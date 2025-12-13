package config

import (
	"os"
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
		PollInterval int // 轮询间隔（秒），默认 5秒
		
		// 评估配置
		Evaluation struct {
			BatchSize int // 批量评估卡片数量，默认 10
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
	
	// 报警服务配置
	cfg.Alarm.Cache.RealtimeKeyPrefix = getEnv("CACHE_REALTIME_PREFIX", "vital-focus:card:")
	cfg.Alarm.Cache.RealtimeSuffix = ":realtime"
	cfg.Alarm.Cache.AlarmKeyPrefix = getEnv("CACHE_ALARM_PREFIX", "vital-focus:card:")
	cfg.Alarm.Cache.AlarmSuffix = ":alarms"
	cfg.Alarm.Cache.AlarmTTL = 30 // 30秒
	cfg.Alarm.Cache.StateKeyPrefix = getEnv("CACHE_STATE_PREFIX", "alarm:state:")
	
	cfg.Alarm.PollInterval = 5 // 5秒轮询一次
	cfg.Alarm.Evaluation.BatchSize = 10
	
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

