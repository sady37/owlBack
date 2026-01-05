#!/bin/bash

# wisefido-card-aggregator 服务启动脚本

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting wisefido-card-aggregator service...${NC}"

# 检查环境变量
if [ -z "$TENANT_ID" ]; then
    echo -e "${YELLOW}Warning: TENANT_ID not set, using default demo tenant ID${NC}"
    export TENANT_ID="bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c"
fi

# 设置默认环境变量（如果未设置）
export DB_HOST="${DB_HOST:-localhost}"
export DB_PORT="${DB_PORT:-5432}"
export DB_USER="${DB_USER:-postgres}"
export DB_PASSWORD="${DB_PASSWORD:-postgres}"
export DB_NAME="${DB_NAME:-owlrd}"
export DB_SSLMODE="${DB_SSLMODE:-disable}"

export REDIS_ADDR="${REDIS_ADDR:-localhost:6379}"
export REDIS_PASSWORD="${REDIS_PASSWORD:-}"

export CARD_TRIGGER_MODE="${CARD_TRIGGER_MODE:-polling}"
export CARD_POLLING_INTERVAL="${CARD_POLLING_INTERVAL:-3600}"

export LOG_LEVEL="${LOG_LEVEL:-info}"
export LOG_FORMAT="${LOG_FORMAT:-json}"

# 显示配置
echo -e "${GREEN}Configuration:${NC}"
echo "  TENANT_ID: $TENANT_ID"
echo "  DB_HOST: $DB_HOST"
echo "  DB_NAME: $DB_NAME"
echo "  REDIS_ADDR: $REDIS_ADDR"
echo "  CARD_TRIGGER_MODE: $CARD_TRIGGER_MODE"
echo "  CARD_POLLING_INTERVAL: $CARD_POLLING_INTERVAL"
echo ""

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed or not in PATH${NC}"
    exit 1
fi

# 切换到服务目录
cd "$(dirname "$0")"

# 启动服务
echo -e "${GREEN}Starting service...${NC}"
go run cmd/wisefido-card-aggregator/main.go

