package httpapi

import (
	"net/http"
	"strconv"

	"wisefido-data/internal/service"

	"go.uber.org/zap"
)

// CardOverviewHandler 卡片概览 Handler
type CardOverviewHandler struct {
	base        *StubHandler
	cardService service.CardService
	logger      *zap.Logger
}

// NewCardOverviewHandler 创建卡片概览 Handler
func NewCardOverviewHandler(
	base *StubHandler,
	cardService service.CardService,
	logger *zap.Logger,
) *CardOverviewHandler {
	return &CardOverviewHandler{
		base:        base,
		cardService: cardService,
		logger:      logger,
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

	var isMultiPersonRoom *bool
	if isMultiPersonRoomStr := r.URL.Query().Get("is_multi_person_room"); isMultiPersonRoomStr != "" {
		if val, err := strconv.ParseBool(isMultiPersonRoomStr); err == nil {
			isMultiPersonRoom = &val
		}
	}

	sort := r.URL.Query().Get("sort")
	if sort == "" {
		sort = "card_name"
	}

	direction := r.URL.Query().Get("direction")
	if direction == "" {
		direction = "asc"
	}

	// 3. 获取用户信息（从 HTTP Header）
	currentUserID := r.Header.Get("X-User-Id")
	currentUserType := r.Header.Get("X-User-Type")
	currentUserRole := r.Header.Get("X-User-Role")

	// 4. 构建请求
	req := service.GetCardOverviewRequest{
		TenantID:          tenantID,
		CardID:            cardID,
		Search:            search,
		CardType:          cardType,
		UnitType:          unitType,
		IsPublicSpace:     isPublicSpace,
		IsMultiPersonRoom: isMultiPersonRoom,
		Sort:              sort,
		Direction:         direction,
		CurrentUserID:     currentUserID,
		CurrentUserType:   currentUserType,
		CurrentUserRole:   currentUserRole,
	}

	// 5. 调用 Service
	resp, err := h.cardService.GetCardOverview(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get card overview",
			zap.Error(err),
			zap.String("tenant_id", tenantID),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 6. 返回响应（所有可见的卡片）
	// 注意：前端期望 pagination 包含更多字段，但后端不处理分页，只返回 total
	directionNum := 0 // 0=asc, 1=desc
	if direction == "desc" {
		directionNum = 1
	}
	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"items": resp.Items,
		"pagination": map[string]any{
			"total": resp.Total, // 总数（前端用于分页）
			// 以下字段由前端控制，后端不处理，但为了兼容性可以返回默认值
			"page":      1,
			"size":      10,
			"count":     resp.Total,
			"sort":      sort,
			"direction": directionNum, // 前端使用数字：0=asc, 1=desc
		},
	}))
}

