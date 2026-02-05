package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"wisefido-data/internal/store"

	"go.uber.org/zap"
)

const realtimeKeyPrefix = "vital-focus:card:"
const realtimeKeySuffix = ":realtime"
const maxPullRealtimeCards = 40

// CardRealtimeService 卡片实时数据服务：接收 card_id 列表，从 Redis 读取并返回
// 以 server 允许的 card 列表为准，不在清单的跳过；读 Redis 失败则跳过并记录
type CardRealtimeService struct {
	kv     store.KV
	perm   AllowedCardIDsProvider
	logger *zap.Logger
}

// NewCardRealtimeService 创建卡片实时数据服务
func NewCardRealtimeService(kv store.KV, perm AllowedCardIDsProvider, logger *zap.Logger) *CardRealtimeService {
	return &CardRealtimeService{kv: kv, perm: perm, logger: logger}
}

// GetCardRealtimeRequest 拉取实时数据请求（一次最多 40 个，超量请分页请求）
type GetCardRealtimeRequest struct {
	TenantID string
	UserID   string
	UserType string // "resident" | "staff"
	UserRole string
	CardIDs  []string // 用户提交的 card_id 列表，最多 40 个
}

// GetCardRealtimeResponse 拉取实时数据响应
type GetCardRealtimeResponse struct {
	Data            map[string]json.RawMessage `json:"data"`              // card_id -> realtime JSON
	SkippedCardIDs  []string                   `json:"skipped_card_ids"`  // 被跳过（不在允许清单、Redis miss、Get 失败、非法 JSON）
	RefreshCardList bool                       `json:"refresh_card_list"` // 有跳过时 true，建议前端重新调用 ListCardsForUser
}

// GetCardRealtime 从 Redis 拉取指定 card 的实时数据；以 server 允许的 card 为准，错误则跳过并记录
func (s *CardRealtimeService) GetCardRealtime(ctx context.Context, req GetCardRealtimeRequest) (*GetCardRealtimeResponse, error) {
	if req.TenantID == "" || req.UserID == "" || req.UserType == "" {
		return nil, fmt.Errorf("tenant_id, user_id, user_type are required")
	}
	if len(req.CardIDs) > maxPullRealtimeCards {
		return nil, fmt.Errorf("card_ids exceeds limit %d, use pagination", maxPullRealtimeCards)
	}
	if s.perm == nil {
		return nil, fmt.Errorf("permission provider not configured")
	}

	allowed, err := s.perm.GetAllowedCardIDs(ctx, req.TenantID, req.UserID, req.UserType, req.UserRole)
	if err != nil {
		return nil, fmt.Errorf("get allowed cards: %w", err)
	}
	allowedSet := make(map[string]bool, len(allowed))
	for _, id := range allowed {
		allowedSet[id] = true
	}

	// 无允许清单时，建议前端刷新
	resp := &GetCardRealtimeResponse{
		Data:            make(map[string]json.RawMessage),
		SkippedCardIDs:  make([]string, 0),
		RefreshCardList: len(allowed) == 0,
	}

	for _, cardID := range req.CardIDs {
		if cardID == "" {
			continue
		}
		if !allowedSet[cardID] {
			resp.SkippedCardIDs = append(resp.SkippedCardIDs, cardID)
			s.logger.Debug("skip card not in allowed list", zap.String("card_id", cardID))
			continue
		}
		key := realtimeKeyPrefix + cardID + realtimeKeySuffix
		raw, err := s.kv.Get(ctx, key)
		if err != nil {
			if errors.Is(err, store.ErrMiss) {
				// Redis 无值，跳过
				resp.SkippedCardIDs = append(resp.SkippedCardIDs, cardID)
				s.logger.Debug("realtime cache miss", zap.String("card_id", cardID))
			} else {
				resp.SkippedCardIDs = append(resp.SkippedCardIDs, cardID)
				s.logger.Warn("get realtime cache failed", zap.String("card_id", cardID), zap.Error(err))
			}
			continue
		}
		if !json.Valid([]byte(raw)) {
			resp.SkippedCardIDs = append(resp.SkippedCardIDs, cardID)
			s.logger.Warn("invalid realtime json", zap.String("card_id", cardID))
			continue
		}
		resp.Data[cardID] = json.RawMessage(raw)
	}

	// 有跳过则建议前端重新拉取 ListCardsForUser
	resp.RefreshCardList = len(resp.SkippedCardIDs) > 0
	return resp, nil
}
