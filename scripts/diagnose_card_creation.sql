-- 诊断卡片创建问题：检查设备 Sleepad-yfyy 绑定到 Unit E102 时为什么没有创建卡片
-- 使用方法：替换 YOUR_TENANT_ID 等变量后执行；或直接运行 run_diagnose_cards.sh

-- 0. device_store 中的设备（含已删除：devices 无记录、device_store.tenant_id 为未分配/系统）
-- device_store 无 device_name，用 device_uid / device_code 匹配
SELECT 
    ds.device_id,
    ds.device_uid,
    ds.device_code,
    ds.device_type,
    ds.tenant_id AS ds_tenant_id,
    CASE 
        WHEN ds.tenant_id::text = '00000000-0000-0000-0000-000000000000' THEN 'unallocated'
        WHEN ds.tenant_id::text = '00000000-0000-0000-0000-000000000001' THEN 'system'
        ELSE 'tenant'
    END AS ds_tenant_type,
    EXISTS(SELECT 1 FROM devices d WHERE d.device_id = ds.device_id) AS in_devices
FROM device_store ds
WHERE ds.device_uid ILIKE '%Sleepad-yfyy%'
   OR ds.device_code ILIKE '%Sleepad-yfyy%'
ORDER BY ds.device_uid;

-- 1. 查找设备信息（devices 表，仅未删除时存在；unit 通过 room/bed 推导）
SELECT 
    d.device_id,
    d.device_name,
    d.device_uid,
    d.tenant_id,
    ds.device_type,
    ds.device_model,
    d.bound_room_id,
    d.bound_bed_id,
    d.monitoring_enabled,
    d.status,
    d.business_access,
    u.unit_id,
    u.unit_name,
    r.room_name,
    b.bed_name
FROM devices d
JOIN device_store ds ON d.device_id = ds.device_id
LEFT JOIN beds b ON d.bound_bed_id = b.bed_id AND d.tenant_id = b.tenant_id
LEFT JOIN rooms r ON (
    (d.bound_room_id IS NOT NULL AND r.room_id = d.bound_room_id) OR
    (d.bound_bed_id IS NOT NULL AND r.room_id = b.room_id)
) AND r.tenant_id = d.tenant_id
LEFT JOIN units u ON r.unit_id = u.unit_id AND r.tenant_id = u.tenant_id
WHERE d.device_name ILIKE '%Sleepad-yfyy%'
   OR d.device_uid ILIKE '%Sleepad-yfyy%'
ORDER BY d.device_name;

-- 2. 检查 Unit E102 下的 ActiveBed（需要满足：设备绑定到床 + monitoring_enabled = TRUE + status <> 'disabled'）
-- 替换 tenant_id 和 unit_id
SELECT 
    b.bed_id,
    b.bed_name,
    r.room_name,
    u.unit_name,
    COUNT(DISTINCT d.device_id) AS bound_device_count,
    STRING_AGG(DISTINCT d.device_name, ', ') AS device_names,
    STRING_AGG(DISTINCT d.device_uid, ', ') AS device_uids,
    BOOL_OR(d.monitoring_enabled) AS has_monitoring_enabled,
    STRING_AGG(DISTINCT d.status::text, ', ') AS device_statuses
FROM beds b
INNER JOIN rooms r ON b.room_id = r.room_id
INNER JOIN units u ON r.unit_id = u.unit_id
LEFT JOIN devices d ON d.bound_bed_id = b.bed_id AND d.tenant_id = b.tenant_id
WHERE b.tenant_id = 'YOUR_TENANT_ID'  -- 替换为实际 tenant_id
  AND u.unit_name = 'E102'            -- 或使用 unit_id
GROUP BY b.bed_id, b.bed_name, r.room_name, u.unit_name
HAVING COUNT(DISTINCT d.device_id) > 0
ORDER BY b.bed_name;

-- 3. 检查 Unit E102 下未绑床但已启用监控的设备（用于创建 UnitCard）
-- 替换 tenant_id 和 unit_id
SELECT 
    d.device_id,
    d.device_name,
    d.device_uid,
    ds.device_type,
    d.bound_room_id,
    d.bound_bed_id,
    d.monitoring_enabled,
    d.status,
    r.room_name,
    u.unit_name
FROM devices d
JOIN device_store ds ON d.device_id = ds.device_id
LEFT JOIN rooms r ON d.bound_room_id = r.room_id AND d.tenant_id = r.tenant_id
LEFT JOIN units u ON r.unit_id = u.unit_id AND d.tenant_id = u.tenant_id
WHERE d.tenant_id = 'YOUR_TENANT_ID'  -- 替换为实际 tenant_id
  AND u.unit_name = 'E102'            -- 或使用 unit_id
  AND d.bound_room_id IS NOT NULL
  AND d.bound_bed_id IS NULL
  AND d.monitoring_enabled = TRUE
ORDER BY d.device_name;

-- 4. 检查 Unit E102 的配置信息
SELECT 
    u.unit_id,
    u.unit_name,
    u.unit_type,
    u.is_public,
    u.is_shared_unit,
    br.branch_name,
    bld.building_name,
    COUNT(DISTINCT bd.bed_id) AS bed_count,
    COUNT(DISTINCT rv.resident_id) AS resident_count
FROM units u
LEFT JOIN branches br ON u.branch_id = br.branch_id
LEFT JOIN buildings bld ON u.building_id = bld.building_id
LEFT JOIN rooms rm ON rm.unit_id = u.unit_id AND rm.tenant_id = u.tenant_id
LEFT JOIN beds bd ON bd.room_id = rm.room_id AND bd.tenant_id = rm.tenant_id
LEFT JOIN residents rv ON rv.unit_id = u.unit_id AND rv.tenant_id = u.tenant_id
WHERE u.tenant_id = 'YOUR_TENANT_ID'  -- 替换为实际 tenant_id
  AND u.unit_name = 'E102'            -- 或使用 unit_id
GROUP BY u.unit_id, u.unit_name, u.unit_type, u.is_public, u.is_shared_unit, br.branch_name, bld.building_name;

-- 5. 检查已存在的卡片（如果有）
SELECT 
    c.card_id,
    c.card_type,
    c.card_name,
    c.card_address,
    c.bed_id,
    c.unit_id,
    c.resident_id,
    jsonb_array_length(c.devices) AS device_count
FROM cards c
WHERE c.tenant_id = 'YOUR_TENANT_ID'  -- 替换为实际 tenant_id
  AND c.unit_id IN (
      SELECT unit_id FROM units WHERE unit_name = 'E102' AND tenant_id = 'YOUR_TENANT_ID'
  )
ORDER BY c.card_type, c.card_name;

-- 6. Delete 时 device_store.tenant_id 检查（迁回 system / 未分配）
-- 代码约定：DeleteDevice 将 device_store.tenant_id 改为 未分配(UUID all-zero)，devices 行物理删除
-- 若期望迁回 System(...0001)，需改 postgres_devices.DeleteDevice
SELECT 
    ds.device_id,
    ds.device_uid,
    ds.device_code,
    ds.tenant_id::text,
    CASE ds.tenant_id::text
        WHEN '00000000-0000-0000-0000-000000000000' THEN 'unallocated(删后迁回)'
        WHEN '00000000-0000-0000-0000-000000000001' THEN 'system'
        ELSE 'tenant'
    END AS tenant_type
FROM device_store ds
WHERE (ds.device_uid ILIKE '%Sleepad-yfyy%' OR ds.device_code ILIKE '%Sleepad-yfyy%')
  AND ds.tenant_id::text IN (
      '00000000-0000-0000-0000-000000000000',
      '00000000-0000-0000-0000-000000000001'
  );

-- 常见问题排查：
-- 1. 如果设备 bound_bed_id IS NULL：设备只绑定到 unit，需要检查是否有 ActiveBed
--    - 如果没有 ActiveBed，且设备 monitoring_enabled = TRUE，应该创建 UnitCard
--    - 如果有 ActiveBed，未绑床的设备不会单独创建卡片（除非场景 B：多个 ActiveBed）
-- 2. 如果设备 monitoring_enabled = FALSE：不会创建卡片
-- 3. 如果设备 status = 'disabled'：不会创建卡片
-- 4. 如果设备 bound_room_id IS NULL：无法关联到 unit，不会创建卡片
-- 5. 删除(Delete)后：devices 行已删除；device_store.tenant_id 改为 0000..0000(未分配)。非 0000..0001(System)。
