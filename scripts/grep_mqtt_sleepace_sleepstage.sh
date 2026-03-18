#!/usr/bin/env bash
# 从 mqtt-sleepace.log 中筛 BM87224601903 → device_code 1ua3erivl9pv1 的 sleepStage 及 realtime，用于查 awake 来源
# 用法: ./grep_mqtt_sleepace_sleepstage.sh [mqtt-sleepace.log]
# sleepStage 报文含 "dataKey":"sleepStage" 与 "sleepStage":1(清醒)/2(浅睡)/4(深睡)/8(未知)
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG="${1:-${SCRIPT_DIR}/mqtt-sleepace.log}"
DEV="1ua3erivl9pv1"
if [[ ! -f "$LOG" ]]; then
  echo "No log: $LOG. Run ./mqtt_sub_sleepace_to_log.sh first."
  exit 1
fi
echo "=== sleepStage 报文 (deviceId=$DEV) ==="
grep -E "sleepStage|sleepstage" "$LOG" | grep "$DEV" || true
echo ""
echo "=== realtime 报文 (deviceId=$DEV, 含 bedStatus) 最近 20 条 ==="
grep "realtime" "$LOG" | grep "$DEV" | tail -20 || true
