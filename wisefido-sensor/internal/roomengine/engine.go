package roomengine

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"

	"owl-common/observation"
	"owl-common/radarutils"
	rediscommon "owl-common/redis"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// ========================================================================
// RoomConfig：从 layout_config JSONB 解析得到的房间配置
// ========================================================================

// RoomConfig 房间配置（全 int 化 + 对齐 radarutils 类型）
type RoomConfig struct {
	RoomID   string
	RoomName string // rooms.room_name；wisefido-sensor 用 owl-common/roomutil.ClassifyRoomType 判 still fall 区
	RoomW    int    // 画布宽（cm）
	RoomH    int    // 画布深（cm）
	OriginX  int    // grid[0][0] 左上角的画布坐标 X（让 grid 覆盖物体 bbox）
	OriginY  int    // grid[0][0] 左上角的画布坐标 Y

	// Wall 围出的房间多边形（闭合），用于 StampRoomPolygon
	WallPolygon []radarutils.Point

	// 人工标注的矩形先验
	Enters     []radarutils.Rect // AreaEnter
	Beds       []radarutils.Rect // AreaBed
	Toilets    []radarutils.Rect // AreaToilet
	Showers    []radarutils.Rect // AreaShower
	Chairs     []radarutils.Rect // AreaSit（粗标沙发/椅子，Conf=80）
	Furnitures []radarutils.Rect // AreaDeny（家具/桌子）
	Interferes []radarutils.Rect // AreaDeny（镜子/金属反射区/吊灯）

	// 物体顶部高度（cm，地面为 0），与上面同名切片一一对应（Heights[i] ↔ Beds[i]）。
	// 来源：layout JSON 里每个对象的 height 字段，由前端 Toolbar 录入；
	// 缺失时 ParseLayoutConfig 用 FurnitureType 默认值（与前端 FURNITURE_CONFIGS.defaultHeight 对齐）。
	// 当前 RoomEngine 不使用，仅持久化保留——未来用于：
	//   - 空中物体（吊灯 height>200）不阻挡通行 → InFOV 不刷 Deny
	//   - 床/桌面高度参与 Z 轴范围判定 → fall 检测加先验
	EnterHeights     []int
	BedHeights       []int
	ToiletHeights    []int
	ShowerHeights    []int
	ChairHeights     []int
	FurnitureHeights []int
	InterfereHeights []int

	// 雷达安装
	Radar radarutils.RadarMount

	// Sleepad 位置（point 几何，可能多个）— 用于事件路由 + 可视化
	Sleepads []radarutils.Point

	// Timezone IANA 时区字符串（如 "America/Denver"），来自该 room 所属 unit 的 units.timezone。
	// 用于 IsNightTime 判定 risk-time（默认 23:30-07:30 本地时间）。
	// 空值 → engine 退化为 UTC（错位风险，必须设置）。
	Timezone string
}

// ========================================================================
// ParamSet[3] 与 winner 选择
// ========================================================================

// DefaultParamSets 三组并行参数（保守/中庸/激进）
var DefaultParamSets = [3]ParamSet{
	{Alpha: 0.01, Beta: 0.2, FlipTh: 10, Name: "conservative"},
	{Alpha: 0.02, Beta: 0.5, FlipTh: 20, Name: "balanced"},
	{Alpha: 0.05, Beta: 1.0, FlipTh: 30, Name: "aggressive"},
}

// AccuracyTracker 单组参数的准确率累计（供 winner 选择使用）
// TP / FP / TN / FN 由 feedback_loop.go 灌入
type AccuracyTracker struct {
	TruePositive  int
	FalsePositive int
	TrueNegative  int
	FalseNegative int
	LastEvalAt    time.Time
}

// Accuracy 准确率 [0,1]；样本不足返回 -1
func (a *AccuracyTracker) Accuracy() float64 {
	total := a.TruePositive + a.FalsePositive + a.TrueNegative + a.FalseNegative
	if total < 5 {
		return -1
	}
	return float64(a.TruePositive+a.TrueNegative) / float64(total)
}

// ========================================================================
// Engine
// ========================================================================

type Engine struct {
	mu         sync.RWMutex
	rooms      map[string]*TrackManager         // roomID → TrackManager
	grids      map[string]*RoomGrid             // roomID → Grid
	mounts     map[string]radarutils.RadarMount // roomID → Radar 安装参数（坐标转换用）
	cardToRoom map[string]string                // cardID → roomID
	deviceRoom map[string]string                // deviceUID/deviceID → roomID
	// PR-8: AI publish 用 — 源 radar UUID 反查 deviceUID + room 反查 tenant/card
	deviceIDToUID   map[string]string // deviceID(UUID) → device_uid（IoTStreamMessage.DeviceUID）
	deviceIDToType  map[string]string // deviceID(UUID) → 源 sensor 类型（"Radar"/"Sleepad"），AI publish 加 ".AI<node>" 后缀
	roomTenants     map[string]string // roomID → tenant_id（alarm_events 必填）
	// 不再维护 roomCards：AI 是 sensor 层 producer，只发 device 标识；card_id (subject_entity)
	// 由 cardagg IotPreparedHandler 反查 device→cards 路由（多卡共享设备时自然 fan-out）。
	// 协议层北极星 layer 1 / 2 解耦原则。详 doc/TODO.md 第 0 项。

	// Sensor agent 节点完整身份字符串（如 "sensor.caregiver01"），同时作 envelope.Producer
	// + wire fields["source"] + ai_emit log 审计字段。多实例横向扩展时各进程 source 不同 →
	// wire 自动区分实例。来源见 cfg.AIPublish.Source。Phase B：命名规范 sensor.<role><实例号>。
	aiSource string

	// publish 模式："log" | "log&publish"。
	// "log" 模式仅写 sensor.log，不发 redis stream；任何模式都不影响 alarm 触发路径。
	aiPublishMode string

	// layout 几何 hash，RegisterRoom 时计算；snapshot save/load 比对用
	layoutHashes map[string]string

	// 自适应参数
	paramSets [3]ParamSet
	accuracy  [3]AccuracyTracker
	winner    int // 当前 winner 组（0/1/2），-1=无 winner 用 baseline

	// 运行时参数（由 Configure 注入；Decay/Learn 字段语义见 cell.go / cell_learning.go）
	decayParams     DecayParams
	learnParams     LearnParams
	bedsideFallCfg  BedsideFallConfig // R4 床边晕倒参数；RegisterRoom 时下发到 TrackManager

	// 定时器
	decayInterval      time.Duration // 默认 1 小时（Decay 计算一次）
	beliefScanInterval time.Duration // PR-11: 默认 10 分钟（原 5min；降低 CPU 功率）
	winnerEvalInterval time.Duration // 默认 24 小时（winner 重评）
	snapshotInterval   time.Duration // 默认 5 分钟（持久化全量 dump）；0 = 关闭

	// 持久化（nil = 不持久化）
	persister Persister

	// 历史归档（nil = 不归档）。每天 dailySnapshotHour:dailySnapshotMinute 触发，保留 historyRetainDays 天
	historyPersister    HistoryPersister
	dailySnapshotHour   int // 0-23 local；-1=禁用 daily snapshot
	dailySnapshotMinute int // 0-59 local
	historyRetainDays   int // <=0 表示不清理

	// alarm-feedback ingestion（cell.IncrFakeAlarm 反馈链）
	feedbackDB        *sql.DB
	feedbackInterval  time.Duration // 默认 5 分钟；0 = 关闭
	feedbackIngester  *AlarmFeedbackIngester

	// PR-15: 每日定时 layout reload。dailyReloadHour 0-23（local time）；-1=禁用
	// 目的：管理员下班后重读 rooms.layout_config，hash 变 → 重置该 room grid，从 0 重学
	dailyReloadHour int
	dailyReloadDB   *sql.DB // 用于 SELECT layout_config；nil 时跳过

	// 路由表周期热加载（启动后才绑 device→room 的设备永远进不来——见 handleMessage tm==nil）
	// 注入方式：bootstrap 调 SetRoutesReloader 传入"重新跑 mapDevicesToRooms"的闭包
	routesReloader       func(context.Context) error
	routesReloadInterval time.Duration

	// 路由失败的 device 频率限制告警（避免每条 frame 都 warn 一次）。
	// 同一 device key 60s 内只 warn 一次，确保 deviceRoom 缺失能被发现而不淹日志。
	unroutedMu sync.Mutex
	unrouted   map[string]int64 // device key → last warn epoch ms

	redisClient *redis.Client
	logger      *zap.Logger

	onOutput func(roomID string, outputs []TrackOutput)
}

// RuntimeConfig 与 wisefido-sensor/internal/config::RoomEngineConfig 一一对应；
// engine 包不依赖 config 包，wiring 在 cmd/wisefido-sensor/main.go 中完成转换。
type RuntimeConfig struct {
	Decay              DecayParams
	Learn              LearnParams
	ParamSets          [3]ParamSet
	DecayInterval      time.Duration
	BeliefScanInterval time.Duration
	WinnerEvalInterval time.Duration
	SnapshotInterval   time.Duration // 0 = 关闭持久化定时器；Persister 仍可在退出时 dump
	Persister          Persister     // nil = 不持久化

	// HistoryPersister + DailySnapshotHour/Minute + HistoryRetainDays：每日归档
	HistoryPersister    HistoryPersister
	DailySnapshotHour   int // -1=禁用；默认 11（11:50 local）
	DailySnapshotMinute int // 0-59；默认 50
	HistoryRetainDays   int // 默认 365；<=0 不清理

	// 风险时段（夜间）；通过 SetRiskTimeConfig 注入到包级 nightCfg。
	// 全 0 视为未设置，保留 IsNightTime 默认（23:30 - 07:30）。
	RiskTime RiskTimeConfig

	// R4 床边晕倒参数；通过 TrackManager.SetBedsideFallConfig 注入到每个房间。
	// 任一字段 0 保留 defaultBedsideFallCfg 默认值。
	BedsideFall BedsideFallConfig

	// FeedbackDB / FeedbackInterval：alarm_events false_alarm 反馈链。
	// 二者都设置才启用；缺任一退化为不启动 feedbackLoop。
	FeedbackDB       *sql.DB
	FeedbackInterval time.Duration // 默认 5 分钟
}

// NewEngine 创建 Room Engine（默认参数）。生产环境调用 Configure 注入 yaml 配置。
func NewEngine(redisClient *redis.Client, logger *zap.Logger) *Engine {
	return &Engine{
		rooms:              make(map[string]*TrackManager),
		grids:              make(map[string]*RoomGrid),
		mounts:             make(map[string]radarutils.RadarMount),
		cardToRoom:         make(map[string]string),
		deviceRoom:         make(map[string]string),
		deviceIDToUID:      make(map[string]string),
		deviceIDToType:     make(map[string]string),
		roomTenants:        make(map[string]string),
		aiSource:           "sensor.caregiver01",
		aiPublishMode:      "log&publish",
		layoutHashes:       make(map[string]string),
		paramSets:          DefaultParamSets,
		winner:             1, // 默认 balanced
		decayParams:        DefaultDecayParams(),
		learnParams:        DefaultLearnParams(),
		decayInterval:        1 * time.Hour,
		beliefScanInterval:   10 * time.Minute, // PR-11
		winnerEvalInterval:   24 * time.Hour,
		snapshotInterval:     5 * time.Minute,
		dailyReloadHour:      22, // PR-15：22:00 local 重读 layout
		dailySnapshotHour:    11, // 每天 11:50 local 归档 daily history
		dailySnapshotMinute:  50,
		historyRetainDays:    365, // 一年滚动清理
		routesReloadInterval: 60 * time.Second, // 路由表周期热加载默认 60s
		unrouted:             make(map[string]int64),
		redisClient:          redisClient,
		logger:               logger,
	}
}

// SetRoutesReloader 注入路由表（device→room、card→room）周期热加载闭包。
// 解决 handleMessage 在 deviceRoom 缺失时静默丢帧的问题——启动后才绑定/重绑的
// device 不再永远沉默。reloader 通常是"重跑 mapDevicesToRooms"的 wrapper。
// interval ≤ 0 时取默认 60s。
func (e *Engine) SetRoutesReloader(fn func(context.Context) error, interval time.Duration) {
	if interval <= 0 {
		interval = 60 * time.Second
	}
	e.mu.Lock()
	e.routesReloader = fn
	e.routesReloadInterval = interval
	e.mu.Unlock()
}

// warnUnrouted 路由失败的 device 频率限制告警（同一 key 60s 内一次）。
// stream：消息来自哪个 stream（"monitor"/"event"/"alarm"）便于排查。
func (e *Engine) warnUnrouted(stream, cardID, deviceID, deviceUID, deviceType string) {
	key := deviceUID
	if key == "" {
		key = deviceID
	}
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
		zap.String("device_uid", deviceUID),
		zap.String("device_id", deviceID),
		zap.String("card_id", cardID),
		zap.String("device_type", deviceType),
		zap.String("hint", "device not in deviceRoom/cardToRoom; if just bound after startup, will heal at next routes reload"),
	)
}

// routesReloadLoop 周期调用 routesReloader 热加载路由表。
func (e *Engine) routesReloadLoop(ctx context.Context) {
	e.mu.RLock()
	interval := e.routesReloadInterval
	reloader := e.routesReloader
	e.mu.RUnlock()
	if reloader == nil || interval <= 0 {
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := reloader(ctx); err != nil {
				e.logger.Warn("routes_reload_failed", zap.Error(err))
			}
		}
	}
}

// SetDailyLayoutReload 注入 daily layout reload 配置。
// hourLocal：0-23 表示 local time 时刻；-1 禁用。
// db：用于 SELECT rooms.layout_config；nil 时禁用。
// PR-15：管理员下班后定时重读 layout，hash 变化 → 重置该 room grid 从 0 重学。
func (e *Engine) SetDailyLayoutReload(hourLocal int, db *sql.DB) {
	if hourLocal < -1 || hourLocal > 23 {
		hourLocal = -1
	}
	e.dailyReloadHour = hourLocal
	e.dailyReloadDB = db
}

// Configure 注入 yaml 加载的运行时参数。零值字段保留 New 时的默认值。
// Persister 字段单独判断：显式传 nil 即关闭持久化（覆盖默认）。
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
	if cfg.DecayInterval > 0 {
		e.decayInterval = cfg.DecayInterval
	}
	if cfg.BeliefScanInterval > 0 {
		e.beliefScanInterval = cfg.BeliefScanInterval
	}
	if cfg.WinnerEvalInterval > 0 {
		e.winnerEvalInterval = cfg.WinnerEvalInterval
	}
	if cfg.SnapshotInterval > 0 {
		e.snapshotInterval = cfg.SnapshotInterval
	}
	// Persister 直接赋值（nil 也接受，表示禁用）
	e.persister = cfg.Persister

	// HistoryPersister 与 daily snapshot 时刻；时刻字段 0 视为未覆盖（保留默认 11:50/365 天）
	e.historyPersister = cfg.HistoryPersister
	if cfg.DailySnapshotHour != 0 || cfg.DailySnapshotMinute != 0 {
		e.dailySnapshotHour = cfg.DailySnapshotHour
		e.dailySnapshotMinute = cfg.DailySnapshotMinute
	}
	if cfg.HistoryRetainDays > 0 {
		e.historyRetainDays = cfg.HistoryRetainDays
	}

	// 风险时段（IsNightTime 用）—— 包级 var，所有房间共享
	SetRiskTimeConfig(cfg.RiskTime)

	// R4 参数：保存到 engine 级 + 下发到所有已注册房间的 TrackManager。
	// （RegisterRoom 在 Configure 之前调用的房间也要补设。）
	e.bedsideFallCfg = cfg.BedsideFall
	for _, tm := range e.rooms {
		tm.SetBedsideFallConfig(cfg.BedsideFall)
	}

	// alarm feedback：默认 5min；缺 DB 不启用
	e.feedbackDB = cfg.FeedbackDB
	if cfg.FeedbackInterval > 0 {
		e.feedbackInterval = cfg.FeedbackInterval
	} else if e.feedbackInterval == 0 {
		e.feedbackInterval = 5 * time.Minute
	}
	if e.feedbackDB != nil {
		e.feedbackIngester = NewAlarmFeedbackIngester(e.feedbackDB, e, e.logger)
	} else {
		e.feedbackIngester = nil
	}
}

// SetOutputCallback 设置 track 输出回调（发 alarm 等下游）
func (e *Engine) SetOutputCallback(fn func(roomID string, outputs []TrackOutput)) {
	e.onOutput = fn
}

// MapCardToRoom / MapDeviceToRoom 路由表（启动时从 card/device meta 灌入）
func (e *Engine) MapCardToRoom(cardID, roomID string) {
	e.mu.Lock()
	e.cardToRoom[cardID] = roomID
	e.mu.Unlock()
}

// SetRoomStayAlarmEnabled 注入指定房间的 Stay alarm 启用状态。
// 由 bootstrap 查 alarm_device.monitor_config 后调用。
// 房间不存在或未注册 → 静默丢弃（warn 日志）。
func (e *Engine) SetRoomStayAlarmEnabled(roomID string, enabled bool) {
	e.mu.RLock()
	tm := e.rooms[roomID]
	e.mu.RUnlock()
	if tm == nil {
		e.logger.Warn("SetRoomStayAlarmEnabled: room not registered", zap.String("room_id", roomID))
		return
	}
	tm.SetStayAlarmEnabled(enabled)
}

// RoomForDevice 查 deviceKey（device_id 或 device_uid）对应的 room_id。
// 用于 alarm feedback 反查。空字符串 = 未路由（设备未绑定房间或未注册）。
func (e *Engine) RoomForDevice(deviceKey string) string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.deviceRoom[deviceKey]
}

// MountForRoom 取指定房间的雷达安装参数（用于 RadarToCanvas 转换）。
// found=false 表示 room 未注册或未 stamp radar。
func (e *Engine) MountForRoom(roomID string) (radarutils.RadarMount, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	mount, ok := e.mounts[roomID]
	return mount, ok
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

func (e *Engine) MapDeviceToRoom(deviceKey, roomID string) {
	e.mu.Lock()
	e.deviceRoom[deviceKey] = roomID
	e.mu.Unlock()
}

// MapDeviceIDToUID 注册 device UUID → device_uid（PR-8 AI publish 反查源 radar UID 用）。
func (e *Engine) MapDeviceIDToUID(deviceID, deviceUID string) {
	if deviceID == "" || deviceUID == "" {
		return
	}
	e.mu.Lock()
	e.deviceIDToUID[deviceID] = deviceUID
	e.mu.Unlock()
}

// MapDeviceIDToType 注册 device UUID → 源 sensor 类型（"Radar"/"Sleepad"）。
// AI publish 时直接作 message.DeviceType（不再拼 AI 后缀）。缺失时兜底 "Radar"。
func (e *Engine) MapDeviceIDToType(deviceID, deviceType string) {
	if deviceID == "" || deviceType == "" {
		return
	}
	e.mu.Lock()
	e.deviceIDToType[deviceID] = deviceType
	e.mu.Unlock()
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

// SetAIPublishConfig 注入 AI publish 单点配置：mode + node_id（来自 yaml/env）。
// 替换旧的 SetAIDeviceType + SetAIPublishEnabled，单一入口避免漂移。
//
// mode："log" 仅写 ai.log；"log&publish" 还会推 redis stream。任何模式下 alarm
// 的 fire 路径都不变（"宁可误报不可漏报"原则——mode 仅控制下游 stream，不 gate
// 告警生成）。
//
// source：完整节点身份字符串（"AI.Caregiver01" / "AI.Doctor01" 等），同时作
// wire fields["source"] 默认值 + ai_emit log 审计字段。直接由 config 注入，
// 不在此拼接——多实例横向扩展时各进程 source 不同即可。
// device_type 保持源 sensor 类型，AI 派生身份在 dataValue.source 表达。
func (e *Engine) SetAIPublishConfig(mode, source string) {
	if mode == "" {
		mode = "log&publish"
	}
	if source == "" {
		source = "sensor.caregiver01"
	}
	e.mu.Lock()
	e.aiPublishMode = mode
	e.aiSource = source
	e.mu.Unlock()
}

// publishEnabled 当前是否会真发到 redis（mode == "log&publish"）。
func (e *Engine) publishEnabled() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.aiPublishMode == "log&publish"
}

// PublishAIEvent 发布 AI 派生 event 到 iot:event:stream。
// 当前用例：category="track_verdict"（PR5 后 ghost 走这条；旧代码可能仍传 "ghost" 兼容）。
//
// 消息字段（Phase A v2 envelope）：
//   Producer:        e.aiSource（如 "sensor.caregiver01"），sensor agent layer 1 标识
//   SubjectEntity:   留空（AI 不染卡概念，cardagg IotPreparedHandler 反查 device→subject）
//   DeviceUID:       源 sensor 的 device_uid
//   DeviceID:        源 sensor 的 UUID
//   DeviceType:      源 sensor 类型（"Radar" / "Sleepad"）
//   TenantID:        roomTenants[roomID]
//   DataValue:       [{ track_id, ts, position_x/y/z, area_type, pose, track_confidence, source, ... }]
//
// 推送失败仅 warn 日志，不阻塞调用方。任何模式都打 ai_emit 结构化日志，作演示
// 与审计追溯依据；mode=log 时 published=false 仅 log，mode=log&publish 时尝试推流。
func (e *Engine) PublishAIEvent(ctx context.Context, p AIPayload, category string, nowMs int64) {
	e.publishAIMessage(ctx, p, category, "event",
		rediscommon.StreamEvent.Name,
		rediscommon.StreamEvent.MaxLen, rediscommon.StreamEvent.RetentionSeconds, nowMs)
}

// PublishAIAlarm 发布 AI 派生 alarm 到 iot:alarm:stream。
// category ∈ {"silent_fall", "still_fall", "lost_fall", "silent_leftbed_fall", "bedside_fall"}
//
// alarm 路径不受 mode 影响 fire 决策（"宁可误报不可漏报"），mode 仅控制是否推到 stream。
func (e *Engine) PublishAIAlarm(ctx context.Context, p AIPayload, category string, nowMs int64) {
	e.publishAIMessage(ctx, p, category, "alarm",
		rediscommon.StreamAlarm.Name,
		rediscommon.StreamAlarm.MaxLen, rediscommon.StreamAlarm.RetentionSeconds, nowMs)
}

func (e *Engine) publishAIMessage(ctx context.Context, p AIPayload,
	category, topicType, streamName string, maxLen int64, retentionSec int, nowMs int64) {
	e.mu.RLock()
	deviceUID := e.deviceIDToUID[p.DeviceID]
	baseType := e.deviceIDToType[p.DeviceID]
	tenantID := e.roomTenants[p.RoomID]
	defaultSource := e.aiSource
	mode := e.aiPublishMode
	g := e.grids[p.RoomID]
	e.mu.RUnlock()
	// SubjectEntity 留空：AI 是 sensor 层 producer，只发 device 标识；cardagg
	// IotPreparedHandler 反查 device→cards 路由（多卡共享设备时 fan-out）。

	if baseType == "" {
		baseType = "Radar" // 兜底：路由表缺失时默认按 radar 派生
	}
	// device_type 保持源 sensor 类型（不再拼 ".AI<NodeID>" 后缀）。
	// AI 派生身份由 fields["source"] 一等公民字段表达。
	deviceType := baseType

	// alarm/event 流统一走 EventItem 契约（与 qinglan/sleepace publisher 一致）：
	// EventItem 提供生命周期 + first-class 业务字段（TrackID/Pose/HeartRate/RespiratoryRate）；
	// 其余 sensor-specific 字段（position / track_confidence / area_type / source / reason / evidence）
	// 作为 dataValue map 同级平铺补充。
	item := observation.NewEventItem(nowMs, "instant")
	item.TrackID = p.Track.TrackID
	if p.Track.Pose != 0 {
		item.Pose = p.Track.Pose
	}
	if p.Track.HeartRate != 0 {
		item.HeartRate = p.Track.HeartRate
	}
	if p.Track.RespiratoryRate != 0 {
		item.RespiratoryRate = p.Track.RespiratoryRate
	}
	fields, _ := observation.EventItemToDataMap(&item)
	if fields == nil {
		fields = make(map[string]interface{})
	}
	// sensor-specific 业务扩展字段平铺
	if p.Track.PositionX != nil {
		fields[observation.FieldPositionX] = *p.Track.PositionX
	}
	if p.Track.PositionY != nil {
		fields[observation.FieldPositionY] = *p.Track.PositionY
	}
	if p.Track.PositionZ != nil {
		fields[observation.FieldPositionZ] = *p.Track.PositionZ
	}
	if p.Track.TrackConfidence != 0 {
		fields[observation.FieldTrackConfidence] = p.Track.TrackConfidence
	}
	// AI 派生 track_verdict 与床状态无关；仅 sleepad_radar_conflict 显式传 BedStatus=1 才保留。
	if p.Track.BedStatus != 0 {
		fields[observation.FieldBedStatus] = p.Track.BedStatus
	}
	// area_type engine 自己算（observation.Track 的 AreaType 是字符串，engine 这边类型不同）
	if g != nil {
		px, py := 0, 0
		if p.Track.PositionX != nil {
			px = *p.Track.PositionX
		}
		if p.Track.PositionY != nil {
			py = *p.Track.PositionY
		}
		if cell := g.CellAt(px, py); cell != nil {
			fields["area_type"] = areaTypeProtocolStr(cell.Belief[0].Type)
		}
	}
	// PR5b: Source（一等公民）+ Reason / Evidence（审计元数据）
	// Source 默认 = e.aiSource（来自 cfg.AIPublish.Source，如 "AI.Caregiver01"）。
	// p.Source 非空时尊重 caller override（未来多角色场景，如健康风险模块发 verdict）
	source := p.Source
	if source == "" {
		source = defaultSource
	}
	if source != "" {
		fields["source"] = source
	}
	if p.Reason != "" {
		fields["reason"] = p.Reason
	}
	if len(p.Evidence) > 0 {
		fields["evidence"] = p.Evidence
	}

	willPublish := e.redisClient != nil && mode == "log&publish"

	// 任何模式都打 ai_emit 审计日志：sandbox 演示靠这条 log 看 AI 在思考
	e.logger.Info("ai_emit",
		zap.String("source", source), // 节点身份兼审计字段（如 "AI.Caregiver01"）
		zap.String("mode", mode),
		zap.String("device_type", deviceType),
		zap.String("device_uid", deviceUID),
		zap.String("category", category),
		zap.String("topic_type", topicType),
		zap.String("would_publish_to", streamName),
		zap.Bool("published", willPublish),
		zap.Int("track_id", p.Track.TrackID),
		zap.Int("track_confidence", p.Track.TrackConfidence),
	)

	if !willPublish {
		return
	}

	// Phase A: Producer = sensor agent 标识（如 "sensor.caregiver01"）；
	// SubjectEntity 留空——cardagg 反查 device→cards 路由（layer 1 不染卡）。
	producer := source
	if producer == "" {
		producer = defaultSource
	}
	msg := rediscommon.IoTStreamMessage{
		Producer:         producer,
		SubjectEntity:    "",
		SemanticLocation: p.RoomID,
		DeviceUID:        deviceUID,
		DeviceID:         p.DeviceID,
		DeviceType:       deviceType,
		TenantID:         tenantID,
		Timestamp:        nowMs,
		TopicType:        topicType,
		Category:         category,
		DataValue:        []interface{}{fields},
	}
	if _, err := rediscommon.PublishToStream(ctx, e.redisClient, streamName, msg.ToStreamMap(), maxLen, retentionSec); err != nil {
		e.logger.Warn("ai_publish_failed",
			zap.String("source", source),
			zap.String("stream", streamName),
			zap.String("category", category),
			zap.String("device_uid", deviceUID),
			zap.String("device_id", p.DeviceID),
			zap.Error(err),
		)
	}
}

// areaTypeProtocolStr 把 engine 的 AreaType 映射成 observation.EnumAreaType 协议字符串。
// 协议：0=none / 1=custom / 2=bed / 3=interfer / 4=door / 5=monitor_bed / 6=sensing
// engine 内部枚举跟协议不完全对齐；做一个 best-effort 映射，cardagg 端按 device_type=Radar 处理。
func areaTypeProtocolStr(t AreaType) string {
	switch t {
	case AreaBed:
		return "bed"
	case AreaSit:
		return "custom" // 沙发/椅子归 custom（协议没单独 sit 类）
	case AreaToilet, AreaShower:
		return "monitor_bed" // 卫浴归 monitor_bed（最接近的 high-risk 区）
	case AreaEnter:
		return "door"
	case AreaDeny:
		return "interfer"
	case AreaActive:
		return "sensing"
	}
	return "none"
}

// GetRoomOutputs 查询某房间最新 track 输出
func (e *Engine) GetRoomOutputs(roomID string) []TrackOutput {
	e.mu.RLock()
	tm := e.rooms[roomID]
	e.mu.RUnlock()
	if tm == nil {
		return nil
	}
	return tm.GetOutputs()
}

// ========================================================================
// RegisterRoom：构建 grid + rasterize 物理/几何/先验
// ========================================================================

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

	// 1. 创建空 grid，覆盖优化后的 RoomW × RoomH
	grid := NewRoomGrid(cfg.RoomW, cfg.RoomH, radarutils.CellSize)
	grid.OriginX = cfg.OriginX
	grid.OriginY = cfg.OriginY

	// 2. Stamp Wall 多边形 → InRoom
	if len(cfg.WallPolygon) >= 3 {
		grid.StampRoomPolygon(cfg.WallPolygon)
	}

	// 3. Stamp Radar 物理 FOV → InFOV / EdgeDist / MaxZ / MinZ
	grid.StampRadar(cfg.Radar)

	// 4. Stamp Enters → 记 Enters 列表 + 覆写矩形内 InRoom=true（门洞可穿）
	grid.StampEnters(cfg.Enters)

	// 5. SetPrior 人标矩形（AreaType + Confidence + Source）
	for _, r := range cfg.Enters {
		grid.SetPrior(r, AreaEnter, 99, SourceHuman)
	}
	for _, r := range cfg.Beds {
		grid.SetPrior(r, AreaBed, 99, SourceHuman)
	}
	for _, r := range cfg.Toilets {
		grid.SetPrior(r, AreaToilet, 99, SourceHuman)
	}
	for _, r := range cfg.Showers {
		grid.SetPrior(r, AreaShower, 99, SourceHuman)
	}
	for _, r := range cfg.Chairs {
		grid.SetPrior(r, AreaSit, 80, SourceHuman) // 粗标 Conf=80
	}
	for _, r := range cfg.Furnitures {
		grid.SetPrior(r, AreaDeny, 99, SourceHuman)
	}
	for _, r := range cfg.Interferes {
		grid.SetPrior(r, AreaDeny, 99, SourceHuman)
	}

	e.grids[cfg.RoomID] = grid
	e.mounts[cfg.RoomID] = cfg.Radar
	tm := NewTrackManager(cfg.RoomID, grid)
	tm.SetMoveSpeedCms(e.learnParams.MoveSpeedCms)
	tm.SetBedsideFallConfig(e.bedsideFallCfg)
	tm.SetLogger(e.logger)
	tm.SetRoomName(cfg.RoomName)
	tm.SetInterferes(cfg.Interferes)
	// PR-8: 注入 AI 派生事件 / 告警发布器（engine 实现 AIPublisher 接口）
	tm.SetAIPublisher(e)
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

	// 计算 layout hash 并保存（snapshot save/load 都按此 hash 比对）
	hash := LayoutHash(cfg)
	e.layoutHashes[cfg.RoomID] = hash

	e.logger.Info("room registered",
		zap.String("room_id", cfg.RoomID),
		zap.String("room_name", cfg.RoomName),
		zap.Int("raw_w", rawW),
		zap.Int("raw_h", rawH),
		zap.Int("grid_w", cfg.RoomW),
		zap.Int("grid_h", cfg.RoomH),
		zap.Int("enters", len(cfg.Enters)),
		zap.Int("beds", len(cfg.Beds)),
		zap.Int("toilets", len(cfg.Toilets)),
		zap.Int("showers", len(cfg.Showers)),
		zap.Int("furnitures", len(cfg.Furnitures)),
		zap.Int("cells", grid.Width*grid.Height),
		zap.String("layout_hash", hash[:12]),
	)

	// 尝试加载已有 snapshot（hydrate Belief + counters）
	// persister 可能未配置（dev/test 模式），用 background ctx 容忍 DB 慢
	if e.persister != nil {
		e.hydrateRoom(cfg.RoomID, grid, hash)
	}
}

// hydrateRoom 从 persister 取回 snapshot 并灌入 grid。
// 失败/不匹配/无记录都不致命——退化为 fresh start（baseline 已由 SetPrior 烤好）。
// 调用方必须持有 e.mu.Lock。
func (e *Engine) hydrateRoom(roomID string, grid *RoomGrid, expectedHash string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	storedHash, payload, found, err := e.persister.Load(ctx, roomID)
	if err != nil {
		e.logger.Warn("snapshot load failed", zap.String("room_id", roomID), zap.Error(err))
		return
	}
	if !found {
		e.logger.Info("no snapshot found, fresh start", zap.String("room_id", roomID))
		return
	}
	if storedHash != expectedHash {
		e.logger.Warn("snapshot layout_hash mismatch, discarding (layout edited?)",
			zap.String("room_id", roomID),
			zap.String("stored", storedHash[:12]),
			zap.String("expected", expectedHash[:12]),
		)
		return
	}

	snap, err := UnmarshalSnapshot(payload)
	if err != nil {
		e.logger.Warn("snapshot unmarshal failed", zap.String("room_id", roomID), zap.Error(err))
		return
	}
	if err := DecodeSnapshot(snap, grid); err != nil {
		e.logger.Warn("snapshot decode failed", zap.String("room_id", roomID), zap.Error(err))
		return
	}
	e.logger.Info("snapshot hydrated",
		zap.String("room_id", roomID),
		zap.Int("cells", len(snap.Cells)),
	)
}

// ========================================================================
// Run：订阅 iot:monitor:stream 主循环 + 后台定时任务
// ========================================================================

func (e *Engine) Run(ctx context.Context) error {
	const (
		monitorStream = "iot:monitor:stream"
		eventStream   = "iot:event:stream"
		alarmStream   = "iot:alarm:stream"
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

	// 后台定时任务
	go e.decayLoop(ctx)
	go e.beliefScanLoop(ctx)
	go e.winnerEvalLoop(ctx)
	if e.persister != nil && e.snapshotInterval > 0 {
		go e.snapshotLoop(ctx)
	}
	if e.feedbackIngester != nil && e.feedbackInterval > 0 {
		go e.alarmFeedbackLoop(ctx)
	}
	// PR-15: daily layout reload — 管理员下班后重读 layout_config，hash 变 → reset grid
	if e.dailyReloadHour >= 0 && e.dailyReloadDB != nil {
		go e.dailyLayoutReloadLoop(ctx)
	}
	// Daily history snapshot — 每天指定时刻归档 grid 状态，保留 365 天
	if e.historyPersister != nil && e.dailySnapshotHour >= 0 {
		go e.dailySnapshotLoop(ctx)
	}
	// 单独 goroutine 消费 event 流（sleepad + radar 的 InBed/LeftBed/EnterRoom/ExitRoom）
	go e.runEventLoop(ctx, eventStream, group)
	// 单独 goroutine 消费 alarm 流（radar Fall 等）
	go e.runAlarmLoop(ctx, alarmStream, group)
	// 路由表周期热加载（启动后才绑定的 device 不再永远沉默）
	if e.routesReloader != nil {
		go e.routesReloadLoop(ctx)
	}

	e.logger.Info("room engine started",
		zap.String("monitor_stream", monitorStream),
		zap.String("event_stream", eventStream),
		zap.String("alarm_stream", alarmStream),
		zap.String("winner", e.paramSets[e.winner].Name),
		zap.Bool("persist", e.persister != nil),
	)

	for {
		select {
		case <-ctx.Done():
			// 优雅退出：dump 一次再走（避免最近 5 min 学习丢失）
			if e.persister != nil {
				e.saveAllRooms(context.Background())
			}
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

	// 路由到房间
	e.mu.RLock()
	roomID := e.cardToRoom[m.SubjectEntity]
	if roomID == "" {
		roomID = e.deviceRoom[m.DeviceID]
	}
	if roomID == "" {
		roomID = e.deviceRoom[m.DeviceUID]
	}
	tm := e.rooms[roomID]
	e.mu.RUnlock()
	if tm == nil {
		e.warnUnrouted("event", m.SubjectEntity, m.DeviceID, m.DeviceUID, m.DeviceType)
		return
	}

	switch dt {
	case "sleepad", "sleeppad":
		for _, evt := range ParseSleepadBedEvents(m.DataValue, m.DeviceUID, m.Category, ts) {
			tm.ProcessSleepadBedEvent(evt)
		}
	case "radar":
		// 落账 radar EnterRoom/ExitRoom/InBed/LeftBed；同时 InBed/LeftBed 走"事件触发器"
		// 路径：tm.RecordRadarEvent + tm.Tick(ts) → 段 4/5/6 立即跑一次。
		// 当前不消费 EnterRoom/ExitRoom 做行为推断，仅落账供未来段 7 使用。
		evts := ParseRadarTrackEvents(m.DataValue, m.DeviceUID, m.Category, ts)
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

// handleAlarmMessage 处理 iot:alarm:stream 一条消息（仅 radar Fall）。
// 当前阶段：仅落账 + 立即 Tick；不否决 / 不延迟 / 不 verify。
// 未来段 7 (radar fall verify) 在 TrackManager 内消费 recentRadarAlarms 做 narrative。
func (e *Engine) handleAlarmMessage(msg rediscommon.StreamMessage) {
	m, err := rediscommon.FromStreamMap(msg.Values)
	if err != nil {
		return
	}
	if !strings.EqualFold(m.DeviceType, "radar") {
		return // 非 radar 报警暂不消费
	}
	ts := m.Timestamp
	if ts == 0 {
		ts = time.Now().UnixMilli()
	}

	// 路由
	e.mu.RLock()
	roomID := e.cardToRoom[m.SubjectEntity]
	if roomID == "" {
		roomID = e.deviceRoom[m.DeviceID]
	}
	if roomID == "" {
		roomID = e.deviceRoom[m.DeviceUID]
	}
	tm := e.rooms[roomID]
	e.mu.RUnlock()
	if tm == nil {
		e.warnUnrouted("alarm", m.SubjectEntity, m.DeviceID, m.DeviceUID, m.DeviceType)
		return
	}

	alarms := ParseRadarFallAlarm(m.DataValue, m.DeviceUID, m.Category, ts)
	if len(alarms) == 0 {
		return
	}
	for _, a := range alarms {
		tm.RecordRadarAlarm(a)
		// kind=radar_fall_received: radar firmware 发的 Fall 报警；engine 当前不验证（段 7 待做）。
		// 未来 verifier 上线后会进一步分流：fake_fall（否决）/ real_fall（确认）/ lost_fall（应报漏报）
		e.logger.Info("radar_fall_received",
			zap.String("device_uid", m.DeviceUID),
			zap.Int("track_id", a.TrackID),
			zap.String("kind", "radar_firmware_fall"),
			zap.Int("pose", a.Pose),
			zap.String("status", a.Status),
			zap.String("room_id", roomID),
			zap.Int64("ts_ms", a.TMs),
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

	// 路由到房间（radar/sleepad 共用同一路由表）
	e.mu.RLock()
	roomID := e.cardToRoom[m.SubjectEntity]
	if roomID == "" {
		roomID = e.deviceRoom[m.DeviceID]
	}
	if roomID == "" {
		roomID = e.deviceRoom[m.DeviceUID]
	}
	tm := e.rooms[roomID]
	mount, hasMount := e.mounts[roomID]
	e.mu.RUnlock()

	if tm == nil {
		e.warnUnrouted("monitor", m.SubjectEntity, m.DeviceID, m.DeviceUID, m.DeviceType)
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
		frames := ParseRadarTracks(m.DataValue, m.DeviceID, mount, ts)
		if len(frames) == 0 {
			// tid=88 heartbeat / 全零无效帧被 ParseRadarTracks 过滤后，仍要 tick 推进 MissCount，
			// 否则之前活着的 track 永不进入消失判定 → silent/lost fall pending 不会创建。
			tm.Tick(ts)
			return
		}
		outputs := tm.ProcessFrame(frames)
		if e.onOutput != nil && len(outputs) > 0 {
			e.onOutput(roomID, outputs)
		}

	case "sleepad", "sleeppad":
		// 多传感器融合：sleepad 帧喂入 TrackManager，silent fall 触发时做 short-circuit
		obs := ParseSleepadObservations(m.DataValue, m.DeviceID, m.DeviceUID, ts)
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
// 后台定时任务
// ========================================================================

// decayLoop 每 decayInterval 对所有 grid 做一次 DecayAll
func (e *Engine) decayLoop(ctx context.Context) {
	ticker := time.NewTicker(e.decayInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.mu.RLock()
			dtSec := e.decayInterval.Seconds()
			dp := e.decayParams
			for _, grid := range e.grids {
				grid.DecayAll(dtSec, dp)
			}
			e.mu.RUnlock()
			e.logger.Debug("decay all rooms done", zap.Float64("dt_sec", dtSec))
		}
	}
}

// beliefScanLoop 每 beliefScanInterval 对所有 cell 跑 UpdateBelief（3 组并行）
func (e *Engine) beliefScanLoop(ctx context.Context) {
	ticker := time.NewTicker(e.beliefScanInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.scanBeliefAll()
		}
	}
}

// scanBeliefAll 全量扫每个 grid 的每 cell：
//  1. cell_learning（硬阈值规则）：Walk/Sit 升格 + 床外 Lie 异常累计
//  2. UpdateBelief（软概率规则）：3 组参数各 UpdateBelief 一次
//
// 两者顺序：硬规则先跑（确定性强），UpdateBelief 后跑做细粒度调整。
func (e *Engine) scanBeliefAll() {
	e.mu.RLock()
	defer e.mu.RUnlock()
	totalCells := 0
	totalLieAnomalies := 0
	lp := e.learnParams
	nowMs := time.Now().UnixMilli()
	for _, grid := range e.grids {
		grid.LearnCellAreas(lp, nowMs)
		totalLieAnomalies += grid.LearnLyingAnomalies(lp)
		for i := range grid.Cells {
			for g := 0; g < 3; g++ {
				grid.Cells[i].UpdateBelief(g, e.paramSets[g])
			}
		}
		totalCells += len(grid.Cells)
	}
	e.logger.Debug("belief scan done",
		zap.Int("total_cells", totalCells),
		zap.Int("lie_anomalies", totalLieAnomalies),
		zap.Int("winner", e.winner),
	)
}

// alarmFeedbackLoop 每 feedbackInterval 跑一次 alarm_events false_alarm 反馈同步。
// 启动时立即跑一次（首次会全量扫历史，把过往 false_alarm 灌入 cell.FakeAlarmCount）。
// 单次失败只警告，不退出循环。
func (e *Engine) alarmFeedbackLoop(ctx context.Context) {
	if e.feedbackIngester == nil {
		return
	}
	// 启动延迟 5s，避免与 RegisterRoom / hydrate 同时进 DB
	select {
	case <-ctx.Done():
		return
	case <-time.After(5 * time.Second):
	}
	if n, err := e.feedbackIngester.IngestOnce(ctx); err != nil {
		e.logger.Warn("alarm_feedback ingest at startup", zap.Error(err))
	} else {
		e.logger.Info("alarm_feedback startup ingest", zap.Int("processed", n))
	}
	ticker := time.NewTicker(e.feedbackInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if n, err := e.feedbackIngester.IngestOnce(ctx); err != nil {
				e.logger.Warn("alarm_feedback ingest tick", zap.Error(err))
			} else if n > 0 {
				e.logger.Info("alarm_feedback ingest tick", zap.Int("processed", n))
			}
		}
	}
}

// snapshotLoop 每 snapshotInterval 把所有 grid 状态 dump 到 persister。
// 单次 dump 失败只警告，不退出循环（DB 抖动不应拖垮 engine）。
func (e *Engine) snapshotLoop(ctx context.Context) {
	ticker := time.NewTicker(e.snapshotInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.saveAllRooms(ctx)
		}
	}
}

// saveAllRooms 遍历所有 grid，逐房间 UPSERT 到 persister。
// 调用方可来自 snapshotLoop（带 ctx）或 Run 退出（用 background ctx）。
func (e *Engine) saveAllRooms(ctx context.Context) {
	if e.persister == nil {
		return
	}
	e.mu.RLock()
	// 拷贝 (roomID, grid, hash) 三元组到本地切片，释放锁后再做 IO，避免锁内 DB 写
	type roomDump struct {
		id   string
		grid *RoomGrid
		hash string
	}
	dumps := make([]roomDump, 0, len(e.grids))
	for id, g := range e.grids {
		dumps = append(dumps, roomDump{id: id, grid: g, hash: e.layoutHashes[id]})
	}
	e.mu.RUnlock()

	saved, failed := 0, 0
	for _, d := range dumps {
		snap := EncodeSnapshot(d.grid)
		payload, cellCount, err := MarshalSnapshot(snap)
		if err != nil {
			e.logger.Warn("snapshot marshal failed",
				zap.String("room_id", d.id), zap.Error(err))
			failed++
			continue
		}
		if err := e.persister.Save(ctx, d.id, d.hash, cellCount, payload); err != nil {
			e.logger.Warn("snapshot save failed",
				zap.String("room_id", d.id), zap.Error(err))
			failed++
			continue
		}
		saved++
	}
	e.logger.Debug("snapshot batch done",
		zap.Int("saved", saved), zap.Int("failed", failed))
}

// dailySnapshotLoop 每天 dailySnapshotHour:dailySnapshotMinute (local) 归档 grid 状态到 history 表。
// 区别于 snapshotLoop（每 5min 写 live 表，仅最新一份）：history 表按日期保留 365 天用于 playback 历史起点。
func (e *Engine) dailySnapshotLoop(ctx context.Context) {
	for {
		next := nextDailyTriggerHM(time.Now(), e.dailySnapshotHour, e.dailySnapshotMinute)
		wait := time.Until(next)
		e.logger.Info("daily_snapshot_scheduled",
			zap.Time("next", next),
			zap.Duration("wait", wait),
			zap.Int("hour_local", e.dailySnapshotHour),
			zap.Int("minute_local", e.dailySnapshotMinute),
			zap.Int("retain_days", e.historyRetainDays),
		)
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}
		e.saveAllRoomsHistory(ctx, time.Now())
	}
}

// saveAllRoomsHistory 遍历所有 grid，逐房间 SaveDaily 到 history 表。
// snapshotDate 用 nowLocal 的日期部分；retainDays 写入完成后顺手清理。
func (e *Engine) saveAllRoomsHistory(ctx context.Context, nowLocal time.Time) {
	if e.historyPersister == nil {
		return
	}
	e.mu.RLock()
	type roomDump struct {
		id   string
		grid *RoomGrid
		hash string
	}
	dumps := make([]roomDump, 0, len(e.grids))
	for id, g := range e.grids {
		dumps = append(dumps, roomDump{id: id, grid: g, hash: e.layoutHashes[id]})
	}
	e.mu.RUnlock()

	saved, failed := 0, 0
	for _, d := range dumps {
		snap := EncodeSnapshot(d.grid)
		payload, cellCount, err := MarshalSnapshot(snap)
		if err != nil {
			e.logger.Warn("daily snapshot marshal failed",
				zap.String("room_id", d.id), zap.Error(err))
			failed++
			continue
		}
		if err := e.historyPersister.SaveDaily(ctx, d.id, d.hash, nowLocal, cellCount, payload, e.historyRetainDays); err != nil {
			e.logger.Warn("daily snapshot save failed",
				zap.String("room_id", d.id), zap.Error(err))
			failed++
			continue
		}
		saved++
	}
	e.logger.Info("daily_snapshot_done",
		zap.Int("saved", saved),
		zap.Int("failed", failed),
		zap.String("snapshot_date", nowLocal.Format("2006-01-02")),
	)
}

// nextDailyTriggerHM 返回下一次 hour:minute 的 local 时间点；已过则取明天。
func nextDailyTriggerHM(now time.Time, hour, minute int) time.Time {
	t := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	if !t.After(now) {
		t = t.Add(24 * time.Hour)
	}
	return t
}

// dailyLayoutReloadLoop PR-15：每天 dailyReloadHour:00 (local) 重读 rooms.layout_config。
//
// 用途：管理员通过前端 layout 编辑器修改房间布局后，无需重启 wisefido-sensor 即可生效。
// 每日触发时刻应避开管理员上班时间（默认 22:00 = 晚 10 点）。
//
// 流程：
//  1. 等到下次 dailyReloadHour:00:00 local
//  2. 扫所有已注册 room，对比 DB 中 layout_config 的 hash 与内存 layoutHashes
//  3. hash 变 → 重置该 room（替换 TrackManager + 清空 grid 学习状态）
//  4. 立即持久化（覆盖旧 snapshot）
//
// 不变 hash 的 room 跳过；DB 查不到（room 已删）的房间也跳过。
func (e *Engine) dailyLayoutReloadLoop(ctx context.Context) {
	for {
		next := nextDailyTrigger(time.Now(), e.dailyReloadHour)
		wait := time.Until(next)
		e.logger.Info("daily_layout_reload_scheduled",
			zap.Time("next", next),
			zap.Duration("wait", wait),
			zap.Int("hour_local", e.dailyReloadHour),
		)
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}
		e.runDailyLayoutReload(ctx)
	}
}

// nextDailyTrigger 返回下一次 hour:00:00 local 的时间点。
// 若当前已经过当天 hour 时刻，则取明天。
func nextDailyTrigger(now time.Time, hour int) time.Time {
	t := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())
	if !t.After(now) {
		t = t.Add(24 * time.Hour)
	}
	return t
}

// runDailyLayoutReload 扫 rooms 表 → hash diff → 重置变化 room。
// 调用方：dailyLayoutReloadLoop 触发时；不持锁进入。
func (e *Engine) runDailyLayoutReload(ctx context.Context) {
	if e.dailyReloadDB == nil {
		return
	}
	rows, err := e.dailyReloadDB.QueryContext(ctx, `
		SELECT r.room_id::text, r.room_name, r.layout_config::text,
		       COALESCE(u.timezone, ''), r.tenant_id::text
		FROM rooms r
		LEFT JOIN units u ON u.unit_id = r.unit_id
		WHERE r.layout_config IS NOT NULL`)
	if err != nil {
		e.logger.Warn("daily_reload query failed", zap.Error(err))
		return
	}
	defer rows.Close()

	type pendingReload struct {
		cfg      RoomConfig
		tenantID string
		newHash  string
	}
	var pending []pendingReload
	for rows.Next() {
		var roomID, roomName, tenantID, timezone string
		var layoutStr sql.NullString
		if err := rows.Scan(&roomID, &roomName, &layoutStr, &timezone, &tenantID); err != nil {
			continue
		}
		if !layoutStr.Valid || layoutStr.String == "" {
			continue
		}
		cfg, err := ParseLayoutConfig(roomID, []byte(layoutStr.String))
		if err != nil {
			e.logger.Warn("daily_reload parse failed", zap.String("room_id", roomID), zap.Error(err))
			continue
		}
		cfg.RoomName = roomName
		cfg.Timezone = timezone
		ApplyOptimizedExtent(&cfg)
		newHash := LayoutHash(cfg)
		e.mu.RLock()
		oldHash := e.layoutHashes[roomID]
		e.mu.RUnlock()
		if oldHash == newHash {
			continue // 没变
		}
		pending = append(pending, pendingReload{cfg: cfg, tenantID: tenantID, newHash: newHash})
	}
	for _, pr := range pending {
		e.logger.Info("daily_reload room changed, resetting",
			zap.String("room_id", pr.cfg.RoomID),
			zap.String("new_hash", pr.newHash[:12]),
		)
		// RegisterRoom 替换 TrackManager + grid，hydrateRoom 见 hash 不同会丢弃旧 snapshot
		e.RegisterRoom(pr.cfg)
		e.SetRoomTenant(pr.cfg.RoomID, pr.tenantID)
	}
	if len(pending) > 0 && e.persister != nil {
		// 立刻持久化新状态，覆盖旧 snapshot
		e.saveAllRooms(ctx)
	}
}

// winnerEvalLoop 每 24 小时评估一次 winner（需要 feedback_loop 积累 accuracy 数据）
func (e *Engine) winnerEvalLoop(ctx context.Context) {
	ticker := time.NewTicker(e.winnerEvalInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.reevaluateWinner()
		}
	}
}

// reevaluateWinner 依据 accuracy[3] 选择新 winner。
// 规则：
//   - 若 3 组准确率都 < 样本阈值 → 不切换
//   - 若最高准确率 ≥ 雷达 baseline + 20% → 切到该组
//   - 否则维持当前 winner
//
// 实际 baseline（雷达直报准确率）需要 feedback_loop 单独统计，这里暂用 0.5 占位。
func (e *Engine) reevaluateWinner() {
	e.mu.Lock()
	defer e.mu.Unlock()

	const baselineAcc = 0.5 // 占位：二期从 alarm_events 统计雷达直报准确率

	best := -1
	bestAcc := -1.0
	for g := 0; g < 3; g++ {
		acc := e.accuracy[g].Accuracy()
		if acc < 0 {
			continue // 样本不足
		}
		if acc > bestAcc {
			best = g
			bestAcc = acc
		}
	}

	if best < 0 {
		e.logger.Debug("winner eval skipped: not enough samples")
		return
	}
	if bestAcc < baselineAcc+0.20 {
		e.logger.Info("winner eval: AI not beating baseline by 20%, keep current",
			zap.Float64("best_acc", bestAcc),
			zap.Float64("baseline", baselineAcc),
			zap.String("current_winner", e.paramSets[e.winner].Name),
		)
		return
	}
	if best == e.winner {
		return
	}
	e.logger.Info("winner switched",
		zap.String("from", e.paramSets[e.winner].Name),
		zap.String("to", e.paramSets[best].Name),
		zap.Float64("new_acc", bestAcc),
	)
	e.winner = best
}

// RecordGroundTruth 外部（feedback_loop.go）每收到一条家属反馈就调一次
// predicted: engine 在该 cell/track 上对三组的预测（如"是否 fall"），truthReal: 是否真 fall
func (e *Engine) RecordGroundTruth(predicted [3]bool, truthReal bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for g := 0; g < 3; g++ {
		switch {
		case predicted[g] && truthReal:
			e.accuracy[g].TruePositive++
		case predicted[g] && !truthReal:
			e.accuracy[g].FalsePositive++
		case !predicted[g] && truthReal:
			e.accuracy[g].FalseNegative++
		default:
			e.accuracy[g].TrueNegative++
		}
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
