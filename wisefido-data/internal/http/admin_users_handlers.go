package httpapi

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lib/pq"
)

func (s *StubHandler) AdminUsers(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/admin/api/v1/users" {
		switch r.Method {
		case http.MethodGet:
			if s != nil && s.DB != nil {
				tenantID, ok := s.tenantIDFromReq(w, r)
				if !ok {
					fmt.Printf("[AdminUsers] Failed to get tenant_id from request\n")
					return
				}
				fmt.Printf("[AdminUsers] Got tenant_id: %s\n", tenantID)
				search := strings.TrimSpace(r.URL.Query().Get("search"))
				args := []any{tenantID}
				q := `SELECT user_id::text, tenant_id::text, user_account, nickname, email, phone, role, status,
				             COALESCE(alarm_levels, ARRAY[]::varchar[]) as alarm_levels,
				             COALESCE(alarm_channels, ARRAY[]::varchar[]) as alarm_channels,
				             alarm_scope, branch_tag, last_login_at,
				             COALESCE(tags, '[]'::jsonb) as tags,
				             COALESCE(preferences, '{}'::jsonb) as preferences
				      FROM users
				      WHERE tenant_id = $1`
				if search != "" {
					args = append(args, "%"+search+"%")
					q += ` AND (user_account ILIKE $2 OR COALESCE(nickname,'') ILIKE $2 OR COALESCE(email,'') ILIKE $2 OR COALESCE(phone,'') ILIKE $2)`
				}
				q += ` ORDER BY user_account ASC`
				fmt.Printf("[AdminUsers] Executing query: %s with args: %v\n", q, args)
				rows, err := s.DB.QueryContext(r.Context(), q, args...)
				if err != nil {
					// Log the actual error for debugging
					fmt.Printf("[AdminUsers] SQL query error: %v, query: %s, args: %v\n", err, q, args)
					writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to list users: %v", err)))
					return
				}
				fmt.Printf("[AdminUsers] Query executed successfully, starting to scan rows\n")
				defer rows.Close()
				out := []any{}
				rowCount := 0
				for rows.Next() {
					rowCount++
					fmt.Printf("[AdminUsers] Scanning row %d\n", rowCount)
					var userID, tid, userAccount, role, status string
					var nickname, email, phone sql.NullString
					var alarmLevels []string
					var alarmChannels []string
					var alarmScope, branchTag sql.NullString
					var lastLoginAt sql.NullTime
					var tagsRaw, prefRaw []byte
					if err := rows.Scan(
						&userID, &tid, &userAccount, &nickname, &email, &phone, &role, &status,
						pq.Array(&alarmLevels), pq.Array(&alarmChannels), &alarmScope, &branchTag, &lastLoginAt, &tagsRaw, &prefRaw,
					); err != nil {
						// Log the actual error for debugging
						fmt.Printf("[AdminUsers] Row scan error at row %d: %v\n", rowCount, err)
						writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to list users: %v", err)))
						return
					}
					fmt.Printf("[AdminUsers] Successfully scanned row %d: user_account=%s\n", rowCount, userAccount)
					var tags []string
					if len(tagsRaw) > 0 {
						_ = json.Unmarshal(tagsRaw, &tags)
					}
					var prefs any
					if len(prefRaw) > 0 {
						_ = json.Unmarshal(prefRaw, &prefs)
					}
					item := map[string]any{
						"user_id":      userID,
						"tenant_id":    tid,
						"user_account": userAccount,
						"role":         role,
						"status":       status,
					}
					if nickname.Valid {
						item["nickname"] = nickname.String
					}
					if email.Valid {
						item["email"] = email.String
					}
					if phone.Valid {
						item["phone"] = phone.String
					}
					if alarmLevels != nil && len(alarmLevels) > 0 {
						item["alarm_levels"] = alarmLevels
					}
					if alarmChannels != nil && len(alarmChannels) > 0 {
						item["alarm_channels"] = alarmChannels
					}
					if alarmScope.Valid {
						item["alarm_scope"] = alarmScope.String
					}
					if branchTag.Valid {
						item["branch_tag"] = branchTag.String
					}
					if lastLoginAt.Valid {
						item["last_login_at"] = lastLoginAt.Time.Format(time.RFC3339)
					}
					if tags != nil {
						item["tags"] = tags
					}
					if prefs != nil {
						item["preferences"] = prefs
					}
					out = append(out, item)
				}
				// Check for errors from iterating over rows
				if err := rows.Err(); err != nil {
					fmt.Printf("[AdminUsers] Rows iteration error: %v\n", err)
					writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to list users: %v", err)))
					return
				}
				fmt.Printf("[AdminUsers] Successfully listed %d users\n", len(out))
				writeJSON(w, http.StatusOK, Ok(map[string]any{"items": out, "total": len(out)}))
				return
			}
			writeJSON(w, http.StatusOK, Fail("database not available"))
		case http.MethodPost:
			if s != nil && s.DB != nil {
				tenantID, ok := s.tenantIDFromReq(w, r)
				if !ok {
					return
				}
				var payload map[string]any
				if err := readBodyJSON(r, 1<<20, &payload); err != nil {
					writeJSON(w, http.StatusOK, Fail("invalid body"))
					return
				}
				userAccount, _ := payload["user_account"].(string)
				role, _ := payload["role"].(string)
				password, _ := payload["password"].(string)
				if strings.TrimSpace(userAccount) == "" || strings.TrimSpace(role) == "" || password == "" {
					writeJSON(w, http.StatusOK, Fail("user_account, role, password are required"))
					return
				}
				role = strings.TrimSpace(role)
				// Security: system roles can only be assigned by SystemAdmin within System tenant.
				if role == "SystemAdmin" || role == "SystemOperator" {
					if tenantID != SystemTenantID() || !strings.EqualFold(r.Header.Get("X-User-Role"), "SystemAdmin") {
						writeJSON(w, http.StatusOK, Fail("not allowed to assign system role"))
						return
					}
				}
				userAccount = strings.ToLower(strings.TrimSpace(userAccount))
				ah, _ := hex.DecodeString(HashAccount(userAccount))
				// Password hash should only depend on password itself (independent of account/phone/email)
				aph, _ := hex.DecodeString(HashPassword(password))
				if len(ah) == 0 || len(aph) == 0 {
					writeJSON(w, http.StatusOK, Fail("failed to hash credentials"))
					return
				}
				nickname, _ := payload["nickname"].(string)
				email, _ := payload["email"].(string)
				phone, _ := payload["phone"].(string)
				status := "active"
				if st, ok := payload["status"].(string); ok && st != "" {
					status = st
				}

				// Parse alarm configuration fields
				var alarmLevels pq.StringArray
				if levels, ok := payload["alarm_levels"].([]any); ok && len(levels) > 0 {
					alarmLevels = make([]string, 0, len(levels))
					for _, l := range levels {
						if s, ok := l.(string); ok && s != "" {
							alarmLevels = append(alarmLevels, s)
						}
					}
				}
				var alarmChannels pq.StringArray
				if channels, ok := payload["alarm_channels"].([]any); ok && len(channels) > 0 {
					alarmChannels = make([]string, 0, len(channels))
					for _, c := range channels {
						if s, ok := c.(string); ok && s != "" {
							alarmChannels = append(alarmChannels, s)
						}
					}
				}
				// Parse alarm_scope: set default based on role if not provided
				var alarmScope sql.NullString
				if scope, ok := payload["alarm_scope"].(string); ok && scope != "" {
					alarmScope = sql.NullString{String: scope, Valid: true}
				} else {
					// Set default alarm_scope based on role
					roleLower := strings.ToLower(role)
					if roleLower == "caregiver" || roleLower == "nurse" {
						alarmScope = sql.NullString{String: "ASSIGNED_ONLY", Valid: true}
					} else if roleLower == "manager" {
						alarmScope = sql.NullString{String: "BRANCH", Valid: true}
					}
					// Other roles: leave as NULL (no default)
				}

				// Parse tags (JSONB): store as JSON array of strings
				var tagsJSON []byte
				if tags, ok := payload["tags"].([]any); ok && len(tags) > 0 {
					tagsStr := make([]string, 0, len(tags))
					for _, t := range tags {
						if s, ok := t.(string); ok && s != "" {
							tagsStr = append(tagsStr, s)
						}
					}
					if len(tagsStr) > 0 {
						if b, err := json.Marshal(tagsStr); err == nil {
							tagsJSON = b
						}
					}
				}
				var tagsArg any = nil
				if len(tagsJSON) > 0 {
					tagsArg = tagsJSON
				}

				// Check email/phone uniqueness before insert
				if err := checkEmailUniqueness(s.DB, r, tenantID, email, ""); err != nil {
					writeJSON(w, http.StatusOK, Fail(err.Error()))
					return
				}
				if err := checkPhoneUniqueness(s.DB, r, tenantID, phone, ""); err != nil {
					writeJSON(w, http.StatusOK, Fail(err.Error()))
					return
				}

				var userID string
				err := s.DB.QueryRowContext(
					r.Context(),
					`INSERT INTO users (tenant_id, user_account, user_account_hash, password_hash, nickname, email, phone, role, status, alarm_levels, alarm_channels, alarm_scope, tags)
					 VALUES ($1,$2,$3,$4,NULLIF($5,''),NULLIF($6,''),NULLIF($7,''),$8,$9,$10,$11,$12,$13)
					 RETURNING user_id::text`,
					tenantID, userAccount, ah, aph, nickname, email, phone, role, status,
					pq.Array(alarmLevels), pq.Array(alarmChannels), alarmScope, tagsArg,
				).Scan(&userID)
				if err != nil {
					// Check for unique constraint violation
					if msg := checkUniqueConstraintError(err, "email or phone"); msg != "" {
						writeJSON(w, http.StatusOK, Fail(msg))
						return
					}
					writeJSON(w, http.StatusOK, Fail("failed to create user"))
					return
				}
				// Optional: allow dev login via AuthStore as well (keeps current auth flow).
				if s.AuthStore != nil {
					_ = s.AuthStore.UpsertUser(tenantID, userAccount, role, password)
				}
				writeJSON(w, http.StatusOK, Ok(map[string]any{"user_id": userID}))
				return
			}
			writeJSON(w, http.StatusOK, Fail("database not available"))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}
	if strings.HasPrefix(r.URL.Path, "/admin/api/v1/users/") {
		path := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/users/")
		if strings.HasSuffix(path, "/reset-password") {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			if s != nil && s.DB != nil {
				tenantID, ok := s.tenantIDFromReq(w, r)
				if !ok {
					return
				}
				userID := strings.TrimSuffix(path, "/reset-password")
				if userID == "" || strings.Contains(userID, "/") {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				var payload map[string]any
				if err := readBodyJSON(r, 1<<20, &payload); err != nil {
					writeJSON(w, http.StatusOK, Fail("invalid body"))
					return
				}
				newPassword, _ := payload["new_password"].(string)
				if newPassword == "" {
					writeJSON(w, http.StatusOK, Fail("new_password is required"))
					return
				}

				// Look up user_account for hashing
				var userAccount, role string
				err := s.DB.QueryRowContext(
					r.Context(),
					`SELECT user_account, role
					   FROM users
					  WHERE tenant_id = $1 AND user_id::text = $2`,
					tenantID, userID,
				).Scan(&userAccount, &role)
				if err != nil {
					writeJSON(w, http.StatusOK, Fail("user not found"))
					return
				}

				// Hash password: sha256(password) - only depends on password itself (independent of account/phone/email)
				aph, _ := hex.DecodeString(HashPassword(newPassword))
				if len(aph) == 0 {
					writeJSON(w, http.StatusOK, Fail("failed to hash password"))
					return
				}
				_, err = s.DB.ExecContext(
					r.Context(),
					`UPDATE users SET password_hash = $3
					  WHERE tenant_id = $1 AND user_id::text = $2`,
					tenantID, userID, aph,
				)
				if err != nil {
					writeJSON(w, http.StatusOK, Fail("failed to reset password"))
					return
				}
				// Optional: keep AuthStore in sync for dev/stub flows.
				if s.AuthStore != nil {
					_ = s.AuthStore.UpsertUser(tenantID, userAccount, role, newPassword)
				}
				writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true, "message": "ok"}))
				return
			}
			writeJSON(w, http.StatusOK, Fail("database not available"))
			return
		}
		if strings.HasSuffix(path, "/reset-pin") {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			if s != nil && s.DB != nil {
				tenantID, ok := s.tenantIDFromReq(w, r)
				if !ok {
					return
				}
				userID := strings.TrimSuffix(path, "/reset-pin")
				if userID == "" || strings.Contains(userID, "/") {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				var payload map[string]any
				if err := readBodyJSON(r, 1<<20, &payload); err != nil {
					writeJSON(w, http.StatusOK, Fail("invalid body"))
					return
				}
				newPin, _ := payload["new_pin"].(string)
				if newPin == "" {
					writeJSON(w, http.StatusOK, Fail("new_pin is required"))
					return
				}
				// Validate PIN: must be exactly 4 digits
				if len(newPin) != 4 {
					writeJSON(w, http.StatusOK, Fail("PIN must be exactly 4 digits"))
					return
				}
				for _, c := range newPin {
					if c < '0' || c > '9' {
						writeJSON(w, http.StatusOK, Fail("PIN must contain only digits"))
						return
					}
				}

				// Hash PIN: sha256(pin) - only depends on PIN itself
				pinHash, _ := hex.DecodeString(HashPassword(newPin))
				if len(pinHash) == 0 {
					writeJSON(w, http.StatusOK, Fail("failed to hash PIN"))
					return
				}
				_, err := s.DB.ExecContext(
					r.Context(),
					`UPDATE users SET pin_hash = $3
					  WHERE tenant_id = $1 AND user_id::text = $2`,
					tenantID, userID, pinHash,
				)
				if err != nil {
					writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to reset PIN: %v", err)))
					return
				}
				writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
				return
			}
			writeJSON(w, http.StatusOK, Fail("database not available"))
			return
		}
		id := path
		if strings.Contains(id, "/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodPut:
			if s != nil && s.DB != nil {
				tenantID, ok := s.tenantIDFromReq(w, r)
				if !ok {
					return
				}
				var payload map[string]any
				if err := readBodyJSON(r, 1<<20, &payload); err != nil {
					writeJSON(w, http.StatusOK, Fail("invalid body"))
					return
				}
				// Soft delete via {_delete:true}
				if del, ok := payload["_delete"].(bool); ok && del {
					_, err := s.DB.ExecContext(
						r.Context(),
						`UPDATE users SET status = 'left' WHERE tenant_id = $1 AND user_id::text = $2`,
						tenantID, id,
					)
					if err != nil {
						writeJSON(w, http.StatusOK, Fail("failed to delete user"))
						return
					}
					writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
					return
				}

				// Update editable fields
				nickname, _ := payload["nickname"].(string)
				email, _ := payload["email"].(string)
				phone, _ := payload["phone"].(string)
				role, _ := payload["role"].(string)
				status, _ := payload["status"].(string)
				role = strings.TrimSpace(role)
				status = strings.TrimSpace(status)

				// Security: system roles can only be assigned by SystemAdmin within System tenant.
				if role == "SystemAdmin" || role == "SystemOperator" {
					if tenantID != SystemTenantID() || !strings.EqualFold(r.Header.Get("X-User-Role"), "SystemAdmin") {
						writeJSON(w, http.StatusOK, Fail("not allowed to assign system role"))
						return
					}
				}

				// Only allow known status changes; otherwise keep unchanged.
				if status != "" && status != "active" && status != "disabled" && status != "left" {
					writeJSON(w, http.StatusOK, Fail("invalid status"))
					return
				}

				// Parse alarm configuration fields (only update if provided)
				var alarmLevels pq.StringArray
				if levels, ok := payload["alarm_levels"].([]any); ok {
					alarmLevels = make([]string, 0, len(levels))
					for _, l := range levels {
						if s, ok := l.(string); ok && s != "" {
							alarmLevels = append(alarmLevels, s)
						}
					}
				}
				var alarmChannels pq.StringArray
				if channels, ok := payload["alarm_channels"].([]any); ok {
					alarmChannels = make([]string, 0, len(channels))
					for _, c := range channels {
						if s, ok := c.(string); ok && s != "" {
							alarmChannels = append(alarmChannels, s)
						}
					}
				}
				var alarmScope sql.NullString
				if scope, ok := payload["alarm_scope"].(string); ok {
					if scope != "" {
						alarmScope = sql.NullString{String: scope, Valid: true}
					}
				}

				// Parse tags (JSONB): only update if provided
				var tagsJSON []byte
				var tagsProvided bool
				if tags, ok := payload["tags"].([]any); ok {
					tagsProvided = true
					tagsStr := make([]string, 0, len(tags))
					for _, t := range tags {
						if s, ok := t.(string); ok && s != "" {
							tagsStr = append(tagsStr, s)
						}
					}
					// Always marshal (even empty array) to allow clearing tags
					if b, err := json.Marshal(tagsStr); err == nil {
						tagsJSON = b
					}
				}

				// Build dynamic UPDATE query based on what fields are provided
				updates := []string{}
				args := []any{tenantID, id}
				argIdx := 3

				if nickname != "" {
					updates = append(updates, fmt.Sprintf("nickname = $%d", argIdx))
					args = append(args, nickname)
					argIdx++
				}
				if email != "" {
					updates = append(updates, fmt.Sprintf("email = $%d", argIdx))
					args = append(args, email)
					argIdx++
				}
				if phone != "" {
					updates = append(updates, fmt.Sprintf("phone = $%d", argIdx))
					args = append(args, phone)
					argIdx++
				}
				if role != "" {
					updates = append(updates, fmt.Sprintf("role = $%d", argIdx))
					args = append(args, role)
					argIdx++
				}
				if status != "" {
					updates = append(updates, fmt.Sprintf("status = $%d", argIdx))
					args = append(args, status)
					argIdx++
				}
				if _, ok := payload["alarm_levels"]; ok {
					updates = append(updates, fmt.Sprintf("alarm_levels = $%d", argIdx))
					args = append(args, pq.Array(alarmLevels))
					argIdx++
				}
				if _, ok := payload["alarm_channels"]; ok {
					updates = append(updates, fmt.Sprintf("alarm_channels = $%d", argIdx))
					args = append(args, pq.Array(alarmChannels))
					argIdx++
				}
				if _, ok := payload["alarm_scope"]; ok {
					updates = append(updates, fmt.Sprintf("alarm_scope = $%d", argIdx))
					args = append(args, alarmScope)
					argIdx++
				}
				if tagsProvided {
					updates = append(updates, fmt.Sprintf("tags = $%d", argIdx))
					args = append(args, tagsJSON) // tagsJSON is always set when tagsProvided is true
					argIdx++
				}
				if branchTag, ok := payload["branch_tag"].(string); ok {
					updates = append(updates, fmt.Sprintf("branch_tag = $%d", argIdx))
					if branchTag != "" {
						args = append(args, branchTag)
					} else {
						args = append(args, nil) // Set to NULL if empty string
					}
					argIdx++
				}

				if len(updates) == 0 {
					writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
					return
				}

				// Check email/phone uniqueness before update (if being updated)
				if email != "" {
					if err := checkEmailUniqueness(s.DB, r, tenantID, email, id); err != nil {
						writeJSON(w, http.StatusOK, Fail(err.Error()))
						return
					}
				}
				if phone != "" {
					if err := checkPhoneUniqueness(s.DB, r, tenantID, phone, id); err != nil {
						writeJSON(w, http.StatusOK, Fail(err.Error()))
						return
					}
				}

				query := fmt.Sprintf(`UPDATE users SET %s WHERE tenant_id = $1 AND user_id::text = $2`, strings.Join(updates, ", "))
				_, err := s.DB.ExecContext(r.Context(), query, args...)
				if err != nil {
					// Check for unique constraint violation
					if msg := checkUniqueConstraintError(err, "email or phone"); msg != "" {
						writeJSON(w, http.StatusOK, Fail(msg))
						return
					}
					writeJSON(w, http.StatusOK, Fail("failed to update user"))
					return
				}
				writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
				return
			}
			writeJSON(w, http.StatusOK, Fail("database not available"))
			return
		case http.MethodDelete:
			// Soft delete: keep row for audit, mark as left.
			if s != nil && s.DB != nil {
				tenantID, ok := s.tenantIDFromReq(w, r)
				if !ok {
					return
				}
				_, err := s.DB.ExecContext(
					r.Context(),
					`UPDATE users SET status = 'left' WHERE tenant_id = $1 AND user_id::text = $2`,
					tenantID, id,
				)
				if err != nil {
					writeJSON(w, http.StatusOK, Fail("failed to delete user"))
					return
				}
				writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
				return
			}
			writeJSON(w, http.StatusOK, Fail("database not available"))
		case http.MethodGet:
			// 注意：owlFront 这里有一个 getUserApi 误用 Update URL（但 method=GET）
			if s != nil && s.DB != nil {
				tenantID, ok := s.tenantIDFromReq(w, r)
				if !ok {
					return
				}
				var (
					userID, tenantIDStr, userAccount, nickname, email, phone, role, status string
					alarmLevels, alarmChannels                                             []string
					alarmScope, branchTag                                                  sql.NullString
					lastLoginAt                                                            sql.NullTime
					tagsRaw, prefRaw                                                       []byte
				)
				err := s.DB.QueryRowContext(
					r.Context(),
					`SELECT user_id::text,
					        tenant_id::text,
					        user_account,
					        COALESCE(nickname,''),
					        COALESCE(email,''),
					        COALESCE(phone,''),
					        role,
					        COALESCE(status,'active'),
					        COALESCE(alarm_levels, ARRAY[]::varchar[]) as alarm_levels,
					        COALESCE(alarm_channels, ARRAY[]::varchar[]) as alarm_channels,
					        alarm_scope,
					        branch_tag,
					        last_login_at,
					        COALESCE(tags, '[]'::jsonb) as tags,
					        COALESCE(preferences, '{}'::jsonb) as preferences
					   FROM users
					  WHERE tenant_id = $1 AND user_id::text = $2`,
					tenantID, id,
				).Scan(
					&userID, &tenantIDStr, &userAccount, &nickname, &email, &phone, &role, &status,
					pq.Array(&alarmLevels), pq.Array(&alarmChannels), &alarmScope, &branchTag, &lastLoginAt, &tagsRaw, &prefRaw,
				)
				if err != nil {
					writeJSON(w, http.StatusOK, Fail("failed to get user"))
					return
				}
				item := map[string]any{
					"user_id":      userID,
					"tenant_id":    tenantIDStr,
					"user_account": userAccount,
					"nickname":     nickname,
					"email":        email,
					"phone":        phone,
					"role":         role,
					"status":       status,
				}
				if alarmLevels != nil && len(alarmLevels) > 0 {
					item["alarm_levels"] = alarmLevels
				}
				if alarmChannels != nil && len(alarmChannels) > 0 {
					item["alarm_channels"] = alarmChannels
				}
				if alarmScope.Valid {
					item["alarm_scope"] = alarmScope.String
				}
				if branchTag.Valid {
					item["branch_tag"] = branchTag.String
				}
				if lastLoginAt.Valid {
					item["last_login_at"] = lastLoginAt.Time.Format(time.RFC3339)
				}
				if len(tagsRaw) > 0 {
					var tags any
					_ = json.Unmarshal(tagsRaw, &tags)
					item["tags"] = tags
				}
				if len(prefRaw) > 0 {
					var prefs any
					_ = json.Unmarshal(prefRaw, &prefs)
					item["preferences"] = prefs
				}
				writeJSON(w, http.StatusOK, Ok(item))
				return
			}
			writeJSON(w, http.StatusOK, Fail("database not available"))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}
	w.WriteHeader(http.StatusNotFound)
}
