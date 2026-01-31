package aggregator

import (
	"context"
	"encoding/json"
	"fmt"
	"wisefido-card-aggregator/internal/config"
	"wisefido-card-aggregator/internal/converter"
	"wisefido-card-aggregator/internal/models"
	"wisefido-card-aggregator/internal/repository"

	"owl-common/card"

	"go.uber.org/zap"
)

// DataAggregator 数据聚合器（聚合卡片数据）
type DataAggregator struct {
	config      *config.Config
	kv          KVStore
	cardRepo    *repository.CardRepository
	logger      *zap.Logger
}

// NewDataAggregator 创建数据聚合器
func NewDataAggregator(
	cfg *config.Config,
	kv KVStore,
	cardRepo *repository.CardRepository,
	logger *zap.Logger,
) *DataAggregator {
	return &DataAggregator{
		config:      cfg,
		kv:          kv,
		cardRepo:    cardRepo,
		logger:      logger,
	}
}

// AggregateCard 聚合单个卡片的数据
// 输入：
//   - PostgreSQL: cards 表（基础信息）
//   - Redis: vital-focus:card:{card_id}:realtime（实时数据）
//   - Redis: vital-focus:card:{card_id}:alarms（报警数据）
// 输出：
//   - VitalFocusCard 对象
func (a *DataAggregator) AggregateCard(ctx context.Context, tenantID, cardID string) (*models.VitalFocusCard, error) {
	// 1. 从 PostgreSQL 读取卡片基础信息
	cardInfo, err := a.cardRepo.GetCardByID(tenantID, cardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get card info: %w", err)
	}

	// 2. 解析卡片绑定的设备和住户（从 cards.devices 和 cards.residents JSONB）
	devices, err := a.cardRepo.GetCardDevices(cardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get card devices: %w", err)
	}

	residents, err := a.cardRepo.GetCardResidents(cardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get card residents: %w", err)
	}

	// 3. 构建基础 VitalFocusCard
	vitalCard := &models.VitalFocusCard{
		CardID:        cardInfo.CardID,
		TenantID:      cardInfo.TenantID,
		CardType:      cardInfo.CardType,
		BedID:         cardInfo.BedID,
		CardName:      cardInfo.CardName,
		CardAddress:   cardInfo.CardAddress,
		Residents:     convertResidents(residents),
		Devices:       convertDevices(devices),
		DeviceCount:   len(devices),
		ResidentCount: len(residents),
	}

	// 设置 location_id（对于 Location 卡片）
	if cardInfo.CardType == "Location" {
		vitalCard.LocationID = &cardInfo.UnitID
	}

	// 设置报警统计（来自 cards 表）
	vitalCard.UnhandledAlarm0 = cardInfo.UnhandledAlarm0
	vitalCard.UnhandledAlarm1 = cardInfo.UnhandledAlarm1
	vitalCard.UnhandledAlarm2 = cardInfo.UnhandledAlarm2
	vitalCard.UnhandledAlarm3 = cardInfo.UnhandledAlarm3
	vitalCard.UnhandledAlarm4 = cardInfo.UnhandledAlarm4
	vitalCard.IconAlarmLevel = cardInfo.IconAlarmLevel
	vitalCard.PopAlarmEmerge = cardInfo.PopAlarmEmerge

	// 计算总未处理报警数
	total := 0
	if cardInfo.UnhandledAlarm0 != nil {
		total += *cardInfo.UnhandledAlarm0
	}
	if cardInfo.UnhandledAlarm1 != nil {
		total += *cardInfo.UnhandledAlarm1
	}
	if cardInfo.UnhandledAlarm2 != nil {
		total += *cardInfo.UnhandledAlarm2
	}
	if cardInfo.UnhandledAlarm3 != nil {
		total += *cardInfo.UnhandledAlarm3
	}
	if cardInfo.UnhandledAlarm4 != nil {
		total += *cardInfo.UnhandledAlarm4
	}
	if total > 0 {
		vitalCard.TotalUnhandledAlarms = &total
	}

	// 设置 primary_resident_id（对于 ActiveBed 卡片）
	if cardInfo.CardType == "ActiveBed" && cardInfo.ResidentID != nil {
		vitalCard.PrimaryResidentID = cardInfo.ResidentID
	}

	// 4. 从 Redis 读取实时数据
	realtimeData, err := a.getRealtimeData(ctx, cardID)
	if err != nil {
		a.logger.Debug("Failed to get realtime data",
			zap.String("card_id", cardID),
			zap.Error(err),
		)
		// 实时数据不存在不影响聚合，继续处理
	} else {
		// 合并实时数据
		a.mergeRealtimeData(vitalCard, realtimeData)
	}

	// 5. 从 Redis 读取报警数据
	alarms, err := a.getAlarmData(ctx, cardID)
	if err != nil {
		a.logger.Debug("Failed to get alarm data",
			zap.String("card_id", cardID),
			zap.Error(err),
		)
		// 报警数据不存在不影响聚合，继续处理
	} else {
		// 合并报警数据
		vitalCard.Alarms = convertAlarms(alarms)
	}

	return vitalCard, nil
}

// getRealtimeData 从 Redis 读取实时数据（按源结构 radar / sleepad）
func (a *DataAggregator) getRealtimeData(ctx context.Context, cardID string) (*models.RealtimeData, error) {
	key := fmt.Sprintf("vital-focus:card:%s:realtime", cardID)

	val, err := a.kv.Get(ctx, key)
	if err != nil {
		if err == ErrCacheMiss {
			return nil, fmt.Errorf("realtime data not found")
		}
		return nil, fmt.Errorf("failed to get realtime data: %w", err)
	}

	var realtimeData models.RealtimeData
	if err := json.Unmarshal([]byte(val), &realtimeData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal realtime data: %w", err)
	}

	return &realtimeData, nil
}

// getAlarmData 从 Redis 读取报警数据
func (a *DataAggregator) getAlarmData(ctx context.Context, cardID string) ([]AlarmEvent, error) {
	key := fmt.Sprintf("vital-focus:card:%s:alarms", cardID)

	val, err := a.kv.Get(ctx, key)
	if err != nil {
		if err == ErrCacheMiss {
			return nil, fmt.Errorf("alarm data not found")
		}
		return nil, fmt.Errorf("failed to get alarm data: %w", err)
	}

	var alarms []AlarmEvent
	if err := json.Unmarshal([]byte(val), &alarms); err != nil {
		return nil, fmt.Errorf("failed to unmarshal alarm data: %w", err)
	}

	return alarms, nil
}

// mergeRealtimeData 从 radar/sleepad 按源计算 display 后写入 VitalFocusCard
//
// HR / RR / sleep 独立选源、独立优先级：
// - 无新值则保持旧值，直到 device 缓存 TTL 超时（在 updateDeviceDataCache 按字段 merge 实现）。
// - 若本周期内仅一个源有该字段，用该源更新。
// - 若本周期内 Sleepad 与 Radar 都有该字段，按优先级用 Sleepad（RR_sleepad、HR_sleepad、sleep_sleepad）。
func (a *DataAggregator) mergeRealtimeData(vitalCard *models.VitalFocusCard, realtimeData *models.RealtimeData) {
	// 生命体征：HR、RR 各自独立，Sleepad 优先
	var heart, breath *int
	var heartSource, breathSource string
	if realtimeData.Sleepad != nil && realtimeData.Sleepad.Heart != nil {
		heart = realtimeData.Sleepad.Heart
		heartSource = "Sleepace"
	} else if realtimeData.Radar != nil && realtimeData.Radar.Heart != nil {
		heart = realtimeData.Radar.Heart
		heartSource = "Radar"
	}
	if realtimeData.Sleepad != nil && realtimeData.Sleepad.Breath != nil {
		breath = realtimeData.Sleepad.Breath
		breathSource = "Sleepace"
	} else if realtimeData.Radar != nil && realtimeData.Radar.Breath != nil {
		breath = realtimeData.Radar.Breath
		breathSource = "Radar"
	}
	vitalCard.Heart = heart
	vitalCard.Breath = breath
	vitalCard.HeartSource = convertSource(heartSource)
	vitalCard.BreathSource = convertSource(breathSource)

	// 睡眠状态：独立选源，Sleepad 优先
	var sleepStage *string
	if realtimeData.Sleepad != nil && realtimeData.Sleepad.SleepStatus != nil {
		sleepStage = realtimeData.Sleepad.SleepStatus
	} else if realtimeData.Radar != nil && realtimeData.Radar.SleepStatus != nil {
		sleepStage = realtimeData.Radar.SleepStatus
	}
	if sleepStage != nil {
		s := converter.SleepStage(*sleepStage)
		vitalCard.SleepStage = &s
		vitalCard.SleepStateSNOMEDCode = sleepStage
		vitalCard.SleepStateDisplay = converter.SleepStateDisplay(*sleepStage)
	}

	// 床状态 display：sleepad.bed_status ?? radar.bed_status
	var bedStatus *string
	if realtimeData.Sleepad != nil && realtimeData.Sleepad.BedStatus != nil {
		bedStatus = realtimeData.Sleepad.BedStatus
	} else if realtimeData.Radar != nil && realtimeData.Radar.BedStatus != nil {
		bedStatus = realtimeData.Radar.BedStatus
	}
	if bedStatus != nil {
		b := converter.BedStatus(*bedStatus)
		vitalCard.BedStatus = &b
	}

	// 姿态数据（即可以是 Location 卡片，也可以是 ActiveBed）
	vitalCard.PersonCount = intPtr(realtimeData.PersonCount)
	if len(realtimeData.Postures) > 0 {
		postures := make([]int, 0, len(realtimeData.Postures))
		for _, posture := range realtimeData.Postures {
			stage := converter.PostureStage(posture.PostureCode)
			if stage > 0 {
				postures = append(postures, stage)
			}
		}
		if len(postures) > 0 {
			vitalCard.Postures = postures
		}
	}
}

// AlarmEvent 报警事件（与 wisefido-ai 保持一致）
type AlarmEvent struct {
	EventID         string                 `json:"event_id"`
	EventType       string                 `json:"event_type"`
	Category        string                 `json:"category"`
	AlarmLevel      string                 `json:"alarm_level"`
	AlarmStatus     string                 `json:"alarm_status"`
	TriggeredAt     int64                  `json:"triggered_at"` // Unix timestamp
	TriggeredBy     *string                `json:"triggered_by,omitempty"`
	TriggerData     map[string]interface{} `json:"trigger_data,omitempty"`
	IoTTimeSeriesID *int64                 `json:"iot_timeseries_id,omitempty"`
}

// 辅助函数：转换数据格式

func convertResidents(residents []card.ResidentInfo) []models.CardResident {
	result := make([]models.CardResident, 0, len(residents))
	for _, r := range residents {
		result = append(result, models.CardResident{
			ResidentID: r.ResidentID,
			Nickname:   r.Nickname,
			UnitID:     r.UnitID,
			BedID:      r.BedID,
		})
	}
	return result
}

func convertDevices(devices []card.DeviceInfo) []models.CardDevice {
	result := make([]models.CardDevice, 0, len(devices))
	for _, d := range devices {
		result = append(result, models.CardDevice{
			DeviceID:    d.DeviceID,
			DeviceName:  d.DeviceName,
			DeviceType:  d.DeviceType,
			DeviceModel: d.DeviceModel,
			BedID:       d.BoundBedID,
			BedName:     d.BedName,
			RoomID:      d.BoundRoomID,
			RoomName:    d.RoomName,
			UnitID:      d.UnitID,
		})
	}
	return result
}

func convertAlarms(alarms []AlarmEvent) []models.AlarmItem {
	result := make([]models.AlarmItem, 0, len(alarms))
	for _, a := range alarms {
		item := models.AlarmItem{
			EventID:         a.EventID,
			EventType:       a.EventType,
			AlarmLevel:      a.AlarmLevel,
			AlarmStatus:     a.AlarmStatus,
			TriggeredAt:     a.TriggeredAt,
			TriggeredBy:     a.TriggeredBy,
			TriggerData:     a.TriggerData,
			IoTTimeSeriesID: a.IoTTimeSeriesID,
		}
		if a.Category != "" {
			item.Category = &a.Category
		}
		result = append(result, item)
	}
	return result
}

func convertSource(source string) *string {
	if source == "" {
		return strPtr("-")
	}
	// 转换为小写：Sleepace -> s, Radar -> r
	switch source {
	case "Sleepace", "SleepPad":
		return strPtr("s")
	case "Radar":
		return strPtr("r")
	default:
		return strPtr("-")
	}
}

func intPtr(i int) *int {
	return &i
}

func strPtr(s string) *string {
	return &s
}

