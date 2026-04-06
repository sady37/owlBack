package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"wisefido-data/internal/service"

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

// RegisterStubRoutes: 先把 owlFront 写死的其它 API 路由补齐（避免 404）
func (r *Router) RegisterStubRoutes(s *StubHandler) {
	// admin
	// residents 路由已迁移到 ResidentHandler（见 RegisterResidentRoutes）
	// 如果数据库未启用，这些路由将不可用（返回 404）

	// tags - 已删除（tags 表不存在）

	// users - 已迁移到 UserHandler，不再使用 StubHandler.AdminUsers
	// 新路由在 RegisterUsersRoutes 中注册（需要数据库连接）
	// 如果数据库未启用，这些路由将不可用（返回 404）

	// roles - 已迁移到 RolesHandler，不再使用 StubHandler.AdminRoles
	// 新路由在 RegisterRolesRoutes 中注册（需要数据库连接）
	// 如果数据库未启用，这些路由将不可用（返回 404）
	// r.Handle("/admin/api/v1/roles", s.AdminRoles)
	// r.Handle("/admin/api/v1/roles/", s.AdminRoles)

	// role-permissions - 已迁移到 RolePermissionsHandler
	// 新路由在 RegisterRolePermissionsRoutes 中注册（需要数据库连接）
	// 旧的 StubHandler.AdminRolePermissions 已删除

	r.Handle("/admin/api/v1/service-levels", s.AdminServiceLevels)

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

// RegisterRadarRoutes 注册 Radar 设备 API 路由（所有端点均受 AuthMiddleware 保护）
func (r *Router) RegisterRadarRoutes(h *RadarHandler) {
	// 卡片设备列表：GET /radar-device/api/v1/radar-device/card/{cardId}/devices
	r.Handle("/radar-device/api/v1/radar-device/card/", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		path := req.URL.Path
		if strings.HasSuffix(path, "/devices") {
			h.GetCardDevices(w, req)
			return
		}
		http.NotFound(w, req)
	})

	// 设备路由（多个端点）
	r.Handle("/radar-device/api/v1/radar-device/device/", func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path
		// GET /radar-device/api/v1/radar-device/device/{deviceId}/card-devices
		if strings.HasSuffix(path, "/card-devices") && req.Method == http.MethodGet {
			h.GetCardDevicesByDeviceID(w, req)
			return
		}
		// GET /radar-device/api/v1/radar-device/device/{deviceId}/stream (SSE)
		if strings.HasSuffix(path, "/stream") && req.Method == http.MethodGet {
			h.SubscribeRealtimeStream(w, req)
			return
		}
		// GET /radar-device/api/v1/radar-device/device/{deviceId}/original-properties
		if strings.HasSuffix(path, "/original-properties") && req.Method == http.MethodGet {
			h.GetOriginalProperties(w, req)
			return
		}
		// PUT /radar-device/api/v1/radar-device/device/{deviceId}/config
		if strings.HasSuffix(path, "/config") && req.Method == http.MethodPut {
			h.UpdateConfig(w, req)
			return
		}
		// POST /radar-device/api/v1/radar-device/device/{deviceId}/bind
		if strings.HasSuffix(path, "/bind") && req.Method == http.MethodPost {
			h.BindDevice(w, req)
			return
		}
		// POST /radar-device/api/v1/radar-device/device/{deviceId}/unbind
		if strings.HasSuffix(path, "/unbind") && req.Method == http.MethodPost {
			h.UnbindDevice(w, req)
			return
		}
		// POST /radar-device/api/v1/radar-device/device/{deviceId}/control
		if strings.HasSuffix(path, "/control") && req.Method == http.MethodPost {
			h.Control(w, req)
			return
		}
		http.NotFound(w, req)
	})

	// 房间布局：PUT /radar-device/api/v1/radar-device/room/{roomId}/layout
	r.Handle("/radar-device/api/v1/radar-device/room/", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		path := req.URL.Path
		if strings.HasSuffix(path, "/layout") {
			h.PutRoomLayout(w, req)
			return
		}
		http.NotFound(w, req)
	})
}

// RegisterPlaybackRoutes 历史回放（防扫库：见 service/track_playback_service.go ValidatePlaybackWindow）
func (r *Router) RegisterPlaybackRoutes(h *PlaybackHandler) {
	r.Handle("/api/radar/playback", h.PostRadarPlayback)
	r.Handle("/api/vital/playback", h.PostVitalPlayback)
	r.Handle("/api/radar/alarm-replay-context", h.PostAlarmReplayContext)
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

	// device-store：RegisterDeviceStoreRoutes（DeviceStoreHandler）
}

// RegisterRolesRoutes 注册角色管理路由
func (r *Router) RegisterRolesRoutes(h *RolesHandler) {
	r.Handle("/admin/api/v1/roles", h.ServeHTTP)
	r.Handle("/admin/api/v1/roles/", h.ServeHTTP)
}

// RegisterBranchesRoutes 注册院区管理路由
func (r *Router) RegisterBranchesRoutes(h *BranchesHandler) {
	r.Handle("/admin/api/v1/branches", h.ServeHTTP)
	r.Handle("/admin/api/v1/branches/", h.ServeHTTP)
}

// RegisterRolePermissionsRoutes 注册角色权限管理路由
func (r *Router) RegisterRolePermissionsRoutes(h *RolePermissionsHandler) {
	r.Handle("/admin/api/v1/role-permissions", h.ServeHTTP)
	r.Handle("/admin/api/v1/role-permissions/", h.ServeHTTP)
	r.Handle("/admin/api/v1/role-permissions/batch", h.ServeHTTP)
	r.Handle("/admin/api/v1/role-permissions/resource-types", h.ServeHTTP)
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
	r.Handle("/auth/api/v1/verify-pin", h.ServeHTTP)
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

	// Beds (beds-with-details 必须在 beds 之前注册，否则会被 beds 匹配)
	r.Handle("/admin/api/v1/beds-with-details", h.ServeHTTP)
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
	r.Handle("/settings/api/v1/monitor/sleepad/", h.ServeHTTP)
	r.Handle("/settings/api/v1/monitor/radar/", h.ServeHTTP)
	// 注册默认设置路由
	r.Handle("/settings/api/v1/monitor/default/sleepad", h.ServeHTTP)
	r.Handle("/settings/api/v1/monitor/default/radar", h.ServeHTTP)
}

// RegisterAlarmEventRoutes 注册报警事件管理路由（网页端告警记录 Vue：owlFront /admin/api/v1/alarm-events）
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

// RegisterCardOverviewRoutes 注册卡片概览相关路由

// CardRealtimeHandler 专门处理卡片实时数据的处理器
type CardRealtimeHandler struct {
	realtimeSvc *service.CardRealtimeService
	staticSvc   *service.CardStaticService
	logger      *zap.Logger
}

// NewCardRealtimeHandler 创建卡片实时数据处理器
func NewCardRealtimeHandler(realtimeSvc *service.CardRealtimeService, staticSvc *service.CardStaticService, logger *zap.Logger) *CardRealtimeHandler {
	return &CardRealtimeHandler{
		realtimeSvc: realtimeSvc,
		staticSvc:   staticSvc,
		logger:      logger,
	}
}

// PullRealtimeData 处理实时数据拉取请求
func (h *CardRealtimeHandler) PullRealtimeData(w http.ResponseWriter, req *http.Request) {
	// 从请求中提取数据
	ctx := req.Context()

	// 解析请求体：cards_by_branch（branchID → cardIDs）
	var body struct {
		CardsByBranch map[string][]string `json:"cards_by_branch"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		h.logger.Error("Failed to decode pull realtime request", zap.Error(err))
		writeJSON(w, http.StatusBadRequest, Fail("Invalid request body"))
		return
	}
	if body.CardsByBranch == nil {
		body.CardsByBranch = map[string][]string{}
	}

	tenantID := req.Header.Get("X-Tenant-Id")
	userID := req.Header.Get("X-User-Id")
	if tenantID == "" || userID == "" {
		h.logger.Error("Missing required authentication headers")
		writeJSON(w, http.StatusUnauthorized, Fail("Missing required authentication headers"))
		return
	}

	serviceReq := service.GetCardRealtimeRequest{
		TenantID:      tenantID,
		UserID:        userID,
		CardsByBranch: body.CardsByBranch,
	}

	resp, err := h.realtimeSvc.GetCardRealtime(ctx, serviceReq)
	if err != nil {
		h.logger.Error("Failed to pull realtime data", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(resp))
}

// RegisterCardRealtimeRoutes 注册卡片实时数据相关路由
func (r *Router) RegisterCardRealtimeRoutes(h *CardRealtimeHandler) {
	// pull realtime data (POST)
	r.Handle("/data/api/v1/data/vital-focus/pull-realtime", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		h.PullRealtimeData(w, req)
	})
}

// RegisterVitalFocusRoutes 注册 VitalFocus 前端 API 路由
func (r *Router) RegisterMonitorRoutes(h *MonitorHandler) {
	// GET /data/api/v1/data/vital-focus/cards - 获取卡片列表
	r.Handle("/data/api/v1/data/vital-focus/cards", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		h.GetCards(w, req)
	})

	// POST /data/api/v1/data/vital-focus/cards-by-ids - 按 cardID 拉静态数据
	r.Handle("/data/api/v1/data/vital-focus/cards-by-ids", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		h.GetCardsByCardIDs(w, req)
	})

	// GET /data/api/v1/data/vital-focus/cards/stream - Overview 多卡 SSE
	r.Handle("/data/api/v1/data/vital-focus/cards/stream", func(w http.ResponseWriter, req *http.Request) {
		h.SubscribeCardsStream(w, req)
	})

	// POST /data/api/v1/data/vital-focus/cards/stream/init - SSE 初始化（传 watchIDs + viewIDs）
	r.Handle("/data/api/v1/data/vital-focus/cards/stream/init", func(w http.ResponseWriter, req *http.Request) {
		h.InitSSE(w, req)
	})

	// POST /data/api/v1/data/vital-focus/cards/stream/view - 切页更新 viewIDs
	r.Handle("/data/api/v1/data/vital-focus/cards/stream/view", func(w http.ResponseWriter, req *http.Request) {
		h.UpdateSSEView(w, req)
	})

	// POST /data/api/v1/data/vital-focus/cards/stream/watch - card_change 后热更新 watchIDs
	r.Handle("/data/api/v1/data/vital-focus/cards/stream/watch", func(w http.ResponseWriter, req *http.Request) {
		h.UpdateSSEWatch(w, req)
	})

	// GET /data/api/v1/data/vital-focus/card/{id} - 获取卡片详情
	r.Handle("/data/api/v1/data/vital-focus/card/", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		rest := strings.TrimPrefix(req.URL.Path, "/data/api/v1/data/vital-focus/card/")
		if rest == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		// rest 格式: {id} | {id}/stream | {id}/caregivers
		parts := strings.SplitN(rest, "/", 2)
		cardID := parts[0]
		suffix := ""
		if len(parts) == 2 {
			suffix = parts[1]
		}
		switch suffix {
		case "stream":
			h.SubscribeRealtimeStream(w, req, cardID)
		case "caregivers":
			h.GetCardCaregivers(w, req, cardID)
		case "status":
			h.GetCardStatus(w, req, cardID)
		case "":
			h.GetCardInfo(w, req, cardID)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

// RegisterRoundsRoutes 注册巡房记录 API（POST/GET /data/api/v1/data/vital-focus/rounds）
func (r *Router) RegisterRoundsRoutes(roundsHandler *RoundsHandler) {
	r.Handle("/data/api/v1/data/vital-focus/rounds", roundsHandler.ServeHTTP)
	r.Handle("/data/api/v1/data/vital-focus/rounds/", roundsHandler.ServeHTTP)
}

// RegisterAPNSRoutes 注册 iOS APNs 设备 Token 路由
// POST   /data/api/v1/auth/devices/apns  — 登录后注册设备
// DELETE /data/api/v1/auth/devices/apns  — 登出时注销设备
func (r *Router) RegisterAPNSRoutes(h *APNSHandler) {
	r.Handle("/data/api/v1/auth/devices/apns", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			h.RegisterDevice(w, req)
		case http.MethodDelete:
			h.UnregisterDevice(w, req)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}


// RegisterAlarmPushRoute cardagg → APNs：POST /internal/v1/push/alarm（X-Internal-Secret）
func (r *Router) RegisterAlarmPushRoute(h *AlarmPushHandler) {
	r.Handle("/internal/v1/push/alarm", h.ServeHTTP)
}