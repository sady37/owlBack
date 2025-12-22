// +build ignore

package main

import (
	"context"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"owl-common/config"
	"owl-common/database"
)

func main() {
	// 加载配置
	cfg := &config.DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		Database: getEnv("DB_NAME", "owlrd"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	// 连接数据库
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// 检查表是否存在
	var exists bool
	err = db.QueryRowContext(context.Background(),
		`SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'sleepace_report'
		)`,
	).Scan(&exists)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to check table: %v\n", err)
		os.Exit(1)
	}

	if exists {
		fmt.Println("✅ sleepace_report table exists!")
	} else {
		fmt.Println("❌ sleepace_report table does not exist!")
		os.Exit(1)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}

