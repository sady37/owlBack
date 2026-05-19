// track-status-tail — sensor_v2 PR-3c dev playback 工具
//
// 用途：tail sensor:track:status:stream（Layer 1 → Layer 2 TrackStatus 投影流），
// 实时打印每帧每 track 的 verdict / person / 位置 / cell 区域类型，方便 PR-3/4/5
// dev 阶段验证 SuiteCensus + TrackStatus 派生正确性。
//
// 使用：
//   REDIS_ADDR=localhost:6379 ./track-status-tail
//   ./track-status-tail -room fd00:0:3:111:80::/128  # 仅看特定 room
//   ./track-status-tail -compact                       # 单行紧凑模式
//
// 设计选择：
//   - 用 XRead $ (latest only) 而非 consumer group：dev tail 不需要 ack/重放语义，
//     多实例并行 tail 不互斥（与 prod consumer 0 冲突）。
//   - 流自身 MaxLen=5000 + 30s TTL → 落后 N s 即丢，符合"瞬态投影"语义。

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	rediscommon "owl-common/redis"

	"github.com/go-redis/redis/v8"
)

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	addr := flag.String("addr", getEnv("REDIS_ADDR", "localhost:6379"), "redis addr")
	password := flag.String("password", getEnv("REDIS_PASSWORD", ""), "redis password")
	roomFilter := flag.String("room", "", "only print this room_id (full /128 INET)")
	compact := flag.Bool("compact", false, "single-line per record")
	flag.Parse()

	client := redis.NewClient(&redis.Options{Addr: *addr, Password: *password, DB: 0})
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		cancel()
	}()

	if err := client.Ping(ctx).Err(); err != nil {
		fmt.Fprintf(os.Stderr, "redis ping: %v\n", err)
		os.Exit(1)
	}

	streamName := rediscommon.StreamSensorTrackStatus.Name
	// 硬编码 "$" = 只读后续新消息，绝不回放历史。
	// 故意不提供 -from 0 / -from-beginning 之类的开关：dev tail 工具如果回放历史，
	// 多实例并行 + lastID 推进，仍然不会和 prod consumer group 抢消息（XRead 无 group 概念），
	// 但会让本工具读到 30s 内堆积的旧帧 → 误导调试。明确不开。
	lastID := "$"
	fmt.Fprintf(os.Stderr, "tailing %s (latest-only; ctrl-c to stop)\n", streamName)

	for {
		if ctx.Err() != nil {
			return
		}
		res, err := client.XRead(ctx, &redis.XReadArgs{
			Streams: []string{streamName, lastID},
			Block:   5 * time.Second,
			Count:   100,
		}).Result()
		if err == redis.Nil || err == context.Canceled {
			continue
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "xread: %v\n", err)
			time.Sleep(time.Second)
			continue
		}
		for _, s := range res {
			for _, msg := range s.Messages {
				lastID = msg.ID
				if *roomFilter != "" {
					if v, _ := msg.Values["room_id"].(string); v != *roomFilter {
						continue
					}
				}
				print(msg.Values, *compact)
			}
		}
	}
}

// print 渲染一条 TrackStatus（Values map）。compact=true 单行；否则多行 key=value。
// 字段顺序与 sensor_v2 §10.1.1 一致，方便和文档对照阅读。
func print(v map[string]interface{}, compact bool) {
	keys := []string{
		"updated_at_ms", "room_id", "device_id", "track_id",
		"verdict", "ghost_penalty",
		"x", "y", "z", "pose", "still_sec", "cell_area_type", "enter_target",
		"in_bed_zone_id", "in_room_zone_id", "in_bathroom_zone_id",
		"person_id", "person_role",
	}
	if compact {
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			if val, ok := v[k]; ok {
				parts = append(parts, fmt.Sprintf("%s=%v", k, val))
			}
		}
		fmt.Println(strings.Join(parts, " "))
		return
	}
	fmt.Println("---")
	for _, k := range keys {
		if val, ok := v[k]; ok {
			fmt.Printf("  %-22s %v\n", k, val)
		}
	}
}
