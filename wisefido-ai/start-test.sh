#!/bin/bash

# wisefido-ai 测试启动脚本
# 使用测试配置启动 AI 智能推理服务

set -e

echo "========================================="
echo "Starting wisefido-ai (AI智能推理服务)"
echo "========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 检查必需环境变量
check_env() {
    local var_name=$1
    local var_value=${!var_name}
    
    if [ -z "$var_value" ]; then
        echo -e "${YELLOW}⚠${NC} $var_name 未设置，使用默认值"
        return 1
    else
        echo -e "${GREEN}✓${NC} $var_name: $var_value"
        return 0
    fi
}

echo -e "${BLUE}环境变量检查:${NC}"

# TENANT_ID 可选（不设也能起 wisefido-ai；建业务租户后再填）
if [ -z "$TENANT_ID" ]; then
    echo -e "${YELLOW}⚠${NC} TENANT_ID 未设置（可选）"
else
    check_env "TENANT_ID"
fi

# 数据库配置（使用默认值或环境变量）
export DB_HOST=${DB_HOST:-"127.0.0.1"}
export DB_PORT=${DB_PORT:-"5432"}
export DB_USER=${DB_USER:-"postgres"}
export DB_PASSWORD=${DB_PASSWORD:-"postgres"}
export DB_NAME=${DB_NAME:-"owlrd"}
export DB_SSLMODE=${DB_SSLMODE:-"disable"}

check_env "DB_HOST"
check_env "DB_USER"
check_env "DB_NAME"

# Redis 配置（使用默认值或环境变量）
export REDIS_ADDR=${REDIS_ADDR:-"127.0.0.1:6379"}
export REDIS_PASSWORD=${REDIS_PASSWORD:-"TeLunSu-36kr"}
export REDIS_DB=${REDIS_DB:-0}

check_env "REDIS_ADDR"

# 日志配置
export LOG_LEVEL=${LOG_LEVEL:-"info"}
export LOG_FORMAT=${LOG_FORMAT:-"json"}

echo ""
echo -e "${BLUE}服务配置:${NC}"
echo "  服务名称: wisefido-ai (AI智能推理)"
echo "  功能范围:"
echo "    - 高级推理报警（事件1, 3, 4）"
echo "    - 访客识别和智能分析"
echo "    - 巡房优化和模式识别"
echo "  数据源:"
echo "    - 消费 iot:data:stream"
echo "    - 读取 Redis 缓存（卡片实时数据）"
echo "    - 写入 PostgreSQL（报警事件）"
echo ""

echo -e "${BLUE}启动服务...${NC}"
echo "  按 Ctrl+C 停止服务"
echo ""

# 启动服务
cd "$(dirname "$0")"
BIN_DIR="$PWD/.bin"
mkdir -p "$BIN_DIR"
if go build -o "$BIN_DIR/wisefido-ai" ./cmd/wisefido-ai >/dev/null 2>&1; then
    "$BIN_DIR/wisefido-ai"
else
    echo -e "${YELLOW}Warning: go build failed, falling back to go run${NC}"
    go run cmd/wisefido-ai/main.go
fi