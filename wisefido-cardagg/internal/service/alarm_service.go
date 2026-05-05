package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"owl-common/alarm"
	"owl-common/card"
	"owl-common/observation"
	"owl-common/redis"

	"go.uber.org/zap"
)

const redisPendingAlarmKey = "alarm:pending"

// RedisPendingStore 计时型告警待处理存储，由 main 传入 *redis.Client。
type RedisPendingStore interface {
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HGet(ctx context.Context, key, field string) (string, error)
	HSet(ctx context.Context, key string, values ...interface{}) error
	HDel(ctx context.Context, key string, fields ...string) error
}

type AlarmService struct {
	writer       *card.Writer
	reader       *card.Reader
	db           *sql.DB
	enablement   *AlarmEnablementCache
	redisPending RedisPendingStore
	logger       *zap.Logger
}

func NewAlarmService(writer *card.Writer, reader *card.Reader, db *sql.DB, enablement *AlarmEnablementCache, logger *zap.Logger) *AlarmService {
	return &AlarmService{writer: writer, reader: reader, db: db, enablement: enablement, logger: logger}
}

func (s *AlarmService) SetRedisPending(store RedisPendingStore) {
	s.redisPending = store
}

// AHIQueryReady 是否已具备 AHI 查询条件（DB + iot_timeseries）。未实现时返回 false，alarm 层不依赖 CheckAH。
func (s *AlarmService) AHIQueryReady() bool {
	return s.db != nil
}

// CheckAH 查 iot_timeseries 本 device 前 120 秒：RR 每 2 秒一发，无 RR 或 RR<8 为坏周期，连续 8 个坏周期记 1 次 apnea；
// 若连续 60 秒内≥2 次或 120 秒内≥4 次则返回 true（应发 ApneaHypopnea 报警），否则 false。表不存在或查询失败返回 false。
func (s *AlarmService) CheckAH(ctx context.Context, deviceID string) bool {
	if s.db == nil || deviceID == "" {
		return false
	}
	sinceMs := time.Now().Add(-120 * time.Second).UnixMilli()
	rows, err := s.db.QueryContext(ctx, `
		SELECT its."timestamp", its.data_value FROM iot_timeseries its
		WHERE its.device_id = $1::uuid AND its."timestamp" >= $2 AND its.topic_type = 'monitor'
		ORDER BY its."timestamp"`,
		deviceID, sinceMs)
	if err != nil {
		s.logger.Debug("CheckAH query", zap.String("device_id", deviceID), zap.Error(err))
		return false
	}
	defer rows.Close()

	type point struct {
		ts time.Time
		rr int
	}
	since := time.Now().Add(-120 * time.Second)
	var points []point
	for rows.Next() {
		var tsMs int64
		var dataValueJSON []byte
		if err := rows.Scan(&tsMs, &dataValueJSON); err != nil {
			continue
		}
		ts := time.UnixMilli(tsMs)
		rr := parseRRFromDataValue(dataValueJSON)
		points = append(points, point{ts: ts, rr: rr})
	}
	// 2 秒一发 RR；无 RR 或 RR<8 为坏周期，连续 8 个坏周期记 1 次 apnea
	const periodSec = 2
	const periodsPerApnea = 8
	numSlots := 120 / periodSec
	slotRR := make([]int, numSlots)
	for i := range slotRR {
		slotRR[i] = -1
	}
	for _, p := range points {
		elapsed := p.ts.Sub(since).Seconds()
		if elapsed < 0 || elapsed >= 120 {
			continue
		}
		idx := int(elapsed) / periodSec
		if idx >= numSlots {
			idx = numSlots - 1
		}
		if p.rr > slotRR[idx] {
			slotRR[idx] = p.rr
		}
	}
	var apneaEndTimes []time.Time
	i := 0
	for i < numSlots {
		bad := slotRR[i] < 0 || slotRR[i] < 8
		if !bad {
			i++
			continue
		}
		j := i
		for j < numSlots && (slotRR[j] < 0 || slotRR[j] < 8) {
			j++
		}
		if j-i >= periodsPerApnea {
			end := since.Add(time.Duration(j*periodSec) * time.Second)
			apneaEndTimes = append(apneaEndTimes, end)
		}
		i = j
	}
	// 某次 apnea 前 60 秒内≥2 次 或 前 120 秒内≥4 次 则报警
	for _, e := range apneaEndTimes {
		c60, c120 := 0, 0
		for _, t := range apneaEndTimes {
			sec := e.Sub(t).Seconds()
			if sec >= 0 && sec <= 60 {
				c60++
			}
			if sec >= 0 && sec <= 120 {
				c120++
			}
		}
		if c60 >= 2 || c120 >= 4 {
			return true
		}
	}
	return false
}

// parseRRFromDataValue 从 data_value JSONB 数组首项，按 observation.Track 解析 RespiratoryRate；失败返回 0。
func parseRRFromDataValue(dataValueJSON []byte) int {
	var arr []interface{}
	if err := json.Unmarshal(dataValueJSON, &arr); err != nil || len(arr) == 0 {
		return 0
	}
	obj, ok := arr[0].(map[string]interface{})
	if !ok {
		return 0
	}
	var t observation.Track
	t.FromFieldMap(obj)
	return t.RespiratoryRate
}

// ----- 工具 1：按设备查使能并解析 level、duration、min/max（alarm_device.monitor_config → AlarmEnablementCache；min/max 来自 Registry 或 ea.AlarmParams） -----
// ResolveEnablementByDevice 统一入口：Input tenantID, deviceID, alarmType(流中 category/event_name)。Output 标准化 alarmType, level, durationSec(0=即时型 >0=时间触发型), min, max, enabled。
func (s *AlarmService) ResolveEnablementByDevice(ctx context.Context, tenantID, deviceID, alarmType string) (outAlarmType, level string, durationSec int, min, max *int, enabled bool) {
	if alarmType == "" {
		return "", "", 0, nil, nil, false
	}
	def := alarm.LookupAlarm(alarmType)
	if def != nil {
		outAlarmType = def.Key
	} else {
		outAlarmType = alarmType
	}
	var ea *EnabledAlarm
	if s.enablement != nil {
		ea, enabled = s.enablement.IsEnabled(ctx, tenantID, deviceID, outAlarmType)
		if !enabled || ea == nil {
			s.logger.Debug("alarm not enabled", zap.String("did", deviceID), zap.String("event", outAlarmType))
			return outAlarmType, "", 0, nil, nil, false
		}
		if ea.AlarmLevel != "" {
			level = observation.NormalizeEventLevel(ea.AlarmLevel)
		}
	} else {
		enabled = false
		return outAlarmType, "", 0, nil, nil, false
	}
	if level == "" && def != nil {
		level = observation.NormalizeEventLevel(def.DefaultLevel)
	}
	if level != "" {
		level = observation.NormalizeEventLevel(level)
	}
	durationSec = getDurationSec(def, ea)
	min = getParamInt(ea, def, alarm.ParamMin)
	max = getParamInt(ea, def, alarm.ParamMax)
	enabled = level != "" // 有等级才视为使能可落库
	return outAlarmType, level, durationSec, min, max, enabled
}

// ----- 工具 2：根据 Registry 决定立即型还是持续型 -----
// AlarmDef 返回 Registry 中的定义，nil 表示未知类型。
func (s *AlarmService) AlarmDef(eventName string) *alarm.AlarmDef {
	return alarm.LookupAlarm(eventName)
}

// IsImmediateProcess 是否立即落库型（ProcessTypeImmediate）；持续型（如 TimeBased）返回 false。
func (s *AlarmService) IsImmediateProcess(eventName string) bool {
	def := alarm.LookupAlarm(eventName)
	return def != nil && def.ProcessType == alarm.ProcessTypeImmediate
}

// GetDeviceHRThreshold 返回该设备的心率报警阀值（高、低）。有配置则返回值，没有则返 nil。基于 ResolveEnablementByDevice 的 min/max。
func (s *AlarmService) GetDeviceHRThreshold(ctx context.Context, tenantID, deviceID string) (high, low *int) {
	_, _, _, _, max, ok := s.ResolveEnablementByDevice(ctx, tenantID, deviceID, alarm.HeartRateAlertHigh)
	if ok && max != nil && *max > 0 {
		high = max
	}
	_, _, _, min, _, ok := s.ResolveEnablementByDevice(ctx, tenantID, deviceID, alarm.HeartRateAlertLow)
	if ok && min != nil && *min >= 0 {
		low = min
	}
	return high, low
}

// ----- 工具 3：更新 DB 并发布 cardAlarmState -----
// PersistAlarmAndPublish 落库 alarm_events、更新 cards 报警计数与 popAlarm、写 Redis+stream。eventName/level 需已标准化。
func (s *AlarmService) PersistAlarmAndPublish(ctx context.Context, msg *redis.IoTStreamMessage, eventName, level string) error {
	if s.db == nil {
		s.logger.Warn("no DB, alarm not persisted", zap.String("cid", msg.SubjectEntity))
		return nil
	}
	if eventName == "" || level == "" {
		return nil
	}
	triggerData, _ := json.Marshal(redis.FirstDataValue(msg.DataValue))
	fhirCategory := alarm.GetFHIRCategory(eventName)
	triggeredAt := time.UnixMilli(msg.Timestamp)
	result, cardAlarmState, err := card.InsertAlarmAndUpdateCard(ctx, s.db, msg.SubjectEntity, card.AlarmInsertParams{
		TenantID:    msg.TenantID,
		DeviceID:    msg.DeviceID,
		EventType:   eventName,
		Category:    fhirCategory,
		AlarmLevel:  level,
		TriggeredAt: triggeredAt,
		TriggerData: triggerData,
	})
	if err != nil {
		s.logger.Warn("insert alarm failed", zap.String("cid", msg.SubjectEntity), zap.Error(err))
		return err
	}
	if result.Deduped {
		s.logger.Debug("alarm onset deduped (active exists)", zap.String("cid", msg.SubjectEntity), zap.String("device_id", msg.DeviceID), zap.String("type", eventName), zap.String("active_event_id", result.EventID))
		return nil
	}
	s.logger.Info("alarm inserted", zap.String("cid", msg.SubjectEntity), zap.String("event_id", result.EventID), zap.String("level", level), zap.String("type", eventName))
	if err := s.writeAlarmState(ctx, msg.SubjectEntity, cardAlarmState); err != nil {
		return err
	}
	s.notifyAlarmPushAsync(msg.TenantID, msg.SubjectEntity, msg.DeviceID, result.EventID, eventName, level)
	return nil
}

// PersistAlarmFromTrack 包头 + 当前 track 落库：EventType=eventName，TriggeredAt=track.Ts，TriggerData=track，Category=GetFHIRCategory(eventName)。alarm_status 由 InsertAlarmAndUpdateCard 写为 active。
func (s *AlarmService) PersistAlarmFromTrack(ctx context.Context, cardID, tenantID, deviceID, eventName, level string, track *observation.Track) error {
	if s.db == nil {
		s.logger.Warn("no DB, alarm not persisted", zap.String("cid", cardID))
		return nil
	}
	if eventName == "" || level == "" {
		return nil
	}
	triggerData, _ := json.Marshal(track)
	triggeredAt := time.UnixMilli(track.Ts)
	result, cardAlarmState, err := card.InsertAlarmAndUpdateCard(ctx, s.db, cardID, card.AlarmInsertParams{
		TenantID:    tenantID,
		DeviceID:    deviceID,
		EventType:   eventName,
		Category:    alarm.GetFHIRCategory(eventName),
		AlarmLevel:  level,
		TriggeredAt: triggeredAt,
		TriggerData: triggerData,
	})
	if err != nil {
		s.logger.Warn("insert alarm failed", zap.String("cid", cardID), zap.Error(err))
		return err
	}
	if result.Deduped {
		s.logger.Debug("alarm onset deduped (active exists)", zap.String("cid", cardID), zap.String("device_id", deviceID), zap.String("type", eventName), zap.String("active_event_id", result.EventID))
		return nil
	}
	s.logger.Info("alarm inserted", zap.String("cid", cardID), zap.String("event_id", result.EventID), zap.String("level", level), zap.String("type", eventName))
	if err := s.writeAlarmState(ctx, cardID, cardAlarmState); err != nil {
		return err
	}
	s.notifyAlarmPushAsync(tenantID, cardID, deviceID, result.EventID, eventName, level)
	return nil
}

// PersistAlarmWithTriggerData 落库告警，TriggeredAt=triggeredAt，TriggerData=triggerData（如 source 标明 LeftBedFallActivity 产生）。triggerData 为 nil 时写 {}。
func (s *AlarmService) PersistAlarmWithTriggerData(ctx context.Context, cardID, tenantID, deviceID, eventName, level string, triggeredAt time.Time, triggerData map[string]interface{}) error {
	if s.db == nil {
		s.logger.Warn("no DB, alarm not persisted", zap.String("cid", cardID))
		return nil
	}
	if eventName == "" || level == "" {
		return nil
	}
	var raw json.RawMessage
	if triggerData != nil {
		raw, _ = json.Marshal(triggerData)
	}
	if len(raw) == 0 {
		raw = json.RawMessage("{}")
	}
	result, cardAlarmState, err := card.InsertAlarmAndUpdateCard(ctx, s.db, cardID, card.AlarmInsertParams{
		TenantID:    tenantID,
		DeviceID:    deviceID,
		EventType:   eventName,
		Category:    alarm.GetFHIRCategory(eventName),
		AlarmLevel:  level,
		TriggeredAt: triggeredAt,
		TriggerData: raw,
	})
	if err != nil {
		s.logger.Warn("insert alarm failed", zap.String("cid", cardID), zap.Error(err))
		return err
	}
	if result.Deduped {
		s.logger.Debug("alarm onset deduped (active exists)", zap.String("cid", cardID), zap.String("device_id", deviceID), zap.String("type", eventName), zap.String("active_event_id", result.EventID))
		return nil
	}
	s.logger.Info("alarm inserted", zap.String("cid", cardID), zap.String("event_id", result.EventID), zap.String("level", level), zap.String("type", eventName))
	if err := s.writeAlarmState(ctx, cardID, cardAlarmState); err != nil {
		return err
	}
	s.notifyAlarmPushAsync(tenantID, cardID, deviceID, result.EventID, eventName, level)
	return nil
}

// RecordDeviceFailure 落库 alarm_events 为 DeviceFailure 并写 Redis，供前端告警列表展示。triggerData 可为 nil，内部会包一层 reason。
func (s *AlarmService) RecordDeviceFailure(ctx context.Context, cardID, tenantID, deviceID, reason string, triggerData []byte) error {
	if s.db == nil {
		s.logger.Warn("no DB, device_failure not persisted", zap.String("cid", cardID))
		return nil
	}
	if cardID == "" || deviceID == "" {
		return nil
	}
	eventName := alarm.AlarmTypeDeviceFailure
	level := alarm.DefaultDeviceFailure
	triggeredAt := time.Now()
	if len(triggerData) == 0 {
		triggerData, _ = json.Marshal(map[string]string{"reason": reason})
	}
	result, cardAlarmState, err := card.InsertAlarmAndUpdateCard(ctx, s.db, cardID, card.AlarmInsertParams{
		TenantID:    tenantID,
		DeviceID:    deviceID,
		EventType:   eventName,
		Category:    alarm.GetFHIRCategory(eventName),
		AlarmLevel:  level,
		TriggeredAt: triggeredAt,
		TriggerData: triggerData,
	})
	if err != nil {
		s.logger.Warn("insert device_failure alarm failed", zap.String("cid", cardID), zap.Error(err))
		return err
	}
	if result.Deduped {
		s.logger.Debug("device_failure deduped (active exists)", zap.String("cid", cardID), zap.String("device_id", deviceID), zap.String("active_event_id", result.EventID))
		return nil
	}
	s.logger.Info("device_failure alarm inserted", zap.String("cid", cardID), zap.String("event_id", result.EventID), zap.String("reason", reason))
	if err := s.writeAlarmState(ctx, cardID, cardAlarmState); err != nil {
		return err
	}
	s.notifyAlarmPushAsync(tenantID, cardID, deviceID, result.EventID, eventName, level)
	return nil
}

// pendingAlarmPayload 计时型待处理告警，存 Redis Hash 的 value（JSON）。
// 不落 card_id：卡重建后 card_id 会变，落库时按 device_id 解析当前绑定卡。
type pendingAlarmPayload struct {
	TenantID    string          `json:"tenant_id"`
	DeviceID    string          `json:"device_id"`
	AlarmType   string          `json:"alarm_type"`
	AlarmLevel  string          `json:"alarm_level"`
	EventSince  int64           `json:"event_since"` // ms
	DurationSec int             `json:"duration_sec"`
	UpgradeTo   string          `json:"upgrade_to,omitempty"`
	TriggerData json.RawMessage `json:"trigger_data,omitempty"`
}

// pendingField Redis Hash field：tenant_id:device_id:alarmType（设备维度，卡重建不影响 key）
func pendingField(tenantID, deviceID, alarmType string) string {
	return tenantID + ":" + deviceID + ":" + alarmType
}

// pendingFieldLegacy 旧版 field（card_id:device_id:alarmType），仅 RemovePendingAlarm 用于清理历史项
func pendingFieldLegacy(cardID, deviceID, alarmType string) string {
	return cardID + ":" + deviceID + ":" + alarmType
}

func getDurationSec(def *alarm.AlarmDef, ea *EnabledAlarm) int {
	if ea != nil && ea.AlarmParams != nil {
		if v, ok := ea.AlarmParams[alarm.ParamDurationSec]; ok {
			switch t := v.(type) {
			case int:
				return t
			case int64:
				return int(t)
			case float64:
				return int(t)
			}
		}
	}
	if def == nil {
		return 0
	}
	if def.DurationSec > 0 {
		return def.DurationSec
	}
	if def.AlarmParams == nil {
		return 0
	}
	v, ok := def.AlarmParams[alarm.ParamDurationSec]
	if !ok {
		return 0
	}
	switch t := v.(type) {
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	default:
		return 0
	}
}

// getParamInt 从 ea.AlarmParams 或 def.AlarmParams 取 key 对应整数值，ea 优先；无或非数字返回 nil。
func getParamInt(ea *EnabledAlarm, def *alarm.AlarmDef, key string) *int {
	var v interface{}
	var ok bool
	if ea != nil && ea.AlarmParams != nil {
		v, ok = ea.AlarmParams[key]
	}
	if !ok && def != nil && def.AlarmParams != nil {
		v, ok = def.AlarmParams[key]
	}
	if !ok || v == nil {
		return nil
	}
	var n int
	switch t := v.(type) {
	case int:
		n = t
	case int64:
		n = int(t)
	case float64:
		n = int(t)
	default:
		return nil
	}
	return &n
}

// isWithinResetTimeWindow 离床过久等需限定在 ResetTimeParams 内；InBedTime/OutBedTime 为单元本地时间，now 为 UTC，需用 timezone 转成单元本地再比较。
func isWithinResetTimeWindow(now time.Time, params *alarm.ResetTimeParams, timezone string) bool {
	if params == nil || params.InBedTime == "" || params.OutBedTime == "" {
		return true
	}
	inBed, errIn := time.Parse("15:04", params.InBedTime)
	outBed, errOut := time.Parse("15:04", params.OutBedTime)
	if errIn != nil || errOut != nil {
		return true
	}
	if timezone == "" {
		timezone = "UTC"
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	nowLocal := now.In(loc)
	t := time.Date(0, 1, 1, nowLocal.Hour(), nowLocal.Minute(), 0, 0, time.UTC)
	inBedT := time.Date(0, 1, 1, inBed.Hour(), inBed.Minute(), 0, 0, time.UTC)
	outBedT := time.Date(0, 1, 1, outBed.Hour(), outBed.Minute(), 0, 0, time.UTC)
	if inBedT.After(outBedT) {
		return t.After(inBedT) || t.Before(outBedT)
	}
	return t.After(inBedT) && t.Before(outBedT)
}

// getEffectiveTimezoneForCard 卡片/单元 IANA 时区，供 ResetTime 与本地 InBed/OutBed 对齐；查库失败或空则 America/Denver（与 unit 创建默认一致）。
func (s *AlarmService) getEffectiveTimezoneForCard(ctx context.Context, cardID string) string {
	const fallback = "America/Denver"
	if s.db == nil || cardID == "" {
		return fallback
	}
	var ctz, utz sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT NULLIF(TRIM(c.timezone), ''), NULLIF(TRIM(u.timezone), '')
		FROM cards c
		LEFT JOIN units u ON c.unit_id = u.unit_id AND c.tenant_id = u.tenant_id
		WHERE c.card_id = $1
	`, cardID).Scan(&ctz, &utz)
	if err != nil {
		return fallback
	}
	if ctz.Valid && ctz.String != "" {
		return ctz.String
	}
	if utz.Valid && utz.String != "" {
		return utz.String
	}
	return fallback
}

// InRestTimeWindow 当前时间是否在租户 ResetTime 休息窗口内（按 card 所在单元本地时钟），供 event_handler 等调用。
func (s *AlarmService) InRestTimeWindow(ctx context.Context, tenantID, cardID string) bool {
	params := s.getResetTimeParamsForTenant(ctx, tenantID)
	tz := s.getEffectiveTimezoneForCard(ctx, cardID)
	return isWithinResetTimeWindow(time.Now(), params, tz)
}

// getResetTimeParamsForTenant 从 alarm_cloud.metadata 读取租户 TenantResetTime.ResetTime，无配置或解析失败返回 nil（视为在窗口内）。
func (s *AlarmService) getResetTimeParamsForTenant(ctx context.Context, tenantID string) *alarm.ResetTimeParams {
	if s.db == nil || tenantID == "" {
		return nil
	}
	var metadata sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT metadata FROM alarm_cloud WHERE tenant_id = $1`, tenantID).Scan(&metadata)
	if err != nil || !metadata.Valid || metadata.String == "" {
		return nil
	}
	var tr alarm.TenantResetTime
	if err := json.Unmarshal([]byte(metadata.String), &tr); err != nil {
		return nil
	}
	if tr.ResetTime.InBedTime == "" || tr.ResetTime.OutBedTime == "" {
		return nil
	}
	return &tr.ResetTime
}

// TryAddLeftBedPendingAtTrigger 在触发地收到 LeftBed 时调用：若为计时型且满足 ResetTime、在床≥5min 则写入 alarm:pending。应在更新 bed_state 之前调用。返回是否已加入 pending（未加入时调用方直接 ResolveEnablementByDevice + PersistAlarmAndPublish）。
func (s *AlarmService) TryAddLeftBedPendingAtTrigger(ctx context.Context, msg *redis.IoTStreamMessage) bool {
	data := redis.FirstDataValue(msg.DataValue)
	if data == nil {
		return false
	}
	// Stage 1a：envelope.Category 是事件类型唯一权威
	eventName := msg.Category
	if eventName != alarm.LeftBed {
		return false
	}
	def := alarm.LookupAlarm(eventName)
	if def == nil {
		return false
	}
	_, level, durationSec, _, _, enabled := s.ResolveEnablementByDevice(ctx, msg.TenantID, msg.DeviceID, eventName)
	if !enabled || level == "" || durationSec < 10*60 || def.ProcessType != alarm.ProcessTypeTimeBased {
		return false
	}
	params := s.getResetTimeParamsForTenant(ctx, msg.TenantID)
	tz := s.getEffectiveTimezoneForCard(ctx, msg.SubjectEntity)
	if !isWithinResetTimeWindow(time.Now(), params, tz) {
		s.logger.Debug("LeftBed pending skip: outside ResetTimeParams", zap.String("cid", msg.SubjectEntity))
		return false
	}
	if s.reader != nil {
		status, _ := s.reader.ReadCardStatus(ctx, msg.SubjectEntity)
		if status == nil || status.BedState == nil || status.BedState.BedStatus != 0 {
			s.logger.Debug("LeftBed pending skip: not in_bed", zap.String("cid", msg.SubjectEntity))
			return false
		}
		if status.BedState.StartTime == 0 {
			s.logger.Debug("LeftBed pending skip: start_time not set", zap.String("cid", msg.SubjectEntity))
			return false
		}
		inBedDurationMs := time.Now().UnixMilli() - status.BedState.StartTime
		if inBedDurationMs < 5*60*1000 {
			s.logger.Debug("LeftBed pending skip: in_bed < 5min", zap.String("cid", msg.SubjectEntity), zap.Int64("ms", inBedDurationMs))
			return false
		}
	}
	if s.redisPending == nil {
		return false
	}
	triggerData, _ := json.Marshal(redis.FirstDataValue(msg.DataValue))
	upgradeTo := ""
	if def.UpgradeTo != "" {
		upgradeTo = def.UpgradeTo
	}
	if err := s.AddPendingAlarm(ctx, msg.TenantID, msg.DeviceID, eventName, level, msg.Timestamp, durationSec, upgradeTo, triggerData); err != nil {
		s.logger.Warn("add LeftBed pending at trigger", zap.String("cid", msg.SubjectEntity), zap.Error(err))
		return false
	}
	return true
}

// AddPendingAlarm 写入 Redis Hash alarm:pending，field = tenant_id:device_id:alarmType。
func (s *AlarmService) AddPendingAlarm(ctx context.Context, tenantID, deviceID, alarmType, alarmLevel string, eventSinceMs int64, durationSec int, upgradeTo string, triggerData json.RawMessage) error {
	if s.redisPending == nil {
		return nil
	}
	if tenantID == "" || deviceID == "" {
		return nil
	}
	p := pendingAlarmPayload{
		TenantID:    tenantID,
		DeviceID:    deviceID,
		AlarmType:   alarmType,
		AlarmLevel:  alarmLevel,
		EventSince:  eventSinceMs,
		DurationSec: durationSec,
		UpgradeTo:   upgradeTo,
		TriggerData: triggerData,
	}
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	err = s.redisPending.HSet(ctx, redisPendingAlarmKey, pendingField(tenantID, deviceID, alarmType), string(b))
	if err == nil {
		s.logger.Info("pending.added",
			zap.String("device_id", deviceID),
			zap.String("alarm_type", alarmType),
			zap.String("level", alarmLevel),
			zap.Int("duration_sec", durationSec),
			zap.Int64("event_since", eventSinceMs),
		)
	}
	return err
}

// RemovePendingAlarm 删除待处理项。会删新版 key；若 legacyCardID 非空同时删旧版 card_id:device_id:type，避免卡重建后删不掉。
func (s *AlarmService) RemovePendingAlarm(ctx context.Context, tenantID, legacyCardID, deviceID, alarmType string) error {
	if s.redisPending == nil {
		return nil
	}
	if deviceID == "" || alarmType == "" {
		return nil
	}
	var keys []string
	if tenantID != "" {
		keys = append(keys, pendingField(tenantID, deviceID, alarmType))
	}
	if legacyCardID != "" {
		keys = append(keys, pendingFieldLegacy(legacyCardID, deviceID, alarmType))
	}
	if len(keys) == 0 {
		return nil
	}
	err := s.redisPending.HDel(ctx, redisPendingAlarmKey, keys...)
	if err == nil {
		s.logger.Info("pending.removed",
			zap.String("device_id", deviceID),
			zap.String("alarm_type", alarmType),
			zap.String("tenant_id", tenantID),
			zap.String("reason", "InBed_cancel"),
		)
	}
	return err
}

// PurgeDisabledPendingForDevice 扫描该 device 在 alarm:pending 中的所有 entry，
// 凡是当前 enablement 已变成 disabled 的 alarm_type 一律删除。
// 用于 UI 把某 alarm 从 enabled 改成 disabled 时同步清理旧 pending（参考厂家"关闭报警清除统计"逻辑），
// 防止用户已关闭的告警因为旧 pending 还在到期触发。
func (s *AlarmService) PurgeDisabledPendingForDevice(ctx context.Context, tenantID, deviceID string) {
	if s.redisPending == nil || tenantID == "" || deviceID == "" {
		return
	}
	all, err := s.redisPending.HGetAll(ctx, redisPendingAlarmKey)
	if err != nil {
		s.logger.Warn("PurgeDisabledPending HGetAll", zap.Error(err))
		return
	}
	prefix := tenantID + ":" + deviceID + ":"
	var toDelete []string
	for field := range all {
		if !strings.HasPrefix(field, prefix) {
			continue
		}
		alarmType := strings.TrimPrefix(field, prefix)
		_, _, _, _, _, enabled := s.ResolveEnablementByDevice(ctx, tenantID, deviceID, alarmType)
		if !enabled {
			toDelete = append(toDelete, field)
		}
	}
	if len(toDelete) == 0 {
		return
	}
	if err := s.redisPending.HDel(ctx, redisPendingAlarmKey, toDelete...); err != nil {
		s.logger.Warn("PurgeDisabledPending HDel", zap.Strings("fields", toDelete), zap.Error(err))
		return
	}
	s.logger.Info("pending.purged_disabled",
		zap.String("device_id", deviceID),
		zap.String("tenant_id", tenantID),
		zap.Strings("fields", toDelete),
	)
}

// AddStayPendingIfEnabled 若该设备已开启 Stay 告警且配置了 duration_sec，则写入 alarm:pending，供 ScanPendingAlarms 到时落库。
func (s *AlarmService) AddStayPendingIfEnabled(ctx context.Context, tenantID, deviceID string, eventSinceMs int64) error {
	if s.redisPending == nil || deviceID == "" || tenantID == "" {
		return nil
	}
	ea, ok := s.enablement.IsEnabled(ctx, tenantID, deviceID, alarm.Stay)
	if !ok || ea == nil {
		return nil
	}
	def := alarm.LookupAlarm(alarm.Stay)
	durationSec := getDurationSec(def, ea)
	if durationSec <= 0 {
		return nil
	}
	level := ea.AlarmLevel
	if level == "" && def != nil {
		level = def.DefaultLevel
	}
	return s.AddPendingAlarm(ctx, tenantID, deviceID, alarm.Stay, level, eventSinceMs, durationSec, "", nil)
}

// ScanPendingAlarms 主程序每 5 分钟调用：若 now-EventSince>=DurationSec 则落库并发布 AlarmState，并删除该 pending。
func (s *AlarmService) ScanPendingAlarms(ctx context.Context) error {
	if s.redisPending == nil || s.db == nil {
		return nil
	}
	all, err := s.redisPending.HGetAll(ctx, redisPendingAlarmKey)
	if err != nil {
		return err
	}
	nowMs := time.Now().UnixMilli()
	for field, val := range all {
		var p pendingAlarmPayload
		if err := json.Unmarshal([]byte(val), &p); err != nil {
			continue
		}
		if p.TenantID == "" || p.DeviceID == "" {
			continue
		}
		durationMs := int64(p.DurationSec) * 1000
		if nowMs-p.EventSince < durationMs {
			continue
		}
		eventType := p.AlarmType
		if p.UpgradeTo != "" {
			eventType = p.UpgradeTo
		}
		triggeredAt := time.Unix(0, p.EventSince*int64(time.Millisecond))
		// 在 trigger_data 中补充 pending 信息，区分即时报警与计时到期报警
		triggerData := p.TriggerData
		if len(triggerData) > 0 {
			var td map[string]interface{}
			if json.Unmarshal(triggerData, &td) == nil {
				td["alarm_source"] = "pending_expired"
				td["pending_duration_sec"] = p.DurationSec
				td["pending_fired_at"] = nowMs
				triggerData, _ = json.Marshal(td)
			}
		}
		// cardID 传空：事务内按 device_id 解析当前绑定卡，避免卡重建后 pending 内 card_id 失效
		result, cardAlarmState, err := card.InsertAlarmAndUpdateCard(ctx, s.db, "", card.AlarmInsertParams{
			TenantID:    p.TenantID,
			DeviceID:    p.DeviceID,
			EventType:   eventType,
			Category:    alarm.GetFHIRCategory(eventType),
			AlarmLevel:  p.AlarmLevel,
			TriggeredAt: triggeredAt,
			TriggerData: triggerData,
		})
		if err != nil {
			s.logger.Warn("scan pending insert alarm failed",
				zap.String("tenant_id", p.TenantID),
				zap.String("device_id", p.DeviceID),
				zap.String("field", field),
				zap.Error(err))
			continue
		}
		if result.Deduped {
			// pending 触发命中已 active 行：清掉 pending 记录但不重发通知
			s.logger.Debug("pending alarm deduped (active exists)", zap.String("device_id", p.DeviceID), zap.String("type", eventType), zap.String("active_event_id", result.EventID))
			_ = s.redisPending.HDel(ctx, redisPendingAlarmKey, field)
			continue
		}
		resolvedCardID, lookupErr := card.LookupCardIDByDeviceID(ctx, s.db, p.DeviceID)
		if lookupErr != nil {
			s.logger.Warn("pending alarm fired: lookup card_id for stream",
				zap.String("device_id", p.DeviceID),
				zap.Error(lookupErr))
			resolvedCardID = ""
		}
		s.logger.Info("pending alarm fired",
			zap.String("cid", resolvedCardID),
			zap.String("device_id", p.DeviceID),
			zap.String("event_id", result.EventID),
			zap.String("type", eventType))
		if resolvedCardID != "" {
			if err := s.writeAlarmState(ctx, resolvedCardID, cardAlarmState); err != nil {
				s.logger.Warn("write alarm state after pending", zap.String("cid", resolvedCardID), zap.Error(err))
			} else {
				s.notifyAlarmPushAsync(p.TenantID, resolvedCardID, p.DeviceID, result.EventID, eventType, p.AlarmLevel)
			}
		}
		_ = s.redisPending.HDel(ctx, redisPendingAlarmKey, field)
	}
	return nil
}

// HandleAlarmProcess handles alarm acknowledgment from config:alarmProcess:stream.
func (s *AlarmService) HandleAlarmProcess(ctx context.Context, cardID, tenantID, alarmLevel, alarmType, eventID string) error {
	if eventID == "" {
		return nil
	}
	return s.refreshAlarmStateFromDB(ctx, cardID)
}

// SyncAllCardsAlarmState queries alarm state from DB and writes to Hash on startup.
func (s *AlarmService) SyncAllCardsAlarmState(ctx context.Context) error {
	if s.db == nil {
		return nil
	}

	cardRows, err := s.db.QueryContext(ctx, `SELECT card_id FROM cards`)
	if err != nil {
		return fmt.Errorf("query cards: %w", err)
	}
	defer cardRows.Close()

	var cardIDs []string
	for cardRows.Next() {
		var cid string
		if err := cardRows.Scan(&cid); err == nil {
			cardIDs = append(cardIDs, cid)
		}
	}

	synced := 0
	for _, cid := range cardIDs {
		cas, err := card.QueryCardAlarmState(ctx, s.db, cid)
		if err != nil {
			s.logger.Warn("query alarm state", zap.String("cid", cid), zap.Error(err))
			continue
		}
		if err := PublishCardStatusSilent(ctx, s.writer, cid, PublishFields{
			AlarmState: cas.ToAlarmState(),
		}); err != nil {
			s.logger.Warn("sync alarm state", zap.String("cid", cid), zap.Error(err))
			continue
		}
		synced++
	}

	s.logger.Info("startup alarm sync done", zap.Int("cards", synced))
	return nil
}

// SyncDeviceStatusFromActiveAlarms 启动时从 alarm_events.active 重建 device:status hash 的 alarm flags。
//
// 解决场景：
//   - WriteDeviceStatus 历史 bug 把 alarm flags 重置为 0，bug 修复后 qinglan/sleepace transition dedup 已
//     cache=1，不再 republish onset，hash 永远停留在 0。
//   - cardagg 重启后丢失内存状态，但 alarm_events.active 是 DB 真值，借此恢复 hash。
//
// 这是幂等的：每次启动都按 active 行重建 hash flag。已 auto_resolved/acked 的不计入（hash 应为 0）。
// 设备类只覆盖 3 个 hash flag：signal_poor / angle_abnormal / sensor_detached。
// offline 字段由 device-gateway 心跳/健康检测维护，不在此处覆写。
func (s *AlarmService) SyncDeviceStatusFromActiveAlarms(ctx context.Context) error {
	if s.db == nil || s.writer == nil {
		return nil
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT DISTINCT device_id::text, event_type
		FROM alarm_events
		WHERE alarm_status='active'
		  AND event_type IN ('SignalPoor','AngleException','SensorDetached')
		  AND (metadata->>'deleted_at' IS NULL)
	`)
	if err != nil {
		return fmt.Errorf("query active device alarms: %w", err)
	}
	defer rows.Close()

	patched := 0
	for rows.Next() {
		var deviceID, eventType string
		if err := rows.Scan(&deviceID, &eventType); err != nil {
			continue
		}
		var fieldKey string
		switch eventType {
		case alarm.SignalPoor:
			fieldKey = "signal_poor"
		case alarm.AngleException:
			fieldKey = "angle_abnormal"
		case alarm.SensorDetached:
			fieldKey = "sensor_detached"
		default:
			continue
		}
		if err := s.writer.PatchDeviceStatus(ctx, deviceID, map[string]interface{}{
			fieldKey: "1",
		}); err != nil {
			s.logger.Warn("sync device flag failed",
				zap.String("device_id", deviceID),
				zap.String("field", fieldKey),
				zap.Error(err))
			continue
		}
		patched++
	}
	s.logger.Info("startup device:status flag sync done", zap.Int("patched", patched))
	return nil
}

// HandleCardChange processes card add/update/delete from config:card:stream.
func (s *AlarmService) HandleCardChange(ctx context.Context, cardID, op string) error {
	if cardID == "" {
		return nil
	}

	if op == "deleted" || op == "delete" {
		s.logger.Info("card deleted, removing state", zap.String("cid", cardID))
		return s.writer.DeleteCardState(ctx, cardID)
	}

	s.logger.Info("card changed, refreshing alarms", zap.String("cid", cardID), zap.String("op", op))
	return s.refreshAlarmStateFromDB(ctx, cardID)
}

// HandleRecoveryWithTypes 将 device 下指定类型的 active 告警改为 auto_resolved 并写库、发布 AlarmState。
// 仅用于设备类恢复事件（OfflineRecover/SensorDetachedRecover/SignalPoorRecover/AngleExceptionRecover）；生理类不调用，须人工恢复。
// 供 alarm_handler/event_handler 在恢复型 case 直接调用。
func (s *AlarmService) HandleRecoveryWithTypes(ctx context.Context, msg *redis.IoTStreamMessage, alarmTypes []string) error {
	if s.db == nil || len(alarmTypes) == 0 {
		return nil
	}

	cardAlarmState, result, err := card.AutoResolveDeviceAlarms(ctx, s.db, msg.SubjectEntity, msg.TenantID, msg.DeviceID, alarmTypes)
	if err != nil {
		s.logger.Warn("auto resolve failed", zap.String("cid", msg.SubjectEntity), zap.Error(err))
		return err
	}
	if result.ResolvedCount == 0 {
		return nil
	}

	s.logger.Info("alarm auto-resolved",
		zap.String("cid", msg.SubjectEntity),
		zap.Int("resolved", result.ResolvedCount))

	return s.writeAlarmState(ctx, msg.SubjectEntity, cardAlarmState)
}

// writeAlarmState converts CardAlarmState → AlarmState and writes to Redis Hash + stream.
func (s *AlarmService) writeAlarmState(ctx context.Context, cardID string, cas *card.CardAlarmState) error {
	if cas == nil {
		return nil
	}
	return PublishCardStatus(ctx, s.writer, cardID, PublishFields{
		AlarmState: cas.ToAlarmState(),
	})
}

// notifyAlarmPushAsync 在 cardStatus 发布成功后，异步通知 wisefido-data 走 APNs（仅 staff；与 wisefido-sensor/alarmpush 同 URL）
func (s *AlarmService) notifyAlarmPushAsync(tenantID, cardID, deviceID, eventID, eventType, alarmLevel string) {
	base := strings.TrimSpace(os.Getenv("WISEFIDO_DATA_ALARM_PUSH_URL"))
	sec := strings.TrimSpace(os.Getenv("INTERNAL_ALARM_PUSH_SECRET"))
	if base == "" || sec == "" || tenantID == "" || cardID == "" {
		return
	}
	base = strings.TrimRight(base, "/")
	payload := map[string]string{
		"tenant_id":   tenantID,
		"card_id":     cardID,
		"device_id":   deviceID,
		"event_id":    eventID,
		"event_type":  eventType,
		"alarm_level": alarmLevel,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}
	go func() {
		c, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(c, http.MethodPost, base+"/internal/v1/push/alarm", bytes.NewReader(body))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Internal-Secret", sec)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			s.logger.Warn("alarm push to wisefido-data failed", zap.Error(err))
			return
		}
		_ = resp.Body.Close()
		if resp.StatusCode >= 300 {
			s.logger.Warn("alarm push to wisefido-data bad status", zap.Int("status", resp.StatusCode))
		}
	}()
}

// refreshAlarmStateFromDB re-queries alarm state from DB and writes to Hash.
// Used by HandleAlarmProcess / HandleCardChange where we don't have CardAlarmState in hand.
func (s *AlarmService) refreshAlarmStateFromDB(ctx context.Context, cardID string) error {
	if s.db == nil {
		return nil
	}
	cas, err := card.QueryCardAlarmState(ctx, s.db, cardID)
	if err != nil {
		return fmt.Errorf("query alarm state for card %s: %w", cardID, err)
	}
	return s.writeAlarmState(ctx, cardID, cas)
}

// ────────────────────────────────────────────────────────────────────
// NightAbsence：每天 OutBedTime（默认 7:30）检查，整夜无 InBed/LeftBed → alarm
// Sleepace 和 Radar 均适用：只要该 card 绑的设备开启了 NightAbsence 就检查。
// ────────────────────────────────────────────────────────────────────

const nightAbsenceCheckedKey = "alarm:night_absence_checked" // Redis Hash: cardID → "YYYY-MM-DD"

// CheckNightAbsence 每 10 分钟调用。对所有 ActiveBedCard，判断当地时间是否刚过 OutBedTime，
// 若今天尚未检查，则查 alarm_events 表昨晚 InBedTime~OutBedTime 有无 InBed/LeftBed 事件。
// 无事件 + NightAbsence enabled → PersistAlarmWithTriggerData。
func (s *AlarmService) CheckNightAbsence(ctx context.Context, metaCache *DeviceMetaCache) {
	if s.db == nil || s.redisPending == nil {
		return
	}

	rows, err := s.db.QueryContext(ctx, `SELECT card_id, tenant_id FROM cards WHERE card_type = 'ActiveBedCard'`)
	if err != nil {
		s.logger.Warn("night_absence: query cards failed", zap.Error(err))
		return
	}
	defer rows.Close()

	type cardRow struct {
		cardID, tenantID string
	}
	var cards []cardRow
	for rows.Next() {
		var c cardRow
		if err := rows.Scan(&c.cardID, &c.tenantID); err != nil {
			continue
		}
		cards = append(cards, c)
	}

	for _, c := range cards {
		s.checkNightAbsenceForCard(ctx, c.cardID, c.tenantID, metaCache)
	}
}

// getEffectiveTimezoneForCardLA 与 getEffectiveTimezoneForCard 相同，但 fallback 为 America/Los_Angeles。
func (s *AlarmService) getEffectiveTimezoneForCardLA(ctx context.Context, cardID string) string {
	const fallback = "America/Los_Angeles"
	if s.db == nil || cardID == "" {
		return fallback
	}
	var ctz, utz sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT NULLIF(TRIM(c.timezone), ''), NULLIF(TRIM(u.timezone), '')
		FROM cards c
		LEFT JOIN units u ON c.unit_id = u.unit_id AND c.tenant_id = u.tenant_id
		WHERE c.card_id = $1
	`, cardID).Scan(&ctz, &utz)
	if err != nil {
		return fallback
	}
	if ctz.Valid && ctz.String != "" {
		return ctz.String
	}
	if utz.Valid && utz.String != "" {
		return utz.String
	}
	return fallback
}

func (s *AlarmService) checkNightAbsenceForCard(ctx context.Context, cardID, tenantID string, metaCache *DeviceMetaCache) {
	tz := s.getEffectiveTimezoneForCardLA(ctx, cardID)
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc, _ = time.LoadLocation("America/Los_Angeles")
	}
	now := time.Now().In(loc)

	// 获取 ResetTime 参数（InBedTime/OutBedTime），默认 21:30/07:30
	params := s.getResetTimeParamsForTenant(ctx, tenantID)
	outBedHour, outBedMin := 7, 30
	if params != nil && params.OutBedTime != "" {
		fmt.Sscanf(params.OutBedTime, "%d:%d", &outBedHour, &outBedMin)
	}
	inBedHour, inBedMin := 21, 30
	if params != nil && params.InBedTime != "" {
		fmt.Sscanf(params.InBedTime, "%d:%d", &inBedHour, &inBedMin)
	}

	// 只在 OutBedTime 后 30 分钟内检查（避免重复 + 给数据缓冲）
	checkTime := time.Date(now.Year(), now.Month(), now.Day(), outBedHour, outBedMin, 0, 0, loc)
	if now.Before(checkTime) || now.After(checkTime.Add(30*time.Minute)) {
		return
	}

	// 今天已检查？
	todayStr := now.Format("2006-01-02")
	if checked, _ := s.redisPending.HGet(ctx, nightAbsenceCheckedKey, cardID); checked == todayStr {
		return
	}

	// 标记今天已检查
	_ = s.redisPending.HSet(ctx, nightAbsenceCheckedKey, cardID, todayStr)

	// 从 metaCache 获取设备列表
	meta := metaCache.GetOrLoad(ctx, cardID)
	if meta == nil || len(meta.Devices) == 0 {
		return
	}

	nightStart := time.Date(now.Year(), now.Month(), now.Day()-1, inBedHour, inBedMin, 0, 0, loc)
	nightEnd := time.Date(now.Year(), now.Month(), now.Day(), outBedHour, outBedMin, 0, 0, loc)

	// 两段式：
	// Pass 1：先把 card 上**所有设备**状态都评估一遍（presence / absent / offline）。
	//         Sleepad 的 presence 用于 Pass 2 里 Radar 的"合查"判定（同 card Sleepad 在床 → 抵消 Radar absent）。
	// Pass 2：按设备类型各自发对应事件：
	//         - Sleepad absent → BedNightAbsence（床层面）
	//         - Radar   absent AND 没有同 card 的 Sleepad 报 presence → NightAbsence（房间层面）
	//         - offline 不贡献，不发任何事件（Q1）
	type assessed struct {
		deviceType string
		state      devicePresenceState
	}
	results := make(map[string]assessed, len(meta.Devices))
	anySleepadPresence := false
	for devID, dm := range meta.Devices {
		st := s.assessDevicePresence(ctx, tenantID, devID, dm.DeviceType, nightStart, nightEnd)
		results[devID] = assessed{deviceType: dm.DeviceType, state: st}
		if st == presenceYes && strings.Contains(strings.ToLower(dm.DeviceType), "sleep") {
			anySleepadPresence = true
		}
	}

	for devID, r := range results {
		if r.state != presenceAbsent {
			continue // offline / presence 都不发
		}
		devType := strings.ToLower(r.deviceType)

		// 选 event_type + 看 enablement
		var eventType string
		switch {
		case strings.Contains(devType, "sleep"):
			eventType = alarm.BedNightAbsence
		case strings.Contains(devType, "radar"):
			// Radar 合查同 card Sleepad：若 Sleepad 报 presence 说明房间有人，Radar 的 absent 是盲区
			if anySleepadPresence {
				s.logger.Debug("night_absence: radar absent but sibling sleepad presence, skip",
					zap.String("cid", cardID), zap.String("device_id", devID))
				continue
			}
			eventType = alarm.NightAbsence
		default:
			continue
		}

		_, level, _, _, _, enabled := s.ResolveEnablementByDevice(ctx, tenantID, devID, eventType)
		if !enabled || level == "" {
			continue
		}

		s.logger.Info("night_absence: triggering",
			zap.String("cid", cardID),
			zap.String("device_id", devID),
			zap.String("device_type", r.deviceType),
			zap.String("event", eventType),
			zap.String("night", nightStart.Format("15:04")+"~"+nightEnd.Format("15:04")),
			zap.String("tz", tz),
		)
		triggerData := map[string]interface{}{
			"alarm_source": "night_absence_check",
			"night_start":  nightStart.Unix(),
			"night_end":    nightEnd.Unix(),
			"timezone":     tz,
		}
		_ = s.PersistAlarmWithTriggerData(ctx, cardID, tenantID, devID, eventType, level, nightEnd, triggerData)
	}
}

// devicePresenceState 单设备整夜 3 态判定。
type devicePresenceState int

const (
	// presenceOffline 整夜无任何 track（设备掉线/断电）：状态未知，不贡献 absence/presence
	presenceOffline devicePresenceState = iota
	// presenceAbsent 有 track 但没任何"在床/在房间"信号 → 贡献 absence
	presenceAbsent
	// presenceYes 至少一条在床/在房间信号 → 贡献 presence
	presenceYes
)

// assessDevicePresence 判断单个设备在夜间窗口的 3 态。
//
// Sleepad：
//  1. iot_timeseries 有任何 track → 有数据；无 → offline
//  2. 床上信号（bed_status=0）或 alarm_events.InBed 或 sleepace_report.sleep_state>=2 → presence
//  3. 否则 absent
//
// Radar：
//  1. iot_timeseries 有任何 track → 有数据；无 → offline
//  2. number_people>0 → presence
//  3. 否则 absent
//
// 同 room 有 Sleepad + Radar 两台设备时在 card iteration 自然覆盖："Radar 房间 absent + Sleepad 床 presence" → card presence。
func (s *AlarmService) assessDevicePresence(ctx context.Context, tenantID, deviceID, deviceType string, nightStart, nightEnd time.Time) devicePresenceState {
	if s.db == nil {
		return presenceOffline
	}
	nightStartMs := nightStart.UnixMilli()
	nightEndMs := nightEnd.UnixMilli()
	devType := strings.ToLower(deviceType)

	// 步骤 0：device.status='offline'（由 cardagg/qinglan device_healthCheck 维护，90s 无心跳 → offline）
	// 直接短路，避免依赖 iot_timeseries 数据查询。两层防护：
	//   ① 这里 device.status 短路（依赖 healthCheck 准确）
	//   ② 步骤 1 monitor 数据过滤（healthCheck 漏判时兜底）
	var devStatus sql.NullString
	_ = s.db.QueryRowContext(ctx, `SELECT status FROM devices WHERE device_id = $1`, deviceID).Scan(&devStatus)
	if devStatus.Valid && devStatus.String == "offline" {
		return presenceOffline
	}

	// 步骤 1：设备整夜是否有真实监测数据（必须 topic_type='monitor'）。
	// 注意：不能算 alarm/event 类（如 connectionStatus 产生的 Offline/OfflineRecover、
	// deviceSenSor 产生的 SensorDetached），那些是设备元事件，offline 设备照样会产生，
	// 会把 offline 误判成 absent 触发 NightAbsence/BedNightAbsence 误报。
	var trackCount int
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM iot_timeseries
		WHERE tenant_id = $1 AND device_id = $2
		  AND topic_type = 'monitor'
		  AND timestamp >= $3 AND timestamp < $4
	`, tenantID, deviceID, nightStartMs, nightEndMs).Scan(&trackCount); err != nil || trackCount == 0 {
		return presenceOffline
	}

	// 步骤 2：按类型查"有人"
	if strings.Contains(devType, "radar") {
		var cnt int
		_ = s.db.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM iot_timeseries
			WHERE tenant_id = $1 AND device_id = $2
			  AND category = 'number_people'
			  AND timestamp >= $3 AND timestamp < $4
			  AND EXISTS (
			    SELECT 1 FROM jsonb_array_elements(data_value) elem
			    WHERE (elem->>'number_people')::int > 0
			  )
		`, tenantID, deviceID, nightStartMs, nightEndMs).Scan(&cnt)
		if cnt > 0 {
			return presenceYes
		}
		return presenceAbsent
	}

	if strings.Contains(devType, "sleep") {
		// 2a. iot_timeseries 有 bed_status=0 (在床) / HR>0 / RR>0 (生命体征上报意味着人在)
		var bedCnt int
		_ = s.db.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM iot_timeseries
			WHERE tenant_id = $1 AND device_id = $2
			  AND topic_type = 'monitor'
			  AND timestamp >= $3 AND timestamp < $4
			  AND EXISTS (
			    SELECT 1 FROM jsonb_array_elements(data_value) elem
			    WHERE (elem->>'bed_status')::int = 0
			       OR (elem->>'heart_rate')::int > 0
			       OR (elem->>'respiratory_rate')::int > 0
			  )
		`, tenantID, deviceID, nightStartMs, nightEndMs).Scan(&bedCnt)
		if bedCnt > 0 {
			return presenceYes
		}

		// 2b. alarm_events 有 InBed 事件（设备上报的在床事件）
		var inBedCnt int
		_ = s.db.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM alarm_events
			WHERE tenant_id = $1 AND device_id = $2
			  AND event_type = $3
			  AND triggered_at >= to_timestamp($4::double precision / 1000)
			  AND triggered_at <  to_timestamp($5::double precision / 1000)
		`, tenantID, deviceID, alarm.InBed, nightStartMs, nightEndMs).Scan(&inBedCnt)
		if inBedCnt > 0 {
			return presenceYes
		}

		// 2c. sleepace_report 兜底：找窗口覆盖 nightStart 的报告，检查 sleep_state>=2
		nightStartSec := nightStart.Unix()
		nightEndSec := nightEnd.Unix()
		var sleepState sql.NullString
		var startTime int64
		var timeStepSec, recordCount int
		if err := s.db.QueryRowContext(ctx, `
			SELECT sr.sleep_state, sr.start_time,
			       COALESCE(NULLIF(sr.time_step, 0), 60) AS ts,
			       COALESCE(NULLIF(sr.record_count, 0), 1440) AS rc
			FROM sleepace_report sr
			JOIN device_store ds ON sr.device_code = ds.device_code
			WHERE ds.device_id = $1
			  AND sr.start_time <= $2
			  AND sr.start_time + COALESCE(NULLIF(sr.time_step, 0), 60) * COALESCE(NULLIF(sr.record_count, 0), 1440) > $2
			ORDER BY sr.start_time DESC
			LIMIT 1
		`, deviceID, nightStartSec).Scan(&sleepState, &startTime, &timeStepSec, &recordCount); err == nil && sleepState.Valid {
			raw := strings.Trim(sleepState.String, "\"[]")
			vals := strings.Split(raw, ",")
			if timeStepSec <= 0 {
				timeStepSec = 60
			}
			idxFrom := int((nightStartSec - startTime) / int64(timeStepSec))
			idxTo := int((nightEndSec - startTime) / int64(timeStepSec))
			if idxFrom < 0 {
				idxFrom = 0
			}
			if idxTo > len(vals) {
				idxTo = len(vals)
			}
			// sleep_state 厂家文档：0=未监测；1=离床；2=清醒；3=深睡；4=浅睡；6=坐起。>=2 才是人在床。
			for i := idxFrom; i < idxTo && i < len(vals); i++ {
				n := 0
				fmt.Sscanf(strings.TrimSpace(vals[i]), "%d", &n)
				if n >= 2 {
					return presenceYes
				}
			}
		}

		return presenceAbsent
	}

	// 未识别类型：保守按 offline 处理，不贡献判定
	return presenceOffline
}
