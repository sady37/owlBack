package belief

// mm.go — §2 MM 几何原语（[[mm_relationship_matrix]]）。cell 枚举一次产三标量，belief 只消费。
// covers=雷达覆盖床比例(连续)；onbed=垫是否绑此床(二元，垫严格绑一床)；overlap=雷达视野∩垫足迹实际交叠
// （非 min(covers,onbed) 上界代理）。三标量由 adapter 经 cell 枚举填，belief 不自算几何。

// BedGeom 单床的三标量（§2）。M 只存这三个。
type BedGeom struct {
	Covers  float64 // covers(r,b)=|R(r,b)|/|C_b| ∈[0,1]
	Onbed   float64 // onbed(s,b)=|P(s,b)|/|C_b| ∈{0,1}
	Overlap float64 // o_b(r,s)=|R∩P|/|C_b| ∈[0,1]（§F C1：≡ §4 的 o_j）
}

// kappaColdStart κ^(0)_{r,s_j}=covers·onbed（§3 分级冷启）。几何只播种初值，时间在线修正（coupling.go）。
func (g BedGeom) kappaColdStart() float64 { return g.Covers * g.Onbed }
