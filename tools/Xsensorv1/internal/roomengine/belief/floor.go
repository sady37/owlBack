package belief

// floor.go — FN-safe last-resort floor（契约 6A 其十五）。召回兜底:still-box **总时长** ≥ T_floor(按 area)
// 且**无正向休息证据** → 强制 fall-suspect。接住 emission 被 area误学/z假阳/接触假阳**误压**的真摔。
//
// 与被否的 60s 直投区别:60s 是**主路径+短阈**(海量 FP);floor 是**兜底+保守长阈**(正常活动不触,
// 只接被误压真摔)。绕过前向滤波——floor 不看 P(Fallen),只看总时长 + 可观测豁免。
//
// 豁免**挂可观测证据,不挂 label**(否则 area 误学的 FN 从兜底又漏):
//   - AreaDeny(15天高bar 静态反射)→ 非真占用,交 realness。
//   - 接触 InBed → 真在床。
//   - z 坐/站高**持续 ≥N 帧**(守门★:单帧 z 噪声不解豁免,防 area误学+z假阳组合漏真摔)。

type FloorParams struct {
	TFloorDefaultSec float64 // 开阔/未知区总时长阈(默认 600s=10min)
	TFloorBathSec    float64 // 卫浴区阈(默认 1200s=20min,constipation 余量)
	ZSustainFrames   int     // z 坐/站高持续帧阈(守门★,默认 5;单帧噪声不解豁免)
}

// DefaultFloorParams 形态默认(铁律 [[fall_data_is_artificial_test]]:非权威值,留 oracle)。
func DefaultFloorParams() FloorParams {
	return FloorParams{TFloorDefaultSec: 600, TFloorBathSec: 1200, ZSustainFrames: 5}
}

// FloorGuard 单 logicID 兜底守门(跨帧持 z 直立持续计数)。
type FloorGuard struct {
	p         FloorParams
	zUpFrames int // 连续 z 坐/站高帧数(守门★)
}

func NewFloorGuard() *FloorGuard { return &FloorGuard{p: DefaultFloorParams()} }

// Step 一帧兜底判定。zUpright=z 在坐/站档(ZSit/ZStand);contactInBed=任一床 InBed。
// 返回 floorFire。track 消失帧不调(无观测);present 帧每帧调。
func (g *FloorGuard) Step(stillSec float64, areaType int, zUpright, contactInBed bool) bool {
	if zUpright {
		g.zUpFrames++
	} else {
		g.zUpFrames = 0
	}
	// 豁免(可观测证据):
	switch {
	case areaType == areaDeny:
		return false // 静态反射 → realness 管
	case contactInBed:
		return false // 真在床
	case g.zUpFrames >= g.p.ZSustainFrames:
		return false // z 持续直立/坐 = 真直立(守门★)
	}
	tFloor := g.p.TFloorDefaultSec
	if areaType == areaToilet || areaType == areaShower {
		tFloor = g.p.TFloorBathSec
	}
	return stillSec >= tFloor
}
