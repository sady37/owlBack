-- 诊断设备卡片创建问题
-- 设备: Sleepad BM8701-2, Serial: BM87224700978, Device ID: 8amzqonkfyfyy
-- 绑在: DV-A-E102

-- 1. 检查设备基本信息
SELECT 
    d.device_id,
    d.device_name,
    d.device_uid,
    d.tenant_id,
    d.unit_id,
    d.bound_room_id,
    d.bound_bed_id,
    d.monitoring_enabled,
    d.business_access,
    d.status,
    ds.device_type,
    ds.device_model,
    u.unit_name,
    r.room_name,
    b.bed_name
FROM devices d
JOIN device_store ds ON d.device_id = ds.device_id
LEFT JOIN units u ON d.unit_id = u.unit_id AND d.tenant_id = u.tenant_id
LEFT JOIN rooms r ON d.bound_room_id = r.room_id AND d.tenant_id = r.tenant_id
LEFT JOIN beds b ON d.bound_bed_id = b.bed_id AND d.tenant_id = b.tenant_id
WHERE d.device_uid = '8amzqonkfyfyy'
   OR d.device_name LIKE '%BM87224700978%'
   OR ds.device_code = 'BM87224700978';

-- 2. 检查设备是否满足 ActiveBed 条件
-- ActiveBed 需要: bound_bed_id IS NOT NULL AND monitoring_enabled = TRUE AND status <> 'disabled'
SELECT 
    'ActiveBed 条件检查' as check_type,
    CASE 
        WHEN d.bound_bed_id IS NULL THEN '❌ 设备未绑床 (bound_bed_id IS NULL)'
        WHEN d.monitoring_enabled = FALSE THEN '❌ 设备监护未启用 (monitoring_enabled = FALSE)'
        WHEN d.status = 'disabled' THEN '❌ 设备状态为 disabled'
        ELSE '✅ 满足 ActiveBed 条件'
    END as result
FROM devices d
WHERE d.device_uid = '8amzqonkfyfyy'
   OR d.device_name LIKE '%BM87224700978%';

-- 3. 检查设备是否满足 UnitCard 条件
-- UnitCard 需要: bound_room_id IS NOT NULL AND bound_bed_id IS NULL AND monitoring_enabled = TRUE
SELECT 
    'UnitCard 条件检查' as check_type,
    CASE 
        WHEN d.bound_room_id IS NULL THEN '❌ 设备未绑房间 (bound_room_id IS NULL) - 无法关联到 unit'
        WHEN d.bound_bed_id IS NOT NULL THEN '❌ 设备已绑床 (bound_bed_id IS NOT NULL) - 应创建 ActiveBed 卡片'
        WHEN d.monitoring_enabled = FALSE THEN '❌ 设备监护未启用 (monitoring_enabled = FALSE)'
        ELSE '✅ 满足 UnitCard 条件'
    END as result
FROM devices d
WHERE d.device_uid = '8amzqonkfyfyy'
   OR d.device_name LIKE '%BM87224700978%';

-- 4. 检查 Unit E102 下的 ActiveBed 数量
SELECT 
    u.unit_id,
    u.unit_name,
    COUNT(DISTINCT b.bed_id) as active_bed_count
FROM units u
INNER JOIN rooms r ON u.unit_id = r.unit_id AND u.tenant_id = r.tenant_id
INNER JOIN beds b ON r.room_id = b.room_id AND r.tenant_id = b.tenant_id
INNER JOIN devices d ON d.bound_bed_id = b.bed_id AND d.tenant_id = b.tenant_id
WHERE u.unit_name = 'E102'
  AND d.monitoring_enabled = TRUE
  AND d.status <> 'disabled'
GROUP BY u.unit_id, u.unit_name;

-- 5. 检查 Unit E102 下未绑床的设备
SELECT 
    d.device_id,
    d.device_name,
    d.bound_room_id,
    d.bound_bed_id,
    d.monitoring_enabled,
    d.status,
    r.room_name,
    u.unit_name
FROM devices d
LEFT JOIN rooms r ON d.bound_room_id = r.room_id AND d.tenant_id = r.tenant_id
LEFT JOIN units u ON r.unit_id = u.unit_id AND r.tenant_id = u.tenant_id
WHERE u.unit_name = 'E102'
  AND d.bound_bed_id IS NULL
  AND d.bound_room_id IS NOT NULL
  AND d.monitoring_enabled = TRUE;

-- 6. 检查设备是否在 devices 表中存在
SELECT 
    '设备存在性检查' as check_type,
    CASE 
        WHEN COUNT(*) = 0 THEN '❌ 设备不存在于 devices 表'
        ELSE '✅ 设备存在于 devices 表'
    END as result
FROM devices d
WHERE d.device_uid = '8amzqonkfyfyy'
   OR d.device_name LIKE '%BM87224700978%';
