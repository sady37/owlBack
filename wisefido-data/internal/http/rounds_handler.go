package httpapi

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"
	"wisefido-data/internal/service"

	"go.uber.org/zap"
)

// RoundsHandler 巡房记录 HTTP（Automatic Rounds 完成、审计列表）
type RoundsHandler struct {
	svc    *service.RoundsService
	db     *sql.DB // 用于 Phase 3 Current Branch scope 查询
	logger *zap.Logger
}

// NewRoundsHandler 创建 handler
func NewRoundsHandler(svc *service.RoundsService, db *sql.DB, logger *zap.Logger) *RoundsHandler {
	return &RoundsHandler{svc: svc, db: db, logger: logger}
}

// CreateRoundRequest 与 service.CreateRoundRequest 一致，供 JSON 绑定
type CreateRoundRequest struct {
	StartedAt    *time.Time                 `json:"started_at"`
	ItemsChecked []domain.RoundItemChecked `json:"items_checked"`
}

// ServeHTTP 路由：POST /rounds 创建，GET /rounds 列表
func (h *RoundsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	base := "/data/api/v1/data/vital-focus/rounds"
	if path != base && path != base+"/" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	switch r.Method {
	case http.MethodPost:
		h.createRound(w, r)
	case http.MethodGet:
		h.listRounds(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
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
		StartedAt:    body.StartedAt,
		ItemsChecked: body.ItemsChecked,
	}
	roundID, err := h.svc.CreateRound(ctx, tenantID, userID, req)
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
	// Phase 3: 按 user Current Branch (is_primary) 过滤；admin/sysadmin 无挂载 → 不限制
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
	if v := r.URL.Query().Get("executor_id"); v != "" {
		filters.ExecutorID = v
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
	// 将 items_checked JSON 转为前端可读
	items := make([]map[string]interface{}, 0, len(rounds))
	for _, rd := range rounds {
		row := map[string]interface{}{
			"round_id":     rd.RoundID,
			"tenant_id":    rd.TenantID,
			"round_type":   rd.RoundType,
			"unit_id":      rd.UnitID,
			"executor_id":  rd.ExecutorID,
			"round_time":   rd.RoundTime,
			"notes":        rd.Notes,
			"status":       rd.Status,
		}
		if rd.StartedAt != nil {
			row["started_at"] = rd.StartedAt.Format(time.RFC3339)
		}
		if len(rd.ItemsChecked) > 0 {
			var arr []domain.RoundItemChecked
			if _ = json.Unmarshal(rd.ItemsChecked, &arr); len(arr) > 0 {
				row["items_checked"] = arr
			}
		}
		items = append(items, row)
	}
	writeJSON(w, http.StatusOK, Ok(map[string]interface{}{
		"items": items,
		"total": total,
		"page":  page,
		"size":  size,
	}))
}
