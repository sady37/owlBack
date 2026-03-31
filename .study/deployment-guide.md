# Owl Platform — 部署方案

> 面向全新 Linux 服务器的完整部署流程，基于 `owl_sady` 现有工程结构。

---

## 一、前置条件

### 1.1 服务器要求

| 项目 | 最低 | 推荐 |
|------|------|------|
| OS | Ubuntu 22.04 LTS | Ubuntu 22.04 / Debian 12 |
| CPU | 2 核 | 4 核 |
| 内存 | 4 GB | 8 GB |
| 磁盘 | 40 GB | 100 GB SSD |
| 网络 | 公网 IP（设备接入需要） | 固定 IP |

### 1.2 软件依赖

```bash
# Go（wisefido-data / wisefido-qinglan 需要 1.24+；其余 1.21+）
# 统一安装 1.24
wget https://go.dev/dl/go1.24.1.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.24.1.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version   # 验证

# Node.js 18+（owlFront）
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs
node -v && npm -v   # 验证

# Docker Engine + Compose Plugin
sudo apt-get update
sudo apt-get install -y ca-certificates curl gnupg lsb-release
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] \
  https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
sudo usermod -aG docker $USER   # 当前用户免 sudo 使用 docker（重新登录生效）
docker compose version   # 验证

# 其他工具
sudo apt-get install -y git lsof net-tools
```

---

## 二、拉取代码

```bash
# 目标目录与现有工程保持一致
sudo mkdir -p /home/wisefido/owl
sudo chown $USER:$USER /home/wisefido/owl
cd /home/wisefido/owl

# 若使用 Git（推荐）
git clone <your-repo-url> owl_sady
cd owl_sady

# 若从现有机器打包传输
# 在源机器：tar -czf owl_sady.tar.gz owl_sady/
# scp owl_sady.tar.gz user@new-server:/home/wisefido/owl/
# tar -xzf owl_sady.tar.gz
```

> 以下步骤均以 `/home/wisefido/owl/owl_sady` 为工程根目录。

---

## 三、配置 .env

```bash
cd /home/wisefido/owl/owl_sady/owlBack
cp .env.example .env
vim .env
```

**必须修改的配置项：**

```ini
# ── 数据库 ──
DB_PASSWORD=<强密码>

# ── Redis ──
REDIS_PASSWORD=<强密码>
# 同步修改 docker-compose.yml 中 redis command 里的密码，保持一致：
#   command: redis-server --requirepass <同上密码> --loglevel warning
# 同步修改 healthcheck test 里的 -a 参数

# ── MQTT（后端服务连接 Broker） ──
MQTT_BROKER=tcp://127.0.0.1:1883

# ── MQTT（雷达设备端配置） ──
# 填写本机的公网/内网 IP 或域名，雷达设备将连接此地址
RADAR_MQTT_SERVER=<本机公网IP或域名>
RADAR_MQTT_PORT=1883

# ── 租户 ID（初始化后从数据库查取，先保持默认） ──
TENANT_ID=bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c

# ── 服务端口（默认即可，如有冲突按需修改） ──
DATA_SERVICE_PORT=8080
QINGLAN_PORT=8081
CARDAGG_PORT=8082
SLEEPACE_PORT=8083
TIMESERIES_PORT=8085

# ── 报警推送回调（cardagg → wisefido-data） ──
WISEFIDO_DATA_ALARM_PUSH_URL=http://127.0.0.1:8080
INTERNAL_ALARM_PUSH_SECRET=<随机字符串>

# ── APNs（iOS 推送，无 iOS 需求可留空） ──
APNS_KEY_ID=
APNS_TEAM_ID=
APNS_BUNDLE_ID=
APNS_P8_KEY=
```

> ⚠️ `.env` 含有敏感信息，确保文件权限为 `600`：`chmod 600 .env`

---

## 四、启动 Docker 基础设施

```bash
cd /home/wisefido/owl/owl_sady/owlBack

# 启动 PostgreSQL/TimescaleDB、Redis、MQTT、MySQL
docker compose up -d

# 等待健康检查通过
docker compose ps   # 所有容器状态应为 healthy

# 查看日志（如有问题）
docker compose logs postgresql
docker compose logs redis
docker compose logs mqtt
```

**验证连通性：**

```bash
# PostgreSQL
docker exec owl-postgresql pg_isready -U postgres

# Redis
docker exec owl-redis redis-cli -a <REDIS_PASSWORD> ping
# 返回 PONG 即正常

# MQTT
docker exec owl-mqtt nc -z localhost 1883 && echo "MQTT OK"
```

---

## 五、初始化数据库 Schema

```bash
# 全量建表（首次部署必须执行）
docker exec -i owl-postgresql bash -c \
  'cd /docker-entrypoint-initdb.d && bash rebuild_database.sh'

# 验证建表结果
docker exec owl-postgresql psql -U postgres -d owlrd -c "\dt" | head -40
```

**初始化系统用户（默认管理员账号）：**

```bash
cd /home/wisefido/owl/owl_sady/owlRD/db
go run init_system_users.go
# 或使用封装脚本
bash run_init_system_users.sh
```

---

## 六、配置 MQTT TLS 证书

雷达设备通过 8443 端口走 HTTPS 认证，`owl-common` 已包含自签名证书，
**如需替换为正式证书：**

```bash
# 替换 owl-common 下的证书（服务启动时引用此路径）
cp your_server.crt /home/wisefido/owl/owl_sady/owlBack/owl-common/server.crt
cp your_server.key /home/wisefido/owl/owl_sady/owlBack/owl-common/server.key

# 同时替换 MQTT TLS 证书（Mosquitto 容器挂载）
cp your_server.crt /home/wisefido/owl/owl_sady/owlBack/mqtt/config/server.crt
cp your_server.key /home/wisefido/owl/owl_sady/owlBack/mqtt/config/server.key
docker compose restart mqtt
```

> 若仅内网测试，自签名证书直接使用即可，无需替换。

---

## 七、启动后端 Go 微服务

```bash
cd /home/wisefido/owl/owl_sady/owlBack
./start-owlback.sh
```

启动脚本依次以 `go run` 启动以下服务，日志写入 `../log/`：

| 服务 | 日志文件 |
|------|----------|
| wisefido-data | `log/wisefido-data.log` |
| wisefido-cardagg | `log/wisefido-cardagg.log` |
| wisefido-qinglan | `log/wisefido-qinglan.log` |
| wisefido-sleepace | `log/wisefido-sleepace.log` |
| wisefido-iot | `log/wisefido-iot.log` |
| wisefido-ai | `log/wisefido-ai.log` |

**验证服务启动：**

```bash
# 检查端口监听
ss -tlnp | grep -E '8080|8081|8082|8083|8085'

# 健康检查
curl -s http://127.0.0.1:8080/health
curl -s http://127.0.0.1:8081/health
curl -s http://127.0.0.1:8083/health
curl -s http://127.0.0.1:8085/health

# 实时查看日志
tail -f /home/wisefido/owl/log/wisefido-data.log
```

---

## 八、启动 Sleepace Java 服务（可选）

> 仅部署 Sleepace 可穿戴设备时需要此步骤。

```bash
cd /home/wisefido/owl/owl_sady/sleepace
bash sleepace.sh start

# 验证（默认 8090 端口）
curl -s http://127.0.0.1:8090/health || true
```

---

## 九、部署前端 owlFront

### 9.1 开发/测试环境（Vite dev server）

```bash
cd /home/wisefido/owl/owl_sady/owlFront
npm install
npm run dev   # 默认监听 :5173（或 vite.config.ts 中配置的端口）
```

### 9.2 生产环境（Nginx 静态托管）

```bash
cd /home/wisefido/owl/owl_sady/owlFront
npm install
npm run build   # 产物在 dist/

# 安装 Nginx
sudo apt-get install -y nginx

# 部署静态文件
sudo cp -r dist/* /var/www/html/owl/

# Nginx 配置示例（/etc/nginx/sites-available/owl）
sudo tee /etc/nginx/sites-available/owl > /dev/null <<'EOF'
server {
    listen 80;
    server_name _;
    root /var/www/html/owl;
    index index.html;

    # SPA fallback
    location / {
        try_files $uri $uri/ /index.html;
    }

    # 代理后端 API（避免跨域）
    location /admin/api/    { proxy_pass http://127.0.0.1:8080; proxy_set_header Host $host; }
    location /auth/api/     { proxy_pass http://127.0.0.1:8080; proxy_set_header Host $host; }
    location /data/api/     { proxy_pass http://127.0.0.1:8080; proxy_set_header Host $host; }
    location /settings/api/ { proxy_pass http://127.0.0.1:8080; proxy_set_header Host $host; }
    location /sleepace/api/ { proxy_pass http://127.0.0.1:8080; proxy_set_header Host $host; }
    location /radar-device/ { proxy_pass http://127.0.0.1:8080; proxy_set_header Host $host; }
    location /device/api/   { proxy_pass http://127.0.0.1:8080; proxy_set_header Host $host; }

    # SSE：禁用缓冲（实时推送必须）
    location ~* /stream {
        proxy_pass http://127.0.0.1:8080;
        proxy_buffering off;
        proxy_cache off;
        proxy_set_header Connection '';
        proxy_http_version 1.1;
        chunked_transfer_encoding on;
    }
}
EOF

sudo ln -sf /etc/nginx/sites-available/owl /etc/nginx/sites-enabled/owl
sudo nginx -t && sudo systemctl reload nginx
```

---

## 十、设置开机自启（systemd）

### 10.1 后端服务

```bash
# 修改 service 文件中的用户和路径（如不是 wisefido）
sudo cp /home/wisefido/owl/owl_sady/owlBack/systemd/owlback.service /etc/systemd/system/

# 若用户/路径不同，用 override 覆盖
sudo systemctl edit owlback
# 写入（根据实际情况修改）：
# [Service]
# User=youruser
# WorkingDirectory=/path/to/owlBack
# ExecStart=/path/to/owlBack/start-owlback.sh
# ExecStop=/path/to/owlBack/stop-owlback.sh

sudo systemctl daemon-reload
sudo systemctl enable owlback
sudo systemctl start owlback
sudo systemctl status owlback
```

### 10.2 前端（dev server，可选）

```bash
sudo cp /home/wisefido/owl/owl_sady/owlFront/systemd/*.service /etc/systemd/system/owlfront.service
# 同上修改 User / WorkingDirectory
sudo systemctl daemon-reload
sudo systemctl enable owlfront
sudo systemctl start owlfront
```

### 10.3 Docker 基础设施开机自启

Docker 服务本身已由 systemd 管理，容器设置 restart policy 即可：

```bash
cd /home/wisefido/owl/owl_sady/owlBack
# docker-compose.yml 中各服务加 restart: unless-stopped（已有则忽略）
docker compose up -d   # 重启后 Docker 守护进程会自动拉起容器
```

---

## 十一、防火墙端口开放

```bash
# UFW 示例（按需开放）
sudo ufw allow 80/tcp      # HTTP（Nginx 前端）
sudo ufw allow 443/tcp     # HTTPS（可选）
sudo ufw allow 1883/tcp    # MQTT（雷达设备接入）
sudo ufw allow 8883/tcp    # MQTTS（TLS MQTT）
sudo ufw allow 8443/tcp    # HTTPS 设备认证（wisefido-qinglan）
# 8080~8085 仅内部访问，不对外暴露（Nginx 代理）

sudo ufw enable
sudo ufw status
```

---

## 十二、部署验证清单

```bash
# 1. 基础设施
docker compose ps                          # 全部 healthy

# 2. 数据库
docker exec owl-postgresql psql -U postgres -d owlrd -c "SELECT COUNT(*) FROM tenants;"

# 3. 后端服务
ss -tlnp | grep -E '8080|8081|8082|8083|8085'
curl -sf http://127.0.0.1:8080/health && echo "data OK"
curl -sf http://127.0.0.1:8081/health && echo "qinglan OK"

# 4. 登录测试
curl -s -X POST http://127.0.0.1:8080/auth/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"<初始密码>"}' | python3 -m json.tool

# 5. Redis Stream（确认 cardagg 在消费）
docker exec owl-redis redis-cli -a <REDIS_PASSWORD> \
  XINFO GROUPS iot:monitor:stream

# 6. 前端可访问
curl -s http://127.0.0.1/ | grep -o '<title>.*</title>'
```

---

## 十三、常见问题

| 现象 | 排查步骤 |
|------|----------|
| 服务启动后立即退出 | `tail -50 log/wisefido-data.log` — 通常是 DB 或 Redis 连接失败，检查 .env 密码和 Docker 容器状态 |
| 端口已被占用 | `ss -tlnp \| grep 808x`，`./stop-owlback.sh` 后重启 |
| 雷达设备无法接入 | 确认 `RADAR_MQTT_SERVER` 填的是设备能到达的 IP；检查防火墙 1883/8443 |
| cardagg 卡片不更新 | 检查 `iot:monitor:stream` 是否有数据：`XLEN iot:monitor:stream`；检查 cardagg 日志 |
| TimescaleDB 扩展报错 | `docker exec owl-postgresql psql -U postgres -c "CREATE EXTENSION IF NOT EXISTS timescaledb;"` |
| Sleepace 数据不同步 | 检查 sleepace-service 日志，确认 MySQL 容器健康，MQTT 连接正常 |

---

## 十四、目录结构速查

```
/home/wisefido/owl/
├── owl_sady/
│   ├── owlBack/            # 后端微服务 + docker-compose.yml + .env
│   │   ├── start-owlback.sh
│   │   ├── stop-owlback.sh
│   │   ├── mqtt/config/    # Mosquitto 配置 + TLS 证书
│   │   └── owl-common/     # 共享库 + 自签名证书
│   ├── owlFront/           # Vue3 前端
│   ├── owlRD/db/           # SQL schema（数据库唯一真值来源）
│   └── sleepace/           # Sleepace Java 服务
└── log/                    # 所有后端服务日志（统一输出位置）
```
