# 实施顺序和规划总结

## 📋 实施顺序（已完成）

### ✅ 阶段 1：准备报警处理依赖（基础层）

**目标**：创建必要的 Repository 和 Model，为后续功能提供基础

**实施内容**：
1. 创建 `wisefido-sensor-fusion/internal/models/alarm_event.go`
   - 定义 `AlarmEvent` 结构
   - 定义 `TriggerData` 结构

2. 创建 `wisefido-sensor-fusion/internal/repository/alarm_events.go`
   - 实现 `AlarmEventsRepository`
   - 实现 `CreateAlarmEvent` 方法

3. 创建 `wisefido-sensor-fusion/internal/repository/alarm_device.go`
   - 实现 `AlarmDeviceRepository`
   - 实现 `GetAlarmDeviceConfig` 方法

**为什么先做这个**：
- 所有后续功能都依赖这些基础结构
- 必须先完成，才能进行后续开发

---

### ✅ 阶段 2：实现报警处理工具函数

**目标**：创建可复用的报警处理函数

**实施内容**：
1. 创建 `wisefido-sensor-fusion/internal/alarm/alarm_handler.go`
   - `IsDeviceDirectAlarm` - 判断是否是设备直接报警
   - `GetAlarmCategory` - 获取报警分类（safety/clinical/behavioral/device）
   - `GetAlarmLevel` - 获取报警级别（EMERGENCY/WARNING）
   - `IsAlarmEnabled` - 判断报警是否在设备配置中启用
   - `BuildDeviceAlarmEvent` - 构建设备直接报警事件
   - `CreateDeviceAlarm` - 创建报警事件（完整流程）

**为什么先做这个**：
- 提供可复用的函数，便于测试和维护
- 逻辑集中，便于后续集成

---

### ✅ 阶段 3：集成到数据流中

**目标**：将报警处理集成到数据融合流程中

**实施内容**：
1. 更新 `IoTDataMessage` 模型
   - 添加 `EventType` 字段（可选）

2. 更新 `StreamConsumer` 结构
   - 添加 `alarmEventsRepo`, `alarmDeviceRepo`, `alarmHandler` 字段

3. 更新 `NewStreamConsumer` 构造函数
   - 接收报警处理依赖参数

4. 在 `processMessage` 中添加报警检测
   - 检测 `iotData.EventType`
   - 调用 `alarmHandler.CreateDeviceAlarm`

5. 更新 `wisefido-data-transformer`
   - 发布消息时包含 `event_type`（如果存在）

**为什么先做这个**：
- 核心功能，将报警处理集成到数据流
- 必须完成，才能实现设备直接报警处理

---

### ✅ 阶段 4：更新服务初始化

**目标**：确保服务能正确启动和运行

**实施内容**：
1. 更新 `FusionService.NewFusionService`
   - 创建 `AlarmEventsRepository`
   - 创建 `AlarmDeviceRepository`
   - 创建 `AlarmHandler`
   - 传递给 `NewStreamConsumer`

2. 修复编译错误
   - 添加缺失的导入
   - 修复函数调用参数

**为什么先做这个**：
- 确保服务能正确启动
- 必须完成，才能测试功能

---

### ✅ 阶段 5：文档和测试准备

**目标**：完善文档，为测试做准备

**实施内容**：
1. 创建架构文档
   - `docs/ALARM_PROCESSING_ARCHITECTURE.md`

2. 更新 README
   - `wisefido-alarm/README.md` - 明确职责划分

3. 创建实施总结
   - `docs/IMPLEMENTATION_SUMMARY.md`
   - `docs/IMPLEMENTATION_ORDER.md`（本文档）

**为什么先做这个**：
- 文档帮助理解架构和职责
- 为后续测试提供指导

---

## 🔄 依赖关系图

```
阶段 1 (Repository/Model)
    ↓
阶段 2 (工具函数)
    ↓
阶段 3 (集成)
    ↓
阶段 4 (服务初始化)
    ↓
阶段 5 (文档)
```

## 📝 关键设计决策

### 1. 为什么在 wisefido-sensor-fusion 中处理设备直接报警？

**理由**：
- ✅ 数据流中处理，实时性好（0延迟）
- ✅ 不需要额外查询数据库（从消息中直接获取 `event_type`）
- ✅ 逻辑简单，职责清晰

**替代方案**：
- ❌ 在 `wisefido-alarm` 中处理：需要轮询 `iot_timeseries` 表，有延迟
- ❌ 在 `wisefido-data-transformer` 中处理：职责可能混乱

### 2. 为什么在 wisefido-alarm 中处理云端事件报警？

**理由**：
- ✅ 规则评估逻辑集中
- ✅ 支持复杂的状态管理和定时器
- ✅ 可以访问融合后的实时数据

### 3. 为什么 wisefido-data-transformer 要发布 event_type？

**理由**：
- ✅ 避免在 `wisefido-sensor-fusion` 中查询 `iot_timeseries` 表
- ✅ 减少数据库负载
- ✅ 提高实时性

## ⚠️ 注意事项

1. **错误处理**：
   - 报警创建失败不应该影响数据融合流程
   - 只记录警告日志，继续处理

2. **配置检查**：
   - 必须检查设备报警配置
   - 只有启用的报警才创建记录

3. **性能考虑**：
   - 报警检测和处理应该是轻量级的
   - 不影响数据融合性能

4. **代码复用**：
   - 目前复制了报警 Repository 代码
   - 后续可以考虑提取到 `owl-common`

## 🎯 下一步

### 待测试 ⏳

1. **测试设备直接报警**
   - 发送包含 `event_type: "Fall"` 的测试数据
   - 验证是否能正确创建 `alarm_events` 记录
   - 验证前端是否能显示报警

2. **验证云端事件报警**
   - 验证事件1-4 不受影响
   - 验证两个报警处理流程互不干扰

3. **性能测试**
   - 验证报警处理不影响数据融合性能
   - 验证数据库负载在可接受范围内

## 📊 实施统计

- **新建文件**：7 个
- **修改文件**：5 个
- **新增代码行数**：约 500 行
- **编译状态**：✅ 通过
- **测试状态**：⏳ 待测试

