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

// CardManageClient 卡片管理服务客户端
type CardManageClient struct {
	apiBaseURL string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewCardManageClient 创建卡片管理服务客户端
func NewCardManageClient(apiBaseURL string, logger *zap.Logger) *CardManageClient {
	return &CardManageClient{
		apiBaseURL: apiBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// CreateCardsForUnitRequest 创建卡片请求
type CreateCardsForUnitRequest struct {
	TenantID string `json:"tenant_id"`
	UnitID   string `json:"unit_id"`
}

// CreateCardsForUnitResponse 创建卡片响应
type CreateCardsForUnitResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Stats   *struct {
		ExistingCount  int `json:"existing_count"`
		CreatedCount   int `json:"created_count"`
		UpdatedCount   int `json:"updated_count"`
		DeletedCount   int `json:"deleted_count"`
		UnchangedCount int `json:"unchanged_count"`
	} `json:"stats,omitempty"`
}

// CreateCardsForUnit 调用 wisefido-card-manage API 创建/更新卡片
func (c *CardManageClient) CreateCardsForUnit(ctx context.Context, tenantID, unitID string) error {
	url := fmt.Sprintf("%s/api/v1/cards/create", c.apiBaseURL)

	reqBody := CreateCardsForUnitRequest{
		TenantID: tenantID,
		UnitID:   unitID,
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
		c.logger.Error("Card manage API returned error",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(body)),
			zap.String("tenant_id", tenantID),
			zap.String("unit_id", unitID),
		)
		return fmt.Errorf("card manage API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response CreateCardsForUnitResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("card manage API returned error: %s", response.Message)
	}

	c.logger.Debug("Successfully created cards via API",
		zap.String("tenant_id", tenantID),
		zap.String("unit_id", unitID),
		zap.Int("created_count", response.Stats.CreatedCount),
		zap.Int("updated_count", response.Stats.UpdatedCount),
	)

	return nil
}

