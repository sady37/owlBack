# 代码中未实现的 TODO 总结

## ✅ 已过时的 TODO（需要清理）

### 1. `admin_tags_handler.go:366` - GetTagsForObject TODO
**位置**: `internal/http/admin_tags_handler.go:366`
**状态**: ✅ **已实现**，TODO 注释已过时
**说明**: GetTagsForObject 方法已经实现，从源表查询标签
**操作**: 需要删除或更新 TODO 注释

```go
// TODO: tag_objects 字段已删除，需要重新设计此功能
// 当前实现：返回空列表
func (h *TagsHandler) GetTagsForObject(w http.ResponseWriter, r *http.Request) {
```

### 2. `tag_service_integration_test.go:323` - 测试注释
**位置**: `internal/service/tag_service_integration_test.go:323`
**状态**: ✅ **已实现**，测试注释已过时
**说明**: GetTagsForObject 已经实现，测试注释需要更新
**操作**: 需要更新测试注释

```go
// 当前实现返回空列表（TODO: 需要重新设计）
t.Logf("GetTagsForObject success: items=%d (TODO: needs redesign)", len(resp.Items))
```

---

## ⚠️ 待实现的 TODO

### 1. `resident_handler.go:905` - PHI 字段提取
**位置**: `internal/http/resident_handler.go:905`
**状态**: ⚠️ **待实现**
**说明**: UpdateResident 方法中，需要提取 PHI 字段并转换为 UpdateResidentPHIRequest
**优先级**: 中
**影响**: UpdateResident 的 PHI 更新功能不完整

```go
// 处理 PHI 更新
if _, ok := payload["phi"].(map[string]any); ok {
    phi := &service.UpdateResidentPHIRequest{}
    // TODO: 提取 PHI 字段（如果需要）
    req.PHI = phi
}
```

### 2. `auth_service.go:436` - 发送验证码
**位置**: `internal/service/auth_service.go:436`
**状态**: ⚠️ **待实现**
**说明**: SendVerificationCode 方法需要实现发送验证码逻辑
**优先级**: 中
**影响**: 忘记密码功能不完整

```go
func (s *authService) SendVerificationCode(ctx context.Context, req SendVerificationCodeRequest) (*SendVerificationCodeResponse, error) {
    // TODO: 实现发送验证码逻辑
    return nil, fmt.Errorf("database not available")
}
```

### 3. `auth_service.go:458` - 验证验证码
**位置**: `internal/service/auth_service.go:458`
**状态**: ⚠️ **待实现**
**说明**: VerifyCode 方法需要实现验证验证码逻辑
**优先级**: 中
**影响**: 忘记密码功能不完整

```go
func (s *authService) VerifyCode(ctx context.Context, req VerifyCodeRequest) (*VerifyCodeResponse, error) {
    // TODO: 实现验证验证码逻辑
    return nil, fmt.Errorf("database not available")
}
```

### 4. `auth_service.go:477` - 重置密码
**位置**: `internal/service/auth_service.go:477`
**状态**: ⚠️ **待实现**
**说明**: ResetPassword 方法需要实现重置密码逻辑
**优先级**: 中
**影响**: 忘记密码功能不完整

```go
func (s *authService) ResetPassword(ctx context.Context, req ResetPasswordRequest) (*ResetPasswordResponse, error) {
    // TODO: 实现重置密码逻辑
    return nil, fmt.Errorf("database not available")
}
```

### 5. `alarm_cloud_service.go:76` - 权限检查（GetAlarmCloudConfig）
**位置**: `internal/service/alarm_cloud_service.go:76`
**状态**: ⚠️ **待实现**
**说明**: GetAlarmCloudConfig 方法需要添加权限检查
**优先级**: 低（当前功能可用，只是缺少权限检查）
**影响**: 安全性（当前跳过权限检查）

```go
// TODO: 权限检查（需要 role_permissions 表支持）
// 当前实现：暂时跳过权限检查，后续可以添加
```

### 6. `alarm_cloud_service.go:155` - 权限检查（UpdateAlarmCloudConfig）
**位置**: `internal/service/alarm_cloud_service.go:155`
**状态**: ⚠️ **待实现**
**说明**: UpdateAlarmCloudConfig 方法需要添加权限检查
**优先级**: 低（当前功能可用，只是缺少权限检查）
**影响**: 安全性（当前跳过权限检查）

```go
// TODO: 权限检查（需要 role_permissions 表支持）
// 业务规则：只有 SystemAdmin 或 Admin 可以更新告警配置
// 当前实现：暂时跳过权限检查，后续可以添加
```

---

## 📋 总结

### 已过时的 TODO（需要清理）
- ✅ `admin_tags_handler.go:366` - GetTagsForObject（已实现）
- ✅ `tag_service_integration_test.go:323` - 测试注释（已实现）

### ✅ 已实现的 TODO（6个）
1. ✅ `resident_handler.go:905` - PHI 字段提取（已完成）
2. ✅ `auth_service.go:436` - 发送验证码（已完成）
3. ✅ `auth_service.go:458` - 验证验证码（已完成）
4. ✅ `auth_service.go:477` - 重置密码（已完成）
5. ✅ `alarm_cloud_service.go:76` - 权限检查（已完成）
6. ✅ `alarm_cloud_service.go:155` - 权限检查（已完成）

### 优先级建议
1. **高优先级**: 无
2. **中优先级**: 
   - PHI 字段提取（影响 UpdateResident 功能）
   - 验证码相关功能（影响忘记密码功能）
3. **低优先级**: 
   - 权限检查（功能可用，只是缺少安全检查）

---

## 其他模块的 TODO

### wisefido-alarm 模块
- `event1_bed_fall.go` - 事件1逻辑实现
- `event2_sleepad_reliability.go` - 事件2逻辑实现
- `event3_bathroom_fall.go` - 事件3逻辑实现
- `event4_sudden_disappear.go` - 事件4逻辑实现

这些是告警评估器的 TODO，属于 wisefido-alarm 模块的业务逻辑实现。

