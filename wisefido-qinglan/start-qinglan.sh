#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

# 与 start-owlback 一致：可从 owlBack/.env 继承 RADAR_* / QINGLAN_*（不覆盖脚本内随后显式 export 的 DB 等）
ENV_FILE="$SCRIPT_DIR/../.env"
if [ -f "$ENV_FILE" ]; then
	set -a
	# shellcheck source=/dev/null
	source "$ENV_FILE"
	set +a
fi

echo "🚀 Starting wisefido-qinglan service..."

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

check_port() {
    local port=$1
    local service_name=$2
    
    if ! command -v lsof &> /dev/null; then
        echo -e "${YELLOW}⚠️  lsof not available, skipping port check${NC}"
        return 0
    fi
    
    if lsof -i :$port > /dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  Port $port ($service_name) is already in use${NC}"
        local pids=$(lsof -ti :$port)
        if [ -n "$pids" ]; then
            echo -e "${YELLOW}Killing processes on port $port...${NC}"
            for pid in $pids; do
                kill -9 $pid 2>/dev/null
            done
            sleep 1
        fi
    else
        echo -e "${GREEN}✅ Port $port ($service_name) is available${NC}"
    fi
}

check_and_start_dependencies() {
    echo ""
    echo "🔍 Checking dependencies..."
    
    DOCKER_COMPOSE_FILE="${SCRIPT_DIR}/../docker-compose.yml"
    
    if command -v nc > /dev/null 2>&1; then
        if ! nc -zv 127.0.0.1 5432 > /dev/null 2>&1; then
            echo -e "${RED}❌ PostgreSQL (127.0.0.1:5432) is not accessible${NC}"
            return 1
        fi
        if ! nc -zv 127.0.0.1 6379 > /dev/null 2>&1; then
            echo -e "${RED}❌ Redis (127.0.0.1:6379) is not accessible${NC}"
            return 1
        fi
        
        if ! nc -zv 127.0.0.1 1883 > /dev/null 2>&1; then
            echo -e "${YELLOW}⚠️  MQTT Broker (127.0.0.1:1883) is not accessible${NC}"
            if [ -f "$DOCKER_COMPOSE_FILE" ] && command -v docker-compose > /dev/null 2>&1; then
                echo -e "${BLUE}Starting MQTT Broker...${NC}"
                cd "$(dirname "$DOCKER_COMPOSE_FILE")"
                docker-compose up -d mqtt 2>&1 | grep -v "is up-to-date" || true
                sleep 2
                if nc -zv 127.0.0.1 1883 > /dev/null 2>&1; then
                    echo -e "${GREEN}✅ MQTT Broker started${NC}"
                else
                    echo -e "${RED}❌ Failed to start MQTT Broker${NC}"
                    return 1
                fi
            else
                echo -e "${RED}❌ Cannot start MQTT Broker (docker-compose not found)${NC}"
                return 1
            fi
        else
            echo -e "${GREEN}✅ MQTT Broker (127.0.0.1:1883) is accessible${NC}"
        fi
    fi
    
    return 0
}

check_port 8081 "HTTP Server"

if ! check_and_start_dependencies; then
    echo -e "${RED}❌ Dependencies are not ready${NC}"
    exit 1
fi

echo ""
echo "📋 Setting environment variables..."

export DB_HOST="${DB_HOST:-127.0.0.1}"
export DB_PORT="${DB_PORT:-5432}"
export DB_USER="${DB_USER:-postgres}"
export DB_PASSWORD="${DB_PASSWORD:-postgres}"
export DB_NAME="${DB_NAME:-owlrd}"
export DB_SSLMODE="${DB_SSLMODE:-disable}"

export REDIS_ADDR="${REDIS_ADDR:-127.0.0.1:6379}"
export REDIS_PASSWORD="${REDIS_PASSWORD:-TeLunSu-36kr}"
export REDIS_DB="${REDIS_DB:-0}"

export MQTT_BROKER="${MQTT_BROKER:-127.0.0.1}"
export MQTT_PORT="${MQTT_PORT:-1883}"
export MQTT_CLIENT_ID="${MQTT_CLIENT_ID:-wisefido-qinglan}"
export MQTT_USERNAME="${MQTT_USERNAME:-}"
export MQTT_PASSWORD="${MQTT_PASSWORD:-}"

export RADAR_MQTT_PREFIX=
export RADAR_MQTT_PRODUCT_ID=88

# MQTT 配置：本地 broker 同时支持 MQTT (1883) 和 MQTTS (8883)
# 默认使用 MQTTS (8883, protocol=2) 进行测试
# 如需使用不加密，设置：export RADAR_MQTT_PORT=1883 RADAR_MQTT_PROTOCOL=1
export RADAR_MQTT_PORT=${RADAR_MQTT_PORT:-8883}
export RADAR_MQTT_PROTOCOL=${RADAR_MQTT_PROTOCOL:-2}

# MQTT 配置（返回给设备的配置）
# 注意：必须设置为设备可以访问的 IP 地址，不能使用 127.0.0.1
# 如果未设置，将使用默认值并显示警告
export RADAR_MQTT_SERVER="${RADAR_MQTT_SERVER:-10.0.0.30}"
export RADAR_MQTT_ACCOUNT="${RADAR_MQTT_ACCOUNT:-wfiot}"
export RADAR_MQTT_PASSWORD="${RADAR_MQTT_PASSWORD:-}"

export HTTP_HOST="${HTTP_HOST:-0.0.0.0}"
export HTTP_PORT="${HTTP_PORT:-${QINGLAN_PORT:-8081}}"

# HTTPS 服务器配置（用于设备认证；与 .env 中 RADAR_HTTPS_PORT 一致）
export QINGLAN_HTTPS_PORT="${QINGLAN_HTTPS_PORT:-${RADAR_HTTPS_PORT:-8443}}"

# 证书：优先 .env 中 QINGLAN_HTTPS_* / RADAR_HTTPS_*，否则 owl-common/server.crt|key
# 生产 Let’s Encrypt 覆盖 owl-common：sudo LE_DOMAIN=test.wisefido.com bash ../scripts/copy-letsencrypt-to-owl-common.sh
# 开发自签：cd .. && bash scripts/generate-cert.sh
COMMON_DIR="$(cd "$SCRIPT_DIR/../owl-common" && pwd)"
CERT_DIR="${QINGLAN_CERT_DIR:-$COMMON_DIR}"

# 优先使用 owl-common 的证书
COMMON_CERT_FILE="$COMMON_DIR/server.crt"
COMMON_KEY_FILE="$COMMON_DIR/server.key"

# 启动前检查 HTTPS 证书有效期：若剩余天数 <= 阈值则自动从系统 Let's Encrypt 复制到 owl-common
refresh_common_tls_cert_if_needed() {
	local threshold_days now expiry_epoch remain_days
	local cert_file copy_script le_domain
	threshold_days="${TLS_CERT_REFRESH_THRESHOLD_DAYS:-5}"
	cert_file="$COMMON_CERT_FILE"
	copy_script="$SCRIPT_DIR/../scripts/copy-letsencrypt-to-owl-common.sh"
	le_domain="${LE_DOMAIN:-${RADAR_MQTT_SERVER:-test.wisefido.com}}"

	if ! [[ "$threshold_days" =~ ^[0-9]+$ ]]; then
		threshold_days=5
	fi

	if [ ! -f "$copy_script" ]; then
		echo -e "${YELLOW}⚠️  Cert refresh script not found: $copy_script${NC}"
		return 0
	fi

	if [ ! -f "$cert_file" ]; then
		echo -e "${YELLOW}⚠️  TLS cert missing, refreshing from Let's Encrypt: $le_domain${NC}"
		LE_DOMAIN="$le_domain" bash "$copy_script" || echo -e "${YELLOW}⚠️  TLS cert refresh failed, continue startup${NC}"
		return 0
	fi

	expiry_epoch="$(openssl x509 -in "$cert_file" -noout -enddate 2>/dev/null | cut -d= -f2 | xargs -I{} date -d "{}" +%s 2>/dev/null || true)"
	now="$(date +%s)"
	if [ -z "$expiry_epoch" ]; then
		echo -e "${YELLOW}⚠️  Cannot parse cert expiry: $cert_file${NC}"
		return 0
	fi
	remain_days=$(( (expiry_epoch - now) / 86400 ))
	echo "  🔎 TLS cert check: $cert_file, remaining ${remain_days} day(s), threshold ${threshold_days} day(s)"

	if [ "$remain_days" -le "$threshold_days" ]; then
		echo -e "${YELLOW}⚠️  TLS cert expires soon, refreshing from Let's Encrypt: $le_domain${NC}"
		LE_DOMAIN="$le_domain" bash "$copy_script" || echo -e "${YELLOW}⚠️  TLS cert refresh failed, continue startup${NC}"
	fi
}

refresh_common_tls_cert_if_needed

use_env_pair() {
	local c="$1" k="$2"
	[ -n "$c" ] && [ -n "$k" ] && [ -f "$c" ] && [ -f "$k" ]
}

if use_env_pair "${QINGLAN_HTTPS_CERT_FILE:-}" "${QINGLAN_HTTPS_KEY_FILE:-}"; then
	CERT_FILE="$QINGLAN_HTTPS_CERT_FILE"
	KEY_FILE="$QINGLAN_HTTPS_KEY_FILE"
	echo -e "${GREEN}✅ Using QINGLAN_HTTPS_* certificates:${NC}"
	echo "   Certificate: $CERT_FILE"
	echo "   Key: $KEY_FILE"
	echo ""
elif use_env_pair "${RADAR_HTTPS_CERT_FILE:-}" "${RADAR_HTTPS_KEY_FILE:-}"; then
	CERT_FILE="$RADAR_HTTPS_CERT_FILE"
	KEY_FILE="$RADAR_HTTPS_KEY_FILE"
	echo -e "${GREEN}✅ Using RADAR_HTTPS_* certificates:${NC}"
	echo "   Certificate: $CERT_FILE"
	echo "   Key: $KEY_FILE"
	echo ""
elif [ -f "$COMMON_CERT_FILE" ] && [ -f "$COMMON_KEY_FILE" ]; then
	CERT_FILE="$COMMON_CERT_FILE"
	KEY_FILE="$COMMON_KEY_FILE"
	echo -e "${GREEN}✅ Using owl-common shared certificates:${NC}"
	echo "   Certificate: $CERT_FILE"
	echo "   Key: $KEY_FILE"
	echo ""
else
	CERT_FILE="${QINGLAN_HTTPS_CERT_FILE:-$SCRIPT_DIR/server.crt}"
	KEY_FILE="${QINGLAN_HTTPS_KEY_FILE:-$SCRIPT_DIR/server.key}"
    
    # 转换为绝对路径
    CERT_FILE=$(cd "$(dirname "$CERT_FILE")" 2>/dev/null && pwd)/$(basename "$CERT_FILE") 2>/dev/null || echo "$CERT_FILE"
    KEY_FILE=$(cd "$(dirname "$KEY_FILE")" 2>/dev/null && pwd)/$(basename "$KEY_FILE") 2>/dev/null || echo "$KEY_FILE"
    
    # 检查证书文件是否存在（必须存在，禁止回退到 HTTP）
    if [ ! -f "$CERT_FILE" ] || [ ! -f "$KEY_FILE" ]; then
        echo -e "${RED}❌ Error: HTTPS certificate files not found!${NC}"
        echo "   Certificate: $CERT_FILE"
        echo "   Key: $KEY_FILE"
        echo ""
        echo -e "${YELLOW}💡 To generate self-signed certificates, run:${NC}"
        echo "   cd $SCRIPT_DIR/.. && bash scripts/generate-cert.sh"
        echo ""
        echo -e "${YELLOW}   Or set environment variables:${NC}"
        echo "   export QINGLAN_HTTPS_CERT_FILE=/path/to/server.crt"
        echo "   export QINGLAN_HTTPS_KEY_FILE=/path/to/server.key"
        echo ""
        echo -e "${RED}❌ HTTPS server requires TLS certificates. Service will not start without certificates.${NC}"
        echo ""
        exit 1
    else
        echo -e "${GREEN}✅ Using configured certificates:${NC}"
        echo "   Certificate: $CERT_FILE"
        echo "   Key: $KEY_FILE"
        echo ""
    fi
fi

export QINGLAN_HTTPS_CERT_FILE="$CERT_FILE"
export QINGLAN_HTTPS_KEY_FILE="$KEY_FILE"

export LOG_LEVEL=info
export LOG_FORMAT=json

echo ""
echo -e "${BLUE}📊 Configuration:${NC}"
echo "  🌐 HTTP Server (internal): $HTTP_HOST:$HTTP_PORT"
echo "  🔒 HTTPS Server (auth): 0.0.0.0:$QINGLAN_HTTPS_PORT"
echo "  🗄️  Database: $DB_HOST:$DB_PORT/$DB_NAME"
echo "  📡 Redis: $REDIS_ADDR"
echo "  📨 MQTT Broker: $MQTT_BROKER:$MQTT_PORT"
echo "  📨 Device MQTT (returned to devices): ${RADAR_MQTT_SERVER:-127.0.0.1}:${RADAR_MQTT_PORT:-8883}, protocol=${RADAR_MQTT_PROTOCOL:-2}"
echo ""

echo -e "${GREEN}🚀 Starting wisefido-qinglan service...${NC}"
echo ""

# 日志目录（与 start_owlback.sh 统一）
OWL_LOG="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")/../.." && pwd)/log"
LOG_DIR="${LOG_DIR:-$OWL_LOG}"
mkdir -p "$LOG_DIR"
LOG_FILE="${QINGLAN_LOG_FILE:-$LOG_DIR/wisefido-qinglan.log}"

echo -e "${BLUE}📝 Log Configuration:${NC}"
echo "  📄 Log File: $LOG_FILE"
echo "  📁 Log Directory: $LOG_DIR"
echo ""

# 与 start-owlback.sh 一致：每次启动截断当前日志，再追加本次运行内容
: >"$LOG_FILE"

# 写入启动信息到日志文件
echo "==========================================" >> "$LOG_FILE"
echo "wisefido-qinglan service starting at $(date)" >> "$LOG_FILE"
echo "Log file: $LOG_FILE" >> "$LOG_FILE"
echo "==========================================" >> "$LOG_FILE"

# 同时输出到控制台和日志文件
if [ -t 1 ]; then
    echo -e "${GREEN}✅ Logging to: $LOG_FILE${NC}"
    echo -e "${GREEN}✅ Output will be displayed in terminal and saved to log file${NC}"
    echo ""
    go run cmd/wisefido-qinglan/main.go 2>&1 | tee -a "$LOG_FILE"
else
    echo "Logging to: $LOG_FILE" >&2
    go run cmd/wisefido-qinglan/main.go 2>&1 | tee -a "$LOG_FILE"
fi
