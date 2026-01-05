package evaluator

import (
	"context"
	"wisefido-alarm/internal/config"
	"wisefido-alarm/internal/consumer"
	"wisefido-alarm/internal/models"
	"wisefido-alarm/internal/repository"

	"go.uber.org/zap"
)

// Evaluator 报警评估器（实现 consumer.Evaluator 接口）
type Evaluator struct {
	config            *config.Config
	stateManager      *consumer.StateManager
	cardRepo          *repository.CardRepository
	deviceRepo        *repository.DeviceRepository
	roomRepo          *repository.RoomRepository
	alarmCloudRepo    *repository.AlarmCloudRepository
	alarmDeviceRepo   *repository.AlarmDeviceRepository
	alarmEventsRepo   *repository.AlarmEventsRepository
	iotRepo           *repository.IoTTimeSeriesRepository
	configVersionRepo *repository.ConfigVersionRepository
	logger            *zap.Logger

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
	iotRepo *repository.IoTTimeSeriesRepository,
	configVersionRepo *repository.ConfigVersionRepository,
	logger *zap.Logger,
) *Evaluator {
	e := &Evaluator{
		config:            cfg,
		stateManager:      stateManager,
		cardRepo:          cardRepo,
		deviceRepo:        deviceRepo,
		roomRepo:          roomRepo,
		alarmCloudRepo:    alarmCloudRepo,
		alarmDeviceRepo:   alarmDeviceRepo,
		alarmEventsRepo:   alarmEventsRepo,
		iotRepo:           iotRepo,
		configVersionRepo: configVersionRepo,
		logger:            logger,
	}

	// 初始化事件评估器
	e.event1 = NewEvent1Evaluator(e)
	e.event2 = NewEvent2Evaluator(e)
	e.event3 = NewEvent3Evaluator(e)
	e.event4 = NewEvent4Evaluator(e)

	return e
}

// Evaluate 评估卡片数据，返回报警事件列表
// 注意：现在只处理高级推理报警（事件1和事件3）
// 设备直接报警由 wisefido-sensor-fusion 处理
func (e *Evaluator) Evaluate(tenantID string, card repository.CardInfo, realtimeData *models.RealtimeData) ([]models.AlarmEvent, error) {
	var alarms []models.AlarmEvent

	// 评估事件1：床上跌落检测（高级推理：sleepad+radar融合）
	event1Alarms, err := e.event1.Evaluate(tenantID, card, realtimeData)
	if err != nil {
		e.logger.Error("Failed to evaluate event1",
			zap.String("card_id", card.CardID),
			zap.Error(err),
		)
	} else {
		alarms = append(alarms, event1Alarms...)
	}

	// 评估事件3：Bathroom可疑跌倒检测（高级推理：位置模式分析）
	event3Alarms, err := e.event3.Evaluate(tenantID, card, realtimeData)
	if err != nil {
		e.logger.Error("Failed to evaluate event3",
			zap.String("card_id", card.CardID),
			zap.Error(err),
		)
	} else {
		alarms = append(alarms, event3Alarms...)
	}

	// 注意：事件2（Sleepad可靠性判断）已移除
	// 注意：事件4（雷达检测到人突然消失）已移除
	// 设备直接报警由 wisefido-sensor-fusion 处理

	// 写入报警事件到 PostgreSQL
	ctx := context.Background()
	alarmCreated := false
	for _, alarm := range alarms {
		if err := e.alarmEventsRepo.CreateAlarmEvent(ctx, tenantID, &alarm); err != nil {
			e.logger.Error("Failed to create alarm event",
				zap.String("event_id", alarm.EventID),
				zap.String("event_type", alarm.EventType),
				zap.Error(err),
			)
			// 继续处理其他报警，不中断
		} else {
			alarmCreated = true
			e.logger.Info("Alarm event created",
				zap.String("event_id", alarm.EventID),
				zap.String("event_type", alarm.EventType),
				zap.String("alarm_level", alarm.AlarmLevel),
				zap.String("card_id", card.CardID),
			)
		}
	}

	// 如果有报警创建成功，更新卡片的报警计数
	if alarmCreated {
		if err := e.cardRepo.UpdateCardAlarmCounts(ctx, tenantID, card.CardID); err != nil {
			e.logger.Warn("Failed to update card alarm counts",
				zap.String("card_id", card.CardID),
				zap.String("tenant_id", tenantID),
				zap.Error(err),
			)
			// 不返回错误，报警已创建成功
		} else {
			e.logger.Debug("Updated card alarm counts",
				zap.String("card_id", card.CardID),
			)
		}
	}

	return alarms, nil
}
