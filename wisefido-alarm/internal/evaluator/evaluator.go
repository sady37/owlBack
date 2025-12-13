package evaluator

import (
	"wisefido-alarm/internal/config"
	"wisefido-alarm/internal/consumer"
	"wisefido-alarm/internal/models"
	"wisefido-alarm/internal/repository"

	"go.uber.org/zap"
)

// Evaluator 报警评估器（实现 consumer.Evaluator 接口）
type Evaluator struct {
	config          *config.Config
	stateManager    *consumer.StateManager
	cardRepo        *repository.CardRepository
	deviceRepo      *repository.DeviceRepository
	roomRepo        *repository.RoomRepository
	alarmCloudRepo  *repository.AlarmCloudRepository
	alarmDeviceRepo *repository.AlarmDeviceRepository
	alarmEventsRepo *repository.AlarmEventsRepository
	logger          *zap.Logger

	// 事件评估器
	event1 *Event1Evaluator // 床上跌落检测
	event2 *Event2Evaluator // Sleepad可靠性判断
	event3 *Event3Evaluator // Bathroom可疑跌倒检测
	event4 *Event4Evaluator // 雷达检测到人突然消失
}

// NewEvaluator 创建评估器
func NewEvaluator(
	cfg *config.Config,
	stateManager *consumer.StateManager,
	cardRepo *repository.CardRepository,
	deviceRepo *repository.DeviceRepository,
	roomRepo *repository.RoomRepository,
	alarmCloudRepo *repository.AlarmCloudRepository,
	alarmDeviceRepo *repository.AlarmDeviceRepository,
	alarmEventsRepo *repository.AlarmEventsRepository,
	logger *zap.Logger,
) *Evaluator {
	e := &Evaluator{
		config:          cfg,
		stateManager:    stateManager,
		cardRepo:        cardRepo,
		deviceRepo:      deviceRepo,
		roomRepo:        roomRepo,
		alarmCloudRepo:  alarmCloudRepo,
		alarmDeviceRepo: alarmDeviceRepo,
		alarmEventsRepo: alarmEventsRepo,
		logger:          logger,
	}

	// 初始化事件评估器
	e.event1 = NewEvent1Evaluator(e)
	e.event2 = NewEvent2Evaluator(e)
	e.event3 = NewEvent3Evaluator(e)
	e.event4 = NewEvent4Evaluator(e)

	return e
}

// Evaluate 评估卡片数据，返回报警事件列表
func (e *Evaluator) Evaluate(tenantID string, card repository.CardInfo, realtimeData *models.RealtimeData) ([]models.AlarmEvent, error) {
	var alarms []models.AlarmEvent

	// 评估事件1：床上跌落检测
	event1Alarms, err := e.event1.Evaluate(tenantID, card, realtimeData)
	if err != nil {
		e.logger.Error("Failed to evaluate event1",
			zap.String("card_id", card.CardID),
			zap.Error(err),
		)
	} else {
		alarms = append(alarms, event1Alarms...)
	}

	// 评估事件2：Sleepad可靠性判断
	event2Alarms, err := e.event2.Evaluate(tenantID, card, realtimeData)
	if err != nil {
		e.logger.Error("Failed to evaluate event2",
			zap.String("card_id", card.CardID),
			zap.Error(err),
		)
	} else {
		alarms = append(alarms, event2Alarms...)
	}

	// 评估事件3：Bathroom可疑跌倒检测
	event3Alarms, err := e.event3.Evaluate(tenantID, card, realtimeData)
	if err != nil {
		e.logger.Error("Failed to evaluate event3",
			zap.String("card_id", card.CardID),
			zap.Error(err),
		)
	} else {
		alarms = append(alarms, event3Alarms...)
	}

	// 评估事件4：雷达检测到人突然消失
	event4Alarms, err := e.event4.Evaluate(tenantID, card, realtimeData)
	if err != nil {
		e.logger.Error("Failed to evaluate event4",
			zap.String("card_id", card.CardID),
			zap.Error(err),
		)
	} else {
		alarms = append(alarms, event4Alarms...)
	}

	// 写入报警事件到 PostgreSQL
	for _, alarm := range alarms {
		if err := e.alarmEventsRepo.CreateAlarmEvent(&alarm); err != nil {
			e.logger.Error("Failed to create alarm event",
				zap.String("event_id", alarm.EventID),
				zap.String("event_type", alarm.EventType),
				zap.Error(err),
			)
			// 继续处理其他报警，不中断
		} else {
			e.logger.Info("Alarm event created",
				zap.String("event_id", alarm.EventID),
				zap.String("event_type", alarm.EventType),
				zap.String("alarm_level", alarm.AlarmLevel),
				zap.String("card_id", card.CardID),
			)
		}
	}

	return alarms, nil
}
