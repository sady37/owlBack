#!/bin/bash
# cleanup_stale_bed_state.sh — 一次性扫描 card:state:* Redis 哈希，
# 对非 bed-capable 卡（card_type != 'active_bed' AND 无 Sleepad 设备）HDEL 掉 bed_state 字段。
#
# 背景：cardagg state_service.go:622 InitCardRoomAndBathroomState 之前对所有卡
# 写默认 bed_state{BedStatus=8, StartTime=now}，永不被真实 InBed/LeftBed 事件覆盖。
# FE Overview.vue:2131 修复已识别 bed_status=8 为"无信息"不显示 OOB；本脚本清掉残留数据。
#
# Usage:
#   DRY_RUN=1 bash scripts/cleanup_stale_bed_state.sh      # 只统计不删
#   bash scripts/cleanup_stale_bed_state.sh                # 真删
#
# Env:
#   REDIS_HOST=127.0.0.1
#   REDIS_PORT=6379
#   REDIS_PASSWORD=TeLunSu-36kr
#   PG_HOST=127.0.0.1
#   PG_USER=postgres
#   PG_PASSWORD=postgres
#   PG_DB=owl_v2

set -e

REDIS_HOST="${REDIS_HOST:-127.0.0.1}"
REDIS_PORT="${REDIS_PORT:-6379}"
REDIS_PASSWORD="${REDIS_PASSWORD:-TeLunSu-36kr}"
PG_HOST="${PG_HOST:-127.0.0.1}"
PG_USER="${PG_USER:-postgres}"
PG_PASSWORD="${PG_PASSWORD:-postgres}"
PG_DB="${PG_DB:-owl_v2}"
DRY_RUN="${DRY_RUN:-0}"

GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[1;33m'; NC='\033[0m'
info() { echo -e "${YELLOW}• $1${NC}"; }
pass() { echo -e "${GREEN}✓ $1${NC}"; }
warn() { echo -e "${RED}! $1${NC}"; }

REDIS_CLI() { redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" "$@" 2>/dev/null; }
PSQL()       { PGPASSWORD="$PG_PASSWORD" psql -h "$PG_HOST" -U "$PG_USER" -d "$PG_DB" -t -A "$@"; }

# 1. 扫所有 card:state:* keys
info "Scanning Redis card:state:* keys..."
mapfile -t KEYS < <(REDIS_CLI --scan --pattern "card:state:*")
TOTAL=${#KEYS[@]}
info "Found $TOTAL card:state keys"

if [[ "$TOTAL" -eq 0 ]]; then
  warn "No card:state:* keys found — nothing to do"
  exit 0
fi

# 2. 一次性查 DB 拿所有"非 bed-capable"卡的 spatial_prefix 列表
#    bed-capable = card_type='active_bed' OR
#                  存在一个 Sleepad 设备 IPv6 在本卡 spatial_prefix 内 AND 没有更深的 /96 active_bed 卡拥有它
#    (与 cardagg fillDevicesV3 LPM 归属规则一致：sleepad 默认归 /96 bed card，无 /96 时上推 /80 unit)
info "Querying non-bed-capable cards from DB..."
NON_BED_PREFIXES=$(PSQL -c "
  SELECT host(c.spatial_prefix)||'/'||masklen(c.spatial_prefix)
  FROM cards c
  WHERE c.card_type <> 'active_bed'
    AND NOT EXISTS (
      SELECT 1
      FROM device_factory_meta dfm
      JOIN devices d ON d.device_id = dfm.device_id
      WHERE d.device_ipv6 <<= c.spatial_prefix
        AND dfm.device_type = 'Sleepad'
        AND NOT EXISTS (
          SELECT 1 FROM cards c2
          WHERE c2.spatial_prefix = network(set_masklen(d.device_ipv6, 96))
            AND c2.card_type = 'active_bed'
            AND masklen(c2.spatial_prefix) > masklen(c.spatial_prefix)
        )
    )
")
NON_BED_COUNT=$(echo "$NON_BED_PREFIXES" | grep -c '^' || true)
info "Found $NON_BED_COUNT non-bed-capable cards in DB"

# 3. 遍历每个 Redis key，若 cardID ∈ non-bed-capable 集合 + bed_state 字段存在 → HDEL
CHECKED=0
TO_DELETE=0
DELETED=0
NOT_BED_BUT_NO_FIELD=0

declare -A NON_BED_SET
while IFS= read -r line; do
  [[ -n "$line" ]] && NON_BED_SET["$line"]=1
done <<<"$NON_BED_PREFIXES"

for key in "${KEYS[@]}"; do
  CHECKED=$((CHECKED + 1))
  # key 格式: card:state:fd00:0:3:111:3::/80
  CARDID="${key#card:state:}"
  if [[ -z "${NON_BED_SET[$CARDID]:-}" ]]; then
    continue  # bed-capable，跳过
  fi
  # 该 card 非 bed-capable — 检查 hash 中是否有 bed_state 字段
  HAS_FIELD=$(REDIS_CLI HEXISTS "$key" bed_state)
  if [[ "$HAS_FIELD" != "1" ]]; then
    NOT_BED_BUT_NO_FIELD=$((NOT_BED_BUT_NO_FIELD + 1))
    continue
  fi
  TO_DELETE=$((TO_DELETE + 1))
  if [[ "$DRY_RUN" == "1" ]]; then
    echo "  [DRY] would HDEL bed_state from $key"
  else
    REDIS_CLI HDEL "$key" bed_state >/dev/null
    DELETED=$((DELETED + 1))
  fi
done

echo
echo "================ Summary ================"
echo "Keys scanned:                  $CHECKED"
echo "Non-bed-capable cards (DB):    $NON_BED_COUNT"
echo "Non-bed cards w/o bed_state:   $NOT_BED_BUT_NO_FIELD"
if [[ "$DRY_RUN" == "1" ]]; then
  echo "Would HDEL bed_state from:     $TO_DELETE"
  warn "DRY_RUN=1 — no changes made. Run without DRY_RUN to actually delete."
else
  echo "HDEL bed_state executed on:    $DELETED"
  pass "Cleanup done."
fi
