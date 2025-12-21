package repository

import (
	"context"

	"wisefido-data/internal/domain"
)

// SNOMEDMappingFilters SNOMED映射查询过滤器
type SNOMEDMappingFilters struct {
	Category        string // FHIR Category
	FirmwareVersion string // 固件版本
}

// SNOMEDMappingRepository SNOMED编码映射Repository接口
type SNOMEDMappingRepository interface {
	// GetMapping 获取映射（按mapping_type和source_value）
	GetMapping(ctx context.Context, mappingType, sourceValue string) (*domain.SNOMEDMapping, error)

	// GetPostureMapping 获取姿态映射（支持固件版本）
	// 如果指定了firmwareVersion，优先匹配该版本，否则匹配通用版本（firmware_version IS NULL）
	GetPostureMapping(ctx context.Context, sourceValue string, firmwareVersion *string) (*domain.SNOMEDMapping, error)

	// GetEventMapping 获取事件映射
	GetEventMapping(ctx context.Context, sourceValue string) (*domain.SNOMEDMapping, error)

	// ListMappings 列表查询（支持按类型、category、固件版本过滤）
	ListMappings(ctx context.Context, mappingType string, filters *SNOMEDMappingFilters, page, size int) ([]*domain.SNOMEDMapping, int, error)

	// CreateMapping 创建映射
	CreateMapping(ctx context.Context, mapping *domain.SNOMEDMapping) error

	// UpdateMapping 更新映射
	UpdateMapping(ctx context.Context, mappingID string, mapping *domain.SNOMEDMapping) error

	// DeleteMapping 删除映射
	DeleteMapping(ctx context.Context, mappingID string) error
}

