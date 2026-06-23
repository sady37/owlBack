#!/usr/bin/env bash
# cut-to-cutover.sh — 原子切生产：旧 sensor → cutover sensor(DBN_MODE=0 shadow)。
# 🔴 高危：替换线上生产 fall-detection。须架构师显式批准（A·R21）。必带 --go 才执行。
# A·R21.2 强制 DBN_MODE=0 起步（生产含双雷达房，守卫①②未接，1/2=裸跑发真 alarm 风险）。
set -euo pipefail
OWLBACK=/home/wisefido/owl/owlBack
SDIR=$OWLBACK/wisefido-sensor
SVC=owlback.sensor.service

if [[ "${1:-}" != "--go" ]]; then
  echo "🔴 高危操作：替换线上生产 sensor 为 cutover 二进制(DBN_MODE=0 shadow)。"
  echo "   须架构师显式批准。确认后带 --go 执行。"
  echo "   当前 .env DBN_MODE=$(grep -E '^DBN_MODE=' "$OWLBACK/.env" | cut -d= -f2)（本脚本会改 0）。"
  exit 2
fi

echo "▶ [1/6] 备份 .env + 旧 binary"
cp -f "$OWLBACK/.env" "$OWLBACK/.env.precut-backup"
[[ -f "$SDIR/.bin/wisefido-sensor.prod-backup" ]] || cp -f "$SDIR/.bin/wisefido-sensor" "$SDIR/.bin/wisefido-sensor.prod-backup"

echo "▶ [2/6] DBN_MODE → 0（A·R21.2 强制 shadow 起步）"
sed -i 's/^DBN_MODE=.*/DBN_MODE=0/' "$OWLBACK/.env"
grep -E '^DBN_MODE=' "$OWLBACK/.env"

echo "▶ [3/6] 换 cutover 二进制"
cp -f "$SDIR/.bin/wisefido-sensor.cutover" "$SDIR/.bin/wisefido-sensor"

echo "▶ [4/6] systemctl restart（原子，防双跑；systemd 管理）"
systemctl restart "$SVC"
sleep 4

echo "▶ [5/6] 启动自检（fail-safe）"
systemctl is-active "$SVC" | grep -q '^active$' || { echo "❌ 服务未 active，回滚！"; "$OWLBACK/scripts/stagecut/rollback.sh" --go; exit 1; }
journalctl -u "$SVC" --since "1 min ago" -o cat 2>/dev/null | grep -q "DBN router wired" || echo "⚠️ 未见 'DBN router wired'（确认 OnRoomFrame 接通）"
echo "  服务 active；DBN_MODE=0 shadow（固件 floor 保底，DBN 不发真 alarm）"

echo "▶ [6/6] dump dbn_xray（实时生产流 DBN 裁决，验机制规则#3）"
echo "  journalctl -u $SVC -f -o cat | grep dbn_xray"
echo "✅ 切生产完成（DBN_MODE=0 shadow）。回滚：scripts/stagecut/rollback.sh --go"
