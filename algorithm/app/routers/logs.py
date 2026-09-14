import re
import threading

from drain3 import TemplateMiner
from drain3.template_miner_config import TemplateMinerConfig
from fastapi import APIRouter, HTTPException

from ..schemas import LogParseRequest

router = APIRouter(prefix="/api/v1/logs", tags=["logs"])

# 单请求日志条数上限，防止超大批量拖垮服务
MAX_LOGS_PER_REQUEST = 10000

# 按 service_id 缓存的 Drain 挖掘器：每个服务独立的模板树，保证 cluster_id 在服务内稳定
_miners: dict[str, TemplateMiner] = {}
_miners_lock = threading.Lock()

# Drain 挖掘器非线程安全，解析循环整体加锁串行化
_parse_lock = threading.Lock()


def get_miner(service_id: str) -> TemplateMiner:
    """按 service_id 获取独立的模板挖掘器（首次访问时创建并缓存）"""
    with _miners_lock:
        miner = _miners.get(service_id)
        if miner is None:
            miner = TemplateMiner(config=TemplateMinerConfig())
            _miners[service_id] = miner
        return miner


# 结构化级别标记：level=ERROR / level: error / "level":"warn"
_STRUCT_LEVEL_RE = re.compile(r'"?level"?\s*[=:]\s*"?([A-Za-z]+)"?', re.IGNORECASE)
# 括号级别标记：[ERROR]
_BRACKET_LEVEL_RE = re.compile(r"\[\s*([A-Za-z]+)\s*\]")

# 级别词归一化映射
_LEVEL_MAP = {
    "error": "error",
    "err": "error",
    "fatal": "error",
    "critical": "error",
    "crit": "error",
    "severe": "error",
    "alert": "error",
    "emergency": "error",
    "warn": "warn",
    "warning": "warn",
    "info": "info",
    "notice": "info",
    "debug": "info",
    "trace": "info",
}


def detect_level(line: str) -> str:
    """判定日志级别：优先取行内结构化级别标记，缺失时按关键词启发式归类"""
    m = _STRUCT_LEVEL_RE.search(line)
    if m:
        return _LEVEL_MAP.get(m.group(1).lower(), "info")

    m = _BRACKET_LEVEL_RE.search(line)
    if m and m.group(1).lower() in _LEVEL_MAP:
        return _LEVEL_MAP[m.group(1).lower()]

    if re.search(r"error|fatal|exception|failed|timeout", line, re.IGNORECASE):
        return "error"
    if re.search(r"warn", line, re.IGNORECASE):
        return "warn"
    return "info"


@router.post("/parse")
def parse_logs(req: LogParseRequest):
    """批量 Drain 模板解析：返回与请求日志一一对应的模板、参数、聚类 ID 与级别"""
    if len(req.logs) > MAX_LOGS_PER_REQUEST:
        raise HTTPException(
            status_code=400,
            detail=f"单请求日志条数超过上限 {MAX_LOGS_PER_REQUEST}",
        )

    miner = get_miner(req.service_id)
    results = []
    with _parse_lock:
        # 第一遍：逐条挖掘，模板可能随批次内新变体到达而泛化
        cluster_ids = [miner.add_log_message(line)["cluster_id"] for line in req.logs]
        # 第二遍：按聚类最终模板统一抽取参数，保证同批次内结果一致
        for line, cluster_id in zip(req.logs, cluster_ids):
            template = miner.drain.id_to_cluster[cluster_id].get_template()
            params = miner.get_parameter_list(template, line) or []
            results.append({
                "template": template,
                "params": params,
                "cluster_id": cluster_id,
                "level": detect_level(line),
            })
    return {"results": results}
