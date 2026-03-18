#!/usr/bin/env bash
# 导出指定 card_id 的 card:state Hash 各字段
# 用法: ./export_card_status_redis.sh [card_id]
# 原始值（不经 jq）: RAW=1 ./export_card_status_redis.sh [card_id]
# 从 owlBack/.env 读取 REDIS_ADDR REDIS_PASSWORD REDIS_DB

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/../.env"
if [ -f "$ENV_FILE" ]; then
  set -a
  source "$ENV_FILE"
  set +a
fi

CARD_ID="${1:-42077c6d-ed05-46ec-a76d-b45ddb48b24f}"
KEY="card:state:${CARD_ID}"

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

_redis_ping() {
  if [ -n "$REDIS_CMD" ]; then
    eval "$REDIS_CMD" PING &>/dev/null
  elif command -v redis-cli &>/dev/null; then
    redis-cli -h "$R_HOST" -p "$R_PORT" -n "${REDIS_DB:-0}" ${REDIS_PASSWORD:+-a "$REDIS_PASSWORD"} PING &>/dev/null
  else
    docker exec "$REDIS_DOCKER_CONTAINER" redis-cli -a "$REDIS_PASSWORD" -n "${REDIS_DB:-0}" PING &>/dev/null
  fi
}

if ! _redis_ping; then
  echo "Redis unreachable. Start: cd owlBack && docker compose up -d redis"
  echo "Or set REDIS_DOCKER_CONTAINER=your-redis-container (default: owl-redis)"
  exit 1
fi

echo "card_id=$CARD_ID"
echo "redis_key=$KEY"
echo "---"

for f in target room_state bathroom_state bed_state device_status alarm_state message; do
  val=$(run_redis HGET "$KEY" "$f" 2>/dev/null)
  if [ -n "$val" ]; then
    echo "[$f]"
    if [ -n "$RAW" ]; then
      echo "$val"
    else
      echo "$val" | jq '.' 2>/dev/null || echo "$val"
    fi
    echo "---"
  fi
done
