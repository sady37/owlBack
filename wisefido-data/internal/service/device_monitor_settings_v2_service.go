package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"owl-common/alarm"

	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

// DeviceMonitorSettingsV2Service — owl_v2 spatial_config 版的 device 监控配置服务。
//
// 关键差异（vs v1 service）：
//   - v1 用 alarm_device + alarm_cloud + config_versions + device_store 多表 + qinglan/sleepace 联动
//   - v2 只用 spatial_config 一张表存配置；其它 v1 表已 drop
//
// Phase 1b 范围（仅"存储 + tenant 快照 fallback"）：
//   - GetDeviceMonitorSettings: 读 spatial_config@device_ipv6 → 缺失则 fallback tenant@device_type → 缺失则硬编码默认
//   - UpdateDeviceMonitorSettings: upsert spatial_config@device_ipv6；**device push (qinglan/sleepace) 推迟到 Phase 1c**
//   - GetDefaultDeviceMonitorSettings: 返回硬编码默认（owl-common/alarm.GetDefaultAlarmItems）
//   - CheckDeviceOnlineStatus: 调 qinglan client（保留 v1 行为）
//
// 其余 9 个方法（OTA / firmware upload / resync）当前都返回 ErrPhase1bNotImpl，
// 等具体场景时再单独 v2 化。这些方法对应的 handler 入口 (sleepace resync / sleepace firmware)
// 在 main.go 里仍是条件 wire（service != nil），现在 service 不再 nil，调用方会到达这里并拿到错误，
// 而非 nil panic — 这是符合 v2 cutover 安全方向（fail-loud > nil-deref）。
type DeviceMonitorSettingsV2Service struct {
	db             *sql.DB
	alarmCloudRepo repository.AlarmCloudRepository // tenant 层 spatial_config 读取 (Phase 1a 已 v2)
	qinglanClient  *QinglanClient                  // 雷达在线检查 / device push (Phase 1c 用)
	logger         *zap.Logger
}

const deviceAlarmConfigKey = "alarm.device_config"

// deviceAlarmPacked 是 device-level spatial_config.config_value 的 JSONB 包装。
type deviceAlarmPacked struct {
	AlarmItems []alarm.AlarmItem `json:"alarm_items"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

func NewDeviceMonitorSettingsV2Service(
	db *sql.DB,
	alarmCloudRepo repository.AlarmCloudRepository,
	logger *zap.Logger,
) *DeviceMonitorSettingsV2Service {
	return &DeviceMonitorSettingsV2Service{
		db:             db,
		alarmCloudRepo: alarmCloudRepo,
		logger:         logger,
	}
}

func (s *DeviceMonitorSettingsV2Service) SetQinglanClient(c *QinglanClient) {
	s.qinglanClient = c
}

// 编译期保证实现了完整接口（13 个方法）
var _ DeviceMonitorSettingsService = (*DeviceMonitorSettingsV2Service)(nil)

// errPhase1bNotImpl — 标识"Phase 1b 尚未 v2 化"的统一错误。
func errPhase1bNotImpl(method string) error {
	return fmt.Errorf("DeviceMonitorSettingsV2Service.%s: not implemented (Phase 1b — pending sleepace/OTA v2 cutover)", method)
}

// resolveDeviceIPv6 用 device_id (UUID) 查 devices.device_ipv6 (INET /128)。
// 返回 host(...) 文本表示（不带 mask），用作 spatial_config.spatial_prefix 入参。
func (s *DeviceMonitorSettingsV2Service) resolveDeviceIPv6(ctx context.Context, deviceID string) (string, error) {
	if deviceID == "" {
		return "", fmt.Errorf("device_id is required")
	}
	var ipv6 string
	err := s.db.QueryRowContext(ctx,
		`SELECT host(device_ipv6) FROM devices WHERE device_id = $1::uuid`,
		deviceID,
	).Scan(&ipv6)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("device not found: %s", deviceID)
	}
	if err != nil {
		return "", fmt.Errorf("failed to resolve device ipv6: %w", err)
	}
	return ipv6, nil
}

// readDeviceConfig 读 device 自己的 spatial_config 行。返回 (nil, sql.ErrNoRows) 表示未保存。
func (s *DeviceMonitorSettingsV2Service) readDeviceConfig(ctx context.Context, deviceIPv6 string) ([]alarm.AlarmItem, error) {
	var raw []byte
	err := s.db.QueryRowContext(ctx, `
		SELECT config_value
		  FROM spatial_config
		 WHERE spatial_prefix = $1::inet
		   AND config_key = $2
	`, deviceIPv6, deviceAlarmConfigKey).Scan(&raw)
	if err != nil {
		return nil, err
	}

	var p deviceAlarmPacked
	if jerr := json.Unmarshal(raw, &p); jerr != nil {
		return nil, fmt.Errorf("failed to unmarshal device alarm config: %w", jerr)
	}
	return p.AlarmItems, nil
}

// tenantSnapshot 从 tenant 级 AlarmCloud 中抽出对应 deviceType 的 AlarmItems。
// 该 snapshot 仅作"首次显示默认"用，不持久化到 device 层 — 直到用户主动 Save，
// device 才会有独立的 spatial_config 行。
func (s *DeviceMonitorSettingsV2Service) tenantSnapshot(ctx context.Context, tenantID, deviceType string) ([]alarm.AlarmItem, bool) {
	if s.alarmCloudRepo == nil {
		return nil, false
	}
	ac, err := s.alarmCloudRepo.GetAlarmCloud(ctx, tenantID)
	if err != nil || ac == nil {
		return nil, false
	}
	// AlarmCloud.DeviceAlarms 是双层 map {SleepPad: {AlarmType: {is_enabled, alarm_level}}, Radar: {...}}
	// 这是 v1 schema 的扁平存法；Phase 1b 取 tenant 快照时仍按这种结构理解。
	if len(ac.DeviceAlarms) == 0 {
		return nil, false
	}
	// 拿对应设备的硬编码默认作底，用 tenant 双字段（is_enabled / alarm_level）覆盖。
	deviceKey := ""
	switch deviceType {
	case "sleepad", "Sleepad", "sleepace":
		deviceKey = "SleepPad"
	case "radar", "Radar":
		deviceKey = "Radar"
	default:
		return nil, false
	}

	var allMap map[string]map[string]DeviceAlarmEntry
	if err := json.Unmarshal(ac.DeviceAlarms, &allMap); err != nil {
		return nil, false
	}
	overrides, ok := allMap[deviceKey]
	if !ok || len(overrides) == 0 {
		return nil, false
	}

	defaults := alarm.GetDefaultAlarmItems(deviceType)
	result := make([]alarm.AlarmItem, 0, len(defaults))
	for _, item := range defaults {
		// 深拷贝指针字段
		newItem := item
		if item.AlarmLevel != nil {
			lv := *item.AlarmLevel
			newItem.AlarmLevel = &lv
		}
		if item.IsEnabled != nil {
			en := *item.IsEnabled
			newItem.IsEnabled = &en
		}
		if entry, exists := overrides[item.AlarmType]; exists {
			enabled := entry.IsEnabled
			newItem.IsEnabled = &enabled
			if entry.AlarmLevel != "" {
				lv := entry.AlarmLevel
				newItem.AlarmLevel = &lv
			}
		}
		result = append(result, newItem)
	}
	return result, true
}

// ---- DeviceMonitorSettingsService interface ----

// GetDeviceMonitorSettings — Phase 1b 派生时快照模型：
//  1. 先读 device 自己的 spatial_config；
//  2. 行不存在 → 取 tenant snapshot 当 "首次默认"（不持久化）；
//  3. tenant 也无 → 返回 owl-common/alarm 硬编码。
func (s *DeviceMonitorSettingsV2Service) GetDeviceMonitorSettings(ctx context.Context, tenantID, deviceID, deviceType string) ([]alarm.AlarmItem, error) {
	deviceIPv6, err := s.resolveDeviceIPv6(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	items, err := s.readDeviceConfig(ctx, deviceIPv6)
	if err == nil {
		return items, nil
	}
	if err != sql.ErrNoRows {
		// JSON 解析失败或其它 SQL 错误 — 用 warn 记录后退回 tenant 快照，避免 page 完全无法加载
		s.logger.Warn("failed to read device alarm config; falling back to tenant snapshot",
			zap.String("device_id", deviceID),
			zap.String("device_ipv6", deviceIPv6),
			zap.Error(err),
		)
	}

	if snapshot, ok := s.tenantSnapshot(ctx, tenantID, deviceType); ok {
		return snapshot, nil
	}
	return alarm.GetDefaultAlarmItems(deviceType), nil
}

// UpdateDeviceMonitorSettings — Phase 1b 仅写 DB；device push（qinglan/sleepace）推迟。
//
// 返回值兼容 v1 handler 期望：
//   - radar → *UpdateRadarResult (DBWriteSuccess=true, DeviceWriteSuccess=false 表示"已存DB，未推设备")
//   - sleepad → map[string]interface{}{"success": true, "device_write": false, ...}
//
// TODO Phase 1c: 调 s.qinglanClient (radar) / sleepaceGateway (sleepad) 把 alarmItems 实际下发到设备硬件。
func (s *DeviceMonitorSettingsV2Service) UpdateDeviceMonitorSettings(
	ctx context.Context, tenantID, deviceID, deviceType, userID string,
	alarmItems []alarm.AlarmItem, progressCallback ProgressCallback,
) (interface{}, error) {

	deviceIPv6, err := s.resolveDeviceIPv6(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(deviceAlarmPacked{
		AlarmItems: alarmItems,
		UpdatedAt:  time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal alarm items: %w", err)
	}

	var updatedBy interface{}
	if userID != "" {
		updatedBy = userID
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO spatial_config (
			spatial_prefix, config_key, config_value,
			source, state_db, last_synced_at, updated_at, updated_by
		) VALUES (
			$1::inet, $2, $3::jsonb,
			'manual_ui', 3, now(), now(), $4::uuid
		)
		ON CONFLICT (spatial_prefix, config_key) DO UPDATE SET
			config_value   = EXCLUDED.config_value,
			source         = EXCLUDED.source,
			state_db       = EXCLUDED.state_db,
			last_synced_at = EXCLUDED.last_synced_at,
			updated_at     = EXCLUDED.updated_at,
			updated_by     = EXCLUDED.updated_by
	`, deviceIPv6, deviceAlarmConfigKey, string(payload), updatedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert device alarm config: %w", err)
	}

	switch deviceType {
	case "radar", "Radar":
		return &UpdateRadarResult{
			DeviceWriteSuccess: false, // Phase 1c TODO
			DBWriteSuccess:     true,
			FailedAlarmTypes:   nil,
			NoChange:           false,
		}, nil
	default:
		return map[string]interface{}{
			"success":        true,
			"device_write":   false, // Phase 1c TODO
			"database_write": true,
			"no_change":      false,
		}, nil
	}
}

// GetDefaultDeviceMonitorSettings — 直接返回硬编码默认，与 handler 中的旁路逻辑等价。
func (s *DeviceMonitorSettingsV2Service) GetDefaultDeviceMonitorSettings(ctx context.Context, tenantID, deviceType string) ([]alarm.AlarmItem, error) {
	return alarm.GetDefaultAlarmItems(deviceType), nil
}

// CheckDeviceOnlineStatus — 调 qinglan 检查雷达在线（Sleepad 走另一条路径，但 v1 这里只覆盖雷达）。
func (s *DeviceMonitorSettingsV2Service) CheckDeviceOnlineStatus(ctx context.Context, deviceUID string) error {
	if s.qinglanClient == nil {
		return fmt.Errorf("qinglan client not initialized")
	}
	if deviceUID == "" {
		return fmt.Errorf("device_uid is required")
	}
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := s.qinglanClient.GetDeviceProperties(checkCtx, deviceUID, []string{"radar_func_ctrl"})
	return err
}

// ---- 以下 9 个方法在 Phase 1b 暂不实现 ----

func (s *DeviceMonitorSettingsV2Service) ResyncDeviceTimezone(ctx context.Context, tenantID, deviceID string) (string, int, error) {
	return "", 0, errPhase1bNotImpl("ResyncDeviceTimezone")
}

func (s *DeviceMonitorSettingsV2Service) ResyncDeviceReportTime(ctx context.Context, tenantID, deviceID string) (int, error) {
	return 0, errPhase1bNotImpl("ResyncDeviceReportTime")
}

func (s *DeviceMonitorSettingsV2Service) TriggerSleepaceUpgrade(ctx context.Context, tenantID, deviceID, version string) error {
	return errPhase1bNotImpl("TriggerSleepaceUpgrade")
}

func (s *DeviceMonitorSettingsV2Service) UploadSleepaceFirmware(ctx context.Context, filename string, file io.Reader) error {
	return errPhase1bNotImpl("UploadSleepaceFirmware")
}

func (s *DeviceMonitorSettingsV2Service) DeleteSleepaceFirmware(ctx context.Context, deviceType int, deviceVersion string) error {
	return errPhase1bNotImpl("DeleteSleepaceFirmware")
}

func (s *DeviceMonitorSettingsV2Service) ListSleepaceFirmwareVersions(ctx context.Context) ([]SleepaceFirmwareVersion, error) {
	return nil, errPhase1bNotImpl("ListSleepaceFirmwareVersions")
}

func (s *DeviceMonitorSettingsV2Service) LocalUploadSleepaceFirmware(ctx context.Context, filename string, file io.Reader, deviceType, deviceVersion string) (bool, error) {
	return false, errPhase1bNotImpl("LocalUploadSleepaceFirmware")
}

func (s *DeviceMonitorSettingsV2Service) LocalDeleteSleepaceFirmware(ctx context.Context, filename string) error {
	return errPhase1bNotImpl("LocalDeleteSleepaceFirmware")
}

func (s *DeviceMonitorSettingsV2Service) ResolveSleepaceUpgradeVersion(filename string) (string, error) {
	return "", errPhase1bNotImpl("ResolveSleepaceUpgradeVersion")
}
