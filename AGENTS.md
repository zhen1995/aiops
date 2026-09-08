# AIOPS 智能运维平台 - 代理开发指南

## 项目概述

本项目是一个面向大规模分布式系统的 **AIOPS 智能运维平台**，通过融合 Prometheus 指标、ELK 日志、性能剖析等多源监控数据，结合大模型（LLM）Agent 能力，实现智能对话问答、告警规则评估、告警事件生成、定时巡检报告、运维知识库等能力。

当前仓库包含：

- **frontend/**：Vue 3 + Vite 前端，已通过 `src/api/` 接口层对接真实后端（开发环境由 Vite proxy 转发 `/api`）；总览大盘、告警降噪、日志分析等演示页面仍使用 `src/mock/data.js` 的 Mock 数据，知识库页面已接入真实接口。
- **backend/**：Go 后端服务，基于 Gin + GORM + MySQL + Viper + CloudWeGo Eino，已实现认证、对话 Agent、LLM 配置、数据源管理、告警规则评估引擎、告警事件、根因分析、定时巡检、运维知识库、通知媒介、用户与角色权限等完整接口。
- **algorithm/**：Python 算法服务（FastAPI，入口 `app.main:app`），目前承载运维知识库的文档解析、切块、向量化与向量检索，依赖 Qdrant 向量库（仓库根 `docker-compose.yml` 编排）。设计文档中的异常检测、日志聚类等其他算法能力尚未实现。
- **docs/**：系统设计文档（`AIOPS系统设计文档.md`）与前端原型 SPEC（`SPEC.md`）。

> 注意：告警事件不再依赖 Nightingale，由系统内置评估引擎（`backend/internal/alerting`）根据告警规则自动生成；Nightingale 客户端仅保留在「告警引擎配置」的连通性测试中。

项目采用 **Apache License 2.0**。

---

## 技术栈

### 前端

| 技术 | 版本/说明 | 用途 |
|------|----------|------|
| Vue | 3.4+ | 组件化 UI 框架 |
| Vue Router | 4.3+ | 前端路由（hash 模式） |
| Vite | 5.2+ | 构建工具与开发服务器（含 `/api` 代理） |
| ECharts | 5.5+ | 数据可视化图表 |
| CSS 变量 | 自定义主题 | 统一配色与组件样式 |

### 后端

| 技术 | 版本/说明 | 用途 |
|------|----------|------|
| Go | 1.26.4 | 服务端语言 |
| Gin | v1.12.0 | Web 框架 |
| GORM | v1.31.2 | ORM 框架（启动时 AutoMigrate 自动建表） |
| MySQL 驱动 | v1.6.0 | 关系型数据库 |
| Viper | v1.21.0 | 配置文件与环境变量管理 |
| CloudWeGo Eino | v0.9.14 | LLM Agent 框架（Function Calling） |
| eino-ext openai | v0.1.13 | OpenAI 兼容协议的 ChatModel 客户端 |
| robfig/cron | v3.0.1 | 巡检任务 Cron 调度 |
| google/uuid | v1.6.0 | 业务主键 UUID |

### Python 算法服务（可选，运维知识库）

| 技术 | 说明 |
|------|------|
| Python + FastAPI | 知识库文档解析/切块/向量化/检索服务（`algorithm/`，uvicorn 启动，端口 9000） |
| pypdf 等 | 文档解析（扫描版 PDF 解析为空时返回 422，Go 侧置 failed） |
| Qdrant | 向量库（v1.11.0，仓库根 docker-compose 编排，端口 6333/6334） |
| pytest | Python 测试（`algorithm/tests/`） |

---

## 项目结构

```
E:/aiops
├── backend/
│   ├── cmd/center/main.go          # 后端服务入口（路由注册、AutoMigrate、调度器/引擎启动）
│   ├── configs/
│   │   ├── config.go               # Viper 配置加载逻辑
│   │   └── config.yaml             # 默认配置文件（数据库 DSN、端口、前端地址）
│   ├── controllers/                # HTTP 控制器：auth/chat/llm_config/datasource/
│   │                               #   alert_engine/alert_rule/alert_event/root_cause/
│   │                               #   inspection/notify_media/user/role/login
│   ├── middleware/auth.go          # JWT 认证中间件（SSE 支持 query token）
│   ├── internal/
│   │   ├── agent/                  # 公共 LLM Agent：工具注册表 + Function Calling 循环
│   │   │   ├── tools.go            #   query_prometheus / query_elasticsearch /
│   │   │   │                       #   query_pyroscope / list_data_sources
│   │   │   ├── agent.go            #   通用 Agent（Options/Callbacks 配置化）
│   │   │   └── intent.go           #   数据源关键词防幻觉校验
│   │   ├── alerting/engine.go      # 告警规则评估引擎（执行频率探测 PromQL → 告警/恢复事件）
│   │   ├── chat/                   # 对话会话管理、LLM 客户端（对话 Agent 经 internal/agent 实现）
│   │   ├── datasource/             # Prometheus/ElasticSearch/Pyroscope 只读查询客户端
│   │   ├── inspection/             # 巡检调度器（robfig/cron）+ 报告生成执行器
│   │   ├── nightingale/            # Nightingale 客户端（仅告警引擎配置测试使用）
│   │   ├── rca/                    # 根因分析：Eino Graph 工作流（证据收集/综合分析）+ 异步 runner
│   │   └── notify/                 # 通知推送（钉钉机器人 / Webhook）
│   ├── models/                     # GORM 模型：用户/角色/权限/LLM配置/数据源/告警引擎配置/
│   │                               #   告警规则/告警事件/根因分析/巡检任务/巡检报告/通知媒介/会话消息
│   ├── go.mod / go.sum             # Go 依赖
│   └── LeanGo.md                   # Go 依赖管理备忘
├── algorithm/                      # Python 算法服务（FastAPI，运维知识库解析/向量化/检索）
│   ├── app/
│   │   ├── main.py                 # FastAPI 入口（uvicorn app.main:app）
│   │   ├── routers/knowledge.py    # /api/v1/knowledge：index / retrieve / documents
│   │   ├── parser.py / chunker.py  # 文档解析（pypdf 等）与文本切块
│   │   ├── embedder.py / store.py  # embedding 调用与 Qdrant 向量存储
│   │   └── config.py / schemas.py
│   ├── tests/                      # pytest 测试
│   ├── pytest.ini
│   └── requirements.txt
├── docker-compose.yml              # Qdrant 向量库编排（v1.11.0，端口 6333/6334）
├── frontend/
│   ├── src/
│   │   ├── App.vue / main.js       # 应用入口
│   │   ├── router/index.js         # 路由配置（createWebHashHistory，默认重定向 /chat）
│   │   ├── layout/AppLayout.vue    # 侧边栏 + 顶部栏布局
│   │   ├── api/                    # 接口层（request.js 封装 fetch + JWT，按模块分文件）
│   │   ├── components/             # 共享组件（StatCard/LevelTag/ChartBox/PageHeader 等）
│   │   ├── views/                  # 页面视图（含 ai-config/、notification/、org/ 子目录）
│   │   ├── mock/data.js            # 演示页面 Mock 数据（大盘/降噪/日志）
│   │   ├── utils/auth.js           # token 存取与请求头
│   │   └── styles/theme.css        # 主题变量与通用样式
│   ├── index.html
│   ├── package.json
│   ├── vite.config.js              # 含 /api → localhost:8080 开发代理
│   └── dist/                       # 构建产物（由 npm run build 生成）
├── docs/
│   ├── AIOPS系统设计文档.md          # 系统架构与功能设计
│   └── SPEC.md                     # 前端原型 SPEC
├── README.md
├── LICENSE
└── AGENTS.md                       # 本文件
```

---

## 构建与运行

### 前端

```bash
cd frontend
npm install
npm run dev        # 开发服务器（/api 代理到 http://localhost:8080）
npm run build      # 生产构建（输出到 frontend/dist，base 为相对路径 './'）
npm run preview    # 预览构建产物
```

### 后端

```bash
cd backend
go mod tidy
go build ./cmd/center
go run ./cmd/center
```

默认监听端口 `:8080`，配置文件为 `backend/configs/config.yaml`，可用环境变量覆盖：

- `AIOPS_DATABASE_DSN` / `DATABASE_DSN`：MySQL 连接串
- `AIOPS_SERVER_PORT` / `SERVER_PORT`：服务端口（例如 `:8080`）
- `AIOPS_APP_FRONTEND_BASE_URL` / `FRONTEND_BASE_URL`：前端访问地址（用于巡检报告通知中的「完整报告」链接，默认 `http://localhost:5173`）

启动时会自动执行 GORM AutoMigrate 建表/加列，并回填告警规则的默认执行频率（30 秒）。

### Python 算法服务与 Qdrant（运维知识库，可选）

```bash
docker compose up -d qdrant        # 启动 Qdrant 向量库（本机需安装 Docker）
cd algorithm
pip install -r requirements.txt
uvicorn app.main:app --port 9000   # Python 解析/向量化/检索服务
pytest                             # 运行 Python 测试（venv 见 algorithm/.venv）
```

知识库为可选组件，还需在「LLM 管理」配置 model_type=向量化模型 的配置并设为默认；Go 后端默认连接 `http://localhost:9000`（`configs/config.yaml` 的 `knowledge.python_base_url`）。

---

## 核心机制

- **对话 Agent**：`GET /api/chat/sessions/:id/stream`（SSE）→ `controllers/chat.go` 组装 `internal/agent.Agent`，通过 Function Calling 循环调用只读数据源工具回答，命中运维数据关键词时强制先查数据源（防幻觉）。
- **告警评估引擎**：`internal/alerting.Engine` 为每条启用规则按「执行频率」起 Ticker，用 PromQL 即时查询探测；**每条满足条件的时序序列（标签组合）独立计数**：持续命中达到「持续时间」为该序列创建 `alert` 事件（firing，复用同标签未恢复事件防止重启重复告警），持续未命中达到「持续时间」将该序列事件置 `resolved` 并创建 `recovery` 事件；「持续时间」为 0 时表示命中/未命中一次即触发告警/恢复。规则 CRUD/启停/删除实时联动引擎。
- **告警事件查询**：`GET /api/alert-events?scope=active|history` 查询本地 `alert_events` 表；active = 未恢复，history = 已恢复（含恢复事件）；`DELETE /api/alert-events/:id` 删除单条事件，删除未恢复的告警事件会联动引擎清理对应序列状态（条件仍满足时可重新触发）。不再查询 Nightingale。
- **巡检**：`internal/inspection.Scheduler`（robfig/cron，支持 5/6 段表达式）定时触发 `Executor`，通过公共 Agent 生成 Markdown 报告并落库，按任务配置的通知媒介推送（钉钉/Webhook）。
- **根因分析**：`POST /api/alert-events/:id/root-cause` 触发后由 `internal/rca.StartAnalysis` 起后台 goroutine 异步执行（10 分钟超时，状态 running/completed/failed 落库 `root_cause_analyses` 表）；Eino compose Graph 工作流先经 collect 证据收集节点（公共 Agent 调用 Prometheus/ElasticSearch/Pyroscope 只读工具），再经 synthesize 综合分析节点由大模型输出严格 JSON（含修复重试）；`GET /api/root-cause-analyses(:id)` 查询列表/详情，前端 `/rca` 页面对 running 状态轮询展示结构化结果（证据链/可信度/影响范围/修复建议）。
- **运维知识库**：Go 侧 `controllers/knowledge.go` + `internal/knowledge`（编排、文件落盘 `knowledge.upload_dir`、Qdrant/文档元数据维护，REST 前缀 `/api/knowledge-base`）调用 Python 服务（`algorithm/`，FastAPI，`/api/v1/knowledge` 的 index/retrieve/documents）完成文档解析、切块、向量化（embedding 密钥由 Go 从默认的向量化 LLM 配置读取后随请求透传）与向量检索，向量库存 Qdrant（chunk 全文存 Qdrant payload，`kb_chunk` 表只存元数据）；对话 Agent 通过 `search_knowledge_base` 工具（`internal/agent/tools.go`）检索知识库问答，命中结果含 chunk_id/title/score 供引用标注。不同 embedding 模型维度不同，换模型需重建 Qdrant collection。

---

## 代码组织

### 后端约定

- 入口：`cmd/center/main.go`（路由、AutoMigrate、引擎/调度器启动都在此处）。
- 配置：`configs/`（Viper + YAML + 环境变量）。
- 控制器/Handler：`controllers/`（每模块一个文件，构造函数注入 `*gorm.DB`）。
- 数据模型：`models/`（表名显式 `TableName()`，UUID 主键在 `BeforeCreate` 生成）。
- 内部能力：`internal/`（agent / alerting / chat / datasource / inspection / nightingale / rca / notify）。
- 依赖管理：使用 `go get xxx` 后 `go mod tidy`，不要手动编辑 `go.mod`（参考 `backend/LeanGo.md`）。

### 前端约定

- 所有页面在 `src/views/` 下，按模块分子目录（如 `ai-config/`、`notification/`、`org/`）。
- 业务请求统一走 `src/api/`（基于 `request.js` 的 `createRequest(baseURL)`），不要在视图里裸写 fetch。
- 共享组件在 `src/components/`；样式统一使用 `src/styles/theme.css` 中的 CSS 变量，禁止硬编码颜色。
- 路由使用 `createWebHashHistory`，默认重定向到 `/chat`；页面标题通过路由 `meta.title` 驱动。
- 新页面若后端接口未就绪，可临时使用 `src/mock/data.js`，接入真实接口后移除 Mock 依赖。

---

## 测试

当前仓库 Go/前端**没有自动化测试**（未找到 `*_test.go`、Jest、Vitest 或 Playwright 等测试配置）；Python 算法服务有 pytest 测试（`algorithm/tests/`，运行 `cd algorithm && pytest`）。

建议后续补充：

- 后端：`go test ./...` 单元测试与集成测试（告警引擎状态机、Agent 循环适合优先覆盖）。
- 前端：Vitest 或 Jest 单元测试，Playwright 端到端测试。

---

## 安全与敏感信息

- `backend/configs/config.yaml` 中当前包含明文数据库密码。**生产环境请通过环境变量注入**，不要提交真实凭证。
- LLM 配置的 API Key 存储在数据库 `llm_configs` 表中，接口返回时需注意脱敏（当前实现请留意 `controllers/llm_config.go`）。
- 前端 Mock 数据中的 API Key（`sk-****4f2a` 等）为虚构值，不具实际意义。
- 认证为 JWT（`Authorization: Bearer <token>`），SSE 场景允许 `?token=` 传递；token 存于前端 `localStorage`。

---

## 部署

当前没有 Dockerfile、CI/CD 工作流或 Kubernetes 清单文件（仓库根 `docker-compose.yml` 仅编排 Qdrant 向量库）。部署为手动阶段：

1. 前端：`npm run build` 后托管 `frontend/dist/`。
2. 后端：`go build ./cmd/center` 得到二进制，配合环境变量运行。
3. 知识库（可选）：部署 Qdrant（`docker compose up -d qdrant`）与 Python 服务（`uvicorn app.main:app --port 9000`），并配置默认向量化模型。
4. 部署后通过 `AIOPS_APP_FRONTEND_BASE_URL` 配置真实前端地址，保证巡检报告通知中的链接可访问。

---

## 开发注意事项

1. **前后端已联调**：登录、对话、LLM 配置、数据源、告警规则、告警事件、根因分析、巡检、运维知识库、通知媒介、用户/角色均已接真实 API；总览大盘、告警降噪、日志分析仍为 Mock 演示页。
2. **修改告警规则字段时需同步**：模型（`models/alert_rule.go`）、校验（`controllers/alert_rule.go validateRule`）、评估引擎（`internal/alerting/engine.go`）、前端表单（`AlertRuleView.vue`）四处。
3. **新增数据源工具**：在 `internal/agent/tools.go` 的 `ToolMap` 注册，Agent system prompt 会自动带上工具说明。
4. **知识库接口契约需双端同步**：新增/变更知识库接口时，需同时更新 Python 服务端点（`algorithm/app/routers/knowledge.py`）与 Go 客户端（`backend/internal/knowledge/client.go`）的 payload 键名、multipart 字段名及返回结构；Python 侧改动后运行 `cd algorithm && pytest` 验证。
5. **中文优先**：项目文档、注释、界面文本均使用中文。代码标识符保持英文，用户可见文本使用中文。
6. **构建检查**：提交前请确保 `npm run build` 与 `go build ./cmd/center` 均通过。

---

## 关键文件速查

| 文件 | 说明 |
|------|------|
| `backend/cmd/center/main.go` | 后端入口（路由/迁移/引擎启动） |
| `backend/configs/config.yaml` | 后端默认配置 |
| `backend/internal/agent/agent.go` | 公共 LLM Agent 执行循环 |
| `backend/internal/agent/tools.go` | Agent 数据源工具集 |
| `backend/internal/alerting/engine.go` | 告警规则评估引擎 |
| `backend/internal/inspection/scheduler.go` | 巡检 Cron 调度器 |
| `backend/internal/inspection/executor.go` | 巡检报告生成与通知 |
| `backend/internal/rca/runner.go` | 根因分析异步执行器（后台 goroutine、超时控制、状态落库） |
| `backend/internal/knowledge/` | 知识库编排服务（Go↔Python 客户端、文档/块元数据、Qdrant 交互） |
| `backend/controllers/knowledge.go` | 知识库 REST 接口（`/api/knowledge-base`） |
| `algorithm/app/routers/knowledge.py` | Python 知识库服务（解析/向量化/检索，`/api/v1/knowledge`） |
| `docker-compose.yml` | Qdrant 向量库编排 |
| `frontend/src/router/index.js` | 前端路由 |
| `frontend/src/layout/AppLayout.vue` | 整体布局 |
| `frontend/src/api/request.js` | 接口层封装（JWT、401 跳登录） |
| `frontend/src/views/AlertRuleView.vue` | 告警规则配置页 |
| `frontend/src/views/AnomalyView.vue` | 告警事件页（含「根因分析」触发入口） |
| `frontend/src/views/RcaView.vue` | 根因分析页（分析历史列表 + 结构化结果展示，running 状态轮询） |
| `frontend/src/mock/data.js` | 演示页面 Mock 数据 |
| `frontend/src/styles/theme.css` | 主题变量 |
| `docs/AIOPS系统设计文档.md` | 系统架构设计 |
| `docs/SPEC.md` | 前端原型 SPEC |

---

*最后更新：2026-09-04 — 补充运维知识库功能：`algorithm/` Python 服务（FastAPI，解析/切块/向量化/检索）、根 `docker-compose.yml` Qdrant 编排、Go 侧 `internal/knowledge` + `controllers/knowledge_base.go`（`/api/knowledge-base`）、对话 Agent `search_knowledge_base` 工具、前端知识库页面接入真实接口；README 增加知识库（可选组件）启动说明。*
