package httpapi

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
)

func (s *StubHandler) AdminRolePermissions(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/admin/api/v1/role-permissions":
		switch r.Method {
		case http.MethodGet:
			// DB-backed read for UI (view permissions)
			if s != nil && s.DB != nil {
				roleCode := strings.TrimSpace(r.URL.Query().Get("role_code"))
				resourceType := strings.TrimSpace(r.URL.Query().Get("resource_type"))
				permType := strings.TrimSpace(r.URL.Query().Get("permission_type"))
				// is_active is ignored (DB has no is_active per row; record presence means enabled)

				sysT := SystemTenantID()
				args := []any{sysT}
				q := `SELECT permission_id::text, COALESCE(tenant_id::text, NULL), role_code, resource_type, permission_type, assigned_only, branch_only
				      FROM role_permissions
				      WHERE tenant_id = $1`
				// Filters
				argIdx := len(args)
				if roleCode != "" {
					args = append(args, roleCode)
					argIdx++
					q += ` AND role_code = $` + strconv.Itoa(argIdx)
				}
				if resourceType != "" {
					args = append(args, resourceType)
					argIdx++
					q += ` AND resource_type = $` + strconv.Itoa(argIdx)
				}
				if permType != "" && permType != "manage" {
					pt := map[string]string{"read": "R", "create": "C", "update": "U", "delete": "D"}[permType]
					if pt != "" {
						args = append(args, pt)
						argIdx++
						q += ` AND permission_type = $` + strconv.Itoa(argIdx)
					}
				}
				q += ` ORDER BY role_code, resource_type, permission_type`

				rows, err := s.DB.QueryContext(r.Context(), q, args...)
				if err != nil {
					writeJSON(w, http.StatusOK, Fail("failed to list role permissions"))
					return
				}
				defer rows.Close()

				items := []any{}
				for rows.Next() {
					var pid, rc, rt, pt string
					var tenantIDStr sql.NullString
					var assignedOnly, branchOnly bool
					if err := rows.Scan(&pid, &tenantIDStr, &rc, &rt, &pt, &assignedOnly, &branchOnly); err != nil {
						writeJSON(w, http.StatusOK, Fail("failed to list role permissions"))
						return
					}
					perm := map[string]string{"R": "read", "C": "create", "U": "update", "D": "delete"}[pt]
					scope := "all"
					if assignedOnly {
						scope = "assigned_only"
					}
					item := map[string]any{
						"permission_id":   pid,
						"role_code":       rc,
						"resource_type":   rt,
						"permission_type": perm,
						"scope":           scope,
						"branch_only":     branchOnly,
						"is_active":       true,
					}
					if tenantIDStr.Valid {
						item["tenant_id"] = tenantIDStr.String
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
				// Only System tenant's SystemAdmin can modify global role permissions
				if tenantID != SystemTenantID() || !strings.EqualFold(r.Header.Get("X-User-Role"), "SystemAdmin") {
					writeJSON(w, http.StatusOK, Fail("only System tenant's SystemAdmin can modify role permissions"))
					return
				}
				var payload map[string]any
				if err := readBodyJSON(r, 1<<20, &payload); err != nil {
					writeJSON(w, http.StatusOK, Fail("invalid body"))
					return
				}
				roleCode, _ := payload["role_code"].(string)
				resourceType, _ := payload["resource_type"].(string)
				permType, _ := payload["permission_type"].(string)
				scope, _ := payload["scope"].(string)
				branchOnly, _ := payload["branch_only"].(bool)
				roleCode = strings.TrimSpace(roleCode)
				resourceType = strings.TrimSpace(resourceType)
				permType = strings.TrimSpace(permType)
				if roleCode == "" || resourceType == "" || permType == "" {
					writeJSON(w, http.StatusOK, Fail("role_code, resource_type, permission_type are required"))
					return
				}
				assignedOnly := strings.TrimSpace(scope) == "assigned_only"
				sysT := SystemTenantID()
				pt := map[string]string{"read": "R", "create": "C", "update": "U", "delete": "D"}[permType]
				if pt == "" {
					writeJSON(w, http.StatusOK, Fail("invalid permission_type"))
					return
				}
				var permissionID string
				err := s.DB.QueryRowContext(
					r.Context(),
					`INSERT INTO role_permissions (tenant_id, role_code, resource_type, permission_type, assigned_only, branch_only)
					 VALUES ($1, $2, $3, $4, $5, $6)
					 ON CONFLICT ((COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid)), role_code, resource_type, permission_type)
					 DO UPDATE SET assigned_only = EXCLUDED.assigned_only, branch_only = EXCLUDED.branch_only
					 RETURNING permission_id::text`,
					sysT, roleCode, resourceType, pt, assignedOnly, branchOnly,
				).Scan(&permissionID)
				if err != nil {
					writeJSON(w, http.StatusOK, Fail("failed to create permission"))
					return
				}
				writeJSON(w, http.StatusOK, Ok(map[string]any{"permission_id": permissionID}))
				return
			}
			writeJSON(w, http.StatusOK, Fail("database not available"))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	case r.URL.Path == "/admin/api/v1/role-permissions/batch":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// Replace permissions for a role (dev UI "Save" helper).
		// In DB mode: write to role_permissions under SystemTenantID() (global defaults).
		if s != nil && s.DB != nil {
			tenantID, ok := s.tenantIDFromReq(w, r)
			if !ok {
				return
			}
			// Only System tenant's SystemAdmin can modify global role permissions
			if tenantID != SystemTenantID() || !strings.EqualFold(r.Header.Get("X-User-Role"), "SystemAdmin") {
				writeJSON(w, http.StatusOK, Fail("only System tenant's SystemAdmin can modify role permissions"))
				return
			}
			var payload map[string]any
			if err := readBodyJSON(r, 1<<20, &payload); err != nil {
				writeJSON(w, http.StatusOK, Fail("invalid body"))
				return
			}
			roleCode, _ := payload["role_code"].(string)
			roleCode = strings.TrimSpace(roleCode)
			if roleCode == "" {
				writeJSON(w, http.StatusOK, Fail("role_code is required"))
				return
			}
			permsAny, _ := payload["permissions"].([]any)
			sysT := SystemTenantID()

			tx, err := s.DB.BeginTx(r.Context(), nil)
			if err != nil {
				writeJSON(w, http.StatusOK, Fail("failed to start transaction"))
				return
			}
			defer tx.Rollback()

			// Wipe existing global defaults for this role_code (we treat presence as enabled).
			if _, err := tx.ExecContext(r.Context(), `DELETE FROM role_permissions WHERE tenant_id = $1 AND role_code = $2`, sysT, roleCode); err != nil {
				writeJSON(w, http.StatusOK, Fail("failed to update permissions"))
				return
			}

			success := 0
			failed := 0
			for _, p := range permsAny {
				m, ok := p.(map[string]any)
				if !ok {
					failed++
					continue
				}
				resourceType, _ := m["resource_type"].(string)
				permType, _ := m["permission_type"].(string)
				scope, _ := m["scope"].(string)
				branchOnly, _ := m["branch_only"].(bool)
				isActive := true
				if v, ok := m["is_active"].(bool); ok {
					isActive = v
				}
				if !isActive {
					continue
				}
				resourceType = strings.TrimSpace(resourceType)
				permType = strings.TrimSpace(permType)
				if resourceType == "" || permType == "" {
					failed++
					continue
				}
				assignedOnly := strings.TrimSpace(scope) == "assigned_only"

				letters := []string{}
				switch permType {
				case "manage":
					letters = []string{"R", "C", "U", "D"}
				case "read":
					letters = []string{"R"}
				case "create":
					letters = []string{"C"}
				case "update":
					letters = []string{"U"}
				case "delete":
					letters = []string{"D"}
				default:
					failed++
					continue
				}
				for _, l := range letters {
					_, err := tx.ExecContext(
						r.Context(),
						`INSERT INTO role_permissions (tenant_id, role_code, resource_type, permission_type, assigned_only, branch_only)
						 VALUES ($1, $2, $3, $4, $5, $6)
						 ON CONFLICT ((COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid)), role_code, resource_type, permission_type)
						 DO UPDATE SET assigned_only = EXCLUDED.assigned_only, branch_only = EXCLUDED.branch_only`,
						sysT, roleCode, resourceType, l, assignedOnly, branchOnly,
					)
					if err != nil {
						failed++
						continue
					}
					success++
				}
			}

			if err := tx.Commit(); err != nil {
				writeJSON(w, http.StatusOK, Fail("failed to commit permissions"))
				return
			}
			writeJSON(w, http.StatusOK, Ok(map[string]any{"success_count": success, "failed_count": failed}))
			return
		}
		writeJSON(w, http.StatusOK, Fail("database not available"))
		return
	case r.URL.Path == "/admin/api/v1/role-permissions/resource-types":
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if s != nil && s.DB != nil {
			rows, err := s.DB.QueryContext(r.Context(), `SELECT DISTINCT resource_type FROM role_permissions ORDER BY resource_type`)
			if err != nil {
				writeJSON(w, http.StatusOK, Fail("failed to list resource types"))
				return
			}
			defer rows.Close()
			out := []string{}
			for rows.Next() {
				var rt string
				if err := rows.Scan(&rt); err == nil && rt != "" {
					out = append(out, rt)
				}
			}
			writeJSON(w, http.StatusOK, Ok(map[string]any{"resource_types": out}))
			return
		}
		writeJSON(w, http.StatusOK, Fail("database not available"))
		return
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/role-permissions/") && strings.HasSuffix(r.URL.Path, "/status"):
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if s != nil && s.DB != nil {
			tenantID, ok := s.tenantIDFromReq(w, r)
			if !ok {
				return
			}
			// Only System tenant's SystemAdmin can modify global role permissions
			if tenantID != SystemTenantID() || !strings.EqualFold(r.Header.Get("X-User-Role"), "SystemAdmin") {
				writeJSON(w, http.StatusOK, Fail("only System tenant's SystemAdmin can modify role permissions"))
				return
			}
			permissionID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/admin/api/v1/role-permissions/"), "/status")
			if permissionID == "" || strings.Contains(permissionID, "/") {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			var payload map[string]any
			if err := readBodyJSON(r, 1<<20, &payload); err != nil {
				writeJSON(w, http.StatusOK, Fail("invalid body"))
				return
			}
			isActive, _ := payload["is_active"].(bool)
			if !isActive {
				// Delete the permission (presence means enabled)
				_, err := s.DB.ExecContext(r.Context(), `DELETE FROM role_permissions WHERE permission_id::text = $1`, permissionID)
				if err != nil {
					writeJSON(w, http.StatusOK, Fail("failed to update permission status"))
					return
				}
			} else {
				// Re-insert if needed (this shouldn't happen as presence means enabled)
				writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
				return
			}
			writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
			return
		}
		writeJSON(w, http.StatusOK, Fail("database not available"))
		return
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/role-permissions/"):
		if r.Method != http.MethodPut && r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if s != nil && s.DB != nil {
			tenantID, ok := s.tenantIDFromReq(w, r)
			if !ok {
				return
			}
			// Only System tenant's SystemAdmin can modify global role permissions
			if tenantID != SystemTenantID() || !strings.EqualFold(r.Header.Get("X-User-Role"), "SystemAdmin") {
				writeJSON(w, http.StatusOK, Fail("only System tenant's SystemAdmin can modify role permissions"))
				return
			}
			permissionID := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/role-permissions/")
			if permissionID == "" || strings.Contains(permissionID, "/") {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			if r.Method == http.MethodDelete {
				_, err := s.DB.ExecContext(r.Context(), `DELETE FROM role_permissions WHERE permission_id::text = $1`, permissionID)
				if err != nil {
					writeJSON(w, http.StatusOK, Fail("failed to delete permission"))
					return
				}
				writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
				return
			}
			// PUT: update permission
			var payload map[string]any
			if err := readBodyJSON(r, 1<<20, &payload); err != nil {
				writeJSON(w, http.StatusOK, Fail("invalid body"))
				return
			}
			scope, _ := payload["scope"].(string)
			branchOnly, _ := payload["branch_only"].(bool)
			assignedOnly := strings.TrimSpace(scope) == "assigned_only"
			_, err := s.DB.ExecContext(r.Context(), `UPDATE role_permissions SET assigned_only = $2, branch_only = $3 WHERE permission_id::text = $1`, permissionID, assignedOnly, branchOnly)
			if err != nil {
				writeJSON(w, http.StatusOK, Fail("failed to update permission"))
				return
			}
			writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
			return
		}
		writeJSON(w, http.StatusOK, Fail("database not available"))
		return
	}
	w.WriteHeader(http.StatusNotFound)
}
