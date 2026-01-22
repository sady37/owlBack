#!/bin/bash

# wisefido-card-aggregator 服务启动脚本

set -e

# 切换到脚本所在目录
cd "$(dirname "$0")"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Starting wisefido-card-aggregator service${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 检查服务是否已在运行
if pgrep -f "go run.*wisefido-card-aggregator" > /dev/null 2>&1 || \
   pgrep -f "wisefido-card-aggregator" > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Warning: wisefido-card-aggregator is already running${NC}"
    echo -e "${YELLOW}   Use stop_cardagg.sh to stop it first${NC}"
    exit 1
fi

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Error: Go is not installed or not in PATH${NC}"
    exit 1
fi

# 设置默认环境变量（如果未设置）
# 使用 127.0.0.1 而不是 localhost 以避免 IPv6 解析问题
# 注意：如果使用 docker-compose，PostgreSQL 端口映射为 5433:5432，所以从宿主机连接应使用 5433
export DB_HOST="${DB_HOST:-127.0.0.1}"
export DB_PORT="${DB_PORT:-5433}"
export DB_USER="${DB_USER:-postgres}"
export DB_PASSWORD="${DB_PASSWORD:-postgres}"
export DB_NAME="${DB_NAME:-owlrd}"
export DB_SSLMODE="${DB_SSLMODE:-disable}"

# 使用 127.0.0.1 而不是 localhost 以避免 IPv6 解析问题
export REDIS_ADDR="${REDIS_ADDR:-127.0.0.1:6379}"
# Redis 密码（根据 docker-compose.yml 配置）
export REDIS_PASSWORD="${REDIS_PASSWORD:-TeLunSu-36kr}"

export CARD_TRIGGER_MODE="${CARD_TRIGGER_MODE:-polling}"
export CARD_POLLING_INTERVAL="${CARD_POLLING_INTERVAL:-3600}"

export LOG_LEVEL="${LOG_LEVEL:-info}"
export LOG_FORMAT="${LOG_FORMAT:-json}"

# 显示配置
echo -e "${BLUE}📋 Configuration:${NC}"
echo "  DB_HOST: $DB_HOST"
echo "  DB_PORT: $DB_PORT"
echo "  DB_NAME: $DB_NAME"
echo "  REDIS_ADDR: $REDIS_ADDR"
echo "  CARD_TRIGGER_MODE: $CARD_TRIGGER_MODE"
echo "  CARD_POLLING_INTERVAL: $CARD_POLLING_INTERVAL"
echo "  LOG_LEVEL: $LOG_LEVEL"
echo "  LOG_FORMAT: $LOG_FORMAT"
echo -e "${YELLOW}  Note: Service aggregates cards for all tenants${NC}"
echo ""

# 检查是否在后台运行
if [ "$1" = "--background" ] || [ "$1" = "-b" ]; then
    # 后台运行，输出到日志文件
    LOG_DIR="${LOG_DIR:-/tmp/owlBack_logs}"
    mkdir -p "$LOG_DIR"
    LOG_FILE="$LOG_DIR/wisefido-card-aggregator.log"
    
    echo -e "${BLUE}🚀 Starting service in background...${NC}"
    echo -e "${BLUE}   Log file: $LOG_FILE${NC}"
    
    nohup go run cmd/wisefido-card-aggregator/main.go > "$LOG_FILE" 2>&1 &
    PID=$!
    
    # 等待一下，检查进程是否成功启动
    sleep 2
    if ps -p $PID > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Service started successfully (PID: $PID)${NC}"
        echo -e "${GREEN}   View logs: tail -f $LOG_FILE${NC}"
    else
        echo -e "${RED}❌ Service failed to start${NC}"
        echo -e "${RED}   Check logs: $LOG_FILE${NC}"
        exit 1
    fi
else
    # 前台运行
    echo -e "${BLUE}🚀 Starting service in foreground...${NC}"
    echo -e "${YELLOW}   Press Ctrl+C to stop${NC}"
    echo ""
    
    go run cmd/wisefido-card-aggregator/main.go
fi
