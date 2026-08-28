# AIOPS 智能运维助手 — 对话查询数据源设计

> 版本：v1.0  
> 日期：2026-08-28  
> 状态：设计稿 / 待讨论  

---

## 1. 背景与目标

当前助手已实现基于 `llm_config` 默认模型的持久化多轮对话（SSE 流式输出）。下一步希望在对话中直接查询已配置的 **Prometheus** 或 **Elasticsearch** 数据源，例如：

- “过去 1 小时 CPU 使用率 Top5 的容器”
- “查一下 nginx-error 索引最近 10 条 ERROR 日志”
- “对比两个数据中心的 QPS 趋势”

本文回答两个核心疑问：

1. 直接查询特定数据源的能力，是否要通过 **Eino 的 Tool API** 声明？
2. **Agent 应该怎么编排**？

---

## 2. 关键决策

| 决策项 | 推荐方案 | 原因 |
|---|---|---|
| 是否用 Eino Tool API | **是**，将每种查询能力声明为 Function Tool | 让 LLM 自己决定何时查、查哪个源、传什么参数，保持对话自然；也便于扩展新数据源 |
| 工具粒度 | 每个数据源 1~2 个查询工具 + 1 个元数据工具 | 太粗（一个 `query_data_source`）会让模型填参困难；太细会增加维护成本 |
| Agent 形态 | **单循环 Function Calling Agent**（ReAct 风格） | 当前只有两类数据源，足够简单；把编排写在 Go 代码里，便于观测、限流、审计 |
| 多 Agent 拆分 | 暂不拆分，预留一个 `IntentRouter` 扩展点 | 后续新增告警处理、根因分析等复杂能力时，再引入路由 Agent |
| 查询执行位置 | 后端 Go 执行，结果返回给 LLM 再生成回答 | 避免把数据源凭证或大量原始数据暴露给前端 |
| 安全性 | 只读查询 + 超时 + 结果截断 + 数据源权限校验 | 防止模型生成破坏性 DSL 或拖垮存储 |

---

## 3. 架构概览

```
用户提问
   │
   ▼
┌────────────────────────────────────────────┐
│  Chat Service（会话/历史/SSE 封装）          │
│  - 组装 system prompt + 历史 + tools        │
│  - 调用 Eino ChatModel（带 Tools）           │
└────────────────────────────────────────────┘
   │
   ▼  模型返回 tool_calls
┌────────────────────────────────────────────┐
│  Agent Loop（Go 代码实现）                   │
│  1. 解析 tool_calls                         │
│  2. 并发/串行执行工具                        │
│  3. 将结果以 tool 消息回传模型                │
│  4. 重复直到模型给出最终回答（最多 N 轮）      │
└────────────────────────────────────────────┘
   │
   ▼
┌─────────────────────────┐    ┌─────────────────────────┐
│  Prometheus Tool        │    │  Elasticsearch Tool     │
│  - query_instant        │    │  - search_logs          │
│  - query_range          │    │  - search_metrics/doc   │
│  - list_metric_names    │    │  - list_indices         │
└─────────────────────────┘    └─────────────────────────┘
   │                                  │
   ▼                                  ▼
Prometheus Client              Elasticsearch Client
（只读 query API）              （只读 _search）
```

---

## 4. Eino Tool API 的设计

### 4.1 为什么推荐用 Tool API

不用 Tool API 的替代方案是：**在代码里先做一次意图分类，再硬编码调用对应客户端**。这种方式的问题：

- 每新增一种查询意图就要改一次分类逻辑和参数解析。
- 用户的问法千变万化（“Top5 CPU”、“CPU 最高的几个 Pod”、“容器 CPU 排行”），分类器很难覆盖。
- 多轮对话里的指代、时间范围、过滤条件很难靠规则提取。

通过 Eino `FunctionTool` 把这些能力暴露给模型后，LLM 会基于工具描述和 Schema 自己决定调用哪个工具、填入哪些参数。后端只需负责：

- 描述清楚每个工具能做什么、参数含义、返回格式。
- 安全地执行工具并返回结果。

### 4.2 推荐声明的工具

| 工具名 | 所属数据源 | 作用 | 关键参数 |
|---|---|---|---|
| `list_data_sources` | 通用 | 让模型知道当前有哪些可用数据源 | 无 |
| `query_prometheus` | Prometheus | 执行 PromQL 查询 | `data_source_name`, `query`, `start`, `end`, `step` |
| `query_elasticsearch` | Elasticsearch | 执行日志/文档检索 | `data_source_name`, `index_pattern`, `query_string`, `start`, `end`, `size` |
| `get_prometheus_labels`（可选） | Prometheus | 帮助模型补全 metric/label | `data_source_name`, `metric_name` |
| `get_elasticsearch_indices`（可选） | Elasticsearch | 帮助模型确认索引名 | `data_source_name` |

> 工具数量不是越多越好。起步阶段建议先实现 `list_data_sources`、`query_prometheus`、`query_elasticsearch` 三个；如果模型经常写错 metric 名或索引名，再补充元数据工具。

### 4.3 参数设计要点

- **时间范围**：让模型用自然语言填 `start`/`end`，后端支持相对时间（如 `1h`, `now`）和 ISO 时间两种解析。
- **数据源名称**：使用 `datasources.name`，查询前校验该用户是否有权限访问。
- **PromQL / QueryString**：直接让模型生成。对 LLM 来说，PromQL 比 Lucene/DSL 更容易一次写对；ES 侧也可以先支持 `query_string`，后续再支持 DSL。
- **结果限制**：工具内部固定最大返回行数/字节数，不要把原始时序点或日志全文直接塞回 LLM。

---

## 5. Agent 编排方案

### 5.1 推荐：单循环 Function Calling Agent

在 `internal/chat` 里新增一个 `Agent` 层，核心是一个 `for` 循环：

1. **构造请求**：`system prompt` + `历史消息` + `当前用户消息` + `可用 tools`。
2. **调用 LLM**：使用 Eino `ChatModel.Generate`（或 Stream），传入 tools。
3. **判断结果**：
   - 若返回 `tool_calls`，进入步骤 4。
   - 若返回普通文本，直接作为最终回答输出。
4. **执行工具**：
   - 解析每个 `tool_call`，映射到已注册的 tool handler。
   - 并发执行（注意 Prometheus/ES 各自的超时）。
   - 收集结果，格式化为 JSON。
5. **回填消息**：将 tool 结果以 `role=tool` 的消息追加到上下文。
6. **循环**：回到步骤 2，直到模型给出最终回答或达到最大迭代次数（如 5 次）。
7. **流式输出**：最终回答通过 SSE `event: message` 流式返回前端。

```text
用户：过去1小时CPU使用率Top5的容器
LLM → tool_call: query_prometheus({query="topk(5,...)", ...})
后端 → 执行 PromQL → 返回 JSON
LLM → 基于结果生成中文回答
SSE → “过去1小时CPU使用率最高的5个容器是…”
```

### 5.2 为什么不先上多 Agent

当前场景只有两类数据源查询，单循环足够。拆成多个 Agent 的收益不明显，反而带来：

- 状态传递复杂（上下文、工具结果、错误）。
- 延迟增加。
- 调试困难。

建议把多 Agent 作为**未来扩展点**：当助手需要同时处理“查指标 → 查日志 → 根因分析 → 创建工单”这类多步骤任务时，再引入一个顶层 `IntentRouter`，根据意图把请求分发给：

- `DataQueryAgent`（本文设计）
- `KnowledgeAgent`（已有 RAG）
- `AlertAgent`（告警处理）
- `OpsAgent`（执行操作/创建工单）

### 5.3 可选增强：意图预过滤

如果担心模型乱调工具，可以在进入循环前加一层轻量意图判断：

- 简单规则：问题包含 “CPU/内存/指标/QPS/容器” → 启用 Prometheus 工具；包含 “日志/index/ES/错误” → 启用 Elasticsearch 工具。
- 或者再用一次 LLM 做 `IntentRouter`（很小，很快）。

这样既能减少无关工具对模型的干扰，也保留了工具调用的灵活性。

---

## 6. 与现有对话流程的集成

基于 `2026-08-26-chat-design.md` 的 `GET /api/chat/sessions/:id/stream` 进行扩展：

- 在 `internal/chat/client.go` 初始化 ChatModel 时，把工具列表传入 `Tools` 配置。
- 在 `internal/chat/service.go` 的对话逻辑里，把原来的“单次调用 LLM”替换为“Agent 循环”。
- SSE 事件增加 `event: status` 或 `event: tool_call`，用于在前端展示“正在查询指标平台…”等中间状态，避免长时间静默。
- 工具执行报错时，把错误信息包装成 tool 结果回传模型，让模型向用户解释，而不是直接 SSE `event: error`。

---

## 7. 数据源客户端与安全性

### 7.1 客户端抽象

新增 `internal/datasource` 包：

- `PrometheusClient`：封装 `/api/v1/query`、`/api/v1/query_range`。
- `ElasticsearchClient`：封装 `GET /<index>/_search`，只读。
- 统一读取 `datasources` 表获取 URL、用户名、密码、超时、跳过 SSL 等配置。

### 7.2 安全措施

| 风险 | 防护手段 |
|---|---|
| 模型构造删除/写入 DSL | 工具只调用只读 API，禁止 `_delete_by_query`、`_update_by_query`、索引创建等 |
| 查询拖垮存储 | 统一超时（如 15s）、限制返回行数（如 100 条）、限制单次返回字节（如 64KB） |
| 越权访问数据源 | 当前阶段按全局可用处理，仅校验数据源存在且 `is_enabled=1`；后续如需权限再扩展 |
| 凭证泄露 | 凭证只存在于后端内存，不返回给前端，不写入 `chat_message` |
| Prompt 注入 | system prompt 中明确“只能查询，不能执行修改操作”，并配合只读 API 兜底 |

---

## 8. 数据预聚合策略（新增）

由于确认工具返回的原始数据需要后端先做轻量聚合再交给 LLM，设计上把“聚合/截断”作为工具执行层的标准步骤，而不是让 LLM 直接面对海量原始点或日志。

### 8.1 通用原则

- 每次工具返回给 LLM 的数据量控制在固定阈值内，例如：
  - 时序：最多 100 个序列，每个序列最多 200 个采样点；
  - 日志：最多 20 条原始日志或聚合后的桶；
  - 总 JSON 大小不超过 32KB~64KB。
- 超过阈值时优先做聚合，而非简单截断，避免丢失关键信息。
- 聚合逻辑由后端根据数据源类型实现，对 LLM 透明；工具描述中说明“返回已聚合的可读摘要”。

### 8.2 Prometheus 聚合

- **TopN**：如果查询结果是多序列，按最新值或均值取 TopN（N 默认 10，可被 `topn` 参数覆盖）。
- **降采样**：按查询时长自动选择 step，避免返回过多点。
- **对齐**：多序列时间对齐，缺失值填充或忽略。
- **统计摘要**：可返回每个序列的 `min/max/avg/latest`，减少 LLM 解析负担。

### 8.3 Elasticsearch 聚合

- 按时间分桶（histogram/date_histogram）给出趋势，而不是返回全部原始日志。
- 对日志内容做 terms 聚合（如 ERROR 级别、异常类型）。
- 如需原始日志，则只返回前 N 条（按时间倒序）并限制字段数。
- 高亮关键字，但只保留关键字段。

### 8.4 对 Tool Schema 的影响

在 `query_prometheus` / `query_elasticsearch` 中增加可选参数：

| 参数 | 类型 | 说明 |
|---|---|---|
| `limit` | int | 返回条数/序列数上限，默认由后端控制 |
| `topn` | int | 仅 Prometheus，取 TopN |
| `aggregation` | string | 可选提示，如 `trend`/`topk`/`latest`，后端可据此优化聚合 |

模型不一定准确填写这些参数，后端以安全阈值为准。

---

## 9. 可扩展性

新增一种数据源（例如 Loki、ClickHouse、MySQL）时，只需：

1. 新增 `internal/datasource/xxx_client.go`。
2. 新增 1~2 个 Eino FunctionTool，内部实现该数据源的聚合/截断逻辑。
3. 在工具注册表中添加该工具。

Agent 循环、SSE 流程、会话持久化都不需要改动。

---

## 10. 已确认问题与设计调整

| 问题 | 确认结果 | 设计调整 |
|---|---|---|
| 1. 默认模型是否支持 Function Calling？ | 支持 | 直接采用 Eino Tool API，无需降级到意图分类 + 硬编码调用 |
| 2. `datasources` 是否需权限处理？ | 不需要 | 当前按全局可用处理，执行前仅校验存在且 `is_enabled=1` |
| 3. 是否允许跨数据源提问？ | 允许 | 单循环 Agent 在一次迭代中可同时发起多个 `tool_calls`，后端并发执行后合并结果 |
| 4. 原始数据是否需后端聚合？ | 需要 | 新增第 8 节数据预聚合策略，作为工具执行层默认行为 |

---

## 11. 结论

- **用 Eino Tool API 声明查询能力**，比手工意图分类更自然、更可扩展。
- **Agent 采用单循环 Function Calling**：模型决定调哪些工具，后端执行（含预聚合）后回填结果，循环直到生成最终回答。
- 在当前阶段**不必拆多 Agent**，但预留 `IntentRouter` 扩展点。
- 安全上坚持**只读、超时、截断**三原则；权限校验当前按全局可用处理。
