package httpapi

import (
	"database/sql"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

func (s *StubHandler) Auth(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/auth/api/v1/login":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// 对齐 owlFront LoginResult（authModel.ts）
		// loginApi 会把 tenant_id/userType 等放在 JSON body（axios beforeRequestHook 会把 params 作为 data）
		var req map[string]any
		_ = readBodyJSON(r, 1<<20, &req)
		// Some clients may wrap params in {params:{...}}
		if p, ok := req["params"].(map[string]any); ok && p != nil {
			if _, ok2 := req["tenant_id"]; !ok2 {
				req["tenant_id"] = p["tenant_id"]
			}
			if _, ok2 := req["userType"]; !ok2 {
				req["userType"] = p["userType"]
			}
			if _, ok2 := req["accountHash"]; !ok2 {
				req["accountHash"] = p["accountHash"]
			}
			if _, ok2 := req["passwordHash"]; !ok2 {
				req["passwordHash"] = p["passwordHash"]
			}
		}

		tenantID, _ := req["tenant_id"].(string)
		if tenantID == "" {
			tenantID = r.URL.Query().Get("tenant_id")
		}
		userType, _ := req["userType"].(string)
		if userType == "" {
			userType = r.URL.Query().Get("userType")
		}
		if userType == "" {
			userType = "staff"
		}
		accountHash, _ := req["accountHash"].(string)
		if accountHash == "" {
			accountHash = r.URL.Query().Get("accountHash")
		}
		// passwordHash is independent: password_hash = SHA256(password)
		passwordHash, _ := req["passwordHash"].(string)
		if passwordHash == "" {
			passwordHash = r.URL.Query().Get("passwordHash")
		}

		accountHash = strings.TrimSpace(accountHash)
		passwordHash = strings.TrimSpace(passwordHash)
		if accountHash == "" {
			if s != nil && s.Logger != nil {
				s.Logger.Warn("User login failed: missing credentials",
					zap.String("ip_address", getClientIP(r)),
					zap.String("user_agent", r.UserAgent()),
					zap.String("reason", "missing_credentials"),
					zap.String("missing_field", "accountHash"),
				)
			}
			writeJSON(w, http.StatusOK, Fail("missing credentials"))
			return
		}
		// passwordHash must be provided (account_hash and password_hash are independent)
		if passwordHash == "" {
			if s != nil && s.Logger != nil {
				s.Logger.Warn("User login failed: missing credentials",
					zap.String("ip_address", getClientIP(r)),
					zap.String("user_agent", r.UserAgent()),
					zap.String("reason", "missing_credentials"),
					zap.String("missing_field", "passwordHash"),
				)
			}
			writeJSON(w, http.StatusOK, Fail("missing credentials"))
			return
		}

		normalizedUserType := strings.ToLower(strings.TrimSpace(userType))
		if normalizedUserType == "" {
			normalizedUserType = "staff"
		}

		// Prefer DB auth when available (AuthStore is in-memory and will be lost after restart).
		// Decode hashes once (DB stores BYTEA).
		// account_hash and password_hash are independent: account_hash = SHA256(account), password_hash = SHA256(password)
		var ah, ph []byte
		if s != nil && s.DB != nil {
			var err1, err2 error
			ah, err1 = hex.DecodeString(accountHash)
			ph, err2 = hex.DecodeString(passwordHash)
			if err1 != nil || len(ah) == 0 {
				if s != nil && s.Logger != nil {
					s.Logger.Warn("User login failed: invalid account hash format",
						zap.String("ip_address", getClientIP(r)),
						zap.String("user_agent", r.UserAgent()),
						zap.String("reason", "invalid_account_hash"),
						zap.Error(err1),
					)
				}
				writeJSON(w, http.StatusOK, Fail("invalid credentials"))
				return
			}
			if err2 != nil || len(ph) == 0 {
				if s != nil && s.Logger != nil {
					s.Logger.Warn("User login failed: invalid password hash format",
						zap.String("ip_address", getClientIP(r)),
						zap.String("user_agent", r.UserAgent()),
						zap.String("reason", "invalid_password_hash"),
						zap.Error(err2),
					)
				}
				writeJSON(w, http.StatusOK, Fail("invalid credentials"))
				return
			}
		}

		// If tenant_id is not provided, resolve it from DB by (accountHash, passwordHash, userType).
		// account_hash and password_hash are independent: account_hash = SHA256(account), password_hash = SHA256(password)
		// owlFront behavior:
		// - 0 match: invalid credentials
		// - 1 match: auto-login into that tenant (tenant_id optional)
		// - >1 match: frontend must let user choose an institution
		if tenantID == "" && s != nil && s.DB != nil {
			var rows *sql.Rows
			var err error
			switch normalizedUserType {
			case "resident":
				// When userType="resident", search logic:
				// Step 1: Query resident_contacts table by password_hash first, then filter by accountHash
				//   - Find records matching password_hash (passwordHash)
				//   - Then filter by email_hash (priority 1) OR phone_hash (priority 2)
				//   - Must be active (is_enabled = true)
				// Step 2: If no match, query residents table by password_hash, then filter by accountHash
				//   - Find records matching password_hash (passwordHash)
				//   - Then filter by email_hash (priority 1) OR phone_hash (priority 2) OR resident_account_hash (priority 3)
				//   - Must be active (status = 'active')
				// Priority: email_hash > phone_hash > resident_account_hash

				// Step 1: Query resident_contacts table
				// Priority: email_hash > phone_hash
				rows, err = s.DB.QueryContext(
					r.Context(),
					`SELECT DISTINCT rc.tenant_id::text,
					        CASE
					          WHEN rc.email_hash = $1 THEN 'email'
					          WHEN rc.phone_hash = $1 THEN 'phone'
					          ELSE 'phone'
					        END as account_type,
					        CASE
					          WHEN rc.email_hash = $1 THEN 1
					          WHEN rc.phone_hash = $1 THEN 2
					          ELSE 3
					        END as priority
					   FROM resident_contacts rc
					  WHERE rc.password_hash = $2
					    AND COALESCE(rc.is_enabled,true) = true
					    AND (rc.email_hash = $1 OR rc.phone_hash = $1)
					  ORDER BY priority ASC, rc.tenant_id::text ASC`,
					ah, ph,
				)
				if err == nil {
					var count int
					for rows.Next() {
						count++
					}
					rows.Close()
					if count == 0 {
						// Step 2: No match in resident_contacts table, try residents table
						// Priority: email_hash > phone_hash > resident_account_hash
						rows, err = s.DB.QueryContext(
							r.Context(),
							`SELECT DISTINCT r.tenant_id::text,
							        CASE
							          WHEN r.email_hash = $1 THEN 'email'
							          WHEN r.phone_hash = $1 THEN 'phone'
							          WHEN r.resident_account_hash = $1 THEN 'account'
							          ELSE 'account'
							        END as account_type,
							        CASE
							          WHEN r.email_hash = $1 THEN 1
							          WHEN r.phone_hash = $1 THEN 2
							          WHEN r.resident_account_hash = $1 THEN 3
							          ELSE 4
							        END as priority
							   FROM residents r
							  WHERE r.password_hash = $2
							    AND COALESCE(r.status,'active') = 'active'
							    AND (r.email_hash = $1 OR r.phone_hash = $1 OR r.resident_account_hash = $1)
							  ORDER BY priority ASC, r.tenant_id::text ASC`,
							ah, ph,
						)
					} else {
						// Re-query to get all matching tenant_ids with account_type
						rows, err = s.DB.QueryContext(
							r.Context(),
							`SELECT DISTINCT rc.tenant_id::text,
							        CASE
							          WHEN rc.email_hash = $1 THEN 'email'
							          WHEN rc.phone_hash = $1 THEN 'phone'
							          ELSE 'phone'
							        END as account_type,
							        CASE
							          WHEN rc.email_hash = $1 THEN 1
							          WHEN rc.phone_hash = $1 THEN 2
							          ELSE 3
							        END as priority
							   FROM resident_contacts rc
							  WHERE rc.password_hash = $2
							    AND COALESCE(rc.is_enabled,true) = true
							    AND (rc.email_hash = $1 OR rc.phone_hash = $1)
							  ORDER BY priority ASC, rc.tenant_id::text ASC`,
							ah, ph,
						)
					}
				}
			default: // staff
				// When userType="staff", search logic:
				// Step 1: Query users table by password_hash first, then filter by accountHash
				//   - Find records matching password_hash (passwordHash)
				//   - Then filter by email_hash (priority 1) OR phone_hash (priority 2) OR user_account_hash (priority 3)
				// Priority: email_hash > phone_hash > user_account_hash
				rows, err = s.DB.QueryContext(
					r.Context(),
					`SELECT DISTINCT u.tenant_id::text,
					        CASE
					          WHEN u.email_hash = $1 THEN 'email'
					          WHEN u.phone_hash = $1 THEN 'phone'
					          WHEN u.user_account_hash = $1 THEN 'account'
					          ELSE 'account'
					        END as account_type,
					        CASE
					          WHEN u.email_hash = $1 THEN 1
					          WHEN u.phone_hash = $1 THEN 2
					          WHEN u.user_account_hash = $1 THEN 3
					          ELSE 4
					        END as priority
					   FROM users u
					  WHERE u.password_hash = $2
					    AND COALESCE(u.status,'active') = 'active'
					    AND (u.email_hash = $1 OR u.phone_hash = $1 OR u.user_account_hash = $1)
					  ORDER BY priority ASC, u.tenant_id::text ASC`,
					ah, ph,
				)
			}
			if err != nil {
				// Log detailed error for debugging
				writeJSON(w, http.StatusOK, Fail("failed to resolve tenant: "+err.Error()))
				return
			}
			defer rows.Close()
			var tids []string
			for rows.Next() {
				var tid string
				var accountType string // For staff and resident, we get account_type but don't need it here
				var priority int       // Priority value for ordering (we don't need it after ordering)
				if normalizedUserType == "staff" || normalizedUserType == "resident" {
					if err := rows.Scan(&tid, &accountType, &priority); err == nil && tid != "" {
						tids = append(tids, tid)
					}
				} else {
					if err := rows.Scan(&tid); err == nil && tid != "" {
						tids = append(tids, tid)
					}
				}
			}
			if len(tids) == 0 {
				if s != nil && s.Logger != nil {
					s.Logger.Warn("User login failed: invalid credentials",
						zap.String("ip_address", getClientIP(r)),
						zap.String("user_agent", r.UserAgent()),
						zap.String("user_type", normalizedUserType),
						zap.String("reason", "invalid_credentials"),
						zap.String("note", "no matching tenant found"),
					)
				}
				writeJSON(w, http.StatusOK, Fail("invalid credentials"))
				return
			}
			if len(tids) > 1 {
				// IMPORTANT: keep message aligned with owlFront expectations.
				writeJSON(w, http.StatusOK, Fail("Multiple institutions found, please select one"))
				return
			}
			tenantID = tids[0]
		}

		role := "Manager"
		userID := "stub-user"
		userAccount := "stub"
		nickName := ""
		tenantName := "Stub Tenant"
		domain := ""
		var branchTag sql.NullString
		var residentIDForContact, contactSlot string // For resident_contact password reset

		if s != nil && s.DB != nil {
			if tenantID == "" {
				writeJSON(w, http.StatusOK, Fail("tenant_id is required"))
				return
			}
			var status string
			switch normalizedUserType {
			case "resident":
				// When userType="resident", login verification logic:
				// Step 1: Try resident_contacts table first
				//   - Priority: email_hash > phone_hash
				//   - Must be active (is_enabled = true)
				// Step 2: If no match, try residents table
				//   - Priority: email_hash > phone_hash > resident_account_hash
				//   - Must be active (status = 'active')
				// account_hash and password_hash are independent: account_hash = SHA256(account), password_hash = SHA256(password)

				// Step 1: Try family contact login first
				var enabled bool
				var first, last string
				var familyBranchTag sql.NullString
				var accountType string
				err := s.DB.QueryRowContext(
					r.Context(),
					`SELECT rc.contact_id::text,
					        rc.resident_id::text,
					        rc.slot,
					        COALESCE(rc.contact_first_name,''),
					        COALESCE(rc.contact_last_name,''),
					        rc.role,
					        COALESCE(rc.is_enabled,true),
					        COALESCE(t.tenant_name,''),
					        COALESCE(t.domain,''),
					        COALESCE(u.branch_tag, '') as branch_tag,
					        CASE
					          WHEN rc.email_hash = $2 THEN 'email'
					          WHEN rc.phone_hash = $2 THEN 'phone'
					          ELSE 'phone'
					        END as account_type
					   FROM resident_contacts rc
					   JOIN tenants t ON t.tenant_id = rc.tenant_id
					   JOIN residents r ON r.resident_id = rc.resident_id AND r.tenant_id = rc.tenant_id
					   LEFT JOIN units u ON u.unit_id = r.unit_id
					  WHERE rc.tenant_id = $1
					    AND rc.password_hash = $3
					    AND (rc.email_hash = $2 OR rc.phone_hash = $2)
					  ORDER BY 
					    CASE
					      WHEN rc.email_hash = $2 THEN 1
					      WHEN rc.phone_hash = $2 THEN 2
					      ELSE 3
					    END ASC
					  LIMIT 1`,
					tenantID, ah, ph,
				).Scan(&userID, &residentIDForContact, &contactSlot, &first, &last, &role, &enabled, &tenantName, &domain, &familyBranchTag, &accountType)
				if err != nil {
					// Step 2: Try resident login
					var residentBranchTag sql.NullString
					err2 := s.DB.QueryRowContext(
						r.Context(),
						`SELECT r.resident_id::text,
						        r.resident_account,
						        COALESCE(r.nickname,''),
						        r.role,
						        COALESCE(r.status,'active'),
						        COALESCE(t.tenant_name,''),
						        COALESCE(t.domain,''),
						        COALESCE(u.branch_tag, '') as branch_tag,
						        CASE
						          WHEN r.email_hash = $2 THEN 'email'
						          WHEN r.phone_hash = $2 THEN 'phone'
						          WHEN r.resident_account_hash = $2 THEN 'account'
						          ELSE 'account'
						        END as account_type
						   FROM residents r
						   JOIN tenants t ON t.tenant_id = r.tenant_id
						   LEFT JOIN units u ON u.unit_id = r.unit_id
						  WHERE r.tenant_id = $1
						    AND r.password_hash = $3
						    AND (r.email_hash = $2 OR r.phone_hash = $2 OR r.resident_account_hash = $2)
						  ORDER BY 
						    CASE
						      WHEN r.email_hash = $2 THEN 1
						      WHEN r.phone_hash = $2 THEN 2
						      WHEN r.resident_account_hash = $2 THEN 3
						      ELSE 4
						    END ASC
						  LIMIT 1`,
						tenantID, ah, ph,
					).Scan(&userID, &userAccount, &nickName, &role, &status, &tenantName, &domain, &residentBranchTag, &accountType)
					if err2 == nil {
						branchTag = residentBranchTag
						if status != "active" {
							if s != nil && s.Logger != nil {
								s.Logger.Warn("User login failed: account not active",
									zap.String("user_id", userID),
									zap.String("tenant_id", tenantID),
									zap.String("user_type", normalizedUserType),
									zap.String("status", status),
									zap.String("ip_address", getClientIP(r)),
									zap.String("reason", "account_not_active"),
								)
							}
							writeJSON(w, http.StatusOK, Fail("user is not active"))
							return
						}
					} else {
						if s != nil && s.Logger != nil {
							s.Logger.Warn("User login failed: invalid credentials",
								zap.String("tenant_id", tenantID),
								zap.String("user_type", normalizedUserType),
								zap.String("ip_address", getClientIP(r)),
								zap.String("user_agent", r.UserAgent()),
								zap.String("reason", "invalid_credentials"),
								zap.String("note", "resident login failed"),
							)
						}
						writeJSON(w, http.StatusOK, Fail("invalid credentials"))
						return
					}
				} else {
					// Family contact login succeeded
					branchTag = familyBranchTag
					if !enabled {
						if s != nil && s.Logger != nil {
							s.Logger.Warn("User login failed: account not active",
								zap.String("user_id", userID),
								zap.String("tenant_id", tenantID),
								zap.String("user_type", normalizedUserType),
								zap.String("reason", "account_not_active"),
								zap.String("note", "family contact not enabled"),
								zap.String("ip_address", getClientIP(r)),
							)
						}
						writeJSON(w, http.StatusOK, Fail("user is not active"))
						return
					}
					// For family contacts, expose a stable identifier as user_account.
					userAccount = userID
					if strings.TrimSpace(first+" "+last) != "" {
						nickName = strings.TrimSpace(first + " " + last)
					} else {
						nickName = role
					}
					status = "active"
				}
			default: // staff
				// account_hash and password_hash are independent: account_hash = SHA256(account), password_hash = SHA256(password)
				// Priority: email_hash > phone_hash > user_account_hash
				var staffBranchTag sql.NullString
				var accountType string
				err := s.DB.QueryRowContext(
					r.Context(),
					`SELECT u.user_id::text,
					        u.user_account,
					        COALESCE(u.nickname,''),
					        u.role,
					        COALESCE(u.status,'active'),
					        COALESCE(t.tenant_name,''),
					        COALESCE(t.domain,''),
					        COALESCE(u.branch_tag, '') as branch_tag,
					        CASE
					          WHEN u.email_hash = $2 THEN 'email'
					          WHEN u.phone_hash = $2 THEN 'phone'
					          WHEN u.user_account_hash = $2 THEN 'account'
					          ELSE 'account'
					        END as account_type
					   FROM users u
					   JOIN tenants t ON t.tenant_id = u.tenant_id
					  WHERE u.tenant_id = $1
					    AND u.password_hash = $3
					    AND (u.email_hash = $2 OR u.phone_hash = $2 OR u.user_account_hash = $2)
					  ORDER BY 
					    CASE
					      WHEN u.email_hash = $2 THEN 1
					      WHEN u.phone_hash = $2 THEN 2
					      WHEN u.user_account_hash = $2 THEN 3
					      ELSE 4
					    END ASC
					  LIMIT 1`,
					tenantID, ah, ph,
				).Scan(&userID, &userAccount, &nickName, &role, &status, &tenantName, &domain, &staffBranchTag, &accountType)
				if err == nil {
					branchTag = staffBranchTag
				}
				if err != nil {
					if s != nil && s.Logger != nil {
						s.Logger.Warn("User login failed: invalid credentials",
							zap.String("tenant_id", tenantID),
							zap.String("user_type", normalizedUserType),
							zap.String("ip_address", getClientIP(r)),
							zap.String("user_agent", r.UserAgent()),
							zap.String("reason", "invalid_credentials"),
							zap.String("note", "staff login failed"),
						)
					}
					writeJSON(w, http.StatusOK, Fail("invalid credentials"))
					return
				}
				if status != "active" {
					if s != nil && s.Logger != nil {
						s.Logger.Warn("User login failed: account not active",
							zap.String("user_id", userID),
							zap.String("user_account", userAccount),
							zap.String("tenant_id", tenantID),
							zap.String("user_type", normalizedUserType),
							zap.String("status", status),
							zap.String("ip_address", getClientIP(r)),
							zap.String("reason", "account_not_active"),
						)
					}
					writeJSON(w, http.StatusOK, Fail("user is not active"))
					return
				}
			}
		} else if s != nil && s.AuthStore != nil && allowAuthStoreFallback() {
			// Fallback: in-memory auth (not used in production, only for stub/dev mode)
			// Note: AuthStore uses accountPasswordHash, but we only have passwordHash now
			// For now, skip AuthStore fallback as it's not compatible with the new passwordHash approach
			// AuthStore fallback is disabled - return invalid credentials
			writeJSON(w, http.StatusOK, Fail("invalid credentials"))
			return
		}
		if (s == nil || s.DB == nil) && !allowAuthStoreFallback() {
			if s != nil && s.Logger != nil {
				s.Logger.Warn("User login failed: database not configured",
					zap.String("ip_address", getClientIP(r)),
					zap.String("user_agent", r.UserAgent()),
					zap.String("reason", "db_not_configured"),
				)
			}
			writeJSON(w, http.StatusOK, Fail("db auth not configured"))
			return
		}

		if nickName == "" {
			// Prefer nickname; fall back to role/userAccount for display
			if role != "" {
				nickName = role
			} else {
				nickName = userAccount
			}
		}

		// Update last_login_at for staff users (in users table)
		if normalizedUserType == "staff" && s != nil && s.DB != nil {
			_, _ = s.DB.ExecContext(r.Context(),
				"UPDATE users SET last_login_at = $1 WHERE user_id = $2",
				time.Now(), userID,
			)
		}

		// Log successful login
		if s != nil && s.Logger != nil {
			s.Logger.Info("User login successful",
				zap.String("user_id", userID),
				zap.String("user_account", userAccount),
				zap.String("user_type", normalizedUserType),
				zap.String("tenant_id", tenantID),
				zap.String("tenant_name", tenantName),
				zap.String("role", role),
				zap.String("ip_address", getClientIP(r)),
				zap.String("user_agent", r.UserAgent()),
				zap.Time("login_time", time.Now()),
			)
		}

		result := map[string]any{
			"accessToken":  "stub-access-token",
			"refreshToken": "stub-refresh-token",
			"userId":       userID,
			"user_account": userAccount,
			"userType":     normalizedUserType,
			"role":         role,
			"nickName":     nickName,
			"tenant_id":    tenantID,
			"tenant_name":  tenantName,
			"domain":       domain,
			"homePath":     "/monitoring/overview",
		}
		// Add branchTag if available (from users.branch_tag for staff, or units.branch_tag for resident)
		if branchTag.Valid && branchTag.String != "" {
			result["branchTag"] = branchTag.String
		}
		// Note: For resident_contact login, userId is already contact_id, no need for additional fields
		writeJSON(w, http.StatusOK, Ok(result))
		return
	case "/auth/api/v1/institutions/search":
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// 返回 Institution[]（authModel.ts: {id,name,domain?}）
		// account_hash and password_hash are independent: account_hash = SHA256(account), password_hash = SHA256(password)
		accountHash := r.URL.Query().Get("accountHash")
		passwordHash := r.URL.Query().Get("passwordHash")
		userType := strings.TrimSpace(r.URL.Query().Get("userType"))
		normalizedUserType := strings.ToLower(strings.TrimSpace(userType))
		if normalizedUserType == "" {
			normalizedUserType = "staff"
		}
		accountHash = strings.TrimSpace(accountHash)
		passwordHash = strings.TrimSpace(passwordHash)

		// Prefer DB lookup when available.
		// passwordHash must be provided (account_hash and password_hash are independent)
		if s != nil && s.DB != nil && accountHash != "" && passwordHash != "" {
			ah, err1 := hex.DecodeString(accountHash)
			if err1 != nil || len(ah) == 0 {
				writeJSON(w, http.StatusOK, Ok([]any{}))
				return
			}
			ph, err2 := hex.DecodeString(passwordHash)
			if err2 != nil || len(ph) == 0 {
				writeJSON(w, http.StatusOK, Ok([]any{}))
				return
			}
			var rows *sql.Rows
			var err error
			switch normalizedUserType {
			case "resident":
				// When userType="resident", search logic:
				// Step 1: Query resident_contacts table by password_hash first, then filter by accountHash
				//   - Find records matching password_hash (passwordHash)
				//   - Then filter by email_hash (priority 1) OR phone_hash (priority 2)
				//   - Must be active (is_enabled = true)
				// Step 2: If no match, query residents table by password_hash, then filter by accountHash
				//   - Find records matching password_hash (passwordHash)
				//   - Then filter by email_hash (priority 1) OR phone_hash (priority 2) OR resident_account_hash (priority 3)
				//   - Must be active (status = 'active')
				// Priority: email_hash > phone_hash > resident_account_hash

				// Step 1: Query resident_contacts table
				// Security: Only return tenant_id and account_type for matched institutions (already verified by password)
				// Priority: email_hash > phone_hash
				rows, err = s.DB.QueryContext(
					r.Context(),
					`SELECT DISTINCT rc.tenant_id::text,
					        CASE
					          WHEN rc.email_hash = $1 THEN 'email'
					          WHEN rc.phone_hash = $1 THEN 'phone'
					          ELSE 'phone'
					        END as account_type,
					        CASE
					          WHEN rc.email_hash = $1 THEN 1
					          WHEN rc.phone_hash = $1 THEN 2
					          ELSE 3
					        END as priority
					   FROM resident_contacts rc
					  WHERE rc.password_hash = $2
					    AND COALESCE(rc.is_enabled,true) = true
					    AND (rc.email_hash = $1 OR rc.phone_hash = $1)
					  ORDER BY priority ASC, rc.tenant_id::text ASC`,
					ah, ph,
				)
				if err == nil {
					var count int
					for rows.Next() {
						count++
					}
					rows.Close()
					if count == 0 {
						// Step 2: No match in resident_contacts table, try residents table
						// Priority: email_hash > phone_hash > resident_account_hash
						rows, err = s.DB.QueryContext(
							r.Context(),
							`SELECT DISTINCT r.tenant_id::text,
							        CASE
							          WHEN r.email_hash = $1 THEN 'email'
							          WHEN r.phone_hash = $1 THEN 'phone'
							          WHEN r.resident_account_hash = $1 THEN 'account'
							          ELSE 'account'
							        END as account_type,
							        CASE
							          WHEN r.email_hash = $1 THEN 1
							          WHEN r.phone_hash = $1 THEN 2
							          WHEN r.resident_account_hash = $1 THEN 3
							          ELSE 4
							        END as priority
							   FROM residents r
							  WHERE r.password_hash = $2
							    AND COALESCE(r.status,'active') = 'active'
							    AND (r.email_hash = $1 OR r.phone_hash = $1 OR r.resident_account_hash = $1)
							  ORDER BY priority ASC, r.tenant_id::text ASC`,
							ah, ph,
						)
					} else {
						// Re-query to get all matching tenant_ids with account_type
						rows, err = s.DB.QueryContext(
							r.Context(),
							`SELECT DISTINCT rc.tenant_id::text,
							        CASE
							          WHEN rc.email_hash = $1 THEN 'email'
							          WHEN rc.phone_hash = $1 THEN 'phone'
							          ELSE 'phone'
							        END as account_type,
							        CASE
							          WHEN rc.email_hash = $1 THEN 1
							          WHEN rc.phone_hash = $1 THEN 2
							          ELSE 3
							        END as priority
							   FROM resident_contacts rc
							  WHERE rc.password_hash = $2
							    AND COALESCE(rc.is_enabled,true) = true
							    AND (rc.email_hash = $1 OR rc.phone_hash = $1)
							  ORDER BY priority ASC, rc.tenant_id::text ASC`,
							ah, ph,
						)
					}
				}
			default: // staff
				// When userType="staff", search logic:
				// Step 1: Query users table by password_hash first, then filter by accountHash
				//   - Find records matching password_hash (passwordHash)
				//   - Then filter by email_hash (priority 1) OR phone_hash (priority 2) OR user_account_hash (priority 3)
				// Priority: email_hash > phone_hash > user_account_hash
				// Security: Only return tenant_id and account_type for matched institutions (already verified by password)
				rows, err = s.DB.QueryContext(
					r.Context(),
					`SELECT DISTINCT u.tenant_id::text,
					        CASE
					          WHEN u.email_hash = $1 THEN 'email'
					          WHEN u.phone_hash = $1 THEN 'phone'
					          WHEN u.user_account_hash = $1 THEN 'account'
					          ELSE 'account'
					        END as account_type,
					        CASE
					          WHEN u.email_hash = $1 THEN 1
					          WHEN u.phone_hash = $1 THEN 2
					          WHEN u.user_account_hash = $1 THEN 3
					          ELSE 4
					        END as priority
					   FROM users u
					  WHERE u.password_hash = $2
					    AND COALESCE(u.status,'active') = 'active'
					    AND (u.email_hash = $1 OR u.phone_hash = $1 OR u.user_account_hash = $1)
					  ORDER BY priority ASC, u.tenant_id::text ASC`,
					ah, ph,
				)
			}
			if err != nil {
				writeJSON(w, http.StatusOK, Ok([]any{}))
				return
			}
			defer rows.Close()
			items := []any{}
			type tenantInfo struct {
				id          string
				accountType string
			}
			tenantInfos := []tenantInfo{}
			for rows.Next() {
				var id string
				var accountType string
				var priority int // Priority value for ordering (we don't need it after ordering)
				if err := rows.Scan(&id, &accountType, &priority); err != nil {
					continue
				}
				tenantInfos = append(tenantInfos, tenantInfo{id: id, accountType: accountType})
			}
			// Return tenant_id, tenant_name, and accountType for matched institutions only
			// Security: Only return institutions that match both account and password (already verified)
			// This allows frontend to display and match by name, while preventing enumeration of all institutions
			if s != nil && s.Tenants != nil && len(tenantInfos) > 0 {
				ts, _, err := s.Tenants.ListTenants(r.Context(), "", 1, 1000)
				if err == nil {
					for _, ti := range tenantInfos {
						if ti.id == SystemTenantID() {
							items = append(items, map[string]any{
								"id":          SystemTenantID(),
								"name":        "System",
								"accountType": ti.accountType,
							})
							continue
						}
						for _, t := range ts {
							if t.TenantID == ti.id && t.Status != "deleted" {
								items = append(items, map[string]any{
									"id":          t.TenantID,
									"name":        t.TenantName,
									"accountType": ti.accountType,
								})
								break
							}
						}
					}
				}
			}
			// If Tenants service is not available, still return tenant_ids with accountType
			if len(items) == 0 && len(tenantInfos) > 0 {
				for _, ti := range tenantInfos {
					items = append(items, map[string]any{
						"id":          ti.id,
						"accountType": ti.accountType,
					})
				}
			}
			writeJSON(w, http.StatusOK, Ok(items))
			return
		}

		// If auth store fallback is explicitly enabled, only return tenants where this account exists.
		if s != nil && s.AuthStore != nil && allowAuthStoreFallback() && accountHash != "" {
			// AuthStore fallback is disabled as it uses accountPasswordHash, but we now only have passwordHash
			// This fallback is only used in stub/dev mode, and DB lookup is preferred anyway
			var tenantIDs []string
			items := []any{}
			// Always allow "System" if it's in tenantIDs
			for _, tid := range tenantIDs {
				if tid == SystemTenantID() {
					items = append(items, map[string]any{"id": SystemTenantID(), "name": "System", "domain": "system.local"})
				}
			}
			if s.Tenants != nil {
				ts, _, err := s.Tenants.ListTenants(r.Context(), "", 1, 1000)
				if err == nil {
					for _, tid := range tenantIDs {
						if tid == SystemTenantID() {
							continue
						}
						for _, t := range ts {
							if t.TenantID == tid && t.Status != "deleted" {
								items = append(items, map[string]any{
									"id":     t.TenantID,
									"name":   t.TenantName,
									"domain": t.Domain,
								})
								break
							}
						}
					}
				}
			}
			writeJSON(w, http.StatusOK, Ok(items))
			return
		}

		// No DB: return empty list (no fallback to prevent information disclosure)
		if s == nil || s.DB == nil {
			writeJSON(w, http.StatusOK, Ok([]any{}))
			return
		}

		// No match found: return empty list (do not return all tenants for security)
		writeJSON(w, http.StatusOK, Ok([]any{}))
		return
	case "/auth/api/v1/forgot-password/send-code":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, Fail("database not available"))
		return
	case "/auth/api/v1/forgot-password/verify-code":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, Fail("database not available"))
		return
	case "/auth/api/v1/forgot-password/reset":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, Fail("database not available"))
		return
	}
	w.WriteHeader(http.StatusNotFound)
}
