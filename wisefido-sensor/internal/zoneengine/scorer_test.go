package zoneengine

import (
	"testing"
)

// helper: 用默认规则建一个 small_bed scorer
func newBedScorer() *Scorer {
	rules := DefaultRules()
	return NewScorer(&rules.Bed, "small")
}

func newLargeBedScorer() *Scorer {
	rules := DefaultRules()
	return NewScorer(&rules.Bed, "large")
}

func newRoomScorer() *Scorer {
	rules := DefaultRules()
	return NewScorer(&rules.Room, "")
}

func TestScorer_EnterRaisesScore(t *testing.T) {
	s := newBedScorer()
	now := int64(1_000_000_000_000)
	score, accepted := s.Apply(SignalEvidence{
		Source: "sleepace", Kind: "enter", Ts: now,
	}, now)
	if !accepted {
		t.Fatalf("sleepace enter should be accepted")
	}
	if score != 90 {
		t.Errorf("after sleepace enter, score = %d, want 90", score)
	}
}

func TestScorer_RadarEnterLowerThanSleepace(t *testing.T) {
	s := newBedScorer()
	now := int64(1_000_000_000_000)
	score, _ := s.Apply(SignalEvidence{Source: "radar", Kind: "enter", Ts: now}, now)
	if score != 80 {
		t.Errorf("radar enter score = %d, want 80", score)
	}
}

func TestScorer_SustainAddsRecentEnterBonus(t *testing.T) {
	s := newBedScorer()
	now := int64(1_000_000_000_000)
	// enter event + 5s later sustain → max(decayedEnter, sustain) + bonus
	// 5s 衰减后 enterEff ≈ 86，sustain=80，max=86，bonus=+15 → 101
	s.Apply(SignalEvidence{Source: "sleepace", Kind: "enter", Ts: now}, now)
	score, _ := s.Apply(SignalEvidence{Source: "sleepad", Kind: "sustain", Ts: now + 5000}, now+5000)
	expectedEnter := decayStrength(90, 5000) // 5s decay
	want := expectedEnter + 15               // bonus when sustain active
	if score != want {
		t.Errorf("enter+sustain within window, score = %d, want %d (enterEff=%d + bonus 15)",
			score, want, expectedEnter)
	}
}

func TestScorer_SustainAloneNoBonus(t *testing.T) {
	s := newBedScorer()
	now := int64(1_000_000_000_000)
	// 没有 enter event 直接 sustain → 只有 sustain 80，没 bonus
	score, _ := s.Apply(SignalEvidence{Source: "radar", Kind: "sustain", Ts: now}, now)
	if score != 80 {
		t.Errorf("sustain alone score = %d, want 80", score)
	}
}

func TestScorer_LeavePushesNegative(t *testing.T) {
	s := newBedScorer()
	now := int64(1_000_000_000_000)
	// sleepace LeftBed: small_bed bucket → 80
	score, _ := s.Apply(SignalEvidence{Source: "sleepace", Kind: "leave", Ts: now}, now)
	if score != -80 {
		t.Errorf("sleepace leave score = %d, want -80", score)
	}
}

func TestScorer_LargeBedLeaveDeprioritized(t *testing.T) {
	s := newLargeBedScorer()
	now := int64(1_000_000_000_000)
	score, _ := s.Apply(SignalEvidence{Source: "sleepace", Kind: "leave", Ts: now}, now)
	if score != -70 {
		t.Errorf("large bed sleepace leave should drop to 70 (覆盖不全降权), got %d", score)
	}
}

func TestScorer_EnterLatchBlocksLeave(t *testing.T) {
	s := newBedScorer()
	now := int64(1_000_000_000_000)
	// sleepace enter (latch 10s)
	s.Apply(SignalEvidence{Source: "sleepace", Kind: "enter", Ts: now}, now)
	// 5s 后 LeftBed 试图翻转 → 应被 enter latch 拒绝
	score, accepted := s.Apply(SignalEvidence{Source: "sleepace", Kind: "leave", Ts: now + 5000}, now+5000)
	if accepted {
		t.Errorf("leave within enter latch should be rejected")
	}
	if score < 50 {
		t.Errorf("score should remain occupied during latch, got %d", score)
	}
}

func TestScorer_LeaveLatchBlocksEnter(t *testing.T) {
	s := newBedScorer()
	now := int64(1_000_000_000_000)
	s.Apply(SignalEvidence{Source: "sleepace", Kind: "leave", Ts: now}, now)
	// 1s 后 InBed 试图翻转 → 被 leave latch (2s small_bed) 拒绝
	_, accepted := s.Apply(SignalEvidence{Source: "sleepace", Kind: "enter", Ts: now + 1000}, now+1000)
	if accepted {
		t.Errorf("enter within leave latch (2s) should be rejected at t+1s")
	}
}

func TestScorer_SameDirectionTakesMax(t *testing.T) {
	s := newBedScorer()
	now := int64(1_000_000_000_000)
	// radar enter 80 → 然后 sleepace enter 90 → score 应保持 90（取最强非 sum）
	s.Apply(SignalEvidence{Source: "radar", Kind: "enter", Ts: now}, now)
	score, _ := s.Apply(SignalEvidence{Source: "sleepace", Kind: "enter", Ts: now + 100}, now+100)
	if score != 90 {
		t.Errorf("max over same direction, score = %d, want 90 (not 170)", score)
	}
}

func TestScorer_DecaysOverTime(t *testing.T) {
	s := newBedScorer()
	now := int64(1_000_000_000_000)
	s.Apply(SignalEvidence{Source: "sleepace", Kind: "enter", Ts: now}, now)

	// 半衰减点：60s → 应约为 45
	mid := s.Score(now + 60*1000)
	if mid < 40 || mid > 50 {
		t.Errorf("at 60s decay, score = %d, want ~45 (40-50)", mid)
	}

	// 完全衰减：120s → 0
	full := s.Score(now + 120*1000)
	if full != 0 {
		t.Errorf("at 120s decay, score = %d, want 0", full)
	}
}

func TestScorer_RoomEnterLeaveBasic(t *testing.T) {
	s := newRoomScorer()
	now := int64(1_000_000_000_000)
	// Room enter 不依赖 bed_size_dependent
	score, _ := s.Apply(SignalEvidence{Source: "radar", Kind: "enter", Ts: now}, now)
	if score != 80 {
		t.Errorf("room radar enter = %d, want 80", score)
	}
	// radar leave (80 strength)。
	// 注意：leave event 会清零 enterStrength（互斥语义）→ pos=0, neg=80, score=-80
	score, _ = s.Apply(SignalEvidence{Source: "radar", Kind: "leave", Ts: now + 5_000}, now+5_000)
	if score != -80 {
		t.Errorf("after radar enter + 5s + radar leave, score = %d, want -80 (enter strength cleared by leave)", score)
	}
}

// TestScorer_NumberPeopleZeroNotConfiguredAsLeave 确认 number_people_0 不再作为 leave 配置。
// 若未来有 adapter 误用 Kind="leave" + Source="number_people_0" 发送，scorer 应当 reject（lookup 返回 0 → accepted=false）。
// 这是为了防止"静止人 / 雷达盲区"被误判为"人离开"。
func TestScorer_NumberPeopleZeroNotConfiguredAsLeave(t *testing.T) {
	s := newRoomScorer()
	now := int64(1_000_000_000_000)
	score, accepted := s.Apply(SignalEvidence{Source: "number_people_0", Kind: "leave", Ts: now}, now)
	if accepted {
		t.Errorf("number_people_0 as leave should be rejected (not configured) —防止 count=0 被误判为离开")
	}
	if score != 0 {
		t.Errorf("rejected leave should not change score, got %d", score)
	}
}
