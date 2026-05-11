package httpapi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// CareTeamsHandler 管理 care_teams + care_team_members 表
//
// v2 schema：
//
//	care_teams (team_id UUID, tenant_id INET, team_code VARCHAR, team_name, team_kind, ...)
//	care_team_members (team_id, user_id, role_in_team, valid_from, valid_to ...)
//
// 范围：当前 tenant_id 下；不允许跨 tenant 操作。
type CareTeamsHandler struct {
	db     *sql.DB
	logger *zap.Logger
	base   *StubHandler
}

func NewCareTeamsHandler(db *sql.DB, logger *zap.Logger) *CareTeamsHandler {
	return &CareTeamsHandler{db: db, logger: logger, base: &StubHandler{}}
}

func (h *CareTeamsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")
	switch {
	case path == "/admin/api/v1/care-teams" && r.Method == http.MethodGet:
		h.list(w, r)
	case path == "/admin/api/v1/care-teams" && r.Method == http.MethodPost:
		h.create(w, r)
	case strings.HasPrefix(path, "/admin/api/v1/care-teams/") && strings.HasSuffix(path, "/members") && r.Method == http.MethodGet:
		h.listMembers(w, r, extractTeamIDFromMembersPath(path))
	case strings.HasPrefix(path, "/admin/api/v1/care-teams/") && strings.HasSuffix(path, "/members") && r.Method == http.MethodPost:
		h.addMember(w, r, extractTeamIDFromMembersPath(path))
	case strings.HasPrefix(path, "/admin/api/v1/care-teams/") && strings.Contains(path, "/members/") && r.Method == http.MethodDelete:
		teamID, userID := extractTeamUserFromMembersPath(path)
		h.removeMember(w, r, teamID, userID)
	case strings.HasPrefix(path, "/admin/api/v1/care-teams/") && r.Method == http.MethodPut:
		h.update(w, r, strings.TrimPrefix(path, "/admin/api/v1/care-teams/"))
	case strings.HasPrefix(path, "/admin/api/v1/care-teams/") && r.Method == http.MethodDelete:
		h.delete(w, r, strings.TrimPrefix(path, "/admin/api/v1/care-teams/"))
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// extractTeamIDFromMembersPath: /admin/api/v1/care-teams/{teamID}/members → teamID
func extractTeamIDFromMembersPath(path string) string {
	rest := strings.TrimPrefix(path, "/admin/api/v1/care-teams/")
	rest = strings.TrimSuffix(rest, "/members")
	return rest
}

// extractTeamUserFromMembersPath: /admin/api/v1/care-teams/{teamID}/members/{userID}
func extractTeamUserFromMembersPath(path string) (string, string) {
	rest := strings.TrimPrefix(path, "/admin/api/v1/care-teams/")
	parts := strings.Split(rest, "/members/")
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

// list — GET /admin/api/v1/care-teams
func (h *CareTeamsHandler) list(w http.ResponseWriter, r *http.Request) {
	tenantPrefix, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	rows, err := h.db.QueryContext(r.Context(), `
		SELECT t.team_id::text,
		       t.team_code,
		       t.team_name,
		       t.team_kind,
		       COALESCE(t.description,'') AS description,
		       COALESCE(t.color_hex,'')  AS color_hex,
		       t.is_active,
		       (SELECT COUNT(*) FROM care_team_members m WHERE m.team_id = t.team_id AND m.valid_to IS NULL) AS member_count,
		       COALESCE(
		         (SELECT string_agg(COALESCE(NULLIF(u.nickname,''), u.user_account), ',' ORDER BY u.user_account)
		            FROM care_team_members m
		            JOIN users u ON u.user_id = m.user_id
		           WHERE m.team_id = t.team_id AND m.valid_to IS NULL), ''
		       ) AS member_names
		  FROM care_teams t
		 WHERE t.tenant_id = $1::INET
		 ORDER BY t.team_kind, t.team_name`, tenantPrefix)
	if err != nil {
		h.logger.Error("list care_teams", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, code, name, kind, desc, color, memberNames string
		var active bool
		var memberCount int
		if err := rows.Scan(&id, &code, &name, &kind, &desc, &color, &active, &memberCount, &memberNames); err == nil {
			names := []string{}
			if memberNames != "" {
				for _, n := range strings.Split(memberNames, ",") {
					if n != "" {
						names = append(names, n)
					}
				}
			}
			out = append(out, map[string]any{
				"team_id":       id,
				"team_code":     code,
				"team_name":     name,
				"team_kind":     kind,
				"description":   desc,
				"color_hex":     color,
				"is_active":     active,
				"member_count":  memberCount,
				"member_names":  names,
			})
		}
	}
	writeJSON(w, http.StatusOK, Ok(map[string]any{"items": out, "total": len(out)}))
}

// create — POST /admin/api/v1/care-teams { team_code?, team_name, team_kind?, description?, color_hex? }
func (h *CareTeamsHandler) create(w http.ResponseWriter, r *http.Request) {
	tenantPrefix, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	var body struct {
		TeamCode    string `json:"team_code"`
		TeamName    string `json:"team_name"`
		TeamKind    string `json:"team_kind"`
		Description string `json:"description"`
		ColorHex    string `json:"color_hex"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid JSON"))
		return
	}
	body.TeamName = strings.TrimSpace(body.TeamName)
	if body.TeamName == "" {
		writeJSON(w, http.StatusOK, Fail("team_name is required"))
		return
	}
	if body.TeamCode == "" {
		// team_code 默认 = team_name 转 lowercase + spaces→_，避免 unique conflict 时返回明确错
		body.TeamCode = strings.ToLower(strings.ReplaceAll(body.TeamName, " ", "_"))
	}
	if body.TeamKind == "" {
		body.TeamKind = "nurse_station"
	}
	var newID string
	err := h.db.QueryRowContext(r.Context(), `
		INSERT INTO care_teams (tenant_id, team_code, team_name, team_kind, description, color_hex, is_active)
		VALUES ($1::INET, $2, $3, $4, $5, NULLIF($6,''), TRUE)
		RETURNING team_id::text`,
		tenantPrefix, body.TeamCode, body.TeamName, body.TeamKind, body.Description, body.ColorHex,
	).Scan(&newID)
	if err != nil {
		h.logger.Error("create care_team", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(map[string]any{"team_id": newID}))
}

// update — PUT /admin/api/v1/care-teams/{id} { team_name?, team_code?, description?, color_hex?, is_active? }
func (h *CareTeamsHandler) update(w http.ResponseWriter, r *http.Request, teamID string) {
	tenantPrefix, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid JSON"))
		return
	}
	sets := []string{}
	args := []any{teamID, tenantPrefix}
	argN := 3
	for _, f := range []string{"team_name", "team_code", "team_kind", "description", "color_hex"} {
		if v, exists := body[f]; exists {
			if v == nil {
				sets = append(sets, fmt.Sprintf("%s = NULL", f))
			} else if s, ok := v.(string); ok {
				sets = append(sets, fmt.Sprintf("%s = $%d", f, argN))
				args = append(args, s)
				argN++
			}
		}
	}
	if v, exists := body["is_active"]; exists {
		if b, ok := v.(bool); ok {
			sets = append(sets, fmt.Sprintf("is_active = $%d", argN))
			args = append(args, b)
			argN++
		}
	}
	if len(sets) == 0 {
		writeJSON(w, http.StatusOK, Ok(map[string]any{"updated": 0}))
		return
	}
	q := "UPDATE care_teams SET " + strings.Join(sets, ", ") + ", updated_at = NOW() WHERE team_id = $1::UUID AND tenant_id = $2::INET"
	res, err := h.db.ExecContext(r.Context(), q, args...)
	if err != nil {
		h.logger.Error("update care_team", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	n, _ := res.RowsAffected()
	writeJSON(w, http.StatusOK, Ok(map[string]any{"updated": n}))
}

// delete — DELETE /admin/api/v1/care-teams/{id}
//
//	实施：硬删 care_teams 行 + cascade 自动清 care_team_members（FK 已 ON DELETE CASCADE 验证一下）。
//	若想软删，改 SET is_active = FALSE 即可（保留历史 members 关联）。
func (h *CareTeamsHandler) delete(w http.ResponseWriter, r *http.Request, teamID string) {
	tenantPrefix, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(r.Context(),
		`DELETE FROM care_team_members WHERE team_id = $1::UUID`, teamID); err != nil {
		h.logger.Error("delete care_team_members", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	res, err := tx.ExecContext(r.Context(),
		`DELETE FROM care_teams WHERE team_id = $1::UUID AND tenant_id = $2::INET`,
		teamID, tenantPrefix)
	if err != nil {
		h.logger.Error("delete care_team", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	if err := tx.Commit(); err != nil {
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	n, _ := res.RowsAffected()
	writeJSON(w, http.StatusOK, Ok(map[string]any{"deleted": n}))
}

// listMembers — GET /admin/api/v1/care-teams/{id}/members
func (h *CareTeamsHandler) listMembers(w http.ResponseWriter, r *http.Request, teamID string) {
	tenantPrefix, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	rows, err := h.db.QueryContext(r.Context(), `
		SELECT u.user_id::text, u.user_account, COALESCE(u.nickname,'') AS nickname,
		       COALESCE(u.role,'') AS role, m.role_in_team
		  FROM care_team_members m
		  JOIN users u ON u.user_id = m.user_id
		  JOIN care_teams t ON t.team_id = m.team_id
		 WHERE m.team_id = $1::UUID
		   AND m.valid_to IS NULL
		   AND t.tenant_id = $2::INET
		 ORDER BY u.user_account`, teamID, tenantPrefix)
	if err != nil {
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var uid, uname, nick, role, roleInTeam string
		if err := rows.Scan(&uid, &uname, &nick, &role, &roleInTeam); err == nil {
			out = append(out, map[string]any{
				"user_id":      uid,
				"user_account":     uname,
				"nickname":     nick,
				"role":         role,
				"role_in_team": roleInTeam,
			})
		}
	}

	// 同时返回 tenant 下所有可加入的 user（未在 team 里的），方便 FE Add Member picker
	availRows, err := h.db.QueryContext(r.Context(), `
		SELECT u.user_id::text, u.user_account, COALESCE(u.nickname,'') AS nickname, COALESCE(u.role,'') AS role
		  FROM users u
		 WHERE u.tenant_id = $1::INET
		   AND u.status = 'active'
		   AND NOT EXISTS (
		     SELECT 1 FROM care_team_members m
		      WHERE m.team_id = $2::UUID AND m.user_id = u.user_id AND m.valid_to IS NULL
		   )
		 ORDER BY u.user_account`, tenantPrefix, teamID)
	available := []map[string]any{}
	if err == nil {
		defer availRows.Close()
		for availRows.Next() {
			var uid, uname, nick, role string
			if err := availRows.Scan(&uid, &uname, &nick, &role); err == nil {
				available = append(available, map[string]any{
					"user_id":  uid,
					"user_account": uname,
					"nickname": nick,
					"role":     role,
				})
			}
		}
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"members":   out,
		"available": available,
	}))
}

// addMember — POST /admin/api/v1/care-teams/{id}/members { user_id, role_in_team? }
func (h *CareTeamsHandler) addMember(w http.ResponseWriter, r *http.Request, teamID string) {
	tenantPrefix, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	var body struct {
		UserID     string `json:"user_id"`
		RoleInTeam string `json:"role_in_team"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid JSON"))
		return
	}
	if body.UserID == "" {
		writeJSON(w, http.StatusOK, Fail("user_id is required"))
		return
	}
	if body.RoleInTeam == "" {
		body.RoleInTeam = "member"
	}
	// 验证 team 属于当前 tenant
	var dummy string
	if err := h.db.QueryRowContext(r.Context(),
		`SELECT team_id::text FROM care_teams WHERE team_id = $1::UUID AND tenant_id = $2::INET`,
		teamID, tenantPrefix).Scan(&dummy); err != nil {
		writeJSON(w, http.StatusOK, Fail("care team not found in this tenant"))
		return
	}
	// 验证 user 属于当前 tenant
	if err := h.db.QueryRowContext(r.Context(),
		`SELECT user_id::text FROM users WHERE user_id = $1::UUID AND tenant_id = $2::INET`,
		body.UserID, tenantPrefix).Scan(&dummy); err != nil {
		writeJSON(w, http.StatusOK, Fail("user not found in this tenant"))
		return
	}
	_, err := h.db.ExecContext(r.Context(),
		`INSERT INTO care_team_members (team_id, user_id, role_in_team) VALUES ($1::UUID, $2::UUID, $3)
		 ON CONFLICT (team_id, user_id, valid_from) DO NOTHING`,
		teamID, body.UserID, body.RoleInTeam)
	if err != nil {
		h.logger.Error("add member", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(map[string]any{"added": 1}))
}

// removeMember — DELETE /admin/api/v1/care-teams/{teamID}/members/{userID}
func (h *CareTeamsHandler) removeMember(w http.ResponseWriter, r *http.Request, teamID, userID string) {
	tenantPrefix, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	res, err := h.db.ExecContext(r.Context(), `
		DELETE FROM care_team_members
		 WHERE team_id = $1::UUID AND user_id = $2::UUID
		   AND team_id IN (SELECT team_id FROM care_teams WHERE tenant_id = $3::INET)`,
		teamID, userID, tenantPrefix)
	if err != nil {
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	n, _ := res.RowsAffected()
	writeJSON(w, http.StatusOK, Ok(map[string]any{"deleted": n}))
}
