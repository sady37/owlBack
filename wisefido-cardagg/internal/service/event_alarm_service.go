package service

import (
	"context"
	"strings"
	"time"

	"owl-common/card"
	"owl-common/redis"
	"wisefido-cardagg/internal/publisher"
	"wisefido-cardagg/internal/repository"

	"go.uber.org/zap"
)

// 报警级别映射：字符串 → 优先级数字（0 最高）
var alarmLevelPriority = map[string]int{
	"EMERG":   0,
	"ALERT":   1,
	"CRIT":    2,
	"ERR":     3,
	"WARNING": 4,
	"NOTICE":  5,
}

// EventAlarmService event/alarm 消息处理服务
type EventAlarmService struct {
	cacheRepo     repository.CacheRepository
	cardPublisher *publisher.CardStreamPublisher
	logger        *zap.Logger
}

// NewEventAlarmService 创建 event/alarm 服务
func NewEventAlarmService(repo repository.CacheRepository, cardPublisher *publisher.CardStreamPublisher, logger *zap.Logger) *EventAlarmService {
	return &EventAlarmService{
		cacheRepo:     repo,
		cardPublisher: cardPublisher,
		logger:        logger,
	}
}

// GetRealtimeData 获取实时数据（供 handler 时序检查使用）
func (s *EventAlarmService) GetRealtimeData(ctx context.Context, cardID string) (*card.RealtimeData, error) {
	return s.cacheRepo.GetRealtimeData(ctx, cardID)
}

// HandleAlarmProcessMessage 处理报警处理消息（来自 config:alarm.process:stream）
// 当报警被确认或解决时，更新卡片的 ActiveAlarms 状态
// 参数 data 包含：alarm_level, alarm_type, alarm_timestamp
func (s *EventAlarmService) HandleAlarmProcessMessage(ctx context.Context, cardID, alarmLevel, alarmType string, alarmTimestamp int64) error {
	// 获取卡片的实时数据
	realtimeData, err := s.cacheRepo.GetRealtimeData(ctx, cardID)
	if err != nil {
		s.logger.Warn("Failed to get realtime data for alarm process",
			zap.String("card_id", cardID),
			zap.Error(err))
		return nil // 不返回错误，防止消息重试
	}

	if realtimeData == nil {
		s.logger.Debug("RealtimeData not found for alarm process",
			zap.String("card_id", cardID))
		return nil
	}

	// 检查是否有 ActiveAlarms
	if realtimeData.ActiveAlarms == nil {
		s.logger.Debug("No active alarms to process",
			zap.String("card_id", cardID),
			zap.String("alarm_level", alarmLevel))
		return nil
	}

	// 减少相应级别的计数（确保 >= 0）
	s.decrementAlarmCount(realtimeData, alarmLevel)

	// 检查是否需要清空 NowAlarm
	// 根据 alarmType, alarmLevel, alarmTimestamp 来判断是否是同一个报警
	if s.shouldClearNowAlarm(realtimeData, alarmLevel, alarmType, alarmTimestamp) {
		realtimeData.ActiveAlarms.NowAlarm = ""
		s.logger.Info("NowAlarm cleared",
			zap.String("card_id", cardID),
			zap.String("alarm_level", alarmLevel),
			zap.String("alarm_type", alarmType),
			zap.Int64("alarm_timestamp", alarmTimestamp))
	}

	// 更新实时数据到 Redis（TTL 5 秒）
	if err := s.cacheRepo.SetRealtimeData(ctx, cardID, realtimeData, 5*time.Second); err != nil {
		s.logger.Warn("Failed to set realtime data after alarm process",
			zap.String("card_id", cardID),
			zap.Error(err))
		return err
	}

	// 异步发布到 iot:card:stream（非阻塞）
	if s.cardPublisher != nil {
		go s.cardPublisher.PublishCardRealtime(ctx, "", cardID, realtimeData)
	}

	s.logger.Info("Alarm process message handled",
		zap.String("card_id", cardID),
		zap.String("alarm_level", alarmLevel),
		zap.String("alarm_type", alarmType))

	return nil
}

// decrementAlarmCount 减少指定级别的报警计数（确保 >= 0）
func (s *EventAlarmService) decrementAlarmCount(realtimeData *card.RealtimeData, alarmLevel string) {
	if realtimeData == nil || realtimeData.ActiveAlarms == nil {
		return
	}

	switch alarmLevel {
	case "EMERG":
		if realtimeData.ActiveAlarms.EMERG > 0 {
			realtimeData.ActiveAlarms.EMERG--
		}
	case "ALERT":
		if realtimeData.ActiveAlarms.ALERT > 0 {
			realtimeData.ActiveAlarms.ALERT--
		}
	case "CRIT":
		if realtimeData.ActiveAlarms.CRIT > 0 {
			realtimeData.ActiveAlarms.CRIT--
		}
	case "ERR":
		if realtimeData.ActiveAlarms.ERR > 0 {
			realtimeData.ActiveAlarms.ERR--
		}
	case "WARNING":
		if realtimeData.ActiveAlarms.WARNING > 0 {
			realtimeData.ActiveAlarms.WARNING--
		}
	case "NOTICE":
		if realtimeData.ActiveAlarms.NOTICE > 0 {
			realtimeData.ActiveAlarms.NOTICE--
		}
	}

	s.logger.Debug("Alarm count decremented",
		zap.String("alarm_level", alarmLevel),
		zap.Int("emerg", realtimeData.ActiveAlarms.EMERG),
		zap.Int("alert", realtimeData.ActiveAlarms.ALERT),
		zap.Int("crit", realtimeData.ActiveAlarms.CRIT),
		zap.Int("err", realtimeData.ActiveAlarms.ERR),
		zap.Int("warning", realtimeData.ActiveAlarms.WARNING),
		zap.Int("notice", realtimeData.ActiveAlarms.NOTICE))
}

// shouldClearNowAlarm 判断是否应该清空 NowAlarm
// 条件：alarmType, alarmLevel, alarmTimestamp 与当前 NowAlarm 一致
func (s *EventAlarmService) shouldClearNowAlarm(realtimeData *card.RealtimeData, alarmLevel, alarmType string, alarmTimestamp int64) bool {
	if realtimeData == nil || realtimeData.ActiveAlarms == nil || realtimeData.ActiveAlarms.NowAlarm == "" {
		return false
	}

	// NowAlarm 格式为 "AlarmLevel.AlarmType"，如 "EMERG.Fall"
	// 注意：在 updateActiveAlarm 中，NowAlarm 被设置为 "alarmLevel + '.Alarm'"
	// 所以这里需要解析 NowAlarm 来提取 alarmLevel
	parts := strings.Split(realtimeData.ActiveAlarms.NowAlarm, ".")
	if len(parts) < 1 {
		return false
	}

	nowAlarmLevel := parts[0]

	// 检查级别是否相同
	if nowAlarmLevel != alarmLevel {
		return false
	}

	// 检查时间戳是否相同（精确匹配，防止清空其他报警）
	if realtimeData.ActiveAlarms.Timestamp != alarmTimestamp {
		s.logger.Debug("NowAlarm timestamp mismatch, not clearing",
			zap.String("now_alarm", realtimeData.ActiveAlarms.NowAlarm),
			zap.Int64("existing_ts", realtimeData.ActiveAlarms.Timestamp),
			zap.Int64("new_ts", alarmTimestamp))
		return false
	}

	return true
}

// HandleMessage 处理 event/alarm 消息
// 【入口处统一时序检查】msg.Timestamp + 5s > realtimeData.Timestamp（5秒时间窗口）
func (s *EventAlarmService) HandleMessage(ctx context.Context, msg *redis.IoTStreamMessage) error {
	// 【时序检查 - 统一入口】5秒时间窗口
	// 获取现有的 RealtimeData 用于时间戳比较
	realtimeData, err := s.cacheRepo.GetRealtimeData(ctx, msg.CardID)
	if err == nil && realtimeData != nil {
		// 如果 msg.Timestamp + 5s <= realtimeData.Timestamp，说明消息太旧，直接拒绝
		if msg.Timestamp+5 < realtimeData.Timestamp {
			s.logger.Info("Event/alarm message dropped: outside 5s window",
				zap.String("card_id", msg.CardID),
				zap.String("topic_type", msg.TopicType),
				zap.Int64("msg_ts", msg.Timestamp),
				zap.Int64("realtime_ts", realtimeData.Timestamp),
				zap.Int64("window_end", msg.Timestamp+5))
			return nil
		}
	}

	// ✓ 通过时间窗口检查，分流处理 event/alarm
	switch msg.TopicType {
	case "event":
		return s.handleEvent(ctx, msg)
	case "alarm":
		return s.handleAlarm(ctx, msg)
	default:
		s.logger.Warn("Unknown topic_type", zap.String("topic_type", msg.TopicType))
		return nil
	}
}

// handleEvent 处理 event 类消息
// DataValue 为数组，每项为事件对象，项内含 category 字段
func (s *EventAlarmService) handleEvent(ctx context.Context, msg *redis.IoTStreamMessage) error {
	// 获取或创建 RealtimeData
	realtimeData, err := s.cacheRepo.GetRealtimeData(ctx, msg.CardID)
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

	// 遍历 data_value 数组处理每条事件
	for _, item := range msg.DataValue {
		dataItem, ok := item.(map[string]interface{})
		if !ok {
			s.logger.Debug("Invalid event data item type", zap.Any("item", item))
			continue
		}

		// 提取 category
		category, _ := dataItem["category"].(string)

		// 根据 category 分类处理
		switch category {
		case "pose":
			s.handlePoseEvent(ctx, msg, dataItem, realtimeData)
		case "BedState":
			s.handleBedStateEvent(ctx, msg, dataItem, realtimeData)
		case "RoomState":
			s.handleRoomStateEvent(ctx, msg, dataItem, realtimeData)
		default:
			s.logger.Debug("Unknown event category",
				zap.String("category", category),
				zap.String("card_id", msg.CardID))
		}
	}

	// 更新实时数据到 Redis（TTL 12H）
	if msg.Timestamp > realtimeData.Timestamp {
		realtimeData.Timestamp = msg.Timestamp
	}
	if err := s.cacheRepo.SetRealtimeData(ctx, msg.CardID, realtimeData, 12*time.Hour); err != nil {
		s.logger.Warn("Failed to set realtime data", zap.Error(err))
		return err
	}

	// 异步发布到 iot:card:stream（非阻塞）
	if s.cardPublisher != nil {
		go s.cardPublisher.PublishCardRealtime(ctx, msg.TenantID, msg.CardID, realtimeData)
	}

	s.logger.Info("Event message processed",
		zap.String("card_id", msg.CardID),
		zap.String("device_id", msg.DeviceID),
		zap.String("category", msg.Category),
		zap.Int("data_count", len(msg.DataValue)))

	return nil
}

// handleAlarm 处理 alarm 类消息
// DataValue 为数组，每项为报警对象，项内含 category 字段
func (s *EventAlarmService) handleAlarm(ctx context.Context, msg *redis.IoTStreamMessage) error {
	// 获取或创建 RealtimeData
	realtimeData, err := s.cacheRepo.GetRealtimeData(ctx, msg.CardID)
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

	// 遍历 data_value 数组处理每条报警
	for _, item := range msg.DataValue {
		dataItem, ok := item.(map[string]interface{})
		if !ok {
			s.logger.Debug("Invalid alarm data item type", zap.Any("item", item))
			continue
		}

		// 提取报警字段（不需要知道具体的报警类型）
		alarmLevel, _ := dataItem["alarm_level"].(string)
		alarmTimestamp, _ := dataItem["timestamp"].(float64)

		// 统一处理：更新 ActiveAlarms（时间戳 + 级别双重检查）
		s.updateActiveAlarm(realtimeData, alarmLevel, int64(alarmTimestamp))
	}

	// 更新实时数据到 Redis（TTL 5 秒）
	if msg.Timestamp > realtimeData.Timestamp {
		realtimeData.Timestamp = msg.Timestamp
	}
	if err := s.cacheRepo.SetRealtimeData(ctx, msg.CardID, realtimeData, 5*time.Second); err != nil {
		s.logger.Warn("Failed to set realtime data", zap.Error(err))
		return err
	}

	// 异步发布到 iot:card:stream（非阻塞）
	if s.cardPublisher != nil {
		go s.cardPublisher.PublishCardRealtime(ctx, msg.TenantID, msg.CardID, realtimeData)
	}

	s.logger.Info("Alarm message processed",
		zap.String("card_id", msg.CardID),
		zap.String("device_id", msg.DeviceID),
		zap.String("category", msg.Category),
		zap.Int("data_count", len(msg.DataValue)))

	return nil
}

// ========== Event 类别处理 ==========

// handleEnter2OutEvent 处理 enter2out 事件：设备离开区域，清理 vital 数据
func (s *EventAlarmService) handleEnter2OutEvent(ctx context.Context, msg *redis.IoTStreamMessage, dataItem map[string]interface{}, realtimeData *card.RealtimeData) {
	s.logger.Info("Handling enter2out event",
		zap.String("card_id", msg.CardID),
		zap.String("device_id", msg.DeviceID),
		zap.Any("data", dataItem))

	// 清理该设备的 vital 数据
	if err := s.cacheRepo.DeleteVitalSimplified(ctx, msg.CardID, msg.DeviceID); err != nil {
		s.logger.Warn("Failed to delete vital simplified",
			zap.String("card_id", msg.CardID),
			zap.String("device_id", msg.DeviceID),
			zap.Error(err))
	}

	// 从 realtimeData 中移除该设备的 vital
	s.removeDeviceVitalFromRealtime(realtimeData, msg.DeviceID)
}

// handlePoseEvent 处理 pose 事件
// dataItem 包含 {"category": "pose", "track_id": 1, "pose": "Fall", ...}
func (s *EventAlarmService) handlePoseEvent(ctx context.Context, msg *redis.IoTStreamMessage, dataItem map[string]interface{}, realtimeData *card.RealtimeData) {
	trackID, _ := dataItem["track_id"].(float64)
	pose, _ := dataItem["pose"].(string)

	s.logger.Info("Handling pose event",
		zap.String("card_id", msg.CardID),
		zap.String("device_id", msg.DeviceID),
		zap.Int("track_id", int(trackID)),
		zap.String("pose", pose))

	// TODO: 根据 pose 值更新或删除 posture 缓存
}

// handleBedStateEvent 处理 BedState 事件
// dataItem 包含 {"category": "BedState", "bed_id": "uuid-bbb", "CurrentState": "in_bed", "StateSince": 1234567890}
// 检查：初始状态为空 → 直接填充；有值 → 检查 Timestamp > 原记录才更新
func (s *EventAlarmService) handleBedStateEvent(ctx context.Context, msg *redis.IoTStreamMessage, dataItem map[string]interface{}, realtimeData *card.RealtimeData) {
	bedID, _ := dataItem["bed_id"].(string)
	currentState, _ := dataItem["CurrentState"].(string)
	stateSince, _ := dataItem["StateSince"].(float64)

	s.logger.Info("Handling BedState event",
		zap.String("card_id", msg.CardID),
		zap.String("bed_id", bedID),
		zap.String("current_state", currentState),
		zap.Int64("state_since", int64(stateSince)))

	// 检查初始状态是否为空
	if realtimeData.BedState == nil {
		// 初始状态为空，直接填充
		realtimeData.BedState = &card.BedState{
			BedID:        bedID,
			CurrentState: currentState,
			Timestamp:    int64(stateSince),
		}
		s.logger.Info("BedState initialized",
			zap.String("bed_id", bedID),
			zap.String("current_state", currentState))
		return
	}

	// 【设备级时序检查】只有在 5s 窗口内才进行检查和更新
	// msg.Timestamp + 5s > realtimeData.Timestamp 已在 handler 层检查通过
	// 现在检查设备级别的时间戳
	if stateSince <= float64(realtimeData.BedState.Timestamp) {
		s.logger.Debug("BedState not updated: device timestamp not newer",
			zap.Float64("new_ts", stateSince),
			zap.Int64("existing_ts", realtimeData.BedState.Timestamp))
		return
	}

	// 有值且时间戳更新，直接更新
	realtimeData.BedState.BedID = bedID
	realtimeData.BedState.CurrentState = currentState
	realtimeData.BedState.Timestamp = int64(stateSince)
	s.logger.Info("BedState updated",
		zap.String("bed_id", bedID),
		zap.String("current_state", currentState),
		zap.Int64("state_since", int64(stateSince)))
}

// handleRoomStateEvent 处理 RoomState 事件
// dataItem 包含 {"category": "RoomState", "room_id": "uuid-rrr", "room_name": "101", "PeopleCount": 1, "StayTime": 5, "StateSince": 1234567890}
// 检查：初始状态为空 → 直接填充；有值 → 检查 Timestamp > 原记录才更新
func (s *EventAlarmService) handleRoomStateEvent(ctx context.Context, msg *redis.IoTStreamMessage, dataItem map[string]interface{}, realtimeData *card.RealtimeData) {
	roomID, _ := dataItem["room_id"].(string)
	roomName, _ := dataItem["room_name"].(string)
	peopleCount, _ := dataItem["PeopleCount"].(float64)
	stayTime, _ := dataItem["StayTime"].(float64)
	stateSince, _ := dataItem["StateSince"].(float64)

	s.logger.Info("Handling RoomState event",
		zap.String("card_id", msg.CardID),
		zap.String("room_id", roomID),
		zap.String("room_name", roomName),
		zap.Int("people_count", int(peopleCount)),
		zap.Int("stay_time", int(stayTime)),
		zap.Int64("state_since", int64(stateSince)))

	// 检查初始状态是否为空
	if realtimeData.RoomState == nil {
		// 初始状态为空，直接填充
		realtimeData.RoomState = &card.RoomState{
			RoomID:      roomID,
			RoomName:    roomName,
			PeopleCount: int(peopleCount),
			StayTime:    int(stayTime),
			Timestamp:   int64(stateSince),
		}
		s.logger.Info("RoomState initialized",
			zap.String("room_id", roomID),
			zap.String("room_name", roomName),
			zap.Int("people_count", int(peopleCount)))
		return
	}

	// 【设备级时序检查】只有在 5s 窗口内才进行检查和更新
	// msg.Timestamp + 5s > realtimeData.Timestamp 已在 handler 层检查通过
	// 现在检查设备级别的时间戳
	if stateSince <= float64(realtimeData.RoomState.Timestamp) {
		s.logger.Debug("RoomState not updated: device timestamp not newer",
			zap.Float64("new_ts", stateSince),
			zap.Int64("existing_ts", realtimeData.RoomState.Timestamp))
		return
	}

	// 有值且时间戳更新，直接更新
	realtimeData.RoomState.RoomID = roomID
	realtimeData.RoomState.RoomName = roomName
	realtimeData.RoomState.PeopleCount = int(peopleCount)
	realtimeData.RoomState.StayTime = int(stayTime)
	realtimeData.RoomState.Timestamp = int64(stateSince)
	s.logger.Info("RoomState updated",
		zap.String("room_id", roomID),
		zap.String("room_name", roomName),
		zap.Int("people_count", int(peopleCount)),
		zap.Int64("state_since", int64(stateSince)))
}

// ========== Alarm 类别处理 ==========

// updateActiveAlarm 更新 ActiveAlarms：双重比对策略
// 比对 1：timestamp 比较（防止旧数据覆盖新数据）
// 比对 2：alarm_level 比较（防止低级别报警覆盖高级别报警）
func (s *EventAlarmService) updateActiveAlarm(realtimeData *card.RealtimeData, alarmLevel string, alarmTimestamp int64) {
	if realtimeData == nil {
		return
	}

	// 初始化 ActiveAlarms（如果为空）
	if realtimeData.ActiveAlarms == nil {
		realtimeData.ActiveAlarms = &card.ActiveAlarmState{
			Timestamp: alarmTimestamp,
		}
	}

	// 【比对 1】检查 timestamp：只有 alarm.timestamp > existing.timestamp 才更新
	if alarmTimestamp <= realtimeData.ActiveAlarms.Timestamp {
		s.logger.Debug("Active alarm not updated: timestamp not newer",
			zap.Int64("new_ts", alarmTimestamp),
			zap.Int64("existing_ts", realtimeData.ActiveAlarms.Timestamp),
			zap.String("alarm_level", alarmLevel))
		return
	}

	// 【比对 2】检查 alarm_level：防止低级别报警覆盖高级别报警
	// 从 NowAlarm 解析现有的最高级别
	highestLevelStr := ""
	if realtimeData.ActiveAlarms.NowAlarm != "" {
		// NowAlarm 格式为 "AlarmLevel.AlarmType"，如 "EMERG.Fall"
		parts := splitAlarmCategory(realtimeData.ActiveAlarms.NowAlarm)
		if len(parts) >= 1 {
			highestLevelStr = parts[0]
		}
	}

	newPriority := alarmLevelPriority[alarmLevel]
	existingPriority := alarmLevelPriority[highestLevelStr]

	// 如果新报警级别不如现有的高（优先级数字更大），不更新
	if highestLevelStr != "" && newPriority > existingPriority {
		s.logger.Debug("Active alarm not updated: lower alarm level",
			zap.String("new_level", alarmLevel),
			zap.String("existing_level", highestLevelStr),
			zap.Int("new_priority", newPriority),
			zap.Int("existing_priority", existingPriority))
		return
	}

	// ✓ 通过双重比对，更新 timestamp、计数和 NowAlarm
	realtimeData.ActiveAlarms.Timestamp = alarmTimestamp

	// 更新对应级别的计数（增加 1）
	switch alarmLevel {
	case "EMERG":
		realtimeData.ActiveAlarms.EMERG++
	case "ALERT":
		realtimeData.ActiveAlarms.ALERT++
	case "CRIT":
		realtimeData.ActiveAlarms.CRIT++
	case "ERR":
		realtimeData.ActiveAlarms.ERR++
	case "WARNING":
		realtimeData.ActiveAlarms.WARNING++
	case "NOTICE":
		realtimeData.ActiveAlarms.NOTICE++
	}

	// 更新 NowAlarm（保留当前最高级别的报警）
	realtimeData.ActiveAlarms.NowAlarm = alarmLevel + ".Alarm"

	s.logger.Info("Active alarm updated",
		zap.String("alarm_level", alarmLevel),
		zap.Int64("timestamp", alarmTimestamp),
		zap.String("now_alarm", realtimeData.ActiveAlarms.NowAlarm))
}

// ========== 辅助函数 ==========

// splitAlarmCategory 分割 "AlarmLevel.AlarmType" 格式
func splitAlarmCategory(category string) []string {
	// 使用 "." 分割，返回 [AlarmLevel, AlarmType]
	parts := strings.Split(category, ".")
	return parts
}

// removeDeviceVitalFromRealtime 从 RealtimeData 中移除指定设备的 vital 数据（与 MonitorService 共用）
func (s *EventAlarmService) removeDeviceVitalFromRealtime(realtimeData *card.RealtimeData, deviceID string) {
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
