package config

import (
	"os"
	"strconv"
	"owl-common/config"
)

// Config 报警服务配置
type Config struct {
	Database config.DatabaseConfig
	Redis    config.RedisConfig
	
	// 报警服务特定配置
	Alarm struct {
		// Redis 缓存配置
		Cache struct {
			RealtimeKeyPrefix string // 实时数据缓存键前缀，如 "vital-focus:card:"
			RealtimeSuffix    string // 实时数据缓存键后缀，如 ":realtime"
			AlarmKeyPrefix    string // 报警数据缓存键前缀，如 "vital-focus:card:"
			AlarmSuffix       string // 报警数据缓存键后缀，如 ":alarms"
			AlarmTTL          int    // 报警数据 TTL（秒），默认 30秒
			StateKeyPrefix    string // 报警状态缓存键前缀，如 "alarm:state:"
		}
		
		// 轮询配置（如果使用轮询方式）
		PollInterval int // 轮询间隔（秒），默认 10秒
		
		// 评估配置
		Evaluation struct {
			BatchSize int // 批量评估卡片数量，默认 10
		}
		
		// IoT Stream 配置（直接订阅设备级别的 streams）
		IoTStream struct {
			Enabled       bool   // 是否启用 IoT Stream 消费（事件驱动）
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
	
	cfg.Redis.Addr = getEnv("REDIS_ADDR", "localhost:6379")
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "")
	cfg.Redis.DB = 0
	
	// 报警服务配置
	cfg.Alarm.Cache.RealtimeKeyPrefix = getEnv("CACHE_REALTIME_PREFIX", "vital-focus:card:")
	cfg.Alarm.Cache.RealtimeSuffix = ":realtime"
	cfg.Alarm.Cache.AlarmKeyPrefix = getEnv("CACHE_ALARM_PREFIX", "vital-focus:card:")
	cfg.Alarm.Cache.AlarmSuffix = ":alarms"
	cfg.Alarm.Cache.AlarmTTL = 30 // 30秒
	cfg.Alarm.Cache.StateKeyPrefix = getEnv("CACHE_STATE_PREFIX", "alarm:state:")
	
	cfg.Alarm.PollInterval = 10 // 10秒轮询一次
	cfg.Alarm.Evaluation.BatchSize = 10
	
	// IoT Stream 配置 - 设备级别 streams
	cfg.Alarm.IoTStream.Enabled = getEnv("AI_IOT_STREAM_ENABLED", "true") == "true"
	cfg.Alarm.IoTStream.RadarMonitor = getEnv("AI_STREAM_RADAR_MONITOR", "radar:monitor:stream")
	cfg.Alarm.IoTStream.RadarStat = getEnv("AI_STREAM_RADAR_STAT", "radar:stat:stream")
	cfg.Alarm.IoTStream.RadarEvent = getEnv("AI_STREAM_RADAR_EVENT", "radar:event:stream")
	cfg.Alarm.IoTStream.RadarAlarm = getEnv("AI_STREAM_RADAR_ALARM", "radar:alarm:stream")
	cfg.Alarm.IoTStream.SleepaceMonitor = getEnv("AI_STREAM_SLEEPACE_MONITOR", "sleepace:monitor:stream")
	cfg.Alarm.IoTStream.SleepaceEvent = getEnv("AI_STREAM_SLEEPACE_EVENT", "sleepace:event:stream")
	cfg.Alarm.IoTStream.SleepaceAlarm = getEnv("AI_STREAM_SLEEPACE_ALARM", "sleepace:alarm:stream")
	cfg.Alarm.IoTStream.ConsumerGroup = getEnv("AI_IOT_CONSUMER_GROUP", "wisefido-alarm-events")
	cfg.Alarm.IoTStream.ConsumerName = getEnv("AI_IOT_CONSUMER_NAME", "alarm-consumer-1")
	batchSizeStr := getEnv("AI_IOT_BATCH_SIZE", "10")
	if v, err := strconv.ParseInt(batchSizeStr, 10, 64); err == nil && v > 0 {
		cfg.Alarm.IoTStream.BatchSize = v
	} else {
		cfg.Alarm.IoTStream.BatchSize = 10
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

