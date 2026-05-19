// 从 Redis card:state 导出指定 card_id 各字段。不依赖 redis-cli。
// 用法: go run export_card_status.go [card_id]
// 环境变量: REDIS_ADDR, REDIS_PASSWORD, REDIS_DB（可从 owlBack/.env  source 后执行）

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v8"
)

func main() {
	cardID := "42077c6d-ed05-46ec-a76d-b45ddb48b24f"
	if len(os.Args) > 1 {
		cardID = os.Args[1]
	}
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	parts := strings.SplitN(addr, ":", 2)
	host, port := parts[0], "6379"
	if len(parts) == 2 {
		port = parts[1]
	}
	db, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	opt := &redis.Options{
		Addr:     host + ":" + port,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
	}
	ctx := context.Background()
	client := redis.NewClient(opt)
	defer client.Close()
	if err := client.Ping(ctx).Err(); err != nil {
		fmt.Fprintf(os.Stderr, "redis ping: %v\n", err)
		os.Exit(1)
	}
	key := "card:state:" + cardID
	// Phase A：device_status 已迁出 card:state，独立到 device:status:{deviceID} Hash。
	fields := []string{"target", "room_state", "bed_state", "alarm_state", "message"}
	fmt.Printf("card_id=%s\nredis_key=%s\n---\n", cardID, key)
	for _, f := range fields {
		val, err := client.HGet(ctx, key, f).Result()
		if err == redis.Nil || val == "" {
			continue
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "[%s] %v\n", f, err)
			continue
		}
		fmt.Printf("[%s]\n", f)
		var j interface{}
		if json.Unmarshal([]byte(val), &j) == nil {
			b, _ := json.MarshalIndent(j, "", "  ")
			fmt.Println(string(b))
		} else {
			fmt.Println(val)
		}
		fmt.Println("---")
	}
}
