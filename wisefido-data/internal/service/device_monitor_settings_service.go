package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/notifier"
	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

// DeviceMonitorSettingsService 设备监控配置服务接口
type DeviceMonitorSettingsService interface {
	// 获取设备监控配置（根据设备类型返回 Sleepace 或 Radar 配置）
	GetDeviceMonitorSettings(ctx context.Context, req GetDeviceMonitorSettingsRequest) (*GetDeviceMonitorSettingsResponse, error)

	// 更新设备监控配置（根据设备类型更新 Sleepace 或 Radar 配置）
	UpdateDeviceMonitorSettings(ctx context.Context, req UpdateDeviceMonitorSettingsRequest) (*UpdateDeviceMonitorSettingsResponse, error)

	// 获取默认配置（硬编码阈值，Alarm Level 优先从当前租户的 alarm_cloud 读取）
	GetDefaultDeviceMonitorSettings(ctx context.Context, tenantID string, deviceType string) (*GetDeviceMonitorSettingsResponse, error)
}

// deviceMonitorSettingsService 实现
type deviceMonitorSettingsService struct {
	alarmDeviceRepo repository.AlarmDeviceRepository
	alarmCloudRepo  repository.AlarmCloudRepository
	devicesRepo     repository.DevicesRepository
	deviceStoreRepo repository.DeviceStoreRepository
	sleepaceClient  *SleepaceClient // Sleepace 硬件 API 客户端（可选）
	configNotifier  *notifier.ConfigNotifier // 配置变更通知器
	logger          *zap.Logger
}

// NewDeviceMonitorSettingsService 创建设备监控配置服务实例
func NewDeviceMonitorSettingsService(
	alarmDeviceRepo repository.AlarmDeviceRepository,
	alarmCloudRepo repository.AlarmCloudRepository,
	devicesRepo repository.DevicesRepository,
	deviceStoreRepo repository.DeviceStoreRepository,
	configNotifier *notifier.ConfigNotifier,
	logger *zap.Logger,
) DeviceMonitorSettingsService {
	return &deviceMonitorSettingsService{
		alarmDeviceRepo: alarmDeviceRepo,
		alarmCloudRepo:  alarmCloudRepo,
		devicesRepo:     devicesRepo,
		deviceStoreRepo: deviceStoreRepo,
		sleepaceClient:  nil, // 通过 SetSleepaceClient 延迟设置
		configNotifier:  configNotifier,
		logger:          logger,
	}
}

// SetSleepaceClient 设置 Sleepace 客户端（延迟初始化，避免循环依赖）
func (s *deviceMonitorSettingsService) SetSleepaceClient(client *SleepaceClient) {
	s.sleepaceClient = client
}

// ============================================
// Request/Response DTOs
// ============================================

// GetDeviceMonitorSettingsRequest 获取设备监控配置请求
type GetDeviceMonitorSettingsRequest struct {
	TenantID   string // 租户ID
	DeviceID   string // 设备ID
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

	// 验证设备类型匹配（不区分大小写）
	expectedType := ""
	if req.DeviceType == "sleepace" {
		expectedType = "sleepad" // 数据库中的设备类型是 "sleepad"（小写）
	} else if req.DeviceType == "radar" {
		expectedType = "radar" // 数据库中的设备类型可能是 "radar"（小写）
	}
	if !strings.EqualFold(deviceType, expectedType) {
		return nil, fmt.Errorf("device type mismatch: expected %s, got %s", expectedType, deviceType)
	}

	// 对于 Sleepace 设备，优先从硬件读取配置（参考 v1.0 实现）
	// 如果硬件读取失败（设备离线或网络问题），则从数据库读取配置
	// 注意：仅当 Device Type = Sleepad 且 Device Model = BM8701-2 时才使用 Sleepace 硬件配置
	// TODO: 暂时注释掉硬件读取功能，后续需要时再启用
	/*
		if req.DeviceType == "sleepace" && s.sleepaceClient != nil {
			// 检查设备型号是否为 BM8701-2（Sleepace 厂家）
			deviceType := ""
			deviceModel := ""
			if device.DeviceType.Valid {
				deviceType = device.DeviceType.String
			}
			if device.DeviceModel.Valid {
				deviceModel = device.DeviceModel.String
			}
			// 仅当 Device Type = Sleepad 且 Device Model = BM8701-2 时才从硬件读取
			if strings.EqualFold(deviceType, "sleepad") && deviceModel == "BM8701-2" {
				hardwareSettings, err := s.getSleepaceSettingsFromHardware(ctx, device, req.DeviceID)
				if err != nil {
					s.logger.Warn("Failed to get settings from hardware, falling back to database",
						zap.String("tenant_id", req.TenantID),
						zap.String("device_id", req.DeviceID),
						zap.String("device_status", device.Status),
						zap.Error(err),
					)
					// 如果硬件读取失败，继续从数据库读取
				} else {
					// 成功从硬件读取，直接返回
					return hardwareSettings, nil
				}
			} else {
				// 设备型号不是 BM8701-2，跳过硬件读取，从数据库读取
				s.logger.Info("Device model is not BM8701-2, skipping hardware read",
					zap.String("device_id", req.DeviceID),
					zap.String("device_type", deviceType),
					zap.String("device_model", deviceModel),
				)
			}
		}
	*/

	// 获取设备的监控配置（从数据库）
	alarmDevice, err := s.alarmDeviceRepo.GetAlarmDevice(ctx, req.TenantID, req.DeviceID)
	if err != nil {
		// 如果配置不存在，返回空配置（不使用 System 租户模板或硬编码默认值）
		return &GetDeviceMonitorSettingsResponse{
			Settings: make(map[string]interface{}),
		}, nil
	}

	// 解析 monitor_config JSONB
	var monitorConfig map[string]interface{}
	if len(alarmDevice.MonitorConfig) > 0 {
		if err := json.Unmarshal(alarmDevice.MonitorConfig, &monitorConfig); err != nil {
			s.logger.Warn("Failed to parse monitor_config, returning empty settings",
				zap.String("tenant_id", req.TenantID),
				zap.String("device_id", req.DeviceID),
				zap.Error(err),
			)
			// 解析失败，返回空配置
			return &GetDeviceMonitorSettingsResponse{
				Settings: make(map[string]interface{}),
			}, nil
		}
	} else {
		// monitor_config 为空，返回空配置
		return &GetDeviceMonitorSettingsResponse{
			Settings: make(map[string]interface{}),
		}, nil
	}

	// 根据设备类型转换为 flat 结构
	settings, err := s.convertMonitorConfigToFlat(ctx, req.DeviceType, monitorConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to convert monitor config: %w", err)
	}

	// 从当前租户的 alarm_cloud 获取报警级别，填充缺失的值（不回退到 System 租户）
	// 注意：Load 时不应该回退到 System 租户，Alarm Level 默认值应该由前端初始化（'disabled'）
	if err := s.getAlarmLevelFromCloud(ctx, req.TenantID, req.DeviceType, settings); err != nil {
		s.logger.Warn("Failed to fill default alarm levels from cloud, using existing values",
			zap.String("tenant_id", req.TenantID),
			zap.String("device_type", req.DeviceType),
			zap.Error(err),
		)
	}

	// 应用设备默认值（雷达特定默认值）
	s.applyDeviceDefaults(req.DeviceType, settings)

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

	// 验证设备类型匹配（不区分大小写）
	expectedType := ""
	if req.DeviceType == "sleepace" {
		expectedType = "sleepad" // 数据库中的设备类型是 "sleepad"（小写）
	} else if req.DeviceType == "radar" {
		expectedType = "radar" // 数据库中的设备类型可能是 "radar"（小写）
	}
	if !strings.EqualFold(deviceType, expectedType) {
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

	// 对于 Sleepace 设备，同步到硬件（参考 v1.0 实现）
	// 注意：仅当 Device Type = Sleepad 且 Device Model = BM8701-2 时才写入硬件
	// Save 顺序：先写 DB（已完成），然后写硬件（如果失败，不影响 DB 保存）
	if req.DeviceType == "sleepace" && s.sleepaceClient != nil {
		// 检查设备型号是否为 BM8701-2（Sleepace 厂家）
		deviceType := ""
		deviceModel := ""
		if device.DeviceType.Valid {
			deviceType = device.DeviceType.String
		}
		if device.DeviceModel.Valid {
			deviceModel = device.DeviceModel.String
		}
		// 仅当 Device Type = Sleepad 且 Device Model = BM8701-2 时才写入硬件
		if strings.EqualFold(deviceType, "sleepad") && deviceModel == "BM8701-2" {
			if err := s.updateSleepaceSettingsToHardware(ctx, device, req.DeviceID, req.Settings); err != nil {
				s.logger.Warn("Failed to sync settings to hardware, but database update succeeded",
					zap.String("tenant_id", req.TenantID),
					zap.String("device_id", req.DeviceID),
					zap.Error(err),
				)
				// 硬件同步失败不影响数据库保存，只记录警告
			} else {
				s.logger.Info("Settings synced to hardware successfully",
					zap.String("tenant_id", req.TenantID),
					zap.String("device_id", req.DeviceID),
				)
			}
		} else {
			// 设备型号不是 BM8701-2，跳过硬件写入
			s.logger.Info("Device model is not BM8701-2, skipping hardware write",
				zap.String("device_id", req.DeviceID),
				zap.String("device_type", deviceType),
				zap.String("device_model", deviceModel),
			)
		}
	}

	s.logger.Info("Device monitor settings updated",
		zap.String("tenant_id", req.TenantID),
		zap.String("device_id", req.DeviceID),
		zap.String("device_type", req.DeviceType),
	)

	// 发布配置变更通知
	if s.configNotifier != nil {
		if err := s.configNotifier.NotifyAlarmDeviceUpdated(ctx, req.TenantID, req.DeviceID); err != nil {
			s.logger.Warn("Failed to notify config change, but database update succeeded",
				zap.String("tenant_id", req.TenantID),
				zap.String("device_id", req.DeviceID),
				zap.Error(err),
			)
			// 通知失败不影响数据库保存，只记录警告
		}
	}

	return &UpdateDeviceMonitorSettingsResponse{
		Success: true,
	}, nil
}

// ============================================
// 辅助方法
// ============================================

// getDefaultSettingsHardValue 获取硬编码的默认配置（根据设备类型）
// 硬编码阈值和 Alarm Level 默认值
// 阈值：硬编码（与 System 租户模板设备的值相同）
// Alarm Level：硬编码（与 AlarmCloud.vue 中的 DEFAULT_ALARM_LEVELS 相同）
func (s *deviceMonitorSettingsService) getDefaultSettingsHardValue(ctx context.Context, deviceType string) map[string]interface{} {
	settings := make(map[string]interface{})

	if deviceType == "sleepace" {
		// Sleepace 硬编码阈值（与 System 租户模板设备 Sleepad001 的值相同）
		settings["left_bed_start_hour"] = 21
		settings["left_bed_start_minute"] = 30
		settings["left_bed_end_hour"] = 7
		settings["left_bed_end_minute"] = 30
		settings["left_bed_duration"] = 2700 // 45分钟 = 2700秒
		settings["min_heart_rate"] = 45
		settings["max_heart_rate"] = 120
		settings["heart_rate_slow_duration"] = 600 // 10分钟 = 600秒
		settings["heart_rate_fast_duration"] = 600 // 10分钟 = 600秒
		settings["min_breath_rate"] = 10
		settings["max_breath_rate"] = 24
		settings["breath_rate_slow_duration"] = 300 // 5分钟 = 300秒
		settings["breath_rate_fast_duration"] = 120 // 2分钟 = 120秒
		settings["breath_pause_duration"] = 30
		settings["body_move_duration"] = 10     // 分钟
		settings["nobody_move_duration"] = 45   // 分钟
		settings["no_turn_over_duration"] = 120 // 分钟
		settings["onbed_duration"] = 300        // 5分钟 = 300秒

		// Sleepace 硬编码 Alarm Level 默认值（与 AlarmCloud.vue 中的 DEFAULT_ALARM_LEVELS.SleepPad 相同）
		// SleepPad_ApneaHypopnea: EMERGENCY -> "0"
		settings["breath_pause_alarm_level"] = "0"
		// SleepPad_AbnormalHeartRate: EMERGENCY -> "0"
		settings["heart_rate_slow_alarm_level"] = "0"
		settings["heart_rate_fast_alarm_level"] = "0"
		// SleepPad_AbnormalRespiratoryRate: EMERGENCY -> "0"
		settings["breath_rate_slow_alarm_level"] = "0"
		settings["breath_rate_fast_alarm_level"] = "0"
		// SleepPad_AbnormalBodyMovement: WARNING -> "4"
		settings["body_move_alarm_level"] = "4"
		settings["nobody_move_alarm_level"] = "4"
		settings["no_turn_over_alarm_level"] = "4"
		// SleepPad_LeftBed: WARNING -> "4"
		settings["left_bed_alarm_level"] = "4"
		// SleepPad_SitUp: WARNING -> "4"
		settings["situp_alarm_level"] = "4"
		// SleepPad_InBed: DISABLE -> "disabled"
		settings["onbed_alarm_level"] = "disabled"
		// Sensor fall: Enable by default (fallFlag = true, boolean)
		settings["fallFlag"] = true

	} else if deviceType == "radar" {
		// Radar 硬编码阈值（与 System 租户模板设备 Radar001 的值相同）
		settings["radar_function_mode"] = 15 // Full Function
		settings["leave_detection_start_hour"] = 21
		settings["leave_detection_start_minute"] = 30
		settings["leave_detection_end_hour"] = 7
		settings["leave_detection_end_minute"] = 30
		settings["leave_detection_duration"] = 45 // 分钟
		settings["lower_heart_rate"] = 50
		settings["upper_heart_rate"] = 100
		settings["lower_breath_rate"] = 8
		settings["upper_breath_rate"] = 26
		settings["suspected_fall_duration"] = 30    // 秒（前端发送秒，DB 存储秒）
		settings["sitting_on_ground_duration"] = 90 // 秒
		settings["stay_detection_duration"] = 45    // 分钟
		settings["weak_vital_duration"] = 5         // 分钟
		settings["weak_vital_sensitivity"] = 35

		// Radar 硬编码 Alarm Level 默认值（与 AlarmCloud.vue 中的 DEFAULT_ALARM_LEVELS.Radar 相同）
		// Radar_ApneaHypopnea: DISABLE -> "disabled"
		settings["breath_pause_alarm_level"] = "disabled"
		// Radar_AbnormalHeartRate: EMERGENCY -> "0"
		settings["heart_rate_slow_alarm_level"] = "0"
		settings["heart_rate_fast_alarm_level"] = "0"
		// Radar_AbnormalRespiratoryRate: EMERGENCY -> "0"
		settings["breath_rate_slow_alarm_level"] = "0"
		settings["breath_rate_fast_alarm_level"] = "0"
		// Fall: EMERGENCY -> "0" (优先使用 Fall，因为它是 EMERGENCY)
		// 注意：SuspectedFall 和 Fall 都映射到 fall_alarm_level，但 Fall 的优先级更高（EMERGENCY）
		settings["fall_alarm_level"] = "0"
		// VitalsWeak: EMERGENCY -> "0"
		settings["weak_vital_alarm_level"] = "0"
		// Radar_LeftBed: WARNING -> "4"
		settings["leave_alarm_level"] = "4"
		// Stay: WARNING -> "4"
		settings["stay_alarm_level"] = "4"
		// NoActivity24h: WARNING -> "4"
		settings["inactivity_alarm_level"] = "4"
		// Posture detection: Enable by default (posture_detection_enabled = true)
		settings["posture_detection_enabled"] = true
		// SittingOnGround 使用 AngleException 的值（从数据库模板设备和 AlarmCloud.vue 看，SittingOnGround 使用 AngleException 的配置）
		settings["sitting_on_ground_alarm_level"] = "4"
	}

	return settings
}

// convertMonitorConfigToFlat 将 monitor_config JSONB 转换为 flat 结构（根据设备类型）
func (s *deviceMonitorSettingsService) convertMonitorConfigToFlat(ctx context.Context, deviceType string, monitorConfig map[string]interface{}) (map[string]interface{}, error) {
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
					settings["left_bed_alarm_level"] = s.convertAlarmLevelToFrontendFormat(level)
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
					// 支持 slow_duration 和 fast_duration，如果不存在则使用 duration
					if slowDuration, ok := threshold["slow_duration"].(float64); ok {
						settings["heart_rate_slow_duration"] = int(slowDuration)
					} else if duration, ok := threshold["duration"].(float64); ok {
						settings["heart_rate_slow_duration"] = int(duration)
					}
					if fastDuration, ok := threshold["fast_duration"].(float64); ok {
						settings["heart_rate_fast_duration"] = int(fastDuration)
					} else if duration, ok := threshold["duration"].(float64); ok {
						settings["heart_rate_fast_duration"] = int(duration)
					}
				}
				if level, ok := heartRate["level"].(string); ok {
					settings["heart_rate_slow_alarm_level"] = s.convertAlarmLevelToFrontendFormat(level)
					settings["heart_rate_fast_alarm_level"] = s.convertAlarmLevelToFrontendFormat(level)
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
					// 支持 slow_duration 和 fast_duration，如果不存在则使用 duration
					if slowDuration, ok := threshold["slow_duration"].(float64); ok {
						settings["breath_rate_slow_duration"] = int(slowDuration)
					} else if duration, ok := threshold["duration"].(float64); ok {
						settings["breath_rate_slow_duration"] = int(duration)
					}
					if fastDuration, ok := threshold["fast_duration"].(float64); ok {
						settings["breath_rate_fast_duration"] = int(fastDuration)
					} else if duration, ok := threshold["duration"].(float64); ok {
						settings["breath_rate_fast_duration"] = int(duration)
					}
				}
				if level, ok := breathRate["level"].(string); ok {
					settings["breath_rate_slow_alarm_level"] = s.convertAlarmLevelToFrontendFormat(level)
					settings["breath_rate_fast_alarm_level"] = s.convertAlarmLevelToFrontendFormat(level)
				} else {
					settings["breath_rate_slow_alarm_level"] = "disabled"
					settings["breath_rate_fast_alarm_level"] = "disabled"
				}
			}

			// 其他报警配置
			if breathPause, ok := alarms["BreathPause"].(map[string]interface{}); ok {
				if threshold, ok := breathPause["threshold"].(map[string]interface{}); ok {
					if duration, ok := threshold["duration"].(float64); ok {
						settings["breath_pause_duration"] = int(duration) // 秒
					}
				}
				if level, ok := breathPause["level"].(string); ok {
					settings["breath_pause_alarm_level"] = s.convertAlarmLevelToFrontendFormat(level)
				} else {
					settings["breath_pause_alarm_level"] = "disabled"
				}
			}

			// BodyMove 配置（duration 单位：分钟）
			if bodyMove, ok := alarms["BodyMove"].(map[string]interface{}); ok {
				if threshold, ok := bodyMove["threshold"].(map[string]interface{}); ok {
					if duration, ok := threshold["duration"].(float64); ok {
						settings["body_move_duration"] = int(duration) // 分钟
					}
				}
				if level, ok := bodyMove["level"].(string); ok {
					settings["body_move_alarm_level"] = s.convertAlarmLevelToFrontendFormat(level)
				} else {
					settings["body_move_alarm_level"] = "disabled"
				}
			}

			// NobodyMove 配置（duration 单位：分钟）
			if nobodyMove, ok := alarms["NobodyMove"].(map[string]interface{}); ok {
				if threshold, ok := nobodyMove["threshold"].(map[string]interface{}); ok {
					if duration, ok := threshold["duration"].(float64); ok {
						settings["nobody_move_duration"] = int(duration) // 分钟
					}
				}
				if level, ok := nobodyMove["level"].(string); ok {
					settings["nobody_move_alarm_level"] = s.convertAlarmLevelToFrontendFormat(level)
				} else {
					settings["nobody_move_alarm_level"] = "disabled"
				}
			}

			// NoTurnOver 配置（duration 单位：分钟）
			if noTurnOver, ok := alarms["NoTurnOver"].(map[string]interface{}); ok {
				if threshold, ok := noTurnOver["threshold"].(map[string]interface{}); ok {
					if duration, ok := threshold["duration"].(float64); ok {
						settings["no_turn_over_duration"] = int(duration) // 分钟
					}
				}
				if level, ok := noTurnOver["level"].(string); ok {
					settings["no_turn_over_alarm_level"] = s.convertAlarmLevelToFrontendFormat(level)
				} else {
					settings["no_turn_over_alarm_level"] = "disabled"
				}
			}

			// SitUp 配置（只有 level，没有 duration）
			if sitUp, ok := alarms["SitUp"].(map[string]interface{}); ok {
				if level, ok := sitUp["level"].(string); ok {
					settings["situp_alarm_level"] = s.convertAlarmLevelToFrontendFormat(level)
				} else {
					settings["situp_alarm_level"] = "disabled"
				}
			}

			// OnBed 配置（duration 单位：秒）
			if onBed, ok := alarms["OnBed"].(map[string]interface{}); ok {
				if threshold, ok := onBed["threshold"].(map[string]interface{}); ok {
					if duration, ok := threshold["duration"].(float64); ok {
						settings["onbed_duration"] = int(duration) // 秒
					}
				}
				if level, ok := onBed["level"].(string); ok {
					settings["onbed_alarm_level"] = s.convertAlarmLevelToFrontendFormat(level)
				} else {
					settings["onbed_alarm_level"] = "disabled"
				}
			}

			// Fall 配置（Sensor fall）：返回 fallFlag (boolean)
			if fall, ok := alarms["Fall"].(map[string]interface{}); ok {
				if enabled, ok := fall["enabled"].(bool); ok {
					settings["fallFlag"] = enabled
				} else {
					settings["fallFlag"] = false
				}
			}

			// 设置默认值（如果字段不存在）
			s.setSleepaceDefaults(settings)
		} else {
			// 如果没有 alarms 配置，返回默认值
			return s.getDefaultSettingsHardValue(ctx, "sleepace"), nil
		}
	} else if deviceType == "radar" {
		// Radar 配置转换
		// 从 monitor_config 中提取配置项

		// 提取 alarms 配置
		if alarms, ok := monitorConfig["alarms"].(map[string]interface{}); ok {
			// 跌倒配置
			// 注意：DB 存储秒，前端显示秒（不需要转换）
			if fall, ok := alarms["Fall"].(map[string]interface{}); ok {
				if threshold, ok := fall["threshold"].(map[string]interface{}); ok {
					if duration, ok := threshold["duration"].(float64); ok {
						// 直接使用 DB 存储的秒值，不进行转换
						settings["suspected_fall_duration"] = int(duration)
					}
				}
				if level, ok := fall["level"].(string); ok {
					settings["fall_alarm_level"] = s.convertAlarmLevelToFrontendFormat(level)
				} else {
					settings["fall_alarm_level"] = "disabled"
				}
			}

			// 姿态检测配置（AngleException）：返回 posture_detection_enabled (boolean)
			if angleException, ok := alarms["AngleException"].(map[string]interface{}); ok {
				if enabled, ok := angleException["enabled"].(bool); ok {
					settings["posture_detection_enabled"] = enabled
				} else {
					settings["posture_detection_enabled"] = false
				}
			}

			// 坐地配置（SittingOnGround）
			// 注意：SittingOnGround 使用 AngleException 的 enabled，但有自己的 duration
			if sittingOnGround, ok := alarms["SittingOnGround"].(map[string]interface{}); ok {
				// SittingOnGround 的 alarm_level 使用 AngleException 的 enabled（如果存在）
				if angleException, ok := alarms["AngleException"].(map[string]interface{}); ok {
					if enabled, ok := angleException["enabled"].(bool); ok && enabled {
						settings["sitting_on_ground_alarm_level"] = "4" // enabled -> '4' (WARNING)
					} else {
						settings["sitting_on_ground_alarm_level"] = "disabled"
					}
				} else {
					// 如果没有 AngleException，使用 SittingOnGround 自己的 enabled
					if enabled, ok := sittingOnGround["enabled"].(bool); ok && enabled {
						settings["sitting_on_ground_alarm_level"] = "4" // enabled -> '4' (WARNING)
					} else {
						settings["sitting_on_ground_alarm_level"] = "disabled"
					}
				}
				// SittingOnGround duration
				if threshold, ok := sittingOnGround["threshold"].(map[string]interface{}); ok {
					if duration, ok := threshold["duration"].(float64); ok {
						settings["sitting_on_ground_duration"] = int(duration) // 秒
					}
				}
			}

			// 离床配置（Radar_LeftBed）
			if leftBed, ok := alarms["Radar_LeftBed"].(map[string]interface{}); ok {
				if threshold, ok := leftBed["threshold"].(map[string]interface{}); ok {
					if startHour, ok := threshold["start_hour"].(float64); ok {
						settings["leave_detection_start_hour"] = int(startHour)
					}
					if startMinute, ok := threshold["start_minute"].(float64); ok {
						settings["leave_detection_start_minute"] = int(startMinute)
					}
					if endHour, ok := threshold["end_hour"].(float64); ok {
						settings["leave_detection_end_hour"] = int(endHour)
					}
					if endMinute, ok := threshold["end_minute"].(float64); ok {
						settings["leave_detection_end_minute"] = int(endMinute)
					}
					if duration, ok := threshold["duration"].(float64); ok {
						settings["leave_detection_duration"] = int(duration) // 分钟
					}
				}
				if level, ok := leftBed["level"].(string); ok {
					settings["leave_alarm_level"] = s.convertAlarmLevelToFrontendFormat(level)
				} else {
					settings["leave_alarm_level"] = "disabled"
				}
			}

			// 心率配置（Radar_AbnormalHeartRate）
			if heartRate, ok := alarms["Radar_AbnormalHeartRate"].(map[string]interface{}); ok {
				if threshold, ok := heartRate["threshold"].(map[string]interface{}); ok {
					if min, ok := threshold["min"].(float64); ok {
						settings["lower_heart_rate"] = int(min)
					}
					if max, ok := threshold["max"].(float64); ok {
						settings["upper_heart_rate"] = int(max)
					}
				}
				if level, ok := heartRate["level"].(string); ok {
					convertedLevel := s.convertAlarmLevelToFrontendFormat(level)
					settings["heart_rate_slow_alarm_level"] = convertedLevel
					settings["heart_rate_fast_alarm_level"] = convertedLevel
				} else {
					settings["heart_rate_slow_alarm_level"] = "disabled"
					settings["heart_rate_fast_alarm_level"] = "disabled"
				}
			}

			// 呼吸率配置（Radar_AbnormalRespiratoryRate）
			if breathRate, ok := alarms["Radar_AbnormalRespiratoryRate"].(map[string]interface{}); ok {
				if threshold, ok := breathRate["threshold"].(map[string]interface{}); ok {
					if min, ok := threshold["min"].(float64); ok {
						settings["lower_breath_rate"] = int(min)
					}
					if max, ok := threshold["max"].(float64); ok {
						settings["upper_breath_rate"] = int(max)
					}
				}
				if level, ok := breathRate["level"].(string); ok {
					convertedLevel := s.convertAlarmLevelToFrontendFormat(level)
					settings["breath_rate_slow_alarm_level"] = convertedLevel
					settings["breath_rate_fast_alarm_level"] = convertedLevel
				} else {
					settings["breath_rate_slow_alarm_level"] = "disabled"
					settings["breath_rate_fast_alarm_level"] = "disabled"
				}
			}

			// 呼吸暂停配置（Radar_ApneaHypopnea）
			if breathPause, ok := alarms["Radar_ApneaHypopnea"].(map[string]interface{}); ok {
				if level, ok := breathPause["level"].(string); ok {
					settings["breath_pause_alarm_level"] = s.convertAlarmLevelToFrontendFormat(level)
				} else {
					settings["breath_pause_alarm_level"] = "disabled"
				}
			}

			// 滞留配置（Stay）
			if stay, ok := alarms["Stay"].(map[string]interface{}); ok {
				if threshold, ok := stay["threshold"].(map[string]interface{}); ok {
					if duration, ok := threshold["duration"].(float64); ok {
						settings["stay_detection_duration"] = int(duration) // 分钟
					}
				}
				if level, ok := stay["level"].(string); ok {
					settings["stay_alarm_level"] = s.convertAlarmLevelToFrontendFormat(level)
				} else {
					settings["stay_alarm_level"] = "disabled"
				}
			}

			// 弱生命体征配置（VitalsWeak）
			if vitalsWeak, ok := alarms["VitalsWeak"].(map[string]interface{}); ok {
				if threshold, ok := vitalsWeak["threshold"].(map[string]interface{}); ok {
					if duration, ok := threshold["duration"].(float64); ok {
						settings["weak_vital_duration"] = int(duration) // 分钟
					}
					if sensitivity, ok := threshold["sensitivity"].(float64); ok {
						settings["weak_vital_sensitivity"] = int(sensitivity)
					}
				}
				if level, ok := vitalsWeak["level"].(string); ok {
					settings["weak_vital_alarm_level"] = s.convertAlarmLevelToFrontendFormat(level)
				} else {
					settings["weak_vital_alarm_level"] = "disabled"
				}
			}

			// 24小时无活动配置（NoActivity24h）
			if noActivity, ok := alarms["NoActivity24h"].(map[string]interface{}); ok {
				if level, ok := noActivity["level"].(string); ok {
					settings["inactivity_alarm_level"] = s.convertAlarmLevelToFrontendFormat(level)
				} else {
					settings["inactivity_alarm_level"] = "disabled"
				}
			}

			// 设置默认值（如果字段不存在）
			s.setRadarDefaults(settings)
		} else {
			// 如果没有 alarms 配置，返回默认值
			return s.getDefaultSettingsHardValue(ctx, "radar"), nil
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
								"level":   leftBedLevel,
								"enabled": !s.isDisabled(leftBedLevel),
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
				threshold := map[string]interface{}{
					"min": minHeartRate,
					"max": maxHeartRate,
				}
				// 支持 slow_duration 和 fast_duration（如果存在）
				if slowDuration, ok := settings["heart_rate_slow_duration"]; ok {
					if dur, ok := slowDuration.(float64); ok {
						threshold["slow_duration"] = int(dur)
					} else if dur, ok := slowDuration.(int); ok {
						threshold["slow_duration"] = dur
					}
				}
				if fastDuration, ok := settings["heart_rate_fast_duration"]; ok {
					if dur, ok := fastDuration.(float64); ok {
						threshold["fast_duration"] = int(dur)
					} else if dur, ok := fastDuration.(int); ok {
						threshold["fast_duration"] = dur
					}
				}
				// 如果只有 slow_duration，也保存为 duration（兼容旧数据）
				if _, hasSlow := threshold["slow_duration"]; hasSlow {
					if _, hasFast := threshold["fast_duration"]; !hasFast {
						threshold["duration"] = threshold["slow_duration"]
					}
				}
				alarms["HeartRate"] = map[string]interface{}{
					"level":     heartRateLevel,
					"enabled":   heartRateLevel != "disabled",
					"threshold": threshold,
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
				threshold := map[string]interface{}{
					"min": minBreathRate,
					"max": maxBreathRate,
				}
				// 支持 slow_duration 和 fast_duration（如果存在）
				if slowDuration, ok := settings["breath_rate_slow_duration"]; ok {
					if dur, ok := slowDuration.(float64); ok {
						threshold["slow_duration"] = int(dur)
					} else if dur, ok := slowDuration.(int); ok {
						threshold["slow_duration"] = dur
					}
				}
				if fastDuration, ok := settings["breath_rate_fast_duration"]; ok {
					if dur, ok := fastDuration.(float64); ok {
						threshold["fast_duration"] = int(dur)
					} else if dur, ok := fastDuration.(int); ok {
						threshold["fast_duration"] = dur
					}
				}
				// 如果只有 slow_duration，也保存为 duration（兼容旧数据）
				if _, hasSlow := threshold["slow_duration"]; hasSlow {
					if _, hasFast := threshold["fast_duration"]; !hasFast {
						threshold["duration"] = threshold["slow_duration"]
					}
				}
				alarms["BreathRate"] = map[string]interface{}{
					"level":     breathRateLevel,
					"enabled":   breathRateLevel != "disabled",
					"threshold": threshold,
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

		// BodyMove 配置（duration 单位：分钟）
		if bodyMoveDuration, ok := settings["body_move_duration"]; ok {
			bodyMoveLevel := "disabled"
			if level, ok := settings["body_move_alarm_level"].(string); ok && level != "" {
				bodyMoveLevel = level
			}
			alarms["BodyMove"] = map[string]interface{}{
				"level":   bodyMoveLevel,
				"enabled": bodyMoveLevel != "disabled",
				"threshold": map[string]interface{}{
					"duration": bodyMoveDuration, // 分钟
				},
			}
		}

		// NobodyMove 配置（duration 单位：分钟）
		if nobodyMoveDuration, ok := settings["nobody_move_duration"]; ok {
			nobodyMoveLevel := "disabled"
			if level, ok := settings["nobody_move_alarm_level"].(string); ok && level != "" {
				nobodyMoveLevel = level
			}
			alarms["NobodyMove"] = map[string]interface{}{
				"level":   nobodyMoveLevel,
				"enabled": nobodyMoveLevel != "disabled",
				"threshold": map[string]interface{}{
					"duration": nobodyMoveDuration, // 分钟
				},
			}
		}

		// NoTurnOver 配置（duration 单位：分钟）
		if noTurnOverDuration, ok := settings["no_turn_over_duration"]; ok {
			noTurnOverLevel := "disabled"
			if level, ok := settings["no_turn_over_alarm_level"].(string); ok && level != "" {
				noTurnOverLevel = level
			}
			alarms["NoTurnOver"] = map[string]interface{}{
				"level":   noTurnOverLevel,
				"enabled": noTurnOverLevel != "disabled",
				"threshold": map[string]interface{}{
					"duration": noTurnOverDuration, // 分钟
				},
			}
		}

		// SitUp 配置（只有 level，没有 duration）
		if level, ok := settings["situp_alarm_level"].(string); ok && level != "" {
			alarms["SitUp"] = map[string]interface{}{
				"level":   level,
				"enabled": level != "disabled",
			}
		}

		// OnBed 配置（duration 单位：秒）
		if onbedDuration, ok := settings["onbed_duration"]; ok {
			onbedLevel := "disabled"
			if level, ok := settings["onbed_alarm_level"].(string); ok && level != "" {
				onbedLevel = level
			}
			alarms["OnBed"] = map[string]interface{}{
				"level":   onbedLevel,
				"enabled": onbedLevel != "disabled",
				"threshold": map[string]interface{}{
					"duration": onbedDuration, // 秒
				},
			}
		}
		// Sensor fall: 前端发送 fallFlag (boolean)
		if fallFlag, ok := settings["fallFlag"].(bool); ok {
			alarms["Fall"] = map[string]interface{}{
				"enabled": fallFlag,
			}
		}

		monitorConfig["alarms"] = alarms
	} else if deviceType == "radar" {
		// Radar 配置转换
		// 跌倒配置
		// 注意：前端发送秒，DB 存储秒（不需要转换），传给硬件会转换*0.1
		if fallDuration, ok := settings["suspected_fall_duration"]; ok {
			fallLevel := "disabled"
			if level, ok := settings["fall_alarm_level"].(string); ok && level != "" {
				fallLevel = level
			}
			// 直接使用前端发送的秒值，不进行转换
			fallDurationInSeconds := 0
			if dur, ok := fallDuration.(float64); ok {
				fallDurationInSeconds = int(dur)
			} else if dur, ok := fallDuration.(int); ok {
				fallDurationInSeconds = dur
			}
			alarms["Fall"] = map[string]interface{}{
				"level":   fallLevel,
				"enabled": fallLevel != "disabled",
				"threshold": map[string]interface{}{
					"duration": fallDurationInSeconds, // 秒
				},
			}
		}

		// 姿态检测配置（AngleException）
		// 作为阈值（enabled），不存储 level
		// 前端发送 posture_detection_enabled (boolean)
		if postureEnabled, ok := settings["posture_detection_enabled"].(bool); ok {
			alarms["AngleException"] = map[string]interface{}{
				"enabled": postureEnabled,
			}
			// SittingOnGround 使用 AngleException 的 enabled
			if sittingOnGroundDuration, ok := settings["sitting_on_ground_duration"]; ok {
				// SittingOnGround 的 enabled 使用 AngleException 的 enabled
				sittingOnGroundEnabled := postureEnabled
				if level, ok := settings["sitting_on_ground_alarm_level"].(string); ok && level != "" {
					sittingOnGroundEnabled = level != "disabled" && level != ""
				}
				alarms["SittingOnGround"] = map[string]interface{}{
					"enabled": sittingOnGroundEnabled,
					"threshold": map[string]interface{}{
						"duration": sittingOnGroundDuration, // 秒
					},
				}
			}
		}

		// 离床配置（Radar_LeftBed）
		if leaveStartHour, ok := settings["leave_detection_start_hour"]; ok {
			if leaveStartMinute, ok := settings["leave_detection_start_minute"]; ok {
				if leaveEndHour, ok := settings["leave_detection_end_hour"]; ok {
					if leaveEndMinute, ok := settings["leave_detection_end_minute"]; ok {
						if leaveDuration, ok := settings["leave_detection_duration"]; ok {
							leaveLevel := "disabled"
							if level, ok := settings["leave_alarm_level"].(string); ok && level != "" {
								leaveLevel = level
							}
							alarms["Radar_LeftBed"] = map[string]interface{}{
								"level":   leaveLevel,
								"enabled": leaveLevel != "disabled",
								"threshold": map[string]interface{}{
									"start_hour":   leaveStartHour,
									"start_minute": leaveStartMinute,
									"end_hour":     leaveEndHour,
									"end_minute":   leaveEndMinute,
									"duration":     leaveDuration, // 分钟
								},
							}
						}
					}
				}
			}
		}

		// 心率配置（Radar_AbnormalHeartRate）
		if lowerHeartRate, ok := settings["lower_heart_rate"]; ok {
			if upperHeartRate, ok := settings["upper_heart_rate"]; ok {
				heartRateLevel := "disabled"
				if level, ok := settings["heart_rate_slow_alarm_level"].(string); ok && level != "" {
					heartRateLevel = level
				}
				// Radar 的心率报警级别：slow 和 fast 使用相同的 level
				if level, ok := settings["heart_rate_fast_alarm_level"].(string); ok && level != "" {
					heartRateLevel = level
				}
				alarms["Radar_AbnormalHeartRate"] = map[string]interface{}{
					"level":   heartRateLevel,
					"enabled": heartRateLevel != "disabled",
					"threshold": map[string]interface{}{
						"min": lowerHeartRate,
						"max": upperHeartRate,
						// Radar 的心率持续时间固定为 1 分钟（60秒），不配置
					},
				}
			}
		}

		// 呼吸率配置（Radar_AbnormalRespiratoryRate）
		if lowerBreathRate, ok := settings["lower_breath_rate"]; ok {
			if upperBreathRate, ok := settings["upper_breath_rate"]; ok {
				breathRateLevel := "disabled"
				if level, ok := settings["breath_rate_slow_alarm_level"].(string); ok && level != "" {
					breathRateLevel = level
				}
				// Radar 的呼吸率报警级别：slow 和 fast 使用相同的 level
				if level, ok := settings["breath_rate_fast_alarm_level"].(string); ok && level != "" {
					breathRateLevel = level
				}
				alarms["Radar_AbnormalRespiratoryRate"] = map[string]interface{}{
					"level":   breathRateLevel,
					"enabled": breathRateLevel != "disabled",
					"threshold": map[string]interface{}{
						"min": lowerBreathRate,
						"max": upperBreathRate,
						// Radar 的呼吸率持续时间固定为 1 分钟（60秒），不配置
					},
				}
			}
		}

		// 呼吸暂停配置（Radar_ApneaHypopnea）
		if breathPauseLevel, ok := settings["breath_pause_alarm_level"].(string); ok && breathPauseLevel != "" {
			alarms["Radar_ApneaHypopnea"] = map[string]interface{}{
				"level":   breathPauseLevel,
				"enabled": breathPauseLevel != "disabled",
				// Radar 的呼吸暂停持续时间固定为 20-30 秒，不配置
			}
		}

		// 滞留配置（Stay）
		if stayDuration, ok := settings["stay_detection_duration"]; ok {
			stayLevel := "disabled"
			if level, ok := settings["stay_alarm_level"].(string); ok && level != "" {
				stayLevel = level
			}
			alarms["Stay"] = map[string]interface{}{
				"level":   stayLevel,
				"enabled": stayLevel != "disabled",
				"threshold": map[string]interface{}{
					"duration": stayDuration, // 分钟
				},
			}
		}

		// 弱生命体征配置（VitalsWeak）
		if weakVitalDuration, ok := settings["weak_vital_duration"]; ok {
			if weakVitalSensitivity, ok := settings["weak_vital_sensitivity"]; ok {
				weakVitalLevel := "disabled"
				if level, ok := settings["weak_vital_alarm_level"].(string); ok && level != "" {
					weakVitalLevel = level
				}
				alarms["VitalsWeak"] = map[string]interface{}{
					"level":   weakVitalLevel,
					"enabled": weakVitalLevel != "disabled",
					"threshold": map[string]interface{}{
						"duration":    weakVitalDuration,    // 分钟
						"sensitivity": weakVitalSensitivity, // 1-99
					},
				}
			}
		}

		// 24小时无活动配置（NoActivity24h）
		if inactivityLevel, ok := settings["inactivity_alarm_level"].(string); ok && inactivityLevel != "" {
			alarms["NoActivity24h"] = map[string]interface{}{
				"level":   inactivityLevel,
				"enabled": inactivityLevel != "disabled",
				// 24小时无活动的时间固定为 24 小时，不配置
			}
		}

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
		"disabled":      true, // 前端发送的格式（小写）
		"DISABLED":      true, // 大写格式（兼容）
		"EMERGENCY":     true,
		"WARNING":       true,
		"ERROR":         true,
		"INFORMATIONAL": true,
		// 前端发送的格式：'0' = EMERGENCY, '4' = WARNING
		"0": true,
		"4": true,
	}

	for key, value := range settings {
		if strings.HasSuffix(key, "_alarm_level") {
			// 跳过 fall_alarm_level 和 posture_detection_alarm_level，它们作为阈值（enabled）而不是 alarm_level
			if key == "fall_alarm_level" || key == "posture_detection_alarm_level" {
				// 只验证，不转换（在 convertFlatToMonitorConfig 中会转换为 enabled）
				if level, ok := value.(string); ok {
					if !validLevels[level] {
						return fmt.Errorf("invalid alarm level '%s' for field '%s'", level, key)
					}
				}
				continue
			}

			if level, ok := value.(string); ok {
				// 转换前端格式到后端格式
				convertedLevel := level
				if level == "0" {
					convertedLevel = "EMERGENCY"
				} else if level == "4" {
					convertedLevel = "WARNING"
				} else if strings.EqualFold(level, "disabled") || strings.EqualFold(level, "DISABLED") {
					// 统一转换为 "DISABLED"（大写）
					convertedLevel = "DISABLED"
				}

				if !validLevels[level] {
					return fmt.Errorf("invalid alarm level '%s' for field '%s'", level, key)
				}

				// 更新设置中的值为转换后的格式
				settings[key] = convertedLevel
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
		"left_bed_start_hour":          0,
		"left_bed_start_minute":        0,
		"left_bed_end_hour":            0,
		"left_bed_end_minute":          0,
		"left_bed_duration":            0,
		"left_bed_alarm_level":         "disabled",
		"min_heart_rate":               0,
		"heart_rate_slow_duration":     0,
		"heart_rate_slow_alarm_level":  "disabled",
		"max_heart_rate":               0,
		"heart_rate_fast_duration":     0,
		"heart_rate_fast_alarm_level":  "disabled",
		"min_breath_rate":              0,
		"breath_rate_slow_duration":    0,
		"breath_rate_slow_alarm_level": "disabled",
		"max_breath_rate":              0,
		"breath_rate_fast_duration":    0,
		"breath_rate_fast_alarm_level": "disabled",
		"breath_pause_duration":        0,
		"breath_pause_alarm_level":     "disabled",
		"body_move_duration":           0,
		"body_move_alarm_level":        "disabled",
		"nobody_move_duration":         0,
		"nobody_move_alarm_level":      "disabled",
		"no_turn_over_duration":        0,
		"no_turn_over_alarm_level":     "disabled",
		"situp_alarm_level":            "disabled",
		"onbed_duration":               0,
		"onbed_alarm_level":            "disabled",
		"fallFlag":                     false, // Sensor fall: boolean (default: disabled)
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

// GetDefaultDeviceMonitorSettings 获取默认配置
// 阈值：硬编码（与 System 租户模板设备的值相同）
// Alarm Level：优先从当前租户的 alarm_cloud 读取，如果没有则使用硬编码值（与 AlarmCloud.vue 中的默认值相同）
func (s *deviceMonitorSettingsService) GetDefaultDeviceMonitorSettings(ctx context.Context, tenantID string, deviceType string) (*GetDeviceMonitorSettingsResponse, error) {
	if deviceType != "sleepace" && deviceType != "radar" {
		return nil, fmt.Errorf("invalid device_type: %s (must be 'sleepace' or 'radar')", deviceType)
	}

	// 1. 获取硬编码的阈值和 Alarm Level 默认值
	settings := s.getDefaultSettingsHardValue(ctx, deviceType)

	// 2. 优先从当前租户的 alarm_cloud 读取 Alarm Level，覆盖硬编码值
	if err := s.getAlarmLevelFromCloud(ctx, tenantID, deviceType, settings); err != nil {
		s.logger.Warn("Failed to fill default alarm levels from cloud, using hardcoded values",
			zap.String("tenant_id", tenantID),
			zap.String("device_type", deviceType),
			zap.Error(err),
		)
		// 如果失败，使用硬编码值（已经在 settings 中）
	}

	return &GetDeviceMonitorSettingsResponse{
		Settings: settings,
	}, nil
}

// getAlarmLevelFromCloud 从 alarm_cloud 获取报警级别并填充到 settings
// 优先从当前租户的 alarm_cloud 读取，如果没有则使用硬编码值（不回退到 System 租户）
func (s *deviceMonitorSettingsService) getAlarmLevelFromCloud(ctx context.Context, tenantID, deviceType string, settings map[string]interface{}) error {
	// 获取当前租户的 alarm_cloud 配置
	alarmCloud, err := s.alarmCloudRepo.GetAlarmCloud(ctx, tenantID)
	if err != nil {
		// 如果租户配置不存在，直接返回（使用硬编码值，已经在 settings 中）
		return nil
	}

	// 解析 device_alarms JSONB
	var deviceAlarms map[string]map[string]interface{}
	if len(alarmCloud.DeviceAlarms) > 0 {
		if err := json.Unmarshal(alarmCloud.DeviceAlarms, &deviceAlarms); err != nil {
			return fmt.Errorf("failed to parse device_alarms: %w", err)
		}
	} else {
		return nil // 没有 device_alarms 配置，使用硬编码值
	}

	// 获取设备类型的报警配置
	deviceTypeKey := ""
	if deviceType == "sleepace" {
		deviceTypeKey = "SleepPad"
	} else if deviceType == "radar" {
		deviceTypeKey = "Radar"
	} else {
		return nil // 未知设备类型
	}

	deviceAlarmsConfig, ok := deviceAlarms[deviceTypeKey]
	if !ok {
		return nil // 没有该设备类型的配置，使用硬编码值
	}

	// 映射关系：alarm_cloud 的报警类型 -> settings 的字段名
	alarmTypeMapping := s.getAlarmTypeMapping(deviceType)

	// 填充默认报警级别（覆盖硬编码值）
	for alarmType, defaultLevel := range deviceAlarmsConfig {
		if levelStr, ok := defaultLevel.(string); ok {
			// 转换 DangerLevel 到 AlarmLevel
			alarmLevel := s.convertDangerLevelToAlarmLevel(levelStr)

			// 查找对应的 settings 字段
			fieldNames, ok := alarmTypeMapping[alarmType]
			if !ok {
				continue // 未知的报警类型
			}

			for _, fieldName := range fieldNames {
				// 特殊处理：fallFlag 是 boolean，不是 alarm_level
				if fieldName == "fallFlag" {
					// fallFlag: "0" (EMERGENCY) 或 "disabled" -> boolean
					settings[fieldName] = alarmLevel != "disabled" && alarmLevel != ""
				} else {
					// 其他字段：直接使用 alarm_level 字符串
					settings[fieldName] = alarmLevel
				}
			}
		}
	}

	return nil
}

// getAlarmTypeMapping 获取报警类型到 settings 字段名的映射
func (s *deviceMonitorSettingsService) getAlarmTypeMapping(deviceType string) map[string][]string {
	if deviceType == "sleepace" {
		return map[string][]string{
			"SleepPad_LeftBed":                 {"left_bed_alarm_level"},
			"SleepPad_AbnormalHeartRate":       {"heart_rate_slow_alarm_level", "heart_rate_fast_alarm_level"},
			"SleepPad_AbnormalRespiratoryRate": {"breath_rate_slow_alarm_level", "breath_rate_fast_alarm_level"},
			"SleepPad_ApneaHypopnea":           {"breath_pause_alarm_level"},
			"Fall":                             {"fallFlag"}, // Sensor fall: boolean (not alarm_level)
			"SleepPad_AbnormalBodyMovement":    {"body_move_alarm_level", "nobody_move_alarm_level", "no_turn_over_alarm_level"},
			"SleepPad_SitUp":                   {"situp_alarm_level"},
			"SleepPad_InBed":                   {"onbed_alarm_level"},
		}
	} else if deviceType == "radar" {
		return map[string][]string{
			"Fall":                          {"fall_alarm_level"},
			"SuspectedFall":                 {"fall_alarm_level"},
			"Radar_LeftBed":                 {"leave_alarm_level"},
			"Radar_AbnormalHeartRate":       {"heart_rate_slow_alarm_level", "heart_rate_fast_alarm_level"},
			"Radar_AbnormalRespiratoryRate": {"breath_rate_slow_alarm_level", "breath_rate_fast_alarm_level"},
			"Radar_ApneaHypopnea":           {"breath_pause_alarm_level"},
			"VitalsWeak":                    {"weak_vital_alarm_level"},
			"Stay":                          {"stay_alarm_level"},
			"NoActivity24h":                 {"inactivity_alarm_level"},
			// AngleException 不映射到 posture_detection_enabled
			// posture_detection_enabled 是独立的 boolean 开关，不属于 alarm_level
			// AngleException 只影响 sitting_on_ground_alarm_level（如果存在）
			"SittingOnGround": {"sitting_on_ground_alarm_level"},
		}
	}
	return map[string][]string{}
}

// convertDangerLevelToAlarmLevel 转换 DangerLevel 到 AlarmLevel
func (s *deviceMonitorSettingsService) convertDangerLevelToAlarmLevel(dangerLevel string) string {
	mapping := map[string]string{
		"DISABLE":       "disabled",
		"EMERGENCY":     "0",
		"WARNING":       "4",
		"ERROR":         "3",
		"CRITICAL":      "2",
		"ALERT":         "1",
		"INFORMATIONAL": "5",
	}
	if level, ok := mapping[dangerLevel]; ok {
		return level
	}
	return "disabled"
}

// convertAlarmLevelToFrontendFormat 将数据库中的 alarm_level 格式转换为前端格式
// 数据库存储：'EMERGENCY', 'WARNING', 'DISABLED'
// 前端期望：'0', '4', 'disabled'（小写）
func (s *deviceMonitorSettingsService) convertAlarmLevelToFrontendFormat(level string) string {
	mapping := map[string]string{
		"EMERGENCY":     "0",
		"WARNING":       "4",
		"ERROR":         "3",
		"CRITICAL":      "2",
		"ALERT":         "1",
		"INFORMATIONAL": "5",
		"DISABLED":      "disabled", // 数据库存储大写，转换为前端小写
		"DISABLE":       "disabled",
		"disabled":      "disabled", // 兼容小写（如果数据库中有）
		// 如果已经是前端格式，直接返回
		"0": "0",
		"4": "4",
	}
	if converted, ok := mapping[level]; ok {
		return converted
	}
	// 未知格式，默认返回 disabled
	return "disabled"
}

// isDisabled 检查 alarm level 是否为 disabled（不区分大小写）
func (s *deviceMonitorSettingsService) isDisabled(level string) bool {
	return level == "" || strings.EqualFold(level, "disabled") || strings.EqualFold(level, "DISABLED")
}

// applyDeviceDefaults 应用设备特定默认值
func (s *deviceMonitorSettingsService) applyDeviceDefaults(deviceType string, settings map[string]interface{}) {
	if deviceType == "radar" {
		// 雷达默认值
		if val, exists := settings["radar_function_mode"]; !exists || val == nil || val == 0 {
			settings["radar_function_mode"] = 15 // Full Function
		}
		if val, exists := settings["leave_detection_start_hour"]; !exists || val == nil || val == 0 {
			settings["leave_detection_start_hour"] = 21 // 9:00 PM
		}
		if val, exists := settings["leave_detection_start_minute"]; !exists || val == nil {
			settings["leave_detection_start_minute"] = 0
		}
		if val, exists := settings["leave_detection_end_hour"]; !exists || val == nil || val == 0 {
			settings["leave_detection_end_hour"] = 8 // 8:00 AM
		}
		if val, exists := settings["leave_detection_end_minute"]; !exists || val == nil {
			settings["leave_detection_end_minute"] = 0
		}
	}
}

// setRadarDefaults 设置 Radar 默认值（如果字段不存在）
func (s *deviceMonitorSettingsService) setRadarDefaults(settings map[string]interface{}) {
	defaults := map[string]interface{}{
		"radar_function_mode":           0,
		"suspected_fall_duration":       0,
		"fall_alarm_level":              "disabled",
		"sitting_on_ground_duration":    0,
		"sitting_on_ground_alarm_level": "disabled",
		"stay_detection_duration":       0,
		"stay_alarm_level":              "disabled",
		"leave_detection_start_hour":    0,
		"leave_detection_start_minute":  0,
		"leave_detection_end_hour":      0,
		"leave_detection_end_minute":    0,
		"leave_detection_duration":      0,
		"leave_alarm_level":             "disabled",
		"lower_heart_rate":              0,
		"heart_rate_slow_alarm_level":   "disabled",
		"upper_heart_rate":              0,
		"heart_rate_fast_alarm_level":   "disabled",
		"lower_breath_rate":             0,
		"breath_rate_slow_alarm_level":  "disabled",
		"upper_breath_rate":             0,
		"breath_rate_fast_alarm_level":  "disabled",
		"breath_pause_alarm_level":      "disabled",
		"weak_vital_duration":           0,
		"weak_vital_sensitivity":        0,
		"weak_vital_alarm_level":        "disabled",
		"inactivity_alarm_level":        "disabled",
		"posture_detection_enabled":     false, // Posture detection: boolean (default: disabled)
	}

	for key, defaultValue := range defaults {
		if _, exists := settings[key]; !exists {
			settings[key] = defaultValue
		}
	}
}

// getSleepaceSettingsFromHardware 从硬件读取 Sleepace 设备配置（参考 v1.0 实现）
// 参考：wisefido-backend/wisefido-sleepace/modules/sleepace_service.go::GetDeviceMonitorSettings
func (s *deviceMonitorSettingsService) getSleepaceSettingsFromHardware(ctx context.Context, device *domain.Device, deviceID string) (*GetDeviceMonitorSettingsResponse, error) {
	// 调用 Sleepace API 从硬件读取配置
	request := SleepaceRequest{
		Token: s.sleepaceClient.token,
		Data: map[string]any{
			"userId": deviceID,
		},
	}

	var response SleepaceResponse
	resp, err := s.sleepaceClient.httpClient.R().
		SetBody(request).
		SetResult(&response).
		Post("/sleepace/getalarmnotifyconfig")

	if err != nil {
		return nil, fmt.Errorf("failed to call Sleepace API: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("Sleepace API returned status code: %d", resp.StatusCode())
	}

	if response.Status != 0 {
		return nil, fmt.Errorf("Sleepace API error: %s (status: %d)", response.Msg, response.Status)
	}

	// 解析硬件返回的配置
	// 参考 v1.0: models.SleepaceMonitorSettings 结构
	var hardwareSettings map[string]interface{}
	if err := json.Unmarshal(response.Data, &hardwareSettings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal hardware settings: %w", err)
	}

	// 将硬件返回的配置转换为 flat 结构（与前端期望的格式一致）
	settings := make(map[string]interface{})

	// 离床时间配置
	if leftBedStartHour, ok := hardwareSettings["leftBedStartHour"].(float64); ok {
		settings["left_bed_start_hour"] = int(leftBedStartHour)
	}
	if leftBedStartMinute, ok := hardwareSettings["leftBedStartMinute"].(float64); ok {
		settings["left_bed_start_minute"] = int(leftBedStartMinute)
	}
	if leftBedEndHour, ok := hardwareSettings["leftBedEndHour"].(float64); ok {
		settings["left_bed_end_hour"] = int(leftBedEndHour)
	}
	if leftBedEndMinute, ok := hardwareSettings["leftBedEndMinute"].(float64); ok {
		settings["left_bed_end_minute"] = int(leftBedEndMinute)
	}
	if leftBedDuration, ok := hardwareSettings["leftBedDuration"].(float64); ok {
		settings["left_bed_duration"] = int(leftBedDuration)
	}
	if leftBedFlag, ok := hardwareSettings["leftBedFlag"].(float64); ok {
		if leftBedFlag == 1 {
			settings["left_bed_alarm_level"] = "0" // EMERG (默认)
		} else {
			settings["left_bed_alarm_level"] = "disabled"
		}
	}

	// 心率配置
	if minHeartRate, ok := hardwareSettings["minHeartRate"].(float64); ok {
		settings["min_heart_rate"] = int(minHeartRate)
	}
	if maxHeartRate, ok := hardwareSettings["maxHeartRate"].(float64); ok {
		settings["max_heart_rate"] = int(maxHeartRate)
	}
	if heartRateSlowDuration, ok := hardwareSettings["heartRateSlowDuration"].(float64); ok {
		settings["heart_rate_slow_duration"] = int(heartRateSlowDuration)
	}
	if heartRateFastDuration, ok := hardwareSettings["heartRateFastDuration"].(float64); ok {
		settings["heart_rate_fast_duration"] = int(heartRateFastDuration)
	}
	if heartRateSlowFlag, ok := hardwareSettings["heartRateSlowFlag"].(float64); ok {
		if heartRateSlowFlag == 1 {
			settings["heart_rate_slow_alarm_level"] = "0" // EMERG (默认)
		} else {
			settings["heart_rate_slow_alarm_level"] = "disabled"
		}
	}
	if heartRateFastFlag, ok := hardwareSettings["heartRateFastFlag"].(float64); ok {
		if heartRateFastFlag == 1 {
			settings["heart_rate_fast_alarm_level"] = "0" // EMERG (默认)
		} else {
			settings["heart_rate_fast_alarm_level"] = "disabled"
		}
	}

	// 呼吸率配置
	if minBreathRate, ok := hardwareSettings["minBreathRate"].(float64); ok {
		settings["min_breath_rate"] = int(minBreathRate)
	}
	if maxBreathRate, ok := hardwareSettings["maxBreathRate"].(float64); ok {
		settings["max_breath_rate"] = int(maxBreathRate)
	}
	if breathRateSlowDuration, ok := hardwareSettings["breathRateSlowDuration"].(float64); ok {
		settings["breath_rate_slow_duration"] = int(breathRateSlowDuration)
	}
	if breathRateFastDuration, ok := hardwareSettings["breathRateFastDuration"].(float64); ok {
		settings["breath_rate_fast_duration"] = int(breathRateFastDuration)
	}
	if breathRateSlowFlag, ok := hardwareSettings["breathRateSlowFlag"].(float64); ok {
		if breathRateSlowFlag == 1 {
			settings["breath_rate_slow_alarm_level"] = "0" // EMERG (默认)
		} else {
			settings["breath_rate_slow_alarm_level"] = "disabled"
		}
	}
	if breathRateFastFlag, ok := hardwareSettings["breathRateFastFlag"].(float64); ok {
		if breathRateFastFlag == 1 {
			settings["breath_rate_fast_alarm_level"] = "0" // EMERG (默认)
		} else {
			settings["breath_rate_fast_alarm_level"] = "disabled"
		}
	}

	// 呼吸暂停配置
	if breathPauseDuration, ok := hardwareSettings["breathPauseDuration"].(float64); ok {
		settings["breath_pause_duration"] = int(breathPauseDuration)
	}
	if breathPauseFlag, ok := hardwareSettings["breathPauseFlag"].(float64); ok {
		if breathPauseFlag == 1 {
			settings["breath_pause_alarm_level"] = "0" // EMERG (默认)
		} else {
			settings["breath_pause_alarm_level"] = "disabled"
		}
	}

	// 身体移动配置
	if bodyMoveDuration, ok := hardwareSettings["bodyMoveDuration"].(float64); ok {
		settings["body_move_duration"] = int(bodyMoveDuration)
	}
	if bodyMoveFlag, ok := hardwareSettings["bodyMoveFlag"].(float64); ok {
		if bodyMoveFlag == 1 {
			settings["body_move_alarm_level"] = "0" // EMERG (默认)
		} else {
			settings["body_move_alarm_level"] = "disabled"
		}
	}

	// 无身体移动配置
	if nobodyMoveDuration, ok := hardwareSettings["nobodyMoveDuration"].(float64); ok {
		settings["nobody_move_duration"] = int(nobodyMoveDuration)
	}
	if nobodyMoveFlag, ok := hardwareSettings["nobodyMoveFlag"].(float64); ok {
		if nobodyMoveFlag == 1 {
			settings["nobody_move_alarm_level"] = "0" // EMERG (默认)
		} else {
			settings["nobody_move_alarm_level"] = "disabled"
		}
	}

	// 无翻身配置
	if noTurnOverDuration, ok := hardwareSettings["noTurnOverDuration"].(float64); ok {
		settings["no_turn_over_duration"] = int(noTurnOverDuration)
	}
	if noTurnOverFlag, ok := hardwareSettings["noTurnOverFlag"].(float64); ok {
		if noTurnOverFlag == 1 {
			settings["no_turn_over_alarm_level"] = "0" // EMERG (默认)
		} else {
			settings["no_turn_over_alarm_level"] = "disabled"
		}
	}

	// 坐起配置
	if situpFlag, ok := hardwareSettings["situpFlag"].(float64); ok {
		if situpFlag == 1 {
			settings["situp_alarm_level"] = "0" // EMERG (默认)
		} else {
			settings["situp_alarm_level"] = "disabled"
		}
	}

	// 在床配置
	if onbedDuration, ok := hardwareSettings["onbedDuration"].(float64); ok {
		settings["onbed_duration"] = int(onbedDuration)
	}
	if onbedFlag, ok := hardwareSettings["onbedFlag"].(float64); ok {
		if onbedFlag == 1 {
			settings["onbed_alarm_level"] = "0" // EMERG (默认)
		} else {
			settings["onbed_alarm_level"] = "disabled"
		}
	}

	// 传感器跌落配置：返回 fallFlag (boolean)
	if fallFlag, ok := hardwareSettings["fallFlag"].(float64); ok {
		settings["fallFlag"] = fallFlag == 1
	} else {
		settings["fallFlag"] = false
	}

	return &GetDeviceMonitorSettingsResponse{
		Settings: settings,
	}, nil
}

// updateSleepaceSettingsToHardware 将配置同步到 Sleepace 硬件（参考 v1.0 实现）
// 参考：wisefido-backend/wisefido-sleepace/modules/sleepace_service.go::SetDeviceMonitorSettings
func (s *deviceMonitorSettingsService) updateSleepaceSettingsToHardware(ctx context.Context, device *domain.Device, deviceID string, settings map[string]interface{}) error {
	// 获取设备代码（deviceCode），优先使用 serial_number，其次使用 uid
	deviceCode := ""
	if device.SerialNumber.Valid && device.SerialNumber.String != "" {
		deviceCode = device.SerialNumber.String
	} else if device.UID.Valid && device.UID.String != "" {
		deviceCode = device.UID.String
	} else {
		return fmt.Errorf("device has no serial_number or uid")
	}

	// 将 flat settings 转换为 SleepaceMonitorSettings 格式（用于硬件 API）
	hardwareSettings := s.convertFlatSettingsToSleepaceFormat(deviceID, deviceCode, settings)

	// 调用 Sleepace API 同步到硬件
	request := struct {
		Token *SleepaceToken         `json:"token"`
		Data  map[string]interface{} `json:"data"`
	}{
		Token: s.sleepaceClient.token,
		Data:  hardwareSettings,
	}

	var response SleepaceResponse
	resp, err := s.sleepaceClient.httpClient.R().
		SetBody(request).
		SetResult(&response).
		Post("/sleepace/updatealarmnotifyconfig")

	if err != nil {
		return fmt.Errorf("failed to call Sleepace API: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("Sleepace API returned status code: %d", resp.StatusCode())
	}

	if response.Status != 0 {
		return fmt.Errorf("Sleepace API error: %s (status: %d)", response.Msg, response.Status)
	}

	return nil
}

// convertFlatSettingsToSleepaceFormat 将 flat settings 转换为 SleepaceMonitorSettings 格式（用于硬件 API）
// 参考：wisefido-backend/wisefido-sleepace/models/settings.go::SleepaceMonitorSettings
func (s *deviceMonitorSettingsService) convertFlatSettingsToSleepaceFormat(deviceID, deviceCode string, settings map[string]interface{}) map[string]interface{} {
	hardwareSettings := make(map[string]interface{})

	// 基本字段
	hardwareSettings["userId"] = deviceID
	hardwareSettings["deviceId"] = deviceCode

	// 离床时间配置
	if leftBedStartHour, ok := settings["left_bed_start_hour"]; ok {
		if hour, ok := leftBedStartHour.(float64); ok {
			hardwareSettings["leftBedStartHour"] = int(hour)
		} else if hour, ok := leftBedStartHour.(int); ok {
			hardwareSettings["leftBedStartHour"] = hour
		}
	}
	if leftBedStartMinute, ok := settings["left_bed_start_minute"]; ok {
		if minute, ok := leftBedStartMinute.(float64); ok {
			hardwareSettings["leftBedStartMinute"] = int(minute)
		} else if minute, ok := leftBedStartMinute.(int); ok {
			hardwareSettings["leftBedStartMinute"] = minute
		}
	}
	if leftBedEndHour, ok := settings["left_bed_end_hour"]; ok {
		if hour, ok := leftBedEndHour.(float64); ok {
			hardwareSettings["leftBedEndHour"] = int(hour)
		} else if hour, ok := leftBedEndHour.(int); ok {
			hardwareSettings["leftBedEndHour"] = hour
		}
	}
	if leftBedEndMinute, ok := settings["left_bed_end_minute"]; ok {
		if minute, ok := leftBedEndMinute.(float64); ok {
			hardwareSettings["leftBedEndMinute"] = int(minute)
		} else if minute, ok := leftBedEndMinute.(int); ok {
			hardwareSettings["leftBedEndMinute"] = minute
		}
	}
	if leftBedDuration, ok := settings["left_bed_duration"]; ok {
		if duration, ok := leftBedDuration.(float64); ok {
			hardwareSettings["leftBedDuration"] = int(duration)
		} else if duration, ok := leftBedDuration.(int); ok {
			hardwareSettings["leftBedDuration"] = duration
		}
	}
	// leftBedFlag: 1 = enabled, 0 = disabled
	if leftBedLevel, ok := settings["left_bed_alarm_level"].(string); ok {
		if leftBedLevel != "disabled" && leftBedLevel != "" {
			hardwareSettings["leftBedFlag"] = 1
		} else {
			hardwareSettings["leftBedFlag"] = 0
		}
	}

	// 心率配置
	if minHeartRate, ok := settings["min_heart_rate"]; ok {
		if rate, ok := minHeartRate.(float64); ok {
			hardwareSettings["minHeartRate"] = int(rate)
		} else if rate, ok := minHeartRate.(int); ok {
			hardwareSettings["minHeartRate"] = rate
		}
	}
	if maxHeartRate, ok := settings["max_heart_rate"]; ok {
		if rate, ok := maxHeartRate.(float64); ok {
			hardwareSettings["maxHeartRate"] = int(rate)
		} else if rate, ok := maxHeartRate.(int); ok {
			hardwareSettings["maxHeartRate"] = rate
		}
	}
	if heartRateSlowDuration, ok := settings["heart_rate_slow_duration"]; ok {
		if duration, ok := heartRateSlowDuration.(float64); ok {
			hardwareSettings["heartRateSlowDuration"] = int(duration)
		} else if duration, ok := heartRateSlowDuration.(int); ok {
			hardwareSettings["heartRateSlowDuration"] = duration
		}
	}
	if heartRateFastDuration, ok := settings["heart_rate_fast_duration"]; ok {
		if duration, ok := heartRateFastDuration.(float64); ok {
			hardwareSettings["heartRateFastDuration"] = int(duration)
		} else if duration, ok := heartRateFastDuration.(int); ok {
			hardwareSettings["heartRateFastDuration"] = duration
		}
	}
	if heartRateSlowLevel, ok := settings["heart_rate_slow_alarm_level"].(string); ok {
		if heartRateSlowLevel != "disabled" && heartRateSlowLevel != "" {
			hardwareSettings["heartRateSlowFlag"] = 1
		} else {
			hardwareSettings["heartRateSlowFlag"] = 0
		}
	}
	if heartRateFastLevel, ok := settings["heart_rate_fast_alarm_level"].(string); ok {
		if heartRateFastLevel != "disabled" && heartRateFastLevel != "" {
			hardwareSettings["heartRateFastFlag"] = 1
		} else {
			hardwareSettings["heartRateFastFlag"] = 0
		}
	}

	// 呼吸率配置
	if minBreathRate, ok := settings["min_breath_rate"]; ok {
		if rate, ok := minBreathRate.(float64); ok {
			hardwareSettings["minBreathRate"] = int(rate)
		} else if rate, ok := minBreathRate.(int); ok {
			hardwareSettings["minBreathRate"] = rate
		}
	}
	if maxBreathRate, ok := settings["max_breath_rate"]; ok {
		if rate, ok := maxBreathRate.(float64); ok {
			hardwareSettings["maxBreathRate"] = int(rate)
		} else if rate, ok := maxBreathRate.(int); ok {
			hardwareSettings["maxBreathRate"] = rate
		}
	}
	if breathRateSlowDuration, ok := settings["breath_rate_slow_duration"]; ok {
		if duration, ok := breathRateSlowDuration.(float64); ok {
			hardwareSettings["breathRateSlowDuration"] = int(duration)
		} else if duration, ok := breathRateSlowDuration.(int); ok {
			hardwareSettings["breathRateSlowDuration"] = duration
		}
	}
	if breathRateFastDuration, ok := settings["breath_rate_fast_duration"]; ok {
		if duration, ok := breathRateFastDuration.(float64); ok {
			hardwareSettings["breathRateFastDuration"] = int(duration)
		} else if duration, ok := breathRateFastDuration.(int); ok {
			hardwareSettings["breathRateFastDuration"] = duration
		}
	}
	if breathRateSlowLevel, ok := settings["breath_rate_slow_alarm_level"].(string); ok {
		if breathRateSlowLevel != "disabled" && breathRateSlowLevel != "" {
			hardwareSettings["breathRateSlowFlag"] = 1
		} else {
			hardwareSettings["breathRateSlowFlag"] = 0
		}
	}
	if breathRateFastLevel, ok := settings["breath_rate_fast_alarm_level"].(string); ok {
		if breathRateFastLevel != "disabled" && breathRateFastLevel != "" {
			hardwareSettings["breathRateFastFlag"] = 1
		} else {
			hardwareSettings["breathRateFastFlag"] = 0
		}
	}

	// 呼吸暂停配置
	if breathPauseDuration, ok := settings["breath_pause_duration"]; ok {
		if duration, ok := breathPauseDuration.(float64); ok {
			hardwareSettings["breathPauseDuration"] = int(duration)
		} else if duration, ok := breathPauseDuration.(int); ok {
			hardwareSettings["breathPauseDuration"] = duration
		}
	}
	if breathPauseLevel, ok := settings["breath_pause_alarm_level"].(string); ok {
		if breathPauseLevel != "disabled" && breathPauseLevel != "" {
			hardwareSettings["breathPauseFlag"] = 1
		} else {
			hardwareSettings["breathPauseFlag"] = 0
		}
	}

	// 身体移动配置
	if bodyMoveDuration, ok := settings["body_move_duration"]; ok {
		if duration, ok := bodyMoveDuration.(float64); ok {
			hardwareSettings["bodyMoveDuration"] = int(duration)
		} else if duration, ok := bodyMoveDuration.(int); ok {
			hardwareSettings["bodyMoveDuration"] = duration
		}
	}
	if bodyMoveLevel, ok := settings["body_move_alarm_level"].(string); ok {
		if bodyMoveLevel != "disabled" && bodyMoveLevel != "" {
			hardwareSettings["bodyMoveFlag"] = 1
		} else {
			hardwareSettings["bodyMoveFlag"] = 0
		}
	}

	// 无身体移动配置
	if nobodyMoveDuration, ok := settings["nobody_move_duration"]; ok {
		if duration, ok := nobodyMoveDuration.(float64); ok {
			hardwareSettings["nobodyMoveDuration"] = int(duration)
		} else if duration, ok := nobodyMoveDuration.(int); ok {
			hardwareSettings["nobodyMoveDuration"] = duration
		}
	}
	if nobodyMoveLevel, ok := settings["nobody_move_alarm_level"].(string); ok {
		if nobodyMoveLevel != "disabled" && nobodyMoveLevel != "" {
			hardwareSettings["nobodyMoveFlag"] = 1
		} else {
			hardwareSettings["nobodyMoveFlag"] = 0
		}
	}

	// 无翻身配置
	if noTurnOverDuration, ok := settings["no_turn_over_duration"]; ok {
		if duration, ok := noTurnOverDuration.(float64); ok {
			hardwareSettings["noTurnOverDuration"] = int(duration)
		} else if duration, ok := noTurnOverDuration.(int); ok {
			hardwareSettings["noTurnOverDuration"] = duration
		}
	}
	if noTurnOverLevel, ok := settings["no_turn_over_alarm_level"].(string); ok {
		if noTurnOverLevel != "disabled" && noTurnOverLevel != "" {
			hardwareSettings["noTurnOverFlag"] = 1
		} else {
			hardwareSettings["noTurnOverFlag"] = 0
		}
	}

	// 坐起配置
	if situpLevel, ok := settings["situp_alarm_level"].(string); ok {
		if situpLevel != "disabled" && situpLevel != "" {
			hardwareSettings["situpFlag"] = 1
		} else {
			hardwareSettings["situpFlag"] = 0
		}
	}

	// 在床配置
	if onbedDuration, ok := settings["onbed_duration"]; ok {
		if duration, ok := onbedDuration.(float64); ok {
			hardwareSettings["onbedDuration"] = int(duration)
		} else if duration, ok := onbedDuration.(int); ok {
			hardwareSettings["onbedDuration"] = duration
		}
	}
	if onbedLevel, ok := settings["onbed_alarm_level"].(string); ok {
		if onbedLevel != "disabled" && onbedLevel != "" {
			hardwareSettings["onbedFlag"] = 1
		} else {
			hardwareSettings["onbedFlag"] = 0
		}
	}

	// 传感器跌落配置：前端发送 fallFlag (boolean)
	if fallFlag, ok := settings["fallFlag"].(bool); ok {
		if fallFlag {
			hardwareSettings["fallFlag"] = 1
		} else {
			hardwareSettings["fallFlag"] = 0
		}
	}

	return hardwareSettings
}
