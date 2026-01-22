#!/bin/bash
# 执行删除 device_uid 为空的设备记录
# 用法: ./run_delete_empty_uid_devices.sh [preview|execute]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SQL_FILE="${SCRIPT_DIR}/delete_devices_with_empty_uid.sql"

# 检查 SQL 文件是否存在
if [ ! -f "$SQL_FILE" ]; then
    echo "Error: SQL file not found: $SQL_FILE"
    exit 1
fi

# 从环境变量或 docker-compose 获取数据库连接信息
if [ -f "${SCRIPT_DIR}/../docker-compose.yml" ]; then
    # 从 docker-compose.yml 读取数据库配置
    DB_HOST="${DB_HOST:-localhost}"
    DB_PORT="${DB_PORT:-5432}"
    DB_NAME="${DB_NAME:-wisefido}"
    DB_USER="${DB_USER:-postgres}"
    DB_PASSWORD="${DB_PASSWORD:-postgres}"
    
    # 尝试从 docker-compose 获取
    if command -v docker-compose &> /dev/null; then
        COMPOSE_PROJECT_DIR="${SCRIPT_DIR}/.."
        export $(docker-compose -f "${COMPOSE_PROJECT_DIR}/docker-compose.yml" config 2>/dev/null | grep -A 5 "postgres:" | grep -E "POSTGRES_|POSTGRESDB" | sed 's/.*=//' | xargs) 2>/dev/null || true
    fi
else
    # 从环境变量读取
    DB_HOST="${DB_HOST:-localhost}"
    DB_PORT="${DB_PORT:-5432}"
    DB_NAME="${DB_NAME:-wisefido}"
    DB_USER="${DB_USER:-postgres}"
    DB_PASSWORD="${DB_PASSWORD:-postgres}"
fi

export PGPASSWORD="$DB_PASSWORD"

ACTION="${1:-preview}"

if [ "$ACTION" = "preview" ]; then
    echo "=== 预览：查看将要删除的设备 ==="
    echo ""
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$SQL_FILE" -c "
        SELECT 
            d.device_id,
            d.device_name,
            d.device_uid,
            d.tenant_id,
            ds.device_type,
            ds.device_code,
            is_device_used(d.device_id) AS is_used,
            CASE 
                WHEN is_device_used(d.device_id) THEN '已使用，将跳过'
                ELSE '未使用，将被删除'
            END AS action
        FROM devices d
        JOIN device_store ds ON d.device_id = ds.device_id
        WHERE d.device_uid IS NULL OR d.device_uid = ''
        ORDER BY d.tenant_id, d.device_name;
    "
    echo ""
    echo "=== 统计信息 ==="
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
        SELECT 
            COUNT(*) AS total_devices_with_empty_uid,
            COUNT(*) FILTER (WHERE NOT is_device_used(d.device_id)) AS devices_to_delete,
            COUNT(*) FILTER (WHERE is_device_used(d.device_id)) AS devices_to_skip
        FROM devices d
        WHERE d.device_uid IS NULL OR d.device_uid = '';
    "
    echo ""
    echo "提示: 要执行删除，请运行: $0 execute"
    
elif [ "$ACTION" = "execute" ]; then
    echo "=== 执行删除 device_uid 为空的设备记录 ==="
    echo "警告: 此操作将永久删除未使用的设备记录！"
    echo ""
    read -p "确认执行删除? (yes/no): " confirm
    if [ "$confirm" != "yes" ]; then
        echo "操作已取消"
        exit 0
    fi
    
    # 修改 SQL 文件，取消 COMMIT 的注释
    TEMP_SQL=$(mktemp)
    sed 's/^-- COMMIT;/COMMIT;/' "$SQL_FILE" > "$TEMP_SQL"
    
    echo "正在执行删除..."
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$TEMP_SQL"
    
    rm -f "$TEMP_SQL"
    
    echo ""
    echo "=== 删除完成 ==="
    echo "查看剩余的空 device_uid 设备:"
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
        SELECT COUNT(*) AS remaining_empty_uid_devices
        FROM devices
        WHERE device_uid IS NULL OR device_uid = '';
    "
else
    echo "用法: $0 [preview|execute]"
    echo "  preview  - 预览将要删除的设备（默认）"
    echo "  execute  - 执行删除操作"
    exit 1
fi
