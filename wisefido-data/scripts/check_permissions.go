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

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping: %v", err)
	}

	systemTenantID := "00000000-0000-0000-0000-000000000001"

	// Check permissions count by role
	rows, err := db.Query(`
		SELECT role_code, COUNT(*) as count
		FROM role_permissions
		WHERE tenant_id = $1
		GROUP BY role_code
		ORDER BY role_code
	`, systemTenantID)

	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}
	defer rows.Close()

	fmt.Println("📊 Current permissions count by role:")
	fmt.Println("Role Code          | Count")
	fmt.Println("----------------------------")

	totalCount := 0
	for rows.Next() {
		var roleCode string
		var count int
		if err := rows.Scan(&roleCode, &count); err != nil {
			log.Fatalf("Failed to scan: %v", err)
		}
		fmt.Printf("%-18s | %d\n", roleCode, count)
		totalCount += count
	}

	fmt.Printf("\nTotal: %d permissions\n", totalCount)

	// Check if System tenant exists
	var tenantExists bool
	err = db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM tenants WHERE tenant_id = $1)
	`, systemTenantID).Scan(&tenantExists)

	if err != nil {
		log.Fatalf("Failed to check tenant: %v", err)
	}

	if !tenantExists {
		fmt.Println("\n❌ System tenant does not exist!")
	} else {
		fmt.Println("\n✅ System tenant exists")
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
