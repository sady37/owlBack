package httpapi

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"wisefido-data/internal/repository"

	"github.com/lib/pq"
)

// StubHandler：用于 DB/真实逻辑未就绪时，先保证 owlFront 不 404、页面可渲染（code=2000 + 空数据）
type StubHandler struct {
	Tenants   repository.TenantsRepo
	AuthStore *AuthStore
	DB        *sql.DB // optional: when set, some admin endpoints read/write real DB
}

func NewStubHandler(tenants repository.TenantsRepo, auth *AuthStore, db *sql.DB) *StubHandler {
	return &StubHandler{Tenants: tenants, AuthStore: auth, DB: db}
}

func allowAuthStoreFallback() bool {
	// Security hardening:
	// - AuthStore is in-memory and should NOT be used in real deployments.
	// - Only allow it when explicitly enabled for local dev.
	return os.Getenv("ALLOW_AUTHSTORE_FALLBACK") == "true"
}

func (s *StubHandler) tenantIDFromReq(w http.ResponseWriter, r *http.Request) (string, bool) {
	if tid := r.URL.Query().Get("tenant_id"); tid != "" && tid != "null" {
		return tid, true
	}
	if tid := r.Header.Get("X-Tenant-Id"); tid != "" && tid != "null" {
		return tid, true
	}
	// Try to resolve tenant from user ID via DB query (if DB is available)
	if s != nil && s.DB != nil {
		userID := r.Header.Get("X-User-Id")
		if userID != "" {
			var tenantID string
			err := s.DB.QueryRowContext(r.Context(), "SELECT tenant_id::text FROM users WHERE user_id = $1", userID).Scan(&tenantID)
			if err == nil && tenantID != "" {
				return tenantID, true
			}
		}
	}
	// Convenience: SystemAdmin without tenant header falls back to System tenant.
	if strings.EqualFold(r.Header.Get("X-User-Role"), "SystemAdmin") {
		return SystemTenantID(), true
	}
	writeJSON(w, http.StatusOK, Fail("tenant_id is required"))
	return "", false
}

// --- admin/api/v1 ---

func (s *StubHandler) AdminDevices(w http.ResponseWriter, r *http.Request) {
	// GET /admin/api/v1/devices
	// GET /admin/api/v1/devices/:id
	// PUT /admin/api/v1/devices/:id
	// DELETE /admin/api/v1/devices/:id
	if r.URL.Path == "/admin/api/v1/devices" {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, Ok(map[string]any{"items": []any{}, "total": 0}))
		return
	}
	if strings.HasPrefix(r.URL.Path, "/admin/api/v1/devices/") {
		id := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/devices/")
		if id == "" || strings.Contains(id, "/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodGet:
			// 参考 owlFront Device model 的必填字段：device_id/tenant_id/device_name/status/business_access
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"device_id":          id,
				"tenant_id":          "",
				"device_name":        "stub-" + id,
				"status":             "offline",
				"business_access":    "pending",
				"monitoring_enabled": false,
			}))
		case http.MethodPut:
			writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
		case http.MethodDelete:
			writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

func (s *StubHandler) AdminUnits(w http.ResponseWriter, r *http.Request) {
	// buildings
	switch {
	case r.URL.Path == "/admin/api/v1/buildings":
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, Ok([]any{}))
		case http.MethodPost:
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"building_id":   "stub-building",
				"building_name": "stub",
				"floors":        0,
				"tenant_id":     "",
			}))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/buildings/"):
		id := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/buildings/")
		if id == "" || strings.Contains(id, "/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodPut:
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"building_id":   id,
				"building_name": "stub",
				"floors":        0,
				"tenant_id":     "",
			}))
		case http.MethodDelete:
			writeJSON(w, http.StatusOK, Ok[any](nil))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	// units
	switch {
	case r.URL.Path == "/admin/api/v1/units":
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, Ok(map[string]any{"items": []any{}, "total": 0}))
		case http.MethodPost:
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"unit_id":     "stub-unit",
				"tenant_id":   "",
				"unit_name":   "stub",
				"unit_number": "stub",
				"unit_type":   "Facility",
			}))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/units/"):
		id := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/units/")
		if id == "" || strings.Contains(id, "/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"unit_id":     id,
				"tenant_id":   "",
				"unit_name":   "stub-" + id,
				"unit_number": "stub",
				"unit_type":   "Facility",
			}))
		case http.MethodPut:
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"unit_id":     id,
				"tenant_id":   "",
				"unit_name":   "stub-" + id,
				"unit_number": "stub",
				"unit_type":   "Facility",
			}))
		case http.MethodDelete:
			writeJSON(w, http.StatusOK, Ok[any](nil))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	// rooms
	switch {
	case r.URL.Path == "/admin/api/v1/rooms":
		switch r.Method {
		case http.MethodGet:
			// getRoomsApi 期待 RoomWithBeds[]
			writeJSON(w, http.StatusOK, Ok([]any{}))
		case http.MethodPost:
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"room_id":    "stub-room",
				"unit_id":    "",
				"room_name":  "stub",
				"is_default": false,
			}))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/rooms/"):
		id := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/rooms/")
		if id == "" || strings.Contains(id, "/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodPut:
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"room_id":    id,
				"unit_id":    "",
				"room_name":  "stub-" + id,
				"is_default": false,
			}))
		case http.MethodDelete:
			writeJSON(w, http.StatusOK, Ok[any](nil))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	// beds
	switch {
	case r.URL.Path == "/admin/api/v1/beds":
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, Ok([]any{}))
		case http.MethodPost:
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"bed_id":   "stub-bed",
				"room_id":  "",
				"bed_name": "stub",
			}))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/beds/"):
		id := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/beds/")
		if id == "" || strings.Contains(id, "/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodPut:
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"bed_id":   id,
				"room_id":  "",
				"bed_name": "stub-" + id,
			}))
		case http.MethodDelete:
			writeJSON(w, http.StatusOK, Ok[any](nil))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

func (s *StubHandler) AdminResidents(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/admin/api/v1/residents" {
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, Ok(map[string]any{"items": []any{}, "total": 0}))
		case http.MethodPost:
			writeJSON(w, http.StatusOK, Ok(map[string]any{"resident_id": "stub-resident"}))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}
	if strings.HasPrefix(r.URL.Path, "/admin/api/v1/residents/") {
		path := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/residents/")
		// subresources
		if strings.HasSuffix(path, "/phi") {
			if r.Method != http.MethodPut {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
			return
		}
		if strings.HasSuffix(path, "/contacts") {
			if r.Method != http.MethodPut {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
			return
		}
		id := path
		if strings.Contains(id, "/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"resident_id": id,
				"tenant_id":   "",
				"status":      "active",
			}))
		case http.MethodPut:
			writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
		case http.MethodDelete:
			writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

func (s *StubHandler) AdminTags(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/admin/api/v1/tags":
		switch r.Method {
		case http.MethodDelete:
			// DELETE /admin/api/v1/tags?tenant_id=xxx&tag_name=xxx
			if s != nil && s.DB != nil {
				tenantID, ok := s.tenantIDFromReq(w, r)
				if !ok {
					return
				}
				tagName := strings.TrimSpace(r.URL.Query().Get("tag_name"))
				if tagName == "" {
					writeJSON(w, http.StatusOK, Fail("tag_name is required"))
					return
				}

				// Call drop_tag function
				var result bool
				err := s.DB.QueryRowContext(
					r.Context(),
					`SELECT drop_tag($1::uuid, $2)`,
					tenantID, tagName,
				).Scan(&result)
				if err != nil {
					fmt.Printf("[AdminTags] Failed to delete tag: %v\n", err)
					writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to delete tag: %v", err)))
					return
				}
				writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
				return
			}
			writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
			return
		case http.MethodGet:
			if s != nil && s.DB != nil {
				tenantID, ok := s.tenantIDFromReq(w, r)
				if !ok {
					return
				}
				tagTypeFilter := strings.TrimSpace(r.URL.Query().Get("tag_type"))
				includeSystem := r.URL.Query().Get("include_system_tag_types") != "false"

				where := "tenant_id = $1"
				args := []any{tenantID}
				argIdx := 2

				if tagTypeFilter != "" {
					where += fmt.Sprintf(" AND tag_type = $%d", argIdx)
					args = append(args, tagTypeFilter)
					argIdx++
				} else if !includeSystem {
					// Exclude system predefined tag types
					where += fmt.Sprintf(" AND tag_type NOT IN ($%d, $%d, $%d)", argIdx, argIdx+1, argIdx+2)
					args = append(args, "branch_tag", "family_tag", "area_tag")
					argIdx += 3
				}

				q := fmt.Sprintf(`
					SELECT tag_id::text, tenant_id::text, tag_type, tag_name, tag_objects
					FROM tags_catalog
					WHERE %s
					ORDER BY tag_type, tag_name
				`, where)

				rows, err := s.DB.QueryContext(r.Context(), q, args...)
				if err != nil {
					fmt.Printf("[AdminTags] Query error: %v\n", err)
					writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to list tags: %v", err)))
					return
				}
				defer rows.Close()

				items := []any{}
				for rows.Next() {
					var tagID, tid, tagType, tagName string
					var tagObjectsRaw []byte
					if err := rows.Scan(&tagID, &tid, &tagType, &tagName, &tagObjectsRaw); err != nil {
						fmt.Printf("[AdminTags] Scan error: %v\n", err)
						writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to scan tag: %v", err)))
						return
					}
					var tagObjects map[string]any
					if len(tagObjectsRaw) > 0 {
						_ = json.Unmarshal(tagObjectsRaw, &tagObjects)
					}
					if tagObjects == nil {
						tagObjects = make(map[string]any)
					}
					items = append(items, map[string]any{
						"tag_id":      tagID,
						"tenant_id":   tid,
						"tag_type":    tagType,
						"tag_name":    tagName,
						"tag_objects": tagObjects,
					})
				}

				writeJSON(w, http.StatusOK, Ok(map[string]any{
					"items":                       items,
					"total":                       len(items),
					"available_tag_types":         []string{"branch_tag", "family_tag", "area_tag", "user_tag"},
					"system_predefined_tag_types": []string{"branch_tag", "family_tag", "area_tag"},
				}))
				return
			}
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"items":                       []any{},
				"total":                       0,
				"available_tag_types":         []string{"branch_tag", "family_tag", "area_tag", "user_tag"},
				"system_predefined_tag_types": []string{"branch_tag", "family_tag", "area_tag"},
			}))
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

				tagName, _ := payload["tag_name"].(string)
				tagName = strings.TrimSpace(tagName)
				if tagName == "" {
					writeJSON(w, http.StatusOK, Fail("tag_name is required"))
					return
				}

				tagType, _ := payload["tag_type"].(string)
				if tagType == "" {
					tagType = "user_tag" // Default to user_tag
				}

				// Validate tag_type
				allowedTypes := map[string]bool{
					"branch_tag": true,
					"family_tag": true,
					"area_tag":   true,
					"user_tag":   true,
				}
				if !allowedTypes[tagType] {
					writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("invalid tag_type: %s", tagType)))
					return
				}

				// Use upsert_tag_to_catalog function
				var tagID string
				err := s.DB.QueryRowContext(
					r.Context(),
					`SELECT upsert_tag_to_catalog($1, $2, $3)::text`,
					tenantID, tagName, tagType,
				).Scan(&tagID)
				if err != nil {
					fmt.Printf("[AdminTags] Create tag error: %v, tenant_id=%s, tag_name=%s, tag_type=%s\n", err, tenantID, tagName, tagType)
					writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to create or find tag: %v", err)))
					return
				}

				writeJSON(w, http.StatusOK, Ok(map[string]any{"tag_id": tagID}))
				return
			}
			writeJSON(w, http.StatusOK, Ok(map[string]any{"tag_id": "stub-tag"}))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/tags/") && strings.HasSuffix(r.URL.Path, "/objects"):
		if r.Method == http.MethodPost {
			if s != nil && s.DB != nil {
				// Extract tag_id from path: /admin/api/v1/tags/{tag_id}/objects
				path := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/tags/")
				tagID := strings.TrimSuffix(path, "/objects")
				if tagID == "" || strings.Contains(tagID, "/") {
					w.WriteHeader(http.StatusNotFound)
					return
				}

				var payload map[string]any
				if err := readBodyJSON(r, 1<<20, &payload); err != nil {
					writeJSON(w, http.StatusOK, Fail("invalid body"))
					return
				}

				objectType, _ := payload["object_type"].(string)
				objects, _ := payload["objects"].([]any)

				if objectType == "" || len(objects) == 0 {
					writeJSON(w, http.StatusOK, Fail("object_type and objects are required"))
					return
				}

				// Add each object using update_tag_objects function
				for _, obj := range objects {
					objMap, ok := obj.(map[string]any)
					if !ok {
						continue
					}
					objectID, _ := objMap["object_id"].(string)
					objectName, _ := objMap["object_name"].(string)

					if objectID == "" || objectName == "" {
						continue
					}

					// Use PostgreSQL UUID casting
					_, err := s.DB.ExecContext(
						r.Context(),
						`SELECT update_tag_objects($1::uuid, $2, $3::uuid, $4, 'add')`,
						tagID, objectType, objectID, objectName,
					)
					if err != nil {
						fmt.Printf("[AdminTags] Failed to add tag object: %v\n", err)
						writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to add tag object: %v", err)))
						return
					}
				}

				writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
				return
			}
			writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
			return
		}
		if r.Method == http.MethodDelete {
			// Delete tag objects
			if s != nil && s.DB != nil {
				path := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/tags/")
				tagID := strings.TrimSuffix(path, "/objects")
				if tagID == "" || strings.Contains(tagID, "/") {
					w.WriteHeader(http.StatusNotFound)
					return
				}

				var payload map[string]any
				if err := readBodyJSON(r, 1<<20, &payload); err != nil {
					writeJSON(w, http.StatusOK, Fail("invalid body"))
					return
				}

				objectType, _ := payload["object_type"].(string)
				// Support both formats:
				// 1. object_ids: string[] (from frontend)
				// 2. objects: [{object_id, object_name}] (alternative format)
				objectIDs, _ := payload["object_ids"].([]any)
				objects, _ := payload["objects"].([]any)

				if objectType == "" {
					writeJSON(w, http.StatusOK, Fail("object_type is required"))
					return
				}

				// Handle object_ids format (array of strings)
				if objectIDs != nil && len(objectIDs) > 0 {
					// Get tag_name for user_tag type (needed for syncing users.tags)
					var tagName string
					var tagType string
					err := s.DB.QueryRowContext(
						r.Context(),
						`SELECT tag_name, tag_type FROM tags_catalog WHERE tag_id = $1`,
						tagID,
					).Scan(&tagName, &tagType)
					if err != nil {
						fmt.Printf("[AdminTags] Failed to get tag info: %v\n", err)
						writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to get tag info: %v", err)))
						return
					}

					for _, objIDAny := range objectIDs {
						objectID, ok := objIDAny.(string)
						if !ok {
							continue
						}
						if objectID == "" {
							continue
						}

						// For remove action, object_name is optional (can be empty string)
						// The function will work with just object_id
						_, err := s.DB.ExecContext(
							r.Context(),
							`SELECT update_tag_objects($1::uuid, $2, $3::uuid, '', 'remove')`,
							tagID, objectType, objectID,
						)
						if err != nil {
							fmt.Printf("[AdminTags] Failed to remove tag object: %v\n", err)
							writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to remove tag object: %v", err)))
							return
						}

						// If removing a user from user_tag, also remove the tag from user's tags JSONB
						if objectType == "user" && tagType == "user_tag" {
							// Use jsonb - operator to remove the tag from array
							_, err = s.DB.ExecContext(
								r.Context(),
								`UPDATE users 
								 SET tags = tags - $1
								 WHERE user_id = $2::uuid
								   AND tags IS NOT NULL
								   AND tags ? $1`,
								tagName, objectID,
							)
							if err != nil {
								fmt.Printf("[AdminTags] Failed to remove tag from user's tags: %v\n", err)
								// Don't fail the whole operation, just log the error
							}
						}
					}
					writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
					return
				}

				// Handle objects format (array of {object_id, object_name})
				if objects != nil && len(objects) > 0 {
					// Get tag_name for user_tag type (needed for syncing users.tags)
					var tagName string
					var tagType string
					err := s.DB.QueryRowContext(
						r.Context(),
						`SELECT tag_name, tag_type FROM tags_catalog WHERE tag_id = $1`,
						tagID,
					).Scan(&tagName, &tagType)
					if err != nil {
						fmt.Printf("[AdminTags] Failed to get tag info: %v\n", err)
						writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to get tag info: %v", err)))
						return
					}

					for _, obj := range objects {
						objMap, ok := obj.(map[string]any)
						if !ok {
							continue
						}
						objectID, _ := objMap["object_id"].(string)
						objectName, _ := objMap["object_name"].(string)

						if objectID == "" {
							continue
						}

						_, err := s.DB.ExecContext(
							r.Context(),
							`SELECT update_tag_objects($1::uuid, $2, $3::uuid, $4, 'remove')`,
							tagID, objectType, objectID, objectName,
						)
						if err != nil {
							fmt.Printf("[AdminTags] Failed to remove tag object: %v\n", err)
							writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to remove tag object: %v", err)))
							return
						}

						// If removing a user from user_tag, also remove the tag from user's tags JSONB
						if objectType == "user" && tagType == "user_tag" {
							// Use jsonb - operator to remove the tag from array
							_, err = s.DB.ExecContext(
								r.Context(),
								`UPDATE users 
								 SET tags = tags - $1
								 WHERE user_id = $2::uuid
								   AND tags IS NOT NULL
								   AND tags ? $1`,
								tagName, objectID,
							)
							if err != nil {
								fmt.Printf("[AdminTags] Failed to remove tag from user's tags: %v\n", err)
								// Don't fail the whole operation, just log the error
							}
						}
					}
					writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
					return
				}

				writeJSON(w, http.StatusOK, Fail("object_ids or objects is required"))
				return
			}
			writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	case r.URL.Path == "/admin/api/v1/tags/types":
		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
		return
	case r.URL.Path == "/admin/api/v1/tags/for-object":
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, Ok([]any{}))
		return
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/tags/"):
		// PUT /admin/api/v1/tags/:id
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

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
				             alarm_scope, last_login_at,
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
					var alarmScope sql.NullString
					var lastLoginAt sql.NullTime
					var tagsRaw, prefRaw []byte
					if err := rows.Scan(
						&userID, &tid, &userAccount, &nickname, &email, &phone, &role, &status,
						pq.Array(&alarmLevels), pq.Array(&alarmChannels), &alarmScope, &lastLoginAt, &tagsRaw, &prefRaw,
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
			// No DB: return users from AuthStore (dev/stub)
			if s != nil && s.AuthStore != nil {
				tenantID, ok := s.tenantIDFromReq(w, r)
				if !ok {
					return
				}
				users := s.AuthStore.ListUsersByTenant(tenantID)
				out := make([]any, 0, len(users))
				for _, u := range users {
					out = append(out, map[string]any{
						"user_id":        u.UserID,
						"tenant_id":      u.TenantID,
						"user_account":   u.UserAccount,
						"role":           u.Role,
						"status":         "active",
						"alarm_levels":   []string{},
						"alarm_channels": []string{},
					})
				}
				writeJSON(w, http.StatusOK, Ok(map[string]any{"items": out, "total": len(out)}))
				return
			}
			writeJSON(w, http.StatusOK, Ok(map[string]any{"items": []any{}, "total": 0}))
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
				aph, _ := hex.DecodeString(HashAccountPassword(userAccount, password))
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
						alarmScope = sql.NullString{String: "BRANCH-TAG", Valid: true}
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
			writeJSON(w, http.StatusOK, Ok(map[string]any{"user_id": "stub-user"}))
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
				newPassword = strings.TrimSpace(newPassword)
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

				aph, _ := hex.DecodeString(HashAccountPassword(userAccount, newPassword))
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
			}
			writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true, "message": "ok"}))
			return
		}
		if strings.HasSuffix(path, "/reset-pin") {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true, "message": "stub"}))
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

				if len(updates) == 0 {
					writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
					return
				}

				query := fmt.Sprintf(`UPDATE users SET %s WHERE tenant_id = $1 AND user_id::text = $2`, strings.Join(updates, ", "))
				_, err := s.DB.ExecContext(r.Context(), query, args...)
				if err != nil {
					writeJSON(w, http.StatusOK, Fail("failed to update user"))
					return
				}
				writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
				return
			}
			writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
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
			}
			writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
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
					alarmScope                                                             sql.NullString
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
					        last_login_at,
					        COALESCE(tags, '[]'::jsonb) as tags,
					        COALESCE(preferences, '{}'::jsonb) as preferences
					   FROM users
					  WHERE tenant_id = $1 AND user_id::text = $2`,
					tenantID, id,
				).Scan(
					&userID, &tenantIDStr, &userAccount, &nickname, &email, &phone, &role, &status,
					pq.Array(&alarmLevels), pq.Array(&alarmChannels), &alarmScope, &lastLoginAt, &tagsRaw, &prefRaw,
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
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"user_id":        id,
				"tenant_id":      "",
				"user_account":   "stub",
				"role":           "Staff",
				"status":         "active",
				"alarm_levels":   []string{},
				"alarm_channels": []string{},
			}))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

func (s *StubHandler) AdminRoles(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/admin/api/v1/roles" {
		switch r.Method {
		case http.MethodGet:
			// If DB is not configured, still return system preset roles so the UI isn't blank.
			// In the "global = System tenant" model, global defaults are stored under SystemTenantID().
			if s == nil || s.DB == nil {
				sysT := SystemTenantID()
				items := []any{
					map[string]any{"role_id": "00000000-0000-0000-0000-000000000101", "tenant_id": sysT, "role_code": "SystemAdmin", "display_name": "SystemAdmin", "description": "System administrator", "is_system": true, "is_active": true},
					map[string]any{"role_id": "00000000-0000-0000-0000-000000000102", "tenant_id": sysT, "role_code": "Admin", "display_name": "Admin", "description": "Administrator", "is_system": true, "is_active": true},
					map[string]any{"role_id": "00000000-0000-0000-0000-000000000103", "tenant_id": sysT, "role_code": "Manager", "display_name": "Manager", "description": "Manager", "is_system": true, "is_active": true},
					map[string]any{"role_id": "00000000-0000-0000-0000-000000000104", "tenant_id": sysT, "role_code": "IT", "display_name": "IT", "description": "IT Support", "is_system": true, "is_active": true},
					map[string]any{"role_id": "00000000-0000-0000-0000-000000000105", "tenant_id": sysT, "role_code": "Nurse", "display_name": "Nurse", "description": "Nurse", "is_system": true, "is_active": true},
					map[string]any{"role_id": "00000000-0000-0000-0000-000000000106", "tenant_id": sysT, "role_code": "Caregiver", "display_name": "Caregiver", "description": "Caregiver", "is_system": true, "is_active": true},
					map[string]any{"role_id": "00000000-0000-0000-0000-000000000107", "tenant_id": sysT, "role_code": "Resident", "display_name": "Resident", "description": "Resident", "is_system": true, "is_active": true},
					map[string]any{"role_id": "00000000-0000-0000-0000-000000000108", "tenant_id": sysT, "role_code": "Family", "display_name": "Family", "description": "Family", "is_system": true, "is_active": true},
				}
				writeJSON(w, http.StatusOK, Ok(map[string]any{"items": items, "total": len(items)}))
				return
			}
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
			writeJSON(w, http.StatusOK, Ok(map[string]any{"items": []any{}, "total": 0}))
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
			writeJSON(w, http.StatusOK, Ok(map[string]any{"role_id": "stub-role"}))
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
			writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
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
					_, err := s.DB.ExecContext(r.Context(), `UPDATE roles SET is_active = $2 WHERE role_id::text = $1`, path, isActive)
					if err != nil {
						writeJSON(w, http.StatusOK, Fail("failed to update role"))
						return
					}
					writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
					return
				}

				if isSystem {
					// For system preset roles, only is_active is mutable (enforced here).
					writeJSON(w, http.StatusOK, Fail("system roles cannot be modified"))
					return
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
			writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
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
			writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

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
				q := `SELECT permission_id::text, COALESCE(tenant_id::text, NULL), role_code, resource_type, permission_type, assigned_only
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
					var assignedOnly bool
					if err := rows.Scan(&pid, &tenantIDStr, &rc, &rt, &pt, &assignedOnly); err != nil {
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
			writeJSON(w, http.StatusOK, Ok(map[string]any{"items": []any{}, "total": 0}))
		case http.MethodPost:
			writeJSON(w, http.StatusOK, Ok(map[string]any{"permission_id": "stub-permission"}))
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
						`INSERT INTO role_permissions (tenant_id, role_code, resource_type, permission_type, assigned_only)
						 VALUES ($1, $2, $3, $4, $5)
						 ON CONFLICT ((COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid)), role_code, resource_type, permission_type)
						 DO UPDATE SET assigned_only = EXCLUDED.assigned_only`,
						sysT, roleCode, resourceType, l, assignedOnly,
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

		// No DB: accept but no-op
		writeJSON(w, http.StatusOK, Ok(map[string]any{"success_count": 0, "failed_count": 0}))
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
		writeJSON(w, http.StatusOK, Ok(map[string]any{"resource_types": []string{}}))
		return
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/role-permissions/") && strings.HasSuffix(r.URL.Path, "/status"):
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
		return
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/role-permissions/"):
		if r.Method != http.MethodPut && r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

func (s *StubHandler) AdminServiceLevels(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/admin/api/v1/service-levels" || r.Method != http.MethodGet {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, Ok(map[string]any{"items": []any{}, "total": 0}))
}

func (s *StubHandler) AdminCardOverview(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/admin/api/v1/card-overview" || r.Method != http.MethodGet {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// 对齐 GetCardOverviewResult
	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"items": []any{},
		"pagination": map[string]any{
			"size":      10,
			"page":      1,
			"count":     0,
			"total":     0,
			"sort":      "",
			"direction": 0,
		},
	}))
}

func (s *StubHandler) AdminAddresses(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/admin/api/v1/addresses":
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, Ok(map[string]any{"items": []any{}, "total": 0}))
		case http.MethodPost:
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"address_id":   "stub-address",
				"tenant_id":    "",
				"address_name": "stub",
				"is_active":    true,
			}))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/addresses/"):
		path := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/addresses/")
		// allocate/*
		if strings.Contains(path, "/allocate/") {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
			return
		}
		id := path
		if strings.Contains(id, "/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"address_id":   id,
				"tenant_id":    "",
				"address_name": "stub-" + id,
				"is_active":    true,
			}))
		case http.MethodPut:
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"address_id":   id,
				"tenant_id":    "",
				"address_name": "stub-" + id,
				"is_active":    true,
			}))
		case http.MethodDelete:
			writeJSON(w, http.StatusOK, Ok[any](nil))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

func (s *StubHandler) AdminAlarm(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/admin/api/v1/alarm-cloud":
		if r.Method == http.MethodGet {
			if s != nil && s.DB != nil {
				// Get tenant_id from query or header
				tenantID := r.URL.Query().Get("tenant_id")
				if tenantID == "" || tenantID == "null" {
					tenantID = r.Header.Get("X-Tenant-Id")
				}
				// Normalize: empty string or "null" means use SystemTenantID
				if tenantID == "" || tenantID == "null" {
					tenantID = SystemTenantID()
				}
				// Query alarm_cloud table
				var offlineAlarm, lowBattery, deviceFailure sql.NullString
				var deviceAlarmsRaw, conditionsRaw, notificationRulesRaw []byte
				var found bool
				err := s.DB.QueryRowContext(
					r.Context(),
					`SELECT OfflineAlarm, LowBattery, DeviceFailure, device_alarms, conditions, notification_rules
					 FROM alarm_cloud WHERE tenant_id = $1`,
					tenantID,
				).Scan(&offlineAlarm, &lowBattery, &deviceFailure, &deviceAlarmsRaw, &conditionsRaw, &notificationRulesRaw)
				if err == nil {
					found = true
				} else if err == sql.ErrNoRows {
					// If not found, try SystemTenantID as fallback
					if tenantID != SystemTenantID() {
						err = s.DB.QueryRowContext(
							r.Context(),
							`SELECT OfflineAlarm, LowBattery, DeviceFailure, device_alarms, conditions, notification_rules
							 FROM alarm_cloud WHERE tenant_id = $1`,
							SystemTenantID(),
						).Scan(&offlineAlarm, &lowBattery, &deviceFailure, &deviceAlarmsRaw, &conditionsRaw, &notificationRulesRaw)
						if err == nil {
							found = true
							tenantID = SystemTenantID() // Update tenant_id to reflect the actual source
						}
					}
				}
				if !found {
					// Return empty config if not found (should not happen if init script was run)
					writeJSON(w, http.StatusOK, Ok(map[string]any{
						"tenant_id":          tenantID,
						"OfflineAlarm":       nil,
						"LowBattery":         nil,
						"DeviceFailure":      nil,
						"device_alarms":      map[string]any{},
						"conditions":         nil,
						"notification_rules": nil,
					}))
					return
				}
				var deviceAlarms map[string]any
				if len(deviceAlarmsRaw) > 0 {
					_ = json.Unmarshal(deviceAlarmsRaw, &deviceAlarms)
				} else {
					deviceAlarms = map[string]any{}
				}
				var conditions any
				if len(conditionsRaw) > 0 {
					_ = json.Unmarshal(conditionsRaw, &conditions)
				}
				var notificationRules any
				if len(notificationRulesRaw) > 0 {
					_ = json.Unmarshal(notificationRulesRaw, &notificationRules)
				}
				result := map[string]any{
					"tenant_id":     tenantID,
					"device_alarms": deviceAlarms,
				}
				if offlineAlarm.Valid {
					result["OfflineAlarm"] = offlineAlarm.String
				}
				if lowBattery.Valid {
					result["LowBattery"] = lowBattery.String
				}
				if deviceFailure.Valid {
					result["DeviceFailure"] = deviceFailure.String
				}
				if conditions != nil {
					result["conditions"] = conditions
				}
				if notificationRules != nil {
					result["notification_rules"] = notificationRules
				}
				writeJSON(w, http.StatusOK, Ok(result))
				return
			}
			// No DB: return stub
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"tenant_id":     nil,
				"device_alarms": map[string]any{},
			}))
			return
		}
		if r.Method == http.MethodPut {
			if s != nil && s.DB != nil {
				var payload map[string]any
				if err := readBodyJSON(r, 1<<20, &payload); err != nil {
					writeJSON(w, http.StatusOK, Fail("invalid body"))
					return
				}
				// Get tenant_id from payload or header
				tenantID, _ := payload["tenant_id"].(string)
				if tenantID == "" {
					tenantID = r.Header.Get("X-Tenant-Id")
				}
				if tenantID == "" {
					tenantID = SystemTenantID()
				}
				// Parse fields
				var offlineAlarm, lowBattery, deviceFailure sql.NullString
				if val, ok := payload["OfflineAlarm"].(string); ok && val != "" {
					offlineAlarm = sql.NullString{String: val, Valid: true}
				}
				if val, ok := payload["LowBattery"].(string); ok && val != "" {
					lowBattery = sql.NullString{String: val, Valid: true}
				}
				if val, ok := payload["DeviceFailure"].(string); ok && val != "" {
					deviceFailure = sql.NullString{String: val, Valid: true}
				}
				var deviceAlarmsJSON []byte
				if val, ok := payload["device_alarms"].(map[string]any); ok && val != nil {
					deviceAlarmsJSON, _ = json.Marshal(val)
				} else {
					deviceAlarmsJSON = []byte("{}")
				}
				var conditionsJSON []byte
				if val, ok := payload["conditions"]; ok && val != nil {
					conditionsJSON, _ = json.Marshal(val)
				}
				var notificationRulesJSON []byte
				if val, ok := payload["notification_rules"]; ok && val != nil {
					notificationRulesJSON, _ = json.Marshal(val)
				}
				// Upsert: INSERT ... ON CONFLICT DO UPDATE
				_, err := s.DB.ExecContext(
					r.Context(),
					`INSERT INTO alarm_cloud (tenant_id, OfflineAlarm, LowBattery, DeviceFailure, device_alarms, conditions, notification_rules)
					 VALUES ($1, $2, $3, $4, $5, $6, $7)
					 ON CONFLICT (tenant_id) DO UPDATE SET
					   OfflineAlarm = EXCLUDED.OfflineAlarm,
					   LowBattery = EXCLUDED.LowBattery,
					   DeviceFailure = EXCLUDED.DeviceFailure,
					   device_alarms = EXCLUDED.device_alarms,
					   conditions = EXCLUDED.conditions,
					   notification_rules = EXCLUDED.notification_rules`,
					tenantID, offlineAlarm, lowBattery, deviceFailure, deviceAlarmsJSON, conditionsJSON, notificationRulesJSON,
				)
				if err != nil {
					writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to update alarm_cloud: %v", err)))
					return
				}
				// Return updated config
				var resultOfflineAlarm, resultLowBattery, resultDeviceFailure sql.NullString
				var resultDeviceAlarmsRaw, resultConditionsRaw, resultNotificationRulesRaw []byte
				_ = s.DB.QueryRowContext(
					r.Context(),
					`SELECT OfflineAlarm, LowBattery, DeviceFailure, device_alarms, conditions, notification_rules
					 FROM alarm_cloud WHERE tenant_id = $1`,
					tenantID,
				).Scan(&resultOfflineAlarm, &resultLowBattery, &resultDeviceFailure, &resultDeviceAlarmsRaw, &resultConditionsRaw, &resultNotificationRulesRaw)
				var resultDeviceAlarms map[string]any
				if len(resultDeviceAlarmsRaw) > 0 {
					_ = json.Unmarshal(resultDeviceAlarmsRaw, &resultDeviceAlarms)
				} else {
					resultDeviceAlarms = map[string]any{}
				}
				var resultConditions any
				if len(resultConditionsRaw) > 0 {
					_ = json.Unmarshal(resultConditionsRaw, &resultConditions)
				}
				var resultNotificationRules any
				if len(resultNotificationRulesRaw) > 0 {
					_ = json.Unmarshal(resultNotificationRulesRaw, &resultNotificationRules)
				}
				result := map[string]any{
					"tenant_id":     tenantID,
					"device_alarms": resultDeviceAlarms,
				}
				if resultOfflineAlarm.Valid {
					result["OfflineAlarm"] = resultOfflineAlarm.String
				}
				if resultLowBattery.Valid {
					result["LowBattery"] = resultLowBattery.String
				}
				if resultDeviceFailure.Valid {
					result["DeviceFailure"] = resultDeviceFailure.String
				}
				if resultConditions != nil {
					result["conditions"] = resultConditions
				}
				if resultNotificationRules != nil {
					result["notification_rules"] = resultNotificationRules
				}
				writeJSON(w, http.StatusOK, Ok(result))
				return
			}
			// No DB: return stub
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"tenant_id":     nil,
				"device_alarms": map[string]any{},
			}))
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	case r.URL.Path == "/admin/api/v1/alarm-events":
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// 对齐 GetAlarmEventsResult
		writeJSON(w, http.StatusOK, Ok(map[string]any{
			"items": []any{},
			"pagination": map[string]any{
				"size":  10,
				"page":  1,
				"count": 0,
				"total": 0,
			},
		}))
		return
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/alarm-events/") && strings.HasSuffix(r.URL.Path, "/handle"):
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

// --- settings/api/v1 ---

func (s *StubHandler) SettingsMonitor(w http.ResponseWriter, r *http.Request) {
	// /settings/api/v1/monitor/sleepace/:deviceId
	// /settings/api/v1/monitor/radar/:deviceId
	if strings.HasPrefix(r.URL.Path, "/settings/api/v1/monitor/sleepace/") {
		if r.Method == http.MethodGet || r.Method == http.MethodPut {
			// 返回 SleepaceMonitorSettings（字段较多，给默认值即可）
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"left_bed_start_hour": 0, "left_bed_start_minute": 0,
				"left_bed_end_hour": 0, "left_bed_end_minute": 0,
				"left_bed_duration": 0, "left_bed_alarm_level": "disabled",
				"min_heart_rate": 0, "heart_rate_slow_duration": 0, "heart_rate_slow_alarm_level": "disabled",
				"max_heart_rate": 0, "heart_rate_fast_duration": 0, "heart_rate_fast_alarm_level": "disabled",
				"min_breath_rate": 0, "breath_rate_slow_duration": 0, "breath_rate_slow_alarm_level": "disabled",
				"max_breath_rate": 0, "breath_rate_fast_duration": 0, "breath_rate_fast_alarm_level": "disabled",
				"breath_pause_duration": 0, "breath_pause_alarm_level": "disabled",
				"body_move_duration": 0, "body_move_alarm_level": "disabled",
				"nobody_move_duration": 0, "nobody_move_alarm_level": "disabled",
				"no_turn_over_duration": 0, "no_turn_over_alarm_level": "disabled",
				"situp_alarm_level": "disabled",
				"onbed_duration":    0, "onbed_alarm_level": "disabled",
				"fall_alarm_level": "disabled",
			}))
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/settings/api/v1/monitor/radar/") {
		if r.Method == http.MethodGet || r.Method == http.MethodPut {
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"radar_function_mode":     0,
				"suspected_fall_duration": 0, "fall_alarm_level": "disabled",
				"posture_detection_alarm_level": "disabled",
				"sitting_on_ground_duration":    0, "sitting_on_ground_alarm_level": "disabled",
				"stay_detection_duration": 0, "stay_alarm_level": "disabled",
				"leave_detection_start_hour": 0, "leave_detection_start_minute": 0,
				"leave_detection_end_hour": 0, "leave_detection_end_minute": 0,
				"leave_detection_duration": 0, "leave_alarm_level": "disabled",
				"lower_heart_rate": 0, "heart_rate_slow_alarm_level": "disabled",
				"upper_heart_rate": 0, "heart_rate_fast_alarm_level": "disabled",
				"lower_breath_rate": 0, "breath_rate_slow_alarm_level": "disabled",
				"upper_breath_rate": 0, "breath_rate_fast_alarm_level": "disabled",
				"breath_pause_alarm_level": "disabled",
				"weak_vital_duration":      0, "weak_vital_sensitivity": 0, "weak_vital_alarm_level": "disabled",
				"inactivity_alarm_level": "disabled",
			}))
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

// --- sleepace/api/v1 ---

func (s *StubHandler) SleepaceReports(w http.ResponseWriter, r *http.Request) {
	// GET /sleepace/api/v1/sleepace/reports/:id
	// GET /sleepace/api/v1/sleepace/reports/:id/detail
	// GET /sleepace/api/v1/sleepace/reports/:id/dates
	if !strings.HasPrefix(r.URL.Path, "/sleepace/api/v1/sleepace/reports/") || r.Method != http.MethodGet {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if strings.HasSuffix(r.URL.Path, "/detail") {
		writeJSON(w, http.StatusOK, Ok(map[string]any{
			"id": 0, "deviceId": "", "deviceCode": "", "recordCount": 0,
			"startTime": 0, "endTime": 0, "date": 0, "stopMode": 0, "timeStep": 0, "timezone": 0,
			"report": "",
		}))
		return
	}
	if strings.HasSuffix(r.URL.Path, "/dates") {
		writeJSON(w, http.StatusOK, Ok([]int{}))
		return
	}
	// list
	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"items": []any{},
		"pagination": map[string]any{
			"size": 10, "page": 1, "count": 0, "sort": "", "direction": 0,
		},
	}))
}

// --- device/api/v1 ---

func (s *StubHandler) DeviceRelations(w http.ResponseWriter, r *http.Request) {
	// GET /device/api/v1/device/:id/relations
	if !strings.HasPrefix(r.URL.Path, "/device/api/v1/device/") || !strings.HasSuffix(r.URL.Path, "/relations") {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/device/api/v1/device/"), "/relations")
	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"deviceId": id, "deviceName": "stub", "deviceInternalCode": "", "deviceType": 0,
		"addressId": "", "addressName": "", "addressType": 0,
		"residents": []any{},
	}))
}

// --- auth/api/v1 ---

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
			if _, ok2 := req["accountPasswordHash"]; !ok2 {
				req["accountPasswordHash"] = p["accountPasswordHash"]
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
		accountPasswordHash, _ := req["accountPasswordHash"].(string)
		if accountPasswordHash == "" {
			accountPasswordHash = r.URL.Query().Get("accountPasswordHash")
		}

		accountHash = strings.TrimSpace(accountHash)
		accountPasswordHash = strings.TrimSpace(accountPasswordHash)
		if accountHash == "" || accountPasswordHash == "" {
			writeJSON(w, http.StatusOK, Fail("missing credentials"))
			return
		}

		normalizedUserType := strings.ToLower(strings.TrimSpace(userType))
		if normalizedUserType == "" {
			normalizedUserType = "staff"
		}

		// Prefer DB auth when available (AuthStore is in-memory and will be lost after restart).
		// Decode hashes once (DB stores BYTEA).
		var ah, aph []byte
		if s != nil && s.DB != nil {
			var err1, err2 error
			ah, err1 = hex.DecodeString(accountHash)
			aph, err2 = hex.DecodeString(accountPasswordHash)
			if err1 != nil || err2 != nil || len(ah) == 0 || len(aph) == 0 {
				writeJSON(w, http.StatusOK, Fail("invalid credentials"))
				return
			}
		}

		// If tenant_id is not provided, resolve it from DB by (accountHash, accountPasswordHash, userType).
		// owlFront behavior:
		// - 0 match: invalid credentials
		// - 1 match: auto-login into that tenant (tenant_id optional)
		// - >1 match: frontend must let user choose an institution
		if tenantID == "" && s != nil && s.DB != nil {
			var rows *sql.Rows
			var err error
			switch normalizedUserType {
			case "resident":
				// Residents can login via:
				// - resident_account_hash (userAccount)
				// - phone_hash / email_hash (personal id)
				// Family contacts can login via:
				// - phone_hash / email_hash (personal id)
				rows, err = s.DB.QueryContext(
					r.Context(),
					`SELECT DISTINCT x.tenant_id::text
					   FROM (
					         SELECT r.tenant_id
					           FROM residents r
					          WHERE (r.resident_account_hash = $1 OR r.phone_hash = $1 OR r.email_hash = $1)
					            AND r.password_hash = $2
					            AND COALESCE(r.status,'active') = 'active'
					         UNION
					         SELECT rc.tenant_id
					           FROM resident_contacts rc
					          WHERE (rc.phone_hash = $1 OR rc.email_hash = $1)
					            AND rc.password_hash = $2
					            AND COALESCE(rc.is_enabled,true) = true
					   ) x`,
					ah, aph,
				)
			default: // staff
				rows, err = s.DB.QueryContext(
					r.Context(),
					`SELECT DISTINCT u.tenant_id::text
					   FROM users u
					  WHERE u.user_account_hash = $1
					    AND u.password_hash = $2
					    AND COALESCE(u.status,'active') = 'active'`,
					ah, aph,
				)
			}
			if err != nil {
				writeJSON(w, http.StatusOK, Fail("failed to resolve tenant"))
				return
			}
			defer rows.Close()
			var tids []string
			for rows.Next() {
				var tid string
				if err := rows.Scan(&tid); err == nil && tid != "" {
					tids = append(tids, tid)
				}
			}
			if len(tids) == 0 {
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

		if s != nil && s.DB != nil {
			if tenantID == "" {
				writeJSON(w, http.StatusOK, Fail("tenant_id is required"))
				return
			}
			var status string
			switch normalizedUserType {
			case "resident":
				// 1) Try resident login
				err := s.DB.QueryRowContext(
					r.Context(),
					`SELECT r.resident_id::text,
					        r.resident_account,
					        COALESCE(r.nickname,''),
					        r.role,
					        COALESCE(r.status,'active'),
					        COALESCE(t.tenant_name,''),
					        COALESCE(t.domain,'')
					   FROM residents r
					   JOIN tenants t ON t.tenant_id = r.tenant_id
					  WHERE r.tenant_id = $1
					    AND (r.resident_account_hash = $2 OR r.phone_hash = $2 OR r.email_hash = $2)
					    AND r.password_hash = $3
					  LIMIT 1`,
					tenantID, ah, aph,
				).Scan(&userID, &userAccount, &nickName, &role, &status, &tenantName, &domain)
				if err != nil {
					// 2) Try family contact login (phone/email only)
					var enabled bool
					var first, last string
					err2 := s.DB.QueryRowContext(
						r.Context(),
						`SELECT rc.contact_id::text,
						        COALESCE(rc.contact_first_name,''),
						        COALESCE(rc.contact_last_name,''),
						        rc.role,
						        COALESCE(rc.is_enabled,true),
						        COALESCE(t.tenant_name,''),
						        COALESCE(t.domain,'')
						   FROM resident_contacts rc
						   JOIN tenants t ON t.tenant_id = rc.tenant_id
						  WHERE rc.tenant_id = $1
						    AND (rc.phone_hash = $2 OR rc.email_hash = $2)
						    AND rc.password_hash = $3
						  LIMIT 1`,
						tenantID, ah, aph,
					).Scan(&userID, &first, &last, &role, &enabled, &tenantName, &domain)
					if err2 != nil {
						writeJSON(w, http.StatusOK, Fail("invalid credentials"))
						return
					}
					if !enabled {
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
				if status != "active" {
					writeJSON(w, http.StatusOK, Fail("user is not active"))
					return
				}
			default: // staff
				err := s.DB.QueryRowContext(
					r.Context(),
					`SELECT u.user_id::text,
					        u.user_account,
					        COALESCE(u.nickname,''),
					        u.role,
					        COALESCE(u.status,'active'),
					        COALESCE(t.tenant_name,''),
					        COALESCE(t.domain,'')
					   FROM users u
					   JOIN tenants t ON t.tenant_id = u.tenant_id
					  WHERE u.tenant_id = $1
					    AND u.user_account_hash = $2
					    AND u.password_hash = $3
					  LIMIT 1`,
					tenantID, ah, aph,
				).Scan(&userID, &userAccount, &nickName, &role, &status, &tenantName, &domain)
				if err != nil {
					writeJSON(w, http.StatusOK, Fail("invalid credentials"))
					return
				}
				if status != "active" {
					writeJSON(w, http.StatusOK, Fail("user is not active"))
					return
				}
			}
		} else if s != nil && s.AuthStore != nil && allowAuthStoreFallback() {
			// Fallback: in-memory auth
			if u, ok := s.AuthStore.FindUser(tenantID, accountHash, accountPasswordHash); ok {
				role = u.Role
				userID = u.UserID
				userAccount = u.UserAccount
				nickName = u.Role
			} else {
				writeJSON(w, http.StatusOK, Fail("invalid credentials"))
				return
			}
			// Resolve tenant name/domain for better UX
			if tenantID == SystemTenantID() {
				tenantName = "System"
				domain = "system.local"
			} else if s.Tenants != nil {
				ts, _, err := s.Tenants.ListTenants(r.Context(), "", 1, 1000)
				if err == nil {
					for _, t := range ts {
						if t.TenantID == tenantID {
							tenantName = t.TenantName
							domain = t.Domain
							break
						}
					}
				}
			}
		}
		if (s == nil || s.DB == nil) && !allowAuthStoreFallback() {
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
		writeJSON(w, http.StatusOK, Ok(map[string]any{
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
		}))
		return
	case "/auth/api/v1/institutions/search":
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// 返回 Institution[]（authModel.ts: {id,name,domain?}）
		accountHash := r.URL.Query().Get("accountHash")
		accountPasswordHash := r.URL.Query().Get("accountPasswordHash")
		userType := strings.TrimSpace(r.URL.Query().Get("userType"))
		normalizedUserType := strings.ToLower(strings.TrimSpace(userType))
		if normalizedUserType == "" {
			normalizedUserType = "staff"
		}
		accountHash = strings.TrimSpace(accountHash)
		accountPasswordHash = strings.TrimSpace(accountPasswordHash)

		// Prefer DB lookup when available.
		if s != nil && s.DB != nil && accountHash != "" && accountPasswordHash != "" {
			ah, err1 := hex.DecodeString(accountHash)
			aph, err2 := hex.DecodeString(accountPasswordHash)
			if err1 != nil || err2 != nil || len(ah) == 0 || len(aph) == 0 {
				writeJSON(w, http.StatusOK, Ok([]any{}))
				return
			}
			var rows *sql.Rows
			var err error
			switch normalizedUserType {
			case "resident":
				rows, err = s.DB.QueryContext(
					r.Context(),
					`SELECT DISTINCT t.tenant_id::text, t.tenant_name, COALESCE(t.domain,'')
					   FROM (
					         SELECT r.tenant_id
					           FROM residents r
					          WHERE (r.resident_account_hash = $1 OR r.phone_hash = $1 OR r.email_hash = $1)
					            AND r.password_hash = $2
					            AND COALESCE(r.status,'active') = 'active'
					         UNION
					         SELECT rc.tenant_id
					           FROM resident_contacts rc
					          WHERE (rc.phone_hash = $1 OR rc.email_hash = $1)
					            AND rc.password_hash = $2
					            AND COALESCE(rc.is_enabled,true) = true
					   ) x
					   JOIN tenants t ON t.tenant_id = x.tenant_id
					  ORDER BY t.tenant_name ASC`,
					ah, aph,
				)
			default: // staff
				rows, err = s.DB.QueryContext(
					r.Context(),
					`SELECT DISTINCT t.tenant_id::text, t.tenant_name, COALESCE(t.domain,'')
					   FROM users u
					   JOIN tenants t ON t.tenant_id = u.tenant_id
					  WHERE u.user_account_hash = $1
					    AND u.password_hash = $2
					    AND COALESCE(u.status,'active') = 'active'
					  ORDER BY t.tenant_name ASC`,
					ah, aph,
				)
			}
			if err != nil {
				writeJSON(w, http.StatusOK, Ok([]any{}))
				return
			}
			defer rows.Close()
			items := []any{}
			for rows.Next() {
				var id, name, dom string
				if err := rows.Scan(&id, &name, &dom); err != nil {
					continue
				}
				items = append(items, map[string]any{"id": id, "name": name, "domain": dom})
			}
			writeJSON(w, http.StatusOK, Ok(items))
			return
		}

		// If auth store fallback is explicitly enabled, only return tenants where this account exists.
		if s != nil && s.AuthStore != nil && allowAuthStoreFallback() && accountHash != "" {
			tenantIDs := s.AuthStore.TenantsForLogin(accountHash, accountPasswordHash)
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

		// No DB and fallback disabled: return empty list.
		if s == nil || s.DB == nil {
			writeJSON(w, http.StatusOK, Ok([]any{}))
			return
		}

		// Legacy fallback: return System + all active tenants
		items := []any{
			map[string]any{"id": SystemTenantID(), "name": "System", "domain": "system.local"},
		}
		if s != nil && s.Tenants != nil {
			ts, _, err := s.Tenants.ListTenants(r.Context(), "", 1, 1000)
			if err == nil {
				for _, t := range ts {
					if t.Status == "deleted" {
						continue
					}
					if t.TenantID == SystemTenantID() {
						continue
					}
					items = append(items, map[string]any{"id": t.TenantID, "name": t.TenantName, "domain": t.Domain})
				}
			}
		}
		writeJSON(w, http.StatusOK, Ok(items))
		return
	case "/auth/api/v1/forgot-password/send-code":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true, "message": "stub", "expired_at": time.Now().Add(5 * time.Minute).Unix()}))
		return
	case "/auth/api/v1/forgot-password/verify-code":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true, "message": "stub"}))
		return
	case "/auth/api/v1/forgot-password/reset":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true, "message": "stub"}))
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

// SystemTenantID is the fixed platform tenant id used for SystemAdmin (dev bootstrap).
func SystemTenantID() string {
	// IMPORTANT:
	// - Do NOT use 00000000-0000-0000-0000-000000000000 because owlRD uses it as a sentinel
	//   meaning "unassigned" (e.g. device_store.tenant_id).
	return "00000000-0000-0000-0000-000000000001"
}

// --- example ---

func (s *StubHandler) Example(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/api/v1/example/items" && r.Method == http.MethodGet:
		writeJSON(w, http.StatusOK, Ok(map[string]any{"items": []any{}, "total": 0}))
		return
	case strings.HasPrefix(r.URL.Path, "/api/v1/example/") && r.Method == http.MethodGet:
		writeJSON(w, http.StatusOK, Ok(map[string]any{}))
		return
	case r.URL.Path == "/api/v1/example/item" && r.Method == http.MethodPost:
		writeJSON(w, http.StatusOK, Ok(map[string]any{"id": "stub"}))
		return
	}
	w.WriteHeader(http.StatusNotFound)
}
