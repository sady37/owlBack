#!/usr/bin/env bash
# 每日轮转 owl/log 下 sleepace 相关日志，避免无限 append 撑爆磁盘。
# 1) sleepace/sleepace.sh rotate — Java server.out/err 聚合到 sleepace.out / sleepace.err
# 2) wisefido-sleepace.log — start-owlback 里 go run 重定向；同 inode 上 cp 后备份再截断
# 3) 可选删除 RETAIN_DAYS 天前的归档（默认 14）
#
# crontab 示例（每天 0:10，LOG_DIR 与 start-owlback 一致）：
#   10 0 * * * LOG_DIR=/home/wisefido/owl/log /home/wisefido/owl/owlBack/scripts/rotate-owl-sleepace-logs.sh >>/tmp/rotate-owl-sleepace-cron.log 2>&1

set -u
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OWL_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
LOG_DIR="${LOG_DIR:-$OWL_ROOT/log}"
RETAIN_DAYS="${RETAIN_DAYS:-14}"

mkdir -p "$LOG_DIR"

if [ -x "$OWL_ROOT/sleepace/sleepace.sh" ]; then
  "$OWL_ROOT/sleepace/sleepace.sh" rotate || true
fi

GF="$LOG_DIR/wisefido-sleepace.log"
if [ -f "$GF" ] && [ -s "$GF" ]; then
  d=$(date +%Y%m%d)
  bk="$GF.$d"
  if [ -f "$bk" ]; then
    bk="$GF.$d.$(date +%H%M%S)"
  fi
  if cp -a "$GF" "$bk" 2>/dev/null; then
    : >"$GF"
    echo "rotated $GF -> $bk (truncated in place)"
  fi
fi

if [ "${RETAIN_DAYS:-0}" -gt 0 ] 2>/dev/null && command -v find >/dev/null 2>&1; then
  find "$LOG_DIR" -maxdepth 1 -type f \( \
    -name 'sleepace.out.*' -o \
    -name 'sleepace.err.*' -o \
    -name 'wisefido-sleepace.log.*' \
    \) -mtime +"$RETAIN_DAYS" -delete 2>/dev/null || true
fi
