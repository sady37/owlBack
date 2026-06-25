#!/bin/bash
# replay-case.sh — 标准化一键回放：清 test 流 → 重启 Xsensor(清空日志) → 喂 case → 日志 cp 回 case 目录。
#
# 用法:
#   ./replay-case.sh <case 目录> [speed]
# 例:
#   ./replay-case.sh ../../doc/cases/case-cabb-0616-17441802 8
#
# 注:speed>1 仅冒烟数据流连通性;confirmMs/dwell/decay 按真实墙钟,加速会压缩时间窗 → fire/confirm 失真。
set -euo pipefail

CASE_DIR="${1:?用法: ./replay-case.sh <case 目录> [speed]}"
SPEED="${2:-8}"

XSENSOR_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OWLBACK_DIR="$(cd "$XSENSOR_DIR/../.." && pwd)"
REPLAY_DIR="$OWLBACK_DIR/tools/replay"
LOG=/home/wisefido/owl/log/Xsensor.log
CASE_ABS="$(cd "$CASE_DIR" && pwd)"

export CONFIG_PATH="${CONFIG_PATH:-$OWLBACK_DIR/wisefido-sensor/config.yaml}"
if [ -f "$OWLBACK_DIR/.env" ]; then set -a; source "$OWLBACK_DIR/.env"; set +a; fi
export REDISCLI_AUTH="${REDIS_PASSWORD:-${REDIS_AUTH:-}}"

echo "▶ [1/6] 清 test:* 流"
redis-cli -h 127.0.0.1 -p 6379 -n 0 --scan --pattern 'test:*' | xargs -r redis-cli -h 127.0.0.1 -p 6379 -n 0 DEL

echo "▶ [2/6] 停旧 Xsensor"
pkill -f '\.bin/xsensor' 2>/dev/null || true
sleep 1

echo "▶ [3/6] 清空日志 + 起 Xsensor"
: > "$LOG"
cd "$XSENSOR_DIR"
go build -o .bin/xsensor ./cmd/xsensor
nohup ./.bin/xsensor >> "$LOG" 2>&1 &
XPID=$!
sleep 3
if ! kill -0 "$XPID" 2>/dev/null; then echo "❌ Xsensor 启动失败,见 $LOG"; tail -20 "$LOG"; exit 1; fi
echo "  Xsensor PID=$XPID"

echo "▶ [4/6] 喂 case @ ${SPEED}x: $CASE_ABS"
cd "$REPLAY_DIR"
go run . --fixture "$CASE_ABS" --streams monitor,event,alarm --stream-prefix test: --speed "$SPEED"

echo "  等待 Xsensor 排空..."
sleep 5
pkill -f '\.bin/xsensor' 2>/dev/null || true

echo "▶ [5/6] cp 日志回 case 目录"
cp "$LOG" "$CASE_ABS/Xsensor.log"
echo "  Xsensor.log → $CASE_ABS/Xsensor.log"

echo "▶ [6/6] 生成 track_event_timeline.md"
# room_pfx + 显示时区从 meta.json 派生（tz 由导出侧烤入；主 radar = case 名 last4，否则首个 Radar）。
read -r ROOM_PFX CASE_TZ < <(python3 - "$CASE_ABS" <<'PY'
import sys, json, os, re
case_dir = sys.argv[1]
m = json.load(open(os.path.join(case_dir, 'meta.json')))
tz = m.get('tz', 'America/Denver')
mo = re.match(r'case-([0-9A-Za-z]{4})-', os.path.basename(case_dir))
last4 = mo.group(1).upper() if mo else None
radars = [d for d in m['devices'] if d['device_type'] == 'Radar']
if not radars:
    sys.exit('meta.json 无 Radar，无法定 room_pfx')
main = next((d for d in radars if d['device_uid'][-4:].upper() == last4), radars[0])
print(':'.join(main['device_addr'].split(':')[:6]), tz)
PY
)
echo "  room_pfx=$ROOM_PFX  tz=$CASE_TZ"
CASE_TZ="$CASE_TZ" python3 "$OWLBACK_DIR/tools/timeline_from_xray.py" "$CASE_ABS" "$ROOM_PFX" "$CASE_ABS/Xsensor.log"
echo "✅ 完成: $CASE_ABS/{Xsensor.log, track_event_timeline.md}"
echo "  fire: grep xsensor_dbn_fire $CASE_ABS/Xsensor.log"
