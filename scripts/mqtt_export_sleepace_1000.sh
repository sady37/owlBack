#!/usr/bin/env bash
# 直接订阅 Sleepace MQTT 主题，收取 1000 条后退出并导出到文件
# 用法: ./mqtt_export_sleepace_1000.sh [输出文件路径]
# 默认输出: scripts/mqtt_sleepace_1000_<YYYYMMDD_HHMMSS>.txt

set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/../.env"
TOPIC="${MQTT_TOPIC:-sleepace-57136}"
MQTT_HOST="${MQTT_HOST:-host.docker.internal}"
MQTT_PORT="${MQTT_PORT:-1883}"
MQTT_USER="${MQTT_USERNAME:-wfiot}"
MQTT_PASS="${MQTT_PASSWORD:-tt@wf@2025}"
COUNT="${MQTT_COUNT:-1000}"

[ -f "$ENV_FILE" ] && set -a && source "$ENV_FILE" && set +a
[ -n "$MQTT_PASSWORD" ] && MQTT_PASS="$MQTT_PASSWORD"

OUTFILE="${1:-${SCRIPT_DIR}/mqtt_sleepace_${COUNT}_$(date +%Y%m%d_%H%M%S).txt}"
echo "Topic: $TOPIC | Host: $MQTT_HOST:$MQTT_PORT | Count: $COUNT -> $OUTFILE"
echo "Subscribing..."
docker run --rm --add-host=host.docker.internal:host-gateway eclipse-mosquitto:2 \
  mosquitto_sub -h "$MQTT_HOST" -p "$MQTT_PORT" -t "$TOPIC" -u "$MQTT_USER" -P "$MQTT_PASS" -v -C "$COUNT" 2>/dev/null \
  | tee "$OUTFILE"
echo ""
echo "Exported $COUNT messages to $OUTFILE"
