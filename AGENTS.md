# AIOPS 智能运维平台 - 代理开发指南

## 项目概述

本项目是一个面向大规模分布式系统的 **AIOPS 智能运维平台** 雏形，目标是通过融合 Prometheus 指标、ELK 日志、Nightingale 告警等多源监控数据，结合机器学习与深度学习算法，实现异常检测、根因分析、告警降噪、运维知识库与智能运维助手等能力。

当前仓库包含：

- **frontend/**：Vue 3 + Vite 前端原型，使用纯 Mock 数据，暂未接入真实后端 API。
- **backend/**：Go 后端服务雏形，基于 Gin + GORM + MySQL，目前仅实现了配置加载、数据库连接与基础登录接口示例。
- **docs/**：系统设计文档（`AIOPS系统设计文档.md`）与前端原型 SPEC（`SPEC.md`）。

> 注意：设计文档中描述的 Python 算法引擎层（FastAPI / Celery / 异常检测 / 根因分析 / 知识库服务等）目前**尚未在仓库中实现**，仅作为架构设计参考。

项目采用 **Apache License 2.0**。

---

## 技术栈

### 前端

| 技术 | 版本/说明 | 用途 |
|------|----------|------|
| Vue | 3.4+ | 组件化 UI 框架 |
| Vue Router | 4.3+ | 前端路由 |
| Vite | 5.2+ | 构建工具与开发服务器 |
| ECharts | 5.5+ | 数据可视化图表 |
| CSS 变量 | 自定义主题 | 统一配色与组件样式 |

### 后端

| 技术 | 版本/说明 | 用途 |
|------|----------|------|
| Go | 1.26.4 | 服务端语言 |
| Gin | v1.12.0 | Web 框架 |
| GORM | v1.31.2 | ORM 框架 |
| MySQL 驱动 | v1.6.0 / v1.10.0 | 关系型数据库 |
| Viper | v1.21.0 | 配置文件与环境变量管理 |
| CloudWeGo Eino | v0.9.14 | AI 应用框架（已引入 deepseek 扩展） |

### 设计文档提及但尚未落地的组件

- Python + FastAPI 算法服务
- VictoriaMetrics / Elasticsearch / Kafka / Redis / 向量数据库
- Celery 任务调度
- Kubernetes / Istio / Grafana / Jaeger

---

## 项目结构

```
E:/aiops
├── backend/
│   ├── cmd/center/main.go          # 后端服务入口
│   ├── configs/
│   │   ├── config.go               # Viper 配置加载逻辑
│   │   └── config.yaml             # 默认配置文件（含数据库 DSN）
│   ├── controllers/
│   │   └── login.go                # /login 示例接口
│   ├── internal/deepseekchat/      # Eino + DeepSeek 调用示例
│   ├── models/
│   │   └── sys_user.go             # 用户模型（GORM）
│   ├── go.mod / go.sum             # Go 依赖
│   └── LeanGo.md                   # Go 依赖管理备忘
├── frontend/
│   ├── src/
│   │   ├── App.vue / main.js       # 应用入口
│   │   ├── router/index.js         # 路由配置
│   │   ├── layout/AppLayout.vue    # 侧边栏 + 顶部栏布局
│   │   ├── components/             # 共享组件（StatCard/LevelTag/ChartBox/PageHeader）
│   │   ├── views/                  # 页面视图
│   │   ├── mock/data.js            # 全部 Mock 数据
│   │   └── styles/theme.css        # 主题变量与通用样式
│   ├── index.html
│   ├── package.json
│   ├── vite.config.js
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
npm run dev        # 开发服务器
npm run build      # 生产构建（输出到 frontend/dist）
npm run preview    # 预览构建产物
```

- 前端当前使用 `base: './'`（相对路径），构建产物可直接用静态服务器托管。
- 前端没有代理后端 API 的配置，当前所有数据来自 `src/mock/data.js`。

### 后端

```bash
cd backend
go mod tidy
go build ./cmd/center
go run ./cmd/center
```

默认监听端口 `:8080`，可通过环境变量或配置文件修改：

- `AIOPS_DATABASE_DSN` / `DATABASE_DSN`：MySQL 连接串
- `AIOPS_SERVER_PORT` / `SERVER_PORT`：服务端口（例如 `:8080`）

配置文件路径由 `cmd/center/main.go` 中的 `getConfigPath()` 自动推导，期望 `config.yaml` 位于 `backend/configs/config.yaml`。

---

## 代码组织

### 后端约定

- 入口：`cmd/center/main.go`
- 配置：`configs/`（Viper + YAML + 环境变量）
- 控制器/Handler：`controllers/`
- 数据模型：`models/`
- 内部工具/示例：`internal/`
- 依赖管理：使用 `go mod tidy` 与 `go get`；不要手动编辑 `go.mod`（参考 `backend/LeanGo.md`）。

### 前端约定

- 所有页面在 `src/views/` 下，按模块分子目录（如 `ai-config/`、`notification/`、`org/`）。
- 共享组件在 `src/components/`。
- 样式统一使用 `src/styles/theme.css` 中的 CSS 变量，禁止硬编码颜色。
- Mock 数据集中在 `src/mock/data.js`。
- 路由使用 `createWebHashHistory`，默认重定向到 `/chat`。
- 页面标题通过路由 `meta.title` 驱动。

---

## 测试

当前仓库**没有自动化测试**（未找到 `*_test.go`、Jest、Vitest 或 Playwright 等测试配置）。

建议后续补充：

- 后端：`go test ./...` 单元测试与集成测试。
- 前端：Vitest 或 Jest 单元测试，Playwright 端到端测试。

---

## 安全与敏感信息

- `backend/configs/config.yaml` 中当前包含明文数据库密码（`root:a.123456R+=@...`）。**生产环境请通过环境变量注入**，不要提交真实凭证。
- 前端 Mock 数据中的 API Key（`sk-****4f2a` 等）为虚构值，不具实际意义。
- DeepSeek / LLM 调用示例从环境变量读取 `DEEPSEEK_API_KEY` 与 `DEEPSEEK_BASE_URL`，符合密钥不硬编码的原则。

---

## 部署

当前没有 Dockerfile、docker-compose、CI/CD 工作流或 Kubernetes 清单文件。部署为手动阶段：

1. 前端：`npm run build` 后托管 `frontend/dist/`。
2. 后端：`go build ./cmd/center` 得到二进制，配合环境变量运行。

设计文档中描述的容器化/K8s 部署方案尚未落地。

---

## 开发注意事项

1. **前端与后端尚未联调**：所有视图数据来自 Mock，新增真实 API 时需要同步修改前端 `src/mock/data.js` 或对接口层进行替换。
2. **后端接口极少**：目前仅 `/login` 一个示例接口，业务接口（告警、异常检测、根因分析等）待实现。
3. **算法引擎未实现**：异常检测、日志聚类、知识库向量化等算法服务目前只存在于设计文档。
4. **中文优先**：项目文档、注释、界面文本均使用中文。代码标识符保持英文，用户可见文本使用中文。
5. **Go 依赖管理**：新增依赖后先 `go get xxx`，再 `go mod tidy`。
6. **构建检查**：提交前请确保 `npm run build` 与 `go build ./cmd/center` 均通过。

---

## 关键文件速查

| 文件 | 说明 |
|------|------|
| `backend/cmd/center/main.go` | 后端入口 |
| `backend/configs/config.yaml` | 后端默认配置 |
| `backend/controllers/login.go` | 登录接口示例 |
| `frontend/src/router/index.js` | 前端路由 |
| `frontend/src/layout/AppLayout.vue` | 整体布局 |
| `frontend/src/mock/data.js` | 全部 Mock 数据 |
| `frontend/src/styles/theme.css` | 主题变量 |
| `docs/AIOPS系统设计文档.md` | 系统架构设计 |
| `docs/SPEC.md` | 前端原型 SPEC |

---

*最后更新：2026-08-18 — 基于仓库当前实际内容整理。*
