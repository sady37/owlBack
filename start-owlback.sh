#!/bin/bash

# 统一启动脚本：启动所有后台服务
# 输出日志到统一文件，并显示在终端
#
# 架构说明：
# - wisefido-data: 数据管理 API + 卡片创建/更新（Redis 缓存 + config.card.*）
# - wisefido-cardagg: 数据聚合（从 PostgreSQL + Redis 聚合卡片数据并缓存到 Redis）
# - wisefido-iot: 数据消费服务（从 Redis Streams 消费数据，存储到 TimescaleDB）
# - wisefido-ai: AI 智能推理服务（高级推理、访客识别、巡房优化）
#
# Radar 网关由 wisefido-qinglan 提供（本脚本已启动）

# 注意：不使用 set -e，以便在服务失败时能执行清理函数

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 脚本所在目录（先定义，供 check_running_services 里 stop-owlback.sh 使用）
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OWL_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# 日志统一放到 owl/log/
LOG_DIR="${LOG_DIR:-$OWL_ROOT/log}"
mkdir -p "$LOG_DIR"

# 日志文件
DATA_LOG="$LOG_DIR/wisefido-data.log"
AGGREGATOR_LOG="$LOG_DIR/wisefido-cardagg.log"
QINGLAN_LOG="$LOG_DIR/wisefido-qinglan.log"
SLEEPACE_LOG="$LOG_DIR/wisefido-sleepace.log"
IOT_LOG="$LOG_DIR/wisefido-iot.log"
AI_LOG="$LOG_DIR/wisefido-ai.log"
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

# 检查是否有服务正在运行（只看端口）
check_running_services() {
    local any_running=false
    local running_names=()

    if lsof -ti :8080 >/dev/null 2>&1; then any_running=true; running_names+=("wisefido-data(:8080)"); fi
    if lsof -ti :8081 >/dev/null 2>&1; then any_running=true; running_names+=("wisefido-qinglan(:8081)"); fi
    if lsof -ti :8083 >/dev/null 2>&1; then any_running=true; running_names+=("wisefido-sleepace(:8083)"); fi
    if pgrep -f "wisefido-cardagg" >/dev/null 2>&1; then any_running=true; running_names+=("wisefido-cardagg"); fi
    if pgrep -f "wisefido-ai" >/dev/null 2>&1; then any_running=true; running_names+=("wisefido-ai"); fi

    if [ "$any_running" = true ]; then
        echo -e "${YELLOW}Warning: Services are already running${NC}"
        for n in "${running_names[@]}"; do echo -e "${YELLOW}  - $n${NC}"; done
        echo ""
        echo -e "${BLUE}Options:${NC}"
        echo "  1) Stop existing services and restart"
        echo "  2) Exit (keep existing services running)"
        echo ""
        read -p "Choose an option (1 or 2): " -n 1 -r
        echo ""
        if [[ $REPLY =~ ^[1]$ ]]; then
            echo -e "${YELLOW}Stopping existing services...${NC}"
            bash "$SCRIPT_DIR/stop-owlback.sh"
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

# 检查端口
# 注意：wisefido-cardagg 和 wisefido-ai 不需要端口，它们是纯后台服务
check_port 8080 "wisefido-data"
check_port 8081 "wisefido-qinglan"
check_port 8083 "wisefido-sleepace"
check_port 8085 "wisefido-iot"

# 加载统一配置文件
# 获取脚本所在目录（owlBack 根目录）
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/.env"

if [ -f "$ENV_FILE" ]; then
    echo -e "${BLUE}Loading configuration from: $ENV_FILE${NC}"
    set -a  # 自动导出所有变量
    source "$ENV_FILE"
    set +a  # 关闭自动导出
else
    echo -e "${YELLOW}Warning: .env file not found, using default values${NC}"
    # 设置默认环境变量（如果 .env 不存在）
    export DB_HOST="${DB_HOST:-127.0.0.1}"
    export DB_PORT="${DB_PORT:-5432}"
    export DB_USER="${DB_USER:-postgres}"
    export DB_PASSWORD="${DB_PASSWORD:-postgres}"
    export DB_NAME="${DB_NAME:-owlrd}"
    export DB_SSLMODE="${DB_SSLMODE:-disable}"
    export REDIS_ADDR="${REDIS_ADDR:-127.0.0.1:6379}"
    export REDIS_PASSWORD="${REDIS_PASSWORD:-TeLunSu-36kr}"
    export MQTT_BROKER="${MQTT_BROKER:-127.0.0.1}"
    export MQTT_PORT="${MQTT_PORT:-1883}"
fi

# 设置默认值（如果环境变量未设置）
export DB_HOST="${DB_HOST:-127.0.0.1}"
export DB_PORT="${DB_PORT:-5432}"
export DB_USER="${DB_USER:-postgres}"
export DB_PASSWORD="${DB_PASSWORD:-postgres}"
export DB_NAME="${DB_NAME:-owlrd}"
export DB_SSLMODE="${DB_SSLMODE:-disable}"

# 使用 127.0.0.1 而不是 localhost 以避免 IPv6 解析问题
export REDIS_ADDR="${REDIS_ADDR:-127.0.0.1:6379}"
export REDIS_PASSWORD="${REDIS_PASSWORD:-TeLunSu-36kr}"

# MQTT配置（qinglan 用 MQTT_BROKER+MQTT_PORT 拼接，sleepace 用 MQTT_BROKER 含端口，故全局只设 host，sleepace 启动行单独传 host:port）
export MQTT_BROKER="${MQTT_BROKER:-127.0.0.1}"
export MQTT_PORT="${MQTT_PORT:-1883}"

# wisefido-cardagg 配置（如果没有从 .env 加载）
export TENANT_ID="${TENANT_ID:-bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c}"
export CARD_TRIGGER_MODE="${CARD_TRIGGER_MODE:-polling}"
export CARD_POLLING_INTERVAL="${CARD_POLLING_INTERVAL:-86400}"
export CARD_AGGREGATION_ENABLED="${CARD_AGGREGATION_ENABLED:-true}"
export CARD_AGGREGATION_INTERVAL="${CARD_AGGREGATION_INTERVAL:-2}"

# 日志配置（如果未从 .env 加载）
export LOG_LEVEL="${LOG_LEVEL:-info}"
export LOG_FORMAT="${LOG_FORMAT:-json}"

# 显示配置
echo -e "${BLUE}Configuration:${NC}"
echo "  DB_HOST: $DB_HOST"
echo "  DB_NAME: $DB_NAME"
echo "  REDIS_ADDR: $REDIS_ADDR"
echo "  MQTT_BROKER: $MQTT_BROKER:$MQTT_PORT"
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
echo "  - wisefido-cardagg:"
echo "    * Full card creation on startup (in startPollingMode)"
echo "    * Data aggregation on startup + every ${CARD_AGGREGATION_INTERVAL}s (in startDataAggregation)"
echo ""
echo -e "${BLUE}Log files:${NC}"
echo "  wisefido-data: $DATA_LOG"
echo "  wisefido-cardagg: $AGGREGATOR_LOG"
echo "  wisefido-qinglan: $QINGLAN_LOG"
echo "  wisefido-sleepace: $SLEEPACE_LOG"
echo "  wisefido-iot: $IOT_LOG"
echo "  wisefido-ai: $AI_LOG"
echo "  combined: $COMBINED_LOG"
echo ""

# 启动依赖服务（PostgreSQL, Redis, MQTT）
echo -e "${BLUE}Starting dependency services (PostgreSQL, Redis, MQTT)...${NC}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
if [ -f "${SCRIPT_DIR}/docker-compose.yml" ]; then
    cd "$SCRIPT_DIR"
    docker-compose up -d postgresql redis mqtt sleepace-mysql 2>&1 | grep -v "is up-to-date" || true
    echo -e "${GREEN}Dependency services started${NC}"
    echo ""
fi

# 清理旧日志（可选）
if [ "$1" == "--clean" ]; then
    echo -e "${YELLOW}Cleaning old log files...${NC}"
    rm -f "$DATA_LOG" "$AGGREGATOR_LOG" "$QINGLAN_LOG" "$SLEEPACE_LOG" "$IOT_LOG" "$AI_LOG" "$COMBINED_LOG"
fi

# 函数：清理后台进程
cleanup() {
    echo ""
    echo -e "${YELLOW}Shutting down services...${NC}"
    pkill -f "go run.*wisefido-data" || true
    pkill -f "go run.*wisefido-cardagg" || true
    pkill -f "go run.*wisefido-sleepace" || true
    pkill -f "go run.*wisefido-iot" || true
    pkill -f "go run.*wisefido-ai" || true
    pkill -f "wisefido-data" || true
    pkill -f "wisefido-cardagg" || true
    pkill -f "wisefido-sleepace" || true
    pkill -f "wisefido-iot-timeseries" || true
    pkill -f "wisefido-ai" || true
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

if [ ! -d "$OWLBACK_DIR/wisefido-cardagg" ]; then
    echo -e "${RED}Error: wisefido-cardagg directory not found at $OWLBACK_DIR/wisefido-cardagg${NC}"
    exit 1
fi

if [ ! -d "$OWLBACK_DIR/wisefido-qinglan" ]; then
    echo -e "${RED}Error: wisefido-qinglan directory not found at $OWLBACK_DIR/wisefido-qinglan${NC}"
    exit 1
fi

# wisefido-sleepace 可选：目录不存在则跳过（如 Alitest 未拉取或不需要 Sleepad 网关）
SLEEPACE_DIR_MISSING=false
if [ ! -d "$OWLBACK_DIR/wisefido-sleepace" ]; then
    echo -e "${YELLOW}Warning: wisefido-sleepace directory not found at $OWLBACK_DIR/wisefido-sleepace, skipping${NC}"
    SLEEPACE_DIR_MISSING=true
fi

# 启动 wisefido-data 服务
# 功能：数据管理 API + 卡片创建/更新（直接同步调用 CardCreator）
echo -e "${GREEN}[1/6] Starting wisefido-data service...${NC}"
echo -e "${BLUE}  Function: Data management API + Card creation/update (direct sync)${NC}"
cd "$OWLBACK_DIR/wisefido-data"
# 设置环境变量，包括Qinglan服务地址
export HTTP_ADDR="${HTTP_ADDR:-:8080}"
export QINGLAN_API_BASE_URL="${QINGLAN_API_BASE_URL:-http://localhost:8081}"
# 直接输出到终端，同时使用 tee 写入日志文件
go run cmd/wisefido-data/main.go 2>&1 | tee "$DATA_LOG" &
DATA_PID=$!
echo "  PID: $DATA_PID"
echo "  Log: $DATA_LOG (also displayed in terminal)"

# 等待一下确保服务启动
sleep 2

# 启动 wisefido-cardagg 服务
# 功能：数据聚合（从 PostgreSQL + Redis 聚合卡片数据并缓存）
# 注意：卡片创建/更新现在由 wisefido-data 直接处理，aggregator 主要用于数据聚合
echo -e "${GREEN}[2/6] Starting wisefido-cardagg service...${NC}"
echo -e "${BLUE}  Function: Data aggregation (PostgreSQL + Redis → full card cache)${NC}"
echo -e "${YELLOW}  Note: Card creation/update is now handled by wisefido-data${NC}"
cd "$OWLBACK_DIR/wisefido-cardagg"
# 直接输出到终端，同时使用 tee 写入日志文件
go run main.go 2>&1 | tee "$AGGREGATOR_LOG" &
AGGREGATOR_PID=$!
echo "  PID: $AGGREGATOR_PID"
echo "  Log: $AGGREGATOR_LOG (also displayed in terminal)"

# 等待一下确保服务启动
sleep 2

# 启动 wisefido-qinglan 服务
# 功能：设备接入与HTTPS认证 + MQTT消息路由（设备房间消费集成）
echo -e "${GREEN}[3/6] Starting wisefido-qinglan service...${NC}"
echo -e "${BLUE}  Function: Device access + MQTT message routing (internal control)${NC}"
cd "$OWLBACK_DIR/wisefido-qinglan"
# 设置一些关键环境变量（简化配置）
export HTTP_HOST="${HTTP_HOST:-0.0.0.0}"
export HTTP_PORT="${HTTP_PORT:-8081}"
export QINGLAN_HTTPS_PORT="${QINGLAN_HTTPS_PORT:-8443}"
# .env 中 RADAR_HTTPS_* 会传给 qinglan 作为 QINGLAN_HTTPS_*（未设则用 qinglan 默认 owl-common 证书）
export QINGLAN_HTTPS_CERT_FILE="${QINGLAN_HTTPS_CERT_FILE:-$RADAR_HTTPS_CERT_FILE}"
export QINGLAN_HTTPS_KEY_FILE="${QINGLAN_HTTPS_KEY_FILE:-$RADAR_HTTPS_KEY_FILE}"
# 直接输出到终端，同时使用 tee 写入日志文件
go run cmd/wisefido-qinglan/main.go 2>&1 | tee "$QINGLAN_LOG" &
QINGLAN_PID=$!
echo "  PID: $QINGLAN_PID"
echo "  Log: $QINGLAN_LOG (also displayed in terminal)"

# 等待一下确保服务启动
sleep 2

# 启动 wisefido-sleepace 服务（目录存在时）
# 功能：Sleepad 设备网关（MQTT 消费 + HTTP 透传代理）
if [ "$SLEEPACE_DIR_MISSING" != "true" ]; then
  echo -e "${GREEN}[4/6] Starting wisefido-sleepace service...${NC}"
  echo -e "${BLUE}  Function: Sleepad device gateway (MQTT consumer + HTTP proxy)${NC}"
  cd "$OWLBACK_DIR/wisefido-sleepace"
  MQTT_CLIENT_ID=wisefido-sleepace-2 go run cmd/wisefido-sleepace/main.go -env dev 2>&1 | tee "$SLEEPACE_LOG" &
  SLEEPACE_PID=$!
  echo "  PID: $SLEEPACE_PID"
  echo "  Log: $SLEEPACE_LOG (also displayed in terminal)"
else
  echo -e "${YELLOW}[4/6] Skipping wisefido-sleepace (directory not found)${NC}"
  SLEEPACE_PID=""
fi

# 等待一下确保服务启动
sleep 2

# 启动 wisefido-iot 服务
# 功能：从 Redis Streams 消费数据，存储到 TimescaleDB
echo -e "${GREEN}[5/6] Starting wisefido-iot service...${NC}"
echo -e "${BLUE}  Function: Consume data from Redis Streams → TimescaleDB${NC}"
cd "$OWLBACK_DIR/wisefido-iot"
# 设置环境变量（如果未从 .env 加载）
export HTTP_ADDR="${HTTP_ADDR:-:8085}"
export REDIS_DB="${REDIS_DB:-0}"
# Stream 配置 - 统一 IoT streams（所有设备类型）
export STREAM_IOT_MONITOR="${STREAM_IOT_MONITOR:-iot:monitor:stream}"
export STREAM_IOT_STAT="${STREAM_IOT_STAT:-iot:stat:stream}"
export STREAM_IOT_EVENT="${STREAM_IOT_EVENT:-iot:event:stream}"
export STREAM_IOT_ALARM="${STREAM_IOT_ALARM:-iot:alarm:stream}"
export STREAM_IOT_AUTH="${STREAM_IOT_AUTH:-iot:auth:stream}"
# 兼容性配置（旧名称，指向统一的 stream）
export STREAM_RADAR_MONITOR="${STREAM_RADAR_MONITOR:-iot:monitor:stream}"
export STREAM_RADAR_STAT="${STREAM_RADAR_STAT:-iot:stat:stream}"
export STREAM_RADAR_EVENT="${STREAM_RADAR_EVENT:-iot:event:stream}"
export STREAM_RADAR_ALARM="${STREAM_RADAR_ALARM:-iot:alarm:stream}"
export STREAM_RADAR_AUTH="${STREAM_RADAR_AUTH:-iot:auth:stream}"
export STREAM_SLEEPACE_MONITOR="${STREAM_SLEEPACE_MONITOR:-iot:monitor:stream}"
export STREAM_SLEEPACE_STAT="${STREAM_SLEEPACE_STAT:-iot:stat:stream}"
export STREAM_SLEEPACE_EVENT="${STREAM_SLEEPACE_EVENT:-iot:event:stream}"
export STREAM_SLEEPACE_ALARM="${STREAM_SLEEPACE_ALARM:-iot:alarm:stream}"
export STREAM_SLEEPACE_AUTH="${STREAM_SLEEPACE_AUTH:-iot:auth:stream}"
export CONSUMER_GROUP="${CONSUMER_GROUP:-iot-timeseries-group}"
export CONSUMER_NAME="${CONSUMER_NAME:-iot-timeseries-1}"
# 直接输出到终端，同时使用 tee 写入日志文件
go run cmd/wisefido-iot/main.go 2>&1 | tee "$IOT_LOG" &
IOT_PID=$!
echo "  PID: $IOT_PID"
echo "  Log: $IOT_LOG (also displayed in terminal)"

# 等待一下确保服务启动
sleep 2

# 启动 wisefido-ai 服务
# 功能：AI 智能推理服务（高级推理、访客识别、巡房优化）
echo -e "${GREEN}[6/6] Starting wisefido-ai service...${NC}"
echo -e "${BLUE}  Function: AI inference (advanced reasoning, visitor detection, patrol optimization)${NC}"
cd "$OWLBACK_DIR/wisefido-ai"
# 设置环境变量（如果未从 .env 加载）
export REDIS_DB="${REDIS_DB:-0}"
# 直接输出到终端，同时使用 tee 写入日志文件
go run cmd/wisefido-ai/main.go 2>&1 | tee "$AI_LOG" &
AI_PID=$!
echo "  PID: $AI_PID"
echo "  Log: $AI_LOG (also displayed in terminal)"

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}All backend services are running${NC}"
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
        echo "  Stopping wisefido-cardagg (PID: $AGGREGATOR_PID)"
        kill $AGGREGATOR_PID 2>/dev/null || true
    fi
    if [ -n "$QINGLAN_PID" ] && ps -p $QINGLAN_PID > /dev/null 2>&1; then
        echo "  Stopping wisefido-qinglan (PID: $QINGLAN_PID)"
        kill $QINGLAN_PID 2>/dev/null || true
    fi
    if [ -n "$SLEEPACE_PID" ] && ps -p $SLEEPACE_PID > /dev/null 2>&1; then
        echo "  Stopping wisefido-sleepace (PID: $SLEEPACE_PID)"
        kill $SLEEPACE_PID 2>/dev/null || true
    fi
    if [ -n "$IOT_PID" ] && ps -p $IOT_PID > /dev/null 2>&1; then
        echo "  Stopping wisefido-iot (PID: $IOT_PID)"
        kill $IOT_PID 2>/dev/null || true
    fi
    if [ -n "$AI_PID" ] && ps -p $AI_PID > /dev/null 2>&1; then
        echo "  Stopping wisefido-ai (PID: $AI_PID)"
        kill $AI_PID 2>/dev/null || true
    fi
    # 等待进程退出
    sleep 2
    # 强制杀死仍在运行的进程
    pkill -f "go run.*wisefido-data" 2>/dev/null || true
    pkill -f "go run.*wisefido-cardagg" 2>/dev/null || true
    pkill -f "go run.*wisefido-qinglan" 2>/dev/null || true
    pkill -f "go run.*wisefido-sleepace" 2>/dev/null || true
    pkill -f "go run.*wisefido-iot" 2>/dev/null || true
    pkill -f "go run.*wisefido-ai" 2>/dev/null || true
    echo -e "${GREEN}Cleanup completed${NC}"
}

# 注册清理函数（在脚本退出时执行）
trap cleanup EXIT INT TERM

# 检查服务是否正常运行
DATA_FAILED=false
AGGREGATOR_FAILED=false
QINGLAN_FAILED=false
SLEEPACE_FAILED=false
IOT_FAILED=false
AI_FAILED=false

if ! ps -p $DATA_PID > /dev/null 2>&1; then
    echo -e "${RED}Error: wisefido-data service failed to start${NC}"
    echo "Check log: $DATA_LOG"
    tail -20 "$DATA_LOG"
    DATA_FAILED=true
fi

if ! ps -p $AGGREGATOR_PID > /dev/null 2>&1; then
    echo -e "${RED}Error: wisefido-cardagg service failed to start${NC}"
    echo "Check log: $AGGREGATOR_LOG"
    tail -20 "$AGGREGATOR_LOG"
    AGGREGATOR_FAILED=true
fi

if ! ps -p $QINGLAN_PID > /dev/null 2>&1; then
    echo -e "${RED}Error: wisefido-qinglan service failed to start${NC}"
    echo "Check log: $QINGLAN_LOG"
    tail -20 "$QINGLAN_LOG"
    QINGLAN_FAILED=true
fi

if [ -n "$SLEEPACE_PID" ] && ! ps -p $SLEEPACE_PID > /dev/null 2>&1; then
    echo -e "${RED}Error: wisefido-sleepace service failed to start${NC}"
    echo "Check log: $SLEEPACE_LOG"
    tail -20 "$SLEEPACE_LOG"
    SLEEPACE_FAILED=true
fi

if ! ps -p $IOT_PID > /dev/null 2>&1; then
    echo -e "${RED}Error: wisefido-iot service failed to start${NC}"
    echo "Check log: $IOT_LOG"
    tail -20 "$IOT_LOG"
    IOT_FAILED=true
fi

if ! ps -p $AI_PID > /dev/null 2>&1; then
    echo -e "${RED}Error: wisefido-ai service failed to start${NC}"
    echo "Check log: $AI_LOG"
    tail -20 "$AI_LOG"
    AI_FAILED=true
fi

# 统计失败的服务数量
FAILED_COUNT=0
[ "$DATA_FAILED" = true ] && FAILED_COUNT=$((FAILED_COUNT + 1))
[ "$AGGREGATOR_FAILED" = true ] && FAILED_COUNT=$((FAILED_COUNT + 1))
[ "$QINGLAN_FAILED" = true ] && FAILED_COUNT=$((FAILED_COUNT + 1))
[ "$SLEEPACE_FAILED" = true ] && FAILED_COUNT=$((FAILED_COUNT + 1))
[ "$IOT_FAILED" = true ] && FAILED_COUNT=$((FAILED_COUNT + 1))
[ "$AI_FAILED" = true ] && FAILED_COUNT=$((FAILED_COUNT + 1))

# 如果所有服务都失败，退出
if [ $FAILED_COUNT -eq 6 ]; then
    echo -e "${RED}All services failed to start. Exiting.${NC}"
    exit 1
fi

# 如果有服务失败，警告但继续
if [ "$DATA_FAILED" = true ]; then
    echo -e "${YELLOW}Warning: wisefido-data failed, but continuing with other services${NC}"
    echo "  You can check the log: $DATA_LOG"
    echo ""
fi

if [ "$AGGREGATOR_FAILED" = true ]; then
    echo -e "${YELLOW}Warning: wisefido-cardagg failed, but continuing with other services${NC}"
    echo "  You can check the log: $AGGREGATOR_LOG"
    echo ""
fi

if [ "$QINGLAN_FAILED" = true ]; then
    echo -e "${YELLOW}Warning: wisefido-qinglan failed, but continuing with other services${NC}"
    echo "  You can check the log: $QINGLAN_LOG"
    echo ""
fi

if [ "$SLEEPACE_FAILED" = true ]; then
    echo -e "${YELLOW}Warning: wisefido-sleepace failed, but continuing with other services${NC}"
    echo "  You can check the log: $SLEEPACE_LOG"
    echo ""
fi

if [ "$IOT_FAILED" = true ]; then
    echo -e "${YELLOW}Warning: wisefido-iot failed, but continuing with other services${NC}"
    echo "  You can check the log: $IOT_LOG"
    echo ""
fi

if [ "$AI_FAILED" = true ]; then
    echo -e "${YELLOW}Warning: wisefido-ai failed, but continuing with other services${NC}"
    echo "  You can check the log: $AI_LOG"
    echo ""
fi

echo -e "${GREEN}Services started successfully!${NC}"
echo ""
echo -e "${BLUE}Logs are being displayed in this terminal and saved to:${NC}"
echo "  - $DATA_LOG"
echo "  - $AGGREGATOR_LOG"
echo "  - $QINGLAN_LOG"
echo "  - $SLEEPACE_LOG"
echo "  - $IOT_LOG"
echo "  - $AI_LOG"
echo ""
echo ""
echo -e "${YELLOW}Press Ctrl+C to stop all services${NC}"
echo ""

# 等待进程（这样 Ctrl+C 可以正确捕获并停止服务）
# 只等待成功启动的进程
# 注意：由于使用了管道和 tee，$! 可能捕获的是 tee 进程的 PID
# 所以这里使用 pgrep 来查找实际的进程，而不是直接使用 $PID
WAIT_PIDS=""
[ "$DATA_FAILED" = false ] && WAIT_PIDS="$WAIT_PIDS $DATA_PID"
[ "$AGGREGATOR_FAILED" = false ] && WAIT_PIDS="$WAIT_PIDS $AGGREGATOR_PID"
[ "$QINGLAN_FAILED" = false ] && WAIT_PIDS="$WAIT_PIDS $QINGLAN_PID"
[ "$SLEEPACE_FAILED" = false ] && [ -n "$SLEEPACE_PID" ] && WAIT_PIDS="$WAIT_PIDS $SLEEPACE_PID"
[ "$IOT_FAILED" = false ] && WAIT_PIDS="$WAIT_PIDS $IOT_PID"
[ "$AI_FAILED" = false ] && WAIT_PIDS="$WAIT_PIDS $AI_PID"

if [ -n "$WAIT_PIDS" ]; then
    wait $WAIT_PIDS 2>/dev/null || true
fi

