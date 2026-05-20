// sensor_state_projector_test.go — S5b mergeBedStateSensorOwner / mergeBedStateSleepStage T1 单元测试。
//
// 覆盖：
//   - bed.state 路径：sensor owner 字段（5 核 + TrackNumber + BedConfidence）从 incoming 取，
//     SleepStage/SleepConfidence 保留 prev
//   - bed.sleepstage 路径：仅更 SleepStage/SleepConfidence/UpdatedAt，bed.state 字段保留 prev

package consumer

import (
	"testing"

	"owl-common/card"
)

func TestMergeBedStateSensorOwner_PreservesSleepStage(t *testing.T) {
	prev := &card.BedState{
		BedStatus:       0,
		SleepStage:      2, // light
		SleepConfidence: 90,
		TrackNumber:     1,
	}
	incoming := &card.BedState{
		UpdatedAt:     1_700_000_000_000,
		BedStatus:     1,
		BedEvent:      1,
		StartTime:     1_700_000_000_000,
		DurationSec:   120,
		TrackNumber:   0, // S5b: incoming 写
		BedConfidence: 60, // S5b: incoming 写
	}
	out := mergeBedStateSensorOwner(prev, incoming)

	if out.BedStatus != 1 || out.BedEvent != 1 || out.DurationSec != 120 {
		t.Errorf("core fields not overwritten: %+v", out)
	}
	if out.TrackNumber != 0 {
		t.Errorf("TrackNumber should take incoming 0, got %d", out.TrackNumber)
	}
	if out.BedConfidence != 60 {
		t.Errorf("BedConfidence should take incoming 60, got %d", out.BedConfidence)
	}
	// SleepStage / SleepConfidence 保留 prev（bed.state 路径不动）
	if out.SleepStage != 2 || out.SleepConfidence != 90 {
		t.Errorf("SleepStage/Conf should be preserved (2/90), got %d/%d",
			out.SleepStage, out.SleepConfidence)
	}
}

func TestMergeBedStateSensorOwner_NilPrev(t *testing.T) {
	incoming := &card.BedState{
		UpdatedAt:     1_700_000_000_000,
		BedStatus:     0,
		TrackNumber:   1,
		BedConfidence: 90,
	}
	out := mergeBedStateSensorOwner(nil, incoming)
	if out.TrackNumber != 1 || out.BedConfidence != 90 {
		t.Errorf("nil prev should pass-through: %+v", out)
	}
}

func TestMergeBedStateSensorOwner_NilIncoming(t *testing.T) {
	prev := &card.BedState{BedStatus: 0, SleepStage: 2}
	out := mergeBedStateSensorOwner(prev, nil)
	if out != prev {
		t.Errorf("nil incoming should return prev unchanged")
	}
}

func TestMergeBedStateSleepStage_PreservesBedStatus(t *testing.T) {
	prev := &card.BedState{
		BedStatus:       0,
		BedEvent:        0,
		StartTime:       1_700_000_000_000,
		DurationSec:     30,
		TrackNumber:     1,
		BedConfidence:   90,
		SleepStage:      0,
		SleepConfidence: 0,
	}
	incoming := &card.BedState{
		UpdatedAt:       1_700_000_060_000,
		SleepStage:      2,
		SleepConfidence: 90,
	}
	out := mergeBedStateSleepStage(prev, incoming)

	if out.SleepStage != 2 || out.SleepConfidence != 90 {
		t.Errorf("SleepStage path should update: stage=%d conf=%d", out.SleepStage, out.SleepConfidence)
	}
	if out.UpdatedAt != 1_700_000_060_000 {
		t.Errorf("UpdatedAt should advance: %d", out.UpdatedAt)
	}
	// bed.state owner 字段全保留 prev
	if out.BedStatus != 0 || out.BedEvent != 0 || out.DurationSec != 30 {
		t.Errorf("bed.state fields should NOT be touched: %+v", out)
	}
	if out.TrackNumber != 1 || out.BedConfidence != 90 {
		t.Errorf("TrackNumber/BedConfidence should NOT be touched: %d / %d",
			out.TrackNumber, out.BedConfidence)
	}
}

// 防回归：bed.state 在 T=100 写后，bed.sleepstage 带较早采样时间 T=95 到达，
// UpdatedAt 必须保留 T=100（max 语义），不能被 sleepstage event 时间倒推。
// 否则 FE 用 UpdatedAt 算 bed transition anchor 会拿到 stale 时间。
func TestMergeBedStateSleepStage_DoesNotRegressUpdatedAt(t *testing.T) {
	prev := &card.BedState{
		UpdatedAt:     1_700_000_100_000, // bed.state 刚写
		BedStatus:     1,
		BedEvent:      1,
		StartTime:     1_700_000_100_000,
		DurationSec:   0,
		TrackNumber:   1,
		BedConfidence: 80,
	}
	incoming := &card.BedState{
		UpdatedAt:       1_700_000_095_000, // sleepstage 采样时间早 5s
		SleepStage:      3,
		SleepConfidence: 90,
	}
	out := mergeBedStateSleepStage(prev, incoming)

	if out.UpdatedAt != 1_700_000_100_000 {
		t.Errorf("UpdatedAt must not regress: got %d, want %d", out.UpdatedAt, 1_700_000_100_000)
	}
	if out.SleepStage != 3 || out.SleepConfidence != 90 {
		t.Errorf("SleepStage path should still update: stage=%d conf=%d", out.SleepStage, out.SleepConfidence)
	}
}
