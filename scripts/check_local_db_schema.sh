#!/bin/bash
# 检查本地数据库表结构与 owlRD/db 中的 SQL 文件是否一致

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

# 项目路径
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OWLBACK_DIR="$(dirname "$SCRIPT_DIR")"
OWLRD_DB_DIR="${OWLBACK_DIR%/owlBack}/owlRD/db"

# 设置 PATH（如果 psql 不在默认路径）
export PATH="/usr/local/opt/postgresql@15/bin:$PATH"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}本地数据库表结构检查工具${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "数据库: $DB_HOST:$DB_PORT/$DB_NAME"
echo "SQL 文件目录: $OWLRD_DB_DIR"
echo ""

# 检查 PostgreSQL 客户端
if ! command -v psql &> /dev/null; then
    echo -e "${RED}错误: psql 未安装${NC}"
    echo "请安装: brew install postgresql@15"
    exit 1
fi

# 设置密码环境变量
export PGPASSWORD="$DB_PASSWORD"

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

# 获取所有表名
echo -e "${BLUE}获取数据库中的表...${NC}"
TABLES=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
    SELECT tablename 
    FROM pg_tables 
    WHERE schemaname = 'public' 
    ORDER BY tablename;
" | sed 's/^[[:space:]]*//' | sed 's/[[:space:]]*$//' | grep -v '^$')

TABLE_COUNT=$(echo "$TABLES" | wc -l | xargs)
echo -e "${GREEN}找到 $TABLE_COUNT 个表${NC}"
echo ""

# 对比每个表的结构
DIFF_COUNT=0
MATCH_COUNT=0
MISSING_SQL_COUNT=0

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}表结构对比${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

for table in $TABLES; do
    echo -e "${BLUE}表: $table${NC}"
    
    # 获取数据库中的列信息
    DB_COLS=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT 
            column_name || '|' || 
            data_type || 
            COALESCE('(' || character_maximum_length || ')', '') || 
            COALESCE('(' || numeric_precision || ',' || numeric_scale || ')', '') || '|' ||
            CASE WHEN is_nullable = 'YES' THEN 'NULL' ELSE 'NOT NULL' END
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = '$table'
        ORDER BY ordinal_position;
    " | sed 's/^[[:space:]]*//' | sed 's/[[:space:]]*$//' | grep -v '^$')
    
    # 查找对应的 SQL 文件
    SQL_FILE=$(find "$OWLRD_DB_DIR" -name "*.sql" -type f ! -name "migrate*" ! -name "*.md" ! -name "*.sh" | \
        xargs grep -l "CREATE TABLE.*$table" 2>/dev/null | head -1)
    
    if [ -z "$SQL_FILE" ]; then
        echo -e "${YELLOW}  ⚠️  未找到对应的 SQL 文件${NC}"
        echo "  数据库列数: $(echo "$DB_COLS" | wc -l | xargs)"
        MISSING_SQL_COUNT=$((MISSING_SQL_COUNT + 1))
        DIFF_COUNT=$((DIFF_COUNT + 1))
        echo ""
        continue
    fi
    
    echo "  SQL 文件: $(basename "$SQL_FILE")"
    
    # 从 SQL 文件提取列定义（简化）
    SQL_COLS=$(grep -A 200 "CREATE TABLE.*$table" "$SQL_FILE" 2>/dev/null | \
        grep -E "^\s+[a-zA-Z_][a-zA-Z0-9_]*\s+" | \
        grep -v "CONSTRAINT\|PRIMARY KEY\|FOREIGN KEY\|UNIQUE\|CHECK\|INDEX\|--" | \
        head -100 | \
        sed 's/^[[:space:]]*//' | \
        sed 's/[[:space:]]*$//' | \
        sed 's/--.*$//' | \
        grep -v '^$' | \
        awk '{print $1}' | \
        grep -v '^$')
    
    DB_COL_COUNT=$(echo "$DB_COLS" | wc -l | xargs)
    SQL_COL_COUNT=$(echo "$SQL_COLS" | wc -l | xargs)
    
    echo "  数据库列数: $DB_COL_COUNT"
    echo "  SQL 文件列数: $SQL_COL_COUNT"
    
    if [ "$DB_COL_COUNT" = "$SQL_COL_COUNT" ]; then
        echo -e "${GREEN}  ✅ 列数匹配${NC}"
        MATCH_COUNT=$((MATCH_COUNT + 1))
    else
        echo -e "${YELLOW}  ⚠️  列数不匹配${NC}"
        DIFF_COUNT=$((DIFF_COUNT + 1))
        
        # 显示前几列作为示例
        echo -e "${YELLOW}  数据库列（前5列）:${NC}"
        echo "$DB_COLS" | head -5 | sed 's/^/    /'
        if [ "$DB_COL_COUNT" -gt 5 ]; then
            echo "    ... (共 $DB_COL_COUNT 列)"
        fi
    fi
    echo ""
done

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}对比结果${NC}"
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}匹配: $MATCH_COUNT${NC}"
echo -e "${YELLOW}差异: $DIFF_COUNT${NC}"
if [ "$MISSING_SQL_COUNT" -gt 0 ]; then
    echo -e "${YELLOW}未找到 SQL 文件: $MISSING_SQL_COUNT${NC}"
fi
echo ""
