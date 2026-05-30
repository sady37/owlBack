package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/netip"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"wisefido-sensor/internal/config"
	"wisefido-sensor/internal/consumer"
	"wisefido-sensor/internal/service"
	"wisefido-sensor/internal/zoneengine/wiring"

	logpkg "owl-common/logger"
	"owl-common/spatial"

	_ "github.com/lib/pq"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// sensorAgentIdentity 解析 wisefido-sensor 进程的 platform agent IPv6 + UID
// 身份，校验 UID 与 IPv6 派生值一致（不一致 log WARN 但不 fail，dev 友好）。
//
// 详见 owlBack/doc/platform_agent_addressing.md §3。
func sensorAgentIdentity(cfg *config.Config, logger *zap.Logger) consumer.AgentIdentity {
	const agentName = "wisefido-sensor"
	id := consumer.AgentIdentity{AgentName: agentName}

	ipv6Str := strings.TrimSpace(cfg.Identity.IPv6)
	if ipv6Str == "" {
		logger.Warn("platform identity ipv6 empty; envelope.producer will be NULL — pin SENSOR_IPV6 in .env",
			zap.String("agent", agentName))
		return id
	}
	addr, err := netip.ParseAddr(ipv6Str)
	if err != nil {
		logger.Warn("platform identity ipv6 invalid; envelope.producer will be NULL",
			zap.String("agent", agentName), zap.String("ipv6", ipv6Str), zap.Error(err))
		return id
	}
	id.IPv6 = addr

	// 校验 UID（如已配置）与派生值一致；不一致 warn（pin 错了别静默漂移）
	derived, derr := spatial.DerivePlatformUID(agentName, addr.String())
	if derr != nil {
		logger.Warn("platform identity UID derivation failed", zap.Error(derr))
	} else if pinned := strings.TrimSpace(cfg.Identity.UID); pinned != "" && pinned != derived.String() {
		logger.Warn("platform identity UID mismatch — pinned .env value differs from derived value",
			zap.String("pinned", pinned), zap.String("derived", derived.String()),
			zap.String("hint", "regenerate via spatial.DerivePlatformUID and update .env / config.yaml"))
	} else if pinned == "" {
		logger.Info("platform identity UID not pinned; using derived value (please pin in .env)",
			zap.String("derived", derived.String()))
	}
	logger.Info("platform identity wired",
		zap.String("agent", agentName),
		zap.String("ipv6", addr.String()),
		zap.String("uid", derived.String()))
	return id
}

func main() {
	// 1. 加载配置
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// 2. 初始化日志
	logger, err := logpkg.NewLogger(cfg.Log.Level, cfg.Log.Format, "wisefido-sensor")
	if err != nil {
		panic(fmt.Sprintf("Failed to init logger: %v", err))
	}
	defer logger.Sync()

	// 3. 创建上下文（支持优雅关闭）
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 5.1 RoomEngine + Zone Engine + Fall rules — 三层 wire 不变量（sensor_v2 §6.A 设计原则 #3）：
	//   - zoneengine.Supervisor: Warning 兜底 (Stay rule 10min armed)
	//   - BathroomFallRules: Critical bathroom (§6.A 4 类)
	//   - BedroomFallRules:  Critical bedroom (§6.B 11b/11c + PR-11.1 sleepad gating)
	// 任一缺失 log.Fatal 拒启动（不接受降级运行，sensor_v2_known_limitations.md L4）。
	engineDB, engineRedis, err := openEngineDeps(cfg)
	if err != nil {
		logger.Fatal("roomengine deps init failed — refusing to start",
			zap.Error(err),
			zap.String("hint", "DB or Redis unreachable; check config + connectivity"))
	}

	engine, err := startRoomEngine(ctx, cfg, engineDB, engineRedis, logger)
	if err != nil {
		logger.Fatal("roomengine startup failed — refusing to start",
			zap.Error(err))
	}

	// PR-Bootstrap invariant 检查：3 层 wire 都必须 armed
	if engine.BathroomFallRulesWired() == false {
		logger.Fatal("BathroomFallRules not wired — refusing to start (sensor_v2_known_limitations.md L4)")
	}
	if engine.BedroomFallRulesWired() == false {
		logger.Fatal("BedroomFallRules not wired — refusing to start (sensor_v2_known_limitations.md L4)")
	}

	// 5.2 PR1 (A7): sensor 端 monitor 流消费者 + buffer。
	monitorBuf := service.NewMonitorBuffer()
	monitorConsumer := consumer.NewMonitorConsumer(engineRedis, monitorBuf, logger)
	monitorConsumer.Start(ctx)

	// 5.2.1 Alarm enablement cache + BackChannel gate
	// sensor 在源头按 device /128 LPM 查 spatial_config alarm.cloud_config，未启用的 alarm
	// 在此处统一 drop，不发往 cardagg。closure 注入 BackChannel，避免 consumer→service 导入环。
	enablementCache := service.NewAlarmEnablementCache(engineDB, logger)
	backChannel := consumer.NewAlarmBackChannel(engineRedis, sensorAgentIdentity(cfg, logger))
	backChannel.SetEnablement(func(ctx context.Context, deviceAddr, alarmType string) bool {
		_, ok := enablementCache.IsEnabled(ctx, deviceAddr, alarmType)
		return ok
	})

	// 5.2.2 alarm.cloud_config 失效订阅：wisefido-data 改 spatial_config 后 publish 精确
	// device_addr → 本 consumer 调 enablementCache.Invalidate；与 cardagg AlarmDeviceHandler 对称。
	// 详 [[alarm_enablement_invalidate_design]]。
	alarmDeviceConsumer := consumer.NewAlarmDeviceConfigConsumer(engineRedis, enablementCache, logger)
	alarmDeviceConsumer.Start(ctx)

	// S6: DeviceFitnessTracker — per-device 健康状态 gate；AlarmConsumer fan-out 设备类 alarm
	// 写 fitness state；adapter_radar / adapter_sleepace handleMsg 入口查 IsFit → unfit skip
	// 防 SensorDetached/AngleException 后垃圾数据污染 zone FSM 决策。
	fitnessTracker := service.NewDeviceFitnessTracker(logger)

	// 5.3 Zone Engine 子系统：Bed/Room/Bathroom 状态唯一权威源 + zonealarm Warning 兜底
	// （含 TargetStateAggregator 构造 + ZoneEvent 订阅 + StreamPublisher 60s tick 主循环）
	zone, err := wiring.Setup(wiring.SetupOptions{
		DB:            engineDB,
		Redis:         engineRedis,
		MonitorBuffer: monitorBuf,
		BackChannel:   backChannel,
		Identity:      sensorAgentIdentity(cfg, logger),
		Fitness:       fitnessTracker,
		Enablement:    enablementCache,
		Logger:        logger,
	})
	if err != nil {
		logger.Fatal("zone engine subsystem init failed — Warning floor lost, refusing to start",
			zap.Error(err),
			zap.String("hint", "sensor_v2_known_limitations.md L4 — zonealarm.Supervisor Stay rule is Warning 兜底"))
	}
	zone.Start(ctx)

	// 5.3.0 A 风险放大消费者: roomengine fall verifier 查 zone.TargetAggregator WeakBio≥80
	// → 强制 verdict=real（短路三档评分）。WeakBio 30min 滑窗 score 由 S1 alarm consumer
	// 累加；engine 转发给所有 TrackManager（已注册 + 新 RegisterRoom）。
	engine.SetWeakBioSource(zone.TargetAggregator)

	// 5.3.1 producer-side consumer wire (S1/S2/S4)
	//   S1 alarm:    iot:alarm:stream WeakBio 关联 → aggregator score 累加
	//   S2 activity: iot:event:stream category=activity → aggregator LastActive/Standing
	//   S4 sleepstage: iot:event:stream category=SleepStage → confidence ladder → bed.sleepstage publish
	// 各 consumer 用独立 consumer group，与 roomengine "roomengine" group 隔离；
	// Producer 在 fd00:0:fff1::/48 slot 的自家派生消息跳过防 loop。
	alarmConsumer := consumer.NewAlarmConsumer(engineRedis, zone.TargetAggregator, logger)
	alarmConsumer.SetFitnessSink(fitnessTracker) // S6: 设备类 alarm fan-out 到 fitness state
	alarmConsumer.Start(ctx)
	activityConsumer := consumer.NewActivityConsumer(engineRedis, zone.TargetAggregator, logger)
	activityConsumer.Start(ctx)
	sleepStageConsumer := consumer.NewSleepStageConsumer(engineRedis, zone.StreamPublisher, logger)
	// C OOB 守卫: bed FSM Vacant 时 sleepace 报 SleepStage 视作 device_failure
	// D SleepStage clear on bed transition: 通过 wiring adapter 注册 ZoneEventListener
	sleepStageConsumer.SetBedChecker(zone.Engine)
	sleepStageConsumer.SetDeviceFailureEmitter(backChannel)
	zone.Engine.AddListener(wiring.NewSleepStageClearAdapter(ctx, sleepStageConsumer))
	sleepStageConsumer.Start(ctx)

	// "device offline = 内存重启"原则（用户 2026-05-19 拍板）：fit→unfit 边沿广播清各
	// 持 per-device in-memory state 的组件。已 fire 入 DB 的 alarm 不丢；只清没 fire 的
	// pending / cache / accumulator → 同设备恢复后从干净起点累积。
	fitnessTracker.RegisterUnfitCallback(engine.OnDeviceUnfit) // roomengine: clear bedSessions/sleepadStates
	fitnessTracker.RegisterUnfitCallback(zone.TargetAggregator.ForgetDevice)
	fitnessTracker.RegisterUnfitCallback(sleepStageConsumer.OnDeviceUnfit)
	logger.Info("v2 fall detection wired — all 3 layers armed",
		zap.String("warning_floor", "zonealarm.Supervisor Stay rule (alone-anchored: day 45min / night 30min)"),
		zap.String("critical_bathroom", "BathroomFallRules §6.A 4 rules"),
		zap.String("critical_bedroom", "BedroomFallRules §6.B 11b+11c"))

	// 5.2.3 config:card:stream 订阅：wisefido-data PublishConfigChanged 后失效 + reload。
	// 替代旧 60s SetRoutesReloader 轮询；事件驱动，monitor toggle 1s 内同步到 spatial cache。
	configCardConsumer := consumer.NewConfigCardConsumer(engineRedis,
		&engineReloader{ctx: ctx, engine: engine, db: engineDB, logger: logger},
		zone.Spatial,
		logger)
	configCardConsumer.Start(ctx)

	// 重启清残留：sensor 启动后主动 publish 所有 room/bed 的 vacant 初始态，
	// OOR/OOB anchor 重置为 now，避免 cardagg card:state hash 保留昨日 LastExitTime
	// （如「OOR 29h」残留显示）。详 initial_publish.go。
	publishInitialResetState(ctx, engineDB, zone.StreamPublisher, zone.Spatial, logger)

	// 6. 等待信号（优雅关闭）
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	logger.Info("Received signal, shutting down", zap.String("signal", sig.String()))
	cancel()

	logger.Info("wisefido-sensor stopped")
}

// openEngineDeps 为 RoomEngine 单独建 db + redis 连接（与 alarm service 隔离）
func openEngineDeps(cfg *config.Config) (*sql.DB, *redis.Client, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.Database, cfg.Database.SSLMode)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, nil, fmt.Errorf("ping db: %w", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("ping redis: %w", err)
	}
	return db, rdb, nil
}

