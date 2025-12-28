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

	fmt.Println("🚀 Inserting SystemAdmin permissions...")

	// SystemAdmin permissions: tenants, roles, role_permissions, device_store (all RCDU)
	systemAdminSQL := `
		INSERT INTO role_permissions (tenant_id, role_code, resource_type, permission_type, assigned_only) VALUES
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'tenants', 'R', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'tenants', 'C', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'tenants', 'U', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'tenants', 'D', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'roles', 'R', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'roles', 'C', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'roles', 'U', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'roles', 'D', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'role_permissions', 'R', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'role_permissions', 'C', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'role_permissions', 'U', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'role_permissions', 'D', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'device_store', 'R', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'device_store', 'C', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'device_store', 'U', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'device_store', 'D', FALSE)
		ON CONFLICT ((COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid)), role_code, resource_type, permission_type) 
		DO UPDATE SET assigned_only = EXCLUDED.assigned_only
	`

	result, err := db.Exec(systemAdminSQL)
	if err != nil {
		log.Fatalf("Failed to insert SystemAdmin permissions: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Fatalf("Failed to get rows affected: %v", err)
	}

	fmt.Printf("✅ Inserted/Updated %d SystemAdmin permissions\n", rowsAffected)

	// Verify
	rows, err := db.Query(`
		SELECT resource_type, permission_type
		FROM role_permissions
		WHERE tenant_id = $1 AND role_code = 'SystemAdmin'
		ORDER BY resource_type, permission_type
	`, systemTenantID)
	if err != nil {
		log.Fatalf("Failed to verify: %v", err)
	}
	defer rows.Close()

	fmt.Println("\n📊 SystemAdmin Permissions (after insert):")
	fmt.Println("Resource Type      | Permission Type")
	fmt.Println("-----------------------------------")

	count := 0
	for rows.Next() {
		var resourceType, permissionType string
		if err := rows.Scan(&resourceType, &permissionType); err != nil {
			log.Fatalf("Failed to scan: %v", err)
		}
		fmt.Printf("%-19s | %s\n", resourceType, permissionType)
		count++
	}
	fmt.Printf("\n✅ Total: %d SystemAdmin permissions\n", count)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
