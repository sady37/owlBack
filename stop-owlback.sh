#!/bin/bash

# 统一停止脚本：停止所有后台服务
#
# 架构说明：
# - wisefido-data: 数据管理 API + 卡片创建/更新（Redis 缓存 + config.card.*）
# - wisefido-cardagg: 数据聚合（从 PostgreSQL + Redis 聚合卡片数据并缓存到 Redis）
# - wisefido-qinglan: 设备接入与HTTPS认证 + MQTT消息路由（设备房间消费集成）
# - wisefido-iot-timeseries: 数据消费服务（从 Redis Streams 消费数据，存储到 TimescaleDB）
# - wisefido-ai: AI 智能推理服务（高级推理、访客识别、巡房优化）
#
# 注意：设备接入服务（wisefido-sleepace, wisefido-radar）不包含在此脚本中
# 这些服务需要手动停止

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
LOG_DIR="${LOG_DIR:-/tmp/owlBack_logs}"

# 获取当前 shell 的 PID 和父进程，用于保护
SHELL_PID=$$
SHELL_PPID=$PPID

# 按进程组停止：安全地杀死进程，保护 shell 和父进程
stop_service_group_kill() {
    local name="$1"
    local pattern_go="$2"
    local pattern_bin="$3"
    local log_file="$4"
    local port="$5"
    printf "[*] Stopping %s service...\n" "$name"
    PIDS=()
    
    # 收集匹配 pattern_go 的进程
    while IFS= read -r pid; do
        [ -z "$pid" ] && continue
        [ "$pid" = "$SHELL_PID" ] && continue
        [ "$pid" = "$SHELL_PPID" ] && continue
        PIDS+=("$pid")
        # 也收集其子进程
        for child in $(pgrep -P "$pid" 2>/dev/null || true); do 
            [ -n "$child" ] && [ "$child" != "$SHELL_PID" ] && [ "$child" != "$SHELL_PPID" ] && PIDS+=("$child")
        done
    done < <(pgrep -f "$pattern_go" 2>/dev/null || true)
    
    # 收集匹配 pattern_bin 的进程
    while IFS= read -r pid; do
        [ -z "$pid" ] && continue
        [ "$pid" = "$SHELL_PID" ] && continue
        [ "$pid" = "$SHELL_PPID" ] && continue
        PIDS+=("$pid")
        # 也收集其子进程
        for child in $(pgrep -P "$pid" 2>/dev/null || true); do 
            [ -n "$child" ] && [ "$child" != "$SHELL_PID" ] && [ "$child" != "$SHELL_PPID" ] && PIDS+=("$child")
        done
    done < <(pgrep -f "$pattern_bin" 2>/dev/null || true)
    
    # 查找占用日志文件的进程
    if [ -n "$log_file" ] && command -v lsof &>/dev/null; then
        while IFS= read -r pid; do 
            [ -z "$pid" ] && continue
            [ "$pid" = "$SHELL_PID" ] && continue
            [ "$pid" = "$SHELL_PPID" ] && continue
            PIDS+=("$pid")
        done < <(lsof -t "$log_file" 2>/dev/null || true)
    fi
    
    # 查找占用指定端口的进程
    if [ -n "$port" ] && command -v lsof &>/dev/null; then
        for pid in $(lsof -ti ":$port" 2>/dev/null || true); do
            [ -z "$pid" ] && continue
            [ "$pid" = "$SHELL_PID" ] && continue
            [ "$pid" = "$SHELL_PPID" ] && continue
            CWD=$(lsof -p "$pid" 2>/dev/null | grep cwd | awk '{print $NF}' || true)
            echo "$CWD" | grep -q "$name" && PIDS+=("$pid")
        done
    fi
    
    # 去重
    ALL_PIDS=($(printf '%s\n' "${PIDS[@]}" | sort -u))
    
    if [ ${#ALL_PIDS[@]} -eq 0 ]; then
        printf "  No %s process found\n" "$name"
    else
        # 只杀死具体的 PID，不使用进程组操作（避免杀死 shell）
        for pid in "${ALL_PIDS[@]}"; do
            if kill -0 "$pid" 2>/dev/null; then
                kill -9 "$pid" 2>/dev/null || true
            fi
        done
        sleep 1
        printf "  [✓] %s stopped\n" "$name"
    fi
}

printf "========================================\n"
printf "Stopping OwlBack Services\n"
printf "========================================\n"
printf "\n"

stop_service_group_kill "wisefido-data" "go run.*wisefido-data" "wisefido-data" "$LOG_DIR/wisefido-data.log" "8080"

# wisefido-cardagg：先按占用日志文件的进程杀进程组（go run | tee 时 pgrep 可能匹配不到，tee 占 log）
printf "[*] Stopping wisefido-cardagg service...\n"
if command -v lsof &>/dev/null; then
    while IFS= read -r pid; do
        [ -z "$pid" ] && continue
        [ "$pid" = "$SHELL_PID" ] && continue
        [ "$pid" = "$SHELL_PPID" ] && continue
        if kill -0 "$pid" 2>/dev/null; then
            kill -9 "$pid" 2>/dev/null || true
        fi
    done < <(lsof 2>/dev/null | grep "wisefido-cardagg.log" | awk '{print $2}' | sort -u)
    sleep 1
fi
if [ -x "$SCRIPT_DIR/wisefido-cardagg/stop-cardagg.sh" ]; then
    LOG_DIR="$LOG_DIR" "$SCRIPT_DIR/wisefido-cardagg/stop-cardagg.sh" || true
else
    stop_service_group_kill "wisefido-cardagg" "go run.*wisefido-cardagg" "wisefido-cardagg" "$LOG_DIR/wisefido-cardagg.log" ""
fi

stop_service_group_kill "wisefido-qinglan" "go run.*wisefido-qinglan" "wisefido-qinglan" "$LOG_DIR/wisefido-qinglan.log" "8081"

stop_service_group_kill "wisefido-iot-timeseries" "go run.*wisefido-iot-timeseries" "wisefido-iot-timeseries" "$LOG_DIR/wisefido-iot.log" "8083"

stop_service_group_kill "wisefido-ai" "go run.*wisefido-ai" "wisefido-ai" "$LOG_DIR/wisefido-ai.log" ""

printf "\n"

# 验证所有进程已停止
printf "Verifying all services are stopped...\n"
REMAINING=$(pgrep -f "wisefido-data|wisefido-cardagg|wisefido-qinglan|wisefido-iot-timeseries|wisefido-ai" 2>/dev/null || true)
if [ -z "$REMAINING" ]; then
    printf "========================================\n"
    printf "[✓] All backend services stopped successfully\n"
    printf "========================================\n"
    printf "\n"
    printf "Note: Device access services (wisefido-sleepace) are not stopped by this script\n"
    printf "  Stop them manually if needed\n"
else
    printf "Warning: Some processes may still be running:\n"
    echo "$REMAINING"
    printf "\n"
    printf "Attempting force kill...\n"
    # 不使用 pkill -f，用更精确的方式
    while IFS= read -r pid; do
        [ -z "$pid" ] && continue
        [ "$pid" = "$SHELL_PID" ] && continue
        [ "$pid" = "$SHELL_PPID" ] && continue
        if kill -0 "$pid" 2>/dev/null; then
            kill -9 "$pid" 2>/dev/null || true
        fi
    done < <(echo "$REMAINING" | awk '{print $1}')
    sleep 1
    REMAINING2=$(pgrep -f "wisefido-data|wisefido-cardagg|wisefido-qinglan|wisefido-iot-timeseries|wisefido-ai" 2>/dev/null || true)
    if [ -z "$REMAINING2" ]; then
        printf "[✓] All processes terminated\n"
    else
        printf "Error: Some processes could not be terminated:\n"
        echo "$REMAINING2"
        exit 1
    fi
fi

