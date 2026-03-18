#!/usr/bin/env bash
# 清空 iot:*:stream 与 card:*:stream（重启 owlBack 前清理 MQTT/Redis 数据）
# 用法: ./redis_trim_streams.sh [redis_password]
# 默认密码 TeLunSu-36kr
set -e
PWD="${1:-TeLunSu-36kr}"
REDIS_CMD="docker exec owl-redis redis-cli -a $PWD"

echo "=== 清空 iot:*:stream ==="
for key in $($REDIS_CMD KEYS "iot:*:stream" 2>/dev/null); do
  $REDIS_CMD DEL "$key" 2>/dev/null && echo "DEL $key"
done

echo "=== 清空 card:*:stream ==="
for key in $($REDIS_CMD KEYS "card:*:stream" 2>/dev/null); do
  $REDIS_CMD DEL "$key" 2>/dev/null && echo "DEL $key"
done

echo "=== 当前 iot/card stream 键 ==="
$REDIS_CMD KEYS "iot:*:stream" 2>/dev/null || true
$REDIS_CMD KEYS "card:*:stream" 2>/dev/null || true
echo "完成。"
