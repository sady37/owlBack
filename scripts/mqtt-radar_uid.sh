#!/usr/bin/env bash
# 订阅 MQTT，仅用于 Radar device_uid（可选指定单个 uid）
# 用法: ./mqtt_subscribe_by_radar_uid.sh [DEVICE_UID]

set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/../.env"

# defaults (可通过环境变量覆盖)
MQTT_HOST="${MQTT_HOST:-host.docker.internal}"
MQTT_PORT="${MQTT_PORT:-1883}"
MQTT_USER="${MQTT_USERNAME:-wfiot}"
MQTT_PASS="${MQTT_PASSWORD:-tt@wf@2025}"

[ -f "$ENV_FILE" ] && set -a && source "$ENV_FILE" && set +a
[ -n "$MQTT_PASSWORD" ] && MQTT_PASS="$MQTT_PASSWORD"
[ -n "$RADAR_MQTT_PASSWORD" ] && MQTT_PASS="${MQTT_PASS:-$RADAR_MQTT_PASSWORD}"

UID_FILTER="$1"
DEFAULT_TOPIC="/monitor/88/+/post"

if [ -n "$UID_FILTER" ]; then
  TOPIC="/monitor/88/${UID_FILTER}/post"
  LOGNAME="radar_${UID_FILTER}.log"
  echo "Subscribing topic for single radar UID: ${UID_FILTER}"
else
  TOPIC="${MQTT_TOPIC:-$DEFAULT_TOPIC}"
  LOGNAME="radar_all.log"
  echo "Subscribing all radar messages (topic: ${TOPIC})"
fi

LOGDIR="${SCRIPT_DIR}"
LOGFILE="${LOGDIR}/${LOGNAME}"

# 如果有本地 mqtt 容器，优先加入同一网络以直接连接服务名 mqtt
NET_OPT=""
if [[ -z "$MQTT_HOST" || "$MQTT_HOST" == "host.docker.internal" ]]; then
  MQTT_NET=$(docker inspect owl-mqtt --format '{{range $k,$v := .NetworkSettings.Networks}}{{$k}}{{end}}' 2>/dev/null | head -1)
  if [[ -n "$MQTT_NET" ]]; then
    NET_OPT="--network $MQTT_NET"
    MQTT_HOST="mqtt"
  else
    MQTT_HOST="host.docker.internal"
  fi
fi

echo "MQTT: ${MQTT_HOST}:${MQTT_PORT}  USER: ${MQTT_USER}  LOG -> ${LOGFILE}"
echo "To watch live output: tail -f ${LOGFILE}"
echo "(Press Ctrl+C to stop subscription)"
echo ""

# Run mosquitto_sub inside small alpine container and write to mounted log file
docker run --rm ${NET_OPT} --add-host=host.docker.internal:host-gateway -v "${LOGDIR}:/log:rw" -w /log \
  -e "MQTT_HOST=${MQTT_HOST}" -e "MQTT_PORT=${MQTT_PORT}" -e "TOPIC=${TOPIC}" -e "MQTT_USER=${MQTT_USER}" -e "MQTT_PASS=${MQTT_PASS}" -e "LOGNAME=${LOGNAME}" \
  alpine sh -c '
    apk add -q mosquitto-clients util-linux 2>/dev/null || true
    script -q -c "mosquitto_sub -h \"$MQTT_HOST\" -p \"$MQTT_PORT\" -t \"$TOPIC\" -u \"$MQTT_USER\" -P \"$MQTT_PASS\" -v 2>&1" /log/$LOGNAME
  '
