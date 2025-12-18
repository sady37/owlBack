package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
	"wisefido-data/internal/config"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <migration_file.sql>", os.Args[0])
	}

	migrationFile := os.Args[1]
	sqlContent, err := ioutil.ReadFile(migrationFile)
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	cfg := config.Load()

	dbNames := []string{"wisefido_data", "owlrd"}
	var db *sql.DB
	var connectedDB string

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

	// Split SQL by semicolon and execute each statement
	statements := strings.Split(string(sqlContent), ";")
	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		
		// Skip verification queries (commented out)
		if strings.Contains(stmt, "Verification query") {
			continue
		}

		fmt.Printf("Executing statement %d/%d...\n", i+1, len(statements))
		_, err := db.Exec(stmt)
		if err != nil {
			log.Fatalf("Failed to execute statement %d: %v\nStatement: %s", i+1, err, stmt[:min(100, len(stmt))])
		}
		fmt.Printf("✅ Statement %d executed successfully\n\n", i+1)
	}

	fmt.Println("✅ Migration completed successfully!")
}
