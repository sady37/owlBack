#!/bin/bash
#
# 启停成对：./start-cardagg.sh ↔ ./stop-cardagg.sh（尽量杀光本模块相关进程）
# systemd：systemctl start owlback.cardagg ↔ systemctl stop owlback.cardagg
#

cd "$(dirname "$0")"

echo "🛑 Stopping wisefido-cardagg service..."

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "🔍 Looking for wisefido-cardagg processes..."

# 收集所有相关进程 PID（包括父进程和子进程）
PIDS=()

# 方法1: 通过进程名查找 (go run)
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
done < <(pgrep -f "go run.*wisefido-cardagg" 2>/dev/null || true)

# 方法2: 通过进程名查找 (binary)
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
done < <(pgrep -f "wisefido-cardagg" 2>/dev/null || true)

# 方法2b: go run 仅带 main.go 路径（cmdline 可能无 wisefido-cardagg 字样）
while IFS= read -r pid; do
    if [ -n "$pid" ]; then
        PIDS+=("$pid")
        CHILD_PIDS=$(pgrep -P "$pid" 2>/dev/null || true)
        for child_pid in $CHILD_PIDS; do
            [ -n "$child_pid" ] && PIDS+=("$child_pid")
        done
    fi
done < <(pgrep -f "cardagg/main.go" 2>/dev/null || true)

# 方法3: 通过日志文件查找（如果 tee 占用日志文件）
if command -v lsof &> /dev/null; then
    OWL_LOG="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")/../.." && pwd)/log"
    LOG_DIR="${LOG_DIR:-$OWL_LOG}"
    LOG_FILE="${CARDAGG_LOG_FILE:-$LOG_DIR/wisefido-cardagg.log}"
    
    if [ -f "$LOG_FILE" ]; then
        LOG_PIDS=$(lsof -t "$LOG_FILE" 2>/dev/null || true)
        if [ -n "$LOG_PIDS" ]; then
            for pid in $LOG_PIDS; do
                if [ -n "$pid" ] && [ "$pid" != "$$" ]; then
                    PIDS+=("$pid")
                    # 查找该进程的所有子进程
                    CHILD_PIDS=$(pgrep -P "$pid" 2>/dev/null || true)
                    for child_pid in $CHILD_PIDS; do
                        if [ -n "$child_pid" ]; then
                            PIDS+=("$child_pid")
                        fi
                    done
                fi
            done
        fi
    fi
fi

# 去重
PIDS=($(printf '%s\n' "${PIDS[@]}" | sort -u))

if [ ${#PIDS[@]} -eq 0 ]; then
    echo -e "${YELLOW}  No wisefido-cardagg process found${NC}"
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
    if pgrep -f "go run.*wisefido-cardagg" > /dev/null 2>&1 || \
       pgrep -f "wisefido-cardagg" > /dev/null 2>&1; then
        echo -e "${YELLOW}  Force killing with pkill...${NC}"
        pkill -9 -f "go run.*wisefido-cardagg" 2>/dev/null || true
        pkill -9 -f "wisefido-cardagg" 2>/dev/null || true
        sleep 1
    fi
fi

echo -e "${GREEN}✅ wisefido-cardagg service stopped${NC}"

echo ""
echo -e "${BLUE}📊 Current status:${NC}"
echo "  wisefido-cardagg processes:"
if pgrep -f "wisefido-cardagg" > /dev/null 2>&1; then
    echo -e "    ${RED}❌ Still running${NC}"
    pgrep -f "wisefido-cardagg" | xargs ps -o pid,command -p 2>/dev/null || true
else
    echo -e "    ${GREEN}✅ Stopped${NC}"
fi

echo ""
