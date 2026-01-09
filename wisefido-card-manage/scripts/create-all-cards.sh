#!/bin/bash

# wisefido-card-manage 全量卡片创建脚本
# 
# 用途：通过 HTTP API 触发全量卡片创建/更新
# 使用场景：crontab 定时任务（每天早上 8:00）

set -e

# 配置
CARD_MANAGE_URL="${CARD_MANAGE_URL:-http://localhost:8082}"
TENANT_ID="${TENANT_ID:-}"

# 检查参数
if [ -z "$TENANT_ID" ]; then
    echo "Error: TENANT_ID environment variable is required"
    exit 1
fi

# 发送 HTTP 请求
echo "Creating all cards for tenant: $TENANT_ID"
response=$(curl -s -w "\n%{http_code}" -X POST \
    "${CARD_MANAGE_URL}/api/v1/cards/create-all" \
    -H "Content-Type: application/json" \
    -d "{\"tenant_id\":\"${TENANT_ID}\"}")

# 解析响应
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

# 检查 HTTP 状态码
if [ "$http_code" -eq 200 ]; then
    echo "Success: All cards created/updated"
    echo "$body"
    exit 0
else
    echo "Error: HTTP $http_code"
    echo "$body"
    exit 1
fi

