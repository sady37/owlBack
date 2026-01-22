#!/bin/bash
# 导出本地数据库数据

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 数据库配置（本地 Docker）
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_NAME="${DB_NAME:-owlrd}"

# 导出配置
EXPORT_DIR="${EXPORT_DIR:-./exports}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
EXPORT_FILE="${EXPORT_DIR}/owlrd_backup_${TIMESTAMP}.sql"
EXPORT_FILE_COMPRESSED="${EXPORT_FILE}.gz"

# 设置 PATH
export PATH="/usr/local/opt/postgresql@15/bin:$PATH"
export PGPASSWORD="$DB_PASSWORD"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}导出数据库数据${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "数据库: $DB_HOST:$DB_PORT/$DB_NAME"
echo "说明: $DB_NAME 是数据库名称（Database Name），不是实例名"
echo "      PostgreSQL 实例运行在 $DB_HOST:$DB_PORT"
echo "导出目录: $EXPORT_DIR"
echo ""

# 检查 pg_dump
if ! command -v pg_dump &> /dev/null; then
    echo -e "${RED}错误: pg_dump 未安装${NC}"
    echo "请安装: brew install postgresql@15"
    exit 1
fi

# 创建导出目录
mkdir -p "$EXPORT_DIR"

# 测试数据库连接
echo -e "${BLUE}测试数据库连接...${NC}"
if ! psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" > /dev/null 2>&1; then
    echo -e "${RED}错误: 无法连接到数据库${NC}"
    echo "请确保："
    echo "  1. Docker Compose 服务正在运行: cd owlBack && docker-compose up -d"
    echo "  2. 数据库配置正确"
    exit 1
fi
echo -e "${GREEN}✓ 数据库连接成功${NC}"
echo ""

# 导出选项
echo -e "${BLUE}选择导出选项:${NC}"
echo "1. 仅导出数据（不包含表结构）"
echo "2. 仅导出表结构（不包含数据）"
echo "3. 导出完整数据库（表结构 + 数据）"
echo "4. 导出完整数据库（压缩）"
read -p "请选择 (1/2/3/4): " export_option

case $export_option in
    1)
        echo -e "${BLUE}导出数据（不包含表结构）...${NC}"
        pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
            --data-only \
            --no-owner \
            --no-privileges \
            -f "$EXPORT_FILE"
        ;;
    2)
        echo -e "${BLUE}导出表结构（不包含数据）...${NC}"
        pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
            --schema-only \
            --no-owner \
            --no-privileges \
            -f "$EXPORT_FILE"
        ;;
    3)
        echo -e "${BLUE}导出完整数据库（表结构 + 数据）...${NC}"
        pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
            --no-owner \
            --no-privileges \
            -f "$EXPORT_FILE"
        ;;
    4)
        echo -e "${BLUE}导出完整数据库（压缩）...${NC}"
        pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
            --no-owner \
            --no-privileges \
            -F c \
            -f "${EXPORT_FILE%.sql}.dump"
        EXPORT_FILE="${EXPORT_FILE%.sql}.dump"
        EXPORT_FILE_COMPRESSED="${EXPORT_FILE}"
        ;;
    *)
        echo "无效选择，使用默认：完整数据库"
        pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
            --no-owner \
            --no-privileges \
            -f "$EXPORT_FILE"
        ;;
esac

# 如果是 SQL 文件，尝试压缩
if [ "$export_option" != "4" ] && [ -f "$EXPORT_FILE" ]; then
    echo ""
    echo -e "${BLUE}压缩导出文件...${NC}"
    gzip -f "$EXPORT_FILE"
    EXPORT_FILE_COMPRESSED="${EXPORT_FILE}.gz"
    EXPORT_FILE="$EXPORT_FILE_COMPRESSED"
fi

# 显示文件信息
echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}导出完成${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
if [ -f "$EXPORT_FILE" ]; then
    FILE_SIZE=$(du -h "$EXPORT_FILE" | cut -f1)
    echo "导出文件: $EXPORT_FILE"
    echo "文件大小: $FILE_SIZE"
    echo ""
    echo "可以使用以下命令导入到远程服务器："
    echo "  scp $EXPORT_FILE user@remote-server:/path/to/destination/"
    echo ""
    echo "然后在远程服务器上执行："
    echo "  注意: 如果远程数据库名不是 'owlrd'，请修改命令中的数据库名"
    if [[ "$EXPORT_FILE" == *.dump ]]; then
        echo "  pg_restore -h localhost -U postgres -d owlrd -c $EXPORT_FILE"
        echo "  # 或使用其他数据库名: pg_restore -h localhost -U postgres -d <数据库名> -c $EXPORT_FILE"
    elif [[ "$EXPORT_FILE" == *.gz ]]; then
        echo "  gunzip -c $EXPORT_FILE | psql -h localhost -U postgres -d owlrd"
        echo "  # 或使用其他数据库名: gunzip -c $EXPORT_FILE | psql -h localhost -U postgres -d <数据库名>"
    else
        echo "  psql -h localhost -U postgres -d owlrd < $EXPORT_FILE"
        echo "  # 或使用其他数据库名: psql -h localhost -U postgres -d <数据库名> < $EXPORT_FILE"
    fi
else
    echo -e "${RED}错误: 导出文件未生成${NC}"
    exit 1
fi
