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
	"strconv"
	"strings"
	"time"

	"owl-common/alarm"
	"owl-common/card"
	"owl-common/observation"
	owlredis "owl-common/redis"
	"owl-common/spatial"

	"wisefido-cardagg/internal/alarmpush"
	"wisefido-cardagg/internal/service"

	"go.uber.org/zap"
)

// EnablementResolver alarm enablement 查询；service.AlarmEnablementCache 实现。
type EnablementResolver interface {
	Resolve(ctx context.Context, tenantPref, deviceAddr, alarmType string) (level string, enabled bool)
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

const staleAlarmMs = 300_000 // gateway 时钟漂移 / 消息积压上限；alarm 稀疏，宽容到 5min 优先保不丢

type AlarmRouter struct {
	db         *sql.DB
	writer     *card.Writer
	reader     *card.Reader
	enable     EnablementResolver
	cache      *service.SpatialCache
	devSignal  DeviceSignal
	rebuilder  *DisplayRebuilder
	monitorBuf *service.MonitorBuffer // alarm 触发时取报警设备 monitor 缓存写 evidence
	logger     *zap.Logger
}

func NewAlarmRouter(db *sql.DB, writer *card.Writer, reader *card.Reader, enable EnablementResolver, cache *service.SpatialCache, devSignal DeviceSignal, rebuilder *DisplayRebuilder, monitorBuf *service.MonitorBuffer, logger *zap.Logger) *AlarmRouter {
	return &AlarmRouter{
		db:         db,
		writer:     writer,
		reader:     reader,
		enable:     enable,
		cache:      cache,
		devSignal:  devSignal,
		rebuilder:  rebuilder,
		monitorBuf: monitorBuf,
		logger:     logger,
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
	resolvedAddr := ac.DeviceAddr
	if resolvedAddr == "" {
		return nil
	}
	// cardID 解析：sensor producer 留 SubjectEntity 空时通过 entry 反查 device → card_id
	cardID := msg.SubjectEntity
	if cardID == "" {
		if addr := msg.DeviceAddr; addr.IsValid() {
			if owner := r.cache.LookupCardByDevice(addr); owner.IsValid() {
				cardID = owner.String()
			}
		}
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

	// platform-agent trust：sensor 已在 AlarmBackChannel.gate 源头按 device_config is_enabled gate，
	// 未启用直接 drop（或改发 event），不会到达此处。故 cardagg 不重复 gate enablement，trust 放行。
	// level 不取 producer：alarm_level 由 cardagg Resolve 从 device_config 单点决定（与 sleepace 直发同源）。
	// 详 owlBack/doc/platform_agent_addressing.md §6 + cardagg_sensor_split.md §2.3。
	if spatial.IsPlatformAgentAddr(msg.Producer) {
		enabled = true
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

	if !enabled {
		return nil
	}
	// level 唯一来自 device_config（Resolve；生成时已填 default 作保底，无 Registry 兜底）。
	// enabled 但 level 空 = 配置缺陷，留痕丢弃不静默（可观测性，非无声黑洞）。
	if level == "" {
		r.logger.Warn("alarm dropped: no resolvable level",
			zap.String("event", eventName),
			zap.String("device_addr", resolvedAddr),
			zap.String("card_id", cardID))
		return nil
	}
	return r.persist(ctx, msg, cardID, eventName, level, data)
}

// isFirmwareOriginReason 判 fall origin = device 直发：reason 形如 <device_uid_hex>.<alarmType>。
// sensor 自发 reason=dbn.<alarmType>（前缀非 hex）→ false；空 / 无 "." → false。
func isFirmwareOriginReason(reason string) bool {
	i := strings.IndexByte(reason, '.')
	if i <= 0 {
		return false
	}
	for _, c := range reason[:i] {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// trackIDKey 把 payload 里的 track_id（JSON 解出为 float64）转 MonitorBuffer 的字符串键（strconv.Itoa 同源）。
func trackIDKey(v any) string {
	switch n := v.(type) {
	case float64:
		return strconv.Itoa(int(n))
	case int:
		return strconv.Itoa(n)
	case string:
		return n
	}
	return ""
}

func (r *AlarmRouter) persist(ctx context.Context, msg *owlredis.IoTStreamMessage, cardID, eventName, level string, data map[string]interface{}) error {
	ac := service.AddrCtxFromMsg(msg)
	reasonStr, _ := data["reason"].(string) // origin：<uid>.alarmType 固件直发 / dbn.alarmType sensor 自发 → alarm_events.reason

	// 固件直发 fall（reason=<uid_hex>.<type>）payload 无坐标 → 按 track_id 从 MonitorBuffer 补 position 进
	// payload，与 DBN 腿（payload 自带）统一落点 payload.position_x/y/z，下游/feedback 单一读法不再 lookback。
	if r.monitorBuf != nil && isFirmwareOriginReason(reasonStr) {
		if _, has := data[observation.FieldPositionX]; !has {
			if x, y, z, ok := r.monitorBuf.TrackPosition(cardID, ac.DeviceAddr, trackIDKey(data[observation.FieldTrackID])); ok {
				data[observation.FieldPositionX] = x
				data[observation.FieldPositionY] = y
				data[observation.FieldPositionZ] = z
			}
		}
	}
	triggerData, _ := json.Marshal(data)
	parentSpan := service.BuildParentSpan(msg.Producer, msg.SequenceNumber)

	// alarm 触发时刻取报警设备 monitor 缓存（HR/RR/position/pose 等所有 raw 字段），
	// 写 alarm_events.evidence.monitor —— elder care 临床场景必需"事发现场"的瞬时值，
	// 尤其 sleepace HR/RR alarm 之前只发 type+status 不带瞬时值（sleepace 厂家 payload 也不含），
	// 由 cardagg 集中从 MonitorBuffer 取报警设备最近一帧统一处理，支持多种设备（sleepad/radar/...）。
	var metadata json.RawMessage
	evidence := map[string]any{}
	if r.monitorBuf != nil {
		if snap := r.monitorBuf.SnapshotByDevice(cardID, ac.DeviceAddr); snap != nil {
			evidence["monitor"] = snap
		}
	}
	// 合并 sensor DBN 腿在 payload.evidence 带的 fire/pin 判据快照（取证：band/tfloor/cell + 该点 pin 状态；
	// pin=null 即"摔倒点没被任何 sit pin 覆盖"=误报根因）。与 monitor 同写 alarm_events.evidence，不互相覆盖。
	if pe, ok := data["evidence"].(map[string]interface{}); ok {
		if f, ok2 := pe["fire"]; ok2 {
			evidence["fire"] = f
		}
		if pin, ok2 := pe["pin"]; ok2 {
			evidence["pin"] = pin
		}
	}
	if len(evidence) > 0 {
		metadata, _ = json.Marshal(evidence)
	}

	// 推断类 fall（silent/lost/still/bedside）传 incident_ts_ms = 真实发生时刻；envelope.Timestamp = 决策时刻。
	// triggered_at = incident（FE 据此 [t-1, t+2] 拉 replay 窗口）；alerted_at = nowMs（延迟标记审计用）。
	// 缺失（firmware Fall 等直发类）→ triggered_at = alerted_at = msg.Timestamp。
	alertedAt := time.UnixMilli(msg.Timestamp)
	triggeredAt := alertedAt
	if incidentMs := parseIncidentTsMs(data); incidentMs > 0 {
		triggeredAt = time.UnixMilli(incidentMs)
	}

	result, cas, err := card.InsertAlarmAndUpdateCard(ctx, r.db, cardID, card.AlarmInsertParams{
		TenantID:    ac.TenantPref,
		DeviceAddr:  ac.DeviceAddr,
		EventType:   eventName,
		Category:    alarm.GetFHIRCategory(eventName),
		AlarmLevel:  level,
		TriggeredAt: triggeredAt,
		AlertedAt:   alertedAt,
		TriggerData: triggerData,
		Metadata:    metadata,
		Reason:      reasonStr,
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
	if result.Deduped {
		// 可观测性（2026-06-05）：dedup 不再无声 return（此前"被 dedup 丢"在 cardagg 无任何日志，
		// 排查 Case2"报了没到 UI"时全靠人肉推。Debug 级避免 firmware 同一摔倒 1Hz 刷屏；
		// active_event_id = 撞上的那条 5min 内 active 同类告警。
		r.logger.Debug("alarm deduped",
			zap.String("cid", cardID),
			zap.String("type", eventName),
			zap.String("level", level),
			zap.String("active_event_id", result.EventID))
		return nil
	}
	if result.SkippedNotify {
		return nil // device-class：已落库审计、仅跳 pop/notify，非丢弃，不另记
	}
	r.logger.Info("alarm inserted",
		zap.String("trace_id", parentSpan),
		zap.String("cid", cardID),
		zap.String("event_id", result.EventID),
		zap.String("type", eventName),
		zap.String("level", level))
	// APNs 推送 trigger：alarm 已落 alarm_events 且非 dedup / 非 SkippedNotify。
	// notify_mode / 节假日 / login_only 等过滤在 wisefido-data 端 should_notify_user() 统一执行。
	alarmpush.NotifyWisefidoData(r.logger, ac.TenantPref, cardID, ac.DeviceAddr, result.EventID, eventName, level)
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

// parseIncidentTsMs 从 alarm dataValue map 提 incident_ts_ms 数字字段。
// sensor 推断类 fall 写入 fields["incident_ts_ms"]（engine.publishAIMessage）；
// JSON unmarshal 后数字类型可能是 float64 / json.Number / int64，分别 cover。
// 缺失 / 非法 → 返回 0（caller 退化 triggered_at = alerted_at = msg.Timestamp）。
func parseIncidentTsMs(data map[string]interface{}) int64 {
	v, ok := data["incident_ts_ms"]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int64:
		return n
	case int:
		return int64(n)
	case json.Number:
		ms, _ := n.Int64()
		return ms
	}
	return 0
}

// writeAlarmState 写 AlarmState 的同时合并 prev 状态重算 Display，单次 WriteCardStatus 落两块。
// Display 由 BuildCardDisplay 从 AlarmState.PopAlarm 派生 Section1DownRight 简称。
// writeAlarmState 写 alarm state，display + 父卡重算交给 picker.RebuildDisplay。
func (r *AlarmRouter) writeAlarmState(ctx context.Context, cardID string, cas *card.CardAlarmState) error {
	if cas == nil {
		return nil
	}
	out := &card.CardStatus{
		CardID:     cardID,
		AlarmState: cas.ToAlarmState(),
	}
	if err := r.writer.WriteCardStatus(ctx, out); err != nil {
		return err
	}
	if r.rebuilder != nil {
		r.rebuilder.Rebuild(ctx, cardID)
	}
	return nil
}
