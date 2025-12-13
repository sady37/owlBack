# Sleepace v1.0 架构分析

## 📋 确认：Sleepace 数据流（v1.0）

根据 v1.0 代码分析，确认了 Sleepace 的架构：

### 数据流

```
Sleepad 设备 → Sleepace 厂家服务（第三方，有独立数据库）
    ↓
Sleepace 厂家服务 → MQTT Broker（厂家提供的 MQTT）
    ↓
wisefido-sleepace 服务（我们的服务）
    ├─ MQTT 订阅（订阅 Sleepace 厂家 MQTT）
    ├─ 数据处理
    └─→ MySQL 数据库（直接写入）
```

**关键点**：
1. **Sleepace 厂家服务**：第三方服务，有独立的数据库和 HTTP API
2. **MQTT 推送**：Sleepace 厂家服务通过 MQTT 推送数据到我们
3. **wisefido-sleepace 服务**：我们的服务，订阅 Sleepace 厂家的 MQTT，处理数据并写入 MySQL
4. **直接写入数据库**：数据直接写入 MySQL，**不经过 Redis Streams**

---

## 🔍 v1.0 实现细节

### 1. Sleepace 厂家服务

**HTTP API 地址**：`http://47.90.180.176:8080`

**主要功能**：
- 设备绑定/解绑
- 设备配置（心率模式、实时数据间隔、离床灵敏度等）
- 获取报警记录
- 获取睡眠报告
- **MQTT 推送配置**：设置 `pushType: "MQTT"`

### 2. wisefido-sleepace 服务

**功能**：
1. **HTTP API 服务**：提供 RESTful API 接口
2. **MQTT 订阅**：订阅 Sleepace 厂家提供的 MQTT 主题
3. **数据处理**：处理 MQTT 消息，写入 MySQL

**MQTT 配置**：
```yaml
mqtt:
  address: "mqtt://47.90.180.176:1883"  # Sleepace 厂家的 MQTT Broker
  username: "wisefido"
  password: "env(MQTT_PASSWORD)"
  client_id: "wisefido-sleepace-dev"
  topic_id: "sleepace-57136"  # 订阅的主题
```

**数据存储**：
- 直接写入 MySQL 数据库（`sleepace_*` 表）
- 不经过 Redis Streams
- 存储在多个表中：
  - `sleepace_realtime_record` - 实时数据
  - `sleepace_report` - 睡眠报告
  - `sleepace_connection_status` - 连接状态
  - 等等

---

## 📊 v1.5 的变化

### v1.0 架构

```
Sleepad → Sleepace 厂家服务（独立 DB + HTTP API + MQTT）
    ↓
Sleepace 厂家 MQTT Broker
    ↓
wisefido-sleepace 服务
    ├─ MQTT 订阅（厂家 MQTT）
    └─→ MySQL 数据库（直接写入）
```

### v1.5 架构（保持不变）

```
Sleepad → Sleepace 厂家服务（独立 DB + HTTP API + MQTT）
    ↓
Sleepace 厂家 MQTT Broker（不变）
    ↓
wisefido-sleepace 服务（保持不变，v1.0 格式）
    ├─ MQTT 订阅（厂家 MQTT）
    └─→ PostgreSQL 数据库（直接写入 iot_timeseries 表）
```

**关键变化**：
- ✅ **通信方式不变**：仍通过 Sleepace 厂家 MQTT
- ✅ **服务不变**：wisefido-sleepace 服务保持 v1.0 格式
- ⚠️ **数据库变化**：从 MySQL 迁移到 PostgreSQL
- ⚠️ **表结构变化**：可能需要将数据写入 `iot_timeseries` 表（标准化格式）

---

## 🎯 对 wisefido-data-transformer 的影响

### 结论：**不需要处理 Sleepace 数据**

**原因**：
1. Sleepace 数据**直接写入数据库**（不经过 Redis Streams）
2. wisefido-sleepace 服务**保持 v1.0 格式不变**
3. 数据可能已经标准化，或由 wisefido-sleepace 服务转换

### 需要确认的问题

1. **Sleepace 数据写入位置**
   - 是否写入 PostgreSQL `iot_timeseries` 表？
   - 还是写入其他表（如 v1.0 的 `sleepace_*` 表）？

2. **数据格式**
   - 如果写入 `iot_timeseries` 表，是否已经是标准格式？
   - 如果写入其他表，是否需要迁移到 `iot_timeseries`？

3. **数据转换**
   - 如果 Sleepace 数据写入其他表，是否需要转换到 `iot_timeseries`？
   - 转换是否在 wisefido-sleepace 服务中完成？

---

## 📝 建议

### 方案 A：Sleepace 数据直接写入 iot_timeseries 表（推荐）

如果 wisefido-sleepace 服务在 v1.5 中直接写入 `iot_timeseries` 表：

- ✅ `wisefido-data-transformer` **不需要处理** Sleepace 数据
- ✅ 只处理 Radar 数据
- ✅ 移除 Sleepace 相关代码

### 方案 B：Sleepace 数据写入其他表

如果 Sleepace 数据写入其他表（如 v1.0 的 `sleepace_*` 表）：

- ⚠️ 可能需要数据迁移或转换
- ⚠️ 需要考虑如何与 `iot_timeseries` 数据统一

---

## 🔧 代码调整建议

### 当前实现

`wisefido-data-transformer` 中：
- ✅ 消费 `radar:data:stream`（已实现）
- ⚠️ `sleepace:data:stream` 消费代码已注释（待确认）

### 调整建议

**如果确认 Sleepace 数据直接写入数据库**：
1. ✅ 移除 `sleepace:data:stream` 相关配置
2. ✅ 移除 Sleepace 数据消费代码
3. ✅ 只保留 Radar 数据处理

**如果 Sleepace 数据进入 Redis Streams**：
1. ⚠️ 启用 `sleepace:data:stream` 消费代码
2. ⚠️ 实现 Sleepace 数据转换器

---

## 📋 需要确认的清单

- [ ] Sleepace 数据在 v1.5 中写入哪个表？
- [ ] 如果写入 `iot_timeseries` 表，是否已经是标准格式？
- [ ] 如果写入其他表，是否需要迁移或转换？
- [ ] `wisefido-data-transformer` 是否需要处理 Sleepace 数据？

---

## 📚 参考文件

- `wisefido-backend/wisefido-sleepace/modules/sleepace_service.go` - Sleepace 服务逻辑
- `wisefido-backend/wisefido-sleepace/modules/borker.go` - MQTT 消息处理
- `wisefido-backend/wisefido-sleepace/sleepace-dev.yaml` - 配置文件

