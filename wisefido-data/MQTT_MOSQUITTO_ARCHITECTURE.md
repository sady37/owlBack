# v1.5 MQTT 架构说明：Mosquitto 作为 MQTT Broker

## 📋 概述

**是的，v1.5 统一使用 Mosquitto 作为 MQTT Broker**。

Mosquitto 是一个开源的 MQTT 消息代理（Message Broker），用于接收和转发 MQTT 消息。

---

## 🏗️ v1.5 架构中的 MQTT

### 1. Docker Compose 配置

```yaml
mqtt:
  image: eclipse-mosquitto:2.0
  container_name: owl-mqtt
  ports:
    - "1883:1883"    # MQTT 协议端口
    - "9001:9001"    # WebSocket 端口
  volumes:
    - ./mqtt/config:/mosquitto/config
    - ./mqtt/data:/mosquitto/data
    - ./mqtt/log:/mosquitto/log
```

**说明**：
- **Mosquitto** 是 MQTT Broker（消息代理）
- **端口 1883**：标准 MQTT 协议端口
- **端口 9001**：WebSocket 端口（用于 Web 客户端）

---

## 🔄 MQTT 工作原理

### MQTT 是什么？

**MQTT (Message Queuing Telemetry Transport)** 是一个轻量级的消息传输协议，用于 IoT 设备通信。

### 架构角色

```
设备（Publisher） → MQTT Broker（Mosquitto） → 服务（Subscriber）
    发布消息             转发消息                   订阅消息
```

**流程**：
1. **设备发布消息**：设备通过 MQTT 协议发送消息到 Mosquitto
2. **Mosquitto 转发**：Mosquitto 根据主题（Topic）将消息转发给订阅者
3. **服务订阅消息**：服务订阅特定主题，接收消息并处理

---

## 📊 v1.5 中的 MQTT 使用

### 1. 设备层

**Radar 设备**：
- 通过 MQTT over TLS 连接到 `ql-mosquitto`
- 发布数据到主题（如 `radar/device001/data`）

**Sleepace 设备**：
- 通过 `xs-services`（享睡Java服务集）连接到 `mosquitto`
- 发布数据到主题（如 `sleepace-57136`）

### 2. 服务层

**wisefido-radar**：
- 使用 `owl-common/mqtt/client.go` 作为 MQTT 客户端
- 连接到 Mosquitto（默认：`tcp://localhost:1883`）
- 订阅主题：`radar/+/data`
- 处理消息并发布到 Redis Streams

**wisefido-sleepace**：
- 使用 `owl-common/mqtt/client.go` 作为 MQTT 客户端
- 连接到 MQTT Broker（默认：`mqtt://47.90.180.176:1883`，Sleepace 厂家的 MQTT）
- 订阅主题：`sleepace-57136`（Sleepace 厂家提供的主题）
- 处理消息并发布到 Redis Streams

**wisefido-data**：
- 目前**不直接使用 MQTT**
- 只提供 HTTP API
- 可以通过 HTTP API 被其他服务调用

---

## 🔍 架构图中的 MQTT 流程

根据你提供的架构图：

```
rardar 设备
    ↓ (MQTT over TLS)
ql-mosquitto (MQTT Broker)
    ↓
device-service / stream-service

sleepboard 设备
    ↓ (TLS)
xs-services (享睡Java服务集)
    ↓
mosquitto (MQTT Broker)
    ↓
device-service / stream-service
```

**说明**：
- **ql-mosquitto** 和 **mosquitto** 都是 MQTT Broker（Mosquitto 实例）
- 设备通过 MQTT 协议连接到这些 Broker
- Broker 将消息转发给订阅的服务（`device-service`、`stream-service`）

---

## 🎯 与 Sleepace 报告下载的关系

### 当前状态

**wisefido-data**：
- ❌ 不直接使用 MQTT
- ✅ 提供 HTTP API（手动触发下载）

**wisefido-sleepace**：
- ✅ 使用 MQTT 订阅 Sleepace 设备消息
- ✅ 处理实时数据、睡眠阶段等
- ❌ 目前不处理报告下载触发

### 如果实现 MQTT 触发下载

**选项 1：在 wisefido-sleepace 中实现**（推荐）
```
Sleepace 设备 → Sleepace 厂家 MQTT → wisefido-sleepace
    ↓
收到 analysis 类型消息
    ↓
调用 wisefido-data HTTP API
    ↓
触发报告下载
```

**选项 2：在 wisefido-data 中实现**
```
Sleepace 设备 → Sleepace 厂家 MQTT → wisefido-data
    ↓
收到 analysis 类型消息
    ↓
直接调用 Service.DownloadReport
```

---

## 📝 配置说明

### wisefido-sleepace 配置

```go
// internal/config/config.go
cfg.MQTT.Broker = getEnv("MQTT_BROKER", "mqtt://47.90.180.176:1883")
cfg.MQTT.ClientID = getEnv("MQTT_CLIENT_ID", "wisefido-sleepace")
cfg.MQTT.Username = getEnv("MQTT_USERNAME", "wisefido")
cfg.MQTT.Password = getEnv("MQTT_PASSWORD", "")

cfg.Sleepace.Topic = getEnv("SLEEPACE_MQTT_TOPIC", "sleepace-57136")
```

**说明**：
- 默认连接到 Sleepace 厂家的 MQTT Broker（`47.90.180.176:1883`）
- 订阅主题：`sleepace-57136`（Sleepace 厂家提供的主题）
- 可以配置为连接到本地的 Mosquitto

### wisefido-radar 配置

```go
// internal/config/config.go
cfg.MQTT.Broker = getEnv("MQTT_BROKER", "tcp://localhost:1883")
cfg.MQTT.ClientID = getEnv("MQTT_CLIENT_ID", "wisefido-radar")
```

**说明**：
- 默认连接到本地 Mosquitto（`localhost:1883`）
- 订阅主题：`radar/+/data`

---

## ✅ 总结

### 1. Mosquitto 是什么？

**Mosquitto 是 MQTT Broker（消息代理）**，用于：
- 接收设备发布的 MQTT 消息
- 根据主题（Topic）将消息转发给订阅者
- 管理 MQTT 连接和会话

### 2. v1.5 是否统一使用 Mosquitto？

**是的**：
- ✅ Docker Compose 中配置了 Mosquitto（`eclipse-mosquitto:2.0`）
- ✅ 服务使用 `owl-common/mqtt/client.go` 连接到 MQTT Broker
- ✅ 可以连接到本地 Mosquitto 或外部 MQTT Broker

### 3. 这是对接 MQTT 的吗？

**是的**：
- ✅ Mosquitto 是 MQTT 协议的实现
- ✅ 设备和服务都通过 MQTT 协议连接到 Mosquitto
- ✅ 这是标准的 MQTT 架构

### 4. 与 Sleepace 报告下载的关系

**当前**：
- `wisefido-data` 不直接使用 MQTT
- 只提供 HTTP API（手动触发下载）

**如果实现 MQTT 触发下载**：
- 可以在 `wisefido-sleepace` 中实现（已有 MQTT 消费者）
- 或在 `wisefido-data` 中添加 MQTT 处理模块
- 两种方案都可以连接到 Mosquitto 或 Sleepace 厂家的 MQTT

---

## 🔗 相关文档

- [MQTT 客户端设计](../docs/02_MQTT_Client_Design.md)
- [系统架构完整说明](../docs/system_architecture_complete.md)
- [Sleepace 数据流 v1.5](../docs/10_Sleepace_Data_Flow_v1.5.md)

