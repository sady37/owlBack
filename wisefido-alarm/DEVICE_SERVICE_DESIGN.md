# DeviceService 设计文档

## 📋 设备管理需求分析

### 1. 设备状态管理

**设备状态值**：
- `online` - 设备在线
- `offline` - 设备离线
- `error` - 设备错误
- `disabled` - 设备已禁用

**状态转换规则**：
- `disabled` → `online`：需要业务访问权限为 `approved`
- `online` → `disabled`：允许（禁用设备）
- `offline` → `online`：允许（设备上线）
- `error` → `online`：允许（错误恢复）
- 其他转换：需要业务规则验证

**状态管理职责**：
- 验证状态转换是否合法
- 验证业务访问权限（`pending`, `approved`, `rejected`）
- 更新设备状态

---

### 2. 设备绑定管理

**绑定类型**：
- 绑定到 Room（`bound_room_id`）
- 绑定到 Bed（`bound_bed_id`）
- 互斥：设备不能同时绑定到 Room 和 Bed

**绑定验证规则**：
1. 验证 room/bed 是否属于该租户
2. 验证 room/bed 是否存在
3. 验证设备是否已绑定到其他位置（如果需要）
4. 验证设备状态（disabled 的设备不能绑定）

**绑定变更后的业务编排**：
- 发布 card 更新事件（通知 card-aggregator 重新聚合）
- 更新设备状态（如果需要）
- 更新监控状态（`monitoring_enabled`）

---

### 3. 权限检查

**设备管理权限**：
- 查看设备：所有有权限的用户
- 更新设备：Admin, Manager, IT
- 绑定设备：Admin, Manager, IT
- 禁用设备：Admin, Manager

**设备绑定权限**：
- 绑定到 Room：需要验证用户是否有权限访问该 Unit
- 绑定到 Bed：需要验证用户是否有权限访问该 Room

---

## 🏗️ DeviceService 设计

### 接口定义

```go
package service

import (
    "context"
    "wisefido-data/internal/domain"
    "wisefido-data/internal/repository"
    "go.uber.org/zap"
)

type DeviceService struct {
    devicesRepo *repository.DevicesRepository
    unitsRepo  *repository.UnitsRepository
    permissionChecker *PermissionChecker
    eventPublisher *EventPublisher // 用于发布 card 更新事件
    logger *zap.Logger
}

func NewDeviceService(
    devicesRepo *repository.DevicesRepository,
    unitsRepo *repository.UnitsRepository,
    permissionChecker *PermissionChecker,
    eventPublisher *EventPublisher,
    logger *zap.Logger,
) *DeviceService {
    return &DeviceService{
        devicesRepo: devicesRepo,
        unitsRepo: unitsRepo,
        permissionChecker: permissionChecker,
        eventPublisher: eventPublisher,
        logger: logger,
    }
}
```

### 方法定义

#### 1. CRUD 方法

```go
// ListDevices 获取设备列表
func (s *DeviceService) ListDevices(
    ctx context.Context,
    tenantID, userID, userRole string,
    filters repository.DeviceFilters,
    page, size int,
) ([]*domain.Device, int, error) {
    // 1. 权限检查
    if !s.permissionChecker.CanViewDevices(ctx, tenantID, userID, userRole) {
        return nil, 0, ErrPermissionDenied
    }
    
    // 2. 调用 Repository
    return s.devicesRepo.ListDevices(ctx, tenantID, filters, page, size)
}

// GetDevice 获取设备详情
func (s *DeviceService) GetDevice(
    ctx context.Context,
    tenantID, userID, userRole, deviceID string,
) (*domain.Device, error) {
    // 1. 权限检查
    if !s.permissionChecker.CanViewDevices(ctx, tenantID, userID, userRole) {
        return nil, ErrPermissionDenied
    }
    
    // 2. 调用 Repository
    return s.devicesRepo.GetDevice(ctx, tenantID, deviceID)
}

// UpdateDevice 更新设备信息
func (s *DeviceService) UpdateDevice(
    ctx context.Context,
    tenantID, userID, userRole, deviceID string,
    payload map[string]any,
) error {
    // 1. 权限检查
    if !s.permissionChecker.CanUpdateDevices(ctx, tenantID, userID, userRole) {
        return ErrPermissionDenied
    }
    
    // 2. 获取当前设备
    device, err := s.devicesRepo.GetDevice(ctx, tenantID, deviceID)
    if err != nil {
        return err
    }
    
    // 3. 业务规则验证
    if err := s.validateDeviceUpdate(device, payload); err != nil {
        return err
    }
    
    // 4. 数据转换
    updatedDevice := s.convertPayloadToDevice(payload, device)
    
    // 5. 调用 Repository
    if err := s.devicesRepo.UpdateDevice(ctx, tenantID, deviceID, updatedDevice); err != nil {
        return err
    }
    
    // 6. 业务编排：如果绑定变更，发布 card 更新事件
    if s.isBindingChanged(device, updatedDevice) {
        if err := s.publishCardUpdateEvent(ctx, deviceID); err != nil {
            s.logger.Warn("failed to publish card update event", zap.Error(err))
        }
    }
    
    return nil
}
```

#### 2. 设备状态管理

```go
// UpdateDeviceStatus 更新设备状态
func (s *DeviceService) UpdateDeviceStatus(
    ctx context.Context,
    tenantID, userID, userRole, deviceID, newStatus string,
) error {
    // 1. 权限检查
    if !s.permissionChecker.CanUpdateDevices(ctx, tenantID, userID, userRole) {
        return ErrPermissionDenied
    }
    
    // 2. 获取当前设备
    device, err := s.devicesRepo.GetDevice(ctx, tenantID, deviceID)
    if err != nil {
        return err
    }
    
    // 3. 验证状态转换
    if err := s.validateStatusTransition(device.Status, newStatus, device.BusinessAccess); err != nil {
        return err
    }
    
    // 4. 更新状态
    updatedDevice := &domain.Device{
        Status: newStatus,
    }
    
    return s.devicesRepo.UpdateDevice(ctx, tenantID, deviceID, updatedDevice)
}

// DisableDevice 禁用设备
func (s *DeviceService) DisableDevice(
    ctx context.Context,
    tenantID, userID, userRole, deviceID string,
) error {
    // 1. 权限检查
    if !s.permissionChecker.CanDisableDevices(ctx, tenantID, userID, userRole) {
        return ErrPermissionDenied
    }
    
    // 2. 调用 Repository（软删除）
    return s.devicesRepo.DisableDevice(ctx, tenantID, deviceID)
}

// validateStatusTransition 验证状态转换
func (s *DeviceService) validateStatusTransition(
    oldStatus, newStatus, businessAccess string,
) error {
    // 状态转换规则
    switch {
    case oldStatus == newStatus:
        return nil // 无变化
    case newStatus == "online" && oldStatus == "disabled":
        // disabled → online：需要业务访问权限为 approved
        if businessAccess != "approved" {
            return fmt.Errorf("cannot enable device: business_access must be approved")
        }
    case newStatus == "disabled":
        // 任何状态 → disabled：允许
        return nil
    case newStatus == "online" && (oldStatus == "offline" || oldStatus == "error"):
        // offline/error → online：允许
        return nil
    default:
        // 其他转换：需要业务规则验证
        return fmt.Errorf("invalid status transition: %s → %s", oldStatus, newStatus)
    }
    
    return nil
}
```

#### 3. 设备绑定管理

```go
// BindDeviceToRoom 绑定设备到房间
func (s *DeviceService) BindDeviceToRoom(
    ctx context.Context,
    tenantID, userID, userRole, deviceID, roomID string,
) error {
    // 1. 权限检查
    if !s.permissionChecker.CanBindDevices(ctx, tenantID, userID, userRole) {
        return ErrPermissionDenied
    }
    
    // 2. 验证 room 是否属于该租户
    room, err := s.unitsRepo.GetRoom(ctx, tenantID, roomID)
    if err != nil {
        return fmt.Errorf("room not found: %w", err)
    }
    
    // 3. 验证用户是否有权限访问该 Unit
    if !s.permissionChecker.CanAccessUnit(ctx, tenantID, userID, userRole, room.UnitID) {
        return ErrPermissionDenied
    }
    
    // 4. 获取当前设备
    device, err := s.devicesRepo.GetDevice(ctx, tenantID, deviceID)
    if err != nil {
        return err
    }
    
    // 5. 验证设备状态
    if device.Status == "disabled" {
        return fmt.Errorf("cannot bind disabled device")
    }
    
    // 6. 验证绑定规则（不能同时绑定 room 和 bed）
    if device.BoundBedID.Valid {
        return fmt.Errorf("device is already bound to bed: %s", device.BoundBedID.String)
    }
    
    // 7. 更新绑定
    updatedDevice := &domain.Device{
        BoundRoomID: sql.NullString{String: roomID, Valid: true},
        BoundBedID:  sql.NullString{Valid: false}, // 清除 bed 绑定
    }
    
    if err := s.devicesRepo.UpdateDevice(ctx, tenantID, deviceID, updatedDevice); err != nil {
        return err
    }
    
    // 8. 发布 card 更新事件
    if err := s.publishCardUpdateEvent(ctx, deviceID); err != nil {
        s.logger.Warn("failed to publish card update event", zap.Error(err))
    }
    
    return nil
}

// BindDeviceToBed 绑定设备到床位
func (s *DeviceService) BindDeviceToBed(
    ctx context.Context,
    tenantID, userID, userRole, deviceID, bedID string,
) error {
    // 1. 权限检查
    if !s.permissionChecker.CanBindDevices(ctx, tenantID, userID, userRole) {
        return ErrPermissionDenied
    }
    
    // 2. 验证 bed 是否属于该租户
    bed, err := s.unitsRepo.GetBed(ctx, tenantID, bedID)
    if err != nil {
        return fmt.Errorf("bed not found: %w", err)
    }
    
    // 3. 验证用户是否有权限访问该 Room
    if !s.permissionChecker.CanAccessRoom(ctx, tenantID, userID, userRole, bed.RoomID) {
        return ErrPermissionDenied
    }
    
    // 4. 获取当前设备
    device, err := s.devicesRepo.GetDevice(ctx, tenantID, deviceID)
    if err != nil {
        return err
    }
    
    // 5. 验证设备状态
    if device.Status == "disabled" {
        return fmt.Errorf("cannot bind disabled device")
    }
    
    // 6. 验证绑定规则（不能同时绑定 room 和 bed）
    if device.BoundRoomID.Valid {
        return fmt.Errorf("device is already bound to room: %s", device.BoundRoomID.String)
    }
    
    // 7. 更新绑定
    updatedDevice := &domain.Device{
        BoundBedID:  sql.NullString{String: bedID, Valid: true},
        BoundRoomID: sql.NullString{Valid: false}, // 清除 room 绑定
    }
    
    if err := s.devicesRepo.UpdateDevice(ctx, tenantID, deviceID, updatedDevice); err != nil {
        return err
    }
    
    // 8. 发布 card 更新事件
    if err := s.publishCardUpdateEvent(ctx, deviceID); err != nil {
        s.logger.Warn("failed to publish card update event", zap.Error(err))
    }
    
    return nil
}

// UnbindDevice 解绑设备
func (s *DeviceService) UnbindDevice(
    ctx context.Context,
    tenantID, userID, userRole, deviceID string,
) error {
    // 1. 权限检查
    if !s.permissionChecker.CanBindDevices(ctx, tenantID, userID, userRole) {
        return ErrPermissionDenied
    }
    
    // 2. 更新绑定（清除 room 和 bed 绑定）
    updatedDevice := &domain.Device{
        BoundRoomID: sql.NullString{Valid: false},
        BoundBedID:  sql.NullString{Valid: false},
    }
    
    if err := s.devicesRepo.UpdateDevice(ctx, tenantID, deviceID, updatedDevice); err != nil {
        return err
    }
    
    // 3. 发布 card 更新事件
    if err := s.publishCardUpdateEvent(ctx, deviceID); err != nil {
        s.logger.Warn("failed to publish card update event", zap.Error(err))
    }
    
    return nil
}

// publishCardUpdateEvent 发布 card 更新事件
func (s *DeviceService) publishCardUpdateEvent(ctx context.Context, deviceID string) error {
    // 发布事件到消息队列，通知 card-aggregator 重新聚合
    event := &CardUpdateEvent{
        DeviceID: deviceID,
        EventType: "device_binding_changed",
        Timestamp: time.Now(),
    }
    
    return s.eventPublisher.Publish(ctx, "card-updates", event)
}
```

---

## 📋 总结

### DeviceService 职责

1. **权限检查**：验证用户是否有权限管理设备
2. **业务规则验证**：
   - 设备状态转换规则
   - 设备绑定规则（互斥、状态验证）
3. **数据转换**：前端格式 ↔ 领域模型
4. **业务编排**：
   - 设备绑定变更后发布 card 更新事件
   - 更新设备状态（如果需要）

### 设备绑定管理位置

**设备绑定管理应该放在 DeviceService 中**，而不是 UnitService 中，因为：
- 设备绑定是设备管理的核心功能
- 设备绑定涉及设备状态、业务访问权限等设备相关逻辑
- 设备绑定变更需要发布 card 更新事件，这是设备管理的职责
- UnitService 主要负责地址层级管理（Buildings, Units, Rooms, Beds），不涉及设备绑定

---

## 🚀 实现优先级

**Phase 2: 高优先级**（复杂度高）
- ✅ **DeviceService** - 设备状态管理、设备绑定管理、业务编排（card 更新事件）

