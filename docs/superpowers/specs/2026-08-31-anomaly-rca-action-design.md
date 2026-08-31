# 异常检测列表根因分析按钮设计

## 背景

异常检测页面已接入 Nightingale 告警事件（活跃与历史）。运维人员看到具体告警后，希望一键触发根因分析，由 AI 自动排查相关指标与日志，定位宕机原因。

## 目标

在异常检测列表增加“操作”列，每行提供“根因分析”按钮。点击后跳转到对话页面，由现有 LLM Agent 调用 Prometheus/Elasticsearch 工具完成根因分析。

## 方案

采用“跳转对话页 + Agent 自动排查”方案（方案 A）。

## 详细设计

### 1. 前端：AnomalyView.vue

- 表格表头增加 `<th>操作</th>`。
- 每行事件数据后增加操作单元格，包含“根因分析”按钮。
- 空状态/加载状态的 `colspan` 从 `7` 改为 `8`。
- 点击按钮时，将事件关键字段编码为 URL 查询参数，跳转至 `/#/chat?rca=1&event=<base64(json)>`。
- 携带字段：
  - `rule_name`：规则名称
  - `target_ident`：告警对象
  - `tags`：标签
  - `trigger_time`：触发时间（秒级时间戳）
  - `trigger_value`：触发值
  - `severity`：级别
  - `is_recovered`：是否已恢复

### 2. 前端：ChatView.vue

- 路由挂载后检测 `rca=1` 与 `event` 参数。
- 解码并解析事件 JSON。
- 自动生成根因分析 Prompt，例如：

  ```text
  请对以下告警事件进行根因分析：
  - 规则名称：服务宕机，业务中断，需紧急干预
  - 告警对象：10.2.209.145
  - 触发时间：2026/08/31 14:13:32
  - 级别：P1-紧急
  - 状态：告警中
  - 标签：__name__=up,alertName=Instance,...
  - 触发值：0.00

  请查询相关 Prometheus 指标和 Elasticsearch 日志，排查宕机原因，并给出根因、证据链和修复建议。
  ```

- 将 Prompt 自动填入输入框并触发 `send()`，实现“一键分析”。
- 发送后清除路由参数，避免刷新重复触发。

### 3. 后端

复用现有能力，无需新增接口：

- `/api/chat/sessions/:id/stream` 保持不变。
- `internal/chat/agent.go` 已有 Function Calling 循环，自动调用 `query_prometheus` / `query_elasticsearch`。
- `internal/chat/tools/registry.go` 已有工具实现。
- 根因分析能力取决于用户是否已配置 Prometheus/Elasticsearch 数据源。

### 4. 错误与兜底

- 若事件 JSON 解码失败，在输入框回显原始参数并提示用户手动检查。
- 若当前没有会话，先自动创建新会话再发送 RCA Prompt。
- 若模型未调用工具直接回答，由现有 Agent 的拒绝逻辑兜底，提示用户检查数据源配置。

## 依赖

- 已启用的 Prometheus 数据源（用于查询指标）。
- 已启用的 Elasticsearch 数据源（用于查询日志）。
- 已配置默认 LLM。

## 验收标准

- 异常检测列表出现“操作”列和“根因分析”按钮。
- 点击按钮后跳转到对话页并自动发送根因分析请求。
- Agent 回复中包含对指标/日志的分析结论。
- 页面刷新不会重复触发分析。
