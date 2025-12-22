package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

// DeviceService 设备管理服务接口
type DeviceService interface {
	// 查询
	ListDevices(ctx context.Context, req ListDevicesRequest) (*ListDevicesResponse, error)
	GetDevice(ctx context.Context, req GetDeviceRequest) (*GetDeviceResponse, error)
	GetDeviceRelations(ctx context.Context, req GetDeviceRelationsRequest) (*GetDeviceRelationsResponse, error)

	// 更新
	UpdateDevice(ctx context.Context, req UpdateDeviceRequest) (*UpdateDeviceResponse, error)

	// 删除
	DeleteDevice(ctx context.Context, req DeleteDeviceRequest) (*DeleteDeviceResponse, error)
}

// deviceService 实现
type deviceService struct {
	devicesRepo repository.DevicesRepository
	logger      *zap.Logger
}

// NewDeviceService 创建 DeviceService 实例
func NewDeviceService(devicesRepo repository.DevicesRepository, logger *zap.Logger) DeviceService {
	return &deviceService{
		devicesRepo: devicesRepo,
		logger:      logger,
	}
}

// ListDevicesRequest 查询设备列表请求
type ListDevicesRequest struct {
	TenantID       string   // 必填
	Status         []string // 可选：设备状态过滤（online, offline, error）
	BusinessAccess string   // 可选：业务访问权限（pending, approved, rejected）
	DeviceType     string   // 可选：设备类型
	SearchType     string   // 可选：搜索类型（device_name, serial_number, uid）
	SearchKeyword  string   // 可选：搜索关键词
	Page           int      // 可选，默认 1
	Size           int      // 可选，默认 20
}

// ListDevicesResponse 查询设备列表响应
type ListDevicesResponse struct {
	Items []*domain.Device // 设备列表
	Total int              // 总数量
}

// ListDevices 查询设备列表
func (s *deviceService) ListDevices(ctx context.Context, req ListDevicesRequest) (*ListDevicesResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	// 2. 处理 status 参数（支持逗号分隔）
	statuses := req.Status
	if len(statuses) == 1 && strings.Contains(statuses[0], ",") {
		statuses = strings.Split(statuses[0], ",")
		// 清理空格
		for i := range statuses {
			statuses[i] = strings.TrimSpace(statuses[i])
		}
	}

	// 3. 构建过滤器
	filters := repository.DeviceFilters{
		Status:         statuses,
		BusinessAccess: strings.TrimSpace(req.BusinessAccess),
		DeviceType:     strings.TrimSpace(req.DeviceType),
		SearchType:     strings.TrimSpace(req.SearchType),
		SearchKeyword:  strings.TrimSpace(req.SearchKeyword),
	}

	// 4. 分页参数
	page := req.Page
	if page <= 0 {
		page = 1
	}
	size := req.Size
	if size <= 0 {
		size = 20
	}

	// 5. 调用 Repository
	items, total, err := s.devicesRepo.ListDevices(ctx, req.TenantID, filters, page, size)
	if err != nil {
		s.logger.Error("ListDevices failed",
			zap.String("tenant_id", req.TenantID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to list devices")
	}

	return &ListDevicesResponse{
		Items: items,
		Total: total,
	}, nil
}

// GetDeviceRequest 查询设备详情请求
type GetDeviceRequest struct {
	TenantID string // 必填
	DeviceID string // 必填
}

// GetDeviceResponse 查询设备详情响应
type GetDeviceResponse struct {
	Device *domain.Device // 设备信息
}

// GetDevice 查询设备详情
func (s *deviceService) GetDevice(ctx context.Context, req GetDeviceRequest) (*GetDeviceResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.DeviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}

	// 2. 调用 Repository
	device, err := s.devicesRepo.GetDevice(ctx, req.TenantID, req.DeviceID)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.Warn("Device not found",
				zap.String("tenant_id", req.TenantID),
				zap.String("device_id", req.DeviceID),
			)
			return nil, fmt.Errorf("device not found")
		}
		s.logger.Error("GetDevice failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("device_id", req.DeviceID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get device")
	}

	return &GetDeviceResponse{
		Device: device,
	}, nil
}

// UpdateDeviceRequest 更新设备请求
type UpdateDeviceRequest struct {
	TenantID string         // 必填
	DeviceID string         // 必填
	Device   *domain.Device // 设备信息（部分更新）
}

// UpdateDeviceResponse 更新设备响应
type UpdateDeviceResponse struct {
	Success bool // 更新成功
}

// UpdateDevice 更新设备
func (s *deviceService) UpdateDevice(ctx context.Context, req UpdateDeviceRequest) (*UpdateDeviceResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.DeviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}
	if req.Device == nil {
		return nil, fmt.Errorf("device is required")
	}

	// 2. 业务规则验证
	// 注意：unit_id 验证在 Handler 层处理（因为 domain.Device 中没有 unit_id 字段）
	// Service 层只验证 bound_room_id 和 bound_bed_id 的逻辑

	// 3. 调用 Repository
	if err := s.devicesRepo.UpdateDevice(ctx, req.TenantID, req.DeviceID, req.Device); err != nil {
		s.logger.Error("UpdateDevice failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("device_id", req.DeviceID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to update device")
	}

	return &UpdateDeviceResponse{
		Success: true,
	}, nil
}

// DeleteDeviceRequest 删除设备请求
type DeleteDeviceRequest struct {
	TenantID string // 必填
	DeviceID string // 必填
}

// DeleteDeviceResponse 删除设备响应
type DeleteDeviceResponse struct {
	Success bool // 删除成功
}

// DeleteDevice 删除设备（软删除）
func (s *deviceService) DeleteDevice(ctx context.Context, req DeleteDeviceRequest) (*DeleteDeviceResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.DeviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}

	// 2. 调用 Repository（软删除）
	if err := s.devicesRepo.DisableDevice(ctx, req.TenantID, req.DeviceID); err != nil {
		s.logger.Error("DeleteDevice failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("device_id", req.DeviceID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to delete device")
	}

	return &DeleteDeviceResponse{
		Success: true,
	}, nil
}

// GetDeviceRelationsRequest 查询设备关联关系请求
type GetDeviceRelationsRequest struct {
	TenantID string // 必填
	DeviceID string // 必填
}

// GetDeviceRelationsResponse 查询设备关联关系响应
type GetDeviceRelationsResponse struct {
	DeviceID           string
	DeviceName         string
	DeviceInternalCode string
	DeviceType         int
	AddressID          string
	AddressName        string
	AddressType        int
	Residents          []DeviceRelationResidentItem
}

// DeviceRelationResidentItem 设备关联的住户信息
type DeviceRelationResidentItem struct {
	ID       string
	Name     string
	Gender   string
	Birthday string
}

// GetDeviceRelations 查询设备关联关系
func (s *deviceService) GetDeviceRelations(ctx context.Context, req GetDeviceRelationsRequest) (*GetDeviceRelationsResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.DeviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}

	// 2. 调用 Repository
	relations, err := s.devicesRepo.GetDeviceRelations(ctx, req.TenantID, req.DeviceID)
	if err != nil {
		s.logger.Error("GetDeviceRelations failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("device_id", req.DeviceID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get device relations: %w", err)
	}

	// 3. 转换响应格式
	residents := make([]DeviceRelationResidentItem, len(relations.Residents))
	for i, r := range relations.Residents {
		residents[i] = DeviceRelationResidentItem{
			ID:       r.ID,
			Name:     r.Name,
			Gender:   r.Gender,
			Birthday: r.Birthday,
		}
	}

	return &GetDeviceRelationsResponse{
		DeviceID:           relations.DeviceID,
		DeviceName:         relations.DeviceName,
		DeviceInternalCode: relations.DeviceInternalCode,
		DeviceType:         relations.DeviceType,
		AddressID:          relations.AddressID,
		AddressName:        relations.AddressName,
		AddressType:        relations.AddressType,
		Residents:          residents,
	}, nil
}

