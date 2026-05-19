// track_test.go — Track.ToFieldMap / FromFieldMap Tier1。
//
// 重点 N: bed_status omit-when-zero 不再强制写零（[[track_tofieldmap_bedstatus_leak]]）。

package observation

import "testing"

func TestTrack_ToFieldMap_OmitsZeroBedStatus(t *testing.T) {
	// 默认 BedStatus=0（雷达 heartbeat / radarTrackToData 路径）→ 不应出现 bed_status 键
	tr := Track{TrackID: 11, TrackConfidence: 80}
	m := tr.ToFieldMap()
	if _, ok := m[FieldBedStatus]; ok {
		t.Errorf("BedStatus=0 must be omitted; got map=%+v", m)
	}
}

func TestTrack_ToFieldMap_WritesNonZeroBedStatus(t *testing.T) {
	// sleepace LeftBed / silent_fall alarm 显式设 BedStatus=1 → 必须写出
	tr := Track{TrackID: 0, BedStatus: 1}
	m := tr.ToFieldMap()
	got, ok := m[FieldBedStatus]
	if !ok {
		t.Fatalf("BedStatus=1 must be written; got map=%+v", m)
	}
	if got != 1 {
		t.Errorf("want bed_status=1, got %v", got)
	}
}

func TestTrack_RoundTrip_BedStatusZero(t *testing.T) {
	// 0 round-trip：encode omit → decode 看不到字段 → 默认 0
	in := Track{TrackID: 11, BedStatus: 0}
	m := in.ToFieldMap()
	var out Track
	out.FromFieldMap(m)
	if out.BedStatus != 0 {
		t.Errorf("round-trip 0: want 0, got %d", out.BedStatus)
	}
}

func TestTrack_RoundTrip_BedStatusOne(t *testing.T) {
	in := Track{TrackID: 0, BedStatus: 1}
	m := in.ToFieldMap()
	var out Track
	out.FromFieldMap(m)
	if out.BedStatus != 1 {
		t.Errorf("round-trip 1: want 1, got %d", out.BedStatus)
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
	// 必写：TrackID 是顶层标识，0 也要写（与现行 ToFieldMap 一致）
	if _, ok := m[FieldTrackID]; !ok {
		t.Errorf("TrackID must always be written; got %+v", m)
	}
}
