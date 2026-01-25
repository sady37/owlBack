package service

import (
	"context"
	"encoding/json"
	"fmt"
	"wisefido-data/internal/config"
	"wisefido-data/internal/repository"
	
	"go.uber.org/zap"
)

// RadarInstall Radar 安装配置服务
// 通过 wisefido-qinglan 与设备通信，提供雷达设备安装配置接口（v1.0 API 兼容）
// 主要用于处理安装方式、安装高度、检测边界等安装相关配置
type RadarInstall struct {
	config       *config.Config
	devicesRepo  repository.DevicesRepository
	qinglanClient *QinglanClient
	logger       *zap.Logger
}

// NewRadarInstall 创建 Radar 安装配置服务
func NewRadarInstall(cfg *config.Config, devicesRepo repository.DevicesRepository, qinglanClient *QinglanClient, logger *zap.Logger) *RadarInstall {
	return &RadarInstall{
		config:        cfg,
		devicesRepo:   devicesRepo,
		qinglanClient: qinglanClient,
		logger:        logger,
	}
}

// GetDeviceUID 根据 device_id 和 tenant_id 获取设备 UID
func (s *RadarInstall) GetDeviceUID(ctx context.Context, tenantID, deviceID string) (string, error) {
	if s.devicesRepo == nil {
		return "", fmt.Errorf("devices repository not available")
	}
	
	device, err := s.devicesRepo.GetDevice(ctx, tenantID, deviceID)
	if err != nil {
		return "", fmt.Errorf("failed to get device: %w", err)
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
// 返回 JSON 字符串，包含雷达所有配置参数
func (s *RadarInstall) GetOriginalProperties(ctx context.Context, tenantID, deviceID string) (string, error) {
	// 1. 获取设备 UID
	uid, err := s.GetDeviceUID(ctx, tenantID, deviceID)
	if err != nil {
		return "", err
	}
	
	// 2. 读取所有属性（keys 为空数组表示读取所有属性）
	properties, err := s.GetDeviceProperties(ctx, uid, []string{})
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

// UpdateConfig 更新设备配置（v1.0 API 格式）
// 将前端配置转换为 Radar 设备属性并设置（安装方式/高度/边界，可能触发重启）
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

