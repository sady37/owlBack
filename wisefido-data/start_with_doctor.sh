#!/bin/bash

# 启用 wisefido-data Doctor 功能的启动脚本

echo "🚀 启动 wisefido-data 服务（Doctor 功能已启用）"
echo "=========================================="

# 检查 Docker 是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker 未运行，请先启动 Docker"
    exit 1
fi

# 切换到项目目录
cd /Users/sady3721/project/owlBack

echo "📦 检查依赖服务..."
docker-compose ps postgresql redis > /dev/null 2>&1

if [ $? -ne 0 ]; then
    echo "🔧 启动依赖服务（PostgreSQL 和 Redis）..."
    docker-compose up -d postgresql redis
    
    echo "⏳ 等待依赖服务就绪..."
    sleep 10
fi

echo "🔨 构建并启动 wisefido-data..."
docker-compose up -d --build wisefido-data

echo "⏳ 等待服务启动..."
sleep 5

echo "📊 检查服务状态..."
docker-compose ps wisefido-data

echo "📝 查看日志..."
docker-compose logs --tail=10 wisefido-data | grep -i "doctor\|starting\|error\|fatal"

echo "✅ 服务启动完成！"
echo ""
echo "🔍 测试 Doctor 端点："
echo "  健康检查：curl http://localhost:8080/health"
echo "  就绪检查：curl http://localhost:8080/ready"
echo ""
echo "📋 查看完整日志：docker-compose logs -f wisefido-data"
echo "🛑 停止服务：docker-compose stop wisefido-data"
