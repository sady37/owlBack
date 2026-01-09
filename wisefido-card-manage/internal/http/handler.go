package http

import (
	"encoding/json"
	"net/http"
	"wisefido-card-manage/internal/service"

	"go.uber.org/zap"
)

// Handler HTTP 请求处理器
type Handler struct {
	cardService *service.CardService
	logger      *zap.Logger
}

// NewHandler 创建 HTTP 处理器
func NewHandler(cardService *service.CardService, logger *zap.Logger) *Handler {
	return &Handler{
		cardService: cardService,
		logger:      logger,
	}
}

// CreateCardsForUnitRequest 创建卡片请求
type CreateCardsForUnitRequest struct {
	TenantID string `json:"tenant_id"`
	UnitID   string `json:"unit_id"`
}

// CreateCardsForUnitResponse 创建卡片响应
type CreateCardsForUnitResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Stats   *struct {
		ExistingCount  int `json:"existing_count"`
		CreatedCount   int `json:"created_count"`
		UpdatedCount   int `json:"updated_count"`
		DeletedCount   int `json:"deleted_count"`
		UnchangedCount int `json:"unchanged_count"`
	} `json:"stats,omitempty"`
}

// CreateCardsForUnit 处理创建卡片请求
// POST /api/v1/cards/create
func (h *Handler) CreateCardsForUnit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateCardsForUnitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.TenantID == "" || req.UnitID == "" {
		http.Error(w, "tenant_id and unit_id are required", http.StatusBadRequest)
		return
	}

	stats, err := h.cardService.CreateCardsForUnit(r.Context(), req.TenantID, req.UnitID)
	if err != nil {
		h.logger.Error("Failed to create cards",
			zap.String("tenant_id", req.TenantID),
			zap.String("unit_id", req.UnitID),
			zap.Error(err),
		)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(CreateCardsForUnitResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(CreateCardsForUnitResponse{
		Success: true,
		Stats: &struct {
			ExistingCount  int `json:"existing_count"`
			CreatedCount   int `json:"created_count"`
			UpdatedCount   int `json:"updated_count"`
			DeletedCount   int `json:"deleted_count"`
			UnchangedCount int `json:"unchanged_count"`
		}{
			ExistingCount:  stats.ExistingCount,
			CreatedCount:   stats.CreatedCount,
			UpdatedCount:   stats.UpdatedCount,
			DeletedCount:   stats.DeletedCount,
			UnchangedCount: stats.UnchangedCount,
		},
	})
}

// CreateAllCardsRequest 全量创建卡片请求
type CreateAllCardsRequest struct {
	TenantID string `json:"tenant_id"`
}

// CreateAllCardsResponse 全量创建卡片响应
type CreateAllCardsResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// CreateAllCards 处理全量创建卡片请求
// POST /api/v1/cards/create-all
func (h *Handler) CreateAllCards(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateAllCardsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.TenantID == "" {
		http.Error(w, "tenant_id is required", http.StatusBadRequest)
		return
	}

	err := h.cardService.CreateAllCards(r.Context())
	if err != nil {
		h.logger.Error("Failed to create all cards",
			zap.String("tenant_id", req.TenantID),
			zap.Error(err),
		)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(CreateAllCardsResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(CreateAllCardsResponse{
		Success: true,
		Message: "All cards created/updated successfully",
	})
}

