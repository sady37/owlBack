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
	"wisefido-data/internal/publisher"
	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

// DeviceStoreService 设备库存业务：Sleepace 绑定/解绑、InitialAll 查询回写等。
// Handler 只做请求解析与响应写入，业务逻辑在此层。
type DeviceStoreService struct {
	db              *sql.DB
	deviceStoreRepo repository.DeviceStoreRepository
	devicesRepo     repository.DevicesRepository
	unitsRepo       repository.UnitsRepository
	sleepaceGateway *SleepaceGatewayClient
	configPublisher *publisher.ConfigPublisher
	// onSleepadVendorBound — sleepad 在厂家成功 bind+InitializeDevice 之后调（applyDefaultSleepadRealtime success）。
	// 用于通知 SleepaceIntervalScheduler 立即清掉 1h vendor-unbound backoff cache，
	// 让重新 bind 后下次 60s tick 就能 push interval（不等 1h TTL）。可空（未注入即 noop）。
	onSleepadVendorBound func(deviceID string)
	logger               *zap.Logger
}

// NewDeviceStoreService 创建设备库存 Service。sleepaceGateway 可为 nil（未配置时不执行绑定相关逻辑）。
// db 用于 bind-init 时读 spatial_config 的 user-configured realtime_interval；nil 时落 alarm.go 默认。
func NewDeviceStoreService(
	db *sql.DB,
	deviceStoreRepo repository.DeviceStoreRepository,
	devicesRepo repository.DevicesRepository,
	unitsRepo repository.UnitsRepository,
	sleepaceGateway *SleepaceGatewayClient,
	logger *zap.Logger,
) *DeviceStoreService {
	return &DeviceStoreService{
		db:              db,
		deviceStoreRepo: deviceStoreRepo,
		devicesRepo:     devicesRepo,
		unitsRepo:       unitsRepo,
		sleepaceGateway: sleepaceGateway,
		logger:          logger,
	}
}

func (s *DeviceStoreService) SetConfigPublisher(pub *publisher.ConfigPublisher) {
	s.configPublisher = pub
}

// SetOnSleepadVendorBound — 注入 sleepad 厂家 bind 成功回调。
// main.go wire scheduler 时调一次：deviceStoreService.SetOnSleepadVendorBound(intervalScheduler.InvalidateUnbound)
func (s *DeviceStoreService) SetOnSleepadVendorBound(cb func(deviceID string)) {
	s.onSleepadVendorBound = cb
}

// BatchUpdateDeviceStoresNotify 批量更新 device_store 成功后发 config.card，并按 tenant 边界转移触发 Sleepace bind/unbind。
//
// Sleepace 厂家 bind 规则（业务约定，与 v2 pivot tenants 配合）：
//   - 转入 **真实 tenant**（非 System fd00:0:1::/48、非 Trash fd00:0:2::/48）→ 触发 BindSleepadOne
//   - 转入 **Trash** (fd00:0:2::/48) → 触发 UnbindSleepadOne
//   - 转入 System 或 tenant 内部移位（pool ↔ unit/bed）→ noop（Sleepace 端不动）
//
// 调用方按 update.TenantID 设置目标 /48；本函数 update 前抓 old tenant /48 做差分。
func (s *DeviceStoreService) BatchUpdateDeviceStoresNotify(ctx context.Context, updates []*domain.DeviceStore) error {
	// update 前抓每个 device 的旧 tenant /48（仅对 Sleepad 关心）。
	type transition struct {
		oldTenant string
		newTenant string
	}
	transitions := make(map[string]transition)
	for _, u := range updates {
		if u == nil || u.DeviceUID == "" || u.TenantID == "" {
			continue
		}
		old := s.lookupCurrentTenantPrefix(ctx, u.DeviceUID)
		if old == u.TenantID {
			continue
		}
		transitions[u.DeviceUID] = transition{oldTenant: old, newTenant: u.TenantID}
	}

	if err := s.deviceStoreRepo.BatchUpdateDeviceStores(ctx, updates); err != nil {
		return err
	}
	NotifyDeviceStoreBatchAfterUpdate(ctx, s.deviceStoreRepo, s.configPublisher, updates, s.logger)

	// 异步触发 Sleepace bind/unbind（不阻塞调用方；失败 Debug log 不报错给用户）。
	for uid, tr := range transitions {
		uidCopy, oldCopy, newCopy := uid, tr.oldTenant, tr.newTenant
		go s.applyTenantTransitionToSleepace(context.Background(), uidCopy, oldCopy, newCopy)
	}
	return nil
}

// lookupCurrentTenantPrefix 取 device 当前 device_addr 的 /48 host repr+"/48"；查不到（factory-only）返 ""。
func (s *DeviceStoreService) lookupCurrentTenantPrefix(ctx context.Context, deviceUID string) string {
	if s.db == nil || deviceUID == "" {
		return ""
	}
	var prefix sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT host(network(set_masklen(device_addr, 48))) || '/48'
		  FROM devices
		 WHERE device_uid = $1
	`, deviceUID).Scan(&prefix)
	if err != nil || !prefix.Valid {
		return ""
	}
	return prefix.String
}

// applyTenantTransitionToSleepace 根据 tenant /48 转移方向决定 Sleepace bind / unbind / noop。
// 调用方需在 repo update 成功后再调，确保 device_addr 已是 newTenant。
func (s *DeviceStoreService) applyTenantTransitionToSleepace(ctx context.Context, deviceUID, oldTenant, newTenant string) {
	if s.sleepaceGateway == nil {
		return
	}
	const trashTenant = "fd00:0:2::/48"
	newIsTrash := newTenant == trashTenant
	newIsReal := newTenant != "" && !newIsTrash && !isSystemOrTrashTenant(newTenant)
	oldIsReal := oldTenant != "" && !isSystemOrTrashTenant(oldTenant)

	switch {
	case newIsReal && !oldIsReal:
		// 转入真实 tenant（从 factory-only / system / trash 进来）→ bind
		if err := s.BindSleepadOne(ctx, deviceUID); err != nil {
			s.logger.Debug("tenant-transition: BindSleepadOne skipped/failed",
				zap.String("device_uid", deviceUID),
				zap.String("from", oldTenant), zap.String("to", newTenant),
				zap.Error(err))
		}
	case newIsTrash && oldIsReal:
		// 真实 tenant → trash → unbind
		if err := s.UnbindSleepadOne(ctx, deviceUID); err != nil {
			s.logger.Debug("tenant-transition: UnbindSleepadOne skipped/failed",
				zap.String("device_uid", deviceUID),
				zap.String("from", oldTenant), zap.String("to", newTenant),
				zap.Error(err))
		}
	}
	// 其他组合（真实↔真实需经 pivot；system 中转；tenant 内部移位）都不动 Sleepace。
}

// isSystemOrTrashTenant 判 pivot tenant（系统保留 slot 1/2，业务不在此 bind/unbind）。
func isSystemOrTrashTenant(prefix string) bool {
	p := strings.TrimSpace(prefix)
	return p == "fd00:0:1::/48" || p == "fd00:0:2::/48"
}

// ImportDeviceStoresNotify 导入成功后按插入行发 config.card；若目标 tenant 是真实 tenant 顺手触发 Sleepace bind。
// 导入路径默认 tenant = System (fd00:0:1::/48)，此时 newIsReal=false，bind 不会触发——admin 后续调拨到真实 tenant 时
// 由 BatchUpdateDeviceStoresNotify 的 tenant-transition 逻辑负责 bind。
func (s *DeviceStoreService) ImportDeviceStoresNotify(ctx context.Context, items []*domain.DeviceStore) (successCount int, inserted []*domain.DeviceStore, skipped []*domain.DeviceStore, errors []*domain.DeviceStore, err error) {
	successCount, inserted, skipped, errors, err = s.deviceStoreRepo.ImportDeviceStores(ctx, items)
	if err != nil {
		return
	}
	NotifyDeviceStoreFromStores(ctx, s.configPublisher, inserted, "device_store_imported", s.logger)
	// 直接导入到真实 tenant 的 Sleepad → bind（factory-only 是没有 oldTenant 的转入）
	for _, ds := range inserted {
		if ds == nil || ds.DeviceUID == "" {
			continue
		}
		newTenant := strings.TrimSpace(ds.TenantID)
		if newTenant == "" || isSystemOrTrashTenant(newTenant) {
			continue
		}
		uidCopy, newCopy := ds.DeviceUID, newTenant
		go s.applyTenantTransitionToSleepace(context.Background(), uidCopy, "", newCopy)
	}
	return
}

func (s *DeviceStoreService) getTimezoneForDevice(ctx context.Context, tenantID, deviceID string) int {
	if s.devicesRepo == nil || s.unitsRepo == nil || tenantID == "" || deviceID == "" {
		return DefaultTimezoneOffsetSeconds
	}
	dev, err := s.devicesRepo.GetDevice(ctx, tenantID, deviceID)
	if err != nil || dev == nil {
		return DefaultTimezoneOffsetSeconds
	}
	if !dev.UnitID.Valid || dev.UnitID.String == "" {
		return DefaultTimezoneOffsetSeconds
	}
	u, err := s.unitsRepo.GetUnit(ctx, tenantID, dev.UnitID.String)
	if err != nil || u == nil || u.Timezone == "" {
		return DefaultTimezoneOffsetSeconds
	}
	return IANAToOffsetSeconds(u.Timezone)
}

// SyncSleepaceBindResult 同步 Sleepace 绑定结果
type SyncSleepaceBindResult struct {
	Synced  int
	Skipped int
	Failed  int
	Errors  []string
}

// SyncSleepaceBind 先查 bindInfo(userId=device_uid)，仅对未绑定的 Sleepad 执行绑定。
func (s *DeviceStoreService) SyncSleepaceBind(ctx context.Context) (*SyncSleepaceBindResult, error) {
	if s.sleepaceGateway == nil {
		return nil, fmt.Errorf("sleepace gateway not configured")
	}
	filters := repository.DeviceStoreFilters{DeviceType: "Sleepad"}
	items, _, err := s.deviceStoreRepo.ListDeviceStores(ctx, filters, 1, 1000, "", "asc")
	if err != nil {
		s.logger.Error("SyncSleepaceBind list failed", zap.Error(err))
		return nil, fmt.Errorf("list device_store: %w", err)
	}
	var synced, skipped, failed int
	var errMsgs []string
	for _, row := range items {
		if row.DeviceUID == "" {
			continue
		}
		status, _, bindItems, getErr := s.sleepaceGateway.GetBindInfo(ctx, row.DeviceUID)
		if getErr != nil {
			failed++
			msg := getErr.Error()
			if msg == "" {
				msg = "getBindInfo failed"
			}
			errMsgs = append(errMsgs, row.DeviceUID+": getBindInfo "+msg)
			continue
		}
		if status == 0 && len(bindItems) > 0 {
			skipped++
			continue
		}
		tz := s.getTimezoneForDevice(ctx, row.TenantID, row.DeviceUID)
		if !row.DeviceCode.Valid || row.DeviceCode.String == "" {
			failed++
			errMsgs = append(errMsgs, row.DeviceUID+": device_code required for bind (from import)")
			continue
		}
		_, initErr := s.sleepaceGateway.InitializeDevice(ctx, row.DeviceCode.String, row.DeviceUID, &tz)
		if initErr != nil {
			failed++
			msg := initErr.Error()
			if msg == "" {
				msg = "initialize failed (check wisefido-sleepace / sleepace-service logs)"
			}
			errMsgs = append(errMsgs, row.DeviceUID+": "+msg)
			continue
		}
		s.applyDefaultSleepadRealtime(ctx, row)
		synced++
	}
	return &SyncSleepaceBindResult{Synced: synced, Skipped: skipped, Failed: failed, Errors: errMsgs}, nil
}

// InitialAllSleepadResult InitialAll 结果：1) 查询回写 2) 未绑定单独绑定
type InitialAllSleepadResult struct {
	Synced           int
	Failed           int
	NoData           int
	NoDataList       []string
	NoDataBound      int
	NoDataBindFailed int
	Errors           []string
	SuccessDetails   []map[string]any
}

// InitialAllSleepad 1) 查询、回写：查所有 Sleepad 的 bindInfo，有数据则把 deviceVersion 回写 firmware_version；
// 2) 把未绑定单独绑定：对 bindInfo 无数据的设备调用 initialize 绑定。
func (s *DeviceStoreService) InitialAllSleepad(ctx context.Context) (*InitialAllSleepadResult, error) {
	if s.sleepaceGateway == nil {
		return nil, fmt.Errorf("sleepace gateway not configured")
	}
	filters := repository.DeviceStoreFilters{DeviceType: "Sleepad"}
	items, _, err := s.deviceStoreRepo.ListDeviceStores(ctx, filters, 1, 1000, "", "asc")
	if err != nil {
		s.logger.Error("InitialAllSleepad list failed", zap.Error(err))
		return nil, fmt.Errorf("list: %w", err)
	}
	var synced, failed, noData int
	var noDataBound, noDataBindFailed int
	var successDetails []map[string]any
	var errMsgs []string
	var noDataList []string
	for _, row := range items {
		if row.DeviceUID == "" {
			continue
		}
		status, msg, bindItems, getErr := s.sleepaceGateway.GetBindInfo(ctx, row.DeviceUID)
		if getErr != nil {
			failed++
			errMsgs = append(errMsgs, row.DeviceUID+": "+getErr.Error())
			continue
		}
		if status != 0 {
			failed++
			if msg == "" {
				msg = "bindInfo status non-zero"
			}
			errMsgs = append(errMsgs, row.DeviceUID+": "+msg)
			continue
		}
		if len(bindItems) == 0 {
			noData++
			noDataList = append(noDataList, row.DeviceUID)
			tz := s.getTimezoneForDevice(ctx, row.TenantID, row.DeviceUID)
			if !row.DeviceCode.Valid || row.DeviceCode.String == "" {
				noDataBindFailed++
				errMsgs = append(errMsgs, row.DeviceUID+": device_code required for bind (from import)")
				continue
			}
			_, initErr := s.sleepaceGateway.InitializeDevice(ctx, row.DeviceCode.String, row.DeviceUID, &tz)
			if initErr != nil {
				noDataBindFailed++
				initMsg := initErr.Error()
				if initMsg == "" {
					initMsg = "initialize failed"
				}
				errMsgs = append(errMsgs, row.DeviceUID+": bind "+initMsg)
			} else {
				s.applyDefaultSleepadRealtime(ctx, row)
				noDataBound++
			}
			continue
		}
		first := bindItems[0]
		fwVer := fmt.Sprintf("%.2f", first.DeviceVersion)
		if err := s.deviceStoreRepo.UpdateFirmwareVersion(ctx, row.DeviceUID, fwVer); err != nil {
			failed++
			errMsgs = append(errMsgs, row.DeviceUID+": update firmware "+err.Error())
			continue
		}
		if s.configPublisher != nil && row.DeviceUID != "" {
			_ = s.configPublisher.PublishConfigChanged(ctx, "update", nil, nil, []string{row.DeviceUID})
		}
		synced++
		successDetails = append(successDetails, map[string]any{
			"device_uid": row.DeviceUID, "device_type": first.DeviceType, "firmware_version": fwVer,
		})
	}
	return &InitialAllSleepadResult{
		Synced:           synced,
		Failed:           failed,
		NoData:           noData,
		NoDataList:       noDataList,
		NoDataBound:      noDataBound,
		NoDataBindFailed: noDataBindFailed,
		Errors:           errMsgs,
		SuccessDetails:   successDetails,
	}, nil
}

// BindSleepadOne 单条 Sleepad 绑定：调用 initialize(device_code, device_uid)。
// 厂家映射：Sleepace device_id = device_factory_meta.device_code，Sleepace userid = device_factory_meta.device_uid。
// 当前仅绑定在 left（leftRight=0）。
func (s *DeviceStoreService) BindSleepadOne(ctx context.Context, deviceUID string) error {
	if s.sleepaceGateway == nil {
		return fmt.Errorf("sleepace gateway not configured")
	}
	ds, err := s.deviceStoreRepo.GetDeviceStore(ctx, deviceUID)
	if err != nil || ds == nil {
		return fmt.Errorf("device not found")
	}
	if !strings.EqualFold(ds.DeviceType, "Sleepad") {
		return fmt.Errorf("not a Sleepad device")
	}
	if !ds.DeviceCode.Valid || ds.DeviceCode.String == "" {
		return fmt.Errorf("device_code required for bind (from import)")
	}
	tz := s.getTimezoneForDevice(ctx, ds.TenantID, ds.DeviceUID)
	_, initErr := s.sleepaceGateway.InitializeDevice(ctx, ds.DeviceCode.String, ds.DeviceUID, &tz)
	if initErr != nil {
		return initErr
	}
	s.applyDefaultSleepadRealtime(ctx, ds)
	if s.configPublisher != nil && ds.DeviceUID != "" {
		_ = s.configPublisher.PublishConfigChanged(ctx, "update", nil, nil, []string{ds.DeviceUID})
	}
	return nil
}

// UnbindSleepadOne 单条 Sleepad 解绑：仅调用 Sleepace unbind(device_code)。不清空 device_code。
func (s *DeviceStoreService) UnbindSleepadOne(ctx context.Context, deviceUID string) error {
	if s.sleepaceGateway == nil {
		return fmt.Errorf("sleepace gateway not configured")
	}
	ds, err := s.deviceStoreRepo.GetDeviceStore(ctx, deviceUID)
	if err != nil || ds == nil {
		return fmt.Errorf("device not found")
	}
	if !strings.EqualFold(ds.DeviceType, "Sleepad") {
		return fmt.Errorf("not a Sleepad device")
	}
	if !ds.DeviceCode.Valid || ds.DeviceCode.String == "" {
		return fmt.Errorf("device has no device_code, already unbound")
	}
	if err := s.sleepaceGateway.UnbindDevice(ctx, ds.DeviceCode.String); err != nil {
		return err
	}
	if s.configPublisher != nil && ds.DeviceUID != "" {
		_ = s.configPublisher.PublishConfigChanged(ctx, "update", nil, nil, []string{ds.DeviceUID})
	}
	return nil
}

// DeleteDeviceStoreAndNotify 删除 device_store 行并发送 config.changed。
func (s *DeviceStoreService) DeleteDeviceStoreAndNotify(ctx context.Context, deviceUID string) error {
	if deviceUID == "" {
		return fmt.Errorf("device_uid required")
	}
	row, err := s.deviceStoreRepo.GetDeviceStore(ctx, deviceUID)
	if err != nil {
		return err
	}
	if row == nil {
		return fmt.Errorf("device not found")
	}
	if err := s.deviceStoreRepo.DeleteDeviceStore(ctx, deviceUID); err != nil {
		return err
	}
	if s.configPublisher != nil && row.DeviceUID != "" {
		_ = s.configPublisher.PublishConfigChanged(ctx, "update", nil, nil, []string{row.DeviceUID})
	}
	return nil
}

// applyDefaultSleepadRealtime 在 Sleepace bind 成功后下发实时上报配置：
//   - SetRealtimeInterval —— 取值优先：spatial_config (FE 保存值) > alarm.go SleepadSetting 默认 (10)
//   - SetRealtimeModeAfterLeave(1) —— 离床后停止上报；仅 BM8701-2 + 固件 ≥ 6.67 支持
//
// 任何错误降级为 Warn 日志，不阻塞 bind 流程。偶发瞬时错误（实测厂家 status:3 空 msg）做一次重试。
// 这是 bind 后的"最后一公里"配置，失败也不影响绑定成功；后续可由 cmd/sleepace-set-realtime -apply-all 兜底。
// 运行期 SleepaceIntervalScheduler 在 reset_time + post-meal(12:00-14:00) 窗口内会切 2s 高频，离开窗口切回 user/default 值。
func (s *DeviceStoreService) applyDefaultSleepadRealtime(ctx context.Context, ds *domain.DeviceStore) {
	if s.sleepaceGateway == nil || ds == nil {
		return
	}
	if !ds.DeviceCode.Valid || ds.DeviceCode.String == "" {
		return
	}
	targetInterval := s.resolveSleepadRealtimeInterval(ctx, ds.DeviceUID)
	const targetMode = 1

	if err := setIntervalWithRetry(ctx, s.sleepaceGateway, ds.DeviceUID, ds.DeviceCode.String, targetInterval); err != nil {
		s.logger.Warn("post-bind SetRealtimeInterval failed",
			zap.String("device_uid", ds.DeviceUID),
			zap.String("device_id", ds.DeviceUID),
			zap.Error(err))
	} else if s.onSleepadVendorBound != nil {
		// bind+init 成功 → 通知 scheduler 清 unbound backoff，下次 tick 立即重新评估 push（不等 1h TTL）
		s.onSleepadVendorBound(ds.DeviceUID)
	}

	if !sleepadSupportsRealtimeMode(ds) {
		return
	}
	if err := setRealtimeModeWithRetry(ctx, s.sleepaceGateway, ds.DeviceUID, ds.DeviceCode.String, targetMode); err != nil {
		s.logger.Warn("post-bind SetRealtimeModeAfterLeave failed",
			zap.String("device_uid", ds.DeviceUID),
			zap.String("device_id", ds.DeviceUID),
			zap.Error(err))
	}
}

// resolveSleepadRealtimeInterval 取 bind-init 时 cloud 应下发的 realtime_interval：
//   1. spatial_config alarm.device_config (FE 保存值)  ←—— 优先
//   2. alarm.go SleepadSetting 默认 10                  ←—— 缺失/解析失败时兜底
//
// scheduler 运行期会在窗口内切 2s 高频；此函数只决定 bind-init 的初始值。
func (s *DeviceStoreService) resolveSleepadRealtimeInterval(ctx context.Context, deviceUID string) int {
	fallback := sleepadDefaultRealtimeInterval()
	if s.db == nil || deviceUID == "" {
		return fallback
	}
	var raw []byte
	err := s.db.QueryRowContext(ctx, `
		SELECT sc.config_value
		  FROM spatial_config sc
		  JOIN devices d ON d.device_addr = sc.spatial_prefix
		 WHERE d.device_uid = $1
		   AND sc.config_key = 'alarm.device_config'
	`, deviceUID).Scan(&raw)
	if err != nil || len(raw) == 0 {
		return fallback
	}
	var packed struct {
		AlarmItems []alarm.AlarmItem `json:"alarm_items"`
	}
	if json.Unmarshal(raw, &packed) != nil {
		return fallback
	}
	for _, item := range packed.AlarmItems {
		if item.AlarmType != alarm.SleepadSetting || item.AlarmParams == nil {
			continue
		}
		switch v := item.AlarmParams["realtime_interval"].(type) {
		case float64:
			if int(v) > 0 {
				return int(v)
			}
		case int:
			if v > 0 {
				return v
			}
		case json.Number:
			if n, err := v.Int64(); err == nil && n > 0 {
				return int(n)
			}
		}
		return fallback
	}
	return fallback
}

// sleepadDefaultRealtimeInterval 从 alarm.go 默认 SleepadSetting 提 realtime_interval（10）。
// 单一来源：始终跟着 alarm.go 变，不在此处硬编码 10。
func sleepadDefaultRealtimeInterval() int {
	for _, item := range alarm.GetDefaultAlarmItemsSleepPad() {
		if item.AlarmType != alarm.SleepadSetting || item.AlarmParams == nil {
			continue
		}
		switch v := item.AlarmParams["realtime_interval"].(type) {
		case float64:
			return int(v)
		case int:
			return v
		}
	}
	return 10
}

func setIntervalWithRetry(ctx context.Context, gw *SleepaceGatewayClient, deviceID, deviceCode string, interval int) error {
	if err := gw.SetRealtimeInterval(ctx, deviceID, deviceCode, interval); err == nil {
		return nil
	}
	time.Sleep(2 * time.Second)
	return gw.SetRealtimeInterval(ctx, deviceID, deviceCode, interval)
}

func setRealtimeModeWithRetry(ctx context.Context, gw *SleepaceGatewayClient, deviceID, deviceCode string, mode int) error {
	if err := gw.SetRealtimeModeAfterLeave(ctx, deviceID, deviceCode, mode); err == nil {
		return nil
	}
	time.Sleep(2 * time.Second)
	return gw.SetRealtimeModeAfterLeave(ctx, deviceID, deviceCode, mode)
}

// sleepadSupportsRealtimeMode 仅 BM8701-2 + 固件 ≥ 6.67。fw "0.00" / 空 / 非数字均视为不支持（保守）。
func sleepadSupportsRealtimeMode(ds *domain.DeviceStore) bool {
	if !ds.DeviceModel.Valid || !strings.EqualFold(ds.DeviceModel.String, "BM8701-2") {
		return false
	}
	if !ds.FirmwareVersion.Valid {
		return false
	}
	return parseSleepadFirmwareScore(ds.FirmwareVersion.String) >= 667
}

// parseSleepadFirmwareScore "6.89" → 689；"6.67" → 667；"0.00" → 0；非数字/空 → -1。
func parseSleepadFirmwareScore(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return -1
	}
	parts := strings.SplitN(s, ".", 2)
	major, ok := atoiAllDigits(parts[0])
	if !ok {
		return -1
	}
	minor := 0
	if len(parts) == 2 {
		var ok2 bool
		minor, ok2 = atoiAllDigits(parts[1])
		if !ok2 {
			return -1
		}
	}
	return major*100 + minor
}

func atoiAllDigits(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int(c-'0')
	}
	return n, true
}

