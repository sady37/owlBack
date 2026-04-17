package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"

	"golang.org/x/crypto/hkdf"
)

// DeriveKey derives a 32-byte AES key using HKDF-SHA256.
func DeriveKey(masterKey, salt []byte, info []byte) ([]byte, error) {
	if len(masterKey) == 0 {
		return nil, errors.New("masterKey is empty")
	}
	hk := hkdf.New(sha256.New, masterKey, salt, info)
	key := make([]byte, 32)
	if _, err := io.ReadFull(hk, key); err != nil {
		return nil, err
	}
	return key, nil
}

// EncryptWithDataKey uses a raw AES-256 key to AEAD-encrypt plaintext with tenantAAD as AAD.
// Returned payload format: nonce(12) || ciphertext
func EncryptWithDataKey(key, tenantAAD, plaintext []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("data key must be 32 bytes for AES-256")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	ct := aesgcm.Seal(nil, nonce, plaintext, tenantAAD)
	out := make([]byte, 0, len(nonce)+len(ct))
	out = append(out, nonce...)
	out = append(out, ct...)
	return out, nil
}

// DecryptWithDataKey decrypts payload produced by EncryptWithDataKey.
func DecryptWithDataKey(key, tenantAAD, payload []byte) ([]byte, error) {
	if len(payload) < 12 {
		return nil, errors.New("payload too short")
	}
	if len(key) != 32 {
		return nil, errors.New("data key must be 32 bytes for AES-256")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := payload[:12]
	ct := payload[12:]
	pt, err := aesgcm.Open(nil, nonce, ct, tenantAAD)
	if err != nil {
		return nil, err
	}
	return pt, nil
}
