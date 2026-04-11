#!/usr/bin/env bash
# 安装 owlback.*.service（各模块）+ owlback.service（oneshot，仅 systemctl start/stop 编排各模块）
# 子模块无 [Install]；总控 owlback.service 可 enable 实现开机拉齐（按需）
#
# 磁盘上的 unit: owlback.data.service, …, 以及 owlback.service
#
# 实际启动: 调用 scripts/systemd/owlback-run-service.sh <wisefido-xxx>，
#           即 go run 对应模块，日志仍在 ../log/wisefido-*.log
#
# sudo ./install-modular-units.sh
set -euo pipefail
ROOT="$(cd "$(dirname "$0")" && pwd)"
if [[ "$(id -u)" -ne 0 ]]; then
  echo "请使用: sudo $0" >&2
  exit 1
fi
shopt -s nullglob
for f in "$ROOT"/owlback.*.service; do
  [[ -f "$f" ]] || continue
  install -m 644 "$f" "/etc/systemd/system/$(basename "$f")"
  echo "installed $(basename "$f")"
done
shopt -u nullglob
if [[ -f "$ROOT/owlback.service" ]]; then
  install -m 644 "$ROOT/owlback.service" "/etc/systemd/system/owlback.service"
  echo "installed owlback.service (orchestrator oneshot)"
fi
systemctl daemon-reload
echo "整栈: sudo systemctl start owlback   单模块: sudo systemctl restart owlback.data"
echo "或脚本: sudo ./start-owlback-full.sh | sudo ./stop-owlback-full.sh"
echo "若曾装 owlback-data 等连字符 unit，请 rm 后 daemon-reload"
