package config

import (
	"os"
	"owl-common/config"
)

// Config 雷达服务配置
type Config struct {
	Database config.DatabaseConfig
	Redis    config.RedisConfig
	MQTT     config.MQTTConfig
	
	// 雷达服务特定配置
	Radar struct {
		Topics struct {
			Data    string // 数据主题，如 "radar/+/data"
			Command string // 命令主题，如 "radar/+/command"
			OTA     string // OTA主题，如 "radar/+/ota"
		}
		OTA struct {
			Enabled        bool
			FirmwarePath   string // 固件文件路径
			CheckInterval  string // 检查间隔
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
	
	cfg.MQTT.Broker = getEnv("MQTT_BROKER", "tcp://localhost:1883")
	cfg.MQTT.ClientID = getEnv("MQTT_CLIENT_ID", "wisefido-radar")
	cfg.MQTT.Username = getEnv("MQTT_USERNAME", "")
	cfg.MQTT.Password = getEnv("MQTT_PASSWORD", "")
	
	// 雷达服务配置
	cfg.Radar.Topics.Data = getEnv("RADAR_TOPIC_DATA", "radar/+/data")
	cfg.Radar.Topics.Command = getEnv("RADAR_TOPIC_COMMAND", "radar/+/command")
	cfg.Radar.Topics.OTA = getEnv("RADAR_TOPIC_OTA", "radar/+/ota")
	
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

