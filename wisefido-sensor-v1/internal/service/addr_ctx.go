package service

import (
	"net/netip"

	owlredis "owl-common/redis"
)

// AddrCtx 从 IoTStreamMessage.DeviceAddr 派生的全套字符串身份上下文。
//
// device_ipv6 单程票 (doc/device_ipv6_migration_checklist.md) 落地后，
// 上游 envelope 不再带 TenantID/DeviceID/UnitID/RoomID/BedID 等独立字段；
// 各 consumer 在入口处一次派生 AddrCtx，下游业务代码用其字符串字段 (Redis hash key 等)。
//
// 字段全部是 INET CIDR 文本（"fd00:0:3::/48"）或 /128 host 文本（"fd00:0:3:111:3:101:a2ac:d523"）。
// 无效 DeviceAddr 时各字段返回 ""，caller 可早退。
type AddrCtx struct {
	Addr        netip.Addr // 原始 /128
	DeviceAddr  string     // canonical /128 host text，无 mask
	TenantPref  string     // /48 CIDR
	BranchPref  string     // /56 CIDR
	UnitPref    string     // /80 CIDR
	RoomPref    string     // /88 CIDR
	BedPref     string     // /96 CIDR
}

// AddrCtxFromMsg 从消息 envelope 派生。
func AddrCtxFromMsg(m *owlredis.IoTStreamMessage) AddrCtx {
	if m == nil {
		return AddrCtx{}
	}
	return AddrCtxFromAddr(m.DeviceAddr)
}

// AddrCtxFromAddr 由 raw netip.Addr 派生。
func AddrCtxFromAddr(addr netip.Addr) AddrCtx {
	if !addr.IsValid() {
		return AddrCtx{}
	}
	return AddrCtx{
		Addr:       addr,
		DeviceAddr: addr.String(),
		TenantPref: netip.PrefixFrom(addr, 48).Masked().String(),
		BranchPref: netip.PrefixFrom(addr, 56).Masked().String(),
		UnitPref:   netip.PrefixFrom(addr, 80).Masked().String(),
		RoomPref:   netip.PrefixFrom(addr, 88).Masked().String(),
		BedPref:    netip.PrefixFrom(addr, 96).Masked().String(),
	}
}
