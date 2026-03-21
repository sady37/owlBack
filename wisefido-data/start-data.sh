#!/bin/bash

cd "$(dirname "$0")"

echo "🚀 Starting wisefido-data service..."

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

check_port() {
    local port=$1
    local service_name=$2
    
    if ! command -v lsof &> /dev/null; then
        echo -e "${YELLOW}⚠️  lsof not available, skipping port check${NC}"
        return 0
    fi
    
    if lsof -i :$port > /dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  Port $port ($service_name) is already in use${NC}"
        local pids=$(lsof -ti :$port)
        if [ -n "$pids" ]; then
            echo -e "${YELLOW}Killing processes on port $port...${NC}"
            for pid in $pids; do
                kill -9 $pid 2>/dev/null
            done
            sleep 1
        fi
    else
        echo -e "${GREEN}✅ Port $port ($service_name) is available${NC}"
    fi
}

check_and_start_dependencies() {
    echo ""
    echo "🔍 Checking dependencies..."
    
    SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
    DOCKER_COMPOSE_FILE="${SCRIPT_DIR}/../docker-compose.yml"
    
    if command -v nc > /dev/null 2>&1; then
        if ! nc -zv 127.0.0.1 5432 > /dev/null 2>&1; then
            echo -e "${RED}❌ PostgreSQL (127.0.0.1:5432) is not accessible${NC}"
            return 1
        fi
        if ! nc -zv 127.0.0.1 6379 > /dev/null 2>&1; then
            echo -e "${RED}❌ Redis (127.0.0.1:6379) is not accessible${NC}"
            return 1
        fi
    fi
    
    return 0
}

check_port 8080 "HTTP Server"

if ! check_and_start_dependencies; then
    echo -e "${RED}❌ Dependencies are not ready${NC}"
    exit 1
fi

echo ""
echo "📋 Setting environment variables..."

export HTTP_ADDR="${HTTP_ADDR:-:8080}"

export DB_ENABLED="${DB_ENABLED:-true}"
export DB_HOST="${DB_HOST:-127.0.0.1}"
export DB_PORT="${DB_PORT:-5432}"
export DB_USER="${DB_USER:-postgres}"
export DB_PASSWORD="${DB_PASSWORD:-postgres}"
export DB_NAME="${DB_NAME:-owlrd}"
export DB_SSLMODE="${DB_SSLMODE:-disable}"

export REDIS_ADDR="${REDIS_ADDR:-127.0.0.1:6379}"
export REDIS_PASSWORD="${REDIS_PASSWORD:-TeLunSu-36kr}"
export REDIS_DB=0

export LOG_LEVEL="${LOG_LEVEL:-info}"
export LOG_FORMAT="${LOG_FORMAT:-json}"

# Qinglan 服务配置（用于下发雷达监控设置：工作模式、跌倒/呼吸心率参数）
export QINGLAN_API_BASE_URL="${QINGLAN_API_BASE_URL:-http://localhost:8081}"

# Sleepace 配置（可选）
export SLEEPACE_HTTP_ADDRESS="${SLEEPACE_HTTP_ADDRESS:-http://127.0.0.1:8090}"
export SLEEPACE_APP_ID="${SLEEPACE_APP_ID:-}"
export SLEEPACE_CHANNEL_ID="${SLEEPACE_CHANNEL_ID:-}"
export SLEEPACE_SECRET_KEY="${SLEEPACE_SECRET_KEY:-}"
export SLEEPACE_TIMEZONE="${SLEEPACE_TIMEZONE:-28800}"

# MQTT 配置（用于触发报告下载，默认禁用）
export MQTT_ENABLED="${MQTT_ENABLED:-false}"
export MQTT_BROKER="${MQTT_BROKER:-tcp://localhost:1883}"
export MQTT_CLIENT_ID="${MQTT_CLIENT_ID:-wisefido-data-sleepace}"
export MQTT_USERNAME="${MQTT_USERNAME:-}"
export MQTT_PASSWORD="${MQTT_PASSWORD:-}"
export MQTT_TOPIC="${MQTT_TOPIC:-sleepace-57136}"

# Radar 服务配置（用于调用 wisefido-radar 内部 API）
export RADAR_INTERNAL_API_BASE_URL="${RADAR_INTERNAL_API_BASE_URL:-http://localhost:8443}"

# CardManage 服务配置（用于调用 wisefido-card-manage API）
export CARD_MANAGE_API_BASE_URL="${CARD_MANAGE_API_BASE_URL:-http://localhost:8082}"

# IoTTimeSeries 服务配置（用于调用 wisefido-iot-timeseries 内部 API）
export IOT_TIMESERIES_INTERNAL_API_BASE_URL="${IOT_TIMESERIES_INTERNAL_API_BASE_URL:-http://localhost:8085}"

echo ""
echo -e "${BLUE}📊 Configuration:${NC}"
echo "  🌐 HTTP Server: $HTTP_ADDR"
echo "  🗄️  Database: $DB_HOST:$DB_PORT/$DB_NAME (enabled: $DB_ENABLED)"
echo "  📡 Redis: $REDIS_ADDR"
echo "  🔗 Qinglan API: $QINGLAN_API_BASE_URL"
if [ "$MQTT_ENABLED" = "true" ]; then
    echo "  📨 MQTT Broker: $MQTT_BROKER"
fi
echo ""

echo -e "${GREEN}🚀 Starting wisefido-data service...${NC}"
echo ""

# 日志目录（与 start_owlback.sh 统一）
OWL_LOG="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")/../.." && pwd)/log"
LOG_DIR="${LOG_DIR:-$OWL_LOG}"
mkdir -p "$LOG_DIR"
LOG_FILE="${DATA_LOG_FILE:-$LOG_DIR/wisefido-data.log}"

echo -e "${BLUE}📝 Log Configuration:${NC}"
echo "  📄 Log File: $LOG_FILE"
echo "  📁 Log Directory: $LOG_DIR"
echo ""

# 写入启动信息到日志文件
echo "==========================================" >> "$LOG_FILE"
echo "wisefido-data service starting at $(date)" >> "$LOG_FILE"
echo "Log file: $LOG_FILE" >> "$LOG_FILE"
echo "==========================================" >> "$LOG_FILE"

# 同时输出到控制台和日志文件
if [ -t 1 ]; then
    echo -e "${GREEN}✅ Logging to: $LOG_FILE${NC}"
    echo -e "${GREEN}✅ Output will be displayed in terminal and saved to log file${NC}"
    echo ""
    go run cmd/wisefido-data/main.go 2>&1 | tee -a "$LOG_FILE"
else
    echo "Logging to: $LOG_FILE" >&2
    go run cmd/wisefido-data/main.go 2>&1 | tee -a "$LOG_FILE"
fi
