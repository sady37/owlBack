#!/bin/bash

# 【旧版】单 bash 后台拉六个 go run；与 systemd 分模块 owlback.* 勿同开（会抢端口）。
# 推荐：systemctl start owlback（仓库内 owlback.service 已为 oneshot 编排各 owlback.*）
#       或 ./start-owlback-full.sh ；单独启停：systemctl start/stop owlback.data 等
# 本脚本若仍作手动排障用，启停成对：↔ stop-owlback.sh
#
# 统一启动脚本：启动所有后台服务
# 日志写入 owl/log 各文件；不用 | tee（否则 $! 是 tee，ps -p 与健康检查会误判）
#
# 架构说明：
# - wisefido-data: 数据管理 API + 卡片创建/更新（Redis 缓存 + config.card.*）
# - wisefido-cardagg: 数据聚合（从 PostgreSQL + Redis 聚合卡片数据并缓存到 Redis）
# - wisefido-iot: 数据消费服务（从 Redis Streams 消费数据，存储到 TimescaleDB）
# - wisefido-sensor: AI 智能推理服务（高级推理、访客识别、巡房优化）
#
# Radar 网关由 wisefido-qinglan 提供（本脚本已启动）

# 注意：不使用 set -e，以便在服务失败时能执行清理函数

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 脚本所在目录（先定义，供 check_running_services 里 stop-owlback.sh 使用）
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OWL_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# 默认 owl 根目录下 log/；Alitest 可 export LOG_DIR=/home/wisefido/owl/log
LOG_DIR="${LOG_DIR:-$OWL_ROOT/log}"
mkdir -p "$LOG_DIR"

tcp_listen() {
    command -v lsof >/dev/null 2>&1 && lsof -nP -iTCP:"$1" -sTCP:LISTEN -t >/dev/null 2>&1
}

# 与 stop-owlback.sh 一致：多源解析 LISTEN PID（避免仅 lsof 漏检）
pids_listen_tcp_port_start() {
    local port="$1"
    local p ssout fu
    {
        if command -v lsof >/dev/null 2>&1; then
            p=$(lsof -nP -iTCP:"$port" -sTCP:LISTEN -t 2>/dev/null || true)
            [ -z "$p" ] && p=$(lsof -nP -i :"$port" -sTCP:LISTEN -t 2>/dev/null || true)
            for x in $p; do printf '%s\n' "$x"; done
        fi
        if command -v ss >/dev/null 2>&1; then
            ssout=$(ss -lntp "sport = :$port" 2>/dev/null || true)
            [ -z "$ssout" ] && ssout=$(ss -lntp 2>/dev/null | grep -E ":${port}([^0-9]|$)" || true)
            echo "$ssout" | sed -n 's/.*pid=\([0-9][0-9]*\).*/\1/p'
        fi
        if command -v fuser >/dev/null 2>&1; then
            fu=$(fuser "${port}/tcp" 2>&1 || true)
            echo "$fu" | sed 's/.*:[[:space:]]*//' | tr -s ' ' '\n' | grep -E '^[0-9]+$'
        fi
    } | sort -u -n
}

# 仅杀 TCP LISTEN 占用者（与 stop-owlback 策略一致）
START_SHELL_PID=$$
START_SHELL_PPID=$PPID
safe_kill_start() {
    local pid="$1"
    [ -z "$pid" ] && return
    [ "$pid" = "$START_SHELL_PID" ] || [ "$pid" = "$START_SHELL_PPID" ] && return
    kill -9 "$pid" 2>/dev/null || true
}
kill_tree_start() {
    local pid="$1"
    [ -z "$pid" ] && return
    local c
    for c in $(pgrep -P "$pid" 2>/dev/null || true); do
        kill_tree_start "$c"
    done
    safe_kill_start "$pid"
}
free_listen_port() {
    local port=$1
    local service_name=$2
    if ! command -v lsof >/dev/null 2>&1 && ! command -v ss >/dev/null 2>&1 && ! command -v fuser >/dev/null 2>&1; then
        echo -e "${YELLOW}Note: lsof/ss/fuser unavailable, skip freeing port $port ($service_name)${NC}"
        return 0
    fi
    local pids
    pids=$(pids_listen_tcp_port_start "$port" | tr '\n' ' ')
    if [ -z "$(echo "$pids" | tr -d '[:space:]')" ]; then
        return 0
    fi
    echo -e "${YELLOW}Port $port ($service_name): LISTEN — killing PID tree: $pids${NC}"
    for pid in $pids; do
        [ -n "$pid" ] && kill_tree_start "$pid"
    done
    sleep 1
    if [ -n "$(pids_listen_tcp_port_start "$port" | head -1)" ]; then
        if command -v fuser >/dev/null 2>&1; then
            echo -e "${YELLOW}Port $port still busy, fuser -k ${port}/tcp${NC}"
            fuser -k "${port}/tcp" 2>/dev/null || true
            sleep 1
        fi
    fi
    if [ -n "$(pids_listen_tcp_port_start "$port" | head -1)" ]; then
        echo -e "${RED}Error: port $port still LISTEN after kill${NC}"
        exit 1
    fi
}

# 日志文件
DATA_LOG="$LOG_DIR/wisefido-data.log"
AGGREGATOR_LOG="$LOG_DIR/wisefido-cardagg.log"
QINGLAN_LOG="$LOG_DIR/wisefido-qinglan.log"
SLEEPACE_LOG="$LOG_DIR/wisefido-sleepace.log"
IOT_LOG="$LOG_DIR/wisefido-iot.log"
AI_LOG="$LOG_DIR/wisefido-sensor.log"
COMBINED_LOG="$LOG_DIR/combined.log"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Starting OwlBack Services${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed or not in PATH${NC}"
    exit 1
fi

# 检查是否有服务仍在运行（端口 + 无固定口的进程）
check_running_services() {
    local any_running=false
    local running_names=()

    if tcp_listen 8080; then any_running=true; running_names+=("wisefido-data(:8080)"); fi
    if tcp_listen 8081; then any_running=true; running_names+=("wisefido-qinglan(:8081)"); fi
    if tcp_listen 8083; then any_running=true; running_names+=("wisefido-sleepace(:8083)"); fi
    if tcp_listen 8085; then any_running=true; running_names+=("wisefido-iot(:8085)"); fi
    if pgrep -f "wisefido-cardagg" >/dev/null 2>&1; then any_running=true; running_names+=("wisefido-cardagg"); fi
    if pgrep -f "wisefido-sensor" >/dev/null 2>&1; then any_running=true; running_names+=("wisefido-sensor"); fi

    if [ "$any_running" = true ]; then
        echo -e "${YELLOW}Warning: Services are already running${NC}"
        for n in "${running_names[@]}"; do echo -e "${YELLOW}  - $n${NC}"; done
        echo ""
        echo -e "${BLUE}Options:${NC}"
        echo "  1) Stop existing services and restart"
        echo "  2) Exit (keep existing services running)"
        echo ""
        read -p "Choose an option (1 or 2): " -n 1 -r
        echo ""
        if [[ $REPLY =~ ^[1]$ ]]; then
            echo -e "${YELLOW}Stopping existing services...${NC}"
            bash "$SCRIPT_DIR/stop-owlback.sh"
            echo -e "${GREEN}Existing services stopped${NC}"
            echo ""
        else
            echo -e "${BLUE}Exiting. Existing services will continue running.${NC}"
            exit 0
        fi
    fi
}

# 先清端口与典型 go run 残留，再询问是否仍有占用（避免「已 stop 仍提示已运行」）
free_listen_port 8080 "wisefido-data"
free_listen_port 8081 "wisefido-qinglan"
free_listen_port 8083 "wisefido-sleepace"
free_listen_port 8085 "wisefido-iot"
if command -v pgrep >/dev/null 2>&1; then
    for pid in $(pgrep -f "cmd/wisefido-sensor/main.go" 2>/dev/null || true); do
        [ -n "$pid" ] && kill_tree_start "$pid"
    done
    for pid in $(pgrep -f "wisefido-cardagg/main.go" 2>/dev/null || true); do
        [ -n "$pid" ] && kill_tree_start "$pid"
    done
    for pid in $(pgrep -f "go run.*wisefido-sensor" 2>/dev/null || true); do
        [ -n "$pid" ] && kill_tree_start "$pid"
    done
    for pid in $(pgrep -f "go run.*wisefido-cardagg" 2>/dev/null || true); do
        [ -n "$pid" ] && kill_tree_start "$pid"
    done
fi
sleep 1

check_running_services

# 加载统一配置文件
# 获取脚本所在目录（owlBack 根目录）
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/.env"

if [ -f "$ENV_FILE" ]; then
    echo -e "${BLUE}Loading configuration from: $ENV_FILE${NC}"
    set -a  # 自动导出所有变量
    source "$ENV_FILE"
    set +a  # 关闭自动导出
else
    echo -e "${YELLOW}Warning: .env file not found, using default values${NC}"
    # 设置默认环境变量（如果 .env 不存在）
    export DB_HOST="${DB_HOST:-127.0.0.1}"
    export DB_PORT="${DB_PORT:-5432}"
    export DB_USER="${DB_USER:-postgres}"
    export DB_PASSWORD="${DB_PASSWORD:-postgres}"
    export DB_NAME="${DB_NAME:-owlrd}"
    export DB_SSLMODE="${DB_SSLMODE:-disable}"
    export REDIS_ADDR="${REDIS_ADDR:-127.0.0.1:6379}"
    export REDIS_PASSWORD="${REDIS_PASSWORD:-TeLunSu-36kr}"
    export MQTT_BROKER="${MQTT_BROKER:-127.0.0.1}"
    export MQTT_PORT="${MQTT_PORT:-1883}"
fi

# 设置默认值（如果环境变量未设置）
export DB_HOST="${DB_HOST:-127.0.0.1}"
export DB_PORT="${DB_PORT:-5432}"
export DB_USER="${DB_USER:-postgres}"
export DB_PASSWORD="${DB_PASSWORD:-postgres}"
export DB_NAME="${DB_NAME:-owlrd}"
export DB_SSLMODE="${DB_SSLMODE:-disable}"

# 使用 127.0.0.1 而不是 localhost 以避免 IPv6 解析问题
export REDIS_ADDR="${REDIS_ADDR:-127.0.0.1:6379}"
export REDIS_PASSWORD="${REDIS_PASSWORD:-TeLunSu-36kr}"

# MQTT配置（qinglan 用 MQTT_BROKER+MQTT_PORT 拼接，sleepace 用 MQTT_BROKER 含端口，故全局只设 host，sleepace 启动行单独传 host:port）
export MQTT_BROKER="${MQTT_BROKER:-127.0.0.1}"
export MQTT_PORT="${MQTT_PORT:-1883}"

# wisefido-sensor 可读 TENANT_ID（可选）；不设也能启动，管理员建业务租户后再填即可。不在此写默认 UUID。
export CARD_TRIGGER_MODE="${CARD_TRIGGER_MODE:-polling}"
export CARD_POLLING_INTERVAL="${CARD_POLLING_INTERVAL:-86400}"
export CARD_AGGREGATION_ENABLED="${CARD_AGGREGATION_ENABLED:-true}"
export CARD_AGGREGATION_INTERVAL="${CARD_AGGREGATION_INTERVAL:-2}"

# 日志配置（如果未从 .env 加载）
export LOG_LEVEL="${LOG_LEVEL:-info}"
export LOG_FORMAT="${LOG_FORMAT:-json}"

# 启动前检查 HTTPS 证书有效期：若剩余天数 <= 阈值则自动从系统 Let's Encrypt 复制
refresh_qinglan_tls_cert_if_needed() {
    local threshold_days now expiry_epoch remain_days
    local cert_file copy_script le_domain
    threshold_days="${TLS_CERT_REFRESH_THRESHOLD_DAYS:-5}"
    cert_file="${RADAR_HTTPS_CERT_FILE:-$SCRIPT_DIR/owl-common/server.crt}"
    copy_script="$SCRIPT_DIR/scripts/copy-letsencrypt-to-owl-common.sh"
    le_domain="${LE_DOMAIN:-${RADAR_MQTT_SERVER:-test.wisefido.com}}"

    if ! [[ "$threshold_days" =~ ^[0-9]+$ ]]; then
        threshold_days=5
    fi

    if [ ! -f "$copy_script" ]; then
        echo -e "${YELLOW}Warning: cert refresh script not found: $copy_script${NC}"
        return 0
    fi

    if [ ! -f "$cert_file" ]; then
        echo -e "${YELLOW}TLS cert not found, copying from Let's Encrypt for domain: $le_domain${NC}"
        if LE_DOMAIN="$le_domain" bash "$copy_script"; then
            echo -e "${GREEN}TLS cert copied successfully${NC}"
        else
            echo -e "${YELLOW}Warning: TLS cert copy failed, continue startup${NC}"
        fi
        return 0
    fi

    expiry_epoch="$(openssl x509 -in "$cert_file" -noout -enddate 2>/dev/null | cut -d= -f2 | xargs -I{} date -d "{}" +%s 2>/dev/null || true)"
    now="$(date +%s)"
    if [ -z "$expiry_epoch" ]; then
        echo -e "${YELLOW}Warning: cannot parse cert expiry: $cert_file${NC}"
        return 0
    fi
    remain_days=$(( (expiry_epoch - now) / 86400 ))

    echo "  TLS cert check: $cert_file, remaining ${remain_days} day(s), threshold ${threshold_days} day(s)"
    if [ "$remain_days" -le "$threshold_days" ]; then
        echo -e "${YELLOW}TLS cert expires soon, refreshing from Let's Encrypt (domain: $le_domain)...${NC}"
        if LE_DOMAIN="$le_domain" bash "$copy_script"; then
            echo -e "${GREEN}TLS cert refreshed successfully${NC}"
        else
            echo -e "${YELLOW}Warning: TLS cert refresh failed, continue startup${NC}"
        fi
    fi
}

# 显示配置
echo -e "${BLUE}Configuration:${NC}"
echo "  DB_HOST: $DB_HOST"
echo "  DB_NAME: $DB_NAME"
echo "  REDIS_ADDR: $REDIS_ADDR"
echo "  MQTT_BROKER: $MQTT_BROKER:$MQTT_PORT"
echo "  TENANT_ID: ${TENANT_ID:-<unset — 可选；建业务租户后填 wisefido-sensor 轮询用>}"
echo "  CARD_TRIGGER_MODE: $CARD_TRIGGER_MODE (card creation now handled by wisefido-data, this is backup)"
if [ "$CARD_POLLING_INTERVAL" -ge 86400 ]; then
    echo "  CARD_POLLING_INTERVAL: $CARD_POLLING_INTERVAL seconds (auto-switched to daily 8:00 AM scheduled task)"
else
    echo "  CARD_POLLING_INTERVAL: $CARD_POLLING_INTERVAL seconds (fixed interval polling)"
fi
echo "  CARD_AGGREGATION_ENABLED: $CARD_AGGREGATION_ENABLED (main function: aggregate card data)"
echo "  CARD_AGGREGATION_INTERVAL: $CARD_AGGREGATION_INTERVAL seconds (aggregate every ${CARD_AGGREGATION_INTERVAL}s to match heart/breath rate update frequency, runs on startup)"
echo "  LOG_LEVEL: $LOG_LEVEL"
echo ""
refresh_qinglan_tls_cert_if_needed

echo -e "${BLUE}Service startup behavior:${NC}"
echo "  - wisefido-data: Full card check/update on startup (after 2s delay, in main.go)"
echo "  - wisefido-cardagg:"
echo "    * Full card creation on startup (in startPollingMode)"
echo "    * Data aggregation on startup + every ${CARD_AGGREGATION_INTERVAL}s (in startDataAggregation)"
echo ""
echo -e "${BLUE}Log files:${NC}"
echo "  wisefido-data: $DATA_LOG"
echo "  wisefido-cardagg: $AGGREGATOR_LOG"
echo "  wisefido-qinglan: $QINGLAN_LOG"
echo "  wisefido-sleepace: $SLEEPACE_LOG"
echo "  wisefido-iot: $IOT_LOG"
echo "  wisefido-sensor: $AI_LOG"
echo "  combined: $COMBINED_LOG"
echo ""

# 启动依赖服务（PostgreSQL, Redis, MQTT）
echo -e "${BLUE}Starting dependency services (PostgreSQL, Redis, MQTT)...${NC}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
if [ -f "${SCRIPT_DIR}/docker-compose.yml" ]; then
    cd "$SCRIPT_DIR"
    docker-compose up -d postgresql redis mqtt sleepace-mysql 2>&1 | grep -v "is up-to-date" || true
    echo -e "${GREEN}Dependency services started${NC}"
    echo ""
fi

# 清理旧日志（可选）
if [ "$1" == "--clean" ]; then
    echo -e "${YELLOW}Cleaning old log files...${NC}"
    rm -f "$DATA_LOG" "$AGGREGATOR_LOG" "$QINGLAN_LOG" "$SLEEPACE_LOG" "$IOT_LOG" "$AI_LOG" "$COMBINED_LOG"
fi

# 获取脚本所在目录（owlBack 根目录）
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OWLBACK_DIR="$SCRIPT_DIR"

# 检查服务目录是否存在
if [ ! -d "$OWLBACK_DIR/wisefido-data" ]; then
    echo -e "${RED}Error: wisefido-data directory not found at $OWLBACK_DIR/wisefido-data${NC}"
    exit 1
fi

if [ ! -d "$OWLBACK_DIR/wisefido-cardagg" ]; then
    echo -e "${RED}Error: wisefido-cardagg directory not found at $OWLBACK_DIR/wisefido-cardagg${NC}"
    exit 1
fi

if [ ! -d "$OWLBACK_DIR/wisefido-qinglan" ]; then
    echo -e "${RED}Error: wisefido-qinglan directory not found at $OWLBACK_DIR/wisefido-qinglan${NC}"
    exit 1
fi

# wisefido-sleepace 可选：目录不存在则跳过（如 Alitest 未拉取或不需要 Sleepad 网关）
SLEEPACE_DIR_MISSING=false
if [ ! -d "$OWLBACK_DIR/wisefido-sleepace" ]; then
    echo -e "${YELLOW}Warning: wisefido-sleepace directory not found at $OWLBACK_DIR/wisefido-sleepace, skipping${NC}"
    SLEEPACE_DIR_MISSING=true
fi

EXPECTED_SERVICES=6
[ "$SLEEPACE_DIR_MISSING" = true ] && EXPECTED_SERVICES=5

# 启动 wisefido-data 服务
# 功能：数据管理 API + 卡片创建/更新（直接同步调用 CardCreator）
echo -e "${GREEN}[1/6] Starting wisefido-data service...${NC}"
echo -e "${BLUE}  Function: Data management API + Card creation/update (direct sync)${NC}"
cd "$OWLBACK_DIR/wisefido-data"
# 设置环境变量，包括Qinglan服务地址
export HTTP_ADDR="${HTTP_ADDR:-:8080}"
export QINGLAN_API_BASE_URL="${QINGLAN_API_BASE_URL:-http://localhost:8081}"
: >"$DATA_LOG"
go run cmd/wisefido-data/main.go >>"$DATA_LOG" 2>&1 &
DATA_PID=$!
echo "  PID: $DATA_PID"
echo "  Log: $DATA_LOG  (tail -f \"$DATA_LOG\")"

# 等待一下确保服务启动
sleep 5

# 启动 data 后等待 5 秒，确保卡片初始化/同步先完成
echo -e "${YELLOW}Waiting 5 seconds after starting wisefido-data...${NC}"
sleep 5

# 启动 wisefido-iot 服务
# 功能：从 Redis Streams 消费数据，存储到 TimescaleDB
echo -e "${GREEN}[2/6] Starting wisefido-iot service...${NC}"
echo -e "${BLUE}  Function: Consume data from Redis Streams → TimescaleDB${NC}"
cd "$OWLBACK_DIR/wisefido-iot"
# 设置环境变量（如果未从 .env 加载）
export HTTP_ADDR="${HTTP_ADDR:-:8085}"
export REDIS_DB="${REDIS_DB:-0}"
# Stream 配置 - 统一 IoT streams（所有设备类型）
export STREAM_IOT_MONITOR="${STREAM_IOT_MONITOR:-iot:monitor:stream}"
export STREAM_IOT_STAT="${STREAM_IOT_STAT:-iot:stat:stream}"
export STREAM_IOT_EVENT="${STREAM_IOT_EVENT:-iot:event:stream}"
export STREAM_IOT_ALARM="${STREAM_IOT_ALARM:-iot:alarm:stream}"
export STREAM_IOT_AUTH="${STREAM_IOT_AUTH:-iot:auth:stream}"
# 兼容性配置（旧名称，指向统一的 stream）
export STREAM_RADAR_MONITOR="${STREAM_RADAR_MONITOR:-iot:monitor:stream}"
export STREAM_RADAR_STAT="${STREAM_RADAR_STAT:-iot:stat:stream}"
export STREAM_RADAR_EVENT="${STREAM_RADAR_EVENT:-iot:event:stream}"
export STREAM_RADAR_ALARM="${STREAM_RADAR_ALARM:-iot:alarm:stream}"
export STREAM_RADAR_AUTH="${STREAM_RADAR_AUTH:-iot:auth:stream}"
export STREAM_SLEEPACE_MONITOR="${STREAM_SLEEPACE_MONITOR:-iot:monitor:stream}"
export STREAM_SLEEPACE_STAT="${STREAM_SLEEPACE_STAT:-iot:stat:stream}"
export STREAM_SLEEPACE_EVENT="${STREAM_SLEEPACE_EVENT:-iot:event:stream}"
export STREAM_SLEEPACE_ALARM="${STREAM_SLEEPACE_ALARM:-iot:alarm:stream}"
export STREAM_SLEEPACE_AUTH="${STREAM_SLEEPACE_AUTH:-iot:auth:stream}"
export CONSUMER_GROUP="${CONSUMER_GROUP:-iot-timeseries-group}"
export CONSUMER_NAME="${CONSUMER_NAME:-iot-timeseries-1}"
: >"$IOT_LOG"
mkdir -p "$OWLBACK_DIR/wisefido-iot/.bin"
if go build -o "$OWLBACK_DIR/wisefido-iot/.bin/wisefido-iot" ./cmd/wisefido-iot >/dev/null 2>&1; then
    "$OWLBACK_DIR/wisefido-iot/.bin/wisefido-iot" >>"$IOT_LOG" 2>&1 &
    IOT_PID=$!
else
    echo -e "${YELLOW}Warning: go build failed for wisefido-iot, falling back to go run${NC}"
    go run cmd/wisefido-iot/main.go >>"$IOT_LOG" 2>&1 &
    IOT_PID=$!
fi
echo "  PID: $IOT_PID"
echo "  Log: $IOT_LOG  (tail -f \"$IOT_LOG\")"

# 等待一下确保服务启动
sleep 2

# 启动 wisefido-sensor 服务
# 功能：AI 智能推理服务（高级推理、访客识别、巡房优化）
echo -e "${GREEN}[3/6] Starting wisefido-sensor service...${NC}"
echo -e "${BLUE}  Function: AI inference (advanced reasoning, visitor detection, patrol optimization)${NC}"
cd "$OWLBACK_DIR/wisefido-sensor"
# 设置环境变量（如果未从 .env 加载）
export REDIS_DB="${REDIS_DB:-0}"
: >"$AI_LOG"
mkdir -p "$OWLBACK_DIR/wisefido-sensor/.bin"
if go build -o "$OWLBACK_DIR/wisefido-sensor/.bin/wisefido-sensor" ./cmd/wisefido-sensor >/dev/null 2>&1; then
    "$OWLBACK_DIR/wisefido-sensor/.bin/wisefido-sensor" >>"$AI_LOG" 2>&1 &
    AI_PID=$!
else
    echo -e "${YELLOW}Warning: go build failed for wisefido-sensor, falling back to go run${NC}"
    go run cmd/wisefido-sensor/main.go >>"$AI_LOG" 2>&1 &
    AI_PID=$!
fi
echo "  PID: $AI_PID"
echo "  Log: $AI_LOG  (tail -f \"$AI_LOG\")"

# 等待一下确保服务启动
sleep 2

# 启动 wisefido-cardagg 服务
# 功能：数据聚合（从 PostgreSQL + Redis 聚合卡片数据并缓存）
# 注意：卡片创建/更新现在由 wisefido-data 直接处理，aggregator 主要用于数据聚合
echo -e "${GREEN}[4/6] Starting wisefido-cardagg service...${NC}"
echo -e "${BLUE}  Function: Data aggregation (PostgreSQL + Redis → full card cache)${NC}"
echo -e "${YELLOW}  Note: Card creation/update is now handled by wisefido-data${NC}"
cd "$OWLBACK_DIR/wisefido-cardagg"
: >"$AGGREGATOR_LOG"
mkdir -p "$OWLBACK_DIR/wisefido-cardagg/.bin"
if go build -o "$OWLBACK_DIR/wisefido-cardagg/.bin/wisefido-cardagg" . >/dev/null 2>&1; then
    "$OWLBACK_DIR/wisefido-cardagg/.bin/wisefido-cardagg" >>"$AGGREGATOR_LOG" 2>&1 &
    AGGREGATOR_PID=$!
else
    echo -e "${YELLOW}Warning: go build failed for wisefido-cardagg, falling back to go run${NC}"
    go run "$OWLBACK_DIR/wisefido-cardagg/main.go" >>"$AGGREGATOR_LOG" 2>&1 &
    AGGREGATOR_PID=$!
fi
echo "  PID: $AGGREGATOR_PID"
echo "  Log: $AGGREGATOR_LOG  (tail -f \"$AGGREGATOR_LOG\")"

# 等待一下确保服务启动
sleep 3

# 启动 wisefido-sleepace 服务（目录存在时）
# 功能：Sleepad 设备网关（MQTT 消费 + HTTP 透传代理）
if [ "$SLEEPACE_DIR_MISSING" != "true" ]; then
    echo -e "${GREEN}[5/6] Starting wisefido-sleepace service...${NC}"
    echo -e "${BLUE}  Function: Sleepad device gateway (MQTT consumer + HTTP proxy)${NC}"
    cd "$OWLBACK_DIR/wisefido-sleepace"
    : >"$SLEEPACE_LOG"
    mkdir -p "$OWLBACK_DIR/wisefido-sleepace/.bin"
    if MQTT_CLIENT_ID=wisefido-sleepace-2 go build -o "$OWLBACK_DIR/wisefido-sleepace/.bin/wisefido-sleepace" ./cmd/wisefido-sleepace >/dev/null 2>&1; then
        MQTT_CLIENT_ID=wisefido-sleepace-2 "$OWLBACK_DIR/wisefido-sleepace/.bin/wisefido-sleepace" -env dev >>"$SLEEPACE_LOG" 2>&1 &
        SLEEPACE_PID=$!
    else
        echo -e "${YELLOW}Warning: go build failed for wisefido-sleepace, falling back to go run${NC}"
        MQTT_CLIENT_ID=wisefido-sleepace-2 go run cmd/wisefido-sleepace/main.go -env dev >>"$SLEEPACE_LOG" 2>&1 &
        SLEEPACE_PID=$!
    fi
    echo "  PID: $SLEEPACE_PID"
    echo "  Log: $SLEEPACE_LOG  (tail -f \"$SLEEPACE_LOG\")"
else
    echo -e "${YELLOW}[5/6] Skipping wisefido-sleepace (directory not found)${NC}"
    SLEEPACE_PID=""
fi

# 等待一下确保服务启动
sleep 2

# 启动 wisefido-qinglan 服务
# 功能：设备接入与HTTPS认证 + MQTT消息路由（设备房间消费集成）
echo -e "${GREEN}[6/6] Starting wisefido-qinglan service...${NC}"
echo -e "${BLUE}  Function: Device access + MQTT message routing (internal control)${NC}"
cd "$OWLBACK_DIR/wisefido-qinglan"
# 设置一些关键环境变量（简化配置）
export HTTP_HOST="${HTTP_HOST:-0.0.0.0}"
export HTTP_PORT="${HTTP_PORT:-8081}"
# HTTPS 认证端口：与 .env 中 RADAR_HTTPS_PORT 一致（默认 8443）
export QINGLAN_HTTPS_PORT="${QINGLAN_HTTPS_PORT:-${RADAR_HTTPS_PORT:-8443}}"
# .env 中 RADAR_HTTPS_* 会传给 qinglan 作为 QINGLAN_HTTPS_*（未设则用 qinglan 默认 owl-common 证书）
export QINGLAN_HTTPS_CERT_FILE="${QINGLAN_HTTPS_CERT_FILE:-$RADAR_HTTPS_CERT_FILE}"
export QINGLAN_HTTPS_KEY_FILE="${QINGLAN_HTTPS_KEY_FILE:-$RADAR_HTTPS_KEY_FILE}"
: >"$QINGLAN_LOG"
mkdir -p "$OWLBACK_DIR/wisefido-qinglan/.bin"
if LOG_LEVEL=info go build -o "$OWLBACK_DIR/wisefido-qinglan/.bin/wisefido-qinglan" ./cmd/wisefido-qinglan >/dev/null 2>&1; then
    LOG_LEVEL=info "$OWLBACK_DIR/wisefido-qinglan/.bin/wisefido-qinglan" >>"$QINGLAN_LOG" 2>&1 &
    QINGLAN_PID=$!
else
    echo -e "${YELLOW}Warning: go build failed for wisefido-qinglan, falling back to go run${NC}"
    LOG_LEVEL=info go run cmd/wisefido-qinglan/main.go >>"$QINGLAN_LOG" 2>&1 &
    QINGLAN_PID=$!
fi
echo "  PID: $QINGLAN_PID"
echo "  Log: $QINGLAN_LOG  (tail -f \"$QINGLAN_LOG\")"

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}All backend services are running${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${BLUE}Logs: tail -f \"$DATA_LOG\" …；Ctrl+C 停止所有服务${NC}"
echo ""

# 等待服务启动
sleep 2

# 清理函数：在退出时停止所有后台服务
cleanup() {
    echo ""
    echo -e "${YELLOW}Cleaning up services...${NC}"
    if [ -n "$QINGLAN_PID" ] && ps -p "$QINGLAN_PID" > /dev/null 2>&1; then
        echo "  Stopping wisefido-qinglan (PID: $QINGLAN_PID)"
        kill "$QINGLAN_PID" 2>/dev/null || true
    fi
    if [ -n "$SLEEPACE_PID" ] && ps -p "$SLEEPACE_PID" > /dev/null 2>&1; then
        echo "  Stopping wisefido-sleepace (PID: $SLEEPACE_PID)"
        kill "$SLEEPACE_PID" 2>/dev/null || true
    fi
    if [ -n "$AGGREGATOR_PID" ] && ps -p "$AGGREGATOR_PID" > /dev/null 2>&1; then
        echo "  Stopping wisefido-cardagg (PID: $AGGREGATOR_PID)"
        kill "$AGGREGATOR_PID" 2>/dev/null || true
    fi
    if [ -n "$AI_PID" ] && ps -p "$AI_PID" > /dev/null 2>&1; then
        echo "  Stopping wisefido-sensor (PID: $AI_PID)"
        kill "$AI_PID" 2>/dev/null || true
    fi
    if [ -n "$IOT_PID" ] && ps -p "$IOT_PID" > /dev/null 2>&1; then
        echo "  Stopping wisefido-iot (PID: $IOT_PID)"
        kill "$IOT_PID" 2>/dev/null || true
    fi
    if [ -n "$DATA_PID" ] && ps -p "$DATA_PID" > /dev/null 2>&1; then
        echo "  Stopping wisefido-data (PID: $DATA_PID)"
        kill "$DATA_PID" 2>/dev/null || true
    fi
    # 等待进程退出
    sleep 2
    # 强制杀死仍在运行的进程
    pkill -f "go run.*wisefido-qinglan" 2>/dev/null || true
    pkill -f "go run.*wisefido-sleepace" 2>/dev/null || true
    pkill -f "go run.*wisefido-cardagg" 2>/dev/null || true
    pkill -f "go run.*wisefido-sensor" 2>/dev/null || true
    pkill -f "go run.*wisefido-iot" 2>/dev/null || true
    pkill -f "go run.*wisefido-data" 2>/dev/null || true
    echo -e "${GREEN}Cleanup completed${NC}"
}

# 注册清理函数（在脚本退出时执行）
trap cleanup EXIT INT TERM

# 检查服务是否正常运行
DATA_FAILED=false
AGGREGATOR_FAILED=false
QINGLAN_FAILED=false
SLEEPACE_FAILED=false
IOT_FAILED=false
AI_FAILED=false

if ps -p "$DATA_PID" >/dev/null 2>&1 || tcp_listen 8080; then
    DATA_FAILED=false
else
    echo -e "${RED}Error: wisefido-data service failed to start${NC}"
    echo "Check log: $DATA_LOG"
    tail -20 "$DATA_LOG"
    DATA_FAILED=true
fi

if ps -p "$AGGREGATOR_PID" >/dev/null 2>&1 || pgrep -f "wisefido-cardagg" >/dev/null 2>&1; then
    AGGREGATOR_FAILED=false
else
    echo -e "${RED}Error: wisefido-cardagg service failed to start${NC}"
    echo "Check log: $AGGREGATOR_LOG"
    tail -20 "$AGGREGATOR_LOG"
    AGGREGATOR_FAILED=true
fi

if ps -p "$QINGLAN_PID" >/dev/null 2>&1 || tcp_listen 8081; then
    QINGLAN_FAILED=false
else
    echo -e "${RED}Error: wisefido-qinglan service failed to start${NC}"
    echo "Check log: $QINGLAN_LOG"
    tail -20 "$QINGLAN_LOG"
    QINGLAN_FAILED=true
fi

if [ -n "$SLEEPACE_PID" ]; then
    if ps -p "$SLEEPACE_PID" >/dev/null 2>&1 || tcp_listen 8083; then
        SLEEPACE_FAILED=false
    else
        echo -e "${RED}Error: wisefido-sleepace service failed to start${NC}"
        echo "Check log: $SLEEPACE_LOG"
        tail -20 "$SLEEPACE_LOG"
        SLEEPACE_FAILED=true
    fi
else
    SLEEPACE_FAILED=false
fi

if ps -p "$IOT_PID" >/dev/null 2>&1 || tcp_listen 8085; then
    IOT_FAILED=false
else
    echo -e "${RED}Error: wisefido-iot service failed to start${NC}"
    echo "Check log: $IOT_LOG"
    tail -20 "$IOT_LOG"
    IOT_FAILED=true
fi

if ps -p "$AI_PID" >/dev/null 2>&1 || pgrep -f "wisefido-sensor" >/dev/null 2>&1; then
    AI_FAILED=false
else
    echo -e "${RED}Error: wisefido-sensor service failed to start${NC}"
    echo "Check log: $AI_LOG"
    tail -20 "$AI_LOG"
    AI_FAILED=true
fi

# 统计失败的服务数量
FAILED_COUNT=0
[ "$DATA_FAILED" = true ] && FAILED_COUNT=$((FAILED_COUNT + 1))
[ "$AGGREGATOR_FAILED" = true ] && FAILED_COUNT=$((FAILED_COUNT + 1))
[ "$QINGLAN_FAILED" = true ] && FAILED_COUNT=$((FAILED_COUNT + 1))
[ "$SLEEPACE_FAILED" = true ] && FAILED_COUNT=$((FAILED_COUNT + 1))
[ "$IOT_FAILED" = true ] && FAILED_COUNT=$((FAILED_COUNT + 1))
[ "$AI_FAILED" = true ] && FAILED_COUNT=$((FAILED_COUNT + 1))

# 如果所有服务都失败，退出
if [ "$FAILED_COUNT" -eq "$EXPECTED_SERVICES" ]; then
    echo -e "${RED}All services failed to start. Exiting.${NC}"
    exit 1
fi

# 如果有服务失败，警告但继续
if [ "$DATA_FAILED" = true ]; then
    echo -e "${YELLOW}Warning: wisefido-data failed, but continuing with other services${NC}"
    echo "  You can check the log: $DATA_LOG"
    echo ""
fi

if [ "$AGGREGATOR_FAILED" = true ]; then
    echo -e "${YELLOW}Warning: wisefido-cardagg failed, but continuing with other services${NC}"
    echo "  You can check the log: $AGGREGATOR_LOG"
    echo ""
fi

if [ "$QINGLAN_FAILED" = true ]; then
    echo -e "${YELLOW}Warning: wisefido-qinglan failed, but continuing with other services${NC}"
    echo "  You can check the log: $QINGLAN_LOG"
    echo ""
fi

if [ "$SLEEPACE_FAILED" = true ]; then
    echo -e "${YELLOW}Warning: wisefido-sleepace failed, but continuing with other services${NC}"
    echo "  You can check the log: $SLEEPACE_LOG"
    echo ""
fi

if [ "$IOT_FAILED" = true ]; then
    echo -e "${YELLOW}Warning: wisefido-iot failed, but continuing with other services${NC}"
    echo "  You can check the log: $IOT_LOG"
    echo ""
fi

if [ "$AI_FAILED" = true ]; then
    echo -e "${YELLOW}Warning: wisefido-sensor failed, but continuing with other services${NC}"
    echo "  You can check the log: $AI_LOG"
    echo ""
fi

echo -e "${GREEN}Services started successfully!${NC}"
echo ""
echo -e "${BLUE}日志已写入下列文件（非终端回显），请 tail -f：${NC}"
echo "  - $DATA_LOG"
echo "  - $AGGREGATOR_LOG"
echo "  - $QINGLAN_LOG"
echo "  - $SLEEPACE_LOG"
echo "  - $IOT_LOG"
echo "  - $AI_LOG"
echo ""
echo ""
echo -e "${YELLOW}Press Ctrl+C to stop all services${NC}"
echo ""

# 等待后台 go run 子进程（$! 即为 go）
WAIT_PIDS=""
[ "$DATA_FAILED" = false ] && WAIT_PIDS="$WAIT_PIDS $DATA_PID"
[ "$AGGREGATOR_FAILED" = false ] && WAIT_PIDS="$WAIT_PIDS $AGGREGATOR_PID"
[ "$QINGLAN_FAILED" = false ] && WAIT_PIDS="$WAIT_PIDS $QINGLAN_PID"
[ "$SLEEPACE_FAILED" = false ] && [ -n "$SLEEPACE_PID" ] && WAIT_PIDS="$WAIT_PIDS $SLEEPACE_PID"
[ "$IOT_FAILED" = false ] && WAIT_PIDS="$WAIT_PIDS $IOT_PID"
[ "$AI_FAILED" = false ] && WAIT_PIDS="$WAIT_PIDS $AI_PID"

if [ -n "$WAIT_PIDS" ]; then
    wait $WAIT_PIDS 2>/dev/null || true
fi

