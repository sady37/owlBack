package service

import (
	"context"
	"encoding/json"
	"fmt"
	"wisefido-data/internal/config"
	"wisefido-data/internal/repository"
	
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// RadarService Radar 服务
// 调用 wisefido-radar 内部 API，提供统一的 Radar 设备管理接口
type RadarService struct {
	config       *config.Config
	devicesRepo  repository.DevicesRepository
	httpClient   *resty.Client
	logger       *zap.Logger
}

// NewRadarService 创建 Radar 服务
func NewRadarService(cfg *config.Config, devicesRepo repository.DevicesRepository, logger *zap.Logger) *RadarService {
	client := resty.New().
		SetBaseURL(cfg.Radar.InternalAPIBaseURL).
		SetTimeout(30).
		SetRetryCount(2).
		SetRetryWaitTime(1)
	
	return &RadarService{
		config:      cfg,
		devicesRepo: devicesRepo,
		httpClient:  client,
		logger:      logger,
	}
}

// GetDeviceUID 根据 device_id 和 tenant_id 获取设备 UID
func (s *RadarService) GetDeviceUID(ctx context.Context, tenantID, deviceID string) (string, error) {
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
// 调用 wisefido-radar 内部 API: POST /internal/api/v1/radar/devices/{uid}/properties/get
func (s *RadarService) GetDeviceProperties(ctx context.Context, uid string, keys []string) (map[string]interface{}, error) {
	var response struct {
		Code int                    `json:"code"`
		Msg  string                 `json:"msg"`
		Data map[string]interface{} `json:"data"`
	}
	
	reqBody := map[string]interface{}{
		"keys": keys,
	}
	
	resp, err := s.httpClient.R().
		SetContext(ctx).
		SetBody(reqBody).
		SetResult(&response).
		Post(fmt.Sprintf("/internal/api/v1/radar/devices/%s/properties/get", uid))
	
	if err != nil {
		s.logger.Error("Failed to call radar internal API for get properties",
			zap.String("uid", uid),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to call radar service: %w", err)
	}
	
	if resp.StatusCode() != 200 {
		s.logger.Error("Radar internal API returned error",
			zap.String("uid", uid),
			zap.Int("status_code", resp.StatusCode()),
			zap.String("response", string(resp.Body())),
		)
		return nil, fmt.Errorf("radar service returned status %d", resp.StatusCode())
	}
	
	if response.Code != 200 {
		return nil, fmt.Errorf("radar service error: %s (code: %d)", response.Msg, response.Code)
	}
	
	return response.Data, nil
}

// SetDeviceProperties 设置设备属性
// 调用 wisefido-radar 内部 API: POST /internal/api/v1/radar/devices/{uid}/properties/set
func (s *RadarService) SetDeviceProperties(ctx context.Context, uid string, properties map[string]interface{}) error {
	var response struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data interface{} `json:"data"`
	}
	
	reqBody := map[string]interface{}{
		"properties": properties,
	}
	
	resp, err := s.httpClient.R().
		SetContext(ctx).
		SetBody(reqBody).
		SetResult(&response).
		Post(fmt.Sprintf("/internal/api/v1/radar/devices/%s/properties/set", uid))
	
	if err != nil {
		s.logger.Error("Failed to call radar internal API for set properties",
			zap.String("uid", uid),
			zap.Error(err),
		)
		return fmt.Errorf("failed to call radar service: %w", err)
	}
	
	if resp.StatusCode() != 200 {
		s.logger.Error("Radar internal API returned error",
			zap.String("uid", uid),
			zap.Int("status_code", resp.StatusCode()),
			zap.String("response", string(resp.Body())),
		)
		return fmt.Errorf("radar service returned status %d", resp.StatusCode())
	}
	
	if response.Code != 200 {
		return fmt.Errorf("radar service error: %s (code: %d)", response.Msg, response.Code)
	}
	
	return nil
}

// SubscribeRealtimeData 订阅实时数据
// 调用 wisefido-radar 内部 API: POST /internal/api/v1/radar/devices/{uid}/realtime/subscribe
func (s *RadarService) SubscribeRealtimeData(ctx context.Context, uid string, content interface{}, duration int) error {
	var response struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data interface{} `json:"data"`
	}
	
	reqBody := map[string]interface{}{
		"content":  content,
		"duration": duration,
	}
	
	resp, err := s.httpClient.R().
		SetContext(ctx).
		SetBody(reqBody).
		SetResult(&response).
		Post(fmt.Sprintf("/internal/api/v1/radar/devices/%s/realtime/subscribe", uid))
	
	if err != nil {
		s.logger.Error("Failed to call radar internal API for subscribe realtime",
			zap.String("uid", uid),
			zap.Error(err),
		)
		return fmt.Errorf("failed to call radar service: %w", err)
	}
	
	if resp.StatusCode() != 200 {
		s.logger.Error("Radar internal API returned error",
			zap.String("uid", uid),
			zap.Int("status_code", resp.StatusCode()),
			zap.String("response", string(resp.Body())),
		)
		return fmt.Errorf("radar service returned status %d", resp.StatusCode())
	}
	
	if response.Code != 200 {
		return fmt.Errorf("radar service error: %s (code: %d)", response.Msg, response.Code)
	}
	
	return nil
}

// CallDeviceFunction 调用设备功能（重启等）
// 调用 wisefido-radar 内部 API: POST /internal/api/v1/radar/devices/{uid}/commands
func (s *RadarService) CallDeviceFunction(ctx context.Context, uid string, dev int) error {
	var response struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data interface{} `json:"data"`
	}
	
	reqBody := map[string]interface{}{
		"dev": dev,
	}
	
	resp, err := s.httpClient.R().
		SetContext(ctx).
		SetBody(reqBody).
		SetResult(&response).
		Post(fmt.Sprintf("/internal/api/v1/radar/devices/%s/commands", uid))
	
	if err != nil {
		s.logger.Error("Failed to call radar internal API for device function",
			zap.String("uid", uid),
			zap.Int("dev", dev),
			zap.Error(err),
		)
		return fmt.Errorf("failed to call radar service: %w", err)
	}
	
	if resp.StatusCode() != 200 {
		s.logger.Error("Radar internal API returned error",
			zap.String("uid", uid),
			zap.Int("status_code", resp.StatusCode()),
			zap.String("response", string(resp.Body())),
		)
		return fmt.Errorf("radar service returned status %d", resp.StatusCode())
	}
	
	if response.Code != 200 {
		return fmt.Errorf("radar service error: %s (code: %d)", response.Msg, response.Code)
	}
	
	return nil
}

// GetOriginalProperties 获取设备原始属性（v1.0 API 格式）
// 返回 JSON 字符串，包含雷达所有配置参数
func (s *RadarService) GetOriginalProperties(ctx context.Context, tenantID, deviceID string) (string, error) {
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
// 将前端配置转换为 Radar 设备属性并设置
func (s *RadarService) UpdateConfig(ctx context.Context, tenantID, deviceID string, config map[string]interface{}) error {
	// 1. 获取设备 UID
	uid, err := s.GetDeviceUID(ctx, tenantID, deviceID)
	if err != nil {
		return err
	}
	
	// 2. 转换配置格式（v1.0 格式 → Radar 设备属性格式）
	properties := s.convertConfigToProperties(config)
	
	// 3. 设置设备属性
	return s.SetDeviceProperties(ctx, uid, properties)
}

// convertConfigToProperties 将 v1.0 配置格式转换为 Radar 设备属性格式
// v1.0 格式：install_model, height, boundary_left, boundary_right, boundary_front, boundary_rear, area_*_*
// Radar 属性格式：radar_func_ctrl, radar_install_style, radar_install_height, rectangle, declare_area
func (s *RadarService) convertConfigToProperties(config map[string]interface{}) map[string]interface{} {
	properties := make(map[string]interface{})
	
	// 安装方式：install_model (wall/ceiling) → radar_install_style (0=顶装, 1=侧装)
	if installModel, ok := config["install_model"].(string); ok {
		if installModel == "ceiling" {
			properties["radar_install_style"] = "0"
		} else if installModel == "wall" {
			properties["radar_install_style"] = "1"
		}
	}
	
	// 安装高度：height (dm) → radar_install_height
	if height, ok := config["height"].(float64); ok {
		properties["radar_install_height"] = fmt.Sprintf("%.0f", height)
	}
	
	// 检测边界：boundary_left, boundary_right, boundary_front, boundary_rear → rectangle
	// 格式：{x1, y1; x2, y2, x3, y3, x4, y4}
	if boundaryLeft, ok := config["boundary_left"].(float64); ok {
		if boundaryRight, ok := config["boundary_right"].(float64); ok {
			if boundaryFront, ok := config["boundary_front"].(float64); ok {
				if boundaryRear, ok := config["boundary_rear"].(float64); ok {
					// 根据安装方式构建边界坐标
					installModel := "ceiling"
					if im, ok := config["install_model"].(string); ok {
						installModel = im
					}
					
					rectangle := s.buildRectangle(boundaryLeft, boundaryRight, boundaryFront, boundaryRear, installModel)
					properties["rectangle"] = rectangle
				}
			}
		}
	}
	
	// 区域配置：area_*_* → declare_area
	// TODO: 实现区域配置转换（需要根据实际的前端格式）
	
	return properties
}

// buildRectangle 构建边界矩形坐标字符串
// 格式：{x1, y1; x2, y2, x3, y3, x4, y4}
func (s *RadarService) buildRectangle(left, right, front, rear float64, installModel string) string {
	// 顶装：X 轴 ±left/right，Y 轴 ±front/rear
	// 侧装：X 轴 ±left/right，Y 轴 0 到 front/rear
	if installModel == "ceiling" {
		return fmt.Sprintf("{%d, %d; %d, %d; %d, %d; %d, %d}",
			int(-left), int(-front),
			int(right), int(-front),
			int(-left), int(rear),
			int(right), int(rear))
	} else {
		// 侧装
		return fmt.Sprintf("{%d, %d; %d, %d; %d, %d; %d, %d}",
			int(-left), 0,
			int(right), 0,
			int(-left), int(rear),
			int(right), int(rear))
	}
}

