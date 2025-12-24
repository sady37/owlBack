# User AccountSettings API 测试指南

## 测试账号信息

- **租户**: Demo
- **用户**: s1
- **User ID**: `7f3fa6ff-3a07-49b2-a649-6ea086e93863`
- **Tenant ID**: `095c1ab6-5143-47ea-8670-5476158b6cad`
- **Email**: s1@demo.com
- **Password**: Ts123@123
- **Role**: Nurse

## 前置条件

1. 启动后端服务：
```bash
cd /Users/sady3721/project/owlBack/wisefido-data
go run cmd/wisefido-data/main.go
```

2. 确认服务运行在 `http://localhost:8080`

## 测试步骤

### 1. 测试 GetAccountSettings

```bash
curl -X GET "http://localhost:8080/admin/api/v1/users/7f3fa6ff-3a07-49b2-a649-6ea086e93863/account-settings" \
  -H "X-User-Id: 7f3fa6ff-3a07-49b2-a649-6ea086e93863" \
  -H "X-Tenant-Id: 095c1ab6-5143-47ea-8670-5476158b6cad"
```

**预期响应**：
```json
{
  "success": true,
  "data": {
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
- ✅ `email` 存在时返回邮箱值
- ✅ `phone` 不存在时返回 `null`

### 2. 测试 UpdateAccountSettings - 更新 phone

```bash
curl -X PUT "http://localhost:8080/admin/api/v1/users/7f3fa6ff-3a07-49b2-a649-6ea086e93863/account-settings" \
  -H "X-User-Id: 7f3fa6ff-3a07-49b2-a649-6ea086e93863" \
  -H "X-Tenant-Id: 095c1ab6-5143-47ea-8670-5476158b6cad" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "1234567890",
    "phone_hash": "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3"
  }'
```

**预期响应**：
```json
{
  "success": true,
  "message": "Account settings updated successfully"
}
```

**验证点**：
- ✅ 更新成功
- ✅ 数据库中 `phone` 字段有值
- ✅ 数据库中 `phone_hash` 字段有值

### 3. 测试 UpdateAccountSettings - 删除 phone 明文但保留 hash

```bash
curl -X PUT "http://localhost:8080/admin/api/v1/users/7f3fa6ff-3a07-49b2-a649-6ea086e93863/account-settings" \
  -H "X-User-Id: 7f3fa6ff-3a07-49b2-a649-6ea086e93863" \
  -H "X-Tenant-Id: 095c1ab6-5143-47ea-8670-5476158b6cad" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": null
  }'
```

**预期响应**：
```json
{
  "success": true,
  "message": "Account settings updated successfully"
}
```

**验证点**：
- ✅ 更新成功
- ✅ 数据库中 `phone` 字段为 `NULL`
- ✅ 数据库中 `phone_hash` 字段**仍然有值**（保留用于登录）

### 4. 验证数据库状态

```sql
SELECT 
  user_id, 
  user_account, 
  email, 
  phone, 
  encode(email_hash, 'hex') as email_hash_hex,
  encode(phone_hash, 'hex') as phone_hash_hex
FROM users 
WHERE user_id::text = '7f3fa6ff-3a07-49b2-a649-6ea086e93863';
```

**预期结果**（删除 phone 明文后）：
- `email`: `s1@demo.com`
- `phone`: `NULL`
- `email_hash_hex`: 有值（64 字符 hex）
- `phone_hash_hex`: **有值**（64 字符 hex，保留用于登录）

## 测试脚本

使用提供的测试脚本：
```bash
cd /Users/sady3721/project/owlBack
./scripts/test_user_account_settings.sh
```

## 注意事项

1. 确保后端服务正在运行
2. 确保数据库连接正常
3. 测试前先查看当前数据状态
4. 测试后验证数据库状态是否符合预期

