package config

import (
	"os"
	"owl-common/config"
)

// Config 数据转换服务配置
type Config struct {
	Database config.DatabaseConfig
	Redis    config.RedisConfig
	
	// 数据转换服务特定配置
	Transformer struct {
		// Redis Streams 配置
		Streams struct {
			Radar    string // 雷达数据流，如 "radar:data:stream"
			Sleepace string // Sleepace 数据流，如 "sleepace:data:stream"
			Output   string // 输出数据流，如 "iot:data:stream"
		}
		ConsumerGroup string // 消费者组名称
		ConsumerName  string // 消费者名称
		BatchSize     int64  // 批量处理大小
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
	
	// 数据转换服务配置
	cfg.Transformer.Streams.Radar = getEnv("STREAM_RADAR", "radar:data:stream")
	cfg.Transformer.Streams.Sleepace = getEnv("STREAM_SLEEPACE", "sleepace:data:stream")
	cfg.Transformer.Streams.Output = getEnv("STREAM_OUTPUT", "iot:data:stream")
	cfg.Transformer.ConsumerGroup = getEnv("CONSUMER_GROUP", "data-transformer-group")
	cfg.Transformer.ConsumerName = getEnv("CONSUMER_NAME", "data-transformer-1")
	cfg.Transformer.BatchSize = 10
	
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

