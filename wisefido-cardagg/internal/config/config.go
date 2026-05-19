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
	AI      AIConfig                    `yaml:"ai"`
}

// AIConfig — AI 派生 verdict 合并行为配置。
//
// 部署约定（[[deploy_mode_by_domain]]）：
//   - test.wisefido.com  → mode=sandbox（仅 log，不动 track_confidence；演示"AI 在思考"）
//   - owl.wisefido.com   → mode=release（覆写 track_confidence + ai_source，UI 渲染 AI 调整后值）
//
// override 优先级：env CARDAGG_AI_OVERRIDE_MODE > config.yaml ai.override_mode > 默认 sandbox。
type AIConfig struct {
	OverrideMode  string `yaml:"override_mode"`     // "sandbox" | "release"
	OverrideTTLSec int   `yaml:"override_ttl_sec"`  // verdict cache TTL；≤0 默认 60s
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
	if v := os.Getenv("CARDAGG_AI_OVERRIDE_MODE"); v != "" {
		c.AI.OverrideMode = v
	}
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
	// AI 默认 sandbox + 60s TTL；生产 deploy 通过 yaml/env 切 release
	if c.AI.OverrideMode == "" {
		c.AI.OverrideMode = "sandbox"
	}
	if c.AI.OverrideTTLSec <= 0 {
		c.AI.OverrideTTLSec = 60
	}
}

func (c *Config) GetDatabaseURL() string { return c.DB.GetDSN() }
