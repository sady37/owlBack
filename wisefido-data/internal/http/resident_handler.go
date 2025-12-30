package httpapi

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/service"

	"go.uber.org/zap"
)

// ResidentHandler 住户管理 Handler
type ResidentHandler struct {
	residentService service.ResidentService
	logger          *zap.Logger
	base            *StubHandler // 用于 tenantIDFromReq
	db              *sql.DB      // 用于权限检查
}

// NewResidentHandler 创建住户管理 Handler
func NewResidentHandler(residentService service.ResidentService, db *sql.DB, logger *zap.Logger) *ResidentHandler {
	return &ResidentHandler{
		residentService: residentService,
		logger:          logger,
		base:            &StubHandler{},
		db:              db,
	}
}

// ServeHTTP 实现 http.Handler 接口
func (h *ResidentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	// ListResidents
	case path == "/admin/api/v1/residents" && r.Method == http.MethodGet:
		h.ListResidents(w, r)
	// CreateResident
	case path == "/admin/api/v1/residents" && r.Method == http.MethodPost:
		h.CreateResident(w, r)
	// GetResidentAccountSettings (必须在 GetResident 之前，因为路径更具体)
	// 处理 /admin/api/v1/residents/:id/account-settings
	case strings.HasPrefix(path, "/admin/api/v1/residents/") && strings.HasSuffix(path, "/account-settings") && r.Method == http.MethodGet:
		residentID := strings.TrimSuffix(path, "/account-settings")
		residentID = strings.TrimPrefix(residentID, "/admin/api/v1/residents/")
		if residentID != "" && !strings.Contains(residentID, "/") {
			h.GetResidentAccountSettings(w, r, residentID)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	// GetResidentAccountSettings for contacts (必须在 GetResident 之前，因为路径更具体)
	// 处理 /admin/api/v1/contacts/:id/account-settings
	case strings.HasPrefix(path, "/admin/api/v1/contacts/") && strings.HasSuffix(path, "/account-settings") && r.Method == http.MethodGet:
		contactID := strings.TrimSuffix(path, "/account-settings")
		contactID = strings.TrimPrefix(contactID, "/admin/api/v1/contacts/")
		if contactID != "" && !strings.Contains(contactID, "/") {
			h.GetResidentAccountSettings(w, r, contactID)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	// UpdateResidentAccountSettings (必须在 UpdateResident 之前，因为路径更具体)
	// 处理 /admin/api/v1/residents/:id/account-settings
	case strings.HasPrefix(path, "/admin/api/v1/residents/") && strings.HasSuffix(path, "/account-settings") && r.Method == http.MethodPut:
		residentID := strings.TrimSuffix(path, "/account-settings")
		residentID = strings.TrimPrefix(residentID, "/admin/api/v1/residents/")
		if residentID != "" && !strings.Contains(residentID, "/") {
			h.UpdateResidentAccountSettings(w, r, residentID)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	// UpdateResidentAccountSettings for contacts (必须在 UpdateResident 之前，因为路径更具体)
	// 处理 /admin/api/v1/contacts/:id/account-settings
	case strings.HasPrefix(path, "/admin/api/v1/contacts/") && strings.HasSuffix(path, "/account-settings") && r.Method == http.MethodPut:
		contactID := strings.TrimSuffix(path, "/account-settings")
		contactID = strings.TrimPrefix(contactID, "/admin/api/v1/contacts/")
		if contactID != "" && !strings.Contains(contactID, "/") {
			h.UpdateResidentAccountSettings(w, r, contactID)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	// GetResident
	case strings.HasPrefix(path, "/admin/api/v1/residents/") && r.Method == http.MethodGet:
		residentID := strings.TrimPrefix(path, "/admin/api/v1/residents/")
		// 处理子路径（如 /contacts, /reset-password）
		if strings.Contains(residentID, "/") {
			parts := strings.Split(residentID, "/")
			if len(parts) == 2 {
				switch parts[1] {
				case "contacts":
					// GET /admin/api/v1/residents/:id/contacts - 获取联系人列表（已包含在 GetResident 中）
					h.GetResident(w, r, parts[0])
				case "reset-password":
					// POST /admin/api/v1/residents/:id/reset-password
					if r.Method == http.MethodPost {
						h.ResetResidentPassword(w, r, parts[0])
					} else {
						w.WriteHeader(http.StatusMethodNotAllowed)
					}
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		} else if residentID != "" {
			h.GetResident(w, r, residentID)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	// UpdateResidentPHI - PUT /admin/api/v1/residents/:id/phi
	case strings.HasSuffix(path, "/phi") && r.Method == http.MethodPut:
		residentID := strings.TrimSuffix(path, "/phi")
		residentID = strings.TrimPrefix(residentID, "/admin/api/v1/residents/")
		if residentID != "" && !strings.Contains(residentID, "/") {
			h.UpdateResidentPHI(w, r, residentID)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	// UpdateResidentContact - PUT /admin/api/v1/residents/:id/contacts (必须在 UpdateResident 之前，因为路径更具体)
	case strings.HasSuffix(path, "/contacts") && strings.HasPrefix(path, "/admin/api/v1/residents/") && r.Method == http.MethodPut:
		residentID := strings.TrimSuffix(path, "/contacts")
		residentID = strings.TrimPrefix(residentID, "/admin/api/v1/residents/")
		if residentID != "" && !strings.Contains(residentID, "/") {
			h.UpdateResidentContact(w, r, residentID)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	// UpdateResident - PUT /admin/api/v1/residents/:id
	case strings.HasPrefix(path, "/admin/api/v1/residents/") && r.Method == http.MethodPut:
		residentID := strings.TrimPrefix(path, "/admin/api/v1/residents/")
		if residentID != "" && !strings.Contains(residentID, "/") {
			h.UpdateResident(w, r, residentID)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	// DeleteResident
	case strings.HasPrefix(path, "/admin/api/v1/residents/") && r.Method == http.MethodDelete:
		residentID := strings.TrimPrefix(path, "/admin/api/v1/residents/")
		if residentID != "" && !strings.Contains(residentID, "/") {
			h.DeleteResident(w, r, residentID)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	// ResetResidentPassword
	case strings.HasSuffix(path, "/reset-password") && r.Method == http.MethodPost:
		residentID := strings.TrimSuffix(path, "/reset-password")
		residentID = strings.TrimPrefix(residentID, "/admin/api/v1/residents/")
		if residentID != "" && !strings.Contains(residentID, "/") {
			h.ResetResidentPassword(w, r, residentID)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// ============================================
// ListResidents 查询住户列表
// ============================================

func (h *ResidentHandler) ListResidents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	currentUserID := r.Header.Get("X-User-Id")
	currentUserType := r.Header.Get("X-User-Type")
	currentUserRole := r.Header.Get("X-User-Role")

	// 获取查询参数
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	serviceLevel := strings.TrimSpace(r.URL.Query().Get("service_level"))
	page := parseInt(r.URL.Query().Get("page"), 1)
	pageSize := parseInt(r.URL.Query().Get("size"), 20)

	// 权限检查（仅 Staff 需要）
	// Service 层会自己查询用户的 branch_id（通过 user_branches 表），这里不需要传递 UserBranchTag
	var permCheck *service.PermissionCheckResult
	// 注意：resident_contacts 不能登录系统，所以 currentUserType 永远不会是 "family"
	// 保留此检查是为了向后兼容，但实际上只会是 "resident" 或 "staff"
	if currentUserType != "resident" && currentUserType != "family" && currentUserRole != "" && h.db != nil {
		// 检查权限
		perm, err := GetResourcePermission(h.db, ctx, currentUserRole, "residents", "R")
		if err == nil {
			permCheck = &service.PermissionCheckResult{
				AssignedOnly: perm.AssignedOnly,
				BranchOnly:   perm.BranchOnly,
			}
		}
	}

	req := service.ListResidentsRequest{
		TenantID:        tenantID,
		CurrentUserID:   currentUserID,
		CurrentUserType: currentUserType,
		CurrentUserRole: currentUserRole,
		PermissionCheck: permCheck,
		Search:          search,
		Status:          status,
		ServiceLevel:    serviceLevel,
		Page:            page,
		PageSize:        pageSize,
	}

	resp, err := h.residentService.ListResidents(ctx, req)
	if err != nil {
		h.logger.Error("ListResidents failed",
			zap.String("tenant_id", tenantID),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 转换为旧 Handler 格式
	items := make([]any, 0, len(resp.Items))
	for _, item := range resp.Items {
		itemMap := map[string]any{
			"resident_id":       item.ResidentID,
			"tenant_id":         item.TenantID,
			"nickname":          item.Nickname,
			"status":            item.Status,
			"is_access_enabled": item.IsAccessEnabled,
		}
		if item.ResidentAccount != nil {
			itemMap["resident_account"] = *item.ResidentAccount
		}
		if item.ServiceLevel != nil {
			itemMap["service_level"] = *item.ServiceLevel
		}
		if item.AdmissionDate != nil {
			itemMap["admission_date"] = time.Unix(*item.AdmissionDate, 0).Format("2006-01-02")
		}
		if item.DischargeDate != nil {
			itemMap["discharge_date"] = time.Unix(*item.DischargeDate, 0).Format("2006-01-02")
		}
		if item.UnitID != nil {
			itemMap["unit_id"] = *item.UnitID
		}
		if item.UnitName != nil {
			itemMap["unit_name"] = *item.UnitName
		}
		if item.BranchID != nil {
			itemMap["branch_id"] = *item.BranchID
		}
		if item.BranchName != nil {
			itemMap["branch_name"] = *item.BranchName
			// 保持向后兼容：同时设置 branch_tag
			itemMap["branch_tag"] = *item.BranchName
		}
		if item.BuildingName != nil {
			itemMap["building_name"] = *item.BuildingName
			// 保持向后兼容：同时设置 building
			itemMap["building"] = *item.BuildingName
		}
		// is_shared_unit 现在为 *bool，未绑定 unit 时为 nil
		if item.IsSharedUnit != nil {
			itemMap["is_shared_unit"] = *item.IsSharedUnit
			// 保持向后兼容：同时设置 is_multi_person_room
			itemMap["is_multi_person_room"] = *item.IsSharedUnit
		}
		if item.RoomID != nil {
			itemMap["room_id"] = *item.RoomID
		}
		if item.RoomName != nil {
			itemMap["room_name"] = *item.RoomName
		}
		if item.BedID != nil {
			itemMap["bed_id"] = *item.BedID
		}
		if item.BedName != nil {
			itemMap["bed_name"] = *item.BedName
		}
		items = append(items, itemMap)
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"items": items,
		"total": resp.Total,
	}))
}

// ============================================
// Handler 层请求/响应结构定义
// ============================================

// CreateResidentRequest Handler 层创建住户请求结构
// 按照 Service 层的三个结构体组织：InherentAttributes + UnitRelation + CaregiverRelation
type CreateResidentRequest struct {
	// Resident 固有属性（3张表：residents, resident_phi, resident_contacts）
	InherentAttributes *CreateResidentInherentAttributesRequest `json:"inherent_attributes"`

	// Resident 与 Unit 的关系（位置分配）
	UnitRelation *CreateResidentUnitRelationRequest `json:"unit_relation,omitempty"`

	// Resident 与 Caregiver 的关系（护理人员分配）
	CaregiverRelation *CreateResidentCaregiverRelationRequest `json:"caregiver_relation,omitempty"`
}

// CreateResidentInherentAttributesRequest Handler 层 Resident 固有属性请求结构
type CreateResidentInherentAttributesRequest struct {
	// ========== residents 表字段 ==========
	// 必填字段
	ResidentAccount string `json:"resident_account"` // 住户账号（必填）
	Nickname        string `json:"nickname"`         // 昵称（必填）

	// 可选字段
	PasswordHash    string          `json:"password_hash"`     // password_hash hex 字符串（可选，前端已计算）
	Status          string          `json:"status"`            // 状态（可选，默认 "active"）
	ServiceLevel    string          `json:"service_level"`     // 护理级别（可选）
	AdmissionDate   string          `json:"admission_date"`    // 入院日期 YYYY-MM-DD 格式（可选，默认当前日期）
	DischargeDate   string          `json:"discharge_date"`    // 出院日期 YYYY-MM-DD 格式（可选，仅在 status='discharged' 或 'transferred' 时有效）
	BranchID        string          `json:"branch_id"`         // 院区ID（可选）
	IsAccessEnabled bool            `json:"is_access_enabled"` // 是否允许查看状态（可选，默认 false）
	Note            string          `json:"note"`              // 备注（可选）
	PhoneHash       string          `json:"phone_hash"`        // phone_hash hex 字符串（可选，前端已计算）
	EmailHash       string          `json:"email_hash"`        // email_hash hex 字符串（可选，前端已计算）
	Metadata        json.RawMessage `json:"metadata"`          // metadata JSONB（可选）

	// ========== resident_phi 表字段 ==========
	PHI *PHIRequest `json:"phi,omitempty"`

	// ========== resident_contacts 表字段 ==========
	Contacts []*ContactRequest `json:"contacts,omitempty"`
}

// PHIRequest Handler 层 PHI 请求结构
type PHIRequest struct {
	FirstName             string   `json:"first_name"`
	LastName              string   `json:"last_name"`
	Gender                string   `json:"gender"`
	DateOfBirth           string   `json:"date_of_birth"` // YYYY-MM-DD 格式
	ResidentPhone         string   `json:"resident_phone"`
	ResidentEmail         string   `json:"resident_email"`
	SavePhone             bool     `json:"save_phone"`
	SaveEmail             bool     `json:"save_email"`
	WeightLb              *float64 `json:"weight_lb"`
	HeightFt              *float64 `json:"height_ft"`
	HeightIn              *float64 `json:"height_in"`
	MobilityLevel         *int     `json:"mobility_level"`
	TremorStatus          string   `json:"tremor_status"`
	MobilityAid           string   `json:"mobility_aid"`
	ADLAssistance         string   `json:"adl_assistance"`
	CommStatus            string   `json:"comm_status"`
	HasHypertension       *bool    `json:"has_hypertension"`
	HasHyperlipaemia      *bool    `json:"has_hyperlipaemia"`
	HasHyperglycaemia     *bool    `json:"has_hyperglycaemia"`
	HasStrokeHistory      *bool    `json:"has_stroke_history"`
	HasParalysis          *bool    `json:"has_paralysis"`
	HasAlzheimer          *bool    `json:"has_alzheimer"`
	MedicalHistory        string   `json:"medical_history"`
	HomeAddressStreet     string   `json:"home_address_street"`
	HomeAddressCity       string   `json:"home_address_city"`
	HomeAddressState      string   `json:"home_address_state"`
	HomeAddressPostalCode string   `json:"home_address_postal_code"`
	PlusCode              string   `json:"plus_code"`
}

// ContactRequest Handler 层联系人请求结构
// 对应 resident_contacts 表，注意：联系人不登录系统，仅作为住户的属性
type ContactRequest struct {
	Slot             string          `json:"slot"`               // 槽位 'A', 'B', 'C', 'D', 'E'（必填）
	IsEnabled        bool            `json:"is_enabled"`         // 是否启用该联系人（对应前端的 "slot Enable" 复选框）
	Relationship     string          `json:"relationship"`       // 关系（可选）：Child/Spouse/Friend/Caregiver/Other
	ContactFirstName string          `json:"contact_first_name"` // 联系人名（可选）
	ContactLastName  string          `json:"contact_last_name"`  // 联系人姓（可选）
	ContactPhone     string          `json:"contact_phone"`      // 联系人电话（可选），明文保存
	ContactEmail     string          `json:"contact_email"`      // 联系人邮箱（可选），明文保存
	ContactPhoneHash string          `json:"contact_phone_hash"` // 联系人电话 hash（可选，前端计算的 hex 字符串，用于搜索）
	ContactEmailHash string          `json:"contact_email_hash"` // 联系人邮箱 hash（可选，前端计算的 hex 字符串，用于搜索）
	ReceiveSMS       bool            `json:"receive_sms"`        // 是否接收短信（可选，默认 false）
	ReceiveEmail     bool            `json:"receive_email"`      // 是否接收邮件（可选，默认 false）
	AlertTimeWindow  json.RawMessage `json:"alert_time_window"`  // 告警接收时间窗口 JSONB（可选）
}

// CreateResidentUnitRelationRequest Handler 层 Resident 与 Unit 的关系请求结构
type CreateResidentUnitRelationRequest struct {
	UnitID string `json:"unit_id"` // 单元ID（可选）
	RoomID string `json:"room_id"` // 房间ID（可选）
	BedID  string `json:"bed_id"`  // 床位ID（可选）
	// 业务规则：bed → room → unit（如果指定 bed_id，则必须同时指定 room_id 和 unit_id）
}

// CreateResidentCaregiverRelationRequest Handler 层 Resident 与 Caregiver 的关系请求结构
type CreateResidentCaregiverRelationRequest struct {
	UserList  []string `json:"user_list"`  // 用户ID列表（可选，JSONB array）
	GroupList []string `json:"group_list"` // 用户组标签列表（可选，JSONB array，用于匹配 users.user_tags）
	// 说明：
	//   - 每个租户+住户最多一条记录（UNIQUE(tenant_id, resident_id)）
	//   - 如果 user_list 和 group_list 都为空，使用默认告警路由规则（由应用层处理）
}

// ============================================
// GetResident 获取住户详情
// ============================================

func (h *ResidentHandler) GetResident(w http.ResponseWriter, r *http.Request, residentID string) {
	ctx := r.Context()

	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	currentUserID := r.Header.Get("X-User-Id")
	currentUserType := r.Header.Get("X-User-Type")
	currentUserRole := r.Header.Get("X-User-Role")

	// 获取查询参数
	includePHI := r.URL.Query().Get("include_phi") == "true"
	includeContacts := r.URL.Query().Get("include_contacts") == "true"

	// 权限检查（仅 Staff 需要）
	// 注意：resident_contacts 不能登录系统，所以 currentUserType 永远不会是 "family"
	// currentUserRole 也不会是 "Family"，因为只有 residents 可以登录
	// 保留这些检查是为了向后兼容，但实际上只会是 "resident" 或 "staff"
	// Service 层会自己查询用户的 branch_id（通过 user_branches 表），这里不需要传递 UserBranchTag
	var permCheck *service.PermissionCheckResult
	isResident := currentUserType == "resident" || currentUserRole == "Resident"
	if !isResident && currentUserRole != "" && h.db != nil {
		perm, err := GetResourcePermission(h.db, ctx, currentUserRole, "residents", "R")
		if err == nil {
			permCheck = &service.PermissionCheckResult{
				AssignedOnly: perm.AssignedOnly,
				BranchOnly:   perm.BranchOnly,
			}
		}
	}

	req := service.GetResidentRequest{
		TenantID:        tenantID,
		ResidentID:      residentID,
		CurrentUserID:   currentUserID,
		CurrentUserType: currentUserType,
		CurrentUserRole: currentUserRole,
		PermissionCheck: permCheck,
		IncludePHI:      includePHI,
		IncludeContacts: includeContacts,
	}

	resp, err := h.residentService.GetResident(ctx, req)
	if err != nil {
		h.logger.Error("GetResident failed",
			zap.String("tenant_id", tenantID),
			zap.String("resident_id", residentID),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 转换为旧 Handler 格式
	item := map[string]any{
		"resident_id":       resp.Resident.ResidentID,
		"tenant_id":         resp.Resident.TenantID,
		"nickname":          resp.Resident.Nickname,
		"status":            resp.Resident.Status,
		"is_access_enabled": resp.Resident.IsAccessEnabled,
	}
	if resp.Resident.ResidentAccount != nil {
		item["resident_account"] = *resp.Resident.ResidentAccount
	}
	if resp.Resident.ServiceLevel != nil {
		item["service_level"] = *resp.Resident.ServiceLevel
	}
	if resp.Resident.AdmissionDate != nil {
		item["admission_date"] = time.Unix(*resp.Resident.AdmissionDate, 0).Format("2006-01-02")
	}
	if resp.Resident.DischargeDate != nil {
		item["discharge_date"] = time.Unix(*resp.Resident.DischargeDate, 0).Format("2006-01-02")
	}
	if resp.Resident.UnitID != nil {
		item["unit_id"] = *resp.Resident.UnitID
	}
	if resp.Resident.UnitName != nil {
		item["unit_name"] = *resp.Resident.UnitName
	}
	if resp.Resident.BranchID != nil {
		item["branch_id"] = *resp.Resident.BranchID
	}
	if resp.Resident.BranchName != nil {
		item["branch_name"] = *resp.Resident.BranchName
		// 保持向后兼容：同时设置 branch_tag
		item["branch_tag"] = *resp.Resident.BranchName
	}
	if resp.Resident.BuildingName != nil {
		item["building_name"] = *resp.Resident.BuildingName
		// 保持向后兼容：同时设置 building
		item["building"] = *resp.Resident.BuildingName
	}
	// is_shared_unit 现在为 *bool，未绑定 unit 时为 nil
	if resp.Resident.IsSharedUnit != nil {
		item["is_shared_unit"] = *resp.Resident.IsSharedUnit
		// 保持向后兼容：同时设置 is_multi_person_room
		item["is_multi_person_room"] = *resp.Resident.IsSharedUnit
	}
	if resp.Resident.RoomID != nil {
		item["room_id"] = *resp.Resident.RoomID
	}
	if resp.Resident.RoomName != nil {
		item["room_name"] = *resp.Resident.RoomName
	}
	if resp.Resident.BedID != nil {
		item["bed_id"] = *resp.Resident.BedID
	}
	if resp.Resident.BedName != nil {
		item["bed_name"] = *resp.Resident.BedName
	}
	if resp.Resident.Note != nil {
		item["note"] = *resp.Resident.Note
	}

	// 添加 email 和 phone（从 PHI 中获取，用于前端显示和创建时的 hash 计算）
	// 注意：这些字段不在 residents 表中，但在前端 Resident 模型中定义
	if resp.PHI != nil {
		if resp.PHI.ResidentEmail != nil {
			item["email"] = *resp.PHI.ResidentEmail
		} else {
			// 检查 email_hash 是否存在（如果存在但 email 为 NULL，返回占位符）
			var emailHash []byte
			err := h.db.QueryRowContext(ctx,
				`SELECT email_hash FROM residents WHERE tenant_id = $1 AND resident_id::text = $2`,
				tenantID, residentID,
			).Scan(&emailHash)
			if err == nil && len(emailHash) > 0 {
				item["email"] = "***@***" // Placeholder when hash exists but email is not saved
			}
		}
		if resp.PHI.ResidentPhone != nil {
			item["phone"] = *resp.PHI.ResidentPhone
		} else {
			// 检查 phone_hash 是否存在（如果存在但 phone 为 NULL，返回占位符）
			var phoneHash []byte
			err := h.db.QueryRowContext(ctx,
				`SELECT phone_hash FROM residents WHERE tenant_id = $1 AND resident_id::text = $2`,
				tenantID, residentID,
			).Scan(&phoneHash)
			if err == nil && len(phoneHash) > 0 {
				item["phone"] = "xxx-xxx-xxxx" // Placeholder when hash exists but phone is not saved
			}
		}
	} else {
		// 如果没有 PHI 数据，检查 hash 是否存在
		var phoneHash, emailHash []byte
		err := h.db.QueryRowContext(ctx,
			`SELECT phone_hash, email_hash FROM residents WHERE tenant_id = $1 AND resident_id::text = $2`,
			tenantID, residentID,
		).Scan(&phoneHash, &emailHash)
		if err == nil {
			if len(phoneHash) > 0 {
				item["phone"] = "xxx-xxx-xxxx" // Placeholder when hash exists but phone is not saved
			}
			if len(emailHash) > 0 {
				item["email"] = "***@***" // Placeholder when hash exists but email is not saved
			}
		}
	}

	// 添加 PHI 数据
	if resp.PHI != nil {
		phi := map[string]any{
			"phi_id":      resp.PHI.PhiID,
			"resident_id": residentID, // PHI DTO 中没有 ResidentID 字段，使用传入的 residentID
		}
		if resp.PHI.FirstName != nil {
			phi["first_name"] = *resp.PHI.FirstName
		}
		if resp.PHI.LastName != nil {
			phi["last_name"] = *resp.PHI.LastName
		}
		if resp.PHI.Gender != nil {
			phi["gender"] = *resp.PHI.Gender
		}
		if resp.PHI.DateOfBirth != nil {
			phi["date_of_birth"] = time.Unix(*resp.PHI.DateOfBirth, 0).Format("2006-01-02")
		}
		// 处理 phone/email（需要检查 hash 是否存在）
		// 注意：Service 层返回的 PHI 中，如果 phone_hash 存在但 phone 为 NULL，应该返回占位符
		// 但当前 Service 层实现中，如果 phone 为 NULL，则不会在 DTO 中设置
		// 这里需要从数据库查询 phone_hash 和 email_hash 来判断
		if resp.PHI.ResidentPhone != nil {
			phi["resident_phone"] = *resp.PHI.ResidentPhone
		} else {
			// 检查 phone_hash 是否存在（如果存在但 phone 为 NULL，返回占位符）
			var phoneHash []byte
			err := h.db.QueryRowContext(ctx,
				`SELECT phone_hash FROM residents WHERE tenant_id = $1 AND resident_id::text = $2`,
				tenantID, residentID,
			).Scan(&phoneHash)
			if err == nil && phoneHash != nil && len(phoneHash) > 0 {
				phi["resident_phone"] = "xxx-xxx-xxxx" // Placeholder when hash exists but phone is not saved
			}
		}
		if resp.PHI.ResidentEmail != nil {
			phi["resident_email"] = *resp.PHI.ResidentEmail
		} else {
			// 检查 email_hash 是否存在
			var emailHash []byte
			err := h.db.QueryRowContext(ctx,
				`SELECT email_hash FROM residents WHERE tenant_id = $1 AND resident_id::text = $2`,
				tenantID, residentID,
			).Scan(&emailHash)
			if err == nil && emailHash != nil && len(emailHash) > 0 {
				phi["resident_email"] = "***@***" // Placeholder when hash exists but email is not saved
			}
		}
		// 其他 PHI 字段
		if resp.PHI.WeightLb != nil {
			phi["weight_lb"] = *resp.PHI.WeightLb
		}
		if resp.PHI.HeightFt != nil {
			phi["height_ft"] = *resp.PHI.HeightFt
		}
		if resp.PHI.HeightIn != nil {
			phi["height_in"] = *resp.PHI.HeightIn
		}
		if resp.PHI.MobilityLevel != nil {
			phi["mobility_level"] = *resp.PHI.MobilityLevel
		}
		if resp.PHI.TremorStatus != nil {
			phi["tremor_status"] = *resp.PHI.TremorStatus
		}
		if resp.PHI.MobilityAid != nil {
			phi["mobility_aid"] = *resp.PHI.MobilityAid
		}
		if resp.PHI.ADLAssistance != nil {
			phi["adl_assistance"] = *resp.PHI.ADLAssistance
		}
		if resp.PHI.CommStatus != nil {
			phi["comm_status"] = *resp.PHI.CommStatus
		}
		if resp.PHI.HasHypertension != nil {
			phi["has_hypertension"] = *resp.PHI.HasHypertension
		}
		if resp.PHI.HasHyperlipaemia != nil {
			phi["has_hyperlipaemia"] = *resp.PHI.HasHyperlipaemia
		}
		if resp.PHI.HasHyperglycaemia != nil {
			phi["has_hyperglycaemia"] = *resp.PHI.HasHyperglycaemia
		}
		if resp.PHI.HasStrokeHistory != nil {
			phi["has_stroke_history"] = *resp.PHI.HasStrokeHistory
		}
		if resp.PHI.HasParalysis != nil {
			phi["has_paralysis"] = *resp.PHI.HasParalysis
		}
		if resp.PHI.HasAlzheimer != nil {
			phi["has_alzheimer"] = *resp.PHI.HasAlzheimer
		}
		if resp.PHI.MedicalHistory != nil {
			phi["medical_history"] = *resp.PHI.MedicalHistory
		}
		if resp.PHI.HomeAddressStreet != nil {
			phi["home_address_street"] = *resp.PHI.HomeAddressStreet
		}
		if resp.PHI.HomeAddressCity != nil {
			phi["home_address_city"] = *resp.PHI.HomeAddressCity
		}
		if resp.PHI.HomeAddressState != nil {
			phi["home_address_state"] = *resp.PHI.HomeAddressState
		}
		if resp.PHI.HomeAddressPostalCode != nil {
			phi["home_address_postal_code"] = *resp.PHI.HomeAddressPostalCode
		}
		if resp.PHI.PlusCode != nil {
			phi["plus_code"] = *resp.PHI.PlusCode
		}
		item["phi"] = phi
	}

	// 添加联系人数据
	if len(resp.Contacts) > 0 {
		contacts := make([]any, 0, len(resp.Contacts))
		for _, c := range resp.Contacts {
			contact := map[string]any{
				"contact_id":    c.ContactID,
				"resident_id":   residentID, // 添加 resident_id
				"slot":          c.Slot,
				"is_enabled":    c.IsEnabled,
				"receive_sms":   c.ReceiveSMS,
				"receive_email": c.ReceiveEmail,
			}
			if c.Relationship != nil {
				contact["relationship"] = *c.Relationship
			}
			if c.ContactFirstName != nil {
				contact["contact_first_name"] = *c.ContactFirstName
			}
			if c.ContactLastName != nil {
				contact["contact_last_name"] = *c.ContactLastName
			}
			// 处理 phone/email（需要检查 hash 是否存在）
			if c.ContactPhone != nil {
				contact["contact_phone"] = *c.ContactPhone
			} else {
				// 检查 phone_hash 是否存在
				var phoneHash []byte
				err := h.db.QueryRowContext(ctx,
					`SELECT phone_hash FROM resident_contacts WHERE tenant_id = $1 AND contact_id::text = $2`,
					tenantID, c.ContactID,
				).Scan(&phoneHash)
				if err == nil && phoneHash != nil && len(phoneHash) > 0 {
					contact["contact_phone"] = "xxx-xxx-xxxx" // Placeholder when hash exists but phone is not saved
				}
			}
			if c.ContactEmail != nil {
				contact["contact_email"] = *c.ContactEmail
			} else {
				// 检查 email_hash 是否存在
				var emailHash []byte
				err := h.db.QueryRowContext(ctx,
					`SELECT email_hash FROM resident_contacts WHERE tenant_id = $1 AND contact_id::text = $2`,
					tenantID, c.ContactID,
				).Scan(&emailHash)
				if err == nil && emailHash != nil && len(emailHash) > 0 {
					contact["contact_email"] = "***@***" // Placeholder when hash exists but email is not saved
				}
			}
			if c.ContactFamilyTag != nil {
				contact["contact_family_tag"] = *c.ContactFamilyTag
			}
			contacts = append(contacts, contact)
		}
		item["contacts"] = contacts
	}

	// 添加 caregivers 数据（默认包含，与旧 Handler 一致）
	{
		var userListRaw, groupListRaw []byte
		err := h.db.QueryRowContext(ctx,
			`SELECT user_list, group_list
			 FROM resident_caregivers
			 WHERE tenant_id = $1 AND resident_id::text = $2`,
			tenantID, residentID,
		).Scan(&userListRaw, &groupListRaw)
		if err == nil {
			var userList []string
			var groupList []string
			if len(userListRaw) > 0 {
				if err := json.Unmarshal(userListRaw, &userList); err == nil {
					// userList parsed successfully
				}
			}
			if len(groupListRaw) > 0 {
				if err := json.Unmarshal(groupListRaw, &groupList); err == nil {
					// groupList parsed successfully
				}
			}
			item["caregivers"] = map[string]any{
				"userList":  userList,
				"groupList": groupList,
			}
		}
	}

	writeJSON(w, http.StatusOK, Ok(item))
}

// ============================================
// CreateResident 创建住户
// ============================================

func (h *ResidentHandler) CreateResident(w http.ResponseWriter, r *http.Request) {
	// 原有实现已注释，使用新的实现（见下方）
	/*
	ctx := r.Context()

	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	currentUserID := r.Header.Get("X-User-Id")
	currentUserType := r.Header.Get("X-User-Type")
	currentUserRole := r.Header.Get("X-User-Role")

	// 权限检查（需要 C 权限）
	// 注意：resident_contacts 不能登录系统，所以 currentUserType 永远不会是 "family"
	// 保留此检查是为了向后兼容，但实际上只会是 "resident" 或 "staff"
	if currentUserType != "resident" && currentUserType != "family" && currentUserRole != "" && h.db != nil {
		hasCPermission := false
		err := h.db.QueryRowContext(ctx,
			`SELECT EXISTS(
				SELECT 1 FROM role_permissions
				WHERE tenant_id = $1 AND role_code = $2 AND resource_type = 'residents' AND permission_type = 'C'
			)`,
			SystemTenantID(), currentUserRole,
		).Scan(&hasCPermission)
		if err != nil || !hasCPermission {
			writeJSON(w, http.StatusOK, Fail("permission denied: no create permission for residents"))
			return
		}
	}

	// 解析请求体
	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	// 权限检查结果
	// Service 层会自己查询用户的 branch_id（通过 user_branches 表），这里不需要传递 UserBranchTag
	var permCheck *service.PermissionCheckResult
	if currentUserRole != "" && h.db != nil {
		perm, err := GetResourcePermission(h.db, ctx, currentUserRole, "residents", "C")
		if err == nil {
			permCheck = &service.PermissionCheckResult{
				AssignedOnly: perm.AssignedOnly,
				BranchOnly:   perm.BranchOnly,
			}
		}
	}

	// 构建请求
	req := service.CreateResidentRequest{
		TenantID:        tenantID,
		CurrentUserID:   currentUserID,
		CurrentUserRole: currentUserRole,
		PermissionCheck: permCheck,
	}

	// 提取必填字段
	if residentAccount, ok := payload["resident_account"].(string); ok {
		req.ResidentAccount = strings.TrimSpace(residentAccount)
	}
	if nickname, ok := payload["nickname"].(string); ok {
		req.Nickname = strings.TrimSpace(nickname)
	}
	if password, ok := payload["password"].(string); ok {
		req.Password = password
	}
	if status, ok := payload["status"].(string); ok {
		req.Status = status
	}
	if serviceLevel, ok := payload["service_level"].(string); ok {
		req.ServiceLevel = serviceLevel
	}
	if unitID, ok := payload["unit_id"].(string); ok {
		req.UnitID = unitID
	}
	if familyTag, ok := payload["family_tag"].(string); ok {
		req.FamilyTag = familyTag
	}
	if isAccessEnabled, ok := payload["is_access_enabled"].(bool); ok {
		req.IsAccessEnabled = isAccessEnabled
	}
	if note, ok := payload["note"].(string); ok {
		req.Note = note
	}
	if phoneHash, ok := payload["phone_hash"].(string); ok {
		req.PhoneHash = phoneHash
	}
	if emailHash, ok := payload["email_hash"].(string); ok {
		req.EmailHash = emailHash
	}

	// 处理 admission_date
	if admDate, ok := payload["admission_date"].(string); ok && admDate != "" {
		if t, err := time.Parse("2006-01-02", admDate); err == nil {
			ts := t.Unix()
			req.AdmissionDate = &ts
		}
	}

	// 处理 PHI 数据
	if phiData, ok := payload["phi"].(map[string]any); ok {
		phi := &service.CreateResidentPHIRequest{}
		if firstName, ok := phiData["first_name"].(string); ok {
			phi.FirstName = firstName
		}
		if lastName, ok := phiData["last_name"].(string); ok {
			phi.LastName = lastName
		}
		if gender, ok := phiData["gender"].(string); ok {
			phi.Gender = gender
		}
		if dob, ok := phiData["date_of_birth"].(string); ok && dob != "" {
			if t, err := time.Parse("2006-01-02", dob); err == nil {
				ts := t.Unix()
				phi.DateOfBirth = &ts
			}
		}
		if residentPhone, ok := phiData["resident_phone"].(string); ok {
			phi.ResidentPhone = residentPhone
		}
		if residentEmail, ok := phiData["resident_email"].(string); ok {
			phi.ResidentEmail = residentEmail
		}
		if savePhone, ok := phiData["save_phone"].(bool); ok {
			phi.SavePhone = savePhone
		}
		if saveEmail, ok := phiData["save_email"].(bool); ok {
			phi.SaveEmail = saveEmail
		}
		// ... 其他 PHI 字段
		req.PHI = phi
	}

	// 处理联系人数据
	if contacts, ok := payload["contacts"].([]any); ok {
		req.Contacts = make([]*service.CreateResidentContactRequest, 0, len(contacts))
		for _, contactRaw := range contacts {
			if contact, ok := contactRaw.(map[string]any); ok {
				contactReq := &service.CreateResidentContactRequest{}
				if slot, ok := contact["slot"].(string); ok {
					contactReq.Slot = slot
				}
				if isEnabled, ok := contact["is_enabled"].(bool); ok {
					contactReq.IsEnabled = isEnabled
				}
				if relationship, ok := contact["relationship"].(string); ok {
					contactReq.Relationship = relationship
				}
				if contactFirstName, ok := contact["contact_first_name"].(string); ok {
					contactReq.ContactFirstName = contactFirstName
				}
				if contactLastName, ok := contact["contact_last_name"].(string); ok {
					contactReq.ContactLastName = contactLastName
				}
				if contactPhone, ok := contact["contact_phone"].(string); ok {
					contactReq.ContactPhone = contactPhone
				}
				if contactEmail, ok := contact["contact_email"].(string); ok {
					contactReq.ContactEmail = contactEmail
				}
				if receiveSMS, ok := contact["receive_sms"].(bool); ok {
					contactReq.ReceiveSMS = receiveSMS
				}
				if receiveEmail, ok := contact["receive_email"].(bool); ok {
					contactReq.ReceiveEmail = receiveEmail
				}
				if phoneHash, ok := contact["phone_hash"].(string); ok {
					contactReq.PhoneHash = phoneHash
				}
				if emailHash, ok := contact["email_hash"].(string); ok {
					contactReq.EmailHash = emailHash
				}
				if contactFamilyTag, ok := contact["contact_family_tag"].(string); ok {
					contactReq.ContactFamilyTag = contactFamilyTag
				}
				req.Contacts = append(req.Contacts, contactReq)
			}
		}
	}

	resp, err := h.residentService.CreateResident(ctx, req)
				if err != nil {
					h.logger.Error("CreateResident failed",
						zap.String("tenant_id", tenantID),
						zap.Error(err),
					)
					writeJSON(w, http.StatusOK, Fail(err.Error()))
					return
				}

				writeJSON(w, http.StatusOK, Ok(map[string]any{
					"resident_id": resp.ResidentID,
				}))
	*/

	// ========== 新实现开始 ==========
	ctx := r.Context()

	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	currentUserID := r.Header.Get("X-User-Id")
	currentUserType := r.Header.Get("X-User-Type")
	currentUserRole := r.Header.Get("X-User-Role")

	// 权限检查（需要 C 权限）
	// 注意：resident_contacts 不能登录系统，所以 currentUserType 永远不会是 "family"
	// 保留此检查是为了向后兼容，但实际上只会是 "resident" 或 "staff"
	if currentUserType != "resident" && currentUserType != "family" && currentUserRole != "" && h.db != nil {
		hasCPermission := false
		err := h.db.QueryRowContext(ctx,
			`SELECT EXISTS(
				SELECT 1 FROM role_permissions
				WHERE tenant_id = $1 AND role_code = $2 AND resource_type = 'residents' AND permission_type = 'C'
			)`,
			SystemTenantID(), currentUserRole,
		).Scan(&hasCPermission)
		if err != nil || !hasCPermission {
			writeJSON(w, http.StatusOK, Fail("permission denied: no create permission for residents"))
			return
		}
	}

	// 解析请求体到 Handler 层结构
	var handlerReq CreateResidentRequest
	if err := json.NewDecoder(r.Body).Decode(&handlerReq); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	// 权限检查结果
	// Service 层会自己查询用户的 branch_id（通过 user_branches 表），这里不需要传递 UserBranchTag
	var permCheck *service.PermissionCheckResult
	if currentUserRole != "" && h.db != nil {
		perm, err := GetResourcePermission(h.db, ctx, currentUserRole, "residents", "C")
		if err == nil {
			permCheck = &service.PermissionCheckResult{
				AssignedOnly: perm.AssignedOnly,
				BranchOnly:   perm.BranchOnly,
			}
		}
	}

	// 注意：AvailableBranches 不应传递给 Service 层
	// Service 层会自己从数据库查询用户的 branch 信息（这是用户本身的属性，不能信任前端传递的值）
	// 如果前端需要获取可用 branch 列表，应该调用专门的 API（如 GetAvailableBranches）

	// 转换为 Service 层请求 - 按照三个结构体组织：InherentAttributes + UnitRelation + CaregiverRelation
	serviceReq := service.CreateResidentRequest{
		TenantID:        tenantID,
		CurrentUserID:   currentUserID,
		CurrentUserRole: currentUserRole,
		PermissionCheck: permCheck,
	}

	// 1. 构建 InherentAttributes（固有属性：residents + resident_phi + resident_contacts）
	if handlerReq.InherentAttributes != nil {
		inherentAttrs := &service.CreateResidentInherentAttributes{
			ResidentAccount: strings.TrimSpace(handlerReq.InherentAttributes.ResidentAccount),
			Nickname:        strings.TrimSpace(handlerReq.InherentAttributes.Nickname),
			PasswordHash:    handlerReq.InherentAttributes.PasswordHash,
			Status:          handlerReq.InherentAttributes.Status,
			ServiceLevel:    handlerReq.InherentAttributes.ServiceLevel,
			BranchID:        handlerReq.InherentAttributes.BranchID,
			IsAccessEnabled: handlerReq.InherentAttributes.IsAccessEnabled,
			Note:            handlerReq.InherentAttributes.Note,
			PhoneHash:       handlerReq.InherentAttributes.PhoneHash,
			EmailHash:       handlerReq.InherentAttributes.EmailHash,
			Metadata:        handlerReq.InherentAttributes.Metadata,
		}

		// 处理 admission_date
		if handlerReq.InherentAttributes.AdmissionDate != "" {
			if t, err := time.Parse("2006-01-02", handlerReq.InherentAttributes.AdmissionDate); err == nil {
				ts := t.Unix()
				inherentAttrs.AdmissionDate = &ts
			}
		}

		// 处理 discharge_date（创建时通常为空，但如果提供了则设置）
		if handlerReq.InherentAttributes.DischargeDate != "" {
			if t, err := time.Parse("2006-01-02", handlerReq.InherentAttributes.DischargeDate); err == nil {
				ts := t.Unix()
				inherentAttrs.DischargeDate = &ts
			}
		}

		// 处理 PHI 数据
		if handlerReq.InherentAttributes.PHI != nil {
			phi := &service.CreateResidentPHIRequest{
				FirstName:             handlerReq.InherentAttributes.PHI.FirstName,
				LastName:              handlerReq.InherentAttributes.PHI.LastName,
				Gender:                handlerReq.InherentAttributes.PHI.Gender,
				ResidentPhone:         handlerReq.InherentAttributes.PHI.ResidentPhone,
				ResidentEmail:         handlerReq.InherentAttributes.PHI.ResidentEmail,
				SavePhone:             handlerReq.InherentAttributes.PHI.SavePhone,
				SaveEmail:             handlerReq.InherentAttributes.PHI.SaveEmail,
				WeightLb:              handlerReq.InherentAttributes.PHI.WeightLb,
				HeightFt:              handlerReq.InherentAttributes.PHI.HeightFt,
				HeightIn:              handlerReq.InherentAttributes.PHI.HeightIn,
				MobilityLevel:         handlerReq.InherentAttributes.PHI.MobilityLevel,
				TremorStatus:          handlerReq.InherentAttributes.PHI.TremorStatus,
				MobilityAid:           handlerReq.InherentAttributes.PHI.MobilityAid,
				ADLAssistance:         handlerReq.InherentAttributes.PHI.ADLAssistance,
				CommStatus:            handlerReq.InherentAttributes.PHI.CommStatus,
				HasHypertension:       handlerReq.InherentAttributes.PHI.HasHypertension,
				HasHyperlipaemia:      handlerReq.InherentAttributes.PHI.HasHyperlipaemia,
				HasHyperglycaemia:     handlerReq.InherentAttributes.PHI.HasHyperglycaemia,
				HasStrokeHistory:      handlerReq.InherentAttributes.PHI.HasStrokeHistory,
				HasParalysis:          handlerReq.InherentAttributes.PHI.HasParalysis,
				HasAlzheimer:          handlerReq.InherentAttributes.PHI.HasAlzheimer,
				MedicalHistory:        handlerReq.InherentAttributes.PHI.MedicalHistory,
				HomeAddressStreet:     handlerReq.InherentAttributes.PHI.HomeAddressStreet,
				HomeAddressCity:       handlerReq.InherentAttributes.PHI.HomeAddressCity,
				HomeAddressState:      handlerReq.InherentAttributes.PHI.HomeAddressState,
				HomeAddressPostalCode: handlerReq.InherentAttributes.PHI.HomeAddressPostalCode,
				PlusCode:              handlerReq.InherentAttributes.PHI.PlusCode,
			}
			if handlerReq.InherentAttributes.PHI.DateOfBirth != "" {
				if t, err := time.Parse("2006-01-02", handlerReq.InherentAttributes.PHI.DateOfBirth); err == nil {
					ts := t.Unix()
					phi.DateOfBirth = &ts
				}
			}
			inherentAttrs.PHI = phi
		}

		// 处理联系人数据（注意：联系人不登录系统，但需要保存 phone_hash 和 email_hash 用于搜索）
		if len(handlerReq.InherentAttributes.Contacts) > 0 {
			inherentAttrs.Contacts = make([]*service.CreateResidentContactRequest, 0, len(handlerReq.InherentAttributes.Contacts))
			for _, handlerContact := range handlerReq.InherentAttributes.Contacts {
				contactReq := &service.CreateResidentContactRequest{
					Slot:             handlerContact.Slot,
					IsEnabled:        handlerContact.IsEnabled,
					Relationship:     handlerContact.Relationship,
					ContactFirstName: handlerContact.ContactFirstName,
					ContactLastName:  handlerContact.ContactLastName,
					ContactPhone:     handlerContact.ContactPhone,
					ContactEmail:     handlerContact.ContactEmail,
					ContactPhoneHash: handlerContact.ContactPhoneHash,
					ContactEmailHash: handlerContact.ContactEmailHash,
					ReceiveSMS:       handlerContact.ReceiveSMS,
					ReceiveEmail:     handlerContact.ReceiveEmail,
					AlertTimeWindow:  handlerContact.AlertTimeWindow,
				}
				inherentAttrs.Contacts = append(inherentAttrs.Contacts, contactReq)
			}
		}

		serviceReq.InherentAttributes = inherentAttrs
	}

	// 2. 构建 UnitRelation（位置分配）
	if handlerReq.UnitRelation != nil {
		serviceReq.UnitRelation = &service.CreateResidentUnitRelation{
			UnitID: handlerReq.UnitRelation.UnitID,
			RoomID: handlerReq.UnitRelation.RoomID,
			BedID:  handlerReq.UnitRelation.BedID,
		}
	}

	// 3. 构建 CaregiverRelation（护理人员分配）
	if handlerReq.CaregiverRelation != nil {
		serviceReq.CaregiverRelation = &service.CreateResidentCaregiverRelation{
			UserList:  handlerReq.CaregiverRelation.UserList,
			GroupList: handlerReq.CaregiverRelation.GroupList,
		}
	}

	resp, err := h.residentService.CreateResident(ctx, serviceReq)
	if err != nil {
		h.logger.Error("CreateResident failed",
			zap.String("tenant_id", tenantID),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"resident_id": resp.ResidentID,
	}))
}

// ============================================
// UpdateResident 更新住户
// ============================================
// 前端数据格式转换规则：
// - 字段不存在（undefined）→ nil（不更新）
// - 字段存在且有值 → UpdateActionUpdate（更新为新值）
// - 字段存在但值为 null → UpdateActionDelete（删除/设置为 NULL）

func (h *ResidentHandler) UpdateResident(w http.ResponseWriter, r *http.Request, residentID string) {
	ctx := r.Context()

	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	currentUserID := r.Header.Get("X-User-Id")
	currentUserRole := r.Header.Get("X-User-Role")

	// 解析请求体
	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	// 构建 Service 请求（权限检查由 Service 层自己处理）
	req := service.UpdateResidentRequest{
		TenantID:        tenantID,
		ResidentID:      residentID,
		CurrentUserID:   currentUserID,
		CurrentUserRole: currentUserRole,
		PermissionCheck: nil, // Service 层自己查询权限
	}

	// 构建 InherentAttributes（residents 表字段 + PHI + Contacts）
	inherentAttrs := &service.UpdateResidentInherentAttributes{}

	// 1. 处理 residents 表字段
	if val, exists := payload["resident_account"]; exists {
		if val == nil {
			inherentAttrs.ResidentAccount = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
		} else if str, ok := val.(string); ok && str != "" {
			inherentAttrs.ResidentAccount = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
		}
	}
	if val, exists := payload["nickname"]; exists {
		if val == nil {
			inherentAttrs.Nickname = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
		} else if str, ok := val.(string); ok {
			inherentAttrs.Nickname = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
		}
	}
	if val, exists := payload["password_hash"]; exists {
		if val == nil {
			inherentAttrs.PasswordHash = &domain.UpdateBytes{Action: domain.UpdateActionDelete, Value: nil}
		} else if str, ok := val.(string); ok && str != "" {
			// password_hash 是 hex 字符串，需要转换为 []byte
			if hashBytes, err := hex.DecodeString(str); err == nil {
				inherentAttrs.PasswordHash = &domain.UpdateBytes{Action: domain.UpdateActionUpdate, Value: hashBytes}
			}
		}
	}
	if val, exists := payload["status"]; exists {
		if val == nil {
			inherentAttrs.Status = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
		} else if str, ok := val.(string); ok {
			inherentAttrs.Status = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
		}
	}
	if val, exists := payload["service_level"]; exists {
		if val == nil {
			inherentAttrs.ServiceLevel = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
		} else if str, ok := val.(string); ok {
			inherentAttrs.ServiceLevel = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
		}
	}
	if val, exists := payload["admission_date"]; exists {
		if val == nil {
			inherentAttrs.AdmissionDate = &domain.UpdateTime{Action: domain.UpdateActionDelete, Value: nil}
		} else if str, ok := val.(string); ok && str != "" {
			if t, err := time.Parse("2006-01-02", str); err == nil {
				inherentAttrs.AdmissionDate = &domain.UpdateTime{Action: domain.UpdateActionUpdate, Value: &t}
			}
		}
	}
	if val, exists := payload["discharge_date"]; exists {
		if val == nil {
			inherentAttrs.DischargeDate = &domain.UpdateTime{Action: domain.UpdateActionDelete, Value: nil}
		} else if str, ok := val.(string); ok {
			if str == "" {
				inherentAttrs.DischargeDate = &domain.UpdateTime{Action: domain.UpdateActionDelete, Value: nil}
		} else {
				if t, err := time.Parse("2006-01-02", str); err == nil {
					inherentAttrs.DischargeDate = &domain.UpdateTime{Action: domain.UpdateActionUpdate, Value: &t}
				}
			}
		}
	}
	if val, exists := payload["branch_id"]; exists {
		if val == nil {
			inherentAttrs.BranchID = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
		} else if str, ok := val.(string); ok {
			inherentAttrs.BranchID = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
		}
	}
	if val, exists := payload["is_access_enabled"]; exists {
		if val == nil {
			inherentAttrs.IsAccessEnabled = &domain.UpdateBool{Action: domain.UpdateActionDelete, Value: false}
		} else if b, ok := val.(bool); ok {
			inherentAttrs.IsAccessEnabled = &domain.UpdateBool{Action: domain.UpdateActionUpdate, Value: b}
		}
	}
	if val, exists := payload["note"]; exists {
		if val == nil {
			inherentAttrs.Note = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
		} else if str, ok := val.(string); ok {
			inherentAttrs.Note = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
		}
	}
	if val, exists := payload["phone"]; exists {
		if val == nil {
			inherentAttrs.Phone = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
		} else if str, ok := val.(string); ok {
			inherentAttrs.Phone = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
		}
	}
	if val, exists := payload["email"]; exists {
		if val == nil {
			inherentAttrs.Email = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
		} else if str, ok := val.(string); ok {
			inherentAttrs.Email = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
		}
	}
	if val, exists := payload["phone_hash"]; exists {
		if val == nil {
			inherentAttrs.PhoneHash = &domain.UpdateBytes{Action: domain.UpdateActionDelete, Value: nil}
		} else if str, ok := val.(string); ok && str != "" {
			if hashBytes, err := hex.DecodeString(str); err == nil {
				inherentAttrs.PhoneHash = &domain.UpdateBytes{Action: domain.UpdateActionUpdate, Value: hashBytes}
			}
		}
	}
	if val, exists := payload["email_hash"]; exists {
		if val == nil {
			inherentAttrs.EmailHash = &domain.UpdateBytes{Action: domain.UpdateActionDelete, Value: nil}
		} else if str, ok := val.(string); ok && str != "" {
			if hashBytes, err := hex.DecodeString(str); err == nil {
				inherentAttrs.EmailHash = &domain.UpdateBytes{Action: domain.UpdateActionUpdate, Value: hashBytes}
			}
		}
	}
	if val, exists := payload["metadata"]; exists {
		if val == nil {
			inherentAttrs.Metadata = &domain.UpdateJSON{Action: domain.UpdateActionDelete, Value: nil}
		} else {
			if jsonBytes, err := json.Marshal(val); err == nil {
				inherentAttrs.Metadata = &domain.UpdateJSON{Action: domain.UpdateActionUpdate, Value: jsonBytes}
			}
		}
	}

	// 2. 处理 PHI 数据（resident_phi 表字段）
	if phiData, ok := payload["phi"].(map[string]any); ok {
		phi := &service.UpdateResidentPHIRequest{}

		// 使用辅助函数处理各个字段
		if val, exists := phiData["first_name"]; exists {
			if val == nil {
				phi.FirstName = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
			} else if str, ok := val.(string); ok && str != "" {
				phi.FirstName = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
			}
		}
		if val, exists := phiData["last_name"]; exists {
			if val == nil {
				phi.LastName = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
			} else if str, ok := val.(string); ok && str != "" {
				phi.LastName = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
			}
		}
		if val, exists := phiData["gender"]; exists {
			if val == nil {
				phi.Gender = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
			} else if str, ok := val.(string); ok && str != "" {
				phi.Gender = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
			}
		}
		if val, exists := phiData["date_of_birth"]; exists {
			if val == nil {
				phi.DateOfBirth = &domain.UpdateTime{Action: domain.UpdateActionDelete, Value: nil}
			} else if str, ok := val.(string); ok && str != "" {
				if t, err := time.Parse("2006-01-02", str); err == nil {
					phi.DateOfBirth = &domain.UpdateTime{Action: domain.UpdateActionUpdate, Value: &t}
				}
			}
		}
		if val, exists := phiData["resident_phone"]; exists {
			if val == nil {
				phi.ResidentPhone = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
			} else if str, ok := val.(string); ok {
				phi.ResidentPhone = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
			}
		}
		if val, exists := phiData["resident_email"]; exists {
			if val == nil {
				phi.ResidentEmail = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
			} else if str, ok := val.(string); ok {
				phi.ResidentEmail = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
			}
		}
		if val, exists := phiData["save_phone"]; exists {
			if val == nil {
				phi.SavePhone = &domain.UpdateBool{Action: domain.UpdateActionDelete, Value: false}
			} else if b, ok := val.(bool); ok {
				phi.SavePhone = &domain.UpdateBool{Action: domain.UpdateActionUpdate, Value: b}
			}
		}
		if val, exists := phiData["save_email"]; exists {
			if val == nil {
				phi.SaveEmail = &domain.UpdateBool{Action: domain.UpdateActionDelete, Value: false}
			} else if b, ok := val.(bool); ok {
				phi.SaveEmail = &domain.UpdateBool{Action: domain.UpdateActionUpdate, Value: b}
			}
		}
		if val, exists := phiData["weight_lb"]; exists {
			if val == nil {
				phi.WeightLb = &domain.UpdateFloat64{Action: domain.UpdateActionDelete, Value: 0}
			} else if f, ok := val.(float64); ok {
				phi.WeightLb = &domain.UpdateFloat64{Action: domain.UpdateActionUpdate, Value: f}
			}
		}
		if val, exists := phiData["height_ft"]; exists {
			if val == nil {
				phi.HeightFt = &domain.UpdateFloat64{Action: domain.UpdateActionDelete, Value: 0}
			} else if f, ok := val.(float64); ok {
				phi.HeightFt = &domain.UpdateFloat64{Action: domain.UpdateActionUpdate, Value: f}
			}
		}
		if val, exists := phiData["height_in"]; exists {
			if val == nil {
				phi.HeightIn = &domain.UpdateFloat64{Action: domain.UpdateActionDelete, Value: 0}
			} else if f, ok := val.(float64); ok {
				phi.HeightIn = &domain.UpdateFloat64{Action: domain.UpdateActionUpdate, Value: f}
			}
		}
		if val, exists := phiData["mobility_level"]; exists {
			if val == nil {
				phi.MobilityLevel = &domain.UpdateInt{Action: domain.UpdateActionDelete, Value: 0}
			} else if f, ok := val.(float64); ok {
				phi.MobilityLevel = &domain.UpdateInt{Action: domain.UpdateActionUpdate, Value: int(f)}
			}
		}
		if val, exists := phiData["tremor_status"]; exists {
			if val == nil {
				phi.TremorStatus = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
			} else if str, ok := val.(string); ok && str != "" {
				phi.TremorStatus = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
			}
		}
		if val, exists := phiData["mobility_aid"]; exists {
			if val == nil {
				phi.MobilityAid = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
			} else if str, ok := val.(string); ok && str != "" {
				phi.MobilityAid = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
			}
		}
		if val, exists := phiData["adl_assistance"]; exists {
			if val == nil {
				phi.ADLAssistance = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
			} else if str, ok := val.(string); ok && str != "" {
				phi.ADLAssistance = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
			}
		}
		if val, exists := phiData["comm_status"]; exists {
			if val == nil {
				phi.CommStatus = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
			} else if str, ok := val.(string); ok && str != "" {
				phi.CommStatus = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
			}
		}
		if val, exists := phiData["has_hypertension"]; exists {
			if val == nil {
				phi.HasHypertension = &domain.UpdateBool{Action: domain.UpdateActionDelete, Value: false}
			} else if b, ok := val.(bool); ok {
				phi.HasHypertension = &domain.UpdateBool{Action: domain.UpdateActionUpdate, Value: b}
			}
		}
		if val, exists := phiData["has_hyperlipaemia"]; exists {
			if val == nil {
				phi.HasHyperlipaemia = &domain.UpdateBool{Action: domain.UpdateActionDelete, Value: false}
			} else if b, ok := val.(bool); ok {
				phi.HasHyperlipaemia = &domain.UpdateBool{Action: domain.UpdateActionUpdate, Value: b}
			}
		}
		if val, exists := phiData["has_hyperglycaemia"]; exists {
			if val == nil {
				phi.HasHyperglycaemia = &domain.UpdateBool{Action: domain.UpdateActionDelete, Value: false}
			} else if b, ok := val.(bool); ok {
				phi.HasHyperglycaemia = &domain.UpdateBool{Action: domain.UpdateActionUpdate, Value: b}
			}
		}
		if val, exists := phiData["has_stroke_history"]; exists {
			if val == nil {
				phi.HasStrokeHistory = &domain.UpdateBool{Action: domain.UpdateActionDelete, Value: false}
			} else if b, ok := val.(bool); ok {
				phi.HasStrokeHistory = &domain.UpdateBool{Action: domain.UpdateActionUpdate, Value: b}
			}
		}
		if val, exists := phiData["has_paralysis"]; exists {
			if val == nil {
				phi.HasParalysis = &domain.UpdateBool{Action: domain.UpdateActionDelete, Value: false}
			} else if b, ok := val.(bool); ok {
				phi.HasParalysis = &domain.UpdateBool{Action: domain.UpdateActionUpdate, Value: b}
			}
		}
		if val, exists := phiData["has_alzheimer"]; exists {
			if val == nil {
				phi.HasAlzheimer = &domain.UpdateBool{Action: domain.UpdateActionDelete, Value: false}
			} else if b, ok := val.(bool); ok {
				phi.HasAlzheimer = &domain.UpdateBool{Action: domain.UpdateActionUpdate, Value: b}
			}
		}
		if val, exists := phiData["medical_history"]; exists {
			if val == nil {
				phi.MedicalHistory = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
			} else if str, ok := val.(string); ok && str != "" {
				phi.MedicalHistory = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
			}
		}
		if val, exists := phiData["home_address_street"]; exists {
			if val == nil {
				phi.HomeAddressStreet = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
			} else if str, ok := val.(string); ok && str != "" {
				phi.HomeAddressStreet = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
			}
		}
		if val, exists := phiData["home_address_city"]; exists {
			if val == nil {
				phi.HomeAddressCity = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
			} else if str, ok := val.(string); ok && str != "" {
				phi.HomeAddressCity = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
			}
		}
		if val, exists := phiData["home_address_state"]; exists {
			if val == nil {
				phi.HomeAddressState = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
			} else if str, ok := val.(string); ok && str != "" {
				phi.HomeAddressState = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
			}
		}
		if val, exists := phiData["home_address_postal_code"]; exists {
			if val == nil {
				phi.HomeAddressPostalCode = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
			} else if str, ok := val.(string); ok && str != "" {
				phi.HomeAddressPostalCode = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
			}
		}
		if val, exists := phiData["plus_code"]; exists {
			if val == nil {
				phi.PlusCode = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
			} else if str, ok := val.(string); ok && str != "" {
				phi.PlusCode = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
			}
		}

		inherentAttrs.PHI = phi
	}

	// 3. 处理 Contacts（resident_contacts 表字段）
	// 注意：前端可能通过单独的 API 更新 contacts，这里先支持在 UpdateResident 中一起更新
	if contacts, ok := payload["contacts"].([]any); ok {
		contactReqs := make([]*service.UpdateResidentContactRequest, 0, len(contacts))
		for _, contactRaw := range contacts {
			if contactData, ok := contactRaw.(map[string]any); ok {
				slotVal, slotExists := contactData["slot"]
				if !slotExists {
					continue // slot 是必填字段
				}
				slot, ok := slotVal.(string)
				if !ok || slot == "" {
					continue
				}

				contactReq := &service.UpdateResidentContactRequest{
					Slot: slot,
				}

				if val, exists := contactData["is_enabled"]; exists {
					if val == nil {
						contactReq.IsEnabled = &domain.UpdateBool{Action: domain.UpdateActionDelete, Value: false}
					} else if b, ok := val.(bool); ok {
						contactReq.IsEnabled = &domain.UpdateBool{Action: domain.UpdateActionUpdate, Value: b}
					}
				}
				if val, exists := contactData["relationship"]; exists {
					if val == nil {
						contactReq.Relationship = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
					} else if str, ok := val.(string); ok && str != "" {
						contactReq.Relationship = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
					}
				}
				if val, exists := contactData["contact_first_name"]; exists {
					if val == nil {
						contactReq.ContactFirstName = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
					} else if str, ok := val.(string); ok && str != "" {
						contactReq.ContactFirstName = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
					}
				}
				if val, exists := contactData["contact_last_name"]; exists {
					if val == nil {
						contactReq.ContactLastName = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
					} else if str, ok := val.(string); ok && str != "" {
						contactReq.ContactLastName = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
					}
				}
				if val, exists := contactData["contact_phone"]; exists {
					if val == nil {
						contactReq.ContactPhone = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
					} else if str, ok := val.(string); ok {
						contactReq.ContactPhone = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
					}
				}
				if val, exists := contactData["contact_email"]; exists {
					if val == nil {
						contactReq.ContactEmail = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
					} else if str, ok := val.(string); ok {
						contactReq.ContactEmail = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
					}
				}
				// Handle contact_phone_hash (for search)
				if val, exists := contactData["contact_phone_hash"]; exists {
					if val == nil {
						contactReq.ContactPhoneHash = &domain.UpdateBytes{Action: domain.UpdateActionDelete, Value: nil}
					} else if str, ok := val.(string); ok && str != "" {
						hashBytes, err := hex.DecodeString(str)
						if err == nil && len(hashBytes) > 0 {
							contactReq.ContactPhoneHash = &domain.UpdateBytes{Action: domain.UpdateActionUpdate, Value: hashBytes}
						}
					}
				}
				// Handle contact_email_hash (for search)
				if val, exists := contactData["contact_email_hash"]; exists {
					if val == nil {
						contactReq.ContactEmailHash = &domain.UpdateBytes{Action: domain.UpdateActionDelete, Value: nil}
					} else if str, ok := val.(string); ok && str != "" {
						hashBytes, err := hex.DecodeString(str)
						if err == nil && len(hashBytes) > 0 {
							contactReq.ContactEmailHash = &domain.UpdateBytes{Action: domain.UpdateActionUpdate, Value: hashBytes}
						}
					}
				}
				if val, exists := contactData["receive_sms"]; exists {
					if val == nil {
						contactReq.ReceiveSMS = &domain.UpdateBool{Action: domain.UpdateActionDelete, Value: false}
					} else if b, ok := val.(bool); ok {
						contactReq.ReceiveSMS = &domain.UpdateBool{Action: domain.UpdateActionUpdate, Value: b}
					}
				}
				if val, exists := contactData["receive_email"]; exists {
					if val == nil {
						contactReq.ReceiveEmail = &domain.UpdateBool{Action: domain.UpdateActionDelete, Value: false}
					} else if b, ok := val.(bool); ok {
						contactReq.ReceiveEmail = &domain.UpdateBool{Action: domain.UpdateActionUpdate, Value: b}
					}
				}
				if val, exists := contactData["alert_time_window"]; exists {
					if val == nil {
						contactReq.AlertTimeWindow = &domain.UpdateJSON{Action: domain.UpdateActionDelete, Value: nil}
					} else {
						if jsonBytes, err := json.Marshal(val); err == nil {
							contactReq.AlertTimeWindow = &domain.UpdateJSON{Action: domain.UpdateActionUpdate, Value: jsonBytes}
						}
					}
				}

				contactReqs = append(contactReqs, contactReq)
			}
		}
		if len(contactReqs) > 0 {
			inherentAttrs.Contacts = contactReqs
		}
	}

	// 只有至少有一个字段需要更新时，才设置 InherentAttributes
	if inherentAttrs != nil && (inherentAttrs.ResidentAccount != nil || inherentAttrs.Nickname != nil ||
		inherentAttrs.PasswordHash != nil || inherentAttrs.Status != nil || inherentAttrs.ServiceLevel != nil ||
		inherentAttrs.AdmissionDate != nil || inherentAttrs.DischargeDate != nil || inherentAttrs.BranchID != nil ||
		inherentAttrs.IsAccessEnabled != nil || inherentAttrs.Note != nil || inherentAttrs.Phone != nil ||
		inherentAttrs.Email != nil || inherentAttrs.PhoneHash != nil || inherentAttrs.EmailHash != nil ||
		inherentAttrs.Metadata != nil || inherentAttrs.PHI != nil || len(inherentAttrs.Contacts) > 0) {
		req.InherentAttributes = inherentAttrs
	}

	// 4. 处理 UnitRelation（位置分配）
	unitRelation := &service.UpdateResidentUnitRelation{}
	hasUnitRelation := false

	if val, exists := payload["unit_id"]; exists {
		hasUnitRelation = true
		if val == nil {
			unitRelation.UnitID = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
		} else if str, ok := val.(string); ok {
			unitRelation.UnitID = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
		}
	}
	if val, exists := payload["room_id"]; exists {
		hasUnitRelation = true
		if val == nil {
			unitRelation.RoomID = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
		} else if str, ok := val.(string); ok {
			unitRelation.RoomID = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
		}
	}
	if val, exists := payload["bed_id"]; exists {
		hasUnitRelation = true
		if val == nil {
			unitRelation.BedID = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
		} else if str, ok := val.(string); ok {
			unitRelation.BedID = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
		}
	}

	if hasUnitRelation {
		req.UnitRelation = unitRelation
	}

	// 5. 处理 CaregiverRelation（护理人员分配）
	// 注意：前端发送的格式是 {userList: [], groupList: []}（驼峰格式）
	if caregivers, ok := payload["caregivers"].(map[string]any); ok {
		cgRelation := &service.UpdateResidentCaregiverRelation{}

		if val, exists := caregivers["userList"]; exists {
			if val == nil {
				cgRelation.UserList = &domain.UpdateJSON{Action: domain.UpdateActionDelete, Value: nil}
			} else if userList, ok := val.([]any); ok {
				if jsonBytes, err := json.Marshal(userList); err == nil {
					cgRelation.UserList = &domain.UpdateJSON{Action: domain.UpdateActionUpdate, Value: jsonBytes}
				}
			}
		}
		if val, exists := caregivers["groupList"]; exists {
			if val == nil {
				cgRelation.GroupList = &domain.UpdateJSON{Action: domain.UpdateActionDelete, Value: nil}
			} else if groupList, ok := val.([]any); ok {
				if jsonBytes, err := json.Marshal(groupList); err == nil {
					cgRelation.GroupList = &domain.UpdateJSON{Action: domain.UpdateActionUpdate, Value: jsonBytes}
				}
			}
		}

		// 只有至少有一个字段需要更新时，才设置 CaregiverRelation
		if cgRelation.UserList != nil || cgRelation.GroupList != nil {
			req.CaregiverRelation = cgRelation
		}
	}

	resp, err := h.residentService.UpdateResident(ctx, req)
	if err != nil {
		h.logger.Error("UpdateResident failed",
			zap.String("tenant_id", tenantID),
			zap.String("resident_id", residentID),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"success": resp.Success,
	}))
}

// ============================================
// UpdateResidentPHI 更新住户 PHI
// ============================================

func (h *ResidentHandler) UpdateResidentPHI(w http.ResponseWriter, r *http.Request, residentID string) {
	ctx := r.Context()

	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	currentUserID := r.Header.Get("X-User-Id")
	currentUserType := r.Header.Get("X-User-Type")
	currentUserRole := r.Header.Get("X-User-Role")

	// Permission check: Resident/Family cannot update PHI
	// 注意：resident_contacts 不能登录系统，所以 currentUserType 永远不会是 "family"
	// 保留此检查是为了向后兼容，但实际上只会是 "resident" 或 "staff"
	if currentUserType == "resident" || currentUserType == "family" {
		writeJSON(w, http.StatusOK, Fail("permission denied: resident/family cannot update PHI"))
		return
	}

	// 解析请求体
	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	// 构建 UpdateResidentRequest，只包含 PHI 数据（权限检查由 Service 层自己处理）
	req := service.UpdateResidentRequest{
		TenantID:        tenantID,
		ResidentID:      residentID,
		CurrentUserID:   currentUserID,
		CurrentUserRole: currentUserRole,
		PermissionCheck: nil, // Service 层自己查询权限
	}

	// 构建 InherentAttributes（residents 表字段 + PHI）
	inherentAttrs := &service.UpdateResidentInherentAttributes{}

	// 1. 处理 phone_hash 和 email_hash（residents 表字段）
	// 这些字段可能在 payload 的顶层，也可能在 phi 对象中
	if val, exists := payload["phone_hash"]; exists {
		if val == nil {
			inherentAttrs.PhoneHash = &domain.UpdateBytes{Action: domain.UpdateActionDelete, Value: nil}
		} else if str, ok := val.(string); ok && str != "" {
			if hashBytes, err := hex.DecodeString(str); err == nil {
				inherentAttrs.PhoneHash = &domain.UpdateBytes{Action: domain.UpdateActionUpdate, Value: hashBytes}
			}
		}
	}
	if val, exists := payload["email_hash"]; exists {
		if val == nil {
			inherentAttrs.EmailHash = &domain.UpdateBytes{Action: domain.UpdateActionDelete, Value: nil}
		} else if str, ok := val.(string); ok && str != "" {
			if hashBytes, err := hex.DecodeString(str); err == nil {
				inherentAttrs.EmailHash = &domain.UpdateBytes{Action: domain.UpdateActionUpdate, Value: hashBytes}
			}
		}
	}

	// 2. 处理 PHI 数据（resident_phi 表字段）
	// 支持两种格式：payload["phi"] 对象，或 payload 顶层字段
	var phiData map[string]any
	if phiObj, ok := payload["phi"].(map[string]any); ok {
		phiData = phiObj
		// 如果 phone_hash 和 email_hash 在 phi 对象中，也要处理
		if val, exists := phiData["phone_hash"]; exists {
			if val == nil {
				inherentAttrs.PhoneHash = &domain.UpdateBytes{Action: domain.UpdateActionDelete, Value: nil}
			} else if str, ok := val.(string); ok && str != "" {
				if hashBytes, err := hex.DecodeString(str); err == nil {
					inherentAttrs.PhoneHash = &domain.UpdateBytes{Action: domain.UpdateActionUpdate, Value: hashBytes}
				}
			}
		}
		if val, exists := phiData["email_hash"]; exists {
			if val == nil {
				inherentAttrs.EmailHash = &domain.UpdateBytes{Action: domain.UpdateActionDelete, Value: nil}
			} else if str, ok := val.(string); ok && str != "" {
				if hashBytes, err := hex.DecodeString(str); err == nil {
					inherentAttrs.EmailHash = &domain.UpdateBytes{Action: domain.UpdateActionUpdate, Value: hashBytes}
				}
			}
		}
	} else {
		// 如果没有 phi 字段，直接使用 payload 作为 PHI 数据（前端直接发送字段，不在 phi 对象中）
		phiData = payload
	}

	// 处理 PHI 字段（使用 domain.UpdateX 类型）
	phi := &service.UpdateResidentPHIRequest{}

	if val, exists := phiData["first_name"]; exists {
		if val == nil {
			phi.FirstName = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
		} else if str, ok := val.(string); ok && str != "" {
			phi.FirstName = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
		}
	}
	if val, exists := phiData["last_name"]; exists {
		if val == nil {
			phi.LastName = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
		} else if str, ok := val.(string); ok && str != "" {
			phi.LastName = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
		}
	}
	if val, exists := phiData["gender"]; exists {
		if val == nil {
			phi.Gender = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
		} else if str, ok := val.(string); ok && str != "" {
			phi.Gender = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
		}
	}
	if val, exists := phiData["date_of_birth"]; exists {
		if val == nil {
			phi.DateOfBirth = &domain.UpdateTime{Action: domain.UpdateActionDelete, Value: nil}
		} else if str, ok := val.(string); ok && str != "" {
			if t, err := time.Parse("2006-01-02", str); err == nil {
				phi.DateOfBirth = &domain.UpdateTime{Action: domain.UpdateActionUpdate, Value: &t}
			}
		}
	}
	if val, exists := phiData["resident_phone"]; exists {
		if val == nil {
			phi.ResidentPhone = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
		} else if str, ok := val.(string); ok {
			phi.ResidentPhone = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
		}
	}
	if val, exists := phiData["resident_email"]; exists {
		if val == nil {
			phi.ResidentEmail = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
		} else if str, ok := val.(string); ok {
			phi.ResidentEmail = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
		}
	}
	if val, exists := phiData["save_phone"]; exists {
		if val == nil {
			phi.SavePhone = &domain.UpdateBool{Action: domain.UpdateActionDelete, Value: false}
		} else if b, ok := val.(bool); ok {
			phi.SavePhone = &domain.UpdateBool{Action: domain.UpdateActionUpdate, Value: b}
		}
	}
	if val, exists := phiData["save_email"]; exists {
		if val == nil {
			phi.SaveEmail = &domain.UpdateBool{Action: domain.UpdateActionDelete, Value: false}
		} else if b, ok := val.(bool); ok {
			phi.SaveEmail = &domain.UpdateBool{Action: domain.UpdateActionUpdate, Value: b}
		}
	}
	if val, exists := phiData["weight_lb"]; exists {
		if val == nil {
			phi.WeightLb = &domain.UpdateFloat64{Action: domain.UpdateActionDelete, Value: 0}
		} else if f, ok := val.(float64); ok {
			phi.WeightLb = &domain.UpdateFloat64{Action: domain.UpdateActionUpdate, Value: f}
		}
	}
	if val, exists := phiData["height_ft"]; exists {
		if val == nil {
			phi.HeightFt = &domain.UpdateFloat64{Action: domain.UpdateActionDelete, Value: 0}
		} else if f, ok := val.(float64); ok {
			phi.HeightFt = &domain.UpdateFloat64{Action: domain.UpdateActionUpdate, Value: f}
		}
	}
	if val, exists := phiData["height_in"]; exists {
		if val == nil {
			phi.HeightIn = &domain.UpdateFloat64{Action: domain.UpdateActionDelete, Value: 0}
		} else if f, ok := val.(float64); ok {
			phi.HeightIn = &domain.UpdateFloat64{Action: domain.UpdateActionUpdate, Value: f}
		}
	}
	if val, exists := phiData["mobility_level"]; exists {
		if val == nil {
			phi.MobilityLevel = &domain.UpdateInt{Action: domain.UpdateActionDelete, Value: 0}
		} else if f, ok := val.(float64); ok {
			phi.MobilityLevel = &domain.UpdateInt{Action: domain.UpdateActionUpdate, Value: int(f)}
		}
	}
	if val, exists := phiData["tremor_status"]; exists {
		if val == nil {
			phi.TremorStatus = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
		} else if str, ok := val.(string); ok && str != "" {
			phi.TremorStatus = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
		}
	}
	if val, exists := phiData["mobility_aid"]; exists {
		if val == nil {
			phi.MobilityAid = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
		} else if str, ok := val.(string); ok && str != "" {
			phi.MobilityAid = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
		}
	}
	if val, exists := phiData["adl_assistance"]; exists {
		if val == nil {
			phi.ADLAssistance = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
		} else if str, ok := val.(string); ok && str != "" {
			phi.ADLAssistance = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
		}
	}
	if val, exists := phiData["comm_status"]; exists {
		if val == nil {
			phi.CommStatus = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
		} else if str, ok := val.(string); ok && str != "" {
			phi.CommStatus = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
		}
	}
	if val, exists := phiData["has_hypertension"]; exists {
		if val == nil {
			phi.HasHypertension = &domain.UpdateBool{Action: domain.UpdateActionDelete, Value: false}
		} else if b, ok := val.(bool); ok {
			phi.HasHypertension = &domain.UpdateBool{Action: domain.UpdateActionUpdate, Value: b}
		}
	}
	if val, exists := phiData["has_hyperlipaemia"]; exists {
		if val == nil {
			phi.HasHyperlipaemia = &domain.UpdateBool{Action: domain.UpdateActionDelete, Value: false}
		} else if b, ok := val.(bool); ok {
			phi.HasHyperlipaemia = &domain.UpdateBool{Action: domain.UpdateActionUpdate, Value: b}
		}
	}
	if val, exists := phiData["has_hyperglycaemia"]; exists {
		if val == nil {
			phi.HasHyperglycaemia = &domain.UpdateBool{Action: domain.UpdateActionDelete, Value: false}
		} else if b, ok := val.(bool); ok {
			phi.HasHyperglycaemia = &domain.UpdateBool{Action: domain.UpdateActionUpdate, Value: b}
		}
	}
	if val, exists := phiData["has_stroke_history"]; exists {
		if val == nil {
			phi.HasStrokeHistory = &domain.UpdateBool{Action: domain.UpdateActionDelete, Value: false}
		} else if b, ok := val.(bool); ok {
			phi.HasStrokeHistory = &domain.UpdateBool{Action: domain.UpdateActionUpdate, Value: b}
		}
	}
	if val, exists := phiData["has_paralysis"]; exists {
		if val == nil {
			phi.HasParalysis = &domain.UpdateBool{Action: domain.UpdateActionDelete, Value: false}
		} else if b, ok := val.(bool); ok {
			phi.HasParalysis = &domain.UpdateBool{Action: domain.UpdateActionUpdate, Value: b}
		}
	}
	if val, exists := phiData["has_alzheimer"]; exists {
		if val == nil {
			phi.HasAlzheimer = &domain.UpdateBool{Action: domain.UpdateActionDelete, Value: false}
		} else if b, ok := val.(bool); ok {
			phi.HasAlzheimer = &domain.UpdateBool{Action: domain.UpdateActionUpdate, Value: b}
		}
	}
	if val, exists := phiData["medical_history"]; exists {
		if val == nil {
			phi.MedicalHistory = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
		} else if str, ok := val.(string); ok && str != "" {
			phi.MedicalHistory = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
		}
	}
	if val, exists := phiData["home_address_street"]; exists {
		if val == nil {
			phi.HomeAddressStreet = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
		} else if str, ok := val.(string); ok && str != "" {
			phi.HomeAddressStreet = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
		}
	}
	if val, exists := phiData["home_address_city"]; exists {
		if val == nil {
			phi.HomeAddressCity = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
		} else if str, ok := val.(string); ok && str != "" {
			phi.HomeAddressCity = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
		}
	}
	if val, exists := phiData["home_address_state"]; exists {
		if val == nil {
			phi.HomeAddressState = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
		} else if str, ok := val.(string); ok && str != "" {
			phi.HomeAddressState = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
		}
	}
	if val, exists := phiData["home_address_postal_code"]; exists {
		if val == nil {
			phi.HomeAddressPostalCode = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
		} else if str, ok := val.(string); ok && str != "" {
			phi.HomeAddressPostalCode = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
		}
	}
	if val, exists := phiData["plus_code"]; exists {
		if val == nil {
			phi.PlusCode = &domain.UpdateString{Action: domain.UpdateActionDelete, Value: ""}
		} else if str, ok := val.(string); ok && str != "" {
			phi.PlusCode = &domain.UpdateString{Action: domain.UpdateActionUpdate, Value: str}
		}
	}

	// 只有至少有一个 PHI 字段需要更新时，才设置 PHI
	hasPHI := phi.FirstName != nil || phi.LastName != nil || phi.Gender != nil ||
		phi.DateOfBirth != nil || phi.ResidentPhone != nil || phi.ResidentEmail != nil ||
		phi.SavePhone != nil || phi.SaveEmail != nil || phi.WeightLb != nil ||
		phi.HeightFt != nil || phi.HeightIn != nil || phi.MobilityLevel != nil ||
		phi.TremorStatus != nil || phi.MobilityAid != nil || phi.ADLAssistance != nil ||
		phi.CommStatus != nil || phi.HasHypertension != nil || phi.HasHyperlipaemia != nil ||
		phi.HasHyperglycaemia != nil || phi.HasStrokeHistory != nil || phi.HasParalysis != nil ||
		phi.HasAlzheimer != nil || phi.MedicalHistory != nil || phi.HomeAddressStreet != nil ||
		phi.HomeAddressCity != nil || phi.HomeAddressState != nil || phi.HomeAddressPostalCode != nil ||
		phi.PlusCode != nil

	if hasPHI {
		inherentAttrs.PHI = phi
	}

	// 只有至少有一个字段需要更新时，才设置 InherentAttributes
	hasInherentAttrs := inherentAttrs.PhoneHash != nil || inherentAttrs.EmailHash != nil || inherentAttrs.PHI != nil

	if hasInherentAttrs {
		req.InherentAttributes = inherentAttrs
	}

	resp, err := h.residentService.UpdateResident(ctx, req)
	if err != nil {
		h.logger.Error("UpdateResidentPHI failed",
			zap.String("tenant_id", tenantID),
			zap.String("resident_id", residentID),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"success": resp.Success,
	}))
}

// ============================================
// DeleteResident 删除住户
// ============================================

func (h *ResidentHandler) DeleteResident(w http.ResponseWriter, r *http.Request, residentID string) {
	ctx := r.Context()

	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	currentUserID := r.Header.Get("X-User-Id")
	currentUserType := r.Header.Get("X-User-Type")
	currentUserRole := r.Header.Get("X-User-Role")

	// 权限检查结果（Service 层会自己查询用户信息和验证权限，这里只传递权限配置）
	var permCheck *service.PermissionCheckResult
	if currentUserRole != "" && h.db != nil {
		perm, err := GetResourcePermission(h.db, ctx, currentUserRole, "residents", "D")
		if err == nil {
			// Service 层会自己查询用户的 branch_id，这里不需要传递 UserBranchTag
			permCheck = &service.PermissionCheckResult{
				AssignedOnly:  perm.AssignedOnly,
				BranchOnly:    perm.BranchOnly,
			}
		}
	}

	req := service.DeleteResidentRequest{
		TenantID:        tenantID,
		ResidentID:      residentID,
		CurrentUserID:   currentUserID,
		CurrentUserType: currentUserType,
		CurrentUserRole: currentUserRole,
		PermissionCheck: permCheck,
	}

	resp, err := h.residentService.DeleteResident(ctx, req)
	if err != nil {
		h.logger.Error("DeleteResident failed",
			zap.String("tenant_id", tenantID),
			zap.String("resident_id", residentID),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"success": resp.Success,
	}))
}

// ============================================
// ResetResidentPassword 重置住户密码
// ============================================

func (h *ResidentHandler) ResetResidentPassword(w http.ResponseWriter, r *http.Request, residentID string) {
	ctx := r.Context()

	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	currentUserID := r.Header.Get("X-User-Id")
	currentUserType := r.Header.Get("X-User-Type")
	currentUserRole := r.Header.Get("X-User-Role")

	// 解析请求体（可选）
	var payload map[string]any
	var passwordHash string
	if err := readBodyJSON(r, 1<<20, &payload); err == nil {
		if pwd, ok := payload["password_hash"].(string); ok {
			passwordHash = pwd
		}
	}

	// 权限检查结果（Service 层会自己查询用户信息和验证权限，这里只传递权限配置）
	var permCheck *service.PermissionCheckResult
	if currentUserRole != "" && h.db != nil {
		perm, err := GetResourcePermission(h.db, ctx, currentUserRole, "residents", "U")
		if err == nil {
			// Service 层会自己查询用户的 branch_id，这里不需要传递 UserBranchTag
			permCheck = &service.PermissionCheckResult{
				AssignedOnly:  perm.AssignedOnly,
				BranchOnly:    perm.BranchOnly,
			}
		}
	}

	req := service.ResetResidentPasswordRequest{
		TenantID:        tenantID,
		ResidentID:      residentID,
		CurrentUserID:   currentUserID,
		CurrentUserType: currentUserType,
		CurrentUserRole: currentUserRole,
		PermissionCheck: permCheck,
		NewPassword:     passwordHash,
	}

	resp, err := h.residentService.ResetResidentPassword(ctx, req)
	if err != nil {
		h.logger.Error("ResetResidentPassword failed",
			zap.String("tenant_id", tenantID),
			zap.String("resident_id", residentID),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"success":      resp.Success,
		"new_password": resp.NewPassword,
	}))
}


// ============================================
// GetResidentAccountSettings 获取住户/联系人账户设置
// ============================================

// GetResidentAccountSettings 获取住户/联系人账户设置
func (h *ResidentHandler) GetResidentAccountSettings(w http.ResponseWriter, r *http.Request, residentID string) {
	ctx := r.Context()

	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	currentUserID := r.Header.Get("X-User-Id")
	if currentUserID == "" {
		writeJSON(w, http.StatusOK, Fail("user ID is required"))
		return
	}

	currentUserType := r.Header.Get("X-User-Type")
	currentUserRole := r.Header.Get("X-User-Role")

	req := service.GetResidentAccountSettingsRequest{
		TenantID:        tenantID,
		ResidentID:      residentID,
		CurrentUserID:   currentUserID,
		CurrentUserType: currentUserType,
		CurrentUserRole: currentUserRole,
	}

	// Input log
	fmt.Printf("[GetResidentAccountSettings] INPUT: tenant_id=%s, resident_id=%s, current_user_id=%s, current_user_role=%s\n",
		req.TenantID, req.ResidentID, req.CurrentUserID, req.CurrentUserRole)

	resp, err := h.residentService.GetResidentAccountSettings(ctx, req)
	if err != nil {
		h.logger.Error("GetResidentAccountSettings failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	item := map[string]any{
		"id":         residentID, // UUID: resident_id 或 contact_id
		"nickname":   resp.Nickname,
		"is_contact": resp.IsContact,
		"role":       currentUserRole, // 角色代码
	}
	if resp.ResidentAccount != nil {
		item["account"] = *resp.ResidentAccount
	} else {
		// resp.ResidentAccount is nil, item["account"] will not be set
	}

	// Output log
	accountValue := ""
	if resp.ResidentAccount != nil {
		accountValue = *resp.ResidentAccount
	}
	fmt.Printf("[GetResidentAccountSettings] OUTPUT: tenant_id=%s, resident_id=%s, account=%s, nickname=%s, is_contact=%v\n",
		req.TenantID, req.ResidentID, accountValue, resp.Nickname, resp.IsContact)
	if resp.Email != nil {
		item["email"] = *resp.Email
	}
	if resp.Phone != nil {
		item["phone"] = *resp.Phone
	}
	// resident 和 contact 都需要返回 save 标志
	item["save_email"] = resp.SaveEmail
	item["save_phone"] = resp.SavePhone

	writeJSON(w, http.StatusOK, Ok(item))
}

// ============================================
// UpdateResidentAccountSettings 更新住户/联系人账户设置
// ============================================

// UpdateResidentAccountSettings 更新住户/联系人账户设置（统一 API）
func (h *ResidentHandler) UpdateResidentAccountSettings(w http.ResponseWriter, r *http.Request, residentID string) {
	ctx := r.Context()

	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	currentUserID := r.Header.Get("X-User-Id")
	if currentUserID == "" {
		writeJSON(w, http.StatusOK, Fail("user ID is required"))
		return
	}

	currentUserType := r.Header.Get("X-User-Type")
	currentUserRole := r.Header.Get("X-User-Role")

	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	req := service.UpdateResidentAccountSettingsRequest{
		TenantID:        tenantID,
		ResidentID:      residentID,
		CurrentUserID:   currentUserID,
		CurrentUserType: currentUserType,
		CurrentUserRole: currentUserRole,
	}

	// 解析 password_hash
	if passwordHash, ok := payload["password_hash"].(string); ok && passwordHash != "" {
		req.PasswordHash = &passwordHash
	}

	// 解析 email 和 email_hash
	if email, ok := payload["email"].(string); ok {
		req.Email = &email
	}
	if emailHash, ok := payload["email_hash"].(string); ok && emailHash != "" {
		req.EmailHash = &emailHash
	}

	// 解析 phone 和 phone_hash
	if phone, ok := payload["phone"].(string); ok {
		req.Phone = &phone
	}
	if phoneHash, ok := payload["phone_hash"].(string); ok && phoneHash != "" {
		req.PhoneHash = &phoneHash
	}

	// 解析 save_email 和 save_phone（仅 resident 需要）
	if saveEmail, ok := payload["save_email"].(bool); ok {
		req.SaveEmail = &saveEmail
	}
	if savePhone, ok := payload["save_phone"].(bool); ok {
		req.SavePhone = &savePhone
	}

	// 检查是否有任何更新
	if req.PasswordHash == nil && req.Email == nil && req.Phone == nil {
		writeJSON(w, http.StatusOK, Fail("no fields to update"))
		return
	}

	resp, err := h.residentService.UpdateResidentAccountSettings(ctx, req)
	if err != nil {
		h.logger.Error("UpdateResidentAccountSettings failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"success": resp.Success,
		"message": resp.Message,
	}))
}

// ============================================
// UpdateResidentContact 更新联系人信息
// ============================================

func (h *ResidentHandler) UpdateResidentContact(w http.ResponseWriter, r *http.Request, residentID string) {
	ctx := r.Context()

	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	currentUserID := r.Header.Get("X-User-Id")
	currentUserType := r.Header.Get("X-User-Type")
	currentUserRole := r.Header.Get("X-User-Role")

	// 解析请求体
	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	// 获取 slot（必填）：通过 resident_id + slot 定位 contact
	slot, ok := payload["slot"].(string)
	if !ok || slot == "" {
		writeJSON(w, http.StatusOK, Fail("slot is required"))
		return
	}

	// 构建 Service 请求（权限检查由 Service 层自己处理）
	req := service.UpdateResidentContactStandaloneRequest{
		TenantID:        tenantID,
		ResidentID:      residentID,
		Slot:            slot, // slot 是必填的，用于定位 contact
		CurrentUserID:   currentUserID,
		CurrentUserType: currentUserType,
		CurrentUserRole: currentUserRole,
	}

	// 解析字段（使用指针表示可选）
	// 规则：
	//   - 字段不存在 → nil（不更新）
	//   - 字段为 null → ""（删除，Repository 会转换为 NULL）
	//   - 字段为 "" → ""（删除，Repository 会转换为 NULL）
	//   - 字段有值 → 有值（更新）
	if isEnabled, ok := payload["is_enabled"].(bool); ok {
		req.IsEnabled = &isEnabled
	}
	// 处理 contact_first_name：支持 string 和 null
	if firstName, ok := payload["contact_first_name"].(string); ok {
		req.ContactFirstName = &firstName // "" 表示删除
	} else if _, exists := payload["contact_first_name"]; exists && payload["contact_first_name"] == nil {
		// Vue 发送 null 时，转换为 ""（删除）
		emptyStr := ""
		req.ContactFirstName = &emptyStr
	}
	// contact_first_name 字段不存在 → nil（不更新）
	// 处理 contact_last_name：支持 string 和 null
	if lastName, ok := payload["contact_last_name"].(string); ok {
		req.ContactLastName = &lastName // "" 表示删除
	} else if _, exists := payload["contact_last_name"]; exists && payload["contact_last_name"] == nil {
		// Vue 发送 null 时，转换为 ""（删除）
		emptyStr := ""
		req.ContactLastName = &emptyStr
	}
	// contact_last_name 字段不存在 → nil（不更新）
	// 处理 relationship：支持 string 和 null
	if relationship, ok := payload["relationship"].(string); ok {
		req.Relationship = &relationship // "" 表示删除
	} else if _, exists := payload["relationship"]; exists && payload["relationship"] == nil {
		// Vue 发送 null 时，转换为 ""（删除）
		emptyStr := ""
		req.Relationship = &emptyStr
	}
	// relationship 字段不存在 → nil（不更新）
	// 处理 contact_phone：支持 string 和 null
	if phone, ok := payload["contact_phone"].(string); ok {
		req.ContactPhone = &phone // "" 表示删除
	} else if payload["contact_phone"] == nil {
		// Vue 发送 null 时，转换为 ""（删除）
		emptyStr := ""
		req.ContactPhone = &emptyStr
	}
	// contact_phone 字段不存在 → nil（不更新）
	// 处理 contact_email：支持 string 和 null
	if email, ok := payload["contact_email"].(string); ok {
		req.ContactEmail = &email // "" 表示删除
	} else if payload["contact_email"] == nil {
		// Vue 发送 null 时，转换为 ""（删除）
		emptyStr := ""
		req.ContactEmail = &emptyStr
	}
	// contact_email 字段不存在 → nil（不更新）
	
	// 处理 contact_phone_hash（用于搜索）：支持 string 和 null
	if phoneHash, ok := payload["contact_phone_hash"].(string); ok && phoneHash != "" {
		req.PhoneHash = &phoneHash // 有效的 hash hex 字符串
	} else if payload["contact_phone_hash"] == nil {
		// Vue 发送 null 时，表示删除 hash（phone 被删除）
		emptyStr := ""
		req.PhoneHash = &emptyStr
	}
	// contact_phone_hash 字段不存在 → nil（不更新）
	
	// 处理 contact_email_hash（用于搜索）：支持 string 和 null
	if emailHash, ok := payload["contact_email_hash"].(string); ok && emailHash != "" {
		req.EmailHash = &emailHash // 有效的 hash hex 字符串
	} else if payload["contact_email_hash"] == nil {
		// Vue 发送 null 时，表示删除 hash（email 被删除）
		emptyStr := ""
		req.EmailHash = &emptyStr
	}
	// contact_email_hash 字段不存在 → nil（不更新）
	
	if receiveSMS, ok := payload["receive_sms"].(bool); ok {
		req.ReceiveSMS = &receiveSMS
	}
	if receiveEmail, ok := payload["receive_email"].(bool); ok {
		req.ReceiveEmail = &receiveEmail
	}

	// 处理 password_hash（已废弃，联系人不登录系统）
	// 规则：passwd 是不回显的，没有从密码改为无密码的状态转换，所以不能发送 ""
	// vue 要么发送有效 password 的 hash，要么不发送该字段，表示 passwd 未修改
	// 如果前端未发送 password_hash 字段，req.PasswordHash 为 nil（不更新）
	if passwordHash, ok := payload["password_hash"].(string); ok && passwordHash != "" {
		// 前端发送有效的 password_hash（hex 字符串）
		req.PasswordHash = &passwordHash
	}
	// password_hash 字段不存在或为空字符串 → req.PasswordHash 为 nil（不更新）

	// 调用 Service 层
	resp, err := h.residentService.UpdateResidentContact(ctx, req)
	if err != nil {
		h.logger.Error("UpdateResidentContact failed",
			zap.String("tenant_id", tenantID),
			zap.String("resident_id", residentID),
			zap.String("slot", slot),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"success": resp.Success,
	}))
}
