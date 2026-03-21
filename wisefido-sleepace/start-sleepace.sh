#!/bin/bash

cd "$(dirname "$0")"

echo "Starting wisefido-sleepace service..."

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

check_port() {
    local port=$1
    local service_name=$2
    
    if ! command -v lsof &> /dev/null; then
        echo -e "${YELLOW}lsof not available, skipping port check${NC}"
        return 0
    fi
    
    if lsof -i :$port > /dev/null 2>&1; then
        echo -e "${YELLOW}Port $port ($service_name) is already in use${NC}"
        local pids=$(lsof -ti :$port)
        if [ -n "$pids" ]; then
            echo -e "${YELLOW}Killing processes on port $port...${NC}"
            for pid in $pids; do
                kill -9 $pid 2>/dev/null
            done
            sleep 1
        fi
    else
        echo -e "${GREEN}Port $port ($service_name) is available${NC}"
    fi
}

check_and_start_dependencies() {
    echo ""
    echo "Checking dependencies..."
    
    SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
    DOCKER_COMPOSE_FILE="${SCRIPT_DIR}/../docker-compose.yml"
    
    if command -v nc > /dev/null 2>&1; then
        if ! nc -zv 127.0.0.1 5432 > /dev/null 2>&1; then
            echo -e "${RED}PostgreSQL (127.0.0.1:5432) is not accessible${NC}"
            return 1
        fi
        if ! nc -zv 127.0.0.1 6379 > /dev/null 2>&1; then
            echo -e "${RED}Redis (127.0.0.1:6379) is not accessible${NC}"
            return 1
        fi
        
        if ! nc -zv 127.0.0.1 1883 > /dev/null 2>&1; then
            echo -e "${YELLOW}MQTT Broker (127.0.0.1:1883) is not accessible${NC}"
            if [ -f "$DOCKER_COMPOSE_FILE" ] && command -v docker-compose > /dev/null 2>&1; then
                echo -e "${BLUE}Starting MQTT Broker...${NC}"
                cd "$(dirname "$DOCKER_COMPOSE_FILE")"
                docker-compose up -d mqtt 2>&1 | grep -v "is up-to-date" || true
                sleep 2
                if nc -zv 127.0.0.1 1883 > /dev/null 2>&1; then
                    echo -e "${GREEN}MQTT Broker started${NC}"
                else
                    echo -e "${RED}Failed to start MQTT Broker${NC}"
                    return 1
                fi
            else
                echo -e "${RED}Cannot start MQTT Broker (docker-compose not found)${NC}"
                return 1
            fi
        else
            echo -e "${GREEN}MQTT Broker (127.0.0.1:1883) is accessible${NC}"
        fi
    fi
    
    return 0
}

check_port 8083 "HTTP Server"

if ! check_and_start_dependencies; then
    echo -e "${RED}Dependencies are not ready${NC}"
    exit 1
fi

ENV="${SLEEPACE_ENV:-dev}"

echo ""
echo -e "${BLUE}Configuration:${NC}"
echo "  HTTP Server: 0.0.0.0:8083"
echo "  Config file: sleepace-${ENV}.yaml"
echo "  Database: localhost:5432/owlrd"
echo "  Redis: localhost:6379"
echo "  MQTT Broker: localhost:1883"
echo ""

echo -e "${GREEN}Starting wisefido-sleepace service (env=${ENV})...${NC}"
echo ""

OWL_LOG="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")/../.." && pwd)/log"
LOG_DIR="${LOG_DIR:-$OWL_LOG}"
mkdir -p "$LOG_DIR"
LOG_FILE="${SLEEPACE_LOG_FILE:-$LOG_DIR/wisefido-sleepace.log}"

echo -e "${BLUE}Log Configuration:${NC}"
echo "  Log File: $LOG_FILE"
echo "  Log Directory: $LOG_DIR"
echo ""

echo "==========================================" >> "$LOG_FILE"
echo "wisefido-sleepace service starting at $(date)" >> "$LOG_FILE"
echo "Log file: $LOG_FILE" >> "$LOG_FILE"
echo "==========================================" >> "$LOG_FILE"

if [ -t 1 ]; then
    echo -e "${GREEN}Logging to: $LOG_FILE${NC}"
    echo ""
    go run cmd/wisefido-sleepace/main.go -env "$ENV" 2>&1 | tee -a "$LOG_FILE"
else
    echo "Logging to: $LOG_FILE" >&2
    go run cmd/wisefido-sleepace/main.go -env "$ENV" 2>&1 | tee -a "$LOG_FILE"
fi
