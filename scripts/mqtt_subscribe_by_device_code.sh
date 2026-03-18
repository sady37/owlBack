#!/usr/bin/env bash
# 订阅 MQTT，看 deviceId(device_code) + dataKey + timeStamp 即可快速判定 SensorDetached 来源
# 用法: ./mqtt_subscribe_by_device_code.sh [device_code]
# 不传 device_code 则打全量；传了则只 grep 该设备

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
[ -n "$RADAR_MQTT_PASSWORD" ] && MQTT_PASS="${MQTT_PASS:-$RADAR_MQTT_PASSWORD}"

FILTER="$1"
LOGFILE="${SCRIPT_DIR}/sleepace.log"
LOGDIR="${SCRIPT_DIR}"

# 优先加入 owl-mqtt 所在网络，直连服务名 mqtt 避免 Network unreachable
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

echo "topic=$TOPIC  (device_code 过滤: ${FILTER:-无}，仅提示；筛设备请 grep deviceId $LOGFILE)"
echo "append -> $LOGFILE  MQTT: $MQTT_HOST:$MQTT_PORT  实时: tail -f $LOGFILE"
echo ""

# 用 alpine + script 给 mosquitto_sub 提供 TTY，避免 Bad file descriptor；实时看用 tail -f sleepace.log
docker run --rm $NET_OPT --add-host=host.docker.internal:host-gateway -v "${LOGDIR}:/log:rw" -w /log \
  -e "MQTT_HOST=$MQTT_HOST" -e "MQTT_PORT=$MQTT_PORT" -e "TOPIC=$TOPIC" -e "MQTT_USER=$MQTT_USER" -e "MQTT_PASS=$MQTT_PASS" \
  alpine sh -c '
    apk add -q mosquitto-clients util-linux 2>/dev/null
    script -q -c "mosquitto_sub -h \"\$MQTT_HOST\" -p \"\$MQTT_PORT\" -t \"\$TOPIC\" -u \"\$MQTT_USER\" -P \"\$MQTT_PASS\" -v 2>&1" /log/sleepace.log
  '
