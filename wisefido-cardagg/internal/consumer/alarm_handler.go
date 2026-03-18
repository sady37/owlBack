package consumer

import (
	"context"
	"strings"
	"time"

	"owl-common/alarm"
	"owl-common/redis"
	"wisefido-cardagg/internal/service"

	"go.uber.org/zap"
)

// Sleepad 专有报警类型，Radar 不支持；来自 Radar 的此类报警直接丢弃。（InBed/LeftBed 已统一走 alarm，不在此丢弃）
var sleepadOnlyAlarmTypes = map[string]bool{
	alarm.AbnormalBodyMovement: true,
	alarm.NoTurnOver:           true,
}

// AlarmHandler 消费 iot:alarm:stream，落库并更新 card 告警状态；设备类告警时同步更新 buffer/metaCache。
// InBed/LeftBed：Step1 PersistAlarmAndPublish（或 pending），Step2 PublishBedStateFromEvent，Step3 ReconcileRoomStateFromBedState。
type AlarmHandler struct {
	alarms    *service.AlarmService
	state     *service.StateService
	buffer    *service.MonitorBuffer
	metaCache *service.DeviceMetaCache
	resolver  *service.DeviceCardResolver
	logger    *zap.Logger
}

func NewAlarmHandler(alarms *service.AlarmService, state *service.StateService, buffer *service.MonitorBuffer, metaCache *service.DeviceMetaCache, resolver *service.DeviceCardResolver, logger *zap.Logger) *AlarmHandler {
	return &AlarmHandler{
		alarms: alarms, state: state, buffer: buffer, metaCache: metaCache, resolver: resolver, logger: logger,
	}
}

func (h *AlarmHandler) Handle(ctx context.Context, msg interface{}) error {
	raw, ok := msg.(map[string]interface{})
	if !ok {
		return nil
	}
	m, err := ParseMessage(raw)
	if err != nil {
		h.logger.Warn("parse alarm", zap.Error(err))
		return nil
	}
	data := redis.FirstDataValue(m.DataValue)
	if data == nil {
		data = make(map[string]interface{})
	}
	eventName, _ := data[alarm.FieldEventName].(string)
	if eventName == "" {
		return nil
	}
	// card_id / device_uid 已由 IotPreparedHandler 填充（若流里带 device_uid）
	if time.Now().UnixMilli()-m.Timestamp > 30_000 {
		h.logger.Debug("alarm dropped (stale)", zap.String("cid", m.CardID), zap.String("event", eventName))
		return nil
	}
	dt := strings.ToLower(m.DeviceType)
	isSleepad := strings.Contains(dt, "sleepad") || strings.Contains(dt, "sleeppad")
	isRadar := strings.Contains(dt, "radar")
	// Radar 不支持的类型（仅 Sleepad）：来自 Radar 则丢弃
	if isRadar && sleepadOnlyAlarmTypes[eventName] {
		h.logger.Debug("alarm dropped (radar does not support)", zap.String("event", eventName))
		return nil
	}
	payload := &redis.IoTStreamMessage{
		DeviceUID:  m.DeviceUID,
		DeviceType: m.DeviceType,
		CardID:     m.CardID,
		TenantID:   m.TenantID,
		Timestamp:  m.Timestamp,
		TopicType:  m.TopicType,
		Category:   m.Category,
		DataValue:  []interface{}{data},
	}
	if meta := h.metaCache.GetOrLoad(ctx, m.CardID); meta != nil {
		dm := meta.Devices[m.DeviceID]
		if dm == nil && m.DeviceUID != "" {
			dm = h.metaCache.GetDeviceMetaByUID(ctx, m.CardID, m.DeviceUID)
		}
		if dm != nil {
			payload.DeviceID = dm.DeviceID
		}
	}
	_, level, _, _, _, enabled := h.alarms.ResolveEnablementByDevice(ctx, m.TenantID, payload.DeviceID, eventName)

	switch eventName {

	case alarm.Fall, alarm.SittingOnGround:
		if enabled && level != "" {
			if err := h.alarms.PersistAlarmAndPublish(ctx, payload, eventName, level); err != nil {
				return err
			}
		}
	// InBed/LeftBed 是 sleepace 直接发的 alarm：先使能检查、报警，再与 event_handler 一致更新 bed status
	case alarm.InBed:
		if enabled && level != "" {
			if err := h.alarms.PersistAlarmAndPublish(ctx, payload, eventName, level); err != nil {
				return err
			}
		}
		if h.state != nil {
			deviceType := ""
			if dm := h.metaCache.GetDeviceMeta(ctx, m.CardID, m.DeviceID); dm != nil {
				deviceType = dm.DeviceType
			} else if m.DeviceUID != "" {
				if dm := h.metaCache.GetDeviceMetaByUID(ctx, m.CardID, m.DeviceUID); dm != nil {
					deviceType = dm.DeviceType
				}
			}
			trackOverride := sleepadTrackOverride(ctx, h.metaCache, h.buffer, m.CardID)
			written, _ := h.state.PublishBedStateFromEvent(ctx, m.CardID, alarm.InBed, deviceType, m.Timestamp, 0, trackOverride)
			if written {
				_ = h.alarms.RemovePendingAlarm(ctx, m.CardID, payload.DeviceID, alarm.LeftBed)
			}
			_ = h.state.ReconcileRoomStateFromBedState(ctx, m.CardID)
		}
	case alarm.LeftBed:
		if enabled && level != "" {
			if err := h.alarms.PersistAlarmAndPublish(ctx, payload, eventName, level); err != nil {
				return err
			}
		}
		if h.state != nil {
			deviceType := ""
			if dm := h.metaCache.GetDeviceMeta(ctx, m.CardID, m.DeviceID); dm != nil {
				deviceType = dm.DeviceType
			} else if m.DeviceUID != "" {
				if dm := h.metaCache.GetDeviceMetaByUID(ctx, m.CardID, m.DeviceUID); dm != nil {
					deviceType = dm.DeviceType
				}
			}
			trackOverride := sleepadTrackOverride(ctx, h.metaCache, h.buffer, m.CardID)
			_, _ = h.state.PublishBedStateFromEvent(ctx, m.CardID, alarm.LeftBed, deviceType, m.Timestamp, 0, trackOverride)
			_ = h.state.ReconcileRoomStateFromBedState(ctx, m.CardID)
		}
		if meta := h.metaCache.GetOrLoad(ctx, m.CardID); meta != nil {
			if radarID := service.RadarDeviceIDBoundToBed(meta); radarID != "" {
				service.StartLeftBedFall(m.CardID, radarID)
			}
		}

	// SingalPoorRecover/AngleExceptionRecover 已切到 event 流，此处丢弃
	case alarm.SingalPoorRecover, alarm.AngleExceptionRecover:
		return nil

	// Sleepad 专有：使能则落库并推送   / 异动抽搐 /翻身/体动/ 床上坐起
	case alarm.AbnormalBodyMovement, alarm.NoTurnOver, alarm.NoBodyMove, alarm.BedSitUp:
		if isSleepad && enabled && level != "" {
			if err := h.alarms.PersistAlarmAndPublish(ctx, payload, eventName, level); err != nil {
				return err
			}
		}

	case alarm.ApneaHypopnea:
		if isSleepad {
			if enabled && level != "" {
				if err := h.alarms.PersistAlarmAndPublish(ctx, payload, eventName, level); err != nil {
					return err
				}
			}
		} else if isRadar {
			if enabled && level != "" && h.alarms.AHIQueryReady() && payload.DeviceID != "" && h.alarms.CheckAH(ctx, payload.DeviceID) {
				if err := h.alarms.PersistAlarmAndPublish(ctx, payload, eventName, level); err != nil {
					return err
				}
			}
		}
	case alarm.WeakBiometricSignal:
		if err := h.alarms.PersistAlarmAndPublish(ctx, payload, eventName, level); err != nil {
			return err
		}

		// 心率呼吸：Sleepad 使能落库推送；Radar 使能落库推送，未启用 TODO
	case alarm.HeartRateAlertHigh, alarm.HeartRateAlertLow,
		alarm.RespRateAlertHigh, alarm.RespRateAlertLow:
		if err := h.alarms.PersistAlarmAndPublish(ctx, payload, eventName, level); err != nil {
			return err
		}
	//ToDo LogicTargetManage 负责将vital，WeakBiometricSignal 更新到LogicTarget

	case alarm.AlarmTypeOffline:
		if enabled && level != "" {
			if err := h.alarms.PersistAlarmAndPublish(ctx, payload, eventName, level); err != nil {
				return err
			}
		}
		if m.DeviceUID != "" {
			deviceKey := m.DeviceID
			if deviceKey == "" {
				deviceKey = m.DeviceUID
			}
			h.buffer.RemoveDevice(m.CardID, deviceKey)
			// device_status 下线时从 Redis 删除该设备（仅上线时置 Offline=0）
			if m.CardID != "" && h.state != nil {
				_ = h.state.SetDeviceOnline(ctx, m.CardID, m.DeviceID, m.DeviceUID, m.DeviceType, false)
			}
		}
	case alarm.AlarmTypeOfflineRecover:
		_ = h.alarms.HandleRecoveryWithTypes(ctx, payload, []string{alarm.AlarmTypeOffline, alarm.AlarmTypeDeviceFailure})
		if m.DeviceUID != "" && m.CardID != "" && h.state != nil {
			_ = h.state.SetDeviceOnline(ctx, m.CardID, m.DeviceID, m.DeviceUID, m.DeviceType, true)
		}
	case alarm.SensorDetached:
		if enabled && level != "" {
			if err := h.alarms.PersistAlarmAndPublish(ctx, payload, eventName, level); err != nil {
				return err
			}
		}
		//当前DeviceStatus仅包含 offline ,SensorDetached不用更新devcieStatus
	// status=1 时 gateway 发 SensorDetachedRecover；此处仅恢复该设备既有 SensorDetached，不落新告警
	case alarm.SensorDetachedRecover:
		_ = h.alarms.HandleRecoveryWithTypes(ctx, payload, []string{alarm.SensorDetached})

	case alarm.Stay, alarm.NightAbsence:
		// alarm 流收到的已满足报警条件，直接落库发布
		if enabled && level != "" {
			if err := h.alarms.PersistAlarmAndPublish(ctx, payload, eventName, level); err != nil {
				return err
			}
		}

	default:
		if enabled && level != "" {
			if err := h.alarms.PersistAlarmAndPublish(ctx, payload, eventName, level); err != nil {
				return err
			}
		}
	}
	return nil
}
