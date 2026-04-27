package httpapi

import (
	"net/http"

	"wisefido-sleepace/internal/config"
	"wisefido-sleepace/internal/service"

	"go.uber.org/zap"
)

type Router struct {
	mux           *http.ServeMux
	proxy         *SleepaceProxy
	deviceAPI     *DeviceAPI
	statusTracker *service.DeviceStatusTracker
	logger        *zap.Logger
}

func NewRouter(
	cfg *config.SleepaceConfig,
	sleepaceAPI *service.SleepaceAPI,
	cardMapping *service.CardMappingService,
	statusTracker *service.DeviceStatusTracker,
	logger *zap.Logger,
) *Router {
	r := &Router{
		mux:           http.NewServeMux(),
		proxy:         NewSleepaceProxy(cfg, logger),
		deviceAPI:     NewDeviceAPI(sleepaceAPI, cardMapping, logger),
		statusTracker: statusTracker,
		logger:        logger,
	}
	r.routes()
	return r
}

func (r *Router) routes() {
	// Multipart 单独走专用 handler；必须在通用 JSON proxy 前注册（更长路径优先匹配）。
	r.mux.HandleFunc("/api/v1/proxy/sleepace/firmware/uploadFile", r.proxy.UploadFirmwareHandler())
	// Proxy to sleepace-service (JSON, 自动注入 appId/secureKey)
	r.mux.HandleFunc("/api/v1/proxy/", r.proxy.Handler())

	// Device lifecycle
	r.mux.HandleFunc("/api/v1/sleepace/device/initialize", r.deviceAPI.HandleInitialize)
	r.mux.HandleFunc("/api/v1/sleepace/device/", r.deviceAPI.HandleUnbind)

	// Device status
	r.mux.HandleFunc("/api/v1/sleepace/devices/status", r.GetAllDeviceStatus)
	r.mux.HandleFunc("/api/v1/sleepace/devices/", r.deviceSubRoute)

	// Health
	r.mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
}

// deviceSubRoute handles /api/v1/sleepace/devices/{uid}/status
func (r *Router) deviceSubRoute(w http.ResponseWriter, req *http.Request) {
	const prefix = "/api/v1/sleepace/devices/"
	path := req.URL.Path[len(prefix):]

	idx := len(path) - len("/status")
	if idx > 0 && path[idx:] == "/status" {
		uid := path[:idx]
		r.GetDeviceStatus(w, req, uid)
		return
	}

	http.NotFound(w, req)
}

func (r *Router) GetDeviceStatus(w http.ResponseWriter, _ *http.Request, uid string) {
	if r.statusTracker == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "status tracker not available"})
		return
	}
	item := r.statusTracker.GetByUID(uid)
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": item})
}

func (r *Router) GetAllDeviceStatus(w http.ResponseWriter, _ *http.Request) {
	if r.statusTracker == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "status tracker not available"})
		return
	}
	items := r.statusTracker.GetAll()
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": items, "count": len(items)})
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
