package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"owl-common/alarm"

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
	alarmDeviceRepo    repository.AlarmDeviceRepository
	alarmCloudRepo     repository.AlarmCloudRepository
	configVersionsRepo repository.ConfigVersionsRepository // 配置版本仓库（用于审计）
	devicesRepo        repository.DevicesRepository
	deviceStoreRepo    repository.DeviceStoreRepository
	sleepaceClient     *SleepaceClient          // Sleepace 硬件 API 客户端（可选）
	configNotifier     *notifier.ConfigNotifier // 配置变更通知器
	qinglanClient      *QinglanClient          // 下发雷达属性（工作模式、跌倒/呼吸心率），不重启
	db                 *sql.DB                  // 用于事务操作
	logger             *zap.Logger
}

// NewDeviceMonitorSettingsService 创建设备监控配置服务实例
func NewDeviceMonitorSettingsService(
	alarmDeviceRepo repository.AlarmDeviceRepository,
	alarmCloudRepo repository.AlarmCloudRepository,
	configVersionsRepo repository.ConfigVersionsRepository,
	devicesRepo repository.DevicesRepository,
	deviceStoreRepo repository.DeviceStoreRepository,
	db *sql.DB,
	configNotifier *notifier.ConfigNotifier,
	logger *zap.Logger,
) DeviceMonitorSettingsService {
	return &deviceMonitorSettingsService{
		alarmDeviceRepo:    alarmDeviceRepo,
		alarmCloudRepo:     alarmCloudRepo,
		configVersionsRepo: configVersionsRepo,
		devicesRepo:        devicesRepo,
		deviceStoreRepo:    deviceStoreRepo,
		db:                 db,
		sleepaceClient:     nil, // 通过 SetSleepaceClient 延迟设置
		configNotifier:     configNotifier,
		logger:             logger,
	}
}

// SetSleepaceClient 设置 Sleepace 客户端（延迟初始化，避免循环依赖）
func (s *deviceMonitorSettingsService) SetSleepaceClient(client *SleepaceClient) {
	s.sleepaceClient = client
}

// SetQinglanClient 设置 Qinglan 客户端（下发雷达工作模式、跌倒/呼吸心率参数，不重启）
func (s *deviceMonitorSettingsService) SetQinglanClient(client *QinglanClient) {
	s.qinglanClient = client
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
// 返回统一的 DefaultAlarmSetting 格式（AlarmItem 数组）
type GetDeviceMonitorSettingsResponse struct {
	AlarmItems []alarm.AlarmItem `json:"alarm_items"` // 报警项列表（DefaultAlarmSetting 格式）
}

// UpdateDeviceMonitorSettingsRequest 更新设备监控配置请求
type UpdateDeviceMonitorSettingsRequest struct {
	TenantID   string            // 租户ID
	DeviceID   string            // 设备ID
	DeviceType string            // 设备类型：'sleepace' 或 'radar'
	UserID     string            // 当前用户ID（用于审计）
	AlarmItems []alarm.AlarmItem // 报警项列表（DefaultAlarmSetting 格式）
}

// UpdateDeviceMonitorSettingsResponse 更新设备监控配置响应
type UpdateDeviceMonitorSettingsResponse struct {
	Success bool `json:"success"` // 是否成功
}

// ============================================
// Service 方法实现
// ============================================

// buildDeviceAlarmItemsFromCloudOrDefault 从 alarm_cloud 或 DefaultAlarmSetting 构建设备的 AlarmItem 数组
// deviceType: 数据库中的 device_type（如 "Sleepad", "radar"），直接使用 device_store.device_type
func (s *deviceMonitorSettingsService) buildDeviceAlarmItemsFromCloudOrDefault(ctx context.Context, tenantID, deviceType string) ([]alarm.AlarmItem, error) {
	// 标准化 device_type（统一转换为小写）
	deviceTypeLower := strings.ToLower(deviceType)

	// 判断是 Sleepad 还是 Radar
	isSleepad := deviceTypeLower == "sleepace" || deviceTypeLower == "sleepad" || deviceTypeLower == "sleeppad"
	isRadar := deviceTypeLower == "radar"

	if !isSleepad && !isRadar {
		return nil, fmt.Errorf("invalid device_type: %s (must be Sleepad/sleepace/sleepad or radar)", deviceType)
	}

	// 1. 尝试从 alarm_cloud 获取配置
	alarmCloudConfig, err := s.getAlarmCloudConfig(ctx, tenantID)
	if err == nil && alarmCloudConfig != nil {
		// 从 alarm_cloud 获取对应设备类型的 AlarmItems
		if isSleepad {
			return alarmCloudConfig.AlarmSetting.Sleepad, nil
		} else if isRadar {
			return alarmCloudConfig.AlarmSetting.Radar, nil
		}
	}

	// 2. 如果 alarm_cloud 不存在，使用 DefaultAlarmSetting
	if isSleepad {
		return alarm.DefaultAlarmSetting.Sleepad, nil
	} else if isRadar {
		return alarm.DefaultAlarmSetting.Radar, nil
	}

	return nil, fmt.Errorf("invalid device_type: %s", deviceType)
}

// getAlarmCloudConfig 获取 alarm_cloud 配置（简化版本，不进行权限检查）
func (s *deviceMonitorSettingsService) getAlarmCloudConfig(ctx context.Context, tenantID string) (*alarm.AlarmCloudConfig, error) {
	alarmCloud, err := s.alarmCloudRepo.GetAlarmCloud(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// 从 domain.AlarmCloud 构建 alarm.AlarmCloudConfig
	// 这里简化处理，直接使用 buildAlarmCloudConfigFromDomain 的逻辑
	config := &alarm.AlarmCloudConfig{
		TenantResetTime: alarm.DefaultTenantResetTime,
		CommonAlarm: alarm.CommonAlarm{
			OfflineAlarm:  alarm.DefaultOfflineAlarm,
			LowBattery:    alarm.DefaultLowBattery,
			DeviceFailure: alarm.DefaultDeviceFailure,
		},
		AlarmSetting: struct {
			Sleepad []alarm.AlarmItem `json:"sleepad,omitempty"`
			Radar   []alarm.AlarmItem `json:"radar,omitempty"`
		}{
			Sleepad: make([]alarm.AlarmItem, 0),
			Radar:   make([]alarm.AlarmItem, 0),
		},
		CloudVitalAlarmThreshold: alarm.DefaultCloudVitalAlarmThreshold,
	}

	// 设置 CommonAlarm
	if alarmCloud.OfflineAlarm != "" {
		config.CommonAlarm.OfflineAlarm = alarmCloud.OfflineAlarm
	}
	if alarmCloud.LowBattery != "" {
		config.CommonAlarm.LowBattery = alarmCloud.LowBattery
	}
	if alarmCloud.DeviceFailure != "" {
		config.CommonAlarm.DeviceFailure = alarmCloud.DeviceFailure
	}

	// 解析 device_alarms
	var deviceAlarmsMap map[string]map[string]string
	if len(alarmCloud.DeviceAlarms) > 0 {
		if err := json.Unmarshal(alarmCloud.DeviceAlarms, &deviceAlarmsMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal device_alarms: %w", err)
		}
	}

	// 处理 SleepPad
	for _, item := range alarm.DefaultAlarmSetting.Sleepad {
		newItem := item
		if sleepPadMap, ok := deviceAlarmsMap["SleepPad"]; ok {
			if level, exists := sleepPadMap[item.AlarmType]; exists {
				newItem.AlarmLevel = &level
				if level == "DISABLE" || level == "DISABLED" {
					enabled := alarm.IsEnabledOff
					newItem.IsEnabled = &enabled
				} else {
					enabled := alarm.IsEnabledOn
					newItem.IsEnabled = &enabled
				}
			}
		}
		config.AlarmSetting.Sleepad = append(config.AlarmSetting.Sleepad, newItem)
	}

	// 处理 Radar
	for _, item := range alarm.DefaultAlarmSetting.Radar {
		newItem := item
		if radarMap, ok := deviceAlarmsMap["Radar"]; ok {
			if level, exists := radarMap[item.AlarmType]; exists {
				newItem.AlarmLevel = &level
				if level == "DISABLE" || level == "DISABLED" {
					enabled := alarm.IsEnabledOff
					newItem.IsEnabled = &enabled
				} else {
					enabled := alarm.IsEnabledOn
					newItem.IsEnabled = &enabled
				}
			}
		}
		config.AlarmSetting.Radar = append(config.AlarmSetting.Radar, newItem)
	}

	return config, nil
}

// convertAlarmItemsToMonitorConfig 将 AlarmItem 数组转换为 monitor_config JSONB 格式
func (s *deviceMonitorSettingsService) convertAlarmItemsToMonitorConfig(alarmItems []alarm.AlarmItem) ([]byte, error) {
	// monitor_config 格式：{"alarms": {"AlarmType": {"level": "...", "enabled": true/false, "threshold": {...}}}}
	monitorConfig := make(map[string]interface{})
	alarms := make(map[string]interface{})

	for _, item := range alarmItems {
		alarmConfig := make(map[string]interface{})

		// level
		if item.AlarmLevel != nil {
			alarmConfig["level"] = *item.AlarmLevel
		}

		// enabled
		if item.IsEnabled != nil {
			alarmConfig["enabled"] = *item.IsEnabled == alarm.IsEnabledOn
		} else {
			alarmConfig["enabled"] = false
		}

		// threshold (从 alarm_params 转换)
		if item.AlarmParams != nil && len(item.AlarmParams) > 0 {
			alarmConfig["threshold"] = item.AlarmParams
		}

		alarms[item.AlarmType] = alarmConfig
	}

	monitorConfig["alarms"] = alarms
	return json.Marshal(monitorConfig)
}

// convertMonitorConfigToAlarmItems 将 monitor_config JSONB 转换为 AlarmItem 数组
func (s *deviceMonitorSettingsService) convertMonitorConfigToAlarmItems(monitorConfigJSON []byte, deviceType string) ([]alarm.AlarmItem, error) {
	// 解析 monitor_config
	var monitorConfig map[string]interface{}
	if len(monitorConfigJSON) > 0 {
		if err := json.Unmarshal(monitorConfigJSON, &monitorConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal monitor_config: %w", err)
		}
	}

	// 获取默认 AlarmItems（根据 device_store.device_type 直接判断）
	deviceTypeLower := strings.ToLower(deviceType)
	isSleepad := deviceTypeLower == "sleepace" || deviceTypeLower == "sleepad" || deviceTypeLower == "sleeppad"
	isRadar := deviceTypeLower == "radar"

	var defaultItems []alarm.AlarmItem
	if isSleepad {
		defaultItems = alarm.DefaultAlarmSetting.Sleepad
	} else if isRadar {
		defaultItems = alarm.DefaultAlarmSetting.Radar
	} else {
		return nil, fmt.Errorf("invalid device_type: %s (must be Sleepad/sleepace/sleepad or radar)", deviceType)
	}

	// 从 monitor_config 构建映射
	alarmsMap := make(map[string]map[string]interface{})
	if alarms, ok := monitorConfig["alarms"].(map[string]interface{}); ok {
		for k, v := range alarms {
			if alarmConfig, ok := v.(map[string]interface{}); ok {
				alarmsMap[k] = alarmConfig
			}
		}
	}

	// 合并默认值和 monitor_config
	result := make([]alarm.AlarmItem, 0, len(defaultItems))
	for _, defaultItem := range defaultItems {
		newItem := defaultItem

		if alarmConfig, exists := alarmsMap[defaultItem.AlarmType]; exists {
			// 更新 level
			if level, ok := alarmConfig["level"].(string); ok {
				newItem.AlarmLevel = &level
			}

			// 更新 enabled
			if enabled, ok := alarmConfig["enabled"].(bool); ok {
				if enabled {
					enabledVal := alarm.IsEnabledOn
					newItem.IsEnabled = &enabledVal
				} else {
					enabledVal := alarm.IsEnabledOff
					newItem.IsEnabled = &enabledVal
				}
			}

			// 更新 threshold (alarm_params)
			if threshold, ok := alarmConfig["threshold"].(map[string]interface{}); ok {
				newItem.AlarmParams = threshold
			}
		}

		result = append(result, newItem)
	}

	return result, nil
}

// GetDeviceMonitorSettings 获取设备监控配置（简化版本）
func (s *deviceMonitorSettingsService) GetDeviceMonitorSettings(ctx context.Context, req GetDeviceMonitorSettingsRequest) (*GetDeviceMonitorSettingsResponse, error) {
	// 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.DeviceID == "" || req.DeviceID == "undefined" || req.DeviceID == "null" {
		return nil, fmt.Errorf("device_id is required and must be a valid UUID")
	}
	if req.DeviceType != "sleepace" && req.DeviceType != "radar" {
		return nil, fmt.Errorf("invalid device_type: %s (must be 'sleepace' or 'radar')", req.DeviceType)
	}

	// 验证设备存在
	device, err := s.devicesRepo.GetDevice(ctx, req.TenantID, req.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	// 验证设备类型匹配（安全性检查：防止恶意修改 URL 参数）
	// req.DeviceType 是前端传入的 "sleepace" 或 "radar"
	// device.DeviceType 是从 devices 表 JOIN device_store 获取的 device_type（如 "Sleepad", "radar"）
	if !device.DeviceType.Valid {
		return nil, fmt.Errorf("device has no device_type")
	}

	reqTypeLower := strings.ToLower(req.DeviceType)
	deviceTypeLower := strings.ToLower(device.DeviceType.String)

	// 判断是否匹配
	reqIsSleepad := reqTypeLower == "sleepace"
	reqIsRadar := reqTypeLower == "radar"
	dbIsSleepad := deviceTypeLower == "sleepace" || deviceTypeLower == "sleepad" || deviceTypeLower == "sleeppad"
	dbIsRadar := deviceTypeLower == "radar"

	if !((reqIsSleepad && dbIsSleepad) || (reqIsRadar && dbIsRadar)) {
		return nil, fmt.Errorf("device type mismatch: expected %s (from request), got %s (from device_store)", req.DeviceType, device.DeviceType.String)
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

	// 1. 获取设备的监控配置（从数据库）
	alarmDevice, err := s.alarmDeviceRepo.GetAlarmDevice(ctx, req.TenantID, req.DeviceID)
	if err != nil {
		// 如果配置不存在，从 alarm_cloud 或 DefaultAlarmSetting 构建
		// 直接使用 req.DeviceType（前端已经知道是 radar 还是 sleepace）
		alarmItems, buildErr := s.buildDeviceAlarmItemsFromCloudOrDefault(ctx, req.TenantID, req.DeviceType)
		if buildErr != nil {
			return nil, fmt.Errorf("failed to build default alarm items: %w", buildErr)
		}

		// 转换为 monitor_config 格式并写入数据库
		monitorConfigJSON, convertErr := s.convertAlarmItemsToMonitorConfig(alarmItems)
		if convertErr != nil {
			return nil, fmt.Errorf("failed to convert alarm items to monitor config: %w", convertErr)
		}

		// 创建新记录
		newAlarmDevice := &domain.AlarmDevice{
			DeviceID:      req.DeviceID,
			TenantID:      req.TenantID,
			MonitorConfig: monitorConfigJSON,
		}

		// 使用事务写入数据库（设置 created_by）
		if s.db != nil {
			tx, txErr := s.db.BeginTx(ctx, nil)
			if txErr == nil {
				defer tx.Rollback()

				query := `
					INSERT INTO alarm_device (
						device_id,
						tenant_id,
						monitor_config
					) VALUES ($1, $2, $3::jsonb)
				`

				_, execErr := tx.ExecContext(ctx, query, req.DeviceID, req.TenantID, string(monitorConfigJSON))
				if execErr == nil {
					if commitErr := tx.Commit(); commitErr == nil {
						// 保存成功，重新读取记录
						alarmDevice, readErr := s.alarmDeviceRepo.GetAlarmDevice(ctx, req.TenantID, req.DeviceID)
						if readErr == nil {
							// 转换为 AlarmItem 数组并返回
							// 直接使用 req.DeviceType（前端已经知道是 radar 还是 sleepace）
							alarmItems, convertErr := s.convertMonitorConfigToAlarmItems(alarmDevice.MonitorConfig, req.DeviceType)
							if convertErr != nil {
								return nil, fmt.Errorf("failed to convert monitor config to alarm items: %w", convertErr)
							}
							return &GetDeviceMonitorSettingsResponse{
								AlarmItems: alarmItems,
							}, nil
						}
					}
				}
				// 如果事务保存失败，继续使用 UpsertAlarmDevice 作为后备方案
			}
		}

		// 后备方案：使用 UpsertAlarmDevice（无法设置 created_by）
		if saveErr := s.alarmDeviceRepo.UpsertAlarmDevice(ctx, req.TenantID, req.DeviceID, newAlarmDevice); saveErr != nil {
			s.logger.Warn("Failed to save default alarm device config",
				zap.String("tenant_id", req.TenantID),
				zap.String("device_id", req.DeviceID),
				zap.Error(saveErr),
			)
			// 保存失败不影响返回，继续返回默认配置
		}

		// 返回默认配置
		return &GetDeviceMonitorSettingsResponse{
			AlarmItems: alarmItems,
		}, nil
	}

	// 2. 如果配置存在，转换为 AlarmItem 数组
	// 直接使用 req.DeviceType（前端已经知道是 radar 还是 sleepace）
	alarmItems, err := s.convertMonitorConfigToAlarmItems(alarmDevice.MonitorConfig, req.DeviceType)
	if err != nil {
		return nil, fmt.Errorf("failed to convert monitor config to alarm items: %w", err)
	}

	return &GetDeviceMonitorSettingsResponse{
		AlarmItems: alarmItems,
	}, nil
}

// UpdateDeviceMonitorSettings 更新设备监控配置（简化版本，使用事务）
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
	if req.AlarmItems == nil || len(req.AlarmItems) == 0 {
		return nil, fmt.Errorf("alarm_items is required")
	}

	// 验证设备存在
	device, err := s.devicesRepo.GetDevice(ctx, req.TenantID, req.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	// 验证设备类型匹配（安全性检查：防止恶意修改 URL 参数）
	// req.DeviceType 是前端传入的 "sleepace" 或 "radar"
	// device.DeviceType 是从 devices 表 JOIN device_store 获取的 device_type（如 "Sleepad", "radar"）
	if !device.DeviceType.Valid {
		return nil, fmt.Errorf("device has no device_type")
	}

	reqTypeLower := strings.ToLower(req.DeviceType)
	deviceTypeLower := strings.ToLower(device.DeviceType.String)

	// 判断是否匹配
	reqIsSleepad := reqTypeLower == "sleepace"
	reqIsRadar := reqTypeLower == "radar"
	dbIsSleepad := deviceTypeLower == "sleepace" || deviceTypeLower == "sleepad" || deviceTypeLower == "sleeppad"
	dbIsRadar := deviceTypeLower == "radar"

	if !((reqIsSleepad && dbIsSleepad) || (reqIsRadar && dbIsRadar)) {
		return nil, fmt.Errorf("device type mismatch: expected %s (from request), got %s (from device_store)", req.DeviceType, device.DeviceType.String)
	}

	// 1. 获取现有配置（如果存在），用于比对和保存到 config_versions
	var existingAlarmItems []alarm.AlarmItem
	existingAlarmDevice, err := s.alarmDeviceRepo.GetAlarmDevice(ctx, req.TenantID, req.DeviceID)
	if err != nil {
		// 如果不存在，从 alarm_cloud 或 DefaultAlarmSetting 构建
		// 直接使用 req.DeviceType（前端已经知道是 radar 还是 sleepace）
		existingAlarmItems, err = s.buildDeviceAlarmItemsFromCloudOrDefault(ctx, req.TenantID, req.DeviceType)
		if err != nil {
			return nil, fmt.Errorf("failed to build default alarm items: %w", err)
		}
	} else {
		// 转换为 AlarmItem 数组
		// 直接使用 req.DeviceType（前端已经知道是 radar 还是 sleepace）
		existingAlarmItems, err = s.convertMonitorConfigToAlarmItems(existingAlarmDevice.MonitorConfig, req.DeviceType)
		if err != nil {
			return nil, fmt.Errorf("failed to convert existing monitor config: %w", err)
		}
	}

	// 2. 比对配置是否有变化（简单比对 JSON 字符串）
	existingJSON, err := json.Marshal(existingAlarmItems)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal existing alarm items: %w", err)
	}
	newJSON, err := json.Marshal(req.AlarmItems)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal new alarm items: %w", err)
	}

	hasChanges := string(existingJSON) != string(newJSON)

	// 3. 将 AlarmItem 数组转换为 monitor_config JSONB
	monitorConfigJSON, err := s.convertAlarmItemsToMonitorConfig(req.AlarmItems)
	if err != nil {
		return nil, fmt.Errorf("failed to convert alarm items to monitor config: %w", err)
	}

	// 4. 使用事务保证原子性：要么全部成功，要么全部失败
	if s.db == nil {
		return nil, fmt.Errorf("database connection is required for transaction")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // 确保在函数返回时回滚（如果未提交）

	// 4.1. 如果有变化，先保存旧配置到 config_versions（在事务中）
	// 保存完整配置：包括 alarm_cloud 配置和设备特定的 monitor_config
	if hasChanges && s.configVersionsRepo != nil {
		// 获取 alarm_cloud 配置（用于合并生成完整配置）
		alarmCloudConfig, err := s.getAlarmCloudConfig(ctx, req.TenantID)
		if err != nil {
			// 如果获取失败，构建默认配置
			alarmCloudConfig = &alarm.AlarmCloudConfig{
				TenantResetTime: alarm.DefaultTenantResetTime,
				CommonAlarm: alarm.CommonAlarm{
					OfflineAlarm:  alarm.DefaultOfflineAlarm,
					LowBattery:    alarm.DefaultLowBattery,
					DeviceFailure: alarm.DefaultDeviceFailure,
				},
				AlarmSetting: struct {
					Sleepad []alarm.AlarmItem `json:"sleepad,omitempty"`
					Radar   []alarm.AlarmItem `json:"radar,omitempty"`
				}{
					Sleepad: alarm.DefaultAlarmSetting.Sleepad,
					Radar:   alarm.DefaultAlarmSetting.Radar,
				},
				CloudVitalAlarmThreshold: alarm.DefaultCloudVitalAlarmThreshold,
			}
		}

		// 保存前端修改的 []AlarmItem（即 monitor_config 中的内容）
		// 前端修改的是 DefaultAlarmSetting 中的 Radar 或 Sleepad []AlarmItem
		// 这些数据保存在 monitor_config 中，需要保存到 config_versions
		// 使用 existingAlarmItems（已在第 538-555 行从 monitor_config 转换，或从 alarm_cloud/默认值构建）
		alarmItems := existingAlarmItems

		// 处理 vendor_config：从 existingAlarmDevice 获取
		var vendorConfig map[string]interface{}
		if existingAlarmDevice != nil && len(existingAlarmDevice.VendorConfig) > 0 {
			if err := json.Unmarshal(existingAlarmDevice.VendorConfig, &vendorConfig); err != nil {
				vendorConfig = make(map[string]interface{})
			}
		} else {
			vendorConfig = make(map[string]interface{})
		}

		// 处理 metadata：从 existingAlarmDevice 获取
		var metadata map[string]interface{}
		if existingAlarmDevice != nil && len(existingAlarmDevice.Metadata) > 0 {
			if err := json.Unmarshal(existingAlarmDevice.Metadata, &metadata); err != nil {
				metadata = make(map[string]interface{})
			}
		} else {
			metadata = make(map[string]interface{})
		}

		// 构建完整的配置数据：保存所有字段
		// 包括 alarm_cloud 的完整配置和 alarm_device 的完整配置（前端修改的 []AlarmItem 在 monitor_config 中）
		configData := map[string]interface{}{
			// alarm_cloud 配置（完整保存）
			"TenantResetTime":          alarmCloudConfig.TenantResetTime,
			"common_alarm":             alarmCloudConfig.CommonAlarm,
			"AlarmSetting":             alarmCloudConfig.AlarmSetting,
			"CloudVitalAlarmThreshold": alarmCloudConfig.CloudVitalAlarmThreshold,
			// alarm_device 的全部字段（前端修改的 []AlarmItem 保存在 monitor_config 中）
			"alarm_items":   alarmItems, // 前端修改的 []AlarmItem（Radar 或 Sleepad）
			"vendor_config": vendorConfig,
			"metadata":      metadata,
		}
		configDataJSON, err := json.Marshal(configData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal config_data for config_version: %w", err)
		}

		configVersion := &domain.ConfigVersion{
			ConfigType:      "alarm_device",
			EntityID:        req.DeviceID, // alarm_device 的 entity_id 就是 device_id
			CurrentEntityID: req.DeviceID,
			ConfigData:      json.RawMessage(configDataJSON),
			ValidFrom:       time.Now().UTC(),
		}
		if req.UserID != "" {
			configVersion.CreatedBy = &req.UserID
		}

		// 在事务中执行 config_versions 的插入
		// 将旧版本的valid_to设置为当前时间
		updateOldQuery := `
			UPDATE config_versions
			SET valid_to = $4
			WHERE tenant_id = $1 
				AND config_type = $2 
				AND entity_id = $3
				AND (valid_to IS NULL OR valid_to > $4)
		`
		_, err = tx.ExecContext(ctx, updateOldQuery, req.TenantID, configVersion.ConfigType, configVersion.EntityID, configVersion.ValidFrom)
		if err != nil {
			return nil, fmt.Errorf("failed to update old config versions: %w", err)
		}

		// 创建新版本
		insertQuery := `
			INSERT INTO config_versions (
				tenant_id,
				config_type,
				entity_id,
				current_entity_id,
				config_data,
				valid_from,
				valid_to,
				created_by
			) VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7, $8)
			RETURNING version_id::text
		`

		var currentEntityID interface{}
		if configVersion.CurrentEntityID != "" {
			currentEntityID = configVersion.CurrentEntityID
		}

		var validTo interface{}
		if configVersion.ValidTo != nil {
			validTo = *configVersion.ValidTo
		}

		var createdBy interface{}
		if configVersion.CreatedBy != nil && *configVersion.CreatedBy != "" {
			createdBy = *configVersion.CreatedBy
		}

		var versionID string
		err = tx.QueryRowContext(ctx, insertQuery, req.TenantID, configVersion.ConfigType, configVersion.EntityID,
			currentEntityID, string(configVersion.ConfigData), configVersion.ValidFrom, validTo, createdBy).Scan(&versionID)
		if err != nil {
			return nil, fmt.Errorf("failed to create config version: %w", err)
		}
	}

	// 4.2. 更新 alarm_device（在事务中）
	upsertQuery := `
		INSERT INTO alarm_device (
			device_id,
			tenant_id,
			monitor_config,
			updated_at,
			updated_by
		) VALUES ($1, $2, $3::jsonb, $4, $5)
		ON CONFLICT (device_id) DO UPDATE SET
			tenant_id = EXCLUDED.tenant_id,
			monitor_config = EXCLUDED.monitor_config,
			updated_at = EXCLUDED.updated_at,
			updated_by = EXCLUDED.updated_by
	`

	var updatedBy interface{}
	if req.UserID != "" {
		updatedBy = req.UserID
	}

	_, err = tx.ExecContext(ctx, upsertQuery, req.DeviceID, req.TenantID, string(monitorConfigJSON), time.Now().UTC(), updatedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to update alarm device: %w", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Info("Device monitor settings updated",
		zap.String("tenant_id", req.TenantID),
		zap.String("device_id", req.DeviceID),
		zap.String("device_type", req.DeviceType),
	)

	// 发布配置变更通知
	if s.configNotifier != nil {
		// 获取 device_uid、device_code 和 device_type（radar 服务需要）
		deviceUID := ""
		if device.DeviceUID != "" {
			deviceUID = device.DeviceUID
		}

		deviceCode := ""
		if device.DeviceCode.Valid && device.DeviceCode.String != "" {
			deviceCode = device.DeviceCode.String
		}

		// 获取设备类型（转换为标准格式）
		deviceType := ""
		if device.DeviceType.Valid {
			deviceTypeStr := device.DeviceType.String
			// 转换为标准格式：Sleepad/sleepad/sleepace/sleeppad -> "Sleepad", radar -> "Radar"
			switch strings.ToLower(deviceTypeStr) {
			case "Sleepad", "sleepad", "sleepace", "sleeppad":
				deviceType = "Sleepad"
			case "radar":
				deviceType = "Radar"
			default:
				deviceType = deviceTypeStr
			}
		}

		// 位置信息可选，可为空（当前不传递，后续如需可添加获取逻辑）
		if err := s.configNotifier.NotifyAlarmDeviceUpdated(ctx, req.TenantID, req.DeviceID, deviceUID, deviceCode, deviceType, nil); err != nil {
			s.logger.Warn("Failed to notify config change, but database update succeeded",
				zap.String("tenant_id", req.TenantID),
				zap.String("device_id", req.DeviceID),
				zap.String("device_uid", deviceUID),
				zap.String("device_code", deviceCode),
				zap.String("device_type", deviceType),
				zap.Error(err),
			)
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

// getDeviceType 已废弃：不再需要单独查询 device_store
// device.DeviceType 已经从 GetDevice 查询中通过 JOIN device_store 获取
// 保留此函数仅用于向后兼容（如果其他地方还在使用）
// TODO: 检查并删除所有调用此函数的地方
func (s *deviceMonitorSettingsService) getDeviceType(ctx context.Context, device *domain.Device) (string, error) {
	if !device.DeviceType.Valid {
		return "", fmt.Errorf("device has no device_type")
	}
	return device.DeviceType.String, nil
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

	// 3. 从 alarm_cloud 合并 reset_time（覆盖硬编码值）
	if err := s.mergeResetTimeFromCloud(ctx, tenantID, deviceType, settings); err != nil {
		s.logger.Warn("Failed to merge reset time from cloud, using hardcoded values",
			zap.String("tenant_id", tenantID),
			zap.String("device_type", deviceType),
			zap.Error(err),
		)
	}

	// 4. 从 alarm_cloud 合并阈值（覆盖硬编码值）
	if err := s.mergeThresholdFromCloud(ctx, tenantID, deviceType, settings); err != nil {
		s.logger.Warn("Failed to merge threshold from cloud, using hardcoded values",
			zap.String("tenant_id", tenantID),
			zap.String("device_type", deviceType),
			zap.Error(err),
		)
	}

	// 转换为 AlarmItem 数组
	alarmItems, err := s.buildDeviceAlarmItemsFromCloudOrDefault(ctx, tenantID, deviceType)
	if err != nil {
		return nil, fmt.Errorf("failed to build alarm items: %w", err)
	}

	return &GetDeviceMonitorSettingsResponse{
		AlarmItems: alarmItems,
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

// mergeResetTimeFromCloud 从 alarm_cloud.metadata 合并 reset_time 到 settings
// 配置优先级：设备配置（已存在） > alarm_cloud.metadata（租户统一） > 默认值（已在 settings 中）
// 如果设备配置中已有 reset_time 相关字段，则不覆盖
func (s *deviceMonitorSettingsService) mergeResetTimeFromCloud(ctx context.Context, tenantID, deviceType string, settings map[string]interface{}) error {
	// 获取当前租户的 alarm_cloud 配置
	alarmCloud, err := s.alarmCloudRepo.GetAlarmCloud(ctx, tenantID)
	if err != nil {
		// 如果租户配置不存在，直接返回（使用 settings 中已有的值）
		return nil
	}

	// 解析 metadata JSONB
	if len(alarmCloud.Metadata) == 0 {
		return nil // 没有 metadata 配置，使用 settings 中已有的值
	}

	var metadata alarm.TenantResetTime
	if err := json.Unmarshal(alarmCloud.Metadata, &metadata); err != nil {
		return fmt.Errorf("failed to parse metadata: %w", err)
	}

	if deviceType == "sleepace" {
		// Sleepace: left_bed_start_hour, left_bed_start_minute, left_bed_end_hour, left_bed_end_minute
		// 如果设备配置中没有这些值，从 alarm_cloud.metadata.reset_time 填充
		_, hasStartHour := settings["left_bed_start_hour"]
		_, hasStartMinute := settings["left_bed_start_minute"]
		_, hasEndHour := settings["left_bed_end_hour"]
		_, hasEndMinute := settings["left_bed_end_minute"]

		if !hasStartHour || !hasStartMinute || !hasEndHour || !hasEndMinute {
			// 解析 reset_time
			if metadata.ResetTime.InBedTime != "" {
				inBedHour, inBedMinute, err := alarm.ParseTimeString(metadata.ResetTime.InBedTime)
				if err == nil {
					if !hasStartHour {
						settings["left_bed_start_hour"] = inBedHour
					}
					if !hasStartMinute {
						settings["left_bed_start_minute"] = inBedMinute
					}
				}
			}
			if metadata.ResetTime.OutBedTime != "" {
				outBedHour, outBedMinute, err := alarm.ParseTimeString(metadata.ResetTime.OutBedTime)
				if err == nil {
					if !hasEndHour {
						settings["left_bed_end_hour"] = outBedHour
					}
					if !hasEndMinute {
						settings["left_bed_end_minute"] = outBedMinute
					}
				}
			}
		}
	} else if deviceType == "radar" {
		// Radar: leave_detection_start_hour, leave_detection_start_minute, leave_detection_end_hour, leave_detection_end_minute
		// 如果设备配置中没有这些值，从 alarm_cloud.metadata.reset_time 填充
		_, hasStartHour := settings["leave_detection_start_hour"]
		_, hasStartMinute := settings["leave_detection_start_minute"]
		_, hasEndHour := settings["leave_detection_end_hour"]
		_, hasEndMinute := settings["leave_detection_end_minute"]

		if !hasStartHour || !hasStartMinute || !hasEndHour || !hasEndMinute {
			// 解析 reset_time
			if metadata.ResetTime.InBedTime != "" {
				inBedHour, inBedMinute, err := alarm.ParseTimeString(metadata.ResetTime.InBedTime)
				if err == nil {
					if !hasStartHour {
						settings["leave_detection_start_hour"] = inBedHour
					}
					if !hasStartMinute {
						settings["leave_detection_start_minute"] = inBedMinute
					}
				}
			}
			if metadata.ResetTime.OutBedTime != "" {
				outBedHour, outBedMinute, err := alarm.ParseTimeString(metadata.ResetTime.OutBedTime)
				if err == nil {
					if !hasEndHour {
						settings["leave_detection_end_hour"] = outBedHour
					}
					if !hasEndMinute {
						settings["leave_detection_end_minute"] = outBedMinute
					}
				}
			}
		}
	}

	return nil
}

// mergeThresholdFromCloud 从 alarm_cloud 合并阈值到 settings
// 配置优先级：设备配置（已存在） > alarm_cloud（租户统一） > 默认值（已在 settings 中）
// 如果设备配置中已有阈值字段，则不覆盖
// 从 AlarmCloudConfig 中获取：
// 1. Heart Rate / Respiratory Rate 阈值：从 CloudVitalAlarmThreshold.Conditions 获取
// 2. VitalSignsWeak 参数：从 AlarmSetting.Radar 中 Radar_VitalsWeak 的 alarm_params 获取
func (s *deviceMonitorSettingsService) mergeThresholdFromCloud(ctx context.Context, tenantID, deviceType string, settings map[string]interface{}) error {
	// 获取当前租户的 alarm_cloud 配置
	alarmCloud, err := s.alarmCloudRepo.GetAlarmCloud(ctx, tenantID)
	if err != nil {
		// 如果租户配置不存在，直接返回（使用 settings 中已有的值）
		return nil
	}

	// 解析 conditions JSONB 获取 CloudVitalAlarmThreshold
	var threshold alarm.CloudVitalAlarmThreshold
	if len(alarmCloud.Conditions) > 0 {
		if err := json.Unmarshal(alarmCloud.Conditions, &threshold); err != nil {
			return fmt.Errorf("failed to parse conditions: %w", err)
		}
	}

	// 解析 device_alarms JSONB 获取 AlarmSetting
	var deviceAlarms map[string]map[string]interface{}
	if len(alarmCloud.DeviceAlarms) > 0 {
		if err := json.Unmarshal(alarmCloud.DeviceAlarms, &deviceAlarms); err != nil {
			return fmt.Errorf("failed to parse device_alarms: %w", err)
		}
	}

	// 从 conditions.Conditions 获取 Heart Rate 和 Respiratory Rate 阈值
	if threshold.Conditions != nil {
		// Heart Rate: 从 Normal ranges[0] 获取 min/max
		if threshold.Conditions.HeartRate != nil && threshold.Conditions.HeartRate.Normal != nil {
			if len(threshold.Conditions.HeartRate.Normal.Ranges) > 0 {
				hrRange := threshold.Conditions.HeartRate.Normal.Ranges[0]
				hrMin := 0
				hrMax := 0
				if hrRange.Min != nil {
					hrMin = *hrRange.Min
				}
				if hrRange.Max != nil {
					hrMax = *hrRange.Max
				}

				if deviceType == "sleepace" {
					if _, exists := settings["min_heart_rate"]; !exists && hrMin > 0 {
						settings["min_heart_rate"] = hrMin
					}
					if _, exists := settings["max_heart_rate"]; !exists && hrMax > 0 {
						settings["max_heart_rate"] = hrMax
					}
				} else if deviceType == "radar" {
					if _, exists := settings["lower_heart_rate"]; !exists && hrMin > 0 {
						settings["lower_heart_rate"] = hrMin
					}
					if _, exists := settings["upper_heart_rate"]; !exists && hrMax > 0 {
						settings["upper_heart_rate"] = hrMax
					}
				}
			}
		}

		// Respiratory Rate: 从 Normal ranges[0] 获取 min/max
		if threshold.Conditions.RespiratoryRate != nil && threshold.Conditions.RespiratoryRate.Normal != nil {
			if len(threshold.Conditions.RespiratoryRate.Normal.Ranges) > 0 {
				rrRange := threshold.Conditions.RespiratoryRate.Normal.Ranges[0]
				rrMin := 0
				rrMax := 0
				if rrRange.Min != nil {
					rrMin = *rrRange.Min
				}
				if rrRange.Max != nil {
					rrMax = *rrRange.Max
				}

				if deviceType == "sleepace" {
					if _, exists := settings["min_breath_rate"]; !exists && rrMin > 0 {
						settings["min_breath_rate"] = rrMin
					}
					if _, exists := settings["max_breath_rate"]; !exists && rrMax > 0 {
						settings["max_breath_rate"] = rrMax
					}
				} else if deviceType == "radar" {
					if _, exists := settings["lower_breath_rate"]; !exists && rrMin > 0 {
						settings["lower_breath_rate"] = rrMin
					}
					if _, exists := settings["upper_breath_rate"]; !exists && rrMax > 0 {
						settings["upper_breath_rate"] = rrMax
					}
				}
			}
		}
	}

	// 从 AlarmSetting.Radar 中 Radar_VitalsWeak 的 alarm_params 获取 weak_vital 参数
	// 需要从 device_alarms 构建完整的 AlarmSetting，然后查找 Radar_VitalsWeak
	if deviceAlarms != nil {
		radarAlarms, ok := deviceAlarms["Radar"]
		if ok {
			// 从 DefaultAlarmSetting 构建完整的 AlarmSetting，查找 Radar_VitalsWeak
			for _, item := range alarm.DefaultAlarmSetting.Radar {
				if item.AlarmType == alarm.RadarVitalsWeak {
					// 检查 device_alarms 中是否有该报警项的配置
					if _, exists := radarAlarms[item.AlarmType]; exists {
						// 如果存在配置，使用 DefaultAlarmSetting 中的 alarm_params
						if item.AlarmParams != nil {
							if durationMin, ok := item.AlarmParams["duration_min"].(int); ok {
								if _, exists := settings["weak_vital_duration"]; !exists && durationMin > 0 {
									settings["weak_vital_duration"] = durationMin
								}
							} else if durationMin, ok := item.AlarmParams["duration_min"].(float64); ok {
								if _, exists := settings["weak_vital_duration"]; !exists && int(durationMin) > 0 {
									settings["weak_vital_duration"] = int(durationMin)
								}
							}
							if sensitivity, ok := item.AlarmParams["sensitivity"].(int); ok {
								if _, exists := settings["weak_vital_sensitivity"]; !exists && sensitivity > 0 {
									settings["weak_vital_sensitivity"] = sensitivity
								}
							} else if sensitivity, ok := item.AlarmParams["sensitivity"].(float64); ok {
								if _, exists := settings["weak_vital_sensitivity"]; !exists && int(sensitivity) > 0 {
									settings["weak_vital_sensitivity"] = int(sensitivity)
								}
							}
						}
					}
					break
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

	// 转换为 AlarmItem 数组（简化处理，返回空数组）
	// 注意：硬件读取功能暂时不使用，如果需要可以后续实现转换逻辑
	alarmItems := make([]alarm.AlarmItem, 0)

	return &GetDeviceMonitorSettingsResponse{
		AlarmItems: alarmItems,
	}, nil
}

// updateSleepaceSettingsToHardware 将配置同步到 Sleepace 硬件（参考 v1.0 实现）
// 参考：wisefido-backend/wisefido-sleepace/modules/sleepace_service.go::SetDeviceMonitorSettings
func (s *deviceMonitorSettingsService) updateSleepaceSettingsToHardware(ctx context.Context, device *domain.Device, deviceID string, settings map[string]interface{}) error {
	// 获取设备代码（deviceCode），使用 device_uid
	if device.DeviceUID == "" {
		return fmt.Errorf("device has no device_uid")
	}
	deviceCode := device.DeviceUID

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
