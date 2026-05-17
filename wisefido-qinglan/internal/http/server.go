package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	commonconfig "owl-common/config"
	"wisefido-qinglan/internal/config"
	"wisefido-qinglan/internal/domain"
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
	GetDeviceOnlineStatus(deviceUID string) string
	GetDeviceOnlineStatusByDeviceID(ctx context.Context, deviceID string) (string, error)
	GetAllDeviceStatuses(tenantID string) []domain.DeviceStatusItem
}

// Server HTTP服务器
type Server struct {
	config              *commonconfig.HTTPConfig
	radarService        *service.RadarService
	authService         *AuthService
	subscriptionManager DeviceSubscriptionManager
	server              *http.Server
	otaHandler          *OTAHandler
	logger              *zap.Logger
}

// SetOTAHandler sets the OTA handler for the HTTP server
func (s *Server) SetOTAHandler(h *OTAHandler) {
	s.otaHandler = h
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
	cardMapping *service.CardMappingService,
) *Server {
	if logger == nil {
		logger = zap.NewNop()
	}
	authService := NewAuthService(appConfig, db, deviceRepo, redisClient, logger, subscriptionManager, cardMapping)
	return &Server{
		config:              cfg,
		radarService:        radarService,
		authService:         authService,
		subscriptionManager: subscriptionManager,
		logger:              logger,
	}
}

// Start 启动HTTP服务器
func (s *Server) Start() error {
	router := mux.NewRouter()

	apiHandler := NewAPIHandler(s.radarService, s.subscriptionManager, s.logger)
	apiHandler.RegisterRoutes(router)

	if s.otaHandler != nil {
		s.otaHandler.RegisterRoutes(router)
	}

	fwDir := filepath.Join("..", "ota")
	router.PathPrefix("/firmware/").Handler(http.StripPrefix("/firmware/", http.FileServer(http.Dir(fwDir))))
	s.logger.Info("firmware http server",
		zap.String("path", "/firmware/"),
		zap.String("dir", fwDir),
	)

	s.server = &http.Server{
		Addr:         s.config.GetAddr(),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.logger.Info("starting http server", zap.String("addr", s.config.GetAddr()))
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server error: %w", err)
	}
	return nil
}

// handleAuth 处理设备认证请求
func (s *Server) handleAuth(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("auth request received", zap.String("remote", r.RemoteAddr))

	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Warn("decode auth body", zap.Error(err))
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	s.logger.Info("processing auth request", zap.String("uid", req.UID))

	ctx := r.Context()
	remoteAddr := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		remoteAddr = forwarded
	}

	response, err := s.authService.AuthenticateDevice(ctx, &req, remoteAddr)
	if err != nil {
		s.logger.Warn("auth failed", zap.String("uid", req.UID), zap.Error(err))
		http.Error(w, fmt.Sprintf("Authentication failed: %v", err), http.StatusInternalServerError)
		return
	}

	s.logger.Info("auth complete", zap.String("uid", req.UID), zap.Int("code", response.Code))
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
