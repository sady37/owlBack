#!/usr/bin/env bash
# 每日轮转 owl/log 下所有服务日志，避免无限 append 撑爆磁盘。
# cp 备份 → 截断原文件（同 inode，不影响 tee -a 进程）。
# 可选删除 RETAIN_DAYS 天前的归档（默认 14）。
#
# crontab 示例（每天 0:05）：
#   5 0 * * * LOG_DIR=/home/wisefido/owl/log /home/wisefido/owl/owlBack/scripts/rotate-owl-logs.sh >>/tmp/rotate-owl-logs-cron.log 2>&1

set -u
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OWL_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
LOG_DIR="${LOG_DIR:-$OWL_ROOT/log}"
RETAIN_DAYS="${RETAIN_DAYS:-14}"

mkdir -p "$LOG_DIR"

# 需要轮转的日志文件列表
LOG_FILES=(
  "wisefido-cardagg.log"
  "wisefido-qinglan.log"
  "wisefido-data.log"
  "wisefido-iot.log"
  "wisefido-sensor.log"
)

d=$(date +%Y%m%d)

for name in "${LOG_FILES[@]}"; do
  GF="$LOG_DIR/$name"
  if [ -f "$GF" ] && [ -s "$GF" ]; then
    bk="$GF.$d"
    if [ -f "$bk" ]; then
      bk="$GF.$d.$(date +%H%M%S)"
    fi
    if cp -a "$GF" "$bk" 2>/dev/null; then
      : >"$GF"
      echo "$(date '+%Y-%m-%d %H:%M:%S') rotated $name -> $(basename "$bk")"
    fi
  fi
done

# 清理旧归档
if [ "${RETAIN_DAYS:-0}" -gt 0 ] 2>/dev/null && command -v find >/dev/null 2>&1; then
  find "$LOG_DIR" -maxdepth 1 -type f \( \
    -name 'wisefido-cardagg.log.*' -o \
    -name 'wisefido-qinglan.log.*' -o \
    -name 'wisefido-data.log.*' -o \
    -name 'wisefido-iot.log.*' -o \
    -name 'wisefido-sensor.log.*' \
    \) -mtime +"$RETAIN_DAYS" -delete 2>/dev/null || true
  echo "$(date '+%Y-%m-%d %H:%M:%S') cleaned archives older than ${RETAIN_DAYS} days"
fi
