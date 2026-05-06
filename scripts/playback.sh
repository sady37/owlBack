#!/usr/bin/env bash
# 拉指定雷达的 room playback HTML（自动启动 roomengine-api dev server，60s timeout）。
#
# 用法:
#   scripts/playback.sh <device_uid> <start> <end> [snap_min]
#
# 例:
#   scripts/playback.sh 25A859B8333B "2026-05-05T01:00:00-07:00" "2026-05-05T08:00:00-07:00"
#   scripts/playback.sh 25A859B8333B "2026-05-05 01:00" "2026-05-05 08:00" 30
#
# 输出: /tmp/playback_<uid>_<start_short>.html  +  浏览器打开命令
set -euo pipefail

UID_ARG="${1:?usage: $0 <device_uid> <start> <end> [snap_min]}"
START="${2:?missing start (RFC3339 or 'YYYY-MM-DD HH:MM')}"
END="${3:?missing end}"
SNAP_MIN="${4:-30}"
PORT="${PORT:-7788}"
OWLBACK="${OWLBACK:-/home/wisefido/owl/owlBack}"
LOG_DIR="${LOG_DIR:-/home/wisefido/owl/log}"
OUT_DIR="${OUT_DIR:-${OWLBACK}/out}"
mkdir -p "$OUT_DIR"

# 时间格式补全：'YYYY-MM-DD HH:MM' → 当前时区 RFC3339
to_rfc3339() {
  if [[ "$1" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}T ]]; then echo "$1"; return; fi
  date -d "$1" --rfc-3339=seconds | tr ' ' 'T'
}
START_RFC=$(to_rfc3339 "$START")
END_RFC=$(to_rfc3339 "$END")

OUT="${OUT_DIR}/playback_${UID_ARG}_$(date -d "$START_RFC" +%Y%m%d-%H%M).html"

# 1) 检查 dev server，没起就起
if ! curl -sS --max-time 2 "http://127.0.0.1:${PORT}/api/health" >/dev/null 2>&1; then
  echo "[playback] starting roomengine-api on :${PORT} ..."
  cd "$OWLBACK"
  set -a; [[ -f .env ]] && source .env; set +a
  export DB_HOST="${DB_HOST:-127.0.0.1}" DB_PORT="${DB_PORT:-5432}" \
         DB_USER="${DB_USER:-postgres}" DB_PASSWORD="${DB_PASSWORD:-postgres}" \
         DB_NAME="${DB_NAME:-owlrd}" DB_SSLMODE="${DB_SSLMODE:-disable}"
  cd wisefido-sensor
  nohup go run ./cmd/roomengine-api --listen ":${PORT}" \
    > "${LOG_DIR}/roomengine-api.log" 2>&1 & disown
  for i in {1..30}; do
    sleep 1
    curl -sS --max-time 1 "http://127.0.0.1:${PORT}/api/health" >/dev/null 2>&1 && break
  done
  if ! curl -sS --max-time 2 "http://127.0.0.1:${PORT}/api/health" >/dev/null 2>&1; then
    echo "[playback] api failed to start, see ${LOG_DIR}/roomengine-api.log" >&2
    exit 1
  fi
fi

# 2) 拉 HTML
echo "[playback] uid=${UID_ARG}  ${START_RFC} → ${END_RFC}  snap_min=${SNAP_MIN}"
URL="http://127.0.0.1:${PORT}/api/playback?uid=${UID_ARG}&start=${START_RFC}&end=${END_RFC}&snap_min=${SNAP_MIN}&format=html"
curl -sSf "$URL" --max-time 60 -o "$OUT"

# 3) 颜色统计（一眼看 InRoom 比例 / 是否有学习色）
SIZE=$(wc -c <"$OUT")
echo "[playback] saved ${OUT}  (${SIZE} bytes)"

python3 - "$OUT" <<'PY'
import re, sys
from collections import Counter
with open(sys.argv[1]) as f: html = f.read()
fills = re.findall(r'fill=\\"(#[0-9a-fA-F]{6})\\"', html)
c = Counter(fills); total = sum(c.values())
if total == 0:
    print("  (no fills found — empty playback or window has no monitor data)"); sys.exit()
out  = c.get('#2a2a2a', 0)
walk = c.get('#ffffff', 0)
ent  = c.get('#44cc66', 0)
bed  = c.get('#4488dd', 0)
sit  = c.get('#ff9933', 0)
print(f"  cells x snaps = {total}")
print(f"  !InRoom (Out)   {out:>7}  {out*100//total:>3}%")
print(f"  InRoom          {total-out:>7}  {100-out*100//total:>3}%")
print(f"  Walk (#ffffff)  {walk:>7}   ← cell_learning 累积出的活动区")
print(f"  Enter (#44cc66) {ent:>7}")
print(f"  Bed (#4488dd)   {bed:>7}")
print(f"  Sit (#ff9933)   {sit:>7}")
PY

echo ""
echo "[playback] open in browser:"
echo "  xdg-open ${OUT}"
echo "  file://${OUT}"
echo "[playback] list all:  ls -lh ${OUT_DIR}/playback_*.html"
