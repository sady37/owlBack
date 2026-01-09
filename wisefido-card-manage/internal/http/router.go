package http

import (
	"net/http"
)

// Router HTTP 路由
type Router struct {
	mux     *http.ServeMux
	handler *Handler
}

// NewRouter 创建路由
func NewRouter(handler *Handler) *Router {
	return &Router{
		mux:     http.NewServeMux(),
		handler: handler,
	}
}

// RegisterRoutes 注册路由
func (r *Router) RegisterRoutes() {
	r.mux.HandleFunc("/api/v1/cards/create", r.handler.CreateCardsForUnit)
	r.mux.HandleFunc("/api/v1/cards/create-all", r.handler.CreateAllCards)
	r.mux.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}

// ServeHTTP 实现 http.Handler 接口
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

