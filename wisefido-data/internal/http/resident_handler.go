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
		hoa := extractHoa(path, "/clear-check")
		h.clearCheck(w, r, hoa)
	case strings.HasPrefix(path, "/admin/api/v2/residents/") && strings.HasSuffix(path, "/discharge") && r.Method == http.MethodPost:
		hoa := extractHoa(path, "/discharge")
		h.discharge(w, r, hoa)
	case strings.HasPrefix(path, "/admin/api/v2/residents/") && r.Method == http.MethodGet:
		hoa := strings.TrimPrefix(path, "/admin/api/v2/residents/")
		h.get(w, r, hoa)
	case strings.HasPrefix(path, "/admin/api/v2/residents/") && r.Method == http.MethodPut:
		hoa := strings.TrimPrefix(path, "/admin/api/v2/residents/")
		h.update(w, r, hoa)
	case strings.HasPrefix(path, "/admin/api/v2/residents/") && r.Method == http.MethodDelete:
		hoa := strings.TrimPrefix(path, "/admin/api/v2/residents/")
		h.delete(w, r, hoa)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func extractHoa(path, suffix string) string {
	rest := strings.TrimPrefix(path, "/admin/api/v2/residents/")
	return strings.TrimSuffix(rest, suffix)
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
