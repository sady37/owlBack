#!/usr/bin/env bash
# belief shadow vs gate-list 实时对比监控（§9 3b）。
# 把 belief shadow 的 belief_shadow_fall（sensor journal）与 gate-list 的 alarm_events Fall
# 拉到同一时间窗对照，flag 分歧。log-only，纯只读，不改任何状态。
#
# 用法: ./belief_shadow_monitor.sh [窗口分钟，默认 20]
set -euo pipefail
WIN_MIN="${1:-20}"
SINCE="$WIN_MIN min ago"
export PGPASSWORD="${DB_PASSWORD:-postgres}"
PSQL() { psql -h "${DB_HOST:-localhost}" -p "${DB_PORT:-5432}" -U "${DB_USER:-postgres}" -d "${DB_NAME:-owl_v2}" -tA "$@"; }

echo "===== belief shadow vs gate-list  [窗口: 近 ${WIN_MIN}min]  $(date '+%H:%M:%S') ====="

# A) belief shadow fire（journal）
SHADOW=$(sudo journalctl -u owlback.sensor --since "$SINCE" --no-pager 2>/dev/null | grep "belief_shadow_fall" || true)
SHADOW_N=$(printf '%s' "$SHADOW" | grep -c "belief_shadow_fall" || true)
echo "[belief shadow] fire 条数: ${SHADOW_N:-0}"
if [[ -n "$SHADOW" ]]; then
  printf '%s\n' "$SHADOW" | grep -oE '"room_id":"[^"]*".*"p_fallen":[0-9.]*' | sed 's/^/  /' | head -10
fi

# B) gate-list Fall（DB），belief 相关类（lost_track / person_silent）+ firmware 计数
echo "[gate-list] 近 ${WIN_MIN}min Fall 分类:"
PSQL -c "SELECT '  '||reason||' x'||cnt FROM (SELECT COALESCE(payload->>'reason','?') reason, count(*) cnt FROM alarm_events WHERE event_type='Fall' AND triggered_at > now()-interval '${WIN_MIN} minutes' GROUP BY 1) s ORDER BY cnt DESC;" 2>/dev/null || echo "  (db err)"
echo "[gate-list] lost_track / person_silent 明细（belief 应对账这些）:"
PSQL -c "SELECT '  '||to_char(triggered_at AT TIME ZONE 'Asia/Shanghai','HH24:MI')||' '||unit_name||'/'||room_name||' '||alarm_status||' reason='||COALESCE(payload->>'reason','?')||' still_box='||COALESCE(payload->'evidence'->>'frozen_start_ms','-') FROM alarm_events WHERE event_type='Fall' AND payload->>'reason' IN ('lost_track','bedroom_person_silent') AND triggered_at > now()-interval '${WIN_MIN} minutes' ORDER BY triggered_at DESC;" 2>/dev/null || echo "  (db err)"

# C) 分歧摘要
GATE_RELEVANT=$(PSQL -c "SELECT count(*) FROM alarm_events WHERE event_type='Fall' AND payload->>'reason' IN ('lost_track','bedroom_person_silent') AND triggered_at > now()-interval '${WIN_MIN} minutes';" 2>/dev/null | tr -d '[:space:]')
echo "----- 分歧 -----"
echo "  belief shadow fire=${SHADOW_N:-0} ; gate-list lost/silent fire=${GATE_RELEVANT:-0}"
if [[ "${SHADOW_N:-0}" == "0" && "${GATE_RELEVANT:-0}" == "0" ]]; then
  echo "  ✓ 两边静默一致（止血+belief 都不报）"
elif [[ "${SHADOW_N:-0}" == "0" && "${GATE_RELEVANT:-0}" != "0" ]]; then
  echo "  ⚠ gate-list 报了 belief 没报 → belief 疑似压掉 gate-list FP（需 triage 是真压对还是漏）"
elif [[ "${SHADOW_N:-0}" != "0" && "${GATE_RELEVANT:-0}" == "0" ]]; then
  echo "  ⚠ belief 报了 gate-list 没报 → belief-only（需 triage 是抓到 gate 漏的还是 belief FP）"
else
  echo "  ~ 两边都报（需按 room/time 逐条核对是否同一事件）"
fi
