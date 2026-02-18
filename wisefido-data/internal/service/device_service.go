package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/publisher"
	"wisefido-data/internal/repository"

	rediscommon "owl-common/redis"

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
	devicesRepo     repository.DevicesRepository
	cardSync        *CardSyncService // 设备/单元变更后同步卡片（Redis + config.card.*）
	configPublisher *publisher.ConfigPublisher
	qinglanClient   *QinglanClient
	logger          *zap.Logger
}

// NewDeviceService 创建 DeviceService 实例
func NewDeviceService(devicesRepo repository.DevicesRepository, cardSync *CardSyncService, qinglanClient *QinglanClient, logger *zap.Logger) DeviceService {
	return &deviceService{
		devicesRepo:   devicesRepo,
		cardSync:      cardSync,
		qinglanClient: qinglanClient,
		logger:        logger,
	}
}

// NewDeviceServiceWithPublisher 创建 DeviceService 实例（包含 ConfigPublisher 用于发送 device_store 变化信号）
func NewDeviceServiceWithPublisher(devicesRepo repository.DevicesRepository, cardSync *CardSyncService, configPublisher *publisher.ConfigPublisher, qinglanClient *QinglanClient, logger *zap.Logger) DeviceService {
	return &deviceService{
		devicesRepo:     devicesRepo,
		cardSync:        cardSync,
		configPublisher: configPublisher,
		qinglanClient:   qinglanClient,
		logger:          logger,
	}
}

// ListDevicesRequest 查询设备列表请求
type ListDevicesRequest struct {
	TenantID       string   // 必填（SystemAdmin 查看所有设备时，此字段仍需要但会被忽略）
	IsSystemAdmin  bool     // SystemAdmin 查看所有租户的设备
	Status         []string // 可选：设备状态过滤（online, offline, error）
	BusinessAccess string   // 可选：业务访问权限（pending, approved, rejected）
	DeviceType     string   // 可选：设备类型
	SearchType     string   // 可选：搜索类型（device_name, device_uid）
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
		IsSystemAdmin:  req.IsSystemAdmin, // SystemAdmin 查看所有设备
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
	// SystemAdmin 查看所有设备时，tenantID 会被忽略
	items, total, err := s.devicesRepo.ListDevices(ctx, req.TenantID, filters, page, size)
	if err != nil {
		s.logger.Error("ListDevices failed",
			zap.String("tenant_id", req.TenantID),
			zap.Bool("is_system_admin", req.IsSystemAdmin),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	// 6. 通过 wisefido-qinglan HTTP API 查询设备在线状态并填充到 items
	if s.qinglanClient != nil && len(items) > 0 {
		s.fillDeviceOnlineStatus(ctx, items)
	} else if s.qinglanClient == nil {
		// qinglanClient 为 nil，所有设备默认设置为 "offline"
		for _, device := range items {
			if device.OnlineStatus == "" {
				device.OnlineStatus = "offline"
			}
		}
	}

	return &ListDevicesResponse{
		Items: items,
		Total: total,
	}, nil
}

// fillDeviceOnlineStatus 通过 wisefido-qinglan HTTP API 批量查询设备在线状态并填充到设备列表
// 使用并发查询提高性能
func (s *deviceService) fillDeviceOnlineStatus(ctx context.Context, devices []*domain.Device) {
	if len(devices) == 0 {
		return
	}

	// 过滤出有 device_uid 的设备
	devicesWithUID := make([]*domain.Device, 0, len(devices))
	for _, device := range devices {
		if device.DeviceUID != "" {
			devicesWithUID = append(devicesWithUID, device)
		}
	}

	if len(devicesWithUID) == 0 {
		// 没有 device_uid 的设备，默认设置为 "offline"
		for _, device := range devices {
			if device.OnlineStatus == "" {
				device.OnlineStatus = "offline"
			}
		}
		return
	}

	// 并发查询设备状态
	type statusResult struct {
		deviceUID string
		status    string
		err       error
	}
	results := make(chan statusResult, len(devicesWithUID))

	// 启动 goroutine 并发查询
	for _, device := range devicesWithUID {
		go func(d *domain.Device) {
			status, err := s.qinglanClient.GetDeviceStatus(ctx, d.DeviceUID)
			if err != nil {
				// 查询失败，根据错误类型决定状态
				// 如果是连接被拒绝或网络错误，表示无法确定设备状态，应设为"unknown"
				statusValue := "offline"
				if strings.Contains(err.Error(), "connection refused") || 
				   strings.Contains(err.Error(), "timeout") || 
				   strings.Contains(err.Error(), "network") ||
				   strings.Contains(err.Error(), "no route to host") ||
				   strings.Contains(err.Error(), "connection reset by peer") {
					statusValue = "unknown"
				}
				results <- statusResult{deviceUID: d.DeviceUID, status: statusValue, err: err}
			} else {
				results <- statusResult{deviceUID: d.DeviceUID, status: status, err: nil}
			}
		}(device)
	}

	// 收集结果
	deviceUIDToStatus := make(map[string]string)
	for i := 0; i < len(devicesWithUID); i++ {
		result := <-results
		deviceUIDToStatus[result.deviceUID] = result.status
	}

	// 填充状态到设备列表
	for _, device := range devices {
		if device.DeviceUID != "" {
			if status, exists := deviceUIDToStatus[device.DeviceUID]; exists {
				device.OnlineStatus = status
			} else {
				device.OnlineStatus = "offline"
			}
		} else {
			// 没有 device_uid 的设备，默认设置为 "offline"
			if device.OnlineStatus == "" {
				device.OnlineStatus = "offline"
			}
		}
	}
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
	TenantID          string         // 必填
	DeviceID          string         // 必填
	Device            *domain.Device // 设备信息（部分更新）
	UpdateBoundRoomID bool           // 是否更新 bound_room_id（即使为 null）
	UpdateBoundBedID  bool           // 是否更新 bound_bed_id（即使为 null）
	// 标记哪些字段在 payload 中（用于部分更新）
	UpdateBusinessAccess    bool // 是否更新 business_access
	UpdateMonitoringEnabled bool // 是否更新 monitoring_enabled
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

	// 2. 获取旧设备信息（用于比较 monitoring_enabled 是否变化）
	var oldDevice *domain.Device
	var monitoringEnabledChanged bool
	oldDevice, _ = s.devicesRepo.GetDevice(ctx, req.TenantID, req.DeviceID)
	if oldDevice != nil && req.Device.MonitoringEnabled != oldDevice.MonitoringEnabled {
		monitoringEnabledChanged = true
	}

	// 3. 业务规则验证
	// 注意：unit_id 验证在 Handler 层处理（因为 domain.Device 中没有 unit_id 字段）
	// Service 层只验证 bound_room_id 和 bound_bed_id 的逻辑

	// 4. 调用 Repository（传递更新标志）
	if err := s.devicesRepo.UpdateDeviceWithFlags(ctx, req.TenantID, req.DeviceID, req.Device, req.UpdateBoundRoomID, req.UpdateBoundBedID, req.UpdateBusinessAccess, req.UpdateMonitoringEnabled); err != nil {
		s.logger.Error("UpdateDevice failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("device_id", req.DeviceID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to update device")
	}

	// 5. 仅在 monitoring_enabled 变化时触发相关服务更新
	if monitoringEnabledChanged {
		// ...existing code...

		// 仅在 monitoringEnabledChanged 时触发卡片同步
		if monitoringEnabledChanged && s.cardSync != nil {
			newDevice, err := s.devicesRepo.GetDevice(ctx, req.TenantID, req.DeviceID)
			if err != nil {
				s.logger.Warn("Failed to get updated device for card sync", zap.Error(err), zap.String("tenant_id", req.TenantID), zap.String("device_id", req.DeviceID))
			} else if newDevice != nil && newDevice.UnitID.Valid && newDevice.UnitID.String != "" {
				if _, err := s.cardSync.CreateCardsForUnit(ctx, req.TenantID, newDevice.UnitID.String); err != nil {
					s.logger.Warn("Failed to sync cards after monitoring status change", zap.Error(err), zap.String("tenant_id", req.TenantID), zap.String("device_id", req.DeviceID), zap.String("unit_id", newDevice.UnitID.String))
				} else {
					s.logger.Info("Synced cards after monitoring status change", zap.String("tenant_id", req.TenantID), zap.String("device_id", req.DeviceID), zap.String("unit_id", newDevice.UnitID.String))
				}
			}
		}

		// ...existing code...
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

// DeleteDevice 删除设备（软删除：移至 Trash 租户）
// 功能：将设备移至 Trash 租户，设置 business_access='rejected', monitoring_enabled=FALSE
// 流程：先执行数据库事务（保证原子性），事务提交成功后再通知 card_manager
func (s *deviceService) DeleteDevice(ctx context.Context, req DeleteDeviceRequest) (*DeleteDeviceResponse, error) {
	// 1. 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.DeviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}

	// 2. 获取设备信息（删除前获取 unit_id，用于后续通知 card_manager）
	var unitID string
	device, err := s.devicesRepo.GetDevice(ctx, req.TenantID, req.DeviceID)
	if err == nil && device != nil && device.UnitID.Valid && device.UnitID.String != "" {
		unitID = device.UnitID.String
	}

	// 3. 调用 Repository（硬删除：在事务中删除 devices 记录，更新 device_store tenant_id）
	// 先执行数据库操作，保证事务的原子性
	if err := s.devicesRepo.DeleteDevice(ctx, req.TenantID, req.DeviceID); err != nil {
		s.logger.Error("DeleteDevice failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("device_id", req.DeviceID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to delete device: %w", err)
	}

	if s.cardSync != nil && unitID != "" {
		if _, err := s.cardSync.CreateCardsForUnit(ctx, req.TenantID, unitID); err != nil {
			s.logger.Warn("Failed to sync cards after device deletion", zap.Error(err), zap.String("tenant_id", req.TenantID), zap.String("device_id", req.DeviceID), zap.String("unit_id", unitID))
		} else {
			s.logger.Info("Synced cards after device deletion", zap.String("tenant_id", req.TenantID), zap.String("device_id", req.DeviceID), zap.String("unit_id", unitID))
		}
	}

	// 发送 device_store 变化信号（device_deleted）
	if s.configPublisher != nil {
		extraData := map[string]interface{}{
			"device_id":   req.DeviceID,
			"change_type": "device_deleted",
		}
		if err := s.configPublisher.PublishCardChangeMessageWithExtraAndType(ctx, req.TenantID, "", unitID, "", rediscommon.ConfigCardDeviceStoreChanged, extraData); err != nil {
			s.logger.Warn("Failed to publish device_store change signal for device deletion",
				zap.String("device_id", req.DeviceID),
				zap.Error(err),
			)
		}
	}

	s.logger.Info("Device deleted successfully",
		zap.String("tenant_id", req.TenantID),
		zap.String("device_id", req.DeviceID),
		zap.String("unit_id", unitID),
	)

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
