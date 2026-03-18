#!/usr/bin/env bash
# 订阅 sleepace MQTT 主题并追加到 mqtt-sleepace.log，用于排查 BM87224601903(1ua3erivl9pv1) / card 42077c6d 的 sleepStage、awake 来源
# 用法: ./mqtt_sub_sleepace_to_log.sh [日志路径]
# 默认: scripts/mqtt-sleepace.log ；Ctrl+C 停止
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

LOGFILE="${1:-${SCRIPT_DIR}/mqtt-sleepace.log}"
LOGDIR="$(dirname "$LOGFILE")"
LOGNAME="$(basename "$LOGFILE")"

# 若未指定 MQTT_HOST 或为 host.docker.internal，优先加入 owl-mqtt 所在网络，直连服务名 mqtt 避免 Network unreachable
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

# 若未加入 broker 网络会 Network unreachable，请先启动 owl-mqtt
[ -z "$NET_OPT" ] && echo "Warning: owl-mqtt not found, using MQTT_HOST=$MQTT_HOST (container may get Network unreachable)"
echo "Subscribing $TOPIC -> $LOGFILE (append). MQTT: $MQTT_HOST:$MQTT_PORT  Ctrl+C to stop. Live: tail -f $LOGFILE"
# 与 mqtt_subscribe_by_device_code.sh 一致：alpine + script 提供 TTY；-a 追加到日志
docker run --rm $NET_OPT --add-host=host.docker.internal:host-gateway -v "${LOGDIR}:/log:rw" -w /log \
  -e "MQTT_HOST=$MQTT_HOST" -e "MQTT_PORT=$MQTT_PORT" -e "TOPIC=$TOPIC" -e "MQTT_USER=$MQTT_USER" -e "MQTT_PASS=$MQTT_PASS" -e "LOGNAME=$LOGNAME" \
  alpine sh -c '
    apk add -q mosquitto-clients util-linux 2>/dev/null
    script -q -a -c "mosquitto_sub -h \$MQTT_HOST -p \$MQTT_PORT -t \$TOPIC -u \$MQTT_USER -P \$MQTT_PASS -v 2>&1" /log/"$LOGNAME"
  '