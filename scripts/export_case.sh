#!/usr/bin/env bash
# 导出一个 case fixture（room_layout + iot_timeseries 时间窗）到 doc/cases/<name>/，
# 格式与 doc/cases/d5f7-ghost/ 完全一致。
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
#     room_layout.json         整行 rooms 行 (room_id/room_name/layout_config)
#     <range>.json             iot_timeseries 行数组，字段同 d5f7-ghost fixture
#                              (id, device_uid, device_id, timestamp, topic_type,
#                               category, data_value)

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
PGDATABASE="${DB_NAME:-owlrd}"
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

ROOM_ID=$(RUN_PSQL -t -A -c "
  SELECT bound_room_id FROM devices WHERE device_uid='$UID_ARG' LIMIT 1;
" | tr -d '[:space:]')

if [[ -z "$ROOM_ID" ]]; then
  echo "ERROR: device_uid=$UID_ARG not found in devices, or no bound_room_id" >&2
  exit 2
fi
echo "room_id   : $ROOM_ID"

RUN_PSQL -t -A -c "
  SELECT json_build_object(
    'room_id',     room_id,
    'room_name',   room_name,
    'layout_config', layout_config
  )::text
  FROM rooms WHERE room_id='$ROOM_ID';
" > "$LAYOUT_FILE"
echo "layout    : $(wc -c < "$LAYOUT_FILE") bytes -> $LAYOUT_FILE"

RUN_PSQL -t -A -c "
  SELECT COALESCE(
    json_agg(
      json_build_object(
        'id',          id,
        'device_uid',  device_uid,
        'device_id',   device_id,
        'timestamp',   timestamp,
        'topic_type',  topic_type,
        'category',    category,
        'data_value',  data_value
      )
      ORDER BY timestamp, id
    ),
    '[]'::json
  )::text
  FROM iot_timeseries
  WHERE device_uid='$UID_ARG'
    AND timestamp >= $START_MS
    AND timestamp <= $END_MS;
" > "$WINDOW_FILE"

ROW_COUNT=$(RUN_PSQL -t -A -c "
  SELECT count(*) FROM iot_timeseries
  WHERE device_uid='$UID_ARG' AND timestamp BETWEEN $START_MS AND $END_MS;
" | tr -d '[:space:]')
echo "rows      : $ROW_COUNT  -> $WINDOW_FILE"
echo "Done."
