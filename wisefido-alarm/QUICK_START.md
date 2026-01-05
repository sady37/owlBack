# wisefido-alarm 快速启动指南

## 🚀 快速启动

### 1. 环境验证

```bash
cd /Users/sady3721/project/owlBack/wisefido-alarm
bash scripts/verify_setup.sh
```

### 2. 设置环境变量

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
```

### 3. 启动服务

```bash
# 方式1：直接运行
go run cmd/wisefido-alarm/main.go

# 方式2：编译后运行
go build -o wisefido-alarm cmd/wisefido-alarm/main.go
./wisefido-alarm
```

## 📊 服务行为

### 轮询模式
- 每 **10秒** 轮询一次所有卡片
- 批量评估（每批 **10** 张卡片）
- 读取 Redis 实时数据缓存
- 评估报警事件（事件1-4）
- 更新报警缓存

### 日志输出
```json
{"level":"info","msg":"Starting alarm service","tenant_id":"your-tenant-id"}
{"level":"info","msg":"Cache consumer started","tenant_id":"your-tenant-id","poll_interval":10}
{"level":"debug","msg":"Evaluating cards","card_count":10}
```

## ✅ 验证服务运行

### 1. 检查日志
- 确认服务启动成功
- 确认定期轮询（每10秒）
- 确认卡片评估过程

### 2. 检查 Redis 缓存
```bash
# 检查报警缓存（如果有报警生成）
redis-cli KEYS "vital-focus:card:*:alarms"

# 检查状态缓存（事件1-4的状态）
redis-cli KEYS "alarm:state:*"
```

### 3. 检查数据库
```sql
-- 检查报警事件（待实现写入功能）
SELECT * FROM alarm_events ORDER BY created_at DESC LIMIT 10;
```

## 🛑 停止服务

按 `Ctrl+C` 优雅停止服务。

## 📝 注意事项

1. **前置依赖**：
   - 需要先运行 `wisefido-card-aggregator` 创建卡片
   - 需要先运行 `wisefido-sensor-fusion` 生成实时数据

2. **当前状态**：
   - 基础框架已完成 ✅
   - 事件1-4的评估逻辑为简化版本（返回空列表，待完善）
   - 报警事件写入功能已实现 ✅

3. **性能**：
   - 当前通过扫描 Redis 键获取卡片ID（效率较低）
   - 建议后续优化为从 PostgreSQL 查询

## 🔗 相关文档

- `VERIFY.md` - 详细验证指南
- `IMPLEMENTATION_SUMMARY.md` - 实现总结
- `REPOSITORY_LAYER_SUMMARY.md` - Repository 层总结

