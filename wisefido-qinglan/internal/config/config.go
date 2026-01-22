package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	commonconfig "owl-common/config"

	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	Server       ServerConfig                `yaml:"server"`
	MQTT         MQTTConfig                  `yaml:"mqtt"`
	HTTP         commonconfig.HTTPConfig     `yaml:"http"`
	HTTPS        HTTPSConfig                 `yaml:"https"` // HTTPS 服务器配置（用于设备认证）
	Redis        commonconfig.RedisConfig    `yaml:"redis"`
	DB           commonconfig.DatabaseConfig `yaml:"database"`
	Logging      commonconfig.LogConfig      `yaml:"logging"`
	Cache        commonconfig.CacheConfig    `yaml:"cache"`
	Streams      commonconfig.StreamsConfig  `yaml:"streams"`
	Subscription SubscriptionConfig          `yaml:"subscription"`
	Alarm        commonconfig.AlarmConfig    `yaml:"alarm"`
}

// HTTPSConfig HTTPS 服务器配置（用于设备认证）
type HTTPSConfig struct {
	Port     int    `yaml:"port"`      // HTTPS 服务器端口（默认 8443）
	CertFile string `yaml:"cert_file"` // TLS 证书文件路径
	KeyFile  string `yaml:"key_file"`  // TLS 私钥文件路径
}

// ServerConfig 服务器配置（服务特定配置）
type ServerConfig struct {
	Port    int    `yaml:"port"`
	Address string `yaml:"address"`
}

// MQTTConfig MQTT配置（扩展commonconfig.MQTTConfig）
type MQTTConfig struct {
	Broker   string `yaml:"broker"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	ClientID string `yaml:"client_id"`

	// 雷达设备MQTT主题配置（青兰特定）
	RadarDeviceMQTT RadarDeviceMQTTConfig `yaml:"radar_device_mqtt"`

	// 订阅主题（青兰特定）
	Topics TopicsConfig `yaml:"topics"`
}

// ToCommonConfig 转换为owl-common的MQTT配置
func (c *MQTTConfig) ToCommonConfig() *commonconfig.MQTTConfig {
	return &commonconfig.MQTTConfig{
		Broker:   fmt.Sprintf("tcp://%s:%d", c.Broker, c.Port),
		ClientID: c.ClientID,
		Username: c.Username,
		Password: c.Password,
		QoS:      1,
	}
}

// RadarDeviceMQTTConfig 雷达设备MQTT配置（青兰特定）
type RadarDeviceMQTTConfig struct {
	Server         string `yaml:"server"`           // MQTT 服务器地址
	Port           int    `yaml:"port"`             // MQTT 服务器端口
	Account        string `yaml:"account"`          // MQTT 账号
	Password       string `yaml:"password"`         // MQTT 密码
	Protocol       string `yaml:"protocol"`         // 协议类型（"1"=不加密, "2"=加密）
	Prefix         string `yaml:"prefix"`           // 主题前缀（可为空）
	ProductID      string `yaml:"product_id"`       // 产品 ID (0-255)
	Timeout        int    `yaml:"timeout"`          // 连接超时（秒）
	Keepalive      int    `yaml:"keepalive"`        // 心跳间隔（秒）
	ClientIDPrefix string `yaml:"client_id_prefix"` // ClientID 前缀（可选）
}

// TopicsConfig 主题配置（青兰特定）
type TopicsConfig struct {
	Property string `yaml:"property"` // 属性响应主题
	Monitor  string `yaml:"monitor"`  // 实时数据主题
	Function string `yaml:"function"` // 功能响应主题
	Stat     string `yaml:"stat"`     // 统计数据主题
	Event    string `yaml:"event"`    // 事件主题
	Alarm    string `yaml:"alarm"`    // 告警主题
}

// SubscriptionConfig 订阅配置（青兰特定）
type SubscriptionConfig struct {
	AutoSubscribe   bool `yaml:"auto_subscribe"`   // 自动订阅
	DefaultDuration int  `yaml:"default_duration"` // 默认订阅时长（秒）
}

// GetStreamConfig 获取流配置
func (c *Config) GetStreamConfig(streamName string) (maxLen int64, retentionSeconds int) {
	// 使用默认配置
	if c.Streams.Default.MaxLen > 0 {
		maxLen = c.Streams.Default.MaxLen
	} else {
		maxLen = 1000 // 默认值
	}

	if c.Streams.Default.RetentionSeconds > 0 {
		retentionSeconds = c.Streams.Default.RetentionSeconds
	} else {
		retentionSeconds = 86400 // 24小时默认值
	}

	// 检查是否有特定stream的配置
	if c.Streams.Streams != nil {
		if streamConfig, ok := c.Streams.Streams[streamName]; ok {
			if streamConfig.MaxLen > 0 {
				maxLen = streamConfig.MaxLen
			}
			if streamConfig.RetentionSeconds > 0 {
				retentionSeconds = streamConfig.RetentionSeconds
			}
		}
	}

	return maxLen, retentionSeconds
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
		return nil, fmt.Errorf("failed to read config file: %w", err)
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

// setDefaults 设置默认值
func (c *Config) setDefaults() {
	// 日志配置默认值
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
		c.Logging.FilePath = "/var/log/wisefido-qinglan.log"
	}

	// 缓存配置默认值
	if c.Cache.DefaultTTL == 0 {
		c.Cache.DefaultTTL = 300 * time.Second // 5分钟
	}
	if c.Cache.DeviceTTL == 0 {
		c.Cache.DeviceTTL = 300 * time.Second // 5分钟
	}
	if c.Cache.LocationTTL == 0 {
		c.Cache.LocationTTL = 600 * time.Second // 10分钟
	}

	// Stream配置默认值
	if c.Streams.Default.MaxLen == 0 {
		c.Streams.Default.MaxLen = 1000
	}
	if c.Streams.Default.RetentionSeconds == 0 {
		c.Streams.Default.RetentionSeconds = 86400 // 24小时
	}

	// 订阅配置默认值
	if c.Subscription.DefaultDuration == 0 {
		c.Subscription.DefaultDuration = 3600 // 1小时
	}

	// MQTT主题默认值
	if c.MQTT.Topics.Property == "" {
		c.MQTT.Topics.Property = "/prop/88/+/post"
	}
	if c.MQTT.Topics.Monitor == "" {
		c.MQTT.Topics.Monitor = "/monitor/88/+/post"
	}
	if c.MQTT.Topics.Function == "" {
		c.MQTT.Topics.Function = "/func/88/+/post"
	}
	if c.MQTT.Topics.Stat == "" {
		c.MQTT.Topics.Stat = "/stat/88/+/post"
	}
	if c.MQTT.Topics.Event == "" {
		c.MQTT.Topics.Event = "/event/88/+/post"
	}
	if c.MQTT.Topics.Alarm == "" {
		c.MQTT.Topics.Alarm = "/alarm/88/+/post"
	}
}

// LoadFromEnv 从环境变量加载配置
func LoadFromEnv() (*Config, error) {
	cfg := &Config{}

	// 数据库配置
	cfg.DB.LoadFromEnv("DB")
	// 兼容 DB_NAME（与 wisefido-data 保持一致）
	if cfg.DB.Database == "" {
		cfg.DB.Database = getEnv("DB_NAME", "owlrd")
	}
	if cfg.DB.SSLMode == "" {
		cfg.DB.SSLMode = getEnv("DB_SSLMODE", "disable")
	}

	// Redis配置
	cfg.Redis.LoadFromEnv("REDIS")

	// MQTT配置
	cfg.MQTT.Broker = getEnv("MQTT_BROKER", "127.0.0.1")
	cfg.MQTT.Port = parseInt(getEnv("MQTT_PORT", "1883"), 1883)
	cfg.MQTT.ClientID = getEnv("MQTT_CLIENT_ID", "wisefido-qinglan")
	cfg.MQTT.Username = getEnv("MQTT_USERNAME", "")
	cfg.MQTT.Password = getEnv("MQTT_PASSWORD", "")

	// 雷达设备MQTT配置
	// 雷达设备 MQTT 配置（返回给设备的配置）
	cfg.MQTT.RadarDeviceMQTT.Server = getEnv("RADAR_MQTT_SERVER", "127.0.0.1")
	// 检查服务器地址是否为本地回环地址（设备无法连接）
	if cfg.MQTT.RadarDeviceMQTT.Server == "127.0.0.1" || cfg.MQTT.RadarDeviceMQTT.Server == "localhost" || cfg.MQTT.RadarDeviceMQTT.Server == "::1" {
		log.Printf("⚠️ WARNING: RADAR_MQTT_SERVER is set to %s (localhost), devices cannot connect! Please set RADAR_MQTT_SERVER environment variable to a reachable IP address or hostname", cfg.MQTT.RadarDeviceMQTT.Server)
	}
	cfg.MQTT.RadarDeviceMQTT.Port = parseInt(getEnv("RADAR_MQTT_PORT", "8883"), 8883) // 默认加密端口
	cfg.MQTT.RadarDeviceMQTT.Account = getEnv("RADAR_MQTT_ACCOUNT", "wfiot")
	cfg.MQTT.RadarDeviceMQTT.Password = getEnv("RADAR_MQTT_PASSWORD", "")
	cfg.MQTT.RadarDeviceMQTT.Protocol = getEnv("RADAR_MQTT_PROTOCOL", "2") // 默认加密
	cfg.MQTT.RadarDeviceMQTT.Prefix = getEnv("RADAR_MQTT_PREFIX", "")
	cfg.MQTT.RadarDeviceMQTT.ProductID = getEnv("RADAR_MQTT_PRODUCT_ID", "88")
	cfg.MQTT.RadarDeviceMQTT.Timeout = parseInt(getEnv("RADAR_MQTT_TIMEOUT", "30"), 30)
	cfg.MQTT.RadarDeviceMQTT.Keepalive = parseInt(getEnv("RADAR_MQTT_KEEPALIVE", "60"), 60)
	cfg.MQTT.RadarDeviceMQTT.ClientIDPrefix = getEnv("RADAR_MQTT_CLIENT_ID_PREFIX", "radar")

	// HTTP配置
	cfg.HTTP.LoadFromEnv("HTTP")

	// HTTPS配置（用于设备认证）
	// 优先使用 owl-common 的共享证书（如果存在）
	cfg.HTTPS.Port = parseInt(getEnv("QINGLAN_HTTPS_PORT", "8443"), 8443)

	// 如果未指定证书路径，尝试使用 owl-common 的证书
	certFile := getEnv("QINGLAN_HTTPS_CERT_FILE", "")
	keyFile := getEnv("QINGLAN_HTTPS_KEY_FILE", "")

	if certFile == "" || keyFile == "" {
		// 尝试查找 owl-common 的证书
		commonCertFile := "../owl-common/server.crt"
		commonKeyFile := "../owl-common/server.key"
		if _, err := os.Stat(commonCertFile); err == nil {
			if _, err := os.Stat(commonKeyFile); err == nil {
				certFile = commonCertFile
				keyFile = commonKeyFile
			}
		}
	}

	cfg.HTTPS.CertFile = certFile
	cfg.HTTPS.KeyFile = keyFile

	// 日志配置
	cfg.Logging.LoadFromEnv("LOG")

	// 设置默认值
	cfg.setDefaults()

	return cfg, nil
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseInt 解析整数，如果解析失败则返回默认值
func parseInt(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}
	return val
}
