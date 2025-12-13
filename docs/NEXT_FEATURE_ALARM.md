# 下一步功能开发：wisefido-alarm 服务

## 📋 功能概述

`wisefido-alarm` 服务负责对融合后的实时数据进行报警规则评估，生成报警事件并更新缓存。

## 🎯 核心功能

### 1. 数据输入
- **Redis 缓存**：`vital-focus:card:{card_id}:realtime`（融合后的实时数据）
- **PostgreSQL**：`alarm_cloud`（租户级别报警策略）、`alarm_device`（设备级别报警配置）

### 2. 报警评估
- **生命体征异常**：心率异常、呼吸率异常、呼吸暂停
- **行为事件**：跌倒、离床、床上坐起等
- **设备状态**：设备离线、低电量、设备故障等

### 3. 数据输出
- **PostgreSQL**：`alarm_events` 表（报警事件记录）
- **Redis 缓存**：`vital-focus:card:{card_id}:alarms`（报警数据，TTL: 30秒）

## 📊 数据流

```
Redis (vital-focus:card:{card_id}:realtime)
    ↓ 读取融合后的实时数据
wisefido-alarm 服务
    ├─ 读取报警规则（alarm_cloud, alarm_device）
    ├─ 评估报警条件
    ├─ 生成报警事件
    ├─→ PostgreSQL (alarm_events)
    └─→ Redis (vital-focus:card:{card_id}:alarms)
```

## 🔍 报警规则（根据 alarm_rule.md）

### 1. 心率异常（Heart Rate, HR）
- **EMERGENCY**：`[0-44]` 或 `[116-∞]`，持续 ≥ 60秒
- **WARNING**：`[45-54]` 或 `[96-115]`，持续 ≥ 300秒
- **Normal**：`[55-95]`

### 2. 呼吸率异常（Respiratory Rate, RR）
- **EMERGENCY**：`[0-7]` 或 `[27-∞]`，持续 ≥ 60秒
- **WARNING**：`[8-9]` 或 `[24-26]`，持续 ≥ 300秒
- **Normal**：`[10-23]`

### 3. 呼吸暂停（Apnea Hypopnea）
- 检测呼吸率是否为 0 或接近 0
- 持续时间阈值（从 `alarm_device.monitor_config.alarms` 读取）

### 4. 跌倒（Fall）
- 从姿态数据中检测跌倒姿态
- 支持 AI 智能评估（事件1：防止雷达漏报）

### 5. 离床（Left Bed）
- 检测床状态变化（`bed_status` 从 "in bed" 变为 "out of bed"）
- 持续时间阈值

### 6. 设备状态报警
- 设备离线、低电量、设备故障等

## 🏗️ 实现步骤

### 阶段 1：项目基础结构 ✅ 部分完成
- [x] 项目结构已创建
- [x] Models 已定义（realtime_data, alarm_event, alarm_config）
- [x] Config 已定义
- [ ] 完善配置加载逻辑
- [ ] 创建 main.go 入口

### 阶段 2：Repository 层
- [ ] 实现 `alarm_cloud` 表操作
- [ ] 实现 `alarm_device` 表操作
- [ ] 实现 `alarm_events` 表操作（写入报警事件）

### 阶段 3：Consumer 层
- [ ] 实现 Redis 缓存消费者
  - 方案 A：定期轮询 `vital-focus:card:{card_id}:realtime`
  - 方案 B：监听 Redis Streams（如果 sensor-fusion 发布事件）
- [ ] 实现缓存管理器（读取和更新报警缓存）

### 阶段 4：Evaluator 层
- [ ] 实现生命体征评估器（HR/RR 异常、呼吸暂停）
- [ ] 实现行为事件评估器（跌倒、离床）
- [ ] 实现设备状态评估器（离线、低电量、故障）

### 阶段 5：Service 层
- [ ] 实现报警服务主逻辑
- [ ] 整合 Repository、Consumer、Evaluator
- [ ] 实现报警事件生成和持久化

### 阶段 6：测试
- [ ] 单元测试（Repository、Evaluator）
- [ ] 集成测试（完整流程）
- [ ] 性能测试

## 📝 当前状态

### ✅ 已完成
- 项目结构已创建
- Models 已定义（部分）

### ⏳ 待实现
- Repository 层
- Consumer 层
- Evaluator 层
- Service 层
- 测试

## 🔗 相关文档

- `docs/13_Alarm_Fusion_Implementation.md` - 详细设计文档
- `docs/alarm_rule.md` - 报警规则详细说明
- `docs/system_architecture_complete.md` - 系统架构文档

## 🚀 开始实现

建议从 Repository 层开始，逐步实现各层功能。

