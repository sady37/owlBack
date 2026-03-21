#!/bin/bash

# OwlBack 服务状态检测

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

check_port() {
    local port=$1
    local name=$2
    local pid=""
    if command -v lsof &>/dev/null; then
        pid=$(lsof -nP -iTCP:"$port" -sTCP:LISTEN -t 2>/dev/null | head -1)
    fi
    if [ -n "$pid" ]; then
        echo -e "  ${GREEN}[UP]${NC}   $name  port:$port  pid:$pid  (LISTEN)"
    else
        echo -e "  ${RED}[DOWN]${NC} $name  port:$port  (no LISTEN)"
    fi
}

check_process() {
    local pattern=$1
    local name=$2
    local pid=$(pgrep -f "$pattern" 2>/dev/null | head -1)
    if [ -n "$pid" ]; then
        echo -e "  ${GREEN}[UP]${NC}   $name  pid:$pid"
    else
        echo -e "  ${RED}[DOWN]${NC} $name"
    fi
}

echo ""
echo "=== OwlBack Services ==="
check_port 8080 "wisefido-data"
check_process "wisefido-cardagg" "wisefido-cardagg"
check_port 8081 "wisefido-qinglan"
check_port 8083 "wisefido-sleepace (Go gateway)"
check_port 8085 "wisefido-iot"
check_process "wisefido-ai" "wisefido-ai"

echo ""
echo "=== Infrastructure ==="
check_port 5432 "PostgreSQL"
check_port 6379 "Redis"
check_port 1883 "MQTT (mosquitto)"
check_port 3306 "MySQL"

echo ""
echo "=== Sleepace vendor (Java on host; Go 网关见上 :8083) ==="
check_port 8090 "sleepace-service (Java)"

echo ""
echo "=== Log dir (owl/log) ==="
OWL_LOG="${LOG_DIR:-$(cd "$(dirname "$0")/.." && pwd)/log}"
[ -d "$OWL_LOG" ] && echo "  $OWL_LOG" || echo "  (not created yet)"
echo ""
echo "=== Frontend ==="
check_port 3100 "owlFront (Vite)"
echo ""
