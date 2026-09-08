from fastapi import APIRouter, File, Form, HTTPException, UploadFile
from openai import APIConnectionError, APIError, AuthenticationError, NotFoundError
from qdrant_client import QdrantClient

from ..chunker import split_text
from ..config import get_settings
from ..embedder import EmbedConfig, get_embedder
from ..parser import parse_document
from ..schemas import RetrieveRequest
from ..store import VectorStore

router = APIRouter(prefix="/api/v1/knowledge", tags=["knowledge"])

# 单个上传文件的大小上限（50MB）
MAX_FILE_SIZE = 50 * 1024 * 1024

_store: VectorStore | None = None


def _check_size(n: int) -> None:
    """校验上传文件大小，超限抛 413（抽成独立函数便于单测）"""
    if n > MAX_FILE_SIZE:
        raise HTTPException(status_code=413, detail="文件超过 50MB 限制")


def _embed_texts(embedder, texts: list[str]) -> list[list[float]]:
    """调用向量化，并把服务商异常翻译为可读的 502 错误，便于 Go 端记录真实原因"""
    try:
        return embedder.embed_texts(texts)
    except NotFoundError as e:
        raise HTTPException(
            status_code=502,
            detail=f"向量化模型不存在或该服务商不支持 Embeddings 接口：{e.message}",
        )
    except AuthenticationError:
        raise HTTPException(status_code=502, detail="向量化服务鉴权失败，请检查 API Key 是否有效")
    except APIConnectionError as e:
        raise HTTPException(status_code=502, detail=f"无法连接向量化服务：{e}")
    except APIError as e:
        raise HTTPException(status_code=502, detail=f"向量化服务调用失败：{e.message}")


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
    _check_size(len(data))
    try:
        text = parse_document(file.filename or "", data)
    except ValueError as e:
        raise HTTPException(status_code=415, detail=str(e))

    s = get_settings()
    chunks = split_text(text, s.chunk_size, s.chunk_overlap)
    if not chunks:
        raise HTTPException(status_code=422, detail="文档解析后无有效文本内容")

    embedder = get_embedder(EmbedConfig(base_url=base_url, api_key=api_key, model=model))
    vectors = _embed_texts(embedder, chunks)

    store = get_store()
    store.ensure_collection(len(vectors[0]))
    n = store.upsert_chunks(document_id, chunks, vectors, title)
    return {"document_id": document_id, "chunks": n}


@router.post("/retrieve")
def retrieve(req: RetrieveRequest):
    embedder = get_embedder(req.embedding)
    vector = _embed_texts(embedder, [req.query])[0]
    results = get_store().search(vector, req.top_k)
    return {"results": results}


@router.delete("/documents/{document_id}")
def delete_document(document_id: str):
    get_store().delete_document(document_id)
    return {"deleted": True}
