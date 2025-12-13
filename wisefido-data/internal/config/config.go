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
