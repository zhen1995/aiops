-- =============================================================
-- AIOPS 初始化脚本：创建用户/角色/权限相关表 + root 管理员
-- 数据库：aiops
-- 说明：
--   1. 全部为幂等写法（CREATE TABLE IF NOT EXISTS + 存在即跳过），可重复执行
--   2. 后端登录为明文密码比对（controllers/auth.go Login），故 password 直接存 'sanquan'
--   3. sys_auth 覆盖前端路由全部 17 个菜单（frontend/src/router/index.js meta.auth）
-- =============================================================

USE aiops;

-- -------------------------------------------------------------
-- 1. 建表（与 backend/models 中 GORM 模型结构一致）
-- -------------------------------------------------------------

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

-- -------------------------------------------------------------
-- 2. 插入 root 管理员（用户名 root / 密码 sanquan），已存在则跳过
-- -------------------------------------------------------------
INSERT INTO `sys_user` (`id`, `username`, `password`, `name`, `created_at`, `update_at`)
SELECT UUID(), 'root', 'sanquan', '管理员', NOW(), NOW()
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM `sys_user` WHERE `username` = 'root');

-- -------------------------------------------------------------
-- 3. 插入 admin 角色，已存在则跳过
-- -------------------------------------------------------------
INSERT INTO `sys_role` (`id`, `name`, `created_at`, `updated`)
SELECT UUID(), 'admin', NOW(), NOW()
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM `sys_role` WHERE `name` = 'admin');

SET @admin_role_id := (SELECT `id` FROM `sys_role` WHERE `name` = 'admin' LIMIT 1);

-- -------------------------------------------------------------
-- 4. 插入全部菜单权限（与前端路由 meta.auth 一一对应），已存在则跳过
-- -------------------------------------------------------------
INSERT INTO `sys_auth` (`id`, `name`, `type`, `created_at`, `updated`)
SELECT UUID(), t.`name`, 1, NOW(), NOW()
FROM (
  SELECT '对话'     AS `name` UNION ALL
  SELECT '总览大盘' UNION ALL
  SELECT '告警控制台' UNION ALL
  SELECT '告警降噪' UNION ALL
  SELECT '告警事件' UNION ALL
  SELECT '告警规则' UNION ALL
  SELECT '根因分析' UNION ALL
  SELECT '日志分析' UNION ALL
  SELECT 'LLM 管理' UNION ALL
  SELECT 'Skill 管理' UNION ALL
  SELECT '运维知识库' UNION ALL
  SELECT '巡检任务' UNION ALL
  SELECT '巡检报告' UNION ALL
  SELECT '通知媒介' UNION ALL
  SELECT '用户管理' UNION ALL
  SELECT '角色管理' UNION ALL
  SELECT '数据源接入'
) t
WHERE NOT EXISTS (SELECT 1 FROM `sys_auth` a WHERE a.`name` = t.`name`);

-- -------------------------------------------------------------
-- 5. root 用户绑定 admin 角色，已绑定则跳过
-- -------------------------------------------------------------
SET @root_user_id := (SELECT `id` FROM `sys_user` WHERE `username` = 'root' LIMIT 1);

INSERT INTO `sys_user_role_relation` (`id`, `user_id`, `role_id`, `created_at`, `updated_at`)
SELECT UUID(), @root_user_id, @admin_role_id, NOW(), NOW()
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM `sys_user_role_relation`
  WHERE `user_id` = @root_user_id AND `role_id` = @admin_role_id
);

-- -------------------------------------------------------------
-- 6. admin 角色绑定所有菜单权限，已绑定则跳过
-- -------------------------------------------------------------
INSERT INTO `sys_role_auth_relation`
  (`id`, `role_id`, `auth_id`, `role_name`, `auth_name`, `created_at`, `updated_at`, `updated`, `roleId`)
SELECT UUID(), @admin_role_id, a.`id`, 'admin', a.`name`, NOW(), NOW(), NOW(), @admin_role_id
FROM `sys_auth` a
WHERE NOT EXISTS (
  SELECT 1 FROM `sys_role_auth_relation` r
  WHERE r.`role_id` = @admin_role_id AND r.`auth_id` = a.`id`
);

-- -------------------------------------------------------------
-- 7. 验证查询
-- -------------------------------------------------------------
SELECT u.`username`, u.`name` AS `姓名`, r.`name` AS `角色`, COUNT(ra.`auth_id`) AS `权限数`
FROM `sys_user` u
JOIN `sys_user_role_relation` ur ON ur.`user_id` = u.`id`
JOIN `sys_role` r               ON r.`id` = ur.`role_id`
JOIN `sys_role_auth_relation` ra ON ra.`role_id` = r.`id`
WHERE u.`username` = 'root'
GROUP BY u.`username`, u.`name`, r.`name`;
