#!/bin/bash

# OwlBack 代码验证脚本
# 用途: 快速验证代码质量和编译状态

set -e  # 遇到错误立即退出

echo "=== OwlBack 代码验证 ==="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查函数
check() {
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ $1${NC}"
        return 0
    else
        echo -e "${RED}❌ $1${NC}"
        return 1
    fi
}

# 1. 代码格式
echo "1. 检查代码格式..."
go fmt ./... > /dev/null 2>&1
check "代码格式正确"

# 2. 代码规范
echo "2. 检查代码规范..."
go vet ./... 2>&1 | grep -v "vendor" || true
if [ ${PIPESTATUS[0]} -eq 0 ]; then
    echo -e "${GREEN}✅ 代码规范检查通过${NC}"
else
    echo -e "${YELLOW}⚠️  代码规范检查有警告（请查看上方输出）${NC}"
fi

# 3. 编译检查
echo ""
echo "3. 编译所有服务..."
services=("wisefido-radar" "wisefido-sleepace" "wisefido-data-transformer" "wisefido-sensor-fusion")
failed_services=()

for service in "${services[@]}"; do
    echo "  编译 $service..."
    cd "$service" 2>/dev/null || {
        echo -e "${YELLOW}⚠️  跳过 $service（目录不存在）${NC}"
        continue
    }
    
    if go build ./cmd/$service > /dev/null 2>&1; then
        echo -e "  ${GREEN}✅ $service 编译成功${NC}"
    else
        echo -e "  ${RED}❌ $service 编译失败${NC}"
        failed_services+=("$service")
    fi
    cd ..
done

if [ ${#failed_services[@]} -gt 0 ]; then
    echo -e "${RED}❌ 以下服务编译失败: ${failed_services[*]}${NC}"
    exit 1
fi

# 4. 依赖检查
echo ""
echo "4. 检查依赖..."
if go mod verify > /dev/null 2>&1; then
    echo -e "${GREEN}✅ 依赖验证通过${NC}"
else
    echo -e "${YELLOW}⚠️  依赖验证有警告${NC}"
fi

# 5. 统计信息
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

