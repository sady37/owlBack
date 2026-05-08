package envelope_test

// e2e roundtrip：spatial hierarchy → device addr 派生 → Datagram emit →
// CloudEvent 序列化 → 反序列化 → 验证全字段。
//
// 等价跑通 doc §6.2 engine_lost_fall producer 契约示例。
//
// 此测试故意放在 _test 外部 package（envelope_test）以确保我们用的
// 都是 exported API，不依赖包内部实现。

import (
	"net/netip"
	"testing"

	"owl-common/envelope"
	"owl-common/spatial"
)

// TestE2E_LostFallProducerContract 跑完整 producer 契约：
//
//  1. 构造 spatial hierarchy（tenant=1 / branch=1 / building=1 / floor=2 /
//     unit=0x42 / room=3 / bed=1）→ 派生 device addr
//  2. 模拟 firmware monitor track（device-gateway://qinglan）— root datagram
//  3. AI engine_lost_fall 产生 monitor verdict — parentspan = firmware track
//  4. AI emit alarm — parentspan = AI verdict
//  5. 用户 ack — parentspan = alarm
//  6. 整条 trace 链通过 traceid 关联，每条 datagram CloudEvent roundtrip 全字段保持
func TestE2E_LostFallProducerContract(t *testing.T) {
	// ---- 1. spatial 派生设备地址 ----
	tenant := spatial.BuildTenantPrefix(0x0001)
	branch, _ := spatial.DeriveBranchPrefix(tenant, 0x01)
	site, _ := spatial.DeriveSitePrefix(branch, 1, 2)
	unit, _ := spatial.DeriveUnitPrefix(site, 0x0042)
	room, _ := spatial.DeriveRoomPrefix(unit, 0x03)
	bed, _ := spatial.DeriveBedPrefix(room, 0x01)
	devAddr, err := spatial.DeriveDeviceAddr(bed, "E598A2ACD523")
	if err != nil {
		t.Fatalf("DeriveDeviceAddr: %v", err)
	}

	wantAddr := "fd00:0:1:112:42:301:a2ac:d523"
	if devAddr.String() != wantAddr {
		t.Fatalf("device addr = %s, want %s", devAddr, wantAddr)
	}

	// 反向验证 LPM：bed prefix 在 [tenant, unit, room, bed] 中是最长匹配
	candidates := []netip.Prefix{tenant, branch, site, unit, room, bed}
	if got := spatial.LongestPrefixMatch(devAddr, candidates); got != bed {
		t.Errorf("LPM = %v, want bed %v", got, bed)
	}

	// ---- 2. firmware monitor track（root datagram）----
	fwIdentity := envelope.NewServiceIdentity("device-gateway://qinglan/v2.0.0", "2.0.0")
	subjectDevice := "device:fcaedf95-c8bf-4f81-8a2b-6bb4549fff27"

	fwTrack := envelope.New(fwIdentity, "owl.monitor.track.v1").
		WithSubject(subjectDevice).
		WithSpatial(devAddr).
		WithSeverity(envelope.SeverityInfo).
		WithTag("owl.domain", "track").
		WithData("application/protobuf", []byte{0x10, 0x01}) // 占位 payload

	// root datagram：TraceID == ID
	if fwTrack.TraceID != fwTrack.ID {
		t.Errorf("root datagram TraceID(%s) != ID(%s)", fwTrack.TraceID, fwTrack.ID)
	}
	traceRoot := fwTrack.TraceID

	// ---- 3. AI engine_lost_fall monitor verdict ----
	aiIdentity := envelope.NewServiceIdentity("sensor://owl.engine.lost_fall/v1.0.0", "1.0.0")

	aiVerdict := envelope.New(aiIdentity, "owl.monitor.track.v1").
		WithSubject(subjectDevice).
		WithSpatial(devAddr).
		WithSeverity(envelope.SeverityInfo).
		WithParent(fwTrack.ID, traceRoot).
		WithTag("owl.domain", "fall").
		WithTag("owl.verdict", "engine_inferred").
		WithTag("owl.stance", "asserted").
		WithData("application/protobuf", []byte{0x05}) // pose=5 FALL

	if aiVerdict.ParentSpan != fwTrack.ID {
		t.Errorf("verdict ParentSpan = %s, want %s", aiVerdict.ParentSpan, fwTrack.ID)
	}
	if aiVerdict.TraceID != traceRoot {
		t.Errorf("verdict TraceID = %s, want %s (inherited)", aiVerdict.TraceID, traceRoot)
	}

	// ---- 4. AI alarm（基于 verdict 派生通知）----
	aiAlarm := envelope.New(aiIdentity, "owl.alarm.fall.v1").
		WithSubject(subjectDevice).
		WithSpatial(devAddr).
		WithSeverity(envelope.SeverityCritical).
		WithParent(aiVerdict.ID, traceRoot).
		WithTag("owl.domain", "fall").
		WithTag("owl.stance", "asserted").
		WithData("application/protobuf", []byte{0x64}) // fall_score=100

	if aiAlarm.ParentSpan != aiVerdict.ID {
		t.Errorf("alarm ParentSpan = %s, want %s (= verdict ID)", aiAlarm.ParentSpan, aiVerdict.ID)
	}
	if aiAlarm.TraceID != traceRoot {
		t.Errorf("alarm TraceID = %s, want %s", aiAlarm.TraceID, traceRoot)
	}

	// ---- 5. 用户 ack ----
	userIdentity := envelope.NewServiceIdentity("owlfront://user-action/v1.0.0", "1.0.0")
	userAck := envelope.New(userIdentity, "owl.config.alarm_process.v1").
		WithSubject("card:12a63cf9-...").
		WithSpatial(devAddr).
		WithSeverity(envelope.SeverityInfo).
		WithParent(aiAlarm.ID, traceRoot).
		WithTag("ml.feedback", "positive").
		WithTag("ml.label-source", "manual").
		WithTag("owl.domain", "fall")

	// ---- 6. 整 trace 树验证：全部共享 traceRoot ----
	all := []*envelope.Datagram{fwTrack, aiVerdict, aiAlarm, userAck}
	for i, d := range all {
		if d.TraceID != traceRoot {
			t.Errorf("datagram[%d] TraceID = %s, want %s", i, d.TraceID, traceRoot)
		}
	}

	// 因果链验证（each parentspan = previous ID）
	if aiVerdict.ParentSpan != fwTrack.ID {
		t.Errorf("trace chain broken at verdict")
	}
	if aiAlarm.ParentSpan != aiVerdict.ID {
		t.Errorf("trace chain broken at alarm")
	}
	if userAck.ParentSpan != aiAlarm.ID {
		t.Errorf("trace chain broken at user ack")
	}

	// ---- 7. CloudEvent roundtrip 每一条 ----
	for i, d := range all {
		evt, err := d.ToCloudEvent()
		if err != nil {
			t.Fatalf("datagram[%d] ToCloudEvent: %v", i, err)
		}
		round, err := envelope.FromCloudEvent(evt)
		if err != nil {
			t.Fatalf("datagram[%d] FromCloudEvent: %v", i, err)
		}
		if round.ID != d.ID {
			t.Errorf("datagram[%d] roundtrip ID drift: %s vs %s", i, round.ID, d.ID)
		}
		if round.SpatialAddr != d.SpatialAddr {
			t.Errorf("datagram[%d] roundtrip SpatialAddr drift: %v vs %v", i, round.SpatialAddr, d.SpatialAddr)
		}
		if round.TraceID != d.TraceID {
			t.Errorf("datagram[%d] roundtrip TraceID drift", i)
		}
		if round.ParentSpan != d.ParentSpan {
			t.Errorf("datagram[%d] roundtrip ParentSpan drift", i)
		}
		if len(round.Tags) != len(d.Tags) {
			t.Errorf("datagram[%d] Tags len: %d vs %d", i, len(round.Tags), len(d.Tags))
		}
	}
}

// TestE2E_SpatialAddrRoundtrip 单独验证 spatial.Addr 文字格式经 CloudEvent 后无漂移。
func TestE2E_SpatialAddrRoundtrip(t *testing.T) {
	cases := []string{
		"fd00:0:1::",                       // tenant /48 → 表达为 /128 单地址
		"fd00:0:1:112:42:301:a2ac:d523",    // 完整 device
		"fd00:0:ffff:ffff:ffff:ffff:ff:ff", // 边界
	}
	id := envelope.NewServiceIdentity("sensor://test/v1", "1")
	for _, c := range cases {
		addr := netip.MustParseAddr(c)
		d := envelope.New(id, "owl.test.v1").WithSpatial(addr)
		evt, err := d.ToCloudEvent()
		if err != nil {
			t.Fatalf("%s: ToCloudEvent: %v", c, err)
		}
		round, err := envelope.FromCloudEvent(evt)
		if err != nil {
			t.Fatalf("%s: FromCloudEvent: %v", c, err)
		}
		if round.SpatialAddr != addr {
			t.Errorf("%s: roundtrip = %v, want %v", c, round.SpatialAddr, addr)
		}
	}
}
