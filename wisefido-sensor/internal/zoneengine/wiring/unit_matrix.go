// unit_matrix.go — MM per-unit 空间关系方阵的 maintainer。
//
// 从 SpatialCache(前缀：contains/onbed) + Geometry(covers 几何) Build 出 *spatial.Matrix。
// 只物化**静态关系**，不落库、不微服务。失效=启动 BuildAll + config:card commit 按 scope 精准失效（受影响
// device/room/card → mask 到 /80 unit 只重建那些 unit；无 scope 兜底全清）。无定时刷新。
// 注：covers 几何源自 server room_visual_layout（用户画的 layout），非 firmware 状态，故触发点是
// layout/binding 的 DB 提交（config:card publish），不是雷达固件回 ack。
// 运行时学习（15s 同步把候选 0.5 推向 1）不回写本矩阵——那是 DBN 后验/独立 learned 层的事。
//
// 当前 shadow：built + 可查询，尚未接 alarm 决策路径（消费留 #5 多床房 / #7 DBN 映射）。

package wiring

import (
	"context"
	"net/netip"
	"strings"
	"sync"

	"owl-common/spatial"

	"go.uber.org/zap"
)

// Geometry MM covers 几何源。roomengine.Engine 天然满足；结构化接口避免 wiring→roomengine import 环。
type Geometry interface {
	// RadarBedReachCount 数本 radar 边界内有几张 canvas bed（m）。
	RadarBedReachCount(deviceAddr string) int
}

// MatrixCache per-unit 关系方阵 maintainer。
type MatrixCache struct {
	spatial *SpatialCache
	geo     Geometry
	logger  *zap.Logger

	mu    sync.RWMutex
	units map[netip.Prefix]*spatial.Matrix
}

func NewMatrixCache(sp *SpatialCache, geo Geometry, logger *zap.Logger) *MatrixCache {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &MatrixCache{
		spatial: sp,
		geo:     geo,
		logger:  logger,
		units:   make(map[netip.Prefix]*spatial.Matrix),
	}
}

// MatrixFor 取 unit 方阵；miss → 即时 Build 并缓存。
func (mc *MatrixCache) MatrixFor(unit netip.Prefix) *spatial.Matrix {
	if mc == nil {
		return nil
	}
	mc.mu.RLock()
	mx, ok := mc.units[unit]
	mc.mu.RUnlock()
	if ok {
		return mx
	}
	mx = mc.buildUnit(unit)
	if mx == nil {
		return nil
	}
	mc.mu.Lock()
	mc.units[unit] = mx
	mc.mu.Unlock()
	return mx
}

// BuildAll 启动预热：扫 SpatialCache 已知 unit 逐个 Build。
func (mc *MatrixCache) BuildAll(ctx context.Context) {
	if mc == nil || mc.spatial == nil {
		return
	}
	_ = mc.spatial.LoadAll(ctx) // 幂等：确保 unit 列表已装载
	mc.spatial.mu.RLock()
	unitIDs := make([]netip.Prefix, 0, len(mc.spatial.units))
	for u := range mc.spatial.units {
		unitIDs = append(unitIDs, u)
	}
	mc.spatial.mu.RUnlock()

	built, edges := 0, 0
	for _, u := range unitIDs {
		if mx := mc.buildUnit(u); mx != nil {
			mc.mu.Lock()
			mc.units[u] = mx
			mc.mu.Unlock()
			built++
			edges += countEdges(mx)
		}
	}
	mc.logger.Info("MM matrix built", zap.Int("units", built), zap.Int("nonzero_edges", edges))
}

// InvalidateAll 清空所有 unit 方阵（无 scope 信息时兜底）；下次 MatrixFor 时 lazy 重建。
func (mc *MatrixCache) InvalidateAll() {
	if mc == nil {
		return
	}
	mc.mu.Lock()
	mc.units = make(map[netip.Prefix]*spatial.Matrix)
	mc.mu.Unlock()
}

// InvalidateScope 按受影响 scope（device_addr host/CIDR 或 room/card CIDR）精准失效：各自 mask 到 /80 unit，
// 只删那些 unit（下次 MatrixFor lazy 重建该 unit 下全部 radar covers）。空 scope → 调用方兜底 InvalidateAll。
func (mc *MatrixCache) InvalidateScope(prefixes []string) {
	if mc == nil {
		return
	}
	mc.mu.Lock()
	defer mc.mu.Unlock()
	for _, ps := range prefixes {
		if u, ok := unitOfString(ps); ok {
			delete(mc.units, u)
		}
	}
}

// unitOfString 把 device_addr(host 或 /128) 或 room/card CIDR 解析到所属 /80 unit。
func unitOfString(s string) (netip.Prefix, bool) {
	if p, err := netip.ParsePrefix(s); err == nil {
		u := maskToUnit(p)
		return u, u.IsValid()
	}
	if a, err := netip.ParseAddr(s); err == nil {
		u := maskToUnit(netip.PrefixFrom(a, a.BitLen()))
		return u, u.IsValid()
	}
	return netip.Prefix{}, false
}

// bedResolveConf radar 床事件归床的 covers 解析阈：单床房 covers=1.0≥阈 → 解析；多床候选 0.5<阈 →
// 留空（退回 radar 自己 /96）。设在 0.5 与 1.0 之间，使 DBN 把候选推到 0.95 后自动越阈激活多床房路由。
const bedResolveConf float32 = 0.8

// ResolveBed satisfy zoneengine.BedResolver — 按 radar covers 置信解析它床事件该归的床 /96，
// 同时返回该 covers 置信（喂 bed_bayesian radar 簇加权）。取 BedsCoveredBy 的最高置信床：
// ≥阈且严格压过次高 → 解析（单床房 1.0 / DBN 后 0.95），返回 (床, 置信)；多床候选未定 → ("",0)；
// 无覆盖（mount 未配/不可达）→ 退回 SpatialCache 单床房 parity，权重 1.0（无 covers 信息=不打折，不回归）。
func (mc *MatrixCache) ResolveBed(radarAddr, roomPref string) (string, float64) {
	if mc == nil {
		return "", 0
	}
	rp, err := netip.ParsePrefix(roomPref)
	if err != nil {
		return "", 0
	}
	if ra, err := netip.ParseAddr(radarAddr); err == nil {
		if mx := mc.MatrixFor(maskToUnit(rp)); mx != nil {
			var best, second spatial.Source
			for _, s := range mx.BedsCoveredBy(netip.PrefixFrom(ra, 128)) {
				if s.Conf > best.Conf {
					second, best = best, s
				} else if s.Conf > second.Conf {
					second = s
				}
			}
			if best.Conf >= bedResolveConf && best.Conf > second.Conf {
				return best.Prefix.String(), float64(best.Conf)
			}
			if best.Conf > 0 {
				return "", 0 // 多床候选 0.5/0.5 或并列 1.0 → 几何分不清，留 DBN
			}
		}
	}
	if mc.spatial != nil {
		return mc.spatial.ResolveSingleBed(roomPref), 1.0
	}
	return "", 0
}

// MatrixDump debug：一个 unit 方阵的可读快照（节点 + 非零有向边）。
type MatrixDump struct {
	Unit  string       `json:"unit"`
	Nodes []MatrixNode `json:"nodes"`
	Edges []MatrixEdge `json:"edges"`
}

type MatrixNode struct {
	Prefix string `json:"prefix"`
	Kind   string `json:"kind"`
}

type MatrixEdge struct {
	From string  `json:"from"`
	To   string  `json:"to"`
	Rel  string  `json:"rel"`
	Conf float64 `json:"conf"`
}

// Dump 导出 unit 方阵（debug HTTP 用）。列出节点 + 所有非零有向边（含 covers 候选/samebed 派生）。
func (mc *MatrixCache) Dump(unit netip.Prefix) *MatrixDump {
	mx := mc.MatrixFor(unit)
	if mx == nil {
		return nil
	}
	d := &MatrixDump{Unit: unit.String()}
	for _, o := range mx.Obj {
		d.Nodes = append(d.Nodes, MatrixNode{Prefix: o.Prefix.String(), Kind: kindName(o.Kind)})
	}
	for i := range mx.M {
		for j := range mx.M[i] {
			if i == j || mx.M[i][j] <= 0 {
				continue
			}
			d.Edges = append(d.Edges, MatrixEdge{
				From: mx.Obj[i].Prefix.String(),
				To:   mx.Obj[j].Prefix.String(),
				Rel:  relName(mx.Obj[i].Kind, mx.Obj[j].Kind),
				Conf: float64(mx.M[i][j]),
			})
		}
	}
	return d
}

func kindName(k spatial.ObjectKind) string {
	switch k {
	case spatial.ObjRoom:
		return "room"
	case spatial.ObjBed:
		return "bed"
	case spatial.ObjRadar:
		return "radar"
	case spatial.ObjSleepad:
		return "sleepad"
	default:
		return "device"
	}
}

// relName 由 kind 对推关系名（与矩阵"类型由 kind 对隐含"一致）。
func relName(from, to spatial.ObjectKind) string {
	switch {
	case from == spatial.ObjRadar && to == spatial.ObjSleepad, from == spatial.ObjSleepad && to == spatial.ObjRadar:
		return "samebed"
	case from == spatial.ObjRadar && to == spatial.ObjBed:
		return "covers"
	case from.IsDevice() && to == spatial.ObjBed:
		return "onbed"
	case to == spatial.ObjBed || to.IsDevice() || to == spatial.ObjRoom:
		return "contains"
	default:
		return "rel"
	}
}

// buildUnit 从 SpatialCache UnitData + Geometry 构造一个 unit 方阵。
func (mc *MatrixCache) buildUnit(unit netip.Prefix) *spatial.Matrix {
	ud := mc.spatial.ensureUnit(unit)
	if ud == nil {
		return nil
	}
	var objs []spatial.Object
	for p := range ud.Rooms {
		objs = append(objs, spatial.Object{Prefix: p, Kind: spatial.ObjRoom})
	}
	for p := range ud.Beds {
		objs = append(objs, spatial.Object{Prefix: p, Kind: spatial.ObjBed})
	}
	for addr, dm := range ud.Devices {
		objs = append(objs, spatial.Object{Prefix: netip.PrefixFrom(addr, 128), Kind: deviceKind(dm.DeviceType)})
	}
	if len(objs) == 0 {
		return nil
	}
	return spatial.Build(unit, objs, mc.makeCovers(ud))
}

// makeCovers covers(radar,bed) = min(m/n, 1)：m=radar 边界内 canvas bed 数(Geometry)，
// n=同房 bed 前缀数(SpatialCache DB beds 表)。单床房 m=n=1→1.0 确定；m=1,n=2→0.5 候选。
// 同房判定用 netip /88 掩码比较（可靠，不碰 DB 字符串）。
//
// 为何必须 cap 在 1：long-sofa 在 layout 里以 typeName "Bed"/"MonitorBed" 进 cfg.Beds，几何层面
// 与真床**分不出**，但 DB beds 表**没有** long-sofa。故 m(canvas，含 sofa) 可 > n(DB，仅真床)，
// 比值会 >1。cap 后 sofa 污染只会把候选 saturate 到并列 1.0 → ResolveBed 支配规则判不出 → 留空
// 退回 radar /96，**绝不误归床**（安全）。彻底分 sofa 需前端给 long-sofa 独立 typeName，不在此。
func (mc *MatrixCache) makeCovers(ud *UnitData) func(radar, bed netip.Prefix) float32 {
	return func(radarPfx, bedPfx netip.Prefix) float32 {
		room := roomOf(bedPfx)
		if !room.IsValid() || roomOf(radarPfx) != room {
			return 0
		}
		m := mc.geo.RadarBedReachCount(radarPfx.Addr().String())
		if m == 0 {
			return 0
		}
		n := 0
		for bp := range ud.Beds {
			if roomOf(bp) == room {
				n++
			}
		}
		if n == 0 {
			return 0
		}
		v := float32(m) / float32(n)
		if v > 1 {
			v = 1
		}
		return v
	}
}

func deviceKind(deviceType string) spatial.ObjectKind {
	switch {
	case strings.EqualFold(deviceType, "Radar"):
		return spatial.ObjRadar
	case strings.EqualFold(deviceType, "Sleepad"):
		return spatial.ObjSleepad
	default:
		return spatial.ObjDevice
	}
}

// roomOf 任意 /96+ 前缀截到 /88 room（netip 掩码，与 DB 文本无关）。
func roomOf(p netip.Prefix) netip.Prefix {
	if !p.IsValid() || p.Bits() < 88 {
		return netip.Prefix{}
	}
	return netip.PrefixFrom(p.Addr(), 88).Masked()
}

func countEdges(mx *spatial.Matrix) int {
	n := 0
	for i := range mx.M {
		for j := range mx.M[i] {
			if i != j && mx.M[i][j] > 0 {
				n++
			}
		}
	}
	return n
}
