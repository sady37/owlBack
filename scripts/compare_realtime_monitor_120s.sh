#!/bin/bash
# 120 秒内按 device_id 同时统计 iot:monitor:stream 与 card:realtime:stream，对照周期与中断次数
# Sleepad BM87224601903 → device_id ba204a77-9b51-4ba9-802b-2f1d9f1245f7, card 42077c6d-ed05-46ec-a76d-b45ddb48b24f
DEVICE_ID="ba204a77-9b51-4ba9-802b-2f1d9f1245f7"
CARD_ID="42077c6d-ed05-46ec-a76d-b45ddb48b24f"
REDIS_AUTH="${REDIS_PASSWORD:-TeLunSu-36kr}"
GAP_MS=3000   # 间隔超过此值计为一次中断
GAP_2S=2000   # 另统计 >2s 次数（约 7 次中断时可用）

NOW_MS=$(($(date +%s) * 1000))
START_MS=$((NOW_MS - 120000))
START_ID="${START_MS}-0"

TMP_M=$(mktemp)
TMP_C=$(mktemp)
trap "rm -f $TMP_M $TMP_C" EXIT

echo "Device: $DEVICE_ID (Sleepad BM87224601903)"
echo "Card:   $CARD_ID"
echo "Window: last 120s (from $START_ID)"
echo "Gap > ${GAP_MS}ms = 1 interruption; also reporting gap > ${GAP_2S}ms"
echo "=========================================="

# 1) iot:monitor:stream: 按 device_id 过滤，收集 stream ID (ms)
docker exec owl-redis redis-cli -a "$REDIS_AUTH" XRANGE iot:monitor:stream "$START_ID" + COUNT 5000 2>/dev/null | awk -v dev="$DEVICE_ID" '
/^[0-9]+-[0-9]+$/ { id=$0; next }
$0=="device_id" { getline val; if(val==dev) print id; next }
' | awk -F'-' '{print $1}' > "$TMP_M"

# 2) card:realtime:stream: 该 card 下 data 含此 device_id 的消息（data 的 key 为 device_id）
docker exec owl-redis redis-cli -a "$REDIS_AUTH" XRANGE card:realtime:stream "$START_ID" + COUNT 5000 2>/dev/null | awk -v card="$CARD_ID" -v dev="$DEVICE_ID" '
/^[0-9]+-[0-9]+$/ { id=$0; want=0; next }
$0=="card_id" { getline val; want=(val==card)?1:0; next }
$0=="data" && want { getline json; if(index(json, dev)>0) print id; next }
' | awk -F'-' '{print $1}' > "$TMP_C"

n_m=$(wc -l < "$TMP_M" | tr -d ' ')
n_c=$(wc -l < "$TMP_C" | tr -d ' ')

echo ""
echo "--- iot:monitor:stream (device_id=$DEVICE_ID) ---"
echo "Messages: $n_m"
if [ "$n_m" -ge 2 ]; then
  awk 'NR>1{print $0-prev} {prev=$0}' "$TMP_M" | sort -n > "${TMP_M}.gap"
  int_m=$(awk -v g="$GAP_MS" '$1>g{c++} END{print c+0}' "${TMP_M}.gap")
  int2_m=$(awk -v g="$GAP_2S" '$1>g{c++} END{print c+0}' "${TMP_M}.gap")
  min_m=$(head -1 "${TMP_M}.gap")
  max_m=$(tail -1 "${TMP_M}.gap")
  avg_m=$(awk '{s+=$1;n++} END{printf "%.0f", (n>0)?s/n:0}' "${TMP_M}.gap")
  echo "Intervals (ms): min=$min_m max=$max_m avg=$avg_m"
  echo "Interruptions: gap>${GAP_2S}ms: $int2_m  gap>${GAP_MS}ms: $int_m"
  echo "Gap distribution: 2-3s: $(awk '$1>2000&&$1<=3000{c++} END{print c+0}' "${TMP_M}.gap")  3-5s: $(awk '$1>3000&&$1<=5000{c++} END{print c+0}' "${TMP_M}.gap")  5-7s: $(awk '$1>5000&&$1<=7000{c++} END{print c+0}' "${TMP_M}.gap")  >7s: $(awk '$1>7000{c++} END{print c+0}' "${TMP_M}.gap")"
fi

echo ""
echo "--- card:realtime:stream (card + data contains device) ---"
echo "Messages: $n_c"
if [ "$n_c" -ge 2 ]; then
  awk 'NR>1{print $0-prev} {prev=$0}' "$TMP_C" | sort -n > "${TMP_C}.gap"
  int_c=$(awk -v g="$GAP_MS" '$1>g{c++} END{print c+0}' "${TMP_C}.gap")
  int2_c=$(awk -v g="$GAP_2S" '$1>g{c++} END{print c+0}' "${TMP_C}.gap")
  min_c=$(head -1 "${TMP_C}.gap")
  max_c=$(tail -1 "${TMP_C}.gap")
  avg_c=$(awk '{s+=$1;n++} END{printf "%.0f", (n>0)?s/n:0}' "${TMP_C}.gap")
  echo "Intervals (ms): min=$min_c max=$max_c avg=$avg_c"
  echo "Interruptions: gap>${GAP_2S}ms: $int2_c  gap>${GAP_MS}ms: $int_c"
  echo "Gap distribution: 2-3s: $(awk '$1>2000&&$1<=3000{c++} END{print c+0}' "${TMP_C}.gap")  3-5s: $(awk '$1>3000&&$1<=5000{c++} END{print c+0}' "${TMP_C}.gap")  5-7s: $(awk '$1>5000&&$1<=7000{c++} END{print c+0}' "${TMP_C}.gap")  >7s: $(awk '$1>7000{c++} END{print c+0}' "${TMP_C}.gap")"
fi

echo ""
echo "--- Summary ---"
echo "iot:monitor (upstream) count=$n_m  card:realtime (agg) count=$n_c"
if [ "$n_m" -ge 2 ] && [ "$n_c" -ge 2 ]; then
  echo "iot:monitor interruptions(>3s)=$int_m  card:realtime interruptions(>3s)=$int_c"
fi
echo ""
echo "Note: iot:monitor = Sleepad->wisefido-sleepace->stream; card:realtime = cardagg aggregate (same card may include radar)."
