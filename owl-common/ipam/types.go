package ipam

import (
	"errors"
	"net/netip"
)

// 错误
var (
	ErrNotFound       = errors.New("ipam: spatial entity not found")
	ErrSlotExhausted  = errors.New("ipam: slot range exhausted at this level")
	ErrInvalidParent  = errors.New("ipam: parent prefix length does not match expected scope")
	ErrInvalidAttrs   = errors.New("ipam: required attribute missing")
	ErrAlreadyExists  = errors.New("ipam: entity already exists at this slot")
)

// BranchAttrs are passed to AllocateBranch. Name 必填；timezone NULL 继承 tenant。
type BranchAttrs struct {
	Name     string // 显示名 e.g. "Denver"
	Timezone string // IANA tz e.g. "America/Denver"; 空字符串 = 继承 tenant
	Address  string // 物理地址 free-form text
}

// SiteAttrs are passed to AllocateSite. building 0..14 / floor 0..14 显式由参数传，不在此结构。
type SiteAttrs struct {
	Name    string // e.g. "Building A Floor 2"
	Address string
}

// UnitAttrs are passed to AllocateUnit.
// UnitType 与 units.unit_type SMALLINT 同语义；常量在 owl-common/card.UnitType*.
// is_public / is_shared 由 unit_type 派生（=3 / =2），不单独存。
type UnitAttrs struct {
	Name     string // e.g. "201"
	UnitType int    // 0=Unknown / 1=Private / 2=Share / 3=Public
	Timezone string // IANA tz; 空 = 继承 site/branch/tenant
}

// RoomAttrs are passed to AllocateRoom.
// RoomType 取 owl-common/card.RoomType*（0=Default / 1=Bathroom / 2=Kitchen）。
type RoomAttrs struct {
	Name      string
	RoomType  int  // 0=Default / 1=Bathroom / 2=Kitchen
	IsPrimary bool // unit 内每 unit 仅 1 primary
}

// BedAttrs are passed to AllocateBed.
type BedAttrs struct {
	Name        string // e.g. "BedA"
	Description string
}

// TenantInfo is read-only result from LookupTenant.
type TenantInfo struct {
	Prefix       netip.Prefix
	Slot         uint16
	Name         string
	Timezone     string
	ContactName  string
	ContactEmail string
	ContactPhone string
}
