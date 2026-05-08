package spatial

import (
	"errors"
	"net/netip"
	"testing"
)

// =============================================================================
// Build / Derive 链路
// =============================================================================

func TestBuildTenantPrefix(t *testing.T) {
	got := BuildTenantPrefix(0x0001)
	want := netip.MustParsePrefix("fd00:0:1::/48")
	if got != want {
		t.Errorf("BuildTenantPrefix(1) = %v, want %v", got, want)
	}
	if ScopeOf(got) != Scope6Tenant {
		t.Errorf("ScopeOf = %v, want tenant", ScopeOf(got))
	}
}

func TestBuildTenantPrefix_AllSlotsLegal(t *testing.T) {
	// 全范围 0..0xFFFF 都合法（无 sentinel 拒分配）
	for _, slot := range []uint16{0x0000, 0x0001, 0x00FF, 0xFFFE, 0xFFFF} {
		got := BuildTenantPrefix(slot)
		if !got.IsValid() {
			t.Errorf("BuildTenantPrefix(%#x) returned invalid prefix", slot)
		}
		if ScopeOf(got) != Scope6Tenant {
			t.Errorf("BuildTenantPrefix(%#x) wrong scope", slot)
		}
	}
}

// TestDeriveChain 完整复现 doc §2.2 的例：
//
//	tenant=1 branch=1 building=1 floor=2 unit=0x42 room=3 bed=1
//	device_uid=E598A2ACD523 → fd00:0:1:112:42:301:a2ac:d523/128
func TestDeriveChain(t *testing.T) {
	tenant := BuildTenantPrefix(0x0001)
	if got := tenant.String(); got != "fd00:0:1::/48" {
		t.Errorf("tenant prefix = %s, want fd00:0:1::/48", got)
	}

	branch, err := DeriveBranchPrefix(tenant, 0x01)
	if err != nil {
		t.Fatalf("DeriveBranchPrefix: %v", err)
	}
	if got := branch.String(); got != "fd00:0:1:100::/56" {
		t.Errorf("branch prefix = %s, want fd00:0:1:100::/56", got)
	}

	site, err := DeriveSitePrefix(branch, 1, 2) // building 1, floor 2
	if err != nil {
		t.Fatalf("DeriveSitePrefix: %v", err)
	}
	if got := site.String(); got != "fd00:0:1:112::/64" {
		t.Errorf("site prefix = %s, want fd00:0:1:112::/64", got)
	}

	unit, err := DeriveUnitPrefix(site, 0x0042)
	if err != nil {
		t.Fatalf("DeriveUnitPrefix: %v", err)
	}
	if got := unit.String(); got != "fd00:0:1:112:42::/80" {
		t.Errorf("unit prefix = %s, want fd00:0:1:112:42::/80", got)
	}

	room, err := DeriveRoomPrefix(unit, 0x03)
	if err != nil {
		t.Fatalf("DeriveRoomPrefix: %v", err)
	}
	if got := room.String(); got != "fd00:0:1:112:42:300::/88" {
		t.Errorf("room prefix = %s, want fd00:0:1:112:42:300::/88", got)
	}

	bed, err := DeriveBedPrefix(room, 0x01)
	if err != nil {
		t.Fatalf("DeriveBedPrefix: %v", err)
	}
	if got := bed.String(); got != "fd00:0:1:112:42:301::/96" {
		t.Errorf("bed prefix = %s, want fd00:0:1:112:42:301::/96", got)
	}

	dev, err := DeriveDeviceAddr(bed, "E598A2ACD523")
	if err != nil {
		t.Fatalf("DeriveDeviceAddr: %v", err)
	}
	want := "fd00:0:1:112:42:301:a2ac:d523"
	if got := dev.String(); got != want {
		t.Errorf("device addr = %s, want %s", got, want)
	}

	// 整链 contains 关系（== "device 在 room/unit/branch X 内吗"
	// 全靠 prefix.Contains() 标准 IPv6 语义，零自定义代码）
	for _, p := range []netip.Prefix{tenant, branch, site, unit, room, bed} {
		if !p.Contains(dev) {
			t.Errorf("%s does not contain device %s", p, dev)
		}
	}
}

// =============================================================================
// PackSiteSlot
// =============================================================================

func TestPackUnpackSiteSlot(t *testing.T) {
	cases := []struct {
		building uint8
		floor    uint8
		want     uint8
	}{
		{0, 0, 0x00},
		{1, 1, 0x11},
		{14, 14, 0xEE},
		{1, 2, 0x12},
		{3, 7, 0x37},
		{15, 15, 0xFF}, // 全 0xF 现在合法（无 sentinel）
	}
	for _, c := range cases {
		slot := PackSiteSlot(c.building, c.floor)
		if slot != c.want {
			t.Errorf("PackSiteSlot(%#x,%#x) = %#x, want %#x", c.building, c.floor, slot, c.want)
		}
		gotB, gotF := UnpackSiteSlot(slot)
		if gotB != c.building || gotF != c.floor {
			t.Errorf("Unpack: got (%#x,%#x), want (%#x,%#x)", gotB, gotF, c.building, c.floor)
		}
	}
	// 超界值被截断到 4 bit
	if got := PackSiteSlot(0xFF, 0xFF); got != 0xFF {
		t.Errorf("PackSiteSlot(255,255) = %#x, want 0xFF (clipped to 4+4 bit)", got)
	}
}

// =============================================================================
// 错误路径
// =============================================================================

func TestDerive_WrongPrefixLen(t *testing.T) {
	tenant := BuildTenantPrefix(1)
	branch, _ := DeriveBranchPrefix(tenant, 1)
	if _, err := DeriveUnitPrefix(branch, 1); !errors.Is(err, ErrInvalidPrefixLen) {
		t.Errorf("expected ErrInvalidPrefixLen, got %v", err)
	}
	if _, err := DeriveRoomPrefix(tenant, 1); !errors.Is(err, ErrInvalidPrefixLen) {
		t.Errorf("expected ErrInvalidPrefixLen for tenant→room, got %v", err)
	}
}

func TestDeriveDeviceAddr_BadUID(t *testing.T) {
	bed := buildTestBed(t)

	// 短 UID
	if _, err := DeriveDeviceAddr(bed, "ABC"); !errors.Is(err, ErrInvalidDeviceUID) {
		t.Errorf("expected ErrInvalidDeviceUID for short UID, got %v", err)
	}
	// 非 hex
	if _, err := DeriveDeviceAddr(bed, "ZZZZZZZZ"); !errors.Is(err, ErrInvalidDeviceUID) {
		t.Errorf("expected ErrInvalidDeviceUID for non-hex UID, got %v", err)
	}
	// /128 base 不允许
	bigPrefix := netip.PrefixFrom(bed.Addr(), 128)
	if _, err := DeriveDeviceAddr(bigPrefix, "A2ACD523"); !errors.Is(err, ErrInvalidPrefixLen) {
		t.Errorf("expected ErrInvalidPrefixLen for /128 base, got %v", err)
	}
}

// buildTestBed 一次走完 tenant→bed 链；任一步失败 t.Fatal。
func buildTestBed(t *testing.T) netip.Prefix {
	t.Helper()
	tenant := BuildTenantPrefix(1)
	branch, err := DeriveBranchPrefix(tenant, 1)
	if err != nil {
		t.Fatalf("DeriveBranchPrefix: %v", err)
	}
	site, err := DeriveSitePrefix(branch, 1, 1)
	if err != nil {
		t.Fatalf("DeriveSitePrefix: %v", err)
	}
	unit, err := DeriveUnitPrefix(site, 1)
	if err != nil {
		t.Fatalf("DeriveUnitPrefix: %v", err)
	}
	room, err := DeriveRoomPrefix(unit, 1)
	if err != nil {
		t.Fatalf("DeriveRoomPrefix: %v", err)
	}
	bed, err := DeriveBedPrefix(room, 1)
	if err != nil {
		t.Fatalf("DeriveBedPrefix: %v", err)
	}
	return bed
}

// =============================================================================
// IsOwlAddr / Containment / LPM
// =============================================================================

func TestIsOwlAddr(t *testing.T) {
	tenant := BuildTenantPrefix(1)
	dev, _ := DeriveDeviceAddr(tenant, "A2ACD523")
	if !IsOwlAddr(dev) {
		t.Errorf("IsOwlAddr(%s) = false, want true", dev)
	}
	// 其他 ULA namespace（fd00:1234::）— 不在 owl namespace
	other := netip.MustParseAddr("fd00:1234::1")
	if IsOwlAddr(other) {
		t.Errorf("IsOwlAddr(%s) = true, want false (non-owl ULA)", other)
	}
	// IPv4 — 不是 IPv6
	v4 := netip.MustParseAddr("10.0.0.1")
	if IsOwlAddr(v4) {
		t.Errorf("IsOwlAddr(IPv4) = true, want false")
	}
}

func TestLongestPrefixMatch(t *testing.T) {
	tenant := BuildTenantPrefix(1)
	branch, _ := DeriveBranchPrefix(tenant, 1)
	site, _ := DeriveSitePrefix(branch, 1, 2)
	unit, _ := DeriveUnitPrefix(site, 0x42)
	room, _ := DeriveRoomPrefix(unit, 3)
	bed, _ := DeriveBedPrefix(room, 1)
	dev, _ := DeriveDeviceAddr(bed, "A2ACD523")

	// 多 candidate：bed 应胜（最长）
	candidates := []netip.Prefix{tenant, branch, site, unit, room, bed}
	got := LongestPrefixMatch(dev, candidates)
	if got != bed {
		t.Errorf("LPM = %v, want bed %v", got, bed)
	}

	// 移除 bed → room 应胜
	got2 := LongestPrefixMatch(dev, []netip.Prefix{tenant, branch, site, unit, room})
	if got2 != room {
		t.Errorf("LPM (no bed) = %v, want room %v", got2, room)
	}

	// 全不匹配（另一个 tenant 的 prefix）
	otherTenant := BuildTenantPrefix(99)
	got3 := LongestPrefixMatch(dev, []netip.Prefix{otherTenant})
	if got3.IsValid() {
		t.Errorf("LPM with no match = %v, want zero Prefix", got3)
	}
}

func TestContainsAddr_Equivalent(t *testing.T) {
	// ContainsAddr 是核心"X 在 Y 内吗"原语；验证它就是标准 IPv6 prefix.Contains
	tenant := BuildTenantPrefix(1)
	branch, _ := DeriveBranchPrefix(tenant, 1)
	dev, _ := DeriveDeviceAddr(branch, "A2ACD523")

	if !ContainsAddr(tenant, dev) {
		t.Errorf("tenant should contain device")
	}
	if !ContainsAddr(branch, dev) {
		t.Errorf("branch should contain device")
	}
	other := BuildTenantPrefix(99)
	if ContainsAddr(other, dev) {
		t.Errorf("other tenant should NOT contain device")
	}
}

// =============================================================================
// SlotsOf 反解
// =============================================================================

func TestSlotsOf(t *testing.T) {
	tenantPx := BuildTenantPrefix(0x0001)
	branchPx, _ := DeriveBranchPrefix(tenantPx, 0x01)
	sitePx, _ := DeriveSitePrefix(branchPx, 1, 2)
	unitPx, _ := DeriveUnitPrefix(sitePx, 0x0042)
	roomPx, _ := DeriveRoomPrefix(unitPx, 0x03)
	bedPx, _ := DeriveBedPrefix(roomPx, 0x01)
	dev, err := DeriveDeviceAddr(bedPx, "E598A2ACD523")
	if err != nil {
		t.Fatalf("derive err: %v", err)
	}

	tenant, branch, site, unit, room, bed, host, err := SlotsOf(dev)
	if err != nil {
		t.Fatalf("SlotsOf err: %v", err)
	}
	if tenant != 0x0001 || branch != 0x01 || site != 0x12 || unit != 0x0042 || room != 0x03 || bed != 0x01 {
		t.Errorf("slots: tenant=%#x branch=%#x site=%#x unit=%#x room=%#x bed=%#x; want 1/1/0x12/0x42/3/1",
			tenant, branch, site, unit, room, bed)
	}
	if host != 0xA2ACD523 {
		t.Errorf("device host = %#x, want 0xA2ACD523", host)
	}

	b, f := UnpackSiteSlot(site)
	if b != 1 || f != 2 {
		t.Errorf("UnpackSiteSlot(0x12) = (%d,%d), want (1,2)", b, f)
	}
}

func TestSlotsOf_NotOwl(t *testing.T) {
	other := netip.MustParseAddr("2001:db8::1")
	if _, _, _, _, _, _, _, err := SlotsOf(other); !errors.Is(err, ErrNotOwlNamespace) {
		t.Errorf("expected ErrNotOwlNamespace, got %v", err)
	}
}

// =============================================================================
// ScopeOf
// =============================================================================

func TestScopeOf(t *testing.T) {
	tenant := BuildTenantPrefix(1)
	branch, _ := DeriveBranchPrefix(tenant, 1)
	site, _ := DeriveSitePrefix(branch, 0, 0)
	unit, _ := DeriveUnitPrefix(site, 1)
	room, _ := DeriveRoomPrefix(unit, 1)
	bed, _ := DeriveBedPrefix(room, 1)

	cases := []struct {
		p    netip.Prefix
		want Scope6
	}{
		{tenant, Scope6Tenant},
		{branch, Scope6Branch},
		{site, Scope6Site},
		{unit, Scope6Unit},
		{room, Scope6Room},
		{bed, Scope6Bed},
	}
	for _, c := range cases {
		if got := ScopeOf(c.p); got != c.want {
			t.Errorf("ScopeOf(%s) = %v, want %v", c.p, got, c.want)
		}
	}

	// 非标准长度
	weird := netip.PrefixFrom(netip.MustParseAddr("fd00::1"), 50)
	if got := ScopeOf(weird); got != Scope6Unknown {
		t.Errorf("ScopeOf(/50) = %v, want Unknown", got)
	}
}

// =============================================================================
// ReverseDNS（ip6.arpa）— BIND zone 兼容性硬指标
// =============================================================================

func TestReverseDNS(t *testing.T) {
	// 标准 RFC 3596 ip6.arpa 反向命名：每 nibble 倒序，加 .ip6.arpa
	cases := []struct {
		addr string
		want string
	}{
		{
			// 全零 fd00::
			addr: "fd00::",
			want: "0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.d.f.ip6.arpa",
		},
		{
			// doc §2.2 完整示例
			addr: "fd00:0:1:112:42:301:a2ac:d523",
			want: "3.2.5.d.c.a.2.a.1.0.3.0.2.4.0.0.2.1.1.0.1.0.0.0.0.0.0.0.0.0.d.f.ip6.arpa",
		},
		{
			// loopback ::1
			addr: "::1",
			want: "1.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.ip6.arpa",
		},
	}
	for _, c := range cases {
		addr := netip.MustParseAddr(c.addr)
		got := ReverseDNS(addr)
		if got != c.want {
			t.Errorf("ReverseDNS(%s):\n got: %s\nwant: %s", c.addr, got, c.want)
		}
	}
}
