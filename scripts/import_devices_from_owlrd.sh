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
    -- 历史归一：当前所有 Radar=HC2 / Sleepad=BM8701-2；db1.0 中 NULL 或 'M871W' 都映射到正典型号
    CASE
        WHEN s.device_type = 'Radar'   THEN 'HC2'
        WHEN s.device_type = 'Sleepad' THEN 'BM8701-2'
        ELSE NULLIF(s.device_model, '')
    END,
    NULLIF(s.mac, ''),
    NULLIF(s.imei, ''),
    NULLIF(s.comm_mode, ''),
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
