#!/usr/bin/env bash
# 看 MQTT (sleepace 主题) 中 payload 里 timeStamp 最早的时间
# 用法:
#   ./mqtt_earliest_time.sh                    # 订阅约 15 秒，取收到消息中的最早 timeStamp
#   ./mqtt_earliest_time.sh /path/to/mqtt.log  # 从已有日志文件提取最早 timeStamp
# timeStamp 为秒；与 iot:monitor:stream 对比时需统一单位（monitor 的 stream id 为 ms）
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/../.env"
TOPIC="${MQTT_TOPIC:-sleepace-57136}"
MQTT_HOST="${MQTT_HOST:-host.docker.internal}"
MQTT_PORT="${MQTT_PORT:-1883}"
MQTT_USER="${MQTT_USERNAME:-wfiot}"
MQTT_PASS="${MQTT_PASSWORD:-tt@wf@2025}"

[ -f "$ENV_FILE" ] && set -a && source "$ENV_FILE" && set +a
[ -n "$MQTT_PASSWORD" ] && MQTT_PASS="$MQTT_PASSWORD"

extract_earliest() {
  grep -oE '"timeStamp"[[:space:]]*:[[:space:]]*[0-9]+' "$@" | grep -oE '[0-9]+' | sort -n | head -1
}

if [[ -n "$1" && -f "$1" ]]; then
  echo "=== 从文件读取: $1 ==="
  FIRST_SEC=$(extract_earliest "$1")
else
  echo "=== MQTT 订阅 ${TOPIC} 约 15 秒，取最早 timeStamp ==="
  TMP=$(mktemp)
  trap "rm -f $TMP" EXIT
  timeout 15 docker run --rm --add-host=host.docker.internal:host-gateway eclipse-mosquitto:2 \
    mosquitto_sub -h "$MQTT_HOST" -p "$MQTT_PORT" -t "$TOPIC" -u "$MQTT_USER" -P "$MQTT_PASS" -v -C 500 2>/dev/null > "$TMP" || true
  FIRST_SEC=$(extract_earliest "$TMP")
fi

if [[ -z "$FIRST_SEC" ]]; then
  echo "未取到 timeStamp（无消息或格式不符）"
  exit 0
fi

FIRST_MS=$((FIRST_SEC * 1000))
echo "最早 timeStamp(秒): $FIRST_SEC"
echo "最早 timeStamp(ms): $FIRST_MS"
if command -v date &>/dev/null; then
  (date -r "$FIRST_SEC" 2>/dev/null || date -d "@$FIRST_SEC" 2>/dev/null) | xargs echo "对应时间:"
fi
echo ""
echo "与 iot:monitor:stream 对比: 运行 ./scripts/redis_monitor_earliest.sh 看 monitor 最早条目的 id(ms)。"
