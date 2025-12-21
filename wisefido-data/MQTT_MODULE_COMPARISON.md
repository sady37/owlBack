# MQTT 处理模块对比：v1.0 vs v1.5

## 📊 对比总结

| 项目 | v1.0 (wisefido-backend) | v1.5 (owlBack) |
|------|------------------------|----------------|
| **独立的 MQTT 处理模块** | ✅ 有 | ❌ 无 |
| **实现位置** | `modules/borker.go` | 无 |
| **MQTT 客户端库** | `github.com/eclipse/paho.mqtt.golang` | `owl-common/mqtt/client.go`（通用客户端） |
| **消息处理机制** | 消息队列 + Worker | 无 |
| **Sleepace 报告下载触发** | ✅ MQTT 触发 | ❌ 仅手动触发 |

---

## 🔍 v1.0 实现分析

### 架构

**位置**：`wisefido-backend/wisefido-sleepace/`

**文件结构**：
```
wisefido-sleepace/
├── main.go                    # 启动 MQTT 客户端
├── modules/
│   └── borker.go              # MQTT 消息处理模块（独立）
└── internal/
    └── config/
        └── config.go          # MQTT 配置
```

### 实现细节

#### 1. MQTT 客户端初始化 (`main.go`)

```go
func initMqtt() (mqtt.Client, error) {
    opts := mqtt.NewClientOptions().
        SetClientID(config.Cfg.Mqtt.ClientId).
        SetUsername(config.Cfg.Mqtt.Username).
        SetPassword(config.Cfg.Mqtt.Password).
        SetBroker(config.Cfg.Mqtt.Address).
        SetAutoReconnect(true).
        SetKeepAlive(60 * time.Second).
        SetPingTimeout(10 * time.Second)
    
    client := mqtt.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        return nil, token.Error()
    }
    
    // 订阅 MQTT topic
    topic := config.Cfg.Mqtt.TopicId
    client.Subscribe(topic, 0, modules.MqttBroker)
    
    return client, nil
}
```

#### 2. MQTT 消息处理模块 (`modules/borker.go`)

**核心组件**：
- **消息队列**：`chan *Message`（缓冲 1000）
- **Worker 池**：10 个并发 worker
- **消息处理函数**：`handleMessage`

**架构**：
```
MQTT 消息
    ↓
MqttBroker (消息接收)
    ↓
messageQueue (消息队列)
    ↓
worker (并发处理)
    ↓
handleMessage (消息路由)
    ↓
handleAnalysisEvent (事件处理)
    ↓
DownloadReport (业务逻辑)
```

**关键代码**：
```go
// 初始化消息队列和 worker
func InitBroker() {
    messageQueue = make(chan *Message, 1000)
    var ctx context.Context
    ctx, cancel = context.WithCancel(context.Background())
    for i := 0; i < wokerCount; i++ {
        wg.Add(1)
        go worker(ctx, &wg, messageQueue)
    }
}

// MQTT 消息接收
func MqttBroker(client mqtt.Client, msg mqtt.Message) {
    messageQueue <- &Message{
        Topic:   msg.Topic(),
        Payload: msg.Payload(),
    }
}

// Worker 处理消息
func worker(ctx context.Context, wg *sync.WaitGroup, queue <-chan *Message) {
    defer wg.Done()
    for {
        select {
        case <-ctx.Done():
            return
        case msg := <-queue:
            handleMessage(msg)
        }
    }
}

// 消息路由和处理
func handleMessage(msg *Message) {
    // 解析消息
    var messages []*models.ReceivedMessage
    json.Unmarshal(msg.Payload, &messages)
    
    // 根据消息类型路由
    for _, m := range messages {
        switch m.Type {
        case "analysis":
            handleAnalysisEvent(m.Data)
        case "upgradeProgress":
            handleUpgradeProgress(m.Data)
        // ... 其他消息类型
        }
    }
}

// 处理分析事件（触发报告下载）
func handleAnalysisEvent(data *models.AnalysisData) error {
    // 保存分析数据
    record := &models.SleepaceAnalysis{...}
    database.Engine.Insert(record)
    
    // 触发报告下载
    return DownloadReport(
        utils.LongId(utils.Atoi64(data.UserId, 0)),
        data.DeviceId,
        data.StartTime+1,
        data.TimeStamp,
    )
}
```

#### 3. 配置 (`internal/config/config.go`)

```go
type MqttConfig struct {
    Address  string `yaml:"address"`
    ClientId string `yaml:"client_id"`
    Username string `yaml:"username"`
    Password string `yaml:"password"`
    TopicId  string `yaml:"topic_id"`  // 如 "sleepace-57136"
}
```

---

## 🔍 v1.5 实现分析

### 当前状态

**位置**：`owlBack/wisefido-data/`

**文件结构**：
```
wisefido-data/
├── cmd/wisefido-data/
│   └── main.go                # 主程序（无 MQTT 初始化）
├── internal/
│   ├── http/                  # HTTP Handler
│   ├── service/               # Service 层
│   └── repository/             # Repository 层
└── (无 MQTT 处理模块)
```

### 相关组件

#### 1. 通用 MQTT 客户端 (`owl-common/mqtt/client.go`)

**用途**：通用的 MQTT 客户端封装，供多个服务使用

**特点**：
- 不是专门为 `wisefido-data` 设计
- 可能被其他服务（如 `wisefido-sleepace`、`wisefido-radar`）使用
- 需要检查是否被 `wisefido-data` 使用

#### 2. 其他服务的 MQTT 实现

**`wisefido-sleepace`** (`owlBack/wisefido-sleepace/`):
- 有独立的 MQTT 消费者：`internal/consumer/mqtt_consumer.go`
- 用于处理 Sleepace 设备的 MQTT 消息

**`wisefido-radar`** (`owlBack/wisefido-radar/`):
- 有独立的 MQTT 消费者：`internal/consumer/mqtt_consumer.go`
- 用于处理 Radar 设备的 MQTT 消息

**注意**：这些是独立的微服务，不是 `wisefido-data` 的一部分。

---

## 📋 对比总结

### v1.0

✅ **有独立的 MQTT 处理模块**
- 位置：`modules/borker.go`
- 功能：
  - MQTT 消息接收
  - 消息队列缓冲
  - Worker 池并发处理
  - 消息路由和处理
  - 触发 Sleepace 报告下载

✅ **集成在同一个服务中**
- `wisefido-sleepace` 服务同时处理：
  - HTTP API（报告查询）
  - MQTT 消息（报告下载触发）

### v1.5

❌ **没有独立的 MQTT 处理模块**
- `wisefido-data` 服务目前只处理 HTTP API
- 没有 MQTT 消息监听和处理逻辑

✅ **有通用的 MQTT 客户端**
- `owl-common/mqtt/client.go` 提供通用封装
- 但未被 `wisefido-data` 使用

✅ **其他服务有 MQTT 实现**
- `wisefido-sleepace` 和 `wisefido-radar` 有独立的 MQTT 消费者
- 但这些是独立的微服务，不是 `wisefido-data` 的一部分

---

## 🎯 结论

### v1.0
- ✅ **有独立的 MQTT 处理模块**（`modules/borker.go`）
- ✅ 在同一服务中集成 HTTP 和 MQTT 处理

### v1.5
- ❌ **没有独立的 MQTT 处理模块**
- ❌ `wisefido-data` 服务目前只支持手动触发下载（HTTP API）
- ✅ 有通用的 MQTT 客户端库（`owl-common/mqtt/client.go`），但未被使用
- ✅ 其他微服务（`wisefido-sleepace`、`wisefido-radar`）有独立的 MQTT 实现

---

## 💡 建议

如果要实现 v1.5 的 MQTT 触发下载，需要：

1. **创建独立的 MQTT 处理模块**
   - 位置：`internal/mqtt/sleepace_broker.go`
   - 参考 v1.0 的 `modules/borker.go`

2. **使用通用的 MQTT 客户端**
   - 使用 `owl-common/mqtt/client.go`
   - 或直接使用 `github.com/eclipse/paho.mqtt.golang`

3. **集成到主程序**
   - 在 `main.go` 中初始化 MQTT 客户端
   - 订阅 Sleepace 相关的 MQTT topic
   - 调用 Service 层的 `DownloadReport` 方法

