#!/usr/bin/env bash
# 导出一个 case fixture（room_layout + 时序时间窗）到 doc/cases/<name>/。
# owl_v2 schema：track 帧取 monitor_stream（设备级），事件取 event_log（房级，含 sleepad
# InBed/LeftBed —— bed-fall fixture 必需）；layout 取 room_visual_layout.canvas。
# 输出 record 形如 {device_uid,timestamp(ms),topic_type,category,data_value}，
# category='track' 对齐 belief_replay_test 的 case "track"。
#
# 用法:
#   ./export_case.sh <device_uid> <start> <end> --tz <IANA_TZ> [<case_name>]
#
# 示例:
#   ./export_case.sh 9923003AB17F "2026-04-29 07:10:00" "2026-04-29 07:30:00" \
#       --tz America/Denver 9923-kitchen-0429
#   ./export_case.sh E598A2ACD5F7 "2026-04-25 17:20:00" "2026-04-25 17:35:00" \
#       --tz America/Denver d5f7-ghost-replay
#
# 参数:
#   <start> <end>  本地时间字符串 ("YYYY-MM-DD HH:MM:SS")，时区按 --tz 解释。
#   --tz <TZ>      **必填**，IANA 时区名 (e.g. America/Denver, America/Los_Angeles)。
#                  服务器 $TZ 与设备所在地不一定相同；显式指定避免偏 1 小时。
#                  iot_timeseries.timestamp 存的是 UTC epoch ms，本脚本只在
#                  把"本地时间"换成 epoch-ms 时使用 --tz。
#   <case_name>    可选；缺省自动拼成 "<uid 后 4>_<startISO>_to_<endISO>_<TZ_abbr>"
#
# 输出:
#   doc/cases/<case_name>/
#     room_layout.json         {room_id, room_name, layout_config:canvas}（/128 设备级优先，回落 /88）
#     <range>.json             record 数组 {device_uid, timestamp(ms), topic_type, category, data_value}
#                              track 帧 category='track'；事件 category=event_kind（InBed/LeftBed/...）

set -euo pipefail

if [[ $# -lt 3 ]]; then
  sed -n '2,28p' "$0"
  exit 1
fi

UID_ARG="$1"; START_ARG="$2"; END_ARG="$3"; shift 3
CASE_NAME=""
TZ_ARG=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --tz) TZ_ARG="$2"; shift 2 ;;
    --tz=*) TZ_ARG="${1#*=}"; shift ;;
    *) CASE_NAME="$1"; shift ;;
  esac
done

if [[ -z "$TZ_ARG" ]]; then
  echo "ERROR: --tz <IANA_TZ> is required (e.g. America/Denver, America/Los_Angeles)." >&2
  echo "       The DB stores UTC epoch ms; you must declare which timezone you mean" >&2
  echo "       when you write '07:10:00'. Don't assume server \$TZ matches the device." >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
CASES_DIR="$ROOT_DIR/doc/cases"

if [[ -f "$ROOT_DIR/.env" ]]; then
  set -a; source "$ROOT_DIR/.env"; set +a
fi
PGHOST="${DB_HOST:-127.0.0.1}"
PGPORT="${DB_PORT:-5432}"
PGUSER="${DB_USER:-postgres}"
PGPASSWORD="${DB_PASSWORD:-postgres}"
PGDATABASE="${DB_NAME:-owl_v2}"
export PGHOST PGPORT PGUSER PGPASSWORD PGDATABASE

RUN_PSQL() {
  # match container by suffix（some setups have hashed prefix like e8cd…_owl-postgresql）
  local pg_container
  pg_container=$(docker ps --format '{{.Names}}' 2>/dev/null | grep 'owl-postgresql' | head -1 || true)
  if [[ -n "$pg_container" ]]; then
    docker exec -i "$pg_container" psql -U "$PGUSER" -d "$PGDATABASE" "$@"
  else
    psql "$@"
  fi
}

START_MS=$(TZ="$TZ_ARG" date -d "$START_ARG" +%s%3N)
END_MS=$(TZ="$TZ_ARG"   date -d "$END_ARG"   +%s%3N)

if [[ -z "$CASE_NAME" ]]; then
  short_uid="$(echo "$UID_ARG" | tr 'A-Z' 'a-z' | tail -c 5)"
  s_iso=$(TZ="$TZ_ARG" date -d "$START_ARG" +%Y-%m-%d_%H-%M)
  e_iso=$(TZ="$TZ_ARG" date -d "$END_ARG"   +%H-%M)
  tz_abbr=$(TZ="$TZ_ARG" date -d "$START_ARG" +%Z)
  CASE_NAME="${short_uid}_${s_iso}_to_${e_iso}_${tz_abbr}"
fi

OUT_DIR="$CASES_DIR/$CASE_NAME"
mkdir -p "$OUT_DIR"

s_iso_full=$(TZ="$TZ_ARG" date -d "$START_ARG" +%Y-%m-%d_%H-%M)
e_iso_full=$(TZ="$TZ_ARG" date -d "$END_ARG"   +%H-%M)
tz_abbr=$(TZ="$TZ_ARG" date -d "$START_ARG" +%Z)
WINDOW_FILE="$OUT_DIR/${s_iso_full}_to_${e_iso_full}_${tz_abbr}.json"
LAYOUT_FILE="$OUT_DIR/room_layout.json"

echo "case_name : $CASE_NAME"
echo "device_uid: $UID_ARG"
echo "tz        : $TZ_ARG"
echo "window    : $START_ARG -> $END_ARG  ($START_MS -> $END_MS ms)"
echo "out_dir   : $OUT_DIR"

# owl_v2: devices 无 bound_room_id；device_addr(/128) 反查所属 /88 room（IPv6 前缀包含）。
DEV_ADDR=$(RUN_PSQL -t -A -c "
  SELECT host(device_addr) FROM devices WHERE device_uid='$UID_ARG' LIMIT 1;
" | tr -d '[:space:]')
if [[ -z "$DEV_ADDR" ]]; then
  echo "ERROR: device_uid=$UID_ARG not found in devices" >&2
  exit 2
fi
ROOM_ID=$(RUN_PSQL -t -A -c "
  SELECT room_id FROM rooms WHERE masklen(room_id)=88 AND '${DEV_ADDR}'::inet <<= room_id LIMIT 1;
" | tr -d '[:space:]')
if [[ -z "$ROOM_ID" ]]; then
  echo "ERROR: no /88 room contains device $DEV_ADDR" >&2
  exit 2
fi
echo "dev_addr  : $DEV_ADDR"
echo "room_id   : $ROOM_ID"

# owl_v2: layout 在 room_visual_layout.canvas（按 spatial_prefix）；优先 /128 设备级（playback 口径），
# 回落 /88 房级。包成 {room_id,room_name,layout_config} 与 buildGridFromLayout 的解壳一致。
RUN_PSQL -t -A -c "
  SELECT json_build_object(
    'room_id',   r.room_id,
    'room_name', r.room_name,
    'layout_config', COALESCE(
      (SELECT canvas FROM room_visual_layout WHERE spatial_prefix='${DEV_ADDR}'::inet),
      (SELECT canvas FROM room_visual_layout WHERE spatial_prefix=r.room_id)
    )
  )::text
  FROM rooms r WHERE r.room_id='$ROOM_ID';
" > "$LAYOUT_FILE"
echo "layout    : $(wc -c < "$LAYOUT_FILE") bytes -> $LAYOUT_FILE"

# owl_v2: iot_timeseries 拆成 monitor_stream(track 帧,设备级,带坐标) + event_log(InBed/LeftBed/
# Enter/Exit 等,房级 — 含 sleepad 事件,bed-fall fixture 必需)。映射成 replay fxRecord
# {device_uid,timestamp(ms),topic_type,category,data_value}；category='track' 对齐 belief_replay 的 case "track"。
RUN_PSQL -t -A -c "
  SELECT COALESCE(json_agg(rec ORDER BY ts_ms), '[]'::json)::text FROM (
    SELECT (extract(epoch FROM m.ts)*1000)::bigint AS ts_ms,
           json_build_object(
             'device_uid', dm.device_uid,
             'timestamp',  (extract(epoch FROM m.ts)*1000)::bigint,
             'topic_type', 'monitor',
             'category',   replace(replace(m.stream_type::text, 'radar.', ''), 'sleepad.', ''),
             'data_value', m.payload
           ) AS rec
      FROM monitor_stream m JOIN devices dm ON dm.device_addr=m.device_addr
     WHERE m.device_addr <<= '$ROOM_ID'::inet
       AND m.ts BETWEEN to_timestamp($START_MS/1000.0) AND to_timestamp($END_MS/1000.0)
    UNION ALL
    SELECT (extract(epoch FROM e.ts)*1000)::bigint AS ts_ms,
           json_build_object(
             'device_uid', d.device_uid,
             'timestamp',  (extract(epoch FROM e.ts)*1000)::bigint,
             'topic_type', 'event',
             'category',   e.event_kind,
             'data_value', e.payload
           ) AS rec
      FROM event_log e JOIN devices d ON d.device_addr=e.device_addr
     WHERE e.device_addr <<= '$ROOM_ID'::inet
       AND e.ts BETWEEN to_timestamp($START_MS/1000.0) AND to_timestamp($END_MS/1000.0)
  ) t;
" > "$WINDOW_FILE"

ROW_COUNT=$(RUN_PSQL -t -A -c "
  SELECT
    (SELECT count(*) FROM monitor_stream WHERE device_addr='${DEV_ADDR}'::inet
       AND ts BETWEEN to_timestamp($START_MS/1000.0) AND to_timestamp($END_MS/1000.0))
   +(SELECT count(*) FROM event_log WHERE device_addr <<= '$ROOM_ID'::inet
       AND ts BETWEEN to_timestamp($START_MS/1000.0) AND to_timestamp($END_MS/1000.0));
" | tr -d '[:space:]')
echo "rows      : $ROW_COUNT  -> $WINDOW_FILE"
echo "Done."
