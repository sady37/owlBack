package config

import (
	"os"
	"strconv"
	"owl-common/config"
)

// Config 卡片聚合服务配置
type Config struct {
	Database config.DatabaseConfig
	Redis    config.RedisConfig
	
	// 卡片聚合服务特定配置
	Aggregator struct {
		// 租户 ID（用于多租户场景，当前先支持单个租户）
		TenantID string
		
		// 卡片创建触发条件
		// 监听设备/住户/床位绑定关系变化的方式
		// 选项：polling（轮询）、events（事件驱动，待实现）
		// 📝 当前状态：使用轮询模式（每60秒全量更新）
		//     事件驱动模式待 wisefido-data 服务实现后再启用
		//     详见：docs/PENDING_FEATURES.md
		TriggerMode string // "polling" 或 "events"
		
		// 轮询模式配置
		Polling struct {
			Interval int // 轮询间隔（秒），默认 60 秒
		}
		
		// Redis Streams 配置（用于接收事件）
		EventStream      string // 事件流名称，如 "card:events"
		ConsumerGroup    string // 消费者组名称，如 "card-aggregator-group"
		ConsumerName     string // 消费者名称，如 "card-aggregator-1"
		BatchSize        int    // 批量处理大小，默认 10
		
		// 数据聚合配置
		Aggregation struct {
			Enabled  bool // 是否启用数据聚合功能
			Interval int  // 聚合间隔（秒），默认 10 秒
		}
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
	
	cfg.Redis.Addr = getEnv("REDIS_ADDR", "localhost:6379")
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "")
	cfg.Redis.DB = 0
	
	// 卡片聚合服务配置
	cfg.Aggregator.TenantID = getEnv("TENANT_ID", "")
	cfg.Aggregator.TriggerMode = getEnv("CARD_TRIGGER_MODE", "polling")
	cfg.Aggregator.Polling.Interval = 60 // 默认 60 秒
	cfg.Aggregator.EventStream = getEnv("CARD_EVENT_STREAM", "card:events")
	cfg.Aggregator.ConsumerGroup = getEnv("CARD_CONSUMER_GROUP", "card-aggregator-group")
	cfg.Aggregator.ConsumerName = getEnv("CARD_CONSUMER_NAME", "card-aggregator-1")
	cfg.Aggregator.BatchSize = 10 // 默认批量处理 10 条消息
	
	// 数据聚合配置
	cfg.Aggregator.Aggregation.Enabled = getEnv("CARD_AGGREGATION_ENABLED", "true") == "true"
	aggIntervalStr := getEnv("CARD_AGGREGATION_INTERVAL", "10")
	if v, err := strconv.Atoi(aggIntervalStr); err == nil && v > 0 {
		cfg.Aggregator.Aggregation.Interval = v
	} else {
		cfg.Aggregator.Aggregation.Interval = 10 // 默认 10 秒聚合一次
	}
	
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

