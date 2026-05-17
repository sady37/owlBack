package service

import (
	"net/netip"

	owlredis "owl-common/redis"
)

// AddrCtx 从 IoTStreamMessage.DeviceAddr 派生的字符串身份上下文。
// 字段为 INET CIDR 文本（"fd00:0:3::/48"）或 /128 host 文本。
type AddrCtx struct {
	Addr       netip.Addr
	DeviceAddr string
	TenantPref string
	BranchPref string
	UnitPref   string
	RoomPref   string
	BedPref    string
}

func AddrCtxFromMsg(m *owlredis.IoTStreamMessage) AddrCtx {
	if m == nil {
		return AddrCtx{}
	}
	return AddrCtxFromAddr(m.DeviceAddr)
}

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
