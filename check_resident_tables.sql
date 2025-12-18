-- 检查 residents, resident_phi, resident_contacts 三张表的数据
-- 用于排查：Done修改、allow access、phi丢失、contact保存问题、登录问题

-- 1. 检查 residents 表（包含 can_view_status, status, password_hash）
SELECT 
    r.resident_id,
    r.tenant_id,
    r.resident_account,
    r.nickname,
    r.status,
    r.can_view_status,
    r.service_level,
    r.family_tag,
    CASE WHEN r.password_hash IS NULL THEN 'NULL' ELSE 'HAS_PASSWORD' END as password_hash_status,
    CASE WHEN r.phone_hash IS NULL THEN 'NULL' ELSE 'HAS_PHONE_HASH' END as phone_hash_status,
    CASE WHEN r.email_hash IS NULL THEN 'NULL' ELSE 'HAS_EMAIL_HASH' END as email_hash_status,
    r.admission_date,
    r.discharge_date
FROM residents r
ORDER BY r.tenant_id, r.nickname
LIMIT 20;

-- 2. 检查 resident_phi 表（检查是否丢失数据）
SELECT 
    rp.phi_id,
    rp.tenant_id,
    rp.resident_id,
    r.nickname,
    rp.first_name,
    rp.last_name,
    rp.gender,
    rp.date_of_birth,
    CASE WHEN rp.resident_phone IS NULL THEN 'NULL' ELSE 'HAS_PHONE' END as resident_phone_status,
    CASE WHEN rp.resident_email IS NULL THEN 'NULL' ELSE 'HAS_EMAIL' END as resident_email_status,
    rp.weight_lb,
    rp.height_ft,
    rp.height_in,
    rp.mobility_level,
    rp.has_hypertension,
    rp.has_hyperlipaemia,
    rp.has_hyperglycaemia,
    rp.has_stroke_history,
    rp.has_paralysis,
    rp.has_alzheimer
FROM resident_phi rp
LEFT JOIN residents r ON rp.resident_id = r.resident_id
ORDER BY rp.tenant_id, r.nickname
LIMIT 20;

-- 3. 检查 resident_contacts 表（检查 slot A 和 B 的 email 保存情况）
SELECT 
    rc.contact_id,
    rc.tenant_id,
    rc.resident_id,
    r.nickname,
    rc.slot,
    rc.is_enabled,
    rc.relationship,
    rc.contact_family_tag,
    -- 检查 email 明文保存情况
    CASE WHEN rc.contact_email IS NULL THEN 'NULL' ELSE rc.contact_email END as contact_email,
    CASE WHEN rc.contact_phone IS NULL THEN 'NULL' ELSE rc.contact_phone END as contact_phone,
    -- 检查 hash 保存情况
    CASE WHEN rc.email_hash IS NULL THEN 'NULL' ELSE 'HAS_EMAIL_HASH' END as email_hash_status,
    CASE WHEN rc.phone_hash IS NULL THEN 'NULL' ELSE 'HAS_PHONE_HASH' END as phone_hash_status,
    -- 检查 password_hash
    CASE WHEN rc.password_hash IS NULL THEN 'NULL' ELSE 'HAS_PASSWORD_HASH' END as password_hash_status,
    rc.receive_sms,
    rc.receive_email
FROM resident_contacts rc
LEFT JOIN residents r ON rc.resident_id = r.resident_id
WHERE rc.slot IN ('A', 'B')
ORDER BY rc.tenant_id, r.nickname, rc.slot;

-- 4. 检查特定邮箱的登录信息（ding@gmail.com 和 ding2@gmail.com）
-- 需要计算 hash 来查找，这里先查找包含这些邮箱的联系人
SELECT 
    rc.contact_id,
    rc.tenant_id,
    rc.resident_id,
    r.nickname,
    rc.slot,
    rc.contact_email,
    CASE WHEN rc.email_hash IS NULL THEN 'NULL' ELSE 'HAS_EMAIL_HASH' END as email_hash_status,
    CASE WHEN rc.password_hash IS NULL THEN 'NULL' ELSE 'HAS_PASSWORD_HASH' END as password_hash_status,
    rc.is_enabled
FROM resident_contacts rc
LEFT JOIN residents r ON rc.resident_id = r.resident_id
WHERE rc.contact_email IN ('ding@gmail.com', 'ding2@gmail.com')
   OR LOWER(rc.contact_email) LIKE '%ding%'
ORDER BY rc.contact_email, rc.slot;

-- 5. 检查 residents 表中是否有这些邮箱的 hash（用于登录）
-- 注意：需要手动计算 hash 来查找，这里只显示有 email_hash 的记录
SELECT 
    r.resident_id,
    r.tenant_id,
    r.resident_account,
    r.nickname,
    CASE WHEN r.email_hash IS NULL THEN 'NULL' ELSE 'HAS_EMAIL_HASH' END as email_hash_status,
    CASE WHEN r.password_hash IS NULL THEN 'NULL' ELSE 'HAS_PASSWORD_HASH' END as password_hash_status,
    r.can_view_status,
    r.status
FROM residents r
WHERE r.email_hash IS NOT NULL
ORDER BY r.tenant_id, r.nickname;

-- 6. 统计信息
SELECT 
    'residents' as table_name,
    COUNT(*) as total_count,
    COUNT(CASE WHEN can_view_status = TRUE THEN 1 END) as can_view_status_true,
    COUNT(CASE WHEN can_view_status = FALSE THEN 1 END) as can_view_status_false,
    COUNT(CASE WHEN password_hash IS NOT NULL THEN 1 END) as has_password_hash,
    COUNT(CASE WHEN email_hash IS NOT NULL THEN 1 END) as has_email_hash,
    COUNT(CASE WHEN phone_hash IS NOT NULL THEN 1 END) as has_phone_hash
FROM residents
UNION ALL
SELECT 
    'resident_phi' as table_name,
    COUNT(*) as total_count,
    COUNT(CASE WHEN resident_phone IS NOT NULL THEN 1 END) as has_phone,
    COUNT(CASE WHEN resident_email IS NOT NULL THEN 1 END) as has_email,
    COUNT(CASE WHEN first_name IS NOT NULL THEN 1 END) as has_first_name,
    COUNT(CASE WHEN last_name IS NOT NULL THEN 1 END) as has_last_name,
    COUNT(CASE WHEN date_of_birth IS NOT NULL THEN 1 END) as has_dob
FROM resident_phi
UNION ALL
SELECT 
    'resident_contacts' as table_name,
    COUNT(*) as total_count,
    COUNT(CASE WHEN is_enabled = TRUE THEN 1 END) as is_enabled_true,
    COUNT(CASE WHEN contact_email IS NOT NULL THEN 1 END) as has_contact_email,
    COUNT(CASE WHEN email_hash IS NOT NULL THEN 1 END) as has_email_hash,
    COUNT(CASE WHEN password_hash IS NOT NULL THEN 1 END) as has_password_hash,
    COUNT(CASE WHEN slot = 'A' THEN 1 END) as slot_a_count
FROM resident_contacts;

