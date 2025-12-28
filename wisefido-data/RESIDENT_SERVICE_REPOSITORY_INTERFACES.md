# resident_service.go 与 Repository 接口调用关系

## Repository 接口定义

**接口类型**：`repository.ResidentsRepository`

**文件位置**：`internal/repository/residents_repo.go`

---

## 1. Residents 表操作接口

### 1.1 GetResident
**接口定义**：
```go
GetResident(ctx context.Context, tenantID, residentID string) (*domain.Resident, error)
```

**在 resident_service.go 中的调用**：
- **第 1976 行**：`UpdateResident` 方法中，获取现有住户信息用于验证
- **第 2077 行**：`UpdateResident` 方法中，更新后重新获取住户信息

**用途**：
- 获取单个住户的完整信息
- 用于更新前的数据验证
- 用于更新后的数据返回

---

### 1.2 CreateResident
**接口定义**：
```go
CreateResident(ctx context.Context, tenantID string, resident *domain.Resident) (string, error)
```

**在 resident_service.go 中的调用**：
- **第 1356 行**：`CreateResident` 方法中，创建新住户

**用途**：
- 创建新住户记录
- 返回新创建的 `resident_id`

---

### 1.3 UpdateResident
**接口定义**：
```go
UpdateResident(ctx context.Context, tenantID, residentID string, resident *domain.Resident) error
```

**在 resident_service.go 中的调用**：
- **第 2065 行**：`UpdateResident` 方法中，更新住户基本信息
- **第 2195 行**：`UpdateResident` 方法中，更新密码哈希
- **第 2417 行**：`DeleteResident` 方法中，软删除住户（更新 status）

**用途**：
- 更新住户基本信息
- 更新密码哈希
- 软删除住户（通过更新 status）

---

### 1.4 GetResidentContacts
**接口定义**：
```go
GetResidentContacts(ctx context.Context, tenantID, residentID string) ([]*domain.ResidentContact, error)
```

**在 resident_service.go 中的调用**：
- **第 1099 行**：`GetResident` 方法中，获取住户的联系人列表

**用途**：
- 获取住户的所有联系人信息
- 用于返回完整的住户信息

---

### 1.5 CreateResidentContact
**接口定义**：
```go
CreateResidentContact(ctx context.Context, tenantID, residentID string, contact *domain.ResidentContact) (string, error)
```

**在 resident_service.go 中的调用**：
- **第 1495 行**：`CreateResident` 方法中，创建住户时同时创建联系人
- **第 3221 行**：`UpdateResidentContact` 方法中，如果联系人不存在则创建新联系人

**用途**：
- 创建新的住户联系人
- 返回新创建的 `contact_id`（实际是 slot 值）

---

### 1.6 UpdateResidentContact
**接口定义**：
```go
UpdateResidentContact(ctx context.Context, tenantID, contactID string, contact *domain.ResidentContact) error
```

**在 resident_service.go 中的调用**：
- **第 3347 行**：`UpdateResidentContact` 方法中，更新联系人信息

**用途**：
- 更新现有联系人的信息
- 注意：`contactID` 参数实际上是 `slot` 值（INTEGER）

---

## 2. ResidentPHI 表操作接口

### 2.1 GetResidentPHI
**接口定义**：
```go
GetResidentPHI(ctx context.Context, tenantID, residentID string) (*domain.ResidentPHI, error)
```

**在 resident_service.go 中的调用**：
- **第 1062 行**：`GetResident` 方法中，获取住户的 PHI（受保护健康信息）
- **第 2211 行**：`UpdateResident` 方法中，获取现有 PHI 用于合并更新

**用途**：
- 获取住户的受保护健康信息（PHI）
- 用于返回完整的住户信息
- 用于更新前的数据合并

---

### 2.2 UpsertResidentPHI
**接口定义**：
```go
UpsertResidentPHI(ctx context.Context, tenantID, residentID string, phi *domain.ResidentPHI) error
```

**在 resident_service.go 中的调用**：
- **第 1417 行**：`CreateResident` 方法中，创建住户时同时创建/更新 PHI
- **第 2306 行**：`UpdateResident` 方法中，更新住户的 PHI 信息

**用途**：
- 创建或更新住户的 PHI 信息
- 使用 UPSERT 语义（如果存在则更新，不存在则创建）

---

## 3. ResidentCaregivers 表操作接口

### 3.1 UpsertResidentCaregiver
**接口定义**：
```go
UpsertResidentCaregiver(ctx context.Context, tenantID, residentID string, caregiver *domain.ResidentCaregiver) error
```

**在 resident_service.go 中的调用**：
- **第 2331 行**：`UpdateResident` 方法中，更新住户的护理人员关联配置

**用途**：
- 创建或更新住户的护理人员关联配置
- 使用 UPSERT 语义（如果存在则更新，不存在则创建）
- 注意：`resident_caregivers` 表有 `UNIQUE(tenant_id, resident_id)` 约束

---

## 接口调用统计

### 按接口类型分类：

1. **Residents 表操作**：
   - `GetResident`: 2 次调用
   - `CreateResident`: 1 次调用
   - `UpdateResident`: 3 次调用
   - `GetResidentContacts`: 1 次调用
   - `CreateResidentContact`: 2 次调用
   - `UpdateResidentContact`: 1 次调用

2. **ResidentPHI 表操作**：
   - `GetResidentPHI`: 2 次调用
   - `UpsertResidentPHI`: 2 次调用

3. **ResidentCaregivers 表操作**：
   - `UpsertResidentCaregiver`: 1 次调用

### 未使用的接口：

以下接口在 `residents_repo.go` 中定义，但在 `resident_service.go` 中**未使用**：

1. `GetResidentByAccount` - 通过账户哈希查询住户
2. `GetResidentByEmail` - 通过邮箱哈希查询住户
3. `GetResidentByPhone` - 通过手机哈希查询住户
4. `ListResidents` - 查询住户列表（Service 层直接使用 SQL 查询）
5. `DeleteResident` - 删除住户（Service 层使用 `UpdateResident` 进行软删除）
6. `BindResidentToLocation` - 绑定住户到位置（可能在 Service 层直接使用 SQL）
7. `GetResidentCaregivers` - 获取住户的护理人员关联（可能在 Service 层直接使用 SQL）
8. `DeleteResidentContact` - 删除联系人（可能在 Service 层直接使用 SQL）

---

## 注意事项

1. **Service 层直接使用 SQL 查询**：
   - `ListResidents` 方法在 Service 层直接使用 `s.db.QueryContext` 进行复杂查询（包含 JOIN、权限过滤等）
   - 没有使用 Repository 的 `ListResidents` 接口

2. **软删除**：
   - `DeleteResident` 方法使用 `UpdateResident` 进行软删除（更新 `status` 字段）
   - 没有使用 Repository 的 `DeleteResident` 接口

3. **权限检查**：
   - Service 层在调用 Repository 之前进行权限检查
   - 权限检查使用 `s.db.QueryRowContext` 直接查询数据库

4. **事务处理**：
   - Service 层负责事务管理
   - Repository 层不处理事务，只负责单表操作

