// stream_subscriber.go — 6 流订阅，按消息形态分两类 handler：
//
//   iot:monitor / iot:alarm / sensor:derived              → IoTHandler（解析后传 IoTStreamMessage）
//   config:card / config:alarmDevice / config:alarmProcess → RawHandler（CloudEvents envelope，传 raw map）

package consumer

import (
	"context"
	"time"

	owlredis "owl-common/redis"

	redislib "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type IoTHandler interface {
	Handle(ctx context.Context, msg *owlredis.IoTStreamMessage) error
}

type RawHandler interface {
	Handle(ctx context.Context, raw map[string]interface{}) error
}

type Handlers struct {
	Monitor      IoTHandler // iot:monitor:stream
	Alarm        IoTHandler // iot:alarm:stream
	SensorState  IoTHandler // sensor:state:stream
	CardChange   RawHandler // config:card:stream
	AlarmDevice  RawHandler // config:alarmDevice:stream
	AlarmProcess RawHandler // config:alarmProcess:stream
}

const (
	consumerGroup = "$cons-cardagg"
	readCount     = 10
	readBlock     = 5 * time.Second
)

func SubscribeAll(ctx context.Context, logger *zap.Logger, client *redislib.Client, h Handlers) {
	iotSubs := []struct {
		stream, consumer string
		handler          IoTHandler
	}{
		{owlredis.StreamMonitor.Name, "consumer-monitor", h.Monitor},
		{owlredis.StreamAlarm.Name, "consumer-alarm", h.Alarm},
		{owlredis.StreamSensorDerived.Name, "consumer-sensor-derived", h.SensorState},
	}
	rawSubs := []struct {
		stream, consumer string
		handler          RawHandler
	}{
		{owlredis.StreamConfigCard.Name, "consumer-card-change", h.CardChange},
		{owlredis.StreamConfigAlarmDevice.Name, "consumer-alarm-device", h.AlarmDevice},
		{owlredis.StreamConfigAlarmProcess.Name, "consumer-alarm-process", h.AlarmProcess},
	}
	for _, s := range iotSubs {
		if err := owlredis.CreateConsumerGroup(ctx, client, s.stream, consumerGroup); err != nil {
			logger.Warn("create consumer group", zap.String("stream", s.stream), zap.Error(err))
		}
	}
	for _, s := range rawSubs {
		if err := owlredis.CreateConsumerGroup(ctx, client, s.stream, consumerGroup); err != nil {
			logger.Warn("create consumer group", zap.String("stream", s.stream), zap.Error(err))
		}
	}
	for _, s := range iotSubs {
		stream, consumer, handler := s.stream, s.consumer, s.handler
		go runIoTSubscriber(ctx, logger, client, stream, consumer, handler)
	}
	for _, s := range rawSubs {
		stream, consumer, handler := s.stream, s.consumer, s.handler
		go runRawSubscriber(ctx, logger, client, stream, consumer, handler)
	}
}

func runIoTSubscriber(ctx context.Context, logger *zap.Logger, client *redislib.Client, stream, consumer string, handler IoTHandler) {
	logger.Info("iot subscriber started", zap.String("stream", stream))
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		msgs, err := owlredis.ReadFromStreamWithBlock(ctx, client, stream, consumerGroup, consumer, readCount, readBlock)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			logger.Warn("read iot stream", zap.String("stream", stream), zap.Error(err))
			time.Sleep(time.Second)
			continue
		}
		for _, m := range msgs {
			parsed, err := owlredis.FromStreamMap(m.Values)
			if err != nil {
				logger.Warn("parse iot msg", zap.String("stream", stream), zap.Error(err))
				client.XAck(ctx, stream, consumerGroup, m.ID)
				continue
			}
			if err := handler.Handle(ctx, parsed); err != nil {
				logger.Warn("handle iot msg", zap.String("stream", stream), zap.Error(err))
			}
			client.XAck(ctx, stream, consumerGroup, m.ID)
		}
	}
}

func runRawSubscriber(ctx context.Context, logger *zap.Logger, client *redislib.Client, stream, consumer string, handler RawHandler) {
	logger.Info("raw subscriber started", zap.String("stream", stream))
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		msgs, err := owlredis.ReadFromStreamWithBlock(ctx, client, stream, consumerGroup, consumer, readCount, readBlock)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			logger.Warn("read raw stream", zap.String("stream", stream), zap.Error(err))
			time.Sleep(time.Second)
			continue
		}
		for _, m := range msgs {
			if err := handler.Handle(ctx, m.Values); err != nil {
				logger.Warn("handle raw msg", zap.String("stream", stream), zap.Error(err))
			}
			client.XAck(ctx, stream, consumerGroup, m.ID)
		}
	}
}
