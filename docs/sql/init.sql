-- =============================================================
-- AIOPS 智能运维平台 · 全量初始化脚本
-- 数据库：MySQL 8.x  (aiops)
-- 说明：
--   1. 覆盖后端 GORM AutoMigrate 中的全部 20 张表（与 backend/models/ 一一对应）
--   2. 全部幂等写法（CREATE TABLE IF NOT EXISTS + INSERT ... WHERE NOT EXISTS），可重复执行
--   3. 内置 root 管理员、admin 角色、全量菜单权限、默认钉钉/Webhook 通知模板
--   4. 后端登录为明文密码比对（controllers/auth.go Login），root 密码固定 sanquan
-- =============================================================

CREATE DATABASE IF NOT EXISTS `aiops` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `aiops`;


-- =============================================================
-- 一、用户 / 角色 / 权限（系统基础表）
-- =============================================================

CREATE TABLE IF NOT EXISTS `sys_user` (
  `id`         varchar(40)  NOT NULL COMMENT 'id',
  `username`   varchar(30)  NOT NULL COMMENT '用户名',
  `password`   varchar(30)  NOT NULL COMMENT '密码',
  `name`       varchar(10)  NOT NULL COMMENT '姓名',
  `created_at` datetime     NULL     COMMENT '创建时间',
  `deleted_at` datetime     NULL     COMMENT '删除时间',
  `update_at`  datetime     NULL     COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

CREATE TABLE IF NOT EXISTS `sys_role` (
  `id`         varchar(40) NOT NULL COMMENT 'id',
  `name`       varchar(30) NULL     COMMENT '名称',
  `created_at` datetime    NULL     COMMENT '创建时间',
  `updated`    datetime    NULL     COMMENT '更新时间',
  `deleted_at` datetime    NULL     COMMENT '删除时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色表';

CREATE TABLE IF NOT EXISTS `sys_auth` (
  `id`         varchar(40) NOT NULL COMMENT 'id',
  `name`       varchar(50) NULL     COMMENT '名称',
  `type`       bigint      NOT NULL COMMENT '1-菜单 2-操作',
  `created_at` datetime    NULL     COMMENT '创建时间',
  `updated`    datetime    NULL     COMMENT '更新时间',
  `deleted_at` datetime    NULL     COMMENT '删除时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='权限表（菜单）';

CREATE TABLE IF NOT EXISTS `sys_user_role_relation` (
  `id`         varchar(40) NOT NULL COMMENT 'id',
  `user_id`    varchar(40) NOT NULL COMMENT '用户id',
  `role_id`    varchar(40) NOT NULL COMMENT '角色id',
  `created_at` datetime    NULL     COMMENT '创建时间',
  `updated_at` datetime    NULL     COMMENT '更新时间',
  `deleted_at` datetime    NULL     COMMENT '删除时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户-角色关联表';

CREATE TABLE IF NOT EXISTS `sys_role_auth_relation` (
  `id`         varchar(40) NOT NULL COMMENT 'id',
  `role_id`    varchar(40) NOT NULL COMMENT '角色id',
  `auth_id`    varchar(40) NOT NULL COMMENT '权限id',
  `role_name`  varchar(40) NULL     COMMENT '角色名称',
  `auth_name`  varchar(40) NULL     COMMENT '权限名称',
  `created_at` datetime    NULL     COMMENT '创建时间',
  `deleted_at` datetime    NULL     COMMENT '删除时间',
  `updated_at` datetime    NULL     COMMENT '更新时间',
  `updated`    datetime    NULL     COMMENT '更新时间',
  `roleId`     varchar(40) NOT NULL COMMENT '角色id(冗余列)',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色-权限关联表';


-- =============================================================
-- 二、对话模块（合并 chat_tables.sql）
-- =============================================================

CREATE TABLE IF NOT EXISTS `chat_session` (
  `id`              VARCHAR(36)  NOT NULL COMMENT '会话ID',
  `user_id`         VARCHAR(36)  NOT NULL COMMENT '用户ID',
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
  `id`                VARCHAR(36) NOT NULL COMMENT '消息ID',
  `session_id`        VARCHAR(36) NOT NULL COMMENT '会话ID',
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


-- =============================================================
-- 三、LLM / 数据源 / 告警引擎配置
-- =============================================================

CREATE TABLE IF NOT EXISTS `llm_config` (
  `id`                 varchar(40)  NOT NULL COMMENT 'id',
  `name`               varchar(30)  NULL     COMMENT 'LLM 名称',
  `description`        varchar(100) NULL     COMMENT '描述',
  `supplier_category`  varchar(10)  NULL     COMMENT '供应商类型 openai/claude/gemini等',
  `model_type`         varchar(10)  NOT NULL DEFAULT 'chat' COMMENT '模型类型 chat-对话 embedding-向量化',
  `model`              varchar(20)  NULL     COMMENT '模型名称',
  `base_url`           varchar(50)  NULL     COMMENT 'api url',
  `api_key`            varchar(100) NULL     COMMENT '密钥',
  `is_default`         int          NOT NULL DEFAULT 0 COMMENT '0 否，1是',
  `is_enabled`         int                   DEFAULT 1 COMMENT '是否启用 0-否 1-是',
  `created_at`         datetime     NULL     COMMENT '创建时间',
  `updated_at`         datetime     NULL     COMMENT '更新时间',
  `deleted_at`         datetime     NULL     COMMENT '删除时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='大模型配置';

CREATE TABLE IF NOT EXISTS `datasources` (
  `id`          varchar(40)  NOT NULL COMMENT 'id',
  `name`        varchar(30)  NULL     COMMENT '名称',
  `type`        varchar(20)  NULL     COMMENT '类型 Prometheus/ElasticSearch/Pyroscope',
  `url`         varchar(100) NOT NULL COMMENT 'http地址',
  `timeout`     int          NULL     COMMENT '超时(ms)',
  `username`    varchar(100) NULL     COMMENT '用户名',
  `password`    varchar(100) NULL     COMMENT '密码',
  `is_skip_ssl` int          NULL     COMMENT '是否跳过SSL验证 0-否 1-是',
  `remark`      varchar(100) NULL     COMMENT '备注',
  `is_enabled`  int          NULL     COMMENT '状态 0-不启用 1-启用',
  `created_at`  datetime     NULL     COMMENT '创建时间',
  `updated_at`  datetime     NULL     COMMENT '更新时间',
  `deleted_at`  datetime     NULL     COMMENT '删除时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据源';

CREATE TABLE IF NOT EXISTS `alert_engine_config` (
  `id`         varchar(40)  NOT NULL COMMENT 'id',
  `name`       varchar(30)  NOT NULL COMMENT '名称',
  `base_url`   varchar(100) NOT NULL COMMENT '夜莺地址',
  `token`      varchar(100) NOT NULL COMMENT '用户Token',
  `gids`       varchar(100) NULL     COMMENT '业务组ID,逗号分隔',
  `is_enabled` int                   DEFAULT 1 COMMENT '0-禁用 1-启用',
  `is_default` int                   DEFAULT 0 COMMENT '0-否 1-是',
  `remark`     varchar(100) NULL     COMMENT '备注',
  `created_at` datetime     NULL     COMMENT '创建时间',
  `updated_at` datetime     NULL     COMMENT '更新时间',
  `deleted_at` datetime     NULL     COMMENT '删除时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='告警引擎配置';

CREATE TABLE IF NOT EXISTS `n9e_config` (
  `id`         varchar(40)  NOT NULL COMMENT 'id',
  `name`       varchar(64)  NOT NULL DEFAULT '夜莺' COMMENT '名称',
  `address`    varchar(255) NOT NULL COMMENT '夜莺服务地址',
  `token`      varchar(128) NOT NULL COMMENT 'X-User-Token',
  `is_enabled` int                   DEFAULT 1 COMMENT '是否启用',
  `created_at` datetime     NULL     COMMENT '创建时间',
  `updated_at` datetime     NULL     COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='夜莺引擎配置';


-- =============================================================
-- 四、告警规则 / 告警事件 / 根因分析
-- =============================================================

CREATE TABLE IF NOT EXISTS `alert_rules` (
  `id`                      varchar(64)  NOT NULL COMMENT 'id',
  `name`                    varchar(128) NOT NULL COMMENT '规则名称',
  `prom_ql`                 text         NOT NULL COMMENT 'PromQL 表达式',
  `eval_interval`           int          NULL     COMMENT '执行频率(秒)',
  `duration`                int          NULL     COMMENT '持续时间(秒)',
  `severity`                int          NULL     COMMENT '告警级别 1-P1紧急 2-P2警告 3-P3提醒',
  `notify_rule_id`          varchar(64)  NULL     COMMENT '通知规则ID',
  `repeat_interval_minutes` int          NOT NULL DEFAULT 0 COMMENT '重复通知间隔(分钟) 0=不重复',
  `max_send_count`          int          NOT NULL DEFAULT 0 COMMENT '最大发送次数 0=不限制',
  `is_enabled`              int          NULL     COMMENT '启用状态 0-停用 1-启用',
  `created_at`              datetime     NULL     COMMENT '创建时间',
  `updated_at`              datetime     NULL     COMMENT '更新时间',
  `deleted_at`              datetime     NULL     COMMENT '删除时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='告警规则';

CREATE TABLE IF NOT EXISTS `alert_events` (
  `id`               varchar(64) NOT NULL COMMENT 'id',
  `rule_id`          varchar(64) NULL     COMMENT '关联告警规则ID',
  `rule_name`        varchar(128) NULL    COMMENT '规则名称（冗余）',
  `severity`         int         NULL     COMMENT '告警级别',
  `type`             varchar(16) NULL      COMMENT '事件类型 alert/recovery',
  `status`           varchar(16) NULL      COMMENT '状态 firing/resolved',
  `target_ident`     varchar(256) NULL    COMMENT '告警对象',
  `tags`             text        NULL     COMMENT '标签序列化',
  `trigger_value`    varchar(64) NULL      COMMENT '触发值',
  `trigger_time`     datetime    NULL     COMMENT '触发时间',
  `recovered_at`     datetime    NULL     COMMENT '恢复时间',
  `notify_count`     int         NOT NULL DEFAULT 0 COMMENT '已发送通知次数',
  `last_notified_at` datetime    NULL     COMMENT '上次发送通知时间',
  `created_at`       datetime    NULL     COMMENT '创建时间',
  `updated_at`       datetime    NULL     COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_rule_id` (`rule_id`),
  KEY `idx_type` (`type`),
  KEY `idx_status` (`status`),
  KEY `idx_trigger_time` (`trigger_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='告警事件';

CREATE TABLE IF NOT EXISTS `root_cause_analyses` (
  `id`            varchar(64) NOT NULL COMMENT 'id',
  `alert_event_id` varchar(64) NULL    COMMENT '关联告警事件ID',
  `rule_name`      varchar(128) NULL    COMMENT '规则名称',
  `target_ident`  varchar(256) NULL    COMMENT '告警对象',
  `tags`          text         NULL    COMMENT '标签序列化',
  `severity`      int          NULL    COMMENT '告警级别',
  `trigger_time`  datetime     NULL    COMMENT '告警触发时间',
  `status`        varchar(16)  NULL    COMMENT '状态 running/completed/failed',
  `error`         text         NULL    COMMENT '失败原因',
  `result`        text         NULL    COMMENT '结构化分析结果 JSON',
  `content`       text         NULL    COMMENT '完整分析报告 Markdown',
  `created_at`    datetime     NULL    COMMENT '创建时间',
  `updated_at`    datetime     NULL    COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_alert_event_id` (`alert_event_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='根因分析';


-- =============================================================
-- 五、巡检任务 / 巡检报告
-- =============================================================

CREATE TABLE IF NOT EXISTS `inspection_tasks` (
  `id`               uint         NOT NULL AUTO_INCREMENT COMMENT '任务ID',
  `name`             varchar(128) NOT NULL COMMENT '任务名称',
  `cron_expr`        varchar(64)  NOT NULL COMMENT 'cron 表达式',
  `prompt`           text         NOT NULL COMMENT '任务提示词',
  `notify_media_ids` text         NULL     COMMENT '通知媒介ID列表JSON',
  `enabled`          tinyint(1)   NOT NULL DEFAULT 1 COMMENT '是否启用',
  `last_run_at`      datetime     NULL     COMMENT '上次执行时间',
  `next_run_at`      datetime     NULL     COMMENT '预计下次执行时间',
  `created_at`       datetime     NULL     COMMENT '创建时间',
  `updated_at`       datetime     NULL     COMMENT '更新时间',
  `deleted_at`       datetime     NULL     COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='巡检任务';

CREATE TABLE IF NOT EXISTS `inspection_reports` (
  `id`        uint         NOT NULL AUTO_INCREMENT COMMENT '报告ID',
  `task_id`   uint         NOT NULL COMMENT '关联任务 ID',
  `task_name` varchar(128) NULL     COMMENT '冗余任务名',
  `title`     varchar(256) NOT NULL COMMENT '报告标题',
  `content`   longtext     NOT NULL COMMENT 'Markdown 报告正文',
  `summary`   varchar(512) NULL     COMMENT '简短摘要',
  `score`     int          NULL     COMMENT '综合评分 0-100',
  `status`    varchar(32)  NULL     DEFAULT 'completed' COMMENT 'generating/completed/failed',
  `error`     varchar(512) NULL     COMMENT '失败原因',
  `created_at` datetime    NULL     COMMENT '创建时间',
  `deleted_at` datetime    NULL     COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_task_id` (`task_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='巡检报告';


-- =============================================================
-- 六、通知媒介 / 模板 / 规则
-- =============================================================

CREATE TABLE IF NOT EXISTS `notify_media` (
  `id`         varchar(64)  NOT NULL COMMENT 'id',
  `name`       varchar(60)  NOT NULL COMMENT '媒介名称',
  `type`       varchar(20)  NOT NULL COMMENT '类型 webhook/dingtalk',
  `config`     text         NULL     COMMENT '媒介配置JSON',
  `remark`     varchar(200) NULL     COMMENT '备注',
  `is_enabled` int          NULL     COMMENT '状态 0-停用 1-启用',
  `created_at` datetime     NULL     COMMENT '创建时间',
  `updated_at` datetime     NULL     COMMENT '更新时间',
  `deleted_at` datetime     NULL     COMMENT '删除时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='通知媒介';

CREATE TABLE IF NOT EXISTS `notify_template` (
  `id`            varchar(64)  NOT NULL COMMENT 'id',
  `name`          varchar(64)  NOT NULL COMMENT '模板名称',
  `description`   varchar(255) NULL     COMMENT '描述',
  `type`          varchar(16)  NOT NULL DEFAULT 'all' COMMENT '适用场景 firing/recovered/all',
  `media_type`    varchar(32)  NOT NULL DEFAULT 'dingtalk' COMMENT '媒介类型 dingtalk/webhook/email/wecom',
  `content`       text         NOT NULL COMMENT '模板内容（Go template 语法）',
  `variables_hint` text        NULL    COMMENT '变量提示 JSON',
  `is_enabled`    int          NOT NULL DEFAULT 1 COMMENT '是否启用',
  `created_at`    datetime     NULL     COMMENT '创建时间',
  `updated_at`    datetime     NULL     COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='通知消息模板';

CREATE TABLE IF NOT EXISTS `notify_rule` (
  `id`              varchar(64)  NOT NULL COMMENT 'id',
  `name`            varchar(64)  NOT NULL COMMENT '规则名称',
  `remark`          varchar(255) NULL     COMMENT '备注',
  `media_id`        varchar(64)  NULL     COMMENT '关联通知媒介ID',
  `template_id`     varchar(64)  NULL     COMMENT '关联消息模板ID',
  `trigger_types`   varchar(16)  NULL     DEFAULT 'all' COMMENT '触发类型 firing/recovered/all',
  `severity_filter` varchar(128) NULL     COMMENT '级别过滤 JSON 数组字符串',
  `is_enabled`      int          NOT NULL DEFAULT 1 COMMENT '是否启用',
  `created_at`      datetime     NULL     COMMENT '创建时间',
  `updated_at`      datetime     NULL     COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='通知规则';


-- =============================================================
-- 七、知识库（元数据表，向量存 Qdrant）
-- =============================================================

CREATE TABLE IF NOT EXISTS `kb_document` (
  `id`         varchar(40)  NOT NULL COMMENT 'id',
  `name`       varchar(200) NOT NULL COMMENT '文件名',
  `type`       varchar(10)  NOT NULL COMMENT '扩展名',
  `size`       bigint       NULL     COMMENT '文件大小(字节)',
  `status`     varchar(10)  NOT NULL DEFAULT 'pending' COMMENT '状态 pending/indexing/indexed/failed',
  `error_msg`  varchar(500) NULL     COMMENT '失败原因',
  `uploader`   varchar(30)  NULL     COMMENT '上传人',
  `file_path`  varchar(300) NULL     COMMENT '文件存储路径',
  `created_at` datetime     NULL     COMMENT '创建时间',
  `updated_at` datetime     NULL     COMMENT '更新时间',
  `deleted_at` datetime     NULL     COMMENT '删除时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='知识库文档';

CREATE TABLE IF NOT EXISTS `kb_chunk` (
  `id`          varchar(40) NOT NULL COMMENT 'id',
  `document_id` varchar(40) NOT NULL COMMENT '文档id',
  `chunk_index` int         NULL     COMMENT '块序号',
  `content`     text        NULL     COMMENT '块文本',
  `created_at`  datetime    NULL     COMMENT '创建时间',
  `deleted_at`  datetime    NULL     COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_document_id` (`document_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='知识库分块';


-- =============================================================
-- 八、种子数据：root 管理员 / admin 角色 / 菜单权限 / 默认通知模板
-- =============================================================

-- 1. root 管理员（用户名 root / 密码 sanquan）
INSERT INTO `sys_user` (`id`, `username`, `password`, `name`, `created_at`, `update_at`)
SELECT UUID(), 'root', 'sanquan', '管理员', NOW(), NOW()
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM `sys_user` WHERE `username` = 'root');

-- 2. admin 角色
INSERT INTO `sys_role` (`id`, `name`, `created_at`, `updated`)
SELECT UUID(), 'admin', NOW(), NOW()
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM `sys_role` WHERE `name` = 'admin');

SET @admin_role_id := (SELECT `id` FROM `sys_role` WHERE `name` = 'admin' LIMIT 1);

-- 3. 全量菜单权限（与前端路由 meta.auth 一一对应）
INSERT INTO `sys_auth` (`id`, `name`, `type`, `created_at`, `updated`)
SELECT UUID(), t.`name`, 1, NOW(), NOW()
FROM (
  SELECT '对话'         AS `name` UNION ALL
  SELECT '总览大盘'     UNION ALL
  SELECT '告警控制台'   UNION ALL
  SELECT '告警降噪'     UNION ALL
  SELECT '告警事件'     UNION ALL
  SELECT '告警规则'     UNION ALL
  SELECT '根因分析'     UNION ALL
  SELECT '日志分析'     UNION ALL
  SELECT 'LLM 管理'     UNION ALL
  SELECT 'Skill 管理'   UNION ALL
  SELECT '运维知识库'   UNION ALL
  SELECT '巡检任务'     UNION ALL
  SELECT '巡检报告'     UNION ALL
  SELECT '通知媒介'     UNION ALL
  SELECT '消息模板'     UNION ALL
  SELECT '通知规则'     UNION ALL
  SELECT '夜莺引擎配置' UNION ALL
  SELECT '夜莺告警规则' UNION ALL
  SELECT '夜莺告警事件' UNION ALL
  SELECT '用户管理'     UNION ALL
  SELECT '角色管理'     UNION ALL
  SELECT '数据源接入'   UNION ALL
  SELECT '服务注册'
) t
WHERE NOT EXISTS (SELECT 1 FROM `sys_auth` a WHERE a.`name` = t.`name`);

-- 4. root 用户绑定 admin 角色
SET @root_user_id := (SELECT `id` FROM `sys_user` WHERE `username` = 'root' LIMIT 1);

INSERT INTO `sys_user_role_relation` (`id`, `user_id`, `role_id`, `created_at`, `updated_at`)
SELECT UUID(), @root_user_id, @admin_role_id, NOW(), NOW()
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM `sys_user_role_relation`
  WHERE `user_id` = @root_user_id AND `role_id` = @admin_role_id
);

-- 5. admin 角色绑定所有菜单权限
INSERT INTO `sys_role_auth_relation`
  (`id`, `role_id`, `auth_id`, `role_name`, `auth_name`, `created_at`, `updated_at`, `updated`, `roleId`)
SELECT UUID(), @admin_role_id, a.`id`, 'admin', a.`name`, NOW(), NOW(), NOW(), @admin_role_id
FROM `sys_auth` a
WHERE NOT EXISTS (
  SELECT 1 FROM `sys_role_auth_relation` r
  WHERE r.`role_id` = @admin_role_id AND r.`auth_id` = a.`id`
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
-- 九、验证查询
-- =============================================================

SELECT 'sys_user' AS `表名`, COUNT(*) AS `行数` FROM `sys_user`
UNION ALL SELECT 'sys_role', COUNT(*) FROM `sys_role`
UNION ALL SELECT 'sys_auth', COUNT(*) FROM `sys_auth`
UNION ALL SELECT 'sys_user_role_relation', COUNT(*) FROM `sys_user_role_relation`
UNION ALL SELECT 'sys_role_auth_relation', COUNT(*) FROM `sys_role_auth_relation`
UNION ALL SELECT 'chat_session', COUNT(*) FROM `chat_session`
UNION ALL SELECT 'chat_message', COUNT(*) FROM `chat_message`
UNION ALL SELECT 'llm_config', COUNT(*) FROM `llm_config`
UNION ALL SELECT 'datasources', COUNT(*) FROM `datasources`
UNION ALL SELECT 'alert_engine_config', COUNT(*) FROM `alert_engine_config`
UNION ALL SELECT 'n9e_config', COUNT(*) FROM `n9e_config`
UNION ALL SELECT 'alert_rules', COUNT(*) FROM `alert_rules`
UNION ALL SELECT 'alert_events', COUNT(*) FROM `alert_events`
UNION ALL SELECT 'root_cause_analyses', COUNT(*) FROM `root_cause_analyses`
UNION ALL SELECT 'inspection_tasks', COUNT(*) FROM `inspection_tasks`
UNION ALL SELECT 'inspection_reports', COUNT(*) FROM `inspection_reports`
UNION ALL SELECT 'notify_media', COUNT(*) FROM `notify_media`
UNION ALL SELECT 'notify_template', COUNT(*) FROM `notify_template`
UNION ALL SELECT 'notify_rule', COUNT(*) FROM `notify_rule`
UNION ALL SELECT 'kb_document', COUNT(*) FROM `kb_document`
UNION ALL SELECT 'kb_chunk', COUNT(*) FROM `kb_chunk`;
