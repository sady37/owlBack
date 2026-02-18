#!/bin/bash

# 统一停止脚本：停止所有后台服务
# 策略：1) 端口杀进程+子进程  2) 名称匹配兜底  3) 日志文件占用兜底

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
LOG_DIR="${LOG_DIR:-/tmp/owlBack_logs}"
SHELL_PID=$$
SHELL_PPID=$PPID

printf "========================================\n"
printf "Stopping OwlBack Services\n"
printf "========================================\n\n"

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

    # 1) 端口找进程
    if [ -n "$port" ]; then
        for pid in $(lsof -ti ":$port" 2>/dev/null || true); do
            [ -z "$pid" ] && continue
            seen[$pid]=1
        done
    fi

    # 2) 名称匹配
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
    if [ -n "$port" ] && lsof -ti ":$port" >/dev/null 2>&1; then
        printf "  [!] port %s still occupied, force killing...\n" "$port"
        for pid in $(lsof -ti ":$port" 2>/dev/null || true); do
            safe_kill "$pid"
        done
        sleep 1
    fi
    printf "  [✓] %s stopped\n" "$name"
}

stop_service "wisefido-data"           "8080" "$LOG_DIR/wisefido-data.log"    "wisefido-data"
stop_service "wisefido-cardagg"        ""     "$LOG_DIR/wisefido-cardagg.log" "wisefido-cardagg"
stop_service "wisefido-qinglan"        "8081" "$LOG_DIR/wisefido-qinglan.log" "wisefido-qinglan"
stop_service "wisefido-iot-timeseries" "8083" "$LOG_DIR/wisefido-iot.log"     "wisefido-iot-timeseries" "wisefido-iot"
stop_service "wisefido-ai"             ""     "$LOG_DIR/wisefido-ai.log"      "wisefido-ai"

printf "\n========================================\n"
printf "[✓] Done\n"
printf "========================================\n"
printf "\nNote: Device access services (wisefido-sleepace) are not stopped by this script\n"
