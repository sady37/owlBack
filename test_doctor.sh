#!/bin/bash

echo "🚀 测试 wisefido-data Doctor 功能"
echo "================================"

# 设置环境变量
export DB_HOST=127.0.0.1
export DB_PORT=5433
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=owlrd
export DB_SSLMODE=disable
export REDIS_ADDR=127.0.0.1:6379
export REDIS_PASSWORD=TeLunSu-36kr
export DOCTOR_ENABLED=true
export DOCTOR_PPROF=false
export HTTP_ADDR=:8080

echo "📊 配置:"
echo "  DB_HOST: $DB_HOST"
echo "  DB_PORT: $DB_PORT"
echo "  DOCTOR_ENABLED: $DOCTOR_ENABLED"
echo "  DOCTOR_PPROF: $DOCTOR_PPROF"
echo ""

echo "🔧 启动 wisefido-data..."
cd /Users/sady3721/project/owlBack/wisefido-data

# 在后台启动服务
go run cmd/wisefido-data/main.go > /tmp/wisefido-data.log 2>&1 &
PID=$!

echo "⏳ 等待服务启动..."
sleep 3

echo "📝 查看启动日志:"
tail -10 /tmp/wisefido-data.log

echo ""
echo "🔍 测试 Doctor 端点:"

# 测试健康检查
echo "1. 健康检查 (/health):"
curl -s http://localhost:8080/health | jq . 2>/dev/null || curl -s http://localhost:8080/health

echo ""
echo "2. 就绪检查 (/ready):"
curl -s http://localhost:8080/ready | jq . 2>/dev/null || curl -s http://localhost:8080/ready

echo ""
echo "3. 健康检查 (/healthz):"
curl -s http://localhost:8080/healthz | head -c 200

echo ""
echo "4. 就绪检查 (/readyz):"
curl -s http://localhost:8080/readyz | head -c 200

echo ""
echo "✅ 测试完成"
echo "🛑 停止服务: kill $PID"
kill $PID
