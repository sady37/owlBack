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

	// 检查 SystemAdmin 权限记录数量
	var count int
	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM role_permissions
		WHERE tenant_id = $1
		  AND role_code = 'SystemAdmin'
		  AND resource_type != 'service_levels'
	`, systemTenantID).Scan(&count)
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}

	fmt.Fprintf(os.Stdout, "SystemAdmin permissions count: %d\n", count)

	if count == 0 {
		fmt.Fprintf(os.Stdout, "❌ SystemAdmin has no permissions in database!\n")
		fmt.Fprintf(os.Stdout, "Need to insert permissions from 03_role_permissions.sql\n")
	} else {
		fmt.Fprintf(os.Stdout, "✅ SystemAdmin has %d permissions\n", count)
		
		// 列出所有权限
		rows, err := db.Query(`
			SELECT resource_type, permission_type, assigned_only, branch_only
			FROM role_permissions
			WHERE tenant_id = $1
			  AND role_code = 'SystemAdmin'
			  AND resource_type != 'service_levels'
			ORDER BY resource_type, permission_type
		`, systemTenantID)
		if err != nil {
			log.Fatalf("Failed to query: %v", err)
		}
		defer rows.Close()

		fmt.Fprintf(os.Stdout, "\nSystemAdmin permissions:\n")
		for rows.Next() {
			var resourceType, permissionType string
			var assignedOnly, branchOnly bool
			if err := rows.Scan(&resourceType, &permissionType, &assignedOnly, &branchOnly); err != nil {
				log.Fatalf("Failed to scan: %v", err)
			}
			fmt.Fprintf(os.Stdout, "  - %s: %s (assigned_only=%v, branch_only=%v)\n",
				resourceType, permissionType, assignedOnly, branchOnly)
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
