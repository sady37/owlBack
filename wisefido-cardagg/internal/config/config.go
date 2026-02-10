package config

import (
	"fmt"
	"log"
	"os"

	commonconfig "owl-common/config"

	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	DB      commonconfig.DatabaseConfig `yaml:"database"`
	Redis   commonconfig.RedisConfig    `yaml:"redis"`
	Logging commonconfig.LogConfig      `yaml:"logging"`
	Streams commonconfig.StreamsConfig  `yaml:"streams"`
}

// Load 加载配置
func Load() (*Config, error) {
	// 默认配置文件路径
	configPath := "config.yaml"
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		configPath = path
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		// 配置文件不存在或无法读取，降级为从环境变量加载
		log.Printf("(info) config file not found, using environment variables: %v", err)
		return LoadFromEnv()
	}

	// 解析YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 设置默认值
	cfg.setDefaults()

	return &cfg, nil
}

// LoadFromEnv 从环境变量加载配置
func LoadFromEnv() (*Config, error) {
	cfg := &Config{}

	// 数据库配置
	cfg.DB.LoadFromEnv("DB")
	// 兼容 DB_NAME（与 wisefido-data 保持一致）
	if cfg.DB.Database == "" {
		cfg.DB.Database = getEnv("DB_NAME", "owl")
	}
	if cfg.DB.Host == "" {
		cfg.DB.Host = "localhost"
	}
	if cfg.DB.Port == 0 {
		cfg.DB.Port = 5432
	}
	if cfg.DB.User == "" {
		cfg.DB.User = "postgres"
	}

	// Redis配置
	cfg.Redis.LoadFromEnv("REDIS")
	// 提供默认值
	if cfg.Redis.Addr == "" {
		cfg.Redis.Addr = "localhost:6379"
	}

	// 日志配置
	cfg.Logging.LoadFromEnv("LOG")
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "text"
	}
	if cfg.Logging.Output == "" {
		cfg.Logging.Output = "stdout"
	}
	if cfg.Logging.FilePath == "" {
		cfg.Logging.FilePath = "/var/log/wisefido-cardagg.log"
	}

	// 设置默认值
	cfg.setDefaults()

	return cfg, nil
}

// setDefaults 设置默认值
func (c *Config) setDefaults() {
	// Stream 配置默认值
	if c.Streams.Default.MaxLen == 0 {
		c.Streams.Default.MaxLen = 1000
	}
	if c.Streams.Default.RetentionSeconds == 0 {
		c.Streams.Default.RetentionSeconds = 86400 // 24小时
	}
}

// GetDatabaseURL 获取数据库连接串（PostgreSQL）
func (c *Config) GetDatabaseURL() string {
	return c.DB.GetDSN()
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
