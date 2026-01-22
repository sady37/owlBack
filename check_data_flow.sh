#!/bin/bash

echo "🔍 检查数据流问题"
echo "=================="

echo ""
echo "1. 检查 PostgreSQL TimescaleDB:"
docker-compose exec postgresql psql -U postgres -d owlrd -c "SELECT extname, extversion FROM pg_extension WHERE extname = 'timescaledb';"
echo ""

echo "2. 检查 iot_timeseries 表数据:"
docker-compose exec postgresql psql -U postgres -d owlrd -c "SELECT COUNT(*) FROM iot_timeseries;"
echo ""

echo "3. 检查 Redis Streams:"
echo "   检查 radar:*:stream:"
redis-cli -h 127.0.0.1 -p 6379 -a TeLunSu-36kr KEYS "radar:*" 2>/dev/null || echo "   无 radar 相关键"
echo ""

echo "4. 检查服务状态:"
echo "   wisefido-radar:"
ps aux | grep "wisefido-radar" | grep -v grep | head -1
echo "   wisefido-iot-timeseries:"
ps aux | grep "wisefido-iot-timeseries" | grep -v grep | head -1
echo ""

echo "5. 可能的解决方案:"
echo "   A. wisefido-radar 可能没有发布数据到 Redis Streams"
echo "   B. 检查 wisefido-radar 配置中的 Redis Stream 名称"
echo "   C. 检查 wisefido-iot-timeseries 消费的 Stream 名称"
echo "   D. 确保 TimescaleDB 扩展已启用"
echo ""

echo "6. 测试 wisefido-radar 数据发布:"
echo "   查看 wisefido-radar 日志中是否有 'Published radar data to Redis Streams'"
echo ""

echo "7. 手动测试 Redis 连接:"
redis-cli -h 127.0.0.1 -p 6379 -a TeLunSu-36kr PING 2>/dev/null && echo "   ✅ Redis 连接正常" || echo "   ❌ Redis 连接失败"
echo ""

echo "📋 总结:"
echo "   - TimescaleDB 扩展: ✅ 已安装"
echo "   - iot_timeseries 表: ✅ 已创建为超表"
echo "   - 表中数据: ❌ 无数据"
echo "   - Redis Streams: ❌ 无数据"
echo "   - 问题: 数据没有从 wisefido-radar 流向 Redis Streams"
