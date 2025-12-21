# 自底向上架构设计

## 设计原则

**从底层开始，逐步向上设计**：
1. **领域模型（Domain Models）** - 基于数据库表结构
2. **Repository 接口** - 数据访问抽象
3. **Repository 实现** - 数据访问 + 数据一致性（替代触发器）
4. **Service 接口** - 业务逻辑抽象
5. **Service 实现** - 业务逻辑 + 权限检查
6. **Handler** - HTTP 处理

---

## 一、领域模型设计（Domain Models）

### 1.1 核心实体识别

基于数据库表结构，识别核心实体：

```
核心实体：
├── Tenant（租户）
├── User（用户）
├── Resident（住户）
│   ├── ResidentPHI（住户 PHI）
│   ├── ResidentContact（住户联系人）
│   └── ResidentCaregiver（护工分配）
├── Location（地址层级）
│   ├── Building（楼栋）
│   ├── Unit（单元）
│   ├── Room（房间）
│   └── Bed（床位）
├── Device（设备）
│   └── DeviceStore（设备库存）
├── Tag（标签）
│   └── TagsCatalog（标签目录）
├── Role（角色）
│   └── RolePermission（角色权限）
└── Card（卡片）
```

### 1.2 领域模型定义

#### 1.2.1 Resident（住户）

```go
// internal/domain/resident.go

package domain

import (
    "database/sql"
    "time"
)

// Resident 住户领域模型
type Resident struct {
    // 标识
    ResidentID      string
    TenantID        string
    
    // 账号信息
    ResidentAccount string
    ResidentAccountHash []byte
    
    // 基本信息
    Nickname        string
    Status          ResidentStatus  // active, discharged, transferred
    ServiceLevel    sql.NullString
    Role            string  // 固定为 'Resident'，用于前端路由识别和权限控制
    
    // 日期
    AdmissionDate  time.Time
    DischargeDate  sql.NullTime
    
    // 位置关联
    UnitID          string
    RoomID          sql.NullString
    BedID           sql.NullString
    
    // Tag 关联
    FamilyTag       sql.NullString
    
    // 权限
    CanViewStatus   bool
    
    // 联系方式（Hash）
    PhoneHash       []byte
    EmailHash       []byte
    PasswordHash    []byte
    
    // 元数据
    Note            sql.NullString
    Metadata        sql.NullString
    
    // 注意：DB 表中没有 created_at 和 updated_at 字段
}

type ResidentStatus string

const (
    ResidentStatusActive      ResidentStatus = "active"
    ResidentStatusDischarged  ResidentStatus = "discharged"
    ResidentStatusTransferred ResidentStatus = "transferred"
)

// ResidentPHI 住户 PHI（个人健康信息）
// 注意：HIS 系统相关字段已移除（HIS_resident_name, HIS_resident_admission_date, HIS_resident_discharge_date, HIS_resident_metadata）
type ResidentPHI struct {
    PHIID           string
    TenantID        string
    ResidentID      string
    
    // 基本信息（PII，加密存储）
    FirstName       sql.NullString
    LastName        sql.NullString
    Gender          sql.NullString
    DateOfBirth     sql.NullTime
    
    // 联系方式（PII）
    ResidentPhone   sql.NullString
    ResidentEmail   sql.NullString
    
    // 健康信息
    WeightLb        sql.NullFloat64
    HeightFt        sql.NullFloat64
    HeightIn        sql.NullFloat64
    MobilityLevel   sql.NullInt64
    TremorStatus    sql.NullString
    MobilityAid     sql.NullString
    ADLAssistance   sql.NullString
    CommStatus      sql.NullString
    
    // 疾病史
    HasHypertension     sql.NullBool
    HasHyperlipaemia    sql.NullBool
    HasHyperglycaemia   sql.NullBool
    HasStrokeHistory    sql.NullBool
    HasParalysis        sql.NullBool
    HasAlzheimer        sql.NullBool
    MedicalHistory      sql.NullString
    
    // 家庭地址（PII，仅用于 Home 场景）
    HomeAddressStreet    sql.NullString
    HomeAddressCity      sql.NullString
    HomeAddressState     sql.NullString
    HomeAddressPostalCode sql.NullString
    PlusCode             sql.NullString
}

// ResidentContact 住户联系人
type ResidentContact struct {
    ContactID           string
    TenantID            string
    ResidentID          string
    Slot                string  // A/B/C/D/E
    
    // 状态
    IsEnabled           bool
    IsEmergencyContact  bool
    
    // 关系
    Relationship        sql.NullString
    Role                string  // 固定为 'Family'
    
    // 联系方式
    ContactFirstName    sql.NullString
    ContactLastName     sql.NullString
    ContactPhone        sql.NullString
    ContactEmail        sql.NullString
    
    // 通知设置
    ReceiveSMS          bool
    ReceiveEmail        bool
    
    // 密码（Hash）
    PhoneHash           []byte
    EmailHash           []byte
    PasswordHash        []byte
    
    // 告警时间窗口
    AlertTimeWindow     sql.NullString  // JSONB
}

// ResidentCaregiver 护工分配
type ResidentCaregiver struct {
    CaregiverID     string  // DB 字段名是 `caregiver_id`
    TenantID        string
    ResidentID      string
    
    // 用户组和用户列表（JSONB）
    GroupList       sql.NullString  // JSONB: ["tag1", "tag2"]
    UserList        sql.NullString  // JSONB: ["user_id1", "user_id2"]
}
```

#### 1.2.2 User（用户）

```go
// internal/domain/user.go

package domain

import (
    "database/sql"
    "time"
)

// User 用户领域模型
type User struct {
    // 标识
    UserID          string
    TenantID        string
    
    // 账号信息
    UserAccount     string
    UserAccountHash []byte
    
    // 基本信息
    Nickname        sql.NullString
    
    // 角色和权限
    Role            string  // 角色代码（引用 roles.role_code）
    BranchTag       sql.NullString  // 分支标签（用于权限过滤）
    
    // 状态
    Status          UserStatus  // active, suspended, deleted
    
    // 联系方式（明文 + Hash）
    Email           sql.NullString
    Phone           sql.NullString
    EmailHash       []byte
    PhoneHash       []byte
    
    // 密码
    PasswordHash    []byte
    PinHash         []byte
    
    // Tag 关联（JSONB 数组）
    Tags            []string  // user_tag 类型
    
    // 告警配置
    AlarmLevels     []string
    AlarmChannels   []string
    AlarmScope      sql.NullString
    
    // 用户偏好
    Preferences     sql.NullString  // JSONB
    
    // 时间戳
    LastLoginAt     sql.NullTime
    
    // 注意：DB 表中没有 created_at 和 updated_at 字段
}

type UserStatus string

const (
    UserStatusActive    UserStatus = "active"
    UserStatusSuspended UserStatus = "suspended"
    UserStatusDeleted   UserStatus = "deleted"
)
```

#### 1.2.3 Tag（标签）

```go
// internal/domain/tag.go

package domain

// Tag 标签领域模型
type Tag struct {
    TagID       string
    TenantID    string
    TagType     TagType
    TagName     string
}

type TagType string

const (
    TagTypeBranchTag  TagType = "branch_tag"   // 系统预定义
    TagTypeFamilyTag  TagType = "family_tag"    // 系统预定义
    TagTypeAreaTag    TagType = "area_tag"      // 系统预定义
    TagTypeUserTag    TagType = "user_tag"      // 系统定义（租户新建）
)

// TagFilter 标签查询过滤器
type TagFilter struct {
    TenantID    string
    TagType     *TagType  // 可选，nil 表示查询所有类型
    TagName     string    // 可选，用于搜索
}
```

---

## 二、Repository 接口设计

### 2.1 设计原则

1. **强类型**：使用领域模型，不使用 `map[string]any`
2. **数据一致性**：替代数据库触发器
3. **事务管理**：单 Repository 内的事务
4. **接口抽象**：便于测试和替换实现

### 2.2 Repository 接口定义

#### 2.2.1 ResidentsRepository

```go
// internal/repository/residents_repo.go

package repository

import (
    "context"
    "wisefido-data/internal/domain"
)

// ResidentsRepository 住户数据访问接口
type ResidentsRepository interface {
    // 查询
    GetResident(ctx context.Context, tenantID, residentID string) (*domain.Resident, error)
    ListResidents(ctx context.Context, filter ResidentsFilter) ([]*domain.Resident, int, error)
    
    // 创建（替代触发器：trigger_sync_family_tag）
    CreateResident(ctx context.Context, tenantID string, resident *domain.Resident) (string, error)
    
    // 更新（替代触发器：trigger_sync_family_tag）
    UpdateResident(ctx context.Context, tenantID, residentID string, resident *domain.Resident) error
    
    // 删除（替代触发器：trigger_cleanup_resident_from_tags）
    DeleteResident(ctx context.Context, tenantID, residentID string) error
    
    // PHI 操作
    GetResidentPHI(ctx context.Context, tenantID, residentID string) (*domain.ResidentPHI, error)
    UpsertResidentPHI(ctx context.Context, tenantID, residentID string, phi *domain.ResidentPHI) error
    
    // Contact 操作
    GetResidentContacts(ctx context.Context, tenantID, residentID string) ([]*domain.ResidentContact, error)
    CreateResidentContact(ctx context.Context, tenantID, residentID string, contact *domain.ResidentContact) (string, error)
    UpdateResidentContact(ctx context.Context, tenantID, contactID string, contact *domain.ResidentContact) error
    DeleteResidentContact(ctx context.Context, tenantID, contactID string) error
    
    // Caregiver 操作
    GetResidentCaregivers(ctx context.Context, tenantID, residentID string) ([]*domain.ResidentCaregiver, error)
    UpsertResidentCaregiver(ctx context.Context, tenantID, residentID string, caregiver *domain.ResidentCaregiver) error
}

// ResidentsFilter 住户查询过滤器
type ResidentsFilter struct {
    TenantID        string
    Search          string  // 搜索 nickname 或 unit_name
    Status          string
    ServiceLevel    string
    FamilyTag       string
    UnitID          string
    RoomID          string
    BedID           string
    
    // 权限过滤
    AssignedUserID  string  // 仅查询分配给该用户的住户
    BranchTag       string  // 仅查询该分支的住户
    
    // 分页
    Page            int
    Size            int
}
```

---

## 三、关键设计决策

### 3.1 强类型 vs map[string]any

**决策**：使用强类型领域模型
- ✅ 类型安全
- ✅ IDE 自动补全
- ✅ 编译时检查
- ❌ 不使用 `map[string]any`

### 3.2 数据一致性

**决策**：在 Repository 层维护数据一致性（替代触发器）
- ✅ 可测试
- ✅ 可调试
- ✅ 可维护
- ❌ 不使用数据库触发器

### 3.3 事务管理

**决策**：
- Repository 层：单 Repository 内的事务
- Service 层：跨 Repository 的事务

---

## 四、实现顺序

### Phase 1: 领域模型 + Repository 接口
1. ❌ 定义领域模型（`internal/domain/`）- **待实现**
2. ❌ 定义 Repository 接口（`internal/repository/*_repo.go`）- **待实现**

### Phase 2: Repository 实现（替代触发器）
1. ⚠️ 实现 `PostgresResidentsRepository`（部分实现，使用 map[string]any，需改为强类型）
   - ✅ 已替代：`trigger_sync_family_tag`（CreateResident, UpdateResident）
   - ✅ 已替代：`trigger_cleanup_resident_from_tags`（DeleteResident）
2. ❌ 实现 `PostgresUsersRepository` - **待实现**
   - ❌ 需替代：`trigger_sync_user_tags`（INSERT/UPDATE users.tags）
   - ❌ 需替代：`trigger_cleanup_user_from_tags`（DELETE users）
3. ❌ 实现 `PostgresTagsRepository` - **待实现**
4. ⚠️ 实现 `PostgresUnitsRepository`（部分实现，使用 map[string]any）
   - ✅ 已替代：`trigger_sync_units_groupList_to_cards`（UpdateUnit 中的 syncUnitGroupListToCards）
5. ❌ 实现 `PostgresCardsRepository` - **待实现**
   - **注意**：cards 是实体表，但数据由应用层计算和维护（devices, residents 是 JSONB）
   - 需要维护 cards 表的数据一致性（替代相关触发器）

### Phase 3: Service 层
1. ❌ 实现 `ResidentService` - **待实现**
2. ❌ 实现 `UserService` - **待实现**
3. ❌ 实现 `TagService` - **待实现**

### Phase 4: Handler 层
1. ❌ 重构 `admin_residents_handlers.go`（3032 行 → ~200 行）- **待实现**
2. ❌ 重构 `admin_users_handlers.go`（1257 行 → ~200 行）- **待实现**
3. ❌ 重构 `admin_tags_handlers.go`（576 行 → ~200 行）- **待实现**

---

## 五、注意事项

### 5.1 已移除的字段

**ResidentPHI 表中已移除的字段**：
- ❌ `HIS_resident_name` - 已移除
- ❌ `HIS_resident_admission_date` - 已移除
- ❌ `HIS_resident_discharge_date` - 已移除
- ❌ `HIS_resident_metadata` - 已移除

**当前 ResidentPHI 表包含的字段**（基于 `09_resident_phi.sql`）：
- ✅ 基本信息：first_name, last_name, gender, date_of_birth
- ✅ 联系方式：resident_phone, resident_email
- ✅ 健康信息：weight_lb, height_ft, height_in, mobility_level
- ✅ 功能性健康：tremor_status, mobility_aid, adl_assistance, comm_status
- ✅ 疾病史：has_hypertension, has_hyperlipaemia, has_hyperglycaemia, has_stroke_history, has_paralysis, has_alzheimer, medical_history
- ✅ 家庭地址：home_address_street, home_address_city, home_address_state, home_address_postal_code, plus_code

### 5.2 字段对比修正

**已修正的不一致**：

1. **Resident 表**：
   - ✅ 添加了 `Role` 字段（DB 中有，固定为 'Resident'）
   - ✅ 移除了 `CreatedAt` 和 `UpdatedAt`（DB 中没有）

2. **ResidentCaregiver 表**：
   - ✅ 将 `CaregiverID` 改为 `ID`（DB 中的主键字段名是 `id`）

3. **User 表**：
   - ✅ 移除了 `CreatedAt` 和 `UpdatedAt`（DB 中没有）

4. **ResidentPHI 表**：
   - ✅ 已确认移除了所有 HIS 相关字段

### 5.3 触发器替代状态

**重要说明**：既然 `tag_objects` 字段已删除，没有冗余数据了，**不需要反向索引维护**。

**trigger_sync_family_tag 的作用**：
- 现在只是维护 `tags_catalog` 目录（调用 `upsert_tag_to_catalog()`）
- **不是反向索引**（因为 `tag_objects` 已删除）
- 应用层只需要在创建/更新 resident 时调用 `upsert_tag_to_catalog()` 即可
- **不需要** `update_tag_objects()` 和 `drop_object_from_all_tags()`（这些函数会报错，因为 `tag_objects` 字段已删除）

**当前状态**：
- ⚠️ `trigger_sync_family_tag` - **可以保留**（自动维护目录），或**应用层手动调用** `upsert_tag_to_catalog()`
- ⚠️ `trigger_cleanup_resident_from_tags` - **已删除**（不需要清理，因为没有 `tag_objects`）
- ✅ `trigger_sync_units_groupList_to_cards` - **保留**（维护 cards.routing_alarm_tags，不是反向索引）

**待处理的触发器**：
- ⚠️ `trigger_sync_user_tags` - **可以保留**（自动维护目录），或**应用层手动调用** `upsert_tag_to_catalog()`
- ⚠️ `trigger_cleanup_user_from_tags` - **已删除**（不需要清理，因为没有 `tag_objects`）

**PostgresResidentsRepo 中的错误**：
- ❌ `syncFamilyTagToCatalogTx()` 调用了 `update_tag_objects()` - **会报错**（`tag_objects` 字段已删除）
- ❌ `removeResidentFromTagTx()` 调用了 `update_tag_objects()` - **会报错**
- ❌ `dropResidentFromAllTagsTx()` 调用了 `drop_object_from_all_tags()` - **会报错**

**修正方案**：
- 只需要调用 `upsert_tag_to_catalog()` 维护目录
- 删除所有 `update_tag_objects()` 和 `drop_object_from_all_tags()` 的调用

**保留的触发器**（数据验证类，可能保留）：
- ⚠️ `trigger_validate_*` - 数据验证类触发器（25_integrity_constraints.sql）
- ⚠️ `trigger_residents_lowercase_account` - 数据格式转换（可保留或移到应用层）
- ⚠️ `trigger_users_lowercase_account` - 数据格式转换（可保留或移到应用层）

### 5.4 Cards 表说明

**Cards 表特性**：
- ✅ **实体表**（CREATE TABLE），不是视图
- ⚠️ **数据由应用层维护**：
  - `devices` (JSONB) - 应用层计算和维护
  - `residents` (JSONB) - 应用层计算和维护
  - `routing_alarm_tags` (VARCHAR[]) - 应用层维护（或通过触发器 `trigger_sync_units_groupList_to_cards`）
  - `unhandled_alarm_*` (INTEGER) - 应用层维护
- **设计考虑**：
  - 需要 `CardsRepository` 来维护 cards 表的数据一致性
  - 或者通过其他 Repository（如 UnitsRepository）在更新时同步更新 cards 表

### 5.5 代码清理

**需要检查的遗留代码**：
- `admin_residents_handlers.go` 中可能还有对 HIS 字段的引用（第 1426-1454 行）
- 需要清理这些遗留代码，确保与数据库结构一致

