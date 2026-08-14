# AIOPS 系统设计文档

## 文档信息

| 项目     | 内容         |
| -------- | ------------ |
| 文档版本 | v1.2         |
| 创建日期 | 2026-08-11   |
| 编写人   | AIOPS 项目组 |
| 审核状态 | 待评审       |

---

## 目录

1. [项目背景](#1-项目背景)
2. [系统概述](#2-系统概述)
3. [技术架构](#3-技术架构)
4. [数据源接入方案](#4-数据源接入方案)
5. [核心功能模块](#5-核心功能模块)
   - [5.5 运维知识库模块](#55-运维知识库模块)
   - [5.6 智能运维助手模块](#56-智能运维助手模块)
6. [技术栈选型](#6-技术栈选型)
7. [实施路线图](#7-实施路线图)
8. [接口设计](#8-接口设计)
9. [部署架构](#9-部署架构)
10. [附录](#10-附录)

---

## 1. 项目背景

### 1.1 业务痛点

当前运维体系面临以下核心挑战：

| 痛点         | 现状                                  | 影响                             |
| ------------ | ------------------------------------- | -------------------------------- |
| 告警风暴     | 日均告警量超 5000+，大量重复/关联告警 | 运维人员疲于应对，关键告警被淹没 |
| 故障定位慢   | 平均 MTTR 超过 30 分钟                | 业务中断时间长，SLA 难以保障     |
| 容量规划难   | 依赖人工经验判断资源需求              | 资源浪费或不足，成本不可控       |
| 日志分析低效 | 海量日志人工排查，缺乏智能分析        | 问题发现滞后，根因定位困难       |
| 缺乏预测能力 | 被动式运维，事后响应                  | 无法提前规避潜在风险             |

### 1.2 建设目标

通过引入 AIOPS 系统，实现以下目标：

- **异常检测自动化**：自动识别指标和日志中的异常模式，减少人工监控成本
- **根因分析智能化**：将平均故障定位时间（MTTR）缩短 60% 以上
- **告警降噪**：告警压缩率达到 80% 以上，提升告警信噪比
- **预测性维护**：提前 24-72 小时预警潜在故障
- **容量智能规划**：基于历史数据实现自动化容量预测与建议

### 1.3 现有数据基础

| 数据源                                   | 类型     | 状态   | 接入方式                 |
| ---------------------------------------- | -------- | ------ | ------------------------ |
| ELK（Elasticsearch + Logstash + Kibana） | 日志数据 | 已部署 | API 查询                 |
| Prometheus                               | 指标数据 | 已部署 | Remote Read / Federation |

---

## 2. 系统概述

### 2.1 系统定位

AIOPS 智能运维平台是一个面向大规模分布式系统的智能运维中枢，通过融合多源监控数据，运用机器学习与深度学习技术，实现从被动响应到主动预防的运维模式转变。

### 2.2 核心能力矩阵

```
                    数据采集
                       |
        +--------------+--------------+
        |                             |
    指标分析                      日志分析
        |                             |
    异常检测                      日志聚类
    时序预测                      模式识别
    关联分析                      异常日志发现
        |                             |
        +--------------+--------------+
                       |
                   根因分析引擎
                       |
              +--------+--------+
              |                 |
          告警降噪          智能告警
          告警聚合          修复建议
          动态阈值          自动化修复
              |                 |
              +--------+--------+
                       |
                   可视化大盘
                       |
              +--------+--------+
              |                 |
          运维知识库        智能运维助手
          文档管理            对话问答
          语义检索            告警解读
          经验沉淀            排障建议
              |                 |
              +--------+--------+
                       |
                   反馈闭环
```

> **注**：静态阈值类原始告警由外部规则引擎 **夜莺（Nightingale）** 生成，通过 webhook 接入后，与指标/日志异常检测结果共同进入告警降噪与智能告警流程。

### 2.3 设计原则

1. **高可用**：核心服务多实例部署，算法服务故障时不影响基础告警能力
2. **可扩展**：模块化设计，支持新数据源和算法模型的快速接入
3. **可解释**：所有 AI 决策输出附带可解释性说明，便于运维人员理解和信任
4. **松耦合**：业务层与算法层解耦，通过标准化接口通信
5. **反馈闭环**：支持人工反馈，持续优化模型效果

---

## 3. 技术架构

### 3.1 总体架构图

```
+------------------------------------------------------------------+
|                         用户交互层                                |
|  +----------------+  +----------------+  +------------------+   |
|  |   告警控制台    |  |   可视化大盘    |  |   配置管理中心    |   |
|  +----------------+  +----------------+  +------------------+   |
+------------------------------------------------------------------+
                              |
+------------------------------------------------------------------+
|                         API 网关层 (Go)                          |
|  +----------------+  +----------------+  +------------------+   |
|  |   请求路由      |  |   认证鉴权      |  |   限流熔断        |   |
|  +----------------+  +----------------+  +------------------+   |
+------------------------------------------------------------------+
                              |
+------------------------------------------------------------------+
|                        业务服务层 (Go)                            |
|  +--------+  +--------+  +--------+  +--------+  +----------+   |
|  | 告警管理 |  | 事件编排 |  | 工单系统 |  | 夜莺接入 |  | 用户权限   |   |
|  +--------+  +--------+  +--------+  +--------+  +----------+   |
|  +--------+  +--------+  +--------+  +--------+  +----------+   |
|  | 策略管理 |  | 通知中心 |  | 配置管理 |  | 拓扑管理 |  | 智能运维助手|   |
|  +--------+  +--------+  +--------+  +--------+  +----------+   |
+------------------------------------------------------------------+
                              |
+------------------------------------------------------------------+
|                       算法引擎层 (Python)                        |
|  +----------------+  +----------------+  +------------------+  +------------------+   |
|  |   异常检测服务   |  |   根因分析服务   |  |   日志分析服务    |  |   知识库服务    |   |
|  |   (FastAPI)      |  |   (FastAPI)      |  |   (FastAPI)       |  |   (FastAPI)     |   |
|  +----------------+  +----------------+  +------------------+  +------------------+   |
|  +----------------+  +----------------+  +------------------+                        |
|  |   时序预测服务   |  |   告警降噪服务   |  |   知识图谱服务    |                        |
|  |   (FastAPI)      |  |   (FastAPI)      |  |   (FastAPI)       |                        |
|  +----------------+  +----------------+  +------------------+                        |
+------------------------------------------------------------------+
                              |
+------------------------------------------------------------------+
|                       数据采集层 (Go)                             |
|  +----------------+  +----------------+  +------------------+   |
|  | Prometheus 采集器|  |   ELK 采集器     |  |  扩展采集器接口   |   |
|  +----------------+  +----------------+  +------------------+   |
+------------------------------------------------------------------+
                              |
+------------------------------------------------------------------+
|                         数据存储层                                |
|  +------------+  +------------+  +------------+  +----------+  +----------+  |
|  | VictoriaMetrics|  | Elasticsearch|  |  PostgreSQL  |  |  Redis   |  | Vector DB|  |
|  |  (时序数据)   |  |  (日志数据)   |  |  (业务数据)   |  |  (缓存)   |  | (向量数据)|  |
|  +------------+  +------------+  +------------+  +----------+  +----------+  |
|  +------------+  +------------+                                   |
|  |   Kafka    |  |  MinIO/S3  |                                   |
|  | (消息队列)  |  | (对象存储)  |                                   |
|  +------------+  +------------+                                   |
+------------------------------------------------------------------+
```

> **外部规则引擎说明**：业务服务层中的“夜莺接入”指独立部署的 **Nightingale** 规则引擎，通过 webhook 向 AIOPS 推送静态阈值告警事件；AIOPS 负责告警标准化、降噪、分级与通知。

### 3.2 语言分工说明

| 层级     | 语言   | 框架/库                                   | 选择理由                                          |
| -------- | ------ | ----------------------------------------- | ------------------------------------------------- |
| API 网关 | Go     | Gin / Echo                                | 高并发、低延迟、编译型语言，适合处理海量实时请求  |
| 业务服务 | Go     | Gin + GORM                                | 标准业务 CRUD，Go 的 goroutine 天然适合高并发场景 |
| 数据采集 | Go     | prometheus/client_golang, olivere/elastic | 资源占用低，适合长期运行的采集 Agent              |
| 算法引擎 | Python | FastAPI + scikit-learn / PyTorch          | 丰富的 ML/DL 生态，快速迭代算法模型               |
| 知识库服务 | Python | FastAPI + SentenceTransformers / LangChain | 文本向量化、语义检索、RAG 流程编排               |
| 向量数据库 | -      | Milvus / Qdrant / pgvector                | 高维向量相似度检索，支撑语义召回                 |
| 任务调度 | Python | Celery + Redis                            | 异步任务处理，模型训练、知识库索引与推理调度      |

### 3.3 服务间通信

```
Go 业务服务  <--gRPC/HTTP-->  Python 算法服务
     |                              |
     +---------- Kafka -------------+
     |         (异步数据流)            |
     +---------- Redis -------------+
               (缓存/状态)

智能运维助手 (Go)  <--HTTP/gRPC-->  知识库服务 (Python)
                                          |
                                    向量数据库 (Vector DB)
```

- **同步通信**：Go 业务服务通过 gRPC/HTTP 调用 Python 算法服务进行实时推理；智能运维助手同步调用知识库服务获取检索上下文
- **异步通信**：通过 Kafka 传输大规模时序数据和日志数据，解耦采集与处理；知识库文档解析与向量化通过 Celery 异步执行
- **状态共享**：Redis 用于缓存告警状态、会话信息、热点数据、AI 助手多轮对话上下文

---

## 4. 数据源接入方案

### 4.1 Prometheus 指标接入

#### 4.1.1 架构流程

```
Prometheus Server / Thanos Sidecar / VictoriaMetrics
                    |
            Remote Read API / Federation / remote_write
                    |
        +-----------+-----------+
        |                       |
   实时流 (remote_write)    批量查询 (HTTP API)
        |                       |
    Kafka Topic            Go 采集服务
    "metrics-raw"          (prometheus/client_golang)
        |                       |
        +-----------+-----------+
                    |
            数据清洗 & 预处理
                    |
            +-------+-------+
            |               |
        时序数据库      算法引擎
    VictoriaMetrics    (异常检测/预测)
```

#### 4.1.2 接入配置

**Remote Write 配置（Prometheus）**

```yaml
remote_write:
  - url: "http://aiops-collector:9090/api/v1/receive"
    queue_config:
      capacity: 10000
      max_samples_per_send: 2000
      max_shards: 200
    write_relabel_configs:
      - source_labels: [__name__]
        regex: 'up|node_cpu_seconds_total|node_memory_MemAvailable_bytes|.*'
        action: keep
```

**Go 采集服务核心逻辑**

```go
// 使用 prometheus/client_golang 查询历史数据
import (
    "github.com/prometheus/client_golang/api"
    v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

func QueryMetricsRange(query string, start, end time.Time, step time.Duration) (model.Matrix, error) {
    client, err := api.NewClient(api.Config{
        Address: "http://prometheus:9090",
    })
    if err != nil {
        return nil, err
    }
    
    api := v1.NewAPI(client)
    r := v1.Range{
        Start: start,
        End:   end,
        Step:  step,
    }
    
    result, warnings, err := api.QueryRange(context.Background(), query, r)
    // 处理结果...
}
```

#### 4.1.3 指标类型覆盖

| 指标类型  | 说明             | 异常检测适用场景           |
| --------- | ---------------- | -------------------------- |
| Counter   | 累计值，只增不减 | 请求总量、错误总量突变检测 |
| Gauge     | 可增可减的瞬时值 | CPU、内存、磁盘使用率异常  |
| Histogram | 分桶统计         | 请求延迟分布异常           |
| Summary   | 分位数统计       | P99/P95 延迟突增检测       |

### 4.2 ELK 日志接入

#### 4.2.1 架构流程

```
应用服务 → Filebeat / Fluentd → Logstash → Elasticsearch
                                                  |
                                          Go 日志采集服务
                                          (olivere/elastic v7)
                                                  |
                                          +-------+-------+
                                          |               |
                                      日志解析        DSL 查询
                                      结构化处理      时间窗口拉取
                                          |               |
                                          +-------+-------+
                                                  |
                                          Kafka Topic "logs-raw"
                                                  |
                                          +-------+-------+
                                          |               |
                                      日志存储        算法引擎
                                  Elasticsearch     (日志聚类/异常)
```

#### 4.2.2 接入配置

**Go 日志采集服务核心逻辑**

```go
import "github.com/olivere/elastic/v7"

func FetchLogs(index string, query elastic.Query, from, size int) (*elastic.SearchResult, error) {
    client, err := elastic.NewClient(
        elastic.SetURL("http://elasticsearch:9200"),
        elastic.SetSniff(false),
    )
    if err != nil {
        return nil, err
    }
    
    searchResult, err := client.Search().
        Index(index).
        Query(query).
        From(from).Size(size).
        Sort("@timestamp", true).
        Do(context.Background())
    
    return searchResult, err
}
```

#### 4.2.3 日志结构化字段

| 字段名          | 类型    | 说明         | 示例                     |
| --------------- | ------- | ------------ | ------------------------ |
| `@timestamp`    | date    | 日志时间     | 2026-08-11T08:00:00.000Z |
| `service`       | keyword | 服务名       | order-service            |
| `level`         | keyword | 日志级别     | ERROR, WARN, INFO        |
| `trace_id`      | keyword | 链路追踪ID   | abc123def456             |
| `host`          | keyword | 主机名       | node-01                  |
| `message`       | text    | 原始日志内容 | Connection timeout...    |
| `error_type`    | keyword | 错误类型     | TimeoutException         |
| `response_time` | long    | 响应时间(ms) | 5000                     |

### 4.3 夜莺告警接入

#### 4.3.1 定位

夜莺（Nightingale）作为 AIOPS 的**外部规则引擎**，负责静态阈值规则的触发与首次收敛。AIOPS 通过 webhook 接收实时告警事件，并可选通过夜莺 API 查询规则列表与历史告警，用于大盘展示、根因分析 enrichment。

#### 4.3.2 架构流程

```
+-----------------------------------+       webhook        +------------------+
|  夜莺 (Nightingale)               | -------------------> |  AIOPS API 网关   |
|  - 静态阈值规则                    |                      |  /api/v1/alerts  |
|  - 告警事件推送                    |                      +------------------+
|  - 规则/历史 API                   |                               |
+-----------------------------------+                               v
        ^                                                   +------------------+
        |                                                   |  告警管理/标准化  |
        +---------------- 可选查询 -------------------------+  - source 标记    |
                                                              |  - severity 映射 |
                                                              +------------------+
```

#### 4.3.3 接入方式

| 方式 | 接口/机制 | 用途 | 说明 |
| ---- | -------- | ---- | ---- |
| **实时推送** | Nightingale webhook → AIOPS | 接收触发后的告警事件 | 核心路径，需配置接收端点与鉴权 |
| **规则查询** | Nightingale API → AIOPS | 展示规则列表、规则详情 | 只读展示，避免在 AIOPS 中重复维护规则 |
| **历史查询** | Nightingale API → AIOPS | 查询历史告警事件 | 用于降噪复盘、根因分析 enrichment |

#### 4.3.4 告警标准化字段

| 夜莺字段 | AIOPS 字段 | 说明 |
| -------- | ---------- | ---- |
| `rule_id` / `rule_name` | `rule_id` / `alert_name` | 规则标识 |
| `severity` | `severity` | 需映射到 P0‑P4 |
| `labels`（含 `service`、`instance` 等） | `labels` | 用于拓扑抑制、服务关联 |
| `value` / `trigger_value` | `value` | 触发时的指标值或错误计数 |
| `trigger_time` | `timestamp` | 告警触发时间 |
| `description` / `content` | `description` | 告警描述 |

标准化后统一打上 `source=nightingale`，以便与异常检测产生的告警区分。

#### 4.3.5 严重级别映射建议

| 夜莺 severity | AIOPS 级别 | 备注 |
| ------------- | ---------- | ---- |
| 1 / Critical  | P0         | 紧急，需立即响应 |
| 2 / High      | P1         | 重要，尽快响应 |
| 3 / Medium    | P2         | 中等，按排期处理 |
| 4 / Low       | P3/P4      | 提示/信息，可聚合后展示 |

#### 4.3.6 通知与高可用

- **通知 ownership**：建议关闭夜莺侧的普通通知，仅保留兜底通知；由 AIOPS 统一完成降噪、分级后的通知路由。
- **高可用**：夜莺故障时，AIOPS 内部的异常检测、时序预测等能力仍可独立产生告警，保证基础智能告警能力不中断。

---

## 5. 核心功能模块

### 5.1 异常检测模块

#### 5.1.1 模块架构

```
+--------------------------------------------------+
|                 异常检测引擎 (Python/FastAPI)       |
+--------------------------------------------------+
|  +----------------+  +----------------+         |
|  |   数据预处理    |  |   特征工程      |         |
|  | - 缺失值处理    |  | - 统计特征      |         |
|  | - 归一化       |  | - 时序特征      |         |
|  | - 降采样       |  | - 频域特征      |         |
|  +----------------+  +----------------+         |
|  +----------------+  +----------------+         |
|  |   模型推理      |  |   结果后处理    |         |
|  | - 加载模型      |  | - 阈值过滤      |         |
|  | - 批量预测      |  | - 置信度校准    |         |
|  | - 实时推理      |  | - 结果格式化    |         |
|  +----------------+  +----------------+         |
+--------------------------------------------------+
```

#### 5.1.2 算法选型矩阵

| 场景       | 算法             | 模型类型 | 优势                          | 适用数据           |
| ---------- | ---------------- | -------- | ----------------------------- | ------------------ |
| 单指标异常 | Isolation Forest | 无监督   | 无需标注数据，适合高维        | CPU、内存、QPS     |
| 单指标异常 | Prophet          | 时序分解 | 可解释性强，自动处理趋势/季节 | 有周期性规律的指标 |
| 单指标异常 | LSTM Autoencoder | 深度学习 | 捕捉复杂时序模式              | 高频指标（秒级）   |
| 多指标关联 | VAE              | 深度学习 | 学习正常模式流形              | 多维度指标向量     |
| 多指标关联 | GMM              | 概率模型 | 计算快，可解释                | 中等维度（<50）    |
| 日志异常   | Drain            | 模板提取 | 高效，在线学习                | 非结构化日志       |
| 日志异常   | LogBERT          | NLP      | 语义理解能力强                | 含语义信息的日志   |
| 日志异常   | DeepLog          | LSTM     | 序列模式学习                  | 系统调用序列       |

#### 5.1.3 异常检测 API 设计

**请求**

```json
POST /api/v1/anomaly/detect
Content-Type: application/json

{
  "data_source": "prometheus",
  "metric_name": "node_cpu_seconds_total",
  "labels": {
    "instance": "10.0.0.1:9100",
    "mode": "idle"
  },
  "time_range": {
    "start": "2026-08-10T00:00:00Z",
    "end": "2026-08-11T00:00:00Z"
  },
  "algorithm": "prophet",
  "sensitivity": "medium",
  "return_explanation": true
}
```

**响应**

```json
{
  "status": "success",
  "request_id": "req_abc123",
  "result": {
    "anomaly_detected": true,
    "anomaly_score": 0.89,
    "confidence": 0.92,
    "anomaly_points": [
      {
        "timestamp": "2026-08-10T14:30:00Z",
        "value": 95.2,
        "expected_value": 45.0,
        "deviation": 50.2,
        "severity": "critical"
      }
    ],
    "explanation": {
      "method": "基于Prophet时序分解",
      "trend": "整体呈上升趋势",
      "seasonality": "检测到日周期模式",
      "reason": "实际值超出预测区间3个标准差"
    }
  }
}
```

### 5.2 根因分析模块

#### 5.2.1 模块架构

```
+--------------------------------------------------+
|                 根因分析引擎 (Python)               |
+--------------------------------------------------+
|  +----------------+  +----------------+         |
|  |   数据融合层    |  |   知识图谱层    |         |
|  | - 异常指标      |  | - 服务拓扑      |         |
|  | - 异常日志      |  | - 调用链关系    |         |
|  | - 告警事件      |  | - 基础设施关系  |         |
|  | - 变更记录      |  | - 历史故障库    |         |
|  +----------------+  +----------------+         |
|  +----------------+  +----------------+         |
|  |   推理引擎      |  |   结果生成      |         |
|  | - 图神经网络    |  | - 根因排名      |         |
|  | - 因果推断      |  | - 影响范围      |         |
|  | - 关联规则      |  | - 修复建议      |         |
|  +----------------+  +----------------+         |
+--------------------------------------------------+
```

#### 5.2.2 根因分析流程

```
输入：异常事件（指标异常 + 日志异常 + 告警风暴）
  |
  +--> 时间窗口对齐（提取异常前后 5-15 分钟数据）
  |
  +--> 拓扑关联分析（基于 CMDB/服务网格获取关联实体）
  |
  +--> 异常传播路径追踪（调用链分析）
  |
  +--> 变更关联（检查异常时间窗口内的发布/变更记录）
  |
  +--> 历史模式匹配（与历史故障库比对相似模式）
  |
  +--> 根因评分排序（综合多维度证据计算根因概率）
  |
输出：根因列表（Top-N）+ 影响范围 + 修复建议
```

#### 5.2.3 根因分析 API 设计

**请求**

```json
POST /api/v1/rca/analyze
Content-Type: application/json

{
  "incident_id": "inc_2026081101",
  "anomaly_events": [
    {
      "type": "metric",
      "metric": "http_requests_duration_seconds",
      "service": "order-service",
      "timestamp": "2026-08-11T08:00:00Z"
    },
    {
      "type": "log",
      "pattern": "Connection timeout to database",
      "service": "order-service",
      "count": 150,
      "timestamp": "2026-08-11T08:00:00Z"
    }
  ],
  "time_window": {
    "start": "2026-08-11T07:55:00Z",
    "end": "2026-08-11T08:05:00Z"
  }
}
```

**响应**

```json
{
  "status": "success",
  "incident_id": "inc_2026081101",
  "root_causes": [
    {
      "rank": 1,
      "confidence": 0.88,
      "entity": {
        "type": "database",
        "name": "order-db-primary",
        "namespace": "production"
      },
      "evidence": [
        "数据库连接超时日志激增（150条/分钟）",
        "数据库CPU使用率从30%飙升至98%",
        "慢查询数量在08:00突增500%"
      ],
      "impact": {
        "affected_services": ["order-service", "payment-service", "inventory-service"],
        "affected_users": "约 12,000 用户",
        "severity": "P1"
      },
      "suggested_actions": [
        "检查数据库慢查询日志，定位耗时SQL",
        "评估是否需要扩容数据库连接池",
        "考虑启用读写分离或数据库主从切换"
      ]
    }
  ]
}
```

### 5.3 告警降噪模块

#### 5.3.1 降噪策略

| 策略             | 说明                       | 预期效果           |
| ---------------- | -------------------------- | ------------------ |
| **时间窗口聚合** | 5 分钟内相同告警合并为一条 | 减少 40% 重复告警  |
| **相似度合并**   | 基于文本相似度合并同类告警 | 减少 30% 相似告警  |
| **拓扑抑制**     | 父节点故障抑制子节点告警   | 减少 50% 级联告警  |
| **动态阈值**     | 基于历史数据自适应调整阈值 | 减少 60% 阈值误报  |
| **智能分级**     | 基于影响面和紧急度自动定级 | 提升关键告警识别率 |

#### 5.3.2 告警生命周期

```
原始告警生成
     |
     +--> 夜莺规则引擎（静态阈值）
     |
     +--> AIOPS 异常检测算法（指标/日志/关联）
     |
     +--> 来源标记 / 去重（source=nightingale / source=aiops）
     |
     +--> 规则过滤（黑名单/白名单）
     |
     +--> 时间窗口聚合
     |
     +--> 相似度去重（文本向量化 + 聚类）
     |
     +--> 拓扑抑制（依赖关系分析）
     |
     +--> 动态阈值判断
     |
     +--> 智能分级（P0-P4）
     |
     +--> 通知路由（根据分级和值班表）
     |
     v
  有效告警通知
```

> **说明**：原始告警来自夜莺规则引擎与 AIOPS 异常检测两类来源，统一打上 `source` 标签后再进入降噪管道，避免同一故障被重复处理。

### 5.4 日志分析模块

#### 5.4.1 日志处理流水线

```
原始日志
   |
   +--> 日志解析（正则/Grok/自动模式识别）
   |
   +--> 模板提取（Drain 算法）
   |
   +--> 参数分离（常量模板 + 变量参数）
   |
   +--> 向量化（TF-IDF / Word2Vec / BERT Embedding）
   |
   +--> 聚类分析（K-Means / DBSCAN）
   |
   +--> 异常检测（孤立日志 / 频率异常 / 模式突变）
   |
   v
结构化日志事件
```

#### 5.4.2 Drain 日志模板提取

Drain 是一种在线日志解析算法，核心思想：

1. **固定深度树遍历**：按日志词序列构建解析树
2. **模板匹配**：新日志与已有模板比对，相似则合并
3. **在线学习**：无需预训练，实时更新模板库

**示例**

```
原始日志：
  "Connection to database order-db failed after 5000ms"
  "Connection to database user-db failed after 3000ms"

提取模板：
  "Connection to database <*> failed after <*>ms"

参数：
  ["order-db", "5000"]
  ["user-db", "3000"]
```

### 5.5 运维知识库模块

#### 5.5.1 模块定位

运维知识库是 AIOPS 的“记忆中枢”，沉淀历史故障处理经验、Runbook/SOP、系统文档、告警处置记录等结构化与非结构化知识。它向上为智能运维助手、根因分析、告警降噪提供检索与推理依据，向下通过向量化与语义检索实现高效召回。

#### 5.5.2 知识库内容来源

| 来源              | 内容示例                          | 接入方式                  |
| ----------------- | --------------------------------- | ------------------------- |
| 历史 RCA 报告     | 故障现象、根因、影响范围、修复动作 | 工单/事件系统 API 导入    |
| Runbook / SOP     | 标准排查步骤、应急操作手册        | 文档上传 / 在线文档同步   |
| 运维文档 & FAQ    | 系统架构说明、常见错误排查        | Markdown / Word / PDF 上传 |
| 告警处置反馈      | 人工确认告警后的处理备注          | 告警管理模块自动沉淀      |
| 日志/指标异常模式 | 已知异常模式说明与处理建议        | 日志/异常检测模块自动写入 |

#### 5.5.3 知识处理流水线

```
原始文档 / 记录
      |
      +--> 文档解析（PDF/Word/Markdown/网页）
      |
      +--> 文本清洗与分段（按语义/标题/固定长度）
      |
      +--> 向量化（Embedding 模型：BGE / M3E / text2vec）
      |
      +--> 向量索引（Vector DB：Milvus / Qdrant / pgvector）
      |
      +--> 元数据关联（服务、标签、版本、来源、权限）
      |
      v
可检索知识单元（Chunk）
```

#### 5.5.4 检索与召回策略

- **语义检索**：基于用户问题向量，在向量数据库中召回 Top-K 相关文本片段
- **关键词检索**：作为语义检索的补充，提升专有名词、命令行、错误码的召回精度
- **混合召回**：语义得分 + BM25 得分加权融合，返回最相关片段
- **过滤与排序**：支持按服务、标签、时间、来源过滤；按相似度与发布时间排序
- **引用溯源**：每个返回片段必须附带来源文档 ID、标题、段落位置、相似度得分

#### 5.5.5 知识库管理 API 设计

**上传文档**

```json
POST /api/v1/knowledge/documents
Content-Type: multipart/form-data

{
  "file": "database-troubleshooting.md",
  "metadata": {
    "source": "runbook",
    "service": "order-service",
    "tags": ["mysql", "timeout"]
  }
}
```

**语义检索**

```json
POST /api/v1/knowledge/retrieve
Content-Type: application/json

{
  "query": "order-service 数据库连接超时如何排查？",
  "top_k": 5,
  "filters": {
    "service": "order-service",
    "source": ["runbook", "rca"]
  }
}
```

**响应示例**

```json
{
  "status": "success",
  "results": [
    {
      "chunk_id": "chunk_abc123",
      "document_id": "doc_runbook_001",
      "title": "MySQL 连接超时排查手册",
      "content": "1. 检查数据库连接池配置；2. 查看慢查询日志...",
      "score": 0.89,
      "metadata": {
        "source": "runbook",
        "service": "order-service"
      }
    }
  ]
}
```

### 5.6 智能运维助手模块

#### 5.6.1 模块定位

智能运维助手是基于大语言模型（LLM）的对话式运维入口，面向运维人员提供告警解读、排障建议、知识查询、操作辅助等能力。助手在回答问题时必须优先检索运维知识库，基于检索结果生成可解释、可溯源的回答，避免模型幻觉。

#### 5.6.2 RAG 增强回答流程

```
用户问题
   |
   +--> 意图识别（闲聊 / 告警解读 / 排障建议 / 知识查询）
   |
   +--> 查询改写（结合多轮对话上下文生成检索 query）
   |
   +--> 知识库检索（调用 /api/v1/knowledge/retrieve）
   |
   +--> 上下文拼接（检索结果 + 对话历史 + 系统 Prompt）
   |
   +--> LLM 生成回答
   |
   +--> 后处理（格式化、引用标注、安全审查）
   |
   v
带引用来源的回答
```

#### 5.6.3 回答引用规范

- 每条事实性建议必须标注来源知识库条目 `[source: chunk_id]`
- 若知识库未命中相关内容，助手应明确说明“未在知识库中找到相关记录”，并给出通用性建议
- 支持点击引用跳转至原始文档或 RCA 报告

#### 5.6.4 智能运维助手 API 设计

**对话请求**

```json
POST /api/v1/assistant/chat
Content-Type: application/json

{
  "session_id": "sess_2026081301",
  "message": "order-service 数据库连接超时怎么排查？",
  "context": {
    "current_alert_id": "alert_2026081301",
    "service": "order-service"
  }
}
```

**对话响应**

```json
{
  "status": "success",
  "session_id": "sess_2026081301",
  "message": "根据知识库中的《MySQL 连接超时排查手册》[source: chunk_abc123]，建议按以下步骤排查：...",
  "references": [
    {
      "chunk_id": "chunk_abc123",
      "title": "MySQL 连接超时排查手册",
      "score": 0.89,
      "url": "/knowledge/doc_runbook_001"
    }
  ],
  "retrieval_summary": {
    "query": "order-service MySQL 连接超时 排查",
    "hit_count": 3
  }
}
```

---

## 6. 技术栈选型

### 6.1 完整技术栈

| 层级           | 组件             | 版本建议      | 用途                                    |
| -------------- | ---------------- | ------------- | --------------------------------------- |
| **网关层**     | Go + Gin         | v1.9+         | API 网关，请求路由、鉴权、限流          |
| **业务服务**   | Go + GORM        | v2.0+         | 业务逻辑、数据持久化                    |
| **算法服务**   | Python + FastAPI | 0.100+        | ML 模型服务化，异步高性能               |
| **任务调度**   | Python + Celery  | 5.3+          | 异步任务、定时任务、模型训练            |
| **时序数据库** | VictoriaMetrics  | v1.90+        | 兼容 Prometheus，高性能时序存储         |
| **日志存储**   | Elasticsearch    | 7.x/8.x       | 已有 ELK 基础设施复用                   |
| **关系数据库** | PostgreSQL       | 14+           | 业务数据、元数据、配置                  |
| **缓存**       | Redis            | 7.0+          | 告警状态、会话、热点数据、Celery Broker |
| **向量数据库** | Milvus / Qdrant  | 2.3+ / 1.7+   | 知识库文本向量存储与语义检索            |
| **消息队列**   | Kafka            | 3.5+          | 高吞吐指标/日志流、事件总线             |
| **对象存储**   | MinIO            | RELEASE.2024+ | 模型文件、训练数据、知识库原始文档快照  |
| **容器编排**   | Kubernetes       | 1.28+         | 服务部署、弹性伸缩、服务发现            |
| **服务网格**   | Istio (可选)     | 1.20+         | 微服务治理、流量管理、可观测性          |
| **可观测性**   | Grafana + Jaeger | 10.x / 1.50+  | 监控大盘、链路追踪                      |
| **模型服务**   | MLflow (可选)    | 2.10+         | 模型版本管理、实验追踪                  |

### 6.2 Go 依赖库

```go
// go.mod 核心依赖
require (
    github.com/gin-gonic/gin v1.9.1           // Web 框架
    github.com/prometheus/client_golang v1.17.0 // Prometheus 客户端
    github.com/olivere/elastic/v7 v7.0.32     // Elasticsearch 客户端
    github.com/segmentio/kafka-go v0.4.44    // Kafka 客户端
    github.com/redis/go-redis/v9 v9.3.0       // Redis 客户端
    gorm.io/gorm v1.25.5                      // ORM
    gorm.io/driver/postgres v1.5.4            // PostgreSQL 驱动
    google.golang.org/grpc v1.59.0            // gRPC
    github.com/casbin/casbin/v2 v2.79.0       // 权限管理
    github.com/robfig/cron/v3 v3.0.1          // 定时任务
    github.com/spf13/viper v1.18.0            // 配置管理
    go.uber.org/zap v1.26.0                   // 日志
)
```

### 6.3 Python 依赖库

```python
# requirements.txt 核心依赖
fastapi==0.104.1              # Web 框架
uvicorn[standard]==0.24.0     # ASGI 服务器
celery==5.3.4                 # 异步任务
redis==5.0.1                  # Redis 客户端
kafka-python==2.0.2           # Kafka 客户端

# 数据科学
numpy==1.24.3
pandas==2.0.3
scikit-learn==1.3.2
scipy==1.11.4

# 时序分析
prophet==1.1.5                # 时序预测
statsmodels==0.14.1           # 统计模型

# 深度学习 (按需安装)
torch==2.1.1                  # PyTorch
transformers==4.36.0          # Hugging Face Transformers

# 日志分析
drain3==0.9.11                # Drain 日志解析

# 图计算 (根因分析)
neo4j==5.15.0                 # Neo4j 图数据库客户端
networkx==3.2.1               # 图算法

# 知识库与 RAG
sentence-transformers==2.2.2  # 文本 Embedding 模型
langchain==0.1.0              # RAG 流程编排（可选）
pymilvus==2.3.4               # Milvus 向量数据库客户端
qdrant-client==1.7.0          # Qdrant 向量数据库客户端
openai==1.6.0                 # LLM API 调用（或兼容 OpenAI 协议的本地模型）

# 可观测性
prometheus-client==0.19.0     # 指标暴露
```

---

## 7. 实施路线图

### 7.1 阶段划分

```
Phase 1 (1-2月)        Phase 2 (2-3月)        Phase 3 (3-4月)        Phase 4 (持续)
     |                       |                       |                      |
     v                       v                       v                      v
+----------+            +----------+            +----------+           +----------+
| 基础平台  |            | 智能检测  |            | 高级特性  |           | 持续优化  |
| 搭建     |            | 上线     |            | 开发     |           | 迭代     |
+----------+            +----------+            +----------+           +----------+
     |                       |                       |                      |
- 服务框架搭建          - 单指标异常检测         - 多指标关联检测        - 模型在线学习
- 数据采集接入          - 日志模板提取           - 根因分析引擎          - 自动化修复
- 基础告警规则          - 日志异常检测           - 告警智能降噪          - 多租户与 SaaS 化
- 告警通知渠道          - 时序预测               - 容量规划              - 知识库持续运营
- 可视化大盘            - 动态阈值               - 故障演练              - 助手能力迭代
-                       -                       - 运维知识库             -
-                       -                       - 智能运维助手           -
```

### 7.2 详细里程碑

#### Phase 1：基础平台搭建（第 1-2 月）

| 周次  | 任务                    | 交付物                         | 负责人 |
| ----- | ----------------------- | ------------------------------ | ------ |
| W1-W2 | 技术选型确认 & 环境搭建 | 技术方案评审通过、K8s 集群就绪 | 架构组 |
| W2-W3 | Go 基础服务框架开发     | 用户/权限/配置服务上线         | 后端组 |
| W3-W4 | 数据采集器开发          | Prometheus/ELK 采集器上线      | 后端组 |
| W4-W6 | 夜莺规则引擎接入 & 告警管理 & 通知中心 | 支持夜莺 webhook 接入、告警标准化、多渠道通知 | 后端组 |
| W6-W8 | 可视化大盘开发          | Grafana 大盘/自研看板上线      | 前端组 |

#### Phase 2：智能检测上线（第 3-5 月）

| 周次    | 任务                | 交付物                        | 负责人 |
| ------- | ------------------- | ----------------------------- | ------ |
| W9-W10  | Python 算法服务框架 | FastAPI 服务、模型加载机制    | 算法组 |
| W10-W12 | 单指标异常检测      | Prophet/Isolation Forest 上线 | 算法组 |
| W12-W14 | 日志分析模块        | Drain 模板提取、日志聚类上线  | 算法组 |
| W14-W16 | 时序预测模块        | 容量预测、趋势预测上线        | 算法组 |
| W16-W18 | 动态阈值模块        | 自适应阈值策略上线            | 算法组 |

#### Phase 3：高级特性开发（第 6-9 月）

| 周次    | 任务           | 交付物                     | 负责人 |
| ------- | -------------- | -------------------------- | ------ |
| W19-W22 | 多指标关联检测 | VAE/GMM 关联异常检测上线   | 算法组 |
| W22-W26 | 根因分析引擎   | 知识图谱构建、GNN 推理上线 | 算法组 |
| W26-W30 | 告警智能降噪   | 聚合/抑制/分级策略上线     | 算法组 |
| W30-W34 | 容量规划系统   | 资源预测、扩缩容建议上线   | 算法组 |
| W32-W36 | 运维知识库     | 文档上传、向量化、语义检索上线 | 算法组 |
| W34-W36 | 故障演练平台   | 混沌工程集成、演练自动化   | 运维组 |
| W36-W38 | 智能运维助手   | LLM 接入、RAG 检索增强、对话上线 | 算法组 |

#### Phase 4：持续优化（第 10 月起）

- 模型在线学习与反馈闭环
- Runbook 自动化执行
- 知识库持续运营与质量评估
- 智能运维助手能力迭代（多模态、工具调用）
- 多租户与 SaaS 化支持

---

## 8. 接口设计

### 8.1 内部服务接口规范

所有内部服务接口遵循 RESTful + gRPC 双协议设计：

- **RESTful API**：面向前端和外部系统，HTTP/JSON
- **gRPC**：服务间内部通信，高性能二进制协议

### 8.2 核心接口列表

#### 8.2.1 数据采集接口

| 接口                      | 方法 | 说明                              |
| ------------------------- | ---- | --------------------------------- |
| `/api/v1/collect/metrics` | POST | 接收 Prometheus remote_write 数据 |
| `/api/v1/collect/logs`    | POST | 接收结构化日志数据                |
| `/api/v1/collect/query`   | GET  | 查询采集状态与统计                |

#### 8.2.2 异常检测接口

| 接口                       | 方法 | 说明             |
| -------------------------- | ---- | ---------------- |
| `/api/v1/anomaly/detect`   | POST | 执行异常检测     |
| `/api/v1/anomaly/batch`    | POST | 批量异常检测     |
| `/api/v1/anomaly/models`   | GET  | 获取可用模型列表 |
| `/api/v1/anomaly/feedback` | POST | 提交检测反馈     |

#### 8.2.3 根因分析接口

| 接口                   | 方法 | 说明             |
| ---------------------- | ---- | ---------------- |
| `/api/v1/rca/analyze`  | POST | 执行根因分析     |
| `/api/v1/rca/topology` | GET  | 获取服务拓扑图   |
| `/api/v1/rca/history`  | GET  | 查询历史根因案例 |

#### 8.2.4 告警管理接口

| 接口                          | 方法 | 说明         |
| ----------------------------- | ---- | ------------ |
| `/api/v1/alerts`                      | GET  | 查询告警列表 |
| `/api/v1/alerts/{id}/ack`             | POST | 确认告警     |
| `/api/v1/alerts/{id}/resolve`         | POST | 解决告警     |
| `/api/v1/alerts/rules`                | GET  | 告警规则聚合展示（含夜莺规则） |
| `/api/v1/alerts/policies`             | CRUD | 降噪策略管理 |
| `/api/v1/alerts/webhook/nightingale`  | POST | 接收夜莺 webhook 告警 |
| `/api/v1/alerts/external/rules`       | GET  | 查询夜莺告警规则 |
| `/api/v1/alerts/external/events`      | GET  | 查询夜莺历史告警事件 |

#### 8.2.5 知识库接口

| 接口                                  | 方法   | 说明                     |
| ------------------------------------- | ------ | ------------------------ |
| `/api/v1/knowledge/documents`         | POST   | 上传知识库文档           |
| `/api/v1/knowledge/documents`         | GET    | 查询文档列表             |
| `/api/v1/knowledge/documents/{id}`    | GET    | 查询文档详情             |
| `/api/v1/knowledge/documents/{id}`    | DELETE | 删除文档                 |
| `/api/v1/knowledge/retrieve`          | POST   | 语义检索知识库片段       |
| `/api/v1/knowledge/feedback`          | POST   | 提交检索结果反馈         |

#### 8.2.6 智能运维助手接口

| 接口                                  | 方法 | 说明                     |
| ------------------------------------- | ---- | ------------------------ |
| `/api/v1/assistant/chat`              | POST | 发起对话/提问            |
| `/api/v1/assistant/sessions`          | GET  | 查询会话列表             |
| `/api/v1/assistant/sessions/{id}`     | GET  | 查询会话历史             |
| `/api/v1/assistant/sessions/{id}`     | DELETE | 删除会话               |

### 8.3 数据流接口

```
Prometheus --remote_write--> Go Collector --Kafka--> Python Analyzer
                                                        |
Elasticsearch <--query-- Go Collector --Kafka----------+
                                                        |
                                                    Kafka
                                                        |
                              Go Business Service <--consume
                                        |
                                   PostgreSQL/Redis

Nightingale --webhook--> AIOPS Gateway --> Go Alert Service --> PostgreSQL/Redis
      ^                                                        |
      +-------------------- API 查询 --------------------------+
```

---

## 9. 部署架构

### 9.1 Kubernetes 部署拓扑

```
Namespace: aiops-production

+-------------------------------------------------------------+
|                      Ingress Controller                       |
|              (Nginx / Traefik / Istio Gateway)                |
+-------------------------------------------------------------+
                              |
+-------------------------------------------------------------+
|  Deployment: aiops-gateway (Go/Gin)                       |
|  Replicas: 3                                                |
|  Resources: 2 CPU / 4GB RAM per pod                         |
+-------------------------------------------------------------+
                              |
        +-------------------+-------------------+
        |                   |                   |
+---------------+   +---------------+   +---------------+
|  Deployment:  |   |  Deployment:  |   |  Deployment:  |
|  aiops-api    |   |  aiops-collector| |  aiops-alert  |
|  (Go)         |   |  (Go)         |   |  (Go)         |
|  Replicas: 5  |   |  Replicas: 3  |   |  Replicas: 3  |
+---------------+   +---------------+   +---------------+
        |                   |                   |
        +-------------------+-------------------+
                              |
+-------------------------------------------------------------+
|  Deployment: aiops-analyzer (Python/FastAPI)                |
|  Replicas: 5 (HPA: 5-20 based on GPU/CPU)                   |
|  Resources: 4 CPU / 16GB RAM / optional GPU                 |
+-------------------------------------------------------------+
                              |
        +-------------------+-------------------+
        |                   |                   |
+---------------+   +---------------+   +---------------+
|  StatefulSet: |   |  StatefulSet: |   |  Deployment:  |
|  kafka        |   |  postgres     |   |  redis        |
|  (3 brokers)  |   |  (primary+replica)| |  (sentinel)   |
+---------------+   +---------------+   +---------------+
        |                   |                   |
+---------------+   +---------------+   +---------------+
|  StatefulSet: |   |  Deployment:  |   |  Deployment:  |
|  victoriametrics| |  elasticsearch|   |  minio        |
|  (cluster)    |   |  (3 nodes)    |   |  (distributed)|
+---------------+   +---------------+   +---------------+
```

### 9.2 资源规划

| 组件            | 实例数 | CPU/实例 | 内存/实例 | 存储  | 备注         |
| --------------- | ------ | -------- | --------- | ----- | ------------ |
| aiops-gateway   | 3      | 2        | 4GB       | -     | 无状态       |
| aiops-api       | 5      | 2        | 4GB       | -     | 无状态       |
| aiops-collector | 3      | 4        | 8GB       | -     | 高 I/O       |
| aiops-analyzer  | 5-20   | 4        | 16GB      | 100GB | HPA 弹性伸缩 |
| Kafka           | 3      | 4        | 8GB       | 500GB | SSD          |
| PostgreSQL      | 2      | 4        | 8GB       | 200GB | 主从         |
| Redis           | 3      | 2        | 8GB       | -     | Sentinel     |
| VictoriaMetrics | 3      | 8        | 16GB      | 2TB   | SSD          |
| Elasticsearch   | 3      | 8        | 32GB      | 1TB   | SSD          |
| MinIO           | 4      | 4        | 8GB       | 5TB   | 对象存储     |

---

## 10. 附录

### 10.1 术语表

| 术语  | 英文                                      | 说明              |
| ----- | ----------------------------------------- | ----------------- |
| AIOPS | Artificial Intelligence for IT Operations | 智能运维          |
| MTTR  | Mean Time To Repair                       | 平均修复时间      |
| MTBF  | Mean Time Between Failures                | 平均故障间隔时间  |
| SLA   | Service Level Agreement                   | 服务等级协议      |
| CMDB  | Configuration Management Database         | 配置管理数据库    |
| RCA   | Root Cause Analysis                       | 根因分析          |
| SLO   | Service Level Objective                   | 服务等级目标      |
| SRE   | Site Reliability Engineering              | 站点可靠性工程    |
| HPA   | Horizontal Pod Autoscaler                 | 水平 Pod 自动伸缩 |
| VAE   | Variational Autoencoder                   | 变分自编码器      |
| GNN   | Graph Neural Network                      | 图神经网络        |
| GMM   | Gaussian Mixture Model                    | 高斯混合模型      |
| RAG   | Retrieval-Augmented Generation            | 检索增强生成，通过检索外部知识提升 LLM 回答准确性 |
| Embedding | Embedding                               | 嵌入向量，将文本/数据映射到高维语义空间 |
| Vector DB | Vector Database                         | 向量数据库，专门存储与检索高维向量 |
| LLM   | Large Language Model                      | 大语言模型        |
| Runbook | Runbook                                 | 标准操作流程手册，记录故障排查与处置步骤 |

### 10.2 参考资源

- [Prometheus Remote Write Protocol](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#remote_write)
- [VictoriaMetrics Documentation](https://docs.victoriametrics.com/)
- [Elasticsearch Go Client](https://github.com/olivere/elastic)
- [FastAPI Documentation](https://fastapi.tiangolo.com/)
- [Drain3 Log Parser](https://github.com/IBM/drain3)
- [Prophet Documentation](https://facebook.github.io/prophet/)

### 10.3 文档修订记录

| 版本 | 日期       | 修订内容 | 修订人       |
| ---- | ---------- | -------- | ------------ |
| v1.2 | 2026-08-13 | 新增运维知识库模块与智能运维助手（RAG）设计 | AIOPS 项目组 |
| v1.1 | 2026-08-12 | 引入夜莺（Nightingale）作为外部规则引擎 | AIOPS 项目组 |
| v1.0 | 2026-08-11 | 初始版本 | AIOPS 项目组 |

---

> **文档说明**：本文档为 AIOPS 系统设计的总体方案，各模块的详细设计将在后续迭代中补充。如有疑问或建议，请联系项目组。