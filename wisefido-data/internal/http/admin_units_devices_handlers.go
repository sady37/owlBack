package httpapi

import (
	"net/http"
	"wisefido-data/internal/repository"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type AdminAPI struct {
	Units       repository.UnitsRepository
	Devices     repository.DevicesRepository
	DeviceStore repository.DeviceStoreRepository
	Tenant      repository.TenantResolver
	Stub        *StubHandler
	Log         *zap.Logger
	RedisClient *redis.Client
}

func NewAdminAPI(units repository.UnitsRepository, devices repository.DevicesRepository, deviceStore repository.DeviceStoreRepository, tenant repository.TenantResolver, stub *StubHandler, log *zap.Logger, redisClient *redis.Client) *AdminAPI {
	return &AdminAPI{
		Units:       units,
		Devices:     devices,
		DeviceStore: deviceStore,
		Tenant:      tenant,
		Stub:        stub,
		Log:         log,
		RedisClient: redisClient,
	}
}

// --- Units ---
// 注意：Units 路由已迁移到 UnitHandler（见 RegisterUnitRoutes）
// 这里保留作为备用，但不再被调用

func (a *AdminAPI) UnitsHandler(w http.ResponseWriter, r *http.Request) {
	// Units 路由已迁移到 UnitHandler，这里返回 stub
	a.Stub.AdminUnits(w, r)
}

// --- Buildings (实体表) ---
// 注意：Buildings 路由已迁移到 UnitHandler（见 RegisterUnitRoutes）
// 这里保留作为备用，但不再被调用

func (a *AdminAPI) BuildingsHandler(w http.ResponseWriter, r *http.Request) {
	// Buildings 路由已迁移到 UnitHandler，这里返回 stub
	a.Stub.AdminUnits(w, r)
}

// Rooms 路由已迁移到 UnitHandler
func (a *AdminAPI) RoomsHandler(w http.ResponseWriter, r *http.Request) {
	a.Stub.AdminUnits(w, r)
}

func (a *AdminAPI) RoomByIDHandler(w http.ResponseWriter, r *http.Request) {
	a.Stub.AdminUnits(w, r)
}

// Beds 路由已迁移到 UnitHandler
func (a *AdminAPI) BedsHandler(w http.ResponseWriter, r *http.Request) {
	a.Stub.AdminUnits(w, r)
}

func (a *AdminAPI) BedByIDHandler(w http.ResponseWriter, r *http.Request) {
	a.Stub.AdminUnits(w, r)
}

// --- Devices ---
// 注意：Devices 路由已迁移到 DeviceHandler（见 RegisterDeviceRoutes）
// 这里保留作为备用，但不再被调用

func (a *AdminAPI) DevicesHandler(w http.ResponseWriter, r *http.Request) {
	// Devices 路由已迁移到 DeviceHandler，这里返回 stub
	a.Stub.AdminDevices(w, r)
}

// --- Device Store ---

func (a *AdminAPI) DeviceStoreHandler(w http.ResponseWriter, r *http.Request) {
	if a.DeviceStore == nil {
		a.Stub.AdminDeviceStore(w, r)
		return
	}

	switch {
	case r.URL.Path == "/admin/api/v1/device-store":
		switch r.Method {
		case http.MethodGet:
			a.getDeviceStores(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	case r.URL.Path == "/admin/api/v1/device-store/batch":
		switch r.Method {
		case http.MethodPut:
			a.batchUpdateDeviceStores(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	case r.URL.Path == "/admin/api/v1/device-store/import":
		switch r.Method {
		case http.MethodPost:
			a.importDeviceStores(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	case r.URL.Path == "/admin/api/v1/device-store/import-template":
		switch r.Method {
		case http.MethodGet:
			a.getImportTemplate(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	case r.URL.Path == "/admin/api/v1/device-store/export":
		switch r.Method {
		case http.MethodGet:
			a.exportDeviceStores(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}
	a.Stub.AdminDeviceStore(w, r)
}
