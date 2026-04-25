#!/bin/bash

# OwlBack 服务状态检测

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 与 start-owlback / stop-owlback 一致：多源解析 LISTEN（避免仅 lsof 漏检或权限下看不到 PID）
pids_listen_tcp_port_status() {
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

ss_listen_no_pid() {
    local port="$1"
    command -v ss >/dev/null 2>&1 || return 1
    ss -lnt "sport = :$port" 2>/dev/null | grep -q LISTEN && return 0
    ss -lnt 2>/dev/null | grep -E ":${port}([^0-9]|$)" | grep -q LISTEN
}

check_port() {
    local port=$1
    local name=$2
    local pid
    pid=$(pids_listen_tcp_port_status "$port" | head -1)
    if [ -n "$pid" ]; then
        echo -e "  ${GREEN}[UP]${NC}   $name  port:$port  pid:$pid  (LISTEN)"
    elif ss_listen_no_pid "$port"; then
        echo -e "  ${GREEN}[UP]${NC}   $name  port:$port  pid:?  (LISTEN, ss)"
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

# wisefido-cardagg 由 start-owlback 用「go run main.go」或「go run .../wisefido-cardagg/main.go」启动，前者 cmdline 不含 wisefido-cardagg，需多模式匹配
check_cardagg_process() {
    local name="wisefido-cardagg"
    local pid
    pid=$(pgrep -f "wisefido-cardagg/main.go" 2>/dev/null | head -1)
    [ -z "$pid" ] && pid=$(pgrep -f "wisefido-cardagg" 2>/dev/null | head -1)
    [ -z "$pid" ] && pid=$(pgrep -f "go run main.go" 2>/dev/null | head -1)
    if [ -n "$pid" ]; then
        echo -e "  ${GREEN}[UP]${NC}   $name  pid:$pid"
    else
        echo -e "  ${RED}[DOWN]${NC} $name"
    fi
}

echo ""
echo "=== OwlBack Services ==="
check_port 8080 "wisefido-data"
check_cardagg_process
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
