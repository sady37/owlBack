package cache

import (
	"context"
	"encoding/json"
	"fmt"

	"wisefido-data/internal/repository"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const vitalFocusStaticKeyPrefix = "vital-focus:card:"
const vitalFocusStaticKeySuffix = ":static"

// VitalFocusStaticCache 将 cards 表+unit 生成 VitalFocusCardInfo 静态 JSON 写入 Redis，供前端快速返回
type VitalFocusStaticCache struct {
	redis  *redis.Client
	repo   *repository.PostgresCardRepository
	logger *zap.Logger
}

// NewVitalFocusStaticCache 创建静态缓存写入器
func NewVitalFocusStaticCache(redis *redis.Client, repo *repository.PostgresCardRepository, logger *zap.Logger) *VitalFocusStaticCache {
	return &VitalFocusStaticCache{redis: redis, repo: repo, logger: logger}
}

// WriteCardStatic 从 DB 取卡+unit，生成静态 JSON 写入 Redis
func (c *VitalFocusStaticCache) WriteCardStatic(ctx context.Context, tenantID, cardID string) error {
	row, err := c.repo.GetCardRowForCache(ctx, tenantID, cardID)
	if err != nil || row == nil {
		if err != nil {
			return err
		}
		return nil
	}
	// 构建与 VitalFocusCardInfo 静态部分兼容的 JSON
	payload := map[string]interface{}{
		"card_id":             row.CardID,
		"tenant_id":           row.TenantID,
		"card_type":           row.CardType,
		"bed_id":              row.BedID,
		"unit_id":             row.UnitID,
		"card_name":           row.CardName,
		"card_address":        row.CardAddress,
		"timezone":            row.Timezone,
		"branch_id":           row.BranchID,
		"branch_name":         row.BranchName,
		"primary_resident_id": row.ResidentID,
		"icon_alarm_level":    row.IconAlarmLevel,
		"pop_alarm_emerge":    row.PopAlarmEmerge,
	}
	if row.DevicesJSON != nil {
		payload["devices"] = json.RawMessage(row.DevicesJSON)
	} else {
		payload["devices"] = json.RawMessage([]byte("[]"))
	}
	if row.ResidentsJSON != nil {
		payload["residents"] = json.RawMessage(row.ResidentsJSON)
	} else {
		payload["residents"] = json.RawMessage([]byte("[]"))
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal card static: %w", err)
	}
	key := vitalFocusStaticKeyPrefix + cardID + vitalFocusStaticKeySuffix
	if err := c.redis.Set(ctx, key, string(b), 0).Err(); err != nil {
		return err
	}
	return nil
}

// DeleteCardStatic 删除该卡的静态缓存
func (c *VitalFocusStaticCache) DeleteCardStatic(ctx context.Context, cardID string) error {
	key := vitalFocusStaticKeyPrefix + cardID + vitalFocusStaticKeySuffix
	return c.redis.Del(ctx, key).Err()
}
