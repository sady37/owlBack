package roomengine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	rediscommon "owl-common/redis"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// belief_b_replay_test.go — P9 载体 B(审查㊸ 批准:B-1=(a)直驱 handleMessage / B-2=(a)zap-observer 日志为主)。
//
// 保真硬条件(㊸ 机械堵"合成绿"):**只**喂 raw fixture record 进生产 `handleMessage`/`handleEventMessage`
// + setup(NewEngine/RegisterRoom/wire deviceRoom/mount)+ observer 读 `belief_shadow_*` 日志。
// **禁**手搓 `TrackStatusBase`、禁调非生产 tick、禁 fork shadow 逻辑。fixture record → 真 StreamMessage →
// FromStreamMap → ParseRadarTracks → ProcessFrame → SnapshotTrackStatuses → 真 beliefShadowTick(零手搓零 fork)。
//
// 首 oracle = 复现 C 诊断:8 CD2B 卧室 fall 逐案分类对账(α→bed 权威压制 / β→nodetect 门控 / γ→escalate)。
// B 结果与 C 手工诊断分歧 = pipeline bug(B 同时是 C 的交叉校验)。

// bRecord 一条 fixture 记录(monitor/event 流原始行)。
type bRecord struct {
	DeviceUID string                   `json:"device_uid"`
	Timestamp int64                    `json:"timestamp"`
	TopicType string                   `json:"topic_type"`
	Category  string                   `json:"category"`
	DataValue []map[string]interface{} `json:"data_value"`
}

// bLoadRecords 读 fixture window json → 按 ts 升序记录。
func bLoadRecords(t *testing.T, dir, file string) []bRecord {
	b, err := os.ReadFile(filepath.Join(casesDir, dir, file))
	if err != nil {
		t.Skipf("fixture 缺失 %s/%s: %v", dir, file, err)
	}
	var recs []bRecord
	if err := json.Unmarshal(b, &recs); err != nil {
		t.Fatalf("解析 %s/%s: %v", dir, file, err)
	}
	sort.SliceStable(recs, func(i, j int) bool { return recs[i].Timestamp < recs[j].Timestamp })
	return recs
}

// bLayout 读 room_layout.json → cfg(ParseLayoutConfig)+ radar addr(layout radar 映射首值)+ roomID。
func bLayout(t *testing.T, dir string) (RoomConfig, string, string) {
	raw, err := os.ReadFile(filepath.Join(casesDir, dir, "room_layout.json"))
	if err != nil {
		t.Skipf("layout 缺失 %s: %v", dir, err)
	}
	// 双壳:cd2b export = {room_id, room_name, layout_config{...}};裸 canvas 直接是 layout_config。
	var wrap struct {
		RoomID       string          `json:"room_id"`
		LayoutConfig json.RawMessage `json:"layout_config"`
	}
	body := raw
	roomID := "fd00:0:3:112:3:100::/88"
	if json.Unmarshal(raw, &wrap) == nil && len(wrap.LayoutConfig) > 0 {
		body = wrap.LayoutConfig
		if wrap.RoomID != "" {
			roomID = wrap.RoomID
		}
	}
	cfg, err := ParseLayoutConfig(roomID, body)
	if err != nil {
		t.Fatalf("ParseLayoutConfig %s: %v", dir, err)
	}
	// radar addr:layout_config.radar = {radar_<n>: <INET>};取首值。
	var lc struct {
		Radar map[string]string `json:"radar"`
	}
	radarAddr := ""
	if json.Unmarshal(body, &lc) == nil {
		for _, v := range lc.Radar {
			radarAddr = v
			break
		}
	}
	if radarAddr == "" {
		t.Fatalf("layout %s 无 radar 映射", dir)
	}
	return cfg, radarAddr, cfg.RoomID
}

// bShadowLog observer 捕到的一条 belief_shadow_* 日志(msg + 关键字段)。
type bShadowLog struct {
	Msg    string
	Fields map[string]interface{}
}

// bReplay setup 真 Engine + 回放 fixture → 返回 observer 捕到的全部 belief_shadow_* 日志。
// **保真**:只喂 raw record 进 handleMessage/handleEventMessage;不手搓 bases。
func bReplay(t *testing.T, dir, file string) []bShadowLog {
	recs := bLoadRecords(t, dir, file)
	cfg, radarAddr, roomID := bLayout(t, dir)

	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	e := NewEngine(nil, logger)
	e.RegisterRoom(cfg)

	// 识别 radar/sleepad device_uid:有 track 记录的 uid = radar;其余 = sleepad。
	radarUID := ""
	for _, r := range recs {
		if r.Category == "track" {
			radarUID = r.DeviceUID
			break
		}
	}
	const sleepadAddr = "fd00:0:3:112:3:100:5111:ad01" // 合成 sleepad 寻址(仅路由用;sleepad obs 按 UID 存)
	// wire 路由:radar/sleepad addr → roomID(RegisterRoom 已设 deviceMounts[roomID]=cfg.Radar 兜底 mount)。
	e.deviceRoom[radarAddr] = roomID
	e.deviceRoom[sleepadAddr] = roomID
	e.deviceMounts[radarAddr] = cfg.Radar // 显式兜底:确保 handleMessage hasMount(RadarAddrs 可能空/异)

	mk := func(addr, devType, topic, cat string, dv []map[string]interface{}, ts int64) rediscommon.StreamMessage {
		dvJSON, _ := json.Marshal(dv)
		return rediscommon.StreamMessage{Values: map[string]interface{}{
			"device_addr": addr,
			"device_type": devType,
			"topic_type":  topic,
			"category":    cat,
			"timestamp":   strconv.FormatInt(ts, 10),
			"dataValue":   string(dvJSON),
		}}
	}

	for _, r := range recs {
		isRadar := r.DeviceUID == radarUID
		addr := radarAddr
		devType := "radar"
		if !isRadar {
			addr = sleepadAddr
			devType = "sleepad"
		}
		switch {
		case r.TopicType == "event":
			// 床/门/人数事件 → 真 event 入口(InBed/LeftBed/EnterRoom/ExitRoom/number_people)。
			e.handleEventMessage(mk(addr, devType, r.TopicType, r.Category, r.DataValue, r.Timestamp))
		case r.Category == "track":
			// radar track 帧 → 真 monitor 入口(ParseRadarTracks→ProcessFrame→beliefShadowTick)。
			e.handleMessage(nil, mk(addr, devType, r.TopicType, r.Category, r.DataValue, r.Timestamp))
		case !isRadar:
			// sleepad monitor 帧(床占用/vital)→ 真 monitor 入口(ProcessSleepadObservation)。
			e.handleMessage(nil, mk(addr, devType, r.TopicType, r.Category, r.DataValue, r.Timestamp))
		default:
			// radar 非 track monitor 子流(heart/activity 等)不入 shadow lost/bed 决策路径,跳过(非 fork:不喂≠手搓)。
		}
	}

	var out []bShadowLog
	allCounts := map[string]int{}
	for _, entry := range logs.All() {
		allCounts[entry.Message]++
		if !strings.HasPrefix(entry.Message, "belief_shadow_") {
			continue
		}
		out = append(out, bShadowLog{Msg: entry.Message, Fields: entry.ContextMap()})
	}
	if os.Getenv("BDIAG") != "" {
		t.Logf("ALL logs %s: %v", dir, allCounts)
	}
	return out
}

// bCase 8 CD2B 卧室 fall 案(C 诊断分类)。
type bCase struct {
	dir   string
	file  string
	class string // α / β / γ
}

var bCases = []bCase{
	{"cd2b-fall-0605-0712", "", "α"},
	{"cd2b-fall-0606-0917", "", "α"},
	{"cd2b-fall-0606-0929", "", "α"},
	{"cd2b-fall-0607-0127", "", "α"},
	{"cd2b-fall-0607-1021", "", "α"},
	{"cd2b-fall-0604-2233", "", "β"},
	{"cd2b-fall-0605-0717", "", "β"},
	{"cd2b-fall-0605-0142", "", "γ"},
}

// bFindWindowFile dir 下唯一的 window json(*_to_*.json)。
func bFindWindowFile(t *testing.T, dir string) string {
	entries, err := os.ReadDir(filepath.Join(casesDir, dir))
	if err != nil {
		t.Skipf("case 目录缺失 %s: %v", dir, err)
	}
	for _, en := range entries {
		n := en.Name()
		if strings.Contains(n, "_to_") && strings.HasSuffix(n, ".json") {
			return n
		}
	}
	t.Skipf("case %s 无 window json", dir)
	return ""
}

// TestBReplaySmoke — B harness 冒烟:每案能跑通真生产路径且 observer 捕到 belief_shadow_* 日志(无 panic)。
// 这是 B 的第一步保真验证(全生产路径走通);分类对账在 TestBReproduceCDiagnosis。
func TestBReplaySmoke(t *testing.T) {
	for _, c := range bCases {
		c := c
		t.Run(c.dir, func(t *testing.T) {
			file := bFindWindowFile(t, c.dir)
			logs := bReplay(t, c.dir, file)
			counts := map[string]int{}
			for _, l := range logs {
				counts[l.Msg]++
			}
			t.Logf("case=%s class=%s shadow_logs=%v", c.dir, c.class, counts)
		})
	}
}
