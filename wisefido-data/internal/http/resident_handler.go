package httpapi

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

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
	case strings.HasSuffix(path, "/account-settings") && r.Method == http.MethodGet:
		residentID := strings.TrimSuffix(path, "/account-settings")
		residentID = strings.TrimPrefix(residentID, "/admin/api/v1/residents/")
		if residentID != "" && !strings.Contains(residentID, "/") {
			h.GetResidentAccountSettings(w, r, residentID)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	// UpdateResidentAccountSettings (必须在 UpdateResident 之前，因为路径更具体)
	case strings.HasSuffix(path, "/account-settings") && r.Method == http.MethodPut:
		residentID := strings.TrimSuffix(path, "/account-settings")
		residentID = strings.TrimPrefix(residentID, "/admin/api/v1/residents/")
		if residentID != "" && !strings.Contains(residentID, "/") {
			h.UpdateResidentAccountSettings(w, r, residentID)
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
	// ResetContactPassword (必须在 ResetResidentPassword 之前，因为路径更具体)
	case strings.HasPrefix(path, "/admin/api/v1/contacts/") && strings.HasSuffix(path, "/reset-password") && r.Method == http.MethodPost:
		// /admin/api/v1/contacts/:contact_id/reset-password
		contactID := strings.TrimSuffix(path, "/reset-password")
		contactID = strings.TrimPrefix(contactID, "/admin/api/v1/contacts/")
		if contactID != "" && !strings.Contains(contactID, "/") {
			h.ResetContactPassword(w, r, contactID)
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
	var permCheck *service.PermissionCheckResult
	var userBranchTag *string
	if currentUserType != "resident" && currentUserType != "family" && currentUserRole != "" && h.db != nil {
		// 获取用户 branch_tag
		var branchTag sql.NullString
		err := h.db.QueryRowContext(ctx,
			`SELECT branch_tag FROM users WHERE tenant_id = $1 AND user_id::text = $2`,
			tenantID, currentUserID,
		).Scan(&branchTag)
		if err == nil && branchTag.Valid {
			tag := branchTag.String
			userBranchTag = &tag
		}

		// 检查权限
		perm, err := GetResourcePermission(h.db, ctx, currentUserRole, "residents", "R")
		if err == nil {
			userBranchTagStr := ""
			if userBranchTag != nil {
				userBranchTagStr = *userBranchTag
			}
			permCheck = &service.PermissionCheckResult{
				AssignedOnly:  perm.AssignedOnly,
				BranchOnly:    perm.BranchOnly,
				UserBranchTag: userBranchTagStr,
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
		if item.FamilyTag != nil {
			itemMap["family_tag"] = *item.FamilyTag
		}
		if item.UnitID != nil {
			itemMap["unit_id"] = *item.UnitID
		}
		if item.UnitName != nil {
			itemMap["unit_name"] = *item.UnitName
		}
		if item.BranchTag != nil {
			itemMap["branch_tag"] = *item.BranchTag
		}
		if item.AreaTag != nil {
			itemMap["area_tag"] = *item.AreaTag
		}
		if item.UnitNumber != nil {
			itemMap["unit_number"] = *item.UnitNumber
		}
		itemMap["is_multi_person_room"] = item.IsMultiPersonRoom
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
	// 注意：对于 resident/family 用户，userType 是 "resident"，role 可能是 "Resident" 或 "Family"
	// 所以需要同时检查 userType 和 role
	var permCheck *service.PermissionCheckResult
	var userBranchTag *string
	isResidentOrFamily := currentUserType == "resident" || currentUserRole == "Resident" || currentUserRole == "Family"
	if !isResidentOrFamily && currentUserRole != "" && h.db != nil {
		var branchTag sql.NullString
		err := h.db.QueryRowContext(ctx,
			`SELECT branch_tag FROM users WHERE tenant_id = $1 AND user_id::text = $2`,
			tenantID, currentUserID,
		).Scan(&branchTag)
		if err == nil && branchTag.Valid {
			tag := branchTag.String
			userBranchTag = &tag
		}

		perm, err := GetResourcePermission(h.db, ctx, currentUserRole, "residents", "R")
		if err == nil {
			userBranchTagStr := ""
			if userBranchTag != nil {
				userBranchTagStr = *userBranchTag
			}
			permCheck = &service.PermissionCheckResult{
				AssignedOnly:  perm.AssignedOnly,
				BranchOnly:    perm.BranchOnly,
				UserBranchTag: userBranchTagStr,
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
	if resp.Resident.FamilyTag != nil {
		item["family_tag"] = *resp.Resident.FamilyTag
	}
	if resp.Resident.UnitID != nil {
		item["unit_id"] = *resp.Resident.UnitID
	}
	if resp.Resident.UnitName != nil {
		item["unit_name"] = *resp.Resident.UnitName
	}
	if resp.Resident.BranchTag != nil {
		item["branch_tag"] = *resp.Resident.BranchTag
	}
	if resp.Resident.AreaTag != nil {
		item["area_tag"] = *resp.Resident.AreaTag
	}
	if resp.Resident.UnitNumber != nil {
		item["unit_number"] = *resp.Resident.UnitNumber
	}
	item["is_multi_person_room"] = resp.Resident.IsMultiPersonRoom
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
			"phi_id": resp.PHI.PhiID,
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
				"contact_id":           c.ContactID,
				"resident_id":          residentID, // 添加 resident_id
				"slot":                 c.Slot,
				"is_enabled":           c.IsEnabled,
				"receive_sms":          c.ReceiveSMS,
				"receive_email":        c.ReceiveEmail,
				"is_emergency_contact": c.IsEmergencyContact,
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
			`SELECT userList, groupList
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
	ctx := r.Context()

	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	currentUserID := r.Header.Get("X-User-Id")
	currentUserType := r.Header.Get("X-User-Type")
	currentUserRole := r.Header.Get("X-User-Role")

	// 权限检查（需要 C 权限）
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

	// 获取用户 branch_tag
	var userBranchTag *string
	if h.db != nil {
		var branchTag sql.NullString
		err := h.db.QueryRowContext(ctx,
			`SELECT branch_tag FROM users WHERE tenant_id = $1 AND user_id::text = $2`,
			tenantID, currentUserID,
		).Scan(&branchTag)
		if err == nil && branchTag.Valid {
			tag := branchTag.String
			userBranchTag = &tag
		}
	}

	// 权限检查结果
	var permCheck *service.PermissionCheckResult
	if currentUserRole != "" && h.db != nil {
		perm, err := GetResourcePermission(h.db, ctx, currentUserRole, "residents", "C")
		if err == nil {
			userBranchTagStr := ""
			if userBranchTag != nil {
				userBranchTagStr = *userBranchTag
			}
			permCheck = &service.PermissionCheckResult{
				AssignedOnly:  perm.AssignedOnly,
				BranchOnly:    perm.BranchOnly,
				UserBranchTag: userBranchTagStr,
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
}

// ============================================
// UpdateResident 更新住户
// ============================================

func (h *ResidentHandler) UpdateResident(w http.ResponseWriter, r *http.Request, residentID string) {
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

	// 构建 Service 请求（权限检查由 Service 层自己处理）
	req := service.UpdateResidentRequest{
		TenantID:        tenantID,
		ResidentID:      residentID,
		CurrentUserID:   currentUserID,
		CurrentUserType: currentUserType,
		CurrentUserRole: currentUserRole,
		// PermissionCheck 不再需要，Service 层自己查询
	}

	// 提取可更新字段
	if residentAccount, ok := payload["resident_account"].(string); ok && residentAccount != "" {
		req.ResidentAccount = &residentAccount
	}
	if nickname, ok := payload["nickname"].(string); ok {
		req.Nickname = &nickname
	}
	if status, ok := payload["status"].(string); ok {
		req.Status = &status
	}
	if serviceLevel, ok := payload["service_level"].(string); ok {
		req.ServiceLevel = &serviceLevel
	}
	if admDate, ok := payload["admission_date"].(string); ok && admDate != "" {
		if t, err := time.Parse("2006-01-02", admDate); err == nil {
			ts := t.Unix()
			req.AdmissionDate = &ts
		}
	}
	if disDate, ok := payload["discharge_date"].(string); ok {
		if disDate != "" {
			if t, err := time.Parse("2006-01-02", disDate); err == nil {
				ts := t.Unix()
				req.DischargeDate = &ts
			}
		} else {
			// 空字符串表示清除
			var zero int64
			req.DischargeDate = &zero
		}
	}
	if unitID, ok := payload["unit_id"].(string); ok {
		req.UnitID = &unitID
	}
	if familyTag, ok := payload["family_tag"].(string); ok {
		req.FamilyTag = &familyTag
	}
	if isAccessEnabled, ok := payload["is_access_enabled"].(bool); ok {
		req.IsAccessEnabled = &isAccessEnabled
	}
	if note, ok := payload["note"].(string); ok {
		req.Note = &note
	}

	// 处理 PHI 更新
	if phiData, ok := payload["phi"].(map[string]any); ok {
		phi := &service.UpdateResidentPHIRequest{}

		// 提取所有 PHI 字段
		if firstName, ok := phiData["first_name"].(string); ok && firstName != "" {
			phi.FirstName = &firstName
		}
		if lastName, ok := phiData["last_name"].(string); ok && lastName != "" {
			phi.LastName = &lastName
		}
		if gender, ok := phiData["gender"].(string); ok && gender != "" {
			phi.Gender = &gender
		}
		if dob, ok := phiData["date_of_birth"].(string); ok && dob != "" {
			if t, err := time.Parse("2006-01-02", dob); err == nil {
				ts := t.Unix()
				phi.DateOfBirth = &ts
			}
		}
		if phone, ok := phiData["resident_phone"].(string); ok && phone != "" {
			phi.ResidentPhone = &phone
		}
		if email, ok := phiData["resident_email"].(string); ok && email != "" {
			phi.ResidentEmail = &email
		}
		if weightLb, ok := phiData["weight_lb"].(float64); ok {
			phi.WeightLb = &weightLb
		}
		if heightFt, ok := phiData["height_ft"].(float64); ok {
			phi.HeightFt = &heightFt
		}
		if heightIn, ok := phiData["height_in"].(float64); ok {
			phi.HeightIn = &heightIn
		}
		if mobilityLevel, ok := phiData["mobility_level"].(float64); ok {
			level := int(mobilityLevel)
			phi.MobilityLevel = &level
		}
		if tremorStatus, ok := phiData["tremor_status"].(string); ok && tremorStatus != "" {
			phi.TremorStatus = &tremorStatus
		}
		if mobilityAid, ok := phiData["mobility_aid"].(string); ok && mobilityAid != "" {
			phi.MobilityAid = &mobilityAid
		}
		if adlAssistance, ok := phiData["adl_assistance"].(string); ok && adlAssistance != "" {
			phi.ADLAssistance = &adlAssistance
		}
		if commStatus, ok := phiData["comm_status"].(string); ok && commStatus != "" {
			phi.CommStatus = &commStatus
		}
		if hasHypertension, ok := phiData["has_hypertension"].(bool); ok {
			phi.HasHypertension = &hasHypertension
		}
		if hasHyperlipaemia, ok := phiData["has_hyperlipaemia"].(bool); ok {
			phi.HasHyperlipaemia = &hasHyperlipaemia
		}
		if hasHyperglycaemia, ok := phiData["has_hyperglycaemia"].(bool); ok {
			phi.HasHyperglycaemia = &hasHyperglycaemia
		}
		if hasStrokeHistory, ok := phiData["has_stroke_history"].(bool); ok {
			phi.HasStrokeHistory = &hasStrokeHistory
		}
		if hasParalysis, ok := phiData["has_paralysis"].(bool); ok {
			phi.HasParalysis = &hasParalysis
		}
		if hasAlzheimer, ok := phiData["has_alzheimer"].(bool); ok {
			phi.HasAlzheimer = &hasAlzheimer
		}
		if medicalHistory, ok := phiData["medical_history"].(string); ok && medicalHistory != "" {
			phi.MedicalHistory = &medicalHistory
		}
		if hisResidentName, ok := phiData["HIS_resident_name"].(string); ok && hisResidentName != "" {
			phi.HISResidentName = &hisResidentName
		}
		if hisAdmissionDate, ok := phiData["HIS_resident_admission_date"].(string); ok && hisAdmissionDate != "" {
			if t, err := time.Parse("2006-01-02", hisAdmissionDate); err == nil {
				ts := t.Unix()
				phi.HISResidentAdmissionDate = &ts
			}
		}
		if hisDischargeDate, ok := phiData["HIS_resident_discharge_date"].(string); ok && hisDischargeDate != "" {
			if t, err := time.Parse("2006-01-02", hisDischargeDate); err == nil {
				ts := t.Unix()
				phi.HISResidentDischargeDate = &ts
			}
		}
		if homeAddressStreet, ok := phiData["home_address_street"].(string); ok && homeAddressStreet != "" {
			phi.HomeAddressStreet = &homeAddressStreet
		}
		if homeAddressCity, ok := phiData["home_address_city"].(string); ok && homeAddressCity != "" {
			phi.HomeAddressCity = &homeAddressCity
		}
		if homeAddressState, ok := phiData["home_address_state"].(string); ok && homeAddressState != "" {
			phi.HomeAddressState = &homeAddressState
		}
		if homeAddressPostalCode, ok := phiData["home_address_postal_code"].(string); ok && homeAddressPostalCode != "" {
			phi.HomeAddressPostalCode = &homeAddressPostalCode
		}
		if plusCode, ok := phiData["plus_code"].(string); ok && plusCode != "" {
			phi.PlusCode = &plusCode
		}

		req.PHI = phi
	}

	// 处理 Caregivers 更新
	if caregivers, ok := payload["caregivers"].(map[string]any); ok {
		cg := &service.UpdateResidentCaregiversRequest{}
		if userList, ok := caregivers["userList"].([]any); ok {
			cg.UserList = make([]string, 0, len(userList))
			for _, uid := range userList {
				if uidStr, ok := uid.(string); ok {
					cg.UserList = append(cg.UserList, uidStr)
				}
			}
		}
		if groupList, ok := caregivers["groupList"].([]any); ok {
			cg.GroupList = make([]string, 0, len(groupList))
			for _, gid := range groupList {
				if gidStr, ok := gid.(string); ok {
					cg.GroupList = append(cg.GroupList, gidStr)
				}
			}
		}
		req.Caregivers = cg
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
		CurrentUserType: currentUserType,
		CurrentUserRole: currentUserRole,
		// PermissionCheck 不再需要，Service 层自己查询
	}

	// 处理 PHI 更新
	if phiData, ok := payload["phi"].(map[string]any); ok {
		phi := &service.UpdateResidentPHIRequest{}

		// 提取所有 PHI 字段
		if firstName, ok := phiData["first_name"].(string); ok {
			phi.FirstName = &firstName
		}
		if lastName, ok := phiData["last_name"].(string); ok {
			phi.LastName = &lastName
		}
		if gender, ok := phiData["gender"].(string); ok {
			phi.Gender = &gender
		}
		if dob, ok := phiData["date_of_birth"].(string); ok && dob != "" {
			if t, err := time.Parse("2006-01-02", dob); err == nil {
				ts := t.Unix()
				phi.DateOfBirth = &ts
			}
		}
		// 处理 resident_phone（可能为 null，用于删除）
		if phoneVal, exists := phiData["resident_phone"]; exists {
			if phoneVal == nil {
				var empty string
				phi.ResidentPhone = &empty // null 表示删除，设置为空字符串
			} else if phone, ok := phoneVal.(string); ok {
				phi.ResidentPhone = &phone
			}
		}
		// 处理 resident_email（可能为 null，用于删除）
		if emailVal, exists := phiData["resident_email"]; exists {
			if emailVal == nil {
				var empty string
				phi.ResidentEmail = &empty // null 表示删除，设置为空字符串
			} else if email, ok := emailVal.(string); ok {
				phi.ResidentEmail = &email
			}
		}
		// 提取 phone_hash 和 email_hash（前端已计算，用于更新 residents 表）
		if phoneHash, ok := phiData["phone_hash"].(string); ok {
			phi.PhoneHash = &phoneHash
		} else if phoneHashVal, exists := phiData["phone_hash"]; exists && phoneHashVal == nil {
			// null 表示删除 hash
			var empty string
			phi.PhoneHash = &empty
		}
		if emailHash, ok := phiData["email_hash"].(string); ok {
			phi.EmailHash = &emailHash
		} else if emailHashVal, exists := phiData["email_hash"]; exists && emailHashVal == nil {
			// null 表示删除 hash
			var empty string
			phi.EmailHash = &empty
		}
		if weightLb, ok := phiData["weight_lb"].(float64); ok {
			phi.WeightLb = &weightLb
		}
		if heightFt, ok := phiData["height_ft"].(float64); ok {
			phi.HeightFt = &heightFt
		}
		if heightIn, ok := phiData["height_in"].(float64); ok {
			phi.HeightIn = &heightIn
		}
		if mobilityLevel, ok := phiData["mobility_level"].(float64); ok {
			level := int(mobilityLevel)
			phi.MobilityLevel = &level
		}
		if tremorStatus, ok := phiData["tremor_status"].(string); ok {
			phi.TremorStatus = &tremorStatus
		}
		if mobilityAid, ok := phiData["mobility_aid"].(string); ok {
			phi.MobilityAid = &mobilityAid
		}
		if adlAssistance, ok := phiData["adl_assistance"].(string); ok {
			phi.ADLAssistance = &adlAssistance
		}
		if commStatus, ok := phiData["comm_status"].(string); ok {
			phi.CommStatus = &commStatus
		}
		if hasHypertension, ok := phiData["has_hypertension"].(bool); ok {
			phi.HasHypertension = &hasHypertension
		}
		if hasHyperlipaemia, ok := phiData["has_hyperlipaemia"].(bool); ok {
			phi.HasHyperlipaemia = &hasHyperlipaemia
		}
		if hasHyperglycaemia, ok := phiData["has_hyperglycaemia"].(bool); ok {
			phi.HasHyperglycaemia = &hasHyperglycaemia
		}
		if hasStrokeHistory, ok := phiData["has_stroke_history"].(bool); ok {
			phi.HasStrokeHistory = &hasStrokeHistory
		}
		if hasParalysis, ok := phiData["has_paralysis"].(bool); ok {
			phi.HasParalysis = &hasParalysis
		}
		if hasAlzheimer, ok := phiData["has_alzheimer"].(bool); ok {
			phi.HasAlzheimer = &hasAlzheimer
		}
		if medicalHistory, ok := phiData["medical_history"].(string); ok {
			phi.MedicalHistory = &medicalHistory
		}
		if homeAddressStreet, ok := phiData["home_address_street"].(string); ok {
			phi.HomeAddressStreet = &homeAddressStreet
		}
		if homeAddressCity, ok := phiData["home_address_city"].(string); ok {
			phi.HomeAddressCity = &homeAddressCity
		}
		if homeAddressState, ok := phiData["home_address_state"].(string); ok {
			phi.HomeAddressState = &homeAddressState
		}
		if homeAddressPostalCode, ok := phiData["home_address_postal_code"].(string); ok {
			phi.HomeAddressPostalCode = &homeAddressPostalCode
		}
		if plusCode, ok := phiData["plus_code"].(string); ok {
			phi.PlusCode = &plusCode
		}

		req.PHI = phi
	} else {
		// 如果没有 phi 字段，直接使用 payload 作为 PHI 数据（前端直接发送字段，不在 phi 对象中）
		phi := &service.UpdateResidentPHIRequest{}
		if firstName, ok := payload["first_name"].(string); ok {
			phi.FirstName = &firstName
		}
		if lastName, ok := payload["last_name"].(string); ok {
			phi.LastName = &lastName
		}
		if gender, ok := payload["gender"].(string); ok {
			phi.Gender = &gender
		}
		if dob, ok := payload["date_of_birth"].(string); ok && dob != "" {
			if t, err := time.Parse("2006-01-02", dob); err == nil {
				ts := t.Unix()
				phi.DateOfBirth = &ts
			}
		}
		// 处理 resident_phone（可能为 null，用于删除）
		if phoneVal, exists := payload["resident_phone"]; exists {
			if phoneVal == nil {
				var empty string
				phi.ResidentPhone = &empty // null 表示删除，设置为空字符串
			} else if phone, ok := phoneVal.(string); ok {
				phi.ResidentPhone = &phone
			}
		}
		// 处理 resident_email（可能为 null，用于删除）
		if emailVal, exists := payload["resident_email"]; exists {
			if emailVal == nil {
				var empty string
				phi.ResidentEmail = &empty // null 表示删除，设置为空字符串
			} else if email, ok := emailVal.(string); ok {
				phi.ResidentEmail = &email
			}
		}
		// 提取 phone_hash 和 email_hash（前端已计算，用于更新 residents 表）
		if phoneHash, ok := payload["phone_hash"].(string); ok {
			phi.PhoneHash = &phoneHash
		} else if phoneHashVal, exists := payload["phone_hash"]; exists && phoneHashVal == nil {
			// null 表示删除 hash
			var empty string
			phi.PhoneHash = &empty
		}
		if emailHash, ok := payload["email_hash"].(string); ok {
			phi.EmailHash = &emailHash
		} else if emailHashVal, exists := payload["email_hash"]; exists && emailHashVal == nil {
			// null 表示删除 hash
			var empty string
			phi.EmailHash = &empty
		}
		if weightLb, ok := payload["weight_lb"].(float64); ok {
			phi.WeightLb = &weightLb
		}
		if heightFt, ok := payload["height_ft"].(float64); ok {
			phi.HeightFt = &heightFt
		}
		if heightIn, ok := payload["height_in"].(float64); ok {
			phi.HeightIn = &heightIn
		}
		if mobilityLevel, ok := payload["mobility_level"].(float64); ok {
			level := int(mobilityLevel)
			phi.MobilityLevel = &level
		}
		if tremorStatus, ok := payload["tremor_status"].(string); ok {
			phi.TremorStatus = &tremorStatus
		}
		if mobilityAid, ok := payload["mobility_aid"].(string); ok {
			phi.MobilityAid = &mobilityAid
		}
		if adlAssistance, ok := payload["adl_assistance"].(string); ok {
			phi.ADLAssistance = &adlAssistance
		}
		if commStatus, ok := payload["comm_status"].(string); ok {
			phi.CommStatus = &commStatus
		}
		if hasHypertension, ok := payload["has_hypertension"].(bool); ok {
			phi.HasHypertension = &hasHypertension
		}
		if hasHyperlipaemia, ok := payload["has_hyperlipaemia"].(bool); ok {
			phi.HasHyperlipaemia = &hasHyperlipaemia
		}
		if hasHyperglycaemia, ok := payload["has_hyperglycaemia"].(bool); ok {
			phi.HasHyperglycaemia = &hasHyperglycaemia
		}
		if hasStrokeHistory, ok := payload["has_stroke_history"].(bool); ok {
			phi.HasStrokeHistory = &hasStrokeHistory
		}
		if hasParalysis, ok := payload["has_paralysis"].(bool); ok {
			phi.HasParalysis = &hasParalysis
		}
		if hasAlzheimer, ok := payload["has_alzheimer"].(bool); ok {
			phi.HasAlzheimer = &hasAlzheimer
		}
		if medicalHistory, ok := payload["medical_history"].(string); ok {
			phi.MedicalHistory = &medicalHistory
		}
		if hisResidentName, ok := payload["HIS_resident_name"].(string); ok {
			phi.HISResidentName = &hisResidentName
		}
		if hisAdmissionDate, ok := payload["HIS_resident_admission_date"].(string); ok && hisAdmissionDate != "" {
			if t, err := time.Parse("2006-01-02", hisAdmissionDate); err == nil {
				ts := t.Unix()
				phi.HISResidentAdmissionDate = &ts
			}
		}
		if hisDischargeDate, ok := payload["HIS_resident_discharge_date"].(string); ok && hisDischargeDate != "" {
			if t, err := time.Parse("2006-01-02", hisDischargeDate); err == nil {
				ts := t.Unix()
				phi.HISResidentDischargeDate = &ts
			}
		}
		if homeAddressStreet, ok := payload["home_address_street"].(string); ok {
			phi.HomeAddressStreet = &homeAddressStreet
		}
		if homeAddressCity, ok := payload["home_address_city"].(string); ok {
			phi.HomeAddressCity = &homeAddressCity
		}
		if homeAddressState, ok := payload["home_address_state"].(string); ok {
			phi.HomeAddressState = &homeAddressState
		}
		if homeAddressPostalCode, ok := payload["home_address_postal_code"].(string); ok {
			phi.HomeAddressPostalCode = &homeAddressPostalCode
		}
		if plusCode, ok := payload["plus_code"].(string); ok {
			phi.PlusCode = &plusCode
		}
		req.PHI = phi
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

	// 获取用户 branch_tag
	var userBranchTag *string
	if h.db != nil {
		var branchTag sql.NullString
		err := h.db.QueryRowContext(ctx,
			`SELECT branch_tag FROM users WHERE tenant_id = $1 AND user_id::text = $2`,
			tenantID, currentUserID,
		).Scan(&branchTag)
		if err == nil && branchTag.Valid {
			tag := branchTag.String
			userBranchTag = &tag
		}
	}

	// 权限检查结果
	var permCheck *service.PermissionCheckResult
	if currentUserRole != "" && h.db != nil {
		perm, err := GetResourcePermission(h.db, ctx, currentUserRole, "residents", "D")
		if err == nil {
			userBranchTagStr := ""
			if userBranchTag != nil {
				userBranchTagStr = *userBranchTag
			}
			permCheck = &service.PermissionCheckResult{
				AssignedOnly:  perm.AssignedOnly,
				BranchOnly:    perm.BranchOnly,
				UserBranchTag: userBranchTagStr,
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

	// 获取用户 branch_tag
	var userBranchTag *string
	if h.db != nil {
		var branchTag sql.NullString
		err := h.db.QueryRowContext(ctx,
			`SELECT branch_tag FROM users WHERE tenant_id = $1 AND user_id::text = $2`,
			tenantID, currentUserID,
		).Scan(&branchTag)
		if err == nil && branchTag.Valid {
			tag := branchTag.String
			userBranchTag = &tag
		}
	}

	// 权限检查结果
	var permCheck *service.PermissionCheckResult
	if currentUserRole != "" && h.db != nil {
		perm, err := GetResourcePermission(h.db, ctx, currentUserRole, "residents", "U")
		if err == nil {
			userBranchTagStr := ""
			if userBranchTag != nil {
				userBranchTagStr = *userBranchTag
			}
			permCheck = &service.PermissionCheckResult{
				AssignedOnly:  perm.AssignedOnly,
				BranchOnly:    perm.BranchOnly,
				UserBranchTag: userBranchTagStr,
			}
		}
	}

	req := service.ResetResidentPasswordRequest{
		TenantID:        tenantID,
		ResidentID:      residentID,
		CurrentUserID:   currentUserID,
		CurrentUserType: currentUserType,
		CurrentUserRole: currentUserRole,
		UserBranchTag:   userBranchTag,
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
// ResetContactPassword 重置联系人密码（通过 contact_id）
// ============================================

func (h *ResidentHandler) ResetContactPassword(w http.ResponseWriter, r *http.Request, contactID string) {
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

	// 获取用户 branch_tag（如果数据库可用）
	var userBranchTag *string
	if h.db != nil {
		var branchTag sql.NullString
		err := h.db.QueryRowContext(ctx,
			`SELECT branch_tag FROM users WHERE tenant_id = $1 AND user_id::text = $2`,
			tenantID, currentUserID,
		).Scan(&branchTag)
		if err == nil && branchTag.Valid {
			tag := branchTag.String
			userBranchTag = &tag
		}
	}

	// 权限检查结果（如果数据库可用）
	var permCheck *service.PermissionCheckResult
	if currentUserRole != "" && h.db != nil {
		perm, err := GetResourcePermission(h.db, ctx, currentUserRole, "resident_contacts", "U")
		if err == nil {
			userBranchTagStr := ""
			if userBranchTag != nil {
				userBranchTagStr = *userBranchTag
			}
			permCheck = &service.PermissionCheckResult{
				AssignedOnly:  perm.AssignedOnly,
				BranchOnly:    perm.BranchOnly,
				UserBranchTag: userBranchTagStr,
			}
		}
	}

	req := service.ResetContactPasswordRequest{
		TenantID:        tenantID,
		ContactID:       contactID,
		CurrentUserID:   currentUserID,
		CurrentUserType: currentUserType,
		CurrentUserRole: currentUserRole,
		UserBranchTag:   userBranchTag,
		PermissionCheck: permCheck,
		NewPassword:     passwordHash,
	}

	resp, err := h.residentService.ResetContactPassword(ctx, req)
	if err != nil {
		h.logger.Error("ResetContactPassword failed",
			zap.String("tenant_id", tenantID),
			zap.String("contact_id", contactID),
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

	resp, err := h.residentService.GetResidentAccountSettings(ctx, req)
	if err != nil {
		h.logger.Error("GetResidentAccountSettings failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	item := map[string]any{
		"nickname":   resp.Nickname,
		"is_contact": resp.IsContact,
	}
	if resp.ResidentAccount != nil {
		item["resident_account"] = *resp.ResidentAccount
	}
	if resp.Email != nil {
		item["email"] = *resp.Email
	}
	if resp.Phone != nil {
		item["phone"] = *resp.Phone
	}
	// 仅 resident 需要返回 save 标志
	if !resp.IsContact {
		item["save_email"] = resp.SaveEmail
		item["save_phone"] = resp.SavePhone
	}

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
	req := service.UpdateResidentContactRequest{
		TenantID:        tenantID,
		ResidentID:      residentID,
		Slot:            slot, // slot 是必填的，用于定位 contact
		CurrentUserID:   currentUserID,
		CurrentUserType: currentUserType,
		CurrentUserRole: currentUserRole,
		// PermissionCheck 不再需要，Service 层自己查询
	}

	// 解析字段（使用指针表示可选）
	// 规则：
	//   - 字段不存在或为 null → nil（不更新）
	//   - 字段为 "" → ""（更新为空，Repository 会转换为 NULL）
	//   - 字段有值 → 有值（更新）
	if isEnabled, ok := payload["is_enabled"].(bool); ok {
		req.IsEnabled = &isEnabled
	}
	if firstName, ok := payload["contact_first_name"].(string); ok {
		req.ContactFirstName = &firstName // "" 表示删除
	}
	// contact_first_name 不存在或为 null → nil（不更新）
	if lastName, ok := payload["contact_last_name"].(string); ok {
		req.ContactLastName = &lastName // "" 表示删除
	}
	// contact_last_name 不存在或为 null → nil（不更新）
	if relationship, ok := payload["relationship"].(string); ok {
		req.Relationship = &relationship // "" 表示删除
	}
	// relationship 不存在或为 null → nil（不更新）
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
	if receiveSMS, ok := payload["receive_sms"].(bool); ok {
		req.ReceiveSMS = &receiveSMS
	}
	if receiveEmail, ok := payload["receive_email"].(bool); ok {
		req.ReceiveEmail = &receiveEmail
	}

	// 处理 password_hash
	// 规则：passwd 是不回显的，没有从密码改为无密码的状态转换，所以不能发送 ""
	// vue 要么发送有效 password 的 hash，要么不发送该字段，表示 passwd 未修改
	// 如果前端未发送 password_hash 字段，req.PasswordHash 为 nil（不更新）
	if passwordHash, ok := payload["password_hash"].(string); ok && passwordHash != "" {
		// 前端发送有效的 password_hash（hex 字符串）
		req.PasswordHash = &passwordHash
	}
	// password_hash 字段不存在或为空字符串 → req.PasswordHash 为 nil（不更新）

	// 处理 email_hash：支持 string 和 null
	if emailHash, ok := payload["email_hash"].(string); ok {
		req.EmailHash = &emailHash // "" 表示删除
	} else if payload["email_hash"] == nil {
		// Vue 发送 null 时，转换为 ""（删除）
		emptyStr := ""
		req.EmailHash = &emptyStr
	}
	// email_hash 字段不存在 → nil（不更新）

	// 处理 phone_hash：支持 string 和 null
	if phoneHash, ok := payload["phone_hash"].(string); ok {
		req.PhoneHash = &phoneHash // "" 表示删除
	} else if payload["phone_hash"] == nil {
		// Vue 发送 null 时，转换为 ""（删除）
		emptyStr := ""
		req.PhoneHash = &emptyStr
	}
	// phone_hash 字段不存在 → nil（不更新）

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
