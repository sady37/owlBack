package models

// AuthRequest 雷达设备认证请求
// 参考 radar-ql-v3/simple-https.py 和协议文档
type AuthRequest struct {
	UID  string `json:"uid"`  // 设备序列号
	Type int    `json:"type"` // 设备类型
	Auth string `json:"auth"` // 认证随机数，服务器可忽略
	Salt string `json:"salt"` // 认证随机数，服务器可忽略
	
	// 注意：协议文档使用 MCU 和 Radar（大写），但测试程序使用 mcu 和 radar（小写）
	// 当前实现支持小写格式（与测试程序一致），如需支持大写，可添加 UnmarshalJSON 方法
	MCU   MCUInfo   `json:"mcu"`   // 主控信息（兼容 MCU）
	Radar RadarInfo `json:"radar"` // 雷达信息（兼容 Radar）
}

// MCUInfo 主控信息
type MCUInfo struct {
	HW    string `json:"hw"`    // 主控硬件类型（4G: 1.0, 1.1; Wi-Fi: HC2-2.0, TK2-2.3）
	SW    string `json:"sw"`    // 主控固件编译日期
	MAC   string `json:"mac"`   // Wi-Fi 设备属性
	ICCID string `json:"iccid"` // 4G 设备属性
}

// RadarInfo 雷达信息
type RadarInfo struct {
	HW  string `json:"hw"`  // 雷达硬件类型（T型号: 1.0; S型号: 2.3）
	SW  string `json:"sw"`  // 雷达固件编译日期
	Cap string `json:"cap"` // 雷达功能类别
}

