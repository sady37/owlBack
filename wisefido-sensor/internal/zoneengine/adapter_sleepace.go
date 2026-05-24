package zoneengine

// adapter_sleepace.go — sleepace 床垫 → zoneengine SignalEvidence 翻译层。
//
// 订阅 iot:event:stream（独立 consumer group "wisefido-zoneengine-sleepace"），
// 把 sleepace 日常 InBed / LeftBed 翻成 SignalEvidence 喂 Engine。
// SignalEvidence.Source 固定 "sleepace"，方向：
//
//	InBed     → kind="enter"
//	LeftBed   → kind="leave"
//
// 2026-05-20 修订：从 iot:alarm:stream 改为 iot:event:stream。
// 历史 bug：sleepace mqtt_consumer 把 inBedStatus / bedStatus change 发到 event 流，
// 厂家原生 alarmInBed/alarmLeftBed 才发 alarm 流。adapter 只订 alarm 流时，日常上下床事件
// 不进 zone engine，bed 状态全靠 VitalAdapter HR/RR 心跳间接推动，sleepad 离床后送 0 值
// vital 会让 zone 卡 Occupied → display.bed_status 永久不翻 1。
//
// **不消费 BedSitUp**（设计决策 2026-05-14）：BedSitUp 是行为分类 alarm（"坐起"），
// 不是 presence 信号；让它驱动 zone state 会污染"在/不在床"语义。BedSitUp 留作
// cardagg / 其他业务路径（坐起报警 / 离床预警 / sleep stage 等）独立判断。
//
// 设计约束：
//   - 仅 device_type 含 "sleepad"/"sleeppad"（与 cardagg alarm_handler 同判定）。
//   - SubjectEntity 空（未绑卡）跳过。
//   - event_status="end" 跳过：sleepace 用 end 表示 alarm 条件解除，与 zone state 翻转语义无关
//     （状态翻转应由 LeftBed 这类显式 leave 信号驱动）。
//   - 30s inbound age 守一致。

import (
	"context"
	"strings"
	"time"

	"owl-common/alarm"
	"owl-common/observation"
	rediscommon "owl-common/redis"

	redislib "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// SleepaceAdapter 订阅 iot:alarm:stream 并把 sleepace 床垫 alarm 翻译为 SignalEvidence 喂 engine。
type SleepaceAdapter struct {
	client   *redislib.Client
	engine   *Engine
	dedup    *BedEventDedup       // S5b: per-device per-kind 10s dedup（防 firmware spam）
	fitness  DeviceFitnessChecker // S6: 设备类 alarm gate
	presence BedPresenceSink      // RoomState dedup 旁路：sleepad InBed/LeftBed → X
	logger   *zap.Logger
}

const (
	sleepaceConsumerGroup   = "wisefido-zoneengine-sleepace"
	sleepaceConsumerName    = "consumer-zoneengine-sleepace"
	sleepaceReadCount       = 16
	sleepaceReadBlock       = 5 * time.Second
	sleepaceMaxInboundAgeMs = 30_000
)

func NewSleepaceAdapter(client *redislib.Client, engine *Engine, logger *zap.Logger) *SleepaceAdapter {
	return &SleepaceAdapter{client: client, engine: engine, dedup: NewBedEventDedup(), logger: logger}
}

// SetFitnessChecker main wiring 注入；不调时退化为"所有 device 都 fit"。
func (a *SleepaceAdapter) SetFitnessChecker(c DeviceFitnessChecker) {
	if a == nil {
		return
	}
	a.fitness = c
}

// SetBedPresence main wiring 注入；不调时 dedup 旁路 disable。
func (a *SleepaceAdapter) SetBedPresence(s BedPresenceSink) {
	if a == nil {
		return
	}
	a.presence = s
}

func (a *SleepaceAdapter) Start(ctx context.Context) {
	if err := rediscommon.CreateConsumerGroup(ctx, a.client, rediscommon.StreamEvent.Name, sleepaceConsumerGroup); err != nil {
		a.logger.Warn("zoneengine sleepace adapter: create consumer group", zap.Error(err))
	}
	go a.runLoop(ctx)
	a.logger.Info("zoneengine sleepace adapter started",
		zap.String("stream", rediscommon.StreamEvent.Name),
		zap.String("group", sleepaceConsumerGroup),
	)
}

func (a *SleepaceAdapter) runLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		msgs, err := rediscommon.ReadFromStreamWithBlock(ctx, a.client,
			rediscommon.StreamEvent.Name, sleepaceConsumerGroup, sleepaceConsumerName,
			sleepaceReadCount, sleepaceReadBlock)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			a.logger.Warn("zoneengine sleepace adapter: read stream", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}
		for _, m := range msgs {
			a.handleRaw(m.Values)
			a.client.XAck(ctx, rediscommon.StreamEvent.Name, sleepaceConsumerGroup, m.ID)
		}
	}
}

func (a *SleepaceAdapter) handleRaw(raw map[string]interface{}) {
	msg, err := rediscommon.FromStreamMap(raw)
	if err != nil {
		return
	}
	a.handleMsg(msg)
}

func (a *SleepaceAdapter) handleMsg(msg *rediscommon.IoTStreamMessage) {
	if msg == nil || !msg.DeviceAddr.IsValid() {
		return
	}
	// S6 fitness gate：unfit sleepace（SensorDetached 是 P0 风险——detached 后床垫会乱发数据）
	// 不喂 engine 防 bed FSM 被假数据污染。详 service.DeviceFitnessTracker。
	if a.fitness != nil && !a.fitness.IsFit(msg.DeviceAddr.String()) {
		return
	}
	if msg.SubjectEntity == "" {
		return
	}
	dt := strings.ToLower(msg.DeviceType)
	if !strings.Contains(dt, "sleepad") && !strings.Contains(dt, "sleeppad") {
		return
	}
	nowMs := time.Now().UnixMilli()
	if nowMs-msg.Timestamp > sleepaceMaxInboundAgeMs {
		return
	}
	fields := rediscommon.FirstDataValue(msg.DataValue)
	if fields == nil {
		fields = map[string]interface{}{}
	}
	// event_status="end" 跳过 —— 状态翻转由显式 leave 信号驱动，不消费 alarm 解除事件。
	if status, _ := fields[observation.FieldEventStatus].(string); status == "end" {
		return
	}

	bedPref := prefixOf(msg.DeviceAddr, 96)
	if bedPref == "" {
		return
	}

	var kind string
	switch msg.Category {
	case alarm.InBed:
		kind = "enter"
	case alarm.LeftBed:
		kind = "leave"
	case alarm.SleepStage:
		// sleep-stage event = sleepad 仍在测床上人（sleepad 设计：只有压感持续才发 sleep_stage）。
		// 当 sleep_stage ∈ {1,2,4} (awake/light/deep) 时发 sustain SignalEvidence，
		// 让 scorer.LastEvidenceTs 持续刷新，防止 repairSubsetInvariant 因为 InBed event 长时间
		// 不来（sleepad 只在状态翻转时发 InBed/LeftBed）而把活生生在床的人 force Vacant。
		// sleep_stage=0 (Initial) / 8 (Unknown) 跳过 — 不算有效在床信号。
		stage := intFromAny(fields[observation.FieldSleepStage])
		if stage != 1 && stage != 2 && stage != 4 {
			return
		}
		a.engine.Apply(SignalEvidence{
			ZoneType:    ZoneTypeBed,
			ZoneID:      bedPref,
			Source:      "sleepace",
			Kind:        "sustain",
			Ts:          msg.Timestamp,
			TriggerData: fields,
		})
		return
	default:
		// BedSitUp / 其他 sleepace alarm 不入 zone engine —— presence 信号通道纯净。
		return
	}

	// S5b 10s 同类 dedup：防 firmware spam（同一 device 短窗内重发 InBed/LeftBed）
	if !a.dedup.Admit(msg.DeviceAddr.String(), kind, msg.Timestamp) {
		return
	}

	a.engine.Apply(SignalEvidence{
		ZoneType:    ZoneTypeBed,
		ZoneID:      bedPref,
		Source:      "sleepace",
		Kind:        kind,
		Ts:          msg.Timestamp,
		TriggerData: fields,
	})
	if a.presence != nil {
		a.presence.SetSleepad(bedPref, kind == "enter", msg.Timestamp)
	}
}
