#!/bin/bash

# wisefido-qinglan 启动脚本
# 与现有 owlBack 环境保持一致

cd "$(dirname "$0")"

echo "🚀 Starting wisefido-qinglan service..."
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 检查端口是否被占用
check_port() {
    local port=$1
    local service_name=$2
    
    if ! command -v lsof &> /dev/null; then
        echo -e "${YELLOW}⚠️  lsof not available, skipping port check${NC}"
        return 0
    fi
    
    if lsof -i :$port > /dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  Port $port ($service_name) is already in use${NC}"
        echo "Processes using port $port:"
        lsof -i :$port | tail -n +2
        echo ""
        read -p "Do you want to kill the process(es) using port $port? (Y/n): " answer
        answer=${answer:-Y}
        
        if [[ "$answer" =~ ^[Yy]$ ]]; then
            local pids=$(lsof -ti :$port)
            if [ -n "$pids" ]; then
                echo -e "${YELLOW}Killing processes on port $port...${NC}"
                for pid in $pids; do
                    echo "  Killing PID: $pid"
                    kill -9 $pid 2>/dev/null
                done
                sleep 1
                if lsof -i :$port > /dev/null 2>&1; then
                    echo -e "${RED}❌ Failed to free port $port${NC}"
                    exit 1
                else
                    echo -e "${GREEN}✅ Port $port is now free${NC}"
                fi
            fi
        else
            echo -e "${RED}❌ Port $port is still in use. Exiting.${NC}"
            exit 1
        fi
    else
        echo -e "${GREEN}✅ Port $port ($service_name) is available${NC}"
    fi
}

# 检查依赖服务状态（仅检查，不启动）
check_dependencies() {
    echo ""
    echo "🔍 Checking dependencies..."
    
    # 检查 PostgreSQL（端口 5433，与 start_owlback.sh 保持一致）
    if command -v nc > /dev/null 2>&1; then
        if nc -zv 127.0.0.1 5433 > /dev/null 2>&1; then
            echo -e "${GREEN}✅ PostgreSQL (127.0.0.1:5433) is accessible${NC}"
        else
            echo -e "${RED}❌ PostgreSQL (127.0.0.1:5433) is not accessible${NC}"
            echo "  Please start it: docker-compose up -d postgresql"
            return 1
        fi
    fi
    
    # 检查 Redis
    if command -v nc > /dev/null 2>&1; then
        if nc -zv 127.0.0.1 6379 > /dev/null 2>&1; then
            echo -e "${GREEN}✅ Redis (127.0.0.1:6379) is accessible${NC}"
        else
            echo -e "${RED}❌ Redis (127.0.0.1:6379) is not accessible${NC}"
            echo "  Please start it: docker-compose up -d redis"
            return 1
        fi
    fi
    
    # 检查 MQTT Broker
    if command -v nc > /dev/null 2>&1; then
        if nc -zv 127.0.0.1 1883 > /dev/null 2>&1; then
            echo -e "${GREEN}✅ MQTT Broker (127.0.0.1:1883) is accessible${NC}"
        else
            echo -e "${RED}❌ MQTT Broker (127.0.0.1:1883) is not accessible${NC}"
            echo "  Please start it: docker-compose up -d mqtt"
            return 1
        fi
    fi
    
    return 0
}

# 检查端口冲突
echo "🔍 Checking for port conflicts..."
check_port 8081 "HTTP Server"

# 检查依赖服务
if ! check_dependencies; then
    echo ""
    echo -e "${RED}❌ Dependencies are not ready. Please start them first.${NC}"
    echo ""
    echo "To start dependencies:"
    echo "  cd ../.."
    echo "  docker-compose up -d postgresql redis mqtt"
    echo ""
    exit 1
fi

echo ""
echo "📋 Setting environment variables..."

# 设置环境变量（与 start_owlback.sh 和 start-radar.sh 保持一致）
export DB_HOST=127.0.0.1
export DB_PORT=5433
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=owlrd
export DB_SSLMODE=disable

export REDIS_ADDR=127.0.0.1:6379
export REDIS_PASSWORD=TeLunSu-36kr
export REDIS_DB=0

# MQTT配置（wisefido-qinglan 连接 MQTT）
export MQTT_BROKER=127.0.0.1
export MQTT_PORT=1883
export MQTT_CLIENT_ID=wisefido-qinglan
export MQTT_USERNAME=
export MQTT_PASSWORD=

# 雷达设备MQTT配置（与 wisefido-radar 一致）
export RADAR_MQTT_PREFIX=
export RADAR_MQTT_PRODUCT_ID=88

# HTTP配置
export HTTP_HOST=0.0.0.0
export HTTP_PORT=8081

# 日志配置
export LOG_LEVEL=info
export LOG_FORMAT=json

# 显示配置信息
echo ""
echo -e "${BLUE}📊 Configuration:${NC}"
echo "  🌐 HTTP Server: $HTTP_HOST:$HTTP_PORT"
echo "  🗄️  Database: $DB_HOST:$DB_PORT/$DB_NAME"
echo "  📡 Redis: $REDIS_ADDR"
echo "  📨 MQTT Broker: $MQTT_BROKER:$MQTT_PORT"
echo "  📨 MQTT Product ID: $RADAR_MQTT_PRODUCT_ID"
echo ""

# 启动服务
echo -e "${GREEN}🚀 Starting wisefido-qinglan service...${NC}"
echo ""

# 日志文件路径
LOG_FILE="${QINGLAN_LOG_FILE:-/tmp/wisefido_qinglan.log}"

# 确保日志目录存在
LOG_DIR=$(dirname "$LOG_FILE")
mkdir -p "$LOG_DIR" 2>/dev/null || true

# 记录启动信息到日志
echo "==========================================" >> "$LOG_FILE"
echo "wisefido-qinglan service starting at $(date)" >> "$LOG_FILE"
echo "==========================================" >> "$LOG_FILE"

# 如果通过 nohup 启动（stdout 已重定向），直接运行
# 如果是直接运行，同时输出到控制台和日志
if [ -t 1 ]; then
    # 交互式终端，同时输出到控制台和日志
    go run cmd/wisefido-qinglan/main.go 2>&1 | tee -a "$LOG_FILE"
else
    # 非交互式（nohup），stdout 已重定向，直接运行
    go run cmd/wisefido-qinglan/main.go 2>&1
fi