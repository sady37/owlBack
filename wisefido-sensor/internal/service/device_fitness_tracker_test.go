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
}
