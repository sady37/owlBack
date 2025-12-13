# Sleepace 数据来源说明

## 📋 背景

根据确认：
- **wisefido-radar 服务** - 已由其他同事实现（MQTT 订阅，发布到 Redis Streams）
- **wisefido-sleepace 服务** - **错误理解**，实际情况：
  - Sleepace 仍然是 **v1.0 的格式，保持不变**
  - 仍是由 **sleepace 厂家提供的服务**
  - 不是我们实现的 Go 服务

## 🔍 Sleepace 数据来源（v1.0 格式）

### 可能的数据来源

1. **Sleepace 厂家服务**（第三方服务）
   - 可能是 HTTP API
   - 可能是直接写入数据库
   - 可能是其他方式

2. **数据格式**
   - 保持 v1.0 格式不变
   - 可能已经标准化，或需要转换

## 📊 数据流（修正后）

### 雷达数据流（v1.5，已实现）

```
Radar 设备 → MQTT Broker（直接）
    ↓
wisefido-radar 服务（已实现）
    ├─ MQTT 订阅
    ├─ 设备验证
    └─→ Redis Streams (radar:data:stream)
```

### Sleepace 数据流（v1.0 格式，保持不变）

```
Sleepace 设备 → Sleepace 厂家服务（第三方）
    ├─ 可能通过 HTTP API
    ├─ 可能直接写入数据库
    └─ 或其他方式（待确认）
```

## ⚠️ 对 wisefido-data-transformer 的影响

### 需要确认的问题

1. **Sleepace 数据是否进入 Redis Streams？**
   - 如果 Sleepace 厂家服务直接写入数据库，则不需要消费 Streams
   - 如果 Sleepace 厂家服务发布到 Redis Streams，则需要消费

2. **Sleepace 数据格式**
   - 是否已经是标准格式？
   - 是否需要转换？

3. **数据写入方式**
   - Sleepace 数据是直接写入 `iot_timeseries` 表？
   - 还是需要经过 `wisefido-data-transformer` 转换？

## 🔧 可能的调整

### 方案 1：Sleepace 数据直接写入数据库

如果 Sleepace 厂家服务直接写入 `iot_timeseries` 表：

- **wisefido-data-transformer** 只需要消费 `radar:data:stream`
- 不需要消费 `sleepace:data:stream`
- Sleepace 数据可能已经是标准格式，或由厂家服务转换

### 方案 2：Sleepace 数据进入 Redis Streams

如果 Sleepace 厂家服务发布到 Redis Streams：

- **wisefido-data-transformer** 需要消费 `sleepace:data:stream`
- 需要实现 Sleepace 数据转换器
- 数据格式可能与 v1.0 相同

### 方案 3：Sleepace 数据通过其他方式

如果 Sleepace 数据通过其他方式（如 HTTP API）进入：

- 可能需要额外的适配服务
- 或直接写入数据库

## 📝 建议

1. **确认 Sleepace 数据来源**
   - 与 Sleepace 厂家确认数据提供方式
   - 确认数据格式和写入方式

2. **调整 wisefido-data-transformer**
   - 根据实际情况调整 Stream 消费逻辑
   - 如果 Sleepace 数据不需要转换，移除相关代码

3. **文档更新**
   - 更新架构文档，明确 Sleepace 数据流
   - 更新开发计划

