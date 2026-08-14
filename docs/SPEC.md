# AIOPS 前端原型 SPEC（单一事实来源）

## 项目
AIOPS 智能运维平台前端原型。Vue 3 + Vite + Vue Router + ECharts，纯前端，无后端，所有数据来自 `src/mock/data.js`。

## 设计规范（必须严格遵守）
- 配色：白色为主 + 深青色主题色。CSS 变量定义在 `src/styles/theme.css`，一律使用变量，禁止硬编码颜色。
  - `--c-primary: #0E7C72`（深青），`--c-primary-dark: #0A5F57`，`--c-primary-tint: #E7F3F1`
  - `--c-bg: #F6F8F7`（页面底），`--c-surface: #FFFFFF`，`--c-border: #E4E9E7`
  - 严重程度色：`--c-p0: #C93B3B` `--c-p1: #D97B29` `--c-p2: #C9A227` `--c-p3: #0E7C72` `--c-p4: #8A9693`
- 禁止蓝色-紫色渐变、禁止高饱和背景；整体低饱和、留白充足、层级清晰。
- 圆角：卡片 10px，标签/徽章 6px；阴影用 `--shadow-card`。
- 图表：ECharts，统一使用 `src/components/ChartBox.vue` 封装，配色取主题变量，网格/坐标轴用浅灰。

## 布局
`src/layout/AppLayout.vue`：左侧固定导航（240px，白底，Logo + 7 个菜单项 + 底部系统状态）+ 顶部栏（页面标题、全局搜索框、环境标识、告警铃铛、用户头像）。内容区 `router-view`。所有页面都在此布局内。

## 路由（已在 src/router/index.js 注册）
| path | name | 文件 | 菜单名 |
|---|---|---|---|
| / | dashboard | src/views/Dashboard.vue | 总览大盘（已完成，作为示例） |
| /alerts | alerts | src/views/AlertsView.vue | 告警控制台 |
| /anomaly | anomaly | src/views/AnomalyView.vue | 异常检测 |
| /rca | rca | src/views/RcaView.vue | 根因分析 |
| /logs | logs | src/views/LogAnalysisView.vue | 日志分析 |
| /denoise | denoise | src/views/DenoiseView.vue | 告警降噪 |
| /datasource | datasource | src/views/DatasourceView.vue | 数据源接入 |

## 共享组件（src/components/）
- `StatCard.vue`：KPI 卡片（label、value、delta、icon slot）
- `LevelTag.vue`：级别标签（props: level → P0-P4 / critical/major/minor/warning/info，渲染对应颜色徽章）
- `ChartBox.vue`：ECharts 封装（props: option、height；自动 resize）
- `PageHeader.vue`：页头（标题、描述、右侧操作 slot）

## Mock 数据（src/mock/data.js 已导出）
- `kpiStats`、`alertTrend`、`severityDist`、`serviceHealth`
- `alerts`（20+ 条：id/title/service/level/status/time/source/count）、`alertRules`
- `anomalyResults`、`anomalyAlgorithms`、`metricSeries`
- `rcaCases`（含 rootCauses/evidence/impact/actions）、`topologyNodes`、`topologyEdges`
- `logClusters`、`logTemplates`、`logSeries`
- `denoisePolicies`、`denoiseStats`
- `dataSources`、`collectorStats`
- 工具函数：`fmtTime`、`seededSeries`
新增 mock 数据时追加到该文件，保持命名风格。

## 分工
- 主代理：骨架 + Dashboard（示例页，已在 main 分支）
- coder-A（分支 feat-a）：AlertsView.vue、AnomalyView.vue、RcaView.vue
- coder-B（分支 feat-b）：LogAnalysisView.vue、DenoiseView.vue、DatasourceView.vue

## 页面内容要求（对应设计文档功能）
### AlertsView 告警控制台
顶部统计条（活动告警/今日新增/已确认/已解决/告警压缩率）→ 筛选栏（级别、状态、服务、关键字）→ 告警表格（级别徽章、标题、服务、来源、次数、状态、时间、操作：确认/解决，操作后本地状态更新）→ 右侧或下方：告警生命周期流程示意（文档 5.3.2：规则过滤→窗口聚合→相似度去重→拓扑抑制→动态阈值→智能分级→通知路由）。
### AnomalyView 异常检测
顶部：检测任务统计卡片 → 左侧指标时序图（ChartBox，叠加异常点标注 markPoint/markArea，正常区间带）→ 右侧检测结果详情（异常评分、置信度、偏差、严重程度、可解释性说明：method/trend/seasonality/reason）→ 下方算法选型矩阵表（文档 5.1.2：场景/算法/模型类型/优势/适用数据）+ 模型列表。
### RcaView 根因分析
顶部事件选择（incident 列表）→ 根因排名列表（rank、置信度进度条、实体信息、证据列表、影响范围、修复建议，参考文档 5.2.3 响应结构）→ 服务拓扑示意（用 ECharts graph 或 SVG 绘制节点与调用边，异常节点高亮）。
### LogAnalysisView 日志分析
日志处理流水线示意（文档 5.4.1：解析→模板提取→参数分离→向量化→聚类→异常检测）→ 日志聚类列表（模式、数量趋势、级别分布）→ Drain 模板提取展示（模板 + 参数示例表格，参考文档 5.4.2）→ 异常日志时间线图。
### DenoiseView 告警降噪
降噪效果统计（原始告警→有效告警漏斗/对比、压缩率）→ 五种策略卡片（时间窗口聚合/相似度合并/拓扑抑制/动态阈值/智能分级，含说明、预期效果、开关、状态）→ 降噪策略配置表格。
### DatasourceView 数据源接入
数据源卡片（ELK/Prometheus：类型、状态、接入方式、连接信息）→ 采集统计（指标速率、日志速率图表）→ 日志结构化字段表（文档 4.2.3）→ 指标类型覆盖表（文档 4.1.3）。

## 质量要求
- 全部中文界面；数据要"像真的"（服务名 order-service/payment-service 等，与文档一致）。
- 每页内容充实、信息密度合理，不是占位空页。
- `npm run build` 必须通过。
