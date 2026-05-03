// Package scheduler 用 stdlib time.Timer 实现"每日 HH:MM" / "每月 D 号 HH:MM" 触发。
//
// 不引入 robfig/cron 等第三方调度库——AI Health 的调度需求只有 daily + monthly 两条线，
// 自实现 ~50 行更易审计，依赖更少。
package scheduler

import (
	"context"
	"time"

	"wisefido-ai-health/internal/config"

	"go.uber.org/zap"
)

// JobFunc 单次任务签名。fireFor 是触发时刻对应的"目标时段"
//   - daily:   fireFor = 已完成的那一天（runDate；T+1 凌晨 02:00 跑前一天）
//   - monthly: fireFor = 已完成月份的第 1 天
type JobFunc func(ctx context.Context, fireFor time.Time) error

// Scheduler 持有所有 Job 与时区
type Scheduler struct {
	cfg    config.ScheduleConfig
	loc    *time.Location
	logger *zap.Logger
}

// New 构造（解析时区失败 fallback 到 UTC 并 warn）
func New(cfg config.ScheduleConfig, logger *zap.Logger) *Scheduler {
	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		logger.Warn("invalid timezone, fallback to UTC",
			zap.String("timezone", cfg.Timezone), zap.Error(err))
		loc = time.UTC
	}
	return &Scheduler{cfg: cfg, loc: loc, logger: logger}
}

// RunDaily 阻塞循环：每天 HH:MM 触发 fn(prevDay)，直到 ctx 取消
func (s *Scheduler) RunDaily(ctx context.Context, fn JobFunc) {
	hh, mm, _ := config.ParseHHMM(s.cfg.DailyAt)
	s.runLoop(ctx, "daily", fn, func(now time.Time) (time.Time, time.Time) {
		next := nextDailyAt(now.In(s.loc), hh, mm)
		runDate := next.AddDate(0, 0, -1) // T+1 02:00 处理 T 的数据
		return next, time.Date(runDate.Year(), runDate.Month(), runDate.Day(), 0, 0, 0, 0, s.loc)
	})
}

// RunMonthly 每月 day 号 HH:MM 触发 fn(prevMonthFirstDay)
func (s *Scheduler) RunMonthly(ctx context.Context, fn JobFunc) {
	hh, mm, _ := config.ParseHHMM(s.cfg.MonthlyAt)
	day := s.cfg.MonthlyDay
	s.runLoop(ctx, "monthly", fn, func(now time.Time) (time.Time, time.Time) {
		next := nextMonthlyAt(now.In(s.loc), day, hh, mm)
		// 触发时算"上一月"
		prev := next.AddDate(0, -1, 0)
		runMonth := time.Date(prev.Year(), prev.Month(), 1, 0, 0, 0, 0, s.loc)
		return next, runMonth
	})
}

// runLoop 通用循环：每轮重新计算下一触发时刻（避免漂移）
func (s *Scheduler) runLoop(
	ctx context.Context,
	kind string,
	fn JobFunc,
	planner func(now time.Time) (next time.Time, fireFor time.Time),
) {
	logger := s.logger.With(zap.String("loop", kind))
	for {
		now := time.Now()
		next, fireFor := planner(now)
		wait := time.Until(next)
		logger.Info("next fire scheduled",
			zap.Time("next_at", next),
			zap.Time("fire_for", fireFor),
			zap.Duration("wait", wait))

		t := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			t.Stop()
			logger.Info("scheduler stopped")
			return
		case <-t.C:
			// 单次任务超时不影响下一轮调度
			runCtx, cancel := context.WithTimeout(ctx, 6*time.Hour)
			if err := fn(runCtx, fireFor); err != nil {
				logger.Error("job failed", zap.Error(err))
			}
			cancel()
		}
	}
}

// nextDailyAt 计算下一次 hh:mm（在 now 所在时区），若今天的时刻已过则取明天
func nextDailyAt(now time.Time, hh, mm int) time.Time {
	cand := time.Date(now.Year(), now.Month(), now.Day(), hh, mm, 0, 0, now.Location())
	if !cand.After(now) {
		cand = cand.AddDate(0, 0, 1)
	}
	return cand
}

// nextMonthlyAt 计算下一次"D 号 hh:mm"（D 限制在 1..28，避开月份天数差异）
func nextMonthlyAt(now time.Time, day, hh, mm int) time.Time {
	cand := time.Date(now.Year(), now.Month(), day, hh, mm, 0, 0, now.Location())
	if !cand.After(now) {
		cand = cand.AddDate(0, 1, 0)
	}
	return cand
}
