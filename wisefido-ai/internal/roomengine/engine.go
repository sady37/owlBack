package roomengine

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"

	rediscommon "owl-common/redis"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RoomConfig 房间配置（从 layout_config 解析）。
type RoomConfig struct {
	RoomID    string
	RoomW     float64 // 房间宽度 cm
	RoomH     float64 // 房间深度 cm
	Entries   []Point // 门/入口坐标 cm
	Beds      []Rect  // 床区域
	Chairs    []Rect  // 椅子区域
	Toilets   []Rect  // 马桶区域
	Showers   []Rect  // 淋浴区域
	FOVEdges  []Point // 雷达 FOV 边缘采样点
}

// Rect 矩形区域 [X1,Y1] - [X2,Y2] cm。
type Rect struct {
	X1, Y1, X2, Y2 float64
}

// Engine Room Engine 主体：管理多个房间，订阅 iot:monitor:stream，逐帧打分。
type Engine struct {
	mu          sync.RWMutex
	rooms       map[string]*TrackManager // roomID → manager
	grids       map[string]*RoomGrid     // roomID → grid
	cardToRoom  map[string]string        // cardID → roomID（从 meta 加载）
	deviceRoom  map[string]string        // deviceID → roomID

	redisClient *redis.Client
	logger      *zap.Logger

	// 网格衰减周期（默认 1 小时执行一次，半衰期 24 小时）
	decayInterval time.Duration
	decayHalfLife float64 // 秒

	// 输出回调（通知下游：evaluator / alarm）
	onOutput func(roomID string, outputs []TrackOutput)
}

// NewEngine 创建 Room Engine。
func NewEngine(redisClient *redis.Client, logger *zap.Logger) *Engine {
	return &Engine{
		rooms:         make(map[string]*TrackManager),
		grids:         make(map[string]*RoomGrid),
		cardToRoom:    make(map[string]string),
		deviceRoom:    make(map[string]string),
		redisClient:   redisClient,
		logger:        logger,
		decayInterval: 1 * time.Hour,
		decayHalfLife: 86400, // 24 小时
	}
}

// SetOutputCallback 设置输出回调。
func (e *Engine) SetOutputCallback(fn func(roomID string, outputs []TrackOutput)) {
	e.onOutput = fn
}

// RegisterRoom 注册一个房间（启动时从 DB 加载 layout 后调用）。
func (e *Engine) RegisterRoom(cfg RoomConfig) {
	e.mu.Lock()
	defer e.mu.Unlock()

	grid := NewRoomGrid(cfg.RoomW, cfg.RoomH, 10)

	// 设置入口
	grid.Entries = append(grid.Entries, cfg.Entries...)
	grid.Entries = append(grid.Entries, cfg.FOVEdges...)
	for _, p := range cfg.Entries {
		grid.SetCellType(p.X-15, p.Y-15, p.X+15, p.Y+15, CellDoor, true)
	}

	// 设置家具区域
	for _, r := range cfg.Beds {
		grid.SetCellType(r.X1, r.Y1, r.X2, r.Y2, CellBed, true)
	}
	for _, r := range cfg.Chairs {
		grid.SetCellType(r.X1, r.Y1, r.X2, r.Y2, CellChair, true)
	}
	for _, r := range cfg.Toilets {
		grid.SetCellType(r.X1, r.Y1, r.X2, r.Y2, CellToilet, true)
	}
	for _, r := range cfg.Showers {
		grid.SetCellType(r.X1, r.Y1, r.X2, r.Y2, CellShower, true)
	}

	// 设置 FOV 边缘
	for _, p := range cfg.FOVEdges {
		grid.SetCellType(p.X-5, p.Y-5, p.X+5, p.Y+5, CellFOVEdge, true)
	}

	e.grids[cfg.RoomID] = grid
	e.rooms[cfg.RoomID] = NewTrackManager(cfg.RoomID, grid)

	e.logger.Info("room registered",
		zap.String("room_id", cfg.RoomID),
		zap.Float64("width_cm", cfg.RoomW),
		zap.Float64("height_cm", cfg.RoomH),
		zap.Int("entries", len(cfg.Entries)),
	)
}

// MapCardToRoom 映射 cardID → roomID（启动时从 card meta 加载）。
func (e *Engine) MapCardToRoom(cardID, roomID string) {
	e.mu.Lock()
	e.cardToRoom[cardID] = roomID
	e.mu.Unlock()
}

// MapDeviceToRoom 映射 deviceID → roomID。
func (e *Engine) MapDeviceToRoom(deviceID, roomID string) {
	e.mu.Lock()
	e.deviceRoom[deviceID] = roomID
	e.mu.Unlock()
}

// Run 主循环：订阅 iot:monitor:stream，逐帧处理。阻塞直到 ctx 取消。
func (e *Engine) Run(ctx context.Context) error {
	streamName := "iot:monitor:stream"
	group := "roomengine"
	consumer := "roomengine-1"

	// 创建消费者组
	if err := rediscommon.CreateConsumerGroup(ctx, e.redisClient, streamName, group); err != nil {
		e.logger.Warn("create consumer group", zap.Error(err))
	}

	// 启动衰减 goroutine
	go e.decayLoop(ctx)

	e.logger.Info("room engine started", zap.String("stream", streamName))

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

// handleMessage 处理单条 iot:monitor:stream 消息。
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

	// 只处理雷达数据
	if !strings.EqualFold(deviceType, "radar") {
		return
	}

	// 查找房间
	e.mu.RLock()
	roomID := e.cardToRoom[cardID]
	if roomID == "" {
		roomID = e.deviceRoom[deviceID]
	}
	if roomID == "" {
		roomID = e.deviceRoom[deviceUID]
	}
	tm := e.rooms[roomID]
	e.mu.RUnlock()

	if tm == nil {
		return
	}

	// 解析 data_value 中的 track 数据
	frames := e.parseTrackFrames(raw, deviceID)
	if len(frames) == 0 {
		return
	}

	// 处理一帧
	outputs := tm.ProcessFrame(frames)

	// 回调输出
	if e.onOutput != nil && len(outputs) > 0 {
		e.onOutput(roomID, outputs)
	}
}

// parseTrackFrames 从 iot:monitor:stream 的 data_value 解析 track 坐标。
func (e *Engine) parseTrackFrames(raw map[string]interface{}, deviceID string) []TrackFrame {
	dv := raw[rediscommon.DataValueKey]
	if dv == nil {
		// 尝试直接从顶层字段解析（兼容两种格式）
		dv = raw["dataValue"]
	}
	if dv == nil {
		return nil
	}

	var items []map[string]interface{}
	switch v := dv.(type) {
	case []interface{}:
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				items = append(items, m)
			}
		}
	case map[string]interface{}:
		items = append(items, v)
	default:
		return nil
	}

	var frames []TrackFrame
	for _, item := range items {
		trackID := intFromAny(item["track_id"])
		if trackID < 0 || trackID > 8 {
			continue // 只要有效人员 track
		}
		x := floatFromAny(item["position_x"])
		y := floatFromAny(item["position_y"])
		z := floatFromAny(item["position_z"])
		pose := intFromAny(item["pose"])
		areaType := intFromAny(item["area_type"])
		ts := int64FromAny(item["timestamp"])
		if ts == 0 {
			ts = time.Now().UnixMilli()
		}

		frames = append(frames, TrackFrame{
			TrackID:  trackID,
			DeviceID: deviceID,
			X:        x, Y: y, Z: z,
			Pose:     pose,
			AreaType: areaType,
			TMs:      ts,
		})
	}
	return frames
}

// decayLoop 定期衰减所有房间网格。
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
				grid.DecayAll(dtSec, e.decayHalfLife)
			}
			e.mu.RUnlock()
		}
	}
}

// GetRoomOutputs 查询某房间的最新 track 评估结果。
func (e *Engine) GetRoomOutputs(roomID string) []TrackOutput {
	e.mu.RLock()
	tm := e.rooms[roomID]
	e.mu.RUnlock()
	if tm == nil {
		return nil
	}
	return tm.GetOutputs()
}

// ======================== 辅助 ========================

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

func floatFromAny(v interface{}) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case json.Number:
		f, _ := x.Float64()
		return f
	case string:
		f, _ := strconv.ParseFloat(x, 64)
		return f
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
