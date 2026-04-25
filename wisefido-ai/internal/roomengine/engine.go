package roomengine

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"

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
	RoomID  string
	RoomW   int // 画布宽（cm）
	RoomH   int // 画布深（cm）
	OriginX int // grid[0][0] 左上角的画布坐标 X（让 grid 覆盖物体 bbox）
	OriginY int // grid[0][0] 左上角的画布坐标 Y

	// Wall 围出的房间多边形（闭合），用于 StampRoomPolygon
	WallPolygon []radarutils.Point

	// 人工标注的矩形先验
	Enters     []radarutils.Rect // AreaEnter
	Beds       []radarutils.Rect // AreaBed
	Toilets    []radarutils.Rect // AreaToilet
	Showers    []radarutils.Rect // AreaShower
	Chairs     []radarutils.Rect // AreaSit（粗标沙发/椅子，Conf=80）
	Furnitures []radarutils.Rect // AreaDeny（家具/桌子）
	Interferes []radarutils.Rect // AreaDeny（镜子/金属反射区）

	// 雷达安装
	Radar radarutils.RadarMount

	// Sleepad 位置（point 几何，可能多个）— 用于事件路由 + 可视化
	Sleepads []radarutils.Point
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
	rooms      map[string]*TrackManager        // roomID → TrackManager
	grids      map[string]*RoomGrid            // roomID → Grid
	mounts     map[string]radarutils.RadarMount // roomID → Radar 安装参数（坐标转换用）
	cardToRoom map[string]string                // cardID → roomID
	deviceRoom map[string]string                // deviceUID/deviceID → roomID

	// 自适应参数
	paramSets [3]ParamSet
	accuracy  [3]AccuracyTracker
	winner    int // 当前 winner 组（0/1/2），-1=无 winner 用 baseline

	// 定时器
	decayInterval      time.Duration // 默认 1 小时（Decay 计算一次）
	decayHalfLifeSec   float64       // 短档半衰期（秒）
	beliefScanInterval time.Duration // 默认 5 分钟（全图 UpdateBelief）
	winnerEvalInterval time.Duration // 默认 24 小时（winner 重评）

	redisClient *redis.Client
	logger      *zap.Logger

	onOutput func(roomID string, outputs []TrackOutput)
}

// NewEngine 创建 Room Engine
func NewEngine(redisClient *redis.Client, logger *zap.Logger) *Engine {
	return &Engine{
		rooms:              make(map[string]*TrackManager),
		grids:              make(map[string]*RoomGrid),
		mounts:             make(map[string]radarutils.RadarMount),
		cardToRoom:         make(map[string]string),
		deviceRoom:         make(map[string]string),
		paramSets:          DefaultParamSets,
		winner:             1, // 默认 balanced
		decayInterval:      1 * time.Hour,
		decayHalfLifeSec:   float64(HalfLifeShort), // 15 min（cell.go 定义）
		beliefScanInterval: 5 * time.Minute,
		winnerEvalInterval: 24 * time.Hour,
		redisClient:        redisClient,
		logger:             logger,
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

func (e *Engine) MapDeviceToRoom(deviceKey, roomID string) {
	e.mu.Lock()
	e.deviceRoom[deviceKey] = roomID
	e.mu.Unlock()
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

	// 1. 创建空 grid，覆盖 cfg.RoomW × cfg.RoomH，origin 由 layout 决定
	grid := NewRoomGrid(cfg.RoomW, cfg.RoomH, radarutils.CellSize)
	if cfg.OriginX != 0 || cfg.OriginY != 0 {
		grid.OriginX = cfg.OriginX
		grid.OriginY = cfg.OriginY
	}

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
	e.rooms[cfg.RoomID] = NewTrackManager(cfg.RoomID, grid)

	e.logger.Info("room registered",
		zap.String("room_id", cfg.RoomID),
		zap.Int("width_cm", cfg.RoomW),
		zap.Int("height_cm", cfg.RoomH),
		zap.Int("enters", len(cfg.Enters)),
		zap.Int("beds", len(cfg.Beds)),
		zap.Int("toilets", len(cfg.Toilets)),
		zap.Int("showers", len(cfg.Showers)),
		zap.Int("furnitures", len(cfg.Furnitures)),
		zap.Int("cells", grid.Width*grid.Height),
	)
}

// ========================================================================
// Run：订阅 iot:monitor:stream 主循环 + 后台定时任务
// ========================================================================

func (e *Engine) Run(ctx context.Context) error {
	streamName := "iot:monitor:stream"
	group := "roomengine"
	consumer := "roomengine-1"

	if err := rediscommon.CreateConsumerGroup(ctx, e.redisClient, streamName, group); err != nil {
		e.logger.Warn("create consumer group", zap.Error(err))
	}

	// 后台定时任务
	go e.decayLoop(ctx)
	go e.beliefScanLoop(ctx)
	go e.winnerEvalLoop(ctx)

	e.logger.Info("room engine started",
		zap.String("stream", streamName),
		zap.String("winner", e.paramSets[e.winner].Name),
	)

	for {
		select {
		case <-ctx.Done():
			e.logger.Info("room engine stopped")
			return nil
		default:
		}

		messages, err := rediscommon.ReadFromStream(ctx, e.redisClient, streamName, group, consumer, 50)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}

		for _, msg := range messages {
			e.handleMessage(ctx, msg)
		}
	}
}

// ========================================================================
// 消息处理 + 坐标转换 + 无效帧过滤
// ========================================================================

func (e *Engine) handleMessage(_ context.Context, msg rediscommon.StreamMessage) {
	dataStr, ok := msg.Values["data"].(string)
	if !ok {
		return
	}

	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(dataStr), &raw); err != nil {
		return
	}

	deviceID, _ := raw["device_id"].(string)
	deviceUID, _ := raw["device_uid"].(string)
	cardID, _ := raw["card_id"].(string)
	deviceType, _ := raw["device_type"].(string)

	if !strings.EqualFold(deviceType, "radar") {
		return
	}

	// 路由到房间
	e.mu.RLock()
	roomID := e.cardToRoom[cardID]
	if roomID == "" {
		roomID = e.deviceRoom[deviceID]
	}
	if roomID == "" {
		roomID = e.deviceRoom[deviceUID]
	}
	tm := e.rooms[roomID]
	mount, hasMount := e.mounts[roomID]
	e.mu.RUnlock()

	if tm == nil || !hasMount {
		return
	}

	// 解析 + 坐标转换 + 过滤
	frames := e.parseTrackFrames(raw, deviceID, mount)
	if len(frames) == 0 {
		return
	}

	outputs := tm.ProcessFrame(frames)

	if e.onOutput != nil && len(outputs) > 0 {
		e.onOutput(roomID, outputs)
	}
}

// parseTrackFrames 把 iot:monitor:stream 一条消息里的多 track 拆解。
// 内部委托 ParseRadarTracks（同包，也供 playback 工具直接调用）。
func (e *Engine) parseTrackFrames(raw map[string]interface{}, deviceID string, mount radarutils.RadarMount) []TrackFrame {
	dv := raw[rediscommon.DataValueKey]
	if dv == nil {
		dv = raw["dataValue"]
	}
	ts := int64FromAny(raw["timestamp"])
	if ts == 0 {
		ts = time.Now().UnixMilli()
	}
	return ParseRadarTracks(dv, deviceID, mount, ts)
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
			for _, grid := range e.grids {
				grid.DecayAll(dtSec, e.decayHalfLifeSec)
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
	for _, grid := range e.grids {
		grid.LearnCellAreas()
		totalLieAnomalies += grid.LearnLyingAnomalies()
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
