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
