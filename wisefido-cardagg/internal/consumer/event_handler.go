package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"strings"
	"time"

	"owl-common/alarm"
	"owl-common/observation"
	"owl-common/redis"
	"wisefido-cardagg/internal/service"

	"go.uber.org/zap"
)

type EventHandler struct {
	state      *service.StateService
	alarms     *service.AlarmService
	buffer     *service.MonitorBuffer
	metaCache  *service.DeviceMetaCache
	enablement *service.AlarmEnablementCache
	bedCoord   *service.BedEventCoordinator
	aiOverrides *service.AIOverrideCache // PR6: track_verdict 收口 + Enter/ExitRoom 清理
	logger      *zap.Logger
	staySesMu   sync.Mutex
	staySessions map[string]*staySession // tenant:device -> Stay 状态机
	inBedSinceMu sync.Mutex
	inBedSince   map[string]int64 // deviceID -> InBed 事件时间戳（毫秒），LeftBed 时用于计算在床时长
}

const leftBedMinInBedMs = 5 * 60 * 1000 // 在床≥5分钟才视为稳定上床，LeftBed 才进 pending

func NewEventHandler(state *service.StateService, alarms *service.AlarmService, buffer *service.MonitorBuffer, metaCache *service.DeviceMetaCache, enablement *service.AlarmEnablementCache, bedCoord *service.BedEventCoordinator, aiOverrides *service.AIOverrideCache, logger *zap.Logger) *EventHandler {
	return &EventHandler{
		state: state, alarms: alarms, buffer: buffer, metaCache: metaCache,
		enablement: enablement, bedCoord: bedCoord, aiOverrides: aiOverrides, logger: logger,
		staySessions: make(map[string]*staySession),
		inBedSince:   make(map[string]int64),
	}
}

func (h *EventHandler) Handle(ctx context.Context, msg interface{}) error {
	raw, ok := msg.(map[string]interface{})
	if !ok {
		h.logger.Warn("event dropped (invalid type)", zap.String("type", fmt.Sprintf("%T", msg)))
		return nil
	}

	m, err := ParseMessage(raw)
	if err != nil {
		h.logger.Warn("parse event", zap.Error(err))
		return nil
	}

	if m.DeviceUID == "" {
		h.logger.Warn("event dropped (empty did)",
			zap.String("device_uid", streamFieldStr(raw, "device_uid")),
			zap.String("card_id", streamFieldStr(raw, "card_id")),
			zap.String("tenant_id", streamFieldStr(raw, "tenant_id")),
		)
		return nil
	}
	// card_id / device_uid 已由 IotPreparedHandler 填充

	if ageMs := time.Now().UnixMilli() - m.Timestamp; ageMs > 30_000 {
		h.logger.Info("stream.consume", append(streamLogFields("event", m, ""),
			zap.String("status", "drop"),
			zap.String("reason", "stale"),
			zap.Int64("age_ms", ageMs),
		)...)
		return nil
	}
	resolved := h.metaCache.ResolveDeviceID(ctx, m.CardID, m.DeviceID, m.DeviceUID)
	if resolved == "" {
		h.logger.Info("stream.consume", append(streamLogFields("event", m, ""),
			zap.String("status", "drop"),
			zap.String("reason", "no_device_id"),
		)...)
		return nil
	}
	m.DeviceID = resolved

	// radar/Sleepace 上游已按条拆分，dataValue 仅单条；取首项即可。
	data := redis.FirstDataValue(m.DataValue)
	if data == nil {
		data = make(map[string]interface{})
	}
	evName := streamEventName(m, data)
	// 常规事件（activity/track/number_people 等）用 Debug 减少日志量；
	// 报警相关事件（Fall/LeftBed/Offline 等）保留 Info
	logLevel := h.logger.Debug
	if isAlarmRelatedEvent(evName) {
		logLevel = h.logger.Info
	}
	logLevel("stream.consume", append(streamLogFields("event", m, evName),
		zap.String("status", "recv"),
	)...)
	dataJSON, _ := json.Marshal(data)
	logLevel("stream.payload",
		zap.String("stream", "event"),
		zap.String("card_id", m.CardID),
		zap.String("device_id", m.DeviceID),
		zap.String("device_uid", m.DeviceUID),
		zap.String("event", evName),
		zap.Int64("timestamp", m.Timestamp),
		zap.ByteString("data", dataJSON))

	// --- 1. 报警：有 event_name 则 ResolveEnablementByDevice + PersistAlarmAndPublish。LeftBed 由 case 先 TryAddLeftBedPendingAtTrigger，未加入 pending 再 ResolveEnablementByDevice + PersistAlarmAndPublish。---
	var alarmPayload *redis.IoTStreamMessage
	if _, hasEventName := data[alarm.FieldEventName]; hasEventName {
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
		payload.DeviceID = m.DeviceID
		alarmPayload = payload

	}

	// --- 2. Event 分支：按 event_name 分发 ---
	if evName == "" {
		h.logger.Warn("stream.consume", append(streamLogFields("event", m, evName),
			zap.String("status", "drop"),
			zap.String("reason", "empty_event_name"),
		)...)
	}
	switch evName {
	case alarm.SleepStage:
		h.routeSleepStageEvent(ctx, m, data)
	case alarm.EnterRoom:
		// PR6: 新人进 → 该 device 旧 verdict 全部作废（track_id 可能复用）
		if h.aiOverrides != nil && m.DeviceUID != "" {
			h.aiOverrides.ClearDevice(m.DeviceUID)
		}
		h.routeRoomStateEvent(ctx, m, data, evName)
	case alarm.ExitRoom:
		// PR6: 房间空了 → 清空该 device 所有 verdicts
		if h.aiOverrides != nil && m.DeviceUID != "" {
			h.aiOverrides.ClearDevice(m.DeviceUID)
		}
		h.routeRoomStateEvent(ctx, m, data, evName)
	case alarm.NumberPeople:
		h.routeRoomStateEvent(ctx, m, data, evName)

	case "track_verdict":
		// PR6: wisefido-ai 派生的 track 事后裁决（目前主要是 ghost 判定，confidence=20）
		// 收入 aiOverrides cache，后续 monitor 流合并 track 字段时套用。
		// 不参与 alarm 路径——仅影响 UI 渲染（"宁可误报不可漏报"原则）。
		if h.aiOverrides != nil && m.DeviceUID != "" {
			tid := intFromAny(data["track_id"])
			conf := intFromAny(data["track_confidence"])
			source, _ := data["source"].(string)
			reason, _ := data["reason"].(string)
			h.aiOverrides.Set(m.DeviceUID, tid, service.AIVerdict{
				Confidence: conf,
				Source:     source,
				Reason:     reason,
				UpdatedMs:  m.Timestamp,
			})
		}

	case alarm.InBed:
		// 记录 InBed 时间，供 LeftBed 计算在床时长（不依赖 BedState.StartTime，它会被 derive 覆盖）
		h.inBedSinceMu.Lock()
		h.inBedSince[m.DeviceID] = m.Timestamp
		h.inBedSinceMu.Unlock()
		if h.state != nil {
			deviceType := ""
			if dm := h.metaCache.GetDeviceMeta(ctx, m.CardID, m.DeviceID); dm != nil {
				deviceType = dm.DeviceType
			} else if m.DeviceUID != "" {
				if dm := h.metaCache.GetDeviceMetaByUID(ctx, m.CardID, m.DeviceUID); dm != nil {
					deviceType = dm.DeviceType
				}
			}
			skip := false
			if h.bedCoord != nil {
				skip, _ = h.bedCoord.InBed(ctx, h.state, h.alarms, h.metaCache, h.buffer, m.CardID, m.TenantID, m.DeviceID, m.DeviceUID, deviceType, m.Timestamp, func(wr bool) {
					if wr && alarmPayload != nil {
						_ = h.alarms.RemovePendingAlarm(ctx, m.TenantID, m.CardID, alarmPayload.DeviceID, alarm.LeftBed)
					}
				})
			}
			if !skip {
				trackOverride := sleepadTrackOverrideForInBed(ctx, h.metaCache, h.buffer, m.CardID)
				written, _ := h.state.PublishBedStateFromEvent(ctx, m.CardID, alarm.InBed, deviceType, m.Timestamp, 0, trackOverride)
				if written && alarmPayload != nil {
					_ = h.alarms.RemovePendingAlarm(ctx, m.TenantID, m.CardID, alarmPayload.DeviceID, alarm.LeftBed)
				}
				_ = h.state.ReconcileRoomStateFromBedState(ctx, m.CardID)
			}
		}
	case alarm.LeftBed:
		payload := alarmPayload
		if payload == nil {
			payload = &redis.IoTStreamMessage{CardID: m.CardID, TenantID: m.TenantID, DeviceID: m.DeviceID, DeviceUID: m.DeviceUID, DataValue: []interface{}{data}, Timestamp: m.Timestamp}
		}
		_, level, durationSec, _, _, enabled := h.alarms.ResolveEnablementByDevice(ctx, m.TenantID, payload.DeviceID, alarm.LeftBed)
		inRestWindow := h.alarms.InRestTimeWindow(ctx, m.TenantID, m.CardID)
		// 从 inBedSince 取上床时间（不依赖 BedState.StartTime，它会被 derive 覆盖）
		h.inBedSinceMu.Lock()
		inBedTs := h.inBedSince[m.DeviceID]
		delete(h.inBedSince, m.DeviceID) // 用完清除
		h.inBedSinceMu.Unlock()
		inBedMs := m.Timestamp - inBedTs
		stableInBed := inBedTs > 0 && inBedMs >= leftBedMinInBedMs
		h.logger.Info("LeftBed.pending.check",
			zap.String("device_uid", m.DeviceUID),
			zap.Bool("enabled", enabled),
			zap.Bool("inRestWindow", inRestWindow),
			zap.Bool("stableInBed", stableInBed),
			zap.Int64("inBedMs", inBedMs),
			zap.Int("durationSec", durationSec),
		)
		if enabled && level != "" && inRestWindow && stableInBed {
			if durationSec == 0 {
				_ = h.alarms.PersistAlarmAndPublish(ctx, payload, alarm.LeftBed, level)
			} else {
				triggerData, _ := json.Marshal(redis.FirstDataValue(payload.DataValue))
				_ = h.alarms.AddPendingAlarm(ctx, m.TenantID, payload.DeviceID, alarm.LeftBed, level, m.Timestamp, durationSec, "", triggerData)
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
			skip, written := false, false
			if h.bedCoord != nil {
				skip, written = h.bedCoord.LeftBed(ctx, h.state, h.metaCache, h.buffer, m.CardID, m.TenantID, m.DeviceID, m.DeviceUID, deviceType, m.Timestamp, func() {
					meta := h.metaCache.GetOrLoad(ctx, m.CardID)
					if meta == nil {
						return
					}
					if meta.TenantID != "" && h.alarms != nil {
						service.PersistSuspectedFallPoseLyingIfEnabled(ctx, h.alarms, m.CardID, meta.TenantID, h.buffer, meta, "ImmediateLeftBedFall_lying_coord")
					}
					if radarID := service.RadarDeviceIDBoundToBed(meta); radarID != "" {
						service.StartLeftBedFall(m.CardID, radarID)
					}
				})
			}
			if !skip {
				trackOverride := sleepadTrackOverride(ctx, h.metaCache, h.buffer, m.CardID)
				pubWritten, _ := h.state.PublishBedStateFromEvent(ctx, m.CardID, alarm.LeftBed, deviceType, m.Timestamp, 0, trackOverride)
				if pubWritten {
					_ = h.state.ReconcileRoomStateFromBedState(ctx, m.CardID)
					meta := h.metaCache.GetOrLoad(ctx, m.CardID)
					if meta != nil {
						if meta.TenantID != "" && h.alarms != nil {
							service.PersistSuspectedFallPoseLyingIfEnabled(ctx, h.alarms, m.CardID, meta.TenantID, h.buffer, meta, "ImmediateLeftBedFall_lying_event")
						}
						if radarID := service.RadarDeviceIDBoundToBed(meta); radarID != "" {
							service.StartLeftBedFall(m.CardID, radarID)
						}
					}
				}
			} else if !written {
				// pending：等 buffer 对齐或超时后再落离床与 StartLeftBedFall
			}
		}

	case alarm.Activity:
		h.routeActivityEvent(ctx, m, data)

	case alarm.PressureSensor:
		h.logger.Debug("pressureSensor event", zap.String("device_uid", m.DeviceUID), zap.String("category", evName))

	case alarm.SingalPoorRecover:
		if alarmPayload != nil {
			_ = h.alarms.HandleRecoveryWithTypes(ctx, alarmPayload, []string{alarm.SignalPoor})
		}
	case alarm.AngleExceptionRecover:
		if alarmPayload != nil {
			_ = h.alarms.HandleRecoveryWithTypes(ctx, alarmPayload, []string{alarm.AngleException})
		}

	case alarm.SignalPoor:
		h.metaCache.UpdateStatus(m.CardID, m.DeviceID, "signal_poor", "1")
	case alarm.AngleException:
		h.metaCache.UpdateStatus(m.CardID, m.DeviceID, "angle_abnormal", "1")

	case alarm.WarningArea:
		return nil
	default:
		// 未知事件名或 evName=="" 仅走通用字段
	}

	return nil
}

// routeRoomStateEvent EnterRoom/ExitRoom/NumberPeople → 卫生间设备写 BathRoomState，非卫生间（房间）雷达写 RoomState。HasMulti/HasRisk 规则一致；ReconcileRoomStateFromBedState 仅抬升 RoomState，不抬升 BathRoomState。
func (h *EventHandler) routeRoomStateEvent(ctx context.Context, m *redis.IoTStreamMessage, data map[string]interface{}, evName string) {
	if m.DeviceUID == "" || h.state == nil {
		return
	}
	var kind service.RoomStateEventKind
	totalPeople := -1
	switch evName {
	case alarm.EnterRoom:
		kind = service.RoomStateEventEnter
	case alarm.ExitRoom:
		kind = service.RoomStateEventExit
	case alarm.NumberPeople:
		kind = service.RoomStateEventNumberPeople
		totalPeople = intFromAny(data[observation.FieldNumberPeople])
	default:
		return
	}
	meta := h.metaCache.GetOrLoad(ctx, m.CardID)
	deviceID := m.DeviceID
	if service.IsBathroomDevice(ctx, meta, deviceID, m.TenantID, h.enablement) {
		var dm *service.DeviceMeta
		if meta != nil {
			dm = meta.Devices[deviceID]
		}
		rid, rname := service.BathroomRoomFieldsFromDevice(dm)
		_ = h.state.PublishBathRoomStateFromEvent(ctx, m.CardID, deviceID, m.DeviceUID, kind, totalPeople, m.Timestamp, rid, rname)
		switch kind {
		case service.RoomStateEventEnter:
			h.stayOnEnter(ctx, m, deviceID)
		case service.RoomStateEventExit:
			h.stayOnExit(ctx, m, deviceID)
		case service.RoomStateEventNumberPeople:
			if totalPeople < 0 {
				totalPeople = 0
			}
			h.stayOnNumberPeople(ctx, m, deviceID, totalPeople)
		}
	} else {
		_ = h.state.PublishRoomStateFromEvent(ctx, m.CardID, m.DeviceUID, kind, totalPeople, m.Timestamp)
	}
}

// routeSleepStageEvent Sleepad(sleep_stage)/Radar(Sleep) 睡眠状态事件统一入口。防静电：当前 BedState.CurrentState=out_of_bed 则 drop；若床绑 Radar+Sleepad 且该 bed 雷达未检测到人（无 track）则 drop 并打告警日志。通过后按 SleepStageSource 置信度更新 BedState.SleepStage。
func (h *EventHandler) routeSleepStageEvent(ctx context.Context, m *redis.IoTStreamMessage, data map[string]interface{}) {
	if m.DeviceUID == "" || h.state == nil {
		return
	}
	// 1. 当前卡 BedState 为 out_of_bed → drop（以聚合层状态为准，payload 的 BedStatus 有延时且 Radar 无此字段）
	curr, err := h.state.ReadCardStatus(ctx, m.CardID)
	if err == nil && curr != nil && curr.BedState != nil && curr.BedState.BedStatus == 1 {
		return
	}
	dm := h.metaCache.GetDeviceMeta(ctx, m.CardID, m.DeviceID)
	if dm == nil && m.DeviceUID != "" {
		dm = h.metaCache.GetDeviceMetaByUID(ctx, m.CardID, m.DeviceUID)
	}
	bedID := ""
	if dm != nil {
		bedID = dm.BoundBedID
	}
	meta := h.metaCache.GetOrLoad(ctx, m.CardID)
	// 2. 床绑 Radar+Sleepad 时，需该 bed 上雷达检测到人；未检测到人则 SleepStage 大概率假（干扰/设备故障）→ drop 并告警日志
	if meta != nil && bedID != "" && service.CardHasRadarAndSleepadOnBed(meta, bedID) {
		if curr == nil {
			curr, _ = h.state.ReadCardStatus(ctx, m.CardID)
		}
		hasTrack := false
		if curr != nil && curr.Target != nil {
			ts := curr.Target
			if ts.TrackID > 0 && ts.TrackID != 88 {
				hasTrack = true
			}
		}
		if !hasTrack {
			monitorSnap := h.buffer.SnapshotCard(m.CardID)
			reason := "out-of-bed, but report Sleepstage, maybe device_failure or electromagnetic interference"
			h.logger.Warn(reason,
				zap.String("card_id", m.CardID),
				zap.String("device_uid", m.DeviceUID),
				zap.String("bed_id", bedID),
				zap.Any("card_state", curr),
				zap.Any("monitor_buffer", monitorSnap),
			)
			triggerData, _ := json.Marshal(map[string]interface{}{
				"reason":         reason,
				"card_state":     curr,
				"monitor_buffer": monitorSnap,
			})
			_ = h.alarms.RecordDeviceFailure(ctx, m.CardID, m.TenantID, m.DeviceID, reason, triggerData)
			return
		}
	}
	sleep := intFromAny(data[observation.FieldSleepStage])
	if sleep == 0 || h.state == nil {
		return
	}
	source := sleepStageSourceFromDeviceType("")
	if dm != nil {
		source = sleepStageSourceFromDeviceType(dm.DeviceType)
	}
	_ = h.state.PublishBedStateSleepStage(ctx, m.CardID, sleep, source)
}

// routeActivityEvent Activity 事件：根据设备绑定地址（Bed/Bathroom/Room）更新 RoomState/BathRoomState 和 Target；卫生间且开启 Stay 时维护 alarm pending。
/*
1. 前置与解析（299–310）
无 DeviceUID 或无 state 直接返回。从 data 取出行走距离/时长、站立时长、multi_person_duration 等，用 meta 判断当前设备是否卫生间 isBathroom。

2. 状态与跌倒（312–330）
调用 UpdateStateFromActivity：按绑定区域把本次 activity 合并进 Redis 里的 RoomState / BathRoomState（含站立累计、多人时长等），并可能触发推送。
若 NeedBedFallCheck(card)，用站立/行走/轨迹等跑 LeftBedFallActivity，满足条件则落库 SuspectedFall 告警。

3. Stay pending（332–352，仅卫生间）

非卫生间：上面状态更新完就结束。
是卫生间且租户对该设备开启了 alarm.Stay：再看 Stay 窗口（由 routeRoomStateEvent 里 Enter/Exit 打开的 90s）。
窗口未开：直接 return，不加也不删 Stay pending。
窗口开：读当前卡片的 BathRoomState.TotalPeople：== 1 则 AddStayPendingIfEnabled，否则 RemovePendingAlarm(Stay)。
要点：activity 流负责改 bathroom 人数等状态；Stay pending 的增删只在「窗口仍有效」时根据当前已合并后的 total_people 做一次同步；multi_person_duration 仍传给 UpdateStateFromActivity，但不再单独用来算 Stay pending（那段已迁到「窗口 + total_people」）。
*/
// routeActivityEvent Activity 事件：根据设备绑定地址（Bed/Bathroom/Room）更新 RoomState/BathRoomState 和 Target；卫生间且开启 Stay 时由 stay_fsm 在 Enter 后 150s 武装窗内累计 activity，与 number 条件一起 AddStayPending。
func (h *EventHandler) routeActivityEvent(ctx context.Context, m *redis.IoTStreamMessage, data map[string]interface{}) {
	if m.DeviceUID == "" || h.state == nil {
		return
	}
	walkDuration := intFromAny(data[observation.FieldWalkDuration])
	walkDistance := intFromAny(data[observation.FieldWalkDistance])
	standDuration := intFromAny(data[observation.FieldStandDuration])
	multiPersonDuration := intFromAny(data[observation.FieldMultiPersonDuration])

	meta := h.metaCache.GetOrLoad(ctx, m.CardID)
	deviceID := m.DeviceID
	isBathroom := service.IsBathroomDevice(ctx, meta, deviceID, m.TenantID, h.enablement)
	var dm *service.DeviceMeta
	if meta != nil {
		dm = meta.Devices[deviceID]
	}
	rid, rname := service.BathroomRoomFieldsFromDevice(dm)

	// 行走/站立/多人阈值见 service/weights.go（WalkDistanceMetersThreshold、WalkSecThresholdOR）
	_, _, _ = h.state.UpdateStateFromActivity(ctx, m.CardID, deviceID, m.DeviceUID, isBathroom, walkDuration, walkDistance, standDuration, multiPersonDuration, m.Timestamp, rid, rname)

	if service.NeedBedFallCheck(m.CardID) {
		trackCount := intFromAny(data[observation.FieldTrackCount])
		done, suspectedFall, reportDeviceID, fallPath := service.LeftBedFallActivity(ctx, m.CardID, deviceID, standDuration, walkDuration, 0, trackCount, h.state, h.buffer, meta)
		if done && suspectedFall && reportDeviceID != "" {
			meta2 := h.metaCache.GetOrLoad(ctx, m.CardID)
			if meta2 != nil && meta2.TenantID != "" {
				_, level, _, _, _, enabled := h.alarms.ResolveEnablementByDevice(ctx, meta2.TenantID, reportDeviceID, alarm.SuspectedFall)
				if !enabled || level == "" {
					level = alarm.AlarmLevelWarn
				}
				nowMs := time.Now().UnixMilli()
				triggerData := map[string]interface{}{"source": "LeftBedFallActivity", "path": fallPath, "ts": nowMs}
				_ = h.alarms.PersistAlarmWithTriggerData(ctx, m.CardID, meta2.TenantID, reportDeviceID, alarm.SuspectedFall, level, time.UnixMilli(nowMs), triggerData)
			}
		}
	}

	if isBathroom {
		h.stayOnActivity(ctx, m, deviceID)
	}
}

// sleepStageSourceFromDeviceType 睡眠阶段置信度 100 分制：Sleepad=90，Radar=60，其它=0。
// 用 BaseDeviceType 解 AI 后缀（"Radar.AI01" → "Radar"），AI 派生事件继承源类型权重。
func sleepStageSourceFromDeviceType(deviceType string) int {
	switch strings.ToLower(redis.BaseDeviceType(deviceType)) {
	case "sleepad", "sleeppad":
		return 90
	case "radar":
		return 60
	default:
		return 0
	}
}

func intFromAny(v any) int {
	if v == nil {
		return 0
	}
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	default:
		return 0
	}
}

func stringFromAny(v any) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return x
	case int, int64:
		return fmt.Sprintf("%d", x)
	case float64:
		return fmt.Sprintf("%.0f", x)
	default:
		return fmt.Sprintf("%v", x)
	}
}

func streamFieldStr(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return strings.TrimSpace(s)
	case []byte:
		return strings.TrimSpace(string(s))
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", s))
	}
}

// sleepadTrackOverrideForInBed 与 sleepadTrackOverride 相同，但若缓冲里尚无 Sleepad 在床轨（count==0），返回 nil。
// 否则 InBed 会与 bedStatusFromTrackNumber(0) 组合被写成离床，且 DeriveBedStateFromRealtime 在 BedStatus==1 时不再推导，前端会一直离床。
func sleepadTrackOverrideForInBed(ctx context.Context, metaCache *service.DeviceMetaCache, buffer *service.MonitorBuffer, cardID string) *int {
	o := sleepadTrackOverride(ctx, metaCache, buffer, cardID)
	if o != nil && *o == 0 {
		return nil
	}
	return o
}

// sleepadTrackOverride 若该卡床绑含 Sleepad 则从 monitor buffer 快照统计在床轨数并返回，否则返回 nil。
func sleepadTrackOverride(ctx context.Context, metaCache *service.DeviceMetaCache, buffer *service.MonitorBuffer, cardID string) *int {
	if metaCache == nil || buffer == nil || cardID == "" {
		return nil
	}
	meta := metaCache.GetOrLoad(ctx, cardID)
	if meta == nil || meta.BedID == "" {
		return nil
	}
	bedDevs := meta.BedBoundDeviceIDs()
	hasSleepad := false
	for _, devID := range bedDevs {
		if dm := meta.Devices[devID]; dm != nil {
			t := strings.ToLower(dm.DeviceType)
			if strings.Contains(t, "sleepad") || strings.Contains(t, "sleeppad") {
				hasSleepad = true
				break
			}
		}
	}
	if !hasSleepad {
		return nil
	}
	snap := buffer.SnapshotCard(cardID)
	if snap == nil {
		return nil
	}
	count := service.SleepadTrackCountFromSnapshot(snap, meta, bedDevs)
	return &count
}

// isAlarmRelatedEvent 报警/安全相关事件保留 Info 级别日志
func isAlarmRelatedEvent(evName string) bool {
	switch evName {
	case "Fall", "SuspectedFall", "LeftBed", "InBed",
		"Offline", "OfflineRecover",
		"SensorDetached", "SensorDetachedRecover",
		"SignalPoor", "SignalPoorRecover",
		"AngleException", "AngleExceptionRecover",
		"HeartRateAlert", "RespRateAlert",
		"ApneaHypopnea", "WeakBiometricSignal",
		"BedSitUp", "SittingOnGround",
		"NightAbsence",
		"deviceStatus":
		return true
	}
	return false
}
