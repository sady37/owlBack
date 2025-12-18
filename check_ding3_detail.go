package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
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

	// Find contact with ding3@gmail.com
	rows, err := db.Query(`
		SELECT 
			r.nickname,
			rc.slot,
			rc.contact_id::text,
			rc.contact_email,
			rc.email_hash IS NOT NULL as has_email_hash,
			encode(rc.email_hash, 'hex') as email_hash_hex,
			rc.password_hash IS NOT NULL as has_password_hash,
			rc.is_enabled
		FROM residents r
		JOIN resident_contacts rc ON r.resident_id = rc.resident_id
		WHERE rc.contact_email LIKE '%ding3%' OR r.nickname = 'Done'
		ORDER BY r.nickname, rc.slot
	`)
	if err != nil {
		log.Fatalf("Failed to query contacts: %v", err)
	}
	defer rows.Close()

	fmt.Println("=== Checking contacts for Done (including ding3) ===\n")

	for rows.Next() {
		var nickname, slot, contactID, contactEmail sql.NullString
		var hasEmailHash, hasPasswordHash, isEnabled bool
		var emailHashHex sql.NullString

		err = rows.Scan(
			&nickname, &slot, &contactID, &contactEmail,
			&hasEmailHash, &emailHashHex, &hasPasswordHash, &isEnabled,
		)
		if err != nil {
			log.Fatalf("Failed to scan contact: %v", err)
		}

		fmt.Printf("Resident: %s, Slot: %s\n", nickname.String, slot.String)
		fmt.Printf("  contact_id: %s\n", contactID.String)
		fmt.Printf("  contact_email: %s\n", contactEmail.String)
		fmt.Printf("  is_enabled: %v\n", isEnabled)

		if hasEmailHash {
			fmt.Printf("  email_hash: %s\n", emailHashHex.String)
			// Calculate expected hash for ding3@gmail.com
			if contactEmail.String == "ding3@gmail.com" {
				expectedEmail := "ding3@gmail.com"
				expectedHash := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(expectedEmail))))
				expectedHashHex := hex.EncodeToString(expectedHash[:])
				fmt.Printf("  Expected hash for 'ding3@gmail.com': %s\n", expectedHashHex)
				if emailHashHex.String == expectedHashHex {
					fmt.Printf("  ✅ Email hash matches\n")
				} else {
					fmt.Printf("  ❌ Email hash does NOT match!\n")
				}
			}
		} else {
			fmt.Printf("  ❌ email_hash is NULL\n")
		}

		if hasPasswordHash {
			fmt.Printf("  ✅ password_hash exists\n")
		} else {
			fmt.Printf("  ❌ password_hash is NULL\n")
		}

		if contactEmail.String == "ding3@gmail.com" {
			if !hasEmailHash {
				fmt.Printf("\n  ❌ PROBLEM: ding3@gmail.com has no email_hash - CANNOT login\n")
			} else if !hasPasswordHash {
				fmt.Printf("\n  ❌ PROBLEM: ding3@gmail.com has no password_hash - CANNOT login\n")
			} else if !isEnabled {
				fmt.Printf("\n  ❌ PROBLEM: ding3@gmail.com is_enabled=false - CANNOT login\n")
			} else {
				fmt.Printf("\n  ✅ ding3@gmail.com should be able to login\n")
			}
		}

		fmt.Println()
	}

	// Also check done@gmail.com and smith@gmail.com in residents table
	fmt.Println("=== Checking done@gmail.com and smith@gmail.com in residents ===\n")
	emails := []string{"done@gmail.com", "smith@gmail.com"}
	for _, email := range emails {
		emailLower := strings.ToLower(strings.TrimSpace(email))
		emailHash := sha256.Sum256([]byte(emailLower))

		var residentID, nickname sql.NullString
		var hasEmailHash, hasPasswordHash bool
		var status sql.NullString
		err := db.QueryRow(`
			SELECT 
				r.resident_id::text,
				r.nickname,
				r.email_hash IS NOT NULL as has_email_hash,
				r.password_hash IS NOT NULL as has_password_hash,
				r.status
			FROM residents r
			WHERE r.email_hash = $1
		`, emailHash[:]).Scan(&residentID, &nickname, &hasEmailHash, &hasPasswordHash, &status)
		if err != nil {
			if err == sql.ErrNoRows {
				fmt.Printf("❌ %s: NOT found in residents table\n", email)
			} else {
				fmt.Printf("❌ %s: Error querying: %v\n", email, err)
			}
		} else {
			fmt.Printf("✅ %s:\n", email)
			fmt.Printf("  resident_id: %s\n", residentID.String)
			fmt.Printf("  nickname: %s\n", nickname.String)
			fmt.Printf("  status: %s\n", status.String)
			fmt.Printf("  has_email_hash: %v\n", hasEmailHash)
			fmt.Printf("  has_password_hash: %v\n", hasPasswordHash)
			if !hasEmailHash {
				fmt.Printf("  ❌ PROBLEM: No email_hash - CANNOT login\n")
			} else if !hasPasswordHash {
				fmt.Printf("  ❌ PROBLEM: No password_hash - CANNOT login\n")
			} else if status.String != "active" {
				fmt.Printf("  ❌ PROBLEM: status='%s' (not 'active') - CANNOT login\n", status.String)
			} else {
				fmt.Printf("  ✅ Should be able to login\n")
			}
		}
		fmt.Println()
	}
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

