#!/bin/bash
# 验证 migrate 文件的修改是否已合并到主 SQL 文件

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
echo -e "${GREEN}验证 migrate 文件是否已合并${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 检查每个 migrate 文件
MIGRATE_FILES=$(find "$OWLRD_DB_DIR" -name "migrate*.sql" -type f | sort)

for migrate_file in $MIGRATE_FILES; do
    filename=$(basename "$migrate_file")
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}文件: $filename${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    # 提取表名
    TABLE_NAME=$(grep -oP "ALTER TABLE\s+\K[a-zA-Z_][a-zA-Z0-9_]*" "$migrate_file" 2>/dev/null | head -1)
    
    if [ -z "$TABLE_NAME" ]; then
        echo -e "${YELLOW}  ⚠️  无法提取表名${NC}"
        echo ""
        continue
    fi
    
    echo "  影响的表: $TABLE_NAME"
    
    # 查找对应的主 SQL 文件
    MAIN_FILE=$(find "$OWLRD_DB_DIR" -name "*.sql" -type f ! -name "migrate*" ! -name "*.md" ! -name "*.sh" | \
        xargs grep -l "CREATE TABLE.*$TABLE_NAME" 2>/dev/null | head -1)
    
    if [ -z "$MAIN_FILE" ]; then
        echo -e "${YELLOW}  ⚠️  未找到对应的主表文件${NC}"
        echo ""
        continue
    fi
    
    echo "  主表文件: $(basename "$MAIN_FILE")"
    
    # 检查添加的列
    ADDED_COLS=$(grep -oP "ADD COLUMN\s+\K[a-zA-Z_][a-zA-Z0-9_]*" "$migrate_file" 2>/dev/null | sort -u)
    
    if [ -n "$ADDED_COLS" ]; then
        echo "  添加的列:"
        ALL_MERGED=true
        for col in $ADDED_COLS; do
            if grep -q "\b$col\b" "$MAIN_FILE" 2>/dev/null; then
                echo -e "    ${GREEN}✓${NC} $col (已在主文件中)"
            else
                echo -e "    ${RED}✗${NC} $col (未在主文件中)"
                ALL_MERGED=false
            fi
        done
        
        # 检查数据库中是否存在这些列
        echo "  数据库中的列:"
        for col in $ADDED_COLS; do
            EXISTS=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
                SELECT COUNT(*) FROM information_schema.columns 
                WHERE table_name = '$TABLE_NAME' AND column_name = '$col';
            " | xargs)
            
            if [ "$EXISTS" = "1" ]; then
                echo -e "    ${GREEN}✓${NC} $col (数据库中存在)"
            else
                echo -e "    ${RED}✗${NC} $col (数据库中不存在)"
            fi
        done
        
        if [ "$ALL_MERGED" = true ]; then
            echo -e "${GREEN}  ✅ 所有添加的列已在主文件中${NC}"
        else
            echo -e "${YELLOW}  ⚠️  部分列未在主文件中，需要合并${NC}"
        fi
    fi
    
    # 检查删除的列
    DROPPED_COLS=$(grep -oP "DROP COLUMN\s+\K[a-zA-Z_][a-zA-Z0-9_]*" "$migrate_file" 2>/dev/null | sort -u)
    
    if [ -n "$DROPPED_COLS" ]; then
        echo "  删除的列:"
        for col in $DROPPED_COLS; do
            if grep -q "\b$col\b" "$MAIN_FILE" 2>/dev/null; then
                echo -e "    ${YELLOW}⚠️${NC} $col (仍在主文件中，需要删除)"
            else
                echo -e "    ${GREEN}✓${NC} $col (已从主文件删除)"
            fi
            
            # 检查数据库中是否还存在
            EXISTS=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
                SELECT COUNT(*) FROM information_schema.columns 
                WHERE table_name = '$TABLE_NAME' AND column_name = '$col';
            " | xargs)
            
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
echo "如果所有修改都已合并到主文件且数据库已应用，可以删除 migrate 文件"
