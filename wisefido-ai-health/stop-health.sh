#!/bin/bash
#
# 停止 wisefido-ai-health 守护进程
#

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PIDS=$(pgrep -f "wisefido-ai-health" | grep -v $$ || true)
if [ -z "$PIDS" ]; then
    echo -e "${YELLOW}wisefido-ai-health 未运行${NC}"
    exit 0
fi

echo -e "${GREEN}停止进程: $PIDS${NC}"
kill -TERM $PIDS 2>/dev/null || true
sleep 1

# 兜底强杀
PIDS=$(pgrep -f "wisefido-ai-health" | grep -v $$ || true)
if [ -n "$PIDS" ]; then
    echo -e "${YELLOW}TERM 未生效，发送 KILL${NC}"
    kill -KILL $PIDS 2>/dev/null || true
fi
echo -e "${GREEN}✅ wisefido-ai-health 已停止${NC}"
