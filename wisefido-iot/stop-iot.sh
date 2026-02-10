#!/bin/bash

# IoT 时序数据服务停止脚本
# 
# 职责：
# 1. 查找并停止 wisefido-iot 服务进程
# 2. 清理相关资源
#
# 使用方法：
#   ./stop-iot.sh          # 正常停止服务
#   ./stop-iot.sh force    # 强制停止服务

cd "$(dirname "$0")"

echo "🛑 Stopping wisefido-iot service..."
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 检查是否有强制停止标志
FORCE_STOP=false
if [ "$1" = "force" ]; then
    FORCE_STOP=true
    echo -e "${YELLOW}⚠️  Force stopping mode enabled${NC}"
    echo ""
fi

# 查找并停止 wisefido-iot 进程
echo "🔍 Looking for wisefido-iot processes..."

# 方式1: 通过进程名查找
IOT_PIDS=$(pgrep -f "go run.*wisefido-iot" 2>/dev/null)
if [ -z "$IOT_PIDS" ]; then
    # 方式2: 通过可执行文件名查找
    IOT_PIDS=$(pgrep -f "wisefido-iot" 2>/dev/null)
fi

if [ -n "$IOT_PIDS" ]; then
    echo -e "${GREEN}✅ Found wisefido-iot processes:${NC}"
    echo "$IOT_PIDS" | while read -r pid; do
        if [ -n "$pid" ]; then
            # 显示进程信息
            ps -p "$pid" -o pid,ppid,cmd 2>/dev/null || echo "  Process ID: $pid"
            
            if [ "$FORCE_STOP" = true ]; then
                # 强制停止进程
                echo -e "  ${YELLOW}Force killing process $pid${NC}"
                kill -9 "$pid" 2>/dev/null || echo -e "  ${RED}Failed to kill process $pid${NC}"
            else
                # 正常停止进程
                echo -e "  ${BLUE}Sending termination signal to process $pid${NC}"
                kill -TERM "$pid" 2>/dev/null || echo -e "  ${RED}Failed to terminate process $pid${NC}"
            fi
        fi
    done
    
    # 等待进程结束
    echo ""
    echo -e "${BLUE}⏳ Waiting for processes to terminate...${NC}"
    if [ "$FORCE_STOP" = true ]; then
        sleep 1
    else
        sleep 3
    fi
    
    # 再次检查进程是否还存在
    REMAINING_PIDS=$(pgrep -f "go run.*wisefido-iot" 2>/dev/null || pgrep -f "wisefido-iot" 2>/dev/null)
    if [ -n "$REMAINING_PIDS" ]; then
        echo -e "${YELLOW}⚠️  Some processes are still running, forcing termination...${NC}"
        echo "$REMAINING_PIDS" | while read -r pid; do
            if [ -n "$pid" ]; then
                echo -e "  Force killing remaining process $pid"
                kill -9 "$pid" 2>/dev/null
            fi
        done
        
        # 再次等待
        sleep 1
    fi
    
    echo -e "${GREEN}✅ wisefido-iot processes stopped${NC}"
else
    echo -e "${YELLOW}⚠️  No wisefido-iot processes found${NC}"
fi

# 检查端口占用情况（wisefido-iot 默认端口）
echo ""
echo -e "${BLUE}🔍 Checking port 8085 status...${NC}"
if command -v lsof &> /dev/null; then
    PORT_PIDS=$(lsof -ti :8085 2>/dev/null || true)
    if [ -n "$PORT_PIDS" ]; then
        echo -e "${RED}❌ Port 8085 is still in use:${NC}"
        for pid in $PORT_PIDS; do
            if [ -n "$pid" ]; then
                CWD=$(lsof -p "$pid" 2>/dev/null | grep cwd | awk '{print $NF}' || true)
                CMD=$(ps -p "$pid" -o args= 2>/dev/null || echo "Unknown process")
                
                # 检查是否与wisefido-iot相关
                if echo "$CWD" | grep -q "wisefido-iot" || echo "$CMD" | grep -q "wisefido-iot"; then
                    echo -e "  ${YELLOW}Process $pid (wisefido-iot related):${NC} $CMD"
                    if [ "$FORCE_STOP" = true ]; then
                        kill -9 "$pid" 2>/dev/null
                        echo -e "  ${YELLOW}Force killed process $pid${NC}"
                    else
                        kill -TERM "$pid" 2>/dev/null
                        echo -e "  ${BLUE}Sent termination signal to process $pid${NC}"
                    fi
                fi
            fi
        done
    else
        echo -e "${GREEN}✅ Port 8085 is free${NC}"
    fi
else
    echo "lsof not available, skipping port check"
fi

# 清理完成
echo ""
echo -e "${GREEN}🎉 wisefido-iot service stop process completed${NC}"

# 验证是否还有wisefido-iot进程在运行
FINAL_CHECK=$(pgrep -f "go run.*wisefido-iot" 2>/dev/null || pgrep -f "wisefido-iot" 2>/dev/null)
if [ -n "$FINAL_CHECK" ]; then
    echo ""
    echo -e "${RED}❌ Warning: Some wisefido-iot processes may still be running${NC}"
    echo "  Run 'ps aux | grep wisefido-iot' to check"
    if [ "$FORCE_STOP" = false ]; then
        echo "  Try running './stop-iot.sh force' to force stop"
    fi
else
    echo -e "${GREEN}✅ All wisefido-iot processes have been stopped${NC}"
fi