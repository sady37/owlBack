package httpapi

import (
	"context"
	"net/http"
	"strings"
	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

type AdminAPI struct {
	Units   repository.UnitsRepo
	Devices repository.DevicesRepo
	Tenant  repository.TenantResolver
	Stub    *StubHandler
	Log     *zap.Logger
}

func NewAdminAPI(units repository.UnitsRepo, devices repository.DevicesRepo, tenant repository.TenantResolver, stub *StubHandler, log *zap.Logger) *AdminAPI {
	return &AdminAPI{
		Units:   units,
		Devices: devices,
		Tenant:  tenant,
		Stub:    stub,
		Log:     log,
	}
}

// --- Units ---

func (a *AdminAPI) UnitsHandler(w http.ResponseWriter, r *http.Request) {
	if a.Units == nil {
		a.Stub.AdminUnits(w, r)
		return
	}
	switch r.URL.Path {
	case "/admin/api/v1/units":
		switch r.Method {
		case http.MethodGet:
			a.getUnits(w, r)
		case http.MethodPost:
			a.createUnit(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	default:
		if strings.HasPrefix(r.URL.Path, "/admin/api/v1/units/") {
			id := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/units/")
			if id == "" || strings.Contains(id, "/") {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			switch r.Method {
			case http.MethodGet:
				a.getUnitDetail(w, r, id)
			case http.MethodPut:
				a.updateUnit(w, r, id)
			case http.MethodDelete:
				a.deleteUnit(w, r, id)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}
	}
	a.Stub.AdminUnits(w, r)
}

// --- Buildings (virtual via units) ---

func (a *AdminAPI) BuildingsHandler(w http.ResponseWriter, r *http.Request) {
	if a.Units == nil {
		a.Stub.AdminUnits(w, r)
		return
	}
	// building write support is optional (memory repo provides it; postgres repo may not)
	type buildingWriter interface {
		CreateBuilding(ctx context.Context, tenantID string, payload map[string]any) (map[string]any, error)
		UpdateBuilding(ctx context.Context, tenantID, buildingID string, payload map[string]any) (map[string]any, error)
		DeleteBuilding(ctx context.Context, tenantID, buildingID string) error
	}

	switch {
	case r.URL.Path == "/admin/api/v1/buildings":
		switch r.Method {
		case http.MethodGet:
			a.getBuildings(w, r)
			return
		case http.MethodPost:
			bw, ok := a.Units.(buildingWriter)
			if !ok {
				a.Stub.AdminUnits(w, r)
				return
			}
			tenantID, ok2 := a.tenantIDFromReq(w, r)
			if !ok2 {
				return
			}
			var payload map[string]any
			if err := readBodyJSON(r, 1<<20, &payload); err != nil {
				writeJSON(w, http.StatusOK, Fail("invalid body"))
				return
			}
			out, err := bw.CreateBuilding(r.Context(), tenantID, payload)
			if err != nil {
				writeJSON(w, http.StatusOK, Fail("failed to create building"))
				return
			}
			writeJSON(w, http.StatusOK, Ok(out))
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/buildings/"):
		id := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/buildings/")
		if id == "" || strings.Contains(id, "/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		bw, ok := a.Units.(buildingWriter)
		if !ok {
			a.Stub.AdminUnits(w, r)
			return
		}
		tenantID, ok2 := a.tenantIDFromReq(w, r)
		if !ok2 {
			return
		}
		switch r.Method {
		case http.MethodPut:
			var payload map[string]any
			if err := readBodyJSON(r, 1<<20, &payload); err != nil {
				writeJSON(w, http.StatusOK, Fail("invalid body"))
				return
			}
			out, err := bw.UpdateBuilding(r.Context(), tenantID, id, payload)
			if err != nil {
				writeJSON(w, http.StatusOK, Fail("failed to update building"))
				return
			}
			writeJSON(w, http.StatusOK, Ok(out))
			return
		case http.MethodDelete:
			_ = bw.DeleteBuilding(r.Context(), tenantID, id)
			writeJSON(w, http.StatusOK, Ok[any](nil))
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}

	a.Stub.AdminUnits(w, r)
}

func (a *AdminAPI) RoomsHandler(w http.ResponseWriter, r *http.Request) {
	if a.Units == nil {
		a.Stub.AdminUnits(w, r)
		return
	}
	if r.URL.Path != "/admin/api/v1/rooms" {
		a.Stub.AdminUnits(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		a.getRoomsWithBeds(w, r)
	case http.MethodPost:
		a.createRoom(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *AdminAPI) RoomByIDHandler(w http.ResponseWriter, r *http.Request) {
	if a.Units == nil {
		a.Stub.AdminUnits(w, r)
		return
	}
	if !strings.HasPrefix(r.URL.Path, "/admin/api/v1/rooms/") {
		a.Stub.AdminUnits(w, r)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/rooms/")
	if id == "" || strings.Contains(id, "/") {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	switch r.Method {
	case http.MethodPut:
		a.updateRoom(w, r, id)
	case http.MethodDelete:
		a.deleteRoom(w, r, id)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *AdminAPI) BedsHandler(w http.ResponseWriter, r *http.Request) {
	if a.Units == nil {
		a.Stub.AdminUnits(w, r)
		return
	}
	if r.URL.Path != "/admin/api/v1/beds" {
		a.Stub.AdminUnits(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		a.getBeds(w, r)
	case http.MethodPost:
		a.createBed(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *AdminAPI) BedByIDHandler(w http.ResponseWriter, r *http.Request) {
	if a.Units == nil {
		a.Stub.AdminUnits(w, r)
		return
	}
	if !strings.HasPrefix(r.URL.Path, "/admin/api/v1/beds/") {
		a.Stub.AdminUnits(w, r)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/beds/")
	if id == "" || strings.Contains(id, "/") {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	switch r.Method {
	case http.MethodPut:
		a.updateBed(w, r, id)
	case http.MethodDelete:
		a.deleteBed(w, r, id)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// --- Devices ---

func (a *AdminAPI) DevicesHandler(w http.ResponseWriter, r *http.Request) {
	if a.Devices == nil {
		a.Stub.AdminDevices(w, r)
		return
	}
	if r.URL.Path == "/admin/api/v1/devices" {
		switch r.Method {
		case http.MethodGet:
			a.getDevices(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
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
			a.getDeviceDetail(w, r, id)
		case http.MethodPut:
			a.updateDevice(w, r, id)
		case http.MethodDelete:
			a.deleteDevice(w, r, id)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}
	a.Stub.AdminDevices(w, r)
}


