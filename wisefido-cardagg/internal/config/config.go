package config

import (
	"fmt"
	"log"
	"os"

	commonconfig "owl-common/config"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DB      commonconfig.DatabaseConfig `yaml:"database"`
	Redis   commonconfig.RedisConfig    `yaml:"redis"`
	Logging commonconfig.LogConfig      `yaml:"logging"`
	Streams commonconfig.StreamsConfig  `yaml:"streams"`
}

func Load() (*Config, error) {
	path := "config.yaml"
	if v := os.Getenv("CONFIG_PATH"); v != "" {
		path = v
	}
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("(info) config file not found, using env: %v", err)
		return loadFromEnv(), nil
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	cfg.applyEnvOverrides()
	cfg.setDefaults()
	return &cfg, nil
}

func loadFromEnv() *Config {
	cfg := &Config{}
	cfg.applyEnvOverrides()
	cfg.setDefaults()
	return cfg
}

func (c *Config) applyEnvOverrides() {
	c.DB.LoadFromEnv("DB")
	if v := os.Getenv("DB_NAME"); v != "" {
		c.DB.Database = v
	}
	c.Redis.LoadFromEnv("REDIS")
	c.Logging.LoadFromEnv("LOG")
}

func (c *Config) setDefaults() {
	if c.DB.Host == "" {
		c.DB.Host = "localhost"
	}
	if c.DB.Port == 0 {
		c.DB.Port = 5432
	}
	if c.DB.User == "" {
		c.DB.User = "postgres"
	}
	if c.DB.Database == "" {
		c.DB.Database = "owl"
	}
	if c.Redis.Addr == "" {
		c.Redis.Addr = "localhost:6379"
	}
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "text"
	}
	if c.Logging.Output == "" {
		c.Logging.Output = "stdout"
	}
	if c.Logging.FilePath == "" {
		c.Logging.FilePath = "/var/log/wisefido-cardagg.log"
	}
	if c.Streams.Default.MaxLen == 0 {
		c.Streams.Default.MaxLen = 1000
	}
	if c.Streams.Default.RetentionSeconds == 0 {
		c.Streams.Default.RetentionSeconds = 86400
	}
}

func (c *Config) GetDatabaseURL() string { return c.DB.GetDSN() }
