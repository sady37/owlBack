#!/bin/bash

# Docker下完整的card表重建脚本（含数据同步）
# 用途：生产环境或测试环境中完整重建card表结构
# 执行时机：在停服或维护窗口内执行

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOCKER_COMPOSE_FILE="${1:-$SCRIPT_DIR}"
DB_NAME="${2:-owlback}"
DB_USER="${3:-owlback}"

echo "════════════════════════════════════════════════════════"
echo "  Card Table Complete Rebuild (Docker Environment)"
echo "════════════════════════════════════════════════════════"
echo ""
echo "Docker Compose: $DOCKER_COMPOSE_FILE"
echo "Database:       $DB_NAME"
echo "User:           $DB_USER"
echo ""

# 确保compose文件存在
if [ ! -f "$DOCKER_COMPOSE_FILE/docker-compose.yml" ]; then
    echo "❌ Error: docker-compose.yml not found at $DOCKER_COMPOSE_FILE"
    exit 1
fi

# Step 1: 启动PostgreSQL容器
echo "📋 Step 1: Ensure PostgreSQL container is running..."
if docker-compose -f "$DOCKER_COMPOSE_FILE/docker-compose.yml" ps postgres 2>/dev/null | grep -q "Up"; then
    echo "✅ PostgreSQL is already running"
else
    echo "🚀 Starting PostgreSQL..."
    docker-compose -f "$DOCKER_COMPOSE_FILE/docker-compose.yml" up -d postgres
    sleep 3
    echo "✅ PostgreSQL started"
fi

# Step 2: 备份现有数据（可选）
echo ""
echo "📋 Step 2: Backup existing cards data (optional)..."
BACKUP_FILE="/tmp/cards_backup_$(date +%Y%m%d_%H%M%S).sql"
echo "Creating backup at $BACKUP_FILE..."

docker-compose -f "$DOCKER_COMPOSE_FILE/docker-compose.yml" exec -T postgres pg_dump \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    -t cards \
    --data-only > "$BACKUP_FILE" 2>/dev/null || true

if [ -s "$BACKUP_FILE" ]; then
    echo "✅ Backup created: $BACKUP_FILE"
else
    echo "ℹ️  No existing data to backup (first time setup)"
    rm -f "$BACKUP_FILE"
fi

# Step 3: 执行表重建SQL
echo ""
echo "📋 Step 3: Rebuild cards table..."
echo "⚠️  Existing cards table will be dropped!"
echo ""

# 创建SQL脚本
SQL_SCRIPT=$(mktemp)
cat > "$SQL_SCRIPT" << 'EOSQL'
-- ============================================================
-- Card Table Complete Rebuild
-- ============================================================

-- Drop old table (cascade to remove dependent objects)
DROP TABLE IF NOT EXISTS cards CASCADE;

-- Create optimized table
CREATE TABLE cards (
    card_id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    branch_id  UUID REFERENCES branches(branch_id) ON DELETE SET NULL,

    -- Card type
    card_type  VARCHAR(20) NOT NULL,

    -- Binding target
    bed_id     UUID REFERENCES beds(bed_id) ON DELETE CASCADE,
    unit_id    UUID NOT NULL REFERENCES units(unit_id) ON DELETE CASCADE,

    -- Display
    card_name     VARCHAR(255) NOT NULL,
    card_address  VARCHAR(255) NOT NULL,

    -- Primary resident
    resident_id UUID REFERENCES residents(resident_id) ON DELETE SET NULL,

    -- Precomputed data
    devices   JSONB NOT NULL DEFAULT '[]'::jsonb,
    residents JSONB NOT NULL DEFAULT '[]'::jsonb,

    -- Unit timezone
    timezone  VARCHAR(50) NOT NULL DEFAULT 'UTC',

    -- Alarm counters
    unhandled_alarm_0 INTEGER NOT NULL DEFAULT 0,
    unhandled_alarm_1 INTEGER NOT NULL DEFAULT 0,
    unhandled_alarm_2 INTEGER NOT NULL DEFAULT 0,
    unhandled_alarm_3 INTEGER NOT NULL DEFAULT 0,
    unhandled_alarm_4 INTEGER NOT NULL DEFAULT 0,

    -- UI thresholds
    icon_alarm_level INTEGER NOT NULL DEFAULT 3 CHECK (icon_alarm_level >= 0 AND icon_alarm_level <= 4),
    pop_alarm_emerge INTEGER NOT NULL DEFAULT 0 CHECK (pop_alarm_emerge >= 0 AND pop_alarm_emerge <= 4),

    -- Type constraint
    CONSTRAINT chk_card_type_binding CHECK (
        (card_type = 'ActiveBed' AND bed_id IS NOT NULL) OR
        (card_type = 'Location' AND unit_id IS NOT NULL AND bed_id IS NULL)
    )
);

-- Unique indexes
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

-- Performance indexes
CREATE INDEX idx_cards_unhandled_alarms ON cards(
  tenant_id,
  (unhandled_alarm_0 + unhandled_alarm_1 + unhandled_alarm_2 + unhandled_alarm_3 + unhandled_alarm_4)
) WHERE (unhandled_alarm_0 + unhandled_alarm_1 + unhandled_alarm_2 + unhandled_alarm_3 + unhandled_alarm_4) > 0;

CREATE INDEX idx_cards_devices ON cards USING GIN (devices) 
WHERE jsonb_typeof(devices) = 'array' AND jsonb_array_length(devices) > 0;

CREATE INDEX idx_cards_residents ON cards USING GIN (residents) 
WHERE jsonb_typeof(residents) = 'array' AND jsonb_array_length(residents) > 0;

-- Comments
COMMENT ON TABLE cards IS 'Card (卡片) - 房间/床位的聚合视图，包含展示信息和告警配置';
COMMENT ON COLUMN cards.branch_id IS '院区ID（从units.branch_id冗余存储，用于快速查询、消息发送、缓存失效）';
COMMENT ON COLUMN cards.unit_id IS '单元ID（Not Null，来自units）';
COMMENT ON COLUMN cards.timezone IS 'Unit timezone (IANA format)';
COMMENT ON COLUMN cards.devices IS 'Associated devices JSON array';
COMMENT ON COLUMN cards.residents IS 'Associated residents JSON array';

-- Verification
SELECT 'Cards table rebuilt successfully' as status;
SELECT COUNT(*) as table_size FROM information_schema.tables WHERE table_name = 'cards';

EOSQL

# 执行SQL脚本
echo "Running SQL script..."
docker-compose -f "$DOCKER_COMPOSE_FILE/docker-compose.yml" exec -T postgres psql \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    -f - < "$SQL_SCRIPT"

rm -f "$SQL_SCRIPT"
echo "✅ Table rebuild complete"

# Step 4: 验证新表结构
echo ""
echo "📋 Step 4: Verify table structure..."
echo ""
echo "Columns:"
docker-compose -f "$DOCKER_COMPOSE_FILE/docker-compose.yml" exec -T postgres psql \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    -c "
    SELECT 
        column_name, 
        data_type, 
        is_nullable,
        column_default
    FROM information_schema.columns 
    WHERE table_name = 'cards' 
    ORDER BY ordinal_position;
    "

echo ""
echo "Indexes:"
docker-compose -f "$DOCKER_COMPOSE_FILE/docker-compose.yml" exec -T postgres psql \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    -c "
    SELECT 
        indexname,
        indexdef
    FROM pg_indexes 
    WHERE tablename = 'cards' 
    ORDER BY indexname;
    " | head -20

echo ""
echo "Constraints:"
docker-compose -f "$DOCKER_COMPOSE_FILE/docker-compose.yml" exec -T postgres psql \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    -c "
    SELECT 
        constraint_name,
        constraint_type
    FROM information_schema.table_constraints 
    WHERE table_name = 'cards' 
    ORDER BY constraint_name;
    "

# Step 5: 完成
echo ""
echo "════════════════════════════════════════════════════════"
echo "✅ Card Table Rebuild Complete!"
echo "════════════════════════════════════════════════════════"
echo ""
echo "Next steps:"
echo "1. Deploy updated wisefido-data (postgres_card.go + card_sync_service.go)"
echo "2. Restart application containers"
echo "3. Monitor logs for any issues"
echo "4. Verify card creation/update operations"
echo ""
if [ -f "$BACKUP_FILE" ]; then
    echo "Backup location: $BACKUP_FILE"
fi
echo ""
echo "For testing card operations:"
echo "  - Create a new card: POST /api/cards"
echo "  - Update existing card: PUT /api/cards/{cardID}"
echo "  - Verify branch_id is populated correctly"
echo ""
