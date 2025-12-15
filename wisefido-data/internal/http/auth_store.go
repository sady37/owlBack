package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type AuthUser struct {
	UserID              string
	TenantID            string
	UserAccount         string
	Role                string
	AccountHash         string
	AccountPasswordHash string
}

// AuthStore is a minimal in-memory auth DB for dev/stub mode.
// It matches owlFront hashing rules (see owlFront/src/utils/crypto.ts):
// - accountHash = sha256(lower(account))
// - accountPasswordHash = sha256(lower(account) + ":" + password)
type AuthStore struct {
	mu sync.RWMutex
	// tenantID -> accountHash -> user
	byTenant map[string]map[string]AuthUser
}

func NewAuthStore() *AuthStore {
	return &AuthStore{
		byTenant: map[string]map[string]AuthUser{},
	}
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func normalizeAccount(account string) string {
	return strings.TrimSpace(strings.ToLower(account))
}

func HashAccount(account string) string {
	return sha256Hex(normalizeAccount(account))
}

func HashAccountPassword(account, password string) string {
	return sha256Hex(normalizeAccount(account) + ":" + password)
}

// HashPassword hashes password only (independent of account/phone/email)
// Used for contact password_hash which should only depend on password itself
func HashPassword(password string) string {
	return sha256Hex(password)
}

func (s *AuthStore) UpsertUser(tenantID, userAccount, role, password string) AuthUser {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.byTenant[tenantID] == nil {
		s.byTenant[tenantID] = map[string]AuthUser{}
	}
	ah := HashAccount(userAccount)
	aph := HashAccountPassword(userAccount, password)
	u := AuthUser{
		UserID:              uuid.NewString(),
		TenantID:            tenantID,
		UserAccount:         userAccount,
		Role:                role,
		AccountHash:         ah,
		AccountPasswordHash: aph,
	}
	s.byTenant[tenantID][ah] = u
	return u
}

func (s *AuthStore) FindUser(tenantID, accountHash, accountPasswordHash string) (AuthUser, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m := s.byTenant[tenantID]
	if m == nil {
		return AuthUser{}, false
	}
	u, ok := m[accountHash]
	if !ok {
		return AuthUser{}, false
	}
	if accountPasswordHash != "" && u.AccountPasswordHash != accountPasswordHash {
		return AuthUser{}, false
	}
	return u, true
}

func (s *AuthStore) TenantsForLogin(accountHash, accountPasswordHash string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []string{}
	for tenantID, m := range s.byTenant {
		u, ok := m[accountHash]
		if !ok {
			continue
		}
		if accountPasswordHash != "" && u.AccountPasswordHash != accountPasswordHash {
			continue
		}
		out = append(out, tenantID)
	}
	return out
}

// ListUsersByTenant returns a snapshot of users for a tenant (dev/stub helpers).
func (s *AuthStore) ListUsersByTenant(tenantID string) []AuthUser {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m := s.byTenant[tenantID]
	if m == nil {
		return []AuthUser{}
	}
	out := make([]AuthUser, 0, len(m))
	for _, u := range m {
		out = append(out, u)
	}
	return out
}
