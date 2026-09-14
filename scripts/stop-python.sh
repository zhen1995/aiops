#!/bin/bash
# 停止 Python 算法服务（uvicorn，端口 9000）
# 优先使用 systemd 单元；未注册 systemd 时回退到按进程名终止。
set -e

SERVICE_NAME="aiops-analyzer"
PATTERN="uvicorn app.main:app"
PORT=9000
WAIT_SECONDS=10

# 1. systemd 方式（部署文档推荐的 aiops-analyzer.service）
if command -v systemctl >/dev/null 2>&1 && systemctl list-unit-files 2>/dev/null | grep -q "^${SERVICE_NAME}\.service"; then
    echo "[stop-python] 通过 systemctl 停止 ${SERVICE_NAME} ..."
    systemctl stop "$SERVICE_NAME"
    echo "[stop-python] 已停止。"
    exit 0
fi

# 2. nohup 手动启动方式：按命令行特征匹配 uvicorn 主进程与子进程（worker）
PIDS=$(pgrep -f "$PATTERN" || true)
if [ -z "$PIDS" ]; then
    echo "[stop-python] 未发现运行中的 uvicorn 进程（端口 ${PORT}），无需停止。"
    exit 0
fi

echo "[stop-python] 发现进程: $PIDS，发送 SIGTERM ..."
kill $PIDS 2>/dev/null || true

# 优雅等待退出，超时后强制 kill
for i in $(seq 1 "$WAIT_SECONDS"); do
    PIDS=$(pgrep -f "$PATTERN" || true)
    [ -z "$PIDS" ] && break
    sleep 1
done

PIDS=$(pgrep -f "$PATTERN" || true)
if [ -n "$PIDS" ]; then
    echo "[stop-python] ${WAIT_SECONDS}s 内未退出，强制 SIGKILL: $PIDS"
    kill -9 $PIDS 2>/dev/null || true
fi

# 校验端口已释放
if command -v ss >/dev/null 2>&1 && ss -lnt | grep -q ":${PORT} "; then
    echo "[stop-python] 警告：端口 ${PORT} 仍被占用！" >&2
    exit 1
fi

echo "[stop-python] Python 服务已停止。"
