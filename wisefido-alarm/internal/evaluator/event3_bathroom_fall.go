package evaluator

import (
	"context"
	"strings"
	"wisefido-alarm/internal/models"
	"wisefido-alarm/internal/repository"

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

// Evaluate 评估事件3
// 目的：检测卫生间内长时间站立不动，可能是跌倒后无法移动
func (e *Event3Evaluator) Evaluate(tenantID string, card repository.CardInfo, realtimeData *models.RealtimeData) ([]models.AlarmEvent, error) {
	// TODO: 实现完整的事件3逻辑
	// 当前返回空列表，待后续实现

	// 检查条件：
	// 1. 检查房间是否是 bathroom
	isBathroom, err := e.checkBathroom(tenantID, card)
	if err != nil {
		return nil, err
	}
	if !isBathroom {
		return nil, nil
	}

	// 2. 检查雷达检测范围内是否仅1人
	if realtimeData.PersonCount != 1 {
		return nil, nil
	}

	// 3. 检查是否有1个人处于站立状态（不是坐着）
	// TODO: 需要检查 postures 中的姿态，判断是否是站立状态
	// 当前简化处理，暂时不评估

	e.evaluator.logger.Debug("Event3 evaluation (simplified)",
		zap.String("card_id", card.CardID),
		zap.Bool("is_bathroom", isBathroom),
		zap.Int("person_count", realtimeData.PersonCount),
	)

	return nil, nil
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
			roomNameLower := strings.ToLower(*device.RoomName)
			if strings.Contains(roomNameLower, "bathroom") ||
				strings.Contains(roomNameLower, "restroom") ||
				strings.Contains(roomNameLower, "toilet") {
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
