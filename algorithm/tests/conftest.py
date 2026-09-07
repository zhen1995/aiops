import pytest
from fastapi.testclient import TestClient
from qdrant_client import QdrantClient

from app.main import app
from app.store import VectorStore


class FakeEmbedder:
    """固定向量假 embedder：每个文本映射到 hash 决定的单位向量"""

    def __init__(self, dim: int = 8):
        self.dim = dim

    def embed_texts(self, texts: list[str]) -> list[list[float]]:
        return [[1.0] + [0.0] * (self.dim - 1) for _ in texts]


@pytest.fixture()
def memory_store():
    return VectorStore(QdrantClient(":memory:"), collection="test_kb")


@pytest.fixture()
def client(memory_store, monkeypatch):
    from app.routers import knowledge
    monkeypatch.setattr(knowledge, "get_store", lambda: memory_store)
    monkeypatch.setattr(knowledge, "get_embedder", lambda cfg: FakeEmbedder())
    with TestClient(app) as c:
        yield c
