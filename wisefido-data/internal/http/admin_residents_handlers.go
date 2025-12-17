package httpapi

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (s *StubHandler) AdminResidents(w http.ResponseWriter, r *http.Request) {
	// Handle contact password reset: /admin/api/v1/contacts/:contact_id/reset-password
	if strings.HasPrefix(r.URL.Path, "/admin/api/v1/contacts/") {
		path := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/contacts/")
		if strings.HasSuffix(path, "/reset-password") {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			contactID := strings.TrimSuffix(path, "/reset-password")
			if contactID == "" || strings.Contains(contactID, "/") {
				w.WriteHeader(http.StatusNotFound)
				return
			}
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
				newPassword, _ := payload["password"].(string)
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
					`UPDATE resident_contacts SET password_hash = $3
					  WHERE tenant_id = $1 AND contact_id::text = $2`,
					tenantID, contactID, aph,
				)
				if err != nil {
					if err == sql.ErrNoRows {
						writeJSON(w, http.StatusOK, Fail("contact not found"))
					} else {
						writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to reset contact password: %v", err)))
					}
					return
				}
				writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
				return
			}
			writeJSON(w, http.StatusOK, Fail("database not available"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
		return
	}
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
					// Check if this is a resident_contact login (contact_id) or resident login (resident_id)
					// Try to find if userID exists in resident_contacts table
					if userID != "" && s != nil && s.DB != nil {
						var foundResidentID sql.NullString
						err := s.DB.QueryRowContext(r.Context(),
							`SELECT resident_id::text FROM resident_contacts 
							 WHERE tenant_id = $1 AND contact_id::text = $2`,
							tenantID, userID,
						).Scan(&foundResidentID)
						if err == nil && foundResidentID.Valid {
							// This is a resident_contact login
							residentIDForSelf = foundResidentID
						} else {
							// This is a resident login, X-User-Id is resident_id
							residentIDForSelf = sql.NullString{String: userID, Valid: true}
						}
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
				             COALESCE(b.bed_name, '') as bed_name,
				             r.can_view_status
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
					var canViewStatus bool
					if err := rows.Scan(
						&residentID, &tid, &residentAccount, &nickname,
						&status, &serviceLevel, &admissionDate, &dischargeDate,
						&familyTag, &unitID, &roomID, &bedID,
						&unitName, &branchTag, &areaTag, &unitNumber, &isMultiPersonRoom,
						&roomName, &bedName, &canViewStatus,
					); err != nil {
						fmt.Printf("[AdminResidents] Scan error: %v\n", err)
						writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to scan resident: %v", err)))
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

				// Get phone_hash and email_hash from frontend (calculated on frontend)
				// Note: phone/email are stored in resident_phi table, but phone_hash/email_hash are stored in residents table for login
				var phoneHashArg, emailHashArg any = nil, nil
				var phoneHashBytes, emailHashBytes []byte
				if phoneHashHex, exists := payload["phone_hash"].(string); exists {
					if phoneHashHex != "" {
						ph, _ := hex.DecodeString(phoneHashHex)
						if len(ph) > 0 {
							phoneHashArg = ph
							phoneHashBytes = ph
						}
					} else {
						phoneHashArg = nil // Empty string means null
					}
				}
				if emailHashHex, exists := payload["email_hash"].(string); exists {
					if emailHashHex != "" {
						eh, _ := hex.DecodeString(emailHashHex)
						if len(eh) > 0 {
							emailHashArg = eh
							emailHashBytes = eh
						}
					} else {
						emailHashArg = nil // Empty string means null
					}
				}

				// Check uniqueness of phone_hash and email_hash in residents table (only if hash is provided)
				// Note: email/phone can be empty, so we only check hash uniqueness when hash is provided
				if err := checkHashUniqueness(s.DB, r, tenantID, "residents", phoneHashBytes, emailHashBytes, "", ""); err != nil {
					fmt.Printf("[AdminResidents] Create resident uniqueness check error: %v\n", err)
					writeJSON(w, http.StatusOK, Fail(err.Error()))
					return
				}

				// Insert into residents table
				// phone_hash and email_hash are stored in residents table for login
				var residentID string
				err := s.DB.QueryRowContext(
					r.Context(),
					`INSERT INTO residents (tenant_id, resident_account, resident_account_hash, password_hash, nickname, status, service_level, admission_date, unit_id, family_tag, can_view_status, note, phone_hash, email_hash)
					 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
					 RETURNING resident_id::text`,
					tenantID, residentAccount, ah, aph, nickname, status, serviceLevelArg, admissionDate, unitIDArg, familyTagArg, isAccessEnabled, noteArg,
					phoneHashArg, emailHashArg,
				).Scan(&residentID)
				if err != nil {
					fmt.Printf("[AdminResidents] Create error: %v\n", err)
					writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to create resident: %v", err)))
					return
				}

				// Always create PHI record (even if empty) to simplify update logic
				// This ensures that UPDATE operations always work without needing INSERT ... ON CONFLICT
				// Note: residentPhone and residentEmail were already read above for hash calculation
				firstName, _ := payload["first_name"].(string)
				firstName = strings.TrimSpace(firstName)
				// first_name is required when creating resident
				if firstName == "" {
					writeJSON(w, http.StatusOK, Fail("first_name is required"))
					return
				}
				lastName, _ := payload["last_name"].(string)
				lastName = strings.TrimSpace(lastName)

				// Build dynamic INSERT for PHI (include provided fields, or create empty record)
				// first_name is always included (required)
				phiCols := []string{"tenant_id", "resident_id", "first_name"}
				phiVals := []string{"$1", "$2", "$3"}
				phiArgs := []any{tenantID, residentID, firstName}
				phiArgIdx := 4

				if lastName != "" {
					phiCols = append(phiCols, "last_name")
					phiVals = append(phiVals, fmt.Sprintf("$%d", phiArgIdx))
					phiArgs = append(phiArgs, lastName)
					phiArgIdx++
				}
				if lastName != "" {
					phiCols = append(phiCols, "last_name")
					phiVals = append(phiVals, fmt.Sprintf("$%d", phiArgIdx))
					phiArgs = append(phiArgs, lastName)
					phiArgIdx++
				}
				// Only save phone/email plaintext to resident_phi if save_phone/save_email flags are true
				// Note: phone_hash/email_hash are already saved to residents table above
				savePhone, _ := payload["save_phone"].(bool)
				saveEmail, _ := payload["save_email"].(bool)
				residentPhone, _ := payload["resident_phone"]
				residentEmail, _ := payload["resident_email"]
				if savePhone {
					phiCols = append(phiCols, "resident_phone")
					phiVals = append(phiVals, fmt.Sprintf("$%d", phiArgIdx))
					if str, ok := residentPhone.(string); ok && str != "" {
						phiArgs = append(phiArgs, str)
					} else {
						phiArgs = append(phiArgs, nil) // null if not provided or empty
					}
					phiArgIdx++
				}
				if saveEmail {
					phiCols = append(phiCols, "resident_email")
					phiVals = append(phiVals, fmt.Sprintf("$%d", phiArgIdx))
					if str, ok := residentEmail.(string); ok && str != "" {
						phiArgs = append(phiArgs, str)
					} else {
						phiArgs = append(phiArgs, nil) // null if not provided or empty
					}
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
						// Note: email/phone can be empty, so we only check hash uniqueness when hash is provided
						if err := checkHashUniqueness(s.DB, r, tenantID, "resident_contacts", phoneHashBytes, emailHashBytes, "", ""); err != nil {
							fmt.Printf("[AdminResidents] Create contact uniqueness check error: %v\n", err)
							writeJSON(w, http.StatusOK, Fail(err.Error()))
							return
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

				// Check permissions: resident can only update self, resident_contact can only update linked resident
				userID := r.Header.Get("X-User-Id")
				userType := r.Header.Get("X-User-Type")
				if (userType == "resident" || userType == "family") && userID != "" {
					// Check if this is a resident_contact login
					var foundResidentID sql.NullString
					err := s.DB.QueryRowContext(r.Context(),
						`SELECT resident_id::text FROM resident_contacts 
						 WHERE tenant_id = $1 AND contact_id::text = $2`,
						tenantID, userID,
					).Scan(&foundResidentID)
					if err == nil && foundResidentID.Valid {
						// This is a resident_contact login - can only update linked resident
						if foundResidentID.String != residentID {
							writeJSON(w, http.StatusOK, Fail("access denied: can only update linked resident"))
							return
						}
					} else {
						// This is a resident login - can only update self
						if userID != residentID {
							writeJSON(w, http.StatusOK, Fail("access denied: can only update own information"))
							return
						}
					}
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
				// Track if phone/email are being updated to also update residents table hash
				// Frontend sends phone_hash/email_hash (calculated on frontend) and resident_phone/resident_email
				// If resident_phone/resident_email is provided (even if null), update it
				// If not provided, don't update (keep existing value)
				var phoneUpdated, emailUpdated bool
				var phoneHashArg, emailHashArg any = nil, nil

				// Get phone_hash from frontend (calculated on frontend)
				// If phone_hash is provided (even if null), it means phone is being updated
				if phoneHashHex, exists := payload["phone_hash"]; exists {
					phoneUpdated = true
					if str, ok := phoneHashHex.(string); ok {
						if str != "" {
							ph, _ := hex.DecodeString(str)
							if len(ph) > 0 {
								phoneHashArg = ph
							}
						} else {
							phoneHashArg = nil // Empty string means null
						}
					} else if phoneHashHex == nil {
						phoneHashArg = nil // null means null (delete)
					}
				}
				// Update resident_phone if provided in payload (even if null, means delete)
				if residentPhone, exists := payload["resident_phone"]; exists {
					updates = append(updates, fmt.Sprintf("resident_phone = $%d", argIdx))
					if str, ok := residentPhone.(string); ok && str != "" {
						args = append(args, str)
					} else {
						args = append(args, nil) // null or empty means delete/clear
					}
					argIdx++
				}

				// Get email_hash from frontend (calculated on frontend)
				// If email_hash is provided (even if null), it means email is being updated
				if emailHashHex, exists := payload["email_hash"]; exists {
					emailUpdated = true
					if str, ok := emailHashHex.(string); ok {
						if str != "" {
							eh, _ := hex.DecodeString(str)
							if len(eh) > 0 {
								emailHashArg = eh
							}
						} else {
							emailHashArg = nil // Empty string means null
						}
					} else if emailHashHex == nil {
						emailHashArg = nil // null means null (delete)
					}
				}
				// Update resident_email if provided in payload (even if null, means delete)
				if residentEmail, exists := payload["resident_email"]; exists {
					updates = append(updates, fmt.Sprintf("resident_email = $%d", argIdx))
					if str, ok := residentEmail.(string); ok && str != "" {
						args = append(args, str)
					} else {
						args = append(args, nil) // null or empty means delete/clear
					}
					argIdx++
				}
				// Add more PHI fields (gender, date_of_birth, biometric, functional, chronic conditions, etc.)
				if gender, exists := payload["gender"]; exists {
					if str, ok := gender.(string); ok {
						updates = append(updates, fmt.Sprintf("gender = $%d", argIdx))
						if str != "" {
							args = append(args, str)
						} else {
							args = append(args, nil) // Set to NULL
						}
						argIdx++
					}
				}
				if dateOfBirth, exists := payload["date_of_birth"]; exists {
					if str, ok := dateOfBirth.(string); ok {
						updates = append(updates, fmt.Sprintf("date_of_birth = $%d", argIdx))
						if str != "" {
							if t, err := time.Parse("2006-01-02", str); err == nil {
								args = append(args, t)
							} else {
								args = append(args, nil) // Invalid date, set to NULL
							}
						} else {
							args = append(args, nil) // Set to NULL
						}
						argIdx++
					}
				}
				// Biometric PHI
				if weightLb, exists := payload["weight_lb"]; exists {
					updates = append(updates, fmt.Sprintf("weight_lb = $%d", argIdx))
					if num, ok := weightLb.(float64); ok && num > 0 {
						args = append(args, num)
					} else {
						args = append(args, nil) // Set to NULL
					}
					argIdx++
				}
				if heightFt, exists := payload["height_ft"]; exists {
					updates = append(updates, fmt.Sprintf("height_ft = $%d", argIdx))
					if num, ok := heightFt.(float64); ok && num > 0 {
						args = append(args, num)
					} else {
						args = append(args, nil) // Set to NULL
					}
					argIdx++
				}
				if heightIn, exists := payload["height_in"]; exists {
					updates = append(updates, fmt.Sprintf("height_in = $%d", argIdx))
					if num, ok := heightIn.(float64); ok && num > 0 {
						args = append(args, num)
					} else {
						args = append(args, nil) // Set to NULL
					}
					argIdx++
				}
				// Functional Mobility
				if mobilityLevel, exists := payload["mobility_level"]; exists {
					updates = append(updates, fmt.Sprintf("mobility_level = $%d", argIdx))
					if num, ok := mobilityLevel.(float64); ok {
						args = append(args, int(num))
					} else if num, ok := mobilityLevel.(int); ok {
						args = append(args, num)
					} else {
						args = append(args, nil) // Set to NULL
					}
					argIdx++
				}
				// Functional Health
				if tremorStatus, exists := payload["tremor_status"]; exists {
					if str, ok := tremorStatus.(string); ok {
						updates = append(updates, fmt.Sprintf("tremor_status = $%d", argIdx))
						if str != "" {
							args = append(args, str)
						} else {
							args = append(args, nil) // Set to NULL
						}
						argIdx++
					}
				}
				if mobilityAid, exists := payload["mobility_aid"]; exists {
					if str, ok := mobilityAid.(string); ok {
						updates = append(updates, fmt.Sprintf("mobility_aid = $%d", argIdx))
						if str != "" {
							args = append(args, str)
						} else {
							args = append(args, nil) // Set to NULL
						}
						argIdx++
					}
				}
				if adlAssistance, exists := payload["adl_assistance"]; exists {
					if str, ok := adlAssistance.(string); ok {
						updates = append(updates, fmt.Sprintf("adl_assistance = $%d", argIdx))
						if str != "" {
							args = append(args, str)
						} else {
							args = append(args, nil) // Set to NULL
						}
						argIdx++
					}
				}
				if commStatus, exists := payload["comm_status"]; exists {
					if str, ok := commStatus.(string); ok {
						updates = append(updates, fmt.Sprintf("comm_status = $%d", argIdx))
						if str != "" {
							args = append(args, str)
						} else {
							args = append(args, nil) // Set to NULL
						}
						argIdx++
					}
				}
				// Chronic Conditions
				if hasHypertension, exists := payload["has_hypertension"]; exists {
					updates = append(updates, fmt.Sprintf("has_hypertension = $%d", argIdx))
					if b, ok := hasHypertension.(bool); ok {
						args = append(args, b)
					} else {
						args = append(args, nil) // Set to NULL
					}
					argIdx++
				}
				if hasHyperlipaemia, exists := payload["has_hyperlipaemia"]; exists {
					updates = append(updates, fmt.Sprintf("has_hyperlipaemia = $%d", argIdx))
					if b, ok := hasHyperlipaemia.(bool); ok {
						args = append(args, b)
					} else {
						args = append(args, nil) // Set to NULL
					}
					argIdx++
				}
				if hasHyperglycaemia, exists := payload["has_hyperglycaemia"]; exists {
					updates = append(updates, fmt.Sprintf("has_hyperglycaemia = $%d", argIdx))
					if b, ok := hasHyperglycaemia.(bool); ok {
						args = append(args, b)
					} else {
						args = append(args, nil) // Set to NULL
					}
					argIdx++
				}
				if hasStrokeHistory, exists := payload["has_stroke_history"]; exists {
					updates = append(updates, fmt.Sprintf("has_stroke_history = $%d", argIdx))
					if b, ok := hasStrokeHistory.(bool); ok {
						args = append(args, b)
					} else {
						args = append(args, nil) // Set to NULL
					}
					argIdx++
				}
				if hasParalysis, exists := payload["has_paralysis"]; exists {
					updates = append(updates, fmt.Sprintf("has_paralysis = $%d", argIdx))
					if b, ok := hasParalysis.(bool); ok {
						args = append(args, b)
					} else {
						args = append(args, nil) // Set to NULL
					}
					argIdx++
				}
				if hasAlzheimer, exists := payload["has_alzheimer"]; exists {
					updates = append(updates, fmt.Sprintf("has_alzheimer = $%d", argIdx))
					if b, ok := hasAlzheimer.(bool); ok {
						args = append(args, b)
					} else {
						args = append(args, nil) // Set to NULL
					}
					argIdx++
				}
				if medicalHistory, exists := payload["medical_history"]; exists {
					if str, ok := medicalHistory.(string); ok {
						updates = append(updates, fmt.Sprintf("medical_history = $%d", argIdx))
						if str != "" {
							args = append(args, str)
						} else {
							args = append(args, nil) // Set to NULL
						}
						argIdx++
					}
				}
				// HIS Integration
				if hisResidentName, exists := payload["HIS_resident_name"]; exists {
					if str, ok := hisResidentName.(string); ok {
						updates = append(updates, fmt.Sprintf("his_resident_name = $%d", argIdx))
						if str != "" {
							args = append(args, str)
						} else {
							args = append(args, nil) // Set to NULL
						}
						argIdx++
					}
				}
				if hisResidentAdmissionDate, exists := payload["HIS_resident_admission_date"]; exists {
					if str, ok := hisResidentAdmissionDate.(string); ok {
						updates = append(updates, fmt.Sprintf("his_resident_admission_date = $%d", argIdx))
						if str != "" {
							if t, err := time.Parse("2006-01-02", str); err == nil {
								args = append(args, t)
							} else {
								args = append(args, nil) // Invalid date, set to NULL
							}
						} else {
							args = append(args, nil) // Set to NULL
						}
						argIdx++
					}
				}
				if hisResidentDischargeDate, exists := payload["HIS_resident_discharge_date"]; exists {
					if str, ok := hisResidentDischargeDate.(string); ok {
						updates = append(updates, fmt.Sprintf("his_resident_discharge_date = $%d", argIdx))
						if str != "" {
							if t, err := time.Parse("2006-01-02", str); err == nil {
								args = append(args, t)
							} else {
								args = append(args, nil) // Invalid date, set to NULL
							}
						} else {
							args = append(args, nil) // Set to NULL
						}
						argIdx++
					}
				}
				// Home Address
				if homeAddressStreet, exists := payload["home_address_street"]; exists {
					if str, ok := homeAddressStreet.(string); ok {
						updates = append(updates, fmt.Sprintf("home_address_street = $%d", argIdx))
						if str != "" {
							args = append(args, str)
						} else {
							args = append(args, nil) // Set to NULL
						}
						argIdx++
					}
				}
				if homeAddressCity, exists := payload["home_address_city"]; exists {
					if str, ok := homeAddressCity.(string); ok {
						updates = append(updates, fmt.Sprintf("home_address_city = $%d", argIdx))
						if str != "" {
							args = append(args, str)
						} else {
							args = append(args, nil) // Set to NULL
						}
						argIdx++
					}
				}
				if homeAddressState, exists := payload["home_address_state"]; exists {
					if str, ok := homeAddressState.(string); ok {
						updates = append(updates, fmt.Sprintf("home_address_state = $%d", argIdx))
						if str != "" {
							args = append(args, str)
						} else {
							args = append(args, nil) // Set to NULL
						}
						argIdx++
					}
				}
				if homeAddressPostalCode, exists := payload["home_address_postal_code"]; exists {
					if str, ok := homeAddressPostalCode.(string); ok {
						updates = append(updates, fmt.Sprintf("home_address_postal_code = $%d", argIdx))
						if str != "" {
							args = append(args, str)
						} else {
							args = append(args, nil) // Set to NULL
						}
						argIdx++
					}
				}
				if plusCode, exists := payload["plus_code"]; exists {
					if str, ok := plusCode.(string); ok {
						updates = append(updates, fmt.Sprintf("plus_code = $%d", argIdx))
						if str != "" {
							args = append(args, str)
						} else {
							args = append(args, nil) // Set to NULL
						}
						argIdx++
					}
				}

				if len(updates) > 0 {
					// Use INSERT ... ON CONFLICT DO UPDATE to ensure record exists
					// This handles both create and update cases (if record doesn't exist, create it)
					// Build INSERT columns and values from updates
					// Note: args array structure: args[0]=tenantID, args[1]=residentID, args[2+]=update values in order
					phiCols := []string{"tenant_id", "resident_id"}
					phiVals := []string{"$1", "$2"}
					phiArgs := []any{tenantID, residentID}
					phiArgIdx := 3

					// Extract column names and values from updates in order
					// The args array already contains values in the same order as updates
					// args[0]=tenantID, args[1]=residentID, args[2]=first update value, args[3]=second update value, etc.
					for i, update := range updates {
						// Parse "column = $N" format
						parts := strings.Split(update, " = $")
						if len(parts) == 2 {
							colName := parts[0]
							phiCols = append(phiCols, colName)
							phiVals = append(phiVals, fmt.Sprintf("$%d", phiArgIdx))
							// Get the corresponding value from args (skip tenantID and residentID, so use i+2)
							// args[0]=tenantID, args[1]=residentID, args[2+]=update values in same order as updates
							if i+2 < len(args) {
								phiArgs = append(phiArgs, args[i+2])
							} else {
								// Fallback: try to get value by parsing the original argIdx
								argIdx, _ := strconv.Atoi(parts[1])
								if argIdx >= 3 && argIdx-3 < len(args) {
									phiArgs = append(phiArgs, args[argIdx-3])
								} else {
									phiArgs = append(phiArgs, nil) // Safety fallback
								}
							}
							phiArgIdx++
						}
					}

					// Build conflict updates using EXCLUDED (references the values being inserted)
					// This avoids duplicating parameter values
					conflictUpdates := []string{}
					for _, update := range updates {
						parts := strings.Split(update, " = $")
						if len(parts) == 2 {
							colName := parts[0]
							conflictUpdates = append(conflictUpdates, fmt.Sprintf("%s = EXCLUDED.%s", colName, colName))
						}
					}

					updateQuery := fmt.Sprintf(`INSERT INTO resident_phi (%s) VALUES (%s)
					                        ON CONFLICT (tenant_id, resident_id) DO UPDATE SET %s`,
						strings.Join(phiCols, ", "), strings.Join(phiVals, ", "), strings.Join(conflictUpdates, ", "))
					_, err := s.DB.ExecContext(r.Context(), updateQuery, phiArgs...)
					if err != nil {
						writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to update PHI: %v", err)))
						return
					}

					// Also update phone_hash and email_hash in residents table if phone/email were updated
					// Note: phone/email are stored in resident_phi table, but phone_hash/email_hash are stored in residents table for login
					if phoneUpdated || emailUpdated {
						// Check uniqueness of phone_hash and email_hash in residents table (only if hash is provided)
						// Note: email/phone can be empty, so we only check hash uniqueness when hash is provided
						// Exclude current resident from uniqueness check
						var phoneHashBytes, emailHashBytes []byte
						if phoneHashArg != nil {
							if ph, ok := phoneHashArg.([]byte); ok {
								phoneHashBytes = ph
							}
						}
						if emailHashArg != nil {
							if eh, ok := emailHashArg.([]byte); ok {
								emailHashBytes = eh
							}
						}
						if err := checkHashUniqueness(s.DB, r, tenantID, "residents", phoneHashBytes, emailHashBytes, residentID, "resident_id"); err != nil {
							writeJSON(w, http.StatusOK, Fail(err.Error()))
							return
						}

						residentUpdates := []string{}
						residentArgs := []any{tenantID, residentID}
						residentArgIdx := 3
						if phoneUpdated {
							residentUpdates = append(residentUpdates, fmt.Sprintf("phone_hash = $%d", residentArgIdx))
							residentArgs = append(residentArgs, phoneHashArg)
							residentArgIdx++
						}
						if emailUpdated {
							residentUpdates = append(residentUpdates, fmt.Sprintf("email_hash = $%d", residentArgIdx))
							residentArgs = append(residentArgs, emailHashArg)
							residentArgIdx++
						}
						if len(residentUpdates) > 0 {
							residentUpdateQuery := fmt.Sprintf(`UPDATE residents SET %s WHERE tenant_id = $1 AND resident_id::text = $2`, strings.Join(residentUpdates, ", "))
							_, err = s.DB.ExecContext(r.Context(), residentUpdateQuery, residentArgs...)
							if err != nil {
								writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to update resident hash: %v", err)))
								return
							}
						}
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

				// Check permissions: resident_contact can only modify own slot
				userID := r.Header.Get("X-User-Id")
				userType := r.Header.Get("X-User-Type")
				if (userType == "resident" || userType == "family") && userID != "" {
					// Check if this is a resident_contact login
					var foundResidentID sql.NullString
					var foundSlot sql.NullString
					err := s.DB.QueryRowContext(r.Context(),
						`SELECT resident_id::text, slot FROM resident_contacts 
						 WHERE tenant_id = $1 AND contact_id::text = $2`,
						tenantID, userID,
					).Scan(&foundResidentID, &foundSlot)
					if err == nil && foundResidentID.Valid {
						// This is a resident_contact login - can only modify own slot
						if foundResidentID.String != residentID {
							writeJSON(w, http.StatusOK, Fail("access denied: can only modify contacts for linked resident"))
							return
						}
						// Verify that the slot matches the contact's own slot
						if foundSlot.Valid && foundSlot.String != slot {
							writeJSON(w, http.StatusOK, Fail("access denied: can only modify own slot"))
							return
						}
					} else {
						// This is a resident login - can modify contacts for self
						if userID != residentID {
							writeJSON(w, http.StatusOK, Fail("access denied: can only modify contacts for self"))
							return
						}
					}
				}
				isEnabled, _ := payload["is_enabled"].(bool)
				relationship, _ := payload["relationship"].(string)
				contactFirstName, _ := payload["contact_first_name"].(string)
				contactLastName, _ := payload["contact_last_name"].(string)
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
				// Get phone_hash and email_hash from frontend (calculated on frontend)
				var phoneHashArg, emailHashArg any = nil, nil
				var phoneHashBytes, emailHashBytes []byte
				if phoneHashHex, exists := payload["phone_hash"]; exists {
					if str, ok := phoneHashHex.(string); ok {
						if str != "" {
							ph, _ := hex.DecodeString(str)
							if len(ph) > 0 {
								phoneHashArg = ph
								phoneHashBytes = ph
							}
						} else {
							phoneHashArg = nil // Empty string means null
						}
					} else if phoneHashHex == nil {
						phoneHashArg = nil // null means null
					}
				}
				if emailHashHex, exists := payload["email_hash"]; exists {
					if str, ok := emailHashHex.(string); ok {
						if str != "" {
							eh, _ := hex.DecodeString(str)
							if len(eh) > 0 {
								emailHashArg = eh
								emailHashBytes = eh
							}
						} else {
							emailHashArg = nil // Empty string means null
						}
					} else if emailHashHex == nil {
						emailHashArg = nil // null means null
					}
				}

				// Handle phone/email plaintext based on save flags (frontend sends null if not saving)
				// If contact_phone/contact_email is provided (even if null), update it
				var contactPhoneArg any = nil
				if contactPhoneVal, exists := payload["contact_phone"]; exists {
					if contactPhoneVal == nil {
						contactPhoneArg = nil // Explicitly null: delete phone
					} else if str, ok := contactPhoneVal.(string); ok && str != "" {
						contactPhoneArg = str
					} else {
						contactPhoneArg = nil // Empty string means null
					}
				}
				var contactEmailArg any = nil
				if contactEmailVal, exists := payload["contact_email"]; exists {
					if contactEmailVal == nil {
						contactEmailArg = nil // Explicitly null: delete email
					} else if str, ok := contactEmailVal.(string); ok && str != "" {
						contactEmailArg = str
					} else {
						contactEmailArg = nil // Empty string means null
					}
				}
				var contactFamilyTagArg any = nil
				if contactFamilyTag != "" {
					contactFamilyTagArg = contactFamilyTag
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

				// Handle password_hash if contact_password is provided (like create user, password is included in INSERT)
				// Password hash should only depend on password itself (independent of account/phone/email)
				var passwordHashArg any = nil
				hasPassword := false
				if contactPassword != "" {
					// Hash password: sha256(password) - only depends on password
					aph, _ := hex.DecodeString(HashPassword(contactPassword))
					if len(aph) == 0 {
						writeJSON(w, http.StatusOK, Fail("failed to hash password"))
						return
					}
					passwordHashArg = aph
					hasPassword = true
				}

				// Build UPDATE query for contact fields
				// Include password_hash in INSERT ... ON CONFLICT DO UPDATE only if password is provided (like create user)
				var contactQuery string
				var contactArgs []any
				if hasPassword {
					// Include password_hash in INSERT and UPDATE
					contactQuery = `INSERT INTO resident_contacts 
					                (tenant_id, resident_id, slot, is_enabled, relationship,
					                 contact_first_name, contact_last_name, contact_phone, contact_email,
					                 contact_family_tag, receive_sms, receive_email, phone_hash, email_hash, password_hash)
					                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
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
					                email_hash = EXCLUDED.email_hash,
					                password_hash = EXCLUDED.password_hash`
					contactArgs = []any{tenantID, residentID, slot, isEnabled, relationshipArg,
						contactFirstNameArg, contactLastNameArg, contactPhoneArg, contactEmailArg,
						contactFamilyTagArg, receiveSms, receiveEmail, phoneHashArg, emailHashArg, passwordHashArg}
				} else {
					// Don't update password_hash if password is not provided (keep existing value)
					contactQuery = `INSERT INTO resident_contacts 
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
					contactArgs = []any{tenantID, residentID, slot, isEnabled, relationshipArg,
						contactFirstNameArg, contactLastNameArg, contactPhoneArg, contactEmailArg,
						contactFamilyTagArg, receiveSms, receiveEmail, phoneHashArg, emailHashArg}
				}

				_, err = s.DB.ExecContext(r.Context(), contactQuery, contactArgs...)
				if err != nil {
					// Check for unique constraint violation
					if msg := checkUniqueConstraintError(err, "phone or email"); msg != "" {
						writeJSON(w, http.StatusOK, Fail(msg))
						return
					}
					writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to update contact: %v", err)))
					return
				}

				writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
				return
			}
			writeJSON(w, http.StatusOK, Fail("database not available"))
			return
		}
		// Check for contact password reset: /residents/:id/contacts/:slot/reset-password
		// OR /contacts/:contact_id/reset-password (simpler, uses contact_id directly)
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
				// Extract from path: could be "resident_id/contacts/slot/reset-password" or "contacts/contact_id/reset-password"
				parts := strings.Split(path, "/")
				var contactID string
				var residentID, slot string

				if len(parts) == 3 && parts[0] == "contacts" {
					// New format: /contacts/:contact_id/reset-password
					contactID = parts[1]
				} else if len(parts) == 4 && parts[1] == "contacts" {
					// Old format: /residents/:id/contacts/:slot/reset-password (for backward compatibility)
					residentID = parts[0]
					slot = parts[2]
				} else {
					w.WriteHeader(http.StatusNotFound)
					return
				}

				var payload map[string]any
				if err := readBodyJSON(r, 1<<20, &payload); err != nil {
					writeJSON(w, http.StatusOK, Fail("invalid body"))
					return
				}
				newPassword, _ := payload["password"].(string)
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

				var err error
				if contactID != "" {
					// Use contact_id directly (simpler, each slot is independent)
					_, err = s.DB.ExecContext(
						r.Context(),
						`UPDATE resident_contacts SET password_hash = $3
						  WHERE tenant_id = $1 AND contact_id::text = $2`,
						tenantID, contactID, aph,
					)
				} else if residentID != "" && slot != "" {
					// Use resident_id + slot (backward compatibility)
					_, err = s.DB.ExecContext(
						r.Context(),
						`UPDATE resident_contacts SET password_hash = $4
					  WHERE tenant_id = $1 AND resident_id::text = $2 AND slot = $3`,
						tenantID, residentID, slot, aph,
					)
				} else {
					writeJSON(w, http.StatusOK, Fail("invalid path: contact_id or resident_id+slot required"))
					return
				}

				if err != nil {
					if err == sql.ErrNoRows {
						writeJSON(w, http.StatusOK, Fail("contact not found"))
					} else {
						writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to reset contact password: %v", err)))
					}
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

				// Support both resident_id and contact_id in the path parameter
				// If id is a contact_id, find the associated resident_id first
				var actualResidentID string = id
				var foundContactID sql.NullString
				err := s.DB.QueryRowContext(r.Context(),
					`SELECT contact_id::text FROM resident_contacts 
					 WHERE tenant_id = $1 AND contact_id::text = $2`,
					tenantID, id,
				).Scan(&foundContactID)
				if err == nil && foundContactID.Valid {
					// id is a contact_id, find the associated resident_id
					var linkedResidentID sql.NullString
					err2 := s.DB.QueryRowContext(r.Context(),
						`SELECT resident_id::text FROM resident_contacts 
						 WHERE tenant_id = $1 AND contact_id::text = $2`,
						tenantID, id,
					).Scan(&linkedResidentID)
					if err2 == nil && linkedResidentID.Valid {
						actualResidentID = linkedResidentID.String
					} else {
						writeJSON(w, http.StatusOK, Fail("contact not found or not linked to any resident"))
						return
					}
				}
				// If id is not a contact_id, assume it's a resident_id (existing behavior)

				// Check permissions: resident can only view self, resident_contact can only view linked resident
				userID := r.Header.Get("X-User-Id")
				userType := r.Header.Get("X-User-Type")
				if (userType == "resident" || userType == "family") && userID != "" {
					// Check if this is a resident_contact login
					var foundResidentID sql.NullString
					err := s.DB.QueryRowContext(r.Context(),
						`SELECT resident_id::text FROM resident_contacts 
						 WHERE tenant_id = $1 AND contact_id::text = $2`,
						tenantID, userID,
					).Scan(&foundResidentID)
					if err == nil && foundResidentID.Valid {
						// This is a resident_contact login - can only view linked resident
						if foundResidentID.String != actualResidentID {
							writeJSON(w, http.StatusOK, Fail("access denied: can only view linked resident"))
							return
						}
					} else {
						// This is a resident login - can only view self
						if userID != actualResidentID {
							writeJSON(w, http.StatusOK, Fail("access denied: can only view own information"))
							return
						}
					}
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

				var residentPhoneHash, residentEmailHash []byte
				err = s.DB.QueryRowContext(
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
					        r.note, r.can_view_status,
					        r.phone_hash, r.email_hash
					 FROM residents r
					 LEFT JOIN units u ON u.unit_id = r.unit_id
					 LEFT JOIN rooms rm ON rm.room_id = r.room_id
					 LEFT JOIN beds b ON b.bed_id = r.bed_id
					 WHERE r.tenant_id = $1 AND r.resident_id::text = $2`,
					tenantID, actualResidentID,
				).Scan(
					&residentID, &tid, &residentAccount, &nickname,
					&status, &serviceLevel, &admissionDate, &dischargeDate,
					&familyTag, &unitID, &roomID, &bedID,
					&unitName, &branchTag, &areaTag, &unitNumber, &isMultiPersonRoom,
					&roomName, &bedName, &note, &canViewStatus,
					&residentPhoneHash, &residentEmailHash,
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
					var phiID, phiFirstName, phiLastName, gender, residentPhone, residentEmail, tremorStatus, mobilityAid, adlAssistance, commStatus, medicalHistory, hisResidentName, homeAddressStreet, homeAddressCity, homeAddressState, homeAddressPostalCode, plusCode sql.NullString
					var dateOfBirth, hisAdmissionDate, hisDischargeDate sql.NullTime
					var weightLb, heightFt, heightIn sql.NullFloat64
					var mobilityLevel sql.NullInt64
					var hasHypertension, hasHyperlipaemia, hasHyperglycaemia, hasStrokeHistory, hasParalysis, hasAlzheimer sql.NullBool
					err = s.DB.QueryRowContext(
						r.Context(),
						`SELECT phi_id::text, first_name, last_name, gender, date_of_birth,
						        resident_phone, resident_email, weight_lb, height_ft, height_in,
						        mobility_level, tremor_status, mobility_aid, adl_assistance, comm_status,
						        has_hypertension, has_hyperlipaemia, has_hyperglycaemia, has_stroke_history, has_paralysis, has_alzheimer,
						        medical_history, HIS_resident_name, HIS_resident_admission_date, HIS_resident_discharge_date,
						        home_address_street, home_address_city, home_address_state, home_address_postal_code, plus_code
						 FROM resident_phi
						 WHERE tenant_id = $1 AND resident_id = $2`,
						tenantID, actualResidentID,
					).Scan(&phiID, &phiFirstName, &phiLastName, &gender, &dateOfBirth,
						&residentPhone, &residentEmail, &weightLb, &heightFt, &heightIn,
						&mobilityLevel, &tremorStatus, &mobilityAid, &adlAssistance, &commStatus,
						&hasHypertension, &hasHyperlipaemia, &hasHyperglycaemia, &hasStrokeHistory, &hasParalysis, &hasAlzheimer,
						&medicalHistory, &hisResidentName, &hisAdmissionDate, &hisDischargeDate,
						&homeAddressStreet, &homeAddressCity, &homeAddressState, &homeAddressPostalCode, &plusCode)
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
						if gender.Valid {
							phi["gender"] = gender.String
						}
						if dateOfBirth.Valid {
							phi["date_of_birth"] = dateOfBirth.Time.Format("2006-01-02")
						}
						// If phone_hash exists but phone is NULL, return placeholder
						if residentPhoneHash != nil && len(residentPhoneHash) > 0 {
							if residentPhone.Valid {
								phi["resident_phone"] = residentPhone.String
							} else {
								phi["resident_phone"] = "xxx-xxx-xxxx" // Placeholder when hash exists but phone is not saved
							}
						} else if residentPhone.Valid {
							phi["resident_phone"] = residentPhone.String
						}
						// If email_hash exists but email is NULL, return placeholder
						if residentEmailHash != nil && len(residentEmailHash) > 0 {
							if residentEmail.Valid {
								phi["resident_email"] = residentEmail.String
							} else {
								phi["resident_email"] = "***@***" // Placeholder when hash exists but email is not saved
							}
						} else if residentEmail.Valid {
							phi["resident_email"] = residentEmail.String
						}
						if weightLb.Valid {
							phi["weight_lb"] = weightLb.Float64
						}
						if heightFt.Valid {
							phi["height_ft"] = heightFt.Float64
						}
						if heightIn.Valid {
							phi["height_in"] = heightIn.Float64
						}
						if mobilityLevel.Valid {
							phi["mobility_level"] = mobilityLevel.Int64
						}
						if tremorStatus.Valid {
							phi["tremor_status"] = tremorStatus.String
						}
						if mobilityAid.Valid {
							phi["mobility_aid"] = mobilityAid.String
						}
						if adlAssistance.Valid {
							phi["adl_assistance"] = adlAssistance.String
						}
						if commStatus.Valid {
							phi["comm_status"] = commStatus.String
						}
						if hasHypertension.Valid {
							phi["has_hypertension"] = hasHypertension.Bool
						}
						if hasHyperlipaemia.Valid {
							phi["has_hyperlipaemia"] = hasHyperlipaemia.Bool
						}
						if hasHyperglycaemia.Valid {
							phi["has_hyperglycaemia"] = hasHyperglycaemia.Bool
						}
						if hasStrokeHistory.Valid {
							phi["has_stroke_history"] = hasStrokeHistory.Bool
						}
						if hasParalysis.Valid {
							phi["has_paralysis"] = hasParalysis.Bool
						}
						if hasAlzheimer.Valid {
							phi["has_alzheimer"] = hasAlzheimer.Bool
						}
						if medicalHistory.Valid {
							phi["medical_history"] = medicalHistory.String
						}
						if hisResidentName.Valid {
							phi["HIS_resident_name"] = hisResidentName.String
						}
						if hisAdmissionDate.Valid {
							phi["HIS_resident_admission_date"] = hisAdmissionDate.Time.Format("2006-01-02")
						}
						if hisDischargeDate.Valid {
							phi["HIS_resident_discharge_date"] = hisDischargeDate.Time.Format("2006-01-02")
						}
						if homeAddressStreet.Valid {
							phi["home_address_street"] = homeAddressStreet.String
						}
						if homeAddressCity.Valid {
							phi["home_address_city"] = homeAddressCity.String
						}
						if homeAddressState.Valid {
							phi["home_address_state"] = homeAddressState.String
						}
						if homeAddressPostalCode.Valid {
							phi["home_address_postal_code"] = homeAddressPostalCode.String
						}
						if plusCode.Valid {
							phi["plus_code"] = plusCode.String
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
						        contact_family_tag, receive_sms, receive_email,
						        phone_hash, email_hash
						 FROM resident_contacts
						 WHERE tenant_id = $1 AND resident_id = $2
						 ORDER BY slot ASC`,
						tenantID, actualResidentID,
					)
					if err == nil {
						defer rows.Close()
						contacts := []any{}
						for rows.Next() {
							var contactID, slot, relationship sql.NullString
							var isEnabled, receiveSMS, receiveEmail bool
							var firstName, lastName, phone, email, familyTag sql.NullString
							var phoneHash, emailHash []byte
							if err := rows.Scan(
								&contactID, &slot, &isEnabled, &relationship,
								&firstName, &lastName, &phone, &email, &familyTag,
								&receiveSMS, &receiveEmail, &phoneHash, &emailHash,
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
								// If phone_hash exists but phone is NULL, return placeholder
								if phoneHash != nil && len(phoneHash) > 0 {
									if phone.Valid {
										contact["contact_phone"] = phone.String
									} else {
										contact["contact_phone"] = "xxx-xxx-xxxx" // Placeholder when hash exists but phone is not saved
									}
								} else if phone.Valid {
									contact["contact_phone"] = phone.String
								}
								// If email_hash exists but email is NULL, return placeholder
								if emailHash != nil && len(emailHash) > 0 {
									if email.Valid {
										contact["contact_email"] = email.String
									} else {
										contact["contact_email"] = "***@***" // Placeholder when hash exists but email is not saved
									}
								} else if email.Valid {
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

				// Check permissions: resident can only update self, resident_contact can only update linked resident
				userID := r.Header.Get("X-User-Id")
				userType := r.Header.Get("X-User-Type")
				if (userType == "resident" || userType == "family") && userID != "" {
					// Check if this is a resident_contact login
					var foundResidentID sql.NullString
					err := s.DB.QueryRowContext(r.Context(),
						`SELECT resident_id::text FROM resident_contacts 
						 WHERE tenant_id = $1 AND contact_id::text = $2`,
						tenantID, userID,
					).Scan(&foundResidentID)
					if err == nil && foundResidentID.Valid {
						// This is a resident_contact login - can only update linked resident
						if foundResidentID.String != id {
							writeJSON(w, http.StatusOK, Fail("access denied: can only update linked resident"))
							return
						}
					} else {
						// This is a resident login - can only update self
						if userID != id {
							writeJSON(w, http.StatusOK, Fail("access denied: can only update own information"))
							return
						}
					}
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
