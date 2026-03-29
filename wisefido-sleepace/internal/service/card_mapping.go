package service

import (
	"context"
	"sync"

	"owl-common/card"

	"go.uber.org/zap"
)

var sleepadDeviceTypes = []string{"Sleepad", "SleepPad", "sleepad"}

// CardMappingService provides MQTT deviceId → card info lookup.
//
// Two layers of in-memory cache:
//   - codeToUID: device_code → device_uid (loaded at startup, refreshed on config event)
//   - cardCache: device_uid → DeviceBaseline (populated on first lookup, cleared on config event)
type CardMappingService struct {
	cardDB *card.CardDB
	logger *zap.Logger

	mu        sync.RWMutex
	codeToUID map[string]string              // device_code → device_uid
	cardCache map[string]*card.DeviceBaseline // device_uid → full identity
}

func NewCardMappingService(cardDB *card.CardDB, logger *zap.Logger) *CardMappingService {
	return &CardMappingService{
		cardDB:    cardDB,
		logger:    logger,
		codeToUID: make(map[string]string),
		cardCache: make(map[string]*card.DeviceBaseline),
	}
}

// LoadCodeToUIDMap loads device_code → device_uid mapping from device_store at startup.
func (s *CardMappingService) LoadCodeToUIDMap(ctx context.Context) error {
	m, err := s.cardDB.LoadCodeToUIDMap(ctx, sleepadDeviceTypes)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.codeToUID = m
	s.mu.Unlock()
	s.logger.Info("loaded device_code → device_uid map", zap.Int("count", len(m)))
	return nil
}

// InvalidateCache clears both code→uid map and card cache, then reloads code→uid.
func (s *CardMappingService) InvalidateCache(ctx context.Context) {
	s.mu.Lock()
	s.cardCache = make(map[string]*card.DeviceBaseline)
	s.mu.Unlock()

	if err := s.LoadCodeToUIDMap(ctx); err != nil {
		s.logger.Error("reload code-to-uid map after invalidate", zap.Error(err))
	}
	s.logger.Info("card cache invalidated")
}

// InvalidateByDeviceUID 删除单设备在 cardCache 中的条目（device_uid 键）。
func (s *CardMappingService) InvalidateByDeviceUID(deviceUID string) {
	if deviceUID == "" {
		return
	}
	s.mu.Lock()
	delete(s.cardCache, deviceUID)
	s.mu.Unlock()
}

// InvalidateByTenantUnit 按 unit 失效该 unit 下 cards.devices 中出现的 device 键。
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
		s.evictLookupKeyLocked(id)
	}
	for _, id := range uids {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		s.evictLookupKeyLocked(id)
	}
	s.mu.Unlock()
	if s.logger != nil {
		s.logger.Info("card cache evicted for tenant unit",
			zap.String("tenant_id", tenantID), zap.String("unit_id", unitID), zap.Int("keys", len(seen)))
	}
}

// InvalidateByCardID removes cached entries belonging to a specific card.
func (s *CardMappingService) InvalidateByCardID(cardID string) int {
	s.mu.Lock()
	n := 0
	for uid, m := range s.cardCache {
		if m.CardID == cardID {
			delete(s.cardCache, uid)
			n++
		}
	}
	s.mu.Unlock()
	if n > 0 {
		s.logger.Info("card cache entries evicted", zap.String("card_id", cardID), zap.Int("count", n))
	}
	return n
}

func (s *CardMappingService) evictLookupKeyLocked(key string) {
	delete(s.cardCache, key)
	for uid, m := range s.cardCache {
		if m == nil {
			continue
		}
		if m.DeviceID == key || m.DeviceUID == key {
			delete(s.cardCache, uid)
		}
	}
}

// PutMapping adds a single device_code → device_uid entry to the in-memory map.
func (s *CardMappingService) PutMapping(deviceCode, deviceUID string) {
	if deviceCode == "" || deviceUID == "" {
		return
	}
	s.mu.Lock()
	s.codeToUID[deviceCode] = deviceUID
	s.mu.Unlock()
}

// CodeToUID returns device_uid for a given device_code, or "" if not found.
func (s *CardMappingService) CodeToUID(code string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.codeToUID[code]
}

// UIDToCode returns device_code for a given device_uid (reverse lookup).
func (s *CardMappingService) UIDToCode(uid string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for code, u := range s.codeToUID {
		if u == uid {
			return code
		}
	}
	return ""
}

// ResolveToDeviceUID 将 device_code 或 device_id 转为 device_uid，内部业务统一用 device_uid。
func (s *CardMappingService) ResolveToDeviceUID(ctx context.Context, id string) string {
	if id == "" {
		return ""
	}
	if u := s.CodeToUID(id); u != "" {
		return u
	}
	uid, _, err := s.cardDB.ResolveDevice(ctx, id)
	if err != nil {
		return ""
	}
	return uid
}

// GetCardInfo resolves MQTT deviceId → DeviceBaseline.
// Uses in-memory cache; DB is only hit on first lookup per device_uid.
func (s *CardMappingService) GetCardInfo(ctx context.Context, mqttDeviceID string) (*card.DeviceBaseline, error) {
	s.mu.RLock()
	uid := s.codeToUID[mqttDeviceID]
	s.mu.RUnlock()

	key := uid
	if key == "" {
		key = mqttDeviceID
	}

	s.mu.RLock()
	cached, ok := s.cardCache[key]
	s.mu.RUnlock()
	if ok {
		return cached, nil
	}

	info, err := s.cardDB.LookupCard(ctx, key)
	if err != nil {
		info, err = s.cardDB.LookupDeviceOnly(ctx, key)
		if err != nil {
			info, err = s.cardDB.LookupDeviceStoreOnly(ctx, key)
			if err != nil {
				return nil, err
			}
		}
	}

	s.mu.Lock()
	s.cardCache[info.DeviceUID] = info
	s.mu.Unlock()

	return info, nil
}

// ResolveBaseline 与 CardDB.ResolveDeviceBaseline 一致，供 health_check 等按 device_id / device_uid 解析策略与 device_code。
func (s *CardMappingService) ResolveBaseline(ctx context.Context, deviceKey string) (card.DeviceBaseline, bool) {
	if s.cardDB == nil || deviceKey == "" {
		return card.DeviceBaseline{}, false
	}
	return s.cardDB.ResolveDeviceBaseline(ctx, deviceKey)
}

// DeviceKeysInTenantUnit 透传 CardDB，供 config.card 按 unit 枚举设备。
func (s *CardMappingService) DeviceKeysInTenantUnit(ctx context.Context, tenantID, unitID string) (deviceIDs, deviceUIDs []string) {
	if s.cardDB == nil {
		return nil, nil
	}
	return s.cardDB.DeviceKeysInTenantUnit(ctx, tenantID, unitID)
}

// DeviceKeysInCard 透传 CardDB，按 card_id 枚举 devices JSON 中的设备键。
func (s *CardMappingService) DeviceKeysInCard(ctx context.Context, cardID string) (deviceIDs, deviceUIDs []string) {
	if s.cardDB == nil {
		return nil, nil
	}
	return s.cardDB.DeviceKeysInCard(ctx, cardID)
}

// ListSleepadBaselinesForHealth 定时全量探测用。
func (s *CardMappingService) ListSleepadBaselinesForHealth(ctx context.Context, deviceTypes []string) ([]card.DeviceBaseline, error) {
	if s.cardDB == nil {
		return nil, nil
	}
	if len(deviceTypes) == 0 {
		deviceTypes = sleepadDeviceTypes
	}
	return s.cardDB.ListSleepadBaselinesForHealth(ctx, deviceTypes)
}
