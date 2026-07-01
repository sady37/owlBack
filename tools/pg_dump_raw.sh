#!/usr/bin/env bash
# pg_dump_raw.sh — 按 device_uid + 本地起止时间，从 PG 导出 monitor_stream + event_log
# 完整原始记录（整行所有列，to_jsonb 一列不丢），两表合并按 ts 升序。
# 按 device_uid 过滤（免疫 addr rebind）。
#
# 用法: ./pg_dump_raw.sh <device_uid> '<本地起>' '<本地止>' [tz=America/Denver]
#   例: ./pg_dump_raw.sh 9923003AB197 '2026-06-30 08:02:00' '2026-06-30 08:19:00'
#
# 起止时间按 tz 解释。输出到 stdout：每行一条记录 = 该行整行 JSON
#   注入两个便利字段：_src(MON/EVT) 与 _local(本地时间)；其余键 = 表原始列名/值（含 payload 全文）。
# DB 密码单源 owlBack/.env。
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
[[ -f "$ROOT_DIR/.env" ]] && { set -a; source "$ROOT_DIR/.env"; set +a; }
export PGPASSWORD="${DB_PASSWORD:-postgres}"
PSQL() { psql -h "${DB_HOST:-localhost}" -p "${DB_PORT:-5432}" -U "${DB_USER:-postgres}" -d "${DB_NAME:-owl_v2}" -tA "$@"; }

UID_ARG="${1:?用法: pg_dump_raw.sh <device_uid> '<本地起>' '<本地止>' [tz]}"
START_LOCAL="${2:?缺本地起时间，如 '2026-06-30 08:02:00'}"
END_LOCAL="${3:?缺本地止时间，如 '2026-06-30 08:19:00'}"
TZ_ARG="${4:-America/Denver}"

START_MS=$(( $(TZ="$TZ_ARG" date -d "$START_LOCAL" +%s) * 1000 ))
END_MS=$((   $(TZ="$TZ_ARG" date -d "$END_LOCAL"   +%s) * 1000 ))
echo "# uid=$UID_ARG tz=$TZ_ARG  $START_LOCAL .. $END_LOCAL  (ms ${START_MS}..${END_MS})" >&2

# to_jsonb(t.*) 保整行所有列；jsonb_build_object 注入 _src/_local 后用 || 合并；ORDER BY 真实 ts。
PSQL -c "
WITH mon AS (
  SELECT ts,
    jsonb_build_object('_src','MON','_local', to_char(ts AT TIME ZONE '$TZ_ARG','YYYY-MM-DD HH24:MI:SS.MS'))
      || to_jsonb(m.*) AS row
  FROM monitor_stream m
  WHERE device_uid='$UID_ARG'
    AND ts BETWEEN to_timestamp($START_MS/1000.0) AND to_timestamp($END_MS/1000.0)
), evt AS (
  SELECT ts,
    jsonb_build_object('_src','EVT','_local', to_char(ts AT TIME ZONE '$TZ_ARG','YYYY-MM-DD HH24:MI:SS.MS'))
      || to_jsonb(e.*) AS row
  FROM event_log e
  WHERE device_uid='$UID_ARG'
    AND ts BETWEEN to_timestamp($START_MS/1000.0) AND to_timestamp($END_MS/1000.0)
)
SELECT row::text FROM (SELECT ts, row FROM mon UNION ALL SELECT ts, row FROM evt) u
ORDER BY ts;
"
