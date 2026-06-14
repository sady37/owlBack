#!/bin/bash
# run-tsensor.sh — 起 Tsensor(DBN 分析/优化靶子)。
#
# Tsensor = wisefido-sensor 源码克隆 + 精简,专供 DBN 网络分析优化:
#   - 订 test:*:stream(生产订 iot:*,流名隔离 → 同 redis 也不串)
#   - cell engine 只读:AreaType 从 owl_v2 grid_snapshot 真实加载,不学习、不写库
#   - 所有对外发送改为落 owl/log/Tsensor.log(不推 redis)
#   - 每 belief tick 无条件输出 tsensor_belief(9 态向量 + 喂入 obs + fall reason + dominant LR)
#
# 与生产隔离三重:① 订 test:* 不碰 iot:* ② 输出全进日志不发 redis ③ 学习/写库循环关停不动 owl_v2。
#
# 用法:
#   ./run-tsensor.sh                       # 起 Tsensor(前台,Ctrl-C 停)
#   # 另开终端喂 case:
#   cd ../replay && go run . --fixture <case 目录> --stream-prefix test: --speed 1
#   # 读分析:
#   tail -f /home/wisefido/owl/log/Tsensor.log | grep tsensor_belief
#
# 注:--speed 1 才保 fire/confirm 时序真实;加速仅数据流冒烟。

set -euo pipefail
TSENSOR_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OWLBACK_DIR="$(cd "$TSENSOR_DIR/../.." && pwd)"

# 复用生产 config(DB owl_v2 + redis 连接);Tsensor 自己订 test:* 流做隔离。
export CONFIG_PATH="${CONFIG_PATH:-$OWLBACK_DIR/wisefido-sensor/config.yaml}"
if [ -f "$OWLBACK_DIR/.env" ]; then
    set -a; source "$OWLBACK_DIR/.env"; set +a
fi

LOG=/home/wisefido/owl/log/Tsensor.log
echo "▶ Tsensor: 订 test:*:stream | config=$CONFIG_PATH"
echo "  日志:$LOG(每 tick tsensor_belief = belief 9 态 + obs + reason)"
echo "  喂 case:cd $OWLBACK_DIR/tools/replay && go run . --fixture <case> --stream-prefix test: --speed 1"

cd "$TSENSOR_DIR"
go build -o .bin/tsensor ./cmd/wisefido-sensor
exec ./.bin/tsensor
