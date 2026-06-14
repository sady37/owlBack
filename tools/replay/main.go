// redis-replay：把 case fixture（window.json[+window_sleepad.json][+alarm.json]+meta.json）按原始时序
// 重放回 Redis 实时流，时间戳 rebase 到「当前」（t1 -> now），供 wisefido-sensor / cardagg 当实时数据消费。
//
// replay 只读文件、绝不连库：fixture 由 tools/export（PG→文件）或 tools/simulate-make（合成）产出，
// uid→addr→type 全经 meta.json（导出侧已烤进）。本工具零 DB 依赖（不 import database 包即编译期证明）。
//
// 消费者丢弃 >6s 的消息（sensorMonitorMaxInboundAgeMs=6000），所以必须 rebase + 按真实
// 间隔节奏 sleep 重放，否则整批被当陈旧丢掉。消息信封复用 rediscommon.IoTStreamMessage
// .ToStreamMap()，与真实生产者 XADD 格式完全一致。
//
// 用法:
//   go run ./replay/ --fixture ../doc/cases/case-cd2b-0606-10271037 \
//       [--streams monitor,event,alarm] [--speed 1] [--dry-run]
//
// 环境变量: REDIS_ADDR(默认 localhost:6379)/REDIS_PASSWORD(默认走 owlBack/.env)

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/netip"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	commonconfig "owl-common/config"
	rediscommon "owl-common/redis"

	"github.com/go-redis/redis/v8"
)

const localLayout = "2006-01-02 15:04:05"

type record struct {
	tsMs      int64
	addr      string
	dtype     string // "Radar" / "Sleepad"
	topicType string // "monitor" / "event" / "alarm"
	category  string // "track" / "heart" / "EnterRoom" / "Fall" / ...
	payload   []interface{}
}

func main() {
	loadDotEnv() // owlBack/.env → 填未设的 REDIS_* env（密码不硬编码，单源 .env）

	var (
		fixtureA  = flag.String("fixture", "", "case 目录（含 window.json[+window_sleepad.json][+alarm.json]+meta.json），纯文件回放")
		speed     = flag.Float64("speed", 1.0, "重放倍速（>1 加快，<1 放慢；验 fire 行为必须 1）")
		streamSel = flag.String("streams", "monitor,event", "重放哪些流：monitor,event,alarm 逗号组合")
		streamPfx = flag.String("stream-prefix", "", "stream 名前缀（喂 Tsensor 用 \"test:\" → 推 test:iot:*:stream；默认空=生产 iot:*）")
		dryRun    = flag.Bool("dry-run", false, "只打印不实际 XADD")
	)
	flag.Parse()

	wantMonitor := strings.Contains(*streamSel, "monitor")
	wantEvent := strings.Contains(*streamSel, "event")
	wantAlarm := strings.Contains(*streamSel, "alarm")
	if !wantMonitor && !wantEvent && !wantAlarm {
		fatal("--streams 至少含 monitor、event 或 alarm")
	}
	if *fixtureA == "" {
		fmt.Fprintln(os.Stderr, "ERROR: 必须指定 --fixture <case 目录>（replay 只读文件；导出走 tools/export）")
		flag.Usage()
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	loc := time.UTC // 时间戳全程 UTC：fixture 存 UTC epoch ms，replay 按绝对 ms rebase，显示也走 UTC（不引服务器本地时区）
	rows, addrs, err := loadFixtureRecords(*fixtureA, wantMonitor, wantEvent, wantAlarm)
	if err != nil {
		fatal("加载 fixture 失败: %v", err)
	}
	if len(rows) == 0 {
		fatal("fixture 内无匹配 stream 的记录：%s", *fixtureA)
	}
	// SliceStable：同毫秒并列时保持文件追加序(window→window_sleepad→alarm)，回放可复现；对齐 lego.go
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].tsMs < rows[j].tsMs })

	var rdb *redis.Client
	if !*dryRun {
		rdb = rediscommon.NewRedisClient(&commonconfig.RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", "TeLunSu-36kr"),
			DB:       0,
		})
		defer rediscommon.Close(rdb)
		if err := rediscommon.Ping(ctx, rdb); err != nil {
			fatal("连接 Redis 失败: %v（设 REDIS_ADDR/REDIS_PASSWORD）", err)
		}
	}

	streamNames := []string{}
	if wantMonitor {
		streamNames = append(streamNames, "monitor")
	}
	if wantEvent {
		streamNames = append(streamNames, "event")
	}
	if wantAlarm {
		streamNames = append(streamNames, "alarm")
	}
	winFrom := time.UnixMilli(rows[0].tsMs).In(loc).Format(localLayout)
	winTo := time.UnixMilli(rows[len(rows)-1].tsMs).In(loc).Format(localLayout)
	fmt.Printf("重放 %d 条 (%s) | 设备 %d 台 | 窗口 %s~%s UTC | 倍速 %.2gx | rebase t1->now%s\n",
		len(rows), strings.Join(streamNames, "+"), len(addrs), winFrom, winTo, *speed, dryRunTag(*dryRun))
	if *speed != 1.0 {
		fmt.Printf("⚠️  倍速 %.2gx ≠ 1：sensor 的 confirmMs/dwell/decay 按真实墙钟计,加速会压缩时间窗 → fire/confirm 判断失真,仅供数据流连通性冒烟。验 fire 行为必须 --speed 1。\n", *speed)
	}

	baseMs := rows[0].tsMs // 基准 t0 = 全局最早一帧（radar/sleepad 合并排序后第一条）；每条 fire 在 now + (ts−t0)/speed
	startWall := time.Now()
	startWallMs := startWall.UnixMilli() // 方案一(虚拟时钟):时间戳盖 startWall+(ts−t0) 保真 dt,允许超前墙钟
	seq := map[string]uint64{}
	var nMon, nEvt, nAlarm int

	for _, r := range rows {
		rel := time.Duration(float64(r.tsMs-baseMs)/(*speed)) * time.Millisecond
		if d := time.Until(startWall.Add(rel)); d > 0 {
			select {
			case <-ctx.Done():
				fmt.Printf("\n中断：已重放 monitor=%d event=%d alarm=%d\n", nMon, nEvt, nAlarm)
				return
			case <-time.After(d):
			}
		}

		addr, _ := netip.ParseAddr(r.addr)
		seq[r.addr]++
		msg := rediscommon.IoTStreamMessage{
			Producer:       r.addr, // device-direct producer = device_addr
			SequenceNumber: seq[r.addr],
			DeviceAddr:     addr,
			DeviceType:     r.dtype,
			SubjectEntity:  r.addr, // 生产 raw envelope subject 非空(=device_uid);PG 未存,回放补 device_addr 过 adapter_radar 非空闸(否则 radar 事件被 zoneengine 丢)
			Timestamp:      startWallMs + (r.tsMs - baseMs), // 方案一:保真 dt(原始数据间距),发送节奏仍按 speed 加速 → 虚拟时间轴
			TopicType:      r.topicType,
			Category:       r.category,
			DataValue:      r.payload,
		}
		def := rediscommon.StreamMonitor
		if r.topicType == "event" {
			def = rediscommon.StreamEvent
		} else if r.topicType == "alarm" {
			def = rediscommon.StreamAlarm
		}
		streamName := *streamPfx + def.Name // 前缀=""→生产 iot:*;"test:"→喂 Tsensor 订的 test:iot:*

		rel0 := time.UnixMilli(r.tsMs).In(loc).Format("15:04:05")
		line := fmt.Sprintf("  [%s] %-12s %-7s %-14s %s", rel0, short(r.addr), r.topicType, r.category, summary(r.payload))

		if *dryRun {
			fmt.Println(line)
		} else {
			if _, err := rediscommon.PublishToStream(ctx, rdb, streamName, msg.ToStreamMap(), def.MaxLen, def.RetentionSeconds); err != nil {
				fmt.Fprintf(os.Stderr, "  XADD 失败 %s: %v\n", streamName, err)
				continue
			}
			fmt.Println(line)
		}
		switch r.topicType {
		case "event":
			nEvt++
		case "alarm":
			nAlarm++
		default:
			nMon++
		}
	}
	fmt.Printf("完成：monitor=%d event=%d alarm=%d\n", nMon, nEvt, nAlarm)
}

// ── 文件模式（fixture）──────────────────────────────────────────────────────

// loadDotEnv 从 owlBack/.env 读 KEY=VALUE 填入**未设**的环境变量（Redis 密码单源 .env，工具不硬编码）。
// tools/replay 运行 cwd 不定 → 向上逐级找第一个 .env。
func loadDotEnv() {
	for _, p := range []string{".env", "../.env", "../../.env", "../../../.env"} {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(b), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			k, v, ok := strings.Cut(line, "=")
			if !ok {
				continue
			}
			k, v = strings.TrimSpace(k), strings.TrimSpace(v)
			if _, set := os.LookupEnv(k); !set {
				os.Setenv(k, v)
			}
		}
		return
	}
}

// fixtureRec 是 window.json / window_sleepad.json / alarm.json 单条记录（与 export / simulate-make 产出同格式）。
type fixtureRec struct {
	Category  string        `json:"category"`
	DeviceUID string        `json:"device_uid"`
	Timestamp int64         `json:"timestamp"`
	DataValue []interface{} `json:"data_value"`
}

// devMeta uid → addr + 物理类型（Radar/Sleepad）。dtype 由 meta 权威定，不靠文件来源
// （window.json 也可能含 sleepad 的 event，标 Radar 会被 radar adapter 错处理）。
type devMeta struct {
	addr  string
	dtype string
}

// fixtureTopicType category → topic_type（monitor=track/heart；其余=event）。
// alarm 不靠 category 推（Fall/LeftBed 会被推成 event），由 alarm.json 文件来源强制为 alarm。
func fixtureTopicType(cat string) string {
	switch cat {
	case "track", "heart":
		return "monitor"
	default:
		return "event"
	}
}

// loadFixtureRecords 从 case 目录读 window.json[+window_sleepad.json][+alarm.json] → []record（只读文件）。
// uid→addr 走 meta.json（导出侧权威，无 DB 兜底）。
func loadFixtureRecords(dir string, wantMon, wantEvt, wantAlarm bool) ([]record, []string, error) {
	meta, err := loadFixtureMeta(dir)
	if err != nil {
		return nil, nil, err
	}
	var out []record
	addrSet := map[string]bool{}
	// forceTopic=="" → 按 category 推导（window*.json）；非空 → 强制（alarm.json 来源即 topic）。
	read := func(name, forceTopic string) error {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil // 可选文件（window_sleepad / alarm 可能不存在）
		}
		var recs []fixtureRec
		if err := json.Unmarshal(b, &recs); err != nil {
			return fmt.Errorf("%s 解析失败: %w", name, err)
		}
		for _, r := range recs {
			tt := forceTopic
			if tt == "" {
				tt = fixtureTopicType(r.Category)
			}
			if (tt == "monitor" && !wantMon) || (tt == "event" && !wantEvt) || (tt == "alarm" && !wantAlarm) {
				continue
			}
			dm := meta[r.DeviceUID]
			if dm.addr == "" {
				return fmt.Errorf("device_uid=%s 无 addr 映射（meta.json 缺该设备）", r.DeviceUID)
			}
			out = append(out, record{tsMs: r.Timestamp, addr: dm.addr, dtype: dm.dtype, topicType: tt, category: r.Category, payload: r.DataValue})
			addrSet[dm.addr] = true
		}
		return nil
	}
	if err := read("window.json", ""); err != nil {
		return nil, nil, err
	}
	if err := read("window_sleepad.json", ""); err != nil {
		return nil, nil, err
	}
	if err := read("alarm.json", "alarm"); err != nil {
		return nil, nil, err
	}
	addrs := make([]string, 0, len(addrSet))
	for a := range addrSet {
		addrs = append(addrs, a)
	}
	return out, addrs, nil
}

// loadFixtureMeta 读 meta.json {"devices":[{device_uid,device_addr,device_type}]} → uid→{addr,dtype}。
// 导出侧（tools/export）必须写全 unit 内活跃设备；未覆盖即报错，绝不回查 DB（replay 只读文件）。
func loadFixtureMeta(dir string) (map[string]devMeta, error) {
	b, err := os.ReadFile(filepath.Join(dir, "meta.json"))
	if err != nil {
		return nil, fmt.Errorf("读 meta.json 失败（uid→addr 由导出侧 meta.json 提供，replay 不连库）: %w", err)
	}
	var m struct {
		Devices []struct {
			DeviceUID  string `json:"device_uid"`
			DeviceAddr string `json:"device_addr"`
			DeviceType string `json:"device_type"`
		} `json:"devices"`
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("meta.json 解析失败: %w", err)
	}
	out := map[string]devMeta{}
	for _, d := range m.Devices {
		if d.DeviceUID != "" && d.DeviceAddr != "" {
			dt := d.DeviceType
			if dt == "" {
				dt = "Radar"
			}
			out[d.DeviceUID] = devMeta{addr: d.DeviceAddr, dtype: dt}
		}
	}
	uids := scanFixtureUIDs(dir)
	if len(uids) == 0 {
		return nil, fmt.Errorf("fixture 无 device_uid（window.json 空？）")
	}
	var missing []string
	for _, u := range uids {
		if out[u].addr == "" {
			missing = append(missing, u)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("meta.json 未覆盖设备 %v（导出侧 tools/export 应写全 unit 活跃设备）", missing)
	}
	return out, nil
}

// scanFixtureUIDs 扫 window.json + window_sleepad.json + alarm.json 的 distinct device_uid。
func scanFixtureUIDs(dir string) []string {
	set := map[string]bool{}
	for _, name := range []string{"window.json", "window_sleepad.json", "alarm.json"} {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		var recs []fixtureRec
		if json.Unmarshal(b, &recs) == nil {
			for _, r := range recs {
				if r.DeviceUID != "" {
					set[r.DeviceUID] = true
				}
			}
		}
	}
	uids := make([]string, 0, len(set))
	for u := range set {
		uids = append(uids, u)
	}
	return uids
}

func summary(payload []interface{}) string {
	if len(payload) == 0 {
		return "{}"
	}
	m, ok := payload[0].(map[string]interface{})
	if !ok {
		return fmt.Sprintf("%v", payload[0])
	}
	keys := []string{"track_id", "pose", "position_x", "position_y", "position_z", "number_people", "event_status", "bed_status"}
	parts := []string{}
	for _, k := range keys {
		if v, ok := m[k]; ok {
			parts = append(parts, fmt.Sprintf("%s=%v", k, v))
		}
	}
	if len(parts) == 0 {
		return fmt.Sprintf("len=%d", len(payload))
	}
	return strings.Join(parts, " ")
}

func short(addr string) string {
	if i := strings.LastIndexByte(addr, ':'); i >= 0 {
		return addr[i+1:]
	}
	return addr
}

func dryRunTag(d bool) string {
	if d {
		return " [DRY-RUN]"
	}
	return ""
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
	os.Exit(1)
}
