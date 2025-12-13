package evaluator

import (
	"context"
	"wisefido-alarm/internal/models"
	"wisefido-alarm/internal/repository"

	"go.uber.org/zap"
)

// Event2Evaluator 事件2：Sleepad可靠性判断评估器
type Event2Evaluator struct {
	evaluator *Evaluator
}

// NewEvent2Evaluator 创建事件2评估器
func NewEvent2Evaluator(evaluator *Evaluator) *Event2Evaluator {
	return &Event2Evaluator{
		evaluator: evaluator,
	}
}

// Evaluate 评估事件2
// 目的：避免电磁或振动干扰导致的误报
func (e *Event2Evaluator) Evaluate(tenantID string, card repository.CardInfo, realtimeData *models.RealtimeData) ([]models.AlarmEvent, error) {
	// TODO: 实现完整的事件2逻辑
	// 当前返回空列表，待后续实现
	
	// 检查条件：
	// 1. 必须是 ActiveBed 卡片
	if card.CardType != "ActiveBed" {
		return nil, nil
	}

	// 2. 检查是否有 Sleepace 设备（需要 HR/RR 数据）
	if realtimeData.Heart == nil && realtimeData.Breath == nil {
		// 没有 HR/RR 数据，无法判断可靠性
		return nil, nil
	}

	// 3. 检查床上是否绑定了 Radar 设备
	// TODO: 需要查询卡片绑定的设备，检查是否有 Radar 设备
	// 当前简化处理，暂时不评估

	e.evaluator.logger.Debug("Event2 evaluation (simplified)",
		zap.String("card_id", card.CardID),
	)

	return nil, nil
}

// checkRadarOnBed 检查床上是否绑定了 Radar 设备
func (e *Event2Evaluator) checkRadarOnBed(ctx context.Context, card repository.CardInfo) (bool, error) {
	// 获取卡片绑定的设备
	devices, err := e.evaluator.cardRepo.GetCardDevices(card.CardID)
	if err != nil {
		return false, err
	}

	// 检查是否有 Radar 设备
	for _, device := range devices {
		if device.DeviceType == "Radar" {
			// 检查是否绑定到床（bed_id 有效）
			if device.BedID != nil && *device.BedID != "" {
				return true, nil
			}
		}
	}

	return false, nil
}

