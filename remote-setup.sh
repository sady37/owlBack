#!/bin/bash
# 远程服务器初始化脚本
# 在远程 Linux 测试开发机上执行，用于安装依赖和配置环境

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Remote Server Setup${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DEPLOY_PATH="$(dirname "$SCRIPT_DIR")"

# 加载配置文件（如果存在）
CONFIG_FILE="${SCRIPT_DIR}/deploy-config.sh"
if [ -f "$CONFIG_FILE" ]; then
    source "$CONFIG_FILE"
fi

# 检测操作系统
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
    VER=$VERSION_ID
else
    echo -e "${RED}Error: Cannot detect operating system${NC}"
    exit 1
fi

echo -e "${BLUE}Detected OS: $OS $VER${NC}"
echo ""

# 检查并安装 Docker
if ! command -v docker &> /dev/null; then
    echo -e "${YELLOW}Docker not found. Installing Docker...${NC}"
    if [ "$OS" = "ubuntu" ] || [ "$OS" = "debian" ]; then
        sudo apt-get update
        sudo apt-get install -y docker.io docker-compose
        sudo systemctl enable docker
        sudo systemctl start docker
        # 将当前用户添加到 docker 组（避免每次都用 sudo）
        sudo usermod -aG docker $USER
        echo -e "${GREEN}Docker installed. Please logout and login again for group changes to take effect.${NC}"
    elif [ "$OS" = "centos" ] || [ "$OS" = "rhel" ] || [ "$OS" = "fedora" ]; then
        sudo yum install -y docker docker-compose
        sudo systemctl enable docker
        sudo systemctl start docker
        sudo usermod -aG docker $USER
        echo -e "${GREEN}Docker installed. Please logout and login again for group changes to take effect.${NC}"
    else
        echo -e "${RED}Error: Unsupported OS for automatic Docker installation${NC}"
        echo "Please install Docker manually: https://docs.docker.com/get-docker/"
        exit 1
    fi
else
    echo -e "${GREEN}Docker is already installed${NC}"
fi

# 检查 Docker Compose
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo -e "${YELLOW}Docker Compose not found. Installing...${NC}"
    if [ "$OS" = "ubuntu" ] || [ "$OS" = "debian" ]; then
        sudo apt-get install -y docker-compose
    elif [ "$OS" = "centos" ] || [ "$OS" = "rhel" ] || [ "$OS" = "fedora" ]; then
        sudo yum install -y docker-compose
    fi
else
    echo -e "${GREEN}Docker Compose is already installed${NC}"
fi

echo ""

# 检查并安装 Go
if ! command -v go &> /dev/null; then
    echo -e "${YELLOW}Go not found. Installing Go ${GO_VERSION:-1.21}...${NC}"
    
    GO_VERSION="${GO_VERSION:-1.21}"
    GO_ARCH="linux-amd64"
    GO_TAR="go${GO_VERSION}.${GO_ARCH}.tar.gz"
    GO_URL="https://go.dev/dl/${GO_TAR}"
    
    cd /tmp
    wget "$GO_URL"
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf "$GO_TAR"
    rm "$GO_TAR"
    
    # 添加到 PATH（如果还没有）
    if ! grep -q "/usr/local/go/bin" ~/.bashrc 2>/dev/null; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
        export PATH=$PATH:/usr/local/go/bin
    fi
    
    echo -e "${GREEN}Go ${GO_VERSION} installed${NC}"
else
    INSTALLED_GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    echo -e "${GREEN}Go is already installed (version: $INSTALLED_GO_VERSION)${NC}"
fi

echo ""

# 检查并安装 Node.js（如果需要部署前端）
if [ "$DEPLOY_FRONTEND" != "false" ]; then
    if ! command -v node &> /dev/null; then
        echo -e "${YELLOW}Node.js not found. Installing Node.js ${NODE_VERSION:-20}...${NC}"
        
        NODE_VERSION="${NODE_VERSION:-20}"
        
        if [ "$OS" = "ubuntu" ] || [ "$OS" = "debian" ]; then
            curl -fsSL https://deb.nodesource.com/setup_${NODE_VERSION}.x | sudo -E bash -
            sudo apt-get install -y nodejs
        elif [ "$OS" = "centos" ] || [ "$OS" = "rhel" ] || [ "$OS" = "fedora" ]; then
            curl -fsSL https://rpm.nodesource.com/setup_${NODE_VERSION}.x | sudo bash -
            sudo yum install -y nodejs
        else
            echo -e "${YELLOW}Please install Node.js manually from https://nodejs.org/${NC}"
        fi
        
        echo -e "${GREEN}Node.js installed${NC}"
    else
        INSTALLED_NODE_VERSION=$(node --version)
        echo -e "${GREEN}Node.js is already installed (version: $INSTALLED_NODE_VERSION)${NC}"
    fi
    echo ""
fi

# 创建必要的目录
echo -e "${BLUE}Creating necessary directories...${NC}"
OWL_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
mkdir -p "${LOG_DIR:-$OWL_ROOT/log}"
mkdir -p "${DEPLOY_PATH}/owlBack/mqtt/{config,data,log}"
echo -e "${GREEN}Directories created${NC}"
echo ""

# 配置 MQTT（如果需要）
if [ -d "${SCRIPT_DIR}/mqtt" ]; then
    echo -e "${BLUE}Setting up MQTT configuration...${NC}"
    if [ ! -f "${SCRIPT_DIR}/mqtt/config/mosquitto.conf" ]; then
        echo -e "${YELLOW}Creating default MQTT configuration...${NC}"
        mkdir -p "${SCRIPT_DIR}/mqtt/config"
        cat > "${SCRIPT_DIR}/mqtt/config/mosquitto.conf" << 'EOF'
listener 1883
allow_anonymous true
persistence true
persistence_location /mosquitto/data/
log_dest file /mosquitto/log/mosquitto.log
EOF
    fi
    echo -e "${GREEN}MQTT configuration ready${NC}"
    echo ""
fi

# 初始化数据库（如果需要）
if [ "$1" = "--init-db" ]; then
    echo -e "${BLUE}Initializing database...${NC}"
    
    # 等待 PostgreSQL 启动
    echo "Waiting for PostgreSQL to be ready..."
    sleep 5
    
    # 运行数据库初始化脚本
    if [ -f "${SCRIPT_DIR}/scripts/init-db.sh" ]; then
        bash "${SCRIPT_DIR}/scripts/init-db.sh"
    else
        echo -e "${YELLOW}Warning: init-db.sh not found${NC}"
        echo "Database initialization skipped"
    fi
    
    echo -e "${GREEN}Database initialization completed${NC}"
    echo ""
fi

# 显示环境信息
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Environment Information${NC}"
echo -e "${GREEN}========================================${NC}"
echo "Go version: $(go version 2>/dev/null || echo 'Not installed')"
if command -v node &> /dev/null; then
    echo "Node.js version: $(node --version)"
    echo "npm version: $(npm --version)"
fi
echo "Docker version: $(docker --version 2>/dev/null || echo 'Not installed')"
echo "Docker Compose version: $(docker-compose --version 2>/dev/null || docker compose version 2>/dev/null || echo 'Not installed')"
echo ""
echo "Deployment path: $DEPLOY_PATH"
echo "Log directory: ${LOG_DIR:-$OWL_ROOT/log}"
echo ""

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Setup Completed!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Next steps:"
echo "  1. Start infrastructure: cd ${SCRIPT_DIR} && docker-compose up -d"
echo "  2. Wait for services to be ready: sleep 10"
echo "  3. Initialize database (if needed): bash remote-setup.sh --init-db"
echo "  4. Start backend services: bash start_all_services.sh"
echo ""
