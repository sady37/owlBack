package consumer

// PR1 (A7): sensor-side monitor consumer.
//
// 订阅 iot:monitor:stream 用独立 consumer group "wisefido-sensor-monitor"，
// 解析后写入 sensor 自己的 MonitorBuffer（通过 BufferWriter 接口注入，避免与 service 包形成
// 循环依赖；现有 sensor service 包已 import consumer 包）。**不发布 realtime 流，不触状态/报警**——
// 那些都是 B/C 组迁移后续才接管。本文件只做"喂 buffer"，让后续 DeriveAndWriteState /
// DeriveBedStateFromRealtime 之类的派生函数能在 sensor 内取到原始 monitor 数据。

import (
	"context"
	"encoding/json"
	"net/netip"
	"strconv"
	"strings"
	"time"

	"owl-common/observation"
	rediscommon "owl-common/redis"

	redislib "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const (
	sensorMonitorGroup    = "wisefido-sensor-monitor"
	sensorMonitorConsumer = "consumer-sensor-monitor"
	sensorReadCount       = 10
	sensorReadBlock       = 5 * time.Second

	// sensorMonitorMaxInboundAgeMs 与 cardagg 同步：> 6s 的旧消息丢弃。
	sensorMonitorMaxInboundAgeMs = 6000
)

// BufferWriter 仅声明 sensor MonitorConsumer 所需的 buffer 写入面，便于注入而不引入 service 包依赖。
// service.MonitorBuffer.Write 满足此接口。
type BufferWriter interface {
	Write(cardID, deviceAddr, deviceType, trackKey string, fields map[string]any, tsMs int64)
}

// MonitorConsumer 喂 sensor 端 MonitorBuffer 的薄消费者。
type MonitorConsumer struct {
	client *redislib.Client
	buffer BufferWriter
	logger *zap.Logger
}

func NewMonitorConsumer(client *redislib.Client, buffer BufferWriter, logger *zap.Logger) *MonitorConsumer {
	return &MonitorConsumer{client: client, buffer: buffer, logger: logger}
}

// Start 单独 goroutine 跑读流循环，consumer group 与 cardagg 隔离。
func (c *MonitorConsumer) Start(ctx context.Context) {
	if err := rediscommon.CreateConsumerGroup(ctx, c.client, ("test:" + rediscommon.StreamMonitor.Name), sensorMonitorGroup); err != nil {
		c.logger.Warn("sensor monitor: create consumer group", zap.Error(err))
	}
	go c.runLoop(ctx)
	c.logger.Info("sensor monitor consumer started",
		zap.String("stream", ("test:" + rediscommon.StreamMonitor.Name)),
		zap.String("group", sensorMonitorGroup))
}

func (c *MonitorConsumer) runLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		msgs, err := rediscommon.ReadFromStreamWithBlock(ctx, c.client,
			("test:" + rediscommon.StreamMonitor.Name), sensorMonitorGroup, sensorMonitorConsumer,
			sensorReadCount, sensorReadBlock)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.logger.Warn("sensor monitor: read stream", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}
		for _, m := range msgs {
			c.handleRaw(m.Values)
			c.client.XAck(ctx, ("test:" + rediscommon.StreamMonitor.Name), sensorMonitorGroup, m.ID)
		}
	}
}

func (c *MonitorConsumer) handleRaw(raw map[string]interface{}) {
	msg, err := rediscommon.FromStreamMap(raw)
	if err != nil {
		c.logger.Warn("sensor monitor: parse", zap.Error(err))
		return
	}
	if !msg.DeviceAddr.IsValid() {
		return
	}
	nowMs := time.Now().UnixMilli()
	if nowMs-msg.Timestamp > sensorMonitorMaxInboundAgeMs {
		return
	}
	fields := rediscommon.FirstDataValue(msg.DataValue)
	if fields == nil {
		return
	}
	deviceKey := msg.DeviceAddr.String()
	trackID := sensorResolveTrackID(fields)
	// MonitorBuffer 历史 API 第一参数叫 cardID，sensor 物理寻址用 /96 bed spatial prefix 替代
	// （raw envelope SubjectEntity = device_uid，不进入 sensor 内部索引）；VitalSource 扫描时
	// 仍按"groupKey"迭代，键名是 spatial 而非 card，与 sensor-Card 解耦原则一致。
	bedPrefix := netip.PrefixFrom(msg.DeviceAddr, 96).Masked().String()
	c.buffer.Write(bedPrefix, deviceKey, msg.DeviceType, strconv.Itoa(trackID), fields, msg.Timestamp)
}

// sensorResolveTrackID 与 cardagg consumer.resolveTrackID 同语义；无效返回 observation.TrackInvalid。
func sensorResolveTrackID(fields map[string]interface{}) int {
	if fields == nil {
		return observation.TrackInvalid
	}
	v, ok := fields["track_id"]
	if !ok || v == nil {
		return observation.TrackInvalid
	}
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case int64:
		return int(x)
	case json.Number:
		n, err := x.Int64()
		if err != nil {
			return observation.TrackInvalid
		}
		return int(n)
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return observation.TrackInvalid
		}
		n, err := strconv.Atoi(s)
		if err != nil {
			return observation.TrackInvalid
		}
		return n
	default:
		return observation.TrackInvalid
	}
}
