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

	emails := []string{"done@gmail.com", "ding3@gmail.com", "smith@gmail.com"}

	fmt.Println("=== Checking email login capability ===\n")

	for _, email := range emails {
		fmt.Printf("Email: %s\n", email)
		emailLower := strings.ToLower(email)
		emailHash := sha256.Sum256([]byte(emailLower))
		emailHashHex := hex.EncodeToString(emailHash[:])
		fmt.Printf("  Expected hash: %s\n", emailHashHex)

		// Check in residents table
		fmt.Println("\n  Checking residents table:")
		rows, err := db.Query(`
			SELECT 
				r.resident_id::text,
				r.nickname,
				r.email_hash IS NOT NULL as has_email_hash,
				encode(r.email_hash, 'hex') as email_hash_hex,
				r.password_hash IS NOT NULL as has_password_hash
			FROM residents r
			WHERE r.email_hash = $1
		`, emailHash[:])
		if err != nil {
			log.Printf("Failed to query residents: %v", err)
			continue
		}

		foundInResidents := false
		for rows.Next() {
			foundInResidents = true
			var residentID, nickname sql.NullString
			var hasEmailHash, hasPasswordHash bool
			var emailHashHexDB sql.NullString
			err = rows.Scan(&residentID, &nickname, &hasEmailHash, &emailHashHexDB, &hasPasswordHash)
			if err != nil {
				log.Printf("Failed to scan resident: %v", err)
				continue
			}
			fmt.Printf("    ✅ Found in residents:\n")
			fmt.Printf("      resident_id: %s\n", residentID.String)
			fmt.Printf("      nickname: %s\n", nickname.String)
			fmt.Printf("      email_hash matches: %v\n", emailHashHexDB.String == emailHashHex)
			fmt.Printf("      has_password_hash: %v\n", hasPasswordHash)
			if !hasPasswordHash {
				fmt.Printf("      ❌ NO password_hash - CANNOT login\n")
			} else {
				fmt.Printf("      ✅ Has password_hash - can login\n")
			}
		}
		rows.Close()

		if !foundInResidents {
			fmt.Printf("    ❌ NOT found in residents table\n")
		}

		// Check in resident_contacts table
		fmt.Println("\n  Checking resident_contacts table:")
		rows2, err := db.Query(`
			SELECT 
				rc.contact_id::text,
				rc.slot,
				r.nickname as resident_nickname,
				rc.email_hash IS NOT NULL as has_email_hash,
				encode(rc.email_hash, 'hex') as email_hash_hex,
				rc.password_hash IS NOT NULL as has_password_hash,
				rc.is_enabled
			FROM resident_contacts rc
			JOIN residents r ON r.resident_id = rc.resident_id
			WHERE rc.email_hash = $1
		`, emailHash[:])
		if err != nil {
			log.Printf("Failed to query contacts: %v", err)
			continue
		}

		foundInContacts := false
		for rows2.Next() {
			foundInContacts = true
			var contactID, slot, residentNickname sql.NullString
			var hasEmailHash, hasPasswordHash, isEnabled bool
			var emailHashHexDB sql.NullString
			err = rows2.Scan(&contactID, &slot, &residentNickname, &hasEmailHash, &emailHashHexDB, &hasPasswordHash, &isEnabled)
			if err != nil {
				log.Printf("Failed to scan contact: %v", err)
				continue
			}
			fmt.Printf("    ✅ Found in resident_contacts:\n")
			fmt.Printf("      contact_id: %s\n", contactID.String)
			fmt.Printf("      slot: %s\n", slot.String)
			fmt.Printf("      resident: %s\n", residentNickname.String)
			fmt.Printf("      email_hash matches: %v\n", emailHashHexDB.String == emailHashHex)
			fmt.Printf("      is_enabled: %v\n", isEnabled)
			fmt.Printf("      has_password_hash: %v\n", hasPasswordHash)
			if !hasPasswordHash {
				fmt.Printf("      ❌ NO password_hash - CANNOT login\n")
			} else if !isEnabled {
				fmt.Printf("      ⚠️  password_hash exists but is_enabled=false - CANNOT login\n")
			} else {
				fmt.Printf("      ✅ Has password_hash and is_enabled=true - can login\n")
			}
		}
		rows2.Close()

		if !foundInContacts {
			fmt.Printf("    ❌ NOT found in resident_contacts table\n")
		}

		fmt.Println("\n" + strings.Repeat("-", 80) + "\n")
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

