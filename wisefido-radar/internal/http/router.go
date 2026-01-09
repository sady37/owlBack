package http

import (
	"net/http"
	
	"go.uber.org/zap"
)

// Router HTTP 路由器
// 参考 wisefido-data/internal/http/router.go 的实现
type Router struct {
	mux    *http.ServeMux
	logger *zap.Logger
}

// NewRouter 创建路由器
func NewRouter(logger *zap.Logger) *Router {
	return &Router{
		mux:    http.NewServeMux(),
		logger: logger,
	}
}

// Handle 注册路由处理器
func (r *Router) Handle(pattern string, handler http.Handler) {
	r.mux.Handle(pattern, handler)
}

// ServeHTTP 实现 http.Handler 接口
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// RegisterAuthRoutes 注册认证路由
func (r *Router) RegisterAuthRoutes(h *AuthHandler) {
	// POST /prod-api/thirdmqtt/v2/auth/device - 设备认证（协议文档标准路径）
	r.Handle("/prod-api/thirdmqtt/v2/auth/device", h)
	// 兼容旧路径（保留向后兼容）
	r.Handle("/radar/api/v1/auth", h)
	// 兼容根路径（参考 simple-https.py 监听根路径）
	r.Handle("/", h)
}

// RegisterCommandRoutes 注册命令路由（内部 API）
func (r *Router) RegisterCommandRoutes(h *CommandHandler) {
	// 属性读取
	r.mux.HandleFunc("/internal/api/v1/radar/devices/", func(w http.ResponseWriter, req *http.Request) {
		// 路由分发逻辑
		path := req.URL.Path
		if endsWith(path, "/properties/get") {
			h.GetProperties(w, req)
		} else if endsWith(path, "/properties/set") {
			h.SetProperties(w, req)
		} else if endsWith(path, "/realtime/subscribe") {
			h.SubscribeRealtime(w, req)
		} else if endsWith(path, "/commands") {
			h.SendCommand(w, req)
		} else {
			http.NotFound(w, req)
		}
	})
}

// endsWith 辅助函数
func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

