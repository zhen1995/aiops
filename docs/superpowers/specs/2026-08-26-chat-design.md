# AIOPS 智能运维助手 — 持久化多轮对话设计

> 版本：v1.0  
> 日期：2026-08-26  
> 状态：待实现

---

## 1. 目标与范围

为 AIOPS 平台实现**持久化多轮对话**功能：

- 支持多会话管理（新建、切换、删除）。
- 支持 SSE 流式输出，逐字显示 LLM 回复。
- 后端使用 Go + Gin + GORM + Eino（OpenAI 兼容协议）。
- 大模型配置从 `llm_config` 表中 `is_default=1` 且 `is_enabled=1` 的记录读取。
- 上下文保留最近 10 轮 `user/assistant` 消息。
- 输出 MySQL 建表语句并提供 GORM 自动迁移。

---

## 2. 关键决策

| 决策项 | 选择 | 原因 |
|--------|------|------|
| 通信协议 | SSE（Server-Sent Events） | 单向流式足够，实现简单，浏览器原生支持 |
| 会话模型 | 多会话列表 | 符合主流 Chat 体验，便于长期运维知识沉淀 |
| 上下文长度 | 最近 10 轮 | 控制 token 与延迟 |
| 模型选择 | 仅使用默认配置 | 产品定义：后端始终读取 `llm_config.is_default=1` |
| 历史示例 | 删除 `internal/openaichat` | 用新的 `internal/chat` 包统一实现 |

---

## 3. 数据模型

### 3.1 实体关系

```
chat_session 1 ── N chat_message
sys_user     1 ── N chat_session
```

### 3.2 chat_session

| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(32) PK | UUID |
| user_id | VARCHAR(32) | 关联 sys_user.id |
| title | VARCHAR(100) | 会话标题 |
| status | VARCHAR(20) | active / archived，默认 active |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |
| deleted_at | DATETIME(3) | GORM 软删除 |

### 3.3 chat_message

| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(32) PK | UUID |
| session_id | VARCHAR(32) FK | 关联 chat_session.id |
| role | VARCHAR(20) | system / user / assistant |
| content | LONGTEXT | 完整消息内容 |
| prompt_tokens | INT | 可选，输入 token 数 |
| completion_tokens | INT | 可选，输出 token 数 |
| total_tokens | INT | 可选，总 token 数 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |
| deleted_at | DATETIME(3) | GORM 软删除 |

### 3.4 MySQL 建表语句

```sql
-- AIOPS 智能运维助手对话模块

CREATE TABLE IF NOT EXISTS `chat_session` (
    `id`              VARCHAR(32)  NOT NULL COMMENT '会话ID',
    `user_id`         VARCHAR(32)  NOT NULL COMMENT '用户ID',
    `title`           VARCHAR(100) NOT NULL DEFAULT '新会话' COMMENT '会话标题',
    `status`          VARCHAR(20)  NOT NULL DEFAULT 'active' COMMENT '状态 active/archived',
    `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`      DATETIME(3)           DEFAULT NULL COMMENT '删除时间（软删除）',
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_updated_at` (`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='对话会话';

CREATE TABLE IF NOT EXISTS `chat_message` (
    `id`                VARCHAR(32) NOT NULL COMMENT '消息ID',
    `session_id`        VARCHAR(32) NOT NULL COMMENT '会话ID',
    `role`              VARCHAR(20) NOT NULL COMMENT '角色 system/user/assistant',
    `content`           LONGTEXT    NOT NULL COMMENT '消息内容',
    `prompt_tokens`     INT                  DEFAULT 0 COMMENT '输入token数',
    `completion_tokens` INT                  DEFAULT 0 COMMENT '输出token数',
    `total_tokens`      INT                  DEFAULT 0 COMMENT '总token数',
    `created_at`        DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`        DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`        DATETIME(3)          DEFAULT NULL COMMENT '删除时间（软删除）',
    PRIMARY KEY (`id`),
    KEY `idx_session_id` (`session_id`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='对话消息';
```

---

## 4. 后端设计

### 4.1 新增文件

```
backend/
├── models/chat.go              # ChatSession、ChatMessage 模型
├── controllers/chat.go         # ChatController + SSE Handler
├── internal/chat/
│   ├── llm.go                  # Eino ChatModel 初始化与默认配置读取
│   └── service.go              # 会话/消息持久化与上下文组装
└── cmd/center/main.go          # 注册路由与 AutoMigrate
```

### 4.2 模型定义（GORM）

```go
type ChatSession struct {
    ID        string         `gorm:"primaryKey;size:32"`
    UserID    string         `gorm:"size:32;not null;index"`
    Title     string         `gorm:"size:100;not null;default:'新会话'"`
    Status    string         `gorm:"size:20;not null;default:'active'"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

type ChatMessage struct {
    ID               string         `gorm:"primaryKey;size:32"`
    SessionID        string         `gorm:"size:32;not null;index"`
    Role             string         `gorm:"size:20;not null"`
    Content          string         `gorm:"type:longtext;not null"`
    PromptTokens     int
    CompletionTokens int
    TotalTokens      int
    CreatedAt        time.Time
    UpdatedAt        time.Time
    DeletedAt        gorm.DeletedAt `gorm:"index"`
}
```

### 4.3 路由设计

所有接口位于 `/api/chat`，需登录认证。

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/chat/sessions` | 当前用户会话列表 |
| POST | `/api/chat/sessions` | 新建会话 |
| DELETE | `/api/chat/sessions/:id` | 删除会话（软删） |
| GET | `/api/chat/sessions/:id/messages` | 获取会话历史消息 |
| GET | `/api/chat/sessions/:id/stream` | SSE 流式对话（`?content=...`） |

### 4.4 SSE 事件格式

```text
event: message
data: {"chunk": "第一片段"}

event: message
data: {"chunk": "第二片段"}

event: done
data: {"message_id": "msg-xxx", "session_id": "sess-xxx", "done": true}
```

错误时：

```text
event: error
data: {"code": 500, "message": "调用大模型失败"}
```

### 4.5 核心流程

1. 校验会话归属（`user_id` 匹配）。
2. 保存用户问题到 `chat_message`（role=user）。
3. 查询默认 LLM 配置：`is_default=1 AND is_enabled=1`。
4. 组装上下文：固定 system prompt + 最近 10 轮历史 + 当前问题。
5. 使用 Eino `openai.ChatModel.Stream` 生成流式响应。
6. 流式输出 `event: message` 到客户端。
7. 流结束后，将完整内容写入 `chat_message`（role=assistant）。
8. 发送 `event: done` 并关闭连接。

---

## 5. 前端设计

### 5.1 新增文件

```
frontend/src/
├── api/chat.js           # 聊天相关 API 与 SSE 封装
└── views/ChatView.vue    # 改造现有页面，新增左侧会话列表
```

### 5.2 页面布局

- 左侧边栏：会话列表（新建、切换、删除）。
- 右侧主区域：
  - 顶部标题栏（助手名称 + 当前默认模型名称）。
  - 中部消息气泡区。
  - 底部快捷提问工具栏 + 输入框。

### 5.3 交互流程

1. 进入页面：
   - 调用 `GET /api/chat/sessions` 加载列表。
   - 若为空，调用 `POST /api/chat/sessions` 自动创建。
2. 切换会话：
   - 调用 `GET /api/chat/sessions/:id/messages` 渲染历史。
3. 发送消息：
   - 本地 push user 消息。
   - 创建 `EventSource('/api/chat/sessions/:id/stream?content=...')`。
   - 监听 `message` 追加 `chunk`。
   - 监听 `done` 关闭连接、刷新会话列表。
4. 错误处理：
   - SSE 返回 `event: error` 时提示用户。
   - 自动重连一次，失败则提示刷新。

### 5.4 状态管理

使用 `ref` + 组合式 API：

```js
const sessions = ref([])
const currentSessionId = ref('')
const messages = ref([])
const input = ref('')
const isStreaming = ref(false)
```

---

## 6. 错误处理

| 场景 | 行为 |
|------|------|
| 未配置默认 LLM | SSE `event: error`，前端引导到 AI 配置 |
| LLM 调用失败 | 输出 error 事件，不存 assistant 消息 |
| 会话不存在/越权 | 返回 404/403 JSON |
| SSE 连接中断 | 前端自动重连一次 |

---

## 7. 实现顺序

1. 数据模型：`backend/models/chat.go` + `docs/sql/chat_tables.sql`。
2. 后端服务层：`backend/internal/chat/`。
3. 后端控制器与路由：`backend/controllers/chat.go` + `main.go` 注册。
4. 前端 API 封装：`frontend/src/api/chat.js`。
5. 前端页面改造：`frontend/src/views/ChatView.vue`。
6. 联调与验证：`go build ./cmd/center`、`npm run build`。

---

## 8. 依赖与约束

- 后端依赖：`github.com/cloudwego/eino-ext/components/model/openai` 已存在。
- 前端不新增第三方依赖，使用原生 `EventSource`。
- 保持现有 `theme.css` 变量，不引入新颜色。
- 代码注释/用户可见文本使用中文，标识符保持英文。
