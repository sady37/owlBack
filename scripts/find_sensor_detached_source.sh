#!/usr/bin/env bash
# 定位 SensorDetached 报警源：在 MQTT 消费侧日志中查找是否有 deviceSenSor status=0
# 用法: ./find_sensor_detached_source.sh <device_uid> [alarm_time]
#  例: ./find_sensor_detached_source.sh BM87224601903 "2026-03-09 14:32:55"
#  不传 alarm_time 时自动取该设备最近一条 SensorDetached 的 triggered_at
# 可选第3参: sleepace 日志路径；第4参: wisefido-sleepace 日志路径（Docker 需先 docker logs > file）
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/../.env"
[ -f "$ENV_FILE" ] && set -a && source "$ENV_FILE" && set +a

DEVICE_UID="${1:?usage: $0 <device_uid> [alarm_time]}"
ALARM_TIME_RAW="${2:-}"
SLEEPACE_LOG="${3:-${SCRIPT_DIR}/../../sleepace/sleepace-service/log/server.out}"
WISEFIDO_LOG="${4:-}"

# 解析 device_code
DEVICE_CODES=$(docker exec owl-postgresql psql -U postgres -d owlrd -t -A -c "
  SELECT device_code FROM device_store WHERE device_uid = '$DEVICE_UID' AND device_code IS NOT NULL AND device_code != '';
" 2>/dev/null | tr -d '\r' | xargs)
if [ -z "$DEVICE_CODES" ]; then
  echo "未解析到 device_uid=$DEVICE_UID 的 device_code"
  exit 1
fi
echo "device_uid=$DEVICE_UID → device_code=$DEVICE_CODES"

# 若未传时间，取该设备最近一条 SensorDetached 的 triggered_at
if [ -z "$ALARM_TIME_RAW" ]; then
  ALARM_TIME_RAW=$(docker exec owl-postgresql psql -U postgres -d owlrd -t -A -c "
    SELECT triggered_at::text FROM alarm_events ae
    JOIN device_store ds ON ds.device_id = ae.device_id
    WHERE ae.event_type = 'SensorDetached' AND ds.device_uid = '$DEVICE_UID'
    ORDER BY ae.triggered_at DESC LIMIT 1;
  " 2>/dev/null | tr -d '\r')
  [ -z "$ALARM_TIME_RAW" ] && echo "该设备无 SensorDetached 记录" && exit 1
  echo "使用最近一条 SensorDetached: triggered_at=$ALARM_TIME_RAW"
fi

# 从 DB 查该设备最近 SensorDetached（不依赖传入时间窗，避免时区导致 0 rows）
echo ""
echo "========== DB 中该告警记录 =========="
docker exec owl-postgresql psql -U postgres -d owlrd -c "
  SELECT ae.event_id, ae.triggered_at, ae.triggered_at AT TIME ZONE 'UTC' AS triggered_utc, ae.event_type
  FROM alarm_events ae
  JOIN device_store ds ON ds.device_id = ae.device_id
  WHERE ae.event_type = 'SensorDetached' AND ds.device_uid = '$DEVICE_UID'
  ORDER BY ae.triggered_at DESC
  LIMIT 5;
" 2>/dev/null

SEARCH_DATE=$(echo "$ALARM_TIME_RAW" | sed 's/T/ /; s/Z//; s/+.*//' | cut -d' ' -f1)
SEARCH_HHMM=$(echo "$ALARM_TIME_RAW" | sed 's/T/ /; s/Z//' | cut -d' ' -f2 | cut -d':' -f1,2)
# DB 最近一条的 triggered_at (UTC) 时:分，用于日志为 UTC 时
DB_HHMM=$(docker exec owl-postgresql psql -U postgres -d owlrd -t -A -c "
  SELECT to_char(triggered_at AT TIME ZONE 'UTC', 'HH24:MI') FROM alarm_events ae
  JOIN device_store ds ON ds.device_id = ae.device_id
  WHERE ae.event_type = 'SensorDetached' AND ds.device_uid = '$DEVICE_UID'
  ORDER BY ae.triggered_at DESC LIMIT 1;
" 2>/dev/null | tr -d '\r')
# 避免 grep -E "14:32|" 产生 empty subexpression
if [ -n "$DB_HHMM" ] && [ -n "$SEARCH_HHMM" ] && [ "$DB_HHMM" != "$SEARCH_HHMM" ]; then
  TIMEGREP="$SEARCH_HHMM|$DB_HHMM"
else
  TIMEGREP="${DB_HHMM:-$SEARCH_HHMM}"
fi
echo ""
echo "告警时间窗: 日期=$SEARCH_DATE 用户时:分=$SEARCH_HHMM DB_UTC时:分=$DB_HHMM → grep时:分=$TIMEGREP"

# ----- 1) sleepace-service 日志（MQTT 上游）-----
echo ""
echo "========== 1) sleepace-service 日志中 deviceSenSor / alarmNotify (device_code=$DEVICE_CODES) =========="
if [ -f "$SLEEPACE_LOG" ]; then
  for code in $DEVICE_CODES; do
    echo "--- deviceId=$code 且含 deviceSenSor 或 alarmNotify，日期 $SEARCH_DATE 时间约 $SEARCH_HHMM 或 UTC $DB_HHMM ---"
    { grep -E "deviceSenSor|alarmNotify" "$SLEEPACE_LOG" 2>/dev/null | grep "\"deviceId\":\"$code\"" | grep "$SEARCH_DATE" | grep -E "$TIMEGREP"; } || true
    echo "--- 时间窗内 status=0（脱落）---"
    { grep "deviceSenSor" "$SLEEPACE_LOG" 2>/dev/null | grep "\"deviceId\":\"$code\"" | grep '"status":0' | grep "$SEARCH_DATE" | grep -E "$TIMEGREP"; } || true
  done
else
  echo "日志不存在: $SLEEPACE_LOG"
fi

# ----- 2) Redis iot:alarm:stream 中该设备的 SensorDetached 消息（wisefido-sleepace 写入）-----
REDIS_PASSWORD="${REDIS_PASSWORD:-TeLunSu-36kr}"
echo ""
echo "========== 2) Redis iot:alarm:stream 近期含 device_uid=$DEVICE_UID 或 device_code 的 SensorDetached =========="
docker exec owl-redis redis-cli -a "$REDIS_PASSWORD" XREVRANGE iot:alarm:stream + - COUNT 300 2>/dev/null | grep -B1 "SensorDetached" | grep -A1 "$DEVICE_UID\|$(echo $DEVICE_CODES | tr ' ' '|')" || true

# ----- 3) wisefido-sleepace 日志（消费 MQTT 后写 Redis，含 device_uid + status）-----
echo ""
echo "========== 3) wisefido-sleepace 日志中 deviceSenSor (device_uid=$DEVICE_UID, status=0=脱落) =========="
if [ -n "$WISEFIDO_LOG" ] && [ -f "$WISEFIDO_LOG" ]; then
  { grep "deviceSenSor" "$WISEFIDO_LOG" 2>/dev/null | grep "$DEVICE_UID" | grep "$SEARCH_DATE" | grep -E "$TIMEGREP"; } || true
  echo "--- 脱落源：status 0 ---"
  { grep "deviceSenSor" "$WISEFIDO_LOG" 2>/dev/null | grep "$DEVICE_UID" | grep 'status.*0\|"status":0' | grep "$SEARCH_DATE" | grep -E "$TIMEGREP"; } || true
else
  echo "未提供 wisefido-sleepace 日志路径（第4参）。可先: docker logs wisefido-sleepace 2>&1 | tee wisefido-sleepace.log"
fi

# ----- 判定 -----
echo ""
echo "========== 报警源判定 =========="
FOUND=0
if [ -f "$SLEEPACE_LOG" ]; then
  for code in $DEVICE_CODES; do
    if grep "deviceSenSor" "$SLEEPACE_LOG" 2>/dev/null | grep "\"deviceId\":\"$code\"" | grep '"status":0' | grep "$SEARCH_DATE" | grep -qE "$TIMEGREP"; then
      echo "【报警源】sleepace 日志中在该时间窗内存在 deviceSenSor status=0 (device_code=$code)，即设备/平台上报了传感器脱落。"
      FOUND=1
      break
    fi
  done
fi
if [ $FOUND -eq 0 ] && [ -n "$WISEFIDO_LOG" ] && [ -f "$WISEFIDO_LOG" ]; then
  if grep "deviceSenSor" "$WISEFIDO_LOG" 2>/dev/null | grep "$DEVICE_UID" | grep -E 'status.*0|"status":0' | grep "$SEARCH_DATE" | grep -qE "$TIMEGREP"; then
    echo "【报警源】wisefido-sleepace 日志中在该时间窗内存在 deviceSenSor status=0，即 MQTT 曾上报脱落并被本服务转发。"
    FOUND=1
  fi
fi
if [ $FOUND -eq 0 ]; then
  echo "【未在日志中找到报警源】该时间窗内未发现 deviceSenSor status=0。"
  # 若该时刻仅有 status=1，说明设备上报的是“恢复”，报警可能由更早的 status=0 触发或存在逻辑错误
  if [ -f "$SLEEPACE_LOG" ]; then
    for code in $DEVICE_CODES; do
      ONLY_RECOVER=$(grep "deviceSenSor" "$SLEEPACE_LOG" 2>/dev/null | grep "\"deviceId\":\"$code\"" | grep "$SEARCH_DATE" | grep -E "$TIMEGREP" | grep '"status":1' | wc -l | tr -d ' ')
      if [ -n "$ONLY_RECOVER" ] && [ "$ONLY_RECOVER" -gt 0 ]; then
        echo "  → 该时间窗内仅见 deviceSenSor status=1（已插上），未见 status=0。报警若非更早的 status=0 触发，则可能为误报或逻辑错误。"
        echo "  → 建议：查该时刻前 2 分钟内是否有 status=0（见下）。"
      fi
      break
    done
  fi
  echo "  请确认：1) 日志时间与 DB triggered_at 是否同一时区 2) 日志文件是否覆盖该时刻。"
fi
echo ""
echo "========== 告警时刻前 2 分钟内 deviceSenSor（查更早的 status=0）=========="
for code in $DEVICE_CODES; do
  grep "deviceSenSor" "$SLEEPACE_LOG" 2>/dev/null | grep "\"deviceId\":\"$code\"" | grep "$SEARCH_DATE" | head -50
done 2>/dev/null || true
echo ""
echo "完成。"
