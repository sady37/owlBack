-- 查询 demo 租户 admin 用户的 PIN hash
-- 使用方法: psql -U your_user -d your_database -f query_admin_pin_hash.sql

-- 1. 查找 demo 租户中 admin 用户的 PIN hash
SELECT 
    u.user_id::text as user_id,
    u.user_account,
    u.nickname,
    t.tenant_name,
    CASE 
        WHEN u.pin_hash IS NULL THEN 'NULL (未设置 PIN)'
        ELSE encode(u.pin_hash, 'hex')
    END as pin_hash_hex,
    CASE 
        WHEN u.pin_hash IS NULL THEN 0
        ELSE length(u.pin_hash)
    END as pin_hash_length_bytes,
    u.pin_hash as pin_hash_raw
FROM users u
JOIN tenants t ON u.tenant_id = t.tenant_id
WHERE t.tenant_name = 'demo' 
  AND u.user_account = 'admin';

-- 2. 计算 PIN 1212 和 1234 的 SHA256 hash (用于对比)
-- 注意: PostgreSQL 的 digest 函数返回的是二进制，需要转换为 hex
SELECT 
    '1212' as pin_value,
    encode(digest('1212', 'sha256'), 'hex') as pin_hash_hex_1212,
    length(digest('1212', 'sha256')) as hash_length_bytes;

SELECT 
    '1234' as pin_value,
    encode(digest('1234', 'sha256'), 'hex') as pin_hash_hex_1234,
    length(digest('1234', 'sha256')) as hash_length_bytes;

-- 3. 对比: 检查数据库中存储的 PIN hash 是否匹配 1212 或 1234
SELECT 
    u.user_id::text as user_id,
    u.user_account,
    encode(u.pin_hash, 'hex') as stored_pin_hash_hex,
    encode(digest('1212', 'sha256'), 'hex') as expected_hash_1212,
    encode(digest('1234', 'sha256'), 'hex') as expected_hash_1234,
    CASE 
        WHEN u.pin_hash = digest('1212', 'sha256') THEN 'MATCHES 1212 ✓'
        WHEN u.pin_hash = digest('1234', 'sha256') THEN 'MATCHES 1234 ✓'
        ELSE 'NO MATCH ✗'
    END as match_result
FROM users u
JOIN tenants t ON u.tenant_id = t.tenant_id
WHERE t.tenant_name = 'demo' 
  AND u.user_account = 'admin'
  AND u.pin_hash IS NOT NULL;

