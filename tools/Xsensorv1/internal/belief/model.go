package belief

// model.go — S 轴转移 T_S（DBN-Zone-Room §6：P5 的 9×9 propensity，不被 B' 调制）。
// 与 Tsensor belief/model.go 对齐拷贝，保证 Xsensorv1 与 Tsensor 的 S 轴一致。

// Matrix 转移阵 A，A[i][j]=P(s_t=j | s_{t-1}=i)。
type Matrix [numStates][numStates]float64

// transitionPropensity 转移倾向（行未归一，init 按行归一）。★=进 Fallen 条目，刻意极小——
// 直立态自发倒地罕见，A 只保证 Fallen 可达让观测放大；造跌倒交给观测，绝不靠 A 自漂（种子过大→凭空 FP）。
// BlindRest 无 Left 逃生阀（rest 区非门）；BlindOpen 有（reachable-exit 放大）。
//
//	→E    →Bed  →Sit  →Open →Bath  →Fa★  →BR   →BO   →Lf
var transitionPropensity = Matrix{
	SEmpty:     {92, 1, 1, 5, 1, 0, 0, 0, 0},
	SBed:       {0, 80, 1, 7, 0.5, 0.05, 10, 0, 1.2},
	SSit:       {0, 1, 80, 10, 0, 0.3, 3, 3, 2.7},
	SOpenFloor: {0, 5, 6, 65, 4, 0.5, 0, 10, 6},
	SBath:      {0, 0, 0, 8, 78, 0.3, 10, 0, 3.7},
	SFallen:    {0, 0, 0, 0.7, 0, 99, 0, 0.3, 0}, // 近吸收：倒地不自愈
	SBlindRest: {0, 8, 0, 1, 6, 0.5, 84, 0, 0},   // 无 Left 逃生阀
	SBlindOpen: {0, 0, 0, 8, 0, 0.5, 0, 71, 12},  // →Left 逃生阀
	SLeft:      {85, 0, 0, 0, 0, 0, 0, 0, 15},
}

// Model 持有归一化后的 A 与初始 prior。无状态，可共享。
type Model struct {
	A     Matrix
	Prior Vector
}

// DefaultModel 返回 v1 标定模型（行归一化 A + Empty-主导 prior）。
func DefaultModel() *Model {
	m := &Model{A: rowNormalized(transitionPropensity)}
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
			a[i][i] = 1.0 // 防御 0 行：自持
			continue
		}
		for j := 0; j < numStates; j++ {
			a[i][j] = p[i][j] / sum
		}
	}
	return a
}
