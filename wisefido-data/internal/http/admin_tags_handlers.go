package httpapi

import (
	"fmt"
	"net/http"
	"strings"
)

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
			writeJSON(w, http.StatusOK, Fail("database not available"))
			return
		case http.MethodGet:
			if s != nil && s.DB != nil {
				tenantID, ok := s.tenantIDFromReq(w, r)
				if !ok {
					return
				}
				tagTypeFilter := strings.TrimSpace(r.URL.Query().Get("tag_type"))
				includeSystem := r.URL.Query().Get("include_system_tag_types") != "false"

				// 使用 get_tags_for_tenant 函数（不包含 tag_objects，因为该字段已删除）
				// tag 的使用情况需要从源表查询：residents.family_tag, units.branch_tag, units.area_tag, users.tags
				var q string
				var queryArgs []any
				
				if tagTypeFilter != "" {
					// 指定 tag_type
					q = `SELECT tag_id::text, tenant_id::text, tag_type, tag_name
					      FROM get_tags_for_tenant($1, $2)
					      ORDER BY tag_type, tag_name`
					queryArgs = []any{tenantID, tagTypeFilter}
				} else {
					// 不指定 tag_type，使用函数返回所有类型
					q = `SELECT tag_id::text, tenant_id::text, tag_type, tag_name
					      FROM get_tags_for_tenant($1, NULL)
					      ORDER BY tag_type, tag_name`
					queryArgs = []any{tenantID}
					
					// 如果需要排除系统预定义类型，在应用层过滤
					if !includeSystem {
						// 先查询所有，然后在应用层过滤
					}
				}

				rows, err := s.DB.QueryContext(r.Context(), q, queryArgs...)
				if err != nil {
					fmt.Printf("[AdminTags] Query error: %v\n", err)
					writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to list tags: %v", err)))
					return
				}
				defer rows.Close()

				items := []any{}
				for rows.Next() {
					var tagID, tid, tagType, tagName string
					if err := rows.Scan(&tagID, &tid, &tagType, &tagName); err != nil {
						fmt.Printf("[AdminTags] Scan error: %v\n", err)
						writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to scan tag: %v", err)))
						return
					}
					
					// 如果需要排除系统预定义类型，在应用层过滤
					if !includeSystem && tagTypeFilter == "" {
						if tagType == "branch_tag" || tagType == "family_tag" || tagType == "area_tag" {
							continue
						}
					}
					
					// tag_objects 字段已删除，不再返回
					// 如果需要 tag 的使用情况，需要从源表查询：
					// - family_tag: SELECT COUNT(*) FROM residents WHERE family_tag = tag_name
					// - branch_tag: SELECT COUNT(*) FROM units WHERE branch_tag = tag_name
					// - area_tag: SELECT COUNT(*) FROM units WHERE area_tag = tag_name
					// - user_tag: SELECT COUNT(*) FROM users WHERE tags ? tag_name
					items = append(items, map[string]any{
						"tag_id":    tagID,
						"tenant_id": tid,
						"tag_type":  tagType,
						"tag_name":  tagName,
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
			writeJSON(w, http.StatusOK, Fail("database not available"))
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

			// Get tag_name and tag_type for user_tag sync (needed for syncing users.tags)
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

				// If adding a user to user_tag, also add the tag to user's tags JSONB
				if objectType == "user" && tagType == "user_tag" {
					// Use jsonb concatenation to add the tag to array (avoid duplicates)
					_, err = s.DB.ExecContext(
						r.Context(),
						`UPDATE users 
						 SET tags = COALESCE(tags, '[]'::jsonb) || jsonb_build_array($1::text)
						 WHERE user_id = $2::uuid
						   AND (tags IS NULL OR NOT (tags ? $1))`,
						tagName, objectID,
					)
					if err != nil {
						fmt.Printf("[AdminTags] Failed to add tag to user's tags: %v\n", err)
						// Don't fail the whole operation, just log the error
					}
				}
			}

				writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
				return
			}
			writeJSON(w, http.StatusOK, Fail("database not available"))
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

						// If removing a resident from family_tag, also clear the resident's family_tag field
						if objectType == "resident" && tagType == "family_tag" {
							_, err = s.DB.ExecContext(
								r.Context(),
								`UPDATE residents 
								 SET family_tag = NULL
								 WHERE resident_id = $1::uuid
								   AND family_tag = $2`,
								objectID, tagName,
							)
							if err != nil {
								fmt.Printf("[AdminTags] Failed to clear family_tag from resident: %v\n", err)
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

						// If removing a resident from family_tag, also clear the resident's family_tag field
						if objectType == "resident" && tagType == "family_tag" {
							_, err = s.DB.ExecContext(
								r.Context(),
								`UPDATE residents 
								 SET family_tag = NULL
								 WHERE resident_id = $1::uuid
								   AND family_tag = $2`,
								objectID, tagName,
							)
							if err != nil {
								fmt.Printf("[AdminTags] Failed to clear family_tag from resident: %v\n", err)
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
			writeJSON(w, http.StatusOK, Fail("database not available"))
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	case r.URL.Path == "/admin/api/v1/tags/types":
		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
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
			tagType, _ := payload["tag_type"].(string)
			if tagType == "" {
				writeJSON(w, http.StatusOK, Fail("tag_type is required"))
				return
			}
			// Delete all tags of this type for the tenant
			_, err := s.DB.ExecContext(
				r.Context(),
				`DELETE FROM tags_catalog WHERE tenant_id = $1 AND tag_type = $2`,
				tenantID, tagType,
			)
			if err != nil {
				writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to delete tag type: %v", err)))
				return
			}
			writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
			return
		}
		writeJSON(w, http.StatusOK, Fail("database not available"))
		return
	case r.URL.Path == "/admin/api/v1/tags/for-object":
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if s != nil && s.DB != nil {
			tenantID, ok := s.tenantIDFromReq(w, r)
			if !ok {
				return
			}
			objectType := r.URL.Query().Get("object_type")
			objectID := r.URL.Query().Get("object_id")
			if objectType == "" || objectID == "" {
				writeJSON(w, http.StatusOK, Fail("object_type and object_id are required"))
				return
			}
			// Query tags for the object
			rows, err := s.DB.QueryContext(
				r.Context(),
				`SELECT DISTINCT tc.tag_id::text, tc.tag_name, tc.tag_type
				 FROM tags_catalog tc
				 WHERE tc.tenant_id = $1
				   AND tc.tag_objects->$2->>'object_id' = $3`,
				tenantID, objectType, objectID,
			)
			if err != nil {
				writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to get tags: %v", err)))
				return
			}
			defer rows.Close()
			items := []any{}
			for rows.Next() {
				var tagID, tagName, tagType string
				if err := rows.Scan(&tagID, &tagName, &tagType); err != nil {
					continue
				}
				items = append(items, map[string]any{
					"tag_id":   tagID,
					"tag_name": tagName,
					"tag_type": tagType,
				})
			}
			writeJSON(w, http.StatusOK, Ok(items))
			return
		}
		writeJSON(w, http.StatusOK, Fail("database not available"))
		return
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/tags/"):
		// PUT /admin/api/v1/tags/:id
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if s != nil && s.DB != nil {
			tenantID, ok := s.tenantIDFromReq(w, r)
			if !ok {
				return
			}
			tagID := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/tags/")
			if tagID == "" || strings.Contains(tagID, "/") {
				w.WriteHeader(http.StatusNotFound)
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
			_, err := s.DB.ExecContext(
				r.Context(),
				`UPDATE tags_catalog SET tag_name = $3 WHERE tenant_id = $1 AND tag_id = $2`,
				tenantID, tagID, tagName,
			)
			if err != nil {
				writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to update tag: %v", err)))
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
