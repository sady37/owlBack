package tcp

import (
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
)

// Server TCP server for V2 qinglan device connections
type Server struct {
	Sessions   *SessionManager
	ServerAddr string
	ServerPort uint32
	OnProgress OTAProgressCallback
	OnRegister OnRegisterCallback
}

// NewServer creates TCP server
func NewServer(serverAddr string, serverPort uint32) *Server {
	return &Server{
		Sessions:   NewSessionManager(),
		ServerAddr: serverAddr,
		ServerPort: serverPort,
	}
}

// SetLogger injects the zap logger into the underlying SessionManager and frame handlers.
func (s *Server) SetLogger(logger *zap.Logger) {
	s.Sessions.SetLogger(logger)
}

// Serve accepts connections from a listener (cmux mode)
func (s *Server) Serve(listener net.Listener) error {
	lg := s.Sessions.logger
	lg.Info("tcp ota server started",
		zap.String("dispatch_addr", s.ServerAddr),
		zap.Uint32("dispatch_port", s.ServerPort),
	)
	for {
		conn, err := listener.Accept()
		if err != nil {
			lg.Warn("tcp accept", zap.Error(err))
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	lg := s.Sessions.logger
	defer func() {
		uid := s.Sessions.GetUIDByConn(conn)
		s.Sessions.Disconnect(conn)
		conn.Close()
		if uid != "" {
			lg.Info("tcp connection closed", zap.String("uid", uid))
		}
	}()

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	firstFrame, err := ReadFrame(conn)
	if err != nil {
		return
	}

	if firstFrame.Type != TypeGetServer && firstFrame.Type != TypeRegister {
		lg.Warn("tcp invalid first frame",
			zap.Uint8("type", firstFrame.Type),
			zap.String("remote", conn.RemoteAddr().String()),
		)
		return
	}

	lg.Info("tcp new connection",
		zap.String("remote", conn.RemoteAddr().String()),
		zap.Uint8("first_type", firstFrame.Type),
	)
	s.Sessions.UpdateHeartbeat(conn)
	HandleFrame(conn, firstFrame, s.Sessions, s.ServerAddr, s.ServerPort, s.OnProgress, s.OnRegister)

	if firstFrame.Type == TypeGetServer && s.Sessions.GetUIDByConn(conn) == "" {
		conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		nextFrame, err := ReadFrame(conn)
		if err != nil {
			return
		}
		s.Sessions.UpdateHeartbeat(conn)
		HandleFrame(conn, nextFrame, s.Sessions, s.ServerAddr, s.ServerPort, s.OnProgress, s.OnRegister)
	}

	for {
		conn.SetReadDeadline(time.Now().Add(90 * time.Second))
		frame, err := ReadFrame(conn)
		if err != nil {
			uid := s.Sessions.GetUIDByConn(conn)
			if uid != "" {
				lg.Info("tcp disconnect", zap.String("uid", uid), zap.Error(err))
			}
			return
		}
		s.Sessions.UpdateHeartbeat(conn)
		HandleFrame(conn, frame, s.Sessions, s.ServerAddr, s.ServerPort, s.OnProgress, s.OnRegister)
	}
}

// PushOTA sends OTA frame to device
func (s *Server) PushOTA(uid string, frame *Frame) error {
	session := s.Sessions.GetByUID(uid)
	if session == nil {
		return fmt.Errorf("device offline: %s", uid)
	}
	return WriteFrame(session.Conn, frame)
}
