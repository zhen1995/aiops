from fastapi import FastAPI

from .routers import knowledge

app = FastAPI(title="aiops-analyzer", version="0.1.0")
app.include_router(knowledge.router)


@app.get("/health")
def health():
    return {"status": "ok"}
