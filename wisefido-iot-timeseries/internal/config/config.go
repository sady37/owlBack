package config

import (
	"os"
	"owl-common/config"
	"strconv"
)

// Config IoT 时序数据服务配置
type Config struct {
	Database config.DatabaseConfig
	Redis    config.RedisConfig

	// IoT 时序数据服务特定配置
	Streams struct {
		// 统一 IoT streams（所有设备类型）
		Monitor string // iot:monitor:stream - 所有设备的实时数据
		Stat    string // iot:stat:stream    - 所有设备的统计数据
		Event   string // iot:event:stream   - 所有设备的事件数据
		Alarm   string // iot:alarm:stream   - 所有设备的告警数据
		// 注意：认证流 iot:auth:stream 由其他服务处理
	}
	ConsumerGroup string // 消费者组名称
	ConsumerName  string // 消费者名称
	BatchSize     int64  // 批量处理大小

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
	// 默认端口使用环境变量，如果没有则使用 5433（与 start_owlback.sh 保持一致）
	if portStr := getEnv("DB_PORT", ""); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil && port > 0 {
			cfg.Database.Port = port
		} else {
			cfg.Database.Port = 5433 // 默认使用 5433（与 start_owlback.sh 保持一致）
		}
	} else {
		cfg.Database.Port = 5433 // 默认使用 5433（与 start_owlback.sh 保持一致）
	}
	cfg.Database.User = getEnv("DB_USER", "postgres")
	cfg.Database.Password = getEnv("DB_PASSWORD", "postgres")
	cfg.Database.Database = getEnv("DB_NAME", "owlrd")
	cfg.Database.SSLMode = getEnv("DB_SSLMODE", "disable")

	cfg.Redis.Addr = getEnv("REDIS_ADDR", "localhost:6379")
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "")
	cfg.Redis.DB = 0

	// IoT 时序数据服务配置 - 统一 IoT streams
	cfg.Streams.Monitor = getEnv("STREAM_IOT_MONITOR", "iot:monitor:stream")
	cfg.Streams.Stat = getEnv("STREAM_IOT_STAT", "iot:stat:stream")
	cfg.Streams.Event = getEnv("STREAM_IOT_EVENT", "iot:event:stream")
	cfg.Streams.Alarm = getEnv("STREAM_IOT_ALARM", "iot:alarm:stream")
	cfg.ConsumerGroup = getEnv("CONSUMER_GROUP", "iot-timeseries-group")
	cfg.ConsumerName = getEnv("CONSUMER_NAME", "iot-timeseries-1")
	cfg.BatchSize = 10

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
