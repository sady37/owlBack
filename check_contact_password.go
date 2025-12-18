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

	// Calculate expected password hash for "Ts123@123"
	password := "Ts123@123"
	hash := sha256.Sum256([]byte(password))
	expectedHashHex := hex.EncodeToString(hash[:])
	fmt.Printf("Expected password hash for 'Ts123@123': %s\n\n", expectedHashHex)

	// Query contacts for "done" resident
	rows, err := db.Query(`
		SELECT 
			rc.contact_id::text,
			rc.slot,
			rc.contact_first_name,
			rc.contact_last_name,
			rc.contact_phone,
			rc.contact_email,
			rc.password_hash,
			rc.phone_hash,
			rc.email_hash
		FROM residents r
		JOIN resident_contacts rc ON r.resident_id = rc.resident_id
		WHERE LOWER(r.nickname) = 'done'
		ORDER BY rc.slot
	`)
	if err != nil {
		log.Fatalf("Failed to query contacts: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var contactID, slot, firstName, lastName, contactPhone, contactEmail sql.NullString
		var passwordHash, phoneHash, emailHash []byte

		err = rows.Scan(
			&contactID, &slot, &firstName, &lastName,
			&contactPhone, &contactEmail,
			&passwordHash, &phoneHash, &emailHash,
		)
		if err != nil {
			log.Fatalf("Failed to scan contact: %v", err)
		}

		fmt.Printf("=== Contact Slot %s ===\n", slot.String)
		fmt.Printf("  contact_id: %s\n", contactID.String)
		fmt.Printf("  contact_first_name: %s\n", firstName.String)
		fmt.Printf("  contact_last_name: %s\n", lastName.String)
		fmt.Printf("  contact_phone: %s\n", contactPhone.String)
		fmt.Printf("  contact_email: %s\n", contactEmail.String)

		if passwordHash != nil {
			actualHashHex := hex.EncodeToString(passwordHash)
			fmt.Printf("  password_hash: %s\n", actualHashHex)
			if actualHashHex == expectedHashHex {
				fmt.Printf("  ✅ Password hash matches 'Ts123@123'\n")
			} else {
				fmt.Printf("  ❌ Password hash does NOT match 'Ts123@123'\n")
				fmt.Printf("     Expected: %s\n", expectedHashHex)
				fmt.Printf("     Actual:   %s\n", actualHashHex)
			}
		} else {
			fmt.Printf("  password_hash: NULL\n")
			fmt.Printf("  ❌ Password hash is NULL\n")
		}

		if phoneHash != nil {
			fmt.Printf("  phone_hash: %s\n", hex.EncodeToString(phoneHash))
		} else {
			fmt.Printf("  phone_hash: NULL\n")
		}

		if emailHash != nil {
			fmt.Printf("  email_hash: %s\n", hex.EncodeToString(emailHash))
		} else {
			fmt.Printf("  email_hash: NULL\n")
		}

		// Check if email should be saved
		if slot.String == "A" {
			// Slot A should have ding@gmail.com saved
			expectedEmail := "ding@gmail.com"
			if contactEmail.String == expectedEmail {
				fmt.Printf("  ✅ Email saved correctly: %s\n", expectedEmail)
			} else {
				fmt.Printf("  ❌ Email NOT saved! Expected: %s, Actual: %s\n", expectedEmail, contactEmail.String)
			}
			// Calculate expected email hash
			emailHashCalc := sha256.Sum256([]byte(expectedEmail))
			expectedEmailHashHex := hex.EncodeToString(emailHashCalc[:])
			if emailHash != nil {
				actualEmailHashHex := hex.EncodeToString(emailHash)
				if actualEmailHashHex == expectedEmailHashHex {
					fmt.Printf("  ✅ Email hash matches\n")
				} else {
					fmt.Printf("  ❌ Email hash mismatch! Expected: %s, Actual: %s\n", expectedEmailHashHex, actualEmailHashHex)
				}
			} else {
				fmt.Printf("  ❌ Email hash is NULL (should exist)\n")
			}
		} else if slot.String == "B" {
			// Slot B should NOT have ding2@gmail.com saved (only hash)
			expectedEmail := "ding2@gmail.com"
			if contactEmail.String == "" {
				fmt.Printf("  ✅ Email NOT saved (as expected for slot B)\n")
			} else {
				fmt.Printf("  ❌ Email should NOT be saved! Actual: %s\n", contactEmail.String)
			}
			// Calculate expected email hash
			emailHashCalc := sha256.Sum256([]byte(expectedEmail))
			expectedEmailHashHex := hex.EncodeToString(emailHashCalc[:])
			if emailHash != nil {
				actualEmailHashHex := hex.EncodeToString(emailHash)
				if actualEmailHashHex == expectedEmailHashHex {
					fmt.Printf("  ✅ Email hash matches (email not saved, only hash)\n")
				} else {
					fmt.Printf("  ❌ Email hash mismatch! Expected: %s, Actual: %s\n", expectedEmailHashHex, actualEmailHashHex)
				}
			} else {
				fmt.Printf("  ❌ Email hash is NULL (should exist for slot B)\n")
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

