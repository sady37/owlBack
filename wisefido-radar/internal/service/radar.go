package service

import (
	"context"
	"database/sql"
	"fmt"
	"wisefido-radar/internal/config"
	"wisefido-radar/internal/consumer"
	"wisefido-radar/internal/http"
	"wisefido-radar/internal/publisher"
	"wisefido-radar/internal/repository"
	
	"go.uber.org/zap"
	"github.com/go-redis/redis/v8"
	"owl-common/database"
	rediscommon "owl-common/redis"
	mqttcommon "owl-common/mqtt"
)

// RadarService 雷达服务
type RadarService struct {
	config      *config.Config
	logger      *zap.Logger
	db          *sql.DB
	redis       *redis.Client
	mqttClient  *mqttcommon.Client
	consumer    *consumer.MQTTConsumer
	httpServer  *http.Server // 新增：HTTPS 服务器
}

// NewRadarService 创建雷达服务
func NewRadarService(cfg *config.Config, logger *zap.Logger) (*RadarService, error) {
	// 初始化数据库
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	
	// 初始化Redis
	redisClient := rediscommon.NewRedisClient(&cfg.Redis)
	if err := rediscommon.Ping(context.Background(), redisClient); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}
	
	// 初始化MQTT
	mqttClient, err := mqttcommon.NewClient(&cfg.MQTT, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MQTT: %w", err)
	}
	
	// 创建Repository
	deviceRepo := repository.NewDeviceRepository(db, logger)
	
	// 创建Publisher（用于发送命令）
	mqttPublisher := publisher.NewMQTTPublisher(cfg, mqttClient, logger)
	
	// 创建CommandService（封装命令发送逻辑）
	// 注意：CommandService 在 http 包中定义，避免循环依赖
	commandService := http.NewCommandService(cfg, mqttPublisher, deviceRepo, redisClient, logger)
	
	// 创建Consumer
	mqttConsumer := consumer.NewMQTTConsumer(cfg, mqttClient, redisClient, deviceRepo, logger)
	
	// 创建认证服务（在 http 包中，避免循环依赖）
	authService := http.NewAuthService(cfg, db, deviceRepo, logger)
	
	// 创建 HTTP 路由和处理器
	router := http.NewRouter(logger)
	authHandler := http.NewAuthHandler(authService, logger)
	router.RegisterAuthRoutes(authHandler)
	
	// 注册命令相关的内部 HTTP API 路由
	commandHandler := http.NewCommandHandler(commandService, logger)
	router.RegisterCommandRoutes(commandHandler)
	
	// 创建 HTTPS 服务器
	httpsAddr := fmt.Sprintf(":%d", cfg.HTTPS.Port)
	httpServer, err := http.NewServer(httpsAddr, router, cfg.HTTPS.CertFile, cfg.HTTPS.KeyFile, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTPS server: %w", err)
	}
	
	return &RadarService{
		config:     cfg,
		logger:     logger,
		db:         db,
		redis:      redisClient,
		mqttClient: mqttClient,
		consumer:   mqttConsumer,
		httpServer: httpServer,
	}, nil
}

// Start 启动服务
func (s *RadarService) Start(ctx context.Context) error {
	s.logger.Info("Starting radar service components")
	
	// 启动 HTTPS 服务器（在 goroutine 中运行）
	go func() {
		if err := s.httpServer.Start(); err != nil {
			s.logger.Error("HTTPS server error", zap.Error(err))
		}
	}()
	
	// 启动MQTT消费者
	if err := s.consumer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start MQTT consumer: %w", err)
	}
	
	s.logger.Info("Radar service started successfully")
	return nil
}

// Stop 停止服务
func (s *RadarService) Stop(ctx context.Context) error {
	s.logger.Info("Stopping radar service")
	
	// 停止 HTTPS 服务器
	if s.httpServer != nil {
		if err := s.httpServer.Stop(ctx); err != nil {
			s.logger.Error("Error stopping HTTPS server", zap.Error(err))
		}
	}
	
	// 停止Consumer
	if s.consumer != nil {
		if err := s.consumer.Stop(ctx); err != nil {
			s.logger.Error("Error stopping consumer", zap.Error(err))
		}
	}
	
	// 断开MQTT
	if s.mqttClient != nil {
		s.mqttClient.Disconnect()
	}
	
	// 关闭Redis
	if s.redis != nil {
		rediscommon.Close(s.redis)
	}
	
	// 关闭数据库
	if s.db != nil {
		database.Close(s.db)
	}
	
	s.logger.Info("Radar service stopped")
	return nil
}

