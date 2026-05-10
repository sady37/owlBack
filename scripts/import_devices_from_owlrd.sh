#!/bin/bash
# import_devices_from_owlrd.sh
# 将 owlrd (db1.0) device_store 导入 owl_v2 device_factory_meta + device_runtime_state
#
# 设计：
#   - "device 自身项" = factory metadata + 当前 firmware version
#   - 不导 spatial 绑定（dbv2 devices 表需 unit /80 存在；待 unit/room 重建后再做）
#   - 不导 OTA 字段（dbv2 device_ota 是独立表，待 OTA workflow 复用时再做）
#   - 跨库走 dblink（owlrd / owl_v2 同实例）
#
# 用法（从 owlBack/ 目录）：
#   bash scripts/import_devices_from_owlrd.sh           # 导入；冲突跳过
#   bash scripts/import_devices_from_owlrd.sh --truncate # 先清空 dbv2 device_factory_meta+runtime
#   bash scripts/import_devices_from_owlrd.sh --check    # 仅显示 owlrd vs owl_v2 计数

set -euo pipefail

CONTAINER=${OWL_PG_CONTAINER:-owl-postgresql}
SRC_DB=${OWL_V1_DB:-owlrd}
DST_DB=${OWL_V2_DB:-owl_v2}

cmd_check() {
    echo "[check] owlrd device_store vs owl_v2 device_factory_meta"
    docker exec -i "$CONTAINER" psql -U postgres <<EOF
\c $SRC_DB
SELECT 'owlrd.device_store' AS src, device_type, COALESCE(NULLIF(device_model,''), '<NULL>') AS model, COUNT(*)
FROM device_store GROUP BY 2,3 ORDER BY 2,3;
\c $DST_DB
SELECT 'owl_v2.device_factory_meta' AS dst, device_type::text, COALESCE(device_model, '<NULL>') AS model, COUNT(*)
FROM device_factory_meta GROUP BY 2,3 ORDER BY 2,3;
EOF
}

cmd_truncate() {
    echo "[!] TRUNCATE owl_v2 device_runtime_state, device_factory_meta"
    docker exec -i "$CONTAINER" psql -U postgres -d "$DST_DB" <<'EOF'
TRUNCATE device_runtime_state, device_factory_meta CASCADE;
EOF
}

cmd_import() {
    echo "[+] dblink import: $SRC_DB.device_store → $DST_DB.device_factory_meta + .device_runtime_state"
    docker exec -i "$CONTAINER" psql -U postgres -d "$DST_DB" -v ON_ERROR_STOP=1 <<EOF
CREATE EXTENSION IF NOT EXISTS dblink;

-- factory_meta
INSERT INTO device_factory_meta (
    device_id, device_uid, device_code, device_type, device_model,
    mac_wifi, imei, comm_mode, mcu_model, import_date, notes
)
SELECT
    s.device_id,
    s.device_uid,
    NULLIF(s.device_code, ''),
    s.device_type::device_type_enum,
    -- 型号归一规则（按 device_uid 前缀，避免误覆盖新款）：
    --   Radar:   全部 HC2（v1 库存已确认无其它型号）
    --   Sleepad: device_uid 前缀决定 — M871* → M871W (新款) / BM87* → BM8701-2 / 其它保留 v1 原值
    CASE
        WHEN s.device_type = 'Radar'   THEN 'HC2'
        WHEN s.device_type = 'Sleepad' AND s.device_uid LIKE 'M871%' THEN 'M871W'
        WHEN s.device_type = 'Sleepad' AND s.device_uid LIKE 'BM87%' THEN 'BM8701-2'
        WHEN s.device_type = 'Sleepad' THEN COALESCE(NULLIF(s.device_model, ''), 'BM8701-2')
        ELSE NULLIF(s.device_model, '')
    END,
    NULLIF(s.mac, ''),
    NULLIF(s.imei, ''),
    -- Radar/Sleepad 当前库存全 wifi；v1 NULL 兜底为 'wifi'
    COALESCE(NULLIF(s.comm_mode, ''),
             CASE WHEN s.device_type IN ('Radar','Sleepad') THEN 'wifi' END),
    NULLIF(s.mcu_model, ''),
    s.import_date,
    'Imported from owlrd.device_store on ' || NOW()::DATE
FROM dblink('dbname=$SRC_DB',
    'SELECT device_id, device_uid, device_code, device_type, device_model,
            mac, imei, comm_mode, mcu_model, import_date
     FROM device_store'
) AS s(device_id UUID, device_uid VARCHAR, device_code VARCHAR, device_type VARCHAR,
       device_model VARCHAR, mac VARCHAR, imei VARCHAR, comm_mode VARCHAR,
       mcu_model VARCHAR, import_date TIMESTAMPTZ)
WHERE s.device_uid IS NOT NULL                          -- 跳过 device_uid 为 NULL 的（无 logMAC 记录）
ON CONFLICT (device_id) DO NOTHING;

-- runtime_state（仅 firmware_version snapshot，online=false）
INSERT INTO device_runtime_state (
    device_id, online, firmware_version, updated_at
)
SELECT
    s.device_id,
    FALSE,
    NULLIF(s.firmware_version, ''),
    NOW()
FROM dblink('dbname=$SRC_DB',
    'SELECT device_id, firmware_version FROM device_store'
) AS s(device_id UUID, firmware_version VARCHAR)
WHERE s.device_id IN (SELECT device_id FROM device_factory_meta)
ON CONFLICT (device_id) DO NOTHING;

-- devices（按 v1 device_store.tenant_id 分配 spatial_addr = tenant_prefix:0:0:0:0:MAC32）
-- 每条 factory_meta 都建一条 devices 行，落在 tenant /48 unbound pool（最浅池）。
-- v1 tenant UUID → v2 tenant /48 prefix：System=fd00:0:1 / demo=fd00:0:3 / wisefido=fd00:0:4 / WeCare=fd00:0:5
WITH v1_assign AS (
    SELECT * FROM dblink('dbname=$SRC_DB',
        \$\$SELECT device_uid, tenant_id::text FROM device_store WHERE device_uid IS NOT NULL\$\$
    ) AS x(device_uid VARCHAR, v1_tenant_uuid TEXT)
),
tenant_map AS (
    SELECT * FROM (VALUES
        ('00000000-0000-0000-0000-000000000001', 'fd00:0:1'),
        ('43b8fbf7-b55f-4b48-bd8b-27bb14f48870', 'fd00:0:3'),
        ('6f1cbcdc-a7f6-4080-86f9-6ae2ecdfc4fe', 'fd00:0:4'),
        ('261a9010-8b4e-4c93-9b16-e6b3d6f01a13', 'fd00:0:5')
    ) AS t(v1_uuid, v2_prefix6)
),
dev_with_addr AS (
    SELECT
        dfm.device_id,
        COALESCE(tm.v2_prefix6, 'fd00:0:1') AS tenant_prefix6,  -- sleepace-only / 未知 → System
        LOWER(
            CASE
                WHEN dfm.mac_wifi IS NOT NULL AND LENGTH(REGEXP_REPLACE(dfm.mac_wifi, '[^0-9A-Fa-f]', '', 'g')) >= 8
                    THEN RIGHT(REGEXP_REPLACE(dfm.mac_wifi, '[^0-9A-Fa-f]', '', 'g'), 8)
                WHEN dfm.device_uid ~ '^[0-9A-Fa-f]{12}$' THEN RIGHT(dfm.device_uid, 8)
                ELSE LEFT(MD5(dfm.device_uid), 8)
            END
        ) AS mac32_hex
    FROM device_factory_meta dfm
    LEFT JOIN v1_assign va ON va.device_uid = dfm.device_uid
    LEFT JOIN tenant_map tm ON tm.v1_uuid = va.v1_tenant_uuid
)
INSERT INTO devices (spatial_addr, device_id, monitoring_enabled)
SELECT
    (d.tenant_prefix6 || '::' || SUBSTR(d.mac32_hex, 1, 4) || ':' || SUBSTR(d.mac32_hex, 5, 4) || '/128')::INET,
    d.device_id, TRUE
FROM dev_with_addr d
ON CONFLICT (device_id) DO UPDATE SET spatial_addr = EXCLUDED.spatial_addr;

-- 报告
SELECT
    (SELECT COUNT(*) FROM device_factory_meta)    AS factory_meta_total,
    (SELECT COUNT(*) FROM device_runtime_state)   AS runtime_state_total;

SELECT device_type, COALESCE(device_model, '<NULL>') AS model, COUNT(*)
FROM device_factory_meta GROUP BY 1,2 ORDER BY 1,2;
EOF
}

case "${1:-}" in
    --check)
        cmd_check
        ;;
    --truncate)
        cmd_truncate
        cmd_import
        ;;
    *)
        cmd_import
        ;;
esac

echo "[done] device import complete"
echo
echo "Note: spatial 绑定 (devices 表) deferred — 等 branch/site/unit/room/bed 从备份恢复后再跑"
