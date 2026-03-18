package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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
	since := time.Now().Add(-120 * time.Second)
	rows, err := s.db.QueryContext(ctx, `
		SELECT timestamp, data_values FROM iot_timeseries
		WHERE device_id = $1 AND timestamp >= $2 AND topic_type = 'monitor'
		ORDER BY timestamp`,
		deviceID, since)
	if err != nil {
		s.logger.Debug("CheckAH query", zap.String("device_id", deviceID), zap.Error(err))
		return false
	}
	defer rows.Close()

	type point struct {
		ts time.Time
		rr int
	}
	var points []point
	for rows.Next() {
		var ts time.Time
		var dataValuesJSON []byte
		if err := rows.Scan(&ts, &dataValuesJSON); err != nil {
			continue
		}
		rr := parseRRFromDataValues(dataValuesJSON)
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

// parseRRFromDataValues 从 data_values JSONB 取 data_value/dataValue 首项，按 observation.Track 解析，返回 RespiratoryRate；缺失或解析失败返回 0（视为坏周期）。
func parseRRFromDataValues(dataValuesJSON []byte) int {
	var m map[string]interface{}
	if err := json.Unmarshal(dataValuesJSON, &m); err != nil {
		return 0
	}
	dv, _ := m["data_value"].(interface{})
	if dv == nil {
		dv, _ = m["dataValue"].(interface{})
	}
	if arr, ok := dv.([]interface{}); ok && len(arr) > 0 {
		dv = arr[0]
	}
	obj, ok := dv.(map[string]interface{})
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
		s.logger.Warn("no DB, alarm not persisted", zap.String("cid", msg.CardID))
		return nil
	}
	if eventName == "" || level == "" {
		return nil
	}
	triggerData, _ := json.Marshal(redis.FirstDataValue(msg.DataValue))
	fhirCategory := alarm.GetFHIRCategory(eventName)
	triggeredAt := time.UnixMilli(msg.Timestamp)
	result, cardAlarmState, err := card.InsertAlarmAndUpdateCard(ctx, s.db, msg.CardID, card.AlarmInsertParams{
		TenantID:    msg.TenantID,
		DeviceID:    msg.DeviceID,
		EventType:   eventName,
		Category:    fhirCategory,
		AlarmLevel:  level,
		TriggeredAt: triggeredAt,
		TriggerData: triggerData,
	})
	if err != nil {
		s.logger.Warn("insert alarm failed", zap.String("cid", msg.CardID), zap.Error(err))
		return err
	}
	s.logger.Info("alarm inserted", zap.String("cid", msg.CardID), zap.String("event_id", result.EventID), zap.String("level", level), zap.String("type", eventName))
	return s.writeAlarmState(ctx, msg.CardID, cardAlarmState)
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
	s.logger.Info("alarm inserted", zap.String("cid", cardID), zap.String("event_id", result.EventID), zap.String("level", level), zap.String("type", eventName))
	return s.writeAlarmState(ctx, cardID, cardAlarmState)
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
	s.logger.Info("alarm inserted", zap.String("cid", cardID), zap.String("event_id", result.EventID), zap.String("level", level), zap.String("type", eventName))
	return s.writeAlarmState(ctx, cardID, cardAlarmState)
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
	s.logger.Info("device_failure alarm inserted", zap.String("cid", cardID), zap.String("event_id", result.EventID), zap.String("reason", reason))
	return s.writeAlarmState(ctx, cardID, cardAlarmState)
}

// pendingAlarmPayload 计时型待处理告警，存 Redis Hash 的 value（JSON）
type pendingAlarmPayload struct {
	CardID      string          `json:"card_id"`
	TenantID    string          `json:"tenant_id"`
	DeviceID    string          `json:"device_id"`
	AlarmType   string          `json:"alarm_type"`
	AlarmLevel  string          `json:"alarm_level"`
	EventSince  int64           `json:"event_since"` // ms
	DurationSec int             `json:"duration_sec"`
	UpgradeTo   string          `json:"upgrade_to,omitempty"`
	TriggerData json.RawMessage `json:"trigger_data,omitempty"`
}

func pendingField(cardID, deviceID, alarmType string) string {
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

// InRestTimeWindow 当前时间是否在租户 ResetTime 休息窗口内，供 event_handler 等调用。
func (s *AlarmService) InRestTimeWindow(ctx context.Context, tenantID string) bool {
	params := s.getResetTimeParamsForTenant(ctx, tenantID)
	return isWithinResetTimeWindow(time.Now(), params, "")
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
	eventName, _ := data[alarm.FieldEventName].(string)
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
	if !isWithinResetTimeWindow(time.Now(), params, "") {
		s.logger.Debug("LeftBed pending skip: outside ResetTimeParams", zap.String("cid", msg.CardID))
		return false
	}
	if s.reader != nil {
		status, _ := s.reader.ReadCardStatus(ctx, msg.CardID)
		if status == nil || status.BedState == nil || status.BedState.BedStatus != 0 {
			s.logger.Debug("LeftBed pending skip: not in_bed", zap.String("cid", msg.CardID))
			return false
		}
		if status.BedState.StartTime == 0 {
			s.logger.Debug("LeftBed pending skip: start_time not set", zap.String("cid", msg.CardID))
			return false
		}
		inBedDurationMs := time.Now().UnixMilli() - status.BedState.StartTime
		if inBedDurationMs < 5*60*1000 {
			s.logger.Debug("LeftBed pending skip: in_bed < 5min", zap.String("cid", msg.CardID), zap.Int64("ms", inBedDurationMs))
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
	if err := s.AddPendingAlarm(ctx, msg.CardID, msg.TenantID, msg.DeviceID, eventName, level, msg.Timestamp, durationSec, upgradeTo, triggerData); err != nil {
		s.logger.Warn("add LeftBed pending at trigger", zap.String("cid", msg.CardID), zap.Error(err))
		return false
	}
	return true
}

// AddPendingAlarm 写入 Redis Hash alarm:pending，field = cardID:deviceID:alarmType。
func (s *AlarmService) AddPendingAlarm(ctx context.Context, cardID, tenantID, deviceID, alarmType, alarmLevel string, eventSinceMs int64, durationSec int, upgradeTo string, triggerData json.RawMessage) error {
	if s.redisPending == nil {
		return nil
	}
	p := pendingAlarmPayload{
		CardID:      cardID,
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
	return s.redisPending.HSet(ctx, redisPendingAlarmKey, pendingField(cardID, deviceID, alarmType), string(b))
}

// RemovePendingAlarm 删除待处理项（如收到 event_status=end）。
func (s *AlarmService) RemovePendingAlarm(ctx context.Context, cardID, deviceID, alarmType string) error {
	if s.redisPending == nil {
		return nil
	}
	return s.redisPending.HDel(ctx, redisPendingAlarmKey, pendingField(cardID, deviceID, alarmType))
}

// AddStayPendingIfEnabled 若该设备已开启 Stay 告警且配置了 duration_sec，则写入 alarm:pending，供 ScanPendingAlarms 到时落库。
func (s *AlarmService) AddStayPendingIfEnabled(ctx context.Context, cardID, tenantID, deviceID string, eventSinceMs int64) error {
	if s.redisPending == nil || deviceID == "" {
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
	return s.AddPendingAlarm(ctx, cardID, tenantID, deviceID, alarm.Stay, level, eventSinceMs, durationSec, "", nil)
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
		durationMs := int64(p.DurationSec) * 1000
		if nowMs-p.EventSince < durationMs {
			continue
		}
		eventType := p.AlarmType
		if p.UpgradeTo != "" {
			eventType = p.UpgradeTo
		}
		triggeredAt := time.Unix(0, p.EventSince*int64(time.Millisecond))
		result, cardAlarmState, err := card.InsertAlarmAndUpdateCard(ctx, s.db, p.CardID, card.AlarmInsertParams{
			TenantID:    p.TenantID,
			DeviceID:    p.DeviceID,
			EventType:   eventType,
			Category:    alarm.GetFHIRCategory(eventType),
			AlarmLevel:  p.AlarmLevel,
			TriggeredAt: triggeredAt,
			TriggerData: p.TriggerData,
		})
		if err != nil {
			s.logger.Warn("scan pending insert alarm failed", zap.String("cid", p.CardID), zap.String("field", field), zap.Error(err))
			continue
		}
		s.logger.Info("pending alarm fired", zap.String("cid", p.CardID), zap.String("event_id", result.EventID), zap.String("type", eventType))
		_ = s.writeAlarmState(ctx, p.CardID, cardAlarmState)
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

	cardAlarmState, result, err := card.AutoResolveDeviceAlarms(ctx, s.db, msg.CardID, msg.TenantID, msg.DeviceID, alarmTypes)
	if err != nil {
		s.logger.Warn("auto resolve failed", zap.String("cid", msg.CardID), zap.Error(err))
		return err
	}
	if result.ResolvedCount == 0 {
		return nil
	}

	s.logger.Info("alarm auto-resolved",
		zap.String("cid", msg.CardID),
		zap.Int("resolved", result.ResolvedCount))

	return s.writeAlarmState(ctx, msg.CardID, cardAlarmState)
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
