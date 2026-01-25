package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	commonconfig "owl-common/config"
	"wisefido-qinglan/internal/config"
	"wisefido-qinglan/internal/models"
	"wisefido-qinglan/internal/repository"
	"wisefido-qinglan/internal/service"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// DeviceSubscriptionManager 设备订阅管理器接口（与auth_service.go中的定义一致）
type DeviceSubscriptionManager interface {
	SubscribeDevice(ctx context.Context, deviceUID, deviceID string) error
	EnablePeriodicSubscription(ctx context.Context, deviceUID, deviceID string) error
	ClearForceUnsubscribed(deviceUID string)
}

// Server HTTP服务器
type Server struct {
	config       *commonconfig.HTTPConfig
	radarService *service.RadarService
	authService  *AuthService
	server       *http.Server
}

// NewServer 创建HTTP服务器
func NewServer(
	cfg *commonconfig.HTTPConfig,
	radarService *service.RadarService,
	appConfig *config.Config,
	db *sql.DB,
	deviceRepo repository.DeviceRepository,
	redisClient *redis.Client,
	logger *zap.Logger,
	subscriptionManager DeviceSubscriptionManager,
) *Server {
	authService := NewAuthService(appConfig, db, deviceRepo, redisClient, logger, subscriptionManager)
	return &Server{
		config:       cfg,
		radarService: radarService,
		authService:  authService,
	}
}

// Start 启动HTTP服务器
func (s *Server) Start() error {
	router := mux.NewRouter()
	
	// 创建API处理器
	apiHandler := NewAPIHandler(s.radarService)
	apiHandler.RegisterRoutes(router)
	
	// 注意：auth 路由已移至独立的 HTTPS 服务器，此处不再注册
	
	// 创建HTTP服务器
	s.server = &http.Server{
		Addr:         s.config.GetAddr(),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	log.Printf("Starting HTTP server on %s", s.config.GetAddr())
	
	// 启动服务器
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server error: %w", err)
	}
	
	return nil
}

// handleAuth 处理设备认证请求
func (s *Server) handleAuth(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received auth request from %s", r.RemoteAddr)
	
	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode auth request body: %v", err)
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	log.Printf("Processing auth request for device UID: %s", req.UID)

	ctx := r.Context()
	remoteAddr := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		remoteAddr = forwarded
	}

	response, err := s.authService.AuthenticateDevice(ctx, &req, remoteAddr)
	if err != nil {
		log.Printf("Authentication failed for device %s: %v", req.UID, err)
		http.Error(w, fmt.Sprintf("Authentication failed: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Authentication completed for device %s, response code: %d", req.UID, response.Code)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Shutdown 优雅关闭HTTP服务器
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}