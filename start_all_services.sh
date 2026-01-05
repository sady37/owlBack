#!/bin/bash

# 统一启动脚本：同时启动 wisefido-data 和 wisefido-card-aggregator 服务
# 输出日志到统一文件，并显示在终端
#
# 架构说明：
# - wisefido-data: 数据管理 API + 卡片创建/更新（直接同步调用 CardCreator）
# - wisefido-card-aggregator: 数据聚合（从 PostgreSQL + Redis 聚合卡片数据并缓存到 Redis）
#   注意：卡片创建/更新现在由 wisefido-data 直接处理，aggregator 主要用于数据聚合

# 注意：不使用 set -e，以便在服务失败时能执行清理函数

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志目录
LOG_DIR="${LOG_DIR:-/tmp/owlBack_logs}"
mkdir -p "$LOG_DIR"

# 日志文件
DATA_LOG="$LOG_DIR/wisefido-data.log"
AGGREGATOR_LOG="$LOG_DIR/wisefido-card-aggregator.log"
COMBINED_LOG="$LOG_DIR/combined.log"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Starting OwlBack Services${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed or not in PATH${NC}"
    exit 1
fi

# 检查是否有服务正在运行
check_running_services() {
    local data_running=false
    local aggregator_running=false
    
    # 检查 wisefido-data
    if pgrep -f "go run.*wisefido-data" > /dev/null 2>&1 || \
       pgrep -f "wisefido-data" > /dev/null 2>&1 || \
       (command -v lsof &> /dev/null && lsof -ti :8080 > /dev/null 2>&1); then
        data_running=true
    fi
    
    # 检查 wisefido-card-aggregator
    if pgrep -f "go run.*wisefido-card-aggregator" > /dev/null 2>&1 || \
       pgrep -f "wisefido-card-aggregator" > /dev/null 2>&1; then
        aggregator_running=true
    fi
    
    if [ "$data_running" = true ] || [ "$aggregator_running" = true ]; then
        echo -e "${YELLOW}Warning: Services are already running${NC}"
        if [ "$data_running" = true ]; then
            echo -e "${YELLOW}  - wisefido-data is running${NC}"
        fi
        if [ "$aggregator_running" = true ]; then
            echo -e "${YELLOW}  - wisefido-card-aggregator is running${NC}"
        fi
        echo ""
        echo -e "${BLUE}Options:${NC}"
        echo "  1) Stop existing services and restart"
        echo "  2) Exit (keep existing services running)"
        echo ""
        read -p "Choose an option (1 or 2): " -n 1 -r
        echo ""
        if [[ $REPLY =~ ^[1]$ ]]; then
            echo -e "${YELLOW}Stopping existing services...${NC}"
            # 停止现有服务
            pkill -f "go run.*wisefido-data" 2>/dev/null || true
            pkill -f "go run.*wisefido-card-aggregator" 2>/dev/null || true
            pkill -f "wisefido-data" 2>/dev/null || true
            pkill -f "wisefido-card-aggregator" 2>/dev/null || true
            if command -v lsof &> /dev/null; then
                PORT_PIDS=$(lsof -ti :8080 2>/dev/null || true)
                for pid in $PORT_PIDS; do
                    if [ -n "$pid" ]; then
                        CWD=$(lsof -p "$pid" 2>/dev/null | grep cwd | awk '{print $NF}' || true)
                        if echo "$CWD" | grep -q "wisefido-data"; then
                            kill -9 "$pid" 2>/dev/null || true
                        fi
                    fi
                done
            fi
            sleep 2
            echo -e "${GREEN}Existing services stopped${NC}"
            echo ""
        else
            echo -e "${BLUE}Exiting. Existing services will continue running.${NC}"
            exit 0
        fi
    fi
}

# 检查是否有服务正在运行
check_running_services

# 检查端口占用（可选，如果 lsof 不可用则跳过）
check_port() {
    local port=$1
    local service_name=$2
    
    # 检查 lsof 是否可用
    if ! command -v lsof &> /dev/null; then
        echo -e "${YELLOW}Note: lsof not available, skipping port check for $service_name${NC}"
        return 0
    fi
    
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo -e "${YELLOW}Warning: Port $port is already in use${NC}"
        echo "  This may prevent $service_name from starting"
        echo "  To find the process using port $port, run:"
        echo "    lsof -i :$port"
        echo ""
        read -p "Continue anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo -e "${RED}Aborted by user${NC}"
            exit 1
        fi
    fi
}

# 检查 wisefido-data 端口 (8080)
# 注意：wisefido-card-aggregator 不需要端口，它是纯后台服务
check_port 8080 "wisefido-data"

# 设置默认环境变量
export DB_HOST="${DB_HOST:-localhost}"
export DB_PORT="${DB_PORT:-5432}"
export DB_USER="${DB_USER:-postgres}"
export DB_PASSWORD="${DB_PASSWORD:-postgres}"
export DB_NAME="${DB_NAME:-owlrd}"
export DB_SSLMODE="${DB_SSLMODE:-disable}"

export REDIS_ADDR="${REDIS_ADDR:-localhost:6379}"
export REDIS_PASSWORD="${REDIS_PASSWORD:-}"

# wisefido-card-aggregator 配置
# 注意：卡片创建/更新现在由 wisefido-data 直接处理（同步调用）
# wisefido-card-aggregator 主要用于数据聚合（从 PostgreSQL + Redis 聚合数据并缓存）
export TENANT_ID="${TENANT_ID:-bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c}"
# 卡片创建轮询模式配置（作为保底机制）
# 注意：由于 wisefido-data 已经直接处理卡片更新，这里只是保底机制
# 如果设置为 "polling"：
#   - 如果 CARD_POLLING_INTERVAL >= 86400 (24小时)，会自动使用定时任务（每天8点执行）
#   - 如果 CARD_POLLING_INTERVAL < 86400，使用固定间隔轮询
# 如果设置为 "events"，wisefido-card-aggregator 会监听 Redis Streams 事件（当前未使用）
export CARD_TRIGGER_MODE="${CARD_TRIGGER_MODE:-polling}"
# 轮询间隔（秒）：默认 86400 秒（24 小时）
# 当 >= 86400 时，会自动使用每天8点的定时任务（而不是固定间隔）
# 当 < 86400 时，使用固定间隔轮询
export CARD_POLLING_INTERVAL="${CARD_POLLING_INTERVAL:-86400}"  # 24 小时（自动切换为每天8点执行）
# 数据聚合配置（wisefido-card-aggregator 的主要功能）
# CARD_AGGREGATION_ENABLED: 是否启用数据聚合（从 PostgreSQL + Redis 聚合并缓存到 Redis）
export CARD_AGGREGATION_ENABLED="${CARD_AGGREGATION_ENABLED:-true}"  # 启用数据聚合
# CARD_AGGREGATION_INTERVAL: 数据聚合间隔（秒），默认 2 秒
# 功能：每 2 秒聚合一次所有卡片数据（从 PostgreSQL + Redis 读取，组装成完整的 VitalFocusCard 对象，缓存到 Redis）
# 原因：心率呼吸数据每 2 秒更新一次（前端 mockRadarData.ts），聚合间隔应匹配数据更新频率
# 启动时：服务启动时会立即执行一次全量聚合（在 startDataAggregation 函数中）
export CARD_AGGREGATION_INTERVAL="${CARD_AGGREGATION_INTERVAL:-2}"  # 每 2 秒聚合一次（匹配心率呼吸数据更新频率）

# 日志配置
export LOG_LEVEL="${LOG_LEVEL:-info}"
export LOG_FORMAT="${LOG_FORMAT:-json}"

# 显示配置
echo -e "${BLUE}Configuration:${NC}"
echo "  DB_HOST: $DB_HOST"
echo "  DB_NAME: $DB_NAME"
echo "  REDIS_ADDR: $REDIS_ADDR"
echo "  TENANT_ID: $TENANT_ID"
echo "  CARD_TRIGGER_MODE: $CARD_TRIGGER_MODE (card creation now handled by wisefido-data, this is backup)"
if [ "$CARD_POLLING_INTERVAL" -ge 86400 ]; then
    echo "  CARD_POLLING_INTERVAL: $CARD_POLLING_INTERVAL seconds (auto-switched to daily 8:00 AM scheduled task)"
else
    echo "  CARD_POLLING_INTERVAL: $CARD_POLLING_INTERVAL seconds (fixed interval polling)"
fi
echo "  CARD_AGGREGATION_ENABLED: $CARD_AGGREGATION_ENABLED (main function: aggregate card data)"
echo "  CARD_AGGREGATION_INTERVAL: $CARD_AGGREGATION_INTERVAL seconds (aggregate every ${CARD_AGGREGATION_INTERVAL}s to match heart/breath rate update frequency, runs on startup)"
echo "  LOG_LEVEL: $LOG_LEVEL"
echo ""
echo -e "${BLUE}Service startup behavior:${NC}"
echo "  - wisefido-data: Full card check/update on startup (after 2s delay, in main.go)"
echo "  - wisefido-card-aggregator:"
echo "    * Full card creation on startup (in startPollingMode)"
echo "    * Data aggregation on startup + every ${CARD_AGGREGATION_INTERVAL}s (in startDataAggregation)"
echo ""
echo -e "${BLUE}Log files:${NC}"
echo "  wisefido-data: $DATA_LOG"
echo "  wisefido-card-aggregator: $AGGREGATOR_LOG"
echo "  combined: $COMBINED_LOG"
echo ""

# 清理旧日志（可选）
if [ "$1" == "--clean" ]; then
    echo -e "${YELLOW}Cleaning old log files...${NC}"
    rm -f "$DATA_LOG" "$AGGREGATOR_LOG" "$COMBINED_LOG"
fi

# 函数：清理后台进程
cleanup() {
    echo ""
    echo -e "${YELLOW}Shutting down services...${NC}"
    pkill -f "go run.*wisefido-data" || true
    pkill -f "go run.*wisefido-card-aggregator" || true
    pkill -f "wisefido-data" || true
    pkill -f "wisefido-card-aggregator" || true
    echo -e "${GREEN}Services stopped${NC}"
    exit 0
}

# 捕获退出信号
trap cleanup SIGINT SIGTERM

# 获取脚本所在目录（owlBack 根目录）
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OWLBACK_DIR="$SCRIPT_DIR"

# 检查服务目录是否存在
if [ ! -d "$OWLBACK_DIR/wisefido-data" ]; then
    echo -e "${RED}Error: wisefido-data directory not found at $OWLBACK_DIR/wisefido-data${NC}"
    exit 1
fi

if [ ! -d "$OWLBACK_DIR/wisefido-card-aggregator" ]; then
    echo -e "${RED}Error: wisefido-card-aggregator directory not found at $OWLBACK_DIR/wisefido-card-aggregator${NC}"
    exit 1
fi

# 启动 wisefido-data 服务
# 功能：数据管理 API + 卡片创建/更新（直接同步调用 CardCreator）
echo -e "${GREEN}[1/2] Starting wisefido-data service...${NC}"
echo -e "${BLUE}  Function: Data management API + Card creation/update (direct sync)${NC}"
cd "$OWLBACK_DIR/wisefido-data"
# 直接输出到终端，同时使用 tee 写入日志文件
go run cmd/wisefido-data/main.go 2>&1 | tee "$DATA_LOG" &
DATA_PID=$!
echo "  PID: $DATA_PID"
echo "  Log: $DATA_LOG (also displayed in terminal)"

# 等待一下确保服务启动
sleep 2

# 启动 wisefido-card-aggregator 服务
# 功能：数据聚合（从 PostgreSQL + Redis 聚合卡片数据并缓存）
# 注意：卡片创建/更新现在由 wisefido-data 直接处理，aggregator 主要用于数据聚合
echo -e "${GREEN}[2/2] Starting wisefido-card-aggregator service...${NC}"
echo -e "${BLUE}  Function: Data aggregation (PostgreSQL + Redis → full card cache)${NC}"
echo -e "${YELLOW}  Note: Card creation/update is now handled by wisefido-data${NC}"
cd "$OWLBACK_DIR/wisefido-card-aggregator"
# 直接输出到终端，同时使用 tee 写入日志文件
go run cmd/wisefido-card-aggregator/main.go 2>&1 | tee "$AGGREGATOR_LOG" &
AGGREGATOR_PID=$!
echo "  PID: $AGGREGATOR_PID"
echo "  Log: $AGGREGATOR_LOG (also displayed in terminal)"

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Both services are running${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${BLUE}Monitoring logs (Ctrl+C to stop)...${NC}"
echo ""

# 等待服务启动
sleep 3

# 清理函数：在退出时停止所有后台服务
cleanup() {
    echo ""
    echo -e "${YELLOW}Cleaning up services...${NC}"
    if [ -n "$DATA_PID" ] && ps -p $DATA_PID > /dev/null 2>&1; then
        echo "  Stopping wisefido-data (PID: $DATA_PID)"
        kill $DATA_PID 2>/dev/null || true
    fi
    if [ -n "$AGGREGATOR_PID" ] && ps -p $AGGREGATOR_PID > /dev/null 2>&1; then
        echo "  Stopping wisefido-card-aggregator (PID: $AGGREGATOR_PID)"
        kill $AGGREGATOR_PID 2>/dev/null || true
    fi
    # 等待进程退出
    sleep 2
    # 强制杀死仍在运行的进程
    pkill -f "go run.*wisefido-data" 2>/dev/null || true
    pkill -f "go run.*wisefido-card-aggregator" 2>/dev/null || true
    echo -e "${GREEN}Cleanup completed${NC}"
}

# 注册清理函数（在脚本退出时执行）
trap cleanup EXIT INT TERM

# 检查服务是否正常运行
DATA_FAILED=false
AGGREGATOR_FAILED=false

if ! ps -p $DATA_PID > /dev/null 2>&1; then
    echo -e "${RED}Error: wisefido-data service failed to start${NC}"
    echo "Check log: $DATA_LOG"
    tail -20 "$DATA_LOG"
    DATA_FAILED=true
fi

if ! ps -p $AGGREGATOR_PID > /dev/null 2>&1; then
    echo -e "${RED}Error: wisefido-card-aggregator service failed to start${NC}"
    echo "Check log: $AGGREGATOR_LOG"
    tail -20 "$AGGREGATOR_LOG"
    AGGREGATOR_FAILED=true
fi

# 如果两个服务都失败，退出
if [ "$DATA_FAILED" = true ] && [ "$AGGREGATOR_FAILED" = true ]; then
    echo -e "${RED}Both services failed to start. Exiting.${NC}"
    exit 1
fi

# 如果只有一个服务失败，警告但继续
if [ "$DATA_FAILED" = true ]; then
    echo -e "${YELLOW}Warning: wisefido-data failed, but continuing with wisefido-card-aggregator${NC}"
    echo "  You can check the log: $DATA_LOG"
    echo ""
fi

if [ "$AGGREGATOR_FAILED" = true ]; then
    echo -e "${YELLOW}Warning: wisefido-card-aggregator failed, but continuing with wisefido-data${NC}"
    echo "  You can check the log: $AGGREGATOR_LOG"
    echo ""
fi

echo -e "${GREEN}Services started successfully!${NC}"
echo ""
echo -e "${BLUE}Logs are being displayed in this terminal and saved to:${NC}"
echo "  - $DATA_LOG"
echo "  - $AGGREGATOR_LOG"
echo ""
echo -e "${YELLOW}Press Ctrl+C to stop all services${NC}"
echo ""

# 等待进程（这样 Ctrl+C 可以正确捕获并停止服务）
# 只等待成功启动的进程
if [ "$DATA_FAILED" = false ] && [ "$AGGREGATOR_FAILED" = false ]; then
    wait $DATA_PID $AGGREGATOR_PID 2>/dev/null || true
elif [ "$DATA_FAILED" = false ]; then
    wait $DATA_PID 2>/dev/null || true
elif [ "$AGGREGATOR_FAILED" = false ]; then
    wait $AGGREGATOR_PID 2>/dev/null || true
fi

