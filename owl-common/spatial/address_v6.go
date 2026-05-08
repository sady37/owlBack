package spatial

// IPv6 ULA-based spatial addressing for owl datagrams (Phase 2 / Datagram v1).
//
// 设计参见 owlBack/doc/datagram_envelope.md §2。本文件提供基于 net/netip 的
// 类型安全 API；旧的 7-段字符串路径 API（spatial.go）暂保留作过渡兼容，但所有
// 新代码应使用本文件的 IPv6 函数。
//
// # 设计原则：100% 标准 IPv6
//
// 地址完全是合法 IPv6 ULA（RFC 4193）；所有 wildcarding 走标准 prefix length
// 语义（不引入 inverse mask / 私有 sentinel）。这样：
//
//   - net/netip / Postgres INET / iproute2 / ip6tables / BGP 直接 work
//   - 反向 DNS（ip6.arpa）/ AAAA 正向 / BIND zone 文件直接配
//   - DHCPv6 server（ISC dhcpd / kea / dnsmasq）按 prefix scope 分配地址
//   - "device 在 room X 内吗" = 标准 prefix.Contains(addr)，零自定义代码
//
// # 128-bit 布局（per doc §2.1）
//
//	bits   0-15 : ULA fixed              = 0xFD00
//	bits  16-31 : owl namespace reserved = 0x0000
//	bits  32-47 : tenant slot   (uint16, 全局分配)
//	bits  48-55 : branch slot   (uint8 , per tenant)
//	bits  56-63 : site slot     (uint8 , per branch；= building<<4 | floor)
//	bits  64-79 : unit slot     (uint16, per site)
//	bits  80-87 : room slot     (uint8 , per unit)
//	bits  88-95 : bed slot      (uint8 , per room)
//	bits  96-127: device host   (uint32, = device_uid 末 32 bit；stateless 派生)
//
// # 各 scope 对应 prefix 长度
//
//	Tenant /48   Branch /56   Site /64   Unit /80   Room /88   Bed /96   Device /128
//
// "site" 段把 building 高 4 bit + floor 低 4 bit 打包成 1 字节（PackSiteSlot）。

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/netip"
)

// IPv6 layout 常量
const (
	// ULA + owl namespace 固定前 32 位
	ULAByte0   byte = 0xFD
	ULAByte1   byte = 0x00
	OwlNSByte0 byte = 0x00
	OwlNSByte1 byte = 0x00

	// Prefix 长度（按 scope）
	PrefixLenTenant = 48
	PrefixLenBranch = 56
	PrefixLenSite   = 64
	PrefixLenUnit   = 80
	PrefixLenRoom   = 88
	PrefixLenBed    = 96
	PrefixLenDevice = 128
)

// Scope6 为 IPv6 prefix 长度对应的语义 scope 标签。
type Scope6 int

const (
	Scope6Unknown Scope6 = iota
	Scope6Tenant
	Scope6Branch
	Scope6Site
	Scope6Unit
	Scope6Room
	Scope6Bed
	Scope6Device
)

func (s Scope6) String() string {
	switch s {
	case Scope6Tenant:
		return "tenant"
	case Scope6Branch:
		return "branch"
	case Scope6Site:
		return "site"
	case Scope6Unit:
		return "unit"
	case Scope6Room:
		return "room"
	case Scope6Bed:
		return "bed"
	case Scope6Device:
		return "device"
	}
	return "unknown"
}

// ScopeOf 根据 prefix 长度返回 scope 语义；非标准长度返回 Scope6Unknown。
func ScopeOf(prefix netip.Prefix) Scope6 {
	switch prefix.Bits() {
	case PrefixLenTenant:
		return Scope6Tenant
	case PrefixLenBranch:
		return Scope6Branch
	case PrefixLenSite:
		return Scope6Site
	case PrefixLenUnit:
		return Scope6Unit
	case PrefixLenRoom:
		return Scope6Room
	case PrefixLenBed:
		return Scope6Bed
	case PrefixLenDevice:
		return Scope6Device
	}
	return Scope6Unknown
}

// 错误
var (
	ErrInvalidPrefixLen = errors.New("spatial: prefix length does not match expected scope")
	ErrInvalidDeviceUID = errors.New("spatial: device_uid must end with 8+ hex chars")
	ErrNotOwlNamespace  = errors.New("spatial: address not in fd00:0000::/32 owl namespace")
)

// =============================================================================
// Builders / Derivers
// =============================================================================
//
// 所有 slot 值（uint8 / uint16）全范围 0..max 都合法；wildcarding 走 prefix
// length 语义，不引入 sentinel。"任意 X" = 用 X 的父 scope prefix 即可。

// BuildTenantPrefix 返回 tenant /48 prefix：fd00:0000:TTTT::/48。
//
//	slot=0x0001 → fd00:0:1::/48
func BuildTenantPrefix(slot uint16) netip.Prefix {
	var b [16]byte
	b[0] = ULAByte0
	b[1] = ULAByte1
	b[2] = OwlNSByte0
	b[3] = OwlNSByte1
	b[4] = byte(slot >> 8)
	b[5] = byte(slot)
	return netip.PrefixFrom(netip.AddrFrom16(b), PrefixLenTenant)
}

// DeriveBranchPrefix 在 tenant /48 下分配 branch /56（写 byte 6）。
func DeriveBranchPrefix(tenant netip.Prefix, branchSlot uint8) (netip.Prefix, error) {
	if tenant.Bits() != PrefixLenTenant {
		return netip.Prefix{}, fmt.Errorf("%w: expected /%d (tenant), got /%d",
			ErrInvalidPrefixLen, PrefixLenTenant, tenant.Bits())
	}
	b := tenant.Masked().Addr().As16()
	b[6] = branchSlot
	return netip.PrefixFrom(netip.AddrFrom16(b), PrefixLenBranch), nil
}

// DeriveSitePrefix 在 branch /56 下分配 site /64（写 byte 7）。
//
// 入参分别接收 building (4 bit, 0..15) 与 floor (4 bit, 0..15)；内部 pack
// 成单字节 site_slot。约定：building 0-15 都允许（0 = 单楼宇租户的默认楼）；
// floor 0 通常表示地面/特殊，1-15 是楼层号。
func DeriveSitePrefix(branch netip.Prefix, building, floor uint8) (netip.Prefix, error) {
	if branch.Bits() != PrefixLenBranch {
		return netip.Prefix{}, fmt.Errorf("%w: expected /%d (branch), got /%d",
			ErrInvalidPrefixLen, PrefixLenBranch, branch.Bits())
	}
	b := branch.Masked().Addr().As16()
	b[7] = PackSiteSlot(building, floor)
	return netip.PrefixFrom(netip.AddrFrom16(b), PrefixLenSite), nil
}

// PackSiteSlot 把 building 高 4 bit + floor 低 4 bit 打包成 1 字节。
func PackSiteSlot(building, floor uint8) uint8 {
	return (building&0x0F)<<4 | (floor & 0x0F)
}

// UnpackSiteSlot 是 PackSiteSlot 的逆运算。
func UnpackSiteSlot(slot uint8) (building, floor uint8) {
	return slot >> 4, slot & 0x0F
}

// DeriveUnitPrefix 在 site /64 下分配 unit /80（写 byte 8-9）。
func DeriveUnitPrefix(site netip.Prefix, unitSlot uint16) (netip.Prefix, error) {
	if site.Bits() != PrefixLenSite {
		return netip.Prefix{}, fmt.Errorf("%w: expected /%d (site), got /%d",
			ErrInvalidPrefixLen, PrefixLenSite, site.Bits())
	}
	b := site.Masked().Addr().As16()
	b[8] = byte(unitSlot >> 8)
	b[9] = byte(unitSlot)
	return netip.PrefixFrom(netip.AddrFrom16(b), PrefixLenUnit), nil
}

// DeriveRoomPrefix 在 unit /80 下分配 room /88（写 byte 10）。
func DeriveRoomPrefix(unit netip.Prefix, roomSlot uint8) (netip.Prefix, error) {
	if unit.Bits() != PrefixLenUnit {
		return netip.Prefix{}, fmt.Errorf("%w: expected /%d (unit), got /%d",
			ErrInvalidPrefixLen, PrefixLenUnit, unit.Bits())
	}
	b := unit.Masked().Addr().As16()
	b[10] = roomSlot
	return netip.PrefixFrom(netip.AddrFrom16(b), PrefixLenRoom), nil
}

// DeriveBedPrefix 在 room /88 下分配 bed /96（写 byte 11）。
func DeriveBedPrefix(room netip.Prefix, bedSlot uint8) (netip.Prefix, error) {
	if room.Bits() != PrefixLenRoom {
		return netip.Prefix{}, fmt.Errorf("%w: expected /%d (room), got /%d",
			ErrInvalidPrefixLen, PrefixLenRoom, room.Bits())
	}
	b := room.Masked().Addr().As16()
	b[11] = bedSlot
	return netip.PrefixFrom(netip.AddrFrom16(b), PrefixLenBed), nil
}

// DeriveDeviceAddr 派生 /128 device 完整地址：基础 prefix + device_uid 末 32 bit 填 host。
//
// per doc §2.3：完全 stateless。device_uid 形如 "E598A2ACD523"；末 8 hex
// 直接当 host part 写入 byte 12-15。基础 prefix 通常是 bed /96，少数共享区
// device 用 unit /80 也可（host 段会落在更短 prefix 之后的零位上）。
//
// 例：bed prefix fd00:0:1:112:42:301::/96 + device_uid "E598A2ACD523"
//
//	→ fd00:0:1:112:42:301:a2ac:d523/128
func DeriveDeviceAddr(base netip.Prefix, deviceUID string) (netip.Addr, error) {
	if base.Bits() > 96 {
		return netip.Addr{}, fmt.Errorf("%w: device base prefix must be /96 or shorter, got /%d",
			ErrInvalidPrefixLen, base.Bits())
	}
	if len(deviceUID) < 8 {
		return netip.Addr{}, ErrInvalidDeviceUID
	}
	macSuffix := deviceUID[len(deviceUID)-8:]
	hostBits, err := hex.DecodeString(macSuffix)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("%w: hex decode failed for %q: %v",
			ErrInvalidDeviceUID, macSuffix, err)
	}
	b := base.Masked().Addr().As16()
	copy(b[12:], hostBits)
	return netip.AddrFrom16(b), nil
}

// =============================================================================
// Inspectors / Matchers
// =============================================================================

// IsOwlAddr 判断 addr 是否落在 fd00:0000::/32 owl 命名空间内。
// 用于消费方反向校验：拒绝来自其他 ULA namespace 的伪造消息。
func IsOwlAddr(addr netip.Addr) bool {
	if !addr.Is6() {
		return false
	}
	b := addr.As16()
	return b[0] == ULAByte0 && b[1] == ULAByte1 && b[2] == OwlNSByte0 && b[3] == OwlNSByte1
}

// ContainsAddr 判断 prefix 是否包含 addr（标准 IPv6 prefix containment）。
// 等价 Postgres `prefix >>= addr`；longest-prefix-match consumer 用。
//
// 这是核心查询原语："device 在 room/unit/branch X 内吗" 全部走这一个函数，
// 不需要写自定义 device→room→unit→branch 多表 JOIN 代码。
func ContainsAddr(prefix netip.Prefix, addr netip.Addr) bool {
	return prefix.Contains(addr)
}

// LongestPrefixMatch 在 candidates 中找包含 addr 且 mask 最长的 prefix。
// 命中返回该 prefix；未命中返回零值 Prefix（调用方用 .IsValid() 判断）。
//
// cardagg-style spatial→card_id 路由的核心算法。生产路径建议直接走 Postgres
// GiST 索引（O(log n)）；本函数提供 in-memory cache 的等价语义。
func LongestPrefixMatch(addr netip.Addr, candidates []netip.Prefix) netip.Prefix {
	var best netip.Prefix
	bestBits := -1
	for _, c := range candidates {
		if c.Contains(addr) && c.Bits() > bestBits {
			best = c
			bestBits = c.Bits()
		}
	}
	return best
}

// SlotsOf 反解 addr 各级 slot（仅 owl 命名空间内地址有效）。
// 返回值仅高于实际 prefix 长度的字段才有意义；建议结合 ScopeOf 使用。
func SlotsOf(addr netip.Addr) (tenant uint16, branch, site uint8, unit uint16, room, bed uint8, deviceHost uint32, err error) {
	if !IsOwlAddr(addr) {
		return 0, 0, 0, 0, 0, 0, 0, ErrNotOwlNamespace
	}
	b := addr.As16()
	tenant = uint16(b[4])<<8 | uint16(b[5])
	branch = b[6]
	site = b[7]
	unit = uint16(b[8])<<8 | uint16(b[9])
	room = b[10]
	bed = b[11]
	deviceHost = uint32(b[12])<<24 | uint32(b[13])<<16 | uint32(b[14])<<8 | uint32(b[15])
	return
}

// ReverseDNS 返回 addr 的 ip6.arpa 反向 DNS 域名（标准 RFC 3596）。
//
//	fd00:0:1:112:42:301:a2ac:d523
//	→ 3.2.5.d.c.a.2.a.1.0.3.0.2.4.0.0.2.1.1.0.1.0.0.0.0.0.0.0.0.0.0.d.f.ip6.arpa
//
// 用途：BIND 反向 zone 配置 / dig -x 查询 / 工业 DNS 解析器直接喂这个名字。
// owl 不强制运行 BIND，但 zone 文件能跑通是"跟工业 IPv6 一致"的硬指标之一。
func ReverseDNS(addr netip.Addr) string {
	b := addr.As16()
	out := make([]byte, 0, 73) // 32 hex chars × 2 (含点) + ".ip6.arpa"
	for i := 15; i >= 0; i-- {
		// 低 4 bit 在前
		lo := b[i] & 0x0F
		hi := (b[i] >> 4) & 0x0F
		out = append(out, hexNibble(lo), '.', hexNibble(hi), '.')
	}
	out = append(out, []byte("ip6.arpa")...)
	return string(out)
}

func hexNibble(n byte) byte {
	if n < 10 {
		return '0' + n
	}
	return 'a' + (n - 10)
}
