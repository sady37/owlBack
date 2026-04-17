package phi

import (
	"fmt"
	"strconv"
	"time"

	kmscrypto "owl-common/crypto"
)

// PHICryptor 提供 resident_phi 全字段加解密。
// 使用 tenant_key (32字节，由 K 派生) + AES-256-GCM，tenant_id 作 AAD。
type PHICryptor struct {
	keyStore *TenantKeyStore
}

func NewPHICryptor(keyStore *TenantKeyStore) *PHICryptor {
	return &PHICryptor{keyStore: keyStore}
}

func (c *PHICryptor) getKey(tenantID string) ([]byte, error) {
	return c.keyStore.GetKey(tenantID)
}

// String
func (c *PHICryptor) EncryptString(tenantID, plaintext string) ([]byte, error) {
	if plaintext == "" { return nil, nil }
	key, err := c.getKey(tenantID)
	if err != nil { return nil, err }
	return kmscrypto.EncryptWithDataKey(key, []byte(tenantID), []byte(plaintext))
}

func (c *PHICryptor) DecryptString(tenantID string, ct []byte) (string, error) {
	if len(ct) == 0 { return "", nil }
	key, err := c.getKey(tenantID)
	if err != nil { return "", err }
	pt, err := kmscrypto.DecryptWithDataKey(key, []byte(tenantID), ct)
	if err != nil { return "", err }
	return string(pt), nil
}

// Time
func (c *PHICryptor) EncryptTime(tenantID string, t *time.Time) ([]byte, error) {
	if t == nil { return nil, nil }
	return c.EncryptString(tenantID, t.Format("2006-01-02"))
}

func (c *PHICryptor) DecryptTime(tenantID string, ct []byte) (*time.Time, error) {
	s, err := c.DecryptString(tenantID, ct)
	if err != nil || s == "" { return nil, err }
	t, err := time.Parse("2006-01-02", s)
	if err != nil { return nil, fmt.Errorf("parse time: %w", err) }
	return &t, nil
}

// Float
func (c *PHICryptor) EncryptFloat(tenantID string, v *float64) ([]byte, error) {
	if v == nil { return nil, nil }
	return c.EncryptString(tenantID, strconv.FormatFloat(*v, 'f', 2, 64))
}

func (c *PHICryptor) DecryptFloat(tenantID string, ct []byte) (*float64, error) {
	s, err := c.DecryptString(tenantID, ct)
	if err != nil || s == "" { return nil, err }
	f, err := strconv.ParseFloat(s, 64)
	if err != nil { return nil, err }
	return &f, nil
}

// Int
func (c *PHICryptor) EncryptInt(tenantID string, v *int) ([]byte, error) {
	if v == nil { return nil, nil }
	return c.EncryptString(tenantID, strconv.Itoa(*v))
}

func (c *PHICryptor) DecryptInt(tenantID string, ct []byte) (*int, error) {
	s, err := c.DecryptString(tenantID, ct)
	if err != nil || s == "" { return nil, err }
	i, err := strconv.Atoi(s)
	if err != nil { return nil, err }
	return &i, nil
}

// Bool
func (c *PHICryptor) EncryptBool(tenantID string, v bool) ([]byte, error) {
	return c.EncryptString(tenantID, strconv.FormatBool(v))
}

func (c *PHICryptor) DecryptBool(tenantID string, ct []byte) (bool, error) {
	s, err := c.DecryptString(tenantID, ct)
	if err != nil { return false, err }
	if s == "" { return false, nil }
	return strconv.ParseBool(s)
}
