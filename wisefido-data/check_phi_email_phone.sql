-- 检查 resident_phi 表中的 email 和 phone
-- r1 的 nickname 是 smith
-- r3 的 nickname 是 test1
SELECT 
    r.resident_account,
    r.nickname,
    r.email_hash IS NOT NULL as has_email_hash,
    r.phone_hash IS NOT NULL as has_phone_hash,
    encode(r.email_hash, 'hex') as email_hash_hex,
    encode(r.phone_hash, 'hex') as phone_hash_hex,
    rp.resident_email,
    rp.resident_phone,
    CASE 
        WHEN r.email_hash IS NOT NULL AND rp.resident_email IS NULL THEN 'Hash exists but email is NULL'
        WHEN r.email_hash IS NOT NULL AND rp.resident_email IS NOT NULL THEN 'Both hash and email exist'
        WHEN r.email_hash IS NULL AND rp.resident_email IS NULL THEN 'Both are NULL'
        ELSE 'Other'
    END as email_status,
    CASE 
        WHEN r.phone_hash IS NOT NULL AND rp.resident_phone IS NULL THEN 'Hash exists but phone is NULL'
        WHEN r.phone_hash IS NOT NULL AND rp.resident_phone IS NOT NULL THEN 'Both hash and phone exist'
        WHEN r.phone_hash IS NULL AND rp.resident_phone IS NULL THEN 'Both are NULL'
        ELSE 'Other'
    END as phone_status
FROM residents r
LEFT JOIN resident_phi rp ON r.tenant_id = rp.tenant_id AND r.resident_id = rp.resident_id
WHERE r.resident_account IN ('r1', 'r3') OR r.nickname IN ('smith', 'test1')
ORDER BY r.resident_account;
