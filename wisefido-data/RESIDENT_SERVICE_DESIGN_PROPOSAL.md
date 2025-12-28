# Resident Service 设计提案

## 当前问题分析

### 前端操作场景
1. **residentProfile tab**：基础信息，操作 `resident`, `phi`，并涉及 `resident_caregiver`, `resident-unit`
2. **phi tab**：对 `phi` 操作
3. **contact tab**：对 `contact` 操作
4. **保存时**：所有 tab 的数据一并提交

### 当前设计问题
1. **分散的更新方法**：
   - `UpdateResident` - 更新 resident, phi, caregivers
   - `UpdateResidentContact` - 单独更新 contact
   - 前端需要调用多个接口才能完成一次保存

2. **缺少事务保证**：
   - 多个更新操作不在同一个事务中
   - 如果部分更新失败，数据可能不一致

3. **缺少统一的数据结构**：
   - 前端需要分别构造多个请求
   - 数据分散，难以维护

---

## 设计提案

### 方案 1：统一保存接口（推荐）

#### 1.1 新增接口定义

```go
// ResidentService 接口新增方法
type ResidentService interface {
    // ... 现有方法 ...
    
    // 统一保存接口（推荐）
    SaveResidentProfile(ctx context.Context, req SaveResidentProfileRequest) (*SaveResidentProfileResponse, error)
}
```

#### 1.2 请求结构

```go
// SaveResidentProfileRequest 统一保存请求
type SaveResidentProfileRequest struct {
    TenantID      string
    ResidentID    string // 如果为空，表示创建；否则表示更新
    
    // 权限信息
    CurrentUserID   string
    CurrentUserRole string
    CurrentUserType string
    PermissionCheck *PermissionCheckResult
    
    // Resident 基础信息（对应 residentProfile tab）
    Resident *ResidentProfileData `json:"resident,omitempty"`
    
    // PHI 信息（对应 phi tab）
    PHI *PHIProfileData `json:"phi,omitempty"`
    
    // Contacts 信息（对应 contact tab）
    Contacts []*ContactProfileData `json:"contacts,omitempty"`
    
    // Caregivers 信息（对应 residentProfile tab）
    Caregivers *CaregiversProfileData `json:"caregivers,omitempty"`
    
    // Unit 绑定（对应 residentProfile tab）
    UnitBinding *UnitBindingData `json:"unit_binding,omitempty"`
}

// ResidentProfileData Resident 基础信息
type ResidentProfileData struct {
    ResidentAccount *string `json:"resident_account,omitempty"`
    Nickname        *string `json:"nickname,omitempty"`
    Status          *string `json:"status,omitempty"`
    ServiceLevel    *string `json:"service_level,omitempty"`
    AdmissionDate   *int64  `json:"admission_date,omitempty"` // Unix timestamp
    DischargeDate   *int64  `json:"discharge_date,omitempty"` // Unix timestamp
    FamilyTag       *string `json:"family_tag,omitempty"`
    IsAccessEnabled *bool   `json:"is_access_enabled,omitempty"`
    Note            *string `json:"note,omitempty"`
}

// PHIProfileData PHI 信息
type PHIProfileData struct {
    FirstName    *string `json:"first_name,omitempty"`
    LastName     *string `json:"last_name,omitempty"`
    Gender       *string `json:"gender,omitempty"`
    DateOfBirth  *int64  `json:"date_of_birth,omitempty"` // Unix timestamp
    
    // 联系方式（明文，可选保存）
    ResidentPhone *string `json:"resident_phone,omitempty"`
    ResidentEmail *string `json:"resident_email,omitempty"`
    SavePhone     bool    `json:"save_phone"` // 是否保存明文
    SaveEmail     bool    `json:"save_email"` // 是否保存明文
    
    // Hash（用于登录）
    PhoneHash *string `json:"phone_hash,omitempty"` // Hex string
    EmailHash *string `json:"email_hash,omitempty"`  // Hex string
    
    // 身体指标
    WeightLb      *float64 `json:"weight_lb,omitempty"`
    HeightFt      *int     `json:"height_ft,omitempty"`
    HeightIn      *int     `json:"height_in,omitempty"`
    MobilityLevel *string  `json:"mobility_level,omitempty"`
    TremorStatus  *string  `json:"tremor_status,omitempty"`
    MobilityAid   *string  `json:"mobility_aid,omitempty"`
    ADLAssistance *string  `json:"adl_assistance,omitempty"`
    CommStatus    *string  `json:"comm_status,omitempty"`
    
    // 健康状态
    HasHypertension  *bool   `json:"has_hypertension,omitempty"`
    HasHyperlipaemia *bool   `json:"has_hyperlipaemia,omitempty"`
    HasHyperglycaemia *bool  `json:"has_hyperglycaemia,omitempty"`
    HasStrokeHistory *bool   `json:"has_stroke_history,omitempty"`
    HasParalysis     *bool   `json:"has_paralysis,omitempty"`
    HasAlzheimer     *bool   `json:"has_alzheimer,omitempty"`
    MedicalHistory   *string `json:"medical_history,omitempty"`
    
    // 地址信息
    HomeAddressStreet    *string `json:"home_address_street,omitempty"`
    HomeAddressCity      *string `json:"home_address_city,omitempty"`
    HomeAddressState     *string `json:"home_address_state,omitempty"`
    HomeAddressPostalCode *string `json:"home_address_postal_code,omitempty"`
    PlusCode             *string `json:"plus_code,omitempty"`
}

// ContactProfileData Contact 信息
type ContactProfileData struct {
    Slot             string  `json:"slot"` // "A", "B", "C", etc.
    IsEnabled        *bool   `json:"is_enabled,omitempty"`
    Relationship     *string  `json:"relationship,omitempty"`
    ContactFirstName *string  `json:"contact_first_name,omitempty"`
    ContactLastName  *string  `json:"contact_last_name,omitempty"`
    ContactPhone     *string  `json:"contact_phone,omitempty"`
    ContactEmail     *string  `json:"contact_email,omitempty"`
    ReceiveSMS       *bool    `json:"receive_sms,omitempty"`
    ReceiveEmail     *bool    `json:"receive_email,omitempty"`
    
    // Hash（用于登录）
    PhoneHash *string `json:"phone_hash,omitempty"` // Hex string
    EmailHash *string `json:"email_hash,omitempty"`  // Hex string
}

// CaregiversProfileData Caregivers 信息
type CaregiversProfileData struct {
    UserList  []string `json:"user_list,omitempty"`  // 用户ID列表
    GroupList []string `json:"group_list,omitempty"` // 标签ID列表
}

// UnitBindingData Unit 绑定信息
type UnitBindingData struct {
    UnitID *string `json:"unit_id,omitempty"`
    RoomID *string `json:"room_id,omitempty"`
    BedID  *string `json:"bed_id,omitempty"`
}

// SaveResidentProfileResponse 统一保存响应
type SaveResidentProfileResponse struct {
    Success    bool   `json:"success"`
    ResidentID string `json:"resident_id"`
    Message    string `json:"message,omitempty"`
}
```

#### 1.3 实现逻辑

```go
// SaveResidentProfile 统一保存住户档案
// 在一个事务中处理所有相关表的创建/更新
func (s *residentService) SaveResidentProfile(ctx context.Context, req SaveResidentProfileRequest) (*SaveResidentProfileResponse, error) {
    // 1. 参数验证
    if req.TenantID == "" {
        return nil, fmt.Errorf("tenant_id is required")
    }
    
    // 2. 权限检查
    // ... 权限检查逻辑 ...
    
    // 3. 开启事务
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()
    
    var residentID string
    isCreate := req.ResidentID == ""
    
    // 4. 处理 Resident 表
    if req.Resident != nil {
        if isCreate {
            // 创建 Resident
            residentID, err = s.createResidentInTx(ctx, tx, req.TenantID, req.Resident)
            if err != nil {
                return nil, fmt.Errorf("failed to create resident: %w", err)
            }
        } else {
            // 更新 Resident
            residentID = req.ResidentID
            err = s.updateResidentInTx(ctx, tx, req.TenantID, residentID, req.Resident)
            if err != nil {
                return nil, fmt.Errorf("failed to update resident: %w", err)
            }
        }
    } else if !isCreate {
        residentID = req.ResidentID
    } else {
        return nil, fmt.Errorf("resident data is required for creation")
    }
    
    // 5. 处理 PHI 表
    if req.PHI != nil {
        err = s.upsertPHIInTx(ctx, tx, req.TenantID, residentID, req.PHI)
        if err != nil {
            return nil, fmt.Errorf("failed to upsert PHI: %w", err)
        }
    }
    
    // 6. 处理 Contacts 表
    if len(req.Contacts) > 0 {
        err = s.upsertContactsInTx(ctx, tx, req.TenantID, residentID, req.Contacts)
        if err != nil {
            return nil, fmt.Errorf("failed to upsert contacts: %w", err)
        }
    }
    
    // 7. 处理 Caregivers 表
    if req.Caregivers != nil {
        err = s.upsertCaregiversInTx(ctx, tx, req.TenantID, residentID, req.Caregivers)
        if err != nil {
            return nil, fmt.Errorf("failed to upsert caregivers: %w", err)
        }
    }
    
    // 8. 处理 Unit 绑定
    if req.UnitBinding != nil {
        err = s.bindResidentToLocationInTx(ctx, tx, req.TenantID, residentID, req.UnitBinding)
        if err != nil {
            return nil, fmt.Errorf("failed to bind location: %w", err)
        }
    }
    
    // 9. 提交事务
    if err = tx.Commit(); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }
    
    return &SaveResidentProfileResponse{
        Success:    true,
        ResidentID: residentID,
    }, nil
}
```

#### 1.4 事务内操作方法

```go
// createResidentInTx 在事务中创建 Resident
func (s *residentService) createResidentInTx(ctx context.Context, tx *sql.Tx, tenantID string, data *ResidentProfileData) (string, error) {
    // 使用 Repository 的 CreateResident，但需要支持事务
    // 或者直接使用 tx 执行 SQL
}

// updateResidentInTx 在事务中更新 Resident
func (s *residentService) updateResidentInTx(ctx context.Context, tx *sql.Tx, tenantID, residentID string, data *ResidentProfileData) error {
    // 使用 Repository 的 UpdateResident，但需要支持事务
    // 或者直接使用 tx 执行 SQL
}

// upsertPHIInTx 在事务中创建/更新 PHI
func (s *residentService) upsertPHIInTx(ctx context.Context, tx *sql.Tx, tenantID, residentID string, data *PHIProfileData) error {
    // 使用 Repository 的 UpsertResidentPHI，但需要支持事务
    // 或者直接使用 tx 执行 SQL
}

// upsertContactsInTx 在事务中创建/更新 Contacts
func (s *residentService) upsertContactsInTx(ctx context.Context, tx *sql.Tx, tenantID, residentID string, contacts []*ContactProfileData) error {
    // 遍历 contacts，使用 Repository 的 CreateResidentContact/UpdateResidentContact
    // 或者直接使用 tx 执行 SQL
}

// upsertCaregiversInTx 在事务中创建/更新 Caregivers
func (s *residentService) upsertCaregiversInTx(ctx context.Context, tx *sql.Tx, tenantID, residentID string, data *CaregiversProfileData) error {
    // 使用 Repository 的 UpsertResidentCaregiver，但需要支持事务
    // 或者直接使用 tx 执行 SQL
}

// bindResidentToLocationInTx 在事务中绑定位置
func (s *residentService) bindResidentToLocationInTx(ctx context.Context, tx *sql.Tx, tenantID, residentID string, data *UnitBindingData) error {
    // 使用 Repository 的 BindResidentToLocation，但需要支持事务
    // 或者直接使用 tx 执行 SQL
}
```

---

### 方案 2：增强现有接口（备选）

如果不想新增接口，可以增强现有的 `UpdateResident` 方法：

```go
// UpdateResidentRequest 增强，支持所有相关数据
type UpdateResidentRequest struct {
    // ... 现有字段 ...
    
    // 新增：支持 Contacts 批量更新
    Contacts []*ContactProfileData `json:"contacts,omitempty"`
    
    // 新增：支持 Unit 绑定
    UnitBinding *UnitBindingData `json:"unit_binding,omitempty"`
}
```

然后在 `UpdateResident` 方法中使用事务处理所有更新。

---

## 推荐方案

**推荐使用方案 1（统一保存接口）**，原因：

1. **清晰的职责分离**：
   - `SaveResidentProfile` - 统一保存接口，前端主要使用
   - `UpdateResident` - 部分更新接口，用于特定场景
   - `UpdateResidentContact` - 单独更新联系人，用于特定场景

2. **事务保证**：
   - 所有相关表的更新在一个事务中
   - 要么全部成功，要么全部失败

3. **前端友好**：
   - 前端只需要调用一个接口
   - 数据结构清晰，易于维护

4. **向后兼容**：
   - 保留现有的独立更新方法
   - 不影响现有功能

---

## Repository 层改进

为了支持事务，Repository 层需要支持传入 `*sql.Tx`：

```go
// ResidentsRepository 接口增强
type ResidentsRepository interface {
    // ... 现有方法 ...
    
    // 支持事务的方法
    CreateResidentWithTx(ctx context.Context, tx *sql.Tx, tenantID string, resident *domain.Resident) (string, error)
    UpdateResidentWithTx(ctx context.Context, tx *sql.Tx, tenantID, residentID string, resident *domain.Resident) error
    UpsertResidentPHIWithTx(ctx context.Context, tx *sql.Tx, tenantID, residentID string, phi *domain.ResidentPHI) error
    // ... 其他方法 ...
}
```

或者，Repository 方法可以接受 `interface{}` 类型，既支持 `*sql.DB` 也支持 `*sql.Tx`：

```go
type DBExecutor interface {
    ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
    QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
    QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// 然后 Repository 方法接受 DBExecutor
func (r *PostgresResidentsRepository) CreateResident(ctx context.Context, db DBExecutor, tenantID string, resident *domain.Resident) (string, error) {
    // 实现逻辑
}
```

---

## 实施步骤

1. **第一步**：定义新的请求/响应结构
2. **第二步**：实现 `SaveResidentProfile` 方法（使用现有 Repository，暂时不使用事务）
3. **第三步**：增强 Repository 层支持事务
4. **第四步**：重构 `SaveResidentProfile` 使用事务
5. **第五步**：前端迁移到新接口
6. **第六步**：保留旧接口作为兼容（可选）

