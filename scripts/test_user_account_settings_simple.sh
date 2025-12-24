#!/bin/bash

# 简单的测试脚本 - 测试 User AccountSettings API
# 使用 demo/s1 用户

TENANT_ID="095c1ab6-5143-47ea-8670-5476158b6cad"
USER_ID="7f3fa6ff-3a07-49b2-a649-6ea086e93863"

echo "=== 1. 测试 GetAccountSettings ==="
curl -s -X GET "http://localhost:8080/admin/api/v1/users/${USER_ID}/account-settings" \
  -H "X-User-Id: ${USER_ID}" \
  -H "X-Tenant-Id: ${TENANT_ID}"
echo ""
echo ""

echo "=== 2. 测试 UpdateAccountSettings (更新 phone) ==="
curl -s -X PUT "http://localhost:8080/admin/api/v1/users/${USER_ID}/account-settings" \
  -H "X-User-Id: ${USER_ID}" \
  -H "X-Tenant-Id: ${TENANT_ID}" \
  -H "Content-Type: application/json" \
  -d '{"phone": "1234567890", "phone_hash": "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3"}'
echo ""
echo ""

echo "=== 3. 验证更新后的数据 ==="
curl -s -X GET "http://localhost:8080/admin/api/v1/users/${USER_ID}/account-settings" \
  -H "X-User-Id: ${USER_ID}" \
  -H "X-Tenant-Id: ${TENANT_ID}"
echo ""
echo ""

echo "=== 4. 数据库验证 ==="
docker exec owl-postgresql psql -U postgres -d owlrd -c "SELECT user_account, phone, encode(phone_hash, 'hex') as phone_hash_hex FROM users WHERE user_id::text = '${USER_ID}';"

