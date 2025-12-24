#!/bin/bash

# 测试 User AccountSettings API
# 使用 demo/s1 用户

TENANT_ID="095c1ab6-5143-47ea-8670-5476158b6cad"
USER_ID="7f3fa6ff-3a07-49b2-a649-6ea086e93863"
EMAIL="s1@demo.com"
PASSWORD="Ts123@123"
BASE_URL="http://localhost:8080"

echo "=== 1. 登录获取 Token ==="
LOGIN_RESP=$(curl -s -X POST "${BASE_URL}/admin/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"account\": \"${EMAIL}\",
    \"password\": \"${PASSWORD}\",
    \"userType\": \"staff\",
    \"tenant_id\": \"${TENANT_ID}\"
  }")

echo "$LOGIN_RESP" | python3 -m json.tool 2>/dev/null || echo "$LOGIN_RESP"
echo ""

TOKEN=$(echo "$LOGIN_RESP" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
if [ -z "$TOKEN" ]; then
  echo "登录失败，无法获取 token"
  exit 1
fi

echo "Token: ${TOKEN:0:50}..."
echo ""

echo "=== 2. 测试 GetAccountSettings ==="
GET_RESP=$(curl -s -X GET "${BASE_URL}/admin/api/v1/users/${USER_ID}/account-settings" \
  -H "X-User-Id: ${USER_ID}" \
  -H "X-Tenant-Id: ${TENANT_ID}")

echo "$GET_RESP" | python3 -m json.tool 2>/dev/null || echo "$GET_RESP"
echo ""

echo "=== 3. 测试 UpdateAccountSettings (更新 phone) ==="
UPDATE_RESP=$(curl -s -X PUT "${BASE_URL}/admin/api/v1/users/${USER_ID}/account-settings" \
  -H "X-User-Id: ${USER_ID}" \
  -H "X-Tenant-Id: ${TENANT_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "1234567890",
    "phone_hash": "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3"
  }')

echo "$UPDATE_RESP" | python3 -m json.tool 2>/dev/null || echo "$UPDATE_RESP"
echo ""

echo "=== 4. 再次测试 GetAccountSettings (验证更新) ==="
GET_RESP2=$(curl -s -X GET "${BASE_URL}/admin/api/v1/users/${USER_ID}/account-settings" \
  -H "X-User-Id: ${USER_ID}" \
  -H "X-Tenant-Id: ${TENANT_ID}")

echo "$GET_RESP2" | python3 -m json.tool 2>/dev/null || echo "$GET_RESP2"
echo ""

echo "=== 5. 测试 UpdateAccountSettings (删除 phone 明文，保留 hash) ==="
UPDATE_RESP2=$(curl -s -X PUT "${BASE_URL}/admin/api/v1/users/${USER_ID}/account-settings" \
  -H "X-User-Id: ${USER_ID}" \
  -H "X-Tenant-Id: ${TENANT_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": null
  }')

echo "$UPDATE_RESP2" | python3 -m json.tool 2>/dev/null || echo "$UPDATE_RESP2"
echo ""

echo "=== 6. 验证 phone 已删除但 hash 保留 ==="
GET_RESP3=$(curl -s -X GET "${BASE_URL}/admin/api/v1/users/${USER_ID}/account-settings" \
  -H "X-User-Id: ${USER_ID}" \
  -H "X-Tenant-Id: ${TENANT_ID}")

echo "$GET_RESP3" | python3 -m json.tool 2>/dev/null || echo "$GET_RESP3"
echo ""

echo "=== 7. 验证数据库中的 phone_hash 是否保留 ==="
docker exec owl-postgresql psql -U postgres -d owlrd -c "SELECT user_id, user_account, phone, encode(phone_hash, 'hex') as phone_hash_hex FROM users WHERE user_id::text = '${USER_ID}';"

