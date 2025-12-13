package httpapi

import (
	"database/sql"
	"net/http"
	"strings"
)

// -------- tenant helpers --------

func (a *AdminAPI) tenantIDFromReq(w http.ResponseWriter, r *http.Request) (string, bool) {
	if tid := r.URL.Query().Get("tenant_id"); tid != "" {
		return tid, true
	}
	// Prefer tenant header (owlFront axios injects it for all requests after login)
	if tid := r.Header.Get("X-Tenant-Id"); tid != "" && tid != "null" {
		return tid, true
	}
	// Try to resolve tenant from user ID
	if a.Tenant != nil {
		userID := r.Header.Get("X-User-Id")
		if userID != "" {
			if tid, err := a.Tenant.TenantIDByUserID(r.Context(), userID); err == nil && tid != "" {
				return tid, true
			}
		}
	}
	// Convenience: SystemAdmin without tenant header falls back to System tenant
	if strings.EqualFold(r.Header.Get("X-User-Role"), "SystemAdmin") {
		// SystemTenantID is defined in stub_handlers.go
		return "00000000-0000-0000-0000-000000000001", true
	}
	writeJSON(w, http.StatusOK, Fail("tenant_id is required"))
	return "", false
}

func (a *AdminAPI) getBuildings(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := a.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	branchTag := r.URL.Query().Get("branch_tag")
	items, err := a.Units.ListBuildings(r.Context(), tenantID, branchTag)
	if err != nil {
		writeJSON(w, http.StatusOK, Fail("failed to list buildings"))
		return
	}
	writeJSON(w, http.StatusOK, Ok(items))
}

// -------- Units impl --------

func (a *AdminAPI) getUnits(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := a.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	filters := map[string]string{
		"branch_tag":  r.URL.Query().Get("branch_tag"),
		"building":    r.URL.Query().Get("building"),
		"floor":       r.URL.Query().Get("floor"),
		"area_tag":    r.URL.Query().Get("area_tag"),
		"unit_number": r.URL.Query().Get("unit_number"),
		"unit_name":   r.URL.Query().Get("unit_name"),
		"unit_type":   r.URL.Query().Get("unit_type"),
	}
	page := parseInt(r.URL.Query().Get("page"), 1)
	size := parseInt(r.URL.Query().Get("size"), 100)

	items, total, err := a.Units.ListUnits(r.Context(), tenantID, filters, page, size)
	if err != nil {
		writeJSON(w, http.StatusOK, Fail("failed to list units"))
		return
	}
	out := make([]any, 0, len(items))
	for _, u := range items {
		out = append(out, u.ToJSON())
	}
	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"items": out,
		"total": total,
	}))
}

func (a *AdminAPI) getUnitDetail(w http.ResponseWriter, r *http.Request, unitID string) {
	tenantID, ok := a.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	u, err := a.Units.GetUnit(r.Context(), tenantID, unitID)
	if err != nil {
		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusOK, Fail("unit not found"))
			return
		}
		writeJSON(w, http.StatusOK, Fail("failed to get unit"))
		return
	}
	writeJSON(w, http.StatusOK, Ok(u.ToJSON()))
}

func (a *AdminAPI) createUnit(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := a.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}
	u, err := a.Units.CreateUnit(r.Context(), tenantID, payload)
	if err != nil {
		writeJSON(w, http.StatusOK, Fail("failed to create unit"))
		return
	}
	writeJSON(w, http.StatusOK, Ok(u.ToJSON()))
}

func (a *AdminAPI) updateUnit(w http.ResponseWriter, r *http.Request, unitID string) {
	tenantID, ok := a.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}
	u, err := a.Units.UpdateUnit(r.Context(), tenantID, unitID, payload)
	if err != nil {
		writeJSON(w, http.StatusOK, Fail("failed to update unit"))
		return
	}
	writeJSON(w, http.StatusOK, Ok(u.ToJSON()))
}

func (a *AdminAPI) deleteUnit(w http.ResponseWriter, r *http.Request, unitID string) {
	tenantID, ok := a.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	if err := a.Units.DeleteUnit(r.Context(), tenantID, unitID); err != nil {
		writeJSON(w, http.StatusOK, Fail("failed to delete unit"))
		return
	}
	writeJSON(w, http.StatusOK, Ok[any](nil))
}

func (a *AdminAPI) getRoomsWithBeds(w http.ResponseWriter, r *http.Request) {
	unitID := r.URL.Query().Get("unit_id")
	if unitID == "" {
		writeJSON(w, http.StatusOK, Fail("unit_id is required"))
		return
	}
	out, err := a.Units.ListRoomsWithBeds(r.Context(), unitID)
	if err != nil {
		writeJSON(w, http.StatusOK, Fail("failed to list rooms"))
		return
	}
	writeJSON(w, http.StatusOK, Ok(out))
}

func (a *AdminAPI) createRoom(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}
	unitID, _ := payload["unit_id"].(string)
	if unitID == "" {
		writeJSON(w, http.StatusOK, Fail("unit_id is required"))
		return
	}
	rr, err := a.Units.CreateRoom(r.Context(), unitID, payload)
	if err != nil {
		writeJSON(w, http.StatusOK, Fail("failed to create room"))
		return
	}
	writeJSON(w, http.StatusOK, Ok(rr.ToJSON()))
}

func (a *AdminAPI) updateRoom(w http.ResponseWriter, r *http.Request, roomID string) {
	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}
	rr, err := a.Units.UpdateRoom(r.Context(), roomID, payload)
	if err != nil {
		writeJSON(w, http.StatusOK, Fail("failed to update room"))
		return
	}
	writeJSON(w, http.StatusOK, Ok(rr.ToJSON()))
}

func (a *AdminAPI) deleteRoom(w http.ResponseWriter, r *http.Request, roomID string) {
	if err := a.Units.DeleteRoom(r.Context(), roomID); err != nil {
		writeJSON(w, http.StatusOK, Fail("failed to delete room"))
		return
	}
	writeJSON(w, http.StatusOK, Ok[any](nil))
}

func (a *AdminAPI) getBeds(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room_id")
	if roomID == "" {
		writeJSON(w, http.StatusOK, Fail("room_id is required"))
		return
	}
	beds, err := a.Units.ListBeds(r.Context(), roomID)
	if err != nil {
		writeJSON(w, http.StatusOK, Fail("failed to list beds"))
		return
	}
	out := make([]any, 0, len(beds))
	for _, b := range beds {
		out = append(out, b.ToJSON())
	}
	writeJSON(w, http.StatusOK, Ok(out))
}

func (a *AdminAPI) createBed(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}
	roomID, _ := payload["room_id"].(string)
	if roomID == "" {
		writeJSON(w, http.StatusOK, Fail("room_id is required"))
		return
	}
	b, err := a.Units.CreateBed(r.Context(), roomID, payload)
	if err != nil {
		writeJSON(w, http.StatusOK, Fail("failed to create bed"))
		return
	}
	writeJSON(w, http.StatusOK, Ok(b.ToJSON()))
}

func (a *AdminAPI) updateBed(w http.ResponseWriter, r *http.Request, bedID string) {
	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}
	b, err := a.Units.UpdateBed(r.Context(), bedID, payload)
	if err != nil {
		writeJSON(w, http.StatusOK, Fail("failed to update bed"))
		return
	}
	writeJSON(w, http.StatusOK, Ok(b.ToJSON()))
}

func (a *AdminAPI) deleteBed(w http.ResponseWriter, r *http.Request, bedID string) {
	if err := a.Units.DeleteBed(r.Context(), bedID); err != nil {
		writeJSON(w, http.StatusOK, Fail("failed to delete bed"))
		return
	}
	writeJSON(w, http.StatusOK, Ok[any](nil))
}

// -------- Devices impl --------

func (a *AdminAPI) getDevices(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := a.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	// status can be repeated ?status=online&status=offline or status[]=...
	statuses := r.URL.Query()["status"]
	// Some frontend uses status as array directly; if it's comma-separated, split
	if len(statuses) == 1 && strings.Contains(statuses[0], ",") {
		statuses = strings.Split(statuses[0], ",")
	}
	filters := map[string]any{
		"status":          statuses,
		"business_access": r.URL.Query().Get("business_access"),
		"device_type":     r.URL.Query().Get("device_type"),
		"search_type":     r.URL.Query().Get("search_type"),
		"search_keyword":  r.URL.Query().Get("search_keyword"),
		"page":            parseInt(r.URL.Query().Get("page"), 1),
		"size":            parseInt(r.URL.Query().Get("size"), 20),
	}
	items, total, err := a.Devices.ListDevices(r.Context(), tenantID, filters)
	if err != nil {
		writeJSON(w, http.StatusOK, Fail("failed to list devices"))
		return
	}
	out := make([]any, 0, len(items))
	for _, d := range items {
		out = append(out, d.ToJSON())
	}
	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"items": out,
		"total": total,
	}))
}

func (a *AdminAPI) getDeviceDetail(w http.ResponseWriter, r *http.Request, deviceID string) {
	tenantID, ok := a.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	d, err := a.Devices.GetDevice(r.Context(), tenantID, deviceID)
	if err != nil {
		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusOK, Fail("device not found"))
			return
		}
		writeJSON(w, http.StatusOK, Fail("failed to get device"))
		return
	}
	writeJSON(w, http.StatusOK, Ok(d.ToJSON()))
}

func (a *AdminAPI) updateDevice(w http.ResponseWriter, r *http.Request, deviceID string) {
	tenantID, ok := a.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}
	// 关键对齐：前端不会“只传 unit_id”，它会先 ensureUnitRoom 再传 bound_room_id
	// 因此这里收紧：如果请求里携带了 unit_id，但 bound_room_id/bound_bed_id 都为空/缺失，直接报错，避免后端兜底掩盖问题
	unitID, _ := payload["unit_id"].(string)
	if unitID != "" {
		roomVal, hasRoom := payload["bound_room_id"]
		bedVal, hasBed := payload["bound_bed_id"]
		roomEmpty := !hasRoom || roomVal == nil || roomVal == ""
		bedEmpty := !hasBed || bedVal == nil || bedVal == ""
		if roomEmpty && bedEmpty {
			writeJSON(w, http.StatusOK, Fail("invalid binding: unit_id provided but bound_room_id/bound_bed_id missing"))
			return
		}
	}

	if err := a.Devices.UpdateDevice(r.Context(), tenantID, deviceID, payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("failed to update device"))
		return
	}
	writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
}

func (a *AdminAPI) deleteDevice(w http.ResponseWriter, r *http.Request, deviceID string) {
	tenantID, ok := a.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	if err := a.Devices.DisableDevice(r.Context(), tenantID, deviceID); err != nil {
		writeJSON(w, http.StatusOK, Fail("failed to delete device"))
		return
	}
	writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
}
