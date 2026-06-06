#!/usr/bin/env bash
# export_bed_test_record.sh
# 导出一次"床态测试"的人类可读完整时间线 txt（radar.track + radar.heart + sleepad.track
# + event_log + alarm_events），逐秒按时间排序，并【自动派生】 bed_state / room_state 的
# >>> STATE 转换标记行（基于 InBed/LeftBed 与 EnterRoom/ExitRoom/number_people 事件，不硬编码）。
#
# 与 export_case.sh 的区别：那个导 JSON replay fixture 喂 belief_replay_test；本脚本导
# 人类可读 txt 供测试复盘/对账（bed 融合、固件 fall、sleepad-radar 矛盾等）。
#
# 用法:
#   ./export_bed_test_record.sh <radar_uid> <sleepad_uid> <start> <end> --tz <IANA_TZ> [<case_name>]
#
# 示例:
#   ./export_bed_test_record.sh 9D8A326309E7 BM87224700978 \
#       "2026-06-05 13:08:30" "2026-06-05 13:18:45" --tz America/Denver bed-test-0605-2
#
# 参数:
#   <radar_uid>    雷达 device_uid（发 radar.track/heart + 固件 InBed/LeftBed/Fall）
#   <sleepad_uid>  床垫 device_uid（发 sleepad.track + sleepad InBed/LeftBed）；'-' 表示无床垫
#   <start> <end>  本地时间 "YYYY-MM-DD HH:MM:SS"，按 --tz 解释（DB 内部存 UTC）
#   --tz <TZ>      必填，IANA 时区（如 America/Denver）。服务器 $TZ 未必等于设备所在地。
#   <case_name>    可选；缺省自动拼 "bedtest_<radar后4>_<startISO>"
#
# 输出:
#   doc/cases/<case_name>/test_record.txt
#
# 环境变量（可覆盖）: PGHOST=localhost PGPORT=5432 PGUSER=postgres PGPASSWORD=postgres PGDB=owl_v2

set -euo pipefail

if [[ $# -lt 4 ]]; then sed -n '2,30p' "$0"; exit 1; fi

RADAR_UID="$1"; SLEEPAD_UID="$2"; START_ARG="$3"; END_ARG="$4"; shift 4
TZ_ARG=""; CASE_NAME=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --tz) TZ_ARG="$2"; shift 2 ;;
    *) CASE_NAME="$1"; shift ;;
  esac
done
[[ -z "$TZ_ARG" ]] && { echo "ERROR: --tz <IANA_TZ> 必填"; exit 1; }

PGHOST="${PGHOST:-localhost}"; PGPORT="${PGPORT:-5432}"
PGUSER="${PGUSER:-postgres}"; PGPASSWORD="${PGPASSWORD:-postgres}"; PGDB="${PGDB:-owl_v2}"
export PGPASSWORD
PSQL=(psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDB" -tA)

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OWLBACK_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# uid -> /128 device_addr
RADAR_ADDR="$("${PSQL[@]}" -c "SELECT host(device_addr) FROM devices WHERE device_uid='${RADAR_UID}' LIMIT 1;")"
[[ -z "$RADAR_ADDR" ]] && { echo "ERROR: 找不到 radar uid=$RADAR_UID"; exit 1; }
SLEEPAD_ADDR=""
if [[ "$SLEEPAD_UID" != "-" ]]; then
  SLEEPAD_ADDR="$("${PSQL[@]}" -c "SELECT host(device_addr) FROM devices WHERE device_uid='${SLEEPAD_UID}' LIMIT 1;")"
  [[ -z "$SLEEPAD_ADDR" ]] && { echo "ERROR: 找不到 sleepad uid=$SLEEPAD_UID"; exit 1; }
fi

# 本地时间 -> UTC timestamptz 字面量（psql 内用 AT TIME ZONE 反算）
WS="$("${PSQL[@]}" -c "SELECT (timestamp '${START_ARG}' AT TIME ZONE '${TZ_ARG}')::text;")"
WE="$("${PSQL[@]}" -c "SELECT (timestamp '${END_ARG}'   AT TIME ZONE '${TZ_ARG}')::text;")"

if [[ -z "$CASE_NAME" ]]; then
  R4="${RADAR_UID: -4}"; ISO="$(echo "$START_ARG" | tr ' :' '__')"
  CASE_NAME="bedtest_${R4}_${ISO}"
fi
OUT_DIR="${OWLBACK_DIR}/doc/cases/${CASE_NAME}"
mkdir -p "$OUT_DIR"
OUT="${OUT_DIR}/test_record.txt"

# 设备地址集合（event/alarm 同时查两台；track/heart 限 radar；sleepad.track 限 sleepad）
if [[ -n "$SLEEPAD_ADDR" ]]; then DEVSET="'${RADAR_ADDR}','${SLEEPAD_ADDR}'"; else DEVSET="'${RADAR_ADDR}'"; fi
SLP_FILTER="AND device_addr='${SLEEPAD_ADDR}'"; [[ -z "$SLEEPAD_ADDR" ]] && SLP_FILTER="AND false"

# 计数
cnt(){ "${PSQL[@]}" -c "$1"; }
N_TRACK=$(cnt "SELECT count(*) FROM monitor_stream WHERE stream_type='radar.track' AND device_addr='${RADAR_ADDR}' AND ts BETWEEN '${WS}' AND '${WE}';")
N_HEART=$(cnt "SELECT count(*) FROM monitor_stream WHERE stream_type='radar.heart' AND device_addr='${RADAR_ADDR}' AND ts BETWEEN '${WS}' AND '${WE}';")
N_SLP=$(cnt "SELECT count(*) FROM monitor_stream WHERE stream_type='sleepad.track' ${SLP_FILTER} AND ts BETWEEN '${WS}' AND '${WE}';")
N_EVT=$(cnt "SELECT count(*) FROM event_log WHERE device_addr IN (${DEVSET}) AND ts BETWEEN '${WS}' AND '${WE}';")
N_ALM=$(cnt "SELECT count(*) FROM alarm_events WHERE device_addr IN (${DEVSET}) AND triggered_at BETWEEN '${WS}' AND '${WE}';")

# 主体（含自动派生 STATE 行）
BODY="$("${PSQL[@]}" -c "
WITH base AS (
  SELECT ts AS ts, 1 AS pri, right(host(device_addr),4) AS dev,
    coalesce(payload->0->>'track_id','-') AS tid, 'track         ' AS typ,
    format('pose=%s xy=(%s,%s) z=%s area=%s conf=%s', payload->0->>'pose', payload->0->>'position_x',
      payload->0->>'position_y', payload->0->>'position_z', payload->0->>'area_id',
      payload->0->>'track_confidence') AS line
  FROM monitor_stream WHERE stream_type='radar.track' AND device_addr='${RADAR_ADDR}' AND ts BETWEEN '${WS}' AND '${WE}'
  UNION ALL
  SELECT ts, 1, right(host(device_addr),4), coalesce(payload->0->>'track_id','-'), 'radar.heart   ',
    format('conf=%s (payload 不含 HR/RR)', payload->0->>'track_confidence')
  FROM monitor_stream WHERE stream_type='radar.heart' AND device_addr='${RADAR_ADDR}' AND ts BETWEEN '${WS}' AND '${WE}'
  UNION ALL
  SELECT ts, 1, right(host(device_addr),4), coalesce(payload->0->>'track_id','-'), 'sleepad.track ',
    format('bed=%s move=%s turn=%s HR=%s RR=%s pc=%s vc=%s', coalesce(payload->0->>'bed_status','-'),
      coalesce(payload->0->>'body_move','-'), coalesce(payload->0->>'turn_over','-'), coalesce(payload->0->>'heart_rate','-'),
      coalesce(payload->0->>'respiratory_rate','-'), coalesce(payload->0->>'pose_confidence','-'),
      coalesce(payload->0->>'vital_confidence','-'))
  FROM monitor_stream WHERE stream_type='sleepad.track' ${SLP_FILTER} AND ts BETWEEN '${WS}' AND '${WE}'
  UNION ALL
  SELECT ts, 1, right(host(device_addr),4), coalesce(payload->0->>'track_id','-'), 'EVENT '||rpad(event_kind,12),
    (SELECT string_agg(k||'='||v,' ' ORDER BY k) FROM jsonb_each_text(payload->0) e(k,v))
  FROM event_log WHERE device_addr IN (${DEVSET}) AND ts BETWEEN '${WS}' AND '${WE}'
  UNION ALL
  -- 按 triggered_at(事故时刻)过滤+落位：事故在窗内就保留，落在摔倒那一刻；真 fire 时间靠标注说明
  -- (回填型 alarm triggered_at=事故锚点 ≠ alerted_at=真 fire；reason 列空时回落 payload.reason)
  SELECT triggered_at, 1, right(host(device_addr),4), coalesce(payload->>'track_id','-'), '*** ALARM     ',
    format('%s lvl=%s reason=%s status=%s%s', event_type, alarm_level,
      coalesce(nullif(reason,''), payload->>'reason', '-'), alarm_status,
      CASE WHEN alerted_at IS NOT NULL AND alerted_at <> triggered_at
        THEN format(' [事故@%s 回填→fire@%s]', to_char(triggered_at AT TIME ZONE 'UTC','HH24:MI:SS'),
               to_char(alerted_at AT TIME ZONE 'UTC','HH24:MI:SS'))
        ELSE coalesce(' pose='||(payload->>'pose'), '') END)
  FROM alarm_events WHERE device_addr IN (${DEVSET}) AND triggered_at BETWEEN '${WS}' AND '${WE}'
  -- ===== 自动派生 STATE（bed_state: InBed/LeftBed 事件; room_state: Enter/Exit/np=0）=====
  UNION ALL
  SELECT ts, 0, '---', '-', '>>> STATE     ',
    'bed_state -> '||CASE event_kind WHEN 'InBed' THEN 'InBed' ELSE 'NotInBed' END
      ||'  (来源: '||right(host(device_addr),4)||' '||event_kind||')'
  FROM event_log WHERE device_addr IN (${DEVSET}) AND event_kind IN ('InBed','LeftBed') AND ts BETWEEN '${WS}' AND '${WE}'
  UNION ALL
  SELECT ts, 0, '---', '-', '>>> STATE     ',
    'room_state -> '||CASE event_kind WHEN 'EnterRoom' THEN 'Occupied' ELSE 'Vacant' END
      ||'  (来源: '||right(host(device_addr),4)||' '||event_kind||')'
  FROM event_log WHERE device_addr IN (${DEVSET}) AND event_kind IN ('EnterRoom','ExitRoom') AND ts BETWEEN '${WS}' AND '${WE}'
)
SELECT to_char(ts AT TIME ZONE 'UTC','HH24:MI:SS') ||'  | '|| dev ||' | tid'|| lpad(tid,3) ||' | '|| typ ||' | '|| line
FROM base ORDER BY ts, pri
")"

{
  echo "================================================================================"
  echo " 床态测试完整记录   case=${CASE_NAME}"
  echo " 窗口(本地 ${TZ_ARG}): ${START_ARG} – ${END_ARG}"
  echo " 窗口(UTC): ${WS} – ${WE}   (下表时间列为 UTC, HH:MM:SS)"
  echo " radar   = ${RADAR_UID} (${RADAR_ADDR})"
  [[ -n "$SLEEPAD_ADDR" ]] && echo " sleepad = ${SLEEPAD_UID} (${SLEEPAD_ADDR})"
  echo "--------------------------------------------------------------------------------"
  echo " 列: 时间(UTC) | dev(addr后4) | tid(track_id) | 类型 | 明细"
  echo " tid: 真人 0-8 / 9=vital / 10=space(np 事件) / 11=device(心跳) / 88=无目标 / - =派生态无轨"
  echo " 行类型: track(radar.track ${N_TRACK}) / radar.heart(${N_HEART},无HR/RR) / sleepad.track(${N_SLP})"
  echo "        / EVENT(event_log ${N_EVT}) / *** ALARM(${N_ALM}) / >>> STATE(自动派生 bed/room 转换)"
  echo "--------------------------------------------------------------------------------"
  echo " 数据流水线: 固件 --MQTT--> qinglan(解码 mqtt rx) --Redis--> wisefido-iot(写 event_log/monitor) --> PG"
  echo " STATE 派生说明: bed_state 取 InBed/LeftBed 事件(radar 固件 Enter2Out 与 sleepad 各发各的);"
  echo "   room_state 取 EnterRoom/ExitRoom。均为转换点标注, 非逐秒填充。窗口起始前的态需看更早事件。"
  echo "================================================================================"
  echo ""
  echo "$BODY"
} > "$OUT"

echo "wrote: $OUT"
echo "  lines: $(wc -l < "$OUT")  (track=$N_TRACK heart=$N_HEART sleepad=$N_SLP event=$N_EVT alarm=$N_ALM)"
