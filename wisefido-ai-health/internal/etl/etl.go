// Package etl 包含 AI Health 离线 ETL runner。
//
// Phase 0 骨架：runner 只跑通"读 watermark → 跑空业务逻辑 → 写 watermark"链路，
// 不计算任何指标。各指标公式（gait / frailty / Cheyne-Stokes / restlessness 等）
// 在 Phase 1+ 按 doc/AI_health.md §9 顺序实现。
package etl

import (
	"context"
	"time"

	"wisefido-ai-health/internal/config"
	"wisefido-ai-health/internal/repo"

	"go.uber.org/zap"
)

// Runner 负责按调度执行单次 ETL（daily 或 monthly）
type Runner struct {
	cfg    *config.Config
	repo   *repo.Repo
	logger *zap.Logger
}

// New 构造 Runner
func New(cfg *config.Config, r *repo.Repo, logger *zap.Logger) *Runner {
	return &Runner{cfg: cfg, repo: r, logger: logger}
}

// RunDaily 按 cards × 日期 跑 daily_etl 骨架
//
// runDate 是 "目标日期"（已完成的一天，T+1 跑 T 的数据）。Phase 0 仅 log + watermark 写入。
func (r *Runner) RunDaily(ctx context.Context, runDate time.Time) error {
	const taskName = "daily_etl"
	logger := r.logger.With(
		zap.String("task", taskName),
		zap.String("run_date", runDate.Format("2006-01-02")),
		zap.Bool("dry_run", r.cfg.ETL.DryRun),
	)
	logger.Info("daily_etl start (Phase 0 skeleton)")

	// Phase 0：只跑 global watermark 链路。Phase 1+ 改为 cards 列表迭代。
	wm, err := r.repo.GetWatermark(ctx, taskName, "")
	if err != nil {
		logger.Error("get watermark failed", zap.Error(err))
		return err
	}
	logger.Info("watermark loaded",
		zap.Time("last_run_at", wm.LastRunAt),
		zap.Time("last_complete_day", wm.LastCompleteDay),
		zap.String("last_status", wm.LastStatus),
	)

	// TODO Phase 1: 按设计文档 §9 Phase 1 实现
	//   1) SELECT card_id FROM cards WHERE deleted_at IS NULL AND ...
	//   2) parallel 8 workers 处理每个 card
	//   3) 每 card 调用 mobility / sleep / vital 子模块算指标
	//   4) INSERT INTO daily_health_metrics ... ON CONFLICT (card_id, metric_date) DO UPDATE
	//   5) 每 card 失败时 LogError 并继续，不阻塞批次（§8.5）

	if r.cfg.ETL.DryRun {
		logger.Info("dry_run=true, skipping watermark write")
		return nil
	}

	wm.LastStatus = "success"
	wm.LastCompleteDay = runDate
	wm.LastRunAt = time.Now().UTC()
	wm.ErrorMessage = ""
	if err := r.repo.UpsertWatermark(ctx, wm); err != nil {
		logger.Error("upsert watermark failed", zap.Error(err))
		return err
	}
	logger.Info("daily_etl done")
	return nil
}

// RunMonthly 跑 monthly_etl 骨架
func (r *Runner) RunMonthly(ctx context.Context, runMonth time.Time) error {
	const taskName = "monthly_etl"
	logger := r.logger.With(
		zap.String("task", taskName),
		zap.String("run_month", runMonth.Format("2006-01")),
		zap.Bool("dry_run", r.cfg.ETL.DryRun),
	)
	logger.Info("monthly_etl start (Phase 0 skeleton)")

	wm, err := r.repo.GetWatermark(ctx, taskName, "")
	if err != nil {
		logger.Error("get watermark failed", zap.Error(err))
		return err
	}
	logger.Info("watermark loaded",
		zap.Time("last_run_at", wm.LastRunAt),
		zap.String("last_status", wm.LastStatus),
	)

	// TODO Phase 3: monthly 聚合 + Vitality Preservation Score 0–100
	// TODO Phase 4: cohort_health_metrics 月聚合 + 签名报告

	if r.cfg.ETL.DryRun {
		logger.Info("dry_run=true, skipping watermark write")
		return nil
	}

	wm.LastStatus = "success"
	wm.LastRunAt = time.Now().UTC()
	wm.ErrorMessage = ""
	if err := r.repo.UpsertWatermark(ctx, wm); err != nil {
		logger.Error("upsert watermark failed", zap.Error(err))
		return err
	}
	logger.Info("monthly_etl done")
	return nil
}
