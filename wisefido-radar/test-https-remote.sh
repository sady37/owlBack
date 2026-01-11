#!/bin/bash

# 远端测试 HTTPS 服务的脚本
# 用于检测 wisefido-radar HTTPS 认证服务是否开启

SERVER_IP="${1:-47.77.194.143}"
HTTPS_PORT="${2:-8443}"

echo "=========================================="
echo "测试 HTTPS 服务连接"
echo "=========================================="
echo "服务器: $SERVER_IP"
echo "端口: $HTTPS_PORT"
echo ""

# 1. 测试端口连通性
echo "1. 测试端口连通性..."
if command -v nc > /dev/null 2>&1; then
    if nc -zv -w 5 "$SERVER_IP" "$HTTPS_PORT" 2>&1; then
        echo "   ✅ 端口 $HTTPS_PORT 可访问"
    else
        echo "   ❌ 端口 $HTTPS_PORT 不可访问"
    fi
else
    echo "   ⚠️  nc (netcat) 未安装，跳过端口测试"
fi

echo ""

# 2. 测试 HTTPS 连接
echo "2. 测试 HTTPS 连接..."
if command -v curl > /dev/null 2>&1; then
    RESPONSE=$(curl -k -s -w "\nHTTP_CODE:%{http_code}\nTIME:%{time_total}" \
        -X POST \
        -H "Content-Type: application/json" \
        -d '{"uid":"test","type":1,"mcu":{"hw":"HC2-2.0","sw":"20240101"},"radar":{"hw":"1.0","sw":"20240101"}}' \
        --max-time 10 \
        "https://$SERVER_IP:$HTTPS_PORT/prod-api/thirdmqtt/v2/auth/device" 2>&1)
    
    HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP_CODE" | cut -d: -f2)
    TIME=$(echo "$RESPONSE" | grep "TIME" | cut -d: -f2)
    
    if [ -n "$HTTP_CODE" ] && [ "$HTTP_CODE" != "000" ]; then
        echo "   ✅ HTTPS 服务响应正常"
        echo "   HTTP 状态码: $HTTP_CODE"
        echo "   响应时间: ${TIME}s"
        echo "   响应内容:"
        echo "$RESPONSE" | grep -v "HTTP_CODE\|TIME" | head -5
    else
        echo "   ❌ HTTPS 服务无响应或连接失败"
        echo "   错误信息:"
        echo "$RESPONSE" | head -5
    fi
else
    echo "   ⚠️  curl 未安装，跳过 HTTPS 测试"
fi

echo ""

# 3. 测试证书信息
echo "3. 测试证书信息..."
if command -v openssl > /dev/null 2>&1; then
    CERT_INFO=$(echo | openssl s_client -connect "$SERVER_IP:$HTTPS_PORT" -servername "$SERVER_IP" 2>&1 | grep -E "subject=|issuer=|Verify return code" | head -3)
    if [ -n "$CERT_INFO" ]; then
        echo "   ✅ 证书信息:"
        echo "$CERT_INFO" | sed 's/^/   /'
    else
        echo "   ❌ 无法获取证书信息"
    fi
else
    echo "   ⚠️  openssl 未安装，跳过证书测试"
fi

echo ""
echo "=========================================="
echo "测试完成"
echo "=========================================="
