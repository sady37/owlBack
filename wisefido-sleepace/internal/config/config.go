package config

import (
	"os"
	"owl-common/config"
)

// Config Sleepace 服务配置
type Config struct {
	Database config.DatabaseConfig
	Redis    config.RedisConfig
	MQTT     config.MQTTConfig
	
	// Sleepace 服务特定配置
	Sleepace struct {
		HttpAddress      string // Sleepace 厂家 HTTP API 地址
		AppId            string // App ID
		ChannelId        string // Channel ID
		SecretKey        string // Secret Key
		Timezone         int    // 时区
		RealtimeInterval int    // 实时数据间隔
		LeaveSensibility int    // 离床灵敏度
		ReportUploadType int    // 报告上传类型
		ReportUploadTime int    // 报告上传时间
		Topic            string // MQTT 主题（Sleepace 厂家提供的主题，如 "sleepace-57136"）
		Stream           string // Redis Streams 输出流，如 "sleepace:data:stream"
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
	
	cfg.MQTT.Broker = getEnv("MQTT_BROKER", "mqtt://47.90.180.176:1883")
	cfg.MQTT.ClientID = getEnv("MQTT_CLIENT_ID", "wisefido-sleepace")
	cfg.MQTT.Username = getEnv("MQTT_USERNAME", "wisefido")
	cfg.MQTT.Password = getEnv("MQTT_PASSWORD", "")
	
	// Sleepace 服务配置
	cfg.Sleepace.HttpAddress = getEnv("SLEEPACE_HTTP_ADDRESS", "http://47.90.180.176:8080")
	cfg.Sleepace.AppId = getEnv("SLEEPACE_APP_ID", "")
	cfg.Sleepace.ChannelId = getEnv("SLEEPACE_CHANNEL_ID", "")
	cfg.Sleepace.SecretKey = getEnv("SLEEPACE_SECRET_KEY", "")
	cfg.Sleepace.Timezone = 8
	cfg.Sleepace.RealtimeInterval = 30
	cfg.Sleepace.LeaveSensibility = 1
	cfg.Sleepace.ReportUploadType = 0
	cfg.Sleepace.ReportUploadTime = 0
	cfg.Sleepace.Topic = getEnv("SLEEPACE_MQTT_TOPIC", "sleepace-57136")
	cfg.Sleepace.Stream = getEnv("SLEEPACE_STREAM", "sleepace:data:stream")
	
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

