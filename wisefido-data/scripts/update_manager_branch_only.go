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
	// Database connection parameters
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "owlrd")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")

	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("✅ Connected to database")

	// Update Manager role permissions
	systemTenantID := "00000000-0000-0000-0000-000000000001"

	updateSQL := `
		UPDATE role_permissions
		SET branch_only = TRUE,
		    assigned_only = FALSE
		WHERE role_code = 'Manager'
		  AND tenant_id = $1
	`

	result, err := db.Exec(updateSQL, systemTenantID)
	if err != nil {
		log.Fatalf("Failed to update Manager role permissions: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Fatalf("Failed to get rows affected: %v", err)
	}

	fmt.Printf("✅ Updated %d Manager role permissions to use branch_only = TRUE\n", rowsAffected)

	// Verify the update
	verifySQL := `
		SELECT 
		    resource_type,
		    permission_type,
		    assigned_only,
		    branch_only,
		    COUNT(*) as count
		FROM role_permissions
		WHERE role_code = 'Manager'
		  AND tenant_id = $1
		GROUP BY resource_type, permission_type, assigned_only, branch_only
		ORDER BY resource_type, permission_type
	`

	rows, err := db.Query(verifySQL, systemTenantID)
	if err != nil {
		log.Fatalf("Failed to verify update: %v", err)
	}
	defer rows.Close()

	fmt.Println("\n📊 Verification Results:")
	fmt.Println("Resource Type | Permission Type | assigned_only | branch_only | Count")
	fmt.Println("----------------------------------------------------------------------------")

	totalCount := 0
	for rows.Next() {
		var resourceType, permissionType string
		var assignedOnly, branchOnly bool
		var count int

		if err := rows.Scan(&resourceType, &permissionType, &assignedOnly, &branchOnly, &count); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}

		fmt.Printf("%-13s | %-15s | %-13v | %-12v | %d\n",
			resourceType, permissionType, assignedOnly, branchOnly, count)
		totalCount += count
	}

	fmt.Printf("\n✅ Total: %d Manager role permissions\n", totalCount)
	fmt.Println("\n✅ Update completed successfully!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
