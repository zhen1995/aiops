from openai import OpenAI
from pydantic import BaseModel


class EmbedConfig(BaseModel):
    """向量化模型配置（由 Go 端透传）"""
    base_url: str
    api_key: str
    model: str


class Embedder:
    BATCH_SIZE = 64

    def __init__(self, config: EmbedConfig):
        self.config = config
        # timeout 必须显著小于 Go 端调用本服务的 120s 超时，max_retries 限制重试次数，
        # 保证 embedding 接口不可达时快速失败并返回真实错误，而不是拖垮整个 uvicorn worker
        self._client = OpenAI(base_url=config.base_url, api_key=config.api_key,
                              timeout=30, max_retries=1)

    def embed_texts(self, texts: list[str]) -> list[list[float]]:
        """批量向量化，自动按 BATCH_SIZE 分批，返回与输入等长的向量列表"""
        vectors: list[list[float]] = []
        for i in range(0, len(texts), self.BATCH_SIZE):
            batch = texts[i:i + self.BATCH_SIZE]
            resp = self._client.embeddings.create(model=self.config.model, input=batch)
            ordered = sorted(resp.data, key=lambda d: d.index)
            vectors.extend(d.embedding for d in ordered)
        return vectors


def get_embedder(config: EmbedConfig) -> Embedder:
    return Embedder(config)
