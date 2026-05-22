package httpapi

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"
	"wisefido-data/internal/service"

	"go.uber.org/zap"
)

// RoundsHandler 巡房记录 HTTP
type RoundsHandler struct {
	svc    *service.RoundsService
	db     *sql.DB
	logger *zap.Logger
}

func NewRoundsHandler(svc *service.RoundsService, db *sql.DB, logger *zap.Logger) *RoundsHandler {
	return &RoundsHandler{svc: svc, db: db, logger: logger}
}

// CreateRoundRequest 与 service.CreateRoundRequest 一致，供 JSON 绑定
type CreateRoundRequest struct {
	StartedAt      *time.Time                `json:"started_at"`
	Notes          string                    `json:"notes"`
	ItemsChecked   []domain.RoundItemChecked `json:"items_checked"`
	ReportSnapshot json.RawMessage           `json:"report_snapshot,omitempty"`
}

// ServeHTTP 路由：
//   POST /rounds        创建一轮
//   GET  /rounds        列表（支持 ?user_id=&status=&...&page=&size=）
//   GET  /rounds/{id}   单条详情（RoundSafetyReport 用）
func (h *RoundsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	base := "/data/api/v1/data/vital-focus/rounds"
	if path == base || path == base+"/" {
		switch r.Method {
		case http.MethodPost:
			h.createRound(w, r)
		case http.MethodGet:
			h.listRounds(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}
	if strings.HasPrefix(path, base+"/") {
		roundID := strings.TrimPrefix(path, base+"/")
		if roundID == "" || strings.Contains(roundID, "/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		h.getRound(w, r, roundID)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

func roundToWire(rd *domain.Round) map[string]interface{} {
	row := map[string]interface{}{
		"round_id":   rd.RoundID,
		"round_type": rd.RoundType,
		"unit_id":    rd.UnitID,
		"user_id":    rd.UserID,
		"notes":      rd.Notes,
		"status":     rd.Status,
	}
	if rd.StartedAt != nil {
		row["started_at"] = rd.StartedAt.Format(time.RFC3339)
	}
	if rd.EndedAt != nil {
		row["ended_at"] = rd.EndedAt.Format(time.RFC3339)
	}
	if rd.ExecutorDisplay != "" {
		row["executor_display"] = rd.ExecutorDisplay
	}
	if rd.FacilityName != "" {
		row["facility_name"] = rd.FacilityName
	}
	if rd.UnitName != "" {
		row["unit_name"] = rd.UnitName
	}
	// snapshot JSONB: 拆 items_checked 到顶层（FE 列表/汇总用），
	// 其余字段（rows / schema_version / facility_license）整体出在 report_snapshot 给 detail 视图。
	if len(rd.Snapshot) > 0 {
		var snap map[string]json.RawMessage
		if json.Unmarshal(rd.Snapshot, &snap) == nil {
			if raw, ok := snap["items_checked"]; ok {
				var arr []domain.RoundItemChecked
				if json.Unmarshal(raw, &arr) == nil && len(arr) > 0 {
					row["items_checked"] = arr
				}
			}
			delete(snap, "items_checked")
			if len(snap) > 0 {
				row["report_snapshot"] = snap
			}
		}
	}
	return row
}

func (h *RoundsHandler) createRound(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, tenantID, _, _, ok := service.MustSession(ctx)
	if !ok || userID == "" || tenantID == "" {
		writeJSON(w, http.StatusUnauthorized, Fail("missing or invalid authorization"))
		return
	}
	var body CreateRoundRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, Fail("invalid request body"))
		return
	}
	req := &service.CreateRoundRequest{
		StartedAt:      body.StartedAt,
		Notes:          body.Notes,
		ItemsChecked:   body.ItemsChecked,
		ReportSnapshot: body.ReportSnapshot,
	}
	roundID, err := h.svc.CreateRound(ctx, userID, req)
	if err != nil {
		h.logger.Warn("CreateRound failed", zap.String("user_id", userID), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(map[string]interface{}{"round_id": roundID}))
}

func (h *RoundsHandler) listRounds(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUserID, tenantID, _, _, ok := service.MustSession(ctx)
	if !ok || tenantID == "" {
		writeJSON(w, http.StatusUnauthorized, Fail("missing or invalid authorization"))
		return
	}
	filters := &repository.RoundFilters{}
	if currentUserID != "" && h.db != nil {
		var bp sql.NullString
		_ = h.db.QueryRowContext(ctx, `
			SELECT host(branch_prefix) || '/56' FROM user_branches
			 WHERE user_id = $1::UUID AND is_primary = TRUE AND valid_to IS NULL LIMIT 1`,
			currentUserID,
		).Scan(&bp)
		if bp.Valid && bp.String != "" {
			filters.BranchPrefix = bp.String
		}
	}
	if v := r.URL.Query().Get("user_id"); v != "" {
		filters.UserID = v
	}
	if v := r.URL.Query().Get("round_type"); v != "" {
		filters.RoundType = v
	}
	if v := r.URL.Query().Get("status"); v != "" {
		filters.Status = v
	}
	if v := r.URL.Query().Get("start_time"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filters.StartTime = &t
		}
	}
	if v := r.URL.Query().Get("end_time"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filters.EndTime = &t
		}
	}
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	size := 20
	if s := r.URL.Query().Get("size"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			size = n
			if size > 500 {
				size = 500
			}
		}
	}
	rounds, total, err := h.svc.ListRounds(ctx, tenantID, filters, page, size)
	if err != nil {
		h.logger.Warn("ListRounds failed", zap.String("tenant_id", tenantID), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	items := make([]map[string]interface{}, 0, len(rounds))
	for _, rd := range rounds {
		items = append(items, roundToWire(rd))
	}
	writeJSON(w, http.StatusOK, Ok(map[string]interface{}{
		"items": items,
		"total": total,
		"page":  page,
		"size":  size,
	}))
}

func (h *RoundsHandler) getRound(w http.ResponseWriter, r *http.Request, roundID string) {
	ctx := r.Context()
	_, tenantID, _, _, ok := service.MustSession(ctx)
	if !ok || tenantID == "" {
		writeJSON(w, http.StatusUnauthorized, Fail("missing or invalid authorization"))
		return
	}
	rd, err := h.svc.GetRound(ctx, tenantID, roundID)
	if err != nil {
		h.logger.Warn("GetRound failed", zap.String("round_id", roundID), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(roundToWire(rd)))
}
