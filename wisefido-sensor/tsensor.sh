#!/bin/bash
# tsensor.sh — Tsensor: 生产 wisefido-sensor 同一份二进制,跑在独立 redis 库号 +
# 逐 tick 全状态 verbose,供回放隔离演算/查原因。零 fork:只灌隔离 env,不改代码。
#
#   隔离靠 REDIS_DB(键空间独立)而非改 stream 名 → Tsensor 读写都在 DB N,
#   生产在 DB 0,物理上不可能互相污染。配合 tools/replay --redis-db <N> 推录像进同库。
#
# 用法:
#   ./tsensor.sh                 # REDIS_DB=1, verbose on, DBN_MODE=2
#   REDIS_DB=2 ./tsensor.sh      # 换隔离库
#   DBN_MODE=1 ./tsensor.sh      # 换档
#
# 安全闸:REDIS_DB=0 = 生产库,拒绝启动(防误把 Tsensor 指向生产)。

set -euo pipefail

OWLBACK_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# 复用生产同套 .env(DB / redis addr / identity 等),再用下方 override 覆盖隔离项。
if [ -f "$OWLBACK_DIR/load_env.sh" ]; then
    # shellcheck disable=SC1091
    source "$OWLBACK_DIR/load_env.sh"
fi

export REDIS_DB="${REDIS_DB:-1}"
if [ "$REDIS_DB" = "0" ]; then
    echo "❌ REDIS_DB=0 是生产库,Tsensor 拒绝启动(改用 1+)。" >&2
    exit 1
fi
export TSENSOR_VERBOSE="${TSENSOR_VERBOSE:-1}"
export DBN_MODE="${DBN_MODE:-2}"

LOG="$OWLBACK_DIR/wisefido-sensor/.bin/tsensor.log"
mkdir -p "$(dirname "$LOG")"

echo "▶ Tsensor: REDIS_DB=$REDIS_DB  TSENSOR_VERBOSE=$TSENSOR_VERBOSE  DBN_MODE=$DBN_MODE"
echo "  隔离:键空间 DB=$REDIS_DB(生产=0),stream 名不变;tools/replay --redis-db $REDIS_DB 推录像进同库"
echo "  日志:$LOG(逐 tick tsensor_tick = belief 9 态向量 + p_fallen + bed 态 + 每 track realness + τ*)"

cd "$OWLBACK_DIR/wisefido-sensor"
go build -o .bin/wisefido-sensor ./cmd/wisefido-sensor
exec ./.bin/wisefido-sensor 2>&1 | tee "$LOG"
