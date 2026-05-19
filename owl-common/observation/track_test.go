// track_test.go — Track.ToFieldMap / FromFieldMap Tier1。
//
// BedStatus *int 三态：nil = 未知 / 不适用；*0 = 在床；*1 = 离床。
// 雷达 / heartbeat 不填即 nil；sleepace InBed 写 *0；LeftBed/silent_fall 写 *1。

package observation

import "testing"

func ptrInt(v int) *int { return &v }

func TestTrack_ToFieldMap_OmitsNilBedStatus(t *testing.T) {
	// 默认 BedStatus nil（雷达 heartbeat / radarTrackToData 路径）→ 不应出现 bed_status 键
	tr := Track{TrackID: 11, TrackConfidence: 80}
	m := tr.ToFieldMap()
	if _, ok := m[FieldBedStatus]; ok {
		t.Errorf("BedStatus=nil must be omitted; got map=%+v", m)
	}
}

func TestTrack_ToFieldMap_WritesInBedZero(t *testing.T) {
	// sleepace 在床 = *0 → 必须写出 0（区别于其它 int 字段的 omit-when-zero）
	tr := Track{TrackID: 0, BedStatus: ptrInt(0)}
	m := tr.ToFieldMap()
	got, ok := m[FieldBedStatus]
	if !ok {
		t.Fatalf("BedStatus=*0 must be written; got map=%+v", m)
	}
	if got != 0 {
		t.Errorf("want bed_status=0, got %v", got)
	}
}

func TestTrack_ToFieldMap_WritesLeftBedOne(t *testing.T) {
	// sleepace LeftBed / silent_fall alarm 写 *1
	tr := Track{TrackID: 0, BedStatus: ptrInt(1)}
	m := tr.ToFieldMap()
	got, ok := m[FieldBedStatus]
	if !ok {
		t.Fatalf("BedStatus=*1 must be written; got map=%+v", m)
	}
	if got != 1 {
		t.Errorf("want bed_status=1, got %v", got)
	}
}

func TestTrack_FromFieldMap_MissingKeyLeavesNil(t *testing.T) {
	// 流里没 bed_status 键（雷达 heartbeat 等）→ BedStatus 留 nil
	var out Track
	out.FromFieldMap(map[string]any{FieldTrackID: 11})
	if out.BedStatus != nil {
		t.Errorf("missing key must leave BedStatus nil, got *%d", *out.BedStatus)
	}
}

func TestTrack_RoundTrip_InBedZero(t *testing.T) {
	in := Track{TrackID: 0, BedStatus: ptrInt(0)}
	m := in.ToFieldMap()
	var out Track
	out.FromFieldMap(m)
	if out.BedStatus == nil || *out.BedStatus != 0 {
		t.Errorf("round-trip *0: want *0, got %v", out.BedStatus)
	}
}

func TestTrack_RoundTrip_LeftBedOne(t *testing.T) {
	in := Track{TrackID: 0, BedStatus: ptrInt(1)}
	m := in.ToFieldMap()
	var out Track
	out.FromFieldMap(m)
	if out.BedStatus == nil || *out.BedStatus != 1 {
		t.Errorf("round-trip *1: want *1, got %v", out.BedStatus)
	}
}

func TestTrack_RoundTrip_NilUnknown(t *testing.T) {
	// 雷达 Track{} 默认 BedStatus nil → encode 不写 → decode 仍 nil
	in := Track{TrackID: 11}
	m := in.ToFieldMap()
	var out Track
	out.FromFieldMap(m)
	if out.BedStatus != nil {
		t.Errorf("round-trip nil: want nil, got *%d", *out.BedStatus)
	}
}

func TestTrack_ToFieldMap_PreservesOtherZeroOmissions(t *testing.T) {
	// 其它 int 字段维持 omit-when-zero（HeartRate=0 / Pose=0 / etc）
	tr := Track{TrackID: 11}
	m := tr.ToFieldMap()
	for _, k := range []string{FieldHeartRate, FieldRespiratoryRate, FieldPose, FieldEvent, FieldBodyMove, FieldTurnOver, FieldBedStatus} {
		if _, ok := m[k]; ok {
			t.Errorf("zero %q must be omitted; got %+v", k, m)
		}
	}
	if _, ok := m[FieldTrackID]; !ok {
		t.Errorf("TrackID must always be written; got %+v", m)
	}
}
