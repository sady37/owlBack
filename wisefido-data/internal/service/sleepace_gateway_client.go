package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// IANAToOffsetSeconds 将 IANA 时区（如 America/Denver）转为当前时刻的 UTC 偏移秒数。
func IANAToOffsetSeconds(iana string) int {
	if iana == "" {
		return DefaultTimezoneOffsetSeconds
	}
	loc, err := time.LoadLocation(iana)
	if err != nil {
		return DefaultTimezoneOffsetSeconds
	}
	_, offset := time.Now().In(loc).Zone()
	return offset
}

// SleepaceGatewayClient proxies sleepad cloud API requests through wisefido-sleepace.
// The proxy injects the auth token automatically; callers only provide the data payload.
type SleepaceGatewayClient struct {
	apiBaseURL string
	httpClient *http.Client
	logger     *zap.Logger
}

func NewSleepaceGatewayClient(apiBaseURL string, logger *zap.Logger) *SleepaceGatewayClient {
	return &SleepaceGatewayClient{
		apiBaseURL: apiBaseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		logger:     logger,
	}
}

// InitializeDevice 调用 Sleepace 绑定。仅需两参数，与 Sleepace API 一致：
//   - Sleepace device_id = device_code (wisefido device_store.device_code，厂家导入)
//   - Sleepace userid    = deviceID (wisefido device_store.device_id UUID)
//
// timezoneSeconds 可选。
func (c *SleepaceGatewayClient) InitializeDevice(ctx context.Context, deviceCode, deviceID string, timezoneSeconds *int) (outCode string, err error) {
	url := c.apiBaseURL + "/api/v1/sleepace/device/initialize"
	payload := map[string]interface{}{"device_code": deviceCode, "user_id": deviceID}
	if timezoneSeconds != nil {
		payload["timezone"] = *timezoneSeconds
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	c.logger.Debug("sleepace initialize call", zap.String("url", url), zap.String("device_code", deviceCode), zap.String("user_id", deviceID))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Warn("sleepace initialize request failed", zap.String("device_code", deviceCode), zap.String("url", url), zap.Error(err))
		return "", fmt.Errorf("request to sleepace initialize: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)
	if resp.StatusCode != http.StatusOK {
		if bodyStr == "" {
			bodyStr = "(empty body)"
		}
		errMsg := fmt.Sprintf("sleepace initialize status %d: %s", resp.StatusCode, bodyStr)
		if strings.Contains(strings.ToLower(bodyStr), "device not found") {
			errMsg += " (device must be online and connected via MQTT at least once before bind)"
		}
		c.logger.Warn("sleepace initialize non-200", zap.String("device_code", deviceCode), zap.Int("status", resp.StatusCode), zap.String("body", bodyStr))
		return "", fmt.Errorf("%s", errMsg)
	}
	var out struct {
		Success    bool   `json:"success"`
		DeviceCode string `json:"device_code"`
		Error      string `json:"error"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}
	if !out.Success {
		msg := out.Error
		if msg == "" {
			msg = bodyStr
		}
		if msg == "" {
			msg = "unknown"
		}
		if strings.Contains(strings.ToLower(msg), "device not found") {
			msg = "device not found in Sleepace (device must be online and connected via MQTT at least once before bind)"
		}
		c.logger.Warn("sleepace initialize success=false", zap.String("device_code", deviceCode), zap.String("error", msg))
		return "", fmt.Errorf("sleepace initialize failed: %s", msg)
	}
	return out.DeviceCode, nil
}

// UnbindDevice 解绑：调用 wisefido-sleepace DELETE /api/v1/sleepace/device/{device_code}。
func (c *SleepaceGatewayClient) UnbindDevice(ctx context.Context, deviceCode string) error {
	url := c.apiBaseURL + "/api/v1/sleepace/device/" + deviceCode
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request unbind: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unbind status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// 默认时区：Denver (America/Denver)。用于 bindInfo 无返回时区或 bind 请求时未配置时。
const (
	DefaultTimezoneIANA          = "America/Denver"
	DefaultTimezoneOffsetSeconds = -25200 // UTC-7 (Mountain Standard)
)

// BindInfoItem 用户绑定信息单条（bindInfo 返回的 data 数组元素）。时区：可选，未返回时使用默认时区 Denver。
type BindInfoItem struct {
	DeviceID      string  `json:"deviceId"`
	DeviceName    string  `json:"deviceName"`
	DeviceType    int     `json:"deviceType"`
	LeftRight     int     `json:"leftRight"`
	DeviceVersion float64 `json:"deviceVersion"`
	Timezone      *int    `json:"timezone,omitempty"` // 用户所在时区，单位秒；空则用 DefaultTimezoneOffsetSeconds (Denver)
}

// GetBindInfo 查询用户绑定信息（Sleepace bindInfo），userId 为 device_id (UUID)。
func (c *SleepaceGatewayClient) GetBindInfo(ctx context.Context, userID string) (status int, msg string, items []BindInfoItem, err error) {
	url := c.apiBaseURL + "/api/v1/proxy/sleepace/bindInfo"
	payload := map[string]string{"userId": userID}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return 0, "", nil, fmt.Errorf("marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return 0, "", nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, "", nil, fmt.Errorf("request bindInfo: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var out struct {
		Status int            `json:"status"`
		Msg    string         `json:"msg"`
		Data   []BindInfoItem `json:"data"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return 0, "", nil, fmt.Errorf("unmarshal bindInfo response: %w", err)
	}
	return out.Status, out.Msg, out.Data, nil
}

// Ping 检查 wisefido-sleepace 网关是否可达（GET /health），启动时调用以便日志显示连接结果。
func (c *SleepaceGatewayClient) Ping(ctx context.Context) error {
	url := c.apiBaseURL + "/health"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("ping request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ping %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ping %s status %d: %s", url, resp.StatusCode, string(body))
	}
	return nil
}

// UpdateAlarmConfig pushes alarm config to sleepad hardware via wisefido-sleepace proxy.
// sleepaceData is the data payload (userId, deviceId, flags/params) for the cloud API.
func (c *SleepaceGatewayClient) UpdateAlarmConfig(ctx context.Context, sleepaceData map[string]interface{}) error {
	proxyURL := fmt.Sprintf("%s/api/v1/proxy/sleepace/updatealarmnotifyconfig", c.apiBaseURL)

	jsonData, err := json.Marshal(sleepaceData)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, proxyURL, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request to sleepad gateway proxy: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sleepad proxy status %d: %s", resp.StatusCode, string(body))
	}

	var out struct {
		Status int    `json:"status"`
		Msg    string `json:"msg"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return fmt.Errorf("unmarshal sleepad response: %w", err)
	}
	if out.Status != 0 {
		return fmt.Errorf("sleepad API error: %s (status: %d)", out.Msg, out.Status)
	}

	return nil
}

// SetRealtimeInterval pushes realtime data interval via wisefido-sleepace proxy.
// userId = wisefido devices.device_id (UUID), deviceId = device_store.device_code.
func (c *SleepaceGatewayClient) SetRealtimeInterval(ctx context.Context, deviceID, deviceCode string, interval int) error {
	return c.postProxy(ctx, "/sleepace/device/updateconfig", map[string]interface{}{
		"userId": deviceID, "deviceId": deviceCode, "leftRight": 0, "interval": interval,
	})
}

// SetLeaveSensibility pushes leave-bed sensitivity via wisefido-sleepace proxy.
func (c *SleepaceGatewayClient) SetLeaveSensibility(ctx context.Context, deviceID, deviceCode string, mode int) error {
	return c.postProxy(ctx, "/sleepace/device/updateAlgMode", map[string]interface{}{
		"userId": deviceID, "deviceId": deviceCode, "leftRight": 0, "mode": mode,
	})
}

// SetBedParameters pushes mattress thickness/material via wisefido-sleepace proxy.
func (c *SleepaceGatewayClient) SetBedParameters(ctx context.Context, deviceID, deviceCode string, thickness, material int) error {
	return c.postProxy(ctx, "/sleepace/updateSetting", map[string]interface{}{
		"userId": deviceID, "deviceId": deviceCode, "leftRight": 0, "thickness": thickness, "material": material,
	})
}

// SetReportUploadType pushes report upload type via wisefido-sleepace proxy.
func (c *SleepaceGatewayClient) SetReportUploadType(ctx context.Context, deviceID, deviceCode string, uploadType int) error {
	return c.postProxy(ctx, "/sleepace/reportUploadType/set", map[string]interface{}{
		"userId": deviceID, "deviceId": deviceCode, "reportUploadType": uploadType,
	})
}

// SetReportUploadTime pushes report upload time via wisefido-sleepace proxy.
func (c *SleepaceGatewayClient) SetReportUploadTime(ctx context.Context, deviceID, deviceCode string, uploadTime int) error {
	return c.postProxy(ctx, "/sleepace/setReportUploadTime", map[string]interface{}{
		"userId": deviceID, "deviceId": deviceCode, "leftRight": 0, "reportUploadTime": uploadTime,
	})
}

func (c *SleepaceGatewayClient) postProxy(ctx context.Context, sleepacePath string, data map[string]interface{}) error {
	proxyURL := fmt.Sprintf("%s/api/v1/proxy%s", c.apiBaseURL, sleepacePath)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, proxyURL, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request to sleepad gateway proxy: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sleepad proxy status %d: %s", resp.StatusCode, string(body))
	}

	var out struct {
		Status int    `json:"status"`
		Msg    string `json:"msg"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return fmt.Errorf("unmarshal sleepad response: %w", err)
	}
	if out.Status != 0 {
		return fmt.Errorf("sleepad API error: %s (status: %d)", out.Msg, out.Status)
	}
	return nil
}

// GetDeviceConfig 查询设备配置（语雀 getconfig），返回 realtimeDataInterval(秒)。用于加载时回填设备当前值。
func (c *SleepaceGatewayClient) GetDeviceConfig(ctx context.Context, deviceID, deviceCode string) (realtimeDataInterval int, err error) {
	proxyURL := fmt.Sprintf("%s/api/v1/proxy/sleepace/device/getconfig", c.apiBaseURL)
	payload := map[string]interface{}{"userId": deviceID, "deviceId": deviceCode, "leftRight": 0}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("marshal getconfig: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, proxyURL, bytes.NewReader(jsonData))
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("getconfig request: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("getconfig status %d: %s", resp.StatusCode, string(body))
	}
	var out struct {
		Status int    `json:"status"`
		Msg    string `json:"msg"`
		Data   struct {
			RealtimeDataInterval int `json:"realtimeDataInterval"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return 0, fmt.Errorf("unmarshal getconfig: %w", err)
	}
	if out.Status != 0 {
		return 0, fmt.Errorf("getconfig API: %s (status %d)", out.Msg, out.Status)
	}
	if out.Data.RealtimeDataInterval > 0 {
		return out.Data.RealtimeDataInterval, nil
	}
	return 0, nil
}

// GetAlarmConfig queries current alarm config from sleepad hardware via proxy.
// userId = wisefido devices.device_id, deviceId = device_store.device_code.
func (c *SleepaceGatewayClient) GetAlarmConfig(ctx context.Context, deviceID, deviceCode string) (map[string]interface{}, error) {
	proxyURL := fmt.Sprintf("%s/api/v1/proxy/sleepace/getalarmnotifyconfig", c.apiBaseURL)

	payload := map[string]interface{}{"userId": deviceID, "deviceId": deviceCode}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, proxyURL, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to sleepad gateway proxy: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sleepad proxy status %d: %s", resp.StatusCode, string(body))
	}

	var out struct {
		Status int                    `json:"status"`
		Msg    string                 `json:"msg"`
		Data   map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("unmarshal sleepad response: %w", err)
	}
	if out.Status != 0 {
		return nil, fmt.Errorf("sleepad API error: %s (status: %d)", out.Msg, out.Status)
	}

	return out.Data, nil
}
