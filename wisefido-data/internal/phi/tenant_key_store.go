package phi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
)

// TenantKeyStore 管理 tenant_key 缓存，按需从 K 服务获取。
type TenantKeyStore struct {
	mu         sync.RWMutex
	keys       map[string][]byte
	socketPath string
	masterPin  string
}

func NewTenantKeyStore(socketPath, masterPin string) *TenantKeyStore {
	return &TenantKeyStore{keys: make(map[string][]byte), socketPath: socketPath, masterPin: masterPin}
}

func (s *TenantKeyStore) GetKey(tenantID string) ([]byte, error) {
	s.mu.RLock()
	if k, ok := s.keys[tenantID]; ok {
		s.mu.RUnlock()
		return k, nil
	}
	s.mu.RUnlock()
	key, err := s.fetchFromK(tenantID)
	if err != nil { return nil, err }
	s.mu.Lock()
	s.keys[tenantID] = key
	s.mu.Unlock()
	return key, nil
}

func (s *TenantKeyStore) fetchFromK(tenantID string) ([]byte, error) {
	body, _ := json.Marshal(map[string]string{"tenant_id": tenantID, "master_pin": s.masterPin})
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", s.socketPath)
			},
		},
	}
	resp, err := client.Post("http://kms/tenant-key", "application/json", strings.NewReader(string(body)))
	if err != nil { return nil, fmt.Errorf("connect to K (%s): %w", s.socketPath, err) }
	defer resp.Body.Close()
	if resp.StatusCode != 200 { return nil, fmt.Errorf("K returned %d for tenant %s", resp.StatusCode, tenantID) }
	var result struct{ TenantKey string `json:"tenant_key"` }
	json.NewDecoder(resp.Body).Decode(&result)
	key, err := base64.StdEncoding.DecodeString(result.TenantKey)
	if err != nil { return nil, err }
	if len(key) != 32 { return nil, fmt.Errorf("tenant_key wrong size: %d", len(key)) }
	return key, nil
}
