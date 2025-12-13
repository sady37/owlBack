# Repository 层实现总结

## ✅ 已完成

### 1. alarm_cloud.go - 报警策略仓库
- ✅ `GetAlarmCloudConfig` - 获取租户的报警策略配置（支持租户覆盖）
- ✅ `GetDeviceTypeAlarmConfig` - 获取设备类型的报警配置（使用数据库函数）

**功能**：
- 匹配优先级：1) 租户特定配置，2) 系统默认配置（tenant_id = NULL）
- 返回通用报警（OfflineAlarm, LowBattery, DeviceFailure）和设备特定报警配置

### 2. alarm_device.go - 设备报警配置仓库
- ✅ `GetAlarmDeviceConfig` - 获取设备的报警配置
- ✅ `GetDeviceMonitorConfig` - 获取设备的完整监控配置（使用数据库函数）
- ✅ `GetDeviceDefaultMonitorConfig` - 获取设备类型的默认配置（用于初次配置）

**功能**：
- 读取设备的个性化配置（覆盖 alarm_cloud 的默认值）
- 包含睡眠时间、各报警项及其级别、阈值等完整配置

### 3. alarm_events.go - 报警事件仓库
- ✅ `CreateAlarmEvent` - 创建报警事件
- ✅ `GetRecentAlarmEvent` - 获取最近的报警事件（用于去重检查）
- ✅ `UpdateAlarmEvent` - 更新报警事件（用于延长持续时间等）

**功能**：
- 写入报警事件到 PostgreSQL（alarm_events 表）
- 支持报警去重检查（检查最近 N 分钟内是否已有相同类型的报警）

### 4. card.go - 卡片仓库
- ✅ `GetCardByID` - 根据卡片ID获取卡片信息
- ✅ `GetCardDevices` - 获取卡片绑定的设备列表（从 cards.devices JSONB 字段）

**功能**：
- 查询卡片信息（用于报警评估）
- 解析卡片绑定的设备列表

### 5. device.go - 设备仓库
- ✅ `GetDeviceBindingInfo` - 获取设备的绑定信息
- ✅ `GetDevicesByRoom` - 获取房间内的所有设备
- ✅ `GetDevicesByBed` - 获取床上的所有设备

**功能**：
- 查询设备绑定关系（用于事件2：Sleepad可靠性判断）
- 查询房间/床上的设备列表

### 6. room.go - 房间仓库
- ✅ `GetRoomInfo` - 获取房间信息
- ✅ `IsBathroom` - 判断房间是否为卫生间（用于事件3：Bathroom可疑跌倒检测）
- ✅ `GetRoomByBedID` - 根据 bed_id 获取房间信息

**功能**：
- 查询房间信息
- 识别卫生间房间（通过 room_name 或 unit_name 中是否包含 bathroom/restroom/toilet）

## 📊 文件结构

```
wisefido-alarm/internal/repository/
├── alarm_cloud.go    ✅ 报警策略仓库
├── alarm_device.go   ✅ 设备报警配置仓库
├── alarm_events.go   ✅ 报警事件仓库
├── card.go           ✅ 卡片仓库
├── device.go         ✅ 设备仓库
└── room.go           ✅ 房间仓库
```

## 🔗 数据库表映射

| Repository | 数据库表 | 主要功能 |
|-----------|---------|---------|
| AlarmCloudRepository | `alarm_cloud` | 读取租户级别报警策略 |
| AlarmDeviceRepository | `alarm_device` | 读取设备级别报警配置 |
| AlarmEventsRepository | `alarm_events` | 写入报警事件 |
| CardRepository | `cards` | 查询卡片信息 |
| DeviceRepository | `devices`, `device_store` | 查询设备绑定关系 |
| RoomRepository | `rooms`, `units` | 查询房间信息 |

## ✅ 编译状态

- ✅ 所有文件编译通过
- ✅ 无编译错误

## 🚀 下一步

Repository 层已完成，下一步实现：
1. **Consumer 层** - 读取 Redis realtime 缓存
2. **Evaluator 层** - 实现事件1-4的评估逻辑
3. **Service 层** - 整合各层
4. **Main 入口** - 启动服务

