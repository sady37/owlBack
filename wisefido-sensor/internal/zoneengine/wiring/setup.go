package wiring

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	rediscommon "owl-common/redis"
	"wisefido-sensor/internal/service"
	"wisefido-sensor/internal/zoneengine"

	redislib "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// Subsystem zone engine 子系统的运行时容器（一站式工厂返回）。
//
// 持有 Engine + 4 adapter + 3 lookup 实现，由 caller 调 Start 启动各 goroutine。
type Subsystem struct {
	Engine          *zoneengine.Engine
	RadarAdapter    *zoneengine.RadarAdapter
	SleepaceAdapter *zoneengine.SleepaceAdapter
	VitalAdapter    *zoneengine.VitalAdapter
	RedisAdapter    *zoneengine.RedisAdapter

	BedSizeLookup  *BedSizeLookup
	BathroomLookup *BathroomLookup
	VitalSource    *MonitorVitalSource

	RulesPath string
	logger    *zap.Logger
}

// SetupOptions Setup 入参。
type SetupOptions struct {
	DB            *sql.DB              // 用于 BedSizeLookup / BathroomLookup
	Redis         *redislib.Client     // 用于 4 adapter
	MonitorBuffer *service.MonitorBuffer // 用于 VitalSource
	RulesPath     string               // zone_rules.yaml 绝对路径；空则查 ZONE_RULES_PATH env，再回退默认 "config/zone_rules.yaml"
	Logger        *zap.Logger
}

// defaultRulesPath  yaml 兜底位置（与 sensor 默认 cwd 一致）。
const defaultRulesPath = "config/zone_rules.yaml"

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

	// 3) engine + redis listener
	engine := zoneengine.NewEngine(rules, bedLookup, opts.Logger)
	redisAdapter := zoneengine.NewRedisAdapter(opts.Redis,
		rediscommon.StreamCardStatus.MaxLen,
		rediscommon.StreamCardRealTime.MaxLen,
		0, opts.Logger)
	engine.AddListener(redisAdapter)

	// 4) input adapters
	radar := zoneengine.NewRadarAdapter(opts.Redis, engine, bathLookup, opts.Logger)
	sleepace := zoneengine.NewSleepaceAdapter(opts.Redis, engine, opts.Logger)
	vital := zoneengine.NewVitalAdapter(vitalSrc, engine, opts.Logger)

	return &Subsystem{
		Engine:          engine,
		RadarAdapter:    radar,
		SleepaceAdapter: sleepace,
		VitalAdapter:    vital,
		RedisAdapter:    redisAdapter,
		BedSizeLookup:   bedLookup,
		BathroomLookup:  bathLookup,
		VitalSource:     vitalSrc,
		RulesPath:       rulesPath,
		logger:          opts.Logger,
	}, nil
}

// Start 启动 4 adapter + Engine.Tick 1s 周期。返回后子系统全跑起来；ctx 取消时全部退出。
func (s *Subsystem) Start(ctx context.Context) {
	s.RadarAdapter.Start(ctx)
	s.SleepaceAdapter.Start(ctx)
	s.VitalAdapter.Start(ctx)
	go s.runTickLoop(ctx)
	s.logger.Info("zone engine subsystem started",
		zap.String("rules_path", s.RulesPath))
}

// runTickLoop 1s tick 推 score decay + leaving timer + subset_invariant 周期巡检。
func (s *Subsystem) runTickLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			s.Engine.Tick(t.UnixMilli())
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
	s.BedSizeLookup.InvalidateAll()
	s.BathroomLookup.InvalidateAll()
	s.logger.Info("zone engine rules reloaded", zap.String("path", s.RulesPath))
	return nil
}
