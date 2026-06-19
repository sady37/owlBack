package belief

// floor.go — FN-safe last-resort floor（契约 6A 其十五）。召回兜底:**有效 still 总时长**（obs.StillSec，
// 已含直立折扣 stillDiscount，SnapshotTrackStatuses 算）≥ per-area T_floor 且无正向休息证据 → 强制
// fall-suspect。接住 emission 被 area误学/接触假阳**误压**的真摔。
//
// 与被否的 60s 直投区别:60s 是**主路径+短阈**(海量 FP);floor 是**兜底+保守 per-area 长阈**(正常活动不触,
// 只接被误压真摔)。绕过前向滤波——floor 不看 P(Fallen),只看有效总时长 + 可观测豁免。
//
// z/pose 直立证据已并入有效 still（stillDiscount：pose=sit×0.8 / z 坐×0.9 / z 站×0.5），floor 不再单独
// 看 z（避免同一 z 双压 SFallen → 站立瘫倒过压漏报）。豁免**挂可观测证据,不挂 label**:
//   - AreaDeny(15天高bar 静态反射)→ 非真占用,交 realness。
//   - **本 track 所在床** 接触 InBed → 真在床（按 NearBedMask 收窄到本床，非任一房间 sleepad——
//     否则双人房里他人在床会误豁免地板上的摔者 = FN）。

// ============================================================================
// per-area 静止时长统计模型（单源）：floor 兜底 + D/DU 保底窗(engine/unit.go)
//   + (规划)emission 高斯 CDF 共用此锚，避免数字各漂。
//
// 标定法（用户 2026-06-18 定）：各区"正常停留时长" ~ 高斯(μ,σ)，**异常阈 = μ+1.5σ**（上侧分位）。
//   ±1.5σ 含 86.64%；只取上侧（停留太久=疑似摔；太短=人正常离开，非异常）→ 单侧上尾 6.68%
//   = 异常阈自带 FP 率（选 1.5σ 而非 2σ=2.3%：高危跌倒宁多扰、少漏）。
//   emission 用法(规划)：SFallen 贡献 = Φ((StillboxSec−μ)/σ)，正常值≈0.5、异常阈≈0.93。
//
// per-area (μ,σ) 数据来源（铁律 [[fall_data_is_artificial_test]]：无真摔数据标定，留 oracle）：
//   Bath μ=12min σ=4min 【文献硬支撑】医学建议正常 5–10min；健康人 >20min 仅 0.5%(≈μ+2σ) 反推 σ≈4；
//        老人便秘倾向取 μ=12 → 异常 18min(μ+1.5σ) 覆盖 ~80% 便秘，severe 20min(μ+2σ)。
//        来源：CNN 马桶<10min / anorectal PMC12669168(>20min 0.5%健康 vs 8.1%病)
//   Sit·Lying(含 Bed) μ=60min σ=20min(μ/3) 【保守】文献久坐单次 MBD 11.7–16min、prolonged ≥30min；
//        μ=60 取久坐久卧"容忍上限"(看电视/午睡，FN-safe 高于平均、低危别老打扰) → 异常 90min。
//        来源：sedentary PMC8679788(MBD 11.7-16) / stroke PMC9166254(≥17min 高危)
//   default(开阔/站立) μ=8min σ=2.67min(μ/3) 【经验】无直接文献(非标准观测场景)，类比久坐短 bout<10min
//        → 异常 12min。oracle 待真实数据调。
// ============================================================================
const (
	// 各区正常停留 (μ, σ) 秒——emission 高斯 CDF 与 floor 异常阈(tFloorFor=μ+1.5σ, room×cell 保守)共用单源(§H/§I)
	MuDefaultSec, SigmaDefaultSec = 480, 160   // 8min, 2.67min
	MuSitSec, SigmaSitSec         = 3600, 1200 // 60min, 20min
	MuBathSec, SigmaBathSec       = 720, 240   // 12min, 4min
)

// FloorGuard 单 logicID 兜底守门（§I stillbox 计时器；异常阈 tFloorFor=μ+1.5σ room×cell 保守，无跨帧状态）。
type FloorGuard struct{}

func NewFloorGuard() *FloorGuard { return &FloorGuard{} }

// Step 一帧兜底判定（吃整条 obs，自抽派生量——engine 只 OR verdict 不碰 obs）。有效 still 时长
// （obs.StillSec，已含直立折扣）≥ per-area 阈 → floorFire。track 消失帧不调(无观测);present 帧每帧调。
func (g *FloorGuard) Step(obs Observation) bool {
	contactInBed := false // 仅**本 track 所在床**(NearBedMask)的 InBed 才豁免——他人在床不豁免本摔者
	for j, br := range obs.Sleepad {
		if br == BedInBed && obs.NearBedMask[j] {
			contactInBed = true
			break
		}
	}
	switch {
	case obs.AreaType == areaDeny:
		return false // 静态反射 → realness 管
	case contactInBed:
		return false // 真在床
	}
	return obs.StillSec >= tFloorFor(obs.AreaType, obs.RoomType)
}

// tFloorFor floor(= §I stillbox 计时器)异常阈 = stillMuSigma(area,room) 的 μ+1.5σ
// （room×cell 保守合并，跟 emission CDF 同源 §H/§I）：bathroom 未画 toilet 也用 bathsec(18min)，
// 不再 cell-only 12min 抢先于 CDF 的 18min。
func tFloorFor(area, roomType int) float64 {
	mu, sigma := stillMuSigma(area, roomType)
	return mu + 1.5*sigma
}
