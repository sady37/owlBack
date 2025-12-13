#!/bin/bash

# wisefido-alarm 单元测试运行脚本

set -e

echo "========================================="
echo "wisefido-alarm 单元测试"
echo "========================================="
echo ""

cd "$(dirname "$0")/.."

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 运行所有测试
echo "运行所有测试..."
go test ./... -v

echo ""
echo "========================================="
echo "测试覆盖率"
echo "========================================="
echo ""

go test ./... -cover

echo ""
echo "========================================="
echo "生成覆盖率报告"
echo "========================================="
echo ""

go test ./... -coverprofile=coverage.out
echo -e "${GREEN}✓${NC} 覆盖率报告已生成: coverage.out"
echo ""
echo "查看 HTML 报告:"
echo "  go tool cover -html=coverage.out"

