package belief

// state.go — S 轴隐状态（DBN-Zone-Room §1，P5 全空间 9 态）。与 Tsensor belief 的 S 轴/T_S 对齐，
// §9 三态对照才有意义。阶段 1 骨架只用 S 轴定义 + Vector 边缘读出；emission（likelihood/temper）阶段 2 补。

// State 全空间区域占用 × {直立,倒地}；SFallen 唯一 fire Fall 的近吸收态。
type State int

const (
	SEmpty State = iota
	SBed
	SSit
	SOpenFloor
	SBath
	SFallen
	SBlindRest
	SBlindOpen
	SLeft
)

const numStates = 9

var stateLabel = [numStates]string{
	"Empty", "Bed", "Sit", "Open-Floor", "Bath",
	"Floor-Fallen", "Blind-Rest", "Blind-Open", "Left-via-Door",
}

func (s State) String() string { return stateLabel[s] }

// Vector 是 Δ(S) 上的概率向量，Σ=1。
type Vector [numStates]float64

// normalize 归一化为分布；和为 0（理论不可达）退均匀。
func (v *Vector) normalize() {
	sum := 0.0
	for _, p := range v {
		sum += p
	}
	if sum <= 0 {
		for i := range v {
			v[i] = 1.0 / numStates
		}
		return
	}
	for i := range v {
		v[i] /= sum
	}
}
