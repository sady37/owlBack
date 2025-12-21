#!/bin/bash

# Auth 日志监控脚本
# 使用方法: ./scripts/monitor_auth_logs.sh

set -e

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 监控 Docker 容器日志
monitor_docker_logs() {
    echo -e "${BLUE}监控 wisefido-data 容器日志...${NC}"
    echo "按 Ctrl+C 停止监控"
    echo ""
    
    docker-compose logs -f wisefido-data 2>&1 | grep -i --color=always -E "(auth|login|error|failed|success)" || true
}

# 监控特定端点
monitor_endpoints() {
    echo -e "${BLUE}监控特定端点...${NC}"
    
    ENDPOINTS=(
        "/auth/api/v1/login"
        "/auth/api/v1/institutions/search"
    )
    
    for ENDPOINT in "${ENDPOINTS[@]}"; do
        echo -e "${YELLOW}监控: ${ENDPOINT}${NC}"
        docker-compose logs -f wisefido-data 2>&1 | grep --color=always "${ENDPOINT}" || true
    done
}

# 统计错误
count_errors() {
    echo -e "${BLUE}统计错误...${NC}"
    
    ERROR_COUNT=$(docker-compose logs wisefido-data 2>&1 | grep -i "error\|failed" | wc -l | tr -d ' ')
    echo -e "${RED}总错误数: ${ERROR_COUNT}${NC}"
    
    if [ "${ERROR_COUNT}" -gt 0 ]; then
        echo -e "${YELLOW}最近的错误:${NC}"
        docker-compose logs wisefido-data 2>&1 | grep -i "error\|failed" | tail -10
    fi
}

# 统计登录成功/失败
count_login_stats() {
    echo -e "${BLUE}统计登录统计...${NC}"
    
    SUCCESS_COUNT=$(docker-compose logs wisefido-data 2>&1 | grep -i "login successful" | wc -l | tr -d ' ')
    FAILED_COUNT=$(docker-compose logs wisefido-data 2>&1 | grep -i "login failed" | wc -l | tr -d ' ')
    
    echo -e "${GREEN}登录成功: ${SUCCESS_COUNT}${NC}"
    echo -e "${RED}登录失败: ${FAILED_COUNT}${NC}"
    
    if [ "${FAILED_COUNT}" -gt 0 ]; then
        echo -e "${YELLOW}最近的登录失败:${NC}"
        docker-compose logs wisefido-data 2>&1 | grep -i "login failed" | tail -5
    fi
}

# 主菜单
main() {
    echo "=========================================="
    echo "Auth 日志监控工具"
    echo "=========================================="
    echo "1. 实时监控所有日志"
    echo "2. 监控特定端点"
    echo "3. 统计错误"
    echo "4. 统计登录统计"
    echo "5. 退出"
    echo "=========================================="
    
    read -p "请选择 (1-5): " choice
    
    case $choice in
        1)
            monitor_docker_logs
            ;;
        2)
            monitor_endpoints
            ;;
        3)
            count_errors
            ;;
        4)
            count_login_stats
            ;;
        5)
            echo "退出"
            exit 0
            ;;
        *)
            echo "无效选择"
            exit 1
            ;;
    esac
}

# 运行主菜单
main

