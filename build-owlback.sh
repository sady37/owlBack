#!/bin/bash
# 构建 owlback 各模块二进制到 <module>/.bin/ 下。
# 与 start-owlback-full.sh（systemd 编排）解耦：systemd 只拉起已有 .bin/*。
# 用法：./build-owlback.sh [module ...]
#   无参数：构建全部模块
#   带参数：只构建指定模块，如 ./build-owlback.sh qinglan ai

set -u

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR" || exit 1

if ! command -v go >/dev/null 2>&1; then
    echo -e "${RED}Error: go not found in PATH${NC}"
    exit 1
fi

# module_name  dir                         package
MODULES=(
    "data     wisefido-data      ./cmd/wisefido-data"
    "cardagg  wisefido-cardagg   ."
    "qinglan  wisefido-qinglan   ./cmd/wisefido-qinglan"
    "sleepace wisefido-sleepace  ./cmd/wisefido-sleepace"
    "iot      wisefido-iot       ./cmd/wisefido-iot"
    "ai       wisefido-ai        ./cmd/wisefido-ai"
)

build_one() {
    local name="$1" dir="$2" pkg="$3"
    local bin="wisefido-$name"
    local full_dir="$SCRIPT_DIR/$dir"

    if [ ! -d "$full_dir" ]; then
        echo -e "  ${YELLOW}SKIP $name${NC}  ($dir not found)"
        return 2
    fi

    mkdir -p "$full_dir/.bin"
    echo -n "  Building $bin ... "
    local out
    if out=$(cd "$full_dir" && go build -o ".bin/$bin" "$pkg" 2>&1); then
        echo -e "${GREEN}OK${NC}"
        return 0
    else
        echo -e "${RED}FAIL${NC}"
        echo "$out" | sed 's/^/    /'
        return 1
    fi
}

# 解析参数
SELECTED=()
if [ "$#" -gt 0 ]; then
    for arg in "$@"; do
        SELECTED+=("$arg")
    done
fi

is_selected() {
    local name="$1"
    [ "${#SELECTED[@]}" -eq 0 ] && return 0
    for s in "${SELECTED[@]}"; do
        [ "$s" = "$name" ] && return 0
    done
    return 1
}

echo "=== building owlback modules ==="
FAILED=()
BUILT=0
for entry in "${MODULES[@]}"; do
    read -r name dir pkg <<<"$entry"
    is_selected "$name" || continue
    if build_one "$name" "$dir" "$pkg"; then
        BUILT=$((BUILT + 1))
    else
        [ "$?" -eq 1 ] && FAILED+=("$name")
    fi
done

echo ""
if [ "${#FAILED[@]}" -eq 0 ]; then
    echo -e "${GREEN}all builds OK ($BUILT module(s))${NC}"
    echo "next: sudo systemctl restart owlback"
    exit 0
else
    echo -e "${RED}build failed: ${FAILED[*]}${NC}"
    exit 1
fi
