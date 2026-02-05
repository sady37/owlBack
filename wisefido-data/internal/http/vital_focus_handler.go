package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"wisefido-data/internal/repository"
	"wisefido-data/internal/service"

	commoncard "owl-common/card"

	"go.uber.org/zap"
)

// VitalFocusHandler 实现 owlFront Monitor API 所需接口
type VitalFocusHandler struct {
	cardService cardServiceInterface       // 用于权限过滤的卡片服务
	realtime    realtimeServiceInterface   // 用于获取实时数据
	usersRepo   repository.UsersRepository // 用于验证用户角色
	logger      *zap.Logger
}

func NewVitalFocusHandler(logger *zap.Logger) *VitalFocusHandler {
	return &VitalFocusHandler{logger: logger}
}

// SetCardService 设置卡片服务（用于权限过滤）
type cardServiceInterface interface {
	ListCards(ctx context.Context, tenantID, userID, userRole string, branchIDs []string) ([]commoncard.CardIndexItem, error)
	GetCardInfo(ctx context.Context, tenantID, userID, userRole, cardID string) (*commoncard.VitalFocusCardInfo, error)
}

func (h *VitalFocusHandler) SetCardService(cardService cardServiceInterface) {
	h.cardService = cardService
}

// SetUsersRepo 设置用户 repository 用于会话验证
func (h *VitalFocusHandler) SetUsersRepo(usersRepo repository.UsersRepository) {
	h.usersRepo = usersRepo
}

// realtimeServiceInterface 定义了需要的实时数据接口
type realtimeServiceInterface interface {
	GetCardRealtime(ctx context.Context, req service.GetCardRealtimeRequest) (*service.GetCardRealtimeResponse, error)
}

// SetRealtimeService 设置实时数据服务
func (h *VitalFocusHandler) SetRealtimeService(realtime realtimeServiceInterface) {
	h.realtime = realtime
}

// GET /data/api/v1/data/vital-focus/cards
// params:
// - tenant_id? string
// - branch_ids? string (comma-separated)
// headers:
// - X-User-Id: 用户 ID（必填）
// - X-User-Role: 用户角色（可选，用于权限检查，但需验证与实际用户角色是否一致）
func (h *VitalFocusHandler) GetCards(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 从 Query 获取参数
	tenantID := r.URL.Query().Get("tenant_id")
	branchIDsStr := r.URL.Query().Get("branch_ids")

	var branchIDs []string
	if branchIDsStr != "" {
		branchIDs = strings.Split(branchIDsStr, ",")
	}

	// 从 Header 获取用户信息
	currentUserID := r.Header.Get("X-User-Id")
	claimedUserRole := r.Header.Get("X-User-Role")

	if currentUserID == "" {
		h.logger.Warn("Missing required header: X-User-Id")
		writeJSON(w, http.StatusOK, Fail("missing required header: X-User-Id"))
		return
	}

	// 验证用户存在且获取真实角色（防止伪造）
	var actualUserRole string
	if h.usersRepo != nil {
		user, err := h.usersRepo.GetUser(ctx, tenantID, currentUserID)
		if err != nil {
			h.logger.Warn("Failed to verify user",
				zap.String("user_id", currentUserID),
				zap.String("tenant_id", tenantID),
				zap.Error(err),
			)
			writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("user verification failed: %v", err)))
			return
		}
		actualUserRole = user.Role

		// 如果提供了 claimed role，验证是否与实际 role 匹配
		if claimedUserRole != "" && !strings.EqualFold(actualUserRole, claimedUserRole) {
			h.logger.Warn("User role mismatch",
				zap.String("user_id", currentUserID),
				zap.String("claimed_role", claimedUserRole),
				zap.String("actual_role", actualUserRole),
			)
			writeJSON(w, http.StatusOK, Fail("user role verification failed: role mismatch"))
			return
		}
	} else {
		// usersRepo 未配置，使用提供的 role（不推荐用于生产环境）
		actualUserRole = claimedUserRole
	}

	// 调用 service 层获取卡片列表（使用验证过的 actual role）
	cards, err := h.cardService.ListCards(ctx, tenantID, currentUserID, actualUserRole, branchIDs)
	if err != nil {
		h.logger.Error("Failed to list cards for user",
			zap.String("user_id", currentUserID),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(map[string]interface{}{
		"items": cards,
	}))
}

// GET /data/api/v1/data/vital-focus/card/{id}
// 获取卡片详细信息
// headers:
// - X-User-Id: 用户 ID（必填）
// - X-User-Role: 用户角色（可选，用于权限检查）
func (h *VitalFocusHandler) GetCardInfo(w http.ResponseWriter, r *http.Request, cardID string) {
	ctx := r.Context()

	// 从 Query 获取参数
	tenantID := r.URL.Query().Get("tenant_id")

	// 从 Header 获取用户信息
	currentUserID := r.Header.Get("X-User-Id")
	claimedUserRole := r.Header.Get("X-User-Role")

	if currentUserID == "" {
		h.logger.Warn("Missing required header: X-User-Id")
		writeJSON(w, http.StatusOK, Fail("missing required header: X-User-Id"))
		return
	}

	// 验证用户存在且获取真实角色（防止伪造）
	var actualUserRole string
	if h.usersRepo != nil {
		user, err := h.usersRepo.GetUser(ctx, tenantID, currentUserID)
		if err != nil {
			h.logger.Warn("Failed to verify user",
				zap.String("user_id", currentUserID),
				zap.String("tenant_id", tenantID),
				zap.Error(err),
			)
			writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("user verification failed: %v", err)))
			return
		}
		actualUserRole = user.Role

		// 如果提供了 claimed role，验证是否与实际 role 匹配
		if claimedUserRole != "" && !strings.EqualFold(actualUserRole, claimedUserRole) {
			h.logger.Warn("User role mismatch",
				zap.String("user_id", currentUserID),
				zap.String("claimed_role", claimedUserRole),
				zap.String("actual_role", actualUserRole),
			)
			writeJSON(w, http.StatusOK, Fail("user role verification failed: role mismatch"))
			return
		}
	} else {
		// usersRepo 未配置，使用提供的 role（不推荐用于生产环境）
		actualUserRole = claimedUserRole
	}

	if h.cardService == nil {
		writeJSON(w, http.StatusOK, Fail("card service not available"))
		return
	}

	// 调用服务层获取卡片详情（使用验证过的 actual role）
	cardInfo, err := h.cardService.GetCardInfo(ctx, tenantID, currentUserID, actualUserRole, cardID)
	if err != nil {
		h.logger.Error("Failed to get card info",
			zap.String("card_id", cardID),
			zap.String("user_id", currentUserID),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	if cardInfo == nil {
		writeJSON(w, http.StatusOK, Fail("card not found"))
		return
	}

	writeJSON(w, http.StatusOK, Ok(cardInfo))
}

// params:
// - tenant_id? string
// - card_ids? comma separated list OR repeated `card_id` params
// headers:
// - X-User-Id: 用户 ID（必填）
// - X-User-Type: 用户类型 "resident" | "staff"（必填）
// - X-User-Role: 用户角色（staff 可选）
func (h *VitalFocusHandler) GetCardRealtime(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := r.URL.Query().Get("tenant_id")

	// collect card ids from `card_ids` (comma separated) or repeated `card_id`
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
		// check repeated card_id params
		cardIDs = r.URL.Query()["card_id"]
	}

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

	if h.realtime == nil {
		writeJSON(w, http.StatusOK, Fail("realtime service not available"))
		return
	}

	req := service.GetCardRealtimeRequest{
		TenantID: tenantID,
		UserID:   currentUserID,
		UserType: currentUserType,
		UserRole: currentUserRole,
		CardIDs:  cardIDs,
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

// POST /data/api/v1/data/vital-focus/selection
// 后端未提供此功能 - 已移除
