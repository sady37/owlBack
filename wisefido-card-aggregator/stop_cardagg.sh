#!/bin/bash

# wisefido-card-aggregator 服务停止脚本

cd "$(dirname "$0")"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}Stopping wisefido-card-aggregator service${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""

echo -e "${BLUE}🔍 Looking for wisefido-card-aggregator processes...${NC}"

# 方法1: 通过进程名查找
AGGREGATOR_PIDS=$(pgrep -f "go run.*wisefido-card-aggregator" 2>/dev/null || true)
AGGREGATOR_PIDS="$AGGREGATOR_PIDS $(pgrep -f "wisefido-card-aggregator" 2>/dev/null || true)"

# 去重
AGGREGATOR_PIDS=$(echo $AGGREGATOR_PIDS | tr ' ' '\n' | sort -u | tr '\n' ' ')

if [ -z "$AGGREGATOR_PIDS" ] || [ "$AGGREGATOR_PIDS" = " " ]; then
    echo -e "${YELLOW}⚠️  No wisefido-card-aggregator process found${NC}"
else
    echo -e "${BLUE}📋 Found processes:${NC}"
    for pid in $AGGREGATOR_PIDS; do
        if [ -n "$pid" ] && [ "$pid" != " " ]; then
            ps -p "$pid" -o pid,command --no-headers 2>/dev/null || true
        fi
    done
    echo ""
    
    # 优雅停止（发送 TERM 信号）
    echo -e "${BLUE}🛑 Stopping processes gracefully...${NC}"
    for pid in $AGGREGATOR_PIDS; do
        if [ -n "$pid" ] && [ "$pid" != " " ]; then
            echo -e "${GREEN}   Sending TERM signal to process $pid${NC}"
            kill -TERM "$pid" 2>/dev/null || true
        fi
    done
    
    # 等待进程退出
    sleep 2
    
    # 检查是否还有进程在运行
    REMAINING=$(pgrep -f "go run.*wisefido-card-aggregator" 2>/dev/null || true)
    REMAINING="$REMAINING $(pgrep -f "wisefido-card-aggregator" 2>/dev/null || true)"
    REMAINING=$(echo $REMAINING | tr ' ' '\n' | sort -u | tr '\n' ' ')
    
    if [ -n "$REMAINING" ] && [ "$REMAINING" != " " ]; then
        echo -e "${YELLOW}⚠️  Some processes are still running, forcing kill...${NC}"
        for pid in $REMAINING; do
            if [ -n "$pid" ] && [ "$pid" != " " ]; then
                echo -e "${YELLOW}   Force killing process $pid${NC}"
                kill -9 "$pid" 2>/dev/null || true
            fi
        done
        sleep 1
    fi
    
    # 最终清理
    pkill -9 -f "go run.*wisefido-card-aggregator" 2>/dev/null || true
    pkill -9 -f "wisefido-card-aggregator" 2>/dev/null || true
    
    echo -e "${GREEN}✅ Service stopped${NC}"
fi

echo ""
echo -e "${BLUE}📊 Current status:${NC}"
echo "  wisefido-card-aggregator processes:"
if pgrep -f "wisefido-card-aggregator" > /dev/null 2>&1; then
    echo -e "    ${RED}❌ Still running${NC}"
    pgrep -f "wisefido-card-aggregator" | xargs ps -o pid,command -p 2>/dev/null || true
else
    echo -e "    ${GREEN}✅ Stopped${NC}"
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Stop completed${NC}"
echo -e "${GREEN}========================================${NC}"
