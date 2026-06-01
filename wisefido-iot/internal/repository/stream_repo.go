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
			ts, device_addr, device_type, stream_type, payload,
			datagram_id, trace_id, parent_span, severity, tags
		) VALUES ($1, $2::INET, $3::device_type_enum, $4, $5::JSONB,
		          NULL, NULLIF($6, ''), NULLIF($6, ''), 6, NULL)
	`, ts, msg.DeviceAddr.String(), msg.DeviceType, streamType, string(payload), traceID)
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

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO event_log (
			ts, device_addr, event_kind, subject_addr, payload,
			datagram_id, trace_id, parent_span, severity, tags
		) VALUES ($1, $2::INET, $3, $4::INET, $5::JSONB,
		          NULL, NULLIF($6, ''), NULLIF($6, ''), 6, NULL)
	`, ts, msg.DeviceAddr.String(), msg.Category, subjectAddr(msg.SubjectEntity), string(payload), traceID)
	if err != nil {
		return fmt.Errorf("insert event_log: %w", err)
	}
	return nil
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
