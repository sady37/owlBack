#!/bin/bash
# 在远程服务器上导入数据库的辅助脚本
# 使用方法：在远程服务器上执行此脚本

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 数据库配置
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_NAME="${DB_NAME:-owlrd}"

# 备份文件位置
BACKUP_DIR="${BACKUP_DIR:-~/owl-project/owlBack/exports}"
LATEST_BACKUP=$(ls -t "$BACKUP_DIR"/*.sql.gz 2>/dev/null | head -1)

export PGPASSWORD="$DB_PASSWORD"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}导入数据库备份${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "数据库: $DB_HOST:$DB_PORT/$DB_NAME"
echo "备份目录: $BACKUP_DIR"
echo ""

# 检查备份文件
if [ -z "$LATEST_BACKUP" ]; then
    echo -e "${RED}错误: 未找到备份文件${NC}"
    echo "请确保备份文件在: $BACKUP_DIR"
    exit 1
fi

echo -e "${BLUE}找到最新备份文件:${NC}"
echo "  $(basename "$LATEST_BACKUP")"
FILE_SIZE=$(du -h "$LATEST_BACKUP" | cut -f1)
echo "  文件大小: $FILE_SIZE"
echo ""

# 测试数据库连接
echo -e "${BLUE}测试数据库连接...${NC}"
if ! psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" > /dev/null 2>&1; then
    echo -e "${RED}错误: 无法连接到数据库${NC}"
    echo "请确保："
    echo "  1. PostgreSQL 服务正在运行"
    echo "  2. 数据库 $DB_NAME 已创建"
    echo "  3. 数据库配置正确"
    exit 1
fi
echo -e "${GREEN}✓ 数据库连接成功${NC}"
echo ""

# 警告
echo -e "${RED}警告: 导入操作将修改数据库${NC}"
echo -e "${YELLOW}如果数据库中有重要数据，请先备份！${NC}"
echo ""
read -p "是否继续？(yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "已取消"
    exit 0
fi

# 导入选项
echo ""
echo -e "${BLUE}选择导入方式:${NC}"
echo "1. 仅导入数据（保留现有表结构，需要表已存在）"
echo "2. 完整导入（表结构 + 数据，会覆盖现有表）"
read -p "请选择 (1/2): " import_option

case $import_option in
    1)
        echo -e "${BLUE}导入数据（保留表结构）...${NC}"
        gunzip -c "$LATEST_BACKUP" | \
            psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
            --set ON_ERROR_STOP=on
        ;;
    2)
        echo -e "${BLUE}导入完整数据库（表结构 + 数据）...${NC}"
        echo -e "${YELLOW}注意: 这将覆盖现有表结构${NC}"
        gunzip -c "$LATEST_BACKUP" | \
            psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
            --set ON_ERROR_STOP=on
        ;;
    *)
        echo "无效选择"
        exit 1
        ;;
esac

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}导入完成${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "验证数据："
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
    SELECT 
        'users' as table_name, COUNT(*) as row_count FROM users
    UNION ALL
    SELECT 'residents', COUNT(*) FROM residents
    UNION ALL
    SELECT 'devices', COUNT(*) FROM devices
    UNION ALL
    SELECT 'units', COUNT(*) FROM units;
"
