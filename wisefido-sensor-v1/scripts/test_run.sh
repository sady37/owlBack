#!/bin/bash

# wisefido-alarm 测试运行脚本（带超时）

set -e

echo "========================================="
echo "wisefido-alarm 测试运行"
echo "========================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 检查环境变量
if [ -z "$TENANT_ID" ]; then
    echo -e "${YELLOW}⚠${NC} TENANT_ID 未设置，使用默认值: test-tenant"
    export TENANT_ID="test-tenant"
fi

echo -e "${GREEN}✓${NC} TENANT_ID = $TENANT_ID"
echo ""

# 切换到项目目录
cd "$(dirname "$0")/.."

# 检查可执行文件
if [ ! -f "./wisefido-alarm" ]; then
    echo "编译服务..."
    go build -o wisefido-alarm cmd/wisefido-alarm/main.go
fi

echo "========================================="
echo "启动 wisefido-alarm 服务（10秒后自动停止）"
echo "========================================="
echo ""
echo "提示："
echo "  - 服务将运行10秒后自动停止（用于测试）"
echo "  - 查看日志了解运行状态"
echo "  - 如需长时间运行，请直接运行: ./wisefido-alarm"
echo ""
echo "开始运行..."
echo ""

# 在后台运行服务
./wisefido-alarm > /tmp/wisefido-alarm-test.log 2>&1 &
SERVICE_PID=$!

# 等待10秒
sleep 10

# 停止服务
echo ""
echo "停止服务..."
kill $SERVICE_PID 2>/dev/null || true
wait $SERVICE_PID 2>/dev/null || true

echo ""
echo "========================================="
echo "服务日志（最后20行）"
echo "========================================="
tail -20 /tmp/wisefido-alarm-test.log || echo "无日志输出"

echo ""
echo "========================================="
echo "测试完成"
echo "========================================="

