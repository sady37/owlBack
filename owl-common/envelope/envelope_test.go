package envelope

import (
	"net/netip"
	"strings"
	"testing"
	"time"
)

// =============================================================================
// ULID & Identity
// =============================================================================

func TestNewID_uniqueAndSortable(t *testing.T) {
	a := NewID()
	b := NewID()
	if a == b {
		t.Errorf("NewID returned duplicate: %s", a)
	}
	// ULID 26 chars, lex-sortable; 同毫秒内单调递增
	if len(a) != 26 || len(b) != 26 {
		t.Errorf("ULID length: a=%d b=%d, want 26", len(a), len(b))
	}
	if a >= b {
		t.Errorf("expected a < b lex-sortable, got a=%s b=%s", a, b)
	}
}

func TestNewServiceIdentity(t *testing.T) {
	id := NewServiceIdentity("sensor://owl.engine.lost_fall/v1.0.0", "1.0.0")
	if id.Source != "sensor://owl.engine.lost_fall/v1.0.0" {
		t.Errorf("Source = %q", id.Source)
	}
	if id.Version != "1.0.0" {
		t.Errorf("Version = %q", id.Version)
	}
	if len(id.NodeID) != 36 { // UUID v4 string format
		t.Errorf("NodeID length = %d, want 36 (UUID)", len(id.NodeID))
	}

	// 两次创建 NodeID 不同
	id2 := NewServiceIdentity("sensor://owl.engine.lost_fall/v1.0.0", "1.0.0")
	if id.NodeID == id2.NodeID {
		t.Errorf("two NewServiceIdentity returned same NodeID")
	}
}

// =============================================================================
// Datagram New + Builders
// =============================================================================

func TestNew_AutoFields(t *testing.T) {
	id := NewServiceIdentity("sensor://test/v1", "1")
	d := New(id, "owl.test.v1")

	if d.ID == "" {
		t.Errorf("ID empty")
	}
	if d.SpanID != d.ID {
		t.Errorf("SpanID(%s) != ID(%s) — must be 一物两用", d.SpanID, d.ID)
	}
	if d.TraceID != d.ID {
		t.Errorf("TraceID(%s) != ID(%s) — root datagram TraceID == ID by default", d.TraceID, d.ID)
	}
	if d.NodeID != id.NodeID {
		t.Errorf("NodeID = %q, want %q", d.NodeID, id.NodeID)
	}
	if d.Source != id.Source {
		t.Errorf("Source = %q, want %q", d.Source, id.Source)
	}
	if d.Type != "owl.test.v1" {
		t.Errorf("Type = %q", d.Type)
	}
	if d.Severity != SeverityInfo {
		t.Errorf("default Severity = %d, want %d (Info)", d.Severity, SeverityInfo)
	}
	if d.Time.IsZero() {
		t.Errorf("Time is zero")
	}
	if d.Time.Location() != time.UTC {
		t.Errorf("Time not UTC: %s", d.Time)
	}
}

func TestDatagram_Builders(t *testing.T) {
	id := NewServiceIdentity("sensor://test/v1", "1")
	addr := netip.MustParseAddr("fd00:0:1:112:42:301:a2ac:d523")

	d := New(id, "owl.alarm.fall.v1").
		WithSubject("device:abc-123").
		WithSpatial(addr).
		WithSeverity(SeverityCritical).
		WithTag("owl.domain", "fall").
		WithTag("owl.stance", "asserted").
		WithData("application/protobuf", []byte{0x01, 0x02, 0x03})

	if d.Subject != "device:abc-123" {
		t.Errorf("Subject = %q", d.Subject)
	}
	if d.SpatialAddr != addr {
		t.Errorf("SpatialAddr = %v, want %v", d.SpatialAddr, addr)
	}
	if d.Severity != SeverityCritical {
		t.Errorf("Severity = %d, want %d", d.Severity, SeverityCritical)
	}
	if len(d.Tags) != 2 {
		t.Fatalf("Tags len = %d, want 2", len(d.Tags))
	}
	if d.Tags[0] != (Tag{Type: "owl.domain", Value: "fall"}) {
		t.Errorf("Tags[0] = %+v", d.Tags[0])
	}
	if d.DataContentType != "application/protobuf" {
		t.Errorf("DataContentType = %q", d.DataContentType)
	}
	if len(d.Data) != 3 {
		t.Errorf("Data len = %d", len(d.Data))
	}
}

func TestDatagram_WithParent_inheritsTrace(t *testing.T) {
	id := NewServiceIdentity("sensor://test/v1", "1")

	parent := New(id, "owl.monitor.track.v1")
	child := New(id, "owl.alarm.fall.v1").
		WithParent(parent.ID, parent.TraceID)

	if child.ParentSpan != parent.ID {
		t.Errorf("child.ParentSpan = %s, want %s", child.ParentSpan, parent.ID)
	}
	if child.TraceID != parent.TraceID {
		t.Errorf("child.TraceID = %s, want %s (inherited)", child.TraceID, parent.TraceID)
	}
	// 自己 SpanID 不应被 WithParent 改
	if child.SpanID != child.ID {
		t.Errorf("child.SpanID = %s, want = child.ID %s", child.SpanID, child.ID)
	}
}

func TestDatagram_WithParent_emptyTraceKeepsOwn(t *testing.T) {
	id := NewServiceIdentity("sensor://test/v1", "1")
	d := New(id, "owl.test.v1")
	originalTrace := d.TraceID

	d.WithParent("some-parent-id", "")

	if d.ParentSpan != "some-parent-id" {
		t.Errorf("ParentSpan = %q", d.ParentSpan)
	}
	if d.TraceID != originalTrace {
		t.Errorf("TraceID changed to %s, want %s (no override)", d.TraceID, originalTrace)
	}
}

// =============================================================================
// CloudEvents Roundtrip
// =============================================================================

func TestRoundtrip_CloudEvent(t *testing.T) {
	id := NewServiceIdentity("sensor://owl.engine.lost_fall/v1.0.0", "1.0.0")
	addr := netip.MustParseAddr("fd00:0:1:112:42:301:a2ac:d523")

	original := New(id, "owl.alarm.fall.v1").
		WithSubject("device:fcaedf95-c8bf-4f81-8a2b-6bb4549fff27").
		WithSpatial(addr).
		WithSeverity(SeverityCritical).
		WithTag("owl.domain", "fall").
		WithTag("owl.stance", "asserted").
		WithSupersedes("01HXY-OLD-1", "01HXY-OLD-2").
		WithData("application/protobuf", []byte{0xDE, 0xAD, 0xBE, 0xEF})

	// 设置 parentspan + traceid 模拟下游
	original.ParentSpan = "01HXY-PARENT"
	original.TraceID = "01HXY-TRACE-ROOT"

	evt, err := original.ToCloudEvent()
	if err != nil {
		t.Fatalf("ToCloudEvent: %v", err)
	}

	// CloudEvents 必填字段
	if evt.SpecVersion() != SpecVersion {
		t.Errorf("specversion = %s", evt.SpecVersion())
	}
	if evt.ID() != original.ID {
		t.Errorf("evt.ID = %s", evt.ID())
	}
	if evt.Source() != original.Source {
		t.Errorf("evt.Source = %s", evt.Source())
	}
	if evt.Type() != original.Type {
		t.Errorf("evt.Type = %s", evt.Type())
	}

	// 反向解析
	round, err := FromCloudEvent(evt)
	if err != nil {
		t.Fatalf("FromCloudEvent: %v", err)
	}

	if round.ID != original.ID {
		t.Errorf("ID: %s vs %s", round.ID, original.ID)
	}
	if round.Source != original.Source {
		t.Errorf("Source: %s vs %s", round.Source, original.Source)
	}
	if round.Type != original.Type {
		t.Errorf("Type: %s vs %s", round.Type, original.Type)
	}
	if round.Subject != original.Subject {
		t.Errorf("Subject: %s vs %s", round.Subject, original.Subject)
	}
	if round.SpatialAddr != original.SpatialAddr {
		t.Errorf("SpatialAddr: %s vs %s", round.SpatialAddr, original.SpatialAddr)
	}
	if round.Severity != original.Severity {
		t.Errorf("Severity: %d vs %d", round.Severity, original.Severity)
	}
	if round.TraceID != original.TraceID {
		t.Errorf("TraceID: %s vs %s", round.TraceID, original.TraceID)
	}
	if round.SpanID != original.SpanID {
		t.Errorf("SpanID: %s vs %s", round.SpanID, original.SpanID)
	}
	if round.ParentSpan != original.ParentSpan {
		t.Errorf("ParentSpan: %s vs %s", round.ParentSpan, original.ParentSpan)
	}
	if round.NodeID != original.NodeID {
		t.Errorf("NodeID: %s vs %s", round.NodeID, original.NodeID)
	}
	if len(round.Supersedes) != 2 || round.Supersedes[0] != "01HXY-OLD-1" {
		t.Errorf("Supersedes = %+v", round.Supersedes)
	}
	if len(round.Tags) != 2 {
		t.Fatalf("Tags len = %d, want 2", len(round.Tags))
	}
	if round.Tags[0] != (Tag{Type: "owl.domain", Value: "fall"}) {
		t.Errorf("Tags[0] = %+v", round.Tags[0])
	}
	if round.DataContentType != "application/protobuf" {
		t.Errorf("DataContentType = %q", round.DataContentType)
	}
	if string(round.Data) != string(original.Data) {
		t.Errorf("Data: %x vs %x", round.Data, original.Data)
	}
	// 时间精度可能丢失到 ms，给 1ms 容差
	if diff := round.Time.Sub(original.Time); diff > time.Millisecond || diff < -time.Millisecond {
		t.Errorf("Time drifted: %s vs %s (diff %v)", round.Time, original.Time, diff)
	}
}

func TestToCloudEvent_RequiredFields(t *testing.T) {
	cases := []struct {
		name string
		d    *Datagram
		want string
	}{
		{"missing ID", &Datagram{Source: "s", Type: "t"}, "ID is required"},
		{"missing Source", &Datagram{ID: "x", Type: "t"}, "Source is required"},
		{"missing Type", &Datagram{ID: "x", Source: "s"}, "Type is required"},
	}
	for _, c := range cases {
		_, err := c.d.ToCloudEvent()
		if err == nil {
			t.Errorf("%s: expected err, got nil", c.name)
			continue
		}
		if !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: err = %v, want substring %q", c.name, err, c.want)
		}
	}
}

func TestFromCloudEvent_BadSpecVersion(t *testing.T) {
	// 假设 0.3 旧版 — sdk-go 仍支持，但我们应拒绝
	id := NewServiceIdentity("sensor://test/v1", "1")
	d := New(id, "owl.test.v1")
	evt, err := d.ToCloudEvent()
	if err != nil {
		t.Fatalf("ToCloudEvent: %v", err)
	}
	// 篡改 specversion
	evt.SetSpecVersion("0.3")

	if _, err := FromCloudEvent(evt); err == nil {
		t.Errorf("expected error on specversion mismatch")
	}
}
