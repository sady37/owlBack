package notify

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/net/http2"
)

const (
	apnsHostProd    = "https://api.push.apple.com"
	apnsHostSandbox = "https://api.development.push.apple.com"
	tokenTTL        = 50 * time.Minute // APNs JWT 最长有效 60 分钟，提前 10 分钟刷新
)

// ── Payload 结构（与 iOS Models.swift APNSPayload 对应）────────────────

type APNSPayload struct {
	APS        APSDict `json:"aps"`
	CardID     string  `json:"card_id,omitempty"`
	EventType  string  `json:"event_type,omitempty"`
	AlarmLevel int     `json:"alarm_level,omitempty"`
}

type APSDict struct {
	Alert             APSAlert `json:"alert"`
	Sound             string   `json:"sound,omitempty"`
	Badge             *int     `json:"badge,omitempty"` // 告警：与 cards.pop_alarm_* 待 pop 卡数对齐
	ContentAvailable  int      `json:"content-available,omitempty"`
	InterruptionLevel string   `json:"interruption-level,omitempty"` // "critical" | "time-sensitive"
}

type APSAlert struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// ErrDeviceTokenInvalid APNs 返回 410：token 已失效（用户卸载 App）
var ErrDeviceTokenInvalid = fmt.Errorf("apns: device token invalid (410 Gone)")

// ── APNSSender ────────────────────────────────────────────────────────

type APNSSender struct {
	keyID    string
	teamID   string
	bundleID string
	privKey  *ecdsa.PrivateKey
	client   *http.Client // HTTP/2，APNs 强制要求

	mu        sync.Mutex
	jwtToken  string
	jwtExpiry time.Time
}

// NewAPNSSender 从 base64 编码的 .p8 私钥创建发送器
// p8Base64：去掉 -----BEGIN/END PRIVATE KEY----- 行后的纯 base64 内容
func NewAPNSSender(keyID, teamID, bundleID, p8Base64 string) (*APNSSender, error) {
	// 兼容带 padding 和不带 padding 的 base64
	raw, err := base64.StdEncoding.DecodeString(p8Base64)
	if err != nil {
		raw, err = base64.RawStdEncoding.DecodeString(p8Base64)
		if err != nil {
			return nil, fmt.Errorf("apns: decode p8 base64: %w", err)
		}
	}

	// 如果传入的是完整 PEM（含 header/footer），直接用；否则包装一下
	if !bytes.HasPrefix(raw, []byte("-----")) {
		raw = append([]byte("-----BEGIN PRIVATE KEY-----\n"), raw...)
		raw = append(raw, []byte("\n-----END PRIVATE KEY-----")...)
	}

	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, fmt.Errorf("apns: invalid PEM block")
	}
	pk, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("apns: parse PKCS8 key: %w", err)
	}
	ecKey, ok := pk.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("apns: key is not ECDSA (got %T)", pk)
	}

	transport := &http2.Transport{}
	return &APNSSender{
		keyID:    keyID,
		teamID:   teamID,
		bundleID: bundleID,
		privKey:  ecKey,
		client:   &http.Client{Transport: transport, Timeout: 10 * time.Second},
	}, nil
}

// Send 向单个设备发送推送
// env: "production" | "sandbox"
func (s *APNSSender) Send(ctx context.Context, deviceToken, env string, payload APNSPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("apns: marshal payload: %w", err)
	}

	host := apnsHostProd
	if env == "sandbox" {
		host = apnsHostSandbox
	}
	url := fmt.Sprintf("%s/3/device/%s", host, deviceToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("apns: build request: %w", err)
	}

	token, err := s.bearerToken()
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "bearer "+token)
	req.Header.Set("apns-topic", s.bundleID)
	req.Header.Set("apns-push-type", "alert")
	req.Header.Set("Content-Type", "application/json")
	if payload.APS.InterruptionLevel == "critical" {
		req.Header.Set("apns-priority", "10") // 立即送达
	} else {
		req.Header.Set("apns-priority", "5")
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("apns: http request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusGone: // 410：token 已失效
		return ErrDeviceTokenInvalid
	default:
		var errBody struct {
			Reason string `json:"reason"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		return fmt.Errorf("apns: status %d: %s", resp.StatusCode, errBody.Reason)
	}
}

// bearerToken 返回有效的 APNs JWT，过期前自动刷新（线程安全）
func (s *APNSSender) bearerToken() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if time.Now().Before(s.jwtExpiry) {
		return s.jwtToken, nil
	}
	claims := jwt.MapClaims{
		"iss": s.teamID,
		"iat": time.Now().Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	t.Header["kid"] = s.keyID
	signed, err := t.SignedString(s.privKey)
	if err != nil {
		return "", fmt.Errorf("apns: sign JWT: %w", err)
	}
	s.jwtToken = signed
	s.jwtExpiry = time.Now().Add(tokenTTL)
	return signed, nil
}