"""异常检测技能：对时序数据做确定性分析（3-sigma / IQR / 环比基线漂移 / 突变点），输出有限异常信号。
原则：计算交给程序（可复现），模型只消费信号摘要。纯标准库实现。
"""
import statistics
from datetime import datetime, timezone
from typing import Any

from fastapi import APIRouter

router = APIRouter(prefix="/api/v1/anomaly", tags=["anomaly"])

DEFAULTS = {
    "sigma": 3.0,        # 3-sigma 离群阈值
    "iqr_k": 1.5,        # IQR 离群系数
    "recent_ratio": 0.34,  # 最近窗口占比（约 1/3）
    "shift_z": 2.0,      # 环比基线漂移 z 阈值
    "max_signals": 10,   # 返回信号上限
}


def _to_ms(ts: Any) -> int:
    """支持秒/毫秒时间戳或 ISO 字符串，统一输出毫秒时间戳。"""
    if isinstance(ts, (int, float)):
        return int(ts * 1000) if ts < 1e12 else int(ts)
    if isinstance(ts, str):
        s = ts.replace("Z", "+00:00")
        try:
            return int(datetime.fromisoformat(s).timestamp() * 1000)
        except ValueError:
            pass
    raise ValueError(f"无法解析时间点: {ts!r}")


def _fmt_ms(ms: int) -> str:
    return datetime.fromtimestamp(ms / 1000, tz=timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")


def _detect(series: dict, opts: dict) -> dict | None:
    """对单条时序做检测，返回异常信号或 None。"""
    points = sorted(
        ((_to_ms(p[0]), float(p[1])) for p in series.get("points", [])),
        key=lambda x: x[0],
    )
    if len(points) < 6:
        return None
    values = [v for _, v in points]
    n = len(values)
    recent_n = max(2, int(n * opts["recent_ratio"]))
    recent = values[-recent_n:]
    baseline = values[: n - recent_n]
    if len(baseline) < 3:
        baseline = values[: max(3, n - 1)]
        recent = values[-max(2, n - 3):]

    mean = statistics.fmean(values)
    std = statistics.pstdev(values) or 1e-9
    recent_mean = statistics.fmean(recent)
    baseline_mean = statistics.fmean(baseline)
    baseline_std = statistics.pstdev(baseline) or 1e-9
    # 环比：最近窗口相对前置基线的偏移（z 分数）
    deviation = (recent_mean - baseline_mean) / baseline_std

    signal_type = None
    # 环比基线漂移：最近窗口相对基线显著偏移
    if abs(deviation) >= opts["shift_z"]:
        signal_type = "level_shift" if abs(deviation) < 6 else ("spike" if deviation > 0 else "drop")
    else:
        # 3-sigma / IQR 离群（最近窗口内逐点，保留偏离符号）
        q1, _, q3 = statistics.quantiles(values, n=4)
        iqr = (q3 - q1) or 1e-9
        lo, hi = q1 - opts["iqr_k"] * iqr, q3 + opts["iqr_k"] * iqr
        outliers = [v for v in recent if abs(v - mean) > opts["sigma"] * std or v < lo or v > hi]
        if outliers:
            signal_type = "spike" if statistics.fmean(outliers) > mean else "drop"
            worst = max(outliers, key=lambda v: abs((v - mean) / std))
            deviation = (worst - mean) / std

    if signal_type is None:
        return None
    severity = "high" if abs(deviation) >= 6 else ("medium" if abs(deviation) >= 3 else "low")
    return {
        "metric": series.get("metric", ""),
        "labels": series.get("labels", {}),
        "type": signal_type,
        "deviation": round(deviation, 2),
        "window": [_fmt_ms(points[0][0]), _fmt_ms(points[-1][0])],
        "recent_mean": round(recent_mean, 4),
        "overall_mean": round(mean, 4),
        "severity": severity,
    }


@router.post("/detect")
def detect(payload: dict):
    """输入多条时序，输出异常信号列表（Top-N）与人读摘要。"""
    opts = {**DEFAULTS, **(payload.get("options") or {})}
    signals = []
    for series in payload.get("series", []):
        try:
            sig = _detect(series, opts)
        except (ValueError, TypeError, statistics.StatisticsError):
            continue
        if sig:
            signals.append(sig)
    signals.sort(key=lambda s: abs(s["deviation"]), reverse=True)
    signals = signals[: int(opts["max_signals"])]

    if signals:
        top = signals[0]
        summary = (
            f"检测到 {len(signals)} 条异常信号，最显著：{top['metric']} "
            f"{top['type']}（偏离 {top['deviation']}σ，严重度 {top['severity']}）"
        )
    else:
        summary = "未检测到显著异常"
    return {"signals": signals, "summary": summary}
