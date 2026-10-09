#!/bin/bash
# Docker 一键部署专用：创建后端连接用的最小权限用户（root 仅保留给 DBA 场景）。
# MySQL 镜像对 .sql 文件不做变量替换，故用 .sh 脚本读取 MYSQL_PASSWORD 环境变量执行。
set -e

mysql -u root -p"${MYSQL_ROOT_PASSWORD}" <<SQL
CREATE USER IF NOT EXISTS 'aiops'@'%' IDENTIFIED BY '${MYSQL_PASSWORD}';
GRANT ALL PRIVILEGES ON \`aiops\`.* TO 'aiops'@'%';
FLUSH PRIVILEGES;
SQL
