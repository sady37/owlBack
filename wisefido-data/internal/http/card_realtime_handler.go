package httpapi

import (
	"encoding/json"
	"net/http"

	"wisefido-data/internal/service"

	"go.uber.org/zap"
)

// CardRealtimeHandler 处理卡片实时数据拉取请求（按 branch 分组的 cardIDs → 每张卡当前 vital/state 快照）。
type CardRealtimeHandler struct {
	realtimeSvc *service.CardRealtimeService
	staticSvc   *service.CardStaticService
	logger      *zap.Logger
}

func NewCardRealtimeHandler(realtimeSvc *service.CardRealtimeService, staticSvc *service.CardStaticService, logger *zap.Logger) *CardRealtimeHandler {
	return &CardRealtimeHandler{
		realtimeSvc: realtimeSvc,
		staticSvc:   staticSvc,
		logger:      logger,
	}
}

// PullRealtimeData 处理实时数据拉取请求 (POST /data/api/v1/data/vital-focus/pull-realtime)。
// Body: { cards_by_branch: { branchID → []cardID } } —— 按 branch 分组传，便于 service 端 scope 校验。
func (h *CardRealtimeHandler) PullRealtimeData(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	var body struct {
		CardsByBranch map[string][]string `json:"cards_by_branch"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		h.logger.Error("Failed to decode pull realtime request", zap.Error(err))
		writeJSON(w, http.StatusBadRequest, Fail("Invalid request body"))
		return
	}
	if body.CardsByBranch == nil {
		body.CardsByBranch = map[string][]string{}
	}

	tenantID := req.Header.Get("X-Tenant-Id")
	userID := req.Header.Get("X-User-Id")
	if tenantID == "" || userID == "" {
		h.logger.Error("Missing required authentication headers")
		writeJSON(w, http.StatusUnauthorized, Fail("Missing required authentication headers"))
		return
	}

	serviceReq := service.GetCardRealtimeRequest{
		TenantID:      tenantID,
		UserID:        userID,
		CardsByBranch: body.CardsByBranch,
	}

	resp, err := h.realtimeSvc.GetCardRealtime(ctx, serviceReq)
	if err != nil {
		h.logger.Error("Failed to pull realtime data", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(resp))
}
