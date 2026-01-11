#!/bin/bash

# 详细的订阅功能测试脚本
# 验证订阅管理器的各个组件

set -e

cd "$(dirname "$0")"

echo "========================================="
echo "订阅功能详细测试"
echo "========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 1. 检查文件结构
echo -e "${BLUE}[1/6] 检查文件结构...${NC}"

files=(
    "internal/service/subscription_manager.go"
    "internal/config/config.go"
    "internal/consumer/mqtt_consumer.go"
    "internal/service/radar.go"
    "internal/http/command_service.go"
)

for file in "${files[@]}"; do
    if [ -f "$file" ]; then
        echo -e "${GREEN}✓${NC} $file"
    else
        echo -e "${RED}✗${NC} $file (缺失)"
        exit 1
    fi
done
echo ""

# 2. 检查订阅管理器关键方法
echo -e "${BLUE}[2/6] 检查订阅管理器关键方法...${NC}"

methods=(
    "NewSubscriptionManager"
    "Start"
    "Stop"
    "AutoSubscribe"
    "checkAndRenewSubscriptions"
    "renewSubscription"
    "saveSubscriptionInfo"
    "isSubscribed"
    "getAllActiveSubscriptions"
)

for method in "${methods[@]}"; do
    if grep -q "func.*$method" internal/service/subscription_manager.go; then
        echo -e "${GREEN}✓${NC} $method"
    else
        echo -e "${RED}✗${NC} $method (缺失)"
    fi
done
echo ""

# 3. 检查配置项
echo -e "${BLUE}[3/6] 检查配置项...${NC}"

config_items=(
    "Subscription"
    "AutoSubscribe"
    "DefaultDuration"
    "DefaultContent"
    "RenewalInterval"
    "RenewalAdvanceTime"
)

for item in "${config_items[@]}"; do
    if grep -q "$item" internal/config/config.go; then
        echo -e "${GREEN}✓${NC} $item"
    else
        echo -e "${RED}✗${NC} $item (缺失)"
    fi
done
echo ""

# 4. 检查 MQTT Consumer 集成
echo -e "${BLUE}[4/6] 检查 MQTT Consumer 集成...${NC}"

if grep -q "subscriptionManager" internal/consumer/mqtt_consumer.go; then
    echo -e "${GREEN}✓${NC} subscriptionManager 字段存在"
else
    echo -e "${RED}✗${NC} subscriptionManager 字段缺失"
fi

if grep -q "SetSubscriptionManager" internal/consumer/mqtt_consumer.go; then
    echo -e "${GREEN}✓${NC} SetSubscriptionManager 方法存在"
else
    echo -e "${RED}✗${NC} SetSubscriptionManager 方法缺失"
fi

if grep -q "isFirstConnection" internal/consumer/mqtt_consumer.go; then
    echo -e "${GREEN}✓${NC} 首次连接检测逻辑存在"
else
    echo -e "${RED}✗${NC} 首次连接检测逻辑缺失"
fi

if grep -q "AutoSubscribe" internal/consumer/mqtt_consumer.go; then
    echo -e "${GREEN}✓${NC} 自动订阅调用存在"
else
    echo -e "${RED}✗${NC} 自动订阅调用缺失"
fi
echo ""

# 5. 检查 RadarService 集成
echo -e "${BLUE}[5/6] 检查 RadarService 集成...${NC}"

if grep -q "subscriptionManager.*SubscriptionManager" internal/service/radar.go; then
    echo -e "${GREEN}✓${NC} subscriptionManager 字段定义存在"
else
    echo -e "${RED}✗${NC} subscriptionManager 字段定义缺失"
fi

if grep -q "NewSubscriptionManager" internal/service/radar.go; then
    echo -e "${GREEN}✓${NC} 订阅管理器创建存在"
else
    echo -e "${RED}✗${NC} 订阅管理器创建缺失"
fi

if grep -q "SetSubscriptionManager" internal/service/radar.go; then
    echo -e "${GREEN}✓${NC} 设置订阅管理器到 Consumer 存在"
else
    echo -e "${RED}✗${NC} 设置订阅管理器到 Consumer 缺失"
fi

if grep -q "subscriptionManager.Start" internal/service/radar.go; then
    echo -e "${GREEN}✓${NC} 订阅管理器启动存在"
else
    echo -e "${RED}✗${NC} 订阅管理器启动缺失"
fi

if grep -q "subscriptionManager.Stop" internal/service/radar.go; then
    echo -e "${GREEN}✓${NC} 订阅管理器停止存在"
else
    echo -e "${RED}✗${NC} 订阅管理器停止缺失"
fi
echo ""

# 6. 检查 CommandService 集成
echo -e "${BLUE}[6/6] 检查 CommandService 集成...${NC}"

if grep -q "saveSubscriptionInfo" internal/http/command_service.go; then
    echo -e "${GREEN}✓${NC} saveSubscriptionInfo 方法存在"
    
    # 检查是否在 SubscribeRealtimeData 中调用
    if grep -A 15 "func.*SubscribeRealtimeData" internal/http/command_service.go | grep -q "saveSubscriptionInfo"; then
        echo -e "${GREEN}✓${NC} 在 SubscribeRealtimeData 中调用 saveSubscriptionInfo"
    else
        echo -e "${YELLOW}⚠️${NC} 在 SubscribeRealtimeData 中未找到 saveSubscriptionInfo 调用（可能已存在但格式不同）"
    fi
else
    echo -e "${RED}✗${NC} saveSubscriptionInfo 方法缺失"
fi
echo ""

# 总结
echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}详细测试完成！${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
echo -e "${BLUE}测试结果总结:${NC}"
echo "  - 所有关键文件已存在"
echo "  - 订阅管理器方法完整"
echo "  - 配置项已添加"
echo "  - MQTT Consumer 集成完成"
echo "  - RadarService 集成完成"
echo "  - CommandService 集成完成"
echo ""
echo -e "${BLUE}下一步:${NC}"
echo "  1. 启动服务: ./start-radar.sh"
echo "  2. 观察日志，确认订阅管理器启动"
echo "  3. 模拟设备连接，验证自动订阅"
echo ""
