package httpapi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/lib/pq"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func parseInt(s string, def int) int {
	if s == "" {
		return def
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return i
}

func readBodyJSON(r *http.Request, maxBytes int64, out any) error {
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBytes))
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return nil
	}
	return json.Unmarshal(body, out)
}

// checkUniqueConstraintError checks if an error is a unique constraint violation
// and returns a user-friendly error message
func checkUniqueConstraintError(err error, fieldName string) string {
	if err == nil {
		return ""
	}
	// Check for PostgreSQL unique constraint violation (error code 23505)
	if pqErr, ok := err.(*pq.Error); ok {
		if pqErr.Code == "23505" { // unique_violation
			// Check which constraint was violated
			if strings.Contains(pqErr.Constraint, "email") {
				return "email already exists in this organization"
			}
			if strings.Contains(pqErr.Constraint, "phone") {
				return "phone already exists in this organization"
			}
			if strings.Contains(pqErr.Constraint, "email_hash") {
				return "email already exists in this organization"
			}
			if strings.Contains(pqErr.Constraint, "phone_hash") {
				return "phone already exists in this organization"
			}
			return fmt.Sprintf("%s already exists in this organization", fieldName)
		}
	}
	return ""
}

// checkUnitUniqueConstraintError checks if an error is a unit uniqueness constraint violation
// and returns a user-friendly error message
func checkUnitUniqueConstraintError(err error) string {
	if err == nil {
		return ""
	}
	// Check for PostgreSQL unique constraint violation (error code 23505)
	if pqErr, ok := err.(*pq.Error); ok {
		if pqErr.Code == "23505" { // unique_violation
			// Check if it's a units table unique constraint
			if strings.Contains(pqErr.Constraint, "units_unique") || strings.Contains(pqErr.Message, "units") {
				// Extract building, floor, and unit_name from error message if possible
				// Error message format: "duplicate key value violates unique constraint ... Key (tenant_id, branch_tag, building, floor, unit_name)=(...)"
				if strings.Contains(pqErr.Message, "building") && strings.Contains(pqErr.Message, "floor") {
					return "A unit with the same name already exists in this building and floor. Please use a different unit name or select a different floor."
				}
				return "A unit with the same name already exists in this location. Please use a different unit name."
			}
		}
	}
	return ""
}

// checkEmailUniqueness checks if email already exists in users table for the given tenant
func checkEmailUniqueness(db *sql.DB, r *http.Request, tenantID, email, excludeUserID string) error {
	if email == "" {
		return nil
	}
	var query string
	var args []interface{}
	if excludeUserID != "" {
		query = `SELECT COUNT(*) FROM users WHERE tenant_id = $1 AND email = $2 AND user_id::text != $3`
		args = []interface{}{tenantID, email, excludeUserID}
	} else {
		query = `SELECT COUNT(*) FROM users WHERE tenant_id = $1 AND email = $2`
		args = []interface{}{tenantID, email}
	}
	var count int
	if err := db.QueryRowContext(r.Context(), query, args...).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("email already exists in this organization")
	}
	return nil
}

// checkPhoneUniqueness checks if phone already exists in users table for the given tenant
func checkPhoneUniqueness(db *sql.DB, r *http.Request, tenantID, phone, excludeUserID string) error {
	if phone == "" {
		return nil
	}
	var query string
	var args []interface{}
	if excludeUserID != "" {
		query = `SELECT COUNT(*) FROM users WHERE tenant_id = $1 AND phone = $2 AND user_id::text != $3`
		args = []interface{}{tenantID, phone, excludeUserID}
	} else {
		query = `SELECT COUNT(*) FROM users WHERE tenant_id = $1 AND phone = $2`
		args = []interface{}{tenantID, phone}
	}
	var count int
	if err := db.QueryRowContext(r.Context(), query, args...).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("phone already exists in this organization")
	}
	return nil
}

// checkHashUniqueness checks if phone_hash or email_hash already exists in residents/resident_contacts table
func checkHashUniqueness(db *sql.DB, r *http.Request, tenantID, tableName string, phoneHash, emailHash []byte, excludeID, excludeField string) error {
	if phoneHash != nil && len(phoneHash) > 0 {
		var query string
		var args []interface{}
		if excludeID != "" {
			query = fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE tenant_id = $1 AND phone_hash = $2 AND %s::text != $3`, tableName, excludeField)
			args = []interface{}{tenantID, phoneHash, excludeID}
		} else {
			query = fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE tenant_id = $1 AND phone_hash = $2`, tableName)
			args = []interface{}{tenantID, phoneHash}
		}
		var count int
		if err := db.QueryRowContext(r.Context(), query, args...).Scan(&count); err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("phone already exists in this organization")
		}
	}
	if emailHash != nil && len(emailHash) > 0 {
		var query string
		var args []interface{}
		if excludeID != "" {
			query = fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE tenant_id = $1 AND email_hash = $2 AND %s::text != $3`, tableName, excludeField)
			args = []interface{}{tenantID, emailHash, excludeID}
		} else {
			query = fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE tenant_id = $1 AND email_hash = $2`, tableName)
			args = []interface{}{tenantID, emailHash}
		}
		var count int
		if err := db.QueryRowContext(r.Context(), query, args...).Scan(&count); err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("email already exists in this organization")
		}
	}
	return nil
}




