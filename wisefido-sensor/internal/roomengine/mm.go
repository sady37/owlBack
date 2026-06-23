package roomengine

// mm.go — 房级静态空间关系方阵（MM）的 prior 权威（切片2：只 prior；30s learned overlay = 切片3，§9.2）。
//
// 复用 owl-common spatial.Build（与 wisefido-sensor MatrixCache 同源配方 unit_matrix.go）：
//   - onbed  = device/128 ∈ bed/96（前缀包含，确定）→ sleepad→床是**确定绑定**（接触式压一张床）。
//   - covers = radar 边界内 canvas 床数 m / 房床数 n（几何候选）→ radar→床是**不确定的"或"**（FoV 大看多床）。
//   - samebed(radar,sleepad) = max_bed min(covers, onbed)（派生 prior）。onbed 确定 0/1 → min 退场，
//     samebed 塌成 radar 那侧 covers（§8.1 模型纠正）。
//
// prefix 空间算 samebed，化掉 canvas 床 index 对齐。单床 n=1 → covers=1 → samebed=1（立即吸纳，零回归）；
// 多床 n=2,m=1 → 0.5 候选（吸纳读 prior 时 0.5<阈 → 不吸纳 → uncovered，§9.2④ FN-safe，靠切片3 learned 收敛）。

import (
	"net/netip"

	"owl-common/spatial"
)

// RoomBed DB beds 表一行投影：bed_id /96 prefix + slot + name（A~D，forensic）。
type RoomBed struct {
	Prefix netip.Prefix
	Slot   int
	Name   string
}

// RoomMM 单房静态关系方阵（samebed prior 权威）。内存态、bootstrap 一次建、layout 变重注册即重建（§8.4③）。
type RoomMM struct {
	mx *spatial.Matrix
}

// BuildRoomMM 照 unit_matrix.go 配方建房级静态 MM。
//
//	roomPrefix = room /88；beds = DB beds /96；radarAddrs/sleepadAddrs = 设备 /128。
//	radarBedReach(radarAddr) = 该 radar 边界内 canvas 床数 m（复用 Engine.RadarBedReachCount）；n = len(beds)。
//	无任何 bed 或无设备 → 返回 nil（SamebedConf 退化 0，吸纳侧自然不吸纳）。
func BuildRoomMM(roomPrefix netip.Prefix, beds []RoomBed, radarAddrs, sleepadAddrs []netip.Addr,
	radarBedReach func(radarAddr string) int) *RoomMM {
	if len(beds) == 0 || (len(radarAddrs) == 0 && len(sleepadAddrs) == 0) {
		return nil
	}
	objs := make([]spatial.Object, 0, 1+len(beds)+len(radarAddrs)+len(sleepadAddrs))
	objs = append(objs, spatial.Object{Prefix: roomPrefix, Kind: spatial.ObjRoom})
	for _, b := range beds {
		objs = append(objs, spatial.Object{Prefix: b.Prefix, Kind: spatial.ObjBed})
	}
	for _, a := range radarAddrs {
		objs = append(objs, spatial.Object{Prefix: netip.PrefixFrom(a, 128), Kind: spatial.ObjRadar})
	}
	for _, a := range sleepadAddrs {
		objs = append(objs, spatial.Object{Prefix: netip.PrefixFrom(a, 128), Kind: spatial.ObjSleepad})
	}
	n := len(beds)
	// covers(radar,bed)=min(m/n,1)（单源同 unit_matrix.go makeCovers：m=Geometry 边界内床数，n=房床数）。
	covers := func(radarPfx, _ netip.Prefix) float32 {
		m := radarBedReach(radarPfx.Addr().String())
		if m <= 0 {
			return 0
		}
		if v := float32(m) / float32(n); v < 1 {
			return v
		}
		return 1
	}
	return &RoomMM{mx: spatial.Build(roomPrefix, objs, covers)}
}

// SamebedConf radar/128 ↔ sleepad/128 同床 prior 置信（0=不同床 / 0.5=候选 / 1=确定）。
//
//	吸纳侧据此判：≥阈 ∧ 该 radar N-in-bed → 吸纳该 sleepad。nil MM → 0（不吸纳，uncovered，FN-safe 多报）。
func (m *RoomMM) SamebedConf(radarAddr, sleepadAddr netip.Addr) float32 {
	if m == nil || m.mx == nil {
		return 0
	}
	return m.mx.SameBedConf(netip.PrefixFrom(radarAddr, 128), netip.PrefixFrom(sleepadAddr, 128))
}
