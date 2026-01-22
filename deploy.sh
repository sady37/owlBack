#!/bin/bash
# 远程部署脚本
# 功能：将 owlBack、owlFront 和 owlRD 部署到远程 Linux 测试开发机

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 加载配置文件
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CONFIG_FILE="${SCRIPT_DIR}/deploy-config.sh"

if [ ! -f "$CONFIG_FILE" ]; then
    echo -e "${RED}Error: Configuration file not found: $CONFIG_FILE${NC}"
    echo "Please create deploy-config.sh first (you can copy deploy-config.sh.example)"
    exit 1
fi

source "$CONFIG_FILE"

# 检查配置
if [ "$REMOTE_HOST" = "your-remote-host-ip-or-domain" ] || [ -z "$REMOTE_HOST" ]; then
    echo -e "${RED}Error: Please configure REMOTE_HOST in deploy-config.sh${NC}"
    exit 1
fi

if [ "$REMOTE_USER" = "your-username" ] || [ -z "$REMOTE_USER" ]; then
    echo -e "${RED}Error: Please configure REMOTE_USER in deploy-config.sh${NC}"
    exit 1
fi

# SSH 命令构建
SSH_CMD="ssh"
if [ -n "$SSH_KEY_PATH" ] && [ -f "$SSH_KEY_PATH" ]; then
    SSH_CMD="$SSH_CMD -i $SSH_KEY_PATH"
fi
SSH_CMD="$SSH_CMD -p $REMOTE_PORT ${REMOTE_USER}@${REMOTE_HOST}"

# SCP 命令构建
SCP_CMD="scp"
if [ -n "$SSH_KEY_PATH" ] && [ -f "$SSH_KEY_PATH" ]; then
    SCP_CMD="$SCP_CMD -i $SSH_KEY_PATH"
fi
SCP_CMD="$SCP_CMD -P $REMOTE_PORT"

# 远程执行命令函数
remote_exec() {
    $SSH_CMD "$@"
}

# 显示配置信息
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Remote Deployment Configuration${NC}"
echo -e "${GREEN}========================================${NC}"
echo "Remote Host: $REMOTE_HOST:$REMOTE_PORT"
echo "Remote User: $REMOTE_USER"
echo "Remote Path: $REMOTE_DEPLOY_PATH"
echo "Deploy Frontend: $DEPLOY_FRONTEND"
echo "Deploy Backend: $DEPLOY_BACKEND"
echo "Deploy DB Scripts: $DEPLOY_DB_SCRIPTS"
echo "Build on Remote: $BUILD_ON_REMOTE"
echo "Auto Start Services: $AUTO_START_SERVICES"
echo ""

# 测试 SSH 连接
echo -e "${BLUE}Testing SSH connection...${NC}"
if ! remote_exec "echo 'SSH connection successful'" > /dev/null 2>&1; then
    echo -e "${RED}Error: Cannot connect to remote server${NC}"
    echo "Please check:"
    echo "  1. SSH connection: ssh -p $REMOTE_PORT ${REMOTE_USER}@${REMOTE_HOST}"
    echo "  2. SSH key or password authentication"
    exit 1
fi
echo -e "${GREEN}SSH connection OK${NC}"
echo ""

# 创建远程目录结构
echo -e "${BLUE}Creating remote directory structure...${NC}"
remote_exec "mkdir -p $REMOTE_DEPLOY_PATH/{owlBack,owlFront,owlRD}"
echo -e "${GREEN}Remote directories created${NC}"
echo ""

# 部署后端（owlBack）
if [ "$DEPLOY_BACKEND" = "true" ]; then
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}Deploying Backend (owlBack)${NC}"
    echo -e "${GREEN}========================================${NC}"
    
    # 同步代码（排除不必要的文件）
    echo -e "${BLUE}Syncing owlBack code...${NC}"
    rsync -avz --delete \
        --exclude='.git' \
        --exclude='*.log' \
        --exclude='/tmp' \
        --exclude='node_modules' \
        --exclude='.vscode' \
        --exclude='*.swp' \
        --exclude='*.swo' \
        -e "ssh -p $REMOTE_PORT ${SSH_KEY_PATH:+-i $SSH_KEY_PATH}" \
        "$LOCAL_OWLBACK_PATH/" \
        "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DEPLOY_PATH}/owlBack/"
    echo -e "${GREEN}owlBack code synced${NC}"
    
    # 上传部署配置（如果需要）
    if [ -f "$CONFIG_FILE" ]; then
        echo -e "${BLUE}Uploading deployment config...${NC}"
        $SCP_CMD "$CONFIG_FILE" "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DEPLOY_PATH}/owlBack/deploy-config.sh"
    fi
    
    # 上传远程初始化脚本
    if [ -f "${SCRIPT_DIR}/remote-setup.sh" ]; then
        echo -e "${BLUE}Uploading remote setup script...${NC}"
        $SCP_CMD "${SCRIPT_DIR}/remote-setup.sh" "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DEPLOY_PATH}/owlBack/remote-setup.sh"
        remote_exec "chmod +x ${REMOTE_DEPLOY_PATH}/owlBack/remote-setup.sh"
    fi
    
    echo -e "${GREEN}Backend deployment completed${NC}"
    echo ""
fi

# 部署前端（owlFront）
if [ "$DEPLOY_FRONTEND" = "true" ]; then
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}Deploying Frontend (owlFront)${NC}"
    echo -e "${GREEN}========================================${NC}"
    
    if [ ! -d "$LOCAL_OWLFRONT_PATH" ]; then
        echo -e "${YELLOW}Warning: owlFront directory not found at $LOCAL_OWLFRONT_PATH${NC}"
        echo "Skipping frontend deployment"
    else
        # 同步代码
        echo -e "${BLUE}Syncing owlFront code...${NC}"
        rsync -avz --delete \
            --exclude='.git' \
            --exclude='node_modules' \
            --exclude='dist' \
            --exclude='.vscode' \
            --exclude='*.swp' \
            --exclude='*.swo' \
            -e "ssh -p $REMOTE_PORT ${SSH_KEY_PATH:+-i $SSH_KEY_PATH}" \
            "$LOCAL_OWLFRONT_PATH/" \
            "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DEPLOY_PATH}/owlFront/"
        echo -e "${GREEN}owlFront code synced${NC}"
        
        # 在远程服务器安装依赖和构建
        if [ "$BUILD_ON_REMOTE" = "true" ]; then
            echo -e "${BLUE}Installing frontend dependencies on remote server...${NC}"
            remote_exec "cd ${REMOTE_DEPLOY_PATH}/owlFront && npm install"
            echo -e "${BLUE}Building frontend on remote server...${NC}"
            remote_exec "cd ${REMOTE_DEPLOY_PATH}/owlFront && npm run build"
            echo -e "${GREEN}Frontend built on remote server${NC}"
        fi
    fi
    
    echo -e "${GREEN}Frontend deployment completed${NC}"
    echo ""
fi

# 部署数据库脚本（owlRD）
if [ "$DEPLOY_DB_SCRIPTS" = "true" ]; then
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}Deploying Database Scripts (owlRD)${NC}"
    echo -e "${GREEN}========================================${NC}"
    
    if [ ! -d "$LOCAL_OWLRD_PATH" ]; then
        echo -e "${YELLOW}Warning: owlRD directory not found at $LOCAL_OWLRD_PATH${NC}"
        echo "Skipping database scripts deployment"
    else
        # 只同步 db 目录
        echo -e "${BLUE}Syncing database scripts...${NC}"
        remote_exec "mkdir -p ${REMOTE_DEPLOY_PATH}/owlRD/db"
        rsync -avz \
            --exclude='.git' \
            -e "ssh -p $REMOTE_PORT ${SSH_KEY_PATH:+-i $SSH_KEY_PATH}" \
            "$LOCAL_OWLRD_PATH/db/" \
            "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DEPLOY_PATH}/owlRD/db/"
        echo -e "${GREEN}Database scripts synced${NC}"
    fi
    
    echo -e "${GREEN}Database scripts deployment completed${NC}"
    echo ""
fi

# 在远程服务器执行初始化（如果需要）
if [ "$1" = "--setup" ]; then
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}Running Remote Setup${NC}"
    echo -e "${GREEN}========================================${NC}"
    
    if [ -f "${SCRIPT_DIR}/remote-setup.sh" ]; then
        echo -e "${BLUE}Executing remote setup script...${NC}"
        remote_exec "cd ${REMOTE_DEPLOY_PATH}/owlBack && bash remote-setup.sh"
        echo -e "${GREEN}Remote setup completed${NC}"
    else
        echo -e "${YELLOW}Warning: remote-setup.sh not found${NC}"
    fi
    echo ""
fi

# 自动启动服务
if [ "$AUTO_START_SERVICES" = "true" ] && [ "$DEPLOY_BACKEND" = "true" ]; then
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}Starting Services on Remote Server${NC}"
    echo -e "${GREEN}========================================${NC}"
    
    echo -e "${BLUE}Starting infrastructure services (Docker Compose)...${NC}"
    remote_exec "cd ${REMOTE_DEPLOY_PATH}/owlBack && docker-compose up -d"
    
    echo -e "${BLUE}Waiting for services to be ready...${NC}"
    sleep 5
    
    echo -e "${BLUE}Starting backend services...${NC}"
    remote_exec "cd ${REMOTE_DEPLOY_PATH}/owlBack && bash start_all_services.sh" &
    
    echo -e "${GREEN}Services started${NC}"
    echo ""
    echo -e "${BLUE}Note: Services are running in background${NC}"
    echo "To check logs, SSH to the server and run:"
    echo "  tail -f ${LOG_DIR}/wisefido-data.log"
    echo "  tail -f ${LOG_DIR}/wisefido-card-aggregator.log"
    echo ""
fi

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Deployment Completed!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Remote deployment path: $REMOTE_DEPLOY_PATH"
echo ""
echo "Next steps:"
echo "  1. SSH to remote server: ssh -p $REMOTE_PORT ${REMOTE_USER}@${REMOTE_HOST}"
echo "  2. If not auto-started, run setup: cd ${REMOTE_DEPLOY_PATH}/owlBack && bash remote-setup.sh"
echo "  3. Start services: cd ${REMOTE_DEPLOY_PATH}/owlBack && bash start_all_services.sh"
echo "  4. Check logs: tail -f ${LOG_DIR}/wisefido-data.log"
echo ""
