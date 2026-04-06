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
	sleepaceGateway    *SleepaceGatewayClient // wisefido-sleepace 网关（厂家 HTTP 代理 + 下发硬件配置）
	configPublisher    ConfigPublisher        // 配置消息发布器
	qinglanClient      *QinglanClient         // 雷达设备仅经此客户端：查询状态/属性、下发属性（工作模式、跌倒/呼吸心率）
	db                 *sql.DB                // 用于事务操作
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
		configPublisher:    configPublisher,
		logger:             logger,
	}
}

// SetSleepaceGatewayClient 设置 wisefido-sleepace 网关客户端（下发 Sleepace 硬件配置）
func (s *deviceMonitorSettingsService) SetSleepaceGatewayClient(client *SleepaceGatewayClient) {
	s.sleepaceGateway = client
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
		return nil, fmt.Errorf("invalid device_type: %s (must be 'sleepad' or 'radar')", deviceType)
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

		// alarm_cloud only provides AlarmLevel for initialization; IsEnabled stays from defaults
		if deviceTypeKey != "" {
			if typeMap, ok := deviceAlarmsMap[deviceTypeKey]; ok {
				if level, exists := typeMap[item.AlarmType]; exists {
					level = strings.TrimSpace(level)
					level = strings.Trim(level, `"`)
					newItem.AlarmLevel = &level
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

// fullMonitorConfigPayload 全量存储格式：保存完整 AlarmItem[]，避免后续默认值变更覆盖客户已保存配置
type fullMonitorConfigPayload struct {
	Items []alarm.AlarmItem `json:"items"`
}

// marshalAlarmItemsToFullConfig 将 AlarmItem 数组序列化为全量 monitor_config（{"items": [...]}）
func (s *deviceMonitorSettingsService) marshalAlarmItemsToFullConfig(alarmItems []alarm.AlarmItem) ([]byte, error) {
	if alarmItems == nil {
		alarmItems = []alarm.AlarmItem{}
	}
	return json.Marshal(fullMonitorConfigPayload{Items: alarmItems})
}

// parseMonitorConfigToAlarmItems 仅支持全量格式 {"items": [...]}。非全量或解析失败返回空切片，不自动用默认值；客户需用「引入默认值」再保存。
func (s *deviceMonitorSettingsService) parseMonitorConfigToAlarmItems(monitorConfigJSON []byte, deviceType string) ([]alarm.AlarmItem, error) {
	if len(monitorConfigJSON) == 0 {
		return nil, fmt.Errorf("empty monitor_config")
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(monitorConfigJSON, &raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal monitor_config: %w", err)
	}
	if _, hasItems := raw["items"]; !hasItems {
		return []alarm.AlarmItem{}, nil
	}
	var wrapper fullMonitorConfigPayload
	if err := json.Unmarshal(monitorConfigJSON, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to unmarshal full monitor_config: %w", err)
	}
	return wrapper.Items, nil
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
	if deviceType != "sleepad" && deviceType != "radar" {
		return nil, fmt.Errorf("invalid device_type: %s (must be 'sleepad' or 'radar')", deviceType)
	}

	// 验证设备存在
	device, err := s.devicesRepo.GetDevice(ctx, tenantID, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	// 验证设备类型匹配（安全性检查：防止恶意修改 URL 参数）
	// deviceType 是前端传入的 "sleepad" 或 "radar"
	// device.DeviceType 是从 devices 表 JOIN device_store 获取的 device_type（如 "Sleepad", "Radar"）
	if !device.DeviceType.Valid {
		return nil, fmt.Errorf("device has no device_type")
	}

	if !s.isDeviceTypeMatch(deviceType, device.DeviceType.String) {
		return nil, fmt.Errorf("device type mismatch: expected %s (from request), got %s (from device_store)", deviceType, device.DeviceType.String)
	}

	// Read DB items (needed for AlarmLevel regardless of hardware query result)
	var dbAlarmItems []alarm.AlarmItem
	alarmDevice, dbErr := s.alarmDeviceRepo.GetAlarmDevice(ctx, tenantID, deviceID)
	if dbErr == nil {
		converted, convertErr := s.parseMonitorConfigToAlarmItems(alarmDevice.MonitorConfig, deviceType)
		if convertErr != nil {
			s.logger.Warn("[GET_SETTINGS] DB config conversion failed",
				zap.String("device_id", deviceID),
				zap.Error(convertErr),
			)
		} else {
			dbAlarmItems = converted
		}
	}

	// For sleepad: use DB (or defaults) as baseline for display; overlay hardware IsEnabled/AlarmParams only. AlarmLevel always from DB. If HW differs from DB, sync merged to DB.
	if deviceType == "sleepad" && s.sleepaceGateway != nil && device.DeviceCode.Valid && device.DeviceCode.String != "" {
		s.logger.Info("[GET_SETTINGS] querying sleepad hardware for alarm config",
			zap.String("device_id", deviceID),
			zap.String("device_code", device.DeviceCode.String),
		)
		hwData, hwErr := s.sleepaceGateway.GetAlarmConfig(ctx, device.DeviceID, device.DeviceCode.String)
		if hwErr == nil && hwData != nil {
			hwItems := ConvertHardwareResponseToAlarmItems(hwData)
			if len(hwItems) > 0 {
				baseline := dbAlarmItems
				merged := MergeHardwareIntoBaseline(baseline, hwItems)
				s.logger.Info("[GET_SETTINGS] returning DB baseline + hardware overlay (AlarmLevel from DB)",
					zap.String("device_id", deviceID),
					zap.Int("baseline_items", len(baseline)),
					zap.Int("hw_items", len(hwItems)),
				)
				changedTypes := s.getChangedAlarmTypes(baseline, merged)
				if len(changedTypes) > 0 {
					go s.syncSleepadHardwareToDB(context.Background(), tenantID, deviceID, baseline, merged)
				}
				// 语雀 getconfig：回填设备当前 realtimeDataInterval
				if interval, cfgErr := s.sleepaceGateway.GetDeviceConfig(ctx, device.DeviceID, device.DeviceCode.String); cfgErr == nil && interval > 0 {
					for i := range merged {
						if merged[i].AlarmType == alarm.SleepadSetting && merged[i].AlarmParams != nil {
							merged[i].AlarmParams["realtime_interval"] = interval
							break
						}
					}
				}
				merged = s.overlaySleepadLightModeFromHardware(ctx, device.DeviceCode.String, merged)
				return merged, nil
			}
		}
		if hwErr != nil {
			s.logger.Warn("[GET_SETTINGS] hardware query failed, falling back to DB",
				zap.String("device_id", deviceID),
				zap.Error(hwErr),
			)
		}
	}

	// 仅用 DB：无行或空则返回空，不自动导入默认值；客户要点「引入默认值」再保存
	baselineAlarmItems := []alarm.AlarmItem{}
	dbExists := false
	if dbErr != nil || len(dbAlarmItems) == 0 {
		if dbErr != nil {
			isNotFound := dbErr == sql.ErrNoRows ||
				strings.Contains(dbErr.Error(), "no rows in result set") ||
				strings.Contains(dbErr.Error(), "not found")
			if !isNotFound {
				s.logger.Error("[GET_SETTINGS] database query failed",
					zap.String("device_id", deviceID),
					zap.Error(dbErr),
				)
				return nil, fmt.Errorf("failed to get alarm device: %w", dbErr)
			}
		}
		s.logger.Info("[GET_SETTINGS] DB not found or empty, returning empty",
			zap.String("device_id", deviceID),
		)
	} else {
		baselineAlarmItems = dbAlarmItems
		dbExists = true
		s.logger.Info("[GET_SETTINGS] got alarm config from DB",
			zap.String("device_id", deviceID),
			zap.Int("alarm_items_count", len(baselineAlarmItems)),
		)
	}

	// For radar: Step2 查 device，Step3 比较 DB vs device，有不同则 device 覆盖 DB 对应项、保存并返回合并结果给前端
	if deviceType == "radar" && s.qinglanClient != nil && device.DeviceUID != "" {
		deviceProps, devErr := s.getRadarDeviceConfig(ctx, device.DeviceUID)
		if devErr == nil && deviceProps != nil && len(deviceProps) > 0 {
			deviceAlarmItems, decodeErr := decode.DecodeDevicePropsToAlarmItems(baselineAlarmItems, deviceProps)
			if decodeErr == nil && len(deviceAlarmItems) > 0 {
				merged := MergeHardwareIntoBaseline(baselineAlarmItems, deviceAlarmItems)
				changedTypes := s.getChangedAlarmTypes(baselineAlarmItems, deviceAlarmItems)
				if len(changedTypes) > 0 {
					if err := s.updateAlarmDevicePartial(ctx, tenantID, deviceID, "radar", baselineAlarmItems, deviceAlarmItems); err == nil {
						s.logger.Info("[GET_SETTINGS] radar device differs from DB, synced to DB and return merged",
							zap.String("device_id", deviceID), zap.Strings("changed_types", changedTypes))
					}
				}
				return merged, nil
			}
		}
		// 未拿到 device 或解码失败：仅异步 sync，本次仍返回 DB baseline
		go s.syncDeviceValuesToDB(context.Background(), tenantID, deviceID, device.DeviceUID, deviceType, baselineAlarmItems, dbExists)
	}

	// Sleepad fallback：查询 getconfig 回填 realtime_interval；deviceLightConf/get 回填 light_mode
	if deviceType == "sleepad" && s.sleepaceGateway != nil && device.DeviceCode.Valid && device.DeviceCode.String != "" {
		if interval, err := s.sleepaceGateway.GetDeviceConfig(ctx, device.DeviceID, device.DeviceCode.String); err == nil && interval > 0 {
			for i := range baselineAlarmItems {
				if baselineAlarmItems[i].AlarmType == alarm.SleepadSetting && baselineAlarmItems[i].AlarmParams != nil {
					baselineAlarmItems[i].AlarmParams["realtime_interval"] = interval
					break
				}
			}
		}
		baselineAlarmItems = s.overlaySleepadLightModeFromHardware(ctx, device.DeviceCode.String, baselineAlarmItems)
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
	if deviceType != "sleepad" && deviceType != "radar" {
		return nil, fmt.Errorf("invalid device_type: %s (must be 'sleepad' or 'radar')", deviceType)
	}
	if alarmItems == nil || len(alarmItems) == 0 {
		return nil, fmt.Errorf("alarm_items is required")
	}

	// 根据设备类型调用对应的实现
	if deviceType == "radar" {
		return s.UpdateRadarMonitorSettings(ctx, tenantID, deviceID, userID, alarmItems, progressCallback)
	} else if deviceType == "sleepad" {
		deviceWrite, dbWrite, noChange, err := s.UpdateSleepadMonitorSettings(ctx, tenantID, deviceID, userID, alarmItems, progressCallback)
		return map[string]interface{}{
			"success":      dbWrite,
			"no_change":    noChange,
			"device_write": deviceWrite,
			"db_write":     dbWrite,
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
		alarm.MonitoringMode:      true,
		alarm.Fall:                true,
		alarm.PostureDetection:    true,
		alarm.BedSitUp:            true,
		alarm.SittingOnGround:     true,
		alarm.HeartRateAlert:      true,
		alarm.RespRateAlert:       true,
		alarm.WeakBiometricSignal: true,
	}
	// Radar 设备会返回的 AlarmType（从设备属性解码得到）
	deviceReturnAlarmTypes := map[string]bool{
		alarm.MonitoringMode:      true,
		alarm.Fall:                true,
		alarm.PostureDetection:    true,
		alarm.BedSitUp:            true,
		alarm.SittingOnGround:     true,
		alarm.HeartRateAlert:      true,
		alarm.RespRateAlert:       true,
		alarm.WeakBiometricSignal: true,
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

// UpdateSleepadMonitorSettings updates sleepad device monitor settings.
// Flow: compare -> push to hardware (via sleepad gateway) -> write DB -> publish config:alarmDevice:stream
// Same as Radar: gateway 未配置或设备写入失败则不允许保存，不入库。
func (s *deviceMonitorSettingsService) UpdateSleepadMonitorSettings(ctx context.Context, tenantID, deviceID, userID string, alarmItems []alarm.AlarmItem, progressCallback ProgressCallback) (deviceWrite bool, dbWrite bool, noChange bool, err error) {
	deviceType := "sleepad"

	device, err := s.devicesRepo.GetDevice(ctx, tenantID, deviceID)
	if err != nil {
		return false, false, false, fmt.Errorf("failed to get device: %w", err)
	}
	if !device.DeviceType.Valid {
		return false, false, false, fmt.Errorf("device has no device_type")
	}
	if !s.isDeviceTypeMatch(deviceType, device.DeviceType.String) {
		return false, false, false, fmt.Errorf("device type mismatch: expected %s, got %s", deviceType, device.DeviceType.String)
	}

	existingAlarmItems, err := s.getExistingAlarmItems(ctx, tenantID, deviceID, deviceType)
	if err != nil {
		return false, false, false, err
	}

	// Sleepad 仅向 Sleepace 设备写入其支持的报警类型（与 Radar 仅写设备支持的项一致）。
	// NightAbsence/SensorDetached/ResetTime/NapTime 等由后端或 event 计算，不写入 Sleepace 硬件；ConvertAlarmItemsToSleepaceConfig 也只含 level/param 有映射的项。
	deviceWriteAlarmTypes := map[string]bool{
		alarm.HeartRateAlert:       true,
		alarm.RespRateAlert:        true,
		alarm.ApneaHypopnea:        true,
		alarm.LeftBed:              true,
		alarm.InBed:                true,
		alarm.BedSitUp:             true,
		alarm.AbnormalBodyMovement: true,
		alarm.NoBodyMove:           true,
		alarm.NoTurnOver:           true,
		alarm.SleepadSetting:       true,
		alarm.MaterialSetting:      true,
	}
	deviceReturnAlarmTypes := map[string]bool{}

	_, noChange, hasDeviceWriteChanges, err := s.compareAndLogChanges(ctx, tenantID, deviceID, deviceType, existingAlarmItems, alarmItems, deviceWriteAlarmTypes, deviceReturnAlarmTypes)
	if err != nil {
		return false, false, false, err
	}
	if noChange {
		return true, true, true, nil
	}

	if progressCallback != nil {
		progressCallback(10, "config comparison done, pushing to hardware...")
	}

	// 1. Push to hardware via sleepad gateway（与雷达一致：未配置则不允许保存）
	if hasDeviceWriteChanges {
		if s.sleepaceGateway == nil {
			return false, false, false, fmt.Errorf("sleepad gateway client not configured")
		}
		if !device.DeviceCode.Valid || device.DeviceCode.String == "" {
			return false, false, false, fmt.Errorf("device has no device_code")
		}

		var resetTime *alarm.ResetTimeParams
		if ac, acErr := s.alarmCloudRepo.GetAlarmCloud(ctx, tenantID); acErr == nil && len(ac.Metadata) > 0 {
			var tr alarm.TenantResetTime
			if json.Unmarshal(ac.Metadata, &tr) == nil && tr.ResetTime.InBedTime != "" && tr.ResetTime.OutBedTime != "" {
				resetTime = &tr.ResetTime
			}
		}
		sleepaceConfig := ConvertAlarmItemsToSleepaceConfig(device.DeviceCode.String, device.DeviceID, alarmItems, resetTime)

		s.logger.Info("[SLEEPAD_WRITE] sending alarm config to hardware",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.String("device_code", device.DeviceCode.String),
			zap.Any("left_bed", map[string]interface{}{
				"leftBedFlag":        sleepaceConfig["leftBedFlag"],
				"leftBedDuration":    sleepaceConfig["leftBedDuration"],
				"leftBedStartHour":   sleepaceConfig["leftBedStartHour"],
				"leftBedStartMinute": sleepaceConfig["leftBedStartMinute"],
				"leftBedEndHour":     sleepaceConfig["leftBedEndHour"],
				"leftBedEndMinute":   sleepaceConfig["leftBedEndMinute"],
			}),
		)

		if progressCallback != nil {
			progressCallback(30, "pushing alarm config to sleepad hardware...")
		}

		if err := s.sleepaceGateway.UpdateAlarmConfig(ctx, sleepaceConfig); err != nil {
			s.logger.Error("[SLEEPAD_WRITE] hardware write failed, aborting DB update",
				zap.String("tenant_id", tenantID),
				zap.String("device_id", deviceID),
				zap.String("device_code", device.DeviceCode.String),
				zap.Error(err),
			)
			return false, false, false, fmt.Errorf("failed to write alarm config to hardware: %w", err)
		}
		deviceWrite = true

		s.logger.Info("[SLEEPAD_WRITE] hardware write succeeded",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
		)

		s.pushDeviceSettings(ctx, device.DeviceID, device.DeviceCode.String, alarmItems)
	} else {
		// 没有需要下发硬件的变更，也视为设备写入阶段成功
		deviceWrite = true
	}

	if progressCallback != nil {
		progressCallback(60, "hardware push succeeded, updating database...")
	}

	// 2. Write DB + publish config:alarmDevice:stream
	if err := s.updateAlarmDeviceDB(ctx, tenantID, deviceID, userID, alarmItems); err != nil {
		s.logger.Error("[SLEEPAD_WRITE] DB update failed",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.Error(err),
		)
		return deviceWrite, false, false, fmt.Errorf("failed to update database: %w", err)
	}
	dbWrite = true

	if progressCallback != nil {
		progressCallback(100, "config update completed")
	}

	s.logger.Info("[SLEEPAD_WRITE] full update succeeded (hardware + DB + stream)",
		zap.String("tenant_id", tenantID),
		zap.String("device_id", deviceID),
		zap.String("device_uid", device.DeviceUID),
	)

	return deviceWrite, dbWrite, false, nil
}

// overlaySleepadLightModeFromHardware 厂家 deviceLightConf/get 回填 light_mode（在线时以厂家为准；DB 仅在 get 失败或未查询时保留原样）。
func (s *deviceMonitorSettingsService) overlaySleepadLightModeFromHardware(ctx context.Context, deviceCode string, items []alarm.AlarmItem) []alarm.AlarmItem {
	if s.sleepaceGateway == nil || deviceCode == "" {
		return items
	}
	mode, err := s.sleepaceGateway.GetDeviceLightConf(ctx, deviceCode)
	if err != nil {
		s.logger.Debug("[GET_SETTINGS] deviceLightConf/get skipped",
			zap.String("device_code", deviceCode),
			zap.Error(err),
		)
		return items
	}
	if mode != 0 && mode != 1 {
		return items
	}
	for i := range items {
		if items[i].AlarmType == alarm.SleepadSetting {
			if items[i].AlarmParams == nil {
				items[i].AlarmParams = make(map[string]interface{})
			}
			items[i].AlarmParams["light_mode"] = mode
			break
		}
	}
	return items
}

// pushDeviceSettings pushes SleepadSetting / MaterialSetting params to hardware
// via the sleepace gateway proxy (individual API calls per setting type).
// 仅用本次请求体中的 alarm_params，不以 DB 补全下发（DB 在厂家不可达时仅作展示缓存，见 Get 侧 overlay）。
func (s *deviceMonitorSettingsService) pushDeviceSettings(ctx context.Context, deviceID, deviceCode string, items []alarm.AlarmItem) {
	if s.sleepaceGateway == nil || deviceID == "" || deviceCode == "" {
		return
	}
	for _, item := range items {
		switch item.AlarmType {
		case alarm.SleepadSetting:
			p := item.AlarmParams
			if p == nil || len(p) == 0 {
				continue
			}
			intervalV, intervalOk := toIntParam(p["realtime_interval"])
			if !intervalOk || intervalV <= 0 {
				intervalV = 2
			}
			if err := s.sleepaceGateway.SetRealtimeInterval(ctx, deviceID, deviceCode, intervalV); err != nil {
				s.logger.Warn("[SLEEPAD_WRITE] push realtime_interval", zap.Error(err))
			}
			sensV, sensOk := toIntParam(p["Bed_Exit_Sensitivity"])
			if !sensOk {
				sensV, sensOk = toIntParam(p["leave_sensibility"])
			}
			if sensOk {
				if err := s.sleepaceGateway.SetLeaveSensibility(ctx, deviceID, deviceCode, sensV); err != nil {
					s.logger.Warn("[SLEEPAD_WRITE] push Bed_Exit_Sensitivity", zap.Error(err))
				}
			}
			reportType, reportOk := toIntParam(p["report_upload_type"])
			if reportOk {
				if err := s.sleepaceGateway.SetReportUploadType(ctx, deviceID, deviceCode, reportType); err != nil {
					s.logger.Warn("[SLEEPAD_WRITE] push report_upload_type", zap.Error(err))
				}
				if reportType == 0 {
					if t, ok := toIntParam(p["report_upload_time"]); ok {
						if err := s.sleepaceGateway.SetReportUploadTime(ctx, deviceID, deviceCode, t); err != nil {
							s.logger.Warn("[SLEEPAD_WRITE] push report_upload_time", zap.Error(err))
						}
					}
				}
			}
			if lightV, lightOk := toIntParam(p["light_mode"]); lightOk {
				if err := s.sleepaceGateway.SetDeviceLightConf(ctx, deviceCode, lightV); err != nil {
					s.logger.Warn("[SLEEPAD_WRITE] deviceLightConf/set failed",
						zap.String("device_id", deviceID),
						zap.String("device_code", deviceCode),
						zap.Int("light_mode", lightV),
						zap.Error(err),
					)
				} else {
					s.logger.Info("[SLEEPAD_WRITE] deviceLightConf/set ok",
						zap.String("device_id", deviceID),
						zap.String("device_code", deviceCode),
						zap.Int("light_mode", lightV),
					)
				}
			}
		case alarm.MaterialSetting:
			p := item.AlarmParams
			if p == nil {
				continue
			}
			thickness, _ := toIntParam(p["thickness"])
			material, _ := toIntParam(p["material_type"])
			if thickness > 0 || material > 0 {
				if err := s.sleepaceGateway.SetBedParameters(ctx, deviceID, deviceCode, thickness, material); err != nil {
					s.logger.Warn("[SLEEPAD_WRITE] push bed parameters", zap.Error(err))
				}
			}
		}
	}
}

func toIntParam(v interface{}) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	}
	return 0, false
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
			s.logger.Info("Alarm device not found, returning empty",
				zap.String("tenant_id", tenantID),
				zap.String("device_id", deviceID),
				zap.String("device_type", deviceType),
			)
			return []alarm.AlarmItem{}, nil
		}
		// 其他错误直接返回
		return nil, fmt.Errorf("failed to get alarm device: %w", err)
	}

	existingAlarmItems, err := s.parseMonitorConfigToAlarmItems(existingAlarmDevice.MonitorConfig, deviceType)
	if err != nil {
		return nil, fmt.Errorf("failed to parse existing monitor config: %w", err)
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
	//   - heart_breath_param (从 HeartRateAlert, RespRateAlert, WeakBiometricSignal)
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

// updateAlarmDeviceDB 更新 alarm_device 表和 config_version。保存全量 AlarmItem[]，避免后续默认值变更覆盖客户已保存配置。
func (s *deviceMonitorSettingsService) updateAlarmDeviceDB(ctx context.Context, tenantID, deviceID, userID string, alarmItems []alarm.AlarmItem) error {
	monitorConfigJSON, err := s.marshalAlarmItemsToFullConfig(alarmItems)
	if err != nil {
		return fmt.Errorf("failed to marshal alarm items to full config: %w", err)
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
	switch strings.ToLower(deviceType) {
	case "sleepad":
		return "sleepad"
	case "radar":
		return "radar"
	default:
		return ""
	}
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
	if deviceType != "sleepad" && deviceType != "radar" {
		return nil, fmt.Errorf("invalid device_type: %s (must be 'sleepad' or 'radar')", deviceType)
	}

	// 从 alarm_cloud 或 DefaultAlarmSetting 获取默认配置
	// buildDeviceAlarmItemsFromCloudOrDefault 会先尝试从 alarm_cloud 获取，如果没有则从 alarm.go 中的默认值获取
	alarmItems, err := s.buildDeviceAlarmItemsFromCloudOrDefault(ctx, tenantID, deviceType)
	if err != nil {
		return nil, fmt.Errorf("failed to build alarm items: %w", err)
	}

	return alarmItems, nil
}

// getRadarDeviceConfig 查询雷达当前报警配置（工作模式、跌倒参数、呼吸心率参数），超时 5 秒。仅做 Step2 查询，由调用方决定是否使用。
func (s *deviceMonitorSettingsService) getRadarDeviceConfig(ctx context.Context, deviceUID string) (map[string]interface{}, error) {
	if s.qinglanClient == nil || deviceUID == "" {
		return nil, fmt.Errorf("qinglan client or device_uid is empty")
	}
	deviceReadCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.qinglanClient.GetDeviceProperties(deviceReadCtx, deviceUID, []string{
		"radar_func_ctrl", "fall_param", "heart_breath_param",
	})
}

// syncDeviceValuesToDB 异步查询设备并更新数据库（只更新alarm_params或enabled，updated_by=null，不更新config_version）
func (s *deviceMonitorSettingsService) syncDeviceValuesToDB(ctx context.Context, tenantID, deviceID, deviceUID, deviceType string, baselineAlarmItems []alarm.AlarmItem, dbExists bool) {
	s.logger.Info("[SYNC_DEVICE] Starting async device merge and DB sync",
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

	// 部分更新数据库：device 覆盖 DB 对应字段，保存全量
	if err := s.updateAlarmDevicePartial(ctx, tenantID, deviceID, "radar", baselineAlarmItems, deviceAlarmItems); err != nil {
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

// syncSleepadHardwareToDB writes merged (DB alarm levels + HW switch/params) to alarm_device when Sleepace returned different params/enabled. Does not bump config_version.
func (s *deviceMonitorSettingsService) syncSleepadHardwareToDB(ctx context.Context, tenantID, deviceID string, baseline, merged []alarm.AlarmItem) {
	if err := s.updateAlarmDevicePartial(ctx, tenantID, deviceID, "sleepad", baseline, merged); err != nil {
		s.logger.Warn("[GET_SETTINGS] sleepad hardware sync to DB failed",
			zap.String("device_id", deviceID),
			zap.Error(err),
		)
		return
	}
	s.logger.Info("[GET_SETTINGS] sleepad hardware values synced to DB",
		zap.String("device_id", deviceID),
	)
}

// updateAlarmDevicePartial 用 device 返回值覆盖 DB 中对应字段，保存全量。device 为真实值。
func (s *deviceMonitorSettingsService) updateAlarmDevicePartial(ctx context.Context, tenantID, deviceID, deviceType string, baselineAlarmItems, deviceAlarmItems []alarm.AlarmItem) error {
	if s.db == nil {
		return fmt.Errorf("database connection is not available")
	}

	deviceMap := make(map[string]alarm.AlarmItem)
	for _, item := range deviceAlarmItems {
		if item.AlarmType != "" {
			deviceMap[item.AlarmType] = item
		}
	}

	// 读取现有 monitor_config
	alarmDevice, err := s.alarmDeviceRepo.GetAlarmDevice(ctx, tenantID, deviceID)
	if err != nil {
		// 无 DB 记录：不写入，避免客户误以为已配置。只有客户在页面点保存才会建行。
		return nil
	}

	// 有记录：解析为 []AlarmItem，用 device 覆盖对应项后存全量
	existingItems, parseErr := s.parseMonitorConfigToAlarmItems(alarmDevice.MonitorConfig, deviceType)
	if parseErr != nil || len(existingItems) == 0 {
		existingItems = baselineAlarmItems
	}
	// device 覆盖 DB 中对应 alarm_type 的 AlarmParams、IsEnabled（设备为真实值）
	for i := range existingItems {
		typ := existingItems[i].AlarmType
		if dev, ok := deviceMap[typ]; ok {
			if len(dev.AlarmParams) > 0 {
				existingItems[i].AlarmParams = dev.AlarmParams
			}
			if dev.IsEnabled != nil {
				existingItems[i].IsEnabled = dev.IsEnabled
			}
		}
	}

	updatedMonitorConfigJSON, marshalErr := s.marshalAlarmItemsToFullConfig(existingItems)
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal updated monitor config: %w", marshalErr)
	}
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
