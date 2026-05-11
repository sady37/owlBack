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

// ResidentV2 主体（list / get 公共字段）
//
// 字段名稳定保留 v1 业务命名；type 改为 IPv6 CIDR 字符串：
//   resident_id (= HoA, IPv6 /128 host string e.g. "fd00:0:3:ff01:1::")
//   tenant_id   (IPv6 /48 CIDR)
//   branch_id   (IPv6 /56 CIDR — 反推自 active resident_unit)
//
// DB 列名 2026-05-10 已 rename 回 v1：admission_date / discharge_date / service_level / note
type ResidentV2 struct {
	ResidentID      string  `json:"resident_id"`
	TenantID        string  `json:"tenant_id"`
	BranchID        string  `json:"branch_id,omitempty"`
	ResidentSlot    int     `json:"resident_slot"`
	ResidentAccount string  `json:"resident_account,omitempty"`
	Nickname        string  `json:"nickname"`
	Status          string  `json:"status"`
	ServiceLevel    *string `json:"service_level,omitempty"`
	Gender          *string `json:"gender,omitempty"`
	BirthYear       *int    `json:"birth_year,omitempty"`
	AdmissionDate   *string `json:"admission_date,omitempty"`
	DischargeDate   *string `json:"discharge_date,omitempty"`
	Note            *string `json:"note,omitempty"`
}

// ResidentV2Detail = main + 关联（GetResident 用）
type ResidentV2Detail struct {
	ResidentV2

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

	// 关联（resident_caregivers 一表二选一）
	Caregivers []ResidentCaregiverV2 `json:"caregivers,omitempty"`
	Teams      []ResidentTeamV2      `json:"teams,omitempty"`

	// PHI（Phase 3b 加解密）
	PHI *ResidentPHIv2 `json:"phi,omitempty"`
}

// ResidentCaregiverV2 — caregiver 直接绑定 (resident_caregivers.caregiver_id)
type ResidentCaregiverV2 struct {
	HoA       string `json:"hoa"`                  // caregiver user hoa
	UserID    string `json:"user_id"`              // users.user_id
	UserAccount string `json:"user_account"`
	Nickname  string `json:"nickname,omitempty"`
	Role      string `json:"role,omitempty"`       // Caregiver / Nurse / Manager / Individual
	IsPrimary bool   `json:"is_primary,omitempty"`
}

// ResidentTeamV2 — care team 间接关联 (resident_caregivers.care_team_id)
type ResidentTeamV2 struct {
	TeamID   string `json:"team_id"`
	TeamName string `json:"team_name"`
	TeamKind string `json:"team_kind,omitempty"`
}

// ResidentPHIv2 — Phase 3b 接通加密链路；当前 stub 显示明文 read-only
type ResidentPHIv2 struct {
	// HIPAA min-necessary: 所有字段 optional，空值不写入
	FirstName             *string  `json:"first_name,omitempty"`
	LastName              *string  `json:"last_name,omitempty"`
	DateOfBirth           *string  `json:"date_of_birth,omitempty"` // ISO date
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
}

// ===========================================================================
// Input types (Create / Update) — 平铺，无嵌套；JSON friendly
// ===========================================================================

// ResidentV2CreateInput — POST /admin/api/v2/residents
type ResidentV2CreateInput struct {
	Nickname string `json:"nickname"` // 必填

	ResidentAccount *string `json:"resident_account,omitempty"` // 留空 → 后端 'R{slot:0000}'
	ServiceLevel    *string `json:"service_level,omitempty"`    // 写 DB service_tier
	Gender          *string `json:"gender,omitempty"`
	BirthYear       *int    `json:"birth_year,omitempty"`
	AdmissionDate   *string `json:"admission_date,omitempty"`   // 写 DB move_in_date
	Note            *string `json:"note,omitempty"`             // 写 DB notes

	// 空间分配（值 IPv6 CIDR string）
	UnitID *string `json:"unit_id,omitempty"` // /80 CIDR
	RoomID *string `json:"room_id,omitempty"` // /88 CIDR
	BedID  *string `json:"bed_id,omitempty"`  // /96 CIDR

	// 关联（FE 二选一）
	CaregiverUserIDs []string `json:"caregiver_user_ids,omitempty"`
	CareTeamIDs      []string `json:"care_team_ids,omitempty"`

	PHI *ResidentPHIv2 `json:"phi,omitempty"`
}

// ResidentV2UpdateInput — PUT /admin/api/v2/residents/{hoa}
// 所有字段 pointer / sentinel：nil = 不改；非 nil = 改（含空字符串 / 空数组 = 清空）
type ResidentV2UpdateInput struct {
	// 主表
	Nickname        *string `json:"nickname,omitempty"`
	ResidentAccount *string `json:"resident_account,omitempty"`
	Status          *string `json:"status,omitempty"`
	ServiceLevel    *string `json:"service_level,omitempty"`
	Gender          *string `json:"gender,omitempty"`
	BirthYear       *int    `json:"birth_year,omitempty"`
	AdmissionDate   *string `json:"admission_date,omitempty"` // Nurse 不可改（business rule）
	DischargeDate   *string `json:"discharge_date,omitempty"` // Nurse 不可改
	Note            *string `json:"note,omitempty"`

	// 空间分配（任一字段提供即视为重新分配；空字符串 = 解绑）
	UnitID *string `json:"unit_id,omitempty"`
	RoomID *string `json:"room_id,omitempty"`
	BedID  *string `json:"bed_id,omitempty"`

	// 关联（提供即重置；FE 二选一时另一方提空数组清残留）
	CaregiverUserIDs *[]string `json:"caregiver_user_ids,omitempty"` // *nil*=不改 / *empty*=清空 / *N items*=替换
	CareTeamIDs      *[]string `json:"care_team_ids,omitempty"`

	// PHI（提供即更新；Phase 3b 实际加密落库）
	PHI *ResidentPHIv2 `json:"phi,omitempty"`
}

// ResidentV2ListFilter
type ResidentV2ListFilter struct {
	Status        string // active / deleted / 空=非 deleted
	Search        string // ILIKE nickname/resident_account
	IncludeDelete bool   // 默认 false
	Page          int
	PageSize      int
}

// ResidentV2ClearCheckResult — 硬删前置 check
type ResidentV2ClearCheckResult struct {
	CanClear         bool   `json:"can_clear"`
	AlarmEventsCount int    `json:"alarm_events_count"`
	EventLogCount    int    `json:"event_log_count"`
	MonitorCount     int    `json:"monitor_count"`
	Reason           string `json:"reason,omitempty"` // 不能 clear 时的解释
}
