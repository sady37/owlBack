package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

var errRoundBadRequest = errors.New("tenant_id and executor_id are required")

// RoundsService 巡房记录服务（Automatic Rounds 完成时落库、审计列表与导出）
type RoundsService struct {
	repo   repository.RoundsRepository
	logger *zap.Logger
}

// NewRoundsService 创建巡房记录服务
func NewRoundsService(repo repository.RoundsRepository, logger *zap.Logger) *RoundsService {
	return &RoundsService{repo: repo, logger: logger}
}

// CreateRoundRequest 创建一轮请求（Overview Complete Rounds 调用）
type CreateRoundRequest struct {
	StartedAt    *time.Time           `json:"started_at"`    // 可选，本轮开始时间
	ItemsChecked []domain.RoundItemChecked `json:"items_checked"` // 勾选项快照
}

// CreateRound 创建一轮巡房记录（round_type=manual）
func (s *RoundsService) CreateRound(ctx context.Context, tenantID, executorID string, req *CreateRoundRequest) (string, error) {
	if tenantID == "" || executorID == "" {
		return "", errRoundBadRequest
	}
	round := &domain.Round{
		TenantID:   tenantID,
		RoundType:  "manual",
		ExecutorID: executorID,
		RoundTime:  time.Now(),
		Status:     "completed",
	}
	if req != nil {
		round.StartedAt = req.StartedAt
		if len(req.ItemsChecked) > 0 {
			data, err := json.Marshal(req.ItemsChecked)
			if err != nil {
				return "", err
			}
			round.ItemsChecked = data
		}
	}
	return s.repo.CreateRound(ctx, tenantID, round)
}

// ListRounds 分页列表（审计用）
func (s *RoundsService) ListRounds(ctx context.Context, tenantID string, filters *repository.RoundFilters, page, size int) ([]*domain.Round, int, error) {
	if tenantID == "" {
		return nil, 0, nil
	}
	return s.repo.ListRounds(ctx, tenantID, filters, page, size)
}
