# wisefido-alarm 运行和测试指南

## 🚀 快速运行

### 步骤 1：环境验证

```bash
cd /Users/sady3721/project/owlBack/wisefido-alarm
bash scripts/verify_setup.sh
```

### 步骤 2：设置环境变量

```bash
# 必需：设置租户ID
export TENANT_ID="your-tenant-id"

# 可选：数据库配置（有默认值）
export DB_HOST="localhost"
export DB_USER="postgres"
export DB_PASSWORD="postgres"
export DB_NAME="owlrd"

# 可选：Redis 配置（有默认值）
export REDIS_ADDR="localhost:6379"

# 可选：日志级别（debug/info/warn/error）
export LOG_LEVEL="info"
```

### 步骤 3：启动服务

```bash
# 方式1：直接运行
go run cmd/wisefido-alarm/main.go

# 方式2：编译后运行
go build -o wisefido-alarm cmd/wisefido-alarm/main.go
./wisefido-alarm
```

## ✅ 验证服务运行

### 1. 检查日志输出

服务启动后，应该看到：

```json
{"level":"info","msg":"Starting alarm service","tenant_id":"your-tenant-id"}
{"level":"info","msg":"Cache consumer started","tenant_id":"your-tenant-id","poll_interval":5}
{"level":"debug","msg":"Evaluating cards","card_count":10}
```

### 2. 检查数据库

```sql
-- 连接到数据库
psql -h localhost -U postgres -d owlrd

-- 检查报警事件（如果有报警生成）
SELECT 
    event_id,
    event_type,
    alarm_level,
    alarm_status,
    triggered_at,
    device_id
FROM alarm_events
ORDER BY triggered_at DESC
LIMIT 10;
```

### 3. 检查 Redis 缓存

```bash
# 检查报警缓存（如果有报警生成）
redis-cli KEYS "vital-focus:card:*:alarms"

# 查看特定卡片的报警缓存
redis-cli GET "vital-focus:card:{card_id}:alarms"

# 检查状态缓存（事件1-4的状态）
redis-cli KEYS "alarm:state:*"
```

## 🧪 测试场景

### 测试 1：服务启动测试

**目标**：验证服务能正常启动

**步骤**：
```bash
export TENANT_ID="test-tenant"
go run cmd/wisefido-alarm/main.go
```

**预期结果**：
- ✅ 服务成功启动
- ✅ 无错误日志
- ✅ 显示 "Cache consumer started"
- ✅ 定期轮询（每5秒）

### 测试 2：卡片查询测试

**目标**：验证服务能正确查询卡片

**验证SQL**：
```sql
SELECT COUNT(*) FROM cards WHERE tenant_id = 'test-tenant';
```

**预期结果**：
- ✅ 日志显示正确的卡片数量
- ✅ 能成功读取卡片信息

### 测试 3：实时数据读取测试

**目标**：验证服务能正确读取实时数据

**验证命令**：
```bash
redis-cli KEYS "vital-focus:card:*:realtime"
redis-cli GET "vital-focus:card:{card_id}:realtime"
```

**预期结果**：
- ✅ 能成功读取实时数据
- ✅ 如果卡片没有实时数据，服务能正确处理（跳过）

### 测试 4：报警事件写入测试

**目标**：验证报警事件能正确写入数据库

**前置条件**：
- 事件评估器生成报警（需要完善评估逻辑）

**验证SQL**：
```sql
SELECT * FROM alarm_events 
WHERE tenant_id = 'test-tenant' 
ORDER BY triggered_at DESC 
LIMIT 10;
```

**预期结果**：
- ✅ 报警事件能正确写入数据库
- ✅ 字段完整（event_id, event_type, alarm_level 等）
- ✅ trigger_data 和 metadata 正确序列化

### 测试 5：报警缓存更新测试

**目标**：验证报警缓存能正确更新

**验证命令**：
```bash
redis-cli GET "vital-focus:card:{card_id}:alarms"
redis-cli TTL "vital-focus:card:{card_id}:alarms"
```

**预期结果**：
- ✅ 报警缓存键格式正确
- ✅ TTL 设置正确（30秒）
- ✅ 数据格式正确（JSON）

## 🐛 问题排查

### 问题 1：服务启动失败

**症状**：服务启动后立即退出

**排查步骤**：
1. 检查环境变量（特别是 `TENANT_ID`）
2. 检查数据库连接
3. 检查 Redis 连接
4. 查看错误日志

**解决方案**：
```bash
# 检查环境变量
echo $TENANT_ID

# 测试数据库连接
psql -h localhost -U postgres -d owlrd -c "SELECT 1;"

# 测试 Redis 连接
redis-cli ping
```

### 问题 2：没有卡片数据

**症状**：日志显示 `card_count: 0`

**排查步骤**：
1. 检查 `cards` 表是否有数据
2. 检查租户ID是否正确
3. 确认是否运行了 `wisefido-card-aggregator`

**解决方案**：
```sql
-- 检查卡片数据
SELECT COUNT(*) FROM cards WHERE tenant_id = 'your-tenant-id';

-- 如果没有数据，运行 wisefido-card-aggregator
cd /Users/sady3721/project/owlBack/wisefido-card-aggregator
go run cmd/wisefido-card-aggregator/main.go
```

### 问题 3：无法读取实时数据

**症状**：日志显示 "Realtime data not found"

**排查步骤**：
1. 检查 Redis 中是否有实时数据缓存
2. 检查缓存键格式是否正确
3. 确认是否运行了 `wisefido-sensor-fusion`

**解决方案**：
```bash
# 检查实时数据缓存
redis-cli KEYS "vital-focus:card:*:realtime"

# 如果没有数据，运行 wisefido-sensor-fusion
cd /Users/sady3721/project/owlBack/wisefido-sensor-fusion
go run cmd/wisefido-sensor-fusion/main.go
```

### 问题 4：报警事件未写入

**症状**：评估器执行但没有报警事件写入

**排查步骤**：
1. 检查事件评估器是否生成报警（当前为简化版本，返回空列表）
2. 检查数据库连接
3. 查看错误日志

**解决方案**：
- 当前事件1-4的评估逻辑是简化版本，需要完善评估逻辑才能生成报警
- 检查日志中是否有 "Failed to create alarm event" 错误

## 📊 性能监控

### 监控指标

1. **处理速度**：
   - 每5秒轮询一次
   - 批量评估（每批10张卡片）

2. **内存使用**：
   - 观察长时间运行的内存使用情况

3. **错误率**：
   - 观察日志中的错误频率

### 日志分析

```bash
# 查看错误日志
go run cmd/wisefido-alarm/main.go 2>&1 | grep -i error

# 查看报警事件创建日志
go run cmd/wisefido-alarm/main.go 2>&1 | grep "Alarm event created"
```

## 🔗 相关文档

- `QUICK_START.md` - 快速启动指南
- `VERIFY.md` - 详细验证指南
- `TESTING_GUIDE.md` - 测试指南
- `IMPLEMENTATION_SUMMARY.md` - 实现总结
- `ALARM_EVENT_WRITE.md` - 报警事件写入说明

