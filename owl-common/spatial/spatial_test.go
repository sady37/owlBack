package spatial

import "testing"

func TestBuildPath_emptyToWildcard(t *testing.T) {
	got := BuildPath("T", "", "", "", "U101", "R1", "B1")
	want := "T.0.0.0.U101.R1.B1"
	if got != want {
		t.Errorf("BuildPath() = %q, want %q", got, want)
	}
}

func TestBuildPath_full(t *testing.T) {
	got := BuildPath("T", "Br", "Bd", "1", "U101", "R1", "B1")
	want := "T.Br.Bd.1.U101.R1.B1"
	if got != want {
		t.Errorf("BuildPath() = %q, want %q", got, want)
	}
}

func TestBuildPath_unitScope(t *testing.T) {
	got := BuildPath("T", "Br", "Bd", "1", "U101", "", "")
	want := "T.Br.Bd.1.U101.0.0"
	if got != want {
		t.Errorf("BuildPath() = %q, want %q", got, want)
	}
}

func TestSegments(t *testing.T) {
	got := Segments("T.Br.Bd.1.U.R.B")
	want := [NumSegments]string{"T", "Br", "Bd", "1", "U", "R", "B"}
	if got != want {
		t.Errorf("Segments() = %v, want %v", got, want)
	}
}

func TestSegments_short(t *testing.T) {
	got := Segments("T.0.0.0.U")
	want := [NumSegments]string{"T", "0", "0", "0", "U", "0", "0"}
	if got != want {
		t.Errorf("Segments() = %v, want %v", got, want)
	}
}

func TestPrefixOf(t *testing.T) {
	cases := []struct {
		path string
		n    int
		want string
	}{
		{"T.Br.Bd.1.U.R.B", 5, "T.Br.Bd.1.U"},
		{"T.Br.Bd.1.U.R.B", 1, "T"},
		{"T.Br.Bd.1.U.R.B", 7, "T.Br.Bd.1.U.R.B"},
		{"T.Br.Bd.1.U.R.B", 9, "T.Br.Bd.1.U.R.B"},
		{"T.Br.Bd.1.U.R.B", 0, ""},
	}
	for _, c := range cases {
		if got := PrefixOf(c.path, c.n); got != c.want {
			t.Errorf("PrefixOf(%q, %d) = %q, want %q", c.path, c.n, got, c.want)
		}
	}
}

func TestMatchPrefix(t *testing.T) {
	cases := []struct {
		path, prefix string
		want         bool
	}{
		{"T.Br.Bd.1.U.R.B", "T.Br.Bd.1.U", true},
		{"T.Br.Bd.1.U2.R.B", "T.Br.Bd.1.U", false}, // boundary: U2 ≠ U
		{"T.Br.Bd.1.U", "T.Br.Bd.1.U", true},        // exact
		{"T.Br", "T.Br.Bd", false},                    // shorter than prefix
	}
	for _, c := range cases {
		if got := MatchPrefix(c.path, c.prefix); got != c.want {
			t.Errorf("MatchPrefix(%q, %q) = %v, want %v", c.path, c.prefix, got, c.want)
		}
	}
}

func TestMatchAddress_wildcard(t *testing.T) {
	cases := []struct {
		path, pattern string
		want          bool
	}{
		{"T.Br.Bd.1.U101.R1.B1", "T.0.0.0.U101.0.0", true},      // unit scope wildcard match
		{"T.Br.Bd.1.U101.R1.B1", "T.0.0.0.U102.0.0", false},     // different unit
		{"T.0.0.0.U101.R1.B1", "T.Br.0.0.U101.0.0", true},       // path missing branch, pattern has it → still match
		{"T.Br.Bd.1.U101.R1.B1", "T.Br.Bd.1.U101.R1.B1", true},  // exact
		{"T.Br.Bd.1.U101.R1.B1", "T.Br.Bd.1.U101.R1.B2", false}, // bed differs
	}
	for _, c := range cases {
		if got := MatchAddress(c.path, c.pattern); got != c.want {
			t.Errorf("MatchAddress(%q, %q) = %v, want %v", c.path, c.pattern, got, c.want)
		}
	}
}

func TestLongestCommonPrefix(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"T.Br.Bd.1.U101.R1.B1", "T.Br.Bd.1.U101.R2.B2", 5}, // unit same, room differs
		{"T.Br.Bd.1.U101.R1.B1", "T.Br.Bd.1.U101.R1.B1", 7}, // exact
		{"T.Br.Bd.1.U101.R1.B1", "T.Br.Bd.1.U101.0.0", 5},   // wildcard at room → break at depth 5
		{"T.Br.Bd.1.U101.R1.B1", "T2.Br.Bd.1.U101.R1.B1", 0}, // tenant differs
	}
	for _, c := range cases {
		if got := LongestCommonPrefix(c.a, c.b); got != c.want {
			t.Errorf("LongestCommonPrefix(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestAddress_Path(t *testing.T) {
	a := Address{Tenant: "T", Unit: "U101", Room: "R1", Bed: "B1"}
	got := a.Path()
	want := "T.0.0.0.U101.R1.B1"
	if got != want {
		t.Errorf("Address.Path() = %q, want %q", got, want)
	}
}

func TestParse(t *testing.T) {
	a := Parse("T.Br.Bd.1.U.R.B")
	if a.Tenant != "T" || a.Branch != "Br" || a.Bed != "B" || a.Prefix != NumSegments {
		t.Errorf("Parse() = %+v, want full Address with Prefix=7", a)
	}
}

func TestAddress_Within(t *testing.T) {
	bedroom := Address{Tenant: "T", Unit: "U101", Room: "R1", Bed: "B1"}
	unitScope := Address{Tenant: "T", Unit: "U101"}
	otherUnit := Address{Tenant: "T", Unit: "U102"}
	if !bedroom.Within(unitScope) {
		t.Error("bedroom should be within unitScope")
	}
	if bedroom.Within(otherUnit) {
		t.Error("bedroom should NOT be within otherUnit")
	}
}

func TestAddress_Validate(t *testing.T) {
	tenantUUID := "43b8fbf7-b55f-4b48-bd8b-27bb14f48870"
	unitUUID := "12a63cf9-9455-4f56-9c51-8bc32d27bdbc"

	// 合法：完整 UUID + Floor 短串
	full := Address{Tenant: tenantUUID, Floor: "1", Unit: unitUUID}
	if err := full.Validate(); err != nil {
		t.Errorf("full.Validate() = %v, want nil", err)
	}

	// 合法：仅 Tenant，其它通配（生产者只到 tenant 精度）
	tenantOnly := Address{Tenant: tenantUUID}
	if err := tenantOnly.Validate(); err != nil {
		t.Errorf("tenantOnly.Validate() = %v, want nil", err)
	}

	// 合法：Floor 段用 "B1"
	basement := Address{Tenant: tenantUUID, Floor: "B1", Unit: unitUUID}
	if err := basement.Validate(); err != nil {
		t.Errorf("basement.Validate() = %v, want nil", err)
	}

	// 非法：Tenant 空
	if err := (Address{}).Validate(); err == nil {
		t.Error("Validate() should fail when Tenant empty")
	}

	// 非法：Tenant 是通配
	if err := (Address{Tenant: FieldIgnore}).Validate(); err == nil {
		t.Error("Validate() should fail when Tenant is wildcard")
	}

	// 非法：Tenant 不是 UUID
	if err := (Address{Tenant: "Denver-LakeW"}).Validate(); err == nil {
		t.Error("Validate() should fail when Tenant is plain text")
	}

	// 非法：Unit 不是 UUID（明文）
	if err := (Address{Tenant: tenantUUID, Unit: "101corn"}).Validate(); err == nil {
		t.Error("Validate() should fail when Unit is plain text")
	}

	// 非法：Floor 太长
	if err := (Address{Tenant: tenantUUID, Floor: "very-long-floor-name"}).Validate(); err == nil {
		t.Error("Validate() should fail when Floor exceeds 8 chars")
	}

	// 非法：Floor 含特殊字符
	if err := (Address{Tenant: tenantUUID, Floor: "1!@"}).Validate(); err == nil {
		t.Error("Validate() should fail when Floor has special chars")
	}
}

func TestValidatePath(t *testing.T) {
	tenantUUID := "43b8fbf7-b55f-4b48-bd8b-27bb14f48870"
	unitUUID := "12a63cf9-9455-4f56-9c51-8bc32d27bdbc"
	roomUUID := "028ee067-d36e-4154-b230-ef29b94abaf7"
	bedUUID := "abcdef01-2345-6789-abcd-ef0123456789"

	good := tenantUUID + ".0.0.1." + unitUUID + "." + roomUUID + "." + bedUUID
	if err := ValidatePath(good); err != nil {
		t.Errorf("ValidatePath(good) = %v, want nil", err)
	}

	wildcardTenant := "0.0.0.0.0.0.0"
	if err := ValidatePath(wildcardTenant); err == nil {
		t.Error("ValidatePath should fail when tenant is wildcard")
	}

	plainText := "Denver-LakeW.0.0.0." + unitUUID + ".0.0"
	if err := ValidatePath(plainText); err == nil {
		t.Error("ValidatePath should fail when tenant is plain text")
	}
}

func TestIsValidSegment(t *testing.T) {
	uuid := "43b8fbf7-b55f-4b48-bd8b-27bb14f48870"

	// FieldIgnore 永远合法
	if !IsValidSegment("0", SegmentUUID) {
		t.Error(`IsValidSegment("0", UUID) should be true`)
	}
	if !IsValidSegment("0", SegmentFloor) {
		t.Error(`IsValidSegment("0", Floor) should be true`)
	}

	// 空字符串视为 FieldIgnore
	if !IsValidSegment("", SegmentUUID) {
		t.Error(`IsValidSegment("", UUID) should be true（empty = wildcard）`)
	}

	// UUID kind
	if !IsValidSegment(uuid, SegmentUUID) {
		t.Error("UUID should be valid for SegmentUUID")
	}
	if IsValidSegment("not-a-uuid", SegmentUUID) {
		t.Error("plain text should NOT be valid for SegmentUUID")
	}
	if IsValidSegment("1", SegmentUUID) {
		t.Error("floor-style should NOT be valid for SegmentUUID")
	}

	// Floor kind
	for _, ok := range []string{"1", "2", "B1", "G", "M", "12A"} {
		if !IsValidSegment(ok, SegmentFloor) {
			t.Errorf("%q should be valid for SegmentFloor", ok)
		}
	}
	for _, bad := range []string{"floor-1", "1!", "very-long-floor-name", "FloorWithSpaces "} {
		if IsValidSegment(bad, SegmentFloor) {
			t.Errorf("%q should NOT be valid for SegmentFloor", bad)
		}
	}
}

func TestAddress_InferPrefix(t *testing.T) {
	cases := []struct {
		a    Address
		want int
	}{
		{Address{Tenant: "T"}, 1},
		{Address{Tenant: "T", Branch: "Br", Building: "Bd", Floor: "1", Unit: "U"}, 5},
		{Address{Tenant: "T", Branch: "Br", Building: "Bd", Floor: "1", Unit: "U", Room: "R", Bed: "B"}, 7},
		{Address{Tenant: "T", Unit: "U"}, 1}, // 中间断了，从 Branch 起就是 "0"
	}
	for _, c := range cases {
		if got := c.a.InferPrefix(); got != c.want {
			t.Errorf("InferPrefix(%+v) = %d, want %d", c.a, got, c.want)
		}
	}
}
