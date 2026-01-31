package radar

// CanonicalConfig 雷达配置规范格式（API GET 返回、PUT 入参）。
// 单位：cm；json key 与 keys.go 常量一致。Boundary* 仅定义/读取用，传递用 Rectangle。
type CanonicalConfig struct {
	InstallModel   interface{} `json:"install_model,omitempty"`   // 0/1/2，见 radar.InstallModelCeiling/Wall/Corn
	InstallHeight  *int        `json:"install_height,omitempty"` // 安装高度 cm，与 KeyInstallHeight 一致
	BoundaryLeft   *int        `json:"boundary_left,omitempty"`   // 仅定义/读取，传递用 Rectangle
	BoundaryRight  *int        `json:"boundary_right,omitempty"`
	BoundaryFront  *int        `json:"boundary_front,omitempty"`
	BoundaryRear   *int        `json:"boundary_rear,omitempty"`
	Rectangle      string      `json:"rectangle,omitempty"` // 边界传递：{x1,y1;x2,y2;x3,y3;x4,y4}（cm）
	DeclareArea    interface{} `json:"declare_area,omitempty"`   // 数组或 JSON 字符串
	WorkModel      string      `json:"work_model,omitempty"`
	Accelera       string      `json:"accelera,omitempty"`
	WifiRssi       string      `json:"wifi_rssi,omitempty"`
	IPPort         string      `json:"ip_port,omitempty"`
	WifiSsid       string      `json:"wifi_ssid,omitempty"`
	WifiPassword   string      `json:"wifi_password,omitempty"`
	RunHorizontal  string      `json:"run_Horizontal,omitempty"`
	VoiceEndTip    string      `json:"voice_end_tip,omitempty"`
}

// DeclareAreaItem 规范格式下 declare_area 单条（id, type, x1..y4 单位 cm）
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
