package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"wisefido-data/internal/config"
)

func main() {
	cfg := config.Load()

	dbNames := []string{"wisefido_data", "owlrd"}
	var db *sql.DB
	var connectedDB string
	var err error

	for _, name := range dbNames {
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, name, cfg.Database.SSLMode)
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			continue
		}
		if err = db.Ping(); err != nil {
			db.Close()
			continue
		}
		connectedDB = name
		break
	}

	if db == nil || err != nil {
		log.Fatalf("Cannot connect to database: %v", err)
	}
	defer db.Close()

	fmt.Printf("Connected to database: %s\n\n", connectedDB)

	// Direct query to check branch_only column
	fmt.Println("=== Direct query: Check branch_only column ===")
	rows, err := db.Query(`
		SELECT 
			permission_id::text,
			COALESCE(tenant_id::text, 'System') as tenant_id,
			role_code,
			resource_type,
			permission_type,
			assigned_only,
			branch_only
		FROM role_permissions
		WHERE role_code = 'Manager'
		  AND tenant_id = '00000000-0000-0000-0000-000000000001'
		ORDER BY resource_type, permission_type
		LIMIT 10
	`)
	if err != nil {
		log.Fatalf("Query error: %v", err)
	}
	defer rows.Close()

	fmt.Println("Permission ID | Tenant ID | Role Code | Resource Type | Permission Type | Assigned Only | Branch Only")
	fmt.Println("--------------|-----------|-----------|---------------|----------------|---------------|-------------")
	for rows.Next() {
		var permissionID, tenantID, roleCode, resourceType, permissionType string
		var assignedOnly, branchOnly bool
		if err := rows.Scan(&permissionID, &tenantID, &roleCode, &resourceType, &permissionType, &assignedOnly, &branchOnly); err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}
		fmt.Printf("%-12s | %-9s | %-9s | %-13s | %-14s | %-13v | %-11v\n",
			permissionID[:12]+"...", tenantID[:12]+"...", roleCode, resourceType, permissionType, assignedOnly, branchOnly)
	}
}
