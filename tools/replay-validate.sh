#!/usr/bin/env bash
# replay-validate.sh — 标准化 replay 验证一条龙:清 test:* → 重启 xsensor(新 build) → replay → dump 目标房 xray 时间线 + fire。
# 用法: ./replay-validate.sh <fixture-dir> [speed=1] [room-grep]
#   fixture-dir = 含 meta.json+window.json 的目录;room-grep = 只 dump 含此子串的房(如 411:1:200);省略=dump 所有 fire。
set -euo pipefail

FIXTURE="$(realpath "${1:?用法: replay-validate.sh <fixture-dir> [speed] [room-grep]}")" # 绝对路径：step3 cd 进 replay 后相对路径会失效
SPEED="${2:-1}"
ROOM="${3:-}"
XDIR=/home/wisefido/owl/owlBack/tools/Xsensorv1
RDIR=/home/wisefido/owl/owlBack/tools/replay
LOG=/home/wisefido/owl/log/Xsensor.log
RPW="${REDIS_PASSWORD:-TeLunSu-36kr}"

echo "▶ [1/4] 停 xsensor + 清 test:*"
pkill -x xsensor 2>/dev/null || true; sleep 1
redis-cli -a "$RPW" -n 0 --scan --pattern 'test:*' 2>/dev/null | xargs -r redis-cli -a "$RPW" -n 0 del >/dev/null 2>&1 || true

echo "▶ [2/4] 重启 xsensor(新 build) + 清日志"
: > "$LOG"
( cd "$XDIR" && setsid ./run-xsensor.sh </dev/null >/home/wisefido/owl/log/xsensor-run.out 2>&1 & ) # setsid+</dev/null：彻底脱离，否则后台 xsensor 继承外层管道 fd 致调用方 tail 永等不到 EOF
for i in $(seq 1 20); do sleep 1; pgrep -x xsensor >/dev/null && grep -q "units built" "$LOG" 2>/dev/null && break; done
pgrep -x xsensor >/dev/null || { echo "  ✗ xsensor 起不来"; exit 1; }
echo "  ✓ xsensor 起 ($(grep -c 'room registered' "$LOG") 房)"

echo "▶ [3/4] replay $FIXTURE @${SPEED}x"
( cd "$RDIR" && go run . --fixture "$FIXTURE" --stream-prefix test: --speed "$SPEED" ) | tail -1
sleep 2

echo "▶ [4/4] 分析 xray"
python3 - "$LOG" "$ROOM" <<'PY'
import sys, json
log, room = sys.argv[1], sys.argv[2]
rows = []
for l in open(log):
    if '"xsensor_xray"' not in l: continue
    try: d = json.loads(l)
    except: continue
    if room and room not in d.get('room',''): continue
    rows.append(d)
# fire 汇总(全房)
fires = [d for d in rows if d.get('fire')]
print(f"  xray 行数(过滤后) {len(rows)} | fire 帧数 {len(fires)}")
if fires:
    print("  🔥 FIRE:")
    for d in fires[:20]:
        print(f"    ts {d['ts']} room {d['room'][-12:]} band {d['band']!r} pF {round(d['p_fallen'],3)} top_s {d['top_s']}")
# 目标房状态变化时间线
if room:
    prev = None
    print(f"  时间线({room}) — 仅状态变化:")
    for d in rows:
        key = (d['present_cnt'], round(d['p_fallen'],2), d['top_s'], d['band'], d.get('ud_active'), d.get('unit_has_track'))
        if key != prev:
            dbn = [{'lid':t['lid'],'pR':round(t['p_real'],2),'refl':t.get('is_refl'),'pF':round(t['p_fallen'],2)} for t in d.get('dbn',[])]
            print(f"    ts {d['ts']} | pres {d['present_cnt']} raw {d['raw_tracks']} | pF {round(d['p_fallen'],3)} top_s {d['top_s']} band {d['band']!r} fire {d['fire']} | ud {d.get('ud_active')} rem_min {round(d.get('ud_remain_ms',0)/60000,1)} nbr {d.get('has_neighbor')} unit_trk {d.get('unit_has_track')} | dbn {dbn}")
            prev = key
PY
echo "▶ 完成"
