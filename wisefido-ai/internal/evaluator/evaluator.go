package evaluator

import (
	"context"
	"database/sql"
	"encoding/json"
	"wisefido-ai/internal/config"
	"wisefido-ai/internal/consumer"
	"wisefido-ai/internal/models"
	"wisefido-ai/internal/alarmpush"
	"wisefido-ai/internal/repository"

	commoncard "owl-common/card"

	"go.uber.org/zap"
)

// Evaluator 报警评估器（实现 consumer.Evaluator 接口）
type Evaluator struct {
	config            *config.Config
	db                *sql.DB
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
	db *sql.DB,
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
		db:                db,
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
// 设备直接报警由 wisefido-card-aggregator 处理
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
	// 设备直接报警由 wisefido-card-aggregator 处理

	// 写入报警事件到 PostgreSQL（通过 owl-common 原子操作：INSERT alarm + UPDATE card counts）
	// AI 使用虚拟 device_id，findCardIDByDevice 查不到；
	// 必须在 metadata 中注入 card_id，供后续 UpdateAlarmAndUpdateCard fallback 使用。
	ctx := context.Background()
	for i, alarm := range alarms {
		metadata := ensureMetadataCardID(alarm.Metadata, card.CardID)
		params := commoncard.AlarmInsertParams{
			TenantID:    alarm.TenantID,
			DeviceID:    alarm.DeviceID,
			EventType:   alarm.EventType,
			Category:    alarm.Category,
			AlarmLevel:  alarm.AlarmLevel,
			TriggeredAt: alarm.TriggeredAt,
			TriggerData: alarm.TriggerData,
			Metadata:    metadata,
		}
		result, _, err := commoncard.InsertAlarmAndUpdateCard(ctx, e.db, card.CardID, params)
		if err != nil {
			e.logger.Error("Failed to insert alarm and update card",
				zap.String("event_type", alarm.EventType),
				zap.String("card_id", card.CardID),
				zap.Error(err),
			)
			continue
		}
		alarms[i].EventID = result.EventID
		alarmpush.NotifyWisefidoData(e.logger, alarm.TenantID, card.CardID, alarm.DeviceID, result.EventID, alarm.EventType, alarm.AlarmLevel)
		e.logger.Info("Alarm event created and card updated",
			zap.String("event_id", result.EventID),
			zap.String("event_type", alarm.EventType),
			zap.String("alarm_level", alarm.AlarmLevel),
			zap.String("card_id", card.CardID),
		)
	}

	return alarms, nil
}

// ensureMetadataCardID 确保 metadata 中包含 card_id（AI 虚拟设备 fallback 必需）
func ensureMetadataCardID(raw json.RawMessage, cardID string) json.RawMessage {
	if cardID == "" {
		return raw
	}
	m := map[string]interface{}{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &m)
	}
	if _, ok := m["card_id"]; ok {
		return raw
	}
	m["card_id"] = cardID
	out, err := json.Marshal(m)
	if err != nil {
		return raw
	}
	return out
}
