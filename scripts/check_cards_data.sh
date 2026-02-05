#!/bin/bash
# 检查 cards 表是否有数据，以及 card-overview API 是否返回数据
# 用法: ./scripts/check_cards_data.sh [tenant_id]
# 默认 tenant_id 使用 card-manage 日志中的 bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c

set -e
TENANT_ID="${1:-bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c}"
DB_HOST="${DB_HOST:-127.0.0.1}"
DB_PORT="${DB_PORT:-5433}"

echo "=== 1. DB: cards 表条数 (tenant_id=$TENANT_ID) ==="
docker exec -i owl-postgresql psql -U postgres -d owlrd -t -c \
  "SELECT COUNT(*) FROM cards WHERE tenant_id = '$TENANT_ID';" 2>/dev/null || \
  echo "  (若未用 docker，请设置 DB_HOST/DB_PORT 并用 psql 执行上述 SQL)"

echo ""
echo "=== 2. DB: 各 unit 下卡片数 ==="
docker exec -i owl-postgresql psql -U postgres -d owlrd -c \
  "SELECT unit_id, COUNT(*) AS card_count FROM cards WHERE tenant_id = '$TENANT_ID' GROUP BY unit_id ORDER BY unit_id;" 2>/dev/null || true

echo ""
echo "=== 3. API: card-overview 返回 (需 wisefido-data 运行且带 X-User-* header) ==="
BASE_URL="${BASE_URL:-http://localhost:8080}"
# 无 header 时可能被鉴权拦截或走默认逻辑
curl -s -X GET "$BASE_URL/admin/api/v1/card-overview?tenant_id=$TENANT_ID" \
  -H "X-Tenant-Id: $TENANT_ID" \
  -H "X-User-Id: e0f23dda-ee49-4915-9d53-d23b2e5e045a" \
  -H "X-User-Type: staff" \
  -H "X-User-Role: Admin" | head -c 500
echo ""
echo "  (若返回 401/403 或非 JSON，请先登录拿 token 或检查 wisefido-data 是否 8080)"
