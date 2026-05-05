package evaluator

import (
	"context"
	"fmt"
	"time"
	"wisefido-sensor/internal/consumer"
	"wisefido-sensor/internal/models"
	"wisefido-sensor/internal/repository"

	"go.uber.org/zap"
)

// Event3Evaluator 事件3：Bathroom可疑跌倒检测评估器
type Event3Evaluator struct {
	evaluator *Evaluator
}

// NewEvent3Evaluator 创建事件3评估器
func NewEvent3Evaluator(evaluator *Evaluator) *Event3Evaluator {
	return &Event3Evaluator{
		evaluator: evaluator,
	}
}

// Evaluate 评估事件3：Bathroom可疑跌倒检测
//
// 完整流程：
// 1. 检查是否是 bathroom 房间
// 2. 检测进入事件（仅1个 track_id 时）
// 3. 持续监测（每10秒轮询一次）
//    - 检查是否仍在 bathroom
//    - 检查 track_id 数量（中途有新 track_id 进入时退出）
//    - 检查人员数量（必须 == 1）
//    - 检查姿态（必须是站立）
//    - 检查位置变化（>= 10cm 重置计时，< 10cm 继续计时）
//    - 检查时间阈值（位置不动 >= 30分钟 → Fall，站姿 >= 10分钟 → SuspectedFall）
//
// 数据源：仅使用Redis实时数据（RealtimeData），不查询数据库
func (e *Event3Evaluator) Evaluate(tenantID string, card repository.CardInfo, realtimeData *models.RealtimeData) ([]models.AlarmEvent, error) {
	ctx := context.Background()
	var alarms []models.AlarmEvent

	// 1. 检查是否是 bathroom 房间
	isBathroom, err := e.checkBathroom(tenantID, card)
	if err != nil {
		return nil, err
	}
	if !isBathroom {
		return nil, nil
	}

	// 2. 检查人员数量（必须 == 1）
	if realtimeData.PersonCount != 1 {
		return nil, nil
	}

	// 3. 检查 track_id 数量（必须 == 1）
	if len(realtimeData.Postures) != 1 {
		// 如果之前有状态，清除（中途有新的 track_id 进入）
		if len(realtimeData.Postures) > 1 {
			// 清除所有可能的状态
			for _, posture := range realtimeData.Postures {
				e.clearState(ctx, card.CardID, posture.TrackingID)
			}
		}
		return nil, nil
	}

	// 4. 选择要跟踪的 track_id（仅1个）
	trackID := selectTrackIDInBathroom(realtimeData)
	if trackID == "" {
		return nil, nil
	}

	// 5. 获取或初始化状态
	state, err := e.getOrInitState(ctx, card.CardID, trackID, realtimeData)
	if err != nil {
		e.evaluator.logger.Error("Failed to get or init state",
			zap.String("card_id", card.CardID),
			zap.String("track_id", trackID),
			zap.Error(err),
		)
		return nil, nil
	}

	// 6. 检查是否已经触发过报警（避免重复报警）
	if state.AlarmTriggered {
		return nil, nil
	}

	// 7. 找到当前 track_id 的 posture
	currentPosture, found := findTrackInBathroom(realtimeData, trackID)
	if !found {
		// track_id 消失，清除状态
		e.clearState(ctx, card.CardID, trackID)
		return nil, nil
	}

	// 8. 检查姿态（必须是站立）
	if !checkStandingPosture(currentPosture) {
		// 不是站立状态（可能是坐下，正常行为），清除状态
		e.clearState(ctx, card.CardID, trackID)
		return nil, nil
	}

	// 9. 检查位置变化
	positionChange, exceeded := checkPositionChange(state, currentPosture, 10.0) // 阈值 10cm
	state.PositionChange = positionChange

	if exceeded {
		// 位置变化 >= 10cm，重置计时（人在移动，继续监测）
		resetTimers(state)
		updatePositionInEvent3State(state, currentPosture)
		e.setEvent3State(ctx, card.CardID, state)
		return nil, nil
	}

	// 10. 更新位置信息（位置变化 < 10cm）
	updatePositionInEvent3State(state, currentPosture)

	// 11. 更新站姿时间（如果之前不是站立，现在变成站立）
	if state.StandingTime == nil {
		now := time.Now().Unix()
		state.StandingTime = &now
	}

	// 12. 计算持续时间
	standingDuration, stillDuration := calculateDurations(state)
	state.StandingDuration = standingDuration
	state.StillDuration = stillDuration

	// 13. 检查时间阈值
	// 优先级：位置不动 >= 30分钟 → Fall (ALERT)（更严重）
	if stillDuration >= 1800 { // 30分钟 = 1800秒
		// 触发 Fall 报警
		alarm, err := e.triggerFallAlarm(ctx, tenantID, card, realtimeData, state, currentPosture, stillDuration)
		if err != nil {
			return nil, err
		}
		alarms = append(alarms, *alarm)
		state.AlarmTriggered = true
		e.setEvent3State(ctx, card.CardID, state)
		return alarms, nil
	}

	// 站姿 >= 10分钟 → SuspectedFall (WARNING)
	if standingDuration >= 600 { // 10分钟 = 600秒
		// 触发 SuspectedFall 报警
		alarm, err := e.triggerSuspectedFallAlarm(ctx, tenantID, card, realtimeData, state, currentPosture, standingDuration)
		if err != nil {
			return nil, err
		}
		alarms = append(alarms, *alarm)
		state.AlarmTriggered = true
		e.setEvent3State(ctx, card.CardID, state)
		return alarms, nil
	}

	// 14. 保存状态
	if err := e.setEvent3State(ctx, card.CardID, state); err != nil {
		e.evaluator.logger.Error("Failed to save state",
			zap.String("card_id", card.CardID),
			zap.Error(err),
		)
	}

	return alarms, nil
}

// checkBathroom 检查房间是否是 bathroom
func (e *Event3Evaluator) checkBathroom(tenantID string, card repository.CardInfo) (bool, error) {
	// 方法1：从卡片绑定的设备中获取 room_name
	devices, err := e.evaluator.cardRepo.GetCardDevices(card.CardID)
	if err != nil {
		return false, err
	}

	// 检查设备绑定的房间名称
	for _, device := range devices {
		if device.RoomName != nil {
			if checkBathroomRoomName(*device.RoomName) {
				return true, nil
			}
		}
	}

	// 方法2：如果卡片有 room_id，直接查询房间信息
	if card.RoomID != nil {
		isBathroom, err := e.evaluator.roomRepo.IsBathroom(context.Background(), tenantID, *card.RoomID)
		if err != nil {
			return false, err
		}
		return isBathroom, nil
	}

	return false, nil
}

// getOrInitState 获取或初始化状态
func (e *Event3Evaluator) getOrInitState(
	ctx context.Context,
	cardID, trackID string,
	realtimeData *models.RealtimeData,
) (*consumer.Event3State, error) {
	// 尝试获取现有状态
	state, err := e.getEvent3State(ctx, cardID, trackID)
	if err != nil {
		return nil, err
	}

	// 如果状态已存在，直接返回
	if state.EnterTime != nil {
		return state, nil
	}

	// 初始化新状态（T0：进入 bathroom 的时间）
	now := time.Now().Unix()
	state.EnterTime = &now
	state.TrackID = trackID

	// 如果进入时就是站立，记录站姿时间
	currentPosture, found := findTrackInBathroom(realtimeData, trackID)
	if found && checkStandingPosture(currentPosture) {
		state.StandingTime = &now
	}

	// 记录初始位置
	if currentPosture != nil && currentPosture.PositionX != nil && currentPosture.PositionY != nil {
		if state.LastPosition == nil {
			state.LastPosition = &struct {
				X float64 `json:"x"`
				Y float64 `json:"y"`
			}{}
		}
		state.LastPosition.X = float64(*currentPosture.PositionX)
		state.LastPosition.Y = float64(*currentPosture.PositionY)
		state.LastPositionTime = &now
	}

	// 保存状态
	if err := e.setEvent3State(ctx, cardID, state); err != nil {
		return nil, err
	}

	return state, nil
}

// triggerFallAlarm 触发 Fall 报警（ALERT 级别）
func (e *Event3Evaluator) triggerFallAlarm(
	ctx context.Context,
	tenantID string,
	card repository.CardInfo,
	realtimeData *models.RealtimeData,
	state *consumer.Event3State,
	posture *models.Posture,
	stillDuration int64,
) (*models.AlarmEvent, error) {
	// 获取设备ID（优先使用 Radar 设备）
	deviceID := ""
	if len(realtimeData.Postures) > 0 {
		deviceID = realtimeData.Postures[0].DeviceID
	}

	builder := NewAlarmEventBuilder(tenantID, deviceID)

	postureCode := ""
	postureDisplay := ""
	if posture != nil {
		postureCode = posture.PostureCode
		postureDisplay = posture.PostureDisplay
	}
	snomedCode := "248220002"
	snomedDisplay := "Fall"
	durationSec := int(stillDuration)
	triggerData := BuildTriggerData(
		"Fall",
		"Radar",
		realtimeData.Heart,
		realtimeData.Breath,
		&postureCode,
		&postureDisplay,
		&snomedCode,
		&snomedDisplay,
		nil, // confidence
		&durationSec,
	)

	metadata := map[string]interface{}{
		"card_id":         card.CardID,
		"track_id":        state.TrackID,
		"trigger_point":   "bathroom_still_30min",
		"enter_time":      state.EnterTime,
		"still_duration":  stillDuration,
		"position_change": state.PositionChange,
		"reason":          "No one can stand still for more than 10 minutes, likely fall due to room interference",
	}

	alarm, err := builder.BuildAlarmEvent(
		"Fall",
		"safety",
		"ALERT", // ALERT 级别
		triggerData,
		metadata,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build alarm: %w", err)
	}

	return alarm, nil
}

// triggerSuspectedFallAlarm 触发 SuspectedFall 报警（WARNING 级别）
func (e *Event3Evaluator) triggerSuspectedFallAlarm(
	ctx context.Context,
	tenantID string,
	card repository.CardInfo,
	realtimeData *models.RealtimeData,
	state *consumer.Event3State,
	posture *models.Posture,
	standingDuration int64,
) (*models.AlarmEvent, error) {
	// 获取设备ID（优先使用 Radar 设备）
	deviceID := ""
	if len(realtimeData.Postures) > 0 {
		deviceID = realtimeData.Postures[0].DeviceID
	}

	builder := NewAlarmEventBuilder(tenantID, deviceID)

	postureCode := ""
	postureDisplay := ""
	if posture != nil {
		postureCode = posture.PostureCode
		postureDisplay = posture.PostureDisplay
	}
	snomedCode := "248220002"
	snomedDisplay := "Suspected Fall"
	durationSec := int(standingDuration)
	triggerData := BuildTriggerData(
		"SuspectedFall",
		"Radar",
		realtimeData.Heart,
		realtimeData.Breath,
		&postureCode,
		&postureDisplay,
		&snomedCode,
		&snomedDisplay,
		nil, // confidence
		&durationSec,
	)

	metadata := map[string]interface{}{
		"card_id":           card.CardID,
		"track_id":          state.TrackID,
		"trigger_point":     "bathroom_stand_still_10min",
		"enter_time":        state.EnterTime,
		"standing_duration": standingDuration,
		"position_change":   state.PositionChange,
	}

	alarm, err := builder.BuildAlarmEvent(
		"SuspectedFall",
		"safety",
		"WARNING", // WARNING 级别
		triggerData,
		metadata,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build alarm: %w", err)
	}

	return alarm, nil
}

// getEvent3State 获取事件3的状态
func (e *Event3Evaluator) getEvent3State(ctx context.Context, cardID, trackID string) (*consumer.Event3State, error) {
	stateKey := e.evaluator.stateManager.GetStateKey(cardID, trackID, "event3")

	var state consumer.Event3State
	err := e.evaluator.stateManager.GetState(ctx, stateKey, &state)
	if err != nil {
		// 状态不存在，返回空状态
		return &consumer.Event3State{
			TrackID: trackID,
		}, nil
	}

	return &state, nil
}

// setEvent3State 设置事件3的状态
func (e *Event3Evaluator) setEvent3State(ctx context.Context, cardID string, state *consumer.Event3State) error {
	stateKey := e.evaluator.stateManager.GetStateKey(cardID, state.TrackID, "event3")

	// 设置 TTL 为 60 分钟（事件3的最长持续时间）
	ttl := 60 * time.Minute
	return e.evaluator.stateManager.SetState(ctx, stateKey, state, ttl)
}

// clearState 清除状态
func (e *Event3Evaluator) clearState(ctx context.Context, cardID, trackID string) {
	stateKey := e.evaluator.stateManager.GetStateKey(cardID, trackID, "event3")
	_ = e.evaluator.stateManager.DeleteState(ctx, stateKey)
}
