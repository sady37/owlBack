#!/bin/bash

# 统一停止脚本：停止所有后台服务
#
# 架构说明：
# - wisefido-data: 数据管理 API + 卡片创建/更新（直接同步调用 CardCreator）
# - wisefido-card-aggregator: 数据聚合（从 PostgreSQL + Redis 聚合卡片数据并缓存到 Redis）
# - wisefido-card-manage: 卡片管理服务（提供 HTTP API 用于创建/更新卡片）
# - wisefido-iot-timeseries: 数据消费服务（从 Redis Streams 消费数据，存储到 TimescaleDB）
# - wisefido-ai: AI 智能推理服务（高级推理、访客识别、巡房优化）
#
# 注意：设备接入服务（wisefido-sleepace, wisefido-radar）不包含在此脚本中
# 这些服务需要手动停止

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

# 查找并停止 wisefido-iot-timeseries 服务
echo -e "${BLUE}Stopping wisefido-iot-timeseries service...${NC}"

# 方法1: 通过进程名查找
IOT_TIMESERIES_PIDS=$(pgrep -f "go run.*wisefido-iot-timeseries" 2>/dev/null || true)
IOT_TIMESERIES_PIDS="$IOT_TIMESERIES_PIDS $(pgrep -f "wisefido-iot-timeseries" 2>/dev/null || true)"

# 方法2: 通过端口 8083 查找
if command -v lsof &> /dev/null; then
    PORT_PIDS=$(lsof -ti :8083 2>/dev/null || true)
    if [ -n "$PORT_PIDS" ]; then
        # 检查这些进程的工作目录是否是 wisefido-iot-timeseries
        for pid in $PORT_PIDS; do
            if [ -n "$pid" ]; then
                CWD=$(lsof -p "$pid" 2>/dev/null | grep cwd | awk '{print $NF}' || true)
                if echo "$CWD" | grep -q "wisefido-iot-timeseries"; then
                    IOT_TIMESERIES_PIDS="$IOT_TIMESERIES_PIDS $pid"
                fi
            fi
        done
    fi
fi

# 去重
IOT_TIMESERIES_PIDS=$(echo $IOT_TIMESERIES_PIDS | tr ' ' '\n' | sort -u | tr '\n' ' ')

if [ -z "$IOT_TIMESERIES_PIDS" ] || [ "$IOT_TIMESERIES_PIDS" = " " ]; then
    echo -e "${YELLOW}  No wisefido-iot-timeseries process found${NC}"
else
    for pid in $IOT_TIMESERIES_PIDS; do
        if [ -n "$pid" ] && [ "$pid" != " " ]; then
            echo -e "${GREEN}  Stopping process $pid${NC}"
            kill -TERM "$pid" 2>/dev/null || kill -9 "$pid" 2>/dev/null || true
        fi
    done
    # 等待进程退出
    sleep 2
    # 强制杀死残留进程（包括通过端口查找）
    if command -v lsof &> /dev/null; then
        PORT_PIDS=$(lsof -ti :8083 2>/dev/null || true)
        for pid in $PORT_PIDS; do
            if [ -n "$pid" ]; then
                CWD=$(lsof -p "$pid" 2>/dev/null | grep cwd | awk '{print $NF}' || true)
                if echo "$CWD" | grep -q "wisefido-iot-timeseries"; then
                    echo -e "${YELLOW}  Force killing process $pid on port 8083${NC}"
                    kill -9 "$pid" 2>/dev/null || true
                fi
            fi
        done
    fi
    pkill -9 -f "go run.*wisefido-iot-timeseries" 2>/dev/null || true
    pkill -9 -f "wisefido-iot-timeseries" 2>/dev/null || true
    echo -e "${GREEN}  wisefido-iot-timeseries stopped${NC}"
fi

echo ""

# 查找并停止 wisefido-ai 服务
echo -e "${BLUE}Stopping wisefido-ai service...${NC}"
AI_PIDS=$(pgrep -f "go run.*wisefido-ai" 2>/dev/null || true)
AI_PIDS="$AI_PIDS $(pgrep -f "wisefido-ai" 2>/dev/null || true)"

# 去重
AI_PIDS=$(echo $AI_PIDS | tr ' ' '\n' | sort -u | tr '\n' ' ')

if [ -z "$AI_PIDS" ] || [ "$AI_PIDS" = " " ]; then
    echo -e "${YELLOW}  No wisefido-ai process found${NC}"
else
    for pid in $AI_PIDS; do
        if [ -n "$pid" ] && [ "$pid" != " " ]; then
            echo -e "${GREEN}  Stopping process $pid${NC}"
            kill -TERM "$pid" 2>/dev/null || kill -9 "$pid" 2>/dev/null || true
        fi
    done
    # 等待进程退出
    sleep 2
    # 强制杀死残留进程
    pkill -9 -f "go run.*wisefido-ai" 2>/dev/null || true
    pkill -9 -f "wisefido-ai" 2>/dev/null || true
    echo -e "${GREEN}  wisefido-ai stopped${NC}"
fi

echo ""

# 验证所有进程已停止
echo -e "${BLUE}Verifying all services are stopped...${NC}"
REMAINING=$(pgrep -f "wisefido-data\|wisefido-card-aggregator\|wisefido-card-manage\|wisefido-iot-timeseries\|wisefido-ai" || true)
if [ -z "$REMAINING" ]; then
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}All backend services stopped successfully${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo -e "${YELLOW}Note: Device access services (wisefido-sleepace, wisefido-radar) are not stopped by this script${NC}"
    echo -e "${YELLOW}  Stop them manually if needed${NC}"
else
    echo -e "${RED}Warning: Some processes may still be running:${NC}"
    echo "$REMAINING"
    echo -e "${YELLOW}Attempting force kill...${NC}"
    pkill -9 -f "wisefido-data\|wisefido-card-aggregator\|wisefido-card-manage\|wisefido-iot-timeseries\|wisefido-ai" 2>/dev/null || true
    sleep 1
    REMAINING2=$(pgrep -f "wisefido-data\|wisefido-card-aggregator\|wisefido-card-manage\|wisefido-iot-timeseries\|wisefido-ai" || true)
    if [ -z "$REMAINING2" ]; then
        echo -e "${GREEN}All processes terminated${NC}"
    else
        echo -e "${RED}Some processes could not be terminated:${NC}"
        echo "$REMAINING2"
        exit 1
    fi
fi

