#!/bin/bash

# Docker下重建cards表脚本
# 用途：完全删除旧表并重新创建优化后的表结构
# 使用方式：
#   1. 确保 docker-compose 中数据库容器正在运行
#   2. ./rebuild_cards_table.sh
#   3. 运行完成后验证

set -e

DOCKER_COMPOSE_FILE="${1:-.}"

echo "=== Docker Cards Table Rebuild ==="
echo "Compose file: $DOCKER_COMPOSE_FILE"

# 检查 docker-compose.yml 存在
if [ ! -f "$DOCKER_COMPOSE_FILE/docker-compose.yml" ]; then
    echo "Error: docker-compose.yml not found in $DOCKER_COMPOSE_FILE"
    exit 1
fi

# 确保容器运行
echo "Checking PostgreSQL container..."
docker-compose -f "$DOCKER_COMPOSE_FILE/docker-compose.yml" ps | grep postgres || {
    echo "Starting PostgreSQL container..."
    docker-compose -f "$DOCKER_COMPOSE_FILE/docker-compose.yml" up -d postgres
    sleep 3
}

# 获取数据库连接信息（从 docker-compose.yml 读取）
DB_NAME=$(docker-compose -f "$DOCKER_COMPOSE_FILE/docker-compose.yml" config | grep -A5 'postgres:' | grep 'POSTGRES_DB' | head -1 | grep -oP '(?<=:\s)[^:]+(?=\s*$)' || echo "owlback")
DB_USER=$(docker-compose -f "$DOCKER_COMPOSE_FILE/docker-compose.yml" config | grep -A5 'postgres:' | grep 'POSTGRES_USER' | head -1 | grep -oP '(?<=:\s)[^:]+(?=\s*$)' || echo "owlback")
DB_PASSWORD=$(docker-compose -f "$DOCKER_COMPOSE_FILE/docker-compose.yml" config | grep -A5 'postgres:' | grep 'POSTGRES_PASSWORD' | head -1 | grep -oP '(?<=:\s)[^:]+(?=\s*$)' || echo "owlback")
DB_HOST="localhost"
DB_PORT="5432"

echo "Database: $DB_NAME"
echo "User: $DB_USER"

# 执行重建SQL
echo ""
echo "=== Executing reconstruction SQL ==="

# 创建临时SQL脚本
TEMP_SQL=$(mktemp)
cat > "$TEMP_SQL" << 'EOSQL'
-- 重建 cards 表脚本
-- 警告：此脚本将删除现有 cards 表及其所有数据！

-- Step 1: 检查依赖（显示所有引用 cards 表的视图和对象）
-- 查询结果会显示需要手动处理的依赖

-- Step 2: 删除旧表（级联删除依赖对象）
DROP TABLE IF NOT EXISTS cards CASCADE;

-- Step 3: 重新创建优化后的表结构
CREATE TABLE cards (
    card_id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    branch_id  UUID REFERENCES branches(branch_id) ON DELETE SET NULL,

    -- Card type: 'ActiveBed' / 'Location'
    card_type  VARCHAR(20) NOT NULL,

    -- Binding target
    bed_id     UUID REFERENCES beds(bed_id) ON DELETE CASCADE,
    unit_id    UUID NOT NULL REFERENCES units(unit_id) ON DELETE CASCADE,

    -- Display
    card_name     VARCHAR(255) NOT NULL,
    card_address  VARCHAR(255) NOT NULL,

    -- Primary resident for ActiveBed (Location may be NULL)
    resident_id UUID REFERENCES residents(resident_id) ON DELETE SET NULL,

    -- Precomputed associations (application-maintained)
    devices   JSONB NOT NULL DEFAULT '[]'::jsonb,
    residents JSONB NOT NULL DEFAULT '[]'::jsonb,

    -- Unit timezone (IANA format)
    timezone  VARCHAR(50) NOT NULL DEFAULT 'UTC',

    -- Unhandled alarm counters
    unhandled_alarm_0 INTEGER NOT NULL DEFAULT 0,
    unhandled_alarm_1 INTEGER NOT NULL DEFAULT 0,
    unhandled_alarm_2 INTEGER NOT NULL DEFAULT 0,
    unhandled_alarm_3 INTEGER NOT NULL DEFAULT 0,
    unhandled_alarm_4 INTEGER NOT NULL DEFAULT 0,

    -- UI alarm thresholds
    icon_alarm_level INTEGER NOT NULL DEFAULT 3 CHECK (icon_alarm_level >= 0 AND icon_alarm_level <= 4),
    pop_alarm_emerge INTEGER NOT NULL DEFAULT 0 CHECK (pop_alarm_emerge >= 0 AND pop_alarm_emerge <= 4),

    -- Constraint
    CONSTRAINT chk_card_type_binding CHECK (
        (card_type = 'ActiveBed' AND bed_id IS NOT NULL) OR
        (card_type = 'Location' AND unit_id IS NOT NULL AND bed_id IS NULL)
    )
);

-- Step 4: 创建索引
-- Uniqueness
CREATE UNIQUE INDEX ux_cards_activebed_bed
  ON cards(tenant_id, bed_id)
  WHERE card_type = 'ActiveBed' AND bed_id IS NOT NULL;

CREATE UNIQUE INDEX ux_cards_location_unit
  ON cards(tenant_id, unit_id)
  WHERE card_type = 'Location' AND unit_id IS NOT NULL;

-- Query indexes
CREATE INDEX idx_cards_tenant ON cards(tenant_id);
CREATE INDEX idx_cards_tenant_branch ON cards(tenant_id, branch_id);
CREATE INDEX idx_cards_type ON cards(tenant_id, card_type);
CREATE INDEX idx_cards_bed_id ON cards(tenant_id, bed_id) WHERE bed_id IS NOT NULL;
CREATE INDEX idx_cards_unit_id ON cards(tenant_id, unit_id) WHERE unit_id IS NOT NULL;
CREATE INDEX idx_cards_resident_id ON cards(tenant_id, resident_id) WHERE resident_id IS NOT NULL;
CREATE INDEX idx_cards_unit ON cards(unit_id) WHERE unit_id IS NOT NULL;
CREATE INDEX idx_cards_tenant_type_resident ON cards(tenant_id, card_type, resident_id) WHERE resident_id IS NOT NULL;

CREATE INDEX idx_cards_unhandled_alarms ON cards(
  tenant_id,
  (unhandled_alarm_0 + unhandled_alarm_1 + unhandled_alarm_2 + unhandled_alarm_3 + unhandled_alarm_4)
) WHERE (unhandled_alarm_0 + unhandled_alarm_1 + unhandled_alarm_2 + unhandled_alarm_3 + unhandled_alarm_4) > 0;

CREATE INDEX idx_cards_devices ON cards USING GIN (devices) WHERE jsonb_typeof(devices) = 'array' AND jsonb_array_length(devices) > 0;
CREATE INDEX idx_cards_residents ON cards USING GIN (residents) WHERE jsonb_typeof(residents) = 'array' AND jsonb_array_length(residents) > 0;

-- Step 5: 添加 COMMENT
COMMENT ON TABLE cards IS 'Card (卡片) - 房间/床位的聚合视图，包含展示信息和告警配置';
COMMENT ON COLUMN cards.branch_id IS '院区ID（从units.branch_id冗余存储，用于快速查询、消息发送、缓存失效）';
COMMENT ON COLUMN cards.unit_id IS '单元ID（NotNull，来自units）';
COMMENT ON COLUMN cards.timezone IS 'Unit timezone (IANA), from units.timezone at card create/update';
COMMENT ON COLUMN cards.devices IS 'Associated devices JSON array (device_id, device_uid, device_name, device_type)';
COMMENT ON COLUMN cards.residents IS 'Associated residents JSON array (resident_id, nickname)';

EOSQL

# 通过 docker-compose exec 执行 SQL
echo "Executing SQL via docker-compose..."
docker-compose -f "$DOCKER_COMPOSE_FILE/docker-compose.yml" exec -T postgres psql \
    -h "$DB_HOST" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    -f - < "$TEMP_SQL"

# 清理临时文件
rm -f "$TEMP_SQL"

echo ""
echo "=== Verification ==="

# 验证表结构
docker-compose -f "$DOCKER_COMPOSE_FILE/docker-compose.yml" exec -T postgres psql \
    -h "$DB_HOST" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    -c "SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_name = 'cards' ORDER BY ordinal_position;"

echo ""
echo "=== Index Verification ==="
docker-compose -f "$DOCKER_COMPOSE_FILE/docker-compose.yml" exec -T postgres psql \
    -h "$DB_HOST" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    -c "SELECT indexname FROM pg_indexes WHERE tablename = 'cards' ORDER BY indexname;"

echo ""
echo "✅ Cards table rebuild complete!"
echo "Note: Application layer must update CreateCard() to populate branch_id from units.branch_id"
