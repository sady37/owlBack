//go:build ignore
// +build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "owlrd")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	fmt.Println("📊 验证迁移结果:")
	fmt.Println("")

	// 1. 检查表结构
	fmt.Println("1. 表结构检查:")
	rows, err := db.Query(`
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_name = 'role_permissions'
		ORDER BY ordinal_position
	`)
	if err != nil {
		log.Fatalf("Failed to query columns: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var colName, dataType, nullable string
		if err := rows.Scan(&colName, &dataType, &nullable); err != nil {
			continue
		}
		fmt.Printf("   - %s: %s (nullable: %s)\n", colName, dataType, nullable)
	}
	fmt.Println("")

	// 2. 检查约束
	fmt.Println("2. 约束检查:")
	var constraintExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.table_constraints
			WHERE table_name = 'role_permissions'
			AND constraint_name = 'chk_permission_scope'
		)
	`).Scan(&constraintExists)
	if err != nil {
		log.Fatalf("Failed to check constraint: %v", err)
	}
	if constraintExists {
		fmt.Println("   ✅ chk_permission_scope 约束存在")
	} else {
		fmt.Println("   ❌ chk_permission_scope 约束不存在")
	}
	fmt.Println("")

	// 3. 数据分布
	fmt.Println("3. 数据分布:")
	rows2, err := db.Query(`
		SELECT permission_scope, COUNT(*) as count,
		       ROUND(100.0 * COUNT(*) / SUM(COUNT(*)) OVER (), 2) as percentage
		FROM role_permissions
		GROUP BY permission_scope
		ORDER BY permission_scope
	`)
	if err != nil {
		log.Fatalf("Failed to query distribution: %v", err)
	}
	defer rows2.Close()
	for rows2.Next() {
		var scope string
		var count int
		var pct float64
		if err := rows2.Scan(&scope, &count, &pct); err != nil {
			continue
		}
		scopeName := map[string]string{"A": "All", "S": "assigned_only", "B": "branch_only"}[scope]
		fmt.Printf("   - %s (%s): %d 条 (%.1f%%)\n", scope, scopeName, count, pct)
	}
	fmt.Println("")

	// 4. 验证数据完整性
	fmt.Println("4. 数据完整性验证:")
	var nullCount, invalidCount, totalCount int
	err = db.QueryRow(`
		SELECT 
			COUNT(*) FILTER (WHERE permission_scope IS NULL) as null_count,
			COUNT(*) FILTER (WHERE permission_scope NOT IN ('A', 'S', 'B')) as invalid_count,
			COUNT(*) as total_count
		FROM role_permissions
	`).Scan(&nullCount, &invalidCount, &totalCount)
	if err != nil {
		log.Fatalf("Failed to validate: %v", err)
	}

	if nullCount == 0 && invalidCount == 0 {
		fmt.Printf("   ✅ 所有 %d 条记录数据完整\n", totalCount)
	} else {
		if nullCount > 0 {
			fmt.Printf("   ❌ 发现 %d 条记录的 permission_scope 为 NULL\n", nullCount)
		}
		if invalidCount > 0 {
			fmt.Printf("   ❌ 发现 %d 条记录的 permission_scope 值无效\n", invalidCount)
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

