package evaluator

import (
	"wisefido-alarm/internal/models"
	"wisefido-alarm/internal/repository"

	"go.uber.org/zap"
)

// Event4Evaluator 事件4：雷达检测到人突然消失评估器
type Event4Evaluator struct {
	evaluator *Evaluator
}

// NewEvent4Evaluator 创建事件4评估器
func NewEvent4Evaluator(evaluator *Evaluator) *Event4Evaluator {
	return &Event4Evaluator{
		evaluator: evaluator,
	}
}

// Evaluate 评估事件4
// 目的：检测质心降低 + 突然消失，可能是跌倒
func (e *Event4Evaluator) Evaluate(tenantID string, card repository.CardInfo, realtimeData *models.RealtimeData) ([]models.AlarmEvent, error) {
	// TODO: 实现完整的事件4逻辑
	// 当前返回空列表，待后续实现

	// 检查条件：
	// 1. 检查是否有 track_id 消失
	// TODO: 需要维护 track_id 的历史状态，检测是否突然消失
	// 当前简化处理，暂时不评估

	// 2. 检查质心是否降低（高度降低超过60cm）
	// TODO: 需要维护历史高度数据，检测高度变化
	// 当前简化处理，暂时不评估

	e.evaluator.logger.Debug("Event4 evaluation (simplified)",
		zap.String("card_id", card.CardID),
		zap.Int("person_count", realtimeData.PersonCount),
	)

	return nil, nil
}
