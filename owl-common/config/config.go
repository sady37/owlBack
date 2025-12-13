package config

import (
	"fmt"
	"os"
	"time"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
	MaxConns int
	MaxIdle  int
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// MQTTConfig MQTT配置
type MQTTConfig struct {
	Broker   string
	ClientID string
	Username string
	Password string
	QoS      byte
}

// GetDatabaseDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode)
}

// LoadFromEnv 从环境变量加载配置
func (c *DatabaseConfig) LoadFromEnv(prefix string) {
	if host := os.Getenv(prefix + "_HOST"); host != "" {
		c.Host = host
	}
	if port := os.Getenv(prefix + "_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &c.Port)
	}
	if user := os.Getenv(prefix + "_USER"); user != "" {
		c.User = user
	}
	if password := os.Getenv(prefix + "_PASSWORD"); password != "" {
		c.Password = password
	}
	if database := os.Getenv(prefix + "_DATABASE"); database != "" {
		c.Database = database
	}
	if sslMode := os.Getenv(prefix + "_SSLMODE"); sslMode != "" {
		c.SSLMode = sslMode
	}
}

// LoadFromEnv 从环境变量加载Redis配置
func (c *RedisConfig) LoadFromEnv(prefix string) {
	if addr := os.Getenv(prefix + "_ADDR"); addr != "" {
		c.Addr = addr
	}
	if password := os.Getenv(prefix + "_PASSWORD"); password != "" {
		c.Password = password
	}
	if db := os.Getenv(prefix + "_DB"); db != "" {
		fmt.Sscanf(db, "%d", &c.DB)
	}
}

// LoadFromEnv 从环境变量加载MQTT配置
func (c *MQTTConfig) LoadFromEnv(prefix string) {
	if broker := os.Getenv(prefix + "_BROKER"); broker != "" {
		c.Broker = broker
	}
	if clientID := os.Getenv(prefix + "_CLIENT_ID"); clientID != "" {
		c.ClientID = clientID
	}
	if username := os.Getenv(prefix + "_USERNAME"); username != "" {
		c.Username = username
	}
	if password := os.Getenv(prefix + "_PASSWORD"); password != "" {
		c.Password = password
	}
}

// AlarmConfig 报警服务配置
type AlarmConfig struct {
	RuleBased struct {
		Enabled       bool
		CheckInterval time.Duration
		ConfigCacheTTL time.Duration
	}
	AI struct {
		Enabled            bool
		ModelPath          string
		CheckInterval      time.Duration
		HistoryWindow      time.Duration
		InspectionInterval time.Duration
		InspectionBatchSize int
		ConfidenceThreshold float64
	}
}

