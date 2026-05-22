package service

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"owl-common/card"
	"wisefido-qinglan/internal/domain"

	"go.uber.org/zap"
)

// CardMappingService 只缓存 DeviceBaseline（与 CardAPIClient / 流头一致）。
type CardMappingService struct {
	api    *card.CardAPIClient
	logger *zap.Logger

	mu            sync.RWMutex
	baselineByUID map[string]card.DeviceBaseline

	readyCh chan struct{} // closed when cache is ready; re-created on invalidate
}

func NewCardMappingService(api *card.CardAPIClient, logger *zap.Logger) *CardMappingService {
	ch := make(chan struct{})
	close(ch) // initially ready
	return &CardMappingService{
		api:           api,
		logger:        logger,
		baselineByUID: make(map[string]card.DeviceBaseline),
		readyCh:       ch,
	}
}

// WaitReady blocks until the cache is ready after an invalidation cycle.
func (s *CardMappingService) WaitReady() {
	<-s.readyCh
}

func effectiveCardIDFromBaseline(b card.DeviceBaseline) string {
	if c := strings.TrimSpace(b.CardID); c != "" {
		return c
	}
	// Phase 2 一刀切：device_id UUID 退役；fallback 用 DeviceAddr canonical text
	if b.DeviceAddr.IsValid() {
		return b.DeviceAddr.String()
	}
	return ""
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
		domain.AllowAccessCache.Store(k, b.Access)
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
	<-s.readyCh
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.baselineByUID[deviceUID]
	return b, ok
}

// RefreshBaseline 从 API 重算并写入 baseline + AllowAccessCache。
func (s *CardMappingService) RefreshBaseline(ctx context.Context, lookupKey string) {
	if s == nil || s.api == nil || lookupKey == "" {
		return
	}
	b, err := s.api.LookupBaseline(ctx, lookupKey)
	s.mu.Lock()
	defer s.mu.Unlock()
	if err != nil || b == nil {
		s.evictDeviceKey(lookupKey)
		domain.AllowAccessCache.Store(lookupKey, false)
		return
	}
	s.storeBaseline(lookupKey, *b)
}

// InvalidateCache clears baseline cache with readyCh gate.
func (s *CardMappingService) InvalidateCache() {
	ch := make(chan struct{})
	s.mu.Lock()
	s.readyCh = ch
	s.baselineByUID = make(map[string]card.DeviceBaseline)
	s.mu.Unlock()
	close(ch)
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

// GetCardIDByDeviceUID resolves deviceUID → DeviceBaseline（懒加载；字段含 card_id 等，与 CardAPIClient 一致）。
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

	if s.api == nil {
		return nil, fmt.Errorf("card api not configured")
	}
	bp, err := s.api.LookupBaseline(ctx, deviceUID)
	if err != nil || bp == nil {
		return nil, fmt.Errorf("device not in cards: %s", deviceUID)
	}

	s.mu.Lock()
	s.storeBaseline(deviceUID, *bp)
	s.mu.Unlock()

	return bp, nil
}
