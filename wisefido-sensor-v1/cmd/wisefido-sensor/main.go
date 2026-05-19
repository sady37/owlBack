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
	"wisefido-sensor-v1/internal/config"
	"wisefido-sensor-v1/internal/consumer"
	"wisefido-sensor-v1/internal/service"
	"wisefido-sensor-v1/internal/zoneengine/wiring"

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

	// 5.1 RoomEngine 实时学习（独立 db/redis 连接，避免与 alarm service 共享状态）
	//      Phase 5（家属反馈接入 RecordGroundTruth）暂未开启 —— winner 用 yaml 默认 balanced
	if engineDB, engineRedis, err := openEngineDeps(cfg); err != nil {
		logger.Warn("roomengine deps init failed; engine disabled", zap.Error(err))
	} else {
		if _, err := startRoomEngine(ctx, cfg, engineDB, engineRedis, logger); err != nil {
			logger.Warn("roomengine startup failed; engine disabled", zap.Error(err))
		}
	}

	// 5.2 PR1 (A7): sensor 端 monitor 流消费者 + buffer。
	//      与 cardagg 用独立 consumer group (wisefido-sensor-monitor)，offset 互不干扰。
	//      Phase 2 (zone engine) 复用本 buffer 喂 vital adapter；故拆出 monitorBuf 跨步骤共享。
	monitorBuf := service.NewMonitorBuffer()
	if engineDB, engineRedis, err := openEngineDeps(cfg); err == nil {
		monitorConsumer := consumer.NewMonitorConsumer(engineRedis, monitorBuf, logger)
		monitorConsumer.Start(ctx)

		// 5.3 Zone Engine 子系统：Bed/Room/Bathroom 状态唯一权威源 + zonealarm 派生 alarm。
		// 注入 AlarmBackChannel 让 zonealarm 把 derived alarm（Stay/LeftBed/NightAbsence/
		// BedNightAbsence）经 iot:alarm:stream 回流给 cardagg 落库。
		zone, err := wiring.Setup(wiring.SetupOptions{
			DB:            engineDB,
			Redis:         engineRedis,
			MonitorBuffer: monitorBuf,
			BackChannel:   consumer.NewAlarmBackChannel(engineRedis, sensorAgentIdentity(cfg, logger)),
			Logger:        logger,
		})
		if err != nil {
			logger.Warn("zone engine subsystem init failed; presence 派生退化为 cardagg 旧路径",
				zap.Error(err))
		} else {
			zone.Start(ctx)
		}
	} else {
		logger.Warn("sensor monitor consumer + zone engine disabled (deps init failed)", zap.Error(err))
	}

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

