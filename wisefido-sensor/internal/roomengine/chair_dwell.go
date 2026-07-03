package roomengine

// chair_dwell.go — per-chair 久坐学习态（object_id 锚定，跨 layout 编辑存活）。
//
// 原学习态挂 grid anchor cell（cell index），layout 编辑重建 grid → cold-start 清零。改挂 canvas
// object_id（uuid，跨编辑保号），独立 sibling 表 chair_dwell_state（sensor 单一 maintainer，layout 不碰）。
// 一椅一 ChairDwell：14 日滚动窗（每 >5min walk-away 久坐 episode 进当天桶）+ 缓存 μ/σ/N 供 floor 热路径
// O(1) 读 + maxSit 反馈棘轮（false_alarm+"Sit on Chair"）。衰减/棘轮/直方图公式与旧 anchor-cell 版逐字一致，
// 只把存活载体从 grid cell 换成 per-object。

import (
	"math"

	"owl-common/radarutils"
)

// ChairRect layout 解析出的 chair pin（rect + object_id）。object_id 透传自 canvas 对象 id，
// 是 per-chair 学习态跨 layout 编辑存活的稳定 key。
type ChairRect struct {
	Rect     radarutils.Rect
	ObjectID string
}

// DwellColdMinN 久坐窗冷启门控：窗口样本 < 此值 → floor 回退 90min（声明椅学够前不误报）。
const DwellColdMinN = 3

const (
	dwellWindowDays      = 14  // 久坐窗滚动天数
	dwellHistBins        = 18  // 5min bin：[0]=[5,10)…[16]=[85,90)，[17]=[90,∞)
	dwellHistLowMin      = 5   // 第一个 bin 下界(min)；<5min 已过滤不入
	dwellHistBinMin      = 5   // bin 宽(min)
	dwellMaxSitMarginSec = 600 // maxSit 棘轮在实际久坐时长上加的 10min margin（抬到习惯久坐之上压复发）
	dwellMaxSitFloorSec  = 60  // maxSit 衰减到 <1min 视作归零（清掉残尾，退回 μ+1.5σ 自然兜底）
	chairDisplacementCm  = 200 // object 中心位移 ≥ 此值 → 视作换了把椅子，学习态打折（一次性）
)

// dwellMaxSitHalfLifeMs maxSit 反馈棘轮的指数衰减半衰期（14 天，对齐 14 日久坐学习窗，统一时间尺度）。
// 习惯长坐者窗内自学会→自然公式接管;一次性异常与窗同步淡出→FN 保护更快自愈。
// 衰减回 0 不漏：tFloor 的底永远是 μ+1.5σ+10min（实测分布），maxSit 只是补 14 日窗遗忘的短期空窗。
const dwellMaxSitHalfLifeMs = int64(dwellWindowDays) * 24 * 3600 * 1000

// DwellBucket 一天的 >5min 久坐统计：N/Sum/SumSq 精确算 μ+1.5σ；Hist(5min bin) 留 median/分位后期用。
type DwellBucket struct {
	Day   int32                 `json:"d"`  // epoch day = unixSec/86400 (UTC)
	N     int32                 `json:"n"`  // 样本数
	Sum   float64               `json:"s"`  // Σ时长(秒)
	SumSq float64               `json:"sq"` // Σ时长²
	Hist  [dwellHistBins]uint16 `json:"h"`  // 5min 直方图
}

func epochDay(nowMs int64) int32 { return int32(nowMs / 1000 / 86400) }

func dwellHistIdx(durSec float64) int {
	i := (int(durSec/60) - dwellHistLowMin) / dwellHistBinMin
	if i < 0 {
		i = 0
	}
	if i >= dwellHistBins {
		i = dwellHistBins - 1
	}
	return i
}

// ChairDwell 单把椅子的久坐学习态（keyed by object_id）。sensor 单向写；Xsensorv1 侧只读镜像不含 Window。
type ChairDwell struct {
	Window   []DwellBucket // 14 日环
	Mu       float64       // 缓存：窗口 μ（秒，AV）；0=冷启
	Sig      float64       // 缓存：窗口 σ（秒）
	N        int           // 缓存：窗口总样本数（冷启门控 <DwellColdMinN）
	MaxSit   float64       // false_alarm+"Sit on Chair" 反馈棘轮（秒）：人工确认安全的久坐下限，抗 14 日窗遗忘致复发。
	MaxSitMs int64         // MaxSit 当前值的 as-of 时刻（指数衰减基准）
	CX, CY   int           // object 中心 canvas 坐标（位移打折基准，reload 时比对）
}

// AppendDwell 记一条 >5min walk-away 久坐(秒)到当天桶(14 日环)。调用方持锁。
func (cd *ChairDwell) AppendDwell(durSec float64, nowMs int64) {
	day := epochDay(nowMs)
	var b *DwellBucket
	for i := range cd.Window {
		if cd.Window[i].Day == day {
			b = &cd.Window[i]
			break
		}
	}
	if b == nil {
		if len(cd.Window) < dwellWindowDays {
			cd.Window = append(cd.Window, DwellBucket{Day: day})
			b = &cd.Window[len(cd.Window)-1]
		} else {
			oldest := 0
			for i := 1; i < len(cd.Window); i++ {
				if cd.Window[i].Day < cd.Window[oldest].Day {
					oldest = i
				}
			}
			cd.Window[oldest] = DwellBucket{Day: day}
			b = &cd.Window[oldest]
		}
	}
	b.N++
	b.Sum += durSec
	b.SumSq += durSec * durSec
	b.Hist[dwellHistIdx(durSec)]++
}

// RecomputeDwell 慢周期(hourly)重算：衰减 maxSit 棘轮 + 丢 >14 天槽 → 聚合本窗 → 写缓存 μ/σ/N。调用方持锁。
func (cd *ChairDwell) RecomputeDwell(nowMs int64) {
	cd.decayMaxSit(nowMs) // 棘轮慢衰减（不依赖窗口，空窗也要衰减），须在早返回前
	if len(cd.Window) == 0 {
		cd.Mu, cd.Sig, cd.N = 0, 0, 0
		return
	}
	cutoff := epochDay(nowMs) - (dwellWindowDays - 1)
	live := cd.Window[:0]
	var n int
	var sum, sumSq float64
	for _, b := range cd.Window {
		if b.Day < cutoff {
			continue // 丢 >14 天
		}
		live = append(live, b)
		n += int(b.N)
		sum += b.Sum
		sumSq += b.SumSq
	}
	cd.Window = live
	cd.N = n
	if n == 0 {
		cd.Mu, cd.Sig = 0, 0
		return
	}
	cd.Mu = sum / float64(n)
	v := sumSq/float64(n) - cd.Mu*cd.Mu
	if v < 0 {
		v = 0
	}
	cd.Sig = math.Sqrt(v)
}

// RatchetChairMaxSit false_alarm+"Sit on Chair" 反馈棘轮：把"本次被误报的实际久坐时长 + margin"
// 抬进本椅 maxSit（取 max）。sitDurSec=fire 时观测到的久坐 still 秒（evidence.fire.stillbox_sec）。
// 下次同把椅子坐同样久 → tFloor=max(μ+1.5σ+10min, maxSit) 已抬到该时长之上 → 不再误火。
// 先把已有值衰减到现在再比较抬升（否则旧高值不衰减永久占顶）；仅累不夹 90min（硬顶在 tFloorFor 统一夹）。
func (cd *ChairDwell) RatchetChairMaxSit(sitDurSec float64, nowMs int64) {
	cd.decayMaxSit(nowMs)
	if v := sitDurSec + dwellMaxSitMarginSec; v > cd.MaxSit {
		cd.MaxSit = v
	}
	cd.MaxSitMs = nowMs
}

// decayMaxSit 对 maxSit 棘轮做指数慢衰减（半衰期 dwellMaxSitHalfLifeMs）。MaxSitMs=当前值的 as-of 时刻；
// 每次衰减后推进到 nowMs（在 RecomputeDwell hourly + 棘轮前调）。无记录/冷启（MaxSitMs=0）→ 从现在起算不立即衰减。
// 衰减到 <1min 归零，退回 μ+1.5σ 自然兜底。调用方持锁。
func (cd *ChairDwell) decayMaxSit(nowMs int64) {
	if cd.MaxSit <= 0 {
		cd.MaxSitMs = 0
		return
	}
	if cd.MaxSitMs == 0 || nowMs <= cd.MaxSitMs {
		cd.MaxSitMs = nowMs // 初始化/校正 as-of（迁移 + 时钟回拨防护）
		return
	}
	cd.MaxSit *= math.Exp(-math.Ln2 * float64(nowMs-cd.MaxSitMs) / float64(dwellMaxSitHalfLifeMs))
	cd.MaxSitMs = nowMs
	if cd.MaxSit < dwellMaxSitFloorSec {
		cd.MaxSit, cd.MaxSitMs = 0, 0
	}
}

// applyDisplacementDiscount object 中心搬 ≥200cm（换了把椅子/挪位）时一次性打折：dms×0.5 + dwin 各槽
// 样本数 N 减半（降置信让新习惯更快接管），Sum/SumSq 按 N 的实际减半比例同步缩放 → μ、σ 精确保留。
// （直接 ×0.5 会因 N 整数截断 5→2 而漂 μ；故用 newN/N 比例缩放。）调用方随后 RecomputeDwell 刷缓存。
func (cd *ChairDwell) applyDisplacementDiscount() {
	for i := range cd.Window {
		b := &cd.Window[i]
		newN := b.N / 2
		if b.N > 0 {
			r := float64(newN) / float64(b.N)
			b.Sum *= r
			b.SumSq *= r
		} else {
			b.Sum, b.SumSq = 0, 0
		}
		b.N = newN
		for j := range b.Hist {
			b.Hist[j] /= 2
		}
	}
	cd.MaxSit *= 0.5
}
