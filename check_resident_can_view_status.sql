-- 检查 done 和 smith 的 can_view_status 值
SELECT 
    r.resident_id,
    r.tenant_id,
    r.resident_account,
    r.nickname,
    r.can_view_status,
    r.status,
    r.service_level,
    r.admission_date,
    r.family_tag
FROM residents r
WHERE LOWER(r.nickname) IN ('done', 'smith')
   OR LOWER(r.resident_account) IN ('done', 'smith')
ORDER BY r.nickname;

