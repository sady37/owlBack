// Package consts 硬件/设备相关常量：雷达 key、安装/工作模式、pose、床浴室状态、设备状态等
// 与 radar_convert_table、Radar_MQTT_v3.0 协议一致，统一所有硬件常量
package consts

// --- Key 常量 ---
// 规范格式 key 名（API 与前端统一使用 snake_case）
const (
	KeyInstallModel  = "install_model"
	KeyInstallHeight = "install_height"
	KeyBoundaryLeft  = "boundary_left"
	KeyBoundaryRight = "boundary_right"
	KeyBoundaryFront = "boundary_front"
	KeyBoundaryRear  = "boundary_rear"
	KeyRectangle     = "rectangle"
	KeyDeclareArea   = "declare_area"
	KeyWorkModel     = "work_model"
	KeyAccelera      = "accelera"
	KeyWifiRssi      = "wifi_rssi"
	KeyIPPort        = "ip_port"
	KeyWifiSsid      = "wifi_ssid"
	KeyWifiPassword  = "wifi_password"
	KeyRunHorizontal = "run_Horizontal"
	KeyVoiceEndTip   = "voice_end_tip"
)

// --- 安装/工作模式数值 ---
// InstallModel 安装方式（0=顶装 1=侧装 2=墙角，与 radar_install_style 一致）
const (
	InstallModelCeiling = 0
	InstallModelWall    = 1
	InstallModelCorn    = 2
)

// WorkModel 工作模式（与 radar_func_ctrl 一致）
const (
	WorkModelTrajectory = 3  // 人数轨迹
	WorkModelFall       = 7  // 跌倒检测
	WorkModelSleep      = 11 // 呼吸睡眠
	WorkModelFull       = 15 // 全功能（床位监护）
)

// BedRoomState 床/浴室状态（与协议一致）
type BedRoomState int

const (
	StateInBed         BedRoomState = 0 // 上床
	StateLeftBed       BedRoomState = 1 // 离床
	StateEnterBathRoom BedRoomState = 3 // 进入浴室
	StateOutBathRoom   BedRoomState = 4 // 离开浴室
)

// 兼容旧常量名
const (
	InBed         = StateInBed
	LeftBed       = StateLeftBed
	EnterBathRoom = StateEnterBathRoom
	OutBathRoom   = StateOutBathRoom
)

// --- 类型 ---
// CanonicalConfig 雷达配置规范格式（API GET 返回、PUT 入参）
type CanonicalConfig struct {
	InstallModel  interface{} `json:"install_model,omitempty"`
	InstallHeight *int        `json:"install_height,omitempty"`
	BoundaryLeft  *int        `json:"boundary_left,omitempty"`
	BoundaryRight *int        `json:"boundary_right,omitempty"`
	BoundaryFront *int        `json:"boundary_front,omitempty"`
	BoundaryRear  *int        `json:"boundary_rear,omitempty"`
	Rectangle     string      `json:"rectangle,omitempty"`
	DeclareArea   interface{} `json:"declare_area,omitempty"`
	WorkModel     string      `json:"work_model,omitempty"`
	Accelera      string      `json:"accelera,omitempty"`
	WifiRssi      string      `json:"wifi_rssi,omitempty"`
	IPPort        string      `json:"ip_port,omitempty"`
	WifiSsid      string      `json:"wifi_ssid,omitempty"`
	WifiPassword  string      `json:"wifi_password,omitempty"`
	RunHorizontal string      `json:"run_Horizontal,omitempty"`
	VoiceEndTip   string      `json:"voice_end_tip,omitempty"`
}

// DeclareAreaItem declare_area 单条（id, type, x1..y4 单位 cm）
type DeclareAreaItem struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
	X1   int    `json:"x1"`
	Y1   int    `json:"y1"`
	X2   int    `json:"x2"`
	Y2   int    `json:"y2"`
	X3   int    `json:"x3"`
	Y3   int    `json:"y3"`
	X4   int    `json:"x4"`
	Y4   int    `json:"y4"`
}

// --- Pose 姿态数值（与协议 track 字节 13 一致）---
const (
	PoseInitialization         = 0
	PoseWalking                = 1
	PoseSuspectedFall          = 2
	PoseSitting                = 3
	PoseStanding               = 4
	PoseFall                   = 5
	PoseLying                  = 6
	PoseSuspectedSittingGround = 7
	PoseSittingGround          = 8
	PoseBedSitUp               = 9
	PoseSuspectedBedSitUp      = 10
	PoseBedSitUpConfirm        = 11 // 11 同 9 均为 BedSitUp
)

// DeviceStatus 设备状态（协议单次上报值，与 Radar_MQTT 一致）
// 协议每次上报一个值；多状态用 DeviceStatusSlice 数组
type DeviceStatus int

// DeviceStatusSlice 多状态数组，如 []DeviceStatus{StatusOnline, StatusSignalPoor}
type DeviceStatusSlice []DeviceStatus

// Contains 判断是否包含某状态
func (s DeviceStatusSlice) Contains(d DeviceStatus) bool {
	for _, v := range s {
		if v == d {
			return true
		}
	}
	return false
}

const (
	StatusOffline        DeviceStatus = 0  // 离线
	StatusOnline         DeviceStatus = 1  // 在线
	StatusAngleAbnormal  DeviceStatus = 3  // 倾角异常
	StatusAngleRecovery  DeviceStatus = 4  // 倾角恢复
	StatusSignalPoor     DeviceStatus = 5  // 信号差
	StatusSignalRecovery DeviceStatus = 6  // 信号恢复
	StatusSensorDetached DeviceStatus = 7  // 传感器脱落
	StatusSensorRetached DeviceStatus = 8  // 传感器重连
	StatusDeviceFailure  DeviceStatus = 14 // 设备故障
	StatusDeviceNormal   DeviceStatus = 15 // 设备正常
)

// PoseNumToDisplay track pose 数值 (0-11) → display_en 字符串
func PoseNumToDisplay(n int) string {
	switch n {
	case PoseInitialization:
		return "Initialization"
	case PoseWalking:
		return "Walking"
	case PoseSuspectedFall:
		return "SuspectedFall"
	case PoseSitting:
		return "Sitting"
	case PoseStanding:
		return "Standing"
	case PoseFall:
		return "Fall"
	case PoseLying:
		return "Lying"
	case PoseSuspectedSittingGround:
		return "SuspectedSittingOnGround"
	case PoseSittingGround:
		return "SittingOnGround"
	case PoseBedSitUp, PoseBedSitUpConfirm:
		return "BedSitUp"
	case PoseSuspectedBedSitUp:
		return "SuspectedBedSitUp"
	default:
		return ""
	}
}

// PoseDisplayToNum display_en 字符串 → pose 数值 (0-11)，用于兼容 string 来源
func PoseDisplayToNum(s string) (int, bool) {
	switch s {
	case "Initialization":
		return 0, true
	case "Walking":
		return 1, true
	case "SuspectedFall":
		return 2, true
	case "Sitting":
		return 3, true
	case "Standing":
		return 4, true
	case "Fall":
		return 5, true
	case "Lying":
		return 6, true
	case "SuspectedSittingOnGround":
		return 7, true
	case "SittingOnGround":
		return 8, true
	case "BedSitUp":
		return 9, true
	case "SuspectedBedSitUp":
		return 10, true
	default:
		return 0, false
	}
}

// --- 安装/工作模式 display（数值与协议一致）---
const (
	InstallModelDisplayCeiling = 0 // ceiling
	InstallModelDisplayWall    = 1 // wall
	InstallModelDisplayCorn    = 2 // corn
)

// InstallModelToDisplay 安装模式数值 → display 字符串
func InstallModelToDisplay(model int) string {
	switch model {
	case InstallModelDisplayCeiling:
		return "ceiling"
	case InstallModelDisplayWall:
		return "wall"
	case InstallModelDisplayCorn:
		return "corn"
	default:
		return ""
	}
}

const (
	WorkModelDisplayTrajectory = 3  // 人数轨迹
	WorkModelDisplayFall       = 7  // 跌倒检测
	WorkModelDisplaySleep      = 11 // 呼吸睡眠
	WorkModelDisplayFull       = 15 // 全功能
)

// WorkModelToDisplay 工作模式数值 → display 字符串
func WorkModelToDisplay(model int) string {
	switch model {
	case WorkModelDisplayTrajectory:
		return "trajectory"
	case WorkModelDisplayFall:
		return "fall"
	case WorkModelDisplaySleep:
		return "sleep"
	case WorkModelDisplayFull:
		return "full"
	default:
		return ""
	}
}
