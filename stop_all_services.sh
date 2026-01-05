#!/bin/bash

# 统一停止脚本：停止 wisefido-data 和 wisefido-card-aggregator 服务
#
# 架构说明：
# - wisefido-data: 数据管理 API + 卡片创建/更新（直接同步调用 CardCreator）
# - wisefido-card-aggregator: 数据聚合（从 PostgreSQL + Redis 聚合卡片数据并缓存到 Redis）

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}Stopping OwlBack Services${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""

# 查找并停止 wisefido-data 服务
echo -e "${BLUE}Stopping wisefido-data service...${NC}"

# 方法1: 通过进程名查找
DATA_PIDS=$(pgrep -f "go run.*wisefido-data" 2>/dev/null || true)
DATA_PIDS="$DATA_PIDS $(pgrep -f "wisefido-data" 2>/dev/null || true)"

# 方法2: 通过端口 8080 查找（更可靠，因为 go run 的进程名是 main）
if command -v lsof &> /dev/null; then
    PORT_PIDS=$(lsof -ti :8080 2>/dev/null || true)
    if [ -n "$PORT_PIDS" ]; then
        # 检查这些进程的工作目录是否是 wisefido-data
        for pid in $PORT_PIDS; do
            if [ -n "$pid" ]; then
                CWD=$(lsof -p "$pid" 2>/dev/null | grep cwd | awk '{print $NF}' || true)
                if echo "$CWD" | grep -q "wisefido-data"; then
                    DATA_PIDS="$DATA_PIDS $pid"
                fi
            fi
        done
    fi
fi

# 去重
DATA_PIDS=$(echo $DATA_PIDS | tr ' ' '\n' | sort -u | tr '\n' ' ')

if [ -z "$DATA_PIDS" ] || [ "$DATA_PIDS" = " " ]; then
    echo -e "${YELLOW}  No wisefido-data process found${NC}"
else
    for pid in $DATA_PIDS; do
        if [ -n "$pid" ] && [ "$pid" != " " ]; then
            echo -e "${GREEN}  Stopping process $pid${NC}"
            kill -TERM "$pid" 2>/dev/null || kill -9 "$pid" 2>/dev/null || true
        fi
    done
    # 等待进程退出
    sleep 2
    # 强制杀死残留进程（包括通过端口查找）
    if command -v lsof &> /dev/null; then
        PORT_PIDS=$(lsof -ti :8080 2>/dev/null || true)
        for pid in $PORT_PIDS; do
            if [ -n "$pid" ]; then
                CWD=$(lsof -p "$pid" 2>/dev/null | grep cwd | awk '{print $NF}' || true)
                if echo "$CWD" | grep -q "wisefido-data"; then
                    echo -e "${YELLOW}  Force killing process $pid on port 8080${NC}"
                    kill -9 "$pid" 2>/dev/null || true
                fi
            fi
        done
    fi
    pkill -9 -f "go run.*wisefido-data" 2>/dev/null || true
    pkill -9 -f "wisefido-data" 2>/dev/null || true
    echo -e "${GREEN}  wisefido-data stopped${NC}"
fi

echo ""

# 查找并停止 wisefido-card-aggregator 服务
echo -e "${BLUE}Stopping wisefido-card-aggregator service...${NC}"
AGGREGATOR_PIDS=$(pgrep -f "go run.*wisefido-card-aggregator" || pgrep -f "wisefido-card-aggregator" || true)
if [ -z "$AGGREGATOR_PIDS" ]; then
    echo -e "${YELLOW}  No wisefido-card-aggregator process found${NC}"
else
    for pid in $AGGREGATOR_PIDS; do
        echo -e "${GREEN}  Stopping process $pid${NC}"
        kill -TERM "$pid" 2>/dev/null || kill -9 "$pid" 2>/dev/null || true
    done
    # 等待进程退出
    sleep 2
    # 强制杀死残留进程
    pkill -9 -f "go run.*wisefido-card-aggregator" 2>/dev/null || true
    pkill -9 -f "wisefido-card-aggregator" 2>/dev/null || true
    echo -e "${GREEN}  wisefido-card-aggregator stopped${NC}"
fi

echo ""

# 验证所有进程已停止
echo -e "${BLUE}Verifying all services are stopped...${NC}"
REMAINING=$(pgrep -f "wisefido-data\|wisefido-card-aggregator" || true)
if [ -z "$REMAINING" ]; then
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}All services stopped successfully${NC}"
    echo -e "${GREEN}========================================${NC}"
else
    echo -e "${RED}Warning: Some processes may still be running:${NC}"
    echo "$REMAINING"
    echo -e "${YELLOW}Attempting force kill...${NC}"
    pkill -9 -f "wisefido-data\|wisefido-card-aggregator" 2>/dev/null || true
    sleep 1
    REMAINING2=$(pgrep -f "wisefido-data\|wisefido-card-aggregator" || true)
    if [ -z "$REMAINING2" ]; then
        echo -e "${GREEN}All processes terminated${NC}"
    else
        echo -e "${RED}Some processes could not be terminated:${NC}"
        echo "$REMAINING2"
        exit 1
    fi
fi

