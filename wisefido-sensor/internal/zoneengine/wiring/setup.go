package wiring

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"wisefido-sensor/internal/consumer"
	"wisefido-sensor/internal/service"
	"wisefido-sensor/internal/zonealarm"
	"wisefido-sensor/internal/zoneengine"

	redislib "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// Subsystem zone engine 子系统的运行时容器（一站式工厂返回）。
//
// 持有 Engine + 4 adapter + 3 lookup 实现 + zonealarm Supervisor，由 caller 调 Start
// 启动各 goroutine。
type Subsystem struct {
	Engine          *zoneengine.Engine
	RadarAdapter    *zoneengine.RadarAdapter
	SleepaceAdapter *zoneengine.SleepaceAdapter
	VitalAdapter    *zoneengine.VitalAdapter
	StreamPublisher *zoneengine.StreamPublisher

	// TargetAggregator P2 scaffold：纯 state holder，StreamPublisher 60s tick 主动 pull。
	// 监听 ZoneEvent (TotalPeople gate) + 后续 P3/P4 接 monitor / alarm 流。
	TargetAggregator *service.TargetStateAggregator

	BedSizeLookup  *BedSizeLookup
	BathroomLookup *BathroomLookup
	VitalSource    *MonitorVitalSource

	// Zonealarm zone-derived alarm 子系统（4 条规则订阅 ZoneEvent）。
	// nil 时表示禁用（AlarmBackChannel 未注入或 yaml 加载失败）。
	Zonealarm    *zonealarm.Supervisor
	AlarmFirer   *BackChannelAlarmFirer
	AlarmRulesPath string

	RulesPath string
	logger    *zap.Logger
}

// SetupOptions Setup 入参。
type SetupOptions struct {
	DB             *sql.DB                       // 用于 BedSizeLookup / BathroomLookup
	Redis          *redislib.Client              // 用于 4 adapter
	MonitorBuffer  *service.MonitorBuffer        // 用于 VitalSource
	RulesPath      string                        // zone_rules.yaml 绝对路径；空则查 ZONE_RULES_PATH env，再回退默认 "config/zone_rules.yaml"
	AlarmRulesPath string                        // zone_alarm.yaml 绝对路径；空则查 ZONE_ALARM_PATH env，再回退默认 "config/zone_alarm.yaml"
	BackChannel    *consumer.AlarmBackChannel    // sensor 现成的 alarm 回流；nil 禁用 zonealarm fire
	Identity       consumer.AgentIdentity        // sensor agent IPv6 + name；StreamPublisher 写 envelope.Producer 用
	Fitness        *service.DeviceFitnessTracker // S6: per-device 健康状态 gate；nil 时 adapter 不做 fitness 过滤
	Logger         *zap.Logger
}

// defaultRulesPath  yaml 兜底位置（与 sensor 默认 cwd 一致）。
const (
	defaultRulesPath      = "config/zone_rules.yaml"
	defaultAlarmRulesPath = "config/zone_alarm.yaml"
)

// Setup 一站式工厂：load yaml + 实例化 4 adapter + lookups + listener，但不启动 goroutine。
//
// 调用方拿 *Subsystem 后调 Start(ctx) 启动。
func Setup(opts SetupOptions) (*Subsystem, error) {
	if opts.Redis == nil {
		return nil, fmt.Errorf("zone engine wiring: Redis client required")
	}
	if opts.Logger == nil {
		opts.Logger = zap.NewNop()
	}

	// 1) load yaml；失败用 default rules
	rulesPath := opts.RulesPath
	if rulesPath == "" {
		if env := os.Getenv("ZONE_RULES_PATH"); env != "" {
			rulesPath = env
		} else {
			rulesPath = defaultRulesPath
		}
	}
	rules := zoneengine.DefaultRules()
	if _, err := os.Stat(rulesPath); err == nil {
		if r, err := zoneengine.LoadRulesFromFile(rulesPath); err != nil {
			opts.Logger.Warn("zone_rules.yaml load failed; using DefaultRules",
				zap.String("path", rulesPath), zap.Error(err))
		} else {
			rules = r
			opts.Logger.Info("zone_rules.yaml loaded", zap.String("path", rulesPath))
		}
	} else {
		opts.Logger.Warn("zone_rules.yaml not found; using DefaultRules",
			zap.String("path", rulesPath))
	}

	// 2) lookups
	bedLookup := NewBedSizeLookup(opts.DB, opts.Logger)
	bathLookup := NewBathroomLookup(opts.DB, opts.Logger)
	vitalSrc := NewMonitorVitalSource(opts.MonitorBuffer)

	// 3) engine + stream publisher（替代旧 RedisAdapter 直写 card:status；
	//    cardagg 是 card:status 单 writer，sensor 通过 sensor:derived:stream 推消息）
	engine := zoneengine.NewEngine(rules, bedLookup, opts.Logger)
	streamPublisher := zoneengine.NewStreamPublisher(opts.Redis, opts.Logger)
	streamPublisher.SetIdentity(opts.Identity)
	engine.AddListener(streamPublisher)

	// 3.1) TargetStateAggregator：纯 state holder，订阅 ZoneEvent 缓存 TotalPeople；
	//      monitor / alarm 流消费 FollowUp 接；WeakBio score 仅作风险描述符，不 publish escalation。
	aggregator := service.NewTargetStateAggregator(opts.Redis, opts.Logger)
	engine.AddListener(targetAggregatorListenerAdapter{agg: aggregator})
	streamPublisher.SetAggregator(aggregator)

	// 3.2) RoomState people count dedup ([[bed_presence_fusion]])：
	//        radarRoomCount  = adapter_radar applyCount 旁路记 raw Z
	//        bedPresence     = adapter_sleepace/radar bed enter/leave 旁路记 X/Y
	//      publisher OnZoneEvent ZoneTypeRoom 用 (Z + Σ X where !Y) 覆写 TotalPeople。
	radarRoomCount := service.NewRadarRoomCountCache()
	bedPresence := service.NewBedPresenceFusion()
	streamPublisher.SetRoomDedup(radarRoomCount, bedPresence)

	// 4) input adapters
	radar := zoneengine.NewRadarAdapter(opts.Redis, engine, bathLookup, opts.Logger)
	sleepace := zoneengine.NewSleepaceAdapter(opts.Redis, engine, opts.Logger)
	if opts.Fitness != nil {
		radar.SetFitnessChecker(opts.Fitness)
		sleepace.SetFitnessChecker(opts.Fitness)
	}
	radar.SetBedPresence(bedPresence)
	radar.SetRoomCountSink(radarRoomCount)
	sleepace.SetBedPresence(bedPresence)
	vital := zoneengine.NewVitalAdapter(vitalSrc, engine, opts.Logger)

	// 5) zonealarm — zone-derived alarm 子系统（4 条规则订阅 ZoneEvent）
	alarmRulesPath := opts.AlarmRulesPath
	if alarmRulesPath == "" {
		if env := os.Getenv("ZONE_ALARM_PATH"); env != "" {
			alarmRulesPath = env
		} else {
			alarmRulesPath = defaultAlarmRulesPath
		}
	}
	alarmRules := zonealarm.DefaultRules()
	if _, err := os.Stat(alarmRulesPath); err == nil {
		if r, err := zonealarm.LoadFromFile(alarmRulesPath); err != nil {
			opts.Logger.Warn("zone_alarm.yaml load failed; using DefaultRules",
				zap.String("path", alarmRulesPath), zap.Error(err))
		} else {
			alarmRules = r
			opts.Logger.Info("zone_alarm.yaml loaded",
				zap.String("path", alarmRulesPath), zap.Int("rules", len(r)))
		}
	} else {
		opts.Logger.Warn("zone_alarm.yaml not found; using DefaultRules",
			zap.String("path", alarmRulesPath))
	}

	bedDeviceLookup := NewBedDeviceLookup(opts.DB, opts.Logger)

	// 启动时清空 Redis alarm:pending Hash（platform-agent cutover 安全网）。
	// 详见 owlBack/doc/platform_agent_addressing.md §6/§7：
	// sensor.Supervisor 用 in-memory pending map + timer，不做 HA recovery；
	// cardagg 端的 Redis Hash 已不再读写，启动 cleanup 防上次崩溃残留导致 stale state。
	if opts.Redis != nil {
		const alarmPendingKey = "alarm:pending"
		if err := opts.Redis.Del(context.Background(), alarmPendingKey).Err(); err != nil {
			opts.Logger.Warn("startup cleanup alarm:pending failed (non-fatal)",
				zap.String("key", alarmPendingKey), zap.Error(err))
		} else {
			opts.Logger.Info("startup cleanup alarm:pending done", zap.String("key", alarmPendingKey))
		}
	}

	var firer *BackChannelAlarmFirer
	var supervisor *zonealarm.Supervisor
	if opts.BackChannel != nil {
		firer = NewBackChannelAlarmFirer(opts.BackChannel, bedDeviceLookup, opts.Logger)
		supervisor = zonealarm.NewSupervisor(alarmRules, firer, opts.Logger)
		engine.AddListener(zoneAlarmListenerAdapter{sup: supervisor})
	} else {
		opts.Logger.Warn("zonealarm disabled (BackChannel not provided)")
	}

	return &Subsystem{
		Engine:           engine,
		RadarAdapter:     radar,
		SleepaceAdapter:  sleepace,
		VitalAdapter:     vital,
		StreamPublisher:  streamPublisher,
		TargetAggregator: aggregator,
		BedSizeLookup:    bedLookup,
		BathroomLookup:   bathLookup,
		VitalSource:      vitalSrc,
		Zonealarm:        supervisor,
		AlarmFirer:       firer,
		AlarmRulesPath:   alarmRulesPath,
		RulesPath:        rulesPath,
		logger:           opts.Logger,
	}, nil
}

// zoneAlarmListenerAdapter — zonealarm.Supervisor 实现 zoneengine.ZoneEventListener。
// 不直接让 Supervisor 嵌入 zoneengine 接口（避免 zonealarm 包反向依赖 zoneengine 接口
// 形态）；本 adapter 是窄边界。
type zoneAlarmListenerAdapter struct{ sup *zonealarm.Supervisor }

func (z zoneAlarmListenerAdapter) OnZoneEvent(e zoneengine.ZoneEvent) { z.sup.OnZoneEvent(e) }

// targetAggregatorListenerAdapter — service.TargetStateAggregator 通过窄接口订阅 ZoneEvent。
// Aggregator 的 OnZoneEvent 是 (cardID, zoneID, totalPeople, ts) 四参数（避免反向依赖 zoneengine 类型），
// 此 adapter 做翻译。
type targetAggregatorListenerAdapter struct{ agg *service.TargetStateAggregator }

func (t targetAggregatorListenerAdapter) OnZoneEvent(e zoneengine.ZoneEvent) {
	t.agg.OnZoneEvent(e.CardID, e.ZoneID, e.NewState.Count, e.NewState.UpdatedAt)
}

// sleepStageClearListenerAdapter (D) — consumer.SleepStageConsumer 通过 zone bed FSM
// transition 触发清零。filter ZoneType=Bed + Transition∈{Leaving, Vacant} 后调用
// OnBedVacant 清本地 ladder state + emit bed.sleepstage(0,0)。
//
// Leaving 也触发：人开始离床的瞬间就清 SleepStage，避免 leaving 过渡期 sleepace 还在
// 假报 SleepStage（设备脱床 / 干扰）污染 FE 显示。
type sleepStageClearListenerAdapter struct {
	c   *consumer.SleepStageConsumer
	ctx context.Context
}

func (s sleepStageClearListenerAdapter) OnZoneEvent(e zoneengine.ZoneEvent) {
	if e.ZoneType != zoneengine.ZoneTypeBed {
		return
	}
	if e.Transition != zoneengine.TransitionLeaving && e.Transition != zoneengine.TransitionVacant {
		return
	}
	s.c.OnBedVacant(s.ctx, e.ZoneID, e.Ts)
}

// NewSleepStageClearAdapter (D) main wiring 用：让 SleepStageConsumer 注册到 engine 听 bed FSM
// transition。返回值实现 zoneengine.ZoneEventListener，caller 调 engine.AddListener。
//
// ctx 是长生命周期 context（与 sensor 进程同寿命）；publish failure 仅 warn 不阻塞。
func NewSleepStageClearAdapter(ctx context.Context, c *consumer.SleepStageConsumer) zoneengine.ZoneEventListener {
	return sleepStageClearListenerAdapter{c: c, ctx: ctx}
}

// Start 启动 4 adapter + Engine.Tick 1s 周期 + zonealarm Supervisor.Tick（如已 wire）
//        + TargetStateAggregator + StreamPublisher 60s tick。
// 返回后子系统全跑起来；ctx 取消时全部退出。
func (s *Subsystem) Start(ctx context.Context) {
	s.RadarAdapter.Start(ctx)
	s.SleepaceAdapter.Start(ctx)
	s.VitalAdapter.Start(ctx)
	go s.runTickLoop(ctx)

	// TargetStateAggregator 主循环（消费 monitor/alarm/zone push channel）
	go s.TargetAggregator.Run(ctx)
	// StreamPublisher 60s tick：pull aggregator + 合并 publish RoomState/Target
	go s.StreamPublisher.Run(ctx)

	s.logger.Info("zone engine subsystem started",
		zap.String("rules_path", s.RulesPath),
		zap.Bool("zonealarm_wired", s.Zonealarm != nil),
		zap.String("alarm_rules_path", s.AlarmRulesPath),
		zap.Bool("target_aggregator_wired", s.TargetAggregator != nil),
	)
}

// runTickLoop 1s tick 推 score decay + leaving timer + subset_invariant 周期巡检
// + zonealarm pending fire timer。共用单一 ticker 节奏。
func (s *Subsystem) runTickLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			now := t.UnixMilli()
			s.Engine.Tick(now)
			if s.Zonealarm != nil {
				s.Zonealarm.Tick(now)
			}
		}
	}
}

// ReloadRulesFromFile hot reload 入口（暂未接 stream subscriber，调试 / admin tool 直接调）。
//
// 副作用：替换所有 zoneInstance 的规则指针；运行时状态（score / status / timer）保留。
func (s *Subsystem) ReloadRulesFromFile() error {
	r, err := zoneengine.LoadRulesFromFile(s.RulesPath)
	if err != nil {
		return fmt.Errorf("reload %s: %w", s.RulesPath, err)
	}
	s.Engine.ReloadRules(r)
	// BedSizeLookup / BathroomLookup 用 60s TTL 自动收敛，hot-reload 时不需手动清。
	s.logger.Info("zone engine rules reloaded", zap.String("path", s.RulesPath))
	return nil
}
