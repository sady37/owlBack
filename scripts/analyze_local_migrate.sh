#!/bin/bash
# 分析本地 owlRD/db 中的 migrate* 文件

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

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}分析 migrate* 文件${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 查找所有 migrate* 文件
MIGRATE_FILES=$(find "$OWLRD_DB_DIR" -name "migrate*.sql" -type f | sort)

if [ -z "$MIGRATE_FILES" ]; then
    echo -e "${YELLOW}未找到 migrate* 文件${NC}"
    exit 0
fi

echo -e "${BLUE}找到以下 migrate* 文件:${NC}"
echo "$MIGRATE_FILES" | nl -w2 -s'. '
echo ""

# 分析每个文件
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}详细分析${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

for file in $MIGRATE_FILES; do
    filename=$(basename "$file")
    echo -e "${YELLOW}文件: $filename${NC}"
    
    # 检查文件内容类型
    if grep -q "ALTER TABLE" "$file"; then
        TABLES=$(grep -oP "ALTER TABLE\s+\K[a-zA-Z_][a-zA-Z0-9_]*" "$file" 2>/dev/null | sort -u | tr '\n' ' ')
        echo "  类型: ALTER TABLE 迁移"
        echo "  影响的表: $TABLES"
        
        # 检查对应的主表文件
        if [ -n "$TABLES" ]; then
            MAIN_TABLE=$(echo "$TABLES" | awk '{print $1}')
            MAIN_FILE=$(find "$OWLRD_DB_DIR" -name "*.sql" -type f ! -name "migrate*" ! -name "*.md" ! -name "*.sh" | \
                xargs grep -l "CREATE TABLE.*$MAIN_TABLE" 2>/dev/null | head -1)
            
            if [ -n "$MAIN_FILE" ]; then
                echo "  对应主表文件: $(basename "$MAIN_FILE")"
                
                # 检查主文件中是否已包含 migrate 中的列
                if grep -q "ADD COLUMN" "$file"; then
                    ADDED_COLS=$(grep -oP "ADD COLUMN\s+\K[a-zA-Z_][a-zA-Z0-9_]*" "$file" 2>/dev/null)
                    ALL_FOUND=true
                    for col in $ADDED_COLS; do
                        if ! grep -q "\b$col\b" "$MAIN_FILE" 2>/dev/null; then
                            ALL_FOUND=false
                            break
                        fi
                    done
                    if [ "$ALL_FOUND" = true ] && [ -n "$ADDED_COLS" ]; then
                        echo -e "${GREEN}  ✅ 所有添加的列已在主文件中${NC}"
                    elif [ -n "$ADDED_COLS" ]; then
                        echo -e "${YELLOW}  ⚠️  部分列可能未在主文件中${NC}"
                    fi
                fi
            fi
        fi
    fi
    
    if grep -q "ADD COLUMN" "$file"; then
        COLUMNS=$(grep -oP "ADD COLUMN\s+\K[a-zA-Z_][a-zA-Z0-9_]*" "$file" 2>/dev/null | sort -u | tr '\n' ' ')
        if [ -n "$COLUMNS" ]; then
            echo "  添加的列: $COLUMNS"
        fi
    fi
    
    if grep -q "DROP COLUMN" "$file"; then
        COLUMNS=$(grep -oP "DROP COLUMN\s+\K[a-zA-Z_][a-zA-Z0-9_]*" "$file" 2>/dev/null | sort -u | tr '\n' ' ')
        if [ -n "$COLUMNS" ]; then
            echo "  删除的列: $COLUMNS"
        fi
    fi
    
    SIZE=$(wc -l < "$file" | xargs)
    echo "  行数: $SIZE"
    echo ""
done

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}建议${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "1. 检查每个 migrate 文件的修改是否已合并到对应的主表定义文件"
echo "2. 如果已合并且数据库已应用，可以删除 migrate 文件"
echo "3. 如果未合并，需要先合并到主文件，然后删除 migrate 文件"
echo ""
