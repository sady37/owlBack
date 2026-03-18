#!/bin/bash
# 订阅 Docker MQTT 主题 sleepace-57136，过滤 BM87224601903 对应设备（device_code=8amzqonkfyfyy）
# device_uid=BM87224601903  device_id=ba204a77-9b51-4ba9-802b-2f1d9f1245f7  device_code=8amzqonkfyfyy
MQTT_DEVICE_ID="8amzqonkfyfyy"
TOPIC="sleepace-57136"
HOST="${MQTT_HOST:-host.docker.internal}"
USER="${MQTT_USER:-wfiot}"
PASS="${MQTT_PASS:-tt@wf@2025}"

echo "Subscribing to $TOPIC (deviceId=$MQTT_DEVICE_ID, BM87224601903)..."
echo "Filter: deviceSenSor + realtime + connectionStatus (Ctrl+C to stop)"
echo ""

docker run --rm --add-host=host.docker.internal:host-gateway eclipse-mosquitto:2.0 \
  mosquitto_sub -h "$HOST" -p 1883 -t "$TOPIC" -u "$USER" -P "$PASS" -v 2>/dev/null | while read -r line; do
  if [[ "$line" == *"$MQTT_DEVICE_ID"* ]]; then
    if [[ "$line" == *"deviceSenSor"* ]]; then
      echo "[deviceSenSor] $line"
    elif [[ "$line" == *"connectionStatus"* ]]; then
      echo "[connectionStatus] $line"
    elif [[ "$line" == *"realtime"* ]]; then
      # 可选：只打前 120 字符避免刷屏
      echo "[realtime] ${line:0:200}..."
    else
      echo "$line"
    fi
  fi
done
