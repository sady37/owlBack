package httpapi

import (
	"net/http"
	"strings"
)

// --- admin/api/v1 (stub-only implementations) ---

func (s *StubHandler) AdminDevices(w http.ResponseWriter, r *http.Request) {
	// GET /admin/api/v1/devices
	// GET /admin/api/v1/devices/:id
	// PUT /admin/api/v1/devices/:id
	// DELETE /admin/api/v1/devices/:id
	if r.URL.Path == "/admin/api/v1/devices" {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, Ok(map[string]any{"items": []any{}, "total": 0}))
		return
	}
	if strings.HasPrefix(r.URL.Path, "/admin/api/v1/devices/") {
		id := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/devices/")
		if id == "" || strings.Contains(id, "/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodGet:
			// 参考 owlFront Device model 的必填字段：device_id/tenant_id/device_name/status/business_access
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"device_id":          id,
				"tenant_id":          "",
				"device_name":        "stub-" + id,
				"status":             "offline",
				"business_access":    "pending",
				"monitoring_enabled": false,
			}))
		case http.MethodPut:
			writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
		case http.MethodDelete:
			writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

func (s *StubHandler) AdminUnits(w http.ResponseWriter, r *http.Request) {
	// buildings
	switch {
	case r.URL.Path == "/admin/api/v1/buildings":
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, Ok([]any{}))
		case http.MethodPost:
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"building_id":   "stub-building",
				"building_name": "stub",
				"floors":        0,
				"tenant_id":     "",
			}))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/buildings/"):
		id := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/buildings/")
		if id == "" || strings.Contains(id, "/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodPut:
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"building_id":   id,
				"building_name": "stub",
				"floors":        0,
				"tenant_id":     "",
			}))
		case http.MethodDelete:
			writeJSON(w, http.StatusOK, Ok[any](nil))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	// units
	switch {
	case r.URL.Path == "/admin/api/v1/units":
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, Ok(map[string]any{"items": []any{}, "total": 0}))
		case http.MethodPost:
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"unit_id":     "stub-unit",
				"tenant_id":   "",
				"unit_name":   "stub",
				"unit_number": "stub",
				"unit_type":   "Facility",
			}))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/units/"):
		id := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/units/")
		if id == "" || strings.Contains(id, "/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"unit_id":     id,
				"tenant_id":   "",
				"unit_name":   "stub-" + id,
				"unit_number": "stub",
				"unit_type":   "Facility",
			}))
		case http.MethodPut:
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"unit_id":     id,
				"tenant_id":   "",
				"unit_name":   "stub-" + id,
				"unit_number": "stub",
				"unit_type":   "Facility",
			}))
		case http.MethodDelete:
			writeJSON(w, http.StatusOK, Ok[any](nil))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	// rooms
	switch {
	case r.URL.Path == "/admin/api/v1/rooms":
		switch r.Method {
		case http.MethodGet:
			// getRoomsApi 期待 RoomWithBeds[]
			writeJSON(w, http.StatusOK, Ok([]any{}))
		case http.MethodPost:
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"room_id":    "stub-room",
				"unit_id":    "",
				"room_name":  "stub",
				"is_default": false,
			}))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/rooms/"):
		id := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/rooms/")
		if id == "" || strings.Contains(id, "/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodPut:
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"room_id":    id,
				"unit_id":    "",
				"room_name":  "stub-" + id,
				"is_default": false,
			}))
		case http.MethodDelete:
			writeJSON(w, http.StatusOK, Ok[any](nil))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	// beds
	switch {
	case r.URL.Path == "/admin/api/v1/beds":
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, Ok([]any{}))
		case http.MethodPost:
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"bed_id":   "stub-bed",
				"room_id":  "",
				"bed_name": "stub",
			}))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/beds/"):
		id := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/beds/")
		if id == "" || strings.Contains(id, "/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodPut:
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"bed_id":   id,
				"room_id":  "",
				"bed_name": "stub-" + id,
			}))
		case http.MethodDelete:
			writeJSON(w, http.StatusOK, Ok[any](nil))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	w.WriteHeader(http.StatusNotFound)
}
