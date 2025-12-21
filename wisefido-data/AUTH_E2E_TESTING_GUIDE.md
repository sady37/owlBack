# Auth 端到端测试使用指南

## 📋 概述

本指南提供了完整的 Auth 端点端到端测试流程，包括自动化测试脚本、日志监控和测试报告。

---

## 🚀 快速开始

### 1. 启动服务

```bash
cd /Users/sady3721/project/owlBack
docker-compose up -d wisefido-data
```

### 2. 验证服务状态

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

## 🔍 自动化测试

### 使用测试脚本

**脚本位置**：`scripts/test_auth_endpoints.sh`

**运行测试**：
```bash
cd /Users/sady3721/project/owlBack/wisefido-data
./scripts/test_auth_endpoints.sh
```

**自定义服务地址**：
```bash
BASE_URL=http://localhost:8080 ./scripts/test_auth_endpoints.sh
```

**测试内容**：
- ✅ 服务健康检查
- ✅ POST /auth/api/v1/login - 成功登录
- ✅ POST /auth/api/v1/login - 缺少凭证
- ✅ GET /auth/api/v1/institutions/search - 搜索成功
- ✅ GET /auth/api/v1/institutions/search - 无匹配
- ✅ POST /auth/api/v1/forgot-password/* - 密码重置端点（待实现）

**输出示例**：
```
==========================================
Auth 端点端到端测试
==========================================
服务地址: http://localhost:8080
测试租户: 00000000-0000-0000-0000-000000000001
测试用户: sysadmin
==========================================

=== 检查服务状态 ===
✓ 服务运行正常

=== 测试 POST /auth/api/v1/login ===
账号: sysadmin
账号 Hash: 5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8
密码 Hash: 8f434346648f6b96df89dda901c5176b10a6d83961dd3c1ac88b59b2dc327aa4
HTTP 状态码: 200
响应: {
  "code": 2000,
  "type": "success",
  "message": "ok",
  "result": {
    "accessToken": "...",
    "refreshToken": "...",
    ...
  }
}
✓ 登录成功
✓ 用户账号匹配: sysadmin

==========================================
测试总结
==========================================
总测试数: 6
通过: 6
失败: 0
==========================================
✓ 所有测试通过！
```

---

## 📊 日志监控

### 使用监控脚本

**脚本位置**：`scripts/monitor_auth_logs.sh`

**运行监控**：
```bash
cd /Users/sady3721/project/owlBack/wisefido-data
./scripts/monitor_auth_logs.sh
```

**功能选项**：
1. 实时监控所有日志
2. 监控特定端点
3. 统计错误
4. 统计登录统计

**输出示例**：
```
==========================================
Auth 日志监控工具
==========================================
1. 实时监控所有日志
2. 监控特定端点
3. 统计错误
4. 统计登录统计
5. 退出
==========================================
请选择 (1-5): 1

监控 wisefido-data 容器日志...
按 Ctrl+C 停止监控

[2024-01-01 12:00:00] INFO User login successful user_id=... user_account=sysadmin
[2024-01-01 12:00:01] INFO User login successful user_id=... user_account=sysadmin
```

---

## 📝 手动测试

### 1. 测试登录端点

#### 成功登录

```bash
curl -X POST http://localhost:8080/auth/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "00000000-0000-0000-0000-000000000001",
    "userType": "staff",
    "accountHash": "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8",
    "passwordHash": "8f434346648f6b96df89dda901c5176b10a6d83961dd3c1ac88b59b2dc327aa4"
  }' | jq '.'
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

#### 缺少凭证

```bash
curl -X POST http://localhost:8080/auth/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "00000000-0000-0000-0000-000000000001",
    "userType": "staff"
  }' | jq '.'
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

### 2. 测试搜索机构端点

#### 搜索成功

```bash
curl "http://localhost:8080/auth/api/v1/institutions/search?accountHash=5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8&passwordHash=8f434346648f6b96df89dda901c5176b10a6d83961dd3c1ac88b59b2dc327aa4&userType=staff" | jq '.'
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

#### 无匹配

```bash
curl "http://localhost:8080/auth/api/v1/institutions/search?accountHash=0000000000000000000000000000000000000000000000000000000000000000&passwordHash=0000000000000000000000000000000000000000000000000000000000000000&userType=staff" | jq '.'
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

## 🔍 计算 Hash 值

### 账号 Hash

账号 Hash = SHA256(lowercase(account))

**示例**：
```bash
echo -n "sysadmin" | tr '[:upper:]' '[:lower:]' | sha256sum | cut -d' ' -f1
# 输出: 5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8
```

### 密码 Hash

密码 Hash = SHA256(password)

**示例**：
```bash
echo -n "ChangeMe123!" | sha256sum | cut -d' ' -f1
# 输出: 8f434346648f6b96df89dda901c5176b10a6d83961dd3c1ac88b59b2dc327aa4
```

---

## 📊 测试报告

### 填写测试报告

**报告模板**：`AUTH_E2E_TEST_REPORT.md`

**填写步骤**：
1. 运行自动化测试脚本
2. 记录测试结果
3. 填写测试报告
4. 记录问题和备注

---

## 🔍 监控和日志

### 查看 Docker 日志

```bash
# 实时查看日志
docker-compose logs -f wisefido-data

# 查看最近 100 行日志
docker-compose logs --tail=100 wisefido-data

# 查看包含 "auth" 或 "login" 的日志
docker-compose logs wisefido-data | grep -i "auth\|login"
```

### 统计错误

```bash
# 统计错误数量
docker-compose logs wisefido-data | grep -i "error\|failed" | wc -l

# 查看最近的错误
docker-compose logs wisefido-data | grep -i "error\|failed" | tail -10
```

### 统计登录

```bash
# 统计登录成功
docker-compose logs wisefido-data | grep -i "login successful" | wc -l

# 统计登录失败
docker-compose logs wisefido-data | grep -i "login failed" | wc -l
```

---

## ✅ 验证清单

### 功能验证

- [ ] 所有端点响应格式正确
- [ ] 所有端点 HTTP 状态码正确
- [ ] 错误处理正常
- [ ] 业务逻辑正确

### 前端集成

- [ ] 前端登录功能正常
- [ ] 前端机构选择功能正常
- [ ] 前端错误提示正常
- [ ] 前端路由跳转正常

### 性能

- [ ] 响应时间正常（< 500ms）
- [ ] 无性能问题
- [ ] 无内存泄漏

### 日志

- [ ] 日志记录正常
- [ ] 无异常错误
- [ ] 错误率正常（< 1%）

---

## 🎯 问题排查

### 服务无法启动

```bash
# 检查服务状态
docker-compose ps

# 查看服务日志
docker-compose logs wisefido-data

# 检查数据库连接
docker-compose exec wisefido-data psql -h postgresql -U postgres -d wisefido
```

### 端点返回 404

```bash
# 检查路由注册
docker-compose logs wisefido-data | grep -i "register\|route"

# 检查数据库连接
curl http://localhost:8080/health
```

### 登录失败

```bash
# 检查用户数据
docker-compose exec postgresql psql -U postgres -d wisefido -c \
  "SELECT user_account, user_account_hash, password_hash FROM users WHERE tenant_id = '00000000-0000-0000-0000-000000000001';"

# 检查日志
docker-compose logs wisefido-data | grep -i "login failed"
```

---

## 📝 测试记录

### 测试日期：__________

### 测试环境：
- 服务地址：`http://localhost:8080`
- 数据库：PostgreSQL
- 测试用户：sysadmin

### 测试结果：

| 测试用例 | 状态 | 备注 |
|---------|------|------|
| 自动化测试脚本 | ✅/❌ | |
| 手动测试 - 登录 | ✅/❌ | |
| 手动测试 - 搜索机构 | ✅/❌ | |
| 前端集成测试 | ✅/❌ | |

### 问题记录：

1. 
2. 
3. 

---

## 🎉 完成

完成所有测试后，确认：

1. ✅ 所有端点响应格式正确
2. ✅ 所有端点 HTTP 状态码正确
3. ✅ 前端集成正常
4. ✅ 日志无异常
5. ✅ 性能无异常

**测试完成！**

