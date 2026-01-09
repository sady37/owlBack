#!/bin/bash

# Radar 服务启动脚本
# 
# 职责：
# 1. 检查端口冲突（8443: HTTPS 服务）
# 2. 检查依赖服务状态（PostgreSQL, Redis, MQTT）- 仅检查，不启动
# 3. 设置环境变量
# 4. 启动 wisefido-radar 服务
#
# 注意：
# - 基础服务（PostgreSQL, Redis, MQTT）应由 start_all_services.sh 或 docker-compose.yml 管理
# - 此脚本只负责启动 wisefido-radar 服务本身

cd "$(dirname "$0")"

echo "🚀 Starting wisefido-radar service..."
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 检查端口是否被占用（仅检查 HTTPS 端口 8443）
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
    
    # 检查 PostgreSQL
    if command -v nc > /dev/null 2>&1; then
        if nc -zv 127.0.0.1 5432 > /dev/null 2>&1; then
            echo -e "${GREEN}✅ PostgreSQL (127.0.0.1:5432) is accessible${NC}"
        else
            echo -e "${RED}❌ PostgreSQL (127.0.0.1:5432) is not accessible${NC}"
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
    
    # 检查 MQTT Broker（端口 1883 可能被 Docker 容器使用，这是正常的）
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

# 检查端口冲突（仅检查 HTTPS 端口 8443）
echo "🔍 Checking for port conflicts..."
check_port 8443 "HTTPS Server"

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

# 设置环境变量
export RADAR_HTTPS_PORT=8443

# 证书文件配置
# 如果证书文件不存在，运行 ./generate-cert.sh 生成自签名证书
CERT_DIR="${RADAR_CERT_DIR:-.}"
CERT_FILE="${RADAR_HTTPS_CERT_FILE:-$CERT_DIR/server.crt}"
KEY_FILE="${RADAR_HTTPS_KEY_FILE:-$CERT_DIR/server.key}"

# 检查证书文件是否存在
if [ ! -f "$CERT_FILE" ] || [ ! -f "$KEY_FILE" ]; then
    echo "⚠️  Warning: Certificate files not found!"
    echo "   Certificate: $CERT_FILE"
    echo "   Key: $KEY_FILE"
    echo ""
    echo "💡 To generate self-signed certificates, run:"
    echo "   ./generate-cert.sh"
    echo ""
    echo "   Or set environment variables:"
    echo "   export RADAR_HTTPS_CERT_FILE=/path/to/server.crt"
    echo "   export RADAR_HTTPS_KEY_FILE=/path/to/server.key"
    echo ""
    echo "⚠️  Starting without TLS (HTTP only - not recommended for production)"
    echo ""
else
    export RADAR_HTTPS_CERT_FILE="$CERT_FILE"
    export RADAR_HTTPS_KEY_FILE="$KEY_FILE"
    echo "✅ Using TLS certificates:"
    echo "   Certificate: $CERT_FILE"
    echo "   Key: $KEY_FILE"
    echo ""
fi

# 数据库配置（使用 127.0.0.1 避免 IPv6 问题）
export DB_HOST=127.0.0.1
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=owlrd
export DB_SSLMODE=disable

# Redis 配置（使用 127.0.0.1 避免 IPv6 问题）
export REDIS_ADDR=127.0.0.1:6379
export REDIS_PASSWORD=
export REDIS_DB=0

# MQTT 配置（wisefido-radar 服务连接 MQTT 的配置）
# 使用 127.0.0.1 而不是 localhost 以避免 IPv6 问题
export MQTT_BROKER=tcp://127.0.0.1:1883
export MQTT_CLIENT_ID=wisefido-radar
export MQTT_USERNAME=
export MQTT_PASSWORD=

# MQTT 配置（返回给设备的配置）
export RADAR_MQTT_SERVER=10.0.0.30
export RADAR_MQTT_PORT=1883
export RADAR_MQTT_PROTOCOL=1  # 1=不加密
export RADAR_MQTT_ACCOUNT=wfiot
export RADAR_MQTT_PASSWORD=tt@wf@2025
export RADAR_MQTT_PREFIX=
export RADAR_MQTT_PRODUCT_ID=88
export RADAR_MQTT_TIMEOUT=30
export RADAR_MQTT_KEEPALIVE=60
export RADAR_MQTT_CLIENT_ID_PREFIX=radar

# 日志配置
export LOG_LEVEL=info
export LOG_FORMAT=json

# 显示配置信息
echo ""
echo -e "${BLUE}📊 Configuration:${NC}"
echo "  🔌 HTTPS Port: $RADAR_HTTPS_PORT"
echo "  🗄️  Database: $DB_HOST:$DB_PORT/$DB_NAME"
echo "  📡 Redis: $REDIS_ADDR"
echo "  📨 MQTT Broker: $MQTT_BROKER"
echo "  📨 MQTT (Device): $RADAR_MQTT_SERVER:$RADAR_MQTT_PORT"
echo ""

# 启动服务
echo -e "${GREEN}🚀 Starting wisefido-radar service...${NC}"
echo ""

go run cmd/wisefido-radar/main.go
