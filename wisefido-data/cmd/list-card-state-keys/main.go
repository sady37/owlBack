package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-redis/redis/v8"
)

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	ctx := context.Background()
	addr := getEnv("REDIS_ADDR", "localhost:6379")
	password := getEnv("REDIS_PASSWORD", "")

	client := redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: 0})
	defer client.Close()

	if err := client.Ping(ctx).Err(); err != nil {
		fmt.Fprintf(os.Stderr, "redis: %v\n", err)
		os.Exit(1)
	}

	const prefix = "card:state:"
	var cursor uint64
	var keys []string
	for {
		var batch []string
		var err error
		batch, cursor, err = client.Scan(ctx, cursor, prefix+"*", 100).Result()
		if err != nil {
			fmt.Fprintf(os.Stderr, "scan: %v\n", err)
			os.Exit(1)
		}
		keys = append(keys, batch...)
		if cursor == 0 {
			break
		}
	}

	fmt.Printf("count: %d\n", len(keys))
	for _, k := range keys {
		cardID := strings.TrimPrefix(k, prefix)
		fmt.Println(cardID)
	}
}
