# wisefido-data 架构设计（分层 + 领域模型）

> 合并自原 `ARCHITECTURE_DESIGN.md`（三层架构）+ `BOTTOM_UP_DESIGN.md`（自底向上 + 领域模型），2026-05-02。
> 原文档基线日期 2026-04-16；本文删除已过时的"Phase 进度"描述，保留长期有效的架构原则与领域模型契约。当前实现状态以代码为准。

---

## 1. 业务领域边界

```
┌─────────────────────────────────────────────────────────┐
│ 平台层（Platform）                                       │
│  - 租户管理（Tenants）                                   │
└─────────────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────────┐
│ 用户权限层（Auth & Access）                              │
│  - Users / Roles / RolePermissions / Auth                │
└─────────────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────────┐
│ 业务层（Business）                                        │
│  ├─ 地址层级：Buildings / Units / Rooms / Beds            │
│  ├─ 住户管理：Resident / ResidentPHI / Contact / Caregiver │
│  ├─ 设备管理：Devices / DeviceStore                       │
│  ├─ 标签管理：Tag / TagsCatalog                           │
│  └─ 告警管理：alarm_cloud / alarm_device / alarm_events   │
└─────────────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────────┐
│ 数据查询层（Data）                                       │
│  - Vital Focus / 卡片数据                                │
└─────────────────────────────────────────────────────────┘
```

---

## 2. 三层架构原则

```
┌─────────────┐
│   Handler   │  HTTP 请求/响应处理
└──────┬──────┘  - 解析参数、生成 JSON、错误处理
       ↓
┌─────────────┐
│   Service   │  业务逻辑
└──────┬──────┘  - 权限检查、业务规则、数据转换、跨 Repository 编排
       ↓
┌─────────────┐
│ Repository  │  数据访问
└──────┬──────┘  - SQL 操作、数据一致性（替代触发器）、单 Repo 内事务
       ↓
┌─────────────┐
│  Database   │
└─────────────┘
```

**职责边界**：

| 层 | 负责 | 不负责 |
|---|---|---|
| Handler | 解析 HTTP 请求、生成 JSON 响应、路由分发、错误捕获 | 业务规则、权限、数据转换、SQL |
| Service | 权限检查、业务规则验证、数据转换、跨 Repo 事务编排 | HTTP 处理、SQL |
| Repository | SQL 抽象、领域模型映射、数据一致性、单 Repo 内事务 | 业务规则、权限、HTTP |

**依赖方向**：`Handler → Service → Repository → Database`，**不允许反向依赖**。

**简单领域可省 Service**：Location / Device 当前直接 Handler→Repository，OK。

### 2.1 Service 层直接用 db 的例外

复杂 JOIN + 权限过滤的查询如果硬塞进 Repository 接口会让接口爆炸。允许 Service 层直接 `s.db.QueryContext()`，但必须封装为 Service 内部方法，不暴露给 Handler。

```go
// ✅ 允许：跨表 JOIN + 当前用户权限过滤
func (s *userService) ListUsers(ctx context.Context, req ListUsersRequest) (*ListUsersResponse, error) {
    query := `
        SELECT u.*, b.branch_name
        FROM users u
        LEFT JOIN LATERAL (
            SELECT branch_name FROM branches b
            JOIN user_branches ub ON b.branch_id = ub.branch_id
            WHERE ub.user_id = u.user_id
            LIMIT 1
        ) b ON true
        WHERE u.tenant_id = $1
        -- 还有按当前用户 BranchTag 的权限过滤
    `
    rows, err := s.db.QueryContext(ctx, query, ...)
    // ...
}
```

判定规则：
- **简单查询**（单表 / 用过滤条件就能搞定）→ Repository
- **跨表 JOIN + 业务级权限过滤** → 允许 Service 层 db，但封装为内部方法
- **跨 Repository 写操作的事务** → Service 层 `db.BeginTx`，但**仍调 Repository 方法**（Repository 接口需扩展支持 tx 参数），不要在 Service 里手写 INSERT/UPDATE/DELETE

反例（要改）：
```go
// ❌ Service 直接写跨表 DELETE+INSERT 应改为调 Repository
func (s *userService) updateUserBranches(ctx, tenantID, userID string, branchIDs []string) error {
    tx, _ := s.db.BeginTx(ctx, nil)
    _, _ = tx.ExecContext(ctx, `DELETE FROM user_branches WHERE user_id=$1`, userID)
    // ...
}
```

---

## 3. 设计决策

| 决策 | 选择 | 理由 |
|---|---|---|
| 数据结构 | 强类型领域模型 | 类型安全、IDE 补全、编译时检查 |
| 不用 `map[string]any` | — | 无类型校验，难维护；与触发器语义耦合 |
| 数据一致性 | Repository 层维护 | 可测试、可调试，替代隐式 DB 触发器 |
| 事务边界 | 单 Repo 内事务在 Repo 层；跨 Repo 事务在 Service 层 | 隔离 + 复用 |
| 触发器策略 | 数据格式/合规类保留；反向索引/级联删除迁到应用层 | DB 触发器对应用透明、调试困难 |

---

## 4. 领域模型契约

### 4.1 Resident（住户）

```go
// internal/domain/resident.go
type Resident struct {
    ResidentID          string
    TenantID            string
    ResidentAccount     string
    ResidentAccountHash []byte
    Nickname            string
    Status              ResidentStatus  // active / discharged / transferred
    ServiceLevel        sql.NullString
    Role                string          // 固定 'Resident'

    AdmissionDate       time.Time
    DischargeDate       sql.NullTime

    // 位置
    UnitID              string
    RoomID              sql.NullString
    BedID               sql.NullString

    // Tag
    FamilyTag           sql.NullString

    // 权限
    CanViewStatus       bool

    // 联系方式（Hash）
    PhoneHash           []byte
    EmailHash           []byte
    PasswordHash        []byte

    Note                sql.NullString
    Metadata            sql.NullString
    // 注意：DB 表无 created_at / updated_at
}

type ResidentStatus string
const (
    ResidentStatusActive      ResidentStatus = "active"
    ResidentStatusDischarged  ResidentStatus = "discharged"
    ResidentStatusTransferred ResidentStatus = "transferred"
)
```

### 4.2 ResidentPHI（个人健康信息，HIPAA 加密字段）

```go
type ResidentPHI struct {
    PHIID       string
    TenantID    string
    ResidentID  string

    // 基本信息（PII，AES-256-GCM 加密 → *_enc 列）
    FirstName   sql.NullString
    LastName    sql.NullString
    Gender      sql.NullString
    DateOfBirth sql.NullTime

    // 联系
    ResidentPhone sql.NullString
    ResidentEmail sql.NullString

    // 健康
    WeightLb      sql.NullFloat64
    HeightFt      sql.NullFloat64
    HeightIn      sql.NullFloat64
    MobilityLevel sql.NullInt64
    TremorStatus  sql.NullString
    MobilityAid   sql.NullString
    ADLAssistance sql.NullString
    CommStatus    sql.NullString

    // 疾病史
    HasHypertension   sql.NullBool
    HasHyperlipaemia  sql.NullBool
    HasHyperglycaemia sql.NullBool
    HasStrokeHistory  sql.NullBool
    HasParalysis      sql.NullBool
    HasAlzheimer      sql.NullBool
    MedicalHistory    sql.NullString

    // 家庭地址（Home 场景）
    HomeAddressStreet     sql.NullString
    HomeAddressCity       sql.NullString
    HomeAddressState      sql.NullString
    HomeAddressPostalCode sql.NullString
    PlusCode              sql.NullString
}
```

> PHI 加密设计见 [doc/kms.md](../doc/kms.md)。HIS_* 系列字段已移除（admission/discharge_date 等）。

### 4.3 ResidentContact / ResidentCaregiver

```go
type ResidentContact struct {
    ContactID, TenantID, ResidentID string
    Slot                            string  // A/B/C/D/E
    IsEnabled, IsEmergencyContact   bool
    Relationship                    sql.NullString
    Role                            string  // 固定 'Family'
    ContactFirstName, ContactLastName, ContactPhone, ContactEmail sql.NullString
    ReceiveSMS, ReceiveEmail        bool
    PhoneHash, EmailHash, PasswordHash []byte
    AlertTimeWindow                 sql.NullString  // JSONB
}

type ResidentCaregiver struct {
    CaregiverID, TenantID, ResidentID string
    GroupList sql.NullString  // JSONB: ["tag1", "tag2"]
    UserList  sql.NullString  // JSONB: ["user_id1", "user_id2"]
}
```

### 4.4 User

```go
type User struct {
    UserID, TenantID                 string
    UserAccount                      string
    UserAccountHash                  []byte
    Nickname                         sql.NullString
    Role                             string          // 引用 roles.role_code
    BranchTag                        sql.NullString
    Status                           UserStatus      // active / suspended / deleted

    // 联系方式（明文 + Hash）
    Email, Phone                     sql.NullString
    EmailHash, PhoneHash             []byte
    PasswordHash, PinHash            []byte

    Tags                             []string        // user_tag JSONB
    AlarmLevels, AlarmChannels       []string
    AlarmScope                       sql.NullString
    Preferences                      sql.NullString  // JSONB
    LastLoginAt                      sql.NullTime
    // 注意：DB 表无 created_at / updated_at
}

type UserStatus string
const (
    UserStatusActive    UserStatus = "active"
    UserStatusSuspended UserStatus = "suspended"
    UserStatusDeleted   UserStatus = "deleted"
)
```

### 4.5 Tag

```go
type Tag struct {
    TagID, TenantID string
    TagType         TagType
    TagName         string
}

type TagType string
const (
    TagTypeBranchTag TagType = "branch_tag"  // 系统预定义
    TagTypeFamilyTag TagType = "family_tag"  // 系统预定义
    TagTypeAreaTag   TagType = "area_tag"    // 系统预定义
    TagTypeUserTag   TagType = "user_tag"    // 租户自定义
)
```

---

## 5. Repository 接口模板

### 5.1 ResidentsRepository

```go
type ResidentsRepository interface {
    // 查询
    GetResident(ctx, tenantID, residentID) (*domain.Resident, error)
    ListResidents(ctx, filter ResidentsFilter) ([]*domain.Resident, total int, error)

    // 写（Create/Update 替代 trigger_sync_family_tag）
    CreateResident(ctx, tenantID, *domain.Resident) (residentID string, error)
    UpdateResident(ctx, tenantID, residentID, *domain.Resident) error
    DeleteResident(ctx, tenantID, residentID) error  // 替代 trigger_cleanup_resident_from_tags

    // PHI（透过 KMS 加解密）
    GetResidentPHI(ctx, tenantID, residentID) (*domain.ResidentPHI, error)
    UpsertResidentPHI(ctx, tenantID, residentID, *domain.ResidentPHI) error

    // Contact
    GetResidentContacts(ctx, tenantID, residentID) ([]*domain.ResidentContact, error)
    CreateResidentContact(ctx, tenantID, residentID, *domain.ResidentContact) (string, error)
    UpdateResidentContact(ctx, tenantID, contactID, *domain.ResidentContact) error
    DeleteResidentContact(ctx, tenantID, contactID) error

    // Caregiver
    GetResidentCaregivers(ctx, tenantID, residentID) ([]*domain.ResidentCaregiver, error)
    UpsertResidentCaregiver(ctx, tenantID, residentID, *domain.ResidentCaregiver) error
}

type ResidentsFilter struct {
    TenantID                          string
    Search                            string  // nickname / unit_name
    Status, ServiceLevel, FamilyTag   string
    UnitID, RoomID, BedID             string
    AssignedUserID                    string  // 仅查分配给该用户的
    BranchTag                         string  // 仅查该 branch 的
    Page, Size                        int
}
```

---

## 6. 触发器迁移指引

| 触发器 | 现状 | 应用层做法 |
|---|---|---|
| `trigger_sync_family_tag` | 保留（自动维护 tags_catalog 目录） | Create/Update Resident 时调 `upsert_tag_to_catalog()` |
| `trigger_cleanup_resident_from_tags` | 已删除（`tag_objects` 字段已删，无需反向清理） | 不做 |
| `trigger_sync_units_groupList_to_cards` | 保留（维护 cards.routing_alarm_tags） | UnitsRepo 同步调用即可 |
| `trigger_sync_user_tags` | 保留（自动维护目录） | 同 family_tag |
| `trigger_cleanup_user_from_tags` | 已删除 | 不做 |
| `trigger_validate_*` | 保留（数据校验类） | 不做 |
| `trigger_*_lowercase_account` | 保留或迁应用层（数据格式转换） | 二选一 |

> **PostgresResidentsRepo 已知问题**：旧实现里调 `update_tag_objects()` / `drop_object_from_all_tags()` 会报错（`tag_objects` 字段已删除），需改为只调 `upsert_tag_to_catalog()`。

---

## 7. Cards 表

`cards` 是**实体表**，但 `devices` / `residents` (JSONB) / `routing_alarm_tags` / `unhandled_alarm_*` 全部由**应用层**计算和维护：

- 设备/住户/单元变化时，service 层调 wisefido-card-aggregator API（或事件驱动）触发 card 重算
- 现有 `cardCreator.CreateCardsForUnit()` 是入口
- cardagg 自身定时全量轮询是兜底机制

cards 表数据契约见 [docs/Reside_stream_stand.md](Reside_stream_stand.md) + cardagg 实现。

---

## 8. 已知 follow-up

- 大 Handler（`admin_residents_handlers.go` 3000+ 行 / `admin_users_handlers.go` 1200+ 行 / `auth_handlers.go` 800+ 行 / `admin_tags_handlers.go` 580+ 行）仍待按本文三层原则拆分
- `PostgresResidentsRepo` 改强类型 + 修 `update_tag_objects` 调用
- `PostgresUsersRepo` / `PostgresTagsRepo` / `PostgresRolesRepo` / `PostgresRolePermissionsRepo` 待实现
