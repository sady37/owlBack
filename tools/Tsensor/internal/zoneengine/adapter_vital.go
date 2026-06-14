package zoneengine

// adapter_vital.go — HR/RR 持续生命体征 → Bed zone sustain SignalEvidence。
//
// 周期 tick（默认 4s，对齐 MonitorBuffer 字段 TTL），扫描所有 active card 的 monitor buffer，
// 对每个 (card, bed_id) 二元组：若任一 sleepad/radar device 在 freshness 窗内同时有 HR>0
// 和 RR>0 → 发 ZoneTypeBed sustain SignalEvidence。
//
// sustain 不单独翻转状态（scorer.go: case "sustain"）；它的作用是：
//   1. 让 enter/leave 之间的 score 不被 decay 拖到 exit_threshold（防"睡着了被误判离床"）。
//   2. recent_enter_bonus +15 仅在 sustain 活时加（scorer.go:218），保证刚上床 +15 短暂强化。
//
// 设计约束（来自上游 memory）：
//   - sustain 与 enter/leave 互斥清零无关 —— 物理证据独立通道，sleepace 撒谎时 HR/RR
//     仍然让 score 升回去。
//   - 数据源不在 zone engine 包内构造（避免依赖 service 包），由调用方注入 VitalSource。

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// VitalSource zone engine 拉 vital 的抽象。实现侧（如 sensor 主进程）包装 MonitorBuffer 实现。
//
// emit 回调对每个 fresh (card, bed) 二元组调一次；caller 不应假设顺序或频次。
// freshnessMs：bedZoneID 上一次见到 HR+RR 同时存在的时间不能超过这个窗口。
// source：sleepad→"vital"（OnSleepadVital）；radar→"radar"（OnRadarVital，firmware bed-enter 门控）。
type VitalSource interface {
	ScanActiveBedVitals(nowMs, freshnessMs int64, emit func(bedZoneID, source string, ts int64))
}

// VitalAdapter 周期 tick 把 fresh HR/RR 翻译为 sustain SignalEvidence 喂 engine。
type VitalAdapter struct {
	src         VitalSource
	engine      *Engine
	interval    time.Duration
	freshnessMs int64
	logger      *zap.Logger
}

const (
	// defaultVitalInterval 与 MonitorBuffer.PruneFields 同节奏（4s）；
	// 也短于 Scorer.SustainStaleMs (10s)，保证连续在床期间 sustain 不出现"假断"。
	defaultVitalInterval = 4 * time.Second

	// defaultVitalFreshnessMs 30s — 比 sustain stale 阈值大一倍冗余。
	// sleepad 心跳通常 1~5s 一帧，30s 容忍偶发丢包。
	defaultVitalFreshnessMs int64 = 30_000
)

// NewVitalAdapter 默认 4s tick + 30s freshness。
func NewVitalAdapter(src VitalSource, engine *Engine, logger *zap.Logger) *VitalAdapter {
	return &VitalAdapter{
		src:         src,
		engine:      engine,
		interval:    defaultVitalInterval,
		freshnessMs: defaultVitalFreshnessMs,
		logger:      logger,
	}
}

// WithInterval / WithFreshness 测试 / tuning 用 setter（采链式略冗，直接 expose）。
func (a *VitalAdapter) WithInterval(d time.Duration) *VitalAdapter {
	if d > 0 {
		a.interval = d
	}
	return a
}

func (a *VitalAdapter) WithFreshness(ms int64) *VitalAdapter {
	if ms > 0 {
		a.freshnessMs = ms
	}
	return a
}

// Start 启动周期 tick。
func (a *VitalAdapter) Start(ctx context.Context) {
	go a.runLoop(ctx)
	if a.logger != nil {
		a.logger.Info("zoneengine vital adapter started",
			zap.Duration("interval", a.interval),
			zap.Int64("freshness_ms", a.freshnessMs),
		)
	}
}

func (a *VitalAdapter) runLoop(ctx context.Context) {
	t := time.NewTicker(a.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-t.C:
			a.tickOnce(now.UnixMilli())
		}
	}
}

// tickOnce 单次扫描，公开供测试驱动（不依赖真实 ticker）。
func (a *VitalAdapter) tickOnce(nowMs int64) {
	if a.src == nil {
		return
	}
	a.src.ScanActiveBedVitals(nowMs, a.freshnessMs, func(bedZoneID, source string, ts int64) {
		if bedZoneID == "" || source == "" {
			return
		}
		a.engine.Apply(SignalEvidence{
			ZoneType: ZoneTypeBed,
			ZoneID:   bedZoneID,
			Source:   source,
			Kind:     "sustain",
			Ts:       ts,
		})
	})
}
