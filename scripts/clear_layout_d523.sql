-- 清除 E598A2ACD523 (D523) 相关 layout
-- 用法: psql $DATABASE_URL -f clear_layout_d523.sql
--
-- /128 device scope: fd00:0:3:111:3:100:a2ac:d523
-- /88 room scope:    fd00:0:3:111:3:100:: (D523 所在 room；按需取消注释)
-- /80 unit scope:    fd00:0:3:111:3:: (D523 所在 unit；按需取消注释)

\echo '=== 删除前：D523 相关 layout 行 ==='
SELECT
  spatial_prefix::text AS prefix,
  masklen(spatial_prefix) AS mask,
  version,
  jsonb_array_length(COALESCE(canvas->'objects', '[]'::jsonb)) AS object_count,
  updated_at
FROM room_visual_layout
WHERE spatial_prefix <<= 'fd00:0:3:111:3::/80'::INET
  AND host(spatial_prefix) LIKE '%a2ac:d523%'
   OR spatial_prefix = 'fd00:0:3:111:3:100:a2ac:d523'::INET;

\echo ''
\echo '=== 执行删除（仅 /128 device scope）==='
DELETE FROM room_visual_layout
WHERE spatial_prefix = 'fd00:0:3:111:3:100:a2ac:d523'::INET;

\echo ''
\echo '=== 删除后剩余行（应该是 0 行 D523/128 层）==='
SELECT
  spatial_prefix::text AS prefix,
  masklen(spatial_prefix) AS mask,
  version,
  jsonb_array_length(COALESCE(canvas->'objects', '[]'::jsonb)) AS object_count
FROM room_visual_layout
WHERE spatial_prefix = 'fd00:0:3:111:3:100:a2ac:d523'::INET;

-- 如果要连 /88 room scope 也清掉（注意：会影响该 room 所有设备的 layout），取消下面注释：
-- DELETE FROM room_visual_layout WHERE spatial_prefix = 'fd00:0:3:111:3:100::/88'::INET;
