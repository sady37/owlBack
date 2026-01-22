#!/bin/bash
# 清理本地 owlRD/db 中无用的 migrate* 文件

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
echo -e "${GREEN}清理 migrate* 文件${NC}"
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

# 显示每个文件的简要信息
echo -e "${BLUE}文件详情:${NC}"
for file in $MIGRATE_FILES; do
    filename=$(basename "$file")
    echo ""
    echo -e "${YELLOW}$filename${NC}"
    head -10 "$file" | grep -E "^--|^BEGIN|^ALTER|^ADD|^DROP" | head -5 | sed 's/^/  /'
done

echo ""
echo -e "${RED}警告: 删除文件前请确认这些迁移已经应用到数据库并合并到主表定义中${NC}"
echo ""
read -p "是否继续？(yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "已取消"
    exit 0
fi

echo ""
echo -e "${BLUE}选择要删除的文件:${NC}"
echo "1. 删除所有 migrate* 文件"
echo "2. 手动选择要删除的文件"
echo "3. 取消"
read -p "请选择 (1/2/3): " choice

case $choice in
    1)
        echo ""
        echo -e "${RED}删除所有 migrate* 文件...${NC}"
        for file in $MIGRATE_FILES; do
            echo "  删除: $(basename "$file")"
            rm "$file"
        done
        echo -e "${GREEN}所有 migrate* 文件已删除${NC}"
        ;;
    2)
        echo ""
        echo -e "${BLUE}请输入要删除的文件编号（用空格分隔，例如: 1 3 5）:${NC}"
        read -p "> " file_nums
        
        for num in $file_nums; do
            file=$(echo "$MIGRATE_FILES" | sed -n "${num}p")
            if [ -n "$file" ] && [ -f "$file" ]; then
                echo "  删除: $(basename "$file")"
                rm "$file"
            fi
        done
        echo -e "${GREEN}选定的文件已删除${NC}"
        ;;
    3)
        echo "已取消"
        exit 0
        ;;
    *)
        echo "无效选择"
        exit 1
        ;;
esac

echo ""
echo -e "${GREEN}清理完成${NC}"
