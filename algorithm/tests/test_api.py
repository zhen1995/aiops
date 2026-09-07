def test_health(client):
    assert client.get("/health").json() == {"status": "ok"}


def test_index_and_retrieve(client):
    files = {"file": ("runbook.md", "# 排障\nMySQL 连接超时检查 max_connections。", "text/markdown")}
    data = {"document_id": "doc-1", "title": "MySQL 手册",
            "base_url": "http://x/v1", "api_key": "k", "model": "m"}
    resp = client.post("/api/v1/knowledge/index", data=data, files=files)
    assert resp.status_code == 200
    assert resp.json()["chunks"] >= 1

    resp = client.post("/api/v1/knowledge/retrieve", json={
        "query": "连接超时怎么办", "top_k": 3,
        "embedding": {"base_url": "http://x/v1", "api_key": "k", "model": "m"},
    })
    assert resp.status_code == 200
    results = resp.json()["results"]
    assert len(results) >= 1
    assert results[0]["document_id"] == "doc-1"
    assert results[0]["title"] == "MySQL 手册"


def test_index_unsupported_type(client):
    files = {"file": ("evil.exe", b"MZ", "application/octet-stream")}
    data = {"document_id": "doc-2", "title": "x", "base_url": "u", "api_key": "k", "model": "m"}
    resp = client.post("/api/v1/knowledge/index", data=data, files=files)
    assert resp.status_code == 415


def test_delete_document(client):
    files = {"file": ("a.md", "内容", "text/markdown")}
    data = {"document_id": "doc-3", "title": "t", "base_url": "u", "api_key": "k", "model": "m"}
    assert client.post("/api/v1/knowledge/index", data=data, files=files).status_code == 200
    assert client.delete("/api/v1/knowledge/documents/doc-3").json() == {"deleted": True}
    resp = client.post("/api/v1/knowledge/retrieve", json={
        "query": "内容", "top_k": 5,
        "embedding": {"base_url": "u", "api_key": "k", "model": "m"}})
    assert all(r["document_id"] != "doc-3" for r in resp.json()["results"])
