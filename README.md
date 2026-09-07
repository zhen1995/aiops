# AIOPS 智能运维平台

面向大规模分布式系统的智能运维平台：融合 Prometheus 指标、ELK 日志、性能剖析等多源监控数据，结合大模型 Agent 实现智能问答、告警规则评估、告警事件、定时巡检报告、根因分析与运维知识库。

## 运维知识库（可选组件）

知识库为可选功能，启用需准备三个要素：**Qdrant 向量库**、**Python 解析/向量化服务**、**默认向量化模型配置**。

```bash
# 1. 启动 Qdrant 向量库（本机需安装 Docker；若无 Docker 请自行安装后重试）
docker compose up -d qdrant

# 2. 启动 Python 解析/向量化服务（端口 9000，Go 后端默认连接该地址）
cd algorithm
pip install -r requirements.txt
uvicorn app.main:app --port 9000

# 3. 在「LLM 管理」中配置一个 model_type=向量化模型 的配置并设为默认，
#    上传文档时 Go 后端会把密钥随请求透传给 Python 服务完成 embedding
```

启用后：知识库页面可上传文档（自动解析、切块、向量化入 Qdrant），对话 Agent 自动获得 `search_knowledge_base` 工具进行检索问答。

> 注意：不同向量化模型的向量维度不同（768/1024/1536），首次入库时按所用模型的实际维度自动建 collection；更换模型需重建 collection（`docker compose down -v` 清除 `qdrant_data` 后重新上传）。
