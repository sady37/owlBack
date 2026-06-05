package wiring

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"owl-common/card"

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

	// Spatial 统一替代旧 4 lookup（bathroom/bed_size/bed_device/unit_property）—
	// 一次 LoadAll 全 entries+units，60s tick refresh。
	Spatial     *SpatialCache
	VitalSource *MonitorVitalSource

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
	DB             *sql.DB                       // 用于 SpatialCache (rooms/beds/devices/units) + zonealarm DeviceLookup
	Redis          *redislib.Client              // 用于 4 adapter
	MonitorBuffer  *service.MonitorBuffer        // 用于 VitalSource
	RulesPath      string                        // zone_rules.yaml 绝对路径；空则查 ZONE_RULES_PATH env，再回退默认 "config/zone_rules.yaml"
	AlarmRulesPath string                        // zone_alarm.yaml 绝对路径；空则查 ZONE_ALARM_PATH env，再回退默认 "config/zone_alarm.yaml"
	BackChannel    *consumer.AlarmBackChannel    // sensor 现成的 alarm 回流；nil 禁用 zonealarm fire
	Identity       consumer.AgentIdentity        // sensor agent IPv6 + name；StreamPublisher 写 envelope.Producer 用
	Fitness        *service.DeviceFitnessTracker // S6: per-device 健康状态 gate；nil 时 adapter 不做 fitness 过滤
	Enablement     *service.AlarmEnablementCache // per-device alarm 使能 + 阈值 (spatial_config alarm.cloud_config)；nil → zonealarm 退化静态 yaml
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

	// 2) unified spatial cache (一次 LoadAll 替代旧 4 lookups)
	spatial := NewSpatialCache(opts.DB, opts.Logger)
	if err := spatial.LoadAll(context.Background()); err != nil {
		opts.Logger.Warn("spatial cache initial load failed; continuing with empty cache",
			zap.Error(err))
	}
	// HasRadarInRoom 需要 fitness 才能判 radar offline 等同无 radar；nil 时退化 DB-only。
	if opts.Fitness != nil {
		spatial.SetFitness(opts.Fitness)
	}
	vitalSrc := NewMonitorVitalSource(opts.MonitorBuffer)

	// 3) engine + stream publisher（替代旧 RedisAdapter 直写 card:status；
	//    cardagg 是 card:status 单 writer，sensor 通过 sensor:derived:stream 推消息）
	engine := zoneengine.NewEngine(rules, spatial, opts.Logger)
	// Bayesian bed FSM 启用：unit_property 驱动 Facility/Home LR 档 + 0.70/0.75 阈值。
	// legacy Scorer+StateMachine 仅 room/bathroom 用。
	engine.SetUseBedBayesian(true, spatial)
	// subset_invariant lift_parent 必须知道 /88 是否真有 radar，否则纯 sleepace 房卡死 Occupied。
	engine.SetRadarPresenceLookup(spatial)
	streamPublisher := zoneengine.NewStreamPublisher(opts.Redis, opts.Logger)
	streamPublisher.SetIdentity(opts.Identity)
	streamPublisher.SetEngine(engine) // publisher 派生 AloneContinuousMin 时按 (ZoneType,ZoneID) 查 engine.GetState
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
	radar := zoneengine.NewRadarAdapter(opts.Redis, engine, spatial, opts.Logger)
	// BedResolver 由 main wiring 注入 MatrixCache（需 engine 几何，在 Setup 之后构造）；见 main.go。
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
		firer = NewBackChannelAlarmFirer(opts.BackChannel, spatial, opts.Logger)
		supervisor = zonealarm.NewSupervisor(alarmRules, firer, opts.Logger)
		// state-anchored 重构：Supervisor 不再订阅 ZoneEvent，改 Tick 从 StreamPublisher
		// 拿 snapshot 判 fire；同时让 publisher 每 5s coalesce refresh RiskLevel。
		supervisor.SetProvider(newStreamPublisherSnapshotter(streamPublisher))
		supervisor.SetRefresher(streamPublisher)
		// per-device 阈值/使能（spatial_config alarm.cloud_config）— nil 时 evaluator 退化按 yaml 默认全启用。
		// 失效链路：alarmDeviceConsumer (main.go) 已 wire enablementCache.Invalidate(deviceAddr)。
		supervisor.SetDeviceLookup(&spatialDeviceLookupAdapter{spatial: spatial})
		if opts.Enablement != nil {
			supervisor.SetThresholdResolver(&enablementResolver{cache: opts.Enablement})
		}
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
		Spatial:     spatial,
		VitalSource: vitalSrc,
		Zonealarm:        supervisor,
		AlarmFirer:       firer,
		AlarmRulesPath:   alarmRulesPath,
		RulesPath:        rulesPath,
		logger:           opts.Logger,
	}, nil
}

// streamPublisherSnapshotter 把 sensor StreamPublisher 包装成 zonealarm.StateProvider。
//
// 类型适配：StreamPublisher.SnapshotRoomStates 返 map[string]zoneengine.RoomStateSnapshot
// (含 *card.RoomState + Kind)；zonealarm.RoomEntry 字段一致，深拷贝转字段。
type streamPublisherSnapshotter struct {
	pub *zoneengine.StreamPublisher
}

func newStreamPublisherSnapshotter(pub *zoneengine.StreamPublisher) *streamPublisherSnapshotter {
	return &streamPublisherSnapshotter{pub: pub}
}

func (s *streamPublisherSnapshotter) SnapshotRooms() map[string]zonealarm.RoomEntry {
	if s == nil || s.pub == nil {
		return nil
	}
	raw := s.pub.SnapshotRoomStates()
	out := make(map[string]zonealarm.RoomEntry, len(raw))
	for cidr, entry := range raw {
		out[cidr] = zonealarm.RoomEntry{State: entry.State, Kind: entry.Kind}
	}
	return out
}

func (s *streamPublisherSnapshotter) SnapshotBeds() map[string]*card.BedState {
	if s == nil || s.pub == nil {
		return nil
	}
	return s.pub.SnapshotBedStates()
}

// targetAggregatorListenerAdapter — service.TargetStateAggregator 通过窄接口订阅 ZoneEvent。
// Aggregator 的 OnZoneEvent 是 (spatialPrefix, totalPeople, ts) 三参数（避免反向依赖 zoneengine 类型），
// 此 adapter 做翻译。
type targetAggregatorListenerAdapter struct{ agg *service.TargetStateAggregator }

func (t targetAggregatorListenerAdapter) OnZoneEvent(e zoneengine.ZoneEvent) {
	t.agg.OnZoneEvent(e.ZoneID, e.NewState.Count, e.NewState.UpdatedAt)
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
//        + TargetStateAggregator + StreamPublisher 60s tick + SpatialCache 60s refresh。
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
	// SpatialCache 60s tick refresh（替代旧 4 lookup 的 per-key TTL）
	if s.Spatial != nil {
		go s.Spatial.Run(ctx)
	}

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
	// SpatialCache 60s tick 自动 refresh，hot-reload 时不需手动清。
	s.logger.Info("zone engine rules reloaded", zap.String("path", s.RulesPath))
	return nil
}

// spatialDeviceLookupAdapter 包 SpatialCache.FindPrimaryDevice 为 zonealarm.DeviceLookup 形态。
// (zonealarm 用 string 而非 netip.Addr — 给 Evaluator 用 spatial_config LPM 查 canonical IPv6 文本)
type spatialDeviceLookupAdapter struct {
	spatial *SpatialCache
}

func (a *spatialDeviceLookupAdapter) FindPrimaryDevice(zoneID, alarmType string) string {
	if a == nil || a.spatial == nil {
		return ""
	}
	addr := a.spatial.FindPrimaryDevice(zoneID, alarmType)
	if !addr.IsValid() {
		return ""
	}
	return addr.String()
}

// enablementResolver 把 service.AlarmEnablementCache 包成 zonealarm.ThresholdResolver。
//
// 解析顺序：duration_sec 优先（默认 cloud_config 用秒），duration_min*60 次之（部分
// 历史项用分钟，见 owl-common/alarm DefaultAlarmSetting.Radar Stay = ParamDurationMin:45）。
// 都缺 → fallback。
type enablementResolver struct {
	cache *service.AlarmEnablementCache
}

func (r *enablementResolver) Resolve(ctx context.Context, deviceAddr, alarmType string, fallbackSec int) (int, bool) {
	if r == nil || r.cache == nil {
		return fallbackSec, true
	}
	ea, ok := r.cache.IsEnabled(ctx, deviceAddr, alarmType)
	if !ok || ea == nil {
		return fallbackSec, false
	}
	if sec := paramInt(ea.AlarmParams, "duration_sec"); sec > 0 {
		return sec, true
	}
	if min := paramInt(ea.AlarmParams, "duration_min"); min > 0 {
		return min * 60, true
	}
	return fallbackSec, true
}

// paramInt 从 AlarmParams 取整数值，兼容 float64 / int / int64（JSONB 解出来常是 float64）。
func paramInt(params map[string]interface{}, key string) int {
	if params == nil {
		return 0
	}
	v, ok := params[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	}
	return 0
}
