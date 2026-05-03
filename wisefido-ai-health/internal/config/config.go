package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	commonconfig "owl-common/config"

	"gopkg.in/yaml.v3"
)

// Config wisefido-ai-health 应用配置
type Config struct {
	DB       commonconfig.DatabaseConfig `yaml:"database"`
	Logging  commonconfig.LogConfig      `yaml:"logging"`
	Schedule ScheduleConfig              `yaml:"schedule"`
	ETL      ETLConfig                   `yaml:"etl"`
}

// ScheduleConfig 调度配置（stdlib timer 实现，无 cron 库依赖）
type ScheduleConfig struct {
	Timezone    string `yaml:"timezone"`     // IANA tz, e.g. "America/Denver"
	DailyAt     string `yaml:"daily_at"`     // "HH:MM" 24h
	MonthlyDay  int    `yaml:"monthly_day"`  // 1–28（>28 月份兼容性）
	MonthlyAt   string `yaml:"monthly_at"`   // "HH:MM" 24h
}

// ETLConfig ETL 运行参数
type ETLConfig struct {
	Parallel             int  `yaml:"parallel"`
	PerCardTimeoutSec    int  `yaml:"per_card_timeout_sec"`
	WatermarkLookbackDay int  `yaml:"watermark_lookback_days"`
	DryRun               bool `yaml:"dry_run"`
}

// Load 加载配置（优先 config.yaml，缺失则降级 env）
func Load() (*Config, error) {
	path := "config.yaml"
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		path = p
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// 配置文件不存在 → 走环境变量（容器部署常态）
		return loadFromEnv()
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	cfg.applyDefaults()
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func loadFromEnv() (*Config, error) {
	cfg := &Config{}
	cfg.DB.LoadFromEnv("DB")
	if name := os.Getenv("DB_NAME"); name != "" && cfg.DB.Database == "" {
		cfg.DB.Database = name
	}
	cfg.Logging.LoadFromEnv("LOG")

	cfg.Schedule.Timezone = envOr("SCHEDULE_TIMEZONE", "America/Denver")
	cfg.Schedule.DailyAt = envOr("SCHEDULE_DAILY_AT", "02:00")
	cfg.Schedule.MonthlyDay = envOrInt("SCHEDULE_MONTHLY_DAY", 1)
	cfg.Schedule.MonthlyAt = envOr("SCHEDULE_MONTHLY_AT", "03:00")

	cfg.ETL.Parallel = envOrInt("ETL_PARALLEL", 8)
	cfg.ETL.PerCardTimeoutSec = envOrInt("ETL_PER_CARD_TIMEOUT_SEC", 60)
	cfg.ETL.WatermarkLookbackDay = envOrInt("ETL_WATERMARK_LOOKBACK_DAYS", 7)
	cfg.ETL.DryRun = strings.EqualFold(os.Getenv("ETL_DRY_RUN"), "true")

	cfg.applyDefaults()
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) applyDefaults() {
	if c.DB.Host == "" {
		c.DB.Host = "127.0.0.1"
	}
	if c.DB.Port == 0 {
		c.DB.Port = 5432
	}
	if c.DB.User == "" {
		c.DB.User = "postgres"
	}
	if c.DB.Database == "" {
		c.DB.Database = "owlrd"
	}
	if c.DB.SSLMode == "" {
		c.DB.SSLMode = "disable"
	}
	if c.DB.MaxConns == 0 {
		c.DB.MaxConns = 8
	}
	if c.DB.MaxIdle == 0 {
		c.DB.MaxIdle = 4
	}
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Schedule.Timezone == "" {
		c.Schedule.Timezone = "America/Denver"
	}
	if c.Schedule.DailyAt == "" {
		c.Schedule.DailyAt = "02:00"
	}
	if c.Schedule.MonthlyDay == 0 {
		c.Schedule.MonthlyDay = 1
	}
	if c.Schedule.MonthlyAt == "" {
		c.Schedule.MonthlyAt = "03:00"
	}
	if c.ETL.Parallel == 0 {
		c.ETL.Parallel = 8
	}
	if c.ETL.PerCardTimeoutSec == 0 {
		c.ETL.PerCardTimeoutSec = 60
	}
	if c.ETL.WatermarkLookbackDay == 0 {
		c.ETL.WatermarkLookbackDay = 7
	}
}

func (c *Config) validate() error {
	if _, _, err := parseHHMM(c.Schedule.DailyAt); err != nil {
		return fmt.Errorf("schedule.daily_at: %w", err)
	}
	if _, _, err := parseHHMM(c.Schedule.MonthlyAt); err != nil {
		return fmt.Errorf("schedule.monthly_at: %w", err)
	}
	if c.Schedule.MonthlyDay < 1 || c.Schedule.MonthlyDay > 28 {
		return fmt.Errorf("schedule.monthly_day must be 1..28 (got %d)", c.Schedule.MonthlyDay)
	}
	return nil
}

// ParseHHMM exported helper used by scheduler
func ParseHHMM(s string) (int, int, error) { return parseHHMM(s) }

func parseHHMM(s string) (int, int, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expect HH:MM, got %q", s)
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil || h < 0 || h > 23 {
		return 0, 0, fmt.Errorf("hour out of range: %q", s)
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil || m < 0 || m > 59 {
		return 0, 0, fmt.Errorf("minute out of range: %q", s)
	}
	return h, m, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envOrInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	if n, err := strconv.Atoi(v); err == nil {
		return n
	}
	return def
}
