package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
	"wisefido-data/internal/config"
	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

// RadarInstall Radar 安装配置服务
// 通过 wisefido-qinglan 与设备通信，提供雷达设备安装配置接口（v1.0 API 兼容）
// 画布 layout 可从 config_versions(config_type=room_layout, entity_id=room_id) 读取，LaySave 时回写
type RadarInstall struct {
	config            *config.Config
	devicesRepo       repository.DevicesRepository
	cardsRepo         repository.CardsRepository
	configVersionsRepo repository.ConfigVersionsRepository
	qinglanClient     *QinglanClient
	logger            *zap.Logger
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

// GetDeviceProperties 读取设备属性
// 通过 wisefido-qinglan: GET /api/v1/radar/devices/{uid}/properties
func (s *RadarInstall) GetDeviceProperties(ctx context.Context, uid string, keys []string) (map[string]interface{}, error) {
	return s.qinglanClient.GetDeviceProperties(ctx, uid, keys)
}

// SetDeviceProperties 设置设备属性
// 通过 wisefido-qinglan: PUT /api/v1/radar/devices/{uid}/properties
func (s *RadarInstall) SetDeviceProperties(ctx context.Context, uid string, properties map[string]interface{}) error {
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
// config 与 radarMqttConfig 输出对齐：dm（画布 cm/10），install_model 0/1/2 或 "ceiling"/"wall"/"corn"，
// boundary_left/right/front/rear，可选 area_{i}_id/type/x1..y4；经 V1ConfigToRadarDeviceProps 转成
// radar_install_style、rectangle、declare_area 等后通过 qinglan 写入（安装/边界/区域可能触发重启）。
func (s *RadarInstall) UpdateConfig(ctx context.Context, tenantID, deviceID string, config map[string]interface{}) error {
	// 1. 获取设备 UID
	uid, err := s.GetDeviceUID(ctx, tenantID, deviceID)
	if err != nil {
		return err
	}

	// 2. 转换配置格式（v1.0 格式 → Radar 设备属性格式）
	properties := V1ConfigToRadarDeviceProps(config)

	// 3. 设置设备属性
	return s.SetDeviceProperties(ctx, uid, properties)
}
