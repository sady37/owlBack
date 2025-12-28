//go:build ignore
// +build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

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

	// Check SystemAdmin permissions
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("📊 System Administrator (SystemAdmin) Permissions")
	fmt.Println(strings.Repeat("=", 80))

	rows, err := db.Query(`
		SELECT resource_type, permission_type, assigned_only, branch_only
		FROM role_permissions
		WHERE tenant_id = $1 AND role_code = 'SystemAdmin'
		ORDER BY resource_type, permission_type
	`, systemTenantID)

	if err != nil {
		log.Fatalf("Failed to query SystemAdmin: %v", err)
	}
	defer rows.Close()

	count := 0
	fmt.Println("Resource Type      | Permission | assigned_only | branch_only")
	fmt.Println("-------------------------------------------------------------------")
	for rows.Next() {
		var resourceType, permissionType string
		var assignedOnly, branchOnly bool
		if err := rows.Scan(&resourceType, &permissionType, &assignedOnly, &branchOnly); err != nil {
			log.Fatalf("Failed to scan: %v", err)
		}
		fmt.Printf("%-19s | %-10s | %-13v | %v\n", resourceType, permissionType, assignedOnly, branchOnly)
		count++
	}

	if count == 0 {
		fmt.Println("❌ No permissions found for SystemAdmin")
	} else {
		fmt.Printf("\n✅ Found %d permissions for SystemAdmin\n", count)
	}

	// Check Manager permissions
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("📊 Manager Permissions")
	fmt.Println(strings.Repeat("=", 80))

	rows2, err := db.Query(`
		SELECT resource_type, permission_type, assigned_only, branch_only
		FROM role_permissions
		WHERE tenant_id = $1 AND role_code = 'Manager'
		ORDER BY resource_type, permission_type
	`, systemTenantID)

	if err != nil {
		log.Fatalf("Failed to query Manager: %v", err)
	}
	defer rows2.Close()

	count2 := 0
	fmt.Println("Resource Type      | Permission | assigned_only | branch_only")
	fmt.Println("-------------------------------------------------------------------")
	for rows2.Next() {
		var resourceType, permissionType string
		var assignedOnly, branchOnly bool
		if err := rows2.Scan(&resourceType, &permissionType, &assignedOnly, &branchOnly); err != nil {
			log.Fatalf("Failed to scan: %v", err)
		}
		fmt.Printf("%-19s | %-10s | %-13v | %v\n", resourceType, permissionType, assignedOnly, branchOnly)
		count2++
	}

	if count2 == 0 {
		fmt.Println("❌ No permissions found for Manager")
	} else {
		fmt.Printf("\n✅ Found %d permissions for Manager\n", count2)
	}

	// Summary
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("📋 Summary")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("SystemAdmin: %d permissions\n", count)
	fmt.Printf("Manager: %d permissions\n", count2)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

