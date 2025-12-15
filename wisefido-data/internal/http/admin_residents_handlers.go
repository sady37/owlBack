package httpapi

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func (s *StubHandler) AdminResidents(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/admin/api/v1/residents" {
		switch r.Method {
		case http.MethodGet:
			if s != nil && s.DB != nil {
				tenantID, ok := s.tenantIDFromReq(w, r)
				if !ok {
					return
				}
				search := strings.TrimSpace(r.URL.Query().Get("search"))
				statusFilter := strings.TrimSpace(r.URL.Query().Get("status"))
				serviceLevelFilter := strings.TrimSpace(r.URL.Query().Get("service_level"))

				// Get current user info for permission-based filtering
				userID := r.Header.Get("X-User-Id")
				var userRole, alarmScope, userBranchTag sql.NullString
				var isResidentLogin bool
				var residentIDForSelf sql.NullString

				// Check if this is a resident login (userType from auth)
				userType := r.Header.Get("X-User-Type")
				if userType == "resident" || userType == "family" {
					isResidentLogin = true
					// For resident login, X-User-Id is actually resident_id
					if userID != "" {
						residentIDForSelf = sql.NullString{String: userID, Valid: true}
					}
				} else if userID != "" {
					// For staff login, get role and alarm_scope from users table
					err := s.DB.QueryRowContext(r.Context(),
						`SELECT role, alarm_scope, branch_tag FROM users WHERE tenant_id = $1 AND user_id::text = $2`,
						tenantID, userID,
					).Scan(&userRole, &alarmScope, &userBranchTag)
					if err != nil && err != sql.ErrNoRows {
						fmt.Printf("[AdminResidents] Failed to get user info: %v\n", err)
					}
				}

				// Check role_permissions for residents resource
				var assignedOnly bool
				if userRole.Valid && userRole.String != "" {
					err := s.DB.QueryRowContext(r.Context(),
						`SELECT assigned_only FROM role_permissions
						 WHERE tenant_id = $1 AND role_code = $2 AND resource_type = 'residents' AND permission_type = 'R'
						 LIMIT 1`,
						SystemTenantID(), userRole.String,
					).Scan(&assignedOnly)
					if err != nil && err != sql.ErrNoRows {
						fmt.Printf("[AdminResidents] Failed to check role permissions: %v\n", err)
						// Default to assigned_only=true for safety if query fails
						assignedOnly = true
					}
				} else {
					// If no role found, default to assigned_only=true for safety
					assignedOnly = true
				}

				args := []any{tenantID}
				q := `SELECT r.resident_id::text, r.tenant_id::text, r.resident_account, r.nickname,
				             r.status, r.service_level, r.admission_date, r.discharge_date,
				             r.family_tag, r.unit_id::text, r.room_id::text, r.bed_id::text,
				             COALESCE(u.unit_name, '') as unit_name,
				             COALESCE(u.branch_tag, '') as branch_tag,
				             COALESCE(u.area_tag, '') as area_tag,
				             COALESCE(u.unit_number, '') as unit_number,
				             COALESCE(u.is_multi_person_room, false) as is_multi_person_room,
				             COALESCE(rm.room_name, '') as room_name,
				             COALESCE(b.bed_name, '') as bed_name
				      FROM residents r
				      LEFT JOIN units u ON u.unit_id = r.unit_id
				      LEFT JOIN rooms rm ON rm.room_id = r.room_id
				      LEFT JOIN beds b ON b.bed_id = r.bed_id`

				// Apply permission-based filtering
				if isResidentLogin {
					// Resident/Family: only show self (or linked residents for family)
					// Note: For family contacts, we should show linked residents, but for now we only show self
					if residentIDForSelf.Valid {
						args = append(args, residentIDForSelf.String)
						q += fmt.Sprintf(` WHERE r.tenant_id = $1 AND r.resident_id::text = $%d`, len(args))
					} else {
						// If resident ID not found, return empty list
						q += ` WHERE 1=0`
					}
				} else if assignedOnly && userID != "" {
					// Staff with assigned_only permission: filter by alarm_scope
					if alarmScope.Valid && alarmScope.String == "BRANCH" && userBranchTag.Valid {
						// Filter by branch_tag: match users.branch_tag with units.branch_tag
						args = append(args, userBranchTag.String)
						q += fmt.Sprintf(` WHERE r.tenant_id = $1 AND u.branch_tag = $%d`, len(args))
					} else if alarmScope.Valid && alarmScope.String == "ASSIGNED_ONLY" {
						// Filter by resident_caregivers.userList
						args = append(args, userID)
						q += fmt.Sprintf(` WHERE r.tenant_id = $1
						                  AND EXISTS (
						                      SELECT 1 FROM resident_caregivers rc
						                      WHERE rc.tenant_id = r.tenant_id
						                        AND rc.resident_id = r.resident_id
						                        AND (rc.userList::text LIKE $%d OR rc.userList::text LIKE $%d)
						                  )`, len(args), len(args)+1)
						// Add pattern matching: exact match or in array
						args = append(args, "%\""+userID+"\"%")
					} else {
						// alarm_scope='ALL' or NULL: show all (but this is unusual for assigned_only roles)
						q += ` WHERE r.tenant_id = $1`
					}
				} else {
					// No assigned_only restriction (Admin/Manager) or no user info
					// Special case: Manager role should filter by branch_tag if set (alarm_scope='BRANCH' by default)
					// For Manager with alarm_scope='BRANCH', filter by matching users.branch_tag with units.branch_tag
					if userRole.Valid && userRole.String == "Manager" && userBranchTag.Valid && userBranchTag.String != "" {
						args = append(args, userBranchTag.String)
						q += fmt.Sprintf(` WHERE r.tenant_id = $1 AND u.branch_tag = $%d`, len(args))
					} else {
						// Show all residents (Admin or Manager without branch_tag)
						q += ` WHERE r.tenant_id = $1`
					}
				}
				argIdx := len(args) + 1
				if search != "" {
					args = append(args, "%"+search+"%")
					q += fmt.Sprintf(` AND (r.nickname ILIKE $%d OR COALESCE(u.unit_name,'') ILIKE $%d)`, argIdx, argIdx)
					argIdx++
				}
				if statusFilter != "" {
					args = append(args, statusFilter)
					q += fmt.Sprintf(` AND r.status = $%d`, argIdx)
					argIdx++
				}
				if serviceLevelFilter != "" {
					args = append(args, serviceLevelFilter)
					q += fmt.Sprintf(` AND r.service_level = $%d`, argIdx)
					argIdx++
				}
				q += ` ORDER BY r.nickname ASC`

				rows, err := s.DB.QueryContext(r.Context(), q, args...)
				if err != nil {
					fmt.Printf("[AdminResidents] Query error: %v\n", err)
					writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to list residents: %v", err)))
					return
				}
				defer rows.Close()
				items := []any{}
				for rows.Next() {
					var residentID, tid, residentAccount, nickname, status, serviceLevel sql.NullString
					var admissionDate, dischargeDate sql.NullTime
					var familyTag, unitID, roomID, bedID sql.NullString
					var unitName, branchTag, areaTag, unitNumber sql.NullString
					var isMultiPersonRoom bool
					var roomName, bedName sql.NullString
					if err := rows.Scan(
						&residentID, &tid, &residentAccount, &nickname,
						&status, &serviceLevel, &admissionDate, &dischargeDate,
						&familyTag, &unitID, &roomID, &bedID,
						&unitName, &branchTag, &areaTag, &unitNumber, &isMultiPersonRoom,
						&roomName, &bedName,
					); err != nil {
						fmt.Printf("[AdminResidents] Scan error: %v\n", err)
						writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to scan resident: %v", err)))
						return
					}
					item := map[string]any{
						"resident_id": residentID.String,
						"tenant_id":   tid.String,
						"status":      status.String,
					}
					if residentAccount.Valid {
						item["resident_account"] = residentAccount.String
					}
					if nickname.Valid {
						item["nickname"] = nickname.String
					}
					if serviceLevel.Valid {
						item["service_level"] = serviceLevel.String
					}
					if admissionDate.Valid {
						item["admission_date"] = admissionDate.Time.Format("2006-01-02")
					}
					if dischargeDate.Valid {
						item["discharge_date"] = dischargeDate.Time.Format("2006-01-02")
					}
					if familyTag.Valid {
						item["family_tag"] = familyTag.String
					}
					if unitID.Valid {
						item["unit_id"] = unitID.String
					}
					if unitName.Valid {
						item["unit_name"] = unitName.String
					}
					if branchTag.Valid {
						item["branch_tag"] = branchTag.String
					}
					if areaTag.Valid {
						item["area_tag"] = areaTag.String
					}
					if unitNumber.Valid {
						item["unit_number"] = unitNumber.String
					}
					item["is_multi_person_room"] = isMultiPersonRoom
					if roomID.Valid {
						item["room_id"] = roomID.String
					}
					if roomName.Valid {
						item["room_name"] = roomName.String
					}
					if bedID.Valid {
						item["bed_id"] = bedID.String
					}
					if bedName.Valid {
						item["bed_name"] = bedName.String
					}
					items = append(items, item)
				}
				writeJSON(w, http.StatusOK, Ok(map[string]any{"items": items, "total": len(items)}))
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

				// Extract required fields
				nickname, _ := payload["nickname"].(string)
				nickname = strings.TrimSpace(nickname)
				if nickname == "" {
					writeJSON(w, http.StatusOK, Fail("nickname is required"))
					return
				}

				// resident_account is required (each institution has its own encoding pattern)
				residentAccount, _ := payload["resident_account"].(string)
				residentAccount = strings.TrimSpace(residentAccount)
				if residentAccount == "" {
					writeJSON(w, http.StatusOK, Fail("resident_account is required (each institution has its own encoding pattern)"))
					return
				}
				// Store as lowercase for consistency (DB constraint ensures lowercase)
				residentAccount = strings.ToLower(residentAccount)

				// Hash account (for login)
				ah, _ := hex.DecodeString(HashAccount(residentAccount))
				if len(ah) == 0 {
					writeJSON(w, http.StatusOK, Fail("failed to hash account"))
					return
				}

				// Generate default password hash (if password is provided, use it; otherwise generate default)
				// Password hash should only depend on password itself (independent of account/phone/email)
				password := "ChangeMe123!"
				if pwd, ok := payload["password"].(string); ok && pwd != "" {
					password = pwd
				}
				aph, _ := hex.DecodeString(HashPassword(password))
				if len(aph) == 0 {
					writeJSON(w, http.StatusOK, Fail("failed to hash password"))
					return
				}

				// Extract optional fields
				status := "active"
				if st, ok := payload["status"].(string); ok && st != "" {
					status = st
				}
				serviceLevel, _ := payload["service_level"].(string)
				var serviceLevelArg any = nil
				if serviceLevel != "" {
					serviceLevelArg = serviceLevel
				}

				// Parse admission_date
				var admissionDate sql.NullTime
				if admDate, ok := payload["admission_date"].(string); ok && admDate != "" {
					if t, err := time.Parse("2006-01-02", admDate); err == nil {
						admissionDate = sql.NullTime{Time: t, Valid: true}
					} else {
						// Try current date as fallback
						admissionDate = sql.NullTime{Time: time.Now(), Valid: true}
					}
				} else {
					// Default to current date
					admissionDate = sql.NullTime{Time: time.Now(), Valid: true}
				}

				unitID, _ := payload["unit_id"].(string)
				var unitIDArg any = nil
				if unitID != "" {
					unitIDArg = unitID
				}

				familyTag, _ := payload["family_tag"].(string)
				var familyTagArg any = nil
				if familyTag != "" {
					familyTagArg = familyTag
				}

				isAccessEnabled := false
				if enabled, ok := payload["is_access_enabled"].(bool); ok {
					isAccessEnabled = enabled
				}

				note, _ := payload["note"].(string)
				var noteArg any = nil
				if note != "" {
					noteArg = note
				}

				// Insert into residents table
				// Note: phone_hash and email_hash will be updated separately if provided
				var residentID string
				err := s.DB.QueryRowContext(
					r.Context(),
					`INSERT INTO residents (tenant_id, resident_account, resident_account_hash, password_hash, nickname, status, service_level, admission_date, unit_id, family_tag, can_view_status, note)
					 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
					 RETURNING resident_id::text`,
					tenantID, residentAccount, ah, aph, nickname, status, serviceLevelArg, admissionDate, unitIDArg, familyTagArg, isAccessEnabled, noteArg,
				).Scan(&residentID)
				if err != nil {
					fmt.Printf("[AdminResidents] Create error: %v\n", err)
					writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to create resident: %v", err)))
					return
				}

				// Always create PHI record (even if empty) to simplify update logic
				// This ensures that UPDATE operations always work without needing INSERT ... ON CONFLICT
				firstName, _ := payload["first_name"].(string)
				lastName, _ := payload["last_name"].(string)
				residentPhone, _ := payload["resident_phone"].(string)
				residentEmail, _ := payload["resident_email"].(string)

				// Build dynamic INSERT for PHI (include provided fields, or create empty record)
				phiCols := []string{"tenant_id", "resident_id"}
				phiVals := []string{"$1", "$2"}
				phiArgs := []any{tenantID, residentID}
				phiArgIdx := 3

				if firstName != "" {
					phiCols = append(phiCols, "first_name")
					phiVals = append(phiVals, fmt.Sprintf("$%d", phiArgIdx))
					phiArgs = append(phiArgs, firstName)
					phiArgIdx++
				}
				if lastName != "" {
					phiCols = append(phiCols, "last_name")
					phiVals = append(phiVals, fmt.Sprintf("$%d", phiArgIdx))
					phiArgs = append(phiArgs, lastName)
					phiArgIdx++
				}
				if residentPhone != "" {
					phiCols = append(phiCols, "resident_phone")
					phiVals = append(phiVals, fmt.Sprintf("$%d", phiArgIdx))
					phiArgs = append(phiArgs, residentPhone)
					phiArgIdx++
				}
				if residentEmail != "" {
					phiCols = append(phiCols, "resident_email")
					phiVals = append(phiVals, fmt.Sprintf("$%d", phiArgIdx))
					phiArgs = append(phiArgs, residentEmail)
					phiArgIdx++
				}

				// Always create PHI record (even if only tenant_id and resident_id)
				phiQuery := fmt.Sprintf(`INSERT INTO resident_phi (%s) VALUES (%s)
				                        ON CONFLICT (tenant_id, resident_id) DO NOTHING`,
					strings.Join(phiCols, ", "), strings.Join(phiVals, ", "))
				_, err = s.DB.ExecContext(r.Context(), phiQuery, phiArgs...)
				if err != nil {
					fmt.Printf("[AdminResidents] Create PHI error: %v\n", err)
					// Don't fail the whole operation, just log
				}

				// Create contacts if provided
				if contacts, ok := payload["contacts"].([]any); ok && len(contacts) > 0 {
					for _, contactRaw := range contacts {
						contact, ok := contactRaw.(map[string]any)
						if !ok {
							continue
						}
						slot, _ := contact["slot"].(string)
						if slot == "" {
							slot = "A" // Default slot
						}
						isEnabled, _ := contact["is_enabled"].(bool)
						relationship, _ := contact["relationship"].(string)
						contactFirstName, _ := contact["contact_first_name"].(string)
						contactLastName, _ := contact["contact_last_name"].(string)
						contactPhone, _ := contact["contact_phone"].(string)
						contactEmail, _ := contact["contact_email"].(string)
						contactFamilyTag, _ := contact["contact_family_tag"].(string)
						receiveSms, _ := contact["receive_sms"].(bool)
						receiveEmail, _ := contact["receive_email"].(bool)

						// Calculate phone_hash and email_hash for login (if phone/email provided)
						var phoneHashArg, emailHashArg any = nil, nil
						var phoneHashBytes, emailHashBytes []byte
						if contactPhone != "" {
							ph, _ := hex.DecodeString(HashAccount(contactPhone))
							if len(ph) > 0 {
								phoneHashArg = ph
								phoneHashBytes = ph
							}
						}
						if contactEmail != "" {
							eh, _ := hex.DecodeString(HashAccount(contactEmail))
							if len(eh) > 0 {
								emailHashArg = eh
								emailHashBytes = eh
							}
						}

						// Check uniqueness before insert
						if err := checkHashUniqueness(s.DB, r, tenantID, "resident_contacts", phoneHashBytes, emailHashBytes, "", ""); err != nil {
							fmt.Printf("[AdminResidents] Create contact uniqueness check error: %v\n", err)
							// Don't fail the whole operation, just log
							continue
						}

						contactQuery := `INSERT INTO resident_contacts 
						                (tenant_id, resident_id, slot, is_enabled, relationship,
						                 contact_first_name, contact_last_name, contact_phone, contact_email,
						                 contact_family_tag, receive_sms, receive_email, phone_hash, email_hash)
						                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
						                ON CONFLICT (tenant_id, resident_id, slot) DO UPDATE SET
						                is_enabled = EXCLUDED.is_enabled,
						                relationship = EXCLUDED.relationship,
						                contact_first_name = EXCLUDED.contact_first_name,
						                contact_last_name = EXCLUDED.contact_last_name,
						                contact_phone = EXCLUDED.contact_phone,
						                contact_email = EXCLUDED.contact_email,
						                contact_family_tag = EXCLUDED.contact_family_tag,
						                receive_sms = EXCLUDED.receive_sms,
						                receive_email = EXCLUDED.receive_email,
						                phone_hash = EXCLUDED.phone_hash,
						                email_hash = EXCLUDED.email_hash`

						var contactFamilyTagArg any = nil
						if contactFamilyTag != "" {
							contactFamilyTagArg = contactFamilyTag
						}
						var relationshipArg any = nil
						if relationship != "" {
							relationshipArg = relationship
						}
						var contactFirstNameArg any = nil
						if contactFirstName != "" {
							contactFirstNameArg = contactFirstName
						}
						var contactLastNameArg any = nil
						if contactLastName != "" {
							contactLastNameArg = contactLastName
						}
						var contactPhoneArg any = nil
						if contactPhone != "" {
							contactPhoneArg = contactPhone
						}
						var contactEmailArg any = nil
						if contactEmail != "" {
							contactEmailArg = contactEmail
						}

						_, err = s.DB.ExecContext(r.Context(), contactQuery,
							tenantID, residentID, slot, isEnabled, relationshipArg,
							contactFirstNameArg, contactLastNameArg, contactPhoneArg, contactEmailArg,
							contactFamilyTagArg, receiveSms, receiveEmail, phoneHashArg, emailHashArg)
						if err != nil {
							// Check for unique constraint violation
							if msg := checkUniqueConstraintError(err, "phone or email"); msg != "" {
								fmt.Printf("[AdminResidents] Create contact uniqueness error: %v\n", msg)
								// Don't fail the whole operation, just log
								continue
							}
							fmt.Printf("[AdminResidents] Create contact error: %v\n", err)
							// Don't fail the whole operation, just log
						}
					}
				}

				writeJSON(w, http.StatusOK, Ok(map[string]any{"resident_id": residentID}))
				return
			}
			writeJSON(w, http.StatusOK, Fail("database not available"))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}
	if strings.HasPrefix(r.URL.Path, "/admin/api/v1/residents/") {
		path := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/residents/")
		// subresources
		if strings.HasSuffix(path, "/reset-password") {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			// Reset resident password
			if s != nil && s.DB != nil {
				tenantID, ok := s.tenantIDFromReq(w, r)
				if !ok {
					return
				}
				residentID := strings.TrimSuffix(path, "/reset-password")
				if residentID == "" || strings.Contains(residentID, "/") {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				var payload map[string]any
				if err := readBodyJSON(r, 1<<20, &payload); err != nil {
					writeJSON(w, http.StatusOK, Fail("invalid body"))
					return
				}
				newPassword, _ := payload["password"].(string)
				newPassword = strings.TrimSpace(newPassword)
				if newPassword == "" {
					writeJSON(w, http.StatusOK, Fail("password is required"))
					return
				}

				// Look up resident_account for hashing
				var residentAccount string
				err := s.DB.QueryRowContext(
					r.Context(),
					`SELECT resident_account
					   FROM residents
					  WHERE tenant_id = $1 AND resident_id::text = $2`,
					tenantID, residentID,
				).Scan(&residentAccount)
				if err != nil {
					if err == sql.ErrNoRows {
						writeJSON(w, http.StatusOK, Fail("resident not found"))
					} else {
						writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to find resident: %v", err)))
					}
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
					`UPDATE residents SET password_hash = $3
					  WHERE tenant_id = $1 AND resident_id::text = $2`,
					tenantID, residentID, aph,
				)
				if err != nil {
					writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to reset password: %v", err)))
					return
				}
				writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
				return
			}
			writeJSON(w, http.StatusOK, Fail("database not available"))
			return
		}
		if strings.HasSuffix(path, "/phi") {
			if r.Method != http.MethodPut {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			// Update PHI
			if s != nil && s.DB != nil {
				tenantID, ok := s.tenantIDFromReq(w, r)
				if !ok {
					return
				}
				residentID := strings.TrimSuffix(path, "/phi")
				if residentID == "" || strings.Contains(residentID, "/") {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				var payload map[string]any
				if err := readBodyJSON(r, 1<<20, &payload); err != nil {
					writeJSON(w, http.StatusOK, Fail("invalid body"))
					return
				}

				// Build dynamic UPDATE query for PHI
				// Support partial updates: if a field is present in payload (even if empty string), update it
				// This allows setting fields to NULL/empty by explicitly passing empty string
				updates := []string{}
				args := []any{tenantID, residentID}
				argIdx := 3

				// Check if field exists in payload (not just if it's non-empty)
				// This allows setting fields to empty/NULL
				if firstName, exists := payload["first_name"]; exists {
					if str, ok := firstName.(string); ok {
						updates = append(updates, fmt.Sprintf("first_name = $%d", argIdx))
						if str != "" {
							args = append(args, str)
						} else {
							args = append(args, nil) // Set to NULL
						}
						argIdx++
					}
				}
				if lastName, exists := payload["last_name"]; exists {
					if str, ok := lastName.(string); ok {
						updates = append(updates, fmt.Sprintf("last_name = $%d", argIdx))
						if str != "" {
							args = append(args, str)
						} else {
							args = append(args, nil) // Set to NULL
						}
						argIdx++
					}
				}
				if residentPhone, exists := payload["resident_phone"]; exists {
					if str, ok := residentPhone.(string); ok {
						updates = append(updates, fmt.Sprintf("resident_phone = $%d", argIdx))
						if str != "" {
							args = append(args, str)
						} else {
							args = append(args, nil) // Set to NULL
						}
						argIdx++
					}
				}
				if residentEmail, exists := payload["resident_email"]; exists {
					if str, ok := residentEmail.(string); ok {
						updates = append(updates, fmt.Sprintf("resident_email = $%d", argIdx))
						if str != "" {
							args = append(args, str)
						} else {
							args = append(args, nil) // Set to NULL
						}
						argIdx++
					}
				}
				// Add more PHI fields as needed (gender, date_of_birth, etc.)

				if len(updates) > 0 {
					// Since PHI record is always created when resident is created,
					// we can use a simple UPDATE query (no need for INSERT ... ON CONFLICT)
					updateQuery := fmt.Sprintf(`UPDATE resident_phi SET %s WHERE tenant_id = $1 AND resident_id = $2`, strings.Join(updates, ", "))
					_, err := s.DB.ExecContext(r.Context(), updateQuery, args...)
					if err != nil {
						writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to update PHI: %v", err)))
						return
					}
				}
				writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
				return
			}
			writeJSON(w, http.StatusOK, Fail("database not available"))
			return
		}
		if strings.HasSuffix(path, "/contacts") {
			if r.Method != http.MethodPut {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			if s != nil && s.DB != nil {
				tenantID, ok := s.tenantIDFromReq(w, r)
				if !ok {
					return
				}
				residentID := strings.TrimSuffix(path, "/contacts")
				if residentID == "" || strings.Contains(residentID, "/") {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				var payload map[string]any
				if err := readBodyJSON(r, 1<<20, &payload); err != nil {
					writeJSON(w, http.StatusOK, Fail("invalid body"))
					return
				}

				// Extract contact fields
				slot, _ := payload["slot"].(string)
				if slot == "" {
					writeJSON(w, http.StatusOK, Fail("slot is required"))
					return
				}
				isEnabled, _ := payload["is_enabled"].(bool)
				relationship, _ := payload["relationship"].(string)
				contactFirstName, _ := payload["contact_first_name"].(string)
				contactLastName, _ := payload["contact_last_name"].(string)
				contactPhone, _ := payload["contact_phone"].(string)
				contactEmail, _ := payload["contact_email"].(string)
				contactFamilyTag, _ := payload["contact_family_tag"].(string)
				receiveSms, _ := payload["receive_sms"].(bool)
				receiveEmail, _ := payload["receive_email"].(bool)
				contactPassword, _ := payload["contact_password"].(string)

				// Prepare arguments (handle NULL values)
				var relationshipArg any = nil
				if relationship != "" {
					relationshipArg = relationship
				}
				var contactFirstNameArg any = nil
				if contactFirstName != "" {
					contactFirstNameArg = contactFirstName
				}
				var contactLastNameArg any = nil
				if contactLastName != "" {
					contactLastNameArg = contactLastName
				}
				var contactPhoneArg any = nil
				if contactPhone != "" {
					contactPhoneArg = contactPhone
				}
				var contactEmailArg any = nil
				if contactEmail != "" {
					contactEmailArg = contactEmail
				}
				var contactFamilyTagArg any = nil
				if contactFamilyTag != "" {
					contactFamilyTagArg = contactFamilyTag
				}

				// Calculate phone_hash and email_hash for login (if phone/email provided)
				var phoneHashArg, emailHashArg any = nil, nil
				var phoneHashBytes, emailHashBytes []byte
				if contactPhone != "" {
					ph, _ := hex.DecodeString(HashAccount(contactPhone))
					if len(ph) > 0 {
						phoneHashArg = ph
						phoneHashBytes = ph
					}
				}
				if contactEmail != "" {
					eh, _ := hex.DecodeString(HashAccount(contactEmail))
					if len(eh) > 0 {
						emailHashArg = eh
						emailHashBytes = eh
					}
				}

				// Check uniqueness before insert/update
				// Get existing contact_id if updating
				var existingContactID string
				err := s.DB.QueryRowContext(r.Context(),
					`SELECT contact_id::text FROM resident_contacts 
					 WHERE tenant_id = $1 AND resident_id::text = $2 AND slot = $3`,
					tenantID, residentID, slot,
				).Scan(&existingContactID)
				if err != nil && err != sql.ErrNoRows {
					writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to check contact: %v", err)))
					return
				}

				// Check hash uniqueness (exclude current contact if updating)
				if err := checkHashUniqueness(s.DB, r, tenantID, "resident_contacts", phoneHashBytes, emailHashBytes, existingContactID, "contact_id"); err != nil {
					writeJSON(w, http.StatusOK, Fail(err.Error()))
					return
				}

				// Build UPDATE query for contact fields
				contactQuery := `INSERT INTO resident_contacts 
				                (tenant_id, resident_id, slot, is_enabled, relationship,
				                 contact_first_name, contact_last_name, contact_phone, contact_email,
				                 contact_family_tag, receive_sms, receive_email, phone_hash, email_hash)
				                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
				                ON CONFLICT (tenant_id, resident_id, slot) DO UPDATE SET
				                is_enabled = EXCLUDED.is_enabled,
				                relationship = EXCLUDED.relationship,
				                contact_first_name = EXCLUDED.contact_first_name,
				                contact_last_name = EXCLUDED.contact_last_name,
				                contact_phone = EXCLUDED.contact_phone,
				                contact_email = EXCLUDED.contact_email,
				                contact_family_tag = EXCLUDED.contact_family_tag,
				                receive_sms = EXCLUDED.receive_sms,
				                receive_email = EXCLUDED.receive_email,
				                phone_hash = EXCLUDED.phone_hash,
				                email_hash = EXCLUDED.email_hash`

				_, err = s.DB.ExecContext(r.Context(), contactQuery,
					tenantID, residentID, slot, isEnabled, relationshipArg,
					contactFirstNameArg, contactLastNameArg, contactPhoneArg, contactEmailArg,
					contactFamilyTagArg, receiveSms, receiveEmail, phoneHashArg, emailHashArg)
				if err != nil {
					// Check for unique constraint violation
					if msg := checkUniqueConstraintError(err, "phone or email"); msg != "" {
						writeJSON(w, http.StatusOK, Fail(msg))
						return
					}
					writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to update contact: %v", err)))
					return
				}

				// Handle password update if provided
				// Password hash should only depend on password itself (independent of account/phone/email)
				if contactPassword != "" {
					// Hash password: sha256(password) - only depends on password
					aph, _ := hex.DecodeString(HashPassword(contactPassword))
					if len(aph) == 0 {
						writeJSON(w, http.StatusOK, Fail("failed to hash password"))
						return
					}

					_, err := s.DB.ExecContext(
						r.Context(),
						`UPDATE resident_contacts SET password_hash = $4
						  WHERE tenant_id = $1 AND resident_id::text = $2 AND slot = $3`,
						tenantID, residentID, slot, aph,
					)
					if err != nil {
						writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to update contact password: %v", err)))
						return
					}
				}

				writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
				return
			}
			writeJSON(w, http.StatusOK, Fail("database not available"))
			return
		}
		// Check for contact password reset: /residents/:id/contacts/:slot/reset-password
		if strings.Contains(path, "/contacts/") && strings.HasSuffix(path, "/reset-password") {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			// Reset contact password
			if s != nil && s.DB != nil {
				tenantID, ok := s.tenantIDFromReq(w, r)
				if !ok {
					return
				}
				// Extract resident_id and slot from path: "resident_id/contacts/slot/reset-password"
				parts := strings.Split(path, "/")
				if len(parts) != 4 || parts[1] != "contacts" {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				residentID := parts[0]
				slot := parts[2]
				if residentID == "" || slot == "" {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				var payload map[string]any
				if err := readBodyJSON(r, 1<<20, &payload); err != nil {
					writeJSON(w, http.StatusOK, Fail("invalid body"))
					return
				}
				newPassword, _ := payload["password"].(string)
				newPassword = strings.TrimSpace(newPassword)
				if newPassword == "" {
					writeJSON(w, http.StatusOK, Fail("password is required"))
					return
				}

				// Hash password: sha256(password) - only depends on password itself
				aph, _ := hex.DecodeString(HashPassword(newPassword))
				if len(aph) == 0 {
					writeJSON(w, http.StatusOK, Fail("failed to hash password"))
					return
				}

				_, err := s.DB.ExecContext(
					r.Context(),
					`UPDATE resident_contacts SET password_hash = $4
					  WHERE tenant_id = $1 AND resident_id::text = $2 AND slot = $3`,
					tenantID, residentID, slot, aph,
				)
				if err != nil {
					writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to reset contact password: %v", err)))
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
		case http.MethodGet:
			// Get resident detail
			if s != nil && s.DB != nil {
				tenantID, ok := s.tenantIDFromReq(w, r)
				if !ok {
					return
				}
				includePHI := r.URL.Query().Get("include_phi") == "true"
				includeContacts := r.URL.Query().Get("include_contacts") == "true"

				var residentID, tid, residentAccount, nickname, status, serviceLevel sql.NullString
				var admissionDate, dischargeDate sql.NullTime
				var familyTag, unitID, roomID, bedID sql.NullString
				var unitName, branchTag, areaTag, unitNumber sql.NullString
				var isMultiPersonRoom bool
				var roomName, bedName sql.NullString
				var note sql.NullString
				var canViewStatus bool

				err := s.DB.QueryRowContext(
					r.Context(),
					`SELECT r.resident_id::text, r.tenant_id::text, r.resident_account, r.nickname,
					        r.status, r.service_level, r.admission_date, r.discharge_date,
					        r.family_tag, r.unit_id::text, r.room_id::text, r.bed_id::text,
					        COALESCE(u.unit_name, '') as unit_name,
					        COALESCE(u.branch_tag, '') as branch_tag,
					        COALESCE(u.area_tag, '') as area_tag,
					        COALESCE(u.unit_number, '') as unit_number,
					        COALESCE(u.is_multi_person_room, false) as is_multi_person_room,
					        COALESCE(rm.room_name, '') as room_name,
					        COALESCE(b.bed_name, '') as bed_name,
					        r.note, r.can_view_status
					 FROM residents r
					 LEFT JOIN units u ON u.unit_id = r.unit_id
					 LEFT JOIN rooms rm ON rm.room_id = r.room_id
					 LEFT JOIN beds b ON b.bed_id = r.bed_id
					 WHERE r.tenant_id = $1 AND r.resident_id::text = $2`,
					tenantID, id,
				).Scan(
					&residentID, &tid, &residentAccount, &nickname,
					&status, &serviceLevel, &admissionDate, &dischargeDate,
					&familyTag, &unitID, &roomID, &bedID,
					&unitName, &branchTag, &areaTag, &unitNumber, &isMultiPersonRoom,
					&roomName, &bedName, &note, &canViewStatus,
				)
				if err != nil {
					if err == sql.ErrNoRows {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to get resident: %v", err)))
					return
				}

				item := map[string]any{
					"resident_id":       residentID.String,
					"tenant_id":         tid.String,
					"status":            status.String,
					"is_access_enabled": canViewStatus,
				}
				if residentAccount.Valid {
					item["resident_account"] = residentAccount.String
				}
				if nickname.Valid {
					item["nickname"] = nickname.String
				}
				if serviceLevel.Valid {
					item["service_level"] = serviceLevel.String
				}
				if admissionDate.Valid {
					item["admission_date"] = admissionDate.Time.Format("2006-01-02")
				}
				if dischargeDate.Valid {
					item["discharge_date"] = dischargeDate.Time.Format("2006-01-02")
				}
				if familyTag.Valid {
					item["family_tag"] = familyTag.String
				}
				if unitID.Valid {
					item["unit_id"] = unitID.String
				}
				if unitName.Valid {
					item["unit_name"] = unitName.String
				}
				if branchTag.Valid {
					item["branch_tag"] = branchTag.String
				}
				if areaTag.Valid {
					item["area_tag"] = areaTag.String
				}
				if unitNumber.Valid {
					item["unit_number"] = unitNumber.String
				}
				item["is_multi_person_room"] = isMultiPersonRoom
				if roomID.Valid {
					item["room_id"] = roomID.String
				}
				if roomName.Valid {
					item["room_name"] = roomName.String
				}
				if bedID.Valid {
					item["bed_id"] = bedID.String
				}
				if bedName.Valid {
					item["bed_name"] = bedName.String
				}
				if note.Valid {
					item["note"] = note.String
				}

				// Load PHI if requested
				if includePHI {
					var phiID, phiFirstName, phiLastName sql.NullString
					err = s.DB.QueryRowContext(
						r.Context(),
						`SELECT phi_id::text, first_name, last_name
						 FROM resident_phi
						 WHERE tenant_id = $1 AND resident_id = $2`,
						tenantID, id,
					).Scan(&phiID, &phiFirstName, &phiLastName)
					if err == nil {
						phi := map[string]any{
							"phi_id":      phiID.String,
							"resident_id": residentID.String,
						}
						if phiFirstName.Valid {
							phi["first_name"] = phiFirstName.String
						}
						if phiLastName.Valid {
							phi["last_name"] = phiLastName.String
						}
						item["phi"] = phi
					}
				}

				// Load contacts if requested
				if includeContacts {
					rows, err := s.DB.QueryContext(
						r.Context(),
						`SELECT contact_id::text, slot, is_enabled, relationship,
						        contact_first_name, contact_last_name, contact_phone, contact_email,
						        contact_family_tag, receive_sms, receive_email
						 FROM resident_contacts
						 WHERE tenant_id = $1 AND resident_id = $2
						 ORDER BY slot ASC`,
						tenantID, id,
					)
					if err == nil {
						defer rows.Close()
						contacts := []any{}
						for rows.Next() {
							var contactID, slot, relationship sql.NullString
							var isEnabled, receiveSMS, receiveEmail bool
							var firstName, lastName, phone, email, familyTag sql.NullString
							if err := rows.Scan(
								&contactID, &slot, &isEnabled, &relationship,
								&firstName, &lastName, &phone, &email, &familyTag,
								&receiveSMS, &receiveEmail,
							); err == nil {
								contact := map[string]any{
									"contact_id":    contactID.String,
									"resident_id":   residentID.String,
									"slot":          slot.String,
									"is_enabled":    isEnabled,
									"receive_sms":   receiveSMS,
									"receive_email": receiveEmail,
								}
								if relationship.Valid {
									contact["relationship"] = relationship.String
								}
								if firstName.Valid {
									contact["contact_first_name"] = firstName.String
								}
								if lastName.Valid {
									contact["contact_last_name"] = lastName.String
								}
								if phone.Valid {
									contact["contact_phone"] = phone.String
								}
								if email.Valid {
									contact["contact_email"] = email.String
								}
								if familyTag.Valid {
									contact["contact_family_tag"] = familyTag.String
								}
								contacts = append(contacts, contact)
							}
						}
						item["contacts"] = contacts
					}
				}

				writeJSON(w, http.StatusOK, Ok(item))
				return
			}
			writeJSON(w, http.StatusOK, Fail("database not available"))
		case http.MethodPut:
			// Update resident
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

				// Build dynamic UPDATE query
				updates := []string{}
				args := []any{tenantID, id}
				argIdx := 3

				if val, ok := payload["nickname"].(string); ok && val != "" {
					updates = append(updates, fmt.Sprintf("nickname = $%d", argIdx))
					args = append(args, val)
					argIdx++
				}
				if val, ok := payload["status"].(string); ok && val != "" {
					updates = append(updates, fmt.Sprintf("status = $%d", argIdx))
					args = append(args, val)
					argIdx++
				}
				if val, ok := payload["service_level"].(string); ok {
					if val != "" {
						updates = append(updates, fmt.Sprintf("service_level = $%d", argIdx))
						args = append(args, val)
					} else {
						updates = append(updates, "service_level = NULL")
					}
					argIdx++
				}
				if val, ok := payload["admission_date"].(string); ok && val != "" {
					if t, err := time.Parse("2006-01-02", val); err == nil {
						updates = append(updates, fmt.Sprintf("admission_date = $%d", argIdx))
						args = append(args, t)
						argIdx++
					}
				}
				if val, ok := payload["discharge_date"].(string); ok {
					if val != "" {
						if t, err := time.Parse("2006-01-02", val); err == nil {
							// Only allow discharge_date if status is discharged or transferred
							currentStatus, _ := payload["status"].(string)
							if currentStatus == "" {
								// If status is not being updated, check current status from DB
								var currentStatusFromDB string
								s.DB.QueryRowContext(r.Context(), `SELECT status FROM residents WHERE tenant_id = $1 AND resident_id::text = $2`, tenantID, id).Scan(&currentStatusFromDB)
								currentStatus = currentStatusFromDB
							}
							if currentStatus == "discharged" || currentStatus == "transferred" {
								updates = append(updates, fmt.Sprintf("discharge_date = $%d", argIdx))
								args = append(args, t)
								argIdx++
							}
						}
					} else {
						// Clear discharge_date if empty string is provided
						updates = append(updates, "discharge_date = NULL")
					}
				}
				if val, ok := payload["unit_id"].(string); ok {
					if val != "" {
						updates = append(updates, fmt.Sprintf("unit_id = $%d", argIdx))
						args = append(args, val)
					} else {
						updates = append(updates, "unit_id = NULL")
					}
					argIdx++
				}
				if val, ok := payload["family_tag"].(string); ok {
					if val != "" {
						updates = append(updates, fmt.Sprintf("family_tag = $%d", argIdx))
						args = append(args, val)
					} else {
						updates = append(updates, "family_tag = NULL")
					}
					argIdx++
				}
				if val, ok := payload["is_access_enabled"].(bool); ok {
					updates = append(updates, fmt.Sprintf("can_view_status = $%d", argIdx))
					args = append(args, val)
					argIdx++
				}
				if val, ok := payload["note"].(string); ok {
					if val != "" {
						updates = append(updates, fmt.Sprintf("note = $%d", argIdx))
						args = append(args, val)
					} else {
						updates = append(updates, "note = NULL")
					}
					argIdx++
				}

				if len(updates) > 0 {
					query := fmt.Sprintf(`UPDATE residents SET %s WHERE tenant_id = $1 AND resident_id::text = $2`, strings.Join(updates, ", "))
					_, err := s.DB.ExecContext(r.Context(), query, args...)
					if err != nil {
						writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to update resident: %v", err)))
						return
					}
				}
				writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
				return
			}
			writeJSON(w, http.StatusOK, Fail("database not available"))
			return
		case http.MethodDelete:
			// Soft delete: mark as discharged
			if s != nil && s.DB != nil {
				tenantID, ok := s.tenantIDFromReq(w, r)
				if !ok {
					return
				}
				_, err := s.DB.ExecContext(
					r.Context(),
					`UPDATE residents SET status = 'discharged' WHERE tenant_id = $1 AND resident_id::text = $2`,
					tenantID, id,
				)
				if err != nil {
					writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to delete resident: %v", err)))
					return
				}
				writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
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
