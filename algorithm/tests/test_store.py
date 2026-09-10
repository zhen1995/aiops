from qdrant_client import QdrantClient
from app.store import VectorStore


def _store() -> VectorStore:
    return VectorStore(QdrantClient(":memory:"), collection="test_kb")


def test_ensure_and_upsert_search():
    store = _store()
    store.ensure_collection(3)
    n = store.upsert_chunks(
        "doc-1", ["MySQL 超时排查", "Redis 内存告警处理"], [[1.0, 0, 0], [0, 1.0, 0]], "运维手册")
    assert n == 2
    hits = store.search([1.0, 0, 0], top_k=2)
    assert len(hits) == 2
    assert hits[0]["document_id"] == "doc-1"
    assert hits[0]["content"] == "MySQL 超时排查"
    assert hits[0]["score"] > hits[1]["score"]


def test_delete_document():
    store = _store()
    store.ensure_collection(3)
    store.upsert_chunks("doc-1", ["a"], [[1.0, 0, 0]], "t")
    store.upsert_chunks("doc-2", ["b"], [[0, 1.0, 0]], "t")
    store.delete_document("doc-1")
    hits = store.search([1.0, 0, 0], top_k=10)
    assert all(h["document_id"] != "doc-1" for h in hits)
    assert any(h["document_id"] == "doc-2" for h in hits)


def test_upsert_large_document_batched():
    """超过单批 256 点的大文档分片 upsert 后全部可检索（防 Qdrant 32MB 请求体上限）"""
    store = _store()
    store.ensure_collection(3)
    n = 600
    chunks = [f"chunk-{i}" for i in range(n)]
    vectors = [[1.0, 0, 0] if i % 2 == 0 else [0, 1.0, 0] for i in range(n)]
    assert store.upsert_chunks("big-doc", chunks, vectors, "t") == n
    hits = store.search([1.0, 0, 0], top_k=n)
    assert len(hits) == n
    assert all(h["document_id"] == "big-doc" for h in hits)
