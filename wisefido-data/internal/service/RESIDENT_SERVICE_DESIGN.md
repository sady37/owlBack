# ResidentService 接口设计文档

## 概述

ResidentService 是住户管理服务，负责处理住户的 CRUD 操作、权限检查、业务规则验证和数据转换。

**复杂度**：极高（3032 行 Handler 需要重构）

---

## 1. 接口方法设计

### 1.1 ListResidents - 查询住户列表

**功能**：
- 支持多种过滤条件（status, service_level, search）
- 支持权限过滤（AssignedOnly, BranchOnly）
- 支持 Resident/Family 登录（只能查看自己）
- 支持分页

**请求**：
```go
type ListResidentsRequest struct {
    TenantID      string // 租户ID
    CurrentUserID string // 当前用户ID
    CurrentUserType string // 当前用户类型：'resident' | 'family' | 'staff'
    CurrentUserRole string // 当前用户角色（仅 staff 需要）
    
    // 过滤条件
    Search        string // 搜索关键词（nickname, unit_name）
    Status        string // 状态过滤
    ServiceLevel  string // 护理级别过滤
    
    // 分页
    Page     int // 页码，默认 1
    PageSize int // 每页数量，默认 20
}
```

**响应**：
```go
type ListResidentsResponse struct {
    Items      []*ResidentDTO // 住户列表
    Pagination PaginationDTO   // 分页信息
}

type ResidentDTO struct {
    ResidentID      string
    TenantID        string
    ResidentAccount *string
    Nickname        string
    Status          string
    ServiceLevel    *string
    AdmissionDate   *int64 // Unix timestamp
    DischargeDate   *int64 // Unix timestamp
    FamilyTag       *string
    UnitID          *string
    UnitName        *string
    BranchTag       *string
    AreaTag         *string
    UnitNumber      *string
    IsMultiPersonRoom bool
    RoomID          *string
    RoomName        *string
    BedID           *string
    BedName         *string
    IsAccessEnabled bool
}
```

---

### 1.2 GetResident - 获取住户详情

**功能**：
- 支持通过 resident_id 或 contact_id 查询
- 支持 include_phi 和 include_contacts 参数
- 权限检查（Resident/Family 只能查看自己，Staff 根据权限过滤）

**请求**：
```go
type GetResidentRequest struct {
    TenantID        string // 租户ID
    ResidentID      string // 住户ID（或 contact_id）
    CurrentUserID   string // 当前用户ID
    CurrentUserType string // 当前用户类型
    CurrentUserRole string // 当前用户角色（仅 staff 需要）
    
    // 可选数据
    IncludePHI      bool // 是否包含 PHI 数据
    IncludeContacts bool // 是否包含联系人数据
}
```

**响应**：
```go
type GetResidentResponse struct {
    Resident  *ResidentDTO  // 住户基本信息
    PHI       *ResidentPHIDTO // PHI 数据（如果 IncludePHI=true）
    Contacts  []*ResidentContactDTO // 联系人列表（如果 IncludeContacts=true）
}
```

---

### 1.3 CreateResident - 创建住户

**功能**：
- 创建住户记录
- 可选创建 PHI 记录
- 可选创建联系人记录
- 权限检查（需要 C 权限）
- 业务规则验证（resident_account 必填、discharge_date 验证等）

**请求**：
```go
type CreateResidentRequest struct {
    TenantID        string // 租户ID
    CurrentUserID   string // 当前用户ID
    CurrentUserRole string // 当前用户角色
    
    // 必填字段
    ResidentAccount string // 住户账号（必填）
    Nickname        string // 昵称（必填）
    
    // 可选字段
    Password        string // 密码（默认 "ChangeMe123!"）
    Status          string // 状态（默认 "active"）
    ServiceLevel    string // 护理级别
    AdmissionDate   *int64 // 入院日期（Unix timestamp）
    UnitID          string // 单元ID
    FamilyTag       string // 家庭标签
    IsAccessEnabled bool  // 是否允许查看状态
    Note            string // 备注
    
    // Hash 字段（前端计算）
    PhoneHash       string // phone_hash (hex)
    EmailHash       string // email_hash (hex)
    
    // PHI 数据（可选）
    PHI             *CreateResidentPHIRequest
    
    // 联系人数据（可选）
    Contacts        []*CreateResidentContactRequest
}

type CreateResidentPHIRequest struct {
    FirstName       string
    LastName        string
    Gender          string
    DateOfBirth     *int64
    ResidentPhone   string // 明文（可选保存）
    ResidentEmail   string // 明文（可选保存）
    SavePhone       bool   // 是否保存明文 phone
    SaveEmail       bool   // 是否保存明文 email
    // ... 其他 PHI 字段
}

type CreateResidentContactRequest struct {
    Slot            string // 'A', 'B', 'C', 'D', 'E'
    IsEnabled       bool
    Relationship    string
    ContactFirstName string
    ContactLastName  string
    ContactPhone     string
    ContactEmail     string
    PhoneHash        string // phone_hash (hex)
    EmailHash        string // email_hash (hex)
    ReceiveSMS       bool
    ReceiveEmail     bool
    ContactFamilyTag string
}
```

**响应**：
```go
type CreateResidentResponse struct {
    ResidentID string // 创建的住户ID
}
```

---

### 1.4 UpdateResident - 更新住户

**功能**：
- 支持部分更新
- 更新住户基本信息
- 可选更新 PHI 数据
- 可选更新联系人数据
- 可选更新 Caregivers 数据
- 权限检查（需要 U 权限）
- 业务规则验证（discharge_date 验证等）

**请求**：
```go
type UpdateResidentRequest struct {
    TenantID        string // 租户ID
    ResidentID      string // 住户ID
    CurrentUserID   string // 当前用户ID
    CurrentUserType string // 当前用户类型
    CurrentUserRole string // 当前用户角色
    
    // 可更新字段（使用指针表示可选）
    Nickname        *string
    Status          *string
    ServiceLevel    *string
    AdmissionDate   *int64
    DischargeDate   *int64
    UnitID          *string
    FamilyTag       *string
    IsAccessEnabled *bool
    Note            *string
    
    // PHI 更新（可选）
    PHI             *UpdateResidentPHIRequest
    
    // Caregivers 更新（可选）
    Caregivers      *UpdateResidentCaregiversRequest
}

type UpdateResidentPHIRequest struct {
    // 所有 PHI 字段（使用指针表示可选）
    FirstName       *string
    LastName        *string
    // ... 其他字段
}

type UpdateResidentCaregiversRequest struct {
    UserList  []string // 用户ID列表
    GroupList []string // 标签ID列表
}
```

**响应**：
```go
type UpdateResidentResponse struct {
    Success bool
}
```

---

### 1.5 DeleteResident - 删除住户（软删除）

**功能**：
- 软删除：将 status 设置为 'discharged'
- 权限检查（需要 D 权限）
- Resident/Family 不能删除

**请求**：
```go
type DeleteResidentRequest struct {
    TenantID        string // 租户ID
    ResidentID      string // 住户ID
    CurrentUserID   string // 当前用户ID
    CurrentUserType string // 当前用户类型
    CurrentUserRole string // 当前用户角色
}
```

**响应**：
```go
type DeleteResidentResponse struct {
    Success bool
}
```

---

### 1.6 ResetResidentPassword - 重置住户密码

**功能**：
- 重置住户密码
- 权限检查（Resident 只能重置自己的密码，Staff 需要 U 权限）

**请求**：
```go
type ResetResidentPasswordRequest struct {
    TenantID        string // 租户ID
    ResidentID      string // 住户ID
    CurrentUserID   string // 当前用户ID
    CurrentUserType string // 当前用户类型
    CurrentUserRole string // 当前用户角色
    NewPassword     string // 新密码（可选，默认生成）
}
```

**响应**：
```go
type ResetResidentPasswordResponse struct {
    Success      bool
    NewPassword  string // 生成的新密码
}
```

---

### 1.7 ResetContactPassword - 重置联系人密码

**功能**：
- 重置联系人密码
- 权限检查（Contact 只能重置自己的密码，Resident 只能重置自己的联系人密码，Staff 需要 U 权限）

**请求**：
```go
type ResetContactPasswordRequest struct {
    TenantID        string // 租户ID
    ContactID       string // 联系人ID
    CurrentUserID   string // 当前用户ID
    CurrentUserType string // 当前用户类型
    CurrentUserRole string // 当前用户角色
    NewPassword     string // 新密码（可选，默认生成）
}
```

**响应**：
```go
type ResetContactPasswordResponse struct {
    Success      bool
    NewPassword  string // 生成的新密码
}
```

---

## 2. 权限检查逻辑

### 2.1 权限检查函数

Service 层需要调用权限检查函数，但 `GetResourcePermission` 在 `httpapi` 包中。有两种方案：

**方案 A：在 Service 层实现权限检查**
- 优点：Service 层独立，不依赖 HTTP 层
- 缺点：需要复制权限检查逻辑

**方案 B：将权限检查逻辑提取到独立包**
- 优点：代码复用，逻辑统一
- 缺点：需要重构现有代码

**建议**：方案 B，创建 `internal/permission` 包，将权限检查逻辑提取出来。

### 2.2 权限检查规则

#### ListResidents
- **Resident/Family 登录**：只能查看自己（或关联的住户）
- **Staff 登录**：
  - 检查 R 权限
  - 如果 `AssignedOnly=true`：只返回分配给该用户的住户
  - 如果 `BranchOnly=true`：只返回该分支的住户

#### GetResident
- **Resident/Family 登录**：只能查看自己
- **Staff 登录**：
  - 检查 R 权限
  - 如果 `AssignedOnly=true`：检查是否分配给该用户
  - 如果 `BranchOnly=true`：检查分支匹配

#### CreateResident
- **Resident/Family 登录**：不允许创建
- **Staff 登录**：
  - 检查 C 权限
  - 如果 `BranchOnly=true`：检查 unit 的 branch_tag 匹配

#### UpdateResident
- **Resident 登录**：只能更新自己
- **Family 登录**：只能更新关联的住户
- **Staff 登录**：
  - 检查 U 权限
  - 如果 `AssignedOnly=true`：检查是否分配给该用户
  - 如果 `BranchOnly=true`：检查分支匹配

#### DeleteResident
- **Resident/Family 登录**：不允许删除
- **Staff 登录**：
  - 检查 D 权限
  - 如果 `AssignedOnly=true`：检查是否分配给该用户
  - 如果 `BranchOnly=true`：检查分支匹配

---

## 3. 业务规则验证

### 3.1 创建住户规则

1. **resident_account 必填**：每家机构有自己的编码模式
2. **nickname 必填**
3. **admission_date 默认值**：如果未提供，使用当前日期
4. **status 默认值**：如果未提供，使用 "active"
5. **discharge_date 验证**：仅在 status='discharged' 或 'transferred' 时可以有值
6. **unit_id 验证**：如果提供了 unit_id，需要验证：
   - unit 存在
   - 如果 `BranchOnly=true`，检查 branch_tag 匹配
7. **phone_hash/email_hash**：如果提供了 phone/email，必须提供对应的 hash
8. **PHI 数据**：如果提供了 first_name/last_name，自动创建 PHI 记录
9. **联系人数据**：如果提供了 contacts，创建联系人记录

### 3.2 更新住户规则

1. **discharge_date 验证**：仅在 status='discharged' 或 'transferred' 时允许设置
2. **unit_id 验证**：如果更新了 unit_id，需要验证 unit 存在和权限
3. **部分更新**：只更新提供的字段

### 3.3 删除住户规则

1. **软删除**：将 status 设置为 'discharged'
2. **不允许硬删除**：保持数据完整性

---

## 4. 数据转换

### 4.1 前端格式 → 领域模型

- `map[string]any` → `domain.Resident`
- `map[string]any` → `domain.ResidentPHI`
- `map[string]any` → `domain.ResidentContact`
- 时间格式：`"2006-01-02"` → `time.Time`
- Hash 格式：hex string → `[]byte`

### 4.2 领域模型 → 前端格式

- `domain.Resident` → `ResidentDTO`
- `domain.ResidentPHI` → `ResidentPHIDTO`
- `domain.ResidentContact` → `ResidentContactDTO`
- 时间格式：`time.Time` → Unix timestamp (int64)
- 关联数据：JOIN units, rooms, beds 表获取完整信息

---

## 5. 业务编排

### 5.1 CreateResident 编排

1. 权限检查
2. 业务规则验证
3. 创建 Resident 记录
4. 如果提供了 PHI 数据，创建 ResidentPHI 记录
5. 如果提供了 Contacts 数据，创建 ResidentContact 记录
6. 返回 ResidentID

### 5.2 UpdateResident 编排

1. 权限检查
2. 业务规则验证
3. 更新 Resident 记录
4. 如果提供了 PHI 数据，更新/创建 ResidentPHI 记录
5. 如果提供了 Caregivers 数据，更新 ResidentCaregivers 记录
6. 返回成功

---

## 6. 待确认问题

### 6.1 权限检查实现位置

- **方案 A**：在 Service 层实现权限检查（需要访问 role_permissions 表）
- **方案 B**：在 Handler 层进行权限检查，Service 层只处理业务逻辑

**建议**：方案 B，因为权限检查逻辑复杂，且需要访问 users 表和 role_permissions 表，更适合在 Handler 层处理。

### 6.2 权限检查函数复用

- **问题**：`GetResourcePermission` 在 `httpapi` 包中，Service 层无法直接使用
- **方案 A**：将权限检查逻辑提取到 `internal/permission` 包
- **方案 B**：Service 层接收已检查的权限信息（从 Handler 传入）

**建议**：方案 B，Service 层接收 `PermissionCheck` 结构，由 Handler 层负责权限检查。

### 6.3 数据转换位置

- **方案 A**：在 Service 层进行数据转换（map → domain → DTO）
- **方案 B**：在 Handler 层进行数据转换（map → DTO），Service 层只处理 domain 模型

**建议**：方案 A，Service 层负责业务逻辑和数据转换，Handler 层只负责 HTTP 处理。

---

## 7. 实现计划

### 阶段 1：分析需求 ✅
- [x] 查看旧 Handler 实现
- [x] 了解业务规则和权限检查
- [x] 确认 Repository 接口

### 阶段 2：设计接口
- [ ] 确认接口方法
- [ ] 设计请求/响应 DTO
- [ ] 确认权限检查逻辑

### 阶段 3：实现 Service
- [ ] 实现权限检查辅助函数
- [ ] 实现业务规则验证
- [ ] 实现数据转换
- [ ] 实现业务编排

### 阶段 4：编写测试
- [ ] 编写单元测试
- [ ] 编写集成测试

### 阶段 5：实现 Handler
- [ ] 实现 HTTP 层处理
- [ ] 权限检查调用
- [ ] 响应格式对齐

### 阶段 6：集成和路由注册
- [ ] 在 main.go 中注册
- [ ] 在 router.go 中注册路由

### 阶段 7：验证和测试
- [ ] 对比旧 Handler 响应格式
- [ ] 端到端测试

