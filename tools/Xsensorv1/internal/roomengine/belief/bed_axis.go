package belief

// bed_axis.go — B 轴：床占用隐变量 B^j ∈ {vac, occ}，与 S 轴正交（DBN-Zone-Room §1）。
// T_B 转移核（§6 + 第三部分 §C）：观测可用性 ρ 门控自持。ρ 为 config-static（§32 二态裁定：
// 设备在线 OR 没有，不建模中途掉线）——有 sleepad 房 ρ=1，无 sleepad（radar-only）房 ρ=0。
//   ρ=1（有 sleepad）：K^obs（高自持，上下床稀有，防抖）。
//   ρ=0（无 sleepad，radar-only）：K^unobs_λ（§C 箭头记法单向泄漏 occ→vac，vac 吸收）= 无接触证据
//     时 B 轴合理默认空、由 radar 经 §4 Ψ 涌现占用。
// ε≪λ（§C 机制保留，C 裁过）：occ→vac 比 vac→occ 快。**论证订正（§31/§32 C 共同认领）**：原「治本
// cd2b 漏报=陈旧 occ 离线蒸发」绑在误读的 cd2b-offline 上，作废——真实 cd2b sleepad 在线报 LeftBed，
// B 靠 LeftBed 发射直接推 vac，非靠 λ 蒸发。机制重解释为「config-static 无 sleepad 房 B 轴默认空+软化」。

// BedState 床占用二态。索引固定（joint.go bmask 第 j 位 = B^j）。
type BedState int

const (
	BVac BedState = 0 // 床空
	BOcc BedState = 1 // 床占用
)

const numBedStates = 2

// bedAxisParams T_B 的两套核参数。开发阶段锚形态/方向，标定见 feedback-p6C §5（μ=ε~1e-2、λ 半衰期 10–15s）。
type bedAxisParams struct {
	epsilon float64 // occ→vac 在线翻转率（上下床稀有 → ε 小）
	mu      float64 // vac→occ 在线翻转率（B1：μ=ε 对称默认）
	lambda  float64 // occ→vac 离线泄漏率（ε≪λ，半衰期 < 30s 离线窗）
}

// defaultBedAxisParams 形态默认：μ=ε（对称 B1），ε≪λ（§C 单向泄漏）。
func defaultBedAxisParams() bedAxisParams {
	const eps = 1e-2 // ε 量级（413s 仅 1 翻转，feedback-p6C §5）
	return bedAxisParams{
		epsilon: eps,
		mu:      eps,  // B1：μ=ε 对称默认
		lambda:  0.05, // λ≫ε；1Hz 下半衰期 ≈ ln2/0.05 ≈ 14s < 30s 离线窗（§C / D-2）
	}
}

// bedKernel 2×2 转移核，kernel[from][to] = P(B_t=to | B_{t-1}=from)。行随机（每行 Σ=1）。
type bedKernel [numBedStates][numBedStates]float64

// kObs 在线核 K^obs（ρ=1）。§6：occ→occ=1-ε, vac→occ=μ。
func (p bedAxisParams) kObs() bedKernel {
	return bedKernel{
		BOcc: {BVac: p.epsilon, BOcc: 1 - p.epsilon},
		BVac: {BVac: 1 - p.mu, BOcc: p.mu},
	}
}

// kUnobs 离线核 K^unobs_λ（ρ=0，§C 箭头记法单向泄漏，定义 B：occ 泄漏 vac 吸收）：
//
//	occ→vac = λ     occ→occ = 1−λ
//	vac→vac = 1     vac→occ = 0      ← vac 吸收（空房离线不被弛豫成伪占用，cd2b 根方向）
func (p bedAxisParams) kUnobs() bedKernel {
	return bedKernel{
		BOcc: {BVac: p.lambda, BOcc: 1 - p.lambda},
		BVac: {BVac: 1.0, BOcc: 0.0},
	}
}

// tBKernel 按观测可用性 ρ 选核。ρ=true（sleepad 在线）→ K^obs；ρ=false（离线）→ K^unobs_λ。
// DBN-Zone-Room §6：ρ_t = 1[s_j 在线]。骨架阶段 ρ 由调用方（filter）传入。
func (p bedAxisParams) tBKernel(online bool) bedKernel {
	if online {
		return p.kObs()
	}
	return p.kUnobs()
}

// logKernel 预存 log 域 2×2 核，filter Predict 内层复用，避免每步 log。
type logKernel [numBedStates][numBedStates]float64

func makeLogKernel(k bedKernel) logKernel {
	var lk logKernel
	for from := 0; from < numBedStates; from++ {
		for to := 0; to < numBedStates; to++ {
			lk[from][to] = logP(k[from][to])
		}
	}
	return lk
}
