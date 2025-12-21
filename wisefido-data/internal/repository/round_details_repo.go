package repository

import (
	"context"
	"time"

	"wisefido-data/internal/domain"
)

// RoundDetailFilters 巡房详细记录查询过滤器
type RoundDetailFilters struct {
	BedID      string     // 床位ID
	UnitID     string     // 单元ID
	BedStatus  string     // 在床状态：'in_bed'/'out_of_bed'/'unknown'
	StartTime  *time.Time // 开始时间
	EndTime    *time.Time // 结束时间
}

// RoundDetailsRepository 巡房详细记录Repository接口
type RoundDetailsRepository interface {
	// GetRoundDetail 获取巡房详细记录
	GetRoundDetail(ctx context.Context, tenantID, detailID string) (*domain.RoundDetail, error)

	// GetRoundDetailsByRound 获取某个巡房记录的所有详细记录
	GetRoundDetailsByRound(ctx context.Context, tenantID, roundID string) ([]*domain.RoundDetail, error)

	// GetRoundDetailsByResident 获取某个住户的所有巡房详细记录（支持分页）
	GetRoundDetailsByResident(ctx context.Context, tenantID, residentID string, filters *RoundDetailFilters, page, size int) ([]*domain.RoundDetail, int, error)

	// UpsertRoundDetail 创建或更新巡房详细记录
	// 注意：UNIQUE(round_id, resident_id)，使用UPSERT语义
	UpsertRoundDetail(ctx context.Context, tenantID, roundID string, roundDetail *domain.RoundDetail) (string, error)

	// UpdateRoundDetail 更新巡房详细记录
	UpdateRoundDetail(ctx context.Context, tenantID, detailID string, roundDetail *domain.RoundDetail) error

	// DeleteRoundDetail 删除巡房详细记录
	DeleteRoundDetail(ctx context.Context, tenantID, detailID string) error
}

