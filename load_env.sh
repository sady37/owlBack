#!/bin/bash

# 加载 .env 配置文件的辅助函数
# 用法：source load_env.sh 或 . load_env.sh

# 获取脚本所在目录（owlBack 根目录）
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/.env"

# 如果 .env 文件存在，加载它
if [ -f "$ENV_FILE" ]; then
    echo "📋 Loading configuration from: $ENV_FILE"
    # 读取 .env 文件，忽略注释和空行，并导出变量
    set -a  # 自动导出所有变量
    source "$ENV_FILE"
    set +a  # 关闭自动导出
    echo "✅ Configuration loaded"
else
    echo "⚠️  Warning: .env file not found at $ENV_FILE"
    echo "   Using default values or environment variables"
fi
