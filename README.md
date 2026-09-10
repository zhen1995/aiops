# AIOPS 智能运维平台

中文 | [English](README.en.md)

面向大规模分布式系统的智能运维平台：融合 Prometheus 指标、ELK 日志、性能剖析等多源监控数据，结合大模型（LLM）Agent 实现智能对话问答、告警规则评估、告警事件管理、告警降噪、定时巡检报告、根因分析与运维知识库。

## 功能特性

- **AI 对话助手**：基于大模型 Function Calling 循环调用只读数据源工具（Prometheus / ElasticSearch / Pyroscope），命中运维数据关键词时强制先查数据源，防止幻觉。
- **告警规则评估引擎**：内置引擎按「执行频率」用 PromQL 即时查询探测，每条时序序列（标签组合）独立计数，满足「持续时间」后自动产生 / 恢复告警事件，无需依赖 Nightingale。
- **告警降噪**：时间窗口聚合防抖与拓扑抑制（服务树祖先 firing 时抑制后代告警）两种策略，被拦截的记录单独落库可审计。
- **通知记录**：告警通知全链路（被拦截 / 未配规则 / 级别过滤 / 发送成功或失败）统一落库可追溯。
- **总览大盘**：KPI、原始 vs 降噪趋势、级别分布、服务健康度等多维度统计。
- **根因分析**：Eino Graph 工作流异步执行，先收集证据再综合分析，输出结构化结论（证据链 / 可信度 / 修复建议）。
- **日志分析**：待开发。
- **定时巡检**：Cron 调度 + LLM 生成 Markdown 报告，支持钉钉机器人 / Webhook 推送。
- **运维知识库**：文档解析、切块、向量化入 Qdrant，对话 Agent 可检索引用（可选组件）。
- **全局搜索 / 多语言（中、英）/ RBAC 权限**：用户、角色、权限种子化管理。
## 技术栈

| 层 | 技术 |
|----|------|
| 前端 | Vue 3、Vite、ECharts、Vue Router（hash 模式） |
| 后端 | Go、Gin、GORM、Viper、CloudWeGo Eino、robfig/cron、JWT |
| 算法服务（可选） | Python、FastAPI、Qdrant Client |
| 存储 | MySQL 8.0+、Qdrant 向量库（可选） |
| 监控数据源 | Prometheus、ElasticSearch、Pyroscope |

## 目录结构

```
aiops/
├── backend/                 # Go 后端服务
│   ├── cmd/center/          # 服务入口（路由、AutoMigrate、引擎/调度器启动）
│   ├── configs/             # Viper 配置（config.yaml）
│   ├── controllers/         # HTTP 控制器
│   ├── internal/            # agent / alerting / chat / datasource / inspection / rca / notify / knowledge
│   ├── middleware/          # JWT 认证中间件（SSE 支持 query token）
│   └── models/              # GORM 模型
├── frontend/                # Vue 3 + Vite 前端
│   ├── src/api/             # 接口封装层（JWT、401 跳登录）
│   ├── src/views/           # 页面视图
│   └── vite.config.js       # 含 /api → localhost:8080 开发代理
├── algorithm/               # Python 算法服务（知识库解析/向量化/检索，可选）
├── qdrant/                  # Qdrant 向量库部署文件（docker-compose.yml + config.yaml）
├── nginx.conf               # Nginx 生产部署参考配置
└── sql/                     # 数据库初始化脚本（含 root 用户种子）
```

## 快速开始

### 环境要求

| 依赖 | 版本 |
|------|------|
| Go | 1.26+ |
| MySQL | 8.0+ |
| Node.js | 18+ |
| Docker | 仅知识库功能需要 |
| Python | 3.10+，仅知识库功能需要 |

### 1. 初始化数据库

```bash
mysql -u root -p -e "CREATE DATABASE aiops DEFAULT CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;"
mysql -u root -p aiops < sql/init.sql
```

`sql/init.sql` 会创建全部表结构，并种子化默认用户、角色、权限、通知模板与降噪策略。

> ⚠️ 默认管理员账号为 `root / root`（明文存储），首次登录后请立即在「用户管理」中修改密码。

### 2. 配置并启动后端

编辑 `backend/configs/config.yaml`，把 `database.dsn` 改为你的 MySQL 连接串：

```yaml
database:
  dsn: "root:你的密码@tcp(localhost:3306)/aiops?charset=utf8mb4&parseTime=True&loc=Local"
server:
  port: ":8080"
app:
  # 前端访问地址，用于巡检报告通知中的「完整报告」链接
  frontend_base_url: "http://localhost:5173"
knowledge:
  python_base_url: "http://localhost:9000"
  upload_dir: "./uploads/knowledge"
```

也可以通过环境变量覆盖配置（生产环境推荐，避免明文写入仓库）：

| 环境变量 | 说明 | 默认值 |
|----------|------|--------|
| `AIOPS_DATABASE_DSN` / `DATABASE_DSN` | MySQL 连接串 | 见 config.yaml |
| `AIOPS_SERVER_PORT` / `SERVER_PORT` | 服务端口 | `:8080` |
| `AIOPS_APP_FRONTEND_BASE_URL` / `FRONTEND_BASE_URL` | 前端访问地址 | `http://localhost:5173` |

启动：

```bash
cd backend
go mod tidy
go run ./cmd/center
```

启动时自动执行 GORM AutoMigrate（建表 / 加列）并回填告警规则默认执行频率。

### 3. 启动前端

```bash
cd frontend
npm install
npm run dev        # 开发服务器 http://localhost:5173，/api 代理到 http://localhost:8080
```

浏览器访问 http://localhost:5173 ，使用 `root / root` 登录。

### 4. 配置大模型

登录后进入「LLM 配置」页面，添加 OpenAI 兼容协议的对话模型配置并设为默认（用于对话助手、根因分析、巡检报告）。无需重启，配置即时生效。

## 可选：启用运维知识库

知识库需要三个组件：**Qdrant 向量库**、**Python 解析/向量化服务**、**默认向量化模型配置**。

```bash
# 1. 启动 Qdrant（配置文件见 qdrant/config.yaml，可访问 http://<ip>:6333/dashboard）
cd qdrant
docker compose up -d

# 2. 启动 Python 服务（端口 9000，Go 后端默认连接该地址）
cd ../algorithm
pip install -r requirements.txt
uvicorn app.main:app --port 9000

# 3. 在「LLM 配置」中添加 model_type=向量化模型 的配置并设为默认
```

向量化模型必须提供 **Embeddings API**（OpenAI 兼容协议均可）。推荐：

| 服务商 | Base URL | 模型 | 说明 |
|--------|----------|------|------|
| 阿里云百炼 | `https://dashscope.aliyuncs.com/compatible-mode/v1` | `text-embedding-v3` | 国内直连，中文效果好，按量计费 |
| SiliconFlow | `https://api.siliconflow.cn/v1` | `BAAI/bge-m3` | 个人开发者有免费额度 |
| OpenAI | `https://api.openai.com/v1` | `text-embedding-3-small` | 需可访问 OpenAI 的网络与密钥 |

> 注意：**DeepSeek 官方 API 不提供 Embeddings 端点**，Kimi 编程订阅密钥同样不能用于向量化，配置错误时文档索引会失败（错误原因记录在 `kb_document.error_msg`）。不同向量化模型的向量维度不同（768/1024/1536），首次入库时按所用模型的实际维度自动建 collection；更换模型需清空 Qdrant 数据后重新上传文档。

## 生产部署

1. **前端构建**：`cd frontend && npm run build`，将 `frontend/dist/` 内容发布到 Nginx 站点目录。
2. **后端构建**：`cd backend && go build -o center ./cmd/center`，以环境变量方式注入配置后运行二进制。
3. **Nginx**：参考仓库根目录 `nginx.conf`：

   ```nginx
   root /opt/aiops/web;                    # 前端 dist 所在目录
   location /api/ {
       proxy_pass http://127.0.0.1:8080;   # 后端服务
       proxy_buffering off;                # SSE 流式响应必需
       proxy_read_timeout 600s;
       proxy_set_header X-Accel-Buffering no;
       # ... 其余见 nginx.conf
   }
   ```

   SSE（对话流、根因分析轮询）依赖 `proxy_buffering off` 与 `X-Accel-Buffering no`，请勿省略。
4. **知识库（可选）**：服务器上部署 Qdrant（`cd qdrant && docker compose up -d`）与 Python 服务（uvicorn），注意 `qdrant/docker-compose.yml` 中数据卷挂载路径（`/data/qdrant:/qdrant/storage`）需按你的服务器实际目录调整。
5. **前端地址**：通过 `AIOPS_APP_FRONTEND_BASE_URL` 配置真实前端地址，保证巡检报告通知中的链接可访问。

## 测试

```bash
cd algorithm && pytest      # Python 算法服务单元测试
cd backend && go test ./... # Go 测试（如补充）
```

## 安全说明

- `backend/configs/config.yaml` 默认使用占位 DSN，**生产环境请通过环境变量注入数据库密码**，不要把真实凭证提交到仓库。
- LLM 配置的 API Key 存储在数据库中，接口返回已做脱敏处理。
- 默认账号 `root / root` 仅供首次启动使用，请立即修改。
- JWT token 存于前端 localStorage，生产部署建议启用 HTTPS。

## License

[Apache License 2.0](LICENSE)
