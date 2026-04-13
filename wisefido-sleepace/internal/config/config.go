package config

import (
	"fmt"
	"os"
	"strings"

	commonconfig "owl-common/config"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Service  ServiceConfig              `yaml:"service"`
	Log      commonconfig.LogConfig     `yaml:"log"`
	DB       commonconfig.DatabaseConfig `yaml:"database"`
	Redis    commonconfig.RedisConfig   `yaml:"redis"`
	MQTT     MQTTConfig                 `yaml:"mqtt"`
	Sleepace SleepaceConfig             `yaml:"sleepace"`
	HTTP     commonconfig.HTTPConfig    `yaml:"http"`
	Streams  commonconfig.StreamsConfig `yaml:"streams"`
}

type ServiceConfig struct {
	Name string `yaml:"name"`
}

type MQTTConfig struct {
	Broker   string `yaml:"broker"`
	ClientID string `yaml:"client_id"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	TopicID  string `yaml:"topic_id"`
}

type SleepaceConfig struct {
	HTTPAddress      string `yaml:"http_address"`
	AppID            string `yaml:"app_id"`
	ChannelID        string `yaml:"channel_id"`
	SecretKey        string `yaml:"secret_key"`
	Timezone         int    `yaml:"timezone"`
	RealtimeInterval int    `yaml:"realtime_interval"`
	LeaveSensibility int    `yaml:"leave_sensibility"`
	ReportUploadType int    `yaml:"report_upload_type"`
	ReportUploadTime int    `yaml:"report_upload_time"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	// Allow overriding log settings from environment (LOG_LEVEL, LOG_FORMAT, etc.)
	cfg.Log.LoadFromEnv("LOG")
	cfg.DB.LoadFromEnv("DB")
	cfg.Redis.LoadFromEnv("REDIS")
	cfg.MQTT.loadFromEnv()
	return cfg, nil
}

func (c *MQTTConfig) loadFromEnv() {
	if v := os.Getenv("MQTT_BROKER"); v != "" {
		c.Broker = v
		if !strings.Contains(c.Broker, ":") {
			port := os.Getenv("MQTT_PORT")
			if port == "" {
				port = "1883"
			}
			c.Broker = c.Broker + ":" + port
		}
	}
	if v := os.Getenv("MQTT_CLIENT_ID"); v != "" {
		c.ClientID = v
	}
	if v := os.Getenv("MQTT_USERNAME"); v != "" {
		c.Username = v
	}
	if v := os.Getenv("MQTT_PASSWORD"); v != "" {
		c.Password = v
	}
}

func (c *Config) GetStreamConfig(streamName string) (maxLen int64, retentionSeconds int) {
	if c.Streams.Streams != nil {
		if sc, ok := c.Streams.Streams[streamName]; ok {
			return sc.MaxLen, sc.RetentionSeconds
		}
	}
	if c.Streams.Default.MaxLen > 0 {
		return c.Streams.Default.MaxLen, c.Streams.Default.RetentionSeconds
	}
	return 1000, 0
}
