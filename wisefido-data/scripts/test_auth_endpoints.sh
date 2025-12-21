#!/bin/bash

# Auth 端点端到端测试脚本
# 使用方法: ./scripts/test_auth_endpoints.sh

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"
TEST_TENANT_ID="00000000-0000-0000-0000-000000000001"
TEST_USER_ACCOUNT="sysadmin"
TEST_PASSWORD="ChangeMe123!"

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印测试标题
print_test() {
    echo -e "\n${YELLOW}=== $1 ===${NC}"
}

# 打印成功
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# 打印失败
print_fail() {
    echo -e "${RED}✗ $1${NC}"
}

# 计算 SHA256 hash (hex)
hash_string() {
    echo -n "$1" | sha256sum | cut -d' ' -f1
}

# 计算账号 hash (lowercase)
hash_account() {
    echo -n "$1" | tr '[:upper:]' '[:lower:]' | sha256sum | cut -d' ' -f1
}

# 计算密码 hash
hash_password() {
    echo -n "$1" | sha256sum | cut -d' ' -f1
}

# 检查服务是否运行
check_service() {
    print_test "检查服务状态"
    if curl -s -f "${BASE_URL}/health" > /dev/null; then
        print_success "服务运行正常"
        return 0
    else
        print_fail "服务未运行或无法访问 ${BASE_URL}"
        echo "请确保服务已启动: docker-compose up -d wisefido-data"
        exit 1
    fi
}

# 测试登录端点
test_login() {
    print_test "测试 POST /auth/api/v1/login"
    
    ACCOUNT_HASH=$(hash_account "${TEST_USER_ACCOUNT}")
    PASSWORD_HASH=$(hash_password "${TEST_PASSWORD}")
    
    echo "账号: ${TEST_USER_ACCOUNT}"
    echo "账号 Hash: ${ACCOUNT_HASH}"
    echo "密码 Hash: ${PASSWORD_HASH}"
    
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${BASE_URL}/auth/api/v1/login" \
        -H "Content-Type: application/json" \
        -d "{
            \"tenant_id\": \"${TEST_TENANT_ID}\",
            \"userType\": \"staff\",
            \"accountHash\": \"${ACCOUNT_HASH}\",
            \"passwordHash\": \"${PASSWORD_HASH}\"
        }")
    
    HTTP_CODE=$(echo "${RESPONSE}" | tail -n1)
    BODY=$(echo "${RESPONSE}" | sed '$d')
    
    echo "HTTP 状态码: ${HTTP_CODE}"
    echo "响应: ${BODY}" | jq '.' 2>/dev/null || echo "${BODY}"
    
    if [ "${HTTP_CODE}" = "200" ]; then
        CODE=$(echo "${BODY}" | jq -r '.code' 2>/dev/null || echo "")
        if [ "${CODE}" = "2000" ]; then
            print_success "登录成功"
            USER_ACCOUNT=$(echo "${BODY}" | jq -r '.result.user_account' 2>/dev/null || echo "")
            if [ "${USER_ACCOUNT}" = "${TEST_USER_ACCOUNT}" ]; then
                print_success "用户账号匹配: ${USER_ACCOUNT}"
            else
                print_fail "用户账号不匹配: 期望 ${TEST_USER_ACCOUNT}, 得到 ${USER_ACCOUNT}"
            fi
            return 0
        else
            print_fail "登录失败: code=${CODE}"
            echo "${BODY}" | jq '.' 2>/dev/null || echo "${BODY}"
            return 1
        fi
    else
        print_fail "HTTP 状态码错误: ${HTTP_CODE}"
        return 1
    fi
}

# 测试登录 - 缺少凭证
test_login_missing_credentials() {
    print_test "测试 POST /auth/api/v1/login - 缺少凭证"
    
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${BASE_URL}/auth/api/v1/login" \
        -H "Content-Type: application/json" \
        -d "{
            \"tenant_id\": \"${TEST_TENANT_ID}\",
            \"userType\": \"staff\"
        }")
    
    HTTP_CODE=$(echo "${RESPONSE}" | tail -n1)
    BODY=$(echo "${RESPONSE}" | sed '$d')
    
    echo "HTTP 状态码: ${HTTP_CODE}"
    echo "响应: ${BODY}" | jq '.' 2>/dev/null || echo "${BODY}"
    
    if [ "${HTTP_CODE}" = "200" ]; then
        CODE=$(echo "${BODY}" | jq -r '.code' 2>/dev/null || echo "")
        if [ "${CODE}" = "-1" ]; then
            print_success "错误处理正确: code=-1"
            return 0
        else
            print_fail "错误处理不正确: code=${CODE}"
            return 1
        fi
    else
        print_fail "HTTP 状态码错误: ${HTTP_CODE}"
        return 1
    fi
}

# 测试搜索机构
test_search_institutions() {
    print_test "测试 GET /auth/api/v1/institutions/search"
    
    ACCOUNT_HASH=$(hash_account "${TEST_USER_ACCOUNT}")
    PASSWORD_HASH=$(hash_password "${TEST_PASSWORD}")
    
    RESPONSE=$(curl -s -w "\n%{http_code}" \
        "${BASE_URL}/auth/api/v1/institutions/search?accountHash=${ACCOUNT_HASH}&passwordHash=${PASSWORD_HASH}&userType=staff")
    
    HTTP_CODE=$(echo "${RESPONSE}" | tail -n1)
    BODY=$(echo "${RESPONSE}" | sed '$d')
    
    echo "HTTP 状态码: ${HTTP_CODE}"
    echo "响应: ${BODY}" | jq '.' 2>/dev/null || echo "${BODY}"
    
    if [ "${HTTP_CODE}" = "200" ]; then
        CODE=$(echo "${BODY}" | jq -r '.code' 2>/dev/null || echo "")
        if [ "${CODE}" = "2000" ]; then
            print_success "搜索成功"
            COUNT=$(echo "${BODY}" | jq '.result | length' 2>/dev/null || echo "0")
            if [ "${COUNT}" -gt "0" ]; then
                print_success "找到 ${COUNT} 个机构"
                echo "${BODY}" | jq '.result' 2>/dev/null || echo ""
            else
                print_fail "未找到机构"
            fi
            return 0
        else
            print_fail "搜索失败: code=${CODE}"
            return 1
        fi
    else
        print_fail "HTTP 状态码错误: ${HTTP_CODE}"
        return 1
    fi
}

# 测试搜索机构 - 无匹配
test_search_institutions_no_match() {
    print_test "测试 GET /auth/api/v1/institutions/search - 无匹配"
    
    INVALID_HASH="0000000000000000000000000000000000000000000000000000000000000000"
    
    RESPONSE=$(curl -s -w "\n%{http_code}" \
        "${BASE_URL}/auth/api/v1/institutions/search?accountHash=${INVALID_HASH}&passwordHash=${INVALID_HASH}&userType=staff")
    
    HTTP_CODE=$(echo "${RESPONSE}" | tail -n1)
    BODY=$(echo "${RESPONSE}" | sed '$d')
    
    echo "HTTP 状态码: ${HTTP_CODE}"
    echo "响应: ${BODY}" | jq '.' 2>/dev/null || echo "${BODY}"
    
    if [ "${HTTP_CODE}" = "200" ]; then
        CODE=$(echo "${BODY}" | jq -r '.code' 2>/dev/null || echo "")
        COUNT=$(echo "${BODY}" | jq '.result | length' 2>/dev/null || echo "0")
        if [ "${CODE}" = "2000" ] && [ "${COUNT}" = "0" ]; then
            print_success "无匹配时返回空数组"
            return 0
        else
            print_fail "响应不正确: code=${CODE}, count=${COUNT}"
            return 1
        fi
    else
        print_fail "HTTP 状态码错误: ${HTTP_CODE}"
        return 1
    fi
}

# 测试密码重置端点（待实现）
test_forgot_password_endpoints() {
    print_test "测试密码重置端点（待实现）"
    
    ENDPOINTS=(
        "/auth/api/v1/forgot-password/send-code"
        "/auth/api/v1/forgot-password/verify-code"
        "/auth/api/v1/forgot-password/reset"
    )
    
    for ENDPOINT in "${ENDPOINTS[@]}"; do
        echo -e "\n测试: ${ENDPOINT}"
        RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${BASE_URL}${ENDPOINT}" \
            -H "Content-Type: application/json" \
            -d '{"account": "test"}')
        
        HTTP_CODE=$(echo "${RESPONSE}" | tail -n1)
        BODY=$(echo "${RESPONSE}" | sed '$d')
        
        CODE=$(echo "${BODY}" | jq -r '.code' 2>/dev/null || echo "")
        MESSAGE=$(echo "${BODY}" | jq -r '.message' 2>/dev/null || echo "")
        
        if [ "${HTTP_CODE}" = "200" ] && [ "${CODE}" = "-1" ]; then
            print_success "${ENDPOINT}: 返回错误（待实现）"
        else
            print_fail "${ENDPOINT}: 响应不正确"
            echo "HTTP: ${HTTP_CODE}, Code: ${CODE}, Message: ${MESSAGE}"
        fi
    done
}

# 主测试流程
main() {
    echo "=========================================="
    echo "Auth 端点端到端测试"
    echo "=========================================="
    echo "服务地址: ${BASE_URL}"
    echo "测试租户: ${TEST_TENANT_ID}"
    echo "测试用户: ${TEST_USER_ACCOUNT}"
    echo "=========================================="
    
    TOTAL_TESTS=0
    PASSED_TESTS=0
    FAILED_TESTS=0
    
    # 检查服务
    check_service
    ((TOTAL_TESTS++))
    ((PASSED_TESTS++))
    
    # 测试登录
    if test_login; then
        ((PASSED_TESTS++))
    else
        ((FAILED_TESTS++))
    fi
    ((TOTAL_TESTS++))
    
    # 测试登录 - 缺少凭证
    if test_login_missing_credentials; then
        ((PASSED_TESTS++))
    else
        ((FAILED_TESTS++))
    fi
    ((TOTAL_TESTS++))
    
    # 测试搜索机构
    if test_search_institutions; then
        ((PASSED_TESTS++))
    else
        ((FAILED_TESTS++))
    fi
    ((TOTAL_TESTS++))
    
    # 测试搜索机构 - 无匹配
    if test_search_institutions_no_match; then
        ((PASSED_TESTS++))
    else
        ((FAILED_TESTS++))
    fi
    ((TOTAL_TESTS++))
    
    # 测试密码重置端点
    test_forgot_password_endpoints
    ((TOTAL_TESTS++))
    ((PASSED_TESTS++))
    
    # 测试总结
    echo -e "\n=========================================="
    echo "测试总结"
    echo "=========================================="
    echo "总测试数: ${TOTAL_TESTS}"
    echo -e "${GREEN}通过: ${PASSED_TESTS}${NC}"
    echo -e "${RED}失败: ${FAILED_TESTS}${NC}"
    echo "=========================================="
    
    if [ "${FAILED_TESTS}" -eq 0 ]; then
        echo -e "${GREEN}所有测试通过！${NC}"
        exit 0
    else
        echo -e "${RED}有测试失败，请检查日志${NC}"
        exit 1
    fi
}

# 运行主测试
main

