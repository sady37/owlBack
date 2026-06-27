#!/usr/bin/env bash
# 输出指定 device_uid + 时间窗内的 alarm_events（含 payload / evidence 完整 jsonb）。
# 用法:
#   ./show_alarm_events.sh <UID> [START] [END]
#   ./show_alarm_events.sh E598A2ACD523 '2026-06-27 10:30' '2026-06-27 11:05'
#   ./show_alarm_events.sh E598A2ACD523 '2026-06-27'                 # 当天 00:00 → 次日 00:00
#   ./show_alarm_events.sh E598A2ACD523                              # 最近 24h
# 时间按本地时区 TZ（默认 America/Denver）解释；输出同样换算到该时区显示。
# 单事件深挖某个 event_id：再传第 4 参 → 只打那条的 payload/evidence。
#   ./show_alarm_events.sh E598A2ACD523 '' '' <EVENT_ID>

set -euo pipefail

UID_ARG="${1:?用法: $0 <UID> [START] [END] [EVENT_ID]}"
START_ARG="${2:-}"
END_ARG="${3:-}"
EVENT_ID="${4:-}"

TZ_NAME="${TZ:-America/Denver}"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
if [ -f "$SCRIPT_DIR/../.env" ]; then
  set -a; source "$SCRIPT_DIR/../.env"; set +a
fi
PGHOST="${DB_HOST:-localhost}"
PGPORT="${DB_PORT:-5432}"
PGUSER="${DB_USER:-postgres}"
PGPASSWORD="${DB_PASSWORD:-postgres}"
PGDATABASE="${DB_NAME:-owl_v2}"
export PGHOST PGPORT PGUSER PGPASSWORD PGDATABASE

# 时间窗默认值
if [ -z "$START_ARG" ] && [ -z "$EVENT_ID" ]; then
  START_ARG="now() - interval '24 hours'"; START_SQL="$START_ARG"
elif [ -n "$START_ARG" ]; then
  START_SQL="timestamp '$START_ARG' AT TIME ZONE '$TZ_NAME'"
fi
if [ -n "$END_ARG" ]; then
  END_SQL="timestamp '$END_ARG' AT TIME ZONE '$TZ_NAME'"
elif [ -n "$START_ARG" ] && [[ "$START_ARG" != now* ]] && [ ${#START_ARG} -le 10 ]; then
  # 只给了日期(YYYY-MM-DD) → 取整天
  END_SQL="(timestamp '$START_ARG' AT TIME ZONE '$TZ_NAME') + interval '1 day'"
else
  END_SQL="now()"
fi

PSQL() { psql -X -v ON_ERROR_STOP=1 "$@"; }

# 单事件模式
if [ -n "$EVENT_ID" ]; then
  echo "=== event_id=$EVENT_ID ==="
  PSQL -c "
    SELECT event_id, triggered_at AT TIME ZONE '$TZ_NAME' AS occurred_${TZ_NAME//\//_},
           alerted_at AT TIME ZONE '$TZ_NAME' AS fired,
           event_type, alarm_level, reason, category, room_name, device_uid,
           alarm_status, operation, handler_notes
    FROM alarm_events WHERE event_id = '$EVENT_ID';"
  echo "--- payload ---"
  PSQL -t -A -c "SELECT jsonb_pretty(payload) FROM alarm_events WHERE event_id='$EVENT_ID';"
  echo "--- evidence ---"
  PSQL -t -A -c "SELECT jsonb_pretty(evidence) FROM alarm_events WHERE event_id='$EVENT_ID';"
  exit 0
fi

echo "=== alarm_events: uid=$UID_ARG  window=[$START_SQL, $END_SQL)  tz=$TZ_NAME ==="
PSQL -c "
  SELECT event_id,
         triggered_at AT TIME ZONE '$TZ_NAME' AS occurred,
         alerted_at   AT TIME ZONE '$TZ_NAME' AS fired,
         EXTRACT(epoch FROM (alerted_at - triggered_at))::int AS lag_s,
         event_type, alarm_level AS lvl, reason, category,
         room_name, alarm_status, operation
  FROM alarm_events
  WHERE device_uid = '$UID_ARG'
    AND triggered_at >= ($START_SQL)
    AND triggered_at <  ($END_SQL)
  ORDER BY triggered_at;"

echo ""
echo "=== payload / evidence（逐条）==="
PSQL -t -A -c "
  SELECT string_agg(
           E'\n##### event_id=' || event_id ||
           '  ' || (triggered_at AT TIME ZONE '$TZ_NAME')::text || E' #####\n' ||
           E'--- payload ---\n'  || jsonb_pretty(payload) || E'\n' ||
           E'--- evidence ---\n' || jsonb_pretty(evidence),
           E'\n')
  FROM alarm_events
  WHERE device_uid = '$UID_ARG'
    AND triggered_at >= ($START_SQL)
    AND triggered_at <  ($END_SQL);"
