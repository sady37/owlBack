#!/bin/bash

# wisefido-qinglan 启动脚本

echo "Starting wisefido-qinglan..."

# 检查配置文件
if [ ! -f "config.yaml" ]; then
    echo "Error: config.yaml not found"
    echo "Creating default config..."
    cp config.yaml.example config.yaml 2>/dev/null || echo "Please create config.yaml first"
    exit 1
fi

# 检查go.work
if [ ! -f "go.work" ]; then
    echo "Error: go.work not found"
    echo "Creating go.work..."
    cat > go.work << EOF
go 1.21

use (
    .
    ../owl-common
)

replace owl-common => ../owl-common
EOF
fi

# 下载依赖
echo "Downloading dependencies..."
go mod download

# 启动服务
echo "Starting service..."
go run cmd/wisefido-qinglan/main.go