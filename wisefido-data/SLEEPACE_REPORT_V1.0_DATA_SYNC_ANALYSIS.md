# Sleepace Report v1.0 数据同步机制分析

## 📋 v1.0 数据同步方式

### 方式 1：MQTT 消息触发（主要方式）

**实现位置**：`wisefido-backend/wisefido-sleepace/modules/borker.go`

**流程**：
```
MQTT 消息（设备上报）
    ↓
MqttBroker (消息队列)
    ↓
worker (并发处理)
    ↓
handleMessage
    ↓
handleReportUpload
    ↓
DownloadReport (调用厂家 API)
    ↓
SaveReport (保存到 MySQL)
```

**关键代码**：
```go
// modules/borker.go:366
func handleReportUpload(data *models.ReportUploadData) error {
    // ... 保存上传记录 ...
    return DownloadReport(
        utils.LongId(utils.Atoi64(data.UserId, 0)), 
        data.DeviceId, 
        data.StartTime+1, 
        data.TimeStamp
    )
}
```

**MQTT 配置**：
```yaml
mqtt:
  address: "mqtt://47.90.180.176:1883"
  username: "wisefido"
  password: "env(MQTT_PASSWORD)"
  client_id: "wisefido-sleepace-dev"
  topic_id: "sleepace-57136"  # 订阅的 MQTT topic
```

**特点**：
- ✅ 实时触发：设备上报后立即下载报告
- ✅ 自动同步：无需手动干预
- ✅ 事件驱动：基于 MQTT 消息

---

### 方式 2：手动触发下载（API 端点）

**实现位置**：`wisefido-backend/wisefido-sleepace/controllers/sleepace_controller.go`

**路由**：`GET /reports/:id?startTime={startTime}&endTime={endTime}`

**功能**：手动触发下载历史报告

**关键代码**：
```go
// controllers/sleepace_controller.go:106
func GetHistorySleepReports(ctx *gin.Context) {
    deviceId := utils.LongId(idInt)
    startTime := utils.Atoi64(ctx.Query("startTime"), 0)
    endTime := utils.Atoi64(ctx.Query("endTime"), 0)
    err = modules.GetHistoryDailyReport(deviceId, startTime, endTime)
    // ...
}
```

**特点**：
- ✅ 手动控制：可以指定时间范围
- ✅ 历史数据：可以下载历史报告
- ✅ API 调用：通过 HTTP API 触发

---

### 方式 3：定时任务（❌ v1.0 中未实现）

**状态**：v1.0 中没有定时任务（cron）的实现

**说明**：
- 数据同步主要依赖 MQTT 消息触发
- 手动触发用于补充历史数据
- 没有定期轮询厂家服务的逻辑

---

## 🔄 v1.5 需要实现的功能

### ✅ 已完成

1. **查询功能**
   - ✅ 获取报告列表
   - ✅ 获取报告详情
   - ✅ 获取有效日期列表

### ❌ 待实现（数据同步）

#### 1. MQTT 触发下载（高优先级）

**需求**：监听 MQTT 消息，自动触发报告下载

**实现方式**：
- 监听 MQTT topic（如 `sleepace-{app_id}`）
- 解析消息（ReportUploadData）
- 调用 `SleepaceReportService.DownloadReport`
- 保存到 PostgreSQL

**参考**：`wisefido-backend/wisefido-sleepace/modules/borker.go`

**预计工作量**：2-3 天

---

#### 2. 手动触发下载 API（中优先级）

**需求**：提供 API 端点，手动触发下载历史报告

**实现方式**：
- 在 `SleepaceReportHandler` 中添加 `DownloadReport` 方法
- 路由：`POST /sleepace/api/v1/sleepace/reports/:id/download`
- 参数：`startTime`, `endTime`（Unix 时间戳）
- 调用 `SleepaceReportService.DownloadReport`

**参考**：`wisefido-backend/wisefido-sleepace/controllers/sleepace_controller.go::GetHistorySleepReports`

**预计工作量**：1 天

---

#### 3. 定时任务（可选，低优先级）

**需求**：定期轮询厂家服务，下载缺失的报告

**实现方式**：
- 使用 cron 或类似机制
- 每天凌晨检查是否有缺失的报告
- 自动下载缺失的报告

**预计工作量**：1-2 天

---

## 📊 对比总结

| 功能 | v1.0 | v1.5 当前状态 | v1.5 待实现 |
|------|------|--------------|------------|
| 查询报告列表 | ✅ | ✅ | - |
| 查询报告详情 | ✅ | ✅ | - |
| 查询有效日期 | ✅ | ✅ | - |
| MQTT 触发下载 | ✅ | ❌ | ⏳ 待实现 |
| 手动触发下载 | ✅ | ❌ | ⏳ 待实现 |
| 定时任务 | ❌ | ❌ | ⏳ 可选 |

---

## 🎯 建议的实现顺序

### 阶段 1：手动触发下载 API（优先）

**原因**：
- 实现简单（1 天）
- 可以立即使用
- 不依赖 MQTT 基础设施

**任务**：
1. 实现 `DownloadReport` Service 方法
2. 实现 Sleepace 厂家 API 客户端
3. 实现 `DownloadReport` Handler
4. 添加配置管理

### 阶段 2：MQTT 触发下载（设备接入后）

**原因**：
- 需要 MQTT 基础设施
- 需要设备接入后才能测试
- 实现复杂度较高（2-3 天）

**任务**：
1. 实现 MQTT 客户端
2. 实现消息处理逻辑
3. 集成到 `SleepaceReportService`

### 阶段 3：定时任务（可选）

**原因**：
- 作为补充机制
- 可以定期检查缺失数据
- 实现简单（1-2 天）

---

## 📝 关键发现

1. **v1.0 主要依赖 MQTT 触发**：设备上报后自动下载报告
2. **v1.0 有手动触发 API**：可以手动下载历史报告
3. **v1.0 没有定时任务**：不依赖定期轮询
4. **v1.5 缺少数据同步机制**：只有查询功能，没有数据下载功能

---

## ✅ 结论

**v1.5 当前状态**：
- ✅ 查询功能已完成
- ❌ **缺少数据同步机制**（MQTT 触发 + 手动触发）

**建议**：
1. **立即实现**：手动触发下载 API（可以立即使用）
2. **设备接入后实现**：MQTT 触发下载（需要 MQTT 基础设施）
3. **可选实现**：定时任务（作为补充机制）

