package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"wisefido-data/internal/store"

	"go.uber.org/zap"
)

const userCardsKeyPrefix = "vital-focus:allowed-cards:"
const userCardsIndexPrefix = "vital-focus:allowed-cards:index:"

// UserCardsCache 用户可见卡片 ID 缓存，按 {tenantID}:{branchID}:{userID} 存储
// 无 TTL，依赖失效删除
type UserCardsCache struct {
	kv     store.KV
	logger *zap.Logger
}

// NewUserCardsCache 创建用户卡片缓存
func NewUserCardsCache(kv store.KV, logger *zap.Logger) *UserCardsCache {
	return &UserCardsCache{kv: kv, logger: logger}
}

// Get 读取用户可见的卡片 ID 列表；miss 返回 (nil, nil)，错误返回 (nil, err)
func (c *UserCardsCache) Get(ctx context.Context, tenantID, userID string, branchIDs []string) ([]string, error) {
	if len(branchIDs) == 0 {
		return nil, nil
	}
	var merged []string
	seen := make(map[string]bool)
	for _, branchID := range branchIDs {
		if branchID == "" {
			branchID = "_"
		}
		key := c.branchKey(tenantID, branchID, userID)
		raw, err := c.kv.Get(ctx, key)
		if err != nil {
			if errors.Is(err, store.ErrMiss) {
				return nil, nil // 任一 miss 则整体 miss
			}
			return nil, err
		}
		var ids []string
		if json.Unmarshal([]byte(raw), &ids) != nil {
			c.logger.Warn("unmarshal user cards cache failed", zap.String("key", key))
			return nil, nil
		}
		for _, id := range ids {
			if id != "" && !seen[id] {
				seen[id] = true
				merged = append(merged, id)
			}
		}
	}
	return merged, nil
}

// Set 按 branch 写入缓存；byBranch 的 key 为 branchID（空用 "_"）
func (c *UserCardsCache) Set(ctx context.Context, tenantID, userID string, byBranch map[string][]string) error {
	branchIDs := make([]string, 0, len(byBranch))
	for branchID, ids := range byBranch {
		if len(ids) == 0 {
			continue
		}
		if branchID == "" {
			branchID = "_"
		}
		b, err := json.Marshal(ids)
		if err != nil {
			return fmt.Errorf("marshal card ids: %w", err)
		}
		key := c.branchKey(tenantID, branchID, userID)
		if err := c.kv.Set(ctx, key, string(b), 0); err != nil {
			return err
		}
		branchIDs = append(branchIDs, branchID)
	}
	// 写入 index，供下次读取时知道要取哪些 branch
	indexKey := userCardsIndexPrefix + tenantID + ":" + userID
	idxBytes, err := json.Marshal(branchIDs)
	if err != nil {
		return fmt.Errorf("marshal index: %w", err)
	}
	return c.kv.Set(ctx, indexKey, string(idxBytes), 0)
}

// GetBranchIDsFromIndex 从 index 读取该用户的 branch 列表；miss 返回 nil
func (c *UserCardsCache) GetBranchIDsFromIndex(ctx context.Context, tenantID, userID string) ([]string, error) {
	key := userCardsIndexPrefix + tenantID + ":" + userID
	raw, err := c.kv.Get(ctx, key)
	if err != nil {
		if errors.Is(err, store.ErrMiss) {
			return nil, nil
		}
		return nil, err
	}
	var branchIDs []string
	if json.Unmarshal([]byte(raw), &branchIDs) != nil {
		return nil, nil
	}
	return branchIDs, nil
}

// InvalidateByTenantBranch 按 tenant+branch 失效，删除该 branch 下所有用户的缓存
func (c *UserCardsCache) InvalidateByTenantBranch(ctx context.Context, tenantID, branchID string) error {
	if branchID == "" {
		branchID = "_"
	}
	pattern := c.branchKey(tenantID, branchID, "*")
	keys, err := c.kv.ScanKeys(ctx, pattern)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	userIDs := make(map[string]bool)
	for _, k := range keys {
		userID := c.parseUserIDFromBranchKey(k, tenantID, branchID)
		if userID != "" {
			userIDs[userID] = true
		}
	}
	var toDel []string
	toDel = append(toDel, keys...)
	for userID := range userIDs {
		toDel = append(toDel, userCardsIndexPrefix+tenantID+":"+userID)
	}
	return c.kv.Del(ctx, toDel...)
}

func (c *UserCardsCache) branchKey(tenantID, branchID, userID string) string {
	return userCardsKeyPrefix + tenantID + ":" + branchID + ":" + userID
}

func (c *UserCardsCache) parseUserIDFromBranchKey(key, tenantID, branchID string) string {
	prefix := userCardsKeyPrefix + tenantID + ":" + branchID + ":"
	if !strings.HasPrefix(key, prefix) {
		return ""
	}
	return key[len(prefix):]
}
