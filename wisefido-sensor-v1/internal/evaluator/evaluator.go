package evaluator

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/netip"
	"wisefido-sensor-v1/internal/config"
	"wisefido-sensor-v1/internal/consumer"
	"wisefido-sensor-v1/internal/models"
	"wisefido-sensor-v1/internal/repository"

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
	// PR1 (A11): 不再直插 alarm_events 表，改 publish 到 iot:alarm:stream
	// 让 cardagg 统一持久化（cardagg = data 对接层，单 writer）。
	alarmBackChannel  *consumer.AlarmBackChannel
	logger            *zap.Logger

	// 事件评估器
	event1 *Event1Evaluator // 床上跌落检测
	event2 *Event2Evaluator // Sleepad可靠性判断
	event3 *Event3Evaluator // Bathroom可疑跌倒检测
	event4 *Event4Evaluator // 雷达检测到人突然消失
}

// NewEvaluator 创建评估器
//
// PR1 (A11): alarmBackChannel 为 nil 时降级为"仅评估不报警"（dev/test 模式）；
// 生产环境必须传入有效的 back channel（redis 已建连），否则评估输出无处发送。
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
	alarmBackChannel *consumer.AlarmBackChannel,
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
		alarmBackChannel:  alarmBackChannel,
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

	// PR1 (A11): 不再直插 alarm_events 表 / 不再调 alarmpush 推送。
	// 改 publish 到 iot:alarm:stream，由 cardagg 现有 alarm_handler 统一持久化 + 推送 +
	// device_flags Redis 投影。dedup / SkippedNotify / NotifyWisefidoData 全部由 cardagg 端
	// PersistAlarmAndPublish + notifyAlarmPushAsync 负责。
	if e.alarmBackChannel == nil {
		e.logger.Warn("alarm back channel nil; alarms silently dropped",
			zap.String("card_id", card.CardID),
			zap.Int("alarm_count", len(alarms)),
		)
		return alarms, nil
	}
	ctx := context.Background()
	for i, al := range alarms {
		deviceAddr, parseErr := netip.ParseAddr(al.DeviceID)
		if parseErr != nil {
			e.logger.Warn("alarm device_id parse failed",
				zap.String("event_type", al.EventType),
				zap.String("device_id", al.DeviceID),
				zap.Error(parseErr),
			)
			continue
		}
		trigger := map[string]interface{}{}
		if len(al.TriggerData) > 0 {
			_ = json.Unmarshal(al.TriggerData, &trigger)
		}
		trigger = ensureTriggerCardID(trigger, card.CardID)
		// AlarmCategory 走 envelope.category（eventName）；alarm_level 字段在 trigger map。
		streamID, publishErr := e.alarmBackChannel.PublishAlarmFire(
			ctx, deviceAddr, card.CardID, al.EventType, al.AlarmLevel, al.TriggeredAt.UnixMilli(), trigger)
		if publishErr != nil {
			e.logger.Error("publish alarm to stream failed",
				zap.String("event_type", al.EventType),
				zap.String("card_id", card.CardID),
				zap.Error(publishErr),
			)
			continue
		}
		alarms[i].EventID = streamID // 调用方若依赖 EventID 仅作发布回执（非 DB row id）
		e.logger.Info("alarm published to stream",
			zap.String("stream_id", streamID),
			zap.String("event_type", al.EventType),
			zap.String("alarm_level", al.AlarmLevel),
			zap.String("card_id", card.CardID),
		)
	}

	return alarms, nil
}

// ensureTriggerCardID 注入 card_id 到 trigger map（AI 虚拟设备 fallback 必需）。
func ensureTriggerCardID(m map[string]interface{}, cardID string) map[string]interface{} {
	if cardID == "" {
		return m
	}
	if m == nil {
		m = map[string]interface{}{}
	}
	if _, ok := m["card_id"]; !ok {
		m["card_id"] = cardID
	}
	return m
}

// ensureMetadataCardID 旧 alarm_events.metadata fallback。PR1 (A11) 改 publish 后不再使用，
// 函数保留作为旧契约文档；后续 PR 删除。
var _ = ensureMetadataCardID

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
