# wisefido-alarm 问题排查指南

## 🔍 测试结果

### ✅ 服务编译和启动
- ✅ 代码编译成功
- ✅ 服务可以启动
- ✅ 代码逻辑正常

### ⚠️ 当前问题
- ⚠️ PostgreSQL 数据库连接失败：`dial tcp [::1]:5432: connect: connection refused`

## 🛠️ 解决方案

### 问题 1：PostgreSQL 未运行

**症状**：`connection refused` 错误

**解决方案**：

#### macOS (使用 Homebrew)
```bash
# 检查 PostgreSQL 状态
brew services list | grep postgresql

# 启动 PostgreSQL
brew services start postgresql

# 或者使用 postgres 命令
pg_ctl -D /usr/local/var/postgres start
```

#### Linux
```bash
# 检查 PostgreSQL 状态
sudo systemctl status postgresql

# 启动 PostgreSQL
sudo systemctl start postgresql
```

#### Docker
```bash
# 启动 PostgreSQL 容器
docker run -d \
  --name postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=owlrd \
  -p 5432:5432 \
  postgres:15
```

### 问题 2：数据库连接配置错误

**检查步骤**：

1. **检查环境变量**：
```bash
echo $DB_HOST
echo $DB_USER
echo $DB_PASSWORD
echo $DB_NAME
```

2. **测试数据库连接**：
```bash
# 如果安装了 psql
psql -h localhost -U postgres -d owlrd -c "SELECT 1;"

# 或者使用环境变量
PGPASSWORD=postgres psql -h localhost -U postgres -d owlrd -c "SELECT 1;"
```

3. **检查数据库是否存在**：
```bash
psql -h localhost -U postgres -c "\l" | grep owlrd
```

如果数据库不存在，需要创建：
```sql
CREATE DATABASE owlrd;
```

### 问题 3：Redis 未运行

**检查步骤**：

```bash
# 检查 Redis 状态
redis-cli -h localhost -p 6379 ping

# 如果返回 PONG，说明 Redis 运行正常
# 如果连接失败，需要启动 Redis
```

**启动 Redis**：

#### macOS (使用 Homebrew)
```bash
brew services start redis
```

#### Linux
```bash
sudo systemctl start redis
```

#### Docker
```bash
docker run -d \
  --name redis \
  -p 6379:6379 \
  redis:7
```

## ✅ 完整测试流程

### 步骤 1：检查环境

```bash
# 检查 PostgreSQL
psql -h localhost -U postgres -d owlrd -c "SELECT 1;" || echo "PostgreSQL 未运行"

# 检查 Redis
redis-cli -h localhost -p 6379 ping || echo "Redis 未运行"
```

### 步骤 2：设置环境变量

```bash
export TENANT_ID="test-tenant"
export DB_HOST="localhost"
export DB_USER="postgres"
export DB_PASSWORD="postgres"
export DB_NAME="owlrd"
export REDIS_ADDR="localhost:6379"
```

### 步骤 3：运行服务

```bash
cd /Users/sady3721/project/owlBack/wisefido-alarm

# 方式1：使用测试脚本（运行10秒）
bash scripts/test_run.sh

# 方式2：直接运行（持续运行）
./wisefido-alarm
```

### 步骤 4：验证运行

**检查日志**：
- 应该看到 "Starting alarm service"
- 应该看到 "Cache consumer started"
- 应该看到 "Evaluating cards"

**检查数据库**：
```sql
-- 检查报警事件
SELECT COUNT(*) FROM alarm_events;
```

**检查 Redis**：
```bash
# 检查报警缓存
redis-cli KEYS "vital-focus:card:*:alarms"
```

## 📊 预期行为

### 正常启动日志

```json
{"level":"info","msg":"Starting alarm service","tenant_id":"test-tenant"}
{"level":"info","msg":"Cache consumer started","tenant_id":"test-tenant","poll_interval":10}
{"level":"debug","msg":"Evaluating cards","card_count":10}
```

### 如果数据库连接成功

- ✅ 服务持续运行
- ✅ 每10秒轮询一次卡片
- ✅ 评估报警事件
- ✅ 写入数据库（如果有报警生成）

### 如果数据库连接失败

- ❌ 服务启动后立即退出
- ❌ 日志显示 "Failed to create alarm service"
- ❌ 错误信息：`connection refused` 或 `authentication failed`

## 🔗 相关文档

- `QUICK_START.md` - 快速启动指南
- `VERIFY.md` - 详细验证指南
- `RUN_TEST.md` - 运行测试指南

