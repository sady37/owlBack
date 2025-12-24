# User AccountSettings API 测试结果

## 代码修改完成

### 1. Service 层修改
- ✅ `GetAccountSettingsResponse` 添加了 `ID`, `Role`, `SaveEmail`, `SavePhone` 字段
- ✅ `GetAccountSettings` 实现已更新，返回所有必需字段
- ✅ `UpdateAccountSettings` 实现已更新，正确处理 null 值和只删除明文但保留 hash 的逻辑

### 2. Handler 层修改
- ✅ `GetAccountSettings` handler 返回所有字段
- ✅ `UpdateAccountSettings` handler 处理 null 值
- ✅ 路由顺序已调整（account-settings 路由在 GetUser/UpdateUser 之前）

### 3. 路由修复
- ✅ 路由顺序已修复：`GetAccountSettings` 和 `UpdateAccountSettings` 在 `GetUser` 和 `UpdateUser` 之前

## 需要重启后端服务

**重要**：代码已修改，但后端服务需要**重启**才能加载新代码。

如果使用 `go run`，需要：
1. 停止当前服务（Ctrl+C）
2. 重新运行：`cd /Users/sady3721/project/owlBack/wisefido-data && go run cmd/wisefido-data/main.go`

## 测试步骤（服务重启后）

### 测试账号
- **User ID**: `7f3fa6ff-3a07-49b2-a649-6ea086e93863`
- **Tenant ID**: `095c1ab6-5143-47ea-8670-5476158b6cad`
- **Email**: s1@demo.com
- **Password**: Ts123@123

### 1. 测试 GetAccountSettings

```bash
curl -X GET "http://localhost:8080/admin/api/v1/users/7f3fa6ff-3a07-49b2-a649-6ea086e93863/account-settings" \
  -H "X-User-Id: 7f3fa6ff-3a07-49b2-a649-6ea086e93863" \
  -H "X-Tenant-Id: 095c1ab6-5143-47ea-8670-5476158b6cad"
```

**预期响应**：
```json
{
  "code": 2000,
  "type": "success",
  "message": "ok",
  "result": {
    "id": "7f3fa6ff-3a07-49b2-a649-6ea086e93863",
    "account": "s1",
    "nickname": "s1",
    "email": "s1@demo.com",
    "phone": null,
    "role": "Nurse",
    "save_email": true,
    "save_phone": true
  }
}
```

**验证点**：
- ✅ 包含 `id` 字段
- ✅ 包含 `role` 字段
- ✅ 包含 `save_email` 和 `save_phone` 字段（都是 `true`）

### 2. 测试 UpdateAccountSettings - 更新 phone

```bash
curl -X PUT "http://localhost:8080/admin/api/v1/users/7f3fa6ff-3a07-49b2-a649-6ea086e93863/account-settings" \
  -H "X-User-Id: 7f3fa6ff-3a07-49b2-a649-6ea086e93863" \
  -H "X-Tenant-Id: 095c1ab6-5143-47ea-8670-5476158b6cad" \
  -H "Content-Type: application/json" \
  -d '{"phone": "1234567890", "phone_hash": "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3"}'
```

### 3. 测试 UpdateAccountSettings - 删除 phone 明文但保留 hash

```bash
curl -X PUT "http://localhost:8080/admin/api/v1/users/7f3fa6ff-3a07-49b2-a649-6ea086e93863/account-settings" \
  -H "X-User-Id: 7f3fa6ff-3a07-49b2-a649-6ea086e93863" \
  -H "X-Tenant-Id: 095c1ab6-5143-47ea-8670-5476158b6cad" \
  -H "Content-Type: application/json" \
  -d '{"phone": null}'
```

**验证点**：
- ✅ 数据库中 `phone` 字段为 `NULL`
- ✅ 数据库中 `phone_hash` 字段**仍然有值**（保留用于登录）

### 4. 使用测试脚本

```bash
chmod +x /Users/sady3721/project/owlBack/scripts/test_user_account_settings_simple.sh
/Users/sady3721/project/owlBack/scripts/test_user_account_settings_simple.sh
```

## 当前状态

- ✅ 代码修改完成
- ✅ 路由顺序已修复
- ⚠️ **需要重启后端服务才能测试**

