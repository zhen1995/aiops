from fastapi import FastAPI

from .routers import knowledge, logs

app = FastAPI(title="aiops-analyzer", version="0.1.0")
app.include_router(knowledge.router)
app.include_router(logs.router)


@app.get("/health")
def health():
    return {"status": "ok"}
