package config

import (
	"os"
	"owl-common/config"
	"strconv"
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
		// 注意：数据流分类直接硬编码在代码中，使用设备级别 streams
		// realtime/sleepStage → sleepace:monitor:stream
		// connectionStatus → sleepace:event:stream
		// alarmNotify → sleepace:alarm:stream
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

	cfg.MQTT.Broker = getEnv("MQTT_BROKER", "mqtt://127.0.0.1:1883")
	cfg.MQTT.ClientID = getEnv("MQTT_CLIENT_ID", "wisefido-sleepace")
	cfg.MQTT.Username = getEnv("MQTT_USERNAME", "wisefido")
	cfg.MQTT.Password = getEnv("MQTT_PASSWORD", "")

	// Sleepace 服务配置（Sleepace 厂家程序在同一台机器上，使用内部地址）
	cfg.Sleepace.HttpAddress = getEnv("SLEEPACE_HTTP_ADDRESS", "http://127.0.0.1:8080")
	cfg.Sleepace.AppId = getEnv("SLEEPACE_APP_ID", "")
	cfg.Sleepace.ChannelId = getEnv("SLEEPACE_CHANNEL_ID", "")
	cfg.Sleepace.SecretKey = getEnv("SLEEPACE_SECRET_KEY", "")
	cfg.Sleepace.Timezone = 8
	cfg.Sleepace.RealtimeInterval = 30
	cfg.Sleepace.LeaveSensibility = 1
	cfg.Sleepace.ReportUploadType = 0
	cfg.Sleepace.ReportUploadTime = 0
	cfg.Sleepace.Topic = getEnv("SLEEPACE_MQTT_TOPIC", "sleepace-57136")

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
