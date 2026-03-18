-- 清除 residents 表内所有位置绑定关系（unit_id, room_id, bed_id）
-- 执行前会先查询将受影响的行数
BEGIN;
SELECT COUNT(*) AS "将清除绑定的住户数" FROM residents WHERE unit_id IS NOT NULL OR room_id IS NOT NULL OR bed_id IS NOT NULL;
UPDATE residents SET unit_id = NULL, room_id = NULL, bed_id = NULL WHERE unit_id IS NOT NULL OR room_id IS NOT NULL OR bed_id IS NOT NULL;
SELECT COUNT(*) AS "仍含绑定的住户数(应为0)" FROM residents WHERE unit_id IS NOT NULL OR room_id IS NOT NULL OR bed_id IS NOT NULL;
COMMIT;
