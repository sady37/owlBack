#!/bin/bash
# 同步本地数据库到远程服务器

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 读取部署配置
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DEPLOY_CONFIG="${SCRIPT_DIR}/../deploy-config.sh"

if [ -f "$DEPLOY_CONFIG" ]; then
    source "$DEPLOY_CONFIG"
else
    echo -e "${YELLOW}警告: deploy-config.sh 不存在，使用默认配置${NC}"
    REMOTE_HOST="${REMOTE_HOST:-your-server.com}"
    REMOTE_USER="${REMOTE_USER:-your-user}"
    REMOTE_PATH="${REMOTE_PATH:-/home/your-user/project}"
fi

# 本地数据库配置
LOCAL_DB_HOST="${DB_HOST:-localhost}"
LOCAL_DB_PORT="${DB_PORT:-5432}"
LOCAL_DB_USER="${DB_USER:-postgres}"
LOCAL_DB_PASSWORD="${DB_PASSWORD:-postgres}"
LOCAL_DB_NAME="${DB_NAME:-owlrd}"

# 远程数据库配置（默认与本地相同，可通过环境变量覆盖）
REMOTE_DB_HOST="${REMOTE_DB_HOST:-localhost}"
REMOTE_DB_PORT="${REMOTE_DB_PORT:-5432}"
REMOTE_DB_USER="${REMOTE_DB_USER:-postgres}"
REMOTE_DB_PASSWORD="${REMOTE_DB_PASSWORD:-postgres}"
REMOTE_DB_NAME="${REMOTE_DB_NAME:-owlrd}"

# 临时文件
TEMP_DIR="/tmp/owlrd_sync_$$"
EXPORT_FILE="${TEMP_DIR}/owlrd_backup.sql.gz"

export PATH="/usr/local/opt/postgresql@15/bin:$PATH"
export PGPASSWORD="$LOCAL_DB_PASSWORD"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}同步数据库到远程服务器${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "本地数据库: $LOCAL_DB_HOST:$LOCAL_DB_PORT/$LOCAL_DB_NAME"
echo "远程服务器: $REMOTE_USER@$REMOTE_HOST"
echo "远程数据库: $REMOTE_DB_HOST:$REMOTE_DB_PORT/$REMOTE_DB_NAME"
echo ""

# 检查 pg_dump
if ! command -v pg_dump &> /dev/null; then
    echo -e "${RED}错误: pg_dump 未安装${NC}"
    echo "请安装: brew install postgresql@15"
    exit 1
fi

# 创建临时目录
mkdir -p "$TEMP_DIR"
trap "rm -rf $TEMP_DIR" EXIT

# 测试本地数据库连接
echo -e "${BLUE}测试本地数据库连接...${NC}"
if ! psql -h "$LOCAL_DB_HOST" -p "$LOCAL_DB_PORT" -U "$LOCAL_DB_USER" -d "$LOCAL_DB_NAME" -c "SELECT 1;" > /dev/null 2>&1; then
    echo -e "${RED}错误: 无法连接到本地数据库${NC}"
    exit 1
fi
echo -e "${GREEN}✓ 本地数据库连接成功${NC}"
echo ""

# 导出选项
echo -e "${BLUE}选择同步方式:${NC}"
echo "1. 仅同步数据（保留远程表结构）"
echo "2. 完整同步（表结构 + 数据，会覆盖远程数据库）"
read -p "请选择 (1/2): " sync_option

case $sync_option in
    1)
        echo -e "${BLUE}导出数据（不包含表结构）...${NC}"
        pg_dump -h "$LOCAL_DB_HOST" -p "$LOCAL_DB_PORT" -U "$LOCAL_DB_USER" -d "$LOCAL_DB_NAME" \
            --data-only \
            --no-owner \
            --no-privileges \
            | gzip > "$EXPORT_FILE"
        SYNC_MODE="data-only"
        ;;
    2)
        echo -e "${BLUE}导出完整数据库（表结构 + 数据）...${NC}"
        pg_dump -h "$LOCAL_DB_HOST" -p "$LOCAL_DB_PORT" -U "$LOCAL_DB_USER" -d "$LOCAL_DB_NAME" \
            --no-owner \
            --no-privileges \
            | gzip > "$EXPORT_FILE"
        SYNC_MODE="full"
        ;;
    *)
        echo "无效选择"
        exit 1
        ;;
esac

FILE_SIZE=$(du -h "$EXPORT_FILE" | cut -f1)
echo -e "${GREEN}✓ 导出完成，文件大小: $FILE_SIZE${NC}"
echo ""

# 传输到远程服务器
echo -e "${BLUE}传输文件到远程服务器...${NC}"
REMOTE_TEMP_FILE="/tmp/owlrd_backup_$$.sql.gz"
scp "$EXPORT_FILE" "$REMOTE_USER@$REMOTE_HOST:$REMOTE_TEMP_FILE"
echo -e "${GREEN}✓ 文件传输完成${NC}"
echo ""

# 在远程服务器上导入
echo -e "${BLUE}在远程服务器上导入数据...${NC}"
echo -e "${YELLOW}警告: 这将修改远程数据库${NC}"
read -p "是否继续？(yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "已取消"
    ssh "$REMOTE_USER@$REMOTE_HOST" "rm -f $REMOTE_TEMP_FILE"
    exit 0
fi

# 执行远程导入
if [ "$SYNC_MODE" = "data-only" ]; then
    echo "导入数据（保留表结构）..."
    ssh "$REMOTE_USER@$REMOTE_HOST" "
        export PGPASSWORD='$REMOTE_DB_PASSWORD'
        gunzip -c $REMOTE_TEMP_FILE | psql -h $REMOTE_DB_HOST -p $REMOTE_DB_PORT -U $REMOTE_DB_USER -d $REMOTE_DB_NAME
        rm -f $REMOTE_TEMP_FILE
    "
else
    echo "导入完整数据库（会覆盖现有数据）..."
    ssh "$REMOTE_USER@$REMOTE_HOST" "
        export PGPASSWORD='$REMOTE_DB_PASSWORD'
        # 先删除现有数据库（可选，谨慎使用）
        # psql -h $REMOTE_DB_HOST -p $REMOTE_DB_PORT -U $REMOTE_DB_USER -d postgres -c \"DROP DATABASE IF EXISTS $REMOTE_DB_NAME;\"
        # psql -h $REMOTE_DB_HOST -p $REMOTE_DB_PORT -U $REMOTE_DB_USER -d postgres -c \"CREATE DATABASE $REMOTE_DB_NAME;\"
        gunzip -c $REMOTE_TEMP_FILE | psql -h $REMOTE_DB_HOST -p $REMOTE_DB_PORT -U $REMOTE_DB_USER -d $REMOTE_DB_NAME
        rm -f $REMOTE_TEMP_FILE
    "
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}同步完成${NC}"
echo -e "${GREEN}========================================${NC}"
