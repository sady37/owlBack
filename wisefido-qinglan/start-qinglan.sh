#!/bin/bash

cd "$(dirname "$0")"

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
    
    SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
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

export DB_HOST=127.0.0.1
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=owlrd
export DB_SSLMODE=disable

export REDIS_ADDR=127.0.0.1:6379
export REDIS_PASSWORD=TeLunSu-36kr
export REDIS_DB=0

export MQTT_BROKER=127.0.0.1
export MQTT_PORT=1883
export MQTT_CLIENT_ID=wisefido-qinglan
export MQTT_USERNAME=
export MQTT_PASSWORD=

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

export HTTP_HOST=0.0.0.0
export HTTP_PORT=8081

# HTTPS 服务器配置（用于设备认证）
export QINGLAN_HTTPS_PORT="${QINGLAN_HTTPS_PORT:-8443}"

# 证书统一使用 owl-common 下的 server.crt / server.key：
#   - 本脚本：Qinglan HTTPS 设备认证使用 owl-common 证书
#   - MQTT broker（mqtt/config）：建议与 owl-common 一致，可将 mosquitto 配置为使用 owl-common 证书，或把 owl-common 生成的证书复制到 mqtt/config
# 生成证书：在 owl-common 目录执行 ../wisefido-qinglan/generate-cert.sh
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
COMMON_DIR="$(cd "$SCRIPT_DIR/../owl-common" && pwd)"
CERT_DIR="${QINGLAN_CERT_DIR:-$COMMON_DIR}"

# 优先使用 owl-common 的证书
COMMON_CERT_FILE="$COMMON_DIR/server.crt"
COMMON_KEY_FILE="$COMMON_DIR/server.key"

# 如果 owl-common 的证书存在，使用它；否则使用环境变量或本地证书
if [ -f "$COMMON_CERT_FILE" ] && [ -f "$COMMON_KEY_FILE" ]; then
    CERT_FILE="$COMMON_CERT_FILE"
    KEY_FILE="$COMMON_KEY_FILE"
    echo -e "${GREEN}✅ Using owl-common shared certificates:${NC}"
    echo "   Certificate: $CERT_FILE"
    echo "   Key: $KEY_FILE"
    echo ""
else
    # 回退到环境变量或本地证书
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
        echo "   cd $COMMON_DIR && ./generate-cert.sh"
        echo "   (or)"
        echo "   cd $SCRIPT_DIR && ./generate-cert.sh"
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
LOG_DIR="${LOG_DIR:-/tmp/owlBack_logs}"
mkdir -p "$LOG_DIR"
LOG_FILE="${QINGLAN_LOG_FILE:-$LOG_DIR/wisefido-qinglan.log}"

echo -e "${BLUE}📝 Log Configuration:${NC}"
echo "  📄 Log File: $LOG_FILE"
echo "  📁 Log Directory: $LOG_DIR"
echo ""

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
