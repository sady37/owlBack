#!/bin/bash

# Sleepace 服务启动脚本
# 
# 职责：
# 1. 检查依赖服务状态（PostgreSQL, Redis, MQTT）- 仅检查，不启动
# 2. 设置环境变量
# 3. 启动 wisefido-sleepace 服务
#
# 注意：
# - 基础服务（PostgreSQL, Redis, MQTT）应由 start_all_services.sh 或 docker-compose.yml 管理
# - 此脚本只负责启动 wisefido-sleepace 服务本身

cd "$(dirname "$0")"

echo "🚀 Starting wisefido-sleepace service..."
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# 数据库配置（使用 127.0.0.1 避免 IPv6 问题，端口 5433 与 start_owlback.sh 保持一致）
export DB_HOST=127.0.0.1
export DB_PORT=5433
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=owlrd
export DB_SSLMODE=disable

# Redis 配置（使用 127.0.0.1 避免 IPv6 问题）
export REDIS_ADDR=127.0.0.1:6379
export REDIS_PASSWORD=TeLunSu-36kr
export REDIS_DB=0

# MQTT 配置（wisefido-sleepace 服务连接 MQTT 的配置）
# 使用 127.0.0.1 而不是 localhost 以避免 IPv6 问题
export MQTT_BROKER=mqtt://127.0.0.1:1883
export MQTT_CLIENT_ID=wisefido-sleepace
export MQTT_USERNAME=
export MQTT_PASSWORD=

# Sleepace 配置（Sleepace 厂家程序在同一台机器上，使用内部地址）
export SLEEPACE_HTTP_ADDRESS=http://127.0.0.1:8080
export SLEEPACE_APP_ID=
export SLEEPACE_CHANNEL_ID=
export SLEEPACE_SECRET_KEY=
export SLEEPACE_TIMEZONE=8
export SLEEPACE_MQTT_TOPIC=sleepace-57136

# 日志配置
export LOG_LEVEL=info
export LOG_FORMAT=json

# 显示配置信息
echo ""
echo -e "${BLUE}📊 Configuration:${NC}"
echo "  🗄️  Database: $DB_HOST:$DB_PORT/$DB_NAME"
echo "  📡 Redis: $REDIS_ADDR"
echo "  📨 MQTT Broker: $MQTT_BROKER"
echo "  📨 MQTT Topic: $SLEEPACE_MQTT_TOPIC"
echo "  🌐 Sleepace HTTP: $SLEEPACE_HTTP_ADDRESS"
  echo "  📨 Data Streams: sleepace:monitor:stream, sleepace:event:stream, sleepace:alarm:stream"
echo ""

# 启动服务
echo -e "${GREEN}🚀 Starting wisefido-sleepace service...${NC}"
echo ""

go run cmd/wisefido-sleepace/main.go
