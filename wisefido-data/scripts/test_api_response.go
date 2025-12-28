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

	// Simulate the API query
	q := `SELECT permission_id::text, COALESCE(tenant_id::text, NULL), role_code, resource_type, permission_type, assigned_only, branch_only
	      FROM role_permissions
	      WHERE tenant_id = $1
	      ORDER BY role_code, resource_type, permission_type
	      LIMIT 10`

	rows, err := db.Query(q, systemTenantID)
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}
	defer rows.Close()

	fmt.Println("📊 Testing API query (first 10 records):")
	fmt.Println("permission_id | tenant_id | role_code | resource_type | permission_type | assigned_only | branch_only")
	fmt.Println("---------------------------------------------------------------------------------------------------")

	count := 0
	for rows.Next() {
		var pid, rc, rt, pt string
		var tenantIDStr sql.NullString
		var assignedOnly, branchOnly bool
		if err := rows.Scan(&pid, &tenantIDStr, &rc, &rt, &pt, &assignedOnly, &branchOnly); err != nil {
			log.Fatalf("Failed to scan: %v", err)
		}
		tenantID := "NULL"
		if tenantIDStr.Valid {
			tenantID = tenantIDStr.String
		}
		fmt.Printf("%-13s | %-9s | %-9s | %-13s | %-15s | %-13v | %v\n",
			pid[:8]+"...", tenantID[:8]+"...", rc, rt, pt, assignedOnly, branchOnly)
		count++
	}

	fmt.Printf("\n✅ Found %d records\n", count)

	// Check total count
	var totalCount int
	err = db.QueryRow(`SELECT COUNT(*) FROM role_permissions WHERE tenant_id = $1`, systemTenantID).Scan(&totalCount)
	if err != nil {
		log.Fatalf("Failed to count: %v", err)
	}
	fmt.Printf("📊 Total permissions in System tenant: %d\n", totalCount)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
