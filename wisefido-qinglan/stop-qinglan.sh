#!/bin/bash

# wisefido-qinglan 停止脚本

cd "$(dirname "$0")"

echo "🛑 Stopping wisefido-qinglan service..."

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 停止服务
echo "🔍 Looking for wisefido-qinglan processes..."

# 查找并停止 go run 进程
pkill -f "go run.*wisefido-qinglan" 2>/dev/null

# 查找并停止编译后的二进制进程
pkill -f "wisefido-qinglan" 2>/dev/null

# 检查是否还有进程在运行
if pgrep -f "wisefido-qinglan" > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Some processes are still running, forcing kill...${NC}"
    pkill -9 -f "wisefido-qinglan" 2>/dev/null
fi

# 检查端口 8081 是否被占用
if command -v lsof &> /dev/null; then
    if lsof -i :8081 > /dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  Port 8081 is still in use, killing processes...${NC}"
        pids=$(lsof -ti :8081)
        for pid in $pids; do
            echo "  Killing PID: $pid"
            kill -9 $pid 2>/dev/null
        done
    fi
fi

echo -e "${GREEN}✅ wisefido-qinglan service stopped${NC}"

# 显示状态
echo ""
echo -e "${BLUE}📊 Current status:${NC}"
echo "  wisefido-qinglan processes:"
if pgrep -f "wisefido-qinglan" > /dev/null 2>&1; then
    echo -e "    ${RED}❌ Still running${NC}"
    pgrep -f "wisefido-qinglan" | xargs ps -o pid,command -p
else
    echo -e "    ${GREEN}✅ Stopped${NC}"
fi

echo ""
echo "  Port 8081 (HTTP):"
if command -v lsof &> /dev/null; then
    if lsof -i :8081 > /dev/null 2>&1; then
        echo -e "    ${RED}❌ Still in use${NC}"
        lsof -i :8081
    else
        echo -e "    ${GREEN}✅ Free${NC}"
    fi
else
    echo "    lsof not available, skipping port check"
fi