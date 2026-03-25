package service

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"owl-common/card"
	"wisefido-qinglan/internal/domain"
	"wisefido-qinglan/internal/repository"

	"go.uber.org/zap"
)

// DeviceBoundResolver returns device-bound room_id and bed_id (e.g. from device_store).
type DeviceBoundResolver interface {
	GetBoundRoomAndBed(ctx context.Context, deviceUID string) (roomID, bedID string)
}

// CardMappingService 只缓存 DeviceBaseline（与 CardDB / 流头一致）。
type CardMappingService struct {
	cardDB        *card.CardDB
	logger        *zap.Logger
	boundResolver DeviceBoundResolver

	mu            sync.RWMutex
	baselineByUID map[string]card.DeviceBaseline
}

func NewCardMappingService(cardDB *card.CardDB, logger *zap.Logger) *CardMappingService {
	return &CardMappingService{
		cardDB:        cardDB,
		logger:        logger,
		baselineByUID: make(map[string]card.DeviceBaseline),
	}
}

// SetDeviceBoundResolver sets the optional resolver for RoomID/BedID (e.g. qinglan device_store).
func (s *CardMappingService) SetDeviceBoundResolver(r DeviceBoundResolver) {
	s.boundResolver = r
}

func effectiveCardIDFromBaseline(b card.DeviceBaseline) string {
	if c := strings.TrimSpace(b.CardID); c != "" {
		return c
	}
	return strings.TrimSpace(b.DeviceID)
}

func (s *CardMappingService) enrichBaselineRoomBed(ctx context.Context, lookupKey string, b *card.DeviceBaseline) {
	if s.boundResolver == nil || b == nil {
		return
	}
	uid := strings.TrimSpace(b.DeviceUID)
	if uid == "" {
		uid = lookupKey
	}
	b.RoomID, b.BedID = s.boundResolver.GetBoundRoomAndBed(ctx, uid)
}

// storeBaseline 须在持锁下调用；写入 room/bed 已合并后的快照。
func (s *CardMappingService) storeBaseline(lookupKey string, b card.DeviceBaseline) {
	keys := []string{lookupKey}
	if u := strings.TrimSpace(b.DeviceUID); u != "" && u != lookupKey {
		keys = append(keys, u)
	}
	for _, k := range keys {
		if k == "" {
			continue
		}
		s.baselineByUID[k] = b
		domain.AllowAccessCache.Store(k, b.AllowAccess)
	}
}

func (s *CardMappingService) evictDeviceKey(k string) {
	delete(s.baselineByUID, k)
	domain.AllowAccessCache.Delete(k)
}

// BaselineFor 返回已缓存的 DeviceBaseline（MQTT allow_access / resolveIotPolicy 快路径）。
func (s *CardMappingService) BaselineFor(deviceUID string) (card.DeviceBaseline, bool) {
	if deviceUID == "" {
		return card.DeviceBaseline{}, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.baselineByUID[deviceUID]
	return b, ok
}

// RefreshBaseline 从 DB 重算并写入 baseline + AllowAccessCache。
func (s *CardMappingService) RefreshBaseline(ctx context.Context, lookupKey string) {
	if s == nil || s.cardDB == nil || lookupKey == "" {
		return
	}
	b, ok := s.cardDB.ResolveDeviceBaseline(ctx, lookupKey)
	s.mu.Lock()
	defer s.mu.Unlock()
	if !ok {
		s.evictDeviceKey(lookupKey)
		domain.AllowAccessCache.Store(lookupKey, false)
		return
	}
	s.enrichBaselineRoomBed(ctx, lookupKey, &b)
	s.storeBaseline(lookupKey, b)
}

// InvalidateCache clears baseline cache.
func (s *CardMappingService) InvalidateCache() {
	s.mu.Lock()
	s.baselineByUID = make(map[string]card.DeviceBaseline)
	s.mu.Unlock()
	if s.logger != nil {
		s.logger.Info("baseline cache invalidated")
	}
}

// InvalidateByCardID removes cached entries whose有效 card_id 匹配。
func (s *CardMappingService) InvalidateByCardID(cardID string) int {
	if cardID == "" {
		return 0
	}
	s.mu.Lock()
	n := 0
	for key, b := range s.baselineByUID {
		if effectiveCardIDFromBaseline(b) == cardID {
			s.evictDeviceKey(key)
			n++
		}
	}
	s.mu.Unlock()
	if n > 0 && s.logger != nil {
		s.logger.Info("baseline entries evicted", zap.String("card_id", cardID), zap.Int("count", n))
	}
	return n
}

// InvalidateByDeviceUID removes baseline for one lookup key。
func (s *CardMappingService) InvalidateByDeviceUID(deviceUID string) {
	if deviceUID == "" {
		return
	}
	s.mu.Lock()
	s.evictDeviceKey(deviceUID)
	s.mu.Unlock()
	if s.logger != nil {
		s.logger.Info("baseline evicted for device", zap.String("device_uid", deviceUID))
	}
}

// InvalidateByTenantUnit 按 unit 失效该 unit 下所有出现在 cards.devices 中的 device_id / device_uid 键。
func (s *CardMappingService) InvalidateByTenantUnit(ctx context.Context, tenantID, unitID string) {
	if s.cardDB == nil || tenantID == "" || unitID == "" {
		return
	}
	dids, uids := s.cardDB.DeviceKeysInTenantUnit(ctx, tenantID, unitID)
	seen := make(map[string]struct{})
	s.mu.Lock()
	for _, id := range dids {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		s.evictDeviceKey(id)
	}
	for _, id := range uids {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		s.evictDeviceKey(id)
	}
	s.mu.Unlock()
	if s.logger != nil {
		s.logger.Info("baseline cache evicted for tenant unit",
			zap.String("tenant_id", tenantID), zap.String("unit_id", unitID), zap.Int("keys", len(seen)))
	}
}

// GetCardIDByDeviceUID resolves deviceUID → DeviceBaseline（懒加载；字段含 card_id 等，与 CardDB 一致）。
func (s *CardMappingService) GetCardIDByDeviceUID(ctx context.Context, deviceUID string) (*card.DeviceBaseline, error) {
	if deviceUID == "" {
		return nil, fmt.Errorf("empty device uid")
	}
	s.mu.RLock()
	b, ok := s.baselineByUID[deviceUID]
	s.mu.RUnlock()
	if ok {
		out := b
		return &out, nil
	}

	if s.cardDB == nil {
		return nil, fmt.Errorf("card db not configured")
	}
	b, ok = s.cardDB.ResolveDeviceBaseline(ctx, deviceUID)
	if !ok {
		return nil, fmt.Errorf("device not in cards: %s", deviceUID)
	}
	s.enrichBaselineRoomBed(ctx, deviceUID, &b)

	s.mu.Lock()
	s.storeBaseline(deviceUID, b)
	s.mu.Unlock()

	out := b
	return &out, nil
}

// deviceBoundResolver implements DeviceBoundResolver from device_store (BoundRoomID/BoundBedID).
type deviceBoundResolver struct {
	repo repository.DeviceRepository
}

// NewDeviceBoundResolver returns a resolver that fills RoomID/BedID from device_store.
func NewDeviceBoundResolver(repo repository.DeviceRepository) DeviceBoundResolver {
	return &deviceBoundResolver{repo: repo}
}

func (r *deviceBoundResolver) GetBoundRoomAndBed(ctx context.Context, deviceUID string) (roomID, bedID string) {
	dev, err := r.repo.GetDeviceByUID(ctx, deviceUID)
	if err != nil || dev == nil {
		return "", ""
	}
	if dev.BoundRoomID.Valid {
		roomID = dev.BoundRoomID.String
	}
	if dev.BoundBedID.Valid {
		bedID = dev.BoundBedID.String
	}
	return roomID, bedID
}

var _ DeviceBoundResolver = (*deviceBoundResolver)(nil)
