// track_status.go — cell area type wire 序列化（areaTypeWireName）。

package roomengine

// areaTypeWireName cell area type 序列化字符串（v2 sensor 内部完整字典）。
//
// **与 areaTypeProtocolStr (engine.go) 是不同的序列化函数，对应不同的下游消费方**：
//
//   - areaTypeWireName    → 用于 sensor:track:status:stream（本流，PR-3 引入）
//     完整 v2 内部语义：toilet / shower / sit 等独立映射。
//     消费方：dev playback (track-status-tail) / 未来 zoneengine v2 / fall verifier。
//
//   - areaTypeProtocolStr → 用于 iot:event:stream / iot:alarm:stream（cardagg 消费）
//     向固件协议级 area（best-effort，5 类）投影。
//     消费方：cardagg AI verdict 路由 / alarm 落库。
//
// 设计意图：v2 内部表达力（高维）和 cardagg 协议向后兼容（低维）解耦——
// 改本字典不会牵动 cardagg，反之亦然。新增消费方时按"哪条流"决定用哪套。
func areaTypeWireName(t AreaType) string {
	switch t {
	case AreaEnter:
		return "enter"
	case AreaBed:
		return "bed"
	case AreaMonitorBed:
		return "monitor_bed"
	case AreaReflector:
		return "reflector"
	case AreaInterfer:
		return "interfer"
	case AreaSit:
		return "sit"
	case AreaLying:
		return "lying"
	case AreaActive:
		return "active"
	case AreaDeny:
		return "deny"
	}
	return "unknown"
}
