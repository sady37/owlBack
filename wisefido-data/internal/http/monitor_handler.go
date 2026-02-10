package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"wisefido-data/internal/models"
	"wisefido-data/internal/service"

	commoncard "owl-common/card"

	"go.uber.org/zap"
)

// MonitorHandler 仅做 HTTP 转换：卡片列表/静态/实时/SSE 转发给对应 Service
type MonitorHandler struct {
	cardService cardServiceInterface
	realtime    realtimeServiceInterface
	logger      *zap.Logger
}

func NewMonitorHandler(logger *zap.Logger) *MonitorHandler {
	return &MonitorHandler{logger: logger}
}

type cardServiceInterface interface {
	GetCardList(ctx context.Context, tenantID, userID, userRole string, branchIDs []string, page, pageSize int) ([]commoncard.VitalFocusCardInfo, *models.BackendPagination, error)
	GetCardInfo(ctx context.Context, tenantID, userID, cardID string) (*commoncard.VitalFocusCardInfo, error)
}

func (h *MonitorHandler) SetCardService(cardService cardServiceInterface) {
	h.cardService = cardService
}

type realtimeServiceInterface interface {
	GetCardRealtime(ctx context.Context, req service.GetCardRealtimeRequest) (*service.GetCardRealtimeResponse, error)
	SubscribeRealtimeStream(ctx context.Context, w http.ResponseWriter, cardID, tenantID, userID string)
}

func (h *MonitorHandler) SetRealtimeService(realtime realtimeServiceInterface) {
	h.realtime = realtime
}

// GET /data/api/v1/data/vital-focus/cards
func (h *MonitorHandler) GetCards(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUserID, tenantID, _, claimedUserRole, ok := service.MustSession(ctx)
	if !ok || tenantID == "" {
		writeJSON(w, http.StatusUnauthorized, Fail("missing or invalid authorization"))
		return
	}
	branchIDsStr := r.URL.Query().Get("branch_ids")
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	pageSize := 10
	if ps := r.URL.Query().Get("pageSize"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 {
			pageSize = v
			if pageSize > 100 {
				pageSize = 100
			}
		}
	}
	var branchIDs []string
	if branchIDsStr != "" {
		branchIDs = strings.Split(branchIDsStr, ",")
	}
	if h.cardService == nil {
		writeJSON(w, http.StatusOK, Fail("card service not available"))
		return
	}
	cards, pagination, err := h.cardService.GetCardList(ctx, tenantID, currentUserID, claimedUserRole, branchIDs, page, pageSize)
	if err != nil {
		h.logger.Error("GetCardList failed", zap.String("user_id", currentUserID), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	if pagination == nil {
		pagination = &models.BackendPagination{}
	}
	writeJSON(w, http.StatusOK, Ok(map[string]interface{}{
		"items":      cards,
		"pagination": pagination,
	}))
}

// GET /data/api/v1/data/vital-focus/card/{id}
func (h *MonitorHandler) GetCardInfo(w http.ResponseWriter, r *http.Request, cardID string) {
	ctx := r.Context()
	currentUserID, tenantID, _, _, ok := service.MustSession(ctx)
	if !ok || currentUserID == "" || tenantID == "" {
		writeJSON(w, http.StatusUnauthorized, Fail("missing or invalid authorization"))
		return
	}
	if h.cardService == nil {
		writeJSON(w, http.StatusOK, Fail("card service not available"))
		return
	}
	card, err := h.cardService.GetCardInfo(ctx, tenantID, currentUserID, cardID)
	if err != nil {
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(card))
}

// GetCardRealtime 拉取实时数据
func (h *MonitorHandler) GetCardRealtime(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := r.URL.Query().Get("tenant_id")

	var cardsByBranch map[string][]string
	if r.Body != nil && r.Method != http.MethodGet {
		var body struct {
			CardsByBranch map[string][]string `json:"cards_by_branch"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		cardsByBranch = body.CardsByBranch
	}
	if cardsByBranch == nil {
		if v := r.URL.Query().Get("cards_by_branch"); v != "" {
			_ = json.Unmarshal([]byte(v), &cardsByBranch)
		}
	}
	if cardsByBranch == nil {
		var cardIDs []string
		if v := r.URL.Query().Get("card_ids"); v != "" {
			parts := strings.Split(v, ",")
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					cardIDs = append(cardIDs, p)
				}
			}
		}
		if len(cardIDs) == 0 {
			cardIDs = r.URL.Query()["card_id"]
		}
		if len(cardIDs) > 0 {
			cardsByBranch = map[string][]string{"": cardIDs}
		}
	}
	if cardsByBranch == nil {
		cardsByBranch = map[string][]string{}
	}

	currentUserID, _, _, _, ok := service.MustSession(ctx)
	if !ok || currentUserID == "" {
		h.logger.Warn("Missing session for GetCardRealtime", zap.String("user_id", currentUserID))
		writeJSON(w, http.StatusUnauthorized, Fail("missing or invalid authorization"))
		return
	}

	if h.realtime == nil {
		writeJSON(w, http.StatusOK, Fail("realtime service not available"))
		return
	}

	req := service.GetCardRealtimeRequest{
		TenantID:      tenantID,
		UserID:        currentUserID,
		CardsByBranch: cardsByBranch,
	}

	resp, err := h.realtime.GetCardRealtime(ctx, req)
	if err != nil {
		h.logger.Error("Failed to get card realtime",
			zap.String("user_id", currentUserID),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(resp))
}

// SubscribeRealtimeStream 转发到 CardRealtimeService.SubscribeRealtimeStream（SSE 由 Service 负责）
func (h *MonitorHandler) SubscribeRealtimeStream(w http.ResponseWriter, r *http.Request, cardID string) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if cardID == "" || strings.Contains(cardID, "/") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	currentUserID, tenantID, _, _, ok := service.MustSession(r.Context())
	if !ok || currentUserID == "" || tenantID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if h.realtime == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	h.realtime.SubscribeRealtimeStream(r.Context(), w, cardID, tenantID, currentUserID)
}
