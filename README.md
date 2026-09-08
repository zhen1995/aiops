# AIOPS 智能运维平台

面向大规模分布式系统的智能运维平台：融合 Prometheus 指标、ELK 日志、性能剖析等多源监控数据，结合大模型 Agent 实现智能问答、告警规则评估、告警事件、定时巡检报告、根因分析与运维知识库。

## 运维知识库（可选组件）

知识库为可选功能，启用需准备三个要素：**Qdrant 向量库**、**Python 解析/向量化服务**、**默认向量化模型配置**。

```bash
# 1. 启动 Qdrant 向量库（本机需安装 Docker；若无 Docker 请自行安装后重试）
docker compose up -d qdrant

## 配置文件
https://github.com/qdrant/qdrant/blob/master/config/config.yaml


访问 http://<ip>:6333/dashboard查看 qdrant服务

# 2. 启动 Python 解析/向量化服务（端口 9000，Go 后端默认连接该地址）
cd algorithm
pip install -r requirements.txt
uvicorn app.main:app --port 9000

# 3. 在「LLM 管理」中配置一个 model_type=向量化模型 的配置并设为默认，
#    上传文档时 Go 后端会把密钥随请求透传给 Python 服务完成 embedding
```

向量化模型需选择**提供 Embeddings API** 的服务商（OpenAI 兼容协议均可）。推荐：

| 服务商 | Base URL | 模型 | 说明 |
|--------|----------|------|------|
| 阿里云百炼 | `https://dashscope.aliyuncs.com/compatible-mode/v1` | `text-embedding-v3` | 国内直连，中文效果好，按量计费 |
| SiliconFlow | `https://api.siliconflow.cn/v1` | `BAAI/bge-m3` | 个人开发者有免费额度 |
| OpenAI | `https://api.openai.com/v1` | `text-embedding-3-small` | 需可访问 OpenAI 的网络与密钥 |

> 注意：**DeepSeek 官方 API 不提供 Embeddings 端点**（仅 deepseek-v4-flash / deepseek-v4-pro 等对话模型），配置为向量化模型会导致索引报 404；Kimi 编程订阅（`api.kimi.com/coding/v1`）的密钥同样不能用于向量化。配置错误时文档会索引失败，错误原因记录在 `kb_document.error_msg`。

启用后：知识库页面可上传文档（自动解析、切块、向量化入 Qdrant），对话 Agent 自动获得 `search_knowledge_base` 工具进行检索问答。

> 注意：不同向量化模型的向量维度不同（768/1024/1536），首次入库时按所用模型的实际维度自动建 collection；更换模型需重建 collection（`docker compose down -v` 清除 `qdrant_data` 后重新上传）。

## 生产部署

生产环境（Go 后端 + Python 算法服务 + Vue 前端 + Qdrant + MySQL + Nginx）的完整部署步骤、systemd/Nginx 配置样例、启动顺序、验证清单与安全注意事项，见 [docs/生产部署文档.md](docs/生产部署文档.md)。
