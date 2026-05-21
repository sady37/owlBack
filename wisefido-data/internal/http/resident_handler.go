// Package httpapi — Resident V2 Handler (Forward Design)
//
// REST: /admin/api/v2/residents[/{hoa}[/clear-check]]
//
// 不引用 v1 ResidentHandler；session role 通过 X-User-Role header（auth_v2_handler 注入）。
package httpapi

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/service"

	"go.uber.org/zap"
)

type ResidentHandler struct {
	svc    *service.ResidentService
	logger *zap.Logger
	base   *StubHandler
	db     *sql.DB
}

func NewResidentHandler(db *sql.DB, svc *service.ResidentService, logger *zap.Logger) *ResidentHandler {
	return &ResidentHandler{svc: svc, logger: logger, base: &StubHandler{}, db: db}
}

func (h *ResidentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")
	switch {
	case path == "/admin/api/v2/residents" && r.Method == http.MethodGet:
		h.list(w, r)
	case path == "/admin/api/v2/residents" && r.Method == http.MethodPost:
		h.create(w, r)
	case strings.HasPrefix(path, "/admin/api/v2/residents/") && strings.HasSuffix(path, "/clear-check") && r.Method == http.MethodGet:
		hoa, err := h.resolveResidentHoA(extractHoa(path, "/clear-check"))
		if err != nil {
			writeJSON(w, http.StatusOK, Fail(err.Error()))
			return
		}
		h.clearCheck(w, r, hoa)
	case strings.HasPrefix(path, "/admin/api/v2/residents/") && strings.HasSuffix(path, "/discharge") && r.Method == http.MethodPost:
		hoa, err := h.resolveResidentHoA(extractHoa(path, "/discharge"))
		if err != nil {
			writeJSON(w, http.StatusOK, Fail(err.Error()))
			return
		}
		h.discharge(w, r, hoa)
	// P0 subresources — thin wrappers 折叠到 unified ResidentUpdateInput；
	// 同样的角色 gate（canEditResident: Admin/Manager/Nurse），admission/discharge 等敏感字段
	// 不会经由这里（updateInput 里相关字段为 nil）。FE 由 v1 路径迁过来。
	case strings.HasPrefix(path, "/admin/api/v2/residents/") && strings.HasSuffix(path, "/phi") && r.Method == http.MethodPut:
		hoa, err := h.resolveResidentHoA(extractHoa(path, "/phi"))
		if err != nil {
			writeJSON(w, http.StatusOK, Fail(err.Error()))
			return
		}
		h.updatePHI(w, r, hoa)
	case strings.HasPrefix(path, "/admin/api/v2/residents/") && strings.HasSuffix(path, "/contacts") && r.Method == http.MethodPut:
		hoa, err := h.resolveResidentHoA(extractHoa(path, "/contacts"))
		if err != nil {
			writeJSON(w, http.StatusOK, Fail(err.Error()))
			return
		}
		h.updateContacts(w, r, hoa)
	case strings.HasPrefix(path, "/admin/api/v2/residents/") && strings.HasSuffix(path, "/unit-history") && r.Method == http.MethodPost:
		hoa, err := h.resolveResidentHoA(extractHoa(path, "/unit-history"))
		if err != nil {
			writeJSON(w, http.StatusOK, Fail(err.Error()))
			return
		}
		h.assignUnit(w, r, hoa)
	case strings.HasPrefix(path, "/admin/api/v2/residents/") && strings.HasSuffix(path, "/caregivers") && r.Method == http.MethodPost:
		hoa, err := h.resolveResidentHoA(extractHoa(path, "/caregivers"))
		if err != nil {
			writeJSON(w, http.StatusOK, Fail(err.Error()))
			return
		}
		h.assignCaregivers(w, r, hoa)
	case strings.HasPrefix(path, "/admin/api/v2/residents/") && r.Method == http.MethodGet:
		hoa, err := h.resolveResidentHoA(strings.TrimPrefix(path, "/admin/api/v2/residents/"))
		if err != nil {
			writeJSON(w, http.StatusOK, Fail(err.Error()))
			return
		}
		h.get(w, r, hoa)
	case strings.HasPrefix(path, "/admin/api/v2/residents/") && r.Method == http.MethodPut:
		hoa, err := h.resolveResidentHoA(strings.TrimPrefix(path, "/admin/api/v2/residents/"))
		if err != nil {
			writeJSON(w, http.StatusOK, Fail(err.Error()))
			return
		}
		h.update(w, r, hoa)
	case strings.HasPrefix(path, "/admin/api/v2/residents/") && r.Method == http.MethodDelete:
		hoa, err := h.resolveResidentHoA(strings.TrimPrefix(path, "/admin/api/v2/residents/"))
		if err != nil {
			writeJSON(w, http.StatusOK, Fail(err.Error()))
			return
		}
		h.delete(w, r, hoa)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func extractHoa(path, suffix string) string {
	rest := strings.TrimPrefix(path, "/admin/api/v2/residents/")
	return strings.TrimSuffix(rest, suffix)
}

// resolveResidentHoA — path 段可能是：
//
//  1. IPv6 CIDR /128 (含 ":" 或 "/"，如 "fd00:0:3:111:1:ff01:10:0/128") — 直接当 hoa
//  2. dns_short_name 6 位 base36 (a-z0-9，无 ":/")  — 查 residents.dns_short_name 反解为 hoa
//
// 短码查不到 → error；非短码格式 + 非 IPv6 → 原样返回（让下游 INET cast 报错）。
//
// 同 cards.card_dns resolve 模式（[[short-code-alias-resolve-everywhere]]）：
// 必须在 handler 入口 resolve，service 内部一律假设拿到的是 hoa。
func (h *ResidentHandler) resolveResidentHoA(idOrShort string) (string, error) {
	if idOrShort == "" {
		return "", nil
	}
	if strings.ContainsAny(idOrShort, ":/") {
		return idOrShort, nil
	}
	if !isResidentShortCodeFormat(idOrShort) {
		return idOrShort, nil
	}
	var hoa string
	err := h.db.QueryRow(
		`SELECT host(resident_id) || '/' || masklen(resident_id) FROM residents WHERE dns_short_name = $1`,
		idOrShort,
	).Scan(&hoa)
	if err == sql.ErrNoRows {
		return "", &residentNotFoundErr{shortName: idOrShort}
	}
	if err != nil {
		return "", err
	}
	return hoa, nil
}

func isResidentShortCodeFormat(s string) bool {
	if len(s) != 6 {
		return false
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return true
}

type residentNotFoundErr struct{ shortName string }

func (e *residentNotFoundErr) Error() string {
	return "resident not found (short_name=" + e.shortName + ")"
}

// ============================================================================

func (h *ResidentHandler) list(w http.ResponseWriter, r *http.Request) {
	tenantPrefix, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	role := r.Header.Get("X-User-Role")
	currentHOA := r.Header.Get("X-User-HoA") // 可空；Resident 自查需要

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	includeDel := r.URL.Query().Get("include_deleted") == "true"

	resp, err := h.svc.List(r.Context(), service.ListResidentsV2Request{
		TenantPrefix:    tenantPrefix,
		CurrentUserID:   r.Header.Get("X-User-Id"),
		CurrentUserHOA:  currentHOA,
		CurrentUserRole: role,
		Filter: domain.ResidentListFilter{
			Search:        r.URL.Query().Get("search"),
			Status:        r.URL.Query().Get("status"),
			IncludeDelete: includeDel,
			Page:          page,
			PageSize:      size,
		},
	})
	if err != nil {
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(resp))
}

func (h *ResidentHandler) get(w http.ResponseWriter, r *http.Request, hoa string) {
	tenantPrefix, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	h.logger.Info("[ResidentHandler.get ENTRY]", zap.String("hoa", hoa), zap.String("tenant_prefix", tenantPrefix))
	d, err := h.svc.Get(r.Context(), service.GetResidentRequest{
		TenantPrefix:    tenantPrefix,
		HoA:             hoa,
		CurrentUserID:   r.Header.Get("X-User-Id"),
		CurrentUserHOA:  r.Header.Get("X-User-HoA"),
		CurrentUserRole: r.Header.Get("X-User-Role"),
	})
	if err != nil {
		h.logger.Error("[ResidentHandler.get FAILED]", zap.String("hoa", hoa), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	caregiverCount := 0
	teamCount := 0
	familyCount := 0
	caregiverDetails := make([]map[string]string, 0)
	teamDetails := make([]map[string]string, 0)
	familyDetails := make([]map[string]string, 0)
	if d != nil {
		if len(d.Caregivers) > 0 {
			caregiverCount = len(d.Caregivers)
			for _, c := range d.Caregivers {
				caregiverDetails = append(caregiverDetails, map[string]string{
					"user_id":  c.UserID,
					"nickname": c.Nickname,
					"account":  c.UserAccount,
				})
			}
		}
		if len(d.Teams) > 0 {
			teamCount = len(d.Teams)
			for _, t := range d.Teams {
				teamDetails = append(teamDetails, map[string]string{
					"team_id":   t.TeamID,
					"team_name": t.TeamName,
					"team_kind": t.TeamKind,
				})
			}
		}
		if len(d.Family) > 0 {
			familyCount = len(d.Family)
			for _, f := range d.Family {
				familyDetails = append(familyDetails, map[string]string{
					"user_id":  f.UserID,
					"nickname": f.Nickname,
					"account":  f.UserAccount,
				})
			}
		}
	}
	// HIPAA min-necessary：Caregiver / Viewer / Family 等非临床角色不返回 PHI / Contacts。
	// FE 同时也 hide PHI / Contacts tab；BE 这层是 defense-in-depth，防 curl 直接拉。
	// Manager / Admin / Nurse / Resident（看自己）/ SystemAdmin（跨 tenant 只读）正常返回。
	if d != nil {
		switch r.Header.Get("X-User-Role") {
		case "Caregiver", "Viewer":
			d.PHI = nil
			d.Contacts = nil
		}
	}

	h.logger.Info("[ResidentHandler.get SUCCESS]",
		zap.String("hoa", hoa),
		zap.Int("caregiver_count", caregiverCount),
		zap.Int("team_count", teamCount),
		zap.Int("family_count", familyCount),
		zap.Any("caregivers", caregiverDetails),
		zap.Any("teams", teamDetails),
		zap.Any("family", familyDetails))
	writeJSON(w, http.StatusOK, Ok(d))
}

func (h *ResidentHandler) create(w http.ResponseWriter, r *http.Request) {
	tenantPrefix, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	var input domain.ResidentCreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid JSON: "+err.Error()))
		return
	}
	hoa, err := h.svc.Create(r.Context(), service.CreateResidentRequest{
		TenantPrefix:    tenantPrefix,
		ActorUserID:     r.Header.Get("X-User-Id"),
		CurrentUserRole: r.Header.Get("X-User-Role"),
		Input:           &input,
	})
	if err != nil {
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(map[string]string{"hoa": hoa}))
}

func (h *ResidentHandler) update(w http.ResponseWriter, r *http.Request, hoa string) {
	tenantPrefix, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	var input domain.ResidentUpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Error("[ResidentHandler.update] decode failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("invalid JSON: "+err.Error()))
		return
	}
	caregiverCount := 0
	var caregiverIDs []string
	if input.CaregiverUserIDs != nil {
		caregiverCount = len(*input.CaregiverUserIDs)
		caregiverIDs = *input.CaregiverUserIDs
	}
	familyCount := -1
	var familyIDs []string
	if input.FamilyUserIDs != nil {
		familyCount = len(*input.FamilyUserIDs)
		familyIDs = *input.FamilyUserIDs
	}
	teamCount := -1
	var teamIDs []string
	if input.CareTeamIDs != nil {
		teamCount = len(*input.CareTeamIDs)
		teamIDs = *input.CareTeamIDs
	}
	h.logger.Info("[ResidentHandler.update ENTRY]",
		zap.String("hoa", hoa),
		zap.String("tenant_prefix", tenantPrefix),
		zap.String("role", r.Header.Get("X-User-Role")),
		zap.Int("caregiver_count", caregiverCount),
		zap.Int("family_count", familyCount),
		zap.Int("team_count", teamCount),
		zap.Strings("caregiver_ids", caregiverIDs),
		zap.Strings("family_ids", familyIDs),
		zap.Strings("team_ids", teamIDs))
	if err := h.svc.Update(r.Context(), service.UpdateResidentRequest{
		TenantPrefix:    tenantPrefix,
		HoA:             hoa,
		ActorUserID:     r.Header.Get("X-User-Id"),
		CurrentUserRole: r.Header.Get("X-User-Role"),
		Input:           &input,
	}); err != nil {
		h.logger.Error("[ResidentHandler.update FAILED]", zap.String("hoa", hoa), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	h.logger.Info("[ResidentHandler.update SUCCESS]", zap.String("hoa", hoa))
	writeJSON(w, http.StatusOK, Ok(map[string]bool{"success": true}))
}

func (h *ResidentHandler) delete(w http.ResponseWriter, r *http.Request, hoa string) {
	tenantPrefix, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	hard := r.URL.Query().Get("hard") == "true" || r.URL.Query().Get("clear") == "true"
	if err := h.svc.Delete(r.Context(), service.DeleteResidentRequest{
		TenantPrefix:    tenantPrefix,
		HoA:             hoa,
		CurrentUserRole: r.Header.Get("X-User-Role"),
		Hard:            hard,
	}); err != nil {
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(map[string]bool{"success": true}))
}

func (h *ResidentHandler) clearCheck(w http.ResponseWriter, r *http.Request, hoa string) {
	res, err := h.svc.CheckClearable(r.Context(), hoa)
	if err != nil {
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(res))
}

func (h *ResidentHandler) discharge(w http.ResponseWriter, r *http.Request, hoa string) {
	tenantPrefix, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	if err := h.svc.Discharge(r.Context(), service.UpdateResidentRequest{
		TenantPrefix:    tenantPrefix,
		HoA:             hoa,
		ActorUserID:     r.Header.Get("X-User-Id"),
		CurrentUserRole: r.Header.Get("X-User-Role"),
	}); err != nil {
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(map[string]bool{"success": true}))
}

// callUpdate — 共享底座：wrapper 子路由把局部 payload 折成完整 ResidentUpdateInput 后走同一个 service.Update。
// 共用 admission/discharge gate（仅这两个字段在 wrapper 里**永不**赋值，自然不会触发 nurse 的 finance 边界拒绝）。
func (h *ResidentHandler) callUpdate(w http.ResponseWriter, r *http.Request, hoa string, in *domain.ResidentUpdateInput) {
	tenantPrefix, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	if err := h.svc.Update(r.Context(), service.UpdateResidentRequest{
		TenantPrefix:    tenantPrefix,
		HoA:             hoa,
		ActorUserID:     r.Header.Get("X-User-Id"),
		CurrentUserRole: r.Header.Get("X-User-Role"),
		Input:           in,
	}); err != nil {
		h.logger.Error("[ResidentHandler.subresource update FAILED]",
			zap.String("hoa", hoa),
			zap.String("path", r.URL.Path),
			zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(map[string]bool{"success": true}))
}

// updatePHI — PUT /admin/api/v2/residents/{hoa}/phi
// body = ResidentPHIv2 (flat, FE Partial<ResidentPHIv2>)；折成 UpdateInput{PHI: &body} 走 service.Update。
func (h *ResidentHandler) updatePHI(w http.ResponseWriter, r *http.Request, hoa string) {
	var phi domain.ResidentPHIv2
	if err := json.NewDecoder(r.Body).Decode(&phi); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid JSON: "+err.Error()))
		return
	}
	h.callUpdate(w, r, hoa, &domain.ResidentUpdateInput{PHI: &phi})
}

// updateContacts — PUT /admin/api/v2/residents/{hoa}/contacts
// body = []ResidentContactV2 (替换 resident 的全部紧急联系人；FE 传整个数组)。
// 若 FE 单条改：先 GET 拿完整 contacts → 改本地数组 → PUT 整组。
func (h *ResidentHandler) updateContacts(w http.ResponseWriter, r *http.Request, hoa string) {
	var contacts []domain.ResidentContactV2
	if err := json.NewDecoder(r.Body).Decode(&contacts); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid JSON: "+err.Error()))
		return
	}
	h.callUpdate(w, r, hoa, &domain.ResidentUpdateInput{Contacts: &contacts})
}

// assignUnit — POST /admin/api/v2/residents/{hoa}/unit-history
// body = {unit_id?, room_id?, bed_id?}：unit_id 必填或显式空字符串（解绑）；room_id/bed_id 可选下钻。
// repo.UpdateResident 里实际转化为 resident_unit 表的 valid_from/valid_to 切换（中途换 unit 自动 close 旧行新增新行）。
func (h *ResidentHandler) assignUnit(w http.ResponseWriter, r *http.Request, hoa string) {
	var body struct {
		UnitID *string `json:"unit_id"`
		RoomID *string `json:"room_id"`
		BedID  *string `json:"bed_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid JSON: "+err.Error()))
		return
	}
	h.callUpdate(w, r, hoa, &domain.ResidentUpdateInput{
		UnitID: body.UnitID,
		RoomID: body.RoomID,
		BedID:  body.BedID,
	})
}

// assignCaregivers — POST /admin/api/v2/residents/{hoa}/caregivers
// body = {caregiver_user_ids?, care_team_ids?, family_user_ids?}：任一字段提供即重置该 slice；空数组 = 显式清空。
func (h *ResidentHandler) assignCaregivers(w http.ResponseWriter, r *http.Request, hoa string) {
	var body struct {
		CaregiverUserIDs *[]string `json:"caregiver_user_ids"`
		CareTeamIDs      *[]string `json:"care_team_ids"`
		FamilyUserIDs    *[]string `json:"family_user_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid JSON: "+err.Error()))
		return
	}
	h.callUpdate(w, r, hoa, &domain.ResidentUpdateInput{
		CaregiverUserIDs: body.CaregiverUserIDs,
		CareTeamIDs:      body.CareTeamIDs,
		FamilyUserIDs:    body.FamilyUserIDs,
	})
}
