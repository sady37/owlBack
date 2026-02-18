package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"
	"wisefido-data/internal/config"
	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"
	"wisefido-qinglan/encode"

	"go.uber.org/zap"
)

// RadarInstall Radar 安装配置服务
// 雷达设备的查询与设置仅通过 qinglan_client（PUT/GET wisefido-qinglan /api/v1/radar/devices/{uid}/properties 等），无其他设备路径。
// 画布 layout 从 config_versions(config_type=room_layout, entity_id=room_id) 读写。
type RadarInstall struct {
	config            *config.Config
	devicesRepo       repository.DevicesRepository
	cardsRepo         repository.CardsRepository
	configVersionsRepo repository.ConfigVersionsRepository
	qinglanClient     *QinglanClient
	logger            *zap.Logger
	// 订阅管理器：记录哪些设备需要订阅（device_id -> bool）
	// 当设备被 bind 时，标记为需要订阅；unbind 时，移除订阅标记
	subscribedDevices map[string]bool
	subscribedMutex   sync.RWMutex // 保护 subscribedDevices 的并发访问
}

// NewRadarInstall 创建 Radar 安装配置服务
func NewRadarInstall(cfg *config.Config, devicesRepo repository.DevicesRepository, cardsRepo repository.CardsRepository, configVersionsRepo repository.ConfigVersionsRepository, qinglanClient *QinglanClient, logger *zap.Logger) *RadarInstall {
	return &RadarInstall{
		config:            cfg,
		devicesRepo:       devicesRepo,
		cardsRepo:         cardsRepo,
		configVersionsRepo: configVersionsRepo,
		qinglanClient:     qinglanClient,
		logger:            logger,
		subscribedDevices: make(map[string]bool),
	}
}

// ListCardDevices 按 card_id 取卡片上设备，供 GET /card/:cardId/devices；有 cardId 时前端用。
func (s *RadarInstall) ListCardDevices(ctx context.Context, tenantID, cardID string) ([]repository.CardDeviceItem, error) {
	if s.cardsRepo == nil {
		return nil, fmt.Errorf("cards repository not available")
	}
	return s.cardsRepo.GetCardDevices(ctx, tenantID, cardID)
}

// ListCardDevicesByDeviceID 通过 device_id 查找所属卡片，返回 card_id、room_id、该卡设备列表及 room 的 layout 配置（初始化时一次返回，供画布 Bind 与加载 layout）
// room_id 来自 devices.bound_room_id 或 beds.room_id；layout_config 来自 config_versions(config_type=room_layout, entity_id=room_id) 最新一条，无则 nil
func (s *RadarInstall) ListCardDevicesByDeviceID(ctx context.Context, tenantID, deviceID string) (cardID, roomID string, devices []repository.CardDeviceItem, layoutConfig json.RawMessage, err error) {
	if s.cardsRepo == nil {
		return "", "", nil, nil, fmt.Errorf("cards repository not available")
	}
	cardID, err = s.cardsRepo.GetCardIDByDeviceID(ctx, tenantID, deviceID)
	if err != nil {
		return "", "", nil, nil, err
	}
	devices, err = s.cardsRepo.GetCardDevices(ctx, tenantID, cardID)
	if err != nil {
		return "", "", nil, nil, err
	}
	roomID = ""
	if s.devicesRepo != nil {
		dev, e := s.devicesRepo.GetDevice(ctx, tenantID, deviceID)
		if e == nil && dev.RoomID.Valid && dev.RoomID.String != "" {
			roomID = dev.RoomID.String
		}
	}
	if roomID != "" && s.configVersionsRepo != nil {
		cfg, _ := s.GetLayoutConfigByRoomID(ctx, tenantID, roomID)
		layoutConfig = cfg
	}
	return cardID, roomID, devices, layoutConfig, nil
}

// GetLayoutConfigByRoomID 按 room_id 查询 config_versions 中 config_type='room_layout', entity_id=room_id 的最新生效配置，返回 config_data；无则返回 nil,nil
func (s *RadarInstall) GetLayoutConfigByRoomID(ctx context.Context, tenantID, roomID string) (json.RawMessage, error) {
	if s.configVersionsRepo == nil || roomID == "" {
		return nil, nil
	}
	cv, err := s.configVersionsRepo.GetConfigVersionAtTime(ctx, tenantID, "room_layout", roomID, time.Now())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return cv.ConfigData, nil
}

// SaveRoomLayout 将 config 写入 config_versions（config_type=room_layout, entity_id=room_id）。
// 比较后再保存：与当前最新 config_data 一致则不再插入；否则插入新版本（旧版本 valid_to 由 CreateConfigVersion 关闭）。
func (s *RadarInstall) SaveRoomLayout(ctx context.Context, tenantID, roomID string, configData []byte) error {
	if s.configVersionsRepo == nil || roomID == "" {
		return fmt.Errorf("config_versions repo not available or room_id empty")
	}
	if len(configData) == 0 {
		return fmt.Errorf("config_data is required")
	}

	cv, err := s.configVersionsRepo.GetConfigVersionAtTime(ctx, tenantID, "room_layout", roomID, time.Now())
	if err == nil && len(cv.ConfigData) > 0 {
		var a, b interface{}
		if errA := json.Unmarshal(configData, &a); errA == nil {
			if errB := json.Unmarshal(cv.ConfigData, &b); errB == nil && reflect.DeepEqual(a, b) {
				s.logger.Debug("SaveRoomLayout: config unchanged, skip insert", zap.String("room_id", roomID))
				return nil
			}
		}
	}

	cfg := &domain.ConfigVersion{
		ConfigType:       "room_layout",
		EntityID:         roomID,
		CurrentEntityID:  roomID,
		ConfigData:       configData,
		ValidFrom:        time.Now(),
	}
	_, err = s.configVersionsRepo.CreateConfigVersion(ctx, tenantID, cfg)
	return err
}

// GetDeviceUID 根据 device_id 和 tenant_id 获取设备 UID；设备不存在、未绑定或非雷达类型时返回明确错误
func (s *RadarInstall) GetDeviceUID(ctx context.Context, tenantID, deviceID string) (string, error) {
	if s.devicesRepo == nil {
		return "", fmt.Errorf("devices repository not available")
	}
	device, err := s.devicesRepo.GetDevice(ctx, tenantID, deviceID)
	if err != nil {
		return "", fmt.Errorf("设备不存在或未绑定(device_id=%s)，请先在画布中 Bind 真实雷达设备: %w", deviceID, err)
	}
	if !device.DeviceType.Valid || !strings.EqualFold(device.DeviceType.String, "radar") {
		return "", fmt.Errorf("该设备不是雷达类型，无法操作")
	}
	if device.DeviceUID == "" {
		return "", fmt.Errorf("device UID not found for device_id: %s", deviceID)
	}
	return device.DeviceUID, nil
}

// GetDeviceByUID 根据 device_uid 和 tenant_id 获取设备信息
// 用于从 device_uid 转换为 device_id（用于订阅检查等场景）
func (s *RadarInstall) GetDeviceByUID(ctx context.Context, tenantID, deviceUID string) (*domain.Device, error) {
	if s.devicesRepo == nil {
		return nil, fmt.Errorf("devices repository not available")
	}
	return s.devicesRepo.GetDeviceByUID(ctx, tenantID, deviceUID)
}

// GetDeviceStatus 获取设备在线状态
// 通过 wisefido-qinglan: GET /api/v1/radar/devices/{uid}/status
func (s *RadarInstall) GetDeviceStatus(ctx context.Context, deviceUID string) (string, error) {
	if s.qinglanClient == nil {
		return "", fmt.Errorf("qinglan client not available")
	}
	return s.qinglanClient.GetDeviceStatus(ctx, deviceUID)
}

// GetDeviceProperties 读取设备属性
// 通过 wisefido-qinglan: GET /api/v1/radar/devices/{uid}/properties
func (s *RadarInstall) GetDeviceProperties(ctx context.Context, uid string, keys []string) (map[string]interface{}, error) {
	return s.qinglanClient.GetDeviceProperties(ctx, uid, keys)
}

// SetDeviceProperties 设置设备属性，返回设备响应码（200=成功）和错误
// 通过 wisefido-qinglan: PUT /api/v1/radar/devices/{uid}/properties
func (s *RadarInstall) SetDeviceProperties(ctx context.Context, uid string, properties map[string]interface{}) (deviceCode int, err error) {
	return s.qinglanClient.SetDeviceProperties(ctx, uid, properties)
}

// SubscribeRealtimeData 订阅实时数据
// 通过 wisefido-qinglan: POST /api/v1/radar/devices/{uid}/subscribe
func (s *RadarInstall) SubscribeRealtimeData(ctx context.Context, uid string, content interface{}, duration int) error {
	return s.qinglanClient.SubscribeRealtimeData(ctx, uid, content, duration)
}

// CallDeviceFunction 调用设备功能（重启等）
// 通过 wisefido-qinglan: POST /api/v1/radar/devices/{uid}/function
func (s *RadarInstall) CallDeviceFunction(ctx context.Context, uid string, dev int) error {
	return s.qinglanClient.CallDeviceFunction(ctx, uid, dev)
}

// GetOriginalProperties 获取设备原始属性（v1.0 API 格式）
// keys 为空时读全部；有则按序只查这些 key，wisefido-qinglan 内分笔 MQTT、指令间 50ms。
// 返回 JSON 字符串，包含雷达配置参数
func (s *RadarInstall) GetOriginalProperties(ctx context.Context, tenantID, deviceID string, keys []string) (string, error) {
	// 1. 获取设备 UID
	uid, err := s.GetDeviceUID(ctx, tenantID, deviceID)
	if err != nil {
		return "", err
	}
	// 2. 读取属性：keys 为空读全部，否则按序查指定 key
	properties, err := s.GetDeviceProperties(ctx, uid, keys)
	if err != nil {
		return "", err
	}
	// 返回前对 ssid_password 脱敏：冒号后部分改为 *******（SSID:password 中的 password）
	if v, ok := properties["ssid_password"]; ok && v != nil {
		properties["ssid_password"] = encode.MaskSsidPassword(encode.ToStr(v))
	}

	// 3. 转换为 JSON 字符串
	jsonBytes, err := json.Marshal(properties)
	if err != nil {
		return "", fmt.Errorf("failed to marshal properties to JSON: %w", err)
	}

	return string(jsonBytes), nil
}

// GetOriginalPropertiesFromDB 从 DB 读取雷达安装配置（设备查不到时的回退）
// 当前无持久化表，返回空 JSON；后续可接 config_versions 等
func (s *RadarInstall) GetOriginalPropertiesFromDB(ctx context.Context, tenantID, deviceID string) (string, error) {
	return "{}", nil
}

// UpdateConfig 更新设备配置。
// config 与 radarMqttConfig 输出对齐：dm（画布 cm/10），install_model 统一 0/1/2，
// boundary_left/right/front/rear，可选 area_{i}_id/type/x1..y4；经 encode.EncodeV1ConfigToDeviceProps 转成
// radar_install_style、rectangle、declare_area 等后通过 qinglan 写入（安装/边界/区域可能触发重启）。
// 返回设备响应码（200=成功）和错误，供 HTTP 层透传 device_code 给前端。
func (s *RadarInstall) UpdateConfig(ctx context.Context, tenantID, deviceID string, config map[string]interface{}) (deviceCode int, err error) {
	// 1. 获取设备 UID
	uid, err := s.GetDeviceUID(ctx, tenantID, deviceID)
	if err != nil {
		return 0, err
	}

	// 2. 前端已完成格式转换（cm→dm、boundary→rectangle 等），此处透传
	properties := config
	encodeLogFields := []zap.Field{zap.String("device_id", deviceID), zap.Int("config_keys", len(config)), zap.Int("properties_keys", len(properties))}
	if v := properties["declare_area"]; v != nil {
		if s, ok := v.(string); ok {
			encodeLogFields = append(encodeLogFields, zap.String("declare_area", s))
		} else {
			encodeLogFields = append(encodeLogFields, zap.Any("declare_area", v))
		}
	}
	s.logger.Info("UpdateConfig: encode result", encodeLogFields...)
	if len(properties) == 0 {
		return 0, fmt.Errorf("config produced no device properties (check request body has install_model/height/rectangle/declare_area)")
	}

	// 便于 MQTT 排查：下发 rectangle 时打日志，可与 /prop/productId/uid/get 及设备 /post 回包对照
	if v, ok := properties["rectangle"]; ok {
		s.logger.Info("UpdateConfig: rectangle to device", zap.String("uid", uid), zap.Any("rectangle", v))
	}

	// 3. 设置设备属性，透传设备响应码给发起端
	return s.SetDeviceProperties(ctx, uid, properties)
}

// BindDevice 绑定设备（通知需要订阅该设备的数据）
// 当 vue-radar 画布中 bind 设备时调用
func (s *RadarInstall) BindDevice(ctx context.Context, tenantID, deviceID string) error {
	// 验证设备是否存在且为雷达类型
	device, err := s.devicesRepo.GetDevice(ctx, tenantID, deviceID)
	if err != nil {
		return fmt.Errorf("设备不存在: %w", err)
	}
	if !device.DeviceType.Valid || !strings.EqualFold(device.DeviceType.String, "radar") {
		return fmt.Errorf("该设备不是雷达类型，无法订阅")
	}
	
	// 标记为需要订阅（加锁保护）
	s.subscribedMutex.Lock()
	s.subscribedDevices[deviceID] = true
	s.subscribedMutex.Unlock()
	
	s.logger.Info("[RADAR_BIND] device bound, subscription enabled",
		zap.String("device_id", deviceID),
		zap.String("tenant_id", tenantID))
	return nil
}

// UnbindDevice 解绑设备（取消订阅该设备的数据）
// 当 vue-radar 画布中 unbind 设备时调用
func (s *RadarInstall) UnbindDevice(ctx context.Context, tenantID, deviceID string) error {
	// 移除订阅标记（加锁保护）
	s.subscribedMutex.Lock()
	delete(s.subscribedDevices, deviceID)
	s.subscribedMutex.Unlock()
	
	s.logger.Info("Device unbound, subscription disabled", zap.String("device_id", deviceID), zap.String("tenant_id", tenantID))
	return nil
}

// IsDeviceSubscribed 检查设备是否需要订阅
func (s *RadarInstall) IsDeviceSubscribed(deviceID string) bool {
	s.subscribedMutex.RLock()
	defer s.subscribedMutex.RUnlock()
	return s.subscribedDevices[deviceID]
}

// InitializeSubscriptionsFromLayout 从 layout 配置初始化订阅
// 解析 layout 中的 objects，找出所有已绑定的雷达设备，自动订阅
func (s *RadarInstall) InitializeSubscriptionsFromLayout(ctx context.Context, tenantID string, layoutConfig json.RawMessage) error {
	if len(layoutConfig) == 0 {
		return nil
	}
	
	var layout struct {
		Objects []struct {
			TypeName      string `json:"typeName"`
			BindedDeviceID string `json:"bindedDeviceId,omitempty"`
			DeviceID      string `json:"device_id,omitempty"`
		} `json:"objects"`
	}
	
	if err := json.Unmarshal(layoutConfig, &layout); err != nil {
		s.logger.Warn("Failed to parse layout config for subscription initialization", zap.Error(err))
		return nil // 不阻塞初始化流程
	}
	
	// 找出所有已绑定的雷达设备
	for _, obj := range layout.Objects {
		if obj.TypeName == "Radar" {
			deviceID := obj.BindedDeviceID
			if deviceID == "" {
				deviceID = obj.DeviceID
			}
			if deviceID != "" {
				// 自动订阅
				if err := s.BindDevice(ctx, tenantID, deviceID); err != nil {
					s.logger.Warn("Failed to subscribe device from layout", 
						zap.String("device_id", deviceID), 
						zap.Error(err))
					// 继续处理其他设备，不中断
				}
			}
		}
	}
	
	s.subscribedMutex.RLock()
	count := len(s.subscribedDevices)
	s.subscribedMutex.RUnlock()
	
	s.logger.Info("[RADAR_INIT] initialized subscriptions from layout", 
		zap.Int("subscribed_count", count))
	return nil
}
