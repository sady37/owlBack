# MQTT 触发下载实现 TODO

## 📋 概述

本文档记录 MQTT 触发下载功能的实现 TODO，用于后续开发。

**当前状态**：框架已创建，具体逻辑待实现（默认禁用）

**启用方式**：设置环境变量 `MQTT_ENABLED=true`

---

## ✅ 已完成

1. **框架代码**
   - ✅ 创建 `internal/mqtt/sleepace_broker.go` 框架
   - ✅ 定义 `SleepaceMQTTBroker` 结构体
   - ✅ 定义接口方法（`HandleMessage`, `Start`, `Stop`）

2. **配置**
   - ✅ 添加 `MQTTConfig` 到 `config.go`
   - ✅ 支持环境变量配置（默认禁用）

3. **主程序集成**
   - ✅ 在 `main.go` 中添加条件初始化（TODO 注释）

---

## ⏳ 待实现

### 1. 实现 MQTT 消息解析

**文件**：`internal/mqtt/sleepace_broker.go`

**方法**：`HandleMessage`

**实现步骤**：
1. 解析 MQTT 消息（Sleepace 消息格式：数组）
2. 遍历消息数组，处理每条消息
3. 调用 `processMessage` 处理单条消息

**参考**：
- `wisefido-backend/wisefido-sleepace/modules/borker.go::handleMessage`
- `wisefido-backend/wisefido-sleepace/models/receive.go`

**消息格式**：
```json
[
  {
    "deviceId": "device_code",
    "dataKey": "analysis",
    "timestamp": 1234567890,
    "data": {
      "deviceId": "device_code",
      "userId": "user_id",
      "startTime": 1234567890,
      "timeStamp": 1234567890
    }
  }
]
```

**代码模板**：
```go
func (b *SleepaceMQTTBroker) HandleMessage(topic string, payload []byte) error {
    var messages []ReceivedMessage
    if err := json.Unmarshal(payload, &messages); err != nil {
        return fmt.Errorf("failed to unmarshal message: %w", err)
    }
    
    for _, msg := range messages {
        if err := b.processMessage(&msg); err != nil {
            b.logger.Error("Failed to process message", zap.Error(err))
            // 继续处理下一条消息，不中断
        }
    }
    
    return nil
}
```

---

### 2. 实现消息路由

**文件**：`internal/mqtt/sleepace_broker.go`

**方法**：`processMessage`

**实现步骤**：
1. 根据 `dataKey` 路由到不同的处理函数
2. 支持的消息类型：
   - `"analysis"` -> `handleAnalysisEvent`（触发报告下载）
   - `"upgradeProgress"` -> `handleUpgradeProgress`
   - `"connectionStatus"` -> `handleConnectionStatus`
   - `"alarmNotify"` -> `handleAlarmNotify`

**参考**：
- `wisefido-backend/wisefido-sleepace/modules/borker.go::handleMessage`

**代码模板**：
```go
func (b *SleepaceMQTTBroker) processMessage(msg *ReceivedMessage) error {
    switch msg.DataKey {
    case "analysis":
        return b.handleAnalysisEvent(msg.Data)
    case "upgradeProgress":
        return b.handleUpgradeProgress(msg.Data)
    case "connectionStatus":
        return b.handleConnectionStatus(msg.Data)
    case "alarmNotify":
        return b.handleAlarmNotify(msg.Data)
    default:
        b.logger.Debug("Unhandled data key", zap.String("data_key", msg.DataKey))
        return nil
    }
}
```

---

### 3. 实现分析事件处理（触发报告下载）

**文件**：`internal/mqtt/sleepace_broker.go`

**方法**：`handleAnalysisEvent`

**实现步骤**：
1. 解析 `AnalysisData`
   - `DeviceId`: 设备编码（device_code）
   - `UserId`: 用户 ID（对应 device_id）
   - `StartTime`: 开始时间
   - `TimeStamp`: 结束时间
2. 通过 `device_code` 查询 `device_id`（如果 `UserId` 为空）
3. 获取 `tenant_id`（从设备信息中获取）
4. 构建 `DownloadReportRequest`
5. 调用 Service 层的 `DownloadReport` 方法

**参考**：
- `wisefido-backend/wisefido-sleepace/modules/borker.go::handleAnalysisEvent`
- `wisefido-backend/wisefido-sleepace/modules/sleepace_service.go::DownloadReport`

**AnalysisData 格式**：
```json
{
  "deviceId": "device_code",
  "userId": "user_id",
  "startTime": 1234567890,
  "timeStamp": 1234567890
}
```

**代码模板**：
```go
func (b *SleepaceMQTTBroker) handleAnalysisEvent(data json.RawMessage) error {
    var analysisData struct {
        DeviceId  string `json:"deviceId"`
        UserId    string `json:"userId"`
        StartTime int64  `json:"startTime"`
        TimeStamp int64  `json:"timeStamp"`
    }
    if err := json.Unmarshal(data, &analysisData); err != nil {
        return fmt.Errorf("failed to unmarshal analysis data: %w", err)
    }
    
    // 获取 tenant_id（需要从设备信息中获取）
    // TODO: 通过 device_code 查询设备信息，获取 tenant_id
    tenantID := "" // 从 devices 表查询
    
    // 获取 device_id（如果 UserId 为空，通过 device_code 查询）
    deviceID := analysisData.UserId
    if deviceID == "" {
        // TODO: 通过 device_code 查询 device_id
        // 使用 SleepaceReportsRepository.GetDeviceIDByDeviceCode
    }
    
    // 构建 DownloadReportRequest
    req := service.DownloadReportRequest{
        TenantID:   tenantID,
        DeviceID:   deviceID,
        DeviceCode: analysisData.DeviceId,
        StartTime:  analysisData.StartTime + 1, // v1.0 中加 1
        EndTime:    analysisData.TimeStamp,
    }
    
    // 调用 Service 层
    ctx := context.Background()
    if err := b.sleepaceReportService.DownloadReport(ctx, req); err != nil {
        return fmt.Errorf("failed to download report: %w", err)
    }
    
    b.logger.Info("Successfully triggered report download via MQTT",
        zap.String("device_id", deviceID),
        zap.String("device_code", analysisData.DeviceId),
        zap.Int64("start_time", req.StartTime),
        zap.Int64("end_time", req.EndTime),
    )
    
    return nil
}
```

---

### 4. 实现 MQTT 订阅

**文件**：`internal/mqtt/sleepace_broker.go`

**方法**：`Start`

**实现步骤**：
1. 使用 `owl-common/mqtt/client.go` 创建 MQTT 客户端（已在 main.go 中创建）
2. 订阅 Sleepace MQTT 主题（从配置读取，如 `"sleepace-57136"`）
3. 注册消息处理函数（`HandleMessage`）

**参考**：
- `owl-common/mqtt/client.go::Subscribe`
- `wisefido-backend/wisefido-sleepace/main.go::initMqtt`

**代码模板**：
```go
func (b *SleepaceMQTTBroker) Start(ctx context.Context, mqttClient *mqttcommon.Client) error {
    topic := "" // TODO: 从配置读取 MQTT 主题
    if err := mqttClient.Subscribe(topic, 1, b.HandleMessage); err != nil {
        return fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
    }
    
    b.logger.Info("MQTT broker started",
        zap.String("topic", topic),
    )
    
    return nil
}
```

---

### 5. 实现 MQTT 取消订阅

**文件**：`internal/mqtt/sleepace_broker.go`

**方法**：`Stop`

**实现步骤**：
1. 取消订阅 MQTT 主题
2. 记录日志

**参考**：
- `owl-common/mqtt/client.go::Unsubscribe`

**代码模板**：
```go
func (b *SleepaceMQTTBroker) Stop(ctx context.Context, mqttClient *mqttcommon.Client) error {
    topic := "" // TODO: 从配置读取 MQTT 主题
    if err := mqttClient.Unsubscribe(topic); err != nil {
        b.logger.Error("Failed to unsubscribe", zap.Error(err))
        return err
    }
    
    b.logger.Info("MQTT broker stopped")
    return nil
}
```

---

### 6. 在主程序中启用 MQTT

**文件**：`cmd/wisefido-data/main.go`

**实现步骤**：
1. 检查 `cfg.MQTT.Enabled` 是否为 `true`
2. 如果启用，创建 MQTT 客户端
3. 创建 `SleepaceMQTTBroker` 实例
4. 启动 MQTT Broker
5. 在服务停止时停止 MQTT Broker

**参考**：
- `wisefido-backend/wisefido-sleepace/main.go::initMqtt`
- `owl-common/mqtt/client.go`

**代码模板**：
```go
if cfg.MQTT.Enabled {
    // 使用 owl-common/mqtt/client.go 创建 MQTT 客户端
    mqttConfig := &commoncfg.MQTTConfig{
        Broker:   cfg.MQTT.Broker,
        ClientID: cfg.MQTT.ClientID,
        Username: cfg.MQTT.Username,
        Password: cfg.MQTT.Password,
    }
    mqttClient, err := mqttcommon.NewClient(mqttConfig, logger)
    if err != nil {
        logger.Error("Failed to create MQTT client", zap.Error(err))
    } else {
        // 创建 MQTT Broker
        mqttBroker := mqtt.NewSleepaceMQTTBroker(sleepaceReportService, logger)
        // 启动 MQTT Broker
        if err := mqttBroker.Start(ctx, mqttClient); err != nil {
            logger.Error("Failed to start MQTT broker", zap.Error(err))
        } else {
            logger.Info("MQTT broker started",
                zap.String("broker", cfg.MQTT.Broker),
                zap.String("topic", cfg.MQTT.Topic),
            )
            // 在服务停止时停止 MQTT Broker
            defer mqttBroker.Stop(ctx, mqttClient)
        }
    }
} else {
    logger.Info("MQTT trigger download is disabled (set MQTT_ENABLED=true to enable)")
}
```

---

### 7. 定义消息模型

**文件**：`internal/mqtt/models.go`（新建）

**实现步骤**：
1. 定义 `ReceivedMessage` 结构体
2. 定义 `AnalysisData` 结构体
3. 定义其他消息类型结构体（可选）

**参考**：
- `wisefido-backend/wisefido-sleepace/models/receive.go`
- `owlBack/wisefido-sleepace/internal/models/message.go`

**代码模板**：
```go
package mqtt

import "encoding/json"

// ReceivedMessage Sleepace MQTT 消息结构（v1.0 格式）
type ReceivedMessage struct {
	DeviceId  string          `json:"deviceId"`  // 设备代码（device_code）
	DataKey   string          `json:"dataKey"`  // 数据类型：analysis, upgradeProgress, connectionStatus, alarmNotify
	TimeStamp int64           `json:"timestamp"` // 时间戳
	Data      json.RawMessage `json:"data"`      // 数据内容（JSON）
}

// AnalysisData 分析数据
type AnalysisData struct {
	DeviceId  string `json:"deviceId"`  // 设备代码（device_code）
	UserId    string `json:"userId"`    // 用户 ID（对应 device_id）
	StartTime int64  `json:"startTime"` // 开始时间
	TimeStamp int64  `json:"timeStamp"` // 结束时间
}
```

---

## 📝 实现顺序

1. **定义消息模型**（步骤 7）
   - 创建 `internal/mqtt/models.go`
   - 定义 `ReceivedMessage` 和 `AnalysisData`

2. **实现消息解析**（步骤 1）
   - 实现 `HandleMessage` 方法

3. **实现消息路由**（步骤 2）
   - 实现 `processMessage` 方法

4. **实现分析事件处理**（步骤 3）
   - 实现 `handleAnalysisEvent` 方法
   - 这是核心功能，触发报告下载

5. **实现 MQTT 订阅**（步骤 4、5）
   - 实现 `Start` 和 `Stop` 方法

6. **在主程序中启用**（步骤 6）
   - 在 `main.go` 中启用 MQTT 功能

---

## 🧪 测试计划

### 1. 单元测试

- 测试消息解析
- 测试消息路由
- 测试分析事件处理

### 2. 集成测试

- 测试 MQTT 连接
- 测试消息订阅
- 测试报告下载触发

### 3. 端到端测试

- 模拟 MQTT 消息
- 验证报告下载
- 验证数据保存

---

## 📚 参考文档

- [v1.0 实现分析](./SLEEPACE_REPORT_V1.0_IMPLEMENTATION_ANALYSIS.md)
- [v1.0 数据同步分析](./SLEEPACE_REPORT_V1.0_DATA_SYNC_ANALYSIS.md)
- [架构层次设计](./SLEEPACE_REPORT_ARCHITECTURE_LAYERS.md)
- [MQTT 客户端设计](../docs/02_MQTT_Client_Design.md)

---

## 🔗 相关代码

- `wisefido-backend/wisefido-sleepace/modules/borker.go` - v1.0 MQTT 处理实现
- `wisefido-backend/wisefido-sleepace/modules/sleepace_service.go::DownloadReport` - v1.0 报告下载实现
- `owl-common/mqtt/client.go` - v1.5 MQTT 客户端封装
- `owlBack/wisefido-sleepace/internal/consumer/mqtt_consumer.go` - v1.5 Sleepace MQTT 消费者示例

---

## ✅ 完成标准

- [ ] 所有 TODO 注释已实现
- [ ] 单元测试通过
- [ ] 集成测试通过
- [ ] 端到端测试通过
- [ ] 文档已更新
- [ ] 代码审查通过

