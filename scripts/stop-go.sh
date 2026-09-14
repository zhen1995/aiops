#!/bin/bash
# 停止 Go 后端服务（center 二进制，端口 8080）
# 优先使用 systemd 单元；未注册 systemd 时回退到按进程名终止。
set -e

SERVICE_NAME="aiops-center"
BIN_NAME="center"
PORT=8080
WAIT_SECONDS=10

# 1. systemd 方式（部署文档推荐的 aiops-center.service）
if command -v systemctl >/dev/null 2>&1 && systemctl list-unit-files 2>/dev/null | grep -q "^${SERVICE_NAME}\.service"; then
    echo "[stop-go] 通过 systemctl 停止 ${SERVICE_NAME} ..."
    systemctl stop "$SERVICE_NAME"
    echo "[stop-go] 已停止。"
    exit 0
fi

# 2. nohup 手动启动方式：按进程名精确匹配 center（避免误杀）
PIDS=$(pgrep -x "$BIN_NAME" || true)
if [ -z "$PIDS" ]; then
    echo "[stop-go] 未发现运行中的 ${BIN_NAME} 进程（端口 ${PORT}），无需停止。"
    exit 0
fi

echo "[stop-go] 发现进程: $PIDS，发送 SIGTERM ..."
kill $PIDS 2>/dev/null || true

# 优雅等待退出，超时后强制 kill
for i in $(seq 1 "$WAIT_SECONDS"); do
    PIDS=$(pgrep -x "$BIN_NAME" || true)
    [ -z "$PIDS" ] && break
    sleep 1
done

PIDS=$(pgrep -x "$BIN_NAME" || true)
if [ -n "$PIDS" ]; then
    echo "[stop-go] ${WAIT_SECONDS}s 内未退出，强制 SIGKILL: $PIDS"
    kill -9 $PIDS 2>/dev/null || true
fi

# 校验端口已释放
if command -v ss >/dev/null 2>&1 && ss -lnt | grep -q ":${PORT} "; then
    echo "[stop-go] 警告：端口 ${PORT} 仍被占用！" >&2
    exit 1
fi

echo "[stop-go] Go 后端服务已停止。"
