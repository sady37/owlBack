// config_alarm_device_consumer.go — sensor 端 config:alarmDevice:stream 订阅。
//
// 数据流：wisefido-data UPDATE spatial_config (alarm.cloud_config) → Publish 精确 device_addr
// payload 到 config:alarmDevice:stream → 本 consumer Invalidate sensor AlarmEnablementCache。
//
// 与 cardagg 端 AlarmDeviceHandler ([[alarm_enablement_invalidate_design]] §C) 对称：相同 payload
// 解析，独立 consumer group（"wisefido-sensor-alarm-device"），避免与 cardagg "$cons-cardagg" 抢消息。

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
	sensorAlarmDeviceGroup    = "wisefido-sensor-alarm-device"
	sensorAlarmDeviceConsumer = "consumer-sensor-alarm-device"
)

// AlarmEnablementInvalidator 仅声明本 consumer 所需的失效面，避免 consumer→service 反向导入环。
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
		("test:" + rediscommon.StreamConfigAlarmDevice.Name), sensorAlarmDeviceGroup); err != nil {
		c.logger.Warn("sensor alarm-device config: create consumer group", zap.Error(err))
	}
	go c.runLoop(ctx)
	c.logger.Info("sensor alarm-device config consumer started",
		zap.String("stream", ("test:" + rediscommon.StreamConfigAlarmDevice.Name)),
		zap.String("group", sensorAlarmDeviceGroup))
}

func (c *AlarmDeviceConfigConsumer) runLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		msgs, err := rediscommon.ReadFromStreamWithBlock(ctx, c.client,
			("test:" + rediscommon.StreamConfigAlarmDevice.Name), sensorAlarmDeviceGroup, sensorAlarmDeviceConsumer,
			sensorReadCount, sensorReadBlock)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.logger.Warn("sensor alarm-device config: read stream", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}
		for _, m := range msgs {
			c.handleRaw(m.Values)
			c.client.XAck(ctx, ("test:" + rediscommon.StreamConfigAlarmDevice.Name), sensorAlarmDeviceGroup, m.ID)
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
// payload 结构与 cardagg `AlarmDeviceHandler` 对齐（[[alarm_enablement_invalidate_design]]）。
func (c *AlarmDeviceConfigConsumer) handleRaw(raw map[string]interface{}) {
	dataStr, _ := raw["data"].(string)
	if dataStr == "" {
		return
	}
	var env cloudEventsEnvelope
	if err := json.Unmarshal([]byte(dataStr), &env); err != nil {
		c.logger.Warn("sensor alarm-device config: unmarshal envelope", zap.Error(err))
		return
	}
	var d alarmDeviceData
	if err := json.Unmarshal(env.Data, &d); err != nil {
		c.logger.Warn("sensor alarm-device config: unmarshal data", zap.Error(err))
		return
	}
	if d.DeviceAddr == "" {
		return
	}
	c.invalidator.Invalidate(d.DeviceAddr)
	c.logger.Info("sensor alarm device config invalidated",
		zap.String("device_addr", d.DeviceAddr),
		zap.String("setting", d.SettingType))
}
