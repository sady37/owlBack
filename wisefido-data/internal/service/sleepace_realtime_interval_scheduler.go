package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"owl-common/alarm"
	"wisefido-data/internal/repository"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// SleepaceIntervalScheduler — 按时段切换 sleepad realtime_interval。
//
// 默认策略：
//   - rest 窗口内（默认 21:30~07:30）= 2s 高频，监测 sleep stages
//   - nap 窗口内（默认 12:00~14:00）= 2s 高频
//   - 其它时段 = 10s（省电、降低 broker 流量）
//
// 提前/延后 padding：
//   - 进入窗口**前 10 分钟**开始 2s（让设备进入 sleep period 前已经在采集）
//   - 离开窗口**后 2 分钟**切回 10s（避免临界跳变 + 给睡眠分析留余量）
//
// 用户尊重：device monitor settings 里 SleepadSetting.realtime_interval >= 10
// 视为用户明确选择低频，scheduler 跳过该 device（不强制切换）。
//
// 触发：每 60s tick；去重：Redis hash 缓存上次下发值，相同则跳过 set 调用。
type SleepaceIntervalScheduler struct {
	db              *sql.DB
	redisClient     *redis.Client
	sleepaceGateway *SleepaceGatewayClient
	alarmCloudRepo  repository.AlarmCloudRepository
	logger          *zap.Logger
}

const (
	intervalScheduledKey      = "sleepace:realtime_interval" // Redis hash: device_id → "2"|"10"
	intervalUnboundKey        = "sleepace:scheduler:unbound" // Redis hash: device_id → "1"，命中跳过；TTL 通过 set-with-ttl 模拟
	intervalUnboundTTL        = 1 * time.Hour                // unbound device backoff 时间，避免每分钟 spam log
	intervalEnterPadMin       = 10                           // 进入窗口前提早多少分钟切到高频
	intervalLeavePadMin       = 2                            // 离开窗口后延迟多少分钟切回低频
	intervalHighFreq          = 2                            // 窗口内 interval (秒)
	intervalNormal            = 10                           // 窗口外 interval (秒)
	intervalUserSkipThreshold = 10                           // 用户 device-settings interval >= 此值 → 跳过 scheduler 干预
)

func NewSleepaceIntervalScheduler(
	db *sql.DB,
	redisClient *redis.Client,
	sleepaceGateway *SleepaceGatewayClient,
	alarmCloudRepo repository.AlarmCloudRepository,
	logger *zap.Logger,
) *SleepaceIntervalScheduler {
	return &SleepaceIntervalScheduler{
		db:              db,
		redisClient:     redisClient,
		sleepaceGateway: sleepaceGateway,
		alarmCloudRepo:  alarmCloudRepo,
		logger:          logger,
	}
}

// Start 启动后台 ticker。ctx 取消即退出。
// 启动后立即跑一次（让重启后状态尽快收敛），之后每 60 秒一次。
func (s *SleepaceIntervalScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	s.logger.Info("[INTERVAL_SCHEDULER] started",
		zap.Int("enter_pad_min", intervalEnterPadMin),
		zap.Int("leave_pad_min", intervalLeavePadMin),
		zap.Int("high_freq_s", intervalHighFreq),
		zap.Int("normal_s", intervalNormal),
		zap.Int("user_skip_threshold", intervalUserSkipThreshold),
	)
	s.tick(ctx)
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("[INTERVAL_SCHEDULER] stopped")
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

type sleepadDeviceRow struct {
	deviceID   string
	deviceUID  string
	deviceCode string
	tenantID   string // /48 CIDR
	unitID     string // /80 CIDR
	timezone   string // unit IANA timezone（顺便取，省一次 query）
}

func (s *SleepaceIntervalScheduler) tick(ctx context.Context) {
	devices, err := s.listSleepadDevices(ctx)
	if err != nil {
		s.logger.Warn("[INTERVAL_SCHEDULER] list sleepad devices failed", zap.Error(err))
		return
	}
	if len(devices) == 0 {
		return
	}
	for _, d := range devices {
		s.checkDevice(ctx, d)
	}
}

// listSleepadDevices — 只取**真正 bound 到 unit** 的 Sleepad device（INNER JOIN units）。
// 未 bind 的库存 device（device_ipv6 不在任何 unit /80 内）跳过 — 否则厂家 API 会报
// "user not found" (status 5)，每分钟刷一遍 log。同时一次 query 拿到 unit timezone，省后续 lookup。
func (s *SleepaceIntervalScheduler) listSleepadDevices(ctx context.Context) ([]sleepadDeviceRow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT dfm.device_id::text, dfm.device_uid, COALESCE(dfm.device_code, ''),
		       host(network(set_masklen(d.device_ipv6, 48))) || '/48' AS tenant_id,
		       host(network(set_masklen(d.device_ipv6, 80))) || '/80' AS unit_id,
		       COALESCE(u.timezone, '') AS tz
		  FROM device_factory_meta dfm
		  JOIN devices d ON d.device_id = dfm.device_id
		  JOIN units u   ON u.unit_id   = network(set_masklen(d.device_ipv6, 80))
		 WHERE dfm.device_type = 'Sleepad'
		   AND COALESCE(dfm.device_code, '') <> ''
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []sleepadDeviceRow
	for rows.Next() {
		var r sleepadDeviceRow
		if err := rows.Scan(&r.deviceID, &r.deviceUID, &r.deviceCode, &r.tenantID, &r.unitID, &r.timezone); err != nil {
			continue
		}
		out = append(out, r)
	}
	return out, nil
}

func (s *SleepaceIntervalScheduler) checkDevice(ctx context.Context, d sleepadDeviceRow) {
	// 1. unbound backoff cache：sleepace 厂家返回过 "user not found" 的 device 在 TTL 内跳过，
	//    避免每 60s spam log。bind 路径需主动 InvalidateUnbound 清缓存让重新 bind 后立即生效。
	if exists, _ := s.redisClient.Exists(ctx, intervalUnboundKey+":"+d.deviceID).Result(); exists > 0 {
		return
	}

	// 2. 用户配置 skip：device monitor settings 设了 interval >= 10 → 不干预
	userInterval := s.readUserConfiguredInterval(ctx, d.deviceID)
	if userInterval >= intervalUserSkipThreshold {
		return
	}

	// 3. 决定目标 interval（2 高频 / 10 低频）
	// IANA tz（如 "America/Denver"）走 time.LoadLocation 自动按当年 DST 规则解析；
	// now.In(loc).Hour()/Minute() 已经是 wall-clock 本地时间（夏令时 / 标准时段都正确）。
	tz := d.timezone
	if tz == "" {
		tz = "America/Denver"
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc, _ = time.LoadLocation("America/Denver")
	}
	rtParams := s.resolveTenantResetTime(ctx, d.tenantID)
	target := intervalNormal
	if isInSleepWindow(time.Now().In(loc), rtParams) {
		target = intervalHighFreq
	}

	// 4. 去重：上次下发已经是目标值就不重复推
	targetStr := fmt.Sprintf("%d", target)
	last, _ := s.redisClient.HGet(ctx, intervalScheduledKey, d.deviceID).Result()
	if last == targetStr {
		return
	}

	// 5. 下发到厂家
	if s.sleepaceGateway == nil {
		return
	}
	if err := s.sleepaceGateway.SetRealtimeInterval(ctx, d.deviceID, d.deviceCode, target); err != nil {
		// "user not found" = sleepace 厂家未 BindDevice/InitializeDevice。这种 device 持续 retry 没意义，
		// 加 1h backoff 缓存（只 log 一次）。bind 流程成功时会调 InvalidateUnbound 清此 key。
		if isSleepaceUserNotFound(err) {
			_ = s.redisClient.Set(ctx, intervalUnboundKey+":"+d.deviceID, "1", intervalUnboundTTL).Err()
			// per-device 1h backoff，bind 流程 InvalidateUnbound 时才会重新尝试 — Debug 即可
			s.logger.Debug("[INTERVAL_SCHEDULER] device not bound at sleepace vendor, backoff",
				zap.String("device_id", d.deviceID),
				zap.String("device_uid", d.deviceUID),
				zap.Duration("ttl", intervalUnboundTTL))
			return
		}
		s.logger.Warn("[INTERVAL_SCHEDULER] SetRealtimeInterval failed",
			zap.String("device_id", d.deviceID),
			zap.String("device_uid", d.deviceUID),
			zap.Int("target", target),
			zap.Error(err))
		return
	}
	if err := s.redisClient.HSet(ctx, intervalScheduledKey, d.deviceID, targetStr).Err(); err != nil {
		s.logger.Debug("[INTERVAL_SCHEDULER] cache set failed", zap.Error(err))
	}
	s.logger.Info("[INTERVAL_SCHEDULER] interval pushed",
		zap.String("device_id", d.deviceID),
		zap.String("device_uid", d.deviceUID),
		zap.Int("interval_s", target),
		zap.String("prev", last),
		zap.String("tz", tz),
	)
}

// InvalidateUnbound — bind/InitializeDevice 成功路径调用，清掉 backoff cache 让 scheduler 立即重新 push interval。
// 在 device_store_service.applyDefaultSleepadRealtime 之后调；可选 wire（不调也只是延迟到下次 1h backoff TTL 过期）。
func (s *SleepaceIntervalScheduler) InvalidateUnbound(ctx context.Context, deviceID string) {
	if deviceID == "" {
		return
	}
	_ = s.redisClient.Del(ctx, intervalUnboundKey+":"+deviceID).Err()
}

// isSleepaceUserNotFound 厂家 "user not found (status: 5)" 错误判定。字符串匹配避免依赖 error type。
func isSleepaceUserNotFound(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "user not found") || strings.Contains(s, "status: 5")
}

// readUserConfiguredInterval — 从 spatial_config alarm.device_config 读 SleepadSetting.realtime_interval。
// 找不到 / 解析失败 → 返回 0（视为未设置，不触发 skip）。
func (s *SleepaceIntervalScheduler) readUserConfiguredInterval(ctx context.Context, deviceID string) int {
	var raw []byte
	err := s.db.QueryRowContext(ctx, `
		SELECT sc.config_value
		  FROM spatial_config sc, devices d
		 WHERE d.device_id = $1::uuid
		   AND sc.spatial_prefix = d.device_ipv6
		   AND sc.config_key = 'alarm.device_config'
	`, deviceID).Scan(&raw)
	if err != nil || len(raw) == 0 {
		return 0
	}
	var packed struct {
		AlarmItems []alarm.AlarmItem `json:"alarm_items"`
	}
	if json.Unmarshal(raw, &packed) != nil {
		return 0
	}
	for _, item := range packed.AlarmItems {
		if item.AlarmType != alarm.SleepadSetting || item.AlarmParams == nil {
			continue
		}
		if v, ok := item.AlarmParams["realtime_interval"]; ok {
			switch x := v.(type) {
			case float64:
				return int(x)
			case int:
				return x
			}
		}
	}
	return 0
}

func (s *SleepaceIntervalScheduler) resolveTenantResetTime(ctx context.Context, tenantID string) *alarm.TenantResetTime {
	if s.alarmCloudRepo == nil {
		def := alarm.DefaultTenantResetTime
		return &def
	}
	ac, err := s.alarmCloudRepo.GetAlarmCloud(ctx, tenantID)
	if err != nil || ac == nil || len(ac.Metadata) == 0 {
		def := alarm.DefaultTenantResetTime
		return &def
	}
	var tr alarm.TenantResetTime
	if json.Unmarshal(ac.Metadata, &tr) != nil {
		def := alarm.DefaultTenantResetTime
		return &def
	}
	// reset_time 缺字段时用 default fallback
	if tr.ResetTime.InBedTime == "" || tr.ResetTime.OutBedTime == "" {
		tr.ResetTime = alarm.DefaultTenantResetTime.ResetTime
	}
	if tr.NapTime.InBedTime == "" || tr.NapTime.OutBedTime == "" {
		tr.NapTime = alarm.DefaultTenantResetTime.NapTime
	}
	return &tr
}

// isInSleepWindow — now (local time) 是否在 rest 或 nap 的 padded 窗口内。
// rest 跨午夜（21:30→07:30）需 wrap 处理；nap 不跨午夜。
func isInSleepWindow(now time.Time, params *alarm.TenantResetTime) bool {
	if params == nil {
		return false
	}
	if isInPaddedWindow(now, params.ResetTime.InBedTime, params.ResetTime.OutBedTime, intervalEnterPadMin, intervalLeavePadMin) {
		return true
	}
	if isInPaddedWindow(now, params.NapTime.InBedTime, params.NapTime.OutBedTime, intervalEnterPadMin, intervalLeavePadMin) {
		return true
	}
	return false
}

// isInPaddedWindow — 当前 (now) 是否在 [start-enterPad, end+leavePad] 内。
// 自动检测跨午夜（startHM > endHM 视为 wrap，如 21:30→07:30）。
// HH:MM 解析失败返回 false（不触发任何切换）。
func isInPaddedWindow(now time.Time, startHM, endHM string, enterPadMin, leavePadMin int) bool {
	if startHM == "" || endHM == "" {
		return false
	}
	sh, sm, ok1 := parseHHMM(startHM)
	eh, em, ok2 := parseHHMM(endHM)
	if !ok1 || !ok2 {
		return false
	}
	const dayMin = 24 * 60
	nowMin := now.Hour()*60 + now.Minute()
	startMin := sh*60 + sm - enterPadMin
	endMin := eh*60 + em + leavePadMin
	startBaseMin := sh*60 + sm
	endBaseMin := eh*60 + em
	crossesMidnight := startBaseMin > endBaseMin

	// 把 padding 后的边界规范到 [0, dayMin)
	for startMin < 0 {
		startMin += dayMin
	}
	for endMin >= dayMin {
		endMin -= dayMin
	}

	if crossesMidnight {
		// wrap: [startMin, dayMin) ∪ [0, endMin)
		return nowMin >= startMin || nowMin < endMin
	}
	return nowMin >= startMin && nowMin < endMin
}

func parseHHMM(s string) (int, int, bool) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return 0, 0, false
	}
	var h, m int
	if _, err := fmt.Sscanf(parts[0]+":"+parts[1], "%d:%d", &h, &m); err != nil {
		return 0, 0, false
	}
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, 0, false
	}
	return h, m, true
}
