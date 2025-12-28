#!/bin/bash

# 快速检查数据库连接脚本

set -e

DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_NAME=${DB_NAME:-owlrd}

echo "检查 PostgreSQL 连接..."
echo "  主机: $DB_HOST"
echo "  端口: $DB_PORT"
echo "  用户: $DB_USER"
echo "  数据库: $DB_NAME"
echo ""

if command -v psql &> /dev/null; then
    if psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT version();" &> /dev/null; then
        echo "✓ 数据库连接成功"
        echo ""
        echo "数据库信息:"
        psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT version();" | head -3
        echo ""
        echo "表数量:"
        psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" | tail -1
    else
        echo "✗ 数据库连接失败"
        echo ""
        echo "请检查："
        echo "  1. PostgreSQL 是否运行"
        echo "  2. 数据库 '$DB_NAME' 是否存在"
        echo "  3. 用户 '$DB_USER' 的密码是否正确"
        echo "  4. 连接参数是否正确"
        exit 1
    fi
else
    echo "⚠ psql 命令未找到，无法检查数据库连接"
    echo "请安装 PostgreSQL 客户端工具"
    exit 1
fi

