#!/usr/bin/env bash
# 一次性清理 wisefido-sensor roomengine consumer group 的历史 stale pending。
#
# 背景：旧版 engine 主循环 XReadGroup 后从不 XAck，pending 列表无限增长
# （4.86 天积压 1.39M+ 条）。新版 engine（commit 之后）已修复 XAck，但历史
# 残留需要单独清理一次。
#
# 策略：XGROUP DELCONSUMER —— redis 删 consumer 时连带删它的 PEL 记录，
# 比 XPENDING+XACK 循环（1400+ 批次）快几个数量级。重启 ai 时 redis 自动
# 重建 consumer 条目，新流量不受影响。
#
# 用法：
#   ./clean-roomengine-pending.sh           # 清理所有 3 个流（monitor/event/alarm）
#   REDIS_PASSWORD=xxx ./clean-roomengine-pending.sh
#
# 推荐流程：
#   sudo systemctl stop owlback.sensor
#   ./clean-roomengine-pending.sh
#   sudo systemctl start owlback.sensor
#
# 在线运行也能跑（已修复 XAck 的 engine 不会因 consumer 短暂消失出问题），
# 但停服跑更干净（无需考虑微秒级的 in-flight 消息 race）。

set -u
PASS="${REDIS_PASSWORD:-TeLunSu-36kr}"
HOST="${REDIS_HOST:-127.0.0.1}"
PORT="${REDIS_PORT:-6379}"
GROUP="roomengine"

declare -A CONSUMERS=(
  [iot:monitor:stream]=roomengine-monitor-1
  [iot:event:stream]=roomengine-event-1
  [iot:alarm:stream]=roomengine-alarm-1
)

redis() { redis-cli -h "$HOST" -p "$PORT" -a "$PASS" --no-auth-warning "$@"; }

echo "=== Before ==="
for stream in "${!CONSUMERS[@]}"; do
  pending=$(redis XLEN "$stream" 2>/dev/null)
  group_pending=$(redis XINFO GROUPS "$stream" 2>/dev/null | awk -v g="$GROUP" '
    /^name$/{getline n}
    n==g && /^pending$/{getline p; print p; exit}
  ')
  printf "  %-22s stream_len=%s   roomengine_pending=%s\n" "$stream" "$pending" "${group_pending:-0}"
done

total_deleted=0
echo
echo "=== Deleting consumers ==="
for stream in "${!CONSUMERS[@]}"; do
  consumer=${CONSUMERS[$stream]}
  result=$(redis XGROUP DELCONSUMER "$stream" "$GROUP" "$consumer" 2>&1)
  echo "  $stream / $consumer  → deleted_pending=$result"
  if [[ "$result" =~ ^[0-9]+$ ]]; then
    total_deleted=$((total_deleted + result))
  fi
done

echo
echo "=== After ==="
for stream in "${!CONSUMERS[@]}"; do
  group_pending=$(redis XINFO GROUPS "$stream" 2>/dev/null | awk -v g="$GROUP" '
    /^name$/{getline n}
    n==g && /^pending$/{getline p; print p; exit}
  ')
  printf "  %-22s roomengine_pending=%s\n" "$stream" "${group_pending:-0}"
done

echo
echo "Total stale pending deleted: $total_deleted"
echo "Done. If wisefido-sensor is running, redis will recreate the consumers on next XReadGroup."
