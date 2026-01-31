#!/bin/bash
# 清空 Docker PostgreSQL 中指定表的记录
# 表：config_versions, iot_timeseries, alarm_cloud, alarm_device

echo "正在清空 Docker PostgreSQL 中的表记录..."
echo "表：config_versions, iot_timeseries, alarm_cloud, alarm_device"
echo ""

# 执行 SQL 脚本
docker exec -i owl-postgresql psql -U postgres -d owlrd < "$(dirname "$0")/clear_tables.sql"

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ 表记录已成功清空！"
else
    echo ""
    echo "❌ 清空表记录时出错，请检查错误信息"
    exit 1
fi
