package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
)

// QinglanClient 调用 wisefido-qinglan HTTP API，提供雷达设备交互功能（属性读写、实时数据订阅、设备功能调用）。
// 替代原 wisefido-radar 内部 API，统一通过 wisefido-qinglan 与设备通信。
type QinglanClient struct {
	apiBaseURL string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewQinglanClient 创建 Qinglan 客户端
func NewQinglanClient(apiBaseURL string, logger *zap.Logger) *QinglanClient {
	return &QinglanClient{
		apiBaseURL: apiBaseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		logger:     logger,
	}
}

// GetDeviceProperties 读取设备属性（GET /api/v1/radar/devices/{uid}/properties?keys=key1,key2）。
// keys 为空时读取所有属性。
func (c *QinglanClient) GetDeviceProperties(ctx context.Context, deviceUID string, keys []string) (map[string]interface{}, error) {
	urlStr := fmt.Sprintf("%s/api/v1/radar/devices/%s/properties", c.apiBaseURL, deviceUID)
	if len(keys) > 0 {
		u, err := url.Parse(urlStr)
		if err == nil {
			q := u.Query()
			q.Set("keys", strings.Join(keys, ","))
			u.RawQuery = q.Encode()
			urlStr = u.String()
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		c.logger.Warn("Qinglan GetDeviceProperties non-200",
			zap.String("uid", deviceUID),
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)),
		)
		return nil, fmt.Errorf("qinglan API status %d: %s", resp.StatusCode, string(body))
	}

	var out struct {
		Success bool                   `json:"success"`
		Data    map[string]interface{} `json:"data"`
		Error   string                 `json:"error,omitempty"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	if !out.Success && out.Error != "" {
		return nil, fmt.Errorf("qinglan API error: %s", out.Error)
	}

	return out.Data, nil
}

// SetDeviceProperties 设置设备属性（PUT /api/v1/radar/devices/{uid}/properties）。
// 用于工作模式、跌倒/呼吸心率参数、安装方式/高度/边界等所有属性。
func (c *QinglanClient) SetDeviceProperties(ctx context.Context, deviceUID string, properties map[string]interface{}) error {
	url := fmt.Sprintf("%s/api/v1/radar/devices/%s/properties", c.apiBaseURL, deviceUID)

	reqBody := map[string]interface{}{"properties": properties}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal properties: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		c.logger.Warn("Qinglan SetDeviceProperties non-200",
			zap.String("uid", deviceUID),
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)),
		)
		return fmt.Errorf("qinglan API status %d: %s", resp.StatusCode, string(body))
	}

	var out struct {
		Success bool   `json:"success"`
		Error   string `json:"error,omitempty"`
	}
	_ = json.Unmarshal(body, &out)
	if !out.Success && out.Error != "" {
		return fmt.Errorf("qinglan API error: %s", out.Error)
	}

	return nil
}

// SubscribeRealtimeData 订阅实时数据（POST /api/v1/radar/devices/{uid}/subscribe）。
// content: 0-同时订阅，1-订阅轨迹，2-订阅呼吸心率
// duration: 订阅时长（秒），最大3600
func (c *QinglanClient) SubscribeRealtimeData(ctx context.Context, deviceUID string, content interface{}, duration int) error {
	url := fmt.Sprintf("%s/api/v1/radar/devices/%s/subscribe", c.apiBaseURL, deviceUID)

	reqBody := map[string]interface{}{
		"content":  content,
		"duration": duration,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		c.logger.Warn("Qinglan SubscribeRealtimeData non-200",
			zap.String("uid", deviceUID),
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)),
		)
		return fmt.Errorf("qinglan API status %d: %s", resp.StatusCode, string(body))
	}

	var out struct {
		Success bool   `json:"success"`
		Error   string `json:"error,omitempty"`
	}
	_ = json.Unmarshal(body, &out)
	if !out.Success && out.Error != "" {
		return fmt.Errorf("qinglan API error: %s", out.Error)
	}

	return nil
}

// CallDeviceFunction 调用设备功能（POST /api/v1/radar/devices/{uid}/function）。
// dev: 0-重启雷达和主控，1-只重启雷达，2-只重启主控
func (c *QinglanClient) CallDeviceFunction(ctx context.Context, deviceUID string, dev int) error {
	url := fmt.Sprintf("%s/api/v1/radar/devices/%s/function", c.apiBaseURL, deviceUID)

	reqBody := map[string]interface{}{
		"dev": dev,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		c.logger.Warn("Qinglan CallDeviceFunction non-200",
			zap.String("uid", deviceUID),
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)),
		)
		return fmt.Errorf("qinglan API status %d: %s", resp.StatusCode, string(body))
	}

	var out struct {
		Success bool   `json:"success"`
		Error   string `json:"error,omitempty"`
	}
	_ = json.Unmarshal(body, &out)
	if !out.Success && out.Error != "" {
		return fmt.Errorf("qinglan API error: %s", out.Error)
	}

	return nil
}
