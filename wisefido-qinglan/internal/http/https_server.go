package http

import (
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
)

// tlsErrorFilter 过滤 TLS handshake error 刷屏（外网扫描不信任自签名证书）
type tlsErrorFilter struct {
	w io.Writer
}

func (f *tlsErrorFilter) Write(p []byte) (n int, err error) {
	if strings.Contains(string(p), "TLS handshake error") {
		return len(p), nil
	}
	return f.w.Write(p)
}

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
		log.Printf("Warning: failed to create firmware dir %s: %v", fwDir, err)
	}

	// Initialize TCP server for device connections (shares the HTTPS port via cmux)
	serverAddr := strings.TrimSpace(appConfig.MQTT.RadarDeviceMQTT.Server)
	if serverAddr == "" {
		serverAddr = "0.0.0.0"
	}
	tcpSrv := tcp.NewServer(serverAddr, uint32(cfg.Port))

	// Initialize OTA manager
	fwURL := fmt.Sprintf("https://%s:%d/firmware", serverAddr, cfg.Port)
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
	log.Printf("Loaded TLS certificate: %s, key: %s", certFile, keyFile)

	addr := fmt.Sprintf(":%d", cfg.Port)
	s := &http.Server{
		Addr:              addr,
		Handler:           router,
		TLSConfig:         tlsConfig,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		ErrorLog:          log.New(&tlsErrorFilter{w: os.Stderr}, "", 0),
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
	s.logger.Info("Starting HTTPS+TCP server (cmux)",
		zap.String("addr", s.server.Addr),
		zap.Int("port", s.config.Port),
	)
	log.Printf("Starting HTTPS+TCP server (cmux) on %s", s.server.Addr)

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
		log.Printf("HTTPS server serving on %s (TLS via cmux)", s.server.Addr)
		if err := s.server.Serve(tlsListener); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTPS server error: %v", err)
		}
	}()

	// Start TCP server
	go func() {
		log.Printf("TCP server serving on %s (raw via cmux)", s.server.Addr)
		if err := s.tcpServer.Serve(tcpLn); err != nil {
			log.Printf("TCP server error: %v", err)
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
	log.Printf("Received HTTPS auth request from %s", r.RemoteAddr)
	if h.logger != nil {
		h.logger.Info("http auth request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("x_forwarded_for", r.Header.Get("X-Forwarded-For")),
		)
	}

	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode auth request body: %v", err)
		if h.logger != nil {
			h.logger.Warn("http auth decode failed", zap.Error(err))
		}
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	log.Printf("Processing HTTPS auth request for device UID: %s", req.UID)
	if h.logger != nil {
		h.logger.Info("http auth body decoded", zap.String("uid", req.UID), zap.Int("type", req.Type))
	}

	ctx := r.Context()
	remoteAddr := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		remoteAddr = forwarded
	}

	response, err := h.authService.AuthenticateDevice(ctx, &req, remoteAddr)
	if err != nil {
		log.Printf("Authentication failed for device %s: %v", req.UID, err)
		if h.logger != nil {
			h.logger.Error("http auth AuthenticateDevice error", zap.String("uid", req.UID), zap.Error(err))
		}
		http.Error(w, fmt.Sprintf("Authentication failed: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("HTTPS authentication completed for device %s, response code: %d", req.UID, response.Code)
	if h.logger != nil {
		h.logger.Info("http auth response",
			zap.String("uid", req.UID),
			zap.Int("code", response.Code),
			zap.String("msg", response.Msg),
		)
	}
	
	// 将响应序列化为 JSON 并原封不动地打印到 log（用于检查真实发送的信息）
	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Printf("⚠️ Failed to marshal response to JSON: %v", err)
	} else {
		log.Printf("📤 [HTTPS Response to Device] Complete JSON response for device %s:\n%s", req.UID, string(responseJSON))
	}
	
	// 输出返回给设备的完整响应内容（用于调试）
	if response.Data != nil && response.Data.MQTT != nil {
		log.Printf("📤 Returning auth response to device %s: msg=%s, code=%d, server=%s, port=%d, protocol=%s, prefix=%s, productId=%s, clientId=%s, account=%s",
			req.UID,
			response.Msg,
			response.Code,
			response.Data.MQTT.Server,
			response.Data.MQTT.Port,
			response.Data.MQTT.Protocol,
			response.Data.MQTT.Prefix,
			response.Data.MQTT.ProductID,
			response.Data.MQTT.ClientID,
			response.Data.MQTT.Account,
		)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
