package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// DoctorHandler 诊断处理器
type DoctorHandler struct {
	db           *sql.DB
	redisClient  *redis.Client
	logger       *zap.Logger
	pprofEnabled bool
}

// NewDoctorHandler 创建诊断处理器
func NewDoctorHandler(db *sql.DB, redisClient *redis.Client, logger *zap.Logger) *DoctorHandler {
	return &DoctorHandler{
		db:          db,
		redisClient: redisClient,
		logger:      logger,
	}
}

// EnablePprof 启用 pprof 性能分析
func (d *DoctorHandler) EnablePprof(enabled bool) {
	d.pprofEnabled = enabled
}

// HealthCheckResponse 健康检查响应
type HealthCheckResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

// HealthCheck 健康检查端点
func (d *DoctorHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := "healthy"
	services := make(map[string]string)

	// 检查 Redis
	if d.redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := d.redisClient.Ping(ctx).Err(); err != nil {
			status = "unhealthy"
			services["redis"] = "unhealthy: " + err.Error()
		} else {
			services["redis"] = "healthy"
		}
	} else {
		services["redis"] = "not configured"
	}

	// 检查数据库
	if d.db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := d.db.PingContext(ctx); err != nil {
			status = "unhealthy"
			services["database"] = "unhealthy: " + err.Error()
		} else {
			services["database"] = "healthy"
		}
	} else {
		services["database"] = "not configured"
	}

	response := HealthCheckResponse{
		Status:    status,
		Timestamp: time.Now(),
		Services:  services,
	}

	statusCode := http.StatusOK
	if status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// Ready 就绪检查（用于 Kubernetes liveness/readiness probes）
func (d *DoctorHandler) Ready(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ready := true
	checks := make(map[string]bool)

	// 检查 Redis
	if d.redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		checks["redis"] = d.redisClient.Ping(ctx).Err() == nil
		if !checks["redis"] {
			ready = false
		}
	} else {
		checks["redis"] = false
		ready = false
	}

	// 如果启用了数据库，检查数据库
	if d.db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		checks["database"] = d.db.PingContext(ctx) == nil
		if !checks["database"] {
			ready = false
		}
	} else {
		checks["database"] = true // DB 是可选的
	}

	statusCode := http.StatusOK
	if !ready {
		statusCode = http.StatusServiceUnavailable
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ready":  ready,
		"checks": checks,
	})
}

// RegisterDoctorRoutes 注册诊断路由
func (r *Router) RegisterDoctorRoutes(doctor *DoctorHandler) {
	// 健康检查
	r.Handle("/health", doctor.HealthCheck)
	r.Handle("/healthz", doctor.HealthCheck)

	// 就绪检查
	r.Handle("/ready", doctor.Ready)
	r.Handle("/readyz", doctor.Ready)

	// pprof 性能分析（如果启用）
	if doctor.pprofEnabled {
		r.Handle("/debug/pprof/", pprof.Index)
		r.Handle("/debug/pprof/cmdline", pprof.Cmdline)
		r.Handle("/debug/pprof/profile", pprof.Profile)
		r.Handle("/debug/pprof/symbol", pprof.Symbol)
		r.Handle("/debug/pprof/trace", pprof.Trace)
		// 额外的 pprof 端点（使用 HandleHandler 因为返回的是 http.Handler）
		r.HandleHandler("/debug/pprof/goroutine", pprof.Handler("goroutine"))
		r.HandleHandler("/debug/pprof/heap", pprof.Handler("heap"))
		r.HandleHandler("/debug/pprof/allocs", pprof.Handler("allocs"))
		r.HandleHandler("/debug/pprof/block", pprof.Handler("block"))
		r.HandleHandler("/debug/pprof/mutex", pprof.Handler("mutex"))
	}
}

