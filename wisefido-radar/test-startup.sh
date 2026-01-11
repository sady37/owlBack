#!/bin/bash

# 雷达服务启动测试脚本
# 测试服务启动和订阅管理器初始化

set -e

cd "$(dirname "$0")"

echo "========================================="
echo "雷达服务启动测试"
echo "========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 设置环境变量（端口 5433 与 start_owlback.sh 保持一致）
export DB_HOST=127.0.0.1
export DB_PORT=5433
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=owlrd
export DB_SSLMODE=disable

export REDIS_ADDR=127.0.0.1:6379
export REDIS_PASSWORD=TeLunSu-36kr
export REDIS_DB=0

export MQTT_BROKER=tcp://127.0.0.1:1883
export MQTT_CLIENT_ID=wisefido-radar-test
export MQTT_USERNAME=wisefido
export MQTT_PASSWORD=

# 订阅配置
export RADAR_SUBSCRIPTION_AUTO=true
export RADAR_SUBSCRIPTION_DURATION=3600
export RADAR_SUBSCRIPTION_CONTENT=0
export RADAR_SUBSCRIPTION_RENEWAL_INTERVAL=50
export RADAR_SUBSCRIPTION_RENEWAL_ADVANCE=10

export LOG_LEVEL=info
export LOG_FORMAT=json

echo -e "${BLUE}环境变量配置:${NC}"
echo "  DB_HOST: $DB_HOST"
echo "  REDIS_ADDR: $REDIS_ADDR"
echo "  MQTT_BROKER: $MQTT_BROKER"
echo "  订阅自动: $RADAR_SUBSCRIPTION_AUTO"
echo "  订阅时长: $RADAR_SUBSCRIPTION_DURATION 秒"
echo "  续订间隔: $RADAR_SUBSCRIPTION_RENEWAL_INTERVAL 分钟"
echo ""

# 检查依赖服务
echo -e "${BLUE}检查依赖服务...${NC}"

# 检查 PostgreSQL
if command -v psql &> /dev/null; then
    if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT 1" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ PostgreSQL 连接正常${NC}"
    else
        echo -e "${RED}✗ PostgreSQL 连接失败${NC}"
        echo -e "${YELLOW}  提示: 请确保 PostgreSQL 服务正在运行${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  psql 未安装，跳过 PostgreSQL 检查${NC}"
fi

# 检查 Redis
if command -v redis-cli &> /dev/null; then
    if redis-cli -a "$REDIS_PASSWORD" ping > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Redis 连接正常${NC}"
    else
        echo -e "${RED}✗ Redis 连接失败${NC}"
        echo -e "${YELLOW}  提示: 请确保 Redis 服务正在运行${NC}"
        exit 1
    fi
else
    echo -e "${YELLOW}⚠️  redis-cli 未安装，跳过 Redis 检查${NC}"
fi

# 检查 MQTT（可选）
if command -v mosquitto_sub &> /dev/null; then
    echo -e "${YELLOW}⚠️  MQTT 检查需要手动验证${NC}"
else
    echo -e "${YELLOW}⚠️  mosquitto_sub 未安装，跳过 MQTT 检查${NC}"
fi

echo ""

# 测试服务启动（短暂运行，然后停止）
echo -e "${BLUE}测试服务启动（5秒后自动停止）...${NC}"
echo -e "${YELLOW}提示: 这将启动服务并运行 5 秒，用于验证订阅管理器是否正常启动${NC}"
echo ""

# 创建临时日志文件
LOG_FILE="/tmp/radar_test_$(date +%s).log"

# 在后台启动服务
timeout 5 go run ./cmd/wisefido-radar/main.go > "$LOG_FILE" 2>&1 &
SERVICE_PID=$!

# 等待服务启动
sleep 2

# 检查服务是否还在运行
if ps -p $SERVICE_PID > /dev/null 2>&1; then
    echo -e "${GREEN}✓ 服务已启动 (PID: $SERVICE_PID)${NC}"
    
    # 检查日志中是否有订阅管理器启动信息
    if grep -q "Subscription manager started" "$LOG_FILE" 2>/dev/null; then
        echo -e "${GREEN}✓ 订阅管理器已启动${NC}"
        
        # 显示订阅管理器启动日志
        echo ""
        echo -e "${BLUE}订阅管理器启动日志:${NC}"
        grep "Subscription manager" "$LOG_FILE" | head -3
    else
        echo -e "${YELLOW}⚠️  订阅管理器启动日志未找到${NC}"
        echo -e "${YELLOW}  查看完整日志: cat $LOG_FILE${NC}"
    fi
    
    # 等待服务自然停止（timeout 会杀死它）
    wait $SERVICE_PID 2>/dev/null || true
else
    echo -e "${RED}✗ 服务启动失败${NC}"
    echo -e "${YELLOW}查看日志:${NC}"
    cat "$LOG_FILE" | tail -20
    exit 1
fi

echo ""

# 显示关键日志
echo -e "${BLUE}关键日志信息:${NC}"
if [ -f "$LOG_FILE" ]; then
    echo "--- 服务启动日志 ---"
    grep -E "(Starting|Subscription manager|radar service)" "$LOG_FILE" | head -5
    echo ""
    echo "--- 错误日志（如有） ---"
    grep -i "error\|fatal\|panic" "$LOG_FILE" | head -5 || echo "  无错误"
fi

echo ""

# 清理
rm -f "$LOG_FILE"

echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}启动测试完成！${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
echo -e "${BLUE}下一步：${NC}"
echo "  1. 启动完整服务: ./start-radar.sh"
echo "  2. 查看实时日志，确认订阅管理器运行正常"
echo "  3. 等待设备连接，观察自动订阅日志"
echo ""
