package repository

import (
	"context"
	"time"

	"wisefido-data/internal/domain"
)

// ConfigVersionFilters 配置版本查询过滤器
type ConfigVersionFilters struct {
	StartTime *time.Time // 开始时间
	EndTime   *time.Time // 结束时间
}

// ConfigVersionsRepository 配置版本Repository接口
type ConfigVersionsRepository interface {
	// GetConfigVersion 获取配置版本
	GetConfigVersion(ctx context.Context, tenantID, versionID string) (*domain.ConfigVersion, error)

	// GetConfigVersionAtTime 查询某个时间点的配置（用于回放）
	GetConfigVersionAtTime(ctx context.Context, tenantID, configType, entityID string, atTime time.Time) (*domain.ConfigVersion, error)

	// ListConfigVersions 查询某个实体的所有配置历史（支持分页、时间范围过滤）
	ListConfigVersions(ctx context.Context, tenantID, configType, entityID string, filters *ConfigVersionFilters, page, size int) ([]*domain.ConfigVersion, int, error)

	// CreateConfigVersion 创建新版本
	// 自动设置valid_from，将旧版本的valid_to设置为当前时间
	CreateConfigVersion(ctx context.Context, tenantID string, configVersion *domain.ConfigVersion) (string, error)

	// UpdateConfigVersion 更新配置版本
	UpdateConfigVersion(ctx context.Context, tenantID, versionID string, configVersion *domain.ConfigVersion) error

	// DeleteConfigVersion 删除配置版本
	DeleteConfigVersion(ctx context.Context, tenantID, versionID string) error
}

