package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"owl-common/alarm"
	"owl-common/card"
	"owl-common/redis"
	"wisefido-cardagg/internal/publisher"
	"wisefido-cardagg/internal/repository"

	"github.com/lib/pq"
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

// HandleAlarmProcessMessage 处理报警处理消息（来自 config:alarm.process:stream）
// 当报警被确认或解决时，更新卡片的 ActiveAlarms 状态
// 参数 data 包含：alarm_level, alarm_type, alarm_timestamp
// 【核心】：优先从数据库查询精确计数，确保幂等性和准确性
func (s *EventAlarmService) HandleAlarmProcessMessage(ctx context.Context, cardID, tenantID, alarmLevel, alarmType string, alarmTimestamp int64) error {
	// 1. 优先从数据库查询精确的报警计数
	var dbAlarmCounts *card.ActiveAlarmState
	if s.db != nil {
		var err error
		dbAlarmCounts, err = s.getCardAlarmCountsFromDB(ctx, tenantID, cardID)
		if err != nil {
			s.logger.Warn("Failed to query alarm counts from DB, falling back to Redis cache",
				zap.String("card_id", cardID),
				zap.Error(err))
			// 降级到 Redis 缓存
		}
	}

	// 2. 如果 DB 查询失败或 DB 为空，使用 Redis 缓存
	if dbAlarmCounts == nil {
		cardStatus, err := s.cacheRepo.GetCardStatus(ctx, cardID)
		if err != nil {
			s.logger.Warn("Failed to get card status for alarm process",
				zap.String("card_id", cardID),
				zap.Error(err))
			return nil
		}

		if cardStatus == nil || cardStatus.ActiveAlarms == nil {
			s.logger.Debug("No alarm data found (Redis cache)",
				zap.String("card_id", cardID),
				zap.String("alarm_level", alarmLevel))
			return nil
		}

		dbAlarmCounts = cardStatus.ActiveAlarms
	}

	// 3. 【幂等性检查】：验证该级别的计数 > 0，确保该报警真实存在
	currentCount := s.getAlarmCountFromState(dbAlarmCounts, alarmLevel)
	if currentCount <= 0 {
		s.logger.Info("Alarm already processed or does not exist (count <= 0), skipping decrement",
			zap.String("card_id", cardID),
			zap.String("alarm_level", alarmLevel),
			zap.Int("current_count", currentCount))
		return nil
	}

	// 4. 从数据库查询的计数减 1，更新到 Redis
	// 先从 Redis 获取完整的 CardStatus
	cardStatus, err := s.cacheRepo.GetCardStatus(ctx, cardID)
	if err != nil || cardStatus == nil {
		cardStatus = &card.CardStatus{
			CardID:       cardID,
			Timestamp:    time.Now().Unix(),
			ActiveAlarms: dbAlarmCounts,
		}
	} else {
		// 用 DB 查询的准确计数覆盖 Redis 中的计数
		cardStatus.ActiveAlarms = dbAlarmCounts
	}

	// 5. 减少对应级别的计数
	s.decrementAlarmCountInStatus(cardStatus, alarmLevel)

	// 6. 检查是否需要清空 NowAlarm
	if s.shouldClearNowAlarmInStatus(cardStatus, alarmLevel, alarmType, alarmTimestamp) {
		cardStatus.ActiveAlarms.NowAlarm = ""
		s.logger.Info("NowAlarm cleared",
			zap.String("card_id", cardID),
			zap.String("alarm_level", alarmLevel),
			zap.String("alarm_type", alarmType),
			zap.Int64("alarm_timestamp", alarmTimestamp))
	}

	// 7. 更新状态数据到 Redis（12 小时 TTL）
	if err := s.cacheRepo.SetCardStatus(ctx, cardStatus); err != nil {
		s.logger.Warn("Failed to set card status after alarm process",
			zap.String("card_id", cardID),
			zap.Error(err))
		return err
	}

	s.logger.Info("Alarm process message handled",
		zap.String("card_id", cardID),
		zap.String("alarm_level", alarmLevel),
		zap.String("alarm_type", alarmType))

	return nil
}

// getAlarmCountFromState 从 ActiveAlarmState 中获取指定级别的计数
func (s *EventAlarmService) getAlarmCountFromState(state *card.ActiveAlarmState, alarmLevel string) int {
	if state == nil {
		return 0
	}

	switch alarmLevel {
	case "EMERG":
		return state.ActiveEmerg
	case "ALERT":
		return state.ActiveAlert
	case "CRIT":
		return state.ActiveCrit
	case "ERR":
		return state.ActiveErr
	case "WARNING":
		return state.ActiveWarning
	default:
		return 0
	}
}

// decrementAlarmCountInStatus 减少 CardStatus 中指定级别的报警计数（确保 >= 0）
func (s *EventAlarmService) decrementAlarmCountInStatus(cardStatus *card.CardStatus, alarmLevel string) {
	if cardStatus == nil || cardStatus.ActiveAlarms == nil {
		return
	}

	switch alarmLevel {
	case "EMERG":
		if cardStatus.ActiveAlarms.ActiveEmerg > 0 {
			cardStatus.ActiveAlarms.ActiveEmerg--
		}
	case "ALERT":
		if cardStatus.ActiveAlarms.ActiveAlert > 0 {
			cardStatus.ActiveAlarms.ActiveAlert--
		}
	case "CRIT":
		if cardStatus.ActiveAlarms.ActiveCrit > 0 {
			cardStatus.ActiveAlarms.ActiveCrit--
		}
	case "ERR":
		if cardStatus.ActiveAlarms.ActiveErr > 0 {
			cardStatus.ActiveAlarms.ActiveErr--
		}
	case "WARNING":
		if cardStatus.ActiveAlarms.ActiveWarning > 0 {
			cardStatus.ActiveAlarms.ActiveWarning--
		}
	}

	s.logger.Debug("Alarm count decremented in status",
		zap.String("alarm_level", alarmLevel),
		zap.Int("emerg", cardStatus.ActiveAlarms.ActiveEmerg),
		zap.Int("alert", cardStatus.ActiveAlarms.ActiveAlert),
		zap.Int("crit", cardStatus.ActiveAlarms.ActiveCrit),
		zap.Int("err", cardStatus.ActiveAlarms.ActiveErr),
		zap.Int("warning", cardStatus.ActiveAlarms.ActiveWarning))
}

// getCardAlarmCountsFromDB 从数据库查询卡片的精确报警计数
// 查询 alarm_events 中 status='active' 的记录，按 alarm_level 统计
func (s *EventAlarmService) getCardAlarmCountsFromDB(ctx context.Context, tenantID, cardID string) (*card.ActiveAlarmState, error) {
	if s.db == nil {
		return nil, nil // DB 为空，返回 nil，使用 Redis 缓存
	}

	// 1. 查询卡片关联的所有 device_id（从 cards.devices JSONB 中提取）
	query := `
		SELECT devices
		FROM cards
		WHERE card_id = $1 AND tenant_id = $2
		LIMIT 1
	`

	var devicesJSON sql.NullString
	err := s.db.QueryRowContext(ctx, query, cardID, tenantID).Scan(&devicesJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.Debug("Card not found in database", zap.String("card_id", cardID))
			return &card.ActiveAlarmState{}, nil
		}
		s.logger.Warn("Failed to query card devices", zap.String("card_id", cardID), zap.Error(err))
		return nil, err
	}

	if !devicesJSON.Valid || devicesJSON.String == "" {
		s.logger.Debug("Card has no devices", zap.String("card_id", cardID))
		return &card.ActiveAlarmState{}, nil
	}

	// 2. 解析 devices JSON
	var devices []map[string]interface{}
	if err := json.Unmarshal([]byte(devicesJSON.String), &devices); err != nil {
		s.logger.Warn("Failed to unmarshal devices JSON", zap.String("card_id", cardID), zap.Error(err))
		return &card.ActiveAlarmState{}, nil
	}

	// 3. 提取 device_id 列表
	deviceIDs := make([]string, 0, len(devices))
	for _, device := range devices {
		if deviceID, ok := device["device_id"].(string); ok && deviceID != "" {
			deviceIDs = append(deviceIDs, deviceID)
		}
	}

	if len(deviceIDs) == 0 {
		s.logger.Debug("Card has no valid device IDs", zap.String("card_id", cardID))
		return &card.ActiveAlarmState{}, nil
	}

	// 4. 统计该卡片关联的所有设备的未处理报警（alarm_status = 'active'）
	// 按 alarm_level 分组统计
	countQuery := `
		SELECT 
			COUNT(*) FILTER (WHERE alarm_level IN ('0', 'EMERG')) as count_0,
			COUNT(*) FILTER (WHERE alarm_level IN ('1', 'ALERT')) as count_1,
			COUNT(*) FILTER (WHERE alarm_level IN ('2', 'CRIT')) as count_2,
			COUNT(*) FILTER (WHERE alarm_level IN ('3', 'ERR')) as count_3,
			COUNT(*) FILTER (WHERE alarm_level IN ('4', 'WARNING')) as count_4,
    	FROM alarm_events
		WHERE tenant_id = $1
		  AND device_id = ANY($2::uuid[])
		  AND alarm_status = 'active'
		  AND (metadata->>'deleted_at' IS NULL)
	`

	var count0, count1, count2, count3, count4, count5 int
	err = s.db.QueryRowContext(ctx, countQuery, tenantID, pq.Array(deviceIDs)).Scan(
		&count0, &count1, &count2, &count3, &count4, &count5,
	)
	if err != nil {
		s.logger.Warn("Failed to count alarm events from DB", zap.String("card_id", cardID), zap.Error(err))
		return nil, err
	}

	return &card.ActiveAlarmState{
		ActiveEmerg:   count0,
		ActiveAlert:   count1,
		ActiveCrit:    count2,
		ActiveErr:     count3,
		ActiveWarning: count4,
	}, nil
}

// shouldClearNowAlarmInStatus 判断 CardStatus 中是否应该清空 NowAlarm
// 条件：alarmType, alarmLevel, alarmTimestamp 与当前 NowAlarm 一致
func (s *EventAlarmService) shouldClearNowAlarmInStatus(cardStatus *card.CardStatus, alarmLevel, alarmType string, alarmTimestamp int64) bool {
	if cardStatus == nil || cardStatus.ActiveAlarms == nil || cardStatus.ActiveAlarms.NowAlarm == "" {
		return false
	}

	// NowAlarm 格式为 "AlarmLevel.AlarmType"，如 "EMERG.Fall"
	parts := strings.Split(cardStatus.ActiveAlarms.NowAlarm, ".")
	if len(parts) < 1 {
		return false
	}

	nowAlarmLevel := parts[0]

	// 检查级别是否相同
	if nowAlarmLevel != alarmLevel {
		return false
	}

	// 检查时间戳是否相同（精确匹配，防止清空其他报警）
	if cardStatus.ActiveAlarms.Timestamp != alarmTimestamp {
		s.logger.Debug("NowAlarm timestamp mismatch in status, not clearing",
			zap.String("now_alarm", cardStatus.ActiveAlarms.NowAlarm),
			zap.Int64("existing_ts", cardStatus.ActiveAlarms.Timestamp),
			zap.Int64("new_ts", alarmTimestamp))
		return false
	}

	return true
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
		case "BedState":
			s.handleBedStateEvent(ctx, msg, dataItem, cardStatus)
		case "RoomState":
			s.handleRoomStateEvent(ctx, msg, dataItem, cardStatus)
		default:
			s.logger.Debug("Unknown event category",
				zap.String("category", category),
				zap.String("card_id", msg.CardID))
		}
	}

	// 更新状态数据到 Redis（TTL 12H）
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
// DataValue 为数组，每项为报警对象，项内含 category 字段
func (s *EventAlarmService) handleAlarm(ctx context.Context, msg *redis.IoTStreamMessage) error {
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
		s.updateActiveAlarmInStatus(cardStatus, alarmLevel, int64(alarmTimestamp))
	}

	// 更新状态数据到 Redis（TTL 12H）
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

	s.logger.Info("Alarm message processed",
		zap.String("card_id", msg.CardID),
		zap.String("device_id", msg.DeviceID),
		zap.String("category", msg.Category),
		zap.Int("data_count", len(msg.DataValue)))

	return nil
}

// ========== Event 类别处理 ==========

// handleBedStateEvent 处理 BedState 事件
// dataItem 包含 {"category": "BedState", "bed_id": "uuid-bbb", "CurrentState": "in_bed", "StateSince": 1234567890}
// 检查：初始状态为空 → 直接填充；有值 → 检查 Timestamp > 原记录才更新
func (s *EventAlarmService) handleBedStateEvent(ctx context.Context, msg *redis.IoTStreamMessage, dataItem map[string]interface{}, cardStatus *card.CardStatus) {
	bedID, _ := dataItem["bed_id"].(string)
	currentState, _ := dataItem["CurrentState"].(string)
	stateSince, _ := dataItem["StateSince"].(float64)

	s.logger.Info("Handling BedState event",
		zap.String("card_id", msg.CardID),
		zap.String("bed_id", bedID),
		zap.String("current_state", currentState),
		zap.Int64("state_since", int64(stateSince)))

	// 检查初始状态是否为空
	if cardStatus.BedState == nil {
		// 初始状态为空，直接填充
		cardStatus.BedState = &card.BedState{
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
	// msg.Timestamp + 5s > cardStatus.Timestamp 已在 handler 层检查通过
	// 现在检查设备级别的时间戳
	if stateSince <= float64(cardStatus.BedState.Timestamp) {
		s.logger.Debug("BedState not updated: device timestamp not newer",
			zap.Float64("new_ts", stateSince),
			zap.Int64("existing_ts", cardStatus.BedState.Timestamp))
		return
	}

	// 有值且时间戳更新，直接更新
	cardStatus.BedState.BedID = bedID
	cardStatus.BedState.CurrentState = currentState
	cardStatus.BedState.Timestamp = int64(stateSince)
	s.logger.Info("BedState updated",
		zap.String("bed_id", bedID),
		zap.String("current_state", currentState),
		zap.Int64("state_since", int64(stateSince)))
}

// handleRoomStateEvent 处理 RoomState 事件
// dataItem 包含 {"category": "RoomState", "room_id": "uuid-rrr", "room_name": "101", "PeopleCount": 1, "StayTime": 5, "StateSince": 1234567890}
// 检查：初始状态为空 → 直接填充；有值 → 检查 Timestamp > 原记录才更新
func (s *EventAlarmService) handleRoomStateEvent(ctx context.Context, msg *redis.IoTStreamMessage, dataItem map[string]interface{}, cardStatus *card.CardStatus) {
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
	if cardStatus.RoomState == nil {
		// 初始状态为空，直接填充
		cardStatus.RoomState = &card.RoomState{
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
	// msg.Timestamp + 5s > cardStatus.Timestamp 已在 handler 层检查通过
	// 现在检查设备级别的时间戳
	if stateSince <= float64(cardStatus.RoomState.Timestamp) {
		s.logger.Debug("RoomState not updated: device timestamp not newer",
			zap.Float64("new_ts", stateSince),
			zap.Int64("existing_ts", cardStatus.RoomState.Timestamp))
		return
	}

	// 有值且时间戳更新，直接更新
	cardStatus.RoomState.RoomID = roomID
	cardStatus.RoomState.RoomName = roomName
	cardStatus.RoomState.PeopleCount = int(peopleCount)
	cardStatus.RoomState.StayTime = int(stayTime)
	cardStatus.RoomState.Timestamp = int64(stateSince)
	s.logger.Info("RoomState updated",
		zap.String("room_id", roomID),
		zap.String("room_name", roomName),
		zap.Int("people_count", int(peopleCount)),
		zap.Int64("state_since", int64(stateSince)))
}

// ========== Alarm 类别处理 ==========

// updateActiveAlarmInStatus 更新 CardStatus 中的 ActiveAlarms：双重比对策略
// 比对 1：timestamp 比较（防止旧数据覆盖新数据）
// 比对 2：alarm_level 比较（防止低级别报警覆盖高级别报警）
func (s *EventAlarmService) updateActiveAlarmInStatus(cardStatus *card.CardStatus, alarmLevel string, alarmTimestamp int64) {
	if cardStatus == nil {
		return
	}

	// 初始化 ActiveAlarms（如果为空）
	if cardStatus.ActiveAlarms == nil {
		cardStatus.ActiveAlarms = &card.ActiveAlarmState{
			Timestamp: alarmTimestamp,
		}
	}

	// 【比对 1】检查 timestamp：只有 alarm.timestamp > existing.timestamp 才更新
	if alarmTimestamp <= cardStatus.ActiveAlarms.Timestamp {
		s.logger.Debug("Active alarm not updated in status: timestamp not newer",
			zap.Int64("new_ts", alarmTimestamp),
			zap.Int64("existing_ts", cardStatus.ActiveAlarms.Timestamp),
			zap.String("alarm_level", alarmLevel))
		return
	}

	// 【比对 2】检查 alarm_level：防止低级别报警覆盖高级别报警
	highestLevelStr := ""
	if cardStatus.ActiveAlarms.NowAlarm != "" {
		parts := splitAlarmCategory(cardStatus.ActiveAlarms.NowAlarm)
		if len(parts) >= 1 {
			highestLevelStr = parts[0]
		}
	}

	newPriority := alarm.AlarmLevelPriority[alarmLevel]
	existingPriority := alarm.AlarmLevelPriority[highestLevelStr]

	if highestLevelStr != "" && newPriority > existingPriority {
		s.logger.Debug("Active alarm not updated in status: lower alarm level",
			zap.String("new_level", alarmLevel),
			zap.String("existing_level", highestLevelStr),
			zap.Int("new_priority", newPriority),
			zap.Int("existing_priority", existingPriority))
		return
	}

	// ✓ 通过双重比对，更新 timestamp、计数和 NowAlarm
	cardStatus.ActiveAlarms.Timestamp = alarmTimestamp

	// 更新对应级别的计数（增加 1）
	switch alarmLevel {
	case "EMERG":
		cardStatus.ActiveAlarms.ActiveEmerg++
	case "ALERT":
		cardStatus.ActiveAlarms.ActiveAlert++
	case "CRIT":
		cardStatus.ActiveAlarms.ActiveCrit++
	case "ERR":
		cardStatus.ActiveAlarms.ActiveErr++
	case "WARNING":
		cardStatus.ActiveAlarms.ActiveWarning++
	}

	// 更新 NowAlarm（保留当前最高级别的报警）
	cardStatus.ActiveAlarms.NowAlarm = alarmLevel + ".Alarm"

	s.logger.Info("Active alarm updated in status",
		zap.String("alarm_level", alarmLevel),
		zap.Int64("timestamp", alarmTimestamp),
		zap.String("now_alarm", cardStatus.ActiveAlarms.NowAlarm))
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

// ========== 辅助函数 ===========

// splitAlarmCategory 分割 "AlarmLevel.AlarmType" 格式
func splitAlarmCategory(category string) []string {
	// 使用 "." 分割，返回 [AlarmLevel, AlarmType]
	parts := strings.Split(category, ".")
	return parts
}
