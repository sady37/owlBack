package spatial

// Subject (resident / caregiver) addressing — Mobile IPv6 (RFC 6275) 风格。
//
// owl 把 "人" 看作 RFC 6275 mobile node：
//
//   - Home Address (HoA)  = 永久身份地址，在 owl 编码体系内 deterministic 派生
//   - Care-of Address (CoA) = 当前位置 = 当前 tracking 该 subject 的 device 地址
//   - Binding Cache       = HoA → 当前 CoA 映射（小表，分钟级更新）
//
// 业务查询 "subject 在哪 / 谁在 room X / device D 是否对应 caregiver C" 全部
// 退化成 IPv6 prefix.Contains() 标准原语，零自定义关系判断代码。
//
// # 编码方案：tenant 内保留 branch=0xFF 作 subject namespace
//
//	bits   0-31  : owl ULA (fd00:0)
//	bits  32-47  : tenant slot (uint16)            — 与 spatial 同
//	bits  48-55  : 0xFF                            — branch slot 保留值，subject 标识
//	bits  56-63  : subject kind (uint8)            — 0x01=resident, 0x02=caregiver, ...
//	bits  64-79  : subject_id (uint16)             — 65k subject per (tenant, kind)
//	bits  80-127 : reserved zero (48 bit)          — 未来 subject 子属性扩展位
//
// 例：tenant=1 / 第 0x42 号 resident
//
//	HoA = fd00:0:0001:FF01:0042:0000:0000:0000/128
//	    = fd00:0:1:ff01:42::/128 (canonical 文本)
//	     │   │   │  └ KK=resident (0x01)
//	     │   │   └─ branch=0xFF (subject namespace)
//	     │   └─ tenant slot 1
//	     └─ owl ULA
//
// # 与 spatial 编码宽度一致
//
//	spatial:  tenant(16) | branch(8)  | site(8) | unit(16)     | room(8) | bed(8) | host(32)
//	subject:  tenant(16) | 0xFF(8)    | kind(8) | subject_id(16)| reserved(48)
//
// subject_id 与 unit_slot 同 16 bit，65k/(tenant×kind) 远超任何 elder care 总人数。
//
// # Prefix 查询语义（标准 IPv6 longest-prefix-match）
//
//	"tenant 1 内任意 subject"      = fd00:0:1:FF00::/56
//	"tenant 1 的所有 residents"    = fd00:0:1:FF01::/64
//	"tenant 1 的所有 caregivers"   = fd00:0:1:FF02::/64
//	"特定 resident 0x42"          = fd00:0:1:FF01:42::/128
//
// # admin / manager 不进入此体系
//
// 远程登录用户没有物理位置，不该有 spatial / subject 地址。
// 他们继续用 UUID + role 在 user 表里管理。
//
// # caregiver 占位不接线（Phase A）
//
// SubjectKindCaregiver=0x02 常量永久占住，供未来业务接入；当前阶段：
//   - 不暴露 BuildCaregiverHoA 便利函数（避免误用）
//   - 调用方需要 caregiver HoA 时显式 BuildSubjectHoA(tenant, SubjectKindCaregiver, id)
//   - DB 应预留 caregiver subject_id / home_addr 列（NULL 允许）
//   - alarm 路由 / 访问权限沿用 user role 模型，不接 CoA
//   - 未来引入 caregiver tracking 时，加 caregiver_binding 表 + service，本包不动
//
// # branch=0xFF 与空间 branch 共存
//
// allocator 给空间 branch 分配 slot 时**必须跳过 0xFF**（约定，不在 encoder
// 层面强制——encoder 是纯函数，所有 byte 值合法）。0xFE 个空间 branch 槽位
// 对 owl 业务足够。

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/netip"
)

// SubjectKind 标识 subject 类型（owl namespace 内 site slot byte 位置）。
type SubjectKind uint8

const (
	SubjectKindUnknown   SubjectKind = 0x00
	SubjectKindResident  SubjectKind = 0x01
	SubjectKindCaregiver SubjectKind = 0x02
	// 0x03..0xFE 保留作未来 subject 类型（如 visitor / family / external_doctor / consultant）
)

func (k SubjectKind) String() string {
	switch k {
	case SubjectKindResident:
		return "resident"
	case SubjectKindCaregiver:
		return "caregiver"
	}
	return "unknown"
}

// BranchSlotSubject branch slot 值 0xFF 保留作 subject namespace 标识。
// 任何 spatial branch allocator 必须跳过此值。
const BranchSlotSubject uint8 = 0xFF

// Subject namespace prefix 长度
const (
	// PrefixLenSubjectNS 整 tenant 的 subject namespace（包括所有 kinds）
	PrefixLenSubjectNS = 56
	// PrefixLenSubjectKind 特定 kind 下的 subject 集合（e.g. 所有 residents）
	PrefixLenSubjectKind = 64
	// PrefixLenSubjectHoA 单个 subject 的 Home Address（唯一标识）
	PrefixLenSubjectHoA = 128
)

// 错误
var (
	ErrInvalidSubjectKind = errors.New("spatial: subject kind 0x00 / 0xFF reserved or invalid")
	ErrNotSubjectAddr     = errors.New("spatial: address not in subject namespace (branch != 0xFF)")
	ErrSubjectAddrInvalid = errors.New("spatial: address has malformed subject layout")
)

// =============================================================================
// Builders
// =============================================================================

// BuildSubjectNamespacePrefix 返回 tenant 的 subject namespace /56 前缀，
// 包含该 tenant 内所有类型的 subject（residents / caregivers / ...）。
//
//	tenantSlot=1 → fd00:0:1:ff00::/56
func BuildSubjectNamespacePrefix(tenantSlot uint16) netip.Prefix {
	tenant := BuildTenantPrefix(tenantSlot)
	b := tenant.Masked().Addr().As16()
	b[6] = BranchSlotSubject
	return netip.PrefixFrom(netip.AddrFrom16(b), PrefixLenSubjectNS)
}

// BuildSubjectKindPrefix 返回 tenant 内特定 kind 的 subject /64 前缀。
//
//	tenantSlot=1, kind=Resident → fd00:0:1:ff01::/64
func BuildSubjectKindPrefix(tenantSlot uint16, kind SubjectKind) (netip.Prefix, error) {
	if kind == SubjectKindUnknown || kind == 0xFF {
		return netip.Prefix{}, fmt.Errorf("%w: kind=0x%02X", ErrInvalidSubjectKind, kind)
	}
	tenant := BuildTenantPrefix(tenantSlot)
	b := tenant.Masked().Addr().As16()
	b[6] = BranchSlotSubject
	b[7] = byte(kind)
	return netip.PrefixFrom(netip.AddrFrom16(b), PrefixLenSubjectKind), nil
}

// BuildSubjectHoA 派生 subject 的 Home Address (/128)。
//
// subjectID 是 per-(tenant, kind) 唯一标识，application 层负责唯一性
// （DB sequence / SELECT MAX+1 / 业务 ID 映射）。
//
//	tenantSlot=1, kind=Resident, subjectID=0x42
//	  → fd00:0:1:ff01:42::/128
//
// 同一 (tenant, kind, subjectID) 三元组永远派生同一 HoA — stateless，
// 不需要中央分配（对比：spatial branch slot 需 allocator 协调）。
//
// **caregiver 仍可调用本函数**（kind=SubjectKindCaregiver），但 caregiver 业务
// 在当前阶段不消费 HoA / CoA，调用方需自知所做（见 package doc）。
func BuildSubjectHoA(tenantSlot uint16, kind SubjectKind, subjectID uint16) (netip.Addr, error) {
	if kind == SubjectKindUnknown || kind == 0xFF {
		return netip.Addr{}, fmt.Errorf("%w: kind=0x%02X", ErrInvalidSubjectKind, kind)
	}
	tenant := BuildTenantPrefix(tenantSlot)
	b := tenant.Masked().Addr().As16()
	b[6] = BranchSlotSubject
	b[7] = byte(kind)
	// bytes 8-9: subject_id (big-endian uint16)
	b[8] = byte(subjectID >> 8)
	b[9] = byte(subjectID)
	// bytes 10-15: reserved zero
	return netip.AddrFrom16(b), nil
}

// BuildResidentHoA 便利函数：`BuildSubjectHoA(tenant, Resident, id)` 的 ergonomic 封装。
//
// 这是 Phase A 唯一暴露的 subject 便利 builder。caregiver / visitor / 其他
// kind 必须显式走 BuildSubjectHoA + SubjectKind 常量，强制调用方意识到
// "你正在使用尚未接业务的 subject namespace"。
func BuildResidentHoA(tenantSlot uint16, residentID uint16) netip.Addr {
	addr, _ := BuildSubjectHoA(tenantSlot, SubjectKindResident, residentID)
	return addr
}

// BuildSubjectHoAFromHex 把 hex 字符串（最长 4 字符 = 16 bit）当 subject_id
// 末段填进 HoA。便利函数：直接用 UUID 末段或 device_uid 末段 hex 时省手转 uint16。
//
//	BuildSubjectHoAFromHex(1, Resident, "0042")
//	  → fd00:0:1:ff01:42::
//
// hex 长度不够 4 时左侧补零；超 4 字符返回 ErrSubjectAddrInvalid。
func BuildSubjectHoAFromHex(tenantSlot uint16, kind SubjectKind, idHex string) (netip.Addr, error) {
	if len(idHex) > 4 {
		return netip.Addr{}, fmt.Errorf("%w: id hex %q > 4 chars (max uint16)",
			ErrSubjectAddrInvalid, idHex)
	}
	padded := idHex
	for len(padded) < 4 {
		padded = "0" + padded
	}
	idBytes, err := hex.DecodeString(padded)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("%w: hex decode %q: %v",
			ErrSubjectAddrInvalid, idHex, err)
	}
	id := uint16(idBytes[0])<<8 | uint16(idBytes[1])
	return BuildSubjectHoA(tenantSlot, kind, id)
}

// =============================================================================
// Inspectors / Parsers
// =============================================================================

// IsSubjectAddr 判断 addr 是否在 subject namespace 内（branch byte = 0xFF）。
// addr 必须先在 owl namespace 内（IsOwlAddr）；否则返回 false。
func IsSubjectAddr(addr netip.Addr) bool {
	if !IsOwlAddr(addr) {
		return false
	}
	b := addr.As16()
	return b[6] == BranchSlotSubject
}

// IsSpatialAddr 判断 addr 是否在 spatial namespace 内（device / room / unit / ...）。
// = IsOwlAddr(addr) && !IsSubjectAddr(addr)
func IsSpatialAddr(addr netip.Addr) bool {
	if !IsOwlAddr(addr) {
		return false
	}
	b := addr.As16()
	return b[6] != BranchSlotSubject
}

// SubjectKindOf 返回 subject addr 的 kind；非 subject addr 返回 SubjectKindUnknown。
func SubjectKindOf(addr netip.Addr) SubjectKind {
	if !IsSubjectAddr(addr) {
		return SubjectKindUnknown
	}
	return SubjectKind(addr.As16()[7])
}

// ParseSubjectAddr 反解 subject HoA：返回 tenant_slot / kind / subject_id。
// 非 subject addr 返回 ErrNotSubjectAddr；非 owl namespace 返回 ErrNotOwlNamespace。
func ParseSubjectAddr(addr netip.Addr) (tenantSlot uint16, kind SubjectKind, subjectID uint16, err error) {
	if !IsOwlAddr(addr) {
		return 0, SubjectKindUnknown, 0, ErrNotOwlNamespace
	}
	b := addr.As16()
	if b[6] != BranchSlotSubject {
		return 0, SubjectKindUnknown, 0, ErrNotSubjectAddr
	}
	tenantSlot = uint16(b[4])<<8 | uint16(b[5])
	kind = SubjectKind(b[7])
	subjectID = uint16(b[8])<<8 | uint16(b[9])
	return
}
