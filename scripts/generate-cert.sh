#!/bin/bash

# 生成自签名证书脚本
# 用于 Qinglan HTTPS 认证服务器
# 注意：设备端不验证服务器证书，所以可以使用自签名证书

set -e

# 解析脚本所在目录（支持 bash scripts/generate-cert.sh 或 ./scripts/generate-cert.sh）
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$0})" && pwd)"
OWL_COMMON_DIR="$SCRIPT_DIR/../owl-common"
[ -d "$OWL_COMMON_DIR" ] || mkdir -p "$OWL_COMMON_DIR"
OWL_COMMON_DIR="$(cd "$OWL_COMMON_DIR" && pwd)"

# 证书配置：共用证书统一输出到 owl-common/server.crt、server.key
CERT_DIR="${CERT_DIR:-$OWL_COMMON_DIR}"
CERT_FILE="${CERT_FILE:-server.crt}"
KEY_FILE="${KEY_FILE:-server.key}"
CERT_DAYS="${CERT_DAYS:-3650}"  # 10 年有效期

# 服务器信息（可以根据实际部署环境修改）
# 当前默认使用 demo.wisefido.com；可覆盖 CERT_CN=test.wisefido.com 等
COMMON_NAME="${CERT_CN:-demo.wisefido.com}"  # 服务器 IP 或域名
ORG_NAME="${CERT_ORG:-WiseFido}"
COUNTRY="${CERT_COUNTRY:-US}"
# 可选：额外 SAN（逗号分隔 IP 或 DNS），如 CERT_SAN_IP=172.16.75.83 CERT_SAN_DNS=alitest.example.com
CERT_SAN_IP="${CERT_SAN_IP:-}"
CERT_SAN_DNS="${CERT_SAN_DNS:-}"

echo "🔐 Generating self-signed certificate for shared HTTPS Auth Server (owl-common)"
echo "================================================================================"
echo "Certificate directory: $CERT_DIR"
echo "Certificate file: $CERT_FILE"
echo "Key file: $KEY_FILE"
echo "Common Name (CN): $COMMON_NAME"
echo "Organization: $ORG_NAME"
echo "Validity: $CERT_DAYS days"
echo ""

# 检查是否已存在证书（CERT_FORCE=1 或非交互时直接覆盖）
if [ -f "$CERT_DIR/$CERT_FILE" ] || [ -f "$CERT_DIR/$KEY_FILE" ]; then
    echo "⚠️  Warning: Certificate files already exist:"
    [ -f "$CERT_DIR/$CERT_FILE" ] && echo "  - $CERT_DIR/$CERT_FILE"
    [ -f "$CERT_DIR/$KEY_FILE" ] && echo "  - $CERT_DIR/$KEY_FILE"
    echo ""
    if [[ "${CERT_FORCE:-0}" != "1" ]] && [ -t 0 ]; then
        read -p "Do you want to overwrite them? (y/N): " -n 1 -r
        echo ""
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "❌ Certificate generation cancelled"
            exit 1
        fi
    fi
    echo "🗑️  Removing old certificate files..."
    rm -f "$CERT_DIR/$CERT_FILE" "$CERT_DIR/$KEY_FILE"
fi

# 生成私钥和证书（SAN：CN 为 IP 则加 IP，否则加 DNS；再加 localhost 及可选的 CERT_SAN_*）
build_san() {
  local san="DNS:localhost"
  if echo "$COMMON_NAME" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$'; then
    san="IP:$COMMON_NAME,$san"
  else
    san="DNS:$COMMON_NAME,$san"
  fi
  [ -n "$CERT_SAN_IP" ] && san="$san,IP:${CERT_SAN_IP// /,IP:}"
  [ -n "$CERT_SAN_DNS" ] && san="$san,DNS:${CERT_SAN_DNS// /,DNS:}"
  echo "$san"
}
SAN=$(build_san)
echo "📝 Generating private key and certificate (SAN: $SAN)..."
openssl req -x509 -newkey rsa:2048 -keyout "$CERT_DIR/$KEY_FILE" \
    -out "$CERT_DIR/$CERT_FILE" -days "$CERT_DAYS" -nodes \
    -subj "/C=$COUNTRY/O=$ORG_NAME/CN=$COMMON_NAME" \
    -addext "subjectAltName=$SAN"

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
echo "Certificates are in owl-common ($OWL_COMMON_DIR), shared by:"
echo "  - wisefido-radar"
echo "  - wisefido-qinglan"
echo ""
echo "Override output dir: CERT_DIR=/path/to/dir $0"
echo "Override cert names: RADAR_HTTPS_CERT_FILE=... QINGLAN_HTTPS_CERT_FILE=... (in .env)"
echo ""
echo "Alitest / demo (CN/SAN match host):"
echo "   cd owlBack && CERT_CN=test.wisefido.com CERT_FORCE=1 bash scripts/generate-cert.sh"
echo "   # or: CERT_CN=demo.wisefido.com CERT_FORCE=1 ..."
echo "   # or both: CERT_CN=test.wisefido.com CERT_SAN_DNS='demo.wisefido.com' CERT_FORCE=1 ..."
echo "   (then restart qinglan)"
echo ""
