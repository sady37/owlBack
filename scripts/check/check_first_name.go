package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"owl-common/config"
	"owl-common/database"
)

func main() {
	// Load database config
	dbCfg := &config.DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     parseInt(getEnv("DB_PORT", "5432"), 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		Database: getEnv("DB_NAME", "owlrd"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	db, err := database.NewPostgresDB(dbCfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Check residents and their first_name in resident_phi
	rows, err := db.Query(`
		SELECT 
			r.resident_id::text,
			r.nickname,
			rp.first_name,
			rp.last_name,
			rp.phi_id::text
		FROM residents r
		LEFT JOIN resident_phi rp ON r.resident_id = rp.resident_id
		WHERE r.nickname IS NOT NULL
		ORDER BY r.nickname
		LIMIT 10
	`)
	if err != nil {
		log.Fatalf("Failed to query residents: %v", err)
	}
	defer rows.Close()

	fmt.Println("=== Checking first_name in resident_phi ===\n")
	fmt.Printf("%-40s %-20s %-20s %-20s %-40s\n", "resident_id", "nickname", "first_name", "last_name", "phi_id")
	fmt.Println(strings.Repeat("-", 140))

	foundCount := 0
	missingCount := 0
	for rows.Next() {
		var residentID, nickname sql.NullString
		var firstName, lastName, phiID sql.NullString

		err = rows.Scan(&residentID, &nickname, &firstName, &lastName, &phiID)
		if err != nil {
			log.Fatalf("Failed to scan resident: %v", err)
		}

		firstNameStr := "NULL"
		if firstName.Valid {
			firstNameStr = firstName.String
			foundCount++
		} else {
			missingCount++
		}

		lastNameStr := "NULL"
		if lastName.Valid {
			lastNameStr = lastName.String
		}

		phiIDStr := "NULL"
		if phiID.Valid {
			phiIDStr = phiID.String
		}

		fmt.Printf("%-40s %-20s %-20s %-20s %-40s\n",
			residentID.String, nickname.String, firstNameStr, lastNameStr, phiIDStr)
	}

	fmt.Println()
	fmt.Printf("Total checked: %d\n", foundCount+missingCount)
	fmt.Printf("With first_name: %d\n", foundCount)
	fmt.Printf("Without first_name: %d\n", missingCount)
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func parseInt(s string, def int) int {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	if err != nil {
		return def
	}
	return i
}
