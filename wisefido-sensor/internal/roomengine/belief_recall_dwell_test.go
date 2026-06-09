package roomengine

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"
	"time"

	rediscommon "owl-common/redis"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// belief_recall_dwell_test.go — Tier-1 真 recall（DBN 抓 firmware **漏**的摔）：
// #1 bedtest-0605-1（firmware 零 pose=5 未判出）→ 喂真 track 帧走真 pipeline，诊断 DBN P(Fallen)。
// 委员会授权（#9 firmware 判出=易方向补不了真 recall）。room_layout.json 从 owl_v2 room_visual_layout 导（真，已提交）。
//
// ★真 pose 分布（亲查 txt，改正初稿失实「pose≈3 全程」）：pose=4 走 241 / **pose=6 卧 128** / pose=3 坐 119 / pose=1 36。
// ★卧姿帧 area=1（**126/128 在 GeomInEnter 门区**，非床区 area=2）——床边摔被雷达**误归 Enter 区**；poseLying@Enter
// 走 likelihood default 档（modest LR + SBedLying 竞争），又与 241 走动帧混杂 → 信号冲淡。诊断型（no-silent-caps）：
// 报真实可复现 peak，不硬断言 fire——结果（DBN 抓不抓得到、为何）本身是发现。
//
// txt→生产 StreamMessage 转换器：解析 test_record.txt 的 radar track 行。

var txtTrackRe = regexp.MustCompile(`^(\d\d:\d\d:\d\d)\s*\|\s*:9e7\s*\|\s*track\s*\|\s*pose=(-?\d+)\s+xy=\((-?\d+),(-?\d+)\)\s+z=(-?\d+)\s+tid=(\d+)\s+area=(-?\d+)\s+conf=(\d+)`)

func TestRecallRealFall_FirmwareMissed_Bedtest1(t *testing.T) {
	dir := "bedtest-0605-1-bedside-fall-no-fw-detect"
	cfg, radarAddr, roomID := bLayout(t, dir)

	f, err := os.Open(filepath.Join(casesDir, dir, "test_record.txt"))
	if err != nil {
		t.Skipf("test_record.txt 缺: %v", err)
	}
	defer f.Close()

	core, logs := observer.New(zapcore.DebugLevel)
	e := NewEngine(nil, zap.New(core))
	e.RegisterRoom(cfg)
	e.deviceRoom[radarAddr] = roomID
	e.deviceMounts[radarAddr] = cfg.Radar

	mk := func(dv []map[string]interface{}, ts int64) rediscommon.StreamMessage {
		dvJSON := fmt.Sprintf(`[{"pose":%v,"position_x":%v,"position_y":%v,"position_z":%v,"track_id":%v,"area_id":%v,"track_confidence":%v}]`,
			dv[0]["pose"], dv[0]["position_x"], dv[0]["position_y"], dv[0]["position_z"], dv[0]["track_id"], dv[0]["area_id"], dv[0]["track_confidence"])
		return rediscommon.StreamMessage{Values: map[string]interface{}{
			"device_addr": radarAddr, "device_type": "radar", "topic_type": "monitor", "category": "track",
			"timestamp": strconv.FormatInt(ts, 10), "dataValue": dvJSON,
		}}
	}

	const day = "2026-06-05 "
	atoi := func(s string) int { n, _ := strconv.Atoi(s); return n }
	rows := 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		m := txtTrackRe.FindStringSubmatch(sc.Text())
		if m == nil {
			continue
		}
		tt, err := time.Parse("2006-01-02 15:04:05", day+m[1])
		if err != nil {
			continue
		}
		ts := tt.UTC().UnixMilli()
		e.handleMessage(nil, mk([]map[string]interface{}{{
			"pose": atoi(m[2]), "position_x": atoi(m[3]), "position_y": atoi(m[4]), "position_z": atoi(m[5]),
			"track_id": atoi(m[6]), "area_id": atoi(m[7]), "track_confidence": atoi(m[8]),
		}}, ts))
		rows++
	}

	fired := logs.FilterMessage("belief_shadow_fall").Len()
	var peak, lastP float64
	for _, le := range logs.All() {
		if le.Message == "belief_shadow_trace" {
			if v, ok := le.ContextMap()["p_fallen"].(float64); ok {
				lastP = v
				if v > peak {
					peak = v
				}
			}
		}
	}
	t.Logf("#1 firmware-漏真摔(pose分布 walk241/卧128/sit119;卧帧area=1=GeomInEnter门区)：喂 %d 帧 → belief_shadow_fall=%d  peak P(Fallen)=%.3f  末态=%.3f", rows, fired, peak, lastP)
	if rows == 0 {
		t.Fatalf("txt 解析 0 帧 → 转换器/正则失效")
	}
	// 诊断型：不硬断言 fire（最硬案，结果是发现）。仅锁"管道通+有帧"，peak 留报告。
}
