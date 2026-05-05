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

	AIOverride AIOverrideConfig `yaml:"ai_override"`
}

// AIOverrideConfig wisefido-sensor 派生 track_verdict 在 cardagg 端的合并行为。
//
// 仅作用于 UI 合并（调整 track_confidence 字段供前端渲染），**绝不影响 alarm
// 触发路径**——用户原则"宁可误报不可漏报"。
//
//	mode = "sandbox"（默认）：仅 log，track_confidence 不变。演示"AI 在思考"。
//	mode = "release"：把 AI 写的 confidence 覆写到 monitor 流 track 字段。
//
// env 覆盖：CARDAGG_AI_OVERRIDE_MODE / CARDAGG_AI_OVERRIDE_TTL_SEC
type AIOverrideConfig struct {
	Mode   string `yaml:"mode"`    // "sandbox" | "release"
	TTLSec int    `yaml:"ttl_sec"` // 缓存条目兜底过期秒数
	GCSec  int    `yaml:"gc_sec"`  // GC 间隔秒数（≤0 取默认 30s）
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
	// AI override 默认 sandbox / 60s TTL / 30s GC，env 覆盖优先
	if v := os.Getenv("CARDAGG_AI_OVERRIDE_MODE"); v != "" {
		c.AIOverride.Mode = v
	}
	if v := os.Getenv("CARDAGG_AI_OVERRIDE_TTL_SEC"); v != "" {
		var n int
		_, _ = fmt.Sscanf(v, "%d", &n)
		if n > 0 {
			c.AIOverride.TTLSec = n
		}
	}
	if c.AIOverride.Mode == "" {
		c.AIOverride.Mode = "sandbox"
	}
	if c.AIOverride.TTLSec <= 0 {
		c.AIOverride.TTLSec = 60
	}
	if c.AIOverride.GCSec <= 0 {
		c.AIOverride.GCSec = 30
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
