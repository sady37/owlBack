package models

// AuthResponse 雷达设备认证响应
// 参考 radar-ql-v3/radar-v3-json-massages.md 和协议文档
type AuthResponse struct {
	Msg  string      `json:"msg"`  // 操作结果描述
	Code int         `json:"code"` // 状态码（200=成功）
	Data *AuthData   `json:"data"` // 业务数据
}

// AuthData 认证数据
type AuthData struct {
	UID  string    `json:"uid"`  // 设备序列号
	MQTT *MQTTConfig `json:"mqtt"` // MQTT 连接配置
}

// MQTTConfig MQTT 连接配置
type MQTTConfig struct {
	ClientID  string `json:"clientId"`  // MQTT 客户端标识
	Timeout   int    `json:"timeout"`   // 连接超时（秒）
	Keepalive int    `json:"keepalive"` // 心跳间隔（秒）
	Server    string `json:"server"`    // MQTT 服务器地址
	Port      int    `json:"port"`      // MQTT 服务器端口
	Account   string `json:"account"`   // MQTT 账号
	Pwd       string `json:"pwd"`       // MQTT 密码
	Protocol  string `json:"protocol"`  // 协议类型（"1"=不加密, "2"=加密）
	Prefix    string `json:"prefix"`    // 主题前缀（可为空）
	ProductID string `json:"productId"` // 产品 ID (0-255)
}

