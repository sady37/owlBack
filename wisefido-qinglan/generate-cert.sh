#!/bin/bash

# 生成自签名证书脚本
# 用于 Qinglan HTTPS 认证服务器
# 注意：设备端不验证服务器证书，所以可以使用自签名证书

set -e

# 证书配置
CERT_DIR="${CERT_DIR:-.}"
CERT_FILE="${CERT_FILE:-server.crt}"
KEY_FILE="${KEY_FILE:-server.key}"
CERT_DAYS="${CERT_DAYS:-3650}"  # 10 年有效期

# 服务器信息（可以根据实际部署环境修改）
COMMON_NAME="${CERT_CN:-localhost}"  # 服务器 IP 或域名
ORG_NAME="${CERT_ORG:-WiseFido}"
COUNTRY="${CERT_COUNTRY:-US}"

echo "🔐 Generating self-signed certificate for shared HTTPS Auth Server (owl-common)"
echo "================================================================================"
echo "Certificate directory: $CERT_DIR"
echo "Certificate file: $CERT_FILE"
echo "Key file: $KEY_FILE"
echo "Common Name (CN): $COMMON_NAME"
echo "Organization: $ORG_NAME"
echo "Validity: $CERT_DAYS days"
echo ""

# 检查是否已存在证书
if [ -f "$CERT_DIR/$CERT_FILE" ] || [ -f "$CERT_DIR/$KEY_FILE" ]; then
    echo "⚠️  Warning: Certificate files already exist:"
    [ -f "$CERT_DIR/$CERT_FILE" ] && echo "  - $CERT_DIR/$CERT_FILE"
    [ -f "$CERT_DIR/$KEY_FILE" ] && echo "  - $CERT_DIR/$KEY_FILE"
    echo ""
    read -p "Do you want to overwrite them? (y/N): " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "❌ Certificate generation cancelled"
        exit 1
    fi
    echo "🗑️  Removing old certificate files..."
    rm -f "$CERT_DIR/$CERT_FILE" "$CERT_DIR/$KEY_FILE"
fi

# 生成私钥和证书
echo "📝 Generating private key and certificate..."
openssl req -x509 -newkey rsa:2048 -keyout "$CERT_DIR/$KEY_FILE" \
    -out "$CERT_DIR/$CERT_FILE" -days "$CERT_DAYS" -nodes \
    -subj "/C=$COUNTRY/O=$ORG_NAME/CN=$COMMON_NAME" \
    -addext "subjectAltName=IP:$COMMON_NAME,DNS:$COMMON_NAME,DNS:localhost"

# 设置权限（私钥仅所有者可读）
chmod 600 "$CERT_DIR/$KEY_FILE"
chmod 644 "$CERT_DIR/$CERT_FILE"

echo ""
echo "✅ Certificate generated successfully!"
echo ""
echo "📋 Certificate details:"
openssl x509 -in "$CERT_DIR/$CERT_FILE" -text -noout | grep -E "Subject:|Issuer:|Not Before|Not After|DNS:|IP Address" | head -10

echo ""
echo "📝 Next steps:"
echo "Certificates are generated in owl-common directory and will be automatically used by:"
echo "  - wisefido-radar"
echo "  - wisefido-qinglan"
echo ""
echo "If you need to use different certificates, set environment variables:"
echo "   export RADAR_HTTPS_CERT_FILE=/path/to/server.crt"
echo "   export RADAR_HTTPS_KEY_FILE=/path/to/server.key"
echo "   export QINGLAN_HTTPS_CERT_FILE=/path/to/server.crt"
echo "   export QINGLAN_HTTPS_KEY_FILE=/path/to/server.key"
echo ""
