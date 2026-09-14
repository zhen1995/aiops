import pytest
from fastapi.testclient import TestClient

from app.main import app
from app.routers import logs as logs_router


@pytest.fixture()
def client():
    # 每个用例使用全新的 miner 缓存，避免聚类 ID 互相影响
    logs_router._miners.clear()
    with TestClient(app) as c:
        yield c
    logs_router._miners.clear()


def _parse(client, service_id, lines):
    resp = client.post("/api/v1/logs/parse", json={"service_id": service_id, "logs": lines})
    assert resp.status_code == 200
    return resp.json()["results"]


def test_same_template_variants_share_cluster_and_params(client):
    lines = [
        "Connection to database order-db failed after 5000ms",
        "Connection to database user-db failed after 120ms",
    ]
    # 先预热一条使模板泛化出 <*> 占位符（drain3 对全新聚类首条返回原始行作为模板）
    _parse(client, "svc-1", [lines[0]])
    results = _parse(client, "svc-1", lines)
    assert len(results) == 2
    # 同一模板的多条变体：聚类相同、模板相同、参数按 <*> 位置抽取
    assert results[0]["cluster_id"] == results[1]["cluster_id"]
    assert results[0]["template"] == results[1]["template"]
    assert "<*>" in results[0]["template"]
    assert results[0]["params"] == ["order-db", "5000ms"]
    assert results[1]["params"] == ["user-db", "120ms"]


def test_different_templates_get_different_clusters(client):
    results = _parse(client, "svc-1", [
        "Connection to database order-db failed after 5000ms",
        "user alice logged in from 10.0.0.1",
    ])
    assert results[0]["cluster_id"] != results[1]["cluster_id"]
    assert results[0]["template"] != results[1]["template"]


def test_level_from_structured_marker(client):
    results = _parse(client, "svc-1", ['{"level":"ERROR","msg":"request rejected"}'])
    assert results[0]["level"] == "error"


def test_level_from_bracket_marker(client):
    results = _parse(client, "svc-1", ["[WARN] disk usage above 80%"])
    assert results[0]["level"] == "warn"


def test_level_keyword_heuristic(client):
    results = _parse(client, "svc-1", [
        "something failed with timeout",   # error 关键词
        "cache miss warning triggered",    # warn 关键词
        "service started successfully",    # 无线索 → info
    ])
    assert [r["level"] for r in results] == ["error", "warn", "info"]


def test_level_structured_marker_takes_priority(client):
    # 行内带结构化级别标记时优先用之，而非按关键词推断
    results = _parse(client, "svc-1", ["level=info an error happened during cleanup test"])
    assert results[0]["level"] == "info"


def test_service_id_isolation(client):
    line = "Connection to database order-db failed after 5000ms"
    r1 = _parse(client, "svc-a", [line])
    r2 = _parse(client, "svc-b", [line])
    # 不同 service_id 各自聚类，cluster_id 从 1 独立编号
    assert r1[0]["cluster_id"] == r2[0]["cluster_id"] == 1

    # 同 service_id 再次解析同一模板，聚类保持稳定
    r3 = _parse(client, "svc-a", ["Connection to database pay-db failed after 30ms"])
    assert r3[0]["cluster_id"] == r1[0]["cluster_id"]


def test_empty_service_id_works(client):
    results = _parse(client, "", ["user bob logged in from 192.168.1.1"])
    assert results[0]["cluster_id"] == 1
    assert results[0]["level"] == "info"


def test_results_match_logs_one_to_one(client):
    lines = [f"request {i} handled in {i}ms" for i in range(20)]
    results = _parse(client, "svc-1", lines)
    assert len(results) == len(lines)
    assert results[10]["params"] == ["10", "10ms"]


def test_too_many_logs_rejected(client):
    resp = client.post("/api/v1/logs/parse", json={
        "service_id": "svc-1",
        "logs": ["x"] * (logs_router.MAX_LOGS_PER_REQUEST + 1),
    })
    assert resp.status_code == 400


def test_detect_level_map():
    from app.routers.logs import detect_level

    assert detect_level("level=WARN something") == "warn"
    assert detect_level('{"level":"debug"}') == "info"
    assert detect_level("FATAL: application crashed") == "error"
    assert detect_level("Exception raised in worker") == "error"
    assert detect_level("normal operation") == "info"
