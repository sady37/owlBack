package service

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"

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
	UpdateResidentContact(ctx context.Context, req UpdateResidentContactRequest) (*UpdateResidentContactResponse, error)

	// 删除
	DeleteResident(ctx context.Context, req DeleteResidentRequest) (*DeleteResidentResponse, error)

	// 密码管理
	ResetResidentPassword(ctx context.Context, req ResetResidentPasswordRequest) (*ResetResidentPasswordResponse, error)
	ResetContactPassword(ctx context.Context, req ResetContactPasswordRequest) (*ResetContactPasswordResponse, error)

	// 账户设置管理（统一 API）
	GetResidentAccountSettings(ctx context.Context, req GetResidentAccountSettingsRequest) (*GetResidentAccountSettingsResponse, error)
	UpdateResidentAccountSettings(ctx context.Context, req UpdateResidentAccountSettingsRequest) (*UpdateResidentAccountSettingsResponse, error)
}

// residentService 实现
type residentService struct {
	residentsRepo repository.ResidentsRepository
	db            *sql.DB // 用于复杂查询（JOIN、权限过滤）
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
func (s *residentService) getResourcePermission(ctx context.Context, roleCode, resourceType, permissionType string) (*PermissionCheckResult, error) {
	var assignedOnly, branchOnly bool
	err := s.db.QueryRowContext(ctx,
		`SELECT 
			COALESCE(assigned_only, FALSE) as assigned_only,
			COALESCE(branch_only, FALSE) as branch_only
		 FROM role_permissions
		 WHERE tenant_id = $1 
		   AND role_code = $2 
		   AND resource_type = $3 
		   AND permission_type = $4
		 LIMIT 1`,
		SystemTenantID, roleCode, resourceType, permissionType,
	).Scan(&assignedOnly, &branchOnly)

	if err == sql.ErrNoRows {
		// 记录不存在：返回最严格的权限（安全默认值）
		return &PermissionCheckResult{AssignedOnly: true, BranchOnly: true}, nil
	}
	if err != nil {
		return nil, err
	}

	return &PermissionCheckResult{AssignedOnly: assignedOnly, BranchOnly: branchOnly}, nil
}

// getUserBranchTag 查询用户的 branch_tag（Service 层内部方法）
func (s *residentService) getUserBranchTag(ctx context.Context, tenantID, userID string) (string, error) {
	var branchTag sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT branch_tag FROM users WHERE tenant_id = $1 AND user_id::text = $2`,
		tenantID, userID,
	).Scan(&branchTag)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // 用户不存在或没有 branch_tag
		}
		return "", err
	}
	if branchTag.Valid {
		return branchTag.String, nil
	}
	return "", nil
}

// ============================================
// Request/Response DTOs
// ============================================

// ListResidentsRequest 查询住户列表请求
type ListResidentsRequest struct {
	TenantID        string // 必填
	CurrentUserID   string // 当前用户ID（用于权限过滤）
	CurrentUserType string // 当前用户类型：'resident' | 'family' | 'staff'
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
	AssignedOnly  bool   // 是否仅限分配的资源
	BranchOnly    bool   // 是否仅限同一 Branch 的资源
	UserBranchTag string // 用户的 branch_tag（用于分支过滤）
}

// ListResidentsResponse 查询住户列表响应
type ListResidentsResponse struct {
	Items []*ResidentListItemDTO // 住户列表
	Total int                    // 总数量
}

// ResidentListItemDTO 住户列表项 DTO
type ResidentListItemDTO struct {
	ResidentID        string  `json:"resident_id"`
	TenantID          string  `json:"tenant_id"`
	ResidentAccount   *string `json:"resident_account,omitempty"`
	Nickname          string  `json:"nickname"`
	Status            string  `json:"status"`
	ServiceLevel      *string `json:"service_level,omitempty"`
	AdmissionDate     *int64  `json:"admission_date,omitempty"` // Unix timestamp
	DischargeDate     *int64  `json:"discharge_date,omitempty"` // Unix timestamp
	FamilyTag         *string `json:"family_tag,omitempty"`
	UnitID            *string `json:"unit_id,omitempty"`
	UnitName          *string `json:"unit_name,omitempty"`
	BranchTag         *string `json:"branch_tag,omitempty"`
	AreaTag           *string `json:"area_tag,omitempty"`
	UnitNumber        *string `json:"unit_number,omitempty"`
	IsMultiPersonRoom bool    `json:"is_multi_person_room"`
	RoomID            *string `json:"room_id,omitempty"`
	RoomName          *string `json:"room_name,omitempty"`
	BedID             *string `json:"bed_id,omitempty"`
	BedName           *string `json:"bed_name,omitempty"`
	IsAccessEnabled   bool    `json:"is_access_enabled"`
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

// GetResidentResponse 获取住户详情响应
type GetResidentResponse struct {
	Resident *ResidentDetailDTO    `json:"resident"`
	PHI      *ResidentPHIDTO       `json:"phi,omitempty"`
	Contacts []*ResidentContactDTO `json:"contacts,omitempty"`
}

// ResidentDetailDTO 住户详情 DTO
type ResidentDetailDTO struct {
	ResidentID        string  `json:"resident_id"`
	TenantID          string  `json:"tenant_id"`
	ResidentAccount   *string `json:"resident_account,omitempty"`
	Nickname          string  `json:"nickname"`
	Status            string  `json:"status"`
	ServiceLevel      *string `json:"service_level,omitempty"`
	AdmissionDate     *int64  `json:"admission_date,omitempty"` // Unix timestamp
	DischargeDate     *int64  `json:"discharge_date,omitempty"` // Unix timestamp
	FamilyTag         *string `json:"family_tag,omitempty"`
	UnitID            *string `json:"unit_id,omitempty"`
	UnitName          *string `json:"unit_name,omitempty"`
	BranchTag         *string `json:"branch_tag,omitempty"`
	AreaTag           *string `json:"area_tag,omitempty"`
	UnitNumber        *string `json:"unit_number,omitempty"`
	IsMultiPersonRoom bool    `json:"is_multi_person_room"`
	RoomID            *string `json:"room_id,omitempty"`
	RoomName          *string `json:"room_name,omitempty"`
	BedID             *string `json:"bed_id,omitempty"`
	BedName           *string `json:"bed_name,omitempty"`
	IsAccessEnabled   bool    `json:"is_access_enabled"`
	Note              *string `json:"note,omitempty"`
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

// CreateResidentRequest 创建住户请求
type CreateResidentRequest struct {
	TenantID        string // 必填
	CurrentUserID   string // 当前用户ID
	CurrentUserRole string // 当前用户角色

	// 权限检查结果（由 Handler 层传入）
	PermissionCheck *PermissionCheckResult // 权限检查结果

	// 必填字段
	ResidentAccount string // 住户账号（必填）
	Nickname        string // 昵称（必填）

	// 可选字段
	Password        string // 密码（默认 "ChangeMe123!"）
	Status          string // 状态（默认 "active"）
	ServiceLevel    string // 护理级别
	AdmissionDate   *int64 // 入院日期（Unix timestamp，默认当前日期）
	UnitID          string // 单元ID
	FamilyTag       string // 家庭标签
	IsAccessEnabled bool   // 是否允许查看状态
	Note            string // 备注

	// Hash 字段（前端计算，hex 字符串）
	PhoneHash string // phone_hash (hex)
	EmailHash string // email_hash (hex)

	// PHI 数据（可选）
	PHI *CreateResidentPHIRequest

	// 联系人数据（可选）
	Contacts []*CreateResidentContactRequest
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
type CreateResidentContactRequest struct {
	Slot             string // 'A', 'B', 'C', 'D', 'E'
	IsEnabled        bool
	Relationship     string
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

// CreateResidentResponse 创建住户响应
type CreateResidentResponse struct {
	ResidentID string // 创建的住户ID
}

// UpdateResidentRequest 更新住户请求
type UpdateResidentRequest struct {
	TenantID        string // 必填
	ResidentID      string // 必填
	CurrentUserID   string // 当前用户ID
	CurrentUserType string // 当前用户类型
	CurrentUserRole string // 当前用户角色（Service 层自己查询权限）

	// 可更新字段（使用指针表示可选）
	ResidentAccount *string // 住户账号（机构内部唯一标识）
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
	PHI *UpdateResidentPHIRequest

	// Caregivers 更新（可选）
	Caregivers *UpdateResidentCaregiversRequest
}

// UpdateResidentPHIRequest 更新住户 PHI 请求
type UpdateResidentPHIRequest struct {
	// 所有 PHI 字段（使用指针表示可选）
	FirstName                *string
	LastName                 *string
	Gender                   *string
	DateOfBirth              *int64
	ResidentPhone            *string
	ResidentEmail            *string
	PhoneHash                *string // phone_hash (hex string, 前端已计算)
	EmailHash                *string // email_hash (hex string, 前端已计算)
	WeightLb                 *float64
	HeightFt                 *float64
	HeightIn                 *float64
	MobilityLevel            *int
	TremorStatus             *string
	MobilityAid              *string
	ADLAssistance            *string
	CommStatus               *string
	HasHypertension          *bool
	HasHyperlipaemia         *bool
	HasHyperglycaemia        *bool
	HasStrokeHistory         *bool
	HasParalysis             *bool
	HasAlzheimer             *bool
	MedicalHistory           *string
	HISResidentName          *string
	HISResidentAdmissionDate *int64
	HISResidentDischargeDate *int64
	HomeAddressStreet        *string
	HomeAddressCity          *string
	HomeAddressState         *string
	HomeAddressPostalCode    *string
	PlusCode                 *string
}

// UpdateResidentCaregiversRequest 更新住户 Caregivers 请求
type UpdateResidentCaregiversRequest struct {
	UserList  []string // 用户ID列表
	GroupList []string // 标签ID列表
}

// UpdateResidentResponse 更新住户响应
type UpdateResidentResponse struct {
	Success bool
}

// UpdateResidentContactRequest 更新住户联系人请求
type UpdateResidentContactRequest struct {
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
	UserBranchTag   *string                // 当前用户 BranchTag (用于权限过滤)
	PermissionCheck *PermissionCheckResult // 权限检查结果
	NewPassword     string                 // 新密码（可选，默认生成）
}

// ResetResidentPasswordResponse 重置住户密码响应
type ResetResidentPasswordResponse struct {
	Success     bool
	NewPassword string // 生成的新密码
}

// ResetContactPasswordRequest 重置联系人密码请求
type ResetContactPasswordRequest struct {
	TenantID        string                 // 必填
	ContactID       string                 // 必填
	CurrentUserID   string                 // 当前用户ID
	CurrentUserType string                 // 当前用户类型
	CurrentUserRole string                 // 当前用户角色
	UserBranchTag   *string                // 当前用户 BranchTag (用于权限过滤)
	PermissionCheck *PermissionCheckResult // 权限检查结果
	NewPassword     string                 // 新密码（可选，默认生成）
}

// ResetContactPasswordResponse 重置联系人密码响应
type ResetContactPasswordResponse struct {
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

	// 2. 构建基础查询（JOIN units, rooms, beds）
	args := []any{req.TenantID}
	q := `SELECT r.resident_id::text, r.tenant_id::text, r.resident_account, r.nickname,
	             r.status, r.service_level, r.admission_date, r.discharge_date,
	             r.family_tag, r.unit_id::text, r.room_id::text, r.bed_id::text,
	             COALESCE(u.unit_name, '') as unit_name,
	             COALESCE(u.branch_name, '') as branch_tag,
	             COALESCE(u.area_name, '') as area_tag,
	             COALESCE(u.unit_number, '') as unit_number,
	             COALESCE(u.is_multi_person_room, false) as is_multi_person_room,
	             COALESCE(rm.room_name, '') as room_name,
	             COALESCE(b.bed_name, '') as bed_name,
	             r.can_view_status
	      FROM residents r
	      LEFT JOIN units u ON u.unit_id = r.unit_id
	      LEFT JOIN rooms rm ON rm.room_id = r.room_id
	      LEFT JOIN beds b ON b.bed_id = r.bed_id`

	// 3. 权限过滤
	if req.CurrentUserType == "resident" || req.CurrentUserType == "family" {
		// Resident/Family: 只能查看自己
		// 检查是否是 resident_contact 登录
		var residentIDForSelf sql.NullString
		if req.CurrentUserID != "" && s.db != nil {
			err := s.db.QueryRowContext(ctx,
				`SELECT resident_id::text FROM resident_contacts 
				 WHERE tenant_id = $1 AND contact_id::text = $2`,
				req.TenantID, req.CurrentUserID,
			).Scan(&residentIDForSelf)
			if err == nil && residentIDForSelf.Valid {
				// This is a resident_contact login
			} else {
				// This is a resident login
				residentIDForSelf = sql.NullString{String: req.CurrentUserID, Valid: true}
			}
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
			                        AND (rc.userList::text LIKE $%d OR rc.userList::text LIKE $%d)
			                  )`, len(args), len(args)+1)
			args = append(args, "%\""+req.CurrentUserID+"\"%")
		}

		// BranchOnly 过滤
		if req.PermissionCheck != nil && req.PermissionCheck.BranchOnly {
			userBranchTag := sql.NullString{String: req.PermissionCheck.UserBranchTag, Valid: req.PermissionCheck.UserBranchTag != ""}
			if !userBranchTag.Valid || userBranchTag.String == "" {
				// User branch_tag is NULL: can only view residents in units with branch_tag IS NULL
				q += ` AND (u.branch_name IS NULL OR u.branch_name = '-')`
			} else {
				// User branch_tag has value: can only view residents in matching branch
				args = append(args, userBranchTag.String)
				q += fmt.Sprintf(` AND u.branch_name = $%d`, len(args))
			}
		}
	}

	// 4. 搜索和过滤
	argIdx := len(args) + 1
	if req.Search != "" {
		args = append(args, "%"+req.Search+"%")
		q += fmt.Sprintf(` AND (r.nickname ILIKE $%d OR COALESCE(u.unit_name,'') ILIKE $%d)`, argIdx, argIdx)
		argIdx++
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

	// 5. 排序和分页
	q += ` ORDER BY r.nickname ASC`
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
		var familyTag, unitID, roomID, bedID sql.NullString
		var unitName, branchTag, areaTag, unitNumber sql.NullString
		var isMultiPersonRoom bool
		var roomName, bedName sql.NullString
		var canViewStatus bool

		if err := rows.Scan(
			&residentID, &tid, &residentAccount, &nickname,
			&status, &serviceLevel, &admissionDate, &dischargeDate,
			&familyTag, &unitID, &roomID, &bedID,
			&unitName, &branchTag, &areaTag, &unitNumber, &isMultiPersonRoom,
			&roomName, &bedName, &canViewStatus,
		); err != nil {
			s.logger.Error("ListResidents scan failed",
				zap.String("tenant_id", req.TenantID),
				zap.Error(err),
			)
			return nil, fmt.Errorf("failed to scan resident: %w", err)
		}

		item := &ResidentListItemDTO{
			ResidentID:        residentID.String,
			TenantID:          tid.String,
			Nickname:          nickname.String,
			Status:            status.String,
			IsMultiPersonRoom: isMultiPersonRoom,
			IsAccessEnabled:   canViewStatus,
		}

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
		if familyTag.Valid {
			item.FamilyTag = &familyTag.String
		}
		if unitID.Valid {
			item.UnitID = &unitID.String
		}
		if unitName.Valid && unitName.String != "" {
			item.UnitName = &unitName.String
		}
		if branchTag.Valid && branchTag.String != "" {
			item.BranchTag = &branchTag.String
		}
		if areaTag.Valid && areaTag.String != "" {
			item.AreaTag = &areaTag.String
		}
		if unitNumber.Valid && unitNumber.String != "" {
			item.UnitNumber = &unitNumber.String
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
	countQuery := strings.Replace(q, "SELECT r.resident_id::text, r.tenant_id::text, r.resident_account, r.nickname,\n\t             r.status, r.service_level, r.admission_date, r.discharge_date,\n\t             r.family_tag, r.unit_id::text, r.room_id::text, r.bed_id::text,\n\t             COALESCE(u.unit_name, '') as unit_name,\n\t             COALESCE(u.branch_name, '') as branch_tag,\n\t             COALESCE(u.area_name, '') as area_tag,\n\t             COALESCE(u.unit_number, '') as unit_number,\n\t             COALESCE(u.is_multi_person_room, false) as is_multi_person_room,\n\t             COALESCE(rm.room_name, '') as room_name,\n\t             COALESCE(b.bed_name, '') as bed_name,\n\t             r.can_view_status\n	      FROM residents r\n	      LEFT JOIN units u ON u.unit_id = r.unit_id\n	      LEFT JOIN rooms rm ON rm.room_id = r.room_id\n	      LEFT JOIN beds b ON b.bed_id = r.bed_id", "SELECT COUNT(*)\n	      FROM residents r\n	      LEFT JOIN units u ON u.unit_id = r.unit_id", 1)
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
	// 注意：对于 resident/family 用户，CurrentUserType 是 "resident"，CurrentUserRole 可能是 "Resident" 或 "Family"
	// 所以需要同时检查 CurrentUserType 和 CurrentUserRole
	isResidentOrFamily := req.CurrentUserType == "resident" || req.CurrentUserRole == "Resident" || req.CurrentUserRole == "Family"
	if isResidentOrFamily {
		// Resident/Family: 只能查看自己
		var foundResidentID sql.NullString
		if req.CurrentUserID != "" && s.db != nil {
			err := s.db.QueryRowContext(ctx,
				`SELECT resident_id::text FROM resident_contacts 
				 WHERE tenant_id = $1 AND contact_id::text = $2`,
				req.TenantID, req.CurrentUserID,
			).Scan(&foundResidentID)
			if err == nil && foundResidentID.Valid {
				// This is a resident_contact login - can only view linked resident
				if foundResidentID.String != actualResidentID {
					return nil, fmt.Errorf("access denied: can only view linked resident")
				}
			} else {
				// This is a resident login - can only view self
				if req.CurrentUserID != actualResidentID {
					return nil, fmt.Errorf("access denied: can only view own information")
				}
			}
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
						  AND (rc.userList::text LIKE $3 OR rc.userList::text LIKE $4)
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

			// BranchOnly 检查
			if req.PermissionCheck.BranchOnly && s.db != nil {
				var targetBranchTag sql.NullString
				err := s.db.QueryRowContext(ctx,
					`SELECT COALESCE(u.branch_name, '') as branch_tag
					 FROM residents r
					 LEFT JOIN units u ON u.unit_id = r.unit_id
					 WHERE r.tenant_id = $1 AND r.resident_id::text = $2`,
					req.TenantID, actualResidentID,
				).Scan(&targetBranchTag)
				if err != nil {
					if err == sql.ErrNoRows {
						return nil, fmt.Errorf("resident not found")
					}
					return nil, fmt.Errorf("failed to get resident info: %w", err)
				}

				userBranchTag := req.PermissionCheck.UserBranchTag
				if userBranchTag == "" {
					// User branch_tag is NULL: can only view residents in units with branch_tag IS NULL
					if targetBranchTag.Valid && targetBranchTag.String != "" {
						return nil, fmt.Errorf("permission denied: can only view residents in units with branch_tag IS NULL")
					}
				} else {
					// User branch_tag has value: can only view residents in matching branch
					if !targetBranchTag.Valid || targetBranchTag.String != userBranchTag {
						return nil, fmt.Errorf("permission denied: can only view residents in units with branch_tag = %s", userBranchTag)
					}
				}
			}
		}
	}

	// 3. 查询住户基本信息（JOIN units, rooms, beds）
	var residentID, tid, residentAccount, nickname, status, serviceLevel sql.NullString
	var admissionDate, dischargeDate sql.NullTime
	var familyTag, unitID, roomID, bedID sql.NullString
	var unitName, branchTag, areaTag, unitNumber sql.NullString
	var isMultiPersonRoom bool
	var roomName, bedName sql.NullString
	var note sql.NullString
	var canViewStatus bool

	err := s.db.QueryRowContext(ctx,
		`SELECT r.resident_id::text, r.tenant_id::text, r.resident_account, r.nickname,
		        r.status, r.service_level, r.admission_date, r.discharge_date,
		        r.family_tag, r.unit_id::text, r.room_id::text, r.bed_id::text,
		        COALESCE(u.unit_name, '') as unit_name,
		        COALESCE(u.branch_name, '') as branch_tag,
		        COALESCE(u.area_name, '') as area_tag,
		        COALESCE(u.unit_number, '') as unit_number,
		        COALESCE(u.is_multi_person_room, false) as is_multi_person_room,
		        COALESCE(rm.room_name, '') as room_name,
		        COALESCE(b.bed_name, '') as bed_name,
		        COALESCE(r.note, '') as note,
		        r.can_view_status
		 FROM residents r
		 LEFT JOIN units u ON u.unit_id = r.unit_id
		 LEFT JOIN rooms rm ON rm.room_id = r.room_id
		 LEFT JOIN beds b ON b.bed_id = r.bed_id
		 WHERE r.tenant_id = $1 AND r.resident_id::text = $2`,
		req.TenantID, actualResidentID,
	).Scan(
		&residentID, &tid, &residentAccount, &nickname,
		&status, &serviceLevel, &admissionDate, &dischargeDate,
		&familyTag, &unitID, &roomID, &bedID,
		&unitName, &branchTag, &areaTag, &unitNumber, &isMultiPersonRoom,
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
		ResidentID:        residentID.String,
		TenantID:          tid.String,
		Nickname:          nickname.String,
		Status:            status.String,
		IsMultiPersonRoom: isMultiPersonRoom,
		IsAccessEnabled:   canViewStatus,
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
	if familyTag.Valid {
		resident.FamilyTag = &familyTag.String
	}
	if unitID.Valid {
		resident.UnitID = &unitID.String
	}
	if unitName.Valid && unitName.String != "" {
		resident.UnitName = &unitName.String
	}
	if branchTag.Valid && branchTag.String != "" {
		resident.BranchTag = &branchTag.String
	}
	if areaTag.Valid && areaTag.String != "" {
		resident.AreaTag = &areaTag.String
	}
	if unitNumber.Valid && unitNumber.String != "" {
		resident.UnitNumber = &unitNumber.String
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

	return &GetResidentResponse{
		Resident: resident,
		PHI:      phi,
		Contacts: contacts,
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
		IsEmergencyContact: contact.IsEmergencyContact,
	}
	if contact.Relationship != "" {
		dto.Relationship = &contact.Relationship
	}
	if contact.ContactFirstName != "" {
		dto.ContactFirstName = &contact.ContactFirstName
	}
	if contact.ContactLastName != "" {
		dto.ContactLastName = &contact.ContactLastName
	}
	if contact.ContactPhone != "" {
		dto.ContactPhone = &contact.ContactPhone
	}
	if contact.ContactEmail != "" {
		dto.ContactEmail = &contact.ContactEmail
	}
	// ContactFamilyTag 字段在 domain.ResidentContact 中不存在，需要从数据库查询
	// 暂时跳过，如果需要可以从数据库单独查询
	return dto
}

// CreateResident 创建住户
func (s *residentService) CreateResident(ctx context.Context, req CreateResidentRequest) (*CreateResidentResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.ResidentAccount == "" {
		return nil, fmt.Errorf("resident_account is required (each institution has its own encoding pattern)")
	}
	if req.Nickname == "" {
		return nil, fmt.Errorf("nickname is required")
	}

	// 2. 业务规则验证
	// 2.1 resident_account 转换为小写
	residentAccount := strings.ToLower(strings.TrimSpace(req.ResidentAccount))

	// 2.2 计算 account_hash
	accountHashHex := HashAccount(residentAccount)
	accountHash, err := hex.DecodeString(accountHashHex)
	if err != nil || len(accountHash) == 0 {
		return nil, fmt.Errorf("failed to hash account")
	}

	// 2.3 生成 password_hash
	password := req.Password
	if password == "" {
		password = "ChangeMe123!" // 默认密码
	}
	passwordHashHex := HashPassword(password)
	passwordHash, err := hex.DecodeString(passwordHashHex)
	if err != nil || len(passwordHash) == 0 {
		return nil, fmt.Errorf("failed to hash password")
	}

	// 2.4 处理 phone_hash 和 email_hash（从请求中获取，前端已计算）
	var phoneHash, emailHash []byte
	if req.PhoneHash != "" {
		ph, err := hex.DecodeString(req.PhoneHash)
		if err == nil && len(ph) > 0 {
			phoneHash = ph
		}
	}
	if req.EmailHash != "" {
		eh, err := hex.DecodeString(req.EmailHash)
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
	if req.AdmissionDate != nil {
		admissionDate = *unixTimestampToTime(req.AdmissionDate)
	}

	// 2.7 处理 status（默认 "active"）
	status := req.Status
	if status == "" {
		status = "active"
	}

	// 2.8 unit_id 验证和权限检查
	if req.UnitID != "" {
		// 验证 unit 存在
		var unitExists bool
		err := s.db.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM units WHERE tenant_id = $1 AND unit_id::text = $2)`,
			req.TenantID, req.UnitID,
		).Scan(&unitExists)
		if err != nil {
			return nil, fmt.Errorf("failed to check unit existence: %w", err)
		}
		if !unitExists {
			return nil, fmt.Errorf("unit not found")
		}

		// BranchOnly 权限检查
		if req.PermissionCheck != nil && req.PermissionCheck.BranchOnly {
			var unitBranchTag sql.NullString
			err := s.db.QueryRowContext(ctx,
				`SELECT branch_name FROM units WHERE tenant_id = $1 AND unit_id::text = $2`,
				req.TenantID, req.UnitID,
			).Scan(&unitBranchTag)
			if err != nil {
				return nil, fmt.Errorf("failed to check unit branch: %w", err)
			}

			userBranchTag := req.PermissionCheck.UserBranchTag
			if userBranchTag == "" {
				// User branch_tag is NULL: can only create residents in units with branch_tag IS NULL
				if unitBranchTag.Valid && unitBranchTag.String != "" {
					return nil, fmt.Errorf("permission denied: can only create residents in units with branch_tag IS NULL")
				}
			} else {
				// User branch_tag has value: can only create residents in units with matching branch_tag
				if !unitBranchTag.Valid || unitBranchTag.String != userBranchTag {
					return nil, fmt.Errorf("permission denied: can only create residents in units with branch_tag = %s", userBranchTag)
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
		Nickname:            strings.TrimSpace(req.Nickname),
		Status:              status,
		Role:                "Resident",
		AdmissionDate:       &admissionDate,
		CanViewStatus:       req.IsAccessEnabled,
		UnitID:              req.UnitID,
		FamilyTag:           req.FamilyTag,
		Note:                req.Note,
		PhoneHash:           phoneHash,
		EmailHash:           emailHash,
		PasswordHash:        passwordHash,
	}
	if req.ServiceLevel != "" {
		resident.ServiceLevel = req.ServiceLevel
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
	if req.PHI != nil && req.PHI.FirstName != "" {
		phi := &domain.ResidentPHI{
			FirstName: req.PHI.FirstName,
			LastName:  req.PHI.LastName,
			Gender:    req.PHI.Gender,
		}
		if req.PHI.DateOfBirth != nil {
			phi.DateOfBirth = unixTimestampToTime(req.PHI.DateOfBirth)
		}
		// 只在 save_phone/save_email 为 true 时保存明文
		if req.PHI.SavePhone && req.PHI.ResidentPhone != "" {
			phi.ResidentPhone = req.PHI.ResidentPhone
		}
		if req.PHI.SaveEmail && req.PHI.ResidentEmail != "" {
			phi.ResidentEmail = req.PHI.ResidentEmail
		}
		// 其他 PHI 字段
		phi.WeightLb = req.PHI.WeightLb
		phi.HeightFt = req.PHI.HeightFt
		phi.HeightIn = req.PHI.HeightIn
		phi.MobilityLevel = req.PHI.MobilityLevel
		phi.TremorStatus = req.PHI.TremorStatus
		phi.MobilityAid = req.PHI.MobilityAid
		phi.ADLAssistance = req.PHI.ADLAssistance
		phi.CommStatus = req.PHI.CommStatus
		if req.PHI.HasHypertension != nil {
			phi.HasHypertension = *req.PHI.HasHypertension
		}
		if req.PHI.HasHyperlipaemia != nil {
			phi.HasHyperlipaemia = *req.PHI.HasHyperlipaemia
		}
		if req.PHI.HasHyperglycaemia != nil {
			phi.HasHyperglycaemia = *req.PHI.HasHyperglycaemia
		}
		if req.PHI.HasStrokeHistory != nil {
			phi.HasStrokeHistory = *req.PHI.HasStrokeHistory
		}
		if req.PHI.HasParalysis != nil {
			phi.HasParalysis = *req.PHI.HasParalysis
		}
		if req.PHI.HasAlzheimer != nil {
			phi.HasAlzheimer = *req.PHI.HasAlzheimer
		}
		phi.MedicalHistory = req.PHI.MedicalHistory
		phi.HomeAddressStreet = req.PHI.HomeAddressStreet
		phi.HomeAddressCity = req.PHI.HomeAddressCity
		phi.HomeAddressState = req.PHI.HomeAddressState
		phi.HomeAddressPostalCode = req.PHI.HomeAddressPostalCode
		phi.PlusCode = req.PHI.PlusCode

		if err := s.residentsRepo.UpsertResidentPHI(ctx, req.TenantID, residentID, phi); err != nil {
			s.logger.Warn("Failed to create PHI record",
				zap.String("tenant_id", req.TenantID),
				zap.String("resident_id", residentID),
				zap.Error(err),
			)
			// 不失败整个操作，只记录警告
		}
	}

	// 5. 创建联系人记录（如果提供了 contacts）
	if len(req.Contacts) > 0 {
		for _, contactReq := range req.Contacts {
			// Hash 唯一性检查
			var contactPhoneHash, contactEmailHash []byte
			if contactReq.PhoneHash != "" {
				ph, err := hex.DecodeString(contactReq.PhoneHash)
				if err == nil && len(ph) > 0 {
					contactPhoneHash = ph
				}
			}
			if contactReq.EmailHash != "" {
				eh, err := hex.DecodeString(contactReq.EmailHash)
				if err == nil && len(eh) > 0 {
					contactEmailHash = eh
				}
			}

			// 如果 phone_hash 或 email_hash 已存在，复用已存在的联系人信息
			var existingContact *domain.ResidentContact
			if len(contactPhoneHash) > 0 || len(contactEmailHash) > 0 {
				existingContact = s.findExistingContactByHash(ctx, req.TenantID, contactPhoneHash, contactEmailHash)
			}

			// 生成默认密码 hash（联系人密码独立于 account）
			contactPassword := "ChangeMe123!"
			contactPasswordHashHex := HashPassword(contactPassword)
			contactPasswordHash, _ := hex.DecodeString(contactPasswordHashHex)

			contact := &domain.ResidentContact{
				Slot:               contactReq.Slot,
				IsEnabled:          contactReq.IsEnabled,
				Relationship:       contactReq.Relationship,
				ContactFirstName:   contactReq.ContactFirstName,
				ContactLastName:    contactReq.ContactLastName,
				ContactPhone:       contactReq.ContactPhone,
				ContactEmail:       contactReq.ContactEmail,
				ReceiveSMS:         contactReq.ReceiveSMS,
				ReceiveEmail:       contactReq.ReceiveEmail,
				PhoneHash:          contactPhoneHash,
				EmailHash:          contactEmailHash,
				PasswordHash:       contactPasswordHash,
				Role:               "Family",
				IsEmergencyContact: false,
			}
			if contact.Slot == "" {
				contact.Slot = "A" // 默认 slot
			}

			// 如果找到已存在的联系人，复用其信息（优先使用请求中的值，如果请求中为空则使用已存在的值）
			if existingContact != nil {
				if contact.ContactFirstName == "" && existingContact.ContactFirstName != "" {
					contact.ContactFirstName = existingContact.ContactFirstName
				}
				if contact.ContactLastName == "" && existingContact.ContactLastName != "" {
					contact.ContactLastName = existingContact.ContactLastName
				}
				if contact.ContactPhone == "" && existingContact.ContactPhone != "" {
					contact.ContactPhone = existingContact.ContactPhone
				}
				if contact.ContactEmail == "" && existingContact.ContactEmail != "" {
					contact.ContactEmail = existingContact.ContactEmail
				}
				if contact.Relationship == "" && existingContact.Relationship != "" {
					contact.Relationship = existingContact.Relationship
				}
				// 复用 password_hash（如果已存在联系人已有密码）
				if len(existingContact.PasswordHash) > 0 && len(contactPasswordHash) == 0 {
					contact.PasswordHash = existingContact.PasswordHash
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
// 如果找到，返回联系人信息以便复用
func (s *residentService) findExistingContactByHashExcluding(ctx context.Context, tenantID string, phoneHash, emailHash []byte, excludeContactID string) *domain.ResidentContact {
	var query string
	var args []any

	// 优先使用 phone_hash，如果没有则使用 email_hash
	if len(phoneHash) > 0 {
		if excludeContactID != "" {
			query = `
				SELECT 
					contact_id::text,
					tenant_id::text,
					resident_id::text,
					slot,
					is_enabled,
					relationship,
					role,
					is_emergency_contact,
					COALESCE(alert_time_window, '{}'::jsonb)::text as alert_time_window,
					contact_first_name,
					contact_last_name,
					contact_phone,
					contact_email,
					receive_sms,
					receive_email,
					phone_hash,
					email_hash,
					password_hash
				FROM resident_contacts
				WHERE tenant_id = $1 AND phone_hash = $2 AND contact_id::text != $3
				LIMIT 1
			`
			args = []any{tenantID, phoneHash, excludeContactID}
		} else {
			query = `
				SELECT 
					contact_id::text,
					tenant_id::text,
					resident_id::text,
					slot,
					is_enabled,
					relationship,
					role,
					is_emergency_contact,
					COALESCE(alert_time_window, '{}'::jsonb)::text as alert_time_window,
					contact_first_name,
					contact_last_name,
					contact_phone,
					contact_email,
					receive_sms,
					receive_email,
					phone_hash,
					email_hash,
					password_hash
				FROM resident_contacts
				WHERE tenant_id = $1 AND phone_hash = $2
				LIMIT 1
			`
			args = []any{tenantID, phoneHash}
		}
	} else if len(emailHash) > 0 {
		if excludeContactID != "" {
			query = `
				SELECT 
					contact_id::text,
					tenant_id::text,
					resident_id::text,
					slot,
					is_enabled,
					relationship,
					role,
					is_emergency_contact,
					COALESCE(alert_time_window, '{}'::jsonb)::text as alert_time_window,
					contact_first_name,
					contact_last_name,
					contact_phone,
					contact_email,
					receive_sms,
					receive_email,
					phone_hash,
					email_hash,
					password_hash
				FROM resident_contacts
				WHERE tenant_id = $1 AND email_hash = $2 AND contact_id::text != $3
				LIMIT 1
			`
			args = []any{tenantID, emailHash, excludeContactID}
		} else {
			query = `
				SELECT 
					contact_id::text,
					tenant_id::text,
					resident_id::text,
					slot,
					is_enabled,
					relationship,
					role,
					is_emergency_contact,
					COALESCE(alert_time_window, '{}'::jsonb)::text as alert_time_window,
					contact_first_name,
					contact_last_name,
					contact_phone,
					contact_email,
					receive_sms,
					receive_email,
					phone_hash,
					email_hash,
					password_hash
				FROM resident_contacts
				WHERE tenant_id = $1 AND email_hash = $2
				LIMIT 1
			`
			args = []any{tenantID, emailHash}
		}
	} else {
		return nil
	}

	var contact domain.ResidentContact
	var relationship, contactFirstName, contactLastName, contactPhone, contactEmail sql.NullString
	var alertTimeWindow sql.NullString
	var phoneHashDB, emailHashDB, passwordHashDB sql.Null[[]byte]

	err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&contact.ContactID,
		&contact.TenantID,
		&contact.ResidentID,
		&contact.Slot,
		&contact.IsEnabled,
		&relationship,
		&contact.Role,
		&contact.IsEmergencyContact,
		&alertTimeWindow,
		&contactFirstName,
		&contactLastName,
		&contactPhone,
		&contactEmail,
		&contact.ReceiveSMS,
		&contact.ReceiveEmail,
		&phoneHashDB,
		&emailHashDB,
		&passwordHashDB,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		// 查询错误，返回 nil（不中断流程）
		return nil
	}

	// 处理可空字段
	if relationship.Valid {
		contact.Relationship = relationship.String
	}
	if contactFirstName.Valid {
		contact.ContactFirstName = contactFirstName.String
	}
	if contactLastName.Valid {
		contact.ContactLastName = contactLastName.String
	}
	if contactPhone.Valid {
		contact.ContactPhone = contactPhone.String
	}
	if contactEmail.Valid {
		contact.ContactEmail = contactEmail.String
	}
	if alertTimeWindow.Valid && alertTimeWindow.String != "" {
		contact.AlertTimeWindow = json.RawMessage(alertTimeWindow.String)
	}
	if phoneHashDB.Valid {
		contact.PhoneHash = phoneHashDB.V
	}
	if emailHashDB.Valid {
		contact.EmailHash = emailHashDB.V
	}
	if passwordHashDB.Valid {
		contact.PasswordHash = passwordHashDB.V
	}

	return &contact
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
		// 查询 1：用户的 branch_name 和 role（同时检查是否是 Manager）
		var userBranchName sql.NullString
		var userRole sql.NullString
		err := s.db.QueryRowContext(ctx,
			`SELECT branch_name, role FROM users WHERE tenant_id = $1 AND user_id::text = $2`,
			req.TenantID, req.CurrentUserID,
		).Scan(&userBranchName, &userRole)
		if err != nil {
			return nil, fmt.Errorf("failed to get user info: %w", err)
		}
		if !userRole.Valid || userRole.String != "Manager" {
			return nil, fmt.Errorf("access denied: user role is not Manager")
		}

		// 查询 2：目标 resident 的 branch_name
		var targetBranchName sql.NullString
		err = s.db.QueryRowContext(ctx,
			`SELECT COALESCE(u.branch_name, '') as branch_name
			 FROM residents r
			 LEFT JOIN units u ON u.unit_id = r.unit_id
			 WHERE r.tenant_id = $1 AND r.resident_id::text = $2`,
			req.TenantID, req.ResidentID,
		).Scan(&targetBranchName)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("resident not found")
			}
			return nil, fmt.Errorf("failed to get resident info: %w", err)
		}

		// 如果两者的 branchName 均为 ""，视为相同
		userBranch := ""
		if userBranchName.Valid && userBranchName.String != "" {
			userBranch = userBranchName.String
		}
		targetBranch := ""
		if targetBranchName.Valid && targetBranchName.String != "" {
			targetBranch = targetBranchName.String
		}

		if userBranch != targetBranch {
			return nil, fmt.Errorf("permission denied: can only update residents in same branch")
		}
		// 允许更新
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
					-- 直接分配：userList 中包含 user_id
					rc.userList::text LIKE $3
					OR rc.userList::text LIKE $4
					-- 通过 user_tag 分配：groupList 中的 tag_id 对应的 tag_name 在 users.tags 中
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
									SELECT jsonb_array_elements_text(rc.groupList)::text
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

	// 2. 获取现有住户信息
	existingResident, err := s.residentsRepo.GetResident(ctx, req.TenantID, req.ResidentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get resident: %w", err)
	}

	// 3. 构建更新字段
	updates := &domain.Resident{
		ResidentID: req.ResidentID,
	}

	// 基本字段更新
	if req.ResidentAccount != nil {
		// 更新 resident_account 需要同时更新 resident_account_hash
		updates.ResidentAccount = strings.ToLower(*req.ResidentAccount)
		// HashAccount 返回 hex 字符串，需要转换为 []byte
		hashHex := HashAccount(updates.ResidentAccount)
		hashBytes, err := hex.DecodeString(hashHex)
		if err != nil {
			return nil, fmt.Errorf("failed to decode resident_account_hash: %w", err)
		}
		updates.ResidentAccountHash = hashBytes
	} else {
		updates.ResidentAccount = existingResident.ResidentAccount
		updates.ResidentAccountHash = existingResident.ResidentAccountHash
	}

	if req.Nickname != nil {
		updates.Nickname = *req.Nickname
	} else {
		updates.Nickname = existingResident.Nickname
	}

	if req.Status != nil {
		updates.Status = *req.Status
	} else {
		updates.Status = existingResident.Status
	}

	if req.ServiceLevel != nil {
		updates.ServiceLevel = *req.ServiceLevel
	} else {
		updates.ServiceLevel = existingResident.ServiceLevel
	}

	if req.AdmissionDate != nil {
		updates.AdmissionDate = unixTimestampToTime(req.AdmissionDate)
	} else {
		updates.AdmissionDate = existingResident.AdmissionDate
	}

	if req.DischargeDate != nil {
		// discharge_date 验证：仅在 status='discharged' 或 'transferred' 时可以有值
		if updates.Status != "discharged" && updates.Status != "transferred" {
			return nil, fmt.Errorf("discharge_date can only be set when status is 'discharged' or 'transferred'")
		}
		updates.DischargeDate = unixTimestampToTime(req.DischargeDate)
	} else {
		updates.DischargeDate = existingResident.DischargeDate
	}

	if req.UnitID != nil {
		updates.UnitID = *req.UnitID
	} else {
		updates.UnitID = existingResident.UnitID
	}

	// RoomID 和 BedID 不在 UpdateResidentRequest 中，保持现有值
	updates.RoomID = existingResident.RoomID
	updates.BedID = existingResident.BedID

	if req.FamilyTag != nil {
		updates.FamilyTag = *req.FamilyTag
	} else {
		updates.FamilyTag = existingResident.FamilyTag
	}

	if req.IsAccessEnabled != nil {
		updates.CanViewStatus = *req.IsAccessEnabled
	} else {
		updates.CanViewStatus = existingResident.CanViewStatus
	}

	if req.Note != nil {
		updates.Note = *req.Note
	} else {
		updates.Note = existingResident.Note
	}

	// 4. 更新 Resident 记录
	if err := s.residentsRepo.UpdateResident(ctx, req.TenantID, req.ResidentID, updates); err != nil {
		s.logger.Error("UpdateResident failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("resident_id", req.ResidentID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to update resident: %w", err)
	}

	// 5. 更新 residents 表的 phone_hash/email_hash（如果提供了）
	if req.PHI != nil && (req.PHI.PhoneHash != nil || req.PHI.EmailHash != nil) {
		// 获取现有 resident 数据
		existingResident, err := s.residentsRepo.GetResident(ctx, req.TenantID, req.ResidentID)
		if err != nil {
			s.logger.Warn("Failed to get existing resident for hash update",
				zap.String("tenant_id", req.TenantID),
				zap.String("resident_id", req.ResidentID),
				zap.Error(err),
			)
		} else if existingResident != nil {
			// 准备更新 phone_hash/email_hash
			hashUpdated := false
			var phoneHashBytes, emailHashBytes []byte
			var phoneHashToSet, emailHashToSet *[]byte

			// 处理 phone_hash
			if req.PHI.PhoneHash != nil {
				if *req.PHI.PhoneHash == "" {
					// 空字符串表示删除（设置为 NULL）
					phoneHashToSet = nil
					hashUpdated = true
				} else {
					// 解码 hex 字符串
					decoded, err := hex.DecodeString(*req.PHI.PhoneHash)
					if err != nil {
						s.logger.Warn("Failed to decode phone_hash",
							zap.String("tenant_id", req.TenantID),
							zap.String("resident_id", req.ResidentID),
							zap.Error(err),
						)
					} else if len(decoded) > 0 {
						phoneHashBytes = decoded
						phoneHashToSet = &phoneHashBytes
						hashUpdated = true
					}
				}
			}

			// 处理 email_hash
			if req.PHI.EmailHash != nil {
				if *req.PHI.EmailHash == "" {
					// 空字符串表示删除（设置为 NULL）
					emailHashToSet = nil
					hashUpdated = true
				} else {
					// 解码 hex 字符串
					decoded, err := hex.DecodeString(*req.PHI.EmailHash)
					if err != nil {
						s.logger.Warn("Failed to decode email_hash",
							zap.String("tenant_id", req.TenantID),
							zap.String("resident_id", req.ResidentID),
							zap.Error(err),
						)
					} else if len(decoded) > 0 {
						emailHashBytes = decoded
						emailHashToSet = &emailHashBytes
						hashUpdated = true
					}
				}
			}

			// 检查唯一性（排除当前 resident）
			if phoneHashToSet != nil && len(*phoneHashToSet) > 0 {
				var count int
				err := s.db.QueryRowContext(ctx,
					`SELECT COUNT(*) FROM residents WHERE tenant_id = $1 AND phone_hash = $2 AND resident_id::text != $3`,
					req.TenantID, *phoneHashToSet, req.ResidentID,
				).Scan(&count)
				if err != nil {
					s.logger.Warn("Failed to check phone_hash uniqueness",
						zap.String("tenant_id", req.TenantID),
						zap.String("resident_id", req.ResidentID),
						zap.Error(err),
					)
				} else if count > 0 {
					return nil, fmt.Errorf("phone already exists in this organization")
				}
			}

			if emailHashToSet != nil && len(*emailHashToSet) > 0 {
				var count int
				err := s.db.QueryRowContext(ctx,
					`SELECT COUNT(*) FROM residents WHERE tenant_id = $1 AND email_hash = $2 AND resident_id::text != $3`,
					req.TenantID, *emailHashToSet, req.ResidentID,
				).Scan(&count)
				if err != nil {
					s.logger.Warn("Failed to check email_hash uniqueness",
						zap.String("tenant_id", req.TenantID),
						zap.String("resident_id", req.ResidentID),
						zap.Error(err),
					)
				} else if count > 0 {
					return nil, fmt.Errorf("email already exists in this organization")
				}
			}

			// 更新 residents 表的 phone_hash/email_hash
			if hashUpdated {
				hashUpdate := &domain.Resident{
					ResidentID: existingResident.ResidentID,
					TenantID:   existingResident.TenantID,
				}
				// 设置 phone_hash（nil 表示不更新，空 slice 表示设置为 NULL，非空 slice 表示设置值）
				if phoneHashToSet != nil {
					if len(*phoneHashToSet) > 0 {
						hashUpdate.PhoneHash = *phoneHashToSet
					} else {
						hashUpdate.PhoneHash = []byte{} // 空 slice 表示设置为 NULL
					}
				}
				// 设置 email_hash（nil 表示不更新，空 slice 表示设置为 NULL，非空 slice 表示设置值）
				if emailHashToSet != nil {
					if len(*emailHashToSet) > 0 {
						hashUpdate.EmailHash = *emailHashToSet
					} else {
						hashUpdate.EmailHash = []byte{} // 空 slice 表示设置为 NULL
					}
				}

				// 只更新 phone_hash/email_hash
				if err := s.residentsRepo.UpdateResident(ctx, req.TenantID, req.ResidentID, hashUpdate); err != nil {
					s.logger.Warn("Failed to update phone_hash/email_hash",
						zap.String("tenant_id", req.TenantID),
						zap.String("resident_id", req.ResidentID),
						zap.Error(err),
					)
					// 不失败整个操作，只记录警告
				}
			}
		}
	}

	// 6. 更新 PHI 数据（如果提供了）
	if req.PHI != nil {
		phi := &domain.ResidentPHI{}
		// 从现有 PHI 获取，然后更新提供的字段
		existingPHI, _ := s.residentsRepo.GetResidentPHI(ctx, req.TenantID, req.ResidentID)
		if existingPHI != nil {
			phi = existingPHI
		}

		// 更新提供的字段
		if req.PHI.FirstName != nil {
			phi.FirstName = *req.PHI.FirstName
		}
		if req.PHI.LastName != nil {
			phi.LastName = *req.PHI.LastName
		}
		if req.PHI.Gender != nil {
			phi.Gender = *req.PHI.Gender
		}
		if req.PHI.DateOfBirth != nil {
			phi.DateOfBirth = unixTimestampToTime(req.PHI.DateOfBirth)
		}
		// 处理 resident_phone（空字符串表示删除，设置为 NULL）
		if req.PHI.ResidentPhone != nil {
			if *req.PHI.ResidentPhone == "" {
				phi.ResidentPhone = "" // 空字符串，repository 会设置为 NULL
			} else {
				phi.ResidentPhone = *req.PHI.ResidentPhone
			}
		}
		// 处理 resident_email（空字符串表示删除，设置为 NULL）
		if req.PHI.ResidentEmail != nil {
			if *req.PHI.ResidentEmail == "" {
				phi.ResidentEmail = "" // 空字符串，repository 会设置为 NULL
			} else {
				phi.ResidentEmail = *req.PHI.ResidentEmail
			}
		}
		if req.PHI.WeightLb != nil {
			phi.WeightLb = req.PHI.WeightLb
		}
		if req.PHI.HeightFt != nil {
			phi.HeightFt = req.PHI.HeightFt
		}
		if req.PHI.HeightIn != nil {
			phi.HeightIn = req.PHI.HeightIn
		}
		if req.PHI.MobilityLevel != nil {
			phi.MobilityLevel = req.PHI.MobilityLevel
		}
		if req.PHI.TremorStatus != nil {
			phi.TremorStatus = *req.PHI.TremorStatus
		}
		if req.PHI.MobilityAid != nil {
			phi.MobilityAid = *req.PHI.MobilityAid
		}
		if req.PHI.ADLAssistance != nil {
			phi.ADLAssistance = *req.PHI.ADLAssistance
		}
		if req.PHI.CommStatus != nil {
			phi.CommStatus = *req.PHI.CommStatus
		}
		if req.PHI.HasHypertension != nil {
			phi.HasHypertension = *req.PHI.HasHypertension
		}
		if req.PHI.HasHyperlipaemia != nil {
			phi.HasHyperlipaemia = *req.PHI.HasHyperlipaemia
		}
		if req.PHI.HasHyperglycaemia != nil {
			phi.HasHyperglycaemia = *req.PHI.HasHyperglycaemia
		}
		if req.PHI.HasStrokeHistory != nil {
			phi.HasStrokeHistory = *req.PHI.HasStrokeHistory
		}
		if req.PHI.HasParalysis != nil {
			phi.HasParalysis = *req.PHI.HasParalysis
		}
		if req.PHI.HasAlzheimer != nil {
			phi.HasAlzheimer = *req.PHI.HasAlzheimer
		}
		if req.PHI.MedicalHistory != nil {
			phi.MedicalHistory = *req.PHI.MedicalHistory
		}
		if req.PHI.HomeAddressStreet != nil {
			phi.HomeAddressStreet = *req.PHI.HomeAddressStreet
		}
		if req.PHI.HomeAddressCity != nil {
			phi.HomeAddressCity = *req.PHI.HomeAddressCity
		}
		if req.PHI.HomeAddressState != nil {
			phi.HomeAddressState = *req.PHI.HomeAddressState
		}
		if req.PHI.HomeAddressPostalCode != nil {
			phi.HomeAddressPostalCode = *req.PHI.HomeAddressPostalCode
		}
		if req.PHI.PlusCode != nil {
			phi.PlusCode = *req.PHI.PlusCode
		}

		if err := s.residentsRepo.UpsertResidentPHI(ctx, req.TenantID, req.ResidentID, phi); err != nil {
			s.logger.Warn("Failed to update PHI",
				zap.String("tenant_id", req.TenantID),
				zap.String("resident_id", req.ResidentID),
				zap.Error(err),
			)
			// 不失败整个操作，只记录警告
		}
	}

	// 6. 更新 Caregivers 数据（如果提供了）
	if req.Caregivers != nil {
		// 将 []string 转换为 json.RawMessage
		var userListJSON, groupListJSON json.RawMessage
		if len(req.Caregivers.UserList) > 0 {
			userListJSON, _ = json.Marshal(req.Caregivers.UserList)
		}
		if len(req.Caregivers.GroupList) > 0 {
			groupListJSON, _ = json.Marshal(req.Caregivers.GroupList)
		}

		caregiver := &domain.ResidentCaregiver{
			UserList:  userListJSON,
			GroupList: groupListJSON,
		}
		if err := s.residentsRepo.UpsertResidentCaregiver(ctx, req.TenantID, req.ResidentID, caregiver); err != nil {
			s.logger.Warn("Failed to update caregivers",
				zap.String("tenant_id", req.TenantID),
				zap.String("resident_id", req.ResidentID),
				zap.Error(err),
			)
			// 不失败整个操作，只记录警告
		}
	}

	return &UpdateResidentResponse{
		Success: true,
	}, nil
}

// DeleteResident 删除住户（软删除）
func (s *residentService) DeleteResident(ctx context.Context, req DeleteResidentRequest) (*DeleteResidentResponse, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.ResidentID == "" {
		return nil, fmt.Errorf("resident_id is required")
	}

	// 1. 权限检查
	if req.CurrentUserType == "resident" || req.CurrentUserType == "family" {
		// Resident/Family 不能删除
		return nil, fmt.Errorf("access denied: residents and family members cannot delete residents")
	}

	// Staff: 权限检查（AssignedOnly, BranchOnly）
	if req.PermissionCheck != nil {
		if req.PermissionCheck.AssignedOnly && req.CurrentUserID != "" {
			var isAssigned bool
			err := s.db.QueryRowContext(ctx,
				`SELECT EXISTS(
					SELECT 1 FROM resident_caregivers rc
					WHERE rc.tenant_id = $1
					  AND rc.resident_id::text = $2
					  AND (rc.userList::text LIKE $3 OR rc.userList::text LIKE $4)
				)`,
				req.TenantID, req.ResidentID, req.CurrentUserID, "%\""+req.CurrentUserID+"\"%",
			).Scan(&isAssigned)
			if err != nil {
				return nil, fmt.Errorf("failed to check assignment: %w", err)
			}
			if !isAssigned {
				return nil, fmt.Errorf("permission denied: can only delete assigned residents")
			}
		}

		if req.PermissionCheck.BranchOnly {
			var targetBranchTag sql.NullString
			err := s.db.QueryRowContext(ctx,
				`SELECT COALESCE(u.branch_name, '') as branch_tag
				 FROM residents r
				 LEFT JOIN units u ON u.unit_id = r.unit_id
				 WHERE r.tenant_id = $1 AND r.resident_id::text = $2`,
				req.TenantID, req.ResidentID,
			).Scan(&targetBranchTag)
			if err != nil {
				if err == sql.ErrNoRows {
					return nil, fmt.Errorf("resident not found")
				}
				return nil, fmt.Errorf("failed to get resident info: %w", err)
			}

			userBranchTag := req.PermissionCheck.UserBranchTag
			if userBranchTag == "" {
				if targetBranchTag.Valid && targetBranchTag.String != "" {
					return nil, fmt.Errorf("permission denied: can only delete residents in units with branch_tag IS NULL")
				}
			} else {
				if !targetBranchTag.Valid || targetBranchTag.String != userBranchTag {
					return nil, fmt.Errorf("permission denied: can only delete residents in units with branch_tag = %s", userBranchTag)
				}
			}
		}
	}

	// 2. 软删除：将 status 设置为 'discharged'
	dischargedStatus := "discharged"
	updates := &domain.Resident{
		Status: dischargedStatus,
	}

	if err := s.residentsRepo.UpdateResident(ctx, req.TenantID, req.ResidentID, updates); err != nil {
		s.logger.Error("DeleteResident failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("resident_id", req.ResidentID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to delete resident: %w", err)
	}

	return &DeleteResidentResponse{
		Success: true,
	}, nil
}

// ResetResidentPassword 重置住户密码
func (s *residentService) ResetResidentPassword(ctx context.Context, req ResetResidentPasswordRequest) (*ResetResidentPasswordResponse, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.ResidentID == "" {
		return nil, fmt.Errorf("resident_id is required")
	}

	// 1. 权限检查
	if req.CurrentUserType == "resident" || req.CurrentUserType == "family" {
		// Resident/Family: 只能重置自己的密码
		if req.CurrentUserID != req.ResidentID {
			return nil, fmt.Errorf("access denied: can only reset own password")
		}
	} else {
		// Staff: 权限检查（AssignedOnly, BranchOnly）
		if req.PermissionCheck != nil {
			if req.PermissionCheck.AssignedOnly && req.CurrentUserID != "" {
				var isAssigned bool
				err := s.db.QueryRowContext(ctx,
					`SELECT EXISTS(
						SELECT 1 FROM resident_caregivers rc
						WHERE rc.tenant_id = $1
						  AND rc.resident_id::text = $2
						  AND (rc.userList::text LIKE $3 OR rc.userList::text LIKE $4)
					)`,
					req.TenantID, req.ResidentID, req.CurrentUserID, "%\""+req.CurrentUserID+"\"%",
				).Scan(&isAssigned)
				if err != nil {
					return nil, fmt.Errorf("failed to check assignment: %w", err)
				}
				if !isAssigned {
					return nil, fmt.Errorf("permission denied: can only reset password for assigned residents")
				}
			}

			if req.PermissionCheck.BranchOnly {
				var targetBranchTag sql.NullString
				err := s.db.QueryRowContext(ctx,
					`SELECT COALESCE(u.branch_name, '') as branch_tag
					 FROM residents r
					 LEFT JOIN units u ON u.unit_id = r.unit_id
					 WHERE r.tenant_id = $1 AND r.resident_id::text = $2`,
					req.TenantID, req.ResidentID,
				).Scan(&targetBranchTag)
				if err != nil {
					if err == sql.ErrNoRows {
						return nil, fmt.Errorf("resident not found")
					}
					return nil, fmt.Errorf("failed to get resident info: %w", err)
				}

				userBranchTag := req.PermissionCheck.UserBranchTag
				if userBranchTag == "" {
					if targetBranchTag.Valid && targetBranchTag.String != "" {
						return nil, fmt.Errorf("permission denied: can only reset password for residents in units with branch_tag IS NULL")
					}
				} else {
					if !targetBranchTag.Valid || targetBranchTag.String != userBranchTag {
						return nil, fmt.Errorf("permission denied: can only reset password for residents in units with branch_tag = %s", userBranchTag)
					}
				}
			}
		}
	}

	// 2. 密码哈希（前端已 hash，这里直接解码 hex 字符串）
	// 前端发送的是 SHA256(password) 的 hex 字符串，直接解码为 byte slice
	if req.NewPassword == "" {
		return nil, fmt.Errorf("password hash is required")
	}
	passwordHash, err := hex.DecodeString(req.NewPassword)
	if err != nil || len(passwordHash) == 0 {
		return nil, fmt.Errorf("failed to decode password hash: %w", err)
	}

	// 4. 更新 residents 表
	_, err = s.db.ExecContext(ctx,
		`UPDATE residents SET password_hash = $1 WHERE tenant_id = $2 AND resident_id::text = $3`,
		passwordHash, req.TenantID, req.ResidentID,
	)
	if err != nil {
		s.logger.Error("ResetResidentPassword failed",
			zap.String("tenant_id", req.TenantID),
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

// ResetContactPassword 重置联系人密码
func (s *residentService) ResetContactPassword(ctx context.Context, req ResetContactPasswordRequest) (*ResetContactPasswordResponse, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.ContactID == "" {
		return nil, fmt.Errorf("contact_id is required")
	}

	// 1. 获取联系人信息（用于权限检查）
	var residentID sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT resident_id::text FROM resident_contacts WHERE tenant_id = $1 AND contact_id::text = $2`,
		req.TenantID, req.ContactID,
	).Scan(&residentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("contact not found")
		}
		return nil, fmt.Errorf("failed to get contact info: %w", err)
	}

	// 2. 权限检查
	if req.CurrentUserType == "resident" || req.CurrentUserType == "family" {
		// Resident/Family: 只能重置自己的联系人密码
		if req.CurrentUserID != residentID.String {
			return nil, fmt.Errorf("access denied: can only reset password for own contacts")
		}
	} else {
		// Staff: 权限检查（AssignedOnly, BranchOnly）
		if req.PermissionCheck != nil {
			if req.PermissionCheck.AssignedOnly && req.CurrentUserID != "" {
				var isAssigned bool
				err := s.db.QueryRowContext(ctx,
					`SELECT EXISTS(
						SELECT 1 FROM resident_caregivers rc
						WHERE rc.tenant_id = $1
						  AND rc.resident_id::text = $2
						  AND (rc.userList::text LIKE $3 OR rc.userList::text LIKE $4)
					)`,
					req.TenantID, residentID.String, req.CurrentUserID, "%\""+req.CurrentUserID+"\"%",
				).Scan(&isAssigned)
				if err != nil {
					return nil, fmt.Errorf("failed to check assignment: %w", err)
				}
				if !isAssigned {
					return nil, fmt.Errorf("permission denied: can only reset password for contacts of assigned residents")
				}
			}

			if req.PermissionCheck.BranchOnly {
				var targetBranchTag sql.NullString
				err := s.db.QueryRowContext(ctx,
					`SELECT COALESCE(u.branch_name, '') as branch_tag
					 FROM residents r
					 LEFT JOIN units u ON u.unit_id = r.unit_id
					 WHERE r.tenant_id = $1 AND r.resident_id::text = $2`,
					req.TenantID, residentID.String,
				).Scan(&targetBranchTag)
				if err != nil {
					return nil, fmt.Errorf("failed to get resident info: %w", err)
				}

				userBranchTag := req.PermissionCheck.UserBranchTag
				if userBranchTag == "" {
					if targetBranchTag.Valid && targetBranchTag.String != "" {
						return nil, fmt.Errorf("permission denied: can only reset password for contacts of residents in units with branch_tag IS NULL")
					}
				} else {
					if !targetBranchTag.Valid || targetBranchTag.String != userBranchTag {
						return nil, fmt.Errorf("permission denied: can only reset password for contacts of residents in units with branch_tag = %s", userBranchTag)
					}
				}
			}
		}
	}

	// 3. 密码哈希（前端已 hash，这里直接解码 hex 字符串）
	// 前端发送的是 SHA256(password) 的 hex 字符串，直接解码为 byte slice
	if req.NewPassword == "" {
		return nil, fmt.Errorf("password hash is required")
	}
	passwordHash, err := hex.DecodeString(req.NewPassword)
	if err != nil || len(passwordHash) == 0 {
		return nil, fmt.Errorf("failed to decode password hash: %w", err)
	}

	// 4. 更新 resident_contacts 表
	_, err = s.db.ExecContext(ctx,
		`UPDATE resident_contacts SET password_hash = $1 WHERE tenant_id = $2 AND contact_id::text = $3`,
		passwordHash, req.TenantID, req.ContactID,
	)
	if err != nil {
		s.logger.Error("ResetContactPassword failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("contact_id", req.ContactID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to reset password: %w", err)
	}

	return &ResetContactPasswordResponse{
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

// GetResidentAccountSettings 获取住户/联系人账户设置（只返回账户设置相关字段）
// 注意：这个 API 只能查看自己的账户设置，不允许查看其他用户的
func (s *residentService) GetResidentAccountSettings(ctx context.Context, req GetResidentAccountSettingsRequest) (*GetResidentAccountSettingsResponse, error) {
	// 1. 权限检查：只能查看自己的账户设置
	if req.CurrentUserID != req.ResidentID {
		return nil, fmt.Errorf("permission denied: can only view own account settings")
	}

	// 2. 判断是 resident 还是 contact
	// 使用 role 来判断：Family = contact, Resident = resident
	var isContact bool
	if req.CurrentUserRole == "Family" {
		isContact = true
	} else {
		// Resident 或其他情况，默认为 resident
		isContact = false
	}

	// 3. 构建响应
	resp := &GetResidentAccountSettingsResponse{
		IsContact: isContact,
	}

	if isContact {
		// Contact: 从 resident_contacts 表获取（逻辑与 Resident 一样，处理占位符）
		var contactEmail, contactPhone sql.NullString
		var contactNickname sql.NullString
		var linkedResidentAccount sql.NullString

		err := s.db.QueryRowContext(ctx,
			`SELECT 
				COALESCE(rc.contact_email, '') as contact_email,
				COALESCE(rc.contact_phone, '') as contact_phone,
				COALESCE(rc.contact_first_name || ' ' || rc.contact_last_name, '') as nickname,
				COALESCE(r.resident_account, '') as resident_account
			 FROM resident_contacts rc
			 JOIN residents r ON r.resident_id = rc.resident_id AND r.tenant_id = rc.tenant_id
			 WHERE rc.tenant_id = $1 AND rc.contact_id::text = $2`,
			req.TenantID, req.ResidentID,
		).Scan(&contactEmail, &contactPhone, &contactNickname, &linkedResidentAccount)

		if err != nil {
			return nil, fmt.Errorf("contact not found: %w", err)
		}

		if linkedResidentAccount.Valid {
			account := linkedResidentAccount.String
			resp.ResidentAccount = &account
		}
		if contactNickname.Valid {
			resp.Nickname = contactNickname.String
		}
		// Contact 的 email/phone 处理逻辑与 Resident 一样（处理占位符）
		if contactEmail.Valid && contactEmail.String != "" && contactEmail.String != "***@***" {
			email := contactEmail.String
			resp.Email = &email
			resp.SaveEmail = true // 如果存在且不是占位符，说明已保存
		}
		if contactPhone.Valid && contactPhone.String != "" && contactPhone.String != "xxx-xxx-xxxx" {
			phone := contactPhone.String
			resp.Phone = &phone
			resp.SavePhone = true // 如果存在且不是占位符，说明已保存
		}
	} else {
		// Resident: 从 residents 和 resident_phi 表获取
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
			return nil, fmt.Errorf("resident not found: %w", err)
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
	}

	return resp, nil
}

// ============================================
// UpdateResidentAccountSettings 更新住户/联系人账户设置（统一 API）
// ============================================

// UpdateResidentAccountSettings 更新住户/联系人账户设置（在同一个事务中处理所有更新）
// 注意：这个 API 只能更新自己的账户设置，不允许更新其他用户的
func (s *residentService) UpdateResidentAccountSettings(ctx context.Context, req UpdateResidentAccountSettingsRequest) (*UpdateResidentAccountSettingsResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" || req.ResidentID == "" || req.CurrentUserID == "" {
		return nil, fmt.Errorf("tenant_id, resident_id, and current_user_id are required")
	}

	// 2. 权限检查：只能更新自己的账户设置
	if req.CurrentUserID != req.ResidentID {
		return nil, fmt.Errorf("permission denied: can only update own account settings")
	}

	// 3. 判断是 resident 还是 contact
	// 使用 role 来判断：Family = contact, Resident = resident
	var isContact bool
	if req.CurrentUserRole == "Family" {
		isContact = true
	} else {
		// Resident 或其他情况，默认为 resident
		isContact = false
	}

	// 4. 开始事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if isContact {
		// Contact: 更新 resident_contacts 表
		updates := []string{}
		args := []interface{}{}
		argIdx := 1

		// 更新密码（如果提供，!= nil 就更新，不进行任何判断）
		if req.PasswordHash != nil {
			passwordHashBytes, _ := hex.DecodeString(*req.PasswordHash)
			updates = append(updates, fmt.Sprintf("password_hash = $%d", argIdx))
			args = append(args, passwordHashBytes)
			argIdx++
		}

		// 更新 email（如果提供，!= nil 就更新，不进行任何判断，直接传值）
		if req.Email != nil {
			updates = append(updates, fmt.Sprintf("contact_email = $%d", argIdx))
			args = append(args, *req.Email)
			argIdx++
		}

		// 更新 email_hash（如果提供，!= nil 就更新，不进行任何判断，直接传值）
		if req.EmailHash != nil {
			updates = append(updates, fmt.Sprintf("email_hash = $%d", argIdx))
			emailHashBytes, _ := hex.DecodeString(*req.EmailHash)
			args = append(args, emailHashBytes)
			argIdx++
		}

		// 更新 phone（如果提供，!= nil 就更新，不进行任何判断，直接传值）
		if req.Phone != nil {
			updates = append(updates, fmt.Sprintf("contact_phone = $%d", argIdx))
			args = append(args, *req.Phone)
			argIdx++
		}

		// 更新 phone_hash（如果提供，!= nil 就更新，不进行任何判断，直接传值）
		if req.PhoneHash != nil {
			updates = append(updates, fmt.Sprintf("phone_hash = $%d", argIdx))
			phoneHashBytes, _ := hex.DecodeString(*req.PhoneHash)
			args = append(args, phoneHashBytes)
			argIdx++
		}

		if len(updates) > 0 {
			query := fmt.Sprintf(
				`UPDATE resident_contacts SET %s WHERE tenant_id = $%d AND contact_id::text = $%d`,
				strings.Join(updates, ", "), argIdx, argIdx+1,
			)
			args = append(args, req.TenantID, req.ResidentID)
			_, err = tx.ExecContext(ctx, query, args...)
			if err != nil {
				return nil, fmt.Errorf("failed to update contact: %w", err)
			}
		}
	} else {
		// Resident: 更新 residents 和 resident_phi 表
		// 4.1 更新密码（residents 表，如果提供，!= nil 就更新，不进行任何判断）
		if req.PasswordHash != nil {
			passwordHashBytes, _ := hex.DecodeString(*req.PasswordHash)
			_, err = tx.ExecContext(ctx,
				`UPDATE residents SET password_hash = $1 WHERE tenant_id = $2 AND resident_id::text = $3`,
				passwordHashBytes, req.TenantID, req.ResidentID,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to update password: %w", err)
			}
		}

		// 4.2 更新 email/phone hash（residents 表，用于登录）
		residentUpdates := []string{}
		residentArgs := []interface{}{}
		residentArgIdx := 1

		// 更新 email_hash（如果提供，!= nil 就更新，不进行任何判断，直接传值）
		if req.EmailHash != nil {
			emailHashBytes, _ := hex.DecodeString(*req.EmailHash)
			residentUpdates = append(residentUpdates, fmt.Sprintf("email_hash = $%d", residentArgIdx))
			residentArgs = append(residentArgs, emailHashBytes)
			residentArgIdx++
		}

		// 更新 phone_hash（如果提供，!= nil 就更新，不进行任何判断，直接传值）
		if req.PhoneHash != nil {
			phoneHashBytes, _ := hex.DecodeString(*req.PhoneHash)
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

		// 4.3 更新 email/phone 明文（resident_phi 表，根据 save 标志）
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
func (s *residentService) UpdateResidentContact(ctx context.Context, req UpdateResidentContactRequest) (*UpdateResidentContactResponse, error) {
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

	// 3. 查找或创建 contact（通过 resident_id + slot 定位）
	// 如果不存在，则创建；如果存在，则更新（upsert 逻辑，类似 PHI 的 UpsertResidentPHI）
	var contactID string
	var isNewlyCreated bool                // 标记是否是新创建的 contact
	var newContact *domain.ResidentContact // 保存新创建的 contact 对象，用于更新逻辑
	err := s.db.QueryRowContext(ctx,
		`SELECT contact_id::text FROM resident_contacts WHERE tenant_id = $1 AND resident_id::text = $2 AND slot = $3`,
		req.TenantID, req.ResidentID, req.Slot,
	).Scan(&contactID)
	if err != nil {
		if err == sql.ErrNoRows {
			isNewlyCreated = true
			// Contact 不存在，需要创建（类似 PHI 的 UpsertResidentPHI 逻辑）
			// 先构建一个基本的 contact 对象用于创建
			defaultPassword := "ChangeMe123!"
			defaultPasswordHashHex := HashPassword(defaultPassword)
			defaultPasswordHash, _ := hex.DecodeString(defaultPasswordHashHex)

			newContact = &domain.ResidentContact{
				Slot:               req.Slot,
				IsEnabled:          false, // 默认禁用
				Role:               "Family",
				IsEmergencyContact: false,
				PasswordHash:       defaultPasswordHash,
			}
			// 如果请求中提供了字段，使用请求的值
			if req.IsEnabled != nil {
				newContact.IsEnabled = *req.IsEnabled
			}
			if req.Relationship != nil {
				newContact.Relationship = *req.Relationship
			}
			if req.ContactFirstName != nil {
				newContact.ContactFirstName = *req.ContactFirstName
			}
			if req.ContactLastName != nil {
				newContact.ContactLastName = *req.ContactLastName
			}
			if req.ContactPhone != nil {
				newContact.ContactPhone = *req.ContactPhone
			}
			if req.ContactEmail != nil {
				newContact.ContactEmail = *req.ContactEmail
			}
			if req.ReceiveSMS != nil {
				newContact.ReceiveSMS = *req.ReceiveSMS
			}
			if req.ReceiveEmail != nil {
				newContact.ReceiveEmail = *req.ReceiveEmail
			}
			// 解析 phone_hash 和 email_hash
			if req.PhoneHash != nil && *req.PhoneHash != "" {
				ph, err := hex.DecodeString(*req.PhoneHash)
				if err == nil {
					newContact.PhoneHash = ph
				}
			}
			if req.EmailHash != nil && *req.EmailHash != "" {
				eh, err := hex.DecodeString(*req.EmailHash)
				if err == nil {
					newContact.EmailHash = eh
				}
			}
			if req.PasswordHash != nil && *req.PasswordHash != "" {
				hashBytes, err := hex.DecodeString(*req.PasswordHash)
				if err == nil {
					newContact.PasswordHash = hashBytes
				}
			}

			// 创建 contact
			contactID, err = s.residentsRepo.CreateResidentContact(ctx, req.TenantID, req.ResidentID, newContact)
			if err != nil {
				return nil, fmt.Errorf("failed to create contact: %w", err)
			}
			// 创建成功后，标记为刚创建的 contact，更新逻辑中应保留创建时的值
			// 继续执行更新逻辑（使用新创建的 contactID）以应用其他字段的更新
		} else {
			return nil, fmt.Errorf("failed to get contact: %w", err)
		}
	}

	// 4. 解析 phone_hash 和 email_hash，如果已存在则复用联系人信息
	var phoneHash, emailHash []byte
	if req.PhoneHash != nil {
		if *req.PhoneHash != "" {
			ph, err := hex.DecodeString(*req.PhoneHash)
			if err != nil {
				return nil, fmt.Errorf("invalid phone_hash format: %w", err)
			}
			phoneHash = ph
		}
	}
	if req.EmailHash != nil {
		if *req.EmailHash != "" {
			eh, err := hex.DecodeString(*req.EmailHash)
			if err != nil {
				return nil, fmt.Errorf("invalid email_hash format: %w", err)
			}
			emailHash = eh
		}
	}

	// 如果更新 phone_hash 或 email_hash，查找已存在的联系人信息（排除当前 contact）
	var existingContact *domain.ResidentContact
	if (req.PhoneHash != nil && len(phoneHash) > 0) || (req.EmailHash != nil && len(emailHash) > 0) {
		existingContact = s.findExistingContactByHashExcluding(ctx, req.TenantID, phoneHash, emailHash, contactID)
		// 如果找到已存在的联系人，说明是同一个联系人，复用其信息
	}

	// 5. 构建 domain.ResidentContact 对象
	contact := &domain.ResidentContact{
		ContactID: contactID, // 从步骤3查询得到
		Slot:      req.Slot,  // slot 是必填的，直接使用
	}
	if req.IsEnabled != nil {
		contact.IsEnabled = *req.IsEnabled
	}
	if req.Relationship != nil {
		contact.Relationship = *req.Relationship
	} else if existingContact != nil && existingContact.Relationship != "" {
		// 如果请求中未提供，复用已存在的值
		contact.Relationship = existingContact.Relationship
	}
	if req.ContactFirstName != nil {
		contact.ContactFirstName = *req.ContactFirstName
	} else if existingContact != nil && existingContact.ContactFirstName != "" {
		// 如果请求中未提供，复用已存在的值
		contact.ContactFirstName = existingContact.ContactFirstName
	}
	if req.ContactLastName != nil {
		contact.ContactLastName = *req.ContactLastName
	} else if existingContact != nil && existingContact.ContactLastName != "" {
		// 如果请求中未提供，复用已存在的值
		contact.ContactLastName = existingContact.ContactLastName
	}
	// 更新 contact_phone（只要 != nil 就更新，参考 UpdateAccountSetting）
	if req.ContactPhone != nil {
		contact.ContactPhone = *req.ContactPhone // "" 表示删除（Repository 会处理为 NULL）
	} else if existingContact != nil && existingContact.ContactPhone != "" {
		// 如果请求中未提供，复用已存在的值
		contact.ContactPhone = existingContact.ContactPhone
	}
	// 更新 contact_email（只要 != nil 就更新，参考 UpdateAccountSetting）
	// 注意：如果是新创建的 contact，且创建时已经设置了 contact_email，更新时不应该用空字符串覆盖
	if req.ContactEmail != nil {
		// 如果是新创建的 contact，且创建时已经设置了非空值，更新时如果是空字符串，不应该覆盖
		if isNewlyCreated && newContact != nil && newContact.ContactEmail != "" && *req.ContactEmail == "" {
			// 新创建的 contact，创建时已有值，更新时为空字符串，保留创建时的值
			contact.ContactEmail = newContact.ContactEmail
		} else {
			contact.ContactEmail = *req.ContactEmail // "" 表示删除（Repository 会处理为 NULL）
		}
	} else if existingContact != nil && existingContact.ContactEmail != "" {
		// 如果请求中未提供，复用已存在的值
		contact.ContactEmail = existingContact.ContactEmail
	} else if isNewlyCreated && newContact != nil && newContact.ContactEmail != "" {
		// 如果是新创建的 contact，且创建时已经设置了值，但更新时未提供，保留创建时的值
		contact.ContactEmail = newContact.ContactEmail
	}
	if req.ReceiveSMS != nil {
		contact.ReceiveSMS = *req.ReceiveSMS
	}
	if req.ReceiveEmail != nil {
		contact.ReceiveEmail = *req.ReceiveEmail
	}
	// 更新 phone_hash（只要 != nil 就更新，参考 UpdateAccountSetting）
	if req.PhoneHash != nil {
		if *req.PhoneHash == "" {
			contact.PhoneHash = []byte{} // 空字符串，删除 phone_hash（Repository 会处理为 NULL）
		} else {
			contact.PhoneHash = phoneHash
		}
	}
	// 更新 email_hash（只要 != nil 就更新，参考 UpdateAccountSetting）
	if req.EmailHash != nil {
		if *req.EmailHash == "" {
			contact.EmailHash = []byte{} // 空字符串，删除 email_hash（Repository 会处理为 NULL）
		} else {
			contact.EmailHash = emailHash
		}
	}
	// 更新 password_hash
	// 规则：passwd 是不回显的，没有从密码改为无密码的状态转换，所以不能发送 ""
	// vue 要么发送有效 password 的 hash，要么不发送该字段，表示 passwd 未修改
	// 如果 req.PasswordHash 为 nil（未传递），不设置 contact.PasswordHash（保持为零值 nil，repository 不会更新）
	// 如果 req.PasswordHash 有值，设置 contact.PasswordHash（repository 会更新）
	if req.PasswordHash != nil && *req.PasswordHash != "" {
		hashBytes, err := hex.DecodeString(*req.PasswordHash)
		if err != nil {
			return nil, fmt.Errorf("invalid password_hash format: %w", err)
		}
		contact.PasswordHash = hashBytes
	}
	// 如果 req.PasswordHash 为 nil 或空字符串，contact.PasswordHash 保持为零值 nil（不更新）

	// 6. 调用 Repository 更新
	err = s.residentsRepo.UpdateResidentContact(ctx, req.TenantID, contactID, contact)
	if err != nil {
		return nil, fmt.Errorf("failed to update contact: %w", err)
	}

	return &UpdateResidentContactResponse{
		Success: true,
	}, nil
}
