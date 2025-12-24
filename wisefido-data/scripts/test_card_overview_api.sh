#!/bin/bash

# Card Overview API 测试脚本
# 用于前端联调测试

BASE_URL="${BASE_URL:-http://localhost:8080}"
TENANT_ID="${TENANT_ID:-00000000-0000-0000-0000-000000000001}"
USER_ID="${USER_ID:-00000000-0000-0000-0000-000000000002}"
USER_TYPE="${USER_TYPE:-resident}"

echo "=========================================="
echo "Card Overview API 测试"
echo "=========================================="
echo "Base URL: $BASE_URL"
echo "Tenant ID: $TENANT_ID"
echo "User ID: $USER_ID"
echo "User Type: $USER_TYPE"
echo ""

# 测试 1: 基本查询
echo "测试 1: 基本查询"
echo "----------------------------------------"
curl -s -X GET "$BASE_URL/admin/api/v1/card-overview?tenant_id=$TENANT_ID" \
  -H "X-User-Id: $USER_ID" \
  -H "X-User-Type: $USER_TYPE" \
  -H "Content-Type: application/json" | jq '.' || echo "请求失败或响应不是有效 JSON"
echo ""
echo ""

# 测试 2: 搜索查询
echo "测试 2: 搜索查询 (search=Test)"
echo "----------------------------------------"
curl -s -X GET "$BASE_URL/admin/api/v1/card-overview?tenant_id=$TENANT_ID&search=Test" \
  -H "X-User-Id: $USER_ID" \
  -H "X-User-Type: $USER_TYPE" \
  -H "Content-Type: application/json" | jq '.' || echo "请求失败或响应不是有效 JSON"
echo ""
echo ""

# 测试 3: 按卡片类型过滤
echo "测试 3: 按卡片类型过滤 (card_type=ActiveBed)"
echo "----------------------------------------"
curl -s -X GET "$BASE_URL/admin/api/v1/card-overview?tenant_id=$TENANT_ID&card_type=ActiveBed" \
  -H "X-User-Id: $USER_ID" \
  -H "X-User-Type: $USER_TYPE" \
  -H "Content-Type: application/json" | jq '.' || echo "请求失败或响应不是有效 JSON"
echo ""
echo ""

# 测试 4: 按单元类型过滤
echo "测试 4: 按单元类型过滤 (unit_type=Facility)"
echo "----------------------------------------"
curl -s -X GET "$BASE_URL/admin/api/v1/card-overview?tenant_id=$TENANT_ID&unit_type=Facility" \
  -H "X-User-Id: $USER_ID" \
  -H "X-User-Type: $USER_TYPE" \
  -H "Content-Type: application/json" | jq '.' || echo "请求失败或响应不是有效 JSON"
echo ""
echo ""

# 测试 5: 组合查询（搜索 + 过滤）
echo "测试 5: 组合查询 (search=Test&card_type=ActiveBed)"
echo "----------------------------------------"
curl -s -X GET "$BASE_URL/admin/api/v1/card-overview?tenant_id=$TENANT_ID&search=Test&card_type=ActiveBed" \
  -H "X-User-Id: $USER_ID" \
  -H "X-User-Type: $USER_TYPE" \
  -H "Content-Type: application/json" | jq '.' || echo "请求失败或响应不是有效 JSON"
echo ""
echo ""

# 测试 6: 排序
echo "测试 6: 排序 (sort=card_name&direction=desc)"
echo "----------------------------------------"
curl -s -X GET "$BASE_URL/admin/api/v1/card-overview?tenant_id=$TENANT_ID&sort=card_name&direction=desc" \
  -H "X-User-Id: $USER_ID" \
  -H "X-User-Type: $USER_TYPE" \
  -H "Content-Type: application/json" | jq '.' || echo "请求失败或响应不是有效 JSON"
echo ""
echo ""

# 测试 7: Staff 用户（Caregiver）
if [ "$USER_TYPE" = "staff" ]; then
  echo "测试 7: Staff 用户查询 (AssignedOnly)"
  echo "----------------------------------------"
  curl -s -X GET "$BASE_URL/admin/api/v1/card-overview?tenant_id=$TENANT_ID" \
    -H "X-User-Id: $USER_ID" \
    -H "X-User-Type: staff" \
    -H "X-User-Role: Caregiver" \
    -H "Content-Type: application/json" | jq '.' || echo "请求失败或响应不是有效 JSON"
  echo ""
  echo ""
fi

echo "=========================================="
echo "测试完成"
echo "=========================================="

