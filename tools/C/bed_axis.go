package belief

// bed_axis.go — B 轴：床占用隐变量 B^j ∈ {vac, occ}，与 S 轴正交（DBN-Zone-Room §1）。
// T_B 转移核（§6 + 第三部分 §C）：观测可用性 ρ 门控自持。
//   在线 ρ=1：K^obs（高自持，上下床稀有，防抖）。
//   离线 ρ=0：K^unobs_λ（按 §C 箭头记法单向泄漏 occ→vac，vac 吸收）。
// 关键不等式 ε≪λ：陈旧 occ 在 fall confirm 前快速蒸发（治本 cd2b 漏报 + 替所有 staleness/TTL 补丁）。
//
// 床占用二态。索引固定（joint.go bmask 第 j 位 = B^j）。
type BedState int

const (
	BVac BedState = 0 // 床空
	BOcc BedState = 1 // 床占用
)

const numBedStates = 2

// bedAxisParams T_B 的两套核参数。开发阶段为常量；标定见 feedback-p6C §5（μ=ε~1e-2 对称、λ 半衰期 10–15s）。
// 这里只锚“形态/约束方向”，不锁单 case 反推的精确值（铁律：定形态、不定参数）。
type bedAxisParams struct {
	// ── K^obs（ρ=1 在线）──
	epsilon float64 // occ→vac 在线翻转率（= 1 − occ自持）；上下床稀有 → ε 小
	mu      float64 // vac→occ 在线翻转率；B1：无非对称证据 → 默认 μ=ε（对称）
	// ── K^unobs_λ（ρ=0 离线，§C 箭头记法单向泄漏）──
	lambda float64 // occ→vac 离线泄漏率；约束 ε≪λ 且半衰期 < 离线判定窗(~30s)
}

// defaultBedAxisParams 形态默认：μ=ε（对称，B1），ε≪λ（§C）。
// 数值是占位初值（量级锚 feedback-p6C §5），标定阶段由 oracle 收紧——骨架阶段只验结构与不等式。
func defaultBedAxisParams() bedAxisParams {
	const eps = 1e-2 // ε 量级（413s 仅 1 翻转，feedback-p6C §5）
	return bedAxisParams{
		epsilon: eps,
		mu:      eps,      // B1：μ=ε 对称默认
		lambda:  0.05,     // λ≫ε；1Hz 下半衰期 ≈ ln2/0.05 ≈ 14s < 30s 离线窗（§C / D-2）
	}
}

// bedKernel 2×2 转移核，kernel[from][to] = P(B_t=to | B_{t-1}=from)。行随机（每行 Σ=1）。
type bedKernel [numBedStates][numBedStates]float64

// kObs 在线核 K^obs（ρ=1）：
//
//	occ→occ = 1−ε   occ→vac = ε
//	vac→occ = μ     vac→vac = 1−μ
func (p bedAxisParams) kObs() bedKernel {
	return bedKernel{
		BOcc: {BVac: p.epsilon, BOcc: 1 - p.epsilon},
		BVac: {BVac: 1 - p.mu, BOcc: p.mu},
	}
}

// kUnobs 离线核 K^unobs_λ（ρ=0，§C 箭头记法单向泄漏，定义 B：occ 泄漏 vac 吸收）：
//
//	occ→vac = λ     occ→occ = 1−λ
//	vac→vac = 1     vac→occ = 0      ← vac 吸收（关键：空房离线不被弛豫成伪占用，cd2b 根方向）
//
// 注意：这是单向泄漏，**不是**对称弛豫。vac→occ 恒 0，故陈旧 occ 单向蒸发、空房保持空。
func (p bedAxisParams) kUnobs() bedKernel {
	return bedKernel{
		BOcc: {BVac: p.lambda, BOcc: 1 - p.lambda},
		BVac: {BVac: 1.0, BOcc: 0.0}, // vac 吸收态
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
