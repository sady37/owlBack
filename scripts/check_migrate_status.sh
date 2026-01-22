#!/bin/bash
# 检查 migrate 文件的状态：是否已合并到主文件，是否已应用到数据库

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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
echo -e "${GREEN}检查 migrate 文件状态${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 检查每个 migrate 文件
MIGRATE_FILES=$(find "$OWLRD_DB_DIR" -name "migrate*.sql" -type f | sort)

for migrate_file in $MIGRATE_FILES; do
    filename=$(basename "$migrate_file")
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}$filename${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    # 从文件注释或内容中提取表名
    TABLE_NAME=""
    
    # 方法1: 从注释中提取
    if grep -q "resident_contacts" "$migrate_file"; then
        TABLE_NAME="resident_contacts"
    elif grep -q "iot_timeseries" "$migrate_file"; then
        TABLE_NAME="iot_timeseries"
    elif grep -q "role_permissions" "$migrate_file"; then
        TABLE_NAME="role_permissions"
    elif grep -q "units" "$migrate_file" && grep -q "building" "$migrate_file"; then
        TABLE_NAME="units"
    elif grep -q "residents" "$migrate_file" && grep -q "branch_id" "$migrate_file"; then
        TABLE_NAME="residents"
    elif grep -q "buildings" "$migrate_file"; then
        TABLE_NAME="buildings"
    fi
    
    if [ -z "$TABLE_NAME" ]; then
        # 方法2: 从 ALTER TABLE 中提取（即使是在 DO 块中）
        TABLE_NAME=$(grep -oP "table_name\s*=\s*['\"]\K[a-zA-Z_][a-zA-Z0-9_]*" "$migrate_file" 2>/dev/null | head -1)
    fi
    
    if [ -z "$TABLE_NAME" ]; then
        echo -e "${YELLOW}  ⚠️  无法确定表名，手动检查文件内容${NC}"
        echo ""
        continue
    fi
    
    echo "  影响的表: $TABLE_NAME"
    
    # 查找主 SQL 文件
    MAIN_FILE=$(find "$OWLRD_DB_DIR" -name "*.sql" -type f ! -name "migrate*" ! -name "*.md" ! -name "*.sh" | \
        xargs grep -l "CREATE TABLE.*$TABLE_NAME" 2>/dev/null | head -1)
    
    if [ -z "$MAIN_FILE" ]; then
        echo -e "${YELLOW}  ⚠️  未找到主表文件${NC}"
        echo ""
        continue
    fi
    
    echo "  主表文件: $(basename "$MAIN_FILE")"
    
    # 检查添加的列
    ADDED_COLS=$(grep -oP "ADD COLUMN\s+\K[a-zA-Z_][a-zA-Z0-9_]*" "$migrate_file" 2>/dev/null | sort -u)
    
    if [ -n "$ADDED_COLS" ]; then
        echo "  添加的列检查:"
        ALL_MERGED=true
        ALL_IN_DB=true
        
        for col in $ADDED_COLS; do
            # 检查主文件
            if grep -q "\b$col\b" "$MAIN_FILE" 2>/dev/null; then
                echo -e "    ${GREEN}✓${NC} $col (已在主文件中)"
            else
                echo -e "    ${RED}✗${NC} $col (未在主文件中)"
                ALL_MERGED=false
            fi
            
            # 检查数据库
            EXISTS=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
                SELECT COUNT(*) FROM information_schema.columns 
                WHERE table_name = '$TABLE_NAME' AND column_name = '$col';
            " 2>/dev/null | xargs)
            
            if [ "$EXISTS" = "1" ]; then
                echo -e "      ${GREEN}✓${NC} 数据库中存在"
            else
                echo -e "      ${RED}✗${NC} 数据库中不存在"
                ALL_IN_DB=false
            fi
        done
        
        if [ "$ALL_MERGED" = true ] && [ "$ALL_IN_DB" = true ]; then
            echo -e "${GREEN}  ✅ 可以删除此 migrate 文件${NC}"
        elif [ "$ALL_MERGED" = true ]; then
            echo -e "${YELLOW}  ⚠️  已合并但数据库未应用${NC}"
        else
            echo -e "${RED}  ✗ 需要合并到主文件${NC}"
        fi
    fi
    
    # 检查删除的列
    DROPPED_COLS=$(grep -oP "DROP COLUMN\s+\K[a-zA-Z_][a-zA-Z0-9_]*" "$migrate_file" 2>/dev/null | sort -u)
    
    if [ -n "$DROPPED_COLS" ]; then
        echo "  删除的列检查:"
        for col in $DROPPED_COLS; do
            # 检查主文件
            if grep -q "\b$col\b" "$MAIN_FILE" 2>/dev/null; then
                echo -e "    ${YELLOW}⚠️${NC} $col (仍在主文件中)"
            else
                echo -e "    ${GREEN}✓${NC} $col (已从主文件删除)"
            fi
            
            # 检查数据库
            EXISTS=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
                SELECT COUNT(*) FROM information_schema.columns 
                WHERE table_name = '$TABLE_NAME' AND column_name = '$col';
            " 2>/dev/null | xargs)
            
            if [ "$EXISTS" = "0" ]; then
                echo -e "      ${GREEN}✓${NC} 数据库中已删除"
            else
                echo -e "      ${RED}✗${NC} 数据库中仍存在"
            fi
        done
    fi
    
    echo ""
done

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}总结${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "如果显示 ✅ 可以删除，说明："
echo "  1. migrate 的修改已合并到主 SQL 文件"
echo "  2. 数据库已应用这些修改"
echo "  3. 可以安全删除 migrate 文件"
echo ""
