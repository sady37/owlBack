#!/bin/bash

# 统一停止脚本：停止所有后台服务
# 策略：1) 端口 LISTEN 找 PID（lsof + ss 双保险） 2) 名称 / main.go 匹配  3) 日志文件占用
# 建议与启动使用同一用户执行；若用 sudo 而服务由普通用户启动，一般仍可杀，但需系统有 lsof 或 ss。

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OWL_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
LOG_DIR="${LOG_DIR:-$OWL_ROOT/log}"
SHELL_PID=$$
SHELL_PPID=$PPID

printf "========================================\n"
printf "Stopping OwlBack Services\n"
printf "========================================\n\n"

# 本机谁在 TCP port 上 LISTEN（lsof + ss，去重；go 子进程 cmdline 常无模块名，端口最准）
pids_listen_tcp_port() {
    local port="$1"
    local p ssout
    {
        if command -v lsof >/dev/null 2>&1; then
            p=$(lsof -nP -iTCP:"$port" -sTCP:LISTEN -t 2>/dev/null || true)
            [ -z "$p" ] && p=$(lsof -nP -i :"$port" -sTCP:LISTEN -t 2>/dev/null || true)
            for x in $p; do printf '%s\n' "$x"; done
        fi
        if command -v ss >/dev/null 2>&1; then
            ssout=$(ss -lntp "sport = :$port" 2>/dev/null || true)
            [ -z "$ssout" ] && ssout=$(ss -lntp 2>/dev/null | grep ":$port" || true)
            echo "$ssout" | sed -n 's/.*pid=\([0-9][0-9]*\).*/\1/p'
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

stop_service "wisefido-data"           "8080" "$LOG_DIR/wisefido-data.log"      "wisefido-data" "cmd/wisefido-data/main.go"
stop_service "wisefido-cardagg"        ""     "$LOG_DIR/wisefido-cardagg.log"   "wisefido-cardagg" "wisefido-cardagg/main.go"
stop_service "wisefido-qinglan"        "8081" "$LOG_DIR/wisefido-qinglan.log"   "wisefido-qinglan" "cmd/wisefido-qinglan/main.go"
stop_service "wisefido-sleepace"       "8083" "$LOG_DIR/wisefido-sleepace.log"  "wisefido-sleepace" "cmd/wisefido-sleepace/main.go"
stop_service "wisefido-iot-timeseries" "8085" "$LOG_DIR/wisefido-iot.log"       "wisefido-iot-timeseries" "wisefido-iot" "cmd/wisefido-iot/main.go"
stop_service "wisefido-ai"             ""     "$LOG_DIR/wisefido-ai.log"        "wisefido-ai" "cmd/wisefido-ai/main.go"

printf "\n========================================\n"
printf "[✓] Done\n"
printf "========================================\n"
