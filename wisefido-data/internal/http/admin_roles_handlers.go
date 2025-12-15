package httpapi

import (
	"database/sql"
	"net/http"
	"strings"
)

func (s *StubHandler) AdminRoles(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/admin/api/v1/roles" {
		switch r.Method {
		case http.MethodGet:
			if s != nil && s.DB != nil {
				// Global roles are stored under SystemTenantID() (no tenant custom roles by product rule).
				sysT := SystemTenantID()
				search := strings.TrimSpace(r.URL.Query().Get("search"))
				args := []any{sysT}
				q := `SELECT role_id::text,
				             COALESCE(tenant_id::text, NULL),
				             role_code,
				             description,
				             is_system,
				             is_active
				      FROM roles
				      WHERE tenant_id = $1`
				if search != "" {
					args = append(args, "%"+search+"%")
					q += ` AND (role_code ILIKE $2 OR description ILIKE $2)`
				}
				q += ` ORDER BY is_system DESC, role_code ASC`
				rows, err := s.DB.QueryContext(r.Context(), q, args...)
				if err != nil {
					writeJSON(w, http.StatusOK, Fail("failed to list roles"))
					return
				}
				defer rows.Close()
				items := []any{}
				for rows.Next() {
					var roleID, roleCode, desc string
					var tenantIDStr sql.NullString
					var isSystem, isActive bool
					if err := rows.Scan(&roleID, &tenantIDStr, &roleCode, &desc, &isSystem, &isActive); err != nil {
						writeJSON(w, http.StatusOK, Fail("failed to list roles"))
						return
					}
					displayName := roleCode
					if p := strings.SplitN(desc, "\n", 2); len(p) > 0 && strings.TrimSpace(p[0]) != "" {
						displayName = strings.TrimSpace(p[0])
					}
					item := map[string]any{
						"role_id":      roleID,
						"tenant_id":    nil,
						"role_code":    roleCode,
						"display_name": displayName,
						"description":  desc,
						"is_system":    isSystem,
						"is_active":    isActive,
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
			// Create tenant custom role (non-system)
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
				roleCode, _ := payload["role_code"].(string)
				displayName, _ := payload["display_name"].(string)
				desc, _ := payload["description"].(string)
				roleCode = strings.TrimSpace(roleCode)
				if roleCode == "" {
					writeJSON(w, http.StatusOK, Fail("role_code is required"))
					return
				}
				if displayName == "" {
					displayName = roleCode
				}
				// Store description in the schema's "two-line" format.
				fullDesc := strings.TrimSpace(displayName)
				if strings.TrimSpace(desc) != "" {
					fullDesc = fullDesc + "\n" + strings.TrimSpace(desc)
				}
				var roleID string
				err := s.DB.QueryRowContext(
					r.Context(),
					`INSERT INTO roles (tenant_id, role_code, description, is_system, is_active)
					 VALUES ($1, $2, $3, FALSE, TRUE)
					 RETURNING role_id::text`,
					tenantID, roleCode, fullDesc,
				).Scan(&roleID)
				if err != nil {
					writeJSON(w, http.StatusOK, Fail("failed to create role"))
					return
				}
				writeJSON(w, http.StatusOK, Ok(map[string]any{"role_id": roleID}))
				return
			}
			writeJSON(w, http.StatusOK, Fail("database not available"))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}
	if strings.HasPrefix(r.URL.Path, "/admin/api/v1/roles/") {
		path := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/roles/")
		if strings.HasSuffix(path, "/status") {
			if r.Method != http.MethodPut {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			// Update role active status
			id := strings.TrimSuffix(path, "/status")
			if id == "" || strings.Contains(id, "/") {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			if s != nil && s.DB != nil {
				var payload map[string]any
				if err := readBodyJSON(r, 1<<20, &payload); err != nil {
					writeJSON(w, http.StatusOK, Fail("invalid body"))
					return
				}
				isActive, _ := payload["is_active"].(bool)
				_, err := s.DB.ExecContext(r.Context(), `UPDATE roles SET is_active = $2 WHERE role_id::text = $1`, id, isActive)
				if err != nil {
					writeJSON(w, http.StatusOK, Fail("failed to update role status"))
					return
				}
				writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
				return
			}
			writeJSON(w, http.StatusOK, Fail("database not available"))
			return
		}
		if strings.Contains(path, "/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodPut:
			// Update role: supports {is_active:false} (disable) and {_delete:true} (delete) and edit fields for non-system roles.
			if s != nil && s.DB != nil {
				var payload map[string]any
				if err := readBodyJSON(r, 1<<20, &payload); err != nil {
					writeJSON(w, http.StatusOK, Fail("invalid body"))
					return
				}
				// Fetch current role flags
				var isSystem bool
				var roleCode string
				if err := s.DB.QueryRowContext(r.Context(), `SELECT is_system, role_code FROM roles WHERE role_id::text = $1`, path).Scan(&isSystem, &roleCode); err != nil {
					writeJSON(w, http.StatusOK, Fail("role not found"))
					return
				}

				if del, ok := payload["_delete"].(bool); ok && del {
					if isSystem {
						writeJSON(w, http.StatusOK, Fail("system roles cannot be deleted"))
						return
					}
					_, err := s.DB.ExecContext(r.Context(), `DELETE FROM roles WHERE role_id::text = $1`, path)
					if err != nil {
						writeJSON(w, http.StatusOK, Fail("failed to delete role"))
						return
					}
					writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
					return
				}

				if v, ok := payload["is_active"]; ok {
					isActive, _ := v.(bool)
					// Critical system roles that cannot be disabled: SystemAdmin, SystemOperator, Admin, Manager, Caregiver, Resident, Family
					protectedRoles := []string{"SystemAdmin", "SystemOperator", "Admin", "Manager", "Caregiver", "Resident", "Family"}
					if !isActive {
						for _, protected := range protectedRoles {
							if roleCode == protected {
								writeJSON(w, http.StatusOK, Fail(roleCode+" is a critical system role and cannot be disabled"))
								return
							}
						}
					}
					_, err := s.DB.ExecContext(r.Context(), `UPDATE roles SET is_active = $2 WHERE role_id::text = $1`, path, isActive)
					if err != nil {
						writeJSON(w, http.StatusOK, Fail("failed to update role"))
						return
					}
					writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
					return
				}

				if isSystem {
					// For system preset roles, only SystemAdmin can modify display_name and description
					userRole := r.Header.Get("X-User-Role")
					if !strings.EqualFold(userRole, "SystemAdmin") {
						writeJSON(w, http.StatusOK, Fail("system roles can only be modified by SystemAdmin"))
						return
					}
				}

				displayName, _ := payload["display_name"].(string)
				desc, _ := payload["description"].(string)
				if strings.TrimSpace(displayName) == "" {
					displayName = roleCode
				}
				fullDesc := strings.TrimSpace(displayName)
				if strings.TrimSpace(desc) != "" {
					fullDesc = fullDesc + "\n" + strings.TrimSpace(desc)
				}
				_, err := s.DB.ExecContext(r.Context(), `UPDATE roles SET description = $2 WHERE role_id::text = $1`, path, fullDesc)
				if err != nil {
					writeJSON(w, http.StatusOK, Fail("failed to update role"))
					return
				}
				writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
				return
			}
			writeJSON(w, http.StatusOK, Fail("database not available"))
			return
		case http.MethodDelete:
			// Delete role (non-system)
			if s != nil && s.DB != nil {
				var isSystem bool
				if err := s.DB.QueryRowContext(r.Context(), `SELECT is_system FROM roles WHERE role_id::text = $1`, path).Scan(&isSystem); err != nil {
					writeJSON(w, http.StatusOK, Fail("role not found"))
					return
				}
				if isSystem {
					writeJSON(w, http.StatusOK, Fail("system roles cannot be deleted"))
					return
				}
				_, err := s.DB.ExecContext(r.Context(), `DELETE FROM roles WHERE role_id::text = $1`, path)
				if err != nil {
					writeJSON(w, http.StatusOK, Fail("failed to delete role"))
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
