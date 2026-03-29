package httpapi

import (
	"context"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"wisefido-data/internal/service"
)

// SkippedPath 返回 true 表示该路径跳过认证
type SkippedPath func(path string) bool

// DefaultSkippedPaths 默认跳过认证的路径
func DefaultSkippedPaths(path string) bool {
	switch {
	case path == "/auth/api/v1/login",
		path == "/auth/api/v1/institutions/search",
		strings.HasPrefix(path, "/auth/api/v1/forgot-password/"),
		path == "/auth/api/v1/verify-pin",
		strings.HasPrefix(path, "/doctor"),
		strings.HasPrefix(path, "/internal/"):
		return true
	}
	return false
}

// AuthMiddleware 校验 token 并注入可信的 X-User-Id, X-Tenant-Id, X-User-Type, X-User-Role
type AuthMiddleware struct {
	store   SessionStore
	skipped SkippedPath
	logger  *zap.Logger
	enabled bool
}

// AuthMiddlewareConfig 中间件配置
type AuthMiddlewareConfig struct {
	Store   SessionStore
	Skipped SkippedPath
	Logger  *zap.Logger
	Enabled bool // 为 false 时跳过校验（兼容旧环境，如 AUTH_MIDDLEWARE_ENABLED=false）
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(cfg AuthMiddlewareConfig) *AuthMiddleware {
	skipped := cfg.Skipped
	if skipped == nil {
		skipped = DefaultSkippedPaths
	}
	return &AuthMiddleware{
		store:   cfg.Store,
		skipped: skipped,
		logger:  cfg.Logger,
		enabled: cfg.Enabled,
	}
}

// Wrap 包装 handler，在调用前校验 token 并注入 header
func (m *AuthMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS: 允许前端开发服务器跨域访问（SSE 直连后端时需要）
		// 注意：EventSource 不支持 withCredentials，所以不设 Credentials header
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type")
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if !m.enabled {
			next.ServeHTTP(w, r)
			return
		}
		if m.skipped(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		token := extractBearerToken(r)
		// SSE 连接时，EventSource 无法发送 Authorization header，
		// 允许从查询参数 ?token=xxx 中提取（仅用于 SSE）
		if token == "" && strings.Contains(r.URL.Path, "/stream") {
			token = r.URL.Query().Get("token")
		}
		if token == "" {
			writeJSON(w, http.StatusUnauthorized, Fail("missing or invalid authorization"))
			return
		}
		session, err := m.store.Get(r.Context(), token)
		if err != nil || session == nil {
			writeJSON(w, http.StatusUnauthorized, Fail("invalid or expired token"))
			return
		}
		// 注入可信 header，覆盖客户端提供的值
		r.Header.Set("X-User-Id", session.UserID)
		r.Header.Set("X-Tenant-Id", session.TenantID)
		r.Header.Set("X-User-Type", session.UserType)
		r.Header.Set("X-User-Role", session.Role)

		// 同时把可信值注入到 request context，便于 service 层直接读取
		ctx := context.WithValue(r.Context(), service.ContextKeyUserID, session.UserID)
		ctx = context.WithValue(ctx, service.ContextKeyTenantID, session.TenantID)
		ctx = context.WithValue(ctx, service.ContextKeyUserType, session.UserType)
		ctx = context.WithValue(ctx, service.ContextKeyUserRole, session.Role)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func extractBearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if h == "" {
		return ""
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(h, prefix) {
		return ""
	}
	return strings.TrimSpace(h[len(prefix):])
}
