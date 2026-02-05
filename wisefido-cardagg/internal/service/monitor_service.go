package service

import (
	"context"
	"strconv"
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

// ProcessMonitor 处理 monitor 消息
func (s *MonitorService) ProcessMonitor(ctx context.Context, msg *redis.IoTStreamMessage) error {
	if msg == nil {
		s.logger.Info("Monitor message is nil")
		return nil
	}

	// 解析 category，例如 "track2.vital1" -> track: 2, vital: 1
	trackCount, vitalCount := parseCategoryCount(msg.Category)

	// 获取或创建 RealtimeData
	realtimeData, err := s.repo.GetRealtimeData(ctx, msg.CardID)
	if err != nil {
		s.logger.Warn("Failed to get realtime data", zap.Error(err))
		realtimeData = &card.RealtimeData{
			CardID:    msg.CardID,
			Timestamp: msg.Timestamp,
		}
	}

	if realtimeData == nil {
		realtimeData = &card.RealtimeData{
			CardID:    msg.CardID,
			Timestamp: msg.Timestamp,
		}
	}

	// 处理 track 数据
	trackUpdated := false
	if trackCount > 0 {
		trackUpdated = s.processTrackData(ctx, msg, realtimeData, trackCount)
	}

	if trackUpdated && vitalCount == 0 {
		s.clearVitalForDevice(ctx, msg, realtimeData)
	}

	// 处理 vital 数据
	if vitalCount > 0 {
		s.processVitalData(ctx, msg, realtimeData, trackCount, vitalCount)
	}

	if trackCount == 0 && vitalCount == 0 {
		s.cleanupDeviceCache(ctx, msg, realtimeData)
	}

	// 更新实时数据到 Redis（TTL 5 秒）
	if msg.Timestamp > realtimeData.Timestamp {
		realtimeData.Timestamp = msg.Timestamp
	}
	if err := s.repo.SetRealtimeData(ctx, msg.CardID, realtimeData, 5*time.Second); err != nil {
		s.logger.Warn("Failed to set realtime data", zap.Error(err))
		return err
	}

	s.logger.Info("Monitor message processed",
		zap.String("card_id", msg.CardID),
		zap.String("device_id", msg.DeviceID),
		zap.Int("track_count", trackCount),
		zap.Int("vital_count", vitalCount),
	)

	return nil
}

// processTrackData 处理 track 数据
func (s *MonitorService) processTrackData(ctx context.Context, msg *redis.IoTStreamMessage, realtimeData *card.RealtimeData, trackCount int) bool {
	// 初始化 Postures 如果为 nil
	if realtimeData.Postures == nil {
		realtimeData.Postures = make([]card.DevicePosture, 0)
	}

	trackIndex := 0
	trackUpdated := false
	for i := 0; i < len(msg.DataValue) && trackIndex < trackCount; i++ {
		dataItem, ok := msg.DataValue[i].(map[string]interface{})
		if !ok {
			continue
		}

		category, _ := dataItem["category"].(string)
		if category != "track" {
			continue
		}

		trackIndex++

		// 提取 pose 字段（已是 int）
		posture := extractPose(dataItem, "pose")

		if !s.shouldUpdatePosture(ctx, msg, msg.DeviceID) {
			continue
		}

		// 创建或更新 DevicePosture
		devicePosture := &card.DevicePosture{
			DeviceID:  msg.DeviceID,
			Timestamp: msg.Timestamp,
			Postures:  []int{posture},
		}

		// 保存到 Redis
		if err := s.repo.SetDevicePosture(ctx, msg.CardID, devicePosture); err != nil {
			s.logger.Warn("Failed to set device posture", zap.Error(err))
		}

		// 更新 RealtimeData.Postures（追加到数组）
		realtimeData.Postures = append(realtimeData.Postures, *devicePosture)
		trackUpdated = true
	}

	return trackUpdated
}

// processVitalData 处理 vital 数据
func (s *MonitorService) processVitalData(ctx context.Context, msg *redis.IoTStreamMessage, realtimeData *card.RealtimeData, trackCount int, vitalCount int) {
	// 初始化 Vital 如果为 nil
	if realtimeData.Vital == nil {
		realtimeData.Vital = make([]card.VitalSimplified, 0)
	}

	vitalIndex := 0
	for i := 0; i < len(msg.DataValue) && vitalIndex < vitalCount; i++ {
		dataItem, ok := msg.DataValue[i].(map[string]interface{})
		if !ok {
			continue
		}

		category, _ := dataItem["category"].(string)
		if category != "vital" {
			continue
		}

		vitalIndex++

		// 提取 vital 字段
		respiratoryRate := extractIntPtr(dataItem, "respiratory_rate")
		heartRate := extractIntPtr(dataItem, "heart_rate")
		sleepStatus := extractStringPtr(dataItem, "sleep_status")
		stability := extractStringPtr(dataItem, "stability")

		if !s.shouldUpdateVital(ctx, msg, msg.DeviceID) {
			continue
		}

		vital := &card.VitalSimplified{
			DeviceID:        msg.DeviceID,
			Timestamp:       msg.Timestamp,
			RespiratoryRate: respiratoryRate,
			HeartRate:       heartRate,
			SleepStatus:     sleepStatus,
			Stability:       stability,
		}

		// 保存到 Redis
		if err := s.repo.SetVitalSimplified(ctx, msg.CardID, vital); err != nil {
			s.logger.Warn("Failed to set vital simplified", zap.Error(err))
		}

		// 更新或添加到 RealtimeData.Vital
		realtimeData.Vital = append(realtimeData.Vital, *vital)
	}
}

func (s *MonitorService) shouldUpdatePosture(ctx context.Context, msg *redis.IoTStreamMessage, deviceID string) bool {
	existing, err := s.repo.GetDevicePosture(ctx, msg.CardID, deviceID)
	if err != nil {
		s.logger.Warn("Failed to get existing posture", zap.Error(err))
		return true
	}
	if existing == nil {
		return true
	}
	return msg.Timestamp >= existing.Timestamp
}

func (s *MonitorService) shouldUpdateVital(ctx context.Context, msg *redis.IoTStreamMessage, deviceID string) bool {
	existing, err := s.repo.GetVitalSimplified(ctx, msg.CardID, deviceID)
	if err != nil {
		s.logger.Warn("Failed to get existing vital", zap.Error(err))
		return true
	}
	if existing == nil {
		return true
	}
	return msg.Timestamp >= existing.Timestamp
}

func (s *MonitorService) clearVitalForDevice(ctx context.Context, msg *redis.IoTStreamMessage, realtimeData *card.RealtimeData) {
	if err := s.repo.DeleteVitalSimplified(ctx, msg.CardID, msg.DeviceID); err != nil {
		s.logger.Warn("Failed to delete vital simplified", zap.Error(err))
	}
	s.removeDeviceVitalFromRealtime(realtimeData, msg.DeviceID)
}

func (s *MonitorService) removeDevicePostureFromRealtime(realtimeData *card.RealtimeData, deviceID string) {
	filtered := make([]card.DevicePosture, 0, len(realtimeData.Postures))
	for _, p := range realtimeData.Postures {
		if p.DeviceID != deviceID {
			filtered = append(filtered, p)
		}
	}
	realtimeData.Postures = filtered
}

func (s *MonitorService) cleanupDeviceCache(ctx context.Context, msg *redis.IoTStreamMessage, realtimeData *card.RealtimeData) {
	if err := s.repo.DeleteDevicePosture(ctx, msg.CardID, msg.DeviceID); err != nil {
		s.logger.Warn("Failed to delete device posture", zap.Error(err))
	}
	// 从 Postures 数组中移除该设备的数据
	s.removeDevicePostureFromRealtime(realtimeData, msg.DeviceID)
	s.clearVitalForDevice(ctx, msg, realtimeData)
}

func (s *MonitorService) removeDeviceVitalFromRealtime(realtimeData *card.RealtimeData, deviceID string) {
	if realtimeData == nil || len(realtimeData.Vital) == 0 {
		return
	}
	filtered := realtimeData.Vital[:0]
	for _, vital := range realtimeData.Vital {
		if vital.DeviceID != deviceID {
			filtered = append(filtered, vital)
		}
	}
	realtimeData.Vital = filtered
}

// parseCategoryCount 解析 category 中的类型计数，例如 "track2.vital1" -> (2, 1)
func parseCategoryCount(category string) (int, int) {
	trackCount := 0
	vitalCount := 0

	parts := strings.Split(category, ".")
	for _, part := range parts {
		if strings.HasPrefix(part, "track") {
			// 从 "track2" 中提取数字 2
			n, _ := parseCountFromString(part)
			trackCount = n
		} else if strings.HasPrefix(part, "vital") {
			n, _ := parseCountFromString(part)
			vitalCount = n
		}
	}

	return trackCount, vitalCount
}

// parseCountFromString 从 "track2" 这样的字符串中提取数字
func parseCountFromString(s string) (int, error) {
	count := 0
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			count = count*10 + int(ch-'0')
		}
	}
	return count, nil
}

// extractPose 读取已为整数的 pose 字段
func extractPose(data map[string]interface{}, key string) int {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case int:
			return clampPose(v)
		case float64:
			return clampPose(int(v))
		case string:
			if n, err := strconv.Atoi(v); err == nil {
				return clampPose(n)
			}
		}
	}
	return 0
}

func clampPose(value int) int {
	if value < 0 {
		return 0
	}
	if value > 11 {
		return 11
	}
	return value
}

// CleanupExpired 清理过期的实时数据
// 如果数据的时间戳早于 realtimeData.Timestamp-200秒，则将其删除
func (s *MonitorService) CleanupExpired(ctx context.Context, cardID string) error {
	// 获取实时数据
	realtimeData, err := s.repo.GetRealtimeData(ctx, cardID)
	if err != nil {
		s.logger.Warn("Failed to get realtime data for cleanup", zap.String("card_id", cardID), zap.Error(err))
		return err
	}

	if realtimeData == nil {
		return nil
	}

	// 计算过期时间阈值 (当前时间戳 - 200秒)
	expirationThreshold := realtimeData.Timestamp - 200

	// 清理过期的 Vital 数据
	cleanedVital := make([]card.VitalSimplified, 0)
	for _, vital := range realtimeData.Vital {
		if vital.Timestamp >= expirationThreshold {
			// 保留未过期的数据
			cleanedVital = append(cleanedVital, vital)
		} else {
			// 删除过期的 vital 数据
			if err := s.repo.DeleteVitalSimplified(ctx, cardID, vital.DeviceID); err != nil {
				s.logger.Warn("Failed to delete expired vital simplified",
					zap.String("card_id", cardID),
					zap.String("device_id", vital.DeviceID),
					zap.Error(err))
			}
		}
	}
	realtimeData.Vital = cleanedVital

	// 清理过期的 Postures 数据
	cleanedPostures := make([]card.DevicePosture, 0, len(realtimeData.Postures))
	for _, posture := range realtimeData.Postures {
		devicePosture, err := s.repo.GetDevicePosture(ctx, cardID, posture.DeviceID)
		if err != nil {
			s.logger.Warn("Failed to get device posture for cleanup check",
				zap.String("card_id", cardID),
				zap.String("device_id", posture.DeviceID),
				zap.Error(err))
			// 保留该数据，不清理（发生错误时保留）
			cleanedPostures = append(cleanedPostures, posture)
			continue
		}

		// 如果设备姿势数据未过期，则保留它
		if devicePosture != nil && devicePosture.Timestamp >= expirationThreshold {
			cleanedPostures = append(cleanedPostures, posture)
		} else if devicePosture != nil {
			// 如果已过期，删除缓存中的数据
			if err := s.repo.DeleteDevicePosture(ctx, cardID, posture.DeviceID); err != nil {
				s.logger.Warn("Failed to delete expired device posture",
					zap.String("card_id", cardID),
					zap.String("device_id", posture.DeviceID),
					zap.Error(err))
			}
		} else {
			// 如果 devicePosture 为 nil 但没有错误，保留该数据
			cleanedPostures = append(cleanedPostures, posture)
		}
	}
	realtimeData.Postures = cleanedPostures

	// 更新实时数据到 Redis（保持原有的 TTL 5 秒）
	if err := s.repo.SetRealtimeData(ctx, cardID, realtimeData, 5*time.Second); err != nil {
		s.logger.Warn("Failed to update realtime data after cleanup", zap.Error(err))
		return err
	}

	s.logger.Info("Expired data cleanup completed",
		zap.String("card_id", cardID),
		zap.Int64("threshold_time", expirationThreshold))

	return nil
}

// extractIntPtr 从 map 中提取可选的 int 字段
func extractIntPtr(data map[string]interface{}, key string) *int {
	if val, ok := data[key]; ok {
		if f64, ok := val.(float64); ok {
			v := int(f64)
			return &v
		}
	}
	return nil
}

// extractStringPtr 从 map 中提取可选的 string 字段
func extractStringPtr(data map[string]interface{}, key string) *string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return &str
		}
	}
	return nil
}
