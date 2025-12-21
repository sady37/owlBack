package service

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Server struct {
	httpServer *http.Server
	logger     *zap.Logger
}

func NewServer(addr string, handler http.Handler, logger *zap.Logger) *Server {
	s := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return &Server{httpServer: s, logger: logger}
}

func (s *Server) Start() error {
	s.logger.Info("Starting wisefido-data HTTP server", zap.String("addr", s.httpServer.Addr))
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping wisefido-data HTTP server")
	return s.httpServer.Shutdown(ctx)
}

