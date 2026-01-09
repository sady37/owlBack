package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// IoTTimeSeriesClient IoT 时序数据服务客户端
// 用于调用 wisefido-iot-timeseries 内部 API（如清除位置信息缓存）
type IoTTimeSeriesClient struct {
	apiBaseURL string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewIoTTimeSeriesClient 创建 IoT 时序数据服务客户端
func NewIoTTimeSeriesClient(apiBaseURL string, logger *zap.Logger) *IoTTimeSeriesClient {
	return &IoTTimeSeriesClient{
		apiBaseURL: apiBaseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second, // 较短的超时时间，因为这是内部 API 调用
		},
		logger: logger,
	}
}

// InvalidateLocationCacheRequest 清除位置信息缓存请求
type InvalidateLocationCacheRequest struct {
	DeviceID string `json:"device_id"`
}

// InvalidateLocationCacheResponse 清除位置信息缓存响应
type InvalidateLocationCacheResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// InvalidateLocationCache 清除设备位置信息缓存
// 调用 wisefido-iot-timeseries 内部 API: POST /internal/api/v1/iot-timeseries/cache/invalidate
func (c *IoTTimeSeriesClient) InvalidateLocationCache(ctx context.Context, deviceID string) error {
	if deviceID == "" {
		return fmt.Errorf("device_id is required")
	}

	url := fmt.Sprintf("%s/internal/api/v1/iot-timeseries/cache/invalidate", c.apiBaseURL)

	reqBody := InvalidateLocationCacheRequest{
		DeviceID: deviceID,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Warn("IoT timeseries API returned error",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(body)),
			zap.String("device_id", deviceID),
		)
		return fmt.Errorf("iot timeseries API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response InvalidateLocationCacheResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("iot timeseries API returned error: %s", response.Message)
	}

	c.logger.Debug("Successfully invalidated location cache via API",
		zap.String("device_id", deviceID),
	)

	return nil
}
