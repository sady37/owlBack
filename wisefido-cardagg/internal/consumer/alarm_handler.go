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
	bedCoord  *service.BedEventCoordinator
	logger    *zap.Logger
}

func NewAlarmHandler(alarms *service.AlarmService, state *service.StateService, buffer *service.MonitorBuffer, metaCache *service.DeviceMetaCache, bedCoord *service.BedEventCoordinator, logger *zap.Logger) *AlarmHandler {
	return &AlarmHandler{
		alarms: alarms, state: state, buffer: buffer, metaCache: metaCache, bedCoord: bedCoord, logger: logger,
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
	eventName := streamEventName(m, data)
	// stream.consume 走 Debug：原始 alarm 已由 wisefido-iot 写 iot_timeseries。
	// 各 case 分支自己 Info-log 关键决策（alarm inserted / pending.added 等）。
	h.logger.Debug("stream.consume", append(streamLogFields("alarm", m, eventName),
		zap.String("status", "recv"),
	)...)
	if eventName == "" {
		h.logger.Warn("stream.consume", append(streamLogFields("alarm", m, eventName),
			zap.String("status", "drop"),
			zap.String("reason", "empty_event_name"),
		)...)
		return nil
	}
	// card_id / device_uid 已由 IotPreparedHandler 填充（若流里带 device_uid）
	if time.Now().UnixMilli()-m.Timestamp > 30_000 {
		h.logger.Info("stream.consume", append(streamLogFields("alarm", m, eventName),
			zap.String("status", "drop"),
			zap.String("reason", "stale"),
		)...)
		return nil
	}
	resolved := h.metaCache.ResolveDeviceID(ctx, m.SubjectEntity, m.DeviceID, m.DeviceUID)
	if resolved == "" {
		h.logger.Info("stream.consume", append(streamLogFields("alarm", m, eventName),
			zap.String("status", "drop"),
			zap.String("reason", "no_device_id"),
		)...)
		return nil
	}
	m.DeviceID = resolved
	dt := strings.ToLower(m.DeviceType)
	isSleepad := strings.Contains(dt, "sleepad") || strings.Contains(dt, "sleeppad")
	isRadar := strings.Contains(dt, "radar")
	// Radar 不支持的类型（仅 Sleepad）：来自 Radar 则丢弃
	if isRadar && sleepadOnlyAlarmTypes[eventName] {
		h.logger.Info("stream.consume", append(streamLogFields("alarm", m, eventName),
			zap.String("status", "drop"),
			zap.String("reason", "radar_not_support"),
		)...)
		return nil
	}
	payload := &redis.IoTStreamMessage{
		Producer:      m.Producer,
		SubjectEntity: m.SubjectEntity,
		DeviceUID:     m.DeviceUID,
		DeviceType:    m.DeviceType,
		TenantID:      m.TenantID,
		Timestamp:     m.Timestamp,
		TopicType:     m.TopicType,
		Category:      m.Category,
		DataValue:     []interface{}{data},
	}
	payload.DeviceID = m.DeviceID
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
			if dm := h.metaCache.GetDeviceMeta(ctx, m.SubjectEntity, m.DeviceID); dm != nil {
				deviceType = dm.DeviceType
			} else if m.DeviceUID != "" {
				if dm := h.metaCache.GetDeviceMetaByUID(ctx, m.SubjectEntity, m.DeviceUID); dm != nil {
					deviceType = dm.DeviceType
				}
			}
			skip := false
			if h.bedCoord != nil {
				skip, _ = h.bedCoord.InBed(ctx, h.state, h.alarms, h.metaCache, h.buffer, m.SubjectEntity, m.TenantID, m.DeviceID, m.DeviceUID, deviceType, m.Timestamp, func(wr bool) {
					if wr {
						_ = h.alarms.RemovePendingAlarm(ctx, m.TenantID, m.SubjectEntity, payload.DeviceID, alarm.LeftBed)
					}
				})
			}
			if !skip {
				trackOverride := sleepadTrackOverrideForInBed(ctx, h.metaCache, h.buffer, m.SubjectEntity)
				written, _ := h.state.PublishBedStateFromEvent(ctx, m.SubjectEntity, alarm.InBed, deviceType, m.Timestamp, 0, trackOverride)
				if written {
					_ = h.alarms.RemovePendingAlarm(ctx, m.TenantID, m.SubjectEntity, payload.DeviceID, alarm.LeftBed)
				}
				_ = h.state.ReconcileRoomStateFromBedState(ctx, m.SubjectEntity)
			}
		}
	case alarm.LeftBed:
		if enabled && level != "" {
			if err := h.alarms.PersistAlarmAndPublish(ctx, payload, eventName, level); err != nil {
				return err
			}
		}
		if h.state != nil {
			deviceType := ""
			if dm := h.metaCache.GetDeviceMeta(ctx, m.SubjectEntity, m.DeviceID); dm != nil {
				deviceType = dm.DeviceType
			} else if m.DeviceUID != "" {
				if dm := h.metaCache.GetDeviceMetaByUID(ctx, m.SubjectEntity, m.DeviceUID); dm != nil {
					deviceType = dm.DeviceType
				}
			}
			skip := false
			if h.bedCoord != nil {
				skip, _ = h.bedCoord.LeftBed(ctx, h.state, h.metaCache, h.buffer, m.SubjectEntity, m.TenantID, m.DeviceID, m.DeviceUID, deviceType, m.Timestamp, func() {
					meta := h.metaCache.GetOrLoad(ctx, m.SubjectEntity)
					if meta == nil {
						return
					}
					if meta.TenantID != "" && h.alarms != nil {
						service.PersistSuspectedFallPoseLyingIfEnabled(ctx, h.alarms, m.SubjectEntity, meta.TenantID, h.buffer, meta, "ImmediateLeftBedFall_lying_coord")
					}
					if radarID := service.RadarDeviceIDBoundToBed(meta); radarID != "" {
						service.StartLeftBedFall(m.SubjectEntity, radarID)
					}
				})
			}
			if !skip {
				trackOverride := sleepadTrackOverride(ctx, h.metaCache, h.buffer, m.SubjectEntity)
				pubWritten, _ := h.state.PublishBedStateFromEvent(ctx, m.SubjectEntity, alarm.LeftBed, deviceType, m.Timestamp, 0, trackOverride)
				if pubWritten {
					_ = h.state.ReconcileRoomStateFromBedState(ctx, m.SubjectEntity)
					meta := h.metaCache.GetOrLoad(ctx, m.SubjectEntity)
					if meta != nil {
						if meta.TenantID != "" && h.alarms != nil {
							service.PersistSuspectedFallPoseLyingIfEnabled(ctx, h.alarms, m.SubjectEntity, meta.TenantID, h.buffer, meta, "ImmediateLeftBedFall_lying_alarm")
						}
						if radarID := service.RadarDeviceIDBoundToBed(meta); radarID != "" {
							service.StartLeftBedFall(m.SubjectEntity, radarID)
						}
					}
				}
			}
		}

	// 设备类抖动信号：SignalPoor / AngleException 与 Offline / SensorDetached 同等处理 ——
	// onset 入库报警，*Recover 触发 auto-recover 解除 active alarm（无 active 时 noop，幂等）。
	// 同时把当前真值写到 device:status:{deviceID} hash，前端无需穿越 alarm_events 反推标志位。
	// 原则：设备类报警一律自动恢复，不需人工标记。
	case alarm.SignalPoor:
		if m.DeviceID != "" && h.state != nil {
			_ = h.state.PatchDeviceFlag(ctx, m.DeviceID, "signal_poor", 1)
		}
		if enabled && level != "" {
			if err := h.alarms.PersistAlarmAndPublish(ctx, payload, eventName, level); err != nil {
				return err
			}
		}
	case alarm.AngleException:
		if m.DeviceID != "" && h.state != nil {
			_ = h.state.PatchDeviceFlag(ctx, m.DeviceID, "angle_abnormal", 1)
		}
		if enabled && level != "" {
			if err := h.alarms.PersistAlarmAndPublish(ctx, payload, eventName, level); err != nil {
				return err
			}
		}
	case alarm.SingalPoorRecover:
		if m.DeviceID != "" && h.state != nil {
			_ = h.state.PatchDeviceFlag(ctx, m.DeviceID, "signal_poor", 0)
		}
		_ = h.alarms.HandleRecoveryWithTypes(ctx, payload, []string{alarm.SignalPoor})
	case alarm.AngleExceptionRecover:
		if m.DeviceID != "" && h.state != nil {
			_ = h.state.PatchDeviceFlag(ctx, m.DeviceID, "angle_abnormal", 0)
		}
		_ = h.alarms.HandleRecoveryWithTypes(ctx, payload, []string{alarm.AngleException})

	// Sleepad 行为/状态恢复型 alarm（BedSitUp 床上坐起 / NoTurnOver 长时不翻身）：
	// sleepace 协议用同一 EventName + event_status="start"/"end" 配对。
	// 临床语义上设备 end = 行为信号本身已消失（坐回 / 翻身了），等价 device-driven recovery。
	// onset(start) → 落库推送；relieve(end) → auto-resolve，否则会永久 latch active。
	case alarm.BedSitUp, alarm.NoTurnOver:
		if !isSleepad {
			break
		}
		eventStatus, _ := data["event_status"].(string)
		if eventStatus == "end" {
			_ = h.alarms.HandleRecoveryWithTypes(ctx, payload, []string{eventName})
			break
		}
		if enabled && level != "" {
			if err := h.alarms.PersistAlarmAndPublish(ctx, payload, eventName, level); err != nil {
				return err
			}
		}

	// Sleepad 风险事件型 alarm（AbnormalBodyMovement 异动抽搐 / NoBodyMove 长时无体动）：
	// 同样有 event_status="start"/"end" 配对，但临床语义上设备 end ≠ 病情已评估，
	// 必须保持 active 等护理人员人工 ack（与 Fall/SittingOnGround 同级）。
	// onset(start) → 落库推送；relieve(end) → 仅记日志留痕，不改 active 状态。
	case alarm.AbnormalBodyMovement, alarm.NoBodyMove:
		if !isSleepad {
			break
		}
		eventStatus, _ := data["event_status"].(string)
		if eventStatus == "end" {
			h.logger.Info("sleepace_relieve_signal_kept_active",
				zap.String("cid", m.SubjectEntity),
				zap.String("device_id", m.DeviceID),
				zap.String("event_type", eventName),
				zap.String("reason", "high_risk_requires_manual_ack"),
			)
			break
		}
		if enabled && level != "" {
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
			h.buffer.RemoveDevice(m.SubjectEntity, m.DeviceID)
			// 负向标记 device:status：alarm Offline 显式 offline=1（不刷 last_seen_ms——
			// 设备死了不是它活着的证据）。看门狗 fail-safe 兜底 alarm 流自身故障的场景。
			if m.DeviceID != "" && h.state != nil {
				_ = h.state.SetDeviceOnline(ctx, m.DeviceID, m.DeviceUID, m.DeviceType, false)
			}
		}
	case alarm.AlarmTypeOfflineRecover:
		// 三段处理（清旧 alarm + 设 online + 状态归位）合并成单条 done 日志。
		// 失败路径仍单独 Warn 出来。
		recoveryOK := true
		if err := h.alarms.HandleRecoveryWithTypes(ctx, payload, []string{alarm.AlarmTypeOffline, alarm.AlarmTypeDeviceFailure}); err != nil {
			recoveryOK = false
			h.logger.Warn("offline_recover.recovery_failed",
				zap.String("device_uid", m.DeviceUID),
				zap.String("card_id", m.SubjectEntity),
				zap.Error(err),
			)
		}
		setOnlineOK, setOnlineSkip := false, false
		if m.DeviceID == "" || h.state == nil {
			setOnlineSkip = true
		} else if err := h.state.SetDeviceOnline(ctx, m.DeviceID, m.DeviceUID, m.DeviceType, true); err != nil {
			h.logger.Warn("offline_recover.set_online_failed",
				zap.String("device_uid", m.DeviceUID),
				zap.String("card_id", m.SubjectEntity),
				zap.Error(err),
			)
		} else {
			setOnlineOK = true
		}
		h.logger.Info("offline_recover.done",
			zap.String("device_uid", m.DeviceUID),
			zap.String("card_id", m.SubjectEntity),
			zap.String("device_type", m.DeviceType),
			zap.Bool("recovery_ok", recoveryOK),
			zap.Bool("set_online_ok", setOnlineOK),
			zap.Bool("set_online_skipped", setOnlineSkip),
			zap.Int64("ts", m.Timestamp),
		)
	// DeviceRecover：其它故障恢复；连通性上线由 Sleepace/qinglan 发 OfflineRecover。
	case alarm.AlarmTypeDeviceRecover:
		_ = h.alarms.HandleRecoveryWithTypes(ctx, payload, []string{alarm.AlarmTypeDeviceFailure})
		if enabled && level != "" {
			if err := h.alarms.PersistAlarmAndPublish(ctx, payload, eventName, level); err != nil {
				return err
			}
		}
		if m.DeviceID != "" && h.state != nil {
			_ = h.state.SetDeviceOnline(ctx, m.DeviceID, m.DeviceUID, m.DeviceType, true)
		}
	case alarm.SensorDetached:
		if m.DeviceID != "" && h.state != nil {
			_ = h.state.PatchDeviceFlag(ctx, m.DeviceID, "sensor_detached", 1)
		}
		if enabled && level != "" {
			if err := h.alarms.PersistAlarmAndPublish(ctx, payload, eventName, level); err != nil {
				return err
			}
		}
	// status=1 时 gateway 发 SensorDetachedRecover；此处仅恢复该设备既有 SensorDetached，不落新告警；同时清除 device:status hash 标志位。
	case alarm.SensorDetachedRecover:
		if m.DeviceID != "" && h.state != nil {
			_ = h.state.PatchDeviceFlag(ctx, m.DeviceID, "sensor_detached", 0)
		}
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
