// alarm_router.go — iot:alarm:stream 消费 → alarm_events PG INSERT + card:status.AlarmState 投影。
//
// 唯一职责（CLAUDE.md 规则 #1.3 单源真相）：
//   - alarm 落 PG（owl-common/card.InsertAlarmAndUpdateCard 封装 SQL）
//   - card:status.AlarmState 单 writer 投影（owl-common/card.Writer.WriteCardStatus）
//   - device 类 alarm 信号转 DeviceSignal 接口（不直接改 device:status hash）
//
// 不做的事（已退役 / 归 sensor）：
//   - 床事件协调 / Stay FSM / NightAbsence 计时 / AHI 检测 / pending alarm
//   - bed state / room state 写 card:status（sensor:zone:state:stream → zone_state_projector）

package consumer

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"owl-common/alarm"
	"owl-common/card"
	owlredis "owl-common/redis"
	"owl-common/spatial"

	"wisefido-cardagg/internal/service"

	"go.uber.org/zap"
)

// EnablementResolver alarm enablement 查询；service.AlarmEnablementCache 实现。
type EnablementResolver interface {
	Resolve(ctx context.Context, tenantPref, deviceAddr, alarmType string) (level string, enabled bool)
}

// MetaResolver cardID + addr hint → 解析后的 device_addr；device_addr → cardID LPM 反查；
// HasBed/IsBathroom 反查卡静态属性。service.DeviceMetaCache 实现。
type MetaResolver interface {
	ResolveDeviceAddr(ctx context.Context, cardID, hint string) string
	LookupCardByDevice(ctx context.Context, deviceAddr string) string
	CardHasBed(ctx context.Context, cardID string) bool
	CardIsBathroom(ctx context.Context, cardID string) bool
}

// DeviceSignal device 类 alarm/recovery 通知；service.DeviceStatusTracker 实现。
// alarm_router 不直接改 device:status hash，由 tracker 单 writer 守门。
type DeviceSignal interface {
	OnDeviceAlarm(ctx context.Context, deviceAddr, alarmType string, payload map[string]interface{})
	OnDeviceRecover(ctx context.Context, deviceAddr, recoverType string)
	OnDeviceConnectivity(ctx context.Context, deviceAddr, deviceType string, online bool)
}

// sleepad 专有 alarm — radar 来源直接 drop。
var sleepadOnlyAlarms = map[string]struct{}{
	alarm.AbnormalBodyMovement: {},
	alarm.NoTurnOver:           {},
}

// recovery 类 alarm → 它要解除的 active alarm types。
var recoveryMap = map[string][]string{
	alarm.SingalPoorRecover:       {alarm.SignalPoor},
	alarm.AngleExceptionRecover:   {alarm.AngleException},
	alarm.SensorDetachedRecover:   {alarm.SensorDetached},
	alarm.AlarmTypeOfflineRecover: {alarm.AlarmTypeOffline, alarm.AlarmTypeDeviceFailure},
	alarm.AlarmTypeDeviceRecover:  {alarm.AlarmTypeDeviceFailure},
}

// device 类 onset — 转给 tracker 维护 DeviceState 标志位。
var deviceClassOnset = map[string]struct{}{
	alarm.AlarmTypeOffline: {},
	alarm.SignalPoor:       {},
	alarm.AngleException:   {},
	alarm.SensorDetached:   {},
}

const staleAlarmMs = 30_000 // gateway 时钟漂移 / 消息积压上限

type AlarmRouter struct {
	db        *sql.DB
	writer    *card.Writer
	reader    *card.Reader
	enable    EnablementResolver
	meta      MetaResolver
	devSignal DeviceSignal
	picker    *UnitPicker
	logger    *zap.Logger
}

func NewAlarmRouter(db *sql.DB, writer *card.Writer, reader *card.Reader, enable EnablementResolver, meta MetaResolver, devSignal DeviceSignal, picker *UnitPicker, logger *zap.Logger) *AlarmRouter {
	return &AlarmRouter{
		db:        db,
		writer:    writer,
		reader:    reader,
		enable:    enable,
		meta:      meta,
		devSignal: devSignal,
		picker:    picker,
		logger:    logger,
	}
}

func (r *AlarmRouter) Handle(ctx context.Context, msg *owlredis.IoTStreamMessage) error {
	eventName := msg.Category
	if eventName == "" {
		return nil
	}
	if time.Now().UnixMilli()-msg.Timestamp > staleAlarmMs {
		return nil
	}
	if !msg.DeviceAddr.IsValid() {
		return nil
	}

	ac := service.AddrCtxFromMsg(msg)
	resolvedAddr := r.meta.ResolveDeviceAddr(ctx, msg.SubjectEntity, ac.DeviceAddr)
	if resolvedAddr == "" {
		return nil
	}
	// cardID 解析一次：sensor 等 producer 留 SubjectEntity 空时走 cardagg LPM 兜底
	cardID := msg.SubjectEntity
	if cardID == "" {
		cardID = r.meta.LookupCardByDevice(ctx, resolvedAddr)
	}

	if strings.Contains(strings.ToLower(msg.DeviceType), "radar") {
		if _, ok := sleepadOnlyAlarms[eventName]; ok {
			return nil
		}
	}

	data := owlredis.FirstDataValue(msg.DataValue)
	if data == nil {
		data = make(map[string]interface{})
	}

	level, enabled := r.enable.Resolve(ctx, ac.TenantPref, resolvedAddr, eventName)

	// platform-agent trust：sensor 已在 AlarmBackChannel.gate 源头按 spatial_config LPM 查
	// alarm.cloud_config（per device /128 + alarmType）；未启用直接 drop，不会到达此处。
	// 因此 cardagg 端不重复 gate，trust producer 自带 level + 直接放行。
	// 详 owlBack/doc/platform_agent_addressing.md §6 + cardagg_sensor_split.md §2.3
	// + sensor consumer/alarm_back_channel.go (gate 源)。
	if spatial.IsPlatformAgentAddr(msg.Producer) {
		enabled = true
		if lvl, ok := data["alarm_level"].(string); ok && lvl != "" {
			level = lvl
		}
	}

	// recovery 分支：解除 active，不落新 row
	if recoverTypes, ok := recoveryMap[eventName]; ok {
		r.devSignal.OnDeviceRecover(ctx, resolvedAddr, eventName)
		if eventName == alarm.AlarmTypeOfflineRecover || eventName == alarm.AlarmTypeDeviceRecover {
			r.devSignal.OnDeviceConnectivity(ctx, resolvedAddr, msg.DeviceType, true)
		}
		return r.autoResolve(ctx, msg, cardID, recoverTypes)
	}

	// event_status=end：按 Registry EndPolicy 集中处理
	if status, _ := data["event_status"].(string); status == "end" {
		if def := alarm.LookupAlarm(eventName); def != nil {
			switch def.EndPolicy {
			case alarm.EndPolicyAutoResolve:
				return r.autoResolve(ctx, msg, cardID, []string{eventName})
			case alarm.EndPolicyManualAck:
				return nil
			}
		}
	}

	// device 类 onset：DeviceState 投影 + 无条件落 PG 审计（owl-common alarm Registry
	// 给设备类标了 SkipUnhandledCount=true，PG 层自动跳过 pop/notify 仅做审计落库；
	// HIPAA 要求设备离线/信号差等历史可追溯，不受租户 enablement 影响 —— CLAUDE.md 规则 #2.1）
	if _, ok := deviceClassOnset[eventName]; ok {
		r.devSignal.OnDeviceAlarm(ctx, resolvedAddr, eventName, data)
		if eventName == alarm.AlarmTypeOffline {
			r.devSignal.OnDeviceConnectivity(ctx, resolvedAddr, msg.DeviceType, false)
		}
		effectiveLevel := level
		if effectiveLevel == "" {
			if def := alarm.LookupAlarm(eventName); def != nil {
				effectiveLevel = def.DefaultLevel
			}
		}
		return r.persist(ctx, msg, cardID, eventName, effectiveLevel, data)
	}

	if !enabled || level == "" {
		return nil
	}
	return r.persist(ctx, msg, cardID, eventName, level, data)
}

func (r *AlarmRouter) persist(ctx context.Context, msg *owlredis.IoTStreamMessage, cardID, eventName, level string, data map[string]interface{}) error {
	ac := service.AddrCtxFromMsg(msg)
	triggerData, _ := json.Marshal(data)
	parentSpan := service.BuildParentSpan(msg.Producer, msg.SequenceNumber)

	result, cas, err := card.InsertAlarmAndUpdateCard(ctx, r.db, cardID, card.AlarmInsertParams{
		TenantID:    ac.TenantPref,
		DeviceAddr:  ac.DeviceAddr,
		EventType:   eventName,
		Category:    alarm.GetFHIRCategory(eventName),
		AlarmLevel:  level,
		TriggeredAt: time.UnixMilli(msg.Timestamp),
		TriggerData: triggerData,
		RoomID:      ac.RoomPref,
		Producer:    msg.Producer,
		ParentSpan:  parentSpan,
		TraceID:     parentSpan,
	})
	if err != nil {
		r.logger.Warn("alarm insert",
			zap.String("trace_id", parentSpan),
			zap.String("cid", cardID),
			zap.String("type", eventName),
			zap.Error(err))
		return err
	}
	if result.Deduped || result.SkippedNotify {
		return nil
	}
	r.logger.Info("alarm inserted",
		zap.String("trace_id", parentSpan),
		zap.String("cid", cardID),
		zap.String("event_id", result.EventID),
		zap.String("type", eventName),
		zap.String("level", level))
	return r.writeAlarmState(ctx, cardID, cas)
}

func (r *AlarmRouter) autoResolve(ctx context.Context, msg *owlredis.IoTStreamMessage, cardID string, alarmTypes []string) error {
	ac := service.AddrCtxFromMsg(msg)
	traceID := service.BuildParentSpan(msg.Producer, msg.SequenceNumber)
	cas, result, err := card.AutoResolveDeviceAlarms(ctx, r.db, cardID, ac.TenantPref, ac.DeviceAddr, alarmTypes)
	if err != nil {
		return err
	}
	if result.ResolvedCount == 0 {
		return nil
	}
	r.logger.Info("alarm auto-resolved",
		zap.String("trace_id", traceID),
		zap.String("cid", cardID),
		zap.Int("count", result.ResolvedCount))
	return r.writeAlarmState(ctx, cardID, cas)
}

// writeAlarmState 写 AlarmState 的同时合并 prev 状态重算 Display，单次 WriteCardStatus 落两块。
// Display 由 BuildCardDisplay 从 AlarmState.PopAlarm 派生 Section1DownRight 简称。
// 写完后触发 /80 父卡 picker 重算（子卡 alarm 翻转影响 unit 视图）。
func (r *AlarmRouter) writeAlarmState(ctx context.Context, cardID string, cas *card.CardAlarmState) error {
	if cas == nil {
		return nil
	}
	as := cas.ToAlarmState()
	merged := &card.CardStatus{CardID: cardID, AlarmState: as}
	if prev, err := r.reader.ReadCardStatus(ctx, cardID); err == nil && prev != nil {
		merged.RoomState = prev.RoomState
		merged.BedState = prev.BedState
		merged.Target = prev.Target
	}
	out := &card.CardStatus{
		CardID:     cardID,
		AlarmState: as,
		Display: BuildCardDisplay(merged,
			r.meta != nil && r.meta.CardHasBed(ctx, cardID),
			r.meta != nil && r.meta.CardIsBathroom(ctx, cardID)),
	}
	if err := r.writer.WriteCardStatus(ctx, out); err != nil {
		return err
	}
	if r.picker != nil {
		r.picker.RefreshParent(ctx, cardID)
		r.picker.RefreshSelf(ctx, cardID)
	}
	return nil
}
