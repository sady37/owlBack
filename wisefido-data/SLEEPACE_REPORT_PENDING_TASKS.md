# Sleepace Report 未完成任务清单

## 📊 当前完成状态

### ✅ 已完成

1. **核心查询功能**
   - ✅ 获取报告列表 (`GetSleepaceReports`)
   - ✅ 获取报告详情 (`GetSleepaceReportDetail`)
   - ✅ 获取有效日期列表 (`GetSleepaceReportDates`)

2. **数据下载功能**
   - ✅ 手动触发下载 API (`DownloadReport`)
   - ✅ Sleepace 厂家 API 客户端 (`SleepaceClient`)
   - ✅ 数据解析和保存逻辑

3. **权限检查**
   - ✅ 住户查看自己的报告
   - ✅ 住户不能查看其他住户的报告
   - ✅ Caregiver 查看分配的住户报告
   - ✅ Manager 查看同分支的住户报告
   - ✅ 设备没有关联住户时的 fallback 逻辑
   - ✅ 所有权限检查测试通过

4. **测试**
   - ✅ Handler 层集成测试（5 个测试全部通过）
   - ✅ 测试数据创建和清理
   - ✅ 权限检查测试

5. **数据库**
   - ✅ `sleepace_report` 表创建
   - ✅ Repository 层实现
   - ✅ `device_code` 与 `serial_number`/`uid` 等价关系处理

---

## ⏳ 未完成任务

### 1. MQTT 触发下载（高优先级，待设备接入后实现）

**状态**：⏸️ **框架已创建，具体逻辑待实现**

**文件**：`internal/mqtt/sleepace_broker.go`

**待实现功能**：

#### 1.1 消息解析和路由
- [ ] 实现 `HandleMessage` 方法（解析 MQTT 消息数组）
- [ ] 实现 `processMessage` 方法（根据 `dataKey` 路由）
- [ ] 定义消息模型（`models.go`）

#### 1.2 分析事件处理
- [ ] 实现 `handleAnalysisEvent` 方法（触发报告下载）
- [ ] 解析 `AnalysisData`（设备信息、时间范围）
- [ ] 通过 `device_code` 查询 `device_id` 和 `tenant_id`
- [ ] 调用 Service 层的 `DownloadReport` 方法

#### 1.3 MQTT 订阅管理
- [ ] 实现 `Start` 方法（订阅 MQTT 主题）
- [ ] 实现 `Stop` 方法（取消订阅）
- [ ] 在 `main.go` 中启用 MQTT（条件初始化）

**参考文档**：
- `MQTT_IMPLEMENTATION_TODO.md` - 详细的实现步骤和代码模板
- `SLEEPACE_REPORT_V1.0_DATA_SYNC_ANALYSIS.md` - v1.0 数据同步分析

**预计工作量**：2-3 天（待设备接入后）

**依赖**：设备接入后，MQTT 消息可用

---

### 2. Service 层单元测试（中优先级）

**状态**：⚠️ **未实现**

**文件**：`internal/service/sleepace_report_service_test.go`（需要创建）

**待实现测试**：

- [ ] `TestGetSleepaceReports` - 测试获取报告列表
- [ ] `TestGetSleepaceReportDetail` - 测试获取报告详情
- [ ] `TestGetSleepaceReportDates` - 测试获取有效日期列表
- [ ] `TestDownloadReport` - 测试下载报告（需要 mock Sleepace 客户端）
- [ ] `TestValidateDevice` - 测试设备验证
- [ ] `TestDeviceCodeResolution` - 测试 `device_code` 解析逻辑

**参考**：
- `internal/service/resident_service_test.go`
- `internal/service/user_service_integration_test.go`

**预计工作量**：1-2 天

---

### 3. 错误处理优化（低优先级）

**状态**：⚠️ **基本错误处理已实现，可以进一步优化**

**待优化项**：

- [ ] 更细化的错误分类（设备不存在、权限不足、数据库错误、网络错误等）
- [ ] 更友好的错误消息
- [ ] 错误码标准化

**预计工作量**：0.5 天

---

### 4. 日志优化（低优先级）

**状态**：⚠️ **基本日志已实现，可以进一步优化**

**待优化项**：

- [ ] 添加更详细的日志（请求参数、响应大小等）
- [ ] 性能日志（请求耗时）
- [ ] 结构化日志优化

**预计工作量**：0.5 天

---

### 5. 文档完善（低优先级）

**状态**：⚠️ **基本文档已创建，可以进一步完善**

**待完善项**：

- [ ] API 端点文档（详细的请求/响应格式）
- [ ] 错误码说明文档
- [ ] 权限要求说明文档
- [ ] 使用示例

**预计工作量**：1 天

---

### 6. 可选功能（低优先级，按需实现）

#### 6.1 周期性任务（可选）

**状态**：🟢 **可选功能**

**功能**：自动定期下载报告（如每天凌晨下载前一天的报告）

**预计工作量**：2-3 天

**参考**：v1.0 中没有周期性任务，主要依赖 MQTT 触发

#### 6.2 数据迁移（可选）

**状态**：🟢 **可选功能**

**功能**：如果 v1.0 的 MySQL 数据库中有现有数据，迁移到 PostgreSQL

**预计工作量**：1 天

**前提**：需要从 v1.0 迁移数据

---

## 🎯 优先级排序

### 高优先级（必须实现）

1. **MQTT 触发下载** ⏸️
   - 状态：框架已创建，待设备接入后实现
   - 预计工作量：2-3 天
   - 依赖：设备接入

### 中优先级（建议实现）

2. **Service 层单元测试**
   - 状态：未实现
   - 预计工作量：1-2 天
   - 依赖：无

### 低优先级（可选）

3. **错误处理优化**
   - 预计工作量：0.5 天

4. **日志优化**
   - 预计工作量：0.5 天

5. **文档完善**
   - 预计工作量：1 天

6. **可选功能**（周期性任务、数据迁移）
   - 按需实现

---

## 📋 实施建议

### 当前阶段（开发阶段，无设备）

1. **Service 层单元测试**（中优先级）
   - 可以立即开始
   - 不依赖设备
   - 提高代码质量和可维护性

2. **错误处理和日志优化**（低优先级）
   - 可以并行进行
   - 提升用户体验和可调试性

### 设备接入后

3. **MQTT 触发下载**（高优先级）
   - 设备接入后立即实现
   - 这是生产环境的核心功能
   - 参考 `MQTT_IMPLEMENTATION_TODO.md` 的详细步骤

---

## 📝 相关文档

- `MQTT_IMPLEMENTATION_TODO.md` - MQTT 实现详细步骤
- `SLEEPACE_REPORT_NEXT_STEPS.md` - 后续规划
- `SLEEPACE_REPORT_REMAINING_TASKS.md` - 待处理事项（部分已过期）
- `SLEEPACE_REPORT_V1.0_IMPLEMENTATION_ANALYSIS.md` - v1.0 实现分析
- `SLEEPACE_REPORT_V1.0_DATA_SYNC_ANALYSIS.md` - v1.0 数据同步分析
- `TEST_FINAL_RESULT.md` - 测试结果总结

---

## ✅ 完成标准

### MQTT 触发下载完成标准

- [ ] 所有 TODO 注释已实现
- [ ] 单元测试通过
- [ ] 集成测试通过
- [ ] 端到端测试通过（模拟 MQTT 消息）
- [ ] 文档已更新

### Service 层测试完成标准

- [ ] 测试覆盖率 > 80%
- [ ] 所有主要功能都有测试
- [ ] Mock 外部依赖（Sleepace 客户端）
- [ ] 测试通过

---

## 📊 总结

**当前状态**：
- ✅ 核心功能已完成（查询、下载、权限检查）
- ✅ Handler 层测试已完成
- ⏸️ MQTT 触发下载（框架已创建，待设备接入后实现）
- ⚠️ Service 层测试（未实现）
- ⚠️ 错误处理和日志优化（可选）

**建议**：
1. **立即开始**：Service 层单元测试（不依赖设备）
2. **设备接入后**：MQTT 触发下载（核心功能）
3. **后续优化**：错误处理、日志、文档（按需）

