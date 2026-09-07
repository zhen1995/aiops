from fastapi import APIRouter, File, Form, HTTPException, UploadFile
from qdrant_client import QdrantClient

from ..chunker import split_text
from ..config import get_settings
from ..embedder import EmbedConfig, get_embedder
from ..parser import parse_document
from ..schemas import RetrieveRequest
from ..store import VectorStore

router = APIRouter(prefix="/api/v1/knowledge", tags=["knowledge"])

_store: VectorStore | None = None


def get_store() -> VectorStore:
    """单例 store（测试通过 monkeypatch 替换）"""
    global _store
    if _store is None:
        s = get_settings()
        _store = VectorStore(QdrantClient(url=s.qdrant_url), collection=s.collection)
    return _store


@router.post("/index")
async def index_document(
    document_id: str = Form(...),
    title: str = Form(...),
    file: UploadFile = File(...),
    base_url: str = Form(...),
    api_key: str = Form(...),
    model: str = Form(...),
):
    data = await file.read()
    try:
        text = parse_document(file.filename or "", data)
    except ValueError as e:
        raise HTTPException(status_code=415, detail=str(e))

    s = get_settings()
    chunks = split_text(text, s.chunk_size, s.chunk_overlap)
    if not chunks:
        raise HTTPException(status_code=422, detail="文档解析后无有效文本内容")

    embedder = get_embedder(EmbedConfig(base_url=base_url, api_key=api_key, model=model))
    vectors = embedder.embed_texts(chunks)

    store = get_store()
    store.ensure_collection(len(vectors[0]))
    n = store.upsert_chunks(document_id, chunks, vectors, title)
    return {"document_id": document_id, "chunks": n}


@router.post("/retrieve")
def retrieve(req: RetrieveRequest):
    embedder = get_embedder(req.embedding)
    vector = embedder.embed_texts([req.query])[0]
    results = get_store().search(vector, req.top_k)
    return {"results": results}


@router.delete("/documents/{document_id}")
def delete_document(document_id: str):
    get_store().delete_document(document_id)
    return {"deleted": True}
