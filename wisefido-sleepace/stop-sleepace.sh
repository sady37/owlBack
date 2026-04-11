#!/bin/bash
#
# 启停成对：./start-sleepace.sh ↔ ./stop-sleepace.sh
# systemd：systemctl start owlback.sleepace ↔ systemctl stop owlback.sleepace
#

cd "$(dirname "$0")"

echo "Stopping wisefido-sleepace service..."

# 手动执行时：若 unit 仍 active，先 systemctl stop，否则 pkill 后 Restart= 会立刻拉起。
# systemd 执行本脚本作为 ExecStop 时已处于 stop 流程，禁止再 systemctl stop 同 unit（死锁）。
if [ -z "${INVOCATION_ID:-}" ] && command -v systemctl >/dev/null 2>&1; then
    if systemctl is-active --quiet owlback.sleepace 2>/dev/null; then
        echo "owlback.sleepace is active — stopping unit first..."
        systemctl stop owlback.sleepace 2>/dev/null || true
        sleep 1
    fi
fi

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "Looking for wisefido-sleepace processes..."

pkill -f "go run.*wisefido-sleepace" 2>/dev/null
pkill -f "wisefido-sleepace" 2>/dev/null

if pgrep -f "wisefido-sleepace" > /dev/null 2>&1; then
    echo -e "${YELLOW}Some processes are still running, forcing kill...${NC}"
    pkill -9 -f "wisefido-sleepace" 2>/dev/null
fi

if command -v lsof &> /dev/null; then
    if lsof -nP -iTCP:8083 -sTCP:LISTEN > /dev/null 2>&1; then
        echo -e "${YELLOW}Port 8083 still has LISTEN, killing listener PIDs...${NC}"
        pids=$(lsof -nP -tiTCP:8083 -sTCP:LISTEN 2>/dev/null || true)
        for pid in $pids; do
            kill -9 $pid 2>/dev/null
        done
    fi
fi

echo -e "${GREEN}wisefido-sleepace service stopped${NC}"

echo ""
echo -e "${BLUE}Current status:${NC}"
echo "  wisefido-sleepace processes:"
if pgrep -f "wisefido-sleepace" > /dev/null 2>&1; then
    echo -e "    ${RED}Still running${NC}"
    pgrep -f "wisefido-sleepace" | xargs ps -o pid,command -p
else
    echo -e "    ${GREEN}Stopped${NC}"
fi

echo ""
echo "  Port 8083 (HTTP):"
if command -v lsof &> /dev/null; then
    if lsof -nP -iTCP:8083 -sTCP:LISTEN > /dev/null 2>&1; then
        echo -e "    ${RED}Still in use (LISTEN)${NC}"
        lsof -nP -iTCP:8083 -sTCP:LISTEN
    else
        echo -e "    ${GREEN}Free${NC}"
    fi
else
    echo "    lsof not available, skipping port check"
fi
