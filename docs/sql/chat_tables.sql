-- AIOPS 智能运维助手对话模块

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
