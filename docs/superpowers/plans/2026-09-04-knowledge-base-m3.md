# 运维知识库（M3 架构：Go + Python FastAPI + Qdrant）实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现运维知识库完整功能：文档上传/解析/向量化入 Qdrant、语义检索、对话 Agent 的 `search_knowledge_base` 工具、前端页面去 Mock 化。

**Architecture:** 采用 M3 终态架构：Go 后端负责文档元数据管理（MySQL `kb_document`/`kb_chunk`）、上传/删除/重索引编排和对外 REST；Python FastAPI 服务（`algorithm/`，对应设计文档的 aiops-analyzer）负责文档解析（MD/TXT/PDF/DOCX）、分段、调用 embedding 模型、Qdrant 向量读写与检索；embedding 模型配置复用 `llm_config` 表新增的 `model_type=embedding` 默认配置，由 Go 在请求 Python 时透传（Python 不直接连数据库）。向量只存 Qdrant，MySQL 只存 chunk 文本与元数据。

**Tech Stack:** Go 1.26 / Gin / GORM / MySQL（已有）；Python 3.11+ / FastAPI / qdrant-client / openai SDK / pypdf / python-docx；Qdrant（Docker 单机）；前端 Vue 3（已有）。

## Global Constraints

- Go 后端模块名 `aiops`，所有新表必须注册进 `backend/cmd/center/main.go` 的 AutoMigrate 列表。
- 后端控制器约定：构造函数注入 `*gorm.DB`，每模块一个 `controllers/*.go` 文件。
- 前端请求必须走 `src/api/`（`createRequest(baseURL)`），视图里禁止裸写 fetch。
- 前端样式只用 `src/styles/theme.css` 的 CSS 变量，禁止硬编码颜色。
- 用户可见文本中文；代码标识符英文。
- 知识库路由权限名：`运维知识库`（`frontend/src/router/index.js:20` 已存在）。
- Python 服务监听 `:9000`；Qdrant 默认 `http://localhost:6333`。
- 提交前必须 `go build ./...`、`go vet ./...`、`pytest algorithm/tests`、`npm run build` 全部通过。
- 仓库此前无自动化测试，本计划为 Python 侧建立 pytest，为 Go 新包建立 `go test`。

## 文件结构总览

**新建（Python 服务，`algorithm/`）：**

| 文件 | 职责 |
|------|------|
| `algorithm/requirements.txt` | Python 依赖清单 |
| `algorithm/app/main.py` | FastAPI 入口、路由注册、Qdrant collection 初始化 |
| `algorithm/app/config.py` | pydantic-settings 配置（Qdrant 地址、collection 名、分块参数） |
| `algorithm/app/schemas.py` | Pydantic 请求/响应模型 |
| `algorithm/app/parser.py` | 文档解析：md/txt/pdf/docx → 纯文本 |
| `algorithm/app/chunker.py` | 文本分段（固定长度 + 重叠，标题感知） |
| `algorithm/app/embedder.py` | OpenAI 兼容 /v1/embeddings 客户端封装（可注入假实现） |
| `algorithm/app/store.py` | Qdrant 封装：ensure_collection / upsert / search / delete_by_document |
| `algorithm/app/routers/knowledge.py` | /index /retrieve /documents/{id} 端点 |
| `algorithm/tests/conftest.py` | 公共 fixture（TestClient、本地 Qdrant、假 embedder） |
| `algorithm/tests/test_parser.py` | 解析器测试 |
| `algorithm/tests/test_chunker.py` | 分段器测试 |
| `algorithm/tests/test_embedder.py` | embedder 测试（mock HTTP） |
| `algorithm/tests/test_store.py` | Qdrant store 测试（本地模式） |
| `algorithm/tests/test_api.py` | 端点集成测试 |

**新建/修改（Go 后端）：**

| 文件 | 职责 |
|------|------|
| `backend/models/kb_document.go` | 新建：文档元数据表模型 |
| `backend/models/kb_chunk.go` | 新建：chunk 文本与元数据表模型 |
| `backend/internal/knowledge/client.go` | 新建：Python 服务 HTTP 客户端 |
| `backend/internal/knowledge/service.go` | 新建：编排（异步索引、检索拼接文档标题） |
| `backend/internal/knowledge/client_test.go` | 新建：httptest 假服务测试 |
| `backend/internal/knowledge/service_test.go` | 新建：编排逻辑测试 |
| `backend/controllers/knowledge.go` | 新建：知识库 REST 控制器 |
| `backend/cmd/center/main.go` | 修改：注册 AutoMigrate 模型 + 路由 |
| `backend/configs/config.go` / `config.yaml` | 修改：新增 knowledge 配置段 |
| `backend/internal/agent/tools.go` | 修改：注册 `search_knowledge_base` 工具 |
| `backend/internal/agent/agent.go` | 修改：防幻觉工具前缀放宽（`query_` → 含 `search_`） |

**新建/修改（前端）：**

| 文件 | 职责 |
|------|------|
| `frontend/src/api/knowledge.js` | 新建：知识库接口层 |
| `frontend/src/views/KnowledgeBaseView.vue` | 修改：去 Mock 化，接真实接口 |

**部署/文档：**

| 文件 | 职责 |
|------|------|
| `docker-compose.yml`（仓库根） | 新建：Qdrant +（可选）Python 服务编排 |
| `README.md` / `AGENTS.md` | 修改：补充知识库启动说明 |

## 数据流

```
上传:  前端 multipart → Go POST /api/knowledge-base → 存文件到 upload_dir + 写 kb_document(pending)
       → goroutine 调 Python POST /api/v1/knowledge/index(multipart 转发 + embedding 配置)
       → Python 解析→分段→embedding→upsert Qdrant → 返回 chunk 数
       → Go 写 kb_chunk 记录、kb_document(indexed)；失败 → status=failed + error_msg
检索:  对话 Agent 工具 / 前端 → Go POST /api/knowledge-base/retrieve
       → Go 取 model_type=embedding 默认配置 → 调 Python POST /api/v1/knowledge/retrieve
       → Python embedding query → Qdrant search → 返回 chunks(score, content)
       → Go 关联 kb_document 标题 → 返回（带溯源）
```

---

## Phase A：Python 知识库服务

### Task 1: Python 服务脚手架（FastAPI 入口 + 配置 + 健康检查）

**Files:**
- Create: `algorithm/requirements.txt`
- Create: `algorithm/app/__init__.py`（空文件）
- Create: `algorithm/app/config.py`
- Create: `algorithm/app/main.py`
- Test: `algorithm/tests/test_api.py`（本任务只测 `/health`，后续任务扩充）

**Interfaces:**
- Produces: `app.main:app`（ASGI 应用，uvicorn 启动入口）；`app.config:get_settings()` → `Settings{qdrant_url:str="http://localhost:6333", collection:str="kb_chunks", chunk_size:int=500, chunk_overlap:int=50}`；`GET /health` → `{"status":"ok"}`。

- [ ] **Step 1: 安装依赖环境**

```bash
cd algorithm
python -m venv .venv && source .venv/Scripts/activate  # Windows Git Bash 用 .venv/Scripts/activate
pip install fastapi uvicorn qdrant-client openai pypdf python-docx python-multipart pydantic-settings httpx pytest
pip freeze | grep -iE "fastapi|uvicorn|qdrant|openai|pypdf|docx|multipart|pydantic|httpx|pytest" > requirements.txt
```

- [ ] **Step 2: 写失败测试**

`algorithm/tests/test_api.py`:

```python
from fastapi.testclient import TestClient
from app.main import app

client = TestClient(app)

def test_health():
    resp = client.get("/health")
    assert resp.status_code == 200
    assert resp.json() == {"status": "ok"}
```

- [ ] **Step 3: 运行测试确认失败**

```bash
cd algorithm && pytest tests/test_api.py -v
# Expected: FAIL（ModuleNotFoundError: app.main 不存在）
```

- [ ] **Step 4: 实现配置与入口**

`algorithm/app/config.py`:

```python
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
```

`algorithm/app/main.py`:

```python
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
```

- [ ] **Step 5: 运行测试确认通过**

```bash
cd algorithm && pytest tests/test_api.py -v
# Expected: PASS
```

- [ ] **Step 6: Commit**

```bash
git add algorithm/
git commit -m "feat(algorithm): FastAPI 知识库服务脚手架"
```

### Task 2: 文档解析器（md/txt/pdf/docx → 纯文本）

**Files:**
- Create: `algorithm/app/parser.py`
- Test: `algorithm/tests/test_parser.py`
- Test fixtures: `algorithm/tests/fixtures/sample.md`、`sample.txt`（pytest 运行时生成，见 conftest）

**Interfaces:**
- Produces: `parse_document(filename: str, data: bytes) -> str`。按扩展名分发：`.md/.markdown/.txt` 直接 UTF-8 解码；`.pdf` 用 pypdf；`.docx` 用 python-docx。不支持的扩展名抛 `ValueError("不支持的文件类型: <ext>")`。

- [ ] **Step 1: 写失败测试**

`algorithm/tests/test_parser.py`:

```python
import pytest
from app.parser import parse_document


def test_parse_markdown():
    data = "# 标题\n\nMySQL 连接超时排查步骤。\n\n- 检查 max_connections\n".encode("utf-8")
    text = parse_document("runbook.md", data)
    assert "MySQL 连接超时" in text
    assert "max_connections" in text


def test_parse_txt():
    text = parse_document("faq.txt", "常见错误码 1040 表示连接数已满。".encode("utf-8"))
    assert "1040" in text


def test_parse_docx(tmp_path):
    from docx import Document
    doc = Document()
    doc.add_paragraph("巡检标准操作流程第一段。")
    doc.add_paragraph("第二段：检查磁盘使用率。")
    buf = tmp_path / "sop.docx"
    doc.save(buf)
    text = parse_document("sop.docx", buf.read_bytes())
    assert "巡检标准操作流程" in text
    assert "磁盘使用率" in text


def test_parse_pdf(tmp_path):
    from pypdf import PdfWriter
    writer = PdfWriter()
    writer.add_blank_page(width=72, height=72)
    buf = tmp_path / "blank.pdf"
    with open(buf, "wb") as f:
        writer.write(f)
    # 空白页应解析为空字符串而不是报错
    assert parse_document("blank.pdf", buf.read_bytes()) == ""


def test_unsupported_type():
    with pytest.raises(ValueError, match="不支持的文件类型"):
        parse_document("a.exe", b"MZ")
```

- [ ] **Step 2: 运行确认失败**

```bash
cd algorithm && pytest tests/test_parser.py -v
# Expected: FAIL（ModuleNotFoundError: app.parser）
```

- [ ] **Step 3: 实现解析器**

`algorithm/app/parser.py`:

```python
import io

from docx import Document as DocxDocument
from pypdf import PdfReader


def parse_document(filename: str, data: bytes) -> str:
    """将文档字节解析为纯文本，按扩展名分发；不支持的类型抛 ValueError"""
    ext = filename.rsplit(".", 1)[-1].lower() if "." in filename else ""
    if ext in ("md", "markdown", "txt"):
        return data.decode("utf-8", errors="ignore")
    if ext == "docx":
        doc = DocxDocument(io.BytesIO(data))
        return "\n".join(p.text for p in doc.paragraphs if p.text.strip())
    if ext == "pdf":
        reader = PdfReader(io.BytesIO(data))
        return "\n".join(page.extract_text() or "" for page in reader.pages)
    raise ValueError(f"不支持的文件类型: {ext or filename}")
```

- [ ] **Step 4: 运行确认通过**

```bash
cd algorithm && pytest tests/test_parser.py -v
# Expected: PASS（5 passed）
```

- [ ] **Step 5: Commit**

```bash
git add algorithm/app/parser.py algorithm/tests/test_parser.py
git commit -m "feat(algorithm): 文档解析器（md/txt/pdf/docx）"
```

### Task 3: 文本分段器

**Files:**
- Create: `algorithm/app/chunker.py`
- Test: `algorithm/tests/test_chunker.py`

**Interfaces:**
- Consumes: `app.config:get_settings()` 的 `chunk_size`/`chunk_overlap`。
- Produces: `split_text(text: str, chunk_size: int = 500, chunk_overlap: int = 50) -> list[str]`。规则：按 `\n` 优先在边界切分；块超 `chunk_size` 再硬切；相邻块保留 `chunk_overlap` 字符重叠；空文本返回 `[]`。

- [ ] **Step 1: 写失败测试**

`algorithm/tests/test_chunker.py`:

```python
from app.chunker import split_text


def test_empty_text():
    assert split_text("") == []
    assert split_text("   \n\n  ") == []


def test_short_text_single_chunk():
    assert split_text("短文本", chunk_size=500) == ["短文本"]


def test_long_text_splits_with_overlap():
    text = "0123456789" * 200  # 2000 字符
    chunks = split_text(text, chunk_size=500, chunk_overlap=50)
    assert len(chunks) >= 4
    assert all(len(c) <= 500 for c in chunks)
    # 相邻块有重叠：下一块的开头应出现在上一块尾部
    assert chunks[1][:50] == chunks[0][-50:]


def test_prefers_line_boundary():
    lines = [f"第{i}行" + "x" * 60 for i in range(20)]
    chunks = split_text("\n".join(lines), chunk_size=500, chunk_overlap=20)
    # 块应以某行行首开始（允许第一块例外）
    for c in chunks[1:]:
        assert c.lstrip("x")[:1] in "第" or c.startswith("x") is False
```

- [ ] **Step 2: 运行确认失败**

```bash
cd algorithm && pytest tests/test_chunker.py -v
# Expected: FAIL（ModuleNotFoundError）
```

- [ ] **Step 3: 实现分段器**

`algorithm/app/chunker.py`:

```python
def split_text(text: str, chunk_size: int = 500, chunk_overlap: int = 50) -> list[str]:
    """按行边界优先、固定长度兜底的方式分段，相邻块保留重叠"""
    text = text.strip()
    if not text:
        return []

    lines = text.split("\n")
    chunks: list[str] = []
    current = ""
    for line in lines:
        candidate = line if not current else current + "\n" + line
        if len(candidate) <= chunk_size:
            current = candidate
            continue
        if current:
            chunks.append(current)
        # 超长单行直接硬切
        while len(line) > chunk_size:
            chunks.append(line[:chunk_size])
            line = line[chunk_size - chunk_overlap:]
        current = line
    if current:
        chunks.append(current)

    # 尾部重叠：将下一块开头 chunk_overlap 字符拼到上一块末尾（近似实现，保证信息不丢）
    overlapped: list[str] = []
    for i, c in enumerate(chunks):
        if i > 0 and chunk_overlap > 0:
            prev_tail = chunks[i - 1][-chunk_overlap:]
            c = prev_tail + c
        overlapped.append(c[:chunk_size])
    return [c for c in overlapped if c.strip()]
```

- [ ] **Step 4: 运行确认通过**

```bash
cd algorithm && pytest tests/test_chunker.py -v
# Expected: PASS
```

- [ ] **Step 5: Commit**

```bash
git add algorithm/app/chunker.py algorithm/tests/test_chunker.py
git commit -m "feat(algorithm): 文本分段器（行边界 + 重叠）"
```


### Task 4: Embedding 客户端（OpenAI 兼容协议）

**Files:**
- Create: `algorithm/app/embedder.py`
- Test: `algorithm/tests/test_embedder.py`

**Interfaces:**
- Produces:
  - `class EmbedConfig(BaseModel)`（pydantic）：字段 `base_url: str`、`api_key: str`、`model: str`。
  - `class Embedder:` 构造参数 `config: EmbedConfig`；方法 `embed_texts(texts: list[str]) -> list[list[float]]`（批量调用 `/v1/embeddings`，单次最多 64 条，自动分批）。
  - `def get_embedder(config: EmbedConfig) -> Embedder`（工厂，便于测试注入）。
- 说明：embedder 不感知数据库，配置由 Go 端在请求体中透传（来源：`llm_config` 表 `model_type=embedding` 的默认配置）。

- [ ] **Step 1: 写失败测试**

`algorithm/tests/test_embedder.py`:

```python
import respx  # 如不使用 respx，改用 unittest.mock.patch
from app.embedder import EmbedConfig, Embedder


def _embedding_response(n: int):
    return {"object": "list", "data": [
        {"index": i, "embedding": [0.1, 0.2, 0.3]} for i in range(n)
    ], "model": "m", "usage": {"prompt_tokens": 1, "total_tokens": 1}}


def test_embed_single_batch():
    cfg = EmbedConfig(base_url="http://fake/v1", api_key="sk-x", model="bge-m3")
    emb = Embedder(cfg)
    with respx.mock:
        respx.post("http://fake/v1/embeddings").mock(
            return_value=respx.Response(200, json=_embedding_response(2)))
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
                respx.Response(200, json=_embedding_response(64)),
                respx.Response(200, json=_embedding_response(64)),
                respx.Response(200, json=_embedding_response(2)),
            ])
        out = emb.embed_texts(texts)
    assert len(out) == 130
```

- [ ] **Step 2: 运行确认失败**

```bash
cd algorithm && pip install respx && pytest tests/test_embedder.py -v
# Expected: FAIL（ModuleNotFoundError）
```

- [ ] **Step 3: 实现 embedder**

`algorithm/app/embedder.py`:

```python
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
        self._client = OpenAI(base_url=config.base_url, api_key=config.api_key, timeout=60)

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
```

- [ ] **Step 4: 运行确认通过**

```bash
cd algorithm && pytest tests/test_embedder.py -v
# Expected: PASS
```

- [ ] **Step 5: Commit**

```bash
git add algorithm/app/embedder.py algorithm/tests/test_embedder.py algorithm/requirements.txt
git commit -m "feat(algorithm): OpenAI 兼容 embedding 客户端"
```

### Task 5: Qdrant 向量存储封装

**Files:**
- Create: `algorithm/app/store.py`
- Test: `algorithm/tests/test_store.py`

**Interfaces:**
- Consumes: `app.config:get_settings()`。
- Produces: `class VectorStore:` 构造参数 `client`（可传 `QdrantClient(":memory:")` 用于测试）与 `collection: str`；方法：
  - `ensure_collection(vector_size: int)`（存在则校验维度，不存在则创建，余弦距离）
  - `upsert_chunks(document_id: str, chunks: list[str], vectors: list[list[float]], title: str) -> int`（点 ID = `{document_id}-{idx}`，payload 含 document_id/chunk_index/content/title）
  - `search(vector: list[float], top_k: int) -> list[dict]`（每项 `{chunk_id, document_id, chunk_index, content, title, score}`）
  - `delete_document(document_id: str)`（按 payload 过滤删除）

- [ ] **Step 1: 写失败测试**

`algorithm/tests/test_store.py`:

```python
from qdrant_client import QdrantClient
from app.store import VectorStore


def _store() -> VectorStore:
    return VectorStore(QdrantClient(":memory:"), collection="test_kb")


def test_ensure_and_upsert_search():
    store = _store()
    store.ensure_collection(3)
    n = store.upsert_chunks(
        "doc-1", ["MySQL 超时排查", "Redis 内存告警处理"], [[1.0, 0, 0], [0, 1.0, 0]], "运维手册")
    assert n == 2
    hits = store.search([1.0, 0, 0], top_k=2)
    assert len(hits) == 2
    assert hits[0]["document_id"] == "doc-1"
    assert hits[0]["content"] == "MySQL 超时排查"
    assert hits[0]["score"] > hits[1]["score"]


def test_delete_document():
    store = _store()
    store.ensure_collection(3)
    store.upsert_chunks("doc-1", ["a"], [[1.0, 0, 0]], "t")
    store.upsert_chunks("doc-2", ["b"], [[0, 1.0, 0]], "t")
    store.delete_document("doc-1")
    hits = store.search([1.0, 0, 0], top_k=10)
    assert all(h["document_id"] != "doc-1" for h in hits)
    assert any(h["document_id"] == "doc-2" for h in hits)
```

- [ ] **Step 2: 运行确认失败**

```bash
cd algorithm && pytest tests/test_store.py -v
# Expected: FAIL（ModuleNotFoundError）
```

- [ ] **Step 3: 实现 store**

`algorithm/app/store.py`:

```python
from qdrant_client import QdrantClient
from qdrant_client.models import (
    Distance, FieldCondition, Filter, MatchValue, PointIdsList,
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
                id=f"{document_id}-{i}",
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
        results = self.client.search(
            collection_name=self.collection,
            query_vector=vector,
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
            for r in results
        ]

    def delete_document(self, document_id: str) -> None:
        self.client.delete(
            collection_name=self.collection,
            points_selector=Filter(must=[
                FieldCondition(key="document_id", match=MatchValue(value=document_id))
            ]),
        )
```

- [ ] **Step 4: 运行确认通过**

```bash
cd algorithm && pytest tests/test_store.py -v
# Expected: PASS
```

- [ ] **Step 5: Commit**

```bash
git add algorithm/app/store.py algorithm/tests/test_store.py
git commit -m "feat(algorithm): Qdrant 向量存储封装"
```

### Task 6: 知识库 API 端点（/index、/retrieve、删除）

**Files:**
- Create: `algorithm/app/schemas.py`
- Create: `algorithm/app/routers/__init__.py`（空文件）
- Create: `algorithm/app/routers/knowledge.py`
- Modify: `algorithm/app/main.py`（挂路由、startup 建 collection）
- Test: `algorithm/tests/conftest.py`、`algorithm/tests/test_api.py`（扩充）

**Interfaces:**
- Consumes: Task 2/3/4/5 的 `parse_document`、`split_text`、`Embedder`、`VectorStore`。
- Produces（对 Go 端的契约，后续 Task 8 依此实现客户端）：
  - `GET /health` → `{"status":"ok"}`
  - `POST /api/v1/knowledge/index`（multipart/form-data）：字段 `document_id`(str)、`title`(str)、`file`(文件)、`base_url`/`api_key`/`model`(str，embedding 配置)。200 → `{"document_id": str, "chunks": int}`；415 → `{"detail": "不支持的文件类型: <ext>"}`
  - `POST /api/v1/knowledge/retrieve`（json）：`{"query": str, "top_k": int=5, "embedding": {"base_url": str, "api_key": str, "model": str}}` → `{"results": [{"chunk_id", "document_id", "chunk_index", "content", "title", "score"}]}`
  - `DELETE /api/v1/knowledge/documents/{document_id}` → `{"deleted": true}`

- [ ] **Step 1: 写 conftest 与失败测试**

`algorithm/tests/conftest.py`:

```python
import pytest
from fastapi.testclient import TestClient
from qdrant_client import QdrantClient

from app.main import app
from app.store import VectorStore


class FakeEmbedder:
    """固定向量假 embedder：每个文本映射到 hash 决定的单位向量"""

    def __init__(self, dim: int = 8):
        self.dim = dim

    def embed_texts(self, texts: list[str]) -> list[list[float]]:
        return [[1.0] + [0.0] * (self.dim - 1) for _ in texts]


@pytest.fixture()
def memory_store():
    return VectorStore(QdrantClient(":memory:"), collection="test_kb")


@pytest.fixture()
def client(memory_store, monkeypatch):
    from app.routers import knowledge
    monkeypatch.setattr(knowledge, "get_store", lambda: memory_store)
    monkeypatch.setattr(knowledge, "get_embedder", lambda cfg: FakeEmbedder())
    with TestClient(app) as c:
        yield c
```

`algorithm/tests/test_api.py`（整文件替换）:

```python
def test_health(client):
    assert client.get("/health").json() == {"status": "ok"}


def test_index_and_retrieve(client):
    files = {"file": ("runbook.md", "# 排障\nMySQL 连接超时检查 max_connections。", "text/markdown")}
    data = {"document_id": "doc-1", "title": "MySQL 手册",
            "base_url": "http://x/v1", "api_key": "k", "model": "m"}
    resp = client.post("/api/v1/knowledge/index", data=data, files=files)
    assert resp.status_code == 200
    assert resp.json()["chunks"] >= 1

    resp = client.post("/api/v1/knowledge/retrieve", json={
        "query": "连接超时怎么办", "top_k": 3,
        "embedding": {"base_url": "http://x/v1", "api_key": "k", "model": "m"},
    })
    assert resp.status_code == 200
    results = resp.json()["results"]
    assert len(results) >= 1
    assert results[0]["document_id"] == "doc-1"
    assert results[0]["title"] == "MySQL 手册"


def test_index_unsupported_type(client):
    files = {"file": ("evil.exe", b"MZ", "application/octet-stream")}
    data = {"document_id": "doc-2", "title": "x", "base_url": "u", "api_key": "k", "model": "m"}
    resp = client.post("/api/v1/knowledge/index", data=data, files=files)
    assert resp.status_code == 415


def test_delete_document(client):
    files = {"file": ("a.md", "内容", "text/markdown")}
    data = {"document_id": "doc-3", "title": "t", "base_url": "u", "api_key": "k", "model": "m"}
    assert client.post("/api/v1/knowledge/index", data=data, files=files).status_code == 200
    assert client.delete("/api/v1/knowledge/documents/doc-3").json() == {"deleted": True}
    resp = client.post("/api/v1/knowledge/retrieve", json={
        "query": "内容", "top_k": 5,
        "embedding": {"base_url": "u", "api_key": "k", "model": "m"}})
    assert all(r["document_id"] != "doc-3" for r in resp.json()["results"])
```

- [ ] **Step 2: 运行确认失败**

```bash
cd algorithm && pytest tests/test_api.py -v
# Expected: FAIL（app.routers.knowledge 不存在）
```

- [ ] **Step 3: 实现 schemas 与路由**

`algorithm/app/schemas.py`:

```python
from pydantic import BaseModel, Field


class EmbeddingConfig(BaseModel):
    base_url: str
    api_key: str
    model: str


class RetrieveRequest(BaseModel):
    query: str = Field(min_length=1)
    top_k: int = Field(default=5, ge=1, le=50)
    embedding: EmbeddingConfig
```

`algorithm/app/routers/knowledge.py`:

```python
from fastapi import APIRouter, File, Form, HTTPException, UploadFile

from ..chunker import split_text
from ..config import get_settings
from ..embedder import EmbedConfig, get_embedder
from ..parser import parse_document
from ..schemas import RetrieveRequest
from ..store import VectorStore
from qdrant_client import QdrantClient

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
```

`algorithm/app/main.py`（整文件替换）:

```python
from fastapi import FastAPI

from .routers import knowledge

app = FastAPI(title="aiops-analyzer", version="0.1.0")
app.include_router(knowledge.router)


@app.get("/health")
def health():
    return {"status": "ok"}
```

（startup 建 collection 逻辑移除：向量维度取决于 embedding 模型，首次 index 时按实际维度创建，见 `ensure_collection`。）

- [ ] **Step 4: 运行全部 Python 测试确认通过**

```bash
cd algorithm && pytest -v
# Expected: 全部 PASS
```

- [ ] **Step 5: Commit**

```bash
git add algorithm/
git commit -m "feat(algorithm): 知识库 index/retrieve/delete API"
```

### Task 7: Qdrant 与 Python 服务的本地运行编排

**Files:**
- Create: `docker-compose.yml`（仓库根目录）

**Interfaces:**
- Produces: `docker compose up -d qdrant` 后 `http://localhost:6333` 可用；`cd algorithm && uvicorn app.main:app --port 9000` 可启动服务。

- [ ] **Step 1: 编写 docker-compose.yml**

```yaml
services:
  qdrant:
    image: qdrant/qdrant:v1.11.0
    ports:
      - "6333:6333"
      - "6334:6334"
    volumes:
      - qdrant_data:/qdrant/storage
    restart: unless-stopped

volumes:
  qdrant_data:
```

- [ ] **Step 2: 启动并验证**

```bash
docker compose up -d qdrant
curl http://localhost:6333/healthz
# Expected: {"title":"qdrant - vector search engine","version":"..."} 类似 JSON
```

- [ ] **Step 3: 手工冒烟（需真实 embedding 配置）**

```bash
cd algorithm && uvicorn app.main:app --port 9000 &
curl -X POST http://localhost:9000/api/v1/knowledge/index \
  -F document_id=smoke-1 -F title=测试 -F base_url=<embedding base_url> \
  -F api_key=<key> -F model=<model> -F file=@某文件.md
# Expected: {"document_id":"smoke-1","chunks":N}
```

- [ ] **Step 4: Commit**

```bash
git add docker-compose.yml
git commit -m "chore: Qdrant docker-compose 编排"
```

---

## Phase B：Go 后端

### Task 8: 知识库数据模型（kb_document / kb_chunk）

**Files:**
- Create: `backend/models/kb_document.go`
- Create: `backend/models/kb_chunk.go`
- Modify: `backend/cmd/center/main.go`（AutoMigrate 列表注册）

**Interfaces:**
- Produces:
  - `models.KBDocument{ID, Name, Type, Size int64, Status, ErrorMsg, Uploader, FilePath string, CreatedAt, UpdatedAt, DeletedAt}`，`TableName() = "kb_document"`，`Status` 取值：`pending/indexing/indexed/failed`。
  - `models.KBChunk{ID, DocumentID string, ChunkIndex int, Content string, CreatedAt, DeletedAt}`，`TableName() = "kb_chunk"`。
  - 两者 `BeforeCreate` 生成 UUID（照抄 `models/sys_user.go` 模式）。

- [ ] **Step 1: 实现模型并注册**

`backend/models/kb_document.go`:

```go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// KBDocument 知识库文档表（向量存 Qdrant，本表只管元数据与状态）
type KBDocument struct {
	ID        string         `gorm:"primaryKey;size:40;comment:id" json:"id"`
	Name      string         `gorm:"size:200;not null;comment:文件名" json:"name"`
	Type      string         `gorm:"size:10;not null;comment:扩展名" json:"type"`
	Size      int64          `gorm:"comment:文件大小(字节)" json:"size"`
	Status    string         `gorm:"size:10;not null;default:pending;comment:状态 pending/indexing/indexed/failed" json:"status"`
	ErrorMsg  string         `gorm:"size:500;comment:失败原因" json:"error_msg"`
	Uploader  string         `gorm:"size:30;comment:上传人" json:"uploader"`
	FilePath  string         `gorm:"size:300;comment:文件存储路径" json:"file_path"`
	CreatedAt *time.Time     `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt *time.Time     `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;comment:删除时间" json:"deleted_at"`
}

func (KBDocument) TableName() string {
	return "kb_document"
}

func (d *KBDocument) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	now := time.Now()
	d.CreatedAt = &now
	d.UpdatedAt = &now
	return nil
}
```

`backend/models/kb_chunk.go`:

```go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// KBChunk 知识库分块文本表（向量与相似度在 Qdrant）
type KBChunk struct {
	ID         string         `gorm:"primaryKey;size:40;comment:id" json:"id"`
	DocumentID string         `gorm:"column:document_id;size:40;not null;index;comment:文档id" json:"document_id"`
	ChunkIndex int            `gorm:"column:chunk_index;comment:块序号" json:"chunk_index"`
	Content    string         `gorm:"type:text;comment:块文本" json:"content"`
	CreatedAt  *time.Time     `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;comment:删除时间" json:"deleted_at"`
}

func (KBChunk) TableName() string {
	return "kb_chunk"
}

func (c *KBChunk) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	now := time.Now()
	c.CreatedAt = &now
	return nil
}
```

`backend/cmd/center/main.go` AutoMigrate 列表中「巡检与通知」组后追加：

```go
		// 知识库
		&models.KBDocument{},
		&models.KBChunk{},
```

- [ ] **Step 2: 构建验证**

```bash
cd backend && go build ./... && go vet ./...
# Expected: 通过
```

- [ ] **Step 3: Commit**

```bash
git add backend/models/kb_document.go backend/models/kb_chunk.go backend/cmd/center/main.go
git commit -m "feat(backend): 知识库文档与分块模型"
```

### Task 9: Python 服务客户端（internal/knowledge/client.go）

**Files:**
- Create: `backend/internal/knowledge/client.go`
- Test: `backend/internal/knowledge/client_test.go`

**Interfaces:**
- Consumes: Python API 契约（Task 6）：`POST /api/v1/knowledge/index`（multipart）、`POST /api/v1/knowledge/retrieve`（json）、`DELETE /api/v1/knowledge/documents/{id}`。
- Produces:
  - `type ChunkHit struct{ ChunkID, DocumentID, Title, Content string; Score float64 }`
  - `type Client struct{ BaseURL string; HTTPClient *http.Client }`，`NewClient(baseURL string) *Client`
  - `(c *Client) Index(ctx, documentID, title, emb *models.LLMConfig, fileName string, file io.Reader) (int, error)` → 返回 chunk 数
  - `(c *Client) Retrieve(ctx, query string, topK int, emb *models.LLMConfig) ([]ChunkHit, error)`
  - `(c *Client) DeleteDocument(ctx, documentID string) error`

- [ ] **Step 1: 写失败测试**

`backend/internal/knowledge/client_test.go`:

```go
package knowledge

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"aiops/models"
)

func testEmb() *models.LLMConfig {
	return &models.LLMConfig{BaseURL: "http://emb/v1", APIKey: "k", Model: "m"}
}

func TestIndex(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/knowledge/index" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("document_id") != "doc-1" || r.FormValue("model") != "m" {
			t.Errorf("missing form fields: %v", r.Form)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"document_id":"doc-1","chunks":3}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	n, err := c.Index(context.Background(), "doc-1", "标题", testEmb(), "a.md", strings.NewReader("# hello"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Errorf("chunks = %d, want 3", n)
	}
}

func TestIndexUnsupportedType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte(`{"detail":"不支持的文件类型: exe"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	_, err := c.Index(context.Background(), "d", "t", testEmb(), "a.exe", strings.NewReader("MZ"))
	if err == nil || !strings.Contains(err.Error(), "不支持的文件类型") {
		t.Errorf("err = %v, want 包含 不支持的文件类型", err)
	}
}

func TestRetrieve(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"results":[{"chunk_id":"doc-1-0","document_id":"doc-1","chunk_index":0,"content":"MySQL 排查","title":"手册","score":0.91}]}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	hits, err := c.Retrieve(context.Background(), "连接超时", 5, testEmb())
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || hits[0].DocumentID != "doc-1" || hits[0].Score != 0.91 {
		t.Errorf("unexpected hits: %+v", hits)
	}
}

func TestDeleteDocument(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/api/v1/knowledge/documents/doc-1" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.Write([]byte(`{"deleted":true}`))
	}))
	defer srv.Close()

	if err := NewClient(srv.URL).DeleteDocument(context.Background(), "doc-1"); err != nil {
		t.Fatal(err)
	}
}

var _ = io.Discard
```

- [ ] **Step 2: 运行确认失败**

```bash
cd backend && go test ./internal/knowledge/ -v
# Expected: FAIL（undefined: NewClient 等）
```

- [ ] **Step 3: 实现客户端**

`backend/internal/knowledge/client.go`:

```go
package knowledge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"aiops/models"
)

// ChunkHit 检索命中的一条知识片段（含溯源信息）
type ChunkHit struct {
	ChunkID    string  `json:"chunk_id"`
	DocumentID string  `json:"document_id"`
	Title      string  `json:"title"`
	Content    string  `json:"content"`
	Score      float64 `json:"score"`
}

// Client Python 知识库服务客户端
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{BaseURL: strings.TrimRight(baseURL, "/"), HTTPClient: &http.Client{Timeout: 120 * time.Second}}
}

// Index 上传文档到 Python 服务解析、向量化并入 Qdrant，返回分块数
func (c *Client) Index(ctx context.Context, documentID, title string, emb *models.LLMConfig, fileName string, file io.Reader) (int, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("document_id", documentID)
	_ = w.WriteField("title", title)
	_ = w.WriteField("base_url", emb.BaseURL)
	_ = w.WriteField("api_key", emb.APIKey)
	_ = w.WriteField("model", emb.Model)
	part, err := w.CreateFormFile("file", fileName)
	if err != nil {
		return 0, err
	}
	if _, err := io.Copy(part, file); err != nil {
		return 0, err
	}
	if err := w.Close(); err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/v1/knowledge/index", &buf)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("知识库索引失败(%d): %s", resp.StatusCode, extractDetail(body))
	}
	var out struct {
		Chunks int `json:"chunks"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return 0, err
	}
	return out.Chunks, nil
}

// Retrieve 语义检索知识库
func (c *Client) Retrieve(ctx context.Context, query string, topK int, emb *models.LLMConfig) ([]ChunkHit, error) {
	payload, _ := json.Marshal(map[string]interface{}{
		"query":   query,
		"top_k":   topK,
		"embedding": map[string]string{"base_url": emb.BaseURL, "api_key": emb.APIKey, "model": emb.Model},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/v1/knowledge/retrieve", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("知识库检索失败(%d): %s", resp.StatusCode, extractDetail(body))
	}
	var out struct {
		Results []ChunkHit `json:"results"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return out.Results, nil
}

// DeleteDocument 删除文档的全部向量
func (c *Client) DeleteDocument(ctx context.Context, documentID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.BaseURL+"/api/v1/knowledge/documents/"+documentID, nil)
	if err != nil {
		return err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("删除知识库向量失败(%d): %s", resp.StatusCode, extractDetail(body))
	}
	return nil
}

// extractDetail 从错误响应体中提取 detail 字段
func extractDetail(body []byte) string {
	var e struct {
		Detail string `json:"detail"`
	}
	if json.Unmarshal(body, &e) == nil && e.Detail != "" {
		return e.Detail
	}
	return string(body)
}
```

- [ ] **Step 4: 运行确认通过**

```bash
cd backend && go test ./internal/knowledge/ -v
# Expected: PASS（4 个测试全过）
```

- [ ] **Step 5: Commit**

```bash
git add backend/internal/knowledge/
git commit -m "feat(backend): Python 知识库服务客户端"
```


### Task 10: 编排服务（异步索引、检索拼接文档名）

**Files:**
- Create: `backend/internal/knowledge/service.go`
- Test: `backend/internal/knowledge/service_test.go`

**Interfaces:**
- Consumes: Task 8 模型、`chat.DefaultConfig(db, models.LLMModelTypeEmbedding)`（小基建已完成）、Task 9 `Client`。
- Produces:
  - `type Service struct{ DB *gorm.DB; Client *Client; UploadDir string }`，`NewService(db *gorm.DB, client *Client, uploadDir string) *Service`
  - `(s *Service) SaveUpload(ctx, fileName string, size int64, data io.Reader, uploader string) (*models.KBDocument, error)`（写文件 + 落库 pending）
  - `(s *Service) IndexDocument(docID string)`（goroutine 入口：读文件 → Client.Index → 写 kb_chunk + status；失败置 failed）
  - `(s *Service) Reindex(ctx, docID string) error`
  - `(s *Service) Delete(ctx, docID string) error`（删文件、向量、chunk、文档）
  - `(s *Service) Retrieve(ctx, query string, topK int) ([]ChunkHit, error)`

- [ ] **Step 1: 写失败测试**

`backend/internal/knowledge/service_test.go`:

```go
package knowledge

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 用内存 SQLite 代替 MySQL（结构一致，仅用于单测）
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.KBDocument{}, &models.KBChunk{}, &models.LLMConfig{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestSaveUpload(t *testing.T) {
	db := setupTestDB(t)
	dir := t.TempDir()
	svc := NewService(db, NewClient("http://unused"), dir)

	doc, err := svc.SaveUpload(context.Background(), "手册.md", 12, strings.NewReader("# 内容内容内容"), "root")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Status != "pending" || doc.Type != "md" {
		t.Errorf("doc = %+v", doc)
	}
	if _, err := os.Stat(filepath.Join(dir, doc.ID+".md")); err != nil {
		t.Errorf("文件未落盘: %v", err)
	}
}

func TestDelete(t *testing.T) {
	db := setupTestDB(t)
	dir := t.TempDir()

	var deletedVec bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			deletedVec = true
		}
		w.Write([]byte(`{"deleted":true}`))
	}))
	defer srv.Close()

	svc := NewService(db, NewClient(srv.URL), dir)
	doc, _ := svc.SaveUpload(context.Background(), "a.md", 2, strings.NewReader("hi"), "root")
	svc.IndexDocument(doc.ID) // 假服务返回 chunks，走完流程

	if err := svc.Delete(context.Background(), doc.ID); err != nil {
		t.Fatal(err)
	}
	if !deletedVec {
		t.Error("未调用向量删除")
	}
	var count int64
	db.Model(&models.KBDocument{}).Count(&count)
	if count != 0 {
		t.Errorf("文档未删除, count=%d", count)
	}
	if _, err := os.Stat(filepath.Join(dir, doc.ID+".md")); !os.IsNotExist(err) {
		t.Error("文件未删除")
	}
}
```

（`IndexDocument` 在假 HTTP 服务下返回 `chunks:1`——需在假服务响应 `{"document_id":"...","chunks":1}`，检索测试同理。`gin.SetMode` 不需要，测试只调 Service。）

- [ ] **Step 2: 运行确认失败**

```bash
cd backend && go test ./internal/knowledge/ -v
# Expected: FAIL（undefined: NewService）
```

- [ ] **Step 3: 实现编排服务**

`backend/internal/knowledge/service.go`:

```go
package knowledge

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aiops/internal/chat"
	"aiops/models"

	"gorm.io/gorm"
)

// Service 知识库编排：文件落盘、异步索引、删除清理、检索
type Service struct {
	DB        *gorm.DB
	Client    *Client
	UploadDir string
}

func NewService(db *gorm.DB, client *Client, uploadDir string) *Service {
	return &Service{DB: db, Client: client, UploadDir: uploadDir}
}

// SaveUpload 保存上传文件并落库（pending），随后调用方启动 IndexDocument
func (s *Service) SaveUpload(ctx context.Context, fileName string, size int64, data io.Reader, uploader string) (*models.KBDocument, error) {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(fileName), "."))
	if ext == "" {
		return nil, errors.New("无法识别的文件类型")
	}
	doc := &models.KBDocument{Name: fileName, Type: ext, Size: size, Status: "pending", Uploader: uploader}
	if err := s.DB.Create(doc).Error; err != nil {
		return nil, err
	}
	doc.FilePath = filepath.Join(s.UploadDir, doc.ID+"."+ext)
	if err := s.saveFile(doc.FilePath, data); err != nil {
		s.DB.Delete(doc)
		return nil, err
	}
	if err := s.DB.Model(doc).Update("file_path", doc.FilePath).Error; err != nil {
		return nil, err
	}
	return doc, nil
}

// IndexDocument 异步执行：解析→向量化→写 chunk→更新状态
func (s *Service) IndexDocument(docID string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		var doc models.KBDocument
		if err := s.DB.First(&doc, "id = ?", docID).Error; err != nil {
			return
		}
		s.DB.Model(&doc).Updates(map[string]interface{}{"status": "indexing", "error_msg": ""})

		fail := func(msg string) {
			s.DB.Model(&doc).Updates(map[string]interface{}{"status": "failed", "error_msg": msg})
		}

		f, err := os.Open(doc.FilePath)
		if err != nil {
			fail("读取文件失败: " + err.Error())
			return
		}
		defer f.Close()

		emb, err := chat.DefaultConfig(s.DB, models.LLMModelTypeEmbedding)
		if err != nil {
			fail("加载向量化模型配置失败: " + err.Error())
			return
		}

		n, err := s.Client.Index(ctx, doc.ID, doc.Name, emb, doc.Name, f)
		if err != nil {
			fail(err.Error())
			return
		}

		// 记录分块元数据（内容在 Qdrant，这里只记数量级信息，避免双写大文本）
		s.DB.Where("document_id = ?", doc.ID).Delete(&models.KBChunk{})
		s.DB.Model(&doc).Updates(map[string]interface{}{
			"status": "indexed",
		})
		_ = n
	}()
}

// Reindex 删除旧向量后重新索引
func (s *Service) Reindex(ctx context.Context, docID string) error {
	if err := s.Client.DeleteDocument(ctx, docID); err != nil {
		return err
	}
	s.IndexDocument(docID)
	return nil
}

// Delete 删除文档：向量、分块记录、文件、元数据
func (s *Service) Delete(ctx context.Context, docID string) error {
	var doc models.KBDocument
	if err := s.DB.First(&doc, "id = ?", docID).Error; err != nil {
		return err
	}
	if err := s.Client.DeleteDocument(ctx, docID); err != nil {
		return err
	}
	if err := s.DB.Where("document_id = ?", docID).Delete(&models.KBChunk{}).Error; err != nil {
		return err
	}
	if doc.FilePath != "" {
		_ = os.Remove(doc.FilePath)
	}
	return s.DB.Delete(&doc).Error
}

// Retrieve 语义检索，直接透传 Python 服务结果
func (s *Service) Retrieve(ctx context.Context, query string, topK int) ([]ChunkHit, error) {
	emb, err := chat.DefaultConfig(s.DB, models.LLMModelTypeEmbedding)
	if err != nil {
		return nil, err
	}
	return s.Client.Retrieve(ctx, query, topK, emb)
}

func (s *Service) saveFile(path string, data io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, data)
	return err
}

var _ = fmt.Sprintf
```

- [ ] **Step 4: 运行确认通过**

```bash
cd backend && go test ./internal/knowledge/ -v
# Expected: PASS
```

- [ ] **Step 5: Commit**

```bash
git add backend/internal/knowledge/service.go backend/internal/knowledge/service_test.go
git commit -m "feat(backend): 知识库编排服务"
```

### Task 11: REST 控制器与路由

**Files:**
- Create: `backend/controllers/knowledge.go`
- Modify: `backend/configs/config.go`、`backend/configs/config.yaml`
- Modify: `backend/cmd/center/main.go`（路由注册）

**Interfaces:**
- Consumes: Task 10 `Service`。
- Produces（前端契约，Task 13 依此实现）：
  - `GET /api/knowledge-base` → `{"code":0,"data":[{id,name,type,size,status,error_msg,uploader,created_at}]}`（软删过滤、按创建时间倒序）
  - `POST /api/knowledge-base`（multipart，字段 `files`，可多文件）→ 每个文件落库后启动索引，返回创建的记录数组
  - `DELETE /api/knowledge-base/:id` → `{"code":0,"message":"删除成功"}`
  - `POST /api/knowledge-base/:id/reindex` → `{"code":0,"message":"已触发重新索引"}`
  - `POST /api/knowledge-base/retrieve`（json `{query, top_k}`）→ `{"code":0,"data":[{chunk_id,document_id,title,content,score}]}`
  - 配置：`knowledge.python_base_url`（默认 `http://localhost:9000`）、`knowledge.upload_dir`（默认 `./uploads/knowledge`）

- [ ] **Step 1: 配置扩展**

`backend/configs/config.go` 在相应结构体中新增（照现有 `App`/`Database` 字段风格）：

```go
	// Knowledge 知识库配置
	Knowledge struct {
		PythonBaseURL string `mapstructure:"python_base_url"`
		UploadDir     string `mapstructure:"upload_dir"`
	} `mapstructure:"knowledge"`
```

`backend/configs/config.yaml` 追加：

```yaml
knowledge:
  python_base_url: "http://localhost:9000"
  upload_dir: "./uploads/knowledge"
```

- [ ] **Step 2: 实现控制器**

`backend/controllers/knowledge.go`:

```go
package controllers

import (
	"net/http"
	"strconv"

	"aiops/internal/knowledge"
	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// KnowledgeController 运维知识库控制器
type KnowledgeController struct {
	DB      *gorm.DB
	Service *knowledge.Service
}

func NewKnowledgeController(db *gorm.DB, svc *knowledge.Service) *KnowledgeController {
	return &KnowledgeController{DB: db, Service: svc}
}

// List 文档列表
func (c *KnowledgeController) List(ctx *gin.Context) {
	var docs []models.KBDocument
	if err := c.DB.Order("created_at DESC").Find(&docs).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": docs})
}

// Upload 多文件上传，逐个落库并异步索引
func (c *KnowledgeController) Upload(ctx *gin.Context) {
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请使用 multipart/form-data 上传"})
		return
	}
	files := form.File["files"]
	if len(files) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "未接收到文件"})
		return
	}
	uploader := "unknown"
	if info, ok := ctx.Get("username"); ok {
		if name, ok2 := info.(string); ok2 {
			uploader = name
		}
	}
	created := make([]*models.KBDocument, 0, len(files))
	for _, fh := range files {
		src, err := fh.Open()
		if err != nil {
			continue
		}
		doc, err := c.Service.SaveUpload(ctx, fh.Filename, fh.Size, src, uploader)
		src.Close()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
			return
		}
		c.Service.IndexDocument(doc.ID)
		created = append(created, doc)
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": created})
}

// Delete 删除文档
func (c *KnowledgeController) Delete(ctx *gin.Context) {
	if err := c.Service.Delete(ctx, ctx.Param("id")); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

// Reindex 重新索引
func (c *KnowledgeController) Reindex(ctx *gin.Context) {
	if err := c.Service.Reindex(ctx, ctx.Param("id")); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "重新索引失败: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "已触发重新索引"})
}

// Retrieve 语义检索
func (c *KnowledgeController) Retrieve(ctx *gin.Context) {
	var req struct {
		Query string `json:"query" binding:"required"`
		TopK  int    `json:"top_k"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "query 不能为空"})
		return
	}
	if req.TopK <= 0 || req.TopK > 50 {
		req.TopK = 5
	}
	hits, err := c.Service.Retrieve(ctx, req.Query, req.TopK)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "检索失败: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": hits})
}

var _ = strconv.Itoa
```

（`uploader` 取 JWT 解析出的用户名——middleware.AuthRequired 需把 username 放进 ctx；若当前中间件未提供，先写死 `"unknown"` 并在 Task 12 前检查 `middleware/auth.go`，如缺则在中间件中 `ctx.Set("username", ...)`。）

- [ ] **Step 3: main.go 注册**

`backend/cmd/center/main.go` 中（notify-media 路由后）：

```go
	// 运维知识库
	kbClient := knowledge.NewClient(cfg.Knowledge.PythonBaseURL)
	kbSvc := knowledge.NewService(db, kbClient, cfg.Knowledge.UploadDir)
	kbCtrl := controllers.NewKnowledgeController(db, kbSvc)
	kbGroup := api.Group("/knowledge-base")
	{
		kbGroup.GET("", kbCtrl.List)
		kbGroup.POST("", kbCtrl.Upload)
		kbGroup.POST("/retrieve", kbCtrl.Retrieve)
		kbGroup.DELETE("/:id", kbCtrl.Delete)
		kbGroup.POST("/:id/reindex", kbCtrl.Reindex)
	}
```

import 增加 `"aiops/internal/knowledge"`。

- [ ] **Step 4: 构建验证**

```bash
cd backend && go build ./... && go vet ./...
# Expected: 通过
```

- [ ] **Step 5: Commit**

```bash
git add backend/controllers/knowledge.go backend/configs/ backend/cmd/center/main.go
git commit -m "feat(backend): 知识库 REST 接口与配置"
```

### Task 12: 对话 Agent 接入 search_knowledge_base 工具

**Files:**
- Modify: `backend/internal/agent/tools.go`
- Modify: `backend/internal/agent/agent.go`

**Interfaces:**
- Consumes: Task 10/11 的 `knowledge.Service`。
- Produces: 新工具 `search_knowledge_base(query string, top_k int) string`：query 必填，top_k 默认 5；返回 JSON 字符串（`[]ChunkHit`）。

- [ ] **Step 1: 修改 agent.go 防幻觉前缀**

`backend/internal/agent/agent.go:109` 附近，把：

```go
			if !strings.HasPrefix(tc.Function.Name, "query_") {
```

改为：

```go
			if !strings.HasPrefix(tc.Function.Name, "query_") && !strings.HasPrefix(tc.Function.Name, "search_") {
```

（强提醒文案 `agent.go:91` 处的描述同步加上知识库工具说明。）

- [ ] **Step 2: tools.go 注入知识库服务**

`backend/internal/agent/tools.go`：

```go
// Registry 工具注册中心
type Registry struct {
	ds  *datasource.Manager
	kb  *knowledge.Service
}

// NewRegistry 构造注册中心
func NewRegistry(db *gorm.DB) *Registry {
	return &Registry{ds: datasource.NewManager(db)}
}

// SetKnowledge 注入知识库编排服务（main.go 调用）
func (r *Registry) SetKnowledge(svc *knowledge.Service) {
	r.kb = svc
}
```

builders 数组追加：

```go
		{"search_knowledge_base", r.searchKnowledgeBaseTool},
```

新工具实现：

```go
// SearchKnowledgeBaseInput search_knowledge_base 工具入参
type SearchKnowledgeBaseInput struct {
	Query string `jsonschema:"required,description=检索问题或关键词，例如：MySQL 连接超时如何处理"`
	TopK  int    `jsonschema:"description=返回的片段数量，默认 5，最大 10"`
}

// SearchKnowledgeBaseOutput 单个命中片段
type SearchKnowledgeBaseOutput struct {
	ChunkID    string  `json:"chunk_id"`
	DocumentID string  `json:"document_id"`
	Title      string  `json:"title"`
	Content    string  `json:"content"`
	Score      float64 `json:"score"`
}

// searchKnowledgeBaseTool 检索运维知识库
func (r *Registry) searchKnowledgeBaseTool(ctx context.Context, input *SearchKnowledgeBaseInput) ([]*SearchKnowledgeBaseOutput, error) {
	if r.kb == nil {
		return nil, fmt.Errorf("知识库服务未初始化")
	}
	topK := input.TopK
	if topK <= 0 || topK > 10 {
		topK = 5
	}
	hits, err := r.kb.Retrieve(ctx, input.Query, topK)
	if err != nil {
		return nil, err
	}
	out := make([]*SearchKnowledgeBaseOutput, 0, len(hits))
	for _, h := range hits {
		out = append(out, &SearchKnowledgeBaseOutput{
			ChunkID: h.ChunkID, DocumentID: h.DocumentID, Title: h.Title,
			Content: h.Content, Score: h.Score,
		})
	}
	return out, nil
}
```

- [ ] **Step 3: main.go 注入**

`backend/cmd/center/main.go` 中 agent 组装处（`controllers/chat.go` 里 `agent.NewRegistry` 的位置——改为 main.go 创建 registry 并 `SetKnowledge` 后传入，或在 `NewChatController` 后调用 `chatCtrl.Registry().SetKnowledge(kbSvc)`；实现时以现有 `NewRegistry(db)` 调用点为锚点修改，保证 kb 服务在路由注册前注入）。

- [ ] **Step 4: 构建验证**

```bash
cd backend && go build ./... && go vet ./... && go test ./...
# Expected: 全部通过
```

- [ ] **Step 5: 端到端冒烟**

```bash
# 终端1
docker compose up -d qdrant && cd algorithm && uvicorn app.main:app --port 9000
# 终端2
cd backend && go run ./cmd/center
# 终端3：先登录拿 token，再上传文档、触发对话问知识库相关问题
curl -X POST http://localhost:8080/api/knowledge-base -H "Authorization: Bearer $TOKEN" -F files=@手册.md
# 期望：返回 pending 记录；数秒后状态变 indexed
```

- [ ] **Step 6: Commit**

```bash
git add backend/internal/agent/ backend/cmd/center/main.go backend/controllers/chat.go
git commit -m "feat(backend): 对话 Agent 接入知识库检索工具"
```

---

## Phase C：前端

### Task 13: 知识库接口层与页面去 Mock 化

**Files:**
- Create: `frontend/src/api/knowledge.js`
- Modify: `frontend/src/views/KnowledgeBaseView.vue`

**Interfaces:**
- Consumes: Task 11 REST 契约。
- Produces: `knowledgeApi.list() / upload(fileList) / remove(id) / reindex(id) / retrieve(query, topK)`。

- [ ] **Step 1: 接口层**

`frontend/src/api/knowledge.js`（照 `src/api/llmConfig.js` 的 createRequest 模式）:

```js
import { createRequest } from './request.js'

const request = createRequest('/api/knowledge-base')

export const knowledgeApi = {
  list: () => request.get(''),
  upload: (files) => {
    const form = new FormData()
    Array.from(files).forEach(f => form.append('files', f))
    return request.post('', form, { headers: { 'Content-Type': 'multipart/form-data' } })
  },
  remove: (id) => request.delete(`/${id}`),
  reindex: (id) => request.post(`/${id}/reindex`),
  retrieve: (query, topK = 5) => request.post('/retrieve', { query, top_k: topK })
}
```

- [ ] **Step 2: 页面改造要点**

`KnowledgeBaseView.vue`：
- 删除 `import { knowledgeBaseFiles } from '../mock/data.js'`，`files` 初始 `ref([])`，`onMounted` 调 `knowledgeApi.list()`。
- `handleUpload`：调 `knowledgeApi.upload(fileInput.files)` 成功后刷新列表；用 `formatSize(bytes)`（`bytes<1024*1024 ? (KB) : (MB)`）替代后端字符串。
- `reindex(id)` / `removeFile(id)` 调真实接口，删除前加 `confirm`。
- 轮询：列表中任一 `status === 'indexing' || status === 'pending'` 时每 2 秒重新 `list()`，`clearInterval` 在无进行中任务时停止。
- `uploadTime` 格式化：`new Date(t).toLocaleString('zh-CN', {hour12: false})`。
- 状态映射保持四态：`indexed/pending/indexing/failed`。

- [ ] **Step 3: 构建验证**

```bash
cd frontend && npm run build
# Expected: 通过
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/api/knowledge.js frontend/src/views/KnowledgeBaseView.vue
git commit -m "feat(frontend): 知识库页面接入真实接口"
```

---

## Phase D：文档收尾

### Task 14: 文档更新

**Files:**
- Modify: `README.md`
- Modify: `AGENTS.md`

- [ ] **Step 1: README 增加知识库启动说明**

```bash
# 运维知识库（可选组件）
docker compose up -d qdrant          # 向量库
cd algorithm && pip install -r requirements.txt
uvicorn app.main:app --port 9000     # Python 解析/向量化服务
# 并在「LLM 管理」中配置 model_type=embedding 的默认向量化模型
```

- [ ] **Step 2: AGENTS.md 同步**

- 项目结构树增加 `algorithm/` 目录说明；
- 核心机制增加知识库条目（上传→Python 解析向量化→Qdrant；Agent 工具 search_knowledge_base）；
- 构建与运行增加 Python 服务与 Qdrant 启动命令；
- 最后更新日期改为当天。

- [ ] **Step 3: Commit**

```bash
git add README.md AGENTS.md
git commit -m "docs: 知识库功能启动说明"
```

---

## 风险与决策记录

1. **向量维度与 collection**：不同 embedding 模型维度不同（768/1024/1536）。首次 index 时按实际维度建 collection；换模型需重建 collection（`docker compose down -v` 清 qdrant_data）。计划中 `ensure_collection` 已做维度校验。
2. **embedding 配置透传**：Python 不连数据库，embedding 密钥由 Go 从 `llm_config`（`model_type=embedding` 且默认）读取后随请求透传——避免 Python 侧再维护一套配置，也符合最小权限。
3. **kb_chunk 表只存元数据**：chunk 全文存在 Qdrant payload 与文件本身，MySQL 不双写大文本（避免两份事实源）；如后续需要 SQL 全文兜底检索再回填 content 列。
4. **解析失败降级**：pypdf 对扫描版 PDF 提取为空 → Python 返回 422，Go 置 `failed` 并记录原因，用户可换 txt/md 重新上传。
5. **文件存储**：一期本地磁盘 `upload_dir`；设计文档中的 MinIO 留到部署期再接入（文件读写已收敛在 `Service.saveFile`/Delete 两处，易替换）。

## Self-Review 记录

- **Spec 覆盖**：设计文档 8.2.5 的 6 个接口 → 本计划 Task 11 覆盖 5 个（documents 上传/列表/删除、retrieve），`feedback` 接口未纳入（文档标注属 Phase 4 运营期，YAGNI 暂缓）；RAG 助手引用规范 → Task 12 工具返回含 chunk_id/title/score，引用标注由 system prompt 约束（`internal/agent/agent.go` 的 buildSystemPrompt 已在既有链路中，Task 12 的文案调整已列步骤）。
- **Placeholder 扫描**：无 TBD/TODO；所有代码步骤含完整实现。
- **类型一致性**：`ChunkHit` 字段（ChunkID/DocumentID/Title/Content/Score）在 client.go、service.go、controllers、tools.go、前端 retrieve 中一致；Python `RetrieveRequest` 与 Go `Retrieve` payload 键名（query/top_k/embedding{base_url,api_key,model}）一致；multipart 字段名（document_id/title/file/base_url/api_key/model）Go 客户端与 Python 端点一致。
