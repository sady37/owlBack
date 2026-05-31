package httpapi

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// SessionData 登录会话数据（供 middleware 注入到请求 header）
//
// 字段约定：
//   - TenantID 历史字段，v1 流程存 UUID 字符串，v2 流程存 prefix CIDR 字符串（兼容期占位）
//   - TenantPrefix v2 专用，IPv6 /48 CIDR 字符串（如 "fd00:0:3::/48"）；middleware 注入 X-Tenant-Id
//   - HoA  resident/caregiver 自身 /128 IPv6 host address；staff/admin 为空
type SessionData struct {
	UserID       string
	TenantID     string
	TenantPrefix string
	UserType     string // 恒为 "staff"：v2 仅 users 表账号可登录，resident 不登录（Family 靠 Role 区分）
	Role         string
	HoA          string
}

// SessionStore 会话存储：Login 写入，AuthMiddleware 读取
type SessionStore interface {
	Set(ctx context.Context, token string, data SessionData, ttl time.Duration) error
	Get(ctx context.Context, token string) (*SessionData, error)
}

// MemSessionStore 内存实现（单实例部署适用；多实例需用 Redis）
type MemSessionStore struct {
	mu      sync.RWMutex
	entries map[string]sessionEntry
}

type sessionEntry struct {
	data SessionData
	exp  time.Time
}

// NewMemSessionStore 创建内存会话存储
func NewMemSessionStore() *MemSessionStore {
	s := &MemSessionStore{entries: make(map[string]sessionEntry)}
	go s.cleanupLoop()
	return s
}

func (s *MemSessionStore) Set(ctx context.Context, token string, data SessionData, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[token] = sessionEntry{data: data, exp: time.Now().Add(ttl)}
	return nil
}

func (s *MemSessionStore) Get(ctx context.Context, token string) (*SessionData, error) {
	s.mu.RLock()
	e, ok := s.entries[token]
	s.mu.RUnlock()
	if !ok {
		return nil, nil
	}
	if time.Now().After(e.exp) {
		s.mu.Lock()
		delete(s.entries, token)
		s.mu.Unlock()
		return nil, nil
	}
	return &e.data, nil
}

func (s *MemSessionStore) cleanupLoop() {
	tick := time.NewTicker(5 * time.Minute)
	defer tick.Stop()
	for range tick.C {
		s.mu.Lock()
		now := time.Now()
		for k, v := range s.entries {
			if now.After(v.exp) {
				delete(s.entries, k)
			}
		}
		s.mu.Unlock()
	}
}

// GenerateSessionToken 生成会话 token
func GenerateSessionToken() string {
	return uuid.New().String()
}
