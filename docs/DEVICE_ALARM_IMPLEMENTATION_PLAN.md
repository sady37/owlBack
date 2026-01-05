# 设备直接报警处理实施计划

## 📋 目标

在 `wisefido-sensor-fusion` 中添加设备直接报警处理功能，实现：
- 检测设备上报的报警事件（如 Fall）
- 立即创建 `alarm_events` 记录
- 与云端事件报警（事件1-4）分离处理

## 🎯 实施步骤（按顺序）

### 阶段 1：准备报警处理依赖（基础层）

#### 1.1 在 `wisefido-sensor-fusion` 中添加报警 Repository
**文件**：`wisefido-sensor-fusion/internal/repository/alarm_events.go`
- 复制 `wisefido-alarm/internal/repository/alarm_events.go` 的 `CreateAlarmEvent` 方法
- 简化版本，只保留创建功能

**文件**：`wisefido-sensor-fusion/internal/repository/alarm_device.go`
- 复制 `wisefido-alarm/internal/repository/alarm_device.go` 的 `GetAlarmDeviceConfig` 方法
- 用于读取设备报警配置，判断报警是否启用

#### 1.2 添加报警模型
**文件**：`wisefido-sensor-fusion/internal/models/alarm_event.go`
- 复制 `wisefido-alarm/internal/models/alarm_event.go`
- 或创建简化版本

**文件**：`wisefido-sensor-fusion/internal/models/trigger_data.go`
- 复制 `wisefido-alarm/internal/models/trigger_data.go`（如果存在）
- 或创建简化版本

**优先级**：⭐⭐⭐ 最高（必须先完成）

---

### 阶段 2：实现报警处理工具函数

#### 2.1 创建报警工具函数
**文件**：`wisefido-sensor-fusion/internal/alarm/alarm_handler.go`（新建）
- `IsDeviceDirectAlarm(eventType string) bool` - 判断是否是设备直接报警
- `IsAlarmEnabled(config, eventType) bool` - 判断报警是否启用
- `BuildDeviceAlarmEvent(...) *models.AlarmEvent` - 构建报警事件
- `GetAlarmCategory(eventType string) string` - 获取报警分类（safety/clinical/behavioral/device）
- `GetAlarmLevel(eventType string) string` - 获取报警级别（ALERT/WARNING等）

**优先级**：⭐⭐⭐ 最高（必须先完成）

---

### 阶段 3：集成到数据流中

#### 3.1 更新 `StreamConsumer` 结构
**文件**：`wisefido-sensor-fusion/internal/consumer/stream_consumer.go`
- 添加 `alarmEventsRepo *repository.AlarmEventsRepository`
- 添加 `alarmDeviceRepo *repository.AlarmDeviceRepository`
- 添加 `alarmHandler *alarm.AlarmHandler`

#### 3.2 更新 `NewStreamConsumer` 构造函数
**文件**：`wisefido-sensor-fusion/internal/consumer/stream_consumer.go`
- 初始化 `alarmEventsRepo`
- 初始化 `alarmDeviceRepo`
- 初始化 `alarmHandler`

#### 3.3 在 `processMessage` 中添加报警检测
**文件**：`wisefido-sensor-fusion/internal/consumer/stream_consumer.go`
- 在融合数据后，检测 `iotData.EventType`
- 如果是设备直接报警，调用 `alarmHandler.CreateDeviceAlarm`

**优先级**：⭐⭐⭐ 最高（核心功能）

---

### 阶段 4：更新服务初始化

#### 4.1 更新 `FusionService`
**文件**：`wisefido-sensor-fusion/internal/service/fusion.go`
- 在 `NewFusionService` 中创建报警 Repository
- 传递给 `NewStreamConsumer`

**优先级**：⭐⭐ 高（必须完成）

---

### 阶段 5：文档和测试

#### 5.1 更新文档
**文件**：`wisefido-sensor-fusion/README.md`
- 说明设备直接报警处理功能

**文件**：`wisefido-alarm/README.md`
- 明确说明只处理云端事件报警（事件1-4）

#### 5.2 测试验证
- 测试设备直接报警（Fall）是否能正确创建 `alarm_events`
- 测试云端事件报警不受影响

**优先级**：⭐ 中（完成后验证）

---

## 📝 详细实施清单

### ✅ 阶段 1：准备报警处理依赖

- [ ] 创建 `wisefido-sensor-fusion/internal/repository/alarm_events.go`
  - [ ] 实现 `AlarmEventsRepository` 结构
  - [ ] 实现 `CreateAlarmEvent` 方法
- [ ] 创建 `wisefido-sensor-fusion/internal/repository/alarm_device.go`
  - [ ] 实现 `AlarmDeviceRepository` 结构
  - [ ] 实现 `GetAlarmDeviceConfig` 方法
- [ ] 创建 `wisefido-sensor-fusion/internal/models/alarm_event.go`
  - [ ] 定义 `AlarmEvent` 结构
- [ ] 创建 `wisefido-sensor-fusion/internal/models/trigger_data.go`
  - [ ] 定义 `TriggerData` 结构

### ✅ 阶段 2：实现报警处理工具函数

- [ ] 创建 `wisefido-sensor-fusion/internal/alarm/alarm_handler.go`
  - [ ] 实现 `IsDeviceDirectAlarm` 函数
  - [ ] 实现 `IsAlarmEnabled` 函数
  - [ ] 实现 `BuildDeviceAlarmEvent` 函数
  - [ ] 实现 `GetAlarmCategory` 函数
  - [ ] 实现 `GetAlarmLevel` 函数

### ✅ 阶段 3：集成到数据流中

- [ ] 更新 `StreamConsumer` 结构
- [ ] 更新 `NewStreamConsumer` 构造函数
- [ ] 在 `processMessage` 中添加报警检测逻辑

### ✅ 阶段 4：更新服务初始化

- [ ] 更新 `NewFusionService` 创建报警 Repository
- [ ] 更新依赖注入

### ✅ 阶段 5：文档和测试

- [ ] 更新 README
- [ ] 测试设备直接报警
- [ ] 测试云端事件报警不受影响

---

## 🔄 实施顺序说明

### 为什么按这个顺序？

1. **阶段 1（基础层）**：必须先完成，因为后续所有功能都依赖这些 Repository 和 Model
2. **阶段 2（工具函数）**：提供可复用的函数，便于测试和维护
3. **阶段 3（集成）**：核心功能，将报警处理集成到数据流中
4. **阶段 4（服务初始化）**：确保服务能正确启动
5. **阶段 5（文档和测试）**：验证功能正确性

### 依赖关系

```
阶段 1 (Repository/Model)
    ↓
阶段 2 (工具函数)
    ↓
阶段 3 (集成)
    ↓
阶段 4 (服务初始化)
    ↓
阶段 5 (文档和测试)
```

---

## ⚠️ 注意事项

1. **代码复用**：可以考虑将报警 Repository 提取到 `owl-common`，但目前先复制代码，保持快速实施
2. **错误处理**：报警创建失败不应该影响数据融合流程，只记录警告日志
3. **性能考虑**：报警检测和处理应该是轻量级的，不影响数据融合性能
4. **配置检查**：必须检查设备报警配置，只有启用的报警才创建记录

---

## 🎯 预期结果

实施完成后：
- ✅ 设备直接上报的 Fall 报警能立即创建 `alarm_events` 记录
- ✅ 云端事件报警（事件1-4）继续由 `wisefido-alarm` 处理
- ✅ 两个报警处理流程互不干扰
- ✅ 前端能正确显示所有报警

