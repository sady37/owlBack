package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"

	"github.com/lib/pq"
	"go.uber.org/zap"
)

// ResidentService 住户管理服务接口
type ResidentService interface {
	// 查询
	ListResidents(ctx context.Context, req ListResidentsRequest) (*ListResidentsResponse, error)
	GetResident(ctx context.Context, req GetResidentRequest) (*GetResidentResponse, error)

	// 创建
	CreateResident(ctx context.Context, req CreateResidentRequest) (*CreateResidentResponse, error)

	// 更新
	UpdateResident(ctx context.Context, req UpdateResidentRequest) (*UpdateResidentResponse, error)
	UpdateResidentContact(ctx context.Context, req UpdateResidentContactStandaloneRequest) (*UpdateResidentContactResponse, error)

	// 删除
	DeleteResident(ctx context.Context, req DeleteResidentRequest) (*DeleteResidentResponse, error)

	// 密码管理
	ResetResidentPassword(ctx context.Context, req ResetResidentPasswordRequest) (*ResetResidentPasswordResponse, error)

	// 账户设置管理（统一 API）
	GetResidentAccountSettings(ctx context.Context, req GetResidentAccountSettingsRequest) (*GetResidentAccountSettingsResponse, error)
	UpdateResidentAccountSettings(ctx context.Context, req UpdateResidentAccountSettingsRequest) (*UpdateResidentAccountSettingsResponse, error)
}

// residentService 实现
type residentService struct {
	residentsRepo repository.ResidentsRepository
	db            *sql.DB
	logger        *zap.Logger
}

// NewResidentService 创建 ResidentService 实例
func NewResidentService(residentsRepo repository.ResidentsRepository, db *sql.DB, logger *zap.Logger) ResidentService {
	return &residentService{
		residentsRepo: residentsRepo,
		db:            db,
		logger:        logger,
	}
}

// getResourcePermission 查询资源权限配置（Service 层内部方法）
// 从 role_permissions 表中查询指定角色对指定资源的权限配置
//
// 注意: permission_scope 值映射:
//   - 'A' = All (no restriction) → assigned_only=false, branch_only=false
//   - 'S' = assigned_only → assigned_only=true, branch_only=false
//   - 'B' = branch_only → assigned_only=false, branch_only=true
func (s *residentService) getResourcePermission(ctx context.Context, roleCode, resourceType, permissionType string) (*PermissionCheckResult, error) {
	var permissionScope string
	err := s.db.QueryRowContext(ctx,
		`SELECT permission_scope
		 FROM role_permissions
		 WHERE tenant_id = $1 
		   AND role_code = $2 
		   AND resource_type = $3 
		   AND permission_type = $4
		 LIMIT 1`,
		SystemTenantID, roleCode, resourceType, permissionType,
	).Scan(&permissionScope)

	if err == sql.ErrNoRows {
		// 记录不存在：返回最严格的权限（安全默认值）
		return &PermissionCheckResult{AssignedOnly: true, BranchOnly: true}, nil
	}
	if err != nil {
		return nil, err
	}

	// 将 permission_scope 转换为 assigned_only 和 branch_only 标志
	var assignedOnly, branchOnly bool
	switch permissionScope {
	case "A":
		// All (no restriction)
		assignedOnly = false
		branchOnly = false
	case "S":
		// assigned_only
		assignedOnly = true
		branchOnly = false
	case "B":
		// branch_only
		assignedOnly = false
		branchOnly = true
	default:
		// 未知值，返回最严格的权限（安全默认值）
		assignedOnly = true
		branchOnly = true
	}

	return &PermissionCheckResult{AssignedOnly: assignedOnly, BranchOnly: branchOnly}, nil
}

// getUserBranchIDs 查询用户所属的 branch_id 列表（Service 层内部方法，向后兼容）
// 注意：UserBranchInfo 在 user_service.go 中定义，两个 service 共享使用
// 从 user_branches 表查询用户关联的所有院区 ID
// 返回：
//   - branchIDs: 用户所属的 branch_id 列表（可能为空）
//   - hasBranches: 用户是否有关联的院区（false 表示可以访问所有院区或 NULL 院区）
func (s *residentService) getUserBranchIDs(ctx context.Context, tenantID, userID string) (branchIDs []string, hasBranches bool, err error) {
	branches, hasBranches, err := s.getUserBranches(ctx, tenantID, userID)
	if err != nil {
		return nil, false, err
	}
	ids := make([]string, 0, len(branches))
	for _, branch := range branches {
		ids = append(ids, branch.BranchID)
	}
	return ids, hasBranches, nil
}

// getUserBranches 查询用户所属的院区信息（Service 层内部方法）
// 从 user_branches 表 JOIN branches 表查询用户关联的所有院区（包含 branch_id 和 branch_name）
// 返回：
//   - branches: 用户所属的院区信息列表（包含 branch_id 和 branch_name，可能为空）
//   - hasBranches: 用户是否有关联的院区（false 表示可以访问所有院区或 NULL 院区）
//
// 注意：UserBranchInfo 在 user_service.go 中定义，两个 service 共享使用
func (s *residentService) getUserBranches(ctx context.Context, tenantID, userID string) (branches []UserBranchInfo, hasBranches bool, err error) {
	if tenantID == "" || userID == "" {
		return nil, false, nil
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT 
			ub.branch_id::text,
			COALESCE(b.branch_name, '') as branch_name
		 FROM user_branches ub
		 LEFT JOIN branches b ON b.branch_id = ub.branch_id
		 WHERE ub.tenant_id = $1 AND ub.user_id::text = $2
		 ORDER BY b.branch_name ASC`,
		tenantID, userID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false, nil // 用户没有关联任何院区
		}
		return nil, false, fmt.Errorf("failed to query user branches: %w", err)
	}
	defer rows.Close()

	var branchList []UserBranchInfo
	for rows.Next() {
		var branchID, branchName sql.NullString
		if err := rows.Scan(&branchID, &branchName); err != nil {
			return nil, false, fmt.Errorf("failed to scan branch info: %w", err)
		}
		if branchID.Valid && branchID.String != "" {
			// 使用 user_service.go 中定义的 UserBranchInfo
			branchList = append(branchList, UserBranchInfo{
				BranchID:   branchID.String,
				BranchName: branchName.String, // 如果 branch_name 为 NULL，返回空字符串
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("failed to iterate user branches: %w", err)
	}

	return branchList, len(branchList) > 0, nil
}

// ============================================
// Request/Response DTOs
// ============================================

// ListResidentsRequest 查询住户列表请求
type ListResidentsRequest struct {
	TenantID        string // 必填
	CurrentUserID   string // 当前用户ID（用于权限过滤）
	CurrentUserType string // 当前用户类型：'resident' | 'staff'（注意：'family' 已被禁止，resident_contacts 不能登录系统）
	CurrentUserRole string // 当前用户角色（仅 staff 需要）

	// 权限检查结果（由 Handler 层传入）
	PermissionCheck *PermissionCheckResult // 权限检查结果（仅 staff 需要）

	// 过滤条件
	Search       string // 搜索关键词（nickname, unit_name）
	Status       string // 状态过滤
	ServiceLevel string // 护理级别过滤

	// 分页
	Page     int // 页码，默认 1
	PageSize int // 每页数量，默认 20
}

// PermissionCheckResult 权限检查结果（Service 层内部使用，不信任外部传入）
type PermissionCheckResult struct {
	AssignedOnly bool // 是否仅限分配的资源
	BranchOnly   bool // 是否仅限同一 Branch 的资源
	// 注意：权限检查应通过 branch_id 进行，不再使用 branch_tag
	// Service 层会自己查询用户的 branch_id（通过 user_branches 表）
}

// ListResidentsResponse 查询住户列表响应
type ListResidentsResponse struct {
	Items []*ResidentListItemDTO // 住户列表
	Total int                    // 总数量
}

// ResidentListItemDTO 住户列表项 DTO
type ResidentListItemDTO struct {
	ResidentID      string  `json:"resident_id"`
	TenantID        string  `json:"tenant_id"`
	ResidentAccount *string `json:"resident_account,omitempty"`
	Nickname        string  `json:"nickname"`
	Status          string  `json:"status"`
	ServiceLevel    *string `json:"service_level,omitempty"`
	AdmissionDate   *int64  `json:"admission_date,omitempty"` // Unix timestamp
	DischargeDate   *int64  `json:"discharge_date,omitempty"` // Unix timestamp
	UnitID          *string `json:"unit_id,omitempty"`
	UnitName        *string `json:"unit_name,omitempty"`
	BuildingName    *string `json:"building_name,omitempty"`  // 从 units.building_name 获取
	BranchID        *string `json:"branch_id,omitempty"`      // 从 residents.branch_id 获取
	BranchName      *string `json:"branch_name,omitempty"`    // 从 branches.branch_name 获取（通过 residents.branch_id）
	IsSharedUnit    *bool   `json:"is_shared_unit,omitempty"` // 从 units.is_shared_unit 获取（原 is_multi_person_room），未绑定 unit 时为 nil
	FacilityType    *string `json:"facility_type,omitempty"`  // Share / Public / VIP（由 units.is_shared_unit + is_public 推导）
	RoomID          *string `json:"room_id,omitempty"`
	RoomName        *string `json:"room_name,omitempty"`
	BedID           *string `json:"bed_id,omitempty"`
	BedName         *string `json:"bed_name,omitempty"`
	IsAccessEnabled bool    `json:"is_access_enabled"`
	// Note: 列表不需要 Note 字段
}

// residentListFacilityType 列表用：Share（共享）/ Public（公共）/ VIP（独享）
func residentListFacilityType(shared, public sql.NullBool) *string {
	if !shared.Valid && !public.Valid {
		return nil
	}
	if shared.Valid && shared.Bool {
		s := "Share"
		return &s
	}
	if public.Valid && public.Bool {
		s := "Public"
		return &s
	}
	s := "VIP"
	return &s
}

// GetResidentRequest 获取住户详情请求
type GetResidentRequest struct {
	TenantID        string // 必填
	ResidentID      string // 住户ID（或 contact_id）
	CurrentUserID   string // 当前用户ID
	CurrentUserType string // 当前用户类型
	CurrentUserRole string // 当前用户角色（仅 staff 需要）

	// 权限检查结果（由 Handler 层传入）
	PermissionCheck *PermissionCheckResult // 权限检查结果（仅 staff 需要）

	// 可选数据
	IncludePHI      bool // 是否包含 PHI 数据
	IncludeContacts bool // 是否包含联系人数据
}

// ResidentCaregiverDTO 住户护理人员分配 DTO
type ResidentCaregiverDTO struct {
	UserList  []UserDTO `json:"user_list,omitempty"`  // 用户详细信息列表（包含 user_id, nickname, user_account 等）
	GroupList []string  `json:"group_list,omitempty"` // JSONB array -> []string (tag_name 列表)
}

// GetResidentResponse 获取住户详情响应
type GetResidentResponse struct {
	Resident   *ResidentDetailDTO    `json:"resident"`
	PHI        *ResidentPHIDTO       `json:"phi,omitempty"`
	Contacts   []*ResidentContactDTO `json:"contacts,omitempty"`
	Caregivers *ResidentCaregiverDTO `json:"caregivers,omitempty"` // 从 resident_caregivers 表获取
}

// ResidentDetailDTO 住户详情 DTO
type ResidentDetailDTO struct {
	ResidentID      string  `json:"resident_id"`
	TenantID        string  `json:"tenant_id"`
	ResidentAccount *string `json:"resident_account,omitempty"`
	Nickname        string  `json:"nickname"`
	Status          string  `json:"status"`
	ServiceLevel    *string `json:"service_level,omitempty"`
	AdmissionDate   *int64  `json:"admission_date,omitempty"` // Unix timestamp
	DischargeDate   *int64  `json:"discharge_date,omitempty"` // Unix timestamp
	UnitID          *string `json:"unit_id,omitempty"`
	UnitName        *string `json:"unit_name,omitempty"`
	BuildingName    *string `json:"building_name,omitempty"`  // 从 units.building_name 获取
	BranchID        *string `json:"branch_id,omitempty"`      // 从 residents.branch_id 获取
	BranchName      *string `json:"branch_name,omitempty"`    // 从 branches.branch_name 获取（通过 residents.branch_id）
	IsSharedUnit    *bool   `json:"is_shared_unit,omitempty"` // 从 units.is_shared_unit 获取（原 is_multi_person_room），未绑定 unit 时为 nil
	RoomID          *string `json:"room_id,omitempty"`
	RoomName        *string `json:"room_name,omitempty"`
	BedID           *string `json:"bed_id,omitempty"`
	BedName         *string `json:"bed_name,omitempty"`
	IsAccessEnabled bool    `json:"is_access_enabled"`
	Note            *string `json:"note,omitempty"`
}

// ResidentPHIDTO 住户 PHI 数据 DTO
type ResidentPHIDTO struct {
	PhiID                 string   `json:"phi_id"`
	FirstName             *string  `json:"first_name,omitempty"`
	LastName              *string  `json:"last_name,omitempty"`
	Gender                *string  `json:"gender,omitempty"`
	DateOfBirth           *int64   `json:"date_of_birth,omitempty"` // Unix timestamp
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
	PlusCode              *string  `json:"plus_code,omitempty"`
}

// ResidentContactDTO 住户联系人 DTO
type ResidentContactDTO struct {
	ContactID          string  `json:"contact_id"`
	Slot               string  `json:"slot"`
	IsEnabled          bool    `json:"is_enabled"`
	Relationship       *string `json:"relationship,omitempty"`
	ContactFirstName   *string `json:"contact_first_name,omitempty"`
	ContactLastName    *string `json:"contact_last_name,omitempty"`
	ContactPhone       *string `json:"contact_phone,omitempty"`
	ContactEmail       *string `json:"contact_email,omitempty"`
	ReceiveSMS         bool    `json:"receive_sms"`
	ReceiveEmail       bool    `json:"receive_email"`
	ContactFamilyTag   *string `json:"contact_family_tag,omitempty"`
	IsEmergencyContact bool    `json:"is_emergency_contact"`
}

// CreateResidentInherentAttributes Resident 固有属性创建结构体
// 包含3张表：residents, resident_phi, resident_contacts
type CreateResidentInherentAttributes struct {
	// ========== residents 表字段 ==========
	// 必填字段
	ResidentAccount string // 住户账号（必填）
	Nickname        string // 昵称（必填）

	// 可选字段
	PasswordHash    string          // password_hash hex 字符串（可选，前端已计算）
	Status          string          // 状态（可选，默认 "active"）
	ServiceLevel    string          // 护理级别（可选）
	AdmissionDate   *int64          // 入院日期（Unix 时间戳，可选）
	DischargeDate   *int64          // 出院日期（Unix 时间戳，可选）
	BranchID        string          // 院区ID（可选）
	IsAccessEnabled bool            // 是否允许查看状态（可选，默认 false）
	Note            string          // 备注（可选）
	PhoneHash       string          // phone_hash hex 字符串（可选，前端已计算）
	EmailHash       string          // email_hash hex 字符串（可选，前端已计算）
	Metadata        json.RawMessage // metadata JSONB（可选）

	// ========== resident_phi 表字段 ==========
	PHI *CreateResidentPHIRequest // PHI 数据（可选）

	// ========== resident_contacts 表字段 ==========
	Contacts []*CreateResidentContactRequest // 联系人列表（可选）
}

// CreateResidentUnitRelation Resident 与 Unit 的关系创建结构体
// 业务属性：位置分配（虽然存储在 residents 表中，但属于业务分配属性）
type CreateResidentUnitRelation struct {
	UnitID string // 单元ID（可选）
	RoomID string // 房间ID（可选）
	BedID  string // 床位ID（可选）
	// 业务规则：bed → room → unit（如果指定 bed_id，则必须同时指定 room_id 和 unit_id）
}

// CreateResidentCaregiverRelation Resident 与 Caregiver 的关系创建结构体
// 业务属性：护理人员分配（存储在 resident_caregivers 表中）
type CreateResidentCaregiverRelation struct {
	UserList  []string // 用户ID列表（可选，JSONB array）
	GroupList []string // 用户组标签列表（可选，JSONB array，用于匹配 users.user_tags）
	// 说明：
	//   - 每个租户+住户最多一条记录（UNIQUE(tenant_id, resident_id)）
	//   - 如果 user_list 和 group_list 都为空，使用默认告警路由规则（由应用层处理）
}

// CreateResidentRequest 创建住户请求
// 包含3部分：Resident 固有属性 + 与 Unit 的关系 + 与 Caregiver 的关系
type CreateResidentRequest struct {
	TenantID        string                 // 必填
	CurrentUserID   string                 // 当前用户ID
	CurrentUserRole string                 // 当前用户角色
	PermissionCheck *PermissionCheckResult // 权限检查结果

	// 注意：AvailableBranches 不应由 Handler 传递，Service 层会自己从数据库查询用户的 branch 信息
	// 这是用户本身的属性，不能信任前端传递的值

	// Resident 固有属性（3张表：residents, resident_phi, resident_contacts）
	InherentAttributes *CreateResidentInherentAttributes

	// Resident 与 Unit 的关系（位置分配）
	UnitRelation *CreateResidentUnitRelation

	// Resident 与 Caregiver 的关系（护理人员分配）
	CaregiverRelation *CreateResidentCaregiverRelation
}

// CreateResidentPHIRequest 创建住户 PHI 请求
type CreateResidentPHIRequest struct {
	FirstName             string // 必填（创建时）
	LastName              string // 可选
	Gender                string // 可选
	DateOfBirth           *int64 // Unix timestamp
	ResidentPhone         string // 明文（可选保存）
	ResidentEmail         string // 明文（可选保存）
	SavePhone             bool   // 是否保存明文 phone
	SaveEmail             bool   // 是否保存明文 email
	WeightLb              *float64
	HeightFt              *float64
	HeightIn              *float64
	MobilityLevel         *int
	TremorStatus          string
	MobilityAid           string
	ADLAssistance         string
	CommStatus            string
	HasHypertension       *bool
	HasHyperlipaemia      *bool
	HasHyperglycaemia     *bool
	HasStrokeHistory      *bool
	HasParalysis          *bool
	HasAlzheimer          *bool
	MedicalHistory        string
	HomeAddressStreet     string
	HomeAddressCity       string
	HomeAddressState      string
	HomeAddressPostalCode string
	PlusCode              string
}

// CreateResidentContactRequest 创建住户联系人请求
// 注意：联系人不登录系统，仅作为住户的属性
type CreateResidentContactRequest struct {
	Slot             string          // 槽位 'A', 'B', 'C', 'D', 'E'（必填）
	IsEnabled        bool            // 是否启用该联系人（可选，默认 false）
	Relationship     string          // 关系（可选）：Child/Spouse/Friend/Caregiver/Other
	ContactFirstName string          // 联系人名（可选）
	ContactLastName  string          // 联系人姓（可选）
	ContactPhone     string          // 联系人电话（可选），明文保存
	ContactEmail     string          // 联系人邮箱（可选），明文保存
	ContactPhoneHash string          // 联系人电话 hash（可选，前端计算的 hex 字符串，用于搜索）
	ContactEmailHash string          // 联系人邮箱 hash（可选，前端计算的 hex 字符串，用于搜索）
	ReceiveSMS       bool            // 是否接收短信（可选，默认 false）
	ReceiveEmail     bool            // 是否接收邮件（可选，默认 false）
	AlertTimeWindow  json.RawMessage // 告警接收时间窗口 JSONB（可选）
}

// CreateResidentResponse 创建住户响应
type CreateResidentResponse struct {
	ResidentID string // 创建的住户ID
}

// UpdateResidentInherentAttributes Resident 固有属性更新结构体
// 包含3张表：residents, resident_phi, resident_contacts
type UpdateResidentInherentAttributes struct {
	// ========== residents 表字段（使用 domain.UpdateX 类型）==========
	ResidentAccount *domain.UpdateString // 住户账号（可选更新）
	Nickname        *domain.UpdateString // 昵称（可选更新）
	PasswordHash    *domain.UpdateBytes  // password_hash（可选更新）
	Status          *domain.UpdateString // 状态（可选更新）
	ServiceLevel    *domain.UpdateString // 护理级别（可选更新）
	AdmissionDate   *domain.UpdateTime   // 入院日期（可选更新）
	DischargeDate   *domain.UpdateTime   // 出院日期（可选更新）
	BranchID        *domain.UpdateString // 院区ID（可选更新）
	IsAccessEnabled *domain.UpdateBool   // 是否允许查看状态（可选更新）
	Note            *domain.UpdateString // 备注（可选更新）
	Phone           *domain.UpdateString // phone（可选更新）
	Email           *domain.UpdateString // email（可选更新）
	PhoneHash       *domain.UpdateBytes  // phone_hash（可选更新）
	EmailHash       *domain.UpdateBytes  // email_hash（可选更新）
	Metadata        *domain.UpdateJSON   // metadata JSONB（可选更新）

	// ========== resident_phi 表字段 ==========
	PHI *UpdateResidentPHIRequest // PHI 数据（可选更新）

	// ========== resident_contacts 表字段 ==========
	Contacts []*UpdateResidentContactRequest // 联系人列表（可选更新，每个联系人通过 slot 标识）
}

// UpdateResidentPHIRequest 更新住户 PHI 请求
// 使用 domain.UpdateX 类型来明确表示更新意图
type UpdateResidentPHIRequest struct {
	FirstName             *domain.UpdateString  // 名（可选更新）
	LastName              *domain.UpdateString  // 姓（可选更新）
	Gender                *domain.UpdateString  // 性别（可选更新）
	DateOfBirth           *domain.UpdateTime    // 出生日期（可选更新）
	ResidentPhone         *domain.UpdateString  // 住户电话（可选更新）
	ResidentEmail         *domain.UpdateString  // 住户邮箱（可选更新）
	SavePhone             *domain.UpdateBool    // 是否保存电话（可选更新）
	SaveEmail             *domain.UpdateBool    // 是否保存邮箱（可选更新）
	WeightLb              *domain.UpdateFloat64 // 体重（lb，可选更新）
	HeightFt              *domain.UpdateFloat64 // 身高：feet（可选更新）
	HeightIn              *domain.UpdateFloat64 // 身高：inches（可选更新）
	MobilityLevel         *domain.UpdateInt     // 行动能力（可选更新，0-5）
	TremorStatus          *domain.UpdateString  // 颤抖状态（可选更新）
	MobilityAid           *domain.UpdateString  // 行走辅助（可选更新）
	ADLAssistance         *domain.UpdateString  // 日常活动协助（可选更新）
	CommStatus            *domain.UpdateString  // 沟通状态（可选更新）
	HasHypertension       *domain.UpdateBool    // 高血压（可选更新）
	HasHyperlipaemia      *domain.UpdateBool    // 高血脂（可选更新）
	HasHyperglycaemia     *domain.UpdateBool    // 高血糖/糖尿病（可选更新）
	HasStrokeHistory      *domain.UpdateBool    // 既往脑卒中史（可选更新）
	HasParalysis          *domain.UpdateBool    // 肢体瘫痪/偏瘫（可选更新）
	HasAlzheimer          *domain.UpdateBool    // 阿尔茨海默病/痴呆（可选更新）
	MedicalHistory        *domain.UpdateString  // 其他病史说明（可选更新）
	HomeAddressStreet     *domain.UpdateString  // 街道地址（可选更新）
	HomeAddressCity       *domain.UpdateString  // 城市（可选更新）
	HomeAddressState      *domain.UpdateString  // 州/省（可选更新）
	HomeAddressPostalCode *domain.UpdateString  // 邮编（可选更新）
	PlusCode              *domain.UpdateString  // Google Plus Code（可选更新）
}

// UpdateResidentContactRequest 更新住户联系人请求
// 使用 domain.UpdateX 类型来明确表示更新意图
// 注意：联系人的主键是 (resident_id, slot)，所以 slot 字段必须提供（用于标识要更新哪个联系人）
type UpdateResidentContactRequest struct {
	Slot             string               // 槽位 'A', 'B', 'C', 'D', 'E'（必填，用于标识联系人）
	IsEnabled        *domain.UpdateBool   // 是否启用（可选更新）
	Relationship     *domain.UpdateString // 关系（可选更新）：Child/Spouse/Friend/Caregiver/Other
	ContactFirstName *domain.UpdateString // 联系人名（可选更新）
	ContactLastName  *domain.UpdateString // 联系人姓（可选更新）
	ContactPhone     *domain.UpdateString // 联系人电话（可选更新），明文保存
	ContactEmail     *domain.UpdateString // 联系人邮箱（可选更新），明文保存
	ContactPhoneHash *domain.UpdateBytes  // 联系人电话 hash（可选更新，用于搜索）
	ContactEmailHash *domain.UpdateBytes  // 联系人邮箱 hash（可选更新，用于搜索）
	ReceiveSMS       *domain.UpdateBool   // 是否接收短信（可选更新）
	ReceiveEmail     *domain.UpdateBool   // 是否接收邮件（可选更新）
	AlertTimeWindow  *domain.UpdateJSON   // 告警接收时间窗口（可选更新，JSONB）
}

// UpdateResidentUnitRelation Resident 与 Unit 的关系更新结构体
// 业务属性：位置分配（虽然存储在 residents 表中，但属于业务分配属性）
type UpdateResidentUnitRelation struct {
	UnitID *domain.UpdateString // 单元ID（可选更新）
	RoomID *domain.UpdateString // 房间ID（可选更新）
	BedID  *domain.UpdateString // 床位ID（可选更新）
	// 业务规则：bed → room → unit（如果指定 bed_id，则必须同时指定 room_id 和 unit_id）
}

// UpdateResidentCaregiverRelation Resident 与 Caregiver 的关系更新结构体
// 业务属性：护理人员分配（存储在 resident_caregivers 表中）
type UpdateResidentCaregiverRelation struct {
	UserList  *domain.UpdateJSON // 用户ID列表（可选更新，JSONB array）
	GroupList *domain.UpdateJSON // 用户组标签列表（可选更新，JSONB array，用于匹配 users.user_tags）
	// 说明：
	//   - 每个租户+住户最多一条记录（UNIQUE(tenant_id, resident_id)）
	//   - 如果 user_list 和 group_list 都为空，使用默认告警路由规则（由应用层处理）
}

// UpdateResidentRequest 更新住户请求
// 包含3部分：Resident 固有属性 + 与 Unit 的关系 + 与 Caregiver 的关系
type UpdateResidentRequest struct {
	TenantID        string                 // 必填
	ResidentID      string                 // 必填（要更新的住户ID）
	CurrentUserID   string                 // 当前用户ID
	CurrentUserRole string                 // 当前用户角色
	PermissionCheck *PermissionCheckResult // 权限检查结果

	// Resident 固有属性（3张表：residents, resident_phi, resident_contacts）
	InherentAttributes *UpdateResidentInherentAttributes

	// Resident 与 Unit 的关系（位置分配）
	UnitRelation *UpdateResidentUnitRelation

	// Resident 与 Caregiver 的关系（护理人员分配）
	CaregiverRelation *UpdateResidentCaregiverRelation
}

// UpdateResidentResponse 更新住户响应
type UpdateResidentResponse struct {
	Success bool
}

// UpdateResidentContactStandaloneRequest 更新住户联系人请求（独立接口）
// Deprecated: 此接口用于独立的 UpdateResidentContact 方法
// 建议使用 UpdateResident 方法中的 Contacts 字段来更新联系人
type UpdateResidentContactStandaloneRequest struct {
	TenantID        string // 必填
	ResidentID      string // 必填
	Slot            string // 必填：通过 resident_id + slot 定位 contact
	CurrentUserID   string // 当前用户ID
	CurrentUserType string // 当前用户类型
	CurrentUserRole string // 当前用户角色（Service 层自己查询权限）

	// 可更新字段（使用指针表示可选，nil 表示不更新，空字符串表示删除）
	IsEnabled        *bool
	Relationship     *string
	ContactFirstName *string
	ContactLastName  *string
	ContactPhone     *string // nil=不更新, ""=删除, 有值=更新
	ContactEmail     *string // nil=不更新, ""=删除, 有值=更新
	ReceiveSMS       *bool
	ReceiveEmail     *bool
	PhoneHash        *string // phone_hash (hex string, nil=不更新, ""=删除, 有值=更新)
	EmailHash        *string // email_hash (hex string, nil=不更新, ""=删除, 有值=更新)
	PasswordHash     *string // password_hash (hex string, nil=不更新, ""=删除, 有值=更新)
}

// UpdateResidentContactResponse 更新住户联系人响应
type UpdateResidentContactResponse struct {
	Success bool
}

// DeleteResidentRequest 删除住户请求
type DeleteResidentRequest struct {
	TenantID        string // 必填
	ResidentID      string // 必填
	CurrentUserID   string // 当前用户ID
	CurrentUserType string // 当前用户类型
	CurrentUserRole string // 当前用户角色

	// 权限检查结果（由 Handler 层传入）
	PermissionCheck *PermissionCheckResult // 权限检查结果（仅 staff 需要）
}

// DeleteResidentResponse 删除住户响应
type DeleteResidentResponse struct {
	Success bool
}

// ResetResidentPasswordRequest 重置住户密码请求
type ResetResidentPasswordRequest struct {
	TenantID        string                 // 必填
	ResidentID      string                 // 必填
	CurrentUserID   string                 // 当前用户ID
	CurrentUserType string                 // 当前用户类型
	CurrentUserRole string                 // 当前用户角色
	PermissionCheck *PermissionCheckResult // 权限检查结果
	NewPassword     string                 // 新密码（可选，默认生成）
}

// ResetResidentPasswordResponse 重置住户密码响应
type ResetResidentPasswordResponse struct {
	Success     bool
	NewPassword string // 生成的新密码
}

// GetResidentAccountSettingsRequest 获取住户/联系人账户设置请求
type GetResidentAccountSettingsRequest struct {
	TenantID        string // 必填
	ResidentID      string // 住户ID 或 contact_id
	CurrentUserID   string // 当前用户ID
	CurrentUserType string // 当前用户类型
	CurrentUserRole string // 当前用户角色
}

// GetResidentAccountSettingsResponse 获取住户/联系人账户设置响应
type GetResidentAccountSettingsResponse struct {
	ResidentAccount *string // 住户账号（resident）或关联的 resident_account（contact）
	Nickname        string  // 昵称
	Email           *string // 邮箱（可为空）
	Phone           *string // 电话（可为空）
	IsContact       bool    // 是否是 contact（true=contact, false=resident）
	SaveEmail       bool    // 是否保存 email（仅 resident 需要，contact 总是保存）
	SavePhone       bool    // 是否保存 phone（仅 resident 需要，contact 总是保存）
}

// UpdateResidentAccountSettingsRequest 更新住户/联系人账户设置请求（统一 API）
type UpdateResidentAccountSettingsRequest struct {
	TenantID        string  // 必填
	ResidentID      string  // 住户ID 或 contact_id
	CurrentUserID   string  // 当前用户ID
	CurrentUserType string  // 当前用户类型
	CurrentUserRole string  // 当前用户角色
	PasswordHash    *string // 可选：密码 hash（nil 表示不更新）
	Email           *string // 可选：邮箱（nil 表示不更新，空字符串表示删除）
	EmailHash       *string // 可选：邮箱 hash（前端计算的 hash）
	Phone           *string // 可选：电话（nil 表示不更新，空字符串表示删除）
	PhoneHash       *string // 可选：电话 hash（前端计算的 hash）
	SaveEmail       *bool   // 可选：是否保存 email 明文（仅 resident 需要，contact 总是保存）
	SavePhone       *bool   // 可选：是否保存 phone 明文（仅 resident 需要，contact 总是保存）
}

// UpdateResidentAccountSettingsResponse 更新住户/联系人账户设置响应
type UpdateResidentAccountSettingsResponse struct {
	Success bool   // 是否成功
	Message string // 消息（可选，用于错误详情）
}

// ============================================
// 辅助函数
// ============================================

// timeToUnixTimestamp 将 time.Time 转换为 Unix timestamp（秒）
func timeToUnixTimestamp(t *time.Time) *int64 {
	if t == nil {
		return nil
	}
	ts := t.Unix()
	return &ts
}

// unixTimestampToTime 将 Unix timestamp（秒）转换为 time.Time
func unixTimestampToTime(ts *int64) *time.Time {
	if ts == nil {
		return nil
	}
	t := time.Unix(*ts, 0)
	return &t
}

// HashAccount, HashPassword, sha256Hex 已在 user_service.go 中定义，这里不再重复定义

// sha256Hash 计算 SHA256 hash 并返回 []byte（用于与 domain.UpdateBytes 配合）
func sha256Hash(data string) []byte {
	h := sha256.Sum256([]byte(strings.ToLower(data)))
	return h[:]
}

// equalBytes 比较两个 []byte 是否相等
func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ============================================
// Service 方法实现
// ============================================

// ============================================
// Service 方法实现（待实现）
// ============================================
// 注意：ResidentService 的实现非常复杂（约 3000+ 行代码），需要：
// 1. 权限过滤（AssignedOnly, BranchOnly）
// 2. JOIN 查询（units, rooms, beds）
// 3. 业务规则验证（discharge_date, unit_id 等）
// 4. 多表操作（residents, resident_phi, resident_contacts, resident_caregivers）
// 5. 数据转换（map → domain → DTO）
//
// 当前仅完成接口定义，具体实现待后续完善。
// 参考：internal/http/admin_residents_handlers.go (3032 行)

// ListResidents 查询住户列表
func (s *residentService) ListResidents(ctx context.Context, req ListResidentsRequest) (*ListResidentsResponse, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	// 1. 参数验证和默认值
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	// 2. 构建基础查询（JOIN units, buildings, rooms, beds, branches, resident_phi, resident_contacts）
	args := []any{req.TenantID}
	q := `SELECT DISTINCT r.resident_id::text, r.tenant_id::text, r.resident_account, r.nickname,
	             r.status, r.service_level, r.admission_date, r.discharge_date,
	             r.unit_id::text, r.room_id::text, r.bed_id::text,
	             r.branch_id::text,
	             u.unit_name,
	             bld.building_name,
	             u.is_shared_unit,
	             u.is_public,
	             br.branch_name,
	             rm.room_name,
	             b.bed_name,
	             r.is_access_enabled,
	             COALESCE(r.resident_account, '') as resident_account_for_sort,
	             r.resident_id::text as resident_id_for_sort
	      FROM residents r
	      LEFT JOIN units u ON u.unit_id = r.unit_id
	      LEFT JOIN buildings bld ON bld.building_id = u.building_id
	      LEFT JOIN rooms rm ON rm.room_id = r.room_id
	      LEFT JOIN beds b ON b.bed_id = r.bed_id
	      LEFT JOIN branches br ON br.branch_id = r.branch_id
	      LEFT JOIN resident_phi rp ON rp.resident_id = r.resident_id AND rp.tenant_id = r.tenant_id
	      LEFT JOIN resident_contacts rc ON rc.resident_id = r.resident_id AND rc.tenant_id = r.tenant_id`

	// 3. 权限过滤
	// 注意：resident_contacts 不能登录系统，所以 CurrentUserType 永远不会是 "family"
	// 保留 "family" 检查是为了向后兼容，但实际上只会是 "resident" 或 "staff"
	if req.CurrentUserType == "resident" || req.CurrentUserType == "family" {
		// Resident: 只能查看自己
		// 注意：虽然代码中有检查 resident_contact 的逻辑，但实际上 resident_contacts 不能登录系统
		// 所以这段代码永远不会检测到 resident_contact 登录
		var residentIDForSelf sql.NullString
		if req.CurrentUserID != "" {
			// 直接使用 CurrentUserID 作为 resident_id（因为 resident_contacts 不能登录）
			residentIDForSelf = sql.NullString{String: req.CurrentUserID, Valid: true}
		}

		if residentIDForSelf.Valid {
			args = append(args, residentIDForSelf.String)
			q += fmt.Sprintf(` WHERE r.tenant_id = $1 AND r.resident_id::text = $%d`, len(args))
		} else {
			// If resident ID not found, return empty list
			q += ` WHERE 1=0`
		}
	} else {
		// Staff login: 应用权限过滤
		q += ` WHERE r.tenant_id = $1`

		// AssignedOnly 过滤
		if req.PermissionCheck != nil && req.PermissionCheck.AssignedOnly && req.CurrentUserID != "" {
			args = append(args, req.CurrentUserID)
			q += fmt.Sprintf(` AND EXISTS (
			                      SELECT 1 FROM resident_caregivers rc
			                      WHERE rc.tenant_id = r.tenant_id
			                        AND rc.resident_id = r.resident_id
			                        AND (rc.user_list::text LIKE $%d OR rc.user_list::text LIKE $%d)
			                  )`, len(args), len(args)+1)
			args = append(args, "%\""+req.CurrentUserID+"\"%")
		}

		// BranchOnly 过滤（通过 residents.branch_id 匹配用户的 branch_id）
		// 注意：NULL、""、"default" 在 branch 表中都表示空院区/默认院区
		if req.PermissionCheck != nil && req.PermissionCheck.BranchOnly && req.CurrentUserID != "" {
			userBranchIDs, hasBranches, err := s.getUserBranchIDs(ctx, req.TenantID, req.CurrentUserID)
			if err != nil {
				return nil, fmt.Errorf("failed to get user branch IDs: %w", err)
			}

			if !hasBranches {
				// 用户没有关联任何院区：可以访问所有 NULL/默认院区的住户
				// NULL/默认院区：branch_id IS NULL，或 branch_name IS NULL，或 ''，或 'default'
				q += ` AND (
					r.branch_id IS NULL 
					OR br.branch_name IS NULL 
					OR br.branch_name = '' 
					OR br.branch_name = 'default'
				)`
			} else if len(userBranchIDs) == 1 {
				// 用户只属于一个院区：只查看该院区的住户
				args = append(args, userBranchIDs[0])
				q += fmt.Sprintf(` AND r.branch_id::text = $%d`, len(args))
			} else {
				// 用户属于多个院区：查看所有关联院区的住户（使用 IN 查询）
				placeholders := make([]string, len(userBranchIDs))
				for i, branchID := range userBranchIDs {
					args = append(args, branchID)
					placeholders[i] = fmt.Sprintf("$%d", len(args))
				}
				q += fmt.Sprintf(` AND r.branch_id::text IN (%s)`, strings.Join(placeholders, ", "))
			}
		}
	}

	// 4. 搜索和过滤
	argIdx := len(args) + 1
	if req.Search != "" {
		searchPattern := "%" + strings.ToLower(req.Search) + "%"
		// 搜索支持字段（共3个）：
		// 1. r.nickname - 住户昵称（使用 ILIKE 模糊查询）
		// 2. r.email - residents 表中的 email（通过 r.email_hash 查询，因为是 PHI，加密存储）
		// 3. r.phone - residents 表中的 phone（通过 r.phone_hash 查询，因为是 PHI，加密存储）
		// 注意：first_name, last_name 需要加密存储，字母大小写 hash 不方便，不支持搜索
		// 注意：rp.resident_email 和 rp.resident_phone 是加密存储的，但 resident_phi 表中没有 hash 字段，无法搜索
		// 计算搜索词的 hash（用于匹配 email_hash 和 phone_hash）
		searchLower := strings.ToLower(strings.TrimSpace(req.Search))
		searchHashHex := HashAccount(searchLower)
		searchHash, err := hex.DecodeString(searchHashHex)
		if err != nil || len(searchHash) == 0 {
			searchHash = []byte{} // 空 hash，不会匹配任何记录
		}
		args = append(args, searchPattern, searchHash, searchHash)
		q += fmt.Sprintf(` AND (
			r.nickname ILIKE $%d OR
			-- 对于 email 和 phone，使用 hash 查询（因为它们是 PHI，加密存储）
			r.email_hash = $%d OR
			r.phone_hash = $%d
			-- 注意：rp.resident_email 和 rp.resident_phone 是加密存储的，但 resident_phi 表中没有 hash 字段，无法搜索
		)`, argIdx, argIdx+1, argIdx+2)
		argIdx += 3
	}
	if req.Status != "" {
		args = append(args, req.Status)
		q += fmt.Sprintf(` AND r.status = $%d`, argIdx)
		argIdx++
	}
	if req.ServiceLevel != "" {
		args = append(args, req.ServiceLevel)
		q += fmt.Sprintf(` AND r.service_level = $%d`, argIdx)
		argIdx++
	}

	// 5. 排序和分页（按 account 排序）
	// 注意：在 SELECT DISTINCT 中，ORDER BY 的表达式必须出现在 SELECT 列表中
	q += ` ORDER BY resident_account_for_sort ASC, resident_id_for_sort ASC`
	args = append(args, pageSize, (page-1)*pageSize)
	q += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)

	// 6. 执行查询
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		s.logger.Error("ListResidents query failed",
			zap.String("tenant_id", req.TenantID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to list residents: %w", err)
	}
	defer rows.Close()

	// 7. 扫描结果
	items := []*ResidentListItemDTO{}
	for rows.Next() {
		var residentID, tid, residentAccount, nickname, status, serviceLevel sql.NullString
		var admissionDate, dischargeDate sql.NullTime
		var unitID, roomID, bedID, branchID sql.NullString
		var unitName, buildingName, branchName sql.NullString
		var isSharedUnit, isPublicUnit sql.NullBool // LEFT JOIN units 时可能为 NULL
		var roomName, bedName sql.NullString
		var canViewStatus bool

		var residentAccountForSort, residentIDForSort sql.NullString // 用于排序的字段，不需要使用
		if err := rows.Scan(
			&residentID, &tid, &residentAccount, &nickname,
			&status, &serviceLevel, &admissionDate, &dischargeDate,
			&unitID, &roomID, &bedID,
			&branchID,
			&unitName, &buildingName, &isSharedUnit, &isPublicUnit,
			&branchName,
			&roomName, &bedName, &canViewStatus,
			&residentAccountForSort, // 扫描排序字段（SELECT DISTINCT 要求 ORDER BY 字段必须在 SELECT 中）
			&residentIDForSort,      // 扫描排序字段（SELECT DISTINCT 要求 ORDER BY 字段必须在 SELECT 中）
		); err != nil {
			s.logger.Error("ListResidents scan failed",
				zap.String("tenant_id", req.TenantID),
				zap.Error(err),
			)
			return nil, fmt.Errorf("failed to scan resident: %w", err)
		}

		item := &ResidentListItemDTO{
			ResidentID:      residentID.String,
			TenantID:        tid.String,
			Nickname:        nickname.String,
			Status:          status.String,
			IsAccessEnabled: canViewStatus,
		}

		// 只有当 isSharedUnit 有效时才赋值，未绑定 unit 时为 nil
		if isSharedUnit.Valid {
			item.IsSharedUnit = &isSharedUnit.Bool
		}
		item.FacilityType = residentListFacilityType(isSharedUnit, isPublicUnit)

		if residentAccount.Valid {
			item.ResidentAccount = &residentAccount.String
		}
		if serviceLevel.Valid {
			item.ServiceLevel = &serviceLevel.String
		}
		if admissionDate.Valid {
			ts := admissionDate.Time.Unix()
			item.AdmissionDate = &ts
		}
		if dischargeDate.Valid {
			ts := dischargeDate.Time.Unix()
			item.DischargeDate = &ts
		}
		if unitID.Valid {
			item.UnitID = &unitID.String
		}
		if unitName.Valid && unitName.String != "" {
			item.UnitName = &unitName.String
		}
		if buildingName.Valid && buildingName.String != "" {
			item.BuildingName = &buildingName.String
		}
		if branchID.Valid && branchID.String != "" {
			item.BranchID = &branchID.String
		}
		if branchName.Valid && branchName.String != "" {
			item.BranchName = &branchName.String
		}
		if roomID.Valid {
			item.RoomID = &roomID.String
		}
		if roomName.Valid && roomName.String != "" {
			item.RoomName = &roomName.String
		}
		if bedID.Valid {
			item.BedID = &bedID.String
		}
		if bedName.Valid && bedName.String != "" {
			item.BedName = &bedName.String
		}

		items = append(items, item)
	}

	// 8. 查询总数（使用相同的 WHERE 条件，但不包含 JOIN 和分页）
	countQuery := strings.Replace(q, "SELECT DISTINCT r.resident_id::text, r.tenant_id::text, r.resident_account, r.nickname,\n\t             r.status, r.service_level, r.admission_date, r.discharge_date,\n\t             r.unit_id::text, r.room_id::text, r.bed_id::text,\n\t             r.branch_id::text,\n\t             u.unit_name,\n\t             bld.building_name,\n\t             u.is_shared_unit,\n\t             u.is_public,\n\t             br.branch_name,\n\t             rm.room_name,\n\t             b.bed_name,\n\t             r.is_access_enabled,\n\t             COALESCE(r.resident_account, '') as resident_account_for_sort,\n\t             r.resident_id::text as resident_id_for_sort\n	      FROM residents r\n	      LEFT JOIN units u ON u.unit_id = r.unit_id\n	      LEFT JOIN buildings bld ON bld.building_id = u.building_id\n	      LEFT JOIN rooms rm ON rm.room_id = r.room_id\n	      LEFT JOIN beds b ON b.bed_id = r.bed_id\n	      LEFT JOIN branches br ON br.branch_id = r.branch_id\n	      LEFT JOIN resident_phi rp ON rp.resident_id = r.resident_id AND rp.tenant_id = r.tenant_id\n	      LEFT JOIN resident_contacts rc ON rc.resident_id = r.resident_id AND rc.tenant_id = r.tenant_id", "SELECT COUNT(DISTINCT r.resident_id)\n	      FROM residents r\n	      LEFT JOIN units u ON u.unit_id = r.unit_id\n	      LEFT JOIN buildings bld ON bld.building_id = u.building_id\n	      LEFT JOIN branches br ON br.branch_id = r.branch_id\n	      LEFT JOIN resident_phi rp ON rp.resident_id = r.resident_id AND rp.tenant_id = r.tenant_id\n	      LEFT JOIN resident_contacts rc ON rc.resident_id = r.resident_id AND rc.tenant_id = r.tenant_id", 1)
	countQuery = strings.Replace(countQuery, " ORDER BY resident_account_for_sort ASC, resident_id_for_sort ASC", "", 1)
	countQuery = strings.Replace(countQuery, " ORDER BY r.nickname ASC", "", 1)
	countQuery = strings.Replace(countQuery, fmt.Sprintf(` LIMIT $%d OFFSET $%d`, argIdx, argIdx+1), "", 1)

	// 移除最后两个参数（LIMIT 和 OFFSET）
	countArgs := args[:len(args)-2]

	var total int
	if err := s.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		s.logger.Error("ListResidents count query failed",
			zap.String("tenant_id", req.TenantID),
			zap.Error(err),
		)
		// 如果总数查询失败，使用 items 长度作为 fallback
		total = len(items)
	}

	return &ListResidentsResponse{
		Items: items,
		Total: total,
	}, nil
}

// GetResident 获取住户详情
func (s *residentService) GetResident(ctx context.Context, req GetResidentRequest) (*GetResidentResponse, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.ResidentID == "" {
		return nil, fmt.Errorf("resident_id is required")
	}

	// 1. 支持通过 resident_id 或 contact_id 查询
	actualResidentID := req.ResidentID
	if s.db != nil {
		// 检查是否是 contact_id
		var foundContactID sql.NullString
		err := s.db.QueryRowContext(ctx,
			`SELECT contact_id::text FROM resident_contacts 
			 WHERE tenant_id = $1 AND contact_id::text = $2`,
			req.TenantID, req.ResidentID,
		).Scan(&foundContactID)
		if err == nil && foundContactID.Valid {
			// id is a contact_id, find the associated resident_id
			var linkedResidentID sql.NullString
			err2 := s.db.QueryRowContext(ctx,
				`SELECT resident_id::text FROM resident_contacts 
				 WHERE tenant_id = $1 AND contact_id::text = $2`,
				req.TenantID, req.ResidentID,
			).Scan(&linkedResidentID)
			if err2 == nil && linkedResidentID.Valid {
				actualResidentID = linkedResidentID.String
			} else {
				return nil, fmt.Errorf("contact not found or not linked to any resident")
			}
		}
	}

	// 2. 权限检查
	// 注意：resident_contacts 不能登录系统，所以 CurrentUserType 永远不会是 "family"
	// CurrentUserRole 也不会是 "Family"，因为只有 residents 可以登录
	// 保留这些检查是为了向后兼容，但实际上只会是 "resident" 或 "staff"
	isResident := req.CurrentUserType == "resident" || req.CurrentUserRole == "Resident"
	if isResident {
		// Resident: 只能查看自己
		// 注意：虽然代码中有检查 resident_contact 的逻辑，但实际上 resident_contacts 不能登录系统
		// 所以这段代码永远不会检测到 resident_contact 登录
		if req.CurrentUserID != "" && req.CurrentUserID != actualResidentID {
			return nil, fmt.Errorf("access denied: can only view own information")
		}
	} else {
		// Staff: 权限检查
		if req.PermissionCheck != nil {
			// AssignedOnly 检查
			if req.PermissionCheck.AssignedOnly && req.CurrentUserID != "" && s.db != nil {
				var isAssigned bool
				err := s.db.QueryRowContext(ctx,
					`SELECT EXISTS(
						SELECT 1 FROM resident_caregivers rc
						WHERE rc.tenant_id = $1
						  AND rc.resident_id::text = $2
						  AND (rc.user_list::text LIKE $3 OR rc.user_list::text LIKE $4)
					)`,
					req.TenantID, actualResidentID, req.CurrentUserID, "%\""+req.CurrentUserID+"\"%",
				).Scan(&isAssigned)
				if err != nil {
					return nil, fmt.Errorf("failed to check assignment: %w", err)
				}
				if !isAssigned {
					return nil, fmt.Errorf("permission denied: can only view assigned residents")
				}
			}

			// BranchOnly 权限检查（通过 residents.branch_id 匹配用户的 branch_id）
			// 注意：NULL、""、"default" 在 branch 表中都表示空院区/默认院区
			if req.PermissionCheck.BranchOnly && s.db != nil && req.CurrentUserID != "" {
				// 查询目标住户的 branch_id 和 branch_name
				var residentBranchID sql.NullString
				var residentBranchName sql.NullString
				err := s.db.QueryRowContext(ctx,
					`SELECT r.branch_id::text, br.branch_name
					 FROM residents r
					 LEFT JOIN branches br ON br.branch_id = r.branch_id
					 WHERE r.tenant_id = $1 AND r.resident_id::text = $2`,
					req.TenantID, actualResidentID,
				).Scan(&residentBranchID, &residentBranchName)
				if err != nil {
					if err == sql.ErrNoRows {
						return nil, fmt.Errorf("resident not found")
					}
					return nil, fmt.Errorf("failed to get resident branch info: %w", err)
				}

				// 判断住户是否属于空院区/默认院区（NULL、""、"default"）
				isNullBranch := !residentBranchID.Valid ||
					!residentBranchName.Valid ||
					residentBranchName.String == "" ||
					residentBranchName.String == "default"

				// 查询用户所属的 branch_id 列表
				userBranchIDs, hasBranches, err := s.getUserBranchIDs(ctx, req.TenantID, req.CurrentUserID)
				if err != nil {
					return nil, fmt.Errorf("failed to get user branch IDs: %w", err)
				}

				// 权限检查
				if !hasBranches {
					// 用户没有关联任何院区：只能查看空院区/默认院区的住户（NULL、""、"default"）
					if !isNullBranch {
						return nil, fmt.Errorf("permission denied: can only view residents with null/default branch (null, '', or 'default')")
					}
				} else {
					// 用户有关联院区：只能查看关联院区的住户（不能查看空院区的住户）
					if isNullBranch {
						return nil, fmt.Errorf("permission denied: can only view residents in assigned branches")
					}

					// 检查住户的 branch_id 是否在用户的 branch_id 列表中
					if !residentBranchID.Valid {
						return nil, fmt.Errorf("permission denied: can only view residents in assigned branches")
					}

					allowed := false
					for _, userBranchID := range userBranchIDs {
						if userBranchID == residentBranchID.String {
							allowed = true
							break
						}
					}

					if !allowed {
						return nil, fmt.Errorf("permission denied: can only view residents in assigned branches")
					}
				}
			}
		}
	}

	// 3. 查询住户基本信息（JOIN units, rooms, beds, branches）
	var residentID, tid, residentAccount, nickname, status, serviceLevel sql.NullString
	var admissionDate, dischargeDate sql.NullTime
	var unitID, roomID, bedID, branchID sql.NullString
	var unitName, buildingName, branchName sql.NullString
	var isSharedUnit sql.NullBool // 使用 sql.NullBool 处理 NULL 值（LEFT JOIN 可能导致 NULL）
	var roomName, bedName sql.NullString
	var note sql.NullString
	var canViewStatus bool

	err := s.db.QueryRowContext(ctx,
		`SELECT r.resident_id::text, r.tenant_id::text, r.resident_account, r.nickname,
		        r.status, r.service_level, r.admission_date, r.discharge_date,
		        r.unit_id::text, r.room_id::text, r.bed_id::text,
		        r.branch_id::text,
		        u.unit_name,
		        bld.building_name,
		        u.is_shared_unit,
		        br.branch_name,
		        rm.room_name,
		        b.bed_name,
		        r.note,
		        r.is_access_enabled
		 FROM residents r
		 LEFT JOIN units u ON u.unit_id = r.unit_id
		 LEFT JOIN buildings bld ON bld.building_id = u.building_id
		 LEFT JOIN rooms rm ON rm.room_id = r.room_id
		 LEFT JOIN beds b ON b.bed_id = r.bed_id
		 LEFT JOIN branches br ON br.branch_id = r.branch_id
		 WHERE r.tenant_id = $1 AND r.resident_id::text = $2`,
		req.TenantID, actualResidentID,
	).Scan(
		&residentID, &tid, &residentAccount, &nickname,
		&status, &serviceLevel, &admissionDate, &dischargeDate,
		&unitID, &roomID, &bedID,
		&branchID,
		&unitName, &buildingName, &isSharedUnit,
		&branchName,
		&roomName, &bedName, &note, &canViewStatus,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("resident not found")
		}
		s.logger.Error("GetResident query failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("resident_id", actualResidentID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get resident: %w", err)
	}

	// 4. 转换为 DTO
	resident := &ResidentDetailDTO{
		ResidentID:      residentID.String,
		TenantID:        tid.String,
		Nickname:        nickname.String,
		Status:          status.String,
		IsAccessEnabled: canViewStatus,
	}

	// 只有当 isSharedUnit 有效时才赋值，未绑定 unit 时为 nil
	if isSharedUnit.Valid {
		resident.IsSharedUnit = &isSharedUnit.Bool
	}

	if residentAccount.Valid {
		resident.ResidentAccount = &residentAccount.String
	}
	if serviceLevel.Valid {
		resident.ServiceLevel = &serviceLevel.String
	}
	if admissionDate.Valid {
		ts := admissionDate.Time.Unix()
		resident.AdmissionDate = &ts
	}
	if dischargeDate.Valid {
		ts := dischargeDate.Time.Unix()
		resident.DischargeDate = &ts
	}
	if unitID.Valid {
		resident.UnitID = &unitID.String
	}
	if unitName.Valid && unitName.String != "" {
		resident.UnitName = &unitName.String
	}
	if buildingName.Valid && buildingName.String != "" {
		resident.BuildingName = &buildingName.String
	}
	if branchID.Valid && branchID.String != "" {
		resident.BranchID = &branchID.String
	}
	if branchName.Valid && branchName.String != "" {
		resident.BranchName = &branchName.String
	}
	if roomID.Valid {
		resident.RoomID = &roomID.String
	}
	if roomName.Valid && roomName.String != "" {
		resident.RoomName = &roomName.String
	}
	if bedID.Valid {
		resident.BedID = &bedID.String
	}
	if bedName.Valid && bedName.String != "" {
		resident.BedName = &bedName.String
	}
	if note.Valid && note.String != "" {
		resident.Note = &note.String
	}

	// 5. 可选查询 PHI 数据
	var phi *ResidentPHIDTO
	if req.IncludePHI {
		phiData, err := s.residentsRepo.GetResidentPHI(ctx, req.TenantID, actualResidentID)
		if err == nil && phiData != nil {
			phi = domainPHIToDTO(phiData)

			// 检查 residents 表的 phone_hash/email_hash，如果存在但明文为空，设置占位符
			if s.db != nil {
				var phoneHash, emailHash sql.NullString
				err := s.db.QueryRowContext(ctx,
					`SELECT 
						CASE WHEN phone_hash IS NOT NULL THEN 'exists' ELSE NULL END as phone_hash,
						CASE WHEN email_hash IS NOT NULL THEN 'exists' ELSE NULL END as email_hash
					 FROM residents WHERE tenant_id = $1 AND resident_id = $2`,
					req.TenantID, actualResidentID,
				).Scan(&phoneHash, &emailHash)
				if err == nil {
					// 如果 phone_hash 存在但 resident_phone 为空，设置占位符
					if phoneHash.Valid && phoneHash.String == "exists" {
						if phi.ResidentPhone == nil || *phi.ResidentPhone == "" {
							placeholder := "xxx-xxx-xxxx"
							phi.ResidentPhone = &placeholder
						}
					}
					// 如果 email_hash 存在但 resident_email 为空，设置占位符
					if emailHash.Valid && emailHash.String == "exists" {
						if phi.ResidentEmail == nil || *phi.ResidentEmail == "" {
							placeholder := "***@***"
							phi.ResidentEmail = &placeholder
						}
					}
				}
			}
		}
	}

	// 6. 可选查询联系人数据
	var contacts []*ResidentContactDTO
	if req.IncludeContacts {
		contactList, err := s.residentsRepo.GetResidentContacts(ctx, req.TenantID, actualResidentID)
		if err == nil {
			for _, c := range contactList {
				contacts = append(contacts, domainContactToDTO(c))
			}
		}
	}

	// 7. 查询 Caregivers 数据（必须查询，Profile Tab 需要）
	var caregivers *ResidentCaregiverDTO
	var userListRaw, groupListRaw sql.NullString
	err = s.db.QueryRowContext(ctx,
		`SELECT user_list::text, group_list::text
		 FROM resident_caregivers
		 WHERE tenant_id = $1 AND resident_id::text = $2`,
		req.TenantID, actualResidentID,
	).Scan(&userListRaw, &groupListRaw)
	if err == nil {
		caregivers = &ResidentCaregiverDTO{}

		// 解析 group_list JSONB array -> []string
		if groupListRaw.Valid && groupListRaw.String != "" && groupListRaw.String != "null" {
			var groupList []string
			if err := json.Unmarshal([]byte(groupListRaw.String), &groupList); err == nil {
				caregivers.GroupList = groupList
			}
		}

		// 解析 user_list JSONB array -> []string，然后查询完整的用户信息
		if userListRaw.Valid && userListRaw.String != "" && userListRaw.String != "null" {
			var userIDs []string
			if err := json.Unmarshal([]byte(userListRaw.String), &userIDs); err == nil && len(userIDs) > 0 {
				// 查询完整的用户信息
				query := `
					SELECT 
						u.user_id::text,
						u.user_account,
						COALESCE(u.nickname, '') as nickname,
						u.role,
						u.status
					FROM users u
					WHERE u.tenant_id = $1
					  AND u.user_id::text = ANY($2::text[])
					ORDER BY u.user_account
				`
				rows, err := s.db.QueryContext(ctx, query, req.TenantID, pq.Array(userIDs))
				if err == nil {
					defer rows.Close()
					var userList []UserDTO
					for rows.Next() {
						var user UserDTO
						if err := rows.Scan(
							&user.UserID,
							&user.UserAccount,
							&user.Nickname,
							&user.Role,
							&user.Status,
						); err == nil {
							user.TenantID = req.TenantID
							userList = append(userList, user)
						}
					}
					caregivers.UserList = userList
				} else {
					// 查询用户信息失败，记录日志但不阻止返回
					s.logger.Warn("GetResident failed to query user details",
						zap.String("tenant_id", req.TenantID),
						zap.Strings("user_ids", userIDs),
						zap.Error(err),
					)
				}
			}
		}
	} else if err != sql.ErrNoRows {
		// 查询出错但不是"记录不存在"，记录日志但不阻止返回
		s.logger.Warn("GetResident failed to query caregivers",
			zap.String("tenant_id", req.TenantID),
			zap.String("resident_id", actualResidentID),
			zap.Error(err),
		)
	}

	return &GetResidentResponse{
		Resident:   resident,
		PHI:        phi,
		Contacts:   contacts,
		Caregivers: caregivers,
	}, nil
}

// domainPHIToDTO 将 domain.ResidentPHI 转换为 ResidentPHIDTO
func domainPHIToDTO(phi *domain.ResidentPHI) *ResidentPHIDTO {
	if phi == nil {
		return nil
	}
	dto := &ResidentPHIDTO{
		PhiID: phi.PhiID,
	}
	if phi.FirstName != "" {
		dto.FirstName = &phi.FirstName
	}
	if phi.LastName != "" {
		dto.LastName = &phi.LastName
	}
	if phi.Gender != "" {
		dto.Gender = &phi.Gender
	}
	if phi.DateOfBirth != nil {
		ts := phi.DateOfBirth.Unix()
		dto.DateOfBirth = &ts
	}
	if phi.ResidentPhone != "" {
		dto.ResidentPhone = &phi.ResidentPhone
	}
	if phi.ResidentEmail != "" {
		dto.ResidentEmail = &phi.ResidentEmail
	}
	if phi.WeightLb != nil {
		dto.WeightLb = phi.WeightLb
	}
	if phi.HeightFt != nil {
		dto.HeightFt = phi.HeightFt
	}
	if phi.HeightIn != nil {
		dto.HeightIn = phi.HeightIn
	}
	if phi.MobilityLevel != nil {
		dto.MobilityLevel = phi.MobilityLevel
	}
	if phi.TremorStatus != "" {
		dto.TremorStatus = &phi.TremorStatus
	}
	if phi.MobilityAid != "" {
		dto.MobilityAid = &phi.MobilityAid
	}
	if phi.ADLAssistance != "" {
		dto.ADLAssistance = &phi.ADLAssistance
	}
	if phi.CommStatus != "" {
		dto.CommStatus = &phi.CommStatus
	}
	dto.HasHypertension = &phi.HasHypertension
	dto.HasHyperlipaemia = &phi.HasHyperlipaemia
	dto.HasHyperglycaemia = &phi.HasHyperglycaemia
	dto.HasStrokeHistory = &phi.HasStrokeHistory
	dto.HasParalysis = &phi.HasParalysis
	dto.HasAlzheimer = &phi.HasAlzheimer
	if phi.MedicalHistory != "" {
		dto.MedicalHistory = &phi.MedicalHistory
	}
	if phi.HomeAddressStreet != "" {
		dto.HomeAddressStreet = &phi.HomeAddressStreet
	}
	if phi.HomeAddressCity != "" {
		dto.HomeAddressCity = &phi.HomeAddressCity
	}
	if phi.HomeAddressState != "" {
		dto.HomeAddressState = &phi.HomeAddressState
	}
	if phi.HomeAddressPostalCode != "" {
		dto.HomeAddressPostalCode = &phi.HomeAddressPostalCode
	}
	if phi.PlusCode != "" {
		dto.PlusCode = &phi.PlusCode
	}
	return dto
}

// domainContactToDTO 将 domain.ResidentContact 转换为 ResidentContactDTO
func domainContactToDTO(contact *domain.ResidentContact) *ResidentContactDTO {
	if contact == nil {
		return nil
	}
	dto := &ResidentContactDTO{
		ContactID:          contact.ContactID,
		Slot:               contact.Slot,
		IsEnabled:          contact.IsEnabled,
		ReceiveSMS:         contact.ReceiveSMS,
		ReceiveEmail:       contact.ReceiveEmail,
		IsEmergencyContact: contact.IsEnabled, // 使用 IsEnabled 作为 IsEmergencyContact（向后兼容）
	}
	if contact.Relationship.Valid && contact.Relationship.String != "" {
		dto.Relationship = &contact.Relationship.String
	}
	if contact.ContactFirstName.Valid && contact.ContactFirstName.String != "" {
		dto.ContactFirstName = &contact.ContactFirstName.String
	}
	if contact.ContactLastName.Valid && contact.ContactLastName.String != "" {
		dto.ContactLastName = &contact.ContactLastName.String
	}
	if contact.ContactPhone.Valid && contact.ContactPhone.String != "" {
		dto.ContactPhone = &contact.ContactPhone.String
	}
	if contact.ContactEmail.Valid && contact.ContactEmail.String != "" {
		dto.ContactEmail = &contact.ContactEmail.String
	}
	return dto
}

// CreateResident 创建住户
func (s *residentService) CreateResident(ctx context.Context, req CreateResidentRequest) (*CreateResidentResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.CurrentUserID == "" {
		return nil, fmt.Errorf("current_user_id is required")
	}
	if req.InherentAttributes == nil {
		return nil, fmt.Errorf("inherent_attributes is required")
	}
	if req.InherentAttributes.ResidentAccount == "" {
		return nil, fmt.Errorf("resident_account is required (each institution has its own encoding pattern)")
	}
	if req.InherentAttributes.Nickname == "" {
		return nil, fmt.Errorf("nickname is required")
	}

	// 1.1 角色权限检查：只有 Admin 和 Manager 可以创建 resident
	if req.CurrentUserRole == "" {
		return nil, fmt.Errorf("permission denied: user role is required")
	}
	allowedRoles := []string{"Admin", "Manager"}
	roleAllowed := false
	for _, allowedRole := range allowedRoles {
		if strings.EqualFold(req.CurrentUserRole, allowedRole) {
			roleAllowed = true
			break
		}
	}
	if !roleAllowed {
		return nil, fmt.Errorf("permission denied: only Admin and Manager can create residents (current role: %s)", req.CurrentUserRole)
	}

	// 1.2 验证用户角色（从数据库查询，不信任前端传递的值）
	var userRole sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT role FROM users WHERE tenant_id = $1 AND user_id::text = $2`,
		req.TenantID, req.CurrentUserID,
	).Scan(&userRole)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to verify user role: %w", err)
	}
	if !userRole.Valid {
		return nil, fmt.Errorf("permission denied: user has no role")
	}
	// 再次验证角色（从数据库查询的值）
	roleAllowed = false
	for _, allowedRole := range allowedRoles {
		if strings.EqualFold(userRole.String, allowedRole) {
			roleAllowed = true
			break
		}
	}
	if !roleAllowed {
		return nil, fmt.Errorf("permission denied: only Admin and Manager can create residents (user role: %s)", userRole.String)
	}

	// 2. 业务规则验证
	// 2.1 resident_account 转换为小写
	residentAccount := strings.ToLower(strings.TrimSpace(req.InherentAttributes.ResidentAccount))

	// 2.2 计算 account_hash
	accountHashHex := HashAccount(residentAccount)
	accountHash, err := hex.DecodeString(accountHashHex)
	if err != nil || len(accountHash) == 0 {
		return nil, fmt.Errorf("failed to hash account")
	}

	// 2.3 处理 password_hash（如果提供了）
	var passwordHash []byte
	if req.InherentAttributes.PasswordHash != "" {
		ph, err := hex.DecodeString(req.InherentAttributes.PasswordHash)
		if err == nil && len(ph) > 0 {
			passwordHash = ph
		}
	}
	// 如果没有提供 password_hash，生成默认密码的 hash（无盐，简单方式）
	if len(passwordHash) == 0 {
		passwordHashHex := GeneratePasswordHash("ChangeMe123!") // 默认密码，无盐
		ph, err := hex.DecodeString(passwordHashHex)
		if err == nil && len(ph) > 0 {
			passwordHash = ph
		}
	}

	// 2.4 处理 phone_hash 和 email_hash（从请求中获取，前端已计算）
	var phoneHash, emailHash []byte
	if req.InherentAttributes.PhoneHash != "" {
		ph, err := hex.DecodeString(req.InherentAttributes.PhoneHash)
		if err == nil && len(ph) > 0 {
			phoneHash = ph
		}
	}
	if req.InherentAttributes.EmailHash != "" {
		eh, err := hex.DecodeString(req.InherentAttributes.EmailHash)
		if err == nil && len(eh) > 0 {
			emailHash = eh
		}
	}

	// 2.5 Hash 唯一性检查
	if err := s.checkHashUniqueness(ctx, req.TenantID, "residents", phoneHash, emailHash, "", ""); err != nil {
		return nil, err
	}

	// 2.6 处理 admission_date（默认当前日期）
	admissionDate := time.Now()
	if req.InherentAttributes.AdmissionDate != nil {
		admissionDate = time.Unix(*req.InherentAttributes.AdmissionDate, 0).UTC()
	}

	// 2.7 处理 status（默认 "active"）
	status := req.InherentAttributes.Status
	if status == "" {
		status = "active"
	}

	// 2.8 unit_id 验证和权限检查（从 UnitRelation 获取）
	unitID := ""
	if req.UnitRelation != nil {
		unitID = req.UnitRelation.UnitID
	}
	if unitID != "" {
		// 验证 unit 存在
		var unitExists bool
		err := s.db.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM units WHERE tenant_id = $1 AND unit_id::text = $2)`,
			req.TenantID, unitID,
		).Scan(&unitExists)
		if err != nil {
			return nil, fmt.Errorf("failed to check unit existence: %w", err)
		}
		if !unitExists {
			return nil, fmt.Errorf("unit not found")
		}
		// VIP 单元创建时也必须带 room_id，避免下游业务因缺少 room 失败
		var isPublic, isShared bool
		if err := s.db.QueryRowContext(ctx, `SELECT is_public, is_shared_unit FROM units WHERE tenant_id = $1 AND unit_id = $2`, req.TenantID, unitID).Scan(&isPublic, &isShared); err == nil && !isPublic && !isShared {
			if req.UnitRelation.RoomID == "" {
				return nil, fmt.Errorf("VIP unit requires room_id")
			}
		}

		// BranchOnly 权限检查（使用 branch_id）
		// 注意：NULL、""、"default" 在 branch 表中都表示空院区/默认院区
		if req.PermissionCheck != nil && req.PermissionCheck.BranchOnly && req.CurrentUserID != "" {
			// 查询 unit 的 branch_id 和 branch_name
			var unitBranchID sql.NullString
			var unitBranchName sql.NullString
			err := s.db.QueryRowContext(ctx,
				`SELECT u.branch_id::text, br.branch_name
				 FROM units u
				 LEFT JOIN branches br ON br.branch_id = u.branch_id
				 WHERE u.tenant_id = $1 AND u.unit_id::text = $2`,
				req.TenantID, unitID,
			).Scan(&unitBranchID, &unitBranchName)
			if err != nil {
				return nil, fmt.Errorf("failed to check unit branch info: %w", err)
			}

			// 判断 unit 是否属于空院区/默认院区（NULL、""、"default"）
			isNullBranch := !unitBranchID.Valid ||
				!unitBranchName.Valid ||
				unitBranchName.String == "" ||
				unitBranchName.String == "default"

			// 查询用户所属的 branch_id 列表
			userBranchIDs, hasBranches, err := s.getUserBranchIDs(ctx, req.TenantID, req.CurrentUserID)
			if err != nil {
				return nil, fmt.Errorf("failed to get user branch IDs: %w", err)
			}

			// 权限检查
			if !hasBranches {
				// 用户没有关联任何院区：只能创建空院区/默认院区的住户（NULL、""、"default"）
				if !isNullBranch {
					return nil, fmt.Errorf("permission denied: can only create residents with null/default branch (null, '', or 'default')")
				}
			} else {
				// 用户有关联院区：只能创建关联院区的住户（不能创建空院区的住户）
				if isNullBranch {
					return nil, fmt.Errorf("permission denied: can only create residents in assigned branches")
				}

				// 检查 unit 的 branch_id 是否在用户的 branch_id 列表中
				if !unitBranchID.Valid {
					return nil, fmt.Errorf("permission denied: can only create residents in assigned branches")
				}

				allowed := false
				for _, userBranchID := range userBranchIDs {
					if userBranchID == unitBranchID.String {
						allowed = true
						break
					}
				}

				if !allowed {
					return nil, fmt.Errorf("permission denied: can only create residents in assigned branches")
				}
			}
		}
	}

	// 2.9 discharge_date 验证（仅在 status='discharged' 或 'transferred' 时可以有值）
	// 注意：CreateResident 请求中没有 discharge_date 字段，因为创建时默认是 active 状态

	// 3. 创建 Resident 记录
	resident := &domain.Resident{
		ResidentAccount:     residentAccount,
		ResidentAccountHash: accountHash,
		Nickname:            strings.TrimSpace(req.InherentAttributes.Nickname),
		Status:              status,
		Role:                "Resident",
		AdmissionDate:       &admissionDate,
		IsAccessEnabled:     req.InherentAttributes.IsAccessEnabled,
		Note:                req.InherentAttributes.Note,
		PhoneHash:           phoneHash,
		EmailHash:           emailHash,
		PasswordHash:        passwordHash,
	}
	if req.InherentAttributes.ServiceLevel != "" {
		resident.ServiceLevel = req.InherentAttributes.ServiceLevel
	}

	// 2.10 BranchID 权限验证（在设置之前）
	// Service 层自己查询用户的 branch 信息，不信任 Handler 传递的值
	if req.InherentAttributes.BranchID != "" {
		// 如果是 Manager 且有 BranchOnly 权限，验证 branch_id 必须在用户的 branch 范围内
		if req.PermissionCheck != nil && req.PermissionCheck.BranchOnly && req.CurrentUserID != "" {
			// Service 层自己查询用户的 branch 信息
			userBranches, hasBranches, err := s.getUserBranches(ctx, req.TenantID, req.CurrentUserID)
			if err != nil {
				return nil, fmt.Errorf("failed to get user branches: %w", err)
			}

			if !hasBranches || len(userBranches) == 0 {
				// Manager 没有关联任何院区：只能设置 NULL、""、"default" 的 branch_id
				if req.InherentAttributes.BranchID != "" {
					return nil, fmt.Errorf("permission denied: Manager without branches can only set null branch_id")
				}
			} else {
				// Manager 有关联院区：只能设置自己 branch 的 branch_id；与 default 互迁不要求拥有 default
				allowed := false
				for _, userBranch := range userBranches {
					if userBranch.BranchID == req.InherentAttributes.BranchID {
						allowed = true
						break
					}
				}
				if !allowed && s.db != nil {
					var defaultBranchID string
					if err := s.db.QueryRowContext(ctx, `SELECT branch_id::text FROM branches WHERE tenant_id = $1 AND branch_name = $2`, req.TenantID, domain.DefaultBranchName).Scan(&defaultBranchID); err == nil && defaultBranchID != "" && req.InherentAttributes.BranchID == defaultBranchID {
						allowed = true
					}
				}
				if !allowed {
					return nil, fmt.Errorf("permission denied: can only set branch_id from assigned branches")
				}
			}
		}
		// Admin 无限制，直接设置
		resident.BranchID = req.InherentAttributes.BranchID
	}
	if req.InherentAttributes.DischargeDate != nil {
		dischargeDate := time.Unix(*req.InherentAttributes.DischargeDate, 0).UTC()
		resident.DischargeDate = &dischargeDate
	}
	// UnitRelation 中的 UnitID, RoomID, BedID
	if req.UnitRelation != nil {
		if req.UnitRelation.UnitID != "" {
			resident.UnitID = req.UnitRelation.UnitID
		}
		if req.UnitRelation.RoomID != "" {
			resident.RoomID = req.UnitRelation.RoomID
		}
		if req.UnitRelation.BedID != "" {
			resident.BedID = req.UnitRelation.BedID
		}
	}
	if len(req.InherentAttributes.Metadata) > 0 {
		resident.Metadata = req.InherentAttributes.Metadata
	}

	residentID, err := s.residentsRepo.CreateResident(ctx, req.TenantID, resident)
	if err != nil {
		s.logger.Error("CreateResident failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("resident_account", residentAccount),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to create resident: %w", err)
	}

	// 4. 创建 PHI 记录（如果提供了 PHI 数据）
	// 注意：即使 FirstName 为空，也可以创建 PHI（FirstName 是可选字段）
	if req.InherentAttributes.PHI != nil {
		phiUpdate := &domain.ResidentPHIUpdate{}

		phi := req.InherentAttributes.PHI
		// 转换所有字段为 UpdateX 类型（创建时，有值的字段设置为 UpdateActionUpdate）
		if phi.FirstName != "" {
			phiUpdate.FirstName = &domain.UpdateString{
				Action: domain.UpdateActionUpdate,
				Value:  phi.FirstName,
			}
		}
		if phi.LastName != "" {
			phiUpdate.LastName = &domain.UpdateString{
				Action: domain.UpdateActionUpdate,
				Value:  phi.LastName,
			}
		}
		if phi.Gender != "" {
			phiUpdate.Gender = &domain.UpdateString{
				Action: domain.UpdateActionUpdate,
				Value:  phi.Gender,
			}
		}
		if phi.DateOfBirth != nil {
			dobTime := time.Unix(*phi.DateOfBirth, 0).UTC()
			phiUpdate.DateOfBirth = &domain.UpdateTime{
				Action: domain.UpdateActionUpdate,
				Value:  &dobTime,
			}
		}
		// 只在 save_phone/save_email 为 true 时保存明文到 resident_phi 表
		if phi.SavePhone && phi.ResidentPhone != "" {
			phiUpdate.ResidentPhone = &domain.UpdateString{
				Action: domain.UpdateActionUpdate,
				Value:  phi.ResidentPhone,
			}
		}
		if phi.SaveEmail && phi.ResidentEmail != "" {
			phiUpdate.ResidentEmail = &domain.UpdateString{
				Action: domain.UpdateActionUpdate,
				Value:  phi.ResidentEmail,
			}
		}
		// 其他 PHI 字段
		if phi.WeightLb != nil {
			phiUpdate.WeightLb = &domain.UpdateFloat64{
				Action: domain.UpdateActionUpdate,
				Value:  *phi.WeightLb,
			}
		}
		if phi.HeightFt != nil {
			phiUpdate.HeightFt = &domain.UpdateFloat64{
				Action: domain.UpdateActionUpdate,
				Value:  *phi.HeightFt,
			}
		}
		if phi.HeightIn != nil {
			phiUpdate.HeightIn = &domain.UpdateFloat64{
				Action: domain.UpdateActionUpdate,
				Value:  *phi.HeightIn,
			}
		}
		if phi.MobilityLevel != nil {
			phiUpdate.MobilityLevel = &domain.UpdateInt{
				Action: domain.UpdateActionUpdate,
				Value:  *phi.MobilityLevel,
			}
		}
		if phi.TremorStatus != "" {
			phiUpdate.TremorStatus = &domain.UpdateString{
				Action: domain.UpdateActionUpdate,
				Value:  phi.TremorStatus,
			}
		}
		if phi.MobilityAid != "" {
			phiUpdate.MobilityAid = &domain.UpdateString{
				Action: domain.UpdateActionUpdate,
				Value:  phi.MobilityAid,
			}
		}
		if phi.ADLAssistance != "" {
			phiUpdate.ADLAssistance = &domain.UpdateString{
				Action: domain.UpdateActionUpdate,
				Value:  phi.ADLAssistance,
			}
		}
		if phi.CommStatus != "" {
			phiUpdate.CommStatus = &domain.UpdateString{
				Action: domain.UpdateActionUpdate,
				Value:  phi.CommStatus,
			}
		}
		if phi.HasHypertension != nil {
			phiUpdate.HasHypertension = &domain.UpdateBool{
				Action: domain.UpdateActionUpdate,
				Value:  *phi.HasHypertension,
			}
		}
		if phi.HasHyperlipaemia != nil {
			phiUpdate.HasHyperlipaemia = &domain.UpdateBool{
				Action: domain.UpdateActionUpdate,
				Value:  *phi.HasHyperlipaemia,
			}
		}
		if phi.HasHyperglycaemia != nil {
			phiUpdate.HasHyperglycaemia = &domain.UpdateBool{
				Action: domain.UpdateActionUpdate,
				Value:  *phi.HasHyperglycaemia,
			}
		}
		if phi.HasStrokeHistory != nil {
			phiUpdate.HasStrokeHistory = &domain.UpdateBool{
				Action: domain.UpdateActionUpdate,
				Value:  *phi.HasStrokeHistory,
			}
		}
		if phi.HasParalysis != nil {
			phiUpdate.HasParalysis = &domain.UpdateBool{
				Action: domain.UpdateActionUpdate,
				Value:  *phi.HasParalysis,
			}
		}
		if phi.HasAlzheimer != nil {
			phiUpdate.HasAlzheimer = &domain.UpdateBool{
				Action: domain.UpdateActionUpdate,
				Value:  *phi.HasAlzheimer,
			}
		}
		if phi.MedicalHistory != "" {
			phiUpdate.MedicalHistory = &domain.UpdateString{
				Action: domain.UpdateActionUpdate,
				Value:  phi.MedicalHistory,
			}
		}
		if phi.HomeAddressStreet != "" {
			phiUpdate.HomeAddressStreet = &domain.UpdateString{
				Action: domain.UpdateActionUpdate,
				Value:  phi.HomeAddressStreet,
			}
		}
		if phi.HomeAddressCity != "" {
			phiUpdate.HomeAddressCity = &domain.UpdateString{
				Action: domain.UpdateActionUpdate,
				Value:  phi.HomeAddressCity,
			}
		}
		if phi.HomeAddressState != "" {
			phiUpdate.HomeAddressState = &domain.UpdateString{
				Action: domain.UpdateActionUpdate,
				Value:  phi.HomeAddressState,
			}
		}
		if phi.HomeAddressPostalCode != "" {
			phiUpdate.HomeAddressPostalCode = &domain.UpdateString{
				Action: domain.UpdateActionUpdate,
				Value:  phi.HomeAddressPostalCode,
			}
		}
		if phi.PlusCode != "" {
			phiUpdate.PlusCode = &domain.UpdateString{
				Action: domain.UpdateActionUpdate,
				Value:  phi.PlusCode,
			}
		}

		// 使用 UpsertResidentPHIFields 方法（UPSERT 语义，如果不存在则创建，如果存在则更新）
		if err := s.residentsRepo.UpsertResidentPHIFields(ctx, req.TenantID, residentID, phiUpdate); err != nil {
			s.logger.Warn("Failed to create PHI record",
				zap.String("tenant_id", req.TenantID),
				zap.String("resident_id", residentID),
				zap.Error(err),
			)
			// 不失败整个操作，只记录警告
		}
	}

	// 5. 创建联系人记录（如果提供了 contacts）
	// 注意：联系人不登录系统，但需要保存 phone_hash 和 email_hash 用于搜索
	if req.InherentAttributes.Contacts != nil && len(req.InherentAttributes.Contacts) > 0 {
		for _, contactReq := range req.InherentAttributes.Contacts {
			contact := &domain.ResidentContact{
				Slot:             contactReq.Slot,
				IsEnabled:        contactReq.IsEnabled,
				Relationship:     sql.NullString{String: contactReq.Relationship, Valid: contactReq.Relationship != ""},
				ContactFirstName: sql.NullString{String: contactReq.ContactFirstName, Valid: contactReq.ContactFirstName != ""},
				ContactLastName:  sql.NullString{String: contactReq.ContactLastName, Valid: contactReq.ContactLastName != ""},
				ContactPhone:     sql.NullString{String: contactReq.ContactPhone, Valid: contactReq.ContactPhone != ""},
				ContactEmail:     sql.NullString{String: contactReq.ContactEmail, Valid: contactReq.ContactEmail != ""},
				ReceiveSMS:       contactReq.ReceiveSMS,
				ReceiveEmail:     contactReq.ReceiveEmail,
			}
			if contact.Slot == "" {
				contact.Slot = "A" // 默认 slot
			}

			// 处理 contact_phone_hash 和 contact_email_hash
			if contactReq.ContactPhoneHash != "" {
				phoneHash, err := hex.DecodeString(contactReq.ContactPhoneHash)
				if err == nil && len(phoneHash) > 0 {
					contact.ContactPhoneHash = phoneHash
				}
			} else if contactReq.ContactPhone != "" {
				// 如果前端没有提供 hash，但提供了 phone，则计算 hash
				phone := strings.ToLower(strings.TrimSpace(contactReq.ContactPhone))
				phoneHashHex := HashAccount(phone)
				phoneHash, err := hex.DecodeString(phoneHashHex)
				if err == nil && len(phoneHash) > 0 {
					contact.ContactPhoneHash = phoneHash
				}
			}
			if contactReq.ContactEmailHash != "" {
				emailHash, err := hex.DecodeString(contactReq.ContactEmailHash)
				if err == nil && len(emailHash) > 0 {
					contact.ContactEmailHash = emailHash
				}
			} else if contactReq.ContactEmail != "" {
				// 如果前端没有提供 hash，但提供了 email，则计算 hash
				email := strings.ToLower(strings.TrimSpace(contactReq.ContactEmail))
				emailHashHex := HashAccount(email)
				emailHash, err := hex.DecodeString(emailHashHex)
				if err == nil && len(emailHash) > 0 {
					contact.ContactEmailHash = emailHash
				}
			}

			_, err := s.residentsRepo.CreateResidentContact(ctx, req.TenantID, residentID, contact)
			if err != nil {
				s.logger.Warn("Failed to create contact",
					zap.String("tenant_id", req.TenantID),
					zap.String("resident_id", residentID),
					zap.String("slot", contact.Slot),
					zap.Error(err),
				)
				// 不失败整个操作，只记录警告
			}
		}
	}

	SyncUnitCards(ctx, req.TenantID, unitID)

	return &CreateResidentResponse{
		ResidentID: residentID,
	}, nil
}

// findExistingContactByHash 根据 phone_hash 或 email_hash 查找已存在的联系人
// 如果找到，返回联系人信息以便复用
func (s *residentService) findExistingContactByHash(ctx context.Context, tenantID string, phoneHash, emailHash []byte) *domain.ResidentContact {
	return s.findExistingContactByHashExcluding(ctx, tenantID, phoneHash, emailHash, "")
}

// findExistingContactByHashExcluding 根据 phone_hash 或 email_hash 查找已存在的联系人（排除指定的 contact_id）
// Deprecated: 联系人不登录系统，不再使用 phone_hash/email_hash 字段，此函数暂时返回 nil
// 如果找到，返回联系人信息以便复用
func (s *residentService) findExistingContactByHashExcluding(ctx context.Context, tenantID string, phoneHash, emailHash []byte, excludeContactID string) *domain.ResidentContact {
	// 联系人不登录系统，不再使用 phone_hash/email_hash 字段
	// 此函数已废弃，暂时返回 nil
	return nil
}

// checkHashUniqueness 检查 phone_hash 或 email_hash 的唯一性
func (s *residentService) checkHashUniqueness(ctx context.Context, tenantID, tableName string, phoneHash, emailHash []byte, excludeID, excludeField string) error {
	if len(phoneHash) > 0 {
		var query string
		var args []any
		if excludeID != "" {
			query = fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE tenant_id = $1 AND phone_hash = $2 AND %s::text != $3`, tableName, excludeField)
			args = []any{tenantID, phoneHash, excludeID}
		} else {
			query = fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE tenant_id = $1 AND phone_hash = $2`, tableName)
			args = []any{tenantID, phoneHash}
		}
		var count int
		if err := s.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
			return fmt.Errorf("failed to check phone_hash uniqueness: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("phone already exists in this organization")
		}
	}
	if len(emailHash) > 0 {
		var query string
		var args []any
		if excludeID != "" {
			query = fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE tenant_id = $1 AND email_hash = $2 AND %s::text != $3`, tableName, excludeField)
			args = []any{tenantID, emailHash, excludeID}
		} else {
			query = fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE tenant_id = $1 AND email_hash = $2`, tableName)
			args = []any{tenantID, emailHash}
		}
		var count int
		if err := s.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
			return fmt.Errorf("failed to check email_hash uniqueness: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("email already exists in this organization")
		}
	}
	return nil
}

// UpdateResident 更新住户
func (s *residentService) UpdateResident(ctx context.Context, req UpdateResidentRequest) (*UpdateResidentResponse, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.ResidentID == "" {
		return nil, fmt.Errorf("resident_id is required")
	}

	// 1. 权限检查（细粒度）
	if req.CurrentUserRole == "Resident" {
		// Resident: 只能更新自己
		if req.CurrentUserID != req.ResidentID {
			return nil, fmt.Errorf("access denied: can only update own information")
		}
		// 允许更新
	} else if req.CurrentUserRole == "Family" {
		// Family: 不允许更新 resident（只能更新自己的 contact）
		return nil, fmt.Errorf("access denied: family cannot update resident information")
	} else if req.CurrentUserRole == "Admin" {
		// Admin: 先检查 accountID 的角色是 Admin，然后允许更新所有 resident
		var userRole sql.NullString
		err := s.db.QueryRowContext(ctx,
			`SELECT role FROM users WHERE tenant_id = $1 AND user_id::text = $2`,
			req.TenantID, req.CurrentUserID,
		).Scan(&userRole)
		if err != nil {
			return nil, fmt.Errorf("failed to verify user role: %w", err)
		}
		if !userRole.Valid || userRole.String != "Admin" {
			return nil, fmt.Errorf("access denied: user role is not Admin")
		}
		// 允许更新所有 resident
	} else if req.CurrentUserRole == "Manager" {
		// Manager: resident 与 Manager 的 branch 相同，如果两者的 branchName 均为 ""，视为相同
		// 查询 1：验证用户角色是 Manager
		var userRole sql.NullString
		err := s.db.QueryRowContext(ctx,
			`SELECT role FROM users WHERE tenant_id = $1 AND user_id::text = $2`,
			req.TenantID, req.CurrentUserID,
		).Scan(&userRole)
		if err != nil {
			return nil, fmt.Errorf("failed to verify user role: %w", err)
		}
		if !userRole.Valid || userRole.String != "Manager" {
			return nil, fmt.Errorf("access denied: user role is not Manager")
		}

		// 查询用户的所有 branch_id
		rows, err := s.db.QueryContext(ctx,
			`SELECT ub.branch_id::text FROM user_branches ub
			 WHERE ub.tenant_id = $1 AND ub.user_id::text = $2`,
			req.TenantID, req.CurrentUserID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get user branches: %w", err)
		}
		defer rows.Close()
		var userBranchIDs []string
		for rows.Next() {
			var bid string
			if err := rows.Scan(&bid); err == nil {
				userBranchIDs = append(userBranchIDs, bid)
			}
		}

		// 查询目标 resident 的 branch_id
		var targetBranchID sql.NullString
		err = s.db.QueryRowContext(ctx,
			`SELECT branch_id::text FROM residents
			 WHERE tenant_id = $1 AND resident_id::text = $2`,
			req.TenantID, req.ResidentID,
		).Scan(&targetBranchID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("resident not found")
			}
			return nil, fmt.Errorf("failed to get resident info: %w", err)
		}

		// 用户无 branch 限制时（无 user_branches 记录）允许更新
		if len(userBranchIDs) > 0 {
			targetBranch := ""
			if targetBranchID.Valid {
				targetBranch = targetBranchID.String
			}
			found := false
			for _, bid := range userBranchIDs {
				if bid == targetBranch {
					found = true
					break
				}
			}
			// 与 default branch 互迁不要求用户拥有 default 权限，避免每个 manager 都要挂两个 branch
			if !found && targetBranch != "" && s.db != nil {
				var defaultBranchID string
				if err := s.db.QueryRowContext(ctx, `SELECT branch_id::text FROM branches WHERE tenant_id = $1 AND branch_name = $2`, req.TenantID, domain.DefaultBranchName).Scan(&defaultBranchID); err == nil && defaultBranchID != "" && targetBranch == defaultBranchID {
					found = true
				}
			}
			if !found && targetBranch != "" {
				return nil, fmt.Errorf("permission denied: can only update residents in same branch")
			}
		}
	} else if req.CurrentUserRole == "Nurse" || req.CurrentUserRole == "Caregiver" {
		// Caregiver/Nurse: 首先检查是否有 U 权限，其次检查护理关系
		// 检查 U 权限
		hasUPermission := false
		err := s.db.QueryRowContext(ctx,
			`SELECT EXISTS(
				SELECT 1 FROM role_permissions
				WHERE tenant_id = $1 AND role_code = $2 AND resource_type = 'residents' AND permission_type = 'U'
			)`,
			SystemTenantID, req.CurrentUserRole,
		).Scan(&hasUPermission)
		if err != nil {
			return nil, fmt.Errorf("failed to check U permission: %w", err)
		}
		if !hasUPermission {
			return nil, fmt.Errorf("permission denied: no update permission for residents")
		}

		// 检查护理关系（两种路径）：
		// 1. 直接分配：resident_caregivers.userList 包含 user_id
		// 2. 通过 user_tag 分配：resident_caregivers.groupList 中的 tag_id 对应的 tag_name 在 users.tags 中
		var isAssigned bool
		err = s.db.QueryRowContext(ctx,
			`SELECT EXISTS(
				SELECT 1 FROM resident_caregivers rc
				WHERE rc.tenant_id = $1
				  AND rc.resident_id::text = $2
				  AND (
					-- 直接分配：user_list 中包含 user_id
					rc.user_list::text LIKE $3
					OR rc.user_list::text LIKE $4
					-- 通过 user_tag 分配：group_list 中的 tag_id 对应的 tag_name 在 users.tags 中
					OR EXISTS(
						SELECT 1 FROM users u
						WHERE u.tenant_id = $1
						  AND u.user_id::text = $5
						  AND EXISTS(
							SELECT 1 FROM jsonb_array_elements_text(u.tags) AS user_tag_name
							WHERE EXISTS(
								SELECT 1 FROM tags_catalog tc
								WHERE tc.tenant_id = $1
								  AND tc.tag_type = 'user_tag'
								  AND tc.tag_name = user_tag_name
								  AND tc.tag_id::text = ANY(
									SELECT jsonb_array_elements_text(rc.group_list)::text
								  )
							)
						  )
					)
				  )
			)`,
			req.TenantID, req.ResidentID, req.CurrentUserID, "%\""+req.CurrentUserID+"\"%", req.CurrentUserID,
		).Scan(&isAssigned)
		if err != nil {
			return nil, fmt.Errorf("failed to check assignment: %w", err)
		}
		if !isAssigned {
			return nil, fmt.Errorf("permission denied: can only update assigned residents")
		}
		// 允许更新
	} else {
		// 其他角色：拒绝
		return nil, fmt.Errorf("permission denied: role %s has no update permission", req.CurrentUserRole)
	}

	// 2. 验证住户存在
	existingResident, err := s.residentsRepo.GetResident(ctx, req.TenantID, req.ResidentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get resident: %w", err)
	}

	// 3. 处理 InherentAttributes（residents 表字段 + PHI + Contacts）
	var branchChangedAndUnbound bool // 本请求中因迁院区已解绑 unit/room/bed，后续不再应用 UnitRelation 以免写回旧绑定
	if req.InherentAttributes != nil {
		// 3.1 构建 residents 表字段更新
		residentUpdate := &domain.ResidentUpdate{}

		// 处理 ResidentAccount（需要同时更新 ResidentAccountHash）
		if req.InherentAttributes.ResidentAccount != nil && req.InherentAttributes.ResidentAccount.Action == domain.UpdateActionUpdate {
			residentAccount := strings.ToLower(strings.TrimSpace(req.InherentAttributes.ResidentAccount.Value))
			residentUpdate.ResidentAccount = &domain.UpdateString{
				Action: domain.UpdateActionUpdate,
				Value:  residentAccount,
			}
			// 同时更新 resident_account_hash
			hashHex := HashAccount(residentAccount)
			hashBytes, err := hex.DecodeString(hashHex)
			if err != nil {
				return nil, fmt.Errorf("failed to decode resident_account_hash: %w", err)
			}
			residentUpdate.ResidentAccountHash = &domain.UpdateBytes{
				Action: domain.UpdateActionUpdate,
				Value:  hashBytes,
			}
		}

		// 处理其他 residents 表字段
		residentUpdate.Nickname = req.InherentAttributes.Nickname
		residentUpdate.PasswordHash = req.InherentAttributes.PasswordHash
		residentUpdate.Status = req.InherentAttributes.Status
		residentUpdate.ServiceLevel = req.InherentAttributes.ServiceLevel
		residentUpdate.AdmissionDate = req.InherentAttributes.AdmissionDate
		residentUpdate.DischargeDate = req.InherentAttributes.DischargeDate
		residentUpdate.BranchID = req.InherentAttributes.BranchID
		// 迁院区时解除原 unit/room/bed 绑定：新 branch 与当前 resident 的 branch 不同则清空；或新 branch 为 default 且当前 unit 属于其他院区也清空（修 r3 已在 default 但仍有旧 unit 的情况）
		if req.InherentAttributes.BranchID != nil && req.InherentAttributes.BranchID.Action == domain.UpdateActionUpdate {
			newBranchID := strings.TrimSpace(req.InherentAttributes.BranchID.Value)
			curBranchID := strings.TrimSpace(existingResident.BranchID)
			shouldUnbind := false
			if newBranchID != curBranchID && existingResident.UnitID != "" {
				shouldUnbind = true
			} else if newBranchID != "" && existingResident.UnitID != "" && s.db != nil {
				var defaultBranchID string
				err := s.db.QueryRowContext(ctx, `SELECT branch_id::text FROM branches WHERE tenant_id = $1 AND branch_name = $2`, req.TenantID, domain.DefaultBranchName).Scan(&defaultBranchID)
				if err == nil && newBranchID == defaultBranchID {
					var unitBranchID sql.NullString
					err2 := s.db.QueryRowContext(ctx, `SELECT branch_id::text FROM units WHERE tenant_id = $1 AND unit_id = $2`, req.TenantID, existingResident.UnitID).Scan(&unitBranchID)
					if err2 == nil && (!unitBranchID.Valid || unitBranchID.String != defaultBranchID) {
						shouldUnbind = true
					}
				}
			}
			if shouldUnbind {
				residentUpdate.UnitID = &domain.UpdateString{Action: domain.UpdateActionDelete}
				residentUpdate.RoomID = &domain.UpdateString{Action: domain.UpdateActionDelete}
				residentUpdate.BedID = &domain.UpdateString{Action: domain.UpdateActionDelete}
				branchChangedAndUnbound = true
			}
		}
		// status=transferred：归默认院区（default，视为总部），由总部 Admin 再分配；覆盖请求中的 branch_id
		if residentUpdate.Status != nil && residentUpdate.Status.Action == domain.UpdateActionUpdate && residentUpdate.Status.Value == "transferred" {
			var defaultBranchID string
			err := s.db.QueryRowContext(ctx,
				`SELECT branch_id::text FROM branches WHERE tenant_id = $1 AND branch_name = $2`,
				req.TenantID, domain.DefaultBranchName,
			).Scan(&defaultBranchID)
			if err != nil {
				if err == sql.ErrNoRows {
					return nil, fmt.Errorf("default branch not found for tenant")
				}
				return nil, fmt.Errorf("resolve default branch: %w", err)
			}
			residentUpdate.BranchID = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: defaultBranchID}
			if existingResident.UnitID != "" {
				residentUpdate.UnitID = &domain.UpdateString{Action: domain.UpdateActionDelete}
				residentUpdate.RoomID = &domain.UpdateString{Action: domain.UpdateActionDelete}
				residentUpdate.BedID = &domain.UpdateString{Action: domain.UpdateActionDelete}
				branchChangedAndUnbound = true
			}
		}
		residentUpdate.IsAccessEnabled = req.InherentAttributes.IsAccessEnabled
		residentUpdate.Note = req.InherentAttributes.Note
		residentUpdate.Phone = req.InherentAttributes.Phone
		residentUpdate.Email = req.InherentAttributes.Email
		residentUpdate.PhoneHash = req.InherentAttributes.PhoneHash
		residentUpdate.EmailHash = req.InherentAttributes.EmailHash
		residentUpdate.Metadata = req.InherentAttributes.Metadata

		// discharge_date 验证：仅在 status='discharged' 或 'transferred' 时可以有值
		if residentUpdate.DischargeDate != nil && residentUpdate.DischargeDate.Action == domain.UpdateActionUpdate && residentUpdate.DischargeDate.Value != nil {
			var currentStatus string
			if residentUpdate.Status != nil && residentUpdate.Status.Action == domain.UpdateActionUpdate {
				currentStatus = residentUpdate.Status.Value
			} else {
				currentStatus = existingResident.Status
			}
			if currentStatus != "discharged" && currentStatus != "transferred" {
				return nil, fmt.Errorf("discharge_date can only be set when status is 'discharged' or 'transferred'")
			}
		}

		// 处理 phone_hash/email_hash 的唯一性检查和验证
		var phoneHashToCheck, emailHashToCheck []byte
		// 注意：只有在 email/phone 不是占位符时，才设置 hashToCheck
		// 如果 email/phone 是占位符，即使提供了 hash，也不应该检查唯一性（因为 hash 已存在但值未保存）
		if req.InherentAttributes.PhoneHash != nil && req.InherentAttributes.PhoneHash.Action == domain.UpdateActionUpdate {
			// 只有在 phone 不是占位符时，才设置 phoneHashToCheck
			// 如果 phone 是占位符 "xxx-xxx-xxxx"，即使提供了 hash，也不应该检查唯一性（因为 hash 已存在但值未保存）
			if req.InherentAttributes.Phone != nil && req.InherentAttributes.Phone.Action == domain.UpdateActionUpdate &&
				req.InherentAttributes.Phone.Value == "xxx-xxx-xxxx" {
				// 占位符：不设置 phoneHashToCheck，避免触发唯一性检查
				phoneHashToCheck = nil
			} else if req.InherentAttributes.Phone == nil || req.InherentAttributes.Phone.Action != domain.UpdateActionUpdate ||
				(req.InherentAttributes.Phone.Value != "" && req.InherentAttributes.Phone.Value != "xxx-xxx-xxxx") {
				phoneHashToCheck = req.InherentAttributes.PhoneHash.Value
			}
		}
		if req.InherentAttributes.EmailHash != nil && req.InherentAttributes.EmailHash.Action == domain.UpdateActionUpdate {
			// 只有在 email 不是占位符时，才设置 emailHashToCheck
			// 如果 email 是占位符 "***@***"，即使提供了 hash，也不应该检查唯一性（因为 hash 已存在但值未保存）
			if req.InherentAttributes.Email != nil && req.InherentAttributes.Email.Action == domain.UpdateActionUpdate &&
				req.InherentAttributes.Email.Value == "***@***" {
				// 占位符：不设置 emailHashToCheck，避免触发唯一性检查
				emailHashToCheck = nil
			} else if req.InherentAttributes.Email == nil || req.InherentAttributes.Email.Action != domain.UpdateActionUpdate ||
				(req.InherentAttributes.Email.Value != "" && req.InherentAttributes.Email.Value != "***@***") {
				emailHashToCheck = req.InherentAttributes.EmailHash.Value
			}
		}

		// 检查 phone/email 和 hash 的一致性
		if req.InherentAttributes.Phone != nil && req.InherentAttributes.Phone.Action == domain.UpdateActionUpdate {
			phone := strings.ToLower(strings.TrimSpace(req.InherentAttributes.Phone.Value))
			// 跳过占位符 "xxx-xxx-xxxx"：占位符表示 phone_hash 已存在但 phone 未保存，不需要更新或验证
			if phone != "" && phone != "xxx-xxx-xxxx" {
				expectedHash := sha256Hash(phone)
				if req.InherentAttributes.PhoneHash != nil && req.InherentAttributes.PhoneHash.Action == domain.UpdateActionUpdate {
					if !equalBytes(expectedHash, req.InherentAttributes.PhoneHash.Value) {
						return nil, fmt.Errorf("phone and phone_hash mismatch")
					}
				} else {
					// 如果提供了 phone 但没有提供 phoneHash，需要计算
					residentUpdate.PhoneHash = &domain.UpdateBytes{
						Action: domain.UpdateActionUpdate,
						Value:  expectedHash,
					}
					phoneHashToCheck = expectedHash
				}
			} else if phone == "xxx-xxx-xxxx" {
				// 占位符：清除 phoneHashToCheck，避免触发唯一性检查
				// 占位符表示 hash 已存在但 phone 未保存，不需要更新或验证
				phoneHashToCheck = nil
			}
		}
		if req.InherentAttributes.Email != nil && req.InherentAttributes.Email.Action == domain.UpdateActionUpdate {
			email := strings.ToLower(strings.TrimSpace(req.InherentAttributes.Email.Value))
			// 跳过占位符 "***@***"：占位符表示 email_hash 已存在但 email 未保存，不需要更新或验证
			if email != "" && email != "***@***" {
				expectedHash := sha256Hash(email)
				if req.InherentAttributes.EmailHash != nil && req.InherentAttributes.EmailHash.Action == domain.UpdateActionUpdate {
					if !equalBytes(expectedHash, req.InherentAttributes.EmailHash.Value) {
						return nil, fmt.Errorf("email and email_hash mismatch")
					}
				} else {
					// 如果提供了 email 但没有提供 emailHash，需要计算
					residentUpdate.EmailHash = &domain.UpdateBytes{
						Action: domain.UpdateActionUpdate,
						Value:  expectedHash,
					}
					emailHashToCheck = expectedHash
				}
			} else if email == "***@***" {
				// 占位符：清除 emailHashToCheck，避免触发唯一性检查
				// 占位符表示 hash 已存在但 email 未保存，不需要更新或验证
				emailHashToCheck = nil
			}
		}

		// 检查 hash 唯一性（排除当前 resident）
		if len(phoneHashToCheck) > 0 {
			if err := s.checkHashUniqueness(ctx, req.TenantID, "residents", phoneHashToCheck, nil, req.ResidentID, "resident_id"); err != nil {
				return nil, err
			}
		}
		if len(emailHashToCheck) > 0 {
			if err := s.checkHashUniqueness(ctx, req.TenantID, "residents", nil, emailHashToCheck, req.ResidentID, "resident_id"); err != nil {
				return nil, err
			}
		}

		// 检查是否有需要更新的字段
		hasResidentFields := residentUpdate.ResidentAccount != nil || residentUpdate.Nickname != nil ||
			residentUpdate.PasswordHash != nil || residentUpdate.Status != nil || residentUpdate.ServiceLevel != nil ||
			residentUpdate.AdmissionDate != nil || residentUpdate.DischargeDate != nil || residentUpdate.BranchID != nil ||
			residentUpdate.IsAccessEnabled != nil || residentUpdate.Note != nil || residentUpdate.Phone != nil ||
			residentUpdate.Email != nil || residentUpdate.PhoneHash != nil || residentUpdate.EmailHash != nil ||
			residentUpdate.Metadata != nil

		// 如果有需要更新的字段，调用 UpdateResidentFields
		if hasResidentFields {
			if err := s.residentsRepo.UpdateResidentFields(ctx, req.TenantID, req.ResidentID, residentUpdate); err != nil {
				s.logger.Error("UpdateResidentFields failed",
					zap.String("tenant_id", req.TenantID),
					zap.String("resident_id", req.ResidentID),
					zap.Error(err),
				)
				return nil, fmt.Errorf("failed to update resident: %w", err)
			}
		}

		// 3.2 处理 PHI 数据更新
		if req.InherentAttributes.PHI != nil {
			phiUpdate := &domain.ResidentPHIUpdate{}

			// 映射所有 PHI 字段
			phiUpdate.FirstName = req.InherentAttributes.PHI.FirstName
			phiUpdate.LastName = req.InherentAttributes.PHI.LastName
			phiUpdate.Gender = req.InherentAttributes.PHI.Gender
			phiUpdate.DateOfBirth = req.InherentAttributes.PHI.DateOfBirth
			phiUpdate.ResidentPhone = req.InherentAttributes.PHI.ResidentPhone
			phiUpdate.ResidentEmail = req.InherentAttributes.PHI.ResidentEmail
			// SavePhone 和 SaveEmail 不是 domain.ResidentPHIUpdate 的字段，它们用于控制是否在 residents 表中保存明文
			// 这里只映射其他 PHI 字段
			phiUpdate.WeightLb = req.InherentAttributes.PHI.WeightLb
			phiUpdate.HeightFt = req.InherentAttributes.PHI.HeightFt
			phiUpdate.HeightIn = req.InherentAttributes.PHI.HeightIn
			phiUpdate.MobilityLevel = req.InherentAttributes.PHI.MobilityLevel
			phiUpdate.TremorStatus = req.InherentAttributes.PHI.TremorStatus
			phiUpdate.MobilityAid = req.InherentAttributes.PHI.MobilityAid
			phiUpdate.ADLAssistance = req.InherentAttributes.PHI.ADLAssistance
			phiUpdate.CommStatus = req.InherentAttributes.PHI.CommStatus
			phiUpdate.HasHypertension = req.InherentAttributes.PHI.HasHypertension
			phiUpdate.HasHyperlipaemia = req.InherentAttributes.PHI.HasHyperlipaemia
			phiUpdate.HasHyperglycaemia = req.InherentAttributes.PHI.HasHyperglycaemia
			phiUpdate.HasStrokeHistory = req.InherentAttributes.PHI.HasStrokeHistory
			phiUpdate.HasParalysis = req.InherentAttributes.PHI.HasParalysis
			phiUpdate.HasAlzheimer = req.InherentAttributes.PHI.HasAlzheimer
			phiUpdate.MedicalHistory = req.InherentAttributes.PHI.MedicalHistory
			phiUpdate.HomeAddressStreet = req.InherentAttributes.PHI.HomeAddressStreet
			phiUpdate.HomeAddressCity = req.InherentAttributes.PHI.HomeAddressCity
			phiUpdate.HomeAddressState = req.InherentAttributes.PHI.HomeAddressState
			phiUpdate.HomeAddressPostalCode = req.InherentAttributes.PHI.HomeAddressPostalCode
			phiUpdate.PlusCode = req.InherentAttributes.PHI.PlusCode

			// 处理 SaveEmail/SavePhone 逻辑：如果 SaveEmail/SavePhone 为 true，需要同时更新 residents 表的 email/phone
			saveEmail := req.InherentAttributes.PHI.SaveEmail
			savePhone := req.InherentAttributes.PHI.SavePhone
			if saveEmail != nil && saveEmail.Action == domain.UpdateActionUpdate && saveEmail.Value {
				if phiUpdate.ResidentEmail != nil && phiUpdate.ResidentEmail.Action == domain.UpdateActionUpdate && phiUpdate.ResidentEmail.Value != "" {
					// 需要在 residents 表中保存 email
					if residentUpdate.Email == nil {
						residentUpdate.Email = phiUpdate.ResidentEmail
					}
				}
			}
			if savePhone != nil && savePhone.Action == domain.UpdateActionUpdate && savePhone.Value {
				if phiUpdate.ResidentPhone != nil && phiUpdate.ResidentPhone.Action == domain.UpdateActionUpdate && phiUpdate.ResidentPhone.Value != "" {
					// 需要在 residents 表中保存 phone
					if residentUpdate.Phone == nil {
						residentUpdate.Phone = phiUpdate.ResidentPhone
					}
				}
			}

			// 调用 UpsertResidentPHIFields
			if err := s.residentsRepo.UpsertResidentPHIFields(ctx, req.TenantID, req.ResidentID, phiUpdate); err != nil {
				s.logger.Warn("UpsertResidentPHIFields failed",
					zap.String("tenant_id", req.TenantID),
					zap.String("resident_id", req.ResidentID),
					zap.Error(err),
				)
				// 不失败整个操作，只记录警告
			}

			// 如果 SaveEmail/SavePhone 导致需要更新 residents 表的 email/phone，需要再次调用 UpdateResidentFields
			if (saveEmail != nil && saveEmail.Action == domain.UpdateActionUpdate && saveEmail.Value) ||
				(savePhone != nil && savePhone.Action == domain.UpdateActionUpdate && savePhone.Value) {
				if hasResidentFields || residentUpdate.Email != nil || residentUpdate.Phone != nil {
					// residentUpdate 可能已经在上面被更新，如果新增了 Email/Phone，需要再次更新
					if err := s.residentsRepo.UpdateResidentFields(ctx, req.TenantID, req.ResidentID, residentUpdate); err != nil {
						s.logger.Warn("UpdateResidentFields for email/phone failed",
							zap.String("tenant_id", req.TenantID),
							zap.String("resident_id", req.ResidentID),
							zap.Error(err),
						)
						// 不失败整个操作，只记录警告
					}
				}
			}
		}

		// 3.3 处理 Contacts 更新
		if len(req.InherentAttributes.Contacts) > 0 {
			for _, contactReq := range req.InherentAttributes.Contacts {
				if contactReq.Slot == "" {
					continue // slot 是必填字段
				}

				contactUpdate := &domain.ResidentContactUpdate{}

				// 映射所有 Contact 字段
				contactUpdate.IsEnabled = contactReq.IsEnabled
				contactUpdate.Relationship = contactReq.Relationship
				contactUpdate.ContactFirstName = contactReq.ContactFirstName
				contactUpdate.ContactLastName = contactReq.ContactLastName
				contactUpdate.ContactPhone = contactReq.ContactPhone
				contactUpdate.ContactEmail = contactReq.ContactEmail
				contactUpdate.ReceiveSMS = contactReq.ReceiveSMS
				contactUpdate.ReceiveEmail = contactReq.ReceiveEmail
				contactUpdate.AlertTimeWindow = contactReq.AlertTimeWindow

				// 调用 UpdateResidentContactFields
				if err := s.residentsRepo.UpdateResidentContactFields(ctx, req.TenantID, req.ResidentID, contactReq.Slot, contactUpdate); err != nil {
					s.logger.Warn("UpdateResidentContactFields failed",
						zap.String("tenant_id", req.TenantID),
						zap.String("resident_id", req.ResidentID),
						zap.String("slot", contactReq.Slot),
						zap.Error(err),
					)
					// 不失败整个操作，只记录警告
				}
			}
		}
	}

	// 4. 处理 UnitRelation（unit_id, room_id, bed_id）
	// 若本请求中已因迁院区解绑，不再应用 UnitRelation，避免把旧 unit 写回
	if req.UnitRelation != nil && !branchChangedAndUnbound {
		residentUpdate := &domain.ResidentUpdate{}

		residentUpdate.UnitID = req.UnitRelation.UnitID
		residentUpdate.RoomID = req.UnitRelation.RoomID
		residentUpdate.BedID = req.UnitRelation.BedID

		// 检查是否有需要更新的字段
		hasUnitRelation := residentUpdate.UnitID != nil || residentUpdate.RoomID != nil || residentUpdate.BedID != nil

		if hasUnitRelation {
			// 验证 bed → room → unit 的层级关系
			if residentUpdate.BedID != nil && residentUpdate.BedID.Action == domain.UpdateActionUpdate && residentUpdate.BedID.Value != "" {
				// 如果指定了 bed_id，必须同时指定 room_id 和 unit_id
				if residentUpdate.RoomID == nil || residentUpdate.RoomID.Action != domain.UpdateActionUpdate || residentUpdate.RoomID.Value == "" {
					return nil, fmt.Errorf("bed_id requires room_id")
				}
				if residentUpdate.UnitID == nil || residentUpdate.UnitID.Action != domain.UpdateActionUpdate || residentUpdate.UnitID.Value == "" {
					return nil, fmt.Errorf("bed_id requires unit_id")
				}
			}
			if residentUpdate.RoomID != nil && residentUpdate.RoomID.Action == domain.UpdateActionUpdate && residentUpdate.RoomID.Value != "" {
				// 如果指定了 room_id，必须同时指定 unit_id
				if residentUpdate.UnitID == nil || residentUpdate.UnitID.Action != domain.UpdateActionUpdate || residentUpdate.UnitID.Value == "" {
					return nil, fmt.Errorf("room_id requires unit_id")
				}
			}
			// Public/VIP/Share 绑定规则：Public、VIP 必须 room；Share 必须 bed
			if residentUpdate.UnitID != nil && residentUpdate.UnitID.Action == domain.UpdateActionUpdate && residentUpdate.UnitID.Value != "" && s.db != nil {
				var isPublic, isShared bool
				err := s.db.QueryRowContext(ctx, `SELECT is_public, is_shared_unit FROM units WHERE tenant_id = $1 AND unit_id = $2`, req.TenantID, residentUpdate.UnitID.Value).Scan(&isPublic, &isShared)
				if err == nil {
					if isShared {
						// Share：必须指定到床
						hasBed := residentUpdate.BedID != nil && residentUpdate.BedID.Action == domain.UpdateActionUpdate && residentUpdate.BedID.Value != ""
						if !hasBed {
							return nil, fmt.Errorf("Share unit requires bed_id")
						}
					} else {
						// Public、VIP：必须指定 room
						if residentUpdate.RoomID == nil || residentUpdate.RoomID.Action != domain.UpdateActionUpdate || residentUpdate.RoomID.Value == "" {
							return nil, fmt.Errorf("unit requires room_id")
						}
					}
				}
			}

			// 调用 UpdateResidentFields
			if err := s.residentsRepo.UpdateResidentFields(ctx, req.TenantID, req.ResidentID, residentUpdate); err != nil {
				s.logger.Error("UpdateResidentFields for unit relation failed",
					zap.String("tenant_id", req.TenantID),
					zap.String("resident_id", req.ResidentID),
					zap.Error(err),
				)
				return nil, fmt.Errorf("failed to update resident unit relation: %w", err)
			}
		}
	}

	// 5. 处理 CaregiverRelation（user_list, group_list）
	// 若本请求中已因迁院区解绑，清空 caregiver/caregiver-group，且不再应用请求中的 CaregiverRelation
	if branchChangedAndUnbound {
		caregiverUpdate := &domain.ResidentCaregiverUpdate{
			UserList:  &domain.UpdateJSON{Action: domain.UpdateActionDelete},
			GroupList: &domain.UpdateJSON{Action: domain.UpdateActionDelete},
		}
		if err := s.residentsRepo.UpsertResidentCaregiverFields(ctx, req.TenantID, req.ResidentID, caregiverUpdate); err != nil {
			s.logger.Warn("UpsertResidentCaregiverFields (clear on branch change) failed",
				zap.String("tenant_id", req.TenantID),
				zap.String("resident_id", req.ResidentID),
				zap.Error(err),
			)
		}
	} else if req.CaregiverRelation != nil {
		caregiverUpdate := &domain.ResidentCaregiverUpdate{}

		caregiverUpdate.UserList = req.CaregiverRelation.UserList
		caregiverUpdate.GroupList = req.CaregiverRelation.GroupList

		// 检查是否有需要更新的字段
		hasCaregiverRelation := caregiverUpdate.UserList != nil || caregiverUpdate.GroupList != nil

		if hasCaregiverRelation {
			// 调用 UpsertResidentCaregiverFields
			if err := s.residentsRepo.UpsertResidentCaregiverFields(ctx, req.TenantID, req.ResidentID, caregiverUpdate); err != nil {
				s.logger.Warn("UpsertResidentCaregiverFields failed",
					zap.String("tenant_id", req.TenantID),
					zap.String("resident_id", req.ResidentID),
					zap.Error(err),
				)
				// 不失败整个操作，只记录警告
			} else {
				if true {
					updatedResident, err := s.residentsRepo.GetResident(ctx, req.TenantID, req.ResidentID)
					if err != nil {
						s.logger.Warn("Failed to get updated resident for card sync", zap.Error(err), zap.String("tenant_id", req.TenantID), zap.String("resident_id", req.ResidentID))
					} else if updatedResident != nil && updatedResident.UnitID != "" {
						SyncUnitCards(ctx, req.TenantID, updatedResident.UnitID)
						// 护理关系变更快照
						if err := TakeBindingSnapshot(ctx, s.db, s.logger, req.TenantID, "caregiver_change", req.ResidentID,
							fmt.Sprintf("caregiver changed for resident %s", req.ResidentID),
							[]string{updatedResident.UnitID}, req.CurrentUserID); err != nil {
							s.logger.Warn("TakeBindingSnapshot failed", zap.Error(err))
						}
					}
				}
			}
		}
	}

	{
		updatedResident, err := s.residentsRepo.GetResident(ctx, req.TenantID, req.ResidentID)
		if err != nil {
			s.logger.Warn("Failed to get updated resident for card sync", zap.Error(err), zap.String("tenant_id", req.TenantID), zap.String("resident_id", req.ResidentID))
		} else {
			newUnit := ""
			if updatedResident != nil {
				newUnit = updatedResident.UnitID
			}
			SyncUnitCards(ctx, req.TenantID, existingResident.UnitID)
			SyncUnitCards(ctx, req.TenantID, newUnit)
		}
	}

	// 绑定关系快照：住户 unit/room/bed 发生变更时
	updatedResident2, _ := s.residentsRepo.GetResident(ctx, req.TenantID, req.ResidentID)
	if updatedResident2 != nil && existingResident != nil {
		oldUnit := existingResident.UnitID
		newUnit2 := updatedResident2.UnitID
		oldBed := existingResident.BedID
		newBed := updatedResident2.BedID
		if oldUnit != newUnit2 || oldBed != newBed {
			var affectedUnits []string
			if oldUnit != "" {
				affectedUnits = append(affectedUnits, oldUnit)
			}
			if newUnit2 != "" && newUnit2 != oldUnit {
				affectedUnits = append(affectedUnits, newUnit2)
			}
			name := updatedResident2.Nickname
			if name == "" {
				name = req.ResidentID
			}
			oldLoc := snapshotUnitLabel(ctx, s.db, req.TenantID, oldUnit)
			newLoc := snapshotUnitLabel(ctx, s.db, req.TenantID, newUnit2)
			summary := fmt.Sprintf("%s: %s → %s", name, oldLoc, newLoc)
			if err := TakeBindingSnapshot(ctx, s.db, s.logger, req.TenantID, "resident_move", req.ResidentID, summary, affectedUnits, req.CurrentUserID); err != nil {
				s.logger.Warn("TakeBindingSnapshot failed", zap.Error(err))
			}
		}
	}

	return &UpdateResidentResponse{
		Success: true,
	}, nil
}

// DeleteResident 删除住户（软删除）
func (s *residentService) DeleteResident(ctx context.Context, req DeleteResidentRequest) (*DeleteResidentResponse, error) {
	// ========== 1. 参数验证 ==========
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.ResidentID == "" {
		return nil, fmt.Errorf("resident_id is required")
	}
	if req.CurrentUserID == "" {
		return nil, fmt.Errorf("current_user_id is required")
	}

	// ========== 2. 验证用户身份和权限 ==========

	// 2.1 查询用户信息（不信任前端传入的数据）
	var dbTenantID, dbRole, dbStatus string
	err := s.db.QueryRowContext(ctx,
		`SELECT tenant_id::text, role, status 
		 FROM users 
		 WHERE user_id::text = $1`,
		req.CurrentUserID,
	).Scan(&dbTenantID, &dbRole, &dbStatus)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	// 验证用户状态
	if dbStatus != "active" {
		return nil, fmt.Errorf("user is not active")
	}

	// 2.2 验证 tenant_id 一致性
	if req.TenantID != dbTenantID {
		return nil, fmt.Errorf("tenant_id mismatch")
	}
	tenantID := dbTenantID

	// 2.3 角色验证
	if dbRole == "Individual" {
		return nil, fmt.Errorf("individual users cannot delete residents")
	}

	// ========== 3. 验证住户存在 ==========
	var residentTenantID, residentStatus string
	err = s.db.QueryRowContext(ctx,
		`SELECT tenant_id::text, status 
		 FROM residents 
		 WHERE resident_id::text = $1`,
		req.ResidentID,
	).Scan(&residentTenantID, &residentStatus)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("resident not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query resident: %w", err)
	}

	// 验证 tenant_id 一致性
	if residentTenantID != tenantID {
		return nil, fmt.Errorf("resident tenant_id mismatch")
	}

	// ========== 4. 权限检查（AssignedOnly, BranchOnly）==========

	// 4.1 AssignedOnly 检查
	if req.PermissionCheck != nil && req.PermissionCheck.AssignedOnly {
		var isAssigned bool
		err := s.db.QueryRowContext(ctx,
			`SELECT EXISTS(
					SELECT 1 FROM resident_caregivers rc
					WHERE rc.tenant_id = $1
					  AND rc.resident_id::text = $2
					  AND (rc.user_list::text LIKE $3 OR rc.user_list::text LIKE $4)
				)`,
			tenantID, req.ResidentID, req.CurrentUserID, "%\""+req.CurrentUserID+"\"%",
		).Scan(&isAssigned)
		if err != nil {
			return nil, fmt.Errorf("failed to check assignment: %w", err)
		}
		if !isAssigned {
			return nil, fmt.Errorf("permission denied: can only delete assigned residents")
		}
	}

	// 4.2 BranchOnly 检查
	if req.PermissionCheck != nil && req.PermissionCheck.BranchOnly {
		// 获取住户的 branch_id
		var residentBranchID sql.NullString
		err := s.db.QueryRowContext(ctx,
			`SELECT branch_id::text 
			 FROM residents 
			 WHERE tenant_id = $1 AND resident_id::text = $2`,
			tenantID, req.ResidentID,
		).Scan(&residentBranchID)
		if err != nil {
			return nil, fmt.Errorf("failed to get resident branch_id: %w", err)
		}

		rows, err := s.db.QueryContext(ctx,
			`SELECT branch_id::text FROM user_branches WHERE user_id::text = $1`,
			req.CurrentUserID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get user branches: %w", err)
		}
		defer rows.Close()
		var userBranchIDs []string
		for rows.Next() {
			var bid string
			if err := rows.Scan(&bid); err == nil {
				userBranchIDs = append(userBranchIDs, bid)
			}
		}

		if len(userBranchIDs) > 0 {
			if !residentBranchID.Valid {
				return nil, fmt.Errorf("permission denied: can only delete residents in the same branch")
			}
			found := false
			for _, bid := range userBranchIDs {
				if bid == residentBranchID.String {
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("permission denied: can only delete residents in the same branch")
			}
		} else {
			if residentBranchID.Valid {
				return nil, fmt.Errorf("permission denied: can only delete residents without branch_id")
			}
		}
	}

	var unitIDBeforeDelete string
	if err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(unit_id::text, '') FROM residents WHERE tenant_id = $1 AND resident_id::text = $2`,
		tenantID, req.ResidentID,
	).Scan(&unitIDBeforeDelete); err != nil {
		unitIDBeforeDelete = ""
	}

	// ========== 5. 软删除：将 status 设置为 'discharged' ==========
	residentUpdate := &domain.ResidentUpdate{
		Status: &domain.UpdateString{
			Action: domain.UpdateActionUpdate,
			Value:  "discharged",
		},
	}

	err = s.residentsRepo.UpdateResidentFields(ctx, tenantID, req.ResidentID, residentUpdate)
	if err != nil {
		s.logger.Error("DeleteResident failed",
			zap.String("tenant_id", tenantID),
			zap.String("resident_id", req.ResidentID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to delete resident: %w", err)
	}

	SyncUnitCards(ctx, tenantID, unitIDBeforeDelete)

	return &DeleteResidentResponse{Success: true}, nil
}

// ResetResidentPassword 重置住户密码
func (s *residentService) ResetResidentPassword(ctx context.Context, req ResetResidentPasswordRequest) (*ResetResidentPasswordResponse, error) {
	// ========== 1. 参数验证 ==========
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.ResidentID == "" {
		return nil, fmt.Errorf("resident_id is required")
	}
	if req.CurrentUserID == "" {
		return nil, fmt.Errorf("current_user_id is required")
	}

	// ========== 2. 验证用户身份和权限 ==========

	// 2.1 查询用户信息（不信任前端传入的数据）
	var dbTenantID, dbRole, dbStatus string
	err := s.db.QueryRowContext(ctx,
		`SELECT tenant_id::text, role, status 
		 FROM users 
		 WHERE user_id::text = $1`,
		req.CurrentUserID,
	).Scan(&dbTenantID, &dbRole, &dbStatus)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	// 验证用户状态
	if dbStatus != "active" {
		return nil, fmt.Errorf("user is not active")
	}

	// 2.2 验证 tenant_id 一致性
	if req.TenantID != dbTenantID {
		return nil, fmt.Errorf("tenant_id mismatch")
	}
	tenantID := dbTenantID

	// 2.3 权限检查
	if dbRole == "Individual" {
		// Individual 用户可以重置自己的密码（通过 resident_id 匹配）
		if req.CurrentUserID != req.ResidentID {
			return nil, fmt.Errorf("access denied: individual users can only reset own password")
		}
	} else if dbRole == "Resident" || dbRole == "Family" {
		// Resident/Family 用户只能重置自己的密码
		if req.CurrentUserID != req.ResidentID {
			return nil, fmt.Errorf("access denied: can only reset own password")
		}
	}
	// Staff 用户（Admin, Manager, Caregiver, IT）可以重置其他住户的密码（需要后续权限检查）

	// ========== 3. 验证住户存在 ==========
	var residentTenantID string
	err = s.db.QueryRowContext(ctx,
		`SELECT tenant_id::text 
		 FROM residents 
		 WHERE resident_id::text = $1`,
		req.ResidentID,
	).Scan(&residentTenantID)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("resident not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query resident: %w", err)
	}

	// 验证 tenant_id 一致性
	if residentTenantID != tenantID {
		return nil, fmt.Errorf("resident tenant_id mismatch")
	}

	// ========== 4. Staff 权限检查（AssignedOnly, BranchOnly）==========

	// 只有 Staff 用户才需要检查 AssignedOnly 和 BranchOnly
	if dbRole != "Individual" && dbRole != "Resident" && dbRole != "Family" {
		// 4.1 AssignedOnly 检查
		if req.PermissionCheck != nil && req.PermissionCheck.AssignedOnly {
			var isAssigned bool
			err := s.db.QueryRowContext(ctx,
				`SELECT EXISTS(
						SELECT 1 FROM resident_caregivers rc
						WHERE rc.tenant_id = $1
						  AND rc.resident_id::text = $2
						  AND (rc.user_list::text LIKE $3 OR rc.user_list::text LIKE $4)
					)`,
				tenantID, req.ResidentID, req.CurrentUserID, "%\""+req.CurrentUserID+"\"%",
			).Scan(&isAssigned)
			if err != nil {
				return nil, fmt.Errorf("failed to check assignment: %w", err)
			}
			if !isAssigned {
				return nil, fmt.Errorf("permission denied: can only reset password for assigned residents")
			}
		}

		// 4.2 BranchOnly 检查
		if req.PermissionCheck != nil && req.PermissionCheck.BranchOnly {
			// 获取住户的 branch_id
			var residentBranchID sql.NullString
			err := s.db.QueryRowContext(ctx,
				`SELECT branch_id::text 
				 FROM residents 
				 WHERE tenant_id = $1 AND resident_id::text = $2`,
				tenantID, req.ResidentID,
			).Scan(&residentBranchID)
			if err != nil {
				return nil, fmt.Errorf("failed to get resident branch_id: %w", err)
			}

			rows, err := s.db.QueryContext(ctx,
				`SELECT branch_id::text FROM user_branches WHERE user_id::text = $1`,
				req.CurrentUserID,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to get user branches: %w", err)
			}
			defer rows.Close()
			var userBranchIDs []string
			for rows.Next() {
				var bid string
				if err := rows.Scan(&bid); err == nil {
					userBranchIDs = append(userBranchIDs, bid)
				}
			}

			if len(userBranchIDs) > 0 {
				if !residentBranchID.Valid {
					return nil, fmt.Errorf("permission denied: can only reset password for residents in the same branch")
				}
				found := false
				for _, bid := range userBranchIDs {
					if bid == residentBranchID.String {
						found = true
						break
					}
				}
				if !found {
					return nil, fmt.Errorf("permission denied: can only reset password for residents in the same branch")
				}
			} else {
				if residentBranchID.Valid {
					return nil, fmt.Errorf("permission denied: can only reset password for residents without branch_id")
				}
			}
		}
	}

	// ========== 5. 密码哈希（前端已 hash，这里直接解码 hex 字符串）==========
	// 前端发送的是 SHA256(password) 的 hex 字符串，直接解码为 byte slice
	if req.NewPassword == "" {
		return nil, fmt.Errorf("password hash is required")
	}
	passwordHash, err := hex.DecodeString(req.NewPassword)
	if err != nil || len(passwordHash) == 0 {
		return nil, fmt.Errorf("failed to decode password hash: %w", err)
	}

	// ========== 6. 使用 UpdateResidentFields 更新 password_hash ==========
	residentUpdate := &domain.ResidentUpdate{
		PasswordHash: &domain.UpdateBytes{
			Action: domain.UpdateActionUpdate,
			Value:  passwordHash,
		},
	}

	err = s.residentsRepo.UpdateResidentFields(ctx, tenantID, req.ResidentID, residentUpdate)
	if err != nil {
		s.logger.Error("ResetResidentPassword failed",
			zap.String("tenant_id", tenantID),
			zap.String("resident_id", req.ResidentID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to reset password: %w", err)
	}

	return &ResetResidentPasswordResponse{
		Success:     true,
		NewPassword: "", // 不再返回密码（前端已 hash，后端不存储明文）
	}, nil
}

// generateRandomPassword 生成随机密码
func generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// ============================================
// GetResidentAccountSettings 获取住户/联系人账户设置
// ============================================

// GetResidentAccountSettings 获取住户账户设置（只返回账户设置相关字段）
// 注意：这个 API 只能查看自己的账户设置，不允许查看其他用户的
// 注意：联系人不能登录系统，所以不支持联系人的账户设置
func (s *residentService) GetResidentAccountSettings(ctx context.Context, req GetResidentAccountSettingsRequest) (*GetResidentAccountSettingsResponse, error) {
	// 1. 权限检查：只能查看自己的账户设置
	if req.CurrentUserID != req.ResidentID {
		return nil, fmt.Errorf("permission denied: can only view own account settings")
	}

	// 2. 如果是 Family 角色，不支持（联系人不登录，不需要账户设置）
	if req.CurrentUserRole == "Family" {
		return nil, fmt.Errorf("contacts do not log in, account settings are not available for contacts")
	}

	// 3. 构建响应（只支持 Resident）
	resp := &GetResidentAccountSettingsResponse{
		IsContact: false,
	}

	// 4. Resident: 从 residents 和 resident_phi 表获取
	var residentAccount, nickname sql.NullString
	var residentEmail, residentPhone sql.NullString

	err := s.db.QueryRowContext(ctx,
		`SELECT 
				r.resident_account,
				COALESCE(r.nickname, '') as nickname,
			COALESCE(rp.resident_email, '') as resident_email,
			COALESCE(rp.resident_phone, '') as resident_phone
			 FROM residents r
			 LEFT JOIN resident_phi rp ON rp.resident_id = r.resident_id AND rp.tenant_id = r.tenant_id
			 WHERE r.tenant_id = $1 AND r.resident_id::text = $2`,
		req.TenantID, req.ResidentID,
	).Scan(&residentAccount, &nickname, &residentEmail, &residentPhone)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("resident not found")
		}
		return nil, fmt.Errorf("failed to get resident account settings: %w", err)
	}

	if residentAccount.Valid {
		account := residentAccount.String
		resp.ResidentAccount = &account
	}
	if nickname.Valid {
		resp.Nickname = nickname.String
	}
	if residentEmail.Valid && residentEmail.String != "" && residentEmail.String != "***@***" {
		email := residentEmail.String
		resp.Email = &email
		resp.SaveEmail = true // 如果存在且不是占位符，说明已保存
	}
	if residentPhone.Valid && residentPhone.String != "" && residentPhone.String != "xxx-xxx-xxxx" {
		phone := residentPhone.String
		resp.Phone = &phone
		resp.SavePhone = true // 如果存在且不是占位符，说明已保存
	}

	return resp, nil
}

// ============================================
// UpdateResidentAccountSettings 更新住户/联系人账户设置（统一 API）
// ============================================

// UpdateResidentAccountSettings 更新住户账户设置（在同一个事务中处理所有更新）
// 注意：这个 API 只能更新自己的账户设置，不允许更新其他用户的
// 注意：联系人不能登录系统，所以不支持联系人的账户设置
func (s *residentService) UpdateResidentAccountSettings(ctx context.Context, req UpdateResidentAccountSettingsRequest) (*UpdateResidentAccountSettingsResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" || req.ResidentID == "" || req.CurrentUserID == "" {
		return nil, fmt.Errorf("tenant_id, resident_id, and current_user_id are required")
	}

	// 2. 权限检查：只能更新自己的账户设置
	if req.CurrentUserID != req.ResidentID {
		return nil, fmt.Errorf("permission denied: can only update own account settings")
	}

	// 3. 如果是 Family 角色，不支持（联系人不登录，不需要账户设置）
	if req.CurrentUserRole == "Family" {
		return nil, fmt.Errorf("contacts do not log in, account settings cannot be updated for contacts")
	}

	// 4. 开始事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 5. Resident: 更新 residents 和 resident_phi 表
	{
		// 5.1 更新密码（residents 表，如果提供，!= nil 就更新，不进行任何判断）
		if req.PasswordHash != nil {
			passwordHashBytes, err := hex.DecodeString(*req.PasswordHash)
			if err != nil || len(passwordHashBytes) == 0 {
				return nil, fmt.Errorf("failed to decode password hash: %w", err)
			}
			_, err = tx.ExecContext(ctx,
				`UPDATE residents SET password_hash = $1 WHERE tenant_id = $2 AND resident_id::text = $3`,
				passwordHashBytes, req.TenantID, req.ResidentID,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to update password: %w", err)
			}
		}

		// 5.2 更新 email/phone hash（residents 表，用于登录）
		residentUpdates := []string{}
		residentArgs := []interface{}{}
		residentArgIdx := 1

		// 更新 email_hash（如果提供，!= nil 就更新，不进行任何判断，直接传值）
		if req.EmailHash != nil {
			emailHashBytes, err := hex.DecodeString(*req.EmailHash)
			if err != nil || len(emailHashBytes) == 0 {
				return nil, fmt.Errorf("failed to decode email hash: %w", err)
			}
			residentUpdates = append(residentUpdates, fmt.Sprintf("email_hash = $%d", residentArgIdx))
			residentArgs = append(residentArgs, emailHashBytes)
			residentArgIdx++
		}

		// 更新 phone_hash（如果提供，!= nil 就更新，不进行任何判断，直接传值）
		if req.PhoneHash != nil {
			phoneHashBytes, err := hex.DecodeString(*req.PhoneHash)
			if err != nil || len(phoneHashBytes) == 0 {
				return nil, fmt.Errorf("failed to decode phone hash: %w", err)
			}
			residentUpdates = append(residentUpdates, fmt.Sprintf("phone_hash = $%d", residentArgIdx))
			residentArgs = append(residentArgs, phoneHashBytes)
			residentArgIdx++
		}

		if len(residentUpdates) > 0 {
			query := fmt.Sprintf(
				`UPDATE residents SET %s WHERE tenant_id = $%d AND resident_id::text = $%d`,
				strings.Join(residentUpdates, ", "), residentArgIdx, residentArgIdx+1,
			)
			residentArgs = append(residentArgs, req.TenantID, req.ResidentID)
			_, err = tx.ExecContext(ctx, query, residentArgs...)
			if err != nil {
				return nil, fmt.Errorf("failed to update resident: %w", err)
			}
		}

		// 5.3 更新 email/phone 明文（resident_phi 表，根据 save 标志）
		if req.Email != nil || req.Phone != nil {
			// 检查 resident_phi 是否存在
			var phiExists bool
			err = tx.QueryRowContext(ctx,
				`SELECT EXISTS(SELECT 1 FROM resident_phi WHERE tenant_id = $1 AND resident_id::text = $2)`,
				req.TenantID, req.ResidentID,
			).Scan(&phiExists)
			if err != nil {
				return nil, fmt.Errorf("failed to check resident_phi: %w", err)
			}

			phiUpdates := []string{}
			phiArgs := []interface{}{}
			phiArgIdx := 1

			// 更新 email 明文（如果提供，!= nil 就更新，不进行任何判断，直接传值）
			if req.Email != nil {
				phiUpdates = append(phiUpdates, fmt.Sprintf("resident_email = $%d", phiArgIdx))
				phiArgs = append(phiArgs, *req.Email)
				phiArgIdx++
			}

			// 更新 phone 明文（如果提供，!= nil 就更新，不进行任何判断，直接传值）
			if req.Phone != nil {
				phiUpdates = append(phiUpdates, fmt.Sprintf("resident_phone = $%d", phiArgIdx))
				phiArgs = append(phiArgs, *req.Phone)
				phiArgIdx++
			}

			if len(phiUpdates) > 0 {
				if phiExists {
					// 更新现有记录
					query := fmt.Sprintf(
						`UPDATE resident_phi SET %s WHERE tenant_id = $%d AND resident_id::text = $%d`,
						strings.Join(phiUpdates, ", "), phiArgIdx, phiArgIdx+1,
					)
					phiArgs = append(phiArgs, req.TenantID, req.ResidentID)
					_, err = tx.ExecContext(ctx, query, phiArgs...)
					if err != nil {
						return nil, fmt.Errorf("failed to update resident_phi: %w", err)
					}
				} else {
					// 需要创建新记录（只有当 email 或 phone 不为空时才创建）
					shouldCreate := false
					createEmail := interface{}(nil)
					createPhone := interface{}(nil)

					if req.Email != nil && *req.Email != "" {
						shouldCreate = true
						createEmail = *req.Email
					}
					if req.Phone != nil && *req.Phone != "" {
						shouldCreate = true
						createPhone = *req.Phone
					}

					if shouldCreate {
						// 创建新记录
						insertFields := []string{"tenant_id", "resident_id"}
						insertValues := []string{"$1", "$2"}
						insertArgs := []interface{}{req.TenantID, req.ResidentID}
						argIdx := 3

						if createEmail != nil {
							insertFields = append(insertFields, "resident_email")
							insertValues = append(insertValues, fmt.Sprintf("$%d", argIdx))
							insertArgs = append(insertArgs, createEmail)
							argIdx++
						}
						if createPhone != nil {
							insertFields = append(insertFields, "resident_phone")
							insertValues = append(insertValues, fmt.Sprintf("$%d", argIdx))
							insertArgs = append(insertArgs, createPhone)
							argIdx++
						}

						query := fmt.Sprintf(
							`INSERT INTO resident_phi (%s) VALUES (%s)`,
							strings.Join(insertFields, ", "), strings.Join(insertValues, ", "),
						)
						_, err = tx.ExecContext(ctx, query, insertArgs...)
						if err != nil {
							return nil, fmt.Errorf("failed to create resident_phi: %w", err)
						}
					}
				}
			}
		}
	}

	// 5. 提交事务
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &UpdateResidentAccountSettingsResponse{
		Success: true,
		Message: "Account settings updated successfully",
	}, nil
}

// UpdateResidentContact 更新住户联系人信息
func (s *residentService) UpdateResidentContact(ctx context.Context, req UpdateResidentContactStandaloneRequest) (*UpdateResidentContactResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.ResidentID == "" {
		return nil, fmt.Errorf("resident_id is required")
	}
	if req.Slot == "" {
		return nil, fmt.Errorf("slot is required")
	}

	// 2. 权限检查（细粒度）
	if req.CurrentUserRole == "Resident" {
		// Resident: 只能更新自己的联系人
		if req.CurrentUserID != req.ResidentID {
			return nil, fmt.Errorf("permission denied: can only update contacts for own resident")
		}
		// 允许更新
	} else if req.CurrentUserRole == "Admin" {
		// Admin: 先检查 accountID 的角色是 Admin，然后允许更新所有 contact
		var userRole sql.NullString
		err := s.db.QueryRowContext(ctx,
			`SELECT role FROM users WHERE tenant_id = $1 AND user_id::text = $2`,
			req.TenantID, req.CurrentUserID,
		).Scan(&userRole)
		if err != nil {
			return nil, fmt.Errorf("failed to verify user role: %w", err)
		}
		if !userRole.Valid || userRole.String != "Admin" {
			return nil, fmt.Errorf("access denied: user role is not Admin")
		}
		// 允许更新所有 contact
	} else if req.CurrentUserRole == "Manager" {
		// Manager: UpdateResidentContact 不需要 branch 检查逻辑
		// 因为在 resident 页面进入时，ResidentContact 视为 resident 的信息
		// 允许更新（Manager 可以更新所有 contact，因为 contact 被视为 resident 的一部分）
		// 允许更新
	} else {
		// 其他角色（Family, Nurse, Caregiver 等）：拒绝
		return nil, fmt.Errorf("permission denied: only Admin, Manager, and Resident can update contacts")
	}

	// 3. 检查 contact 是否存在（通过 resident_id + slot 定位）
	// 如果不存在，使用 UpdateResidentContactFields 时会自动创建（repository 层处理）
	var contactExists bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM resident_contacts WHERE tenant_id = $1 AND resident_id::text = $2 AND slot = $3)`,
		req.TenantID, req.ResidentID, req.Slot,
	).Scan(&contactExists)
	if err != nil {
		return nil, fmt.Errorf("failed to check contact existence: %w", err)
	}
	if !contactExists {
		// Contact 不存在，需要先创建
		newContact := &domain.ResidentContact{
			Slot:      req.Slot,
			IsEnabled: false, // 默认禁用
		}
		if req.IsEnabled != nil {
			newContact.IsEnabled = *req.IsEnabled
		}
		if req.Relationship != nil {
			newContact.Relationship = sql.NullString{String: *req.Relationship, Valid: *req.Relationship != ""}
		}
		if req.ContactFirstName != nil {
			newContact.ContactFirstName = sql.NullString{String: *req.ContactFirstName, Valid: *req.ContactFirstName != ""}
		}
		if req.ContactLastName != nil {
			newContact.ContactLastName = sql.NullString{String: *req.ContactLastName, Valid: *req.ContactLastName != ""}
		}
		if req.ContactPhone != nil {
			newContact.ContactPhone = sql.NullString{String: *req.ContactPhone, Valid: *req.ContactPhone != ""}
		}
		if req.ContactEmail != nil {
			newContact.ContactEmail = sql.NullString{String: *req.ContactEmail, Valid: *req.ContactEmail != ""}
		}
		if req.ReceiveSMS != nil {
			newContact.ReceiveSMS = *req.ReceiveSMS
		}
		if req.ReceiveEmail != nil {
			newContact.ReceiveEmail = *req.ReceiveEmail
		}

		_, err = s.residentsRepo.CreateResidentContact(ctx, req.TenantID, req.ResidentID, newContact)
		if err != nil {
			return nil, fmt.Errorf("failed to create contact: %w", err)
		}
	}

	// 4. 构建 domain.ResidentContact 对象
	// 注意：联系人不登录系统，不再使用 phone_hash, email_hash, password_hash 字段
	contact := &domain.ResidentContact{
		Slot: req.Slot, // slot 是必填的，直接使用
	}
	if req.IsEnabled != nil {
		contact.IsEnabled = *req.IsEnabled
	}
	if req.Relationship != nil {
		contact.Relationship = sql.NullString{String: *req.Relationship, Valid: *req.Relationship != ""}
	}
	if req.ContactFirstName != nil {
		contact.ContactFirstName = sql.NullString{String: *req.ContactFirstName, Valid: *req.ContactFirstName != ""}
	}
	if req.ContactLastName != nil {
		contact.ContactLastName = sql.NullString{String: *req.ContactLastName, Valid: *req.ContactLastName != ""}
	}
	if req.ContactPhone != nil {
		contact.ContactPhone = sql.NullString{String: *req.ContactPhone, Valid: *req.ContactPhone != ""}
	}
	if req.ContactEmail != nil {
		contact.ContactEmail = sql.NullString{String: *req.ContactEmail, Valid: *req.ContactEmail != ""}
	}
	if req.ReceiveSMS != nil {
		contact.ReceiveSMS = *req.ReceiveSMS
	}
	if req.ReceiveEmail != nil {
		contact.ReceiveEmail = *req.ReceiveEmail
	}

	// 5. 调用 Repository 更新（使用 UpdateResidentContactFields）
	contactUpdate := &domain.ResidentContactUpdate{
		IsEnabled:        nil,
		Relationship:     nil,
		ContactFirstName: nil,
		ContactLastName:  nil,
		ContactPhone:     nil,
		ContactEmail:     nil,
		ContactPhoneHash: nil,
		ContactEmailHash: nil,
		ReceiveSMS:       nil,
		ReceiveEmail:     nil,
		AlertTimeWindow:  nil,
	}
	if req.IsEnabled != nil {
		contactUpdate.IsEnabled = &domain.UpdateBool{Action: domain.UpdateActionUpdate, Value: *req.IsEnabled}
	}
	if req.Relationship != nil {
		if *req.Relationship == "" {
			contactUpdate.Relationship = &domain.UpdateString{Action: domain.UpdateActionDelete}
		} else {
			contactUpdate.Relationship = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: *req.Relationship}
		}
	}
	if req.ContactFirstName != nil {
		if *req.ContactFirstName == "" {
			contactUpdate.ContactFirstName = &domain.UpdateString{Action: domain.UpdateActionDelete}
		} else {
			contactUpdate.ContactFirstName = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: *req.ContactFirstName}
		}
	}
	if req.ContactLastName != nil {
		if *req.ContactLastName == "" {
			contactUpdate.ContactLastName = &domain.UpdateString{Action: domain.UpdateActionDelete}
		} else {
			contactUpdate.ContactLastName = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: *req.ContactLastName}
		}
	}
	if req.ContactPhone != nil {
		if *req.ContactPhone == "" {
			contactUpdate.ContactPhone = &domain.UpdateString{Action: domain.UpdateActionDelete}
			contactUpdate.ContactPhoneHash = &domain.UpdateBytes{Action: domain.UpdateActionDelete}
		} else {
			contactUpdate.ContactPhone = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: *req.ContactPhone}
			// 处理 phone_hash
			if req.PhoneHash != nil && *req.PhoneHash != "" {
				phoneHash, err := hex.DecodeString(*req.PhoneHash)
				if err == nil && len(phoneHash) > 0 {
					contactUpdate.ContactPhoneHash = &domain.UpdateBytes{Action: domain.UpdateActionUpdate, Value: phoneHash}
				}
			} else {
				// 如果前端没有提供 hash，但提供了 phone，则计算 hash
				phone := strings.ToLower(strings.TrimSpace(*req.ContactPhone))
				phoneHashHex := HashAccount(phone)
				phoneHash, err := hex.DecodeString(phoneHashHex)
				if err == nil && len(phoneHash) > 0 {
					contactUpdate.ContactPhoneHash = &domain.UpdateBytes{Action: domain.UpdateActionUpdate, Value: phoneHash}
				}
			}
		}
	}
	if req.ContactEmail != nil {
		if *req.ContactEmail == "" {
			contactUpdate.ContactEmail = &domain.UpdateString{Action: domain.UpdateActionDelete}
			contactUpdate.ContactEmailHash = &domain.UpdateBytes{Action: domain.UpdateActionDelete}
		} else {
			contactUpdate.ContactEmail = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: *req.ContactEmail}
			// 处理 email_hash
			if req.EmailHash != nil && *req.EmailHash != "" {
				emailHash, err := hex.DecodeString(*req.EmailHash)
				if err == nil && len(emailHash) > 0 {
					contactUpdate.ContactEmailHash = &domain.UpdateBytes{Action: domain.UpdateActionUpdate, Value: emailHash}
				}
			} else {
				// 如果前端没有提供 hash，但提供了 email，则计算 hash
				email := strings.ToLower(strings.TrimSpace(*req.ContactEmail))
				emailHashHex := HashAccount(email)
				emailHash, err := hex.DecodeString(emailHashHex)
				if err == nil && len(emailHash) > 0 {
					contactUpdate.ContactEmailHash = &domain.UpdateBytes{Action: domain.UpdateActionUpdate, Value: emailHash}
				}
			}
		}
	}
	if req.ReceiveSMS != nil {
		contactUpdate.ReceiveSMS = &domain.UpdateBool{Action: domain.UpdateActionUpdate, Value: *req.ReceiveSMS}
	}
	if req.ReceiveEmail != nil {
		contactUpdate.ReceiveEmail = &domain.UpdateBool{Action: domain.UpdateActionUpdate, Value: *req.ReceiveEmail}
	}

	err = s.residentsRepo.UpdateResidentContactFields(ctx, req.TenantID, req.ResidentID, req.Slot, contactUpdate)
	if err != nil {
		return nil, fmt.Errorf("failed to update contact: %w", err)
	}

	return &UpdateResidentContactResponse{
		Success: true,
	}, nil
}
