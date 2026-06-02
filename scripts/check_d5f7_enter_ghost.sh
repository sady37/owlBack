#!/usr/bin/env bash
# 检查某雷达最近 N 天: EnterRoom 会话 → 每会话 trackID 数量 / 越界帧 / pose → ghost 算法是否出判决。
# 用法: ./check_d5f7_enter_ghost.sh [DEVICE_UID] [DAYS]
set -euo pipefail

UID_HEX="${1:-E598A2ACD5F7}"
DAYS="${2:-3}"
TZ_LOCAL="America/Denver"

export PGPASSWORD=postgres
PSQL() { psql -h 127.0.0.1 -U postgres -d owl_v2 -t -A -F'|' "$@"; }
export REDISCLI_AUTH="TeLunSu-36kr"

ADDR=$(PSQL -c "SELECT device_addr FROM devices WHERE device_uid='${UID_HEX}';" | head -1)
[ -z "$ADDR" ] && { echo "device_uid ${UID_HEX} 未找到"; exit 1; }
echo "=== device ${UID_HEX} = ${ADDR}  (最近 ${DAYS} 天, TZ=${TZ_LOCAL}) ==="

echo
echo "### 1. Enter/Exit 事件配平"
PSQL -c "SELECT to_char(ts AT TIME ZONE '${TZ_LOCAL}','MM-DD HH24:MI:SS'), event_kind, coalesce(payload->>'reason','')
         FROM event_log WHERE device_addr='${ADDR}' AND event_kind IN ('EnterRoom','ExitRoom')
           AND ts >= now()-interval '${DAYS} days' ORDER BY ts;" \
| awk -F'|' '{printf "  %-17s %-10s %s\n",$1,$2,$3}'
ENTER=$(PSQL -c "SELECT count(*) FROM event_log WHERE device_addr='${ADDR}' AND event_kind='EnterRoom' AND ts>=now()-interval '${DAYS} days';")
EXIT=$(PSQL -c "SELECT count(*) FROM event_log WHERE device_addr='${ADDR}' AND event_kind='ExitRoom' AND ts>=now()-interval '${DAYS} days';")
echo "  -> EnterRoom=${ENTER}  ExitRoom=${EXIT}  (进无出=$((ENTER-EXIT)) = lost_track 风险会话)"

echo
echo "### 2. 每个 EnterRoom 会话: trackID 数量 / 越界帧 / pose 分布"
echo "    (会话窗 = EnterRoom 起, 至下一个 ExitRoom 或 +20min; 越界 = |V|>90cm 超 FOV 深度)"
PSQL -c "SELECT extract(epoch FROM ts)::bigint, to_char(ts AT TIME ZONE '${TZ_LOCAL}','MM-DD HH24:MI:SS')
         FROM event_log WHERE device_addr='${ADDR}' AND event_kind='EnterRoom'
           AND ts>=now()-interval '${DAYS} days' ORDER BY ts;" \
| while IFS='|' read -r EPOCH LABEL; do
    WIN_END=$(PSQL -c "SELECT coalesce(extract(epoch FROM min(ts))::bigint, ${EPOCH}+1200)
                       FROM event_log WHERE device_addr='${ADDR}' AND event_kind='ExitRoom'
                         AND ts > to_timestamp(${EPOCH});")
    echo "  ── Enter @ ${LABEL}  (窗 $((WIN_END-EPOCH))s)"
    PSQL -c "
      WITH f AS (
        SELECT (t->>'track_id')::int tid, (t->>'position_y')::int v, t->>'pose' pose
        FROM monitor_stream, LATERAL jsonb_array_elements(payload) t
        WHERE device_addr='${ADDR}' AND stream_type='radar.track'
          AND ts BETWEEN to_timestamp(${EPOCH}) AND to_timestamp(${WIN_END})
          AND (t->>'track_id')::int <> 88 )
      SELECT tid, count(*) frames, count(*) FILTER (WHERE abs(v)>90) oob,
             min(v), max(v), string_agg(DISTINCT coalesce(pose,'-'),',' ORDER BY coalesce(pose,'-'))
      FROM f GROUP BY tid ORDER BY tid;" \
    | awk -F'|' 'BEGIN{print "      trackID | frames | 越界帧 | Vmin | Vmax | poses"}
                 {printf "      %-7s | %-6s | %-6s | %-4s | %-4s | %s\n",$1,$2,$3,$4,$5,$6}'
    # 该窗内是否触发 Fall
    FALL=$(PSQL -c "SELECT to_char(triggered_at AT TIME ZONE '${TZ_LOCAL}','HH24:MI:SS')||' '||coalesce(payload->>'reason','')
                    FROM alarm_events WHERE device_addr='${ADDR}' AND event_type='Fall'
                      AND triggered_at BETWEEN to_timestamp(${EPOCH}) AND to_timestamp(${WIN_END}+600);")
    [ -n "$FALL" ] && echo "      >> Fall 触发: ${FALL}" || true
  done

echo
echo "### 3. ghost 算法判决 (ai:track:verdict:stream, 该设备)"
redis-cli XRANGE ai:track:verdict:stream - + 2>/dev/null | python3 -c "
import sys,json,datetime,re
addr='${ADDR}'
lines=[l.rstrip('\n') for l in sys.stdin]
ID=re.compile(r'^\d+-\d+\$')
entries=[]; cur=None
i=0
while i < len(lines):
    if ID.match(lines[i]):
        cur={}; entries.append(cur); i+=1
    elif cur is not None and i+1 < len(lines):
        cur[lines[i]]=lines[i+1]; i+=2
    else:
        i+=1
hits=0
for f in entries:
    if f.get('device_addr')!=addr: continue
    try: dv=json.loads(f.get('dataValue','[]'))
    except Exception: dv=[]
    for d in dv:
        ts=int(f.get('timestamp','0'))/1000
        t=datetime.datetime.fromtimestamp(ts).strftime('%m-%d %H:%M:%S')
        ev=d.get('evidence',{}) or {}
        print('  %s tid=%s conf=%s reason=%-13s pose=%s pos=(%s,%s) ctx=%s'%(
            t,d.get('track_id'),d.get('track_confidence'),d.get('reason'),
            d.get('pose'),d.get('position_x'),d.get('position_y'),ev.get('context')))
        hits+=1
print('  -> 该设备 ghost 判决 %d 条 (stream 全局共 %d 条, 最早 %s)'%(
    hits, len(entries),
    datetime.datetime.fromtimestamp(int(entries[0].get('timestamp','0'))/1000).strftime('%m-%d %H:%M') if entries else '-'))
"
