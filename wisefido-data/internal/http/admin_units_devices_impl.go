package httpapi

import (
	"database/sql"
	"net/http"
	"strings"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"
)

// 注意：以下方法已迁移到 UnitHandler，不再使用
// 保留此文件仅作为参考，实际路由由 UnitHandler 处理
// 如果编译错误，可以暂时注释掉这些方法

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
		return SystemTenantID(), true
	}
	writeJSON(w, http.StatusOK, Fail("tenant_id is required"))
	return "", false
}

// getBuildings 已迁移到 UnitHandler.ListBuildings
func (a *AdminAPI) getBuildings(w http.ResponseWriter, r *http.Request) {
	a.Stub.AdminUnits(w, r)
}

// -------- Units impl --------

// getUnits 已迁移到 UnitHandler.ListUnits
func (a *AdminAPI) getUnits(w http.ResponseWriter, r *http.Request) {
	a.Stub.AdminUnits(w, r)
}

// getUnitDetail 已迁移到 UnitHandler.GetUnit
func (a *AdminAPI) getUnitDetail(w http.ResponseWriter, r *http.Request, unitID string) {
	a.Stub.AdminUnits(w, r)
}

// createUnit 已迁移到 UnitHandler.CreateUnit
func (a *AdminAPI) createUnit(w http.ResponseWriter, r *http.Request) {
	a.Stub.AdminUnits(w, r)
}

// updateUnit 已迁移到 UnitHandler.UpdateUnit
func (a *AdminAPI) updateUnit(w http.ResponseWriter, r *http.Request, unitID string) {
	a.Stub.AdminUnits(w, r)
}

// deleteUnit 已迁移到 UnitHandler.DeleteUnit
func (a *AdminAPI) deleteUnit(w http.ResponseWriter, r *http.Request, unitID string) {
	a.Stub.AdminUnits(w, r)
}

// getRoomsWithBeds 已迁移到 UnitHandler.ListRoomsWithBeds
func (a *AdminAPI) getRoomsWithBeds(w http.ResponseWriter, r *http.Request) {
	a.Stub.AdminUnits(w, r)
}

// createRoom 已迁移到 UnitHandler.CreateRoom
func (a *AdminAPI) createRoom(w http.ResponseWriter, r *http.Request) {
	a.Stub.AdminUnits(w, r)
}

// updateRoom 已迁移到 UnitHandler.UpdateRoom
func (a *AdminAPI) updateRoom(w http.ResponseWriter, r *http.Request, roomID string) {
	a.Stub.AdminUnits(w, r)
}

// deleteRoom 已迁移到 UnitHandler.DeleteRoom
func (a *AdminAPI) deleteRoom(w http.ResponseWriter, r *http.Request, roomID string) {
	a.Stub.AdminUnits(w, r)
}

// getBeds 已迁移到 UnitHandler.ListBeds
func (a *AdminAPI) getBeds(w http.ResponseWriter, r *http.Request) {
	a.Stub.AdminUnits(w, r)
}

// createBed 已迁移到 UnitHandler.CreateBed
func (a *AdminAPI) createBed(w http.ResponseWriter, r *http.Request) {
	a.Stub.AdminUnits(w, r)
}

// updateBed 已迁移到 UnitHandler.UpdateBed
func (a *AdminAPI) updateBed(w http.ResponseWriter, r *http.Request, bedID string) {
	a.Stub.AdminUnits(w, r)
}

// deleteBed 已迁移到 UnitHandler.DeleteBed
func (a *AdminAPI) deleteBed(w http.ResponseWriter, r *http.Request, bedID string) {
	a.Stub.AdminUnits(w, r)
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
	filters := repository.DeviceFilters{
		Status:         statuses,
		BusinessAccess: r.URL.Query().Get("business_access"),
		DeviceType:     r.URL.Query().Get("device_type"),
		SearchType:     r.URL.Query().Get("search_type"),
		SearchKeyword:  r.URL.Query().Get("search_keyword"),
	}
	page := parseInt(r.URL.Query().Get("page"), 1)
	size := parseInt(r.URL.Query().Get("size"), 20)
	items, total, err := a.Devices.ListDevices(r.Context(), tenantID, filters, page, size)
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

	// 转换为domain.Device
	device := payloadToDevice(payload)
	if err := a.Devices.UpdateDevice(r.Context(), tenantID, deviceID, device); err != nil {
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

// payloadToDevice 将map[string]any转换为domain.Device
func payloadToDevice(payload map[string]any) *domain.Device {
	device := &domain.Device{}
	
	if v, ok := payload["device_name"].(string); ok {
		device.DeviceName = v
	}
	if v, ok := payload["device_store_id"].(string); ok && v != "" {
		device.DeviceStoreID = sql.NullString{String: v, Valid: true}
	}
	if v, ok := payload["serial_number"].(string); ok && v != "" {
		device.SerialNumber = sql.NullString{String: v, Valid: true}
	}
	if v, ok := payload["uid"].(string); ok && v != "" {
		device.UID = sql.NullString{String: v, Valid: true}
	}
	if v, ok := payload["bound_room_id"].(string); ok {
		if v != "" {
			device.BoundRoomID = sql.NullString{String: v, Valid: true}
		} else {
			device.BoundRoomID = sql.NullString{Valid: false}
		}
	}
	if v, ok := payload["bound_bed_id"].(string); ok {
		if v != "" {
			device.BoundBedID = sql.NullString{String: v, Valid: true}
		} else {
			device.BoundBedID = sql.NullString{Valid: false}
		}
	}
	if v, ok := payload["status"].(string); ok {
		device.Status = v
	}
	if v, ok := payload["business_access"].(string); ok {
		device.BusinessAccess = v
	}
	if v, ok := payload["monitoring_enabled"].(bool); ok {
		device.MonitoringEnabled = v
	}
	if v, ok := payload["metadata"].(string); ok && v != "" {
		device.Metadata = sql.NullString{String: v, Valid: true}
	}
	
	return device
}
