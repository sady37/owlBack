#!/bin/bash

# 测试账户设置更新功能
# 测试数据: user=S1, resident=r1, contact=r1a

BASE_URL="http://localhost:8080"
TENANT_ID="00000000-0000-0000-0000-000000000001"  # 默认 tenant ID

# 如果环境变量中有设置，使用环境变量
if [ -n "$TEST_TENANT_ID" ]; then
    TENANT_ID="$TEST_TENANT_ID"
fi

echo "=========================================="
echo "测试账户设置更新功能"
echo "=========================================="
echo ""

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试函数
test_api() {
    local method=$1
    local url=$2
    local data=$3
    local description=$4
    local user_type=${5:-""}  # 可选：resident 或 family
    local user_role=${6:-""}  # 可选：Resident 或 Family
    
    echo -e "${YELLOW}测试: ${description}${NC}"
    echo "URL: $method $url"
    if [ -n "$data" ]; then
        echo "Data: $data"
    fi
    
    # 构建 curl 命令
    local curl_cmd="curl -s -w \"\\n%{http_code}\" -X \"$method\" \"$url\""
    curl_cmd="$curl_cmd -H \"Content-Type: application/json\""
    curl_cmd="$curl_cmd -H \"X-User-Id: $CURRENT_USER_ID\""
    curl_cmd="$curl_cmd -H \"X-Tenant-Id: $TENANT_ID\""
    
    # 如果是 resident 或 contact，添加额外的 headers
    if [ -n "$user_type" ]; then
        curl_cmd="$curl_cmd -H \"X-User-Type: $user_type\""
    fi
    if [ -n "$user_role" ]; then
        curl_cmd="$curl_cmd -H \"X-User-Role: $user_role\""
    fi
    
    if [ -n "$data" ]; then
        curl_cmd="$curl_cmd -d '$data'"
    fi
    
    response=$(eval $curl_cmd)
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" = "200" ]; then
        echo -e "${GREEN}✓ 成功 (HTTP $http_code)${NC}"
        echo "Response: $body" | jq '.' 2>/dev/null || echo "Response: $body"
    else
        echo -e "${RED}✗ 失败 (HTTP $http_code)${NC}"
        echo "Response: $body"
    fi
    echo ""
}

# 计算 hash 的辅助函数（使用 sha256sum）
hash_string() {
    echo -n "$1" | sha256sum | cut -d' ' -f1
}

# 测试 User S1
USER_S1_ID="7f3fa6ff-3a07-49b2-a649-6ea086e93863"
echo "=========================================="
echo "1. 测试 User (S1) 账户设置更新"
echo "User ID: $USER_S1_ID"
echo "=========================================="
CURRENT_USER_ID="$USER_S1_ID"

# 1.1 获取当前账户设置
test_api "GET" "$BASE_URL/admin/api/v1/users/$USER_S1_ID/account-settings" "" "获取 User S1 账户设置"

# 1.2 更新 email 和 phone（正常值）
EMAIL="test@example.com"
PHONE="1234567890"
EMAIL_HASH=$(hash_string "$EMAIL")
PHONE_HASH=$(hash_string "$PHONE")

test_api "PUT" "$BASE_URL/admin/api/v1/users/$USER_S1_ID/account-settings" \
    "{\"email\":\"$EMAIL\",\"phone\":\"$PHONE\",\"email_hash\":\"$EMAIL_HASH\",\"phone_hash\":\"$PHONE_HASH\"}" \
    "更新 User S1 email 和 phone（正常值）"

# 1.3 验证更新结果
test_api "GET" "$BASE_URL/admin/api/v1/users/$USER_S1_ID/account-settings" "" "验证 User S1 更新结果"

# 1.4 更新 email 和 phone 为空字符串（应该删除字段）
test_api "PUT" "$BASE_URL/admin/api/v1/users/$USER_S1_ID/account-settings" \
    "{\"email\":\"\",\"phone\":\"\",\"email_hash\":\"\",\"phone_hash\":\"\"}" \
    "更新 User S1 email 和 phone 为空字符串（应该删除字段）"

# 1.5 验证删除结果
test_api "GET" "$BASE_URL/admin/api/v1/users/$USER_S1_ID/account-settings" "" "验证 User S1 删除结果"

echo ""
echo "=========================================="
echo "2. 测试 Resident (r1) 账户设置更新"
RESIDENT_R1_ID="85ddc499-eea5-4b92-b625-546819a841e7"
echo "Resident ID: $RESIDENT_R1_ID"
echo "=========================================="
CURRENT_USER_ID="$RESIDENT_R1_ID"

# 2.1 获取当前账户设置
test_api "GET" "$BASE_URL/admin/api/v1/residents/$RESIDENT_R1_ID/account-settings" "" "获取 Resident r1 账户设置" "resident" "Resident"

# 2.2 更新 email 和 phone（正常值）
EMAIL2="resident@example.com"
PHONE2="9876543210"
EMAIL_HASH2=$(hash_string "$EMAIL2")
PHONE_HASH2=$(hash_string "$PHONE2")

test_api "PUT" "$BASE_URL/admin/api/v1/residents/$RESIDENT_R1_ID/account-settings" \
    "{\"email\":\"$EMAIL2\",\"phone\":\"$PHONE2\",\"email_hash\":\"$EMAIL_HASH2\",\"phone_hash\":\"$PHONE_HASH2\"}" \
    "更新 Resident r1 email 和 phone（正常值）" "resident" "Resident"

# 2.3 验证更新结果
test_api "GET" "$BASE_URL/admin/api/v1/residents/$RESIDENT_R1_ID/account-settings" "" "验证 Resident r1 更新结果" "resident" "Resident"

# 2.4 更新 email 和 phone 为空字符串（应该删除字段）
test_api "PUT" "$BASE_URL/admin/api/v1/residents/$RESIDENT_R1_ID/account-settings" \
    "{\"email\":\"\",\"phone\":\"\",\"email_hash\":\"\",\"phone_hash\":\"\"}" \
    "更新 Resident r1 email 和 phone 为空字符串（应该删除字段）" "resident" "Resident"

# 2.5 验证删除结果
test_api "GET" "$BASE_URL/admin/api/v1/residents/$RESIDENT_R1_ID/account-settings" "" "验证 Resident r1 删除结果" "resident" "Resident"

echo ""
echo "=========================================="
echo "3. 测试 Contact (r1a) 账户设置更新"
CONTACT_R1A_ID="c26f3e4d-2eb6-497e-8d72-b24291b98307"
echo "Contact ID: $CONTACT_R1A_ID"
echo "=========================================="
CURRENT_USER_ID="$CONTACT_R1A_ID"

# 3.1 获取当前账户设置
test_api "GET" "$BASE_URL/admin/api/v1/residents/$CONTACT_R1A_ID/account-settings" "" "获取 Contact r1a 账户设置" "resident" "Family"

# 3.2 更新 email 和 phone（正常值）
EMAIL3="contact@example.com"
PHONE3="5555555555"
EMAIL_HASH3=$(hash_string "$EMAIL3")
PHONE_HASH3=$(hash_string "$PHONE3")

test_api "PUT" "$BASE_URL/admin/api/v1/residents/$CONTACT_R1A_ID/account-settings" \
    "{\"email\":\"$EMAIL3\",\"phone\":\"$PHONE3\",\"email_hash\":\"$EMAIL_HASH3\",\"phone_hash\":\"$PHONE_HASH3\"}" \
    "更新 Contact r1a email 和 phone（正常值）" "resident" "Family"

# 3.3 验证更新结果
test_api "GET" "$BASE_URL/admin/api/v1/residents/$CONTACT_R1A_ID/account-settings" "" "验证 Contact r1a 更新结果" "resident" "Family"

# 3.4 更新 email 和 phone 为空字符串（应该删除字段）
test_api "PUT" "$BASE_URL/admin/api/v1/residents/$CONTACT_R1A_ID/account-settings" \
    "{\"email\":\"\",\"phone\":\"\",\"email_hash\":\"\",\"phone_hash\":\"\"}" \
    "更新 Contact r1a email 和 phone 为空字符串（应该删除字段）" "resident" "Family"

# 3.5 验证删除结果
test_api "GET" "$BASE_URL/admin/api/v1/residents/$CONTACT_R1A_ID/account-settings" "" "验证 Contact r1a 删除结果" "resident" "Family"

echo ""
echo "=========================================="
echo "测试完成"
echo "=========================================="

