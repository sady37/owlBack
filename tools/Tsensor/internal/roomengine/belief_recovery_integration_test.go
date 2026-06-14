package roomengine

import (
	"fmt"
	"strconv"
	"testing"

	"owl-common/alarm"
	rediscommon "owl-common/redis"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// TestRecoveryVetoIntegration — P2 集成测(委员会待办:验 recovery-veto 在 **production 路径**):
// 喂 firmware Fall 事件(category=Fall→recordBeliefShadowFirmwareFall 设 firmwareFallTs)+ 恢复轨迹 track 帧。
// 纯误火(摔后倒地<15s 即起)→ recovery 触发(would-veto);自救真摔(摔后倒地≥15s)→ 不触发(genuine-fall guard)。
func TestRecoveryVetoIntegration(t *testing.T) {
	cfg, radarAddr, roomID := bLayout(t, "unit201-handoff-0609-bathroom-333B")
	run := func(fallenSec int) int {
		core, logs := observer.New(zapcore.DebugLevel)
		e := NewEngine(nil, zap.New(core))
		e.RegisterRoom(cfg)
		e.deviceRoom[radarAddr] = roomID
		e.deviceMounts[radarAddr] = cfg.Radar
		ts := int64(1781000000000)
		feedTrack := func(pose, x, y, z int) {
			dv := fmt.Sprintf(`[{"pose":%d,"position_x":%d,"position_y":%d,"position_z":%d,"track_id":0,"track_confidence":80}]`, pose, x, y, z)
			e.handleMessage(nil, rediscommon.StreamMessage{Values: map[string]interface{}{
				"device_addr": radarAddr, "device_type": "radar", "topic_type": "monitor", "category": "track",
				"timestamp": strconv.FormatInt(ts, 10), "dataValue": dv}})
			ts += 1000
		}
		for i := 0; i < 10; i++ { // 摔前走动 10s
			feedTrack(1, 100+i*10, 100, 100)
		}
		tFire := ts // firmware Fall @ T_fire
		e.handleEventMessage(rediscommon.StreamMessage{Values: map[string]interface{}{
			"device_addr": radarAddr, "device_type": "radar", "topic_type": "event", "category": alarm.Fall,
			"timestamp": strconv.FormatInt(tFire, 10),
			"dataValue": fmt.Sprintf(`[{"event_status":"start","event_since":%d,"track_id":0,"pose":5}]`, tFire)}})
		for i := 0; i < fallenSec; i++ { // 摔后倒地 fallenSec 秒
			feedTrack(5, 200, 100, 0)
		}
		for i := 0; i < 8; i++ { // 起身直立 8s
			feedTrack(4, 200, 100, 100)
		}
		return logs.FilterMessage("belief_dbn_recovery_evidence").Len()
	}
	pureFalse := run(3)   // 倒地 3s <15s = 纯误火
	selfRescue := run(20) // 倒地 20s ≥15s = 自救真摔
	t.Logf("纯误火(倒地3s) recovery=%d / 自救真摔(倒地20s) recovery=%d", pureFalse, selfRescue)
	if pureFalse == 0 {
		t.Errorf("★纯误火(摔后倒地3s即起) 应触发 recovery(would-veto)——production recovery-veto wire 漏")
	}
	if selfRescue > 0 {
		t.Errorf("★自救真摔(摔后倒地20s) 不应触发 recovery(genuine-fall guard 失效)")
	}
}
