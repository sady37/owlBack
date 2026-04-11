#!/bin/bash
#
# 启停成对：./start-iot.sh ↔ ./stop-iot.sh
# systemd：systemctl start owlback.iot ↔ systemctl stop owlback.iot
#
# IoT 时序数据服务启动脚本
# 
# 职责：
# 1. 检查端口冲突（8085: HTTP 服务）
# 2. 检查依赖服务状态（PostgreSQL, Redis）- 仅检查，不启动
# 3. 设置环境变量
# 4. 启动 wisefido-iot 服务
#
# 注意：
# - 基础服务（PostgreSQL, Redis）应由 start_all_services.sh 或 docker-compose.yml 管理
# - 此脚本只负责启动 wisefido-iot 服务本身

cd "$(dirname "$0")"

echo "🚀 Starting wisefido-iot service..."
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

# 检查端口冲突
echo "🔍 Checking for port conflicts..."
check_port 8085 "HTTP Server"

# 检查依赖服务状态（仅检查，不启动）
check_dependencies() {
    echo ""
    echo "🔍 Checking dependencies..."
    
    # 检查 PostgreSQL（端口 5432，与 start_owlback.sh 保持一致）
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
    
    return 0
}

# 检查依赖服务
if ! check_dependencies; then
    echo ""
    echo -e "${RED}❌ Dependencies are not ready. Please start them first.${NC}"
    echo ""
    echo "To start dependencies:"
    echo "  cd ../.."
    echo "  docker-compose up -d postgresql redis"
    echo ""
    exit 1
fi

echo ""
echo "📋 Setting environment variables..."

# 数据库配置（使用 127.0.0.1 避免 IPv6 问题，端口 5432 与 start_owlback.sh 保持一致）
export DB_HOST=127.0.0.1
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=owlrd
export DB_SSLMODE=disable

# Redis 配置（使用 127.0.0.1 避免 IPv6 问题）
export REDIS_ADDR=127.0.0.1:6379
export REDIS_PASSWORD=TeLunSu-36kr
export REDIS_DB=0

# Stream 配置 - 设备级别 streams
export STREAM_RADAR_MONITOR=iot:monitor:stream
export STREAM_RADAR_STAT=iot:stat:stream
export STREAM_RADAR_EVENT=iot:event:stream
export STREAM_IOT_ALARM=iot:alarm:stream
//export STREAM_SLEEPACE_MONITOR=sleepace:monitor:stream
//export STREAM_SLEEPACE_EVENT=sleepace:event:stream
# 注意：Sleepace 没有 stat 数据；报警流 Radar/Sleepad 统一 iot:alarm:stream，wisefido-iot 消费 STREAM_IOT_ALARM

# 消费者配置
export CONSUMER_GROUP=iot-timeseries-group
export CONSUMER_NAME=iot-timeseries-1

# 日志配置
export LOG_LEVEL=info
export LOG_FORMAT=json

# 日志目录（与 start_owlback.sh 统一）
OWL_LOG="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")/../.." && pwd)/log"
LOG_DIR="${LOG_DIR:-$OWL_LOG}"
mkdir -p "$LOG_DIR"
LOG_FILE="${IOT_LOG_FILE:-$LOG_DIR/wisefido-iot.log}"

# HTTP 配置
export HTTP_ADDR=:8085

# 显示配置信息
echo ""
echo -e "${BLUE}📊 Configuration:${NC}"
echo "  🔌 HTTP Port: 8085"
echo "  🗄️  Database: $DB_HOST:$DB_PORT/$DB_NAME"
echo "  📡 Redis: $REDIS_ADDR"
echo "  📨 IoT Monitor: $STREAM_IOT_MONITOR"
echo "  📨 IoT Stat: $STREAM_IOT_STAT"
echo "  📨 IoT Event: $STREAM_IOT_EVENT"
echo "  📨 IoT Alarm: $STREAM_IOT_ALARM"
echo "  📨 Sleepace Monitor: $STREAM_SLEEPACE_MONITOR"
echo "  📨 Sleepace Event: $STREAM_SLEEPACE_EVENT"
echo "  👥 Consumer Group: $CONSUMER_GROUP"
echo "  👤 Consumer Name: $CONSUMER_NAME"
echo ""

# 启动服务
echo -e "${GREEN}🚀 Starting wisefido-iot service...${NC}"
echo ""
echo -e "${BLUE}📝 Log: $LOG_FILE${NC}"
echo ""

if [ -t 1 ]; then
    go run cmd/wisefido-iot/main.go 2>&1 | tee -a "$LOG_FILE"
else
    go run cmd/wisefido-iot/main.go 2>&1 | tee -a "$LOG_FILE"
fi
