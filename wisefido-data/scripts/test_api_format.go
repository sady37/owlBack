//go:build ignore
// +build ignore

package main

import (
	"database/sql"
	"encoding/json"
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

	// Simulate the API response format
	q := `SELECT permission_id::text, COALESCE(tenant_id::text, NULL), role_code, resource_type, permission_type, assigned_only, branch_only
	      FROM role_permissions
	      WHERE tenant_id = $1
	      ORDER BY role_code, resource_type, permission_type
	      LIMIT 5`

	rows, err := db.Query(q, systemTenantID)
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}
	defer rows.Close()

	items := []map[string]interface{}{}
	permMap := map[string]string{"R": "read", "C": "create", "U": "update", "D": "delete"}

	for rows.Next() {
		var pid, rc, rt, pt string
		var tenantIDStr sql.NullString
		var assignedOnly, branchOnly bool
		if err := rows.Scan(&pid, &tenantIDStr, &rc, &rt, &pt, &assignedOnly, &branchOnly); err != nil {
			log.Fatalf("Failed to scan: %v", err)
		}
		perm := permMap[pt]
		scope := "all"
		if assignedOnly {
			scope = "assigned_only"
		}
		item := map[string]interface{}{
			"permission_id":   pid,
			"role_code":       rc,
			"resource_type":   rt,
			"permission_type": perm,
			"scope":           scope,
			"branch_only":     branchOnly,
			"is_active":       true,
		}
		if tenantIDStr.Valid {
			item["tenant_id"] = tenantIDStr.String
		}
		items = append(items, item)
	}

	response := map[string]interface{}{
		"code": 0,
		"data": map[string]interface{}{
			"items": items,
			"total": len(items),
		},
	}

	jsonData, _ := json.MarshalIndent(response, "", "  ")
	fmt.Println("📊 API Response Format (sample):")
	fmt.Println(string(jsonData))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
