package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

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

// getResourcePermission 查询 v2 RBAC 中 (role, resource, action) 是否被允许。
//
// 与 httpapi.GetResourcePermission 行为一致；这里独立一份避免 service→http 循环依赖。
// v2 role_permissions 列 (role_id FK→roles, permission, resource_scope INET)：
// 用 EXISTS 查 role 是否拥有 ({resource}.{action} / {resource}.* / tenant.* / *) 任一匹配。
// scope (assigned_only/branch_only) 在 v2 由 IPv6 prefix 自带层级表达，业务侧用 utils/spatial 派生；
// 这里命中返回 (false,false) 放行，未命中返回 strictest 兜底。
func getResourcePermission(db *sql.DB, ctx context.Context, roleCode, resourceType, permissionType string) (*resourcePermissionCheck, error) {
	v2Role := mapRoleToV2(roleCode)
	action := permWord(permissionType)

	// 候选 permission 字符串：精确 → 资源 .config（CRUD 合一）→ 资源通配 → 租户通配 → 全局通配
	target := resourceType + "." + action
	resourceConfig := resourceType + ".config"
	resourceAll := resourceType + ".*"
	const tenantAll = "tenant.*"
	const platformAll = "*"

	var exists bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			  FROM role_permissions rp
			  JOIN roles r ON r.role_id = rp.role_id
			 WHERE r.role_code = $1
			   AND rp.permission IN ($2, $3, $4, $5, $6)
		)
	`, v2Role, target, resourceConfig, resourceAll, tenantAll, platformAll).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return &resourcePermissionCheck{AssignedOnly: true, BranchOnly: true}, nil
	}
	return &resourcePermissionCheck{AssignedOnly: false, BranchOnly: false}, nil
}

// mapRoleToV2 把 v1 PascalCase role code 映射回 v2 snake_case role_code。
// 与 httpapi.mapRoleToV2 一致；这里独立一份避免循环依赖。
func mapRoleToV2(role string) string {
	switch role {
	case "SystemAdmin", "SystemOperator":
		return "platform_admin"
	case "Admin":
		return "tenant_admin"
	case "Manager":
		return "manager"
	case "Nurse":
		return "nurse"
	case "Caregiver":
		return "caregiver"
	case "Family":
		return "family"
	case "Viewer":
		return "viewer"
	}
	return strings.ToLower(role)
}

// permWord 把 v1 单字母权限码转 v2 动词。
func permWord(p string) string {
	switch strings.ToUpper(p) {
	case "R":
		return "read"
	case "C":
		return "create"
	case "U":
		return "update"
	case "D":
		return "delete"
	}
	return strings.ToLower(p)
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

// alarmCloudService 实现 — tenant 级 alarm.cloud_config 是给 UI（模板 / reset to default）用的，
// 与 sensor/cardagg 运行时无关；故不持有 ConfigPublisher，写后不需 fan-out invalidate。
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
//
// 权限矩阵（IsAlarmAccessAllowed）：alarm_cloud READ 全角色放行（B2B + B2C）。
func (s *alarmCloudService) GetAlarmCloudConfig(ctx context.Context, req GetAlarmCloudConfigRequest) (*alarm.AlarmCloudConfig, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	if req.UserRole != "" && s.db != nil {
		allowed, err := IsAlarmAccessAllowed(ctx, s.db, req.TenantID, req.UserRole, AlarmResourceCloud, AlarmActionRead)
		if err != nil {
			s.logger.Warn("alarm access check failed (GetAlarmCloudConfig)",
				zap.String("user_role", req.UserRole),
				zap.String("tenant_id", req.TenantID),
				zap.Error(err),
			)
			return nil, fmt.Errorf("permission check failed: %w", err)
		}
		if !allowed {
			return nil, fmt.Errorf("permission denied: role %q cannot read alarm_cloud config", req.UserRole)
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

			// v2: alarm_cloud 表已并入 spatial_config(longest-prefix-match KV)；
			// repo.UpsertAlarmCloud 直接写 spatial_config 单行 JSONB；created_by/updated_by
			// 字段已不在表中（spatial_config 只跟踪 updated_by），故不再走独立事务路径。
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

	// 从 device_alarms 构建映射表 — 唯一接受 v2 DeviceAlarmEntry 格式 {is_enabled, alarm_level}。
	// v1 老 string 格式（"Fall": "CRITICAL"）不再兼容；DELETE 老行让新 binary 重建即可。
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

	// 4. 解析 CloudVitalAlarmThreshold（从 conditions 列）
	// 写侧（buildDomainAlarmCloudFromConfig）只把 VitalAlarmConditions 存进 conditions 列，
	// 没有 outer "conditions" wrapper；读侧把它包回 CloudVitalAlarmThreshold.Conditions。
	if len(alarmCloud.Conditions) > 0 {
		var inner alarm.VitalAlarmConditions
		if err := json.Unmarshal(alarmCloud.Conditions, &inner); err == nil && (inner.HeartRate != nil || inner.RespiratoryRate != nil) {
			config.CloudVitalAlarmThreshold = alarm.CloudVitalAlarmThreshold{Conditions: &inner}
		} else {
			// 兼容旧格式：早期错误地把整个 CloudVitalAlarmThreshold 序列化（外层多一个 conditions）。
			// 不再走双层 fallback，让 FE 退到默认；老数据 user 重存一次即升级。
			config.CloudVitalAlarmThreshold = alarm.DefaultCloudVitalAlarmThreshold
		}
	} else {
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
	// 仅写入内部 .Conditions（即 VitalAlarmConditions），避免外层 CloudVitalAlarmThreshold
	// 自身带的 "conditions" json tag 造成 {conditions: {conditions: {...}}} 双层嵌套。
	// FE 读侧期望 `config.CloudVitalAlarmThreshold.conditions` = VitalAlarmConditions，
	// 由 buildAlarmCloudConfigFromDomain 在读出时重新包回 CloudVitalAlarmThreshold.Conditions。
	var conditionsJSON []byte
	if config.CloudVitalAlarmThreshold.Conditions != nil {
		conditionsJSON, err = json.Marshal(config.CloudVitalAlarmThreshold.Conditions)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal conditions: %w", err)
		}
	} else {
		conditionsJSON = []byte("{}")
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

	// 1. 构建 device_alarms JSONB — DeviceAlarmEntry 双字段格式（与 buildDomainAlarmCloudFromConfig / read 路径一致）
	// 旧格式 "Fall": "CRITICAL"（字符串）会让 buildAlarmCloudConfigFromDomain 的 json.Unmarshal 撞类型错误（DeviceAlarmEntry 期望 object）。
	// 这里改成 "Fall": {is_enabled: 1, alarm_level: "CRITICAL"} 双字段，读侧 normalize 一致。
	// **包含全部 displayable 项（含 disabled）**：disabled 项也要存 is_enabled=0，否则 read 侧 buildAlarmCloudConfigFromDomain
	// 只能从 defaults 拿到 enabled 状态，写回 FE 会显示成默认 enabled——LeftBed/BedSitUp 应保持 disabled。
	deviceAlarmsTyped := map[string]map[string]DeviceAlarmEntry{}

	sleepPadAlarms := map[string]DeviceAlarmEntry{}
	for _, item := range defaultSetting.Sleepad {
		if item.DisplaySetting != alarm.DisplayAlarmCloud && item.DisplaySetting != alarm.DisplayAlarmCloudAndDevice {
			continue
		}
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
	for _, item := range defaultSetting.Radar {
		if item.DisplaySetting != alarm.DisplayAlarmCloud && item.DisplaySetting != alarm.DisplayAlarmCloudAndDevice {
			continue
		}
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

	deviceAlarmsJSON, err := json.Marshal(deviceAlarmsTyped)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal device_alarms: %w", err)
	}

	// 2. 构建 conditions JSONB（Cloud Vital Alarm Threshold）
	// 与 buildDomainAlarmCloudFromConfig 一致：只 marshal 内部 VitalAlarmConditions，
	// 避免 conditions.conditions 双层嵌套（读侧用 buildAlarmCloudConfigFromDomain 包回 outer）。
	var conditionsJSON []byte
	if alarm.DefaultCloudVitalAlarmThreshold.Conditions != nil {
		conditionsJSON, err = json.Marshal(alarm.DefaultCloudVitalAlarmThreshold.Conditions)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal conditions: %w", err)
		}
	} else {
		conditionsJSON = []byte("{}")
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

// UpdateAlarmCloudConfig 更新告警配置 — v2 spatial_config 版。
//
// v1 该函数走事务：①UPDATE config_versions 旧版本 + INSERT 新版本（审计）②INSERT alarm_cloud ON CONFLICT
// v2 schema 中 config_versions 表已废（各表自带 history），alarm_cloud 表已并入 spatial_config；
// spatial_config 自带审计字段 (updated_at / last_synced_at / updated_by)，无需独立 history 表。
// 因此整段简化为：权限检查 → 构造 domain → repo.UpsertAlarmCloud。
func (s *alarmCloudService) UpdateAlarmCloudConfig(ctx context.Context, req UpdateAlarmCloudConfigRequest) (*alarm.AlarmCloudConfig, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.Config == nil {
		return nil, fmt.Errorf("config is required")
	}
	if req.UserRole == "" {
		return nil, fmt.Errorf("user_role is required for update operation")
	}

	// 权限矩阵：alarm_cloud WRITE 仅 tenant_admin / platform_admin。
	if s.db != nil {
		allowed, err := IsAlarmAccessAllowed(ctx, s.db, req.TenantID, req.UserRole, AlarmResourceCloud, AlarmActionWrite)
		if err != nil {
			s.logger.Warn("alarm access check failed (UpdateAlarmCloudConfig)",
				zap.String("user_role", req.UserRole),
				zap.String("tenant_id", req.TenantID),
				zap.Error(err),
			)
			return nil, fmt.Errorf("permission check failed: %w", err)
		}
		if !allowed {
			return nil, fmt.Errorf("permission denied: role %q cannot update alarm_cloud config", req.UserRole)
		}
	}

	// 业务规则：禁止覆盖系统默认配置（v1 留下的硬编码哨兵 tenant_id，v2 不再写入此 tenant）
	if req.TenantID == "00000000-0000-0000-0000-000000000001" {
		return nil, fmt.Errorf("cannot update system alarm cloud config")
	}

	alarmCloud, err := buildDomainAlarmCloudFromConfig(req.TenantID, req.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to build domain alarm cloud: %w", err)
	}

	if err := s.alarmCloudRepo.UpsertAlarmCloud(ctx, req.TenantID, alarmCloud); err != nil {
		return nil, fmt.Errorf("failed to upsert alarm cloud: %w", err)
	}

	// 不 fan-out invalidate device cache：tenant 级 alarm.cloud_config 是 UI 模板（给新设备
	// bind 和"reset to default"按钮用），sensor/cardagg 运行时只读 alarm.device_config /128，
	// 与 cloud_config 解耦（用户拍板：snapshot model，不级联）。

	// 返回最新配置（走 GET 路径 = 一次 round-trip 验证）
	return s.GetAlarmCloudConfig(ctx, GetAlarmCloudConfigRequest{
		TenantID: req.TenantID,
		UserID:   req.UserID,
		UserRole: req.UserRole,
	})
}
