#!/usr/bin/env bash
# 模块化 systemd（与总控 owlback.service 二选一，勿混跑同端口）
#
# 磁盘上的 unit 文件名: owlback.data.service, owlback.qinglan.service, …
# 安装后 systemctl 里写:  owlback.data  或  owlback.qinglan（可省略 .service）
# 无 [Install]，勿 enable，开机不自启。
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
systemctl daemon-reload
echo "示例: systemctl restart owlback.qinglan"
echo "若曾装 owlback-data 等连字符 unit，请 rm 后 daemon-reload"
