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
	// residents 路由已迁移到 ResidentHandler（见 RegisterResidentRoutes）
	// 如果数据库未启用，这些路由将不可用（返回 404）
	// r.Handle("/admin/api/v1/residents", s.AdminResidents)
	// r.Handle("/admin/api/v1/residents/", s.AdminResidents)
	// r.Handle("/admin/api/v1/contacts/", s.AdminResidents) // For contact password reset

	// tags - 已迁移到 TagsHandler，不再使用 StubHandler.AdminTags
	// 新路由在 RegisterTagsRoutes 中注册（需要数据库连接）
	// 如果数据库未启用，这些路由将不可用（返回 404）
	// r.Handle("/admin/api/v1/tags", s.AdminTags)
	// r.Handle("/admin/api/v1/tags/", s.AdminTags)
	// r.Handle("/admin/api/v1/tags/types", s.AdminTags)
	// r.Handle("/admin/api/v1/tags/for-object", s.AdminTags)

	// users - 已迁移到 UserHandler，不再使用 StubHandler.AdminUsers
	// 新路由在 RegisterUsersRoutes 中注册（需要数据库连接）
	// 如果数据库未启用，这些路由将不可用（返回 404）

	// roles - 已迁移到 RolesHandler，不再使用 StubHandler.AdminRoles
	// 新路由在 RegisterRolesRoutes 中注册（需要数据库连接）
	// 如果数据库未启用，这些路由将不可用（返回 404）
	// r.Handle("/admin/api/v1/roles", s.AdminRoles)
	// r.Handle("/admin/api/v1/roles/", s.AdminRoles)

	// role-permissions - 已迁移到 RolePermissionsHandler，不再使用 StubHandler.AdminRolePermissions
	// 新路由在 RegisterRolePermissionsRoutes 中注册（需要数据库连接）
	// 如果数据库未启用，这些路由将不可用（返回 404）
	// r.Handle("/admin/api/v1/role-permissions", s.AdminRolePermissions)
	// r.Handle("/admin/api/v1/role-permissions/", s.AdminRolePermissions)
	// r.Handle("/admin/api/v1/role-permissions/batch", s.AdminRolePermissions)
	// r.Handle("/admin/api/v1/role-permissions/resource-types", s.AdminRolePermissions)

	r.Handle("/admin/api/v1/service-levels", s.AdminServiceLevels)

	// card-overview - 已迁移到 CardOverviewHandler，不再使用 StubHandler.AdminCardOverview
	// 新路由在 RegisterCardOverviewRoutes 中注册（需要数据库连接）
	// 如果数据库未启用，这些路由将不可用（返回 404）
	// r.Handle("/admin/api/v1/card-overview", s.AdminCardOverview)

	// addresses - 已被 units 替换，前端未使用，已移除
	// 数据库中没有 addresses 表，地址管理已迁移到 units 表
	// 如果前端需要，应使用 /admin/api/v1/units API
	// r.Handle("/admin/api/v1/addresses", s.AdminAddresses)
	// r.Handle("/admin/api/v1/addresses/", s.AdminAddresses)

	// alarm-cloud - 已迁移到 AlarmCloudHandler，不再使用 StubHandler.AdminAlarm
	// 新路由在 RegisterAlarmCloudRoutes 中注册（需要数据库连接）
	// 如果数据库未启用，这些路由将不可用（返回 404）
	// r.Handle("/admin/api/v1/alarm-cloud", s.AdminAlarm)

	// alarm-events 路由已迁移到 AlarmEventHandler（见 RegisterAlarmEventRoutes）
	// 如果数据库未启用，这些路由将不可用（返回 404）

	// settings - 已迁移到 DeviceMonitorSettingsHandler（见 RegisterDeviceMonitorSettingsRoutes）
	// 如果数据库未启用，这些路由将不可用（返回 404）
	// r.Handle("/settings/api/v1/monitor/sleepace/", s.SettingsMonitor)
	// r.Handle("/settings/api/v1/monitor/radar/", s.SettingsMonitor)

	// sleepace reports - 已迁移到 SleepaceReportHandler（见 RegisterSleepaceReportRoutes）
	// 如果数据库未启用，这些路由将不可用（返回 404）

	// device relations - 已迁移到 DeviceHandler，不再使用 StubHandler.DeviceRelations
	// 新路由在 RegisterDeviceRoutes 中注册（需要数据库连接）
	// 如果数据库未启用，这些路由将不可用（返回 404）
	// r.Handle("/device/api/v1/device/", s.DeviceRelations)

	// auth - 已迁移到 AuthHandler，不再使用 StubHandler.Auth
	// 新路由在 RegisterAuthRoutes 中注册（需要数据库连接）
	// 如果数据库未启用，这些路由将不可用（返回 404）

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
// 注意：Unit/Room/Bed 路由已迁移到 UnitHandler（见 RegisterUnitRoutes）
// 注意：Devices 路由已迁移到 DeviceHandler（见 RegisterDeviceRoutes）
// 此函数已废弃，保留仅为向后兼容（不再注册任何路由）
func (r *Router) RegisterAdminUnitDeviceRoutes(admin *AdminAPI) {
	// Unit/Room/Bed 路由已迁移到 UnitHandler，不再在这里注册
	// r.Handle("/admin/api/v1/buildings", admin.BuildingsHandler)
	// r.Handle("/admin/api/v1/buildings/", admin.BuildingsHandler)
	// r.Handle("/admin/api/v1/units", admin.UnitsHandler)
	// r.Handle("/admin/api/v1/units/", admin.UnitsHandler)
	// r.Handle("/admin/api/v1/rooms", admin.RoomsHandler)
	// r.Handle("/admin/api/v1/rooms/", admin.RoomByIDHandler)
	// r.Handle("/admin/api/v1/beds", admin.BedsHandler)
	// r.Handle("/admin/api/v1/beds/", admin.BedByIDHandler)

	// Devices 路由已迁移到 DeviceHandler（见 RegisterDeviceRoutes），不再在这里注册
	// r.Handle("/admin/api/v1/devices", admin.DevicesHandler)
	// r.Handle("/admin/api/v1/devices/", admin.DevicesHandler)

	// device-store 路由已迁移到独立的 DeviceStoreHandler（见 RegisterDeviceStoreRoutes）
	// 保留 AdminAPI.DeviceStoreHandler 作为备用（向后兼容）
}

// RegisterRolesRoutes 注册角色管理路由
func (r *Router) RegisterRolesRoutes(h *RolesHandler) {
	r.Handle("/admin/api/v1/roles", h.ServeHTTP)
	r.Handle("/admin/api/v1/roles/", h.ServeHTTP)
}

// RegisterRolePermissionsRoutes 注册角色权限管理路由
func (r *Router) RegisterRolePermissionsRoutes(h *RolePermissionsHandler) {
	r.Handle("/admin/api/v1/role-permissions", h.ServeHTTP)
	r.Handle("/admin/api/v1/role-permissions/", h.ServeHTTP)
	r.Handle("/admin/api/v1/role-permissions/batch", h.ServeHTTP)
	r.Handle("/admin/api/v1/role-permissions/resource-types", h.ServeHTTP)
}

// RegisterTagsRoutes 注册标签管理路由
func (r *Router) RegisterTagsRoutes(h *TagsHandler) {
	r.Handle("/admin/api/v1/tags", h.ServeHTTP)
	r.Handle("/admin/api/v1/tags/", h.ServeHTTP)
	r.Handle("/admin/api/v1/tags/types", h.ServeHTTP)
	r.Handle("/admin/api/v1/tags/for-object", h.ServeHTTP)
}

// RegisterAlarmCloudRoutes 注册告警配置管理路由
func (r *Router) RegisterAlarmCloudRoutes(h *AlarmCloudHandler) {
	r.Handle("/admin/api/v1/alarm-cloud", h.ServeHTTP)
}

// RegisterAuthRoutes 注册认证授权路由
func (r *Router) RegisterAuthRoutes(h *AuthHandler) {
	r.Handle("/auth/api/v1/login", h.ServeHTTP)
	r.Handle("/auth/api/v1/institutions/search", h.ServeHTTP)
	r.Handle("/auth/api/v1/forgot-password/send-code", h.ServeHTTP)
	r.Handle("/auth/api/v1/forgot-password/verify-code", h.ServeHTTP)
	r.Handle("/auth/api/v1/forgot-password/reset", h.ServeHTTP)
}

// RegisterDeviceRoutes 注册设备管理路由
func (r *Router) RegisterDeviceRoutes(h *DeviceHandler) {
	r.Handle("/admin/api/v1/devices", h.ServeHTTP)
	r.Handle("/admin/api/v1/devices/", h.ServeHTTP)
	// 设备关联关系查询
	r.Handle("/device/api/v1/device/", h.GetDeviceRelations)
}

// RegisterDeviceStoreRoutes 注册设备库存管理路由
func (r *Router) RegisterDeviceStoreRoutes(h *DeviceStoreHandler) {
	r.Handle("/admin/api/v1/device-store", h.ServeHTTP)
	r.Handle("/admin/api/v1/device-store/", h.ServeHTTP)
	r.Handle("/admin/api/v1/device-store/batch", h.ServeHTTP)
	r.Handle("/admin/api/v1/device-store/import", h.ServeHTTP)
	r.Handle("/admin/api/v1/device-store/import-template", h.ServeHTTP)
	r.Handle("/admin/api/v1/device-store/export", h.ServeHTTP)
}

// RegisterUnitRoutes 注册单元管理路由（Building, Unit, Room, Bed）
func (r *Router) RegisterUnitRoutes(h *UnitHandler) {
	// Buildings
	r.Handle("/admin/api/v1/buildings", h.ServeHTTP)
	r.Handle("/admin/api/v1/buildings/", h.ServeHTTP)

	// Units
	r.Handle("/admin/api/v1/units", h.ServeHTTP)
	r.Handle("/admin/api/v1/units/", h.ServeHTTP)

	// Rooms
	r.Handle("/admin/api/v1/rooms", h.ServeHTTP)
	r.Handle("/admin/api/v1/rooms/", h.ServeHTTP)

	// Beds
	r.Handle("/admin/api/v1/beds", h.ServeHTTP)
	r.Handle("/admin/api/v1/beds/", h.ServeHTTP)
}

// RegisterUsersRoutes 注册用户管理路由
func (r *Router) RegisterUsersRoutes(h *UserHandler) {
	r.Handle("/admin/api/v1/users", h.ServeHTTP)
	r.Handle("/admin/api/v1/users/", h.ServeHTTP)
}

// RegisterDeviceMonitorSettingsRoutes 注册设备监控配置路由
func (r *Router) RegisterDeviceMonitorSettingsRoutes(h *DeviceMonitorSettingsHandler) {
	r.Handle("/settings/api/v1/monitor/sleepace/", h.ServeHTTP)
	r.Handle("/settings/api/v1/monitor/radar/", h.ServeHTTP)
}

// RegisterAlarmEventRoutes 注册报警事件管理路由
func (r *Router) RegisterAlarmEventRoutes(h *AlarmEventHandler) {
	r.Handle("/admin/api/v1/alarm-events", h.ServeHTTP)
	r.Handle("/admin/api/v1/alarm-events/", h.ServeHTTP)
}

// RegisterResidentRoutes 注册住户管理路由
func (r *Router) RegisterResidentRoutes(h *ResidentHandler) {
	r.HandleHandler("/admin/api/v1/residents", h)
	r.HandleHandler("/admin/api/v1/residents/", h)
	// 联系人密码重置路由
	r.HandleHandler("/admin/api/v1/contacts/", h)
}

// RegisterSleepaceReportRoutes 注册 Sleepace 睡眠报告路由
func (r *Router) RegisterSleepaceReportRoutes(h *SleepaceReportHandler) {
	r.Handle("/sleepace/api/v1/sleepace/reports/", h.ServeHTTP)
}

// RegisterCardOverviewRoutes 注册卡片概览路由
func (r *Router) RegisterCardOverviewRoutes(h *CardOverviewHandler) {
	r.Handle("/admin/api/v1/card-overview", h.ServeHTTP)
}
