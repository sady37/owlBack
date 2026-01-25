#!/bin/bash

cd "$(dirname "$0")"

echo "🛑 Stopping wisefido-data service..."

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "🔍 Looking for wisefido-data processes..."

pkill -f "go run.*wisefido-data" 2>/dev/null
pkill -f "wisefido-data" 2>/dev/null

if pgrep -f "wisefido-data" > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Some processes are still running, forcing kill...${NC}"
    pkill -9 -f "wisefido-data" 2>/dev/null
fi

if command -v lsof &> /dev/null; then
    # 检查 HTTP 端口（从 HTTP_ADDR 解析，默认 8080）
    HTTP_ADDR="${HTTP_ADDR:-:8080}"
    if [[ "$HTTP_ADDR" == *":"* ]]; then
        HTTP_PORT="${HTTP_ADDR##*:}"
    else
        HTTP_PORT="8080"
    fi
    
    if lsof -i :$HTTP_PORT > /dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  Port $HTTP_PORT is still in use, killing processes...${NC}"
        pids=$(lsof -ti :$HTTP_PORT)
        for pid in $pids; do
            kill -9 $pid 2>/dev/null
        done
    fi
fi

echo -e "${GREEN}✅ wisefido-data service stopped${NC}"

echo ""
echo -e "${BLUE}📊 Current status:${NC}"
echo "  wisefido-data processes:"
if pgrep -f "wisefido-data" > /dev/null 2>&1; then
    echo -e "    ${RED}❌ Still running${NC}"
    pgrep -f "wisefido-data" | xargs ps -o pid,command -p
else
    echo -e "    ${GREEN}✅ Stopped${NC}"
fi

echo ""
HTTP_ADDR="${HTTP_ADDR:-:8080}"
if [[ "$HTTP_ADDR" == *":"* ]]; then
    HTTP_PORT="${HTTP_ADDR##*:}"
else
    HTTP_PORT="8080"
fi
echo "  Port $HTTP_PORT (HTTP):"
if command -v lsof &> /dev/null; then
    if lsof -i :$HTTP_PORT > /dev/null 2>&1; then
        echo -e "    ${RED}❌ Still in use${NC}"
        lsof -i :$HTTP_PORT
    else
        echo -e "    ${GREEN}✅ Free${NC}"
    fi
else
    echo "    lsof not available, skipping port check"
fi
