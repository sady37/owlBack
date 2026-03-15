package config

import (
	"log"
	"os"
	"strconv"

	commoncfg "owl-common/config"

	"gopkg.in/yaml.v3"
)

// Config wisefido-data（HTTP API）配置
type Config struct {
	HTTP struct {
		Addr string `yaml:"addr"`
	} `yaml:"http"`
	DBEnabled bool                     `yaml:"db_enabled"`
	Database  commoncfg.DatabaseConfig `yaml:"database"`
	Redis     struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`
	Log struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
	} `yaml:"logging"`
	Sleepace        SleepaceConfig        `yaml:"sleepace"`
	SleepaceGateway SleepaceGatewayConfig `yaml:"sleepace_gateway"`
	MQTT            MQTTConfig            `yaml:"mqtt"`
	Radar           RadarConfig           `yaml:"radar"`
	Qinglan         QinglanConfig         `yaml:"qinglan"`
	IoTTimeSeries   IoTTimeSeriesConfig   `yaml:"iot_timeseries"`
}

// QinglanConfig wisefido-qinglan HTTP API 配置（雷达设备交互）
type QinglanConfig struct {
	APIBaseURL string `yaml:"api_base_url"`
}

// SleepaceGatewayConfig wisefido-sleepace 网关配置（通过 wisefido-sleepace 下发硬件配置）
type SleepaceGatewayConfig struct {
	APIBaseURL string `yaml:"api_base_url"`
}

// SleepaceConfig Sleepace 厂家服务配置
type SleepaceConfig struct {
	HttpAddress string `yaml:"http_address"` // Sleepace 厂家服务地址
	AppID       string `yaml:"app_id"`       // App ID
	ChannelID   string `yaml:"channel_id"`   // Channel ID
	SecretKey   string `yaml:"secret_key"`   // Secret Key
	Timezone    int    `yaml:"timezone"`     // 时区偏移（秒）
}

// MQTTConfig MQTT 配置（用于触发报告下载）
type MQTTConfig struct {
	Enabled  bool   `yaml:"enabled"`   // 是否启用 MQTT 触发下载（默认 false）
	Broker   string `yaml:"broker"`    // MQTT Broker 地址（如 "tcp://localhost:1883"）
	ClientID string `yaml:"client_id"` // 客户端 ID
	Username string `yaml:"username"`  // 用户名（可选）
	Password string `yaml:"password"`  // 密码（可选）
	Topic    string `yaml:"topic"`     // 订阅的主题（如 "sleepace-57136"）
}

// RadarConfig Radar 服务配置（用于调用 wisefido-radar 内部 API）
type RadarConfig struct {
	InternalAPIBaseURL string `yaml:"internal_api_base_url"` // wisefido-radar 内部 API 地址（如 "http://localhost:8443"）
}

// IoTTimeSeriesConfig IoT 时序数据服务配置（用于调用 wisefido-iot-timeseries 内部 API）
type IoTTimeSeriesConfig struct {
	InternalAPIBaseURL string `yaml:"internal_api_base_url"` // wisefido-iot-timeseries 内部 API 地址（如 "http://localhost:8083"）
}

func Load() *Config {
	configPath := "config.yaml"
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		configPath = path
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("(info) config file not found, using environment variables: %v", err)
		return LoadFromEnv()
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Printf("(warn) failed to parse config file, falling back to env: %v", err)
		return LoadFromEnv()
	}

	cfg.setDefaults()
	return &cfg
}

func (c *Config) setDefaults() {
	if c.HTTP.Addr == "" {
		c.HTTP.Addr = ":8080"
	}
	if c.Database.SSLMode == "" {
		c.Database.SSLMode = "disable"
	}
	if c.Log.Level == "" {
		c.Log.Level = "info"
	}
	if c.Log.Format == "" {
		c.Log.Format = "json"
	}
	if c.Sleepace.HttpAddress == "" {
		c.Sleepace.HttpAddress = "http://127.0.0.1:8090"
	}
	if c.Sleepace.Timezone == 0 {
		c.Sleepace.Timezone = 28800
	}
	if c.MQTT.Broker == "" {
		c.MQTT.Broker = "tcp://localhost:1883"
	}
	if c.MQTT.ClientID == "" {
		c.MQTT.ClientID = "wisefido-data-sleepace"
	}
	if c.MQTT.Topic == "" {
		c.MQTT.Topic = "sleepace-57136"
	}
	if c.Radar.InternalAPIBaseURL == "" {
		c.Radar.InternalAPIBaseURL = "http://localhost:8443"
	}
	if c.IoTTimeSeries.InternalAPIBaseURL == "" {
		c.IoTTimeSeries.InternalAPIBaseURL = "http://localhost:8085"
	}
	if c.SleepaceGateway.APIBaseURL == "" {
		c.SleepaceGateway.APIBaseURL = "http://127.0.0.1:8083"
	}
}

func LoadFromEnv() *Config {
	cfg := &Config{}
	cfg.HTTP.Addr = getEnv("HTTP_ADDR", ":8080")
	cfg.DBEnabled = getEnv("DB_ENABLED", "true") == "true"
	cfg.Database.Host = getEnv("DB_HOST", "localhost")
	cfg.Database.Port = parseInt(getEnv("DB_PORT", "5433"), 5433)
	cfg.Database.User = getEnv("DB_USER", "postgres")
	cfg.Database.Password = getEnv("DB_PASSWORD", "")
	cfg.Database.Database = getEnv("DB_NAME", "owlrd")
	cfg.Database.SSLMode = getEnv("DB_SSLMODE", "disable")
	cfg.Redis.Addr = getEnv("REDIS_ADDR", "localhost:6379")
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "")
	cfg.Redis.DB = 0
	cfg.Log.Level = getEnv("LOG_LEVEL", "info")
	cfg.Log.Format = getEnv("LOG_FORMAT", "json")
	cfg.Sleepace.HttpAddress = getEnv("SLEEPACE_HTTP_ADDRESS", "http://127.0.0.1:8090")
	cfg.Sleepace.AppID = getEnv("SLEEPACE_APP_ID", "")
	cfg.Sleepace.ChannelID = getEnv("SLEEPACE_CHANNEL_ID", "")
	cfg.Sleepace.SecretKey = getEnv("SLEEPACE_SECRET_KEY", "")
	cfg.Sleepace.Timezone = parseInt(getEnv("SLEEPACE_TIMEZONE", "28800"), 28800)
	cfg.MQTT.Enabled = getEnv("MQTT_ENABLED", "false") == "true"
	cfg.MQTT.Broker = getEnv("MQTT_BROKER", "tcp://localhost:1883")
	cfg.MQTT.ClientID = getEnv("MQTT_CLIENT_ID", "wisefido-data-sleepace")
	cfg.MQTT.Username = getEnv("MQTT_USERNAME", "")
	cfg.MQTT.Password = getEnv("MQTT_PASSWORD", "")
	cfg.MQTT.Topic = getEnv("MQTT_TOPIC", "sleepace-57136")
	cfg.Radar.InternalAPIBaseURL = getEnv("RADAR_INTERNAL_API_BASE_URL", "http://localhost:8443")
	// 本机开发：wisefido-sleepace 默认 8083，与 sleepace-dev.yaml 一致；生产用 env 覆盖
	cfg.SleepaceGateway.APIBaseURL = getEnv("SLEEPACE_GATEWAY_API_BASE_URL", "http://127.0.0.1:8083")
	cfg.Qinglan.APIBaseURL = getEnv("QINGLAN_API_BASE_URL", "http://127.0.0.1:8081")
	cfg.IoTTimeSeries.InternalAPIBaseURL = getEnv("IOT_TIMESERIES_INTERNAL_API_BASE_URL", "http://localhost:8085")
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
