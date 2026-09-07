import uuid

from qdrant_client import QdrantClient
from qdrant_client.models import (
    Distance, FieldCondition, Filter, MatchValue,
    PointStruct, VectorParams,
)


class VectorStore:
    def __init__(self, client: QdrantClient, collection: str):
        self.client = client
        self.collection = collection

    def ensure_collection(self, vector_size: int) -> None:
        cols = [c.name for c in self.client.get_collections().collections]
        if self.collection not in cols:
            self.client.create_collection(
                collection_name=self.collection,
                vectors_config=VectorParams(size=vector_size, distance=Distance.COSINE),
            )

    def upsert_chunks(self, document_id: str, chunks: list[str],
                      vectors: list[list[float]], title: str) -> int:
        points = [
            PointStruct(
                # Qdrant 本地模式要求 UUID/整数点 ID，用 UUID5 由业务 ID 确定性生成
                id=str(uuid.uuid5(uuid.NAMESPACE_URL, f"kb-chunk/{document_id}-{i}")),
                vector=v,
                payload={
                    "document_id": document_id,
                    "chunk_index": i,
                    "content": c,
                    "title": title,
                },
            )
            for i, (c, v) in enumerate(zip(chunks, vectors))
        ]
        self.client.upsert(collection_name=self.collection, points=points)
        return len(points)

    def search(self, vector: list[float], top_k: int) -> list[dict]:
        resp = self.client.query_points(
            collection_name=self.collection,
            query=vector,
            limit=top_k,
            with_payload=True,
        )
        return [
            {
                "chunk_id": r.id,
                "document_id": r.payload["document_id"],
                "chunk_index": r.payload["chunk_index"],
                "content": r.payload["content"],
                "title": r.payload["title"],
                "score": r.score,
            }
            for r in resp.points
        ]

    def delete_document(self, document_id: str) -> None:
        self.client.delete(
            collection_name=self.collection,
            points_selector=Filter(must=[
                FieldCondition(key="document_id", match=MatchValue(value=document_id))
            ]),
        )
