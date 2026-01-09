#!/bin/bash

# Radar HTTPS 认证测试脚本（多设备）
# 用于测试多个 Radar 设备的认证接口

# 服务器地址
SERVER="${RADAR_AUTH_SERVER:-10.0.0.30:8443}"
PROTOCOL="${RADAR_AUTH_PROTOCOL:-http}"  # 开发环境使用 http，生产环境使用 https

# 认证路径（协议文档标准路径）
AUTH_PATH="${RADAR_AUTH_PATH:-/prod-api/thirdmqtt/v2/auth/device}"

# 测试设备列表
# 格式：设备型号|设备代码/SN|UID
declare -a DEVICES=(
    "HC2|25A859B8333B|25A859B8333B"
    "TSL60G442|E598A2ACD523|E598A2ACD523"
)

echo "🧪 Testing Radar HTTPS Authentication (Multiple Devices)"
echo "========================================================="
echo "Server: ${PROTOCOL}://${SERVER}${AUTH_PATH}"
echo "Total devices: ${#DEVICES[@]}"
echo ""

# 测试结果统计
SUCCESS_COUNT=0
FAIL_COUNT=0

# 遍历每个设备进行测试
for device_info in "${DEVICES[@]}"; do
    IFS='|' read -r device_model device_code device_uid <<< "$device_info"
    
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "📱 Device: $device_model"
    echo "   Code/SN: $device_code"
    echo "   UID: $device_uid"
    echo ""
    
    # 构建请求 JSON
    REQUEST_JSON=$(cat <<EOF
{
  "uid": "${device_uid}",
  "type": 1,
  "auth": "test_auth_12345",
  "salt": "12345",
  "mcu": {
    "hw": "2.0",
    "sw": "Oct 23 2024 09:20:18",
    "mac": "00:11:22:33:44:55",
    "iccid": " "
  },
  "radar": {
    "hw": "2.3",
    "sw": "Dec 4 2024 08:53:22",
    "cap": "15"
  }
}
EOF
)
    
    # 发送 POST 请求
    RESPONSE=$(curl -X POST \
      "${PROTOCOL}://${SERVER}${AUTH_PATH}" \
      -H "Content-Type: application/json" \
      -d "$REQUEST_JSON" \
      -w "\nHTTP_STATUS:%{http_code}" \
      -s 2>/dev/null)
    
    # 提取 HTTP 状态码
    HTTP_STATUS=$(echo "$RESPONSE" | grep -o "HTTP_STATUS:[0-9]*" | cut -d: -f2)
    JSON_RESPONSE=$(echo "$RESPONSE" | sed 's/HTTP_STATUS:[0-9]*$//')
    
    # 解析响应
    if [ "$HTTP_STATUS" = "200" ]; then
        CODE=$(echo "$JSON_RESPONSE" | grep -o '"code":[0-9]*' | cut -d: -f2)
        MSG=$(echo "$JSON_RESPONSE" | grep -o '"msg":"[^"]*"' | cut -d: -f2 | tr -d '"')
        
        if [ "$CODE" = "200" ]; then
            echo "✅ Success: $MSG"
            echo "   HTTP Status: $HTTP_STATUS"
            
            # 提取 MQTT 配置信息
            MQTT_SERVER=$(echo "$JSON_RESPONSE" | grep -o '"server":"[^"]*"' | cut -d: -f2 | tr -d '"')
            MQTT_PORT=$(echo "$JSON_RESPONSE" | grep -o '"port":[0-9]*' | cut -d: -f2)
            MQTT_CLIENT_ID=$(echo "$JSON_RESPONSE" | grep -o '"clientId":"[^"]*"' | cut -d: -f2 | tr -d '"')
            
            echo "   MQTT Config:"
            echo "     - Server: $MQTT_SERVER:$MQTT_PORT"
            echo "     - Client ID: $MQTT_CLIENT_ID"
            
            SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
        else
            echo "❌ Failed: $MSG (Code: $CODE)"
            echo "   HTTP Status: $HTTP_STATUS"
            FAIL_COUNT=$((FAIL_COUNT + 1))
        fi
    else
        echo "❌ Failed: HTTP Status $HTTP_STATUS"
        echo "   Response: $JSON_RESPONSE"
        FAIL_COUNT=$((FAIL_COUNT + 1))
    fi
    
    echo ""
    sleep 1  # 避免请求过快
done

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 Test Summary:"
echo "   Total: ${#DEVICES[@]}"
echo "   ✅ Success: $SUCCESS_COUNT"
echo "   ❌ Failed: $FAIL_COUNT"
echo ""
echo "✅ Test completed"

