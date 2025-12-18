package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"

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
			rc.email_hash,
			rc.phone_hash,
			rc.password_hash
		FROM residents r
		JOIN resident_contacts rc ON r.resident_id = rc.resident_id
		WHERE rc.contact_email LIKE '%ding3%'
		ORDER BY r.nickname, rc.slot
	`)
	if err != nil {
		log.Fatalf("Failed to query contacts: %v", err)
	}
	defer rows.Close()

	fmt.Println("=== Contacts with ding3@gmail.com ===\n")

	for rows.Next() {
		var nickname, slot, contactID, contactEmail sql.NullString
		var emailHash, phoneHash, passwordHash []byte

		err = rows.Scan(
			&nickname, &slot, &contactID, &contactEmail,
			&emailHash, &phoneHash, &passwordHash,
		)
		if err != nil {
			log.Fatalf("Failed to scan contact: %v", err)
		}

		fmt.Printf("Resident: %s, Slot: %s\n", nickname.String, slot.String)
		fmt.Printf("  contact_id: %s\n", contactID.String)
		fmt.Printf("  contact_email: %s\n", contactEmail.String)

		// Check if email should be saved (if email exists, it should have hash)
		if contactEmail.String != "" {
			fmt.Printf("  ❌ Email is saved (should NOT be saved if save is unchecked)\n")
		} else {
			fmt.Printf("  ✅ Email is NOT saved (correct)\n")
		}

		// Check email hash
		if emailHash != nil {
			emailHashHex := hex.EncodeToString(emailHash)
			fmt.Printf("  email_hash: %s\n", emailHashHex)

			// Calculate expected hash for ding3@gmail.com
			expectedEmail := "ding3@gmail.com"
			expectedHash := sha256.Sum256([]byte(expectedEmail))
			expectedHashHex := hex.EncodeToString(expectedHash[:])
			fmt.Printf("  Expected hash for 'ding3@gmail.com': %s\n", expectedHashHex)

			if emailHashHex == expectedHashHex {
				fmt.Printf("  ✅ Email hash matches 'ding3@gmail.com'\n")
			} else {
				fmt.Printf("  ❌ Email hash does NOT match 'ding3@gmail.com'\n")
			}
		} else {
			fmt.Printf("  ❌ email_hash is NULL (should exist for login)\n")
		}

		// Check password hash
		if passwordHash != nil {
			fmt.Printf("  password_hash: exists\n")
		} else {
			fmt.Printf("  password_hash: NULL\n")
		}

		fmt.Println()
	}

	// Also check if ding3@gmail.com can login (check hash in residents table)
	fmt.Println("=== Checking if ding3@gmail.com can login (checking residents.email_hash) ===\n")
	expectedEmail := "ding3@gmail.com"
	expectedHash := sha256.Sum256([]byte(expectedEmail))
	expectedHashHex := hex.EncodeToString(expectedHash[:])

	rows2, err := db.Query(`
		SELECT 
			r.nickname,
			r.resident_id::text,
			r.email_hash,
			r.phone_hash
		FROM residents r
		WHERE r.email_hash = $1
	`, expectedHash[:])
	if err != nil {
		log.Fatalf("Failed to query residents: %v", err)
	}
	defer rows2.Close()

	found := false
	for rows2.Next() {
		found = true
		var nickname, residentID sql.NullString
		var emailHash, phoneHash []byte
		err = rows2.Scan(&nickname, &residentID, &emailHash, &phoneHash)
		if err != nil {
			log.Fatalf("Failed to scan resident: %v", err)
		}
		fmt.Printf("✅ Found resident with matching email_hash:\n")
		fmt.Printf("  resident_id: %s\n", residentID.String)
		fmt.Printf("  nickname: %s\n", nickname.String)
		fmt.Printf("  email_hash: %s\n", hex.EncodeToString(emailHash))
	}

	if !found {
		fmt.Printf("❌ No resident found with email_hash matching 'ding3@gmail.com'\n")
		fmt.Printf("  Expected hash: %s\n", expectedHashHex)
		fmt.Printf("  This means ding3@gmail.com CANNOT login as resident\n")
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

