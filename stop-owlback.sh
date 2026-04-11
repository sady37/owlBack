#!/bin/bash

# 启停成对：start-owlback.sh ↔ 本脚本；systemctl stop owlback 会执行本脚本（ExecStop）
# 分模块部署：sudo ./start-owlback-full.sh 启 ； sudo ./stop-owlback-full.sh 停（先各 unit 再本脚本清场再 ServiceStatus）
# 勿在 ExecStop 内再 systemctl stop owlback（已移除）；分模块也可单独 systemctl stop owlback.data 等
# 单服务请用各目录 ./stop-qinglan.sh、./stop-cardagg.sh … 与对应 ./start-*.sh 成对
#
# 统一停止脚本：停止所有后台服务
# 策略：1) 端口 LISTEN 找 PID（lsof + ss 双保险） 2) 名称 / main.go 匹配  3) 日志文件占用
# 建议与启动使用同一用户执行；若用 sudo 而服务由普通用户启动，一般仍可杀，但需系统有 lsof 或 ss。

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OWL_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
LOG_DIR="${LOG_DIR:-$OWL_ROOT/log}"
SHELL_PID=$$
SHELL_PPID=$PPID

# systemd 一体 owlback：ExecStop=... stop-owlback.sh $MAINPID（unit 里由 systemd 展开）
OWL_MONOLITH_MAIN="${1:-}"
[ -z "$OWL_MONOLITH_MAIN" ] && OWL_MONOLITH_MAIN="${MAINPID:-}"
case "$OWL_MONOLITH_MAIN" in ''|0|'$MAINPID') OWL_MONOLITH_MAIN= ;; esac

printf "========================================\n"
printf "Stopping OwlBack Services\n"
printf "========================================\n\n"

# 本机谁在 TCP port 上 LISTEN（lsof + ss + fuser，去重；部分环境 lsof/ss 漏检时用 fuser）
pids_listen_tcp_port() {
    local port="$1"
    local p ssout fu
    {
        if command -v lsof >/dev/null 2>&1; then
            p=$(lsof -nP -iTCP:"$port" -sTCP:LISTEN -t 2>/dev/null || true)
            [ -z "$p" ] && p=$(lsof -nP -i :"$port" -sTCP:LISTEN -t 2>/dev/null || true)
            for x in $p; do printf '%s\n' "$x"; done
        fi
        if command -v ss >/dev/null 2>&1; then
            ssout=$(ss -lntp "sport = :$port" 2>/dev/null || true)
            [ -z "$ssout" ] && ssout=$(ss -lntp 2>/dev/null | grep -E ":${port}([^0-9]|$)" || true)
            echo "$ssout" | sed -n 's/.*pid=\([0-9][0-9]*\).*/\1/p'
        fi
        if command -v fuser >/dev/null 2>&1; then
            fu=$(fuser "${port}/tcp" 2>&1 || true)
            # 典型输出 "8080/tcp:            359985 360001"
            echo "$fu" | sed 's/.*:[[:space:]]*//' | tr -s ' ' '\n' | grep -E '^[0-9]+$'
        fi
    } | sort -u -n
}

# 安全 kill：排除自身和父进程
safe_kill() {
    local pid="$1"
    [ -z "$pid" ] && return
    [ "$pid" = "$SHELL_PID" ] || [ "$pid" = "$SHELL_PPID" ] && return
    kill -9 "$pid" 2>/dev/null || true
}

# 杀 pid 及其所有子进程
kill_tree() {
    local pid="$1"
    [ -z "$pid" ] && return
    for child in $(pgrep -P "$pid" 2>/dev/null || true); do
        kill_tree "$child"
    done
    safe_kill "$pid"
}

# 0) 一体 start-owlback.sh：先杀 MainPID 整棵子树（所有后台 go run），否则仅靠端口/pgrep 易漏
if [ -n "$OWL_MONOLITH_MAIN" ] && [ "$OWL_MONOLITH_MAIN" != "$SHELL_PID" ] && [ "$OWL_MONOLITH_MAIN" != "$SHELL_PPID" ]; then
    if kill -0 "$OWL_MONOLITH_MAIN" 2>/dev/null; then
        printf "[*] kill_tree owlback monolith MainPID=%s (start-owlback.sh → go run children)\n" "$OWL_MONOLITH_MAIN"
        kill_tree "$OWL_MONOLITH_MAIN"
        sleep 1
    fi
fi

# 收集并杀掉一个服务的所有相关进程
stop_service() {
    local name="$1"
    local port="$2"        # 可为空
    local log_file="$3"    # 可为空
    shift 3
    local patterns=("$@")  # 剩余参数为名称匹配 pattern

    printf "[*] Stopping %s...\n" "$name"
    local -A seen  # 去重

    # 1) 端口 LISTEN（lsof + ss；go 子进程 cmdline 常无模块名，端口最可靠）
    if [ -n "$port" ]; then
        for pid in $(pids_listen_tcp_port "$port" | sort -u -n); do
            [ -z "$pid" ] && continue
            seen[$pid]=1
        done
    fi

    # 2) 名称匹配（含 go run / main.go 路径，兼容编译后短 cmdline）
    for pat in "${patterns[@]}"; do
        for pid in $(pgrep -f "$pat" 2>/dev/null || true); do
            [ -z "$pid" ] && continue
            seen[$pid]=1
        done
    done

    # 3) 日志文件占用
    if [ -n "$log_file" ] && [ -f "$log_file" ]; then
        for pid in $(lsof -t "$log_file" 2>/dev/null || true); do
            [ -z "$pid" ] && continue
            seen[$pid]=1
        done
    fi

    if [ ${#seen[@]} -eq 0 ]; then
        printf "  No %s process found\n" "$name"
        return
    fi

    # kill 所有收集到的 pid（含子进程树）
    for pid in "${!seen[@]}"; do
        kill_tree "$pid"
    done
    sleep 1

    # 验证端口释放
    if [ -n "$port" ] && [ -n "$(pids_listen_tcp_port "$port" | head -1)" ]; then
        printf "  [!] port %s still LISTEN, force killing...\n" "$port"
        for pid in $(pids_listen_tcp_port "$port" | sort -u -n); do
            safe_kill "$pid"
        done
        sleep 1
    fi
    printf "  [✓] %s stopped\n" "$name"
}

# systemd：只停「分模块」unit，绝不在这里对 owlback 本体再执行 systemctl stop owlback。
# 原因：本脚本就是 owlback.service 的 ExecStop；若再 stop owlback，会与当前 stop job 嵌套等待，表现为停不掉/超时。
# 一体 owlback 的停服：请只用「systemctl stop owlback」；systemd 会调本脚本杀子进程。
# 若你手动执行本脚本且希望停掉一体 unit：请在另一条命令执行 systemctl stop owlback，不要在本脚本里递归 stop。
if command -v systemctl >/dev/null 2>&1; then
    for u in owlback.data owlback.cardagg owlback.qinglan owlback.sleepace owlback.iot owlback.ai; do
        if systemctl is-active --quiet "$u" 2>/dev/null; then
            printf "[*] systemctl stop %s\n" "$u"
            systemctl stop "$u" 2>/dev/null || true
        fi
    done
    sleep 1
fi

stop_service "wisefido-data"           "8080" "$LOG_DIR/wisefido-data.log"      "wisefido-data" "cmd/wisefido-data/main.go"
stop_service "wisefido-cardagg"        ""     "$LOG_DIR/wisefido-cardagg.log"   "wisefido-cardagg" "wisefido-cardagg/main.go" "cardagg/main.go"
stop_service "wisefido-qinglan"        "8081" "$LOG_DIR/wisefido-qinglan.log"   "wisefido-qinglan" "cmd/wisefido-qinglan/main.go"
stop_service "wisefido-sleepace"       "8083" "$LOG_DIR/wisefido-sleepace.log"  "wisefido-sleepace" "cmd/wisefido-sleepace/main.go"
stop_service "wisefido-iot-timeseries" "8085" "$LOG_DIR/wisefido-iot.log"       "wisefido-iot-timeseries" "wisefido-iot" "cmd/wisefido-iot/main.go"
stop_service "wisefido-ai"             ""     "$LOG_DIR/wisefido-ai.log"        "wisefido-ai" "cmd/wisefido-ai/main.go"

# 再扫一轮约定端口（避免单服务阶段 lsof 与日志路径不一致导致漏杀）
# 8443 = wisefido-qinglan HTTPS 认证；此前未扫会导致「8081 已空但 8443 仍被旧进程占住」
printf "\n[*] Sweep LISTEN on 8080 8081 8083 8085 8443...\n"
for port in 8080 8081 8083 8085 8443; do
    for pid in $(pids_listen_tcp_port "$port"); do
        [ -z "$pid" ] && continue
        printf "  kill_tree pid %s (:%s)\n" "$pid" "$port"
        kill_tree "$pid"
    done
done
sleep 1

# 最后 fuser 强杀仍占用约定端口的进程（仅 OwlBack 常用口）
if command -v fuser >/dev/null 2>&1; then
    for port in 8080 8081 8083 8085 8443; do
        if fuser "${port}/tcp" >/dev/null 2>&1; then
            printf "  fuser -k %s/tcp\n" "$port"
            fuser -k "${port}/tcp" 2>/dev/null || true
        fi
    done
fi

printf "\n========================================\n"
printf "[✓] Done\n"
printf "========================================\n"
