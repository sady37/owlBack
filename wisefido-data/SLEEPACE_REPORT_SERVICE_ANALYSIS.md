# SleepaceReportService 架构分析

## 📋 当前状态

### wisefido-data（v1.5）
- **当前实现**：`StubHandler.SleepaceReports`（stub 占位）
- **路由**：`/sleepace/api/v1/sleepace/reports/:id`
- **状态**：❌ 未实现

### wisefido-backend（v1.0）
- **服务**：`wisefido-sleepace` - HTTP API 服务
- **架构**：Handler → Repository → MySQL
- **是否有 Service 层**：❓ 未知（需要查看 v1.0 代码）

---

## 🔍 数据来源分析

### 方案 A：从时间序列数据聚合生成报告

**数据流**：
```
Sleepace 设备
    ↓
wisefido-sleepace 服务（后台服务）
    ├─ MQTT 订阅（Sleepace 厂家 MQTT）
    └─→ Redis Streams (sleepace:data:stream)
        ↓
wisefido-data-transformer 服务
    ├─ 消费 sleepace:data:stream
    ├─ 数据标准化（SNOMED CT映射）
    └─→ PostgreSQL TimescaleDB (iot_timeseries)
        ↓
wisefido-data 服务（HTTP API）
    ├─ SleepaceReportService
    ├─ 从 iot_timeseries 表聚合生成报告
    └─→ 返回给前端
```

**特点**：
- ✅ 数据已标准化（SNOMED CT 编码、FHIR Category）
- ✅ 数据存储在 `iot_timeseries` 表
- ✅ 需要从时间序列数据聚合生成睡眠报告
- ⚠️ **需要 Service 层**：数据聚合逻辑复杂

### 方案 B：调用 Sleepace 厂家服务

**数据流**：
```
Sleepace 设备
    ↓
Sleepace 厂家服务（第三方）
    ├─ HTTP API（提供睡眠报告）
    └─→ 直接返回报告数据
        ↓
wisefido-data 服务（HTTP API）
    ├─ SleepaceReportService
    ├─ 调用 Sleepace 厂家 HTTP API
    ├─ 数据转换和格式化
    └─→ 返回给前端
```

**特点**：
- ✅ 直接调用外部服务，无需数据聚合
- ✅ Sleepace 厂家服务已生成报告
- ⚠️ **需要 Service 层**：外部服务调用、数据转换、错误处理

### 方案 C：从 MySQL 数据库查询（v1.0 方式）

**数据流**：
```
Sleepace 设备
    ↓
Sleepace 厂家服务（第三方）
    └─→ MySQL 数据库（sleepace_report, sleepace_realtime_record 等表）
        ↓
wisefido-data 服务（HTTP API）
    ├─ SleepaceReportService
    ├─ 从 MySQL 查询报告数据
    ├─ 数据转换和格式化
    └─→ 返回给前端
```

**特点**：
- ✅ 数据已存储在 MySQL（v1.0 格式）
- ✅ 直接查询，无需聚合
- ⚠️ **需要 Service 层**：数据转换（v1.0 格式 → v1.5 格式）

---

## 🎯 架构决策

### 是否需要 SleepaceReportService？

**结论**：✅ **需要 Service 层**

**原因**：

1. **数据转换复杂**
   - 方案 A：从时间序列数据聚合生成报告（复杂）
   - 方案 B：调用外部服务，需要数据转换和错误处理
   - 方案 C：v1.0 格式 → v1.5 格式转换

2. **业务编排复杂**
   - 可能需要调用外部服务（Sleepace 厂家 HTTP API）
   - 需要数据聚合（时间序列数据 → 睡眠报告）
   - 需要权限检查（device_id 验证、tenant_id 过滤）

3. **错误处理**
   - 外部服务调用失败处理
   - 数据聚合失败处理
   - 数据格式验证

---

## 📊 实现方案对比

### 方案 1：直接调用（不推荐）

**架构**：
```
Handler → Repository → Database
```

**缺点**：
- ❌ 数据聚合逻辑在 Handler 中（违反单一职责）
- ❌ 外部服务调用逻辑在 Handler 中（难以测试）
- ❌ 数据转换逻辑在 Handler 中（代码重复）
- ❌ 难以复用（其他 Handler 无法使用）

### 方案 2：使用 Service 层（推荐）✅

**架构**：
```
Handler → Service → Repository / External Service
```

**优点**：
- ✅ 业务逻辑集中在 Service 层
- ✅ 易于测试（Service 可以独立测试）
- ✅ 易于复用（其他 Handler 可以调用）
- ✅ 职责清晰（Handler 只负责 HTTP 处理）

---

## 🔧 推荐实现方式

### 方案 A：从时间序列数据聚合生成报告（推荐）

**实现步骤**：

1. **创建 SleepaceReportService**
   ```go
   type SleepaceReportService interface {
       GetSleepaceReports(ctx context.Context, req GetSleepaceReportsRequest) (*GetSleepaceReportsResponse, error)
       GetSleepaceReportDetail(ctx context.Context, req GetSleepaceReportDetailRequest) (*GetSleepaceReportDetailResponse, error)
       GetSleepaceReportDates(ctx context.Context, req GetSleepaceReportDatesRequest) (*GetSleepaceReportDatesResponse, error)
   }
   ```

2. **实现数据聚合逻辑**
   - 从 `iot_timeseries` 表查询 Sleepace 数据
   - 按日期聚合数据
   - 生成睡眠报告（睡眠阶段、心率、呼吸率等）

3. **实现权限检查**
   - 验证 device_id 是否存在
   - 验证 tenant_id 权限
   - 验证用户是否有权限查看该设备

4. **实现数据转换**
   - 时间序列数据 → 睡眠报告格式
   - 格式化响应数据

### 方案 B：调用 Sleepace 厂家服务（备选）

**实现步骤**：

1. **创建 SleepaceReportService**
   ```go
   type SleepaceReportService interface {
       GetSleepaceReports(ctx context.Context, req GetSleepaceReportsRequest) (*GetSleepaceReportsResponse, error)
       GetSleepaceReportDetail(ctx context.Context, req GetSleepaceReportDetailRequest) (*GetSleepaceReportDetailResponse, error)
       GetSleepaceReportDates(ctx context.Context, req GetSleepaceReportDatesRequest) (*GetSleepaceReportDatesResponse, error)
   }
   ```

2. **实现外部服务调用**
   - HTTP 客户端调用 Sleepace 厂家 API
   - 错误处理和重试逻辑
   - 超时控制

3. **实现数据转换**
   - Sleepace 厂家格式 → v1.5 格式
   - 数据验证和清洗

4. **实现权限检查**
   - 验证 device_id 是否存在
   - 验证 tenant_id 权限

---

## 📋 与 wisefido-backend 对比

### wisefido-backend（v1.0）

**架构推测**：
- **Handler** → **Repository** → **MySQL**
- ❓ **是否有 Service 层**：未知
- 如果 Handler 逻辑简单（直接查询 MySQL），可能没有 Service 层
- 如果 Handler 逻辑复杂（数据转换、外部服务调用），可能有 Service 层

### owlBack（v1.5）

**推荐架构**：
- **Handler** → **Service** → **Repository / External Service**
- ✅ **需要 Service 层**：数据聚合、外部服务调用、数据转换

**原因**：
- v1.5 架构更注重分层和职责分离
- Service 层可以独立测试和复用
- 符合领域驱动设计原则

---

## ✅ 最终建议

### 推荐方案：使用 Service 层

**理由**：
1. ✅ **数据转换复杂**：需要从时间序列数据聚合生成报告，或调用外部服务
2. ✅ **业务编排复杂**：需要权限检查、数据聚合、外部服务调用
3. ✅ **易于测试**：Service 层可以独立测试
4. ✅ **易于复用**：其他 Handler 可以调用 Service
5. ✅ **符合架构原则**：职责分离、单一职责

**实现优先级**：
- **高优先级**：前端已使用，需要尽快实现
- **实现方式**：根据数据来源选择方案 A 或方案 B

---

## 📚 参考文档

- `RADAR_SLEEPACE_SERVICE_ANALYSIS.md` - Radar/Sleepace 服务分析
- `docs/09_Sleepace_v1.0_Architecture_Analysis.md` - Sleepace v1.0 架构分析
- `docs/11_Sleepace_Unified_Data_Flow_Implementation.md` - Sleepace 统一数据流实现
- `SERVICE_COMPLETION_STATUS.md` - Service 完成状态

