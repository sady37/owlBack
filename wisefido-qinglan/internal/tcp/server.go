package tcp

import (
	"fmt"
	"log"
	"net"
	"time"
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

// Serve accepts connections from a listener (cmux mode)
func (s *Server) Serve(listener net.Listener) error {
	log.Printf("[TCP] TCP OTA server started, dispatch addr: %s:%d", s.ServerAddr, s.ServerPort)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("[TCP] accept connection failed: %v", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		uid := s.Sessions.GetUIDByConn(conn)
		s.Sessions.Disconnect(conn)
		conn.Close()
		if uid != "" {
			log.Printf("[TCP] connection closed: UID=%s", uid)
		}
	}()

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	firstFrame, err := ReadFrame(conn)
	if err != nil {
		return
	}

	if firstFrame.Type != TypeGetServer && firstFrame.Type != TypeRegister {
		log.Printf("[TCP] invalid first frame type=%d, closing: %s", firstFrame.Type, conn.RemoteAddr())
		return
	}

	log.Printf("[TCP] new connection: %s (first frame type=%d)", conn.RemoteAddr(), firstFrame.Type)
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
				log.Printf("[TCP] disconnect: UID=%s reason=%v", uid, err)
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
