package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"owl-common/alarm"

	"wisefido-data/internal/publisher"
	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

type ProgressCallback func(percent int, message string)

type DeviceMonitorSettingsService interface {
	GetDeviceMonitorSettings(ctx context.Context, tenantID, deviceAddr, deviceType string) ([]alarm.AlarmItem, error)
	UpdateDeviceMonitorSettings(ctx context.Context, tenantID, deviceAddr, deviceType, userID string, alarmItems []alarm.AlarmItem, progressCallback ProgressCallback) (interface{}, error)
	GetDefaultDeviceMonitorSettings(ctx context.Context, tenantID string, deviceType string) ([]alarm.AlarmItem, error)
	CheckDeviceOnlineStatus(ctx context.Context, deviceUID string) error
	ResyncDeviceTimezone(ctx context.Context, tenantID, deviceAddr string) (string, int, error)
	ResyncDeviceReportTime(ctx context.Context, tenantID, deviceAddr string) (int, error)
	TriggerSleepaceUpgrade(ctx context.Context, tenantID, deviceAddr, version string) error
	UploadSleepaceFirmware(ctx context.Context, filename string, file io.Reader) error
	DeleteSleepaceFirmware(ctx context.Context, deviceType int, deviceVersion string) error
	ListSleepaceFirmwareVersions(ctx context.Context) ([]SleepaceFirmwareVersion, error)
	LocalUploadSleepaceFirmware(ctx context.Context, filename string, file io.Reader, deviceType, deviceVersion string) (pushedToVendor bool, err error)
	LocalDeleteSleepaceFirmware(ctx context.Context, filename string) error
	ResolveSleepaceUpgradeVersion(filename string) (string, error)
}

type RadarWriteResult struct {
	Results map[string]string
}

type UpdateRadarResult struct {
	DeviceWriteSuccess bool
	DBWriteSuccess     bool
	FailedAlarmTypes   []alarm.AlarmItem
	NoChange           bool
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

// deviceMonitorSettingsService backs config storage in spatial_config.
//   - Get reads spatial_config@device_ipv6 → fallback tenant@device_type → hardcoded default
//   - Update upserts spatial_config@device_ipv6; device push (qinglan/sleepace) deferred
//   - 9 OTA/firmware/resync methods return errNotImplemented (Phase 1c+ wiring)
type deviceMonitorSettingsService struct {
	db              *sql.DB
	alarmCloudRepo  repository.AlarmCloudRepository // tenant 层 spatial_config 读取 (Phase 1a 已 v2)
	qinglanClient   *QinglanClient                  // 雷达在线检查 / device push (Phase 1c 用)
	sleepaceGateway *SleepaceGatewayClient          // sleepad 厂家 HTTP 下发；nil = 不可下发，UI 会看到 device_write=false
	configPublisher *publisher.ConfigPublisher      // 写 spatial_config 后 publish config:alarmDevice:stream 让 sensor 失效本地 cache
	logger          *zap.Logger
}

const deviceAlarmConfigKey = "alarm.device_config"

// deviceAlarmPacked 是 device-level spatial_config.config_value 的 JSONB 包装。
type deviceAlarmPacked struct {
	AlarmItems []alarm.AlarmItem `json:"alarm_items"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

func NewDeviceMonitorSettingsService(
	db *sql.DB,
	alarmCloudRepo repository.AlarmCloudRepository,
	logger *zap.Logger,
) *deviceMonitorSettingsService {
	return &deviceMonitorSettingsService{
		db:             db,
		alarmCloudRepo: alarmCloudRepo,
		logger:         logger,
	}
}

func (s *deviceMonitorSettingsService) SetQinglanClient(c *QinglanClient) {
	s.qinglanClient = c
}

// SetSleepaceGatewayClient — main.go 在 startup 时通过 type assertion 调用此方法注入。
// 不实现这个方法 sleepad 配置就不会下发到厂家（device_write 永远 false）。
func (s *deviceMonitorSettingsService) SetSleepaceGatewayClient(c *SleepaceGatewayClient) {
	s.sleepaceGateway = c
}

// SetConfigPublisher — main.go 注入；nil 时跳过 publish（sensor 端依赖 lazy reload，
// 但失效延迟变长——所以生产必须 inject）。
func (s *deviceMonitorSettingsService) SetConfigPublisher(p *publisher.ConfigPublisher) {
	s.configPublisher = p
}

var _ DeviceMonitorSettingsService = (*deviceMonitorSettingsService)(nil)

func errNotImplemented(method string) error {
	return fmt.Errorf("deviceMonitorSettingsService.%s: not implemented", method)
}

// resolveDeviceIPv6 用 device_id (UUID) 查 devices.device_ipv6 (INET /128)。
// 返回 host(...) 文本表示（不带 mask），用作 spatial_config.spatial_prefix 入参。
// deviceAddr 入参承载 device_addr (INET text)。
func (s *deviceMonitorSettingsService) resolveDeviceIPv6(ctx context.Context, deviceAddr string) (string, error) {
	if deviceAddr == "" {
		return "", fmt.Errorf("device_addr is required")
	}
	var ipv6 string
	err := s.db.QueryRowContext(ctx,
		`SELECT host(device_addr) FROM devices WHERE device_addr = $1::INET`,
		deviceAddr,
	).Scan(&ipv6)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("device not found: %s", deviceAddr)
	}
	if err != nil {
		return "", fmt.Errorf("failed to resolve device addr: %w", err)
	}
	return ipv6, nil
}

// resolveDeviceCode 由 device_addr (INET text) 查 device_factory_meta.device_code（合作方 deviceId）。
// 取不到 device_code 不算 hard error — 返回空串，由调用方根据 deviceType 决定是否需要。
func (s *deviceMonitorSettingsService) resolveDeviceCode(ctx context.Context, deviceAddr string) (string, error) {
	if deviceAddr == "" {
		return "", fmt.Errorf("device_addr is required")
	}
	var code sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT dfm.device_code
		   FROM device_factory_meta dfm
		   JOIN devices d ON d.device_uid = dfm.device_uid
		  WHERE d.device_addr = $1::INET`,
		deviceAddr,
	).Scan(&code)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("device_factory_meta not found: %s", deviceAddr)
	}
	if err != nil {
		return "", fmt.Errorf("failed to resolve device_code: %w", err)
	}
	if !code.Valid {
		return "", nil
	}
	return code.String, nil
}

// resolveDeviceUID 由 device_addr (INET text) 查 devices.device_uid（雷达走 qinglan 时的入参）。
// device_uid 是 NOT NULL 字段，缺失即 hard error。
func (s *deviceMonitorSettingsService) resolveDeviceUID(ctx context.Context, deviceAddr string) (string, error) {
	if deviceAddr == "" {
		return "", fmt.Errorf("device_addr is required")
	}
	var uid string
	err := s.db.QueryRowContext(ctx,
		`SELECT device_uid FROM devices WHERE device_addr = $1::INET`,
		deviceAddr,
	).Scan(&uid)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("device not found: %s", deviceAddr)
	}
	if err != nil {
		return "", fmt.Errorf("failed to resolve device_uid: %w", err)
	}
	return uid, nil
}

// resolveTenantResetTime 从 alarm_cloud.metadata 取租户作息（rest 时段）。
// 缺失返回 nil — ConvertAlarmItemsToSleepaceConfig 收到 nil 时不会写 leftBedStart/End 字段。
func (s *deviceMonitorSettingsService) resolveTenantResetTime(ctx context.Context, tenantID string) *alarm.ResetTimeParams {
	if s.alarmCloudRepo == nil || tenantID == "" {
		return nil
	}
	ac, err := s.alarmCloudRepo.GetAlarmCloud(ctx, tenantID)
	if err != nil || ac == nil || len(ac.Metadata) == 0 {
		return nil
	}
	var tr alarm.TenantResetTime
	if json.Unmarshal(ac.Metadata, &tr) != nil {
		return nil
	}
	if tr.ResetTime.InBedTime == "" || tr.ResetTime.OutBedTime == "" {
		return nil
	}
	return &tr.ResetTime
}

// readDeviceConfig 读 device 自己的 spatial_config 行。返回 (nil, sql.ErrNoRows) 表示未保存。
func (s *deviceMonitorSettingsService) readDeviceConfig(ctx context.Context, deviceIPv6 string) ([]alarm.AlarmItem, error) {
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
func (s *deviceMonitorSettingsService) tenantSnapshot(ctx context.Context, tenantID, deviceType string) ([]alarm.AlarmItem, bool) {
	if s.alarmCloudRepo == nil {
		return nil, false
	}
	ac, err := s.alarmCloudRepo.GetAlarmCloud(ctx, tenantID)
	if err != nil || ac == nil {
		return nil, false
	}
	// AlarmCloud.DeviceAlarms 是双层 map {SleepPad: {AlarmType: {is_enabled, alarm_level}}, Radar: {...}}
	// 扁平存法，tenant 快照按这种结构理解。
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

// GetDeviceMonitorSettings — 三层 merge（low→high）：
//   1. alarm.go GetDefaultAlarmItems(deviceType)  ← 出厂默认（必有，完整字段）
//   2. tenant alarm_cloud snapshot（如有）         ← 租户级 override
//   3. device spatial_config（如有）               ← user 显式保存的 per-device override
//
// 每一层按 alarm_type + per-key 覆盖上一层；后一层显式给的字段才覆盖。
// FE 拿到的 alarm_items 永远完整，模板里 hardcoded fallback 只是 BE 出错时的 emergency safety net。
func (s *deviceMonitorSettingsService) GetDeviceMonitorSettings(ctx context.Context, tenantID, deviceAddr, deviceType string) ([]alarm.AlarmItem, error) {
	deviceIPv6, err := s.resolveDeviceIPv6(ctx, deviceAddr)
	if err != nil {
		return nil, err
	}

	// Layer 1+2: default + tenant snapshot（与 GetDefaultDeviceMonitorSettings 同源，保证 "Load Defaults" 视图与正常视图共用 base）
	result := s.buildTenantDefaultLayer(ctx, tenantID, deviceType)

	// Layer 3: user device override（spatial_config 行存在才叠）
	if items, err := s.readDeviceConfig(ctx, deviceIPv6); err == nil {
		result = mergeAlarmItemsOverDefaults(result, items)
	} else if err != sql.ErrNoRows {
		// JSON 解析失败或其它 SQL 错误 — warn 记录，不阻塞 page；前两层已足以渲染。
		s.logger.Warn("failed to read device alarm config; tenant-default layer used",
			zap.String("device_addr", deviceAddr),
			zap.String("device_ipv6", deviceIPv6),
			zap.Error(err),
		)
	}
	return result, nil
}

// buildTenantDefaultLayer 复合 alarm.go 默认 + tenant snapshot —— "Load Defaults" 按钮和正常 page load 共用此 base。
// 没有 tenant snapshot 时退回纯 alarm.go 默认。
func (s *deviceMonitorSettingsService) buildTenantDefaultLayer(ctx context.Context, tenantID, deviceType string) []alarm.AlarmItem {
	defaults := alarm.GetDefaultAlarmItems(deviceType)
	if snapshot, ok := s.tenantSnapshot(ctx, tenantID, deviceType); ok {
		return mergeAlarmItemsOverDefaults(defaults, snapshot)
	}
	return defaults
}

// mergeAlarmItemsOverDefaults 把 user override 叠在 alarm.go 默认之上，按 alarm_type 一一合并；
// alarm_params 内部按 key 合并（user 显式 set 的覆盖；user 未给的 key 取 default）。
// is_enabled / alarm_level / display_setting：user 给了非零/非 nil 则用 user 的，否则用 default。
//
// 为什么放服务层做：FE 拿到的 alarm_items 永远完整 → 模板里 getAlarmParam(..., hardcoded) 的 hardcoded
// 兜底永远不触发（防御性保留即可），alarm.go 任何时候改默认 FE 自动跟。
func mergeAlarmItemsOverDefaults(defaults, overrides []alarm.AlarmItem) []alarm.AlarmItem {
	if len(defaults) == 0 {
		return overrides
	}
	overrideMap := make(map[string]alarm.AlarmItem, len(overrides))
	for _, it := range overrides {
		overrideMap[it.AlarmType] = it
	}
	out := make([]alarm.AlarmItem, 0, len(defaults)+len(overrides))
	seen := make(map[string]struct{}, len(defaults))
	for _, def := range defaults {
		seen[def.AlarmType] = struct{}{}
		ov, ok := overrideMap[def.AlarmType]
		if !ok {
			out = append(out, def)
			continue
		}
		merged := def
		if ov.IsEnabled != nil {
			merged.IsEnabled = ov.IsEnabled
		}
		if ov.AlarmLevel != nil {
			merged.AlarmLevel = ov.AlarmLevel
		}
		if ov.DisplaySetting != 0 {
			merged.DisplaySetting = ov.DisplaySetting
		}
		// alarm_params: per-key merge — default 是底，user 给的 key 覆盖。
		params := make(map[string]interface{}, len(def.AlarmParams)+len(ov.AlarmParams))
		for k, v := range def.AlarmParams {
			params[k] = v
		}
		for k, v := range ov.AlarmParams {
			params[k] = v
		}
		merged.AlarmParams = params
		out = append(out, merged)
	}
	// override 里若有 default 没有的 alarm_type（向前兼容老数据），原样追加。
	for _, ov := range overrides {
		if _, dup := seen[ov.AlarmType]; dup {
			continue
		}
		out = append(out, ov)
	}
	return out
}

// UpdateDeviceMonitorSettings — 写 spatial_config + 推到硬件（sleepad 走 sleepace gateway）。
//
// 返回值兼容 v1 handler 期望：
//   - radar → *UpdateRadarResult (radar push 仍是 Phase 1c TODO)
//   - sleepad → map[string]interface{}{"success", "device_write", "db_write", "no_change"}
//
// 注：v2 字段名升级 "db_write" → "database_write"；FE 也已同步读 database_write。
// 历史 v1 仍输出 db_write，那一段是死代码（main.go 已只 wire v2 service）。
//
// 与 v1 区别：
//   - 不做 diff（每次全量下发，sleepace updatealarmnotifyconfig 是幂等替换语义）
//   - 不读 v1 的 alarm_device 表对比 — 该表已 drop，DB 现在是 spatial_config 单一来源
//   - 错误处理：硬件下发失败仍写 DB（v1 是先推硬件再写库，失败回滚；v2 改为先写库再推，让 UI 有 partial 状态可显示）
func (s *deviceMonitorSettingsService) UpdateDeviceMonitorSettings(
	ctx context.Context, tenantID, deviceAddr, deviceType, userID string,
	alarmItems []alarm.AlarmItem, progressCallback ProgressCallback,
) (interface{}, error) {

	deviceIPv6, err := s.resolveDeviceIPv6(ctx, deviceAddr)
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
			spatial_prefix, config_key, config_value, device_uid,
			source, state_db, last_synced_at, updated_at, updated_by
		) VALUES (
			$1::inet, $2, $3::jsonb, (SELECT device_uid FROM devices WHERE device_addr = $1::inet),
			'manual_ui', 3, now(), now(), $4::uuid
		)
		ON CONFLICT (spatial_prefix, config_key) DO UPDATE SET
			config_value   = EXCLUDED.config_value,
			source         = EXCLUDED.source,
			state_db       = EXCLUDED.state_db,
			device_uid     = EXCLUDED.device_uid,
			last_synced_at = EXCLUDED.last_synced_at,
			updated_at     = EXCLUDED.updated_at,
			updated_by     = EXCLUDED.updated_by
	`, deviceIPv6, deviceAlarmConfigKey, string(payload), updatedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert device alarm config: %w", err)
	}
	dbWrite := true

	// 失效 sensor/cardagg 端 alarm enablement cache（key=alarm.device_config /128）。
	// 失败仅 warn — DB 已写入，sensor cache 自动 lazy reload（窗口最大 ~5min activeNoise，
	// 实践中下一次 alarm 事件命中即触发）。fire-and-forget 不阻 UI 响应。
	if s.configPublisher != nil {
		if perr := s.configPublisher.PublishAlarmDeviceMessage(
			ctx, tenantID, deviceIPv6, "", "alarm.device_config",
			map[string]interface{}{"alarm_items": alarmItems},
		); perr != nil {
			s.logger.Warn("publish alarm.device_config invalidate failed (non-fatal; cache will lazy reload)",
				zap.String("device_addr", deviceIPv6), zap.Error(perr))
		}
	}

	switch deviceType {
	case "radar", "Radar":
		deviceWrite, pushErr := s.pushRadarToHardware(ctx, deviceAddr, alarmItems, progressCallback)
		result := &UpdateRadarResult{
			DeviceWriteSuccess: deviceWrite,
			DBWriteSuccess:     dbWrite,
			FailedAlarmTypes:   nil,
			NoChange:           false,
		}
		if pushErr != nil {
			s.logger.Error("[RADAR_WRITE_V2] qinglan push failed",
				zap.String("device_addr", deviceAddr),
				zap.Error(pushErr),
			)
			// device 推失败，但 DB 已写：返回 result 让 handler 透传到 FE 显示半失败状态；
			// error 字段保留给上游做 log 用，不影响 dbWrite=true 的事实
		}
		return result, nil
	default:
		// sleepad / Sleepad / sleepace
		deviceWrite, pushErr := s.pushSleepadToHardware(ctx, tenantID, deviceAddr, alarmItems, progressCallback)
		result := map[string]interface{}{
			"success":        dbWrite, // 与 v1 一致：success 反映 DB 状态
			"device_write":   deviceWrite,
			"database_write": dbWrite,
			"no_change":      false,
		}
		if pushErr != nil {
			result["device_write_error"] = pushErr.Error()
		}
		return result, nil
	}
}

// pushRadarToHardware 调 qinglan client 把 alarmItems 整体下发到 HC2 雷达。
// wisefido-qinglan 自己解析 _alarm_items_json 并构造 fall_param / heart_breath_param / radar_func_ctrl
// 三组属性按 100ms 间隔分批 PUT 到设备，所以这里只需一次调用。
//
// 返回 (deviceWrite, error)：error 仅作 log 用，调用方根据 deviceWrite=false 决定 UI 提示。
func (s *deviceMonitorSettingsService) pushRadarToHardware(
	ctx context.Context, deviceAddr string,
	alarmItems []alarm.AlarmItem, progressCallback ProgressCallback,
) (bool, error) {
	if s.qinglanClient == nil {
		return false, fmt.Errorf("qinglan client not configured")
	}
	deviceUID, err := s.resolveDeviceUID(ctx, deviceAddr)
	if err != nil {
		return false, err
	}
	if deviceUID == "" {
		return false, fmt.Errorf("device_uid missing for device_addr=%s", deviceAddr)
	}

	if progressCallback != nil {
		progressCallback(30, "pushing alarm config to radar...")
	}

	alarmItemsJSON, err := json.Marshal(alarmItems)
	if err != nil {
		return false, fmt.Errorf("marshal alarm items: %w", err)
	}

	s.logger.Info("[RADAR_WRITE_V2] sending alarm items to device",
		zap.String("device_addr", deviceAddr),
		zap.String("device_uid", deviceUID),
		zap.Int("alarm_items_count", len(alarmItems)),
	)

	// qinglanClient 内部按 fall_param / heart_breath_param / radar_func_ctrl 分组并加 100ms 间隔
	if _, err := s.qinglanClient.SetDeviceProperties(ctx, deviceUID, map[string]interface{}{
		"_alarm_items_json": string(alarmItemsJSON),
	}); err != nil {
		return false, fmt.Errorf("qinglan SetDeviceProperties: %w", err)
	}

	if progressCallback != nil {
		progressCallback(100, "radar config pushed")
	}
	s.logger.Info("[RADAR_WRITE_V2] device push succeeded",
		zap.String("device_addr", deviceAddr),
		zap.String("device_uid", deviceUID),
	)
	return true, nil
}

// pushSleepadToHardware 把 alarmItems 翻译成厂家协议并下发：
//  1. updatealarmnotifyconfig — 报警阈值（heart/breath/leftBed/onbed/...）
//  2. SleepadSetting/MaterialSetting 项的辅助 set 接口（realtime_interval / leave_sensibility / ...）
//
// 任意一步失败都返回 deviceWrite=false + error，让 UI 能显式告警。
func (s *deviceMonitorSettingsService) pushSleepadToHardware(
	ctx context.Context, tenantID, deviceAddr string,
	alarmItems []alarm.AlarmItem, progressCallback ProgressCallback,
) (bool, error) {
	if s.sleepaceGateway == nil {
		return false, fmt.Errorf("sleepace gateway not configured")
	}
	deviceCode, err := s.resolveDeviceCode(ctx, deviceAddr)
	if err != nil {
		return false, err
	}
	if deviceCode == "" {
		return false, fmt.Errorf("device_code missing for device_addr=%s (device_factory_meta.device_code is null)", deviceAddr)
	}
	deviceUID, err := s.resolveDeviceUID(ctx, deviceAddr)
	if err != nil {
		return false, err
	}
	if deviceUID == "" {
		return false, fmt.Errorf("device_uid missing for device_addr=%s", deviceAddr)
	}

	if progressCallback != nil {
		progressCallback(30, "pushing alarm config to sleepad hardware...")
	}

	resetTime := s.resolveTenantResetTime(ctx, tenantID)
	sleepaceConfig := ConvertAlarmItemsToSleepaceConfig(deviceCode, deviceUID, alarmItems, resetTime)

	s.logger.Info("[SLEEPAD_WRITE_V2] sending alarm config to hardware",
		zap.String("tenant_id", tenantID),
		zap.String("device_addr", deviceAddr),
		zap.String("device_uid", deviceUID),
		zap.String("device_code", deviceCode),
		zap.Any("leftBedFlag", sleepaceConfig["leftBedFlag"]),
		zap.Any("leftBedDuration", sleepaceConfig["leftBedDuration"]),
	)

	if err := s.sleepaceGateway.UpdateAlarmConfig(ctx, sleepaceConfig); err != nil {
		s.logger.Error("[SLEEPAD_WRITE_V2] updatealarmnotifyconfig failed",
			zap.String("device_uid", deviceUID),
			zap.String("device_code", deviceCode),
			zap.Error(err),
		)
		return false, fmt.Errorf("updatealarmnotifyconfig: %w", err)
	}

	if progressCallback != nil {
		progressCallback(60, "alarm config pushed; pushing device settings...")
	}

	// SleepadSetting / MaterialSetting 各项辅助参数（best-effort：单项失败 warn 但不阻断主结果）
	s.pushSleepadSettings(ctx, deviceUID, deviceCode, alarmItems)

	// timezone 在 Sleep Monitoring Config 由人设置（不在 bind 处），所以 save 时必须把设备 unit 时区
	// 的当前偏移（含 DST）重 bind 到 sleepace —— 驱动厂家侧 ResetTime/leftbed 判窗与报告 startTime。
	// best-effort：失败 warn 不阻断主结果。
	if iana, off, err := s.ResyncDeviceTimezone(ctx, tenantID, deviceAddr); err != nil {
		s.logger.Warn("resync device timezone on settings save",
			zap.String("device_addr", deviceAddr), zap.Error(err))
	} else {
		s.logger.Info("device timezone re-bound on settings save",
			zap.String("device_addr", deviceAddr), zap.String("iana", iana), zap.Int("offset_seconds", off))
	}

	if progressCallback != nil {
		progressCallback(100, "device push completed")
	}
	return true, nil
}

// pushSleepadSettings — SleepadSetting / MaterialSetting 项的非 alarm-config 设置（realtime/leaveSens/床垫…）。
// 走多个独立 set 接口，单项失败不算整体失败（与 v1 行为一致）。
// deviceUID = device_factory_meta.device_uid (logMAC)，作 Sleepace API userId。
func (s *deviceMonitorSettingsService) pushSleepadSettings(
	ctx context.Context, deviceUID, deviceCode string, items []alarm.AlarmItem,
) {
	for _, item := range items {
		switch item.AlarmType {
		case alarm.SleepadSetting:
			p := item.AlarmParams
			if len(p) == 0 {
				continue
			}
			// realtime_interval 由 SleepaceIntervalScheduler 独占（默认 10s，在 reset_time + post-meal(12:00-14:00) 窗口内切 2s）；
			// FE save 路径若也 push 会与 scheduler 抢，覆盖窗口外的 10 → 永远卡在用户值，scheduler 因 dedup 不会自愈。
			sensV, sensOk := toIntParam(p["Bed_Exit_Sensitivity"])
			if !sensOk {
				sensV, sensOk = toIntParam(p["leave_sensibility"])
			}
			if sensOk {
				if err := s.sleepaceGateway.SetLeaveSensibility(ctx, deviceUID, deviceCode, sensV); err != nil {
					s.logger.Warn("[SLEEPAD_WRITE_V2] SetLeaveSensibility", zap.Error(err))
				}
			}
			if v, ok := toIntParam(p["Empty_Bed_Monitor"]); ok && (v == 0 || v == 1) {
				if err := s.sleepaceGateway.SetRealtimeModeAfterLeave(ctx, deviceUID, deviceCode, v); err != nil {
					s.logger.Warn("[SLEEPAD_WRITE_V2] SetRealtimeModeAfterLeave", zap.Error(err))
				}
			}
			if v, ok := toIntParam(p["report_upload_type"]); ok {
				if err := s.sleepaceGateway.SetReportUploadType(ctx, deviceUID, deviceCode, v); err != nil {
					s.logger.Warn("[SLEEPAD_WRITE_V2] SetReportUploadType", zap.Error(err))
				}
				// report_upload_time 由 SleepaceReportTimeScheduler 独占管理，不在这里下发。
			}
		case alarm.MaterialSetting:
			p := item.AlarmParams
			if len(p) == 0 {
				continue
			}
			thickness, tOk := toIntParam(p["thickness"])
			material, mOk := toIntParam(p["material_type"])
			if tOk && mOk {
				if err := s.sleepaceGateway.SetBedParameters(ctx, deviceUID, deviceCode, thickness, material); err != nil {
					s.logger.Warn("[SLEEPAD_WRITE_V2] SetBedParameters", zap.Error(err))
				}
			}
		}
	}
}

// GetDefaultDeviceMonitorSettings — "Load Defaults" 按钮调用：alarm.go 默认 + tenant snapshot，
// 不叠加 user device override（按设计就是要给出"还原成默认/租户标准"的视图，跟当前 user 保存值无关）。
func (s *deviceMonitorSettingsService) GetDefaultDeviceMonitorSettings(ctx context.Context, tenantID, deviceType string) ([]alarm.AlarmItem, error) {
	return s.buildTenantDefaultLayer(ctx, tenantID, deviceType), nil
}

// CheckDeviceOnlineStatus — 调 qinglan 检查雷达在线（Sleepad 走另一条路径，但 v1 这里只覆盖雷达）。
func (s *deviceMonitorSettingsService) CheckDeviceOnlineStatus(ctx context.Context, deviceUID string) error {
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

// ---- 以下 9 个方法暂不实现（OTA / firmware / resync wiring 待接） ----

// ResyncDeviceTimezone 把设备所在 unit 的 IANA 时区在「当前时刻」的 UTC 偏移（含 DST）重新 bind 到 sleepace。
//
// 背景：timezone 在 /sleepace/bind 设（秒偏移），驱动厂家侧 ResetTime / leftbed 判窗与报告 startTime。
// 初始 bind 曾用全局 cfg.Timezone（一刀切 Denver），导致非 Denver 的 unit（如深圳）行为全错。这里改用
// 设备 unit 的真实 IANA → 当前 offset 重 bind。DST 由调用方（save / 每日 today!=yesterday 检测）触发，
// IANAToOffsetSeconds 用 time.Now() 取当下偏移，切换日自然得到新值。
//
// 返回 (iana, offsetSeconds, error)。
func (s *deviceMonitorSettingsService) ResyncDeviceTimezone(ctx context.Context, tenantID, deviceAddr string) (string, int, error) {
	if s.sleepaceGateway == nil {
		return "", 0, fmt.Errorf("sleepace gateway not wired")
	}
	if tenantID == "" || deviceAddr == "" {
		return "", 0, fmt.Errorf("tenant_id and device_addr are required")
	}

	// IANA 时区级联 unit → branch → tenant（与 GetDeviceMonitorSettings 同源；tenant.timezone NOT NULL）。
	var tz sql.NullString
	if err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(NULLIF(u.timezone,''), NULLIF(br.timezone,''), NULLIF(t.timezone,''))
		  FROM devices d
		  LEFT JOIN units    u  ON u.unit_id    = network(set_masklen(d.device_addr, 80))
		  LEFT JOIN branches br ON br.branch_id = network(set_masklen(d.device_addr, 56))
		  LEFT JOIN tenants  t  ON t.tenant_id  = network(set_masklen(d.device_addr, 48))
		 WHERE d.device_addr = $1::INET
	`, deviceAddr).Scan(&tz); err != nil {
		return "", 0, fmt.Errorf("resolve unit timezone: %w", err)
	}
	iana := strings.TrimSpace(tz.String)
	if iana == "" {
		return "", 0, fmt.Errorf("no IANA timezone for device %s", deviceAddr)
	}
	offset := IANAToOffsetSeconds(iana) // 当前偏移，含 DST

	deviceCode, err := s.resolveDeviceCode(ctx, deviceAddr)
	if err != nil {
		return "", 0, fmt.Errorf("resolve device_code: %w", err)
	}
	deviceUID, err := s.resolveDeviceUID(ctx, deviceAddr)
	if err != nil {
		return "", 0, fmt.Errorf("resolve device_uid: %w", err)
	}

	if _, err := s.sleepaceGateway.InitializeDevice(ctx, deviceCode, deviceUID, &offset); err != nil {
		return "", 0, fmt.Errorf("sleepace re-bind timezone: %w", err)
	}
	s.logger.Info("ResyncDeviceTimezone done",
		zap.String("tenant_id", tenantID),
		zap.String("device_addr", deviceAddr),
		zap.String("iana", iana),
		zap.Int("offset_seconds", offset))
	return iana, offset, nil
}

// ResyncDeviceReportTime 把 tenant 配的 sleepreport_time (unit local hour 1-24) 下发到 sleepace 设备。
//
// timezone 设计：reportUploadTime 是 device 本地时钟触发（设备 RTC 跟用户走，DST 由 device 端自动跟随），
// server 直接传 unit local hour 整数即可，无须 server↔device 时区换算。Sleepace 厂家也不保存 timezone
// （见 sleepace_gateway_client.go IANAToOffsetSeconds 注释）。
//
// 调用路径：POST /internal/sleepace/device/{device_addr}/resync-report-time
// 外部 DST 脚本不再需要（device 端自治）；保留 endpoint 仅作"手动 resync" / UI trigger 用。
func (s *deviceMonitorSettingsService) ResyncDeviceReportTime(ctx context.Context, tenantID, deviceAddr string) (int, error) {
	if s.sleepaceGateway == nil {
		return 0, fmt.Errorf("sleepace gateway not wired")
	}
	if tenantID == "" || deviceAddr == "" {
		return 0, fmt.Errorf("tenant_id and device_addr are required")
	}

	hour := alarm.DefaultSleepReportTime
	if s.alarmCloudRepo != nil {
		if ac, err := s.alarmCloudRepo.GetAlarmCloud(ctx, tenantID); err == nil && ac != nil && len(ac.Metadata) > 0 {
			var tr alarm.TenantResetTime
			if json.Unmarshal(ac.Metadata, &tr) == nil && tr.TenantSleepReportTime >= 1 && tr.TenantSleepReportTime <= 24 {
				hour = tr.TenantSleepReportTime
			}
		}
	}

	deviceCode, err := s.resolveDeviceCode(ctx, deviceAddr)
	if err != nil {
		return 0, fmt.Errorf("resolve device_code: %w", err)
	}
	deviceUID, err := s.resolveDeviceUID(ctx, deviceAddr)
	if err != nil {
		return 0, fmt.Errorf("resolve device_uid: %w", err)
	}

	if err := s.sleepaceGateway.SetReportUploadTime(ctx, deviceUID, deviceCode, hour); err != nil {
		return 0, fmt.Errorf("sleepace SetReportUploadTime: %w", err)
	}
	s.logger.Info("ResyncDeviceReportTime done",
		zap.String("tenant_id", tenantID),
		zap.String("device_addr", deviceAddr),
		zap.Int("report_upload_time_hour", hour))
	return hour, nil
}

func (s *deviceMonitorSettingsService) TriggerSleepaceUpgrade(ctx context.Context, tenantID, deviceAddr, version string) error {
	return errNotImplemented("TriggerSleepaceUpgrade")
}

func (s *deviceMonitorSettingsService) UploadSleepaceFirmware(ctx context.Context, filename string, file io.Reader) error {
	return errNotImplemented("UploadSleepaceFirmware")
}

func (s *deviceMonitorSettingsService) DeleteSleepaceFirmware(ctx context.Context, deviceType int, deviceVersion string) error {
	return errNotImplemented("DeleteSleepaceFirmware")
}

func (s *deviceMonitorSettingsService) ListSleepaceFirmwareVersions(ctx context.Context) ([]SleepaceFirmwareVersion, error) {
	return nil, errNotImplemented("ListSleepaceFirmwareVersions")
}

func (s *deviceMonitorSettingsService) LocalUploadSleepaceFirmware(ctx context.Context, filename string, file io.Reader, deviceType, deviceVersion string) (bool, error) {
	return false, errNotImplemented("LocalUploadSleepaceFirmware")
}

func (s *deviceMonitorSettingsService) LocalDeleteSleepaceFirmware(ctx context.Context, filename string) error {
	return errNotImplemented("LocalDeleteSleepaceFirmware")
}

func (s *deviceMonitorSettingsService) ResolveSleepaceUpgradeVersion(filename string) (string, error) {
	return "", errNotImplemented("ResolveSleepaceUpgradeVersion")
}
