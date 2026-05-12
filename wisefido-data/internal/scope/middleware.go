package scope

import (
	"database/sql"
	"net/http"
)

// Middleware — 装在 auth middleware **之后**，把 ScopeContext 注入 r.Context()。
//
// 上游 auth middleware 已经把 session 信息注入 X-User-Id / X-User-Role / X-Tenant-Id /
// X-User-HoA headers。本 middleware 读 header → DB 查 current_branch → 写 ctx。
//
// 装载位置（router.go）：
//
//	r.Use(authMiddleware)        // 先 auth
//	r.Use(scope.Middleware(db))  // 再 scope
//	r.Handle(...)                // handler 拿 scope.FromContext(ctx)
//
// 即便 Resolve 出错也不阻断请求（保持向后兼容；下游 handler nil-check 后回落到自己拿 header）。
func Middleware(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := r.Header.Get("X-User-Id")
			role := r.Header.Get("X-User-Role")
			tenant := r.Header.Get("X-Tenant-Id")
			hoa := r.Header.Get("X-User-HoA")

			// 未登录 / 内部请求（/internal/* / /iot/* 等）userID 为空，跳过
			if userID == "" && role == "" {
				next.ServeHTTP(w, r)
				return
			}
			sc, _ := Resolve(r.Context(), db, userID, role, tenant, hoa)
			if sc != nil {
				r = r.WithContext(WithContext(r.Context(), sc))
			}
			next.ServeHTTP(w, r)
		})
	}
}
