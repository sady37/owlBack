package card

import "net/netip"

// Entity 是 spatial 子项的统一表示。
//
// prefix mask 自描述层级：
//
//	/48  tenant / /56 branch / /64 site
//	/80  unit
//	/88  card | room
//	/96  bed
//	/128 device
//
// 同 mask 内细分由 Kind 区分（rooms.room_type 0/1/2、device_type 等）。
type Entity struct {
	Prefix netip.Prefix
	Name   string
	Kind   int
}

func (e Entity) IsUnit() bool   { return e.Prefix.Bits() == 80 }
func (e Entity) IsRoom() bool   { return e.Prefix.Bits() == 88 }
func (e Entity) IsBed() bool    { return e.Prefix.Bits() == 96 }
func (e Entity) IsDevice() bool { return e.Prefix.Bits() == 128 }
