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

const consumerGroup = "cardagg-group"
const readBlock = 2 * time.Second
const readCount = 10

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
