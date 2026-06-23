#!/usr/bin/env bash
# cut-to-cutover.sh — 原子切生产：旧 sensor → cutover sensor(DBN_MODE=1，架构师 A·R22 拍定)。
# 🔴 高危：替换线上生产 fall-detection，发真 DBN alarm。须架构师/用户显式批准。必带 --go 才执行。
# A·R22.1 风险在案：双雷达房守卫①②未接 → FP 误报无兜底(打扰护士)；固件 floor 仅兜 FN 漏报。
set -euo pipefail
OWLBACK=/home/wisefido/owl/owlBack
SDIR=$OWLBACK/wisefido-sensor
SVC=owlback.sensor.service

if [[ "${1:-}" != "--go" ]]; then
  echo "🔴 高危操作：替换线上生产 sensor 为 cutover 二进制(DBN_MODE=1，架构师 A·R22 拍定，发真 DBN alarm)。"
  echo "   须架构师/用户显式批准。确认后带 --go 执行。"
  echo "   当前 .env DBN_MODE=$(grep -E '^DBN_MODE=' "$OWLBACK/.env" | cut -d= -f2)（本脚本会设 1）。"
  echo "   ⚠️ 双雷达房守卫①②未接→FP 误报无兜底(打扰护士)；固件 floor 仅兜 FN 漏报。出问题：rollback.sh --go 或 .env DBN_MODE=0。"
  exit 2
fi

echo "▶ [1/6] 备份 .env + 旧 binary"
cp -f "$OWLBACK/.env" "$OWLBACK/.env.precut-backup"
[[ -f "$SDIR/.bin/wisefido-sensor.prod-backup" ]] || cp -f "$SDIR/.bin/wisefido-sensor" "$SDIR/.bin/wisefido-sensor.prod-backup"

echo "▶ [2/6] DBN_MODE → 1（架构师 A·R22 拍定起步 mode=1）"
sed -i 's/^DBN_MODE=.*/DBN_MODE=1/' "$OWLBACK/.env"
grep -E '^DBN_MODE=' "$OWLBACK/.env"

echo "▶ [3/6] 换 cutover 二进制"
cp -f "$SDIR/.bin/wisefido-sensor.cutover" "$SDIR/.bin/wisefido-sensor"

echo "▶ [4/6] systemctl restart（原子，防双跑；systemd 管理）"
sudo systemctl restart "$SVC" || true  # SIGKILL-stop 旧进程令 restart 返非零=假警，靠下方 is-active 判定
sleep 4

echo "▶ [5/6] 启动自检（fail-safe）"
sudo systemctl is-active "$SVC" | grep -q '^active$' || { echo "❌ 服务未 active，回滚！"; "$OWLBACK/scripts/stagecut/rollback.sh" --go; exit 1; }
journalctl -u "$SVC" --since "1 min ago" -o cat 2>/dev/null | grep -q "DBN router wired" || echo "⚠️ 未见 'DBN router wired'（确认 OnRoomFrame 接通）"
echo "  服务 active；DBN_MODE=1（DBN 自发真 alarm + 固件 floor 兜 FN）"

echo "▶ [6/6] 观察（误报监控，头几小时人盯，A·R22.2.4）"
echo "  机制: journalctl -u $SVC -f -o cat | grep dbn_xray"
echo "  误报: 盯 cardagg 实际 alarm 频率(尤双雷达房)；异常 → rollback.sh --go 或 .env DBN_MODE=0"
echo "✅ 切生产完成（DBN_MODE=1）。快回滚：sed -i 's/^DBN_MODE=.*/DBN_MODE=0/' $OWLBACK/.env && systemctl restart $SVC（DBN 静默,固件 floor 接管）"
echo "   整回滚：scripts/stagecut/rollback.sh --go（还原旧 binary+.env）"
