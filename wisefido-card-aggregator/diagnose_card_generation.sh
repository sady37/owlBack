#!/bin/bash

# Card 生成诊断脚本

set -e

TENANT_ID="${TENANT_ID:-bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c}"

echo "=== Card 生成诊断 ==="
echo "Tenant ID: $TENANT_ID"
echo ""

# 1. 检查 units
echo "1. Checking units..."
docker exec -i owl-postgresql psql -U postgres -d owlrd -c "
SELECT 
    unit_id,
    unit_name,
    unit_type,
    is_public,
    is_shared_unit
FROM units
WHERE tenant_id = '$TENANT_ID'
ORDER BY unit_name;
"

# 2. 检查 beds 和设备绑定
echo ""
echo "2. Checking beds with devices..."
docker exec -i owl-postgresql psql -U postgres -d owlrd -c "
SELECT 
    b.bed_id,
    b.bed_name,
    r.room_name,
    u.unit_name,
    COUNT(DISTINCT d.device_id) as device_count,
    COUNT(DISTINCT CASE WHEN d.monitoring_enabled = TRUE AND d.status <> 'disabled' THEN d.device_id END) as active_device_count
FROM beds b
INNER JOIN rooms r ON b.room_id = r.room_id
INNER JOIN units u ON r.unit_id = u.unit_id
LEFT JOIN devices d ON d.bound_bed_id = b.bed_id
WHERE b.tenant_id = '$TENANT_ID'
GROUP BY b.bed_id, b.bed_name, r.room_name, u.unit_name
HAVING COUNT(DISTINCT d.device_id) > 0
ORDER BY u.unit_name, b.bed_name;
"

# 3. 检查 ActiveBeds（符合 card 生成条件）
echo ""
echo "3. Checking ActiveBeds (符合 card 生成条件)..."
docker exec -i owl-postgresql psql -U postgres -d owlrd -c "
SELECT DISTINCT
    b.bed_id,
    b.bed_name,
    r.unit_id,
    u.unit_name,
    COUNT(DISTINCT d.device_id)::int AS bound_device_count,
    r2.resident_id
FROM beds b
INNER JOIN rooms r ON b.room_id = r.room_id
INNER JOIN units u ON r.unit_id = u.unit_id
INNER JOIN devices d ON d.bound_bed_id = b.bed_id
LEFT JOIN residents r2 ON r2.bed_id = b.bed_id AND r2.tenant_id = '$TENANT_ID'
WHERE b.tenant_id = '$TENANT_ID'
  AND d.monitoring_enabled = TRUE
  AND d.status <> 'disabled'
GROUP BY b.bed_id, b.bed_name, r.unit_id, u.unit_name, r2.resident_id
HAVING COUNT(DISTINCT d.device_id) > 0
ORDER BY u.unit_name, b.bed_name;
"

# 4. 检查现有 cards
echo ""
echo "4. Checking existing cards..."
docker exec -i owl-postgresql psql -U postgres -d owlrd -c "
SELECT 
    c.card_id,
    c.card_type,
    c.card_name,
    c.bed_id,
    c.unit_id,
    u.unit_name,
    b.bed_name
FROM cards c
LEFT JOIN units u ON c.unit_id = u.unit_id
LEFT JOIN beds b ON c.bed_id = b.bed_id
WHERE c.tenant_id = '$TENANT_ID'
ORDER BY u.unit_name, b.bed_name;
"

echo ""
echo "=== 诊断完成 ==="
echo ""
echo "如果 ActiveBeds 有数据但 cards 为空，请检查："
echo "1. 服务是否正在运行：ps aux | grep wisefido-card-aggregator"
echo "2. 服务日志是否有错误"
echo "3. TENANT_ID 环境变量是否正确设置"

