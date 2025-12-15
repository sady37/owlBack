package httpapi

import (
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// Router 使用标准库 http.ServeMux（避免引入第三方路由依赖）
type Router struct {
	mux    *http.ServeMux
	logger *zap.Logger
}

func NewRouter(logger *zap.Logger) *Router {
	return &Router{
		mux:    http.NewServeMux(),
		logger: logger,
	}
}

func (r *Router) Handle(pattern string, h http.HandlerFunc) {
	r.mux.HandleFunc(pattern, h)
}

// HandleHandler 支持 http.Handler 接口（用于 pprof 等）
func (r *Router) HandleHandler(pattern string, h http.Handler) {
	r.mux.Handle(pattern, h)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// RegisterVitalFocusRoutes 注册与 owlFront 对齐的路由
func (r *Router) RegisterVitalFocusRoutes(v *VitalFocusHandler) {
	// list
	r.Handle("/data/api/v1/data/vital-focus/cards", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		v.GetCards(w, req)
	})

	// selection
	r.Handle("/data/api/v1/data/vital-focus/selection", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		v.SaveSelection(w, req)
	})

	// card/{id} (兼容 residentId/cardId)
	r.Handle("/data/api/v1/data/vital-focus/card/", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		id := strings.TrimPrefix(req.URL.Path, "/data/api/v1/data/vital-focus/card/")
		if id == "" || strings.Contains(id, "/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		v.GetCardByIDOrResident(w, req, id)
	})
}

// RegisterStubRoutes: 先把 owlFront 写死的其它 API 路由补齐（避免 404）
func (r *Router) RegisterStubRoutes(s *StubHandler) {
	// admin
	r.Handle("/admin/api/v1/residents", s.AdminResidents)
	r.Handle("/admin/api/v1/residents/", s.AdminResidents)

	r.Handle("/admin/api/v1/tags", s.AdminTags)
	r.Handle("/admin/api/v1/tags/", s.AdminTags)
	r.Handle("/admin/api/v1/tags/types", s.AdminTags)
	r.Handle("/admin/api/v1/tags/for-object", s.AdminTags)

	r.Handle("/admin/api/v1/users", s.AdminUsers)
	r.Handle("/admin/api/v1/users/", s.AdminUsers)

	r.Handle("/admin/api/v1/roles", s.AdminRoles)
	r.Handle("/admin/api/v1/roles/", s.AdminRoles)

	r.Handle("/admin/api/v1/role-permissions", s.AdminRolePermissions)
	r.Handle("/admin/api/v1/role-permissions/", s.AdminRolePermissions)
	r.Handle("/admin/api/v1/role-permissions/batch", s.AdminRolePermissions)
	r.Handle("/admin/api/v1/role-permissions/resource-types", s.AdminRolePermissions)

	r.Handle("/admin/api/v1/service-levels", s.AdminServiceLevels)
	r.Handle("/admin/api/v1/card-overview", s.AdminCardOverview)

	r.Handle("/admin/api/v1/addresses", s.AdminAddresses)
	r.Handle("/admin/api/v1/addresses/", s.AdminAddresses)

	r.Handle("/admin/api/v1/alarm-cloud", s.AdminAlarm)
	r.Handle("/admin/api/v1/alarm-events", s.AdminAlarm)
	r.Handle("/admin/api/v1/alarm-events/", s.AdminAlarm)

	// settings
	r.Handle("/settings/api/v1/monitor/sleepace/", s.SettingsMonitor)
	r.Handle("/settings/api/v1/monitor/radar/", s.SettingsMonitor)

	// sleepace reports
	r.Handle("/sleepace/api/v1/sleepace/reports/", s.SleepaceReports)

	// device relations
	r.Handle("/device/api/v1/device/", s.DeviceRelations)

	// auth
	r.Handle("/auth/api/v1/login", s.Auth)
	r.Handle("/auth/api/v1/institutions/search", s.Auth)
	r.Handle("/auth/api/v1/forgot-password/send-code", s.Auth)
	r.Handle("/auth/api/v1/forgot-password/verify-code", s.Auth)
	r.Handle("/auth/api/v1/forgot-password/reset", s.Auth)

	// example
	r.Handle("/api/v1/example/items", s.Example)
	r.Handle("/api/v1/example/", s.Example)
	r.Handle("/api/v1/example/item", s.Example)
}

// RegisterAdminTenantRoutes：Tenant management（platform-level）
func (r *Router) RegisterAdminTenantRoutes(h *TenantsHandler) {
	r.Handle("/admin/api/v1/tenants", h.ServeHTTP)
	r.Handle("/admin/api/v1/tenants/", h.ServeHTTP)
}

// RegisterAdminUnitDeviceRoutes：Unit/Room/Bed + Devices（地址类 + 设备类）
func (r *Router) RegisterAdminUnitDeviceRoutes(admin *AdminAPI) {
	// buildings（虚拟 derived from units）
	r.Handle("/admin/api/v1/buildings", admin.BuildingsHandler)
	r.Handle("/admin/api/v1/buildings/", admin.BuildingsHandler)

	r.Handle("/admin/api/v1/units", admin.UnitsHandler)
	r.Handle("/admin/api/v1/units/", admin.UnitsHandler)

	r.Handle("/admin/api/v1/rooms", admin.RoomsHandler)
	r.Handle("/admin/api/v1/rooms/", admin.RoomByIDHandler)

	r.Handle("/admin/api/v1/beds", admin.BedsHandler)
	r.Handle("/admin/api/v1/beds/", admin.BedByIDHandler)

	r.Handle("/admin/api/v1/devices", admin.DevicesHandler)
	r.Handle("/admin/api/v1/devices/", admin.DevicesHandler)
}


