# wisefido-backend vs owlBack 报警评估层对比

## 📋 检查结果

### wisefido-backend（v1.0 架构）

**存在报警相关代码**：
- ✅ `common/commonmodels/alarm.go` - 报警模型定义（DeviceAlarm）
- ✅ `common/utils/alarm.go` - 报警工具函数
  - `GetSleepaceAlarmNotificationKey()` - Sleepace 报警通知键
  - `GetRadarAlarmNotificationKey()` - Radar 报警通知键
  - `GetAlarmCategory()` - 报警分类
  - `GetAlarmSeverity()` - 报警严重程度

**服务列表**：
- `wisefido-device` - 设备管理服务（HTTP API）
- `wisefido-data` - 数据服务（HTTP API）
- `wisefido-radar` - 雷达服务（HTTP API）
- `wisefido-sleepace` - Sleepace 服务（HTTP API）
- `wisefido-user` - 用户服务（HTTP API）
- `wisefido-resident` - 住户服务（HTTP API）
- `wisefido-address` - 地址服务（HTTP API）
- `wisefido-gateway` - 网关服务
- `wisefido-qinglan` - 清澜服务
- `wisefido-file` - 文件服务

**报警评估层**：
- ❌ **没有独立的报警评估服务**
- ⚠️ 报警可能是由设备直接触发，或由各个服务内部处理
- ⚠️ 没有找到专门的报警规则评估逻辑

### owlBack（v1.5 架构）

**报警评估层**：
- ✅ `wisefido-alarm` - **独立的报警评估服务**（待实现）
  - 读取融合后的实时数据（`vital-focus:card:{card_id}:realtime`）
  - 应用报警规则（alarm_rule.md 中的 4 个事件）
  - 生成报警事件
  - 写入 PostgreSQL（`alarm_events` 表）
  - 更新 Redis 缓存（`vital-focus:card:{card_id}:alarms`）

**架构差异**：
- v1.0：报警可能由设备直接触发，或由各个服务内部处理
- v1.5：独立的报警评估服务，统一处理所有报警规则

## 🔍 详细对比

### v1.0 架构（wisefido-backend）

```
设备 → MQTT → wisefido-radar/wisefido-sleepace
    ↓
HTTP API 服务（wisefido-device, wisefido-data 等）
    ↓
报警处理（可能在各个服务内部）
    ↓
数据库（MySQL）
```

**特点**：
- 报警可能由设备直接触发
- 报警处理分散在各个服务中
- 没有统一的报警评估层

### v1.5 架构（owlBack）

```
设备 → MQTT → wisefido-radar/wisefido-sleepace
    ↓
wisefido-data-transformer（数据标准化）
    ↓
wisefido-sensor-fusion（传感器融合）
    ↓
wisefido-alarm（报警评估）⭐ **新服务**
    ↓
PostgreSQL（alarm_events）+ Redis（alarms 缓存）
```

**特点**：
- 统一的报警评估服务
- 基于融合后的实时数据
- 支持复杂的 AI 智能评估（事件1-4）

## 📊 结论

### wisefido-backend（v1.0）
- ❌ **没有独立的报警评估层**
- ⚠️ 报警处理可能分散在各个服务中
- ⚠️ 没有统一的报警规则评估逻辑

### owlBack（v1.5）
- ✅ **有独立的报警评估层**（wisefido-alarm）
- ✅ 统一的报警规则评估
- ✅ 支持复杂的 AI 智能评估（alarm_rule.md 中的 4 个事件）

## 🎯 实现建议

由于 wisefido-backend 中没有独立的报警评估层，我们需要在 owlBack 中**全新实现** `wisefido-alarm` 服务。

**参考 wisefido-backend 的代码**：
- 可以复用报警类型定义（如果需要兼容）
- 可以复用报警分类和严重程度的逻辑
- 但核心的报警评估逻辑需要全新实现

**实现重点**：
1. 实现 alarm_rule.md 中的 4 个复杂事件
2. 实现报警规则评估逻辑
3. 实现状态管理和定时器（事件1）
4. 实现数据查询和位置跟踪（事件2-4）

