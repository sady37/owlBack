package belief

// bed_axis.go — B 轴：床占用隐变量 B^j ∈ {vac,occ}，与 S 轴正交（DBN-Zone-Room §1）。
// T_B 按观测可用性 ρ 门控自持（§6 + 第三部分 §C）：
//   在线 ρ=1 → K^obs（高自持，上下床稀有防抖）；
//   离线 ρ=0 → K^unobs_λ（§C 箭头记法单向泄漏：occ→vac=λ，vac 吸收）。
// 不等式 ε≪λ 使陈旧 occ 在 fall confirm 前蒸发——治本 cd2b 漏报，替所有 staleness/TTL。

type BedState int

const (
	BVac BedState = 0 // 床空
	BOcc BedState = 1 // 床占用
)

const numBedStates = 2

// bedAxisParams T_B 两套核的参数。骨架阶段为占位形态值；精确标定见 feedback-p6C §5
// （μ=ε~1e-2 对称、λ 半衰期 10–15s）。铁律：定形态、不定参数。
type bedAxisParams struct {
	epsilon float64 // K^obs occ→vac 在线翻转率（上下床稀有 → 小）
	mu      float64 // K^obs vac→occ 在线翻转率（B1：无非对称证据 → μ=ε 对称）
	lambda  float64 // K^unobs occ→vac 离线泄漏率（ε≪λ 且半衰期 < ~30s 离线窗）
}

func defaultBedAxisParams() bedAxisParams {
	const eps = 1e-2 // 413s 仅 1 翻转（feedback-p6C §5）
	return bedAxisParams{
		epsilon: eps,
		mu:      eps,  // B1 对称默认
		lambda:  0.05, // 1Hz 半衰期 ≈ ln2/0.05 ≈ 14s < 30s（§C / D-2）
	}
}

// bedKernel 2×2 转移核，kernel[from][to] = P(B_t=to | B_{t-1}=from)，行随机。
type bedKernel [numBedStates][numBedStates]float64

// kObs 在线核：occ→occ=1−ε / occ→vac=ε / vac→occ=μ / vac→vac=1−μ。
func (p bedAxisParams) kObs() bedKernel {
	var k bedKernel
	k[BOcc][BOcc] = 1 - p.epsilon
	k[BOcc][BVac] = p.epsilon
	k[BVac][BOcc] = p.mu
	k[BVac][BVac] = 1 - p.mu
	return k
}

// kUnobs 离线核（§C 定义 B 单向泄漏）：occ→vac=λ / occ→occ=1−λ / vac→vac=1 / vac→occ=0。
// vac→occ 恒 0 = vac 吸收态：陈旧 occ 单向蒸发，空房离线绝不被弛豫成伪占用（cd2b 根方向）。
func (p bedAxisParams) kUnobs() bedKernel {
	var k bedKernel
	k[BOcc][BVac] = p.lambda
	k[BOcc][BOcc] = 1 - p.lambda
	k[BVac][BVac] = 1.0 // vac 吸收
	k[BVac][BOcc] = 0.0
	return k
}

// tBKernel 按 ρ（sleepad 在线）选核。§6：ρ_t = 1[s_j 在线]。
func (p bedAxisParams) tBKernel(online bool) bedKernel {
	if online {
		return p.kObs()
	}
	return p.kUnobs()
}
