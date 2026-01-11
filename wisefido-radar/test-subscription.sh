#!/bin/bash

# 雷达订阅功能测试脚本
# 用于测试自动订阅和自动续订功能

set -e

cd "$(dirname "$0")"

echo "========================================="
echo "雷达订阅功能测试"
echo "========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 检查 Redis 连接
echo -e "${BLUE}[1/5] 检查 Redis 连接...${NC}"
export REDIS_ADDR=127.0.0.1:6379
export REDIS_PASSWORD=TeLunSu-36kr

if command -v redis-cli &> /dev/null; then
    if redis-cli -a "$REDIS_PASSWORD" ping > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Redis 连接正常${NC}"
    else
        echo -e "${RED}✗ Redis 连接失败${NC}"
        exit 1
    fi
else
    echo -e "${YELLOW}⚠️  redis-cli 未安装，跳过连接检查${NC}"
fi

echo ""

# 检查订阅配置
echo -e "${BLUE}[2/5] 检查订阅配置...${NC}"
export RADAR_SUBSCRIPTION_AUTO=${RADAR_SUBSCRIPTION_AUTO:-true}
export RADAR_SUBSCRIPTION_DURATION=${RADAR_SUBSCRIPTION_DURATION:-3600}
export RADAR_SUBSCRIPTION_CONTENT=${RADAR_SUBSCRIPTION_CONTENT:-0}
export RADAR_SUBSCRIPTION_RENEWAL_INTERVAL=${RADAR_SUBSCRIPTION_RENEWAL_INTERVAL:-50}
export RADAR_SUBSCRIPTION_RENEWAL_ADVANCE=${RADAR_SUBSCRIPTION_RENEWAL_ADVANCE:-10}

echo "  自动订阅: $RADAR_SUBSCRIPTION_AUTO"
echo "  订阅时长: $RADAR_SUBSCRIPTION_DURATION 秒"
echo "  订阅内容: $RADAR_SUBSCRIPTION_CONTENT (0=同时订阅, 1=轨迹, 2=呼吸心率)"
echo "  续订间隔: $RADAR_SUBSCRIPTION_RENEWAL_INTERVAL 分钟"
echo "  提前续订: $RADAR_SUBSCRIPTION_RENEWAL_ADVANCE 分钟"
echo ""

# 编译验证
echo -e "${BLUE}[3/5] 编译验证...${NC}"
if go build ./cmd/wisefido-radar/ 2>&1; then
    echo -e "${GREEN}✓ 编译成功${NC}"
else
    echo -e "${RED}✗ 编译失败${NC}"
    exit 1
fi
echo ""

# 检查订阅管理器代码
echo -e "${BLUE}[4/5] 检查订阅管理器代码...${NC}"
if [ -f "internal/service/subscription_manager.go" ]; then
    echo -e "${GREEN}✓ subscription_manager.go 存在${NC}"
    
    # 检查关键方法
    if grep -q "AutoSubscribe" internal/service/subscription_manager.go; then
        echo -e "${GREEN}✓ AutoSubscribe 方法存在${NC}"
    else
        echo -e "${RED}✗ AutoSubscribe 方法不存在${NC}"
        exit 1
    fi
    
    if grep -q "checkAndRenewSubscriptions" internal/service/subscription_manager.go; then
        echo -e "${GREEN}✓ checkAndRenewSubscriptions 方法存在${NC}"
    else
        echo -e "${RED}✗ checkAndRenewSubscriptions 方法不存在${NC}"
        exit 1
    fi
else
    echo -e "${RED}✗ subscription_manager.go 不存在${NC}"
    exit 1
fi
echo ""

# 检查 Redis 中的订阅状态（如果有）
echo -e "${BLUE}[5/5] 检查 Redis 订阅状态...${NC}"
if command -v redis-cli &> /dev/null; then
    # 查找所有订阅 key
    subscription_keys=$(redis-cli -a "$REDIS_PASSWORD" --scan --pattern "radar:subscription:*" 2>/dev/null | head -5)
    
    if [ -n "$subscription_keys" ]; then
        echo -e "${GREEN}✓ 发现订阅记录:${NC}"
        for key in $subscription_keys; do
            echo "  - $key"
            # 显示订阅信息（前100个字符）
            info=$(redis-cli -a "$REDIS_PASSWORD" GET "$key" 2>/dev/null | head -c 100)
            if [ -n "$info" ]; then
                echo "    信息: ${info}..."
            fi
        done
    else
        echo -e "${YELLOW}⚠️  暂无订阅记录（正常，设备尚未连接）${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  redis-cli 未安装，跳过状态检查${NC}"
fi
echo ""

# 总结
echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}测试完成！${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
echo -e "${BLUE}下一步：${NC}"
echo "  1. 启动服务: ./start-radar.sh"
echo "  2. 查看日志，确认订阅管理器已启动"
echo "  3. 等待设备连接，验证自动订阅功能"
echo "  4. 等待 50 分钟，验证自动续订功能"
echo ""
echo -e "${YELLOW}提示：${NC}"
echo "  - 订阅状态存储在 Redis: radar:subscription:{uid}"
echo "  - 订阅管理器每 50 分钟检查一次"
echo "  - 在订阅过期前 10 分钟自动续订"
echo ""
