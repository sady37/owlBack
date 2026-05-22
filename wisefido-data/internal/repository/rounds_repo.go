package repository

import (
	"context"
	"time"

	"wisefido-data/internal/domain"
)

// RoundFilters 巡房记录查询过滤器
type RoundFilters struct {
	UserID       string // UUID (users.user_id)
	UnitID       string // INET CIDR (units.unit_id)
	BranchPrefix string // /56 CIDR — current branch scope
	RoundType    string
	Status       string // 'in_progress'/'completed'/'aborted'
	StartTime    *time.Time
	EndTime      *time.Time
}

// RoundsRepository 巡房记录 Repository 接口
type RoundsRepository interface {
	GetRound(ctx context.Context, tenantID, roundID string) (*domain.Round, error)
	ListRounds(ctx context.Context, tenantID string, filters *RoundFilters, page, size int) ([]*domain.Round, int, error)
	CreateRound(ctx context.Context, userID string, round *domain.Round) (string, error)
}
