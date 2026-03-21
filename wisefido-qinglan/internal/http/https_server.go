package http

import (
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"wisefido-qinglan/internal/config"
	"wisefido-qinglan/internal/models"
	"wisefido-qinglan/internal/repository"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
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
) (*HTTPSServer, error) {
	authService := NewAuthService(appConfig, db, deviceRepo, redisClient, logger, subscriptionManager)

	// 创建路由器，只注册 /auth 路由
	router := mux.NewRouter()
	authHandler := &AuthHTTPSHandler{authService: authService, logger: logger}
	router.HandleFunc("/auth", authHandler.handleAuth).Methods("POST")
	// 兼容其他路径（参考 wisefido-radar）
	router.HandleFunc("/prod-api/thirdmqtt/v2/auth/device", authHandler.handleAuth).Methods("POST")
	router.HandleFunc("/radar/api/v1/auth", authHandler.handleAuth).Methods("POST")
	router.HandleFunc("/", authHandler.handleAuth).Methods("POST")

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
	}, nil
}

// Start 启动 HTTPS 服务器
func (s *HTTPSServer) Start() error {
	s.logger.Info("Starting HTTPS server for device authentication",
		zap.String("addr", s.server.Addr),
		zap.Int("port", s.config.Port),
	)
	log.Printf("Starting HTTPS server for device authentication on %s", s.server.Addr)

	// 必须使用 TLS，禁止回退到 HTTP
	if s.server.TLSConfig == nil || len(s.server.TLSConfig.Certificates) == 0 {
		return fmt.Errorf("HTTPS server requires TLS certificate, but no certificate is configured")
	}

	return s.server.ListenAndServeTLS("", "")
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

	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode auth request body: %v", err)
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	log.Printf("Processing HTTPS auth request for device UID: %s", req.UID)

	ctx := r.Context()
	remoteAddr := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		remoteAddr = forwarded
	}

	response, err := h.authService.AuthenticateDevice(ctx, &req, remoteAddr)
	if err != nil {
		log.Printf("Authentication failed for device %s: %v", req.UID, err)
		http.Error(w, fmt.Sprintf("Authentication failed: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("HTTPS authentication completed for device %s, response code: %d", req.UID, response.Code)
	
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
