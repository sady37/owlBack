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

	// Step 1: Add branch_only column
	fmt.Println("Step 1: Adding branch_only column...")
	_, err = db.Exec(`
		ALTER TABLE role_permissions 
		ADD COLUMN IF NOT EXISTS branch_only BOOLEAN DEFAULT FALSE
	`)
	if err != nil {
		log.Fatalf("Failed to add column: %v", err)
	}
	fmt.Println("✅ Column added")

	// Step 2: Add comment
	fmt.Println("\nStep 2: Adding comment...")
	_, err = db.Exec(`
		COMMENT ON COLUMN role_permissions.branch_only IS 
		'权限范围：FALSE=所有资源，TRUE=仅限同一 Branch 的资源（匹配 users.branch_tag = units.branch_tag）。当 users.branch_tag IS NULL 时，可以管理 units.branch_tag IS NULL 或 units.branch_tag = ''-'' 的资源（空值匹配空值）'
	`)
	if err != nil {
		log.Printf("Warning: Failed to add comment: %v", err)
	} else {
		fmt.Println("✅ Comment added")
	}

	// Step 3: Update Manager permissions
	fmt.Println("\nStep 3: Updating Manager permissions...")
	result, err := db.Exec(`
		UPDATE role_permissions 
		SET branch_only = TRUE 
		WHERE role_code = 'Manager' 
		  AND resource_type IN ('residents', 'users', 'units', 'rooms', 'beds', 'resident_contacts', 'resident_phi')
		  AND permission_type IN ('R', 'C', 'U', 'D')
		  AND tenant_id = '00000000-0000-0000-0000-000000000001'
	`)
	if err != nil {
		log.Fatalf("Failed to update permissions: %v", err)
	}
	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("✅ Updated %d Manager permission records\n", rowsAffected)

	fmt.Println("\n✅ Migration completed successfully!")
}
