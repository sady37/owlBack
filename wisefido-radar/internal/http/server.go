package http

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"
	
	"go.uber.org/zap"
)

// Server HTTPS 服务器
// 参考 radar-ql-v3/simple-https.py 的 HTTPS 服务器实现
type Server struct {
	httpServer *http.Server
	logger     *zap.Logger
}

// NewServer 创建 HTTPS 服务器
func NewServer(addr string, handler http.Handler, certFile, keyFile string, logger *zap.Logger) (*Server, error) {
	// 创建 TLS 配置
	// 注意：设备端不校验服务器证书，仅用于传输加密（参考测试文档）
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		// 允许使用自签名证书
		InsecureSkipVerify: false, // 服务器端不需要跳过验证
	}
	
	// 加载证书
	if certFile != "" && keyFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	
	s := &http.Server{
		Addr:              addr,
		Handler:           handler,
		TLSConfig:         tlsConfig,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}
	
	return &Server{
		httpServer: s,
		logger:     logger,
	}, nil
}

// Start 启动 HTTPS 服务器
func (s *Server) Start() error {
	s.logger.Info("Starting HTTPS server",
		zap.String("addr", s.httpServer.Addr),
	)
	
	// 如果配置了证书，使用 TLS；否则使用普通 HTTP（仅用于开发）
	if s.httpServer.TLSConfig != nil && len(s.httpServer.TLSConfig.Certificates) > 0 {
		return s.httpServer.ListenAndServeTLS("", "")
	} else {
		s.logger.Warn("HTTPS server running without TLS (development mode)")
		return s.httpServer.ListenAndServe()
	}
}

// Stop 停止服务器
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping HTTPS server")
	return s.httpServer.Shutdown(ctx)
}

