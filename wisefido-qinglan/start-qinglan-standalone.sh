#!/bin/bash
# wisefido-qinglan 独立启动脚本
# Usage: ./start-qinglan-standalone.sh [start|stop|restart|status]

set -e

# 配色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 配置
SERVICE_NAME="wisefido-qinglan"
LOG_DIR="/tmp/owlback-logs"
LOG_FILE="$LOG_DIR/$SERVICE_NAME.log"
PID_FILE="/tmp/$SERVICE_NAME.pid"

# 默认配置（可通过环境变量覆盖）
: ${DB_HOST:=localhost}
: ${DB_PORT:=5432}
: ${DB_USER:=postgres}
: ${DB_PASSWORD:=postgres}
: ${DB_NAME:=owlrd}
: ${DB_SSLMODE:=disable}

: ${REDIS_ADDR:=localhost:6379}
: ${REDIS_PASSWORD:=TeLunSu-36kr}
: ${REDIS_DB:=0}

: ${MQTT_BROKER:=localhost}
: ${MQTT_PORT:=1883}
: ${MQTT_CLIENT_ID:=wisefido-qinglan}

: ${RADAR_MQTT_SERVER:=10.0.0.30}
: ${RADAR_MQTT_PORT:=8883}
: ${RADAR_MQTT_ACCOUNT:=wfiot}
: ${RADAR_MQTT_PASSWORD:=tt@wf@2025}
: ${RADAR_MQTT_PRODUCT_ID:=88}

: ${HTTP_HOST:=0.0.0.0}
: ${HTTP_PORT:=8081}

: ${HTTPS_PORT:=8443}
: ${HTTPS_CERT_FILE:=$PWD/certs/server.crt}
: ${HTTPS_KEY_FILE:=$PWD/certs/server.key}

: ${LOG_LEVEL:=info}

# 创建日志目录
mkdir -p "$LOG_DIR"

# 检查端口占用
check_port() {
    local port=$1
    if netstat -tuln 2>/dev/null | grep -q ":$port "; then
        return 0
    elif ss -tuln 2>/dev/null | grep -q ":$port "; then
        return 0
    fi
    return 1
}

# 检查依赖服务
check_dependencies() {
    echo -e "${YELLOW}Checking dependencies...${NC}"

    local all_ok=true

    # 检查 PostgreSQL
    echo -n "  PostgreSQL ($DB_HOST:$DB_PORT): "
    if timeout 2 bash -c "cat < /dev/null > /dev/tcp/$DB_HOST/$DB_PORT" 2>/dev/null; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗ (not reachable)${NC}"
        all_ok=false
    fi

    # 检查 Redis
    REDIS_HOST=$(echo $REDIS_ADDR | cut -d: -f1)
    REDIS_PORT=$(echo $REDIS_ADDR | cut -d: -f2)
    echo -n "  Redis ($REDIS_ADDR): "
    if timeout 2 bash -c "cat < /dev/null > /dev/tcp/$REDIS_HOST/$REDIS_PORT" 2>/dev/null; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗ (not reachable)${NC}"
        all_ok=false
    fi

    # 检查 MQTT
    echo -n "  MQTT Broker ($MQTT_BROKER:$MQTT_PORT): "
    if timeout 2 bash -c "cat < /dev/null > /dev/tcp/$MQTT_BROKER/$MQTT_PORT" 2>/dev/null; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗ (not reachable)${NC}"
        all_ok=false
    fi

    # 检查 HTTPS 证书（如果配置了HTTPS）
    if [ "$HTTPS_PORT" != "0" ] && [ -n "$HTTPS_CERT_FILE" ]; then
        echo -n "  HTTPS Certificate: "
        if [ -f "$HTTPS_CERT_FILE" ] && [ -f "$HTTPS_KEY_FILE" ]; then
            echo -e "${GREEN}✓${NC}"
        else
            echo -e "${YELLOW}⚠ (not found, HTTPS will not start)${NC}"
        fi
    fi

    # 检查 owl-common 依赖
    echo -n "  owl-common module: "
    if [ -d "../owl-common" ]; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗ (../owl-common not found)${NC}"
        all_ok=false
    fi

    if [ "$all_ok" = false ]; then
        echo -e "${RED}Some dependencies are missing. Please check configuration.${NC}"
        return 1
    fi

    echo -e "${GREEN}All dependencies OK${NC}"
    return 0
}

# 获取进程PID
get_pid() {
    if [ -f "$PID_FILE" ]; then
        cat "$PID_FILE"
    else
        # 尝试通过进程名查找
        pgrep -f "go run.*wisefido-qinglan/main.go" || \
        pgrep -f "wisefido-qinglan" | head -1
    fi
}

# 检查服务状态
is_running() {
    local pid=$(get_pid)
    if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
        return 0
    fi
    return 1
}

# 启动服务
start() {
    echo -e "${GREEN}Starting $SERVICE_NAME...${NC}"

    if is_running; then
        echo -e "${YELLOW}$SERVICE_NAME is already running (PID: $(get_pid))${NC}"
        return 0
    fi

    # 检查依赖
    if ! check_dependencies; then
        echo -e "${RED}Failed to start: dependencies check failed${NC}"
        return 1
    fi

    # 检查端口占用
    if check_port $HTTP_PORT; then
        echo -e "${RED}Port $HTTP_PORT is already in use${NC}"
        return 1
    fi

    if [ "$HTTPS_PORT" != "0" ] && check_port $HTTPS_PORT; then
        echo -e "${RED}Port $HTTPS_PORT is already in use${NC}"
        return 1
    fi

    # 设置环境变量
    export DB_HOST DB_PORT DB_USER DB_PASSWORD DB_NAME DB_SSLMODE
    export REDIS_ADDR REDIS_PASSWORD REDIS_DB
    export MQTT_BROKER MQTT_PORT MQTT_CLIENT_ID
    export RADAR_MQTT_SERVER RADAR_MQTT_PORT RADAR_MQTT_ACCOUNT RADAR_MQTT_PASSWORD RADAR_MQTT_PRODUCT_ID
    export HTTP_HOST HTTP_PORT
    export HTTPS_PORT HTTPS_CERT_FILE HTTPS_KEY_FILE
    export LOG_LEVEL

    # 启动服务
    echo "Starting $SERVICE_NAME..."
    echo "Log file: $LOG_FILE"

    cd "$(dirname "$0")"

    # 检查是否有编译的二进制文件
    if [ -f "./wisefido-qinglan" ]; then
        echo "Using compiled binary..."
        nohup ./wisefido-qinglan >> "$LOG_FILE" 2>&1 &
    else
        echo "Using go run..."
        nohup go run cmd/wisefido-qinglan/main.go >> "$LOG_FILE" 2>&1 &
    fi

    local pid=$!
    echo $pid > "$PID_FILE"

    # 等待服务启动
    echo -n "Waiting for service to start"
    for i in {1..30}; do
        sleep 1
        echo -n "."
        if check_port $HTTP_PORT; then
            echo
            echo -e "${GREEN}✓ $SERVICE_NAME started successfully (PID: $pid)${NC}"
            echo "  HTTP:  http://localhost:$HTTP_PORT"
            if [ "$HTTPS_PORT" != "0" ]; then
                echo "  HTTPS: https://localhost:$HTTPS_PORT"
            fi
            echo "  Log:   tail -f $LOG_FILE"
            return 0
        fi
    done

    echo
    echo -e "${RED}✗ $SERVICE_NAME failed to start${NC}"
    echo "Check logs: tail -f $LOG_FILE"
    return 1
}

# 停止服务
stop() {
    echo -e "${YELLOW}Stopping $SERVICE_NAME...${NC}"

    local pid=$(get_pid)

    if [ -z "$pid" ]; then
        echo -e "${YELLOW}$SERVICE_NAME is not running${NC}"
        rm -f "$PID_FILE"
        return 0
    fi

    # 发送 SIGTERM
    echo "Sending SIGTERM to process $pid..."
    kill -TERM "$pid" 2>/dev/null || true

    # 等待进程退出
    echo -n "Waiting for process to stop"
    for i in {1..10}; do
        if ! kill -0 "$pid" 2>/dev/null; then
            echo
            echo -e "${GREEN}✓ $SERVICE_NAME stopped${NC}"
            rm -f "$PID_FILE"
            return 0
        fi
        sleep 1
        echo -n "."
    done

    # 强制杀死
    echo
    echo "Process did not stop gracefully, forcing..."
    kill -9 "$pid" 2>/dev/null || true

    # 清理其他进程
    pkill -9 -f "go run.*wisefido-qinglan" 2>/dev/null || true

    rm -f "$PID_FILE"
    echo -e "${GREEN}✓ $SERVICE_NAME stopped (forced)${NC}"
}

# 重启服务
restart() {
    stop
    sleep 2
    start
}

# 查看状态
status() {
    echo "=== $SERVICE_NAME Status ==="

    if is_running; then
        local pid=$(get_pid)
        echo -e "Status: ${GREEN}RUNNING${NC} (PID: $pid)"

        # 检查端口
        echo -n "HTTP Port ($HTTP_PORT): "
        if check_port $HTTP_PORT; then
            echo -e "${GREEN}✓ listening${NC}"
        else
            echo -e "${RED}✗ not listening${NC}"
        fi

        if [ "$HTTPS_PORT" != "0" ]; then
            echo -n "HTTPS Port ($HTTPS_PORT): "
            if check_port $HTTPS_PORT; then
                echo -e "${GREEN}✓ listening${NC}"
            else
                echo -e "${YELLOW}⚠ not listening${NC}"
            fi
        fi

        # 进程信息
        echo ""
        echo "Process info:"
        ps -p $pid -o pid,vsz,rss,etime,cmd --no-headers | awk '{printf "  PID: %s\n  Memory: %s KB\n  Runtime: %s\n  Command: %s\n", $1, $3, $4, substr($0, index($0,$5))}'

    else
        echo -e "Status: ${RED}STOPPED${NC}"
    fi

    echo ""
    echo "Log file: $LOG_FILE"
    if [ -f "$LOG_FILE" ]; then
        echo "Last 5 lines:"
        tail -5 "$LOG_FILE" | sed 's/^/  /'
    fi
}

# 查看日志
logs() {
    if [ -f "$LOG_FILE" ]; then
        tail -f "$LOG_FILE"
    else
        echo -e "${RED}Log file not found: $LOG_FILE${NC}"
        exit 1
    fi
}

# 主函数
main() {
    local action=${1:-start}

    case "$action" in
        start)
            start
            ;;
        stop)
            stop
            ;;
        restart)
            restart
            ;;
        status)
            status
            ;;
        logs)
            logs
            ;;
        check)
            check_dependencies
            ;;
        *)
            echo "Usage: $0 {start|stop|restart|status|logs|check}"
            echo ""
            echo "Commands:"
            echo "  start   - Start the service"
            echo "  stop    - Stop the service"
            echo "  restart - Restart the service"
            echo "  status  - Show service status"
            echo "  logs    - Tail service logs"
            echo "  check   - Check dependencies"
            echo ""
            echo "Environment variables (optional):"
            echo "  DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME"
            echo "  REDIS_ADDR, REDIS_PASSWORD"
            echo "  MQTT_BROKER, MQTT_PORT"
            echo "  RADAR_MQTT_SERVER, RADAR_MQTT_PORT, RADAR_MQTT_PASSWORD"
            echo "  HTTP_PORT, HTTPS_PORT"
            echo "  HTTPS_CERT_FILE, HTTPS_KEY_FILE"
            exit 1
            ;;
    esac
}

main "$@"
