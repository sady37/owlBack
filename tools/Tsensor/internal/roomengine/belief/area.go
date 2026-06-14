package belief

// area_type 值契约：canonical 定义在 roomengine/cell.go（type AreaType）。belief 是底层包，
// roomengine 反向 import 它，故 belief 看不到 roomengine.AreaType —— 按值契约镜像（改 cell.go 枚举值必须同步此处）。
// 取代旧 Geom intermediate：likelihood 直接基于权威 cell.area_type + 门距 + bed_state 判位置语义。
const (
	areaUnknown = 0
	areaEnter   = 1
	areaBed     = 2
	areaSit     = 3
	areaActive  = 4
	areaDeny    = 5
	areaShower  = 6
	areaToilet  = 7
)

var areaTypeLabel = [...]string{"Unknown", "Enter", "Bed", "Sit", "Active", "Deny", "Shower", "Toilet"}

// AreaTypeLabel area_type → 名（observability；越界=Unknown）。
func AreaTypeLabel(a int) string {
	if a < 0 || a >= len(areaTypeLabel) {
		return "Unknown"
	}
	return areaTypeLabel[a]
}

// AreaCtx ObsPose / poseLikelihood 的位置上下文（权威源已解析，构造侧填一次）。
// NearDoor=离门≤30cm（门距几何，area_type 不编码动态门距）；BedReleased=bed_state 离床（权威>几何，床区躺的睡/摔判别）。
type AreaCtx struct {
	AreaType    int
	NearDoor    bool
	BedReleased bool
}

func isOpenFloor(a int) bool {
	return a == areaSit || a == areaActive || a == areaDeny
}
