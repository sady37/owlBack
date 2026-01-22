#!/bin/bash
# 清理已合并的 migrate 文件

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
echo -e "${GREEN}清理已合并的 migrate 文件${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 已确认可以删除的文件（已合并到主文件且数据库已应用）
SAFE_TO_DELETE=(
    "migrate_add_contact_hash_columns.sql"
    "migrate_iot_timeseries.sql"
    "migrate_permission_scope.sql"
)

# 需要进一步检查的文件
NEED_REVIEW=(
    "migrate_branch_id_not_null.sql"
    "migrate_residents_branch_id_not_null.sql"
    "migrate_units_building_id_nullable.sql"
    "migrate_units_building_name_to_id.sql"
)

echo -e "${BLUE}可以安全删除的文件（已合并且数据库已应用）:${NC}"
for file in "${SAFE_TO_DELETE[@]}"; do
    if [ -f "$OWLRD_DB_DIR/$file" ]; then
        echo "  - $file"
    fi
done

echo ""
echo -e "${YELLOW}需要进一步检查的文件:${NC}"
for file in "${NEED_REVIEW[@]}"; do
    if [ -f "$OWLRD_DB_DIR/$file" ]; then
        echo "  - $file"
    fi
done

echo ""
echo -e "${RED}警告: 删除前请确认这些迁移已经应用到数据库并合并到主表定义中${NC}"
echo ""
read -p "是否删除已确认的文件？(yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "已取消"
    exit 0
fi

# 删除已确认的文件
DELETED_COUNT=0
for file in "${SAFE_TO_DELETE[@]}"; do
    if [ -f "$OWLRD_DB_DIR/$file" ]; then
        echo "删除: $file"
        rm "$OWLRD_DB_DIR/$file"
        DELETED_COUNT=$((DELETED_COUNT + 1))
    fi
done

echo ""
if [ "$DELETED_COUNT" -gt 0 ]; then
    echo -e "${GREEN}已删除 $DELETED_COUNT 个 migrate 文件${NC}"
else
    echo -e "${YELLOW}没有找到要删除的文件${NC}"
fi

echo ""
echo -e "${BLUE}剩余 migrate 文件:${NC}"
REMAINING=$(find "$OWLRD_DB_DIR" -name "migrate*.sql" -type f | wc -l | xargs)
if [ "$REMAINING" -gt 0 ]; then
    find "$OWLRD_DB_DIR" -name "migrate*.sql" -type f | sed 's/^/  /'
    echo ""
    echo "这些文件需要进一步检查是否已合并"
else
    echo "  无"
fi
