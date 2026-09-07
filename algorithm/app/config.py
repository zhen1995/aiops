from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """知识库服务配置（全部可用环境变量覆盖）"""
    model_config = SettingsConfigDict(env_prefix="KB_")

    qdrant_url: str = "http://localhost:6333"
    collection: str = "kb_chunks"
    chunk_size: int = 500        # 每块目标字符数
    chunk_overlap: int = 50      # 块间重叠字符数


def get_settings() -> Settings:
    return Settings()
