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

	// 删除
	DeleteResident(ctx context.Context, req DeleteResidentRequest) (*DeleteResidentResponse, error)

	// 密码管理
	ResetResidentPassword(ctx context.Context, req ResetResidentPasswordRequest) (*ResetResidentPasswordResponse, error)
	ResetContactPassword(ctx context.Context, req ResetContactPasswordRequest) (*ResetContactPasswordResponse, error)
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

// PermissionCheckResult 权限检查结果（从 Handler 层传入）
type PermissionCheckResult struct {
	AssignedOnly bool   // 是否仅限分配的资源
	BranchOnly   bool   // 是否仅限同一 Branch 的资源
	UserBranchTag string // 用户的 branch_tag（用于分支过滤）
}

// ListResidentsResponse 查询住户列表响应
type ListResidentsResponse struct {
	Items []*ResidentListItemDTO // 住户列表
	Total int                    // 总数量
}

// ResidentListItemDTO 住户列表项 DTO
type ResidentListItemDTO struct {
	ResidentID       string  `json:"resident_id"`
	TenantID         string  `json:"tenant_id"`
	ResidentAccount  *string `json:"resident_account,omitempty"`
	Nickname         string  `json:"nickname"`
	Status           string  `json:"status"`
	ServiceLevel     *string `json:"service_level,omitempty"`
	AdmissionDate    *int64  `json:"admission_date,omitempty"`    // Unix timestamp
	DischargeDate    *int64  `json:"discharge_date,omitempty"`    // Unix timestamp
	FamilyTag        *string `json:"family_tag,omitempty"`
	UnitID           *string `json:"unit_id,omitempty"`
	UnitName         *string `json:"unit_name,omitempty"`
	BranchTag        *string `json:"branch_tag,omitempty"`
	AreaTag          *string `json:"area_tag,omitempty"`
	UnitNumber       *string `json:"unit_number,omitempty"`
	IsMultiPersonRoom bool   `json:"is_multi_person_room"`
	RoomID           *string `json:"room_id,omitempty"`
	RoomName         *string `json:"room_name,omitempty"`
	BedID            *string `json:"bed_id,omitempty"`
	BedName          *string `json:"bed_name,omitempty"`
	IsAccessEnabled  bool   `json:"is_access_enabled"`
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
	Resident *ResidentDetailDTO  `json:"resident"`
	PHI      *ResidentPHIDTO      `json:"phi,omitempty"`
	Contacts []*ResidentContactDTO `json:"contacts,omitempty"`
}

// ResidentDetailDTO 住户详情 DTO
type ResidentDetailDTO struct {
	ResidentID      string  `json:"resident_id"`
	TenantID        string  `json:"tenant_id"`
	ResidentAccount *string `json:"resident_account,omitempty"`
	Nickname        string  `json:"nickname"`
	Status          string  `json:"status"`
	ServiceLevel    *string `json:"service_level,omitempty"`
	AdmissionDate   *int64  `json:"admission_date,omitempty"`   // Unix timestamp
	DischargeDate   *int64  `json:"discharge_date,omitempty"`  // Unix timestamp
	FamilyTag       *string `json:"family_tag,omitempty"`
	UnitID          *string `json:"unit_id,omitempty"`
	UnitName        *string `json:"unit_name,omitempty"`
	BranchTag       *string `json:"branch_tag,omitempty"`
	AreaTag         *string `json:"area_tag,omitempty"`
	UnitNumber      *string `json:"unit_number,omitempty"`
	IsMultiPersonRoom bool  `json:"is_multi_person_room"`
	RoomID          *string `json:"room_id,omitempty"`
	RoomName        *string `json:"room_name,omitempty"`
	BedID           *string `json:"bed_id,omitempty"`
	BedName         *string `json:"bed_name,omitempty"`
	IsAccessEnabled bool   `json:"is_access_enabled"`
	Note            *string `json:"note,omitempty"`
}

// ResidentPHIDTO 住户 PHI 数据 DTO
type ResidentPHIDTO struct {
	PhiID                    string     `json:"phi_id"`
	FirstName                *string    `json:"first_name,omitempty"`
	LastName                 *string    `json:"last_name,omitempty"`
	Gender                   *string    `json:"gender,omitempty"`
	DateOfBirth              *int64    `json:"date_of_birth,omitempty"` // Unix timestamp
	ResidentPhone            *string    `json:"resident_phone,omitempty"`
	ResidentEmail            *string    `json:"resident_email,omitempty"`
	WeightLb                 *float64   `json:"weight_lb,omitempty"`
	HeightFt                 *float64   `json:"height_ft,omitempty"`
	HeightIn                 *float64   `json:"height_in,omitempty"`
	MobilityLevel            *int       `json:"mobility_level,omitempty"`
	TremorStatus             *string    `json:"tremor_status,omitempty"`
	MobilityAid              *string    `json:"mobility_aid,omitempty"`
	ADLAssistance            *string    `json:"adl_assistance,omitempty"`
	CommStatus               *string    `json:"comm_status,omitempty"`
	HasHypertension          *bool      `json:"has_hypertension,omitempty"`
	HasHyperlipaemia         *bool      `json:"has_hyperlipaemia,omitempty"`
	HasHyperglycaemia        *bool      `json:"has_hyperglycaemia,omitempty"`
	HasStrokeHistory         *bool      `json:"has_stroke_history,omitempty"`
	HasParalysis             *bool      `json:"has_paralysis,omitempty"`
	HasAlzheimer             *bool      `json:"has_alzheimer,omitempty"`
	MedicalHistory           *string    `json:"medical_history,omitempty"`
	HISResidentName          *string    `json:"HIS_resident_name,omitempty"`
	HISResidentAdmissionDate *int64     `json:"HIS_resident_admission_date,omitempty"` // Unix timestamp
	HISResidentDischargeDate *int64     `json:"HIS_resident_discharge_date,omitempty"` // Unix timestamp
	HomeAddressStreet        *string    `json:"home_address_street,omitempty"`
	HomeAddressCity          *string    `json:"home_address_city,omitempty"`
	HomeAddressState         *string    `json:"home_address_state,omitempty"`
	HomeAddressPostalCode    *string    `json:"home_address_postal_code,omitempty"`
	PlusCode                 *string    `json:"plus_code,omitempty"`
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
	Password        string  // 密码（默认 "ChangeMe123!"）
	Status          string  // 状态（默认 "active"）
	ServiceLevel    string  // 护理级别
	AdmissionDate   *int64  // 入院日期（Unix timestamp，默认当前日期）
	UnitID          string  // 单元ID
	FamilyTag       string  // 家庭标签
	IsAccessEnabled bool    // 是否允许查看状态
	Note            string  // 备注

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
	FirstName       string  // 必填（创建时）
	LastName        string  // 可选
	Gender          string  // 可选
	DateOfBirth     *int64  // Unix timestamp
	ResidentPhone   string  // 明文（可选保存）
	ResidentEmail   string  // 明文（可选保存）
	SavePhone       bool    // 是否保存明文 phone
	SaveEmail       bool    // 是否保存明文 email
	WeightLb        *float64
	HeightFt        *float64
	HeightIn        *float64
	MobilityLevel   *int
	TremorStatus    string
	MobilityAid      string
	ADLAssistance   string
	CommStatus      string
	HasHypertension *bool
	HasHyperlipaemia *bool
	HasHyperglycaemia *bool
	HasStrokeHistory *bool
	HasParalysis     *bool
	HasAlzheimer     *bool
	MedicalHistory   string
	HISResidentName  string
	HISResidentAdmissionDate *int64
	HISResidentDischargeDate *int64
	HomeAddressStreet        string
	HomeAddressCity          string
	HomeAddressState         string
	HomeAddressPostalCode    string
	PlusCode                 string
}

// CreateResidentContactRequest 创建住户联系人请求
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
	CurrentUserRole string // 当前用户角色

	// 权限检查结果（由 Handler 层传入）
	PermissionCheck *PermissionCheckResult // 权限检查结果（仅 staff 需要）

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
	TenantID        string // 必填
	ResidentID      string // 必填
	CurrentUserID   string // 当前用户ID
	CurrentUserType string // 当前用户类型
	CurrentUserRole string // 当前用户角色
	UserBranchTag   *string // 当前用户 BranchTag (用于权限过滤)
	PermissionCheck *PermissionCheckResult // 权限检查结果
	NewPassword     string // 新密码（可选，默认生成）
}

// ResetResidentPasswordResponse 重置住户密码响应
type ResetResidentPasswordResponse struct {
	Success     bool
	NewPassword string // 生成的新密码
}

// ResetContactPasswordRequest 重置联系人密码请求
type ResetContactPasswordRequest struct {
	TenantID        string // 必填
	ContactID       string // 必填
	CurrentUserID   string // 当前用户ID
	CurrentUserType string // 当前用户类型
	CurrentUserRole string // 当前用户角色
	UserBranchTag   *string // 当前用户 BranchTag (用于权限过滤)
	PermissionCheck *PermissionCheckResult // 权限检查结果
	NewPassword     string // 新密码（可选，默认生成）
}

// ResetContactPasswordResponse 重置联系人密码响应
type ResetContactPasswordResponse struct {
	Success     bool
	NewPassword string // 生成的新密码
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
			ResidentID:       residentID.String,
			TenantID:         tid.String,
			Nickname:         nickname.String,
			Status:           status.String,
			IsMultiPersonRoom: isMultiPersonRoom,
			IsAccessEnabled:  canViewStatus,
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
	countQuery := strings.Replace(q, "SELECT r.resident_id::text, r.tenant_id::text, r.resident_account, r.nickname,\n\t             r.status, r.service_level, r.admission_date, r.discharge_date,\n\t             r.family_tag, r.unit_id::text, r.room_id::text, r.bed_id::text,\n\t             COALESCE(u.unit_name, '') as unit_name,\n\t             COALESCE(u.branch_tag, '') as branch_tag,\n\t             COALESCE(u.area_tag, '') as area_tag,\n\t             COALESCE(u.unit_number, '') as unit_number,\n\t             COALESCE(u.is_multi_person_room, false) as is_multi_person_room,\n\t             COALESCE(rm.room_name, '') as room_name,\n\t             COALESCE(b.bed_name, '') as bed_name,\n\t             r.can_view_status\n	      FROM residents r\n	      LEFT JOIN units u ON u.unit_id = r.unit_id\n	      LEFT JOIN rooms rm ON rm.room_id = r.room_id\n	      LEFT JOIN beds b ON b.bed_id = r.bed_id", "SELECT COUNT(*)\n	      FROM residents r\n	      LEFT JOIN units u ON u.unit_id = r.unit_id", 1)
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
	if req.CurrentUserType == "resident" || req.CurrentUserType == "family" {
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
		        COALESCE(u.branch_tag, '') as branch_tag,
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
		IsAccessEnabled:  canViewStatus,
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
	if phi.HISResidentName != "" {
		dto.HISResidentName = &phi.HISResidentName
	}
	if phi.HISResidentAdmissionDate != nil {
		ts := phi.HISResidentAdmissionDate.Unix()
		dto.HISResidentAdmissionDate = &ts
	}
	if phi.HISResidentDischargeDate != nil {
		ts := phi.HISResidentDischargeDate.Unix()
		dto.HISResidentDischargeDate = &ts
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
		phi.HISResidentName = req.PHI.HISResidentName
		if req.PHI.HISResidentAdmissionDate != nil {
			phi.HISResidentAdmissionDate = unixTimestampToTime(req.PHI.HISResidentAdmissionDate)
		}
		if req.PHI.HISResidentDischargeDate != nil {
			phi.HISResidentDischargeDate = unixTimestampToTime(req.PHI.HISResidentDischargeDate)
		}
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

			if err := s.checkHashUniqueness(ctx, req.TenantID, "resident_contacts", contactPhoneHash, contactEmailHash, "", ""); err != nil {
				s.logger.Warn("Contact hash uniqueness check failed",
					zap.String("tenant_id", req.TenantID),
					zap.String("resident_id", residentID),
					zap.Error(err),
				)
				// 不失败整个操作，只记录警告并跳过该联系人
				continue
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

// checkHashUniqueness 检查 phone_hash 或 email_hash 的唯一性
func (s *residentService) checkHashUniqueness(ctx context.Context, tenantID, tableName string, phoneHash, emailHash []byte, excludeID, excludeField string) error {
	if phoneHash != nil && len(phoneHash) > 0 {
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
	if emailHash != nil && len(emailHash) > 0 {
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

	// 1. 权限检查
	if req.CurrentUserType == "resident" || req.CurrentUserType == "family" {
		// Resident/Family: 只能更新自己
		if req.CurrentUserID != req.ResidentID {
			return nil, fmt.Errorf("access denied: can only update own information")
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
					return nil, fmt.Errorf("permission denied: can only update assigned residents")
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
						return nil, fmt.Errorf("permission denied: can only update residents in units with branch_tag IS NULL")
					}
				} else {
					if !targetBranchTag.Valid || targetBranchTag.String != userBranchTag {
						return nil, fmt.Errorf("permission denied: can only update residents in units with branch_tag = %s", userBranchTag)
					}
				}
			}
		}
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

	// 5. 更新 PHI 数据（如果提供了）
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
		// 其他 PHI 字段类似处理...
		// 注意：phone/email 的更新需要特殊处理（save_phone/save_email 标志）

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
				`SELECT COALESCE(u.branch_tag, '') as branch_tag
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

	// 2. 生成新密码（如果未提供）
	newPassword := req.NewPassword
	if newPassword == "" {
		// 生成随机密码（12位，包含字母和数字）
		newPassword = generateRandomPassword(12)
	}

	// 3. 计算 password_hash
	passwordHashHex := HashPassword(newPassword)
	passwordHash, err := hex.DecodeString(passwordHashHex)
	if err != nil || len(passwordHash) == 0 {
		return nil, fmt.Errorf("failed to hash password")
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
		NewPassword: newPassword,
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

	// 3. 生成新密码（如果未提供）
	newPassword := req.NewPassword
	if newPassword == "" {
		newPassword = generateRandomPassword(12)
	}

	// 4. 计算 password_hash
	passwordHashHex := HashPassword(newPassword)
	passwordHash, err := hex.DecodeString(passwordHashHex)
	if err != nil || len(passwordHash) == 0 {
		return nil, fmt.Errorf("failed to hash password")
	}

	// 5. 更新 resident_contacts 表
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
		NewPassword: newPassword,
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

