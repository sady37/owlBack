#!/bin/bash
# 清理 owlRD/db 目录下的文件

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
echo -e "${GREEN}清理 owlRD/db 目录文件${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 1. 系统文件（可以安全删除）
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}1. 系统文件（可以安全删除）${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
SYSTEM_FILES=(
    ".DS_Store"
)
for file in "${SYSTEM_FILES[@]}"; do
    if [ -f "$OWLRD_DB_DIR/$file" ]; then
        echo "  - $file"
    fi
done
echo ""

# 2. 临时/修复脚本（需要确认）
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}2. 临时/修复脚本（已应用的可以删除）${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
TEMP_SCRIPTS=(
    "check_admin_manager_permissions.sql"
    "check_admin_manager_permissions_summary.sql"
    "check_branches_permissions.sql"
    "check_units_table_structure.sql"
    "clean_and_rebuild_buildings.sql"
    "clean_buildings_and_units_only.sql"
    "clear_branch_unit_building.sql"
    "clear_buildings_units.sql"
    "delete_family_role.sql"
    "diagnose_card_generation.sql"
    "enable_device_monitoring.sql"
    "fix_admin_manager_branches_permissions.sql"
    "fix_branches_permissions.sql"
    "migration_remove_users_branch_id.sql"
    "rebuild_fresh_data.sql"
    "rebuild_units_table.sql"
    "update_manager_branch_only.sql"
)
echo "以下脚本可能是临时修复脚本，如果已应用可以删除："
for file in "${TEMP_SCRIPTS[@]}"; do
    if [ -f "$OWLRD_DB_DIR/$file" ]; then
        echo "  - $file"
    fi
done
echo ""

# 3. 测试/演示数据脚本（可能需要保留或移动到测试目录）
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}3. 测试/演示数据脚本（建议保留或移动到测试目录）${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
TEST_SCRIPTS=(
    "allocate_sleepad_to_demo.sql"
    "insert_default_branches_auto.sql"
    "insert_default_branches_executable.sql"
    "insert_default_branches_for_3_tenants.sql"
    "insert_default_branches_for_tenants.sql"
    "insert_demo_devices.sql"
    "integration_test_data.sql"
    "reload_demo_devices.sql"
)
echo "以下脚本是测试/演示数据，建议保留或移动到测试目录："
for file in "${TEST_SCRIPTS[@]}"; do
    if [ -f "$OWLRD_DB_DIR/$file" ]; then
        echo "  - $file"
    fi
done
echo ""

# 询问是否删除系统文件
echo -e "${RED}警告: 删除文件前请确认这些文件不再需要${NC}"
echo ""
read -p "是否删除系统文件 (.DS_Store)？(yes/no): " confirm_system

if [ "$confirm_system" = "yes" ]; then
    for file in "${SYSTEM_FILES[@]}"; do
        if [ -f "$OWLRD_DB_DIR/$file" ]; then
            echo "删除: $file"
            rm "$OWLRD_DB_DIR/$file"
        fi
    done
    echo -e "${GREEN}系统文件已删除${NC}"
fi

echo ""
read -p "是否删除临时/修复脚本？(yes/no): " confirm_temp

if [ "$confirm_temp" = "yes" ]; then
    echo ""
    echo "请输入要删除的文件编号（用空格分隔，例如: 1 3 5），或输入 'all' 删除所有："
    echo ""
    # 显示文件列表
    idx=1
    for file in "${TEMP_SCRIPTS[@]}"; do
        if [ -f "$OWLRD_DB_DIR/$file" ]; then
            echo "  $idx. $file"
            idx=$((idx + 1))
        fi
    done
    echo ""
    read -p "> " file_nums
    
    if [ "$file_nums" = "all" ]; then
        for file in "${TEMP_SCRIPTS[@]}"; do
            if [ -f "$OWLRD_DB_DIR/$file" ]; then
                echo "删除: $file"
                rm "$OWLRD_DB_DIR/$file"
            fi
        done
    else
        idx=1
        for file in "${TEMP_SCRIPTS[@]}"; do
            if [ -f "$OWLRD_DB_DIR/$file" ]; then
                for num in $file_nums; do
                    if [ "$num" = "$idx" ]; then
                        echo "删除: $file"
                        rm "$OWLRD_DB_DIR/$file"
                        break
                    fi
                done
                idx=$((idx + 1))
            fi
        done
    fi
    echo -e "${GREEN}选定的临时脚本已删除${NC}"
fi

echo ""
echo -e "${GREEN}清理完成${NC}"
echo ""
echo "剩余文件统计："
TOTAL_FILES=$(find "$OWLRD_DB_DIR" -maxdepth 1 -type f | wc -l | xargs)
echo "总文件数: $TOTAL_FILES"
