#!/bin/bash

# 传感器融合功能验证脚本
# 用途：检查环境配置和数据准备情况

set -e

echo "=========================================="
echo "传感器融合功能验证 - 环境检查"
echo "=========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查函数
check_command() {
    if command -v $1 &> /dev/null; then
        echo -e "${GREEN}✅${NC} $1 已安装"
        return 0
    else
        echo -e "${RED}❌${NC} $1 未安装"
        return 1
    fi
}

check_service() {
    if pgrep -x "$1" > /dev/null; then
        echo -e "${GREEN}✅${NC} $1 服务正在运行"
        return 0
    else
        echo -e "${RED}❌${NC} $1 服务未运行"
        return 1
    fi
}

# 1. 检查必需的命令
echo "1. 检查必需的命令..."
check_command "psql" || echo "  提示: 需要安装 PostgreSQL 客户端"
check_command "redis-cli" || echo "  提示: 需要安装 Redis 客户端"
check_command "go" || echo "  提示: 需要安装 Go"
echo ""

# 2. 检查服务状态
echo "2. 检查服务状态..."
check_service "postgres" || echo "  提示: 启动 PostgreSQL: brew services start postgresql"
check_service "redis-server" || echo "  提示: 启动 Redis: brew services start redis"
echo ""

# 3. 检查数据库连接
echo "3. 检查数据库连接..."
DB_HOST=${DB_HOST:-localhost}
DB_USER=${DB_USER:-postgres}
DB_NAME=${DB_NAME:-owlrd}

if psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" &> /dev/null; then
    echo -e "${GREEN}✅${NC} 数据库连接成功"
    
    # 检查 cards 表
    CARD_COUNT=$(psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM cards;" 2>/dev/null | xargs)
    if [ "$CARD_COUNT" -gt 0 ]; then
        echo -e "${GREEN}✅${NC} cards 表有数据: $CARD_COUNT 条记录"
    else
        echo -e "${YELLOW}⚠️${NC} cards 表为空，需要运行 wisefido-card-aggregator 创建卡片"
    fi
    
    # 检查设备数据
    DEVICE_COUNT=$(psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM devices WHERE monitoring_enabled = TRUE;" 2>/dev/null | xargs)
    if [ "$DEVICE_COUNT" -gt 0 ]; then
        echo -e "${GREEN}✅${NC} 有 $DEVICE_COUNT 个启用的设备"
    else
        echo -e "${YELLOW}⚠️${NC} 没有启用的设备"
    fi
    
    # 检查 iot_timeseries 表数据
    DATA_COUNT=$(psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM iot_timeseries;" 2>/dev/null | xargs)
    if [ "$DATA_COUNT" -gt 0 ]; then
        echo -e "${GREEN}✅${NC} iot_timeseries 表有数据: $DATA_COUNT 条记录"
    else
        echo -e "${YELLOW}⚠️${NC} iot_timeseries 表为空，需要设备数据"
    fi
else
    echo -e "${RED}❌${NC} 数据库连接失败"
    echo "  提示: 检查环境变量 DB_HOST, DB_USER, DB_PASSWORD, DB_NAME"
fi
echo ""

# 4. 检查 Redis 连接
echo "4. 检查 Redis 连接..."
REDIS_ADDR=${REDIS_ADDR:-localhost:6379}
REDIS_HOST=$(echo $REDIS_ADDR | cut -d: -f1)
REDIS_PORT=$(echo $REDIS_ADDR | cut -d: -f2)

if redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" ping &> /dev/null; then
    echo -e "${GREEN}✅${NC} Redis 连接成功"
    
    # 检查 iot:data:stream
    STREAM_LENGTH=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" XLEN iot:data:stream 2>/dev/null | xargs)
    if [ "$STREAM_LENGTH" -gt 0 ]; then
        echo -e "${GREEN}✅${NC} iot:data:stream 有数据: $STREAM_LENGTH 条消息"
    else
        echo -e "${YELLOW}⚠️${NC} iot:data:stream 为空，需要设备数据"
    fi
    
    # 检查缓存键
    CACHE_KEYS=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" KEYS "vital-focus:card:*:realtime" 2>/dev/null | wc -l | xargs)
    if [ "$CACHE_KEYS" -gt 0 ]; then
        echo -e "${GREEN}✅${NC} 有 $CACHE_KEYS 个实时数据缓存键"
    else
        echo -e "${YELLOW}⚠️${NC} 没有实时数据缓存，服务运行后会创建"
    fi
else
    echo -e "${RED}❌${NC} Redis 连接失败"
    echo "  提示: 检查环境变量 REDIS_ADDR"
fi
echo ""

# 5. 总结
echo "=========================================="
echo "检查完成"
echo "=========================================="
echo ""
echo "下一步："
echo "1. 如果所有检查通过，可以运行服务:"
echo "   cd /Users/sady3721/project/owlBack/wisefido-sensor-fusion"
echo "   go run cmd/wisefido-sensor-fusion/main.go"
echo ""
echo "2. 如果 cards 表为空，先运行:"
echo "   cd /Users/sady3721/project/owlBack/wisefido-card-aggregator"
echo "   go run cmd/wisefido-card-aggregator/main.go"
echo ""

