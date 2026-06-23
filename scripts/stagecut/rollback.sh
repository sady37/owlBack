#!/usr/bin/env bash
# rollback.sh — 回滚：cutover sensor → 旧生产 binary + 旧 .env(DBN_MODE=2)。带 --go 执行。
set -euo pipefail
OWLBACK=/home/wisefido/owl/owlBack
SDIR=$OWLBACK/wisefido-sensor
SVC=owlback.sensor.service
[[ "${1:-}" == "--go" ]] || { echo "回滚旧生产 binary+.env。确认带 --go。"; exit 2; }
echo "▶ 还原旧 binary + .env"
cp -f "$SDIR/.bin/wisefido-sensor.prod-backup" "$SDIR/.bin/wisefido-sensor"
[[ -f "$OWLBACK/.env.precut-backup" ]] && cp -f "$OWLBACK/.env.precut-backup" "$OWLBACK/.env"
sudo systemctl restart "$SVC"; sleep 3
sudo systemctl is-active "$SVC" && echo "✅ 已回滚旧生产 sensor（DBN_MODE=$(grep -E '^DBN_MODE=' "$OWLBACK/.env"|cut -d= -f2)）"
