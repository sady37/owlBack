# wisefido-sensor-fusion 验证清单

## 📋 验证依据

- `owlRD/db/21_cards.sql` - 卡片表结构定义
- `owlRD/docs/20_Card_Creation_Rules_Final.md` - 卡片创建规则
- `wisefido-sensor-fusion/ISSUES_CHECK.md` - 问题检查报告

## ✅ 代码检查结果

### 1. 设备 JSONB 格式 ✅
- **要求**：`[{"device_id": "...", "device_name": "...", "device_type": "...", "device_model": "...", "binding_type": "direct|indirect"}, ...]`
- **实现**：`DeviceInfo` 结构体完全匹配
- **状态**：✅ 通过

### 2. 设备类型过滤 ✅
- **要求**：支持 Radar、Sleepace、SleepPad
- **实现**：`FuseCardData` 中已正确过滤
- **状态**：✅ 通过

### 3. 卡片类型支持 ✅
- **要求**：支持 ActiveBed 和 Location
- **实现**：`CardInfo.CardType` 支持两种类型
- **状态**：✅ 通过

### 4. GetCardByDeviceID 查询逻辑 ✅ **已修复**
- **场景 1**：设备绑定到床 → 查询 ActiveBed 卡片 ✅
- **场景 2**：设备绑定到房间 → 查询 Location 卡片 ✅
- **场景 3**：设备直接绑定到 unit → 查询 Location 卡片 ✅ **新增**
- **状态**：✅ 已修复并编译通过

### 5. 从 JSONB 读取设备列表 ✅
- **要求**：从 `cards.devices` JSONB 字段读取
- **实现**：`GetCardDevices` 正确解析
- **状态**：✅ 通过

## 🔍 功能验证清单

### 验证 1：设备绑定到床
- [ ] 创建测试数据：设备 `bound_bed_id = 'bed-123'`
- [ ] 运行 `wisefido-card-aggregator` 创建 ActiveBed 卡片
- [ ] 运行 `wisefido-sensor-fusion`，发送设备数据
- [ ] 验证：能正确查询到 ActiveBed 卡片
- [ ] 验证：能正确融合数据并更新缓存

### 验证 2：设备绑定到房间
- [ ] 创建测试数据：设备 `bound_room_id = 'room-789'`
- [ ] 运行 `wisefido-card-aggregator` 创建 Location 卡片
- [ ] 运行 `wisefido-sensor-fusion`，发送设备数据
- [ ] 验证：能正确查询到 Location 卡片（通过 room.unit_id）
- [ ] 验证：能正确融合数据并更新缓存

### 验证 3：设备直接绑定到 unit
- [ ] 创建测试数据：设备 `bound_bed_id IS NULL`, `bound_room_id IS NULL`, `unit_id = 'unit-456'`
- [ ] 运行 `wisefido-card-aggregator` 创建 Location 卡片
- [ ] 运行 `wisefido-sensor-fusion`，发送设备数据
- [ ] 验证：能正确查询到 Location 卡片（直接通过 devices.unit_id）
- [ ] 验证：能正确融合数据并更新缓存

### 验证 4：融合逻辑
- [ ] 测试场景：卡片同时有 Radar 和 Sleepace 设备
- [ ] 验证：HR/RR 优先使用 Sleepace 数据
- [ ] 验证：床状态/睡眠状态优先使用 Sleepace 数据
- [ ] 验证：姿态数据使用 Radar 数据

### 验证 5：缓存更新
- [ ] 验证：Redis 缓存键格式正确（`vital-focus:card:{card_id}:realtime`）
- [ ] 验证：缓存数据格式正确（JSON）
- [ ] 验证：TTL 设置正确（5分钟）

## 📝 验证步骤

### 步骤 1：准备测试数据
```sql
-- 1. 创建测试租户、单元、床位、设备
-- 2. 绑定设备到床/房间/unit
-- 3. 运行 wisefido-card-aggregator 创建卡片
```

### 步骤 2：运行 wisefido-sensor-fusion
```bash
cd wisefido-sensor-fusion
go run cmd/wisefido-sensor-fusion/main.go
```

### 步骤 3：发送测试数据
```bash
# 通过 Redis Streams 发送设备数据
redis-cli XADD iot:data:stream * data '{"device_id": "...", ...}'
```

### 步骤 4：检查结果
```bash
# 检查日志
# 检查 Redis 缓存
redis-cli GET "vital-focus:card:{card_id}:realtime"
```

## ✅ 总结

### 代码层面
- ✅ 所有代码检查通过
- ✅ GetCardByDeviceID 已修复，支持三种查询场景
- ✅ 设备 JSONB 格式匹配
- ✅ 设备类型过滤正确
- ✅ 卡片类型支持完整

### 功能层面
- ⏳ 需要实际运行验证
- ⏳ 需要测试三种设备绑定场景
- ⏳ 需要验证融合逻辑
- ⏳ 需要验证缓存更新

## 🚀 下一步

1. **运行 wisefido-card-aggregator** 创建卡片
2. **运行 wisefido-sensor-fusion** 测试融合功能
3. **验证完整数据流**：设备数据 → 卡片查询 → 数据融合 → 缓存更新

