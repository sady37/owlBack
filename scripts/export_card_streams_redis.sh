#!/usr/bin/env bash
# 从 Redis Stream 中查看与某 card_id 相关的最近条目（供对照 SSE：card:status + card:realtime）
# 用法: ./export_card_streams_redis.sh <card_id> [count]
# count 默认 120（每条流各取最近 count 条再 grep，大流量时请调大）
# 依赖 owlBack/.env 中 REDIS_ADDR / REDIS_PASSWORD / REDIS_DB

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/../.env"
if [ -f "$ENV_FILE" ]; then
  set -a
  source "$ENV_FILE"
  set +a
fi

CARD_ID="${1:?usage: $0 <card_id> [count]}"
COUNT="${2:-120}"

R_HOST="${REDIS_ADDR%:*}"
R_PORT="${REDIS_ADDR#*:}"
[ -z "$R_PORT" ] || [ "$R_PORT" = "$REDIS_ADDR" ] && R_PORT="6379"
export REDISCLI_AUTH="${REDIS_PASSWORD:-}"
REDIS_DOCKER_CONTAINER="${REDIS_DOCKER_CONTAINER:-owl-redis}"

run_redis() {
  if [ -n "$REDIS_CMD" ]; then
    eval "$REDIS_CMD" "$@"
  elif command -v redis-cli &>/dev/null; then
    redis-cli -h "$R_HOST" -p "$R_PORT" -n "${REDIS_DB:-0}" ${REDIS_PASSWORD:+-a "$REDIS_PASSWORD"} "$@"
  else
    docker exec "$REDIS_DOCKER_CONTAINER" redis-cli -a "$REDIS_PASSWORD" -n "${REDIS_DB:-0}" "$@"
  fi
}

echo "card_id=$CARD_ID"
echo "per-stream COUNT=$COUNT (then grep card_id)"
echo ""

echo "=== card:status:stream (owl-common Writer → SSE status 源) ==="
RAW_S=$(run_redis XREVRANGE card:status:stream + - COUNT "$COUNT" 2>/dev/null)
if echo "$RAW_S" | grep -qF "$CARD_ID"; then
  echo "$RAW_S" | grep -F "$CARD_ID"
else
  echo "(no matching lines in last $COUNT entries; try larger count: $0 $CARD_ID 500)"
fi
echo ""

echo "=== card:realtime:stream (PublishMonitor → SSE realtime / 轨迹) ==="
RAW_R=$(run_redis XREVRANGE card:realtime:stream + - COUNT "$COUNT" 2>/dev/null)
if echo "$RAW_R" | grep -qF "$CARD_ID"; then
  echo "$RAW_R" | grep -F "$CARD_ID"
else
  echo "(no matching lines in last $COUNT entries; try larger count)"
fi
echo ""

echo "=== vital-focus:card:{id}:realtime (wisefido-data 聚合 KV，若存在) ==="
run_redis GET "vital-focus:card:${CARD_ID}:realtime" 2>/dev/null | head -c 8000
echo ""
echo ""

echo "=== iot:card:stream (若启用；部分环境可能为空) ==="
run_redis XREVRANGE iot:card:stream + - COUNT 30 2>/dev/null | grep -F "$CARD_ID" || echo "(none or stream missing)"
