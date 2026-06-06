// config_alarm_device_consumer.go — qinglan 端 config:alarmDevice:stream 订阅。
//
// 数据流：wisefido-data UPSERT device_config (alarm.device_config) → Publish 精确 device_addr
// payload 到 config:alarmDevice:stream → 本 consumer Invalidate qinglan AlarmEnablementCache。
//
// 与 sensor/cardagg 同消息源、独立 consumer group ("wisefido-qinglan-alarm-device")，
// 每个 service 独立 ack 互不抢消息。详 [[alarm_enablement_invalidate_design]] §C。

package consumer

import (
	"context"
	"encoding/json"
	"time"

	rediscommon "owl-common/redis"

	redislib "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const (
	qinglanAlarmDeviceGroup    = "wisefido-qinglan-alarm-device"
	qinglanAlarmDeviceConsumer = "consumer-qinglan-alarm-device"
	qinglanReadCount           = 10
	qinglanReadBlock           = 5 * time.Second
)

// AlarmEnablementInvalidator 仅声明本 consumer 所需失效面，避免 consumer→service 反向导入环。
// service.AlarmEnablementCache.Invalidate 隐式满足。
type AlarmEnablementInvalidator interface {
	Invalidate(deviceAddr string)
}

// AlarmDeviceConfigConsumer 订阅 config:alarmDevice:stream，按 device_addr 失效 enablement cache。
type AlarmDeviceConfigConsumer struct {
	client      *redislib.Client
	invalidator AlarmEnablementInvalidator
	logger      *zap.Logger
}

func NewAlarmDeviceConfigConsumer(
	client *redislib.Client,
	invalidator AlarmEnablementInvalidator,
	logger *zap.Logger,
) *AlarmDeviceConfigConsumer {
	return &AlarmDeviceConfigConsumer{client: client, invalidator: invalidator, logger: logger}
}

func (c *AlarmDeviceConfigConsumer) Start(ctx context.Context) {
	if err := rediscommon.CreateConsumerGroup(ctx, c.client,
		rediscommon.StreamConfigAlarmDevice.Name, qinglanAlarmDeviceGroup); err != nil {
		c.logger.Warn("qinglan alarm-device config: create consumer group", zap.Error(err))
	}
	go c.runLoop(ctx)
	c.logger.Info("qinglan alarm-device config consumer started",
		zap.String("stream", rediscommon.StreamConfigAlarmDevice.Name),
		zap.String("group", qinglanAlarmDeviceGroup))
}

func (c *AlarmDeviceConfigConsumer) runLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		msgs, err := rediscommon.ReadFromStreamWithBlock(ctx, c.client,
			rediscommon.StreamConfigAlarmDevice.Name, qinglanAlarmDeviceGroup, qinglanAlarmDeviceConsumer,
			qinglanReadCount, qinglanReadBlock)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.logger.Warn("qinglan alarm-device config: read stream", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}
		for _, m := range msgs {
			c.handleRaw(m.Values)
			c.client.XAck(ctx, rediscommon.StreamConfigAlarmDevice.Name, qinglanAlarmDeviceGroup, m.ID)
		}
	}
}

type cloudEventsEnvelope struct {
	Data json.RawMessage `json:"data"`
}

type alarmDeviceData struct {
	DeviceAddr  string `json:"device_addr"`
	SettingType string `json:"setting_type"`
}

// handleRaw 解析 CloudEvents envelope → alarmDevice data → invalidate。
func (c *AlarmDeviceConfigConsumer) handleRaw(raw map[string]interface{}) {
	dataStr, _ := raw["data"].(string)
	if dataStr == "" {
		return
	}
	var env cloudEventsEnvelope
	if err := json.Unmarshal([]byte(dataStr), &env); err != nil {
		c.logger.Warn("qinglan alarm-device config: unmarshal envelope", zap.Error(err))
		return
	}
	var d alarmDeviceData
	if err := json.Unmarshal(env.Data, &d); err != nil {
		c.logger.Warn("qinglan alarm-device config: unmarshal data", zap.Error(err))
		return
	}
	if d.DeviceAddr == "" {
		return
	}
	c.invalidator.Invalidate(d.DeviceAddr)
	c.logger.Info("qinglan alarm device config invalidated",
		zap.String("device_addr", d.DeviceAddr),
		zap.String("setting", d.SettingType))
}
