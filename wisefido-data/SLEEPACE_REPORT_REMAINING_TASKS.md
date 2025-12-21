# Sleepace Report Service 和 Handler 待处理事项

## 📋 当前状态检查

### ✅ 已完成

1. **核心功能**
   - ✅ 查询报告列表
   - ✅ 查询报告详情
   - ✅ 查询有效日期列表
   - ✅ 手动触发下载报告（Handler + Service）

2. **基础架构**
   - ✅ Repository 层（PostgreSQL）
   - ✅ Service 层（业务逻辑）
   - ✅ Handler 层（HTTP 处理）
   - ✅ 路由注册

3. **数据同步**
   - ✅ 手动触发下载 API
   - ⏳ MQTT 触发下载（框架已创建，待实现）

---

## ⏳ 待处理事项

### 1. 权限检查（高优先级）⚠️

**问题**：当前 Handler 没有权限检查

**现状**：
- ✅ 其他 Handler（如 `ResidentHandler`、`DeviceMonitorSettingsHandler`）有权限检查
- ❌ `SleepaceReportHandler` 没有权限检查
- ✅ 有设备验证（验证设备是否存在且属于该租户）

**需要添加的权限检查**：

#### 1.1 查询权限（GetSleepaceReports, GetSleepaceReportDetail, GetSleepaceReportDates）

**参考**：`DeviceMonitorSettingsHandler` 的权限检查

**实现方式**：
```go
// 在 Handler 方法中添加权限检查
userID := r.Header.Get("X-User-Id")
userRole := r.Header.Get("X-User-Role")

// 检查设备权限
perm, err := GetResourcePermission(h.db, ctx, tenantID, userID, userRole, "device", deviceID, "read")
if err != nil {
    writeJSON(w, http.StatusOK, Fail(err.Error()))
    return
}
if !perm.Allowed {
    writeJSON(w, http.StatusOK, Fail("access denied"))
    return
}

// 应用分支过滤（如果 perm.BranchOnly 为 true）
if perm.BranchOnly {
    // TODO: 应用分支过滤
}
```

**权限类型**：
- `resource_type`: `"device"`
- `permission_type`: `"read"`（查询报告）
- `permission_type`: `"manage"`（下载报告）

#### 1.2 下载权限（DownloadReport）

**需要权限**：
- `resource_type`: `"device"`
- `permission_type`: `"manage"`（管理权限，包括下载报告）

**实现方式**：
```go
// 在 DownloadReport 方法中添加权限检查
userID := r.Header.Get("X-User-Id")
userRole := r.Header.Get("X-User-Role")

// 检查设备管理权限
perm, err := GetResourcePermission(h.db, ctx, tenantID, userID, userRole, "device", deviceID, "manage")
if err != nil {
    writeJSON(w, http.StatusOK, Fail(err.Error()))
    return
}
if !perm.Allowed {
    writeJSON(w, http.StatusOK, Fail("access denied: manage permission required"))
    return
}
```

**参考文件**：
- `internal/http/permission_utils.go` - `GetResourcePermission` 函数
- `internal/http/device_monitor_settings_handler.go` - 设备权限检查示例
- `internal/http/resident_handler.go` - 复杂权限检查示例

---

### 2. 错误处理优化（中优先级）

**问题**：错误处理可以更细化

**当前状态**：
- ✅ 基本的错误处理（返回错误消息）
- ⚠️ 错误分类不够细致（如：设备不存在、权限不足、数据库错误）

**建议改进**：
```go
// 更细化的错误处理
if err != nil {
    h.logger.Error("GetSleepaceReports failed",
        zap.String("tenant_id", tenantID),
        zap.String("device_id", deviceID),
        zap.Error(err),
    )
    
    // 根据错误类型返回不同的错误码
    if strings.Contains(err.Error(), "not found") {
        writeJSON(w, http.StatusOK, Fail("device not found"))
    } else if strings.Contains(err.Error(), "access denied") {
        writeJSON(w, http.StatusOK, Fail("access denied"))
    } else {
        writeJSON(w, http.StatusOK, Fail(err.Error()))
    }
    return
}
```

---

### 3. 单元测试（中优先级）

**问题**：没有单元测试

**需要添加的测试**：

#### 3.1 Service 层测试

**文件**：`internal/service/sleepace_report_service_test.go`（新建）

**测试用例**：
- ✅ `TestGetSleepaceReports` - 测试获取报告列表
- ✅ `TestGetSleepaceReportDetail` - 测试获取报告详情
- ✅ `TestGetSleepaceReportDates` - 测试获取有效日期列表
- ✅ `TestDownloadReport` - 测试下载报告（需要 mock Sleepace 客户端）
- ✅ `TestValidateDevice` - 测试设备验证

**参考**：
- `internal/service/resident_service_test.go`
- `internal/service/user_service_integration_test.go`

#### 3.2 Handler 层测试

**文件**：`internal/http/sleepace_report_handler_test.go`（新建）

**测试用例**：
- ✅ 测试路径解析
- ✅ 测试查询参数解析
- ✅ 测试权限检查
- ✅ 测试错误处理

**参考**：
- `internal/http/auth_handler_test.go`

---

### 4. 响应格式验证（低优先级）

**问题**：需要验证响应格式是否与 v1.0 完全兼容

**检查项**：
- ✅ 报告列表响应格式
- ✅ 报告详情响应格式
- ✅ 日期列表响应格式
- ✅ 错误响应格式

**参考**：
- `SLEEPACE_REPORT_V1.0_IMPLEMENTATION_ANALYSIS.md`
- v1.0 的实际响应格式

---

### 5. 日志优化（低优先级）

**问题**：日志可以更详细

**建议改进**：
```go
// 添加更详细的日志
h.logger.Info("GetSleepaceReports",
    zap.String("tenant_id", tenantID),
    zap.String("device_id", deviceID),
    zap.Int("start_date", startDate),
    zap.Int("end_date", endDate),
    zap.Int("page", page),
    zap.Int("size", size),
)
```

---

### 6. 文档完善（低优先级）

**问题**：需要 API 文档

**需要添加**：
- ✅ API 端点文档
- ✅ 请求参数说明
- ✅ 响应格式说明
- ✅ 错误码说明
- ✅ 权限要求说明

---

## 🎯 优先级排序

### 高优先级（必须处理）

1. **权限检查** ⚠️
   - 查询权限（read）
   - 下载权限（manage）
   - 参考其他 Handler 的实现

### 中优先级（建议处理）

2. **单元测试**
   - Service 层测试
   - Handler 层测试

3. **错误处理优化**
   - 更细化的错误分类
   - 更友好的错误消息

### 低优先级（可选）

4. **响应格式验证**
5. **日志优化**
6. **文档完善**

---

## 📝 实施建议

### 第一步：添加权限检查（最重要）

**文件**：`internal/http/sleepace_report_handler.go`

**需要修改的方法**：
1. `GetSleepaceReports` - 添加 read 权限检查
2. `GetSleepaceReportDetail` - 添加 read 权限检查
3. `GetSleepaceReportDates` - 添加 read 权限检查
4. `DownloadReport` - 添加 manage 权限检查

**参考实现**：
- `internal/http/device_monitor_settings_handler.go`
- `internal/http/permission_utils.go`

---

## ✅ 总结

**当前状态**：
- ✅ 核心功能已完成
- ✅ 基础架构完整
- ⚠️ **缺少权限检查**（最重要）
- ⚠️ 缺少单元测试
- ⚠️ 错误处理可以优化

**建议**：
1. **立即处理**：添加权限检查
2. **后续处理**：添加单元测试
3. **可选处理**：优化错误处理和日志

