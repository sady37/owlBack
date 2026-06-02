#!/usr/bin/env bash
# 导出一个 case fixture（room_layout + monitor_stream 时间窗）到 doc/cases/<name>/，
# 格式与 doc/cases/d5f7-ghost/ 一致，可被 belief_replay_test.go 直接消费。
#
# **owl_v2 版**：源是 monitor_stream（ts=timestamptz / device_addr=inet / payload=data_value）。
# 老 export_case.sh 读 iot_timeseries（owl_v2 已无此表），对 owl_v2 不适用，故另立此脚本。
#
# 用法:
#   ./export_case_v2.sh <device_uid> <start_ms> <end_ms> <case_name>
# 示例:
#   ./export_case_v2.sh 4D8710F41797 1780278311000 1780279261000 mom-bathroom-lost-0953
#
# 参数:
#   <start_ms> <end_ms>  UTC epoch 毫秒（monitor_stream.ts 即 UTC，无需 tz）。
#
# 输出:
#   doc/cases/<case_name>/
#     room_layout.json     room_visual_layout.canvas（/128 device layout，objects[] 顶层）
#     window.json          [{category, device_uid, timestamp, data_value}, ...]（track/heart + 事件）

set -euo pipefail

if [[ $# -lt 4 ]]; then
  sed -n '2,20p' "$0"
  exit 1
fi

UID_ARG="$1"; START_MS="$2"; END_MS="$3"; CASE_NAME="$4"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
OUT_DIR="$ROOT_DIR/doc/cases/$CASE_NAME"
mkdir -p "$OUT_DIR"

export PGPASSWORD="${DB_PASSWORD:-postgres}"
PSQL() { psql -h "${DB_HOST:-localhost}" -p "${DB_PORT:-5432}" -U "${DB_USER:-postgres}" -d "${DB_NAME:-owl_v2}" -tA "$@"; }

ADDR=$(PSQL -c "SELECT host(device_addr) FROM devices WHERE device_uid='$UID_ARG' LIMIT 1;" | tr -d '[:space:]')
if [[ -z "$ADDR" ]]; then echo "ERROR: device_uid=$UID_ARG not found" >&2; exit 2; fi
echo "device_uid: $UID_ARG / addr: $ADDR / window: $START_MS..$END_MS -> $OUT_DIR"

# layout：/128 device canvas（buildGridFromLayout 认顶层 objects[]）
PSQL -c "SELECT canvas::text FROM room_visual_layout WHERE spatial_prefix='$ADDR'::inet LIMIT 1;" > "$OUT_DIR/room_layout.json"
echo "layout: $(wc -c < "$OUT_DIR/room_layout.json") bytes"

# window：monitor_stream（radar.track→track / radar.heart→heart）+ event_log（事件）合并，按 ts 排序
PSQL -c "
WITH mon AS (
  SELECT ts, json_build_object(
    'category', replace(stream_type,'radar.',''),
    'device_uid', '$UID_ARG',
    'timestamp', (extract(epoch from ts)*1000)::bigint,
    'data_value', payload
  ) rec
  FROM monitor_stream
  WHERE device_addr='$ADDR'::inet AND ts BETWEEN to_timestamp($START_MS/1000.0) AND to_timestamp($END_MS/1000.0)
), evt AS (
  SELECT ts, json_build_object(
    'category', event_kind,
    'device_uid', '$UID_ARG',
    'timestamp', (extract(epoch from ts)*1000)::bigint,
    'data_value', COALESCE(payload, '[]'::jsonb)
  ) rec
  FROM event_log
  WHERE device_addr='$ADDR'::inet AND ts BETWEEN to_timestamp($START_MS/1000.0) AND to_timestamp($END_MS/1000.0)
)
SELECT COALESCE(json_agg(rec ORDER BY ts),'[]'::json)::text
FROM (SELECT ts, rec FROM mon UNION ALL SELECT ts, rec FROM evt) u;
" > "$OUT_DIR/window.json"

ROWS=$(PSQL -c "SELECT count(*) FROM monitor_stream WHERE device_addr='$ADDR'::inet AND ts BETWEEN to_timestamp($START_MS/1000.0) AND to_timestamp($END_MS/1000.0);" | tr -d '[:space:]')
echo "window: $(wc -c < "$OUT_DIR/window.json") bytes / monitor rows=$ROWS"
echo "Done."
