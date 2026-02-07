package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"owl-common/alarm"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"
	"wisefido-qinglan/decode"

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
	// 对于 radar 设备，返回 UpdateRadarResult；对于 sleepace 设备，返回 (success, noChange, error)
	UpdateDeviceMonitorSettings(ctx context.Context, tenantID, deviceID, deviceType, userID string, alarmItems []alarm.AlarmItem, progressCallback ProgressCallback) (interface{}, error)

	// 获取默认配置（硬编码阈值，Alarm Level 优先从当前租户的 alarm_cloud 读取）
	GetDefaultDeviceMonitorSettings(ctx context.Context, tenantID string, deviceType string) ([]alarm.AlarmItem, error)

	// 检查设备在线状态（通过 wisefido-qinglan HTTP API 实时查询）
	// 返回 nil 表示设备在线，否则返回错误
	CheckDeviceOnlineStatus(ctx context.Context, deviceUID string) error
}

// deviceMonitorSettingsService 实现
type deviceMonitorSettingsService struct {
	alarmDeviceRepo    repository.AlarmDeviceRepository
	alarmCloudRepo     repository.AlarmCloudRepository
	configVersionsRepo repository.ConfigVersionsRepository // 配置版本仓库（用于审计）
	devicesRepo        repository.DevicesRepository
	deviceStoreRepo    repository.DeviceStoreRepository
	sleepaceClient     *SleepaceClient // Sleepace 硬件 API 客户端（可选）
	configPublisher    ConfigPublisher // 配置消息发布器
	qinglanClient      *QinglanClient  // 雷达设备仅经此客户端：查询状态/属性、下发属性（工作模式、跌倒/呼吸心率）
	db                 *sql.DB         // 用于事务操作
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
	configPublisher ConfigPublisher,
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
		configPublisher:    configPublisher,
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

// CheckDeviceOnlineStatus 检查设备在线状态（通过 wisefido-qinglan HTTP API 实时查询）
// 返回 nil 表示设备在线，否则返回错误
func (s *deviceMonitorSettingsService) CheckDeviceOnlineStatus(ctx context.Context, deviceUID string) error {
	if s.qinglanClient == nil {
		return fmt.Errorf("qinglan client not initialized")
	}

	if deviceUID == "" {
		return fmt.Errorf("device_uid is required")
	}

	status, err := s.qinglanClient.GetDeviceStatus(ctx, deviceUID)
	if err != nil {
		s.logger.Warn("Failed to get device status from wisefido-qinglan",
			zap.String("device_uid", deviceUID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to query device status: %w", err)
	}

	if status != "online" {
		return fmt.Errorf("device is offline (status: %s)", status)
	}

	return nil
}

// ============================================
// Service 方法实现
// ============================================

// buildDeviceAlarmItemsFromCloudOrDefault 从 alarm_cloud 或 DefaultAlarmSetting 构建设备的 AlarmItem 数组
// deviceType: 数据库中的 device_type（如 "Sleepad", "radar"），直接使用 device_store.device_type
// 逻辑：总是使用 alarm.go 中的默认值，但如果 alarm_cloud 存在，用 alarm_cloud 的 alarm_level 覆盖
func (s *deviceMonitorSettingsService) buildDeviceAlarmItemsFromCloudOrDefault(ctx context.Context, tenantID, deviceType string) ([]alarm.AlarmItem, error) {
	// 标准化 device_type
	normalizedType := s.normalizeDeviceType(deviceType)
	if normalizedType == "" {
		return nil, fmt.Errorf("invalid device_type: %s (must be Sleepad/sleepace/sleepad or radar)", deviceType)
	}

	// 1. 总是从 alarm.go 获取所有默认项（包括 DisplayAlarmDevice 的项）
	defaultItems := alarm.GetDefaultAlarmItems(deviceType)
	s.logger.Info("buildDeviceAlarmItemsFromCloudOrDefault: starting",
		zap.String("tenant_id", tenantID),
		zap.String("device_type", deviceType),
		zap.Int("default_items_count", len(defaultItems)),
	)

	// 2. 尝试从 alarm_cloud 获取 alarm_level 覆盖
	alarmCloud, err := s.alarmCloudRepo.GetAlarmCloud(ctx, tenantID)
	if err != nil {
		// alarm_cloud 不存在，直接返回默认值
		s.logger.Info("buildDeviceAlarmItemsFromCloudOrDefault: alarm_cloud not found, returning defaults",
			zap.String("tenant_id", tenantID),
			zap.Error(err),
		)
		return defaultItems, nil
	}

	// 3. 解析 device_alarms，用 alarm_cloud 的 alarm_level 覆盖默认值
	deviceAlarmsMap := make(map[string]map[string]string)
	if len(alarmCloud.DeviceAlarms) > 0 {
		if err := json.Unmarshal(alarmCloud.DeviceAlarms, &deviceAlarmsMap); err != nil {
			// 解析失败，返回默认值
			s.logger.Warn("buildDeviceAlarmItemsFromCloudOrDefault: failed to parse device_alarms, returning defaults",
				zap.String("tenant_id", tenantID),
				zap.Error(err),
			)
			return defaultItems, nil
		}
		s.logger.Info("buildDeviceAlarmItemsFromCloudOrDefault: parsed device_alarms",
			zap.String("tenant_id", tenantID),
			zap.Any("device_alarms_map", deviceAlarmsMap),
		)
	} else {
		s.logger.Info("buildDeviceAlarmItemsFromCloudOrDefault: device_alarms is empty, returning defaults",
			zap.String("tenant_id", tenantID),
		)
		return defaultItems, nil
	}

	// 4. 确定设备类型在 alarm_cloud 中的 key
	deviceTypeKey := ""
	if normalizedType == "sleepad" {
		deviceTypeKey = "SleepPad"
	} else if normalizedType == "radar" {
		deviceTypeKey = "Radar"
	}

	// 5. 遍历默认项，用 alarm_cloud 的 alarm_level 覆盖
	result := make([]alarm.AlarmItem, 0, len(defaultItems))
	for _, item := range defaultItems {
		// 深拷贝 item，避免修改原始默认值
		newItem := item
		// 深拷贝指针字段
		if item.AlarmLevel != nil {
			levelCopy := *item.AlarmLevel
			newItem.AlarmLevel = &levelCopy
		}
		if item.IsEnabled != nil {
			enabledCopy := *item.IsEnabled
			newItem.IsEnabled = &enabledCopy
		}
		// 深拷贝 AlarmParams map
		if item.AlarmParams != nil {
			newParams := make(map[string]interface{})
			for k, v := range item.AlarmParams {
				newParams[k] = v
			}
			newItem.AlarmParams = newParams
		}

		// 如果 alarm_cloud 中存在该项，用 alarm_cloud 的 alarm_level 覆盖
		if deviceTypeKey != "" {
			if typeMap, ok := deviceAlarmsMap[deviceTypeKey]; ok {
				if level, exists := typeMap[item.AlarmType]; exists {
					// 用 alarm_cloud 的 alarm_level 覆盖
					// 去除可能的空格和引号
					level = strings.TrimSpace(level)
					level = strings.Trim(level, `"`)
					newItem.AlarmLevel = &level
					// 转换为大写进行比较，避免大小写问题
					levelUpper := strings.ToUpper(level)
					if levelUpper == "DISABLE" || levelUpper == "DISABLED" {
						enabled := alarm.IsEnabledOff
						newItem.IsEnabled = &enabled
						s.logger.Debug("Setting alarm item to disabled from alarm_cloud",
							zap.String("alarm_type", item.AlarmType),
							zap.String("level", level),
							zap.String("level_upper", levelUpper),
						)
					} else {
						// 任何其他值（EMERG, WARNING, ERR, ERROR等）都设置为 enabled
						enabled := alarm.IsEnabledOn
						newItem.IsEnabled = &enabled
						s.logger.Debug("Setting alarm item to enabled from alarm_cloud",
							zap.String("alarm_type", item.AlarmType),
							zap.String("level", level),
							zap.String("level_upper", levelUpper),
							zap.Int("is_enabled", enabled),
						)
					}
				} else {
					// alarm_cloud 中不存在该项，保持默认值
					s.logger.Debug("Alarm item not found in alarm_cloud, keeping default",
						zap.String("alarm_type", item.AlarmType),
						zap.Any("default_is_enabled", item.IsEnabled),
					)
				}
			}
		}
		result = append(result, newItem)
	}

	// 统计 enabled/disabled 数量
	enabledCount := 0
	disabledCount := 0
	for _, item := range result {
		if item.IsEnabled != nil && *item.IsEnabled == alarm.IsEnabledOn {
			enabledCount++
		} else {
			disabledCount++
		}
	}

	s.logger.Info("buildDeviceAlarmItemsFromCloudOrDefault: completed",
		zap.String("tenant_id", tenantID),
		zap.String("device_type", deviceType),
		zap.Int("total_items", len(result)),
		zap.Int("enabled_count", enabledCount),
		zap.Int("disabled_count", disabledCount),
	)

	return result, nil
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

// GetDeviceMonitorSettings 获取设备监控配置（优化版本：立即返回baseline，异步更新设备值）
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

	// 1. 获取baseline配置（从数据库或默认值）
	s.logger.Info("[GET_SETTINGS] Querying database for alarm_device (baseline)",
		zap.String("tenant_id", tenantID),
		zap.String("device_id", deviceID),
		zap.String("device_type", deviceType),
	)
	alarmDevice, err := s.alarmDeviceRepo.GetAlarmDevice(ctx, tenantID, deviceID)
	baselineAlarmItems := []alarm.AlarmItem{}
	dbExists := false
	if err != nil {
		// 检查是否是"未找到"错误
		isNotFound := err == sql.ErrNoRows ||
			strings.Contains(err.Error(), "no rows in result set") ||
			strings.Contains(err.Error(), "not found")

		if !isNotFound {
			// 其他错误直接返回
			s.logger.Error("[GET_SETTINGS] Database query failed",
				zap.String("tenant_id", tenantID),
				zap.String("device_id", deviceID),
				zap.String("device_type", deviceType),
				zap.Error(err),
			)
			return nil, fmt.Errorf("failed to get alarm device: %w", err)
		}
		// 数据库不存在，使用默认配置作为baseline
		baselineAlarmItems, err = s.buildDeviceAlarmItemsFromCloudOrDefault(ctx, tenantID, deviceType)
		if err != nil {
			return nil, fmt.Errorf("failed to build default alarm items: %w", err)
		}
		s.logger.Info("[GET_SETTINGS] Database query: NOT FOUND (using default as baseline)",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.String("device_type", deviceType),
		)
	} else {
		// 数据库存在，转换为 AlarmItem 数组作为baseline
		baselineAlarmItems, err = s.convertMonitorConfigToAlarmItems(alarmDevice.MonitorConfig, deviceType)
		if err != nil {
			s.logger.Error("[GET_SETTINGS] Database query: SUCCESS but conversion failed",
				zap.String("tenant_id", tenantID),
				zap.String("device_id", deviceID),
				zap.String("device_type", deviceType),
				zap.Error(err),
			)
			return nil, fmt.Errorf("failed to convert monitor config to alarm items: %w", err)
		}
		dbExists = true
		s.logger.Info("[GET_SETTINGS] Database query: SUCCESS (baseline ready)",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.String("device_type", deviceType),
			zap.Int("alarm_items_count", len(baselineAlarmItems)),
		)
	}

	// 2. 立即返回baseline给前端（不等待设备查询）
	s.logger.Info("[GET_SETTINGS] Returning baseline immediately",
		zap.String("tenant_id", tenantID),
		zap.String("device_id", deviceID),
		zap.String("device_type", deviceType),
		zap.Bool("db_exists", dbExists),
	)

	// 3. 对于雷达设备，异步查询设备并更新数据库（不阻塞返回）
	if deviceType == "radar" && s.qinglanClient != nil && device.DeviceUID != "" {
		// 使用goroutine异步执行设备查询和数据库更新
		go s.syncDeviceValuesToDB(context.Background(), tenantID, deviceID, device.DeviceUID, deviceType, baselineAlarmItems, dbExists)
	}

	return baselineAlarmItems, nil
}

// UpdateDeviceMonitorSettings 更新设备监控配置（路由函数，根据设备类型调用对应的实现）
func (s *deviceMonitorSettingsService) UpdateDeviceMonitorSettings(ctx context.Context, tenantID, deviceID, deviceType, userID string, alarmItems []alarm.AlarmItem, progressCallback ProgressCallback) (interface{}, error) {
	// 参数验证
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}
	if deviceType != "sleepace" && deviceType != "radar" {
		return nil, fmt.Errorf("invalid device_type: %s (must be 'sleepace' or 'radar')", deviceType)
	}
	if alarmItems == nil || len(alarmItems) == 0 {
		return nil, fmt.Errorf("alarm_items is required")
	}

	// 根据设备类型调用对应的实现
	if deviceType == "radar" {
		return s.UpdateRadarMonitorSettings(ctx, tenantID, deviceID, userID, alarmItems, progressCallback)
	} else if deviceType == "sleepace" {
		success, noChange, err := s.UpdateSleepadMonitorSettings(ctx, tenantID, deviceID, userID, alarmItems, progressCallback)
		return map[string]interface{}{
			"success":      success,
			"no_change":    noChange,
			"device_write": success,
			"db_write":     success,
		}, err
	}
	return nil, fmt.Errorf("unsupported device_type: %s", deviceType)
}

// UpdateRadarMonitorSettings 更新 Radar 设备监控配置
// 返回：设备执行结果、DB 执行结果、失败项列表、是否有变化、错误
func (s *deviceMonitorSettingsService) UpdateRadarMonitorSettings(ctx context.Context, tenantID, deviceID, userID string, alarmItems []alarm.AlarmItem, progressCallback ProgressCallback) (*UpdateRadarResult, error) {
	deviceType := "radar"

	result := &UpdateRadarResult{
		DeviceWriteSuccess: false,
		DBWriteSuccess:     false,
		FailedAlarmTypes:   []alarm.AlarmItem{},
		NoChange:           false,
	}

	// 验证设备存在
	device, err := s.devicesRepo.GetDevice(ctx, tenantID, deviceID)
	if err != nil {
		return result, fmt.Errorf("failed to get device: %w", err)
	}

	// 验证设备类型匹配
	if !device.DeviceType.Valid {
		return result, fmt.Errorf("device has no device_type")
	}
	if !s.isDeviceTypeMatch(deviceType, device.DeviceType.String) {
		return result, fmt.Errorf("device type mismatch: expected %s (from request), got %s (from device_store)", deviceType, device.DeviceType.String)
	}

	// 获取现有配置
	existingAlarmItems, err := s.getExistingAlarmItems(ctx, tenantID, deviceID, deviceType)
	if err != nil {
		return result, err
	}

	// Radar 设备需要写入的 AlarmType
	deviceWriteAlarmTypes := map[string]bool{
		alarm.MonitoringMode:          true,
		alarm.Fall:                    true,
		alarm.PostureDetection:        true,
		alarm.BedSitUp:                true,
		alarm.SittingOnGround:         true,
		alarm.AbnormalRespiratoryRate: true,
		alarm.AbnormalHeartRate:       true,
		alarm.VitalsWeak:              true,
	}
	// Radar 设备会返回的 AlarmType（从设备属性解码得到）
	deviceReturnAlarmTypes := map[string]bool{
		alarm.MonitoringMode:          true, // 从 radar_func_ctrl 返回
		alarm.Fall:                    true, // 从 fall_param 返回
		alarm.PostureDetection:        true, // 从 fall_param 返回
		alarm.BedSitUp:                true, // 从 fall_param 返回（bit 2）
		alarm.SittingOnGround:         true, // 从 fall_param 返回
		alarm.AbnormalRespiratoryRate: true, // 从 heart_breath_param 返回
		alarm.AbnormalHeartRate:       true, // 从 heart_breath_param 返回
		alarm.VitalsWeak:              true, // 从 heart_breath_param 返回
	}

	_, noChange, hasDeviceWriteChanges, err := s.compareAndLogChanges(ctx, tenantID, deviceID, deviceType, existingAlarmItems, alarmItems, deviceWriteAlarmTypes, deviceReturnAlarmTypes)
	if err != nil {
		return result, err
	}
	if noChange {
		result.NoChange = true
		return result, nil
	}

	// 如果有需要写入设备的变化，调用 radarWrite 写入设备
	if hasDeviceWriteChanges {
		writeResult, writeErr := s.radarWrite(ctx, device.DeviceUID, alarmItems, progressCallback)
		if writeErr != nil {
			// 写入失败，只留log，不更新DB，并把log发给前端
			s.logger.Error("[RADAR_WRITE] Failed to write to device",
				zap.String("tenant_id", tenantID),
				zap.String("device_id", deviceID),
				zap.String("device_uid", device.DeviceUID),
				zap.Error(writeErr),
			)

			// 收集失败的 alarm_type，使用旧值
			if writeResult != nil {
				for _, item := range alarmItems {
					if status, ok := writeResult.Results[item.AlarmType]; ok && status == "Fail" {
						// 从 existingAlarmItems 中找到对应的旧值
						for _, existingItem := range existingAlarmItems {
							if existingItem.AlarmType == item.AlarmType {
								result.FailedAlarmTypes = append(result.FailedAlarmTypes, existingItem)
								break
							}
						}
					}
				}
			}

			result.DeviceWriteSuccess = false
			result.DBWriteSuccess = false
			return result, nil
		}

		// 写入成功，检查是否有失败的项
		result.DeviceWriteSuccess = true
		finalAlarmItems := make([]alarm.AlarmItem, len(alarmItems))
		copy(finalAlarmItems, alarmItems)

		// 对于失败的项，使用旧值替换
		if writeResult != nil {
			for i, item := range finalAlarmItems {
				if status, ok := writeResult.Results[item.AlarmType]; ok && status == "Fail" {
					// 从 existingAlarmItems 中找到对应的旧值
					for _, existingItem := range existingAlarmItems {
						if existingItem.AlarmType == item.AlarmType {
							finalAlarmItems[i] = existingItem
							result.FailedAlarmTypes = append(result.FailedAlarmTypes, existingItem)
							s.logger.Warn("[RADAR_WRITE] Alarm type write failed, using old value",
								zap.String("alarm_type", item.AlarmType),
								zap.String("tenant_id", tenantID),
								zap.String("device_id", deviceID),
							)
							break
						}
					}
				}
			}
		}

		// 更新数据库（使用最终的值，失败项已替换为旧值）
		dbErr := s.updateAlarmDeviceDB(ctx, tenantID, deviceID, userID, finalAlarmItems)
		if dbErr != nil {
			s.logger.Error("[RADAR_WRITE] Failed to update database",
				zap.String("tenant_id", tenantID),
				zap.String("device_id", deviceID),
				zap.Error(dbErr),
			)
			result.DBWriteSuccess = false
			return result, fmt.Errorf("failed to update database: %w", dbErr)
		}

		result.DBWriteSuccess = true
		s.logger.Info("[RADAR_WRITE] Device and database update successful",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.String("device_uid", device.DeviceUID),
			zap.Int("failed_count", len(result.FailedAlarmTypes)),
		)
	}

	return result, nil
}

// UpdateSleepadMonitorSettings 更新 Sleepad 设备监控配置
func (s *deviceMonitorSettingsService) UpdateSleepadMonitorSettings(ctx context.Context, tenantID, deviceID, userID string, alarmItems []alarm.AlarmItem, progressCallback ProgressCallback) (success bool, noChange bool, err error) {
	deviceType := "sleepace"

	// 验证设备存在
	device, err := s.devicesRepo.GetDevice(ctx, tenantID, deviceID)
	if err != nil {
		return false, false, fmt.Errorf("failed to get device: %w", err)
	}

	// 验证设备类型匹配
	if !device.DeviceType.Valid {
		return false, false, fmt.Errorf("device has no device_type")
	}
	if !s.isDeviceTypeMatch(deviceType, device.DeviceType.String) {
		return false, false, fmt.Errorf("device type mismatch: expected %s (from request), got %s (from device_store)", deviceType, device.DeviceType.String)
	}

	// 获取现有配置
	existingAlarmItems, err := s.getExistingAlarmItems(ctx, tenantID, deviceID, deviceType)
	if err != nil {
		return false, false, err
	}

	// Sleepad 设备需要写入的 AlarmType（TODO: 根据实际需求定义）
	deviceWriteAlarmTypes := map[string]bool{
		alarm.MonitoringMode:          true,
		alarm.Fall:                    true,
		alarm.PostureDetection:        true,
		alarm.BedSitUp:                true,
		alarm.SittingOnGround:         true,
		alarm.AbnormalRespiratoryRate: true,
		alarm.AbnormalHeartRate:       true,
		alarm.VitalsWeak:              true,
	}
	// Sleepad 设备会返回的 AlarmType（TODO: 根据实际 sleepad 设备返回项定义）
	deviceReturnAlarmTypes := map[string]bool{
		// Sleepad 设备返回的项与 radar 不同，需要根据实际硬件返回定义
		// 目前先使用空 map，后续根据实际需求补充
	}

	_, noChange, _, err = s.compareAndLogChanges(ctx, tenantID, deviceID, deviceType, existingAlarmItems, alarmItems, deviceWriteAlarmTypes, deviceReturnAlarmTypes)
	return true, noChange, err
}

// getExistingAlarmItems 获取现有配置（如果不存在，使用默认配置创建）
func (s *deviceMonitorSettingsService) getExistingAlarmItems(ctx context.Context, tenantID, deviceID, deviceType string) ([]alarm.AlarmItem, error) {
	existingAlarmDevice, err := s.alarmDeviceRepo.GetAlarmDevice(ctx, tenantID, deviceID)
	if err != nil {
		// 检查是否是"未找到"错误
		isNotFound := err == sql.ErrNoRows ||
			strings.Contains(err.Error(), "no rows in result set") ||
			strings.Contains(err.Error(), "not found")

		if isNotFound {
			// 如果配置不存在，使用默认配置
			s.logger.Info("Alarm device not found, using default settings",
				zap.String("tenant_id", tenantID),
				zap.String("device_id", deviceID),
				zap.String("device_type", deviceType),
			)
			existingAlarmItems, err := s.buildDeviceAlarmItemsFromCloudOrDefault(ctx, tenantID, deviceType)
			if err != nil {
				return nil, fmt.Errorf("failed to build default alarm items: %w", err)
			}
			return existingAlarmItems, nil
		}
		// 其他错误直接返回
		return nil, fmt.Errorf("failed to get alarm device: %w", err)
	}

	// 转换为 AlarmItem 数组
	existingAlarmItems, err := s.convertMonitorConfigToAlarmItems(existingAlarmDevice.MonitorConfig, deviceType)
	if err != nil {
		return nil, fmt.Errorf("failed to convert existing monitor config: %w", err)
	}
	return existingAlarmItems, nil
}

// compareAndLogChanges 比对配置变化并记录日志（通用比对逻辑）
// 返回：success, noChange, hasDeviceWriteChanges, error
func (s *deviceMonitorSettingsService) compareAndLogChanges(ctx context.Context, tenantID, deviceID, deviceType string, existingAlarmItems, alarmItems []alarm.AlarmItem, deviceWriteAlarmTypes, deviceReturnAlarmTypes map[string]bool) (success bool, noChange bool, hasDeviceWriteChanges bool, err error) {
	// 构建现有配置的索引（按 AlarmType）
	existingMap := make(map[string]alarm.AlarmItem)
	for _, item := range existingAlarmItems {
		if item.AlarmType != "" {
			existingMap[item.AlarmType] = item
		}
	}

	// 构建新配置的索引（按 AlarmType）
	newMap := make(map[string]alarm.AlarmItem)
	for _, item := range alarmItems {
		if item.AlarmType != "" {
			newMap[item.AlarmType] = item
		}
	}

	// 逐项比对，记录变化项及是否需要发送给设备
	// 对于设备返回的项，单独记录与前端不同的项
	changedItems := make([]struct {
		AlarmType        string
		Changed          bool
		NeedDeviceWrite  bool
		IsDeviceReturn   bool                      // 是否为设备返回的项（radar 或 sleepad）
		ComparisonResult AlarmItemComparisonResult // 详细的比对结果
		ExistingValue    interface{}
		NewValue         interface{}
	}, 0)

	// 检查新配置中的每个 AlarmItem
	for alarmType, newItem := range newMap {
		existingItem, exists := existingMap[alarmType]
		isChanged := false
		var existingValue, newValue interface{}
		var comparisonResult AlarmItemComparisonResult

		if !exists {
			// 新配置中有现有配置没有的 AlarmType，认为有变化
			isChanged = true
			newValue = newItem
			existingValue = nil
			// 新项，所有字段都视为变化
			comparisonResult = AlarmItemComparisonResult{
				AlarmParamsChanged: true,
				IsEnabledChanged:   true,
				AlarmLevelChanged:  true,
				HasAnyChange:       true,
			}
		} else {
			// 比对每个字段：alarm_params, isEnabled, alarm_level
			comparisonResult = s.compareAlarmItem(existingItem, newItem)
			if comparisonResult.HasAnyChange {
				isChanged = true
				existingValue = existingItem
				newValue = newItem
			}
		}

		if isChanged {
			needDeviceWrite := deviceWriteAlarmTypes[alarmType]
			isDeviceReturn := deviceReturnAlarmTypes[alarmType]

			changedItems = append(changedItems, struct {
				AlarmType        string
				Changed          bool
				NeedDeviceWrite  bool
				IsDeviceReturn   bool
				ComparisonResult AlarmItemComparisonResult
				ExistingValue    interface{}
				NewValue         interface{}
			}{
				AlarmType:        alarmType,
				Changed:          true,
				NeedDeviceWrite:  needDeviceWrite,
				IsDeviceReturn:   isDeviceReturn,
				ComparisonResult: comparisonResult,
				ExistingValue:    existingValue,
				NewValue:         newValue,
			})

			// 如果是设备返回的项，且与前端不同，单独记录
			if isDeviceReturn {
				s.logger.Info("[SAVE_SETTINGS] Device return item differs from frontend",
					zap.String("device_type", deviceType),
					zap.String("alarm_type", alarmType),
					zap.Bool("alarm_params_changed", comparisonResult.AlarmParamsChanged),
					zap.Bool("is_enabled_changed", comparisonResult.IsEnabledChanged),
					zap.Bool("alarm_level_changed", comparisonResult.AlarmLevelChanged),
					zap.Any("existing_value", existingValue),
					zap.Any("new_value", newValue),
				)
			}
		}
	}

	// 检查是否有现有配置中有但新配置中没有的 AlarmType
	for alarmType, existingItem := range existingMap {
		if _, exists := newMap[alarmType]; !exists {
			needDeviceWrite := deviceWriteAlarmTypes[alarmType]
			isDeviceReturn := deviceReturnAlarmTypes[alarmType]

			changedItems = append(changedItems, struct {
				AlarmType        string
				Changed          bool
				NeedDeviceWrite  bool
				IsDeviceReturn   bool
				ComparisonResult AlarmItemComparisonResult
				ExistingValue    interface{}
				NewValue         interface{}
			}{
				AlarmType:       alarmType,
				Changed:         true,
				NeedDeviceWrite: needDeviceWrite,
				IsDeviceReturn:  isDeviceReturn,
				ComparisonResult: AlarmItemComparisonResult{
					AlarmParamsChanged: true,
					IsEnabledChanged:   true,
					AlarmLevelChanged:  true,
					HasAnyChange:       true,
				},
				ExistingValue: existingItem,
				NewValue:      nil,
			})
		}
	}

	// 提取变化的 AlarmType 列表（用于后续处理）
	changedAlarmTypes := make([]string, 0, len(changedItems))
	for _, item := range changedItems {
		if item.Changed {
			changedAlarmTypes = append(changedAlarmTypes, item.AlarmType)
			if item.NeedDeviceWrite {
				hasDeviceWriteChanges = true
			}
		}
	}

	// 2.1. 如果配置无变化，直接返回 "no Change"
	if len(changedItems) == 0 {
		s.logger.Info("Device monitor settings unchanged, skipping update",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.String("device_type", deviceType),
		)
		return true, true, false, nil
	}

	// 记录比对结果
	changedItemsForLog := make([]map[string]interface{}, 0, len(changedItems))
	var deviceReturnDiffItems []map[string]interface{} // 单独列出设备返回项与前端不同的项

	for _, item := range changedItems {
		itemLog := map[string]interface{}{
			"alarm_type":           item.AlarmType,
			"changed":              item.Changed,
			"need_device_write":    item.NeedDeviceWrite,
			"alarm_params_changed": item.ComparisonResult.AlarmParamsChanged,
			"is_enabled_changed":   item.ComparisonResult.IsEnabledChanged,
			"alarm_level_changed":  item.ComparisonResult.AlarmLevelChanged,
		}
		changedItemsForLog = append(changedItemsForLog, itemLog)

		// 如果是设备返回的项，且与前端不同，单独列出
		if item.IsDeviceReturn {
			deviceReturnDiffItems = append(deviceReturnDiffItems, itemLog)
		}
	}

	s.logger.Info("[SAVE_SETTINGS] Comparison completed - all changed items identified",
		zap.String("tenant_id", tenantID),
		zap.String("device_id", deviceID),
		zap.String("device_type", deviceType),
		zap.Int("total_changed_items", len(changedItems)),
		zap.Bool("has_device_write_changes", hasDeviceWriteChanges),
		zap.Any("changed_items", changedItemsForLog),
	)

	// 单独列出设备返回项与前端不同的项
	if len(deviceReturnDiffItems) > 0 {
		s.logger.Info("[SAVE_SETTINGS] Device return items differ from frontend - listed separately",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.String("device_type", deviceType),
			zap.Int("device_return_diff_count", len(deviceReturnDiffItems)),
			zap.Any("device_return_diff_items", deviceReturnDiffItems),
		)
	}

	// 3. 暂时停止：只完成比对，不执行设备写入和数据库更新
	// TODO: 后续步骤：
	//   - 根据比对结果，决定需要发送给设备的属性
	//   - 写入设备
	//   - 更新数据库
	//   - 保存 config_versions

	s.logger.Info("[SAVE_SETTINGS] Comparison step completed - stopping here (device write and DB update logic cleared)",
		zap.String("tenant_id", tenantID),
		zap.String("device_id", deviceID),
		zap.String("device_type", deviceType),
		zap.Strings("changed_alarm_types", changedAlarmTypes),
		zap.Bool("has_device_write_changes", hasDeviceWriteChanges),
	)

	// 暂时返回成功，但不执行实际更新
	return true, false, hasDeviceWriteChanges, nil
}

// RadarWriteResult 雷达写入结果，包含每个 alarm_type 的执行结果
type RadarWriteResult struct {
	Results map[string]string // key: alarm_type, value: "200" 或 "Fail"
}

// UpdateRadarResult 更新雷达配置的结果
type UpdateRadarResult struct {
	DeviceWriteSuccess bool              // 设备写入是否成功
	DBWriteSuccess     bool              // 数据库写入是否成功
	FailedAlarmTypes   []alarm.AlarmItem // 失败的 alarm_type 列表（使用旧值）
	NoChange           bool              // 是否有变化
}

// radarWrite 向 qinglan_client 发送写入完整的指令（工作模式、跌倒和呼吸心率参数设置）
// 直接传递 _alarm_items_json，wisefido-qinglan 会自动构建 fall_param 和 heart_breath_param
// qinglan_client 会自动分组、排序和时间间隔（100ms）
// 返回每个 alarm_type 的执行结果：200 成功，Fail 失败
func (s *deviceMonitorSettingsService) radarWrite(ctx context.Context, deviceUID string, alarmItems []alarm.AlarmItem, progressCallback ProgressCallback) (*RadarWriteResult, error) {
	if s.qinglanClient == nil {
		return nil, fmt.Errorf("qinglan client is not initialized")
	}

	if deviceUID == "" {
		return nil, fmt.Errorf("device_uid is required")
	}

	if len(alarmItems) == 0 {
		return nil, fmt.Errorf("alarm_items is empty")
	}

	// 将 AlarmItems 序列化为 JSON 字符串
	alarmItemsJSON, err := json.Marshal(alarmItems)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal alarm items to JSON: %w", err)
	}

	// 直接传递 _alarm_items_json，wisefido-qinglan 会自动处理
	// wisefido-qinglan 会从 _alarm_items_json 构建：
	//   - radar_func_ctrl (从 MonitoringMode)
	//   - fall_param (从 Fall, SittingOnGround, PostureDetection 等)
	//   - heart_breath_param (从 AbnormalHeartRate, AbnormalRespiratoryRate, VitalsWeak)
	properties := map[string]interface{}{
		"_alarm_items_json": string(alarmItemsJSON),
	}

	s.logger.Info("[RADAR_WRITE] Sending alarm items to device",
		zap.String("device_uid", deviceUID),
		zap.Int("alarm_items_count", len(alarmItems)),
		zap.String("alarm_items_json_length", fmt.Sprintf("%d bytes", len(alarmItemsJSON))),
	)

	// 通知前端开始写入
	if progressCallback != nil {
		progressCallback(10, "开始写入设备配置...")
	}

	// 调用 qinglan_client.SetDeviceProperties 写入设备
	// qinglan_client 会自动分组、排序和时间间隔（100ms）
	// wisefido-qinglan 会从 _alarm_items_json 构建设备属性
	_, err = s.qinglanClient.SetDeviceProperties(ctx, deviceUID, properties)

	// 构建每个 alarm_type 的执行结果
	result := &RadarWriteResult{
		Results: make(map[string]string),
	}

	if err != nil {
		s.logger.Error("[RADAR_WRITE] Failed to write device properties",
			zap.String("device_uid", deviceUID),
			zap.Int("alarm_items_count", len(alarmItems)),
			zap.Error(err),
		)
		// 所有 alarm_type 标记为失败
		for _, item := range alarmItems {
			result.Results[item.AlarmType] = "Fail"
		}
		// 通知前端写入失败
		if progressCallback != nil {
			progressCallback(0, fmt.Sprintf("设备配置写入失败: %v", err))
		}
		return result, fmt.Errorf("failed to write device properties: %w", err)
	}

	s.logger.Info("[RADAR_WRITE] Successfully wrote device properties",
		zap.String("device_uid", deviceUID),
		zap.Int("alarm_items_count", len(alarmItems)),
	)

	// 所有 alarm_type 标记为成功
	for _, item := range alarmItems {
		result.Results[item.AlarmType] = "200"
	}

	// 通知前端写入成功
	if progressCallback != nil {
		progressCallback(50, "设备配置写入成功")
	}

	return result, nil
}

// updateAlarmDeviceDB 更新 alarm_device 表和 config_version
func (s *deviceMonitorSettingsService) updateAlarmDeviceDB(ctx context.Context, tenantID, deviceID, userID string, alarmItems []alarm.AlarmItem) error {
	// 转换为 monitor_config JSON
	monitorConfigJSON, err := s.convertAlarmItemsToMonitorConfig(alarmItems)
	if err != nil {
		return fmt.Errorf("failed to convert alarm items to monitor config: %w", err)
	}

	// 构建 AlarmDevice 对象
	alarmDevice := &domain.AlarmDevice{
		DeviceID:      deviceID,
		TenantID:      tenantID,
		MonitorConfig: monitorConfigJSON,
	}

	// 使用事务更新 alarm_device 和 config_version
	if s.db == nil {
		return fmt.Errorf("database connection is not available")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 更新 alarm_device
	if err := s.alarmDeviceRepo.UpsertAlarmDevice(ctx, tenantID, deviceID, alarmDevice); err != nil {
		return fmt.Errorf("failed to upsert alarm device: %w", err)
	}

	// 创建 config_version 记录（用于审计）
	if s.configVersionsRepo != nil {
		now := time.Now()
		var createdBy *string
		if userID != "" {
			createdBy = &userID
		}
		configVersion := &domain.ConfigVersion{
			TenantID:        tenantID,
			EntityID:        deviceID, // device_id 作为 entity_id
			CurrentEntityID: deviceID,
			ConfigType:      "alarm_device", // 使用 alarm_device 作为配置类型
			ConfigData:      monitorConfigJSON,
			ValidFrom:       now,
			ValidTo:         nil, // NULL 表示当前仍生效
			CreatedBy:       createdBy,
			ChangeReason:    "Update device monitor settings",
		}
		if _, err := s.configVersionsRepo.CreateConfigVersion(ctx, tenantID, configVersion); err != nil {
			s.logger.Warn("[UPDATE_DB] Failed to create config version",
				zap.String("tenant_id", tenantID),
				zap.String("device_id", deviceID),
				zap.Error(err),
			)
			// 不返回错误，因为 config_version 只是审计记录
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Info("[UPDATE_DB] Successfully updated alarm_device and config_version",
		zap.String("tenant_id", tenantID),
		zap.String("device_id", deviceID),
		zap.String("user_id", userID),
	)

	// 异步发送设备告警配置变更消息到 config:device.alarm.setting:stream
	go func() {
		publishCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// 获取设备 UID（用于消费端识别设备）
		device, err := s.devicesRepo.GetDevice(publishCtx, tenantID, deviceID)
		if err != nil {
			s.logger.Warn("Failed to get device UID for publishing alarm setting message",
				zap.String("device_id", deviceID),
				zap.Error(err),
			)
			return
		}

		deviceUID := device.DeviceUID
		if deviceUID == "" {
			s.logger.Warn("Device UID is empty, cannot publish alarm setting message",
				zap.String("device_id", deviceID),
			)
			return
		}

		// 构建 settingData（留空，wisefido-qinglan 从数据库直接查询报警使能配置）
		settingData := map[string]interface{}{}

		// 发送消息
		if err := s.configPublisher.PublishAlarmDeviceMessage(
			publishCtx,
			tenantID,
			deviceID,
			deviceUID,
			"updated", // settingType
			settingData,
		); err != nil {
			s.logger.Warn("Failed to publish device alarm setting message",
				zap.String("device_id", deviceID),
				zap.String("device_uid", deviceUID),
				zap.Error(err),
			)
		}
	}()

	return nil
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

// AlarmItemComparisonResult 比对结果，记录每个字段是否有变化
type AlarmItemComparisonResult struct {
	AlarmParamsChanged bool
	IsEnabledChanged   bool
	AlarmLevelChanged  bool
	HasAnyChange       bool
}

// compareAlarmItem 比对两个 AlarmItem，返回详细的比对结果
// 对每个 alarm_type 进行3个比较：alarm_params, isEnabled, alarm_level
func (s *deviceMonitorSettingsService) compareAlarmItem(existing, new alarm.AlarmItem) AlarmItemComparisonResult {
	result := AlarmItemComparisonResult{}

	// 1. 比对 AlarmParams
	result.AlarmParamsChanged = !s.compareAlarmParams(existing.AlarmParams, new.AlarmParams)

	// 2. 比对 IsEnabled（指针比较）
	result.IsEnabledChanged = !compareIntPtr(existing.IsEnabled, new.IsEnabled)

	// 3. 比对 AlarmLevel（指针比较）
	result.AlarmLevelChanged = !compareStringPtr(existing.AlarmLevel, new.AlarmLevel)

	// 是否有任何变化
	result.HasAnyChange = result.AlarmParamsChanged || result.IsEnabledChanged || result.AlarmLevelChanged

	return result
}

// compareAlarmItemSimple 比对两个 AlarmItem 是否完全相同（向后兼容）
func (s *deviceMonitorSettingsService) compareAlarmItemSimple(existing, new alarm.AlarmItem) bool {
	result := s.compareAlarmItem(existing, new)
	return !result.HasAnyChange
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

// compareValues 深度比对两个值是否相同
// 支持字符串 "8" 与数字 8 的比较（radar 硬件可能返回字符串）
func (s *deviceMonitorSettingsService) compareValues(existing, new interface{}) bool {
	// 如果类型相同，直接比较
	if existing == nil && new == nil {
		return true
	}
	if existing == nil || new == nil {
		return false
	}

	// 尝试转换为相同类型后比较
	existingVal := s.normalizeValue(existing)
	newVal := s.normalizeValue(new)

	// 如果都是数字类型，直接比较
	if existingNum, ok1 := existingVal.(float64); ok1 {
		if newNum, ok2 := newVal.(float64); ok2 {
			return existingNum == newNum
		}
	}

	// 如果都是字符串，直接比较
	if existingStr, ok1 := existingVal.(string); ok1 {
		if newStr, ok2 := newVal.(string); ok2 {
			return existingStr == newStr
		}
	}

	// 其他情况使用 JSON 序列化比对
	existingJSON, _ := json.Marshal(existing)
	newJSON, _ := json.Marshal(new)
	return string(existingJSON) == string(newJSON)
}

// normalizeValue 将值标准化为可比较的类型
// 处理字符串 "8" 与数字 8 的情况
func (s *deviceMonitorSettingsService) normalizeValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case string:
		// 尝试将字符串转换为数字
		if intVal, err := strconv.Atoi(val); err == nil {
			return float64(intVal)
		}
		if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
			return floatVal
		}
		return val
	case int:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	default:
		return v
	}
}

// compareIntPtr 比对两个 *int 是否相同
func compareIntPtr(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// compareStringPtr 比对两个 *string 是否相同
func compareStringPtr(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
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

		// 比对每个字段：alarm_params, isEnabled, alarm_level
		comparisonResult := s.compareAlarmItem(existingItem, newItem)
		if comparisonResult.HasAnyChange {
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

// isDeviceTypeMatch 判断两个设备类型是否匹配（支持多种变体）
func (s *deviceMonitorSettingsService) isDeviceTypeMatch(type1, type2 string) bool {
	normalized1 := s.normalizeDeviceType(type1)
	normalized2 := s.normalizeDeviceType(type2)
	return normalized1 != "" && normalized1 == normalized2
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

// syncDeviceValuesToDB 异步查询设备并更新数据库（只更新alarm_params或enabled，updated_by=null，不更新config_version）
func (s *deviceMonitorSettingsService) syncDeviceValuesToDB(ctx context.Context, tenantID, deviceID, deviceUID, deviceType string, baselineAlarmItems []alarm.AlarmItem, dbExists bool) {
	s.logger.Info("[SYNC_DEVICE] Starting async device query and DB sync",
		zap.String("tenant_id", tenantID),
		zap.String("device_id", deviceID),
		zap.String("device_uid", deviceUID),
		zap.String("device_type", deviceType),
	)

	// 先检查设备在线状态，只有 online 才查询设备属性
	if err := s.CheckDeviceOnlineStatus(ctx, deviceUID); err != nil {
		s.logger.Info("[SYNC_DEVICE] Device is not online, skipping device query",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.String("device_uid", deviceUID),
			zap.String("device_type", deviceType),
			zap.Error(err),
		)
		return
	}

	// 创建独立的超时context（5秒超时，避免长时间阻塞）
	deviceReadCtx, deviceReadCancel := context.WithTimeout(ctx, 5*time.Second)
	defer deviceReadCancel()

	// 读取设备属性（工作模式、跌倒参数、呼吸心率参数）
	deviceProps, deviceErr := s.qinglanClient.GetDeviceProperties(deviceReadCtx, deviceUID, []string{
		"radar_func_ctrl",    // 工作模式
		"fall_param",         // 跌倒参数
		"heart_breath_param", // 呼吸心率参数
	})

	// 检查是否超时或取消
	if deviceErr != nil {
		if deviceReadCtx.Err() == context.DeadlineExceeded {
			s.logger.Warn("[SYNC_DEVICE] Device query: TIMEOUT (5s)",
				zap.String("tenant_id", tenantID),
				zap.String("device_id", deviceID),
				zap.String("device_uid", deviceUID),
			)
		} else if deviceReadCtx.Err() == context.Canceled {
			s.logger.Warn("[SYNC_DEVICE] Device query: CANCELED",
				zap.String("tenant_id", tenantID),
				zap.String("device_id", deviceID),
				zap.String("device_uid", deviceUID),
			)
		} else {
			s.logger.Warn("[SYNC_DEVICE] Device query: FAILED",
				zap.String("tenant_id", tenantID),
				zap.String("device_id", deviceID),
				zap.String("device_uid", deviceUID),
				zap.Error(deviceErr),
			)
		}
		return
	}

	if deviceProps == nil || len(deviceProps) == 0 {
		s.logger.Info("[SYNC_DEVICE] Device query: SUCCESS but no properties returned",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.String("device_uid", deviceUID),
		)
		return
	}

	s.logger.Info("[SYNC_DEVICE] Device query: SUCCESS",
		zap.String("tenant_id", tenantID),
		zap.String("device_id", deviceID),
		zap.String("device_uid", deviceUID),
		zap.Int("props_count", len(deviceProps)),
	)

	// 将设备属性解码为 AlarmItem（基于baseline）
	deviceAlarmItems, decodeErr := decode.DecodeDevicePropsToAlarmItems(baselineAlarmItems, deviceProps)
	if decodeErr != nil {
		s.logger.Warn("[SYNC_DEVICE] Failed to decode device properties to alarm items",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.String("device_uid", deviceUID),
			zap.Error(decodeErr),
		)
		return
	}

	// 比对设备值和baseline值，找出需要更新的项
	// 只更新设备相关的项（alarm_params或enabled），其他项保持baseline值
	changedTypes := s.getChangedAlarmTypes(baselineAlarmItems, deviceAlarmItems)
	if len(changedTypes) == 0 {
		s.logger.Debug("[SYNC_DEVICE] Device values match baseline, no sync needed",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.String("device_uid", deviceUID),
		)
		return
	}

	s.logger.Info("[SYNC_DEVICE] Device values differ from baseline, syncing to database",
		zap.String("tenant_id", tenantID),
		zap.String("device_id", deviceID),
		zap.String("device_uid", deviceUID),
		zap.Strings("changed_types", changedTypes),
	)

	// 部分更新数据库：只更新alarm_params或enabled，不更新config_version
	if err := s.updateAlarmDevicePartial(ctx, tenantID, deviceID, baselineAlarmItems, deviceAlarmItems); err != nil {
		s.logger.Warn("[SYNC_DEVICE] Failed to update database with device values",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.String("device_uid", deviceUID),
			zap.Error(err),
		)
		return
	}

	s.logger.Info("[SYNC_DEVICE] Successfully synced device values to database",
		zap.String("tenant_id", tenantID),
		zap.String("device_id", deviceID),
		zap.String("device_uid", deviceUID),
		zap.Strings("changed_types", changedTypes),
	)
}

// updateAlarmDevicePartial 部分更新alarm_device表（只更新alarm_params或enabled，updated_by=null，不更新config_version）
func (s *deviceMonitorSettingsService) updateAlarmDevicePartial(ctx context.Context, tenantID, deviceID string, baselineAlarmItems, deviceAlarmItems []alarm.AlarmItem) error {
	if s.db == nil {
		return fmt.Errorf("database connection is not available")
	}

	// 构建baseline和device的索引
	baselineMap := make(map[string]alarm.AlarmItem)
	for _, item := range baselineAlarmItems {
		if item.AlarmType != "" {
			baselineMap[item.AlarmType] = item
		}
	}

	deviceMap := make(map[string]alarm.AlarmItem)
	for _, item := range deviceAlarmItems {
		if item.AlarmType != "" {
			deviceMap[item.AlarmType] = item
		}
	}

	// 读取现有的monitor_config
	alarmDevice, err := s.alarmDeviceRepo.GetAlarmDevice(ctx, tenantID, deviceID)
	if err != nil {
		// 如果不存在，创建新记录
		monitorConfigJSON, err := s.convertAlarmItemsToMonitorConfig(deviceAlarmItems)
		if err != nil {
			return fmt.Errorf("failed to convert device alarm items to monitor config: %w", err)
		}

		tx, txErr := s.db.BeginTx(ctx, nil)
		if txErr != nil {
			return fmt.Errorf("failed to begin transaction: %w", txErr)
		}
		defer tx.Rollback()

		insertQuery := `
			INSERT INTO alarm_device (
				device_id,
				tenant_id,
				monitor_config,
				updated_at,
				updated_by
			) VALUES ($1, $2, $3::jsonb, $4, $5)
		`
		_, execErr := tx.ExecContext(ctx, insertQuery, deviceID, tenantID, string(monitorConfigJSON), time.Now().UTC(), nil)
		if execErr != nil {
			return fmt.Errorf("failed to insert alarm device: %w", execErr)
		}

		if commitErr := tx.Commit(); commitErr != nil {
			return fmt.Errorf("failed to commit transaction: %w", commitErr)
		}

		return nil
	}

	// 解析现有的monitor_config
	var monitorConfig map[string]interface{}
	if len(alarmDevice.MonitorConfig) > 0 {
		if err := json.Unmarshal(alarmDevice.MonitorConfig, &monitorConfig); err != nil {
			return fmt.Errorf("failed to unmarshal existing monitor config: %w", err)
		}
	} else {
		monitorConfig = make(map[string]interface{})
	}

	// 获取alarms对象
	alarms, ok := monitorConfig["alarms"].(map[string]interface{})
	if !ok {
		alarms = make(map[string]interface{})
		monitorConfig["alarms"] = alarms
	}

	// 只更新设备相关的项（alarm_params或enabled）
	// 如果设备返回了alarm_params，更新alarm_params；如果没有，只更新enabled
	for alarmType, deviceItem := range deviceMap {
		baselineItem, exists := baselineMap[alarmType]
		if !exists {
			continue // 跳过baseline中没有的项
		}

		// 检查是否有变化
		comparisonResult := s.compareAlarmItem(baselineItem, deviceItem)
		if !comparisonResult.HasAnyChange {
			continue // 没有变化，跳过
		}

		// 获取或创建该alarm_type的配置
		alarmConfig, ok := alarms[alarmType].(map[string]interface{})
		if !ok {
			alarmConfig = make(map[string]interface{})
			alarms[alarmType] = alarmConfig
		}

		// 更新alarm_params（如果设备返回了alarm_params）
		if comparisonResult.AlarmParamsChanged && deviceItem.AlarmParams != nil && len(deviceItem.AlarmParams) > 0 {
			alarmConfig["threshold"] = deviceItem.AlarmParams
		}

		// 更新enabled（如果设备返回了enabled）
		if comparisonResult.IsEnabledChanged && deviceItem.IsEnabled != nil {
			alarmConfig["enabled"] = *deviceItem.IsEnabled == alarm.IsEnabledOn
		}

		// 保持level不变（不更新alarm_level）
		if _, ok := alarmConfig["level"]; !ok && baselineItem.AlarmLevel != nil {
			alarmConfig["level"] = *baselineItem.AlarmLevel
		}
	}

	// 序列化更新后的monitor_config
	updatedMonitorConfigJSON, err := json.Marshal(monitorConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal updated monitor config: %w", err)
	}

	// 使用事务更新数据库（只更新monitor_config，不更新config_version）
	tx, txErr := s.db.BeginTx(ctx, nil)
	if txErr != nil {
		return fmt.Errorf("failed to begin transaction: %w", txErr)
	}
	defer tx.Rollback()

	upsertQuery := `
		INSERT INTO alarm_device (
			device_id,
			tenant_id,
			monitor_config,
			updated_at,
			updated_by
		) VALUES ($1, $2, $3::jsonb, $4, $5)
		ON CONFLICT (device_id) DO UPDATE SET
			monitor_config = EXCLUDED.monitor_config,
			updated_at = EXCLUDED.updated_at,
			updated_by = EXCLUDED.updated_by
	`
	// updated_by设置为NULL（表示系统自动同步）
	_, execErr := tx.ExecContext(ctx, upsertQuery, deviceID, tenantID, string(updatedMonitorConfigJSON), time.Now().UTC(), nil)
	if execErr != nil {
		return fmt.Errorf("failed to update alarm device: %w", execErr)
	}

	// 注意：不更新config_version（不创建审计记录）
	// 因为这是系统自动同步，不是用户操作

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("failed to commit transaction: %w", commitErr)
	}

	return nil
}
