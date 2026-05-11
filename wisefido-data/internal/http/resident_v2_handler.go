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

type ResidentV2Handler struct {
	svc    *service.ResidentV2Service
	logger *zap.Logger
	base   *StubHandler
	db     *sql.DB
}

func NewResidentV2Handler(db *sql.DB, svc *service.ResidentV2Service, logger *zap.Logger) *ResidentV2Handler {
	return &ResidentV2Handler{svc: svc, logger: logger, base: &StubHandler{}, db: db}
}

func (h *ResidentV2Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")
	switch {
	case path == "/admin/api/v2/residents" && r.Method == http.MethodGet:
		h.list(w, r)
	case path == "/admin/api/v2/residents" && r.Method == http.MethodPost:
		h.create(w, r)
	case strings.HasPrefix(path, "/admin/api/v2/residents/") && strings.HasSuffix(path, "/clear-check") && r.Method == http.MethodGet:
		hoa := extractV2Hoa(path, "/clear-check")
		h.clearCheck(w, r, hoa)
	case strings.HasPrefix(path, "/admin/api/v2/residents/") && strings.HasSuffix(path, "/discharge") && r.Method == http.MethodPost:
		hoa := extractV2Hoa(path, "/discharge")
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

func extractV2Hoa(path, suffix string) string {
	rest := strings.TrimPrefix(path, "/admin/api/v2/residents/")
	return strings.TrimSuffix(rest, suffix)
}

// ============================================================================

func (h *ResidentV2Handler) list(w http.ResponseWriter, r *http.Request) {
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
		CurrentUserHOA:  currentHOA,
		CurrentUserRole: role,
		Filter: domain.ResidentV2ListFilter{
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

func (h *ResidentV2Handler) get(w http.ResponseWriter, r *http.Request, hoa string) {
	tenantPrefix, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	d, err := h.svc.Get(r.Context(), service.GetResidentV2Request{
		TenantPrefix:    tenantPrefix,
		HoA:             hoa,
		CurrentUserHOA:  r.Header.Get("X-User-HoA"),
		CurrentUserRole: r.Header.Get("X-User-Role"),
	})
	if err != nil {
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(d))
}

func (h *ResidentV2Handler) create(w http.ResponseWriter, r *http.Request) {
	tenantPrefix, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	var input domain.ResidentV2CreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid JSON: "+err.Error()))
		return
	}
	hoa, err := h.svc.Create(r.Context(), service.CreateResidentV2Request{
		TenantPrefix:    tenantPrefix,
		CurrentUserRole: r.Header.Get("X-User-Role"),
		Input:           &input,
	})
	if err != nil {
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(map[string]string{"hoa": hoa}))
}

func (h *ResidentV2Handler) update(w http.ResponseWriter, r *http.Request, hoa string) {
	tenantPrefix, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	var input domain.ResidentV2UpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid JSON: "+err.Error()))
		return
	}
	if err := h.svc.Update(r.Context(), service.UpdateResidentV2Request{
		TenantPrefix:    tenantPrefix,
		HoA:             hoa,
		CurrentUserRole: r.Header.Get("X-User-Role"),
		Input:           &input,
	}); err != nil {
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(map[string]bool{"success": true}))
}

func (h *ResidentV2Handler) delete(w http.ResponseWriter, r *http.Request, hoa string) {
	tenantPrefix, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	hard := r.URL.Query().Get("hard") == "true" || r.URL.Query().Get("clear") == "true"
	if err := h.svc.Delete(r.Context(), service.DeleteResidentV2Request{
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

func (h *ResidentV2Handler) clearCheck(w http.ResponseWriter, r *http.Request, hoa string) {
	res, err := h.svc.CheckClearable(r.Context(), hoa)
	if err != nil {
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(res))
}

func (h *ResidentV2Handler) discharge(w http.ResponseWriter, r *http.Request, hoa string) {
	tenantPrefix, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	if err := h.svc.Discharge(r.Context(), service.UpdateResidentV2Request{
		TenantPrefix:    tenantPrefix,
		HoA:             hoa,
		CurrentUserRole: r.Header.Get("X-User-Role"),
	}); err != nil {
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(map[string]bool{"success": true}))
}
