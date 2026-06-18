package belief

// Matrix 转移阵 A，A[i][j]=P(s_t=j | s_{t-1}=i)。白盒——没有一个数字不可解释（§5）。
type Matrix [numStates][numStates]float64

// transitionPropensity 转移倾向（行未归一化，init 时按行归一）。空间运动：人在 unit 区域间怎么移动。
// 数字是"相对倾向"非概率，方便人手编辑。★=进 Fallen 条目，刻意取极小（0.05–0.5）——
// 直立态自发倒地不存在/罕见，A 只保证 Fallen 可达让观测能放大（diag(w)·b 只放大已有非零分量），
// 造跌倒的活交给观测。**安全性反转**：Blind* 久缺**应**从沉默 ramp Fallen（lost-fall 本义），
// 故 Blind→Fallen 给**极小种子（0.5）**够 Observe 放大即可——ramp 靠每盲 tick 的 ObsNoDetect 放大（×1.6 复合）
// × Fallen 自持 0.99 + Decider 90s confirm + dwell 生存尾，**绝不靠 A 自漂**（种子过大→纯 Predict 凭空造跌倒=FP）。
// Blind* **无自漂 Left 逃生阀**（§84 步5）：Blind→Left 只由**证据**驱动——neighbor handoff（filter
// GateBlindRow 据 ρ 整流 F→L）或离房趋势（lost 反算）。静态自漂 Left 会让开阔区真摔 lost 在 D 窗到点前
// 被 ramp 进 Left 误 drop（漏报）；改证据驱动后，无 handoff/无离房趋势的 blind 留到 D 满走 deadline fire。
//
//	→E    →Bed  →Sit  →Open →Bath  →Fa★  →BR   →BO   →Lf
var transitionPropensity = Matrix{
	SEmpty:     {92, 1, 1, 5, 1, 0, 0, 0, 0},         // 多数维持空；进入靠观测驱动
	SBed:       {0, 80, 1, 7, 0.5, 0.05, 10, 0, 1.2}, // ★ 床上自发倒 0.05（极小）；起身→Open / 床边→BlindRest
	SSit:       {0, 1, 80, 10, 0, 0.3, 3, 3, 2.7},    // 起身→Open；进盲二义对半（Rest 椅边滑坐 / Open 起身走开），后续观测裁
	SOpenFloor: {0, 5, 6, 65, 4, 0.5, 0, 10, 6},      // ★ 走动→倒 0.5；走动中→BlindOpen / 近门→Left
	SBath:      {0, 0, 0, 8, 78, 0.3, 10, 0, 3.7},    // 卫浴盲→BlindRest
	SFallen:    {0, 0, 0, 0.7, 0, 99, 0, 0.3, 0},     // 近吸收：倒地不自愈；只有返回/走动等观测能压下
	SBlindRest: {0, 8, 0, 1, 6, 0.5, 84, 0, 0},       // 重现→Bed/Bath；★→Fallen 0.5 种子；无 Left 自漂(留 D 满 deadline fire 或 handoff/离房趋势 cancel)
	SBlindOpen: {0, 0, 0, 8, 0, 0.5, 0, 71, 0},       // 重现→Open；★→Fallen 0.5 种子；→Left 证据驱动(handoff/离房趋势)非自漂(§84 步5)
	SLeft:      {85, 0, 0, 0, 0, 0, 0, 0, 15},        // 收敛回 Empty
}

// Model 持有归一化后的 A 与初始 prior。无状态，可全局共享。
type Model struct {
	A     Matrix
	Prior Vector
}

// DefaultModel 返回 v1 标定模型（行归一化 A + Empty-主导 prior）。
func DefaultModel() *Model {
	m := &Model{A: rowNormalized(transitionPropensity)}
	// prior：房间默认空，少量散布到 Open/Left/Bed。
	m.Prior = Vector{SEmpty: 0.85, SOpenFloor: 0.08, SLeft: 0.04, SBed: 0.03}
	m.Prior.normalize()
	return m
}

func rowNormalized(p Matrix) Matrix {
	var a Matrix
	for i := 0; i < numStates; i++ {
		sum := 0.0
		for j := 0; j < numStates; j++ {
			sum += p[i][j]
		}
		if sum <= 0 {
			a[i][i] = 1.0 // 退化行：自持（不该发生，防御 0 行）
			continue
		}
		for j := 0; j < numStates; j++ {
			a[i][j] = p[i][j] / sum
		}
	}
	return a
}
