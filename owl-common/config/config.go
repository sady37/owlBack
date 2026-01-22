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

// HTTPConfig HTTP服务器配置
type HTTPConfig struct {
	Host string
	Port int
	// HTTPS相关配置
	EnableTLS bool
	CertFile  string
	KeyFile   string
}

// LogConfig 日志配置
type LogConfig struct {
	Level    string // debug, info, warn, error
	Format   string // json, text
	Output   string // stdout, file
	FilePath string // 日志文件路径
}

// CacheConfig 缓存配置
type CacheConfig struct {
	DefaultTTL time.Duration // 默认缓存TTL
	// 特定类型缓存TTL
	DeviceTTL   time.Duration // 设备缓存TTL
	LocationTTL time.Duration // 位置信息缓存TTL
	UserTTL     time.Duration // 用户缓存TTL
}

// StreamConfig Redis Stream配置
type StreamConfig struct {
	MaxLen           int64 // Stream最大长度
	RetentionSeconds int   // 数据保留秒数
}

// StreamsConfig 多个Stream配置
type StreamsConfig struct {
	Default StreamConfig
	Streams map[string]StreamConfig // 特定stream的配置
}

// AlarmConfig 报警服务配置
type AlarmConfig struct {
	RuleBased struct {
		Enabled        bool
		CheckInterval  time.Duration
		ConfigCacheTTL time.Duration
	}
	AI struct {
		Enabled             bool
		ModelPath           string
		CheckInterval       time.Duration
		HistoryWindow       time.Duration
		InspectionInterval  time.Duration
		InspectionBatchSize int
		ConfidenceThreshold float64
	}
}

// GetDatabaseDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	sslMode := c.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, sslMode)
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
	sslMode := os.Getenv(prefix + "_SSLMODE")
	if sslMode != "" {
		c.SSLMode = sslMode
	}
	if c.SSLMode == "" {
		c.SSLMode = "disable"
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

// LoadFromEnv 从环境变量加载HTTP配置
func (c *HTTPConfig) LoadFromEnv(prefix string) {
	if host := os.Getenv(prefix + "_HOST"); host != "" {
		c.Host = host
	}
	if port := os.Getenv(prefix + "_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &c.Port)
	}
	if enableTLS := os.Getenv(prefix + "_ENABLE_TLS"); enableTLS != "" {
		c.EnableTLS = enableTLS == "true"
	}
	if certFile := os.Getenv(prefix + "_CERT_FILE"); certFile != "" {
		c.CertFile = certFile
	}
	if keyFile := os.Getenv(prefix + "_KEY_FILE"); keyFile != "" {
		c.KeyFile = keyFile
	}
}

// LoadFromEnv 从环境变量加载日志配置
func (c *LogConfig) LoadFromEnv(prefix string) {
	if level := os.Getenv(prefix + "_LEVEL"); level != "" {
		c.Level = level
	}
	if format := os.Getenv(prefix + "_FORMAT"); format != "" {
		c.Format = format
	}
	if output := os.Getenv(prefix + "_OUTPUT"); output != "" {
		c.Output = output
	}
	if filePath := os.Getenv(prefix + "_FILE_PATH"); filePath != "" {
		c.FilePath = filePath
	}
}

// GetAddr 获取HTTP地址字符串
func (c *HTTPConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetBrokerURL 获取MQTT Broker URL
func (c *MQTTConfig) GetBrokerURL() string {
	return c.Broker
}
