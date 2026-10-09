-- =============================================================
-- AIOPS 智能运维平台 · 全量初始化脚本
-- 数据库：MySQL 8.x  (aiops)
-- 说明：
--   1. 覆盖后端 GORM AutoMigrate 中的全部 31 张表（与 backend/models/ 一一对应）
--   2. 全部幂等写法（CREATE TABLE IF NOT EXISTS + INSERT ... WHERE NOT EXISTS），可重复执行
--   3. 内置 root 管理员、Admin/Standard/Guest 角色与角色-权限绑定、全量菜单权限、默认钉钉/Webhook 通知模板
--   4. 后端登录为明文密码比对（controllers/auth.go Login），root 密码固定 root
-- =============================================================

CREATE DATABASE IF NOT EXISTS `aiops` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `aiops`;

-- =============================================================
-- 一、用户 / 角色 / 权限（系统基础表）
-- =============================================================

CREATE TABLE IF NOT EXISTS `sys_user` (
  `id` varchar(40) NOT NULL COMMENT 'id',
  `username` varchar(30) NOT NULL COMMENT '用户名',
  `password` varchar(30) NOT NULL COMMENT '密码',
  `name` varchar(10) NOT NULL COMMENT '姓名',
  `created_at` datetime(3) NULL COMMENT '创建时间',
  `deleted_at` datetime(3) NULL COMMENT '删除时间',
  `update_at` datetime(3) NULL COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户表';

CREATE TABLE IF NOT EXISTS `sys_role` (
  `id` varchar(40) NOT NULL COMMENT 'id',
  `name` varchar(30) NULL COMMENT '名称',
  `created_at` datetime(3) NULL COMMENT '创建时间',
  `updated` datetime(3) NULL COMMENT '更新时间',
  `deleted_at` datetime(3) NULL COMMENT '删除时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='角色表';

CREATE TABLE IF NOT EXISTS `sys_auth` (
  `id` varchar(40) NOT NULL COMMENT '',
  `name` varchar(50) NULL COMMENT '名称',
  `type` bigint(20) NOT NULL COMMENT '1-菜单 2-操作',
  `created_at` datetime(3) NULL COMMENT '创建时间',
  `updated` datetime(3) NULL COMMENT '更新时间',
  `deleted_at` datetime(3) NULL COMMENT '删除时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='权限';

CREATE TABLE IF NOT EXISTS `sys_user_role_relation` (
  `id` varchar(40) NOT NULL COMMENT '',
  `user_id` varchar(40) NOT NULL COMMENT '用户id',
  `role_id` varchar(40) NOT NULL COMMENT '角色id',
  `created_at` datetime(3) NULL COMMENT '创建时间',
  `updated_at` datetime(3) NULL COMMENT '更新时间',
  `deleted_at` datetime(3) NULL COMMENT '删除时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户角色关联表';

CREATE TABLE IF NOT EXISTS `sys_role_auth_relation` (
  `id` varchar(40) NOT NULL COMMENT 'id',
  `auth_id` varchar(40) NOT NULL COMMENT '权限id',
  `role_id` varchar(40) NOT NULL COMMENT '角色id',
  `role_name` varchar(40) NULL COMMENT '角色名称',
  `auth_name` varchar(40) NULL COMMENT '权限名称',
  `created_at` datetime(3) NULL COMMENT '创建时间',
  `deleted_at` datetime(3) NULL COMMENT '删除时间',
  `updated_at` datetime(3) NULL COMMENT '更新时间',
  `updated` datetime(3) NULL COMMENT '更新时间',
  `roleId` varchar(40) NOT NULL COMMENT '角色id(冗余列)',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='角色-权限关联表';

-- =============================================================
-- 二、对话模块
-- =============================================================

CREATE TABLE IF NOT EXISTS `chat_session` (
  `id` varchar(36) NOT NULL COMMENT '会话ID',
  `user_id` varchar(36) NOT NULL COMMENT '用户ID',
  `title` varchar(100) NOT NULL DEFAULT '新会话' COMMENT '会话标题',
  `status` varchar(20) NOT NULL DEFAULT 'active' COMMENT '状态 active/archived',
  `created_at` datetime(3) NULL COMMENT '创建时间',
  `updated_at` datetime(3) NULL COMMENT '更新时间',
  `deleted_at` datetime(3) NULL COMMENT '删除时间（软删除）',
  PRIMARY KEY (`id`),
  KEY `idx_chat_session_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='';

CREATE TABLE IF NOT EXISTS `chat_message` (
  `id` varchar(36) NOT NULL COMMENT '消息ID',
  `session_id` varchar(36) NOT NULL COMMENT '会话ID',
  `role` varchar(20) NOT NULL COMMENT '角色 system/user/assistant',
  `content` longtext NOT NULL COMMENT '消息内容',
  `prompt_tokens` bigint(20) NULL DEFAULT 0 COMMENT '输入token数',
  `completion_tokens` bigint(20) NULL DEFAULT 0 COMMENT '输出token数',
  `total_tokens` bigint(20) NULL DEFAULT 0 COMMENT '总token数',
  `created_at` datetime(3) NULL COMMENT '创建时间',
  `updated_at` datetime(3) NULL COMMENT '更新时间',
  `deleted_at` datetime(3) NULL COMMENT '删除时间（软删除）',
  PRIMARY KEY (`id`),
  KEY `idx_chat_message_session_id` (`session_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='';

-- =============================================================
-- 三、LLM / 数据源 / 告警引擎 / 系统配置
-- =============================================================

CREATE TABLE IF NOT EXISTS `llm_config` (
  `id` varchar(40) NOT NULL COMMENT 'id',
  `name` varchar(30) NULL COMMENT 'LLM 名称',
  `description` varchar(100) NULL COMMENT '描述',
  `supplier_category` varchar(10) NULL COMMENT '供应商类型 openai/claude/gemini等',
  `model` varchar(50) NULL COMMENT '模型名称',
  `base_url` varchar(200) NULL COMMENT 'api url',
  `api_key` varchar(512) NULL COMMENT '密钥（AES-256-GCM 加密存储，enc:v1: 前缀）',
  `is_default` bigint(20) NOT NULL DEFAULT 0 COMMENT '0 否，1是',
  `is_enabled` bigint(20) NULL DEFAULT 1 COMMENT '是否启用 0-否 1-是',
  `created_at` datetime(3) NULL COMMENT '创建时间',
  `deleted_at` datetime(3) NULL COMMENT '删除时间',
  `updated_at` datetime(3) NULL COMMENT '更新时间',
  `model_type` varchar(10) NOT NULL DEFAULT 'chat' COMMENT '模型类型 chat-对话 embedding-向量化',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='大模型配置';

CREATE TABLE IF NOT EXISTS `datasources` (
  `id` varchar(40) NOT NULL COMMENT '',
  `name` varchar(30) NULL COMMENT '名称',
  `type` varchar(20) NULL COMMENT '类型 Prometheus/ElasticSearch/Pyroscope',
  `url` varchar(100) NOT NULL COMMENT 'http地址',
  `timeout` bigint(20) NULL COMMENT '超时(ms)',
  `username` varchar(100) NULL COMMENT '用户名',
  `password` varchar(100) NULL COMMENT '密码',
  `is_skip_ssl` bigint(20) NULL COMMENT '是否跳过SSL验证 0-否 1-是',
  `remark` varchar(100) NULL COMMENT '备注',
  `is_enabled` bigint(20) NULL COMMENT '状态 0-不启用 1-启用',
  `created_at` datetime(3) NULL COMMENT '创建时间',
  `updated_at` datetime(3) NULL COMMENT '更新时间',
  `deleted_at` datetime(3) NULL COMMENT '删除时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='elasticsearch数据源';

CREATE TABLE IF NOT EXISTS `alert_engine_config` (
  `id` varchar(40) NOT NULL COMMENT 'id',
  `name` varchar(30) NOT NULL COMMENT '名称',
  `base_url` varchar(100) NOT NULL COMMENT '夜莺地址',
  `token` varchar(100) NOT NULL COMMENT '用户Token',
  `gids` varchar(100) NULL COMMENT '业务组ID,逗号分隔,空则使用Token可访问的全部组',
  `is_enabled` bigint(20) NULL DEFAULT 1 COMMENT '0-禁用 1-启用',
  `is_default` bigint(20) NULL DEFAULT 0 COMMENT '0-否 1-是',
  `remark` varchar(100) NULL COMMENT '备注',
  `created_at` datetime(3) NULL COMMENT '创建时间',
  `updated_at` datetime(3) NULL COMMENT '更新时间',
  `deleted_at` datetime(3) NULL COMMENT '删除时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='';

CREATE TABLE IF NOT EXISTS `n9e_config` (
  `id` varchar(40) NOT NULL COMMENT 'id',
  `name` varchar(64) NOT NULL DEFAULT '夜莺' COMMENT '',
  `address` varchar(255) NOT NULL COMMENT '夜莺服务地址，如 http://10.2.209.145:17000',
  `token` varchar(128) NOT NULL COMMENT 'X-User-Token',
  `is_enabled` bigint(20) NULL DEFAULT 1 COMMENT '是否启用',
  `created_at` datetime(3) NULL COMMENT '',
  `updated_at` datetime(3) NULL COMMENT '',
  `deleted_at` datetime(3) NULL COMMENT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='';

CREATE TABLE IF NOT EXISTS `system_configs` (
  `id` varchar(64) NOT NULL COMMENT 'id',
  `key` varchar(128) NOT NULL COMMENT '配置键',
  `value` varchar(512) NOT NULL COMMENT '配置值',
  `updated_at` datetime(3) NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_system_configs_key` (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='';

-- =============================================================
-- 四、告警规则 / 告警事件 / 根因分析 / 服务注册 / 业务分组 / 告警降噪
-- =============================================================

CREATE TABLE IF NOT EXISTS `alert_rules` (
  `id` varchar(64) NOT NULL COMMENT 'id',
  `name` varchar(128) NOT NULL COMMENT '规则名称',
  `prom_ql` text NOT NULL COMMENT 'PromQL 表达式',
  `duration` bigint(20) NULL COMMENT '持续时间(秒)',
  `severity` bigint(20) NULL COMMENT '告警级别 1-P1紧急 2-P2警告 3-P3提醒',
  `is_enabled` bigint(20) NULL COMMENT '启用状态 0-停用 1-启用',
  `created_at` datetime(3) NULL COMMENT '创建时间',
  `updated_at` datetime(3) NULL COMMENT '更新时间',
  `deleted_at` datetime(3) NULL COMMENT '删除时间',
  `eval_interval` bigint(20) NULL COMMENT '执行频率(秒)',
  `notify_rule_id` varchar(64) NULL COMMENT '通知规则ID',
  `repeat_interval_minutes` bigint(20) NULL DEFAULT 0 COMMENT '重复通知间隔(分钟)',
  `max_send_count` bigint(20) NULL DEFAULT 0 COMMENT '最大发送次数 0=不限制',
  `group_id` varchar(64) NULL COMMENT '业务分组ID',
  PRIMARY KEY (`id`),
  KEY `idx_alert_rules_group_id` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='';

CREATE TABLE IF NOT EXISTS `alert_events` (
  `id` varchar(64) NOT NULL COMMENT 'id',
  `rule_id` varchar(64) NULL COMMENT '关联告警规则ID',
  `rule_name` varchar(128) NULL COMMENT '规则名称（冗余，便于展示）',
  `severity` bigint(20) NULL COMMENT '告警级别 1-P1紧急 2-P2警告 3-P3提醒',
  `type` varchar(16) NULL COMMENT '事件类型 alert-告警事件 recovery-告警恢复事件',
  `status` varchar(16) NULL COMMENT '状态 firing-告警中 resolved-已恢复',
  `target_ident` varchar(256) NULL COMMENT '告警对象（如 instance）',
  `tags` text NULL COMMENT '标签序列化 k=v,...',
  `trigger_value` varchar(64) NULL COMMENT '触发值',
  `trigger_time` datetime(3) NULL COMMENT '触发时间',
  `recovered_at` datetime(3) NULL COMMENT '恢复时间',
  `created_at` datetime(3) NULL COMMENT '创建时间',
  `updated_at` datetime(3) NULL COMMENT '更新时间',
  `notify_count` bigint(20) NULL DEFAULT 0 COMMENT '已发送通知次数',
  `last_notified_at` datetime(3) NULL COMMENT '上次发送通知时间',
  `group_name` varchar(128) NULL COMMENT '业务分组名称（冗余，便于展示）',
  PRIMARY KEY (`id`),
  KEY `idx_alert_events_rule_id` (`rule_id`),
  KEY `idx_alert_events_status` (`status`),
  KEY `idx_alert_events_trigger_time` (`trigger_time`),
  KEY `idx_alert_events_type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='';

CREATE TABLE IF NOT EXISTS `root_cause_analyses` (
  `id` varchar(64) NOT NULL COMMENT 'id',
  `alert_event_id` varchar(64) NULL COMMENT '关联告警事件ID',
  `rule_name` varchar(128) NULL COMMENT '规则名称（冗余，便于展示）',
  `target_ident` varchar(256) NULL COMMENT '告警对象（如 instance）',
  `tags` text NULL COMMENT '标签序列化 k=v,...',
  `severity` bigint(20) NULL COMMENT '告警级别 1-P1紧急 2-P2警告 3-P3提醒',
  `trigger_time` datetime(3) NULL COMMENT '告警触发时间',
  `status` varchar(16) NULL COMMENT '状态 running/completed/failed',
  `error` text NULL COMMENT '失败原因',
  `result` text NULL COMMENT '结构化分析结果 JSON',
  `content` text NULL COMMENT '完整分析报告 Markdown',
  `created_at` datetime(3) NULL COMMENT '创建时间',
  `updated_at` datetime(3) NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_root_cause_analyses_alert_event_id` (`alert_event_id`),
  KEY `idx_root_cause_analyses_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='';

CREATE TABLE IF NOT EXISTS `services` (
  `id` varchar(40) NOT NULL COMMENT 'id',
  `name` varchar(50) NULL COMMENT '服务名称',
  `code` varchar(50) NULL COMMENT '服务编码 小写字母数字中划线',
  `es_datasource_id` varchar(40) NULL COMMENT 'ElasticSearch数据源ID',
  `es_index_patterns` text NULL COMMENT 'ES索引模式 JSON数组字符串',
  `prom_datasource_id` varchar(40) NULL COMMENT 'Prometheus数据源ID',
  `prom_job` varchar(100) NULL COMMENT 'Prometheus job',
  `prom_labels` text NULL COMMENT 'Prometheus标签选择器 JSON对象字符串',
  `pyroscope_app` varchar(100) NULL COMMENT 'Pyroscope应用名',
  `owner` varchar(50) NULL COMMENT '负责人',
  `level` varchar(20) NULL COMMENT '服务等级',
  `description` varchar(200) NULL COMMENT '描述',
  `status` bigint(20) NULL COMMENT '状态 1-启用 0-停用',
  `verified` bigint(20) NULL COMMENT '探测是否通过 1-是 0-否',
  `verify_message` text NULL COMMENT '探测结果信息',
  `last_verified_at` datetime(3) NULL COMMENT '最近探测时间',
  `created_at` datetime(3) NULL COMMENT '创建时间',
  `updated_at` datetime(3) NULL COMMENT '更新时间',
  `deleted_at` datetime(3) NULL COMMENT '删除时间',
  `health_status` bigint(20) NULL COMMENT '健康状态 0-未检查 1-正常 2-疑似下线',
  `health_message` text NULL COMMENT '健康检查信息',
  `last_check_at` datetime(3) NULL COMMENT '最近健康检查时间',
  `miss_count` bigint(20) NULL COMMENT '连续无数据次数',
  `parent_id` varchar(40) NULL COMMENT '父服务ID（拓扑抑制用，可选）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_services_code` (`code`),
  KEY `idx_services_parent_id` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='';

CREATE TABLE IF NOT EXISTS `business_groups` (
  `id` varchar(64) NOT NULL COMMENT 'id',
  `name` varchar(128) NOT NULL COMMENT '分组名称',
  `created_at` datetime(3) NULL COMMENT '创建时间',
  `updated_at` datetime(3) NULL COMMENT '更新时间',
  `deleted_at` datetime(3) NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_business_groups_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='';

CREATE TABLE IF NOT EXISTS `denoise_policies` (
  `id` varchar(64) NOT NULL COMMENT 'id',
  `strategy` varchar(64) NULL COMMENT '策略类型 window_aggregation/topology_suppression',
  `name` varchar(128) NULL COMMENT '策略名称',
  `description` varchar(256) NULL COMMENT '策略描述',
  `enabled` tinyint(1) NULL COMMENT '是否启用',
  `config` text NULL COMMENT '策略配置 JSON 文本（窗口策略存 window_minutes）',
  `created_at` datetime(3) NULL COMMENT '创建时间',
  `updated_at` datetime(3) NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_denoise_policies_strategy` (`strategy`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='';

CREATE TABLE IF NOT EXISTS `denoise_records` (
  `id` varchar(64) NOT NULL COMMENT 'id',
  `strategy` varchar(64) NULL COMMENT '拦截策略 window_aggregation/topology_suppression',
  `rule_id` varchar(64) NULL COMMENT '关联告警规则ID',
  `rule_name` varchar(128) NULL COMMENT '规则名称',
  `severity` bigint(20) NULL COMMENT '告警级别 1-P1紧急 2-P2警告 3-P3提醒',
  `target_ident` varchar(256) NULL COMMENT '告警对象',
  `tags` text NULL COMMENT '标签序列化 k=v,...',
  `created_at` datetime(3) NULL COMMENT '创建时间',
  `updated_at` datetime(3) NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_denoise_records_created_at` (`created_at`),
  KEY `idx_denoise_records_strategy` (`strategy`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='';

-- =============================================================
-- 五、巡检任务 / 巡检报告
-- =============================================================

CREATE TABLE IF NOT EXISTS `inspection_tasks` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '',
  `name` varchar(128) NOT NULL COMMENT '',
  `cron_expr` varchar(64) NOT NULL COMMENT '',
  `prompt` text NOT NULL COMMENT '',
  `enabled` tinyint(1) NULL DEFAULT 1 COMMENT '',
  `last_run_at` datetime(3) NULL COMMENT '',
  `next_run_at` datetime(3) NULL COMMENT '',
  `created_at` datetime(3) NULL COMMENT '',
  `updated_at` datetime(3) NULL COMMENT '',
  `deleted_at` datetime(3) NULL COMMENT '',
  `notify_media_ids` text NULL COMMENT '通知媒介ID列表JSON',
  `mode` varchar(16) NULL DEFAULT 'single' COMMENT '执行模式 single-单Agent multi-多维度',
  `dimensions` text NULL COMMENT '巡检维度JSON(multi模式)',
  PRIMARY KEY (`id`),
  KEY `idx_inspection_tasks_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='';

CREATE TABLE IF NOT EXISTS `inspection_reports` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '',
  `task_id` bigint(20) unsigned NULL COMMENT '',
  `task_name` varchar(128) NULL COMMENT '',
  `title` varchar(256) NOT NULL COMMENT '',
  `content` longtext NOT NULL COMMENT '',
  `summary` varchar(512) NULL COMMENT '',
  `score` bigint(20) NULL DEFAULT 0 COMMENT '',
  `status` varchar(32) NULL DEFAULT 'completed' COMMENT '',
  `error` varchar(512) NULL COMMENT '',
  `created_at` datetime(3) NULL COMMENT '',
  `deleted_at` datetime(3) NULL COMMENT '',
  PRIMARY KEY (`id`),
  KEY `idx_inspection_reports_deleted_at` (`deleted_at`),
  KEY `idx_inspection_reports_task_id` (`task_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='';

-- =============================================================
-- 六、通知媒介 / 模板 / 规则 / 记录
-- =============================================================

CREATE TABLE IF NOT EXISTS `notify_media` (
  `id` varchar(64) NOT NULL COMMENT 'id',
  `name` varchar(60) NOT NULL COMMENT '媒介名称',
  `type` varchar(20) NOT NULL COMMENT '类型 webhook/dingtalk',
  `config` text NULL COMMENT '媒介配置JSON',
  `remark` varchar(200) NULL COMMENT '备注',
  `is_enabled` bigint(20) NULL COMMENT '状态 0-停用 1-启用',
  `created_at` datetime(3) NULL COMMENT '创建时间',
  `updated_at` datetime(3) NULL COMMENT '更新时间',
  `deleted_at` datetime(3) NULL COMMENT '删除时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='';

CREATE TABLE IF NOT EXISTS `notify_template` (
  `id` varchar(64) NOT NULL COMMENT '',
  `name` varchar(64) NOT NULL COMMENT '',
  `description` varchar(255) NULL COMMENT '',
  `type` varchar(16) NOT NULL DEFAULT 'all' COMMENT '',
  `content` text NOT NULL COMMENT '',
  `variables_hint` text NULL COMMENT '',
  `is_enabled` bigint(20) NULL DEFAULT 1 COMMENT '',
  `created_at` datetime(3) NULL COMMENT '',
  `updated_at` datetime(3) NULL COMMENT '',
  `deleted_at` datetime(3) NULL COMMENT '',
  `media_type` varchar(32) NOT NULL DEFAULT 'dingtalk' COMMENT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='';

CREATE TABLE IF NOT EXISTS `notify_rule` (
  `id` varchar(64) NOT NULL COMMENT '',
  `name` varchar(64) NOT NULL COMMENT '',
  `remark` varchar(255) NULL COMMENT '',
  `media_id` varchar(64) NULL COMMENT '',
  `template_id` varchar(64) NULL COMMENT '',
  `trigger_types` varchar(16) NULL DEFAULT 'all' COMMENT '',
  `severity_filter` varchar(128) NULL COMMENT '',
  `is_enabled` bigint(20) NULL DEFAULT 1 COMMENT '',
  `created_at` datetime(3) NULL COMMENT '',
  `updated_at` datetime(3) NULL COMMENT '',
  `deleted_at` datetime(3) NULL COMMENT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='';

CREATE TABLE IF NOT EXISTS `notify_records` (
  `id` varchar(64) NOT NULL COMMENT 'id',
  `event_id` varchar(64) NULL COMMENT '关联告警事件ID',
  `rule_id` varchar(64) NULL COMMENT '关联告警规则ID',
  `rule_name` varchar(128) NULL COMMENT '规则名称',
  `severity` bigint(20) NULL COMMENT '告警级别 1-P1紧急 2-P2警告 3-P3提醒',
  `event_type` varchar(16) NULL COMMENT '事件类型 alert/recovery',
  `target_ident` varchar(256) NULL COMMENT '告警对象',
  `tags` text NULL COMMENT '标签序列化 k=v,...',
  `trigger_value` varchar(64) NULL COMMENT '触发值',
  `trigger_time` datetime(3) NULL COMMENT '触发时间',
  `status` varchar(16) NULL COMMENT '状态 success/failed/intercepted/skipped',
  `strategy` varchar(64) NULL COMMENT '拦截策略（intercepted 时）window_aggregation/topology_suppression',
  `reason` varchar(512) NULL COMMENT '原因说明',
  `notify_rule_name` varchar(128) NULL COMMENT '通知规则名称',
  `media_name` varchar(128) NULL COMMENT '通知媒介名称',
  `media_type` varchar(32) NULL COMMENT '通知媒介类型',
  `template_name` varchar(128) NULL COMMENT '消息模板名称',
  `content` text NULL COMMENT '渲染后的消息正文',
  `created_at` datetime(3) NULL COMMENT '创建时间',
  `updated_at` datetime(3) NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_notify_records_created_at` (`created_at`),
  KEY `idx_notify_records_event_id` (`event_id`),
  KEY `idx_notify_records_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='';

-- =============================================================
-- 七、知识库（元数据表，向量存 Qdrant）
-- =============================================================

CREATE TABLE IF NOT EXISTS `kb_document` (
  `id` varchar(40) NOT NULL COMMENT 'id',
  `name` varchar(200) NOT NULL COMMENT '文件名',
  `type` varchar(10) NOT NULL COMMENT '扩展名',
  `size` bigint(20) NULL COMMENT '文件大小(字节)',
  `status` varchar(10) NOT NULL DEFAULT 'pending' COMMENT '状态 pending/indexing/indexed/failed',
  `error_msg` varchar(500) NULL COMMENT '失败原因',
  `uploader` varchar(30) NULL COMMENT '上传人',
  `file_path` varchar(300) NULL COMMENT '文件存储路径',
  `created_at` datetime(3) NULL COMMENT '创建时间',
  `updated_at` datetime(3) NULL COMMENT '更新时间',
  `deleted_at` datetime(3) NULL COMMENT '删除时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='';

CREATE TABLE IF NOT EXISTS `kb_chunk` (
  `id` varchar(40) NOT NULL COMMENT 'id',
  `document_id` varchar(40) NOT NULL COMMENT '文档id',
  `chunk_index` bigint(20) NULL COMMENT '块序号',
  `content` text NULL COMMENT '块文本',
  `created_at` datetime(3) NULL COMMENT '创建时间',
  `deleted_at` datetime(3) NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_kb_chunk_document_id` (`document_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='';

-- =============================================================
-- 八、日志分析（Drain 模板 / 聚类 / 分桶统计 / 分析游标）
-- =============================================================

CREATE TABLE IF NOT EXISTS `log_templates` (
  `id` varchar(64) NOT NULL COMMENT 'id',
  `service_id` varchar(64) NULL COMMENT '关联服务ID',
  `template` text NULL COMMENT 'Drain模板 如 Connection to database <*> failed after <*>ms',
  `level` varchar(16) NULL COMMENT '级别 error/warn/info',
  `count` bigint(20) NULL COMMENT '累计条数',
  `sample_raw` text NULL COMMENT '最近一条原始日志',
  `sample_params` text NULL COMMENT '最近样本JSON数组 [{raw,params}] 最多保留3条',
  `first_seen_at` datetime(3) NULL COMMENT '首次出现',
  `last_seen_at` datetime(3) NULL COMMENT '最近出现',
  `created_at` datetime(3) NULL COMMENT '创建时间',
  `updated_at` datetime(3) NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_log_templates_service_id` (`service_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='';

CREATE TABLE IF NOT EXISTS `log_clusters` (
  `id` varchar(64) NOT NULL COMMENT 'id',
  `service_id` varchar(64) NULL COMMENT '关联服务ID',
  `template_id` varchar(64) NULL COMMENT '主模板ID',
  `code` varchar(16) NULL COMMENT '聚类短码 如 C-018',
  `pattern` text NULL COMMENT '展示用模式（=模板）',
  `level` varchar(16) NULL COMMENT '级别 error/warn/info',
  `count` bigint(20) NULL COMMENT '当前窗口累计条数',
  `trend` varchar(16) NULL COMMENT '趋势 spike/rising/flat',
  `first_seen_at` datetime(3) NULL COMMENT '首次出现',
  `last_seen_at` datetime(3) NULL COMMENT '最近出现',
  `created_at` datetime(3) NULL COMMENT '创建时间',
  `updated_at` datetime(3) NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_log_clusters_service_id` (`service_id`),
  KEY `idx_log_clusters_trend` (`trend`),
  UNIQUE KEY `uk_log_cluster_code` (`service_id`,`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='';

CREATE TABLE IF NOT EXISTS `log_cluster_stats` (
  `id` varchar(64) NOT NULL COMMENT 'id',
  `cluster_id` varchar(64) NULL COMMENT '聚类ID',
  `service_id` varchar(64) NULL COMMENT '关联服务ID',
  `bucket_start` datetime(3) NULL COMMENT '分桶起始（整点对齐，1小时）',
  `total_count` bigint(20) NULL COMMENT '该桶总量',
  `error_count` bigint(20) NULL COMMENT '该桶ERROR量',
  `created_at` datetime(3) NULL COMMENT '创建时间',
  `updated_at` datetime(3) NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_log_cluster_stats_bucket_start` (`bucket_start`),
  KEY `idx_log_cluster_stats_cluster_id` (`cluster_id`),
  KEY `idx_log_cluster_stats_service_id` (`service_id`),
  UNIQUE KEY `uk_log_cluster_stat` (`cluster_id`,`bucket_start`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='';

CREATE TABLE IF NOT EXISTS `log_analysis_cursors` (
  `service_id` varchar(64) NOT NULL COMMENT '服务ID',
  `last_analyzed_at` datetime(3) NULL COMMENT '最近分析到的日志时间（本轮 end）',
  `created_at` datetime(3) NULL COMMENT '创建时间',
  `updated_at` datetime(3) NULL COMMENT '更新时间',
  PRIMARY KEY (`service_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='';

-- =============================================================
-- 九、种子数据：root 管理员 / Admin·Standard·Guest 角色 / 菜单权限 / 默认通知模板
-- =============================================================


-- 1. root 管理员（用户名 root / 密码 root）
INSERT INTO `sys_user` (`id`, `username`, `password`, `name`, `created_at`, `update_at`)
SELECT UUID(), 'root', 'root', '管理员', NOW(), NOW()
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM `sys_user` WHERE `username` = 'root');

-- 2. 内置角色（ID 与生产库一致；name 采用 CI 比对，已存在同名角色则跳过）
--    Admin=全部权限，Standard=日常运维（除用户/角色管理），Guest=仅总览大盘只读
INSERT INTO `sys_role` (`id`, `name`, `created_at`, `updated`)
SELECT t.`id`, t.`name`, NOW(), NOW()
FROM (
  SELECT '11041368-7bb8-48fa-9ba8-fcb8ea269846' AS `id`, 'Admin' AS `name` UNION ALL
  SELECT 'b21c8fd3-59d5-4d46-8cad-604360d6eb21', 'Standard' UNION ALL
  SELECT 'bedcc208-87bb-40ee-971d-2ead81bdd356', 'Guest'
) t
WHERE NOT EXISTS (SELECT 1 FROM `sys_role` r WHERE r.`id` = t.`id` OR r.`name` = t.`name`);

SET @admin_role_id    := (SELECT `id` FROM `sys_role` WHERE `name` = 'Admin' LIMIT 1);
SET @standard_role_id := (SELECT `id` FROM `sys_role` WHERE `name` = 'Standard' LIMIT 1);
SET @guest_role_id    := (SELECT `id` FROM `sys_role` WHERE `name` = 'Guest' LIMIT 1);

-- 3. 全量菜单权限（与前端路由 meta.auth 及生产库一致，共 29 项；auth_menu_* 为稳定语义 ID）
INSERT INTO `sys_auth` (`id`, `name`, `type`, `created_at`, `updated`)
SELECT t.`id`, t.`name`, 1, NOW(), NOW()
FROM (
  SELECT 'auth_menu_chat' AS `id`, '对话' AS `name` UNION ALL
  SELECT 'auth_menu_dashboard', '总览大盘' UNION ALL
  SELECT 'auth_menu_anomaly', '异常检测' UNION ALL
  SELECT 'auth_menu_alerts_console', '告警控制台' UNION ALL
  SELECT 'auth_menu_alerts_denoise', '告警降噪' UNION ALL
  SELECT '6b5ca855-1b1c-43c3-90a5-8b06a913a908', '告警事件' UNION ALL
  SELECT '7c56d5d2-8bb9-4463-af97-9dd2740528f8', '告警引擎' UNION ALL
  SELECT '880a76fa-ebad-462a-b654-c229339a1328', '告警规则' UNION ALL
  SELECT 'd16075d2-c124-41e2-99cf-73372b09940b', '夜莺告警事件' UNION ALL
  SELECT '1e4977ab-0449-4e5d-8ba2-44b3ae0f06a2', '夜莺告警规则' UNION ALL
  SELECT 'b9e1ff70-08b7-4950-a02b-ec0fcc0eafa9', '夜莺引擎配置' UNION ALL
  SELECT 'auth_menu_datasource', '数据源接入' UNION ALL
  SELECT '8dda00db-3e48-4d97-b636-0b53fa9cca37', '巡检任务' UNION ALL
  SELECT '82bf2241-2cc9-44fa-a2e3-0bcb4da72a06', '巡检报告' UNION ALL
  SELECT 'auth_menu_inspection', '巡检报告管理' UNION ALL
  SELECT 'auth_menu_rca', '根因分析' UNION ALL
  SELECT 'd29b5246-cc2c-4660-babb-1eb522be8e83', '消息模板' UNION ALL
  SELECT '71a37f43-2fd5-4ece-8fc1-624612dd7a35', '通知规则' UNION ALL
  SELECT '6b42f539-ac5b-4bc9-9d49-eb86cee3b6f3', '通知记录' UNION ALL
  SELECT 'auth_menu_notify_medium', '通知媒介' UNION ALL
  SELECT 'auth_menu_notify_policy', '通知策略' UNION ALL
  SELECT 'auth_menu_logs', '日志分析' UNION ALL
  SELECT '941764ce-551b-459a-a557-9b67e8800b7f', '服务注册' UNION ALL
  SELECT 'auth_menu_llm', 'LLM 管理' UNION ALL
  SELECT '5a28c991-5118-4feb-99e3-84b276a7567c', '业务分组' UNION ALL
  SELECT 'c41bc236-8935-4700-91c1-b8b44f39691b', '系统配置' UNION ALL
  SELECT 'auth_menu_user', '用户管理' UNION ALL
  SELECT 'auth_menu_role', '角色管理' UNION ALL
  SELECT 'auth_menu_knowledge', '运维知识库'
) t
WHERE NOT EXISTS (SELECT 1 FROM `sys_auth` a WHERE a.`id` = t.`id` OR a.`name` = t.`name`);

-- 4. root 用户绑定 Admin 角色
SET @root_user_id := (SELECT `id` FROM `sys_user` WHERE `username` = 'root' LIMIT 1);

INSERT INTO `sys_user_role_relation` (`id`, `user_id`, `role_id`, `created_at`, `updated_at`)
SELECT UUID(), @root_user_id, @admin_role_id, NOW(), NOW()
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM `sys_user_role_relation`
  WHERE `user_id` = @root_user_id AND `role_id` = @admin_role_id
);

-- 5. 角色绑定菜单权限（与生产库现行有效数据一致；同一角色+权限已存在则跳过）
-- 5.1 Admin：绑定全部菜单权限
INSERT INTO `sys_role_auth_relation`
  (`id`, `role_id`, `auth_id`, `role_name`, `auth_name`, `created_at`, `updated_at`, `updated`, `roleId`)
SELECT UUID(), @admin_role_id, a.`id`, 'Admin', a.`name`, NOW(), NOW(), NOW(), @admin_role_id
FROM `sys_auth` a
WHERE NOT EXISTS (
  SELECT 1 FROM `sys_role_auth_relation` r
  WHERE r.`role_id` = @admin_role_id AND r.`auth_id` = a.`id`
);

-- 5.2 Standard：日常运维权限（不含用户管理、角色管理，共 27 项）
INSERT INTO `sys_role_auth_relation`
  (`id`, `role_id`, `auth_id`, `role_name`, `auth_name`, `created_at`, `updated_at`, `updated`, `roleId`)
SELECT UUID(), @standard_role_id, a.`id`, 'Standard', a.`name`, NOW(), NOW(), NOW(), @standard_role_id
FROM `sys_auth` a
WHERE a.`id` IN (
  'auth_menu_chat',
  'auth_menu_dashboard',
  'auth_menu_anomaly',
  'auth_menu_alerts_console',
  'auth_menu_alerts_denoise',
  '6b5ca855-1b1c-43c3-90a5-8b06a913a908',
  '7c56d5d2-8bb9-4463-af97-9dd2740528f8',
  '880a76fa-ebad-462a-b654-c229339a1328',
  'd16075d2-c124-41e2-99cf-73372b09940b',
  '1e4977ab-0449-4e5d-8ba2-44b3ae0f06a2',
  'b9e1ff70-08b7-4950-a02b-ec0fcc0eafa9',
  'auth_menu_datasource',
  '8dda00db-3e48-4d97-b636-0b53fa9cca37',
  '82bf2241-2cc9-44fa-a2e3-0bcb4da72a06',
  'auth_menu_inspection',
  'auth_menu_rca',
  'd29b5246-cc2c-4660-babb-1eb522be8e83',
  '71a37f43-2fd5-4ece-8fc1-624612dd7a35',
  '6b42f539-ac5b-4bc9-9d49-eb86cee3b6f3',
  'auth_menu_notify_medium',
  'auth_menu_notify_policy',
  'auth_menu_logs',
  '941764ce-551b-459a-a557-9b67e8800b7f',
  'auth_menu_llm',
  '5a28c991-5118-4feb-99e3-84b276a7567c',
  'c41bc236-8935-4700-91c1-b8b44f39691b',
  'auth_menu_knowledge'
)
AND NOT EXISTS (
  SELECT 1 FROM `sys_role_auth_relation` r
  WHERE r.`role_id` = @standard_role_id AND r.`auth_id` = a.`id`
);

-- 5.3 Guest：仅绑定总览大盘（只读）
INSERT INTO `sys_role_auth_relation`
  (`id`, `role_id`, `auth_id`, `role_name`, `auth_name`, `created_at`, `updated_at`, `updated`, `roleId`)
SELECT UUID(), @guest_role_id, a.`id`, 'Guest', a.`name`, NOW(), NOW(), NOW(), @guest_role_id
FROM `sys_auth` a
WHERE a.`id` = 'auth_menu_dashboard'
  AND NOT EXISTS (
    SELECT 1 FROM `sys_role_auth_relation` r
    WHERE r.`role_id` = @guest_role_id AND r.`auth_id` = a.`id`
  );

-- 6. 默认钉钉 markdown 告警模板
INSERT INTO `notify_template`
  (`id`, `name`, `description`, `type`, `media_type`, `content`, `is_enabled`, `created_at`, `updated_at`)
SELECT
  'tpl-dingtalk-default',
  '默认钉钉告警模板',
  '钉钉 markdown 格式告警通知模板，包含触发/恢复场景，使用 $event 变量和 $.domain 站点地址',
  'all',
  'dingtalk',
  '#### {{if $event.IsRecovered}}<font color="#008800">💚{{$event.RuleName}}</font>{{else}}<font color="#FF0000">💔{{$event.RuleName}}</font>{{end}}
---
{{$time_duration := sub now $event.FirstTrigger.Unix }}{{if $event.IsRecovered}}{{$time_duration = sub $event.LastEvalTime.Unix $event.FirstTrigger.Unix }}{{end}}
- **告警级别**: {{$event.SeverityLabel}}
{{- if $event.RuleNote}}
	- **规则备注**: {{$event.RuleNote}}
{{- end}}
{{- if not $event.IsRecovered}}
- **触发时值**: {{$event.TriggerValue}}
- **触发时间**: {{timeformatCN $event.TriggerTime}}
- **告警持续时长**: {{humanizeDuration $time_duration}}
{{- else}}
- **恢复时间**: {{timeformatCN $event.LastEvalTime}}
- **告警持续时长**: {{humanizeDuration $time_duration}}
{{- end}}
- **业务组**: {{$event.BusiGroupName}}
- **告警事件标签**:
{{- range $key, $val := $event.TagsMap}}
	- {{$key}}: {{$val}}
{{- end}}
{{if $event.AnnotationsJSON}}
- **附加信息**:
{{- range $key, $val := $event.AnnotationsJSON}}
	- {{$key}}: {{$val}}
{{- end}}
{{end}}',
  1,
  NOW(),
  NOW()
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM `notify_template` WHERE `id` = 'tpl-dingtalk-default');

-- 7. 默认 Webhook / 企微 / 邮件通用模板
INSERT INTO `notify_template`
  (`id`, `name`, `description`, `type`, `media_type`, `content`, `is_enabled`, `created_at`, `updated_at`)
SELECT
  'tpl-webhook-default',
  '默认 Webhook 通用模板',
  'Webhook / 企微 / 邮件通用文本模板，使用 $event 变量语法',
  'all',
  'webhook',
  '{{if $event.IsRecovered}}🟢 {{$event.RuleName}} 已恢复{{else}}🔴 {{$event.RuleName}} 告警{{end}}
━━━━━━━━━━━━━━━━━━━━━━━━━━━
**告警级别**: {{$event.SeverityLabel}}
**业务组**: {{$event.BusiGroupName}}
{{if $event.IsRecovered}}**恢复时间**: {{timeformatCN $event.LastEvalTime}}{{else}}**触发时间**: {{timeformatCN $event.TriggerTime}}{{end}}
{{if not $event.IsRecovered}}**触发值**: {{$event.TriggerValue}}
{{end}}**标签**: {{$event.TagsJSON}}
**站点**: {{$.domain}}',
  1,
  NOW(),
  NOW()
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM `notify_template` WHERE `id` = 'tpl-webhook-default');


-- =============================================================
-- 十、验证查询
-- =============================================================

SELECT 'sys_user' AS `表名`, COUNT(*) AS `行数` FROM `sys_user` UNION ALL
SELECT 'sys_role', COUNT(*) FROM `sys_role` UNION ALL
SELECT 'sys_auth', COUNT(*) FROM `sys_auth` UNION ALL
SELECT 'sys_user_role_relation', COUNT(*) FROM `sys_user_role_relation` UNION ALL
SELECT 'sys_role_auth_relation', COUNT(*) FROM `sys_role_auth_relation` UNION ALL
SELECT 'chat_session', COUNT(*) FROM `chat_session` UNION ALL
SELECT 'chat_message', COUNT(*) FROM `chat_message` UNION ALL
SELECT 'llm_config', COUNT(*) FROM `llm_config` UNION ALL
SELECT 'datasources', COUNT(*) FROM `datasources` UNION ALL
SELECT 'alert_engine_config', COUNT(*) FROM `alert_engine_config` UNION ALL
SELECT 'n9e_config', COUNT(*) FROM `n9e_config` UNION ALL
SELECT 'alert_rules', COUNT(*) FROM `alert_rules` UNION ALL
SELECT 'alert_events', COUNT(*) FROM `alert_events` UNION ALL
SELECT 'root_cause_analyses', COUNT(*) FROM `root_cause_analyses` UNION ALL
SELECT 'inspection_tasks', COUNT(*) FROM `inspection_tasks` UNION ALL
SELECT 'inspection_reports', COUNT(*) FROM `inspection_reports` UNION ALL
SELECT 'notify_media', COUNT(*) FROM `notify_media` UNION ALL
SELECT 'notify_template', COUNT(*) FROM `notify_template` UNION ALL
SELECT 'notify_rule', COUNT(*) FROM `notify_rule` UNION ALL
SELECT 'notify_records', COUNT(*) FROM `notify_records` UNION ALL
SELECT 'kb_document', COUNT(*) FROM `kb_document` UNION ALL
SELECT 'kb_chunk', COUNT(*) FROM `kb_chunk` UNION ALL
SELECT 'services', COUNT(*) FROM `services` UNION ALL
SELECT 'business_groups', COUNT(*) FROM `business_groups` UNION ALL
SELECT 'denoise_policies', COUNT(*) FROM `denoise_policies` UNION ALL
SELECT 'denoise_records', COUNT(*) FROM `denoise_records` UNION ALL
SELECT 'system_configs', COUNT(*) FROM `system_configs` UNION ALL
SELECT 'log_templates', COUNT(*) FROM `log_templates` UNION ALL
SELECT 'log_clusters', COUNT(*) FROM `log_clusters` UNION ALL
SELECT 'log_cluster_stats', COUNT(*) FROM `log_cluster_stats` UNION ALL
SELECT 'log_analysis_cursors', COUNT(*) FROM `log_analysis_cursors`;
