// Package radar 定义雷达设备配置的规范格式（API GET/PUT 统一使用）。
// 仅保留规范性 key；厂家协议字段名（如 radar_install_style、rectangle、ssid_password 等）仅在网关模块内部使用，不在此定义。
package radar

// 规范格式 key 名（API 与前端统一使用 snake_case，单源真相）。boundary_* 仅定义/读取用，传递用 rectangle。
const (
	KeyInstallModel   = "install_model"
	KeyInstallHeight  = "install_height"
	KeyBoundaryLeft   = "boundary_left"
	KeyBoundaryRight  = "boundary_right"
	KeyBoundaryFront  = "boundary_front"
	KeyBoundaryRear   = "boundary_rear"
	KeyRectangle      = "rectangle"
	KeyDeclareArea    = "declare_area"
	KeyWorkModel      = "work_model"
	KeyAccelera       = "accelera"
	KeyWifiRssi       = "wifi_rssi"
	KeyIPPort         = "ip_port"
	KeyWifiSsid       = "wifi_ssid"
	KeyWifiPassword   = "wifi_password"
	KeyRunHorizontal = "run_Horizontal"
	KeyVoiceEndTip   = "voice_end_tip"
)



// InstallModel 安装方式（项目内统一用数字，不下发字符串）
// 与协议 radar_install_style 一致：0=顶装(吸顶) 1=侧装(贴墙) 2=墙角
const (
	InstallModelCeiling = 0 // 顶装/吸顶
	InstallModelWall    = 1 // 侧装/贴墙
	InstallModelCorn    = 2 // 墙角
)

// WorkModel 工作模式（项目内统一用数字，与协议 radar_func_ctrl 一致）
const (
	WorkModelTrajectory = 3  // 人数轨迹
	WorkModelFall       = 7  // 跌倒检测
	WorkModelSleep      = 11 // 呼吸睡眠
	WorkModelFull       = 15 // 全功能（床位监护）
)
