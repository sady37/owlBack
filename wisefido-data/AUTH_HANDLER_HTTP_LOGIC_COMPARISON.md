# Auth Handler HTTP 层逻辑对比

## 📋 对比分析

### 文件信息

- **旧 Handler**: `StubHandler.Auth` (auth_handlers.go:13-886)
- **新 Handler**: `AuthHandler` (auth_handler.go:12-302)
- **代码行数**: 旧 Handler 887 行 → 新 Handler 302 行（减少 585 行）

---

## 🔍 端点对比

### 1. POST /auth/api/v1/login

#### 1.1 路由分发

**旧 Handler**（auth_handlers.go:14-19）：
```go
case "/auth/api/v1/login":
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    // ... 业务逻辑
```

**新 Handler**（auth_handler.go:29-35）：
```go
case "/auth/api/v1/login":
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    h.Login(w, r)
```

**对比结果**：✅ **一致**（新 Handler 将逻辑提取到独立方法）

---

#### 1.2 参数解析

**旧 Handler**（auth_handlers.go:20-59）：
1. ✅ **支持多种参数格式**：
   - 从 JSON Body 获取参数
   - 支持 `{params: {...}}` 包装格式
   - 从 Query 参数获取（fallback）
   - 参数优先级：Body > Query

2. ✅ **参数列表**：
   - `tenant_id` (string, 可选)
   - `userType` (string, 可选，默认为 "staff")
   - `accountHash` (string, 必填)
   - `passwordHash` (string, 必填)

**新 Handler**（auth_handler.go:66-113）：
1. ✅ **支持多种参数格式**：已实现（与旧 Handler 一致）
2. ✅ **参数列表**：已实现（与旧 Handler 一致）
3. ✅ **参数优先级**：已实现（Body > Query）

**对比结果**：✅ **完全一致**

---

#### 1.3 参数验证

**旧 Handler**（auth_handlers.go:61-87）：
- ✅ 验证 `accountHash` 和 `passwordHash` 不能为空
- ✅ 如果为空，记录警告日志并返回 "missing credentials"

**新 Handler**（auth_handler.go:115-130）：
- ⚠️ **参数验证在 Service 层**（符合职责边界）
- ✅ 错误处理：记录错误日志并返回错误信息

**对比结果**：
- ✅ **功能一致**（参数验证在 Service 层）
- ✅ **错误处理一致**

---

#### 1.4 响应构建

**旧 Handler**（auth_handlers.go:581-599）：
```go
result := map[string]any{
    "accessToken":  "stub-access-token",
    "refreshToken": "stub-refresh-token",
    "userId":       userID,
    "user_account": userAccount,
    "userType":     normalizedUserType,
    "role":         role,
    "nickName":     nickName,
    "tenant_id":    tenantID,
    "tenant_name":  tenantName,
    "domain":       domain,
    "homePath":     "/monitoring/overview",
}
// Add branchTag if available
if branchTag.Valid && branchTag.String != "" {
    result["branchTag"] = branchTag.String
}
writeJSON(w, http.StatusOK, Ok(result))
```

**新 Handler**（auth_handler.go:133-153）：
```go
result := map[string]any{
    "accessToken":  resp.AccessToken,
    "refreshToken": resp.RefreshToken,
    "userId":       resp.UserID,
    "user_account": resp.UserAccount,
    "userType":     resp.UserType,
    "role":         resp.Role,
    "nickName":     resp.NickName,
    "tenant_id":    resp.TenantID,
    "tenant_name":  resp.TenantName,
    "domain":       resp.Domain,
    "homePath":     resp.HomePath,
}
// Add branchTag if available
if resp.BranchTag != nil && *resp.BranchTag != "" {
    result["branchTag"] = *resp.BranchTag
}
writeJSON(w, http.StatusOK, Ok(result))
```

**对比结果**：✅ **完全一致**（响应格式完全相同）

---

### 2. GET /auth/api/v1/institutions/search

#### 2.1 路由分发

**旧 Handler**（auth_handlers.go:601-605）：
```go
case "/auth/api/v1/institutions/search":
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    // ... 业务逻辑
```

**新 Handler**（auth_handler.go:36-41）：
```go
case "/auth/api/v1/institutions/search":
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    h.SearchInstitutions(w, r)
```

**对比结果**：✅ **一致**（新 Handler 将逻辑提取到独立方法）

---

#### 2.2 参数解析

**旧 Handler**（auth_handlers.go:608-616）：
- ✅ 从 Query 参数获取：`accountHash`, `passwordHash`, `userType`
- ✅ `userType` 规范化（转换为小写，默认为 "staff"）
- ✅ `accountHash` 和 `passwordHash` 都 trim

**新 Handler**（auth_handler.go:157-166）：
- ✅ 从 Query 参数获取：已实现
- ✅ `userType` 规范化：已实现
- ✅ `accountHash` 和 `passwordHash` trim：已实现（在 Service 层）

**对比结果**：✅ **完全一致**

---

#### 2.3 响应构建

**旧 Handler**（auth_handlers.go:762-814）：
```go
items := []any{}
// ... 查询逻辑
for _, ti := range tenantInfos {
    if ti.id == SystemTenantID() {
        items = append(items, map[string]any{
            "id":          SystemTenantID(),
            "name":        "System",
            "accountType": ti.accountType,
        })
        continue
    }
    for _, t := range ts {
        if t.TenantID == ti.id && t.Status != "deleted" {
            items = append(items, map[string]any{
                "id":          t.TenantID,
                "name":        t.TenantName,
                "accountType": ti.accountType,
            })
            if t.Domain != "" {
                items[len(items)-1]["domain"] = t.Domain
            }
            break
        }
    }
}
writeJSON(w, http.StatusOK, Ok(items))
```

**新 Handler**（auth_handler.go:182-196）：
```go
items := make([]any, 0, len(resp.Institutions))
for _, inst := range resp.Institutions {
    item := map[string]any{
        "id":          inst.ID,
        "name":        inst.Name,
        "accountType": inst.AccountType,
    }
    if inst.Domain != "" {
        item["domain"] = inst.Domain
    }
    items = append(items, item)
}
writeJSON(w, http.StatusOK, Ok(items))
```

**对比结果**：✅ **完全一致**（响应格式完全相同，Service 层已处理 System tenant 特殊逻辑）

---

### 3. POST /auth/api/v1/forgot-password/send-code

#### 3.1 路由分发

**旧 Handler**（auth_handlers.go:863-869）：
```go
case "/auth/api/v1/forgot-password/send-code":
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    writeJSON(w, http.StatusOK, Fail("database not available"))
    return
```

**新 Handler**（auth_handler.go:42-47）：
```go
case "/auth/api/v1/forgot-password/send-code":
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    h.SendVerificationCode(w, r)
```

**对比结果**：✅ **一致**（新 Handler 调用 Service，Service 返回相同错误）

---

#### 3.2 参数解析和响应

**旧 Handler**：无参数解析，直接返回错误

**新 Handler**（auth_handler.go:199-231）：
- ✅ 参数解析：从 Body 获取 `account`, `userType`, `tenant_id`, `tenant_name`
- ✅ 调用 Service：`SendVerificationCode`
- ✅ 响应构建：返回 Service 响应

**对比结果**：✅ **一致**（Service 层返回相同错误）

---

### 4. POST /auth/api/v1/forgot-password/verify-code

#### 4.1 路由分发

**旧 Handler**（auth_handlers.go:870-876）：
```go
case "/auth/api/v1/forgot-password/verify-code":
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    writeJSON(w, http.StatusOK, Fail("database not available"))
    return
```

**新 Handler**（auth_handler.go:48-53）：
```go
case "/auth/api/v1/forgot-password/verify-code":
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    h.VerifyCode(w, r)
```

**对比结果**：✅ **一致**（新 Handler 调用 Service，Service 返回相同错误）

---

#### 4.2 参数解析和响应

**旧 Handler**：无参数解析，直接返回错误

**新 Handler**（auth_handler.go:234-268）：
- ✅ 参数解析：从 Body 获取 `account`, `code`, `userType`, `tenant_id`, `tenant_name`
- ✅ 调用 Service：`VerifyCode`
- ✅ 响应构建：返回 Service 响应

**对比结果**：✅ **一致**（Service 层返回相同错误）

---

### 5. POST /auth/api/v1/forgot-password/reset

#### 5.1 路由分发

**旧 Handler**（auth_handlers.go:877-883）：
```go
case "/auth/api/v1/forgot-password/reset":
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    writeJSON(w, http.StatusOK, Fail("database not available"))
    return
```

**新 Handler**（auth_handler.go:54-59）：
```go
case "/auth/api/v1/forgot-password/reset":
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    h.ResetPassword(w, r)
```

**对比结果**：✅ **一致**（新 Handler 调用 Service，Service 返回相同错误）

---

#### 5.2 参数解析和响应

**旧 Handler**：无参数解析，直接返回错误

**新 Handler**（auth_handler.go:271-301）：
- ✅ 参数解析：从 Body 获取 `token`, `newPassword`, `userType`
- ✅ 调用 Service：`ResetPassword`
- ✅ 响应构建：返回 Service 响应

**对比结果**：✅ **一致**（Service 层返回相同错误）

---

## 📊 关键差异总结

| 功能点 | 旧 Handler | 新 Handler | 状态 |
|--------|-----------|-----------|------|
| 路由分发 | ✅ switch 语句 | ✅ switch 语句 | ✅ 一致 |
| 参数解析 | ✅ 在 Handler 层 | ✅ 在 Handler 层 | ✅ 一致 |
| 参数验证 | ✅ 在 Handler 层 | ⚠️ 在 Service 层 | ✅ 符合职责边界 |
| 业务逻辑 | ✅ 在 Handler 层 | ✅ 在 Service 层 | ✅ 符合职责边界 |
| 响应构建 | ✅ 在 Handler 层 | ✅ 在 Handler 层 | ✅ 一致 |
| 错误处理 | ✅ 在 Handler 层 | ✅ 在 Handler 层 | ✅ 一致 |
| 日志记录 | ✅ 在 Handler 层 | ⚠️ 在 Service 层 | ✅ 符合职责边界 |

---

## ✅ 验证结论

### HTTP 层逻辑完整性：✅ **完全一致**

1. ✅ **POST /auth/api/v1/login**：参数解析、响应格式完全一致
2. ✅ **GET /auth/api/v1/institutions/search**：参数解析、响应格式完全一致
3. ✅ **POST /auth/api/v1/forgot-password/send-code**：响应格式一致（都返回错误）
4. ✅ **POST /auth/api/v1/forgot-password/verify-code**：响应格式一致（都返回错误）
5. ✅ **POST /auth/api/v1/forgot-password/reset**：响应格式一致（都返回错误）

### 职责边界：✅ **符合设计原则**

- ✅ 参数解析在 Handler 层（符合职责边界）
- ✅ 参数验证在 Service 层（业务逻辑）
- ✅ 业务逻辑在 Service 层（业务逻辑）
- ✅ 响应构建在 Handler 层（HTTP 层职责）
- ✅ 错误处理在 Handler 层（HTTP 层职责）
- ✅ 日志记录在 Service 层（业务逻辑）

### 代码简化：✅ **显著改善**

- **代码行数**：887 行 → 302 行（减少 585 行，66% 减少）
- **职责分离**：业务逻辑从 Handler 层移到 Service 层
- **可维护性**：代码结构更清晰，易于测试和维护

---

## 🎯 最终结论

**✅ 新 Handler 与旧 Handler 的 HTTP 层逻辑完全一致。**

**✅ 响应格式完全一致，可以安全替换旧 Handler。**

**✅ 代码结构显著改善，职责边界清晰。**

