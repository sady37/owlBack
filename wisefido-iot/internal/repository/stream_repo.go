// stream_repo.go — wisefido-iot 单一持久化边界：按 topic_type 路由写 v2 三表。
//
//	iot:monitor:stream → monitor_stream (90d hypertable)
//	iot:event:stream   → event_log      (90d hypertable)
//	iot:alarm:stream   → 不写 (cardagg alarm_router 是 alarm_events 单 writer)
//
// 跟 cardagg 拆责任：cardagg 拥有 alarm domain（PG INSERT + AlarmState 投影 + 业务判定）；
// iot 拥有 transport→DB persistence 边界（hot path raw 监控 + 离散事件 history）。
// 双方都不交叉，CLAUDE.md 规则 #1.3 单源真相。

package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/netip"
	"strings"
	"time"

	owlredis "owl-common/redis"

	"go.uber.org/zap"
)

// StreamRepo iot:monitor / iot:event 流持久化入口。
type StreamRepo struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewStreamRepo(db *sql.DB, logger *zap.Logger) *StreamRepo {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &StreamRepo{db: db, logger: logger}
}

// InsertMonitor 写一条 monitor_stream（高频 raw，每条 device sample）。
func (r *StreamRepo) InsertMonitor(ctx context.Context, msg *owlredis.IoTStreamMessage) error {
	if r.db == nil {
		return nil
	}
	if !msg.DeviceAddr.IsValid() {
		return fmt.Errorf("device_addr invalid")
	}
	streamType := buildStreamType(msg.DeviceType, msg.Category)
	if streamType == "" {
		return fmt.Errorf("cannot derive stream_type from device_type=%q category=%q",
			msg.DeviceType, msg.Category)
	}
	payload, err := json.Marshal(msg.DataValue)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	traceID := buildTraceID(msg.Producer, msg.SequenceNumber)
	ts := tsFromMs(msg.Timestamp)

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO monitor_stream (
			ts, device_addr, device_uid, device_type, stream_type, payload,
			datagram_id, trace_id, parent_span, severity, tags
		) VALUES ($1, $2::INET, NULLIF($3, ''), $4::device_type_enum, $5, $6::JSONB,
		          NULL, NULLIF($7, ''), NULLIF($7, ''), 6, NULL)
	`, ts, msg.DeviceAddr.String(), deviceUID(msg.SubjectEntity), msg.DeviceType, streamType, string(payload), traceID)
	if err != nil {
		return fmt.Errorf("insert monitor_stream: %w", err)
	}
	return nil
}

// InsertEvent 写一条 event_log（离散事件：room.enter / fall.suspect / device.signal_poor / ...）。
// event_kind 直存 envelope.Category（CLAUDE.md 规则 #1.1：observation 包定义的字面量，不做翻译）。
func (r *StreamRepo) InsertEvent(ctx context.Context, msg *owlredis.IoTStreamMessage) error {
	if r.db == nil {
		return nil
	}
	if !msg.DeviceAddr.IsValid() {
		return fmt.Errorf("device_addr invalid")
	}
	if msg.Category == "" {
		return fmt.Errorf("event_kind (category) empty")
	}
	payload, err := json.Marshal(msg.DataValue)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	traceID := buildTraceID(msg.Producer, msg.SequenceNumber)
	ts := tsFromMs(msg.Timestamp)

	subjAddr := subjectAddr(msg.SubjectEntity)
	uid := deviceUID(msg.SubjectEntity)
	if uid == "" {
		if s, ok := subjAddr.(string); ok {
			u, lerr := r.lookupDeviceUID(ctx, s)
			if lerr != nil {
				return fmt.Errorf("lookup device_uid for subject %s: %w", s, lerr)
			}
			uid = u
		}
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO event_log (
			ts, device_addr, device_uid, event_kind, subject_addr, payload,
			datagram_id, trace_id, parent_span, severity, tags
		) VALUES ($1, $2::INET, NULLIF($3, ''), $4, $5::INET, $6::JSONB,
		          NULL, NULLIF($7, ''), NULLIF($7, ''), 6, NULL)
	`, ts, msg.DeviceAddr.String(), uid, msg.Category, subjAddr, string(payload), traceID)
	if err != nil {
		return fmt.Errorf("insert event_log: %w", err)
	}
	return nil
}

// InsertSensorDecision 写一条 sensor_decision_log（旁路审计：ghost verdict / lost-fall 决策；见 dbv2/66_）。
// 来源 ai:track:verdict:stream；仅落带 decision_event 的消息（verdict 流上其它消息跳过）。
func (r *StreamRepo) InsertSensorDecision(ctx context.Context, msg *owlredis.IoTStreamMessage) error {
	if r.db == nil {
		return nil
	}
	if !msg.DeviceAddr.IsValid() {
		return fmt.Errorf("device_addr invalid")
	}
	data := owlredis.FirstDataValue(msg.DataValue)
	if data == nil {
		return nil
	}
	event := decStr(data["decision_event"])
	if event == "" {
		return nil // 非决策审计事件（如纯 realtime override），不落库
	}
	var verdict, conf, evidenceJSON interface{}
	if c, ok := data["track_confidence"]; ok {
		conf = decInt(c)
	}
	if ev, ok := data["evidence"].(map[string]interface{}); ok {
		if v, ok2 := ev["verdict"]; ok2 {
			verdict = decInt(v)
		}
		if b, err := json.Marshal(ev); err == nil {
			evidenceJSON = string(b)
		}
	}
	traceID := buildTraceID(msg.Producer, msg.SequenceNumber)
	ts := tsFromMs(msg.Timestamp)
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sensor_decision_log (
			ts, device_addr, track_id, logic_id, event,
			verdict, track_confidence, reason, evidence, trace_id
		) VALUES ($1, $2::INET, $3, NULLIF($4, ''), $5,
		          $6, $7, NULLIF($8, ''), $9::JSONB, NULLIF($10, ''))
		ON CONFLICT (device_addr, ts, track_id, event) DO NOTHING
	`, ts, msg.DeviceAddr.String(), decInt(data["track_id"]), decStr(data["logic_id"]), event,
		verdict, conf, decStr(data["reason"]), evidenceJSON, traceID)
	if err != nil {
		return fmt.Errorf("insert sensor_decision_log: %w", err)
	}
	return nil
}

// decInt / decStr：从 JSON-decoded（数值多为 float64）的 dataValue map 取值。
func decInt(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	}
	return 0
}

func decStr(v interface{}) string {
	s, _ := v.(string)
	return s
}

// subjectAddr 只在 SubjectEntity 是合法 INET（sensor 派生的 spatial prefix /88·/96·/128）时
// 落 subject_addr；device-gateway 发的 device_uid（vendor MAC 串）非 INET → NULL（匿名事件，
// device 身份已在 device_addr，device_uid 可由 device_addr 反查 dfm）。空串 → NULL。
func subjectAddr(subjectEntity string) interface{} {
	s := strings.TrimSpace(subjectEntity)
	if s == "" {
		return nil
	}
	if _, err := netip.ParsePrefix(s); err == nil {
		return s
	}
	if _, err := netip.ParseAddr(s); err == nil {
		return s
	}
	return nil
}

// lookupDeviceUID 把派生事件的 subject_addr（spatial prefix）解析成对应设备的 device_uid。
// 仅 /128 subject 能等值命中 devices.device_addr；room/bed prefix（/88·/96）不对应单一设备 →
// ErrNoRows → 空串（落 NULL，由调用方写入），属正常业务语义不是错误；真 DB 错误向上传播。
func (r *StreamRepo) lookupDeviceUID(ctx context.Context, addr string) (string, error) {
	var uid string
	err := r.db.QueryRowContext(ctx,
		`SELECT device_uid FROM devices WHERE device_addr = $1::INET`, addr).Scan(&uid)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return uid, nil
}

// deviceUID 是 subjectAddr 的互补：SubjectEntity 是 device-gateway raw 事件填的 device_uid
// （非 INET 的 logMAC 串）时返回它落 device_uid 列；sensor 派生事件填的 spatial prefix（INET）
// 或空串 → 返回空串（NULL）。raw=uid / derived=prefix 二选一，两列互斥落值。
func deviceUID(subjectEntity string) string {
	s := strings.TrimSpace(subjectEntity)
	if s == "" {
		return ""
	}
	if _, err := netip.ParsePrefix(s); err == nil {
		return ""
	}
	if _, err := netip.ParseAddr(s); err == nil {
		return ""
	}
	return s
}

// buildStreamType 构造 dot-namespaced stream_type："<deviceType>.<category>" 全小写。
// 例：(Radar, track) → "radar.track"; (Sleepad, vital) → "sleepad.vital"
func buildStreamType(deviceType, category string) string {
	dt := strings.ToLower(strings.TrimSpace(deviceType))
	cat := strings.ToLower(strings.TrimSpace(category))
	if dt == "" || cat == "" {
		return ""
	}
	return dt + "." + cat
}

// buildTraceID 北极星 datagram ref："<producer>.<seqN>"；producer/seqN 任一缺失返空串。
func buildTraceID(producer string, seqN uint64) string {
	if producer == "" || seqN == 0 {
		return ""
	}
	return fmt.Sprintf("%s.%d", producer, seqN)
}

// tsFromMs unix ms → UTC time.Time；0 fallback now()。
func tsFromMs(ms int64) time.Time {
	if ms <= 0 {
		return time.Now().UTC()
	}
	return time.UnixMilli(ms).UTC()
}
