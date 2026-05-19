#!/bin/bash

# 测试 alarm_events repository 数据库连接
# 使用方法: ./scripts/test_alarm_events_repo.sh

set -e

echo "=========================================="
echo "测试 alarm_events repository 数据库连接"
echo "=========================================="
echo ""

# 检查环境变量
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_NAME=${DB_NAME:-owlrd}

echo "数据库配置:"
echo "  Host: $DB_HOST"
echo "  Port: $DB_PORT"
echo "  User: $DB_USER"
echo "  Database: $DB_NAME"
echo ""

# 检查 PostgreSQL 是否运行
echo "1. 检查 PostgreSQL 连接..."
if ! PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT 1;" > /dev/null 2>&1; then
    echo "❌ PostgreSQL 连接失败"
    echo "   请确保 PostgreSQL 正在运行，并且数据库 '$DB_NAME' 存在"
    exit 1
fi
echo "✅ PostgreSQL 连接成功"
echo ""

# 检查 alarm_events 表是否存在
echo "2. 检查 alarm_events 表..."
if ! PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "\d alarm_events" > /dev/null 2>&1; then
    echo "❌ alarm_events 表不存在"
    echo "   请运行数据库迁移脚本创建表"
    exit 1
fi
echo "✅ alarm_events 表存在"
echo ""

# 检查表结构
echo "3. 检查表结构..."
TABLE_COLS=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'alarm_events';")
if [ "$TABLE_COLS" -lt 18 ]; then
    echo "⚠️  警告: alarm_events 表列数少于预期 (当前: $TABLE_COLS, 预期: >= 18)"
else
    echo "✅ 表结构正常 (列数: $TABLE_COLS)"
fi
echo ""

# 测试查询
echo "4. 测试基本查询..."
QUERY_RESULT=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM alarm_events WHERE metadata->>'deleted_at' IS NULL;")
echo "   当前未删除的报警事件数量: $QUERY_RESULT"
echo ""

# 测试插入（如果表为空，创建一个测试记录）
echo "5. 测试插入功能..."
TEST_TENANT_ID=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT tenant_id FROM tenants LIMIT 1;" | xargs)
if [ -z "$TEST_TENANT_ID" ]; then
    echo "⚠️  警告: 没有找到测试租户，跳过插入测试"
else
    echo "   使用租户 ID: $TEST_TENANT_ID"
    TEST_DEVICE_ID=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT device_id FROM devices WHERE tenant_id = '$TEST_TENANT_ID' LIMIT 1;" | xargs)
    if [ -z "$TEST_DEVICE_ID" ]; then
        echo "⚠️  警告: 没有找到测试设备，跳过插入测试"
    else
        echo "   使用设备 ID: $TEST_DEVICE_ID"
        echo "   (插入测试已跳过，避免污染数据)"
    fi
fi
echo ""

# 测试索引
echo "6. 检查索引..."
INDEX_COUNT=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM pg_indexes WHERE tablename = 'alarm_events';")
echo "   索引数量: $INDEX_COUNT"
if [ "$INDEX_COUNT" -lt 5 ]; then
    echo "⚠️  警告: 索引数量少于预期"
else
    echo "✅ 索引正常"
fi
echo ""

echo "=========================================="
echo "✅ 数据库连接测试完成"
echo "=========================================="
echo ""
echo "下一步: 运行 Go 测试"
echo "  cd /Users/sady3721/project/owlBack/wisefido-alarm"
echo "  go test ./internal/repository -v -run TestAlarm"

