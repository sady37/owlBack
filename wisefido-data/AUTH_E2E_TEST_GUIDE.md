# Auth 端点端到端测试指南

## 📋 测试目标

验证新的 `AuthHandler` 与前端（owlFront）的集成是否正常工作。

---

## 🚀 启动服务

### 1. 启动 wisefido-data 服务

```bash
cd /Users/sady3721/project/owlBack
docker-compose up -d wisefido-data
```

或者直接运行：

```bash
cd /Users/sady3721/project/owlBack/wisefido-data
go run cmd/wisefido-data/main.go
```

### 2. 确认服务已启动

```bash
curl http://localhost:8080/health
```

应该返回：
```json
{
  "status": "healthy",
  "timestamp": "...",
  "services": {
    "redis": "healthy",
    "database": "healthy"
  }
}
```

---

## 🔍 测试端点

### 1. POST /auth/api/v1/login

#### 1.1 准备测试数据

确保数据库中有测试用户：

```sql
-- 创建测试租户
INSERT INTO tenants (tenant_id, tenant_name, domain, status)
VALUES ('00000000-0000-0000-0000-000000000001', 'System', 'system.local', 'active')
ON CONFLICT (tenant_id) DO NOTHING;

-- 创建测试用户（sysadmin）
-- accountHash = SHA256("sysadmin")
-- passwordHash = SHA256("ChangeMe123!")
INSERT INTO users (tenant_id, user_account, user_account_hash, password_hash, nickname, role, status)
VALUES (
  '00000000-0000-0000-0000-000000000001',
  'sysadmin',
  '\x5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8',
  '\x8f434346648f6b96df89dda901c5176b10a6d83961dd3c1ac88b59b2dc327aa4',
  'SystemAdmin',
  'SystemAdmin',
  'active'
)
ON CONFLICT (tenant_id, user_account) DO UPDATE SET
  user_account_hash = EXCLUDED.user_account_hash,
  password_hash = EXCLUDED.password_hash,
  nickname = EXCLUDED.nickname,
  role = EXCLUDED.role,
  status = EXCLUDED.status;
```

#### 1.2 测试登录

```bash
curl -X POST http://localhost:8080/auth/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "00000000-0000-0000-0000-000000000001",
    "userType": "staff",
    "accountHash": "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8",
    "passwordHash": "8f434346648f6b96df89dda901c5176b10a6d83961dd3c1ac88b59b2dc327aa4"
  }'
```

**预期响应**：
```json
{
  "code": 2000,
  "type": "success",
  "message": "ok",
  "result": {
    "accessToken": "...",
    "refreshToken": "...",
    "userId": "...",
    "user_account": "sysadmin",
    "userType": "staff",
    "role": "SystemAdmin",
    "nickName": "SystemAdmin",
    "tenant_id": "00000000-0000-0000-0000-000000000001",
    "tenant_name": "System",
    "domain": "system.local",
    "homePath": "/monitoring/overview"
  }
}
```

#### 1.3 测试错误场景

**缺少 accountHash**：
```bash
curl -X POST http://localhost:8080/auth/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "00000000-0000-0000-0000-000000000001",
    "userType": "staff",
    "passwordHash": "8f434346648f6b96df89dda901c5176b10a6d83961dd3c1ac88b59b2dc327aa4"
  }'
```

**预期响应**：
```json
{
  "code": -1,
  "type": "error",
  "message": "missing credentials",
  "result": null
}
```

---

### 2. GET /auth/api/v1/institutions/search

#### 2.1 测试搜索机构

```bash
curl "http://localhost:8080/auth/api/v1/institutions/search?accountHash=5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8&passwordHash=8f434346648f6b96df89dda901c5176b10a6d83961dd3c1ac88b59b2dc327aa4&userType=staff"
```

**预期响应**：
```json
{
  "code": 2000,
  "type": "success",
  "message": "ok",
  "result": [
    {
      "id": "00000000-0000-0000-0000-000000000001",
      "name": "System",
      "accountType": "account",
      "domain": "system.local"
    }
  ]
}
```

#### 2.2 测试无匹配

```bash
curl "http://localhost:8080/auth/api/v1/institutions/search?accountHash=0000000000000000000000000000000000000000000000000000000000000000&passwordHash=0000000000000000000000000000000000000000000000000000000000000000&userType=staff"
```

**预期响应**：
```json
{
  "code": 2000,
  "type": "success",
  "message": "ok",
  "result": []
}
```

---

### 3. POST /auth/api/v1/forgot-password/send-code

```bash
curl -X POST http://localhost:8080/auth/api/v1/forgot-password/send-code \
  -H "Content-Type: application/json" \
  -d '{
    "account": "sysadmin",
    "userType": "staff",
    "tenant_id": "00000000-0000-0000-0000-000000000001",
    "tenant_name": "System"
  }'
```

**预期响应**（待实现）：
```json
{
  "code": -1,
  "type": "error",
  "message": "database not available",
  "result": null
}
```

---

### 4. POST /auth/api/v1/forgot-password/verify-code

```bash
curl -X POST http://localhost:8080/auth/api/v1/forgot-password/verify-code \
  -H "Content-Type: application/json" \
  -d '{
    "account": "sysadmin",
    "code": "123456",
    "userType": "staff",
    "tenant_id": "00000000-0000-0000-0000-000000000001",
    "tenant_name": "System"
  }'
```

**预期响应**（待实现）：
```json
{
  "code": -1,
  "type": "error",
  "message": "database not available",
  "result": null
}
```

---

### 5. POST /auth/api/v1/forgot-password/reset

```bash
curl -X POST http://localhost:8080/auth/api/v1/forgot-password/reset \
  -H "Content-Type: application/json" \
  -d '{
    "token": "...",
    "newPassword": "...",
    "userType": "staff"
  }'
```

**预期响应**（待实现）：
```json
{
  "code": -1,
  "type": "error",
  "message": "database not available",
  "result": null
}
```

---

## ✅ 验证清单

### 功能验证

- [ ] POST /auth/api/v1/login - 成功登录
- [ ] POST /auth/api/v1/login - 缺少凭证错误
- [ ] POST /auth/api/v1/login - 无效凭证错误
- [ ] GET /auth/api/v1/institutions/search - 搜索成功
- [ ] GET /auth/api/v1/institutions/search - 无匹配返回空数组
- [ ] POST /auth/api/v1/forgot-password/send-code - 返回错误（待实现）
- [ ] POST /auth/api/v1/forgot-password/verify-code - 返回错误（待实现）
- [ ] POST /auth/api/v1/forgot-password/reset - 返回错误（待实现）

### 响应格式验证

- [ ] 成功响应格式：`{code: 2000, type: "success", message: "ok", result: {...}}`
- [ ] 错误响应格式：`{code: -1, type: "error", message: "...", result: null}`
- [ ] HTTP 状态码：200 OK（错误通过 code=-1 表示）

### 前端集成验证

- [ ] 前端登录功能正常
- [ ] 前端机构搜索功能正常
- [ ] 前端错误提示正常
- [ ] 前端路由跳转正常

---

## 🔍 路由优先级验证

由于新 Handler 的路由注册在 `RegisterStubRoutes` 之前，新 Handler 会优先处理请求。

**验证方法**：
1. 检查日志，确认请求被新 Handler 处理
2. 在 Handler 中添加日志，确认请求到达新 Handler
3. 测试响应格式，确认与新 Handler 一致

---

## 📝 测试结果记录

### 测试日期：__________

### 测试环境：
- 服务地址：`http://localhost:8080`
- 数据库：PostgreSQL
- 测试用户：sysadmin

### 测试结果：

| 端点 | 测试场景 | 结果 | 备注 |
|------|---------|------|------|
| POST /auth/api/v1/login | 成功登录 | ✅/❌ | |
| POST /auth/api/v1/login | 缺少凭证 | ✅/❌ | |
| GET /auth/api/v1/institutions/search | 搜索成功 | ✅/❌ | |
| GET /auth/api/v1/institutions/search | 无匹配 | ✅/❌ | |

### 前端集成测试：

- [ ] 登录页面正常
- [ ] 机构选择页面正常
- [ ] 错误提示正常
- [ ] 路由跳转正常

### 问题记录：

1. 
2. 
3. 

---

## 🎯 确认步骤

完成所有测试后，确认以下事项：

1. ✅ 所有端点响应格式正确
2. ✅ 所有端点 HTTP 状态码正确
3. ✅ 前端集成正常
4. ✅ 日志无异常
5. ✅ 性能无异常

**确认后，可以移除 `RegisterStubRoutes` 中的旧 Auth 路由。**

