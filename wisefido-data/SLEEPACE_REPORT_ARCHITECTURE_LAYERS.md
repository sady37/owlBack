# Sleepace Report 架构层次设计

## 📋 v1.0 架构分析

### 手动触发下载 API

**v1.0 实现位置**：
```
controllers/sleepace_controller.go (Handler 层)
    ↓
modules/sleepace_service.go::GetHistoryDailyReport (Module 层，类似 Service 层)
    ↓
modules/sleepace_service.go::DownloadReport (业务逻辑层)
    ↓
models/report.go::SaveReport (Model 层，类似 Repository 层)
    ↓
MySQL 数据库
```

**层次**：
- **Handler 层**：HTTP 请求处理（`controllers/sleepace_controller.go`）
- **Service 层**：业务逻辑（`modules/sleepace_service.go`）
- **Repository 层**：数据访问（`models/report.go`）

---

### MQTT 触发下载

**v1.0 实现位置**：
```
main.go (启动 MQTT 客户端)
    ↓
modules/borker.go::MqttBroker (MQTT 消息接收)
    ↓
modules/borker.go::worker (消息队列处理)
    ↓
modules/borker.go::handleMessage (消息路由)
    ↓
modules/borker.go::handleAnalysisEvent (事件处理)
    ↓
modules/sleepace_service.go::DownloadReport (业务逻辑层)
    ↓
models/report.go::SaveReport (Model 层)
    ↓
MySQL 数据库
```

**层次**：
- **MQTT 处理层**：独立的 MQTT 消息处理模块（`modules/borker.go`）
- **Service 层**：业务逻辑（`modules/sleepace_service.go`）
- **Repository 层**：数据访问（`models/report.go`）

---

## 🏗️ v1.5 架构设计

### 架构原则

1. **Service 层统一业务逻辑**：`DownloadReport` 方法应该在 Service 层实现
2. **Handler 层处理 HTTP 请求**：手动触发下载 API 在 Handler 层
3. **独立的 MQTT 处理模块**：MQTT 触发下载在独立的模块中，直接调用 Service 层

---

### 方案 1：手动触发下载 API

**架构层次**：
```
HTTP Request
    ↓
SleepaceReportHandler.DownloadReport (Handler 层)
    ↓
SleepaceReportService.DownloadReport (Service 层)
    ├─ Sleepace 厂家 API 客户端
    ├─ 数据解析和转换
    └─ SleepaceReportsRepository.SaveReport (Repository 层)
        ↓
PostgreSQL 数据库
```

**实现位置**：
- **Handler 层**：`internal/http/sleepace_report_handler.go`
- **Service 层**：`internal/service/sleepace_report_service.go`
- **Repository 层**：`internal/repository/postgres_sleepace_reports.go`（已实现）

**代码结构**：
```go
// Handler 层
func (h *SleepaceReportHandler) DownloadReport(w http.ResponseWriter, r *http.Request, deviceID string) {
    // 1. 解析请求参数（startTime, endTime）
    // 2. 获取 tenant_id
    // 3. 调用 Service 层
    resp, err := h.sleepaceReportService.DownloadReport(ctx, req)
    // 4. 返回 HTTP 响应
}

// Service 层
func (s *sleepaceReportService) DownloadReport(ctx context.Context, req DownloadReportRequest) error {
    // 1. 验证设备
    // 2. 调用 Sleepace 厂家 API
    // 3. 解析报告数据
    // 4. 调用 Repository 保存
    return s.reportsRepo.SaveReport(ctx, tenantID, report)
}
```

---

### 方案 2：MQTT 触发下载

**架构层次**：
```
MQTT 消息（设备上报）
    ↓
MQTT 客户端（独立的处理模块）
    ↓
MQTT 消息处理（消息队列 + worker）
    ↓
SleepaceReportService.DownloadReport (Service 层)
    ├─ Sleepace 厂家 API 客户端
    ├─ 数据解析和转换
    └─ SleepaceReportsRepository.SaveReport (Repository 层)
        ↓
PostgreSQL 数据库
```

**实现位置**：
- **MQTT 处理层**：`internal/mqtt/sleepace_broker.go`（新建，独立的 MQTT 处理模块）
- **Service 层**：`internal/service/sleepace_report_service.go`（复用）
- **Repository 层**：`internal/repository/postgres_sleepace_reports.go`（已实现）

**代码结构**：
```go
// MQTT 处理层（internal/mqtt/sleepace_broker.go）
type SleepaceMQTTBroker struct {
    sleepaceReportService service.SleepaceReportService
    logger                *zap.Logger
}

func (b *SleepaceMQTTBroker) HandleMessage(msg mqtt.Message) {
    // 1. 解析 MQTT 消息
    // 2. 提取设备信息和时间范围
    // 3. 调用 Service 层
    err := b.sleepaceReportService.DownloadReport(ctx, req)
}

// Service 层（复用，与手动触发相同）
func (s *sleepaceReportService) DownloadReport(ctx context.Context, req DownloadReportRequest) error {
    // 业务逻辑（与手动触发相同）
}
```

---

## 📊 架构对比

| 功能 | v1.0 实现位置 | v1.5 实现位置 | 层次 |
|------|-------------|--------------|------|
| 手动触发下载 | `controllers/sleepace_controller.go` | `internal/http/sleepace_report_handler.go` | Handler 层 |
| 手动触发业务逻辑 | `modules/sleepace_service.go::GetHistoryDailyReport` | `internal/service/sleepace_report_service.go::DownloadReport` | Service 层 |
| MQTT 触发下载 | `modules/borker.go` | `internal/mqtt/sleepace_broker.go`（新建） | MQTT 处理层 |
| MQTT 业务逻辑 | `modules/sleepace_service.go::DownloadReport` | `internal/service/sleepace_report_service.go::DownloadReport` | Service 层（复用） |
| 数据保存 | `models/report.go::SaveReport` | `internal/repository/postgres_sleepace_reports.go::SaveReport` | Repository 层 |

---

## 🎯 关键设计决策

### 1. Service 层统一业务逻辑

**决策**：`DownloadReport` 方法在 Service 层实现，被 Handler 层和 MQTT 处理层共同调用

**优点**：
- ✅ 代码复用：避免重复实现
- ✅ 易于测试：Service 层可以独立测试
- ✅ 职责清晰：业务逻辑集中在 Service 层

**实现**：
```go
// Service 层
type SleepaceReportService interface {
    // 查询功能（已实现）
    GetSleepaceReports(...)
    GetSleepaceReportDetail(...)
    GetSleepaceReportDates(...)
    
    // 数据下载功能（待实现）
    DownloadReport(ctx context.Context, req DownloadReportRequest) error
}
```

---

### 2. Handler 层处理 HTTP 请求

**决策**：手动触发下载 API 在 Handler 层实现

**优点**：
- ✅ 符合现有架构模式（与其他 Handler 一致）
- ✅ HTTP 请求处理逻辑集中
- ✅ 易于路由注册

**实现**：
```go
// Handler 层
func (h *SleepaceReportHandler) DownloadReport(w http.ResponseWriter, r *http.Request, deviceID string) {
    // HTTP 请求处理
    // 调用 Service 层
}
```

---

### 3. 独立的 MQTT 处理模块

**决策**：MQTT 触发下载在独立的模块中实现（`internal/mqtt/`）

**优点**：
- ✅ 职责分离：MQTT 处理与 HTTP 处理分离
- ✅ 易于维护：MQTT 相关代码集中管理
- ✅ 可扩展：可以支持其他 MQTT 消息类型

**实现**：
```go
// MQTT 处理层（internal/mqtt/sleepace_broker.go）
type SleepaceMQTTBroker struct {
    sleepaceReportService service.SleepaceReportService
    logger                *zap.Logger
}

func (b *SleepaceMQTTBroker) HandleMessage(msg mqtt.Message) {
    // MQTT 消息处理
    // 调用 Service 层
}
```

---

## 📁 文件结构

```
owlBack/wisefido-data/
├── internal/
│   ├── http/
│   │   └── sleepace_report_handler.go          # Handler 层（手动触发下载 API）
│   ├── service/
│   │   └── sleepace_report_service.go          # Service 层（DownloadReport 方法）
│   ├── repository/
│   │   └── postgres_sleepace_reports.go        # Repository 层（已实现）
│   └── mqtt/                                    # 新建：MQTT 处理模块
│       └── sleepace_broker.go                   # MQTT 触发下载
└── cmd/wisefido-data/
    └── main.go                                   # 启动 MQTT 客户端
```

---

## 🔄 数据流

### 手动触发下载流程

```
HTTP Request: POST /sleepace/api/v1/sleepace/reports/:id/download
    ↓
SleepaceReportHandler.DownloadReport (Handler 层)
    ├─ 解析请求参数（startTime, endTime）
    ├─ 获取 tenant_id
    └─→ SleepaceReportService.DownloadReport (Service 层)
        ├─ 验证设备
        ├─ 调用 Sleepace 厂家 API
        ├─ 解析报告数据
        └─→ SleepaceReportsRepository.SaveReport (Repository 层)
            └─→ PostgreSQL (sleepace_report 表)
```

### MQTT 触发下载流程

```
MQTT 消息（设备上报）
    ↓
MQTT 客户端（main.go 启动）
    ↓
SleepaceMQTTBroker.HandleMessage (MQTT 处理层)
    ├─ 解析 MQTT 消息
    ├─ 提取设备信息和时间范围
    └─→ SleepaceReportService.DownloadReport (Service 层)
        ├─ 验证设备
        ├─ 调用 Sleepace 厂家 API
        ├─ 解析报告数据
        └─→ SleepaceReportsRepository.SaveReport (Repository 层)
            └─→ PostgreSQL (sleepace_report 表)
```

---

## ✅ 总结

### 层次划分

1. **Handler 层** (`internal/http/`)
   - 手动触发下载 API
   - HTTP 请求处理

2. **Service 层** (`internal/service/`)
   - `DownloadReport` 方法（业务逻辑）
   - 被 Handler 层和 MQTT 处理层共同调用

3. **MQTT 处理层** (`internal/mqtt/`)（新建）
   - MQTT 消息接收和处理
   - 调用 Service 层

4. **Repository 层** (`internal/repository/`)
   - 数据保存（已实现）

### 关键原则

- ✅ **Service 层统一业务逻辑**：`DownloadReport` 在 Service 层实现
- ✅ **Handler 层处理 HTTP**：手动触发在 Handler 层
- ✅ **独立的 MQTT 模块**：MQTT 触发在独立的模块中
- ✅ **代码复用**：Handler 和 MQTT 都调用同一个 Service 方法

