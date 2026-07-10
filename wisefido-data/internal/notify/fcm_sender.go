package notify

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	fcmScope      = "https://www.googleapis.com/auth/firebase.messaging"
	fcmSendURLFmt = "https://fcm.googleapis.com/v1/projects/%s/messages:send"
	accessTokenTTL = 55 * time.Minute // Google access token 有效 1h，提前 5 分钟刷新
)

// ── Payload 结构（FCM HTTP v1 信封，Google schema，不可套用 aps）─────────

type FCMMessage struct {
	Token        string            `json:"token"`
	Notification FCMNotification   `json:"notification"`
	Android      FCMAndroidConfig  `json:"android"`
	Data         map[string]string `json:"data,omitempty"` // 值必须全 string
}

type FCMNotification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type FCMAndroidConfig struct {
	Priority     string                 `json:"priority"` // "high" | "normal"
	Notification FCMAndroidNotification `json:"notification"`
}

type FCMAndroidNotification struct {
	Sound                string `json:"sound,omitempty"`
	ChannelID            string `json:"channel_id,omitempty"`
	NotificationPriority string `json:"notification_priority,omitempty"` // PRIORITY_MAX 等
	NotificationCount    *int   `json:"notification_count,omitempty"`    // 角标 hint
}

// ErrFCMTokenInvalid FCM 返回 UNREGISTERED / INVALID_ARGUMENT：token 已失效或格式错，回收。
var ErrFCMTokenInvalid = fmt.Errorf("fcm: device token invalid")

// ── FCMSender ─────────────────────────────────────────────────────────

type FCMSender struct {
	projectID   string
	clientEmail string
	tokenURI    string
	privKey     *rsa.PrivateKey
	client      *http.Client

	mu          sync.Mutex
	accessToken string
	tokenExpiry time.Time
}

type serviceAccount struct {
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
	TokenURI    string `json:"token_uri"`
}

// NewFCMSender 从 service account JSON 全文创建发送器。
func NewFCMSender(projectID, serviceAccountJSON string) (*FCMSender, error) {
	if strings.TrimSpace(projectID) == "" {
		return nil, fmt.Errorf("fcm: empty project id")
	}
	var sa serviceAccount
	if err := json.Unmarshal([]byte(serviceAccountJSON), &sa); err != nil {
		return nil, fmt.Errorf("fcm: parse service account json: %w", err)
	}
	if sa.ClientEmail == "" || sa.PrivateKey == "" {
		return nil, fmt.Errorf("fcm: service account missing client_email or private_key")
	}
	block, _ := pem.Decode([]byte(sa.PrivateKey))
	if block == nil {
		return nil, fmt.Errorf("fcm: invalid private_key PEM")
	}
	pk, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("fcm: parse PKCS8 key: %w", err)
	}
	rsaKey, ok := pk.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("fcm: key is not RSA (got %T)", pk)
	}
	tokenURI := sa.TokenURI
	if tokenURI == "" {
		tokenURI = "https://oauth2.googleapis.com/token"
	}

	return &FCMSender{
		projectID:   projectID,
		clientEmail: sa.ClientEmail,
		tokenURI:    tokenURI,
		privKey:     rsaKey,
		client:      &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Send 向单个设备发送 FCM 通知。
func (s *FCMSender) Send(ctx context.Context, deviceToken string, msg FCMMessage) error {
	msg.Token = deviceToken
	body, err := json.Marshal(map[string]FCMMessage{"message": msg})
	if err != nil {
		return fmt.Errorf("fcm: marshal message: %w", err)
	}

	token, err := s.bearerToken(ctx)
	if err != nil {
		return err
	}

	sendURL := fmt.Sprintf(fcmSendURLFmt, s.projectID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sendURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("fcm: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("fcm: http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	var errBody struct {
		Error struct {
			Status  string `json:"status"`
			Message string `json:"message"`
			Details []struct {
				ErrorCode string `json:"errorCode"`
			} `json:"details"`
		} `json:"error"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&errBody)

	for _, d := range errBody.Error.Details {
		if d.ErrorCode == "UNREGISTERED" || d.ErrorCode == "INVALID_ARGUMENT" {
			return ErrFCMTokenInvalid
		}
	}
	if errBody.Error.Status == "NOT_FOUND" {
		return ErrFCMTokenInvalid
	}
	return fmt.Errorf("fcm: status %d: %s %s", resp.StatusCode, errBody.Error.Status, errBody.Error.Message)
}

// bearerToken 返回有效的 Google OAuth2 access token，过期前自动刷新（线程安全）。
// 走 JWT-bearer flow：SA-JWT(RS256) → token_uri 换 access_token。
func (s *FCMSender) bearerToken(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if time.Now().Before(s.tokenExpiry) {
		return s.accessToken, nil
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iss":   s.clientEmail,
		"scope": fcmScope,
		"aud":   s.tokenURI,
		"iat":   now.Unix(),
		"exp":   now.Add(time.Hour).Unix(),
	}
	assertion, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(s.privKey)
	if err != nil {
		return "", fmt.Errorf("fcm: sign assertion: %w", err)
	}

	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	form.Set("assertion", assertion)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.tokenURI, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("fcm: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fcm: token request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fcm: token endpoint status %d", resp.StatusCode)
	}

	var tok struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return "", fmt.Errorf("fcm: decode token: %w", err)
	}
	if tok.AccessToken == "" {
		return "", fmt.Errorf("fcm: empty access token")
	}

	ttl := accessTokenTTL
	if tok.ExpiresIn > 0 {
		if d := time.Duration(tok.ExpiresIn-60) * time.Second; d < ttl {
			ttl = d
		}
	}
	s.accessToken = tok.AccessToken
	s.tokenExpiry = time.Now().Add(ttl)
	return tok.AccessToken, nil
}
