#!/bin/bash
#
# 启停成对：./start-qinglan.sh ↔ ./stop-qinglan.sh（本脚本会尽量杀光本模块相关进程）
# systemd 分模块成对：systemctl start owlback.qinglan ↔ systemctl stop owlback.qinglan
# 勿混用：若用 start-owlback / systemctl owlback 一体启动，请停 ./stop-owlback.sh 或 systemctl stop owlback
#

cd "$(dirname "$0")"
SCRIPT_DIR="$(pwd)"
OWL_LOG="$(cd "$SCRIPT_DIR/../.." && pwd)/log"
LOG_DIR="${LOG_DIR:-$OWL_LOG}"
LOG_FILE="${QINGLAN_LOG_FILE:-$LOG_DIR/wisefido-qinglan.log}"
SHELL_PID=$$
SHELL_PPID=$PPID

echo "🛑 Stopping wisefido-qinglan service..."

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

pids_listen_tcp_port() {
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

safe_kill() {
    local pid="$1"
    local signal="${2:-9}"  # 默认为SIGKILL
    [ -z "$pid" ] && return
    [ "$pid" = "$SHELL_PID" ] || [ "$pid" = "$SHELL_PPID" ] && return
    kill "-$signal" "$pid" 2>/dev/null || true
}

kill_tree() {
    local pid="$1"
    [ -z "$pid" ] && return
    local c
    for c in $(pgrep -P "$pid" 2>/dev/null || true); do
        kill_tree "$c"
    done
    safe_kill "$pid" 9  # 最终还是使用SIGKILL确保清理
}

# 尝试优雅关闭进程树
graceful_kill_tree() {
    local pid="$1"
    [ -z "$pid" ] && return
    local c
    # 先发送SIGTERM给子进程，让他们有机会优雅关闭
    for c in $(pgrep -P "$pid" 2>/dev/null || true); do
        graceful_kill_tree "$c"
    done
    # 对主进程发送SIGTERM
    safe_kill "$pid" 15  # SIGTERM
}

declare -A seen

add_pids() {
    local pid
    for pid in "$@"; do
        [ -z "$pid" ] && continue
        seen[$pid]=1
    done
}

echo "🔍 Collecting PIDs (ports 8081+8443, log, cmdline patterns)..."

for port in 8081 8443; do
    add_pids $(pids_listen_tcp_port "$port")
done

for pat in "go run.*wisefido-qinglan" "cmd/wisefido-qinglan/main.go" "wisefido-qinglan"; do
    add_pids $(pgrep -f "$pat" 2>/dev/null || true)
done

if [ -f "$LOG_FILE" ] && command -v lsof >/dev/null 2>&1; then
    add_pids $(lsof -t "$LOG_FILE" 2>/dev/null || true)
fi

if [ ${#seen[@]} -eq 0 ]; then
    echo -e "${YELLOW}  No wisefido-qinglan process found${NC}"
else
    echo -e "${BLUE}  Attempting graceful shutdown of ${#seen[@]} process tree(s)...${NC}"
    
    # 首先尝试优雅关闭所有进程
    for pid in "${!seen[@]}"; do
        graceful_kill_tree "$pid"
    done
    
    # 等待最多5秒让进程正常退出
    echo -e "${BLUE}  Waiting for graceful shutdown...${NC}"
    count=0
    while [ $count -lt 5 ]; do
        running_processes=0
        for pid in "${!seen[@]}"; do
            if kill -0 "$pid" 2>/dev/null; then
                running_processes=$((running_processes + 1))
            fi
        done
        
        if [ $running_processes -eq 0 ]; then
            echo -e "${GREEN}  ✅ All processes terminated gracefully${NC}"
            break
        fi
        
        sleep 1
        count=$((count + 1))
    done
    
    # 检查是否仍有进程未关闭，对未关闭的进程进行强制关闭
    for pid in "${!seen[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            echo -e "${YELLOW}  Force killing PID $pid (did not terminate gracefully)${NC}"
            kill_tree "$pid"
        fi
    done
    
    sleep 1
    for port in 8081 8443; do
        if [ -n "$(pids_listen_tcp_port "$port" | head -1)" ]; then
            echo -e "${YELLOW}  Port $port still busy, retry...${NC}"
            for pid in $(pids_listen_tcp_port "$port"); do
                # 先尝试优雅关闭
                safe_kill "$pid" 15
                sleep 1
                # 检查进程是否仍然存活，如果还存在则强制杀死
                if kill -0 "$pid" 2>/dev/null; then
                    kill_tree "$pid"
                fi
            done
        fi
    done
    if command -v fuser >/dev/null 2>&1; then
        for port in 8081 8443; do
            fuser -k "${port}/tcp" 2>/dev/null || true
        done
    fi
fi

echo -e "${GREEN}✅ wisefido-qinglan stop pass done${NC}"

echo ""
echo -e "${BLUE}📊 Status:${NC}"
still=false
for pat in "cmd/wisefido-qinglan/main.go" "go run.*wisefido-qinglan"; do
    if pgrep -f "$pat" >/dev/null 2>&1; then
        still=true
        break
    fi
done
if [ "$still" = true ]; then
    echo -e "  ${RED}❌ Matching processes still present${NC}"
    pgrep -af "qinglan" 2>/dev/null || true
else
    echo -e "  ${GREEN}✅ No qinglan cmdline matches${NC}"
fi

for port in 8081 8443; do
    echo -n "  Port $port: "
    if command -v lsof >/dev/null 2>&1 && lsof -i :$port -sTCP:LISTEN >/dev/null 2>&1; then
        echo -e "${RED}still LISTEN${NC}"
        lsof -i :$port -sTCP:LISTEN 2>/dev/null || true
    else
        echo -e "${GREEN}free${NC}"
    fi
done