package config

import (
	"os"
	"owl-common/config"
)

// Config 传感器融合服务配置
type Config struct {
	Database config.DatabaseConfig
	Redis    config.RedisConfig
	
	// 传感器融合服务特定配置
	Fusion struct {
		// Redis Streams 配置
		Stream struct {
			Input string // 输入数据流，如 "iot:data:stream"
		}
		ConsumerGroup string // 消费者组名称
		ConsumerName  string // 消费者名称
		BatchSize     int64  // 批量处理大小
		
		// Redis 缓存配置
		Cache struct {
			RealtimeKeyPrefix string // 实时数据缓存键前缀，如 "vital-focus:card:"
			RealtimeTTL       int    // 实时数据 TTL（秒），默认 300（5分钟）
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
	
	// 传感器融合服务配置
	cfg.Fusion.Stream.Input = getEnv("STREAM_INPUT", "iot:data:stream")
	cfg.Fusion.ConsumerGroup = getEnv("CONSUMER_GROUP", "sensor-fusion-group")
	cfg.Fusion.ConsumerName = getEnv("CONSUMER_NAME", "sensor-fusion-1")
	cfg.Fusion.BatchSize = 10
	
	cfg.Fusion.Cache.RealtimeKeyPrefix = getEnv("CACHE_REALTIME_PREFIX", "vital-focus:card:")
	cfg.Fusion.Cache.RealtimeTTL = 300 // 5分钟
	
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

