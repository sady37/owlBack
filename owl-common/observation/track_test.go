// track_test.go — Track.ToFieldMap / FromFieldMap Tier1。
//
// BedStatus int 三态：InBed(0) = 在床；LeftBed(1) = 离床；Unchanged(8) = 无床数据/不适用。
// 雷达 / heartbeat 显式置 Unchanged；sleepace InBed 写 0；LeftBed/silent_fall 写 1。

package observation

import "testing"

func ptrInt(v int) *int { return &v }

func TestTrack_ToFieldMap_OmitsUnchangedBedStatus(t *testing.T) {
	// 无床概念设备（雷达 heartbeat / radarTrackToData 路径）显式置 Unchanged → 不应出现 bed_status 键
	tr := Track{TrackID: 11, TrackConfidence: 80, BedStatus: BedStatusUnchanged}
	m := tr.ToFieldMap()
	if _, ok := m[FieldBedStatus]; ok {
		t.Errorf("BedStatus=Unchanged must be omitted; got map=%+v", m)
	}
}

func TestTrack_ToFieldMap_WritesInBedZero(t *testing.T) {
	// sleepace 在床 = InBed(0) → 必须写出 0（区别于其它 int 字段的 omit-when-zero）
	tr := Track{TrackID: 0, BedStatus: BedStatusInBed}
	m := tr.ToFieldMap()
	got, ok := m[FieldBedStatus]
	if !ok {
		t.Fatalf("BedStatus=InBed must be written; got map=%+v", m)
	}
	if got != 0 {
		t.Errorf("want bed_status=0, got %v", got)
	}
}

func TestTrack_ToFieldMap_WritesLeftBedOne(t *testing.T) {
	// sleepace LeftBed / silent_fall alarm 写 1
	tr := Track{TrackID: 0, BedStatus: BedStatusLeftBed}
	m := tr.ToFieldMap()
	got, ok := m[FieldBedStatus]
	if !ok {
		t.Fatalf("BedStatus=LeftBed must be written; got map=%+v", m)
	}
	if got != 1 {
		t.Errorf("want bed_status=1, got %v", got)
	}
}

func TestTrack_FromFieldMap_MissingKeyDefaultsUnchanged(t *testing.T) {
	// 流里没 bed_status 键（雷达 heartbeat 等）→ BedStatus 默认 Unchanged(8)
	var out Track
	out.FromFieldMap(map[string]any{FieldTrackID: 11})
	if out.BedStatus != BedStatusUnchanged {
		t.Errorf("missing key must default BedStatus=Unchanged, got %d", out.BedStatus)
	}
}

func TestTrack_RoundTrip_InBedZero(t *testing.T) {
	in := Track{TrackID: 0, BedStatus: BedStatusInBed}
	m := in.ToFieldMap()
	var out Track
	out.FromFieldMap(m)
	if out.BedStatus != BedStatusInBed {
		t.Errorf("round-trip InBed: want 0, got %d", out.BedStatus)
	}
}

func TestTrack_RoundTrip_LeftBedOne(t *testing.T) {
	in := Track{TrackID: 0, BedStatus: BedStatusLeftBed}
	m := in.ToFieldMap()
	var out Track
	out.FromFieldMap(m)
	if out.BedStatus != BedStatusLeftBed {
		t.Errorf("round-trip LeftBed: want 1, got %d", out.BedStatus)
	}
}

func TestTrack_RoundTrip_Unchanged(t *testing.T) {
	// 无床概念设备显式置 Unchanged(8) → encode 不写 → decode 回 Unchanged
	in := Track{TrackID: 11, BedStatus: BedStatusUnchanged}
	m := in.ToFieldMap()
	if _, ok := m[FieldBedStatus]; ok {
		t.Errorf("Unchanged must be omitted from map; got %+v", m)
	}
	var out Track
	out.FromFieldMap(m)
	if out.BedStatus != BedStatusUnchanged {
		t.Errorf("round-trip Unchanged: want 8, got %d", out.BedStatus)
	}
}

func TestTrack_ToFieldMap_PreservesOtherZeroOmissions(t *testing.T) {
	// 其它 int 字段维持 omit-when-zero（HeartRate=0 / Pose=0 / etc）；BedStatus 用 Unchanged 同样省略
	tr := Track{TrackID: 11, BedStatus: BedStatusUnchanged}
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
