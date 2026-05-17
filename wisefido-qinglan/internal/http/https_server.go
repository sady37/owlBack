package http

import (
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"wisefido-qinglan/internal/config"
	"wisefido-qinglan/internal/models"
	"wisefido-qinglan/internal/ota"
	"wisefido-qinglan/internal/repository"
	"wisefido-qinglan/internal/service"
	"wisefido-qinglan/internal/tcp"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/soheilhy/cmux"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// HTTPSServer HTTPS 服务器（用于设备认证）
// 参考 wisefido-radar/internal/http/server.go 的实现
type HTTPSServer struct {
	config      *config.HTTPSConfig
	authService *AuthService
	server      *http.Server
	logger      *zap.Logger
	tcpServer   *tcp.Server
	otaManager  *ota.Manager
}

// NewHTTPSServer 创建 HTTPS 服务器
func NewHTTPSServer(
	cfg *config.HTTPSConfig,
	appConfig *config.Config,
	db *sql.DB,
	deviceRepo repository.DeviceRepository,
	redisClient *redis.Client,
	logger *zap.Logger,
	subscriptionManager DeviceSubscriptionManager,
	cardMapping *service.CardMappingService,
) (*HTTPSServer, error) {
	authService := NewAuthService(appConfig, db, deviceRepo, redisClient, logger, subscriptionManager, cardMapping)

	// Create firmware directory
	fwDir := filepath.Join("..", "ota")
	if err := os.MkdirAll(fwDir, 0755); err != nil {
		logger.Warn("create firmware dir", zap.String("dir", fwDir), zap.Error(err))
	}

	// Initialize TCP server for device connections (shares the HTTPS port via cmux)
	serverAddr := strings.TrimSpace(appConfig.MQTT.RadarDeviceMQTT.Server)
	if serverAddr == "" {
		serverAddr = "0.0.0.0"
	}
	tcpSrv := tcp.NewServer(serverAddr, uint32(cfg.Port))

	// Initialize OTA manager
	// Firmware download via nginx 443 (Let's Encrypt cert)
	fwURL := fmt.Sprintf("https://%s/ota", serverAddr)
	otaMgr := &ota.Manager{
		TCPServer:   tcpSrv,
		FirmwareDir: fwDir,
		FirmwareURL: fwURL,
	}

	// 创建路由器，只注册 /auth 路由
	router := mux.NewRouter()
	authHandler := &AuthHTTPSHandler{authService: authService, logger: logger}
	router.HandleFunc("/auth", authHandler.handleAuth).Methods("POST")
	// 兼容其他路径（参考 wisefido-radar）
	router.HandleFunc("/prod-api/thirdmqtt/v2/auth/device", authHandler.handleAuth).Methods("POST")
	router.HandleFunc("/radar/api/v1/auth", authHandler.handleAuth).Methods("POST")
	router.HandleFunc("/", authHandler.handleAuth).Methods("POST")

	// Serve firmware files for OTA download
	router.PathPrefix("/firmware/").Handler(http.StripPrefix("/firmware/", http.FileServer(http.Dir(fwDir))))
	// /ota/ path maps directly to owlBack/ota/ directory
	router.PathPrefix("/ota/").Handler(http.StripPrefix("/ota/", http.FileServer(http.Dir(fwDir))))

	// 创建 TLS 配置
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// 必须配置证书，禁止回退到 HTTP
	if cfg.CertFile == "" || cfg.KeyFile == "" {
		return nil, fmt.Errorf("HTTPS server requires certificate files: QINGLAN_HTTPS_CERT_FILE and QINGLAN_HTTPS_KEY_FILE must be set (or use wisefido-radar certificates)")
	}

	// 加载证书（支持相对路径，自动解析为绝对路径）
	certFile := cfg.CertFile
	keyFile := cfg.KeyFile
	
	// 如果是相对路径，尝试解析
	if !filepath.IsAbs(certFile) {
		if absPath, err := filepath.Abs(certFile); err == nil {
			if _, err := os.Stat(absPath); err == nil {
				certFile = absPath
			}
		}
	}
	if !filepath.IsAbs(keyFile) {
		if absPath, err := filepath.Abs(keyFile); err == nil {
			if _, err := os.Stat(absPath); err == nil {
				keyFile = absPath
			}
		}
	}

	// 加载证书
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificate from %s: %w", certFile, err)
	}
	tlsConfig.Certificates = []tls.Certificate{cert}
	logger.Info("loaded tls cert", zap.String("cert", certFile), zap.String("key", keyFile))

	addr := fmt.Sprintf(":%d", cfg.Port)
	tlsErrLogger, _ := zap.NewStdLogAt(logger.With(zap.String("comp", "http_tls")), zapcore.WarnLevel)
	s := &http.Server{
		Addr:              addr,
		Handler:           router,
		TLSConfig:         tlsConfig,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		ErrorLog:          tlsErrLogger,
	}

	return &HTTPSServer{
		config:      cfg,
		authService: authService,
		server:      s,
		logger:      logger,
		tcpServer:   tcpSrv,
		otaManager:  otaMgr,
	}, nil
}

// TCPServer returns the TCP server instance
func (s *HTTPSServer) TCPServer() *tcp.Server {
	return s.tcpServer
}

// OTAManager returns the OTA manager instance
func (s *HTTPSServer) OTAManager() *ota.Manager {
	return s.otaManager
}

// AuthService returns the auth service instance
func (s *HTTPSServer) AuthService() *AuthService {
	return s.authService
}

// Start 启动 HTTPS 服务器 (cmux: TLS connections go to HTTPS, raw TCP to the TCP server)
func (s *HTTPSServer) Start() error {
	s.logger.Info("starting https+tcp server (cmux)",
		zap.String("addr", s.server.Addr),
		zap.Int("port", s.config.Port),
	)

	// 必须使用 TLS，禁止回退到 HTTP
	if s.server.TLSConfig == nil || len(s.server.TLSConfig.Certificates) == 0 {
		return fmt.Errorf("HTTPS server requires TLS certificate, but no certificate is configured")
	}

	// Create a raw TCP listener
	ln, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", s.server.Addr, err)
	}

	// Create cmux multiplexer
	m := cmux.New(ln)

	// TLS connections go to HTTPS server
	tlsLn := m.Match(cmux.TLS())
	// Everything else (raw TCP protobuf) goes to TCP server
	tcpLn := m.Match(cmux.Any())

	// Wrap TLS listener with TLS config
	tlsListener := tls.NewListener(tlsLn, s.server.TLSConfig)

	// Start HTTPS server
	go func() {
		s.logger.Info("https serving on cmux", zap.String("addr", s.server.Addr))
		if err := s.server.Serve(tlsListener); err != nil && err != http.ErrServerClosed {
			s.logger.Warn("https server", zap.Error(err))
		}
	}()

	// Start TCP server
	go func() {
		s.logger.Info("tcp serving on cmux", zap.String("addr", s.server.Addr))
		if err := s.tcpServer.Serve(tcpLn); err != nil {
			s.logger.Warn("tcp server", zap.Error(err))
		}
	}()

	// Start cmux serving (blocks)
	return m.Serve()
}

// Stop 停止服务器
func (s *HTTPSServer) Stop(ctx context.Context) error {
	s.logger.Info("Stopping HTTPS server")
	return s.server.Shutdown(ctx)
}

// AuthHTTPSHandler HTTPS 认证请求处理器
type AuthHTTPSHandler struct {
	authService *AuthService
	logger      *zap.Logger
}

// handleAuth 处理设备认证请求
func (h *AuthHTTPSHandler) handleAuth(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("http auth request",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("x_forwarded_for", r.Header.Get("X-Forwarded-For")),
	)

	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("http auth decode failed", zap.Error(err))
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	h.logger.Info("http auth body decoded", zap.String("uid", req.UID), zap.Int("type", req.Type))

	ctx := r.Context()
	remoteAddr := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		remoteAddr = forwarded
	}

	response, err := h.authService.AuthenticateDevice(ctx, &req, remoteAddr)
	if err != nil {
		h.logger.Error("http auth AuthenticateDevice", zap.String("uid", req.UID), zap.Error(err))
		http.Error(w, fmt.Sprintf("Authentication failed: %v", err), http.StatusInternalServerError)
		return
	}

	h.logger.Info("http auth response",
		zap.String("uid", req.UID),
		zap.Int("code", response.Code),
		zap.String("msg", response.Msg),
	)

	if response.Data != nil && response.Data.MQTT != nil {
		h.logger.Debug("https auth response payload",
			zap.String("uid", req.UID),
			zap.String("server", response.Data.MQTT.Server),
			zap.Int("port", response.Data.MQTT.Port),
			zap.String("protocol", response.Data.MQTT.Protocol),
			zap.String("client_id", response.Data.MQTT.ClientID),
			zap.String("account", response.Data.MQTT.Account),
		)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
