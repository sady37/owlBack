#!/usr/bin/env bash
# Subscribe to Radar MQTT monitor topic, collect 100 messages and export to file
# Usage: ./mqtt_export_radar_100.sh [output-file]
# Default output: scripts/mqtt_radar_100_<YYYYMMDD_HHMMSS>.txt

set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/../.env"

# Topic defaults to monitor for product_id 88, all uids
TOPIC="${MQTT_TOPIC:-/monitor/88/+/post}"
MQTT_HOST="${MQTT_HOST:-host.docker.internal}"
MQTT_PORT="${MQTT_PORT:-1883}"
MQTT_USER="${MQTT_USERNAME:-wfiot}"
MQTT_PASS="${MQTT_PASSWORD:-tt@wf@2025}"
COUNT="${MQTT_COUNT:-100}"

[ -f "$ENV_FILE" ] && set -a && source "$ENV_FILE" && set +a
[ -n "$MQTT_PASSWORD" ] && MQTT_PASS="$MQTT_PASSWORD"

OUTFILE="${1:-${SCRIPT_DIR}/mqtt_radar_${COUNT}_$(date +%Y%m%d_%H%M%S).txt}"
echo "Topic: $TOPIC | Host: $MQTT_HOST:$MQTT_PORT | Count: $COUNT -> $OUTFILE"
echo "Subscribing..."
docker run --rm --add-host=host.docker.internal:host-gateway eclipse-mosquitto:2 \
  mosquitto_sub -h "$MQTT_HOST" -p "$MQTT_PORT" -t "$TOPIC" -u "$MQTT_USER" -P "$MQTT_PASS" -v -C "$COUNT" 2>/dev/null \
  | tee "$OUTFILE"
echo ""
echo "Exported $COUNT messages to $OUTFILE"
