#!/bin/bash
#
# 停止 roomengine-api
#
set -u

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OWL_LOG="$(cd "$SCRIPT_DIR/../.." && pwd)/log"
PID_FILE="$OWL_LOG/roomengine-api.pid"
PORT="${PORT:-7788}"

stopped=0

# 1) 按 pid 文件停（如有）
if [ -f "$PID_FILE" ]; then
	PID="$(cat "$PID_FILE" 2>/dev/null || true)"
	if [ -n "$PID" ] && kill -0 "$PID" 2>/dev/null; then
		# pid 文件里是 nohup go run 的 PID（父）；要把整个进程组连子杀掉
		pkill -P "$PID" 2>/dev/null || true
		kill "$PID" 2>/dev/null || true
		sleep 1
		kill -9 "$PID" 2>/dev/null || true
		stopped=1
	fi
	rm -f "$PID_FILE"
fi

# 2) 按 cmdline / 端口兜底（pid 文件可能丢失或过期）
pkill -f "go run ./cmd/roomengine-api" 2>/dev/null && stopped=1
pkill -f "/roomengine-api --listen" 2>/dev/null && stopped=1

# 3) 端口仍占着 → 强杀占用进程
if command -v lsof >/dev/null 2>&1; then
	pids="$(lsof -ti ":$PORT" 2>/dev/null || true)"
	if [ -n "$pids" ]; then
		echo "$pids" | xargs -r kill -9 2>/dev/null || true
		stopped=1
	fi
fi

if [ "$stopped" -eq 1 ]; then
	echo "🛑 roomengine-api stopped"
else
	echo "ℹ️  no running roomengine-api found"
fi
