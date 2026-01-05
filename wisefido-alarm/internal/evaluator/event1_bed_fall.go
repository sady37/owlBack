package evaluator

import (
	"context"
	"fmt"
	"time"
	"wisefido-alarm/internal/consumer"
	"wisefido-alarm/internal/models"
	"wisefido-alarm/internal/repository"

	"go.uber.org/zap"
)

// Event1Evaluator 事件1：床上跌落检测评估器
type Event1Evaluator struct {
	evaluator *Evaluator
}

// NewEvent1Evaluator 创建事件1评估器
func NewEvent1Evaluator(evaluator *Evaluator) *Event1Evaluator {
	return &Event1Evaluator{
		evaluator: evaluator,
	}
}

// Evaluate 评估事件1：床上跌落检测
//
// 完整流程：
// 1. 前置条件检查（ActiveBed、Radar、Sleepad、bed_status="off_bed"）
// 2. 获取/初始化状态（T0、track_id、bed_area_id）
// 3. T0-T0+10秒：持续检查（每次收到数据时）
//   - 检查退出条件（姿态、离床、移动）
//   - 检查 track_id 消失（30秒窗口）
//   - 检查高度降低（SuspectedFall）
//
// 4. T0+120秒：检查 HR/RR（Fall）
//
// 数据源：仅使用Redis实时数据（RealtimeData），不查询数据库
func (e *Event1Evaluator) Evaluate(tenantID string, card repository.CardInfo, realtimeData *models.RealtimeData) ([]models.AlarmEvent, error) {
	ctx := context.Background()
	var alarms []models.AlarmEvent

	// 1. 前置条件检查
	if !e.checkPrerequisites(card, realtimeData) {
		return nil, nil
	}

	// 2. 获取床的 area_id（从 config_version 表）
	if card.BedID == nil {
		return nil, nil
	}
	bedAreaID, err := e.getBedAreaID(ctx, tenantID, *card.BedID)
	if err != nil {
		e.evaluator.logger.Debug("Failed to get bed area_id",
			zap.String("card_id", card.CardID),
			zap.String("bed_id", *card.BedID),
			zap.Error(err),
		)
		return nil, nil
	}

	// 3. 选择要跟踪的 track_id（在床区域的第一个 track）
	trackID := selectTrackID(realtimeData, bedAreaID)
	if trackID == "" {
		// 没有人在床区域，退出
		return nil, nil
	}

	// 4. 获取或初始化状态
	state, err := e.getOrInitState(ctx, card.CardID, trackID, realtimeData, bedAreaID)
	if err != nil {
		e.evaluator.logger.Error("Failed to get or init state",
			zap.String("card_id", card.CardID),
			zap.String("track_id", trackID),
			zap.Error(err),
		)
		return nil, nil
	}

	// 5. 检查是否已经触发过报警（避免重复报警）
	if state.AlarmTriggered {
		return nil, nil
	}

	// 6. 找到当前 track_id 的 posture
	currentPosture, found := findTrackInBedArea(realtimeData, bedAreaID, trackID)
	if !found {
		// track_id 不在床区域，检查是否消失
		shouldAlarm, shouldContinue := checkTrackDisappeared(ctx, realtimeData, state, e.evaluator.logger)
		if shouldAlarm {
			// 触发 Fall 报警
			alarm, err := e.triggerFallAlarm(ctx, tenantID, card, realtimeData, state, nil, "track_disappeared")
			if err != nil {
				return nil, err
			}
			alarms = append(alarms, *alarm)
			state.AlarmTriggered = true
			e.setEvent1State(ctx, card.CardID, state)
			return alarms, nil
		}
		if !shouldContinue {
			// 退出检测
			return nil, nil
		}
		// 继续监控
		return nil, nil
	}

	// 7. 检查退出条件
	if checkExitConditions(realtimeData, state, currentPosture) {
		// 满足退出条件，清除状态
		e.clearState(ctx, card.CardID, trackID)
		return nil, nil
	}

	// 8. 更新位置信息
	updatePositionInState(state, currentPosture)

	// 9. 检查高度降低（SuspectedFall）
	heightDropped, dropAmount := checkHeightDrop(state, currentPosture)
	if heightDropped {
		e.evaluator.logger.Warn("Height dropped significantly, triggering SuspectedFall",
			zap.String("card_id", card.CardID),
			zap.String("track_id", trackID),
			zap.Float64("drop_amount", dropAmount),
		)
		alarm, err := e.triggerSuspectedFallAlarm(ctx, tenantID, card, realtimeData, state, currentPosture)
		if err != nil {
			return nil, err
		}
		alarms = append(alarms, *alarm)
		state.AlarmTriggered = true
		e.setEvent1State(ctx, card.CardID, state)
		return alarms, nil
	}

	// 10. 检查 T0+120秒：如果仍是 lying 且无 HR/RR，触发 Fall
	if state.LeftBedTime != nil {
		leftBedTime := time.Unix(*state.LeftBedTime, 0)
		elapsed := time.Since(leftBedTime)
		if elapsed >= 120*time.Second {
			// 检查是否有 HR/RR
			if realtimeData.Heart == nil && realtimeData.Breath == nil {
				// 无 HR/RR，触发 Fall 报警
				alarm, err := e.triggerFallAlarm(ctx, tenantID, card, realtimeData, state, currentPosture, "no_hr_rr_120s")
				if err != nil {
					return nil, err
				}
				alarms = append(alarms, *alarm)
				state.AlarmTriggered = true
				e.setEvent1State(ctx, card.CardID, state)
				return alarms, nil
			} else {
				// 有 HR/RR，退出检测
				e.clearState(ctx, card.CardID, trackID)
				return nil, nil
			}
		}
	}

	// 11. 保存状态
	if err := e.setEvent1State(ctx, card.CardID, state); err != nil {
		e.evaluator.logger.Error("Failed to save state",
			zap.String("card_id", card.CardID),
			zap.Error(err),
		)
	}

	return alarms, nil
}

// checkPrerequisites 检查前置条件
func (e *Event1Evaluator) checkPrerequisites(card repository.CardInfo, realtimeData *models.RealtimeData) bool {
	// 1. 必须是 ActiveBed 卡片
	if card.CardType != "ActiveBed" {
		return false
	}

	// 2. 检查是否有离床事件（通过 bed_status 判断）
	if realtimeData.BedStatus == nil {
		return false
	}

	// 3. 检查 sleepad 是否检测到离床
	if *realtimeData.BedStatus != "off_bed" {
		return false
	}

	// 4. 如果 Sleepace 有 HR/RR，说明人可能回到床上，退出
	if realtimeData.Heart != nil || realtimeData.Breath != nil {
		return false
	}

	// 5. 检查是否有 Radar 数据
	if len(realtimeData.Postures) == 0 {
		return false
	}

	return true
}

// getBedAreaID 获取床的区域ID（从 config_version 表）
func (e *Event1Evaluator) getBedAreaID(ctx context.Context, tenantID, bedID string) (int, error) {
	// 通过 bed_id 查询 room_id
	roomID, err := e.evaluator.roomRepo.GetRoomIDByBedID(ctx, tenantID, bedID)
	if err != nil {
		return 0, fmt.Errorf("failed to get room_id by bed_id: %w", err)
	}

	// 从 config_version 表获取 bed_area_id
	bedAreaID, err := e.evaluator.configVersionRepo.GetBedAreaID(ctx, tenantID, roomID)
	if err != nil {
		return 0, fmt.Errorf("failed to get bed area_id: %w", err)
	}

	return bedAreaID, nil
}

// getOrInitState 获取或初始化状态
func (e *Event1Evaluator) getOrInitState(
	ctx context.Context,
	cardID, trackID string,
	realtimeData *models.RealtimeData,
	bedAreaID int,
) (*consumer.Event1State, error) {
	// 尝试获取现有状态
	state, err := e.getEvent1State(ctx, cardID, trackID)
	if err != nil {
		return nil, err
	}

	// 如果状态已存在，直接返回
	if state.LeftBedTime != nil {
		return state, nil
	}

	// 初始化新状态（T0：离床时间）
	now := time.Now().Unix()
	state.LeftBedTime = &now
	state.TrackID = trackID

	// 记录 lying 状态（如果当前是 lying）
	currentPosture, found := findTrackInBedArea(realtimeData, bedAreaID, trackID)
	if found && currentPosture != nil && currentPosture.PostureCode == "248218002" {
		// 记录 lying 状态
		state.LyingTime = &now
		if currentPosture.Height != nil {
			height := float64(*currentPosture.Height)
			state.LyingHeight = &height
		}
		if currentPosture.PositionX != nil && currentPosture.PositionY != nil {
			if state.LyingPosition == nil {
				state.LyingPosition = &struct {
					X int `json:"x"`
					Y int `json:"y"`
				}{}
			}
			state.LyingPosition.X = *currentPosture.PositionX
			state.LyingPosition.Y = *currentPosture.PositionY
		}
	}

	// 保存状态
	if err := e.setEvent1State(ctx, cardID, state); err != nil {
		return nil, err
	}

	return state, nil
}

// triggerFallAlarm 触发 Fall 报警（ALERT 级别）
func (e *Event1Evaluator) triggerFallAlarm(
	ctx context.Context,
	tenantID string,
	card repository.CardInfo,
	realtimeData *models.RealtimeData,
	state *consumer.Event1State,
	posture *models.Posture,
	triggerPoint string,
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
		nil, // duration_sec
	)

	metadata := map[string]interface{}{
		"card_id":       card.CardID,
		"track_id":      state.TrackID,
		"trigger_point": triggerPoint,
		"left_bed_time": state.LeftBedTime,
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
func (e *Event1Evaluator) triggerSuspectedFallAlarm(
	ctx context.Context,
	tenantID string,
	card repository.CardInfo,
	realtimeData *models.RealtimeData,
	state *consumer.Event1State,
	posture *models.Posture,
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
		nil, // duration_sec
	)

	metadata := map[string]interface{}{
		"card_id":        card.CardID,
		"track_id":       state.TrackID,
		"trigger_point":  "height_dropped",
		"current_height": posture.Height,
		"lying_height":   state.LyingHeight,
		"left_bed_time":  state.LeftBedTime,
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

// getEvent1State 获取事件1的状态
func (e *Event1Evaluator) getEvent1State(ctx context.Context, cardID, trackID string) (*consumer.Event1State, error) {
	stateKey := e.evaluator.stateManager.GetStateKey(cardID, trackID, "event1")

	var state consumer.Event1State
	err := e.evaluator.stateManager.GetState(ctx, stateKey, &state)
	if err != nil {
		// 状态不存在，返回空状态
		return &consumer.Event1State{
			TrackID: trackID,
		}, nil
	}

	return &state, nil
}

// setEvent1State 设置事件1的状态
func (e *Event1Evaluator) setEvent1State(ctx context.Context, cardID string, state *consumer.Event1State) error {
	stateKey := e.evaluator.stateManager.GetStateKey(cardID, state.TrackID, "event1")

	// 设置 TTL 为 5 分钟（事件1的最长持续时间）
	ttl := 5 * time.Minute
	return e.evaluator.stateManager.SetState(ctx, stateKey, state, ttl)
}

// clearState 清除状态
func (e *Event1Evaluator) clearState(ctx context.Context, cardID, trackID string) {
	stateKey := e.evaluator.stateManager.GetStateKey(cardID, trackID, "event1")
	_ = e.evaluator.stateManager.DeleteState(ctx, stateKey)
}
