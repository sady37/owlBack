package consumer

// S4 (FU1): sensor-side SleepStage event consumer + confidence ladder.
//
// 订阅 iot:event:stream 独立 consumer group "wisefido-sensor-sleepstage"，filter
// category=alarm.SleepStage（sleepace + radar 双源 publisher）。
//
// **不是加权融合**（v1 实际也未实现 100=双设备）；是 **confidence ladder 覆盖**
// （详 [[cardagg_v1_to_v2_migration_audit]]）：
//   - Sleepad → confidence 90
//   - Radar   → confidence 60
//   - 其它    → confidence 0（丢）
//   - 仅 incoming.confidence ≥ current.confidence 才覆盖；否则 drop
//
// per-card in-memory state（cardID → (sleepStage, confidence, lastTsMs)）；sensor 重启
// 后丢失，下一次 sleepace/radar push 自动回填（频率 ≤1 分钟）。
//
// 命中 ladder → 通过 sink.PublishBedSleepStage 写 sensor:derived:stream
// category=bed.sleepstage（projector 字段级 merge 仅更 SleepStage / SleepConfidence /
// UpdatedAt，不动 BedStatus 等 zoneengine owner 字段）。
//
// 防 loop：与 alarm/activity 同规则，自家 producer 跳过。
//
// **deferred to follow-up PR（S4 范围之外）**：
//   - OOB drop + device_failure（要 zoneengine bed FSM 暴露 query 接口）
//   - LeftBed/ExitRoom/EnterRoom 触发自动清零（要 zoneengine listener wire）

import (
	"context"
	"net/netip"
	"strings"
	"sync"
	"time"

	"owl-common/alarm"
	"owl-common/observation"
	rediscommon "owl-common/redis"
	"owl-common/spatial"

	redislib "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const (
	sensorSleepStageGroup    = "wisefido-sensor-sleepstage"
	sensorSleepStageConsumer = "consumer-sensor-sleepstage"

	sleepConfidenceSleepad = 90
	sleepConfidenceRadar   = 60
)

// SleepStageSink 解耦 zoneengine 反向 import：StreamPublisher.PublishBedSleepStage 隐式满足。
type SleepStageSink interface {
	PublishBedSleepStage(ctx context.Context, cardID string, sleepStage, sleepConfidence int) error
}

// BedOccupancyChecker (C OOB 守卫): zoneengine.Engine 隐式满足；nil 注入 = 不做 OOB 检查。
type BedOccupancyChecker interface {
	IsBedOccupied(spatialPrefix string) bool
}

// DeviceFailureEmitter (C device_failure publish): AlarmBackChannel.PublishAlarmFire 隐式满足；
// nil 注入 = 不发 device_failure alarm（只 drop event，不上报）。primitive 参数避免反向 import。
type DeviceFailureEmitter interface {
	PublishAlarmFire(ctx context.Context, deviceAddr netip.Addr, subjectEntity, eventName, level string,
		tsMs int64, triggerData map[string]interface{}) (string, error)
}

type SleepStageConsumer struct {
	client     *redislib.Client
	sink       SleepStageSink
	bedChecker BedOccupancyChecker  // C: OOB 守卫；nil = disable
	failEmit   DeviceFailureEmitter // C: device_failure alarm publish；nil = disable
	sensorSlot netip.Prefix
	logger     *zap.Logger

	mu              sync.Mutex
	state           map[string]sleepStageEntry // cardID → 当前 ladder 态
	failDedupTs     map[string]int64           // C: spatialPrefix → 最近 device_failure emit tsMs（5min dedup 防 spam）
}

const deviceFailureDedupMs int64 = 5 * 60 * 1000 // 5min

type sleepStageEntry struct {
	sleepStage int
	confidence int
	tsMs       int64
}

func NewSleepStageConsumer(client *redislib.Client, sink SleepStageSink, logger *zap.Logger) *SleepStageConsumer {
	slot, _ := netip.ParsePrefix(spatial.SlotSensor)
	return &SleepStageConsumer{
		client:      client,
		sink:        sink,
		sensorSlot:  slot,
		logger:      logger,
		state:       make(map[string]sleepStageEntry),
		failDedupTs: make(map[string]int64),
	}
}

// SetBedChecker (C) main wire 注入 bed FSM occupancy 查询；nil = 跳过 OOB 守卫（保持兼容）。
func (c *SleepStageConsumer) SetBedChecker(b BedOccupancyChecker) {
	if c == nil {
		return
	}
	c.bedChecker = b
}

// SetDeviceFailureEmitter (C) main wire 注入 alarm publisher；nil = OOB 时仅 drop 不报警。
func (c *SleepStageConsumer) SetDeviceFailureEmitter(e DeviceFailureEmitter) {
	if c == nil {
		return
	}
	c.failEmit = e
}

func (c *SleepStageConsumer) Start(ctx context.Context) {
	if err := rediscommon.CreateConsumerGroup(ctx, c.client, rediscommon.StreamEvent.Name, sensorSleepStageGroup); err != nil {
		c.logger.Warn("sensor sleepstage: create consumer group", zap.Error(err))
	}
	go c.runLoop(ctx)
	c.logger.Info("sensor sleepstage consumer started",
		zap.String("stream", rediscommon.StreamEvent.Name),
		zap.String("group", sensorSleepStageGroup))
}

func (c *SleepStageConsumer) runLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		msgs, err := rediscommon.ReadFromStreamWithBlock(ctx, c.client,
			rediscommon.StreamEvent.Name, sensorSleepStageGroup, sensorSleepStageConsumer,
			sensorReadCount, sensorReadBlock)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.logger.Warn("sensor sleepstage: read stream", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}
		for _, m := range msgs {
			c.handleRaw(ctx, m.Values)
			c.client.XAck(ctx, rediscommon.StreamEvent.Name, sensorSleepStageGroup, m.ID)
		}
	}
}

func (c *SleepStageConsumer) handleRaw(ctx context.Context, raw map[string]interface{}) {
	msg, err := rediscommon.FromStreamMap(raw)
	if err != nil {
		c.logger.Warn("sensor sleepstage: parse", zap.Error(err))
		return
	}
	if c.isSelfProducer(msg.Producer) {
		return
	}
	if msg.Category != alarm.SleepStage {
		return
	}
	if msg.SubjectEntity == "" {
		return
	}
	fields := rediscommon.FirstDataValue(msg.DataValue)
	if fields == nil {
		return
	}
	sleepStage := intFromSleepStageField(fields[observation.FieldSleepStage])
	if sleepStage <= 0 {
		return // 无效 / 未上报
	}
	confidence := confidenceFromDeviceType(msg.DeviceType)
	if confidence == 0 {
		return // 非 sleepad / radar 来源忽略
	}
	// 边界派生：raw event envelope.SubjectEntity 装 device_uid (CE 主体=设备)；sensor 内部一切
	// 按 /96 床 spatial prefix 索引（zone FSM 查询、ladder state、出口 publish subject）。device_addr
	// 不效 → 不能派生 → 丢弃（防御性，理论上 FromStreamMap 已 validate）。
	if !msg.DeviceAddr.IsValid() {
		return
	}
	bedSpatialPrefix := netip.PrefixFrom(msg.DeviceAddr, 96).Masked().String()

	// C OOB 守卫：sensor 自家 bed FSM 已离床（Vacant）→ sleepace 报 SleepStage 是异常
	// （设备脱床 / 接触不良 / 电磁干扰）。drop event + emit device_failure alarm（5min dedup
	// 防 spam）让运维主动定位。bedChecker nil 或 Engine 无该 prefix entry → 视作 Occupied 放行。
	if c.bedChecker != nil && !c.bedChecker.IsBedOccupied(bedSpatialPrefix) {
		c.emitOOBDeviceFailure(ctx, msg, bedSpatialPrefix, sleepStage)
		return
	}
	if !c.ladderAdmits(bedSpatialPrefix, sleepStage, confidence, msg.Timestamp) {
		return
	}
	if err := c.sink.PublishBedSleepStage(ctx, bedSpatialPrefix, sleepStage, confidence); err != nil {
		c.logger.Warn("sensor sleepstage: publish failed",
			zap.String("spatial_prefix", bedSpatialPrefix),
			zap.Int("sleep_stage", sleepStage),
			zap.Int("confidence", confidence),
			zap.Error(err))
		// publish 失败：回退本地 state 让下次重试（仍是当前 ladder 等待更高 confidence）
		c.rollbackEntry(bedSpatialPrefix, sleepStage, confidence)
	}
}

// OnDeviceUnfit (S6+): device offline/SensorDetached 时 DeviceFitnessTracker broadcast 调用。
// 把 device /96 当作 spatial prefix 调 OnBedVacant 等价处理（清 ladder state + emit (0,0)）。
// "device offline = 内存重启"原则：未 fire 的 cache 清掉，下次 device 上线从干净起点累积。
func (c *SleepStageConsumer) OnDeviceUnfit(deviceAddr string) {
	if c == nil || deviceAddr == "" {
		return
	}
	addr, err := netip.ParseAddr(deviceAddr)
	if err != nil {
		return
	}
	sp := netip.PrefixFrom(addr, 96).Masked().String()
	c.OnBedVacant(context.Background(), sp, time.Now().UnixMilli())
}

// OnBedVacant (D): bed FSM transition Leaving/Vacant 时 wiring adapter 调本方法。
//
// 触发清零：
//   1. 本地 ladder state 清掉（让下次 SleepStage event 重新走 ladder，不被旧 confidence 卡）
//   2. publish bed.sleepstage(0, 0) 让 cardagg 端 BedState.SleepStage / SleepConfidence 归零
//      （避免 sleepace 关机 / 老人起床后 FE Section2.left 仍显示 "Deep Sleep" 等 stale 值）
//
// spatialPrefix = ZoneEvent.ZoneID（/96 bed CIDR）。
func (c *SleepStageConsumer) OnBedVacant(ctx context.Context, spatialPrefix string, tsMs int64) {
	if c == nil || spatialPrefix == "" {
		return
	}
	c.mu.Lock()
	_, had := c.state[spatialPrefix]
	delete(c.state, spatialPrefix)
	c.mu.Unlock()
	if !had {
		return // 没在床过，no-op
	}
	c.logger.Info("sleepstage cleared on bed vacant",
		zap.String("spatial_prefix", spatialPrefix),
		zap.Int64("ts_ms", tsMs))
	if c.sink == nil {
		return
	}
	if err := c.sink.PublishBedSleepStage(ctx, spatialPrefix, 0, 0); err != nil {
		c.logger.Warn("sleepstage clear: publish bed.sleepstage(0,0) failed",
			zap.String("spatial_prefix", spatialPrefix),
			zap.Error(err))
	}
}

// emitOOBDeviceFailure (C): sleepace 报 SleepStage 但 bed FSM 已离床 → drop event +
// publish device_failure alarm（5min dedup 防同设备 spam）。
//
// alarm 由 cardagg AlarmRouter 派发持久化；sensor 通过 AlarmBackChannel 走 iot:alarm:stream。
// 若 failEmit 未注入 → 仅 log 不报警（保持兼容；运维 log 仍可见 OOB drop 事件）。
func (c *SleepStageConsumer) emitOOBDeviceFailure(ctx context.Context, msg *rediscommon.IoTStreamMessage, bedSpatialPrefix string, sleepStage int) {
	c.logger.Warn("sleepstage OOB drop",
		zap.String("device_addr", msg.DeviceAddr.String()),
		zap.String("subject_entity", msg.SubjectEntity),
		zap.String("spatial_prefix", bedSpatialPrefix),
		zap.String("device_type", msg.DeviceType),
		zap.Int("sleep_stage", sleepStage),
		zap.String("reason", "bed FSM Vacant but sleepStage event received — likely SensorDetached / 干扰 / 设备故障"),
	)

	if c.failEmit == nil || !msg.DeviceAddr.IsValid() {
		return
	}
	// 5min dedup 按 device_addr：device_failure 是 per-device 现象（哪个 sleepad/radar 数据异常），
	// 同一设备的 spam 才是要抑制的目标；不按床合并（同床上 sleepad+radar 同时 OOB 时各自独立报）。
	deviceKey := msg.DeviceAddr.String()
	c.mu.Lock()
	last := c.failDedupTs[deviceKey]
	if msg.Timestamp-last < deviceFailureDedupMs {
		c.mu.Unlock()
		return
	}
	c.failDedupTs[deviceKey] = msg.Timestamp
	c.mu.Unlock()

	triggerData := map[string]interface{}{
		"reason":         "sleepstage_out_of_bed",
		"sleep_stage":    sleepStage,
		"device_type":    msg.DeviceType,
		"subject_entity": msg.SubjectEntity,
	}
	// alarm 出口 subject = msg.SubjectEntity (device_uid)：device_failure 是 alarm 类，按规约 subject =
	// device_uid（cardagg alarm_router 走 device_addr LPM 找归属卡，subject 提供 device 身份标识）。
	if _, err := c.failEmit.PublishAlarmFire(ctx, msg.DeviceAddr, msg.SubjectEntity,
		alarm.AlarmTypeDeviceFailure, alarm.AlarmLevelErr, msg.Timestamp, triggerData); err != nil {
		c.logger.Warn("sleepstage OOB: publish device_failure failed",
			zap.String("device_addr", msg.DeviceAddr.String()),
			zap.Error(err))
	}
}

// ladderAdmits confidence ladder 判定 + 更新本地 state（仅 admits 时 update）。
func (c *SleepStageConsumer) ladderAdmits(cardID string, sleepStage, confidence int, tsMs int64) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	cur, ok := c.state[cardID]
	if ok && confidence < cur.confidence {
		return false
	}
	c.state[cardID] = sleepStageEntry{sleepStage: sleepStage, confidence: confidence, tsMs: tsMs}
	return true
}

// rollbackEntry publish 失败时把 state 还原到 prev（如果有），无 prev 直接删；
// 下次同源事件来时再触发 publish。
func (c *SleepStageConsumer) rollbackEntry(cardID string, sleepStage, confidence int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	cur, ok := c.state[cardID]
	if !ok {
		return
	}
	if cur.sleepStage == sleepStage && cur.confidence == confidence {
		delete(c.state, cardID)
	}
}

func (c *SleepStageConsumer) isSelfProducer(producer string) bool {
	if producer == "" {
		return false
	}
	if producer == "wisefido-sensor" || strings.HasPrefix(producer, "sensor.") {
		return true
	}
	addr, err := netip.ParseAddr(producer)
	if err != nil {
		return false
	}
	return c.sensorSlot.Contains(addr)
}

func confidenceFromDeviceType(deviceType string) int {
	switch strings.ToLower(deviceType) {
	case "sleepad", "sleeppad":
		return sleepConfidenceSleepad
	case "radar":
		return sleepConfidenceRadar
	}
	return 0
}

func intFromSleepStageField(v interface{}) int {
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	}
	return 0
}
