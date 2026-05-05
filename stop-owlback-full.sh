#!/bin/bash
# 分模块停服 + stop-owlback.sh 清场 +（可选）ServiceStatus
# 手动: sudo ./stop-owlback-full.sh
# systemd: owlback.service ExecStop 须带 --systemd-exec --no-status（勿在脚本内再 stop owlback，避免死锁）
#
# 选项:
#   --systemd-exec   用 systemctl（须 root）；且跳过「systemctl stop owlback」一步
#   --no-status      末尾不跑 ServiceStatus.sh
#
# 成对启动: sudo ./start-owlback-full.sh

set -u

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR" || exit 1

SYSTEMD_EXEC=false
NO_STATUS=false
for _arg in "$@"; do
	case "$_arg" in
		--systemd-exec) SYSTEMD_EXEC=true ;;
		--no-status) NO_STATUS=true ;;
	esac
done

run_sctl() {
	if $SYSTEMD_EXEC || [[ "$(id -u)" -eq 0 ]]; then
		systemctl "$@"
	else
		sudo systemctl "$@"
	fi
}

is_active() {
	run_sctl is-active --quiet "$1" 2>/dev/null
}

echo "=========================================="
echo "OwlBack modular stop + sweep"
echo "=========================================="

echo ""
echo "=== 1) systemctl stop owlback.* (reverse order) ==="
for u in owlback.sensor owlback.ai owlback.iot owlback.sleepace owlback.qinglan owlback.cardagg owlback.data; do
	if is_active "$u"; then
		echo "[*] stopping $u"
		run_sctl stop "$u" 2>/dev/null || true
	else
		echo "    (inactive) $u"
	fi
done

if ! $SYSTEMD_EXEC; then
	echo ""
	echo "=== 2) systemctl stop owlback (orchestrator / 旧一体, if active) ==="
	if is_active owlback; then
		echo "[*] stopping owlback"
		run_sctl stop owlback 2>/dev/null || true
	else
		echo "    (inactive) owlback"
	fi
fi

sleep 1

echo ""
echo "=== stop-owlback.sh (ports / pgrep / logs) ==="
bash "$SCRIPT_DIR/stop-owlback.sh" || true

if ! $NO_STATUS; then
	echo ""
	echo "=== ServiceStatus.sh ==="
	bash "$SCRIPT_DIR/ServiceStatus.sh"
fi
