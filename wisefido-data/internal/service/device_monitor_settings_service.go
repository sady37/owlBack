package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

// DeviceMonitorSettingsService 设备监控配置服务接口
type DeviceMonitorSettingsService interface {
	// 获取设备监控配置（根据设备类型返回 Sleepace 或 Radar 配置）
	GetDeviceMonitorSettings(ctx context.Context, req GetDeviceMonitorSettingsRequest) (*GetDeviceMonitorSettingsResponse, error)

	// 更新设备监控配置（根据设备类型更新 Sleepace 或 Radar 配置）
	UpdateDeviceMonitorSettings(ctx context.Context, req UpdateDeviceMonitorSettingsRequest) (*UpdateDeviceMonitorSettingsResponse, error)
}

// deviceMonitorSettingsService 实现
type deviceMonitorSettingsService struct {
	alarmDeviceRepo  repository.AlarmDeviceRepository
	devicesRepo      repository.DevicesRepository
	deviceStoreRepo  repository.DeviceStoreRepository
	logger           *zap.Logger
}

// NewDeviceMonitorSettingsService 创建设备监控配置服务实例
func NewDeviceMonitorSettingsService(
	alarmDeviceRepo repository.AlarmDeviceRepository,
	devicesRepo repository.DevicesRepository,
	deviceStoreRepo repository.DeviceStoreRepository,
	logger *zap.Logger,
) DeviceMonitorSettingsService {
	return &deviceMonitorSettingsService{
		alarmDeviceRepo: alarmDeviceRepo,
		devicesRepo:     devicesRepo,
		deviceStoreRepo: deviceStoreRepo,
		logger:          logger,
	}
}

// ============================================
// Request/Response DTOs
// ============================================

// GetDeviceMonitorSettingsRequest 获取设备监控配置请求
type GetDeviceMonitorSettingsRequest struct {
	TenantID  string // 租户ID
	DeviceID  string // 设备ID
	DeviceType string // 设备类型：'sleepace' 或 'radar'
}

// GetDeviceMonitorSettingsResponse 获取设备监控配置响应
type GetDeviceMonitorSettingsResponse struct {
	Settings map[string]interface{} `json:"settings"` // 配置项（flat 结构，与前端对齐）
}

// UpdateDeviceMonitorSettingsRequest 更新设备监控配置请求
type UpdateDeviceMonitorSettingsRequest struct {
	TenantID   string                 // 租户ID
	DeviceID   string                 // 设备ID
	DeviceType string                 // 设备类型：'sleepace' 或 'radar'
	Settings   map[string]interface{} // 配置项（flat 结构，来自前端）
}

// UpdateDeviceMonitorSettingsResponse 更新设备监控配置响应
type UpdateDeviceMonitorSettingsResponse struct {
	Success bool `json:"success"` // 是否成功
}

// ============================================
// Service 方法实现
// ============================================

// GetDeviceMonitorSettings 获取设备监控配置
func (s *deviceMonitorSettingsService) GetDeviceMonitorSettings(ctx context.Context, req GetDeviceMonitorSettingsRequest) (*GetDeviceMonitorSettingsResponse, error) {
	// 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.DeviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}
	if req.DeviceType != "sleepace" && req.DeviceType != "radar" {
		return nil, fmt.Errorf("invalid device_type: %s (must be 'sleepace' or 'radar')", req.DeviceType)
	}

	// 验证设备存在
	device, err := s.devicesRepo.GetDevice(ctx, req.TenantID, req.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	// 获取设备类型（通过 device_store_id）
	deviceType, err := s.getDeviceType(ctx, device)
	if err != nil {
		return nil, fmt.Errorf("failed to get device type: %w", err)
	}

	// 验证设备类型匹配
	expectedType := ""
	if req.DeviceType == "sleepace" {
		expectedType = "Sleepace"
	} else if req.DeviceType == "radar" {
		expectedType = "Radar"
	}
	if deviceType != expectedType {
		return nil, fmt.Errorf("device type mismatch: expected %s, got %s", expectedType, deviceType)
	}

	// 获取设备的监控配置
	alarmDevice, err := s.alarmDeviceRepo.GetAlarmDevice(ctx, req.TenantID, req.DeviceID)
	if err != nil {
		// 如果配置不存在，返回默认配置
		s.logger.Warn("Alarm device config not found, returning default settings",
			zap.String("tenant_id", req.TenantID),
			zap.String("device_id", req.DeviceID),
		)
		return s.getDefaultSettings(req.DeviceType), nil
	}

	// 解析 monitor_config JSONB
	var monitorConfig map[string]interface{}
	if len(alarmDevice.MonitorConfig) > 0 {
		if err := json.Unmarshal(alarmDevice.MonitorConfig, &monitorConfig); err != nil {
			s.logger.Warn("Failed to parse monitor_config, returning default settings",
				zap.String("tenant_id", req.TenantID),
				zap.String("device_id", req.DeviceID),
				zap.Error(err),
			)
			return s.getDefaultSettings(req.DeviceType), nil
		}
	} else {
		return s.getDefaultSettings(req.DeviceType), nil
	}

	// 根据设备类型转换为 flat 结构
	settings, err := s.convertMonitorConfigToFlat(req.DeviceType, monitorConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to convert monitor config: %w", err)
	}

	return &GetDeviceMonitorSettingsResponse{
		Settings: settings,
	}, nil
}

// UpdateDeviceMonitorSettings 更新设备监控配置
func (s *deviceMonitorSettingsService) UpdateDeviceMonitorSettings(ctx context.Context, req UpdateDeviceMonitorSettingsRequest) (*UpdateDeviceMonitorSettingsResponse, error) {
	// 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.DeviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}
	if req.DeviceType != "sleepace" && req.DeviceType != "radar" {
		return nil, fmt.Errorf("invalid device_type: %s (must be 'sleepace' or 'radar')", req.DeviceType)
	}
	if req.Settings == nil {
		return nil, fmt.Errorf("settings is required")
	}

	// 验证设备存在
	device, err := s.devicesRepo.GetDevice(ctx, req.TenantID, req.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	// 获取设备类型（通过 device_store_id）
	deviceType, err := s.getDeviceType(ctx, device)
	if err != nil {
		return nil, fmt.Errorf("failed to get device type: %w", err)
	}

	// 验证设备类型匹配
	expectedType := ""
	if req.DeviceType == "sleepace" {
		expectedType = "Sleepace"
	} else if req.DeviceType == "radar" {
		expectedType = "Radar"
	}
	if deviceType != expectedType {
		return nil, fmt.Errorf("device type mismatch: expected %s, got %s", expectedType, deviceType)
	}

	// 验证配置参数
	if err := s.validateSettings(req.DeviceType, req.Settings); err != nil {
		return nil, fmt.Errorf("invalid settings: %w", err)
	}

	// 将 flat 结构转换为 monitor_config JSONB
	monitorConfig, err := s.convertFlatToMonitorConfig(req.DeviceType, req.Settings)
	if err != nil {
		return nil, fmt.Errorf("failed to convert settings to monitor config: %w", err)
	}

	// 获取或创建 alarm_device 记录
	alarmDevice, err := s.alarmDeviceRepo.GetAlarmDevice(ctx, req.TenantID, req.DeviceID)
	if err != nil {
		// 如果不存在，创建新记录
		alarmDevice = &domain.AlarmDevice{
			DeviceID:      req.DeviceID,
			TenantID:      req.TenantID,
			MonitorConfig: monitorConfig,
		}
	} else {
		// 更新现有记录
		alarmDevice.MonitorConfig = monitorConfig
	}

	// 保存到数据库
	if err := s.alarmDeviceRepo.UpsertAlarmDevice(ctx, req.TenantID, req.DeviceID, alarmDevice); err != nil {
		s.logger.Error("Failed to upsert alarm device",
			zap.String("tenant_id", req.TenantID),
			zap.String("device_id", req.DeviceID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to update device monitor settings: %w", err)
	}

	s.logger.Info("Device monitor settings updated",
		zap.String("tenant_id", req.TenantID),
		zap.String("device_id", req.DeviceID),
		zap.String("device_type", req.DeviceType),
	)

	return &UpdateDeviceMonitorSettingsResponse{
		Success: true,
	}, nil
}

// ============================================
// 辅助方法
// ============================================

// getDefaultSettings 获取默认配置（根据设备类型）
func (s *deviceMonitorSettingsService) getDefaultSettings(deviceType string) *GetDeviceMonitorSettingsResponse {
	if deviceType == "sleepace" {
		return &GetDeviceMonitorSettingsResponse{
			Settings: map[string]interface{}{
				"left_bed_start_hour":        0,
				"left_bed_start_minute":      0,
				"left_bed_end_hour":          0,
				"left_bed_end_minute":        0,
				"left_bed_duration":          0,
				"left_bed_alarm_level":       "disabled",
				"min_heart_rate":             0,
				"heart_rate_slow_duration":   0,
				"heart_rate_slow_alarm_level": "disabled",
				"max_heart_rate":             0,
				"heart_rate_fast_duration":   0,
				"heart_rate_fast_alarm_level": "disabled",
				"min_breath_rate":            0,
				"breath_rate_slow_duration":  0,
				"breath_rate_slow_alarm_level": "disabled",
				"max_breath_rate":            0,
				"breath_rate_fast_duration":  0,
				"breath_rate_fast_alarm_level": "disabled",
				"breath_pause_duration":      0,
				"breath_pause_alarm_level":   "disabled",
				"body_move_duration":        0,
				"body_move_alarm_level":      "disabled",
				"nobody_move_duration":      0,
				"nobody_move_alarm_level":    "disabled",
				"no_turn_over_duration":     0,
				"no_turn_over_alarm_level":   "disabled",
				"situp_alarm_level":          "disabled",
				"onbed_duration":            0,
				"onbed_alarm_level":          "disabled",
				"fall_alarm_level":           "disabled",
			},
		}
	} else if deviceType == "radar" {
		return &GetDeviceMonitorSettingsResponse{
			Settings: map[string]interface{}{
				"radar_function_mode":             0,
				"suspected_fall_duration":         0,
				"fall_alarm_level":                "disabled",
				"posture_detection_alarm_level":   "disabled",
				"sitting_on_ground_duration":      0,
				"sitting_on_ground_alarm_level":   "disabled",
				"stay_detection_duration":         0,
				"stay_alarm_level":                "disabled",
				"leave_detection_start_hour":      0,
				"leave_detection_start_minute":    0,
				"leave_detection_end_hour":        0,
				"leave_detection_end_minute":     0,
				"leave_detection_duration":       0,
				"leave_alarm_level":              "disabled",
				"lower_heart_rate":               0,
				"heart_rate_slow_alarm_level":    "disabled",
				"upper_heart_rate":               0,
				"heart_rate_fast_alarm_level":    "disabled",
				"lower_breath_rate":              0,
				"breath_rate_slow_alarm_level":   "disabled",
				"upper_breath_rate":              0,
				"breath_rate_fast_alarm_level":   "disabled",
				"breath_pause_alarm_level":       "disabled",
				"weak_vital_duration":            0,
				"weak_vital_sensitivity":         0,
				"weak_vital_alarm_level":         "disabled",
				"inactivity_alarm_level":          "disabled",
			},
		}
	}

	return &GetDeviceMonitorSettingsResponse{
		Settings: map[string]interface{}{},
	}
}

// convertMonitorConfigToFlat 将 monitor_config JSONB 转换为 flat 结构（根据设备类型）
func (s *deviceMonitorSettingsService) convertMonitorConfigToFlat(deviceType string, monitorConfig map[string]interface{}) (map[string]interface{}, error) {
	settings := make(map[string]interface{})

	if deviceType == "sleepace" {
		// Sleepace 配置转换
		// 从 monitor_config 中提取配置项
		// 注意：monitor_config 的结构可能是嵌套的，需要转换为 flat 结构

		// 提取 alarms 配置
		if alarms, ok := monitorConfig["alarms"].(map[string]interface{}); ok {
			// 离床配置
			if leftBed, ok := alarms["SleepPad_LeftBed"].(map[string]interface{}); ok {
				if threshold, ok := leftBed["threshold"].(map[string]interface{}); ok {
					if startHour, ok := threshold["start_hour"].(float64); ok {
						settings["left_bed_start_hour"] = int(startHour)
					}
					if startMinute, ok := threshold["start_minute"].(float64); ok {
						settings["left_bed_start_minute"] = int(startMinute)
					}
					if endHour, ok := threshold["end_hour"].(float64); ok {
						settings["left_bed_end_hour"] = int(endHour)
					}
					if endMinute, ok := threshold["end_minute"].(float64); ok {
						settings["left_bed_end_minute"] = int(endMinute)
					}
					if duration, ok := threshold["duration"].(float64); ok {
						settings["left_bed_duration"] = int(duration)
					}
				}
				if level, ok := leftBed["level"].(string); ok {
					settings["left_bed_alarm_level"] = level
				} else {
					settings["left_bed_alarm_level"] = "disabled"
				}
			}

			// 心率配置
			if heartRate, ok := alarms["HeartRate"].(map[string]interface{}); ok {
				if threshold, ok := heartRate["threshold"].(map[string]interface{}); ok {
					if min, ok := threshold["min"].(float64); ok {
						settings["min_heart_rate"] = int(min)
					}
					if max, ok := threshold["max"].(float64); ok {
						settings["max_heart_rate"] = int(max)
					}
					if duration, ok := threshold["duration"].(float64); ok {
						settings["heart_rate_slow_duration"] = int(duration)
						settings["heart_rate_fast_duration"] = int(duration)
					}
				}
				if level, ok := heartRate["level"].(string); ok {
					settings["heart_rate_slow_alarm_level"] = level
					settings["heart_rate_fast_alarm_level"] = level
				} else {
					settings["heart_rate_slow_alarm_level"] = "disabled"
					settings["heart_rate_fast_alarm_level"] = "disabled"
				}
			}

			// 呼吸率配置
			if breathRate, ok := alarms["BreathRate"].(map[string]interface{}); ok {
				if threshold, ok := breathRate["threshold"].(map[string]interface{}); ok {
					if min, ok := threshold["min"].(float64); ok {
						settings["min_breath_rate"] = int(min)
					}
					if max, ok := threshold["max"].(float64); ok {
						settings["max_breath_rate"] = int(max)
					}
					if duration, ok := threshold["duration"].(float64); ok {
						settings["breath_rate_slow_duration"] = int(duration)
						settings["breath_rate_fast_duration"] = int(duration)
					}
				}
				if level, ok := breathRate["level"].(string); ok {
					settings["breath_rate_slow_alarm_level"] = level
					settings["breath_rate_fast_alarm_level"] = level
				} else {
					settings["breath_rate_slow_alarm_level"] = "disabled"
					settings["breath_rate_fast_alarm_level"] = "disabled"
				}
			}

			// 其他报警配置
			if breathPause, ok := alarms["BreathPause"].(map[string]interface{}); ok {
				if threshold, ok := breathPause["threshold"].(map[string]interface{}); ok {
					if duration, ok := threshold["duration"].(float64); ok {
						settings["breath_pause_duration"] = int(duration)
					}
				}
				if level, ok := breathPause["level"].(string); ok {
					settings["breath_pause_alarm_level"] = level
				} else {
					settings["breath_pause_alarm_level"] = "disabled"
				}
			}

			// 设置默认值（如果字段不存在）
			s.setSleepaceDefaults(settings)
		} else {
			// 如果没有 alarms 配置，返回默认值
			return s.getDefaultSettings("sleepace").Settings, nil
		}
	} else if deviceType == "radar" {
		// Radar 配置转换
		// 从 monitor_config 中提取配置项

		// 提取 alarms 配置
		if alarms, ok := monitorConfig["alarms"].(map[string]interface{}); ok {
			// 跌倒配置
			if fall, ok := alarms["Fall"].(map[string]interface{}); ok {
				if threshold, ok := fall["threshold"].(map[string]interface{}); ok {
					if duration, ok := threshold["duration"].(float64); ok {
						settings["suspected_fall_duration"] = int(duration)
					}
				}
				if level, ok := fall["level"].(string); ok {
					settings["fall_alarm_level"] = level
				} else {
					settings["fall_alarm_level"] = "disabled"
				}
			}

			// 其他配置...
			// 注意：Radar 的配置结构可能更复杂，需要根据实际数据结构调整

			// 设置默认值（如果字段不存在）
			s.setRadarDefaults(settings)
		} else {
			// 如果没有 alarms 配置，返回默认值
			return s.getDefaultSettings("radar").Settings, nil
		}
	}

	return settings, nil
}

// convertFlatToMonitorConfig 将 flat 结构转换为 monitor_config JSONB（根据设备类型）
func (s *deviceMonitorSettingsService) convertFlatToMonitorConfig(deviceType string, settings map[string]interface{}) ([]byte, error) {
	monitorConfig := make(map[string]interface{})
	alarms := make(map[string]interface{})

	if deviceType == "sleepace" {
		// Sleepace 配置转换
		// 离床配置
		if leftBedStartHour, ok := settings["left_bed_start_hour"]; ok {
			if leftBedStartMinute, ok := settings["left_bed_start_minute"]; ok {
				if leftBedEndHour, ok := settings["left_bed_end_hour"]; ok {
					if leftBedEndMinute, ok := settings["left_bed_end_minute"]; ok {
						if leftBedDuration, ok := settings["left_bed_duration"]; ok {
							leftBedLevel := "disabled"
							if level, ok := settings["left_bed_alarm_level"].(string); ok && level != "" {
								leftBedLevel = level
							}
							alarms["SleepPad_LeftBed"] = map[string]interface{}{
								"level": leftBedLevel,
								"enabled": leftBedLevel != "disabled",
								"threshold": map[string]interface{}{
									"start_hour":   leftBedStartHour,
									"start_minute": leftBedStartMinute,
									"end_hour":     leftBedEndHour,
									"end_minute":   leftBedEndMinute,
									"duration":     leftBedDuration,
								},
							}
						}
					}
				}
			}
		}

		// 心率配置
		if minHeartRate, ok := settings["min_heart_rate"]; ok {
			if maxHeartRate, ok := settings["max_heart_rate"]; ok {
				heartRateLevel := "disabled"
				if level, ok := settings["heart_rate_slow_alarm_level"].(string); ok && level != "" {
					heartRateLevel = level
				}
				duration := 0
				if d, ok := settings["heart_rate_slow_duration"]; ok {
					if dur, ok := d.(float64); ok {
						duration = int(dur)
					} else if dur, ok := d.(int); ok {
						duration = dur
					}
				}
				alarms["HeartRate"] = map[string]interface{}{
					"level":   heartRateLevel,
					"enabled": heartRateLevel != "disabled",
					"threshold": map[string]interface{}{
						"min":      minHeartRate,
						"max":      maxHeartRate,
						"duration": duration,
					},
				}
			}
		}

		// 呼吸率配置
		if minBreathRate, ok := settings["min_breath_rate"]; ok {
			if maxBreathRate, ok := settings["max_breath_rate"]; ok {
				breathRateLevel := "disabled"
				if level, ok := settings["breath_rate_slow_alarm_level"].(string); ok && level != "" {
					breathRateLevel = level
				}
				duration := 0
				if d, ok := settings["breath_rate_slow_duration"]; ok {
					if dur, ok := d.(float64); ok {
						duration = int(dur)
					} else if dur, ok := d.(int); ok {
						duration = dur
					}
				}
				alarms["BreathRate"] = map[string]interface{}{
					"level":   breathRateLevel,
					"enabled": breathRateLevel != "disabled",
					"threshold": map[string]interface{}{
						"min":      minBreathRate,
						"max":      maxBreathRate,
						"duration": duration,
					},
				}
			}
		}

		// 呼吸暂停配置
		if breathPauseDuration, ok := settings["breath_pause_duration"]; ok {
			breathPauseLevel := "disabled"
			if level, ok := settings["breath_pause_alarm_level"].(string); ok && level != "" {
				breathPauseLevel = level
			}
			alarms["BreathPause"] = map[string]interface{}{
				"level":   breathPauseLevel,
				"enabled": breathPauseLevel != "disabled",
				"threshold": map[string]interface{}{
					"duration": breathPauseDuration,
				},
			}
		}

		// 其他报警配置（简化处理，只设置 level）
		if level, ok := settings["body_move_alarm_level"].(string); ok && level != "" {
			alarms["BodyMove"] = map[string]interface{}{
				"level":   level,
				"enabled": level != "disabled",
			}
		}
		if level, ok := settings["nobody_move_alarm_level"].(string); ok && level != "" {
			alarms["NobodyMove"] = map[string]interface{}{
				"level":   level,
				"enabled": level != "disabled",
			}
		}
		if level, ok := settings["no_turn_over_alarm_level"].(string); ok && level != "" {
			alarms["NoTurnOver"] = map[string]interface{}{
				"level":   level,
				"enabled": level != "disabled",
			}
		}
		if level, ok := settings["situp_alarm_level"].(string); ok && level != "" {
			alarms["SitUp"] = map[string]interface{}{
				"level":   level,
				"enabled": level != "disabled",
			}
		}
		if level, ok := settings["onbed_alarm_level"].(string); ok && level != "" {
			alarms["OnBed"] = map[string]interface{}{
				"level":   level,
				"enabled": level != "disabled",
			}
		}
		if level, ok := settings["fall_alarm_level"].(string); ok && level != "" {
			alarms["Fall"] = map[string]interface{}{
				"level":   level,
				"enabled": level != "disabled",
			}
		}

		monitorConfig["alarms"] = alarms
	} else if deviceType == "radar" {
		// Radar 配置转换
		// 跌倒配置
		if fallDuration, ok := settings["suspected_fall_duration"]; ok {
			fallLevel := "disabled"
			if level, ok := settings["fall_alarm_level"].(string); ok && level != "" {
				fallLevel = level
			}
			alarms["Fall"] = map[string]interface{}{
				"level":   fallLevel,
				"enabled": fallLevel != "disabled",
				"threshold": map[string]interface{}{
					"duration": fallDuration,
				},
			}
		}

		// 其他 Radar 配置...
		// 注意：Radar 的配置结构可能更复杂，需要根据实际数据结构调整

		monitorConfig["alarms"] = alarms
	}

	// 序列化为 JSON
	configJSON, err := json.Marshal(monitorConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal monitor config: %w", err)
	}

	return configJSON, nil
}

// validateSettings 验证配置参数
func (s *deviceMonitorSettingsService) validateSettings(deviceType string, settings map[string]interface{}) error {
	// 验证报警级别
	validLevels := map[string]bool{
		"disabled": true,
		"EMERGENCY": true,
		"WARNING":   true,
		"ERROR":     true,
		"INFORMATIONAL": true,
	}

	for key, value := range settings {
		if strings.HasSuffix(key, "_alarm_level") {
			if level, ok := value.(string); ok {
				if !validLevels[level] {
					return fmt.Errorf("invalid alarm level '%s' for field '%s'", level, key)
				}
			}
		}
	}

	// 验证数值范围（如果需要）
	// 例如：心率范围、呼吸率范围等

	return nil
}

// setSleepaceDefaults 设置 Sleepace 默认值（如果字段不存在）
func (s *deviceMonitorSettingsService) setSleepaceDefaults(settings map[string]interface{}) {
	defaults := map[string]interface{}{
		"left_bed_start_hour":         0,
		"left_bed_start_minute":       0,
		"left_bed_end_hour":           0,
		"left_bed_end_minute":         0,
		"left_bed_duration":           0,
		"left_bed_alarm_level":        "disabled",
		"min_heart_rate":              0,
		"heart_rate_slow_duration":    0,
		"heart_rate_slow_alarm_level": "disabled",
		"max_heart_rate":              0,
		"heart_rate_fast_duration":    0,
		"heart_rate_fast_alarm_level": "disabled",
		"min_breath_rate":             0,
		"breath_rate_slow_duration":   0,
		"breath_rate_slow_alarm_level": "disabled",
		"max_breath_rate":             0,
		"breath_rate_fast_duration":   0,
		"breath_rate_fast_alarm_level": "disabled",
		"breath_pause_duration":       0,
		"breath_pause_alarm_level":    "disabled",
		"body_move_duration":          0,
		"body_move_alarm_level":      "disabled",
		"nobody_move_duration":       0,
		"nobody_move_alarm_level":    "disabled",
		"no_turn_over_duration":      0,
		"no_turn_over_alarm_level":   "disabled",
		"situp_alarm_level":          "disabled",
		"onbed_duration":            0,
		"onbed_alarm_level":          "disabled",
		"fall_alarm_level":           "disabled",
	}

	for key, defaultValue := range defaults {
		if _, exists := settings[key]; !exists {
			settings[key] = defaultValue
		}
	}
}

// getDeviceType 获取设备类型（通过 device_store_id）
func (s *deviceMonitorSettingsService) getDeviceType(ctx context.Context, device *domain.Device) (string, error) {
	if !device.DeviceStoreID.Valid {
		return "", fmt.Errorf("device has no device_store_id")
	}

	// 查询 device_store 获取设备类型
	deviceStore, err := s.deviceStoreRepo.GetDeviceStore(ctx, device.DeviceStoreID.String)
	if err != nil {
		return "", fmt.Errorf("failed to get device store: %w", err)
	}

	return deviceStore.DeviceType, nil
}

// setRadarDefaults 设置 Radar 默认值（如果字段不存在）
func (s *deviceMonitorSettingsService) setRadarDefaults(settings map[string]interface{}) {
	defaults := map[string]interface{}{
		"radar_function_mode":             0,
		"suspected_fall_duration":         0,
		"fall_alarm_level":                "disabled",
		"posture_detection_alarm_level":   "disabled",
		"sitting_on_ground_duration":      0,
		"sitting_on_ground_alarm_level":   "disabled",
		"stay_detection_duration":         0,
		"stay_alarm_level":                "disabled",
		"leave_detection_start_hour":      0,
		"leave_detection_start_minute":    0,
		"leave_detection_end_hour":        0,
		"leave_detection_end_minute":     0,
		"leave_detection_duration":       0,
		"leave_alarm_level":              "disabled",
		"lower_heart_rate":               0,
		"heart_rate_slow_alarm_level":    "disabled",
		"upper_heart_rate":               0,
		"heart_rate_fast_alarm_level":    "disabled",
		"lower_breath_rate":              0,
		"breath_rate_slow_alarm_level":   "disabled",
		"upper_breath_rate":              0,
		"breath_rate_fast_alarm_level":   "disabled",
		"breath_pause_alarm_level":       "disabled",
		"weak_vital_duration":            0,
		"weak_vital_sensitivity":         0,
		"weak_vital_alarm_level":         "disabled",
		"inactivity_alarm_level":          "disabled",
	}

	for key, defaultValue := range defaults {
		if _, exists := settings[key]; !exists {
			settings[key] = defaultValue
		}
	}
}

