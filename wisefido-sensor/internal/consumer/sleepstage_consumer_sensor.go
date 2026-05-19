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

type SleepStageConsumer struct {
	client     *redislib.Client
	sink       SleepStageSink
	sensorSlot netip.Prefix
	logger     *zap.Logger

	mu    sync.Mutex
	state map[string]sleepStageEntry // cardID → 当前 ladder 态
}

type sleepStageEntry struct {
	sleepStage int
	confidence int
	tsMs       int64
}

func NewSleepStageConsumer(client *redislib.Client, sink SleepStageSink, logger *zap.Logger) *SleepStageConsumer {
	slot, _ := netip.ParsePrefix(spatial.SlotSensor)
	return &SleepStageConsumer{
		client:     client,
		sink:       sink,
		sensorSlot: slot,
		logger:     logger,
		state:      make(map[string]sleepStageEntry),
	}
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
	if !c.ladderAdmits(msg.SubjectEntity, sleepStage, confidence, msg.Timestamp) {
		return
	}
	if err := c.sink.PublishBedSleepStage(ctx, msg.SubjectEntity, sleepStage, confidence); err != nil {
		c.logger.Warn("sensor sleepstage: publish failed",
			zap.String("card_id", msg.SubjectEntity),
			zap.Int("sleep_stage", sleepStage),
			zap.Int("confidence", confidence),
			zap.Error(err))
		// publish 失败：回退本地 state 让下次重试（仍是当前 ladder 等待更高 confidence）
		c.rollbackEntry(msg.SubjectEntity, sleepStage, confidence)
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
