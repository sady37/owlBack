package roomengine

import (
	"context"
	"encoding/json"
	"net/netip"
	"strconv"
	"strings"
	"sync"
	"time"

	"owl-common/alarm"
	"owl-common/card"
	"owl-common/radarutils"
	rediscommon "owl-common/redis"
	"owl-common/spatial"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// ========================================================================
// RoomConfig：从 layout_config JSONB 解析得到的房间配置
// ========================================================================

// RoomConfig 房间配置（全 int 化 + 对齐 radarutils 类型）
// AreaZone 一个区域对象的空间语义（空间轴单源）：Rect + AreaType + Conf，engine 统一 SetPrior 刷 cell。
// AreaType 跟对象走（= layout typeValue），不按 ObjectType 分桶（rules.md #S2 正交）。
type AreaZone struct {
	Rect     radarutils.Rect
	AreaType AreaType
	Conf     int
}

type RoomConfig struct {
	RoomID   string
	RoomName string // rooms.room_name；保留兼容；sensor_v2 决定 16 后只用 RoomType 做分支判定，不再用 RoomName 字符串匹配
	RoomType int    // owl-common/card.RoomType*: 0=Default / 1=Bathroom / 2=Kitchen
	RoomW    int    // 画布宽（cm）
	RoomH    int    // 画布深（cm）
	OriginX  int    // grid[0][0] 左上角的画布坐标 X（让 grid 覆盖物体 bbox）
	OriginY  int    // grid[0][0] 左上角的画布坐标 Y

	// Wall 围出的房间多边形（闭合），用于 StampRoomPolygon
	WallPolygon []radarutils.Point

	// 空间轴单源（rules.md #S2）：每个区域对象一条 (Rect+AreaType+Conf)，engine 统一 SetPrior 刷 cell。
	AreaZones []AreaZone

	// 物理轴几何：有 AreaType 之外的物理行为（门口/covers+床闸/tm），故单独留几何；刷 cell 走 AreaZones。
	Enters       []radarutils.Rect // 门口：StampEnters / door-pin
	Beds         []radarutils.Rect // 床几何：covers/MM
	BedAreaTypes []AreaType        // 与 Beds 1:1，仅供 countRealBeds 床闸
	Chairs       []radarutils.Rect // tm.SetChairs
	ChairIDs     []string          // 与 Chairs 1:1：canvas 对象 id（per-chair 久坐学习态锚定 key，跨 layout 编辑存活）
	Interferes   []radarutils.Rect // tm.SetInterferes / mirror_detect

	// BedAreaIDs 床区 area_id 集合（areaType∈{2床,5监护床}），来源=固件活体 declare_area
	// （bootstrap 走 wisefido-data HTTP 覆盖，治 canvas 下发区域 vs 固件几何漂移）。
	// radar track 帧的 area_id 命中此集合 → 在床（TrackManager.fwIsBed membership）→ N=1 驱动 SBed boost。
	BedAreaIDs []int

	// 物体顶部高度（cm，地面为 0），与上面同名切片一一对应（Heights[i] ↔ Beds[i]）。
	// 来源：layout JSON 里每个对象的 height 字段，由前端 Toolbar 录入；
	// 缺失时 ParseLayoutConfig 用 FurnitureType 默认值（与前端 FURNITURE_CONFIGS.defaultHeight 对齐）。
	// 当前 RoomEngine 不使用，仅持久化保留——未来用于：
	//   - 空中物体（吊灯 height>200）不阻挡通行 → InFOV 不刷 Deny
	//   - 床/桌面高度参与 Z 轴范围判定 → fall 检测加先验
	EnterHeights     []int
	BedHeights       []int
	ChairHeights     []int
	InterfereHeights []int

	// sensor_v2 决定 15: EnterTargets 与 Enters 同长度平行 slice。
	// 取值：
	//   ""         → inside_enter (默认；不强制标识)
	//   "outside"  → 通向 unit 外（必须人工标）
	//   "bathroom" → 通向 bathroom（必须人工标；v2 单 bathroom 无需 id）
	// vue 编辑器 dropdown 三选项，写入 layout JSON 每个 enter 对象的 enter_target 字段。
	// 缺失或为 "" → 视作 inside_enter。
	EnterTargets []string

	// 雷达安装
	// Radar = 主雷达（单雷达房 = 唯一一台；多雷达房 = 第一台，供 boundaryPolygon 兜底 +
	// L1 mirror ghost tiebreaker 用）。Radars = 该房全部雷达 mount（per-device 坐标转换），
	// RadarAddrs 平行存各 mount 的 deviceAddr。单雷达房 len(Radars)==1；空 = placeholder。
	Radar      radarutils.RadarMount
	Radars     []radarutils.RadarMount
	RadarAddrs []string

	// DeviceBeds per-device(/128) 各自 canvas 画的床矩形 + 高度。多雷达房不合并：covers 计算
	// 必须用**该雷达自己 layout** 的床（D523 自己没画床 → covers=0），不能借别台 canvas 的床
	// （device-local 帧未配准，跨系判定无意义）。key = deviceAddr。
	DeviceBeds       map[string][]radarutils.Rect
	DeviceBedHeights map[string][]int

	// Sleepad 位置（point 几何，可能多个）— 用于事件路由 + 可视化
	Sleepads []radarutils.Point

	// Timezone IANA 时区字符串（如 "America/Denver"），来自该 room 所属 unit 的 units.timezone。
	// 用于 IsNightTime 判定 risk-time（默认 23:30-07:30 本地时间）。
	// 空值 → engine 退化为 UTC（错位风险，必须设置）。
	Timezone string

	// sensor_v2 PR-3 wiring：
	// SuiteID    = census bucket key；取值规则按 IsPublicBathroom 分两路（决定 24 / PR-7）：
	//                public bathroom (IsPublicBathroom=true)  → bathroom_spatial_prefix /128（自身）
	//                suite bathroom / bedroom / 其他           → unit_spatial_prefix /80
	//              bootstrap 责任：按下面规则计算 SuiteID 传入 RegisterRoom。
	// ResidentID = bedroom 关联的 resident.id；bathroom 不必填；PR-2 SuiteCensusManager 升格用。
	// 空值 = bootstrap 未配置 → TrackStatus.PersonID/PersonRole 一律空。
	//
	// **当前是单 resident 字段，multi-resident bedroom 暂未支持**：
	//   PR-3 范围内 caller 责任传入"该 room 的预期 resident"（单 resident unit 直接传 residents[0]）。
	//   决定 19 multi-resident 场景（dual sleepad bedroom）的正解 = sleepad device_uid → resident.id
	//   静态映射，未来扩展为 []ResidentBinding 或 ResidentLookup func；
	//   PR-5/6 BathroomGate + 实测 multi-resident fixture 时再决（详 sensor_v2.md 决定 19）。
	SuiteID    string
	ResidentID string

	// sensor_v2 决定 24 / PR-7 Public Bathroom Standalone Mode：
	// 仅当 RoomType==card.RoomTypeBathroom 时有意义：
	//   true  = public bathroom（楼层共用，多人不识别身份）→ RegisterRoom 触发 census.MarkPublicBathroom，
	//           BathroomGate.Process / BathroomGhostAdjudicator 规则 2/3 因 IsPublicBathroom 自动降级
	//   false = suite bathroom（附属某 bedroom，elder 私用）→ 走完整 §4.A 判定 + BathroomGate 入口流量
	// 来源：rooms.is_public_bathroom 列（migration 2026-05-17_rooms_add_is_public_bathroom.sql）
	// 显式配置不推断（详 sensor_v2.md §4.A.6 触发条件 + 拒绝拓扑推断的理由）
	IsPublicBathroom bool
}

// DefaultParamSets 三组并行参数（保守/中庸/激进），cell.UpdateBelief 似然先验用。
var DefaultParamSets = [3]ParamSet{
	{Alpha: 0.01, Beta: 0.2, FlipTh: 10, Name: "conservative"},
	{Alpha: 0.02, Beta: 0.5, FlipTh: 20, Name: "balanced"},
	{Alpha: 0.05, Beta: 1.0, FlipTh: 30, Name: "aggressive"},
}

// ========================================================================
// Engine
// ========================================================================

type Engine struct {
	mu           sync.RWMutex
	rooms        map[string]*TrackManager         // roomID → TrackManager
	radarPeople  map[string]int                   // roomID → census 折叠后 radar 真人数（belief 层经 SetRoomRadarPeople 注入）；RealPeopleInRoom 唯一权威，未注入返回 -1
	grids        map[string]*RoomGrid             // roomID → Grid
	deviceMounts map[string]radarutils.RadarMount // deviceAddr(/128) → 该 radar 安装参数（坐标转换用）；多雷达房每台一份
	deviceRoom   map[string]string                // deviceAddr → roomID (sensor 唯一物理寻址)
	deviceBeds   map[string][]radarutils.Rect     // deviceAddr(/128) → 该设备**自己 canvas** 的 bed 矩形（MM covers 只用本台 layout，不跨系借别台床）
	deviceBedHs  map[string][]int                 // deviceAddr → 各 bed 高度（与 deviceBeds 同序）
	// PR-8: AI publish 用 — 源 radar UUID 反查 deviceUID + room 反查 tenant/card
	deviceAddrToUID  map[string]string // deviceAddr(UUID) → device_uid（IoTStreamMessage.DeviceUID）
	deviceAddrToType map[string]string // deviceAddr(UUID) → 源 sensor 类型（"Radar"/"Sleepad"），AI publish 加 ".AI<node>" 后缀
	roomTenants      map[string]string // roomID → tenant_id（alarm_events 必填）

	// layout 几何 hash，RegisterRoom 时计算；snapshot save/load 比对用
	layoutHashes map[string]string

	// cell.UpdateBelief 三组似然先验参数（hydrate 后冻结，DBN 读真实 AreaType）
	paramSets [3]ParamSet

	// 运行时参数（由 Configure 注入；Decay/Learn 字段语义见 cell.go / cell_learning.go）
	decayParams    DecayParams
	learnParams    LearnParams
	bedsideFallCfg BedsideFallConfig // R4 床边晕倒参数；RegisterRoom 时下发到 TrackManager
	staticScanMs   int               // static reflector 扫描节流间隔(ms)；RegisterRoom 时下发到 TrackManager

	decayInterval time.Duration

	// 路由失败的 device 频率限制告警（避免每条 frame 都 warn 一次）。
	// 同一 device key 60s 内只 warn 一次，确保 deviceRoom 缺失能被发现而不淹日志。
	unroutedMu sync.Mutex
	unrouted   map[string]int64 // device key → last warn epoch ms

	redisClient *redis.Client
	logger      *zap.Logger

	onOutput func(roomID string, outputs []TrackOutput)

	// sensor v2 PR-3 wiring：
	//   - suiteCensus 进程级共享 census manager（nil = PR-3 未启用，TrackStatus.PersonID 空）
	//   - roomSuiteID    roomID → SuiteID（unit spatial_prefix INET 字符串）；bootstrap 注入
	//   - roomResidentID roomID → resident.id（PR-2 升格用；bathroom room 不必填）
	//   - roomType       roomID → card.RoomType* (0=Default / 1=Bathroom / 2=Kitchen)
	// 单 reader（handleMessage 调 SnapshotTrackStatuses 后 publish）；上述 map 在 RegisterRoom
	// + SetSuiteCensus 时写入，运行时只读，无并发风险。
	suiteCensus    *SuiteCensusManager
	roomSuiteID    map[string]string
	roomResidentID map[string]string
	roomType       map[string]int
	// smallBathroom P6.1b-D(审查㉛ Opt-1):roomID→是否小卫生间(bbox 最小边≤200 ∧ RoomType==Bathroom)。
	// 小卫生间门距退化(处处近门,审查⑳)→ 走 D 延迟确认/分级,不走常规 reachableExit 抑制。RegisterRoom 算。
	smallBathroom map[string]bool

	// trackLastSeen 是 publishTrackStatuses 的失锁判定状态：roomID → trackID → 上次出现的 nowMs。
	// 用途：firmware track_id 复用场景下，SuiteCensus.AnchorTrackID 必须在 track 真正失锁后清空，
	// 否则新 track 复用旧 trackID 会被当成"同一人活动"持续 update LastActiveMs。
	// 判据：当前 snapshot 未含 trackID 且 (nowMs - lastSeenMs) ≥ 60s → 调 ClearAnchorOnLostTrack。
	// 60s 偏保守（远大于 MaxMissCount=10 帧的 coast 期），避免 firmware coast 期间误清。
	// 写者：publishTrackStatuses（per-room 串行，与 handleMessage 同 goroutine）；无锁需要。
	trackLastSeen map[string]map[int]int64

	// sensor_v2 PR-5 BathroomGate 入口流量子模块（§4.A.2）：
	//   每 bathroom room 一个 gate（key = roomID），lazy 创建在 publishTrackStatuses。
	//   census + suiteID 为空时不创建（fallback noop）。
	//   单 goroutine 读写（publishTrackStatuses per-room 串行），无锁需要。
	bathroomGates map[string]*BathroomGate

	// OnRoomFrame seam：每帧 ProcessFrame + SnapshotTrackStatuses 之后触发，交上层 DBN（engine.Room）裁决。
	// nil = 不裁决（纯馈送，无下游）。Engine 不 import belief/adapter/engine 包，靠回调解耦。
	// 返回 (fired, dropped) 的 LogicID：fired → 复位 still-box（belief 已就地复位）；
	//   dropped（确认离场/空）→ evict track_manager，停 12s coast re-feed（防 census 重发新 logicID = churn）。
	OnRoomFrame func(roomID string, bases []TrackStatusBase, bed card.BedState, nowMs int64, exitLogOdds, ghostLeftLogOdds func(logicID string, atMs int64) float64, hardExited func(logicID string, atMs int64) bool) (fired, dropped []string)
}

// RuntimeConfig 与 owlBack/tools/Xsensorv1/internal/config::RoomEngineConfig 一一对应；
// engine 包不依赖 config 包，wiring 在 cmd/wisefido-sensor/main.go 中完成转换。
type RuntimeConfig struct {
	Decay     DecayParams
	Learn     LearnParams
	ParamSets [3]ParamSet

	// 风险时段（夜间）；通过 SetRiskTimeConfig 注入到包级 nightCfg。
	// 全 0 视为未设置，保留 IsNightTime 默认（23:30 - 07:30）。
	RiskTime RiskTimeConfig

	// R4 床边晕倒参数；通过 TrackManager.SetBedsideFallConfig 注入到每个房间。
	// 任一字段 0 保留 defaultBedsideFallCfg 默认值。
	BedsideFall BedsideFallConfig

	// static reflector 扫描节流间隔（ms）；通过 TrackManager.SetStaticScanIntervalMs 注入。
	// 0 保留 NewTrackManager 默认 5000ms。
	StaticScanIntervalMs int
}

// NewEngine 创建 Room Engine（默认参数）。生产环境调用 Configure 注入 yaml 配置。
func NewEngine(redisClient *redis.Client, logger *zap.Logger) *Engine {
	return &Engine{
		rooms:            make(map[string]*TrackManager),
		radarPeople:      make(map[string]int),
		grids:            make(map[string]*RoomGrid),
		deviceMounts:     make(map[string]radarutils.RadarMount),
		deviceRoom:       make(map[string]string),
		deviceBeds:       make(map[string][]radarutils.Rect),
		deviceBedHs:      make(map[string][]int),
		deviceAddrToUID:  make(map[string]string),
		deviceAddrToType: make(map[string]string),
		roomTenants:      make(map[string]string),
		layoutHashes:     make(map[string]string),
		paramSets:        DefaultParamSets,
		decayParams:      DefaultDecayParams(),
		learnParams:      DefaultLearnParams(),
		decayInterval:    1 * time.Hour,
		unrouted:         make(map[string]int64),
		redisClient:      redisClient,
		logger:           logger,
		roomSuiteID:      make(map[string]string),
		roomResidentID:   make(map[string]string),
		roomType:         make(map[string]int),
		smallBathroom:    make(map[string]bool),
		trackLastSeen:    make(map[string]map[int]int64),
		bathroomGates:    make(map[string]*BathroomGate),
	}
}

// warnUnrouted 路由失败的 device 频率限制告警（同一 key 60s 内一次）。
// stream：消息来自哪个 stream（"monitor"/"event"/"alarm"）便于排查。
func (e *Engine) warnUnrouted(stream, cardID, deviceAddr, deviceType string) {
	key := deviceAddr
	if key == "" {
		key = cardID
	}
	if key == "" {
		return
	}
	nowMs := time.Now().UnixMilli()
	e.unroutedMu.Lock()
	last := e.unrouted[key]
	if nowMs-last < 60_000 {
		e.unroutedMu.Unlock()
		return
	}
	e.unrouted[key] = nowMs
	e.unroutedMu.Unlock()
	e.logger.Warn("dropped_unrouted_message",
		zap.String("stream", stream),
		zap.String("device_addr", deviceAddr),
		zap.String("card_id", cardID),
		zap.String("device_type", deviceType),
		zap.String("hint", "device not in deviceRoom; config:card:stream subscriber should heal on next event"),
	)
}

// Configure 注入 yaml 加载的运行时参数。零值字段保留 New 时的默认值。
func (e *Engine) Configure(cfg RuntimeConfig) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if cfg.Decay.ImmediateSec > 0 {
		e.decayParams = cfg.Decay
	}
	if cfg.Learn.SitActiveX10 > 0 {
		e.learnParams = cfg.Learn
	}
	if cfg.ParamSets[0].Name != "" {
		e.paramSets = cfg.ParamSets
	}

	// 风险时段（IsNightTime 用）—— 包级 var，所有房间共享
	SetRiskTimeConfig(cfg.RiskTime)

	// R4 参数：保存到 engine 级 + 下发到所有已注册房间的 TrackManager。
	// （RegisterRoom 在 Configure 之前调用的房间也要补设。）
	e.bedsideFallCfg = cfg.BedsideFall
	e.staticScanMs = cfg.StaticScanIntervalMs
	for _, tm := range e.rooms {
		tm.SetBedsideFallConfig(cfg.BedsideFall)
		tm.SetStaticScanIntervalMs(cfg.StaticScanIntervalMs)
	}
	e.logger.Info("roomengine_configured",
		zap.Int("static_scan_interval_ms", cfg.StaticScanIntervalMs))
}

// SetOutputCallback 设置 track 输出回调（发 alarm 等下游）
func (e *Engine) SetOutputCallback(fn func(roomID string, outputs []TrackOutput)) {
	e.onOutput = fn
}

// RoomForDevice 查 deviceKey（device_addr 或 device_uid）对应的 room_id。
// 用于 alarm feedback 反查。空字符串 = 未路由（设备未绑定房间或未注册）。
func (e *Engine) RoomForDevice(deviceKey string) string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.deviceRoom[deviceKey]
}

// MountForDevice 取指定 device 的雷达安装参数（用于 RadarToCanvas 转换）。
// per-device：多雷达房每台 radar 各自坐标系；回退 roomID key 容老路径（RegisterRoom 无 RadarAddrs 时）。
// found=false 表示 device 未注册或未 stamp radar。
func (e *Engine) MountForDevice(deviceAddr string) (radarutils.RadarMount, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if m, ok := e.deviceMounts[deviceAddr]; ok {
		return m, true
	}
	m, ok := e.deviceMounts[e.deviceRoom[deviceAddr]]
	return m, ok
}

// RealPeopleInRoom 该房非 ghost track 数（去掉 owl 判出的 ghost）。zoneengine room total_people
// 用作 radar_np（替 firmware number_people——后者含 ghost）。房未注册 → -1（caller 退回 firmware）。
// roomID 用 netip 归一匹配，避免 zoneengine CIDR 文本 与 e.rooms key(DB room_id::text) 格式差异。
func (e *Engine) RealPeopleInRoom(roomID string) int {
	want, err := netip.ParsePrefix(roomID)
	if err != nil {
		return -1
	}
	e.mu.RLock()
	var tm *TrackManager
	matchedKey := roomID
	if t, ok := e.rooms[roomID]; ok {
		tm = t
	} else {
		for k, t := range e.rooms {
			if kp, perr := netip.ParsePrefix(k); perr == nil && kp == want {
				tm = t
				matchedKey = k
				break
			}
		}
	}
	rp, rpOK := e.radarPeople[matchedKey]
	e.mu.RUnlock()
	if tm == nil {
		return -1
	}
	// P1 占用人数：cutover 后唯一权威 = belief 层 census 折叠真人数（PReal 排 ghost + 同房跨雷达同人折叠，
	//   2雷达1人→1）。belief 未注入（未 Tick / 非 cutover；no-layout 房不 Tick）→ -1，caller 退回 firmware。
	if rpOK {
		return rp
	}
	return -1
}

// SetRoomRadarPeople belief 层每 tick 注入该房 census 折叠后的 radar 真人数（RealPeopleInRoom P1 用）。
func (e *Engine) SetRoomRadarPeople(roomID string, n int) {
	e.mu.Lock()
	e.radarPeople[roomID] = n
	e.mu.Unlock()
}

// SetTrackPReal belief 层每 tick 把某轨 census realness 后验按 LogicID 回灌该房 TrackManager
//（ghost 单源：cell 占用 / 床区学习读 TrackState.PReal）。roomID netip 归一查找同 RealPeopleInRoom。
func (e *Engine) SetTrackPReal(roomID, logicID string, pReal, pMirror float64, roomNp int) bool {
	want, err := netip.ParsePrefix(roomID)
	if err != nil {
		return false
	}
	e.mu.RLock()
	tm := e.rooms[roomID]
	if tm == nil {
		for k, t := range e.rooms {
			if kp, perr := netip.ParsePrefix(k); perr == nil && kp == want {
				tm = t
				break
			}
		}
	}
	e.mu.RUnlock()
	if tm == nil {
		return false
	}
	return tm.SetTrackPReal(logicID, pReal, pMirror, roomNp)
}

// GhostSignals lost_fall delete 判据信号读出（§10.3-b；forensic/xray），按 roomID+logicID。
func (e *Engine) GhostSignals(roomID, logicID string) (enterBorn, sustained bool, maxGhost float64, maxNp int, found bool) {
	want, err := netip.ParsePrefix(roomID)
	if err != nil {
		return false, false, 0, 0, false
	}
	e.mu.RLock()
	tm := e.rooms[roomID]
	if tm == nil {
		for k, t := range e.rooms {
			if kp, perr := netip.ParsePrefix(k); perr == nil && kp == want {
				tm = t
				break
			}
		}
	}
	e.mu.RUnlock()
	if tm == nil {
		return false, false, 0, 0, false
	}
	return tm.GhostSignals(logicID)
}

// SnapshotSleepads 按 roomID 取本房在场 sleepad 占用身份快照（forensic + 后续吸纳；只读身份不进裁决）。
// roomID netip 归一查找与 RealPeopleInRoom 同（独立内联，不改 A 的占用函数体，规则 #1.3 协调）。
func (e *Engine) SnapshotSleepads(roomID string, nowMs int64) []SleepadStatus {
	want, err := netip.ParsePrefix(roomID)
	if err != nil {
		return nil
	}
	e.mu.RLock()
	var tm *TrackManager
	if t, ok := e.rooms[roomID]; ok {
		tm = t
	} else {
		for k, t := range e.rooms {
			if kp, perr := netip.ParsePrefix(k); perr == nil && kp == want {
				tm = t
				break
			}
		}
	}
	e.mu.RUnlock()
	if tm == nil {
		return nil
	}
	return tm.SnapshotSleepads(nowMs)
}

// RadarBedReachCount 数本 radar 边界内有几张床——**只用该 radar 自己 canvas 画的床**(e.deviceBeds[addr])，
// 取每张床矩形中心 + 床高判 SignalReachable。covers 谓词用之：m=本数 / n=房内 bed 前缀数。
// 本台 layout 没画床(如 D523) → rects 空 → m=0 → covers=0，不会借别台(9e7)的床跨系误判。
func (e *Engine) RadarBedReachCount(deviceAddr string) int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	roomID := e.deviceRoom[deviceAddr]
	mount, ok := e.deviceMounts[deviceAddr]
	if !ok {
		mount, ok = e.deviceMounts[roomID]
	}
	if !ok {
		return 0
	}
	rects := e.deviceBeds[deviceAddr]
	hs := e.deviceBedHs[deviceAddr]
	boundary := radarutils.BoundaryVertices(mount)
	m := 0
	for i, r := range rects {
		r = r.Norm()
		cx, cy := (r.X1+r.X2)/2, (r.Y1+r.Y2)/2
		z := 0
		if i < len(hs) {
			z = hs[i]
		}
		reach := mount.SignalReachable(cx, cy, z)
		if reach {
			m++
		}
		e.logger.Info("covers_geom_trace",
			zap.String("radar", deviceAddr), zap.String("room", roomID),
			zap.Int("mount_cx", mount.Center.X), zap.Int("mount_cy", mount.Center.Y),
			zap.Int("mount_z", mount.Center.Z), zap.Int("rotation", mount.Rotation),
			zap.Int("install", int(mount.InstallModel)), zap.Int("boundary_verts", len(boundary)),
			zap.Any("boundary", boundary),
			zap.Int("bed_idx", i), zap.Int("bed_x1", r.X1), zap.Int("bed_y1", r.Y1),
			zap.Int("bed_x2", r.X2), zap.Int("bed_y2", r.Y2),
			zap.Int("bed_cx", cx), zap.Int("bed_cy", cy), zap.Bool("reachable", reach))
	}
	return m
}

// MarkFakeAlarmAt 在指定房间的 cell at (x, y) 累计 FakeAlarmCount + 1。
// 用于 alarm_events.operation='false_alarm' 反馈链：
// 人工标 fake → 反查 trigger 位置 → 命中 cell → 影响未来 still fall 阈值。
// 房间不存在 / cell 越界 → false（调用方可记日志或忽略）。
func (e *Engine) MarkFakeAlarmAt(roomID string, x, y int, nowMs int64) bool {
	return e.ApplyToCell(roomID, x, y, nowMs, func(c *Cell) { c.IncrFakeAlarm() })
}

// ApplyToCell 通用 cell 修改接口。返回是否成功（cell 在 grid 内）。
// 用于 PR-6 alarm-feedback 按 conditions 分流到不同 counter。
// 调用方传 closure 直接操作 cell；engine 负责 grid 路由 + LastUpdateMs 更新。
func (e *Engine) ApplyToCell(roomID string, x, y int, nowMs int64, fn func(*Cell)) bool {
	e.mu.RLock()
	g := e.grids[roomID]
	e.mu.RUnlock()
	if g == nil {
		return false
	}
	cell := g.CellAt(x, y)
	if cell == nil {
		return false
	}
	fn(cell)
	if nowMs > cell.LastUpdateMs {
		cell.LastUpdateMs = nowMs
	}
	return true
}

// VetoRect 按 rect（fire 点 40×40 足迹）对每个 cell 清非 Human 抑制/deny（→AreaUnknown）；sticky 时
// 额外 MarkLearnBlocked 永久封自动学习（跨重启）。两触发都由 data 驱动：(1) 删 layout source='Feedback'
// object（sticky=false 仅清）；(2) handle "Clear/Never auto-suppress"（Clear→sticky=false；Never→
// sticky=true 含清 + 永久封）。返回 (清/封 cell 数, ok=房路由命中且 grid 存在)。
func (e *Engine) VetoRect(deviceAddr string, rect radarutils.Rect, sticky bool) (cleared, blocked int, ok bool) {
	roomID := e.RoomForDevice(deviceAddr)
	if roomID == "" {
		return 0, 0, false
	}
	e.mu.RLock()
	g := e.grids[roomID]
	e.mu.RUnlock()
	if g == nil {
		return 0, 0, false
	}
	cleared, blocked = g.VetoRect(rect, sticky)
	return cleared, blocked, true
}

func (e *Engine) MapDeviceToRoom(deviceKey, roomID string) {
	e.mu.Lock()
	e.deviceRoom[deviceKey] = roomID
	e.mu.Unlock()
}

// MapDeviceAddrToUID 注册 device UUID → device_uid（PR-8 AI publish 反查源 radar UID 用）。
func (e *Engine) MapDeviceAddrToUID(deviceAddr, deviceUID string) {
	if deviceAddr == "" || deviceUID == "" {
		return
	}
	e.mu.Lock()
	e.deviceAddrToUID[deviceAddr] = deviceUID
	e.mu.Unlock()
}

// MapDeviceAddrToType 注册 device UUID → 源 sensor 类型（"Radar"/"Sleepad"）。
// AI publish 时直接作 message.DeviceType（不再拼 AI 后缀）。缺失时兜底 "Radar"。
func (e *Engine) MapDeviceAddrToType(deviceAddr, deviceType string) {
	if deviceAddr == "" || deviceType == "" {
		return
	}
	e.mu.Lock()
	e.deviceAddrToType[deviceAddr] = deviceType
	e.mu.Unlock()
}

// DeviceUIDHex 反查 device addr (IPv6 字符串) → device_uid hex MAC（人眼可读）。
// 用于 log 双格式（IPv6 字符串机器友好 / hex MAC 人眼对照设备贴纸）。
// 缺失返回空字符串（caller 用 zap.String 仍可写空，不会 panic）。
func (e *Engine) DeviceUIDHex(deviceAddr string) string {
	if deviceAddr == "" {
		return ""
	}
	e.mu.RLock()
	hex := e.deviceAddrToUID[deviceAddr]
	e.mu.RUnlock()
	return hex
}

// SetRoomTenant 注入 roomID → tenant_id（alarm_events.tenant_id 必填，从 rooms 表查）。
func (e *Engine) SetRoomTenant(roomID, tenantID string) {
	if roomID == "" {
		return
	}
	e.mu.Lock()
	e.roomTenants[roomID] = tenantID
	e.mu.Unlock()
}

// SetSuiteCensus 注入进程级共享 SuiteCensusManager（sensor_v2 PR-2 数据结构 + PR-3 publish 关联）。
// nil = 禁用 PR-3 PersonID 关联（TrackStatus.PersonID 一律空）；
// bootstrap 调用方负责生命周期（含 SaveToRedis 定时任务）。

// OnDeviceUnfit 实现 DeviceFitnessTracker UnfitCallback —— device offline / SensorDetached
// 等触发时清该 device 在 engine 内的所有 in-memory state（"device offline = 内存重启"原则，
// 用户 2026-05-19 拍板）。
//
// 通过 deviceRoom 反查 device 所在 room → 调 tm.ClearDevice。device 跨多 room 不存在（v2
// device→room LPM 命中唯一），单 room dispatch 足够。
func (e *Engine) OnDeviceUnfit(deviceAddr string) {
	if e == nil || deviceAddr == "" {
		return
	}
	e.mu.RLock()
	roomID := e.deviceRoom[deviceAddr]
	tm := e.rooms[roomID]
	e.mu.RUnlock()
	if tm == nil {
		return
	}
	tm.ClearDevice(deviceAddr)
	e.logger.Info("engine cleared device on unfit",
		zap.String("device_addr", deviceAddr),
		zap.String("room_id", roomID),
	)
}

func (e *Engine) SetSuiteCensus(m *SuiteCensusManager) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.suiteCensus = m
}

// GridForRoom 暴露 grid 给 BathroomGhostAdjudicator 用（lookup callback）。
// 走 RLock 读 e.grids — 与 RegisterRoom 加锁路径互斥但快速通过。
// 返回 nil = room 未注册（caller 据此 noop）。
func (e *Engine) GridForRoom(roomID string) *RoomGrid {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.grids[roomID]
}

// SuiteIDForRoom 暴露 roomID → suiteID 给 BathroomGhostAdjudicator 用。
func (e *Engine) SuiteIDForRoom(roomID string) string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.roomSuiteID[roomID]
}

// smallBathroomMaxSideCm P6.1b-D(审查㉛ Opt-1):bbox 最小边 ≤ 此 = 小卫生间。
// 来源:审查⑳ 立项"小卫生间人活动范围≈全程贴门 → 门距信号退化"——门距没用才走 D 跨设备/分级。
const smallBathroomMaxSideCm = 200

// isSmallBathroomCfg P6.1b-D gate:RoomType==Bathroom ∧ bbox 最小边 ≤200cm。
// bbox 优先取 WallPolygon(真实轮廓);无 wall 则退 rawW×rawH(ApplyOptimizedExtent 前的原始房尺寸,
// 非 FOV 扩展后)。RegisterRoom 内调(持写锁)。
func isSmallBathroomCfg(cfg RoomConfig, rawW, rawH int) bool {
	if cfg.RoomType != card.RoomTypeBathroom {
		return false
	}
	w, h := rawW, rawH
	if len(cfg.WallPolygon) >= 3 {
		minX, minY := cfg.WallPolygon[0].X, cfg.WallPolygon[0].Y
		maxX, maxY := minX, minY
		for _, p := range cfg.WallPolygon[1:] {
			if p.X < minX {
				minX = p.X
			}
			if p.X > maxX {
				maxX = p.X
			}
			if p.Y < minY {
				minY = p.Y
			}
			if p.Y > maxY {
				maxY = p.Y
			}
		}
		w, h = maxX-minX, maxY-minY
	}
	minSide := w
	if h < minSide {
		minSide = h
	}
	return minSide > 0 && minSide <= smallBathroomMaxSideCm
}

// IsSmallBathroom 暴露 roomID → 是否小卫生间(P6.1b-D gate)给 belief_shadow 用。
func (e *Engine) IsSmallBathroom(roomID string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.smallBathroom[roomID]
}

// SuiteHasOtherDevice P6.1b-D(资源代理):同 unit 是否有 excludeDevice 以外的设备。
// 用途:小卫生间 lost provisional 的设备富/贫判定——设备富(true)→30min cancel 窗;独苗(false)→短窗压制。
// 设备密度 = 机构资源水平代理(v3 用户洞见)。读 deviceRoom→roomSuiteID 数同 suite 设备(只读)。
func (e *Engine) SuiteHasOtherDevice(suiteID, excludeDevice string) bool {
	if suiteID == "" {
		return false
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	for dev, roomID := range e.deviceRoom {
		if dev == excludeDevice {
			continue
		}
		if e.roomSuiteID[roomID] == suiteID {
			return true
		}
	}
	return false
}

// trackLostAnchorMs 失锁判定阈值：snapshot 未含某 trackID 持续 ≥ 60s → 视为真失锁，
// 调 SuiteCensusManager.ClearAnchorOnLostTrack。偏保守的取值（MaxMissCount=10 帧 coast
// 期之外再加足够 buffer），避免 firmware track coast 期间误清。
// 详见 suite_census.go ClearAnchorOnLostTrack 注释 "PR-3 wire 失锁判定建议"。
const trackLostAnchorMs int64 = 60_000

// routeRoomFrame 馈送出口：失锁清理 census anchor + bathroom 入口流量 + PR-3 PersonID
// 关联（更新 census 身份态），然后把 bases 交 OnRoomFrame seam 给上层 DBN 裁决。
// Engine 不建 TrackStatus / 不跑 GhostAdjudicator / 不推 redis——决策权全在上层。
//
// Bathroom room 不调 UpdatePersonFromTrack —— bathroom 内 person 关联由 PR-5 BathroomGate
// 入口流量 + suiteCensus.MarkPersonExitToBathroom/Return 维护，bathroom 帧只读 AnchorRoomType。
func (e *Engine) routeRoomFrame(roomID string, bases []TrackStatusBase, nowMs int64) {
	e.mu.RLock()
	suiteID := e.roomSuiteID[roomID]
	residentID := e.roomResidentID[roomID]
	roomType := e.roomType[roomID]
	census := e.suiteCensus
	e.mu.RUnlock()

	isBathroom := roomType == card.RoomTypeBathroom

	// 步骤 1：失锁清理（trackLastSeen 仅在本 goroutine 读写，无锁需要）。
	seen := e.trackLastSeen[roomID]
	if seen == nil {
		seen = make(map[int]int64)
		e.trackLastSeen[roomID] = seen
	}
	curIDs := make(map[int]struct{}, len(bases))
	for i := range bases {
		curIDs[bases[i].TrackID] = struct{}{}
	}
	for tid, lastMs := range seen {
		if _, alive := curIDs[tid]; alive {
			continue
		}
		if nowMs-lastMs < trackLostAnchorMs {
			continue
		}
		if census != nil && suiteID != "" {
			census.ClearAnchorOnLostTrack(suiteID, tid)
		}
		delete(seen, tid)
	}

	// 步骤 1.5：PR-5 BathroomGate（仅 bathroom room + 已挂 census + suiteID 三条件齐备）。
	if isBathroom && census != nil && suiteID != "" {
		gate, ok := e.bathroomGates[roomID]
		if !ok {
			gate = NewBathroomGate(roomID, suiteID, census, e.logger)
			e.bathroomGates[roomID] = gate
		}
		gate.Process(bases, nowMs)
	}

	// 步骤 2：PR-3 PersonID 关联——更新 census 身份态（DBN 经 census 读，不再投影到 wire）。
	for i := range bases {
		b := &bases[i]
		seen[b.TrackID] = nowMs
		if census != nil && suiteID != "" && !isBathroom {
			census.UpdatePersonFromTrack(
				suiteID, residentID,
				b.TrackID, b.SleepadInBed,
				b.TraverseDelta, b.MoveActive,
				nowMs,
			)
		}
	}

	if e.OnRoomFrame != nil {
		var bed card.BedState
		var exitLogOdds, ghostLeftLogOdds func(logicID string, atMs int64) float64
		var hardExited func(logicID string, atMs int64) bool
		tm := e.rooms[roomID]
		if tm != nil {
			bed = tm.BedOccupancyState(nowMs) // room 级权威 bed 读数（sleepad+radar 床事件融合）→ B 轴
			// 离房 SLeft 对数几率（ExitRoom 硬 + trend+np 软），按 LogicID 反查：事件无坐标走不了 census 关联，
			//   且丢轨 12s 驱逐后 base 空——闭包持 tm（recentRadarEvents/lostExitInfo 按 age 淘汰，不随 track drop）。
			exitLogOdds = tm.ExitLogOdds
			// present 静止 ghost「已离房」SLeft（room-empty 信号 + 门距 + 时间门，硬门否决真摔），喂 present 分支。
			ghostLeftLogOdds = tm.GhostLeftLogOdds
			hardExited = tm.HardExited // 逐帧离房事件 hard-drop（绕过 SLeft 阈）
		}
		fired, dropped := e.OnRoomFrame(roomID, bases, bed, nowMs, exitLogOdds, ghostLeftLogOdds, hardExited)
		if tm != nil {
			tm.DiagDupLids(nowMs) // 临时诊断，待删
			for _, lid := range fired {
				tm.DiagFireSameLid(lid, nowMs) // 临时诊断，待删
				tm.ResetStillBox(lid)          // fall fire → still-box 从 0 热机（belief 已在 DBN 侧就地复位）
			}
			for _, lid := range dropped {
				tm.EvictTrack(lid) // 确认离场/空 → 立即删 track，停 12s coast 期 re-feed（防 census churn）
			}
		}
	}
}

// ========================================================================
// RegisterRoom：构建 grid + rasterize 物理/几何/先验
// ========================================================================

// chairPriorConf = layout Chair SetPrior 成 AreaSit 的初始置信度（镜像 wisefido-sensor 正本）。
const chairPriorConf = 80

func (e *Engine) RegisterRoom(cfg RoomConfig) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if cfg.RoomW <= 0 || cfg.RoomW > radarutils.MaxRoomWidth {
		e.logger.Warn("room width out of range, using default",
			zap.Int("requested", cfg.RoomW), zap.Int("max", radarutils.MaxRoomWidth))
		cfg.RoomW = radarutils.MaxRoomWidth
	}
	if cfg.RoomH <= 0 || cfg.RoomH > radarutils.MaxRoomHeight {
		cfg.RoomH = radarutils.MaxRoomHeight
	}

	// 优化 grid 范围：bbox(WallPolygon ∪ FOV) + 4 cells，cap 600cm
	// 对 cell 索引语义有影响 → 旧 snapshot 的 layoutHash 也会随之变（如预期）
	rawW, rawH := cfg.RoomW, cfg.RoomH
	ApplyOptimizedExtent(&cfg)

	// 幂等 short-circuit：room 已注册 + layout 未变 → 仅刷 soft 字段，**不**重建 TrackManager/grid。
	// 否则 config:card.changed 触发的 ReloadRooms 会把所有 17 个 room 的 bedSessions / fall pending /
	// trackLastSeen / sleepadStates 等 in-memory 状态归零，导致正在床上 / 跌倒判定窗口丢失。
	// layout 真的改了（用户编辑了 wall/bed/radar 等）→ 通过 hash 比对触发完整重建（旧行为）。
	newHash := LayoutHash(cfg)
	if tm, exists := e.rooms[cfg.RoomID]; exists && e.layoutHashes[cfg.RoomID] == newHash {
		// soft-only：room name / timezone / 静态归属（resident_id / suite_id / room_type）+ public bathroom 标记
		if cfg.RoomName != "" {
			tm.SetRoomName(cfg.RoomName)
		}
		tm.SetRoomType(cfg.RoomType)
		if cfg.Timezone != "" {
			if loc, err := time.LoadLocation(cfg.Timezone); err == nil {
				tm.SetTimezone(loc)
			}
		}
		e.roomSuiteID[cfg.RoomID] = cfg.SuiteID
		e.roomResidentID[cfg.RoomID] = cfg.ResidentID
		e.roomType[cfg.RoomID] = cfg.RoomType
		e.smallBathroom[cfg.RoomID] = isSmallBathroomCfg(cfg, rawW, rawH)
		if cfg.RoomType == card.RoomTypeBathroom && cfg.IsPublicBathroom && e.suiteCensus != nil && cfg.SuiteID != "" {
			e.suiteCensus.MarkPublicBathroom(cfg.SuiteID) // 幂等
		}
		return
	}

	// 1. 创建空 grid，覆盖优化后的 RoomW × RoomH
	grid := NewRoomGrid(cfg.RoomW, cfg.RoomH, radarutils.CellSize)
	grid.OriginX = cfg.OriginX
	grid.OriginY = cfg.OriginY

	// 2. Stamp Wall 多边形 → InRoom
	if len(cfg.WallPolygon) >= 3 {
		grid.StampRoomPolygon(cfg.WallPolygon)
	}

	// 3. Stamp 每台 Radar 物理 FOV → InFOV / EdgeDist / MaxZ / MinZ。
	// 多雷达房逐台 stamp（FOV 取并集）；单雷达房 cfg.Radars 含唯一一台。
	// 空 cfg.Radars（placeholder / 老路径）回退到 cfg.Radar。
	radars := cfg.Radars
	if len(radars) == 0 {
		radars = []radarutils.RadarMount{cfg.Radar}
	}
	for _, m := range radars {
		grid.StampRadar(m)
	}

	// 4. Stamp Enters → 记 Enters 列表 + 覆写矩形内 InRoom=true（门洞可穿）
	grid.StampEnters(cfg.Enters, cfg.EnterTargets)

	// 5. 空间轴单源：声明区留在 grid.AreaZones（DB canvas 记录副本），QueryAreaType 直接查。
	//    不再烙进 belief——人标与 learned 分两源，UpdateBelief 侵蚀碰不到声明区。
	grid.AreaZones = cfg.AreaZones

	e.grids[cfg.RoomID] = grid
	// MM covers 几何：per-device 保留各台**自己 canvas** 的 bed 矩形（RadarBedReachCount 只数本台 layout
	// 的床，不用合并后的房级床——多雷达房 device-local 帧未配准，借别台的床判定无意义）。
	for addr, beds := range cfg.DeviceBeds {
		e.deviceBeds[addr] = append([]radarutils.Rect(nil), beds...)
		e.deviceBedHs[addr] = append([]int(nil), cfg.DeviceBedHeights[addr]...)
	}
	// per-device mount：每台 radar 的 deviceAddr → 自己 mount，handleMessage 按帧来源 device 取。
	// 单雷达 / 老路径无 RadarAddrs 时回退把 roomID 当 key（与历史 e.mounts[roomID] 等价）。
	if len(cfg.RadarAddrs) == len(radars) && len(cfg.RadarAddrs) > 0 {
		for i, m := range radars {
			e.deviceMounts[cfg.RadarAddrs[i]] = m
		}
	} else {
		e.deviceMounts[cfg.RoomID] = cfg.Radar
	}
	tm := NewTrackManager(cfg.RoomID, grid, cfg.BedAreaIDs)
	tm.bedCount = countRealBeds(cfg.BedAreaTypes) // 同房多雷达占用对账单床闸（仅 ==1 启用）；LongSofa 不计
	tm.SetMoveSpeedCms(e.learnParams.MoveSpeedCms)
	tm.SetSitLearnParams(e.learnParams.SitPromoteTau, e.learnParams.SitSpreadCm)
	tm.SetBedsideFallConfig(e.bedsideFallCfg)
	tm.SetStaticScanIntervalMs(e.staticScanMs)
	tm.SetLogger(e.logger)
	tm.SetRoomName(cfg.RoomName)
	tm.SetRoomType(cfg.RoomType)
	tm.SetInterferes(cfg.Interferes)
	tm.SetChairs(cfg.Chairs, cfg.ChairIDs) // floor 连续 tFloor 的 chair 区 gate（电子云几何）+ object_id 锚定久坐学习态
	tm.SetRadarMount(cfg.Radar)        // L1 mirror pair 检测用 radar 中心做 ghost tiebreaker
	tm.SetWallPolygon(cfg.WallPolygon) // 静止反射体检测判"近墙"
	tm.SetUIDHexResolver(e.DeviceUIDHex)
	// 注入 IANA 时区（IsNightTime 用）；空串保持 nil → IsNightTime 退化 UTC
	if cfg.Timezone != "" {
		if loc, err := time.LoadLocation(cfg.Timezone); err == nil {
			tm.SetTimezone(loc)
		} else {
			e.logger.Warn("invalid timezone, falling back to UTC",
				zap.String("room_id", cfg.RoomID),
				zap.String("timezone", cfg.Timezone),
				zap.Error(err))
		}
	} else {
		e.logger.Warn("room registered without timezone, IsNightTime will use UTC (likely wrong)",
			zap.String("room_id", cfg.RoomID))
	}
	e.rooms[cfg.RoomID] = tm

	// sensor_v2 PR-3 wiring：记录 room → (suite_id, resident_id, room_type)。
	e.roomSuiteID[cfg.RoomID] = cfg.SuiteID
	e.roomResidentID[cfg.RoomID] = cfg.ResidentID
	e.roomType[cfg.RoomID] = cfg.RoomType
	e.smallBathroom[cfg.RoomID] = isSmallBathroomCfg(cfg, rawW, rawH) // P6.1b-D gate

	// sensor_v2 PR-7：public bathroom 显式标记 census bucket。
	// 仅当 RoomType==Bathroom 且 IsPublicBathroom=true 才触发，其它 room 即使 IsPublicBathroom=true 也忽略。
	// 幂等：MarkPublicBathroom 多次调同 SuiteID 等价（PR-2 实现）。
	if cfg.RoomType == card.RoomTypeBathroom && cfg.IsPublicBathroom && e.suiteCensus != nil && cfg.SuiteID != "" {
		e.suiteCensus.MarkPublicBathroom(cfg.SuiteID)
		e.logger.Info("room_registered_as_public_bathroom",
			zap.String("room_id", cfg.RoomID),
			zap.String("suite_id", cfg.SuiteID),
		)
	}

	e.layoutHashes[cfg.RoomID] = newHash
	hash := newHash

	e.logger.Info("room registered",
		zap.String("room_id", cfg.RoomID),
		zap.String("room_name", cfg.RoomName),
		zap.Int("raw_w", rawW),
		zap.Int("raw_h", rawH),
		zap.Int("grid_w", cfg.RoomW),
		zap.Int("grid_h", cfg.RoomH),
		zap.Int("enters", len(cfg.Enters)),
		zap.Int("beds", len(cfg.Beds)),
		zap.Int("area_zones", len(cfg.AreaZones)),
		zap.Int("cells", grid.Width*grid.Height),
		zap.String("layout_hash", hash[:12]),
	)
}

// ========================================================================
// Run：订阅 iot:monitor:stream 主循环 + 后台定时任务
// ========================================================================

func (e *Engine) Run(ctx context.Context) error {
	const (
		monitorStream = "test:iot:monitor:stream"
		eventStream   = "test:iot:event:stream"
		alarmStream   = "test:iot:alarm:stream"
		group         = "roomengine"
	)

	if err := rediscommon.CreateConsumerGroup(ctx, e.redisClient, monitorStream, group); err != nil {
		e.logger.Warn("create consumer group (monitor)", zap.Error(err))
	}
	if err := rediscommon.CreateConsumerGroup(ctx, e.redisClient, eventStream, group); err != nil {
		e.logger.Warn("create consumer group (event)", zap.Error(err))
	}
	if err := rediscommon.CreateConsumerGroup(ctx, e.redisClient, alarmStream, group); err != nil {
		e.logger.Warn("create consumer group (alarm)", zap.Error(err))
	}

	// 单独 goroutine 消费 event 流（sleepad + radar 的 InBed/LeftBed/EnterRoom/ExitRoom）
	go e.runEventLoop(ctx, eventStream, group)
	// 单独 goroutine 消费 alarm 流（radar Fall 等）
	go e.runAlarmLoop(ctx, alarmStream, group)
	// 路由表 reload 由 config:card:stream subscriber 事件驱动（cmd/wisefido-sensor main.go configCardConsumer）。

	e.logger.Info("room engine started",
		zap.String("monitor_stream", monitorStream),
		zap.String("event_stream", eventStream),
		zap.String("alarm_stream", alarmStream),
	)

	for {
		select {
		case <-ctx.Done():
			e.logger.Info("room engine stopped")
			return nil
		default:
		}

		messages, err := rediscommon.ReadFromStream(ctx, e.redisClient, monitorStream, group, "roomengine-monitor-1", 50)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}

		for _, msg := range messages {
			e.handleMessage(ctx, msg)
			// XAck 每条消息（即便路由失败也 ACK——路由层已 warnUnrouted 暴露问题，
			// 不 ACK 只会让 pending 列表无限增长直到 redis 内存爆）。
			e.redisClient.XAck(ctx, monitorStream, group, msg.ID)
		}
	}
}

// runEventLoop 消费 iot:event:stream，路由 sleepad InBed/LeftBed 事件到 TrackManager
func (e *Engine) runEventLoop(ctx context.Context, stream, group string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		messages, err := rediscommon.ReadFromStream(ctx, e.redisClient, stream, group, "roomengine-event-1", 50)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}
		for _, msg := range messages {
			e.handleEventMessage(msg)
			e.redisClient.XAck(ctx, stream, group, msg.ID)
		}
	}
}

// handleEventMessage 处理 iot:event:stream 一条消息（sleepad + radar 来源都消费）
func (e *Engine) handleEventMessage(msg rediscommon.StreamMessage) {
	m, err := rediscommon.FromStreamMap(msg.Values)
	if err != nil {
		return
	}
	dt := strings.ToLower(m.DeviceType)
	if dt != "sleepad" && dt != "sleeppad" && dt != "radar" {
		return
	}
	ts := m.Timestamp
	if ts == 0 {
		ts = time.Now().UnixMilli()
	}

	// 路由到房间（addr 是唯一 device 标识）
	addrStr := m.DeviceAddr.String()
	e.mu.RLock()
	roomID := e.deviceRoom[addrStr]
	tm := e.rooms[roomID]
	e.mu.RUnlock()
	if tm == nil {
		e.warnUnrouted("event", m.SubjectEntity, addrStr, m.DeviceType)
		return
	}

	// 触发事件 X 光打点（每个 event 一行，与 per-tick xsensor_xray 按 ts 对齐）。
	e.logger.Info("xsensor_event",
		zap.String("room", roomID), zap.String("dev_type", dt),
		zap.String("category", m.Category), zap.Int64("ts", ts), zap.String("addr", addrStr))

	switch dt {
	case "sleepad", "sleeppad":
		for _, evt := range ParseSleepadBedEvents(m.DataValue, addrStr, m.Category, ts) {
			tm.ProcessSleepadBedEvent(evt)
		}
		// sleepad 事件驱动 DBN tick：sleepad-only 房无 radar 帧、这是唯一驱动；radar 房也吃 bed 翻转更新。
		e.routeRoomFrame(roomID, tm.SnapshotTrackStatuses(ts), ts)
	case "radar":
		// 2026-05-15 起 radar firmware Fall/SittingOnGround 也走 event stream（gateway 分流）。
		// 见 doc/cardagg_sensor_split.md：cardagg 不再处理 radar producer Fall；sensor 接管 verifier。
		// Suspected* (pose=2/7) 是 firmware qualification 前的预警态，30~90s 后自动升级到 Fall/SittingOnGround。
		// Per Option C (2026-05-17): server 不双层 — Suspected* 只让 iot 自动写 event_log（审计链不丢），
		// 不进 verifier / 不 forward 到 iot:alarm:stream / 不入 alarm_events。
		// 见 memory firmware_fall_qualification + doc/cardagg_sensor_split.md
		if m.Category == alarm.SuspectedFall || m.Category == alarm.SuspectedSittingOnGround {
			e.logger.Debug("radar suspected (event_log only, no alarm forward)",
				zap.String("device_addr", addrStr),
				zap.String("category", m.Category),
				zap.Int64("ts_ms", ts),
			)
			return
		}
		if m.Category == alarm.Fall || m.Category == alarm.SittingOnGround {
			alarms := ParseRadarFallAlarm(m.DataValue, addrStr, m.Category, ts)
			for _, a := range alarms {
				tm.RecordRadarAlarm(a)
				e.logger.Info("radar_fall_received_via_event_stream",
					zap.String("device_addr", addrStr),
					zap.String("device_uid_hex", e.DeviceUIDHex(addrStr)),
					zap.String("category", m.Category),
					zap.Int("track_id", a.TrackID),
					zap.Int("pose", a.Pose),
					zap.String("status", a.Status),
					zap.String("room_id", roomID),
					zap.Int64("ts_ms", a.TMs),
					zap.String("ts_human", time.UnixMilli(a.TMs).Format("15:04:05.000")),
				)
			}
			tm.Tick(ts)
			return
		}
		// 落账 radar EnterRoom/ExitRoom/InBed/LeftBed；同时 InBed/LeftBed 走"事件触发器"
		// 路径：tm.RecordRadarEvent + tm.Tick(ts) → 段 4/5/6 立即跑一次。
		evts := ParseRadarTrackEvents(m.DataValue, addrStr, m.Category, ts)
		if len(evts) == 0 {
			return
		}
		for _, evt := range evts {
			tm.RecordRadarEvent(evt)
		}
		tm.Tick(ts)
	}
}

// runAlarmLoop 消费 iot:alarm:stream，路由 radar Fall 报警（仅落账 + Tick 触发，不做 verify）
func (e *Engine) runAlarmLoop(ctx context.Context, stream, group string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		messages, err := rediscommon.ReadFromStream(ctx, e.redisClient, stream, group, "roomengine-alarm-1", 50)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}
		for _, msg := range messages {
			e.handleAlarmMessage(msg)
			e.redisClient.XAck(ctx, stream, group, msg.ID)
		}
	}
}

// handleAlarmMessage 处理 iot:alarm:stream 一条消息。
//
// 2026-05-15 gateway 分流后：alarm stream 上的 radar Fall 几乎全是 sensor 自己 emit 的回声
// （RecordRadarAlarm 转发回 alarm stream，producer="sensor.caregiver01"）。
// 必须 guard 防自循环：检测到 producer 是 sensor 自身 emit 时直接跳过，不再 RecordRadarAlarm + emit。
//
// 真正的 firmware radar Fall 现在走 iot:event:stream → handleEventMessage 接管。
func (e *Engine) handleAlarmMessage(msg rediscommon.StreamMessage) {
	m, err := rediscommon.FromStreamMap(msg.Values)
	if err != nil {
		return
	}
	// 自循环 guard：sensor 自身 emit 的 alarm 不重新处理（否则 RecordRadarAlarm → 再 emit → 再消费 → 无限循环）
	// 2026-05-25 修：sensor publish 用 platform-agent IP (fd00:0:fff1::1) 作 producer
	// （见 memory platform_agent_addressing），旧字符串 "sensor.caregiver01"/"wisefido-sensor"
	// 已不再使用 → 必须用 IsPlatformAgentAddr 识别，否则 self-loop guard 失效 (实测放大 43729×)。
	if spatial.IsPlatformAgentAddr(m.Producer) {
		return
	}
	if !strings.EqualFold(m.DeviceType, "radar") {
		return // 非 radar 报警暂不消费
	}
	ts := m.Timestamp
	if ts == 0 {
		ts = time.Now().UnixMilli()
	}

	// 路由（addr 是唯一 device 标识）
	addrStr := m.DeviceAddr.String()
	e.mu.RLock()
	roomID := e.deviceRoom[addrStr]
	tm := e.rooms[roomID]
	e.mu.RUnlock()
	if tm == nil {
		e.warnUnrouted("alarm", m.SubjectEntity, addrStr, m.DeviceType)
		return
	}

	alarms := ParseRadarFallAlarm(m.DataValue, addrStr, m.Category, ts)
	if len(alarms) == 0 {
		return
	}
	for _, a := range alarms {
		tm.RecordRadarAlarm(a)
		// kind=radar_fall_received: radar firmware 发的 Fall 报警；engine 当前不验证（段 7 待做）。
		// 未来 verifier 上线后会进一步分流：fake_fall（否决）/ real_fall（确认）/ lost_fall（应报漏报）
		e.logger.Info("radar_fall_received",
			zap.String("device_addr", addrStr),
			zap.String("device_uid_hex", e.DeviceUIDHex(addrStr)),
			zap.Int("track_id", a.TrackID),
			zap.String("kind", "radar_firmware_fall"),
			zap.Int("pose", a.Pose),
			zap.String("status", a.Status),
			zap.String("room_id", roomID),
			zap.Int64("ts_ms", a.TMs),
			zap.String("ts_human", time.UnixMilli(a.TMs).Format("15:04:05.000")),
		)
	}
	tm.Tick(ts)
}

// ========================================================================
// 消息处理 + 坐标转换 + 无效帧过滤
// ========================================================================

func (e *Engine) handleMessage(_ context.Context, msg rediscommon.StreamMessage) {
	// 真实生产消息是 flat key（device_uid/dataValue/.../timestamp 都是 stream Values 顶层 key），
	// 不是包在 "data" 单 key 里的 JSON。用 FromStreamMap 与 cardagg/wisefido-data 一致解析。
	m, err := rediscommon.FromStreamMap(msg.Values)
	if err != nil {
		return
	}

	// 路由到房间（addr 是唯一 device 标识）
	addrStr := m.DeviceAddr.String()
	e.mu.RLock()
	roomID := e.deviceRoom[addrStr]
	tm := e.rooms[roomID]
	// per-device mount：按帧来源 device 取自己 mount（多雷达房各台独立坐标系）；
	// 回退 roomID key 容老路径（RegisterRoom 无 RadarAddrs 时按 roomID 存）。
	mount, hasMount := e.deviceMounts[addrStr]
	if !hasMount {
		mount, hasMount = e.deviceMounts[roomID]
	}
	e.mu.RUnlock()

	if tm == nil {
		e.warnUnrouted("monitor", m.SubjectEntity, addrStr, m.DeviceType)
		return
	}

	ts := m.Timestamp
	if ts == 0 {
		ts = time.Now().UnixMilli()
	}

	switch strings.ToLower(m.DeviceType) {
	case "radar":
		if !hasMount {
			return
		}
		frames := ParseRadarTracks(m.DataValue, addrStr, mount, ts)
		if len(frames) == 0 {
			// tid=88 heartbeat / 全零无效帧被 ParseRadarTracks 过滤后，仍要 tick 推进 MissCount，
			// 否则之前活着的 track 永不进入消失判定 → silent/lost fall pending 不会创建。
			// NoTargetTick 标记"固件明示无目标"，触发 88-加速驱逐（治 Case2 142s 陈旧 track）。
			tm.NoTargetTick(ts, addrStr)
			// 空帧（无目标）仍喂 DBN：bases 反映 track 缺席 → engine.Room 对消失 track 仅 Predict
			// 自持（blind 续存，S 留 Fallen，告警连续），fall 确认窗继续累积。若此处 return，丢轨期间
			// DBN 永不 tick → 确认窗冻结 → fall fire 延迟到 track 重现（lost-fall 缺口）。
			bases := tm.SnapshotTrackStatuses(ts)
			e.routeRoomFrame(roomID, bases, ts)
			return
		}
		outputs := tm.ProcessFrame(frames)
		if e.onOutput != nil && len(outputs) > 0 {
			e.onOutput(roomID, outputs)
		}
		// 馈送出口：Layer 1 track 投影 → seam，交上层 DBN 裁决。
		// 只在 radar 帧后派生（sleepad 帧不带 track 几何）；
		// SnapshotTrackStatuses 自带 tm.mu，与 ProcessFrame 串行无竞态。
		bases := tm.SnapshotTrackStatuses(ts)
		e.routeRoomFrame(roomID, bases, ts)

	case "sleepad", "sleeppad":
		// 多传感器融合：sleepad 帧喂入 TrackManager，silent fall 触发时做 short-circuit
		obs := ParseSleepadObservations(m.DataValue, addrStr, addrStr, ts)
		for _, o := range obs {
			tm.ProcessSleepadObservation(o)
		}

		// 兼容老接口：把"在床人数"也推给 TrackManager（occupancyFactor 用）
		inBedCount := 0
		for _, o := range obs {
			if o.InBed {
				inBedCount++
			}
		}
		tm.SetSleepadInBedCount(inBedCount)
	}
}

// ========================================================================
// 类型转换 helper（保留原有）
// ========================================================================

func intFromAny(v interface{}) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case int64:
		return int(x)
	case json.Number:
		n, _ := x.Int64()
		return int(n)
	case string:
		n, _ := strconv.Atoi(x)
		return n
	}
	return 0
}

func int64FromAny(v interface{}) int64 {
	switch x := v.(type) {
	case float64:
		return int64(x)
	case int:
		return int64(x)
	case int64:
		return x
	case json.Number:
		n, _ := x.Int64()
		return n
	}
	return 0
}
