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

# 调用者意图先存下:.env 里也有 REDIS_DB/TSENSOR_VERBOSE/DBN_MODE,source 会覆盖,
# 故隔离项必须在 source 后强制写回调用者意图(默认 1),否则被 .env 的生产值(DB=0)盖掉。
_REQ_DB="${REDIS_DB:-1}"
_REQ_VERBOSE="${TSENSOR_VERBOSE:-1}"
_REQ_DBN="${DBN_MODE:-2}"

# 复用生产同套 .env(DB / redis addr / identity 等)。
if [ -f "$OWLBACK_DIR/load_env.sh" ]; then
    # shellcheck disable=SC1091
    source "$OWLBACK_DIR/load_env.sh"
fi

export REDIS_DB="$_REQ_DB"
if [ "$REDIS_DB" = "0" ]; then
    echo "❌ REDIS_DB=0 是生产库,Tsensor 拒绝启动(改用 1+)。" >&2
    exit 1
fi
export TSENSOR_VERBOSE="$_REQ_VERBOSE"
export DBN_MODE="$_REQ_DBN"

LOG="$OWLBACK_DIR/wisefido-sensor/.bin/tsensor.log"
mkdir -p "$(dirname "$LOG")"

echo "▶ Tsensor: REDIS_DB=$REDIS_DB  TSENSOR_VERBOSE=$TSENSOR_VERBOSE  DBN_MODE=$DBN_MODE"
echo "  隔离:键空间 DB=$REDIS_DB(生产=0),stream 名不变;tools/replay --redis-db $REDIS_DB 推录像进同库"
echo "  日志:$LOG(逐 tick tsensor_tick = belief 9 态向量 + p_fallen + bed 态 + 每 track realness + τ*)"

cd "$OWLBACK_DIR/wisefido-sensor"
go build -o .bin/wisefido-sensor ./cmd/wisefido-sensor
exec ./.bin/wisefido-sensor 2>&1 | tee "$LOG"
