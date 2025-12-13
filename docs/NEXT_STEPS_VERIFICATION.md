# 下一步：验证传感器融合功能

## 📋 当前状态

### ✅ 已完成
1. **卡片创建**（wisefido-card-aggregator）
   - ✅ 卡片创建规则实现
   - ✅ 轮询模式（每60秒全量更新）
   - ✅ 定时任务（每天上午9点）
   - ✅ 事件驱动模式（代码已实现，待 wisefido-data 服务发布事件）

2. **传感器融合**（wisefido-sensor-fusion）
   - ✅ 代码已实现
   - ✅ 依赖卡片表（cards 表）
   - ⏳ **需要验证**：确保能正确读取卡片并融合数据

## 🎯 验证目标

### 1. 验证卡片查询
- [ ] `wisefido-sensor-fusion` 能正确查询到卡片数据
- [ ] `GetCardByDeviceID` 能根据设备ID找到关联的卡片
- [ ] `GetCardDevices` 能获取卡片绑定的所有设备

### 2. 验证融合逻辑
- [ ] HR/RR 融合：优先 Sleepace，无数据则 Radar
- [ ] 床状态/睡眠状态融合：优先 Sleepace
- [ ] 姿态数据：使用所有 Radar 数据
- [ ] 融合条件：同时有 Radar 和 Sleepace 时进行融合

### 3. 验证缓存更新
- [ ] Redis 缓存键格式正确：`vital-focus:card:{card_id}:realtime`
- [ ] 缓存数据格式正确（JSON）
- [ ] TTL 设置正确（5分钟）

### 4. 验证完整数据流
- [ ] 设备数据 → `iot:data:stream`
- [ ] `wisefido-sensor-fusion` 消费数据
- [ ] 查询卡片 → 融合数据 → 更新缓存

## 🔍 验证步骤

### 步骤 1：检查卡片数据
```sql
-- 检查 cards 表是否有数据
SELECT card_id, card_type, bed_id, unit_id, devices 
FROM cards 
WHERE tenant_id = 'your-tenant-id'
LIMIT 10;
```

### 步骤 2：检查设备绑定
```sql
-- 检查设备是否绑定到卡片
SELECT d.device_id, d.device_type, d.bound_bed_id, d.bound_room_id
FROM devices d
WHERE d.tenant_id = 'your-tenant-id'
  AND d.monitoring_enabled = TRUE
LIMIT 10;
```

### 步骤 3：运行 wisefido-sensor-fusion 服务
```bash
# 启动服务
cd wisefido-sensor-fusion
go run cmd/wisefido-sensor-fusion/main.go

# 检查日志
# 应该能看到：
# - 成功查询到卡片
# - 成功融合数据
# - 成功更新缓存
```

### 步骤 4：检查 Redis 缓存
```bash
# 检查缓存键
redis-cli KEYS "vital-focus:card:*:realtime"

# 查看缓存内容
redis-cli GET "vital-focus:card:{card_id}:realtime"
```

## 📝 可能的问题

### 问题 1：cards 表为空
**症状**：`GetCardByDeviceID` 返回错误
**解决**：运行 `wisefido-card-aggregator` 创建卡片

### 问题 2：设备未绑定到卡片
**症状**：设备数据到达，但找不到关联的卡片
**解决**：检查设备绑定关系，确保设备绑定到床位或房间

### 问题 3：融合数据为空
**症状**：卡片存在，但融合后的数据为空
**解决**：检查 `iot_timeseries` 表是否有设备数据

## 🚀 验证后的下一步

验证通过后，下一步应该是：
1. **实现报警评估层**（wisefido-alarm）
   - 读取融合后的实时数据
   - 评估报警规则
   - 生成报警事件

## 🔗 相关文档

- `docs/system_architecture_complete.md` - 系统架构文档
- `docs/12_Sensor_Fusion_Implementation.md` - 传感器融合实现文档
- `wisefido-card-aggregator/docs/IMPLEMENTATION_SUMMARY.md` - 卡片创建实现总结

