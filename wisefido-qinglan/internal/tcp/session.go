package tcp

import (
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
)

// DeviceSession 设备会话，维护UID与TCP连接的映射
type DeviceSession struct {
	UID       string    // 设备UID
	Conn      net.Conn  // TCP连接
	Type      string    // 设备类型: TSL60G442 / TSL60G4G
	SfVer     string    // espsfver软件版本: 2.0(HC2), 2.6(TK2)
	HwVer     string    // v4.7 P3: MCU硬件版本: 2.0(HC2), 2.3(TK2)
	ConnectAt time.Time // 连接时间
	LastHeart time.Time // 最后心跳时间
}

// SessionManager 管理所有设备会话
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*DeviceSession // key: UID
	connMap  map[net.Conn]string       // key: conn → UID (反向索引)
	logger   *zap.Logger
	// v12.2 P1: 设备断线回调（在Connect()关闭旧连接时主动触发）
	OnDisconnect func(uid string)
}

// NewSessionManager 创建会话管理器
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*DeviceSession),
		connMap:  make(map[net.Conn]string),
		logger:   zap.NewNop(),
	}
}

// SetLogger injects the zap logger (default is no-op).
func (sm *SessionManager) SetLogger(logger *zap.Logger) {
	if logger != nil {
		sm.logger = logger
	}
}

// Connect 注册设备连接
func (sm *SessionManager) Connect(conn net.Conn, uid, deviceType, sfVer, hwVer string) {
	var disconnCb func(string)
	var disconnUID string

	sm.mu.Lock()

	if old, ok := sm.sessions[uid]; ok {
		sm.logger.Warn("device duplicate register, closing old conn",
			zap.String("uid", uid),
			zap.String("old_addr", old.Conn.RemoteAddr().String()),
		)
		// v12.2 P1: 在删除旧连接前，记录需要通知断线的UID
		// 根因: 如果先delete(connMap)再Close()，旧goroutine的defer中GetUIDByConn返回""
		//       导致NotifyDisconnect永远不被调用，OTA monitor永远检测不到断线
		// 修复: 记录断线回调，在释放锁后调用（避免持锁调用外部函数导致潜在死锁）
		if sm.OnDisconnect != nil {
			disconnCb = sm.OnDisconnect
			disconnUID = uid
		}
		delete(sm.connMap, old.Conn)
		old.Conn.Close()
	}

	now := time.Now()
	session := &DeviceSession{
		UID:       uid,
		Conn:      conn,
		Type:      deviceType,
		SfVer:     sfVer,
		HwVer:     hwVer,
		ConnectAt: now,
		LastHeart: now,
	}

	sm.sessions[uid] = session
	sm.connMap[conn] = uid
	sm.logger.Info("device registered",
		zap.String("uid", uid),
		zap.String("type", deviceType),
		zap.String("sf_ver", sfVer),
		zap.String("hw_ver", hwVer),
		zap.String("addr", conn.RemoteAddr().String()),
	)
	sm.mu.Unlock()

	// v12.2 P1: 释放锁后调用断线回调（避免持锁调用外部函数导致潜在死锁）
	// 时序：先完成新连接注册 → 释放锁 → 通知OTA服务旧连接断线
	// 这确保NotifyReconnect（在TypeRegister handler中被调用）看到Disconnected=true
	if disconnCb != nil {
		disconnCb(disconnUID)
	}
}

// Disconnect 断开设备连接
// v5.5.1: 添加连接对象匹配验证，防止旧goroutine清除新连接注册的session
// 竞争场景: 设备重连后 Connect() 用新conn替换了sessions[uid]，
//
//	但旧conn的goroutine defer仍会调用 Disconnect(旧conn)，
//	必须验证 sessions[uid].Conn == conn 才能删除。
func (sm *SessionManager) Disconnect(conn net.Conn) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	uid, ok := sm.connMap[conn]
	if !ok {
		return
	}

	// 关键保护：只有当 sessions[uid] 的连接仍然是当前 conn 时才删除 session
	// 如果 sessions[uid].Conn != conn，说明 Connect() 已经用新连接替换了，
	// 旧连接的 Disconnect 不应清除新连接的 session
	if session, exists := sm.sessions[uid]; exists {
		if session.Conn == conn {
			delete(sm.sessions, uid)
			sm.logger.Info("device disconnected",
				zap.String("uid", uid),
				zap.String("addr", conn.RemoteAddr().String()),
			)
		} else {
			sm.logger.Info("stale conn closed, session retained for newer conn",
				zap.String("uid", uid),
				zap.String("old_addr", conn.RemoteAddr().String()),
				zap.String("new_addr", session.Conn.RemoteAddr().String()),
			)
		}
	}
	// 无论如何都清除 connMap 中的旧连接映射
	delete(sm.connMap, conn)
}

// GetByUID 根据UID获取设备会话
func (sm *SessionManager) GetByUID(uid string) *DeviceSession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.sessions[uid]
}

// GetUIDByConn 根据连接获取UID
func (sm *SessionManager) GetUIDByConn(conn net.Conn) string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.connMap[conn]
}

// UpdateHeartbeat 更新心跳时间
func (sm *SessionManager) UpdateHeartbeat(conn net.Conn) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	uid, ok := sm.connMap[conn]
	if !ok {
		return
	}
	if session, ok := sm.sessions[uid]; ok {
		session.LastHeart = time.Now()
	}
}

// GetAllSessions 获取所有在线设备（用于Web API）
func (sm *SessionManager) GetAllSessions() []*DeviceSession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make([]*DeviceSession, 0, len(sm.sessions))
	for _, s := range sm.sessions {
		result = append(result, s)
	}
	return result
}

// DisconnectByUID v14.5 P4: 按UID强制断开TCP连接（黑名单踢设备用）
// 返回true表示找到并断开了该设备的连接
func (sm *SessionManager) DisconnectByUID(uid string) bool {
	sm.mu.Lock()
	session, ok := sm.sessions[uid]
	if !ok {
		sm.mu.Unlock()
		return false
	}
	conn := session.Conn
	delete(sm.sessions, uid)
	delete(sm.connMap, conn)
	sm.mu.Unlock()

	conn.Close()
	sm.logger.Info("force disconnect by uid (security)",
		zap.String("uid", uid),
		zap.String("addr", conn.RemoteAddr().String()),
	)

	// 通知OTA可靠性服务设备断线
	if sm.OnDisconnect != nil {
		sm.OnDisconnect(uid)
	}
	return true
}

// OnlineCount 在线设备数
func (sm *SessionManager) OnlineCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sessions)
}
