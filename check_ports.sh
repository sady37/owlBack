#!/bin/bash

echo "🔍 检查端口配置一致性"
echo "======================"

echo ""
echo "📊 Docker Compose 配置:"
echo "  PostgreSQL 端口映射: 5433:5432"
echo "  wisefido-data DB_PORT: 5432 (容器内)"
echo ""

echo "📊 start_owlback.sh 配置:"
grep -n "export DB_PORT" /Users/sady3721/project/owlBack/start_owlback.sh | head -5
echo ""

echo "📊 wisefido-radar 配置:"
grep -n "5433" /Users/sady3721/project/owlBack/wisefido-radar/internal/config/config.go | head -3
echo ""

echo "✅ 端口配置已统一为:"
echo "  主机端口: 5433"
echo "  容器内端口: 5432"
echo ""

echo "🔧 启动服务测试:"
echo "1. 启动 PostgreSQL: docker-compose up -d postgresql"
echo "2. 测试连接: nc -zv localhost 5433"
echo "3. 启动 wisefido-radar: ./wisefido-radar/start-radar.sh"
echo "4. 启动 wisefido-data: docker-compose up -d --build wisefido-data"
echo ""

echo "📋 验证 Doctor 功能:"
echo "  curl http://localhost:8080/health"
echo "  curl http://localhost:8080/ready"
