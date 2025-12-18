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

	// Check if branch_only column exists
	var columnExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = 'role_permissions' 
			  AND column_name = 'branch_only'
		)
	`).Scan(&columnExists)

	if err != nil {
		log.Fatalf("Failed to check column: %v", err)
	}

	if columnExists {
		fmt.Println("✅ branch_only column already exists")
	} else {
		fmt.Println("❌ branch_only column does NOT exist")
	}

	// Check current table structure
	fmt.Println("\n=== Current role_permissions table structure ===")
	rows, err := db.Query(`
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns
		WHERE table_name = 'role_permissions'
		ORDER BY ordinal_position
	`)
	if err != nil {
		log.Fatalf("Failed to query columns: %v", err)
	}
	defer rows.Close()

	fmt.Println("Column Name | Data Type | Nullable | Default")
	fmt.Println("------------|-----------|----------|---------")
	for rows.Next() {
		var colName, dataType, nullable, defaultValue sql.NullString
		if err := rows.Scan(&colName, &dataType, &nullable, &defaultValue); err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}
		defVal := "NULL"
		if defaultValue.Valid {
			defVal = defaultValue.String
		}
		fmt.Printf("%-11s | %-9s | %-8s | %s\n", 
			getString(colName), getString(dataType), getString(nullable), defVal)
	}

	// Check current Manager permissions
	fmt.Println("\n=== Current Manager permissions (sample) ===")
	managerRows, err := db.Query(`
		SELECT resource_type, permission_type, assigned_only
		FROM role_permissions
		WHERE role_code = 'Manager'
		  AND tenant_id = '00000000-0000-0000-0000-000000000001'
		ORDER BY resource_type, permission_type
		LIMIT 10
	`)
	if err != nil {
		log.Printf("Failed to query Manager permissions: %v", err)
	} else {
		defer managerRows.Close()
		fmt.Println("Resource Type | Permission Type | Assigned Only")
		fmt.Println("--------------|-----------------|---------------")
		for managerRows.Next() {
			var resourceType, permissionType string
			var assignedOnly bool
			if err := managerRows.Scan(&resourceType, &permissionType, &assignedOnly); err != nil {
				log.Printf("Scan error: %v", err)
				continue
			}
			fmt.Printf("%-13s | %-15s | %-13v\n", resourceType, permissionType, assignedOnly)
		}
	}
}

func getString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return "NULL"
}
