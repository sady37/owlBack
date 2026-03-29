package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"wisefido-data/internal/config"
	"wisefido-data/internal/domain"
	httpapi "wisefido-data/internal/http"
	"wisefido-data/internal/notify"
	"wisefido-data/internal/publisher"
	"wisefido-data/internal/repository"
	"wisefido-data/internal/service"
	"wisefido-data/internal/store"
	"wisefido-data/internal/subscriber"

	"owl-common/card"
	"owl-common/database"
	logpkg "owl-common/logger"
	rediscommon "owl-common/redis"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()

	// 初始化Logger（SaaS多租户日志管理：自动添加service_name字段）
	logger, err := logpkg.NewLogger(cfg.Log.Level, cfg.Log.Format, "wisefido-data")
	if err != nil {
		// 如果日志初始化失败，使用标准库log输出错误
		log.Printf("Failed to initialize logger: %v, using default logger", err)
		logger, _ = zap.NewProduction()
	}
	defer logger.Sync()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	kv := store.NewRedisKV(redisClient)

	// Tenants management (platform-level)
	var tenantsRepo repository.TenantsRepository
	authStore := httpapi.NewAuthStore()
	// Dev bootstrap: ensure System tenant has a usable SystemAdmin login even when DB is enabled.
	// This keeps the intended flow: SystemAdmin creates/manages other tenants.
	if os.Getenv("SEED_SYSADMIN") != "false" {
		_ = authStore.UpsertUser(httpapi.SystemTenantID(), "sysadmin", "SystemAdmin", "ChangeMe123!")
	}

	// Optional DB-backed admin APIs (units/rooms/beds/devices)
	var db *sql.DB
	var cardRepo *repository.PostgresCardRepository
	var cardRealtimeSvc *service.CardRealtimeService
	var cardSyncService *service.CardSyncService
	var devicesRepo *repository.PostgresDevicesRepository
	var unitsRepo *repository.PostgresUnitsRepository
	var deviceStoreRepo *repository.PostgresDeviceStoreRepository
	var residentsRepo *repository.PostgresResidentsRepository
	var branchesRepo *repository.PostgresBranchesRepository

	var usersRepo *repository.PostgresUsersRepository
	var sleepaceReportService service.SleepaceReportService
	// Stub depends on tenantsRepo + authStore (used by /auth/api/v1/institutions/search + /auth/api/v1/login)
	stub := httpapi.NewStubHandler(nil, authStore, nil)
	// Always register admin routes; if DB is not available, AdminAPI will fall back to stub (no 404).
	admin := httpapi.NewAdminAPI(stub, logger)
	if cfg.DBEnabled {
		if d, err := database.NewPostgresDB(&cfg.Database); err == nil {
			db = d
			logger.Info("DB enabled for wisefido-data")
		} else {
			logger.Warn("DB enabled but connection failed, falling back to stub", zap.Error(err))
		}
	}

	if db != nil {
		// DB可用时创建完整功能的服务
		unitsRepo = repository.NewPostgresUnitsRepository(db)
		devicesRepo = repository.NewPostgresDevicesRepository(db)
		devicesRepo.SetLogger(logger) // Set logger for device connection logging
		deviceStoreRepo = repository.NewPostgresDeviceStoreRepository(db)
		tenantsRepo = repository.NewPostgresTenantsRepository(db)
		// Note: StubHandler still uses TenantsRepo (old interface), but we need TenantsRepository for AuthService
		// For now, pass nil to StubHandler since it's mainly used for fallback
		stub = httpapi.NewStubHandler(nil, authStore, db)
		stub.SetLogger(logger) // Set logger for user login logging
		admin = httpapi.NewAdminAPI(stub, logger)

		// 创建 Role 和 RolePermission Service（仅创建，不保留变量）
		roleRepo := repository.NewPostgresRolesRepository(db)
		rolePermRepo := repository.NewPostgresRolePermissionsRepository(db)
		usersRepo = repository.NewPostgresUsersRepository(db)
		_ = service.NewRoleService(roleRepo, usersRepo, logger)
		_ = service.NewRolePermissionService(rolePermRepo, logger)

	} else {
		// DB不可用时使用简化的处理器
	}

	router := httpapi.NewRouter(logger)

	var cardStaticSvc *service.CardStaticService
	var dataStreamSubscriber *subscriber.DataStreamSubscriber
	var sessionStore *httpapi.MemSessionStore // 创建会话存储（用于 authService 和 authMiddleware）

	// 初始化会话存储（无论 db 是否存在都需要）
	sessionStore = httpapi.NewMemSessionStore()

	if db != nil {
		// 创建卡片实时数据处理器并注册路由
		residentsRepo = repository.NewPostgresResidentsRepository(db)
		cardRepo := repository.NewPostgresCardRepository(db, logger)
		allowedProvider := service.NewAllowedCardIDsProvider(kv, usersRepo, residentsRepo, db, cardRepo, logger)
		cardRealtimeSvc = service.NewCardRealtimeService(kv, allowedProvider, logger, nil)
		cardStaticSvc = service.NewCardStaticService(db, allowedProvider, cardRealtimeSvc, logger)

		cardRealtimeHandler := httpapi.NewCardRealtimeHandler(cardRealtimeSvc, cardStaticSvc, logger)
		router.RegisterCardRealtimeRoutes(cardRealtimeHandler)

		// 创建 MontitorHandler（前端 Monitor API）
		MontitorHandler := httpapi.NewMonitorHandler(logger)
		MontitorHandler.SetCardService(cardStaticSvc)
		MontitorHandler.SetRealtimeService(cardRealtimeSvc)
		router.RegisterMonitorRoutes(MontitorHandler)

	}

	if db != nil {
		// DB bootstrap: ensure System tenant + sysadmin exist in DB for UI pages that query users/roles.
		// Login still uses AuthStore hashes, but keeping DB in sync makes admin pages behave as expected.
		if os.Getenv("SEED_SYSADMIN") != "false" {
			// 1) Ensure System tenant row exists (FK target for users)
			_, _ = db.Exec(
				`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
				 VALUES ($1, $2, $3, 'active')
				 ON CONFLICT (tenant_id)
				 DO UPDATE SET tenant_name = EXCLUDED.tenant_name,
				              domain = EXCLUDED.domain,
				              status = 'active'`,
				httpapi.SystemTenantID(),
				"System",
				"system.local",
			)

			// 2) Ensure sysadmin user exists in DB (so "User Management" in System tenant isn't empty)
			// Use simple password hash (no salt) for compatibility with existing DB format
			// password_hash = SHA256(password) - no salt, matches existing database format
			ah, _ := hex.DecodeString(httpapi.HashAccount("sysadmin"))
			aph, _ := hex.DecodeString(service.GeneratePasswordHash("ChangeMe123!"))
			if len(ah) > 0 && len(aph) > 0 {
				_, _ = db.Exec(
					`INSERT INTO users (tenant_id, user_account, user_account_hash, password_hash, nickname, role, status)
					 VALUES ($1, $2, $3, $4, $5, $6, 'active')
					 ON CONFLICT (tenant_id, user_account)
					 DO UPDATE SET user_account_hash = EXCLUDED.user_account_hash,
					               password_hash = EXCLUDED.password_hash,
					               nickname = EXCLUDED.nickname,
					               role = EXCLUDED.role,
					               status = 'active'`,
					httpapi.SystemTenantID(),
					"sysadmin",
					ah,
					aph,
					"SystemAdmin",
					"SystemAdmin",
				)
			}
		}

		// 创建 QinglanClient（调用 wisefido-qinglan HTTP API，统一与设备通信）
		qinglanClient := service.NewQinglanClient(cfg.Qinglan.APIBaseURL, logger)

		admin = httpapi.NewAdminAPI(stub, logger)

		// 创建 Role 和 RolePermission Service 和 Handler
		roleRepo := repository.NewPostgresRolesRepository(db)
		rolePermRepo := repository.NewPostgresRolePermissionsRepository(db)
		usersRepo := repository.NewPostgresUsersRepository(db)
		roleService := service.NewRoleService(roleRepo, usersRepo, logger)
		rolePermService := service.NewRolePermissionService(rolePermRepo, logger)
		rolesHandler := httpapi.NewRolesHandler(roleService, logger)
		rolePermHandler := httpapi.NewRolePermissionsHandler(rolePermService, logger)
		router.RegisterRolesRoutes(rolesHandler)
		router.RegisterRolePermissionsRoutes(rolePermHandler)

		// 创建 Card Repository（DB card 数据操作）
		cardRepo = repository.NewPostgresCardRepository(db, logger)

		// 创建 AlarmCloud Service 和 Handler
		alarmCloudRepo := repository.NewPostgresAlarmCloudRepository(db)
		configVersionsRepo := repository.NewPostgresConfigVersionsRepository(db)
		alarmCloudService := service.NewAlarmCloudService(alarmCloudRepo, configVersionsRepo, db, logger)
		alarmCloudHandler := httpapi.NewAlarmCloudHandler(alarmCloudService, logger)
		router.RegisterAlarmCloudRoutes(alarmCloudHandler)

		// 创建 Auth Service 和 Handler
		authRepo := repository.NewPostgresAuthRepository(db)

		// 使用 sessionStore 创建 authService（这样 Login 会返回真实的 UUID token 而不是 stub）
		authService := service.NewAuthServiceWithSessionStore(authRepo, tenantsRepo, db, logger, sessionStore)
		authHandler := httpapi.NewAuthHandler(authService, logger)

		// 关联会话存储到 authHandler（供 middleware 读取会话）
		authHandler.SetSessionStore(sessionStore)

		router.RegisterAuthRoutes(authHandler)

		// 创建 Device Service 和 Handler（qinglanClient 已在上面创建）
		devicesRepo.SetLogger(logger) // 确保 logger 已设置（用于设备连接日志）
		deviceService := service.NewDeviceService(devicesRepo, cardSyncService, qinglanClient, card.NewReader(redisClient), db, logger)
		deviceHandler := httpapi.NewDeviceHandler(deviceService, logger)
		router.RegisterDeviceRoutes(deviceHandler)

		// 创建 DeviceStore Handler（直接使用 Repository，在线状态与 device_service 一致：优先 cardagg Redis）
		deviceStoreHandler := httpapi.NewDeviceStoreHandler(deviceStoreRepo, card.NewReader(redisClient), db, logger)
		router.RegisterDeviceStoreRoutes(deviceStoreHandler)

		// 创建 CardsRepository（供 RadarInstall 查卡片设备，以及后续 CardService 使用）
		cardsRepo := repository.NewPostgresCardsRepository(db)

		// 创建数据流订阅器（订阅 card:realtime:stream, card:status:stream；设备状态在 iot:event:stream）
		dataStreamSubscriber = subscriber.NewDataStreamSubscriber(redisClient, logger)
		initCtx := context.Background()
		if err := dataStreamSubscriber.Start(initCtx); err != nil {
			logger.Warn("Failed to start data stream subscriber", zap.Error(err))
		}
		if cardRealtimeSvc != nil {
			cardRealtimeSvc.SetSSEDependencies(dataStreamSubscriber)
			cardRealtimeSvc.SetGetCardStatusDeps(card.NewReader(redisClient), db)

			// 桥接 subscriber.CardStatusEvent → service.StatusEvent，激活 fan-out 通路
			statusBridgeCh := make(chan service.StatusEvent, 64)
			cardRealtimeSvc.SetStatusEventChan(statusBridgeCh)
			go func() {
				for ev := range dataStreamSubscriber.GetCardStatusEventChan() {
					select {
					case statusBridgeCh <- service.StatusEvent{CardID: ev.CardID, Data: ev.Data}:
					default:
						// drop if service channel full to avoid blocking subscriber
					}
				}
			}()
			cardRealtimeSvc.StartStatusFanout(initCtx)
		}

		// 启动数据流消费协程 (稍后在ctx定义后启动)
		// go subscribeDataStream(ctx, logger, redisClient, dataStreamSubscriber)

		// 创建 Radar Install Service 和 Handler（通过 wisefido-qinglan 与设备通信）
		radarInstall := service.NewRadarInstall(cfg, devicesRepo, cardsRepo, configVersionsRepo, unitsRepo, qinglanClient, logger)
		radarHandler := httpapi.NewRadarHandler(radarInstall, stub, kv, redisClient, logger)
		// 将 dataStreamSubscriber 传给 RadarHandler（供 SSE 推送使用）
		radarHandler.SetDataStreamSubscriber(dataStreamSubscriber)
		router.RegisterRadarRoutes(radarHandler)

		iotTSRepo := repository.NewPostgresIoTTimeSeriesRepository(db)
		trackPlaybackSvc := service.NewTrackPlaybackService(devicesRepo, iotTSRepo, logger)
		playbackHandler := httpapi.NewPlaybackHandler(trackPlaybackSvc, tenantsRepo, logger)
		router.RegisterPlaybackRoutes(playbackHandler)

		// 创建 Branch Service 和 Handler
		branchesRepo = repository.NewPostgresBranchesRepository(db)
		branchService := service.NewBranchService(branchesRepo, db, logger)
		branchesHandler := httpapi.NewBranchesHandler(branchService, db, logger)
		router.RegisterBranchesRoutes(branchesHandler)

		// 创建 Unit Service 和 Handler
		unitService := service.NewUnitService(unitsRepo, branchesRepo, residentsRepo, devicesRepo, db, cardSyncService, logger)
		unitHandler := httpapi.NewUnitHandler(unitService, logger)
		router.RegisterUnitRoutes(unitHandler)

		// 创建 User Service 和 Handler
		// usersRepo 已在上面创建 RoleService 时声明，这里直接使用
		// branchesRepo 已在上面创建 UnitService 时声明，这里直接使用
		userBranchesRepo := repository.NewPostgresUserBranchesRepository(db)
		userService := service.NewUserService(usersRepo, branchesRepo, userBranchesRepo, db, logger)
		userHandler := httpapi.NewUserHandler(userService, db, logger)
		router.RegisterUsersRoutes(userHandler)

		// 创建 ConfigPublisher（用于发送所有 config:* 消息）
		configPublisher := publisher.NewConfigPublisher(redisClient, logger)
		deviceService.SetConfigPublisher(configPublisher)

		// 创建 CardSyncService
		cardSyncService = service.NewCardSyncService(cardRepo, configPublisher, cardRealtimeSvc, logger)

		// 创建 DeviceMonitorSettings Service 和 Handler
		alarmDeviceRepo := repository.NewPostgresAlarmDeviceRepository(db)
		// 使用已创建的 alarmCloudRepo、configVersionsRepo（在 AlarmCloud Service 创建时已创建）
		// 使用已创建的 configPublisher（在上面创建）
		deviceMonitorSettingsService := service.NewDeviceMonitorSettingsService(
			alarmDeviceRepo,
			alarmCloudRepo,
			configVersionsRepo, // 使用已创建的 configVersionsRepo
			devicesRepo,
			deviceStoreRepo,
			db, // 添加 db 参数用于事务操作
			configPublisher,
			logger,
		)

		var sleepaceGateway *service.SleepaceGatewayClient
		if cfg.SleepaceGateway.APIBaseURL != "" {
			sleepaceGateway = service.NewSleepaceGatewayClient(cfg.SleepaceGateway.APIBaseURL, logger)
			if svc, ok := deviceMonitorSettingsService.(interface {
				SetSleepaceGatewayClient(client *service.SleepaceGatewayClient)
			}); ok {
				svc.SetSleepaceGatewayClient(sleepaceGateway)
			}
			if err := sleepaceGateway.SetPushType(context.Background()); err != nil {
				logger.Warn("Failed to set sleepace-service pushType via gateway", zap.Error(err))
			} else {
				logger.Info("sleepace-service pushType set (via gateway)")
			}
			if err := sleepaceGateway.Ping(context.Background()); err != nil {
				logger.Warn("Sleepace gateway unreachable at startup (sleepace-service/wisefido-sleepace may be down)",
					zap.String("api_base_url", cfg.SleepaceGateway.APIBaseURL),
					zap.Error(err))
			} else {
				logger.Info("Sleepace gateway connected",
					zap.String("api_base_url", cfg.SleepaceGateway.APIBaseURL))
			}
		} else {
			logger.Warn("Sleepace gateway client not initialized (SLEEPACE_GATEWAY_API_BASE_URL not set)")
		}
		deviceStoreService := service.NewDeviceStoreService(deviceStoreRepo, devicesRepo, unitsRepo, sleepaceGateway, logger)
		deviceStoreService.SetConfigPublisher(configPublisher)
		deviceStoreHandler.SetDeviceStoreService(deviceStoreService)

		// 设置 QinglanClient 到 DeviceMonitorSettingsService（用于下发雷达监控设置：工作模式、跌倒/呼吸心率参数）
		if svc, ok := deviceMonitorSettingsService.(interface {
			SetQinglanClient(client *service.QinglanClient)
		}); ok {
			svc.SetQinglanClient(qinglanClient)
			logger.Info("Qinglan client set for device monitor settings service")
		} else {
			logger.Warn("DeviceMonitorSettingsService does not support SetQinglanClient")
		}

		deviceMonitorSettingsHandler := httpapi.NewDeviceMonitorSettingsHandler(
			deviceMonitorSettingsService,
			usersRepo,        // 用于安全验证：验证 user_id 和 tenant_id 一致性
			userBranchesRepo, // 用于安全验证：验证 branch_id 与 user_id 一致性
			devicesRepo,      // 用于安全验证：验证 device_id 与 tenant_id 一致性
			unitsRepo,        // 用于安全验证：获取设备的 branch_id
			redisClient,
			db,
			card.NewReader(redisClient), // 设备在线状态（走 device_store/cardagg，不依赖 card）
			logger,
		)
		router.RegisterDeviceMonitorSettingsRoutes(deviceMonitorSettingsHandler)

		// 创建 AlarmEvent Service 和 Handler
		alarmEventsRepo := repository.NewPostgresAlarmEventsRepository(db)
		alarmEventService := service.NewAlarmEventService(
			alarmEventsRepo,
			devicesRepo,
			unitsRepo,
			usersRepo,
			db,
			redisClient,
			configPublisher,
			logger,
		)
		alarmEventHandler := httpapi.NewAlarmEventHandler(alarmEventService, logger)
		router.RegisterAlarmEventRoutes(alarmEventHandler)

		var apnsSender *notify.APNSSender
		if kid := strings.TrimSpace(os.Getenv("APNS_KEY_ID")); kid != "" {
			s, err := notify.NewAPNSSender(kid, os.Getenv("APNS_TEAM_ID"), os.Getenv("APNS_BUNDLE_ID"), os.Getenv("APNS_P8_KEY"))
			if err != nil {
				logger.Warn("APNS sender init failed", zap.Error(err))
			} else {
				apnsSender = s
			}
		}
		apnsDeviceSvc := service.NewAPNSDeviceService(db, apnsSender, logger)
		router.RegisterAPNSRoutes(httpapi.NewAPNSHandler(apnsDeviceSvc, logger))
		router.RegisterAlarmPushRoute(httpapi.NewAlarmPushHandler(db, apnsDeviceSvc, logger))

		// 创建 Sleepace Report Service 和 Handler
		sleepaceReportsRepo := repository.NewPostgresSleepaceReportsRepository(db)
		sleepaceReportService = service.NewSleepaceReportService(sleepaceReportsRepo, db, logger)
		if sleepaceGateway != nil {
			sleepaceGateway.SetSleepaceReportPersistence(sleepaceReportsRepo)
			if svc, ok := sleepaceReportService.(interface {
				SetSleepaceGatewayClient(client *service.SleepaceGatewayClient)
			}); ok {
				svc.SetSleepaceGatewayClient(sleepaceGateway)
			}
		}

		// 创建 Resident Service 和 Handler
		// residentsRepo 已在上面创建 CardStaticService 时声明，这里直接使用
		residentService := service.NewResidentService(residentsRepo, db, cardSyncService, logger)
		residentHandler := httpapi.NewResidentHandler(residentService, db, logger)
		router.RegisterResidentRoutes(residentHandler)

		sleepaceReportHandler := httpapi.NewSleepaceReportHandler(sleepaceReportService, db, logger)
		router.RegisterSleepaceReportRoutes(sleepaceReportHandler)

		// Rounds（Automatic Rounds 完成落库与审计列表）
		roundsRepo := repository.NewPostgresRoundsRepository(db)
		roundsService := service.NewRoundsService(roundsRepo, logger)
		roundsHandler := httpapi.NewRoundsHandler(roundsService, logger)
		router.RegisterRoundsRoutes(roundsHandler)

		// Sleepace MQTT 由 wisefido-sleepace 消费；本进程经 SleepaceGatewayClient 访问厂家 HTTP。
	} else {
		// DB 未就绪：内存租户 + auth，Admin 侧 units/devices 等为 nil（功能受限）
		tenantsRepo = repository.NewMemoryTenantsRepo()
		// Seed "System" tenant for SystemAdmin login in dev (matches httpapi.systemTenantID)
		systemTenant := &domain.Tenant{
			TenantID:   "00000000-0000-0000-0000-000000000001",
			TenantName: "System",
			Domain:     "system.local",
			Status:     "active",
		}
		_, _ = tenantsRepo.CreateTenant(context.Background(), systemTenant)
		// Seed SystemAdmin account
		_ = authStore.UpsertUser("00000000-0000-0000-0000-000000000001", "sysadmin", "SystemAdmin", "ChangeMe123!")
		// 注意：StubHandler 仍使用旧的 TenantsRepo 接口，需要适配器或更新
		// 暂时传递 nil，StubHandler 会处理
		stub = httpapi.NewStubHandler(nil, authStore, nil)
		// AdminAPI units/devices 等为 nil，回退 stub
		admin = httpapi.NewAdminAPI(stub, logger)
	}
	router.RegisterAdminUnitDeviceRoutes(admin)
	// 如果 DB 启用，传入 BranchesRepository 以便创建 tenant 时自动创建默认 branch
	var branchesRepoForTenants repository.BranchesRepository
	if db != nil {
		branchesRepoForTenants = repository.NewPostgresBranchesRepository(db)
		router.RegisterAdminTenantRoutes(httpapi.NewTenantsHandlerWithLogger(tenantsRepo, branchesRepoForTenants, authStore, db, logger))
	} else {
		router.RegisterAdminTenantRoutes(httpapi.NewTenantsHandler(tenantsRepo, branchesRepoForTenants, authStore, db))
	}
	router.RegisterStubRoutes(stub)

	// 注册 Doctor 路由（健康检查和诊断功能）
	doctorEnabled := os.Getenv("DOCTOR_ENABLED")
	if doctorEnabled != "false" {
		doctor := httpapi.NewDoctorHandler(db, redisClient, tenantsRepo, logger)
		// 启用 pprof（如果配置了）
		if os.Getenv("DOCTOR_PPROF") == "true" {
			doctor.EnablePprof(true)
		}
		router.RegisterDoctorRoutes(doctor)
	}

	// 初始化认证中间件
	// AuthMiddleware 从 Authorization header 提取 Bearer token，
	// 查询 sessionStore 获取用户信息，然后注入到请求 header（X-User-Id, X-Tenant-Id, X-User-Type, X-User-Role）
	authMiddlewareEnabled := os.Getenv("AUTH_MIDDLEWARE_ENABLED") != "false" // 默认启用
	authMiddleware := httpapi.NewAuthMiddleware(httpapi.AuthMiddlewareConfig{
		Store:   sessionStore,
		Skipped: httpapi.DefaultSkippedPaths,
		Logger:  logger,
		Enabled: authMiddlewareEnabled,
	})

	// 将 router 通过 authMiddleware 包装，使所有路由（非跳过清单）都需要有效的 token
	wrappedHandler := authMiddleware.Wrap(router)

	// 创建  HTTP 服务器
	srv := service.NewServer(cfg.HTTP.Addr, wrappedHandler, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动数据流消费协程（在ctx定义后）
	if dataStreamSubscriber != nil {
		go subscribeDataStream(ctx, logger, redisClient, dataStreamSubscriber)
	}

	// 启动时全量检查并更新卡片（如果 DB 和 cardSyncService 可用）
	if db != nil && cardSyncService != nil {
		go func() {
			// 延迟启动，等待服务完全初始化
			time.Sleep(2 * time.Second)

			logger.Info("Starting full card check/update on service startup")

			// 初始化以当前 unit 为准重建卡片，不基于库里已有卡（重启后 unit 可能已变化）
			if err := cardSyncService.ClearAllCards(ctx); err != nil {
				logger.Warn("Failed to clear all cards before sync", zap.Error(err))
				return
			}

			// 获取所有活跃租户
			tenants, _, err := tenantsRepo.ListTenants(ctx, repository.TenantFilters{
				Status: "active",
			}, 1, 1000) // 假设不超过1000个租户

			if err != nil {
				logger.Warn("Failed to list tenants for card check",
					zap.Error(err),
				)
				return
			}

			logger.Info("Found tenants for card check",
				zap.Int("tenant_count", len(tenants)),
			)

			// 统计信息
			totalStats := struct {
				ExistingCount  int
				DeletedCount   int
				CreatedCount   int
				UpdatedCount   int
				UnchangedCount int
			}{}
			successCount := 0
			errorCount := 0
			totalUnits := 0

			// 为每个租户的所有单元创建/更新卡片
			for _, tenant := range tenants {
				// 获取该租户的所有单元
				unitIDs, err := cardRepo.GetAllUnits(tenant.TenantID)
				if err != nil {
					logger.Warn("Failed to get units for tenant",
						zap.String("tenant_id", tenant.TenantID),
						zap.Error(err),
					)
					errorCount++
					continue
				}

				totalUnits += len(unitIDs)

				// 为每个单元创建/更新卡片
				for _, unitID := range unitIDs {
					select {
					case <-ctx.Done():
						logger.Info("Card check interrupted by context cancellation")
						return
					default:
						stats, err := cardSyncService.CreateCardsForUnit(ctx, tenant.TenantID, unitID)
						if err != nil {
							logger.Error("Failed to create cards for unit",
								zap.String("tenant_id", tenant.TenantID),
								zap.String("unit_id", unitID),
								zap.Error(err),
							)
							errorCount++
						} else {
							successCount++
							if stats != nil {
								totalStats.ExistingCount += stats.ExistingCount
								totalStats.DeletedCount += stats.DeletedCount
								totalStats.CreatedCount += stats.CreatedCount
								totalStats.UpdatedCount += stats.UpdatedCount
								totalStats.UnchangedCount += stats.UpdatedCount
							}
						}
					}
				}
			}

			// 启动时按 alarm_events 重算并写回 cards（unhandled_alarm_*、pop_alarm_*），与 alarm_events 一致
			// 在 CreateCardsForUnit 之后执行，直接调用内部 CardSyncService
			recalcOk, recalcFail, err := cardSyncService.RecalcAllCardsAlarmState(ctx, db)
			if err != nil {
				logger.Warn("Failed to recalc all cards alarm state", zap.Error(err))
			} else {
				logger.Info("Cards alarm state recalc on startup", zap.Int("ok", recalcOk), zap.Int("fail", recalcFail))
			}

			// 输出统计信息到 stdout 和日志
			updateCount := totalStats.DeletedCount + totalStats.CreatedCount + totalStats.UpdatedCount
			summaryMsg := fmt.Sprintf(
				"\n=== Card Check/Update Statistics (Startup) ===\n"+
					"Tenants processed: %d\n"+
					"Units processed: %d (success: %d, failed: %d)\n"+
					"Existing card count: %d\n"+
					"Updated card count: %d (deleted: %d, created: %d, content updated: %d)\n"+
					"Unchanged cards: %d\n"+
					"==========================================\n",
				len(tenants),
				totalUnits,
				successCount,
				errorCount,
				totalStats.ExistingCount,
				updateCount,
				totalStats.DeletedCount,
				totalStats.CreatedCount,
				totalStats.UpdatedCount,
				totalStats.UnchangedCount,
			)

			// 输出到 stdout
			os.Stdout.WriteString(summaryMsg)

			// 同时记录到日志
			logger.Info("Completed full card check/update on startup",
				zap.Int("tenant_count", len(tenants)),
				zap.Int("unit_count", totalUnits),
				zap.Int("success_count", successCount),
				zap.Int("error_count", errorCount),
				zap.Int("existing_count", totalStats.ExistingCount),
				zap.Int("updated_count", updateCount),
				zap.Int("deleted_count", totalStats.DeletedCount),
				zap.Int("created_count", totalStats.CreatedCount),
				zap.Int("content_updated_count", totalStats.UpdatedCount),
				zap.Int("unchanged_count", totalStats.UnchangedCount),
			)
		}()
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigCh:
		logger.Info("Received shutdown signal")
		cancel()
	case err := <-errCh:
		if err != nil {
			logger.Error("Server stopped with error", zap.Error(err))
		} else {
			logger.Info("Server stopped")
		}
		cancel()
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()
	_ = srv.Stop(shutdownCtx)
	_ = redisClient.Close()
	if db != nil {
		_ = db.Close()
	}
}

// subscribeDataStream 订阅数据流（card:realtime:stream, card:status:stream）
func subscribeDataStream(ctx context.Context, logger *zap.Logger, redisClient *redis.Client, dataStreamSubscriber *subscriber.DataStreamSubscriber) {
	streams := []string{
		"card:realtime:stream",
		"card:status:stream",
	}

	consumerGroup := "wisefido-data-consumer"
	consumerName := "wisefido-data-consumer-1"

	logger.Info("Starting data stream consumer",
		zap.Strings("streams", streams),
		zap.String("consumer_group", consumerGroup),
	)

	// 为每个流创建消费者组（已存在则 BUSYGROUP，视为正常）
	for _, streamName := range streams {
		if err := rediscommon.CreateConsumerGroup(ctx, redisClient, streamName, consumerGroup); err != nil &&
			err.Error() != "BUSYGROUP Consumer Group name already exists" {
			logger.Warn("Failed to create consumer group",
				zap.String("stream", streamName),
				zap.String("consumer_group", consumerGroup),
				zap.Error(err),
			)
		}
	}

	for {
		select {
		case <-ctx.Done():
			logger.Info("Data stream consumer stopped")
			return
		default:
			// 使用 XREADGROUP 从 Consumer Group 读取消息
			msgs, err := rediscommon.ReadFromMultipleStreamsWithBlock(ctx, redisClient, streams, consumerGroup, consumerName, 10, 2*time.Second)
			if err != nil {
				errStr := err.Error()
				if strings.Contains(errStr, "NOGROUP") || strings.Contains(errStr, "no such key") {
					for _, streamName := range streams {
						if createErr := rediscommon.CreateConsumerGroup(ctx, redisClient, streamName, consumerGroup); createErr != nil && createErr.Error() != "BUSYGROUP Consumer Group name already exists" {
							logger.Warn("[SUB] Recreate consumer group after stream missing", zap.String("stream", streamName), zap.Error(createErr))
						}
					}
				}
				logger.Warn("[SUB] Failed to read from data streams",
					zap.Error(err),
				)
				continue
			}

			for _, msg := range msgs {
				var handlerErr error

				// 转换 StreamMessage 为 redis.XMessage
				xMsg := redis.XMessage{
					ID:     msg.ID,
					Values: msg.Values,
				}

				switch msg.Stream {
				case "card:realtime:stream":
					handlerErr = dataStreamSubscriber.HandleCardRealtimeMessage(ctx, xMsg)
				case "card:status:stream":
					handlerErr = dataStreamSubscriber.HandleCardStatusMessage(ctx, xMsg)
				default:
					logger.Warn("Unknown stream",
						zap.String("stream", msg.Stream),
					)
				}

				if handlerErr != nil {
					logger.Error("Failed to handle message",
						zap.String("stream", msg.Stream),
						zap.String("message_id", msg.ID),
						zap.Error(handlerErr),
					)
				}

				// 处理完消息后确认
				if err := redisClient.XAck(ctx, msg.Stream, consumerGroup, msg.ID).Err(); err != nil {
					logger.Warn("Failed to acknowledge message",
						zap.String("stream", msg.Stream),
						zap.String("message_id", msg.ID),
						zap.Error(err),
					)
				}
			}
		}
	}
}
