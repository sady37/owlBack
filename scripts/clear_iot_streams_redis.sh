#!/usr/bin/env bash
# 清空 MQTT 写入的 Redis 流（iot:*:stream）；可选 --all 清空当前 DB 全部键
# 用法:
#   ./clear_iot_streams_redis.sh [--dry-run]   # 仅删 iot:*:stream
#   ./clear_iot_streams_redis.sh --all        # 清空当前 DB 所有键 (FLUSHDB)
#   ./clear_iot_streams_redis.sh --check      # 检查 Redis / MQTT 是否已清空（不执行删除）
# 从 owlBack/.env 读取 REDIS_ADDR REDIS_PASSWORD REDIS_DB

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/../.env"
if [ -f "$ENV_FILE" ]; then
  set -a
  source "$ENV_FILE"
  set +a
fi

DRY_RUN=false
CLEAR_ALL=false
CHECK=false
for arg in "$@"; do
  [[ "$arg" == "--dry-run" ]] && DRY_RUN=true
  [[ "$arg" == "--all" ]] && CLEAR_ALL=true
  [[ "$arg" == "--check" ]] && CHECK=true
done

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

DB_NUM="${REDIS_DB:-0}"

if $CHECK; then
  echo "=== Redis (DB $DB_NUM) ==="
  N=$(run_redis DBSIZE 2>/dev/null | grep -oE '[0-9]+' || echo "?")
  echo "  DBSIZE: $N"
  for pattern in "iot:*:stream" "card:*:stream" "card:state:*"; do
    KEYS=$(run_redis KEYS "$pattern" 2>/dev/null | sed 's/^/  /')
    if [[ -n "$KEYS" ]]; then
      echo "  $pattern:"
      echo "$KEYS"
    else
      echo "  $pattern: (none)"
    fi
  done
  echo ""
  echo "=== MQTT ==="
  if docker ps -a --format '{{.Names}}' 2>/dev/null | grep -q '^owl-mqtt$'; then
    echo "  container owl-mqtt: exists (run 'docker rm -f owl-mqtt' to remove)"
  else
    echo "  container owl-mqtt: (none)"
  fi
  MQTT_DIR="${SCRIPT_DIR}/../mqtt"
  for sub in data log; do
    if [[ -d "$MQTT_DIR/$sub" ]]; then
      COUNT=$(find "$MQTT_DIR/$sub" -type f 2>/dev/null | wc -l | tr -d ' ')
      echo "  mqtt/$sub: $COUNT file(s)"
    else
      echo "  mqtt/$sub: (dir not exist)"
    fi
  done
  echo ""
  if [[ "$N" == "0" || "$N" == "?" ]] && ! docker ps -a --format '{{.Names}}' 2>/dev/null | grep -q '^owl-mqtt$'; then
    echo "结论: Redis 当前 DB 无键且无 MQTT 容器 → 可视为已清空。"
  else
    echo "结论: 仍有数据或容器，未完全清空。"
    echo "  Redis: 先停掉 cardagg/data 等再执行 ./clear_iot_streams_redis.sh --all ，否则会立刻被写回。"
    echo "  MQTT: docker rm -f owl-mqtt；清理 mqtt/data mqtt/log。"
  fi
  exit 0
fi

if $CLEAR_ALL; then
  if $DRY_RUN; then
    echo "[dry-run] would FLUSHDB (clear all keys in DB $DB_NUM)"
  else
    echo "WARNING: About to FLUSHDB (clear ALL keys in DB $DB_NUM). Ctrl+C to cancel..."
    sleep 3
    run_redis FLUSHDB 2>/dev/null && echo "FLUSHDB (DB $DB_NUM) done." || echo "FLUSHDB failed."
  fi
  if $DRY_RUN; then
    echo "Run without --dry-run to actually flush."
  fi
  exit 0
fi

STREAMS=(
  "iot:monitor:stream"
  "iot:stat:stream"
  "iot:event:stream"
  "iot:alarm:stream"
  "iot:auth:stream"
  "iot:other:stream"
  "iot:card:stream"
)

for key in "${STREAMS[@]}"; do
  if $DRY_RUN; then
    n=$(run_redis XLEN "$key" 2>/dev/null || echo "0")
    echo "[dry-run] would DEL $key (current len=$n)"
  else
    run_redis DEL "$key" 2>/dev/null && echo "DEL $key" || echo "DEL $key (failed or missing)"
  fi
done

# card:status:stream 与 card:state:*（业务状态），一并清理
if $DRY_RUN; then
  n=$(run_redis XLEN "card:status:stream" 2>/dev/null || echo "0")
  echo "[dry-run] would DEL card:status:stream (current len=$n)"
  cnt=$(run_redis KEYS "card:state:*" 2>/dev/null | wc -l | tr -d ' ')
  echo "[dry-run] would DEL card:state:* ($cnt key(s))"
else
  run_redis DEL "card:status:stream" 2>/dev/null && echo "DEL card:status:stream" || true
  run_redis KEYS "card:state:*" 2>/dev/null | while read -r k; do
    [[ -n "$k" ]] && run_redis DEL "$k" 2>/dev/null && echo "DEL $k"
  done
fi

if $DRY_RUN; then
  echo "Run without --dry-run to actually delete."
fi
