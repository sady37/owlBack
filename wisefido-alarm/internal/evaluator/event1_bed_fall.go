package evaluator

import (
	"context"
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

// Evaluate 评估事件1
// 注意：这是一个复杂的事件，需要维护状态和定时器
// 当前实现是简化版本，完整实现需要：
// 1. 状态管理（lying基线、离床时间、track_id状态）
// 2. 定时器（T0+5秒、T0+60秒、T0+120秒）
// 3. 退出条件检查（持续检查）
func (e *Event1Evaluator) Evaluate(tenantID string, card repository.CardInfo, realtimeData *models.RealtimeData) ([]models.AlarmEvent, error) {
	// TODO: 实现完整的事件1逻辑
	// 当前返回空列表，待后续实现

	// 检查条件：
	// 1. 必须是 ActiveBed 卡片
	if card.CardType != "ActiveBed" {
		return nil, nil
	}

	// 2. 检查是否有离床事件（通过 bed_status 判断）
	// 注意：bed_status 来自 Sleepace，如果为 "off_bed" 表示离床
	if realtimeData.BedStatus == nil {
		return nil, nil
	}

	// 3. 检查是否有 Sleepace 设备（需要 HR/RR 数据）
	// 如果 Sleepace 有 HR/RR，说明人可能回到床上，退出
	if realtimeData.Heart != nil || realtimeData.Breath != nil {
		// 有 HR/RR，退出事件1
		return nil, nil
	}

	// TODO: 实现完整的状态管理和定时器逻辑
	// 当前仅做基础检查，返回空列表

	e.evaluator.logger.Debug("Event1 evaluation (simplified)",
		zap.String("card_id", card.CardID),
		zap.String("bed_status", *realtimeData.BedStatus),
	)

	return nil, nil
}

// checkExitConditions 检查退出条件
func (e *Event1Evaluator) checkExitConditions(ctx context.Context, card repository.CardInfo, realtimeData *models.RealtimeData) bool {
	// 退出条件：
	// 1. sleepad有HR/RR → 退出
	if realtimeData.Heart != nil || realtimeData.Breath != nil {
		return true
	}

	// 2. sleepad有上床事件 → 退出（bed_status 变为 "on_bed"）
	if realtimeData.BedStatus != nil && *realtimeData.BedStatus == "on_bed" {
		return true
	}

	// 3. radar检测到在移动 → 退出
	// TODO: 需要检查 postures 中是否有移动的 track_id
	// 当前简化处理，如果 person_count > 0 且 postures 不为空，认为可能在移动
	if realtimeData.PersonCount > 0 && len(realtimeData.Postures) > 0 {
		// TODO: 检查位置是否变化（需要历史位置数据）
		// 当前简化处理，暂时不退出
	}

	return false
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
