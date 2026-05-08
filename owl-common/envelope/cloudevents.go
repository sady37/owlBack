package envelope

import (
	"encoding/json"
	"fmt"
	"net/netip"
	"strconv"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

// CloudEvents extension 字段名（小写，per CloudEvents 规范）。
const (
	extSpatialAddr = "spatialaddr"
	extSeverity    = "severity"
	extTraceID     = "traceid"
	extSpanID      = "spanid"
	extParentSpan  = "parentspan"
	extNodeID      = "nodeid"
	extSupersedes  = "supersedes" // JSON encoded []string
	extTags        = "tags"       // JSON encoded []Tag
)

// ToCloudEvent 把 Datagram 序列化为 cloudevents.Event。
//
// 输出可用 cloudevents.HTTP / Kafka / NATS protocol binding 直接发送
// （binary mode 默认）。Extension 字段都用小写名（CloudEvents 要求）。
//
// 列表型字段（Supersedes / Tags）用 JSON 字符串序列化进 extension —— 标准
// extension 只支持 string/int/bool/binary/time。这层 JSON 是 owl 私有约定，
// 反向 FromCloudEvent 解开即可。
func (d *Datagram) ToCloudEvent() (cloudevents.Event, error) {
	evt := cloudevents.NewEvent(SpecVersion)

	if d.ID == "" {
		return evt, fmt.Errorf("envelope: Datagram.ID is required")
	}
	if d.Source == "" {
		return evt, fmt.Errorf("envelope: Datagram.Source is required")
	}
	if d.Type == "" {
		return evt, fmt.Errorf("envelope: Datagram.Type is required")
	}

	evt.SetID(d.ID)
	evt.SetSource(d.Source)
	evt.SetType(d.Type)

	if !d.Time.IsZero() {
		evt.SetTime(d.Time.UTC())
	}
	if d.Subject != "" {
		evt.SetSubject(d.Subject)
	}

	// 必装的 owl extensions
	if d.SpatialAddr.IsValid() {
		evt.SetExtension(extSpatialAddr, d.SpatialAddr.String())
	}
	evt.SetExtension(extSeverity, d.Severity)
	if d.TraceID != "" {
		evt.SetExtension(extTraceID, d.TraceID)
	}
	if d.SpanID != "" {
		evt.SetExtension(extSpanID, d.SpanID)
	}
	if d.ParentSpan != "" {
		evt.SetExtension(extParentSpan, d.ParentSpan)
	}
	if d.NodeID != "" {
		evt.SetExtension(extNodeID, d.NodeID)
	}

	// 列表型 extension 用 JSON 编码
	if len(d.Supersedes) > 0 {
		buf, err := json.Marshal(d.Supersedes)
		if err != nil {
			return evt, fmt.Errorf("envelope: marshal supersedes: %w", err)
		}
		evt.SetExtension(extSupersedes, string(buf))
	}
	if len(d.Tags) > 0 {
		buf, err := json.Marshal(d.Tags)
		if err != nil {
			return evt, fmt.Errorf("envelope: marshal tags: %w", err)
		}
		evt.SetExtension(extTags, string(buf))
	}

	// Payload
	if d.DataContentType != "" || len(d.Data) > 0 {
		if err := evt.SetData(d.DataContentType, d.Data); err != nil {
			return evt, fmt.Errorf("envelope: SetData: %w", err)
		}
	}

	return evt, nil
}

// FromCloudEvent 把 cloudevents.Event 还原为 Datagram。
//
// 反向 ToCloudEvent。未知 extension 会被忽略；缺失的 owl extension 字段保留零值。
func FromCloudEvent(evt cloudevents.Event) (*Datagram, error) {
	if evt.SpecVersion() != SpecVersion {
		return nil, fmt.Errorf("envelope: unsupported specversion %q (want %q)",
			evt.SpecVersion(), SpecVersion)
	}

	d := &Datagram{
		ID:              evt.ID(),
		Source:          evt.Source(),
		Type:            evt.Type(),
		Subject:         evt.Subject(),
		Time:            evt.Time(),
		DataContentType: evt.DataContentType(),
		Data:            evt.Data(),
	}

	exts := evt.Extensions()

	if v, ok := exts[extSpatialAddr]; ok {
		s, _ := v.(string)
		if s != "" {
			addr, err := netip.ParseAddr(s)
			if err != nil {
				return nil, fmt.Errorf("envelope: invalid spatialaddr %q: %w", s, err)
			}
			d.SpatialAddr = addr
		}
	}
	if v, ok := exts[extSeverity]; ok {
		d.Severity = parseSeverity(v)
	}
	if v, ok := exts[extTraceID]; ok {
		d.TraceID, _ = v.(string)
	}
	if v, ok := exts[extSpanID]; ok {
		d.SpanID, _ = v.(string)
	}
	if v, ok := exts[extParentSpan]; ok {
		d.ParentSpan, _ = v.(string)
	}
	if v, ok := exts[extNodeID]; ok {
		d.NodeID, _ = v.(string)
	}
	if v, ok := exts[extSupersedes]; ok {
		s, _ := v.(string)
		if s != "" {
			if err := json.Unmarshal([]byte(s), &d.Supersedes); err != nil {
				return nil, fmt.Errorf("envelope: unmarshal supersedes: %w", err)
			}
		}
	}
	if v, ok := exts[extTags]; ok {
		s, _ := v.(string)
		if s != "" {
			if err := json.Unmarshal([]byte(s), &d.Tags); err != nil {
				return nil, fmt.Errorf("envelope: unmarshal tags: %w", err)
			}
		}
	}

	return d, nil
}

// parseSeverity 兼容 CloudEvents extension 可能拿到 int / int32 / string 的几种形态。
func parseSeverity(v any) int {
	switch x := v.(type) {
	case int:
		return x
	case int32:
		return int(x)
	case int64:
		return int(x)
	case string:
		n, err := strconv.Atoi(x)
		if err == nil {
			return n
		}
	}
	return SeverityInfo
}

// Ensure time package import not removed by linter.
var _ = time.Now
