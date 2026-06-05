package spatial

// relmatrix.go — per-unit 空间关系方阵（MM）。
//
// M[i][j] ∈ [0,1] = Obj[i]→Obj[j] 这条关系的置信。关系**类型由 kind 对隐含**（任意有序
// kind 对最多一种关系：room→bed=contains / device→bed=onbed / radar→bed=covers /
// radar→sleepad=samebed），所以格子只存强度，不存类型：
//   1.0 = 确定（前缀类 contains/onbed，或几何无歧义 covers）
//   (0,1) = 候选（几何有歧义的 covers，值=几何先验，由 covers 谓词给）
//   0   = 无关系
// 方向给不对称：M[room][bed]=1 而 M[bed][room]=0；M[sleepad][bed]=M[bed][sleepad]=1
// （sleepad∈bed 是 onbed，bed⊃sleepad 是 contains，两向都确定）。
//
// 纪律：只存**静态关系**（从 layout/前缀重建）。运行时学习（15s 同步把候选 0.5 推向 1）
// 不回写本矩阵——那是 DBN 后验/独立 learned 层的事。covers 几何由 caller 注入，spatial 不依赖 radar 包。

import "net/netip"

type ObjectKind uint8

const (
	ObjRoom ObjectKind = iota
	ObjBed
	ObjRadar
	ObjSleepad
	ObjDevice
)

func (k ObjectKind) IsDevice() bool { return k == ObjRadar || k == ObjSleepad || k == ObjDevice }

// Object = L[i] = C[i]，方阵的行/列实体。只存身份，几何不入矩阵（covers 谓词外部算）。
type Object struct {
	Prefix netip.Prefix
	Kind   ObjectKind
}

const (
	ConfNone    float32 = 0
	ConfCertain float32 = 1
)

// Matrix per-unit 关系方阵。
type Matrix struct {
	Unit netip.Prefix
	Obj  []Object    // L[] = C[]，行列同一序
	M    [][]float32 // M[i][j] = Obj[i]→Obj[j] 置信
	idx  map[netip.Prefix]int
}

// Source bed 的一个融合源（device + 它对该 bed 的置信权重）。
type Source struct {
	Prefix netip.Prefix
	Conf   float32
}

// RelOf = 外部 helper：算一对实体 a→b 的关系置信（M[i][j] 的值）。矩阵本身不含逻辑，只存结果。
// covers(radar,bed) 返回几何置信（确定盖=1 / 候选=几何先验如 0.5 / 没盖=0），由 caller 用
// radarutils 实现注入；nil → 无 covers。samebed 是派生，不在此，由 Build 后置 deriveSameBed 填。
func RelOf(a, b Object, covers func(radar, bed netip.Prefix) float32) float32 {
	if a.Prefix == b.Prefix {
		return ConfCertain
	}
	// a ⊃ b：contains（room→bed→device，含 bed⊃device），前缀确定
	if a.Prefix.Bits() < b.Prefix.Bits() && a.Prefix.Contains(b.Prefix.Addr()) {
		return ConfCertain
	}
	// a ∈ b 且 a=device、b=bed：onbed，前缀确定
	if a.Kind.IsDevice() && b.Kind == ObjBed && b.Prefix.Contains(a.Prefix.Addr()) {
		return ConfCertain
	}
	// a=radar、b=bed：covers，几何（可候选）。前缀绑定已在上面命中，这里只剩房绑雷达的几何判定
	if a.Kind == ObjRadar && b.Kind == ObjBed && covers != nil {
		return covers(a.Prefix, b.Prefix)
	}
	return ConfNone
}

// Build 构造方阵：每对 (i,j) 调 RelOf 填 M[i][j]，再后置派生 samebed。
func Build(unit netip.Prefix, objs []Object, covers func(radar, bed netip.Prefix) float32) *Matrix {
	n := len(objs)
	mx := &Matrix{Unit: unit, Obj: objs, M: make([][]float32, n), idx: make(map[netip.Prefix]int, n)}
	for i, o := range objs {
		mx.M[i] = make([]float32, n)
		mx.idx[o.Prefix] = i
	}
	for i := range objs {
		for j := range objs {
			mx.M[i][j] = RelOf(objs[i], objs[j], covers)
		}
	}
	mx.deriveSameBed()
	return mx
}

// deriveSameBed radar r 与 sleepad s 同床置信 = max_bed min(covers(r,b), onbed(s,b))。对称写双向。
// covers 候选 0.5 ⊗ onbed 确定 1 → samebed 0.5（不确定性沿合取传播，正好作 DBN 先验）。
func (mx *Matrix) deriveSameBed() {
	for r := range mx.Obj {
		if mx.Obj[r].Kind != ObjRadar {
			continue
		}
		for s := range mx.Obj {
			if mx.Obj[s].Kind != ObjSleepad {
				continue
			}
			var best float32
			for b := range mx.Obj {
				if mx.Obj[b].Kind != ObjBed {
					continue
				}
				if c := min32(mx.M[r][b], mx.M[s][b]); c > best {
					best = c
				}
			}
			mx.M[r][s] = best
			mx.M[s][r] = best
		}
	}
}

func min32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

// ---- 查询 API：散落逻辑都变成查矩阵 ----

// Conf 返回 a→b 的关系置信；任一不在矩阵 → 0。
func (mx *Matrix) Conf(a, b netip.Prefix) float32 {
	i, ok1 := mx.idx[a]
	j, ok2 := mx.idx[b]
	if !ok1 || !ok2 {
		return ConfNone
	}
	return mx.M[i][j]
}

// SameBedConf radar↔sleepad 同床置信（0=不同床，0.5=候选，1=确定）。
func (mx *Matrix) SameBedConf(a, b netip.Prefix) float32 { return mx.Conf(a, b) }

// FuseGroup 返回监视 bed 的所有 device 及其置信权重（covers 或 onbed > 0）——bed_bayesian/DBN
// 直接拿权重做融合。候选源(0.5)也算融合源，权重即其先验。
func (mx *Matrix) FuseGroup(bed netip.Prefix) []Source {
	j, ok := mx.idx[bed]
	if !ok {
		return nil
	}
	var out []Source
	for i := range mx.Obj {
		if mx.Obj[i].Kind.IsDevice() && mx.M[i][j] > ConfNone {
			out = append(out, Source{Prefix: mx.Obj[i].Prefix, Conf: mx.M[i][j]})
		}
	}
	return out
}

// BedsCoveredBy radar 覆盖的所有 bed（含候选），带置信。
func (mx *Matrix) BedsCoveredBy(radar netip.Prefix) []Source {
	i, ok := mx.idx[radar]
	if !ok {
		return nil
	}
	var out []Source
	for j := range mx.Obj {
		if mx.Obj[j].Kind == ObjBed && mx.M[i][j] > ConfNone {
			out = append(out, Source{Prefix: mx.Obj[j].Prefix, Conf: mx.M[i][j]})
		}
	}
	return out
}
