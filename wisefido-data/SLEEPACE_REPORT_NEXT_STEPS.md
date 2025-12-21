# SleepaceReportService 后续规划

## 📊 当前完成状态

### ✅ 已完成（查询功能）

1. **数据库层**
   - ✅ PostgreSQL 表结构 (`sleepace_report`)
   - ✅ Domain 模型
   - ✅ Repository 接口和实现
   - ✅ `device_code` 与 `serial_number`/`uid` 等价关系处理

2. **业务层**
   - ✅ Service 接口和实现（查询功能）
   - ✅ Handler 实现
   - ✅ 路由注册
   - ✅ 设备验证

3. **功能**
   - ✅ 获取报告列表 (`GetSleepaceReports`)
   - ✅ 获取报告详情 (`GetSleepaceReportDetail`)
   - ✅ 获取有效日期列表 (`GetSleepaceReportDates`)

---

## 🎯 后续规划（按优先级）

### 阶段 0：测试数据准备（当前阶段，立即完成）

**目标**：为开发阶段准备测试数据，支持前端开发和测试

**任务清单**：

1. **创建测试数据 SQL 脚本**
   - [ ] 创建 `owlRD/db/test_data_sleepace_report.sql`
   - [ ] 插入示例报告数据（模拟真实报告格式）
   - [ ] 支持不同日期、不同设备的测试数据

2. **测试数据说明文档**
   - [ ] 说明如何加载测试数据
   - [ ] 说明测试数据的结构和用途

**预计工作量**：0.5 天

---

### 阶段 1：数据下载和保存功能（待设备接入后实现）

**状态**：⏸️ **暂停** - 开发阶段无设备，厂家服务无数据，待设备接入后再实现

**目标**：实现从 Sleepace 厂家服务下载报告并保存到数据库

**⚠️ 重要发现**：v1.0 中有两种数据同步方式，v1.5 都需要实现：
1. **MQTT 触发下载**（主要方式）- v1.0 有，v1.5 待实现
2. **手动触发下载 API**（补充方式）- v1.0 有，v1.5 待实现

**详细分析**：见 `SLEEPACE_REPORT_V1.0_DATA_SYNC_ANALYSIS.md`

**任务清单**：

1. **实现数据下载 Service 方法**
   - [ ] 在 `SleepaceReportService` 中添加 `DownloadReport` 方法
   - [ ] 调用 Sleepace 厂家 HTTP API (`/sleepace/get24HourDailyWithMaxReport`)
   - [ ] 解析报告数据（JSON）
   - [ ] 调用 `SleepaceReportsRepository.SaveReport` 保存

2. **实现数据下载 Handler**
   - [ ] 在 `SleepaceReportHandler` 中添加 `DownloadReport` 方法
   - [ ] 路由：`POST /sleepace/api/v1/sleepace/reports/:id/download`
   - [ ] 参数：`startTime`, `endTime`（Unix 时间戳）

3. **配置管理**
   - [ ] 添加 Sleepace 厂家服务配置（HTTP 地址、App ID、Secret Key 等）
   - [ ] 参考：`wisefido-backend/wisefido-sleepace/internal/config/config.go`

4. **HTTP 客户端**
   - [ ] 实现 Sleepace 厂家 API 客户端
   - [ ] 认证逻辑（Token 生成）
   - [ ] 错误处理和重试逻辑
   - [ ] 参考：`wisefido-backend/wisefido-sleepace/modules/sleepace_service.go`

**参考代码**：
- `wisefido-backend/wisefido-sleepace/modules/sleepace_service.go::DownloadReport`
- `wisefido-backend/wisefido-sleepace/modules/sleepace_service.go::initialize`（Token 生成）

**预计工作量**：2-3 天（待设备接入后）

---

### 阶段 2：测试（中优先级）

**目标**：确保功能正确性和稳定性

**任务清单**：

1. **单元测试**
   - [ ] Repository 层测试（数据库操作）
   - [ ] Service 层测试（业务逻辑）
   - [ ] Handler 层测试（HTTP 请求处理）

2. **集成测试**
   - [ ] 端到端测试（从 HTTP 请求到数据库查询）
   - [ ] 设备验证测试
   - [ ] 分页测试
   - [ ] `device_code` 匹配测试

3. **兼容性测试**
   - [ ] 与 v1.0 前端 API 调用兼容性
   - [ ] 响应格式一致性

**预计工作量**：1-2 天

---

### 阶段 3：权限检查增强（中优先级）

**目标**：支持更细粒度的权限控制

**任务清单**：

1. **权限检查**
   - [ ] 检查用户是否有权限查看该设备的报告
   - [ ] 支持 `AssignedOnly` 权限过滤（仅查看分配给用户的设备）
   - [ ] 支持 `BranchOnly` 权限过滤（仅查看同一 Branch 的设备）

2. **参考实现**
   - [ ] 参考 `ResidentService` 的权限检查逻辑
   - [ ] 使用 `GetResourcePermission` 和 `ApplyBranchFilter`

**预计工作量**：1 天

---

### 阶段 4：数据迁移（低优先级，可选）

**目标**：如果 v1.0 的 MySQL 数据库中有现有数据，迁移到 PostgreSQL

**任务清单**：

1. **数据迁移脚本**
   - [ ] 创建数据迁移脚本
   - [ ] 从 MySQL 读取数据
   - [ ] 写入 PostgreSQL `sleepace_report` 表
   - [ ] 数据验证和错误处理

2. **迁移验证**
   - [ ] 数据完整性检查
   - [ ] 数据一致性验证

**预计工作量**：1 天（如果需要）

---

### 阶段 5：后台任务（可选）

**目标**：自动定期下载报告

**任务清单**：

1. **定时任务**
   - [ ] 实现定时任务（如每天凌晨下载前一天的报告）
   - [ ] 使用 cron 或类似机制
   - [ ] 错误处理和日志记录

2. **MQTT 触发**（参考 v1.0）
   - [ ] 监听 MQTT 消息
   - [ ] 触发报告下载
   - [ ] 参考：`wisefido-backend/wisefido-sleepace/modules/borker.go`

**预计工作量**：2-3 天（如果需要）

---

## 📋 详细实现计划

### 阶段 0 详细步骤（测试数据准备）

#### 0.1 创建测试数据 SQL 脚本

**文件**：`owlRD/db/test_data_sleepace_report.sql`

```sql
-- 测试数据：Sleepace 睡眠报告
-- 用途：开发阶段测试，模拟真实报告数据
-- 注意：需要先有 devices 表中的设备记录

-- 示例：为设备 device-uuid-1 插入测试报告
INSERT INTO sleepace_report (
    tenant_id,
    device_id,
    device_code,
    record_count,
    start_time,
    end_time,
    date,
    stop_mode,
    time_step,
    timezone,
    sleep_state,
    report,
    created_at,
    updated_at
) VALUES (
    '00000000-0000-0000-0000-000000000001'::uuid,  -- System tenant
    'device-uuid-1'::uuid,  -- 需要替换为实际的 device_id
    'SP001',  -- device_code（对应 devices.serial_number 或 devices.uid）
    1440,  -- record_count（24小时，每分钟一条）
    EXTRACT(EPOCH FROM '2024-08-20 00:00:00'::timestamptz)::bigint,  -- start_time
    EXTRACT(EPOCH FROM '2024-08-21 00:00:00'::timestamptz)::bigint,  -- end_time
    20240820,  -- date (YYYYMMDD)
    0,  -- stop_mode
    60,  -- time_step（秒）
    28800,  -- timezone（UTC+8，8*3600秒）
    '[1,1,1,2,2,2,3,3,3,2,2,1,1,1]',  -- sleep_state（JSON 数组字符串）
    '[{"summary":{"recordCount":1440,"startTime":1721491200,"stopMode":0,"timeStep":60,"timezone":28800},"analysis":{"sleepStateStr":[1,1,1,2,2,2,3,3,3,2,2,1,1,1]}}]',  -- report（完整 JSON 字符串）
    NOW(),
    NOW()
) ON CONFLICT (tenant_id, device_id, date) DO NOTHING;
```

#### 0.2 使用说明

1. **前提条件**：
   - 确保 `devices` 表中有测试设备记录
   - 确保 `tenant_id` 存在

2. **加载测试数据**：
   ```bash
   psql -h localhost -U postgres -d owlrd -f db/test_data_sleepace_report.sql
   ```

3. **验证数据**：
   ```sql
   SELECT * FROM sleepace_report ORDER BY date DESC LIMIT 10;
   ```

---

### 阶段 1 详细步骤（待设备接入后）

#### 1.1 添加配置

**文件**：`owlBack/wisefido-data/internal/config/config.go`

```go
type SleepaceConfig struct {
    HttpAddress string `yaml:"http_address"`  // Sleepace 厂家服务地址
    AppID       string `yaml:"app_id"`        // App ID
    ChannelID   string `yaml:"channel_id"`    // Channel ID
    SecretKey   string `yaml:"secret_key"`    // Secret Key
    Timezone    int    `yaml:"timezone"`      // 时区偏移（秒）
}
```

#### 1.2 实现 HTTP 客户端

**文件**：`owlBack/wisefido-data/internal/service/sleepace_client.go`（新建）

```go
type SleepaceClient struct {
    httpClient *http.Client
    baseURL    string
    appID      string
    channelID  string
    secretKey  string
    token      string
    tokenExpiry time.Time
    mu         sync.RWMutex
}

func (c *SleepaceClient) Get24HourDailyWithMaxReport(deviceID, deviceCode string, startTime, endTime int64) ([]json.RawMessage, error)
```

#### 1.3 实现 DownloadReport Service 方法

**文件**：`owlBack/wisefido-data/internal/service/sleepace_report_service.go`

```go
type DownloadReportRequest struct {
    TenantID   string
    DeviceID   string
    DeviceCode string
    StartTime  int64  // Unix 时间戳（秒）
    EndTime    int64  // Unix 时间戳（秒）
}

func (s *sleepaceReportService) DownloadReport(ctx context.Context, req DownloadReportRequest) error
```

#### 1.4 实现 DownloadReport Handler

**文件**：`owlBack/wisefido-data/internal/http/sleepace_report_handler.go`

```go
// POST /sleepace/api/v1/sleepace/reports/:id/download
func (h *SleepaceReportHandler) DownloadReport(w http.ResponseWriter, r *http.Request, deviceID string)
```

---

## 🎯 优先级总结

| 阶段 | 优先级 | 预计工作量 | 依赖 | 状态 |
|------|--------|-----------|------|------|
| 阶段 0：测试数据准备 | 🔴 高 | 0.5 天 | 无 | ✅ 已完成 |
| 阶段 1：数据下载和保存 | ⏸️ 暂停 | 2-3 天 | 设备接入 | ⏸️ 待设备接入 |
| 阶段 2：测试 | 🟡 中 | 1-2 天 | 阶段 0 | 🟡 进行中 |
| 阶段 3：权限检查增强 | 🟡 中 | 1 天 | 无 | 🟡 待开始 |
| 阶段 4：数据迁移 | 🟢 低 | 1 天 | 可选 | 🟢 待决定 |
| 阶段 5：后台任务 | 🟢 低 | 2-3 天 | 可选 | 🟢 待决定 |

---

## 📝 建议的执行顺序

### 当前阶段（开发阶段，无设备）

1. **立即开始**：阶段 0（测试数据准备）
   - 创建测试数据 SQL 脚本
   - 支持前端开发和测试
   - 验证查询功能是否正常工作

2. **并行进行**：阶段 2（测试）
   - 编写单元测试和集成测试
   - 使用测试数据验证功能

3. **后续优化**：阶段 3（权限检查增强）
   - 在基本功能稳定后，再增强权限检查

### 设备接入后

4. **设备接入后**：阶段 1（数据下载和保存功能）
   - 实现从厂家服务下载报告的功能
   - 这是生产环境的核心功能

5. **按需进行**：阶段 4 和 5
   - 根据实际需求决定是否实现

---

## 🔗 相关文档

- `SLEEPACE_REPORT_SERVICE_IMPLEMENTATION.md` - 实现总结
- `SLEEPACE_REPORT_V1.0_IMPLEMENTATION_ANALYSIS.md` - v1.0 实现分析
- `SLEEPACE_REPORT_DATA_SOURCE_CLARIFICATION.md` - 数据来源澄清
- `SLEEPACE_REPORT_DEVICE_CODE_CLARIFICATION.md` - device_code 说明

---

## ✅ 完成标准

### 阶段 1 完成标准

- [ ] 可以从 Sleepace 厂家服务下载报告
- [ ] 报告数据正确保存到 PostgreSQL
- [ ] 支持通过 API 手动触发下载
- [ ] 错误处理完善（网络错误、API 错误、数据解析错误）
- [ ] 日志记录完整

### 阶段 2 完成标准

- [ ] 单元测试覆盖率 > 80%
- [ ] 集成测试通过
- [ ] 兼容性测试通过

### 阶段 3 完成标准

- [ ] 权限检查逻辑正确
- [ ] 支持 `AssignedOnly` 和 `BranchOnly` 过滤
- [ ] 测试通过

