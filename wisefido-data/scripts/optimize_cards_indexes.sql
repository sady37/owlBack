-- Optimize cards table indexes for getCardIDsForStaff query performance
-- owlRD/db/26_cards.sql 已定义: idx_cards_tenant, idx_cards_type, idx_cards_unit_id, idx_cards_resident_id
-- 本脚本补充对 getCardIDsForStaff 三种 scope 有用的索引（若 owlRD 已建则 IF NOT EXISTS 跳过）

-- ALL: WHERE c.tenant_id = $1（owlRD 已有 idx_cards_tenant）
-- BRANCH: cards JOIN units ON c.unit_id = u.unit_id，units 有 idx_units_branch_id(tenant_id, branch_id)
-- ASSIGNED_ONLY: tenant_id + (unit_id = ANY OR (card_type + resident_id = ANY))

-- 补充：unit_id 单列索引，便于 cards JOIN units 时按 unit_id 查找
CREATE INDEX IF NOT EXISTS idx_cards_unit ON cards (unit_id) WHERE unit_id IS NOT NULL;

-- 复合索引：ASSIGNED_ONLY 中 card_type = 'ActiveBed' AND resident_id = ANY()
-- owlRD 已有 idx_cards_resident_id(tenant_id, resident_id)，此处可选加速 (tenant_id, card_type, resident_id)
CREATE INDEX IF NOT EXISTS idx_cards_tenant_type_resident ON cards (tenant_id, card_type, resident_id) WHERE resident_id IS NOT NULL;

-- 生产环境可改用 CONCURRENTLY 避免锁表（需单独执行每条）:
-- CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_cards_unit ON cards (unit_id) WHERE unit_id IS NOT NULL;
-- CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_cards_tenant_type_resident ON cards (tenant_id, card_type, resident_id) WHERE resident_id IS NOT NULL;

-- 执行: docker exec -i owl-postgresql psql -U postgres -d owlrd -f scripts/optimize_cards_indexes.sql