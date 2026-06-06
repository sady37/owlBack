// redis-replay：把历史 monitor_stream / event_log 行按原始时序重放回 Redis 实时流，
// 时间戳 rebase 到「当前」（t1 -> now），供 wisefido-sensor / cardagg 当作实时数据消费。
//
// 消费者丢弃 >6s 的消息（sensorMonitorMaxInboundAgeMs=6000），所以必须 rebase + 按真实
// 间隔节奏 sleep 重放，否则整批被当陈旧丢掉。消息信封复用 rediscommon.IoTStreamMessage
// .ToStreamMap()，与真实生产者（qinglan）XADD 格式完全一致。
//
// 用法:
//   cd owlBack/tools
//   go run ./redis-replay/ --device-uids 4D8710D5CABB --t1 "2026-06-05 06:06:44" \
//       --t2 "2026-06-05 06:16:42" --tz Asia/Shanghai [--speed 1] [--streams monitor,event] [--dry-run]
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
	topicType string // "monitor" / "event"
	category  string // "track" / "heart" / "EnterRoom" / "Fall" / ...
	payload   []interface{}
}

func main() {
	var (
		uidsArg   = flag.String("device-uids", "", "逗号分隔 device_uid 列表（必填）")
		t1Arg     = flag.String("t1", "", "起始本地时间 \"YYYY-MM-DD HH:MM:SS\"（必填，按 --tz 解释）")
		t2Arg     = flag.String("t2", "", "结束本地时间（必填）")
		tzArg     = flag.String("tz", "", "IANA 时区，如 Asia/Shanghai（必填）")
		speed     = flag.Float64("speed", 1.0, "重放倍速（>1 加快，<1 放慢）")
		streamSel = flag.String("streams", "monitor,event", "重放哪些流：monitor,event 逗号组合")
		dryRun    = flag.Bool("dry-run", false, "只打印不实际 XADD")
	)
	flag.Parse()

	if *uidsArg == "" || *t1Arg == "" || *t2Arg == "" || *tzArg == "" {
		flag.Usage()
		os.Exit(1)
	}

	loc, err := time.LoadLocation(*tzArg)
	if err != nil {
		fatal("加载时区 %q 失败: %v", *tzArg, err)
	}
	t1, err := time.ParseInLocation(localLayout, *t1Arg, loc)
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

	wantMonitor := strings.Contains(*streamSel, "monitor")
	wantEvent := strings.Contains(*streamSel, "event")
	if !wantMonitor && !wantEvent {
		fatal("--streams 至少含 monitor 或 event")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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

	uids := splitCSV(*uidsArg)
	addrs, dtypeByAddr := resolveDevices(ctx, db, uids)
	if len(addrs) == 0 {
		fatal("没有解析到任何 device_addr（device_uids=%v）", uids)
	}

	rows, err := loadRecords(ctx, db, addrs, dtypeByAddr, t1.UTC(), t2.UTC(), wantMonitor, wantEvent)
	if err != nil {
		fatal("查询时序失败: %v", err)
	}
	if len(rows) == 0 {
		fatal("窗口内无数据：%s ~ %s (%s)", *t1Arg, *t2Arg, *tzArg)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].tsMs < rows[j].tsMs })

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

	fmt.Printf("重放 %d 条 (monitor+event) | 设备 %d 台 | 窗口 %s~%s %s | 倍速 %.2gx | rebase t1->now%s\n",
		len(rows), len(addrs), *t1Arg, *t2Arg, *tzArg, *speed, dryRunTag(*dryRun))

	t1ms := t1.UnixMilli()
	startWall := time.Now()
	seq := map[string]uint64{}
	var nMon, nEvt int

	for _, r := range rows {
		rel := time.Duration(float64(r.tsMs-t1ms)/(*speed)) * time.Millisecond
		if d := time.Until(startWall.Add(rel)); d > 0 {
			select {
			case <-ctx.Done():
				fmt.Printf("\n中断：已重放 monitor=%d event=%d\n", nMon, nEvt)
				return
			case <-time.After(d):
			}
		}

		addr, _ := netip.ParseAddr(r.addr)
		seq[r.addr]++
		msg := rediscommon.IoTStreamMessage{
			Producer:       r.addr, // device-direct producer = device_addr（非 sensor.* 不触发回环过滤）
			SequenceNumber: seq[r.addr],
			DeviceAddr:     addr,
			DeviceType:     r.dtype,
			Timestamp:      time.Now().UnixMilli(), // 发布即时戳，保证消费者 age≈0
			TopicType:      r.topicType,
			Category:       r.category,
			DataValue:      r.payload,
		}
		def := rediscommon.StreamMonitor
		if r.topicType == "event" {
			def = rediscommon.StreamEvent
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
		if r.topicType == "event" {
			nEvt++
		} else {
			nMon++
		}
	}
	fmt.Printf("完成：monitor=%d event=%d\n", nMon, nEvt)
}

// resolveDevices: device_uid -> device_addr，并取 device_type（devices 无 type 列，从 monitor_stream 取）。
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
			dtype[a] = "Radar" // 无 monitor 历史的兜底（event-only 设备）
		}
	}
	return addrs, dtype
}

func loadRecords(ctx context.Context, db *sql.DB, addrs []string, dtype map[string]string, t1, t2 time.Time, wantMon, wantEvt bool) ([]record, error) {
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
				cat = st[i+1:] // "radar.track" -> "track"
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
	return out, nil
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
