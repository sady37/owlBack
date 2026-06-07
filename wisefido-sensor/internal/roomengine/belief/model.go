package belief

// Matrix 转移阵 A，A[i][j]=P(s_t=j | s_{t-1}=i)。白盒——没有一个数字不可解释（§5）。
type Matrix [numStates][numStates]float64

// transitionPropensity 转移倾向（行未归一化，init 时按行归一）。
// 数字是"相对倾向"非概率，方便人手编辑。★=进 Fallen 条目。
//
// 关键设计：★ 列刻意取极小值。进 Fallen 的"灵敏度"来自观测似然（ObsFirmwareFall /
// pose=fallen@open_floor / pose-lying@open_floor 等**正向** pose 证据），不来自时间转移。
// （P2.1 已删 ObsKinematics：z↓ 是环境噪声,不当 fall 证据,R5。）Fallen 是吸收态（自持 92%），
// 若 A 的 →Fallen 给大值，纯 Predict（缺证据期）会把信念慢慢渗进 Fallen ＝ 从沉默里凭空造跌倒
// ＝ 正是 CABB/John.Y 误报。所以 A 只保证 Fallen 可达 + 不自愈，造跌倒的活全交给观测。
//
//	→E    →BL   →BR   →Si   →SW   →Fa    →Tr   →Lf   →Ar
var transitionPropensity = Matrix{
	SEmpty:       {90, 0, 0, 0, 1, 0, 2, 0, 1},     // 多数维持空；进入靠观测驱动
	SBedLying:    {0, 85, 10, 0, 0, 0.05, 5, 0, 0}, // ★ 床上跌床 0.05
	SBedRestless: {0, 40, 47, 3, 2, 0.05, 8, 0, 0}, // ★ 0.05
	SSit:         {0, 1, 1, 82, 8, 0.1, 8, 0, 0},   // ★ 坐→倒 0.1
	SStandWalk:   {0, 0, 0, 6, 82, 0.2, 9, 3, 0},   // ★ 站立/行走→倒 0.2（最主要跌倒入口，仍极小）
	SFallen:      {0, 0, 0, 0, 0.3, 99, 0.7, 0, 0}, // 近吸收：倒地不自愈；缺证据不衰减（人还在地上），
	//                                                  只有"返回/走动"等主动反证（观测似然）能压下来
	STransition: {6, 8, 5, 12, 30, 0.3, 33, 4, 2}, // 过渡枢纽
	SLeft:       {85, 0, 0, 0, 0, 0, 0, 15, 0},    // 收敛回 Empty
	SArtifact:   {30, 0, 0, 0, 2, 0, 3, 0, 65},    // ghost 逐步消散到 Empty
}

// Model 持有归一化后的 A 与初始 prior。无状态，可全局共享。
type Model struct {
	A     Matrix
	Prior Vector
}

// DefaultModel 返回 v1 标定模型（行归一化 A + Empty-主导 prior）。
func DefaultModel() *Model {
	m := &Model{A: rowNormalized(transitionPropensity)}
	// prior：房间默认空，少量散布到过渡/伪迹吸收不确定。
	m.Prior = Vector{SEmpty: 0.80, STransition: 0.08, SStandWalk: 0.05, SArtifact: 0.05, SLeft: 0.02}
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
