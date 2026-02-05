#!/bin/bash

# 统一停止脚本：停止所有后台服务
#
# 架构说明：
# - wisefido-data: 数据管理 API + 卡片创建/更新（Redis 缓存 + config.card.*）
# - wisefido-card-aggregator: 数据聚合（从 PostgreSQL + Redis 聚合卡片数据并缓存到 Redis）
# - wisefido-iot-timeseries: 数据消费服务（从 Redis Streams 消费数据，存储到 TimescaleDB）
# - wisefido-ai: AI 智能推理服务（高级推理、访客识别、巡房优化）
#
# 注意：设备接入服务（wisefido-sleepace, wisefido-radar）不包含在此脚本中
# 这些服务需要手动停止

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
LOG_DIR="${LOG_DIR:-/tmp/owlBack_logs}"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 按进程组停止：收集 PID（含子进程、占用 log 的进程），先 kill -9 -pgid 再逐个 kill，最后 pkill 兜底
stop_service_group_kill() {
    local name="$1"
    local pattern_go="$2"
    local pattern_bin="$3"
    local log_file="$4"
    local port="$5"
    echo -e "${BLUE}Stopping $name service...${NC}"
    PIDS=()
    while IFS= read -r pid; do
        [ -z "$pid" ] && continue
        PIDS+=("$pid")
        for child in $(pgrep -P "$pid" 2>/dev/null); do [ -n "$child" ] && PIDS+=("$child"); done
    done < <(pgrep -f "$pattern_go" 2>/dev/null || true)
    while IFS= read -r pid; do
        [ -z "$pid" ] && continue
        PIDS+=("$pid")
        for child in $(pgrep -P "$pid" 2>/dev/null); do [ -n "$child" ] && PIDS+=("$child"); done
    done < <(pgrep -f "$pattern_bin" 2>/dev/null || true)
    if [ -n "$log_file" ] && command -v lsof &>/dev/null; then
        while IFS= read -r pid; do [ -z "$pid" ] || PIDS+=("$pid"); done < <(lsof -t "$log_file" 2>/dev/null || true)
    fi
    if [ -n "$port" ] && command -v lsof &>/dev/null; then
        for pid in $(lsof -ti ":$port" 2>/dev/null); do
            [ -z "$pid" ] && continue
            CWD=$(lsof -p "$pid" 2>/dev/null | grep cwd | awk '{print $NF}' || true)
            echo "$CWD" | grep -q "$name" && PIDS+=("$pid")
        done
    fi
    ALL_PIDS=($(printf '%s\n' "${PIDS[@]}" | sort -u))
    if [ ${#ALL_PIDS[@]} -eq 0 ]; then
        echo -e "${YELLOW}  No $name process found${NC}"
    else
        for pid in "${ALL_PIDS[@]}"; do
            if kill -0 "$pid" 2>/dev/null; then
                pgid=$(ps -o pgid= -p "$pid" 2>/dev/null | tr -d ' ')
                [ -n "$pgid" ] && [ "$pgid" != "0" ] && kill -9 -"$pgid" 2>/dev/null || true
            fi
        done
        for pid in "${ALL_PIDS[@]}"; do
            kill -0 "$pid" 2>/dev/null && kill -9 "$pid" 2>/dev/null || true
        done
        sleep 1
        pkill -9 -f "$pattern_go" 2>/dev/null || true
        pkill -9 -f "$pattern_bin" 2>/dev/null || true
        echo -e "${GREEN}  $name stopped${NC}"
    fi
    echo ""
}

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}Stopping OwlBack Services${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""

stop_service_group_kill "wisefido-data" "go run.*wisefido-data" "wisefido-data" "$LOG_DIR/wisefido-data.log" "8080"

# wisefido-card-aggregator：先按占用日志文件的进程杀进程组（go run | tee 时 pgrep 可能匹配不到，tee 占 log）
echo -e "${BLUE}Stopping wisefido-card-aggregator service...${NC}"
if command -v lsof &>/dev/null; then
    while IFS= read -r pid; do
        [ -z "$pid" ] || [ "$pid" = "$$" ] && continue
        pgid=$(ps -o pgid= -p "$pid" 2>/dev/null | tr -d ' ')
        if [ -n "$pgid" ] && [ "$pgid" != "0" ]; then
            kill -9 -"$pgid" 2>/dev/null || true
        fi
    done < <(lsof 2>/dev/null | grep "wisefido-card-aggregator.log" | awk '{print $2}' | sort -u)
    sleep 1
fi
if [ -x "$SCRIPT_DIR/wisefido-card-aggregator/stop-cardagg.sh" ]; then
    LOG_DIR="$LOG_DIR" "$SCRIPT_DIR/wisefido-card-aggregator/stop-cardagg.sh" || true
else
    stop_service_group_kill "wisefido-card-aggregator" "go run.*wisefido-card-aggregator" "wisefido-card-aggregator" "$LOG_DIR/wisefido-card-aggregator.log" ""
fi
echo ""

stop_service_group_kill "wisefido-iot-timeseries" "go run.*wisefido-iot-timeseries" "wisefido-iot-timeseries" "$LOG_DIR/wisefido-iot.log" "8083"

stop_service_group_kill "wisefido-ai" "go run.*wisefido-ai" "wisefido-ai" "$LOG_DIR/wisefido-ai.log" ""

echo ""

# 验证所有进程已停止
echo -e "${BLUE}Verifying all services are stopped...${NC}"
REMAINING=$(pgrep -f "wisefido-data\|wisefido-card-aggregator\|wisefido-card-manage\|wisefido-iot-timeseries\|wisefido-ai" || true)
if [ -z "$REMAINING" ]; then
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}All backend services stopped successfully${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo -e "${YELLOW}Note: Device access services (wisefido-sleepace, wisefido-radar) are not stopped by this script${NC}"
    echo -e "${YELLOW}  Stop them manually if needed${NC}"
else
    echo -e "${RED}Warning: Some processes may still be running:${NC}"
    echo "$REMAINING"
    echo -e "${YELLOW}Attempting force kill...${NC}"
    pkill -9 -f "wisefido-data\|wisefido-card-aggregator\|wisefido-iot-timeseries\|wisefido-ai" 2>/dev/null || true
    sleep 1
    REMAINING2=$(pgrep -f "wisefido-data\|wisefido-card-aggregator\|wisefido-iot-timeseries\|wisefido-ai" || true)
    if [ -z "$REMAINING2" ]; then
        echo -e "${GREEN}All processes terminated${NC}"
    else
        echo -e "${RED}Some processes could not be terminated:${NC}"
        echo "$REMAINING2"
        exit 1
    fi
fi

