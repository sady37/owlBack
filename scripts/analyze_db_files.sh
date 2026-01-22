#!/bin/bash
# 分析 owlRD/db 目录下的文件，识别可清理的文件

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

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}分析 owlRD/db 目录文件${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 1. 核心表定义文件（应该保留）
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}1. 核心表定义文件（保留）${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
CORE_FILES=$(find "$OWLRD_DB_DIR" -maxdepth 1 -name "[0-9][0-9]_*.sql" -type f | sort)
CORE_COUNT=$(echo "$CORE_FILES" | wc -l | xargs)
echo "找到 $CORE_COUNT 个核心表定义文件："
echo "$CORE_FILES" | sed 's|.*/||' | nl -w2 -s'. '
echo ""

# 2. 系统文件（可以删除）
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}2. 系统文件（建议删除）${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
SYSTEM_FILES=$(find "$OWLRD_DB_DIR" -maxdepth 1 -name ".DS_Store" -o -name "._*" -type f 2>/dev/null | sort)
if [ -n "$SYSTEM_FILES" ]; then
    echo "$SYSTEM_FILES" | sed 's|.*/||' | nl -w2 -s'. '
else
    echo "  无"
fi
echo ""

# 3. 临时/修复脚本（可能可以删除）
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}3. 临时/修复脚本（需要检查）${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
TEMP_SCRIPTS=$(find "$OWLRD_DB_DIR" -maxdepth 1 -type f \( \
    -name "*fix*.sql" \
    -o -name "*clean*.sql" \
    -o -name "*clear*.sql" \
    -o -name "*rebuild*.sql" \
    -o -name "*migration*.sql" \
    -o -name "*update*.sql" \
    -o -name "*delete*.sql" \
    -o -name "*check*.sql" \
    -o -name "*diagnose*.sql" \
    -o -name "*enable*.sql" \
\) | sort)
if [ -n "$TEMP_SCRIPTS" ]; then
    echo "$TEMP_SCRIPTS" | sed 's|.*/||' | nl -w2 -s'. '
    TEMP_COUNT=$(echo "$TEMP_SCRIPTS" | wc -l | xargs)
    echo ""
    echo "  共 $TEMP_COUNT 个文件"
else
    echo "  无"
fi
echo ""

# 4. 测试/演示数据脚本（可能可以删除或移动到测试目录）
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}4. 测试/演示数据脚本（需要检查）${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
TEST_SCRIPTS=$(find "$OWLRD_DB_DIR" -maxdepth 1 -type f \( \
    -name "*demo*.sql" \
    -o -name "*test*.sql" \
    -o -name "*insert*.sql" \
    -o -name "*allocate*.sql" \
    -o -name "*integration*.sql" \
\) | sort)
if [ -n "$TEST_SCRIPTS" ]; then
    echo "$TEST_SCRIPTS" | sed 's|.*/||' | nl -w2 -s'. '
    TEST_COUNT=$(echo "$TEST_SCRIPTS" | wc -l | xargs)
    echo ""
    echo "  共 $TEST_COUNT 个文件"
else
    echo "  无"
fi
echo ""

# 5. 初始化脚本（需要检查）
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}5. 初始化脚本（需要检查）${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
INIT_SCRIPTS=$(find "$OWLRD_DB_DIR" -maxdepth 1 -type f \( \
    -name "*init*.sql" \
    -o -name "*init*.go" \
    -o -name "*init*.sh" \
\) | sort)
if [ -n "$INIT_SCRIPTS" ]; then
    echo "$INIT_SCRIPTS" | sed 's|.*/||' | nl -w2 -s'. '
else
    echo "  无"
fi
echo ""

# 6. Shell 脚本（需要检查）
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}6. Shell 脚本（需要检查）${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
SHELL_SCRIPTS=$(find "$OWLRD_DB_DIR" -maxdepth 1 -name "*.sh" -type f | sort)
if [ -n "$SHELL_SCRIPTS" ]; then
    echo "$SHELL_SCRIPTS" | sed 's|.*/||' | nl -w2 -s'. '
else
    echo "  无"
fi
echo ""

# 7. 文档文件（保留，但可能需要整理）
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}7. 文档文件（保留）${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
DOC_FILES=$(find "$OWLRD_DB_DIR" -maxdepth 1 -name "*.md" -type f | sort)
if [ -n "$DOC_FILES" ]; then
    echo "$DOC_FILES" | sed 's|.*/||' | nl -w2 -s'. '
else
    echo "  无"
fi
echo ""

# 8. 统计
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}文件统计${NC}"
echo -e "${GREEN}========================================${NC}"
TOTAL_FILES=$(find "$OWLRD_DB_DIR" -maxdepth 1 -type f | wc -l | xargs)
echo "总文件数: $TOTAL_FILES"
echo "核心表定义: $CORE_COUNT"
echo ""

# 9. 建议
echo -e "${YELLOW}清理建议：${NC}"
echo "1. 删除系统文件（.DS_Store）"
echo "2. 检查临时/修复脚本是否还需要（如果已应用，可以删除）"
echo "3. 考虑将测试/演示数据脚本移动到测试目录"
echo "4. 检查初始化脚本是否还需要"
echo ""
