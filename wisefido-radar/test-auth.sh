#!/bin/bash

# Radar HTTPS 认证测试脚本
# 用于测试 /prod-api/thirdmqtt/v2/auth/device 接口（协议文档标准路径）

# 服务器地址
SERVER="${RADAR_AUTH_SERVER:-10.0.0.30:8443}"
PROTOCOL="${RADAR_AUTH_PROTOCOL:-http}"  # 开发环境使用 http，生产环境使用 https

# 认证路径（协议文档标准路径）
AUTH_PATH="${RADAR_AUTH_PATH:-/prod-api/thirdmqtt/v2/auth/device}"

# 测试设备信息（需要根据实际 device_store 表中的设备修改）
# 注意：使用 DEVICE_UID 而不是 UID，因为 UID 是 shell 内置只读变量
# 
# 可用测试设备：
# - HC2: 25A859B8333B
# - TSL60G442: E598A2ACD523
# - Radar-1797: 4D8710F41797
DEVICE_UID="${RADAR_TEST_UID:-25A859B8333B}"  # 默认使用 HC2 设备
DEVICE_TYPE="${RADAR_TEST_TYPE:-1}"

echo "🧪 Testing Radar HTTPS Authentication"
echo "======================================"
echo "Server: ${PROTOCOL}://${SERVER}${AUTH_PATH}"
echo "Device UID: ${DEVICE_UID}"
echo ""

# 构建请求 JSON
# 注意：使用小写 "mcu" 和 "radar"（与 Go 结构体 JSON 标签一致）
REQUEST_JSON=$(cat <<EOF
{
  "uid": "${DEVICE_UID}",
  "type": ${DEVICE_TYPE},
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

echo "📤 Request:"
echo "$REQUEST_JSON" | jq '.' 2>/dev/null || echo "$REQUEST_JSON"
echo ""

# 发送 POST 请求
echo "📥 Response:"
curl -X POST \
  "${PROTOCOL}://${SERVER}${AUTH_PATH}" \
  -H "Content-Type: application/json" \
  -d "$REQUEST_JSON" \
  -w "\n\nHTTP Status: %{http_code}\n" \
  -s | jq '.' 2>/dev/null || curl -X POST \
  "${PROTOCOL}://${SERVER}${AUTH_PATH}" \
  -H "Content-Type: application/json" \
  -d "$REQUEST_JSON" \
  -w "\n\nHTTP Status: %{http_code}\n" \
  -s

echo ""
echo "✅ Test completed"

