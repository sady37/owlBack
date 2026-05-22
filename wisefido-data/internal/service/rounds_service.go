package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

var errRoundBadRequest = errors.New("user_id is required")

// RoundsService 巡房记录服务
type RoundsService struct {
	repo   repository.RoundsRepository
	logger *zap.Logger
}

func NewRoundsService(repo repository.RoundsRepository, logger *zap.Logger) *RoundsService {
	return &RoundsService{repo: repo, logger: logger}
}

// CreateRoundRequest 创建一轮请求（Overview Complete Rounds 调用）
//
// ReportSnapshot 是 FE 构造的不透明 JSONB blob（schema_version / stats / rows / shift_start /
// completed_at / tenant_tz 等），BE 不解析直接持久化。items_checked 单独拍进 blob 顶层。
type CreateRoundRequest struct {
	StartedAt      *time.Time                `json:"started_at"`
	Notes          string                    `json:"notes"`
	ItemsChecked   []domain.RoundItemChecked `json:"items_checked"`
	ReportSnapshot json.RawMessage           `json:"report_snapshot,omitempty"`
}

// CreateRound 创建一轮巡房记录（round_type=manual，status=completed）
// items_checked merge 进 report_snapshot blob 落 rounds_report_snapshot.snapshot 单列。
func (s *RoundsService) CreateRound(ctx context.Context, userID string, req *CreateRoundRequest) (string, error) {
	if userID == "" {
		return "", errRoundBadRequest
	}
	now := time.Now()
	round := &domain.Round{
		RoundType: "manual",
		Status:    "completed",
		EndedAt:   &now,
	}
	if req != nil {
		round.StartedAt = req.StartedAt
		round.Notes = req.Notes
		snap := map[string]any{}
		if len(req.ReportSnapshot) > 0 {
			if err := json.Unmarshal(req.ReportSnapshot, &snap); err != nil {
				return "", fmt.Errorf("invalid report_snapshot: %w", err)
			}
		}
		if len(req.ItemsChecked) > 0 {
			snap["items_checked"] = req.ItemsChecked
		}
		if len(snap) > 0 {
			data, err := json.Marshal(snap)
			if err != nil {
				return "", err
			}
			round.Snapshot = data
		}
	}
	return s.repo.CreateRound(ctx, userID, round)
}

func (s *RoundsService) ListRounds(ctx context.Context, tenantID string, filters *repository.RoundFilters, page, size int) ([]*domain.Round, int, error) {
	if tenantID == "" {
		return nil, 0, nil
	}
	return s.repo.ListRounds(ctx, tenantID, filters, page, size)
}

func (s *RoundsService) GetRound(ctx context.Context, tenantID, roundID string) (*domain.Round, error) {
	if tenantID == "" || roundID == "" {
		return nil, errRoundBadRequest
	}
	return s.repo.GetRound(ctx, tenantID, roundID)
}
