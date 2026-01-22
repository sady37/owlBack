# 远程部署指南

本文档说明如何将 owlBack、owlFront 和 owlRD 项目部署到远程 Linux 测试开发机。

## 目录

- [前置要求](#前置要求)
- [快速开始](#快速开始)
- [详细步骤](#详细步骤)
- [配置说明](#配置说明)
- [服务管理](#服务管理)
- [故障排查](#故障排查)

## 前置要求

### 本地机器要求

- macOS 或 Linux 系统
- SSH 客户端
- rsync 工具（通常已预装）
- 已配置 SSH 密钥或密码认证到远程服务器

### 远程服务器要求

- Linux 系统（Ubuntu/Debian/CentOS/RHEL）
- 至少 4GB RAM
- 至少 20GB 可用磁盘空间
- 网络连接正常
- 开放以下端口：
  - 22 (SSH)
  - 8080 (后端 API)
  - 3100 (前端开发服务器，可选)
  - 5432 (PostgreSQL)
  - 6379 (Redis)
  - 1883 (MQTT)

## 快速开始

### 1. 配置部署参数

复制并编辑配置文件：

```bash
cd /Users/sady3721/project/owlBack
cp deploy-config.sh deploy-config.sh.local
# 编辑 deploy-config.sh.local，修改远程服务器信息
```

编辑 `deploy-config.sh.local`，至少修改以下配置：

```bash
REMOTE_HOST="your-remote-server-ip"
REMOTE_USER="your-username"
REMOTE_DEPLOY_PATH="/home/your-username/owl-project"
```

如果使用 SSH 密钥认证：

```bash
SSH_KEY_PATH="/path/to/your/ssh/key"
```

### 2. 执行部署

```bash
# 使用本地配置文件
source deploy-config.sh.local
bash deploy.sh

# 或者直接修改 deploy-config.sh
bash deploy.sh
```

### 3. 远程服务器初始化（首次部署）

首次部署时，需要在远程服务器上运行初始化脚本：

```bash
# 方式1: 通过部署脚本自动执行
bash deploy.sh --setup

# 方式2: 手动 SSH 到远程服务器执行
ssh your-username@your-remote-server
cd /home/your-username/owl-project/owlBack
bash remote-setup.sh
```

## 详细步骤

### 步骤 1: 配置 SSH 连接

#### 使用 SSH 密钥（推荐）

1. 生成 SSH 密钥（如果还没有）：

```bash
ssh-keygen -t rsa -b 4096 -C "your-email@example.com"
```

2. 将公钥复制到远程服务器：

```bash
ssh-copy-id -p 22 your-username@your-remote-server
```

3. 测试连接：

```bash
ssh -p 22 your-username@your-remote-server
```

#### 使用密码认证

确保 SSH 密码认证已启用（通常默认启用）。

### 步骤 2: 配置部署参数

编辑 `deploy-config.sh`：

```bash
# 远程服务器配置
REMOTE_HOST="192.168.1.100"           # 远程服务器 IP 或域名
REMOTE_PORT="22"                       # SSH 端口
REMOTE_USER="developer"                # 远程用户名
REMOTE_DEPLOY_PATH="/home/developer/owl-project"  # 部署路径

# SSH 密钥路径（如果使用密钥认证）
SSH_KEY_PATH="~/.ssh/id_rsa"

# 部署选项
DEPLOY_FRONTEND=true                   # 是否部署前端
DEPLOY_BACKEND=true                    # 是否部署后端
DEPLOY_DB_SCRIPTS=true                 # 是否部署数据库脚本
BUILD_ON_REMOTE=true                   # 是否在远程构建
AUTO_START_SERVICES=true               # 是否自动启动服务
```

### 步骤 3: 执行部署

```bash
cd /Users/sady3721/project/owlBack

# 基本部署（只同步代码）
bash deploy.sh

# 部署并执行远程初始化（首次部署）
bash deploy.sh --setup
```

部署脚本会：

1. 测试 SSH 连接
2. 创建远程目录结构
3. 同步代码到远程服务器
4. （可选）在远程服务器安装依赖
5. （可选）构建前端
6. （可选）启动服务

### 步骤 4: 远程服务器初始化（首次）

如果使用 `--setup` 选项，脚本会自动执行初始化。否则，手动执行：

```bash
ssh your-username@your-remote-server
cd /home/your-username/owl-project/owlBack
bash remote-setup.sh
```

初始化脚本会：

1. 检测操作系统
2. 安装 Docker 和 Docker Compose
3. 安装 Go（如果未安装）
4. 安装 Node.js（如果需要部署前端）
5. 创建必要的目录
6. 配置 MQTT

### 步骤 5: 启动服务

#### 启动基础设施（Docker Compose）

```bash
cd /home/your-username/owl-project/owlBack
docker-compose up -d
```

这会启动：
- PostgreSQL (端口 5432)
- Redis (端口 6379)
- MQTT (端口 1883)

#### 初始化数据库（首次）

```bash
# 等待 PostgreSQL 启动
sleep 10

# 运行数据库初始化脚本
bash scripts/init-db.sh
```

或者使用 remote-setup.sh：

```bash
bash remote-setup.sh --init-db
```

#### 启动后端服务

```bash
cd /home/your-username/owl-project/owlBack
bash start_all_services.sh
```

这会启动：
- wisefido-data (端口 8080)
- wisefido-card-aggregator (后台服务)

#### 启动前端（开发模式）

```bash
cd /home/your-username/owl-project/owlFront
npm install  # 首次需要
npm run dev  # 开发模式，端口 3100
```

或者构建生产版本：

```bash
npm run build
# 使用 nginx 或其他 web 服务器部署 dist 目录
```

## 配置说明

### 环境变量

后端服务使用环境变量进行配置。可以在 `start_all_services.sh` 中设置，或创建 `.env` 文件。

主要环境变量：

```bash
# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=owlrd

# Redis 配置
REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=

# 服务配置
HTTP_ADDR=:8080
LOG_LEVEL=info
LOG_FORMAT=json

# 卡片聚合配置
TENANT_ID=bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c
CARD_TRIGGER_MODE=polling
CARD_POLLING_INTERVAL=86400
```

### 前端配置

编辑 `owlFront/vite.config.ts` 修改 API 地址：

```typescript
process.env.VITE_API_BASE_URL = 'http://your-remote-server-ip:8080'
```

或者使用环境变量文件 `.env`：

```bash
VITE_API_BASE_URL=http://your-remote-server-ip:8080
```

## 服务管理

### 查看服务状态

```bash
# 查看 Docker 服务
docker-compose ps

# 查看后端服务进程
ps aux | grep wisefido

# 查看端口占用
netstat -tlnp | grep -E '8080|5432|6379|1883'
```

### 查看日志

```bash
# 后端服务日志
tail -f /tmp/owlBack_logs/wisefido-data.log
tail -f /tmp/owlBack_logs/wisefido-card-aggregator.log

# Docker 服务日志
docker-compose logs -f postgresql
docker-compose logs -f redis
docker-compose logs -f mqtt
```

### 停止服务

```bash
# 停止后端服务
cd /home/your-username/owl-project/owlBack
bash stop_all_services.sh

# 停止基础设施
docker-compose down
```

### 重启服务

```bash
# 重启后端服务
bash stop_all_services.sh
bash start_all_services.sh

# 重启基础设施
docker-compose restart
```

## 故障排查

### SSH 连接失败

**问题**: 无法连接到远程服务器

**解决方案**:
1. 检查网络连接
2. 检查 SSH 端口是否正确
3. 检查防火墙设置
4. 验证 SSH 密钥权限：`chmod 600 ~/.ssh/id_rsa`

### 部署失败

**问题**: rsync 同步失败

**解决方案**:
1. 检查远程目录权限
2. 检查磁盘空间：`df -h`
3. 检查 SSH 连接：`ssh -v your-username@your-remote-server`

### 服务启动失败

**问题**: 后端服务无法启动

**解决方案**:
1. 检查端口是否被占用：`lsof -i :8080`
2. 检查数据库连接：`psql -h localhost -U postgres -d owlrd`
3. 检查 Redis 连接：`redis-cli ping`
4. 查看详细日志：`tail -f /tmp/owlBack_logs/wisefido-data.log`

### Docker 服务启动失败

**问题**: Docker Compose 服务无法启动

**解决方案**:
1. 检查 Docker 是否运行：`sudo systemctl status docker`
2. 检查端口冲突：`netstat -tlnp | grep -E '5432|6379|1883'`
3. 查看 Docker 日志：`docker-compose logs`
4. 检查磁盘空间：`df -h`

### 数据库连接失败

**问题**: 无法连接到 PostgreSQL

**解决方案**:
1. 检查 PostgreSQL 是否运行：`docker-compose ps postgresql`
2. 检查数据库配置是否正确
3. 检查防火墙是否阻止连接
4. 查看 PostgreSQL 日志：`docker-compose logs postgresql`

### 前端构建失败

**问题**: npm install 或 npm run build 失败

**解决方案**:
1. 检查 Node.js 版本：`node --version`（需要 >= 18）
2. 清理缓存：`rm -rf node_modules package-lock.json && npm install`
3. 检查网络连接（npm 需要访问外网）
4. 使用国内镜像：`npm config set registry https://registry.npmmirror.com`

## 更新部署

当代码更新后，重新运行部署脚本：

```bash
cd /Users/sady3721/project/owlBack
bash deploy.sh
```

如果需要重启服务：

```bash
ssh your-username@your-remote-server
cd /home/your-username/owl-project/owlBack
bash stop_all_services.sh
bash start_all_services.sh
```

## 安全建议

1. **使用 SSH 密钥认证**，避免使用密码
2. **配置防火墙**，只开放必要的端口
3. **使用强密码**，特别是数据库密码
4. **定期更新系统**和依赖包
5. **限制 SSH 访问**，使用 fail2ban 等工具
6. **使用 HTTPS**（生产环境）
7. **定期备份数据库**

## 生产环境部署

生产环境部署建议：

1. 使用反向代理（Nginx）处理 HTTPS
2. 使用 systemd 管理服务（而不是直接运行脚本）
3. 配置日志轮转
4. 设置监控和告警
5. 使用 CI/CD 自动化部署
6. 配置数据库备份策略

## 联系支持

如果遇到问题，请检查：

1. 日志文件：`/tmp/owlBack_logs/`
2. Docker 日志：`docker-compose logs`
3. 系统日志：`journalctl -u docker`（如果使用 systemd）
