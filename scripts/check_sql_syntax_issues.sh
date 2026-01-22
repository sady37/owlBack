#!/bin/bash
# 检查 SQL 文件中的语法问题

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
echo -e "${GREEN}检查 SQL 语法问题${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 1. 检查 03_role_permissions.sql 中的 INSERT 语句语法
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}1. 检查 03_role_permissions.sql INSERT 语句${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

FILE="${OWLRD_DB_DIR}/03_role_permissions.sql"

# 检查多余空行（permission_scope 字段后）
if grep -A 3 "permission_scope VARCHAR(10)" "$FILE" | grep -q "^$"; then
    echo -e "${YELLOW}  ⚠️  发现 permission_scope 字段后有多余空行${NC}"
    echo "  位置：第 100-103 行附近"
    grep -A 3 "permission_scope VARCHAR(10)" "$FILE" | head -5
else
    echo -e "${GREEN}  ✓ permission_scope 字段后无多余空行${NC}"
fi

# 检查 INSERT 语句中的逗号问题
echo ""
echo "  检查 INSERT 语句语法..."

# 检查 SystemAdmin 的 device_store 权限（第186行附近）
LINE_186=$(sed -n '186p' "$FILE")
LINE_187=$(sed -n '187p' "$FILE")

if echo "$LINE_186" | grep -q "', 'A'),$" && echo "$LINE_187" | grep -q "^ON CONFLICT"; then
    echo -e "${RED}  ✗ 第186行：最后一个值后有逗号，然后直接是 ON CONFLICT${NC}"
    echo "    当前: $LINE_186"
    echo "    应该删除逗号"
else
    echo -e "${GREEN}  ✓ 第186行：语法正确${NC}"
fi

# 检查 IT 的 config_versions 权限（第460行附近）
LINE_460=$(sed -n '460p' "$FILE")
LINE_461=$(sed -n '461p' "$FILE")

if echo "$LINE_460" | grep -q "', 'A'),$" && echo "$LINE_461" | grep -q "^ON CONFLICT"; then
    echo -e "${RED}  ✗ 第460行：最后一个值后有逗号，然后直接是 ON CONFLICT${NC}"
    echo "    当前: $LINE_460"
    echo "    应该删除逗号"
else
    echo -e "${GREEN}  ✓ 第460行：语法正确${NC}"
fi

echo ""

# 2. 检查 16_device_store.sql 中的外键约束
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}2. 检查 16_device_store.sql 外键约束${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

FILE="${OWLRD_DB_DIR}/16_device_store.sql"

# 检查是否有 ALTER TABLE devices
if grep -q "ALTER TABLE devices" "$FILE"; then
    echo -e "${YELLOW}  ⚠️  发现 ALTER TABLE devices 语句${NC}"
    echo "  位置："
    grep -n "ALTER TABLE devices" "$FILE" | head -3
    
    # 检查文件执行顺序
    DEVICE_STORE_NUM=$(basename "$FILE" | cut -d'_' -f1)
    DEVICES_FILE=$(find "$OWLRD_DB_DIR" -name "17_devices.sql" -o -name "*_devices.sql" | head -1)
    
    if [ -n "$DEVICES_FILE" ]; then
        DEVICES_NUM=$(basename "$DEVICES_FILE" | cut -d'_' -f1)
        echo ""
        echo "  文件执行顺序："
        echo "    $DEVICE_STORE_NUM_device_store.sql (第 $DEVICE_STORE_NUM 个)"
        echo "    $DEVICES_NUM_devices.sql (第 $DEVICES_NUM 个)"
        
        if [ "$DEVICE_STORE_NUM" -lt "$DEVICES_NUM" ]; then
            echo -e "${RED}  ✗ 问题：16_device_store.sql 在 17_devices.sql 之前执行${NC}"
            echo "    当执行 ALTER TABLE devices 时，devices 表还不存在！"
            echo "    应该注释掉或移动到 17_devices.sql 之后"
        else
            echo -e "${GREEN}  ✓ 文件顺序正确${NC}"
        fi
    fi
else
    echo -e "${GREEN}  ✓ 未发现 ALTER TABLE devices 语句${NC}"
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}检查完成${NC}"
echo -e "${GREEN}========================================${NC}"
