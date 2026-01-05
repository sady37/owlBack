package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"
	"wisefido-data/internal/models"
	"wisefido-data/internal/service"
	"wisefido-data/internal/store"

	"go.uber.org/zap"
)

// VitalFocusHandler 实现 owlFront Monitor API 所需接口
type VitalFocusHandler struct {
	kv          store.KV
	db          *sql.DB             // 用于更新 users.preferences
	cardService service.CardService // 用于权限过滤的卡片服务
	logger      *zap.Logger
}

func NewVitalFocusHandler(kv store.KV, logger *zap.Logger) *VitalFocusHandler {
	return &VitalFocusHandler{kv: kv, logger: logger}
}

// SetCardService 设置卡片服务（用于权限过滤）
func (h *VitalFocusHandler) SetCardService(cardService service.CardService) {
	h.cardService = cardService
}

// SetDB 设置数据库连接（用于更新 preferences）
func (h *VitalFocusHandler) SetDB(db *sql.DB) {
	h.db = db
}

// GET /data/api/v1/data/vital-focus/cards
// params:
// - tenant_id? string
// - page? number (default 1)
// - pageSize? number (default 10)  <-- 前端 mock 使用
// - size? number (alias)
// headers:
// - X-User-Id: 用户 ID（必填）
// - X-User-Type: 用户类型 "resident" | "staff"（必填）
// - X-User-Role: 用户角色（staff 必填）
func (h *VitalFocusHandler) GetCards(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 如果 cardService 可用，使用 service 层（带权限过滤）
	if h.cardService != nil {
		h.getCardsWithService(ctx, w, r)
		return
	}

	// 向后兼容：如果 cardService 不可用，使用旧的直接扫描 Redis 方式（无权限过滤）
	h.getCardsLegacy(ctx, w, r)
}

// getCardsWithService 使用 service 层获取卡片（带权限过滤）
func (h *VitalFocusHandler) getCardsWithService(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// 从 Query 获取参数
	tenantID := r.URL.Query().Get("tenant_id")
	page := parseInt(r.URL.Query().Get("page"), 1)
	pageSize := parseInt(r.URL.Query().Get("pageSize"), 0)
	if pageSize <= 0 {
		pageSize = parseInt(r.URL.Query().Get("size"), 10)
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 从 Header 获取用户信息
	currentUserID := r.Header.Get("X-User-Id")
	currentUserType := r.Header.Get("X-User-Type")
	currentUserRole := r.Header.Get("X-User-Role")

	if currentUserID == "" || currentUserType == "" {
		h.logger.Warn("Missing required headers: X-User-Id or X-User-Type",
			zap.String("user_id", currentUserID),
			zap.String("user_type", currentUserType),
		)
		writeJSON(w, http.StatusOK, Fail("missing required headers: X-User-Id and X-User-Type"))
		return
	}

	// 调用 service 层
	req := service.ListVitalFocusCardsRequest{
		TenantID:        tenantID,
		Page:            page,
		PageSize:        pageSize,
		CurrentUserID:   currentUserID,
		CurrentUserType: currentUserType,
		CurrentUserRole: currentUserRole,
	}

	resp, err := h.cardService.ListVitalFocusCards(ctx, req)
	if err != nil {
		h.logger.Error("Failed to list vital focus cards",
			zap.String("user_id", currentUserID),
			zap.String("user_type", currentUserType),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 转换为前端模型
	items := make([]models.VitalFocusCard, len(resp.Items))
	for i, card := range resp.Items {
		items[i] = card
	}

	backendResp := models.GetVitalFocusCardsModel{
		Items: items,
		Pagination: models.BackendPagination{
			Size:      resp.Pagination.PageSize,
			Page:      resp.Pagination.Page,
			Count:     resp.Pagination.Total,
			Sort:      "",
			Direction: 0,
		},
	}

	writeJSON(w, http.StatusOK, Ok(backendResp))
}

// getCardsLegacy 向后兼容：直接扫描 Redis（无权限过滤）
func (h *VitalFocusHandler) getCardsLegacy(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	page := parseInt(r.URL.Query().Get("page"), 1)
	pageSize := parseInt(r.URL.Query().Get("pageSize"), 0)
	if pageSize <= 0 {
		pageSize = parseInt(r.URL.Query().Get("size"), 10)
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 从 Redis 扫描 full cache（不依赖 DB）
	keys, err := h.kv.ScanKeys(ctx, "vital-focus:card:*:full")
	if err != nil {
		// 联调友好：当 Redis 不可用/没有跑 aggregator 时，不要让前端报错；返回空列表即可
		h.logger.Warn("ScanKeys failed, returning empty cards list", zap.Error(err))
		resp := models.GetVitalFocusCardsModel{
			Items: []models.VitalFocusCard{},
			Pagination: models.BackendPagination{
				Size:      pageSize,
				Page:      page,
				Count:     0,
				Sort:      "",
				Direction: 0,
			},
		}
		writeJSON(w, http.StatusOK, Ok(resp))
		return
	}

	all := make([]models.VitalFocusCard, 0, len(keys))
	for _, key := range keys {
		raw, err := h.kv.Get(ctx, key)
		if err != nil {
			continue
		}
		card, ok := decodeAndNormalizeFullCard(raw)
		if !ok {
			continue
		}
		if tenantID != "" && card.TenantID != tenantID {
			continue
		}
		all = append(all, card)
	}

	// 简单排序：按 card_id
	// （后续可按 sort/direction 扩展）
	// 这里不引入额外依赖，保持轻量
	sortCardsByID(all)

	total := len(all)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	resp := models.GetVitalFocusCardsModel{
		Items: all[start:end],
		Pagination: models.BackendPagination{
			Size:      pageSize,
			Page:      page,
			Count:     total,
			Sort:      "",
			Direction: 0,
		},
	}

	writeJSON(w, http.StatusOK, Ok(resp))
}

// GET /data/api/v1/data/vital-focus/card/{id}
// 兼容前端两种用法：
// - id = card_id
// - id = resident_id （如果 card_id 未命中，则尝试按 resident 查找）
func (h *VitalFocusHandler) GetCardByIDOrResident(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	// 1) 先当作 card_id 直接读取 full cache
	if card, ok := h.getCardFullByCardID(ctx, id); ok {
		writeJSON(w, http.StatusOK, Ok(toCardInfo(card)))
		return
	}

	// 2) 再按 resident_id 查找（扫描 full cache）
	keys, err := h.kv.ScanKeys(ctx, "vital-focus:card:*:full")
	if err != nil {
		writeJSON(w, http.StatusOK, Fail("failed to scan cards"))
		return
	}

	for _, key := range keys {
		raw, err := h.kv.Get(ctx, key)
		if err != nil {
			continue
		}
		card, ok := decodeAndNormalizeFullCard(raw)
		if !ok {
			continue
		}
		if card.PrimaryResidentID == id {
			writeJSON(w, http.StatusOK, Ok(toCardInfo(card)))
			return
		}
		for _, r := range card.Residents {
			if r.ResidentID == id {
				writeJSON(w, http.StatusOK, Ok(toCardInfo(card)))
				return
			}
		}
	}

	writeJSON(w, http.StatusOK, Fail("card not found"))
}

// GET /data/api/v1/data/vital-focus/preferences
// headers:
// - X-User-Id: 用户 ID（必填）
// - X-Tenant-Id: 租户 ID（可选，用于验证）
func (h *VitalFocusHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := r.Header.Get("X-Tenant-Id")
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		writeJSON(w, http.StatusOK, Fail("user ID is required"))
		return
	}

	// 如果 cardService 可用，使用 service 层
	if h.cardService != nil {
		req := service.GetVitalFocusPreferencesRequest{
			TenantID:      tenantID,
			CurrentUserID: userID,
		}

		resp, err := h.cardService.GetVitalFocusPreferences(ctx, req)
		if err != nil {
			h.logger.Error("Failed to get vital focus preferences",
				zap.String("user_id", userID),
				zap.Error(err),
			)
			writeJSON(w, http.StatusOK, Fail(err.Error()))
			return
		}

		writeJSON(w, http.StatusOK, Ok(map[string]any{
			"selected_card_ids": resp.SelectedCardIDs,
		}))
		return
	}

	// 向后兼容：如果 cardService 不可用，直接查询数据库
	if h.db == nil {
		writeJSON(w, http.StatusOK, Fail("database not available"))
		return
	}

	var preferencesJSON sql.NullString
	err := h.db.QueryRowContext(ctx,
		`SELECT preferences FROM users WHERE user_id = $1`,
		userID,
	).Scan(&preferencesJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusOK, Fail("user not found"))
			return
		}
		writeJSON(w, http.StatusOK, Fail("failed to get user preferences"))
		return
	}

	var selectedCardIDs []string
	if preferencesJSON.Valid && preferencesJSON.String != "" {
		var prefs map[string]interface{}
		if err := json.Unmarshal([]byte(preferencesJSON.String), &prefs); err == nil {
			if vitalFocus, ok := prefs["vitalFocus"].(map[string]interface{}); ok {
				if cardIds, ok := vitalFocus["selectedCardIds"].([]interface{}); ok {
					for _, id := range cardIds {
						if idStr, ok := id.(string); ok {
							selectedCardIDs = append(selectedCardIDs, idStr)
						}
					}
				}
			}
		}
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"selected_card_ids": selectedCardIDs,
	}))
}

// POST /data/api/v1/data/vital-focus/selection
// body: { selected_card_ids: string[] }
// headers:
// - X-User-Id: 用户 ID（必填）
// - X-Tenant-Id: 租户 ID（可选，用于验证）
func (h *VitalFocusHandler) SaveSelection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := r.Header.Get("X-Tenant-Id")
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		writeJSON(w, http.StatusOK, Fail("user ID is required"))
		return
	}
	if tenantID == "" {
		tenantID = SystemTenantID()
	}

	var req struct {
		SelectedCardIDs []string `json:"selected_card_ids"`
	}
	if err := readBodyJSON(r, 1<<20, &req); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	// 如果 cardService 可用，使用 service 层
	if h.cardService != nil {
		serviceReq := service.SaveVitalFocusPreferencesRequest{
			TenantID:        tenantID,
			CurrentUserID:   userID,
			SelectedCardIDs: req.SelectedCardIDs,
		}

		err := h.cardService.SaveVitalFocusPreferences(ctx, serviceReq)
		if err != nil {
			h.logger.Error("Failed to save vital focus preferences",
				zap.String("user_id", userID),
				zap.Error(err),
			)
			writeJSON(w, http.StatusOK, Fail(err.Error()))
			return
		}

		// 同时保存到 Redis（向后兼容）
		key := "vital-focus:selection:user:" + userID
		raw, _ := json.Marshal(req)
		_ = h.kv.Set(ctx, key, string(raw), 7*24*time.Hour) // 保存 7 天

		writeJSON(w, http.StatusOK, Ok(map[string]any{
			"success": true,
			"message": "Focus selection saved successfully",
		}))
		return
	}

	// 向后兼容：如果 cardService 不可用，直接更新数据库
	// 1. 保存到 Redis（向后兼容）
	key := "vital-focus:selection:user:" + userID
	raw, _ := json.Marshal(req)
	_ = h.kv.Set(ctx, key, string(raw), 7*24*time.Hour) // 保存 7 天

	// 2. 保存到 users.preferences.vitalFocus.selectedCardIds（如果 DB 可用）
	if h.db != nil {
		// 获取当前用户的 preferences
		var currentPrefsJSON sql.NullString
		err := h.db.QueryRowContext(ctx,
			`SELECT preferences FROM users WHERE tenant_id = $1 AND user_id = $2`,
			tenantID, userID,
		).Scan(&currentPrefsJSON)

		if err == nil {
			// 解析现有 preferences
			var prefs map[string]interface{}
			if currentPrefsJSON.Valid && currentPrefsJSON.String != "" {
				if err := json.Unmarshal([]byte(currentPrefsJSON.String), &prefs); err != nil {
					prefs = make(map[string]interface{})
				}
			} else {
				prefs = make(map[string]interface{})
			}

			// 更新 vitalFocus.selectedCardIds
			if prefs["vitalFocus"] == nil {
				prefs["vitalFocus"] = make(map[string]interface{})
			}
			vitalFocus, ok := prefs["vitalFocus"].(map[string]interface{})
			if !ok {
				vitalFocus = make(map[string]interface{})
				prefs["vitalFocus"] = vitalFocus
			}
			vitalFocus["selectedCardIds"] = req.SelectedCardIDs

			// 保存回数据库
			updatedPrefsJSON, _ := json.Marshal(prefs)
			_, err = h.db.ExecContext(ctx,
				`UPDATE users SET preferences = $1::jsonb WHERE tenant_id = $2 AND user_id = $3`,
				string(updatedPrefsJSON), tenantID, userID,
			)
			if err != nil {
				h.logger.Warn("Failed to update user preferences", zap.Error(err))
				// 不返回错误，因为 Redis 已保存成功
			}
		} else if err != sql.ErrNoRows {
			h.logger.Warn("Failed to get user preferences", zap.Error(err))
			// 不返回错误，因为 Redis 已保存成功
		}
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"success": true,
		"message": "Focus selection saved successfully",
	}))
}

func (h *VitalFocusHandler) getCardFullByCardID(ctx context.Context, cardID string) (models.VitalFocusCard, bool) {
	key := "vital-focus:card:" + cardID + ":full"
	raw, err := h.kv.Get(ctx, key)
	if err != nil {
		return models.VitalFocusCard{}, false
	}
	card, ok := decodeAndNormalizeFullCard(raw)
	return card, ok
}

func toCardInfo(c models.VitalFocusCard) models.VitalFocusCardInfo {
	return models.VitalFocusCardInfo{
		CardID:            c.CardID,
		TenantID:          c.TenantID,
		CardType:          c.CardType,
		BedID:             c.BedID,
		LocationID:        c.LocationID,
		CardName:          c.CardName,
		CardAddress:       c.CardAddress,
		PrimaryResidentID: c.PrimaryResidentID,
		Residents:         c.Residents,
		Devices:           c.Devices,
	}
}

// minimal sort by card_id without importing sort package (but we can use sort, it's stdlib)
func sortCardsByID(cards []models.VitalFocusCard) {
	// simple insertion sort (n small typically); avoids importing sort
	for i := 1; i < len(cards); i++ {
		j := i
		for j > 0 && strings.Compare(cards[j-1].CardID, cards[j].CardID) > 0 {
			cards[j-1], cards[j] = cards[j], cards[j-1]
			j--
		}
	}
}

// for tests/mocking
var _ = context.Background

// --- normalization layer (align with owlFront interfaces) ---

// decodeAndNormalizeFullCard:
// - card-aggregator 写入的 full cache 目前字段类型与前端不完全一致（例如 device_type 为 string）
// - 这里在 API 层做一次规范化，确保返回结构与 owlFront 的 TypeScript model 对齐
func decodeAndNormalizeFullCard(raw string) (models.VitalFocusCard, bool) {
	// 先用 map 解析，避免类型不一致导致整体 unmarshal 失败
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return models.VitalFocusCard{}, false
	}

	// 再把 map 转回 json，然后 unmarshal 到目标模型（容忍少数字段缺失）
	b, err := json.Marshal(m)
	if err != nil {
		return models.VitalFocusCard{}, false
	}
	var card models.VitalFocusCard
	if err := json.Unmarshal(b, &card); err != nil {
		// 兜底：如果模型解析失败，就直接返回失败
		return models.VitalFocusCard{}, false
	}

	// residents：确保 last_name 有值（前端类型标注为必填）
	for i := range card.Residents {
		if card.Residents[i].LastName == "" {
			if card.Residents[i].Nickname != "" {
				card.Residents[i].LastName = card.Residents[i].Nickname
			} else {
				card.Residents[i].LastName = "-"
			}
		}
	}

	// devices：device_type 规范化为 number（sleepace=1, radar=2）
	for i := range card.Devices {
		switch v := card.Devices[i].DeviceType.(type) {
		case string:
			card.Devices[i].DeviceType = deviceTypeToNumber(v)
		case float64:
			// json number -> float64
			card.Devices[i].DeviceType = int(v)
		default:
			// keep as-is
		}
	}

	// heart_source/breath_source：如果被写成 Sleepace/Radar，规范为 s/r/-
	if card.HeartSource != "" {
		card.HeartSource = normalizeSource(card.HeartSource)
	}
	if card.BreathSource != "" {
		card.BreathSource = normalizeSource(card.BreathSource)
	}

	return card, true
}

func deviceTypeToNumber(s string) int {
	switch s {
	case "Sleepace", "SleepPad", "Sleepad", "SleepAd":
		return 1
	case "Radar":
		return 2
	default:
		return 0
	}
}

func normalizeSource(s string) string {
	switch s {
	case "s", "r", "-":
		return s
	case "Sleepace", "SleepPad":
		return "s"
	case "Radar":
		return "r"
	default:
		return "-"
	}
}
