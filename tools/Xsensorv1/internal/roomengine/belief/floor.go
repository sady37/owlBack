package belief

// floor.go — FN-safe last-resort floor（契约 6A 其十五）。召回兜底:**still raw 总时长**（obs.StillSec=
// StillBoxSec 纯计时器，无直立折扣）≥ per-area T_floor 且无正向休息证据 → 强制 fall-suspect。
// 接住 emission 被 area误学/接触假阳**误压**的真摔。
//
// 与被否的 60s 直投区别:60s 是**主路径+短阈**(海量 FP);floor 是**兜底+保守 per-area 长阈**(正常活动不触,
// 只接被误压真摔)。绕过前向滤波——floor 不看 P(Fallen),只看 raw 总时长 + 可观测豁免。
//
// 直立/活动证据（z≥80 / pose=sit·walk）已单一归 emission（压 SFallen，移自旧 stillDiscount）——floor 是
// **不信 pose/z 的纯时间兜底**（pose/z 错报时正是 floor 该接的场景），故 floor 绝不再看 z/pose（杜绝双压）。
// 豁免**挂可观测证据,不挂 label**:
//   - AreaReflector/AreaInterfer(镜金属静态反射 + 帘扇运动杂波)→ 无真人,交 realness/firmware masking。
//   - AreaBed/AreaMonitorBed → 接触床区按 AreaType 豁免（无 sleepad 时几何分不清在床/床边）。
//   - **本 track 所在床** 接触 InBed → 真在床（按 NearBedMask 收窄到本床，非任一房间 sleepad——
//     否则双人房里他人在床会误豁免地板上的摔者 = FN）。

// ============================================================================
// per-area 静止时长统计模型（单源）：floor 兜底 + D/DU 保底窗(engine/unit.go)
//   - (规划)emission 高斯 CDF 共用此锚，避免数字各漂。
//
// 标定法（用户 2026-06-18 定）：各区"正常停留时长" ~ 高斯(μ,σ)，**异常阈 = μ+1.5σ**（上侧分位）。
//
//	±1.5σ 含 86.64%；只取上侧（停留太久=疑似摔；太短=人正常离开，非异常）→ 单侧上尾 6.68%
//	= 异常阈自带 FP 率（选 1.5σ 而非 2σ=2.3%：高危跌倒宁多扰、少漏）。
//	emission 用法(规划)：SFallen 贡献 = Φ((StillboxSec−μ)/σ)，正常值≈0.5、异常阈≈0.93。
//
// per-area (μ,σ) 数据来源（铁律 [[fall_data_is_artificial_test]]：无真摔数据标定，留 oracle）：
//
//	Bath μ=12min σ=4min 【文献硬支撑】医学建议正常 5–10min；健康人 >20min 仅 0.5%(≈μ+2σ) 反推 σ≈4；
//	     老人便秘倾向取 μ=12 → 异常 18min(μ+1.5σ) 覆盖 ~80% 便秘，severe 20min(μ+2σ)。
//	     来源：CNN 马桶<10min / anorectal PMC12669168(>20min 0.5%健康 vs 8.1%病)
//	Sit·Lying(含 Bed) μ=60min σ=20min(μ/3) 【保守】文献久坐单次 MBD 11.7–16min、prolonged ≥30min；
//	     μ=60 取久坐久卧"容忍上限"(看电视/午睡，FN-safe 高于平均、低危别老打扰) → 异常 90min。
//	     来源：sedentary PMC8679788(MBD 11.7-16) / stroke PMC9166254(≥17min 高危)
//	default(开阔/站立) μ=8min σ=2.67min(μ/3) 【经验】无直接文献(非标准观测场景)，类比久坐短 bout<10min
//	     → 异常 12min。oracle 待真实数据调。
//
// ============================================================================
const (
	// 各区正常停留 (μ, σ) 秒——emission 高斯 CDF 与 floor 异常阈(tFloorFor=μ+1.5σ, room×cell 保守)共用单源(§H/§I)
	MuDefaultSec, SigmaDefaultSec = 480, 160   // 8min, 2.67min
	MuSitSec, SigmaSitSec         = 3600, 1200 // 60min, 20min
	MuBathSec, SigmaBathSec       = 800, 267   // 浴室一律 → 异常 20min（占用区，马桶/淋浴久留紧阈）
)

// FloorGuard 单 logicID 兜底守门（§I stillbox 计时器；异常阈 tFloorFor=μ+1.5σ room×cell 保守，无跨帧状态）。
type FloorGuard struct{}

func NewFloorGuard() *FloorGuard { return &FloorGuard{} }

// Step 一帧兜底判定（吃整条 obs，自抽派生量——engine 只 OR verdict 不碰 obs）。有效 still 时长
// （obs.StillSec，still raw 时长）≥ per-area 阈 → floorFire。track 消失帧不调(无观测);present 帧每帧调。
func (g *FloorGuard) Step(obs Observation) bool {
	// 接触豁免收窄到**唯一一张近床且 InBed**（FN-safe 二值）：近多张床=床归属分不清（NearBedMask 非互斥，
	// 摔者同时在两床 100cm 内 → 他人在床会误豁免本摔者）→ 不豁免照报。1 张近床 = 归属确定（你说的"1张床100%"）。
	// 残留：摔倒紧贴某占用床(仅近这一张)仍会豁免——雷达分不清"我在床"vs"我摔在床边"的固有二义（belief Ψ 同此，
	// 靠 a_j 概率软化），floor 二值兜底取保守不细分。多床二义这条已堵。
	nearCount, nearJ := 0, -1
	for j, near := range obs.NearBedMask {
		if near {
			nearCount++
			nearJ = j
		}
	}
	contactInBed := nearCount == 1 && nearJ < len(obs.Sleepad) && obs.Sleepad[nearJ] == BedInBed
	switch {
	case obs.AreaType == areaBed || obs.AreaType == areaMonitorBed:
		return false // 真床/监护床：接触床区按 AreaType 豁免（无 sleepad 时几何分不清在床/床边摔，宁漏边缘 FN 不让睡眠误报）
	case obs.AreaType == areaReflector:
		// 镜面/金属静态反射：几何确定的反射体，整片豁免。
		// （interfer 杂波区不再整片豁免——会漏报走进去摔倒的真人；其伪迹改在 track 出生时按身份抑制 still-box。）
		return false
	case contactInBed:
		return false // 真在床
	}
	return obs.StillSec >= tFloorFor(obs.AreaType, obs.RoomType, obs.IsRiskTime, obs.InChair, obs.ChairMu, obs.ChairSigma)
}

// tFloorFor floor(= §I stillbox 计时器)异常阈。risktime 纯时间轴：μ,σ 不变(物理),只动风险容忍 k——
// 白天 1.5σ(保守防 FP)/夜间也 1.5σ(用户 2026-06-27:夜间不提前兜底,防误报多致无人用)。不进 C_FN(报警阈与时辰无关)。
//
// 分层（镜像 wisefido-sensor 正本）：
//   bathroom 房     → bath 阈(20min)，压过一切（占用区紧阈）。
//   chair 区(inChair)→ 本格学到的 dwell 分布(hydrate 自 28)：tFloor=clamp(μ+k·σ,[12min,90min])；冷启(μ≤0)回退 90min。
//                      pin 几何 gate，免疫 Belief 翻转；坐得越久的椅子阈越大。
//   AreaLying(沙发) → sit 阈(90min) 二值（独立躺区）。
//   其余            → default 阈(12min)，快报真摔。
func tFloorFor(area, roomType int, isRiskTime, inChair bool, chairMu, chairSigma float64) float64 {
	k := 1.5
	if isRiskTime {
		k = 1.5
	}
	tActive := MuDefaultSec + k*SigmaDefaultSec
	tSit := MuSitSec + k*SigmaSitSec
	switch {
	case roomType == roomBathroom:
		return MuBathSec + k*SigmaBathSec
	case inChair:
		if chairMu <= 0 {
			return tSit // 冷启：声明椅学够前回退 90min
		}
		t := chairMu + k*chairSigma
		if t < tActive {
			t = tActive
		} else if t > tSit {
			t = tSit
		}
		return t
	case area == areaLying:
		return tSit
	default:
		return tActive
	}
}
