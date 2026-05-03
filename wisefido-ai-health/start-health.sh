#!/bin/bash
#
# 启停成对：./start-health.sh ↔ ./stop-health.sh
# 用 --once daily / --once monthly 单次跑：./start-health.sh --once daily
#

cd "$(dirname "$0")"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🚀 Starting wisefido-ai-health service...${NC}"

# 简单依赖检查：PG 5432
if command -v nc > /dev/null 2>&1; then
    if ! nc -zv 127.0.0.1 5432 > /dev/null 2>&1; then
        echo -e "${RED}❌ Postgres (127.0.0.1:5432) not accessible${NC}"
        exit 1
    fi
fi

OWL_LOG="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")/../.." && pwd)/log"
LOG_DIR="${LOG_DIR:-$OWL_LOG}"
mkdir -p "$LOG_DIR"
LOG_FILE="${HEALTH_LOG_FILE:-$LOG_DIR/wisefido-ai-health.log}"

echo -e "${BLUE}📝 Log file:${NC} $LOG_FILE"
echo "==========================================" >> "$LOG_FILE"
echo "wisefido-ai-health starting at $(date)"      >> "$LOG_FILE"
echo "args: $*"                                     >> "$LOG_FILE"
echo "==========================================" >> "$LOG_FILE"

BIN_DIR="$PWD/.bin"
mkdir -p "$BIN_DIR"

if go build -o "$BIN_DIR/wisefido-ai-health" ./cmd/health-etl >/dev/null 2>&1; then
    "$BIN_DIR/wisefido-ai-health" "$@" 2>&1 | tee -a "$LOG_FILE"
else
    echo -e "${YELLOW}go build failed, fallback to go run${NC}"
    go run ./cmd/health-etl "$@" 2>&1 | tee -a "$LOG_FILE"
fi
