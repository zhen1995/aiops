from pydantic import BaseModel, Field


class EmbeddingConfig(BaseModel):
    base_url: str
    api_key: str
    model: str


class RetrieveRequest(BaseModel):
    query: str = Field(min_length=1)
    top_k: int = Field(default=5, ge=1, le=50)
    embedding: EmbeddingConfig
