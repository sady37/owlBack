# wisefido-alarm 服务验证指南

## 📋 验证前准备

### 1. 环境要求
- ✅ PostgreSQL 数据库运行中（包含 `cards`, `alarm_cloud`, `alarm_device`, `alarm_events` 表）
- ✅ Redis 运行中
- ✅ `wisefido-card-aggregator` 已运行并创建了卡片
- ✅ `wisefido-sensor-fusion` 已运行并生成了实时数据缓存

### 2. 检查清单

#### 步骤 1：运行环境验证脚本

```bash
cd /Users/sady3721/project/owlBack/wisefido-alarm
chmod +x scripts/verify_setup.sh
bash scripts/verify_setup.sh
```

#### 步骤 2：检查卡片数据

```sql
-- 连接到数据库
psql -h localhost -U postgres -d owlrd

-- 检查 cards 表是否有数据
SELECT 
    card_id, 
    card_type, 
    bed_id, 
    unit_id, 
    card_name,
    jsonb_array_length(devices) as device_count
FROM cards 
LIMIT 10;

-- 检查卡片绑定的设备
SELECT 
    card_id,
    card_type,
    devices
FROM cards
WHERE jsonb_array_length(devices) > 0
LIMIT 5;
```

#### 步骤 3：检查 Redis 实时数据缓存

```bash
# 检查缓存键
redis-cli KEYS "vital-focus:card:*:realtime"

# 查看特定卡片的缓存
redis-cli GET "vital-focus:card:{card_id}:realtime"

# 检查 TTL
redis-cli TTL "vital-focus:card:{card_id}:realtime"
```

#### 步骤 4：检查报警配置

```sql
-- 检查报警策略配置
SELECT * FROM alarm_cloud LIMIT 5;

-- 检查设备报警配置
SELECT * FROM alarm_device LIMIT 5;
```

## 🚀 运行服务

### 步骤 1：设置环境变量

```bash
# 必需
export TENANT_ID="your-tenant-id"

# 数据库配置（可选，有默认值）
export DB_HOST="localhost"
export DB_PORT=5432
export DB_USER="postgres"
export DB_PASSWORD="postgres"
export DB_NAME="owlrd"
export DB_SSLMODE="disable"

# Redis 配置（可选，有默认值）
export REDIS_ADDR="localhost:6379"
export REDIS_PASSWORD=""

# 日志配置（可选，有默认值）
export LOG_LEVEL="info"
export LOG_FORMAT="json"
```

### 步骤 2：启动服务

```bash
cd /Users/sady3721/project/owlBack/wisefido-alarm
go run cmd/wisefido-alarm/main.go
```

### 步骤 3：检查服务日志

服务启动后，应该看到：

```json
{"level":"info","msg":"Starting alarm service","tenant_id":"your-tenant-id"}
{"level":"info","msg":"Cache consumer started","tenant_id":"your-tenant-id","poll_interval":5}
{"level":"debug","msg":"Evaluating cards","card_count":10}
```

## ✅ 验证检查点

### 1. 服务启动验证
- [ ] 服务能成功启动（无错误日志）
- [ ] 能成功连接 PostgreSQL
- [ ] 能成功连接 Redis
- [ ] CacheConsumer 正常启动

### 2. 卡片查询验证
- [ ] 服务能成功查询到所有卡片（日志中显示 card_count > 0）
- [ ] `GetAllCards` 能正确返回卡片列表
- [ ] 卡片信息包含 card_id, card_type, bed_id, unit_id 等字段

### 3. 实时数据读取验证
- [ ] 服务能成功读取 Redis 实时数据缓存
- [ ] `GetRealtimeData` 能正确解析 JSON 数据
- [ ] 如果卡片没有实时数据，服务能正确处理（跳过）

### 4. 报警评估验证
- [ ] 服务能成功调用 Evaluator.Evaluate
- [ ] 事件1-4的评估器能正常执行（当前返回空列表，待完善）
- [ ] 没有异常错误日志

### 5. 报警缓存更新验证
- [ ] 如果生成报警，能正确更新 Redis 报警缓存
- [ ] 缓存键格式正确：`vital-focus:card:{card_id}:alarms`
- [ ] TTL 设置正确（30秒）

### 6. 状态管理验证
- [ ] 状态管理器能正确设置/获取状态
- [ ] 状态键格式正确：`alarm:state:{card_id}:track_{track_id}:{state_type}`

## 📊 测试场景

### 场景 1：基础运行测试

**目标**：验证服务能正常启动和运行

**步骤**：
1. 运行环境验证脚本
2. 设置环境变量
3. 启动服务
4. 观察日志，确认无错误

**预期结果**：
- 服务正常启动
- 定期轮询卡片（每5秒）
- 日志显示评估过程

### 场景 2：卡片数据读取测试

**目标**：验证服务能正确读取卡片数据

**步骤**：
1. 确保 `cards` 表有数据
2. 启动服务
3. 检查日志中的 `card_count`

**预期结果**：
- 日志显示正确的卡片数量
- 能成功读取卡片信息

### 场景 3：实时数据读取测试

**目标**：验证服务能正确读取实时数据

**步骤**：
1. 确保 Redis 中有实时数据缓存
2. 启动服务
3. 检查日志中的实时数据读取情况

**预期结果**：
- 能成功读取实时数据
- 如果卡片没有实时数据，服务能正确处理（跳过）

### 场景 4：报警评估测试（待完善）

**目标**：验证报警评估逻辑（当前为简化版本）

**步骤**：
1. 准备测试数据（卡片 + 实时数据）
2. 启动服务
3. 观察评估结果

**预期结果**：
- 评估器能正常执行
- 当前返回空列表（待完善评估逻辑）

## 🐛 问题排查

### 问题 1：编译失败
**检查**：
- Go 版本是否 >= 1.21
- 依赖是否完整：`go mod tidy`
- 检查编译错误信息

### 问题 2：数据库连接失败
**检查**：
- PostgreSQL 是否运行
- 环境变量是否正确
- 数据库用户权限是否正确

### 问题 3：Redis 连接失败
**检查**：
- Redis 是否运行
- 环境变量是否正确
- Redis 密码是否正确

### 问题 4：没有卡片数据
**检查**：
- `cards` 表是否有数据
- 是否运行了 `wisefido-card-aggregator`
- 租户ID是否正确

### 问题 5：没有实时数据缓存
**检查**：
- Redis 中是否有 `vital-focus:card:*:realtime` 键
- 是否运行了 `wisefido-sensor-fusion`
- 缓存键格式是否正确

### 问题 6：服务启动后立即退出
**检查**：
- 查看错误日志
- 检查数据库连接
- 检查 Redis 连接
- 检查环境变量

## 📝 下一步

1. **完善事件评估逻辑**：实现事件1-4的完整评估逻辑
2. **实现报警事件写入**：将报警事件写入 PostgreSQL
3. **添加单元测试**：为关键函数添加单元测试
4. **性能优化**：优化批量评估和状态管理
5. **监控和指标**：添加处理速度、报警数量等指标

## 🔗 相关文档

- `IMPLEMENTATION_SUMMARY.md` - 实现总结
- `REPOSITORY_LAYER_SUMMARY.md` - Repository 层总结
- `REQUIREMENTS_ANALYSIS.md` - 需求分析
- `ROOM_NAME_USAGE.md` - room_name 使用说明

