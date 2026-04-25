#!/usr/bin/env bash
# Tail a Redis stream and filter by `device_uid`.
#
# 智能参数识别：
#   ./redis-iot-uid.sh <DEVICE_UID>              # 最常用：默认 iot:monitor:stream + 按 UID 过滤
#   ./redis-iot-uid.sh <STREAM>                  # 看整条 stream（无过滤）
#   ./redis-iot-uid.sh <STREAM> <DEVICE_UID>     # 显式两个参数
#
# 默认 stream: iot:monitor:stream
# 密码读自 ../.env (REDIS_PASSWORD)

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/../.env"
DEFAULT_STREAM="iot:monitor:stream"

usage() {
  cat <<'USAGE'
Usage: redis-iot-uid.sh [ARG1] [ARG2]

参数智能识别：
  - 不含冒号 → 视为 DEVICE_UID，stream 用默认
  - 含冒号   → 视为 STREAM（完整名如 iot:monitor:stream）

Options:
  -h, --help    显示此帮助

Examples:
  ./redis-iot-uid.sh 25A859B8333B                      # 最常用
  ./redis-iot-uid.sh iot:alarm:stream                  # 看整条 stream
  ./redis-iot-uid.sh iot:alarm:stream 9923003AB197     # 显式两个参数
  ./redis-iot-uid.sh iot:event:engine                  # 未来的 engine 输出 stream

Notes:
  - XREAD BLOCK 0 实时跟踪新消息（不重放历史）
  - stream 不存在时脚本会立刻报错并列出可用 stream
USAGE
}

if [ "${1:-}" = "-h" ] || [ "${1:-}" = "--help" ]; then
  usage
  exit 0
fi

# 智能参数解析
STREAM="$DEFAULT_STREAM"
UID_FILTER=""
case $# in
  0) ;;
  1)
    if [[ "$1" == *":"* ]]; then
      STREAM="$1"
    else
      UID_FILTER="$1"
    fi
    ;;
  *)
    STREAM="$1"
    UID_FILTER="$2"
    ;;
esac

# 加载 .env
[ -f "$ENV_FILE" ] && set -a && source "$ENV_FILE" && set +a

REDIS_HOST="${REDIS_HOST:-127.0.0.1}"
REDIS_PORT="${REDIS_PORT:-6379}"
REDIS_PASS="${REDIS_PASSWORD:-${REDIS_PASS:-}}"

if [ -z "$REDIS_PASS" ]; then
  echo "Error: REDIS_PASSWORD not found in ${ENV_FILE}" >&2
  exit 1
fi

# redis-cli 统一参数（含 --no-auth-warning 避免警告污染输出）
RCLI=(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASS" --no-auth-warning)

# 验证 stream 存在
stream_len=$("${RCLI[@]}" XLEN "$STREAM" 2>/dev/null || echo 0)
if [ -z "$stream_len" ] || [ "$stream_len" = "0" ]; then
  echo "Error: stream '${STREAM}' 不存在或为空" >&2
  echo "" >&2
  echo "可能的有效 stream（前 10 条）：" >&2
  "${RCLI[@]}" KEYS 'iot:*:stream' 2>/dev/null | head -10 >&2
  echo "" >&2
  echo "提示：参数 '${1:-}' 如果是 device_uid（不含冒号），会被自动用作过滤。" >&2
  exit 1
fi

echo "Following '${STREAM}' on ${REDIS_HOST}:${REDIS_PORT}"
echo "  filter device_uid = '${UID_FILTER:-(all)}'"
echo "  stream length     = ${stream_len}"
echo "Ctrl+C to stop"
echo ""

# 从 '$' 起点只看新消息（避免重放历史）
last_id='$'

while true; do
  out=$("${RCLI[@]}" XREAD BLOCK 0 STREAMS "$STREAM" "$last_id" 2>/dev/null || true)
  if [ -z "$out" ]; then
    sleep 0.1
    continue
  fi

  # 按 UID 过滤（空串时 echo 全部）
  if [ -n "$UID_FILTER" ]; then
    echo "$out" | tr '\n' ' ' | sed 's/ *"device_uid"/\n"device_uid"/g' | grep --line-buffered -F "$UID_FILTER" || true
  else
    echo "$out"
  fi

  # 更新 last_id 为本次最大 stream ID（匹配 13 位毫秒-序号格式）
  new_id=$(echo "$out" | grep -oE '[0-9]{13}-[0-9]+' | tail -n1 || true)
  if [ -n "$new_id" ]; then
    last_id="$new_id"
  fi
done
