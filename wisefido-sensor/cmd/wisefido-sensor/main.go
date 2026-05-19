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

	// 3. 业务租户 ID（可选；内置租户由建库脚本写入，与启动无关）
	tenantID := os.Getenv("TENANT_ID")
	if tenantID == "" {
		logger.Info("TENANT_ID unset — wisefido-sensor starts; card polling uses no tenant until set in .env")
	}

	// 4. 创建服务
	alarmService, err := service.NewAlarmService(cfg, logger, tenantID)
	if err != nil {
		logger.Fatal("Failed to create alarm service",
			zap.Error(err),
		)
	}
	defer alarmService.Stop()

	// 5. 创建上下文（支持优雅关闭）
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

	// 5.3 Zone Engine 子系统：Bed/Room/Bathroom 状态唯一权威源 + zonealarm Warning 兜底
	zone, err := wiring.Setup(wiring.SetupOptions{
		DB:            engineDB,
		Redis:         engineRedis,
		MonitorBuffer: monitorBuf,
		BackChannel:   backChannel,
		Logger:        logger,
	})
	if err != nil {
		logger.Fatal("zone engine subsystem init failed — Warning floor lost, refusing to start",
			zap.Error(err),
			zap.String("hint", "sensor_v2_known_limitations.md L4 — zonealarm.Supervisor Stay rule is Warning 兜底"))
	}
	zone.Start(ctx)
	logger.Info("v2 fall detection wired — all 3 layers armed",
		zap.String("warning_floor", "zonealarm.Supervisor Stay rule (10min bathroom)"),
		zap.String("critical_bathroom", "BathroomFallRules §6.A 4 rules"),
		zap.String("critical_bedroom", "BedroomFallRules §6.B 11b+11c"))

	// 6. 启动服务（在 goroutine 中）
	serviceErrChan := make(chan error, 1)
	go func() {
		if err := alarmService.Start(ctx); err != nil {
			serviceErrChan <- err
		}
	}()

	// 7. 等待信号（优雅关闭）
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		logger.Info("Received signal, shutting down",
			zap.String("signal", sig.String()),
		)
		cancel() // 取消上下文，停止服务
	case err := <-serviceErrChan:
		logger.Fatal("Service error",
			zap.Error(err),
		)
	}

	logger.Info("Alarm service stopped")
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

