#!/bin/bash
# 检查剩余 migrate 文件的状态：主文件 vs 数据库
# 只报告差异，不做自动判断，由用户决定以哪个为准

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# 项目路径
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OWLBACK_DIR="$(dirname "$SCRIPT_DIR")"
OWLRD_DB_DIR="${OWLBACK_DIR%/owlBack}/owlRD/db"

# 数据库配置
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_NAME="${DB_NAME:-owlrd}"

export PATH="/usr/local/opt/postgresql@15/bin:$PATH"
export PGPASSWORD="$DB_PASSWORD"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}检查剩余 migrate 文件状态${NC}"
echo -e "${GREEN}========================================${NC}"
echo -e "${CYAN}说明：只报告主文件和数据库的差异，不做自动判断${NC}"
echo -e "${CYAN}如果两者不一致，由您决定以哪个为准${NC}"
echo ""

# 检查函数：检查列是否为 NOT NULL
check_column_nullable() {
    local table=$1
    local column=$2
    local result=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT is_nullable 
        FROM information_schema.columns 
        WHERE table_name = '$table' AND column_name = '$column';
    " 2>/dev/null | xargs)
    echo "$result"
}

# 检查函数：检查列是否存在
check_column_exists() {
    local table=$1
    local column=$2
    local result=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT COUNT(*) 
        FROM information_schema.columns 
        WHERE table_name = '$table' AND column_name = '$column';
    " 2>/dev/null | xargs)
    if [ "$result" = "1" ]; then
        echo "YES"
    else
        echo "NO"
    fi
}

# 检查函数：检查主文件中列的定义
check_main_file_column() {
    local main_file=$1
    local column=$2
    # 查找列定义行（包含列名和类型定义）
    local col_line=$(grep -E "^\s+$column\s+" "$main_file" 2>/dev/null | head -1)
    if [ -n "$col_line" ]; then
        # 检查是否为 NOT NULL（在列定义行中）
        if echo "$col_line" | grep -q "NOT NULL"; then
            echo "NOT NULL"
        else
            echo "NULLABLE"
        fi
    else
        echo "NOT FOUND"
    fi
}

# 1. migrate_branch_id_not_null.sql
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}1. migrate_branch_id_not_null.sql${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo "  修改内容：将 buildings.branch_id 和 units.branch_id 改为 NOT NULL"
echo ""

# 检查 buildings.branch_id
echo "  buildings.branch_id:"
MAIN_FILE="${OWLRD_DB_DIR}/06_buildings.sql"
if [ -f "$MAIN_FILE" ]; then
    MAIN_STATUS=$(check_main_file_column "$MAIN_FILE" "branch_id")
    echo -e "    主文件状态: ${CYAN}$MAIN_STATUS${NC}"
else
    echo -e "    主文件状态: ${RED}文件不存在${NC}"
    MAIN_STATUS="NOT FOUND"
fi

DB_STATUS=$(check_column_nullable "buildings" "branch_id")
if [ "$DB_STATUS" = "NO" ]; then
    echo -e "    数据库状态: ${RED}NOT NULL${NC}"
    DB_STATUS="NOT NULL"
elif [ "$DB_STATUS" = "YES" ]; then
    echo -e "    数据库状态: ${YELLOW}NULLABLE${NC}"
    DB_STATUS="NULLABLE"
else
    echo -e "    数据库状态: ${RED}列不存在${NC}"
    DB_STATUS="NOT FOUND"
fi

if [ "$MAIN_STATUS" != "$DB_STATUS" ]; then
    echo -e "    ${RED}⚠️  主文件与数据库不一致！${NC}"
fi
echo ""

# 检查 units.branch_id
echo "  units.branch_id:"
MAIN_FILE="${OWLRD_DB_DIR}/09_units.sql"
if [ -f "$MAIN_FILE" ]; then
    MAIN_STATUS=$(check_main_file_column "$MAIN_FILE" "branch_id")
    echo -e "    主文件状态: ${CYAN}$MAIN_STATUS${NC}"
else
    echo -e "    主文件状态: ${RED}文件不存在${NC}"
    MAIN_STATUS="NOT FOUND"
fi

DB_STATUS=$(check_column_nullable "units" "branch_id")
if [ "$DB_STATUS" = "NO" ]; then
    echo -e "    数据库状态: ${RED}NOT NULL${NC}"
    DB_STATUS="NOT NULL"
elif [ "$DB_STATUS" = "YES" ]; then
    echo -e "    数据库状态: ${YELLOW}NULLABLE${NC}"
    DB_STATUS="NULLABLE"
else
    echo -e "    数据库状态: ${RED}列不存在${NC}"
    DB_STATUS="NOT FOUND"
fi

if [ "$MAIN_STATUS" != "$DB_STATUS" ]; then
    echo -e "    ${RED}⚠️  主文件与数据库不一致！${NC}"
fi
echo ""

# 2. migrate_residents_branch_id_not_null.sql
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}2. migrate_residents_branch_id_not_null.sql${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo "  修改内容：将 residents.branch_id 改为 NOT NULL"
echo ""

echo "  residents.branch_id:"
MAIN_FILE="${OWLRD_DB_DIR}/12_residents.sql"
if [ -f "$MAIN_FILE" ]; then
    MAIN_STATUS=$(check_main_file_column "$MAIN_FILE" "branch_id")
    echo -e "    主文件状态: ${CYAN}$MAIN_STATUS${NC}"
else
    echo -e "    主文件状态: ${RED}文件不存在${NC}"
    MAIN_STATUS="NOT FOUND"
fi

DB_STATUS=$(check_column_nullable "residents" "branch_id")
if [ "$DB_STATUS" = "NO" ]; then
    echo -e "    数据库状态: ${RED}NOT NULL${NC}"
    DB_STATUS="NOT NULL"
elif [ "$DB_STATUS" = "YES" ]; then
    echo -e "    数据库状态: ${YELLOW}NULLABLE${NC}"
    DB_STATUS="NULLABLE"
else
    echo -e "    数据库状态: ${RED}列不存在${NC}"
    DB_STATUS="NOT FOUND"
fi

if [ "$MAIN_STATUS" != "$DB_STATUS" ]; then
    echo -e "    ${RED}⚠️  主文件与数据库不一致！${NC}"
fi
echo ""

# 3. migrate_units_building_id_nullable.sql
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}3. migrate_units_building_id_nullable.sql${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo "  修改内容：将 units.building_id 改为允许 NULL"
echo ""

echo "  units.building_id:"
MAIN_FILE="${OWLRD_DB_DIR}/09_units.sql"
if [ -f "$MAIN_FILE" ]; then
    MAIN_STATUS=$(check_main_file_column "$MAIN_FILE" "building_id")
    echo -e "    主文件状态: ${CYAN}$MAIN_STATUS${NC}"
else
    echo -e "    主文件状态: ${RED}文件不存在${NC}"
    MAIN_STATUS="NOT FOUND"
fi

DB_STATUS=$(check_column_nullable "units" "building_id")
if [ "$DB_STATUS" = "NO" ]; then
    echo -e "    数据库状态: ${RED}NOT NULL${NC}"
    DB_STATUS="NOT NULL"
elif [ "$DB_STATUS" = "YES" ]; then
    echo -e "    数据库状态: ${YELLOW}NULLABLE${NC}"
    DB_STATUS="NULLABLE"
else
    echo -e "    数据库状态: ${RED}列不存在${NC}"
    DB_STATUS="NOT FOUND"
fi

if [ "$MAIN_STATUS" != "$DB_STATUS" ]; then
    echo -e "    ${RED}⚠️  主文件与数据库不一致！${NC}"
fi
echo ""

# 4. migrate_units_building_name_to_id.sql
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}4. migrate_units_building_name_to_id.sql${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo "  修改内容：将 units.building_name 改为 building_id，并删除 building_name 列"
echo ""

echo "  units.building_name:"
MAIN_FILE="${OWLRD_DB_DIR}/09_units.sql"
if [ -f "$MAIN_FILE" ]; then
    # 只检查实际的列定义行（以列名开头，不是注释）
    if grep -E "^\s+building_name\s+" "$MAIN_FILE" 2>/dev/null | grep -v "^[[:space:]]*--" | head -1 | grep -q .; then
        echo -e "    主文件状态: ${YELLOW}存在（应该删除）${NC}"
        MAIN_HAS_NAME="YES"
    else
        echo -e "    主文件状态: ${GREEN}不存在（已删除）${NC}"
        MAIN_HAS_NAME="NO"
    fi
else
    echo -e "    主文件状态: ${RED}文件不存在${NC}"
    MAIN_HAS_NAME="UNKNOWN"
fi

DB_HAS_NAME=$(check_column_exists "units" "building_name")
if [ "$DB_HAS_NAME" = "YES" ]; then
    echo -e "    数据库状态: ${YELLOW}存在（应该删除）${NC}"
else
    echo -e "    数据库状态: ${GREEN}不存在（已删除）${NC}"
fi

if [ "$MAIN_HAS_NAME" = "YES" ] && [ "$DB_HAS_NAME" = "NO" ]; then
    echo -e "    ${RED}⚠️  主文件仍有 building_name，但数据库已删除！${NC}"
elif [ "$MAIN_HAS_NAME" = "NO" ] && [ "$DB_HAS_NAME" = "YES" ]; then
    echo -e "    ${RED}⚠️  主文件已删除 building_name，但数据库仍存在！${NC}"
fi
echo ""

echo "  units.building_id:"
if [ -f "$MAIN_FILE" ]; then
    MAIN_STATUS=$(check_main_file_column "$MAIN_FILE" "building_id")
    echo -e "    主文件状态: ${CYAN}$MAIN_STATUS${NC}"
else
    echo -e "    主文件状态: ${RED}文件不存在${NC}"
    MAIN_STATUS="NOT FOUND"
fi

DB_STATUS=$(check_column_nullable "units" "building_id")
if [ "$DB_STATUS" = "NO" ]; then
    echo -e "    数据库状态: ${RED}NOT NULL${NC}"
    DB_STATUS="NOT NULL"
elif [ "$DB_STATUS" = "YES" ]; then
    echo -e "    数据库状态: ${YELLOW}NULLABLE${NC}"
    DB_STATUS="NULLABLE"
else
    echo -e "    数据库状态: ${RED}列不存在${NC}"
    DB_STATUS="NOT FOUND"
fi

if [ "$MAIN_STATUS" != "$DB_STATUS" ]; then
    echo -e "    ${RED}⚠️  主文件与数据库不一致！${NC}"
fi
echo ""

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}检查完成${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${CYAN}说明：${NC}"
echo "  - 如果主文件与数据库不一致，由您决定以哪个为准"
echo "  - 如果两者都满足 migrate 的要求，可以删除 migrate 文件"
echo "  - 如果只有一方满足，需要同步另一方"
echo ""
