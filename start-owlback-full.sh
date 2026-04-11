#!/bin/bash
# 分模块启动（与 stop-owlback-full.sh 成对）
# 手动: sudo ./start-owlback-full.sh
# systemd: owlback.service 为 oneshot，ExecStart 带 --systemd-exec --no-status（仅顺序 systemctl start owlback.*）
#
# 选项:
#   --systemd-exec   用 systemctl（调用方须为 root）；且不再因「owlback 已 active」报错（供 oneshot 幂等）
#   --no-status      末尾不跑 ServiceStatus.sh
#
# 依赖: install-modular-units.sh 已安装 owlback.*.service

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
echo "OwlBack modular start (owlback.data → 30s → cardagg…qinglan… → owlback.ai)"
echo "=========================================="

echo ""
if ! $SYSTEMD_EXEC; then
	if is_active owlback; then
		echo "ERROR: unit「owlback」为 active（oneshot 编排或旧一体）。"
		echo "       若只需补启某模块: sudo systemctl start owlback.data"
		echo "       若需整栈重开: sudo ./stop-owlback-full.sh && sudo ./start-owlback-full.sh"
		exit 1
	fi
fi

echo "=== systemctl start owlback.* ==="
# 与 start-owlback.sh 一致：先起 wisefido-data（owlback.data），留时间做 card 表/全量同步，再起其它模块
DATA_WAS_ACTIVE=false
is_active owlback.data && DATA_WAS_ACTIVE=true
if $DATA_WAS_ACTIVE; then
	echo "    (already active) owlback.data"
else
	echo "[*] starting owlback.data"
	if ! run_sctl start owlback.data 2>/dev/null; then
		echo "    WARN: start owlback.data failed (journalctl -u owlback.data -n 20 --no-pager)"
	fi
fi
if ! $DATA_WAS_ACTIVE; then
	echo "[*] waiting 30s after owlback.data for card table / startup sync..."
	sleep 30
fi
sleep 1

for u in owlback.cardagg owlback.qinglan owlback.sleepace owlback.iot owlback.ai; do
	if is_active "$u"; then
		echo "    (already active) $u"
	else
		echo "[*] starting $u"
		if ! run_sctl start "$u" 2>/dev/null; then
			echo "    WARN: start $u failed (journalctl -u $u -n 20 --no-pager)"
		fi
	fi
	sleep 1
done

if ! $NO_STATUS; then
	echo ""
	echo "=== ServiceStatus.sh ==="
	bash "$SCRIPT_DIR/ServiceStatus.sh"
fi
