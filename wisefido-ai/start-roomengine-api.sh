#!/bin/bash
#
# 开发期工具：启动 roomengine-api（HTTP 回放服务）
# 不进 systemd；本地手动 ./start-roomengine-api.sh，结束 ./stop-roomengine-api.sh
#
# 用法：
#   ./start-roomengine-api.sh              # 默认端口 7788
#   PORT=8899 ./start-roomengine-api.sh    # 自定义端口
#
# 端点：
#   GET /api/health
#   GET /api/playback?uid=&start=&end=&snap_min=&format=html|json[&layout=path]

set -u

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

PORT="${PORT:-7788}"

# 继承 owlBack/.env 的 DB_*（如果存在）
ENV_FILE="$SCRIPT_DIR/../.env"
if [ -f "$ENV_FILE" ]; then
	set -a
	# shellcheck source=/dev/null
	source "$ENV_FILE"
	set +a
fi

# 默认值（与 internal/playback/db.go 的 OpenDB 兜底一致）
export DB_HOST="${DB_HOST:-127.0.0.1}"
export DB_PORT="${DB_PORT:-5432}"
export DB_USER="${DB_USER:-postgres}"
export DB_PASSWORD="${DB_PASSWORD:-postgres}"
export DB_NAME="${DB_NAME:-owlrd}"
export DB_SSLMODE="${DB_SSLMODE:-disable}"

# 端口占用检查
if command -v lsof >/dev/null 2>&1 && lsof -ti ":$PORT" >/dev/null 2>&1; then
	echo "❌ Port $PORT already in use. Run ./stop-roomengine-api.sh first."
	exit 1
fi

# 日志目录（与其它服务一致）
OWL_LOG="$(cd "$SCRIPT_DIR/../.." && pwd)/log"
mkdir -p "$OWL_LOG"
LOG_FILE="$OWL_LOG/roomengine-api.log"
PID_FILE="$OWL_LOG/roomengine-api.pid"

echo "🚀 starting roomengine-api on :$PORT"
echo "   log:  $LOG_FILE"
echo "   pid:  $PID_FILE"
echo "   db:   $DB_HOST:$DB_PORT/$DB_NAME"
echo ""

: >"$LOG_FILE"
nohup go run ./cmd/roomengine-api --listen ":$PORT" >>"$LOG_FILE" 2>&1 &
echo $! >"$PID_FILE"

# 等服务起来（最多 30s）
for i in {1..30}; do
	if curl -fs "http://127.0.0.1:$PORT/api/health" >/dev/null 2>&1; then
		echo "✅ ready: http://127.0.0.1:$PORT/api/health"
		echo ""
		echo "示例（在 alitest 上 curl 测试）:"
		echo "  curl 'http://127.0.0.1:$PORT/api/playback?uid=9D8A326309E7&start=2026-04-25T13:00:00-07:00&end=2026-04-25T17:00:00-07:00&snap_min=10&layout=../doc/layout-09E7-room101.json' -o /tmp/x.html"
		echo ""
		echo "本地浏览器访问（先 ssh -L 7788:127.0.0.1:$PORT alitest）:"
		echo "  http://localhost:7788/api/playback?uid=...&start=...&end=...&layout=../doc/layout-09E7-room101.json"
		exit 0
	fi
	sleep 1
done

echo "⚠️  startup did not become healthy in 30s; check $LOG_FILE"
exit 1
