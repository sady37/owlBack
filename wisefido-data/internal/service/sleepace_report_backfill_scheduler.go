package service

import (
	"context"
	"database/sql"
	"time"

	"go.uber.org/zap"
)

// SleepaceReportBackfillScheduler — 周期对所有 bound sleepad device 执行 vendor → DB 报告补齐。
//
// 背景：原同步路径（GetSleepaceReports list 入口 + GetSleepaceReportDetail 周效率）在用户进入页面时
// 等待厂家 vendor download 多段串行返回，新 device 30 天 missing 实测阻塞 5-30s。
// 拆分后即时路径只读 cached DB，本 scheduler 在后台周期 backfill。
//
// 与 [SleepaceIntervalScheduler] 同设计原则：
//   - INNER JOIN units 过滤库存（未 bound to unit 的 sleepad）
//   - 仅 device_code 非空（厂家 API 要 device_code）
//   - 同模式可参考 memory: sleepace-interval-scheduler
//
// tick 频率默认 1h（vendor backfill 不需要频繁；新 device 第一次进入可能等下个 tick 才补齐，
// 用户可手动 refresh 触发下一次进入）。
// 回溯窗口：30 天（覆盖 list 默认范围 + detail 周效率窗口）。
type SleepaceReportBackfillScheduler struct {
	db            *sql.DB
	reportService SleepaceReportService
	logger        *zap.Logger

	tickInterval   time.Duration
	lookbackDays   int
}

func NewSleepaceReportBackfillScheduler(
	db *sql.DB,
	reportService SleepaceReportService,
	logger *zap.Logger,
) *SleepaceReportBackfillScheduler {
	return &SleepaceReportBackfillScheduler{
		db:             db,
		reportService:  reportService,
		logger:         logger,
		tickInterval:   1 * time.Hour,
		lookbackDays:   30,
	}
}

// Start 启动后台 ticker。ctx 取消即退出。
// 启动后立即跑一次（让重启 / 新部署快速收敛），之后按 tickInterval 周期跑。
func (s *SleepaceReportBackfillScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.tickInterval)
	defer ticker.Stop()
	s.logger.Info("[REPORT_BACKFILL_SCHEDULER] started",
		zap.Duration("tick", s.tickInterval),
		zap.Int("lookback_days", s.lookbackDays),
	)
	s.tick(ctx)
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("[REPORT_BACKFILL_SCHEDULER] stopped")
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

type backfillDeviceRow struct {
	deviceID string
	tenantID string // /48 CIDR
}

func (s *SleepaceReportBackfillScheduler) tick(ctx context.Context) {
	devices, err := s.listSleepadDevices(ctx)
	if err != nil {
		s.logger.Warn("[REPORT_BACKFILL_SCHEDULER] list devices failed", zap.Error(err))
		return
	}
	if len(devices) == 0 {
		return
	}

	now := time.Now()
	endDate := dateToInt(now)
	startDate := dateToInt(now.AddDate(0, 0, -s.lookbackDays))

	processed, errored := 0, 0
	for _, d := range devices {
		// 每 device 串行（避免对厂家 API 过度并发；count <100 单 tick ~30s 内完成可接受）
		if err := s.reportService.BackfillFromVendor(ctx, d.tenantID, d.deviceID, startDate, endDate); err != nil {
			errored++
			s.logger.Debug("[REPORT_BACKFILL_SCHEDULER] device backfill failed",
				zap.String("device_id", d.deviceID),
				zap.Error(err))
			continue
		}
		processed++
	}
	// no-op tick（无设备处理、无错误）走 Debug；有处理或错误才 Info（运维需看的真事件）
	if processed == 0 && errored == 0 {
		s.logger.Debug("[REPORT_BACKFILL_SCHEDULER] tick done (no-op)",
			zap.Int("devices", len(devices)),
			zap.Int("start_date", startDate),
			zap.Int("end_date", endDate),
		)
		return
	}
	s.logger.Info("[REPORT_BACKFILL_SCHEDULER] tick done",
		zap.Int("devices", len(devices)),
		zap.Int("processed", processed),
		zap.Int("errored", errored),
		zap.Int("start_date", startDate),
		zap.Int("end_date", endDate),
	)
}

// listSleepadDevices — 同 SleepaceIntervalScheduler 过滤策略：INNER JOIN units 排除库存 + device_code 非空。
func (s *SleepaceReportBackfillScheduler) listSleepadDevices(ctx context.Context) ([]backfillDeviceRow, error) {
	// Phase 2 一刀切：identity = device_uid；deviceID 字段承载 device_uid
	rows, err := s.db.QueryContext(ctx, `
		SELECT dfm.device_uid,
		       host(network(set_masklen(d.device_addr, 48))) || '/48' AS tenant_id
		  FROM device_factory_meta dfm
		  JOIN devices d ON d.device_uid = dfm.device_uid
		  JOIN units u   ON u.unit_id   = network(set_masklen(d.device_addr, 80))
		 WHERE dfm.device_type = 'Sleepad'
		   AND COALESCE(dfm.device_code, '') <> ''
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []backfillDeviceRow
	for rows.Next() {
		var r backfillDeviceRow
		if err := rows.Scan(&r.deviceID, &r.tenantID); err != nil {
			continue
		}
		out = append(out, r)
	}
	return out, nil
}
