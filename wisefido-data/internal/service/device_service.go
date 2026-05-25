package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/publisher"
	"wisefido-data/internal/repository"
	"wisefido-data/internal/scope"

	"owl-common/card"

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

	SetConfigPublisher(pub *publisher.ConfigPublisher)
}

// deviceService 实现
type deviceService struct {
	devicesRepo     repository.DevicesRepository
	configPublisher *publisher.ConfigPublisher
	qinglanClient   *QinglanClient
	stateReader     *card.Reader
	db              *sql.DB
	logger          *zap.Logger
}

func NewDeviceService(devicesRepo repository.DevicesRepository, qinglanClient *QinglanClient, stateReader *card.Reader, db *sql.DB, logger *zap.Logger) DeviceService {
	return &deviceService{
		devicesRepo:   devicesRepo,
		qinglanClient: qinglanClient,
		stateReader:   stateReader,
		db:            db,
		logger:        logger,
	}
}

func (s *deviceService) SetConfigPublisher(pub *publisher.ConfigPublisher) {
	s.configPublisher = pub
}

// ListDevicesRequest 查询设备列表请求
type ListDevicesRequest struct {
	TenantID      string   // 必填（SystemAdmin 查看所有设备时，此字段仍需要但会被忽略）
	IsSystemAdmin bool     // SystemAdmin 查看所有租户的设备
	CurrentUserID string   // Phase 3：service 据此查 user_branches.is_primary 当 scope
	Status        []string // 可选：设备状态过滤（online, offline, error）
	DeviceType    string   // 可选：设备类型
	SearchType    string   // 可选：搜索类型（device_name, device_uid）
	SearchKeyword string   // 可选：搜索关键词
	BranchID      string   // 可选：显式 branch /56 CIDR 过滤（FE 列头筛选）；非空时覆盖 session 默认 BranchPrefix
	Page          int      // 可选，默认 1
	Size          int      // 可选，默认 20
	Sort          string   // 可选：排序字段（device_name, device_type, device_model, device_uid, ...）
	Direction     string   // 可选：asc / desc，默认 asc
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
		Status:        statuses,
		DeviceType:    strings.TrimSpace(req.DeviceType),
		SearchType:    strings.TrimSpace(req.SearchType),
		SearchKeyword: strings.TrimSpace(req.SearchKeyword),
		IsSystemAdmin: req.IsSystemAdmin,
	}

	// FE 列头显式 branch 过滤优先：用户从下拉选了 branch_id 则直接使用，覆盖 session 默认
	// 多值（"id1,id2"）取第一段；后续如需多 branch OR 过滤再加 BranchPrefixes []string。
	if bid := strings.TrimSpace(req.BranchID); bid != "" {
		if idx := strings.Index(bid, ","); idx > 0 {
			bid = strings.TrimSpace(bid[:idx])
		}
		filters.BranchPrefix = bid
	} else if !req.IsSystemAdmin {
		// Phase 3: 无显式 branch_id 时走 Current Branch scope via scope.ScopeContext
		if sc := scope.MustFromContext(ctx); sc != nil {
			if !sc.IsTenantWide() && sc.HasCurrentBranch() {
				filters.BranchPrefix = sc.CurrentBranchID
			}
		} else if strings.TrimSpace(req.CurrentUserID) != "" {
			// fallback：ctx 没注入时旧 SQL 路径
			var bid2 sql.NullString
			err := s.db.QueryRowContext(ctx, `
				SELECT host(branch_prefix) || '/56'
				  FROM user_branches
				 WHERE user_id = $1::UUID AND is_primary = TRUE AND valid_to IS NULL
				 LIMIT 1`, req.CurrentUserID).Scan(&bid2)
			if err == nil && bid2.Valid && bid2.String != "" {
				filters.BranchPrefix = bid2.String
			}
		}
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

	// 5. 调用 Repository（sort/direction 用于全部数据排序后分页）
	sort := strings.TrimSpace(req.Sort)
	direction := strings.TrimSpace(req.Direction)
	if direction != "desc" {
		direction = "asc"
	}
	items, total, err := s.devicesRepo.ListDevices(ctx, req.TenantID, filters, page, size, sort, direction)
	if err != nil {
		s.logger.Error("ListDevices failed",
			zap.String("tenant_id", req.TenantID),
			zap.Bool("is_system_admin", req.IsSystemAdmin),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	if len(items) > 0 {
		s.fillDeviceOnlineStatus(ctx, items)
	}

	return &ListDevicesResponse{
		Items: items,
		Total: total,
	}, nil
}

// FillDeviceOnlineStatusFromCardagg 只从 cardagg 读在线状态，返回 ipv6 -> "online"|"offline"。
// Phase A：device_status 已迁出 card:state，按 device_ipv6 直读 device:status:{IPv6} 独立 Hash。
//
// online 判定 = `ds.Offline == 0`（仅网络/心跳维度）。
// SensorDetached / SignalPoor / AngleAbnormal 是独立的 degraded 标志位，**不**聚合到 offline——
// 字段语义保持单一职责，让前端按需决定怎么渲染（例如 Sleepad 卡片在 SensorDetached 时显示
// "Sensor Detached" 文字，Detail 页显示 Online 黄字提示 degraded 等）。
//
// 仅返回字符串状态；要拿完整 DeviceStatus（含 signal_poor/angle_abnormal/sensor_detached/last_seen_ms）
// 用 FillDeviceStatusFromCardagg。
//
// 入参 deviceIPv6s：list of host(device_ipv6) 字符串（cardagg redis key 即此）。
func FillDeviceOnlineStatusFromCardagg(ctx context.Context, stateReader *card.Reader, deviceIPv6s []string, logger *zap.Logger) map[string]string {
	full := FillDeviceStatusFromCardagg(ctx, stateReader, deviceIPv6s, logger)
	out := make(map[string]string, len(deviceIPv6s))
	for _, addr := range deviceIPv6s {
		if addr == "" {
			continue
		}
		ds := full[addr]
		if ds != nil && ds.Offline == 0 {
			out[addr] = "online"
		} else {
			out[addr] = "offline"
		}
	}
	return out
}

// FillDeviceStatusFromCardagg 批量读 device:status:{IPv6} Hash，返回 ipv6 → DeviceStatus map。
// 缺失/读失败的 device 在 map 中不存在；调用方负责决定 fallback（典型：构造 offline=1 占位）。
//
// cardagg 写的 redis key 是 device:status:{IPv6}，调用方必须传 IPv6 字符串。
func FillDeviceStatusFromCardagg(ctx context.Context, stateReader *card.Reader, deviceIPv6s []string, logger *zap.Logger) map[string]*card.DeviceStatus {
	if stateReader == nil || len(deviceIPv6s) == 0 {
		return make(map[string]*card.DeviceStatus)
	}
	ids := make([]string, 0, len(deviceIPv6s))
	for _, id := range deviceIPv6s {
		if id != "" {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return make(map[string]*card.DeviceStatus)
	}
	devStatusBatch, err := stateReader.ReadDeviceStatusByDeviceAddrs(ctx, ids)
	if err != nil {
		if logger != nil {
			logger.Warn("FillDeviceStatusFromCardagg: ReadDeviceStatusByDeviceAddrs failed", zap.Error(err))
		}
		return make(map[string]*card.DeviceStatus)
	}
	return devStatusBatch
}

// fillDeviceOnlineStatus 只从 cardagg 读在线状态，使用 FillDeviceOnlineStatusFromCardagg。
// 用 d.DeviceAddr 作 cardagg redis key（device:status:{addr}）。
func (s *deviceService) fillDeviceOnlineStatus(ctx context.Context, devices []*domain.Device) {
	for _, d := range devices {
		d.OnlineStatus = "offline"
	}
	if s.stateReader == nil {
		return
	}
	addrs := make([]string, 0, len(devices))
	for _, d := range devices {
		if d.DeviceAddr != "" {
			addrs = append(addrs, d.DeviceAddr)
		}
	}
	m := FillDeviceOnlineStatusFromCardagg(ctx, s.stateReader, addrs, s.logger)
	for _, d := range devices {
		if d.DeviceAddr != "" && m[d.DeviceAddr] == "online" {
			d.OnlineStatus = "online"
		}
	}
}

// GetDeviceRequest 查询设备详情请求
type GetDeviceRequest struct {
	TenantID string // 必填
	DeviceAddr string // 必填
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
	if req.DeviceAddr == "" {
		return nil, fmt.Errorf("device_addr is required")
	}

	// 2. 调用 Repository
	device, err := s.devicesRepo.GetDevice(ctx, req.TenantID, req.DeviceAddr)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.Warn("Device not found",
				zap.String("tenant_id", req.TenantID),
				zap.String("device_addr", req.DeviceAddr),
			)
			return nil, fmt.Errorf("device not found")
		}
		s.logger.Error("GetDevice failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("device_addr", req.DeviceAddr),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get device")
	}

	s.fillDeviceOnlineStatus(ctx, []*domain.Device{device})

	return &GetDeviceResponse{
		Device: device,
	}, nil
}

// UpdateDeviceRequest 更新设备请求
type UpdateDeviceRequest struct {
	TenantID          string         // 必填
	DeviceAddr          string         // 必填
	Device            *domain.Device // 设备信息（部分更新）
	UpdateBoundRoomID bool           // 是否更新 bound_room_id（即使为 null）
	UpdateBoundBedID  bool           // 是否更新 bound_bed_id（即使为 null）
	UpdateBranchID    bool           // 是否 rebind 到指定 branch（仅在 bed/room 都不指定时生效）
	// 标记哪些字段在 payload 中（用于部分更新）
	UpdateAccess            bool // 是否更新 business_access
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
	if req.DeviceAddr == "" {
		return nil, fmt.Errorf("device_addr is required")
	}
	if req.Device == nil {
		return nil, fmt.Errorf("device is required")
	}

	// 2. 取更新前设备（用于同步旧 unit 卡片）
	var oldDevice *domain.Device
	oldDevice, _ = s.devicesRepo.GetDevice(ctx, req.TenantID, req.DeviceAddr)

	// 3. 业务规则验证
	// 注意：unit_id 验证在 Handler 层处理（因为 domain.Device 中没有 unit_id 字段）
	// Service 层只验证 bound_room_id 和 bound_bed_id 的逻辑

	// 4. 调用 Repository（传递更新标志）
	if err := s.devicesRepo.UpdateDeviceWithFlags(ctx, req.TenantID, req.DeviceAddr, req.Device, req.UpdateBoundRoomID, req.UpdateBoundBedID, req.UpdateAccess, req.UpdateMonitoringEnabled, req.UpdateBranchID); err != nil {
		s.logger.Error("UpdateDevice failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("device_addr", req.DeviceAddr),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to update device")
	}

	// 5. 任意 devices 更新后同步受影响 unit 的卡片（旧 unit + 新 unit，换绑时两边都刷）
	newDevice, err := s.devicesRepo.GetDevice(ctx, req.TenantID, req.DeviceAddr)
	if err != nil {
		s.logger.Warn("Failed to get updated device for card sync", zap.Error(err), zap.String("tenant_id", req.TenantID), zap.String("device_addr", req.DeviceAddr))
	}

	// 任意 devices 字段更新均发 config.changed，供网关刷新 baseline / health。
	// device_uid 优先取 newDevice；bind 后 device_addr 变化导致 GetDevice(req.DeviceAddr) 找不到
	// 时（newDevice=nil），回退用 oldDevice.DeviceUID（device_uid 是稳定身份，bind 不变）。
	// device_addrs 故意用 req.DeviceAddr（pre-bind addr）：cardagg.enablement cache 持有的就是旧
	// addr，invalidate 旧 addr 正是要的语义；新 addr 还没有 cache 条目，无需特意带。
	if s.configPublisher != nil {
		var uids []string
		switch {
		case newDevice != nil && newDevice.DeviceUID != "":
			uids = []string{newDevice.DeviceUID}
		case oldDevice != nil && oldDevice.DeviceUID != "":
			uids = []string{oldDevice.DeviceUID}
		}
		if err := s.configPublisher.PublishConfigChanged(ctx, "update", nil, []string{req.DeviceAddr}, uids); err != nil {
			s.logger.Warn("PublishConfigChanged failed",
				zap.String("tenant_id", req.TenantID),
				zap.String("device_addr", req.DeviceAddr),
				zap.Error(err))
		}
	}

	if err == nil && newDevice != nil {
		// 新旧 unit 都重算（device bind 跨 unit 时 has_bed/split rule 双向受影响）
		if oldDevice != nil && oldDevice.UnitID.Valid && oldDevice.UnitID.String != "" {
			if rerr := globalCardSync.ReconcileCards(ctx, oldDevice.UnitID.String); rerr != nil {
				s.logger.Warn("UpdateDevice: ReconcileCards (old unit) failed",
					zap.String("unit_id", oldDevice.UnitID.String), zap.Error(rerr))
			}
		}
		if newDevice.UnitID.Valid && newDevice.UnitID.String != "" {
			if rerr := globalCardSync.ReconcileCards(ctx, newDevice.UnitID.String); rerr != nil {
				s.logger.Warn("UpdateDevice: ReconcileCards (new unit) failed",
					zap.String("unit_id", newDevice.UnitID.String), zap.Error(rerr))
			}
		}
	}

	// 绑定关系快照：bound_room_id 或 bound_bed_id 发生变更时
	if oldDevice != nil && newDevice != nil {
		oldRoom := ""
		if oldDevice.RoomID.Valid {
			oldRoom = oldDevice.RoomID.String
		}
		newRoom := ""
		if newDevice.RoomID.Valid {
			newRoom = newDevice.RoomID.String
		}
		oldBed := ""
		if oldDevice.BoundBedID.Valid {
			oldBed = oldDevice.BoundBedID.String
		}
		newBed := ""
		if newDevice.BoundBedID.Valid {
			newBed = newDevice.BoundBedID.String
		}
		if oldRoom != newRoom || oldBed != newBed {
			var affectedUnits []string
			if oldDevice.UnitID.Valid && oldDevice.UnitID.String != "" {
				affectedUnits = append(affectedUnits, oldDevice.UnitID.String)
			}
			if newDevice.UnitID.Valid && newDevice.UnitID.String != "" && (len(affectedUnits) == 0 || affectedUnits[0] != newDevice.UnitID.String) {
				affectedUnits = append(affectedUnits, newDevice.UnitID.String)
			}
			summary := fmt.Sprintf("%s(%s): %s → %s",
				newDevice.DeviceName, newDevice.DeviceUID,
				snapshotLocationLabel(ctx, s.db, req.TenantID, oldRoom, oldBed),
				snapshotLocationLabel(ctx, s.db, req.TenantID, newRoom, newBed))
			if err := TakeBindingSnapshot(ctx, s.db, s.logger, req.TenantID, "device_binding", req.DeviceAddr, summary, affectedUnits, ""); err != nil {
				s.logger.Warn("TakeBindingSnapshot failed", zap.Error(err))
			}
		}
	}

	return &UpdateDeviceResponse{
		Success: true,
	}, nil
}

// DeleteDeviceRequest 删除设备请求
type DeleteDeviceRequest struct {
	TenantID string // 必填
	DeviceAddr string // 必填
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
	if req.DeviceAddr == "" {
		return nil, fmt.Errorf("device_addr is required")
	}

	// 2. 获取设备信息（删除前获取 unit_id / device_uid，用于后续通知）
	var unitID, deviceUID string
	device, err := s.devicesRepo.GetDevice(ctx, req.TenantID, req.DeviceAddr)
	if err == nil && device != nil {
		deviceUID = device.DeviceUID
		if device.UnitID.Valid && device.UnitID.String != "" {
			unitID = device.UnitID.String
		}
	}

	// 3. 调用 Repository（硬删除：在事务中删除 devices 记录，更新 device_store tenant_id）
	// 先执行数据库操作，保证事务的原子性
	if err := s.devicesRepo.DeleteDevice(ctx, req.TenantID, req.DeviceAddr); err != nil {
		s.logger.Error("DeleteDevice failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("device_addr", req.DeviceAddr),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to delete device: %w", err)
	}

	if unitID != "" && globalCardSync != nil {
		if rerr := globalCardSync.ReconcileCards(ctx, unitID); rerr != nil {
			s.logger.Warn("DeleteDevice: ReconcileCards failed",
				zap.String("unit_id", unitID), zap.Error(rerr))
		}
	}

	// 发送 device 删除变化信号
	if s.configPublisher != nil {
		var uids []string
		if deviceUID != "" {
			uids = []string{deviceUID}
		}
		if err := s.configPublisher.PublishConfigChanged(ctx, "update", nil, []string{req.DeviceAddr}, uids); err != nil {
			s.logger.Warn("Failed to publish device deletion signal",
				zap.String("device_addr", req.DeviceAddr),
				zap.Error(err),
			)
		}
	}

	s.logger.Info("Device deleted successfully",
		zap.String("tenant_id", req.TenantID),
		zap.String("device_addr", req.DeviceAddr),
		zap.String("unit_id", unitID),
	)

	return &DeleteDeviceResponse{
		Success: true,
	}, nil
}

// GetDeviceRelationsRequest 查询设备关联关系请求
type GetDeviceRelationsRequest struct {
	TenantID string // 必填
	DeviceAddr string // 必填
}

// GetDeviceRelationsResponse 查询设备关联关系响应
type GetDeviceRelationsResponse struct {
	DeviceAddr           string
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
	if req.DeviceAddr == "" {
		return nil, fmt.Errorf("device_addr is required")
	}

	// 2. 调用 Repository
	relations, err := s.devicesRepo.GetDeviceRelations(ctx, req.TenantID, req.DeviceAddr)
	if err != nil {
		s.logger.Error("GetDeviceRelations failed",
			zap.String("tenant_id", req.TenantID),
			zap.String("device_addr", req.DeviceAddr),
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
		DeviceAddr:           relations.DeviceAddr,
		DeviceName:         relations.DeviceName,
		DeviceInternalCode: relations.DeviceInternalCode,
		DeviceType:         relations.DeviceType,
		AddressID:          relations.AddressID,
		AddressName:        relations.AddressName,
		AddressType:        relations.AddressType,
		Residents:          residents,
	}, nil
}
