package httpapi

import (
	"context"
	"encoding/json"
	"net"
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
	GetCardList(ctx context.Context, tenantID, userID, userRole string, branchIDs []string, page, pageSize int) ([]commoncard.CardStatic, *models.BackendPagination, error)
	GetCardInfo(ctx context.Context, tenantID, userID, cardID string) (*commoncard.CardStatic, error)
	GetCardsByCardIDs(ctx context.Context, tenantID, userID string, cardIDs []string) ([]commoncard.CardStatic, error)
	GetCardCaregivers(ctx context.Context, residentIDs []string) ([]string, []commoncard.CaregiverInfo, error)
}

func (h *MonitorHandler) SetCardService(cardService cardServiceInterface) {
	h.cardService = cardService
}

type realtimeServiceInterface interface {
	GetCardRealtime(ctx context.Context, req service.GetCardRealtimeRequest) (*service.GetCardRealtimeResponse, error)
	GetCardStatus(ctx context.Context, cardID string) (map[string]interface{}, error)
	SubscribeRealtimeStream(ctx context.Context, w http.ResponseWriter, cardID, tenantID, userID string)
	SubscribeCardsStream(ctx context.Context, w http.ResponseWriter, interval int, tenantID, userID, clientIP, userAgent string)
	InitSSE(connID string, watchIDs, viewIDs []string) error
	UpdateSSEView(connID string, newViewIDs []string) error
	UpdateSSEWatch(connID string, watchIDs []string) error
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
	ps := r.URL.Query().Get("pageSize")
	if ps == "" {
		ps = r.URL.Query().Get("page_size")
	}
	if ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 {
			pageSize = v
			if pageSize > 500 {
				pageSize = 500
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
	h.logger.Info("[API] GetCards request",
		zap.String("user_id", currentUserID),
		zap.String("tenant_id", tenantID),
		zap.String("role", claimedUserRole),
		zap.Int("page", page),
		zap.Int("pageSize", pageSize))
	cards, pagination, err := h.cardService.GetCardList(ctx, tenantID, currentUserID, claimedUserRole, branchIDs, page, pageSize)
	if err != nil {
		h.logger.Error("GetCardList failed", zap.String("user_id", currentUserID), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	if pagination == nil {
		pagination = &models.BackendPagination{}
	}
	// 打印返回的 card_id 列表
	cardIDs := make([]string, len(cards))
	for i, c := range cards {
		cardIDs[i] = c.CardID
	}
	h.logger.Info("[API] GetCards response",
		zap.Int("count", len(cards)),
		zap.Strings("card_ids", cardIDs))
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
	h.logger.Info("[API] GetCardInfo request",
		zap.String("user_id", currentUserID),
		zap.String("card_id", cardID))
	card, err := h.cardService.GetCardInfo(ctx, tenantID, currentUserID, cardID)
	if err != nil {
		h.logger.Error("[API] GetCardInfo failed", zap.String("card_id", cardID), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	h.logger.Info("[API] GetCardInfo response",
		zap.String("card_id", cardID),
		zap.Bool("found", card != nil))
	writeJSON(w, http.StatusOK, Ok(card))
}

// GET /data/api/v1/data/vital-focus/card/{id}/status
// cardID 路径参数接收 IPv6 CIDR 或 dns_short_name；下游 realtime/Redis key 必须是 IPv6 spatial_prefix，
// 所以**必须**用 GetCardInfo 返回的 card.CardID（已解析为 IPv6）调下游，不要直接用原始 cardID 字符串。
func (h *MonitorHandler) GetCardStatus(w http.ResponseWriter, r *http.Request, cardID string) {
	ctx := r.Context()
	currentUserID, tenantID, _, _, ok := service.MustSession(ctx)
	if !ok || tenantID == "" {
		writeJSON(w, http.StatusUnauthorized, Fail("missing or invalid authorization"))
		return
	}
	if h.realtime == nil {
		writeJSON(w, http.StatusOK, Fail("realtime service not available"))
		return
	}
	if h.cardService == nil {
		writeJSON(w, http.StatusOK, Fail("card service not available"))
		return
	}
	// Phase 3 scope check + 短码解析（GetCardInfo 内部 resolveCardID 把 dns_short_name 解为 spatial_prefix）
	card, err := h.cardService.GetCardInfo(ctx, tenantID, currentUserID, cardID)
	if err != nil || card == nil {
		writeJSON(w, http.StatusOK, Fail("permission denied: card not in your scope"))
		return
	}
	resolvedCardID := card.CardID // IPv6 spatial_prefix CIDR
	data, err := h.realtime.GetCardStatus(ctx, resolvedCardID)
	if err != nil {
		h.logger.Warn("[API] GetCardStatus redis read failed",
			zap.String("card_id", resolvedCardID), zap.Error(err))
		writeJSON(w, http.StatusOK, Ok(map[string]interface{}(nil)))
		return
	}
	writeJSON(w, http.StatusOK, Ok(data))
}

// GET /data/api/v1/data/vital-focus/card/{id}/caregivers
func (h *MonitorHandler) GetCardCaregivers(w http.ResponseWriter, r *http.Request, cardID string) {
	ctx := r.Context()
	currentUserID, tenantID, userType, role, ok := service.MustSession(ctx)
	if !ok || currentUserID == "" || tenantID == "" {
		writeJSON(w, http.StatusUnauthorized, Fail("missing or invalid authorization"))
		return
	}
	if h.cardService == nil {
		writeJSON(w, http.StatusOK, Fail("card service not available"))
		return
	}
	// 先取卡片确认权限并获取 resident IDs
	card, err := h.cardService.GetCardInfo(ctx, tenantID, currentUserID, cardID)
	if err != nil || card == nil {
		writeJSON(w, http.StatusOK, Fail("card not found or no permission"))
		return
	}
	// facility 下只允许 staff（Admin/Manager/Caregiver/Nurse）查看
	if card.Unit != nil && card.Unit.UnitType == "facility" {
		if userType == "resident" {
			writeJSON(w, http.StatusForbidden, Fail("no permission"))
			return
		}
		allowed := map[string]bool{"Admin": true, "Manager": true, "Caregiver": true, "Nurse": true}
		if !allowed[role] {
			writeJSON(w, http.StatusForbidden, Fail("no permission"))
			return
		}
	}
	residentIDs := make([]string, 0, len(card.Residents))
	for _, r := range card.Residents {
		if r.ResidentID != "" {
			residentIDs = append(residentIDs, r.ResidentID)
		}
	}
	groups, caregivers, err := h.cardService.GetCardCaregivers(ctx, residentIDs)
	if err != nil {
		h.logger.Error("[API] GetCardCaregivers failed", zap.String("card_id", cardID), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"caregiver_groups": groups,
		"caregivers":       caregivers,
	}))
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

// POST /data/api/v1/data/vital-focus/cards-by-ids
// 前端收到 card_change 后按 cardID 拉静态数据
func (h *MonitorHandler) GetCardsByCardIDs(w http.ResponseWriter, r *http.Request) {
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

	var body struct {
		CardIDs []string `json:"card_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, Fail("invalid request body"))
		return
	}
	if len(body.CardIDs) == 0 {
		writeJSON(w, http.StatusOK, Ok([]commoncard.CardStatic{}))
		return
	}

	h.logger.Info("[API] GetCardsByCardIDs request",
		zap.String("user_id", currentUserID),
		zap.Strings("card_ids", body.CardIDs))
	cards, err := h.cardService.GetCardsByCardIDs(ctx, tenantID, currentUserID, body.CardIDs)
	if err != nil {
		h.logger.Error("[API] GetCardsByCardIDs failed", zap.String("user_id", currentUserID), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	if cards == nil {
		cards = []commoncard.CardStatic{}
	}
	retIDs := make([]string, len(cards))
	for i, c := range cards {
		retIDs[i] = c.CardID
	}
	h.logger.Info("[API] GetCardsByCardIDs response",
		zap.Int("requested", len(body.CardIDs)),
		zap.Int("returned", len(cards)),
		zap.Strings("returned_ids", retIDs))
	writeJSON(w, http.StatusOK, Ok(cards))
}

// GET /data/api/v1/data/vital-focus/cards/stream?card_ids=a,b,c&view_ids=a,b&interval=2
// Overview 多卡 SSE：GET 只传 token + interval，连接后前端 POST /init 传 watchIDs + viewIDs
func (h *MonitorHandler) SubscribeCardsStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
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

	interval := 2
	if v := r.URL.Query().Get("interval"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			interval = n
		}
	}

	h.realtime.SubscribeCardsStream(r.Context(), w, interval, tenantID, currentUserID, sseClientIP(r), r.UserAgent())
}

func sseClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// POST /data/api/v1/data/vital-focus/cards/stream/init
// SSE 连接建立后，前端提交 watchIDs + viewIDs 初始化订阅
func (h *MonitorHandler) InitSSE(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
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

	var body struct {
		ConnID   string   `json:"conn_id"`
		WatchIDs []string `json:"watch_ids"`
		ViewIDs  []string `json:"view_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, Fail("invalid request body"))
		return
	}
	if body.ConnID == "" {
		writeJSON(w, http.StatusBadRequest, Fail("conn_id required"))
		return
	}
	if len(body.WatchIDs) == 0 {
		writeJSON(w, http.StatusBadRequest, Fail("watch_ids required"))
		return
	}

	_ = currentUserID
	_ = tenantID

	if err := h.realtime.InitSSE(body.ConnID, body.WatchIDs, body.ViewIDs); err != nil {
		h.logger.Warn("[SSE] init failed",
			zap.String("conn_id", body.ConnID),
			zap.Int("watch_ids", len(body.WatchIDs)),
			zap.Error(err))
		writeJSON(w, http.StatusBadRequest, Fail(err.Error()))
		return
	}
	h.logger.Info("[SSE] init ok",
		zap.String("conn_id", body.ConnID),
		zap.Int("watch_ids", len(body.WatchIDs)),
		zap.Int("view_ids", len(body.ViewIDs)))
	writeJSON(w, http.StatusOK, Ok(map[string]interface{}{"status": "initialized"}))
}

// POST /data/api/v1/data/vital-focus/cards/stream/view
// 切页时更新 SSE 连接的 viewIDs，无需重连
func (h *MonitorHandler) UpdateSSEView(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
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

	var body struct {
		ConnID  string   `json:"conn_id"`
		ViewIDs []string `json:"view_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, Fail("invalid request body"))
		return
	}
	if body.ConnID == "" {
		writeJSON(w, http.StatusBadRequest, Fail("conn_id required"))
		return
	}

	_ = currentUserID
	_ = tenantID

	if err := h.realtime.UpdateSSEView(body.ConnID, body.ViewIDs); err != nil {
		writeJSON(w, http.StatusNotFound, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(map[string]interface{}{"status": "updated", "view_count": len(body.ViewIDs)}))
}

// POST /data/api/v1/data/vital-focus/cards/stream/watch
// card_change 后更新 watchIDs，无需重连 SSE
func (h *MonitorHandler) UpdateSSEWatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
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

	var body struct {
		ConnID   string   `json:"conn_id"`
		WatchIDs []string `json:"watch_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, Fail("invalid request body"))
		return
	}
	if body.ConnID == "" {
		writeJSON(w, http.StatusBadRequest, Fail("conn_id required"))
		return
	}

	_ = currentUserID
	_ = tenantID

	if err := h.realtime.UpdateSSEWatch(body.ConnID, body.WatchIDs); err != nil {
		writeJSON(w, http.StatusNotFound, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(map[string]interface{}{"status": "updated", "watch_count": len(body.WatchIDs)}))
}

// SubscribeRealtimeStream 转发到 CardRealtimeService.SubscribeRealtimeStream（SSE 由 Service 负责）
// cardID 路径参数支持 IPv6 CIDR 或 dns_short_name；下游 realtime/Redis key 必须是 IPv6 spatial_prefix，
// 必须经 GetCardInfo 解析后再用 card.CardID 调下游。
func (h *MonitorHandler) SubscribeRealtimeStream(w http.ResponseWriter, r *http.Request, cardID string) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if cardID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	currentUserID, tenantID, _, _, ok := service.MustSession(r.Context())
	if !ok || currentUserID == "" || tenantID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if h.realtime == nil || h.cardService == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	// 解析短码 + 权限校验
	card, err := h.cardService.GetCardInfo(r.Context(), tenantID, currentUserID, cardID)
	if err != nil || card == nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	h.realtime.SubscribeRealtimeStream(r.Context(), w, card.CardID, tenantID, currentUserID)
}
