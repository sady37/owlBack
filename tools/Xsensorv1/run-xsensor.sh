#!/bin/bash
# run-xsensor.sh — 起 Xsensor（replay-重算 sensor，替换 Tsensor）。
#
# Xsensor = wisefido-sensor 馈送层（track_manager/layout/cell/mirror，从 Tsensor copy）+
#   DBN 四轴 engine.Room 作**顶层唯一决策**（非旧 belief_shadow 旁路子集）：
#   - 订 test:*:stream（生产订 iot:*，流名隔离 → 同 redis 不串）
#   - layout/rooms/devices 启动从 DB 只读加载（同 Tsensor 复用生产 config）
#   - 每 radar 帧 → track_manager 派生 bases → adapter.FrameInput → engine.Room.Tick → fire 落 Xsensor.log
#   - 不推 redis、不写库、无 gate-list/旧 belief：决策全在 DBN 顶层
#
# 用法:
#   ./run-xsensor.sh                       # 起 Xsensor（前台，Ctrl-C 停）
#   # 另开终端喂 case:
#   cd ../replay && go run . --fixture <case 目录> --stream-prefix test: --speed 8
#   # 读 fire:
#   tail -f /home/wisefido/owl/log/Xsensor.log | grep xsensor_dbn_fire
#
# 注:Xsensor 的 stillbox/floor/dwell 计时读消息 Timestamp(虚拟时钟,engine.go ts:=m.Timestamp),
#    与 replay --speed 无关,加速不失真 fire/机制时序;唯一墙钟处是 monitor consumer 6s 新鲜度门,
#    replay 已 rebase t1→now 规避。故 replay 可放心加速。

set -euo pipefail
XSENSOR_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OWLBACK_DIR="$(cd "$XSENSOR_DIR/../.." && pwd)"

# 复用生产 config(DB owl_v2 + redis 连接);Xsensor 自己订 test:* 流做隔离,只读 DB layout,不写。
export CONFIG_PATH="${CONFIG_PATH:-$OWLBACK_DIR/wisefido-sensor/config.yaml}"
if [ -f "$OWLBACK_DIR/.env" ]; then
    set -a; source "$OWLBACK_DIR/.env"; set +a
fi

LOG=/home/wisefido/owl/log/Xsensor.log
echo "▶ Xsensor: 订 test:*:stream | DBN 顶层裁决 | config=$CONFIG_PATH"
echo "  日志:$LOG(xsensor_dbn_fire = DBN 真发; xsensor_belief = 每 tick forensic[Debug])"
echo "  喂 case:cd $OWLBACK_DIR/tools/replay && go run . --fixture <case> --stream-prefix test: --speed 1"

cd "$XSENSOR_DIR"
go build -o .bin/xsensor ./cmd/xsensor
exec ./.bin/xsensor
