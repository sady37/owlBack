#!/bin/bash

cd "$(dirname "$0")"

echo "🛑 Stopping wisefido-data service..."

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "🔍 Looking for wisefido-data processes..."

# 收集所有相关进程 PID（包括父进程和子进程）
PIDS=()

# 方法1: 通过进程名查找
while IFS= read -r pid; do
    if [ -n "$pid" ]; then
        PIDS+=("$pid")
        # 查找该进程的所有子进程
        CHILD_PIDS=$(pgrep -P "$pid" 2>/dev/null || true)
        for child_pid in $CHILD_PIDS; do
            if [ -n "$child_pid" ]; then
                PIDS+=("$child_pid")
            fi
        done
    fi
done < <(pgrep -f "go run.*wisefido-data" 2>/dev/null || true)

while IFS= read -r pid; do
    if [ -n "$pid" ]; then
        PIDS+=("$pid")
        # 查找该进程的所有子进程
        CHILD_PIDS=$(pgrep -P "$pid" 2>/dev/null || true)
        for child_pid in $CHILD_PIDS; do
            if [ -n "$child_pid" ]; then
                PIDS+=("$child_pid")
            fi
        done
    fi
done < <(pgrep -f "wisefido-data" 2>/dev/null || true)

# 去重
PIDS=($(printf '%s\n' "${PIDS[@]}" | sort -u))

# 方法2: 通过端口查找（更可靠）
if command -v lsof &> /dev/null; then
    HTTP_ADDR="${HTTP_ADDR:-:8080}"
    if [[ "$HTTP_ADDR" == *":"* ]]; then
        HTTP_PORT="${HTTP_ADDR##*:}"
    else
        HTTP_PORT="8080"
    fi
    
    PORT_PIDS=$(lsof -ti :$HTTP_PORT 2>/dev/null || true)
    if [ -n "$PORT_PIDS" ]; then
        for pid in $PORT_PIDS; do
            if [ -n "$pid" ]; then
                # 检查进程的工作目录是否是 wisefido-data
                CWD=$(lsof -p "$pid" 2>/dev/null | grep cwd | awk '{print $NF}' || true)
                if echo "$CWD" | grep -q "wisefido-data"; then
                    PIDS+=("$pid")
                    # 查找该进程的所有子进程
                    CHILD_PIDS=$(pgrep -P "$pid" 2>/dev/null || true)
                    for child_pid in $CHILD_PIDS; do
                        if [ -n "$child_pid" ]; then
                            PIDS+=("$child_pid")
                        fi
                    done
                fi
            fi
        done
    fi
fi

# 再次去重
PIDS=($(printf '%s\n' "${PIDS[@]}" | sort -u))

if [ ${#PIDS[@]} -eq 0 ]; then
    echo -e "${YELLOW}  No wisefido-data process found${NC}"
else
    echo -e "${BLUE}  Found ${#PIDS[@]} process(es) to stop${NC}"
    
    # 先尝试优雅停止（发送 TERM 信号）
    for pid in "${PIDS[@]}"; do
        if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
            echo -e "${YELLOW}  Stopping process $pid (TERM)...${NC}"
            kill -TERM "$pid" 2>/dev/null || true
        fi
    done
    
    # 等待进程退出（最多等待 3 秒）
    for i in {1..6}; do
        sleep 0.5
        REMAINING=()
        for pid in "${PIDS[@]}"; do
            if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
                REMAINING+=("$pid")
            fi
        done
        if [ ${#REMAINING[@]} -eq 0 ]; then
            break
        fi
    done
    
    # 如果还有进程在运行，强制杀死
    if [ ${#REMAINING[@]} -gt 0 ]; then
        echo -e "${YELLOW}  Force killing remaining processes...${NC}"
        for pid in "${REMAINING[@]}"; do
            if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
                echo -e "${RED}    Killing process $pid (KILL)...${NC}"
                kill -9 "$pid" 2>/dev/null || true
            fi
        done
        sleep 1
    fi
    
    # 再次检查并强制杀死（使用 pkill 作为备用）
    if pgrep -f "wisefido-data" > /dev/null 2>&1; then
        echo -e "${YELLOW}  Force killing with pkill...${NC}"
        pkill -9 -f "go run.*wisefido-data" 2>/dev/null || true
        pkill -9 -f "wisefido-data" 2>/dev/null || true
        sleep 1
    fi
fi

# 最后检查端口并杀死占用端口的进程
if command -v lsof &> /dev/null; then
    HTTP_ADDR="${HTTP_ADDR:-:8080}"
    if [[ "$HTTP_ADDR" == *":"* ]]; then
        HTTP_PORT="${HTTP_ADDR##*:}"
    else
        HTTP_PORT="8080"
    fi
    
    PORT_PIDS=$(lsof -ti :$HTTP_PORT 2>/dev/null || true)
    if [ -n "$PORT_PIDS" ]; then
        for pid in $PORT_PIDS; do
            if [ -n "$pid" ]; then
                # 检查进程的工作目录是否是 wisefido-data
                CWD=$(lsof -p "$pid" 2>/dev/null | grep cwd | awk '{print $NF}' || true)
                if echo "$CWD" | grep -q "wisefido-data"; then
                    echo -e "${YELLOW}  Force killing process $pid on port $HTTP_PORT...${NC}"
                    kill -9 "$pid" 2>/dev/null || true
                    # 也杀死子进程
                    CHILD_PIDS=$(pgrep -P "$pid" 2>/dev/null || true)
                    for child_pid in $CHILD_PIDS; do
                        if [ -n "$child_pid" ]; then
                            kill -9 "$child_pid" 2>/dev/null || true
                        fi
                    done
                fi
            fi
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
