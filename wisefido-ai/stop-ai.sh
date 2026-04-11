#!/bin/bash
# 成对: owlback-run-service.sh wisefido-ai ↔ 本脚本；systemctl stop owlback.ai 的 ExecStop
cd "$(dirname "$0")"
SCRIPT_DIR="$(pwd)"
OWL_LOG="$(cd "$SCRIPT_DIR/../.." && pwd)/log"
LOG_DIR="${LOG_DIR:-$OWL_LOG}"
LOG_FILE="${AI_LOG_FILE:-$LOG_DIR/wisefido-ai.log}"
SHELL_PID=$$
SHELL_PPID=$PPID

echo "Stopping wisefido-ai..."

safe_kill() {
    local pid="$1"
    [ -z "$pid" ] && return
    [ "$pid" = "$SHELL_PID" ] || [ "$pid" = "$SHELL_PPID" ] && return
    kill -9 "$pid" 2>/dev/null || true
}

kill_tree() {
    local pid="$1"
    [ -z "$pid" ] && return
    local c
    for c in $(pgrep -P "$pid" 2>/dev/null || true); do
        kill_tree "$c"
    done
    safe_kill "$pid"
}

declare -A seen
add_pids() {
    local pid
    for pid in "$@"; do
        [ -z "$pid" ] && continue
        seen[$pid]=1
    done
}

for pat in "go run.*wisefido-ai" "cmd/wisefido-ai/main.go" "wisefido-ai"; do
    add_pids $(pgrep -f "$pat" 2>/dev/null || true)
done

if [ -f "$LOG_FILE" ] && command -v lsof >/dev/null 2>&1; then
    add_pids $(lsof -t "$LOG_FILE" 2>/dev/null || true)
fi

if [ ${#seen[@]} -eq 0 ]; then
    echo "  No wisefido-ai process found"
else
    for pid in "${!seen[@]}"; do
        kill_tree "$pid"
    done
    sleep 1
fi

echo "wisefido-ai stop done"
