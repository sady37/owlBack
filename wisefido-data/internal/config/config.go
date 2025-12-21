package config

import (
	"os"
	"strconv"

	commoncfg "owl-common/config"
)

// Config wisefido-data（HTTP API）配置
type Config struct {
	HTTP struct {
		Addr string
	}
	DBEnabled bool
	Database  commoncfg.DatabaseConfig
	Redis     struct {
		Addr     string
		Password string
		DB       int
	}
	Log struct {
		Level  string
		Format string
	}
	Sleepace SleepaceConfig `yaml:"sleepace"`
	MQTT     MQTTConfig     `yaml:"mqtt"`
}

// SleepaceConfig Sleepace 厂家服务配置
type SleepaceConfig struct {
	HttpAddress string `yaml:"http_address"` // Sleepace 厂家服务地址
	AppID       string `yaml:"app_id"`      // App ID
	ChannelID   string `yaml:"channel_id"`  // Channel ID
	SecretKey   string `yaml:"secret_key"`  // Secret Key
	Timezone    int    `yaml:"timezone"`    // 时区偏移（秒）
}

// MQTTConfig MQTT 配置（用于触发报告下载）
type MQTTConfig struct {
	Enabled  bool   `yaml:"enabled"`  // 是否启用 MQTT 触发下载（默认 false）
	Broker   string `yaml:"broker"`   // MQTT Broker 地址（如 "tcp://localhost:1883"）
	ClientID string `yaml:"client_id"` // 客户端 ID
	Username string `yaml:"username"` // 用户名（可选）
	Password string `yaml:"password"` // 密码（可选）
	Topic    string `yaml:"topic"`    // 订阅的主题（如 "sleepace-57136"）
}

func Load() *Config {
	cfg := &Config{}
	cfg.HTTP.Addr = getEnv("HTTP_ADDR", ":8080")

	// Default to true for local dev: if DB is unavailable, wisefido-data will fall back to stub.
	// This avoids "empty admin pages" when starting with plain `go run`.
	cfg.DBEnabled = getEnv("DB_ENABLED", "true") == "true"
	cfg.Database.Host = getEnv("DB_HOST", "localhost")
	cfg.Database.Port = parseInt(getEnv("DB_PORT", "5432"), 5432)
	cfg.Database.User = getEnv("DB_USER", "postgres")
	cfg.Database.Password = getEnv("DB_PASSWORD", "postgres")
	cfg.Database.Database = getEnv("DB_NAME", "owlrd")
	cfg.Database.SSLMode = getEnv("DB_SSLMODE", "disable")

	cfg.Redis.Addr = getEnv("REDIS_ADDR", "localhost:6379")
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "")
	cfg.Redis.DB = 0
	cfg.Log.Level = getEnv("LOG_LEVEL", "info")
	cfg.Log.Format = getEnv("LOG_FORMAT", "json")

	// Sleepace 配置
	cfg.Sleepace.HttpAddress = getEnv("SLEEPACE_HTTP_ADDRESS", "http://47.90.180.176:8080")
	cfg.Sleepace.AppID = getEnv("SLEEPACE_APP_ID", "")
	cfg.Sleepace.ChannelID = getEnv("SLEEPACE_CHANNEL_ID", "")
	cfg.Sleepace.SecretKey = getEnv("SLEEPACE_SECRET_KEY", "")
	cfg.Sleepace.Timezone = parseInt(getEnv("SLEEPACE_TIMEZONE", "28800"), 28800) // 默认 UTC+8

	// MQTT 配置（用于触发报告下载，默认禁用）
	cfg.MQTT.Enabled = getEnv("MQTT_ENABLED", "false") == "true"
	cfg.MQTT.Broker = getEnv("MQTT_BROKER", "tcp://localhost:1883")
	cfg.MQTT.ClientID = getEnv("MQTT_CLIENT_ID", "wisefido-data-sleepace")
	cfg.MQTT.Username = getEnv("MQTT_USERNAME", "")
	cfg.MQTT.Password = getEnv("MQTT_PASSWORD", "")
	cfg.MQTT.Topic = getEnv("MQTT_TOPIC", "sleepace-57136") // Sleepace 厂家提供的主题

	return cfg
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func parseInt(s string, def int) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return i
}
