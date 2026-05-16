package consumer

import (
	"context"
	"encoding/json"
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
	ac := service.AddrCtxFromMsg(m)
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
	resolved := h.metaCache.ResolveDeviceID(ctx, m.SubjectEntity, ac.DeviceAddr, ac.DeviceAddr)
	if resolved == "" {
		h.logger.Info("stream.consume", append(streamLogFields("alarm", m, eventName),
			zap.String("status", "drop"),
			zap.String("reason", "no_device_id"),
		)...)
		return nil
	}
	ac.DeviceAddr = resolved
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
		DeviceAddr:    m.DeviceAddr,
		DeviceType:    m.DeviceType,
		Timestamp:     m.Timestamp,
		TopicType:     m.TopicType,
		Category:      m.Category,
		DataValue:     []interface{}{data},
	}
	_, level, _, _, _, enabled := h.alarms.ResolveEnablementByDevice(ctx, ac.TenantPref, ac.DeviceAddr, eventName)

	// PR1 (A9): sensor → cardagg 回流通道，event_status=pending_arm/pending_cancel。
	// 协议见 wisefido-sensor/internal/consumer/alarm_back_channel.go（同期 PR1）。
	// 由 sensor 端 Stay / LeftBed-timer / Suspected→Real 升格 等延迟报警决策时触发。
	// cardagg 不参与决策，仅按 sensor 给定 alarmType/level/duration 套用现有 pending 机制。
	if eventStatus, _ := data["event_status"].(string); eventStatus == "pending_arm" || eventStatus == "pending_cancel" {
		alarmType := eventName
		if alarmType == "" {
			if v, ok := data["event_name"].(string); ok {
				alarmType = v
			}
		}
		if alarmType == "" {
			h.logger.Warn("pending.signal.drop", append(streamLogFields("alarm", m, eventName),
				zap.String("reason", "empty_alarm_type"),
				zap.String("event_status", eventStatus),
			)...)
			return nil
		}
		switch eventStatus {
		case "pending_arm":
			alarmLevel, _ := data["alarm_level"].(string)
			durationSec := intFromAny(data["duration_sec"])
			upgradeTo, _ := data["upgrade_to"].(string)
			eventSinceMs := int64(intFromAny(data["event_since_ms"]))
			if eventSinceMs == 0 {
				eventSinceMs = m.Timestamp
			}
			triggerData, _ := json.Marshal(data)
			if err := h.alarms.AddPendingAlarm(ctx, ac.TenantPref, ac.DeviceAddr, alarmType, alarmLevel, eventSinceMs, durationSec, upgradeTo, triggerData); err != nil {
				h.logger.Warn("pending.arm.failed", append(streamLogFields("alarm", m, eventName),
					zap.String("alarm_type", alarmType),
					zap.Error(err))...)
				return err
			}
			h.logger.Info("pending.arm.from_sensor",
				zap.String("cid", m.SubjectEntity),
				zap.String("device", ac.DeviceAddr),
				zap.String("alarm_type", alarmType),
				zap.String("level", alarmLevel),
				zap.Int("duration_sec", durationSec),
				zap.String("upgrade_to", upgradeTo),
			)
		case "pending_cancel":
			if err := h.alarms.RemovePendingAlarm(ctx, ac.TenantPref, m.SubjectEntity, ac.DeviceAddr, alarmType); err != nil {
				h.logger.Warn("pending.cancel.failed", append(streamLogFields("alarm", m, eventName),
					zap.String("alarm_type", alarmType),
					zap.Error(err))...)
				return err
			}
			h.logger.Info("pending.cancel.from_sensor",
				zap.String("cid", m.SubjectEntity),
				zap.String("device", ac.DeviceAddr),
				zap.String("alarm_type", alarmType),
			)
		}
		return nil
	}

	// 集中处理 event_status=end（持续事件解除）：协议层 end 永远不该被当作新报警插入 alarm_events。
	// 仅对 Registry 中 EndPolicy 显式标注的类型拦截，避免误拦截 *Recover 类型自身的 end 语义
	// （OfflineRecover/SignalPoorRecover 等 recovery 事件本身需要进入各自 case 分支处理）。
	// 后续 switch case 只关心 start/instant/pending，不需要再检查 event_status。
	if eventStatus, _ := data["event_status"].(string); eventStatus == "end" {
		if def := alarm.LookupAlarm(eventName); def != nil {
			switch def.EndPolicy {
			case alarm.EndPolicyAutoResolve:
				h.logger.Info("alarm.end.auto_resolve",
					zap.String("cid", m.SubjectEntity),
					zap.String("device_id", ac.DeviceAddr),
					zap.String("type", eventName),
				)
				return h.alarms.HandleRecoveryWithTypes(ctx, payload, []string{eventName})
			case alarm.EndPolicyManualAck:
				h.logger.Info("alarm.end.manual_ack_required",
					zap.String("cid", m.SubjectEntity),
					zap.String("device_id", ac.DeviceAddr),
					zap.String("type", eventName),
					zap.String("reason", "high_risk_requires_manual_ack"),
				)
				return nil
			}
			// EndPolicyIgnore（零值）→ fall through 到下方 switch 让 case 自行处理
		}
	}

	switch eventName {

	case alarm.Fall, alarm.SittingOnGround:
		// Fall 是高优先级事件，无论 enabled 如何都打 Info 留迹（含 trigger 上下文便于追溯）。
		// gateway 已把 pose / track_id 等业务字段拍平到 data 顶层（Plan B），消费方直接读。
		h.logger.Info("fall.event.recv",
			zap.String("cid", m.SubjectEntity),
			zap.String("device_id", ac.DeviceAddr),
			zap.String("device_uid", ac.DeviceAddr),
			zap.String("type", eventName),
			zap.Int("track_id", intFromAny(data["track_id"])),
			zap.Int("pose", intFromAny(data["pose"])),
			zap.Int("track_confidence", intFromAny(data["track_confidence"])),
			zap.String("ai_source", stringFromAny(data["ai_source"])),
			zap.Bool("enabled", enabled),
			zap.String("level", level),
		)
		if enabled && level != "" {
			if err := h.alarms.PersistAlarmAndPublish(ctx, payload, eventName, level); err != nil {
				return err
			}
		}
	// InBed/LeftBed 是 sleepace 直接发的 alarm：先使能检查、报警，再与 event_handler 一致更新 bed status
	case alarm.InBed:
		h.logger.Info("bed.alarm.recv",
			zap.String("cid", m.SubjectEntity),
			zap.String("device_id", ac.DeviceAddr),
			zap.String("device_uid", ac.DeviceAddr),
			zap.String("type", eventName),
			zap.String("source", "sleepace_alarm"),
			zap.Bool("enabled", enabled),
			zap.String("level", level),
		)
		if enabled && level != "" {
			if err := h.alarms.PersistAlarmAndPublish(ctx, payload, eventName, level); err != nil {
				return err
			}
		}
		if h.state != nil {
			deviceType := ""
			if dm := h.metaCache.GetDeviceMeta(ctx, m.SubjectEntity, ac.DeviceAddr); dm != nil {
				deviceType = dm.DeviceType
			} else if ac.DeviceAddr != "" {
				if dm := h.metaCache.GetDeviceMetaByUID(ctx, m.SubjectEntity, ac.DeviceAddr); dm != nil {
					deviceType = dm.DeviceType
				}
			}
			skip := false
			if h.bedCoord != nil {
				skip, _ = h.bedCoord.InBed(ctx, h.state, h.alarms, h.metaCache, h.buffer, m.SubjectEntity, ac.TenantPref, ac.DeviceAddr, ac.DeviceAddr, deviceType, m.Timestamp, func(wr bool) {
					if wr {
						_ = h.alarms.RemovePendingAlarm(ctx, ac.TenantPref, m.SubjectEntity, ac.DeviceAddr, alarm.LeftBed)
					}
				})
			}
			if !skip {
				trackOverride := sleepadTrackOverrideForInBed(ctx, h.metaCache, h.buffer, m.SubjectEntity)
				written, _ := h.state.PublishBedStateFromEvent(ctx, m.SubjectEntity, alarm.InBed, deviceType, m.Timestamp, 0, trackOverride)
				if written {
					_ = h.alarms.RemovePendingAlarm(ctx, ac.TenantPref, m.SubjectEntity, ac.DeviceAddr, alarm.LeftBed)
				}
			}
		}
	case alarm.LeftBed:
		h.logger.Info("bed.alarm.recv",
			zap.String("cid", m.SubjectEntity),
			zap.String("device_id", ac.DeviceAddr),
			zap.String("device_uid", ac.DeviceAddr),
			zap.String("type", eventName),
			zap.String("source", "sleepace_alarm"),
			zap.Bool("enabled", enabled),
			zap.String("level", level),
		)
		if enabled && level != "" {
			if err := h.alarms.PersistAlarmAndPublish(ctx, payload, eventName, level); err != nil {
				return err
			}
		}
		if h.state != nil {
			deviceType := ""
			if dm := h.metaCache.GetDeviceMeta(ctx, m.SubjectEntity, ac.DeviceAddr); dm != nil {
				deviceType = dm.DeviceType
			} else if ac.DeviceAddr != "" {
				if dm := h.metaCache.GetDeviceMetaByUID(ctx, m.SubjectEntity, ac.DeviceAddr); dm != nil {
					deviceType = dm.DeviceType
				}
			}
			skip := false
			if h.bedCoord != nil {
				skip, _ = h.bedCoord.LeftBed(ctx, h.state, h.metaCache, h.buffer, m.SubjectEntity, ac.TenantPref, ac.DeviceAddr, ac.DeviceAddr, deviceType, m.Timestamp, nil)
			}
			if !skip {
				trackOverride := sleepadTrackOverride(ctx, h.metaCache, h.buffer, m.SubjectEntity)
				_, _ = h.state.PublishBedStateFromEvent(ctx, m.SubjectEntity, alarm.LeftBed, deviceType, m.Timestamp, 0, trackOverride)
			}
		}

	// 设备类抖动信号：SignalPoor / AngleException 与 Offline / SensorDetached 同等处理 ——
	// onset 入库报警，*Recover 触发 auto-recover 解除 active alarm（无 active 时 noop，幂等）。
	// 同时把当前真值写到 device:status:{deviceID} hash，前端无需穿越 alarm_events 反推标志位。
	// 原则：设备类报警一律自动恢复，不需人工标记。
	case alarm.SignalPoor:
		if ac.DeviceAddr != "" && h.state != nil {
			_ = h.state.PatchDeviceFlag(ctx, ac.DeviceAddr, "signal_poor", 1)
		}
		if enabled && level != "" {
			if err := h.alarms.PersistAlarmAndPublish(ctx, payload, eventName, level); err != nil {
				return err
			}
		}
	case alarm.AngleException:
		if ac.DeviceAddr != "" && h.state != nil {
			_ = h.state.PatchDeviceFlag(ctx, ac.DeviceAddr, "angle_abnormal", 1)
		}
		if enabled && level != "" {
			if err := h.alarms.PersistAlarmAndPublish(ctx, payload, eventName, level); err != nil {
				return err
			}
		}
	case alarm.SingalPoorRecover:
		if ac.DeviceAddr != "" && h.state != nil {
			_ = h.state.PatchDeviceFlag(ctx, ac.DeviceAddr, "signal_poor", 0)
		}
		_ = h.alarms.HandleRecoveryWithTypes(ctx, payload, []string{alarm.SignalPoor})
	case alarm.AngleExceptionRecover:
		if ac.DeviceAddr != "" && h.state != nil {
			_ = h.state.PatchDeviceFlag(ctx, ac.DeviceAddr, "angle_abnormal", 0)
		}
		_ = h.alarms.HandleRecoveryWithTypes(ctx, payload, []string{alarm.AngleException})

	// Sleepad 行为/状态恢复型 alarm（BedSitUp 床上坐起 / NoTurnOver 长时不翻身）：
	// onset(start) → 落库推送；relieve(end) 由入口 EndPolicyAutoResolve 集中处理，不会进入此 case。
	case alarm.BedSitUp, alarm.NoTurnOver:
		if !isSleepad {
			break
		}
		if enabled && level != "" {
			if err := h.alarms.PersistAlarmAndPublish(ctx, payload, eventName, level); err != nil {
				return err
			}
		}

	// Sleepad 风险事件型 alarm（AbnormalBodyMovement 异动抽搐 / NoBodyMove 长时无体动）：
	// onset(start) → 落库推送；relieve(end) 由入口 EndPolicyManualAck 集中处理，仅记日志保持 active。
	case alarm.AbnormalBodyMovement, alarm.NoBodyMove:
		if !isSleepad {
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
			if enabled && level != "" && h.alarms.AHIQueryReady() && ac.DeviceAddr != "" && h.alarms.CheckAH(ctx, ac.DeviceAddr) {
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
		if ac.DeviceAddr != "" {
			h.buffer.RemoveDevice(m.SubjectEntity, ac.DeviceAddr)
			// 负向标记 device:status：alarm Offline 显式 offline=1（不刷 last_seen_ms——
			// 设备死了不是它活着的证据）。看门狗 fail-safe 兜底 alarm 流自身故障的场景。
			if ac.DeviceAddr != "" && h.state != nil {
				_ = h.state.SetDeviceOnline(ctx, ac.DeviceAddr, ac.DeviceAddr, m.DeviceType, false)
			}
		}
	case alarm.AlarmTypeOfflineRecover:
		// 三段处理（清旧 alarm + 设 online + 状态归位）合并成单条 done 日志。
		// 失败路径仍单独 Warn 出来。
		recoveryOK := true
		if err := h.alarms.HandleRecoveryWithTypes(ctx, payload, []string{alarm.AlarmTypeOffline, alarm.AlarmTypeDeviceFailure}); err != nil {
			recoveryOK = false
			h.logger.Warn("offline_recover.recovery_failed",
				zap.String("device_uid", ac.DeviceAddr),
				zap.String("card_id", m.SubjectEntity),
				zap.Error(err),
			)
		}
		setOnlineOK, setOnlineSkip := false, false
		if ac.DeviceAddr == "" || h.state == nil {
			setOnlineSkip = true
		} else if err := h.state.SetDeviceOnline(ctx, ac.DeviceAddr, ac.DeviceAddr, m.DeviceType, true); err != nil {
			h.logger.Warn("offline_recover.set_online_failed",
				zap.String("device_uid", ac.DeviceAddr),
				zap.String("card_id", m.SubjectEntity),
				zap.Error(err),
			)
		} else {
			setOnlineOK = true
		}
		h.logger.Info("offline_recover.done",
			zap.String("device_uid", ac.DeviceAddr),
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
		if ac.DeviceAddr != "" && h.state != nil {
			_ = h.state.SetDeviceOnline(ctx, ac.DeviceAddr, ac.DeviceAddr, m.DeviceType, true)
		}
	case alarm.SensorDetached:
		if ac.DeviceAddr != "" && h.state != nil {
			_ = h.state.PatchDeviceFlag(ctx, ac.DeviceAddr, "sensor_detached", 1)
		}
		if enabled && level != "" {
			if err := h.alarms.PersistAlarmAndPublish(ctx, payload, eventName, level); err != nil {
				return err
			}
		}
	// status=1 时 gateway 发 SensorDetachedRecover；此处仅恢复该设备既有 SensorDetached，不落新告警；同时清除 device:status hash 标志位。
	case alarm.SensorDetachedRecover:
		if ac.DeviceAddr != "" && h.state != nil {
			_ = h.state.PatchDeviceFlag(ctx, ac.DeviceAddr, "sensor_detached", 0)
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
