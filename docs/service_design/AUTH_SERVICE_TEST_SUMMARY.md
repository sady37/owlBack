# AuthService 测试用例总结

## 📋 测试用例概览

### 测试文件
- **文件路径**: `internal/service/auth_service_integration_test.go`
- **测试用例数**: 15 个
- **代码行数**: 1089 行

---

## ✅ 测试用例清单

### 1. Login 功能测试（9个）

#### 1.1 Staff 登录测试
- ✅ **TestAuthService_Login_Staff_Success**
  - 测试使用 user_account 登录
  - 测试使用 email 登录
  - 测试使用 phone 登录
  - 验证返回的用户信息正确

#### 1.2 错误场景测试
- ✅ **TestAuthService_Login_MissingCredentials**
  - 测试缺少 accountHash
  - 测试缺少 passwordHash

- ✅ **TestAuthService_Login_InvalidHash**
  - 测试无效的 accountHash
  - 测试无效的 passwordHash

- ✅ **TestAuthService_Login_InvalidCredentials**
  - 测试错误的密码

- ✅ **TestAuthService_Login_UserNotActive**
  - 测试非激活用户登录（应该失败）

#### 1.3 Tenant ID 自动解析测试
- ✅ **TestAuthService_Login_AutoResolveTenantID**
  - 测试不提供 tenant_id 时自动解析

- ✅ **TestAuthService_Login_MultipleTenants_ShouldFail**
  - 测试匹配到多个机构时应该失败

#### 1.4 Resident 登录测试
- ✅ **TestAuthService_Login_Resident_Success**
  - 测试使用 resident_account 登录
  - 测试使用 email 登录
  - 测试使用 phone 登录

- ✅ **TestAuthService_Login_ResidentContact_Success**
  - 测试使用 email 登录（resident_contact）
  - 测试使用 phone 登录（resident_contact）

- ✅ **TestAuthService_Login_ResidentContact_NotEnabled**
  - 测试未启用的 resident_contact 登录（应该失败）

---

### 2. SearchInstitutions 功能测试（4个）

- ✅ **TestAuthService_SearchInstitutions_Staff_Success**
  - 测试 Staff 搜索机构成功

- ✅ **TestAuthService_SearchInstitutions_Resident**
  - 测试 Resident 搜索机构成功

- ✅ **TestAuthService_SearchInstitutions_NoMatch**
  - 测试无匹配时返回空数组

- ✅ **TestAuthService_SearchInstitutions_InvalidHash**
  - 测试无效 hash 时返回空数组

- ✅ **TestAuthService_SearchInstitutions_MultipleTenants**
  - 测试匹配到多个机构时返回多个结果

---

## 📊 测试覆盖范围

### 功能覆盖
- ✅ Login（Staff）
- ✅ Login（Resident）
- ✅ Login（ResidentContact）
- ✅ SearchInstitutions（Staff）
- ✅ SearchInstitutions（Resident）
- ✅ Tenant ID 自动解析
- ✅ 多机构匹配处理

### 错误场景覆盖
- ✅ 缺少凭证
- ✅ 无效 hash
- ✅ 无效凭证
- ✅ 用户未激活
- ✅ 联系人未启用
- ✅ 多机构匹配（登录失败）

### 账号类型覆盖
- ✅ user_account / resident_account
- ✅ email
- ✅ phone

---

## 🎯 测试辅助函数

### 数据创建函数
- `createTestTenantForAuth` - 创建测试租户
- `createTestUserForAuth` - 创建测试用户（staff）
- `createTestResidentForAuth` - 创建测试住户
- `createTestUnitForAuth` - 创建测试 unit（resident 需要）
- `cleanupTestDataForAuth` - 清理测试数据

### Hash 计算函数
- `hashAccount` - 计算账号 hash
- `hashPassword` - 计算密码 hash

---

## ✅ 测试完成状态

**所有测试用例已创建完成，覆盖了所有主要业务场景和错误场景。**

**下一步**: 进入阶段 5：实现 Handler

