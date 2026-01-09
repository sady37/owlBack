package config

import (
	"os"
	"strconv"
	"owl-common/config"
)

// Config 雷达服务配置
type Config struct {
	Database config.DatabaseConfig
	Redis    config.RedisConfig
	MQTT     config.MQTTConfig // 用于 wisefido-radar 服务连接 MQTT
	
	// HTTPS 服务器配置
	HTTPS struct {
		Port     int    // HTTPS 服务器端口（默认 443）
		CertFile string // TLS 证书文件路径
		KeyFile  string // TLS 私钥文件路径
	}
	
	// 雷达服务特定配置
	Radar struct {
		// MQTT 配置（返回给设备的配置）
		DeviceMQTT struct {
			Server        string // MQTT 服务器地址
			Port          int    // MQTT 服务器端口
			Account       string // MQTT 账号
			Password      string // MQTT 密码
			Protocol      string // 协议类型（"1"=不加密, "2"=加密）
			Prefix        string // 主题前缀（可为空）
			ProductID     string // 产品 ID (0-255)
			Timeout       int    // 连接超时（秒）
			Keepalive     int    // 心跳间隔（秒）
			ClientIDPrefix string // ClientID 前缀（可选）
		}
		
		Topics struct {
			Data    string // 数据主题，如 "radar/+/data"
			Command string // 命令主题，如 "radar/+/command"
			OTA     string // OTA主题，如 "radar/+/ota"
		}
		OTA struct {
			Enabled        bool
			FirmwarePath   string // 固件文件路径
			CheckInterval  string // 检查间隔
		}
	}
	
	Log struct {
		Level  string
		Format string
	}
}

// Load 加载配置
func Load() (*Config, error) {
	cfg := &Config{}
	
	// 从环境变量加载（默认值）
	cfg.Database.Host = getEnv("DB_HOST", "localhost")
	cfg.Database.Port = 5432
	cfg.Database.User = getEnv("DB_USER", "postgres")
	cfg.Database.Password = getEnv("DB_PASSWORD", "postgres")
	cfg.Database.Database = getEnv("DB_NAME", "owlrd")
	cfg.Database.SSLMode = getEnv("DB_SSLMODE", "disable")
	
	cfg.Redis.Addr = getEnv("REDIS_ADDR", "localhost:6379")
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "")
	cfg.Redis.DB = 0
	
	cfg.MQTT.Broker = getEnv("MQTT_BROKER", "tcp://localhost:1883")
	cfg.MQTT.ClientID = getEnv("MQTT_CLIENT_ID", "wisefido-radar")
	cfg.MQTT.Username = getEnv("MQTT_USERNAME", "")
	cfg.MQTT.Password = getEnv("MQTT_PASSWORD", "")
	
	// HTTPS 服务器配置
	cfg.HTTPS.Port = parseInt(getEnv("RADAR_HTTPS_PORT", "443"), 443)
	cfg.HTTPS.CertFile = getEnv("RADAR_HTTPS_CERT_FILE", "")
	cfg.HTTPS.KeyFile = getEnv("RADAR_HTTPS_KEY_FILE", "")
	
	// 雷达设备 MQTT 配置（返回给设备的配置）
	cfg.Radar.DeviceMQTT.Server = getEnv("RADAR_MQTT_SERVER", "192.168.2.177")
	cfg.Radar.DeviceMQTT.Port = parseInt(getEnv("RADAR_MQTT_PORT", "8883"), 8883)
	cfg.Radar.DeviceMQTT.Account = getEnv("RADAR_MQTT_ACCOUNT", "wfiot")
	cfg.Radar.DeviceMQTT.Password = getEnv("RADAR_MQTT_PASSWORD", "")
	cfg.Radar.DeviceMQTT.Protocol = getEnv("RADAR_MQTT_PROTOCOL", "2") // 默认加密
	cfg.Radar.DeviceMQTT.Prefix = getEnv("RADAR_MQTT_PREFIX", "")      // 默认空（根据测试文档，当前固件可能不支持）
	cfg.Radar.DeviceMQTT.ProductID = getEnv("RADAR_MQTT_PRODUCT_ID", "88")
	cfg.Radar.DeviceMQTT.Timeout = parseInt(getEnv("RADAR_MQTT_TIMEOUT", "30"), 30)
	cfg.Radar.DeviceMQTT.Keepalive = parseInt(getEnv("RADAR_MQTT_KEEPALIVE", "60"), 60)
	cfg.Radar.DeviceMQTT.ClientIDPrefix = getEnv("RADAR_MQTT_CLIENT_ID_PREFIX", "radar")
	
	// 雷达服务配置
	cfg.Radar.Topics.Data = getEnv("RADAR_TOPIC_DATA", "radar/+/data")
	cfg.Radar.Topics.Command = getEnv("RADAR_TOPIC_COMMAND", "radar/+/command")
	cfg.Radar.Topics.OTA = getEnv("RADAR_TOPIC_OTA", "radar/+/ota")
	
	cfg.Log.Level = getEnv("LOG_LEVEL", "info")
	cfg.Log.Format = getEnv("LOG_FORMAT", "json")
	
	return cfg, nil
}

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

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

