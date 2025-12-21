# Sleepace Report Handler 测试指南

## 📋 测试文件

**文件**：`internal/http/sleepace_report_handler_test.go`

**测试类型**：集成测试（需要数据库连接）

**运行方式**：
```bash
# 运行所有 Sleepace Report Handler 测试
go test -tags=integration -v ./internal/http -run TestSleepaceReportHandler

# 运行单个测试
go test -tags=integration -v ./internal/http -run TestSleepaceReportHandler_Resident_CanViewOwnReports
```

---

## ✅ 已实现的测试用例

### 1. TestSleepaceReportHandler_Resident_CanViewOwnReports

**测试目标**：验证住户可以查看自己的睡眠报告

**测试场景**：
- 创建测试数据（unit, room, bed, resident, device）
- 设备关联到住户的床位
- 住户请求查看自己的报告
- **预期结果**：权限检查通过，返回成功

---

### 2. TestSleepaceReportHandler_Resident_CannotViewOtherReports

**测试目标**：验证住户不能查看其他住户的睡眠报告

**测试场景**：
- 创建测试数据（device 关联到住户 A）
- 住户 B 请求查看住户 A 的报告
- **预期结果**：权限检查失败，返回 "access denied"

---

### 3. TestSleepaceReportHandler_Caregiver_CanViewAssignedResidentReports

**测试目标**：验证 Caregiver 可以查看分配的住户报告

**测试场景**：
- 创建测试数据（device, resident, caregiver）
- 配置 `role_permissions`（Caregiver assigned_only=true）
- 创建 `resident_caregivers` 记录（userList 包含 caregiverID）
- Caregiver 请求查看分配的住户报告
- **预期结果**：权限检查通过，返回成功

---

### 4. TestSleepaceReportHandler_Manager_CanViewBranchResidentReports

**测试目标**：验证 Manager 可以查看同分支的住户报告

**测试场景**：
- 创建测试数据（device, resident, manager）
- 配置 `role_permissions`（Manager branch_only=true）
- Manager 的 `branch_tag` = "BranchA"
- 住户的 `branch_tag` = "BranchA"（通过 unit）
- Manager 请求查看同分支的住户报告
- **预期结果**：权限检查通过，返回成功

---

### 5. TestSleepaceReportHandler_DeviceWithoutResident_Allowed

**测试目标**：验证设备没有关联住户时允许访问（fallback）

**测试场景**：
- 创建设备（不关联住户）
- 任何用户请求查看该设备的报告
- **预期结果**：权限检查通过（fallback），返回成功

---

## 🔧 测试辅助函数

### setupSleepaceTestData

**功能**：创建完整的测试数据

**创建的数据**：
1. Unit（单元）- branch_tag = "BranchA"
2. Room（房间）
3. Bed（床位）
4. Resident（住户）
5. Device（设备）- 关联到床位
6. Caregiver 用户
7. Manager 用户 - branch_tag = "BranchA"
8. 权限配置（role_permissions）
9. 住户分配关系（resident_caregivers）

**返回**：deviceID, residentID, unitID, roomID, bedID, caregiverID, managerID

---

### cleanupSleepaceTestData

**功能**：清理测试数据

**清理的表**：
- `resident_caregivers`
- `sleepace_report`
- `devices`
- `residents`
- `beds`
- `rooms`
- `units`
- `users`
- `role_permissions`

---

## 📝 测试数据说明

### 测试租户
- **tenantID**: `00000000-0000-0000-0000-000000000998`（由 `createTestTenantForHandler` 创建）

### 测试设备
- **deviceID**: `00000000-0000-0000-0000-000000000501`
- **device_name**: "Test Sleepace Device"
- **serial_number**: "SN123456"
- **bound_bed_id**: 关联到测试床位

### 测试住户
- **residentID**: `00000000-0000-0000-0000-000000000401`
- **resident_account**: "test_resident"
- **nickname**: "Test Resident"
- **unit_id**: 关联到测试单元（branch_tag = "BranchA"）

### 测试用户
- **caregiverID**: `00000000-0000-0000-0000-000000000601`
- **managerID**: `00000000-0000-0000-0000-000000000701`
- **Manager branch_tag**: "BranchA"

---

## 🧪 运行测试

### 前置条件

1. **数据库连接**：
   - 需要可用的 PostgreSQL 数据库
   - 配置在 `owl-common/config` 中

2. **环境变量**（可选）：
   ```bash
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_USER=postgres
   export DB_PASSWORD=postgres
   export DB_NAME=owlrd
   ```

### 运行命令

```bash
# 进入项目目录
cd /Users/sady3721/project/owlBack/wisefido-data

# 运行所有测试
go test -tags=integration -v ./internal/http -run TestSleepaceReportHandler

# 运行单个测试
go test -tags=integration -v ./internal/http -run TestSleepaceReportHandler_Resident_CanViewOwnReports

# 运行测试并显示覆盖率
go test -tags=integration -v -cover ./internal/http -run TestSleepaceReportHandler
```

---

## 📊 测试覆盖范围

### 已覆盖

- ✅ 住户查看自己的报告（允许）
- ✅ 住户查看其他住户的报告（拒绝）
- ✅ Caregiver 查看分配的住户报告（允许）
- ✅ Manager 查看同分支的住户报告（允许）
- ✅ 设备没有关联住户（fallback 允许）

### 待覆盖

- ⏳ Caregiver 查看未分配的住户报告（拒绝）
- ⏳ Manager 查看不同分支的住户报告（拒绝）
- ⏳ Manager branch_tag 为 NULL 的情况
- ⏳ Family 用户查看报告
- ⏳ DownloadReport 权限检查（manage 权限）

---

## 🐛 已知问题

1. **go.mod 依赖问题**：
   - 当前有 `wisefido-alarm` 模块的依赖问题
   - 不影响测试逻辑，但需要先解决依赖问题才能运行测试

2. **测试数据清理**：
   - 测试数据使用固定的 UUID，可能与其他测试冲突
   - 建议使用随机 UUID 或更好的清理策略

---

## 📝 后续改进

1. **添加更多测试用例**：
   - Caregiver 未分配的情况
   - Manager 不同分支的情况
   - Manager branch_tag 为 NULL 的情况
   - Family 用户的情况

2. **改进测试数据**：
   - 使用随机 UUID
   - 更好的数据隔离
   - 更完整的测试场景

3. **性能测试**：
   - 测试权限检查的性能
   - 测试大量数据的情况

