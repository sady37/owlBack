package service

import (
	"context"
	"database/sql"
	"time"

	"go.uber.org/zap"
)

// SleepaceDstRebindScheduler 每日 0 点检查各 sleepad 设备的 unit IANA 时区是否在「今天」发生 DST 切换，
// 若是则用切换后偏移（采当天正午，确保是 2am 切换之后的值）重 bind 到 sleepace —— 让厂家侧
// ResetTime / leftbed / 报告 startTime 当天即用正确偏移。
//
// 与 Setting-save 触发的 ResyncDeviceTimezone 互补：人改时区走 save，每年 2 次的 DST 切换走本调度器。
// Shanghai 等无 DST 时区永不触发；有 DST 的 unit 一年触发 2 天，每天最多一次（offToday != offYesterday）。
// TimezoneResyncer 是"重绑设备时区"的唯一执行动作（手动 Setting save / 手动端点 / 本调度器 共用同一份）。
// 由 DeviceMonitorSettingsService 实现（ResyncDeviceTimezone：解析 unit IANA → 当前 DST offset → 重 bind）。
type TimezoneResyncer interface {
	ResyncDeviceTimezone(ctx context.Context, tenantID, deviceAddr string) (string, int, error)
}

type SleepaceDstRebindScheduler struct {
	db       *sql.DB
	resyncer TimezoneResyncer
	logger   *zap.Logger
}

func NewSleepaceDstRebindScheduler(db *sql.DB, resyncer TimezoneResyncer, logger *zap.Logger) *SleepaceDstRebindScheduler {
	return &SleepaceDstRebindScheduler{db: db, resyncer: resyncer, logger: logger}
}

// Start 对齐到下一个 UTC 00:00 首跑，之后每 24h 一次。判据用 IANA-local 正午比较，与具体跑点无关，
// 0 点只是"一天一次"的节拍。ctx 取消即退出。
func (s *SleepaceDstRebindScheduler) Start(ctx context.Context) {
	s.logger.Info("[DST_REBIND_SCHEDULER] started")
	// 启动即一次性纠偏：把所有 sleepad 重绑到「当前正确 offset」（含 DST），不依赖"今天是否切换日"。
	// 恢复"错过的 DST 切换日"（owl 在切换日宕机/重启 → 旧逻辑要等半年下次切换才修）+ 新装/漂移设备。
	// 与手动 Setting save 同一份 ResyncDeviceTimezone；厂家 bindDevice 幂等，已正确的设备无副作用。
	// 退避重试：data 可能先于 sleepace gateway 起好（连接拒绝全失败）→ 等它就绪再纠偏，最多 5 次。
	for attempt := 0; attempt < 5; attempt++ {
		if rebound, failed := s.tick(ctx, true); rebound > 0 || failed == 0 {
			break // 有成功 / 无设备 → 完成
		}
		s.logger.Warn("[DST_REBIND_SCHEDULER] startup corrective all failed (gateway not ready?), retry in 30s",
			zap.Int("attempt", attempt+1))
		select {
		case <-ctx.Done():
			return
		case <-time.After(30 * time.Second):
		}
	}
	for {
		now := time.Now().UTC()
		next := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1)
		select {
		case <-ctx.Done():
			s.logger.Info("[DST_REBIND_SCHEDULER] stopped")
			return
		case <-time.After(time.Until(next)):
			s.tick(ctx, false) // 每日：仅 DST 切换日重绑
		}
	}
}

type dstDeviceRow struct {
	deviceAddr string
	tenantID   string
	iana       string
}

// tick 扫所有 sleepad，决定"何时触发"，触发即调用共享动作 ResyncDeviceTimezone（与手动 Setting 同一份执行）。
// corrective=true（启动纠偏）：每台无条件重绑到当前 offset（恢复错过的 DST 切换日 / 新装漂移）；
// corrective=false（每日）：仅 DST 切换日（offToday≠offYesterday）触发。调度器只管 WHEN，动作统一在 resyncer。
// 返回 (rebound, failed)：成功重绑数 / 失败数（启动纠偏据此判断网关是否就绪、要不要重试）。
func (s *SleepaceDstRebindScheduler) tick(ctx context.Context, corrective bool) (int, int) {
	if s.resyncer == nil {
		return 0, 0
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT host(d.device_addr),
		       host(network(set_masklen(d.device_addr, 48))),
		       COALESCE(NULLIF(u.timezone, ''), NULLIF(br.timezone, ''), NULLIF(t.timezone, ''))
		  FROM device_factory_meta dfm
		  JOIN devices d ON d.device_uid = dfm.device_uid
		  JOIN units u   ON u.unit_id   = network(set_masklen(d.device_addr, 80))
		  LEFT JOIN branches br ON br.branch_id = network(set_masklen(d.device_addr, 56))
		  LEFT JOIN tenants  t  ON t.tenant_id  = network(set_masklen(d.device_addr, 48))
		 WHERE dfm.device_type = 'Sleepad' AND COALESCE(dfm.device_code, '') <> ''
	`)
	if err != nil {
		s.logger.Warn("[DST_REBIND_SCHEDULER] list devices failed", zap.Error(err))
		return 0, 0
	}
	defer rows.Close()
	var devs []dstDeviceRow
	for rows.Next() {
		var r dstDeviceRow
		if err := rows.Scan(&r.deviceAddr, &r.tenantID, &r.iana); err != nil {
			continue
		}
		devs = append(devs, r)
	}

	rebound, failed := 0, 0
	for _, d := range devs {
		if d.iana == "" || d.tenantID == "" {
			continue
		}
		// WHEN：每日模式只在 DST 切换日触发（noon-anchored 取 2am 切换之后偏移，比 now 稳）；纠偏模式无条件。
		if !corrective {
			loc, err := time.LoadLocation(d.iana)
			if err != nil {
				continue
			}
			nowLocal := time.Now().In(loc)
			todayNoon := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), 12, 0, 0, 0, loc)
			if IANAToOffsetSecondsAt(d.iana, todayNoon) == IANAToOffsetSecondsAt(d.iana, todayNoon.AddDate(0, 0, -1)) {
				continue // 非切换日
			}
		}
		// ACTION：与手动 Setting save / 手动端点完全同一份。
		iana, off, err := s.resyncer.ResyncDeviceTimezone(ctx, d.tenantID, d.deviceAddr)
		if err != nil {
			failed++
			s.logger.Warn("[DST_REBIND_SCHEDULER] resync failed",
				zap.String("device_addr", d.deviceAddr), zap.Error(err))
			continue
		}
		rebound++
		mode := "dst_transition"
		if corrective {
			mode = "startup_corrective"
		}
		s.logger.Info("[DST_REBIND_SCHEDULER] re-bind",
			zap.String("mode", mode), zap.String("device_addr", d.deviceAddr),
			zap.String("iana", iana), zap.Int("offset", off))
	}
	if rebound > 0 || failed > 0 {
		s.logger.Info("[DST_REBIND_SCHEDULER] tick done",
			zap.Int("devices", len(devs)), zap.Int("rebound", rebound), zap.Int("failed", failed))
	} else {
		s.logger.Debug("[DST_REBIND_SCHEDULER] tick done (no re-bind)", zap.Int("devices", len(devs)))
	}
	return rebound, failed
}
