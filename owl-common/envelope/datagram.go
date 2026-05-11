// Package envelope implements TDP Datagram v1 — owl 协议的统一信封格式。
//
// 设计参见 owlBack/doc/datagram_envelope.md。
//
// Datagram = CloudEvents v1.0 envelope + IPv6 ULA 寻址 + Protobuf payload +
// W3C TraceContext + syslog severity + dot-namespaced Tags。几乎全部借标准，
// owl 只加少量 extension。
//
// # 用法快速参考
//
//	identity := envelope.NewServiceIdentity("sensor://owl.engine.lost_fall/v1.0.0", "1.0.0")
//	dev, _ := spatial.DeriveDeviceAddr(bedPrefix, "E598A2ACD523")
//
//	dg := envelope.New(identity, "owl.alarm.fall.v1").
//	    WithSubject("device:" + deviceUUID).
//	    WithSpatial(dev).
//	    WithSeverity(envelope.SeverityCritical).
//	    WithTag("owl.domain", "fall").
//	    WithData("application/protobuf", protoBytes)
//
//	// CloudEvents binary mode (HTTP/Kafka/NATS 通用)
//	ce, err := dg.ToCloudEvent()
//
//	// 反向：从 CloudEvent 还原 Datagram
//	dg2, err := envelope.FromCloudEvent(ce)
package envelope

import (
	"crypto/rand"
	"net/netip"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

// SpecVersion CloudEvents 协议版本，固定。
const SpecVersion = "1.0"

// Datagram TDP Datagram v1 的内存形态。
//
// 字段语义详见 doc/datagram_envelope.md §3。
type Datagram struct {
	// CloudEvents 必填字段
	ID     string    // ULID, 每条消息全局唯一（也作 span_id）
	Source string    // Producer URN, e.g. "sensor://owl.engine.lost_fall/v1.0.0"
	Type   string    // Schema FQN, e.g. "owl.alarm.fall.v1"
	Time   time.Time // 事发时刻

	// CloudEvents 推荐字段
	Subject         string // "device:<uuid>" / "card:<uuid>" / "resident:<uuid>"
	DataContentType string // 通常 "application/protobuf"
	Data            []byte // payload 字节流（protobuf marshaled）

	// owl extensions
	SpatialAddr netip.Addr // IPv6 ULA full address (5W where)
	Severity    int        // syslog 0-7 (RFC 5424)
	TraceID     string     // W3C TraceContext，整条业务流；通常 = root datagram ID
	SpanID      string     // = ID（一物两用）
	ParentSpan  string     // 因果上游 datagram ID
	NodeID      string     // 进程实例 UUID（从 ServiceIdentity 来）
	Supersedes  []string   // belief revision: 替换的先前 datagram_id
	Tags        []Tag      // 跨 agent 元数据
}

// Tag 跨 agent 元数据：dot-namespaced 分类维度 + 枚举值。
//
//	{Type:"owl.stance", Value:"asserted"}
//	{Type:"owl.domain", Value:"fall"}
//	{Type:"ml.feedback", Value:"positive"}
//	{Type:"external.fhir.category", Value:"safety"}
//
// 命名空间约定见 doc/datagram_envelope.md §4.5。
type Tag struct {
	Type  string
	Value string
}

// Severity 常量（syslog RFC 5424，0=最严重）。
const (
	SeverityEmergency = 0 // system unusable
	SeverityAlert     = 1 // immediate action required
	SeverityCritical  = 2 // 一般 alarm（fall / 整夜未归 等）
	SeverityError     = 3
	SeverityWarning   = 4
	SeverityNotice    = 5
	SeverityInfo      = 6 // 一般 monitor / event 流
	SeverityDebug     = 7
)

// ULID 生成（线程安全 + 单调递增）
//
// per doc §3.5：每条 datagram 自己 mint ULID 当 ID（也是 span_id）。无须中央协调。
// 单调 entropy 保证同毫秒内 lex-sortable。
var (
	ulidMu      sync.Mutex
	ulidEntropy = ulid.Monotonic(rand.Reader, 0)
)

// NewID 生成新 ULID 字符串（26 字符，lex-sortable）。
func NewID() string {
	ulidMu.Lock()
	defer ulidMu.Unlock()
	return ulid.MustNew(ulid.Timestamp(time.Now()), ulidEntropy).String()
}

// New 创建一条新 Datagram，自动填 ID/SpanID（同值）+ 当前时间 + identity。
//
// dgType 是 schema FQN（e.g. "owl.monitor.track.v1"）。
//
// 调用方按需 chain 各 With* 方法补充字段。最终用 ToCloudEvent 序列化。
func New(identity ServiceIdentity, dgType string) *Datagram {
	id := NewID()
	return &Datagram{
		ID:     id,
		Source: identity.Source,
		Type:   dgType,
		Time:   time.Now().UTC(),
		// 一物两用：ID == SpanID（per doc §3.5）
		SpanID:   id,
		TraceID:  id, // 默认 trace 根；下游 WithParent 会替换
		NodeID:   identity.NodeID,
		Severity: SeverityInfo,
	}
}

// WithSubject 设置 about whom（CloudEvents subject 字段）。
func (d *Datagram) WithSubject(subject string) *Datagram {
	d.Subject = subject
	return d
}

// WithSpatial 设置 5W where（device_ipv6 extension）。
func (d *Datagram) WithSpatial(addr netip.Addr) *Datagram {
	d.SpatialAddr = addr
	return d
}

// WithSeverity 覆盖 severity（默认 SeverityInfo=6）。
func (d *Datagram) WithSeverity(s int) *Datagram {
	d.Severity = s
	return d
}

// WithParent 设置因果上游（parentspan extension）+ 继承 trace。
//
// parentTraceID 应该等于 parent datagram 的 TraceID（同 trace 树）；
// 若空字符串则不动 d.TraceID（默认 d 自己作 trace 根）。
func (d *Datagram) WithParent(parentSpanID, parentTraceID string) *Datagram {
	d.ParentSpan = parentSpanID
	if parentTraceID != "" {
		d.TraceID = parentTraceID
	}
	return d
}

// WithSupersedes 标记本 datagram 替换若干旧 datagram（belief revision）。
func (d *Datagram) WithSupersedes(ids ...string) *Datagram {
	d.Supersedes = append(d.Supersedes, ids...)
	return d
}

// WithTag 追加一个 Tag。重复 type 不去重——consumer 自处理（per doc §4.4）。
func (d *Datagram) WithTag(tagType, value string) *Datagram {
	d.Tags = append(d.Tags, Tag{Type: tagType, Value: value})
	return d
}

// WithData 设置 payload（content type + 字节）。
//
// 通常 contentType="application/protobuf" + protoMessage marshal 字节。
// JSON 旧 producer 过渡期可用 "application/json"。
func (d *Datagram) WithData(contentType string, data []byte) *Datagram {
	d.DataContentType = contentType
	d.Data = data
	return d
}
