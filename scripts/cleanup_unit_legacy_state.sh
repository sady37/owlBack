#!/bin/bash
# cleanup_unit_legacy_state.sh — 一次性清理 /80 unit 卡 Redis hash 残留的 bed_state / room_state。
#
# 背景：新 card_display_projector 架构下 /80 unit 卡的 display 由 UnitPicker 跨子 /88 /96
# 合成；/80 自身永不再持 bed_state/room_state（那是子卡的事）。旧代码路径可能在 /80 hash
# 写过这两个字段，FE Overview migrate 到 display 字段前会读到 stale 数据，画错图标。
#
# 本脚本扫所有 card:state:*::/80 keys，HDEL bed_state + room_state；保留 target / alarm_state
# / display / message。同样适用于 /80 public 卡。
#
# Usage:
#   DRY_RUN=1 bash scripts/cleanup_unit_legacy_state.sh   # 只统计不删
#   bash scripts/cleanup_unit_legacy_state.sh             # 真删
#
# Env:
#   REDIS_HOST=127.0.0.1
#   REDIS_PORT=6379
#   REDIS_PASSWORD=TeLunSu-36kr

set -e

REDIS_HOST="${REDIS_HOST:-127.0.0.1}"
REDIS_PORT="${REDIS_PORT:-6379}"
REDIS_PASSWORD="${REDIS_PASSWORD:-TeLunSu-36kr}"
DRY_RUN="${DRY_RUN:-0}"

GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[1;33m'; NC='\033[0m'
info() { echo -e "${YELLOW}• $1${NC}"; }
pass() { echo -e "${GREEN}✓ $1${NC}"; }
warn() { echo -e "${RED}! $1${NC}"; }

REDIS_CLI() { redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" "$@" 2>/dev/null; }

info "Scanning Redis card:state:*::/80 keys..."
mapfile -t KEYS < <(REDIS_CLI --scan --pattern "card:state:*::/80")
TOTAL=${#KEYS[@]}
info "Found $TOTAL /80 unit card hash keys"

if [[ "$TOTAL" -eq 0 ]]; then
  warn "No /80 unit cards in Redis — nothing to do"
  exit 0
fi

CHECKED=0
BED_DEL=0
ROOM_DEL=0
NO_LEGACY=0

for key in "${KEYS[@]}"; do
  CHECKED=$((CHECKED + 1))
  HAS_BED=$(REDIS_CLI HEXISTS "$key" bed_state)
  HAS_ROOM=$(REDIS_CLI HEXISTS "$key" room_state)

  if [[ "$HAS_BED" != "1" && "$HAS_ROOM" != "1" ]]; then
    NO_LEGACY=$((NO_LEGACY + 1))
    continue
  fi

  if [[ "$DRY_RUN" == "1" ]]; then
    [[ "$HAS_BED"  == "1" ]] && echo "  [DRY] would HDEL bed_state  from $key"
    [[ "$HAS_ROOM" == "1" ]] && echo "  [DRY] would HDEL room_state from $key"
  else
    if [[ "$HAS_BED" == "1" ]]; then
      REDIS_CLI HDEL "$key" bed_state >/dev/null
      BED_DEL=$((BED_DEL + 1))
    fi
    if [[ "$HAS_ROOM" == "1" ]]; then
      REDIS_CLI HDEL "$key" room_state >/dev/null
      ROOM_DEL=$((ROOM_DEL + 1))
    fi
  fi
done

echo
echo "================ Summary ================"
echo "/80 keys scanned:           $CHECKED"
echo "Without legacy fields:      $NO_LEGACY"
if [[ "$DRY_RUN" == "1" ]]; then
  warn "DRY_RUN=1 — no changes made. Run without DRY_RUN to actually delete."
else
  echo "bed_state HDEL'd:           $BED_DEL"
  echo "room_state HDEL'd:          $ROOM_DEL"
  pass "Cleanup done."
fi
