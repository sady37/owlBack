#!/usr/bin/env bash
# export —— PG → fixture（统一 window.json 格式），重放/回归的真实数据源。
# **unit 级**：从主设备推 unit(/64)，导出该 unit 窗口内**所有**活跃设备（多 radar + sleepad），
# 因为一个 case（如 cd2b bedroom + 333b bathroom + sleepad）跨整个 unit。
# 与 tools/replay（--fixture 文件模式）+ tools/simulate-make（合成）共用 fixture 契约。
#
# 两种入口：
#   ① case 名（推荐）: ./export.sh case-<uidlast4>-<MMDD>-<startHHMM><endHHMM> [--tz IANA] [--year YYYY]
#        例: ./export.sh case-cd2b-0606-10271037 --tz America/Denver
#   ② 原始 4 参数（UTC ms 直给）: ./export.sh <主device_uid> <start_ms> <end_ms> <case_name>
#
# 输出 doc/cases/<name>/：
#   window.json           unit 内所有 Radar 的 track/heart + event（每条带 device_uid）
#   window_sleepad.json   unit 内所有 Sleepad
#   alarm.json            unit 内设备直发 alarm（producer != sensor），含 evidence（replay --streams alarm）
#   meta.json             {"devices":[{device_uid, device_addr, device_type}]}（窗口内全部活跃设备 → replay 纯文件）
#   room_layout.json      主 radar 的 /128 canvas；其余 radar → room_layout_<last4>.json（供离线 belief_replay）
#
# DB 密码单源 owlBack/.env。

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
[[ -f "$ROOT_DIR/.env" ]] && { set -a; source "$ROOT_DIR/.env"; set +a; }

export PGPASSWORD="${DB_PASSWORD:-postgres}"
PSQL() { psql -h "${DB_HOST:-localhost}" -p "${DB_PORT:-5432}" -U "${DB_USER:-postgres}" -d "${DB_NAME:-owl_v2}" -tA "$@"; }

TZ_ARG="America/Denver"; YEAR_ARG="$(date +%Y)"

# ── 解析入口 ─────────────────────────────────────────────────────────────────
if [[ "${1:-}" =~ ^case-([0-9A-Za-z]{4})-([0-9]{4})-([0-9]{4})([0-9]{4})$ ]]; then
  CASE_NAME="$1"; shift
  LAST4="${BASH_REMATCH[1]}"; MMDD="${BASH_REMATCH[2]}"; SHHMM="${BASH_REMATCH[3]}"; EHHMM="${BASH_REMATCH[4]}"
  while [[ $# -gt 0 ]]; do case "$1" in
      --tz) TZ_ARG="$2"; shift 2;; --year) YEAR_ARG="$2"; shift 2;;
      *) echo "未知参数: $1" >&2; exit 1;; esac
  done
  MAIN_UID=$(PSQL -c "SELECT device_uid FROM devices WHERE device_uid ILIKE '%${LAST4}' LIMIT 1;" | tr -d '[:space:]')
  [[ -z "$MAIN_UID" ]] && { echo "ERROR: 无 device_uid 以 '$LAST4' 结尾" >&2; exit 2; }
  MM="${MMDD:0:2}"; DD="${MMDD:2:2}"
  START_MS=$(( $(TZ="$TZ_ARG" date -d "${YEAR_ARG}-${MM}-${DD} ${SHHMM:0:2}:${SHHMM:2:2}:00" +%s) * 1000 ))
  END_MS=$((   $(TZ="$TZ_ARG" date -d "${YEAR_ARG}-${MM}-${DD} ${EHHMM:0:2}:${EHHMM:2:2}:59" +%s) * 1000 ))
  echo "case 解析: 主uid=$MAIN_UID  tz=$TZ_ARG  ${YEAR_ARG}-${MM}-${DD} ${SHHMM}~${EHHMM}  ms=${START_MS}..${END_MS}"
elif [[ $# -ge 4 ]]; then
  MAIN_UID="$1"; START_MS="$2"; END_MS="$3"; CASE_NAME="$4"
else
  sed -n '2,30p' "$0"; exit 1
fi

# 主设备 addr → unit prefix（权威源 = units 表；一个 /64 是楼层，/80 才是 unit）。
MAIN_ADDR=$(PSQL -c "SELECT host(device_addr) FROM devices WHERE device_uid='$MAIN_UID' LIMIT 1;" | tr -d '[:space:]')
[[ -z "$MAIN_ADDR" ]] && { echo "ERROR: device_uid=$MAIN_UID not found" >&2; exit 2; }
PREFIX=$(PSQL -c "SELECT text(unit_id) FROM units WHERE unit_id >>= '$MAIN_ADDR'::inet ORDER BY masklen(unit_id) DESC LIMIT 1;" | tr -d '[:space:]')
[[ -z "$PREFIX" ]] && { echo "ERROR: $MAIN_ADDR 不属于 units 表任何 unit" >&2; exit 2; }

OUT_DIR="$ROOT_DIR/doc/cases/$CASE_NAME"
mkdir -p "$OUT_DIR"
# 清旧 export 产物（避免上次多/少 radar 的 stale layout 残留）；不动 test_record.txt 等非 export 文件。
rm -f "$OUT_DIR"/room_layout*.json "$OUT_DIR"/window.json "$OUT_DIR"/window_sleepad.json "$OUT_DIR"/alarm.json "$OUT_DIR"/meta.json
echo "unit prefix: $PREFIX  window: $START_MS..$END_MS -> $OUT_DIR"

# ── 窗口内活跃设备（unit 全部，按 type 分） ──────────────────────────────────
ACTIVE=$(PSQL -c "
  SELECT string_agg(DISTINCT d.device_uid || ',' || host(d.device_addr) || ',' ||
    (SELECT m.device_type::text FROM monitor_stream m WHERE m.device_addr=d.device_addr ORDER BY ts DESC LIMIT 1), ';')
  FROM monitor_stream s JOIN devices d ON s.device_addr=d.device_addr
  WHERE s.device_addr << '$PREFIX'::inet AND s.ts BETWEEN to_timestamp($START_MS/1000.0) AND to_timestamp($END_MS/1000.0);")
[[ -z "$ACTIVE" ]] && { echo "ERROR: unit $PREFIX 窗口内无设备数据" >&2; exit 2; }

# meta.json + 各 radar layout（主设备 cd2b → room_layout.json，供离线 belief_replay 认主 radar）
META_DEVS=""
IFS=';' read -ra DEVS <<< "$ACTIVE"
for dev in "${DEVS[@]}"; do
  IFS=',' read -r u a t <<< "$dev"
  META_DEVS="${META_DEVS:+$META_DEVS,}{\"device_uid\":\"$u\",\"device_addr\":\"$a\",\"device_type\":\"$t\"}"
  if [[ "$t" == "Radar" ]]; then
    if [[ "$u" == "$MAIN_UID" ]]; then
      layout_file="room_layout.json"
    else
      layout_file="room_layout_$(echo "$u" | tail -c5 | tr 'A-Z' 'a-z').json"
    fi
    PSQL -c "SELECT canvas::text FROM room_visual_layout WHERE spatial_prefix='$a'::inet LIMIT 1;" > "$OUT_DIR/$layout_file"
    echo "  radar $u → $layout_file ($(wc -c < "$OUT_DIR/$layout_file") bytes)"
  fi
done
printf '{"tz":"%s","devices":[%s]}\n' "$TZ_ARG" "$META_DEVS" > "$OUT_DIR/meta.json"

# ── window.json：unit 内所有 Radar 的 monitor(track/heart) + event ──────────
PSQL -c "
WITH mon AS (
  SELECT s.ts, json_build_object('category', replace(s.stream_type,'radar.',''),
    'device_uid', d.device_uid, 'timestamp', (extract(epoch from s.ts)*1000)::bigint,
    'data_value', s.payload) rec
  FROM monitor_stream s JOIN devices d ON s.device_addr=d.device_addr
  WHERE s.device_addr << '$PREFIX'::inet AND s.device_type='Radar'
    AND s.ts BETWEEN to_timestamp($START_MS/1000.0) AND to_timestamp($END_MS/1000.0)
), evt AS (
  SELECT e.ts, json_build_object('category', e.event_kind,
    'device_uid', d.device_uid, 'timestamp', (extract(epoch from e.ts)*1000)::bigint,
    'data_value', COALESCE(e.payload,'[]'::jsonb)) rec
  FROM event_log e JOIN devices d ON e.device_addr=d.device_addr
  WHERE e.device_addr << '$PREFIX'::inet
    AND e.ts BETWEEN to_timestamp($START_MS/1000.0) AND to_timestamp($END_MS/1000.0)
)
SELECT COALESCE(json_agg(rec ORDER BY ts),'[]'::json)::text
FROM (SELECT ts, rec FROM mon UNION ALL SELECT ts, rec FROM evt) u;
" > "$OUT_DIR/window.json"

# ── window_sleepad.json：unit 内所有 Sleepad ────────────────────────────────
PSQL -c "
SELECT COALESCE(json_agg(json_build_object('category', replace(s.stream_type,'sleepad.',''),
    'device_uid', d.device_uid, 'timestamp', (extract(epoch from s.ts)*1000)::bigint,
    'data_value', s.payload) ORDER BY s.ts),'[]'::json)::text
FROM monitor_stream s JOIN devices d ON s.device_addr=d.device_addr
WHERE s.device_addr << '$PREFIX'::inet AND s.device_type='Sleepad'
  AND s.ts BETWEEN to_timestamp($START_MS/1000.0) AND to_timestamp($END_MS/1000.0);
" > "$OUT_DIR/window_sleepad.json"

# ── alarm.json：unit 内设备直发 alarm（producer != sensor platform-agent），含 evidence ────
# sensor 自产 alarm（producer=fd00:0:fff1::1）不导——replay 中 sensor 应自己重新产出。
# category=event_type；data_value=[payload ∪ {evidence}]（与 replay alarm 信封一致，单元素数组）。
PSQL -c "
SELECT COALESCE(json_agg(json_build_object('category', a.event_type,
    'device_uid', d.device_uid, 'timestamp', (extract(epoch from a.triggered_at)*1000)::bigint,
    'data_value', jsonb_build_array(
       COALESCE(a.payload,'{}'::jsonb)
       || CASE WHEN a.evidence IS NOT NULL AND a.evidence::text NOT IN ('','{}')
               THEN jsonb_build_object('evidence', a.evidence) ELSE '{}'::jsonb END)
    ) ORDER BY a.triggered_at),'[]'::json)::text
FROM alarm_events a JOIN devices d ON a.device_addr=d.device_addr
WHERE a.device_addr << '$PREFIX'::inet AND (a.producer IS NULL OR a.producer != 'fd00:0:fff1::1'::inet)
  AND a.triggered_at BETWEEN to_timestamp($START_MS/1000.0) AND to_timestamp($END_MS/1000.0);
" > "$OUT_DIR/alarm.json"

echo "window.json: $(wc -c < "$OUT_DIR/window.json") bytes / window_sleepad.json: $(wc -c < "$OUT_DIR/window_sleepad.json") bytes / alarm.json: $(wc -c < "$OUT_DIR/alarm.json") bytes"
echo "meta.json: $(echo "$ACTIVE" | tr ';' '\n' | wc -l) 设备"
echo "Done."
