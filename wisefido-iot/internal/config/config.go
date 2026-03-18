package config

import (
	"log"
	"os"
	"owl-common/config"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config IoT 时序数据服务配置
type Config struct {
	Database config.DatabaseConfig `yaml:"database"`
	Redis    config.RedisConfig    `yaml:"redis"`

	Streams struct {
		Monitor string `yaml:"monitor"`
		Stat    string `yaml:"stat"`
		Event   string `yaml:"event"`
		Alarm   string `yaml:"alarm"`
		Auth    string `yaml:"auth"`
	} `yaml:"streams"`
	ConsumerGroup string `yaml:"consumer_group"`
	ConsumerName  string `yaml:"consumer_name"`
	BatchSize     int64  `yaml:"batch_size"`

	Log struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
	} `yaml:"logging"`
}

// Load 加载配置
func Load() (*Config, error) {
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
	return &cfg, nil
}

func (c *Config) setDefaults() {
	if c.Database.SSLMode == "" {
		c.Database.SSLMode = "disable"
	}
	if c.Streams.Monitor == "" {
		c.Streams.Monitor = "iot:monitor:stream"
	}
	if c.Streams.Stat == "" {
		c.Streams.Stat = "iot:stat:stream"
	}
	if c.Streams.Event == "" {
		c.Streams.Event = "iot:event:stream"
	}
	if c.Streams.Alarm == "" {
		c.Streams.Alarm = "iot:alarm:stream"
	}
	if c.Streams.Auth == "" {
		c.Streams.Auth = "iot:auth:stream"
	}
	if c.ConsumerGroup == "" {
		c.ConsumerGroup = "iot-timeseries-group"
	}
	if c.ConsumerName == "" {
		c.ConsumerName = "iot-timeseries-1"
	}
	if c.BatchSize == 0 {
		c.BatchSize = 10
	}
	if c.Log.Level == "" {
		c.Log.Level = "info"
	}
	if c.Log.Format == "" {
		c.Log.Format = "json"
	}
}

func LoadFromEnv() (*Config, error) {
	cfg := &Config{}
	cfg.Database.Host = getEnv("DB_HOST", "localhost")
	if portStr := getEnv("DB_PORT", ""); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil && port > 0 {
			cfg.Database.Port = port
		} else {
			cfg.Database.Port = 5432
		}
	} else {
		cfg.Database.Port = 5432
	}
	cfg.Database.User = getEnv("DB_USER", "postgres")
	cfg.Database.Password = getEnv("DB_PASSWORD", "")
	cfg.Database.Database = getEnv("DB_NAME", "owlrd")
	cfg.Database.SSLMode = getEnv("DB_SSLMODE", "disable")
	cfg.Redis.Addr = getEnv("REDIS_ADDR", "localhost:6379")
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "")
	cfg.Redis.DB = 0
	cfg.Streams.Monitor = getEnv("STREAM_IOT_MONITOR", "iot:monitor:stream")
	cfg.Streams.Stat = getEnv("STREAM_IOT_STAT", "iot:stat:stream")
	cfg.Streams.Event = getEnv("STREAM_IOT_EVENT", "iot:event:stream")
	cfg.Streams.Alarm = getEnv("STREAM_IOT_ALARM", "iot:alarm:stream")
	cfg.Streams.Auth = getEnv("STREAM_IOT_AUTH", "iot:auth:stream")
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
