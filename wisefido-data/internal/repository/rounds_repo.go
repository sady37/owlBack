package repository

import (
	"context"
	"time"

	"wisefido-data/internal/domain"
)

// RoundFilters 巡房记录查询过滤器
type RoundFilters struct {
	ExecutorID string    // 执行人ID
	UnitID     string    // 单元ID
	RoundType  string    // 巡房类型：'location'/'manual'/'scheduled'
	Status     string    // 状态：'draft'/'completed'/'cancelled'
	StartTime  *time.Time // 开始时间
	EndTime    *time.Time // 结束时间
}

// RoundsRepository 巡房记录Repository接口
type RoundsRepository interface {
	// GetRound 获取巡房记录
	GetRound(ctx context.Context, tenantID, roundID string) (*domain.Round, error)

	// ListRounds 批量查询巡房记录（支持过滤和分页）
	ListRounds(ctx context.Context, tenantID string, filters *RoundFilters, page, size int) ([]*domain.Round, int, error)

	// CreateRound 创建巡房记录
	CreateRound(ctx context.Context, tenantID string, round *domain.Round) (string, error)

	// UpdateRound 更新巡房记录
	UpdateRound(ctx context.Context, tenantID, roundID string, round *domain.Round) error

	// DeleteRound 删除巡房记录
	DeleteRound(ctx context.Context, tenantID, roundID string) error

	// SetRoundStatus 更新巡房记录状态
	SetRoundStatus(ctx context.Context, tenantID, roundID, status string) error
}

