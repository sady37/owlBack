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
	phipkg "wisefido-data/internal/phi"
	httpapi "wisefido-data/internal/http"
	"wisefido-data/internal/notify"
	"wisefido-data/internal/publisher"
	"wisefido-data/internal/repository"
	scopepkg "wisefido-data/internal/scope"
	"wisefido-data/internal/service"
	"wisefido-data/internal/store"
	"wisefido-data/internal/subscriber"

	"owl-common/card"
	"owl-common/database"
	"owl-common/ddns"
	"owl-common/ipam"
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

	// DDNS client — 一次性装配 (BIND_HOST 缺失 = nil)；同时供 registerSpatialV2 + ResidentService Phase F 写卡路径用
	ddnsClient := initDDNSClient(logger)

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

		// 创建 Role Service（仅创建，不保留变量）
		roleRepo := repository.NewPostgresRolesRepository(db)
		usersRepo = repository.NewPostgresUsersRepository(db)
		_ = service.NewRoleService(roleRepo, usersRepo, logger)

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

		// 主动刷新卡片设备状态（前端 refresh 按钮 → 推 iot:probe:device:stream → qinglan/sleepace 立即 health_check）
		cardRefreshHandler := httpapi.NewCardRefreshHandler(db, redisClient, logger)
		router.RegisterCardRefreshRoutes(cardRefreshHandler)

	}

	// 提升作用域：scheduler 在外部需要引用（见下方 Start 调用），避免 if db != nil 块内声明出不来
	var sleepaceGateway *service.SleepaceGatewayClient
	// 同上：/internal/sleepace/device/* resync 端点需要在 if db != nil 块外引用
	var deviceMonitorSettingsService service.DeviceMonitorSettingsService

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

		// 创建 Role Service 和 Handler
		roleRepo := repository.NewPostgresRolesRepository(db)
		usersRepo := repository.NewPostgresUsersRepository(db)
		roleService := service.NewRoleService(roleRepo, usersRepo, logger)
		rolesHandler := httpapi.NewRolesHandler(roleService, logger)
		router.RegisterRolesRoutes(rolesHandler)

		// 创建 Card Repository（DB card 数据操作）
		cardRepo = repository.NewPostgresCardRepository(db, logger)

		// 创建 AlarmCloud Service 和 Handler
		alarmCloudRepo := repository.NewPostgresAlarmCloudRepository(db)
		configVersionsRepo := repository.NewPostgresConfigVersionsRepository(db)
		alarmCloudService := service.NewAlarmCloudService(alarmCloudRepo, configVersionsRepo, db, logger)
		alarmCloudHandler := httpapi.NewAlarmCloudHandler(alarmCloudService, logger)
		router.RegisterAlarmCloudRoutes(alarmCloudHandler)

		// 创建 Auth Handler (owl_v2 IPv6-native auth)
		authHandler := httpapi.NewAuthHandler(db, sessionStore, logger)
		router.RegisterAuthRoutes(authHandler)

		// 创建 Device Service 和 Handler（qinglanClient 已在上面创建）
		devicesRepo.SetLogger(logger) // 确保 logger 已设置（用于设备连接日志）
		deviceService := service.NewDeviceService(devicesRepo, qinglanClient, card.NewReader(redisClient), db, logger)
		deviceHandler := httpapi.NewDeviceHandler(deviceService, logger)
		deviceHandler.SetDB(db)
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
		radarInstall := service.NewRadarInstall(cfg, db, devicesRepo, cardsRepo, configVersionsRepo, unitsRepo, qinglanClient, logger)
		radarHandler := httpapi.NewRadarHandler(radarInstall, stub, kv, redisClient, logger)
		// 将 dataStreamSubscriber 传给 RadarHandler（供 SSE 推送使用）
		radarHandler.SetDataStreamSubscriber(dataStreamSubscriber)
		router.RegisterRadarRoutes(radarHandler)

		monitorPlaybackRepo := repository.NewPostgresMonitorPlaybackRepository(db)
		trackPlaybackSvc := service.NewTrackPlaybackService(devicesRepo, monitorPlaybackRepo, logger)
		playbackHandler := httpapi.NewPlaybackHandler(trackPlaybackSvc, tenantsRepo, db, logger)
		router.RegisterPlaybackRoutes(playbackHandler)

		// 创建 Branch Service 和 Handler
		branchesRepo = repository.NewPostgresBranchesRepository(db)
		branchService := service.NewBranchService(branchesRepo, db, logger)
		branchesHandler := httpapi.NewBranchesHandler(branchService, db, logger)
		router.RegisterBranchesRoutes(branchesHandler)

		// 创建 Unit Service 和 Handler（v2: residentsRepo 解耦走 raw SQL）
		unitService := service.NewUnitService(unitsRepo, branchesRepo, devicesRepo, db, logger)
		unitHandler := httpapi.NewUnitHandler(unitService, logger)
		router.RegisterUnitRoutes(unitHandler)


		// 创建 User Service 和 Handler
		// usersRepo 已在上面创建 RoleService 时声明，这里直接使用
		// branchesRepo 已在上面创建 UnitService 时声明，这里直接使用
		userBranchesRepo := repository.NewPostgresUserBranchesRepository(db)
		userService := service.NewUserService(usersRepo, branchesRepo, userBranchesRepo, db, logger)
		userHandler := httpapi.NewUserHandler(userService, db, logger)
		router.RegisterUsersRoutes(userHandler)

		// CareTeam（护士站）— 自包含 raw SQL handler，不依赖 service/repo
		careTeamsHandler := httpapi.NewCareTeamsHandler(db, logger)
		router.RegisterCareTeamsRoutes(careTeamsHandler)

		// Resident (Forward Design)— 独立 /admin/api/v2/residents 路由
		// 注意：复用 line 126 已创建的外层 residentsRepo（用 = 而非 :=，避免变量 shadowing
		// 让 PHI cryptor 注入到错误的实例上 → contacts/PHI 路径假性 nil）
		residentsRepo.SetLogger(logger)

		// 初始化 PHI 加密（K 服务模式）—— 必须在 residentsRepo 拿到 logger 之后、service 创建之前注入
		kmsSocket := os.Getenv("KMS_SOCKET")
		masterPin := os.Getenv("MASTER_PIN")
		var phiCryptor *phipkg.PHICryptor
		if kmsSocket != "" && masterPin != "" {
			keyStore := phipkg.NewTenantKeyStore(kmsSocket, masterPin)
			phiCryptor = phipkg.NewPHICryptor(keyStore)
			residentsRepo.SetPHICryptor(phiCryptor)
			logger.Info("PHI encryption enabled", zap.String("kms_socket", kmsSocket))
		} else {
			logger.Warn("PHI encryption NOT enabled — set KMS_SOCKET and MASTER_PIN")
		}

		residentSvc := service.NewResidentService(residentsRepo)
		residentSvc.SetLogger(logger)
		residentHandler := httpapi.NewResidentHandler(db, residentSvc, logger)
		router.RegisterResidentRoutes(residentHandler)

		// 创建 ConfigPublisher（用于发送所有 config:* 消息）
		configPublisher := publisher.NewConfigPublisher(redisClient, logger)
		deviceService.SetConfigPublisher(configPublisher)

		// 创建 CardSyncService — v2 Card 表唯一写入入口（ReconcileCards）
		cardSyncService = service.NewCardSyncService(cardRepo, configPublisher, cardRealtimeSvc, logger)
		// Phase F+: 启动 reconcile 路径要 db / DDNS / owlDomain 装配
		cardSyncService.SetReconcileDeps(db, ddnsClient, getenvDefault("OWL_DOMAIN", "owl."))
		service.InitGlobalCardSync(cardSyncService)

		// Phase F — ResidentService 接 cards 写入路径（兼容旧 syncCardForResident，渐进迁移期间保留）
		residentSvc.SetCardDeps(cardRepo, configPublisher, ddnsClient, getenvDefault("OWL_DOMAIN", "owl."), cardSyncService)

		// v2 收敛 — Repository hook：写表后自动 reconcile cards，免去各 service 层"记得调"
		// 设计：repository 不依赖 service，只持有 callback；闭包桥接 cardSync.ReconcileCards
		residentsRepo.SetOnResidentUnitChange(func(ctx context.Context, unitScope string) {
			if err := cardSyncService.ReconcileCards(ctx, unitScope); err != nil {
				logger.Warn("residents hook → ReconcileCards failed",
					zap.String("scope", unitScope), zap.Error(err))
			}
		})
		devicesRepo.SetOnDeviceChange(func(ctx context.Context, unitScope string) {
			if err := cardSyncService.ReconcileCards(ctx, unitScope); err != nil {
				logger.Warn("devices hook → ReconcileCards failed",
					zap.String("scope", unitScope), zap.Error(err))
			}
		})

		// v2 Phase 1b: slim DeviceMonitorSettingsV2Service — backed by spatial_config (device /128) + tenant snapshot fallback.
		// v1 service 依赖 alarmDeviceRepo / configVersionsRepo / deviceStoreRepo / residentsRepo 等 v2 已删/重构的表，
		// 整段跳过；slim v2 仅实现 Get/Update/GetDefault/CheckOnline 4 个核心方法，
		// 其它 9 个（OTA / firmware / resync）返回 NotImplemented 错误，等具体场景再 v2 化。
		deviceMonitorSettingsService = service.NewDeviceMonitorSettingsV2Service(db, alarmCloudRepo, logger)

		var intervalScheduler *service.SleepaceIntervalScheduler
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

			// SleepaceIntervalScheduler — 按 rest/nap 时段自动切换 realtime_interval
			// 进入窗口前 10min 切 2s 高频；离开窗口后 2min 切回 10s。
			// 用户在 device monitor settings 设了 interval >= 10 视为明确选择低频，scheduler 跳过。
			intervalScheduler = service.NewSleepaceIntervalScheduler(db, redisClient, sleepaceGateway, alarmCloudRepo, logger)
			go intervalScheduler.Start(context.Background())
		} else {
			logger.Warn("Sleepace gateway client not initialized (SLEEPACE_GATEWAY_API_BASE_URL not set)")
		}
		deviceStoreService := service.NewDeviceStoreService(deviceStoreRepo, devicesRepo, unitsRepo, sleepaceGateway, logger)
		deviceStoreService.SetConfigPublisher(configPublisher)
		// 让 sleepad 厂家 bind 成功后立即清掉 scheduler 的 unbound backoff cache（不等 1h TTL），
		// 下次 60s tick 即可重新评估并 push interval。
		if intervalScheduler != nil {
			deviceStoreService.SetOnSleepadVendorBound(func(deviceID string) {
				intervalScheduler.InvalidateUnbound(context.Background(), deviceID)
			})
		}
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
		apnsAllowed := service.NewAllowedCardIDsProvider(kv, usersRepo, residentsRepo, db, cardRepo, logger)
		apnsPopCounter := service.NewStaffPopPendingCounter(db, apnsAllowed)
		apnsDeviceSvc := service.NewAPNSDeviceService(db, apnsSender, logger, apnsPopCounter)
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

		sleepaceReportHandler := httpapi.NewSleepaceReportHandler(sleepaceReportService, db, logger)

		// 创建 Sleepace Report Service
		router.RegisterSleepaceReportRoutes(sleepaceReportHandler)

		// SleepaceReportBackfillScheduler — 后台周期 backfill vendor 缺失 reports，
		// 替代原来 list/detail 入口同步等 vendor 阻塞 5-30s 的问题。
		// 仅 sleepaceGateway 可用 + reportService 已注入 gateway 时才启 scheduler。
		if sleepaceGateway != nil {
			backfillScheduler := service.NewSleepaceReportBackfillScheduler(db, sleepaceReportService, logger)
			go backfillScheduler.Start(context.Background())
		}

		// Rounds（Automatic Rounds 完成落库与审计列表）
		roundsRepo := repository.NewPostgresRoundsRepository(db)
		roundsService := service.NewRoundsService(roundsRepo, logger)
		roundsHandler := httpapi.NewRoundsHandler(roundsService, db, logger)
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
	var tenantsHandler *httpapi.TenantsHandler
	if db != nil {
		branchesRepoForTenants = repository.NewPostgresBranchesRepository(db)
		tenantsHandler = httpapi.NewTenantsHandlerWithLogger(tenantsRepo, branchesRepoForTenants, authStore, db, logger)
	} else {
		tenantsHandler = httpapi.NewTenantsHandler(tenantsRepo, branchesRepoForTenants, authStore, db)
	}
	router.RegisterAdminTenantRoutes(tenantsHandler)
	router.RegisterStubRoutes(stub)

	// 注册 v2 spatial IPAM/DDNS API（owl_v2 IPv6 体系）— 独立于 v1
	// owl-common/ipam.PGBackend 直接走 dbv2 spatial 表 + 可选 kea audit
	if db != nil {
		registerSpatialV2(router, db, ddnsClient, logger)
	}

	// 注册内部 baseline API（供 gate/cardagg 查询 device→card 映射，/internal/ 跳过 auth）
	if db != nil {
		baselineHandler := httpapi.NewBaselineHandler(card.NewCardDB(db), logger)
		router.RegisterBaselineRoutes(baselineHandler)
	}

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
	// D-001b：tenantStatusCache 5min in-memory cache 用于每请求 tenant.status 检查；
	// admin_tenants_handlers 改 status 后通过 SetTenantCache(...) 主动 invalidate。
	authMiddlewareEnabled := os.Getenv("AUTH_MIDDLEWARE_ENABLED") != "false" // 默认启用
	tenantStatusCache := httpapi.NewTenantStatusCache(db)
	authMiddleware := httpapi.NewAuthMiddleware(httpapi.AuthMiddlewareConfig{
		Store:       sessionStore,
		Skipped:     httpapi.DefaultSkippedPaths,
		Logger:      logger,
		Enabled:     authMiddlewareEnabled,
		TenantCache: tenantStatusCache,
	})
	// 把 cache 注入到 TenantsHandler，让其改 tenant 后自动 invalidate
	if tenantsHandler != nil {
		tenantsHandler.SetTenantStatusCache(tenantStatusCache)
	}

	// 将 router 通过 authMiddleware 包装，使所有路由（非跳过清单）都需要有效的 token
	// 然后再叠加 scope.Middleware：从 X-User-* headers 解析 ScopeContext 并注入 ctx，
	// 下游 handler/service 用 scope.FromContext(ctx) 拿 user/tenant/branch/role 4 维 scope。
	wrappedHandler := authMiddleware.Wrap(scopepkg.Middleware(db)(router))

	// 创建  HTTP 服务器
	srv := service.NewServer(cfg.HTTP.Addr, wrappedHandler, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动数据流消费协程（在ctx定义后）
	if dataStreamSubscriber != nil {
		go subscribeDataStream(ctx, logger, redisClient, dataStreamSubscriber)
	}

	// Sleepad 实时上报配置（interval=2 + realtimeMode=1）走 bind 流程自动下发，
	// 见 service.DeviceStoreService.applyDefaultSleepadRealtime。批量补刷由 cmd/sleepace-set-realtime -apply-all 处理。

	// --- 旧 scheduler 停用 ---
	// 原本计划每日 Denver 13:00 重算 reportUploadTime 对应到 LA 服务器 OS 时钟整点。
	// 基于错误假设（以为厂家用"服务器 OS 时钟"触发报告，实际是"设备本地时间"）。
	// 代码保留在 service/sleepace_report_time_scheduler.go 供参考/回滚。

	// --- DST 脚本触发入口 ---
	// 两个 internal endpoint，外部 DST 脚本按 device_id 调用，service 内部自己查 effective 值下发。
	//   POST /internal/sleepace/device/{device_id}/resync-timezone    → bind with effective timezone + gender/age
	//   POST /internal/sleepace/device/{device_id}/resync-report-time → setReportUploadTime with effective hour
	//   POST /internal/sleepace/device/{device_id}/upgrade            → 触发厂家 OTA（body {"version":"6.89"}）
	// 同样路径在 /sleepace/api/v1/sleepace/device/* 下挂一份给前端走 nginx /sleepace/api/ 转发使用（admin UI）。
	if deviceMonitorSettingsService != nil {
		resyncHandler := httpapi.NewSleepaceResyncHandler(deviceMonitorSettingsService, db, logger)
		router.Handle("/internal/sleepace/device/", resyncHandler.Dispatch)
		router.Handle("/sleepace/api/v1/sleepace/device/", resyncHandler.Dispatch)

		// 厂家固件库管理（uploadFile/delete/deviceVersions）。
		//   POST /sleepace/api/v1/sleepace/firmware/upload    multipart "file"
		//   POST /sleepace/api/v1/sleepace/firmware/delete    body {device_type, device_version}
		//   GET  /sleepace/api/v1/sleepace/firmware/versions  channel 由 wisefido-sleepace 注入
		firmwareHandler := httpapi.NewSleepaceFirmwareHandler(deviceMonitorSettingsService, logger)
		router.Handle("/internal/sleepace/firmware/", firmwareHandler.Dispatch)
		router.Handle("/sleepace/api/v1/sleepace/firmware/", firmwareHandler.Dispatch)
	}

	// 启动时全量检查并更新卡片（如果 DB 和 cardSyncService 可用）
	if db != nil && cardSyncService != nil {
		go func() {
			// 延迟启动，等待服务完全初始化
			time.Sleep(2 * time.Second)

			logger.Info("Starting full card check/update on service startup")

			// 重启时裁剪所有 stream，丢弃积压的旧消息
			trimResults := rediscommon.TrimAllStreamsToNow(ctx, redisClient)
			for stream, result := range trimResults {
				logger.Info("stream trim on startup", zap.String("stream", stream), zap.String("result", result))
			}
			// 通知所有下游服务：stream 已裁剪，需回 data 重查映射
			if err := cardSyncService.PublishConfigCardReset(ctx); err != nil {
				logger.Warn("Failed to publish configCard reset", zap.Error(err))
			}

			// card_id 已与底层物理实体 UUID 一体化（ActiveBedCard.card_id=bed_id, UnitCard.card_id=unit_id,
			// DeviceCard.card_id=device_id），CreateCard 改为 idempotent UPSERT。重启时 unit/bed/device
			// 不变 → 同 card_id 走 UPDATE 路径，业务表里现存的 card_id 引用全部仍有效。
			//
			// 启动 orphan cleanup：物理锚点（bed/unit/device）已不存在的 card 行兜底删除，
			// 同时通过 emitCardChange 通知 cardagg 清 Redis card:state:* 防孤儿键 leak。
			if cleaned, err := cardSyncService.CleanupOrphanCards(ctx); err != nil {
				logger.Warn("CleanupOrphanCards failed", zap.Error(err))
			} else if cleaned > 0 {
				logger.Info("Orphan cards cleaned on startup", zap.Int("count", cleaned))
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

			// v2: 每个 tenant 单次 ReconcileCards 扫整 /48
			// 旧"先 CleanupOrphanCards 再 CreateCardsForUnit"流程统一到 ReconcileCards：
			//   - device monitor-on 才建卡（不再 create→cleanup 来回）
			//   - resident_id 用 LPM resident_unit (masklen ≥ 80) 反查
			for _, tenant := range tenants {
				select {
				case <-ctx.Done():
					logger.Info("Card check interrupted by context cancellation")
					return
				default:
					if err := cardSyncService.ReconcileCards(ctx, tenant.TenantID); err != nil {
						logger.Error("Failed to reconcile cards for tenant",
							zap.String("tenant_id", tenant.TenantID),
							zap.Error(err),
						)
						errorCount++
					} else {
						successCount++
					}
				}
			}
			_ = totalUnits // 旧统计字段 — ReconcileCards 自己已在 log 里输出
			_ = totalStats

			// v2: DeviceCard 概念已退役（改为 card_type='device' /128 极罕见）；
			// 孤儿清理由 CleanupOrphanCards 统一处理（v2 LPM 反查无 device 的 card）。
			_ = cardRepo // keep alive for downstream references

			// 为未绑定 card 的设备创建 DeviceCard（card_id = device_id）
			totalDeviceCards := 0
			for _, tenant := range tenants {
				dcCount, dcErr := cardSyncService.SyncDeviceCards(ctx, tenant.TenantID)
				if dcErr != nil {
					logger.Warn("Failed to sync device cards",
						zap.String("tenant_id", tenant.TenantID),
						zap.Error(dcErr),
					)
				} else {
					totalDeviceCards += dcCount
				}
			}
			if totalDeviceCards > 0 {
				logger.Info("DeviceCards synced on startup", zap.Int("created", totalDeviceCards))
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

// registerSpatialV2 wire up owl_v2 IPAM/DDNS REST API on /admin/api/v2/spatial/...
//
// 配置（环境变量；缺失则跳过相应组件）：
//   KEA_CTRL_URL  http://localhost:8000   (空 = 不接 kea audit，仅 PG)
//   KEA_USER      owl
//   KEA_PASSWORD  owl-kea-dev-2026
//   BIND_HOST     127.0.0.1                (空 = 不推 DDNS)
//   BIND_PORT     5353
//   DDNS_KEY_NAME ddns-update
//   DDNS_ALGO     hmac-sha256
//   DDNS_TSIG_SECRET  base64
//   OWL_DOMAIN    owl.
func registerSpatialV2(router *httpapi.Router, db *sql.DB, ddnsClient *ddns.Client, logger *zap.Logger) {
	// kea audit (可选)
	var keaClient *ipam.KeaClient
	if url := os.Getenv("KEA_CTRL_URL"); url != "" {
		keaClient = ipam.NewKeaClient(url,
			os.Getenv("KEA_USER"),
			os.Getenv("KEA_PASSWORD"))
		// 启动期 ping 一下 kea (失败不阻断启动，仅警告)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if v, err := keaClient.VersionGet(ctx); err != nil {
			logger.Warn("v2 spatial: kea ctrl-agent not reachable; PG IPAM still works without audit",
				zap.String("url", url), zap.Error(err))
			keaClient = nil
		} else {
			logger.Info("v2 spatial: kea audit enabled", zap.String("kea_version", v))
		}
	}
	backend := ipam.NewPGBackendWithKea(db, keaClient)

	handler := httpapi.NewSpatialV2Handler(backend, db, ddnsClient, logger)
	router.RegisterSpatialV2Routes(handler)
	logger.Info("v2 spatial: API registered at /admin/api/v2/spatial/*")
}

// initDDNSClient 启动期一次性装配 DDNS client（BIND_HOST 缺失 → 返回 nil 静默跳过）。
// 同时供 registerSpatialV2 + ResidentService Phase F 写卡路径复用。
func initDDNSClient(logger *zap.Logger) *ddns.Client {
	host := os.Getenv("BIND_HOST")
	if host == "" {
		logger.Info("v2 spatial: DDNS disabled (BIND_HOST unset)")
		return nil
	}
	port := 53
	if p := os.Getenv("BIND_PORT"); p != "" {
		fmt.Sscanf(p, "%d", &port)
	}
	c, err := ddns.New(ddns.Config{
		Server:    host,
		Port:      port,
		KeyName:   getenvDefault("DDNS_KEY_NAME", getenvDefault("DDNS_TSIG_NAME", "ddns-update")),
		Algorithm: getenvDefault("DDNS_ALGO", getenvDefault("DDNS_TSIG_ALGORITHM", "hmac-sha256")),
		Secret:    os.Getenv("DDNS_TSIG_SECRET"),
		OwlDomain: getenvDefault("OWL_DOMAIN", "owl."),
	})
	if err != nil {
		logger.Warn("v2 spatial: DDNS client init failed", zap.Error(err))
		return nil
	}
	logger.Info("v2 spatial: DDNS enabled", zap.String("bind", host), zap.Int("port", port))
	return c
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
