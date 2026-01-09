package config

import (
	"os"
	"owl-common/config"
	"strconv"
)

// Config 卡片管理服务配置
type Config struct {
	Database config.DatabaseConfig
	Redis    config.RedisConfig

	// HTTP 服务器配置
	Server struct {
		Port string // HTTP 服务器端口，默认 "8082"
	}

	// 卡片管理服务特定配置
	CardManage struct {
		// 租户 ID（用于多租户场景，当前先支持单个租户）
		TenantID string

		// 卡片创建触发条件
		// 监听设备/住户/床位绑定关系变化的方式
		// 选项：polling（轮询）、api（API 触发）
		TriggerMode string // "polling" 或 "api"

		// 轮询模式配置
		Polling struct {
			Interval int // 轮询间隔（秒），默认 60 分钟（3600 秒），作为保底机制
		}

		// Redis Streams 配置（用于接收事件，可选）
		EventStream   string // 事件流名称，如 "card:events"
		ConsumerGroup string // 消费者组名称，如 "card-manage-group"
		ConsumerName  string // 消费者名称，如 "card-manage-1"
		BatchSize     int    // 批量处理大小，默认 10
	}

	Log struct {
		Level  string
		Format string
	}
}

// Load 加载配置
func Load() (*Config, error) {
	cfg := &Config{}

	// 从环境变量加载（默认值）
	cfg.Database.Host = getEnv("DB_HOST", "localhost")
	cfg.Database.Port = 5432
	cfg.Database.User = getEnv("DB_USER", "postgres")
	cfg.Database.Password = getEnv("DB_PASSWORD", "postgres")
	cfg.Database.Database = getEnv("DB_NAME", "owlrd")
	cfg.Database.SSLMode = getEnv("DB_SSLMODE", "disable")

	// 使用 127.0.0.1 而不是 localhost 以避免 IPv6 解析问题
	cfg.Redis.Addr = getEnv("REDIS_ADDR", "127.0.0.1:6379")
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "")
	cfg.Redis.DB = 0

	// HTTP 服务器配置
	cfg.Server.Port = getEnv("CARD_MANAGE_PORT", "8082")

	// 卡片管理服务配置
	cfg.CardManage.TenantID = getEnv("TENANT_ID", "")
	cfg.CardManage.TriggerMode = getEnv("CARD_TRIGGER_MODE", "api") // 默认 API 模式
	// 轮询间隔：默认 60 分钟（3600 秒），作为保底机制
	pollingIntervalStr := getEnv("CARD_POLLING_INTERVAL", "3600")
	if v, err := strconv.Atoi(pollingIntervalStr); err == nil && v > 0 {
		cfg.CardManage.Polling.Interval = v
	} else {
		cfg.CardManage.Polling.Interval = 3600 // 默认 60 分钟（3600 秒）
	}
	cfg.CardManage.EventStream = getEnv("CARD_EVENT_STREAM", "card:events")
	cfg.CardManage.ConsumerGroup = getEnv("CARD_CONSUMER_GROUP", "card-manage-group")
	cfg.CardManage.ConsumerName = getEnv("CARD_CONSUMER_NAME", "card-manage-1")
	cfg.CardManage.BatchSize = 10 // 默认批量处理 10 条消息

	cfg.Log.Level = getEnv("LOG_LEVEL", "info")
	cfg.Log.Format = getEnv("LOG_FORMAT", "json")

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

