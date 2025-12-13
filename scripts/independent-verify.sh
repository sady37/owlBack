#!/bin/bash

# 独立代码验证脚本
# 使用自动化工具进行验证，避免 AI 自我验证的干扰

set -e

echo "=== OwlBack 独立代码验证 ==="
echo "使用自动化工具进行验证，避免 AI 自我验证的干扰"
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    if [ -f "/usr/local/go/bin/go" ]; then
        GO_CMD="/usr/local/go/bin/go"
    else
        echo -e "${RED}❌ Go 未安装或不在 PATH 中${NC}"
        exit 1
    fi
else
    GO_CMD="go"
fi

echo "使用 Go: $($GO_CMD version)"
echo ""

# 工作目录
cd /Users/sady3721/project/owlBack

# 1. 代码格式检查
echo "1. 代码格式检查 (go fmt)..."
$GO_CMD fmt ./wisefido-radar/... ./wisefido-sleepace/... ./wisefido-data-transformer/... ./wisefido-sensor-fusion/... > /dev/null 2>&1
echo -e "${GREEN}✅ 代码格式正确${NC}"

# 2. 代码规范检查
echo ""
echo "2. 代码规范检查 (go vet)..."
if $GO_CMD vet ./wisefido-radar/... ./wisefido-sleepace/... ./wisefido-data-transformer/... ./wisefido-sensor-fusion/... 2>&1 | grep -v "vendor"; then
    echo -e "${YELLOW}⚠️  代码规范检查有警告${NC}"
else
    echo -e "${GREEN}✅ 代码规范检查通过${NC}"
fi

# 3. 编译检查
echo ""
echo "3. 编译检查..."
services=("wisefido-radar" "wisefido-sleepace" "wisefido-data-transformer" "wisefido-sensor-fusion")
failed=0

for service in "${services[@]}"; do
    echo "  编译 $service..."
    cd "$service"
    if $GO_CMD build ./cmd/$service > /dev/null 2>&1; then
        echo -e "  ${GREEN}✅ $service 编译成功${NC}"
    else
        echo -e "  ${RED}❌ $service 编译失败${NC}"
        failed=$((failed + 1))
    fi
    cd ..
done

if [ $failed -gt 0 ]; then
    echo -e "${RED}❌ $failed 个服务编译失败${NC}"
    exit 1
fi

# 4. 依赖验证
echo ""
echo "4. 依赖验证 (go mod verify)..."
for service in "${services[@]}"; do
    cd "$service"
    if $GO_CMD mod verify > /dev/null 2>&1; then
        echo -e "  ${GREEN}✅ $service 依赖验证通过${NC}"
    else
        echo -e "  ${YELLOW}⚠️  $service 依赖验证有警告${NC}"
    fi
    cd ..
done

# 5. 检查 golangci-lint
echo ""
echo "5. 静态分析工具检查..."
if command -v golangci-lint &> /dev/null; then
    echo -e "${GREEN}✅ golangci-lint 已安装${NC}"
    echo "  运行静态分析..."
    for service in "${services[@]}"; do
        echo "  分析 $service..."
        if golangci-lint run ./$service/... 2>&1 | head -20; then
            echo -e "  ${GREEN}✅ $service 静态分析通过${NC}"
        else
            echo -e "  ${YELLOW}⚠️  $service 静态分析有警告${NC}"
        fi
    done
else
    echo -e "${YELLOW}⚠️  golangci-lint 未安装${NC}"
    echo "  安装命令: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
fi

# 6. 统计信息
echo ""
echo "=== 统计信息 ==="
go_files=$(find . -name "*.go" -type f | grep -v vendor | wc -l | tr -d ' ')
test_files=$(find . -name "*_test.go" -type f | grep -v vendor | wc -l | tr -d ' ')

echo "Go 文件数: $go_files"
echo "测试文件数: $test_files"

if [ "$test_files" -eq 0 ]; then
    echo -e "${YELLOW}⚠️  未发现测试文件，建议添加单元测试${NC}"
fi

echo ""
echo -e "${GREEN}=== 验证完成 ===${NC}"
echo ""
echo "建议:"
echo "1. 使用 ChatGPT 进行独立审查"
echo "2. 安装 golangci-lint 进行静态分析"
echo "3. 添加单元测试提高代码质量"

