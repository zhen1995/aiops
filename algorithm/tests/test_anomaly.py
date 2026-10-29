"""异常检测端点测试：spike / drop / level_shift / 正常序列 / 空与非法输入。"""
from fastapi.testclient import TestClient

from app.main import app

client = TestClient(app)


def _series(name, values, base_ts=1_700_000_000_000, step_ms=60_000):
    points = [[base_ts + i * step_ms, v] for i, v in enumerate(values)]
    return {"metric": name, "labels": {"instance": "test"}, "points": points}


def _detect(series_list):
    resp = client.post("/api/v1/anomaly/detect", json={"series": series_list})
    assert resp.status_code == 200
    return resp.json()


def test_spike_detected():
    values = [10.0] * 30 + [80.0, 90.0, 85.0]
    out = _detect([_series("cpu", values)])
    assert out["signals"], "spike 应被检出"
    sig = out["signals"][0]
    assert sig["metric"] == "cpu"
    assert sig["type"] in ("spike", "level_shift")
    assert sig["deviation"] > 0
    assert "最显著" in out["summary"]


def test_drop_detected():
    values = [50.0] * 30 + [2.0, 1.0, 3.0]
    out = _detect([_series("qps", values)])
    assert out["signals"]
    assert out["signals"][0]["type"] in ("drop", "level_shift")
    assert out["signals"][0]["deviation"] < 0


def test_level_shift():
    values = [10.0] * 20 + [40.0] * 15
    out = _detect([_series("mem", values)])
    assert out["signals"], "基线漂移应被检出"
    assert out["signals"][0]["type"] == "level_shift"


def test_normal_series_no_signal():
    values = [10.0, 10.5, 9.8, 10.2, 10.1, 9.9, 10.3, 10.0, 10.2, 9.8,
              10.1, 10.0, 10.2, 9.9, 10.1, 10.0, 10.3, 9.9, 10.1, 10.0]
    out = _detect([_series("stable", values)])
    assert out["signals"] == []
    assert out["summary"] == "未检测到显著异常"


def test_empty_and_short_series():
    out = _detect([_series("empty", []), _series("short", [1.0, 2.0])])
    assert out["signals"] == []


def test_iso_timestamp_supported():
    import datetime as dt
    base = dt.datetime(2026, 10, 27, tzinfo=dt.timezone.utc)
    points = []
    for i, v in enumerate([10.0] * 20 + [99.0] * 5):
        ts = (base + dt.timedelta(minutes=i)).isoformat()
        points.append([ts, v])
    out = _detect([{"metric": "iso", "labels": {}, "points": points}])
    assert out["signals"]


def test_max_signals_limit():
    series = [_series(f"m{i}", [1.0] * 20 + [50.0 + i]) for i in range(15)]
    out = _detect(series)
    assert len(out["signals"]) <= 10


def test_baseline_fit_deviation():
    # 稳定上行趋势（每天 +1），末尾突然翻倍——环比均值对比可能不敏感，拟合基线应检出
    values = [float(i) for i in range(40)] + [80.0, 82.0, 84.0, 86.0]
    out = _detect([_series("trend", values)])
    assert out["signals"], "趋势曲线上的突增应被拟合基线检出"
    types = [s["type"] for s in out["signals"]]
    assert any(t in ("baseline_deviation", "spike", "drop", "level_shift") for t in types)


def test_periodic_deviation_with_history():
    # 当前序列平稳在 50，历史同时段也是 50（无异常）→ 无信号
    cur = _series("p1", [50.0] * 24)
    cur["history_points"] = [[p[0], 50.0] for p in cur["points"]]
    assert _detect([cur])["signals"] == []

    # 历史同时段在 10，当前在 50 → 周期性强于平时
    cur2 = _series("p2", [50.0] * 24)
    cur2["history_points"] = [[p[0], 10.0] for p in cur2["points"]]
    out = _detect([cur2])
    assert out["signals"], "高于历史同时段应检出"
    assert out["signals"][0]["type"] == "periodic_deviation"
