-- 查找测试数据的实际 UUID
-- 根据 user_account, resident_account, contact slot 查找

-- 1. 查找 User S1
SELECT 
    'User S1' as type,
    user_id::text as id,
    user_account,
    email,
    phone
FROM users
WHERE user_account = 's1' OR user_account ILIKE '%s1%'
LIMIT 5;

-- 2. 查找 Resident r1
SELECT 
    'Resident r1' as type,
    resident_id::text as id,
    resident_account,
    nickname
FROM residents
WHERE resident_account = 'r1' OR resident_account ILIKE '%r1%'
LIMIT 5;

-- 3. 查找 Contact r1a (slot A)
SELECT 
    'Contact r1a' as type,
    contact_id::text as id,
    slot,
    contact_first_name,
    contact_last_name
FROM resident_contacts
WHERE slot = 'A'
LIMIT 5;

-- 4. 如果知道 resident_id，查找其 contacts
-- 替换下面的 'RESIDENT_UUID_HERE' 为实际的 resident_id
-- SELECT 
--     'Contact for resident' as type,
--     contact_id::text as id,
--     slot,
--     contact_first_name,
--     contact_last_name
-- FROM resident_contacts
-- WHERE resident_id::text = 'RESIDENT_UUID_HERE';

