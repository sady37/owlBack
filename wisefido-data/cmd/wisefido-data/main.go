package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"os"
	"os/signal"
	"syscall"
	"time"
	"wisefido-data/internal/config"
	httpapi "wisefido-data/internal/http"
	"wisefido-data/internal/repository"
	"wisefido-data/internal/service"
	"wisefido-data/internal/store"

	"owl-common/database"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()

	logger, _ := zap.NewProduction()
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
	var tenantsRepo repository.TenantsRepo
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
	admin := httpapi.NewAdminAPI(nil, nil, nil, stub, logger)
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

		unitsRepo := repository.NewPostgresUnitsRepo(db)
		devicesRepo := repository.NewPostgresDevicesRepo(db)
		tenantResolver := repository.NewPostgresTenantResolver(db)
		tenantsRepo = repository.NewPostgresTenantsRepo(db)
		stub = httpapi.NewStubHandler(tenantsRepo, authStore, db)
		admin = httpapi.NewAdminAPI(unitsRepo, devicesRepo, tenantResolver, stub, logger)
	} else {
		// DB 未就绪：使用内存 repo 支持联测（UnitList/Devices 等页面不再 404/不再因无 DB 失败）
		unitsRepo := repository.NewMemoryUnitsRepo()
		tenantsRepo = repository.NewMemoryTenantsRepo()
		// Seed "System" tenant for SystemAdmin login in dev (matches httpapi.systemTenantID)
		_, _ = tenantsRepo.UpdateTenant(context.Background(), "00000000-0000-0000-0000-000000000001", map[string]any{
			"tenant_name": "System",
			"domain":      "system.local",
			"status":      "active",
		})
		// Seed SystemAdmin account
		_ = authStore.UpsertUser("00000000-0000-0000-0000-000000000001", "sysadmin", "SystemAdmin", "ChangeMe123!")
		stub = httpapi.NewStubHandler(tenantsRepo, authStore, nil)
		// Devices 仍可先保持 stub（后续需要再补内存设备库）
		admin = httpapi.NewAdminAPI(unitsRepo, nil, nil, stub, logger)
	}
	router.RegisterAdminUnitDeviceRoutes(admin)
	router.RegisterAdminTenantRoutes(httpapi.NewTenantsHandler(tenantsRepo, authStore, db))
	router.RegisterStubRoutes(stub)

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
