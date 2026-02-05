-- 查询 users 表指定 user_id
SELECT
    user_id::text,
    tenant_id::text,
    user_account,
    user_account_hash,
    nickname,
    email,
    phone,
    role,
    status,
    alarm_levels,
    alarm_channels,
    alarm_scope,
    last_login_at,
    user_tags::text,
    preferences::text
FROM users
WHERE user_id = 'e0f23dda-ee49-4915-9d53-d23b2e5e045a';
