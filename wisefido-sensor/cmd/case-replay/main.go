// case-replay：把 doc/cases/<name>/ 里的 fixture JSON 回放到 redis stream，
// 让 wisefido-sensor 在线服务实时消费，用于端到端验证 wire schema / B1+B2 修复 /
// ghost 算法路径等。
//
// 与 cmd/roomengine-playback 区别：
//   - playback：直接构造 TrackManager + grid 跑学习/检测，**不发 redis**，
//     用于离线分析与生成 HTML 查看器。
//   - case-replay：写 redis stream，让正在跑的 wisefido-sensor 服务接收，
//     验证 publishAIMessage 真实 wire 输出（含 source / reason / 等新字段）。
//
// 用法:
//
//	go run ./cmd/case-replay \
//	  --case      9923-kitchen-0429 \
//	  --tenant-id 43b8fbf7-b55f-4b48-bd8b-27bb14f48870 \
//	  --device-type Radar \
//	  --duration  120s
//
// fixture JSON 中每条 row 的 timestamp 按比例压缩到 --duration 内回放，
// 保持相对时序（ghost penalty 累积、probation 计数等时间相关逻辑不畸变）。
//
// 验证流程：
//  1. 启动本工具回放
//  2. 期间 tail -f /home/wisefido/owl/log/wisefido-sensor.log | grep ai_emit
//  3. 完成后 redis-cli XREVRANGE iot:event:stream + - COUNT 5 看 wire 字段
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/go-redis/redis/v8"
)

// fixtureRow 与 doc/cases/*/2026-*.json 数组项一致（iot_timeseries 行）。
type fixtureRow struct {
	ID         int64                    `json:"id"`
	DeviceUID  string                   `json:"device_uid"`
	DeviceID   string                   `json:"device_id"`
	Timestamp  int64                    `json:"timestamp"`
	TopicType  string                   `json:"topic_type"`
	Category   string                   `json:"category"`
	DataValue  []map[string]interface{} `json:"data_value"`
}

func main() {
	var (
		caseName   = flag.String("case", "", "case 目录名（doc/cases/<name>/）；自动找其下最大 .json 数据文件")
		caseDir    = flag.String("case-dir", "doc/cases", "case 根目录（默认 doc/cases）")
		tenantID   = flag.String("tenant-id", "", "tenant_id（必填，wisefido-sensor 路由查 tenant 用）")
		cardID     = flag.String("card-id", "", "card_id（可选，留空让 cardagg/engine 走 device→room 路由）")
		deviceType = flag.String("device-type", "Radar", "device_type（Radar / Sleepad）")
		duration   = flag.Duration("duration", 120*time.Second, "压缩回放总时长；保留相对时序")
		redisAddr  = flag.String("redis", "127.0.0.1:6379", "redis 地址")
		redisPass  = flag.String("redis-pass", "TeLunSu-36kr", "redis 密码")
	)
	flag.Parse()

	if *caseName == "" || *tenantID == "" {
		log.Fatalf("usage: --case <name> --tenant-id <uuid> [--duration 120s]")
	}

	// 1. 找 case 数据文件
	caseFolder := filepath.Join(*caseDir, *caseName)
	files, err := os.ReadDir(caseFolder)
	if err != nil {
		log.Fatalf("read case dir %s: %v", caseFolder, err)
	}
	var dataFile string
	var dataSize int64
	for _, f := range files {
		if f.IsDir() || filepath.Ext(f.Name()) != ".json" || f.Name() == "room_layout.json" {
			continue
		}
		info, _ := f.Info()
		if info.Size() > dataSize {
			dataSize = info.Size()
			dataFile = filepath.Join(caseFolder, f.Name())
		}
	}
	if dataFile == "" {
		log.Fatalf("no fixture JSON found under %s", caseFolder)
	}
	log.Printf("loading fixture: %s (%d bytes)", dataFile, dataSize)

	// 2. 读 + 解析
	raw, err := os.ReadFile(dataFile)
	if err != nil {
		log.Fatalf("read %s: %v", dataFile, err)
	}
	var rows []fixtureRow
	if err := json.Unmarshal(raw, &rows); err != nil {
		log.Fatalf("unmarshal: %v", err)
	}
	if len(rows) == 0 {
		log.Fatalf("fixture empty")
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Timestamp < rows[j].Timestamp })
	firstTs := rows[0].Timestamp
	lastTs := rows[len(rows)-1].Timestamp
	realSpanMs := lastTs - firstTs
	if realSpanMs <= 0 {
		realSpanMs = 1
	}
	log.Printf("rows=%d  real_span=%s  compress_to=%s  ratio=%.2fx",
		len(rows),
		time.Duration(realSpanMs)*time.Millisecond,
		*duration,
		float64(realSpanMs*int64(time.Millisecond))/float64(*duration))

	// 3. 连 redis
	rdb := redis.NewClient(&redis.Options{Addr: *redisAddr, Password: *redisPass, DB: 0})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis ping: %v", err)
	}
	defer rdb.Close()

	// 4. 回放
	startWall := time.Now()
	startNowMs := startWall.UnixMilli()
	stats := map[string]int{}
	for i, row := range rows {
		// 压缩后的 wall-clock 偏移
		offsetMs := (row.Timestamp - firstTs) * int64(*duration) / int64(time.Millisecond) / realSpanMs
		targetWall := startWall.Add(time.Duration(offsetMs) * time.Millisecond)
		if delay := time.Until(targetWall); delay > 0 {
			time.Sleep(delay)
		}

		// 时间戳也压缩到当前时间（保留相对差距）
		shiftedTs := startNowMs + offsetMs

		// 路由 stream
		stream := streamForTopic(row.TopicType)
		if stream == "" {
			stats["skipped"]++
			continue
		}
		// data_value 转 JSON 字符串（与 firmware producer 一致）
		dvJSON, err := json.Marshal(row.DataValue)
		if err != nil {
			log.Printf("marshal dataValue row=%d: %v", i, err)
			stats["err"]++
			continue
		}

		values := map[string]interface{}{
			"device_uid":  row.DeviceUID,
			"device_id":   row.DeviceID,
			"device_type": *deviceType,
			"tenant_id":   *tenantID,
			"timestamp":   fmt.Sprintf("%d", shiftedTs),
			"topic_type":  row.TopicType,
			"category":    row.Category,
			"dataValue":   string(dvJSON),
		}
		if *cardID != "" {
			values["card_id"] = *cardID
		}

		_, err = rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: stream,
			Values: values,
		}).Result()
		if err != nil {
			log.Printf("xadd row=%d stream=%s: %v", i, stream, err)
			stats["err"]++
			continue
		}
		stats[stream]++

		// 每秒进度
		if i > 0 && i%200 == 0 {
			elapsed := time.Since(startWall)
			log.Printf("progress: %d/%d (%.0f%%)  elapsed=%s  by_stream=%v",
				i, len(rows), float64(i)*100/float64(len(rows)), elapsed.Truncate(100*time.Millisecond), stats)
		}
	}

	elapsed := time.Since(startWall)
	log.Printf("done: %d/%d rows replayed in %s  by_stream=%v",
		len(rows), len(rows), elapsed.Truncate(100*time.Millisecond), stats)
	log.Printf("next: tail -f /home/wisefido/owl/log/wisefido-sensor.log | grep ai_emit")
	log.Printf("      redis-cli XREVRANGE iot:event:stream + - COUNT 5  # 看 track_verdict wire 字段")
}

// streamForTopic 把 fixture 的 topic_type 字段映射到 redis stream 名。
func streamForTopic(topic string) string {
	switch topic {
	case "monitor":
		return "iot:monitor:stream"
	case "event":
		return "iot:event:stream"
	case "alarm":
		return "iot:alarm:stream"
	default:
		return ""
	}
}
