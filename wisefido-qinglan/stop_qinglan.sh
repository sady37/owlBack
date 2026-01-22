#!/bin/bash

cd "$(dirname "$0")"

echo "🛑 Stopping wisefido-qinglan service..."

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "🔍 Looking for wisefido-qinglan processes..."

pkill -f "go run.*wisefido-qinglan" 2>/dev/null
pkill -f "wisefido-qinglan" 2>/dev/null

if pgrep -f "wisefido-qinglan" > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Some processes are still running, forcing kill...${NC}"
    pkill -9 -f "wisefido-qinglan" 2>/dev/null
fi

if command -v lsof &> /dev/null; then
    if lsof -i :8081 > /dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  Port 8081 is still in use, killing processes...${NC}"
        pids=$(lsof -ti :8081)
        for pid in $pids; do
            kill -9 $pid 2>/dev/null
        done
    fi
fi

echo -e "${GREEN}✅ wisefido-qinglan service stopped${NC}"

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
