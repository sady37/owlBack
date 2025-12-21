package main

import (
	"fmt"
	"os"
	"path/filepath"

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

	// 读取 SQL 文件
	sqlFile := filepath.Join("..", "..", "owlRD", "db", "26_sleepace_report.sql")
	sqlBytes, err := os.ReadFile(sqlFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read SQL file: %v\n", err)
		os.Exit(1)
	}

	// 执行 SQL
	_, err = db.Exec(string(sqlBytes))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to execute SQL: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ sleepace_report table created successfully!")
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

