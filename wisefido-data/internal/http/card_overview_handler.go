package httpapi

import (
	"context"
	"net/http"
	"strconv"

	"wisefido-data/internal/service"

	"go.uber.org/zap"
)

// CardOverviewProvider 卡片概览服务接口
type CardOverviewProvider interface {
	GetCardOverview(ctx context.Context, req service.GetCardOverviewRequest) (*service.GetCardOverviewResponse, error)
}

// CardOverviewHandler 卡片概览 Handler
type CardOverviewHandler struct {
	base   *StubHandler
	svc    CardOverviewProvider
	logger *zap.Logger
}

// NewCardOverviewHandler 创建卡片概览 Handler
func NewCardOverviewHandler(
	base *StubHandler,
	svc CardOverviewProvider,
	logger *zap.Logger,
) *CardOverviewHandler {
	return &CardOverviewHandler{
		base:   base,
		svc:    svc,
		logger: logger,
	}
}

// ServeHTTP 处理 HTTP 请求
func (h *CardOverviewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// 只处理 GET /admin/api/v1/card-overview
	if r.URL.Path != "/admin/api/v1/card-overview" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	h.GetCardOverview(w, r)
}

// GetCardOverview 获取卡片概览列表
func (h *CardOverviewHandler) GetCardOverview(w http.ResponseWriter, r *http.Request) {
	// 1. 提取 tenant_id
	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	// 2. 解析查询参数（不包含 page 和 pageSize）
	cardID := r.URL.Query().Get("card_id")
	search := r.URL.Query().Get("search")
	cardType := r.URL.Query().Get("card_type")
	unitType := r.URL.Query().Get("unit_type")

	var isPublicSpace *bool
	if isPublicSpaceStr := r.URL.Query().Get("is_public_space"); isPublicSpaceStr != "" {
		if val, err := strconv.ParseBool(isPublicSpaceStr); err == nil {
			isPublicSpace = &val
		}
	}

	var isSharedUnit *bool
	if isSharedUnitStr := r.URL.Query().Get("is_shared_unit"); isSharedUnitStr != "" {
		if val, err := strconv.ParseBool(isSharedUnitStr); err == nil {
			isSharedUnit = &val
		}
	}
	// 向后兼容：也支持 is_multi_person_room 参数
	if isSharedUnit == nil {
		if isMultiPersonRoomStr := r.URL.Query().Get("is_multi_person_room"); isMultiPersonRoomStr != "" {
			if val, err := strconv.ParseBool(isMultiPersonRoomStr); err == nil {
				isSharedUnit = &val
			}
		}
	}

	sortVal := r.URL.Query().Get("sort")
	if sortVal == "" {
		sortVal = "card_name"
	}

	direction := r.URL.Query().Get("direction")
	if direction == "" {
		direction = "asc"
	}

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	pageSize := 20
	if ps := r.URL.Query().Get("pageSize"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 20 {
			pageSize = v
		}
	} else if ps := r.URL.Query().Get("size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 20 {
			pageSize = v
		}
	}

	// 3. 获取用户信息（从 HTTP Header）
	currentUserID := r.Header.Get("X-User-Id")
	currentUserType := r.Header.Get("X-User-Type")
	currentUserRole := r.Header.Get("X-User-Role")

	h.logger.Debug("GetCardOverview request",
		zap.String("tenant_id", tenantID),
		zap.String("current_user_id", currentUserID),
		zap.String("current_user_type", currentUserType),
		zap.String("current_user_role", currentUserRole),
	)

	// 4. 构建请求
	req := service.GetCardOverviewRequest{
		TenantID:        tenantID,
		CardID:          cardID,
		Search:          search,
		CardType:        cardType,
		UnitType:        unitType,
		IsPublicSpace:   isPublicSpace,
		IsSharedUnit:    isSharedUnit,
		Sort:            sortVal,
		Direction:       direction,
		Page:            page,
		PageSize:        pageSize,
		CurrentUserID:   currentUserID,
		CurrentUserType: currentUserType,
		CurrentUserRole: currentUserRole,
	}

	// 5. 调用 Service（svc 为 nil 时返回空列表）
	if h.svc == nil {
		directionNum := 0
		if req.Direction == "desc" {
			directionNum = 1
		}
		writeJSON(w, http.StatusOK, Ok(map[string]any{
			"items": []any{},
			"pagination": map[string]any{
				"total": 0, "page": 1, "size": 20, "count": 0,
				"sort": req.Sort, "direction": directionNum,
			},
		}))
		return
	}
	resp, err := h.svc.GetCardOverview(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get card overview",
			zap.Error(err),
			zap.String("tenant_id", tenantID),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	h.logger.Info("GetCardOverview response",
		zap.String("tenant_id", tenantID),
		zap.Int("total", resp.Total),
		zap.String("current_user_role", currentUserRole),
	)

	directionNum := 0
	if direction == "desc" {
		directionNum = 1
	}
	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"items": resp.Items,
		"pagination": map[string]any{
			"total":     resp.Total,
			"page":      resp.Page,
			"size":      resp.Size,
			"count":     resp.Total,
			"sort":      req.Sort,
			"direction": directionNum,
		},
	}))
}
