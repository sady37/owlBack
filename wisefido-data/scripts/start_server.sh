#!/bin/bash

# 启动 wisefido-data 服务脚本

echo "=========================================="
echo "启动 wisefido-data 服务"
echo "=========================================="

# 设置默认配置（可通过环境变量覆盖）
export HTTP_ADDR="${HTTP_ADDR:-:8080}"
export DB_ENABLED="${DB_ENABLED:-true}"
export DB_HOST="${DB_HOST:-localhost}"
export DB_PORT="${DB_PORT:-5432}"
export DB_USER="${DB_USER:-postgres}"
export DB_PASSWORD="${DB_PASSWORD:-postgres}"
export DB_NAME="${DB_NAME:-owlrd}"
export REDIS_ADDR="${REDIS_ADDR:-localhost:6379}"
export LOG_LEVEL="${LOG_LEVEL:-info}"

echo "配置:"
echo "  HTTP_ADDR: $HTTP_ADDR"
echo "  DB_ENABLED: $DB_ENABLED"
echo "  DB_HOST: $DB_HOST"
echo "  DB_PORT: $DB_PORT"
echo "  DB_NAME: $DB_NAME"
echo "  REDIS_ADDR: $REDIS_ADDR"
echo ""

# 检查数据库连接（如果启用）
if [ "$DB_ENABLED" = "true" ]; then
  echo "检查数据库连接..."
  # 这里可以添加数据库连接检查
fi

# 启动服务
echo "启动服务..."
cd "$(dirname "$0")/.."
go run cmd/wisefido-data/main.go

