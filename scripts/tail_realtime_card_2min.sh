#!/bin/bash
# 统计 card 42077c6d-ed05-46ec-a76d-b45ddb48b24f (BM87224601903) 的 card:realtime:stream 写入周期
# 从 stream 最近 2 分钟的消息中按 card_id 过滤并计算间隔
CARD_ID="42077c6d-ed05-46ec-a76d-b45ddb48b24f"
STREAM="card:realtime:stream"
REDIS_AUTH="${REDIS_PASSWORD:-TeLunSu-36kr}"

# 2 分钟 = 120000 ms，起始 ID 为当前时间-2min
NOW_MS=$(($(date +%s) * 1000))
START_MS=$((NOW_MS - 120000))
START_ID="${START_MS}-0"

echo "Card: $CARD_ID (BM87224601903)"
echo "Stream: $STREAM, from ID $START_ID (last 2 min)"
echo "---"

TMP=$(mktemp)
trap "rm -f $TMP" EXIT

# XRANGE 返回多行: id, field, value, field, value... 需解析出每条消息的 id 和 card_id
docker exec owl-redis redis-cli -a "$REDIS_AUTH" XRANGE "$STREAM" "$START_ID" + COUNT 5000 2>/dev/null | awk -v card="$CARD_ID" '
/^[0-9]+-[0-9]+$/ { id=$0; next }
$0=="card_id" { getline val; if(val==card) print id; next }
' > "$TMP"

COUNT=$(wc -l < "$TMP" | tr -d ' ')
echo "Collected $COUNT messages for this card."
if [ "$COUNT" -lt 2 ]; then
  echo "Need at least 2 messages to compute period."
  exit 0
fi

# stream ID 格式 ms-seq，取 ms 部分算间隔
awk -F'-' '{print $1}' "$TMP" | awk 'NR>1{print $0-prev} {prev=$0}' | sort -n | awk '
BEGIN { sum=0; n=0 }
{ sum+=$1; n++; d[n]=$1 }
END {
  if(n>0) {
    printf "Intervals (ms): min=%d max=%d avg=%.0f count=%d\n", d[1], d[n], sum/n, n
    printf "Period: ~%.2f s\n", (sum/n)/1000
  }
}'
