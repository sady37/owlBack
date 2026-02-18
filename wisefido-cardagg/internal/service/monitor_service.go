package service

import (
	"context"
	"time"

	"owl-common/card"
	"owl-common/redis"
	"wisefido-cardagg/internal/repository"

	"go.uber.org/zap"
)

// CardRealtimePublisher 发布 CardRealTime 到 stream 的接口
type CardRealtimePublisher interface {
	PublishCardRealTime(ctx context.Context, tenantID, cardID string, rtData *card.CardRealTime)
}

// MonitorService monitor 消息处理服务
type MonitorService struct {
	repo      repository.CacheRepository
	publisher CardRealtimePublisher
	logger    *zap.Logger
}

// NewMonitorService 创建 monitor 服务
func NewMonitorService(repo repository.CacheRepository, publisher CardRealtimePublisher, logger *zap.Logger) *MonitorService {
	return &MonitorService{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
	}
}

// toMapSlice 将 []interface{} 转为 []map[string]interface{}，跳过非 map 项，去除 raw_original
func toMapSlice(dataValue []interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(dataValue))
	for _, item := range dataValue {
		if m, ok := item.(map[string]interface{}); ok {
			delete(m, "raw_original")
			result = append(result, m)
		}
	}
	return result
}

// ProcessMonitor 处理 monitor 消息，按设备原样存储观测数据
func (s *MonitorService) ProcessMonitor(ctx context.Context, msg *redis.IoTStreamMessage) error {
	if msg == nil {
		s.logger.Info("Monitor message is nil")
		return nil
	}

	// 获取或创建 CardRealTime
	realtimeData, err := s.repo.GetCardRealTime(ctx, msg.CardID)
	if err != nil {
		s.logger.Warn("Failed to get realtime data", zap.Error(err))
		realtimeData = &card.CardRealTime{
			CardID:    msg.CardID,
			Timestamp: msg.Timestamp,
		}
	}
	if realtimeData == nil {
		realtimeData = &card.CardRealTime{
			CardID:    msg.CardID,
			Timestamp: msg.Timestamp,
		}
	}
	realtimeData.CardID = msg.CardID
	if realtimeData.Devices == nil {
		realtimeData.Devices = make(map[string]*card.DeviceRealTime)
	}

	// 清除超过 2 秒的过期设备
	now := time.Now().Unix()
	s.pruneStaleDevices(realtimeData, now, 2)

	// 获取或创建该设备的实时数据
	dev := realtimeData.Devices[msg.DeviceID]
	if dev == nil {
		dev = &card.DeviceRealTime{
			DeviceID:   msg.DeviceID,
			DeviceType: msg.DeviceType,
		}
	}

	// 时间戳检查：只接受 >= 当前设备时间戳的数据
	if msg.Timestamp < dev.Timestamp {
		return nil
	}

	// 原样存储，不分拣
	dev.Data = toMapSlice(msg.DataValue)
	dev.Timestamp = msg.Timestamp
	realtimeData.Devices[msg.DeviceID] = dev

	// 更新 CardRealTime 顶层时间戳
	if msg.Timestamp > realtimeData.Timestamp {
		realtimeData.Timestamp = msg.Timestamp
	}

	// 写入 Redis（TTL 300 秒）
	if err := s.repo.SetCardRealTime(ctx, realtimeData, 300*time.Second); err != nil {
		s.logger.Warn("Failed to set realtime data", zap.Error(err))
		return err
	}

	// 发布到 card:realtime:stream 供 wisefido-data SSE 推送
	if s.publisher != nil {
		s.publisher.PublishCardRealTime(ctx, msg.TenantID, msg.CardID, realtimeData)
	}

	s.logger.Debug("Monitor message processed",
		zap.String("card_id", msg.CardID),
		zap.String("device_id", msg.DeviceID),
		zap.String("category", msg.Category),
	)

	return nil
}

// pruneStaleDevices 清除 Devices map 中时间戳超过 maxAgeSec 秒的设备
func (s *MonitorService) pruneStaleDevices(rtData *card.CardRealTime, now, maxAgeSec int64) {
	if rtData == nil || rtData.Devices == nil {
		return
	}
	cutoff := now - maxAgeSec
	for devID, dev := range rtData.Devices {
		if dev.Timestamp < cutoff {
			delete(rtData.Devices, devID)
		}
	}
}
