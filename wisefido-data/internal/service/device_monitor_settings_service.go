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

// ProgressCallback 进度回调函数类型
type ProgressCallback func(percent int, message string)

// DeviceMonitorSettingsService 设备监控配置服务接口
type DeviceMonitorSettingsService interface {
	// 获取设备监控配置（根据设备类型返回 Sleepace 或 Radar 配置）
	GetDeviceMonitorSettings(ctx context.Context, tenantID, deviceID, deviceType string) ([]alarm.AlarmItem, error)

	// 更新设备监控配置（根据设备类型更新 Sleepace 或 Radar 配置）
	// progressCallback: 进度回调函数，用于推送进度更新（可选，如果为 nil 则不推送进度）
	UpdateDeviceMonitorSettings(ctx context.Context, tenantID, deviceID, deviceType, userID string, alarmItems []alarm.AlarmItem, progressCallback ProgressCallback) (bool, bool, error)

	// 获取默认配置（硬编码阈值，Alarm Level 优先从当前租户的 alarm_cloud 读取）
	GetDefaultDeviceMonitorSettings(ctx context.Context, tenantID string, deviceType string) ([]alarm.AlarmItem, error)
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
	qinglanClient      *QinglanClient           // 下发雷达属性（工作模式、跌倒/呼吸心率），不重启
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
// Service 方法实现
// ============================================

// buildDeviceAlarmItemsFromCloudOrDefault 从 alarm_cloud 或 DefaultAlarmSetting 构建设备的 AlarmItem 数组
// deviceType: 数据库中的 device_type（如 "Sleepad", "radar"），直接使用 device_store.device_type
func (s *deviceMonitorSettingsService) buildDeviceAlarmItemsFromCloudOrDefault(ctx context.Context, tenantID, deviceType string) ([]alarm.AlarmItem, error) {
	// 标准化 device_type
	normalizedType := s.normalizeDeviceType(deviceType)
	if normalizedType == "" {
		return nil, fmt.Errorf("invalid device_type: %s (must be Sleepad/sleepace/sleepad or radar)", deviceType)
	}

	// 1. 尝试从 alarm_cloud 获取配置
	alarmCloudConfig, err := s.getAlarmCloudConfig(ctx, tenantID)
	if err == nil && alarmCloudConfig != nil {
		// 从 alarm_cloud 获取对应设备类型的 AlarmItems
		if normalizedType == "sleepad" {
			return alarmCloudConfig.AlarmSetting.Sleepad, nil
		} else if normalizedType == "radar" {
			return alarmCloudConfig.AlarmSetting.Radar, nil
		}
	}

	// 2. 如果 alarm_cloud 不存在，使用 DefaultAlarmSetting（通过 owl-common 函数）
	return alarm.GetDefaultAlarmItems(deviceType), nil
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
	for _, item := range alarm.GetDefaultAlarmItemsSleepPad() {
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
	for _, item := range alarm.GetDefaultAlarmItemsRadar() {
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
	normalizedType := s.normalizeDeviceType(deviceType)
	if normalizedType == "" {
		return nil, fmt.Errorf("invalid device_type: %s (must be Sleepad/sleepace/sleepad or radar)", deviceType)
	}

	// 使用 owl-common 函数获取默认配置
	defaultItems := alarm.GetDefaultAlarmItems(deviceType)
	if defaultItems == nil {
		return nil, fmt.Errorf("invalid device_type: %s", deviceType)
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
func (s *deviceMonitorSettingsService) GetDeviceMonitorSettings(ctx context.Context, tenantID, deviceID, deviceType string) ([]alarm.AlarmItem, error) {
	// 参数验证
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if deviceID == "" || deviceID == "undefined" || deviceID == "null" {
		return nil, fmt.Errorf("device_id is required and must be a valid UUID")
	}
	if deviceType != "sleepace" && deviceType != "radar" {
		return nil, fmt.Errorf("invalid device_type: %s (must be 'sleepace' or 'radar')", deviceType)
	}

	// 验证设备存在
	device, err := s.devicesRepo.GetDevice(ctx, tenantID, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	// 验证设备类型匹配（安全性检查：防止恶意修改 URL 参数）
	// deviceType 是前端传入的 "sleepace" 或 "radar"
	// device.DeviceType 是从 devices 表 JOIN device_store 获取的 device_type（如 "Sleepad", "radar"）
	if !device.DeviceType.Valid {
		return nil, fmt.Errorf("device has no device_type")
	}

	if !s.isDeviceTypeMatch(deviceType, device.DeviceType.String) {
		return nil, fmt.Errorf("device type mismatch: expected %s (from request), got %s (from device_store)", deviceType, device.DeviceType.String)
	}

	// 对于 Sleepace 设备，优先从硬件读取配置（参考 v1.0 实现）
	// 如果硬件读取失败（设备离线或网络问题），则从数据库读取配置
	// 注意：仅当 Device Type = Sleepad 且 Device Model = BM8701-2 时才使用 Sleepace 硬件配置
	// TODO: 暂时注释掉硬件读取功能，后续需要时再启用
	// 如果启用，使用 GetSleepaceSettingsFromHardware(ctx, s.sleepaceClient, device, deviceID)
	/*
		if deviceType == "sleepace" && s.sleepaceClient != nil {
			// 检查设备型号是否为 BM8701-2（Sleepace 厂家）
			deviceTypeStr := ""
			deviceModel := ""
			if device.DeviceType.Valid {
				deviceTypeStr = device.DeviceType.String
			}
			if device.DeviceModel.Valid {
				deviceModel = device.DeviceModel.String
			}
			// 仅当 Device Type = Sleepad 且 Device Model = BM8701-2 时才从硬件读取
			if strings.EqualFold(deviceTypeStr, "sleepad") && deviceModel == "BM8701-2" {
				alarmItems, err := GetSleepaceSettingsFromHardware(ctx, s.sleepaceClient, device, deviceID)
				if err != nil {
					s.logger.Warn("Failed to get settings from hardware, falling back to database",
						zap.String("tenant_id", tenantID),
						zap.String("device_id", deviceID),
						zap.String("device_status", device.Status),
						zap.Error(err),
					)
					// 如果硬件读取失败，继续从数据库读取
				} else {
					// 成功从硬件读取，直接返回
					return alarmItems, nil
				}
			} else {
				// 设备型号不是 BM8701-2，跳过硬件读取，从数据库读取
				s.logger.Info("Device model is not BM8701-2, skipping hardware read",
					zap.String("device_id", deviceID),
					zap.String("device_type", deviceTypeStr),
					zap.String("device_model", deviceModel),
				)
			}
		}
	*/

	// 1. 获取设备的监控配置（从数据库）
	alarmDevice, err := s.alarmDeviceRepo.GetAlarmDevice(ctx, tenantID, deviceID)
	if err != nil {
		// 如果配置不存在，返回错误（用户需要先使用 GetDefaultDeviceMonitorSettings 获取默认配置，然后在此基础上修改）
		return nil, fmt.Errorf("device monitor settings not found, please use GetDefaultDeviceMonitorSettings first: %w", err)
	}

	// 2. 如果配置存在，转换为 AlarmItem 数组
	// 直接使用 deviceType（前端已经知道是 radar 还是 sleepace）
	alarmItems, err := s.convertMonitorConfigToAlarmItems(alarmDevice.MonitorConfig, deviceType)
	if err != nil {
		return nil, fmt.Errorf("failed to convert monitor config to alarm items: %w", err)
	}

	return alarmItems, nil
}

// UpdateDeviceMonitorSettings 更新设备监控配置（简化版本，使用事务）
func (s *deviceMonitorSettingsService) UpdateDeviceMonitorSettings(ctx context.Context, tenantID, deviceID, deviceType, userID string, alarmItems []alarm.AlarmItem, progressCallback ProgressCallback) (success bool, noChange bool, err error) {
	// 参数验证
	if tenantID == "" {
		return false, false, fmt.Errorf("tenant_id is required")
	}
	if deviceID == "" {
		return false, false, fmt.Errorf("device_id is required")
	}
	if deviceType != "sleepace" && deviceType != "radar" {
		return false, false, fmt.Errorf("invalid device_type: %s (must be 'sleepace' or 'radar')", deviceType)
	}
	if alarmItems == nil || len(alarmItems) == 0 {
		return false, false, fmt.Errorf("alarm_items is required")
	}

	// 验证设备存在
	device, err := s.devicesRepo.GetDevice(ctx, tenantID, deviceID)
	if err != nil {
		return false, false, fmt.Errorf("failed to get device: %w", err)
	}

	// 验证设备类型匹配（安全性检查：防止恶意修改 URL 参数）
	// deviceType 是前端传入的 "sleepace" 或 "radar"
	// device.DeviceType 是从 devices 表 JOIN device_store 获取的 device_type（如 "Sleepad", "radar"）
	if !device.DeviceType.Valid {
		return false, false, fmt.Errorf("device has no device_type")
	}

	if !s.isDeviceTypeMatch(deviceType, device.DeviceType.String) {
		return false, false, fmt.Errorf("device type mismatch: expected %s (from request), got %s (from device_store)", deviceType, device.DeviceType.String)
	}

	// 1. 获取现有配置（必须存在，否则返回错误）
	existingAlarmDevice, err := s.alarmDeviceRepo.GetAlarmDevice(ctx, tenantID, deviceID)
	if err != nil {
		// 如果配置不存在，返回错误（用户需要先使用 GetDefaultDeviceMonitorSettings 获取默认配置，然后在此基础上修改）
		return false, false, fmt.Errorf("device monitor settings not found, please use GetDefaultDeviceMonitorSettings first: %w", err)
	}

	// 转换为 AlarmItem 数组
	// 直接使用 deviceType（前端已经知道是 radar 还是 sleepace）
	existingAlarmItems, err := s.convertMonitorConfigToAlarmItems(existingAlarmDevice.MonitorConfig, deviceType)
	if err != nil {
		return false, false, fmt.Errorf("failed to convert existing monitor config: %w", err)
	}

	// 2. 比对配置是否有变化（逐项比对，避免 JSON 序列化问题）
	changedAlarmTypes := s.getChangedAlarmTypes(existingAlarmItems, alarmItems)

	// 2.1. 如果配置无变化，直接返回 "no Change"
	if len(changedAlarmTypes) == 0 {
		s.logger.Info("Device monitor settings unchanged, skipping update",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.String("device_type", deviceType),
		)
		return true, true, nil
	}

	// 2.2. 检查变化的项是否包含需要写入设备的项
	// 需要写入设备的 AlarmType：
	// - Radar_MonitoringMode (工作模式)
	// - Radar_Fall (跌倒告警时间，影响 fall_param 字节 3)
	// - Radar_PostureDetection (姿态检测，影响 fall_param 字节 4 bit1)
	// - Radar_BedSitUp (床上坐起，影响 fall_param 字节 4 bit2)
	// - Radar_SittingOnGround (坐地告警，影响 fall_param 字节 4 bit0 和字节 5)
	// - Radar_AbnormalRespiratoryRate (呼吸异常，影响 heart_breath_param)
	// - Radar_AbnormalHeartRate (心率异常，影响 heart_breath_param)
	// - Radar_VitalsWeak (生命体征微弱，影响 heart_breath_param)
	deviceWriteAlarmTypes := map[string]bool{
		alarm.RadarMonitoringMode:          true,
		alarm.RadarFall:                    true,
		alarm.RadarPostureDetection:        true,
		alarm.RadarBedSitUp:                true,
		alarm.RadarSittingOnGround:         true,
		alarm.RadarAbnormalRespiratoryRate: true,
		alarm.RadarAbnormalHeartRate:       true,
		alarm.RadarVitalsWeak:              true,
	}

	hasDeviceWriteChanges := false
	for _, alarmType := range changedAlarmTypes {
		if deviceWriteAlarmTypes[alarmType] {
			hasDeviceWriteChanges = true
			break
		}
	}

	// 调试日志：检查设备写入条件
	s.logger.Info("Checking device write conditions",
		zap.String("tenant_id", tenantID),
		zap.String("device_id", deviceID),
		zap.String("device_type", deviceType),
		zap.Bool("is_radar", deviceType == "radar"),
		zap.Bool("qinglan_client_not_nil", s.qinglanClient != nil),
		zap.String("device_uid", device.DeviceUID),
		zap.Bool("has_device_uid", device.DeviceUID != ""),
		zap.Bool("has_device_write_changes", hasDeviceWriteChanges),
		zap.Strings("changed_alarm_types", changedAlarmTypes),
	)

	// 3. 对于雷达设备，先写入设备（工作模式、跌倒参数、呼吸心率参数）
	// 如果写入设备成功，再写入DB；如果部分失败，读取设备实际值后写入DB
	// 注意：设备写入设置8秒超时，避免前端10秒超时导致整个请求失败
	var finalAlarmItems []alarm.AlarmItem = alarmItems
	if deviceType == "radar" && s.qinglanClient != nil && device.DeviceUID != "" && hasDeviceWriteChanges {
		s.logger.Info("Entering device write logic, converting alarm items to device properties",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.String("device_uid", device.DeviceUID),
		)

		deviceProps, err := AlarmItemsToRadarDeviceProps(alarmItems)
		if err != nil {
			return false, false, fmt.Errorf("failed to convert alarm items to device properties: %w", err)
		}

		s.logger.Info("Device properties converted",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.Int("device_props_count", len(deviceProps)),
			zap.Any("device_props", deviceProps),
		)

		// 如果有需要写入设备的属性，先写入设备
		if len(deviceProps) > 0 {
			s.logger.Info("Calling qinglan_client.SetDeviceProperties",
				zap.String("tenant_id", tenantID),
				zap.String("device_id", deviceID),
				zap.String("device_uid", device.DeviceUID),
				zap.Any("device_props", deviceProps),
			)
			
			// 为设备写入设置5秒超时，确保在前端10秒超时之前完成或快速失败
			deviceWriteCtx, deviceWriteCancel := context.WithTimeout(ctx, 5*time.Second)
			defer deviceWriteCancel()

			writeErr := s.qinglanClient.SetDeviceProperties(deviceWriteCtx, device.DeviceUID, deviceProps)
			if writeErr != nil {
				// 写入设备失败（包括超时），记录警告但继续使用前端传入的值
				// 不尝试读取设备值，避免再次超时导致整个请求失败
				s.logger.Warn("Failed to write radar device properties, using request values",
					zap.String("tenant_id", tenantID),
					zap.String("device_id", deviceID),
					zap.String("device_uid", device.DeviceUID),
					zap.Error(writeErr),
					zap.String("note", "Device write failed, but will continue with request values to avoid timeout"),
				)
				// 继续使用前端传入的值 (finalAlarmItems = alarmItems)
			} else {
				// 写入设备成功，使用前端传入的值
				s.logger.Info("Successfully wrote radar device properties",
					zap.String("tenant_id", tenantID),
					zap.String("device_id", deviceID),
					zap.String("device_uid", device.DeviceUID),
				)
			}
		}
	}

	// 4. 将最终 AlarmItem 数组转换为 monitor_config JSONB
	monitorConfigJSON, err := s.convertAlarmItemsToMonitorConfig(finalAlarmItems)
	if err != nil {
		return false, false, fmt.Errorf("failed to convert alarm items to monitor config: %w", err)
	}

	// 5. 使用事务保证原子性：要么全部成功，要么全部失败
	if s.db == nil {
		return false, false, fmt.Errorf("database connection is required for transaction")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, false, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // 确保在函数返回时回滚（如果未提交）

	// 4.1. 如果有变化，先保存旧配置到 config_versions（在事务中）
	// 保存完整配置：包括 alarm_cloud 配置和设备特定的 monitor_config
	if len(changedAlarmTypes) > 0 && s.configVersionsRepo != nil {
		// 获取 alarm_cloud 配置（用于合并生成完整配置）
		alarmCloudConfig, err := s.getAlarmCloudConfig(ctx, tenantID)
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
					Sleepad: alarm.GetDefaultAlarmItemsSleepPad(),
					Radar:   alarm.GetDefaultAlarmItemsRadar(),
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
			return false, false, fmt.Errorf("failed to marshal config_data for config_version: %w", err)
		}

		configVersion := &domain.ConfigVersion{
			ConfigType:      "alarm_device",
			EntityID:        deviceID, // alarm_device 的 entity_id 就是 device_id
			CurrentEntityID: deviceID,
			ConfigData:      json.RawMessage(configDataJSON),
			ValidFrom:       time.Now().UTC(),
		}
		if userID != "" {
			configVersion.CreatedBy = &userID
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
		_, err = tx.ExecContext(ctx, updateOldQuery, tenantID, configVersion.ConfigType, configVersion.EntityID, configVersion.ValidFrom)
		if err != nil {
			return false, false, fmt.Errorf("failed to update old config versions: %w", err)
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
		err = tx.QueryRowContext(ctx, insertQuery, tenantID, configVersion.ConfigType, configVersion.EntityID,
			currentEntityID, string(configVersion.ConfigData), configVersion.ValidFrom, validTo, createdBy).Scan(&versionID)
		if err != nil {
			return false, false, fmt.Errorf("failed to create config version: %w", err)
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
	if userID != "" {
		updatedBy = userID
	}

	_, err = tx.ExecContext(ctx, upsertQuery, deviceID, tenantID, string(monitorConfigJSON), time.Now().UTC(), updatedBy)
	if err != nil {
		return false, false, fmt.Errorf("failed to update alarm device: %w", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return false, false, fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Info("Device monitor settings updated",
		zap.String("tenant_id", tenantID),
		zap.String("device_id", deviceID),
		zap.String("device_type", deviceType),
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
		if deviceCode == "" {
			deviceCode = device.DeviceUID // device_code 仅 Sleepace 有，Radar 无；空则用 device_uid
		}

		// 获取设备类型（转换为标准格式）
		notifyDeviceType := ""
		if device.DeviceType.Valid {
			deviceTypeStr := device.DeviceType.String
			// 转换为标准格式：Sleepad/sleepad/sleepace/sleeppad -> "Sleepad", radar -> "Radar"
			switch strings.ToLower(deviceTypeStr) {
			case "Sleepad", "sleepad", "sleepace", "sleeppad":
				notifyDeviceType = "Sleepad"
			case "radar":
				notifyDeviceType = "Radar"
			default:
				notifyDeviceType = deviceTypeStr
			}
		}

		// 位置信息可选，可为空（当前不传递，后续如需可添加获取逻辑）
		if err := s.configNotifier.NotifyAlarmDeviceUpdated(ctx, tenantID, deviceID, deviceUID, deviceCode, notifyDeviceType, nil); err != nil {
			s.logger.Warn("Failed to notify config change, but database update succeeded",
				zap.String("tenant_id", tenantID),
				zap.String("device_id", deviceID),
				zap.String("device_uid", deviceUID),
				zap.String("device_code", deviceCode),
				zap.String("device_type", notifyDeviceType),
				zap.Error(err),
			)
		}
	}

	return true, false, nil
}

// ============================================
// 辅助方法
// ============================================

// normalizeDeviceType 标准化设备类型，返回 "sleepad" 或 "radar"，无效返回空字符串
func (s *deviceMonitorSettingsService) normalizeDeviceType(deviceType string) string {
	deviceTypeLower := strings.ToLower(deviceType)
	if deviceTypeLower == "sleepace" || deviceTypeLower == "sleepad" || deviceTypeLower == "sleeppad" {
		return "sleepad"
	}
	if deviceTypeLower == "radar" {
		return "radar"
	}
	return ""
}

// isDeviceTypeMatch 判断两个设备类型是否匹配（支持多种变体）
func (s *deviceMonitorSettingsService) isDeviceTypeMatch(type1, type2 string) bool {
	normalized1 := s.normalizeDeviceType(type1)
	normalized2 := s.normalizeDeviceType(type2)
	return normalized1 != "" && normalized1 == normalized2
}

// getChangedAlarmTypes 逐项比对两个 AlarmItem 数组，返回有变化的 AlarmType 列表
func (s *deviceMonitorSettingsService) getChangedAlarmTypes(existing, new []alarm.AlarmItem) []string {
	changedTypes := make(map[string]bool)

	// 构建现有配置的索引（按 AlarmType）
	existingMap := make(map[string]alarm.AlarmItem)
	for _, item := range existing {
		if item.AlarmType != "" {
			existingMap[item.AlarmType] = item
		}
	}

	// 构建新配置的索引（按 AlarmType）
	newMap := make(map[string]alarm.AlarmItem)
	for _, item := range new {
		if item.AlarmType != "" {
			newMap[item.AlarmType] = item
		}
	}

	// 检查新配置中的每个 AlarmItem
	for alarmType, newItem := range newMap {
		existingItem, exists := existingMap[alarmType]
		if !exists {
			// 新配置中有现有配置没有的 AlarmType，认为有变化
			changedTypes[alarmType] = true
			continue
		}

		// 比对每个字段
		if !s.compareAlarmItem(existingItem, newItem) {
			changedTypes[alarmType] = true
		}
	}

	// 检查是否有现有配置中有但新配置中没有的 AlarmType
	for alarmType := range existingMap {
		if _, exists := newMap[alarmType]; !exists {
			changedTypes[alarmType] = true
		}
	}

	// 转换为列表
	result := make([]string, 0, len(changedTypes))
	for alarmType := range changedTypes {
		result = append(result, alarmType)
	}

	return result
}

// compareAlarmItem 比对两个 AlarmItem 是否完全相同
func (s *deviceMonitorSettingsService) compareAlarmItem(existing, new alarm.AlarmItem) bool {
	// 比对 AlarmType
	if existing.AlarmType != new.AlarmType {
		return false
	}

	// 比对 IsEnabled（指针比较）
	if !compareIntPtr(existing.IsEnabled, new.IsEnabled) {
		return false
	}

	// 比对 AlarmLevel（指针比较）
	if !compareStringPtr(existing.AlarmLevel, new.AlarmLevel) {
		return false
	}

	// 比对 DisplaySetting
	if existing.DisplaySetting != new.DisplaySetting {
		return false
	}

	// 比对 AlarmParams（map 深度比对）
	if !s.compareAlarmParams(existing.AlarmParams, new.AlarmParams) {
		return false
	}

	return true
}

// compareAlarmParams 比对两个 AlarmParams map 是否完全相同
func (s *deviceMonitorSettingsService) compareAlarmParams(existing, new map[string]interface{}) bool {
	// 如果都为 nil，认为相同
	if existing == nil && new == nil {
		return true
	}

	// 如果只有一个为 nil，认为不同
	if existing == nil || new == nil {
		return false
	}

	// 如果长度不同，认为不同
	if len(existing) != len(new) {
		return false
	}

	// 逐个比对每个 key-value
	for key, existingValue := range existing {
		newValue, exists := new[key]
		if !exists {
			return false
		}

		// 深度比对值
		if !s.compareValues(existingValue, newValue) {
			return false
		}
	}

	return true
}

// compareValues 比对两个值是否相同（支持基本类型）
func (s *deviceMonitorSettingsService) compareValues(v1, v2 interface{}) bool {
	// 如果类型不同，尝试转换为相同类型后比对
	switch val1 := v1.(type) {
	case int:
		switch val2 := v2.(type) {
		case int:
			return val1 == val2
		case int64:
			return int64(val1) == val2
		case float64:
			return float64(val1) == val2
		default:
			return false
		}
	case int64:
		switch val2 := v2.(type) {
		case int:
			return val1 == int64(val2)
		case int64:
			return val1 == val2
		case float64:
			return float64(val1) == val2
		default:
			return false
		}
	case float64:
		switch val2 := v2.(type) {
		case int:
			return val1 == float64(val2)
		case int64:
			return val1 == float64(val2)
		case float64:
			return val1 == val2
		default:
			return false
		}
	case string:
		val2, ok := v2.(string)
		return ok && val1 == val2
	case bool:
		val2, ok := v2.(bool)
		return ok && val1 == val2
	case nil:
		return v2 == nil
	default:
		// 对于其他类型（如 map、slice），使用简单的相等比较
		return v1 == v2
	}
}

// compareIntPtr 比对两个 int 指针是否相同
func compareIntPtr(p1, p2 *int) bool {
	if p1 == nil && p2 == nil {
		return true
	}
	if p1 == nil || p2 == nil {
		return false
	}
	return *p1 == *p2
}

// compareStringPtr 比对两个 string 指针是否相同
func compareStringPtr(p1, p2 *string) bool {
	if p1 == nil && p2 == nil {
		return true
	}
	if p1 == nil || p2 == nil {
		return false
	}
	return *p1 == *p2
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
				// 只验证，不转换（前端已统一使用 AlarmItem 格式）
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
func (s *deviceMonitorSettingsService) GetDefaultDeviceMonitorSettings(ctx context.Context, tenantID string, deviceType string) ([]alarm.AlarmItem, error) {
	if deviceType != "sleepace" && deviceType != "radar" {
		return nil, fmt.Errorf("invalid device_type: %s (must be 'sleepace' or 'radar')", deviceType)
	}

	// 从 alarm_cloud 或 DefaultAlarmSetting 获取默认配置
	// buildDeviceAlarmItemsFromCloudOrDefault 会先尝试从 alarm_cloud 获取，如果没有则从 alarm.go 中的默认值获取
	alarmItems, err := s.buildDeviceAlarmItemsFromCloudOrDefault(ctx, tenantID, deviceType)
	if err != nil {
		return nil, fmt.Errorf("failed to build alarm items: %w", err)
	}

	return alarmItems, nil
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
			for _, item := range alarm.GetDefaultAlarmItemsRadar() {
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
