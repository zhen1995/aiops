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


def test_check_size_limit():
    from app.routers.knowledge import MAX_FILE_SIZE, _check_size

    _check_size(MAX_FILE_SIZE)  # 边界值不报错
    from fastapi import HTTPException
    import pytest

    with pytest.raises(HTTPException) as exc_info:
        _check_size(MAX_FILE_SIZE + 1)
    assert exc_info.value.status_code == 413
    assert "50MB" in exc_info.value.detail


class _RaisingEmbedder:
    """embed_texts 固定抛出指定异常的假 embedder"""

    def __init__(self, exc: Exception):
        self.exc = exc

    def embed_texts(self, texts: list[str]):
        raise self.exc


def _openai_error(cls, status: int, message: str) -> Exception:
    import httpx

    req = httpx.Request("POST", "http://x/v1/embeddings")
    resp = httpx.Response(status, request=req)
    return cls(message, response=resp, body=None)


def test_index_embedding_404_readable(client, monkeypatch):
    from openai import NotFoundError

    from app.routers import knowledge

    err = _openai_error(NotFoundError, 404, "Error code: 404 - model not found")
    monkeypatch.setattr(knowledge, "get_embedder", lambda cfg: _RaisingEmbedder(err))
    files = {"file": ("a.md", "内容", "text/markdown")}
    data = {"document_id": "doc-404", "title": "t", "base_url": "u", "api_key": "k", "model": "m"}
    resp = client.post("/api/v1/knowledge/index", data=data, files=files)
    assert resp.status_code == 502
    assert "向量化" in resp.json()["detail"]


def test_retrieve_embedding_auth_error_readable(client, monkeypatch):
    from openai import AuthenticationError

    from app.routers import knowledge

    err = _openai_error(AuthenticationError, 401, "invalid api key")
    monkeypatch.setattr(knowledge, "get_embedder", lambda cfg: _RaisingEmbedder(err))
    resp = client.post("/api/v1/knowledge/retrieve", json={
        "query": "内容", "top_k": 5,
        "embedding": {"base_url": "u", "api_key": "k", "model": "m"}})
    assert resp.status_code == 502
    assert "API Key" in resp.json()["detail"]
