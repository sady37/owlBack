package consumer

// 订阅的流与消费者（cardagg-group）：
//
//	iot:monitor:stream         → consumer-monitor         → MonitorHandler
//	iot:event:stream           → consumer-event           → EventHandler
//	iot:alarm:stream           → consumer-alarm            → AlarmHandler
//	config:alarmProcess:stream → consumer-alarm-process    → AlarmProcessHandler
//	config:card:stream         → consumer-card-change      → CardChangeHandler
//	config:alarmDevice:stream  → consumer-alarm-device     → AlarmDeviceHandler

import (
	"context"
	"time"

	rediscommon "owl-common/redis"

	redislib "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// StreamHandler 处理单条 stream 消息
type StreamHandler interface {
	Handle(ctx context.Context, msg interface{}) error
}

// Handlers 各流对应的 handler，由 main 构造后传入
type Handlers struct {
	Monitor       StreamHandler // iot:monitor:stream
	Event         StreamHandler // iot:event:stream
	Alarm         StreamHandler // iot:alarm:stream
	AlarmProcess  StreamHandler // config:alarmProcess:stream
	CardChange    StreamHandler // config:card:stream
	AlarmDevice   StreamHandler // config:alarmDevice:stream
}

const consumerGroup = "$cons-cardagg" // 与 state_service.go 一致

const (
	readCount = 10
	readBlock = 5 * time.Second
)

const (
	drainTimeout  = 10 * time.Second
	drainMaxMsgs  = 500
	drainBatch    = 50
)

// SubscribeAll 在本包内直接订阅上述 5 个流，收到后调用 h 中对应 Handler。
func SubscribeAll(ctx context.Context, logger *zap.Logger, client *redislib.Client, h Handlers) {
	subs := []struct {
		stream   string
		consumer string
		handler  StreamHandler
	}{
		{rediscommon.StreamMonitor.Name, "consumer-monitor", h.Monitor},
		{rediscommon.StreamEvent.Name, "consumer-event", h.Event},
		{rediscommon.StreamAlarm.Name, "consumer-alarm", h.Alarm},
		{rediscommon.StreamConfigAlarmProcess.Name, "consumer-alarm-process", h.AlarmProcess},
		{rediscommon.StreamConfigCard.Name, "consumer-card-change", h.CardChange},
		{rediscommon.StreamConfigAlarmDevice.Name, "consumer-alarm-device", h.AlarmDevice},
	}
	for _, s := range subs {
		if err := rediscommon.CreateConsumerGroup(ctx, client, s.stream, consumerGroup); err != nil {
			logger.Warn("create consumer group", zap.String("stream", s.stream), zap.Error(err))
		}
	}
	for _, s := range subs {
		stream, consumer, handler := s.stream, s.consumer, s.handler
		go runSubscriber(ctx, logger, client, consumerGroup, stream, consumer, handler)
	}
}

func runSubscriber(ctx context.Context, logger *zap.Logger, client *redislib.Client, group, stream, consumer string, handler StreamHandler) {
	logger.Info("subscriber started", zap.String("stream", stream))
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		msgs, err := rediscommon.ReadFromStreamWithBlock(ctx, client, stream, group, consumer, readCount, readBlock)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			logger.Warn("read stream", zap.String("stream", stream), zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		for _, m := range msgs {
			if err := handler.Handle(ctx, m.Values); err != nil {
				logger.Warn("handle error", zap.String("stream", stream), zap.Error(err))
			}
			client.XAck(ctx, stream, group, m.ID)
		}
	}
}

// ────────────────────────────────────────────────────────────────────────────
// DrainConfigCardStream
//
// 在 IoT 流（iot:alarm / iot:event / iot:monitor）启动前，同步处理完
// config:card:stream 的全部积压，确保 DeviceCardResolver 与 metaCache
// 预热完毕，避免 IotPreparedHandler 将 device_id 误用为 card_id。
//
// 处理顺序：
//   Phase 1 — PEL：认领上次崩溃/重启遗留的 pending 消息（XPENDING + XCLAIM）
//   Phase 2 — NEW：读取 ">" 新消息直到队列为空（XREADGROUP Block 200ms 超时退出）
//
// 安全上限：
//   drainTimeout  = 10s（全局超时，超时后继续正常启动，打 Warn 日志）
//   maxMessages   = 500（单次启动最多处理条数，超出打 Warn 提示运维）
// ────────────────────────────────────────────────────────────────────────────

func DrainConfigCardStream(
	ctx context.Context,
	logger *zap.Logger,
	client *redislib.Client,
	handler StreamHandler,
) {
	stream := rediscommon.StreamConfigCard.Name
	consumer := "consumer-card-change" // 与 SubscribeAll 常驻订阅同名，共享 ACK offset

	// 确保 consumer group 存在；CreateConsumerGroup 用 "0"，新组能见到所有历史消息
	if err := rediscommon.CreateConsumerGroup(ctx, client, stream, consumerGroup); err != nil {
		logger.Warn("drain: ensure group", zap.String("stream", stream), zap.Error(err))
	}

	drainCtx, cancel := context.WithTimeout(ctx, drainTimeout)
	defer cancel()

	total := 0
	hitLimit := false

	// ── Phase 1: PEL（上次实例未 ACK 的 pending 消息）────────────────────────
	total, hitLimit = drainPendingMessages(drainCtx, logger, client, stream, consumer, handler, total)
	if hitLimit {
		logger.Warn("drain: hit max limit in PEL phase, skipping NEW phase",
			zap.Int("processed", total), zap.Int("max", drainMaxMsgs))
		logger.Info("drain: done (limit)", zap.Int("messages", total))
		return
	}

	// ── Phase 2: NEW（">" 尚未投递的消息）────────────────────────────────────
	for total < drainMaxMsgs {
		if drainCtx.Err() != nil {
			logger.Warn("drain: timeout in NEW phase",
				zap.Int("processed", total), zap.Duration("limit", drainTimeout))
			break
		}

		result, err := client.XReadGroup(drainCtx, &redislib.XReadGroupArgs{
			Group:    consumerGroup,
			Consumer: consumer,
			Streams:  []string{stream, ">"},
			Count:    drainBatch,
			Block:    200 * time.Millisecond,
		}).Result()

		if err != nil {
			// redis.Nil = block 超时无新消息，正常退出
			break
		}

		empty := true
		for _, s := range result {
			for _, m := range s.Messages {
				empty = false
				if herr := handler.Handle(drainCtx, m.Values); herr != nil {
					logger.Warn("drain: handle error", zap.String("id", m.ID), zap.Error(herr))
				}
				client.XAck(drainCtx, stream, consumerGroup, m.ID) //nolint:errcheck
				total++
				if total >= drainMaxMsgs {
					logger.Warn("drain: hit max message limit",
						zap.Int("max", drainMaxMsgs),
						zap.String("hint", "config:card:stream backlog unusually large, check data service"))
					hitLimit = true
					break
				}
			}
			if hitLimit {
				break
			}
		}
		if empty || hitLimit {
			break
		}
	}

	logger.Info("drain: config:card:stream done", zap.Int("messages", total))
}

// drainPendingMessages 处理 PEL 中遗留的 pending 消息（XPENDING + XCLAIM）。
// 返回累计处理数和是否触碰上限。
func drainPendingMessages(
	ctx context.Context,
	logger *zap.Logger,
	client *redislib.Client,
	stream, consumer string,
	handler StreamHandler,
	startCount int,
) (total int, hitLimit bool) {
	total = startCount

	// 先用 XPENDING summary 快速判断是否有 pending，避免无谓的 XCLAIM 调用
	summary, err := client.XPending(ctx, stream, consumerGroup).Result()
	if err != nil || summary.Count == 0 {
		return total, false
	}

	logger.Info("drain: found pending messages", zap.Int64("count", summary.Count))

	for total < drainMaxMsgs {
		if ctx.Err() != nil {
			logger.Warn("drain: timeout in PEL phase", zap.Int("processed", total))
			break
		}

		// XPENDING 拿到 pending 条目的 ID 列表（不含消息体）
		pendingEntries, err := client.XPendingExt(ctx, &redislib.XPendingExtArgs{
			Stream: stream,
			Group:  consumerGroup,
			Start:  "-",
			End:    "+",
			Count:  int64(drainBatch),
		}).Result()
		if err != nil || len(pendingEntries) == 0 {
			break
		}

		ids := make([]string, 0, len(pendingEntries))
		for _, p := range pendingEntries {
			ids = append(ids, p.ID)
		}

		// XCLAIM 认领并取得消息体（MinIdle:0 = 不限空闲时间，无条件认领）
		claimed, err := client.XClaim(ctx, &redislib.XClaimArgs{
			Stream:   stream,
			Group:    consumerGroup,
			Consumer: consumer,
			MinIdle:  0,
			Messages: ids,
		}).Result()
		if err != nil {
			logger.Warn("drain: xclaim error", zap.Error(err))
			break
		}

		for _, m := range claimed {
			if herr := handler.Handle(ctx, m.Values); herr != nil {
				logger.Warn("drain: handle pending error", zap.String("id", m.ID), zap.Error(herr))
			}
			client.XAck(ctx, stream, consumerGroup, m.ID) //nolint:errcheck
			total++
			if total >= drainMaxMsgs {
				return total, true
			}
		}

		// 本批全部处理完；若 pending 条目比 drainBatch 少，说明 PEL 已清空
		if len(pendingEntries) < drainBatch {
			break
		}
	}

	return total, false
}
