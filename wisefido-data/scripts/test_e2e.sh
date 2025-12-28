#!/bin/bash

# 端到端测试快速启动脚本
# 用于快速配置和启动服务进行测试

set -e

echo "=========================================="
echo "端到端测试：Vue 前端 → PostgreSQL 后端"
echo "=========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查 PostgreSQL
echo -e "${YELLOW}检查 PostgreSQL...${NC}"
if command -v psql &> /dev/null; then
    if psql -h localhost -U postgres -d owlrd -c "SELECT 1;" &> /dev/null; then
        echo -e "${GREEN}✓ PostgreSQL 连接成功${NC}"
    else
        echo -e "${RED}✗ PostgreSQL 连接失败${NC}"
        echo "请检查："
        echo "  1. PostgreSQL 是否运行"
        echo "  2. 数据库 'owlrd' 是否存在"
        echo "  3. 用户 'postgres' 的密码是否正确"
        exit 1
    fi
else
    echo -e "${YELLOW}⚠ psql 命令未找到，跳过数据库连接检查${NC}"
fi

# 检查 Redis
echo -e "${YELLOW}检查 Redis...${NC}"
if command -v redis-cli &> /dev/null; then
    if redis-cli ping &> /dev/null; then
        echo -e "${GREEN}✓ Redis 连接成功${NC}"
    else
        echo -e "${RED}✗ Redis 连接失败${NC}"
        echo "请确保 Redis 正在运行: redis-server"
        exit 1
    fi
else
    echo -e "${YELLOW}⚠ redis-cli 命令未找到，跳过 Redis 连接检查${NC}"
fi

# 设置环境变量
echo ""
echo -e "${YELLOW}配置环境变量...${NC}"

# 读取用户输入（可选）
read -p "数据库主机 [localhost]: " DB_HOST
DB_HOST=${DB_HOST:-localhost}

read -p "数据库端口 [5432]: " DB_PORT
DB_PORT=${DB_PORT:-5432}

read -p "数据库用户 [postgres]: " DB_USER
DB_USER=${DB_USER:-postgres}

read -s -p "数据库密码 [postgres]: " DB_PASSWORD
echo ""
DB_PASSWORD=${DB_PASSWORD:-postgres}

read -p "数据库名称 [owlrd]: " DB_NAME
DB_NAME=${DB_NAME:-owlrd}

read -p "HTTP 端口 [:8080]: " HTTP_ADDR
HTTP_ADDR=${HTTP_ADDR:-:8080}

# 导出环境变量
export DB_ENABLED=true
export DB_HOST=$DB_HOST
export DB_PORT=$DB_PORT
export DB_USER=$DB_USER
export DB_PASSWORD=$DB_PASSWORD
export DB_NAME=$DB_NAME
export DB_SSLMODE=disable

export REDIS_ADDR=localhost:6379
export REDIS_PASSWORD=
export HTTP_ADDR=$HTTP_ADDR
export LOG_LEVEL=info
export DOCTOR_ENABLED=true

echo ""
echo -e "${GREEN}配置完成！${NC}"
echo ""
echo "配置摘要:"
echo "  数据库: $DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"
echo "  HTTP: $HTTP_ADDR"
echo "  Redis: localhost:6379"
echo ""

# 切换到项目根目录
cd "$(dirname "$0")/.."

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo -e "${RED}✗ Go 未安装或不在 PATH 中${NC}"
    exit 1
fi

# 启动服务
echo -e "${YELLOW}启动后端服务...${NC}"
echo "按 Ctrl+C 停止服务"
echo ""

go run cmd/wisefido-data/main.go

