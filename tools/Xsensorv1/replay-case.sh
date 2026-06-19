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

echo "▶ [1/5] 清 test:* 流"
redis-cli -h 127.0.0.1 -p 6379 -n 0 --scan --pattern 'test:*' | xargs -r redis-cli -h 127.0.0.1 -p 6379 -n 0 DEL

echo "▶ [2/5] 停旧 Xsensor"
pkill -f '\.bin/xsensor' 2>/dev/null || true
sleep 1

echo "▶ [3/5] 清空日志 + 起 Xsensor"
: > "$LOG"
cd "$XSENSOR_DIR"
go build -o .bin/xsensor ./cmd/xsensor
nohup ./.bin/xsensor >> "$LOG" 2>&1 &
XPID=$!
sleep 3
if ! kill -0 "$XPID" 2>/dev/null; then echo "❌ Xsensor 启动失败,见 $LOG"; tail -20 "$LOG"; exit 1; fi
echo "  Xsensor PID=$XPID"

echo "▶ [4/5] 喂 case @ ${SPEED}x: $CASE_ABS"
cd "$REPLAY_DIR"
go run . --fixture "$CASE_ABS" --streams monitor,event,alarm --stream-prefix test: --speed "$SPEED"

echo "  等待 Xsensor 排空..."
sleep 5
pkill -f '\.bin/xsensor' 2>/dev/null || true

echo "▶ [5/5] cp 日志回 case 目录"
cp "$LOG" "$CASE_ABS/Xsensor.log"
echo "✅ 完成: $CASE_ABS/Xsensor.log"
echo "  fire: grep xsensor_dbn_fire $CASE_ABS/Xsensor.log"
