# Sleepace Report 数据来源澄清

## 🔍 关键发现

### v1.0 实现（wisefido-backend）

**数据库**：**wisefido-backend 自己的 MySQL 数据库**（不是厂家的数据库）

**配置**：`sleepace-dev.yaml`
```yaml
database:
  host: "localhost"
  port: 3306
  username: "root"
  password: "env(MYSQL_PASSWORD)"
  database: "wisefido"  # ← wisefido-backend 自己的数据库
```

**数据获取流程**：
```
Sleepace 厂家服务 (HTTP API: http://47.90.180.176:8080)
    ↓
wisefido-sleepace 服务
    ├─ 调用厂家 API: POST /sleepace/get24HourDailyWithMaxReport
    ├─ 解析报告数据（JSON）
    └─→ 保存到自己的 MySQL (wisefido 数据库，sleepace_report 表)
        ↓
查询时直接从自己的数据库查询
```

**关键代码**：

1. **下载报告**：`modules/sleepace_service.go::DownloadReport`
   ```go
   func DownloadReport(deviceId utils.LongId, deviceCode string, startTime, endTime int64) error {
       // 1. 调用厂家 API
       sleepaceClient.R().SetBody(request).SetResult(&response).Post("/sleepace/get24HourDailyWithMaxReport")
       
       // 2. 解析报告数据
       // ...
       
       // 3. 保存到自己的数据库
       err = models.SaveReport(&r)
   }
   ```

2. **查询报告**：`models/report.go::GetReports`
   ```go
   func GetReports(deviceId utils.LongId, startDate, endDate int, page *utils.Pagination) ([]*SleepaceReportOutline, error) {
       // 直接从自己的 MySQL 数据库查询
       sql := "select * from sleepace_report where device_id = ? and date >= ? and date <= ?"
       err := database.Engine.SQL(sql, args...).Find(&reports)
   }
   ```

3. **触发下载**：
   - **MQTT 触发**：`modules/borker.go` 中通过 MQTT 消息触发 `DownloadReport`
   - **手动触发**：`GetHistorySleepReports` API 可以手动触发下载历史报告

---

## 📊 数据流对比

### v1.0 数据流

```
Sleepace 设备
    ↓
Sleepace 厂家服务（第三方，HTTP API）
    ↓
wisefido-sleepace 服务
    ├─ 调用厂家 API: /sleepace/get24HourDailyWithMaxReport
    ├─ 解析报告数据
    └─→ MySQL (wisefido 数据库，sleepace_report 表)  ← **wisefido-backend 自己的数据库**
        ↓
查询时直接从自己的数据库查询
```

### v1.5 可能的方案

#### 方案 A：从 iot_timeseries 表聚合生成报告

```
Sleepace 设备
    ↓
wisefido-sleepace 服务（后台服务）
    └─→ Redis Streams (sleepace:data:stream)
        ↓
wisefido-data-transformer 服务
    └─→ PostgreSQL TimescaleDB (iot_timeseries)
        ↓
wisefido-data 服务（HTTP API）
    ├─ SleepaceReportService
    ├─ 从 iot_timeseries 表聚合生成报告
    └─→ 返回给前端
```

#### 方案 B：调用 Sleepace 厂家服务（类似 v1.0）

```
Sleepace 设备
    ↓
Sleepace 厂家服务（第三方，HTTP API）
    ↓
wisefido-data 服务（HTTP API）
    ├─ SleepaceReportService
    ├─ 调用厂家 API: /sleepace/get24HourDailyWithMaxReport
    ├─ 数据转换和格式化
    └─→ 返回给前端
```

#### 方案 C：从 wisefido-backend MySQL 迁移到 PostgreSQL

```
wisefido-backend MySQL (sleepace_report 表)
    ↓
数据迁移
    ↓
PostgreSQL (sleepace_report 表)
    ↓
wisefido-data 服务（HTTP API）
    ├─ SleepaceReportService
    ├─ 从 PostgreSQL 查询报告数据
    └─→ 返回给前端
```

---

## ❓ 需要确认的问题

### 1. v1.5 中 sleepace_report 表的状态

**问题**：v1.5 中是否已有 `sleepace_report` 表？

**选项**：
- [ ] 已迁移到 PostgreSQL（需要确认表结构是否一致）
- [ ] 仍在 wisefido-backend 的 MySQL 中（需要跨数据库查询）
- [ ] 不存在（需要从其他数据源生成）

### 2. 数据是否已迁移到 iot_timeseries

**问题**：Sleepace 数据是否已迁移到 `iot_timeseries` 表？

**选项**：
- [ ] 已迁移（可以使用方案 A：从时间序列数据聚合生成报告）
- [ ] 未迁移（需要使用方案 B 或 C）

### 3. 是否继续使用 wisefido-backend 的 MySQL

**问题**：v1.5 是否继续使用 wisefido-backend 的 MySQL 数据库？

**选项**：
- [ ] 继续使用（需要跨数据库查询，不推荐）
- [ ] 迁移到 PostgreSQL（推荐）
- [ ] 不再使用（需要从其他数据源生成）

---

## ✅ 推荐方案

### 如果数据已迁移到 iot_timeseries

**推荐**：方案 A - 从时间序列数据聚合生成报告

**优点**：
- ✅ 数据已标准化（SNOMED CT 编码）
- ✅ 统一的数据源（iot_timeseries）
- ✅ 无需依赖外部服务

### 如果数据未迁移，但需要快速实现

**推荐**：方案 B - 调用 Sleepace 厂家服务

**优点**：
- ✅ 实现简单（类似 v1.0 的 DownloadReport）
- ✅ 无需数据迁移
- ✅ 可以缓存到数据库（类似 v1.0）

### 如果数据已迁移到 PostgreSQL sleepace_report 表

**推荐**：方案 C - 从 PostgreSQL 查询

**优点**：
- ✅ 实现简单（直接查询数据库）
- ✅ 无需数据聚合
- ✅ 性能好（直接查询）

---

## 📋 实现建议

### 第一步：确认数据来源

1. **检查 PostgreSQL 是否有 `sleepace_report` 表**
   ```sql
   SELECT * FROM information_schema.tables 
   WHERE table_schema = 'public' AND table_name = 'sleepace_report';
   ```

2. **检查 `iot_timeseries` 表是否有 Sleepace 数据**
   ```sql
   SELECT DISTINCT device_type FROM iot_timeseries 
   WHERE device_type LIKE '%Sleepace%' OR device_type LIKE '%SleepPad%';
   ```

3. **检查 wisefido-backend MySQL 是否仍在使用**
   - 检查配置文件中是否有 MySQL 连接配置
   - 检查是否有跨数据库查询的需求

### 第二步：根据数据来源选择实现方案

- **如果 iot_timeseries 有数据**：使用方案 A（数据聚合）
- **如果 sleepace_report 表已迁移到 PostgreSQL**：使用方案 C（直接查询）
- **如果都没有**：使用方案 B（调用厂家服务）

---

## 📚 参考代码

- `wisefido-backend/wisefido-sleepace/modules/sleepace_service.go::DownloadReport` - 下载报告
- `wisefido-backend/wisefido-sleepace/models/report.go::GetReports` - 查询报告
- `wisefido-backend/wisefido-sleepace/modules/borker.go` - MQTT 触发下载

