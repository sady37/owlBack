// device_fitness_tracker_test.go — S6 DeviceFitnessTracker T1 单元测试。
//
// 覆盖：
//   - 首次空 = fit；MarkUnfit 后 unfit；ClearReason 后 fit
//   - 多 reason OR 累积；逐项 Clear 才回 fit
//   - 同 reason 重复 Mark 幂等
//   - 多 device 互不污染
//   - 空参 / nil receiver 兜底
//   - Forget 清空 device

package service

import (
	"testing"

	"go.uber.org/zap"
)

func newTestFitnessTracker() *DeviceFitnessTracker {
	return NewDeviceFitnessTracker(zap.NewNop())
}

func TestFitness_EmptyIsFit(t *testing.T) {
	tr := newTestFitnessTracker()
	if !tr.IsFit("dev-a") {
		t.Error("empty tracker should report fit")
	}
}

func TestFitness_MarkUnfitThenIsFitFalse(t *testing.T) {
	tr := newTestFitnessTracker()
	tr.MarkUnfit("dev-a", FitnessReasonOffline)
	if tr.IsFit("dev-a") {
		t.Error("after MarkUnfit should report not fit")
	}
}

func TestFitness_ClearReasonRestoresFit(t *testing.T) {
	tr := newTestFitnessTracker()
	tr.MarkUnfit("dev-a", FitnessReasonOffline)
	tr.ClearReason("dev-a", FitnessReasonOffline)
	if !tr.IsFit("dev-a") {
		t.Error("after Clear should be fit again")
	}
}

func TestFitness_MultipleReasonsAccumulate(t *testing.T) {
	tr := newTestFitnessTracker()
	tr.MarkUnfit("dev-a", FitnessReasonOffline)
	tr.MarkUnfit("dev-a", FitnessReasonSensorDetached)
	if tr.IsFit("dev-a") {
		t.Error("two reasons → still unfit")
	}
	// 只清 Offline，仍 unfit
	tr.ClearReason("dev-a", FitnessReasonOffline)
	if tr.IsFit("dev-a") {
		t.Error("only one reason cleared → still unfit (SensorDetached remains)")
	}
	// 再清 SensorDetached → fit
	tr.ClearReason("dev-a", FitnessReasonSensorDetached)
	if !tr.IsFit("dev-a") {
		t.Error("all reasons cleared → fit")
	}
}

func TestFitness_DuplicateMarkIdempotent(t *testing.T) {
	tr := newTestFitnessTracker()
	tr.MarkUnfit("dev-a", FitnessReasonSignalPoor)
	tr.MarkUnfit("dev-a", FitnessReasonSignalPoor)
	tr.MarkUnfit("dev-a", FitnessReasonSignalPoor)
	tr.ClearReason("dev-a", FitnessReasonSignalPoor)
	if !tr.IsFit("dev-a") {
		t.Error("duplicate Mark + single Clear should restore fit")
	}
}

func TestFitness_DevicesIndependent(t *testing.T) {
	tr := newTestFitnessTracker()
	tr.MarkUnfit("dev-a", FitnessReasonAngleException)
	if !tr.IsFit("dev-b") {
		t.Error("dev-b should not be affected by dev-a unfit")
	}
}

func TestFitness_ClearReasonNonExistentNoOp(t *testing.T) {
	tr := newTestFitnessTracker()
	tr.ClearReason("dev-a", FitnessReasonOffline) // never marked
	if !tr.IsFit("dev-a") {
		t.Error("Clear without prior Mark should be no-op (fit)")
	}
}

func TestFitness_ReasonsBitmask(t *testing.T) {
	tr := newTestFitnessTracker()
	tr.MarkUnfit("dev-a", FitnessReasonOffline)
	tr.MarkUnfit("dev-a", FitnessReasonAngleException)
	got := tr.Reasons("dev-a")
	want := FitnessReasonOffline | FitnessReasonAngleException
	if got != want {
		t.Errorf("Reasons = %b, want %b", got, want)
	}
}

func TestFitness_Forget(t *testing.T) {
	tr := newTestFitnessTracker()
	tr.MarkUnfit("dev-a", FitnessReasonOffline)
	tr.MarkUnfit("dev-b", FitnessReasonOffline)
	tr.Forget("dev-a")
	if !tr.IsFit("dev-a") {
		t.Error("after Forget dev-a should be fit")
	}
	if tr.IsFit("dev-b") {
		t.Error("Forget(dev-a) should not affect dev-b")
	}
}

func TestFitness_EmptyArgsNoCrash(t *testing.T) {
	tr := newTestFitnessTracker()
	tr.MarkUnfit("", FitnessReasonOffline)   // empty addr - no-op
	tr.ClearReason("", FitnessReasonOffline) // empty addr - no-op
	tr.MarkUnfit("dev-a", 0)                 // zero reason - no-op
	if !tr.IsFit("dev-a") {
		t.Error("empty args should be no-op")
	}
	if !tr.IsFit("") {
		t.Error("empty deviceAddr should report fit (defensive default)")
	}
}

func TestFitness_NilReceiver(t *testing.T) {
	var tr *DeviceFitnessTracker
	if !tr.IsFit("dev-a") {
		t.Error("nil receiver should report fit (defensive)")
	}
	// 下面不应 panic
	tr.MarkUnfit("dev-a", FitnessReasonOffline)
	tr.ClearReason("dev-a", FitnessReasonOffline)
	tr.Forget("dev-a")
	tr.RegisterUnfitCallback(func(string) {}) // 防 panic
}

// ---------------------------------------------------------------------------
// UnfitCallback broadcast (fit→unfit edge)
// ---------------------------------------------------------------------------

func TestFitness_FirstEdgeBroadcastsCallback(t *testing.T) {
	tr := newTestFitnessTracker()
	var called []string
	tr.RegisterUnfitCallback(func(addr string) { called = append(called, addr) })

	tr.MarkUnfit("dev-a", FitnessReasonOffline)

	if len(called) != 1 || called[0] != "dev-a" {
		t.Errorf("first edge should broadcast: got %v", called)
	}
}

func TestFitness_AdditionalReasonNoBroadcast(t *testing.T) {
	tr := newTestFitnessTracker()
	var count int
	tr.RegisterUnfitCallback(func(string) { count++ })

	tr.MarkUnfit("dev-a", FitnessReasonOffline)        // first edge → 1
	tr.MarkUnfit("dev-a", FitnessReasonSensorDetached) // 累加 reason，state 已清，不再 broadcast
	tr.MarkUnfit("dev-a", FitnessReasonAngleException)

	if count != 1 {
		t.Errorf("only first edge broadcasts; got %d", count)
	}
}

func TestFitness_DuplicateMarkNoBroadcast(t *testing.T) {
	tr := newTestFitnessTracker()
	var count int
	tr.RegisterUnfitCallback(func(string) { count++ })

	tr.MarkUnfit("dev-a", FitnessReasonOffline)
	tr.MarkUnfit("dev-a", FitnessReasonOffline) // 重复 mark 同 reason

	if count != 1 {
		t.Errorf("duplicate mark same reason: 1 broadcast, got %d", count)
	}
}

func TestFitness_ReFitThenUnfitBroadcastsAgain(t *testing.T) {
	tr := newTestFitnessTracker()
	var count int
	tr.RegisterUnfitCallback(func(string) { count++ })

	tr.MarkUnfit("dev-a", FitnessReasonOffline) // 1
	tr.ClearReason("dev-a", FitnessReasonOffline)
	if !tr.IsFit("dev-a") {
		t.Fatal("setup: should be fit after Clear")
	}
	tr.MarkUnfit("dev-a", FitnessReasonOffline) // 又一次 fit→unfit edge → 2

	if count != 2 {
		t.Errorf("after recover + unfit again, should broadcast twice; got %d", count)
	}
}

func TestFitness_MultiCallbacksAllCalled(t *testing.T) {
	tr := newTestFitnessTracker()
	var got1, got2, got3 string
	tr.RegisterUnfitCallback(func(a string) { got1 = a })
	tr.RegisterUnfitCallback(func(a string) { got2 = a })
	tr.RegisterUnfitCallback(func(a string) { got3 = a })

	tr.MarkUnfit("dev-a", FitnessReasonSensorDetached)

	if got1 != "dev-a" || got2 != "dev-a" || got3 != "dev-a" {
		t.Errorf("3 callbacks all should fire: %q / %q / %q", got1, got2, got3)
	}
}

func TestFitness_NilCallbackIgnored(t *testing.T) {
	tr := newTestFitnessTracker()
	tr.RegisterUnfitCallback(nil) // 不应 panic 也不入队
	// 后续 MarkUnfit 不 panic
	tr.MarkUnfit("dev-a", FitnessReasonOffline)
}

func TestFitness_DifferentDevicesEachBroadcast(t *testing.T) {
	tr := newTestFitnessTracker()
	var called []string
	tr.RegisterUnfitCallback(func(a string) { called = append(called, a) })

	tr.MarkUnfit("dev-a", FitnessReasonOffline)
	tr.MarkUnfit("dev-b", FitnessReasonOffline)
	tr.MarkUnfit("dev-c", FitnessReasonOffline)

	if len(called) != 3 {
		t.Errorf("3 different devices: 3 broadcasts, got %d (%v)", len(called), called)
	}
}
