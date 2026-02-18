package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"owl-common/alarm"
	"owl-common/card"
	"owl-common/redis"
	"wisefido-cardagg/internal/publisher"
	"wisefido-cardagg/internal/repository"

	"go.uber.org/zap"
)

// EventAlarmService event/alarm 消息处理服务
type EventAlarmService struct {
	cacheRepo     repository.CacheRepository
	cardPublisher *publisher.CardStreamPublisher
	db            *sql.DB // 数据库连接，用于查询准确的报警计数
	logger        *zap.Logger
}

// NewEventAlarmService 创建 event/alarm 服务
// db 可为 nil，此时使用 Redis 缓存的计数（降级方案）
func NewEventAlarmService(repo repository.CacheRepository, cardPublisher *publisher.CardStreamPublisher, db *sql.DB, logger *zap.Logger) *EventAlarmService {
	return &EventAlarmService{
		cacheRepo:     repo,
		cardPublisher: cardPublisher,
		db:            db,
		logger:        logger,
	}
}

// SyncAllCardsAlarmState 启动时从 DB 同步所有 card 的 alarm_state 到 Redis
// 防止 Redis 缓存与 DB 不一致（如之前消息被丢弃导致 Redis 未更新）
func (s *EventAlarmService) SyncAllCardsAlarmState(ctx context.Context) error {
	if s.db == nil {
		s.logger.Warn("DB not available, skipping alarm state sync")
		return nil
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT card_id, tenant_id,
		       unhandled_alarm_0, unhandled_alarm_1, unhandled_alarm_2,
		       unhandled_alarm_3, unhandled_alarm_4,
		       pop_alarm_level, pop_alarm_type, pop_alarm_event_id
		FROM cards
	`)
	if err != nil {
		return fmt.Errorf("query cards: %w", err)
	}
	defer rows.Close()

	synced := 0
	for rows.Next() {
		var cardID, tenantID string
		var state card.CardAlarmState
		var popEventId sql.NullString
		if err := rows.Scan(&cardID, &tenantID,
			&state.UnhandledAlarm0, &state.UnhandledAlarm1, &state.UnhandledAlarm2,
			&state.UnhandledAlarm3, &state.UnhandledAlarm4,
			&state.PopAlarmLevel, &state.PopAlarmType, &popEventId,
		); err != nil {
			s.logger.Warn("Scan card row failed", zap.Error(err))
			continue
		}
		if popEventId.Valid {
			state.PopAlarmEventId = popEventId.String
		}

		alarmState := state.ToAlarmState()

		cardStatus, _ := s.cacheRepo.GetCardStatus(ctx, cardID)
		if cardStatus == nil {
			cardStatus = &card.CardStatus{
				CardID:    cardID,
				Timestamp: time.Now().Unix(),
			}
		}

		cardStatus.AlarmState = alarmState
		cardStatus.UpdateType = "AlarmState"

		if err := s.cacheRepo.SetCardStatus(ctx, cardStatus); err != nil {
			s.logger.Warn("Sync card alarm state to Redis failed",
				zap.String("card_id", cardID), zap.Error(err))
			continue
		}

		if s.cardPublisher != nil {
			go s.cardPublisher.PublishCardStatus(ctx, tenantID, cardID, cardStatus)
		}

		synced++
	}

	s.logger.Info("Synced all cards alarm state from DB to Redis",
		zap.Int("synced_count", synced))
	return nil
}

// HandleAlarmProcessMessage 处理报警处理消息（来自 config:alarmProcess:stream）
// 直接用消息中的字段更新 Redis AlarmState，无需查 DB
// alarmLevel: "EMERG"/"ALERT"/"CRIT"/"ERR"/"WARNING"
// alarmType: "Fall"/"AbnormalHeartRate" 等
// eventID: 被处理的 alarm_events.event_id
// alarmTimestamp: hand_time（秒）
func (s *EventAlarmService) HandleAlarmProcessMessage(ctx context.Context, cardID, tenantID, alarmLevel, alarmType, eventID string, alarmTimestamp int64) error {
	// 获取当前 Redis CardStatus
	cardStatus, err := s.cacheRepo.GetCardStatus(ctx, cardID)
	if err != nil || cardStatus == nil {
		cardStatus = &card.CardStatus{
			CardID:    cardID,
			Timestamp: time.Now().Unix(),
		}
	}

	alarmState := cardStatus.AlarmState

	// 按 alarmLevel 递减对应计数
	switch alarmLevel {
	case "EMERG", "0":
		if alarmState.ActiveEmerg > 0 {
			alarmState.ActiveEmerg--
		}
	case "ALERT", "1":
		if alarmState.ActiveAlert > 0 {
			alarmState.ActiveAlert--
		}
	case "CRIT", "2":
		if alarmState.ActiveCrit > 0 {
			alarmState.ActiveCrit--
		}
	case "ERR", "3":
		if alarmState.ActiveErr > 0 {
			alarmState.ActiveErr--
		}
	case "WARNING", "4":
		if alarmState.ActiveWarning > 0 {
			alarmState.ActiveWarning--
		}
	}

	// 如果处理的是当前弹出报警，清除 pop_alarm
	if eventID != "" && eventID == alarmState.EventID {
		alarmState.PopAlarm = ""
		alarmState.EventID = ""
	}

	now := time.Now()
	alarmState.UpdatedAt = now.UnixMilli()

	cardStatus.AlarmState = alarmState
	cardStatus.UpdateType = "AlarmState"
	cardStatus.Timestamp = now.Unix()
	cardStatus.Message = nil // 清除旧的 alarm message

	if err := s.cacheRepo.SetCardStatus(ctx, cardStatus); err != nil {
		s.logger.Warn("Failed to set card status after alarm process",
			zap.String("card_id", cardID),
			zap.Error(err))
		return err
	}

	// 推送 card:cardStatus:stream，供 wisefido-data data-subscriber 取出放入 cache，为客户 SSE 提供数据
	if s.cardPublisher != nil {
		go s.cardPublisher.PublishCardStatus(ctx, tenantID, cardID, cardStatus)
	}

	s.logger.Info("Alarm process → Redis update + card:cardStatus:stream",
		zap.String("card_id", cardID),
		zap.String("event_id", eventID),
		zap.String("alarm_level", alarmLevel),
		zap.String("alarm_type", alarmType),
		zap.String("pop_alarm", alarmState.PopAlarm),
		zap.Int("active_emerg", alarmState.ActiveEmerg))

	return nil
}

// HandleMessage 处理 event/alarm 消息
// 【入口处统一时序检查】使用 CardStatus 的时间戳进行检查
func (s *EventAlarmService) HandleMessage(ctx context.Context, msg *redis.IoTStreamMessage) error {
	// 【时序检查 - 统一入口】检查 CardStatus 时间戳
	// 获取现有的 CardStatus 用于时间戳比较
	cardStatus, err := s.cacheRepo.GetCardStatus(ctx, msg.CardID)
	if err == nil && cardStatus != nil {
		// 如果 msg.Timestamp + 5s <= cardStatus.Timestamp，说明消息太旧，直接拒绝
		if msg.Timestamp+5 < cardStatus.Timestamp {
			s.logger.Info("Event/alarm message dropped: outside 5s window",
				zap.String("card_id", msg.CardID),
				zap.String("topic_type", msg.TopicType),
				zap.Int64("msg_ts", msg.Timestamp),
				zap.Int64("status_ts", cardStatus.Timestamp),
				zap.Int64("window_end", msg.Timestamp+5))
			return nil
		}
	}

	// ✓ 通过时间窗口检查，分流处理 event/alarm/status
	switch msg.TopicType {
	case "event":
		return s.handleEvent(ctx, msg)
	case "alarm":
		return s.handleAlarm(ctx, msg)
	case "status":
		return s.handleDeviceStatus(ctx, msg)
	default:
		s.logger.Warn("Unknown topic_type", zap.String("topic_type", msg.TopicType))
		return nil
	}
}

// handleEvent 处理 event 类消息
// DataValue 为数组，每项为事件对象，项内含 category 字段
func (s *EventAlarmService) handleEvent(ctx context.Context, msg *redis.IoTStreamMessage) error {
	// 获取或创建 CardStatus
	cardStatus, err := s.cacheRepo.GetCardStatus(ctx, msg.CardID)
	if err != nil {
		s.logger.Warn("Failed to get card status", zap.Error(err))
		cardStatus = &card.CardStatus{
			CardID:    msg.CardID,
			Timestamp: msg.Timestamp,
		}
	}

	if cardStatus == nil {
		cardStatus = &card.CardStatus{
			CardID:    msg.CardID,
			Timestamp: msg.Timestamp,
		}
	}

	// 遍历 data_value 数组处理每条事件
	handled := false
	for _, item := range msg.DataValue {
		dataItem, ok := item.(map[string]interface{})
		if !ok {
			s.logger.Debug("Invalid event data item type", zap.Any("item", item))
			continue
		}

		category, _ := dataItem["category"].(string)

		switch category {
		// device 直接发出的进出事件
		case "InBed", "OutBed", "EnterMonitor", "ExitMonitor":
			s.handleDeviceBedStateEvent(ctx, msg, dataItem, cardStatus)
			handled = true
			continue
		case "BedState":
			s.handleBedStateEvent(ctx, msg, dataItem, cardStatus)
			handled = true
		case "RoomState":
			s.handleRoomStateEvent(ctx, msg, dataItem, cardStatus)
			handled = true
		default:
			s.logger.Debug("Unknown event category, skipping",
				zap.String("category", category),
				zap.String("card_id", msg.CardID))
		}
	}

	if !handled {
		return nil
	}

	// 更新状态数据到 Redis（TTL 12H）
	cardStatus.UpdateType = "EventState"
	if msg.Timestamp > cardStatus.Timestamp {
		cardStatus.Timestamp = msg.Timestamp
	}
	if err := s.cacheRepo.SetCardStatus(ctx, cardStatus); err != nil {
		s.logger.Warn("Failed to set card status", zap.Error(err))
		return err
	}

	// 异步发布到 card:status:stream（非阻塞）
	if s.cardPublisher != nil {
		go s.cardPublisher.PublishCardStatus(ctx, msg.TenantID, msg.CardID, cardStatus)
	}

	s.logger.Info("Event message processed",
		zap.String("card_id", msg.CardID),
		zap.String("device_id", msg.DeviceID),
		zap.String("category", msg.Category),
		zap.Int("data_count", len(msg.DataValue)))

	return nil
}

// handleAlarm 处理 alarm 类消息
// qinglan 发布格式：msg.Category = "AlarmLevel.AlarmType"（如 "WARNING.Stay", "EMERG.Fall"）
// 流程：解析 → INSERT alarm_events + UPDATE cards（事务）→ 写 Redis → 推送 card:cardStatus:stream
func (s *EventAlarmService) handleAlarm(ctx context.Context, msg *redis.IoTStreamMessage) error {
	// 从 msg.Category 解析 alarm_level 和 alarm_type
	alarmLevel := ""
	alarmType := ""
	if parts := strings.SplitN(msg.Category, ".", 2); len(parts) >= 1 {
		alarmLevel = parts[0]
		if len(parts) >= 2 {
			alarmType = parts[1]
		}
	}

	if alarmLevel == "" {
		s.logger.Warn("Alarm message missing alarm_level in category",
			zap.String("card_id", msg.CardID),
			zap.String("category", msg.Category))
		return nil
	}

	s.logger.Info("handleAlarm: parsed",
		zap.String("card_id", msg.CardID),
		zap.String("alarm_level", alarmLevel),
		zap.String("alarm_type", alarmType),
		zap.Int64("timestamp", msg.Timestamp))

	// 【设备自恢复检测】
	// 1. alarmType == "DeviceRecovery" → 全量 auto-resolve
	// 2. StatusFieldValue == "0" → 该 alarmType 的恢复事件
	// 3. 兼容旧格式：data_value[0].recovery 以 "_recovery" 结尾
	if alarmType == "DeviceRecovery" {
		return s.handleDeviceRecoveryAlarm(ctx, msg, nil)
	}
	if isRecovery, recoveryAlarmType := s.checkRecoveryFromDataValue(msg, alarmType); isRecovery {
		return s.handleDeviceRecoveryAlarm(ctx, msg, []string{recoveryAlarmType})
	}

	// 构建 trigger_data（整个 iot:alarm:stream 消息，便于追溯）
	triggerData, _ := json.Marshal(msg)

	// 确定 FHIR category（从 data_value[0].category 取，为空时按 alarmType 推断）
	fhirCategory := ""
	if len(msg.DataValue) > 0 {
		if item, ok := msg.DataValue[0].(map[string]interface{}); ok {
			fhirCategory, _ = item["category"].(string)
		}
	}
	if fhirCategory == "" {
		fhirCategory = alarm.GetFHIRCategory(alarmType)
	}

	// 【核心】INSERT alarm_events + UPDATE cards（事务）
	if s.db == nil {
		s.logger.Warn("DB not available, alarm not persisted",
			zap.String("card_id", msg.CardID))
		return nil
	}

	triggeredAt := time.Unix(msg.Timestamp, 0)
	_, dbState, err := card.InsertAlarmAndUpdateCard(ctx, s.db, msg.CardID, card.AlarmInsertParams{
		TenantID:    msg.TenantID,
		DeviceID:    msg.DeviceID,
		EventType:   alarmType,
		Category:    fhirCategory,
		AlarmLevel:  alarmLevel,
		TriggeredAt: triggeredAt,
		TriggerData: triggerData,
	})
	if err != nil {
		s.logger.Warn("InsertAlarmAndUpdateCard failed",
			zap.String("card_id", msg.CardID),
			zap.Error(err))
		return err
	}

	// 构建 AlarmState 从 DB 结果
	alarmState := dbState.ToAlarmState()

	// 获取或创建 CardStatus → 写 Redis
	cardStatus, err := s.cacheRepo.GetCardStatus(ctx, msg.CardID)
	if err != nil || cardStatus == nil {
		cardStatus = &card.CardStatus{
			CardID:    msg.CardID,
			Timestamp: msg.Timestamp,
		}
	}

	cardStatus.AlarmState = alarmState
	cardStatus.UpdateType = "AlarmState"
	cardStatus.Message = map[string]interface{}{
		"device_id":   msg.DeviceID,
		"device_type": msg.DeviceType,
		"card_id":     msg.CardID,
		"tenant_id":   msg.TenantID,
		"timestamp":   msg.Timestamp,
		"topic_type":  msg.TopicType,
		"category":    msg.Category,
		"data_value":  msg.DataValue,
	}

	if msg.Timestamp > cardStatus.Timestamp {
		cardStatus.Timestamp = msg.Timestamp
	}
	if err := s.cacheRepo.SetCardStatus(ctx, cardStatus); err != nil {
		s.logger.Warn("Failed to set card status", zap.Error(err))
		return err
	}

	// 异步发布到 card:status:stream
	if s.cardPublisher != nil {
		go s.cardPublisher.PublishCardStatus(ctx, msg.TenantID, msg.CardID, cardStatus)
	}

	s.logger.Info("Alarm → DB + Redis + card:cardStatus:stream",
		zap.String("card_id", msg.CardID),
		zap.String("alarm_level", alarmLevel),
		zap.String("alarm_type", alarmType),
		zap.String("pop_alarm", cardStatus.AlarmState.PopAlarm),
		zap.Int64("triggered_at", msg.Timestamp))

	return nil
}

// checkRecoveryFromDataValue 检查设备恢复事件
// 优先检查 StatusFieldValue=="0"（新格式），兼容 recovery 含 "_recovery"（旧格式）
func (s *EventAlarmService) checkRecoveryFromDataValue(msg *redis.IoTStreamMessage, alarmType string) (bool, string) {
	if len(msg.DataValue) == 0 {
		return false, ""
	}
	item, ok := msg.DataValue[0].(map[string]interface{})
	if !ok {
		return false, ""
	}
	// 新格式：StatusFieldValue == "0" 表示恢复
	if sfv, ok := item["StatusFieldValue"].(string); ok && sfv == "0" {
		return true, alarmType
	}
	// 旧格式兼容
	if recovery, ok := item["recovery"].(string); ok && strings.HasSuffix(recovery, "_recovery") {
		return true, alarmType
	}
	return false, ""
}

// handleDeviceRecoveryAlarm 处理设备恢复报警
// resolveTypes == nil → resolve 所有设备级报警（DeviceRecovery 场景）
// resolveTypes != nil → 只 resolve 指定的报警类型（如 SignalPoor recovery）
func (s *EventAlarmService) handleDeviceRecoveryAlarm(ctx context.Context, msg *redis.IoTStreamMessage, resolveTypes []string) error {
	if s.db == nil {
		s.logger.Warn("DB not available for device recovery",
			zap.String("card_id", msg.CardID))
		return nil
	}

	dbState, result, err := card.AutoResolveDeviceAlarms(ctx, s.db, msg.CardID, msg.TenantID, msg.DeviceID, resolveTypes)
	if err != nil {
		s.logger.Warn("AutoResolveDeviceAlarms failed",
			zap.String("card_id", msg.CardID),
			zap.String("device_id", msg.DeviceID),
			zap.Error(err))
		return err
	}

	if result.ResolvedCount == 0 {
		s.logger.Info("Device recovery: no active alarms to resolve",
			zap.String("card_id", msg.CardID),
			zap.String("device_id", msg.DeviceID),
			zap.Any("resolve_types", resolveTypes))
		return nil
	}

	alarmState := dbState.ToAlarmState()

	// 更新 Redis + 推送
	cardStatus, err := s.cacheRepo.GetCardStatus(ctx, msg.CardID)
	if err != nil || cardStatus == nil {
		cardStatus = &card.CardStatus{
			CardID:    msg.CardID,
			Timestamp: msg.Timestamp,
		}
	}

	cardStatus.AlarmState = alarmState
	cardStatus.UpdateType = "AlarmState"
	if msg.Timestamp > cardStatus.Timestamp {
		cardStatus.Timestamp = msg.Timestamp
	}
	if err := s.cacheRepo.SetCardStatus(ctx, cardStatus); err != nil {
		s.logger.Warn("Failed to set card status after device recovery",
			zap.String("card_id", msg.CardID),
			zap.Error(err))
		return err
	}

	if s.cardPublisher != nil {
		go s.cardPublisher.PublishCardStatus(ctx, msg.TenantID, msg.CardID, cardStatus)
	}

	s.logger.Info("Device recovery → auto-resolved alarms",
		zap.String("card_id", msg.CardID),
		zap.String("device_id", msg.DeviceID),
		zap.Int("resolved_count", result.ResolvedCount),
		zap.String("top_event_id", result.TopEventId),
		zap.Any("resolve_types", resolveTypes),
		zap.String("pop_alarm", alarmState.PopAlarm))

	return nil
}

// ========== Event 类别处理 ==========

// handleDeviceBedStateEvent 处理 device 直接发出的进出事件（InBed/OutBed/EnterMonitor/ExitMonitor）
// event_alarm_handler 已做时效检查，这里直接填充
func (s *EventAlarmService) handleDeviceBedStateEvent(ctx context.Context, msg *redis.IoTStreamMessage, dataItem map[string]interface{}, cardStatus *card.CardStatus) {
	currentState, _ := dataItem["category"].(string)
	now := time.Now().Unix()

	if cardStatus.Timestamp > msg.Timestamp {
		return
	}

	durationSec := now - msg.Timestamp
	if durationSec < 0 {
		durationSec = 0
	}

	cardStatus.EventState = &card.EventState{
		UpdatedAt:    msg.Timestamp,
		Category:     "BedState",
		CurrentState: currentState,
		StateValue: "Duration",
		StartTime:    msg.Timestamp,
		DurationSec:  int(durationSec),
	}
	cardStatus.UpdateType = "EventState"

	s.logger.Info("DeviceBedState updated",
		zap.String("card_id", msg.CardID),
		zap.String("current_state", currentState),
		zap.Int64("msg_timestamp", msg.Timestamp),
		zap.Int64("duration_sec", durationSec))
}

func (s *EventAlarmService) handleBedStateEvent(ctx context.Context, msg *redis.IoTStreamMessage, dataItem map[string]interface{}, cardStatus *card.CardStatus) {
	currentState, _ := dataItem["CurrentState"].(string)
	stateSince, _ := dataItem["StateSince"].(float64)
	startTime := int64(stateSince)
	if startTime == 0 {
		startTime = msg.Timestamp
	}

	s.logger.Info("Handling BedState event",
		zap.String("card_id", msg.CardID),
		zap.String("current_state", currentState),
		zap.Int64("start_time", startTime),
		zap.Int64("msg_timestamp", msg.Timestamp))

	now := msg.Timestamp
	if startTime > now {
		startTime = now
	}

	// 时序检查：EventState 存在时，只有 StartTime 更新才覆盖
	if cardStatus.EventState != nil && startTime <= cardStatus.EventState.StartTime {
		s.logger.Debug("BedState not updated: timestamp not newer",
			zap.Int64("new_ts", startTime),
			zap.Int64("existing_ts", cardStatus.EventState.StartTime))
		return
	}

	durationSec := now - startTime
	if durationSec < 0 {
		s.logger.Warn("Calculated negative duration, setting to 0",
			zap.String("card_id", msg.CardID),
			zap.Int64("duration_sec", durationSec))
		durationSec = 0
	}

	cardStatus.EventState = &card.EventState{
		UpdatedAt:    now,
		Category:     "BedState",
		CurrentState: currentState,
		StartTime:    startTime,
		DurationSec:  int(durationSec),
	}
	cardStatus.UpdateType = "EventState"

	s.logger.Info("EventState updated (BedState)",
		zap.String("current_state", currentState),
		zap.Int64("start_time", startTime),
		zap.Int64("duration_sec", durationSec))
}

// handleRoomStateEvent 处理 RoomState 事件 → 写入 EventState
func (s *EventAlarmService) handleRoomStateEvent(ctx context.Context, msg *redis.IoTStreamMessage, dataItem map[string]interface{}, cardStatus *card.CardStatus) {
	peopleCount, _ := dataItem["PeopleCount"].(float64)
	stayTime, _ := dataItem["StayTime"].(float64)
	stateSince, _ := dataItem["StateSince"].(float64)
	startTime := int64(stateSince)
	if startTime == 0 {
		startTime = msg.Timestamp
	}

	s.logger.Info("Handling RoomState event",
		zap.String("card_id", msg.CardID),
		zap.Int("people_count", int(peopleCount)),
		zap.Int64("start_time", startTime),
		zap.Int64("msg_timestamp", msg.Timestamp))

	now := msg.Timestamp
	if startTime > now {
		startTime = now
	}

	// 时序检查
	if cardStatus.EventState != nil && startTime <= cardStatus.EventState.StartTime {
		s.logger.Debug("RoomState not updated: timestamp not newer",
			zap.Int64("new_ts", startTime),
			zap.Int64("existing_ts", cardStatus.EventState.StartTime))
		return
	}

	durationSec := int64(stayTime) * 60 // StayTime 是分钟，转秒
	if durationSec < 0 {
		s.logger.Warn("Calculated negative duration from StayTime, setting to 0",
			zap.String("card_id", msg.CardID),
			zap.Float64("stay_time", stayTime))
		durationSec = 0
	}

	now = msg.Timestamp
	cardStatus.EventState = &card.EventState{
		UpdatedAt:    now,
		Category:     "RoomState",
		CurrentState: fmt.Sprintf("%d_people", int(peopleCount)),
		StateValue:   fmt.Sprintf("%d", int(peopleCount)),
		StartTime:    startTime,
		DurationSec:  int(durationSec),
	}
	cardStatus.UpdateType = "EventState"

	s.logger.Info("EventState updated (RoomState)",
		zap.Int("people_count", int(peopleCount)),
		zap.Int64("start_time", startTime),
		zap.Int64("duration_sec", durationSec))
}

// handleDeviceStatus 处理设备状态消息 (来自 iot:DeviceStatus:stream)
// 更新卡片的 CardStatus 中的 DeviceStatus 字段
// 【时间戳比对】只在新状态的时间戳 >= 现有时间戳时才更新
func (s *EventAlarmService) handleDeviceStatus(ctx context.Context, msg *redis.IoTStreamMessage) error {
	if msg.CardID == "" {
		s.logger.Debug("DeviceStatus message without card_id, skipping")
		return nil
	}

	if msg.DeviceID == "" {
		s.logger.Debug("DeviceStatus message without device_id, skipping")
		return nil
	}

	// 获取或创建 CardStatus
	cardStatus, err := s.cacheRepo.GetCardStatus(ctx, msg.CardID)
	if err != nil {
		s.logger.Warn("Failed to get card status for device status update", zap.Error(err))
		cardStatus = &card.CardStatus{
			CardID:    msg.CardID,
			Timestamp: msg.Timestamp,
		}
	}

	if cardStatus == nil {
		cardStatus = &card.CardStatus{
			CardID:    msg.CardID,
			Timestamp: msg.Timestamp,
		}
	}

	// 初始化 DeviceStatus map 如果为 nil
	if cardStatus.DeviceStatus == nil {
		cardStatus.DeviceStatus = make(map[string]*card.DeviceStatus)
	}

	// 从 IoTStreamMessage 中提取设备状态
	devStatus := s.convertToDeviceStatus(msg)
	if devStatus == nil {
		s.logger.Debug("Failed to extract device status from message",
			zap.String("card_id", msg.CardID),
			zap.String("device_id", msg.DeviceID))
		return nil
	}

	// 【时间戳比对】检查该设备是否有旧状态
	if existingDevStatus, exists := cardStatus.DeviceStatus[msg.DeviceID]; exists {
		// 只有新数据时间戳 > 现有时间戳时，才更新
		if devStatus.Timestamp <= existingDevStatus.Timestamp {
			s.logger.Debug("Device status message dropped: older than existing",
				zap.String("card_id", msg.CardID),
				zap.String("device_id", msg.DeviceID),
				zap.Int64("new_ts", devStatus.Timestamp),
				zap.Int64("existing_ts", existingDevStatus.Timestamp))
			return nil
		}
	}

	// 按设备 ID 存储设备状态
	cardStatus.DeviceStatus[msg.DeviceID] = devStatus
	cardStatus.UpdateType = "DeviceStatus"

	// 【只在真正更新时，才更新 CardStatus 的时间戳】
	cardStatus.Timestamp = msg.Timestamp

	// 保存回 Redis
	if err := s.cacheRepo.SetCardStatus(ctx, cardStatus); err != nil {
		s.logger.Warn("Failed to set card status after device status update",
			zap.String("card_id", msg.CardID),
			zap.Error(err))
		return err
	}

	// 发布 CardStatus 到 card:status:stream（供前端实时获取）
	s.cardPublisher.PublishCardStatus(ctx, msg.TenantID, msg.CardID, cardStatus)

	s.logger.Debug("Device status updated",
		zap.String("card_id", msg.CardID),
		zap.String("device_id", msg.DeviceID),
		zap.Int64("timestamp", devStatus.Timestamp),
		zap.Any("statuses", devStatus.Statuses))

	return nil
}

// convertToDeviceStatus 从 IoTStreamMessage 中提取并转换为 DeviceStatus
func (s *EventAlarmService) convertToDeviceStatus(msg *redis.IoTStreamMessage) *card.DeviceStatus {
	if len(msg.DataValue) == 0 {
		return nil
	}

	// 从 data_value[0] 提取 statuses
	dataItem, ok := msg.DataValue[0].(map[string]interface{})
	if !ok {
		return nil
	}

	category, _ := dataItem["category"].(string)
	if category != "deviceStatus" {
		return nil
	}

	// 提取 statuses map
	statusesRaw, ok := dataItem["statuses"].(map[string]interface{})
	if !ok {
		return nil
	}

	// 转换 map[string]interface{} 为 map[string]int
	statuses := make(map[string]int)
	for k, v := range statusesRaw {
		// 处理 float64（JSON 中数字的默认类型）
		switch val := v.(type) {
		case float64:
			statuses[k] = int(val)
		case int:
			statuses[k] = val
		case int64:
			statuses[k] = int(val)
		}
	}

	return &card.DeviceStatus{
		DeviceID:   msg.DeviceID,
		DeviceType: msg.DeviceType,
		Timestamp:  msg.Timestamp,
		Statuses:   statuses,
	}
}
