package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wisefido-data/internal/cache"
	cardcreator "wisefido-data/internal/card"
	"wisefido-data/internal/config"
	"wisefido-data/internal/domain"
	httpapi "wisefido-data/internal/http"
	"wisefido-data/internal/notifier"
	"wisefido-data/internal/publisher"
	"wisefido-data/internal/repository"
	"wisefido-data/internal/service"
	"wisefido-data/internal/store"

	"owl-common/database"
	logpkg "owl-common/logger"

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
	var cardCreator *cardcreator.CardCreator
	var cardSyncService *service.CardSyncService
	var devicesRepo *repository.PostgresDevicesRepository
	var unitsRepo *repository.PostgresUnitsRepository
	var deviceStoreRepo *repository.PostgresDeviceStoreRepository
	var residentsRepo *repository.PostgresResidentsRepository
	var branchesRepo *repository.PostgresBranchesRepository
	var userCardsCache *cache.UserCardsCache
	var usersRepo *repository.PostgresUsersRepository
	var tenantResolver *repository.PostgresTenantResolver
	// Stub depends on tenantsRepo + authStore (used by /auth/api/v1/institutions/search + /auth/api/v1/login)
	stub := httpapi.NewStubHandler(nil, authStore, nil)
	// Always register admin routes; if DB is not available, AdminAPI will fall back to stub (no 404).
	admin := httpapi.NewAdminAPI(nil, nil, nil, nil, stub, logger, nil)
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
		tenantResolver = repository.NewPostgresTenantResolver(db)
		tenantsRepo = repository.NewPostgresTenantsRepository(db)
		// Note: StubHandler still uses TenantsRepo (old interface), but we need TenantsRepository for AuthService
		// For now, pass nil to StubHandler since it's mainly used for fallback
		stub = httpapi.NewStubHandler(nil, authStore, db)
		stub.SetLogger(logger) // Set logger for user login logging
		var qinglanClient *service.QinglanClient
		if cfg.Qinglan.APIBaseURL != "" {
			qinglanClient = service.NewQinglanClient(cfg.Qinglan.APIBaseURL, logger)
		}
		admin = httpapi.NewAdminAPI(unitsRepo, devicesRepo, deviceStoreRepo, tenantResolver, stub, logger, qinglanClient)

		// 创建 Role 和 RolePermission Service（仅创建，不保留变量）
		roleRepo := repository.NewPostgresRolesRepository(db)
		rolePermRepo := repository.NewPostgresRolePermissionsRepository(db)
		usersRepo = repository.NewPostgresUsersRepository(db)
		_ = service.NewRoleService(roleRepo, usersRepo, logger)
		_ = service.NewRolePermissionService(rolePermRepo, logger)

		// tags - 已删除（tags 表不存在）

		// 创建配置变更通知器
		configNotifier := notifier.NewConfigNotifier(redisClient, logger)

		// 卡片同步：写 DB，发 config.card.*，可选写 VitalFocusCard 静态缓存
		cardRepo = repository.NewPostgresCardRepository(db, logger)
		cardSyncService = service.NewCardSyncService(cardRepo, configNotifier, logger)

		// 创建 Card Creator（用于启动时全量更新），现在可以使用已创建的 cardRepo
		cardCreator = cardcreator.NewCardCreator(cardRepo, logger)
		vitalCache := cache.NewVitalFocusStaticCache(redisClient, cardRepo, logger)
		cardSyncService.SetVitalFocusStaticCache(vitalCache)
		userCardsCache = cache.NewUserCardsCache(kv, logger)
		cardSyncService.SetUserCardsCache(userCardsCache)

		// 创建 Unit Service
		branchesRepo = repository.NewPostgresBranchesRepository(db)
		// 注意：residentsRepo 和 devicesRepo 需要在 UnitService 之前创建
		residentsRepo = repository.NewPostgresResidentsRepository(db)
		// 使用已创建的 devicesRepo，避免重新创建导致 logger 丢失
		_ = service.NewUnitService(unitsRepo, branchesRepo, residentsRepo, devicesRepo, db, cardSyncService, logger)

	} else {
		// DB不可用时使用简化的处理器
	}

	router := httpapi.NewRouter(logger)

	if db != nil {
		// 创建卡片实时数据处理器并注册路由
		allowedProvider := service.NewAllowedCardIDsProvider(kv, usersRepo, residentsRepo, db, logger)
		cardStaticSvc := service.NewCardStaticService(kv, unitsRepo, allowedProvider, userCardsCache, residentsRepo, logger)
		cardRealtimeSvc := service.NewCardRealtimeService(kv, allowedProvider, logger)

		cardRealtimeHandler := httpapi.NewCardRealtimeHandler(cardRealtimeSvc, cardStaticSvc, logger)
		router.RegisterCardRealtimeRoutes(cardRealtimeHandler)

		// 创建 VitalFocusHandler（前端 Monitor API）
		vitalFocusHandler := httpapi.NewVitalFocusHandler(logger)
		vitalFocusHandler.SetCardService(cardStaticSvc)
		vitalFocusHandler.SetRealtimeService(cardRealtimeSvc)
		vitalFocusHandler.SetUsersRepo(usersRepo) // 用于会话验证
		router.RegisterVitalFocusRoutes(vitalFocusHandler)
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
			// password_hash should only depend on password itself (independent of account/phone/email)
			ah, _ := hex.DecodeString(httpapi.HashAccount("sysadmin"))
			aph, _ := hex.DecodeString(httpapi.HashPassword("ChangeMe123!"))
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

		admin = httpapi.NewAdminAPI(unitsRepo, devicesRepo, deviceStoreRepo, tenantResolver, stub, logger, qinglanClient)

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

		// tags - 已删除（tags 表不存在）

		// 创建配置变更通知器
		configNotifier := notifier.NewConfigNotifier(redisClient, logger)

		// 创建 Card Creator（用于启动时全量更新，保留向后兼容）
		cardCreator = cardcreator.NewCardCreator(cardRepo, logger)

		// 创建 AlarmCloud Service 和 Handler
		alarmCloudRepo := repository.NewPostgresAlarmCloudRepository(db)
		configVersionsRepo := repository.NewPostgresConfigVersionsRepository(db)
		alarmCloudService := service.NewAlarmCloudService(alarmCloudRepo, configVersionsRepo, db, configNotifier, logger)
		alarmCloudHandler := httpapi.NewAlarmCloudHandler(alarmCloudService, logger)
		router.RegisterAlarmCloudRoutes(alarmCloudHandler)

		// 创建 Auth Service 和 Handler
		authRepo := repository.NewPostgresAuthRepository(db)
		authService := service.NewAuthService(authRepo, tenantsRepo, db, logger)
		authHandler := httpapi.NewAuthHandler(authService, logger)
		router.RegisterAuthRoutes(authHandler)

		// 创建 Card Repository（用于启动时全量更新，保留向后兼容）
		cardRepo = repository.NewPostgresCardRepository(db, logger)

		// 创建 IoTTimeSeriesClient（用于调用 wisefido-iot-timeseries 内部 API）
		iotTimeSeriesClient := service.NewIoTTimeSeriesClient(cfg.IoTTimeSeries.InternalAPIBaseURL, logger)

		// 创建 Device Service 和 Handler（qinglanClient 已在上面创建）
		devicesRepo.SetLogger(logger) // 确保 logger 已设置（用于设备连接日志）
		deviceService := service.NewDeviceService(devicesRepo, cardSyncService, iotTimeSeriesClient, qinglanClient, logger)
		deviceHandler := httpapi.NewDeviceHandler(deviceService, logger)
		router.RegisterDeviceRoutes(deviceHandler)

		// 创建 DeviceStore Handler（直接使用 Repository，不需要 Service 层）
		deviceStoreHandler := httpapi.NewDeviceStoreHandler(deviceStoreRepo, qinglanClient, logger)
		router.RegisterDeviceStoreRoutes(deviceStoreHandler)

		// 创建 CardsRepository（供 RadarInstall 查卡片设备，以及后续 CardService 使用）
		cardsRepo := repository.NewPostgresCardsRepository(db)

		// 创建 Radar Install Service 和 Handler（通过 wisefido-qinglan 与设备通信）
		radarInstall := service.NewRadarInstall(cfg, devicesRepo, cardsRepo, configVersionsRepo, qinglanClient, logger)
		radarHandler := httpapi.NewRadarHandler(radarInstall, stub, kv, redisClient, logger)
		router.RegisterRadarRoutes(radarHandler)

		// 创建 Branch Service 和 Handler
		branchService := service.NewBranchService(branchesRepo, db, logger)
		branchesHandler := httpapi.NewBranchesHandler(branchService, db, logger)
		router.RegisterBranchesRoutes(branchesHandler)

		// 创建 User Service 和 Handler
		// usersRepo 已在上面创建 RoleService 时声明，这里直接使用
		// branchesRepo 已在上面创建 UnitService 时声明，这里直接使用
		userBranchesRepo := repository.NewPostgresUserBranchesRepository(db)
		userService := service.NewUserService(usersRepo, branchesRepo, userBranchesRepo, db, logger)
		userHandler := httpapi.NewUserHandler(userService, db, logger)
		router.RegisterUsersRoutes(userHandler)

		// 创建 ConfigPublisher（用于发送所有 config:* 消息）
		configPublisher := publisher.NewConfigPublisher(redisClient, logger)

		// 创建 DeviceMonitorSettings Service 和 Handler
		alarmDeviceRepo := repository.NewPostgresAlarmDeviceRepository(db)
		// 使用已创建的 alarmCloudRepo、configVersionsRepo（在 AlarmCloud Service 创建时已创建）
		// 使用已创建的 configNotifier（在 AlarmCloud Service 创建时已创建）
		deviceMonitorSettingsService := service.NewDeviceMonitorSettingsService(
			alarmDeviceRepo,
			alarmCloudRepo,
			configVersionsRepo, // 使用已创建的 configVersionsRepo
			devicesRepo,
			deviceStoreRepo,
			db, // 添加 db 参数用于事务操作
			configNotifier,
			configPublisher,
			logger,
		)

		// SleepaceReportService
		sleepaceReportsRepo := repository.NewPostgresSleepaceReportsRepository(db)
		sleepaceReportService := service.NewSleepaceReportService(sleepaceReportsRepo, db, logger)

		// 初始化 Sleepace 客户端（如果配置了 Sleepace 服务）
		var sleepaceClient *service.SleepaceClient
		if cfg.Sleepace.HttpAddress != "" && cfg.Sleepace.AppID != "" && cfg.Sleepace.SecretKey != "" {
			sleepaceClient = service.NewSleepaceClient(
				cfg.Sleepace.HttpAddress,
				cfg.Sleepace.AppID,
				cfg.Sleepace.SecretKey,
				logger,
			)

			// 设置 pushType = MQTT（参考 v1.0: wisefido-backend/wisefido-sleepace/modules/sleepace_service.go::setPushType）
			// 这是全局配置，只需要在服务启动时设置一次
			if err := sleepaceClient.SetPushType(); err != nil {
				logger.Warn("Failed to set Sleepace push type to MQTT (this may affect data reception)",
					zap.Error(err),
				)
				// 不阻止服务启动，只记录警告
			} else {
				logger.Info("Sleepace push type set to MQTT successfully")
			}

			// 设置客户端到 SleepaceReportService（延迟初始化）
			if svc, ok := sleepaceReportService.(interface {
				SetSleepaceClient(client *service.SleepaceClient)
			}); ok {
				svc.SetSleepaceClient(sleepaceClient)
			}
			// 设置客户端到 DeviceMonitorSettingsService（用于从硬件读取配置）
			if svc, ok := deviceMonitorSettingsService.(interface {
				SetSleepaceClient(client *service.SleepaceClient)
			}); ok {
				svc.SetSleepaceClient(sleepaceClient)
				logger.Info("Sleepace client set for device monitor settings service")
			}
			logger.Info("Sleepace client initialized",
				zap.String("http_address", cfg.Sleepace.HttpAddress),
				zap.String("app_id", cfg.Sleepace.AppID),
			)
		} else {
			logger.Warn("Sleepace client not initialized (missing configuration)",
				zap.String("http_address", cfg.Sleepace.HttpAddress),
				zap.String("app_id", cfg.Sleepace.AppID),
			)
		}

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
			redisClient,      // 保留字段（已不再使用，改为通过 HTTP API 查询设备状态）
			db,               // 用于安全验证：数据库查询
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

		// 创建 Resident Service 和 Handler
		residentsRepo = repository.NewPostgresResidentsRepository(db)
		residentService := service.NewResidentService(residentsRepo, db, cardSyncService, logger)
		residentHandler := httpapi.NewResidentHandler(residentService, db, logger)
		router.RegisterResidentRoutes(residentHandler)

		sleepaceReportHandler := httpapi.NewSleepaceReportHandler(sleepaceReportService, db, logger)
		router.RegisterSleepaceReportRoutes(sleepaceReportHandler)

		// TODO: MQTT 触发下载功能（默认禁用）
		// 参考：wisefido-backend/wisefido-sleepace/modules/borker.go
		// 实现步骤：
		// 1. 如果 cfg.MQTT.Enabled == true，初始化 MQTT 客户端
		// 2. 创建 SleepaceMQTTBroker 实例
		// 3. 订阅 MQTT 主题
		// 4. 启动消息处理
		//
		// if cfg.MQTT.Enabled {
		//     // 使用 owl-common/mqtt/client.go 创建 MQTT 客户端
		//     mqttConfig := &commoncfg.MQTTConfig{
		//         Broker:   cfg.MQTT.Broker,
		//         ClientID: cfg.MQTT.ClientID,
		//         Username: cfg.MQTT.Username,
		//         Password: cfg.MQTT.Password,
		//     }
		//     mqttClient, err := mqttcommon.NewClient(mqttConfig, logger)
		//     if err != nil {
		//         logger.Error("Failed to create MQTT client", zap.Error(err))
		//     } else {
		//         // 创建 MQTT Broker
		//         mqttBroker := mqtt.NewSleepaceMQTTBroker(sleepaceReportService, logger)
		//         // 启动 MQTT Broker
		//         if err := mqttBroker.Start(ctx, mqttClient); err != nil {
		//             logger.Error("Failed to start MQTT broker", zap.Error(err))
		//         } else {
		//             logger.Info("MQTT broker started",
		//                 zap.String("broker", cfg.MQTT.Broker),
		//                 zap.String("topic", cfg.MQTT.Topic),
		//             )
		//             // 在服务停止时停止 MQTT Broker
		//             defer mqttBroker.Stop(ctx, mqttClient)
		//         }
		//     }
		// } else {
		//     logger.Info("MQTT trigger download is disabled (set MQTT_ENABLED=true to enable)")
		// }
	} else {
		// DB 未就绪：使用内存 repo 支持联测（UnitList/Devices 等页面不再 404/不再因无 DB 失败）
		// 注意：MemoryUnitsRepo 尚未实现新的 UnitsRepository 接口，暂时不使用
		// unitsRepo := repository.NewMemoryUnitsRepo()
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
		// Devices 仍可先保持 stub（后续需要再补内存设备库）
		// 注意：MemoryUnitsRepo 尚未实现新的 UnitsRepository 接口，暂时传递 nil
		// AdminAPI 会回退到 stub handler
		admin = httpapi.NewAdminAPI(nil, nil, nil, nil, stub, logger, nil)
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
		doctor := httpapi.NewDoctorHandler(db, redisClient, logger)
		// 启用 pprof（如果配置了）
		if os.Getenv("DOCTOR_PPROF") == "true" {
			doctor.EnablePprof(true)
		}
		router.RegisterDoctorRoutes(doctor)
	}

	srv := service.NewServer(cfg.HTTP.Addr, router, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动时全量检查并更新卡片（如果 DB 和 cardCreator 可用）
	if db != nil && cardCreator != nil {
		go func() {
			// 延迟启动，等待服务完全初始化
			time.Sleep(2 * time.Second)

			logger.Info("Starting full card check/update on service startup")

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
						stats, err := cardCreator.CreateCardsForUnit(tenant.TenantID, unitID)
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
