package zoneengine

import (
	"testing"
	"time"

	"owl-common/alarm"
)

// stubBedResolver 测试用：把 radar 床事件解析到指定床 /96（模拟 MM covers 解析）。
type stubBedResolver struct{ bed string }

func (s stubBedResolver) ResolveBed(radarAddr, roomPrefix string) (string, float64) {
	return s.bed, 1.0
}

// TestZoneBed_RadarSleepadFuseSameBed — Layer-0 单床房床身份修复的真值靶。
//
// radar 与 sleepad 装在同一张物理床（room 112 BedA :3:101）。bed_bayesian_scorer 设计为融合
// 双源（LR_S + LR_R + γ schedule），前提是两源喂进同一 bed zone(/96)。
//
// 修复前 adapter_radar 用 radar 自己的 /96(:3:100) → 与 sleepad(:3:101) 两个独立 zone，融合空转。
// 修复后 SetBedResolver 注入"房→唯一注册床 /96"，radar 床事件归到 :3:101 → 与 sleepad 同 zone 融合。
//
// 非 vacuous：去掉 SetBedResolver（或 adapter 不用 resolveBedPref）→ radar 进 :3:100 → radarOwnZone=true → FAIL。
func TestZoneBed_RadarSleepadFuseSameBed(t *testing.T) {
	engine := NewEngine(DefaultRules(), StaticBedSizeLookup{Bucket: "small"}, nil)
	radarA := NewRadarAdapter(nil, engine, nil, nil)
	radarA.SetBedResolver(stubBedResolver{bed: "fd00:0:3:112:3:101::/96"}) // 单床房：radar→床 /96
	sleepA := NewSleepaceAdapter(nil, engine, nil)

	now := time.Now().UnixMilli()
	const radarAddr = "fd00:0:3:112:3:100:32a1:cd2b"   // 自身 /96 = :3:100
	const sleepadAddr = "fd00:0:3:112:3:101:2460:1641" // 注册床 BedA /96 = :3:101

	radarA.handleMsg(mkMsg(radarAddr, "Radar", alarm.InBed, "card-112", now,
		map[string]interface{}{"track_id": float64(1)}))
	sleepA.handleMsg(mkMsg(sleepadAddr, "Sleepad", alarm.InBed, "card-112", now+1000, nil))

	_, radarOwnZone := engine.GetState(StateKey{ZoneType: ZoneTypeBed, ZoneID: "fd00:0:3:112:3:100::/96"})
	_, fusedZone := engine.GetState(StateKey{ZoneType: ZoneTypeBed, ZoneID: "fd00:0:3:112:3:101::/96"})

	if radarOwnZone {
		t.Errorf("radar 床事件仍进了自己的 /96(:3:100)——resolver 未生效；应被归到注册床 /96(:3:101)")
	}
	if !fusedZone {
		t.Errorf("radar+sleepad 应融合进同一 bed zone :3:101，但该 zone 不存在")
	}
}
