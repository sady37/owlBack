#!/bin/bash

cd "$(dirname "$0")"

echo "🚀 Starting wisefido-cardagg service..."

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

check_and_start_dependencies() {
    echo ""
    echo "🔍 Checking dependencies..."
    
    if command -v nc > /dev/null 2>&1; then
        if ! nc -zv 127.0.0.1 6379 > /dev/null 2>&1; then
            echo -e "${RED}❌ Redis (127.0.0.1:6379) is not accessible${NC}"
            return 1
        fi
    else
        echo -e "${YELLOW}⚠️  nc not available, skipping Redis connectivity check${NC}"
    fi
    
    return 0
}

if ! check_and_start_dependencies; then
    echo -e "${RED}❌ Dependencies are not ready${NC}"
    exit 1
fi

echo ""
echo "📋 Setting environment variables..."

export REDIS_ADDR="${REDIS_ADDR:-127.0.0.1:6379}"
export REDIS_PASSWORD="${REDIS_PASSWORD:-TeLunSu-36kr}"
export LOG_LEVEL="${LOG_LEVEL:-info}"
export LOG_FORMAT="${LOG_FORMAT:-json}"

echo ""
echo -e "${BLUE}📊 Configuration:${NC}"
echo "  📡 Redis: $REDIS_ADDR"
echo "  📊 Log Level: $LOG_LEVEL"
echo ""

echo -e "${GREEN}🚀 Starting wisefido-cardagg service...${NC}"
echo ""

# 日志目录（与 start_owlback.sh 统一）
OWL_LOG="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")/../.." && pwd)/log"
LOG_DIR="${LOG_DIR:-$OWL_LOG}"
mkdir -p "$LOG_DIR"
LOG_FILE="${CARDAGG_LOG_FILE:-$LOG_DIR/wisefido-cardagg.log}"

echo -e "${BLUE}📝 Log Configuration:${NC}"
echo "  📄 Log File: $LOG_FILE"
echo "  📁 Log Directory: $LOG_DIR"
echo ""

# 写入启动信息到日志文件
echo "==========================================" >> "$LOG_FILE"
echo "wisefido-cardagg service starting at $(date)" >> "$LOG_FILE"
echo "Log file: $LOG_FILE" >> "$LOG_FILE"
echo "==========================================" >> "$LOG_FILE"

# 同时输出到控制台和日志文件
if [ -t 1 ]; then
    echo -e "${GREEN}✅ Logging to: $LOG_FILE${NC}"
    echo -e "${GREEN}✅ Output will be displayed in terminal and saved to log file${NC}"
    echo ""
    go run main.go 2>&1 | tee -a "$LOG_FILE"
else
    echo "Logging to: $LOG_FILE" >&2
    go run main.go 2>&1 | tee -a "$LOG_FILE"
fi
