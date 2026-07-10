#!/usr/bin/env bash
# case_pipeline.sh —— 单 case 一条龙：export → replay-validate → timeline → 追加 alarm_event payload/evidence。
# 用法: case_pipeline.sh <case-b197-MMDD-startHHMMendHHMM> <IANA-tz> <device_uid> <room_prefix>
# 每 case replay 后**立刻**产 timeline 进 case 目录（共享 log 会被下个 replay 冲，见 [[export_case_script]] 互灌警告）。
set -euo pipefail

CASE="$1"; TZ_="$2"; DEV="$3"; ROOMPFX="$4"; SPEED="${5:-8}"
ROOT="/home/wisefido/owl/owlBack"
cd "$ROOT"
CASE_DIR="doc/cases/$CASE"

# 防孤儿 + 自清：启动前杀残留(上一轮没死透的 xsensor/replay)，退出时(含 timeout/被 kill/Ctrl-C)杀本轮 xsensor。
# 否则被 kill 的后台任务会把 replay-validate+xsensor reparent 到 init，持续写共享 log 互灌下一次 replay。
_selfclean() { pkill -9 -f '\.bin/xsensor' 2>/dev/null || true; pkill -9 -f 'replay-validate' 2>/dev/null || true; }
_selfclean
trap _selfclean EXIT INT TERM

# case 名解析出窗口（MMDD + startHHMM + endHHMM），供 alarm_events 查询（年份取当前上下文 2026）。
YEAR=2026
MMDD="${CASE#case-*-}"; MMDD="${MMDD%%-*}"          # 0703
HHMM="${CASE##*-}"                                   # 00120015
MM="${MMDD:0:2}"; DD="${MMDD:2:2}"
SH="${HHMM:0:2}"; SM="${HHMM:2:2}"; EH="${HHMM:4:2}"; EM="${HHMM:6:2}"
WIN_START="$YEAR-$MM-$DD $SH:$SM:00"
WIN_END="$YEAR-$MM-$DD $EH:$EM:59"

echo "▶ [1/4] export $CASE ($TZ_)"
tools/export/export.sh "$CASE" --tz "$TZ_"

echo "▶ [2/4] replay-validate (新 build xsensor + replay @${SPEED}x)"
tools/replay-validate.sh "$CASE_DIR" "$SPEED"

echo "▶ [3/4] timeline (立刻产出进 case 目录)"
CASE_TZ="$TZ_" python3 tools/timeline_from_xray.py "$CASE_DIR" "$ROOMPFX"

echo "▶ [4/4] 追加 alarm_event payload/evidence 到 timeline 尾"
set -a; source "$ROOT/.env"; set +a
export PGPASSWORD="${DB_PASSWORD:-postgres}"
PSQL() { psql -h "${DB_HOST:-localhost}" -p "${DB_PORT:-5432}" -U "${DB_USER:-postgres}" -d "${DB_NAME:-owl_v2}" -tA "$@"; }
LAST4="${DEV: -4}"
TL="$CASE_DIR/track_event_timeline.md"

EIDS=$(PSQL -c "SELECT event_id FROM alarm_events
  WHERE device_uid ILIKE '%$LAST4'
    AND triggered_at AT TIME ZONE '$TZ_' BETWEEN '$WIN_START' AND '$WIN_END'
  ORDER BY triggered_at;")

if [[ -z "$EIDS" ]]; then
  echo "  (窗口内无 alarm_event，跳过追加)"
else
  {
    echo ""; echo "---"; echo ""
    echo "## alarm_events (production, 窗口 $WIN_START ~ $WIN_END $TZ_) — appended"
    while IFS= read -r eid; do
      [[ -z "$eid" ]] && continue
      echo ""
      PSQL -c "SELECT format('### %s / level %s / %s @ %s %s',
          event_type, alarm_level, reason,
          to_char(triggered_at AT TIME ZONE '$TZ_','YYYY-MM-DD HH24:MI:SS'), '$TZ_')
        FROM alarm_events WHERE event_id='$eid';"
      echo '- event_id: '"$eid"
      echo '#### payload'; echo '```json'
      PSQL -c "SELECT jsonb_pretty(payload::jsonb) FROM alarm_events WHERE event_id='$eid';"
      echo '```'
      echo '#### evidence'; echo '```json'
      PSQL -c "SELECT jsonb_pretty(evidence::jsonb) FROM alarm_events WHERE event_id='$eid';"
      echo '```'
    done <<< "$EIDS"
  } >> "$TL"
  echo "  追加 $(wc -l <<< "$EIDS") 条 alarm_event"
fi
echo "✓ $CASE 完成 → $TL"
