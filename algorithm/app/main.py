from fastapi import FastAPI

from .config import get_settings

app = FastAPI(title="aiops-analyzer", version="0.1.0")


@app.get("/health")
def health():
    return {"status": "ok"}


@app.on_event("startup")
def ensure_qdrant_collection():
    """启动时确保 Qdrant collection 存在（向量维度由 Task 6 首次写入时确定）"""
    # Task 6 填充：调用 store.ensure_collection()
    get_settings()
