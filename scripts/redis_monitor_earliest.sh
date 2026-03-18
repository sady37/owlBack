#!/usr/bin/env bash
# 检查 Redis iot:monitor:stream 中最早一条的时间
# 用法: ./redis_monitor_earliest.sh
set -e
REDIS_PASSWORD="${REDIS_PASSWORD:-TeLunSu-36kr}"
REDIS_CONTAINER="${REDIS_DOCKER_CONTAINER:-owl-redis}"

run_redis() {
  docker exec "$REDIS_CONTAINER" redis-cli -a "$REDIS_PASSWORD" -n "${REDIS_DB:-0}" "$@" 2>/dev/null
}

STREAM="iot:monitor:stream"
echo "=== $STREAM 最早一条 (XRANGE - + COUNT 1) ==="
FIRST=$(run_redis XRANGE "$STREAM" - + COUNT 1)
if [ -z "$FIRST" ]; then
  echo "流为空或不存在"
  exit 0
fi
echo "$FIRST"
# Redis stream ID 格式: ms-seq，取第一个 id 的 ms 部分
ID=$(echo "$FIRST" | head -1)
MS="${ID%-*}"
if [ -n "$MS" ] && [ "$MS" -eq "$MS" ] 2>/dev/null; then
  echo ""
  echo "最早时间戳(ms): $MS"
  if command -v date &>/dev/null; then
    SEC=$((MS/1000))
    (date -r "$SEC" 2>/dev/null || date -d "@$SEC" 2>/dev/null) && true || echo "对应时间: ${SEC}s since epoch"
  fi
fi
