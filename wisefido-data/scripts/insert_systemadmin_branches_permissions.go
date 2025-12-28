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

	// 插入 SystemAdmin 的 branches 权限
	permissions := []struct {
		resourceType   string
		permissionType string
	}{
		{"branches", "R"},
		{"branches", "C"},
		{"branches", "U"},
		{"branches", "D"},
	}

	fmt.Fprintf(os.Stdout, "Inserting SystemAdmin branches permissions...\n")

	for _, perm := range permissions {
		_, err := db.Exec(`
			INSERT INTO role_permissions (tenant_id, role_code, resource_type, permission_type, assigned_only, branch_only)
			VALUES ($1, 'SystemAdmin', $2, $3, FALSE, FALSE)
			ON CONFLICT ((COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid)), role_code, resource_type, permission_type)
			DO UPDATE SET assigned_only = EXCLUDED.assigned_only, branch_only = EXCLUDED.branch_only
		`, systemTenantID, perm.resourceType, perm.permissionType)
		if err != nil {
			log.Fatalf("Failed to insert permission %s/%s: %v", perm.resourceType, perm.permissionType, err)
		}
		fmt.Fprintf(os.Stdout, "  ✅ Inserted: branches/%s\n", perm.permissionType)
	}

	fmt.Fprintf(os.Stdout, "\n✅ Successfully inserted SystemAdmin branches permissions\n")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
