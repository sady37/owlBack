package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"wisefido-data/internal/config"
	"wisefido-data/internal/domain"
	httpapi "wisefido-data/internal/http"
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

	vital := httpapi.NewVitalFocusHandler(kv, logger)
	router := httpapi.NewRouter(logger)
	router.RegisterVitalFocusRoutes(vital)

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
	// Stub depends on tenantsRepo + authStore (used by /auth/api/v1/institutions/search + /auth/api/v1/login)
	stub := httpapi.NewStubHandler(nil, authStore, nil)
	// Always register admin routes; if DB is not available, AdminAPI will fall back to stub (no 404).
	admin := httpapi.NewAdminAPI(nil, nil, nil, nil, stub, logger)
	if cfg.DBEnabled {
		if d, err := database.NewPostgresDB(&cfg.Database); err == nil {
			db = d
			logger.Info("DB enabled for wisefido-data")
		} else {
			logger.Warn("DB enabled but connection failed, falling back to stub", zap.Error(err))
		}
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

		unitsRepo := repository.NewPostgresUnitsRepository(db)
		devicesRepo := repository.NewPostgresDevicesRepository(db)
		devicesRepo.SetLogger(logger) // Set logger for device connection logging
		deviceStoreRepo := repository.NewPostgresDeviceStoreRepository(db)
		tenantResolver := repository.NewPostgresTenantResolver(db)
		tenantsRepo = repository.NewPostgresTenantsRepository(db)
		// Note: StubHandler still uses TenantsRepo (old interface), but we need TenantsRepository for AuthService
		// For now, pass nil to StubHandler since it's mainly used for fallback
		stub = httpapi.NewStubHandler(nil, authStore, db)
		stub.SetLogger(logger) // Set logger for user login logging
		admin = httpapi.NewAdminAPI(unitsRepo, devicesRepo, deviceStoreRepo, tenantResolver, stub, logger)

		// 创建 Role 和 RolePermission Service 和 Handler
		roleRepo := repository.NewPostgresRolesRepository(db)
		rolePermRepo := repository.NewPostgresRolePermissionsRepository(db)
		roleService := service.NewRoleService(roleRepo, logger)
		rolePermService := service.NewRolePermissionService(rolePermRepo, logger)
		rolesHandler := httpapi.NewRolesHandler(roleService, logger)
		rolePermHandler := httpapi.NewRolePermissionsHandler(rolePermService, logger)
		router.RegisterRolesRoutes(rolesHandler)
		router.RegisterRolePermissionsRoutes(rolePermHandler)

		// 创建 Tag Service 和 Handler
		tagRepo := repository.NewPostgresTagsRepository(db)
		tagService := service.NewTagService(tagRepo, db, logger)
		tagsHandler := httpapi.NewTagsHandler(tagService, logger)
		router.RegisterTagsRoutes(tagsHandler)

		// 创建 AlarmCloud Service 和 Handler
		alarmCloudRepo := repository.NewPostgresAlarmCloudRepository(db)
		alarmCloudService := service.NewAlarmCloudService(alarmCloudRepo, db, logger)
		alarmCloudHandler := httpapi.NewAlarmCloudHandler(alarmCloudService, logger)
		router.RegisterAlarmCloudRoutes(alarmCloudHandler)

		// 创建 Auth Service 和 Handler
		authRepo := repository.NewPostgresAuthRepository(db)
		authService := service.NewAuthService(authRepo, tenantsRepo, db, logger)
		authHandler := httpapi.NewAuthHandler(authService, logger)
		router.RegisterAuthRoutes(authHandler)

		// 创建 Device Service 和 Handler
		devicesRepo.SetLogger(logger) // 确保 logger 已设置（用于设备连接日志）
		deviceService := service.NewDeviceService(devicesRepo, logger)
		deviceHandler := httpapi.NewDeviceHandler(deviceService, logger)
		router.RegisterDeviceRoutes(deviceHandler)

		// 创建 DeviceStore Handler（直接使用 Repository，不需要 Service 层）
		deviceStoreHandler := httpapi.NewDeviceStoreHandler(deviceStoreRepo, logger)
		router.RegisterDeviceStoreRoutes(deviceStoreHandler)

		// 创建 Unit Service 和 Handler
		unitService := service.NewUnitService(unitsRepo, logger)
		unitHandler := httpapi.NewUnitHandler(unitService, logger)
		router.RegisterUnitRoutes(unitHandler)

		// 创建 User Service 和 Handler
		usersRepo := repository.NewPostgresUsersRepository(db)
		userService := service.NewUserService(usersRepo, logger)
		userHandler := httpapi.NewUserHandler(userService, logger)
		router.RegisterUsersRoutes(userHandler)

		// 创建 DeviceMonitorSettings Service 和 Handler
		alarmDeviceRepo := repository.NewPostgresAlarmDeviceRepository(db)
		deviceMonitorSettingsService := service.NewDeviceMonitorSettingsService(
			alarmDeviceRepo,
			devicesRepo,
			deviceStoreRepo,
			logger,
		)
		deviceMonitorSettingsHandler := httpapi.NewDeviceMonitorSettingsHandler(deviceMonitorSettingsService, logger)
		router.RegisterDeviceMonitorSettingsRoutes(deviceMonitorSettingsHandler)

		// 创建 AlarmEvent Service 和 Handler
		alarmEventsRepo := repository.NewPostgresAlarmEventsRepository(db)
		alarmEventService := service.NewAlarmEventService(
			alarmEventsRepo,
			devicesRepo,
			unitsRepo,
			usersRepo,
			db,
			logger,
		)
		alarmEventHandler := httpapi.NewAlarmEventHandler(alarmEventService, logger)
		router.RegisterAlarmEventRoutes(alarmEventHandler)

		// 创建 Resident Service 和 Handler
		residentsRepo := repository.NewPostgresResidentsRepository(db)
		residentService := service.NewResidentService(residentsRepo, db, logger)
		residentHandler := httpapi.NewResidentHandler(residentService, db, logger)
		router.RegisterResidentRoutes(residentHandler)

		// SleepaceReportService
		sleepaceReportsRepo := repository.NewPostgresSleepaceReportsRepository(db)
		sleepaceReportService := service.NewSleepaceReportService(sleepaceReportsRepo, db, logger)
		
		// 初始化 Sleepace 客户端（如果配置了 Sleepace 服务）
		if cfg.Sleepace.HttpAddress != "" && cfg.Sleepace.AppID != "" && cfg.Sleepace.SecretKey != "" {
			sleepaceClient := service.NewSleepaceClient(
				cfg.Sleepace.HttpAddress,
				cfg.Sleepace.AppID,
				cfg.Sleepace.SecretKey,
				logger,
			)
			// 设置客户端到 Service（延迟初始化）
			if svc, ok := sleepaceReportService.(interface {
				SetSleepaceClient(client *service.SleepaceClient)
			}); ok {
				svc.SetSleepaceClient(sleepaceClient)
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
		
		sleepaceReportHandler := httpapi.NewSleepaceReportHandler(sleepaceReportService, db, logger)
		router.RegisterSleepaceReportRoutes(sleepaceReportHandler)

		// 创建 Card Service 和 Handler
		cardsRepo := repository.NewPostgresCardsRepository(db)
		cardService := service.NewCardService(
			cardsRepo,
			residentsRepo,
			devicesRepo,
			usersRepo,
			db,
			logger,
		)
		cardOverviewHandler := httpapi.NewCardOverviewHandler(stub, cardService, logger)
		router.RegisterCardOverviewRoutes(cardOverviewHandler)

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
		admin = httpapi.NewAdminAPI(nil, nil, nil, nil, stub, logger)
	}
	router.RegisterAdminUnitDeviceRoutes(admin)
	router.RegisterAdminTenantRoutes(httpapi.NewTenantsHandler(tenantsRepo, authStore, db))
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

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigCh:
		cancel()
	case <-errCh:
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
