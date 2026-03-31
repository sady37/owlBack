# IoT 数据获取方案评估报告

> 基于 owlBack 源码分析，2026-03-31

## 1. 背景

当前需要从 Owl 平台获取物联网硬件设备上报的数据（上行），以及向设备下发指令或配置（下行）。本文对两种可行方案进行分析评估，并给出建议。

### 系统现有数据通路

```
上行: Device -> MQTT -> Gateway(qinglan/sleepace) -> Redis Streams(iot:*:stream) -> 下游服务
下行: Client -> wisefido-data(REST) -> wisefido-qinglan(HTTP) -> MQTT -> Device
```

上行与下行使用完全不同的技术栈：上行通过 Redis Streams 异步流转，下行通过 HTTP + MQTT 同步请求-响应。

### 核心概念

| 概念 | 说明 |
|------|------|
| device_uid | 设备物理标识（硬件唯一编号），所有 iot:\*:stream 消息必带 |
| device_id | 业务 UUID（devices 表主键），wisefido-data API 以此为参数 |
| iot:\*:stream | 设备级原始数据流，以 device_uid 为主键 |

### 现有上行链路

```
Device --MQTT--> Gateway(qinglan) --XADD--> iot:monitor:stream (30s TTL)
                                  --XADD--> iot:event:stream   (24h TTL)
                                  --XADD--> iot:alarm:stream   (24h TTL)
                                  --XADD--> iot:stat:stream    (5min TTL)
                                  --XADD--> iot:auth:stream    (24h TTL)
```

每条 Stream 消息固定包含以下字段：

| 字段 | 说明 | 是否必有 |
|------|------|---------|
| device_uid | 设备物理标识 | 是 |
| device_type | "Radar" / "SleepPad" 等 | 是 |
| tenant_id | 租户 ID | 是 |
| timestamp | 时间戳 | 是 |
| topic_type | "monitor" / "event" / "alarm" / "stat" | 是 |
| category | 数据分类 | 是 |
| dataValue | JSON 数组，具体观测数据 | 是 |
| device_id | 业务 UUID | 条件（可能为空） |

### 现有下行链路

设备控制强绑定 MQTT，完整链路：

```
wisefido-data:8080                    wisefido-qinglan:8081              MQTT Broker
     |                                      |                              |
     |--- HTTP GET/PUT/POST ------------->  |--- MQTT publish ----------> |---> Device
     |                                      |    (.../get topic)          |
     |                                      |<-- MQTT response <--------- |<--- Device
     |                                      |    (.../post topic)         |
     |                                      |--> Redis cmd:response:{id}  |
     |                                      |--> 轮询 100ms / 10-30s超时  |
     |<-- HTTP response ------------------- |                              |
```

可用的下行 API：

| 端点 | 参数 | 用途 |
|------|------|------|
| `GET /api/v1/radar/devices/{uid}/properties` | device_uid | 读设备属性 |
| `PUT /api/v1/radar/devices/{uid}/properties` | device_uid | 写设备属性 |
| `POST /api/v1/radar/devices/{uid}/function` | device_uid | 重启/清数据 |
| `POST /api/v1/radar/devices/{uid}/subscribe` | device_uid | 启动实时数据订阅 |
| `GET /api/v1/radar/devices/{uid}/status` | device_uid | 设备在线状态 |

---

## 2. 方案分析

### 方案一：新建独立网关服务

参照 wisefido-qinglan / wisefido-sleepace 模式，新建 wisefido-{device} 服务。

```
上行: Device -> MQTT -> 新网关 -> iot:*:stream -> 下游无需改动
下行: wisefido-data -> HTTP -> 新网关 -> MQTT -> Device
```

**上行：** 自行连接 MQTT、解码设备协议、按 IoTStreamMessage 格式写入 iot:\*:stream。

**下行：** 新网关暴露 HTTP API，接收调用后转为 MQTT 指令下发到设备。上游调用方（wisefido-data 或外部服务）需新增对应 HTTP Client。

可复用 owl-common 组件：

| 包 | 用途 |
|----|------|
| owl-common/mqtt | MQTT 客户端封装 |
| owl-common/redis | Stream 发布/消费、IoTStreamMessage 构建 |
| owl-common/observation | 60+ 标准化字段名 |
| owl-common/config | MQTT/Redis/Database 配置 |
| owl-common/alarm | 告警类型、严重级别定义 |

**特点：**
- 上下行统一：上行走 MQTT -> Redis Streams，下行走 HTTP -> MQTT，与现有网关模式一致
- 需要开发完整的设备协议解码/编码层
- 需要实现设备连接管理（认证、心跳、订阅续期、断线检测）
- 需要新增一个独立服务的部署和运维
- 如需 wisefido-data 调用，还需在 data 侧新增 Client

### 方案二：消费 Redis Streams + 调用现有 HTTP API

不新建网关，利用已有基础设施，新建一个服务完成双向通信。

```
上行: qinglan 已写入 -> iot:*:stream -> [新 consumer group] -> 新服务
下行: 新服务 -> HTTP -> qinglan:8081 -> MQTT -> Device
```

**上行：** 新服务创建独立 consumer group 消费 iot:monitor/event/alarm:stream，每条消息自带 device_uid，可直接定位设备。不同 consumer group 之间完全隔离，不影响现有服务。

**下行：** 新服务通过 HTTP 调用 wisefido-qinglan:8081 的现有 API（以 device_uid 为参数）。qinglan 负责将 HTTP 请求转为 MQTT 指令并等待设备响应。

**特点：**
- 上下行实现方式不统一：上行通过 Redis Streams 消费，下行通过 HTTP 调用
- 不需要处理 MQTT 连接和设备协议，复杂度低
- 依赖 wisefido-qinglan 的可用性——qinglan 不可用时下行中断
- 仍需新建一个服务来承载消费逻辑和对外接口，但该服务不涉及协议层，复杂度远低于方案一
- 对现有系统零改动

---

## 3. 方案二的关键问题

### 3.1 上下行不统一

这是方案二最本质的架构问题。上行和下行走完全不同的通道：

| 方向 | 通道 | 协议 | 特点 |
|------|------|------|------|
| 上行 | Redis Streams | XREADGROUP | 异步、被动消费 |
| 下行 | HTTP -> qinglan | HTTP + MQTT | 同步、请求-响应 |

这意味着新服务需要同时维护两套通信机制：一个 Redis Stream 消费者和一个 HTTP 客户端。对调用方来说，读数据和控制设备是两种完全不同的交互模式。

根源在于：**当前系统的 Redis Streams 是单向的**（仅上行），没有"指令 Stream"供消费者投递下行指令。要通过 Redis Streams 统一上下行，需要新建 command stream 并让 qinglan 或新网关消费——这实质上又回到了方案一的开发量。

### 3.2 依赖 qinglan

下行完全依赖 wisefido-qinglan 的 HTTP API，意味着：

- qinglan 服务不可用时，下行能力全部丧失
- 下行端点、参数格式受 qinglan 现有接口约束，无法自定义
- qinglan API 面向 Radar 设备设计（endpoint 路径中含 `/radar/`），语义上不够通用

### 3.3 仍需新增服务

方案二并非"零服务"方案。实际落地时，仍需要一个服务来：

- 持续消费 Redis Streams（长驻进程）
- 对外暴露统一的 API（供第三方调用上行查询和下行指令）
- 管理 consumer group 的 ACK 和异常恢复

与方案一的区别在于：该服务**不涉及 MQTT 连接和设备协议**，复杂度主要在业务逻辑层面。

---

## 4. 综合对比

| 维度 | 方案一：新建网关 | 方案二：Redis + HTTP |
|------|----------------|---------------------|
| 适用场景 | 接入新设备（完整双向通信） | 获取已有设备数据 + 控制已有设备 |
| 上行实现 | MQTT 订阅 + 协议解码 + 写 Stream | 消费已有 iot:\*:stream |
| 下行实现 | 自建 HTTP API + MQTT 发布 | 调用 qinglan 已有 HTTP API |
| 上下行一致性 | 统一（均经 MQTT） | 不统一（上行 Redis / 下行 HTTP） |
| 协议层开发 | 需要（解码/编码设备协议） | 不需要 |
| 设备连接管理 | 需要（认证、心跳、续订） | 不需要（由 qinglan 管理） |
| 新增服务 | 是（网关服务） | 是（消费+代理服务） |
| 新服务复杂度 | 高（协议 + 连接 + Stream + HTTP） | 低（Stream 消费 + HTTP 调用） |
| 对现有系统改动 | data 需新增 Client | 无 |
| 下行可用性 | 独立（自有 MQTT） | 依赖 qinglan |
| 架构一致性 | 最佳（与现有网关模式统一） | 可接受（混合模式） |
| 数据完整性 | 全量 | 全量（iot:\*:stream） |
| 总开发量 | 大 | 小 |

---

## 5. 建议

两种方案的选择取决于实际需求的定位：

### 如果目标是"接入新型 IoT 设备"

采用**方案一**。新建独立网关是当前系统的标准模式（qinglan 对 Radar、sleepace 对可穿戴），架构统一、上下行一致、故障隔离。参照 wisefido-sleepace 代码结构可快速搭建。开发量较大，但长期可维护性最好。

### 如果目标是"获取和控制现有设备"

采用**方案二**。通过消费 iot:\*:stream 获取数据、调用 qinglan HTTP API 下发指令，可以在不修改任何现有代码的前提下实现双向通信。虽然上下行实现不统一，但开发量最小、风险最低。需要注意的是仍需新建一个轻量服务来承载消费逻辑，且下行依赖 qinglan 的可用性。

```
                        iot:monitor:stream ---+
Device -> MQTT -> qinglan -> iot:event:stream  ---> [新 consumer group] ---> 新服务 ---> 对外API
                        iot:alarm:stream   ---+
                                                                              |
                                                                              | HTTP (下行)
                                                                              v
                                                                       qinglan:8081
                                                                     /api/v1/radar/devices/{uid}/...
                                                                              |
                                                                              | MQTT
                                                                              v
                                                                           Device
```

### 关于上下行不统一的取舍

方案二的上下行不统一是一个架构妥协，但在"获取和控制现有设备"的场景下是合理的——它用最小的代价复用了现有系统的全部能力，而不是重新实现一遍已有的 MQTT 连接管理和协议编解码。如果后续对一致性有更高要求，可以在方案二的基础上逐步演进到方案一。
