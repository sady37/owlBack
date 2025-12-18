package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"strings"

	"wisefido-data/internal/config"

	_ "github.com/lib/pq"
)

// HashPassword calculates SHA256 hash of password
func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func main() {
	// Parse command line arguments
	var residentIDs = flag.String("ids", "", "Comma-separated list of resident IDs or account names (e.g., 'r1,r2,r3' or 'done')")
	var password = flag.String("password", "", "Password to check hash (e.g., 'Ts123@123')")
	var showAll = flag.Bool("all", false, "Show all residents (limit 100)")
	var showPHI = flag.Bool("phi", false, "Show PHI data (email, phone) for residents")
	var dbName = flag.String("db", "", "Database name (default: try 'wisefido_data' then 'owlrd')")
	flag.Parse()

	// Load config
	cfg := config.Load()

	// Determine database name
	dbNames := []string{"wisefido_data", "owlrd"}
	if *dbName != "" {
		dbNames = []string{*dbName}
	}

	var db *sql.DB
	var connectedDB string
	var err error

	// Try to connect to database
	for _, name := range dbNames {
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.User,
			cfg.Database.Password,
			name,
			cfg.Database.SSLMode,
		)
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

	// If password is provided, show hash
	if *password != "" {
		hash := HashPassword(*password)
		fmt.Printf("Password Hash Comparison:\n")
		fmt.Printf("Password: %s -> Hash: %s\n\n", *password, hash)
	}

	// Query residents
	var rows *sql.Rows
	var queryPHI = *showPHI

	if *showAll {
		// Show all residents (limited)
		if queryPHI {
			rows, err = db.Query(`
				SELECT r.resident_id, r.resident_account, r.nickname, r.status, 
				       encode(r.password_hash, 'hex') as password_hash_hex,
				       encode(r.email_hash, 'hex') as email_hash_hex,
				       encode(r.phone_hash, 'hex') as phone_hash_hex,
				       rp.resident_email, rp.resident_phone,
				       rp.first_name, rp.last_name
				FROM residents r
				LEFT JOIN resident_phi rp ON rp.resident_id = r.resident_id AND rp.tenant_id = r.tenant_id
				ORDER BY r.resident_id
				LIMIT 100
			`)
		} else {
			rows, err = db.Query(`
				SELECT resident_id, resident_account, nickname, status, 
				       encode(password_hash, 'hex') as password_hash_hex
				FROM residents 
				ORDER BY resident_id
				LIMIT 100
			`)
		}
		if err != nil {
			log.Fatalf("Query error: %v", err)
		}
	} else if *residentIDs != "" {
		// Query specific residents by ID or account
		ids := strings.Split(*residentIDs, ",")
		trimmedIDs := make([]string, len(ids))
		for i, id := range ids {
			trimmedIDs[i] = strings.TrimSpace(id)
		}

		// Build query with IN clause - renumber parameters for second condition
		placeholders1 := make([]string, len(trimmedIDs))
		placeholders2 := make([]string, len(trimmedIDs))
		args := make([]interface{}, len(trimmedIDs))
		for i, id := range trimmedIDs {
			placeholders1[i] = fmt.Sprintf("$%d", i+1)
			placeholders2[i] = fmt.Sprintf("$%d", i+1+len(trimmedIDs))
			args[i] = id
		}

		var query string
		if queryPHI {
			query = fmt.Sprintf(`
				SELECT r.resident_id, r.resident_account, r.nickname, r.status, 
				       encode(r.password_hash, 'hex') as password_hash_hex,
				       encode(r.email_hash, 'hex') as email_hash_hex,
				       encode(r.phone_hash, 'hex') as phone_hash_hex,
				       rp.resident_email, rp.resident_phone,
				       rp.first_name, rp.last_name
				FROM residents r
				LEFT JOIN resident_phi rp ON rp.resident_id = r.resident_id AND rp.tenant_id = r.tenant_id
				WHERE r.resident_id::text IN (%s) OR r.resident_account IN (%s)
				ORDER BY r.resident_id
			`, strings.Join(placeholders1, ","), strings.Join(placeholders2, ","))
		} else {
			query = fmt.Sprintf(`
				SELECT resident_id, resident_account, nickname, status, 
				       encode(password_hash, 'hex') as password_hash_hex
				FROM residents 
				WHERE resident_id::text IN (%s) OR resident_account IN (%s)
				ORDER BY resident_id
			`, strings.Join(placeholders1, ","), strings.Join(placeholders2, ","))
		}

		// Duplicate args for both conditions (resident_id and resident_account)
		allArgs := make([]interface{}, len(args)*2)
		copy(allArgs, args)
		copy(allArgs[len(args):], args)

		rows, err = db.Query(query, allArgs...)
		if err != nil {
			log.Fatalf("Query error: %v", err)
		}
	} else {
		// Default: show r1, r2, r3
		if queryPHI {
			rows, err = db.Query(`
				SELECT r.resident_id, r.resident_account, r.nickname, r.status, 
				       encode(r.password_hash, 'hex') as password_hash_hex,
				       encode(r.email_hash, 'hex') as email_hash_hex,
				       encode(r.phone_hash, 'hex') as phone_hash_hex,
				       rp.resident_email, rp.resident_phone,
				       rp.first_name, rp.last_name
				FROM residents r
				LEFT JOIN resident_phi rp ON rp.resident_id = r.resident_id AND rp.tenant_id = r.tenant_id
				WHERE r.resident_id::text IN ('r1', 'r2', 'r3') OR r.resident_account IN ('r1', 'r2', 'r3')
				ORDER BY r.resident_id
			`)
		} else {
			rows, err = db.Query(`
				SELECT resident_id, resident_account, nickname, status, 
				       encode(password_hash, 'hex') as password_hash_hex
				FROM residents 
				WHERE resident_id::text IN ('r1', 'r2', 'r3') OR resident_account IN ('r1', 'r2', 'r3')
				ORDER BY resident_id
			`)
		}
		if err != nil {
			log.Fatalf("Query error: %v", err)
		}
	}
	defer rows.Close()

	if queryPHI {
		fmt.Println("Resident Records with PHI Data:")
		fmt.Println("ID | Account | Nickname | Status | First Name | Last Name | Email | Phone | Email Hash | Phone Hash | Password Hash")
		fmt.Println("---|--------|----------|--------|------------|-----------|-------|-------|------------|------------|---------------")
	} else {
		fmt.Println("Resident Records:")
		fmt.Println("ID | Account | Nickname | Status | Password Hash (hex)")
		fmt.Println("---|--------|----------|--------|-------------------")
	}

	found := false
	for rows.Next() {
		found = true
		if queryPHI {
			var residentID, residentAccount, nickname, status, passwordHashHex sql.NullString
			var emailHashHex, phoneHashHex, residentEmail, residentPhone sql.NullString
			var firstName, lastName sql.NullString
			if err := rows.Scan(&residentID, &residentAccount, &nickname, &status, &passwordHashHex,
				&emailHashHex, &phoneHashHex, &residentEmail, &residentPhone, &firstName, &lastName); err != nil {
				log.Printf("Scan error: %v", err)
				continue
			}
			fmt.Printf("%s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s\n",
				getString(residentID), getString(residentAccount), getString(nickname), getString(status),
				getString(firstName), getString(lastName),
				getString(residentEmail), getString(residentPhone),
				getString(emailHashHex), getString(phoneHashHex), getString(passwordHashHex))
		} else {
			var residentID, residentAccount, nickname, status, passwordHashHex sql.NullString
			if err := rows.Scan(&residentID, &residentAccount, &nickname, &status, &passwordHashHex); err != nil {
				log.Printf("Scan error: %v", err)
				continue
			}
			fmt.Printf("%s | %s | %s | %s | %s\n",
				getString(residentID), getString(residentAccount), getString(nickname), getString(status), getString(passwordHashHex))
		}
	}

	if !found {
		fmt.Println("(No records found)")
	}

	// If password was provided, show matching status
	if *password != "" && found {
		hash := HashPassword(*password)
		fmt.Printf("\nPassword Hash: %s\n", hash)
		fmt.Println("(Compare with password_hash_hex above to verify password)")
	}
}

func getString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return "NULL"
}
