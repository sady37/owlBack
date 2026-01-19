package config

import (
	"os"
	"owl-common/config"
	"strconv"
)

// Config 卡片聚合服务配置
type Config struct {
	Database config.DatabaseConfig
	Redis    config.RedisConfig

	// 卡片聚合服务特定配置
	Aggregator struct {
		// 租户 ID（用于多租户场景，当前先支持单个租户）
		TenantID string

		// 数据聚合配置
		Aggregation struct {
			Enabled  bool // 是否启用数据聚合功能
			Interval int  // 聚合间隔（秒），默认 2 秒（匹配心率呼吸数据更新频率）
		}

		// IoT Stream 配置（直接订阅设备级别的 streams）
		IoTStream struct {
			Enabled bool // 是否启用 IoT Stream 消费（事件驱动）
			// Radar 设备 streams
			RadarMonitor string // radar:monitor:stream
			RadarStat    string // radar:stat:stream
			RadarEvent   string // radar:event:stream
			RadarAlarm   string // radar:alarm:stream
			// Sleepace 设备 streams
			SleepaceMonitor string // sleepace:monitor:stream
			SleepaceEvent   string // sleepace:event:stream
			SleepaceAlarm   string // sleepace:alarm:stream
			// 注意：Sleepace 没有 stat 数据
			ConsumerGroup string // 消费者组名称
			ConsumerName  string // 消费者名称
			BatchSize     int64  // 批量处理大小
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

	// 使用 127.0.0.1 而不是 localhost 以避免 IPv6 解析问题
	cfg.Redis.Addr = getEnv("REDIS_ADDR", "127.0.0.1:6379")
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "")
	cfg.Redis.DB = 0

	// 卡片聚合服务配置
	cfg.Aggregator.TenantID = getEnv("TENANT_ID", "")

	// 数据聚合配置
	cfg.Aggregator.Aggregation.Enabled = getEnv("CARD_AGGREGATION_ENABLED", "true") == "true"
	aggIntervalStr := getEnv("CARD_AGGREGATION_INTERVAL", "2")
	if v, err := strconv.Atoi(aggIntervalStr); err == nil && v > 0 {
		cfg.Aggregator.Aggregation.Interval = v
	} else {
		cfg.Aggregator.Aggregation.Interval = 2 // 默认 2 秒聚合一次（匹配心率呼吸数据更新频率）
	}

	// IoT Stream 配置 - 设备级别 streams
	cfg.Aggregator.IoTStream.Enabled = getEnv("CARD_IOT_STREAM_ENABLED", "true") == "true"
	cfg.Aggregator.IoTStream.RadarMonitor = getEnv("CARD_STREAM_RADAR_MONITOR", "iot:monitor:stream")
	cfg.Aggregator.IoTStream.RadarStat = getEnv("CARD_STREAM_RADAR_STAT", "iot:stat:stream")
	cfg.Aggregator.IoTStream.RadarEvent = getEnv("CARD_STREAM_RADAR_EVENT", "iot:event:stream")
	cfg.Aggregator.IoTStream.RadarAlarm = getEnv("CARD_STREAM_RADAR_ALARM", "radar:alarm:stream")
	cfg.Aggregator.IoTStream.SleepaceMonitor = getEnv("CARD_STREAM_SLEEPACE_MONITOR", "sleepace:monitor:stream")
	cfg.Aggregator.IoTStream.SleepaceEvent = getEnv("CARD_STREAM_SLEEPACE_EVENT", "sleepace:event:stream")
	cfg.Aggregator.IoTStream.SleepaceAlarm = getEnv("CARD_STREAM_SLEEPACE_ALARM", "sleepace:alarm:stream")
	cfg.Aggregator.IoTStream.ConsumerGroup = getEnv("CARD_IOT_CONSUMER_GROUP", "card-aggregator-iot-group")
	cfg.Aggregator.IoTStream.ConsumerName = getEnv("CARD_IOT_CONSUMER_NAME", "card-aggregator-iot-1")
	batchSizeStr := getEnv("CARD_IOT_BATCH_SIZE", "10")
	if v, err := strconv.ParseInt(batchSizeStr, 10, 64); err == nil && v > 0 {
		cfg.Aggregator.IoTStream.BatchSize = v
	} else {
		cfg.Aggregator.IoTStream.BatchSize = 10
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
