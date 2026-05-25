// config_card_consumer.go — sensor 端 config:card:stream 订阅。
//
// 数据流：wisefido-data PublishConfigChanged → 本 consumer 解析 op/cards/device_addrs →
// 按 op 触发 reload 闭包：
//   - reset/delete/update → 重 load rooms (registerAllRooms) + device→room 映射 (mapDevicesToRooms)
//   - 任何 op → SpatialCache.InvalidateAll() (清 per-unit cache，下次 query lazy reload)
//
// "card 变更很少"原则：不区分 affected 范围，每次全量 reload；rooms ≈ 13 + devices ≈ 22 全
// reload < 200ms，10 次/天可忽略。换来 0 引入新概念 / 无 incremental 复杂度。
//
// 替代 engine.SetRoutesReloader 的 60s 轮询：事件驱动，monitor toggle 1s 内同步到 SpatialCache。

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
	sensorConfigCardGroup    = "wisefido-sensor-config-card"
	sensorConfigCardConsumer = "consumer-sensor-config-card"
)

// ConfigChangedReloader sensor 端 reload 闭包接口。
// 实现：bootstrap closure，调 registerAllRooms / mapDevicesToRooms。
type ConfigChangedReloader interface {
	ReloadRooms(ctx context.Context) error
	ReloadDevices(ctx context.Context) error
}

// SpatialInvalidator SpatialCache 失效接口（wiring.SpatialCache 隐式实现）。
type SpatialInvalidator interface {
	InvalidateAll()
}

type ConfigCardConsumer struct {
	client   *redislib.Client
	reloader ConfigChangedReloader
	spatial  SpatialInvalidator
	logger   *zap.Logger
}

func NewConfigCardConsumer(
	client *redislib.Client,
	reloader ConfigChangedReloader,
	spatial SpatialInvalidator,
	logger *zap.Logger,
) *ConfigCardConsumer {
	return &ConfigCardConsumer{
		client:   client,
		reloader: reloader,
		spatial:  spatial,
		logger:   logger,
	}
}

func (c *ConfigCardConsumer) Start(ctx context.Context) {
	if err := rediscommon.CreateConsumerGroup(ctx, c.client,
		rediscommon.StreamConfigCard.Name, sensorConfigCardGroup); err != nil {
		c.logger.Warn("sensor config-card: create consumer group", zap.Error(err))
	}
	go c.runLoop(ctx)
	c.logger.Info("sensor config-card consumer started",
		zap.String("stream", rediscommon.StreamConfigCard.Name),
		zap.String("group", sensorConfigCardGroup))
}

func (c *ConfigCardConsumer) runLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		msgs, err := rediscommon.ReadFromStreamWithBlock(ctx, c.client,
			rediscommon.StreamConfigCard.Name, sensorConfigCardGroup, sensorConfigCardConsumer,
			sensorReadCount, sensorReadBlock)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.logger.Warn("sensor config-card: read stream", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}
		for _, m := range msgs {
			c.handleRaw(ctx, m.Values)
			c.client.XAck(ctx, rediscommon.StreamConfigCard.Name, sensorConfigCardGroup, m.ID)
		}
	}
}

type configChangedDataSensor struct {
	Op          string   `json:"op"`
	Cards       []string `json:"cards"`
	DeviceAddrs []string `json:"device_addrs"`
	DeviceUIDs  []string `json:"device_uids"`
}

func (c *ConfigCardConsumer) handleRaw(ctx context.Context, raw map[string]interface{}) {
	dataStr, _ := raw["data"].(string)
	if dataStr == "" {
		return
	}
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal([]byte(dataStr), &env); err != nil {
		c.logger.Warn("sensor config-card: unmarshal envelope", zap.Error(err))
		return
	}
	var d configChangedDataSensor
	if err := json.Unmarshal(env.Data, &d); err != nil {
		c.logger.Warn("sensor config-card: unmarshal data", zap.Error(err))
		return
	}

	if c.reloader != nil {
		if err := c.reloader.ReloadRooms(ctx); err != nil {
			c.logger.Warn("sensor config-card: reload rooms", zap.Error(err))
		}
		if err := c.reloader.ReloadDevices(ctx); err != nil {
			c.logger.Warn("sensor config-card: reload devices", zap.Error(err))
		}
	}
	if c.spatial != nil {
		c.spatial.InvalidateAll()
	}

	c.logger.Info("sensor config-card processed",
		zap.String("op", d.Op),
		zap.Int("cards", len(d.Cards)),
		zap.Int("device_addrs", len(d.DeviceAddrs)),
		zap.Int("device_uids", len(d.DeviceUIDs)))
}
