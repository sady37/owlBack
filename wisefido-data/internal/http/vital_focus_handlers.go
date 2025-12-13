package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
	"wisefido-data/internal/models"
	"wisefido-data/internal/store"

	"go.uber.org/zap"
)

// VitalFocusHandler 实现 owlFront Monitor API 所需接口
type VitalFocusHandler struct {
	kv     store.KV
	logger *zap.Logger
}

func NewVitalFocusHandler(kv store.KV, logger *zap.Logger) *VitalFocusHandler {
	return &VitalFocusHandler{kv: kv, logger: logger}
}

// GET /data/api/v1/data/vital-focus/cards
// params:
// - tenant_id? string
// - page? number (default 1)
// - pageSize? number (default 10)  <-- 前端 mock 使用
// - size? number (alias)
func (h *VitalFocusHandler) GetCards(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

// POST /data/api/v1/data/vital-focus/selection
// body: { selected_card_ids: string[] }
func (h *VitalFocusHandler) SaveSelection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		// 前端会发送，但为了兼容先不强制
		userID = "anonymous"
	}

	var req struct {
		SelectedCardIDs []string `json:"selected_card_ids"`
	}
	if err := readBodyJSON(r, 1<<20, &req); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	key := "vital-focus:selection:user:" + userID
	raw, _ := json.Marshal(req)
	_ = h.kv.Set(ctx, key, string(raw), 7*24*time.Hour) // 保存 7 天，后续可改为永久或 DB

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


