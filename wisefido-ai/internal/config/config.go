package config

import (
	"log"
	"os"
	"owl-common/config"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config 报警服务配置
type Config struct {
	Database config.DatabaseConfig `yaml:"database"`
	Redis    config.RedisConfig    `yaml:"redis"`

	Alarm struct {
		Cache struct {
			RealtimeKeyPrefix string `yaml:"realtime_key_prefix"`
			RealtimeSuffix    string `yaml:"realtime_suffix"`
			AlarmKeyPrefix    string `yaml:"alarm_key_prefix"`
			AlarmSuffix       string `yaml:"alarm_suffix"`
			AlarmTTL          int    `yaml:"alarm_ttl"`
			StateKeyPrefix    string `yaml:"state_key_prefix"`
		} `yaml:"cache"`
		PollInterval int `yaml:"poll_interval"`
		Evaluation   struct {
			BatchSize int `yaml:"batch_size"`
		} `yaml:"evaluation"`
		IoTStream struct {
			Enabled       bool   `yaml:"enabled"`
			Monitor       string `yaml:"monitor"`
			Stat          string `yaml:"stat"`
			Event         string `yaml:"event"`
			Alarm         string `yaml:"alarm"`
			ConsumerGroup string `yaml:"consumer_group"`
			ConsumerName  string `yaml:"consumer_name"`
			BatchSize     int64  `yaml:"batch_size"`
		} `yaml:"iot_stream"`
	} `yaml:"alarm"`

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
	if c.Alarm.Cache.RealtimeKeyPrefix == "" {
		c.Alarm.Cache.RealtimeKeyPrefix = "vital-focus:card:"
	}
	if c.Alarm.Cache.RealtimeSuffix == "" {
		c.Alarm.Cache.RealtimeSuffix = ":realtime"
	}
	if c.Alarm.Cache.AlarmKeyPrefix == "" {
		c.Alarm.Cache.AlarmKeyPrefix = "vital-focus:card:"
	}
	if c.Alarm.Cache.AlarmSuffix == "" {
		c.Alarm.Cache.AlarmSuffix = ":alarms"
	}
	if c.Alarm.Cache.AlarmTTL == 0 {
		c.Alarm.Cache.AlarmTTL = 30
	}
	if c.Alarm.Cache.StateKeyPrefix == "" {
		c.Alarm.Cache.StateKeyPrefix = "alarm:state:"
	}
	if c.Alarm.PollInterval == 0 {
		c.Alarm.PollInterval = 10
	}
	if c.Alarm.Evaluation.BatchSize == 0 {
		c.Alarm.Evaluation.BatchSize = 10
	}
	if c.Alarm.IoTStream.Monitor == "" {
		c.Alarm.IoTStream.Monitor = "iot:monitor:stream"
	}
	if c.Alarm.IoTStream.Stat == "" {
		c.Alarm.IoTStream.Stat = "iot:stat:stream"
	}
	if c.Alarm.IoTStream.Event == "" {
		c.Alarm.IoTStream.Event = "iot:event:stream"
	}
	if c.Alarm.IoTStream.Alarm == "" {
		c.Alarm.IoTStream.Alarm = "iot:alarm:stream"
	}
	if c.Alarm.IoTStream.ConsumerGroup == "" {
		c.Alarm.IoTStream.ConsumerGroup = "wisefido-alarm-events"
	}
	if c.Alarm.IoTStream.ConsumerName == "" {
		c.Alarm.IoTStream.ConsumerName = "alarm-consumer-1"
	}
	if c.Alarm.IoTStream.BatchSize == 0 {
		c.Alarm.IoTStream.BatchSize = 10
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
	cfg.Alarm.Cache.RealtimeKeyPrefix = getEnv("CACHE_REALTIME_PREFIX", "vital-focus:card:")
	cfg.Alarm.Cache.RealtimeSuffix = ":realtime"
	cfg.Alarm.Cache.AlarmKeyPrefix = getEnv("CACHE_ALARM_PREFIX", "vital-focus:card:")
	cfg.Alarm.Cache.AlarmSuffix = ":alarms"
	cfg.Alarm.Cache.AlarmTTL = 30
	cfg.Alarm.Cache.StateKeyPrefix = getEnv("CACHE_STATE_PREFIX", "alarm:state:")
	cfg.Alarm.PollInterval = 10
	cfg.Alarm.Evaluation.BatchSize = 10
	cfg.Alarm.IoTStream.Enabled = getEnv("AI_IOT_STREAM_ENABLED", "true") == "true"
	cfg.Alarm.IoTStream.Monitor = getEnv("AI_STREAM_MONITOR", "iot:monitor:stream")
	cfg.Alarm.IoTStream.Stat = getEnv("AI_STREAM_STAT", "iot:stat:stream")
	cfg.Alarm.IoTStream.Event = getEnv("AI_STREAM_EVENT", "iot:event:stream")
	cfg.Alarm.IoTStream.Alarm = getEnv("AI_STREAM_ALARM", "iot:alarm:stream")
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
