-- 验证账户设置更新结果
-- 测试数据: user=S1, resident=r1, contact=r1a

-- 1. 查看 User S1 的 email 和 phone 相关字段
SELECT 
    'User S1' as type,
    user_id,
    email,
    phone,
    CASE WHEN email_hash IS NULL THEN 'NULL' ELSE 'NOT NULL' END as email_hash_status,
    CASE WHEN phone_hash IS NULL THEN 'NULL' ELSE 'NOT NULL' END as phone_hash_status
FROM users
WHERE user_id::text = 'S1';

-- 2. 查看 Resident r1 的 email 和 phone 相关字段
SELECT 
    'Resident r1' as type,
    resident_id,
    CASE WHEN email_hash IS NULL THEN 'NULL' ELSE 'NOT NULL' END as email_hash_status,
    CASE WHEN phone_hash IS NULL THEN 'NULL' ELSE 'NOT NULL' END as phone_hash_status
FROM residents
WHERE resident_id::text = 'r1';

-- 3. 查看 Resident r1 的 PHI 表中的 email 和 phone
SELECT 
    'Resident r1 PHI' as type,
    resident_id,
    resident_email,
    resident_phone
FROM resident_phi
WHERE resident_id::text = 'r1';

-- 4. 查看 Contact r1a 的 email 和 phone 相关字段
SELECT 
    'Contact r1a' as type,
    contact_id,
    contact_email,
    contact_phone,
    CASE WHEN email_hash IS NULL THEN 'NULL' ELSE 'NOT NULL' END as email_hash_status,
    CASE WHEN phone_hash IS NULL THEN 'NULL' ELSE 'NOT NULL' END as phone_hash_status
FROM resident_contacts
WHERE contact_id::text = 'r1a';

-- 5. 验证空字符串是否被正确设置为 NULL
-- 如果字段值为空字符串 ''，应该显示为 NULL
SELECT 
    'Verification' as type,
    'User S1 email' as field,
    CASE 
        WHEN email IS NULL THEN 'NULL (correct)'
        WHEN email = '' THEN 'Empty string (should be NULL)'
        ELSE 'Has value: ' || email
    END as status
FROM users
WHERE user_id::text = 'S1'
UNION ALL
SELECT 
    'Verification' as type,
    'User S1 phone' as field,
    CASE 
        WHEN phone IS NULL THEN 'NULL (correct)'
        WHEN phone = '' THEN 'Empty string (should be NULL)'
        ELSE 'Has value: ' || phone
    END as status
FROM users
WHERE user_id::text = 'S1'
UNION ALL
SELECT 
    'Verification' as type,
    'Contact r1a email' as field,
    CASE 
        WHEN contact_email IS NULL THEN 'NULL (correct)'
        WHEN contact_email = '' THEN 'Empty string (should be NULL)'
        ELSE 'Has value: ' || contact_email
    END as status
FROM resident_contacts
WHERE contact_id::text = 'r1a'
UNION ALL
SELECT 
    'Verification' as type,
    'Contact r1a phone' as field,
    CASE 
        WHEN contact_phone IS NULL THEN 'NULL (correct)'
        WHEN contact_phone = '' THEN 'Empty string (should be NULL)'
        ELSE 'Has value: ' || contact_phone
    END as status
FROM resident_contacts
WHERE contact_id::text = 'r1a';

