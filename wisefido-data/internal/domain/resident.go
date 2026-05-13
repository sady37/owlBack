// Package domain — Resident v2 types (Forward Design)
//
// 完全脱离 v1 schema：不复用 v1 DTO / UpdateField / Bool/Time-ptr 模式。
// 所有空间归属用 IPv6 INET CIDR 字符串；hoa 是唯一业务 ID。
//
// v2 schema 参考：
//   residents (hoa INET PK, resident_slot, nickname NOT NULL, resident_account UNIQUE,
//              service_tier, move_in_date, move_out_date, status, gender, birth_year, notes)
//   resident_unit (resident_id, spatial_prefix INET, valid_from/to)
//   resident_caregivers (resident_id, caregiver_id OR care_team_id, role, is_primary)
//   resident_phi (hoa, *_enc/iv/tag 加密字段) — 加密链路 Phase 3b
//   resident_contacts (resident_id, slot, ...) — Phase 3b
package domain

// Resident 主体（list / get 公共字段）
//
// 字段名稳定保留 v1 业务命名；type 改为 IPv6 CIDR 字符串：
//   resident_id (= HoA, IPv6 /128 host string e.g. "fd00:0:3:ff01:1::")
//   tenant_id   (IPv6 /48 CIDR)
//   branch_id   (IPv6 /56 CIDR — 反推自 active resident_unit)
//
// DB 列名 2026-05-10 已 rename 回 v1：admission_date / discharge_date / service_level / note
type Resident struct {
	ResidentID      string  `json:"resident_id"`
	TenantID        string  `json:"tenant_id"`
	BranchID        string  `json:"branch_id,omitempty"`
	ResidentSlot    int     `json:"resident_slot"`
	ResidentAccount string  `json:"resident_account,omitempty"`
	Nickname        string  `json:"nickname"`
	Status          string  `json:"status"`
	ServiceLevel    *string `json:"service_level,omitempty"`
	// gender / birth_year 已挪到 ResidentPHIv2 加密存（2026-05-11 HIPAA Safe Harbor）
	AdmissionDate   *string `json:"admission_date,omitempty"`
	DischargeDate   *string `json:"discharge_date,omitempty"`
	FamilyAccess    *bool   `json:"family_access,omitempty"`  // 是否允许 role=Family 用户登录
	Note            *string `json:"note,omitempty"`

	// 空间反推字段（来自 active resident_unit + sites/branches/units/rooms/beds JOIN）
	// 仅 list 视图填充；非 active 或未分配时为 nil/空。
	UnitName     *string `json:"unit_name,omitempty"`
	RoomName     *string `json:"room_name,omitempty"`
	BedName      *string `json:"bed_name,omitempty"`
	BuildingName *string `json:"building_name,omitempty"`
	BranchName   *string `json:"branch_name,omitempty"`
	// FacilityType 沿用 ResidentDetail.UnitType 枚举：1=Private / 2=Share / 3=Public
	FacilityType *int `json:"facility_type,omitempty"`
	// Property 由 tenants.kind 映射：B2B → "facility" / B2C → "home"
	Property *string `json:"property,omitempty"`
}

// ResidentDetail = main + 关联（GetResident 用）
type ResidentDetail struct {
	Resident

	// 当前空间分配（来自 active resident_unit）— 字段名稳定保留 v1，值是 IPv6 CIDR
	UnitID       *string `json:"unit_id,omitempty"`       // /80 CIDR
	RoomID       *string `json:"room_id,omitempty"`       // /88 CIDR
	BedID        *string `json:"bed_id,omitempty"`        // /96 CIDR
	UnitType     *int    `json:"unit_type,omitempty"`     // 1=Private / 2=Share / 3=Public
	BranchName   *string `json:"branch_name,omitempty"`
	BuildingName *string `json:"building_name,omitempty"`
	Floor        *int    `json:"floor,omitempty"`
	UnitName     *string `json:"unit_name,omitempty"`
	RoomName     *string `json:"room_name,omitempty"`
	BedName      *string `json:"bed_name,omitempty"`

	// 关联：resident_caregivers 一表，护理（caregiver/team 二选一）+ 家属(family_id 独立)
	Caregivers []ResidentCaregiverV2 `json:"caregivers,omitempty"`
	Teams      []ResidentTeamV2      `json:"teams,omitempty"`
	Family     []ResidentFamilyV2    `json:"family,omitempty"`

	// PHI（加解密）
	PHI *ResidentPHIv2 `json:"phi,omitempty"`

	// Contacts（紧急联系人，多行）
	Contacts []ResidentContactV2 `json:"contacts,omitempty"`
}

// ResidentCaregiverV2 — caregiver 直接绑定 (resident_caregivers.caregiver_id)
// v2: 业务标识 = users.user_id (UUID)；不返 hoa（Phase B' 后 admin/family/caregiver 等多数 user 没 hoa）
type ResidentCaregiverV2 struct {
	UserID      string `json:"user_id"`
	UserAccount string `json:"user_account"`
	Nickname    string `json:"nickname,omitempty"`
	Role        string `json:"role,omitempty"` // Caregiver / Nurse / Manager / Individual
}

// ResidentFamilyV2 — family 绑定 (resident_caregivers.family_id, users.role=Family)
type ResidentFamilyV2 struct {
	UserID      string `json:"user_id"`
	UserAccount string `json:"user_account"`
	Nickname    string `json:"nickname,omitempty"`
}

// ResidentTeamV2 — care team 间接关联 (resident_caregivers.care_team_id)
type ResidentTeamV2 struct {
	TeamID   string `json:"team_id"`
	TeamName string `json:"team_name"`
	TeamKind string `json:"team_kind,omitempty"`
}

// ResidentPHIv2 — 全字段 AES-256-GCM 加密；FE/BE/DB 字段名 1:1
type ResidentPHIv2 struct {
	// HIPAA min-necessary: 所有字段 optional，空值不写入
	FirstName             *string  `json:"first_name,omitempty"`
	LastName              *string  `json:"last_name,omitempty"`
	Gender                *string  `json:"gender,omitempty"`        // Male / Female / Other
	DateOfBirth           *string  `json:"date_of_birth,omitempty"` // ISO date (YYYY-MM-DD)；仅选年份 → YYYY-01-01
	ResidentPhone         *string  `json:"resident_phone,omitempty"`
	ResidentEmail         *string  `json:"resident_email,omitempty"`
	WeightLb              *float64 `json:"weight_lb,omitempty"`
	HeightFt              *float64 `json:"height_ft,omitempty"`
	HeightIn              *float64 `json:"height_in,omitempty"`
	MobilityLevel         *int     `json:"mobility_level,omitempty"`
	TremorStatus          *string  `json:"tremor_status,omitempty"`
	MobilityAid           *string  `json:"mobility_aid,omitempty"`
	ADLAssistance         *string  `json:"adl_assistance,omitempty"`
	CommStatus            *string  `json:"comm_status,omitempty"`
	HasHypertension       *bool    `json:"has_hypertension,omitempty"`
	HasHyperlipaemia      *bool    `json:"has_hyperlipaemia,omitempty"`
	HasHyperglycaemia     *bool    `json:"has_hyperglycaemia,omitempty"`
	HasStrokeHistory      *bool    `json:"has_stroke_history,omitempty"`
	HasParalysis          *bool    `json:"has_paralysis,omitempty"`
	HasAlzheimer          *bool    `json:"has_alzheimer,omitempty"`
	MedicalHistory        *string  `json:"medical_history,omitempty"`
	HomeAddressStreet     *string  `json:"home_address_street,omitempty"`
	HomeAddressCity       *string  `json:"home_address_city,omitempty"`
	HomeAddressState      *string  `json:"home_address_state,omitempty"`
	HomeAddressPostalCode *string  `json:"home_address_postal_code,omitempty"`
	PlusCode              *string  `json:"plus_code,omitempty"` // 短码明文存

	// OTP 一次性授权码（独立 feature，不属常规 PHI 字段）
	OTP *ResidentOTPv2 `json:"otp,omitempty"`
}

// ResidentOTPv2 — resident ↔ caregiver 匿名身份认证 OTP
type ResidentOTPv2 struct {
	Code      *string `json:"code,omitempty"`       // 当前有效 OTP
	Purpose   *string `json:"purpose,omitempty"`    // 'caregiver_visit' / 'emergency_dispatch'
	IssuedAt  *string `json:"issued_at,omitempty"`  // ISO timestamp
	ExpiresAt *string `json:"expires_at,omitempty"`
	UsedAt    *string `json:"used_at,omitempty"`
	UsedBy    *string `json:"used_by,omitempty"` // caregiver user_id (uuid)
}

// ResidentContactV2 — 紧急/家属联系人（resident_contacts 表，多行）
type ResidentContactV2 struct {
	ContactID         string  `json:"contact_id"` // uuid
	ResidentID        string  `json:"resident_id,omitempty"`
	LinkedUserID      *string `json:"linked_user_id,omitempty"` // 关联已注册 owl user (uuid)
	Relationship      string  `json:"relationship"`             // NOT NULL
	ContactFirstName  *string `json:"contact_first_name,omitempty"`
	ContactLastName   *string `json:"contact_last_name,omitempty"`
	ContactPhone      *string `json:"contact_phone,omitempty"`
	ContactEmail      *string `json:"contact_email,omitempty"`
	ReceiveSMS        bool    `json:"receive_sms"`
	ReceiveEmail      bool    `json:"receive_email"`
}

// ===========================================================================
// Input types (Create / Update) — 平铺，无嵌套；JSON friendly
// ===========================================================================

// ResidentCreateInput — POST /admin/api/v2/residents
type ResidentCreateInput struct {
	Nickname string `json:"nickname"` // 必填

	ResidentAccount *string `json:"resident_account,omitempty"`
	ServiceLevel    *string `json:"service_level,omitempty"`
	AdmissionDate   *string `json:"admission_date,omitempty"`
	FamilyAccess    *bool   `json:"family_access,omitempty"`
	Note            *string `json:"note,omitempty"`

	// 空间分配（值 IPv6 CIDR string）；优先级 Bed > Room > Unit > Branch
	// BranchID = /56 "招待状态"：先建档落到 branch，后续再分到具体 unit/room/bed
	// （schema 31_resident_unit.sql ru_scope_valid 允许 masklen ∈ {48,56,80,88,96}）
	UnitID   *string `json:"unit_id,omitempty"`   // /80 CIDR
	RoomID   *string `json:"room_id,omitempty"`   // /88 CIDR
	BedID    *string `json:"bed_id,omitempty"`    // /96 CIDR
	BranchID *string `json:"branch_id,omitempty"` // /56 CIDR — 仅 Admin/Manager；写入 spatial_prefix 让 staff scope 可见

	// 关联
	CaregiverUserIDs []string `json:"caregiver_user_ids,omitempty"`
	CareTeamIDs      []string `json:"care_team_ids,omitempty"`
	FamilyUserIDs    []string `json:"family_user_ids,omitempty"`

	PHI      *ResidentPHIv2      `json:"phi,omitempty"`
	Contacts []ResidentContactV2 `json:"contacts,omitempty"`
}

// ResidentUpdateInput — PUT /admin/api/v2/residents/{hoa}
// 所有字段 pointer / sentinel：nil = 不改；非 nil = 改（含空字符串 / 空数组 = 清空）
type ResidentUpdateInput struct {
	// 主表
	Nickname        *string `json:"nickname,omitempty"`
	ResidentAccount *string `json:"resident_account,omitempty"`
	Status          *string `json:"status,omitempty"`
	ServiceLevel    *string `json:"service_level,omitempty"`
	AdmissionDate   *string `json:"admission_date,omitempty"`
	DischargeDate   *string `json:"discharge_date,omitempty"`
	FamilyAccess    *bool   `json:"family_access,omitempty"`
	Note            *string `json:"note,omitempty"`

	// 空间分配（任一字段提供即视为重新分配；空字符串 = 解绑）
	UnitID *string `json:"unit_id,omitempty"`
	RoomID *string `json:"room_id,omitempty"`
	BedID  *string `json:"bed_id,omitempty"`

	// 关联（提供即重置；空数组 = 显式清空）
	CaregiverUserIDs *[]string `json:"caregiver_user_ids,omitempty"`
	CareTeamIDs      *[]string `json:"care_team_ids,omitempty"`
	FamilyUserIDs    *[]string `json:"family_user_ids,omitempty"`

	// PHI / Contacts（提供即更新）
	PHI      *ResidentPHIv2       `json:"phi,omitempty"`
	Contacts *[]ResidentContactV2 `json:"contacts,omitempty"`
}

// ResidentListFilter
type ResidentListFilter struct {
	Status        string // active / deleted / 空=非 deleted
	Search        string // ILIKE nickname/resident_account
	IncludeDelete bool   // 默认 false
	Page          int
	PageSize      int

	// BranchPrefix — Current Branch scope 过滤 (/56 CIDR)；service 填，repo SQL EXISTS resident_unit
	// 用于 Manager/Nurse 等按当前业务 branch 看 resident（不跨 branch）
	BranchPrefix string

	// FamilyUserID — Family role scope 强制：只返回 resident_caregivers.family_id = $FamilyUserID 的 resident
	// service 层根据 X-User-Id + Family role 填入；handler/外部不直接设置
	FamilyUserID string
}

// ResidentClearCheckResult — 硬删前置 check
type ResidentClearCheckResult struct {
	CanClear         bool   `json:"can_clear"`
	AlarmEventsCount int    `json:"alarm_events_count"`
	EventLogCount    int    `json:"event_log_count"`
	MonitorCount     int    `json:"monitor_count"`
	Reason           string `json:"reason,omitempty"` // 不能 clear 时的解释
}
