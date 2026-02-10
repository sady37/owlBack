package subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// CachedCardData 带版本号的缓存数据
type CachedCardData struct {
	Data    map[string]interface{}
	Version uint64
}

// CardStatusEvent 卡片状态事件
type CardStatusEvent struct {
	CardID string
	Data   map[string]interface{}
}

// DataStreamSubscriber 数据流订阅器
// 订阅 card:realtime:stream (track/vital) 、card:status:stream (bed/room/alarms/device status)
// 缓存最新数据供 SSE 推送使用，并通过 channel 通知事件变化
// 注意：deviceStatus 已经集成在 card:status:stream 中，不需单独订阅
type DataStreamSubscriber struct {
	redisClient *redis.Client
	logger      *zap.Logger

	// 缓存最新数据（带版本号）
	cardRealtimeCache map[string]*CachedCardData // cardID -> realtime data + version
	cardStatusCache   map[string]*CachedCardData // cardID -> status data + version

	cacheMu sync.RWMutex

	// dirty set：记录自上次消费以来有更新的 cardID
	dirtySet map[string]struct{}
	dirtyMu  sync.Mutex

	// 事件通知 channel（事件驱动）
	cardStatusEventChan chan CardStatusEvent // 卡片状态变化通知
	cardRealtimeUpdated chan string          // 卡片实时数据更新通知（cardID）

	// Consumer group 管理
	consumerGroup string
	consumerName  string
}

// NewDataStreamSubscriber 创建数据流订阅器
func NewDataStreamSubscriber(redisClient *redis.Client, logger *zap.Logger) *DataStreamSubscriber {
	return &DataStreamSubscriber{
		redisClient:         redisClient,
		logger:              logger,
		cardRealtimeCache:   make(map[string]*CachedCardData),
		cardStatusCache:     make(map[string]*CachedCardData),
		dirtySet:            make(map[string]struct{}),
		cardStatusEventChan: make(chan CardStatusEvent, 10), // 缓冲 10 个事件
		cardRealtimeUpdated: make(chan string, 10),          // 缓冲 10 个 cardID 更新
		consumerGroup:       "wisefido-data-consumer",
		consumerName:        "wisefido-data-consumer-1",
	}
}

// Start 启动数据流订阅器
func (s *DataStreamSubscriber) Start(ctx context.Context) error {
	streams := []string{
		"card:realtime:stream", // 卡片实时数据（track/vital）
		"card:status:stream",   // 卡片状态数据（bed/room/alarms/device status）
	}

	s.logger.Info("Starting data stream subscriber",
		zap.Strings("streams", streams),
		zap.String("consumer_group", s.consumerGroup),
	)

	// 所有stream共用同一个group
	for _, stream := range streams {
		if err := s.createConsumerGroupIfNotExistsWithName(ctx, stream, s.consumerGroup); err != nil {
			if err.Error() == "BUSYGROUP Consumer Group name already exists" {
				s.logger.Info("Consumer group already exists",
					zap.String("stream", stream),
					zap.String("group", s.consumerGroup),
					zap.Error(err),
				)
				continue
			}
			s.logger.Warn("Failed to create consumer group",
				zap.String("stream", stream),
				zap.String("group", s.consumerGroup),
				zap.Error(err),
			)
		}
	}

	// 注意：消费协程由 main.go 的 subscribeDataStream 启动
	return nil
}

// HandleCardRealtimeMessage 处理卡片实时数据消息
func (s *DataStreamSubscriber) HandleCardRealtimeMessage(ctx context.Context, message redis.XMessage) error {
	s.logger.Debug("Received card realtime message",
		zap.String("message_id", message.ID),
	)

	// 解析消息（支持 PublishJSONToStream 的包装格式）
	var msgData map[string]interface{}
	if dataStr, ok := message.Values["data"].(string); ok {
		if err := json.Unmarshal([]byte(dataStr), &msgData); err != nil {
			return fmt.Errorf("failed to parse card realtime data: %w", err)
		}
	}

	cardID, _ := message.Values["card_id"].(string)
	if cardID == "" {
		return fmt.Errorf("card_id not found in message")
	}

	// 缓存最新数据（带版本号），存解析后的 msgData 避免 data 字段嵌套
	if msgData == nil {
		msgData = message.Values
	}
	s.cacheMu.Lock()
	if existing, ok := s.cardRealtimeCache[cardID]; ok {
		existing.Data = msgData
		existing.Version++
	} else {
		s.cardRealtimeCache[cardID] = &CachedCardData{Data: msgData, Version: 1}
	}
	s.cacheMu.Unlock()

	// 仅首次 dirty 时通知 channel，避免同一 card 短时间重复 push
	if s.markDirty(cardID) {
		select {
		case s.cardRealtimeUpdated <- cardID:
		default:
			s.logger.Debug("Card realtime update channel full, dropping notification",
				zap.String("card_id", cardID))
		}
	}

	s.logger.Debug("Cached card realtime data",
		zap.String("card_id", cardID),
		zap.String("message_id", message.ID),
	)

	return nil
}

// HandleCardStatusMessage 处理卡片状态数据消息
func (s *DataStreamSubscriber) HandleCardStatusMessage(ctx context.Context, message redis.XMessage) error {
	s.logger.Debug("Received card status message",
		zap.String("message_id", message.ID),
	)

	// 解析消息
	var msgData map[string]interface{}
	if dataStr, ok := message.Values["data"].(string); ok {
		if err := json.Unmarshal([]byte(dataStr), &msgData); err != nil {
			return fmt.Errorf("failed to parse card status data: %w", err)
		}
	}

	cardID, _ := message.Values["card_id"].(string)
	if cardID == "" {
		return fmt.Errorf("card_id not found in message")
	}

	// 缓存最新数据（带版本号），存解析后的 msgData 避免 data 字段嵌套
	if msgData == nil {
		msgData = message.Values
	}
	s.cacheMu.Lock()
	if existing, ok := s.cardStatusCache[cardID]; ok {
		existing.Data = msgData
		existing.Version++
	} else {
		s.cardStatusCache[cardID] = &CachedCardData{Data: msgData, Version: 1}
	}
	s.cacheMu.Unlock()

	s.markDirty(cardID)

	// 状态事件通知（含 alarm），保留 channel 推送
	select {
	case s.cardStatusEventChan <- CardStatusEvent{CardID: cardID, Data: msgData}:
	default:
		s.logger.Debug("Card status event channel full, dropping notification",
			zap.String("card_id", cardID))
	}

	s.logger.Debug("Cached card status data",
		zap.String("card_id", cardID),
		zap.String("message_id", message.ID),
	)

	return nil
}

// GetCardRealtimeData 获取缓存的卡片实时数据
func (s *DataStreamSubscriber) GetCardRealtimeData(cardID string) map[string]interface{} {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()
	if cached, ok := s.cardRealtimeCache[cardID]; ok {
		return cached.Data
	}
	return nil
}

// GetCardRealtimeVersion 获取缓存的卡片实时数据版本号（0 表示无数据）
func (s *DataStreamSubscriber) GetCardRealtimeVersion(cardID string) uint64 {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()
	if cached, ok := s.cardRealtimeCache[cardID]; ok {
		return cached.Version
	}
	return 0
}

// GetCardStatusData 获取缓存的卡片状态数据
func (s *DataStreamSubscriber) GetCardStatusData(cardID string) map[string]interface{} {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()
	if cached, ok := s.cardStatusCache[cardID]; ok {
		return cached.Data
	}
	return nil
}

// GetCardStatusVersion 获取缓存的卡片状态数据版本号（0 表示无数据）
func (s *DataStreamSubscriber) GetCardStatusVersion(cardID string) uint64 {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()
	if cached, ok := s.cardStatusCache[cardID]; ok {
		return cached.Version
	}
	return 0
}

// GetCardRealtimeUpdatedChan 获取卡片实时数据更新通知 channel
func (s *DataStreamSubscriber) GetCardRealtimeUpdatedChan() <-chan string {
	return s.cardRealtimeUpdated
}

// GetCardStatusEventChan 获取卡片状态事件通知 channel（含 alarm 等即时事件）
func (s *DataStreamSubscriber) GetCardStatusEventChan() <-chan CardStatusEvent {
	return s.cardStatusEventChan
}

// markDirty 标记 cardID 有数据更新，返回 true 表示首次标记（之前不在 dirtySet 中）
func (s *DataStreamSubscriber) markDirty(cardID string) bool {
	s.dirtyMu.Lock()
	_, existed := s.dirtySet[cardID]
	s.dirtySet[cardID] = struct{}{}
	s.dirtyMu.Unlock()
	return !existed
}

// ConsumeDirty 取出并清空 dirtySet，返回自上次调用以来有更新的 cardID 集合
func (s *DataStreamSubscriber) ConsumeDirty() map[string]struct{} {
	s.dirtyMu.Lock()
	if len(s.dirtySet) == 0 {
		s.dirtyMu.Unlock()
		return nil
	}
	old := s.dirtySet
	s.dirtySet = make(map[string]struct{})
	s.dirtyMu.Unlock()
	return old
}

// Stop 停止订阅器
func (s *DataStreamSubscriber) Stop() error {
	s.logger.Info("Data stream subscriber stopped")
	return nil
}

// createConsumerGroupIfNotExistsWithName 创建 Consumer Group（指定 groupName，如果不存在）
func (s *DataStreamSubscriber) createConsumerGroupIfNotExistsWithName(ctx context.Context, stream, groupName string) error {
	err := s.redisClient.XGroupCreateMkStream(ctx, stream, groupName, "0").Err()
	if err != nil {
		// 如果 group 已存在会返回 BUSYGROUP，这里直接返回错误
		return err
	}
	return nil
}
