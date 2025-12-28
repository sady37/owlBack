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

	fmt.Println("🚀 Adding users permissions to SystemAdmin...")
	fmt.Println("   (Needed to create/manage SystemOperator and other system users)")

	// Add users permissions for SystemAdmin
	sql := `
		INSERT INTO role_permissions (tenant_id, role_code, resource_type, permission_type, assigned_only) VALUES
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'users', 'R', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'users', 'C', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'users', 'U', FALSE),
		('00000000-0000-0000-0000-000000000001', 'SystemAdmin', 'users', 'D', FALSE)
		ON CONFLICT ((COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid)), role_code, resource_type, permission_type) 
		DO UPDATE SET assigned_only = EXCLUDED.assigned_only
	`

	result, err := db.Exec(sql)
	if err != nil {
		log.Fatalf("Failed to insert: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Fatalf("Failed to get rows affected: %v", err)
	}

	fmt.Printf("✅ Added/Updated %d SystemAdmin users permissions\n", rowsAffected)

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

	fmt.Println("\n📊 SystemAdmin Permissions (after update):")
	fmt.Println("Resource Type      | Permissions")
	fmt.Println("-----------------------------------")

	permissions := make(map[string][]string)
	for rows.Next() {
		var resourceType, permissionType string
		if err := rows.Scan(&resourceType, &permissionType); err != nil {
			log.Fatalf("Failed to scan: %v", err)
		}
		if permissions[resourceType] == nil {
			permissions[resourceType] = []string{}
		}
		permissions[resourceType] = append(permissions[resourceType], permissionType)
	}

	for resourceType, permTypes := range permissions {
		hasR := contains(permTypes, "R")
		hasC := contains(permTypes, "C")
		hasU := contains(permTypes, "U")
		hasD := contains(permTypes, "D")

		perms := ""
		if hasR {
			perms += "R"
		} else {
			perms += "-"
		}
		if hasC {
			perms += "C"
		} else {
			perms += "-"
		}
		if hasU {
			perms += "U"
		} else {
			perms += "-"
		}
		if hasD {
			perms += "D"
		} else {
			perms += "-"
		}

		fmt.Printf("%-19s | %s\n", resourceType, perms)
	}

	fmt.Printf("\n✅ Total: %d resources, %d permissions\n", len(permissions), countTotal(permissions))
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func countTotal(permissions map[string][]string) int {
	total := 0
	for _, perms := range permissions {
		total += len(perms)
	}
	return total
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
