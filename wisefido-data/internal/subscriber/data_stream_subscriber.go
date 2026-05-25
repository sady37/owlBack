package subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"owl-common/card"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// realtimeCacheTTL 实时数据 cache 过期阈值。与前端 REALTIME_TTL_MS=12000 / cardagg MonitorFieldTTL=12s 对齐。
// 设备静默 ≥ 此阈值后 GetCardRealtimeData 返 nil，避免历史 latch（如人离床后旧 pose 一直被 SSE 推给新连接）。
// 注意：仅 realtime cache 适用；status cache 由 cardagg 显式驱动，不应自动过期。
const realtimeCacheTTL = 12 * time.Second

// CachedCardData 带版本号的缓存数据
type CachedCardData struct {
	Data      map[string]interface{}
	Version   uint64
	UpdatedAt time.Time // 仅 realtime cache 使用，用于 staleness 判定；status cache 留 zero value
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
	stateReader *card.Reader
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
		stateReader:         card.NewReader(redisClient),
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
// stream 格式（cardagg PublishMonitor）：Values 含 type=MsgTypeMonitor, card_id, device_addr, data；data 为 deviceId→{ trackId→fields } 的 JSON
func (s *DataStreamSubscriber) HandleCardRealtimeMessage(ctx context.Context, message redis.XMessage) error {
	if t, _ := message.Values[card.MsgTypeKey].(string); t != "" && t != card.MsgTypeMonitor {
		s.logger.Debug("card:realtime message_type mismatch", zap.String("got", t), zap.String("expected", card.MsgTypeMonitor))
	}
	cardID, _ := message.Values["card_id"].(string)
	if cardID == "" {
		return fmt.Errorf("card_id not found in message, keys=%v", mapKeys(message.Values))
	}

	var msgData map[string]interface{}
	if dataStr, ok := message.Values["data"].(string); ok && dataStr != "" {
		if err := json.Unmarshal([]byte(dataStr), &msgData); err != nil {
			return fmt.Errorf("failed to parse card realtime data: %w", err)
		}
	}
	if msgData == nil {
		msgData = make(map[string]interface{})
	}
	msgData["card_id"] = cardID
	now := time.Now()
	s.cacheMu.Lock()
	if existing, ok := s.cardRealtimeCache[cardID]; ok {
		existing.Data = msgData
		existing.Version++
		existing.UpdatedAt = now
	} else {
		s.cardRealtimeCache[cardID] = &CachedCardData{Data: msgData, Version: 1, UpdatedAt: now}
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

	//s.logger.Info("[SUB] cached card realtime data",
	//	zap.String("card_id", cardID),
	//	zap.String("message_id", message.ID),
	//	zap.Int("cache_size", len(s.cardRealtimeCache)),
	//)

	return nil
}

// HandleCardStatusMessage 处理卡片状态变更通知
// stream 格式：Values 含 type=MsgTypeEvent, card_id, fields, data（CardStatus JSON）；与 cardagg 写入一致
// 推送前端时始终用 Redis 全量状态（含 alarm_state），避免 stream 消息仅含部分字段导致前端图标不显色
func (s *DataStreamSubscriber) HandleCardStatusMessage(ctx context.Context, message redis.XMessage) error {
	if t, _ := message.Values[card.MsgTypeKey].(string); t != "" && t != card.MsgTypeEvent {
		s.logger.Debug("card:status message_type mismatch", zap.String("got", t), zap.String("expected", card.MsgTypeEvent))
	}
	cardID, _ := message.Values["card_id"].(string)
	if cardID == "" {
		return fmt.Errorf("card_id not found in status stream message")
	}

	// 始终从 Redis 读全量状态（含 alarm_state），保证前端 Overview 报警图标能按 active_err/active_warning 正确显色
	status, err := s.stateReader.ReadCardStatus(ctx, cardID)
	if err != nil || status == nil {
		return nil
	}
	status.CardID = cardID

	data := cardStatusToMap(status)

	s.cacheMu.Lock()
	if existing, ok := s.cardStatusCache[cardID]; ok {
		existing.Data = data
		existing.Version++
	} else {
		s.cardStatusCache[cardID] = &CachedCardData{Data: data, Version: 1}
	}
	s.cacheMu.Unlock()

	s.markDirty(cardID)

	select {
	case s.cardStatusEventChan <- CardStatusEvent{CardID: cardID, Data: deepCopyMap(data)}:
	default:
		s.logger.Debug("card status event channel full", zap.String("card_id", cardID))
	}

	s.logger.Debug("card:status updated", zap.String("card_id", cardID))
	return nil
}

func cardStatusToMap(st *card.CardStatus) map[string]interface{} {
	if st == nil {
		return nil
	}
	b, _ := json.Marshal(st)
	var m map[string]interface{}
	if json.Unmarshal(b, &m) != nil {
		return map[string]interface{}{"card_id": st.CardID}
	}
	return m
}

// ReadCardStateSnapshot 从 card:state:{cardID} Hash 读全量 CardStatus（初始 snapshot 用）
func (s *DataStreamSubscriber) ReadCardStateSnapshot(ctx context.Context, cardID string) map[string]interface{} {
	status, err := s.stateReader.ReadCardStatus(ctx, cardID)
	if err != nil || status == nil {
		return nil
	}
	result := cardStatusToMap(status)

	s.cacheMu.Lock()
	if existing, ok := s.cardStatusCache[cardID]; ok {
		existing.Data = result
		existing.Version++
	} else {
		s.cardStatusCache[cardID] = &CachedCardData{Data: deepCopyMap(result), Version: 1}
	}
	s.cacheMu.Unlock()

	return result
}

// GetCardRealtimeData 获取缓存的卡片实时数据。
// staleness 判定：cache 超 realtimeCacheTTL 未刷新即视为静默，返 nil 让上层（SSE ticker / radar handler）跳过推送，
// 避免设备静默期把历史 latch（如人离床后的 sit_up=1/pose=9）推给新建连接的前端。
func (s *DataStreamSubscriber) GetCardRealtimeData(cardID string) map[string]interface{} {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()
	cached, ok := s.cardRealtimeCache[cardID]
	if !ok {
		return nil
	}
	if !cached.UpdatedAt.IsZero() && time.Since(cached.UpdatedAt) > realtimeCacheTTL {
		return nil
	}
	return deepCopyMap(cached.Data)
}

func deepCopyMap(src map[string]interface{}) map[string]interface{} {
	if src == nil {
		return nil
	}
	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		switch val := v.(type) {
		case map[string]interface{}:
			dst[k] = deepCopyMap(val)
		case []interface{}:
			cp := make([]interface{}, len(val))
			for i, item := range val {
				if m, ok := item.(map[string]interface{}); ok {
					cp[i] = deepCopyMap(m)
				} else {
					cp[i] = item
				}
			}
			dst[k] = cp
		default:
			dst[k] = v
		}
	}
	return dst
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

func mapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
