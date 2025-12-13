#!/bin/bash

# wisefido-alarm 服务环境验证脚本

set -e

echo "========================================="
echo "wisefido-alarm 环境验证脚本"
echo "========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查函数
check_command() {
    if command -v $1 &> /dev/null; then
        echo -e "${GREEN}✓${NC} $1 已安装"
        return 0
    else
        echo -e "${RED}✗${NC} $1 未安装"
        return 1
    fi
}

check_service() {
    if pgrep -x "$1" > /dev/null; then
        echo -e "${GREEN}✓${NC} $1 正在运行"
        return 0
    else
        echo -e "${YELLOW}⚠${NC} $1 未运行"
        return 1
    fi
}

# 1. 检查必需命令
echo "1. 检查必需命令..."
check_command "go" || exit 1
check_command "psql" || exit 1
check_command "redis-cli" || exit 1
echo ""

# 2. 检查 PostgreSQL
echo "2. 检查 PostgreSQL 连接..."
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_NAME=${DB_NAME:-owlrd}

if PGPASSWORD=${DB_PASSWORD:-postgres} psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT 1;" > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} PostgreSQL 连接成功"
else
    echo -e "${RED}✗${NC} PostgreSQL 连接失败"
    echo "   请检查环境变量: DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME"
    exit 1
fi
echo ""

# 3. 检查 Redis
echo "3. 检查 Redis 连接..."
REDIS_ADDR=${REDIS_ADDR:-localhost:6379}
REDIS_HOST=$(echo $REDIS_ADDR | cut -d: -f1)
REDIS_PORT=$(echo $REDIS_ADDR | cut -d: -f2)

if redis-cli -h $REDIS_HOST -p $REDIS_PORT ping > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Redis 连接成功"
else
    echo -e "${RED}✗${NC} Redis 连接失败"
    echo "   请检查环境变量: REDIS_ADDR, REDIS_PASSWORD"
    exit 1
fi
echo ""

# 4. 检查数据库表
echo "4. 检查数据库表..."
TABLES=("cards" "alarm_cloud" "alarm_device" "alarm_events" "devices" "rooms" "units")

for table in "${TABLES[@]}"; do
    if PGPASSWORD=${DB_PASSWORD:-postgres} psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "\d $table" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} 表 $table 存在"
    else
        echo -e "${RED}✗${NC} 表 $table 不存在"
        exit 1
    fi
done
echo ""

# 5. 检查卡片数据
echo "5. 检查卡片数据..."
CARD_COUNT=$(PGPASSWORD=${DB_PASSWORD:-postgres} psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM cards;" | xargs)

if [ "$CARD_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓${NC} 找到 $CARD_COUNT 张卡片"
else
    echo -e "${YELLOW}⚠${NC} 没有找到卡片数据"
    echo "   提示: 请先运行 wisefido-card-aggregator 创建卡片"
fi
echo ""

# 6. 检查 Redis 实时数据缓存
echo "6. 检查 Redis 实时数据缓存..."
REALTIME_KEYS=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT KEYS "vital-focus:card:*:realtime" 2>/dev/null | wc -l | xargs)

if [ "$REALTIME_KEYS" -gt 0 ]; then
    echo -e "${GREEN}✓${NC} 找到 $REALTIME_KEYS 个实时数据缓存"
else
    echo -e "${YELLOW}⚠${NC} 没有找到实时数据缓存"
    echo "   提示: 请先运行 wisefido-sensor-fusion 生成实时数据"
fi
echo ""

# 7. 检查环境变量
echo "7. 检查环境变量..."
if [ -z "$TENANT_ID" ]; then
    echo -e "${YELLOW}⚠${NC} TENANT_ID 未设置"
    echo "   提示: 运行服务前需要设置 TENANT_ID 环境变量"
else
    echo -e "${GREEN}✓${NC} TENANT_ID = $TENANT_ID"
fi
echo ""

# 8. 检查 Go 模块
echo "8. 检查 Go 模块..."
cd "$(dirname "$0")/.."
if go mod verify > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Go 模块验证通过"
else
    echo -e "${YELLOW}⚠${NC} Go 模块验证失败，尝试运行 go mod tidy"
    go mod tidy
fi
echo ""

# 9. 检查编译
echo "9. 检查编译..."
if go build ./... > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} 代码编译通过"
else
    echo -e "${RED}✗${NC} 代码编译失败"
    echo "   请运行: go build ./..."
    exit 1
fi
echo ""

echo "========================================="
echo -e "${GREEN}环境验证完成！${NC}"
echo "========================================="
echo ""
echo "下一步："
echo "1. 设置环境变量: export TENANT_ID=your-tenant-id"
echo "2. 运行服务: go run cmd/wisefido-alarm/main.go"
echo ""

