package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"

	"golang.org/x/crypto/argon2"
)

// WrapMasterKey encrypts masterKey with passphrase using Argon2id + AES-256-GCM.
// Returns JSON blob.
func WrapMasterKey(masterKey []byte, passphrase string, keyVersion string) ([]byte, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	derived := argon2.IDKey([]byte(passphrase), salt, 1, 64*1024, 2, 32)
	block, err := aes.NewCipher(derived)
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
	ct := aesgcm.Seal(nil, nonce, masterKey, nil)
	env := wrappedFormat{
		Version:   "OWLMK1",
		ArgonSalt: base64.StdEncoding.EncodeToString(salt),
		Nonce:     base64.StdEncoding.EncodeToString(nonce),
		Cipher:    base64.StdEncoding.EncodeToString(ct),
		KeyVer:    keyVersion,
	}
	return json.Marshal(env)
}

// UnwrapMasterKey decrypts the output of WrapMasterKey.
func UnwrapMasterKey(wrapped []byte, passphrase string) ([]byte, string, error) {
	var env wrappedFormat
	if err := json.Unmarshal(wrapped, &env); err != nil {
		return nil, "", err
	}
	if env.Version != "OWLMK1" {
		return nil, "", errors.New("unsupported version: " + env.Version)
	}
	salt, err := base64.StdEncoding.DecodeString(env.ArgonSalt)
	if err != nil {
		return nil, "", err
	}
	nonce, err := base64.StdEncoding.DecodeString(env.Nonce)
	if err != nil {
		return nil, "", err
	}
	ct, err := base64.StdEncoding.DecodeString(env.Cipher)
	if err != nil {
		return nil, "", err
	}
	derived := argon2.IDKey([]byte(passphrase), salt, 1, 64*1024, 2, 32)
	block, err := aes.NewCipher(derived)
	if err != nil {
		return nil, "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, "", err
	}
	pt, err := aesgcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, "", err
	}
	return pt, env.KeyVer, nil
}

type wrappedFormat struct {
	Version   string `json:"version"`
	ArgonSalt string `json:"argon_salt"`
	Nonce     string `json:"nonce"`
	Cipher    string `json:"cipher"`
	KeyVer    string `json:"key_version,omitempty"`
}
