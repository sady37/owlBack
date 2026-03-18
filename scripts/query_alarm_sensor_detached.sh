#!/usr/bin/env bash
# 查询 SensorDetached 历史：alarm_events、MQTT 日志(deviceSenSor + alarmNotify/alarmSensorFall)、iot:alarm:stream、card:state
# 用法: ./query_alarm_sensor_detached.sh [device_uid] [sleepace_log_path]
# 依赖: owlBack/.env, Docker owl-redis, owl-postgresql；可选 sleepace-service 日志路径

set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/../.env"
if [ -f "$ENV_FILE" ]; then
  set -a
  source "$ENV_FILE"
  set +a
fi

REDIS_PASSWORD="${REDIS_PASSWORD:-TeLunSu-36kr}"
DEVICE_UID="${1:-BM87224601903}"
# sleepace-service 日志：dataKey=deviceSenSor 与 dataKey=alarmNotify(type=alarmSensorFall) 均会打出
SLEEPACE_LOG="${2:-${SCRIPT_DIR}/../../sleepace/sleepace-service/log/server.out}"

echo "=== 1) alarm_events 表 SensorDetached 历史 (device_uid=$DEVICE_UID) ==="
docker exec owl-postgresql psql -U postgres -d owlrd -c "
SELECT ae.event_id, ae.device_id, ae.event_type, ae.alarm_status, ae.triggered_at
FROM alarm_events ae
JOIN device_store ds ON ds.device_id = ae.device_id
WHERE ae.event_type = 'SensorDetached'
  AND (ds.device_uid = '$DEVICE_UID' OR ds.device_code = '$DEVICE_UID' OR ae.device_id::text IN (SELECT device_id::text FROM device_store WHERE device_uid = '$DEVICE_UID' OR device_code = '$DEVICE_UID'))
ORDER BY ae.triggered_at DESC
LIMIT 20;
" 2>/dev/null

echo ""
echo "=== 2) MQTT 历史：sleepace 日志中 deviceSenSor + alarmNotify(alarmSensorFall) (device_code 由 device_uid 解析) ==="
# 解析 device_uid 对应的 device_code（可能多行）
while IFS= read -r code; do
  code=$(echo "$code" | tr -d ' ')
  [ -n "$code" ] && DEVICE_CODES="${DEVICE_CODES:+$DEVICE_CODES }$code"
done < <(docker exec owl-postgresql psql -U postgres -d owlrd -t -A -c "
  SELECT device_code FROM device_store WHERE device_uid = '$DEVICE_UID' AND device_code IS NOT NULL AND device_code != '';
" 2>/dev/null)
if [ -z "$DEVICE_CODES" ]; then
  echo "未解析到 device_code，跳过日志检索"
else
  echo "device_code(s): $DEVICE_CODES"
  if [ -f "$SLEEPACE_LOG" ]; then
    echo "--- dataKey=deviceSenSor (deviceId=上述 code) 最近 30 条 ---"
    for code in $DEVICE_CODES; do
      grep "deviceSenSor" "$SLEEPACE_LOG" 2>/dev/null | grep "\"deviceId\":\"$code\"" | tail -30
    done
    echo "--- dataKey=alarmNotify 且 alarmSensorFall/SensorDetached (deviceId=上述 code) 最近 30 条 ---"
    for code in $DEVICE_CODES; do
      grep "alarmNotify" "$SLEEPACE_LOG" 2>/dev/null | grep -i "alarmSensorFall\|SensorDetached" | grep "\"deviceId\":\"$code\"" | tail -30
    done
  else
    echo "日志不存在: $SLEEPACE_LOG"
  fi
fi

echo ""
echo "=== 3) Redis iot:alarm:stream (最近含 device_uid / SensorDetached) ==="
docker exec owl-redis redis-cli -a "$REDIS_PASSWORD" XREVRANGE iot:alarm:stream + - COUNT 500 2>/dev/null | grep -B1 -A22 "device_uid\|SensorDetached" | head -120

echo ""
echo "=== 4) card:state 中 pop_alarm (该设备所属 card) ==="
CARD_ID=$(docker exec owl-postgresql psql -U postgres -d owlrd -t -c "
  SELECT c.card_id FROM cards c, jsonb_array_elements(c.devices) d
  WHERE d->>'device_id' IN (SELECT device_id::text FROM device_store WHERE device_uid = '$DEVICE_UID')
  LIMIT 1;
" 2>/dev/null | tr -d ' ')
if [ -n "$CARD_ID" ]; then
  echo "card_id: $CARD_ID"
  docker exec owl-redis redis-cli -a "$REDIS_PASSWORD" HGET "card:state:$CARD_ID" alarm_state 2>/dev/null
else
  echo "未找到 device_uid=$DEVICE_UID 的 card"
fi
