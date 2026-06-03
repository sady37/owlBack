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
type SleepaceDstRebindScheduler struct {
	db      *sql.DB
	gateway *SleepaceGatewayClient
	logger  *zap.Logger
}

func NewSleepaceDstRebindScheduler(db *sql.DB, gateway *SleepaceGatewayClient, logger *zap.Logger) *SleepaceDstRebindScheduler {
	return &SleepaceDstRebindScheduler{db: db, gateway: gateway, logger: logger}
}

// Start 对齐到下一个 UTC 00:00 首跑，之后每 24h 一次。判据用 IANA-local 正午比较，与具体跑点无关，
// 0 点只是"一天一次"的节拍。ctx 取消即退出。
func (s *SleepaceDstRebindScheduler) Start(ctx context.Context) {
	s.logger.Info("[DST_REBIND_SCHEDULER] started")
	for {
		now := time.Now().UTC()
		next := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1)
		select {
		case <-ctx.Done():
			s.logger.Info("[DST_REBIND_SCHEDULER] stopped")
			return
		case <-time.After(time.Until(next)):
			s.tick(ctx)
		}
	}
}

type dstDeviceRow struct {
	deviceAddr string
	deviceCode string
	deviceUID  string
	iana       string
}

func (s *SleepaceDstRebindScheduler) tick(ctx context.Context) {
	if s.gateway == nil {
		return
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT host(d.device_addr), COALESCE(dfm.device_code, ''), dfm.device_uid,
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
		return
	}
	defer rows.Close()
	var devs []dstDeviceRow
	for rows.Next() {
		var r dstDeviceRow
		if err := rows.Scan(&r.deviceAddr, &r.deviceCode, &r.deviceUID, &r.iana); err != nil {
			continue
		}
		devs = append(devs, r)
	}

	rebound := 0
	for _, d := range devs {
		if d.iana == "" || d.deviceCode == "" {
			continue
		}
		loc, err := time.LoadLocation(d.iana)
		if err != nil {
			continue
		}
		nowLocal := time.Now().In(loc)
		// 当天正午 = 2am 切换之后，取到的是"切换后偏移"
		todayNoon := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), 12, 0, 0, 0, loc)
		offToday := IANAToOffsetSecondsAt(d.iana, todayNoon)
		offYesterday := IANAToOffsetSecondsAt(d.iana, todayNoon.AddDate(0, 0, -1))
		if offToday == offYesterday {
			continue // 非切换日
		}
		off := offToday
		if _, err := s.gateway.InitializeDevice(ctx, d.deviceCode, d.deviceUID, &off); err != nil {
			s.logger.Warn("[DST_REBIND_SCHEDULER] re-bind failed",
				zap.String("device_addr", d.deviceAddr), zap.String("iana", d.iana), zap.Error(err))
			continue
		}
		rebound++
		s.logger.Info("[DST_REBIND_SCHEDULER] DST transition re-bind",
			zap.String("device_addr", d.deviceAddr), zap.String("iana", d.iana),
			zap.Int("old_offset", offYesterday), zap.Int("new_offset", off))
	}
	if rebound > 0 {
		s.logger.Info("[DST_REBIND_SCHEDULER] tick done", zap.Int("devices", len(devs)), zap.Int("rebound", rebound))
	} else {
		s.logger.Debug("[DST_REBIND_SCHEDULER] tick done (no DST transition)", zap.Int("devices", len(devs)))
	}
}
