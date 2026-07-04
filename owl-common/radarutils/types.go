// Package radarutils 雷达/画布坐标体系与几何工具（后端版）。
//
// ================ 坐标系约定 ================
//
// 画布坐标系（= 房间坐标系，前端 layout_config 所有对象共用此坐标）：
//   - 原点在画布顶部中心
//   - X 轴：左为负(-)，右为正(+)
//   - Y 轴：下为正(+)，上为负(-)
//   - 范围：x ∈ [-W/2, W/2]，y ∈ [0, H]
//   - 单位：cm（整数，四舍五入量化）
//
// 雷达本地坐标系（radar local，ceiling 模式下）：
//   - 原点在雷达中心 (0,0)
//   - H 轴：左为正(+H)，右为负(-H)  ==>  +H = -X
//   - V 轴：下为正(+V)，上为负(-V)  ==>  +V = +Y
//   - Z 轴：雷达安装高度方向（cm），转换时直接透传
//   - 单位：cm（整数）
//
// 注意："X 右为正"和"H 左为正"是一对镜像——所以雷达本地 → 画布转换里 x = -h。
//
// 与前端 owlFront/src/utils/radar/radarUtils.ts 完全对齐。
//
// ================ int 化约定 ================
//
// 所有持久化字段、公共 API 签名全部用 int。原因：
//   1. 系统 10cm 粒度量化，小数无物理意义
//   2. int 存储紧凑、序列化清晰、debug 可读
//   3. 整除自然形成"旧数据遗忘"效果（衰减时）
//
// 函数内部临时浮点（三角计算/矩阵运算）允许保留，但接口边界一律 int，
// 用 math.Round 四舍五入取整。
package radarutils

// InstallModel 雷达安装模式（与 owl-common/radar.InstallModel* 一致）
type InstallModel uint8

const (
	InstallCeiling InstallModel = 0 // 吸顶：矩形 FOV，4 角边界
	InstallWall    InstallModel = 1 // 贴墙：楔形 FOV，rearV=0
	InstallCorn    InstallModel = 2 // 墙角：菱形 FOV，内旋 -45°
)

// Point 画布坐标（= 房间坐标）中的一个点，单位 cm。
type Point struct {
	X, Y, Z int
}

// RadarPoint 雷达本地坐标中的一个点，单位 cm。
type RadarPoint struct {
	H, V, Z int
}

// Boundary 雷达 FOV 的四向距离（单位 cm，均为非负值）。
// 与设备协议 boundary_left/right/front/rear 一致。仅作 UI 参考，判定依据是物理 FOV。
type Boundary struct {
	LeftH  int // 雷达左侧探测距离（+H 方向）
	RightH int // 雷达右侧探测距离（-H 方向）
	FrontV int // 雷达前方探测距离（+V 方向）
	RearV  int // 雷达后方探测距离（-V 方向）；Wall 模式恒为 0
}

// BoundaryTruncationMargin 下发截断框冗余量（cm）：以 ModeBaseBoundary 为 base，四边各外扩此值 → 下发截断框。
// 单源镜像 FE types.ts BOUNDARY_TRUNCATION_MARGIN；改一处必改另一处。
const BoundaryTruncationMargin = 50

// ModeBaseBoundary 各安装模式最大探测边界（cm），镜像 FE RADAR_DEFAULT_CONFIG[mode].boundary
// （**不是** RADAR_BOUNDARY_LIMITS_BY_MODE：corn=500 非 550/600）。作下发截断框 base：
// base+BoundaryTruncationMargin → 截断框，对象与之相交即下发（跨界顶点 clamp 进框），完全在框外才丢。
func ModeBaseBoundary(m InstallModel) Boundary {
	switch m {
	case InstallWall:
		return Boundary{LeftH: 300, RightH: 300, FrontV: 400, RearV: 0}
	case InstallCorn:
		return Boundary{LeftH: 500, RightH: 500, FrontV: 0, RearV: 0}
	default: // ceiling
		return Boundary{LeftH: 300, RightH: 300, FrontV: 200, RearV: 200}
	}
}

// RadarMount 雷达安装配置：位置 + 旋转 + 安装方式 + 物理/安装 FOV + 用户边界。
// 足以把雷达本地坐标转换到画布（房间）坐标，并计算信号可达域、盲区、Z 上下限。
//
// 命名层次（与前端保持一致，同时补齐安装后视角）：
//
//	画布层（前端 layout 可编辑）
//	  Rotation     — 雷达在画布中的旋转角（度）。雷达对此无感。
//	                 对应 BaseObject.angle / radar.rotation（两者同步）。
//
//	物理层（雷达硬件规格，设备常量）
//	  HFOV         — 天线水平视场角 h_fov（度）。
//	  VFOV         — 天线垂直视场角 v_fov（度）。
//	                 对应前端 radar.hfov / vfov。
//
//	安装层（物理 FOV 经 installModel 姿态变换后的世界视角）
//	  AzimuthFOV   — 安装后方位角 a_fov（度）。
//	  ElevationFOV — 安装后俯仰角 e_fov（度）。
//	                 吸顶装：A≈HFOV，E 为向下锥角；
//	                 贴墙装：A=HFOV，E=VFOV；
//	                 墙角装：A 因遮挡可能缩减，E 偏向下方。
//	                 这两个字段是计算信号可达域、盲区、Z 上下限的主要依据。
//
//	用户层（layout 编辑设置的有效探测区）
//	  Boundary     — 用户设置的 XY 有效探测边界（cm）。**仅作 UI 参考**，判定用物理 FOV。
//
// 任一字段为 0 表示"未知"，下游计算应显式判断并回退到保守估计。
type RadarMount struct {
	Center       Point        // 雷达中心画布坐标（Z = installHeight，cm）
	Rotation     int          // 画布旋转角（度，0-359，逆时针为正）
	InstallModel InstallModel // 安装方式 Ceiling / Wall / Corn
	Boundary     Boundary     // 用户设置的 XY 有效边界（UI 参考）

	HFOV         int // 物理水平 FOV h_fov（度）
	VFOV         int // 物理垂直 FOV v_fov（度）
	AzimuthFOV   int // 安装后方位角 FOV a_fov（度）
	ElevationFOV int // 安装后俯仰角 FOV e_fov（度）
}

// internalRotationDeg 内部补偿角。
// 仅 corn 安装模式为 -45°—— corn 在前端只是为了方便安装工人用 left/right 命名，
// 内部转换时需要把这个命名旋转折算回去。
func (m RadarMount) internalRotationDeg() int {
	if m.InstallModel == InstallCorn {
		return -45
	}
	return 0
}

// effectiveRotationDeg 实际用于坐标旋转的角度 = Rotation + internalRotation。
func (m RadarMount) effectiveRotationDeg() int {
	return m.Rotation + m.internalRotationDeg()
}

// Rect 画布坐标下的一个轴对齐矩形（cm）。
// 约定 X1 ≤ X2, Y1 ≤ Y2。
type Rect struct {
	X1, Y1, X2, Y2 int
}

// Norm 规范化矩形，确保 X1≤X2, Y1≤Y2。
func (r Rect) Norm() Rect {
	if r.X1 > r.X2 {
		r.X1, r.X2 = r.X2, r.X1
	}
	if r.Y1 > r.Y2 {
		r.Y1, r.Y2 = r.Y2, r.Y1
	}
	return r
}

// 房间最大尺寸与网格常量（cm）。超过该尺寸的 layout 应被拒绝。
const (
	MaxRoomWidth  = 600
	MaxRoomHeight = 600
	CellSize      = 10 // 默认网格边长 cm
)
