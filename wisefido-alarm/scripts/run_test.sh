#!/bin/bash

# wisefido-alarm 运行测试脚本

set -e

echo "========================================="
echo "wisefido-alarm 运行测试"
echo "========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查环境变量
if [ -z "$TENANT_ID" ]; then
    echo -e "${YELLOW}⚠${NC} TENANT_ID 未设置"
    echo "   请设置: export TENANT_ID=\"your-tenant-id\""
    exit 1
fi

echo -e "${GREEN}✓${NC} TENANT_ID = $TENANT_ID"
echo ""

# 切换到项目目录
cd "$(dirname "$0")/.."

# 检查编译
echo "检查编译..."
if go build -o wisefido-alarm cmd/wisefido-alarm/main.go 2>&1; then
    echo -e "${GREEN}✓${NC} 编译成功"
else
    echo -e "${RED}✗${NC} 编译失败"
    exit 1
fi
echo ""

# 运行服务
echo "========================================="
echo "启动 wisefido-alarm 服务"
echo "========================================="
echo ""
echo "提示："
echo "  - 按 Ctrl+C 停止服务"
echo "  - 服务将每5秒轮询一次卡片"
echo "  - 查看日志了解运行状态"
echo ""
echo "开始运行..."
echo ""

./wisefido-alarm

