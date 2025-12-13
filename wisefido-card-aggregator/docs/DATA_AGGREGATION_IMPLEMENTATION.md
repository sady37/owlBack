# wisefido-card-aggregator 数据聚合功能实现总结

## ✅ 已完成

### 1. 模型定义
- ✅ `internal/models/vital_focus_card.go` - VitalFocusCard 模型
  - 包含基础信息、住户、设备、实时数据、报警数据等所有字段
  - 与前端 TypeScript 接口保持一致

### 2. Repository 层扩展
- ✅ `internal/repository/card_info.go` - 卡片信息查询
  - `GetCardByID` - 获取卡片基础信息（包含报警统计）
  - `GetCardDevices` - 获取卡片绑定的设备列表（从 JSONB）
  - `GetCardResidents` - 获取卡片绑定的住户列表（从 JSONB）
  - `GetAllCards` - 获取所有卡片（用于批量聚合）

### 3. 数据聚合器
- ✅ `internal/aggregator/data_aggregator.go` - 数据聚合器
  - `AggregateCard` - 聚合单个卡片的数据
  - 读取 PostgreSQL: cards 表（基础信息）
  - 读取 Redis: `vital-focus:card:{card_id}:realtime`（实时数据）
  - 读取 Redis: `vital-focus:card:{card_id}:alarms`（报警数据）
  - 组装完整的 VitalFocusCard 对象

### 4. 缓存管理器
- ✅ `internal/aggregator/cache_manager.go` - 缓存管理器
  - `UpdateFullCardCache` - 更新完整卡片缓存
  - 写入 Redis: `vital-focus:card:{card_id}:full`（TTL: 10秒）

### 5. 服务层集成
- ✅ `internal/service/aggregator.go` - 服务层更新
  - 添加数据聚合器和缓存管理器
  - `startDataAggregation` - 启动数据聚合任务
  - `aggregateAllCards` - 批量聚合所有卡片
  - 与卡片创建任务并行运行

### 6. 配置更新
- ✅ `internal/config/config.go` - 配置更新
  - 添加 `Aggregation.Enabled` - 是否启用数据聚合
  - 添加 `Aggregation.Interval` - 聚合间隔（默认 10 秒）

## 📊 数据流

```
PostgreSQL (cards 表)
    ↓
DataAggregator.AggregateCard()
    ├─ 读取卡片基础信息
    ├─ 读取设备列表（cards.devices JSONB）
    ├─ 读取住户列表（cards.residents JSONB）
    ├─ 读取实时数据（Redis: vital-focus:card:{card_id}:realtime）
    ├─ 读取报警数据（Redis: vital-focus:card:{card_id}:alarms）
    └─ 组装 VitalFocusCard 对象
    ↓
CacheManager.UpdateFullCardCache()
    ↓
Redis (vital-focus:card:{card_id}:full, TTL: 10秒)
```

## 🔧 配置说明

### 环境变量

```bash
# 启用数据聚合（默认启用）
export CARD_AGGREGATION_ENABLED="true"

# 聚合间隔（秒，默认 10 秒）
export CARD_AGGREGATION_INTERVAL="10"
```

### 配置结构

```go
Aggregator.Aggregation.Enabled  // 是否启用数据聚合
Aggregator.Aggregation.Interval // 聚合间隔（秒）
```

## 🚀 运行方式

### 1. 启用数据聚合

```bash
# 设置环境变量
export TENANT_ID="your-tenant-id"
export CARD_AGGREGATION_ENABLED="true"
export CARD_AGGREGATION_INTERVAL="10"

# 运行服务
go run cmd/wisefido-card-aggregator/main.go
```

### 2. 服务行为

- **卡片创建任务**：每 60 秒全量创建卡片（轮询模式）
- **数据聚合任务**：每 10 秒聚合所有卡片数据（并行运行）

### 3. 日志输出

```json
{"level":"info","msg":"Starting card aggregator service","trigger_mode":"polling","aggregation_enabled":true}
{"level":"info","msg":"Starting data aggregation","interval":"10s"}
{"level":"debug","msg":"Aggregating cards","card_count":10}
{"level":"info","msg":"Completed aggregating cards","success_count":10,"error_count":0,"total_count":10}
```

## ✅ 验证

### 1. 检查 Redis 缓存

```bash
# 检查完整卡片缓存
redis-cli KEYS "vital-focus:card:*:full"

# 查看特定卡片的缓存
redis-cli GET "vital-focus:card:{card_id}:full"

# 检查 TTL
redis-cli TTL "vital-focus:card:{card_id}:full"
```

### 2. 验证数据完整性

```bash
# 查看缓存内容（JSON 格式）
redis-cli GET "vital-focus:card:{card_id}:full" | jq .
```

**预期内容**：
- ✅ 基础信息（card_id, card_name, card_address 等）
- ✅ 住户列表（residents）
- ✅ 设备列表（devices）
- ✅ 实时数据（heart, breath, sleep_stage, bed_status 等）
- ✅ 报警列表（alarms）

## 📝 注意事项

1. **前置依赖**：
   - 需要先运行 `wisefido-sensor-fusion` 生成实时数据缓存
   - 需要先运行 `wisefido-alarm` 生成报警数据缓存

2. **性能考虑**：
   - 聚合间隔默认 10 秒，可根据实际情况调整
   - 缓存 TTL 为 10 秒，确保数据及时更新

3. **错误处理**：
   - 如果实时数据或报警数据不存在，不影响聚合（继续处理）
   - 记录错误日志，但不中断聚合流程

## 🔗 相关文档

- `IMPLEMENTATION_SUMMARY.md` - 卡片创建功能总结
- `docs/EVENT_DRIVEN_IMPLEMENTATION.md` - 事件驱动模式说明
- `owlBack/docs/system_architecture_complete.md` - 系统架构文档

