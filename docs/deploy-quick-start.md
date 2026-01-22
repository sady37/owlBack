# 快速部署指南

## 三步完成部署

### 1. 配置远程服务器信息

编辑 `deploy-config.sh`，修改以下配置：

```bash
REMOTE_HOST="your-remote-server-ip"      # 改为你的远程服务器 IP
REMOTE_USER="your-username"               # 改为你的远程用户名
REMOTE_DEPLOY_PATH="/home/your-username/owl-project"  # 部署路径
```

如果使用 SSH 密钥：

```bash
SSH_KEY_PATH="~/.ssh/id_rsa"              # 你的 SSH 密钥路径
```

### 2. 执行部署

```bash
cd /Users/sady3721/project/owlBack

# 首次部署（包含环境初始化）
bash deploy.sh --setup

# 后续更新（只同步代码）
bash deploy.sh
```

### 3. 启动服务（如果未自动启动）

SSH 到远程服务器：

```bash
ssh your-username@your-remote-server
cd /home/your-username/owl-project/owlBack

# 启动基础设施（PostgreSQL, Redis, MQTT）
docker-compose up -d

# 等待服务就绪
sleep 10

# 初始化数据库（首次需要）
bash remote-setup.sh --init-db

# 启动后端服务
bash start_all_services.sh
```

## 常用命令

### 查看服务状态

```bash
# 查看 Docker 服务
docker-compose ps

# 查看后端服务进程
ps aux | grep wisefido
```

### 查看日志

```bash
# 后端日志
tail -f /tmp/owlBack_logs/wisefido-data.log
tail -f /tmp/owlBack_logs/wisefido-card-aggregator.log

# Docker 日志
docker-compose logs -f
```

### 停止服务

```bash
# 停止后端
bash stop_all_services.sh

# 停止基础设施
docker-compose down
```

## 故障排查

### SSH 连接失败

```bash
# 测试 SSH 连接
ssh -p 22 your-username@your-remote-server

# 如果使用密钥，检查权限
chmod 600 ~/.ssh/id_rsa
```

### 服务启动失败

```bash
# 检查端口占用
lsof -i :8080
lsof -i :5432

# 检查数据库连接
docker-compose exec postgresql psql -U postgres -d owlrd -c "SELECT 1;"
```

## 详细文档

查看完整文档：`DEPLOYMENT.md`
