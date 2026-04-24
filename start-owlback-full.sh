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

# 兼容旧版：启动前先清理端口和残留进程，确保一键拉起体验
echo "=== Pre-clean: stop-owlback.sh (ports/pgrep/logs) ==="
bash "$SCRIPT_DIR/stop-owlback.sh" || true

# 等待 stop 脚本完成并确保关键端口已释放，避免与刚启动的服务发生 race
echo "[*] waiting for ports to be freed (up to 10s)"
ports=(8080 8081 8082 8083 8085 8443)
for i in {1..10}; do
	busy=0
	for p in "${ports[@]}"; do
		if command -v ss >/dev/null 2>&1; then
			ss -lnt "sport = :$p" 2>/dev/null | grep -q LISTEN && busy=1 && break
		elif command -v lsof >/dev/null 2>&1; then
			lsof -nP -iTCP:"$p" -sTCP:LISTEN -t >/dev/null 2>&1 && busy=1 && break
		fi
	done
	if [ "$busy" -eq 0 ]; then
		echo "    ports free"
		break
	fi
	sleep 1
done

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

# ── 基础设施健康检查：等待 Docker 容器 ready（PG/Redis/MQTT）──
echo "[*] Waiting for infrastructure (PostgreSQL, Redis, MQTT)..."
for i in {1..30}; do
	pg_ok=0; redis_ok=0; mqtt_ok=0
	pg_isready -h 127.0.0.1 -p 5432 -q 2>/dev/null && pg_ok=1
	redis-cli -a 'TeLunSu-36kr' ping 2>/dev/null | grep -q PONG && redis_ok=1
	(echo >/dev/tcp/127.0.0.1/1883) 2>/dev/null && mqtt_ok=1
	if [ "$pg_ok" -eq 1 ] && [ "$redis_ok" -eq 1 ] && [ "$mqtt_ok" -eq 1 ]; then
		echo "    Infrastructure ready (${i}s)"
		break
	fi
	[ "$i" -eq 30 ] && echo "    WARN: timeout waiting for infrastructure (PG=$pg_ok Redis=$redis_ok MQTT=$mqtt_ok)"
	sleep 1
done

echo "=== systemctl start owlback.* ==="
# 注：不在此处 build；二进制由 start-owlback.sh（手动启动）在启动前编译。
# systemd 路径仅负责拉起已有 .bin/*，避免 root 环境下 VCS stamping 失败。

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
	echo "[*] waiting 10s after owlback.data for card table / startup sync..."
	sleep 10
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
