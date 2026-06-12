// redis-replay：把历史 monitor_stream / event_log / alarm_events 行按原始时序重放回 Redis 实时流，
// 时间戳 rebase 到「当前」（t1 -> now），供 wisefido-sensor / cardagg 当作实时数据消费。
//
// 消费者丢弃 >6s 的消息（sensorMonitorMaxInboundAgeMs=6000），所以必须 rebase + 按真实
// 间隔节奏 sleep 重放，否则整批被当陈旧丢掉。消息信封复用 rediscommon.IoTStreamMessage
// .ToStreamMap()，与真实生产者 XADD 格式完全一致。
//
// 用法（device 级，向后兼容）:
//   go run ./redis-replay/ --device-uids 4D8710D5CABB --t1 "2026-06-05 06:06:44" \
//       --t2 "2026-06-05 06:16:42" --tz Asia/Shanghai [--streams monitor,event] [--dry-run]
//
// 用法（unit 级，自动拉所有设备）:
//   go run ./redis-replay/ --unit fd00:0:3:112 --t1 ... --t2 ... --tz America/Denver
//
// 环境变量: DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME(默认 owl_v2)
//           REDIS_ADDR(默认 localhost:6379)/REDIS_PASSWORD(默认 TeLunSu-36kr)

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"net/netip"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	commonconfig "owl-common/config"
	database "owl-common/database"
	rediscommon "owl-common/redis"

	"github.com/go-redis/redis/v8"
	"github.com/lib/pq"
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
	loadDotEnv() // owlBack/.env → 填未设的 DB_*/REDIS_* env（密码不硬编码，单源 .env）

	var (
		uidsArg   = flag.String("device-uids", "", "逗号分隔 device_uid 列表（与 --unit 二选一；DB 模式）")
		unitArg   = flag.String("unit", "", "spatial prefix(fd00:0:3:112 或 201)，自动拉该 unit 所有设备（DB 模式）")
		fixtureA  = flag.String("fixture", "", "case 目录（含 window.json[+window_sleepad.json][+meta.json]），从文件回放（与 --device/--unit 互斥；不查 monitor_stream）")
		t1Arg     = flag.String("t1", "", "起始本地时间 \"YYYY-MM-DD HH:MM:SS\"（DB 模式必填，按 --tz 解释）")
		t2Arg     = flag.String("t2", "", "结束本地时间（DB 模式必填）")
		tzArg     = flag.String("tz", "", "IANA 时区，如 Asia/Shanghai（DB 模式必填）")
		speed     = flag.Float64("speed", 1.0, "重放倍速（>1 加快，<1 放慢）")
		streamSel = flag.String("streams", "monitor,event", "重放哪些流：monitor,event,alarm 逗号组合")
		dryRun    = flag.Bool("dry-run", false, "只打印不实际 XADD")
	)
	flag.Parse()

	wantMonitor := strings.Contains(*streamSel, "monitor")
	wantEvent := strings.Contains(*streamSel, "event")
	wantAlarm := strings.Contains(*streamSel, "alarm")
	if !wantMonitor && !wantEvent && !wantAlarm {
		fatal("--streams 至少含 monitor、event 或 alarm")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var (
		rows  []record
		addrs []string
		loc   = time.Local
		t1    time.Time
	)

	if *fixtureA != "" {
		// ── 文件模式：数据来自 fixture window.json（不查 monitor_stream）。addr 经 meta.json，
		//    无 meta 则 DB fallback（uid→addr，devices 表，轻量一次查，密码走 .env）。
		if *uidsArg != "" || *unitArg != "" {
			fatal("--fixture 与 --device-uids/--unit 互斥")
		}
		var err error
		rows, addrs, err = loadFixtureRecords(ctx, *fixtureA, wantMonitor, wantEvent, wantAlarm)
		if err != nil {
			fatal("加载 fixture 失败: %v", err)
		}
		if len(rows) == 0 {
			fatal("fixture 内无匹配 stream 的记录：%s", *fixtureA)
		}
		sort.Slice(rows, func(i, j int) bool { return rows[i].tsMs < rows[j].tsMs })
		t1 = time.UnixMilli(rows[0].tsMs)
	} else {
		// ── DB 模式（向后兼容）：--device/--unit + t1/t2/tz 查 monitor_stream。
		if *uidsArg == "" && *unitArg == "" {
			fmt.Fprintln(os.Stderr, "ERROR: 必须指定 --fixture 或 --device-uids 或 --unit")
			flag.Usage()
			os.Exit(1)
		}
		if *uidsArg != "" && *unitArg != "" {
			fatal("--device-uids 和 --unit 不能同时指定")
		}
		if *t1Arg == "" || *t2Arg == "" || *tzArg == "" {
			flag.Usage()
			os.Exit(1)
		}
		var err error
		loc, err = time.LoadLocation(*tzArg)
		if err != nil {
			fatal("加载时区 %q 失败: %v", *tzArg, err)
		}
		t1, err = time.ParseInLocation(localLayout, *t1Arg, loc)
		if err != nil {
			fatal("解析 t1 %q 失败: %v", *t1Arg, err)
		}
		t2, err := time.ParseInLocation(localLayout, *t2Arg, loc)
		if err != nil {
			fatal("解析 t2 %q 失败: %v", *t2Arg, err)
		}
		if !t2.After(t1) {
			fatal("t2 必须晚于 t1")
		}

		db, err := database.NewPostgresDB(&commonconfig.DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Database: getEnv("DB_NAME", "owl_v2"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		})
		if err != nil {
			fatal("连接 DB 失败: %v", err)
		}
		defer db.Close()

		dtypeByAddr := map[string]string{}
		if *unitArg != "" {
			prefix := resolveUnitPrefix(*unitArg)
			addrs, dtypeByAddr = discoverUnitDevices(ctx, db, prefix, t1.UTC(), t2.UTC())
			if len(addrs) == 0 {
				fatal("unit %q 在窗口内无活跃设备（prefix=%s）", *unitArg, prefix)
			}
		} else {
			addrs, dtypeByAddr = resolveDevices(ctx, db, splitCSV(*uidsArg))
			if len(addrs) == 0 {
				fatal("没有解析到任何 device_addr（device_uids=%v）", splitCSV(*uidsArg))
			}
		}
		rows, err = loadRecords(ctx, db, addrs, dtypeByAddr, t1.UTC(), t2.UTC(), wantMonitor, wantEvent, wantAlarm)
		if err != nil {
			fatal("查询时序失败: %v", err)
		}
		if len(rows) == 0 {
			fatal("窗口内无数据：%s ~ %s (%s)", *t1Arg, *t2Arg, *tzArg)
		}
		sort.Slice(rows, func(i, j int) bool { return rows[i].tsMs < rows[j].tsMs })
	}

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
	fmt.Printf("重放 %d 条 (%s) | 设备 %d 台 | 窗口 %s~%s | 倍速 %.2gx | rebase t1->now%s\n",
		len(rows), strings.Join(streamNames, "+"), len(addrs), winFrom, winTo, *speed, dryRunTag(*dryRun))

	t1ms := t1.UnixMilli()
	startWall := time.Now()
	seq := map[string]uint64{}
	var nMon, nEvt, nAlarm int

	for _, r := range rows {
		rel := time.Duration(float64(r.tsMs-t1ms)/(*speed)) * time.Millisecond
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
			Timestamp:      time.Now().UnixMilli(),
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

		rel0 := time.UnixMilli(r.tsMs).In(loc).Format("15:04:05")
		line := fmt.Sprintf("  [%s] %-12s %-7s %-14s %s", rel0, short(r.addr), r.topicType, r.category, summary(r.payload))

		if *dryRun {
			fmt.Println(line)
		} else {
			if _, err := rediscommon.PublishToStream(ctx, rdb, def.Name, msg.ToStreamMap(), def.MaxLen, def.RetentionSeconds); err != nil {
				fmt.Fprintf(os.Stderr, "  XADD 失败 %s: %v\n", def.Name, err)
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

// resolveUnitPrefix 将简写 unit 名（"201" / "fd00:0:3:112"）归一化为 /64 INET prefix。
func resolveUnitPrefix(raw string) string {
	s := strings.TrimSpace(raw)
	// "201" 简写 → fd00:0:3:201::/64 (B2B unit，前 4 组 = /64)
	if len(s) <= 4 && s[0] >= '1' && s[0] <= '9' {
		return fmt.Sprintf("fd00:0:3:%s::/64", s)
	}
	// 已是 IP prefix 但没有 CIDR → 补 /64
	if !strings.Contains(s, "/") {
		return s + "::/64"
	}
	return s
}

// discoverUnitDevices 从 monitor_stream 查该 INET prefix 下 window 内所有活跃设备 addr。
func discoverUnitDevices(ctx context.Context, db *sql.DB, prefix string, t1, t2 time.Time) ([]string, map[string]string) {
	addrs := []string{}
	dtype := map[string]string{}
	rows, err := db.QueryContext(ctx,
		`SELECT DISTINCT host(device_addr), device_type::text FROM monitor_stream
		 WHERE device_addr <<= $1::inet AND ts BETWEEN $2 AND $3`, prefix, t1, t2)
	if err != nil {
		fatal("discoverUnitDevices: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var addr, dt string
		if err := rows.Scan(&addr, &dt); err != nil {
			fatal("scan unit device: %v", err)
		}
		addrs = append(addrs, addr)
		dtype[addr] = dt
	}
	return addrs, dtype
}

// resolveDevices: device_uid -> device_addr，并取 device_type。
func resolveDevices(ctx context.Context, db *sql.DB, uids []string) ([]string, map[string]string) {
	addrs := []string{}
	dtype := map[string]string{}
	rows, err := db.QueryContext(ctx,
		`SELECT host(device_addr), device_uid FROM devices WHERE device_uid = ANY($1)`, pq.Array(uids))
	if err != nil {
		fatal("查 devices 失败: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var addr, uid string
		if err := rows.Scan(&addr, &uid); err != nil {
			fatal("scan devices: %v", err)
		}
		addrs = append(addrs, addr)
	}
	for _, a := range addrs {
		var dt sql.NullString
		_ = db.QueryRowContext(ctx,
			`SELECT device_type::text FROM monitor_stream WHERE device_addr=$1::inet ORDER BY ts DESC LIMIT 1`, a).Scan(&dt)
		if dt.Valid && dt.String != "" {
			dtype[a] = dt.String
		} else {
			dtype[a] = "Radar"
		}
	}
	return addrs, dtype
}

func loadRecords(ctx context.Context, db *sql.DB, addrs []string, dtype map[string]string, t1, t2 time.Time, wantMon, wantEvt, wantAlarm bool) ([]record, error) {
	out := []record{}
	if wantMon {
		rows, err := db.QueryContext(ctx, `
			SELECT (extract(epoch FROM ts)*1000)::bigint, host(device_addr), device_type::text, stream_type, payload::text
			FROM monitor_stream WHERE device_addr = ANY($1::inet[]) AND ts BETWEEN $2 AND $3`,
			pq.Array(addrs), t1, t2)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var ms int64
			var addr, dt, st, pl string
			if err := rows.Scan(&ms, &addr, &dt, &st, &pl); err != nil {
				rows.Close()
				return nil, err
			}
			cat := st
			if i := strings.IndexByte(st, '.'); i >= 0 {
				cat = st[i+1:]
			}
			out = append(out, record{tsMs: ms, addr: addr, dtype: dt, topicType: "monitor", category: cat, payload: mustArray(pl)})
		}
		rows.Close()
	}
	if wantEvt {
		rows, err := db.QueryContext(ctx, `
			SELECT (extract(epoch FROM ts)*1000)::bigint, host(device_addr), event_kind, payload::text
			FROM event_log WHERE device_addr = ANY($1::inet[]) AND ts BETWEEN $2 AND $3`,
			pq.Array(addrs), t1, t2)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var ms int64
			var addr, ek, pl string
			if err := rows.Scan(&ms, &addr, &ek, &pl); err != nil {
				rows.Close()
				return nil, err
			}
			out = append(out, record{tsMs: ms, addr: addr, dtype: dtype[addr], topicType: "event", category: ek, payload: mustArray(pl)})
		}
		rows.Close()
	}
	if wantAlarm {
		// alarm_events:仅重放 device-gateway 直发的 alarm(producer=设备 addr，非 sensor platform-agent)。
		// sensor 自己发的 alarm(sensor producer=fd00:0:fff1::1)不应重放——sensor 应在 replay 中自己产出。
		rows, err := db.QueryContext(ctx, `
			SELECT (extract(epoch FROM triggered_at)*1000)::bigint, host(device_addr), event_type, payload, evidence
			FROM alarm_events WHERE device_addr = ANY($1::inet[]) AND triggered_at BETWEEN $2 AND $3
			  AND producer != 'fd00:0:fff1::1'
			ORDER BY triggered_at`,
			pq.Array(addrs), t1, t2)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var ms int64
			var addr, et, pl string
			var ev sql.NullString
			if err := rows.Scan(&ms, &addr, &et, &pl, &ev); err != nil {
				rows.Close()
				return nil, err
			}
			payload := mustMap(pl)
			if ev.Valid && ev.String != "" && ev.String != "{}" {
				payload["evidence"] = json.RawMessage(ev.String)
			}
			out = append(out, record{tsMs: ms, addr: addr, dtype: dtype[addr], topicType: "alarm", category: et, payload: []interface{}{payload}})
		}
		rows.Close()
	}
	return out, nil
}

// ── 文件模式（fixture）──────────────────────────────────────────────────────

// loadDotEnv 从 owlBack/.env 读 KEY=VALUE 填入**未设**的环境变量（DB/Redis 密码单源 .env，工具不硬编码）。
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

// fixtureRec 是 window.json / window_sleepad.json 单条记录（与 export / lego 产出同格式）。
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
func fixtureTopicType(cat string) string {
	switch cat {
	case "track", "heart":
		return "monitor"
	default:
		return "event"
	}
}

// loadFixtureRecords 从 case 目录读 window.json[+window_sleepad.json] → []record（不查 monitor_stream）。
// uid→addr 走 meta.json，无则 DB fallback（devices 表，密码 .env）。
func loadFixtureRecords(ctx context.Context, dir string, wantMon, wantEvt, wantAlarm bool) ([]record, []string, error) {
	meta, err := loadFixtureMeta(ctx, dir)
	if err != nil {
		return nil, nil, err
	}
	var out []record
	addrSet := map[string]bool{}
	readOne := func(name string) error {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil // 可选文件（window_sleepad 可能不存在）
		}
		var recs []fixtureRec
		if err := json.Unmarshal(b, &recs); err != nil {
			return fmt.Errorf("%s 解析失败: %w", name, err)
		}
		for _, r := range recs {
			tt := fixtureTopicType(r.Category)
			if (tt == "monitor" && !wantMon) || (tt == "event" && !wantEvt) {
				continue
			}
			dm := meta[r.DeviceUID]
			if dm.addr == "" {
				return fmt.Errorf("device_uid=%s 无 addr 映射（meta.json 缺且 DB 未命中）", r.DeviceUID)
			}
			out = append(out, record{tsMs: r.Timestamp, addr: dm.addr, dtype: dm.dtype, topicType: tt, category: r.Category, payload: r.DataValue})
			addrSet[dm.addr] = true
		}
		return nil
	}
	if err := readOne("window.json"); err != nil {
		return nil, nil, err
	}
	if err := readOne("window_sleepad.json"); err != nil {
		return nil, nil, err
	}
	addrs := make([]string, 0, len(addrSet))
	for a := range addrSet {
		addrs = append(addrs, a)
	}
	return out, addrs, nil
}

// loadFixtureMeta uid→{addr,dtype}：meta.json {"devices":[{device_uid,device_addr,device_type}]} 提供已知；
// 未覆盖的 uid（如配对 sleepad）走 DB 补缺（addr + device_type）。meta 全 → 纯文件；meta 部分/无 → DB 补。
func loadFixtureMeta(ctx context.Context, dir string) (map[string]devMeta, error) {
	out := map[string]devMeta{}
	if b, err := os.ReadFile(filepath.Join(dir, "meta.json")); err == nil {
		var m struct {
			Devices []struct {
				DeviceUID  string `json:"device_uid"`
				DeviceAddr string `json:"device_addr"`
				DeviceType string `json:"device_type"`
			} `json:"devices"`
		}
		if json.Unmarshal(b, &m) == nil {
			for _, d := range m.Devices {
				if d.DeviceUID != "" && d.DeviceAddr != "" {
					dt := d.DeviceType
					if dt == "" {
						dt = "Radar"
					}
					out[d.DeviceUID] = devMeta{addr: d.DeviceAddr, dtype: dt}
				}
			}
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
	if len(missing) == 0 {
		return out, nil // meta 全覆盖 → 纯文件
	}
	db, err := database.NewPostgresDB(&commonconfig.DatabaseConfig{
		Host: getEnv("DB_HOST", "localhost"), Port: getEnvInt("DB_PORT", 5432),
		User: getEnv("DB_USER", "postgres"), Password: getEnv("DB_PASSWORD", "postgres"),
		Database: getEnv("DB_NAME", "owl_v2"), SSLMode: getEnv("DB_SSLMODE", "disable"),
	})
	if err != nil {
		return nil, fmt.Errorf("meta.json 未覆盖 %v，DB 补缺连接失败: %w", missing, err)
	}
	defer db.Close()
	rows, err := db.QueryContext(ctx, `
		SELECT d.device_uid, host(d.device_addr),
		  COALESCE((SELECT m.device_type::text FROM monitor_stream m WHERE m.device_addr=d.device_addr ORDER BY ts DESC LIMIT 1), 'Radar')
		FROM devices d WHERE d.device_uid = ANY($1)`, pq.Array(missing))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var u, a, t string
		if rows.Scan(&u, &a, &t) == nil {
			out[u] = devMeta{addr: a, dtype: t}
		}
	}
	for _, u := range missing {
		if out[u].addr == "" {
			return nil, fmt.Errorf("device_uid=%s 在 devices 表未找到（补 meta.json 或确认 DB）", u)
		}
	}
	return out, nil
}

// scanFixtureUIDs 扫 window.json + window_sleepad.json 的 distinct device_uid。
func scanFixtureUIDs(dir string) []string {
	set := map[string]bool{}
	for _, name := range []string{"window.json", "window_sleepad.json"} {
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

func mustArray(jsonStr string) []interface{} {
	var arr []interface{}
	if err := json.Unmarshal([]byte(jsonStr), &arr); err == nil {
		return arr
	}
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &obj); err == nil {
		return []interface{}{obj}
	}
	return []interface{}{}
}

func mustMap(jsonStr string) map[string]interface{} {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &m); err == nil {
		return m
	}
	return map[string]interface{}{"payload": jsonStr}
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

func splitCSV(s string) []string {
	out := []string{}
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
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

func getEnvInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
	os.Exit(1)
}
