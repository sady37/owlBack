# Sleepace Report v1.0 实现分析

## 📋 前端调用（wisefido-frontend / owlFront）

### API 端点
```typescript
// owlFront/src/api/report/report.ts
export enum Api {
  SleepaceReports = '/sleepace/api/v1/sleepace/reports/:id',
  SleepaceReportDetail = '/sleepace/api/v1/sleepace/reports/:id/detail',
  SleepaceReportsDates = '/sleepace/api/v1/sleepace/reports/:id/dates',
}
```

### 前端页面
- **列表页**：`views/report/daily-report-sleepace.vue`
  - 调用 `getSleepaceReportsApi(deviceId, params)`
  - 参数：`startDate`, `endDate`, `page`, `size`
  
- **详情页**：`views/report/report-detail.vue`
  - 调用 `getSleepaceReportDetailApi(deviceId, date)`
  - 参数：`date` (日期数字，如 20240820)

---

## 🔍 wisefido-backend (v1.0) 实现

### 1. 路由定义

**文件**：`wisefido-sleepace/routes/v1.go`

```go
v1 := r.Group("/wisefido/sleepace/api/v1/sleepace")
v1.Use(auth.AfterAuthMiddleware())
v1.GET("/reports/:id", controllers.GetSleepReports)
v1.GET("/reports/:id/detail", controllers.GetSleepReportDetail)
v1.GET("/reports/:id/dates", controllers.GetSleepReportDates)
```

**注意**：v1.0 的路由前缀是 `/wisefido/sleepace/api/v1/sleepace`，而 v1.5 使用的是 `/sleepace/api/v1/sleepace`

---

### 2. Controller 层

**文件**：`wisefido-sleepace/controllers/sleepace_controller.go`

#### 2.1 GetSleepReports - 获取报告列表

```go
func GetSleepReports(ctx *gin.Context) {
    id := ctx.Param("id")
    idInt, err := strconv.ParseInt(id, 10, 64)
    // 参数验证
    startDate := utils.Atoi(ctx.Query("startDate"), 0)
    endDate := utils.Atoi(ctx.Query("endDate"), 0)
    page := utils.GeneratePaginationFromRequest(ctx, config.Cfg.Database.DefaultSize)
    deviceId := utils.LongId(idInt)
    
    // 调用 Module 层
    response, err := modules.GetDailyReports(deviceId, startDate, endDate, page)
    if err == nil {
        utils.ResponseSuccessJsonWithPagination(ctx, response, page)
    } else {
        utils.ResponseErrorJson(ctx, err)
    }
}
```

**功能**：
- 解析 device_id（路径参数）
- 解析 startDate, endDate（查询参数）
- 生成分页信息
- 调用 `modules.GetDailyReports`
- 返回分页响应

#### 2.2 GetSleepReportDetail - 获取报告详情

```go
func GetSleepReportDetail(ctx *gin.Context) {
    id := ctx.Param("id")
    idInt, err := strconv.ParseInt(id, 10, 64)
    date := utils.Atoi(ctx.Query("date"), 0)
    deviceId := utils.LongId(idInt)
    
    // 调用 Module 层
    response, err := modules.GetDailyReportDetail(deviceId, date)
    if err != nil {
        utils.ResponseErrorJson(ctx, err)
    } else {
        utils.ResponseSuccessJson(ctx, response)
    }
}
```

**功能**：
- 解析 device_id（路径参数）
- 解析 date（查询参数）
- 调用 `modules.GetDailyReportDetail`
- 返回详情响应

#### 2.3 GetSleepReportDates - 获取有数据的日期列表

```go
func GetSleepReportDates(ctx *gin.Context) {
    id := ctx.Param("id")
    idInt, err := strconv.ParseInt(id, 10, 64)
    deviceId := utils.LongId(idInt)
    
    // 调用 Module 层
    response, err := modules.GetDailyReportsValidDates(deviceId)
    if err != nil {
        utils.ResponseErrorJson(ctx, err)
    } else {
        utils.ResponseSuccessJson(ctx, response)
    }
}
```

**功能**：
- 解析 device_id（路径参数）
- 调用 `modules.GetDailyReportsValidDates`
- 返回日期数组

---

### 3. Module 层（类似 Service 层）

**文件**：`wisefido-sleepace/modules/sleepace_service.go`

#### 3.1 GetDailyReports

```go
func GetDailyReports(deviceId utils.LongId, startDate, endDate int, page *utils.Pagination) ([]*models.SleepaceReportOutline, error) {
    if utils.IsDateValid(startDate) && utils.IsDateValid(endDate) {
        return models.GetReports(deviceId, startDate, endDate, page)
    } else {
        return nil, errors.New("invalid date")
    }
}
```

**功能**：
- 日期验证
- 调用 Model 层查询数据库

#### 3.2 GetDailyReportDetail

```go
func GetDailyReportDetail(deviceId utils.LongId, date int) (*models.SleepaceReportDetail, error) {
    if !utils.IsDateValid(date) {
        return nil, errors.New("invalid date")
    }
    return models.GetReport(deviceId, date)
}
```

**功能**：
- 日期验证
- 调用 Model 层查询数据库

#### 3.3 GetDailyReportsValidDates

```go
func GetDailyReportsValidDates(deviceId utils.LongId) ([]int, error) {
    if deviceId == 0 {
        return []int{}, nil
    }
    return models.GetReportsValidDates(deviceId)
}
```

**功能**：
- 参数验证
- 调用 Model 层查询数据库

---

### 4. Model 层（类似 Repository 层）

**文件**：`wisefido-sleepace/models/report.go`

#### 4.1 数据模型

```go
type SleepaceReport struct {
    Id          uint
    DeviceId    utils.LongId
    DeviceCode  string
    RecordCount int
    StartTime   int64
    EndTime     int64
    Date        int          // 日期数字，如 20240820
    StopMode    int
    TimeStep    int
    Timezone    int
    SleepState  string       // JSON 字符串，如 "[1,2,1,1,1,...]"
    Report      string       // JSON 字符串，包含完整的报告数据
    CreatedAt   int64
    UpdatedAt   int64
}

type SleepaceReportOutline struct {
    // 不包含 Report 字段（用于列表）
}

type SleepaceReportDetail struct {
    // 包含 Report 字段（用于详情）
}
```

#### 4.2 GetReports - 查询报告列表

```go
func GetReports(deviceId utils.LongId, startDate, endDate int, page *utils.Pagination) ([]*SleepaceReportOutline, error) {
    sql := "select * from sleepace_report where device_id = ? and date >= ? and date <= ?"
    args := []any{deviceId, startDate, endDate}
    
    // 分页处理
    if page != nil {
        // 计算总数
        countSql := "select count(1) as count from sleepace_report where device_id = ? and date >= ? and date <= ?"
        // 排序（默认按 date desc）
        // 分页（limit offset, size）
    }
    
    reports := make([]*SleepaceReportOutline, 0)
    err := database.Engine.SQL(sql, args...).Find(&reports)
    return reports, err
}
```

**功能**：
- 直接 SQL 查询 `sleepace_report` 表
- 按 device_id, date 范围过滤
- 支持分页和排序

#### 4.3 GetReport - 查询报告详情

```go
func GetReport(deviceId utils.LongId, date int) (*SleepaceReportDetail, error) {
    report := SleepaceReportDetail{DeviceId: deviceId, Date: date}
    exist, err := database.Engine.Table("sleepace_report").Get(&report)
    if !exist {
        return nil, nil
    }
    // 处理 Report 字段格式（确保是 JSON 数组）
    if report.Report[0] != '[' {
        report.Report = "[" + report.Report + "]"
    }
    return &report, nil
}
```

**功能**：
- 直接查询 `sleepace_report` 表
- 按 device_id, date 查询单条记录
- 格式化 Report 字段（确保是 JSON 数组）

#### 4.4 GetReportsValidDates - 查询有数据的日期列表

```go
func GetReportsValidDates(deviceId utils.LongId) ([]int, error) {
    dates := make([]int, 0)
    err := database.Engine.SQL("select date from sleepace_report where device_id = ?", deviceId).Find(&dates)
    return dates, err
}
```

**功能**：
- 直接 SQL 查询 `sleepace_report` 表
- 返回该设备的所有日期列表

---

## 📊 架构总结

### v1.0 架构

```
Controller (controllers/sleepace_controller.go)
    ↓
Module (modules/sleepace_service.go)  ← 类似 Service 层
    ↓
Model (models/report.go)  ← 类似 Repository 层
    ↓
MySQL (sleepace_report 表)  ← **wisefido-backend 自己的数据库**
```

### 关键发现

1. ✅ **有 Module 层**（类似 Service 层）
   - `modules.GetDailyReports`
   - `modules.GetDailyReportDetail`
   - `modules.GetDailyReportsValidDates`
   - 功能：参数验证、调用 Model 层

2. ✅ **有 Model 层**（类似 Repository 层）
   - `models.GetReports`
   - `models.GetReport`
   - `models.GetReportsValidDates`
   - 功能：直接 SQL 查询 MySQL 数据库

3. ✅ **数据来源**：**wisefido-backend 自己的 MySQL 数据库**
   - **数据库**：`wisefido`（wisefido-backend 自己的数据库，不是厂家的数据库）
   - **表**：`sleepace_report` 表
   - **数据获取方式**：
     - 通过调用 Sleepace 厂家 HTTP API (`/sleepace/get24HourDailyWithMaxReport`) 获取报告
     - 然后保存到自己的数据库 (`models.SaveReport`)
     - 查询时直接从自己的数据库查询，不直接查询厂家数据库
   - **表结构**：包含 `device_id`, `date`, `sleep_state`, `report` 等字段
   - **数据格式**：`sleep_state` 是 JSON 字符串数组，`report` 是 JSON 字符串

4. ✅ **数据同步机制**
   - **主动下载**：`modules.DownloadReport` 调用厂家 API 获取报告并保存
   - **MQTT 触发**：`modules/borker.go` 中通过 MQTT 消息触发报告下载
   - **历史报告**：`GetHistorySleepReports` API 可以手动触发下载历史报告

5. ❌ **没有复杂的业务逻辑**
   - 查询时只是简单的数据库查询
   - 没有数据聚合
   - 没有权限检查（只有认证中间件）

---

## 🎯 v1.5 实现建议

### 方案对比

#### 方案 A：直接调用（类似 v1.0）

**架构**：
```
Handler → Repository → PostgreSQL (sleepace_report 表)
```

**适用场景**：
- 如果 v1.5 也使用 `sleepace_report` 表（MySQL → PostgreSQL 迁移）
- 如果数据格式保持不变
- 如果不需要复杂的数据聚合

**优点**：
- ✅ 实现简单
- ✅ 与 v1.0 保持一致
- ✅ 无需 Service 层

**缺点**：
- ❌ 如果数据来源改变（从时间序列数据聚合），需要 Service 层
- ❌ 如果数据格式改变，需要 Service 层进行转换

#### 方案 B：使用 Service 层（推荐）✅

**架构**：
```
Handler → Service → Repository → PostgreSQL
```

**适用场景**：
- 如果数据来源是 `iot_timeseries` 表（需要聚合）
- 如果数据格式需要转换（v1.0 格式 → v1.5 格式）
- 如果需要权限检查
- 如果需要调用外部服务

**优点**：
- ✅ 灵活（可以支持多种数据来源）
- ✅ 易于测试
- ✅ 易于扩展

**当前实现状态**：
- ✅ **查询功能已完成**：Handler → Service → Repository → PostgreSQL
- ❌ **数据同步功能待实现**：
  - ❌ MQTT 触发下载（v1.0 有，v1.5 待实现）
  - ❌ 手动触发下载 API（v1.0 有，v1.5 待实现）
  - ❌ 定时任务（v1.0 没有，v1.5 可选）

**详细分析**：见 `SLEEPACE_REPORT_V1.0_DATA_SYNC_ANALYSIS.md`

---

## 📋 数据表结构（v1.0）

### sleepace_report 表（MySQL）

**数据库**：`wisefido`（wisefido-backend 自己的 MySQL 数据库，不是厂家的数据库）

**配置**：`sleepace-dev.yaml`
```yaml
database:
  host: "localhost"
  port: 3306
  username: "root"
  password: "env(MYSQL_PASSWORD)"
  database: "wisefido"  # ← wisefido-backend 自己的数据库
```

**表结构**：
```sql
CREATE TABLE sleepace_report (
    id          INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    device_id   BIGINT NOT NULL,
    device_code VARCHAR(64) NOT NULL,
    record_count INT NOT NULL,
    start_time  BIGINT NOT NULL,
    end_time    BIGINT NOT NULL,
    date        INT NOT NULL,           -- 日期数字，如 20240820
    stop_mode   INT NOT NULL,
    time_step   INT NOT NULL,
    timezone    INT NOT NULL,
    sleep_state TEXT,                   -- JSON 字符串数组，如 "[1,2,1,1,1,...]"
    report      LONGTEXT,               -- JSON 字符串，包含完整的报告数据
    created_at  BIGINT NOT NULL,
    updated_at  BIGINT NOT NULL,
    INDEX idx_device_id (device_id),
    INDEX idx_date (date)
);
```

**数据获取流程**：
```
Sleepace 厂家服务 (HTTP API)
    ↓
wisefido-sleepace 服务
    ├─ 调用厂家 API: /sleepace/get24HourDailyWithMaxReport
    ├─ 解析报告数据
    └─→ 保存到自己的 MySQL (sleepace_report 表)
        ↓
查询时直接从自己的数据库查询
```

**触发下载的时机**：
1. **MQTT 消息触发**：`modules/borker.go` 中通过 MQTT 消息触发 `DownloadReport`
2. **手动触发**：`GetHistorySleepReports` API 可以手动触发下载历史报告

---

## ✅ 最终建议

### 是否需要 SleepaceReportService？

**结论**：✅ **需要 Service 层**

**理由**：

1. **数据来源可能不同**
   - v1.0：直接查询 `sleepace_report` 表
   - v1.5：可能需要从 `iot_timeseries` 表聚合生成报告

2. **数据格式可能不同**
   - v1.0：`sleep_state` 是 JSON 字符串数组
   - v1.5：可能需要从时间序列数据聚合生成

3. **需要权限检查**
   - v1.0：只有认证中间件
   - v1.5：需要 device_id 验证、tenant_id 过滤

4. **符合 v1.5 架构**
   - v1.5 其他功能都使用 Service 层
   - 保持架构一致性

### 实现方式

**推荐**：使用 Service 层，但实现可以简单

```go
// Service 层
type SleepaceReportService interface {
    GetSleepaceReports(ctx context.Context, req GetSleepaceReportsRequest) (*GetSleepaceReportsResponse, error)
    GetSleepaceReportDetail(ctx context.Context, req GetSleepaceReportDetailRequest) (*GetSleepaceReportDetailResponse, error)
    GetSleepaceReportDates(ctx context.Context, req GetSleepaceReportDatesRequest) (*GetSleepaceReportDatesResponse, error)
}

// 实现
func (s *sleepaceReportService) GetSleepaceReports(ctx context.Context, req GetSleepaceReportsRequest) (*GetSleepaceReportsResponse, error) {
    // 1. 权限检查（device_id 验证、tenant_id 过滤）
    // 2. 调用 Repository 查询数据库
    // 3. 数据转换（如果需要）
    // 4. 返回响应
}
```

**如果数据来源是 `sleepace_report` 表**：
- Service 层可以很简单（只是权限检查 + 调用 Repository）
- Repository 层直接查询数据库

**如果数据来源是 `iot_timeseries` 表**：
- Service 层需要数据聚合逻辑
- Repository 层查询时间序列数据

---

## 📚 参考文件

- `wisefido-backend/wisefido-sleepace/routes/v1.go` - 路由定义
- `wisefido-backend/wisefido-sleepace/controllers/sleepace_controller.go` - Controller 层
- `wisefido-backend/wisefido-sleepace/modules/sleepace_service.go` - Module 层
- `wisefido-backend/wisefido-sleepace/models/report.go` - Model 层
- `owlFront/src/api/report/report.ts` - 前端 API 定义
- `owlFront/src/views/report/daily-report-sleepace.vue` - 前端列表页
- `owlFront/src/views/report/report-detail.vue` - 前端详情页

