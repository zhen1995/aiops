import httpx
import respx
from app.embedder import EmbedConfig, Embedder


def _embedding_payload(n: int) -> dict:
    return {"object": "list", "data": [
        {"index": i, "embedding": [0.1, 0.2, 0.3]} for i in range(n)
    ], "model": "m", "usage": {"prompt_tokens": 1, "total_tokens": 1}}


def test_embed_single_batch():
    cfg = EmbedConfig(base_url="http://fake/v1", api_key="sk-x", model="bge-m3")
    emb = Embedder(cfg)
    with respx.mock:
        respx.post("http://fake/v1/embeddings").mock(
            return_value=httpx.Response(200, json=_embedding_payload(2)))
        out = emb.embed_texts(["hello", "world"])
    assert len(out) == 2
    assert out[0] == [0.1, 0.2, 0.3]


def test_embed_batches_over_64():
    cfg = EmbedConfig(base_url="http://fake/v1", api_key="sk-x", model="bge-m3")
    emb = Embedder(cfg)
    texts = [f"t{i}" for i in range(130)]
    with respx.mock:
        respx.post("http://fake/v1/embeddings").mock(
            side_effect=[
                httpx.Response(200, json=_embedding_payload(64)),
                httpx.Response(200, json=_embedding_payload(64)),
                httpx.Response(200, json=_embedding_payload(2)),
            ])
        out = emb.embed_texts(texts)
    assert len(out) == 130
