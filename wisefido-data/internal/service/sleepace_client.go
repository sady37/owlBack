package service

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// SleepaceToken Sleepace 认证 Token
type SleepaceToken struct {
	AppId     string `json:"appId"`
	SecureKey string `json:"secureKey"`
}

// SleepaceRequest Sleepace API 请求
type SleepaceRequest struct {
	Token *SleepaceToken `json:"token"`
	Data  map[string]any `json:"data"`
}

// SleepaceResponse Sleepace API 响应
type SleepaceResponse struct {
	Status int             `json:"status"`
	Msg    string          `json:"msg"`
	Data   json.RawMessage `json:"data"`
}

// SleepaceClient Sleepace 厂家 API 客户端
type SleepaceClient struct {
	httpClient *resty.Client
	token      *SleepaceToken
	logger     *zap.Logger
	mu         sync.RWMutex
}

// NewSleepaceClient 创建 Sleepace 客户端
func NewSleepaceClient(baseURL, appID, secretKey string, logger *zap.Logger) *SleepaceClient {
	client := resty.New().
		SetBaseURL(baseURL).
		SetTimeout(30 * time.Second). // 报告下载可能需要较长时间
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(5 * time.Second).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json")

	token := &SleepaceToken{
		AppId:     appID,
		SecureKey: secretKey,
	}

	return &SleepaceClient{
		httpClient: client,
		token:      token,
		logger:     logger,
	}
}

// Get24HourDailyWithMaxReport 获取 24 小时每日报告（最大报告）
// 参考：wisefido-backend/wisefido-sleepace/modules/sleepace_service.go::DownloadReport
func (c *SleepaceClient) Get24HourDailyWithMaxReport(deviceID, deviceCode string, startTime, endTime int64) ([]json.RawMessage, error) {
	request := SleepaceRequest{
		Token: c.token,
		Data: map[string]any{
			"userId":    deviceID,
			"startTime": startTime,
			"endTime":   endTime,
		},
	}

	c.logger.Info("Calling Sleepace API: get24HourDailyWithMaxReport",
		zap.String("device_id", deviceID),
		zap.String("device_code", deviceCode),
		zap.Int64("start_time", startTime),
		zap.Int64("end_time", endTime),
	)

	var response SleepaceResponse
	resp, err := c.httpClient.R().
		SetBody(request).
		SetResult(&response).
		Post("/sleepace/get24HourDailyWithMaxReport")

	if err != nil {
		c.logger.Error("Sleepace API call failed",
			zap.Error(err),
			zap.Int("status_code", resp.StatusCode()),
		)
		return nil, fmt.Errorf("failed to call Sleepace API: %w", err)
	}

	if response.Status != 0 {
		c.logger.Error("Sleepace API returned error",
			zap.Int("status", response.Status),
			zap.String("msg", response.Msg),
		)
		return nil, fmt.Errorf("Sleepace API error: %s (status: %d)", response.Msg, response.Status)
	}

	// 解析报告数据
	var reports []json.RawMessage
	if err := json.Unmarshal(response.Data, &reports); err != nil {
		c.logger.Error("Failed to unmarshal Sleepace API response",
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to unmarshal reports: %w", err)
	}

	c.logger.Info("Successfully retrieved reports from Sleepace API",
		zap.Int("report_count", len(reports)),
	)

	return reports, nil
}

