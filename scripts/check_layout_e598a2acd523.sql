-- 检查 E598A2ACD523 所在 unit/room 是否有 layout
-- 用法: psql $DATABASE_URL -f check_layout_e598a2acd523.sql

\echo '=== 1. 设备 E598A2ACD523 所在 room / unit ==='
SELECT
  d.device_id::text,
  d.device_uid,
  d.device_name,
  d.bound_room_id::text,
  d.bound_bed_id::text,
  COALESCE(
    d.bound_room_id,
    (SELECT b.room_id FROM beds b WHERE b.bed_id = d.bound_bed_id LIMIT 1)
  )::text AS resolved_room_id,
  r.room_name,
  u.unit_id::text,
  u.unit_name
FROM devices d
LEFT JOIN device_store ds ON d.device_id = ds.device_id
LEFT JOIN rooms r ON r.room_id = COALESCE(
  d.bound_room_id,
  (SELECT b.room_id FROM beds b WHERE b.bed_id = d.bound_bed_id AND b.tenant_id = d.tenant_id LIMIT 1)
)
LEFT JOIN units u ON r.unit_id = u.unit_id
WHERE (d.device_uid = 'E598A2ACD523' OR ds.device_uid = 'E598A2ACD523');

\echo ''
\echo '=== 2. unit_name / room_name 含 E102 的 room 及其 layout 情况 ==='
SELECT
  r.room_id::text,
  r.room_name,
  u.unit_name,
  r.layout_config IS NOT NULL AS room_has_layout,
  length(r.layout_config::text) AS room_layout_len,
  u.layout_config IS NOT NULL AS unit_has_layout,
  length(u.layout_config::text) AS unit_layout_len
FROM rooms r
JOIN units u ON r.unit_id = u.unit_id
WHERE u.unit_name ILIKE '%E102%' OR r.room_name ILIKE '%E102%';

\echo ''
\echo '=== 3. 上述设备解析出的 room 的 layout 详情（若 resolved_room_id 来自上面）==='
WITH dev_room AS (
  SELECT
    d.tenant_id,
    COALESCE(
      d.bound_room_id,
      (SELECT b.room_id FROM beds b WHERE b.bed_id = d.bound_bed_id AND b.tenant_id = d.tenant_id LIMIT 1)
    ) AS room_id
  FROM devices d
  LEFT JOIN device_store ds ON d.device_id = ds.device_id
  WHERE d.device_uid = 'E598A2ACD523' OR ds.device_uid = 'E598A2ACD523'
  LIMIT 1
)
SELECT
  r.room_id::text,
  r.room_name,
  u.unit_name,
  r.layout_config IS NOT NULL AS room_has_layout,
  left(r.layout_config::text, 200) AS room_layout_preview,
  u.layout_config IS NOT NULL AS unit_has_layout
FROM dev_room dr
JOIN rooms r ON r.room_id = dr.room_id AND r.tenant_id = dr.tenant_id
JOIN units u ON u.unit_id = r.unit_id;

\echo ''
\echo '=== 4. config_versions 中该 room 的 room_layout 条数 ==='
WITH dev_room AS (
  SELECT COALESCE(
      d.bound_room_id,
      (SELECT b.room_id FROM beds b WHERE b.bed_id = d.bound_bed_id AND b.tenant_id = d.tenant_id LIMIT 1)
    )::text AS room_id
  FROM devices d
  LEFT JOIN device_store ds ON d.device_id = ds.device_id
  WHERE d.device_uid = 'E598A2ACD523' OR ds.device_uid = 'E598A2ACD523'
  LIMIT 1
)
SELECT cv.entity_id, cv.config_type, count(*) AS versions
FROM config_versions cv
JOIN dev_room dr ON cv.entity_id = dr.room_id
WHERE cv.config_type = 'room_layout'
GROUP BY cv.entity_id, cv.config_type;
