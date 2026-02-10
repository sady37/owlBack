package service

import (
	"context"
	"strings"
	"time"

	"owl-common/card"
	"owl-common/redis"
	"wisefido-cardagg/internal/repository"

	"go.uber.org/zap"
)

// MonitorService monitor 消息处理服务
type MonitorService struct {
	repo   repository.CacheRepository
	logger *zap.Logger
}

// NewMonitorService 创建 monitor 服务
func NewMonitorService(repo repository.CacheRepository, logger *zap.Logger) *MonitorService {
	return &MonitorService{
		repo:   repo,
		logger: logger,
	}
}

// FromMapToDeviceTrack 从 IoTStreamMessage 转换为 DeviceTrack（track 数据，1Hz 更新）
// 清理 raw_original 字段，只保留 category="track" 的数据
func (s *MonitorService) FromMapToDeviceTrack(msg *redis.IoTStreamMessage) *card.DeviceTrack {
	if msg == nil {
		return nil
	}

	dt := &card.DeviceTrack{
		DeviceID:   msg.DeviceID,
		DeviceType: msg.DeviceType,
		Timestamp:  msg.Timestamp,
		Category:   msg.Category,
	}

	// 清理 DataValue 中的 raw_original 字段，只保留 "track" 的数据
	cleaned := make([]interface{}, 0, len(msg.DataValue))
	for _, item := range msg.DataValue {
		if itemMap, ok := item.(map[string]interface{}); ok {
			// 只保留 category="track" 的数据
			if category, _ := itemMap["category"].(string); category != "track" {
				continue
			}

			cleanedItem := make(map[string]interface{})
			for k, v := range itemMap {
				if k != "raw_original" {
					cleanedItem[k] = v
				}
			}
			cleaned = append(cleaned, cleanedItem)
		}
	}
	dt.DataValue = cleaned

	// 如果没有 track 数据，返回 nil
	if len(dt.DataValue) == 0 {
		return nil
	}

	return dt
}

// FromMapToDeviceVital 从 IoTStreamMessage 转换为 DeviceVital（vital 数据，2Hz 更新）
// 清理 raw_original 字段，只保留 category="vital" 的数据
func (s *MonitorService) FromMapToDeviceVital(msg *redis.IoTStreamMessage) *card.DeviceVital {
	if msg == nil {
		return nil
	}

	dv := &card.DeviceVital{
		DeviceID:   msg.DeviceID,
		DeviceType: msg.DeviceType,
		Timestamp:  msg.Timestamp,
		Category:   msg.Category,
	}

	// 清理 DataValue 中的 raw_original 字段，只保留 "vital" 的数据
	cleaned := make([]interface{}, 0, len(msg.DataValue))
	for _, item := range msg.DataValue {
		if itemMap, ok := item.(map[string]interface{}); ok {
			// 只保留 category="vital" 的数据
			if category, _ := itemMap["category"].(string); category != "vital" {
				continue
			}

			cleanedItem := make(map[string]interface{})
			for k, v := range itemMap {
				if k != "raw_original" {
					cleanedItem[k] = v
				}
			}
			cleaned = append(cleaned, cleanedItem)
		}
	}
	dv.DataValue = cleaned

	// 如果没有 vital 数据，返回 nil
	if len(dv.DataValue) == 0 {
		return nil
	}

	return dv
}

// ProcessMonitor 处理 monitor 消息，分离 track 和 vital 数据
func (s *MonitorService) ProcessMonitor(ctx context.Context, msg *redis.IoTStreamMessage) error {
	if msg == nil {
		s.logger.Info("Monitor message is nil")
		return nil
	}

	// 获取或创建 CardRealTime（只读realtime数据，不涉及status）
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

	// 检查是否包含 track 数据
	hasTrack := strings.Contains(msg.Category, "track")
	// 检查是否包含 vital 数据
	hasVital := strings.Contains(msg.Category, "vital")

	// 处理 track 数据
	if hasTrack {
		if dt := s.FromMapToDeviceTrack(msg); dt != nil {
			// 直接比对当前设备的时间戳（在 realtimeData 中）
			if s.shouldUpdateDeviceTrack(realtimeData, msg.DeviceID, dt.Timestamp) {
				// 替换该设备的 track 数据（而不是追加）
				s.replaceOrAppendDeviceTrack(realtimeData, dt)
				s.logger.Debug("Track data updated", zap.String("device_id", msg.DeviceID))
			}
		}
	} else {
		// 没有 track 数据时，清理该设备的 track 数据
		s.removeDeviceTrackFromRealtime(realtimeData, msg.DeviceID)
	}

	// 处理 vital 数据
	if hasVital {
		if dv := s.FromMapToDeviceVital(msg); dv != nil {
			// 直接比对当前设备的时间戳（在 realtimeData 中）
			if s.shouldUpdateDeviceVital(realtimeData, msg.DeviceID, dv.Timestamp) {
				// 替换该设备的 vital 数据（而不是追加）
				s.replaceOrAppendDeviceVital(realtimeData, dv)
				s.logger.Debug("Vital data updated", zap.String("device_id", msg.DeviceID))
			}
		}
	}

	// 更新实时数据到 Redis（TTL 5 秒）
	if msg.Timestamp > realtimeData.Timestamp {
		realtimeData.Timestamp = msg.Timestamp
	}
	if err := s.repo.SetCardRealTime(ctx, realtimeData, 5*time.Second); err != nil {
		s.logger.Warn("Failed to set realtime data", zap.Error(err))
		return err
	}

	s.logger.Info("Monitor message processed",
		zap.String("card_id", msg.CardID),
		zap.String("device_id", msg.DeviceID),
		zap.String("category", msg.Category),
	)

	return nil
}

// shouldUpdateDeviceTrack 检查是否应该更新设备的 track 数据
// 直接在 CardRealTime 中检查该设备最后一条 track 记录的时间戳
func (s *MonitorService) shouldUpdateDeviceTrack(rtData *card.CardRealTime, deviceID string, newTimestamp int64) bool {
	if rtData == nil || len(rtData.TrackData) == 0 {
		return true
	}

	// 查找该设备的最后一条 track 数据，比对时间戳
	for _, trackIface := range rtData.TrackData {
		track, ok := trackIface.(*card.DeviceTrack)
		if !ok {
			continue
		}
		if track.DeviceID == deviceID {
			// 只有新数据时间戳 >= 现有记录时，才更新
			return newTimestamp >= track.Timestamp
		}
	}

	// 该设备没有 track 数据，则可以追加
	return true
}

// shouldUpdateDeviceVital 检查是否应该更新设备的 vital 数据
// 直接在 CardRealTime 中检查该设备最后一条 vital 记录的时间戳
func (s *MonitorService) shouldUpdateDeviceVital(rtData *card.CardRealTime, deviceID string, newTimestamp int64) bool {
	if rtData == nil || len(rtData.VitalData) == 0 {
		return true
	}

	// 查找该设备的最后一条 vital 数据，比对时间戳
	for _, vitalIface := range rtData.VitalData {
		vital, ok := vitalIface.(*card.DeviceVital)
		if !ok {
			continue
		}
		if vital.DeviceID == deviceID {
			// 只有新数据时间戳 >= 现有记录时，才更新
			return newTimestamp >= vital.Timestamp
		}
	}

	// 该设备没有 vital 数据，则可以追加
	return true
}

// removeDeviceTrackFromRealtime 从 CardRealTime 中移除指定设备的 track 数据
func (s *MonitorService) removeDeviceTrackFromRealtime(rtData *card.CardRealTime, deviceID string) {
	if rtData == nil || len(rtData.TrackData) == 0 {
		return
	}
	filtered := make([]interface{}, 0, len(rtData.TrackData))
	for _, trackIface := range rtData.TrackData {
		track, ok := trackIface.(*card.DeviceTrack)
		if !ok || track.DeviceID != deviceID {
			filtered = append(filtered, trackIface)
		}
	}
	rtData.TrackData = filtered
}

// replaceOrAppendDeviceTrack 替换或追加设备的 track 数据
// 同一设备只保留最新的一条记录
func (s *MonitorService) replaceOrAppendDeviceTrack(rtData *card.CardRealTime, newTrack *card.DeviceTrack) {
	if rtData == nil || newTrack == nil {
		return
	}

	// 查找是否已存在该设备的 track 数据
	found := false
	for i, trackIface := range rtData.TrackData {
		track, ok := trackIface.(*card.DeviceTrack)
		if !ok || track == nil {
			continue
		}
		if track.DeviceID == newTrack.DeviceID {
			// 替换现有记录
			rtData.TrackData[i] = newTrack
			found = true
			break
		}
	}

	// 如果不存在，则追加
	if !found {
		rtData.TrackData = append(rtData.TrackData, newTrack)
	}
}

// replaceOrAppendDeviceVital 替换或追加设备的 vital 数据
// 同一设备只保留最新的一条记录
func (s *MonitorService) replaceOrAppendDeviceVital(rtData *card.CardRealTime, newVital *card.DeviceVital) {
	if rtData == nil || newVital == nil {
		return
	}

	// 查找是否已存在该设备的 vital 数据
	found := false
	for i, vitalIface := range rtData.VitalData {
		vital, ok := vitalIface.(*card.DeviceVital)
		if !ok || vital == nil {
			continue
		}
		if vital.DeviceID == newVital.DeviceID {
			// 替换现有记录
			rtData.VitalData[i] = newVital
			found = true
			break
		}
	}

	// 如果不存在，则追加
	if !found {
		rtData.VitalData = append(rtData.VitalData, newVital)
	}
}
