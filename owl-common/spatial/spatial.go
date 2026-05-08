// Package spatial provides IPv6-style hierarchical spatial addressing for owl
// datagrams: a 7-segment path "tenant.branch.building.floor.unit.room.bed"
// with CIDR-style prefix scope and per-segment wildcard ("0") matching.
//
// 设计参见 owlBack/doc/datagram_envelope.md。
//
// # 用法快速参考
//
//	// 1) 构造路径（空段自动转 "0" 通配）
//	p := spatial.BuildPath(tenantID, branchID, buildingID, floor, unitID, roomID, bedID)
//	// 单楼宇租户：spatial.BuildPath("T", "", "", "", "U101", "R1", "B1") → "T.0.0.0.U101.R1.B1"
//
//	// 2) 用 Address 结构（更直观，等价）
//	a := spatial.Address{Tenant: "T", Unit: "U101", Room: "R1", Bed: "B1", Prefix: spatial.ScopeBed}
//	a.Path() // → "T.0.0.0.U101.R1.B1"
//
//	// 3) 前缀匹配（CIDR-style 快路径）
//	spatial.MatchPrefix("T.Br.Bd.1.U101.R1.B1", "T.Br.Bd.1.U101") // → true
//
//	// 4) 通配匹配（IPv6 inverse mask 风格，灵活）
//	spatial.MatchAddress("T.Br.Bd.1.U101.R1.B1", "T.0.0.0.U101.0.0") // → true
//
//	// 5) Longest-prefix-match（cardagg device→card 路由风格）
//	spatial.LongestCommonPrefix("T.Br.Bd.1.U101.R1.B1", "T.Br.Bd.1.U101.R1.0") // → 6 (room scope)
package spatial

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// 段分隔符 + 通配符常量
const (
	PathSep     = "."
	FieldIgnore = "0" // wildcard sentinel；空段自动转此值；consumer 通配匹配时识别
	NumSegments = 7
)

// Scope 7 级深度常量（producer 写到的精度，consumer prefix-match 用）
const (
	ScopeTenant   = 1 // T
	ScopeBranch   = 2 // T.Br
	ScopeBuilding = 3 // T.Br.Bd
	ScopeFloor    = 4 // T.Br.Bd.F
	ScopeUnit     = 5 // T.Br.Bd.F.U
	ScopeRoom     = 6 // T.Br.Bd.F.U.R
	ScopeBed      = 7 // T.Br.Bd.F.U.R.B
)

// =============================================================================
// 字符串路径 API（低层，性能友好）
// =============================================================================

// BuildPath 拼 7 段路径，空段自动转 FieldIgnore（"0"）。
//
//	BuildPath("T", "Br", "Bd", "1", "U", "R", "B")  → "T.Br.Bd.1.U.R.B"
//	BuildPath("T", "",   "",   "",  "U", "R", "B")  → "T.0.0.0.U.R.B"   单楼宇租户
//	BuildPath("T", "Br", "Bd", "1", "U", "",  "")   → "T.Br.Bd.1.U.0.0" unit scope
func BuildPath(tenantID, branchID, buildingID, floor, unitID, roomID, bedID string) string {
	parts := [NumSegments]string{tenantID, branchID, buildingID, floor, unitID, roomID, bedID}
	for i := range parts {
		if parts[i] == "" {
			parts[i] = FieldIgnore
		}
	}
	return strings.Join(parts[:], PathSep)
}

// Segments 把 7 段路径切成 string slice。短路径补 FieldIgnore 补到 7。
func Segments(path string) [NumSegments]string {
	var out [NumSegments]string
	for i := range out {
		out[i] = FieldIgnore
	}
	if path == "" {
		return out
	}
	idx := 0
	start := 0
	for i := 0; i < len(path) && idx < NumSegments; i++ {
		if path[i] == PathSep[0] {
			out[idx] = path[start:i]
			idx++
			start = i + 1
		}
	}
	if idx < NumSegments && start <= len(path) {
		out[idx] = path[start:]
	}
	return out
}

// PrefixOf 取 path 的前 n 段（CIDR-style 前缀）；n 越界回退到 NumSegments。
//
//	PrefixOf("T.Br.Bd.1.U.R.B", 5) → "T.Br.Bd.1.U"
func PrefixOf(path string, n int) string {
	if n <= 0 {
		return ""
	}
	if n >= NumSegments {
		return path
	}
	count := 0
	for i, c := range path {
		if c == '.' {
			count++
			if count == n {
				return path[:i]
			}
		}
	}
	return path
}

// MatchPrefix 判断 path 是否落在 prefix 描述的 scope 内（前缀完全匹配）。
// prefix 应当是 PrefixOf 的输出，无 trailing dot。不处理 wildcard——纯前缀语义。
//
//	MatchPrefix("T.Br.Bd.1.U.R.B",  "T.Br.Bd.1.U") → true
//	MatchPrefix("T.Br.Bd.1.U2.R.B", "T.Br.Bd.1.U") → false（边界保护，避免 U 误匹配 U2）
func MatchPrefix(path, prefix string) bool {
	if prefix == "" {
		return true
	}
	if len(path) < len(prefix) {
		return false
	}
	if path[:len(prefix)] != prefix {
		return false
	}
	if len(path) == len(prefix) {
		return true
	}
	return path[len(prefix)] == '.'
}

// MatchAddress 段级 wildcard 匹配（IPv6 inverse mask 风格）。
// pattern / path 任一段是 FieldIgnore（"0"）都视为通配。
//
//	MatchAddress("T.Br.Bd.1.U101.R1.B1", "T.0.0.0.U101.0.0")  → true
//	MatchAddress("T.Br.Bd.1.U101.R1.B1", "T.0.0.0.U102.0.0")  → false
//	MatchAddress("T.0.0.0.U101.R1.B1",   "T.Br1.0.0.U101.0.0") → true（producer 没填 branch，仍匹配）
func MatchAddress(path, pattern string) bool {
	pSeg := Segments(path)
	mSeg := Segments(pattern)
	for i := 0; i < NumSegments; i++ {
		if pSeg[i] == FieldIgnore || mSeg[i] == FieldIgnore {
			continue
		}
		if pSeg[i] != mSeg[i] {
			return false
		}
	}
	return true
}

// LongestCommonPrefix 返回 a 与 b 段对段相同的深度（0-7）。遇到 FieldIgnore 视为不匹配中断。
// 用于 cardagg 风格的 longest-prefix-match 路由：找出两个路径在哪一级开始分歧。
//
//	LongestCommonPrefix("T.Br.Bd.1.U101.R1.B1", "T.Br.Bd.1.U101.R2.B2") → 5（unit 同，room 起分歧）
//	LongestCommonPrefix("T.Br.Bd.1.U101.R1.B1", "T.Br.Bd.1.U101.0.0")   → 5（room 段是 0 = 通配，不算精确匹配深度）
func LongestCommonPrefix(a, b string) int {
	aSeg := Segments(a)
	bSeg := Segments(b)
	for i := 0; i < NumSegments; i++ {
		if aSeg[i] == FieldIgnore || bSeg[i] == FieldIgnore {
			return i
		}
		if aSeg[i] != bSeg[i] {
			return i
		}
	}
	return NumSegments
}

// =============================================================================
// Address 结构 API（高层，typed 友好）
// =============================================================================

// Address 7 段空间坐标的 typed 表达。Path() / Parse 跟字符串 API 互转。
type Address struct {
	Tenant   string
	Branch   string
	Building string
	Floor    string
	Unit     string
	Room     string
	Bed      string

	// Prefix 1-7：producer 自己声明的精度承诺。
	// 0 表示未声明，默认按"非通配段数"推断。
	Prefix int
}

// Path 把 Address 渲染为 7 段字符串路径（空段转 "0"）。
func (a Address) Path() string {
	return BuildPath(a.Tenant, a.Branch, a.Building, a.Floor, a.Unit, a.Room, a.Bed)
}

// Parse 从字符串路径解析为 Address；段不足 7 个时短缺段填 FieldIgnore。
// 不会出错（即使段值带空白），永远返回可用的 Address。
func Parse(path string) Address {
	seg := Segments(path)
	a := Address{
		Tenant:   seg[0],
		Branch:   seg[1],
		Building: seg[2],
		Floor:    seg[3],
		Unit:     seg[4],
		Room:     seg[5],
		Bed:      seg[6],
	}
	a.Prefix = a.InferPrefix()
	return a
}

// InferPrefix 按"从前往后第一个 FieldIgnore 之前的段数"推断精度。
// 用在 producer 没显式声明 Prefix 时。
//
//	{Tenant: "T", Unit: "U", Room: "R"}.InferPrefix() → 1（Branch 起就是 "0"，连续段被打断）
//	{Tenant: "T", Branch: "Br", Building: "Bd", Floor: "1", Unit: "U", Room: "R", Bed: "B"}.InferPrefix() → 7
func (a Address) InferPrefix() int {
	parts := [NumSegments]string{a.Tenant, a.Branch, a.Building, a.Floor, a.Unit, a.Room, a.Bed}
	for i, v := range parts {
		if v == "" || v == FieldIgnore {
			return i
		}
	}
	return NumSegments
}

// Within 判断 a 是否落在 scope 描述的范围内（CIDR-style 前缀匹配 + 通配兼容）。
// scope 为零值（全 "0"）时视为 tenant 全局，匹配任何 a。
func (a Address) Within(scope Address) bool {
	return MatchAddress(a.Path(), scope.Path())
}

// Validate 检查 Address 是否合法：
//   1. Tenant 必填（不能为空也不能是通配 "0"，producer 没资格）
//   2. 每段值必须符合 SegmentKindAt 期望的格式（UUID 或 floor 短串 或 FieldIgnore）
//
// 注：空字符串和 FieldIgnore 视为同义（BuildPath 会把空段转 "0"）。
func (a Address) Validate() error {
	if a.Tenant == "" || a.Tenant == FieldIgnore {
		return ErrTenantRequired
	}
	parts := [NumSegments]string{a.Tenant, a.Branch, a.Building, a.Floor, a.Unit, a.Room, a.Bed}
	for i, v := range parts {
		kind := SegmentKindAt(i)
		if !IsValidSegment(v, kind) {
			return fmt.Errorf("spatial: segment[%d] %q invalid for kind %s (expect UUID/0 or floor short-id)", i, v, kind)
		}
	}
	return nil
}

// ValidatePath 校验整个字符串 path：每段格式合规 + tenant 段非通配。
// 用于消费方接收消息时反向校验，或 producer wire 出口处自检。
func ValidatePath(path string) error {
	seg := Segments(path)
	if seg[0] == "" || seg[0] == FieldIgnore {
		return ErrTenantRequired
	}
	for i, v := range seg {
		kind := SegmentKindAt(i)
		if !IsValidSegment(v, kind) {
			return fmt.Errorf("spatial: segment[%d] %q invalid for kind %s", i, v, kind)
		}
	}
	return nil
}

// SegmentKind 段值类型分类。
type SegmentKind int

const (
	// SegmentUUID 期望 UUID v4 形态（tenant/branch/building/unit/room/bed）
	SegmentUUID SegmentKind = iota
	// SegmentFloor 期望短字母数字串（如 "1" / "2" / "B1" / "G" / "M"）
	SegmentFloor
)

func (k SegmentKind) String() string {
	switch k {
	case SegmentUUID:
		return "UUID"
	case SegmentFloor:
		return "Floor"
	}
	return "Unknown"
}

// SegmentKindAt 返回 7 段路径中第 i 段（0-based）期望的 SegmentKind。
// 索引 0=tenant 1=branch 2=building 3=floor 4=unit 5=room 6=bed
// 只有 floor（index=3）是短串；其余都期望 UUID。
func SegmentKindAt(index int) SegmentKind {
	if index == 3 {
		return SegmentFloor
	}
	return SegmentUUID
}

// IsValidSegment 校验单段值。FieldIgnore（"0"）永远合法（通配）；
// 空字符串视同 FieldIgnore（BuildPath 会自动转）；
// 否则按 kind 校验：UUID 段必须是 UUID v4 格式；Floor 段必须是 1-8 位字母数字。
func IsValidSegment(value string, kind SegmentKind) bool {
	if value == "" || value == FieldIgnore {
		return true
	}
	switch kind {
	case SegmentUUID:
		return uuidRegex.MatchString(value)
	case SegmentFloor:
		return floorRegex.MatchString(value)
	}
	return false
}

// 标准 UUID v4 正则（带连字符，case-insensitive）
var uuidRegex = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// Floor 段：1-8 位字母数字（覆盖 "1" "2" "B1" "G" "M" "12A" 等）
var floorRegex = regexp.MustCompile(`^[A-Za-z0-9]{1,8}$`)

// 错误
var ErrTenantRequired = errors.New("spatial: tenant_id is required and cannot be wildcard")
