#!/bin/bash

# 数据库初始化脚本
# 用途：按顺序执行owlRD/db/目录下的所有SQL文件

set -e

# 配置
DB_NAME="${DB_NAME:-owlrd}"
DB_USER="${DB_USER:-postgres}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
OWLRD_DB_DIR="$PROJECT_ROOT/../owlRD/db"

# 检查owlRD目录是否存在
if [ ! -d "$OWLRD_DB_DIR" ]; then
    echo "Error: owlRD/db directory not found at $OWLRD_DB_DIR"
    echo "Please ensure owlRD is in the project directory"
    exit 1
fi

echo "=========================================="
echo "Database Initialization Script"
echo "=========================================="
echo "Database: $DB_NAME"
echo "User: $DB_USER"
echo "Host: $DB_HOST:$DB_PORT"
echo "SQL Files: $OWLRD_DB_DIR"
echo "=========================================="
echo ""

# 检查数据库连接
echo "Checking database connection..."
if ! PGPASSWORD="${DB_PASSWORD:-postgres}" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" > /dev/null 2>&1; then
    echo "Error: Cannot connect to database"
    echo "Please check your database configuration"
    exit 1
fi
echo "✓ Database connection OK"
echo ""

# 执行SQL文件（按文件名排序）
echo "Executing SQL files..."
for sql_file in $(ls -1 "$OWLRD_DB_DIR"/*.sql | sort); do
    filename=$(basename "$sql_file")
    echo "  → $filename"
    
    if PGPASSWORD="${DB_PASSWORD:-postgres}" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$sql_file" > /dev/null 2>&1; then
        echo "    ✓ Success"
    else
        echo "    ✗ Failed"
        echo "    Please check the error above"
        exit 1
    fi
done

echo ""
echo "=========================================="
echo "Database initialization completed!"
echo "=========================================="

