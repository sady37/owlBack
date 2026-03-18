#!/usr/bin/env bash
# 按时间范围查询 Redis 流：iot:event:stream（含 EnterRoom/ExitRoom/number-people/activity）、iot:alarm:stream、card:status:stream
# 用法: ./redis_stream_events_by_time.sh [ts_ms_center] [ts_ms_start] [ts_ms_end]
# 默认 ts_ms_center=1773100012000，范围 start=ts-1000 end=ts+2000

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/../.env"
if [ -f "$ENV_FILE" ]; then
  set -a
  source "$ENV_FILE"
  set +a
fi

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

TS_MS="${1:-1773100012000}"
START_MS="${2:-$((TS_MS - 1000))}"
END_MS="${3:-$((TS_MS + 2000))}"

echo "=== iot:event:stream ($START_MS ~ $END_MS) ==="
RAW=$(run_redis XRANGE "iot:event:stream" "${START_MS}-0" "${END_MS}-0" 2>/dev/null)
echo "$RAW"
# 统计 EnterRoom / ExitRoom / number-people / activity
if [ -n "$RAW" ]; then
  echo ""
  echo "--- event 统计 (EnterRoom / ExitRoom / number-people / activity) ---"
  printf "EnterRoom: %s\n" "$(echo "$RAW" | grep -c "EnterRoom" 2>/dev/null || echo 0)"
  printf "ExitRoom: %s\n" "$(echo "$RAW" | grep -c "ExitRoom" 2>/dev/null || echo 0)"
  printf "number-people: %s\n" "$(echo "$RAW" | grep -c "number-people" 2>/dev/null || echo 0)"
  printf "activity: %s\n" "$(echo "$RAW" | grep -c 'activity' 2>/dev/null || echo 0)"
fi

echo ""
echo "=== iot:alarm:stream ($START_MS ~ $END_MS) ==="
run_redis XRANGE "iot:alarm:stream" "$START_MS" "$END_MS"

echo ""
echo "=== card:status:stream ($START_MS ~ $END_MS) ==="
run_redis XRANGE "card:status:stream" "$START_MS" "$END_MS"
