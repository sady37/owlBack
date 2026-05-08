package spatial

import (
	"errors"
	"net/netip"
	"testing"
)

// =============================================================================
// HoA build / parse 闭环
// =============================================================================

func TestBuildResidentHoA(t *testing.T) {
	hoA := BuildResidentHoA(0x0001, 0x0042)
	want := netip.MustParseAddr("fd00:0:1:ff01:42::")
	if hoA != want {
		t.Errorf("BuildResidentHoA = %s, want %s", hoA, want)
	}

	tenant, kind, id, err := ParseSubjectAddr(hoA)
	if err != nil {
		t.Fatalf("ParseSubjectAddr: %v", err)
	}
	if tenant != 0x0001 || kind != SubjectKindResident || id != 0x0042 {
		t.Errorf("parsed: tenant=%#x kind=%v id=%#x; want 1/resident/0x42", tenant, kind, id)
	}
}

func TestBuildSubjectHoA_Caregiver(t *testing.T) {
	// caregiver 走通用 builder（没有便利封装）
	hoA, err := BuildSubjectHoA(0x0001, SubjectKindCaregiver, 0x07)
	if err != nil {
		t.Fatalf("BuildSubjectHoA caregiver: %v", err)
	}
	want := netip.MustParseAddr("fd00:0:1:ff02:7::")
	if hoA != want {
		t.Errorf("caregiver HoA = %s, want %s", hoA, want)
	}

	tenant, kind, id, _ := ParseSubjectAddr(hoA)
	if tenant != 0x0001 || kind != SubjectKindCaregiver || id != 0x07 {
		t.Errorf("parsed: tenant=%#x kind=%v id=%#x", tenant, kind, id)
	}
}

func TestBuildSubjectHoA_MaxID(t *testing.T) {
	// 16-bit subject_id 边界
	hoA, err := BuildSubjectHoA(0xABCD, SubjectKindResident, 0xFFFE)
	if err != nil {
		t.Fatalf("BuildSubjectHoA: %v", err)
	}
	tenant, kind, id, _ := ParseSubjectAddr(hoA)
	if tenant != 0xABCD || kind != SubjectKindResident || id != 0xFFFE {
		t.Errorf("roundtrip lost data: tenant=%#x kind=%v id=%#x", tenant, kind, id)
	}
}

func TestBuildSubjectHoA_InvalidKind(t *testing.T) {
	if _, err := BuildSubjectHoA(1, SubjectKindUnknown, 1); !errors.Is(err, ErrInvalidSubjectKind) {
		t.Errorf("kind=Unknown(0): err = %v, want ErrInvalidSubjectKind", err)
	}
	if _, err := BuildSubjectHoA(1, SubjectKind(0xFF), 1); !errors.Is(err, ErrInvalidSubjectKind) {
		t.Errorf("kind=0xFF: err = %v, want ErrInvalidSubjectKind", err)
	}
}

func TestBuildSubjectHoAFromHex(t *testing.T) {
	cases := []struct {
		idHex string
		want  uint16
	}{
		{"42", 0x0042},
		{"0042", 0x0042},
		{"abcd", 0xABCD},
		{"", 0},
	}
	for _, c := range cases {
		hoA, err := BuildSubjectHoAFromHex(1, SubjectKindResident, c.idHex)
		if err != nil {
			t.Errorf("hex=%q: %v", c.idHex, err)
			continue
		}
		_, _, id, _ := ParseSubjectAddr(hoA)
		if id != c.want {
			t.Errorf("hex=%q: parsed id = %#x, want %#x", c.idHex, id, c.want)
		}
	}

	// 超长 hex
	if _, err := BuildSubjectHoAFromHex(1, SubjectKindResident, "12345"); !errors.Is(err, ErrSubjectAddrInvalid) {
		t.Errorf("expected ErrSubjectAddrInvalid for 5-char hex, got %v", err)
	}
	// 非 hex
	if _, err := BuildSubjectHoAFromHex(1, SubjectKindResident, "ZZZ"); !errors.Is(err, ErrSubjectAddrInvalid) {
		t.Errorf("expected ErrSubjectAddrInvalid for non-hex, got %v", err)
	}
}

// =============================================================================
// Subject namespace prefix 查询语义
// =============================================================================

func TestSubjectNamespacePrefix(t *testing.T) {
	ns := BuildSubjectNamespacePrefix(0x0001)
	if got := ns.String(); got != "fd00:0:1:ff00::/56" {
		t.Errorf("namespace prefix = %s, want fd00:0:1:ff00::/56", got)
	}

	// 含所有同 tenant 的 subject（resident + caregiver）
	resident := BuildResidentHoA(1, 0x42)
	caregiver, _ := BuildSubjectHoA(1, SubjectKindCaregiver, 0x07)
	if !ns.Contains(resident) {
		t.Errorf("namespace should contain resident HoA")
	}
	if !ns.Contains(caregiver) {
		t.Errorf("namespace should contain caregiver HoA")
	}

	// 跨 tenant 不应包含
	otherTenantResident := BuildResidentHoA(2, 0x42)
	if ns.Contains(otherTenantResident) {
		t.Errorf("tenant=1 namespace should NOT contain tenant=2 resident")
	}
}

func TestSubjectKindPrefix(t *testing.T) {
	residents, err := BuildSubjectKindPrefix(0x0001, SubjectKindResident)
	if err != nil {
		t.Fatalf("BuildSubjectKindPrefix: %v", err)
	}
	if got := residents.String(); got != "fd00:0:1:ff01::/64" {
		t.Errorf("residents prefix = %s, want fd00:0:1:ff01::/64", got)
	}

	caregivers, _ := BuildSubjectKindPrefix(0x0001, SubjectKindCaregiver)
	if got := caregivers.String(); got != "fd00:0:1:ff02::/64" {
		t.Errorf("caregivers prefix = %s, want fd00:0:1:ff02::/64", got)
	}

	// 包含本类型，不含其他类型
	r1 := BuildResidentHoA(1, 1)
	c1, _ := BuildSubjectHoA(1, SubjectKindCaregiver, 1)
	if !residents.Contains(r1) {
		t.Errorf("residents prefix should contain resident HoA")
	}
	if residents.Contains(c1) {
		t.Errorf("residents prefix should NOT contain caregiver HoA")
	}
	if !caregivers.Contains(c1) {
		t.Errorf("caregivers prefix should contain caregiver HoA")
	}
}

// =============================================================================
// IsSubjectAddr / IsSpatialAddr 二分
// =============================================================================

func TestSubject_vs_Spatial_Disjoint(t *testing.T) {
	// 同 tenant 的 spatial device address
	tenant := BuildTenantPrefix(0x0001)
	branch, _ := DeriveBranchPrefix(tenant, 0x01)
	site, _ := DeriveSitePrefix(branch, 1, 2)
	unit, _ := DeriveUnitPrefix(site, 0x42)
	room, _ := DeriveRoomPrefix(unit, 0x03)
	bed, _ := DeriveBedPrefix(room, 0x01)
	dev, _ := DeriveDeviceAddr(bed, "E598A2ACD523")

	// 同 tenant 的 subject HoA
	resident := BuildResidentHoA(0x0001, 0x42)

	if !IsOwlAddr(dev) || !IsOwlAddr(resident) {
		t.Errorf("both should be IsOwlAddr")
	}

	if !IsSpatialAddr(dev) || IsSubjectAddr(dev) {
		t.Errorf("device addr classification wrong: spatial=%v subject=%v",
			IsSpatialAddr(dev), IsSubjectAddr(dev))
	}
	if IsSpatialAddr(resident) || !IsSubjectAddr(resident) {
		t.Errorf("resident HoA classification wrong: spatial=%v subject=%v",
			IsSpatialAddr(resident), IsSubjectAddr(resident))
	}

	// 同 tenant /48 prefix 包含两者（HIPAA 隔离统一边界）
	if !tenant.Contains(dev) || !tenant.Contains(resident) {
		t.Errorf("tenant /48 should contain both spatial device and subject HoA")
	}

	// spatial branch /56 与 subject namespace /56 互不包含
	subjNS := BuildSubjectNamespacePrefix(0x0001)
	if branch.Contains(resident) {
		t.Errorf("spatial branch should NOT contain resident HoA")
	}
	if subjNS.Contains(dev) {
		t.Errorf("subject namespace should NOT contain device addr")
	}
}

func TestParseSubjectAddr_NotSubject(t *testing.T) {
	tenant := BuildTenantPrefix(1)
	branch, _ := DeriveBranchPrefix(tenant, 1)
	dev, _ := DeriveDeviceAddr(branch, "A2ACD523")
	if _, _, _, err := ParseSubjectAddr(dev); !errors.Is(err, ErrNotSubjectAddr) {
		t.Errorf("expected ErrNotSubjectAddr for device addr, got %v", err)
	}

	// 非 owl namespace
	other := netip.MustParseAddr("fd00:1234::1")
	if _, _, _, err := ParseSubjectAddr(other); !errors.Is(err, ErrNotOwlNamespace) {
		t.Errorf("expected ErrNotOwlNamespace, got %v", err)
	}
}

// =============================================================================
// SubjectKindOf
// =============================================================================

func TestSubjectKindOf(t *testing.T) {
	r := BuildResidentHoA(1, 1)
	c, _ := BuildSubjectHoA(1, SubjectKindCaregiver, 1)

	if SubjectKindOf(r) != SubjectKindResident {
		t.Errorf("SubjectKindOf(resident HoA) = %v, want resident", SubjectKindOf(r))
	}
	if SubjectKindOf(c) != SubjectKindCaregiver {
		t.Errorf("SubjectKindOf(caregiver HoA) = %v, want caregiver", SubjectKindOf(c))
	}

	// 非 subject addr → Unknown
	dev, _ := DeriveDeviceAddr(BuildTenantPrefix(1), "A2ACD523")
	if SubjectKindOf(dev) != SubjectKindUnknown {
		t.Errorf("SubjectKindOf(device) = %v, want Unknown", SubjectKindOf(dev))
	}
}

// =============================================================================
// SubjectKind String
// =============================================================================

func TestSubjectKind_String(t *testing.T) {
	cases := []struct {
		kind SubjectKind
		want string
	}{
		{SubjectKindResident, "resident"},
		{SubjectKindCaregiver, "caregiver"},
		{SubjectKindUnknown, "unknown"},
		{SubjectKind(0x99), "unknown"},
	}
	for _, c := range cases {
		if got := c.kind.String(); got != c.want {
			t.Errorf("SubjectKind(%#x).String() = %q, want %q", c.kind, got, c.want)
		}
	}
}
