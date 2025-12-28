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

	// Check both SystemAdmin and SystemOperator
	roles := []string{"SystemAdmin", "SystemOperator"}

	for _, roleCode := range roles {
		fmt.Printf("\n%s\n", strings.Repeat("=", 80))
		fmt.Printf("📊 %s Permissions (Current in Database):\n", roleCode)
		fmt.Println(strings.Repeat("=", 80))

		rows, err := db.Query(`
			SELECT resource_type, permission_type, assigned_only, branch_only
			FROM role_permissions
			WHERE tenant_id = $1 AND role_code = $2
			ORDER BY resource_type, permission_type
		`, systemTenantID, roleCode)

		if err != nil {
			log.Fatalf("Failed to query: %v", err)
		}

		permissions := make(map[string][]string) // resource -> []permission_types

		for rows.Next() {
			var resourceType, permissionType string
			var assignedOnly, branchOnly bool
			if err := rows.Scan(&resourceType, &permissionType, &assignedOnly, &branchOnly); err != nil {
				log.Fatalf("Failed to scan: %v", err)
			}

			if permissions[resourceType] == nil {
				permissions[resourceType] = []string{}
			}
			permissions[resourceType] = append(permissions[resourceType], permissionType)
		}
		rows.Close()

		if len(permissions) == 0 {
			fmt.Printf("❌ No permissions found for %s\n", roleCode)
		} else {
			fmt.Println("Resource Type      | Permissions | assigned_only | branch_only")
			fmt.Println("-------------------------------------------------------------------")

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

				// Get assigned_only and branch_only from first permission (they should be same for all)
				var assignedOnly, branchOnly bool
				err := db.QueryRow(`
					SELECT assigned_only, branch_only
					FROM role_permissions
					WHERE tenant_id = $1 AND role_code = $2 AND resource_type = $3
					LIMIT 1
				`, systemTenantID, roleCode, resourceType).Scan(&assignedOnly, &branchOnly)

				if err == nil {
					fmt.Printf("%-19s | %-11s | %-13v | %v\n",
						resourceType, perms, assignedOnly, branchOnly)
				} else {
					fmt.Printf("%-19s | %-11s | %-13s | %s\n",
						resourceType, perms, "?", "?")
				}
			}

			fmt.Printf("\n✅ Total: %d resources, %d permissions\n", len(permissions), countTotal(permissions))
		}
	}
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
