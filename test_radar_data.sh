#!/bin/bash

echo "🔍 测试雷达数据流"
echo "=================="

echo ""
echo "1. 检查 Redis Streams 状态:"
cd /Users/sady3721/project/owlBack
for stream in "radar:monitor:stream" "radar:stat:stream" "radar:event:stream" "radar:alarm:stream"; do
    count=$(docker-compose exec redis redis-cli -a TeLunSu-36kr XLEN "$stream" 2>/dev/null)
    echo "   $stream: $count 条数据"
done

echo ""
echo "2. 检查数据库数据:"
docker-compose exec postgresql psql -U postgres -d owlrd -c "SELECT COUNT(*) FROM iot_timeseries;"

echo ""
echo "3. 检查 wisefido-iot-timeseries 服务:"
ps aux | grep "wisefido-iot-timeseries" | grep -v grep | head -1

echo ""
echo "4. 检查 wisefido-radar 服务:"
ps aux | grep "wisefido-radar" | grep -v grep | head -1

echo ""
echo "5. 可能的测试方法:"
echo "   A. 使用 decode-track 工具测试:"
echo "      cd wisefido-radar && go run ./cmd/decode-track/ --source redis --stream radar:monitor:stream --count 1"
echo ""
echo "   B. 手动向 MQTT 发送测试数据:"
echo "      mosquitto_pub -h 127.0.0.1 -p 1883 -t 'radar/E598A2ACD523/data' -m 'test'"
echo ""
echo "   C. 检查设备是否在线发送数据"
echo ""
echo "6. 当前状态总结:"
echo "   - Redis Streams: 已创建但无数据"
echo "   - 数据库: 无数据"
echo "   - 服务: wisefido-radar 和 wisefido-iot-timeseries 在运行"
echo "   - 问题: 设备可能没有发送数据，或 wisefido-radar 没有处理数据"
