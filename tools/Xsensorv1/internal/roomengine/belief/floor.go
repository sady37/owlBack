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
//   - 接触 InBed → 真在床。

type FloorParams struct {
	TFloorDefaultSec float64 // 开阔/未知区总时长阈(默认 600s=10min)
	TFloorSitSec     float64 // 久坐区阈(默认 5400s=90min,看电视/发呆容忍)
	TFloorBathSec    float64 // 卫浴区阈(默认 1200s=20min,constipation 余量)
}

// DefaultFloorParams 形态默认(铁律 [[fall_data_is_artificial_test]]:非权威值,留 oracle)。
func DefaultFloorParams() FloorParams {
	return FloorParams{TFloorDefaultSec: 600, TFloorSitSec: 5400, TFloorBathSec: 1200}
}

// FloorGuard 单 logicID 兜底守门（无跨帧状态：z 直立已并入有效 still 折扣，不再累计 zUpFrames）。
type FloorGuard struct {
	p FloorParams
}

func NewFloorGuard() *FloorGuard { return &FloorGuard{p: DefaultFloorParams()} }

// Step 一帧兜底判定（吃整条 obs，自抽派生量——engine 只 OR verdict 不碰 obs）。有效 still 时长
// （obs.StillSec，已含直立折扣）≥ per-area 阈 → floorFire。track 消失帧不调(无观测);present 帧每帧调。
func (g *FloorGuard) Step(obs Observation) bool {
	contactInBed := false
	for _, br := range obs.Sleepad {
		if br == BedInBed {
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
	return obs.StillSec >= g.tFloor(obs.AreaType)
}

// tFloor per-area 总时长阈：久坐 90min / 卫浴 20min / 开阔·未知 10min。
func (g *FloorGuard) tFloor(area int) float64 {
	switch area {
	case areaSit:
		return g.p.TFloorSitSec
	case areaToilet, areaShower:
		return g.p.TFloorBathSec
	default:
		return g.p.TFloorDefaultSec
	}
}
