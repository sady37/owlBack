#!/bin/bash

# Device 端点端到端测试脚本
# 使用方法: ./scripts/test_device_endpoints.sh

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"
TEST_TENANT_ID="00000000-0000-0000-0000-000000000002"
TEST_DEVICE_ID="00000000-0000-0000-0000-000000000002"

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

# 检查服务是否运行
check_service() {
    print_test "检查服务状态"
    # 尝试访问设备端点来检查服务是否运行
    RESPONSE=$(curl -s -w "\n%{http_code}" "${BASE_URL}/admin/api/v1/devices?tenant_id=${TEST_TENANT_ID}" -H "X-Tenant-Id: ${TEST_TENANT_ID}" 2>&1)
    HTTP_CODE=$(echo "${RESPONSE}" | tail -n1)
    if [ "${HTTP_CODE}" = "200" ] || [ "${HTTP_CODE}" = "404" ]; then
        print_success "服务运行正常 (HTTP ${HTTP_CODE})"
        return 0
    else
        print_fail "服务未运行或无法访问 ${BASE_URL} (HTTP ${HTTP_CODE})"
        echo "请确保服务已启动: docker-compose up -d wisefido-data"
        exit 1
    fi
}

# 测试查询设备列表
test_list_devices() {
    print_test "测试 GET /admin/api/v1/devices"
    
    RESPONSE=$(curl -s -w "\n%{http_code}" -X GET \
        "${BASE_URL}/admin/api/v1/devices?tenant_id=${TEST_TENANT_ID}&page=1&size=20" \
        -H "X-Tenant-Id: ${TEST_TENANT_ID}")
    
    HTTP_CODE=$(echo "${RESPONSE}" | tail -n1)
    BODY=$(echo "${RESPONSE}" | sed '$d')
    
    echo "HTTP 状态码: ${HTTP_CODE}"
    echo "响应: ${BODY}" | jq '.' 2>/dev/null || echo "${BODY}"
    
    if [ "${HTTP_CODE}" = "200" ]; then
        CODE=$(echo "${BODY}" | jq -r '.code' 2>/dev/null || echo "")
        if [ "${CODE}" = "2000" ]; then
            print_success "查询设备列表成功"
            ITEMS_COUNT=$(echo "${BODY}" | jq '.result.items | length' 2>/dev/null || echo "0")
            TOTAL=$(echo "${BODY}" | jq '.result.total' 2>/dev/null || echo "0")
            echo "设备数量: ${ITEMS_COUNT}, 总数: ${TOTAL}"
            return 0
        else
            print_fail "查询失败: code=${CODE}"
            return 1
        fi
    else
        print_fail "HTTP 状态码错误: ${HTTP_CODE}"
        return 1
    fi
}

# 测试查询设备列表 - 过滤条件
test_list_devices_with_filters() {
    print_test "测试 GET /admin/api/v1/devices - 过滤条件"
    
    # 按状态过滤
    RESPONSE=$(curl -s -w "\n%{http_code}" -X GET \
        "${BASE_URL}/admin/api/v1/devices?tenant_id=${TEST_TENANT_ID}&status=online" \
        -H "X-Tenant-Id: ${TEST_TENANT_ID}")
    
    HTTP_CODE=$(echo "${RESPONSE}" | tail -n1)
    BODY=$(echo "${RESPONSE}" | sed '$d')
    
    if [ "${HTTP_CODE}" = "200" ]; then
        CODE=$(echo "${BODY}" | jq -r '.code' 2>/dev/null || echo "")
        if [ "${CODE}" = "2000" ]; then
            print_success "按状态过滤成功"
        else
            print_fail "按状态过滤失败: code=${CODE}"
            return 1
        fi
    else
        print_fail "HTTP 状态码错误: ${HTTP_CODE}"
        return 1
    fi
    
    # 按业务访问权限过滤
    RESPONSE=$(curl -s -w "\n%{http_code}" -X GET \
        "${BASE_URL}/admin/api/v1/devices?tenant_id=${TEST_TENANT_ID}&business_access=approved" \
        -H "X-Tenant-Id: ${TEST_TENANT_ID}")
    
    HTTP_CODE=$(echo "${RESPONSE}" | tail -n1)
    BODY=$(echo "${RESPONSE}" | sed '$d')
    
    if [ "${HTTP_CODE}" = "200" ]; then
        CODE=$(echo "${BODY}" | jq -r '.code' 2>/dev/null || echo "")
        if [ "${CODE}" = "2000" ]; then
            print_success "按业务访问权限过滤成功"
        else
            print_fail "按业务访问权限过滤失败: code=${CODE}"
            return 1
        fi
    else
        print_fail "HTTP 状态码错误: ${HTTP_CODE}"
        return 1
    fi
}

# 测试查询设备详情
test_get_device() {
    print_test "测试 GET /admin/api/v1/devices/:id"
    
    RESPONSE=$(curl -s -w "\n%{http_code}" -X GET \
        "${BASE_URL}/admin/api/v1/devices/${TEST_DEVICE_ID}" \
        -H "X-Tenant-Id: ${TEST_TENANT_ID}")
    
    HTTP_CODE=$(echo "${RESPONSE}" | tail -n1)
    BODY=$(echo "${RESPONSE}" | sed '$d')
    
    echo "HTTP 状态码: ${HTTP_CODE}"
    echo "响应: ${BODY}" | jq '.' 2>/dev/null || echo "${BODY}"
    
    if [ "${HTTP_CODE}" = "200" ]; then
        CODE=$(echo "${BODY}" | jq -r '.code' 2>/dev/null || echo "")
        if [ "${CODE}" = "2000" ]; then
            print_success "查询设备详情成功"
            DEVICE_NAME=$(echo "${BODY}" | jq -r '.result.device_name' 2>/dev/null || echo "")
            if [ "${DEVICE_NAME}" != "" ]; then
                print_success "设备名称: ${DEVICE_NAME}"
            fi
            return 0
        else
            print_fail "查询失败: code=${CODE}"
            return 1
        fi
    else
        print_fail "HTTP 状态码错误: ${HTTP_CODE}"
        return 1
    fi
}

# 测试查询设备详情 - 设备不存在
test_get_device_not_found() {
    print_test "测试 GET /admin/api/v1/devices/:id - 设备不存在"
    
    INVALID_DEVICE_ID="00000000-0000-0000-0000-000000000000"
    
    RESPONSE=$(curl -s -w "\n%{http_code}" -X GET \
        "${BASE_URL}/admin/api/v1/devices/${INVALID_DEVICE_ID}" \
        -H "X-Tenant-Id: ${TEST_TENANT_ID}")
    
    HTTP_CODE=$(echo "${RESPONSE}" | tail -n1)
    BODY=$(echo "${RESPONSE}" | sed '$d')
    
    if [ "${HTTP_CODE}" = "200" ]; then
        CODE=$(echo "${BODY}" | jq -r '.code' 2>/dev/null || echo "")
        MESSAGE=$(echo "${BODY}" | jq -r '.message' 2>/dev/null || echo "")
        if [ "${CODE}" = "-1" ] && [ "${MESSAGE}" = "device not found" ]; then
            print_success "错误处理正确: device not found"
            return 0
        else
            print_fail "错误处理不正确: code=${CODE}, message=${MESSAGE}"
            return 1
        fi
    else
        print_fail "HTTP 状态码错误: ${HTTP_CODE}"
        return 1
    fi
}

# 测试更新设备
test_update_device() {
    print_test "测试 PUT /admin/api/v1/devices/:id"
    
    RESPONSE=$(curl -s -w "\n%{http_code}" -X PUT \
        "${BASE_URL}/admin/api/v1/devices/${TEST_DEVICE_ID}" \
        -H "Content-Type: application/json" \
        -H "X-Tenant-Id: ${TEST_TENANT_ID}" \
        -d '{
            "device_name": "Updated Device Name",
            "status": "offline",
            "business_access": "pending",
            "monitoring_enabled": false
        }')
    
    HTTP_CODE=$(echo "${RESPONSE}" | tail -n1)
    BODY=$(echo "${RESPONSE}" | sed '$d')
    
    echo "HTTP 状态码: ${HTTP_CODE}"
    echo "响应: ${BODY}" | jq '.' 2>/dev/null || echo "${BODY}"
    
    if [ "${HTTP_CODE}" = "200" ]; then
        CODE=$(echo "${BODY}" | jq -r '.code' 2>/dev/null || echo "")
        if [ "${CODE}" = "2000" ]; then
            SUCCESS=$(echo "${BODY}" | jq -r '.result.success' 2>/dev/null || echo "")
            if [ "${SUCCESS}" = "true" ]; then
                print_success "更新设备成功"
                return 0
            else
                print_fail "更新失败: success=${SUCCESS}"
                return 1
            fi
        else
            print_fail "更新失败: code=${CODE}"
            return 1
        fi
    else
        print_fail "HTTP 状态码错误: ${HTTP_CODE}"
        return 1
    fi
}

# 测试更新设备 - 绑定验证
test_update_device_binding_validation() {
    print_test "测试 PUT /admin/api/v1/devices/:id - 绑定验证"
    
    RESPONSE=$(curl -s -w "\n%{http_code}" -X PUT \
        "${BASE_URL}/admin/api/v1/devices/${TEST_DEVICE_ID}" \
        -H "Content-Type: application/json" \
        -H "X-Tenant-Id: ${TEST_TENANT_ID}" \
        -d '{
            "unit_id": "00000000-0000-0000-0000-000000000001"
        }')
    
    HTTP_CODE=$(echo "${RESPONSE}" | tail -n1)
    BODY=$(echo "${RESPONSE}" | sed '$d')
    
    if [ "${HTTP_CODE}" = "200" ]; then
        CODE=$(echo "${BODY}" | jq -r '.code' 2>/dev/null || echo "")
        MESSAGE=$(echo "${BODY}" | jq -r '.message' 2>/dev/null || echo "")
        if [ "${CODE}" = "-1" ] && [[ "${MESSAGE}" == *"invalid binding"* ]]; then
            print_success "绑定验证正确: ${MESSAGE}"
            return 0
        else
            print_fail "绑定验证不正确: code=${CODE}, message=${MESSAGE}"
            return 1
        fi
    else
        print_fail "HTTP 状态码错误: ${HTTP_CODE}"
        return 1
    fi
}

# 测试删除设备
test_delete_device() {
    print_test "测试 DELETE /admin/api/v1/devices/:id"
    
    RESPONSE=$(curl -s -w "\n%{http_code}" -X DELETE \
        "${BASE_URL}/admin/api/v1/devices/${TEST_DEVICE_ID}" \
        -H "X-Tenant-Id: ${TEST_TENANT_ID}")
    
    HTTP_CODE=$(echo "${RESPONSE}" | tail -n1)
    BODY=$(echo "${RESPONSE}" | sed '$d')
    
    echo "HTTP 状态码: ${HTTP_CODE}"
    echo "响应: ${BODY}" | jq '.' 2>/dev/null || echo "${BODY}"
    
    if [ "${HTTP_CODE}" = "200" ]; then
        CODE=$(echo "${BODY}" | jq -r '.code' 2>/dev/null || echo "")
        if [ "${CODE}" = "2000" ]; then
            SUCCESS=$(echo "${BODY}" | jq -r '.result.success' 2>/dev/null || echo "")
            if [ "${SUCCESS}" = "true" ]; then
                print_success "删除设备成功（软删除）"
                return 0
            else
                print_fail "删除失败: success=${SUCCESS}"
                return 1
            fi
        else
            print_fail "删除失败: code=${CODE}"
            return 1
        fi
    else
        print_fail "HTTP 状态码错误: ${HTTP_CODE}"
        return 1
    fi
}

# 主测试流程
main() {
    echo "=========================================="
    echo "Device 端点端到端测试"
    echo "=========================================="
    echo "服务地址: ${BASE_URL}"
    echo "测试租户: ${TEST_TENANT_ID}"
    echo "测试设备: ${TEST_DEVICE_ID}"
    echo "=========================================="
    
    TOTAL_TESTS=0
    PASSED_TESTS=0
    FAILED_TESTS=0
    
    # 检查服务
    check_service
    ((TOTAL_TESTS++))
    ((PASSED_TESTS++))
    
    # 测试查询设备列表
    if test_list_devices; then
        ((PASSED_TESTS++))
    else
        ((FAILED_TESTS++))
    fi
    ((TOTAL_TESTS++))
    
    # 测试查询设备列表 - 过滤条件
    if test_list_devices_with_filters; then
        ((PASSED_TESTS++))
    else
        ((FAILED_TESTS++))
    fi
    ((TOTAL_TESTS++))
    
    # 测试查询设备详情
    if test_get_device; then
        ((PASSED_TESTS++))
    else
        ((FAILED_TESTS++))
    fi
    ((TOTAL_TESTS++))
    
    # 测试查询设备详情 - 设备不存在
    if test_get_device_not_found; then
        ((PASSED_TESTS++))
    else
        ((FAILED_TESTS++))
    fi
    ((TOTAL_TESTS++))
    
    # 测试更新设备
    if test_update_device; then
        ((PASSED_TESTS++))
    else
        ((FAILED_TESTS++))
    fi
    ((TOTAL_TESTS++))
    
    # 测试更新设备 - 绑定验证
    if test_update_device_binding_validation; then
        ((PASSED_TESTS++))
    else
        ((FAILED_TESTS++))
    fi
    ((TOTAL_TESTS++))
    
    # 测试删除设备（注意：删除后设备将不可用）
    echo -e "\n${YELLOW}警告: 删除设备测试将禁用设备，后续测试可能失败${NC}"
    read -p "是否继续删除设备测试? (y/N): " confirm
    if [ "${confirm}" = "y" ] || [ "${confirm}" = "Y" ]; then
        if test_delete_device; then
            ((PASSED_TESTS++))
        else
            ((FAILED_TESTS++))
        fi
        ((TOTAL_TESTS++))
    else
        echo "跳过删除设备测试"
    fi
    
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

