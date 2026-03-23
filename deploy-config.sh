#!/bin/bash
# 远程部署配置文件
# 使用前请修改以下配置为你的远程服务器信息

# ============================================
# 远程服务器配置
# ============================================
# 远程服务器地址（IP 或域名）
REMOTE_HOST="your-remote-host-ip-or-domain"

# 远程服务器 SSH 端口（默认 22）
REMOTE_PORT="22"

# 远程服务器用户名
REMOTE_USER="your-username"

# 远程服务器上的项目部署路径
REMOTE_DEPLOY_PATH="/home/${REMOTE_USER}/owl-project"

# SSH 密钥路径（如果使用密钥认证，留空则使用密码）
SSH_KEY_PATH=""

# ============================================
# 项目路径配置
# ============================================
# 本地项目根目录（自动检测，通常不需要修改）
LOCAL_OWLBACK_PATH="$(cd "$(dirname "$0")" && pwd)"
LOCAL_OWLFRONT_PATH="${LOCAL_OWLBACK_PATH%/owlBack}/owlFront"
LOCAL_OWLRD_PATH="${LOCAL_OWLBACK_PATH%/owlBack}/owlRD"

# ============================================
# 部署选项
# ============================================
# 是否部署前端（true/false）
DEPLOY_FRONTEND=true

# 是否部署后端（true/false）
DEPLOY_BACKEND=true

# 是否部署数据库初始化脚本（true/false）
DEPLOY_DB_SCRIPTS=true

# 是否在远程服务器构建（true）或本地构建后上传（false）
BUILD_ON_REMOTE=true

# 是否在部署后自动启动服务（true/false）
AUTO_START_SERVICES=true

# ============================================
# 远程服务器环境配置
# ============================================
# Go 版本（如果远程服务器需要安装 Go）
GO_VERSION="1.21"

# Node.js 版本（如果远程服务器需要安装 Node.js）
NODE_VERSION="20"

# ============================================
# 服务端口配置（远程服务器）
# ============================================
# 后端 API 端口
BACKEND_PORT="8080"

# 前端开发服务器端口
FRONTEND_PORT="3100"

# PostgreSQL 端口
POSTGRES_PORT="5432"

# Redis 端口
REDIS_PORT="6379"

# MQTT 端口
MQTT_PORT="1883"

# ============================================
# 数据库配置（远程服务器）
# ============================================
DB_HOST="localhost"
DB_PORT="5432"
DB_USER="postgres"
DB_PASSWORD="postgres"
DB_NAME="owlrd"

# ============================================
# Redis 配置（远程服务器）
# ============================================
REDIS_ADDR="127.0.0.1:6379"
REDIS_PASSWORD=""

# ============================================
# 其他配置
# ============================================
# 日志目录
LOG_DIR="$(cd "$(dirname "$0")/.." && pwd)/log"

# 可选；远程 owlBack .env 中业务租户 UUID（不设不影响服务启动）
TENANT_ID=""
