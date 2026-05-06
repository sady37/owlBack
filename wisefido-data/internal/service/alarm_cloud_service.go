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
	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

// DeviceAlarmEntry 是 alarm_cloud.device_alarms JSONB 中"每条 alarm 的存储格式"。
//
// 双字段独立存储，is_enabled 与 alarm_level 解耦：
//
//	{"SleepPad": {"LeftBed": {"is_enabled": 0, "alarm_level": "WARNING"}}}
//
// 老格式（合并 string + sentinel "DISABLED"）已通过一次性 SQL migration 升级为本结构；
// 本类型不再做 UnmarshalJSON 兼容——遇到旧字符串数据会反序列化失败，作为部署侧的
// 硬规则强制 schema 干净。迁移脚本见 doc/migrations/alarm_cloud_device_alarms_split.sql。
type DeviceAlarmEntry struct {
	IsEnabled  int    `json:"is_enabled"`
	AlarmLevel string `json:"alarm_level"`
}

// getResourcePermission 查询资源权限配置（从 role_permissions 表）
// 为了避免循环导入，这里复制了 httpapi.GetResourcePermission 的逻辑
//
// 注意: permission_scope 值映射:
//   - 'A' = All (no restriction) → assigned_only=false, branch_only=false
//   - 'S' = assigned_only → assigned_only=true, branch_only=false
//   - 'B' = branch_only → assigned_only=false, branch_only=true
func getResourcePermission(db *sql.DB, ctx context.Context, roleCode, resourceType, permissionType string) (*resourcePermissionCheck, error) {
	var permissionScope string
	err := db.QueryRowContext(ctx,
		`SELECT permission_scope
		 FROM role_permissions
		 WHERE tenant_id = $1 
		   AND role_code = $2 
		   AND resource_type = $3 
		   AND permission_type = $4
		 LIMIT 1`,
		"00000000-0000-0000-0000-000000000001", // SystemTenantID
		roleCode, resourceType, permissionType,
	).Scan(&permissionScope)

	if err == sql.ErrNoRows {
		// 记录不存在：返回最严格的权限（安全默认值）
		return &resourcePermissionCheck{AssignedOnly: true, BranchOnly: true}, nil
	}
	if err != nil {
		return nil, err
	}

	// 将 permission_scope 转换为 assigned_only 和 branch_only 标志
	var assignedOnly, branchOnly bool
	switch permissionScope {
	case "A":
		// All (no restriction)
		assignedOnly = false
		branchOnly = false
	case "S":
		// assigned_only
		assignedOnly = true
		branchOnly = false
	case "B":
		// branch_only
		assignedOnly = false
		branchOnly = true
	default:
		// 未知值，返回最严格的权限（安全默认值）
		assignedOnly = true
		branchOnly = true
	}

	return &resourcePermissionCheck{AssignedOnly: assignedOnly, BranchOnly: branchOnly}, nil
}

// resourcePermissionCheck 资源权限检查结果（用于 service 包内的权限检查）
type resourcePermissionCheck struct {
	AssignedOnly bool
	BranchOnly   bool
}

// AlarmCloudService 告警配置服务接口
type AlarmCloudService interface {
	GetAlarmCloudConfig(ctx context.Context, req GetAlarmCloudConfigRequest) (*alarm.AlarmCloudConfig, error)
	UpdateAlarmCloudConfig(ctx context.Context, req UpdateAlarmCloudConfigRequest) (*alarm.AlarmCloudConfig, error)
}

// alarmCloudService 实现
type alarmCloudService struct {
	alarmCloudRepo     repository.AlarmCloudRepository
	configVersionsRepo repository.ConfigVersionsRepository // 配置版本仓库（用于审计）
	db                 *sql.DB                             // 用于权限检查
	logger             *zap.Logger
}

// NewAlarmCloudService 创建 AlarmCloudService 实例
func NewAlarmCloudService(
	alarmCloudRepo repository.AlarmCloudRepository,
	configVersionsRepo repository.ConfigVersionsRepository,
	db *sql.DB,
	logger *zap.Logger,
) AlarmCloudService {
	return &alarmCloudService{
		alarmCloudRepo:     alarmCloudRepo,
		configVersionsRepo: configVersionsRepo,
		db:                 db,
		logger:             logger,
	}
}

// GetAlarmCloudConfigRequest 查询告警配置请求
type GetAlarmCloudConfigRequest struct {
	TenantID string
	UserID   string // 当前用户ID（用于权限检查）
	UserRole string // 当前用户角色（用于权限检查）
}

// UpdateAlarmCloudConfigRequest 更新告警配置请求
type UpdateAlarmCloudConfigRequest struct {
	TenantID string
	UserID   string                  // 当前用户ID（用于权限检查）
	UserRole string                  // 当前用户角色（用于权限检查）
	Config   *alarm.AlarmCloudConfig // 完整的配置对象
}

// GetAlarmCloudConfig 查询告警配置
// 返回完整的 AlarmCloudConfig 对象
func (s *alarmCloudService) GetAlarmCloudConfig(ctx context.Context, req GetAlarmCloudConfigRequest) (*alarm.AlarmCloudConfig, error) {
	// 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	// 权限检查：只有 SystemAdmin 或 Admin 可以查看告警配置
	if req.UserRole != "" && s.db != nil {
		normalizedRole := strings.ToLower(strings.TrimSpace(req.UserRole))
		// SystemAdmin 和 Admin 可以查看
		if normalizedRole != "systemadmin" && normalizedRole != "admin" {
			// 检查是否有 read 权限
			permCheck, err := getResourcePermission(s.db, ctx, req.UserRole, "alarm_cloud", "R")
			if err != nil {
				s.logger.Warn("Failed to check permission for GetAlarmCloudConfig",
					zap.String("user_role", req.UserRole),
					zap.Error(err),
				)
				// 权限检查失败时，使用默认严格权限（不允许访问）
				return nil, fmt.Errorf("permission denied: failed to check permissions")
			}
			// 如果权限检查返回最严格权限（assigned_only=true, branch_only=true），表示没有权限记录
			// 这种情况下，只有 SystemAdmin 和 Admin 可以访问
			if permCheck.AssignedOnly && permCheck.BranchOnly {
				return nil, fmt.Errorf("permission denied: only SystemAdmin or Admin can view alarm cloud config")
			}
		}
	}

	// 1. 查询租户特定配置
	alarmCloud, err := s.alarmCloudRepo.GetAlarmCloud(ctx, req.TenantID)
	if err != nil {
		// 检查是否是 "not found" 错误
		isNotFound := err == sql.ErrNoRows ||
			(fmt.Sprintf("%v", err) == "alarm cloud not found: sql: no rows in result set" ||
				strings.Contains(fmt.Sprintf("%v", err), "alarm cloud not found"))

		if isNotFound {
			// 如果租户没有配置，从默认值构建配置并自动保存
			defaultConfig, buildErr := buildDefaultAlarmCloudConfig(req.TenantID)
			if buildErr != nil {
				s.logger.Warn("Failed to build default alarm cloud config",
					zap.String("tenant_id", req.TenantID),
					zap.Error(buildErr),
				)
				// 如果构建失败，返回默认配置
				return buildDefaultAlarmCloudConfigObject(), nil
			}

			// 使用事务方式保存默认配置，以便设置 created_by 和 updated_by
			if s.db != nil && req.UserID != "" {
				tx, txErr := s.db.BeginTx(ctx, nil)
				if txErr == nil {
					defer tx.Rollback()

					// 构建 SQL 语句
					query := `
						INSERT INTO alarm_cloud (
							tenant_id,
							offlinealarm,
							lowbattery,
							devicefailure,
							device_alarms,
							conditions,
							notification_rules,
							metadata,
							created_at,
							created_by,
							updated_at,
							updated_by
						) VALUES ($1, $2, $3, $4, $5::jsonb, $6::jsonb, $7::jsonb, $8::jsonb, $9, $10, $11, $12)
					`

					var offlineAlarm, lowBattery, deviceFailure interface{}
					if defaultConfig.OfflineAlarm != "" {
						offlineAlarm = defaultConfig.OfflineAlarm
					}
					if defaultConfig.LowBattery != "" {
						lowBattery = defaultConfig.LowBattery
					}
					if defaultConfig.DeviceFailure != "" {
						deviceFailure = defaultConfig.DeviceFailure
					}

					var deviceAlarms, conditions, notificationRules, metadata interface{}
					if len(defaultConfig.DeviceAlarms) > 0 {
						deviceAlarms = string(defaultConfig.DeviceAlarms)
					} else {
						deviceAlarms = "{}"
					}
					if len(defaultConfig.Conditions) > 0 {
						conditions = string(defaultConfig.Conditions)
					} else {
						conditions = "{}"
					}
					if len(defaultConfig.NotificationRules) > 0 {
						notificationRules = string(defaultConfig.NotificationRules)
					}
					if len(defaultConfig.Metadata) > 0 {
						metadata = string(defaultConfig.Metadata)
					}

					now := time.Now().UTC()
					_, execErr := tx.ExecContext(ctx, query, req.TenantID, offlineAlarm, lowBattery, deviceFailure,
						deviceAlarms, conditions, notificationRules, metadata, now, req.UserID, now, req.UserID)
					if execErr == nil {
						if commitErr := tx.Commit(); commitErr == nil {
							// 保存成功，重新读取记录
							alarmCloud, readErr := s.alarmCloudRepo.GetAlarmCloud(ctx, req.TenantID)
							if readErr == nil {
								// 使用读取的记录构建响应
								return buildAlarmCloudConfigFromDomain(alarmCloud)
							}
						}
					}
					// 如果事务保存失败，继续使用 UpsertAlarmCloud 作为后备方案
				}
			}

			// 后备方案：使用 UpsertAlarmCloud（无法设置 created_by）
			if saveErr := s.alarmCloudRepo.UpsertAlarmCloud(ctx, req.TenantID, defaultConfig); saveErr != nil {
				s.logger.Warn("Failed to save default alarm cloud config",
					zap.String("tenant_id", req.TenantID),
					zap.Error(saveErr),
				)
				// 保存失败不影响返回，继续返回默认配置
			}

			// 使用构建的默认配置构建响应
			alarmCloud = defaultConfig
		} else {
			// 其他错误，返回错误
			return nil, fmt.Errorf("failed to get alarm cloud: %w", err)
		}
	}

	// 2. 从 domain.AlarmCloud 构建 alarm.AlarmCloudConfig
	// buildAlarmCloudConfigFromDomain 会自动处理空字段，使用默认值填充
	config, err := buildAlarmCloudConfigFromDomain(alarmCloud)
	if err != nil {
		return nil, fmt.Errorf("failed to build config from domain: %w", err)
	}

	return config, nil
}

// buildDefaultAlarmCloudConfigObject 构建默认的 AlarmCloudConfig 对象
func buildDefaultAlarmCloudConfigObject() *alarm.AlarmCloudConfig {
	return &alarm.AlarmCloudConfig{
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

// buildAlarmCloudConfigFromDomain 从 domain.AlarmCloud 构建 alarm.AlarmCloudConfig
func buildAlarmCloudConfigFromDomain(alarmCloud *domain.AlarmCloud) (*alarm.AlarmCloudConfig, error) {
	config := buildDefaultAlarmCloudConfigObject()

	// 1. 设置 CommonAlarm
	if alarmCloud.OfflineAlarm != "" {
		config.CommonAlarm.OfflineAlarm = alarmCloud.OfflineAlarm
	}
	if alarmCloud.LowBattery != "" {
		config.CommonAlarm.LowBattery = alarmCloud.LowBattery
	}
	if alarmCloud.DeviceFailure != "" {
		config.CommonAlarm.DeviceFailure = alarmCloud.DeviceFailure
	}

	// 2. 解析 TenantResetTime（从 metadata）
	if len(alarmCloud.Metadata) > 0 {
		var tenantResetTime alarm.TenantResetTime
		if err := json.Unmarshal(alarmCloud.Metadata, &tenantResetTime); err == nil {
			config.TenantResetTime = tenantResetTime
		}
	}

	// 3. 解析 AlarmSetting（从 device_alarms 和 DefaultAlarmSetting 构建完整的 AlarmSetting）
	config.AlarmSetting.Sleepad = make([]alarm.AlarmItem, 0)
	config.AlarmSetting.Radar = make([]alarm.AlarmItem, 0)

	// 从 device_alarms 构建映射表（DeviceAlarmEntry 自带兼容老 string 格式的 UnmarshalJSON）
	var deviceAlarmsMap map[string]map[string]DeviceAlarmEntry
	if len(alarmCloud.DeviceAlarms) > 0 {
		if err := json.Unmarshal(alarmCloud.DeviceAlarms, &deviceAlarmsMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal device_alarms: %w", err)
		}
	}

	// 处理 SleepPad：用 DefaultAlarmSetting 作基础，alarm_cloud 双字段直接覆盖
	for _, item := range alarm.DefaultAlarmSetting.Sleepad {
		if item.DisplaySetting == alarm.DisplayNone {
			continue
		}
		newItem := item
		if sleepPadMap, ok := deviceAlarmsMap["SleepPad"]; ok {
			if entry, exists := sleepPadMap[item.AlarmType]; exists {
				enabled := entry.IsEnabled
				newItem.IsEnabled = &enabled
				if entry.AlarmLevel != "" {
					level := entry.AlarmLevel
					newItem.AlarmLevel = &level
				}
				// AlarmLevel 空（来自老格式 disabled 项）→ 保留 default level 不动
			}
		}
		config.AlarmSetting.Sleepad = append(config.AlarmSetting.Sleepad, newItem)
	}

	// 处理 Radar（同 SleepPad 双字段语义）
	for _, item := range alarm.DefaultAlarmSetting.Radar {
		if item.DisplaySetting == alarm.DisplayNone {
			continue
		}
		newItem := item
		if radarMap, ok := deviceAlarmsMap["Radar"]; ok {
			if entry, exists := radarMap[item.AlarmType]; exists {
				enabled := entry.IsEnabled
				newItem.IsEnabled = &enabled
				if entry.AlarmLevel != "" {
					level := entry.AlarmLevel
					newItem.AlarmLevel = &level
				}
			}
		}
		config.AlarmSetting.Radar = append(config.AlarmSetting.Radar, newItem)
	}

	// 4. 解析 CloudVitalAlarmThreshold（从 conditions）
	if len(alarmCloud.Conditions) > 0 {
		var threshold alarm.CloudVitalAlarmThreshold
		if err := json.Unmarshal(alarmCloud.Conditions, &threshold); err == nil {
			// 检查解析后的 threshold 是否为空（Conditions 为 nil 或空对象）
			// 如果为空，使用默认值
			if threshold.Conditions == nil {
				// conditions 为空，使用默认值
				config.CloudVitalAlarmThreshold = alarm.DefaultCloudVitalAlarmThreshold
			} else {
				config.CloudVitalAlarmThreshold = threshold
			}
		} else {
			// 解析失败，使用默认值
			config.CloudVitalAlarmThreshold = alarm.DefaultCloudVitalAlarmThreshold
		}
	} else {
		// conditions 字段为空，使用默认值
		config.CloudVitalAlarmThreshold = alarm.DefaultCloudVitalAlarmThreshold
	}

	return config, nil
}

// buildDomainAlarmCloudFromConfig 从 alarm.AlarmCloudConfig 构建 domain.AlarmCloud
func buildDomainAlarmCloudFromConfig(tenantID string, config *alarm.AlarmCloudConfig) (*domain.AlarmCloud, error) {
	// 1. 构建 device_alarms JSONB
	// 新格式：{ "Radar": { "Fall": {"is_enabled":1,"alarm_level":"CRITICAL"} }, "SleepPad": ... }

	// 写 device_alarms JSONB：每条 alarm 用 DeviceAlarmEntry 对象表达 is_enabled + alarm_level 双字段，
	// 不再用合并 string + sentinel "DISABLED"。读侧 buildAlarmCloudConfigFromDomain 直接读双字段，
	// disabled 项的 alarm_level 也保留下来（用户切回 enabled 不丢失上次等级）。
	deviceAlarmsTyped := map[string]map[string]DeviceAlarmEntry{}

	sleepPadAlarms := map[string]DeviceAlarmEntry{}
	for _, item := range config.AlarmSetting.Sleepad {
		if item.IsEnabled == nil {
			continue
		}
		entry := DeviceAlarmEntry{IsEnabled: *item.IsEnabled}
		if item.AlarmLevel != nil {
			entry.AlarmLevel = *item.AlarmLevel
		}
		sleepPadAlarms[item.AlarmType] = entry
	}
	if len(sleepPadAlarms) > 0 {
		deviceAlarmsTyped["SleepPad"] = sleepPadAlarms
	}

	radarAlarms := map[string]DeviceAlarmEntry{}
	for _, item := range config.AlarmSetting.Radar {
		if item.IsEnabled == nil {
			continue
		}
		entry := DeviceAlarmEntry{IsEnabled: *item.IsEnabled}
		if item.AlarmLevel != nil {
			entry.AlarmLevel = *item.AlarmLevel
		}
		radarAlarms[item.AlarmType] = entry
	}
	if len(radarAlarms) > 0 {
		deviceAlarmsTyped["Radar"] = radarAlarms
	}
	deviceAlarms := deviceAlarmsTyped

	deviceAlarmsJSON, err := json.Marshal(deviceAlarms)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal device_alarms: %w", err)
	}

	// 2. 构建 conditions JSONB（Cloud Vital Alarm Threshold）
	conditionsJSON, err := json.Marshal(config.CloudVitalAlarmThreshold)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal conditions: %w", err)
	}

	// 3. 构建 metadata JSONB（Tenant ResetTime/NapTime）
	metadataJSON, err := json.Marshal(config.TenantResetTime)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// 4. 构建完整的 AlarmCloud 配置
	alarmCloud := &domain.AlarmCloud{
		TenantID:      tenantID,
		OfflineAlarm:  config.CommonAlarm.OfflineAlarm,
		LowBattery:    config.CommonAlarm.LowBattery,
		DeviceFailure: config.CommonAlarm.DeviceFailure,
		DeviceAlarms:  json.RawMessage(deviceAlarmsJSON),
		Conditions:    json.RawMessage(conditionsJSON),
		Metadata:      json.RawMessage(metadataJSON),
	}

	return alarmCloud, nil
}

// buildDefaultAlarmCloudConfig 从默认值构建 alarm_cloud 配置
// 从 ExampleDefaultAlarmSettingJSON 或 DefaultAlarmSetting 提取默认配置
// 返回完整的 domain.AlarmCloud 配置，包括：
// - 通用报警（OfflineAlarm, LowBattery, DeviceFailure）
// - 设备特定报警（device_alarms JSONB）
// - 条件/阈值（conditions JSONB - Cloud Vital Alarm Threshold）
// - 租户作息时间（metadata JSONB - reset_time/nap_time）
func buildDefaultAlarmCloudConfig(tenantID string) (*domain.AlarmCloud, error) {
	// 使用 DefaultAlarmSetting Go 结构体（而不是 JSON 字符串）
	defaultSetting := alarm.DefaultAlarmSetting

	// 1. 构建 device_alarms JSONB
	// 格式：{ "Radar": { "Radar_Fall": "CRITICAL", ... }, "SleepPad": { "SleepPad_LeftBed": "WARNING", ... } }
	deviceAlarms := make(map[string]map[string]string)

	// 处理 SleepPad 报警
	sleepPadAlarms := make(map[string]string)
	for _, item := range defaultSetting.Sleepad {
		// 只包含启用的、有报警级别的、且 DisplaySetting 包含 alarm_cloud 的项
		if item.IsEnabled != nil && *item.IsEnabled == alarm.IsEnabledOn && item.AlarmLevel != nil {
			// 过滤掉 DisplayNone 的项，只保留 DisplayAlarmCloud 或 DisplayAlarmCloudAndDevice
			if item.DisplaySetting == alarm.DisplayAlarmCloud || item.DisplaySetting == alarm.DisplayAlarmCloudAndDevice {
				sleepPadAlarms[item.AlarmType] = *item.AlarmLevel
			}
		}
	}
	if len(sleepPadAlarms) > 0 {
		deviceAlarms["SleepPad"] = sleepPadAlarms
	}

	// 处理 Radar 报警
	radarAlarms := make(map[string]string)
	for _, item := range defaultSetting.Radar {
		// 只包含启用的、有报警级别的、且 DisplaySetting 包含 alarm_cloud 的项
		if item.IsEnabled != nil && *item.IsEnabled == alarm.IsEnabledOn && item.AlarmLevel != nil {
			// 过滤掉 DisplayNone 的项，只保留 DisplayAlarmCloud 或 DisplayAlarmCloudAndDevice
			if item.DisplaySetting == alarm.DisplayAlarmCloud || item.DisplaySetting == alarm.DisplayAlarmCloudAndDevice {
				radarAlarms[item.AlarmType] = *item.AlarmLevel
			}
		}
	}
	if len(radarAlarms) > 0 {
		deviceAlarms["Radar"] = radarAlarms
	}

	deviceAlarmsJSON, err := json.Marshal(deviceAlarms)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal device_alarms: %w", err)
	}

	// 2. 构建 conditions JSONB（Cloud Vital Alarm Threshold）
	conditionsJSON, err := json.Marshal(alarm.DefaultCloudVitalAlarmThreshold)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal conditions: %w", err)
	}

	// 3. 构建 metadata JSONB（Tenant ResetTime/NapTime）
	metadataJSON, err := json.Marshal(alarm.DefaultTenantResetTime)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// 4. 构建完整的 AlarmCloud 配置
	alarmCloud := &domain.AlarmCloud{
		TenantID:      tenantID,
		OfflineAlarm:  alarm.DefaultOfflineAlarm,
		LowBattery:    alarm.DefaultLowBattery,
		DeviceFailure: alarm.DefaultDeviceFailure,
		DeviceAlarms:  json.RawMessage(deviceAlarmsJSON),
		Conditions:    json.RawMessage(conditionsJSON),
		Metadata:      json.RawMessage(metadataJSON),
	}

	return alarmCloud, nil
}

// UpdateAlarmCloudConfig 更新告警配置
// 接收完整的 AlarmCloudConfig 对象，比对后更新，旧记录存入 config_versions
func (s *alarmCloudService) UpdateAlarmCloudConfig(ctx context.Context, req UpdateAlarmCloudConfigRequest) (*alarm.AlarmCloudConfig, error) {
	// 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.Config == nil {
		return nil, fmt.Errorf("config is required")
	}

	// 权限检查：只有 SystemAdmin 或 Admin 可以更新告警配置
	if req.UserRole == "" {
		return nil, fmt.Errorf("user_role is required for update operation")
	}
	if s.db != nil {
		normalizedRole := strings.ToLower(strings.TrimSpace(req.UserRole))
		// SystemAdmin 和 Admin 可以更新
		if normalizedRole != "systemadmin" && normalizedRole != "admin" {
			// 检查是否有 update 权限
			permCheck, err := getResourcePermission(s.db, ctx, req.UserRole, "alarm_cloud", "U")
			if err != nil {
				s.logger.Warn("Failed to check permission for UpdateAlarmCloudConfig",
					zap.String("user_role", req.UserRole),
					zap.Error(err),
				)
				return nil, fmt.Errorf("permission denied: failed to check permissions")
			}
			// 如果权限检查返回最严格权限（assigned_only=true, branch_only=true），表示没有权限记录
			// 这种情况下，只有 SystemAdmin 和 Admin 可以更新
			if permCheck.AssignedOnly && permCheck.BranchOnly {
				return nil, fmt.Errorf("permission denied: only SystemAdmin or Admin can update alarm cloud config")
			}
		}
	}

	// 业务规则验证：不能更新系统默认配置
	if req.TenantID == "00000000-0000-0000-0000-000000000001" {
		return nil, fmt.Errorf("cannot update system alarm cloud config")
	}

	// 1. 获取现有配置（如果存在），用于比对和保存到 config_versions
	var existingConfig *alarm.AlarmCloudConfig
	existingAlarmCloud, err := s.alarmCloudRepo.GetAlarmCloud(ctx, req.TenantID)
	if err != nil {
		// 检查是否是 "not found" 错误
		isNotFound := err == sql.ErrNoRows ||
			(fmt.Sprintf("%v", err) == "alarm cloud not found: sql: no rows in result set" ||
				strings.Contains(fmt.Sprintf("%v", err), "alarm cloud not found"))
		if !isNotFound {
			return nil, fmt.Errorf("failed to get existing alarm cloud: %w", err)
		}
		// 如果是 not found，existingConfig 为 nil，使用默认配置
		existingConfig = buildDefaultAlarmCloudConfigObject()
	} else {
		// 从 domain.AlarmCloud 构建 alarm.AlarmCloudConfig
		existingConfig, err = buildAlarmCloudConfigFromDomain(existingAlarmCloud)
		if err != nil {
			return nil, fmt.Errorf("failed to build existing config: %w", err)
		}
	}

	// 2. 比对配置是否有变化（简单比对 JSON 字符串）
	existingJSON, err := json.Marshal(existingConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal existing config: %w", err)
	}
	newJSON, err := json.Marshal(req.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal new config: %w", err)
	}

	hasChanges := string(existingJSON) != string(newJSON)

	// 3. 从 alarm.AlarmCloudConfig 构建 domain.AlarmCloud
	alarmCloud, err := buildDomainAlarmCloudFromConfig(req.TenantID, req.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to build domain alarm cloud: %w", err)
	}

	// 4. 使用事务保证原子性：要么全部成功，要么全部失败
	// 开始事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // 确保在函数返回时回滚（如果未提交）

	// 4.1. 如果有变化，先保存旧配置到 config_versions（在事务中）
	if hasChanges && s.configVersionsRepo != nil {
		configData := map[string]interface{}{
			"metadata": existingConfig,
		}
		configDataJSON, err := json.Marshal(configData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal config_data for config_version: %w", err)
		}

		configVersion := &domain.ConfigVersion{
			ConfigType:      "alarm_cloud",
			EntityID:        req.TenantID, // alarm_cloud 的 entity_id 就是 tenant_id
			CurrentEntityID: req.TenantID,
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

	// 4.2. 更新 alarm_cloud（在事务中）
	query := `
		INSERT INTO alarm_cloud (
			tenant_id,
			offlinealarm,
			lowbattery,
			devicefailure,
			device_alarms,
			conditions,
			notification_rules,
			metadata,
			updated_at,
			updated_by
		) VALUES ($1, $2, $3, $4, $5::jsonb, $6::jsonb, $7::jsonb, $8::jsonb, $9, $10)
		ON CONFLICT (tenant_id) DO UPDATE SET
			offlinealarm = EXCLUDED.offlinealarm,
			lowbattery = EXCLUDED.lowbattery,
			devicefailure = EXCLUDED.devicefailure,
			device_alarms = EXCLUDED.device_alarms,
			conditions = EXCLUDED.conditions,
			notification_rules = EXCLUDED.notification_rules,
			metadata = EXCLUDED.metadata,
			updated_at = EXCLUDED.updated_at,
			updated_by = EXCLUDED.updated_by
	`

	var offlineAlarm, lowBattery, deviceFailure interface{}
	if alarmCloud.OfflineAlarm != "" {
		offlineAlarm = alarmCloud.OfflineAlarm
	}
	if alarmCloud.LowBattery != "" {
		lowBattery = alarmCloud.LowBattery
	}
	if alarmCloud.DeviceFailure != "" {
		deviceFailure = alarmCloud.DeviceFailure
	}

	var deviceAlarms, conditions, notificationRules, metadata interface{}
	if len(alarmCloud.DeviceAlarms) > 0 {
		deviceAlarms = string(alarmCloud.DeviceAlarms)
	} else {
		deviceAlarms = "{}"
	}
	if len(alarmCloud.Conditions) > 0 {
		conditions = string(alarmCloud.Conditions)
	} else {
		conditions = "{}"
	}
	if len(alarmCloud.NotificationRules) > 0 {
		notificationRules = string(alarmCloud.NotificationRules)
	}
	if len(alarmCloud.Metadata) > 0 {
		metadata = string(alarmCloud.Metadata)
	}

	// 设置 updated_at 和 updated_by
	updatedAt := time.Now().UTC()
	var updatedBy interface{}
	if req.UserID != "" {
		updatedBy = req.UserID
	}

	_, err = tx.ExecContext(ctx, query, req.TenantID, offlineAlarm, lowBattery, deviceFailure,
		deviceAlarms, conditions, notificationRules, metadata, updatedAt, updatedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to update alarm cloud: %w", err)
	}

	// 4.3. 提交事务（如果所有操作都成功）
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 5. 返回更新后的配置
	// 注意：alarm_cloud 配置变更不再发布事件（已废弃，不影响已初始化的设备）
	return s.GetAlarmCloudConfig(ctx, GetAlarmCloudConfigRequest{
		TenantID: req.TenantID,
		UserID:   req.UserID,
		UserRole: req.UserRole,
	})
}
