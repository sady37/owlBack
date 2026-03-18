#!/bin/bash
# 2 分钟内同时统计 3 路：MQTT realtime (deviceId=1ua3erivl9pv1)、iot:monitor:stream、card:realtime:stream
# 保证 3 个数据流口径一致，HR/RR 约 2 秒 1 次
# MQTT: docker mosquitto_sub 订阅 sleepace-57136，过滤 realtime + deviceId
# Redis: 同 device_id ba204a77... / card 42077c6d...
MQTT_DEVICE_ID="1ua3erivl9pv1"
DEVICE_ID="ba204a77-9b51-4ba9-802b-2f1d9f1245f7"
CARD_ID="42077c6d-ed05-46ec-a76d-b45ddb48b24f"
REDIS_AUTH="${REDIS_PASSWORD:-TeLunSu-36kr}"
DURATION=120
GAP_MS=3000
GAP_2S=2000

TMP_MQTT=$(mktemp)
TMP_M=$(mktemp)
TMP_C=$(mktemp)
trap "rm -f $TMP_MQTT $TMP_M $TMP_C ${TMP_MQTT}.gap ${TMP_M}.gap ${TMP_C}.gap 2>/dev/null" EXIT

echo "=== 3-stream comparison (${DURATION}s) ==="
echo "MQTT deviceId: $MQTT_DEVICE_ID (Sleepad BM87224601903)"
echo "Redis device_id: $DEVICE_ID  card_id: $CARD_ID"
echo "Gap >${GAP_2S}ms / >${GAP_MS}ms = interruption"
echo ""

# ---- 1) MQTT realtime: 订阅 120s，过滤 realtime + deviceId，提取 payload 内 timeStamp(秒)→转 ms ----
echo "Collecting MQTT realtime (topic sleepace-57136) for ${DURATION}s..."
(
  docker run --rm --add-host=host.docker.internal:host-gateway eclipse-mosquitto:2 \
    mosquitto_sub -h host.docker.internal -p 1883 -t "sleepace-57136" -u wfiot -P "tt@wf@2025" -v -C 999999 2>/dev/null &
  SUB_PID=$!
  sleep "$DURATION"
  kill $SUB_PID 2>/dev/null
  wait $SUB_PID 2>/dev/null
) | while read -r line; do
  if [[ "$line" == *"$MQTT_DEVICE_ID"* && "$line" == *"realtime"* ]]; then
    # payload 内 timeStamp 为秒，转 ms 便于与 Redis 对比
    ts_sec=$(echo "$line" | grep -oE '"timeStamp"[[:space:]]*:[[:space:]]*[0-9]+' | head -1 | grep -oE '[0-9]+')
    if [[ -n "$ts_sec" ]]; then echo $((ts_sec * 1000)); fi
  fi
done | sort -n > "$TMP_MQTT"
n_mqtt=$(wc -l < "$TMP_MQTT" | tr -d ' ')
echo "MQTT realtime messages (deviceId=$MQTT_DEVICE_ID): $n_mqtt"

# ---- 2) Redis: 最近 120s 窗口 ----
NOW_MS=$(($(date +%s) * 1000))
START_MS=$((NOW_MS - 120000))
START_ID="${START_MS}-0"

docker exec owl-redis redis-cli -a "$REDIS_AUTH" XRANGE iot:monitor:stream "$START_ID" + COUNT 5000 2>/dev/null | awk -v dev="$DEVICE_ID" '
/^[0-9]+-[0-9]+$/ { id=$0; next }
$0=="device_id" { getline val; if(val==dev) print id; next }
' | awk -F'-' '{print $1}' > "$TMP_M"

docker exec owl-redis redis-cli -a "$REDIS_AUTH" XRANGE card:realtime:stream "$START_ID" + COUNT 5000 2>/dev/null | awk -v card="$CARD_ID" -v dev="$DEVICE_ID" '
/^[0-9]+-[0-9]+$/ { id=$0; want=0; next }
$0=="card_id" { getline val; want=(val==card)?1:0; next }
$0=="data" && want { getline json; if(index(json, dev)>0) print id; next }
' | awk -F'-' '{print $1}' > "$TMP_C"

n_m=$(wc -l < "$TMP_M" | tr -d ' ')
n_c=$(wc -l < "$TMP_C" | tr -d ' ')

# ---- 3) 统一输出：条数、间隔、中断、分布 ----
report() {
  local name="$1"
  local file="$2"
  local n="$3"
  echo "--- $name ---"
  echo "Messages: $n"
  if [[ "$n" -ge 2 ]]; then
    awk 'NR>1{print $0-prev} {prev=$0}' "$file" | sort -n > "${file}.gap"
    int=$(awk -v g="$GAP_MS" '$1>g{c++} END{print c+0}' "${file}.gap")
    int2=$(awk -v g="$GAP_2S" '$1>g{c++} END{print c+0}' "${file}.gap")
    min=$(head -1 "${file}.gap")
    max=$(tail -1 "${file}.gap")
    avg=$(awk '{s+=$1;n++} END{printf "%.0f", (n>0)?s/n:0}' "${file}.gap")
    echo "Intervals (ms): min=$min max=$max avg=$avg"
    echo "Interruptions: gap>${GAP_2S}ms: $int2  gap>${GAP_MS}ms: $int"
    echo "Gap distribution: 2-3s: $(awk '$1>2000&&$1<=3000{c++} END{print c+0}' "${file}.gap")  3-5s: $(awk '$1>3000&&$1<=5000{c++} END{print c+0}' "${file}.gap")  5-7s: $(awk '$1>5000&&$1<=7000{c++} END{print c+0}' "${file}.gap")  >7s: $(awk '$1>7000{c++} END{print c+0}' "${file}.gap")"
  fi
  echo ""
}

report "1) MQTT realtime (sleepace-57136, deviceId=$MQTT_DEVICE_ID)" "$TMP_MQTT" "$n_mqtt"
report "2) iot:monitor:stream (device_id=$DEVICE_ID)" "$TMP_M" "$n_m"
report "3) card:realtime:stream (card+device)" "$TMP_C" "$n_c"

echo "--- Summary (3 streams should be roughly consistent; HR/RR ~2s expected) ---"
echo "MQTT: $n_mqtt  iot:monitor: $n_m  card:realtime: $n_c"
