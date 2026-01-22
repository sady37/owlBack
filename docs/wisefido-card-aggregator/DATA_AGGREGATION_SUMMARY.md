# wisefido-card-aggregator 数据聚合功能总结

## ✅ 实现完成

### 功能概述

`wisefido-card-aggregator` 现在支持两种模式：
1. **卡片创建模式**（已实现） - 创建和维护卡片
2. **数据聚合模式**（新实现） - 聚合卡片数据，生成完整的 VitalFocusCard 对象

### 已创建的文件

1. **模型层**
   - `internal/models/vital_focus_card.go` - VitalFocusCard 模型定义

2. **Repository 层**
   - `internal/repository/card_info.go` - 卡片信息查询方法
     - `GetCardByID` - 获取卡片基础信息
     - `GetCardDevices` - 获取设备列表
     - `GetCardResidents` - 获取住户列表
     - `GetAllCards` - 获取所有卡片

3. **Aggregator 层**
   - `internal/aggregator/data_aggregator.go` - 数据聚合器
   - `internal/aggregator/cache_manager.go` - 缓存管理器

4. **配置更新**
   - `internal/config/config.go` - 添加聚合配置

5. **服务层更新**
   - `internal/service/aggregator.go` - 集成数据聚合功能

## 📊 数据聚合流程

```
1. 从 PostgreSQL 读取卡片基础信息
   ↓
2. 解析 cards.devices JSONB（设备列表）
   ↓
3. 解析 cards.residents JSONB（住户列表）
   ↓
4. 从 Redis 读取实时数据（vital-focus:card:{card_id}:realtime）
   ↓
5. 从 Redis 读取报警数据（vital-focus:card:{card_id}:alarms）
   ↓
6. 组装完整的 VitalFocusCard 对象
   ↓
7. 写入 Redis 缓存（vital-focus:card:{card_id}:full, TTL: 10秒）
```

## 🎯 关键特性

### 1. 数据源整合
- ✅ PostgreSQL: 卡片基础信息、设备列表、住户列表
- ✅ Redis: 实时数据（来自 wisefido-sensor-fusion）
- ✅ Redis: 报警数据（来自 wisefido-ai）

### 2. 数据转换
- ✅ SNOMED 编码转换为数字（sleep_stage, bed_status, postures）
- ✅ 数据源转换为小写（'s'=sleepace, 'r'=radar, '-'=无数据）
- ✅ 报警事件格式转换

### 3. 错误处理
- ✅ 实时数据不存在时继续处理（不影响聚合）
- ✅ 报警数据不存在时继续处理（不影响聚合）
- ✅ 记录错误日志，但不中断聚合流程

### 4. 性能优化
- ✅ 批量聚合所有卡片
- ✅ 缓存 TTL 设置为 10 秒（及时更新）
- ✅ 聚合间隔可配置（默认 10 秒）

## 🚀 使用方式

### 配置

```bash
# 启用数据聚合（默认启用）
export CARD_AGGREGATION_ENABLED="true"

# 聚合间隔（秒，默认 10 秒）
export CARD_AGGREGATION_INTERVAL="10"
```

### 运行

```bash
cd /Users/sady3721/project/owlBack/wisefido-card-aggregator
export TENANT_ID="your-tenant-id"
go run cmd/wisefido-card-aggregator/main.go
```

### 验证

```bash
# 检查完整卡片缓存
redis-cli KEYS "vital-focus:card:*:full"
redis-cli GET "vital-focus:card:{card_id}:full"
```

## 📝 当前状态

- ✅ 代码编译通过
- ✅ 数据聚合功能已实现
- ✅ 与卡片创建功能并行运行
- ⚠️ 需要测试验证（需要 PostgreSQL 和 Redis 运行）

## 🔗 相关文档

- `docs/DATA_AGGREGATION_IMPLEMENTATION.md` - 详细实现说明
- `IMPLEMENTATION_SUMMARY.md` - 卡片创建功能总结
- `owlBack/docs/system_architecture_complete.md` - 系统架构文档

