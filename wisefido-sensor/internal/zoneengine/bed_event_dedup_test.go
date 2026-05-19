// bed_event_dedup_test.go — S5b BedEventDedup T1 单元测试（纯逻辑无 IO）。
//
// 覆盖：
//   - 首入放行
//   - 同 device 同 kind 10s 内重复 → 拒
//   - 同 device 同 kind 10s 边界（≥10s 重新放行）
//   - 同 device 不同 kind 互不影响
//   - 不同 device 互不影响
//   - 空 deviceAddr/kind 兜底放行（保持 caller 兼容）
//   - Forget 清空 device entries

package zoneengine

import (
	"testing"
)

func TestBedDedup_FirstEntryAdmits(t *testing.T) {
	d := NewBedEventDedup()
	if !d.Admit("dev-a", "enter", 1000) {
		t.Error("first entry should be admitted")
	}
}

func TestBedDedup_SameKindWithin10sRejected(t *testing.T) {
	d := NewBedEventDedup()
	d.Admit("dev-a", "enter", 1000)
	if d.Admit("dev-a", "enter", 5000) {
		t.Error("same kind within 10s should be rejected (gap=4s)")
	}
	if d.Admit("dev-a", "enter", 10_999) {
		t.Error("same kind at <10s boundary should be rejected (gap=9.999s)")
	}
}

func TestBedDedup_SameKindAtOrAfter10sAdmits(t *testing.T) {
	d := NewBedEventDedup()
	d.Admit("dev-a", "enter", 1000)
	if !d.Admit("dev-a", "enter", 11_000) {
		t.Error("same kind at 10s exact boundary should be admitted (gap=10s)")
	}
	// 再 5s 又被拒
	if d.Admit("dev-a", "enter", 16_000) {
		t.Error("after admit, next within 10s should be rejected")
	}
}

func TestBedDedup_DifferentKindsIndependent(t *testing.T) {
	d := NewBedEventDedup()
	d.Admit("dev-a", "enter", 1000)
	// leave 立刻紧跟应放行（不同 kind 互不抑制）
	if !d.Admit("dev-a", "leave", 2000) {
		t.Error("different kind should not be deduped")
	}
}

func TestBedDedup_DifferentDevicesIndependent(t *testing.T) {
	d := NewBedEventDedup()
	d.Admit("dev-a", "enter", 1000)
	if !d.Admit("dev-b", "enter", 2000) {
		t.Error("different device should not be deduped")
	}
}

func TestBedDedup_EmptyArgsAdmits(t *testing.T) {
	d := NewBedEventDedup()
	if !d.Admit("", "enter", 1000) {
		t.Error("empty deviceAddr should pass through (兼容 caller)")
	}
	if !d.Admit("dev-a", "", 1000) {
		t.Error("empty kind should pass through")
	}
}

func TestBedDedup_NilReceiverAdmits(t *testing.T) {
	var d *BedEventDedup
	if !d.Admit("dev-a", "enter", 1000) {
		t.Error("nil receiver should admit (defensive)")
	}
}

func TestBedDedup_ForgetClearsDevice(t *testing.T) {
	d := NewBedEventDedup()
	d.Admit("dev-a", "enter", 1000)
	d.Forget("dev-a")
	if !d.Admit("dev-a", "enter", 2000) {
		t.Error("after Forget, same kind in same window should re-admit (state cleared)")
	}
}

func TestBedDedup_ForgetDoesNotAffectOthers(t *testing.T) {
	d := NewBedEventDedup()
	d.Admit("dev-a", "enter", 1000)
	d.Admit("dev-b", "enter", 1000)
	d.Forget("dev-a")
	if d.Admit("dev-b", "enter", 5000) {
		t.Error("Forget(dev-a) should not affect dev-b state")
	}
}
