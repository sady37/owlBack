# MQTT 客户端设计说明

## 📋 概述

`owl-common/mqtt/client.go` 是一个封装了 Eclipse Paho MQTT 客户端的通用库，为所有后端服务提供统一的 MQTT 连接、订阅、发布功能。

---

## 🏗️ 架构设计

### 1. 设计目标

- **统一封装**: 所有服务使用相同的 MQTT 客户端接口
- **简化使用**: 隐藏底层实现细节，提供简洁的 API
- **自动重连**: 支持自动重连机制
- **错误处理**: 统一的错误处理方式

### 2. 核心组件

```go
// MessageHandler 消息处理函数类型
type MessageHandler func(topic string, payload []byte) error

// Client MQTT客户端封装
type Client struct {
    client mqtt.Client  // 底层 Paho MQTT 客户端
    config *config.MQTTConfig  // 配置信息
}
```

---

## 🔧 实现细节

### 1. 客户端创建 (`NewClient`)

```go
func NewClient(cfg *config.MQTTConfig, logger interface{}) (*Client, error)
```

**功能**:
- 创建 MQTT 客户端连接选项
- 配置 Broker 地址、ClientID、认证信息
- 启用自动重连
- 建立连接

**配置项**:
```go
type MQTTConfig struct {
    Broker   string  // MQTT Broker 地址，如 "tcp://localhost:1883"
    ClientID string  // 客户端唯一标识
    Username string  // 用户名（可选）
    Password string  // 密码（可选）
    QoS      byte    // 默认 QoS 级别
}
```

**关键设置**:
- `SetAutoReconnect(true)` - 自动重连
- `SetCleanSession(true)` - 清理会话（每次连接都是新会话）

**示例**:
```go
cfg := &config.MQTTConfig{
    Broker:   "tcp://localhost:1883",
    ClientID: "wisefido-radar-001",
    Username: "admin",
    Password: "password",
}

client, err := mqttcommon.NewClient(cfg, logger)
if err != nil {
    log.Fatal(err)
}
```

---

### 2. 订阅主题 (`Subscribe`)

```go
func (c *Client) Subscribe(topic string, qos byte, handler MessageHandler) error
```

**功能**:
- 订阅指定的 MQTT 主题
- 注册消息处理函数
- 支持通配符主题（如 `radar/+/data`）

**参数**:
- `topic`: 主题名称，支持通配符
  - `+`: 单级通配符（如 `radar/+/data` 匹配 `radar/device1/data`）
  - `#`: 多级通配符（如 `radar/#` 匹配所有以 `radar/` 开头的主题）
- `qos`: 服务质量级别（0, 1, 2）
  - `0`: 最多一次（可能丢失）
  - `1`: 至少一次（可能重复）
  - `2`: 恰好一次（保证）
- `handler`: 消息处理函数

**消息处理函数签名**:
```go
type MessageHandler func(topic string, payload []byte) error
```

**示例**:
```go
// 订阅雷达数据主题
err := client.Subscribe("radar/+/data", 1, func(topic string, payload []byte) error {
    // 处理消息
    fmt.Printf("Received message on topic: %s\n", topic)
    fmt.Printf("Payload: %s\n", string(payload))
    return nil
})
```

**错误处理**:
- 如果消息处理函数返回错误，会在控制台打印（当前实现）
- 不会中断订阅，继续处理后续消息

---

### 3. 发布消息 (`Publish`)

```go
func (c *Client) Publish(topic string, qos byte, retained bool, payload []byte) error
```

**功能**:
- 向指定主题发布消息
- 支持保留消息（retained message）

**参数**:
- `topic`: 主题名称
- `qos`: 服务质量级别
- `retained`: 是否保留消息
  - `true`: Broker 会保留最后一条消息，新订阅者会立即收到
  - `false`: 不保留
- `payload`: 消息内容（字节数组）

**示例**:
```go
// 发布命令到设备
payload := []byte(`{"command": "restart", "timestamp": 1234567890}`)
err := client.Publish("radar/device1/command", 1, false, payload)
```

**使用场景**:
- 发送设备控制命令
- OTA 升级指令
- 配置更新

---

### 4. 取消订阅 (`Unsubscribe`)

```go
func (c *Client) Unsubscribe(topics ...string) error
```

**功能**:
- 取消订阅一个或多个主题

**示例**:
```go
// 取消订阅单个主题
err := client.Unsubscribe("radar/+/data")

// 取消订阅多个主题
err := client.Unsubscribe("radar/+/data", "radar/+/command")
```

---

### 5. 断开连接 (`Disconnect`)

```go
func (c *Client) Disconnect()
```

**功能**:
- 优雅断开 MQTT 连接
- 等待 250ms 确保消息发送完成

**示例**:
```go
// 服务停止时断开连接
defer client.Disconnect()
```

---

### 6. 连接状态检查 (`IsConnected`)

```go
func (c *Client) IsConnected() bool
```

**功能**:
- 检查客户端是否已连接到 Broker

**示例**:
```go
if client.IsConnected() {
    fmt.Println("MQTT client is connected")
} else {
    fmt.Println("MQTT client is disconnected")
}
```

---

## 📊 在 wisefido-radar 服务中的使用

### 1. 初始化

```go
// 在 service/radar.go 中
mqttClient, err := mqttcommon.NewClient(&cfg.MQTT, logger)
if err != nil {
    return nil, fmt.Errorf("failed to connect to MQTT: %w", err)
}
```

### 2. 订阅数据主题

```go
// 在 consumer/mqtt_consumer.go 中
func (c *MQTTConsumer) Start(ctx context.Context) error {
    // 订阅雷达数据主题: "radar/+/data"
    if err := c.mqttClient.Subscribe(
        c.config.Radar.Topics.Data,  // "radar/+/data"
        1,                            // QoS 1
        c.handleMessage,             // 消息处理函数
    ); err != nil {
        return fmt.Errorf("failed to subscribe to data topic: %w", err)
    }
    
    // 等待上下文取消
    <-ctx.Done()
    return nil
}
```

### 3. 处理消息

```go
func (c *MQTTConsumer) handleMessage(topic string, payload []byte) error {
    // 1. 解析主题，提取设备ID
    // 主题格式: radar/{device_id}/data
    parts := strings.Split(topic, "/")
    deviceIdentifier := parts[1]
    
    // 2. 解析消息内容
    var mqttData map[string]interface{}
    json.Unmarshal(payload, &mqttData)
    
    // 3. 验证设备权限
    device, err := c.deviceRepo.GetDeviceBySerialNumber(deviceIdentifier)
    
    // 4. 处理数据并发布到 Redis Streams
    // ...
    
    return nil
}
```

### 4. 停止服务

```go
func (c *MQTTConsumer) Stop(ctx context.Context) error {
    // 取消订阅
    if err := c.mqttClient.Unsubscribe(c.config.Radar.Topics.Data); err != nil {
        c.logger.Error("Failed to unsubscribe", zap.Error(err))
    }
    return nil
}
```

---

## 🔍 MQTT 主题设计

### 主题命名规范

```
{device_type}/{device_id}/{message_type}
```

**示例**:
- `radar/device001/data` - 雷达设备数据
- `radar/device001/command` - 雷达设备命令
- `radar/device001/ota` - 雷达设备 OTA
- `sleepace/device002/data` - 睡眠垫设备数据

### 通配符使用

- `radar/+/data` - 订阅所有雷达设备的数据
- `radar/#` - 订阅所有雷达设备的所有消息类型

---

## ⚙️ 配置示例

### 环境变量配置

```bash
# MQTT Broker 地址
MQTT_BROKER=tcp://localhost:1883

# 客户端ID（每个服务实例应该唯一）
MQTT_CLIENT_ID=wisefido-radar-001

# 认证信息（可选）
MQTT_USERNAME=admin
MQTT_PASSWORD=password
```

### 代码配置

```go
cfg := &config.MQTTConfig{
    Broker:   "tcp://localhost:1883",
    ClientID: "wisefido-radar-001",
    Username: "admin",
    Password: "password",
}
```

---

## 🚀 最佳实践

### 1. 客户端ID 唯一性

每个服务实例应该使用唯一的 ClientID，避免冲突：

```go
// 使用 UUID 或时间戳确保唯一性
clientID := fmt.Sprintf("wisefido-radar-%s", uuid.New().String())
```

### 2. QoS 级别选择

- **QoS 0**: 用于实时性要求高、允许丢失的数据（如传感器实时数据）
- **QoS 1**: 用于需要保证送达但不允许重复的数据（如设备状态更新）
- **QoS 2**: 用于关键命令（如 OTA 升级指令）

### 3. 错误处理

消息处理函数中的错误不会中断订阅，但应该记录日志：

```go
func handleMessage(topic string, payload []byte) error {
    if err := processMessage(payload); err != nil {
        logger.Error("Failed to process message", zap.Error(err))
        // 返回错误会被记录，但不会中断订阅
        return err
    }
    return nil
}
```

### 4. 资源清理

服务停止时应该取消订阅并断开连接：

```go
defer func() {
    client.Unsubscribe("radar/+/data")
    client.Disconnect()
}()
```

---

## 🔄 与其他服务的集成

### 1. wisefido-sleepace 服务

使用相同的 MQTT 客户端，订阅不同的主题：

```go
// 订阅睡眠垫数据主题
client.Subscribe("sleepace/+/data", 1, handleSleepaceMessage)
```

### 2. 命令发布

服务可以向设备发布命令：

```go
// 发布重启命令
command := map[string]interface{}{
    "command": "restart",
    "timestamp": time.Now().Unix(),
}
payload, _ := json.Marshal(command)
client.Publish("radar/device001/command", 1, false, payload)
```

---

## 📝 总结

### 优点

1. **统一接口**: 所有服务使用相同的 MQTT 客户端封装
2. **自动重连**: 网络中断时自动重连
3. **简洁 API**: 隐藏底层实现，易于使用
4. **灵活配置**: 支持环境变量和代码配置

### 改进建议

1. **日志集成**: 当前错误使用 `fmt.Printf`，应该集成到 logger
2. **重连回调**: 可以添加重连成功/失败的回调函数
3. **连接池**: 如果需要多个连接，可以考虑连接池
4. **消息队列**: 可以添加消息队列缓冲，避免消息丢失

---

## 📚 相关文档

- [Eclipse Paho MQTT Go Client](https://github.com/eclipse/paho.mqtt.golang)
- [MQTT 协议规范](https://mqtt.org/)
- [wisefido-radar 服务实现](../wisefido-radar/internal/consumer/mqtt_consumer.go)

